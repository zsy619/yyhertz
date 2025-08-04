package main

import (
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/example/simple/controllers"
	"github.com/zsy619/yyhertz/framework/middleware"
	"github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
	// 创建应用实例
	app := mvc.HertzApp

	// 修正框架的静态文件路径问题
	app.StaticPath = "./static"

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

	// 自动注册路由 (使用新的AutoRouters方法)
	app.AutoRouters(homeController, userController, adminController, markdownController)

	app.RouterPrefix("/", homeController, "GetIndex", "*:/")
	app.RouterPrefix("/", markdownController, "GetIndex", "*:/")

	fmt.Println("🚀 YYHertz Namespace功能演示启动...", homeController.GetControllerName())
	fmt.Println("		0000🚀🚀🚀 ", homeController.GetControllerName())
	fmt.Println("		0000🚀🚀🚀 ", userController.GetControllerName())
	fmt.Println("		0000🚀🚀🚀 ", adminController.GetControllerName())

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

		// 学生管理命名空间
		mvc.NSNamespace("/student",
			mvc.NSRouter("/register", userController, "*:PostCreate"),
			mvc.NSRouter("/login", userController, "POST:GetInfo"),
			mvc.NSRouter("/logout", userController, "POST:DeleteRemove"),
			mvc.NSRouter("/profile", userController, "GET:GetInfo"),
		),

		// 教师管理命名空间
		mvc.NSNamespace("/teacher",
			mvc.NSRouter("/register", adminController, "*:GetSettings"),
			mvc.NSRouter("/login", adminController, "POST:PostSettings"),
			mvc.NSRouter("/logout", adminController, "POST:GetUsers"),
			mvc.NSRouter("/profile", adminController, "GET:GetDashboard"),
		),

		// 在线功能
		mvc.NSNamespace("/online",
			mvc.NSRouter("/heartbeat", homeController, "*:GetIndex"),
			mvc.NSRouter("/status", homeController, "GET:GetAbout"),
		),

		// 任务管理
		mvc.NSNamespace("/task",
			mvc.NSRouter("/clean", adminController, "*:PostClearCache"),
			mvc.NSRouter("/backup", adminController, "POST:PostSettings"),
		),
	)

	// 添加V2版本的API命名空间
	nsApiV2 := mvc.NewNamespace("/api/v2",
		// 用户管理
		mvc.NSNamespace("/users",
			mvc.NSAutoRouter(userController),
			mvc.NSRouter("/profile", userController, "GET:GetInfo"),
			mvc.NSRouter("/avatar", userController, "POST:PostCreate"),

			// 用户设置子空间
			mvc.NSNamespace("/settings",
				mvc.NSRouter("/password", userController, "PUT:PutUpdate"),
				mvc.NSRouter("/email", userController, "PUT:PutUpdate"),
				mvc.NSRouter("/preferences", userController, "GET:GetInfo"),
			),
		),

		// 管理员功能
		mvc.NSNamespace("/admin",
			mvc.NSAutoRouter(adminController),

			// 系统管理
			mvc.NSNamespace("/system",
				mvc.NSRouter("/config", adminController, "GET:GetSettings"),
				mvc.NSRouter("/config", adminController, "POST:PostSettings"),
				mvc.NSRouter("/logs", adminController, "GET:GetUsers"),
				mvc.NSRouter("/backup", adminController, "POST:PostClearCache"),
			),
		),
	)

	// 添加命名空间到全局应用
	mvc.AddNamespace(nsApi)
	mvc.AddNamespace(nsApiV2)

	fmt.Println("		8888🚀🚀🚀 ", homeController.ControllerName)

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
	log.Println("📋 API V2版本路由:")
	log.Println("GET    /api/v2/users/profile     - 用户资料")
	log.Println("POST   /api/v2/users/avatar      - 上传头像")
	log.Println("PUT    /api/v2/users/settings/password    - 修改密码")
	log.Println("PUT    /api/v2/users/settings/email       - 修改邮箱")
	log.Println("GET    /api/v2/users/settings/preferences - 获取偏好设置")
	log.Println("GET    /api/v2/admin/system/config        - 系统配置")
	log.Println("POST   /api/v2/admin/system/config        - 保存配置")
	log.Println("GET    /api/v2/admin/system/logs          - 系统日志")
	log.Println("POST   /api/v2/admin/system/backup        - 系统备份")
	log.Println("")
	log.Println("💡 测试命令:")
	log.Println("curl http://localhost:8888/api/auth/token")
	log.Println("curl http://localhost:8888/api/student/register")
	log.Println("curl http://localhost:8888/api/v2/users/profile")
	log.Println("curl http://localhost:8888/api/v2/admin/system/config")

	app.Run(":8890")
}
