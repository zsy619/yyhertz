package mvc

// Package mvc 全局管理器模块
//
// 本文件实现了Session、Cookie和Template的全局管理器，
// 通过单例模式减少内存占用和初始化开销，提供统一的配置管理。
//
// 设计原则：
// - 线程安全：使用 sync.Once 确保只初始化一次
// - 懒加载：按需初始化，减少启动时间
// - 配置统一：从配置文件统一加载配置
// - 向后兼容：保持原有API不变
//
// 使用示例：
//
//	// 获取全局Session管理器
//	sessionManager := mvc.GetSessionManager()
//	store := sessionManager.GetOrCreateSession(ctx)
//
//	// 获取全局Cookie辅助器
//	cookieHelper := mvc.GetCookieHelper()
//	cookieHelper.SetCookie(ctx, "key", "value")
//
//	// 获取全局模板引擎
//	templateEngine := mvc.GetTemplateEngine()
//	templateEngine.Render(ctx, "template.html", data)

import (
	"sync"

	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/security"
	"github.com/zsy619/yyhertz/framework/mvc/session"
	"github.com/zsy619/yyhertz/framework/mvc/template"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// ============= 全局变量定义 =============

var (
	// 全局Session管理器
	globalSessionManager *session.Manager
	sessionOnce          sync.Once
	sessionInitialized   bool

	// 全局Cookie辅助器
	globalCookieHelper *cookie.Helper
	cookieOnce         sync.Once
	cookieInitialized  bool

	// 全局模板引擎
	globalTemplateEngine *view.TemplateEngine
	templateOnce         sync.Once
	templateInitialized  bool

	// 全局CSRF管理器
	globalCSRFManager *security.CSRFManager
	csrfOnce          sync.Once
	csrfInitialized   bool

	// 初始化状态标记
	globalsInitialized bool
	globalsInitMutex   sync.RWMutex
)

// ============= Session管理器 =============

// GetSessionManager 获取全局Session管理器（线程安全）
//
// 返回全局单例的Session管理器实例。如果尚未初始化，会使用默认配置自动初始化。
// 该方法是线程安全的，可以在并发环境中安全调用。
//
// 返回：
//   - *session.Manager: 全局Session管理器实例
//
// 示例：
//
//	manager := mvc.GetSessionManager()
//	store := manager.GetOrCreateSession(ctx.Request())
func GetSessionManager() *session.Manager {
	sessionOnce.Do(func() {
		initializeSessionManager()
	})
	return globalSessionManager
}

// SetGlobalSessionManager 设置全局Session管理器
//
// 允许用户自定义Session管理器实例。一旦设置，后续的GetSessionManager()调用
// 都会返回这个实例。
//
// 参数：
//   - manager: 要设置的Session管理器实例
//
// 注意：该方法不是线程安全的，应该在应用启动阶段调用
//
// 示例：
//
//	config := session.LoadFromConfig()
//	manager := session.NewManager(config)
//	mvc.SetGlobalSessionManager(manager)
func SetGlobalSessionManager(manager *session.Manager) {
	globalSessionManager = manager
	sessionInitialized = true
}

// initializeSessionManager 初始化Session管理器（内部使用）
func initializeSessionManager() {
	if !sessionInitialized {
		// 从配置文件加载Session配置
		config := session.LoadFromConfig()
		globalSessionManager = session.NewManager(config)
		sessionInitialized = true
	}
}

// ============= Cookie管理器 =============

// GetCookieHelper 获取全局Cookie辅助器（线程安全）
//
// 返回全局单例的Cookie辅助器实例。如果尚未初始化，会使用默认配置自动初始化。
// 该方法是线程安全的，可以在并发环境中安全调用。
//
// 返回：
//   - *cookie.Helper: 全局Cookie辅助器实例
//
// 示例：
//
//	helper := mvc.GetCookieHelper()
//	helper.SetCookie(ctx, "user_id", "123", cookie.DefaultOptions())
func GetCookieHelper() *cookie.Helper {
	cookieOnce.Do(func() {
		initializeCookieHelper()
	})
	return globalCookieHelper
}

// SetGlobalCookieHelper 设置全局Cookie辅助器
//
// 允许用户自定义Cookie辅助器实例。一旦设置，后续的GetCookieHelper()调用
// 都会返回这个实例。
//
// 参数：
//   - helper: 要设置的Cookie辅助器实例
//
// 注意：该方法不是线程安全的，应该在应用启动阶段调用
//
// 示例：
//
//	config := cookie.LoadFromConfig()
//	helper := cookie.NewHelper(config)
//	mvc.SetGlobalCookieHelper(helper)
func SetGlobalCookieHelper(helper *cookie.Helper) {
	globalCookieHelper = helper
	cookieInitialized = true
}

// initializeCookieHelper 初始化Cookie辅助器（内部使用）
func initializeCookieHelper() {
	if !cookieInitialized {
		// 从配置文件加载Cookie配置，如果失败则使用默认配置
		config := cookie.DefaultConfig()
		globalCookieHelper = cookie.NewHelper(config)
		cookieInitialized = true
	}
}

// ============= 模板引擎管理器 =============

// GetTemplateEngine 获取全局模板引擎（线程安全）
//
// 返回全局单例的模板引擎实例。模板引擎已经在template包中实现了单例模式，
// 这里提供统一的访问接口。
//
// 返回：
//   - *view.TemplateEngine: 全局模板引擎实例
//
// 示例：
//
//	engine := mvc.GetTemplateEngine()
//	err := engine.Render(ctx, "user/profile.html", data)
func GetTemplateEngine() *view.TemplateEngine {
	templateOnce.Do(func() {
		initializeTemplateEngine()
	})
	return globalTemplateEngine
}

// SetGlobalTemplateEngine 设置全局模板引擎
//
// 允许用户自定义模板引擎实例。一旦设置，后续的GetTemplateEngine()调用
// 都会返回这个实例。
//
// 参数：
//   - engine: 要设置的模板引擎实例
//
// 注意：该方法不是线程安全的，应该在应用启动阶段调用
//
// 示例：
//
//	config := view.LoadTemplateConfig()
//	engine, _ := view.NewTemplateEngine(config)
//	mvc.SetGlobalTemplateEngine(engine)
func SetGlobalTemplateEngine(engine *view.TemplateEngine) {
	globalTemplateEngine = engine
	templateInitialized = true
}

// initializeTemplateEngine 初始化模板引擎（内部使用）
func initializeTemplateEngine() {
	if !templateInitialized {
		// 使用template包中已有的单例
		templateManager := template.GetTemplateManager()
		globalTemplateEngine = templateManager.GetEngine()
		templateInitialized = true
	}
}

// ============= CSRF管理器 =============

// GetCSRFManager 获取全局CSRF管理器（线程安全）
//
// 返回全局单例的CSRF管理器实例。如果尚未初始化，会使用默认配置自动初始化。
// 该方法是线程安全的，可以在并发环境中安全调用。
//
// 返回：
//   - *security.CSRFManager: 全局CSRF管理器实例
//
// 示例：
//
//	manager := mvc.GetCSRFManager()
//	token, _ := manager.GenerateToken(userID, clientIP)
func GetCSRFManager() *security.CSRFManager {
	csrfOnce.Do(func() {
		initializeCSRFManager()
	})
	return globalCSRFManager
}

// SetGlobalCSRFManager 设置全局CSRF管理器
//
// 允许用户自定义CSRF管理器实例。一旦设置，后续的GetCSRFManager()调用
// 都会返回这个实例。
//
// 参数：
//   - manager: 要设置的CSRF管理器实例
//
// 注意：该方法不是线程安全的，应该在应用启动阶段调用
//
// 示例：
//
//	config := security.LoadCSRFConfig()
//	manager := security.NewCSRFManager(config)
//	mvc.SetGlobalCSRFManager(manager)
func SetGlobalCSRFManager(manager *security.CSRFManager) {
	globalCSRFManager = manager
	csrfInitialized = true
}

// initializeCSRFManager 初始化CSRF管理器（内部使用）
func initializeCSRFManager() {
	if !csrfInitialized {
		// 从配置文件加载CSRF配置
		config := security.LoadCSRFConfig()
		globalCSRFManager = security.NewCSRFManager(config)

		// 注册CSRF token提供者到view包，避免循环导入
		adapter := security.NewCSRFTokenAdapter(globalCSRFManager)
		view.SetCSRFTokenProvider(adapter)

		csrfInitialized = true
	}
}

// ============= 统一初始化接口 =============

// InitializeGlobalManagers 初始化所有全局管理器
//
// 统一初始化Session管理器、Cookie辅助器和模板引擎。
// 该方法是幂等的，多次调用不会重复初始化。
//
// 示例：
//
//	func init() {
//	    mvc.InitializeGlobalManagers()
//	}
func InitializeGlobalManagers() {
	globalsInitMutex.Lock()
	defer globalsInitMutex.Unlock()

	if globalsInitialized {
		return
	}

	// 初始化Session管理器
	GetSessionManager()

	// 初始化Cookie辅助器
	GetCookieHelper()

	// 初始化模板引擎
	GetTemplateEngine()

	// 初始化CSRF管理器
	GetCSRFManager()

	// 注册全局管理器访问器到 core 包，避免循环导入
	accessor := &globalManagerAccessorImpl{}
	core.SetGlobalManagerAccessor(accessor)

	globalsInitialized = true
}

// IsGlobalsInitialized 检查全局管理器是否已初始化
//
// 返回：
//   - bool: 如果所有全局管理器都已初始化返回true，否则返回false
func IsGlobalsInitialized() bool {
	globalsInitMutex.RLock()
	defer globalsInitMutex.RUnlock()
	return globalsInitialized
}

// ============= 状态查询接口 =============

// GetGlobalManagersStatus 获取全局管理器状态
//
// 返回所有全局管理器的初始化状态，用于调试和监控。
//
// 返回：
//   - map[string]bool: 包含各个管理器初始化状态的映射
func GetGlobalManagersStatus() map[string]bool {
	return map[string]bool{
		"session":  sessionInitialized,
		"cookie":   cookieInitialized,
		"template": templateInitialized,
		"csrf":     csrfInitialized,
		"all":      globalsInitialized,
	}
}

// ResetGlobalManagers 重置所有全局管理器（仅用于测试）
//
// 注意：该方法仅应在测试环境中使用，生产环境中调用可能导致未知问题
func ResetGlobalManagers() {
	globalsInitMutex.Lock()
	defer globalsInitMutex.Unlock()

	globalSessionManager = nil
	globalCookieHelper = nil
	globalTemplateEngine = nil
	globalCSRFManager = nil

	sessionInitialized = false
	cookieInitialized = false
	templateInitialized = false
	csrfInitialized = false
	globalsInitialized = false

	// 重置sync.Once
	sessionOnce = sync.Once{}
	cookieOnce = sync.Once{}
	templateOnce = sync.Once{}
	csrfOnce = sync.Once{}
}

// ============= 全局管理器访问器实现（避免循环导入） =============

// globalManagerAccessorImpl 实现了 core.globalManagerAccessor 接口
// 这个实现可以让 core 包访问 mvc 包的全局管理器，而不需要直接导入 mvc 包
type globalManagerAccessorImpl struct{}

// GetSessionManager 实现接口方法
func (g *globalManagerAccessorImpl) GetSessionManager() *session.Manager {
	return GetSessionManager()
}

// GetCookieHelper 实现接口方法
func (g *globalManagerAccessorImpl) GetCookieHelper() *cookie.Helper {
	return GetCookieHelper()
}

// GetTemplateEngine 实现接口方法
func (g *globalManagerAccessorImpl) GetTemplateEngine() *view.TemplateEngine {
	return GetTemplateEngine()
}

// GetCSRFManager 实现接口方法
func (g *globalManagerAccessorImpl) GetCSRFManager() *security.CSRFManager {
	return GetCSRFManager()
}

// IsInitialized 实现接口方法
func (g *globalManagerAccessorImpl) IsInitialized() bool {
	return IsGlobalsInitialized()
}
