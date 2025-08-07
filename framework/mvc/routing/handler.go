package routing

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// RequestHandler 统一请求处理器（从annotation和comment包提取）
type RequestHandler struct {
	paramBinder *ParamBinder
	app         *core.App
	engine      *route.Engine
}

// NewRequestHandler 创建请求处理器
func NewRequestHandler(app *core.App, engine *route.Engine) *RequestHandler {
	return &RequestHandler{
		paramBinder: NewParamBinder(),
		app:         app,
		engine:      engine,
	}
}

// CreateHandler 创建处理函数（统一从annotation和comment包提取）
func (rh *RequestHandler) CreateHandler(route *RouteInfo) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建控制器实例
		var controllerValue reflect.Value
		var controller interface{}
		
		if route.ControllerType.Kind() == reflect.Ptr {
			// 如果是指针类型，创建元素类型的实例
			controllerValue = reflect.New(route.ControllerType.Elem())
			controller = controllerValue.Interface()
		} else {
			// 如果是值类型，创建指针类型的实例
			controllerValue = reflect.New(route.ControllerType)
			controller = controllerValue.Interface()
		}

		// 如果实现了IController接口，初始化BaseController
		if iController, ok := controller.(core.IController); ok {
			// 创建统一的Context
			enhancedCtx := &contextenhanced.Context{
				RequestContext: c,
			}

			// 初始化控制器
			iController.Init(enhancedCtx, route.TypeName, route.MethodName, rh.app)

			// 调用Prepare方法
			iController.Prepare()
		}

		// 验证参数
		if err := rh.paramBinder.ValidateParams(route.Params, c); err != nil {
			rh.handleError(c, 400, err)
			return
		}

		// 准备方法参数
		methodInfo := &MethodInfo{
			MethodName: route.MethodName,
			Params:     route.Params,
		}
		args, err := rh.paramBinder.PrepareMethodArgs(methodInfo, c, controllerValue)
		if err != nil {
			rh.handleError(c, 400, fmt.Errorf("failed to prepare method arguments: %w", err))
			return
		}

		// 调用控制器方法
		method := controllerValue.MethodByName(route.MethodName)
		if !method.IsValid() {
			rh.handleError(c, 500, fmt.Errorf("method '%s' not found in controller", route.MethodName))
			return
		}

		results := method.Call(args)

		// 处理方法返回值
		rh.handleMethodResults(c, results)

		// 如果实现了IController接口，调用Finish方法
		if iController, ok := controller.(core.IController); ok {
			iController.Finish()
		}
	}
}

// RegisterRoute 注册路由到引擎（统一从annotation和comment包提取）
func (rh *RequestHandler) RegisterRoute(route *RouteInfo) error {
	// 验证路由信息
	if err := rh.validateRoute(route); err != nil {
		return err
	}

	// 创建处理函数
	handler := rh.CreateHandler(route)

	// 根据HTTP方法注册路由
	return rh.registerToEngine(route.HTTPMethod, route.Path, handler)
}

// validateRoute 验证路由信息
func (rh *RequestHandler) validateRoute(route *RouteInfo) error {
	if route == nil {
		return &RouteError{
			Type:    ErrorTypeRegistrationError,
			Message: "route info is nil",
		}
	}

	if err := ValidateHTTPMethod(route.HTTPMethod); err != nil {
		return err
	}

	if err := ValidatePath(route.Path); err != nil {
		return err
	}

	if err := ValidateControllerType(route.ControllerType); err != nil {
		return err
	}

	if err := ValidateMethodName(route.ControllerType, route.MethodName); err != nil {
		return err
	}

	return nil
}

// registerToEngine 注册到Hertz引擎
func (rh *RequestHandler) registerToEngine(httpMethod, path string, handler app.HandlerFunc) error {
	switch strings.ToUpper(httpMethod) {
	case "GET":
		rh.engine.GET(path, handler)
	case "POST":
		rh.engine.POST(path, handler)
	case "PUT":
		rh.engine.PUT(path, handler)
	case "DELETE":
		rh.engine.DELETE(path, handler)
	case "PATCH":
		rh.engine.PATCH(path, handler)
	case "HEAD":
		rh.engine.HEAD(path, handler)
	case "OPTIONS":
		rh.engine.OPTIONS(path, handler)
	case "ANY":
		rh.engine.Any(path, handler)
	default:
		return &RouteError{
			Type:    ErrorTypeInvalidHTTPMethod,
			Message: "unsupported HTTP method: " + httpMethod,
		}
	}

	return nil
}

// handleMethodResults 处理方法返回值（统一从annotation和comment包提取）
func (rh *RequestHandler) handleMethodResults(c *app.RequestContext, results []reflect.Value) {
	if len(results) == 0 {
		return
	}

	// 如果最后一个返回值是error类型
	if len(results) >= 1 && results[len(results)-1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		errorValue := results[len(results)-1]
		if !errorValue.IsNil() {
			err := errorValue.Interface().(error)
			rh.handleError(c, 500, err)
			return
		}
	}

	// 处理第一个返回值作为响应数据
	if len(results) >= 1 {
		result := results[0]
		if !result.IsValid() || result.IsNil() {
			return
		}

		// 根据返回值类型处理响应
		switch result.Kind() {
		case reflect.String:
			c.String(200, result.String())
		case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
			c.JSON(200, result.Interface())
		case reflect.Interface:
			// 检查接口的具体类型
			concreteValue := result.Elem()
			if concreteValue.IsValid() {
				switch concreteValue.Kind() {
				case reflect.String:
					c.String(200, concreteValue.String())
				default:
					c.JSON(200, result.Interface())
				}
			}
		default:
			c.JSON(200, result.Interface())
		}
	}
}

// handleError 处理错误响应（统一错误处理）
func (rh *RequestHandler) handleError(c *app.RequestContext, statusCode int, err error) {
	response := map[string]interface{}{
		"error": err.Error(),
	}

	// 如果是路由错误，添加更多信息
	if routeErr, ok := err.(*RouteError); ok {
		response["error_type"] = string(routeErr.Type)
		if routeErr.Cause != nil {
			response["cause"] = routeErr.Cause.Error()
		}
	}

	c.JSON(statusCode, response)
}

// ControllerLifecycle 控制器生命周期管理（统一从annotation和comment包提取）
type ControllerLifecycle struct {
	handler *RequestHandler
}

// NewControllerLifecycle 创建控制器生命周期管理器
func NewControllerLifecycle(handler *RequestHandler) *ControllerLifecycle {
	return &ControllerLifecycle{
		handler: handler,
	}
}

// CreateControllerInstance 创建控制器实例
func (cl *ControllerLifecycle) CreateControllerInstance(controllerType reflect.Type) (reflect.Value, error) {
	if err := ValidateControllerType(controllerType); err != nil {
		return reflect.Value{}, err
	}

	// 确保是指针类型
	if controllerType.Kind() != reflect.Ptr {
		controllerType = reflect.PtrTo(controllerType)
	}

	// 创建实例
	controllerValue := reflect.New(controllerType.Elem())
	return controllerValue, nil
}

// InitializeController 初始化控制器
func (cl *ControllerLifecycle) InitializeController(controller interface{}, ctx *contextenhanced.Context, typeName, methodName string) {
	if iController, ok := controller.(core.IController); ok {
		iController.Init(ctx, typeName, methodName, cl.handler.app)
		iController.Prepare()
	}
}

// FinalizeController 完成控制器处理
func (cl *ControllerLifecycle) FinalizeController(controller interface{}) {
	if iController, ok := controller.(core.IController); ok {
		iController.Finish()
	}
}

// RequestProcessor 请求处理器（组合所有功能）
type RequestProcessor struct {
	handler   *RequestHandler
	lifecycle *ControllerLifecycle
	binder    *ParamBinder
}

// NewRequestProcessor 创建请求处理器
func NewRequestProcessor(app *core.App, engine *route.Engine) *RequestProcessor {
	handler := NewRequestHandler(app, engine)
	return &RequestProcessor{
		handler:   handler,
		lifecycle: NewControllerLifecycle(handler),
		binder:    handler.paramBinder,
	}
}

// ProcessRequest 处理请求（完整的请求处理流程）
func (rp *RequestProcessor) ProcessRequest(route *RouteInfo, c *app.RequestContext) {
	// 创建控制器实例
	controllerValue, err := rp.lifecycle.CreateControllerInstance(route.ControllerType)
	if err != nil {
		rp.handler.handleError(c, 500, err)
		return
	}

	controller := controllerValue.Interface()

	// 创建增强的上下文
	enhancedCtx := &contextenhanced.Context{
		RequestContext: c,
	}

	// 初始化控制器
	rp.lifecycle.InitializeController(controller, enhancedCtx, route.TypeName, route.MethodName)

	// 确保完成时调用Finish
	defer rp.lifecycle.FinalizeController(controller)

	// 验证参数
	if err := rp.binder.ValidateParams(route.Params, c); err != nil {
		rp.handler.handleError(c, 400, err)
		return
	}

	// 准备方法参数
	methodInfo := &MethodInfo{
		MethodName: route.MethodName,
		Params:     route.Params,
	}
	args, err := rp.binder.PrepareMethodArgs(methodInfo, c, controllerValue)
	if err != nil {
		rp.handler.handleError(c, 400, fmt.Errorf("failed to prepare method arguments: %w", err))
		return
	}

	// 调用控制器方法
	method := controllerValue.MethodByName(route.MethodName)
	if !method.IsValid() {
		rp.handler.handleError(c, 500, fmt.Errorf("method '%s' not found in controller", route.MethodName))
		return
	}

	results := method.Call(args)

	// 处理方法返回值
	rp.handler.handleMethodResults(c, results)
}

// GetHandler 获取处理器
func (rp *RequestProcessor) GetHandler() *RequestHandler {
	return rp.handler
}

// GetParamBinder 获取参数绑定器
func (rp *RequestProcessor) GetParamBinder() *ParamBinder {
	return rp.binder
}

// GetLifecycle 获取生命周期管理器
func (rp *RequestProcessor) GetLifecycle() *ControllerLifecycle {
	return rp.lifecycle
}