package define

import (
	"github.com/hertz-contrib/websocket"
)

// WsConn WebSocket 连接类型
type WsConn = websocket.Conn

// WsClientUpgrader is a helper for upgrading hertz http response to websocket conn.
type WsClientUpgrader = websocket.ClientUpgrader

// WsHandlerFunc 定义 WebSocket 处理函数类型
//
// 该函数类型用于处理 WebSocket 连接，在连接建立后被调用。
// 函数应该包含消息处理循环，负责读取和写入 WebSocket 消息。
type WsHandlerFunc func(*WsConn)

// WsHertzUpgrader 定义 WebSocket 升级器类型
type WsHertzUpgrader = websocket.HertzUpgrader
