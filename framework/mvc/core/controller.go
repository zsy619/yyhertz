package core

import (
	"html/template"

	"github.com/cloudwego/hertz/pkg/app"

	context "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/session"
	templatemanager "github.com/zsy619/yyhertz/framework/template"
	"github.com/zsy619/yyhertz/framework/view"
)

// BaseController 基础控制器结构
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

	// ============= 优化控制器特性 =============

	// 优化功能控制
	optimizationEnabled bool     // 是否启用优化特性
	middlewareList      []string // 中间件列表，支持GetMiddleware()

	// 内部控制字段
	initialized bool // 控制器名称是否已初始化（内部使用）
}

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

		// 辅助工具
		cookieHelper:   cookie.NewHelper(cookie.DefaultConfig()),
		sessionHelper:  session.NewManager(session.DefaultConfig()),
		templateEngine: templatemanager.GetTemplateManager().GetEngine(),
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

// IsOptimizationEnabled 检查是否启用优化特性
func (c *BaseController) IsOptimizationEnabled() bool {
	return c.optimizationEnabled
}
