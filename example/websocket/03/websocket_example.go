// Package example 提供 WebSocket 使用示例
//
// 本文件展示了如何在 YYHertz MVC 框架中使用 WebSocket 功能，
// 包括基本的 Echo 服务器、聊天室和实时数据推送等常见场景。
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hertz-contrib/websocket"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// ============= 示例控制器 =============

// EchoController 简单的 WebSocket Echo 控制器
type EchoController struct {
	core.BaseController
}

// WebSocket 处理 WebSocket 连接 - Echo 服务
func (c *EchoController) WebSocket() {
	c.HandleWebSocket(func(conn *websocket.Conn) {
		log.Printf("Echo WebSocket connection from %s", c.Ctx.ClientIP())
		defer log.Printf("Echo connection closed")

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			log.Printf("Received: %s", string(message))

			// Echo 消息
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				log.Printf("Write error: %v", err)
				break
			}
		}
	}, nil)
}

// ============= 聊天室示例 =============

// ChatRoom 聊天室管理器
type ChatRoom struct {
	rooms map[string]map[*websocket.Conn]bool
	mutex sync.RWMutex
}

// NewChatRoom 创建聊天室管理器
func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		rooms: make(map[string]map[*websocket.Conn]bool),
	}
}

// Join 加入聊天室
func (cr *ChatRoom) Join(roomID string, conn *websocket.Conn) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if cr.rooms[roomID] == nil {
		cr.rooms[roomID] = make(map[*websocket.Conn]bool)
	}
	cr.rooms[roomID][conn] = true

	log.Printf("User joined room %s, total: %d", roomID, len(cr.rooms[roomID]))
}

// Leave 离开聊天室
func (cr *ChatRoom) Leave(roomID string, conn *websocket.Conn) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if room, exists := cr.rooms[roomID]; exists {
		delete(room, conn)
		if len(room) == 0 {
			delete(cr.rooms, roomID)
		}
		log.Printf("User left room %s", roomID)
	}
}

// Broadcast 向房间广播消息
func (cr *ChatRoom) Broadcast(roomID string, message []byte) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	if room, exists := cr.rooms[roomID]; exists {
		for conn := range room {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Broadcast error: %v", err)
				conn.Close()
				delete(room, conn)
			}
		}
	}
}

// 全局聊天室实例
var globalChatRoom = NewChatRoom()

// ChatController 聊天控制器
type ChatController struct {
	core.BaseController
}

// Room 聊天室 WebSocket 处理
func (c *ChatController) Room() {
	roomID := c.Ctx.GetParam("room")
	if roomID == "" {
		c.Abort("400", "Room ID is required")
		return
	}

	c.HandleWebSocket(func(conn *websocket.Conn) {
		log.Printf("User joined chat room %s", roomID)

		// 加入聊天室
		globalChatRoom.Join(roomID, conn)
		defer globalChatRoom.Leave(roomID, conn)

		// 发送欢迎消息
		welcomeMsg := map[string]interface{}{
			"type":    "system",
			"message": fmt.Sprintf("Welcome to room %s", roomID),
			"time":    time.Now().Unix(),
		}
		conn.WriteJSON(welcomeMsg)

		// 消息处理循环
		for {
			var message map[string]interface{}
			err := conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Chat WebSocket error: %v", err)
				}
				break
			}

			// 添加时间戳和房间信息
			message["time"] = time.Now().Unix()
			message["room"] = roomID

			// 广播消息到房间
			messageBytes, _ := json.Marshal(message)
			globalChatRoom.Broadcast(roomID, messageBytes)
		}
	}, nil)
}

// ============= 实时数据推送示例 =============

// StreamController 实时数据流控制器
type StreamController struct {
	core.BaseController
}

// Data 实时数据推送
func (c *StreamController) Data() {
	c.HandleWebSocket(func(conn *websocket.Conn) {
		c.Printf("Data stream connection from %s", c.Ctx.ClientIP())
		defer c.Printf("Data stream connection closed")

		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 模拟实时数据
				data := map[string]interface{}{
					"timestamp": time.Now().Unix(),
					"cpu":       fmt.Sprintf("%.2f%%", float64(time.Now().Second())*1.67),
					"memory":    fmt.Sprintf("%.2f%%", float64(time.Now().Minute())*1.25),
					"disk":      fmt.Sprintf("%.2f%%", float64(time.Now().Hour())*4.16),
				}

				err := conn.WriteJSON(data)
				if err != nil {
					log.Printf("Data stream write error: %v", err)
					return
				}

			default:
				// 检查连接是否关闭
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
				time.Sleep(time.Millisecond * 100)
			}
		}
	}, nil)
}

// ============= 主函数和路由设置 =============

func main() {
	// 创建应用
	app := core.NewApp()

	// ============= 使用命名空间注册 WebSocket 路由 =============

	// 基本 Echo WebSocket
	echoNS := mvc.NewNamespace("/echo",
		mvc.NSWebSocket("/ws", func(conn *websocket.Conn) {
			log.Printf("Namespace Echo WebSocket connection")
			defer log.Printf("Namespace Echo connection closed")

			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					break
				}

				// Echo with prefix
				response := fmt.Sprintf("Echo: %s", string(message))
				err = conn.WriteMessage(messageType, []byte(response))
				if err != nil {
					break
				}
			}
		}),
	)

	// 聊天室 WebSocket
	chatNS := mvc.NewNamespace("/chat",
		mvc.NSWebSocket("/room/:room", func(conn *websocket.Conn) {
			// 这里可以通过某种方式获取路径参数，暂时使用固定房间
			roomID := "general"

			log.Printf("Chat WebSocket connection to room %s", roomID)

			globalChatRoom.Join(roomID, conn)
			defer globalChatRoom.Leave(roomID, conn)

			for {
				var message map[string]interface{}
				err := conn.ReadJSON(&message)
				if err != nil {
					break
				}

				message["time"] = time.Now().Unix()
				messageBytes, _ := json.Marshal(message)
				globalChatRoom.Broadcast(roomID, messageBytes)
			}
		}),
	)

	// 实时数据推送 WebSocket
	streamNS := mvc.NewNamespace("/stream",
		mvc.NSWebSocket("/data", func(conn *websocket.Conn) {
			log.Printf("Stream WebSocket connection")
			defer log.Printf("Stream connection closed")

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					data := map[string]interface{}{
						"timestamp": time.Now().Unix(),
						"value":     time.Now().Second(),
					}

					if err := conn.WriteJSON(data); err != nil {
						return
					}
				default:
					time.Sleep(time.Millisecond * 100)
				}
			}
		}),
	)

	// 自定义升级器示例
	customUpgrader := websocket.HertzUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(ctx *websocket.Context) bool {
			// 允许所有源（生产环境应该限制）
			return true
		},
	}

	customNS := mvc.NewNamespace("/custom",
		mvc.NSWebSocketWithUpgrader("/ws", func(conn *websocket.Conn) {
			log.Printf("Custom WebSocket connection with custom upgrader")

			// 发送欢迎消息
			welcomeMsg := map[string]interface{}{
				"type":    "welcome",
				"message": "Connected with custom upgrader",
				"config": map[string]interface{}{
					"readBuffer":  1024,
					"writeBuffer": 1024,
				},
			}
			conn.WriteJSON(welcomeMsg)

			// 消息处理
			for {
				var message map[string]interface{}
				if err := conn.ReadJSON(&message); err != nil {
					break
				}

				// 回显消息
				response := map[string]interface{}{
					"echo": message,
					"time": time.Now().Unix(),
				}
				conn.WriteJSON(response)
			}
		}, customUpgrader),
	)

	// 注册命名空间
	app.RegisterNamespace(echoNS)
	app.RegisterNamespace(chatNS)
	app.RegisterNamespace(streamNS)
	app.RegisterNamespace(customNS)

	// ============= 使用控制器注册 WebSocket 路由 =============

	// 注册控制器路由
	app.RouterAutoRouter(&EchoController{})
	app.RouterAutoRouter(&ChatController{})
	app.RouterAutoRouter(&StreamController{})

	// 静态文件服务（用于提供 WebSocket 客户端页面）
	app.Static("/static", "./static")

	// 启动服务器
	fmt.Println("WebSocket Example Server Starting...")
	fmt.Println("Available endpoints:")
	fmt.Println("  - ws://localhost:8080/echo/ws (Namespace Echo)")
	fmt.Println("  - ws://localhost:8080/chat/room/general (Namespace Chat)")
	fmt.Println("  - ws://localhost:8080/stream/data (Namespace Stream)")
	fmt.Println("  - ws://localhost:8080/custom/ws (Custom Upgrader)")
	fmt.Println("  - ws://localhost:8080/EchoController/WebSocket (Controller Echo)")
	fmt.Println("  - ws://localhost:8080/ChatController/Room?room=test (Controller Chat)")
	fmt.Println("  - ws://localhost:8080/StreamController/Data (Controller Stream)")

	app.Run(":8080")
}
