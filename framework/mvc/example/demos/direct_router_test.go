package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// DirectAPI演示 - 最简洁的API使用方式
func TestDirectRouter(t *testing.T) {
	// 初始化YYHertz应用
	app := core.NewApp()

	// 设置HertzApp以便路由注册
	mvc.HertzApp = app

	// ============= Direct API 演示 =============
	// 直接传递增强Context，无需额外调用 FromContext()

	// 1. DirectGET - 用户列表
	mvc.DirectGET("/direct/users", func(c *contextenhanced.Context) {
		c.JSON(200, map[string]interface{}{
			"message": "Direct GET API - 用户列表",
			"users": []map[string]interface{}{
				{"id": 1, "name": "张三", "age": 25},
				{"id": 2, "name": "李四", "age": 30},
			},
		})
	})

	// 2. DirectPOST - 创建用户
	mvc.DirectPOST("/direct/users", func(c *contextenhanced.Context) {
		// 直接使用增强Context的所有方法
		name := c.PostForm("name")
		age := c.PostForm("age")

		c.JSON(201, map[string]interface{}{
			"message": "Direct POST API - 用户创建成功",
			"data": map[string]interface{}{
				"id":   100,
				"name": name,
				"age":  age,
			},
		})
	})

	// 3. DirectPUT - 更新用户
	mvc.DirectPUT("/direct/users/:id", func(c *contextenhanced.Context) {
		userID := c.Param("id")
		name := c.PostForm("name")
		age := c.PostForm("age")

		c.JSON(200, map[string]interface{}{
			"message": "Direct PUT API - 用户更新成功",
			"user_id": userID,
			"data": map[string]interface{}{
				"name": name,
				"age":  age,
			},
		})
	})

	// 4. DirectDELETE - 删除用户
	mvc.DirectDELETE("/direct/users/:id", func(c *contextenhanced.Context) {
		userID := c.Param("id")

		c.JSON(200, map[string]interface{}{
			"message":         "Direct DELETE API - 用户删除成功",
			"deleted_user_id": userID,
		})
	})

	// 5. DirectPATCH - 部分更新用户
	mvc.DirectPATCH("/direct/users/:id/status", func(c *contextenhanced.Context) {
		userID := c.Param("id")
		status := c.PostForm("status")

		c.JSON(200, map[string]interface{}{
			"message":    "Direct PATCH API - 用户状态更新",
			"user_id":    userID,
			"new_status": status,
		})
	})

	// 6. DirectHEAD - 检查用户是否存在
	mvc.DirectHEAD("/direct/users/:id", func(c *contextenhanced.Context) {
		userID := c.Param("id")

		// HEAD请求只返回头信息，不返回body
		if userID == "404" {
			c.AbortWithStatus(404)
		} else {
			c.SetHeader("X-User-Exists", "true")
			c.String(200, "") // HEAD请求，返回空内容
		}
	})

	// 7. DirectOPTIONS - CORS预检请求
	mvc.DirectOPTIONS("/direct/users", func(c *contextenhanced.Context) {
		c.SetHeader("Access-Control-Allow-Origin", "*")
		c.SetHeader("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
		c.SetHeader("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.JSON(200, map[string]interface{}{"message": "CORS OK"})
	})

	// 8. DirectAny - 处理任意HTTP方法
	mvc.DirectAny("/direct/webhook", func(c *contextenhanced.Context) {
		c.JSON(200, map[string]interface{}{
			"message": "Direct Any API - Webhook处理",
			"status":  "received",
		})
	})

	// ============= 中间件链式调用演示 =============

	// 认证中间件
	authMiddleware := func(c *contextenhanced.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, map[string]interface{}{
				"error": "缺少Authorization头",
			})
			c.Abort()
			return
		}

		// 设置用户信息到Context
		c.Set("user_id", "12345")
		c.Set("username", "demo_user")
	}

	// 主处理函数
	protectedHandler := func(c *contextenhanced.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		c.JSON(200, map[string]interface{}{
			"message":  "Direct API 中间件演示",
			"user_id":  userID,
			"username": username,
		})
	}

	// 使用中间件的受保护路由
	mvc.DirectGET("/direct/protected", authMiddleware, protectedHandler)

	// ============= Context功能演示 =============

	mvc.DirectPOST("/direct/demo/context", func(c *contextenhanced.Context) {
		// Query参数
		page := c.Query("page")
		if page == "" {
			page = "1"
		}
		size := c.Query("size")
		if size == "" {
			size = "10"
		}

		// Form数据
		title := c.PostForm("title")
		content := c.PostForm("content")

		// 响应数据
		c.JSON(200, map[string]interface{}{
			"query_params": map[string]string{
				"page": page,
				"size": size,
			},
			"form_data": map[string]string{
				"title":   title,
				"content": content,
			},
			"message": "Direct API Context功能演示完成",
		})
	})

	// ============= 错误处理演示 =============

	mvc.DirectGET("/direct/demo/error", func(c *contextenhanced.Context) {
		errorType := c.Query("type")
		if errorType == "" {
			errorType = "none"
		}

		switch errorType {
		case "400":
			c.JSON(400, map[string]interface{}{
				"error":   "Bad Request",
				"message": "请求参数有误",
			})
		case "404":
			c.JSON(404, map[string]interface{}{
				"error":   "Not Found",
				"message": "资源不存在",
			})
		case "500":
			c.JSON(500, map[string]interface{}{
				"error":   "Internal Server Error",
				"message": "服务器内部错误",
			})
		default:
			c.JSON(200, map[string]interface{}{
				"message": "Direct API 正常响应",
				"tip":     "使用 ?type=400/404/500 测试错误响应",
			})
		}
	})

	// ============= 与其他API的对比演示 =============

	// 原始API (复杂)
	mvc.GET("/compare/original", func(ctx context.Context, rc *core.RequestContext) {
		// 原始API需要手动处理Context
		fmt.Println("原始API调用")
		// 这里就简化了，不做复杂转换
	})

	// 简化API (需要FromContext)
	mvc.SimpleGET("/compare/simple", func(ctx context.Context) {
		c := mvc.FromContext(ctx)
		if c != nil {
			c.JSON(200, map[string]interface{}{
				"message":  "简化API - 需要FromContext()获取增强Context",
				"api_type": "simple",
			})
		}
	})

	// Direct API (最简洁)
	mvc.DirectGET("/compare/direct", func(c *contextenhanced.Context) {
		c.JSON(200, map[string]interface{}{
			"message":  "Direct API - 直接使用增强Context，最简洁",
			"api_type": "direct",
		})
	})

	// ============= 启动信息 =============
	fmt.Println("🚀 YYHertz Direct API Demo 启动成功!")
	fmt.Println("📍 监听地址: http://localhost:8080")
	fmt.Println("\n🌟 Direct API 测试路由:")
	fmt.Println("  GET    /direct/users          - 获取用户列表")
	fmt.Println("  POST   /direct/users          - 创建用户")
	fmt.Println("  PUT    /direct/users/:id      - 更新用户")
	fmt.Println("  DELETE /direct/users/:id      - 删除用户")
	fmt.Println("  PATCH  /direct/users/:id/status - 更新用户状态")
	fmt.Println("  HEAD   /direct/users/:id      - 检查用户存在")
	fmt.Println("  GET    /direct/protected      - 需要认证的路由")
	fmt.Println("  POST   /direct/demo/context   - Context功能演示")
	fmt.Println("  GET    /direct/demo/error     - 错误处理演示")
	fmt.Println("\n📊 API 对比路由:")
	fmt.Println("  GET    /compare/original      - 原始API演示")
	fmt.Println("  GET    /compare/simple        - 简化API演示")
	fmt.Println("  GET    /compare/direct        - Direct API演示")
	fmt.Println("\n✨ Direct API 的优势:")
	fmt.Println("  - 直接传递 contextenhanced.Context，无需额外调用")
	fmt.Println("  - 与 Gin、Echo 风格类似，开发体验最佳")
	fmt.Println("  - 零性能开销，无 context.Value 查找成本")
	fmt.Println("  - 保留 YYHertz 所有高级特性")

	// 启动服务
	app.Spin()
}
