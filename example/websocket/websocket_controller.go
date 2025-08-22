// Package mvc 提供 WebSocket 控制器实现
//
// 本文件实现了 WebSocket 连接处理的控制器，用于在命名空间系统中支持 WebSocket 路由。
// 该控制器负责处理 HTTP 到 WebSocket 的协议升级，并管理 WebSocket 连接的生命周期。
package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hertz-contrib/websocket"

	"github.com/zsy619/yyhertz/framework/mvc"
	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// ============= WebSocket 控制器 =============

// WebSocketController WebSocket 连接处理控制器
//
// 该控制器实现了 core.IController 接口，专门用于处理 WebSocket 连接。
// 它封装了 WebSocket 升级过程和连接管理逻辑。
type WebSocketController struct {
	core.BaseController

	// handler WebSocket连接处理函数
	handler mvc.WsHandlerFunc

	// upgrader WebSocket升级器
	upgrader websocket.HertzUpgrader

	// connections 活跃连接管理
	connections map[string]*websocket.Conn
	connMutex   sync.RWMutex
}

// NewWebSocketController 创建 WebSocket 控制器
//
// 参数：
//   - handler: WebSocketHandlerFunc - WebSocket连接处理函数
//   - upgrader: websocket.HertzUpgrader - WebSocket升级器配置
//
// 返回值：
//   - *WebSocketController: WebSocket控制器实例
func NewWebSocketController(handler mvc.WsHandlerFunc, upgrader websocket.HertzUpgrader) *WebSocketController {
	return &WebSocketController{
		handler:     handler,
		upgrader:    upgrader,
		connections: make(map[string]*mvc.WsConn),
	}
}

// HandleWebSocket 处理 WebSocket 连接请求
//
// 该方法是 WebSocket 控制器的主要入口点，负责：
// 1. 验证请求是否为有效的 WebSocket 升级请求
// 2. 执行协议升级
// 3. 调用用户定义的处理函数
// 4. 处理连接错误和清理
func (wsc *WebSocketController) HandleWebSocket() {
	// 检查是否为 WebSocket 请求
	if !wsc.Ctx.IsWebsocket() {
		wsc.Ctx.AbortWithStatus(400)
		return
	}

	// 执行 WebSocket 升级
	err := wsc.upgrader.Upgrade(wsc.Ctx.Request(), func(conn *websocket.Conn) {
		// 生成连接ID
		connID := wsc.generateConnectionID()

		// 注册连接
		wsc.addConnection(connID, conn)
		defer wsc.removeConnection(connID)

		// 设置连接参数（如果需要路径参数）
		wsc.setupConnectionParams(conn)

		// 调用用户定义的处理函数
		wsc.handler(conn)
	})

	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		wsc.Ctx.AbortWithStatus(500)
	}
}

// generateConnectionID 生成连接ID
func (wsc *WebSocketController) generateConnectionID() string {
	// 使用远程地址和时间戳生成简单的连接ID
	// 在生产环境中可能需要更复杂的ID生成策略
	return fmt.Sprintf("%s_%d", wsc.Ctx.ClientIP(), time.Now().UnixNano())
}

// addConnection 添加连接到管理列表
func (wsc *WebSocketController) addConnection(id string, conn *websocket.Conn) {
	wsc.connMutex.Lock()
	defer wsc.connMutex.Unlock()
	wsc.connections[id] = conn
}

// removeConnection 从管理列表中移除连接
func (wsc *WebSocketController) removeConnection(id string) {
	wsc.connMutex.Lock()
	defer wsc.connMutex.Unlock()
	delete(wsc.connections, id)
}

// setupConnectionParams 设置连接参数
//
// 该方法将路由参数等信息传递给 WebSocket 连接，
// 使得在处理函数中可以访问路径参数。
func (wsc *WebSocketController) setupConnectionParams(conn *websocket.Conn) {
	// TODO: 实现参数传递机制
	// 可以通过连接的 locals 或自定义字段传递参数
}

// GetActiveConnections 获取活跃连接数
func (wsc *WebSocketController) GetActiveConnections() int {
	wsc.connMutex.RLock()
	defer wsc.connMutex.RUnlock()
	return len(wsc.connections)
}

// CloseAllConnections 关闭所有连接
func (wsc *WebSocketController) CloseAllConnections() {
	wsc.connMutex.Lock()
	defer wsc.connMutex.Unlock()

	for id, conn := range wsc.connections {
		conn.Close()
		delete(wsc.connections, id)
	}
}

// ============= 控制器接口实现 =============

// Init 初始化控制器
func (wsc *WebSocketController) Init(ctx *mvcContext.Context, controllerName, actionName string, app any) {
	wsc.BaseController.Init(ctx, controllerName, actionName, app)
}

// Prepare 预处理请求
func (wsc *WebSocketController) Prepare() {
	// WebSocket 控制器通常不需要特殊的预处理
	wsc.BaseController.Prepare()
}

// Finish 完成请求处理
func (wsc *WebSocketController) Finish() {
	// 清理资源
	wsc.BaseController.Finish()
}

// Get 处理 GET 请求（WebSocket 升级通常使用 GET）
func (wsc *WebSocketController) Get() {
	wsc.HandleWebSocket()
}

// ============= 辅助函数 =============

// CreateWebSocketController 创建用于命名空间的 WebSocket 控制器工厂函数
//
// 该函数为命名空间系统提供一个便捷的控制器创建方法。
//
// 参数：
//   - handler: WebSocketHandlerFunc - WebSocket连接处理函数
//   - upgrader: websocket.HertzUpgrader - 升级器配置
//
// 返回值：
//   - core.IController: 实现了控制器接口的 WebSocket 控制器
func CreateWebSocketController(handler mvc.WsHandlerFunc, upgrader websocket.HertzUpgrader) core.IController {
	return NewWebSocketController(handler, upgrader)
}

// ============= 连接管理器 =============

// WebSocketManager WebSocket 连接管理器
//
// 提供全局的 WebSocket 连接管理功能，支持广播消息等高级功能。
type WebSocketManager struct {
	connections map[string]*mvc.WsConn
	rooms       map[string]map[string]*mvc.WsConn // room -> connections
	mutex       sync.RWMutex
}

// NewWebSocketManager 创建 WebSocket 管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[string]*mvc.WsConn),
		rooms:       make(map[string]map[string]*mvc.WsConn),
	}
}

// AddConnection 添加连接
func (wsm *WebSocketManager) AddConnection(id string, conn *mvc.WsConn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	wsm.connections[id] = conn
}

// RemoveConnection 移除连接
func (wsm *WebSocketManager) RemoveConnection(id string) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	// 从所有房间中移除
	for roomID, roomConns := range wsm.rooms {
		delete(roomConns, id)
		if len(roomConns) == 0 {
			delete(wsm.rooms, roomID)
		}
	}

	// 从全局连接中移除
	delete(wsm.connections, id)
}

// JoinRoom 加入房间
func (wsm *WebSocketManager) JoinRoom(connID, roomID string) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	conn, exists := wsm.connections[connID]
	if !exists {
		return
	}

	if wsm.rooms[roomID] == nil {
		wsm.rooms[roomID] = make(map[string]*websocket.Conn)
	}

	wsm.rooms[roomID][connID] = conn
}

// LeaveRoom 离开房间
func (wsm *WebSocketManager) LeaveRoom(connID, roomID string) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	if room, exists := wsm.rooms[roomID]; exists {
		delete(room, connID)
		if len(room) == 0 {
			delete(wsm.rooms, roomID)
		}
	}
}

// BroadcastToRoom 向房间广播消息
func (wsm *WebSocketManager) BroadcastToRoom(roomID string, messageType int, message []byte) {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()

	if room, exists := wsm.rooms[roomID]; exists {
		for _, conn := range room {
			conn.WriteMessage(messageType, message)
		}
	}
}

// BroadcastToAll 向所有连接广播消息
func (wsm *WebSocketManager) BroadcastToAll(messageType int, message []byte) {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()

	for _, conn := range wsm.connections {
		conn.WriteMessage(messageType, message)
	}
}

// GetConnectionCount 获取连接总数
func (wsm *WebSocketManager) GetConnectionCount() int {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()
	return len(wsm.connections)
}

// GetRoomCount 获取房间总数
func (wsm *WebSocketManager) GetRoomCount() int {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()
	return len(wsm.rooms)
}

// 全局 WebSocket 管理器实例
var globalWSManager = NewWebSocketManager()

// GetGlobalWebSocketManager 获取全局 WebSocket 管理器
func GetGlobalWebSocketManager() *WebSocketManager {
	return globalWSManager
}
