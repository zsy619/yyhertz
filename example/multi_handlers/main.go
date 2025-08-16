package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/router"
)

func main() {
	// 创建应用实例
	app := mvc.HertzApp

	// 创建支持多处理器的API路由组
	apiGroup := mvc.CreateGroup("/api/v1")

	// 现在可以直接使用所有新的处理器类型
	log.Println("🎉 多处理器路由组已集成!")
	log.Println("✅ 现在可以使用所有7种处理器类型")

	// 设置所有路由
	setupRoutes(apiGroup)

	log.Println("🚀 多处理器类型示例启动成功!")
	log.Println("📍 服务器地址: http://localhost:8888")
	log.Println("")
	log.Println("🔗 可用的API端点:")
	log.Println("GET    /api/v1/health         - 健康检查 (LightHandler)")
	log.Println("GET    /api/v1/ping           - Ping响应 (SimpleHandler)")
	log.Println("GET    /api/v1/users          - 用户列表 (ResponseHandler)")
	log.Println("GET    /api/v1/info           - 系统信息 (DirectHandler)")
	log.Println("POST   /api/v1/process        - 异步处理 (AsyncHandler)")
	log.Println("GET    /api/v1/stream         - 流式数据 (StreamHandler)")
	log.Println("GET    /api/v1/any            - 任意方法 (AnyResponse)")
	log.Println("")

	// 启动服务
	app.Run()
}

func setupRoutes(group *router.Group) {
	// 1. 轻量级处理器 - 健康检查端点
	// 特点: 无参数，最小开销，自动返回200 OK
	group.GETLight("/health", func() {
		log.Println("🟢 Health check performed")
	})

	// 2. 简单处理器 - 基础ping响应
	// 特点: 只需要context，适用于简单业务逻辑
	group.GETSimple("/ping", func(ctx context.Context) {
		log.Println("📡 Ping received at", time.Now().Format("15:04:05"))
	})

	// 3. 响应处理器 - REST API风格的用户列表
	// 特点: 返回任意数据，自动转换为JSON响应
	group.GETResponse("/users", func(c *mvcContext.Context) any {
		// 使用增强Context的功能
		c.Set("handler_type", "ResponseHandler")
		c.SetString("request_time", time.Now().Format(time.RFC3339))

		users := []map[string]any{
			{"id": 1, "name": "Alice", "email": "alice@example.com"},
			{"id": 2, "name": "Bob", "email": "bob@example.com"},
			{"id": 3, "name": "Charlie", "email": "charlie@example.com"},
		}

		return map[string]any{
			"success":          true,
			"data":             users,
			"total":            len(users),
			"message":          "用户列表获取成功",
			"enhanced_context": true,
		}
	})

	// 4. 直接处理器 - 完全控制请求和响应
	// 特点: 直接访问增强Context，最简洁的处理器类型
	group.GETDirect("/info", func(c *mvcContext.Context) {
		// 使用增强Context的方法
		c.Set("handler_type", "DirectHandler")
		c.SetBool("enhanced_features", true)

		// 通过Request()获取原始context
		reqCtx := c.Request()
		reqCtx.SetContentType("application/json")
		reqCtx.SetStatusCode(200)

		info := fmt.Sprintf(`{
			"server": "YYHertz Framework",
			"version": "1.0.0",
			"timestamp": "%s",
			"uptime": "运行中",
			"handlers": "多种处理器类型支持",
			"context_type": "Enhanced mvcContext.Context",
			"keys_count": %d
		}`, time.Now().Format(time.RFC3339), c.KeysCount())

		reqCtx.WriteString(info)
		log.Println("ℹ️ System info requested with enhanced context")
	})

	// 5. 异步处理器 - 模拟耗时操作
	// 特点: 通过channel返回结果，支持异步处理
	group.POSTAsync("/process", func(c *mvcContext.Context) <-chan any {
		// 使用增强Context的功能
		c.Set("handler_type", "AsyncHandler")
		c.SetString("start_time", time.Now().Format(time.RFC3339))

		resultChan := make(chan any, 1)

		go func() {
			defer close(resultChan)
			log.Println("⚙️ Starting async processing with enhanced context...")

			// 模拟耗时操作
			time.Sleep(2 * time.Second)
			log.Println("✅ Async processing completed")
			resultChan <- map[string]any{
				"success":          true,
				"result":           "数据处理完成",
				"processed":        1000,
				"duration":         "2s",
				"timestamp":        time.Now().Format(time.RFC3339),
				"context_enhanced": true,
				"keys_count":       c.KeysCount(),
			}
		}()

		return resultChan
	})

	// 6. 流式处理器 - 实时数据流
	// 特点: 通过channel发送数据流，适用于大数据传输
	group.GETStream("/stream", func(c *mvcContext.Context, dataChan chan<- []byte) error {
		// 使用增强Context的功能
		c.Set("handler_type", "StreamHandler")
		c.SetInt("total_chunks", 10)

		// 通过Request()设置头信息
		reqCtx := c.Request()
		reqCtx.Header("Cache-Control", "no-cache")
		reqCtx.Header("Connection", "keep-alive")
		log.Println("🌊 Starting data stream with enhanced context...")

		for i := 1; i <= 10; i++ {
			c.SetInt("current_chunk", i)
			data := fmt.Sprintf("数据块 %d: %s (Context Keys: %d)\n",
				i, time.Now().Format("15:04:05.000"), c.KeysCount())
			dataChan <- []byte(data)
			log.Printf("📦 Sent chunk %d with enhanced context", i)
			time.Sleep(500 * time.Millisecond) // 模拟数据产生间隔
		}

		log.Println("✅ Stream completed")
		return nil
	})

	// 7. Any处理器 - 支持所有HTTP方法
	// 特点: 可以响应任何HTTP方法的请求
	group.AnyResponse("/any", func(c *mvcContext.Context) any {
		// 使用增强Context的功能
		c.Set("handler_type", "AnyHandler")
		reqCtx := c.Request()
		method := string(reqCtx.Request.Header.Method())
		c.SetString("http_method", method)

		log.Printf("🔄 Received %s request to /any with enhanced context", method)

		return map[string]any{
			"message":          "支持所有HTTP方法",
			"method":           method,
			"path":             string(reqCtx.Request.Path()),
			"time":             time.Now().Format(time.RFC3339),
			"note":             "这个端点可以处理 GET, POST, PUT, DELETE 等所有方法",
			"enhanced_context": true,
			"keys_count":       c.KeysCount(),
		}
	})

	// 8. 演示不同HTTP方法的处理器组合
	group.POSTResponse("/submit", func(c *mvcContext.Context) any {
		c.Set("handler_type", "POSTResponse")
		return map[string]any{
			"success":          true,
			"message":          "数据提交成功",
			"method":           "POST",
			"enhanced_context": true,
		}
	})

	group.PUTResponse("/update", func(c *mvcContext.Context) any {
		c.Set("handler_type", "PUTResponse")
		return map[string]any{
			"success":          true,
			"message":          "数据更新成功",
			"method":           "PUT",
			"enhanced_context": true,
		}
	})

	group.DELETEResponse("/delete", func(c *mvcContext.Context) any {
		c.Set("handler_type", "DELETEResponse")
		return map[string]any{
			"success":          true,
			"message":          "数据删除成功",
			"method":           "DELETE",
			"enhanced_context": true,
		}
	})

	// 9. 展示错误处理
	group.GETResponse("/error", func(c *mvcContext.Context) any {
		c.Set("handler_type", "ErrorResponse")
		c.Request().SetStatusCode(500)
		return map[string]any{
			"success":          false,
			"error":            "这是一个模拟错误",
			"code":             500,
			"enhanced_context": true,
		}
	})

	// 10. 展示返回nil的情况
	group.POSTResponse("/noContent", func(c *mvcContext.Context) any {
		c.Set("handler_type", "NoContentResponse")
		c.SetBool("processed", true)
		log.Println("📝 Processing completed with enhanced context, no content to return")
		return nil // 会自动返回204 No Content
	})
}
