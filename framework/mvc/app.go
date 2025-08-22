// Package mvc 提供MVC框架的统一入口和类型定义
//
// 本文件作为MVC框架的主入口，提供了对所有子模块的统一封装和重新导出。
// 通过类型别名和函数重导出，保持了向后兼容性，同时提供了简洁的API。
//
// 主要功能：
// - 核心类型导出：App、IController、HandlerFunc等
// - 组件系统集成：Session、Cookie、Captcha、Router等
// - 全局实例管理：统一的应用实例和初始化
// - 过滤器支持：完整的过滤器生命周期管理
// - 模板系统：全局模板函数和管理功能
//
// 设计原则：
// - 向后兼容：保持与旧版本的API兼容性
// - 统一管理：所有组件通过一个入口访问
// - 延迟初始化：只在需要时才创建实例
// - 线程安全：所有全局操作都是线程安全的
//
// 使用示例：
//
//	// 基础使用
//	app := mvc.GetAppInstance()
//	mvc.RouterAuto(&UserController{})
//
//	// 组件使用
//	sessionManager := mvc.NewSessionManager(mvc.DefaultSessionConfig())
//	cookieHelper := mvc.NewCookieHelper(mvc.DefaultCookieConfig())
package mvc

// ============= 模块导入 =============

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc/annotation"
	"github.com/zsy619/yyhertz/framework/mvc/captcha"
	"github.com/zsy619/yyhertz/framework/mvc/comment"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/define"
	errPkg "github.com/zsy619/yyhertz/framework/mvc/errors"
	"github.com/zsy619/yyhertz/framework/mvc/router"
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

type (
	H = utils.H // utils.H 用于简化map操作
)

// ============= 核心类型别名 =============

// 保留必要的类型别名，删除过度包装的类型
type (
	// App MVC应用核心类型
	App = core.App

	// RequestContext 请求上下文类型
	RequestContext = define.RequestContext

	// HandlerFunc HTTP处理函数类型
	HandlerFunc = define.HandlerFunc

	// IController 控制器接口
	IController = core.IController

	// FilterFunc 过滤器函数类型
	FilterFunc = define.FilterFunc

	// FilterPattern 过滤器模式
	FilterPattern = core.FilterPattern

	// WsConn WebSocket 连接类型
	WsConn = define.WsConn

	// WsClientUpgrader WebSocket 客户端升级器类型
	WsClientUpgrader = define.WsClientUpgrader

	// WsHertzUpgrader WebSocket Hertz 升级器类型
	WsHertzUpgrader = define.WsHertzUpgrader
)

// ============= 过滤器位置常量 =============

// 过滤器执行位置常量，定义过滤器在请求处理流程中的执行时机
//
// 执行顺序：BeforeStatic -> BeforeRouter -> BeforeExec -> [Controller] -> AfterExec -> FinishRouter
const (
	// BeforeStatic 在静态文件处理之前执行
	// 适用于：全局安全检查、访问限制、请求预处理
	BeforeStatic = constant.BeforeStatic

	// BeforeRouter 在路由匹配之前执行
	// 适用于：CORS处理、请求解析、通用中间件
	BeforeRouter = constant.BeforeRouter

	// BeforeExec 在控制器方法执行之前执行
	// 适用于：认证验证、权限检查、参数验证
	BeforeExec = constant.BeforeExec

	// AfterExec 在控制器方法执行之后执行
	// 适用于：响应处理、日志记录、统计分析
	AfterExec = constant.AfterExec

	// FinishRouter 在整个请求处理完成后执行
	// 适用于：资源清理、性能监控、最终日志
	FinishRouter = constant.FinishRouter
)

// ============= 全局变量和函数导出 =============

// 全局状态管理和函数重导出
var (
	// 核心函数重导出，保持API简洁性
	GetAppInstance      = core.GetAppInstance      // 获取单例应用实例
	NewApp              = core.NewApp              // 创建新的应用实例
	NewAppWithLogConfig = core.NewAppWithLogConfig // 使用日志配置创建应用
	AdaptHandler        = core.AdaptHandler        // 处理函数适配器

	// 全局应用实例，由init()函数初始化
	HertzApp *App

	// 初始化状态标记，用于检查框架是否已完成初始化
	IsInitComplete = false
)

// ============= Session会话管理组件 =============

// Session组件的快捷创建函数 - 保留常用的函数别名
var (
	DefaultSessionConfig = session.DefaultConfig
	NewSessionManager    = session.NewManager
)

// ============= Cookie管理组件 =============

// Cookie组件的快捷创建函数 - 保留常用的函数别名
var (
	DefaultCookieConfig  = cookie.DefaultConfig
	DefaultCookieOptions = cookie.DefaultOptions
	NewCookieHelper      = cookie.NewHelper
)

// ============= 路由管理组件 =============

// 保留必要的路由类型别名
type RouterGroup = router.Group

// Router组件的快捷创建函数
var (
	NewRouter = router.NewRouter
	NewGroup  = router.NewGroup
)

// ============= 验证码组件 =============

// Captcha组件的快捷创建函数和处理器 - 只保留常用的函数别名
var (
	DefaultCaptchaConfig   = captcha.DefaultConfig
	NewCaptchaGenerator    = captcha.NewGenerator
	NewMemoryStore         = captcha.NewMemoryStore
	NewCaptchaMiddleware   = captcha.NewMiddleware
	CaptchaGenerateHandler = captcha.GenerateHandler
	CaptchaImageHandler    = captcha.ImageHandler
	CaptchaVerifyHandler   = captcha.VerifyHandler
)

func init() {
	// 使用 config.GinLogConfig() 设置hertz的日志配置
	logConfig := config.GinLogConfig()
	l := logConfig.CreateLogger()
	hlog.SetLogger(l)

	// 初始化全局管理器（Session、Cookie、Template、CSRF）
	InitializeGlobalManagers()

	// 创建全局Hertz应用实例
	HertzApp = GetAppInstance()

	errPkg.QuickSetup("development")

	// 启用Beego风格的自动错误处理
	HertzApp.EnableAutoErrorHandling()

	// 添加测试路由（演示手动错误触发）
	HertzApp.GET("/test-error", func(ctx context.Context, c *RequestContext) {
		HertzApp.TriggerError(c, 404, fmt.Errorf("测试错误"))
	})

	// 创建注解应用
	AnnotationApp = annotation.NewAnnotationWithApp(HertzApp)

	// 创建注释注解应用
	CommentApp = comment.NewCommentWithApp(HertzApp)

	// 完成初始化
	IsInitComplete = true
}

// ============= HertzApp 静态方法 =============

// SetStaticPath 设置静态文件路径的静态方法
// 参数：localDir - 静态文件本地目录（相对应用所在目录）
//
//	urlPath - URL路径（可选），如果不提供则自动推导
//
// 示例：SetStaticPath("public", "/static") 或 SetStaticPath("public")
func SetStaticPath(localDir string, urlPath ...string) {
	if HertzApp != nil {
		HertzApp.SetStaticPath(localDir, urlPath...)
	}
}

// ============= 模板函数管理静态方法 =============

// AddFuncMap 添加全局模板函数的静态方法
// 参数：name - 函数名字符串，fn - 函数实现
// 示例：AddFuncMap("containString", tool.ContainString)
func AddFuncMap(name string, fn any) error {
	if HertzApp != nil {
		HertzApp.AddFuncMap(name, fn)
		return nil
	}
	return fmt.Errorf("hertz app instance is not initialized")
}

// AddFuncMaps 批量添加全局模板函数的静态方法
func AddFuncMaps(funcs map[string]any) error {
	if HertzApp != nil {
		for name, fn := range funcs {
			HertzApp.AddFuncMap(name, fn)
		}
		return nil
	}
	return fmt.Errorf("hertz app instance is not initialized")
}

// GetGlobalFuncMap 获取全局模板函数映射的静态方法
func GetGlobalFuncMap() map[string]any {
	if HertzApp != nil {
		return HertzApp.GetGlobalFuncMap()
	}
	return make(map[string]any)
}

// RemoveFuncMap 移除全局模板函数的静态方法
func RemoveFuncMap(name string) {
	if HertzApp != nil {
		HertzApp.RemoveFuncMap(name)
	}
}

// ListFuncMap 列出所有已注册的模板函数名称的静态方法
func ListFuncMap() []string {
	if HertzApp != nil {
		return HertzApp.ListFuncMap()
	}
	return []string{}
}

// ============= 过滤器管理静态方法 =============

// InsertFilter 插入过滤器到指定位置的静态方法
// 参数：pattern - 路径模式 (支持通配符 *)
//
//	position - 过滤器位置 (BeforeStatic, BeforeRouter, BeforeExec, AfterExec, FinishRouter)
//	filter - 过滤器函数
//	params - 可选参数 (第一个bool值表示是否启用，默认true)
//
// 示例：InsertFilter("/api/*", BeforeRouter, authFilter)
func InsertFilter(pattern string, position int, filter FilterFunc, params ...bool) {
	if HertzApp != nil {
		HertzApp.InsertFilter(pattern, position, filter, params...)
	}
}

// RemoveFilter 移除指定模式和位置的过滤器的静态方法
func RemoveFilter(pattern string, position int) bool {
	if HertzApp != nil {
		return HertzApp.RemoveFilter(pattern, position)
	}
	return false
}

// ListFilters 列出指定位置的所有过滤器的静态方法
func ListFilters(position int) []*FilterPattern {
	if HertzApp != nil {
		return HertzApp.ListFilters(position)
	}
	return []*FilterPattern{}
}

// GetAllFilters 获取所有位置的过滤器的静态方法
func GetAllFilters() map[int][]*FilterPattern {
	if HertzApp != nil {
		return HertzApp.GetAllFilters()
	}
	return make(map[int][]*FilterPattern)
}

// ============= Beego风格的错误处理静态方法 =============

// EnableAutoErrorHandling 启用自动错误处理的静态方法
func EnableAutoErrorHandling() {
	if HertzApp != nil {
		HertzApp.EnableAutoErrorHandling()
	}
}

// TriggerError 触发错误处理的静态方法
func TriggerError(ctx *RequestContext, statusCode int, err error) error {
	if HertzApp != nil {
		return HertzApp.TriggerError(ctx, statusCode, err)
	}
	return err
}

// Abort 中止请求并触发错误处理的静态方法（类似Beego的Abort）
func Abort(ctx *RequestContext, statusCode int, message ...string) {
	if HertzApp != nil {
		HertzApp.Abort(ctx, statusCode, message...)
	}
}

// AbortWithError 中止请求并使用指定错误的静态方法
func AbortWithError(ctx *RequestContext, statusCode int, err error) {
	if HertzApp != nil {
		HertzApp.AbortWithError(ctx, statusCode, err)
	}
}

// ============= 全局路由组方法 =============

// CreateGroup 创建支持多处理器的路由组（全局方法）
func CreateGroup(relativePath string) *RouterGroup {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	// 创建新的路由器实例
	router := NewRouter(HertzApp)

	// 创建支持多处理器的路由组
	return NewGroup(router, relativePath)
}
