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
	beegoFunctionsController := &controllers.BeeGoFunctionsController{}

	// 注册路由
	setupRoutes(app, exampleController, demoController, templateController, blogsController, beegoFunctionsController)

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
	log.Println("🚀 DemoController 高级功能路由:")
	log.Println("GET    /Demo/AdvancedFeatures - 高级功能演示（自动路由）")
	log.Println("GET    /demo/advancedfeatures - 高级功能演示（小写变体）")
	log.Println("GET    /advanced - 高级功能演示（手动映射）")
	log.Println("GET    /Demo/Performance - 性能监控（自动路由）")
	log.Println("GET    /demo/performance - 性能监控（小写变体）")
	log.Println("GET    /performance - 性能监控（手动映射）")
	log.Println("POST   /Demo/CsrfTest - CSRF测试（自动路由）")
	log.Println("POST   /demo/csrftest - CSRF测试（小写变体）")
	log.Println("POST   /csrf-test - CSRF测试（手动映射）")
	log.Println("")
	log.Println("🧪 模板函数测试路由:")
	log.Println("GET    /test/beego-functions - 模板函数测试首页")
	log.Println("GET    /test/beego-functions/include - Include函数测试")
	log.Println("GET    /test/beego-functions/templateinclude - TemplateInclude函数测试")
	log.Println("GET    /test/beego-functions/partial - Partial函数测试")
	log.Println("GET    /test/beego-functions/componenttemplate - ComponentTemplate函数测试")
	log.Println("GET    /test/beego-functions/rendertemplate - RenderTemplate函数测试")
	log.Println("GET    /test/beego-functions/template - Template函数测试")
	log.Println("GET    /test/beego-functions/include-templateinclude - Include+TemplateInclude组合测试")
	log.Println("GET    /test/beego-functions/partial-component - Partial+Component组合测试")
	log.Println("GET    /test/beego-functions/all-functions - 所有函数综合测试")
	log.Println("GET    /test/beego-functions/nested-includes - 嵌套包含测试")
	log.Println("GET    /test/beego-functions/error-handling - 错误处理测试")
	log.Println("")
	log.Println("🎨 Layout版本测试路由:")
	log.Println("GET    /test/beego-functions/include_layout - Include函数测试 (Layout版)")
	log.Println("GET    /test/beego-functions/templateinclude_layout - TemplateInclude函数测试 (Layout版)")
	log.Println("GET    /test/beego-functions/partial_layout - Partial函数测试 (Layout版)")
	log.Println("GET    /test/beego-functions/componenttemplate_layout - ComponentTemplate函数测试 (Layout版)")
	log.Println("GET    /test/beego-functions/rendertemplate_layout - RenderTemplate函数测试 (Layout版)")
	log.Println("GET    /test/beego-functions/template_layout - Template函数测试 (Layout版)")
	log.Println("")
	log.Println("🚀 组合函数Layout版本测试路由:")
	log.Println("GET    /test/beego-functions/allfunctions_layout - 所有函数综合测试 (Layout版)")
	log.Println("GET    /test/beego-functions/nestedincludes_layout - 嵌套包含测试 (Layout版)")
	log.Println("GET    /test/beego-functions/errorhandling_layout - 错误处理测试 (Layout版)")
	log.Println("GET    /test/beego-functions/includetemplate_layout - Include+TemplateInclude组合测试 (Layout版)")
	log.Println("GET    /test/beego-functions/partialcomponent_layout - Partial+Component组合测试 (Layout版)")
	log.Println("")

	app.Run()
}

// setupRoutes 设置路由
func setupRoutes(app *mvc.App, exampleController *controllers.ExampleController, demoController *controllers.DemoController, templateController *controllers.TemplateController, blogsController *controllers.BlogsController, beegoFunctionsController *controllers.BeeGoFunctionsController) {
	// 使用自动路由注册
	app.RouterAuto(exampleController, demoController, templateController, blogsController, beegoFunctionsController)

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

	// 🧪 模板函数测试路由
	app.RouterPrefix("/test", beegoFunctionsController, true, "Index", "*:/beego-functions")                                      // 测试首页
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestInclude", "*:/beego-functions/include")                       // Include测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestTemplateInclude", "*:/beego-functions/templateinclude")       // TemplateInclude测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestPartial", "*:/beego-functions/partial")                       // Partial测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestComponentTemplate", "*:/beego-functions/componenttemplate")   // ComponentTemplate测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestRenderTemplate", "*:/beego-functions/rendertemplate")         // RenderTemplate测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestTemplate", "*:/beego-functions/template")                   // Template测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestIncludeTemplateInclude", "*:/beego-functions/include-templateinclude") // Include+TemplateInclude组合测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestPartialComponent", "*:/beego-functions/partial-component")    // Partial+Component组合测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestAllFunctions", "*:/beego-functions/all-functions")            // 所有函数综合测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestNestedIncludes", "*:/beego-functions/nested-includes")        // 嵌套包含测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestErrorHandling", "*:/beego-functions/error-handling")          // 错误处理测试

	// 🎨 Layout版本测试路由
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestIncludeLayout", "*:/beego-functions/include_layout")                       // Include Layout测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestTemplateIncludeLayout", "*:/beego-functions/templateinclude_layout")       // TemplateInclude Layout测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestPartialLayout", "*:/beego-functions/partial_layout")                       // Partial Layout测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestComponentTemplateLayout", "*:/beego-functions/componenttemplate_layout")   // ComponentTemplate Layout测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestRenderTemplateLayout", "*:/beego-functions/rendertemplate_layout")         // RenderTemplate Layout测试
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestTemplateLayout", "*:/beego-functions/template_layout")                     // Template Layout测试

	// 🚀 组合函数Layout版本测试路由
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestAllFunctionsLayout", "*:/beego-functions/allfunctions_layout")                       // 所有函数综合测试 (Layout版)
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestNestedIncludesLayout", "*:/beego-functions/nestedincludes_layout")                   // 嵌套包含测试 (Layout版)
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestErrorHandlingLayout", "*:/beego-functions/errorhandling_layout")                     // 错误处理测试 (Layout版)
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestIncludeTemplateIncludeLayout", "*:/beego-functions/includetemplate_layout")         // Include+TemplateInclude组合测试 (Layout版)
	app.RouterPrefix("/test", beegoFunctionsController, true, "TestPartialComponentLayout", "*:/beego-functions/partialcomponent_layout")               // Partial+Component组合测试 (Layout版)
}
