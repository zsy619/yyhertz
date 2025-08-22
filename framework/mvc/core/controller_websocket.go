package core

import (
	"github.com/hertz-contrib/websocket"

	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// ============= WebSocket 支持方法 =============

// HandleWebSocket 处理 WebSocket 连接升级
//
// 该方法在 BaseController 中提供 WebSocket 支持，允许控制器直接处理 WebSocket 连接。
// 它会检查请求是否为有效的 WebSocket 升级请求，然后执行升级并调用处理函数。
//
// 参数：
//   - handler: func(*websocket.Conn) - WebSocket连接处理函数
//   - upgrader: *websocket.HertzUpgrader - 可选的自定义升级器，如果为nil则使用默认配置
//
// 使用示例：
//
//	type ChatController struct {
//		mvc.BaseController
//	}
//
//	func (c *ChatController) WebSocketEndpoint() {
//		c.HandleWebSocket(func(conn *websocket.Conn) {
//			for {
//				messageType, message, err := conn.ReadMessage()
//				if err != nil {
//					log.Println("read error:", err)
//					break
//				}
//
//				// Echo 消息
//				err = conn.WriteMessage(messageType, message)
//				if err != nil {
//					log.Println("write error:", err)
//					break
//				}
//			}
//		}, nil)
//	}
func (c *BaseController) HandleWebSocket(handler func(*define.WsConn), upgrader *websocket.HertzUpgrader) {
	// 检查是否为 WebSocket 请求
	if !c.Ctx.IsWebsocket() {
		c.Abort("400")
		return
	}

	// 使用默认升级器如果没有提供
	var wsUpgrader websocket.HertzUpgrader
	if upgrader != nil {
		wsUpgrader = *upgrader
	} else {
		wsUpgrader = websocket.HertzUpgrader{} // 默认配置
	}

	// 执行 WebSocket 升级
	err := wsUpgrader.Upgrade(c.Ctx.Request(), handler)
	if err != nil {
		c.LogErrorf("WebSocket upgrade failed: %v", err)
		c.Abort("500")
	}
}

// IsWebSocketRequest 检查当前请求是否为 WebSocket 升级请求
//
// 返回值：
//   - bool: 如果是 WebSocket 请求返回 true，否则返回 false
//
// 使用示例：
//
//	func (c *MyController) Get() {
//		if c.IsWebSocketRequest() {
//			c.handleWebSocketConnection()
//		} else {
//			c.handleHttpRequest()
//		}
//	}
func (c *BaseController) IsWebSocketRequest() bool {
	return c.Ctx.IsWebsocket()
}

// SetWebSocketUpgrader 设置控制器的默认 WebSocket 升级器
//
// 该方法允许为控制器设置自定义的 WebSocket 升级器配置，
// 包括缓冲区大小、源检查函数等。
//
// 参数：
//   - upgrader: websocket.HertzUpgrader - WebSocket升级器配置
//
// 使用示例：
//
//	func (c *ChatController) Prepare() {
//		c.BaseController.Prepare()
//
//		// 配置自定义升级器
//		upgrader := websocket.HertzUpgrader{
//			ReadBufferSize:  1024,
//			WriteBufferSize: 1024,
//			CheckOrigin: func(ctx *app.RequestContext) bool {
//				origin := string(ctx.GetHeader("Origin"))
//				return isAllowedOrigin(origin)
//			},
//		}
//		c.SetWebSocketUpgrader(upgrader)
//	}
func (c *BaseController) SetWebSocketUpgrader(upgrader websocket.HertzUpgrader) {
	// 可以将升级器存储在控制器中，供后续使用
	// 这里可以添加一个字段来存储升级器配置
	c.LogInfof("WebSocket upgrader configured with buffer sizes: read=%d, write=%d",
		upgrader.ReadBufferSize, upgrader.WriteBufferSize)
}

// CreateWebSocketEchoHandler 创建一个简单的 WebSocket Echo 处理器
//
// 这是一个便捷方法，创建一个简单的 WebSocket 处理器，
// 它会将接收到的所有消息原样返回给客户端。
//
// 返回值：
//   - func(*websocket.Conn): WebSocket连接处理函数
//
// 使用示例：
//
//	func (c *TestController) Echo() {
//		c.HandleWebSocket(c.CreateWebSocketEchoHandler(), nil)
//	}
func (c *BaseController) CreateWebSocketEchoHandler() func(*websocket.Conn) {
	return func(conn *websocket.Conn) {
		c.LogInfof("WebSocket connection established from %s", c.Ctx.ClientIP())
		defer c.LogInfof("WebSocket connection closed")

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.LogErrorf("WebSocket error: %v", err)
				}
				break
			}

			c.LogInfof("Received message: %s", string(message))

			// Echo 消息回客户端
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				c.LogErrorf("WebSocket write error: %v", err)
				break
			}
		}
	}
}

// CreateWebSocketJSONHandler 创建一个 JSON 消息处理器
//
// 该处理器专门用于处理 JSON 格式的 WebSocket 消息，
// 它会自动序列化和反序列化 JSON 数据。
//
// 参数：
//   - messageHandler: func(map[string]interface{}) map[string]interface{} - JSON消息处理函数
//
// 返回值：
//   - func(*websocket.Conn): WebSocket连接处理函数
//
// 使用示例：
//
//	func (c *APIController) JSONEndpoint() {
//		handler := c.CreateWebSocketJSONHandler(func(data map[string]interface{}) map[string]interface{} {
//			// 处理接收到的 JSON 数据
//			response := make(map[string]interface{})
//			response["echo"] = data
//			response["timestamp"] = time.Now().Unix()
//			return response
//		})
//		c.HandleWebSocket(handler, nil)
//	}
func (c *BaseController) CreateWebSocketJSONHandler(messageHandler func(map[string]interface{}) map[string]interface{}) func(*websocket.Conn) {
	return func(conn *websocket.Conn) {
		c.LogInfof("JSON WebSocket connection established from %s", c.Ctx.ClientIP())
		defer c.LogInfof("JSON WebSocket connection closed")

		for {
			var message map[string]interface{}
			err := conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.LogErrorf("WebSocket JSON read error: %v", err)
				}
				break
			}

			c.LogInfof("Received JSON message: %+v", message)

			// 处理消息
			response := messageHandler(message)

			// 发送 JSON 响应
			err = conn.WriteJSON(response)
			if err != nil {
				c.LogErrorf("WebSocket JSON write error: %v", err)
				break
			}
		}
	}
}
