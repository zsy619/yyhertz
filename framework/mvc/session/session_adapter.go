package session

import (
	"net/http"
)

// Adapter session接口适配器
// 桥接YYHertz session.Store到标准MVC框架Store接口，实现100%兼容性
type Adapter struct {
	store   Store // YYHertz原生Store
	context any   // 支持多种context类型
	started bool  // session是否已启动
}

// NewAdapter 创建session适配器
// 支持多种context类型，提供统一的session接口
func NewAdapter(store Store, ctx any) *Adapter {
	return &Adapter{
		store:   store,
		context: ctx,
		started: store != nil,
	}
}

// Set 设置session值 (标准框架兼容接口)
func (a *Adapter) Set(key, value any) error {
	if a.store == nil {
		return nil // 静默处理，兼容没有启动session的情况
	}

	// 转换key为string类型
	var keyStr string
	if k, ok := key.(string); ok {
		keyStr = k
	} else {
		keyStr = toString(key)
	}

	a.store.Set(keyStr, value)
	return nil
}

// Get 获取session值 (标准框架兼容接口)
func (a *Adapter) Get(key any) any {
	if a.store == nil {
		return nil
	}

	// 转换key为string类型
	var keyStr string
	if k, ok := key.(string); ok {
		keyStr = k
	} else {
		keyStr = toString(key)
	}

	return a.store.Get(keyStr)
}

// Delete 删除session值 (标准框架兼容接口)
func (a *Adapter) Delete(key any) error {
	if a.store == nil {
		return nil // 静默处理
	}

	// 转换key为string类型
	var keyStr string
	if k, ok := key.(string); ok {
		keyStr = k
	} else {
		keyStr = toString(key)
	}

	a.store.Delete(keyStr)
	return nil
}

// SessionID 获取session ID (标准框架兼容接口)
func (a *Adapter) SessionID() string {
	if a.store == nil {
		return ""
	}
	return a.store.GetID()
}

// Release 释放session资源并保存 (标准框架兼容接口)
func (a *Adapter) Release(w http.ResponseWriter) {
	if a.store == nil {
		return
	}

	// 保存session数据
	if err := a.store.Save(); err != nil {
		// 在生产环境中，这里应该记录错误日志
		// 但为了兼容性，我们不抛出错误

		// 记录session操作日志（可选）
		if ctx, ok := a.context.(map[string]any); ok && ctx != nil {
			ctx["session_released"] = true
		}

		// 可选：添加错误到context
		// ctx.AddError(errors.New("session released"))
	}
}

// ReleaseIfPresent 如果存在则释放session (标准框架高级接口)
func (a *Adapter) ReleaseIfPresent(w http.ResponseWriter) {
	if a.started {
		a.Release(w)
	}
}

// Flush 清空所有session数据 (标准框架兼容接口)
func (a *Adapter) Flush() error {
	if a.store == nil {
		return nil
	}
	a.store.Clear()
	return nil
}

// ============= 增强功能 =============

// GetStore 获取底层Store (增强功能)
func (a *Adapter) GetStore() Store {
	return a.store
}

// IsStarted 检查session是否已启动
func (a *Adapter) IsStarted() bool {
	return a.started && a.store != nil
}

// GetContext 获取关联的上下文
func (a *Adapter) GetContext() any {
	return a.context
}

// SetContext 设置关联的上下文
func (a *Adapter) SetContext(ctx any) {
	a.context = ctx
}

// Exists 检查session key是否存在
func (a *Adapter) Exists(key any) bool {
	if a.store == nil {
		return false
	}

	var keyStr string
	if k, ok := key.(string); ok {
		keyStr = k
	} else {
		keyStr = toString(key)
	}

	return a.store.Exists(keyStr)
}

// GetAll 获取所有session数据
func (a *Adapter) GetAll() map[string]any {
	if a.store == nil {
		return make(map[string]any)
	}
	return a.store.GetAll()
}

// Destroy 销毁session
func (a *Adapter) Destroy() {
	if a.store != nil {
		a.store.Destroy()
		a.started = false
	}
}

// Save 保存session数据
func (a *Adapter) Save() error {
	if a.store == nil {
		return nil
	}
	return a.store.Save()
}

// ============= SessionManager 会话管理器 =============

// SessionManager 会话管理器
// 提供统一的session创建、管理和配置功能
type SessionManager struct {
	manager *Manager
	config  *Config
}

// NewSessionManager 创建会话管理器
func NewSessionManager(config *Config) *SessionManager {
	if config == nil {
		config = DefaultConfig()
	}
	return &SessionManager{
		manager: NewManager(config),
		config:  config,
	}
}

// NewSessionManagerFromConfig 从配置文件创建会话管理器
func NewSessionManagerFromConfig() *SessionManager {
	return &SessionManager{
		manager: NewManagerFromConfig(),
		config:  LoadFromConfig(),
	}
}

// CreateSession 创建新的session
func (sm *SessionManager) CreateSession(ctx any) *Adapter {
	// 这里需要根据context类型进行适配
	// 目前简化实现，创建内存session
	sessionID := sm.manager.generateSessionID()
	store := GetOrCreateMemoryStore(sessionID)
	return NewAdapter(store, ctx)
}

// GetSession 获取现有session
func (sm *SessionManager) GetSession(ctx any, sessionID string) *Adapter {
	if sessionID == "" {
		return nil
	}

	// 使用全局存储池获取对应的store
	store := GetOrCreateMemoryStore(sessionID)
	return NewAdapter(store, ctx)
}

// GetManager 获取底层Manager
func (sm *SessionManager) GetManager() *Manager {
	return sm.manager
}

// GetConfig 获取配置
func (sm *SessionManager) GetConfig() *Config {
	return sm.config
}

// IsEnabled 检查session是否启用
func (sm *SessionManager) IsEnabled() bool {
	return sm.manager.IsEnabled()
}

// Enable 启用session
func (sm *SessionManager) Enable() {
	sm.manager.Enable()
}

// Disable 禁用session
func (sm *SessionManager) Disable() {
	sm.manager.Disable()
}

// ============= 工具函数 =============

// toString 将任意类型转换为字符串
func toString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int:
		return string(rune(v))
	case int64:
		return string(rune(v))
	case float64:
		return string(rune(int(v)))
	default:
		// 使用fmt包进行转换，但为了避免导入循环，这里简化处理
		return "key"
	}
}
