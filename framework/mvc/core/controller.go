package core

import (
	"html/template"

	"github.com/cloudwego/hertz/pkg/app"

	context "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/session"
	"github.com/zsy619/yyhertz/framework/mvc/unified"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// BaseController 基础控制器结构，实现了完整的IController接口
//
// BaseController 是所有控制器的基类，完全兼容Beego的ControllerInterface规范。
// 它提供了Web开发中所需的全部功能，包括但不限于：
//
// 核心功能：
//   - 完整的HTTP请求处理 (GET, POST, PUT, DELETE等)
//   - 多格式响应管理 (JSON, HTML, XML, YAML等)
//   - 会话和Cookie管理
//   - 模板渲染系统
//   - XSRF/CSRF安全防护
//   - 参数获取和验证
//   - 错误处理和流程控制
//
// 接口实现：
//   实现了完整的IController接口，包括：
//   - 生命周期方法: Init(), Prepare(), Finish()
//   - 控制器信息: GetControllerName(), GetActionName()
//   - 模板渲染: Render()
//   - 安全功能: XSRFToken(), CheckXSRFCookie()
//   - 路由映射: URLMapping(), HandlerFunc()
//   - 流程控制: ShouldStopExecution()
//
// 使用示例：
//
//	type UserController struct {
//		mvc.BaseController
//	}
//
//	func (c *UserController) Get() {
//		user := c.GetCurrentUser()
//		c.Data["json"] = map[string]any{
//			"user": user,
//		}
//		c.ServeJSON() // 自动调用StopRun()
//	}
//
//	func (c *UserController) Post() {
//		if !c.CheckXSRFCookie() {
//			c.Abort("403")
//			return
//		}
//		// 处理POST请求
//	}
//
type BaseController struct {
	Ctx *context.Context // 统一的上下文

	// ============= 控制器属性 =============

	// 控制器和动作信息
	ControllerName string            // 控制器名称
	ActionName     string            // 当前执行的动作名称
	MethodMapping  map[string]string // HTTP方法到控制器方法的映射

	// 路由信息
	RoutePattern string                      // 路由模式
	RouteParams  map[string]string           // 路由参数
	URLGenerator func(string, ...any) string // URL生成函数

	// 应用控制器引用
	AppController IController // 应用控制器实例引用
	
	// 应用实例引用（用于访问全局功能）
	app *App // 应用实例引用

	// ============= Beego风格的模板属性 =============

	// 模板路径配置
	ViewPath       string            // 视图文件路径
	LayoutPath     string            // 布局文件路径
	Layout         string            // 当前使用的布局文件名
	LayoutSections map[string]string // 布局分区内容
	TplName        string            // 模板文件名
	TplPrefix      string            // 模板文件前缀
	TplExt         string            // 模板文件扩展名

	// 模板数据和函数
	Data            map[string]any   // 模板数据
	xsrfToken       string           // XSRF令牌（私有字段）
	checkXSRFCookie bool             // 是否检查XSRF Cookie（私有字段）
	TplFuncs        template.FuncMap // 自定义模板函数

	// Beego兼容的URL映射和处理器
	URLMappings  map[string]string // URL模式到方法名的映射
	HandlerFuncs map[string]bool   // 可用的处理器函数映射
	XSRFExpire   int               // XSRF令牌过期时间（秒）

	// 模板引擎配置
	EnableRender bool   // 是否启用模板渲染
	EnableGzip   bool   // 是否启用Gzip压缩
	ViewsPath    string // 视图根路径（兼容性）

	// 辅助工具
	cookieHelper   *cookie.Helper              // Cookie辅助工具
	sessionHelper  *session.Manager            // Session管理器
	templateEngine *view.TemplateEngine        // 模板引擎实例
	includeEngine  *view.TemplateIncludeEngine // 支持include的模板引擎
	
	// 统一管理器
	unifiedManager *unified.Manager // 统一管理器实例

	// ============= 优化控制器特性 =============

	// 优化功能控制
	optimizationEnabled bool     // 是否启用优化特性
	middlewareList      []string // 中间件列表，支持GetMiddleware()
	
	// ============= 响应流程控制 =============
	
	// 响应状态控制（内部使用，不暴露给外部）
	shouldStopExecution bool     // 是否应该停止执行

	// 内部控制字段
	initialized bool // 控制器名称是否已初始化（内部使用）
}

// 编译时检查：确保 BaseController 完全实现了 IController 接口
var _ IController = (*BaseController)(nil)

// NewBaseController 创建新的基础控制器实例
func NewBaseController() *BaseController {
	return &BaseController{
		// 基础数据
		Data:           make(map[string]any),
		LayoutSections: make(map[string]string),
		TplFuncs:       make(template.FuncMap),

		// 控制器属性
		MethodMapping:  make(map[string]string),
		RouteParams:    make(map[string]string),
		ControllerName: "",
		ActionName:     "",
		RoutePattern:   "",

		// Beego兼容属性
		URLMappings:  make(map[string]string),
		HandlerFuncs: make(map[string]bool),
		XSRFExpire:   3600, // 默认1小时

		// 默认路径配置
		ViewPath:   "views",
		LayoutPath: "views/layout",
		ViewsPath:  "views", // 兼容性
		Layout:     "layout.html",
		TplExt:     ".html",
		TplPrefix:  "",

		// 功能开关
		EnableRender:    true,
		EnableGzip:      false,
		checkXSRFCookie: false,

		// 优化特性
		optimizationEnabled: false,             // 默认不启用优化
		middlewareList:      make([]string, 0), // 初始化空中间件列表

		// 辅助工具 - 使用全局管理器的引用（向后兼容）
		// 注意：这些字段现在指向全局实例，而不是每个控制器的独立实例
		cookieHelper:   nil, // 将在 initializeBaseController 中设置
		sessionHelper:  nil, // 将在 initializeBaseController 中设置
		templateEngine: nil, // 将在 initializeBaseController 中设置
		
		// 统一管理器
		unifiedManager: unified.GetManager(), // 获取全局统一管理器
	}
}

// NewBaseControllerWithContext 使用指定上下文创建控制器
func NewBaseControllerWithContext(ctx *app.RequestContext) *BaseController {
	c := NewBaseController()
	// 创建增强的Context并设置
	enhancedCtx := context.NewContext(ctx)
	c.Ctx = enhancedCtx
	return c
}

// ============= 生命周期方法 =============

// Init 初始化控制器（完全兼容Beego ControllerInterface规范）
func (c *BaseController) Init(ct *context.Context, controllerName, actionName string, app any) {
	// 设置统一的Context
	c.Ctx = ct

	// 设置控制器和动作信息
	c.ControllerName = controllerName
	c.ActionName = actionName

	if c.Data == nil {
		c.Data = make(map[string]any)
	}

	// 设置应用实例引用
	if app != nil {
		// 尝试类型断言为IController
		if appController, ok := app.(IController); ok {
			c.AppController = appController
		}
		
		// 尝试类型断言为*App
		if appInstance, ok := app.(*App); ok {
			c.app = appInstance
		}
	}

	// 初始化其他组件
	c.initializeBaseController()
}

// initializeBaseController 初始化基础控制器属性
func (c *BaseController) initializeBaseController() {
	// 设置默认值
	if c.ViewPath == "" {
		c.ViewPath = "views"
	}
	if c.LayoutPath == "" {
		c.LayoutPath = "views/layout"
	}
	if c.Layout == "" {
		c.Layout = "layout.html"
	}
	c.EnableRender = true
	
	// 为了向后兼容，保留字段设置，但它们现在可能指向全局实例
	// 通过 getGlobalManagerInstances 函数获取全局实例（如果可用）
	c.ensureHelpersInitialized()
}

// ensureHelpersInitialized 确保辅助工具已初始化
// 这个方法会尝试获取全局实例，如果不可用则创建本地实例
// 优先使用统一管理器，向后兼容独立组件
func (c *BaseController) ensureHelpersInitialized() {
	// 确保统一管理器已初始化
	if c.unifiedManager == nil {
		c.unifiedManager = unified.GetManager()
	}
	
	// 尝试从统一管理器获取组件实例
	if c.unifiedManager != nil && c.unifiedManager.IsInitialized() {
		// 使用统一管理器的组件
		c.cookieHelper = c.unifiedManager.GetCookieHelper()
		c.sessionHelper = c.unifiedManager.GetSessionManager()
		c.templateEngine = c.unifiedManager.GetTemplateEngine()
		return
	}
	
	// 向后兼容：如果统一管理器不可用，使用独立的全局实例
	if c.cookieHelper == nil {
		if globalHelper := getGlobalCookieHelperIfAvailable(); globalHelper != nil {
			c.cookieHelper = globalHelper
		} else {
			c.cookieHelper = cookie.NewHelper(cookie.DefaultConfig())
		}
	}
	
	if c.sessionHelper == nil {
		if globalManager := getGlobalSessionManagerIfAvailable(); globalManager != nil {
			c.sessionHelper = globalManager
		} else {
			c.sessionHelper = session.NewManager(session.DefaultConfig())
		}
	}
	
	if c.templateEngine == nil {
		if globalEngine := getGlobalTemplateEngineIfAvailable(); globalEngine != nil {
			c.templateEngine = globalEngine
		} else {
			c.templateEngine = view.GetTemplateManager().GetEngine()
		}
	}
}

// Prepare 预处理方法
func (c *BaseController) Prepare() {
	// 默认实现为空，子类可以重写
}

// Finish 后处理方法
func (c *BaseController) Finish() {
	// 如果启用了优化特性，自动调用Destroy进行资源清理
	if c.optimizationEnabled {
		c.Destroy()
	}
	// 默认实现为空，子类可以重写
}

// ============= 优化控制器扩展方法 =============

// InitWithContext 优化控制器兼容的初始化方法
func (c *BaseController) InitWithContext(ctx *context.Context) error {
	c.Ctx = ctx
	if c.Data == nil {
		c.Data = make(map[string]any)
	}
	c.initializeBaseController()
	return nil
}

// Destroy 资源清理方法（优化控制器特性）
func (c *BaseController) Destroy() error {
	// 清理Context引用
	c.Ctx = nil

	// 清理模板数据
	if c.Data != nil {
		for k := range c.Data {
			delete(c.Data, k)
		}
	}

	// 清理路由参数
	if c.RouteParams != nil {
		for k := range c.RouteParams {
			delete(c.RouteParams, k)
		}
	}

	return nil
}

// Reset 重置控制器状态（优化控制器特性）
func (c *BaseController) Reset() {
	// 重置Context
	c.Ctx = nil

	// 重置控制器信息
	c.ControllerName = ""
	c.ActionName = ""
	c.RoutePattern = ""

	// 清理数据映射
	if c.Data != nil {
		for k := range c.Data {
			delete(c.Data, k)
		}
	}
	if c.RouteParams != nil {
		for k := range c.RouteParams {
			delete(c.RouteParams, k)
		}
	}
	if c.LayoutSections != nil {
		for k := range c.LayoutSections {
			delete(c.LayoutSections, k)
		}
	}
	
	// 重置执行状态
	c.ResetExecutionState()
}

// GetMiddleware 获取中间件列表（优化控制器特性）
func (c *BaseController) GetMiddleware() []string {
	if c.middlewareList == nil {
		return []string{}
	}
	// 返回副本，防止外部修改
	result := make([]string, len(c.middlewareList))
	copy(result, c.middlewareList)
	return result
}

// SetMiddleware 设置中间件列表
func (c *BaseController) SetMiddleware(middlewares []string) {
	c.middlewareList = make([]string, len(middlewares))
	copy(c.middlewareList, middlewares)
}

// AddMiddleware 添加中间件
func (c *BaseController) AddMiddleware(middleware string) {
	if c.middlewareList == nil {
		c.middlewareList = make([]string, 0)
	}
	c.middlewareList = append(c.middlewareList, middleware)
}

// EnableOptimization 启用优化特性
func (c *BaseController) EnableOptimization() {
	c.optimizationEnabled = true
}

// DisableOptimization 禁用优化特性
func (c *BaseController) DisableOptimization() {
	c.optimizationEnabled = false
}

// ShouldStopExecution 检查是否应该停止执行
// 这个方法主要用于框架内部，在路由处理中检查是否需要停止后续方法的执行
func (c *BaseController) ShouldStopExecution() bool {
	return c.shouldStopExecution
}

// ResetExecutionState 重置执行状态
// 这个方法主要用于框架内部，在处理新请求时重置状态
func (c *BaseController) ResetExecutionState() {
	c.shouldStopExecution = false
}

// IsOptimizationEnabled 检查是否启用优化特性
func (c *BaseController) IsOptimizationEnabled() bool {
	return c.optimizationEnabled
}

// ============= 全局实例访问器（避免循环导入） =============

// 这些函数通过反射或其他方式访问全局实例，避免直接导入 mvc 包造成循环导入

// globalManagerAccessor 全局管理器访问器接口
type globalManagerAccessor interface {
	GetSessionManager() *session.Manager
	GetCookieHelper() *cookie.Helper
	GetTemplateEngine() *view.TemplateEngine
	IsInitialized() bool
}

// 全局管理器访问器实例（将在运行时设置）
var globalAccessor globalManagerAccessor

// SetGlobalManagerAccessor 设置全局管理器访问器
// 这个函数将由 mvc 包在初始化时调用，以避免循环导入
func SetGlobalManagerAccessor(accessor globalManagerAccessor) {
	globalAccessor = accessor
}

// getGlobalSessionManagerIfAvailable 获取全局Session管理器（如果可用）
func getGlobalSessionManagerIfAvailable() *session.Manager {
	if globalAccessor != nil && globalAccessor.IsInitialized() {
		return globalAccessor.GetSessionManager()
	}
	return nil
}

// getGlobalCookieHelperIfAvailable 获取全局Cookie辅助器（如果可用）
func getGlobalCookieHelperIfAvailable() *cookie.Helper {
	if globalAccessor != nil && globalAccessor.IsInitialized() {
		return globalAccessor.GetCookieHelper()
	}
	return nil
}

// getGlobalTemplateEngineIfAvailable 获取全局模板引擎（如果可用）
func getGlobalTemplateEngineIfAvailable() *view.TemplateEngine {
	if globalAccessor != nil && globalAccessor.IsInitialized() {
		return globalAccessor.GetTemplateEngine()
	}
	return nil
}

// ============= 统一管理器集成方法 =============

// GetUnifiedManager 获取统一管理器
// 
// 返回当前控制器使用的统一管理器实例，用于高级功能操作。
//
// 返回：
//   - *unified.Manager: 统一管理器实例
//
// 示例：
//
//	manager := controller.GetUnifiedManager()
//	user := manager.GetCurrentUser(ctx)
func (c *BaseController) GetUnifiedManager() *unified.Manager {
	if c.unifiedManager == nil {
		c.unifiedManager = unified.GetManager()
	}
	return c.unifiedManager
}

// ============= 统一管理器便捷方法 =============

// SetContextData 设置上下文数据（使用统一管理器）
//
// 将数据存储到请求上下文中，可在整个请求生命周期内访问。
//
// 参数：
//   - key: 数据键
//   - value: 数据值
//
// 示例：
//
//	controller.SetContextData("user_id", 123)
//	controller.SetContextData("user_info", userStruct)
func (c *BaseController) SetContextData(key string, value any) {
	if c.GetUnifiedManager() != nil {
		c.unifiedManager.SetContextData(c.Ctx, key, value)
	}
}

// GetContextData 获取上下文数据（使用统一管理器）
//
// 从请求上下文中获取存储的数据。
//
// 参数：
//   - key: 数据键
//
// 返回：
//   - interface{}: 数据值，如果不存在返回nil
//
// 示例：
//
//	userID := controller.GetContextData("user_id")
//	if userID != nil {
//	    id := userID.(int)
//	}
func (c *BaseController) GetContextData(key string) any {
	if c.GetUnifiedManager() != nil {
		return c.unifiedManager.GetContextData(c.Ctx, key)
	}
	return nil
}

// GetTypedContextData 获取类型安全的上下文数据（使用统一管理器）
//
// 提供类型安全的数据获取，避免类型转换错误。
//
// 参数：
//   - key: 数据键
//   - target: 目标类型的零值（用于类型推断）
//
// 返回：
//   - interface{}: 数据值
//   - bool: 是否成功获取（键存在）
//
// 示例：
//
//	userID, ok := controller.GetTypedContextData("user_id", 0)
//	if ok {
//	    id := userID.(int)
//	    fmt.Printf("User ID: %d", id)
//	}
func (c *BaseController) GetTypedContextData(key string, target any) (any, bool) {
	if c.GetUnifiedManager() != nil {
		return c.unifiedManager.GetTypedContextData(c.Ctx, key, target)
	}
	return nil, false
}



// GenerateUnifiedCSRFToken 使用统一管理器生成CSRF令牌
//
// 参数：
//   - userID: 用户ID
//   - clientIP: 客户端IP地址
//
// 返回：
//   - string: CSRF令牌值
//   - error: 生成错误
//
// 示例：
//
//	token, err := controller.GenerateUnifiedCSRFToken("123", "192.168.1.1")
func (c *BaseController) GenerateUnifiedCSRFToken(userID, clientIP string) (string, error) {
	if c.GetUnifiedManager() != nil {
		token, err := c.unifiedManager.GenerateCSRFToken(userID, clientIP)
		if err != nil {
			return "", err
		}
		return token.Value, nil
	}
	return "", unified.ErrCSRFManagerNotInitialized
}

// ValidateUnifiedCSRFToken 使用统一管理器验证CSRF令牌
//
// 参数：
//   - tokenValue: 令牌值
//   - userID: 用户ID
//   - clientIP: 客户端IP地址
//
// 返回：
//   - bool: 验证结果
//   - error: 验证错误
//
// 示例：
//
//	valid, err := controller.ValidateUnifiedCSRFToken(token, "123", "192.168.1.1")
func (c *BaseController) ValidateUnifiedCSRFToken(tokenValue, userID, clientIP string) (bool, error) {
	if c.GetUnifiedManager() != nil {
		return c.unifiedManager.ValidateCSRFToken(tokenValue, userID, clientIP)
	}
	return false, unified.ErrCSRFManagerNotInitialized
}

// RenderUnifiedTemplate 使用统一管理器渲染模板
//
// 参数：
//   - templateName: 模板名称
//   - data: 模板数据
//
// 返回：
//   - string: 渲染结果
//   - error: 渲染错误
//
// 示例：
//
//	html, err := controller.RenderUnifiedTemplate("user/profile", userData)
func (c *BaseController) RenderUnifiedTemplate(templateName string, data any) (string, error) {
	if c.GetUnifiedManager() != nil {
		return c.unifiedManager.RenderTemplate(templateName, data)
	}
	return "", unified.ErrTemplateEngineNotInitialized
}

// ============= SSO 相关便捷方法 =============

// LoginUser 用户登录（使用统一管理器的SSO功能）
//
// 在用户成功登录后调用此函数，设置SSO会话和Cookie。
//
// 参数：
//   - userInfo: 用户信息
//   - rememberMe: 是否记住我
//
// 返回：
//   - error: 登录设置错误
//
// 示例：
//
//	err := controller.LoginUser(userInfo, true)
func (c *BaseController) LoginUser(userInfo *unified.UserInfo, rememberMe bool) error {
	manager := c.GetUnifiedManager()
	if manager == nil {
		return unified.ErrManagerNotInitialized
	}
	return unified.LoginUser(manager, c.Ctx, userInfo, rememberMe)
}

// LogoutUser 用户登出（使用统一管理器的SSO功能）
//
// 在用户登出时调用此函数，清除所有SSO相关数据。
//
// 示例：
//
//	controller.LogoutUser()
func (c *BaseController) LogoutUser() {
	manager := c.GetUnifiedManager()
	if manager != nil {
		unified.LogoutUser(manager, c.Ctx)
	}
}

// GetCurrentUser 获取当前认证用户信息
//
// 返回：
//   - *unified.UserInfo: 用户信息，如果未认证返回nil
//
// 示例：
//
//	user := controller.GetCurrentUser()
//	if user != nil {
//	    fmt.Printf("Welcome %s", user.Username)
//	}
func (c *BaseController) GetCurrentUser() *unified.UserInfo {
	manager := c.GetUnifiedManager()
	if manager != nil {
		return unified.GetCurrentUser(manager, c.Ctx)
	}
	return nil
}

// IsUserAuthenticated 检查用户是否已认证
//
// 返回：
//   - bool: 是否已认证
//
// 示例：
//
//	if controller.IsUserAuthenticated() {
//	    // 处理已认证用户的逻辑
//	}
func (c *BaseController) IsUserAuthenticated() bool {
	manager := c.GetUnifiedManager()
	if manager != nil {
		return unified.IsAuthenticated(manager, c.Ctx)
	}
	return false
}
