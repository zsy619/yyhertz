package yyhertz

import (
	"context"
	"reflect"
	"strings"
)

// RegisterController 注册控制器到指定路径
func (app *App) RegisterController(basePath string, controller IController) {
	reflectVal := reflect.ValueOf(controller)
	rt := reflect.Indirect(reflectVal).Type()

	for i := 0; i < rt.NumMethod(); i++ {
		method := rt.Method(i)
		methodName := method.Name

		// 跳过生命周期方法
		if methodName == "Init" || methodName == "Prepare" || methodName == "Finish" {
			continue
		}

		// 根据方法名前缀确定HTTP方法
		httpMethod := "ANY"
		switch {
		case strings.HasPrefix(methodName, "Get"):
			httpMethod = "GET"
		case strings.HasPrefix(methodName, "Post"):
			httpMethod = "POST"
		case strings.HasPrefix(methodName, "Put"):
			httpMethod = "PUT"
		case strings.HasPrefix(methodName, "Delete"):
			httpMethod = "DELETE"
		case strings.HasPrefix(methodName, "Patch"):
			httpMethod = "PATCH"
		case strings.HasPrefix(methodName, "Head"):
			httpMethod = "HEAD"
		case strings.HasPrefix(methodName, "Options"):
			httpMethod = "OPTIONS"
		case strings.HasPrefix(methodName, "Connect"):
			httpMethod = "CONNECT"
		case strings.HasPrefix(methodName, "Trace"):
			httpMethod = "TRACE"
		}

		// 提取Action名称
		actionName := strings.ToLower(strings.TrimPrefix(methodName, httpMethod))
		if actionName == "" {
			actionName = strings.ToLower(methodName)
		}

		// 构建路由路径
		routePath := basePath + "/" + actionName
		app.createHandler(httpMethod, routePath, rt, methodName)
	}
}

// createHandler 创建路由处理器
func (app *App) createHandler(httpMethod, routePath string, controllerType reflect.Type, methodName string) {
	handler := HandlerFunc(func(c context.Context, ctx *RequestContext) {
		app.executeControllerMethod(controllerType, methodName, ctx)
	})

	// 根据HTTP方法注册路由
	switch httpMethod {
	case "GET":
		app.Hertz.GET(routePath, handler)
	case "POST":
		app.Hertz.POST(routePath, handler)
	case "PUT":
		app.Hertz.PUT(routePath, handler)
	case "DELETE":
		app.Hertz.DELETE(routePath, handler)
	case "PATCH":
		app.Hertz.PATCH(routePath, handler)
	case "HEAD":
		app.Hertz.HEAD(routePath, handler)
	case "OPTIONS":
		app.Hertz.OPTIONS(routePath, handler)
	default:
		app.Hertz.Any(routePath, handler)
	}
}

// executeControllerMethod 执行控制器方法
func (app *App) executeControllerMethod(controllerType reflect.Type, methodName string, ctx *RequestContext) {
	// 创建控制器实例
	vc := reflect.New(controllerType)
	var execController IController
	execController = vc.Elem().Addr().Interface().(IController)

	// 设置BaseController字段
	field := vc.Elem().FieldByName("BaseController")
	if field.IsValid() && field.CanSet() {
		baseCtrl := BaseController{
			Ctx:        ctx,
			ViewPath:   app.ViewPath,
			LayoutPath: "views/layout",
			Data:       make(map[string]any),
		}
		field.Set(reflect.ValueOf(baseCtrl))
	}

	// 执行控制器生命周期
	execController.Init()
	execController.Prepare()

	// 调用具体的Action方法
	method := vc.MethodByName(methodName)
	if method.IsValid() {
		method.Call(nil)
	}

	execController.Finish()
}

// Include 批量注册控制器（自动路径）
func (app *App) Include(controllers ...IController) {
	for _, c := range controllers {
		reflectVal := reflect.ValueOf(c)
		rt := reflect.Indirect(reflectVal).Type()
		controllerName := strings.TrimSuffix(rt.Name(), "Controller")
		controllerName = strings.ToLower(controllerName)

		pattern := "/" + controllerName
		app.RegisterController(pattern, c)
	}
}

// Router 手动注册控制器路由映射
// 用法: app.Router("/base", controller, "MethodName", "HTTP_METHOD:/path", ...)
func (app *App) Router(basePath string, controller IController, routes ...string) {
	if len(routes)%2 != 0 {
		panic("Router: routes must be in pairs of methodName and routePattern")
	}

	reflectVal := reflect.ValueOf(controller)
	rt := reflect.Indirect(reflectVal).Type()

	for i := 0; i < len(routes); i += 2 {
		methodName := routes[i]
		routePattern := routes[i+1]

		// 解析路由模式: "GET:/path" 或 "POST:/api/users"
		parts := strings.SplitN(routePattern, ":", 2)
		if len(parts) != 2 {
			panic("Router: invalid route pattern, expected 'METHOD:/path', got: " + routePattern)
		}

		httpMethod := strings.ToUpper(parts[0])
		routePath := parts[1]

		// 验证方法是否存在
		method := reflectVal.MethodByName(methodName)
		if !method.IsValid() {
			panic("Router: method '" + methodName + "' not found in controller " + rt.Name())
		}

		// 创建并注册处理器
		app.createHandler(httpMethod, routePath, rt, methodName)
	}
}
