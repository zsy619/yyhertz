package mvc

import (
	"github.com/zsy619/yyhertz/framework/mvc/annotation"
	"github.com/zsy619/yyhertz/framework/mvc/comment"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// 全局变量，线程安全且只初始化一次
var (
	// AnnotationApp 全局注解应用实例
	AnnotationApp *annotation.AnnotationApp

	// CommentApp 全局注释注解应用实例
	CommentApp *comment.App
)

// GetHertzApp 获取全局Hertz应用实例（线程安全）
func GetHertzApp() *App {
	// mutex.RLock()
	// defer mutex.RUnlock()
	return HertzApp
}

// GetAnnotationApp 获取全局注解应用实例（线程安全）
func GetAnnotationApp() *annotation.AnnotationApp {
	return AnnotationApp
}

// GetCommentApp 获取全局注释注解应用实例（线程安全）
func GetCommentApp() *comment.App {
	return CommentApp
}

// AutoRegister 全局自动注册控制器（annotation方式）
func AutoRegister(controllers ...core.IController) {
	app := GetAnnotationApp()
	app.AutoRegister(controllers...)
}

// AutoScanAndRegister 全局自动扫描并注册控制器（comment方式）
func AutoScanAndRegister(controllers ...core.IController) {
	app := GetCommentApp()
	app.AutoScanAndRegister(controllers...)
}

// RegisterControllers 混合注册控制器（同时支持annotation和comment）
func RegisterControllers(controllers ...core.IController) {
	// 先注册到annotation系统
	AutoRegister(controllers...)
	// 再注册到comment系统
	AutoScanAndRegister(controllers...)
}

// GetAllRoutes 获取所有路由信息
func GetAllRoutes() (annotationRoutes []*annotation.RouteInfo, commentRoutes []*comment.RouteInfo) {
	annotationApp := GetAnnotationApp()
	commentApp := GetCommentApp()

	annotationRoutes = annotationApp.GetAnnotatedRoutes()
	commentRoutes = commentApp.GetRoutes()

	return
}
