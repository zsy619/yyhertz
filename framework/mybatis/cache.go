// Package mybatis 缓存系统实现
//
// 提供MyBatis风格的一级和二级缓存机制
package mybatis

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(key string) (any, bool)
	// Set 设置缓存值
	Set(key string, value any, ttl time.Duration)
	// Delete 删除缓存
	Delete(key string)
	// Clear 清空缓存
	Clear()
	// Size 缓存大小
	Size() int
}

// CacheEntry 缓存项
type CacheEntry struct {
	Value     any       `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	TTL       time.Duration `json:"ttl"`
}

// IsExpired 检查是否过期
func (e *CacheEntry) IsExpired() bool {
	if e.TTL <= 0 {
		return false // 永不过期
	}
	return time.Since(e.CreatedAt) > e.TTL
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data      map[string]*CacheEntry
	mutex     sync.RWMutex
	maxSize   int
	cleanupInterval time.Duration
	stopCleanup     chan bool
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(maxSize int, cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data:            make(map[string]*CacheEntry),
		maxSize:         maxSize,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan bool),
	}
	
	// 启动清理协程
	go cache.cleanup()
	
	return cache
}

// Get 获取缓存值
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}
	
	if entry.IsExpired() {
		// 过期了，异步删除
		go func() {
			c.mutex.Lock()
			delete(c.data, key)
			c.mutex.Unlock()
		}()
		return nil, false
	}
	
	return entry.Value, true
}

// Set 设置缓存值
func (c *MemoryCache) Set(key string, value any, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// 如果达到最大容量，删除一些旧项目
	if len(c.data) >= c.maxSize {
		c.evictOldEntries()
	}
	
	c.data[key] = &CacheEntry{
		Value:     value,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]*CacheEntry)
}

// Size 缓存大小
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}

// evictOldEntries 淘汰旧项目
func (c *MemoryCache) evictOldEntries() {
	// 简单策略：删除最旧的25%项目
	toDelete := len(c.data) / 4
	if toDelete == 0 {
		toDelete = 1
	}
	
	oldestKeys := make([]string, 0, toDelete)
	oldestTime := time.Now()
	
	for key, entry := range c.data {
		if entry.CreatedAt.Before(oldestTime) || len(oldestKeys) < toDelete {
			if len(oldestKeys) == toDelete {
				// 替换最新的
				for i, k := range oldestKeys {
					if c.data[k].CreatedAt.After(entry.CreatedAt) {
						oldestKeys[i] = key
						break
					}
				}
			} else {
				oldestKeys = append(oldestKeys, key)
			}
		}
	}
	
	for _, key := range oldestKeys {
		delete(c.data, key)
	}
}

// cleanup 清理过期项目
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanupExpired 清理过期项目
func (c *MemoryCache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	expiredKeys := make([]string, 0)
	for key, entry := range c.data {
		if entry.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	for _, key := range expiredKeys {
		delete(c.data, key)
	}
}

// Stop 停止缓存（停止清理协程）
func (c *MemoryCache) Stop() {
	close(c.stopCleanup)
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// 一级缓存配置
	L1CacheEnabled  bool          `json:"l1CacheEnabled"`
	L1CacheSize     int           `json:"l1CacheSize"`
	L1CacheTTL      time.Duration `json:"l1CacheTtl"`
	
	// 二级缓存配置
	L2CacheEnabled  bool          `json:"l2CacheEnabled"`
	L2CacheSize     int           `json:"l2CacheSize"`
	L2CacheTTL      time.Duration `json:"l2CacheTtl"`
	
	// 清理间隔
	CleanupInterval time.Duration `json:"cleanupInterval"`
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		L1CacheEnabled:  true,
		L1CacheSize:     1000,
		L1CacheTTL:      10 * time.Minute,
		L2CacheEnabled:  false,
		L2CacheSize:     5000,
		L2CacheTTL:      30 * time.Minute,
		CleanupInterval: 5 * time.Minute,
	}
}

// CacheManager 缓存管理器
type CacheManager struct {
	l1Cache      Cache         // 一级缓存（会话级别）
	l2Cache      Cache         // 二级缓存（应用级别）
	config       *CacheConfig
	statementCache map[string]bool // 语句缓存配置映射
	mutex        sync.RWMutex
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(config *CacheConfig) *CacheManager {
	var l1Cache, l2Cache Cache
	
	if config.L1CacheEnabled {
		l1Cache = NewMemoryCache(config.L1CacheSize, config.CleanupInterval)
	}
	
	if config.L2CacheEnabled {
		l2Cache = NewMemoryCache(config.L2CacheSize, config.CleanupInterval)
	}
	
	return &CacheManager{
		l1Cache:        l1Cache,
		l2Cache:        l2Cache,
		config:         config,
		statementCache: make(map[string]bool),
	}
}

// GenerateCacheKey 生成缓存键
func (cm *CacheManager) GenerateCacheKey(sql string, args []any) string {
	// 创建键的原始数据
	keyData := struct {
		SQL  string `json:"sql"`
		Args []any  `json:"args"`
	}{
		SQL:  sql,
		Args: args,
	}
	
	// 序列化为JSON
	jsonData, err := json.Marshal(keyData)
	if err != nil {
		// 如果序列化失败，使用简单的字符串拼接
		return fmt.Sprintf("%s:%v", sql, args)
	}
	
	// 生成MD5哈希
	hash := md5.Sum(jsonData)
	return hex.EncodeToString(hash[:])
}

// Get 获取缓存值
func (cm *CacheManager) Get(sql string, args []any) (any, bool) {
	key := cm.GenerateCacheKey(sql, args)
	
	// 先查一级缓存
	if cm.l1Cache != nil {
		if value, found := cm.l1Cache.Get(key); found {
			return value, true
		}
	}
	
	// 再查二级缓存
	if cm.l2Cache != nil {
		if value, found := cm.l2Cache.Get(key); found {
			// 将二级缓存结果放入一级缓存
			if cm.l1Cache != nil {
				cm.l1Cache.Set(key, value, cm.config.L1CacheTTL)
			}
			return value, true
		}
	}
	
	return nil, false
}

// Set 设置缓存值
func (cm *CacheManager) Set(sql string, args []any, value any) {
	key := cm.GenerateCacheKey(sql, args)
	
	// 设置一级缓存
	if cm.l1Cache != nil {
		cm.l1Cache.Set(key, value, cm.config.L1CacheTTL)
	}
	
	// 设置二级缓存
	if cm.l2Cache != nil {
		cm.l2Cache.Set(key, value, cm.config.L2CacheTTL)
	}
}

// ClearL1Cache 清空一级缓存
func (cm *CacheManager) ClearL1Cache() {
	if cm.l1Cache != nil {
		cm.l1Cache.Clear()
	}
}

// ClearL2Cache 清空二级缓存
func (cm *CacheManager) ClearL2Cache() {
	if cm.l2Cache != nil {
		cm.l2Cache.Clear()
	}
}

// ClearAll 清空所有缓存
func (cm *CacheManager) ClearAll() {
	cm.ClearL1Cache()
	cm.ClearL2Cache()
}

// SetStatementCacheEnabled 设置语句缓存开关
func (cm *CacheManager) SetStatementCacheEnabled(statementId string, enabled bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.statementCache[statementId] = enabled
}

// IsStatementCacheEnabled 检查语句是否启用缓存
func (cm *CacheManager) IsStatementCacheEnabled(statementId string) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	enabled, exists := cm.statementCache[statementId]
	return exists && enabled
}

// GetStats 获取缓存统计
func (cm *CacheManager) GetStats() map[string]any {
	stats := make(map[string]any)
	
	if cm.l1Cache != nil {
		stats["l1_cache_size"] = cm.l1Cache.Size()
		stats["l1_cache_enabled"] = true
	} else {
		stats["l1_cache_enabled"] = false
	}
	
	if cm.l2Cache != nil {
		stats["l2_cache_size"] = cm.l2Cache.Size()
		stats["l2_cache_enabled"] = true
	} else {
		stats["l2_cache_enabled"] = false
	}
	
	return stats
}

// Stop 停止缓存管理器
func (cm *CacheManager) Stop() {
	if l1Memory, ok := cm.l1Cache.(*MemoryCache); ok {
		l1Memory.Stop()
	}
	if l2Memory, ok := cm.l2Cache.(*MemoryCache); ok {
		l2Memory.Stop()
	}
}