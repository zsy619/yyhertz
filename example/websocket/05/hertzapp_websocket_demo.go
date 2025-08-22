// Package main 提供 HertzApp WebSocket 静态方法的测试示例
//
// 本示例演示如何使用 HandlerWs、RouterWs 和 RouterWsWithUpgrader 方法
// 来创建 WebSocket 控制器路由，展示 YYHertz 框架的 WebSocket 功能。
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hertz-contrib/websocket"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// ChatController 聊天控制器示例
type ChatController struct {
	core.BaseController
}

// HandleChat 处理聊天 WebSocket 连接
//
// 这个方法将被 RouterWs 注册为 WebSocket 处理器
func (c *ChatController) HandleChat(conn *define.WsConn) {
	log.Printf("聊天连接建立: %s", conn.RemoteAddr())

	// 发送欢迎消息
	err := conn.WriteMessage(websocket.TextMessage, []byte("欢迎进入聊天室！"))
	if err != nil {
		log.Printf("发送欢迎消息失败: %v", err)
		return
	}

	// 消息处理循环
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取聊天消息失败: %v", err)
			break
		}

		log.Printf("收到聊天消息: %s", message)

		// 回显消息，添加时间戳
		response := fmt.Sprintf("[%s] 聊天回复: %s",
			time.Now().Format("15:04:05"), message)
		err = conn.WriteMessage(messageType, []byte(response))
		if err != nil {
			log.Printf("发送聊天回复失败: %v", err)
			break
		}
	}

	log.Printf("聊天连接断开: %s", conn.RemoteAddr())
}

// HandlePrivate 处理私聊 WebSocket 连接
func (c *ChatController) HandlePrivate(conn *define.WsConn) {
	log.Printf("私聊连接建立: %s", conn.RemoteAddr())

	// 发送私聊欢迎消息
	err := conn.WriteMessage(websocket.TextMessage, []byte("私聊模式已启动！"))
	if err != nil {
		log.Printf("发送私聊欢迎消息失败: %v", err)
		return
	}

	// 私聊消息处理循环
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取私聊消息失败: %v", err)
			break
		}

		log.Printf("收到私聊消息: %s", message)

		// 私聊回复
		response := fmt.Sprintf("私聊回复: %s", message)
		err = conn.WriteMessage(messageType, []byte(response))
		if err != nil {
			log.Printf("发送私聊回复失败: %v", err)
			break
		}
	}

	log.Printf("私聊连接断开: %s", conn.RemoteAddr())
}

// GameController 游戏控制器示例
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

// NotificationController 通知控制器示例
type NotificationController struct {
	core.BaseController
}

// HandleNotifications 处理通知 WebSocket 连接
func (c *NotificationController) HandleNotifications(conn *define.WsConn) error {
	log.Printf("通知连接建立: %s", conn.RemoteAddr())

	// 发送连接确认
	err := conn.WriteMessage(websocket.TextMessage, []byte("通知服务已连接"))
	if err != nil {
		return fmt.Errorf("发送连接确认失败: %v", err)
	}

	// 模拟推送通知
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	counter := 1
	for {
		select {
		case <-ticker.C:
			// 定时推送通知
			notification := fmt.Sprintf("通知 #%d: 这是一条测试通知", counter)
			err := conn.WriteMessage(websocket.TextMessage, []byte(notification))
			if err != nil {
				log.Printf("推送通知失败: %v", err)
				return err
			}
			counter++

		default:
			// 检查是否有客户端消息
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("读取通知消息失败: %v", err)
				return err
			}

			if string(message) == "stop" {
				log.Printf("收到停止指令，断开通知连接")
				return nil
			}

			// 确认收到消息
			response := fmt.Sprintf("收到消息: %s", message)
			err = conn.WriteMessage(messageType, []byte(response))
			if err != nil {
				log.Printf("发送确认消息失败: %v", err)
				return err
			}
		}
	}
}

func main() {
	// 创建 YYHertz 应用
	app := mvc.NewApp()

	fmt.Println("创建 HertzApp WebSocket 静态方法测试服务器...")
	fmt.Println("")

	// ========== 1. 使用 HandlerWs 直接注册处理函数 ==========
	fmt.Println("1. 注册 HandlerWs 路由...")

	// 简单的 Echo WebSocket 服务
	mvc.HandlerWs("/ws/echo", func(conn *define.WsConn) {
		log.Printf("Echo 连接建立: %s", conn.RemoteAddr())

		// Echo 消息处理循环
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Echo 读取消息失败: %v", err)
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
	})

	// 时间服务 WebSocket
	mvc.HandlerWs("/ws/time", func(conn *define.WsConn) {
		log.Printf("时间服务连接建立: %s", conn.RemoteAddr())

		// 每秒发送当前时间
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case now := <-ticker.C:
				timeStr := now.Format("2006-01-02 15:04:05")
				message := fmt.Sprintf("当前时间: %s", timeStr)
				err := conn.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					log.Printf("发送时间失败: %v", err)
					return
				}

			default:
				// 检查连接是否断开
				_, _, err := conn.ReadMessage()
				if err != nil {
					log.Printf("时间服务连接断开: %v", err)
					return
				}
			}
		}
	})

	// ========== 2. 使用 RouterWs 注册控制器方法 ==========
	fmt.Println("2. 注册 RouterWs 控制器路由...")

	chatCtrl := &ChatController{}
	mvc.RouterWs("/ws/chat", chatCtrl, "HandleChat")
	mvc.RouterWs("/ws/private", chatCtrl, "HandlePrivate")

	notificationCtrl := &NotificationController{}
	mvc.RouterWs("/ws/notifications", notificationCtrl, "HandleNotifications")

	// ========== 3. 使用 RouterWsWithUpgrader 带自定义配置 ==========
	fmt.Println("3. 注册 RouterWsWithUpgrader 高性能路由...")

	// 高性能游戏 WebSocket 升级器
	gameUpgrader := websocket.HertzUpgrader{
		ReadBufferSize:   2048,
		WriteBufferSize:  2048,
		HandshakeTimeout: 10 * time.Second,
	}
	gameCtrl := &GameController{}
	mvc.RouterWsWithUpgrader("/ws/game", gameCtrl, "HandleGame", gameUpgrader)

	// 大数据传输 WebSocket 升级器
	dataUpgrader := websocket.HertzUpgrader{
		ReadBufferSize:   8192,
		WriteBufferSize:  8192,
		HandshakeTimeout: 15 * time.Second,
	}
	mvc.RouterWsWithUpgrader("/ws/data", chatCtrl, "HandleChat", dataUpgrader)

	// ========== 4. 链式调用示例 ==========
	fmt.Println("4. 使用链式调用注册多个路由...")

	app.HandlerWs("/ws/test1", func(conn *define.WsConn) {
		conn.WriteMessage(websocket.TextMessage, []byte("Test1 连接成功"))
		conn.ReadMessage() // 等待客户端消息
	}).HandlerWs("/ws/test2", func(conn *define.WsConn) {
		conn.WriteMessage(websocket.TextMessage, []byte("Test2 连接成功"))
		conn.ReadMessage() // 等待客户端消息
	}).RouterWs("/ws/test3", chatCtrl, "HandleChat")

	fmt.Println("")
	fmt.Println("HertzApp WebSocket 静态方法测试服务器启动中...")
	fmt.Println("可用的 WebSocket 端点:")
	fmt.Println("")
	fmt.Println("  【HandlerWs 直接处理函数】")
	fmt.Println("  - ws://localhost:8888/ws/echo (Echo 服务)")
	fmt.Println("  - ws://localhost:8888/ws/time (时间服务)")
	fmt.Println("")
	fmt.Println("  【RouterWs 控制器方法】")
	fmt.Println("  - ws://localhost:8888/ws/chat (聊天室)")
	fmt.Println("  - ws://localhost:8888/ws/private (私聊)")
	fmt.Println("  - ws://localhost:8888/ws/notifications (通知推送)")
	fmt.Println("")
	fmt.Println("  【RouterWsWithUpgrader 自定义升级器】")
	fmt.Println("  - ws://localhost:8888/ws/game (游戏，高性能配置)")
	fmt.Println("  - ws://localhost:8888/ws/data (数据传输，大缓冲区)")
	fmt.Println("")
	fmt.Println("  【链式调用测试】")
	fmt.Println("  - ws://localhost:8888/ws/test1 (测试1)")
	fmt.Println("  - ws://localhost:8888/ws/test2 (测试2)")
	fmt.Println("  - ws://localhost:8888/ws/test3 (测试3)")
	fmt.Println("")
	fmt.Println("测试方法:")
	fmt.Println("1. 使用 WebSocket 客户端连接上述地址")
	fmt.Println("2. Echo: 发送任意消息，将原样返回")
	fmt.Println("3. Time: 连接后每秒接收当前时间")
	fmt.Println("4. Chat: 发送消息，接收带时间戳的回复")
	fmt.Println("5. Game: 发送 'start' 开始游戏，'quit' 退出")
	fmt.Println("6. Notifications: 连接后每5秒接收通知，发送 'stop' 断开")
	fmt.Println("")

	// 启动应用
	app.Run()
}
