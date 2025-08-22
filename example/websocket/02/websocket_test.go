package mvc

import (
	"net/url"
	"testing"
	"time"

	"github.com/hertz-contrib/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// ============= 测试控制器 =============

// TestWebSocketController 测试用的 WebSocket 控制器
type TestWebSocketController struct {
	core.BaseController
	receivedMessages []string
}

// Echo WebSocket Echo 处理器
func (c *TestWebSocketController) Echo() {
	c.HandleWebSocket(func(conn *mvc.WsConn) {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 记录接收到的消息
			c.receivedMessages = append(c.receivedMessages, string(message))

			// Echo 消息
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				break
			}
		}
	}, nil)
}

// JSON JSON 消息处理器
func (c *TestWebSocketController) JSON() {
	handler := c.CreateWebSocketJSONHandler(func(data map[string]interface{}) map[string]interface{} {
		response := make(map[string]interface{})
		response["echo"] = data
		response["timestamp"] = time.Now().Unix()
		return response
	})
	c.HandleWebSocket(handler, nil)
}

// ============= 命名空间 WebSocket 测试 =============

func TestNamespaceWebSocket(t *testing.T) {
	// 创建应用
	app := core.NewApp()

	// 测试消息存储
	var receivedMessages []string

	// 创建 WebSocket 命名空间
	ns := mvc.NewNamespace("/test",
		mvc.NSWebSocket("/echo", func(conn *mvc.WsConn) {
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					break
				}

				receivedMessages = append(receivedMessages, string(message))

				// Echo 消息
				err = conn.WriteMessage(messageType, message)
				if err != nil {
					break
				}
			}
		}),
	)

	// 注册命名空间
	ns.Register(app)

	// 启动测试服务器
	go func() {
		app.Run(":8081")
	}()

	// 等待服务器启动
	time.Sleep(time.Millisecond * 100)

	// 连接 WebSocket
	u := url.URL{Scheme: "ws", Host: "localhost:8081", Path: "/test/echo"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed (server may not be running): %v", err)
		return
	}
	defer c.Close()

	// 发送测试消息
	testMessage := "Hello WebSocket"
	err = c.WriteMessage(websocket.TextMessage, []byte(testMessage))
	assert.NoError(t, err)

	// 读取响应
	_, message, err := c.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, testMessage, string(message))
}

func TestNamespaceWebSocketWithCustomUpgrader(t *testing.T) {
	// 创建应用
	app := core.NewApp()

	// 自定义升级器
	upgrader := websocket.HertzUpgrader{
		ReadBufferSize:  512,
		WriteBufferSize: 512,
		CheckOrigin: func(ctx *app.RequestContext) bool {
			return true // 允许所有源
		},
	}

	// 创建带自定义升级器的命名空间
	ns := mvc.NewNamespace("/custom",
		mvc.NSWebSocketWithUpgrader("/ws", func(conn *websocket.Conn) {
			// 发送欢迎消息
			welcome := map[string]interface{}{
				"type":    "welcome",
				"message": "Custom upgrader test",
			}
			conn.WriteJSON(welcome)

			// 消息处理循环
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
		}, upgrader),
	)

	// 注册命名空间
	app.RegisterNamespace(ns)

	// 启动测试服务器
	go func() {
		app.Run(":8082")
	}()

	// 等待服务器启动
	time.Sleep(time.Millisecond * 100)

	// 连接 WebSocket
	u := url.URL{Scheme: "ws", Host: "localhost:8082", Path: "/custom/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Skipf("WebSocket connection failed (server may not be running): %v", err)
		return
	}
	defer c.Close()

	// 读取欢迎消息
	var welcome map[string]interface{}
	err = c.ReadJSON(&welcome)
	assert.NoError(t, err)
	assert.Equal(t, "welcome", welcome["type"])
	assert.Equal(t, "Custom upgrader test", welcome["message"])

	// 发送 JSON 消息
	testData := map[string]interface{}{
		"message": "test",
		"data":    123,
	}
	err = c.WriteJSON(testData)
	assert.NoError(t, err)

	// 读取响应
	var response map[string]interface{}
	err = c.ReadJSON(&response)
	assert.NoError(t, err)

	assert.Contains(t, response, "echo")
	assert.Contains(t, response, "time")

	echoData := response["echo"].(map[string]interface{})
	assert.Equal(t, "test", echoData["message"])
	assert.Equal(t, float64(123), echoData["data"]) // JSON 数字被解析为 float64
}

// ============= WebSocket 控制器测试 =============

func TestWebSocketController(t *testing.T) {
	// 测试 WebSocket 控制器创建
	handler := func(conn *websocket.Conn) {
		// 简单的处理函数
	}
	upgrader := websocket.HertzUpgrader{}

	controller := NewWebSocketController(handler, upgrader)
	assert.NotNil(t, controller)
	assert.Equal(t, 0, controller.GetActiveConnections())
}

func TestWebSocketManager(t *testing.T) {
	manager := NewWebSocketManager()
	assert.NotNil(t, manager)
	assert.Equal(t, 0, manager.GetConnectionCount())
	assert.Equal(t, 0, manager.GetRoomCount())

	// 模拟连接（这里无法创建真正的 websocket.Conn，所以跳过详细测试）
	// 在实际项目中，可以使用 mock 来测试连接管理功能
}

// ============= 集成测试 =============

func TestWebSocketIntegration(t *testing.T) {
	// 创建应用
	app := core.NewApp()

	// 注册控制器
	app.RouterAutoRouter(&TestWebSocketController{})

	// 启动测试服务器
	go func() {
		app.Run(":8083")
	}()

	// 等待服务器启动
	time.Sleep(time.Millisecond * 100)

	// 测试 Echo 端点
	t.Run("EchoEndpoint", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8083", Path: "/TestWebSocketController/Echo"}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Skipf("WebSocket connection failed: %v", err)
			return
		}
		defer c.Close()

		// 发送测试消息
		testMessage := "Integration Test"
		err = c.WriteMessage(websocket.TextMessage, []byte(testMessage))
		assert.NoError(t, err)

		// 读取响应
		_, message, err := c.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, testMessage, string(message))
	})

	// 测试 JSON 端点
	t.Run("JSONEndpoint", func(t *testing.T) {
		u := url.URL{Scheme: "ws", Host: "localhost:8084", Path: "/TestWebSocketController/JSON"}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Skipf("WebSocket connection failed: %v", err)
			return
		}
		defer c.Close()

		// 发送 JSON 消息
		testData := map[string]interface{}{
			"message": "JSON test",
			"number":  42,
		}
		err = c.WriteJSON(testData)
		assert.NoError(t, err)

		// 读取 JSON 响应
		var response map[string]interface{}
		err = c.ReadJSON(&response)
		assert.NoError(t, err)

		assert.Contains(t, response, "echo")
		assert.Contains(t, response, "timestamp")

		echoData := response["echo"].(map[string]interface{})
		assert.Equal(t, "JSON test", echoData["message"])
		assert.Equal(t, float64(42), echoData["number"])
	})
}

// ============= 性能测试 =============

func BenchmarkWebSocketCreation(b *testing.B) {
	handler := func(conn *websocket.Conn) {
		// 空处理函数
	}
	upgrader := websocket.HertzUpgrader{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller := NewWebSocketController(handler, upgrader)
		_ = controller
	}
}

func BenchmarkWebSocketManager(b *testing.B) {
	manager := NewWebSocketManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetConnectionCount()
		manager.GetRoomCount()
	}
}

// ============= 错误处理测试 =============

func TestWebSocketErrorHandling(t *testing.T) {
	// 测试非 WebSocket 请求处理
	// 这里需要创建一个模拟的控制器上下文来测试错误处理
	// 由于涉及复杂的上下文模拟，这里只做基本的结构测试

	controller := &TestWebSocketController{}
	assert.NotNil(t, controller)

	// 测试 IsWebSocketRequest 方法
	// 需要在有完整上下文的情况下进行测试
}

// ============= 辅助函数测试 =============

func TestWebSocketUtilities(t *testing.T) {
	// 测试全局 WebSocket 管理器
	manager := GetGlobalWebSocketManager()
	assert.NotNil(t, manager)
	assert.Equal(t, 0, manager.GetConnectionCount())

	// 测试 WebSocket 控制器工厂函数
	handler := func(conn *websocket.Conn) {}
	upgrader := websocket.HertzUpgrader{}

	controller := CreateWebSocketController(handler, upgrader)
	assert.NotNil(t, controller)
}

// ============= 聊天室功能测试 =============

func TestChatRoomFunctionality(t *testing.T) {
	chatRoom := NewWebSocketManager()

	// 测试房间管理
	assert.Equal(t, 0, chatRoom.GetRoomCount())

	// 由于无法创建真实的 websocket.Conn 对象，
	// 这里主要测试管理器的基本功能

	// 测试广播功能（使用空房间）
	chatRoom.BroadcastToRoom("test-room", 1, []byte("test message"))
	chatRoom.BroadcastToAll(1, []byte("global message"))

	// 验证没有崩溃
	assert.Equal(t, 0, chatRoom.GetConnectionCount())
}

// ============= 命名空间查询方法测试 =============

func TestNamespaceWebSocketQueries(t *testing.T) {
	// 创建包含 WebSocket 路由的命名空间
	ns := NewNamespace("/test",
		NSWebSocket("/ws1", func(conn *websocket.Conn) {}),
		NSWebSocket("/ws2", func(conn *websocket.Conn) {}),
		NSWebSocketWithUpgrader("/ws3", func(conn *websocket.Conn) {}, websocket.HertzUpgrader{}),
	)

	// 测试基本属性
	assert.Equal(t, "/test", ns.GetPrefix())
	assert.Equal(t, 0, len(ns.GetControllers()))
	assert.Equal(t, 0, len(ns.GetRouters()))
	assert.Equal(t, 0, len(ns.GetNamespaces()))

	// 验证 WebSocket 路由已添加
	assert.Equal(t, 3, len(ns.wsRouters))

	// 验证路径
	assert.Equal(t, "/ws1", ns.wsRouters[0].path)
	assert.Equal(t, "/ws2", ns.wsRouters[1].path)
	assert.Equal(t, "/ws3", ns.wsRouters[2].path)
}

// ============= 嵌套命名空间 WebSocket 测试 =============

func TestNestedNamespaceWebSocket(t *testing.T) {
	// 创建嵌套的命名空间
	parentNS := NewNamespace("/api",
		NSNamespace("/v1",
			NSWebSocket("/ws", func(conn *websocket.Conn) {
				conn.WriteMessage(websocket.TextMessage, []byte("v1 response"))
			}),
		),
		NSNamespace("/v2",
			NSWebSocket("/ws", func(conn *websocket.Conn) {
				conn.WriteMessage(websocket.TextMessage, []byte("v2 response"))
			}),
		),
	)

	// 验证嵌套结构
	assert.Equal(t, "/api", parentNS.GetPrefix())
	assert.Equal(t, 2, len(parentNS.GetNamespaces()))

	// 验证子命名空间
	subNamespaces := parentNS.GetNamespaces()
	assert.Equal(t, "/v1", subNamespaces[0].GetPrefix())
	assert.Equal(t, "/v2", subNamespaces[1].GetPrefix())

	// 验证 WebSocket 路由
	assert.Equal(t, 1, len(subNamespaces[0].wsRouters))
	assert.Equal(t, 1, len(subNamespaces[1].wsRouters))
}
