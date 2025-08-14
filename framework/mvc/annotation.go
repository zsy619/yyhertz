// Package mvc 提供MVC框架的注解和注释解析功能
//
// 本文件提供了全局的注解和注释应用实例管理，支持两种路由注册方式：
// 1. Annotation - 基于Go结构体标签的注解方式
// 2. Comment - 基于Go代码注释的解析方式
//
// 设计目标：
// - 提供统一的全局访问接口
// - 支持混合使用两种注册方式
// - 线程安全的实例管理
// - 与主框架的无缝集成
package mvc

import (
	"github.com/zsy619/yyhertz/framework/mvc/annotation"
	"github.com/zsy619/yyhertz/framework/mvc/comment"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// ============= 全局变量定义 =============

// 全局应用实例，线程安全且只初始化一次
var (
	// AnnotationApp 全局注解应用实例
	//
	// 用于管理基于Go结构体标签的路由注册，支持：
	// - @Route 标签定义路由
	// - @Middleware 标签定义中间件
	// - @Param 标签定义参数验证
	// - @Response 标签定义响应格式
	AnnotationApp *annotation.AnnotationApp

	// CommentApp 全局注释注解应用实例
	//
	// 用于管理基于Go代码注释的路由注册，支持：
	// - // @Router 注释定义路由
	// - // @Middleware 注释定义中间件
	// - // @Param 注释定义参数
	// - // @Success/@Failure 注释定义响应
	CommentApp *comment.App
)

// ============= 全局应用实例访问方法 =============

// GetHertzApp 获取全局Hertz应用实例
//
// 返回全局HertzApp实例，该实例在框架初始化时创建。
// 该方法是线程安全的，可以在并发环境中安全调用。
//
// 返回值：
//   - *App: 全局Hertz应用实例，如果未初始化则返回nil
//
// 使用示例：
//
//	app := mvc.GetHertzApp()
//	if app != nil {
//		app.RouterAuto(&MyController{})
//	}
func GetHertzApp() *App {
	// 注意：这里暂时注释了锁的使用，因为HertzApp在init()中初始化
	// 如果需要动态修改HertzApp，需要启用互斥锁保护
	// mutex.RLock()
	// defer mutex.RUnlock()
	return HertzApp
}

// GetAnnotationApp 获取全局注解应用实例
//
// 返回全局AnnotationApp实例，用于管理基于Go结构体标签的注解路由。
// 该方法是线程安全的。
//
// 返回值：
//   - *annotation.AnnotationApp: 注解应用实例
//
// 使用示例：
//
//	app := mvc.GetAnnotationApp()
//	app.AutoRegister(&MyController{})
func GetAnnotationApp() *annotation.AnnotationApp {
	return AnnotationApp
}

// GetCommentApp 获取全局注释注解应用实例
//
// 返回全局CommentApp实例，用于管理基于Go代码注释的注解路由。
// 该方法是线程安全的。
//
// 返回值：
//   - *comment.App: 注释注解应用实例
//
// 使用示例：
//
//	app := mvc.GetCommentApp()
//	app.AutoScanAndRegister(&MyController{})
func GetCommentApp() *comment.App {
	return CommentApp
}

// ============= 控制器注册方法 =============

// AutoRegister 全局自动注册控制器（基于注解标签）
//
// 使用注解标签方式自动扫描并注册控制器路由。该方法会解析控制器结构体上的
// @Route、@Middleware等标签，自动生成对应的路由映射。
//
// 参数：
//   - controllers: ...core.IController - 要注册的控制器实例列表
//
// 支持的注解标签：
//   - @Route: 定义路由路径和HTTP方法
//   - @Middleware: 定义中间件
//   - @Param: 定义参数验证规则
//   - @Response: 定义响应格式
//
// 使用示例：
//
//	type UserController struct {
//		mvc.BaseController `@Route("/users")`
//	}
//
//	func (c *UserController) Get() `@Route("GET /:id")` {
//		// 处理GET请求
//	}
//
//	mvc.AutoRegister(&UserController{})
func AutoRegister(controllers ...core.IController) {
	app := GetAnnotationApp()
	app.AutoRegister(controllers...)
}

// AutoScanAndRegister 全局自动扫描并注册控制器（基于注释解析）
//
// 使用注释解析方式自动扫描并注册控制器路由。该方法会解析控制器方法上的
// // @Router等注释，自动生成对应的路由映射。
//
// 参数：
//   - controllers: ...core.IController - 要注册的控制器实例列表
//
// 支持的注释格式：
//   - // @Router /path [method] - 定义路由
//   - // @Middleware middleware_name - 定义中间件
//   - // @Param name query string true "参数说明" - 定义参数
//   - // @Success 200 {object} ResponseModel "成功响应" - 定义成功响应
//   - // @Failure 400 {object} ErrorModel "错误响应" - 定义错误响应
//
// 使用示例：
//
//	type UserController struct {
//		mvc.BaseController
//	}
//
//	// @Router /users/{id} [GET]
//	// @Param id path int true "用户ID"
//	// @Success 200 {object} User "用户信息"
//	func (c *UserController) Get() {
//		// 处理GET请求
//	}
//
//	mvc.AutoScanAndRegister(&UserController{})
func AutoScanAndRegister(controllers ...core.IController) {
	app := GetCommentApp()
	app.AutoScanAndRegister(controllers...)
}

// RegisterControllers 混合注册控制器
//
// 同时使用注解标签和注释解析两种方式注册控制器。这个方法会：
// 1. 首先尝试使用注解标签方式注册
// 2. 然后尝试使用注释解析方式注册
//
// 这样可以确保不论控制器使用哪种注解方式都能被正确注册。
// 如果同一个路由被两种方式同时定义，后注册的会覆盖先注册的。
//
// 参数：
//   - controllers: ...core.IController - 要注册的控制器实例列表
//
// 注意事项：
//   - 避免在同一个控制器中混用两种注解方式定义相同路由
//   - 建议在项目中统一使用一种注解方式
//   - 该方法主要用于迁移期间的兼容性支持
//
// 使用示例：
//
//	// 控制器可以使用任意一种或两种注解方式
//	mvc.RegisterControllers(&UserController{}, &OrderController{})
func RegisterControllers(controllers ...core.IController) {
	// 先注册到annotation系统
	AutoRegister(controllers...)
	// 再注册到comment系统
	AutoScanAndRegister(controllers...)
}

// ============= 路由信息查询方法 =============

// GetAllRoutes 获取所有已注册的路由信息
//
// 返回系统中所有通过注解标签和注释解析方式注册的路由信息。
// 这个方法对于调试、文档生成和API管理非常有用。
//
// 返回值：
//   - annotationRoutes: []*annotation.RouteInfo - 通过注解标签注册的路由列表
//   - commentRoutes: []*comment.RouteInfo - 通过注释解析注册的路由列表
//
// RouteInfo包含的信息：
//   - Path: 路由路径
//   - Method: HTTP方法
//   - Controller: 控制器名称
//   - Action: 控制器方法名
//   - Middlewares: 中间件列表
//   - Parameters: 参数定义
//   - Responses: 响应定义
//
// 使用示例：
//
//	annotationRoutes, commentRoutes := mvc.GetAllRoutes()
//
//	// 打印所有注解路由
//	for _, route := range annotationRoutes {
//		fmt.Printf("Annotation Route: %s %s -> %s.%s\n",
//			route.Method, route.Path, route.Controller, route.Action)
//	}
//
//	// 打印所有注释路由
//	for _, route := range commentRoutes {
//		fmt.Printf("Comment Route: %s %s -> %s.%s\n",
//			route.Method, route.Path, route.Controller, route.Action)
//	}
func GetAllRoutes() (annotationRoutes []*annotation.RouteInfo, commentRoutes []*comment.RouteInfo) {
	annotationApp := GetAnnotationApp()
	commentApp := GetCommentApp()

	annotationRoutes = annotationApp.GetAnnotatedRoutes()
	commentRoutes = commentApp.GetRoutes()

	return
}
