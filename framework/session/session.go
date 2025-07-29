package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// SessionData Session数据结构
type SessionData struct {
	Data       map[string]any
	LastAccess time.Time
	ExpireTime time.Time
}

// NewSessionData 创建新的Session数据
func NewSessionData(lifetime time.Duration) *SessionData {
	now := time.Now()
	return &SessionData{
		Data:       make(map[string]any),
		LastAccess: now,
		ExpireTime: now.Add(lifetime),
	}
}

// IsExpired 检查Session是否过期
func (s *SessionData) IsExpired() bool {
	return time.Now().After(s.ExpireTime)
}

// Touch 更新最后访问时间
func (s *SessionData) Touch(lifetime time.Duration) {
	now := time.Now()
	s.LastAccess = now
	s.ExpireTime = now.Add(lifetime)
}

// SessionManager Session管理器
type SessionManager struct {
	sessions map[string]*SessionData
	mutex    sync.RWMutex
	lifetime time.Duration
	name     string
}

// NewSessionManager 创建Session管理器
func NewSessionManager(lifetime time.Duration, sessionName string) *SessionManager {
	if sessionName == "" {
		sessionName = "HERTZ_SESSION_ID"
	}
	if lifetime <= 0 {
		lifetime = 30 * time.Minute
	}
	
	sm := &SessionManager{
		sessions: make(map[string]*SessionData),
		lifetime: lifetime,
		name:     sessionName,
	}
	
	// 启动清理过期Session的goroutine
	go sm.gcLoop()
	
	return sm
}

// GenerateSessionID 生成Session ID
func (sm *SessionManager) GenerateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession 创建新Session
func (sm *SessionManager) CreateSession() (string, *SessionData) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sessionID := sm.GenerateSessionID()
	sessionData := NewSessionData(sm.lifetime)
	sm.sessions[sessionID] = sessionData
	
	return sessionID, sessionData
}

// GetSession 获取Session
func (sm *SessionManager) GetSession(sessionID string) (*SessionData, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	sessionData, exists := sm.sessions[sessionID]
	if !exists {
		return nil, false
	}
	
	if sessionData.IsExpired() {
		delete(sm.sessions, sessionID)
		return nil, false
	}
	
	sessionData.Touch(sm.lifetime)
	return sessionData, true
}

// DestroySession 销毁Session
func (sm *SessionManager) DestroySession(sessionID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.sessions, sessionID)
}

// Set 设置Session值
func (sm *SessionManager) Set(sessionID string, key string, value any) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sessionData, exists := sm.sessions[sessionID]
	if !exists || sessionData.IsExpired() {
		return false
	}
	
	sessionData.Data[key] = value
	sessionData.Touch(sm.lifetime)
	return true
}

// Get 获取Session值
func (sm *SessionManager) Get(sessionID string, key string) (any, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	sessionData, exists := sm.sessions[sessionID]
	if !exists || sessionData.IsExpired() {
		return nil, false
	}
	
	value, exists := sessionData.Data[key]
	if exists {
		sessionData.Touch(sm.lifetime)
	}
	
	return value, exists
}

// GetString 获取字符串类型Session值
func (sm *SessionManager) GetString(sessionID string, key string) (string, bool) {
	value, exists := sm.Get(sessionID, key)
	if !exists {
		return "", false
	}
	
	if str, ok := value.(string); ok {
		return str, true
	}
	
	return "", false
}

// GetInt 获取整型Session值
func (sm *SessionManager) GetInt(sessionID string, key string) (int, bool) {
	value, exists := sm.Get(sessionID, key)
	if !exists {
		return 0, false
	}
	
	if i, ok := value.(int); ok {
		return i, true
	}
	
	return 0, false
}

// GetInt64 获取int64类型Session值
func (sm *SessionManager) GetInt64(sessionID string, key string) (int64, bool) {
	value, exists := sm.Get(sessionID, key)
	if !exists {
		return 0, false
	}
	
	if i, ok := value.(int64); ok {
		return i, true
	}
	
	return 0, false
}

// GetBool 获取布尔类型Session值
func (sm *SessionManager) GetBool(sessionID string, key string) (bool, bool) {
	value, exists := sm.Get(sessionID, key)
	if !exists {
		return false, false
	}
	
	if b, ok := value.(bool); ok {
		return b, true
	}
	
	return false, false
}

// Delete 删除Session值
func (sm *SessionManager) Delete(sessionID string, key string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sessionData, exists := sm.sessions[sessionID]
	if !exists || sessionData.IsExpired() {
		return false
	}
	
	delete(sessionData.Data, key)
	sessionData.Touch(sm.lifetime)
	return true
}

// Clear 清空Session所有数据
func (sm *SessionManager) Clear(sessionID string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sessionData, exists := sm.sessions[sessionID]
	if !exists || sessionData.IsExpired() {
		return false
	}
	
	sessionData.Data = make(map[string]any)
	sessionData.Touch(sm.lifetime)
	return true
}

// GetSessionName 获取Session名称
func (sm *SessionManager) GetSessionName() string {
	return sm.name
}

// GetLifetime 获取Session生命周期
func (sm *SessionManager) GetLifetime() time.Duration {
	return sm.lifetime
}

// GetSessionCount 获取Session数量
func (sm *SessionManager) GetSessionCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return len(sm.sessions)
}

// gcLoop 垃圾回收循环
func (sm *SessionManager) gcLoop() {
	ticker := time.NewTicker(sm.lifetime / 2) // 每半个生命周期清理一次
	defer ticker.Stop()
	
	for range ticker.C {
		sm.gc()
	}
}

// gc 垃圾回收
func (sm *SessionManager) gc() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	now := time.Now()
	for sessionID, sessionData := range sm.sessions {
		if now.After(sessionData.ExpireTime) {
			delete(sm.sessions, sessionID)
		}
	}
}

// 全局Session管理器
var DefaultSessionManager = NewSessionManager(30*time.Minute, "HERTZ_SESSION_ID")

// 便捷函数
func CreateSession() (string, *SessionData) {
	return DefaultSessionManager.CreateSession()
}

func GetSession(sessionID string) (*SessionData, bool) {
	return DefaultSessionManager.GetSession(sessionID)
}

func DestroySession(sessionID string) {
	DefaultSessionManager.DestroySession(sessionID)
}

func Set(sessionID string, key string, value any) bool {
	return DefaultSessionManager.Set(sessionID, key, value)
}

func Get(sessionID string, key string) (any, bool) {
	return DefaultSessionManager.Get(sessionID, key)
}

func GetString(sessionID string, key string) (string, bool) {
	return DefaultSessionManager.GetString(sessionID, key)
}

func GetInt(sessionID string, key string) (int, bool) {
	return DefaultSessionManager.GetInt(sessionID, key)
}

func GetInt64(sessionID string, key string) (int64, bool) {
	return DefaultSessionManager.GetInt64(sessionID, key)
}

func GetBool(sessionID string, key string) (bool, bool) {
	return DefaultSessionManager.GetBool(sessionID, key)
}

func Delete(sessionID string, key string) bool {
	return DefaultSessionManager.Delete(sessionID, key)
}

func Clear(sessionID string) bool {
	return DefaultSessionManager.Clear(sessionID)
}

func GetSessionCount() int {
	return DefaultSessionManager.GetSessionCount()
}