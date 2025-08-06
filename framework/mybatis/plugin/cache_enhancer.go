// Package plugin 缓存增强插件实现
//
// 提供高级缓存功能，包括缓存预热、缓存统计、缓存失效策略等
package plugin

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// CacheEnhancerPlugin 缓存增强插件
type CacheEnhancerPlugin struct {
	*BasePlugin
	cacheManager     *CacheManager
	enableStatistics bool
	enablePreload    bool
	statistics       *CacheStatistics
}

// CacheManager 缓存管理器
type CacheManager struct {
	caches map[string]*EnhancedCache
	mutex  sync.RWMutex
}

// EnhancedCache 增强缓存
type EnhancedCache struct {
	name         string
	data         map[string]*CacheEntry
	maxSize      int
	ttl          time.Duration
	hitCount     int64
	missCount    int64
	evictCount   int64
	loadCount    int64
	mutex        sync.RWMutex
	lastAccess   time.Time
	preloadRules []PreloadRule
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Key        string
	Value      any
	CreateTime time.Time
	AccessTime time.Time
	HitCount   int64
	Size       int64
	TTL        time.Duration
}

// PreloadRule 预加载规则
type PreloadRule struct {
	Pattern    string        // SQL模式
	Parameters []any         // 预加载参数
	Schedule   time.Duration // 预加载间隔
	LastLoad   time.Time     // 上次加载时间
}

// CacheStatistics 缓存统计
type CacheStatistics struct {
	TotalHits   int64
	TotalMisses int64
	TotalEvicts int64
	TotalLoads  int64
	HitRate     float64
	MissRate    float64
	AvgLoadTime time.Duration
	CacheCount  int
	TotalSize   int64
	mutex       sync.RWMutex
}

// NewCacheEnhancerPlugin 创建缓存增强插件
func NewCacheEnhancerPlugin() *CacheEnhancerPlugin {
	plugin := &CacheEnhancerPlugin{
		BasePlugin:       NewBasePlugin("cache_enhancer", 4),
		cacheManager:     NewCacheManager(),
		enableStatistics: true,
		enablePreload:    false,
		statistics:       NewCacheStatistics(),
	}
	return plugin
}

// NewCacheManager 创建缓存管理器
func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches: make(map[string]*EnhancedCache),
	}
}

// NewEnhancedCache 创建增强缓存
func NewEnhancedCache(name string, maxSize int, ttl time.Duration) *EnhancedCache {
	return &EnhancedCache{
		name:         name,
		data:         make(map[string]*CacheEntry),
		maxSize:      maxSize,
		ttl:          ttl,
		lastAccess:   time.Now(),
		preloadRules: make([]PreloadRule, 0),
	}
}

// NewCacheStatistics 创建缓存统计
func NewCacheStatistics() *CacheStatistics {
	return &CacheStatistics{
		TotalHits:   0,
		TotalMisses: 0,
		TotalEvicts: 0,
		TotalLoads:  0,
		HitRate:     0.0,
		MissRate:    0.0,
		AvgLoadTime: 0,
		CacheCount:  0,
		TotalSize:   0,
	}
}

// Intercept 拦截方法调用
func (plugin *CacheEnhancerPlugin) Intercept(invocation *Invocation) (any, error) {
	// 生成缓存键
	cacheKey := plugin.generateCacheKey(invocation)

	// 尝试从缓存获取
	if cached := plugin.getFromCache(cacheKey); cached != nil {
		plugin.updateHitStatistics()
		return cached, nil
	}

	// 缓存未命中，执行原方法
	plugin.updateMissStatistics()
	startTime := time.Now()

	result, err := invocation.Proceed()

	loadTime := time.Since(startTime)
	plugin.updateLoadStatistics(loadTime)

	// 如果执行成功，将结果放入缓存
	if err == nil && result != nil {
		plugin.putToCache(cacheKey, result)
	}

	return result, err
}

// Plugin 包装目标对象
func (plugin *CacheEnhancerPlugin) Plugin(target any) any {
	return target
}

// SetProperties 设置插件属性
func (plugin *CacheEnhancerPlugin) SetProperties(properties map[string]any) {
	plugin.BasePlugin.SetProperties(properties)

	plugin.enableStatistics = plugin.GetPropertyBool("enableStatistics", true)
	plugin.enablePreload = plugin.GetPropertyBool("enablePreload", false)
}

// generateCacheKey 生成缓存键
func (plugin *CacheEnhancerPlugin) generateCacheKey(invocation *Invocation) string {
	// 使用方法名和参数生成缓存键
	keyData := fmt.Sprintf("%s:%v", invocation.Method.Name, invocation.Args)

	// 使用MD5生成固定长度的键
	hash := md5.Sum([]byte(keyData))
	return fmt.Sprintf("%x", hash)
}

// getFromCache 从缓存获取数据
func (plugin *CacheEnhancerPlugin) getFromCache(key string) any {
	cache := plugin.cacheManager.GetDefaultCache()
	if cache == nil {
		return nil
	}

	return cache.Get(key)
}

// putToCache 将数据放入缓存
func (plugin *CacheEnhancerPlugin) putToCache(key string, value any) {
	cache := plugin.cacheManager.GetDefaultCache()
	if cache == nil {
		cache = plugin.cacheManager.CreateCache("default", 1000, time.Hour)
	}

	cache.Put(key, value)
}

// updateHitStatistics 更新命中统计
func (plugin *CacheEnhancerPlugin) updateHitStatistics() {
	if !plugin.enableStatistics {
		return
	}

	plugin.statistics.mutex.Lock()
	defer plugin.statistics.mutex.Unlock()

	plugin.statistics.TotalHits++
	plugin.updateRates()
}

// updateMissStatistics 更新未命中统计
func (plugin *CacheEnhancerPlugin) updateMissStatistics() {
	if !plugin.enableStatistics {
		return
	}

	plugin.statistics.mutex.Lock()
	defer plugin.statistics.mutex.Unlock()

	plugin.statistics.TotalMisses++
	plugin.updateRates()
}

// updateLoadStatistics 更新加载统计
func (plugin *CacheEnhancerPlugin) updateLoadStatistics(loadTime time.Duration) {
	if !plugin.enableStatistics {
		return
	}

	plugin.statistics.mutex.Lock()
	defer plugin.statistics.mutex.Unlock()

	plugin.statistics.TotalLoads++

	// 计算平均加载时间
	if plugin.statistics.TotalLoads > 0 {
		totalTime := plugin.statistics.AvgLoadTime*time.Duration(plugin.statistics.TotalLoads-1) + loadTime
		plugin.statistics.AvgLoadTime = totalTime / time.Duration(plugin.statistics.TotalLoads)
	}
}

// updateRates 更新命中率和未命中率
func (plugin *CacheEnhancerPlugin) updateRates() {
	total := plugin.statistics.TotalHits + plugin.statistics.TotalMisses
	if total > 0 {
		plugin.statistics.HitRate = float64(plugin.statistics.TotalHits) / float64(total)
		plugin.statistics.MissRate = float64(plugin.statistics.TotalMisses) / float64(total)
	}
}

// GetCacheReport 获取缓存报告
func (plugin *CacheEnhancerPlugin) GetCacheReport() map[string]any {
	stats := plugin.GetCacheStatistics()

	report := map[string]any{
		"总命中数":   stats.TotalHits,
		"总未命中数":  stats.TotalMisses,
		"总驱逐数":   stats.TotalEvicts,
		"总加载数":   stats.TotalLoads,
		"命中率":    fmt.Sprintf("%.2f%%", stats.HitRate*100),
		"未命中率":   fmt.Sprintf("%.2f%%", stats.MissRate*100),
		"平均加载时间": stats.AvgLoadTime.String(),
		"缓存数量":   stats.CacheCount,
		"总缓存大小":  stats.TotalSize,
	}

	return report
}

// GetCacheStatistics 获取插件缓存统计
func (plugin *CacheEnhancerPlugin) GetCacheStatistics() *CacheStatistics {
	plugin.statistics.mutex.RLock()
	defer plugin.statistics.mutex.RUnlock()

	// 更新缓存数量和总大小
	plugin.statistics.CacheCount = len(plugin.cacheManager.caches)

	var totalSize int64
	for _, cache := range plugin.cacheManager.caches {
		totalSize += int64(cache.Size())
	}
	plugin.statistics.TotalSize = totalSize

	// 返回副本
	return &CacheStatistics{
		TotalHits:   plugin.statistics.TotalHits,
		TotalMisses: plugin.statistics.TotalMisses,
		TotalEvicts: plugin.statistics.TotalEvicts,
		TotalLoads:  plugin.statistics.TotalLoads,
		HitRate:     plugin.statistics.HitRate,
		MissRate:    plugin.statistics.MissRate,
		AvgLoadTime: plugin.statistics.AvgLoadTime,
		CacheCount:  plugin.statistics.CacheCount,
		TotalSize:   plugin.statistics.TotalSize,
	}
}

// CacheManager 方法实现

// CreateCache 创建缓存
func (manager *CacheManager) CreateCache(name string, maxSize int, ttl time.Duration) *EnhancedCache {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	cache := NewEnhancedCache(name, maxSize, ttl)
	manager.caches[name] = cache
	return cache
}

// GetCache 获取缓存
func (manager *CacheManager) GetCache(name string) *EnhancedCache {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	return manager.caches[name]
}

// GetDefaultCache 获取默认缓存
func (manager *CacheManager) GetDefaultCache() *EnhancedCache {
	return manager.GetCache("default")
}

// EnhancedCache 方法实现

// Get 获取缓存项
func (cache *EnhancedCache) Get(key string) any {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	entry, exists := cache.data[key]
	if !exists {
		return nil
	}

	// 检查是否过期
	if cache.isExpired(entry) {
		delete(cache.data, key)
		return nil
	}

	// 更新访问时间和命中次数
	entry.AccessTime = time.Now()
	entry.HitCount++
	cache.hitCount++
	cache.lastAccess = time.Now()

	return entry.Value
}

// Put 放入缓存项
func (cache *EnhancedCache) Put(key string, value any) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// 创建新的缓存项
	entry := &CacheEntry{
		Key:        key,
		Value:      value,
		CreateTime: time.Now(),
		AccessTime: time.Now(),
		HitCount:   0,
		Size:       1, // 简化实现
		TTL:        cache.ttl,
	}

	cache.data[key] = entry
	cache.loadCount++
}

// Size 获取缓存大小
func (cache *EnhancedCache) Size() int {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	return len(cache.data)
}

// isExpired 检查是否过期
func (cache *EnhancedCache) isExpired(entry *CacheEntry) bool {
	if cache.ttl <= 0 {
		return false // 永不过期
	}

	return time.Since(entry.CreateTime) > cache.ttl
}
