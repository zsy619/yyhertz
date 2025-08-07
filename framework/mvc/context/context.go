package context

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// HandlerFunc 处理函数类型
type HandlerFunc func(*Context)

// Context 增强的上下文，支持对象池化
type Context struct {
	// 核心上下文
	Request *app.RequestContext
	RequestContext *app.RequestContext // 兼容性字段别名
	Context context.Context
	
	// 路由相关
	Params   Params // 路由参数
	FullPath string // 完整路径

	// 请求数据
	Keys map[string]interface{} // 上下文键值对
	
	// 响应数据  
	Writer ResponseWriter
	ResponseWriter ResponseWriter // 兼容性字段别名
	
	// 兼容性字段 - 为了向后兼容Beego风格API
	Input  *InputData
	Output *OutputData
	
	// 内部状态
	index    int8           // 中间件索引
	handlers []HandlerFunc  // 处理器链
	mu       sync.RWMutex   // 读写锁
	aborted  bool           // 是否中止
	errors   []error        // 错误列表
	
	// 池化标识
	pooled   bool           // 是否来自池
	acquired time.Time      // 获取时间
}


// Reset 重置Context状态，准备复用
func (ctx *Context) Reset() {
	ctx.Request = nil
	ctx.RequestContext = nil // 同时重置兼容性别名
	ctx.Context = nil
	ctx.Params = ctx.Params[:0]
	ctx.FullPath = ""
	
	// 清空Keys但保留底层数组
	for k := range ctx.Keys {
		delete(ctx.Keys, k)
	}
	
	ctx.Writer = nil
	ctx.index = -1
	ctx.handlers = ctx.handlers[:0]
	ctx.aborted = false
	ctx.errors = ctx.errors[:0]
}

// NewContext 创建新的增强Context（使用池化）
func NewContext(c *app.RequestContext) *Context {
	ctx := defaultPool.Get()
	ctx.Request = c
	ctx.RequestContext = c // 兼容性别名指向同一对象
	ctx.Context = context.Background()
	ctx.Writer = &responseWriter{RequestContext: c}
	ctx.ResponseWriter = ctx.Writer // 兼容性别名指向同一对象
	
	// 初始化Beego风格兼容性字段
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}
	
	return ctx
}

// NewContextWithContext 使用指定context创建增强Context
func NewContextWithContext(c *app.RequestContext, parent context.Context) *Context {
	ctx := defaultPool.Get()
	ctx.Request = c
	ctx.RequestContext = c // 兼容性别名指向同一对象
	ctx.Context = parent
	ctx.Writer = &responseWriter{RequestContext: c}
	ctx.ResponseWriter = ctx.Writer // 兼容性别名指向同一对象
	
	// 初始化Beego风格兼容性字段
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}
	
	return ctx
}

// Release 释放Context到池中
func (ctx *Context) Release() {
	if ctx.pooled {
		defaultPool.Put(ctx)
		atomic.AddInt32(&poolSize, -1)
	}
}

// ============= Context核心方法 =============

// Next 执行下一个中间件
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		if !ctx.aborted {
			ctx.handlers[ctx.index](ctx)
		}
		ctx.index++
	}
}

// Abort 中止执行
func (ctx *Context) Abort() {
	ctx.aborted = true
}

// IsAborted 是否已中止
func (ctx *Context) IsAborted() bool {
	return ctx.aborted
}

// Set 设置键值对
func (ctx *Context) Set(key string, value interface{}) {
	ctx.mu.Lock()
	ctx.Keys[key] = value
	ctx.mu.Unlock()
}

// Get 获取值
func (ctx *Context) Get(key string) (interface{}, bool) {
	ctx.mu.RLock()
	value, exists := ctx.Keys[key]
	ctx.mu.RUnlock()
	return value, exists
}

// MustGet 必须获取值
func (ctx *Context) MustGet(key string) interface{} {
	if value, exists := ctx.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// Param 获取路由参数
func (ctx *Context) Param(key string) string {
	return ctx.Params.ByName(key)
}

// Query 获取查询参数
func (ctx *Context) Query(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.QueryArgs().Peek(key))
}

// PostForm 获取POST表单参数
func (ctx *Context) PostForm(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.PostArgs().Peek(key))
}

// Header 获取请求头
func (ctx *Context) Header(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.GetHeader(key))
}

// GetHeader 获取请求头 (兼容性别名)
func (ctx *Context) GetHeader(key string) string {
	return ctx.Header(key)
}

// ============= 响应方法 =============

// JSON 返回JSON响应
func (ctx *Context) JSON(code int, obj interface{}) {
	if ctx.Request != nil {
		ctx.Request.JSON(code, obj)
	}
}

// String 返回字符串响应
func (ctx *Context) String(code int, format string, values ...interface{}) {
	if ctx.Request != nil {
		ctx.Request.String(code, format, values...)
	}
}

// HTML 返回HTML响应
func (ctx *Context) HTML(code int, name string, obj interface{}) {
	if ctx.Request != nil {
		// 这里需要集成模板引擎
		ctx.Request.HTML(code, name, obj)
	}
}

// SetHandlers 设置处理器链
func (ctx *Context) SetHandlers(handlers []HandlerFunc) {
	ctx.handlers = handlers
	ctx.index = -1
}

// ============= 错误处理方法 =============

// AddError 添加错误
func (ctx *Context) AddError(err error) {
	if err != nil {
		ctx.mu.Lock()
		ctx.errors = append(ctx.errors, err)
		ctx.mu.Unlock()
	}
}

// GetErrors 获取所有错误
func (ctx *Context) GetErrors() []error {
	ctx.mu.RLock()
	errors := make([]error, len(ctx.errors))
	copy(errors, ctx.errors)
	ctx.mu.RUnlock()
	return errors
}

// HasErrors 是否有错误
func (ctx *Context) HasErrors() bool {
	ctx.mu.RLock()
	hasErr := len(ctx.errors) > 0
	ctx.mu.RUnlock()
	return hasErr
}

// ClearErrors 清除所有错误
func (ctx *Context) ClearErrors() {
	ctx.mu.Lock()
	ctx.errors = ctx.errors[:0]
	ctx.mu.Unlock()
}

// LastError 获取最后一个错误
func (ctx *Context) LastError() error {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	
	if len(ctx.errors) == 0 {
		return nil
	}
	return ctx.errors[len(ctx.errors)-1]
}

// ============= 兼容性方法 =============

// AbortWithStatus 终止并设置状态码 (兼容性方法)
func (ctx *Context) AbortWithStatus(code int) {
	ctx.JSON(code, map[string]string{"error": "Request aborted"})
	ctx.Abort()
}

// Write 写入响应数据 (兼容性方法)
func (ctx *Context) Write(data []byte) (int, error) {
	return ctx.Writer.Write(data)
}

// SetHeader 设置响应头 (兼容性方法)
func (ctx *Context) SetHeader(key, value string) {
	if ctx.Request != nil {
		ctx.Request.Response.Header.Set(key, value)
	}
}