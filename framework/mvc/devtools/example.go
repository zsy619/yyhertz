package devtools

import (
	"context"
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
)

// SetupDevTools 设置开发工具
func SetupDevTools(app *mvc.App) error {
	// 1. 设置热重载
	hotReloadConfig := DefaultHotReloadConfig()
	hotReloadConfig.OnReload = func() error {
		log.Println("执行热重载...")
		// 这里可以添加自定义的重载逻辑
		// 比如重新加载配置、重新注册路由等
		return nil
	}

	hotReloader, err := NewHotReloadServer(app, hotReloadConfig)
	if err != nil {
		return err
	}

	// 2. 设置调试中间件
	debugMiddleware := NewDebugMiddleware()
	debugPanel := NewDebugPanel(debugMiddleware)

	// 3. 设置性能监控
	performanceMonitor := NewPerformanceMonitor()
	performancePanel := NewPerformancePanel(performanceMonitor)

	// 启动性能监控
	performanceMonitor.Start()

	// 注册中间件
	app.Use(debugMiddleware.Handler())
	app.Use(performanceMonitor.Middleware())

	// 注册调试和监控路由
	debugPanel.RegisterRoutes(app.Engine)
	performancePanel.RegisterRoutes(app.Engine)

	// 在开发环境下启动热重载服务器
	if isDevelopment() {
		go func() {
			if err := hotReloader.Run(); err != nil {
				log.Printf("热重载服务器错误: %v", err)
			}
		}()
	}

	log.Println("开发工具已启用:")
	log.Println("- 调试面板: http://localhost:8080/debug/panel")
	log.Println("- 性能监控: http://localhost:8080/performance/panel")
	log.Println("- 热重载: 已启用文件监控")

	return nil
}

// isDevelopment 检查是否为开发环境
func isDevelopment() bool {
	// 这里可以根据环境变量或配置文件判断
	// 简单示例，实际项目中应该有更完善的环境判断
	return true
}

// ExampleUsage 使用示例
func ExampleUsage() {
	// 创建应用
	app := mvc.NewApp()

	// 设置开发工具
	if err := SetupDevTools(app); err != nil {
		log.Fatalf("设置开发工具失败: %v", err)
	}

	// 添加一些示例路由
	app.GET("/", func(ctx context.Context, c *mvc.RequestContext) {
		c.JSON(200, map[string]interface{}{
			"message": "Hello YYHertz!",
			"time":    time.Now(),
		})
	})

	app.GET("/api/users", func(ctx context.Context, c *mvc.RequestContext) {
		// 模拟一些处理时间
		time.Sleep(50 * time.Millisecond)

		c.JSON(200, map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": 1, "name": "张三"},
				{"id": 2, "name": "李四"},
			},
		})
	})

	app.POST("/api/users", func(ctx context.Context, c *mvc.RequestContext) {
		// 模拟错误情况
		if string(c.Query("error")) == "true" {
			c.JSON(500, map[string]interface{}{
				"error": "模拟服务器错误",
			})
			return
		}

		c.JSON(201, map[string]interface{}{
			"message": "用户创建成功",
			"id":      123,
		})
	})

	// 启动服务器
	log.Println("服务器启动在 :8080")
	app.Run(":8080")
}
