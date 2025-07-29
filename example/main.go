package main

import (
	"log"
	"time"

	"hertz-controller/example/controllers"
	"hertz-controller/framework/controller"
	"hertz-controller/framework/middleware"
)

func main() {
	app := controller.NewApp()

	// 配置视图和静态文件路径
	app.SetViewPath("example/views")
	app.SetStaticPath("example/static")

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

	// 自动注册路由 (使用Include方法)
	app.Include(homeController, userController, adminController)

	// 手动注册额外的路由
	app.Router("/", homeController,
		"GetIndex", "GET:/",
		"GetAbout", "GET:/about",
		"GetDocs", "GET:/docs",
		"PostContact", "POST:/contact",
	)

	// API路由组
	app.Router("/api/user", userController,
		"GetProfile", "GET:/api/user/profile",
		"PostLogin", "POST:/api/user/login",
	)

	// 管理员路由组 (带权限验证)
	app.Router("/admin", adminController,
		"GetDashboard", "GET:/admin",
		"GetUsers", "GET:/admin/users",
		"GetSettings", "GET:/admin/settings",
		"PostSettings", "POST:/admin/settings",
		"PostClearCache", "POST:/admin/clear-cache",
	)

	log.Println("🚀 Hertz MVC Framework 启动成功!")
	log.Println("📍 服务器地址: http://localhost:8888")
	log.Println("")
	log.Println("📋 路由列表:")
	log.Println("GET    /                - 首页")
	log.Println("GET    /about           - 关于页面")
	log.Println("GET    /docs            - 文档页面")
	log.Println("POST   /contact         - 联系我们")
	log.Println("")
	log.Println("用户管理:")
	log.Println("GET    /user/index      - 用户列表")
	log.Println("GET    /user/info       - 用户详情")
	log.Println("POST   /user/create     - 创建用户")
	log.Println("PUT    /user/update     - 更新用户")
	log.Println("DELETE /user/remove     - 删除用户")
	log.Println("")
	log.Println("管理后台:")
	log.Println("GET    /admin/dashboard - 管理员面板")
	log.Println("GET    /admin/users     - 管理员用户列表")
	log.Println("GET    /admin/settings  - 系统设置")
	log.Println("POST   /admin/settings  - 保存设置")
	log.Println("")
	log.Println("🔧 已启用中间件:")
	log.Println("✅ 异常恢复中间件")
	log.Println("✅ 请求日志中间件")
	log.Println("✅ CORS跨域中间件")
	log.Println("✅ 限流中间件 (100次/分钟)")
	log.Println("")
	log.Println("💡 测试命令:")
	log.Println("curl http://localhost:8888/")
	log.Println("curl http://localhost:8888/user/index")
	log.Println("curl -X POST http://localhost:8888/user/create -d 'name=张三&email=test@example.com&password=123456'")
	log.Println("curl -H 'Authorization: Bearer admin-token' http://localhost:8888/admin/dashboard")

	app.Run(":8888")
}
