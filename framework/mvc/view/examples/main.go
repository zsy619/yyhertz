package main

import (
	"log"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/view/examples/controllers"
)

func main() {
	log.Println("🚀 YYHertz 模板引擎演示应用启动...")

	// 创建应用实例
	app := mvc.HertzApp

	// 设置静态文件路径
	app.SetStaticPath("./static", "/static")

	// 添加全局中间件
	app.Use(
	// middleware.RecoveryMiddleware(),
	// middleware.TracingMiddleware(),
	// middleware.LoggerMiddleware(),
	// middleware.CORSMiddleware(),
	// middleware.RateLimitMiddleware(100, time.Minute),
	)

	// 创建控制器实例
	exampleController := &controllers.ExampleController{}
	demoController := &controllers.DemoController{}
	templateController := &controllers.TemplateController{}
	blogsController := &controllers.BlogsController{}

	// 注册路由
	setupRoutes(app, exampleController, demoController, templateController, blogsController)

	log.Println("🚀 YYHertz 模板引擎演示应用启动成功!")
	log.Println("📍 服务器地址: http://localhost:8888")
	log.Println("")
	log.Println("📋 可用路由:")
	log.Println("GET    /                     - 模板引擎演示首页")
	log.Println("GET    /layout               - Layout继承演示")
	log.Println("GET    /beego-functions      - Beego函数演示")
	log.Println("GET    /advanced             - 高级功能演示")
	log.Println("POST   /csrf-test            - CSRF测试")
	log.Println("GET    /performance          - 性能监控")
	log.Println("GET    /templates            - 模板管理")
	log.Println("GET    /templates/:name      - 查看特定模板")
	log.Println("POST   /templates/preview    - 模板预览")
	log.Println("GET    /blogs                - 技术博客首页")
	log.Println("GET    /blog                 - 技术博客首页（别名）")
	log.Println("")

	app.Run()
}

// setupRoutes 设置路由
func setupRoutes(app *mvc.App, exampleController *controllers.ExampleController, demoController *controllers.DemoController, templateController *controllers.TemplateController, blogsController *controllers.BlogsController) {
	// 使用自动路由注册
	app.RouterAuto(exampleController, demoController, templateController, blogsController)

	// 手动配置特定路由映射
	app.RouterPrefix("/", exampleController, true, "Index", "*:/")                         // 首页
	app.RouterPrefix("/", exampleController, true, "LayoutDemo", "*:/layout")              // Layout演示
	app.RouterPrefix("/", exampleController, true, "BeegoFunctions", "*:/beego-functions") // 函数演示

	app.RouterPrefix("/", demoController, true, "AdvancedFeatures", "*:/advanced") // 高级功能
	app.RouterPrefix("/", demoController, true, "Performance", "*:/performance")   // 性能监控
	app.RouterPrefix("/", demoController, true, "CsrfTest", "POST:/csrf-test")     // CSRF测试

	app.RouterPrefix("/", templateController, true, "Index", "*:/templates")              // 模板管理
	app.RouterPrefix("/", templateController, true, "Show", "*:/templates/:name")         // 查看模板
	app.RouterPrefix("/", templateController, true, "Preview", "POST:/templates/preview") // 模板预览

	// 博客路由
	app.RouterPrefix("/", blogsController, true, "Get", "*:/blogs") // 博客列表
	app.RouterPrefix("/", blogsController, true, "Get", "*:/blog")  // 博客列表（别名）
}
