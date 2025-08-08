package captcha

import (
	"fmt"
	"sync"
	"time"
)

// MemoryStore 内存存储实现
type MemoryStore struct {
	mu    sync.RWMutex
	data  map[string]*Captcha
	close chan struct{}
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		data:  make(map[string]*Captcha),
		close: make(chan struct{}),
	}
	
	// 启动清理协程
	go store.cleanup()
	
	return store
}

// Set 存储验证码
func (s *MemoryStore) Set(id string, captcha *Captcha) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = captcha
	return nil
}

// Get 获取验证码
func (s *MemoryStore) Get(id string) (*Captcha, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	captcha, exists := s.data[id]
	if !exists {
		return nil, fmt.Errorf("captcha not found")
	}
	
	return captcha, nil
}

// Delete 删除验证码
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
	return nil
}

// Clear 清空所有验证码
func (s *MemoryStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]*Captcha)
	return nil
}

// Close 关闭存储
func (s *MemoryStore) Close() {
	close(s.close)
}

// cleanup 定期清理过期验证码
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.cleanExpired()
		case <-s.close:
			return
		}
	}
}

// cleanExpired 清理过期验证码
func (s *MemoryStore) cleanExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now().Unix()
	for id, captcha := range s.data {
		if now > captcha.ExpireAt {
			delete(s.data, id)
		}
	}
}

// SessionStore Session存储实现（适配框架的session）
type SessionStore struct {
	sessionKey string
	getSession func() interface{}
	setSession func(key string, value interface{}) error
}

// NewSessionStore 创建Session存储
func NewSessionStore(sessionKey string, getter func() interface{}, setter func(string, interface{}) error) *SessionStore {
	return &SessionStore{
		sessionKey: sessionKey,
		getSession: getter,
		setSession: setter,
	}
}

// Set 存储验证码到Session
func (s *SessionStore) Set(id string, captcha *Captcha) error {
	sessionData := s.getSessionData()
	sessionData[id] = captcha
	return s.setSession(s.sessionKey, sessionData)
}

// Get 从Session获取验证码
func (s *SessionStore) Get(id string) (*Captcha, error) {
	sessionData := s.getSessionData()
	captcha, exists := sessionData[id]
	if !exists {
		return nil, fmt.Errorf("captcha not found")
	}
	return captcha, nil
}

// Delete 从Session删除验证码
func (s *SessionStore) Delete(id string) error {
	sessionData := s.getSessionData()
	delete(sessionData, id)
	return s.setSession(s.sessionKey, sessionData)
}

// Clear 清空Session中的所有验证码
func (s *SessionStore) Clear() error {
	return s.setSession(s.sessionKey, make(map[string]*Captcha))
}

// getSessionData 获取session数据
func (s *SessionStore) getSessionData() map[string]*Captcha {
	session := s.getSession()
	if session == nil {
		return make(map[string]*Captcha)
	}
	
	data, ok := session.(map[string]*Captcha)
	if !ok {
		return make(map[string]*Captcha)
	}
	
	return data
}

// RedisStore Redis存储实现（如果项目使用Redis）
type RedisStore struct {
	client    interface{} // Redis客户端接口
	keyPrefix string
	ttl       time.Duration
}

// NewRedisStore 创建Redis存储
func NewRedisStore(client interface{}, keyPrefix string, ttl time.Duration) *RedisStore {
	return &RedisStore{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

// Set 存储验证码到Redis
func (s *RedisStore) Set(id string, captcha *Captcha) error {
	// 这里需要根据实际使用的Redis客户端实现
	// 例如：go-redis/redis 或 gomodule/redigo
	// 暂时返回未实现错误
	return fmt.Errorf("redis store not implemented yet")
}

// Get 从Redis获取验证码
func (s *RedisStore) Get(id string) (*Captcha, error) {
	return nil, fmt.Errorf("redis store not implemented yet")
}

// Delete 从Redis删除验证码
func (s *RedisStore) Delete(id string) error {
	return fmt.Errorf("redis store not implemented yet")
}

// Clear 清空Redis中的所有验证码
func (s *RedisStore) Clear() error {
	return fmt.Errorf("redis store not implemented yet")
}
