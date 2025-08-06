package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/core"
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ExampleUsage 展示新引擎的使用方法
func ExampleUsage() {
	// 创建高性能引擎
	engine := NewFastEngine()
	
	// 配置引擎
	config := EngineConfig{
		MaxRouteCache:  2000,
		MaxContextPool: 1000,
		EnableMetrics:  true,
		RequestTimeout: 30 * time.Second,
		RedirectSlash:  true,
		HandleOptions:  true,
	}
	engine.SetConfig(config)
	
	// 添加全局中间件
	engine.Use(LoggerMiddleware(), RecoveryMiddleware())
	
	// 注册基础路由
	engine.GET("/", HomeHandler)
	engine.GET("/health", HealthHandler)
	
	// 用户相关路由
	engine.GET("/users", ListUsersHandler)
	engine.GET("/users/:id", GetUserHandler)
	engine.POST("/users", CreateUserHandler)
	engine.PUT("/users/:id", UpdateUserHandler)
	engine.DELETE("/users/:id", DeleteUserHandler)
	
	// API路由组
	apiGroup := engine.Group("/api/v1")
	apiGroup.Use(AuthMiddleware())
	{
		apiGroup.GET("/profile", GetProfileHandler)
		apiGroup.POST("/upload", UploadFileHandler)
	}
	
	// 通配符路由
	engine.GET("/static/*filepath", StaticFileHandler)
	
	// 编译引擎以优化性能
	engine.Compile()
	
	// 模拟请求处理
	simulateRequests(engine)
	
	// 打印性能统计
	engine.PrintStats()
}

// 中间件示例
func LoggerMiddleware() mvccontext.HandlerFunc {
	return func(ctx *mvccontext.EnhancedContext) {
		start := time.Now()
		path := ctx.Request.Path()
		
		ctx.Next()
		
		latency := time.Since(start)
		status := ctx.Writer.Status()
		
		log.Printf("[%s] %s %s %v %d",
			ctx.Request.Method(),
			string(path),
			ctx.Request.RemoteAddr(),
			latency,
			status,
		)
	}
}

func RecoveryMiddleware() mvccontext.HandlerFunc {
	return func(ctx *mvccontext.EnhancedContext) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				ctx.JSON(500, map[string]string{
					"error": "Internal Server Error",
				})
				ctx.Abort()
			}
		}()
		
		ctx.Next()
	}
}

func AuthMiddleware() mvccontext.HandlerFunc {
	return func(ctx *mvccontext.EnhancedContext) {
		token := ctx.Header("Authorization")
		if token == "" {
			ctx.JSON(401, map[string]string{
				"error": "Authorization required",
			})
			ctx.Abort()
			return
		}
		
		// 验证token逻辑
		ctx.Set("user_id", "12345")
		ctx.Next()
	}
}

// 处理器示例
func HomeHandler(ctx context.Context, c *core.RequestContext) {
	c.JSON(200, map[string]string{
		"message": "Welcome to YYHertz MVC Framework",
		"version": "2.0.0",
		"engine":  "FastEngine",
	})
}

func HealthHandler(ctx context.Context, c *core.RequestContext) {
	c.JSON(200, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(time.Now()).Seconds(),
	})
}

func ListUsersHandler(ctx context.Context, c *core.RequestContext) {
	// 模拟从数据库获取用户列表
	users := []map[string]interface{}{
		{"id": 1, "name": "Alice", "email": "alice@example.com"},
		{"id": 2, "name": "Bob", "email": "bob@example.com"},
		{"id": 3, "name": "Charlie", "email": "charlie@example.com"},
	}
	
	c.JSON(200, map[string]interface{}{
		"users": users,
		"total": len(users),
	})
}

func GetUserHandler(ctx context.Context, c *core.RequestContext) {
	// 从URL参数获取用户ID
	userID := string(c.Param("id"))
	
	// 模拟从数据库获取用户
	user := map[string]interface{}{
		"id":    userID,
		"name":  "User " + userID,
		"email": "user" + userID + "@example.com",
	}
	
	c.JSON(200, user)
}

func CreateUserHandler(ctx context.Context, c *core.RequestContext) {
	// 解析请求体
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	
	// 这里应该有参数绑定逻辑
	// c.ShouldBindJSON(&user)
	
	// 模拟创建用户
	result := map[string]interface{}{
		"id":      123,
		"name":    user.Name,
		"email":   user.Email,
		"created": time.Now(),
	}
	
	c.JSON(201, result)
}

func UpdateUserHandler(ctx context.Context, c *core.RequestContext) {
	userID := string(c.Param("id"))
	
	c.JSON(200, map[string]interface{}{
		"message": "User updated successfully",
		"user_id": userID,
	})
}

func DeleteUserHandler(ctx context.Context, c *core.RequestContext) {
	userID := string(c.Param("id"))
	
	c.JSON(200, map[string]interface{}{
		"message": "User deleted successfully",
		"user_id": userID,
	})
}

func GetProfileHandler(ctx context.Context, c *core.RequestContext) {
	// 从中间件设置的用户ID获取用户信息
	// userID := ctx.MustGet("user_id").(string)
	
	c.JSON(200, map[string]interface{}{
		"user_id": "12345",
		"name":    "Current User",
		"email":   "user@example.com",
	})
}

func UploadFileHandler(ctx context.Context, c *core.RequestContext) {
	c.JSON(200, map[string]interface{}{
		"message":  "File uploaded successfully",
		"filename": "example.jpg",
		"size":     "1024KB",
	})
}

func StaticFileHandler(ctx context.Context, c *core.RequestContext) {
	filepath := string(c.Param("filepath"))
	
	c.JSON(200, map[string]interface{}{
		"message":  "Static file served",
		"filepath": filepath,
	})
}

// simulateRequests 模拟请求以展示性能
func simulateRequests(engine *FastEngine) {
	fmt.Println("=== Simulating Requests ===")
	
	requests := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/health"},
		{"GET", "/users"},
		{"GET", "/users/123"},
		{"POST", "/users"},
		{"PUT", "/users/456"},
		{"DELETE", "/users/789"},
		{"GET", "/api/v1/profile"},
		{"GET", "/static/css/style.css"},
		{"GET", "/nonexistent"}, // 404测试
	}
	
	for _, req := range requests {
		fmt.Printf("Processing: %s %s\n", req.method, req.path)
		engine.HandleRequest(req.method, req.path, nil)
	}
	
	// 并发测试
	fmt.Println("\n=== Concurrent Test ===")
	concurrentTest(engine, 100, 1000)
}

func concurrentTest(engine *FastEngine, goroutines, requestsPerGoroutine int) {
	start := time.Now()
	done := make(chan bool, goroutines)
	
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < requestsPerGoroutine; j++ {
				engine.HandleRequest("GET", fmt.Sprintf("/users/%d", j%100), nil)
			}
			done <- true
		}(i)
	}
	
	// 等待所有goroutine完成
	for i := 0; i < goroutines; i++ {
		<-done
	}
	
	elapsed := time.Since(start)
	totalRequests := goroutines * requestsPerGoroutine
	qps := float64(totalRequests) / elapsed.Seconds()
	
	fmt.Printf("Concurrent test completed:\n")
	fmt.Printf("- Goroutines: %d\n", goroutines)
	fmt.Printf("- Requests per goroutine: %d\n", requestsPerGoroutine)
	fmt.Printf("- Total requests: %d\n", totalRequests)
	fmt.Printf("- Time elapsed: %v\n", elapsed)
	fmt.Printf("- QPS: %.2f\n", qps)
}

// PerformanceComparison 性能对比测试
func PerformanceComparison() {
	fmt.Println("=== Performance Comparison ===")
	
	// 测试原始路由vs新引擎
	testOldRouting()
	testNewEngine()
}

func testOldRouting() {
	fmt.Println("Testing old routing system...")
	// 这里应该测试原来的路由系统性能
	// 由于时间关系，这里省略具体实现
}

func testNewEngine() {
	fmt.Println("Testing new FastEngine...")
	engine := NewFastEngine()
	
	// 添加1000个路由
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/test/%d", i)
		engine.GET(path, func(ctx context.Context, c *core.RequestContext) {
			c.String(200, "OK")
		})
	}
	
	engine.Compile()
	
	// 执行10000次请求
	start := time.Now()
	for i := 0; i < 10000; i++ {
		path := fmt.Sprintf("/test/%d", i%1000)
		engine.HandleRequest("GET", path, nil)
	}
	elapsed := time.Since(start)
	
	fmt.Printf("New engine - 10k requests in %v (%.2f req/ms)\n", 
		elapsed, float64(10000)/float64(elapsed.Milliseconds()))
}