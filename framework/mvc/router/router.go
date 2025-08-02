package router

import (
	"context"
	"reflect"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"

	contextenhanced "github.com/zsy619/yyhertz/framework/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// Router 路由器
type Router struct {
	app *core.App // 应用实例
}

// NewRouter 创建路由器
func NewRouter(app *core.App) *Router {
	return &Router{
		app: app,
	}
}

// RegisterController 注册控制器到指定路径
func (r *Router) RegisterController(basePath string, ctrl core.IController) {
	// 确保控制器实例正确设置（提前初始化）
	if method := reflect.ValueOf(ctrl).MethodByName("SetControllerInstance"); method.IsValid() {
		method.Call([]reflect.Value{reflect.ValueOf(ctrl)})
	}

	reflectVal := reflect.ValueOf(ctrl)
	rt := reflect.Indirect(reflectVal).Type()

	for i := 0; i < rt.NumMethod(); i++ {
		method := rt.Method(i)
		methodName := method.Name

		// 跳过保留方法
		if core.ReservedMethods[methodName] {
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
			actionName = "index"
		}

		// 构建路径
		path := basePath
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		if actionName != "index" {
			path += actionName
		}

		// 创建处理函数
		handler := r.createHandler(ctrl, method)

		// 注册路由（这里需要根据实际的app类型来调用相应的方法）
		r.registerRoute(httpMethod, path, handler)
	}
}

// createHandler 创建处理函数
func (r *Router) createHandler(ctrl core.IController, method reflect.Method) core.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建增强的Context
		enhancedCtx := contextenhanced.NewContext(c)

		// 确保控制器实例正确设置（关键修复）
		// 直接调用方法，因为所有控制器都嵌入了BaseController
		if method := reflect.ValueOf(ctrl).MethodByName("SetControllerInstance"); method.IsValid() {
			method.Call([]reflect.Value{reflect.ValueOf(ctrl)})
		}

		// 提取控制器和动作名称
		controllerName := ctrl.GetControllerName()
		actionName := method.Name

		// 初始化控制器
		ctrl.Init(enhancedCtx, controllerName, actionName, r.app)

		// 执行前置处理
		ctrl.Prepare()

		// 执行具体方法
		methodValue := reflect.ValueOf(ctrl).MethodByName(method.Name)
		if methodValue.IsValid() {
			// 根据方法签名调用
			methodType := methodValue.Type()
			if methodType.NumIn() == 2 {
				// 方法签名: func(context.Context, *RequestContext)
				methodValue.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(c),
				})
			} else if methodType.NumIn() == 0 {
				// 方法签名: func()
				methodValue.Call([]reflect.Value{})
			}
		}

		// 执行后置处理
		ctrl.Finish()
	}
}

// registerRoute 注册路由到应用
func (r *Router) registerRoute(method, path string, handler core.HandlerFunc) {
	// 这里需要根据实际的app类型来调用相应的方法
	// 使用反射或类型断言来调用app的路由注册方法
	appValue := reflect.ValueOf(r.app)

	var routeMethod reflect.Value
	switch strings.ToUpper(method) {
	case "GET":
		routeMethod = appValue.MethodByName("GET")
	case "POST":
		routeMethod = appValue.MethodByName("POST")
	case "PUT":
		routeMethod = appValue.MethodByName("PUT")
	case "DELETE":
		routeMethod = appValue.MethodByName("DELETE")
	case "PATCH":
		routeMethod = appValue.MethodByName("PATCH")
	case "HEAD":
		routeMethod = appValue.MethodByName("HEAD")
	case "OPTIONS":
		routeMethod = appValue.MethodByName("OPTIONS")
	case "ANY":
		routeMethod = appValue.MethodByName("Any")
	default:
		routeMethod = appValue.MethodByName("Any")
	}

	if routeMethod.IsValid() {
		routeMethod.Call([]reflect.Value{
			reflect.ValueOf(path),
			reflect.ValueOf(handler),
		})
	}
}

// RegisterRoutes 批量注册路由的辅助函数
func (r *Router) RegisterRoutes(routes map[string]core.IController) {
	for basePath, ctrl := range routes {
		r.RegisterController(basePath, ctrl)
	}
}
