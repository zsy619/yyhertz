package core

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strings"
	"sync"

	hertzapp "github.com/cloudwego/hertz/pkg/app"

	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// ============= 路由冲突检测 =============

// 全局路由注册表，用于检测路径冲突
var (
	registeredRoutes   = make(map[string]string) // routeKey(method:path) -> controllerInfo
	routeRegistryMutex sync.RWMutex
)

// RouteInfo 路由信息结构
type RouteInfo struct {
	Method     string
	Path       string
	Controller string
	Action     string
}

// clearRouteRegistry 清空路由注册表（主要用于测试）
func clearRouteRegistry() {
	routeRegistryMutex.Lock()
	defer routeRegistryMutex.Unlock()
	registeredRoutes = make(map[string]string)
}

// isRouteRegistered 检查路由是否已注册
func isRouteRegistered(method, path string) (bool, string) {
	routeRegistryMutex.RLock()
	defer routeRegistryMutex.RUnlock()

	routeKey := strings.ToUpper(method) + ":" + path
	existing, exists := registeredRoutes[routeKey]
	return exists, existing
}

// registerRouteInfo 注册路由信息到注册表
func registerRouteInfo(method, path, controllerInfo string) {
	routeRegistryMutex.Lock()
	defer routeRegistryMutex.Unlock()

	routeKey := strings.ToUpper(method) + ":" + path
	registeredRoutes[routeKey] = controllerInfo
}

// ============= 路由注册方法 =============

// RouterAuto 自动注册多个控制器路由（根据控制器方法名自动推导路由）
func (app *App) RouterAuto(controllers ...IController) *App {
	if len(controllers) == 0 {
		return app
	}
	// app.LogInfof("RouterAuto called with %d controllers", len(controllers))
	return app.RouterAutoPrefix("", controllers...)
}

// RouterAutoPrefix 自动注册多个控制器路由，使用指定的路径前缀
func (app *App) RouterAutoPrefix(prefix string, ctrls ...IController) *App {
	if len(ctrls) == 0 {
		return app
	}
	app.LogInfof("RouterAutoPrefix processing %d controllers with prefix '%s'", len(ctrls), prefix)
	for i, ctrl := range ctrls {
		app.LogInfof("Processing controller %d/%d: %s", i+1, len(ctrls), ctrl.GetControllerName())
		app.registerAutoRoutes(prefix, ctrl)
	}
	return app
}

// Include 向后兼容的别名方法，自动注册多个控制器路由
func (app *App) Include(controllers ...IController) *App {
	return app.RouterAutoPrefix("", controllers...)
}

// Router 手动注册控制器路由
func (app *App) Router(ctrl IController, routePair bool, routes ...string) *App {
	return app.RouterPrefix("", ctrl, routePair, routes...)
}

// RouterMap 手动注册控制器路由（使用map格式）
func (app *App) RouterMap(ctrl IController, routes map[string]string) *App {
	return app.RouterPrefixMap("", ctrl, routes)
}

// RouterPrefix 手动注册控制器路由
func (app *App) RouterPrefix(prefix string, ctrl IController, routePair bool, routes ...string) *App {
	if len(routes) == 0 {
		return app
	}
	app.registerManualRoutes(prefix, ctrl, routePair, routes...)
	return app
}

// RouterPrefixMap 手动注册控制器路由（使用map格式）
func (app *App) RouterPrefixMap(prefix string, ctrl IController, routes map[string]string) *App {
	if len(routes) == 0 {
		return app
	}
	app.registerManualRoutesMap(prefix, ctrl, routes)
	return app
}

// ============= 向后兼容的别名方法 =============

// registerAutoRoutes 自动注册控制器路由
func (app *App) registerAutoRoutes(basePath string, controller IController) {
	// 确保控制器实例正确设置（提前初始化）
	if method := reflect.ValueOf(controller).MethodByName("SetControllerInstance"); method.IsValid() {
		method.Call([]reflect.Value{reflect.ValueOf(controller)})
	}

	// 使用反射获取控制器类型信息
	reflectVal := reflect.ValueOf(controller)
	rt := reflectVal.Type() // 获取指针类型的方法，而不是值类型

	// 从控制器名称推导基础路径
	controllerName := rt.Elem().Name() // 获取指针指向的类型名称
	if basePath == "" {
		basePath = "/"
	}

	// 添加调试信息（可选，生产环境可注释掉）
	app.LogInfof("Processing controller: %s, initial basePath: '%s'", controllerName, basePath)

	// 强制为每个控制器分配独立的路径空间，防止路径冲突
	for suffix := range ControllerNameSuffixReserved {
		if strings.HasSuffix(controllerName, suffix) {
			name := strings.TrimSuffix(controllerName, suffix)
			name = strings.ToLower(name)

			if basePath == "/" {
				basePath += name
			} else {
				basePath = path.Join(basePath, name)
			}

			app.LogInfof("Processing Controller %s mapped to basePath: '%s'", controllerName, basePath)
			break
		}
	}

	// 遍历所有公共方法
	for i := 0; i < rt.NumMethod(); i++ {
		method := rt.Method(i)
		methodName := method.Name

		// 跳过生命周期方法和BaseController方法
		if _, ok := ReservedMethods[methodName]; ok {
			continue
		}

		// 根据方法名前缀确定HTTP方法
		httpMethod := "ANY" // 默认ANY
		actionName := methodName

		switch {
		case strings.HasPrefix(methodName, "Get"):
			httpMethod = "GET"
			if methodName == "Get" || methodName == "Index" {
				actionName = "" // 特殊处理Get和Index方法
			} else {
				actionName = strings.TrimPrefix(methodName, "Get")
			}
		case strings.HasPrefix(methodName, "Post"):
			httpMethod = "POST"
			if methodName == "Post" || methodName == "Index" {
				actionName = "" // 特殊处理Post和Index方法
			} else {
				actionName = strings.TrimPrefix(methodName, "Get")
			}
		case strings.HasPrefix(methodName, "Post"):
			httpMethod = "POST"
			if methodName == "Post" || methodName == "Index" {
				actionName = "" // 特殊处理Post和Index方法
			} else {
				actionName = strings.TrimPrefix(methodName, "Post")
			}
		case strings.HasPrefix(methodName, "Put"):
			httpMethod = "PUT"
			if methodName == "Put" || methodName == "Index" {
				actionName = "" // 特殊处理Put和Index方法
			} else {
				actionName = strings.TrimPrefix(methodName, "Put")
			}
		case strings.HasPrefix(methodName, "Delete"):
			httpMethod = "DELETE"
			if methodName == "Delete" || methodName == "Index" {
				actionName = "" // 特殊处理Delete和Index方法
			} else {
				actionName = strings.TrimPrefix(methodName, "Delete")
			}
		case strings.HasPrefix(methodName, "Patch"):
			httpMethod = "PATCH"
			if methodName == "Patch" || methodName == "Index" {
				actionName = "" // 特殊处理Patch和Index方法
			} else {
				actionName = strings.TrimPrefix(methodName, "Patch")
			}
		case strings.HasPrefix(methodName, "Head"):
			httpMethod = "HEAD"
			if methodName == "Head" || methodName == "Index" {
				actionName = "" // 特殊处理Head和Index方法
			} else {
				actionName = strings.TrimPrefix(methodName, "Head")
			}
		case strings.HasPrefix(methodName, "Options"):
			httpMethod = "OPTIONS"
			if methodName == "Options" || methodName == "Index" {
				actionName = "" // 特殊处理Options和Index方法
			} else {
				actionName = strings.TrimPrefix(methodName, "Options")
			}
		}

		// 构建路由路径
		routePath := basePath
		if actionName != "" {
			if !strings.HasSuffix(routePath, "/") {
				routePath += "/"
			}
			routePath += strings.ToLower(actionName)
		}

		// 清理路径，移除重复的斜杠
		for strings.Contains(routePath, "//") {
			routePath = strings.ReplaceAll(routePath, "//", "/")
		}
		if routePath != "/" && strings.HasSuffix(routePath, "/") {
			routePath = strings.TrimSuffix(routePath, "/")
		}

		// 创建处理函数
		handler := app.createControllerHandler(controller, method)

		// 添加调试信息（可选，生产环境可注释掉）
		// app.LogInfof("Processing Registering auto route: method=%s, path='%s', controller=%s.%s", httpMethod, routePath, controller.GetControllerName(), method.Name)

		// 注册路由（带控制器信息）
		controllerInfo := fmt.Sprintf("%s.%s", controller.GetControllerName(), method.Name)
		app.registerRouteWithInfo(httpMethod, routePath, handler, controllerInfo)
	}
}

// registerManualRoutes 手动注册路由
func (app *App) registerManualRoutes(basePath string, controller IController, routePair bool, routes ...string) {
	t := reflect.TypeOf(controller)                       // 返回 *controllers.UserController
	controllerName := strings.TrimPrefix(t.String(), "*") // 得到 "controllers.UserController"
	for suffix := range ControllerNameSuffixReserved {
		if strings.HasSuffix(controllerName, suffix) {
			controllerName = strings.TrimSuffix(controllerName, suffix)
			break
		}
	}
	app.LogDebugf("Registering routes for controller: %s\n", controllerName)
	if routePair == false {
		for i := range routes {
			routeSpec := routes[i]
			// 解析路由规格: "GET:/path|method" 或 "/path" 或 "*:/path"
			httpMethod := "ANY"
			routePath := routeSpec
			methodName := ""

			if colonIndex := strings.Index(routeSpec, ":"); colonIndex != -1 {
				httpMethod = routeSpec[:colonIndex]
				if httpMethod == "*" { // 兼容旧格式的路由语法: *:path
					httpMethod = "ANY"
				}
				rp := routeSpec[colonIndex+1:]
				routePath = rp
				if pipeIndex := strings.Index(rp, "|"); pipeIndex != -1 {
					routePath = rp[:pipeIndex]
					methodName = rp[pipeIndex+1:]
				} else if routePath != "" && !strings.HasPrefix(routePath, "/") {
					// 如果routePath不是以/开头，可能是方法名
					methodName = routePath
					routePath = "" // 使用basePath作为路径
				} else {
					methodName = strings.TrimPrefix(routePath, "/")
				}
			} else {
				methodName = strings.TrimPrefix(routePath, "/")
			}

			routePath = strings.ToLower(routePath) // 确保路由路径小写

			// 确保路由路径以基础路径开头
			if !strings.HasPrefix(routePath, basePath) {
				routePath = basePath + routePath
			}

			// 获取控制器方法
			reflectVal := reflect.ValueOf(controller)
			method := reflectVal.MethodByName(methodName)

			if !method.IsValid() {
				app.LogErrorf("Method %s not found in controller", methodName)
				return
			}

			// 创建处理函数
			handler := app.createMethodHandler(controller, methodName)

			// 注册路由（带控制器信息）
			controllerInfo := fmt.Sprintf("%s.%s", app.getControllerName(controller), methodName)
			app.registerRouteWithInfo(httpMethod, routePath, handler, controllerInfo)
		}
	} else {
		for i := 0; i < len(routes); i += 2 {
			if i+1 >= len(routes) {
				break
			}

			methodName := routes[i]
			routeSpec := routes[i+1]

			// 解析路由规格: "GET:/path" 或 "/path" 或 "*:/path"
			httpMethod := "ANY"
			routePath := routeSpec

			if colonIndex := strings.Index(routeSpec, ":"); colonIndex != -1 {
				httpMethod = routeSpec[:colonIndex]
				if httpMethod == "*" { // 兼容旧格式的路由语法: *:path
					httpMethod = "ANY"
				}
				routePath = routeSpec[colonIndex+1:]
			} else if routePath == "" {
				routePath = methodName
			}

			routePath = strings.ToLower(routePath) // 确保路由路径小写

			// 确保路由路径以基础路径开头
			if !strings.HasPrefix(routePath, basePath) {
				routePath = basePath + routePath
			}

			// 获取控制器方法
			reflectVal := reflect.ValueOf(controller)
			method := reflectVal.MethodByName(methodName)

			if !method.IsValid() {
				app.LogErrorf("Method %s not found in controller", methodName)
				continue
			}

			// 创建处理函数
			handler := app.createMethodHandler(controller, methodName)

			// 注册路由（带控制器信息）
			controllerInfo := fmt.Sprintf("%s.%s", app.getControllerName(controller), methodName)
			app.registerRouteWithInfo(httpMethod, routePath, handler, controllerInfo)
		}
	}
}

func (app *App) registerManualRoutesMap(basePath string, controller IController, routes map[string]string) {
	t := reflect.TypeOf(controller)                       // 返回 *controllers.UserController
	controllerName := strings.TrimPrefix(t.String(), "*") // 得到 "controllers.UserController"
	// controllerName = strings.TrimSuffix(controllerName, "Controller")
	for suffix := range ControllerNameSuffixReserved {
		if strings.HasSuffix(controllerName, suffix) {
			controllerName = strings.TrimSuffix(controllerName, suffix)
			break
		}
	}
	fmt.Printf("Registering routes for controller: %s\n", controllerName)

	for methodName, routeSpec := range routes {
		// 解析路由规格: "GET:/path" 或 "/path" 或 "*:/path"
		httpMethod := "ANY"
		routePath := routeSpec

		if colonIndex := strings.Index(routeSpec, ":"); colonIndex != -1 {
			httpMethod = routeSpec[:colonIndex]
			if httpMethod == "*" { // 兼容旧格式的路由语法: *:path
				httpMethod = "ANY"
			}
			routePath = routeSpec[colonIndex+1:]
		}

		// 确保路由路径以基础路径开头
		if !strings.HasPrefix(routePath, basePath) {
			routePath = basePath + routePath
		}

		// 获取控制器方法
		reflectVal := reflect.ValueOf(controller)
		method := reflectVal.MethodByName(methodName)

		if !method.IsValid() {
			app.LogErrorf("Method %s not found in controller", methodName)
			continue
		}

		// 创建处理函数
		handler := app.createMethodHandler(controller, methodName)

		// 注册路由（带控制器信息）
		controllerInfo := fmt.Sprintf("%s.%s", app.getControllerName(controller), methodName)
		app.registerRouteWithInfo(httpMethod, routePath, handler, controllerInfo)
	}
}

// getControllerName 获取控制器名称
func (app *App) getControllerName(controller IController) string {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}
	controllerName := controllerType.Name()
	for suffix := range ControllerNameSuffixReserved {
		if strings.HasSuffix(controllerName, suffix) {
			controllerName = strings.TrimSuffix(controllerName, suffix)
			break
		}
	}
	return controllerName
}

// createControllerHandler 创建控制器处理函数
func (app *App) createControllerHandler(controller IController, method reflect.Method) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		// 确保控制器实例正确设置（关键修复）
		if method := reflect.ValueOf(controller).MethodByName("SetControllerInstance"); method.IsValid() {
			method.Call([]reflect.Value{reflect.ValueOf(controller)})
		}

		// 初始化增强上下文
		enhancedCtx := contextenhanced.NewContext(c)

		// 执行 BeforeStatic 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeStatic)
		if enhancedCtx.IsAborted() {
			return
		}

		// 执行 BeforeRouter 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeRouter)
		if enhancedCtx.IsAborted() {
			return
		}

		// 初始化控制器
		controllerName := controller.GetControllerName() // 使用修复后的方法
		methodName := method.Name
		controller.Init(enhancedCtx, controllerName, methodName, app)

		// 设置控制器上下文（如果控制器有Ctx字段）
		app.setControllerContext(controller, c)

		// 执行 BeforeExec 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeExec)
		if enhancedCtx.IsAborted() {
			return
		}

		// 执行前置处理
		controller.Prepare()

		// 执行具体方法
		methodValue := reflect.ValueOf(controller).MethodByName(method.Name)
		if methodValue.IsValid() {
			// 根据方法签名调用
			methodType := methodValue.Type()
			if methodType.NumIn() == 2 {
				// 方法签名: func(context.Context, *define.RequestContext)
				methodValue.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(c),
				})
			} else if methodType.NumIn() == 0 {
				// 方法签名: func()
				methodValue.Call([]reflect.Value{})
			}
		}

		// 执行 AfterExec 过滤器
		app.ExecuteFilters(enhancedCtx, AfterExec)

		// 执行后置处理
		controller.Finish()

		// 执行 FinishRouter 过滤器
		app.ExecuteFilters(enhancedCtx, FinishRouter)
	}
}

// createMethodHandler 创建方法处理函数
func (app *App) createMethodHandler(controller IController, methodName string) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		// 初始化增强上下文
		enhancedCtx := contextenhanced.NewContext(c)

		// 执行 BeforeStatic 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeStatic)
		if enhancedCtx.IsAborted() {
			return
		}

		// 执行 BeforeRouter 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeRouter)
		if enhancedCtx.IsAborted() {
			return
		}

		// 初始化控制器
		controllerName := app.getControllerName(controller)
		controller.Init(enhancedCtx, controllerName, methodName, app)

		// 设置控制器上下文
		app.setControllerContext(controller, c)

		// 执行 BeforeExec 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeExec)
		if enhancedCtx.IsAborted() {
			return
		}

		// 执行前置处理
		controller.Prepare()

		// 执行具体方法
		methodValue := reflect.ValueOf(controller).MethodByName(methodName)
		if methodValue.IsValid() {
			methodType := methodValue.Type()
			if methodType.NumIn() == 2 {
				methodValue.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(c),
				})
			} else if methodType.NumIn() == 0 {
				methodValue.Call([]reflect.Value{})
			}
		}

		// 执行 AfterExec 过滤器
		app.ExecuteFilters(enhancedCtx, AfterExec)

		// 执行后置处理
		controller.Finish()

		// 执行 FinishRouter 过滤器
		app.ExecuteFilters(enhancedCtx, FinishRouter)
	}
}

// setControllerContext 设置控制器上下文（重构后版本）
func (app *App) setControllerContext(controller IController, ctx *define.RequestContext) {
	// 创建增强的Context
	enhancedCtx := contextenhanced.NewContext(ctx)

	// 使用反射设置控制器的Ctx字段
	reflectVal := reflect.ValueOf(controller)
	if reflectVal.Kind() == reflect.Ptr {
		reflectVal = reflectVal.Elem()
	}

	// 查找Ctx字段
	if reflectVal.Kind() == reflect.Struct {
		ctxField := reflectVal.FieldByName("Ctx")
		if ctxField.IsValid() && ctxField.CanSet() {
			ctxField.Set(reflect.ValueOf(enhancedCtx))
		}

		// 也尝试设置BaseController字段的Ctx
		baseField := reflectVal.FieldByName("BaseController")
		if baseField.IsValid() && baseField.Kind() == reflect.Ptr {
			baseController := baseField.Elem()
			if baseController.IsValid() {
				baseCtxField := baseController.FieldByName("Ctx")
				if baseCtxField.IsValid() && baseCtxField.CanSet() {
					baseCtxField.Set(reflect.ValueOf(enhancedCtx))
				}
			}
		}
	}
}

// registerRoute 注册路由到应用（带冲突检测）
func (app *App) registerRoute(method, path string, handler define.HandlerFunc) {
	app.registerRouteWithInfo(method, path, handler, "unknown")
}

// registerRouteWithInfo 注册路由到应用（带控制器信息和冲突检测）
func (app *App) registerRouteWithInfo(method, path string, handler define.HandlerFunc, controllerInfo string) {
	// 检查参数有效性
	if app == nil {
		panic("app instance is nil")
	}
	if app.Hertz == nil {
		panic("app.Hertz is nil")
	}
	if method == "" {
		method = "GET" // 默认GET方法
	}
	if path == "" {
		path = "/" // 默认根路径
	}
	if handler == nil {
		panic("handler cannot be nil")
	}

	// 确保路径以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 检查路由冲突
	if exists, existing := isRouteRegistered(method, path); exists {
		app.LogWarnf("Route conflict detected: %s %s already registered by %s, skipping registration for %s",
			method, path, existing, controllerInfo)
		return // 跳过重复注册，防止panic
	}

	// 注册路由信息到注册表
	registerRouteInfo(method, path, controllerInfo)

	// 实际注册路由到Hertz
	switch strings.ToUpper(method) {
	case "GET":
		app.GET(path, handler)
	case "POST":
		app.POST(path, handler)
	case "PUT":
		app.PUT(path, handler)
	case "DELETE":
		app.DELETE(path, handler)
	case "PATCH":
		app.PATCH(path, handler)
	case "HEAD":
		app.HEAD(path, handler)
	case "OPTIONS":
		app.OPTIONS(path, handler)
	default:
		app.Any(path, handler)
	}

	app.LogDebugf("Route registered: %s %s -> %s", method, path, controllerInfo)
}

// ============= 特殊路由处理器（类似Beego） =============

// NoRoute 设置处理未找到路由的处理器（类似Beego风格）
func (app *App) NoRoute(handlers ...func(*contextenhanced.Context, *define.RequestContext)) {
	// 将我们的处理器转换为Hertz的HandlerFunc类型
	hertzHandlers := make([]hertzapp.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hertzHandlers[i] = hertzapp.HandlerFunc(func(ctx context.Context, c *hertzapp.RequestContext) {
			// 创建增强上下文并调用处理器
			enhancedCtx := contextenhanced.NewContext((*define.RequestContext)(c))
			handler(enhancedCtx, (*define.RequestContext)(c))
		})
	}

	// 设置Hertz的NoRoute处理器
	app.Hertz.NoRoute(hertzHandlers...)
}

// NoMethod 设置处理方法不允许的处理器（类似Beego风格）
func (app *App) NoMethod(handlers ...func(*contextenhanced.Context, *define.RequestContext)) {
	// 将我们的处理器转换为Hertz的HandlerFunc类型
	hertzHandlers := make([]hertzapp.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hertzHandlers[i] = hertzapp.HandlerFunc(func(ctx context.Context, c *hertzapp.RequestContext) {
			// 创建增强上下文并调用处理器
			enhancedCtx := contextenhanced.NewContext((*define.RequestContext)(c))
			handler(enhancedCtx, (*define.RequestContext)(c))
		})
	}

	// 设置Hertz的NoMethod处理器
	app.Hertz.NoMethod(hertzHandlers...)
}
