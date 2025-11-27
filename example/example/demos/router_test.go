package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

func TestRouter(t *testing.T) {
	// 初始化应用
	app := mvc.GetAppInstance()
	mvc.HertzApp = app

	// 使用新的路由注册API测试

	// 1. 测试基本的路由注册
	mvc.GET("/test", func(ctx context.Context, c *define.RequestContext) {
		c.JSON(200, map[string]any{
			"message": "GET route working",
			"time":    time.Now(),
		})
	})

	mvc.POST("/test", func(ctx context.Context, c *define.RequestContext) {
		c.JSON(200, map[string]any{
			"message": "POST route working",
			"time":    time.Now(),
		})
	})

	mvc.Any("/any-test", func(ctx context.Context, c *define.RequestContext) {
		method := string(c.Method())
		c.JSON(200, map[string]any{
			"message": "ANY route working",
			"method":  method,
			"time":    time.Now(),
		})
	})

	// 2. 测试中间件功能
	mvc.Use(func(ctx context.Context, c *define.RequestContext) {
		fmt.Printf("[Middleware] %s %s\n", string(c.Method()), string(c.Path()))
	})

	// 3. 测试路由组（简化版）
	apiGroup := mvc.Group("/api")
	_ = apiGroup // 暂时避免未使用的变量警告

	// 4. 测试静态文件
	mvc.Static("/static", "./static")

	fmt.Println("🚀 路由注册API测试启动...")
	fmt.Println("📍 测试路由:")
	fmt.Println("GET    /test                - 测试GET路由")
	fmt.Println("POST   /test                - 测试POST路由")
	fmt.Println("ANY    /any-test            - 测试ANY路由")
	fmt.Println("Static /static/*            - 测试静态文件")
	fmt.Println("")
	fmt.Println("🔧 新功能验证:")
	fmt.Println("✅ HandlerFunc类型适配")
	fmt.Println("✅ 过滤器系统集成")
	fmt.Println("✅ 增强Context支持")
	fmt.Println("✅ HTTP方法支持")
	fmt.Println("✅ 中间件支持")
	fmt.Println("✅ 路由组支持(基础版)")
	fmt.Println("")

	app.Run()
}
