package comment

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/routing"
)

// Router 基于注释的路由器
type Router struct {
	parser    *AnnotationParser
	app       *core.App
	engine    *route.Engine
	processor *routing.RequestProcessor // 使用routing包的请求处理器
}

// NewRouter 创建基于注释的路由器
func NewRouter(app *core.App, engine *route.Engine) *Router {
	return &Router{
		parser:    GetGlobalParser(),
		app:       app,
		engine:    engine,
		processor: routing.NewRequestProcessor(app, engine), // 使用routing包的处理器
	}
}

// ScanAndRegister 扫描并注册控制器
func (r *Router) ScanAndRegister(controllers ...core.IController) error {
	// 先扫描源文件获取注释信息
	for _, controller := range controllers {
		err := r.scanControllerSource(controller)
		if err != nil {
			return fmt.Errorf("failed to scan controller source: %w", err)
		}
	}

	// 然后注册路由
	for _, controller := range controllers {
		err := r.registerController(controller)
		if err != nil {
			return fmt.Errorf("failed to register controller: %w", err)
		}
	}

	return nil
}

// scanControllerSource 扫描控制器源码
func (r *Router) scanControllerSource(controller core.IController) error {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	// 获取包路径和类型名
	packagePath := controllerType.PkgPath()
	typeName := controllerType.Name()

	if packagePath == "" {
		return fmt.Errorf("cannot get package path for type %s", typeName)
	}

	// 尝试获取源文件路径
	sourceFile, err := r.findSourceFile(controller)
	if err != nil {
		return err
	}

	// 解析源文件
	return r.parser.ParseSourceFile(sourceFile)
}

// findSourceFile 查找控制器的源文件
func (r *Router) findSourceFile(controller core.IController) (string, error) {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	// 获取调用者的文件路径作为提示
	_, filename, _, ok := runtime.Caller(4) // 调整调用深度
	if !ok {
		return "", fmt.Errorf("cannot determine source file location")
	}

	// 直接返回调用方的文件，这样可以解析其中的所有控制器
	return filename, nil
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	_, err := filepath.Abs(filename)
	return err == nil
}

// registerController 注册控制器
func (r *Router) registerController(controller core.IController) error {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	packageName := getPackageName(controllerType.PkgPath())
	typeName := controllerType.Name()

	// 获取控制器信息
	controllerInfo := r.parser.GetControllerInfo(packageName, typeName)
	if controllerInfo == nil {
		return fmt.Errorf("no annotation found for controller %s.%s", packageName, typeName)
	}

	// 获取控制器的所有方法
	methods := r.parser.GetControllerMethods(packageName, typeName)

	// 注册每个方法的路由
	for _, methodInfo := range methods {
		err := r.registerMethodRoute(controllerType, controllerInfo, methodInfo)
		if err != nil {
			return fmt.Errorf("failed to register route for method %s: %w", methodInfo.MethodName, err)
		}
	}

	return nil
}

// registerMethodRoute 注册方法路由
func (r *Router) registerMethodRoute(controllerType reflect.Type, controllerInfo *ControllerInfo, methodInfo *MethodInfo) error {
	// 组合完整路径
	fullPath := routing.CombinePath(controllerInfo.BasePath, methodInfo.Path)
	
	// 转换为routing包的RouteInfo
	routingRoute := r.convertToRoutingRoute(controllerType, controllerInfo, methodInfo, fullPath)
	
	// 使用routing包的处理器注册路由
	return r.processor.GetHandler().RegisterRoute(routingRoute)
}

// convertToRoutingRoute 转换为routing包的RouteInfo
func (r *Router) convertToRoutingRoute(controllerType reflect.Type, controllerInfo *ControllerInfo, methodInfo *MethodInfo, fullPath string) *routing.RouteInfo {
	// 转换参数信息
	var params []*routing.ParamInfo
	for _, param := range methodInfo.Params {
		routingParam := &routing.ParamInfo{
			Name:         param.Name,
			Source:       r.convertParamSource(param.Source),
			Required:     param.Required,
			DefaultValue: param.DefaultValue,
			Description:  param.Description,
			Type:         "string", // comment包没有类型信息，默认为string
		}
		params = append(params, routingParam)
	}

	// 获取包名和类型名
	packageName := controllerType.PkgPath()
	if packageName == "" {
		packageName = "main"
	}
	
	typeName := controllerType.Name()
	if controllerType.Kind() == reflect.Ptr {
		typeName = controllerType.Elem().Name()
	}

	return &routing.RouteInfo{
		Path:           fullPath,
		HTTPMethod:     methodInfo.HTTPMethod,
		PackageName:    packageName,
		TypeName:       typeName,
		ControllerType: reflect.PtrTo(controllerType), // 确保是指针类型
		MethodName:     methodInfo.MethodName,
		Description:    methodInfo.Description,
		Params:         params,
		Middlewares:    methodInfo.Middlewares,
		Tags:           methodInfo.Tags,
		Source:         routing.SourceComment, // comment包来源为注释
	}
}

// convertParamSource 转换参数来源类型
func (r *Router) convertParamSource(source ParamSource) routing.ParamSource {
	switch source {
	case ParamSourcePath:
		return routing.ParamSourcePath
	case ParamSourceQuery:
		return routing.ParamSourceQuery
	case ParamSourceBody:
		return routing.ParamSourceBody
	case ParamSourceHeader:
		return routing.ParamSourceHeader
	case ParamSourceCookie:
		return routing.ParamSourceCookie
	case ParamSourceForm:
		return routing.ParamSourceForm
	default:
		return routing.ParamSourceQuery // 默认为查询参数
	}
}

// GetRegisteredRoutes 获取已注册的路由信息
func (r *Router) GetRegisteredRoutes() []*RouteInfo {
	var routes []*RouteInfo

	for _, controllerInfo := range r.parser.ControllerInfos {
		methods := r.parser.GetControllerMethods(controllerInfo.PackageName, controllerInfo.TypeName)

		for _, methodInfo := range methods {
			route := &RouteInfo{
				Path:        routing.CombinePath(controllerInfo.BasePath, methodInfo.Path),
				HTTPMethod:  methodInfo.HTTPMethod,
				PackageName: methodInfo.PackageName,
				TypeName:    methodInfo.TypeName,
				MethodName:  methodInfo.MethodName,
				Description: methodInfo.Description,
				Params:      methodInfo.Params,
				Middlewares: methodInfo.Middlewares,
			}
			routes = append(routes, route)
		}
	}

	return routes
}

// GetRoutingRoutes 获取routing包格式的路由信息
func (r *Router) GetRoutingRoutes() []*routing.RouteInfo {
	var routingRoutes []*routing.RouteInfo
	
	for _, controllerInfo := range r.parser.ControllerInfos {
		methods := r.parser.GetControllerMethods(controllerInfo.PackageName, controllerInfo.TypeName)

		for _, methodInfo := range methods {
			// 由于我们没有控制器类型，创建一个伪类型用于转换
			// 在实际使用中，应该从某处获取真实的控制器类型
			var controllerType reflect.Type
			fullPath := routing.CombinePath(controllerInfo.BasePath, methodInfo.Path)
			
			routingRoute := r.convertToRoutingRoute(controllerType, controllerInfo, methodInfo, fullPath)
			routingRoutes = append(routingRoutes, routingRoute)
		}
	}
	
	return routingRoutes
}

// getPackageName 从包路径获取包名
func getPackageName(pkgPath string) string {
	if pkgPath == "" {
		return ""
	}

	parts := strings.Split(pkgPath, "/")
	return parts[len(parts)-1]
}

// RouteInfo 基于注释的路由信息
type RouteInfo struct {
	Path        string       // 路径
	HTTPMethod  string       // HTTP方法
	PackageName string       // 包名
	TypeName    string       // 类型名
	MethodName  string       // 方法名
	Description string       // 描述
	Params      []*ParamInfo // 参数信息
	Middlewares []string     // 中间件
}

