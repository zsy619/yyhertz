package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func testWebSocketEndpoint(endpoint string, description string) {
	fmt.Printf("\n=== 测试 %s ===\n", description)
	fmt.Printf("连接到: %s\n", endpoint)

	u, err := url.Parse(endpoint)
	if err != nil {
		fmt.Printf("❌ 解析 URL 失败: %v\n", err)
		return
	}

	// 连接 WebSocket
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer c.Close()

	fmt.Printf("✅ 连接成功！\n")

	// 发送测试消息
	testMessages := []string{"test", "ping", "heartbeat", "hello world"}

	for _, msg := range testMessages {
		fmt.Printf("📤 发送消息: %s\n", msg)

		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			fmt.Printf("❌ 发送失败: %v\n", err)
			continue
		}

		// 设置读取超时
		c.SetReadDeadline(time.Now().Add(5 * time.Second))

		// 读取响应
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Printf("❌ 读取响应失败: %v\n", err)
			continue
		}

		fmt.Printf("📨 收到响应: %s\n", string(message))
	}

	fmt.Printf("✅ %s 测试完成\n", description)
}

func main_client() {
	fmt.Println("🚀 YYHertz WebSocket 客户端测试工具")
	fmt.Println("正在测试嵌套命名空间 WebSocket 路由...")

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试嵌套命名空间路由 - 这是主要测试目标
	testWebSocketEndpoint("ws://localhost:8888/admin/online/heartbeat", "嵌套命名空间 (/admin/online/heartbeat)")

	// 对比测试其他路由
	testWebSocketEndpoint("ws://localhost:8888/ws/simple", "简单路由 (/ws/simple)")
	testWebSocketEndpoint("ws://localhost:8888/api/v1/websocket/test", "单级命名空间 (/api/v1/websocket/test)")
	testWebSocketEndpoint("ws://localhost:8888/chat/room/general", "聊天室嵌套 (/chat/room/general)")

	fmt.Println("\n🎉 所有测试完成！")
	fmt.Println("如果所有连接都成功，说明嵌套命名空间 WebSocket 路由问题已经修复。")
}
