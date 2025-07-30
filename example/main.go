package main

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/example/controllers"
	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/middleware"
	"github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
	// 创建增强的日志配置
	logConfig := &config.LogConfig{
		Level:           config.LogLevelDebug,
		Format:          config.LogFormatJSON,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "logs/hertz-mvc.log",
		MaxSize:         50,
		MaxAge:          7,
		MaxBackups:      5,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"service": "hertz-mvc-framework",
			"version": "1.0.0",
			"env":     "demo",
		},
	}

	// 使用增强日志配置创建应用
	app := mvc.NewAppWithLogConfig(logConfig)

	// 配置视图和静态文件路径
	app.SetViewPath("example/views")
	app.SetStaticPath("example/static")

	// 配置增强的日志中间件
	loggerConfig := &middleware.MiddlewareLoggerConfig{
		EnableRequestBody:  true,  // 启用请求体记录用于演示
		EnableResponseBody: false, // 不记录响应体以提高性能
		SkipPaths:          []string{"/health", "/ping"},
		MaxBodySize:        512, // 限制记录的请求体大小
	}

	// 添加全局中间件
	app.Use(
		middleware.RecoveryMiddleware(),
		middleware.TracingMiddleware(),
		middleware.LoggerMiddlewareWithConfig(loggerConfig),
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

	// API路由组 - 使用现有的方法
	app.Router("/api/user", userController,
		"GetInfo", "GET:/api/user/info",
		"PostCreate", "POST:/api/user/create",
	)

	// 管理员路由组 (带权限验证)
	app.Router("/admin", adminController,
		"GetDashboard", "GET:/admin",
		"GetUsers", "GET:/admin/users",
		"GetSettings", "GET:/admin/settings",
		"PostSettings", "POST:/admin/settings",
		"PostClearCache", "POST:/admin/clear-cache",
	)

	// 演示日志功能的路由
	app.LogDebug("应用启动 - Debug级别日志")
	app.LogInfo("应用配置完成 - Info级别日志")
	app.LogWarn("这是一个警告 - Warn级别日志")

	// 健康检查路由（会被日志中间件跳过）
	app.GET("/health", func(c context.Context, ctx *mvc.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
	})

	// ping路由（也会被跳过）
	app.GET("/ping", func(c context.Context, ctx *mvc.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{"message": "pong"})
	})

	log.Println("🚀 Hertz MVC Framework with Logrus 启动成功!")
	log.Println("📍 服务器地址: http://localhost:8888")
	log.Println("📁 日志文件: logs/hertz-mvc.log")
	log.Println("")
	log.Println("📋 主要路由:")
	log.Println("GET    /                - 首页")
	log.Println("GET    /about           - 关于页面")
	log.Println("GET    /health          - 健康检查 (跳过日志)")
	log.Println("GET    /ping            - Ping检查 (跳过日志)")
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
	log.Println("✅ 链路追踪中间件")
	log.Println("✅ Logrus增强日志中间件 (含请求ID生成)")
	log.Println("✅ CORS跨域中间件")
	log.Println("✅ 限流中间件 (100次/分钟)")
	log.Println("")
	log.Println("📊 日志功能:")
	log.Println("🔍 JSON格式结构化日志")
	log.Println("🔍 请求链路追踪 (request_id)")
	log.Println("🔍 请求体记录 (最大512字节)")
	log.Println("🔍 自动日志轮转 (50MB/7天/5备份)")
	log.Println("🔍 根据HTTP状态码智能分级")
	log.Println("")
	log.Println("💡 测试命令:")
	log.Println("curl http://localhost:8888/")
	log.Println("curl http://localhost:8888/health")
	log.Println("curl http://localhost:8888/user/index")
	log.Println("curl -X POST http://localhost:8888/user/create -H 'Content-Type: application/json' -d '{\"name\":\"张三\",\"email\":\"test@example.com\"}'")
	log.Println("curl -H 'Authorization: Bearer admin-token' http://localhost:8888/admin/dashboard")
	log.Println("")
	log.Println("📖 查看日志: tail -f logs/hertz-mvc.log")

	app.Run(":8888")
}
