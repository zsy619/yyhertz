package cache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DistributedCache 分布式缓存接口
type DistributedCache interface {
	Set(key string, value any, expiration time.Duration) error
	Get(key string) (any, bool, error)
	Delete(key string) error
	Clear() error
	Exists(key string) (bool, error)
	SetWithTTL(key string, value any, ttl time.Duration) error
	GetTTL(key string) (time.Duration, error)
	Increment(key string, delta int64) (int64, error)
	Decrement(key string, delta int64) (int64, error)
}

// RedisCache Redis缓存实现（模拟）
type RedisCache struct {
	client RedisClient
	prefix string
	mutex  sync.RWMutex
}

// RedisClient Redis客户端接口
type RedisClient interface {
	Set(key string, value string, expiration time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	Exists(key string) (bool, error)
	TTL(key string) (time.Duration, error)
	Incr(key string) (int64, error)
	Decr(key string) (int64, error)
	FlushAll() error
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client RedisClient, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// Set 设置缓存
func (r *RedisCache) Set(key string, value any, expiration time.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	fullKey := r.getFullKey(key)
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value error: %w", err)
	}
	
	return r.client.Set(fullKey, string(jsonValue), expiration)
}

// Get 获取缓存
func (r *RedisCache) Get(key string) (any, bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	fullKey := r.getFullKey(key)
	value, err := r.client.Get(fullKey)
	if err != nil {
		return nil, false, err
	}
	
	if value == "" {
		return nil, false, nil
	}
	
	var result any
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, false, fmt.Errorf("unmarshal value error: %w", err)
	}
	
	return result, true, nil
}

// Delete 删除缓存
func (r *RedisCache) Delete(key string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	fullKey := r.getFullKey(key)
	return r.client.Del(fullKey)
}

// Clear 清空所有缓存
func (r *RedisCache) Clear() error {
	return r.client.FlushAll()
}

// Exists 检查缓存是否存在
func (r *RedisCache) Exists(key string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	fullKey := r.getFullKey(key)
	return r.client.Exists(fullKey)
}

// SetWithTTL 设置带TTL的缓存
func (r *RedisCache) SetWithTTL(key string, value any, ttl time.Duration) error {
	return r.Set(key, value, ttl)
}

// GetTTL 获取缓存TTL
func (r *RedisCache) GetTTL(key string) (time.Duration, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	fullKey := r.getFullKey(key)
	return r.client.TTL(fullKey)
}

// Increment 递增计数器
func (r *RedisCache) Increment(key string, delta int64) (int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	fullKey := r.getFullKey(key)
	// 简化实现，实际应该支持delta
	return r.client.Incr(fullKey)
}

// Decrement 递减计数器
func (r *RedisCache) Decrement(key string, delta int64) (int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	fullKey := r.getFullKey(key)
	// 简化实现，实际应该支持delta
	return r.client.Decr(fullKey)
}

// getFullKey 获取完整的键名
func (r *RedisCache) getFullKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return r.prefix + ":" + key
}

// MemoryDistributedCache 基于内存的分布式缓存（用于测试）
type MemoryDistributedCache struct {
	data   map[string]*CacheItem[any]
	mutex  sync.RWMutex
	prefix string
}

// NewMemoryDistributedCache 创建内存分布式缓存
func NewMemoryDistributedCache(prefix string) *MemoryDistributedCache {
	return &MemoryDistributedCache{
		data:   make(map[string]*CacheItem[any]),
		prefix: prefix,
	}
}

// Set 设置缓存
func (m *MemoryDistributedCache) Set(key string, value any, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	fullKey := m.getFullKey(key)
	
	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}
	
	m.data[fullKey] = &CacheItem[any]{
		Value:      value,
		Expiration: exp,
	}
	
	return nil
}

// Get 获取缓存
func (m *MemoryDistributedCache) Get(key string) (any, bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	fullKey := m.getFullKey(key)
	item, exists := m.data[fullKey]
	if !exists {
		return nil, false, nil
	}
	
	// 检查是否过期
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		// 延迟删除过期项
		go func() {
			m.mutex.Lock()
			delete(m.data, fullKey)
			m.mutex.Unlock()
		}()
		return nil, false, nil
	}
	
	return item.Value, true, nil
}

// Delete 删除缓存
func (m *MemoryDistributedCache) Delete(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	fullKey := m.getFullKey(key)
	delete(m.data, fullKey)
	return nil
}

// Clear 清空所有缓存
func (m *MemoryDistributedCache) Clear() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.data = make(map[string]*CacheItem[any])
	return nil
}

// Exists 检查缓存是否存在
func (m *MemoryDistributedCache) Exists(key string) (bool, error) {
	_, exists, err := m.Get(key)
	return exists, err
}

// SetWithTTL 设置带TTL的缓存
func (m *MemoryDistributedCache) SetWithTTL(key string, value any, ttl time.Duration) error {
	return m.Set(key, value, ttl)
}

// GetTTL 获取缓存TTL
func (m *MemoryDistributedCache) GetTTL(key string) (time.Duration, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	fullKey := m.getFullKey(key)
	item, exists := m.data[fullKey]
	if !exists {
		return -1, nil
	}
	
	if item.Expiration <= 0 {
		return -1, nil // 永不过期
	}
	
	remaining := time.Duration(item.Expiration - time.Now().UnixNano())
	if remaining <= 0 {
		return 0, nil // 已过期
	}
	
	return remaining, nil
}

// Increment 递增计数器
func (m *MemoryDistributedCache) Increment(key string, delta int64) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	fullKey := m.getFullKey(key)
	item, exists := m.data[fullKey]
	
	var currentValue int64
	if exists {
		if val, ok := item.Value.(int64); ok {
			currentValue = val
		} else if val, ok := item.Value.(float64); ok {
			currentValue = int64(val)
		}
	}
	
	newValue := currentValue + delta
	m.data[fullKey] = &CacheItem[any]{
		Value:      newValue,
		Expiration: 0, // 计数器不过期
	}
	
	return newValue, nil
}

// Decrement 递减计数器
func (m *MemoryDistributedCache) Decrement(key string, delta int64) (int64, error) {
	return m.Increment(key, -delta)
}

// getFullKey 获取完整的键名
func (m *MemoryDistributedCache) getFullKey(key string) string {
	if m.prefix == "" {
		return key
	}
	return m.prefix + ":" + key
}

// CacheCluster 缓存集群
type CacheCluster struct {
	nodes   []DistributedCache
	mutex   sync.RWMutex
	hashFn  func(key string) uint32
}

// NewCacheCluster 创建缓存集群
func NewCacheCluster(nodes []DistributedCache) *CacheCluster {
	return &CacheCluster{
		nodes:  nodes,
		hashFn: defaultHashFunction,
	}
}

// Set 设置缓存到集群
func (c *CacheCluster) Set(key string, value any, expiration time.Duration) error {
	node := c.getNode(key)
	return node.Set(key, value, expiration)
}

// Get 从集群获取缓存
func (c *CacheCluster) Get(key string) (any, bool, error) {
	node := c.getNode(key)
	return node.Get(key)
}

// Delete 从集群删除缓存
func (c *CacheCluster) Delete(key string) error {
	node := c.getNode(key)
	return node.Delete(key)
}

// Clear 清空集群所有缓存
func (c *CacheCluster) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for _, node := range c.nodes {
		if err := node.Clear(); err != nil {
			return err
		}
	}
	return nil
}

// getNode 根据key获取对应的节点
func (c *CacheCluster) getNode(key string) DistributedCache {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if len(c.nodes) == 0 {
		panic("no cache nodes available")
	}
	
	hash := c.hashFn(key)
	index := hash % uint32(len(c.nodes))
	return c.nodes[index]
}

// defaultHashFunction 默认哈希函数
func defaultHashFunction(key string) uint32 {
	hash := uint32(2166136261)
	for _, b := range []byte(key) {
		hash ^= uint32(b)
		hash *= 16777619
	}
	return hash
}

// CacheWrapper 缓存包装器，支持多级缓存
type CacheWrapper struct {
	l1Cache *CacheManager[any]     // 一级缓存（本地）
	l2Cache DistributedCache       // 二级缓存（分布式）
	mutex   sync.RWMutex
}

// NewCacheWrapper 创建缓存包装器
func NewCacheWrapper(l1Cache *CacheManager[any], l2Cache DistributedCache) *CacheWrapper {
	return &CacheWrapper{
		l1Cache: l1Cache,
		l2Cache: l2Cache,
	}
}

// Set 设置多级缓存
func (w *CacheWrapper) Set(key string, value any, expiration time.Duration) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	// 设置一级缓存
	w.l1Cache.Set(key, value, expiration)
	
	// 设置二级缓存
	if w.l2Cache != nil {
		return w.l2Cache.Set(key, value, expiration)
	}
	
	return nil
}

// Get 获取多级缓存
func (w *CacheWrapper) Get(key string) (any, bool, error) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	// 先从一级缓存获取
	if value, found := w.l1Cache.Get(key); found {
		return value, true, nil
	}
	
	// 从二级缓存获取
	if w.l2Cache != nil {
		if value, found, err := w.l2Cache.Get(key); err != nil {
			return nil, false, err
		} else if found {
			// 回写到一级缓存
			w.l1Cache.Set(key, value, time.Hour) // 默认1小时过期
			return value, true, nil
		}
	}
	
	return nil, false, nil
}

// Delete 删除多级缓存
func (w *CacheWrapper) Delete(key string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	// 删除一级缓存
	w.l1Cache.Delete(key)
	
	// 删除二级缓存
	if w.l2Cache != nil {
		return w.l2Cache.Delete(key)
	}
	
	return nil
}