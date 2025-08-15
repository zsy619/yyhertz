package main

import (
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/example/simple/controllers"
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/devtools"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
	// 创建应用实例
	app := mvc.HertzApp

	// 设置开发工具
	if err := devtools.SetupDevTools(app, nil); err != nil {
		mvc.LogErrorf("设置开发工具失败: %v\n", err)
	}

	app.SetStaticPath("./static")

	// 设置静态文件路径 - 避免与默认路径冲突
	app.SetStaticPaths(map[string]string{
		"./assets":  "/assets",
		"./uploads": "/uploads",
		"./cdn":     "/cdn",
		"./public":  "/public",
	})

	// 设置视图模板路径
	app.SetViewPath("./views")

	// 添加全局中间件
	app.Use(
		middleware.RecoveryMiddleware(),
		middleware.TracingMiddleware(),
		middleware.LoggerMiddleware(),
		middleware.CORSMiddleware(),
		middleware.RateLimitMiddleware(100, time.Minute),
	)

	// 创建控制器实例
	homeController := &controllers.HomeController{}
	userController := &controllers.UserController{}
	adminController := &controllers.AdminController{}
	markdownController := &controllers.MarkdownController{}
	docsController := &controllers.DocsController{}

	// 自动注册路由
	app.RouterAuto(homeController, userController, adminController, markdownController, docsController)

	app.RouterPrefix("/", homeController, true, "GetIndex", "*:/")

	fmt.Println("🚀 YYHertz Namespace功能演示启动...", homeController.GetControllerName())
	fmt.Println("		0000🚀🚀🚀 ", homeController.GetControllerName())
	fmt.Println("		0000🚀🚀🚀 ", userController.GetControllerName())
	fmt.Println("		0000🚀🚀🚀 ", adminController.GetControllerName())
	fmt.Println("		0000🚀🚀🚀 ", markdownController.GetControllerName())
	fmt.Println("		0000🚀🚀🚀 ", docsController.GetControllerName())

	nsDocs := mvc.NewNamespace("/docs",
		// ============= 开始使用 =============
		mvc.NSNamespace("/getting-started",
			mvc.NSRouter("/overview", docsController, "*:GetGettingStartedOverview"),
			mvc.NSRouter("/quickstart", docsController, "*:GetGettingStartedQuickstart"),
			mvc.NSRouter("/structure", docsController, "*:GetGettingStartedStructure"),
		),
		// ============= MVC核心 =============
		mvc.NSNamespace("/mvc-core",
			mvc.NSRouter("/application", docsController, "*:GetMvcCoreApplication"),
			mvc.NSRouter("/controller", docsController, "*:GetMvcCoreController"),
			// ============= 控制器子路由 =============
			mvc.NSNamespace("/controller",
				mvc.NSRouter("/overview", docsController, "*:GetMvcCoreControllerOverview"),
				mvc.NSRouter("/lifecycle", docsController, "*:GetMvcCoreControllerLifecycle"),
				mvc.NSRouter("/request-handling", docsController, "*:GetMvcCoreControllerRequestHandling"),
				mvc.NSRouter("/response-output", docsController, "*:GetMvcCoreControllerResponseOutput"),
				mvc.NSRouter("/template-system", docsController, "*:GetMvcCoreControllerTemplateSystem"),
				mvc.NSRouter("/session-cookie", docsController, "*:GetMvcCoreControllerSessionCookie"),
				mvc.NSRouter("/routing-management", docsController, "*:GetMvcCoreControllerRoutingManagement"),
				mvc.NSRouter("/security-features", docsController, "*:GetMvcCoreControllerSecurityFeatures"),
				mvc.NSRouter("/utility-methods", docsController, "*:GetMvcCoreControllerUtilityMethods"),
				mvc.NSRouter("/real-world-examples", docsController, "*:GetMvcCoreControllerRealWorldExamples"),
				mvc.NSRouter("/faq-best-practices", docsController, "*:GetMvcCoreControllerFaqBestPractices"),
			),
			// ============= 错误处理子路由 =============
			mvc.NSNamespace("/error-handling",
				mvc.NSRouter("/overview", docsController, "*:GetMvcCoreErrorHandlingOverview"),
				mvc.NSRouter("/quick-start", docsController, "*:GetMvcCoreErrorHandlingQuickStart"),
				mvc.NSRouter("/default-handlers", docsController, "*:GetMvcCoreErrorHandlingDefaultHandlers"),
				mvc.NSRouter("/custom-handlers", docsController, "*:GetMvcCoreErrorHandlingCustomHandlers"),
				mvc.NSRouter("/error-pages", docsController, "*:GetMvcCoreErrorHandlingErrorPages"),
				mvc.NSRouter("/monitoring", docsController, "*:GetMvcCoreErrorHandlingMonitoring"),
				mvc.NSRouter("/business-errors", docsController, "*:GetMvcCoreErrorHandlingBusinessErrors"),
				mvc.NSRouter("/recovery", docsController, "*:GetMvcCoreErrorHandlingRecovery"),
				mvc.NSRouter("/best-practices", docsController, "*:GetMvcCoreErrorHandlingBestPractices"),
				mvc.NSRouter("/troubleshooting", docsController, "*:GetMvcCoreErrorHandlingTroubleshooting"),
			),
			mvc.NSRouter("/routing", docsController, "*:GetMvcCoreRouting"),
			mvc.NSRouter("/static-path-configuration", docsController, "*:GetMvcCoreStaticPathConfiguration"),
			mvc.NSRouter("/namespace", docsController, "*:GetMvcCoreNamespace"),
			mvc.NSRouter("/annotation", docsController, "*:GetMvcCoreAnnotation"),
			mvc.NSRouter("/comment", docsController, "*:GetMvcCoreComment"),
		),
		// ============= 中间件 =============
		mvc.NSNamespace("/middleware",
			mvc.NSRouter("/overview", docsController, "*:GetMiddlewareOverview"),
			mvc.NSRouter("/builtin", docsController, "*:GetMiddlewareBuiltin"),
			mvc.NSRouter("/custom", docsController, "*:GetMiddlewareCustom"),
			mvc.NSRouter("/config", docsController, "*:GetMiddlewareConfig"),
		),
		// ============= 统一ORM =============
		mvc.NSNamespace("/data-access",
			// 统一ORM概览 - 新增的主要入口
			mvc.NSRouter("/orm-unified", docsController, "*:GetDataAccessOrmUnified"),
			// GORM快速开始 - 新的路由
			mvc.NSRouter("/gorm-quickstart", docsController, "*:GetDataAccessGormQuickstart"),
			// 保持向后兼容的旧GORM路由
			mvc.NSRouter("/gorm", docsController, "*:GetDataAccessGorm"),
			// MyBatis 路由 - 新的分离式结构
			mvc.NSRouter("/mybatis-basic", docsController, "*:GetDataAccessMybatisBasic"),
			mvc.NSRouter("/mybatis-advanced", docsController, "*:GetDataAccessMybatisAdvanced"),
			mvc.NSRouter("/mybatis-performance", docsController, "*:GetDataAccessMybatisPerformance"),
			// 保持向后兼容的旧路由
			mvc.NSRouter("/mybatis", docsController, "*:GetDataAccessMybatis"),
			mvc.NSRouter("/database-config", docsController, "*:GetDataAccessDatabaseConfig"),
			mvc.NSRouter("/transaction", docsController, "*:GetDataAccessTransaction"),
			// 新增的文档路由
			mvc.NSRouter("/database-tuning", docsController, "*:GetDataAccessDatabaseTuning"),
			mvc.NSRouter("/caching-strategies", docsController, "*:GetDataAccessCachingStrategies"),
			mvc.NSRouter("/monitoring-alerting", docsController, "*:GetDataAccessMonitoringAlerting"),
			mvc.NSRouter("/object-mapping", docsController, "*:GetDataAccessObjectMapping"),
		),
		// ============= 视图渲染 =============
		mvc.NSNamespace("/view-template",
			mvc.NSRouter("/overview", docsController, "*:GetViewTemplateOverview"),
			mvc.NSRouter("/template-engine", docsController, "*:GetViewTemplateTemplateEngine"),
			mvc.NSRouter("/view-rendering", docsController, "*:GetViewTemplateViewRendering"),
			mvc.NSRouter("/static-assets", docsController, "*:GetViewTemplateStaticAssets"),
		),
		// ============= 配置管理 =============
		mvc.NSNamespace("/configuration",
			mvc.NSRouter("/app-config", docsController, "*:GetConfigurationAppConfig"),
			mvc.NSRouter("/environment", docsController, "*:GetConfigurationEnvironment"),
			mvc.NSRouter("/logging", docsController, "*:GetConfigurationLogging"),
		),
		// ============= 部署运维 =============
		mvc.NSNamespace("/deployment",
			mvc.NSRouter("/overview", docsController, "*:GetDeploymentOverview"),
			mvc.NSRouter("/docker", docsController, "*:GetDeploymentDocker"),
			mvc.NSRouter("/kubernetes", docsController, "*:GetDeploymentKubernetes"),
			mvc.NSRouter("/monitoring", docsController, "*:GetDeploymentMonitoring"),
		),
		// ============= 高级功能 =============
		mvc.NSNamespace("/advanced",
			mvc.NSRouter("/session", docsController, "*:GetAdvancedSession"),
			mvc.NSRouter("/cache", docsController, "*:GetAdvancedCache"),
			mvc.NSRouter("/validation", docsController, "*:GetAdvancedValidation"),
			mvc.NSRouter("/captcha", docsController, "*:GetAdvancedCaptcha"),
			mvc.NSRouter("/scheduler", docsController, "*:GetAdvancedScheduler"),
		),
		// ============= 开发工具 =============
		mvc.NSNamespace("/dev-tools",
			mvc.NSRouter("/codegen", docsController, "*:GetDevToolsCodegen"),
			mvc.NSRouter("/hot-reload", docsController, "*:GetDevToolsHotReload"),
			mvc.NSRouter("/performance", docsController, "*:GetDevToolsPerformance"),
			mvc.NSRouter("/testing", docsController, "*:GetDevToolsTesting"),
		),
	)

	// 使用Beego风格的Namespace功能
	nsApi := mvc.NewNamespace("/api",
		// 自动路由注册 - 页面控制器
		mvc.NSAutoRouter(homeController),

		// 手动路由映射 - 认证相关
		mvc.NSRouter("/auth/token", userController, "*:GetInfo"),
		mvc.NSRouter("/auth/tokenx", userController, "*:GetIndex"),
		mvc.NSRouter("/auth/refresh", userController, "*:PostCreate"),
		mvc.NSRouter("/dict/get", adminController, "*:GetDashboard"),

		// 地区命名空间
		mvc.NSNamespace("/area",
			mvc.NSRouter("/province", userController, "GET:GetInfo"),
			mvc.NSRouter("/city", userController, "POST:PostCreate"),
			mvc.NSRouter("/county", userController, "PUT:PutUpdate"),
		),
	)

	// 添加命名空间到全局应用
	mvc.AddNamespace(nsDocs)
	mvc.AddNamespace(nsApi)

	log.Println("🚀 YYHertz Namespace功能演示启动成功!")
	log.Println("📍 服务器地址: http://localhost:8888")
	log.Println("")
	log.Println("📋 Beego风格的Namespace路由:")
	log.Println("GET    /api/auth/token           - 获取Token")
	log.Println("POST   /api/auth/refresh         - 刷新Token")
	log.Println("GET    /api/area/province        - 获取省份")
	log.Println("POST   /api/area/city            - 添加城市")
	log.Println("PUT    /api/area/county          - 更新县区")
	log.Println("POST   /api/student/register     - 学生注册")
	log.Println("POST   /api/student/login        - 学生登录")
	log.Println("POST   /api/student/logout       - 学生注销")
	log.Println("GET    /api/student/profile      - 学生资料")
	log.Println("POST   /api/teacher/register     - 教师注册")
	log.Println("POST   /api/teacher/login        - 教师登录")
	log.Println("POST   /api/teacher/logout       - 教师注销")
	log.Println("GET    /api/teacher/profile      - 教师资料")
	log.Println("GET    /api/online/heartbeat     - 心跳检测")
	log.Println("GET    /api/online/status        - 在线状态")
	log.Println("POST   /api/task/clean           - 清理任务")
	log.Println("POST   /api/task/backup          - 备份任务")
	log.Println("")

	app.Run()
}
