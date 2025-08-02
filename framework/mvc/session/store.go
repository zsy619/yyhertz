package session

import (
	"sync"
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
	id       string
	data     map[string]any
	modified bool
	mutex    sync.RWMutex
}

// NewMemoryStore 创建内存Session存储
func NewMemoryStore(id string) *MemoryStore {
	return &MemoryStore{
		id:   id,
		data: make(map[string]any),
	}
}

func (s *MemoryStore) Get(key string) any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data[key]
}

func (s *MemoryStore) Set(key string, value any) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
	s.modified = true
}

func (s *MemoryStore) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, key)
	s.modified = true
}

func (s *MemoryStore) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = make(map[string]any)
	s.modified = true
}

func (s *MemoryStore) GetID() string {
	return s.id
}

func (s *MemoryStore) Destroy() {
	s.Clear()
}

func (s *MemoryStore) Save() error {
	// 内存Session不需要持久化
	s.modified = false
	return nil
}

func (s *MemoryStore) Exists(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.data[key]
	return exists
}

func (s *MemoryStore) GetAll() map[string]any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	copy := make(map[string]any)
	for k, v := range s.data {
		copy[k] = v
	}
	return copy
}