package comment

import (
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// App 基于注释注解的应用
type App struct {
	*core.App
	router *Router
}

// NewCommentApp 创建支持注释注解的应用
func NewCommentApp(baseApp *core.App) *App {
	app := &App{
		App:    baseApp,
		router: NewRouter(baseApp, baseApp.Engine),
	}

	return app
}

// AutoScanAndRegister 自动扫描并注册带注释注解的控制器
func (ca *App) AutoScanAndRegister(controllers ...core.IController) *App {
	err := ca.router.ScanAndRegister(controllers...)
	if err != nil {
		panic("Failed to auto scan and register controllers: " + err.Error())
	}
	return ca
}

// ScanControllers 扫描控制器（不立即注册）
func (ca *App) ScanControllers(controllers ...core.IController) *App {
	for _, controller := range controllers {
		err := ca.router.scanControllerSource(controller)
		if err != nil {
			panic("Failed to scan controller: " + err.Error())
		}
	}
	return ca
}

// RegisterScannedControllers 注册已扫描的控制器
func (ca *App) RegisterScannedControllers(controllers ...core.IController) *App {
	for _, controller := range controllers {
		err := ca.router.registerController(controller)
		if err != nil {
			panic("Failed to register controller: " + err.Error())
		}
	}
	return ca
}

// GetRouter 获取注释路由器
func (ca *App) GetRouter() *Router {
	return ca.router
}

// GetRoutes 获取所有注释路由信息
func (ca *App) GetRoutes() []*RouteInfo {
	return ca.router.GetRegisteredRoutes()
}

// 便捷的创建函数

// NewCommentWithApp 创建支持注释注解的应用
func NewCommentWithApp(baseApp *core.App) *App {
	return NewCommentApp(baseApp)
}

// 全局便捷函数

// ScanSourceFile 全局扫描源文件
func ScanSourceFile(filename string) error {
	return GetGlobalParser().ParseSourceFile(filename)
}

// ScanPackage 全局扫描包
func ScanPackage(packagePath string) error {
	return GetGlobalParser().ScanPackage(packagePath)
}

// GetGlobalControllerInfo 获取全局控制器信息
func GetGlobalControllerInfo(packageName, typeName string) *ControllerInfo {
	return GetGlobalParser().GetControllerInfo(packageName, typeName)
}

// GetGlobalMethodInfo 获取全局方法信息
func GetGlobalMethodInfo(packageName, typeName, methodName string) *MethodInfo {
	return GetGlobalParser().GetMethodInfo(packageName, typeName, methodName)
}

// ListAllAnnotations 列出所有扫描到的注解信息
func ListAllAnnotations() (controllers []*ControllerInfo, methods []*MethodInfo) {
	parser := GetGlobalParser()

	// 获取所有控制器
	for _, info := range parser.ControllerInfos {
		controllers = append(controllers, info)
	}

	// 获取所有方法
	for _, info := range parser.MethodInfos {
		methods = append(methods, info)
	}

	return
}
