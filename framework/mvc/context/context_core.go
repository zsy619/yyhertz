package context

import (
	stdcontext "context"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// ============= 基础类型定义 =============

// Param 路由参数结构
type Param struct {
	Key   string
	Value string
}

// Params 路由参数列表
type Params []Param

// ByName 通过名称获取参数值
func (ps Params) ByName(name string) string {
	for _, param := range ps {
		if param.Key == name {
			return param.Value
		}
	}
	return ""
}

// Get 获取参数值（别名方法）
func (ps Params) Get(name string) (string, bool) {
	for _, param := range ps {
		if param.Key == name {
			return param.Value, true
		}
	}
	return "", false
}

// 注意：ResponseWriter接口在response_writer.go中已定义

// HandlerFunc 处理函数类型
type HandlerFunc func(*Context)

// Context 增强的上下文，支持对象池化和高性能并发访问
//
// 设计原则：
// 1. 高性能：使用sync.Map和原子操作优化并发性能
// 2. 向后兼容：保持与旧版本API的完全兼容
// 3. 模块化：功能按模块拆分，便于维护
// 4. 类型安全：减少运行时错误
//
// 性能特性：
// - 支持对象池化，减少GC压力
// - 原子操作替代粗粒度锁
// - 高效的并发数据存储
type Context struct {
	// ============= 核心字段（优化后） =============

	// 核心上下文 - 私有字段，通过方法访问以确保安全性
	request *app.RequestContext
	writer  ResponseWriter
	Context stdcontext.Context

	// 路由相关
	params   Params // 路由参数
	fullPath string // 完整路径

	// 高性能并发数据存储 - 使用sync.Map替代map+mutex
	keys sync.Map // 上下文键值对，支持高并发访问

	// 兼容性字段 - 为了向后兼容传统MVC风格API
	Input  *InputData
	Output *OutputData

	// ============= 中间件状态（原子操作优化） =============

	index    int8          // 中间件索引
	handlers []HandlerFunc // 处理器链
	aborted  int32         // 是否中止 - 使用原子操作优化

	// ============= 错误处理（细粒度锁优化） =============

	errors []error    // 错误列表
	errMu  sync.Mutex // 专用错误锁，细粒度控制

	// ============= 池化相关 =============

	pooled   bool      // 是否来自池
	acquired time.Time // 获取时间

	engine *Engine // 引擎实例
}

// ============= 构造函数（优化版本） =============

// NewContext 创建新的增强Context（使用池化）
func NewContext(c *app.RequestContext) *Context {
	return NewContextWithContext(c, stdcontext.Background())
}

// NewContextWithContext 使用指定context创建增强Context
func NewContextWithContext(c *app.RequestContext, parent stdcontext.Context) *Context {
	ctx := defaultPool.Get()
	ctx.request = c
	ctx.Context = parent
	ctx.writer = &responseWriter{RequestContext: c}

	// 初始化兼容性字段
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}

	return ctx
}