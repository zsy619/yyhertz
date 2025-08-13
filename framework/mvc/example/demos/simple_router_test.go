package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
)

func TestSimpleRouter(t *testing.T) {
	// 初始化应用
	app := mvc.GetAppInstance()
	mvc.HertzApp = app

	fmt.Println("🚀 YYHertz 简化API测试启动...")

	// ============= 新的简化API测试 =============

	// 1. 测试简化GET路由
	mvc.SimpleGET("/simple/users", func(ctx context.Context) {
		c := mvc.FromContext(ctx)
		c.JSON(200, map[string]any{
			"message": "简化GET API正常工作",
			"users": []map[string]any{
				{"id": 1, "name": "张三"},
				{"id": 2, "name": "李四"},
			},
			"time": time.Now(),
		})
	})

	// 2. 测试简化POST路由
	mvc.SimplePOST("/simple/users", func(ctx context.Context) {
		c := mvc.FromContext(ctx)
		c.JSON(201, map[string]any{
			"message": "简化POST API正常工作",
			"created": map[string]any{
				"id":   123,
				"name": "新用户",
			},
			"time": time.Now(),
		})
	})

	// 3. 测试参数获取
	mvc.SimpleGET("/simple/users/:id", func(ctx context.Context) {
		c := mvc.FromContext(ctx)
		id := c.Param("id")
		c.JSON(200, map[string]any{
			"message": "简化参数获取正常工作",
			"user_id": id,
			"user": map[string]any{
				"id":   id,
				"name": "用户" + id,
			},
			"time": time.Now(),
		})
	})

	// 4. 测试查询参数
	mvc.SimpleGET("/simple/search", func(ctx context.Context) {
		c := mvc.FromContext(ctx)
		keyword := c.Query("q")
		limit := c.Query("limit")
		c.JSON(200, map[string]any{
			"message": "简化查询参数获取正常工作",
			"keyword": keyword,
			"limit":   limit,
			"results": []string{"结果1", "结果2"},
			"time":    time.Now(),
		})
	})

	// 5. 测试错误处理和中断
	mvc.SimpleGET("/simple/error", func(ctx context.Context) {
		c := mvc.FromContext(ctx)

		// 模拟权限检查
		if c.Header("Authorization") == "" {
			c.JSON(401, map[string]any{
				"error": "未授权访问",
				"code":  401,
			})
			c.Abort() // 中断请求处理
			return
		}

		c.JSON(200, map[string]any{
			"message": "授权成功",
			"time":    time.Now(),
		})
	})

	// 6. 测试PUT路由
	mvc.SimplePUT("/simple/users/:id", func(ctx context.Context) {
		c := mvc.FromContext(ctx)
		id := c.Param("id")
		c.JSON(200, map[string]any{
			"message": "简化PUT API正常工作",
			"updated_user": map[string]any{
				"id":   id,
				"name": "更新的用户" + id,
			},
			"time": time.Now(),
		})
	})

	// 7. 测试DELETE路由
	mvc.SimpleDELETE("/simple/users/:id", func(ctx context.Context) {
		c := mvc.FromContext(ctx)
		id := c.Param("id")
		c.JSON(200, map[string]any{
			"message":         "简化DELETE API正常工作",
			"deleted_user_id": id,
			"time":            time.Now(),
		})
	})

	// 8. 测试ANY路由
	mvc.SimpleAny("/simple/health", func(ctx context.Context) {
		c := mvc.FromContext(ctx)
		method := string(c.Request().Method())
		c.JSON(200, map[string]any{
			"message": "简化ANY API正常工作",
			"method":  method,
			"status":  "健康",
			"time":    time.Now(),
		})
	})

	// ============= 向后兼容性测试 =============

	// 9. 测试原有API仍然工作
	mvc.GET("/legacy/test", func(ctx context.Context, c *mvc.RequestContext) {
		c.JSON(200, map[string]any{
			"message": "原有API仍然正常工作",
			"time":    time.Now(),
		})
	})

	// ============= 复杂场景测试 =============

	// 10. 测试复杂业务逻辑
	mvc.SimpleGET("/simple/complex", func(ctx context.Context) {
		c := mvc.FromContext(ctx)

		// 获取各种参数
		userType := c.Query("type")
		page := c.Query("page")

		// 设置上下文值
		c.Set("request_id", fmt.Sprintf("req_%d", time.Now().Unix()))

		// 获取设置的值
		requestID, _ := c.Get("request_id")

		// 复杂响应
		response := map[string]any{
			"message":    "复杂业务逻辑处理成功",
			"request_id": requestID,
			"parameters": map[string]any{
				"type": userType,
				"page": page,
			},
			"processing_info": map[string]any{
				"processed_at": time.Now(),
				"handler":      "SimpleHandlerFunc",
				"context_type": "mvc.FromContext",
			},
		}

		c.JSON(200, response)
	})

	fmt.Println("📍 新的简化API路由:")
	fmt.Println("GET    /simple/users                - 简化GET API测试")
	fmt.Println("POST   /simple/users                - 简化POST API测试")
	fmt.Println("GET    /simple/users/:id            - 简化参数获取测试")
	fmt.Println("GET    /simple/search?q=关键词&limit=10 - 简化查询参数测试")
	fmt.Println("GET    /simple/error               - 错误处理和中断测试")
	fmt.Println("PUT    /simple/users/:id            - 简化PUT API测试")
	fmt.Println("DELETE /simple/users/:id            - 简化DELETE API测试")
	fmt.Println("ANY    /simple/health               - 简化ANY API测试")
	fmt.Println("GET    /simple/complex?type=admin&page=1 - 复杂业务逻辑测试")
	fmt.Println("")
	fmt.Println("📍 向后兼容性路由:")
	fmt.Println("GET    /legacy/test                - 原有API兼容性测试")
	fmt.Println("")
	fmt.Println("🔧 新功能特性:")
	fmt.Println("✅ 简化API: func(context.Context)")
	fmt.Println("✅ Context传递: mvc.FromContext(ctx)")
	fmt.Println("✅ 完整功能: 参数、查询、响应、中断等")
	fmt.Println("✅ 向后兼容: 原有API依然可用")
	fmt.Println("✅ 过滤器集成: 自动执行5层过滤器")
	fmt.Println("✅ 类型安全: 避免手动类型转换")
	fmt.Println("")

	app.Run()
}

// 示例中间件函数
func loggingMiddleware(ctx context.Context) {
	c := mvc.FromContext(ctx)
	start := time.Now()

	// 记录请求开始
	method := string(c.Request().Method())
	path := string(c.Request().Path())
	fmt.Printf("[%s] %s - 开始处理\n", method, path)

	// 模拟处理时间
	defer func() {
		duration := time.Since(start)
		fmt.Printf("[%s] %s - 处理完成，耗时: %v\n", method, path, duration)
	}()
}
