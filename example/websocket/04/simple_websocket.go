// Package main 提供简单的 WebSocket 测试示例
package main

import (
	"fmt"
	"log"

	"github.com/hertz-contrib/websocket"
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// SimpleWebSocketController 简单的 WebSocket 测试控制器
type SimpleWebSocketController struct {
	core.BaseController
}

// Echo WebSocket Echo 处理器
func (c *SimpleWebSocketController) Echo() {
	c.HandleWebSocket(func(conn *websocket.Conn) {
		log.Printf("WebSocket Echo 连接建立")
		defer log.Printf("WebSocket Echo 连接关闭")
		
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket 错误: %v", err)
				}
				break
			}
			
			log.Printf("接收到消息: %s", string(message))
			
			// Echo 消息
			response := fmt.Sprintf("Echo: %s", string(message))
			err = conn.WriteMessage(messageType, []byte(response))
			if err != nil {
				log.Printf("发送消息错误: %v", err)
				break
			}
		}
	}, nil)
}

func main() {
	// 创建应用
	app := core.NewApp()
	
	// 测试 1: 使用命名空间注册 WebSocket 路由
	echoNS := mvc.NewNamespace("/api",
		mvc.NSWebSocket("/echo", func(conn *websocket.Conn) {
			log.Printf("命名空间 WebSocket Echo 连接建立")
			defer log.Printf("命名空间 WebSocket Echo 连接关闭")
			
			// 发送欢迎消息
			welcome := "欢迎使用 YYHertz WebSocket!"
			err := conn.WriteMessage(websocket.TextMessage, []byte(welcome))
			if err != nil {
				log.Printf("发送欢迎消息失败: %v", err)
				return
			}
			
			// 消息处理循环
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					break
				}
				
				log.Printf("命名空间收到消息: %s", string(message))
				
				// Echo 消息
				response := fmt.Sprintf("命名空间 Echo: %s", string(message))
				err = conn.WriteMessage(messageType, []byte(response))
				if err != nil {
					break
				}
			}
		}),
	)
	
	// 测试 2: 聊天室功能
	chatNS := mvc.NewNamespace("/chat",
		mvc.NSWebSocket("/room/:roomId", func(conn *websocket.Conn) {
			log.Printf("聊天室 WebSocket 连接建立")
			defer log.Printf("聊天室 WebSocket 连接关闭")
			
			// 获取全局 WebSocket 管理器
			manager := mvc.GetGlobalWebSocketManager()
			
			// 这里应该从路径参数获取房间ID，暂时使用固定值
			roomID := "room1"
			connID := fmt.Sprintf("conn_%d", conn.RemoteAddr())
			
			// 加入房间
			manager.AddConnection(connID, conn)
			manager.JoinRoom(connID, roomID)
			defer func() {
				manager.LeaveRoom(connID, roomID)
				manager.RemoveConnection(connID)
			}()
			
			// 广播加入消息
			joinMsg := fmt.Sprintf("用户 %s 加入了房间", connID)
			manager.BroadcastToRoom(roomID, websocket.TextMessage, []byte(joinMsg))
			
			// 消息处理循环
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					break
				}
				
				log.Printf("聊天室收到消息: %s", string(message))
				
				// 广播消息到房间
				broadcastMsg := fmt.Sprintf("%s: %s", connID, string(message))
				manager.BroadcastToRoom(roomID, messageType, []byte(broadcastMsg))
			}
		}),
	)
	
	// 测试 3: 自定义升级器
	customUpgrader := websocket.HertzUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	
	customNS := mvc.NewNamespace("/custom",
		mvc.NSWebSocketWithUpgrader("/ws", func(conn *websocket.Conn) {
			log.Printf("自定义升级器 WebSocket 连接建立")
			defer log.Printf("自定义升级器 WebSocket 连接关闭")
			
			// 发送配置信息
			configMsg := map[string]any{
				"type":        "config",
				"readBuffer":  1024,
				"writeBuffer": 1024,
				"message":     "使用自定义升级器连接成功",
			}
			
			err := conn.WriteJSON(configMsg)
			if err != nil {
				log.Printf("发送配置消息失败: %v", err)
				return
			}
			
			// JSON 消息处理
			for {
				var message map[string]any
				err := conn.ReadJSON(&message)
				if err != nil {
					break
				}
				
				log.Printf("自定义升级器收到 JSON: %+v", message)
				
				// 回显 JSON 消息
				response := map[string]any{
					"type":      "echo",
					"original":  message,
					"timestamp": fmt.Sprintf("%d", conn.NetConn().LocalAddr()),
				}
				
				err = conn.WriteJSON(response)
				if err != nil {
					break
				}
			}
		}, customUpgrader),
	)
	
	// 注册命名空间到应用
	// 注意：这里假设 App 有 RegisterNamespace 方法，如果没有，需要添加
	fmt.Println("注册 WebSocket 命名空间...")
	
	// 如果 App 还没有 RegisterNamespace 方法，我们可以直接调用 Register
	fmt.Println("注册 Echo 命名空间...")
	echoNS.Register(app)
	fmt.Println("注册 Chat 命名空间...")
	chatNS.Register(app)
	fmt.Println("注册 Custom 命名空间...")
	customNS.Register(app)
	
	// 注册控制器路由 - 简化版本，直接用路由注册
	// 这里如果有控制器注册方法可以使用，或者创建一个简单的路由来测试控制器
	fmt.Println("  - ws://localhost:8080/simple/Echo (控制器 Echo) - 需要手动实现路由注册")
	
	// 静态文件路由可能已经在框架中设置了，所以这里注释掉
	// app.Static("/static", "./static")
	
	// 启动服务器
	fmt.Println("WebSocket 测试服务器启动中...")
	fmt.Println("可用的 WebSocket 端点:")
	fmt.Println("  - ws://localhost:8888/api/echo (命名空间 Echo)")
	fmt.Println("  - ws://localhost:8888/chat/room/room1 (命名空间聊天室)")  
	fmt.Println("  - ws://localhost:8888/custom/ws (自定义升级器)")
	fmt.Println("  - ws://localhost:8888/simple/Echo (控制器 Echo)")
	fmt.Println("")
	fmt.Println("测试方法:")
	fmt.Println("1. 使用 WebSocket 客户端连接上述地址")
	fmt.Println("2. 发送文本消息测试 Echo 功能")
	fmt.Println("3. 对于聊天室，多个连接会互相广播消息")
	fmt.Println("4. 对于自定义升级器，发送 JSON 格式消息")
	
	log.Printf("服务器启动在端口 8080")
	app.Run(":8080")
}