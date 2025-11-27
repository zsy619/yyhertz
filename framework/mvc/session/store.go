package session

import (
	"sync"
	"time"
)

// Store Session存储接口
type Store interface {
	Get(key string) any
	Set(key string, value any)
	Delete(key string)
	Clear()
	GetID() string
	Destroy()
	Save() error
	Exists(key string) bool
	GetAll() map[string]any
}

// MemoryStore 内存Session存储实现
type MemoryStore struct {
	mutex sync.RWMutex

	SessionID  string
	Data       map[string]any
	Modified   bool
	CreateTime time.Time
	LastAccess time.Time
}

// ============= 全局Session存储池 =============

// SessionStorePool 全局Session存储池
type SessionStorePool struct {
	Stores sync.Map // key: sessionID, value: *MemoryStore
}

// 全局Session存储池实例
var globalStorePool = &SessionStorePool{}

// GetSessionStorePool 获取全局Session存储池
func GetSessionStorePool() *SessionStorePool {
	return globalStorePool
}

// GetOrCreate 从池中获取或创建Session存储
func (p *SessionStorePool) GetOrCreate(sessionID string) *MemoryStore {
	if sessionID == "" {
		return nil
	}

	// 先尝试从池中获取现有的
	if value, ok := p.Stores.Load(sessionID); ok {
		if store, ok := value.(*MemoryStore); ok {
			// 更新最后访问时间
			store.mutex.Lock()
			store.LastAccess = time.Now()
			store.mutex.Unlock()
			return store
		}
	}

	// 不存在则创建新的
	now := time.Now()
	store := &MemoryStore{
		SessionID:  sessionID,
		Data:       make(map[string]any),
		CreateTime: now,
		LastAccess: now,
	}

	// 存储到池中
	p.Stores.Store(sessionID, store)
	return store
}

// Remove 从池中移除Session存储
func (p *SessionStorePool) Remove(sessionID string) {
	if sessionID != "" {
		p.Stores.Delete(sessionID)
	}
}

// Cleanup 清理过期的Session存储
func (p *SessionStorePool) Cleanup(maxAge time.Duration) int {
	if maxAge <= 0 {
		return 0
	}

	cleaned := 0
	now := time.Now()

	p.Stores.Range(func(key, value interface{}) bool {
		if sessionID, ok := key.(string); ok {
			if store, ok := value.(*MemoryStore); ok {
				store.mutex.RLock()
				expired := now.Sub(store.LastAccess) > maxAge
				store.mutex.RUnlock()

				if expired {
					p.Stores.Delete(sessionID)
					cleaned++
				}
			}
		}
		return true
	})

	return cleaned
}

// Count 获取当前Session存储数量
func (p *SessionStorePool) Count() int {
	count := 0
	p.Stores.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// GetAll 获取所有Session存储（用于调试）
func (p *SessionStorePool) GetAll() map[string]*MemoryStore {
	result := make(map[string]*MemoryStore)
	p.Stores.Range(func(key, value interface{}) bool {
		if sessionID, ok := key.(string); ok {
			if store, ok := value.(*MemoryStore); ok {
				result[sessionID] = store
			}
		}
		return true
	})
	return result
}

// NewMemoryStore 创建内存Session存储（已废弃，使用GetOrCreateMemoryStore）
// 保留此函数是为了向后兼容，但推荐使用GetOrCreateMemoryStore
func NewMemoryStore(id string) *MemoryStore {
	return GetOrCreateMemoryStore(id)
}

// GetOrCreateMemoryStore 获取或创建内存Session存储（推荐使用）
// 这个函数使用全局存储池，确保同一个sessionID返回同一个store实例
func GetOrCreateMemoryStore(id string) *MemoryStore {
	return GetSessionStorePool().GetOrCreate(id)
}

func (s *MemoryStore) Get(key string) any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Data[key]
}

func (s *MemoryStore) Set(key string, value any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Data[key] = value
	s.Modified = true
}

func (s *MemoryStore) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.Data, key)
	s.Modified = true
}

func (s *MemoryStore) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Data = make(map[string]any)
	s.Modified = true
}

func (s *MemoryStore) GetID() string {
	return s.SessionID
}

func (s *MemoryStore) Destroy() {
	// 先清空数据
	s.Clear()
	// 从全局存储池中移除
	GetSessionStorePool().Remove(s.SessionID)
}

func (s *MemoryStore) Save() error {
	// 内存Session不需要持久化
	s.Modified = false
	return nil
}

func (s *MemoryStore) Exists(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.Data[key]
	return exists
}

func (s *MemoryStore) GetAll() map[string]any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	copy := make(map[string]any)
	for k, v := range s.Data {
		copy[k] = v
	}
	return copy
}
