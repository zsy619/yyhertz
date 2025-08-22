// Package main 提供 NSRouterWs 功能的测试示例
//
// 本示例演示如何使用 NSRouterWs 方法将控制器方法直接映射为 WebSocket 处理器
package main

import (
	"fmt"
	"log"

	"github.com/hertz-contrib/websocket"
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// ChatController WebSocket 聊天控制器示例
type ChatController struct {
	core.BaseController
}

// HandleChat 处理聊天室 WebSocket 连接
//
// 这个方法将被 NSRouterWs 映射为 WebSocket 处理器
func (c *ChatController) HandleChat(conn *define.WsConn) {
	log.Printf("聊天室连接建立: %s", conn.RemoteAddr())

	// 发送欢迎消息
	err := conn.WriteMessage(websocket.TextMessage, []byte("欢迎来到聊天室！"))
	if err != nil {
		log.Printf("发送欢迎消息失败: %v", err)
		return
	}

	// 消息处理循环
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		log.Printf("收到聊天消息: %s", message)

		// 回显消息，添加前缀
		response := fmt.Sprintf("聊天室回复: %s", message)
		err = conn.WriteMessage(messageType, []byte(response))
		if err != nil {
			log.Printf("发送回复失败: %v", err)
			break
		}
	}

	log.Printf("聊天室连接断开: %s", conn.RemoteAddr())
}

// HandlePrivate 处理私聊 WebSocket 连接
func (c *ChatController) HandlePrivate(conn *define.WsConn) {
	log.Printf("私聊连接建立: %s", conn.RemoteAddr())

	// 发送欢迎消息
	err := conn.WriteMessage(websocket.TextMessage, []byte("欢迎使用私聊功能！"))
	if err != nil {
		log.Printf("发送欢迎消息失败: %v", err)
		return
	}

	// 消息处理循环
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取私聊消息失败: %v", err)
			break
		}

		log.Printf("收到私聊消息: %s", message)

		// 回显消息，添加私聊前缀
		response := fmt.Sprintf("私聊回复: %s", message)
		err = conn.WriteMessage(messageType, []byte(response))
		if err != nil {
			log.Printf("发送私聊回复失败: %v", err)
			break
		}
	}

	log.Printf("私聊连接断开: %s", conn.RemoteAddr())
}

// EchoController WebSocket Echo 控制器示例
type EchoController struct {
	core.BaseController
}

// HandleEcho 处理 Echo WebSocket 连接
func (c *EchoController) HandleEcho(conn *define.WsConn) {
	log.Printf("Echo 连接建立: %s", conn.RemoteAddr())

	// 消息处理循环
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取 Echo 消息失败: %v", err)
			break
		}

		log.Printf("Echo 收到消息: %s", message)

		// 原样回显消息
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("Echo 回复失败: %v", err)
			break
		}
	}

	log.Printf("Echo 连接断开: %s", conn.RemoteAddr())
}

// GameController WebSocket 游戏控制器示例
type GameController struct {
	core.BaseController
}

// HandleGame 处理游戏 WebSocket 连接（带自定义升级器）
func (c *GameController) HandleGame(conn *define.WsConn) {
	log.Printf("游戏连接建立: %s", conn.RemoteAddr())

	// 发送游戏开始消息
	err := conn.WriteMessage(websocket.TextMessage, []byte("游戏开始！发送 'start' 开始游戏"))
	if err != nil {
		log.Printf("发送游戏开始消息失败: %v", err)
		return
	}

	gameStarted := false

	// 游戏消息处理循环
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取游戏消息失败: %v", err)
			break
		}

		log.Printf("游戏收到消息: %s", message)

		messageStr := string(message)
		var response string

		if !gameStarted && messageStr == "start" {
			gameStarted = true
			response = "游戏已开始！发送任意消息继续游戏，发送 'quit' 退出游戏"
		} else if gameStarted {
			if messageStr == "quit" {
				response = "游戏结束！感谢游玩"
				conn.WriteMessage(messageType, []byte(response))
				break
			} else {
				response = fmt.Sprintf("游戏响应: 你发送了 '%s'", messageStr)
			}
		} else {
			response = "请先发送 'start' 开始游戏"
		}

		err = conn.WriteMessage(messageType, []byte(response))
		if err != nil {
			log.Printf("发送游戏回复失败: %v", err)
			break
		}
	}

	log.Printf("游戏连接断开: %s", conn.RemoteAddr())
}

func main() {
	// 创建 YYHertz 应用
	app := mvc.NewApp()

	fmt.Println("创建 NSRouterWs WebSocket 控制器测试服务器...")

	// 测试 1: 基本的 WebSocket 控制器路由
	chatNS := mvc.NewNamespace("/chat",
		mvc.NSRouterWs("/room", &ChatController{}, "HandleChat"),
		mvc.NSRouterWs("/private", &ChatController{}, "HandlePrivate"),
	)

	// 测试 2: Echo 控制器路由
	echoNS := mvc.NewNamespace("/api",
		mvc.NSRouterWs("/echo", &EchoController{}, "HandleEcho"),
	)

	// 测试 3: 带自定义升级器的游戏控制器路由
	gameUpgrader := websocket.HertzUpgrader{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	}
	gameNS := mvc.NewNamespace("/game",
		mvc.NSRouterWsWithUpgrader("/play", &GameController{}, "HandleGame", gameUpgrader),
	)

	// 注册命名空间到应用
	fmt.Println("注册 WebSocket 控制器命名空间...")
	chatNS.Register(app)
	echoNS.Register(app)
	gameNS.Register(app)

	// 启动服务器
	fmt.Println("NSRouterWs WebSocket 控制器测试服务器启动中...")
	fmt.Println("可用的 WebSocket 控制器端点:")
	fmt.Println("  - ws://localhost:8888/chat/room (聊天室)")
	fmt.Println("  - ws://localhost:8888/chat/private (私聊)")
	fmt.Println("  - ws://localhost:8888/api/echo (Echo 服务)")
	fmt.Println("  - ws://localhost:8888/game/play (游戏，自定义升级器)")
	fmt.Println("")
	fmt.Println("测试方法:")
	fmt.Println("1. 使用 WebSocket 客户端连接上述地址")
	fmt.Println("2. 聊天室和私聊：发送任意文本消息")
	fmt.Println("3. Echo 服务：发送消息将原样返回")
	fmt.Println("4. 游戏：发送 'start' 开始，发送 'quit' 退出")

	// 启动应用
	app.Run()
}