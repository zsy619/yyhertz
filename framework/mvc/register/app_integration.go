// Package register App与ControllerRegister集成
// 实现Beego风格的应用与控制器注册器的深度集成
package register

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// AppWithControllerRegister 集成了ControllerRegister的应用
type AppWithControllerRegister struct {
	*core.App                              // 嵌入App
	ControllerRegister *ControllerRegister // 控制器注册器

	// 扩展配置
	autoRouting      bool                  // 是否启用自动路由
	controllerPrefix string                // 控制器路由前缀
	routeFilters     []GlobalFilter        // 全局路由过滤器
	namespaces       map[string]*Namespace // 命名空间管理
	errorHandlers    map[int]ErrorHandler  // 错误处理器

	// 性能优化
	routeCache       map[string]*ControllerInfo // 路由缓存
	enableRouteCache bool                       // 是否启用路由缓存
	mu               sync.RWMutex               // 读写锁
}

// GlobalFilter 全局过滤器
type GlobalFilter struct {
	Name     string
	Pattern  string
	Position int
	Filter   FilterFunc
}

// Namespace 命名空间
type Namespace struct {
	Name               string
	Prefix             string
	ControllerRegister *ControllerRegister
	Filters            []FilterFunc
	Conditions         []Condition
}

// Condition 命名空间条件
type Condition struct {
	Type  string // host, method, header等
	Key   string
	Value any
}

// ErrorHandler 错误处理器
type ErrorHandler func(*contextenhanced.Context, error)

// ControllerRegistrationOptions 控制器注册选项
type ControllerRegistrationOptions struct {
	Pattern         string            // 路由模式
	Methods         []string          // 允许的HTTP方法
	Filters         []FilterFunc      // 控制器级过滤器
	Namespace       string            // 命名空间
	EnableAutoRoute bool              // 是否启用自动路由
	RoutePrefix     string            // 路由前缀
	Middleware      []string          // 中间件名称
	ActionMethods   map[string]string // 动作方法映射
}

// NewAppWithControllerRegister 创建集成ControllerRegister的应用
func NewAppWithControllerRegister() *AppWithControllerRegister {
	app := &AppWithControllerRegister{
		App:                core.NewApp(),
		ControllerRegister: NewControllerRegister(),
		autoRouting:        true,
		controllerPrefix:   "",
		routeFilters:       make([]GlobalFilter, 0),
		namespaces:         make(map[string]*Namespace),
		errorHandlers:      make(map[int]ErrorHandler),
		routeCache:         make(map[string]*ControllerInfo),
		enableRouteCache:   true,
	}

	// 集成ControllerRegister到Hertz应用
	app.integrateControllerRegister()

	return app
}

// RegisterController 核心控制器注册函数（仿照Beego实现）
func (app *AppWithControllerRegister) RegisterController(pattern string, ctrl core.IController, options ...*ControllerRegistrationOptions) error {
	// 合并选项
	opts := app.mergeRegistrationOptions(options...)

	// 验证控制器
	if err := app.validateController(ctrl); err != nil {
		return fmt.Errorf("controller validation failed: %v", err)
	}

	// 应用命名空间
	if opts.Namespace != "" {
		pattern = app.applyNamespace(opts.Namespace, pattern)
	}

	// 应用路由前缀
	if opts.RoutePrefix != "" {
		pattern = app.applyRoutePrefix(opts.RoutePrefix, pattern)
	}

	// 注册控制器到ControllerRegister
	if err := app.registerToControllerRegister(pattern, ctrl, opts); err != nil {
		return fmt.Errorf("failed to register controller: %v", err)
	}

	// 应用过滤器
	app.applyControllerFilters(pattern, opts.Filters)

	// 如果启用自动路由，还要注册自动路由
	if opts.EnableAutoRoute && app.autoRouting {
		app.registerAutoRoutes(ctrl, opts)
	}

	// 缓存路由信息
	if app.enableRouteCache {
		app.cacheRouteInfo(pattern, ctrl, opts)
	}

	config.Infof("Controller registered successfully: %s -> %T", pattern, ctrl)
	return nil
}

// RegisterControllerAuto 自动注册控制器（根据控制器名称自动生成路由）
func (app *AppWithControllerRegister) RegisterControllerAuto(controller core.IController, options ...*ControllerRegistrationOptions) error {
	// 自动生成路由模式
	pattern := app.generateAutoPattern(controller)

	// 设置自动路由选项
	opts := app.mergeRegistrationOptions(options...)
	opts.EnableAutoRoute = true

	return app.RegisterController(pattern, controller, opts)
}

// RegisterControllerWithNamespace 在指定命名空间中注册控制器
func (app *AppWithControllerRegister) RegisterControllerWithNamespace(namespace, pattern string, controller core.IController, options ...*ControllerRegistrationOptions) error {
	opts := app.mergeRegistrationOptions(options...)
	opts.Namespace = namespace

	return app.RegisterController(pattern, controller, opts)
}

// RegisterRESTController 注册RESTful控制器
func (app *AppWithControllerRegister) RegisterRESTController(resourceName string, controller core.IController, options ...*ControllerRegistrationOptions) error {
	opts := app.mergeRegistrationOptions(options...)

	// 设置RESTful路由
	restRoutes := []struct {
		method  string
		pattern string
	}{
		{"GET", "/" + resourceName},             // 列表
		{"POST", "/" + resourceName},            // 创建
		{"GET", "/" + resourceName + "/:id"},    // 详情
		{"PUT", "/" + resourceName + "/:id"},    // 更新
		{"DELETE", "/" + resourceName + "/:id"}, // 删除
	}

	for _, route := range restRoutes {
		opts.Methods = []string{route.method}
		if err := app.RegisterController(route.pattern, controller, opts); err != nil {
			return fmt.Errorf("failed to register REST route %s %s: %v", route.method, route.pattern, err)
		}
	}

	return nil
}

// BatchRegisterControllers 批量注册控制器
func (app *AppWithControllerRegister) BatchRegisterControllers(controllers map[string]core.IController, options ...*ControllerRegistrationOptions) error {
	opts := app.mergeRegistrationOptions(options...)

	for pattern, controller := range controllers {
		if err := app.RegisterController(pattern, controller, opts); err != nil {
			return fmt.Errorf("failed to register controller %s: %v", pattern, err)
		}
	}

	return nil
}

// CreateNamespace 创建命名空间
func (app *AppWithControllerRegister) CreateNamespace(name, prefix string) *Namespace {
	app.mu.Lock()
	defer app.mu.Unlock()

	namespace := &Namespace{
		Name:               name,
		Prefix:             prefix,
		ControllerRegister: NewControllerRegister(),
		Filters:            make([]FilterFunc, 0),
		Conditions:         make([]Condition, 0),
	}

	app.namespaces[name] = namespace
	return namespace
}

// AddGlobalFilter 添加全局过滤器
func (app *AppWithControllerRegister) AddGlobalFilter(name, pattern string, position int, filter FilterFunc) {
	app.mu.Lock()
	defer app.mu.Unlock()

	globalFilter := GlobalFilter{
		Name:     name,
		Pattern:  pattern,
		Position: position,
		Filter:   filter,
	}

	app.routeFilters = append(app.routeFilters, globalFilter)
	app.ControllerRegister.InsertFilter(pattern, position, filter)
}

// SetErrorHandler 设置错误处理器
func (app *AppWithControllerRegister) SetErrorHandler(statusCode int, handler ErrorHandler) {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.errorHandlers[statusCode] = handler
}

// ============= 内部实现方法 =============

// integrateControllerRegister 集成ControllerRegister到Hertz应用（简化版）
func (app *AppWithControllerRegister) integrateControllerRegister() {
	// 将ControllerRegister作为通用处理器注册到Hertz（使用适配器）
	app.Any("/*path", func(ctx context.Context, c *core.RequestContext) {
		// 创建增强的Context
		enhancedCtx := contextenhanced.NewContext(c)
		app.ControllerRegister.ServeHTTP(enhancedCtx)
	})

	// 添加默认的全局过滤器
	app.addDefaultFilters()

	// 设置默认错误处理器
	app.setDefaultErrorHandlers()
}

// mergeRegistrationOptions 合并注册选项
func (app *AppWithControllerRegister) mergeRegistrationOptions(options ...*ControllerRegistrationOptions) *ControllerRegistrationOptions {
	opts := &ControllerRegistrationOptions{
		Methods:         []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		Filters:         make([]FilterFunc, 0),
		EnableAutoRoute: app.autoRouting,
		RoutePrefix:     app.controllerPrefix,
		ActionMethods:   make(map[string]string),
	}

	// 合并传入的选项
	for _, option := range options {
		if option != nil {
			if option.Pattern != "" {
				opts.Pattern = option.Pattern
			}
			if len(option.Methods) > 0 {
				opts.Methods = option.Methods
			}
			if len(option.Filters) > 0 {
				opts.Filters = append(opts.Filters, option.Filters...)
			}
			if option.Namespace != "" {
				opts.Namespace = option.Namespace
			}
			if option.RoutePrefix != "" {
				opts.RoutePrefix = option.RoutePrefix
			}
			if len(option.ActionMethods) > 0 {
				for k, v := range option.ActionMethods {
					opts.ActionMethods[k] = v
				}
			}
		}
	}

	return opts
}

// validateController 验证控制器
func (app *AppWithControllerRegister) validateController(controller core.IController) error {
	if controller == nil {
		return fmt.Errorf("controller cannot be nil")
	}

	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() != reflect.Ptr {
		return fmt.Errorf("controller must be a pointer")
	}

	// 检查是否实现了IController接口
	if !controllerType.Implements(reflect.TypeOf((*core.IController)(nil)).Elem()) {
		return fmt.Errorf("controller must implement IController interface")
	}

	return nil
}

// applyNamespace 应用命名空间
func (app *AppWithControllerRegister) applyNamespace(namespace, pattern string) string {
	app.mu.RLock()
	defer app.mu.RUnlock()

	if ns, exists := app.namespaces[namespace]; exists {
		return "/" + strings.Trim(ns.Prefix, "/") + "/" + strings.Trim(pattern, "/")
	}

	return pattern
}

// applyRoutePrefix 应用路由前缀
func (app *AppWithControllerRegister) applyRoutePrefix(prefix, pattern string) string {
	if prefix == "" {
		return pattern
	}
	return "/" + strings.Trim(prefix, "/") + "/" + strings.Trim(pattern, "/")
}

// registerToControllerRegister 注册到ControllerRegister
func (app *AppWithControllerRegister) registerToControllerRegister(pattern string, controller core.IController, opts *ControllerRegistrationOptions) error {
	// 根据选项决定使用哪种注册方式
	if len(opts.Methods) == 0 || (len(opts.Methods) == 1 && opts.Methods[0] == "*") {
		// 注册所有方法
		app.ControllerRegister.Add(pattern, controller)
	} else {
		// 注册特定方法
		app.ControllerRegister.Add(pattern, controller, opts.Methods...)
	}

	return nil
}

// applyControllerFilters 应用控制器过滤器
func (app *AppWithControllerRegister) applyControllerFilters(pattern string, filters []FilterFunc) {
	for _, filter := range filters {
		app.ControllerRegister.InsertFilter(pattern, BeforeExec, filter)
	}
}

// registerAutoRoutes 注册自动路由
func (app *AppWithControllerRegister) registerAutoRoutes(controller core.IController, opts *ControllerRegistrationOptions) {
	// 自动发现控制器方法并注册路由
	controllerType := reflect.TypeOf(controller)

	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)

		// 检查是否是有效的HTTP方法
		if app.isValidHTTPMethod(method.Name) {
			httpMethod, actionName := app.parseMethodName(method.Name)
			pattern := app.generateMethodPattern(controller, actionName, opts)

			// 注册方法路由
			app.ControllerRegister.AddMethod(httpMethod, pattern, controller)
		}
	}
}

// generateAutoPattern 生成自动路由模式
func (app *AppWithControllerRegister) generateAutoPattern(controller core.IController) string {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	controllerName := strings.ToLower(controllerType.Name())
	if strings.HasSuffix(controllerName, "controller") {
		controllerName = controllerName[:len(controllerName)-10]
	}

	return "/" + controllerName + "/*"
}

// cacheRouteInfo 缓存路由信息
func (app *AppWithControllerRegister) cacheRouteInfo(pattern string, controller core.IController, opts *ControllerRegistrationOptions) {
	app.mu.Lock()
	defer app.mu.Unlock()

	// 创建缓存项
	info := &ControllerInfo{
		pattern:        pattern,
		controllerType: reflect.TypeOf(controller),
		methods:        make(map[string]*MethodInfo),
	}

	app.routeCache[pattern] = info
}

// addDefaultFilters 添加默认过滤器
func (app *AppWithControllerRegister) addDefaultFilters() {
	// 添加日志过滤器
	app.ControllerRegister.InsertFilter("/*", BeforeRouter, LoggingFilter)

	// 添加CORS过滤器
	app.ControllerRegister.InsertFilter("/*", BeforeRouter, CORSFilter)

	// 添加安全过滤器
	app.ControllerRegister.InsertFilter("/*", BeforeRouter, SecurityFilter)
}

// setDefaultErrorHandlers 设置默认错误处理器
func (app *AppWithControllerRegister) setDefaultErrorHandlers() {
	// 404错误处理器
	app.SetErrorHandler(404, func(ctx *contextenhanced.Context, err error) {
		ctx.Output.SetStatus(404)
		ctx.Output.JSON(map[string]any{
			"error": "Not Found",
			"code":  404,
			"path":  string(ctx.RequestContext.URI().Path()),
		}, false, true)
	})

	// 500错误处理器
	app.SetErrorHandler(500, func(ctx *contextenhanced.Context, err error) {
		ctx.Output.SetStatus(500)
		ctx.Output.JSON(map[string]any{
			"error": "Internal Server Error",
			"code":  500,
			"msg":   err.Error(),
		}, false, true)
	})
}

// isValidHTTPMethod 检查是否是有效的HTTP方法
func (app *AppWithControllerRegister) isValidHTTPMethod(methodName string) bool {
	httpMethods := []string{"Get", "Post", "Put", "Delete", "Head", "Options", "Patch"}

	for _, httpMethod := range httpMethods {
		if strings.HasPrefix(methodName, httpMethod) {
			return true
		}
	}

	return false
}

// parseMethodName 解析方法名
func (app *AppWithControllerRegister) parseMethodName(methodName string) (httpMethod, actionName string) {
	httpMethods := []string{"Get", "Post", "Put", "Delete", "Head", "Options", "Patch"}

	for _, method := range httpMethods {
		if strings.HasPrefix(methodName, method) {
			httpMethod = strings.ToUpper(method)
			actionName = strings.ToLower(methodName[len(method):])
			if actionName == "" {
				actionName = "index"
			}
			return
		}
	}

	return "GET", "index"
}

// generateMethodPattern 生成方法路由模式
func (app *AppWithControllerRegister) generateMethodPattern(controller core.IController, actionName string, opts *ControllerRegistrationOptions) string {
	basePattern := app.generateAutoPattern(controller)

	if actionName == "index" {
		return strings.TrimSuffix(basePattern, "/*")
	}

	return strings.TrimSuffix(basePattern, "/*") + "/" + actionName
}

// ============= 便捷方法 =============

// Router 获取ControllerRegister（兼容性方法）
func (app *AppWithControllerRegister) Router() *ControllerRegister {
	return app.ControllerRegister
}

// EnableAutoRouting 启用/禁用自动路由
func (app *AppWithControllerRegister) EnableAutoRouting(enable bool) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.autoRouting = enable
}

// SetControllerPrefix 设置控制器路由前缀
func (app *AppWithControllerRegister) SetControllerPrefix(prefix string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.controllerPrefix = prefix
}

// EnableRouteCache 启用/禁用路由缓存
func (app *AppWithControllerRegister) EnableRouteCache(enable bool) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.enableRouteCache = enable
}

// GetRouteInfo 获取路由信息
func (app *AppWithControllerRegister) GetRouteInfo() map[string]any {
	app.mu.RLock()
	defer app.mu.RUnlock()

	return map[string]any{
		"routes":         app.ControllerRegister.ListRoutes(),
		"request_count":  app.ControllerRegister.GetRequestCount(),
		"route_count":    app.ControllerRegister.GetRouteCount(),
		"cache_enabled":  app.enableRouteCache,
		"cached_routes":  len(app.routeCache),
		"namespaces":     len(app.namespaces),
		"global_filters": len(app.routeFilters),
	}
}
