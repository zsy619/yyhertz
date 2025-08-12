package mvc

// 重新导出核心功能，保持向后兼容
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

// 类型别名，保持向后兼容
type App = core.App
type RequestContext = core.RequestContext
type HandlerFunc = core.HandlerFunc
type IController = core.IController
type FilterFunc = core.FilterFunc
type FilterPattern = core.FilterPattern

// 过滤器位置常量 - 使用统一常量
const (
	BeforeStatic = constant.BeforeStatic // 静态文件处理前
	BeforeRouter = constant.BeforeRouter // 路由匹配前
	BeforeExec   = constant.BeforeExec   // 控制器执行前
	AfterExec    = constant.AfterExec    // 控制器执行后
	FinishRouter = constant.FinishRouter // 请求处理完成后
)

// 重新导出常用功能
var (
	once sync.Once

	mutex sync.RWMutex

	GetAppInstance      = core.GetAppInstance
	NewApp              = core.NewApp
	NewAppWithLogConfig = core.NewAppWithLogConfig
	AdaptHandler        = core.AdaptHandler

	HertzApp *App

	IsInitComplete = false // 是否完成初始化
)

// Session相关类型别名
type (
	SessionConfig  = session.Config
	SessionManager = session.Manager
	SessionStore   = session.Store
)

var (
	DefaultSessionConfig = session.DefaultConfig
	NewSessionManager    = session.NewManager
)

// Cookie相关类型别名
type CookieConfig = cookie.Config
type CookieOptions = cookie.Options
type CookieHelper = cookie.Helper

var (
	DefaultCookieConfig  = cookie.DefaultConfig
	DefaultCookieOptions = cookie.DefaultOptions
	NewCookieHelper      = cookie.NewHelper
)

// Router相关类型别名
type (
	Router      = router.Router
	RouterGroup = router.Group
)

var (
	NewRouter = router.NewRouter
	NewGroup  = router.NewGroup
)

// Captcha相关类型别名
type (
	CaptchaConfig           = captcha.Config
	CaptchaGenerator        = captcha.Generator
	CaptchaStore            = captcha.Store
	CaptchaMiddleware       = captcha.Middleware
	CaptchaMiddlewareConfig = captcha.MiddlewareConfig
)

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
