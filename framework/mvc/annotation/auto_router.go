package annotation

import (
	"fmt"
	"reflect"

	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/routing"
)

// AutoRouter 自动路由器
type AutoRouter struct {
	parser    *AnnotationParser
	registry  *MethodRegistry
	app       *core.App
	engine    *route.Engine
	processor *routing.RequestProcessor // 使用routing包的请求处理器
}

// NewAutoRouter 创建自动路由器
func NewAutoRouter(app *core.App, engine *route.Engine) *AutoRouter {
	return &AutoRouter{
		parser:    NewAnnotationParser(),
		registry:  GetRegistry(),
		app:       app,
		engine:    engine,
		processor: routing.NewRequestProcessor(app, engine), // 使用routing包的处理器
	}
}

// ScanAndRegister 扫描并注册控制器
func (ar *AutoRouter) ScanAndRegister(controllers ...core.IController) error {
	// 解析所有控制器
	infos, err := ar.parser.ParseAllControllers(controllers...)
	if err != nil {
		return fmt.Errorf("failed to parse controllers: %w", err)
	}

	// 注册路由
	for _, info := range infos {
		err := ar.registerControllerRoutes(info)
		if err != nil {
			return fmt.Errorf("failed to register routes for controller %s: %w", info.Name, err)
		}
	}

	return nil
}

// registerControllerRoutes 注册控制器路由
func (ar *AutoRouter) registerControllerRoutes(info *ControllerInfo) error {
	routes := ar.parser.BuildRouteInfo(info)

	for _, route := range routes {
		err := ar.registerRoute(route)
		if err != nil {
			return fmt.Errorf("failed to register route %s %s: %w", route.HTTPMethod, route.Path, err)
		}
	}

	return nil
}

// registerRoute 注册单个路由
func (ar *AutoRouter) registerRoute(route *RouteInfo) error {
	// 转换为routing包的RouteInfo
	routingRoute := ar.convertToRoutingRoute(route)
	
	// 使用routing包的处理器注册路由
	return ar.processor.GetHandler().RegisterRoute(routingRoute)
}

// convertToRoutingRoute 转换为routing包的RouteInfo
func (ar *AutoRouter) convertToRoutingRoute(route *RouteInfo) *routing.RouteInfo {
	// 转换参数信息
	var params []*routing.ParamInfo
	for _, param := range route.Params {
		routingParam := &routing.ParamInfo{
			Name:         param.Name,
			Source:       ar.convertParamSource(param.Kind),
			Required:     param.Required,
			DefaultValue: param.DefaultVal,
			Description:  "", // annotation包没有描述字段
			Type:         param.Type.String(),
		}
		params = append(params, routingParam)
	}

	// 从控制器类型获取包名和类型名
	packageName := route.ControllerType.PkgPath()
	if packageName == "" {
		packageName = "main"
	}
	
	typeName := route.ControllerType.Name()
	if route.ControllerType.Kind() == reflect.Ptr {
		typeName = route.ControllerType.Elem().Name()
	}

	return &routing.RouteInfo{
		Path:           route.Path,
		HTTPMethod:     route.HTTPMethod,
		PackageName:    packageName,
		TypeName:       typeName,
		ControllerType: route.ControllerType,
		MethodName:     route.MethodName,
		Description:    "", // annotation包的RouteInfo没有描述字段
		Params:         params,
		Middlewares:    []string{}, // annotation包的RouteInfo没有中间件字段
		Tags:           route.Tags,
		Source:         routing.SourceStructTag, // annotation包来源为struct标签
	}
}

// convertParamSource 转换参数来源类型
func (ar *AutoRouter) convertParamSource(kind ParamKind) routing.ParamSource {
	switch kind {
	case ParamKindPath:
		return routing.ParamSourcePath
	case ParamKindQuery:
		return routing.ParamSourceQuery
	case ParamKindBody:
		return routing.ParamSourceBody
	case ParamKindHeader:
		return routing.ParamSourceHeader
	case ParamKindCookie:
		return routing.ParamSourceCookie
	case ParamKindForm:
		return routing.ParamSourceForm
	default:
		return routing.ParamSourceQuery // 默认为查询参数
	}
}


// RegisterController 注册单个控制器
func (ar *AutoRouter) RegisterController(controller core.IController) error {
	return ar.ScanAndRegister(controller)
}

// GetRegisteredRoutes 获取已注册的路由信息
func (ar *AutoRouter) GetRegisteredRoutes() []*RouteInfo {
	var allRoutes []*RouteInfo

	controllers := ar.registry.GetAllControllers()
	for _, info := range controllers {
		routes := ar.parser.BuildRouteInfo(info)
		allRoutes = append(allRoutes, routes...)
	}

	return allRoutes
}

// GetRoutingRoutes 获取routing包格式的路由信息
func (ar *AutoRouter) GetRoutingRoutes() []*routing.RouteInfo {
	var routingRoutes []*routing.RouteInfo
	
	annotationRoutes := ar.GetRegisteredRoutes()
	for _, route := range annotationRoutes {
		routingRoute := ar.convertToRoutingRoute(route)
		routingRoutes = append(routingRoutes, routingRoute)
	}
	
	return routingRoutes
}
