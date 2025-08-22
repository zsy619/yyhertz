// Package mvc 提供 WebSocket 工具和错误处理功能
//
// 本文件包含 WebSocket 连接管理、错误处理、连接池等高级功能，
// 用于提升 WebSocket 应用的稳定性和性能。
package mvc

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hertz-contrib/websocket"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

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
func CreateWebSocketController(handler WsHandlerFunc, upgrader websocket.HertzUpgrader) core.IController {
	return core.NewWebSocketController(handler, upgrader)
}

// ============= 连接管理器 =============

// WebSocketManager WebSocket 连接管理器
//
// 提供全局的 WebSocket 连接管理功能，支持广播消息等高级功能。
type WebSocketManager struct {
	connections map[string]*WsConn
	rooms       map[string]map[string]*WsConn // room -> connections
	mutex       sync.RWMutex
}

// NewWebSocketManager 创建 WebSocket 管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[string]*WsConn),
		rooms:       make(map[string]map[string]*WsConn),
	}
}

// AddConnection 添加连接
func (wsm *WebSocketManager) AddConnection(id string, conn *WsConn) {
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

// ============= 错误处理 =============

// WebSocketError WebSocket 错误类型
type WebSocketError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Error 实现 error 接口
func (e *WebSocketError) Error() string {
	return fmt.Sprintf("WebSocket Error [%d:%s] %s", e.Code, e.Type, e.Message)
}

// 预定义错误类型
var (
	ErrWebSocketUpgradeFailed   = &WebSocketError{Code: 1001, Type: "upgrade_failed", Message: "WebSocket upgrade failed"}
	ErrWebSocketConnectionLost  = &WebSocketError{Code: 1002, Type: "connection_lost", Message: "WebSocket connection lost"}
	ErrWebSocketMessageTooLarge = &WebSocketError{Code: 1003, Type: "message_too_large", Message: "WebSocket message too large"}
	ErrWebSocketInvalidMessage  = &WebSocketError{Code: 1004, Type: "invalid_message", Message: "Invalid WebSocket message format"}
	ErrWebSocketRateLimited     = &WebSocketError{Code: 1005, Type: "rate_limited", Message: "WebSocket rate limit exceeded"}
)

// WebSocketErrorHandler WebSocket 错误处理器
type WebSocketErrorHandler struct {
	OnUpgradeFailed  func(error)
	OnConnectionLost func(*websocket.Conn, error)
	OnMessageError   func(*websocket.Conn, error)
	OnRateLimited    func(*websocket.Conn, string)
	OnUnknownError   func(*websocket.Conn, error)
}

// DefaultWebSocketErrorHandler 默认错误处理器
var DefaultWebSocketErrorHandler = &WebSocketErrorHandler{
	OnUpgradeFailed: func(err error) {
		log.Printf("WebSocket upgrade failed: %v", err)
	},
	OnConnectionLost: func(conn *websocket.Conn, err error) {
		log.Printf("WebSocket connection lost: %v", err)
	},
	OnMessageError: func(conn *websocket.Conn, err error) {
		log.Printf("WebSocket message error: %v", err)
	},
	OnRateLimited: func(conn *websocket.Conn, reason string) {
		log.Printf("WebSocket rate limited: %s", reason)
	},
	OnUnknownError: func(conn *websocket.Conn, err error) {
		log.Printf("WebSocket unknown error: %v", err)
	},
}

// ============= 连接池管理 =============

// WebSocketPool WebSocket 连接池
type WebSocketPool struct {
	connections map[string]*WebSocketConnection
	mutex       sync.RWMutex
	maxConns    int
	cleanupTime time.Duration
	stopCleanup chan bool
}

// WebSocketConnection WebSocket 连接包装器
type WebSocketConnection struct {
	ID          string
	Conn        *websocket.Conn
	ConnectedAt time.Time
	LastPing    time.Time
	UserData    map[string]interface{}
	mutex       sync.RWMutex
}

// NewWebSocketPool 创建 WebSocket 连接池
func NewWebSocketPool(maxConns int, cleanupInterval time.Duration) *WebSocketPool {
	pool := &WebSocketPool{
		connections: make(map[string]*WebSocketConnection),
		maxConns:    maxConns,
		cleanupTime: cleanupInterval,
		stopCleanup: make(chan bool),
	}

	// 启动清理 goroutine
	go pool.cleanupRoutine()

	return pool
}

// AddConnection 添加连接到池
func (p *WebSocketPool) AddConnection(id string, conn *websocket.Conn) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 检查连接数限制
	if len(p.connections) >= p.maxConns {
		return fmt.Errorf("connection pool is full (max: %d)", p.maxConns)
	}

	wsConn := &WebSocketConnection{
		ID:          id,
		Conn:        conn,
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
		UserData:    make(map[string]interface{}),
	}

	p.connections[id] = wsConn
	log.Printf("WebSocket connection added to pool: %s (total: %d)", id, len(p.connections))

	return nil
}

// RemoveConnection 从池中移除连接
func (p *WebSocketPool) RemoveConnection(id string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if conn, exists := p.connections[id]; exists {
		conn.Conn.Close()
		delete(p.connections, id)
		log.Printf("WebSocket connection removed from pool: %s (total: %d)", id, len(p.connections))
	}
}

// GetConnection 获取连接
func (p *WebSocketPool) GetConnection(id string) (*WebSocketConnection, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	conn, exists := p.connections[id]
	return conn, exists
}

// GetAllConnections 获取所有连接
func (p *WebSocketPool) GetAllConnections() map[string]*WebSocketConnection {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	result := make(map[string]*WebSocketConnection)
	for id, conn := range p.connections {
		result[id] = conn
	}

	return result
}

// BroadcastMessage 向所有连接广播消息
func (p *WebSocketPool) BroadcastMessage(messageType int, data []byte) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for id, wsConn := range p.connections {
		err := wsConn.Conn.WriteMessage(messageType, data)
		if err != nil {
			log.Printf("Failed to broadcast to connection %s: %v", id, err)
			// 异步移除失败的连接
			go p.RemoveConnection(id)
		}
	}
}

// BroadcastJSON 向所有连接广播 JSON 消息
func (p *WebSocketPool) BroadcastJSON(data interface{}) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for id, wsConn := range p.connections {
		err := wsConn.Conn.WriteJSON(data)
		if err != nil {
			log.Printf("Failed to broadcast JSON to connection %s: %v", id, err)
			go p.RemoveConnection(id)
		}
	}
}

// GetConnectionCount 获取连接数
func (p *WebSocketPool) GetConnectionCount() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.connections)
}

// Close 关闭连接池
func (p *WebSocketPool) Close() {
	p.stopCleanup <- true

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for id, wsConn := range p.connections {
		wsConn.Conn.Close()
		delete(p.connections, id)
	}

	log.Printf("WebSocket pool closed")
}

// cleanupRoutine 清理过期连接
func (p *WebSocketPool) cleanupRoutine() {
	ticker := time.NewTicker(p.cleanupTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanupStaleConnections()
		case <-p.stopCleanup:
			return
		}
	}
}

// cleanupStaleConnections 清理失效连接
func (p *WebSocketPool) cleanupStaleConnections() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	now := time.Now()
	staleTimeout := 5 * time.Minute // 5分钟没有ping就认为连接失效

	for id, wsConn := range p.connections {
		if now.Sub(wsConn.LastPing) > staleTimeout {
			log.Printf("Removing stale WebSocket connection: %s", id)
			wsConn.Conn.Close()
			delete(p.connections, id)
		}
	}
}

// UpdatePing 更新连接的ping时间
func (wsConn *WebSocketConnection) UpdatePing() {
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	wsConn.LastPing = time.Now()
}

// SetUserData 设置用户数据
func (wsConn *WebSocketConnection) SetUserData(key string, value interface{}) {
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	wsConn.UserData[key] = value
}

// GetUserData 获取用户数据
func (wsConn *WebSocketConnection) GetUserData(key string) (interface{}, bool) {
	wsConn.mutex.RLock()
	defer wsConn.mutex.RUnlock()
	value, exists := wsConn.UserData[key]
	return value, exists
}

// ============= 心跳机制 =============

// WebSocketHeartbeat WebSocket 心跳管理器
type WebSocketHeartbeat struct {
	pool      *WebSocketPool
	interval  time.Duration
	timeout   time.Duration
	stopChan  chan bool
	isRunning bool
	mutex     sync.RWMutex
}

// NewWebSocketHeartbeat 创建心跳管理器
func NewWebSocketHeartbeat(pool *WebSocketPool, interval, timeout time.Duration) *WebSocketHeartbeat {
	return &WebSocketHeartbeat{
		pool:     pool,
		interval: interval,
		timeout:  timeout,
		stopChan: make(chan bool),
	}
}

// Start 启动心跳
func (h *WebSocketHeartbeat) Start() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.isRunning {
		return
	}

	h.isRunning = true
	go h.heartbeatRoutine()
	log.Printf("WebSocket heartbeat started (interval: %v, timeout: %v)", h.interval, h.timeout)
}

// Stop 停止心跳
func (h *WebSocketHeartbeat) Stop() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.isRunning {
		return
	}

	h.isRunning = false
	h.stopChan <- true
	log.Printf("WebSocket heartbeat stopped")
}

// heartbeatRoutine 心跳检测例程
func (h *WebSocketHeartbeat) heartbeatRoutine() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.sendPingToAllConnections()
		case <-h.stopChan:
			return
		}
	}
}

// sendPingToAllConnections 向所有连接发送ping
func (h *WebSocketHeartbeat) sendPingToAllConnections() {
	connections := h.pool.GetAllConnections()

	for id, wsConn := range connections {
		err := wsConn.Conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			log.Printf("Failed to send ping to connection %s: %v", id, err)
			h.pool.RemoveConnection(id)
		} else {
			wsConn.UpdatePing()
		}
	}
}

// ============= 限流机制 =============

// WebSocketRateLimiter WebSocket 限流器
type WebSocketRateLimiter struct {
	rates   map[string]*rateLimitInfo
	mutex   sync.RWMutex
	maxRate int           // 每秒最大消息数
	window  time.Duration // 时间窗口
}

type rateLimitInfo struct {
	count     int
	resetTime time.Time
}

// NewWebSocketRateLimiter 创建限流器
func NewWebSocketRateLimiter(maxRate int, window time.Duration) *WebSocketRateLimiter {
	return &WebSocketRateLimiter{
		rates:   make(map[string]*rateLimitInfo),
		maxRate: maxRate,
		window:  window,
	}
}

// CheckRate 检查是否超过限流
func (rl *WebSocketRateLimiter) CheckRate(connectionID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	info, exists := rl.rates[connectionID]
	if !exists {
		rl.rates[connectionID] = &rateLimitInfo{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	// 检查是否需要重置窗口
	if now.After(info.resetTime) {
		info.count = 1
		info.resetTime = now.Add(rl.window)
		return true
	}

	// 检查是否超过限制
	if info.count >= rl.maxRate {
		return false
	}

	info.count++
	return true
}

// CleanupOldEntries 清理过期条目
func (rl *WebSocketRateLimiter) CleanupOldEntries() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	for id, info := range rl.rates {
		if now.After(info.resetTime.Add(rl.window)) {
			delete(rl.rates, id)
		}
	}
}

// ============= 监控和统计 =============

// WebSocketMetrics WebSocket 监控指标
type WebSocketMetrics struct {
	ConnectionsCreated int64
	ConnectionsClosed  int64
	MessagesSent       int64
	MessagesReceived   int64
	ErrorsOccurred     int64
	UpgradeFailures    int64
	CurrentConnections int64
	PeakConnections    int64
	mutex              sync.RWMutex
}

// NewWebSocketMetrics 创建监控指标
func NewWebSocketMetrics() *WebSocketMetrics {
	return &WebSocketMetrics{}
}

// IncrementConnectionsCreated 增加连接创建计数
func (m *WebSocketMetrics) IncrementConnectionsCreated() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ConnectionsCreated++
	m.CurrentConnections++
	if m.CurrentConnections > m.PeakConnections {
		m.PeakConnections = m.CurrentConnections
	}
}

// IncrementConnectionsClosed 增加连接关闭计数
func (m *WebSocketMetrics) IncrementConnectionsClosed() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ConnectionsClosed++
	if m.CurrentConnections > 0 {
		m.CurrentConnections--
	}
}

// IncrementMessagesSent 增加发送消息计数
func (m *WebSocketMetrics) IncrementMessagesSent() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.MessagesSent++
}

// IncrementMessagesReceived 增加接收消息计数
func (m *WebSocketMetrics) IncrementMessagesReceived() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.MessagesReceived++
}

// IncrementErrors 增加错误计数
func (m *WebSocketMetrics) IncrementErrors() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ErrorsOccurred++
}

// IncrementUpgradeFailures 增加升级失败计数
func (m *WebSocketMetrics) IncrementUpgradeFailures() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.UpgradeFailures++
}

// GetMetrics 获取当前指标
func (m *WebSocketMetrics) GetMetrics() map[string]int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]int64{
		"connections_created": m.ConnectionsCreated,
		"connections_closed":  m.ConnectionsClosed,
		"messages_sent":       m.MessagesSent,
		"messages_received":   m.MessagesReceived,
		"errors_occurred":     m.ErrorsOccurred,
		"upgrade_failures":    m.UpgradeFailures,
		"current_connections": m.CurrentConnections,
		"peak_connections":    m.PeakConnections,
	}
}

// 全局实例
var (
	GlobalWebSocketPool    = NewWebSocketPool(1000, time.Minute*5)
	GlobalWebSocketMetrics = NewWebSocketMetrics()
)

// ============= 高级 WebSocket 处理器 =============

// EnhancedWebSocketHandler 增强的 WebSocket 处理器
type EnhancedWebSocketHandler struct {
	pool         *WebSocketPool
	rateLimiter  *WebSocketRateLimiter
	heartbeat    *WebSocketHeartbeat
	metrics      *WebSocketMetrics
	errorHandler *WebSocketErrorHandler
}

// NewEnhancedWebSocketHandler 创建增强的 WebSocket 处理器
func NewEnhancedWebSocketHandler() *EnhancedWebSocketHandler {
	pool := NewWebSocketPool(1000, time.Minute*5)
	rateLimiter := NewWebSocketRateLimiter(100, time.Minute)
	heartbeat := NewWebSocketHeartbeat(pool, time.Second*30, time.Second*10)
	metrics := NewWebSocketMetrics()

	handler := &EnhancedWebSocketHandler{
		pool:         pool,
		rateLimiter:  rateLimiter,
		heartbeat:    heartbeat,
		metrics:      metrics,
		errorHandler: DefaultWebSocketErrorHandler,
	}

	// 启动心跳
	heartbeat.Start()

	return handler
}

// HandleConnection 处理 WebSocket 连接
func (h *EnhancedWebSocketHandler) HandleConnection(connectionID string, conn *websocket.Conn, messageHandler func([]byte) []byte) {
	// 添加连接到池
	if err := h.pool.AddConnection(connectionID, conn); err != nil {
		h.errorHandler.OnUpgradeFailed(err)
		h.metrics.IncrementUpgradeFailures()
		return
	}

	h.metrics.IncrementConnectionsCreated()
	defer func() {
		h.pool.RemoveConnection(connectionID)
		h.metrics.IncrementConnectionsClosed()
	}()

	// 消息处理循环
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.errorHandler.OnConnectionLost(conn, err)
				h.metrics.IncrementErrors()
			}
			break
		}

		h.metrics.IncrementMessagesReceived()

		// 检查限流
		if !h.rateLimiter.CheckRate(connectionID) {
			h.errorHandler.OnRateLimited(conn, "Message rate limit exceeded")
			continue
		}

		// 处理消息
		response := messageHandler(message)
		if response != nil {
			err = conn.WriteMessage(messageType, response)
			if err != nil {
				h.errorHandler.OnMessageError(conn, err)
				h.metrics.IncrementErrors()
				break
			}
			h.metrics.IncrementMessagesSent()
		}
	}
}

// Close 关闭处理器
func (h *EnhancedWebSocketHandler) Close() {
	h.heartbeat.Stop()
	h.pool.Close()
}
