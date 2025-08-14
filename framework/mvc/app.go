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
	"sync"

	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc/annotation"
	"github.com/zsy619/yyhertz/framework/mvc/captcha"
	"github.com/zsy619/yyhertz/framework/mvc/comment"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/router"
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

// ============= 核心类型别名 =============

// 核心类型别名，保持向后兼容性并简化导入
type (
	// App MVC应用核心类型，管理路由、中间件和生命周期
	App = core.App

	// RequestContext 请求上下文类型，封装HTTP请求和响应
	RequestContext = core.RequestContext

	// HandlerFunc HTTP处理函数类型，用于处理请求
	HandlerFunc = core.HandlerFunc

	// IController 控制器接口，定义控制器的基本约定
	IController = core.IController

	// FilterFunc 过滤器函数类型，用于请求拦截和处理
	FilterFunc = core.FilterFunc

	// FilterPattern 过滤器模式，定义过滤器的匹配规则
	FilterPattern = core.FilterPattern
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
	// 线程安全控制
	once  sync.Once    // 确保全局初始化只执行一次
	mutex sync.RWMutex // 保护全局状态的读写锁

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

// Session相关类型别名，提供统一的会话管理功能
type (
	// SessionConfig 会话配置，定义会话的存储、过期等参数
	SessionConfig = session.Config

	// SessionManager 会话管理器，负责会话的创建、存取和销毁
	SessionManager = session.Manager

	// SessionStore 会话存储接口，定义会话数据的存储方式
	SessionStore = session.Store
)

// Session组件的快捷创建函数
var (
	// DefaultSessionConfig 获取默认会话配置
	DefaultSessionConfig = session.DefaultConfig

	// NewSessionManager 创建新的会话管理器
	NewSessionManager = session.NewManager
)

// ============= Cookie管理组件 =============

// Cookie相关类型别名，提供安全的Cookie操作功能
type (
	// CookieConfig Cookie全局配置，设置默认的安全参数
	CookieConfig = cookie.Config

	// CookieOptions 单个Cookie的配置选项
	CookieOptions = cookie.Options

	// CookieHelper Cookie操作辅助工具，提供高级的Cookie管理功能
	CookieHelper = cookie.Helper
)

// Cookie组件的快捷创建函数
var (
	// DefaultCookieConfig 获取默认Cookie配置
	DefaultCookieConfig = cookie.DefaultConfig

	// DefaultCookieOptions 获取默认Cookie选项
	DefaultCookieOptions = cookie.DefaultOptions

	// NewCookieHelper 创建新的Cookie辅助工具
	NewCookieHelper = cookie.NewHelper
)

// ============= 路由管理组件 =============

// Router相关类型别名，提供灵活的路由管理功能
type (
	// RouterAlias 路由器类型，管理URL到处理函数的映射
	RouterAlias = router.Router

	// RouterGroup 路由组，用于将相关路由分组管理
	RouterGroup = router.Group
)

// Router组件的快捷创建函数
var (
	// NewRouter 创建新的路由器实例
	NewRouter = router.NewRouter

	// NewGroup 创建新的路由组
	NewGroup = router.NewGroup
)

// ============= 验证码组件 =============

// Captcha相关类型别名，提供完整的验证码解决方案
type (
	// CaptchaConfig 验证码配置，定义验证码的生成参数
	CaptchaConfig = captcha.Config

	// CaptchaGenerator 验证码生成器，负责生成图片验证码
	CaptchaGenerator = captcha.Generator

	// CaptchaStore 验证码存储接口，管理验证码的存储和验证
	CaptchaStore = captcha.Store

	// CaptchaMiddleware 验证码中间件，自动处理验证码验证
	CaptchaMiddleware = captcha.Middleware

	// CaptchaMiddlewareConfig 验证码中间件配置
	CaptchaMiddlewareConfig = captcha.MiddlewareConfig

	M = map[string]any
)

// Captcha组件的快捷创建函数和处理器
var (
	// DefaultCaptchaConfig 获取默认验证码配置
	DefaultCaptchaConfig = captcha.DefaultConfig

	// NewCaptchaGenerator 创建新的验证码生成器
	NewCaptchaGenerator = captcha.NewGenerator

	// NewMemoryStore 创建内存型验证码存储
	NewMemoryStore = captcha.NewMemoryStore

	// NewCaptchaMiddleware 创建验证码中间件
	NewCaptchaMiddleware = captcha.NewMiddleware

	// 验证码HTTP处理器函数
	CaptchaGenerateHandler = captcha.GenerateHandler // 生成验证码接口
	CaptchaImageHandler    = captcha.ImageHandler    // 验证码图片接口
	CaptchaVerifyHandler   = captcha.VerifyHandler   // 验证码验证接口
)

func init() {
	// 初始化全局Hertz应用实例
	once.Do(func() {
		mutex.Lock()
		defer mutex.Unlock()

		// 创建全局Hertz应用实例
		HertzApp = GetAppInstance()

		// 创建注解应用
		AnnotationApp = annotation.NewAnnotationWithApp(HertzApp)

		// 创建注释注解应用
		CommentApp = comment.NewCommentWithApp(HertzApp)

		// 完成初始化
		IsInitComplete = true
	})
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
func AddFuncMap(name string, fn any) {
	if HertzApp != nil {
		HertzApp.AddFuncMap(name, fn)
	}
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
