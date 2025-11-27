package cache

import (
	"sync"
	"time"
)

// CacheItem 缓存项结构
type CacheItem[T any] struct {
	Value      T
	Expiration int64
}

// CacheManager 缓存管理器
type CacheManager[T any] struct {
	items map[string]*CacheItem[T]
	mutex sync.RWMutex
	name  string
	desc  string
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager[T any](name, desc string) *CacheManager[T] {
	return &CacheManager[T]{
		items: make(map[string]*CacheItem[T]),
		name:  name,
		desc:  desc,
	}
}

// Set 设置缓存项
func (c *CacheManager[T]) Set(key string, value T, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	} else {
		// duration为0表示不过期，设置为一个很大的值
		expiration = time.Now().AddDate(100, 0, 0).UnixNano()
	}
	
	c.items[key] = &CacheItem[T]{
		Value:      value,
		Expiration: expiration,
	}
}

// Get 获取缓存项
func (c *CacheManager[T]) Get(key string) (T, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	item, exists := c.items[key]
	if !exists {
		var zero T
		return zero, false
	}
	
	// 检查是否过期，如果过期则删除
	if time.Now().UnixNano() > item.Expiration {
		delete(c.items, key)
		var zero T
		return zero, false
	}
	
	return item.Value, true
}

// Delete 删除缓存项
func (c *CacheManager[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
}

// Clear 清空所有缓存
func (c *CacheManager[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = make(map[string]*CacheItem[T])
}

// GetName 获取缓存管理器名称
func (c *CacheManager[T]) GetName() string {
	return c.name
}

// GetDesc 获取缓存管理器描述
func (c *CacheManager[T]) GetDesc() string {
	return c.desc
}

// CleanExpired 清理过期缓存
func (c *CacheManager[T]) CleanExpired() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	cleaned := 0
	now := time.Now().UnixNano()
	for key, item := range c.items {
		if now > item.Expiration {
			delete(c.items, key)
			cleaned++
		}
	}
	return cleaned
}

// Size 获取缓存项数量
func (c *CacheManager[T]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// Count 获取缓存项数量（Size的别名）
func (c *CacheManager[T]) Count() int {
	return c.Size()
}

// Exists 检查键是否存在
func (c *CacheManager[T]) Exists(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	item, exists := c.items[key]
	if !exists {
		return false
	}
	
	// 检查是否过期，如果过期则删除
	if time.Now().UnixNano() > item.Expiration {
		delete(c.items, key)
		return false
	}
	
	return true
}

// Keys 获取所有键
func (c *CacheManager[T]) Keys() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	keys := make([]string, 0, len(c.items))
	now := time.Now().UnixNano()
	
	// 清理过期项并收集有效键
	for key, item := range c.items {
		if now <= item.Expiration {
			keys = append(keys, key)
		} else {
			delete(c.items, key)
		}
	}
	
	return keys
}

// GetOrSet 获取或设置缓存项
func (c *CacheManager[T]) GetOrSet(key string, value T, duration time.Duration) T {
	if val, found := c.Get(key); found {
		return val
	}
	
	c.Set(key, value, duration)
	return value
}

// GetWithTTL 获取缓存项及其剩余TTL
func (c *CacheManager[T]) GetWithTTL(key string) (T, time.Duration, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	item, exists := c.items[key]
	if !exists {
		var zero T
		return zero, -1, false
	}
	
	now := time.Now().UnixNano()
	if now > item.Expiration {
		delete(c.items, key)
		var zero T
		return zero, -1, false
	}
	
	// 检查是否为永不过期（100年后的时间表示永不过期）
	century := time.Now().AddDate(50, 0, 0).UnixNano()
	if item.Expiration > century {
		return item.Value, -1, true
	}
	
	ttl := time.Duration(item.Expiration - now)
	return item.Value, ttl, true
}

// CleanupExpired 清理过期缓存的别名
func (c *CacheManager[T]) CleanupExpired() int {
	return c.CleanExpired()
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Keys         int    `json:"keys"`
	Size         int    `json:"size"`
	ItemCount    int    `json:"itemCount"`
	Name         string `json:"name"`
	Desc         string `json:"desc"`
	Description  string `json:"description"`
}

// GetStats 获取缓存统计信息
func (c *CacheManager[T]) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	// 计算有效（未过期）的项目数
	validCount := 0
	now := time.Now().UnixNano()
	for _, item := range c.items {
		if now <= item.Expiration {
			validCount++
		}
	}
	
	return CacheStats{
		Keys:        validCount,
		Size:        validCount,
		ItemCount:   validCount,
		Name:        c.name,
		Desc:        c.desc,
		Description: c.desc,
	}
}

// 全局缓存实例
var (
	DictCache     *CacheManager[[]map[string]any] // 字典缓存
	ConfigCache   *CacheManager[string]           // 配置缓存
	EmailCache    *CacheManager[map[string]any]   // 邮件配置缓存
	GeneralCache  *CacheManager[any]              // 通用缓存
)

// InitCaches 初始化所有缓存管理器
func InitCaches() {
	DictCache = NewCacheManager[[]map[string]any]("DictCache", "字典缓存")
	ConfigCache = NewCacheManager[string]("ConfigCache", "配置缓存")
	EmailCache = NewCacheManager[map[string]any]("EmailCache", "邮件配置缓存")
	GeneralCache = NewCacheManager[any]("GeneralCache", "通用缓存")
}

func init() {
	InitCaches()
	
	// 启动定时清理过期缓存的goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			DictCache.CleanExpired()
			ConfigCache.CleanExpired()
			EmailCache.CleanExpired()
			GeneralCache.CleanExpired()
		}
	}()
}