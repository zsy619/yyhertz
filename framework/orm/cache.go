// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheProvider 缓存提供者接口
type CacheProvider interface {
	// Get 获取缓存
	Get(key string, value interface{}) (bool, error)
	// Set 设置缓存
	Set(key string, value interface{}, expiration time.Duration) error
	// Delete 删除缓存
	Delete(key string) error
	// Clear 清空缓存
	Clear() error
	// GetMulti 批量获取缓存
	GetMulti(keys []string) (map[string][]byte, error)
	// SetMulti 批量设置缓存
	SetMulti(items map[string]interface{}, expiration time.Duration) error
	// DeleteMulti 批量删除缓存
	DeleteMulti(keys []string) error
	// Close 关闭缓存
	Close() error
}

// MemoryCacheProvider 内存缓存提供者
type MemoryCacheProvider struct {
	items     map[string]*memoryCacheItem
	mutex     sync.RWMutex
	janitor   *time.Ticker
	stopChan  chan struct{}
	maxItems  int
	evictType string // LRU, FIFO, Random
}

// memoryCacheItem 内存缓存项
type memoryCacheItem struct {
	value      []byte
	expiration int64
	created    time.Time
	lastAccess time.Time
}

// NewMemoryCacheProvider 创建内存缓存提供者
func NewMemoryCacheProvider(cleanupInterval time.Duration, maxItems int, evictType string) *MemoryCacheProvider {
	if maxItems <= 0 {
		maxItems = 10000 // 默认最大缓存项数
	}

	if evictType == "" {
		evictType = "LRU" // 默认使用LRU淘汰策略
	}

	provider := &MemoryCacheProvider{
		items:     make(map[string]*memoryCacheItem),
		maxItems:  maxItems,
		evictType: evictType,
		stopChan:  make(chan struct{}),
	}

	// 启动清理过期项的定时器
	if cleanupInterval > 0 {
		provider.janitor = time.NewTicker(cleanupInterval)
		go provider.janitorRun()
	}

	return provider
}

// janitorRun 运行清理过期项的定时任务
func (p *MemoryCacheProvider) janitorRun() {
	for {
		select {
		case <-p.janitor.C:
			p.deleteExpired()
		case <-p.stopChan:
			p.janitor.Stop()
			return
		}
	}
}

// deleteExpired 删除过期项
func (p *MemoryCacheProvider) deleteExpired() {
	now := time.Now().UnixNano()
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for key, item := range p.items {
		if item.expiration > 0 && item.expiration < now {
			delete(p.items, key)
		}
	}
}

// Get 获取缓存
func (p *MemoryCacheProvider) Get(key string, value interface{}) (bool, error) {
	p.mutex.RLock()
	item, found := p.items[key]
	if !found {
		p.mutex.RUnlock()
		return false, nil
	}

	// 检查是否过期
	if item.expiration > 0 && item.expiration < time.Now().UnixNano() {
		p.mutex.RUnlock()
		// 异步删除过期项
		go func() {
			p.mutex.Lock()
			delete(p.items, key)
			p.mutex.Unlock()
		}()
		return false, nil
	}

	// 更新最后访问时间（用于LRU）
	item.lastAccess = time.Now()
	p.mutex.RUnlock()

	// 反序列化
	return true, json.Unmarshal(item.value, value)
}

// Set 设置缓存
func (p *MemoryCacheProvider) Set(key string, value interface{}, expiration time.Duration) error {
	// 序列化值
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 计算过期时间
	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 检查是否需要淘汰
	if len(p.items) >= p.maxItems {
		p.evict()
	}

	p.items[key] = &memoryCacheItem{
		value:      b,
		expiration: exp,
		created:    time.Now(),
		lastAccess: time.Now(),
	}

	return nil
}

// evict 淘汰缓存项
func (p *MemoryCacheProvider) evict() {
	if len(p.items) == 0 {
		return
	}

	switch p.evictType {
	case "LRU":
		// 淘汰最久未使用的项
		var oldestKey string
		var oldestAccess time.Time

		first := true
		for k, item := range p.items {
			if first || item.lastAccess.Before(oldestAccess) {
				oldestKey = k
				oldestAccess = item.lastAccess
				first = false
			}
		}

		if oldestKey != "" {
			delete(p.items, oldestKey)
		}

	case "FIFO":
		// 淘汰最先创建的项
		var oldestKey string
		var oldestCreated time.Time

		first := true
		for k, item := range p.items {
			if first || item.created.Before(oldestCreated) {
				oldestKey = k
				oldestCreated = item.created
				first = false
			}
		}

		if oldestKey != "" {
			delete(p.items, oldestKey)
		}

	default:
		// 随机淘汰
		for k := range p.items {
			delete(p.items, k)
			break
		}
	}
}

// Delete 删除缓存
func (p *MemoryCacheProvider) Delete(key string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.items, key)
	return nil
}

// Clear 清空缓存
func (p *MemoryCacheProvider) Clear() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.items = make(map[string]*memoryCacheItem)
	return nil
}

// GetMulti 批量获取缓存
func (p *MemoryCacheProvider) GetMulti(keys []string) (map[string][]byte, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	result := make(map[string][]byte, len(keys))
	now := time.Now().UnixNano()

	for _, key := range keys {
		item, found := p.items[key]
		if !found {
			continue
		}

		// 检查是否过期
		if item.expiration > 0 && item.expiration < now {
			continue
		}

		// 更新最后访问时间
		item.lastAccess = time.Now()
		result[key] = item.value
	}

	return result, nil
}

// SetMulti 批量设置缓存
func (p *MemoryCacheProvider) SetMulti(items map[string]interface{}, expiration time.Duration) error {
	// 计算过期时间
	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for key, value := range items {
		// 序列化值
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}

		// 检查是否需要淘汰
		if len(p.items) >= p.maxItems {
			p.evict()
		}

		p.items[key] = &memoryCacheItem{
			value:      b,
			expiration: exp,
			created:    time.Now(),
			lastAccess: time.Now(),
		}
	}

	return nil
}

// DeleteMulti 批量删除缓存
func (p *MemoryCacheProvider) DeleteMulti(keys []string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, key := range keys {
		delete(p.items, key)
	}

	return nil
}

// Close 关闭缓存提供者
func (p *MemoryCacheProvider) Close() error {
	if p.janitor != nil {
		close(p.stopChan)
	}
	return nil
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// 是否启用缓存
	Enabled bool `json:"enabled" yaml:"enabled"`
	// 缓存提供者类型: memory, redis
	Provider string `json:"provider" yaml:"provider"`
	// 默认过期时间
	DefaultExpiration time.Duration `json:"default_expiration" yaml:"default_expiration"`
	// 清理间隔
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
	// 最大缓存项数
	MaxItems int `json:"max_items" yaml:"max_items"`
	// 淘汰策略: LRU, FIFO, Random
	EvictionPolicy string `json:"eviction_policy" yaml:"eviction_policy"`
	// 缓存命名空间
	Namespace string `json:"namespace" yaml:"namespace"`
	// 缓存统计
	EnableStats bool `json:"enable_stats" yaml:"enable_stats"`
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled:           true,
		Provider:          "memory",
		DefaultExpiration: time.Minute * 5,
		CleanupInterval:   time.Minute,
		MaxItems:          10000,
		EvictionPolicy:    "LRU",
		Namespace:         "orm",
		EnableStats:       true,
	}
}

// CacheManager 缓存管理器
type CacheManager struct {
	provider CacheProvider
	config   *CacheConfig
	stats    *CacheStats
	mutex    sync.RWMutex
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits              int64         `json:"hits"`
	Misses            int64         `json:"misses"`
	Sets              int64         `json:"sets"`
	Deletes           int64         `json:"deletes"`
	Evictions         int64         `json:"evictions"`
	TotalItems        int64         `json:"total_items"`
	AverageAccessTime time.Duration `json:"average_access_time"`
	HitRate           float64       `json:"hit_rate"`
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(config *CacheConfig) (*CacheManager, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	var provider CacheProvider

	switch config.Provider {
	case "memory":
		provider = NewMemoryCacheProvider(config.CleanupInterval, config.MaxItems, config.EvictionPolicy)
	// 可以在这里添加其他缓存提供者，如Redis等
	default:
		return nil, fmt.Errorf("不支持的缓存提供者: %s", config.Provider)
	}

	return &CacheManager{
		provider: provider,
		config:   config,
		stats:    &CacheStats{},
	}, nil
}

// Get 获取缓存
func (cm *CacheManager) Get(key string, value interface{}) (bool, error) {
	startTime := time.Now()
	cacheKey := cm.buildKey(key)

	found, err := cm.provider.Get(cacheKey, value)

	// 更新统计
	if cm.config.EnableStats {
		cm.mutex.Lock()
		if found {
			cm.stats.Hits++
		} else {
			cm.stats.Misses++
		}
		cm.updateAccessTime(time.Since(startTime))
		cm.mutex.Unlock()
	}

	return found, err
}

// Set 设置缓存
func (cm *CacheManager) Set(key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = cm.config.DefaultExpiration
	}

	cacheKey := cm.buildKey(key)
	err := cm.provider.Set(cacheKey, value, expiration)

	// 更新统计
	if cm.config.EnableStats {
		cm.mutex.Lock()
		cm.stats.Sets++
		cm.stats.TotalItems++
		cm.mutex.Unlock()
	}

	return err
}

// Delete 删除缓存
func (cm *CacheManager) Delete(key string) error {
	cacheKey := cm.buildKey(key)
	err := cm.provider.Delete(cacheKey)

	// 更新统计
	if cm.config.EnableStats {
		cm.mutex.Lock()
		cm.stats.Deletes++
		cm.stats.TotalItems--
		cm.mutex.Unlock()
	}

	return err
}

// Clear 清空缓存
func (cm *CacheManager) Clear() error {
	err := cm.provider.Clear()

	// 更新统计
	if cm.config.EnableStats {
		cm.mutex.Lock()
		cm.stats.TotalItems = 0
		cm.mutex.Unlock()
	}

	return err
}

// GetMulti 批量获取缓存
func (cm *CacheManager) GetMulti(keys []string) (map[string][]byte, error) {
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = cm.buildKey(key)
	}

	return cm.provider.GetMulti(cacheKeys)
}

// SetMulti 批量设置缓存
func (cm *CacheManager) SetMulti(items map[string]interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = cm.config.DefaultExpiration
	}

	cacheItems := make(map[string]interface{})
	for key, value := range items {
		cacheItems[cm.buildKey(key)] = value
	}

	// 更新统计
	if cm.config.EnableStats {
		cm.mutex.Lock()
		cm.stats.Sets += int64(len(items))
		cm.stats.TotalItems += int64(len(items))
		cm.mutex.Unlock()
	}

	return cm.provider.SetMulti(cacheItems, expiration)
}

// DeleteMulti 批量删除缓存
func (cm *CacheManager) DeleteMulti(keys []string) error {
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = cm.buildKey(key)
	}

	// 更新统计
	if cm.config.EnableStats {
		cm.mutex.Lock()
		cm.stats.Deletes += int64(len(keys))
		cm.stats.TotalItems -= int64(len(keys))
		cm.mutex.Unlock()
	}

	return cm.provider.DeleteMulti(cacheKeys)
}

// buildKey 构建缓存键
func (cm *CacheManager) buildKey(key string) string {
	if cm.config.Namespace != "" {
		return cm.config.Namespace + ":" + key
	}
	return key
}

// updateAccessTime 更新访问时间统计
func (cm *CacheManager) updateAccessTime(duration time.Duration) {
	total := cm.stats.Hits + cm.stats.Misses
	if total > 0 {
		cm.stats.AverageAccessTime = time.Duration((int64(cm.stats.AverageAccessTime)*(total-1) + int64(duration)) / total)
		cm.stats.HitRate = float64(cm.stats.Hits) / float64(total)
	}
}

// GetStats 获取缓存统计
func (cm *CacheManager) GetStats() *CacheStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 创建副本
	stats := *cm.stats
	return &stats
}

// Close 关闭缓存管理器
func (cm *CacheManager) Close() error {
	return cm.provider.Close()
}

// 全局缓存管理器
var (
	globalCacheManager *CacheManager
	cacheManagerOnce   sync.Once
)

// GetGlobalCacheManager 获取全局缓存管理器
func GetGlobalCacheManager() *CacheManager {
	cacheManagerOnce.Do(func() {
		config := DefaultCacheConfig()

		// 尝试从全局配置获取
		// TODO: 从配置文件中读取缓存配置

		// 直接创建缓存管理器，忽略可能的错误
		globalCacheManager, _ = NewCacheManager(config)
		if globalCacheManager == nil {
			// 如果创建失败，使用默认配置
			fmt.Printf("初始化全局缓存管理器失败，将使用默认配置\n")
			// 创建一个默认的内存缓存提供者
			globalCacheManager = &CacheManager{
				provider: NewMemoryCacheProvider(time.Minute, 10000, "LRU"),
				config:   config,
				stats:    &CacheStats{},
			}
		}
	})

	return globalCacheManager
}

// SetGlobalCacheManager 设置全局缓存管理器
func SetGlobalCacheManager(cm *CacheManager) {
	if globalCacheManager != nil {
		globalCacheManager.Close()
	}

	globalCacheManager = cm
}

// ============= 便捷函数 =============

// GetCache 获取缓存
func GetCache(key string, value interface{}) (bool, error) {
	return GetGlobalCacheManager().Get(key, value)
}

// SetCache 设置缓存
func SetCache(key string, value interface{}, expiration time.Duration) error {
	return GetGlobalCacheManager().Set(key, value, expiration)
}

// DeleteCache 删除缓存
func DeleteCache(key string) error {
	return GetGlobalCacheManager().Delete(key)
}

// ClearCache 清空缓存
func ClearCache() error {
	return GetGlobalCacheManager().Clear()
}
