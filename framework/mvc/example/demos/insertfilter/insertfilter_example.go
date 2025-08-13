package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// ExampleController 示例控制器
type ExampleController struct {
	*core.BaseController
}

func (c *ExampleController) GetIndex() {
	c.Ctx.JSON(200, map[string]any{
		"message": "Hello from Index",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (c *ExampleController) GetApi() {
	// 获取过滤器设置的用户信息
	userID, _ := c.Ctx.Get("user_id")
	requestID, _ := c.Ctx.Get("request_id")

	c.Ctx.JSON(200, map[string]any{
		"message":    "API Response",
		"user_id":    userID,
		"request_id": requestID,
		"timestamp":  time.Now().Unix(),
	})
}

func (c *ExampleController) PostApi() {
	c.Ctx.JSON(200, map[string]any{
		"message": "API POST Response",
		"status":  "success",
	})
}

func (c *ExampleController) GetAdmin() {
	c.Ctx.JSON(200, map[string]any{
		"message": "Admin Panel",
		"level":   "administrator",
	})
}

func main() {
	fmt.Println("InsertFilter 功能演示")
	fmt.Println("===================")

	// 设置各种过滤器演示不同的功能
	setupFilters()

	// 创建应用并注册控制器
	app := mvc.HertzApp
	app.AutoRouters(&ExampleController{})

	fmt.Println("\n🚀 服务器启动在端口 8080")
	fmt.Println("📝 测试URL:")
	fmt.Println("   GET  http://localhost:8080/example/")
	fmt.Println("   GET  http://localhost:8080/example/api")
	fmt.Println("   POST http://localhost:8080/example/api")
	fmt.Println("   GET  http://localhost:8080/example/admin")
	fmt.Println("   GET  http://localhost:8080/static/test.txt")

	// 启动服务器
	app.Run(":8080")
}

// setupFilters 设置演示用的各种过滤器
func setupFilters() {
	fmt.Println("\n📋 设置过滤器...")

	// 1. BeforeStatic - 静态文件处理前的过滤器
	staticFilter := func(ctx *context.Context) {
		path := string(ctx.Request().Path())
		if strings.HasPrefix(path, "/static/") {
			fmt.Printf("🗂️  [BeforeStatic] 访问静态文件: %s\n", path)
			ctx.SetHeader("X-Static-File", "true")
		}
	}
	mvc.InsertFilter("/static/*", constant.BeforeStatic, staticFilter)

	// 2. BeforeRouter - 路由匹配前的全局过滤器
	globalFilter := func(ctx *context.Context) {
		// 生成请求ID
		requestID := generateRequestID()
		ctx.Set("request_id", requestID)
		ctx.SetHeader("X-Request-ID", requestID)

		path := string(ctx.Request().Path())
		method := string(ctx.Request().Method())
		fmt.Printf("🌐 [BeforeRouter] %s %s (ID: %s)\n", method, path, requestID)

		// 记录请求开始时间
		ctx.Set("start_time", time.Now())
	}
	mvc.InsertFilter("/*", constant.BeforeRouter, globalFilter)

	// 3. 认证过滤器 - 只对API路径有效
	authFilter := func(ctx *context.Context) {
		path := string(ctx.Request().Path())
		fmt.Printf("🔐 [Auth] 检查路径权限: %s\n", path)

		// 简单的token验证演示
		token := ctx.Header("Authorization")
		if token == "" {
			// 为演示目的，我们设置一个默认用户
			ctx.Set("user_id", "demo_user")
			fmt.Printf("   ✓ 默认用户认证通过\n")
		} else {
			// 验证token（这里简化处理）
			if token == "Bearer valid_token" {
				ctx.Set("user_id", "authenticated_user")
				fmt.Printf("   ✓ Token认证通过\n")
			} else {
				fmt.Printf("   ✗ Token无效\n")
				ctx.JSON(401, map[string]string{"error": "Unauthorized"})
				ctx.Abort()
				return
			}
		}
	}
	mvc.InsertFilter("/example/api*", constant.BeforeRouter, authFilter)

	// 4. 管理员权限过滤器
	adminFilter := func(ctx *context.Context) {
		fmt.Printf("👑 [Admin] 检查管理员权限\n")

		// 检查是否有admin权限（这里简化为检查特定header）
		adminRole := ctx.Header("X-Admin-Role")
		if adminRole != "admin" {
			fmt.Printf("   ✗ 需要管理员权限\n")
			ctx.JSON(403, map[string]string{"error": "Admin access required"})
			ctx.Abort()
			return
		}

		fmt.Printf("   ✓ 管理员权限验证通过\n")
		ctx.Set("is_admin", true)
	}
	mvc.InsertFilter("/example/admin*", constant.BeforeExec, adminFilter)

	// 5. BeforeExec - 控制器执行前的日志过滤器
	preExecFilter := func(ctx *context.Context) {
		path := string(ctx.Request().Path())
		userID, _ := ctx.Get("user_id")
		fmt.Printf("📝 [BeforeExec] 即将执行控制器 (用户: %v, 路径: %s)\n", userID, path)

		// 设置一些执行前的上下文信息
		ctx.Set("exec_start", time.Now())
	}
	mvc.InsertFilter("/*", constant.BeforeExec, preExecFilter)

	// 6. AfterExec - 控制器执行后的处理
	postExecFilter := func(ctx *context.Context) {
		execStart, exists := ctx.Get("exec_start")
		if exists {
			duration := time.Since(execStart.(time.Time))
			fmt.Printf("⏱️  [AfterExec] 控制器执行完成，耗时: %v\n", duration)

			// 性能监控
			if duration > 100*time.Millisecond {
				fmt.Printf("   ⚠️  慢查询警告: 执行时间超过100ms\n")
			}
		}

		// 添加响应头
		ctx.SetHeader("X-Exec-Time", time.Now().Format(time.RFC3339))
	}
	mvc.InsertFilter("/*", constant.AfterExec, postExecFilter)

	// 7. 限流过滤器（演示）
	rateLimitFilter := func(ctx *context.Context) {
		clientIP := ctx.Request().ClientIP()
		fmt.Printf("🚦 [RateLimit] 检查IP限流: %s\n", clientIP)

		// 简化的限流检查（实际应用中应该用Redis等）
		// 这里只是演示，不做实际限制
		ctx.Set("rate_limit_checked", true)
		fmt.Printf("   ✓ 限流检查通过\n")
	}
	mvc.InsertFilter("/example/api*", constant.BeforeRouter, rateLimitFilter)

	// 8. CORS处理过滤器
	corsFilter := func(ctx *context.Context) {
		origin := ctx.Header("Origin")
		if origin != "" {
			fmt.Printf("🌍 [CORS] 处理跨域请求: %s\n", origin)
			ctx.SetHeader("Access-Control-Allow-Origin", "*")
			ctx.SetHeader("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Admin-Role")
		}

		// 处理预检请求
		if string(ctx.Request().Method()) == "OPTIONS" {
			fmt.Printf("   ✓ 处理OPTIONS预检请求\n")
			ctx.JSON(204, nil)
			ctx.Abort()
			return
		}
	}
	mvc.InsertFilter("/*", constant.BeforeRouter, corsFilter)

	// 9. FinishRouter - 请求完成后的清理和日志
	finishFilter := func(ctx *context.Context) {
		startTime, exists := ctx.Get("start_time")
		if exists {
			totalDuration := time.Since(startTime.(time.Time))
			requestID, _ := ctx.Get("request_id")
			path := string(ctx.Request().Path())
			method := string(ctx.Request().Method())

			fmt.Printf("🏁 [FinishRouter] 请求完成\n")
			fmt.Printf("   📊 总耗时: %v\n", totalDuration)
			fmt.Printf("   📍 请求: %s %s\n", method, path)
			fmt.Printf("   🆔 ID: %v\n", requestID)

			// 添加性能头
			ctx.SetHeader("X-Total-Time", totalDuration.String())
		}

		// 添加安全头
		ctx.SetHeader("X-Content-Type-Options", "nosniff")
		ctx.SetHeader("X-Frame-Options", "DENY")
	}
	mvc.InsertFilter("/*", constant.FinishRouter, finishFilter)

	// 显示过滤器统计
	showFilterStats()
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// showFilterStats 显示过滤器统计信息
func showFilterStats() {
	fmt.Println("\n📈 过滤器统计:")

	positions := []struct {
		pos  int
		name string
	}{
		{constant.BeforeStatic, "BeforeStatic"},
		{constant.BeforeRouter, "BeforeRouter"},
		{constant.BeforeExec, "BeforeExec"},
		{constant.AfterExec, "AfterExec"},
		{constant.FinishRouter, "FinishRouter"},
	}

	for _, p := range positions {
		filters := mvc.ListFilters(p.pos)
		fmt.Printf("   %s: %d 个过滤器\n", p.name, len(filters))

		for i, filter := range filters {
			status := "启用"
			if !filter.Enabled {
				status = "禁用"
			}
			fmt.Printf("     %d. 模式: %s (%s)\n", i+1, filter.Pattern, status)
		}
	}
}

// 演示如何动态管理过滤器
func demonstrateFilterManagement() {
	fmt.Println("\n🔧 过滤器动态管理演示:")

	// 动态添加临时过滤器
	tempFilter := func(ctx *context.Context) {
		fmt.Println("🎯 临时过滤器执行")
	}

	fmt.Println("   ➕ 添加临时过滤器")
	mvc.InsertFilter("/temp/*", constant.BeforeRouter, tempFilter)

	// 检查过滤器
	filters := mvc.ListFilters(constant.BeforeRouter)
	fmt.Printf("   📊 当前 BeforeRouter 过滤器数量: %d\n", len(filters))

	// 移除过滤器
	fmt.Println("   ➖ 移除临时过滤器")
	removed := mvc.RemoveFilter("/temp/*", constant.BeforeRouter)
	fmt.Printf("   ✓ 移除结果: %v\n", removed)

	// 再次检查
	filters = mvc.ListFilters(constant.BeforeRouter)
	fmt.Printf("   📊 移除后 BeforeRouter 过滤器数量: %d\n", len(filters))
}

// 模拟一些辅助函数
func validateToken(token string) bool {
	// 简化的token验证
	return token == "Bearer valid_token"
}

func checkRateLimit(ip string) bool {
	// 简化的限流检查
	return true
}

func updateRequestCount(ip string) {
	// 更新请求计数（实际应用中应该持久化）
}

// 使用说明
/*
运行此示例后，您可以使用以下命令测试：

1. 基本请求（会触发多个过滤器）:
   curl http://localhost:8080/example/

2. API请求（需要认证）:
   curl http://localhost:8080/example/api

3. 带认证的API请求:
   curl -H "Authorization: Bearer valid_token" http://localhost:8080/example/api

4. 管理员请求（需要特殊权限）:
   curl -H "X-Admin-Role: admin" http://localhost:8080/example/admin

5. CORS预检请求:
   curl -X OPTIONS -H "Origin: http://example.com" http://localhost:8080/example/api

6. POST请求:
   curl -X POST -H "Content-Type: application/json" -d '{"test":"data"}' http://localhost:8080/example/api

您将在控制台看到详细的过滤器执行日志，展示请求如何在不同阶段被处理。
*/
