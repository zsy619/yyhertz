package main

import (
	"fmt"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// OnlineController 在线心跳控制器
type OnlineController struct {
	core.BaseController
}

// Index 处理心跳 WebSocket 连接
func (c *OnlineController) Index(conn *define.WsConn) {
	fmt.Printf("心跳连接建立: %s\n", conn.RemoteAddr())

	// 发送欢迎消息
	err := conn.WriteMessage(1, []byte("心跳服务已连接 - 嵌套命名空间正常工作！"))
	if err != nil {
		fmt.Printf("发送欢迎消息失败: %v\n", err)
		return
	}

	// 简单的心跳处理循环
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("心跳连接断开: %v\n", err)
			break
		}

		fmt.Printf("收到心跳消息: %s\n", message)

		// 根据消息类型回复不同内容
		var response string
		switch string(message) {
		case "ping":
			response = "pong"
		case "heartbeat":
			response = "alive"
		case "test":
			response = "嵌套命名空间 WebSocket 路由工作正常！"
		default:
			response = fmt.Sprintf("心跳回复: %s (来自 /admin/online/heartbeat)", string(message))
		}

		err = conn.WriteMessage(1, []byte(response))
		if err != nil {
			fmt.Printf("发送回复失败: %v\n", err)
			break
		}
	}
}

// ChatController 聊天控制器
type ChatController struct {
	core.BaseController
}

// Room 聊天室处理器
func (c *ChatController) Room(conn *define.WsConn) {
	fmt.Printf("聊天室连接建立: %s\n", conn.RemoteAddr())

	// 发送欢迎消息
	err := conn.WriteMessage(1, []byte("欢迎进入聊天室！"))
	if err != nil {
		fmt.Printf("发送欢迎消息失败: %v\n", err)
		return
	}

	// 聊天消息处理循环
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("聊天连接断开: %v\n", err)
			break
		}

		fmt.Printf("聊天消息: %s\n", message)

		// 模拟聊天室广播（实际项目中会广播给所有用户）
		response := fmt.Sprintf("聊天室广播: %s", string(message))
		err = conn.WriteMessage(1, []byte(response))
		if err != nil {
			fmt.Printf("发送聊天消息失败: %v\n", err)
			break
		}
	}
}

// Welcome 首页欢迎页面
func (c *OnlineController) Welcome() {
	c.TplName = "welcome.html"
	c.Data["title"] = "WebSocket 测试服务器"
	c.Data["message"] = "欢迎使用 YYHertz WebSocket 测试服务器"
	c.Data["endpoints"] = []string{
		"ws://localhost:8888/admin/online/heartbeat",
		"ws://localhost:8888/ws/simple",
		"ws://localhost:8888/api/v1/websocket/test",
		"ws://localhost:8888/chat/room/general",
	}
}

func main_server() {
	fmt.Println("YYHertz WebSocket 服务器启动中...")

	// 获取应用实例
	app := mvc.HertzApp

	// ============= 设置静态文件服务 =============
	fmt.Println("设置静态文件服务...")
	app.SetStaticPath("./static")

	// ============= 嵌套命名空间测试 =============
	fmt.Println("注册嵌套命名空间 WebSocket 路由...")

	// 这是用户报告的问题场景：/admin/online/heartbeat
	nsOnline := mvc.NewNamespace("/admin",
		mvc.NSNamespace("/online",
			mvc.NSRouterWs("/heartbeat", &OnlineController{}, "Index"),
		),
	)

	// 添加命名空间
	mvc.AddNamespace(nsOnline)

	// ============= 其他测试路由 =============
	fmt.Println("注册其他测试路由...")

	// 简单的 WebSocket 路由
	app.RouterWs("/ws/simple", &OnlineController{}, "Index")

	// 单级命名空间
	wsNs := mvc.NewNamespace("/api/v1",
		mvc.NSRouterWs("/websocket/test", &OnlineController{}, "Index"),
	)
	mvc.AddNamespace(wsNs)

	// 聊天室命名空间
	chatNs := mvc.NewNamespace("/chat",
		mvc.NSNamespace("/room",
			mvc.NSRouterWs("/general", &ChatController{}, "Room"),
		),
	)
	mvc.AddNamespace(chatNs)

	// ============= 首页路由 =============
	homeController := &OnlineController{}
	app.RouterPrefix("/", homeController, true, "Welcome", "*:/")

	// ============= 启动服务器 =============
	fmt.Println("WebSocket 服务器启动完成！")
	fmt.Println("")
	fmt.Println("可用的 WebSocket 端点:")
	fmt.Println("  1. ws://localhost:8888/admin/online/heartbeat  (嵌套命名空间 - 主要测试)")
	fmt.Println("  2. ws://localhost:8888/ws/simple               (简单路由)")
	fmt.Println("  3. ws://localhost:8888/api/v1/websocket/test   (单级命名空间)")
	fmt.Println("  4. ws://localhost:8888/chat/room/general       (聊天室嵌套)")
	fmt.Println("")
	fmt.Println("测试页面:")
	fmt.Println("  - http://localhost:8888/static/websocket_test.html")
	fmt.Println("  - http://localhost:8888/static/nested_websocket_test.html")
	fmt.Println("")
	fmt.Println("使用方法:")
	fmt.Println("1. 在浏览器中打开测试页面")
	fmt.Println("2. 点击连接按钮测试不同的 WebSocket 端点")
	fmt.Println("3. 发送消息测试双向通信")
	fmt.Println("")

	// 启动服务器在端口 8888
	fmt.Printf("服务器启动在端口 8888...\n")
	app.Run()
}
