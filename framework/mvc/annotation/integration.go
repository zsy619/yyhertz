package annotation

import (
	"reflect"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// AnnotationApp 注解应用扩展
type AnnotationApp struct {
	*core.App
	autoRouter *AutoRouter
}

// NewAnnotationWithApp 创建支持注解的应用
func NewAnnotationWithApp(baseApp *core.App) *AnnotationApp {
	annotationApp := &AnnotationApp{
		App: baseApp,
	}
	annotationApp.autoRouter = NewAutoRouter(baseApp, baseApp.Engine)

	return annotationApp
}

// NewAnnotationWithHertz 创建支持注解的应用（直接传入Hertz实例）
func NewAnnotationWithHertz(hertz *server.Hertz) *AnnotationApp {
	baseApp := core.NewApp()
	baseApp.Engine = hertz.Engine
	return NewAnnotationWithApp(baseApp)
}

// AutoRegister 自动注册控制器（支持注解）
func (aa *AnnotationApp) AutoRegister(controllers ...core.IController) *AnnotationApp {
	err := aa.autoRouter.ScanAndRegister(controllers...)
	if err != nil {
		panic("Failed to auto register controllers: " + err.Error())
	}
	return aa
}

// RegisterAnnotatedController 注册带注解的控制器
func (aa *AnnotationApp) RegisterAnnotatedController(controller core.IController) *AnnotationApp {
	err := aa.autoRouter.RegisterController(controller)
	if err != nil {
		panic("Failed to register annotated controller: " + err.Error())
	}
	return aa
}

// 扩展现有的App方法以支持注解

// AutoRouters 重写以支持注解扫描
func (aa *AnnotationApp) AutoRouters(controllers ...core.IController) *core.App {
	// 先尝试注解方式注册
	var annotatedControllers []core.IController
	var regularControllers []core.IController

	for _, controller := range controllers {
		controllerType := reflect.TypeOf(controller)
		if controllerType.Kind() == reflect.Ptr {
			controllerType = controllerType.Elem()
		}

		// 检查是否有注解标签
		if hasAnnotations(controllerType) {
			annotatedControllers = append(annotatedControllers, controller)
		} else {
			regularControllers = append(regularControllers, controller)
		}
	}

	// 注册带注解的控制器
	if len(annotatedControllers) > 0 {
		err := aa.autoRouter.ScanAndRegister(annotatedControllers...)
		if err != nil {
			panic("Failed to register annotated controllers: " + err.Error())
		}
	}

	// 使用原有方式注册其他控制器
	if len(regularControllers) > 0 {
		aa.App.AutoRouters(regularControllers...)
	}

	return aa.App
}

// hasAnnotations 检查类型是否有注解标签
func hasAnnotations(structType reflect.Type) bool {
	if structType.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// 检查是否有任何注解标签，包括空值的标签
		tags := string(field.Tag)
		if strings.Contains(tags, RestControllerTag) ||
			strings.Contains(tags, ControllerTag) ||
			strings.Contains(tags, RequestMappingTag) ||
			strings.Contains(tags, ServiceTag) ||
			strings.Contains(tags, RepositoryTag) ||
			strings.Contains(tags, ComponentTag) {
			return true
		}
	}

	return false
}

// GetAutoRouter 获取自动路由器
func (aa *AnnotationApp) GetAutoRouter() *AutoRouter {
	return aa.autoRouter
}

// GetAnnotatedRoutes 获取所有注解路由信息
func (aa *AnnotationApp) GetAnnotatedRoutes() []*RouteInfo {
	return aa.autoRouter.GetRegisteredRoutes()
}

// 全局注册函数，用于在init()中注册方法映射

// RegisterMethodMapping 全局注册方法映射
func RegisterMethodMapping(controllerType reflect.Type, methodName, httpMethod, path string) *MethodMappingBuilder {
	registry := GetRegistry()

	switch httpMethod {
	case "GET":
		return registry.RegisterGetMapping(controllerType, methodName, path)
	case "POST":
		return registry.RegisterPostMapping(controllerType, methodName, path)
	case "PUT":
		return registry.RegisterPutMapping(controllerType, methodName, path)
	case "DELETE":
		return registry.RegisterDeleteMapping(controllerType, methodName, path)
	case "PATCH":
		return registry.RegisterPatchMapping(controllerType, methodName, path)
	case "ANY":
		return registry.RegisterAnyMapping(controllerType, methodName, path)
	default:
		return registry.RegisterGetMapping(controllerType, methodName, path)
	}
}

// 便捷的方法映射注册函数

// RegisterGetMethod 注册GET方法
func RegisterGetMethod(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterGetMapping(controllerType, methodName, path)
}

// RegisterPostMethod 注册POST方法
func RegisterPostMethod(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterPostMapping(controllerType, methodName, path)
}

// RegisterPutMethod 注册PUT方法
func RegisterPutMethod(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterPutMapping(controllerType, methodName, path)
}

// RegisterDeleteMethod 注册DELETE方法
func RegisterDeleteMethod(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterDeleteMapping(controllerType, methodName, path)
}

func RegisterHeadMethod(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterHeadMapping(controllerType, methodName, path)
}

// RegisterPatchMethod 注册PATCH方法
func RegisterPatchMethod(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterPatchMapping(controllerType, methodName, path)
}

// RegisterAnyMethod 注册任意方法
func RegisterAnyMethod(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterAnyMapping(controllerType, methodName, path)
}
