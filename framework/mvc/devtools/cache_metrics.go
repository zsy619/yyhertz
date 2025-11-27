// Package devtools 提供缓存监控指标功能
//
// 缓存监控模块用于监控Redis和内存缓存的性能指标，提供：
// - 缓存命中率统计
// - Redis连接池监控
// - 内存缓存使用分析
// - 缓存操作延迟监控
// - 键过期和淘汰统计
// - 缓存热点分析
//
// 功能特性：
// - 支持多种缓存类型（Redis、内存缓存）
// - 实时命中率计算
// - 缓存操作性能分析
// - 自动热点键识别
// - 内存使用优化建议
// - 缓存模式识别
package devtools

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/config"
)

// CacheType 缓存类型枚举
type CacheType string

const (
	CacheTypeRedis  CacheType = "redis"
	CacheTypeMemory CacheType = "memory"
	CacheTypeLocal  CacheType = "local"
)

// CacheOperation 缓存操作枚举
type CacheOperation string

const (
	OperationGet    CacheOperation = "GET"
	OperationSet    CacheOperation = "SET"
	OperationDelete CacheOperation = "DELETE"
	OperationExists CacheOperation = "EXISTS"
	OperationExpire CacheOperation = "EXPIRE"
	OperationIncr   CacheOperation = "INCR"
	OperationDecr   CacheOperation = "DECR"
)

// CacheHitStats 缓存命中统计
type CacheHitStats struct {
	TotalRequests int64   `json:"total_requests"` // 总请求数
	HitCount      int64   `json:"hit_count"`      // 命中次数
	MissCount     int64   `json:"miss_count"`     // 未命中次数
	HitRatio      float64 `json:"hit_ratio"`      // 命中率
	MissRatio     float64 `json:"miss_ratio"`     // 未命中率
}

// CacheOperationStats 缓存操作统计
type CacheOperationStats struct {
	Operation     CacheOperation `json:"operation"`      // 操作类型
	Count         int64          `json:"count"`          // 操作次数
	TotalDuration time.Duration  `json:"total_duration"` // 总耗时
	AvgDuration   time.Duration  `json:"avg_duration"`   // 平均耗时
	MinDuration   time.Duration  `json:"min_duration"`   // 最小耗时
	MaxDuration   time.Duration  `json:"max_duration"`   // 最大耗时
	ErrorCount    int64          `json:"error_count"`    // 错误次数
	SuccessRatio  float64        `json:"success_ratio"`  // 成功率
}

// CacheKeyStats 缓存键统计
type CacheKeyStats struct {
	Key           string        `json:"key"`            // 键名
	AccessCount   int64         `json:"access_count"`   // 访问次数
	HitCount      int64         `json:"hit_count"`      // 命中次数
	LastAccess    time.Time     `json:"last_access"`    // 最后访问时间
	TTL           time.Duration `json:"ttl"`            // 生存时间
	Size          int64         `json:"size"`           // 数据大小（字节）
	IsHot         bool          `json:"is_hot"`         // 是否为热点键
}

// CacheMemoryStats 缓存内存统计
type CacheMemoryStats struct {
	UsedMemory     int64   `json:"used_memory"`     // 已使用内存（字节）
	MaxMemory      int64   `json:"max_memory"`      // 最大内存（字节）
	MemoryUsage    float64 `json:"memory_usage"`    // 内存使用率
	KeyCount       int64   `json:"key_count"`       // 键数量
	ExpiredKeys    int64   `json:"expired_keys"`    // 过期键数量
	EvictedKeys    int64   `json:"evicted_keys"`    // 被淘汰键数量
	AvgKeySize     int64   `json:"avg_key_size"`    // 平均键大小
	Fragmentation  float64 `json:"fragmentation"`   // 内存碎片率
}

// CacheConnectionStats Redis连接统计
type CacheConnectionStats struct {
	ActiveConnections  int64         `json:"active_connections"`  // 活跃连接数
	IdleConnections    int64         `json:"idle_connections"`    // 空闲连接数
	TotalConnections   int64         `json:"total_connections"`   // 总连接数
	ConnectionFailures int64         `json:"connection_failures"` // 连接失败次数
	ConnectionTimeouts int64         `json:"connection_timeouts"` // 连接超时次数
	AvgConnectTime     time.Duration `json:"avg_connect_time"`    // 平均连接时间
	MaxConnections     int64         `json:"max_connections"`     // 最大连接数
	PoolUtilization    float64       `json:"pool_utilization"`    // 连接池使用率
}

// CacheLatencyStats 缓存延迟统计
type CacheLatencyStats struct {
	AvgLatency       time.Duration   `json:"avg_latency"`        // 平均延迟
	P50Latency       time.Duration   `json:"p50_latency"`        // P50延迟
	P95Latency       time.Duration   `json:"p95_latency"`        // P95延迟
	P99Latency       time.Duration   `json:"p99_latency"`        // P99延迟
	MaxLatency       time.Duration   `json:"max_latency"`        // 最大延迟
	LatencyHistogram map[string]int64 `json:"latency_histogram"` // 延迟直方图
}

// CacheMetricsConfig 缓存监控配置
type CacheMetricsConfig struct {
	Enabled              bool          `json:"enabled"`               // 是否启用
	CacheType            CacheType     `json:"cache_type"`            // 缓存类型
	CollectInterval      time.Duration `json:"collect_interval"`      // 收集间隔
	HotKeyThreshold      int64         `json:"hot_key_threshold"`     // 热点键阈值
	MaxHotKeys           int           `json:"max_hot_keys"`          // 最大热点键数量
	EnableKeyTracking    bool          `json:"enable_key_tracking"`   // 启用键跟踪
	EnableLatencyTracking bool         `json:"enable_latency_tracking"` // 启用延迟跟踪
}

// CacheMetricsCollector 缓存指标收集器
type CacheMetricsCollector struct {
	mu                  sync.RWMutex
	config              *CacheMetricsConfig
	enabled             bool
	startTime           time.Time
	
	// 基础统计
	hitStats            *CacheHitStats
	operationStats      map[CacheOperation]*CacheOperationStats
	memoryStats         *CacheMemoryStats
	connectionStats     *CacheConnectionStats
	latencyStats        *CacheLatencyStats
	
	// 键级别统计
	keyStats            map[string]*CacheKeyStats
	hotKeys             []string
	keyStatsMutex       sync.RWMutex
	
	// 延迟跟踪
	latencyRecords      []time.Duration
	latencyMutex        sync.RWMutex
	
	// 收集器控制
	collectTicker       *time.Ticker
	stopChan            chan struct{}
}

// NewCacheMetricsCollector 创建缓存指标收集器
func NewCacheMetricsCollector(config *CacheMetricsConfig) *CacheMetricsCollector {
	if config == nil {
		config = &CacheMetricsConfig{
			Enabled:               true,
			CacheType:             CacheTypeRedis,
			CollectInterval:       5 * time.Second,
			HotKeyThreshold:       100,
			MaxHotKeys:            20,
			EnableKeyTracking:     true,
			EnableLatencyTracking: true,
		}
	}
	
	collector := &CacheMetricsCollector{
		config:    config,
		enabled:   config.Enabled,
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
		
		hitStats: &CacheHitStats{},
		operationStats: make(map[CacheOperation]*CacheOperationStats),
		memoryStats: &CacheMemoryStats{},
		connectionStats: &CacheConnectionStats{},
		latencyStats: &CacheLatencyStats{
			LatencyHistogram: make(map[string]int64),
		},
		
		keyStats:       make(map[string]*CacheKeyStats),
		hotKeys:        make([]string, 0, config.MaxHotKeys),
		latencyRecords: make([]time.Duration, 0, 1000),
	}
	
	// 初始化操作统计
	operations := []CacheOperation{
		OperationGet, OperationSet, OperationDelete,
		OperationExists, OperationExpire, OperationIncr, OperationDecr,
	}
	
	for _, op := range operations {
		collector.operationStats[op] = &CacheOperationStats{
			Operation:   op,
			MinDuration: time.Hour, // 初始化为大值
		}
	}
	
	// 初始化延迟直方图桶
	collector.initializeLatencyHistogram()
	
	return collector
}

// initializeLatencyHistogram 初始化延迟直方图
func (cmc *CacheMetricsCollector) initializeLatencyHistogram() {
	buckets := []string{
		"0-1ms", "1-5ms", "5-10ms", "10-50ms",
		"50-100ms", "100-500ms", "500ms-1s", "1s+",
	}
	
	for _, bucket := range buckets {
		cmc.latencyStats.LatencyHistogram[bucket] = 0
	}
}

// Start 启动缓存指标收集
func (cmc *CacheMetricsCollector) Start() {
	if !cmc.enabled {
		return
	}
	
	cmc.collectTicker = time.NewTicker(cmc.config.CollectInterval)
	go cmc.collectLoop()
}

// Stop 停止缓存指标收集
func (cmc *CacheMetricsCollector) Stop() {
	if cmc.collectTicker != nil {
		cmc.collectTicker.Stop()
	}
	close(cmc.stopChan)
}

// collectLoop 收集循环
func (cmc *CacheMetricsCollector) collectLoop() {
	for {
		select {
		case <-cmc.collectTicker.C:
			cmc.collectMetrics()
		case <-cmc.stopChan:
			return
		}
	}
}

// collectMetrics 收集指标
func (cmc *CacheMetricsCollector) collectMetrics() {
	if !cmc.enabled {
		return
	}
	
	// 更新命中率
	cmc.updateHitRatio()
	
	// 更新操作平均时间
	cmc.updateOperationAverages()
	
	// 更新热点键
	if cmc.config.EnableKeyTracking {
		cmc.updateHotKeys()
	}
	
	// 更新延迟统计
	if cmc.config.EnableLatencyTracking {
		cmc.updateLatencyStats()
	}
}

// RecordCacheHit 记录缓存命中
func (cmc *CacheMetricsCollector) RecordCacheHit(key string, hit bool) {
	if !cmc.enabled {
		return
	}
	
	cmc.mu.Lock()
	atomic.AddInt64(&cmc.hitStats.TotalRequests, 1)
	if hit {
		atomic.AddInt64(&cmc.hitStats.HitCount, 1)
	} else {
		atomic.AddInt64(&cmc.hitStats.MissCount, 1)
	}
	cmc.mu.Unlock()
	
	// 记录键级别统计
	if cmc.config.EnableKeyTracking && key != "" {
		cmc.recordKeyAccess(key, hit)
	}
}

// RecordCacheOperation 记录缓存操作
func (cmc *CacheMetricsCollector) RecordCacheOperation(operation CacheOperation, key string, duration time.Duration, err error) {
	if !cmc.enabled {
		return
	}
	
	cmc.mu.Lock()
	defer cmc.mu.Unlock()
	
	stats, exists := cmc.operationStats[operation]
	if !exists {
		stats = &CacheOperationStats{
			Operation:   operation,
			MinDuration: duration,
		}
		cmc.operationStats[operation] = stats
	}
	
	// 更新统计
	atomic.AddInt64(&stats.Count, 1)
	stats.TotalDuration += duration
	
	if duration < stats.MinDuration {
		stats.MinDuration = duration
	}
	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}
	
	if err != nil {
		atomic.AddInt64(&stats.ErrorCount, 1)
	}
	
	// 记录延迟
	if cmc.config.EnableLatencyTracking {
		cmc.recordLatency(duration)
	}
}

// recordKeyAccess 记录键访问
func (cmc *CacheMetricsCollector) recordKeyAccess(key string, hit bool) {
	cmc.keyStatsMutex.Lock()
	defer cmc.keyStatsMutex.Unlock()
	
	stats, exists := cmc.keyStats[key]
	if !exists {
		stats = &CacheKeyStats{
			Key:        key,
			LastAccess: time.Now(),
		}
		cmc.keyStats[key] = stats
	}
	
	atomic.AddInt64(&stats.AccessCount, 1)
	if hit {
		atomic.AddInt64(&stats.HitCount, 1)
	}
	stats.LastAccess = time.Now()
}

// recordLatency 记录延迟
func (cmc *CacheMetricsCollector) recordLatency(duration time.Duration) {
	cmc.latencyMutex.Lock()
	defer cmc.latencyMutex.Unlock()
	
	// 添加到延迟记录
	if len(cmc.latencyRecords) >= 1000 {
		// 移除最旧的记录
		cmc.latencyRecords = cmc.latencyRecords[1:]
	}
	cmc.latencyRecords = append(cmc.latencyRecords, duration)
	
	// 更新延迟直方图
	cmc.updateLatencyHistogram(duration)
}

// updateLatencyHistogram 更新延迟直方图
func (cmc *CacheMetricsCollector) updateLatencyHistogram(duration time.Duration) {
	ms := duration.Nanoseconds() / 1000000
	
	switch {
	case ms < 1:
		cmc.latencyStats.LatencyHistogram["0-1ms"]++
	case ms < 5:
		cmc.latencyStats.LatencyHistogram["1-5ms"]++
	case ms < 10:
		cmc.latencyStats.LatencyHistogram["5-10ms"]++
	case ms < 50:
		cmc.latencyStats.LatencyHistogram["10-50ms"]++
	case ms < 100:
		cmc.latencyStats.LatencyHistogram["50-100ms"]++
	case ms < 500:
		cmc.latencyStats.LatencyHistogram["100-500ms"]++
	case ms < 1000:
		cmc.latencyStats.LatencyHistogram["500ms-1s"]++
	default:
		cmc.latencyStats.LatencyHistogram["1s+"]++
	}
}

// updateHitRatio 更新命中率
func (cmc *CacheMetricsCollector) updateHitRatio() {
	cmc.mu.Lock()
	defer cmc.mu.Unlock()
	
	total := cmc.hitStats.TotalRequests
	if total > 0 {
		cmc.hitStats.HitRatio = float64(cmc.hitStats.HitCount) / float64(total) * 100
		cmc.hitStats.MissRatio = float64(cmc.hitStats.MissCount) / float64(total) * 100
	}
}

// updateOperationAverages 更新操作平均值
func (cmc *CacheMetricsCollector) updateOperationAverages() {
	cmc.mu.Lock()
	defer cmc.mu.Unlock()
	
	for _, stats := range cmc.operationStats {
		if stats.Count > 0 {
			stats.AvgDuration = time.Duration(int64(stats.TotalDuration) / stats.Count)
			stats.SuccessRatio = float64(stats.Count-stats.ErrorCount) / float64(stats.Count) * 100
		}
	}
}

// updateHotKeys 更新热点键
func (cmc *CacheMetricsCollector) updateHotKeys() {
	cmc.keyStatsMutex.Lock()
	defer cmc.keyStatsMutex.Unlock()
	
	// 收集访问次数超过阈值的键
	type keyAccess struct {
		key   string
		count int64
	}
	
	var hotKeys []keyAccess
	for key, stats := range cmc.keyStats {
		if stats.AccessCount >= cmc.config.HotKeyThreshold {
			hotKeys = append(hotKeys, keyAccess{key: key, count: stats.AccessCount})
			stats.IsHot = true
		} else {
			stats.IsHot = false
		}
	}
	
	// 按访问次数排序
	sort.Slice(hotKeys, func(i, j int) bool {
		return hotKeys[i].count > hotKeys[j].count
	})
	
	// 更新热点键列表
	cmc.hotKeys = cmc.hotKeys[:0]
	maxKeys := cmc.config.MaxHotKeys
	if len(hotKeys) < maxKeys {
		maxKeys = len(hotKeys)
	}
	
	for i := 0; i < maxKeys; i++ {
		cmc.hotKeys = append(cmc.hotKeys, hotKeys[i].key)
	}
}

// updateLatencyStats 更新延迟统计
func (cmc *CacheMetricsCollector) updateLatencyStats() {
	cmc.latencyMutex.RLock()
	defer cmc.latencyMutex.RUnlock()
	
	if len(cmc.latencyRecords) == 0 {
		return
	}
	
	// 复制并排序延迟记录
	latencies := make([]time.Duration, len(cmc.latencyRecords))
	copy(latencies, cmc.latencyRecords)
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	
	// 计算平均延迟
	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	cmc.latencyStats.AvgLatency = time.Duration(int64(total) / int64(len(latencies)))
	
	// 计算百分位数
	cmc.latencyStats.P50Latency = latencies[len(latencies)*50/100]
	cmc.latencyStats.P95Latency = latencies[len(latencies)*95/100]
	cmc.latencyStats.P99Latency = latencies[len(latencies)*99/100]
	cmc.latencyStats.MaxLatency = latencies[len(latencies)-1]
}

// UpdateMemoryStats 更新内存统计
func (cmc *CacheMetricsCollector) UpdateMemoryStats(usedMemory, maxMemory, keyCount, expiredKeys, evictedKeys int64) {
	if !cmc.enabled {
		return
	}
	
	cmc.mu.Lock()
	defer cmc.mu.Unlock()
	
	cmc.memoryStats.UsedMemory = usedMemory
	cmc.memoryStats.MaxMemory = maxMemory
	cmc.memoryStats.KeyCount = keyCount
	cmc.memoryStats.ExpiredKeys = expiredKeys
	cmc.memoryStats.EvictedKeys = evictedKeys
	
	if maxMemory > 0 {
		cmc.memoryStats.MemoryUsage = float64(usedMemory) / float64(maxMemory) * 100
	}
	
	if keyCount > 0 {
		cmc.memoryStats.AvgKeySize = usedMemory / keyCount
	}
}

// UpdateConnectionStats 更新连接统计
func (cmc *CacheMetricsCollector) UpdateConnectionStats(active, idle, total, failures, timeouts int64, avgConnectTime time.Duration, maxConnections int64) {
	if !cmc.enabled {
		return
	}
	
	cmc.mu.Lock()
	defer cmc.mu.Unlock()
	
	cmc.connectionStats.ActiveConnections = active
	cmc.connectionStats.IdleConnections = idle
	cmc.connectionStats.TotalConnections = total
	cmc.connectionStats.ConnectionFailures = failures
	cmc.connectionStats.ConnectionTimeouts = timeouts
	cmc.connectionStats.AvgConnectTime = avgConnectTime
	cmc.connectionStats.MaxConnections = maxConnections
	
	if maxConnections > 0 {
		cmc.connectionStats.PoolUtilization = float64(active) / float64(maxConnections) * 100
	}
}

// GetHitStats 获取命中统计
func (cmc *CacheMetricsCollector) GetHitStats() *CacheHitStats {
	cmc.mu.RLock()
	defer cmc.mu.RUnlock()
	
	stats := *cmc.hitStats
	return &stats
}

// GetOperationStats 获取操作统计
func (cmc *CacheMetricsCollector) GetOperationStats() map[CacheOperation]*CacheOperationStats {
	cmc.mu.RLock()
	defer cmc.mu.RUnlock()
	
	stats := make(map[CacheOperation]*CacheOperationStats)
	for k, v := range cmc.operationStats {
		statsCopy := *v
		stats[k] = &statsCopy
	}
	return stats
}

// GetMemoryStats 获取内存统计
func (cmc *CacheMetricsCollector) GetMemoryStats() *CacheMemoryStats {
	cmc.mu.RLock()
	defer cmc.mu.RUnlock()
	
	stats := *cmc.memoryStats
	return &stats
}

// GetConnectionStats 获取连接统计
func (cmc *CacheMetricsCollector) GetConnectionStats() *CacheConnectionStats {
	cmc.mu.RLock()
	defer cmc.mu.RUnlock()
	
	stats := *cmc.connectionStats
	return &stats
}

// GetLatencyStats 获取延迟统计
func (cmc *CacheMetricsCollector) GetLatencyStats() *CacheLatencyStats {
	cmc.mu.RLock()
	defer cmc.mu.RUnlock()
	
	stats := *cmc.latencyStats
	stats.LatencyHistogram = make(map[string]int64)
	for k, v := range cmc.latencyStats.LatencyHistogram {
		stats.LatencyHistogram[k] = v
	}
	return &stats
}

// GetHotKeys 获取热点键
func (cmc *CacheMetricsCollector) GetHotKeys() []string {
	cmc.keyStatsMutex.RLock()
	defer cmc.keyStatsMutex.RUnlock()
	
	keys := make([]string, len(cmc.hotKeys))
	copy(keys, cmc.hotKeys)
	return keys
}

// GetKeyStats 获取键统计
func (cmc *CacheMetricsCollector) GetKeyStats() map[string]*CacheKeyStats {
	cmc.keyStatsMutex.RLock()
	defer cmc.keyStatsMutex.RUnlock()
	
	stats := make(map[string]*CacheKeyStats)
	for k, v := range cmc.keyStats {
		statsCopy := *v
		stats[k] = &statsCopy
	}
	return stats
}

// Reset 重置指标
func (cmc *CacheMetricsCollector) Reset() {
	cmc.mu.Lock()
	defer cmc.mu.Unlock()
	
	// 重置命中统计
	cmc.hitStats.TotalRequests = 0
	cmc.hitStats.HitCount = 0
	cmc.hitStats.MissCount = 0
	cmc.hitStats.HitRatio = 0
	cmc.hitStats.MissRatio = 0
	
	// 重置操作统计
	for _, stats := range cmc.operationStats {
		stats.Count = 0
		stats.TotalDuration = 0
		stats.AvgDuration = 0
		stats.MinDuration = time.Hour
		stats.MaxDuration = 0
		stats.ErrorCount = 0
		stats.SuccessRatio = 0
	}
	
	// 重置延迟统计
	for k := range cmc.latencyStats.LatencyHistogram {
		cmc.latencyStats.LatencyHistogram[k] = 0
	}
	cmc.latencyStats.AvgLatency = 0
	cmc.latencyStats.P50Latency = 0
	cmc.latencyStats.P95Latency = 0
	cmc.latencyStats.P99Latency = 0
	cmc.latencyStats.MaxLatency = 0
	
	// 重置键统计
	cmc.keyStatsMutex.Lock()
	cmc.keyStats = make(map[string]*CacheKeyStats)
	cmc.hotKeys = cmc.hotKeys[:0]
	cmc.keyStatsMutex.Unlock()
	
	// 重置延迟记录
	cmc.latencyMutex.Lock()
	cmc.latencyRecords = cmc.latencyRecords[:0]
	cmc.latencyMutex.Unlock()
}

// IsEnabled 检查是否启用
func (cmc *CacheMetricsCollector) IsEnabled() bool {
	cmc.mu.RLock()
	defer cmc.mu.RUnlock()
	return cmc.enabled
}

// Enable 启用收集器
func (cmc *CacheMetricsCollector) Enable() {
	cmc.mu.Lock()
	defer cmc.mu.Unlock()
	cmc.enabled = true
}

// Disable 禁用收集器
func (cmc *CacheMetricsCollector) Disable() {
	cmc.mu.Lock()
	defer cmc.mu.Unlock()
	cmc.enabled = false
}

// CacheMetricsPanel 缓存监控面板
type CacheMetricsPanel struct {
	collector *CacheMetricsCollector
}

// NewCacheMetricsPanel 创建缓存监控面板
func NewCacheMetricsPanel(collector *CacheMetricsCollector) *CacheMetricsPanel {
	return &CacheMetricsPanel{
		collector: collector,
	}
}

// RegisterRoutes 注册路由
func (cmp *CacheMetricsPanel) RegisterRoutes(engine any) {
	var cacheGroup *route.RouterGroup
	
	if h, ok := engine.(*route.Engine); ok {
		cacheGroup = h.Group("/yyhertz/cache")
	} else {
		config.Error("无法注册缓存监控路由，未知引擎类型")
		return
	}
	
	// 注册路由
	cacheGroup.GET("/", cmp.getCacheMetrics)
	cacheGroup.GET("/hit-stats", cmp.getHitStats)
	cacheGroup.GET("/operations", cmp.getOperationStats)
	cacheGroup.GET("/memory", cmp.getMemoryStats)
	cacheGroup.GET("/connections", cmp.getConnectionStats)
	cacheGroup.GET("/latency", cmp.getLatencyStats)
	cacheGroup.GET("/hot-keys", cmp.getHotKeys)
	cacheGroup.POST("/reset", cmp.resetMetrics)
	cacheGroup.POST("/enable", cmp.enableCollector)
	cacheGroup.POST("/disable", cmp.disableCollector)
	cacheGroup.GET("/panel", cmp.cachePanel)
}

// getCacheMetrics 获取缓存指标
func (cmp *CacheMetricsPanel) getCacheMetrics(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"hit_stats":        cmp.collector.GetHitStats(),
			"operation_stats":  cmp.collector.GetOperationStats(),
			"memory_stats":     cmp.collector.GetMemoryStats(),
			"connection_stats": cmp.collector.GetConnectionStats(),
			"latency_stats":    cmp.collector.GetLatencyStats(),
			"hot_keys":         cmp.collector.GetHotKeys(),
			"enabled":          cmp.collector.IsEnabled(),
			"timestamp":        time.Now(),
		},
	})
}

// getHitStats 获取命中统计
func (cmp *CacheMetricsPanel) getHitStats(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": cmp.collector.GetHitStats(),
	})
}

// getOperationStats 获取操作统计
func (cmp *CacheMetricsPanel) getOperationStats(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": cmp.collector.GetOperationStats(),
	})
}

// getMemoryStats 获取内存统计
func (cmp *CacheMetricsPanel) getMemoryStats(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": cmp.collector.GetMemoryStats(),
	})
}

// getConnectionStats 获取连接统计
func (cmp *CacheMetricsPanel) getConnectionStats(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": cmp.collector.GetConnectionStats(),
	})
}

// getLatencyStats 获取延迟统计
func (cmp *CacheMetricsPanel) getLatencyStats(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": cmp.collector.GetLatencyStats(),
	})
}

// getHotKeys 获取热点键
func (cmp *CacheMetricsPanel) getHotKeys(ctx context.Context, c *app.RequestContext) {
	hotKeys := cmp.collector.GetHotKeys()
	keyStats := cmp.collector.GetKeyStats()
	
	var hotKeyDetails []map[string]any
	for _, key := range hotKeys {
		if stats, exists := keyStats[key]; exists {
			hotKeyDetails = append(hotKeyDetails, map[string]any{
				"key":          key,
				"access_count": stats.AccessCount,
				"hit_count":    stats.HitCount,
				"last_access":  stats.LastAccess,
				"is_hot":       stats.IsHot,
			})
		}
	}
	
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"hot_keys":     hotKeys,
			"hot_key_details": hotKeyDetails,
			"total_count":  len(hotKeys),
		},
	})
}

// resetMetrics 重置指标
func (cmp *CacheMetricsPanel) resetMetrics(ctx context.Context, c *app.RequestContext) {
	cmp.collector.Reset()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "缓存指标已重置",
	})
}

// enableCollector 启用收集器
func (cmp *CacheMetricsPanel) enableCollector(ctx context.Context, c *app.RequestContext) {
	cmp.collector.Enable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "缓存指标收集已启用",
		"enabled": true,
	})
}

// disableCollector 禁用收集器
func (cmp *CacheMetricsPanel) disableCollector(ctx context.Context, c *app.RequestContext) {
	cmp.collector.Disable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "缓存指标收集已禁用",
		"enabled": false,
	})
}

// cachePanel 缓存监控面板页面
func (cmp *CacheMetricsPanel) cachePanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz 缓存监控面板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .metric-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-card h3 { margin-top: 0; color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
        .metric-item { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #eee; }
        .metric-item:last-child { border-bottom: none; }
        .metric-label { font-weight: bold; color: #555; }
        .metric-value { color: #007bff; font-weight: bold; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-warning { background: #ffc107; color: black; }
        .hot-keys { background: white; margin-top: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .hot-key-item { padding: 15px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; }
        .hot-key-name { font-family: monospace; font-weight: bold; }
        .hot-key-stats { color: #666; font-size: 0.9em; }
        .status-indicator { width: 12px; height: 12px; border-radius: 50%; display: inline-block; margin-right: 5px; }
        .status-healthy { background: #28a745; }
        .status-warning { background: #ffc107; }
        .status-error { background: #dc3545; }
        .hit-ratio-bar { width: 100%; height: 20px; background: #f0f0f0; border-radius: 10px; overflow: hidden; margin-top: 5px; }
        .hit-ratio-fill { height: 100%; background: linear-gradient(to right, #28a745, #007bff); }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz 缓存监控面板</h1>
        <div>
            <button class="btn btn-primary" onclick="refreshMetrics()">刷新指标</button>
            <button class="btn btn-success" onclick="enableCollector()">启用收集</button>
            <button class="btn btn-danger" onclick="disableCollector()">禁用收集</button>
            <button class="btn btn-warning" onclick="resetMetrics()">重置指标</button>
        </div>
    </div>

    <div class="metrics-grid">
        <div class="metric-card">
            <h3><span class="status-indicator status-healthy"></span>缓存命中率</h3>
            <div id="hitStats">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-healthy"></span>操作统计</h3>
            <div id="operationStats">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-healthy"></span>内存使用</h3>
            <div id="memoryStats">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-healthy"></span>连接状态</h3>
            <div id="connectionStats">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-healthy"></span>延迟统计</h3>
            <div id="latencyStats">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>
    </div>

    <div class="hot-keys">
        <h3 style="padding: 20px; margin: 0; border-bottom: 1px solid #eee;">热点键分析</h3>
        <div id="hotKeys">
            <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
        </div>
    </div>

    <script>
        function refreshMetrics() {
            loadCacheMetrics();
        }

        function loadCacheMetrics() {
            Promise.all([
                fetch('/yyhertz/cache/hit-stats'),
                fetch('/yyhertz/cache/operations'),
                fetch('/yyhertz/cache/memory'),
                fetch('/yyhertz/cache/connections'),
                fetch('/yyhertz/cache/latency'),
                fetch('/yyhertz/cache/hot-keys')
            ])
            .then(responses => Promise.all(responses.map(r => r.json())))
            .then(([hitStats, operations, memory, connections, latency, hotKeys]) => {
                showHitStats(hitStats.data);
                showOperationStats(operations.data);
                showMemoryStats(memory.data);
                showConnectionStats(connections.data);
                showLatencyStats(latency.data);
                showHotKeys(hotKeys.data.hot_key_details || []);
            })
            .catch(error => {
                console.error('加载缓存指标失败:', error);
            });
        }

        function showHitStats(data) {
            const container = document.getElementById('hitStats');
            const hitRatio = data.hit_ratio || 0;
            
            container.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">总请求数</span><span class="metric-value">' + (data.total_requests || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">命中次数</span><span class="metric-value">' + (data.hit_count || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">未命中次数</span><span class="metric-value">' + (data.miss_count || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">命中率</span><span class="metric-value">' + hitRatio.toFixed(2) + '%</span></div>' +
                '<div class="hit-ratio-bar"><div class="hit-ratio-fill" style="width: ' + hitRatio + '%"></div></div>';
        }

        function showOperationStats(data) {
            const container = document.getElementById('operationStats');
            let html = '';
            
            for (const [operation, stats] of Object.entries(data)) {
                html += '<div class="metric-item">' +
                    '<span class="metric-label">' + operation + '</span>' +
                    '<span class="metric-value">' + (stats.count || 0) + ' (' + formatDuration(stats.avg_duration || 0) + ')</span>' +
                    '</div>';
            }
            
            container.innerHTML = html || '<div style="text-align: center; padding: 20px; color: #666;">暂无操作数据</div>';
        }

        function showMemoryStats(data) {
            const container = document.getElementById('memoryStats');
            const usage = data.memory_usage || 0;
            
            container.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">已使用内存</span><span class="metric-value">' + formatBytes(data.used_memory || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">最大内存</span><span class="metric-value">' + formatBytes(data.max_memory || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">使用率</span><span class="metric-value">' + usage.toFixed(2) + '%</span></div>' +
                '<div class="metric-item"><span class="metric-label">键数量</span><span class="metric-value">' + (data.key_count || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">过期键</span><span class="metric-value">' + (data.expired_keys || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">淘汰键</span><span class="metric-value">' + (data.evicted_keys || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">平均键大小</span><span class="metric-value">' + formatBytes(data.avg_key_size || 0) + '</span></div>';
        }

        function showConnectionStats(data) {
            const container = document.getElementById('connectionStats');
            const utilization = data.pool_utilization || 0;
            
            container.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">活跃连接</span><span class="metric-value">' + (data.active_connections || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">空闲连接</span><span class="metric-value">' + (data.idle_connections || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">总连接数</span><span class="metric-value">' + (data.total_connections || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">最大连接</span><span class="metric-value">' + (data.max_connections || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">使用率</span><span class="metric-value">' + utilization.toFixed(1) + '%</span></div>' +
                '<div class="metric-item"><span class="metric-label">连接失败</span><span class="metric-value">' + (data.connection_failures || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">连接超时</span><span class="metric-value">' + (data.connection_timeouts || 0) + '</span></div>';
        }

        function showLatencyStats(data) {
            const container = document.getElementById('latencyStats');
            
            container.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">平均延迟</span><span class="metric-value">' + formatDuration(data.avg_latency || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">P50延迟</span><span class="metric-value">' + formatDuration(data.p50_latency || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">P95延迟</span><span class="metric-value">' + formatDuration(data.p95_latency || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">P99延迟</span><span class="metric-value">' + formatDuration(data.p99_latency || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">最大延迟</span><span class="metric-value">' + formatDuration(data.max_latency || 0) + '</span></div>';
        }

        function showHotKeys(hotKeyDetails) {
            const container = document.getElementById('hotKeys');
            
            if (hotKeyDetails.length === 0) {
                container.innerHTML = '<div style="padding: 20px; text-align: center; color: #666;">暂无热点键</div>';
                return;
            }

            let html = '';
            hotKeyDetails.forEach((keyData, index) => {
                const hitRate = keyData.access_count > 0 ? 
                    (keyData.hit_count / keyData.access_count * 100).toFixed(1) : 0;
                
                html += '<div class="hot-key-item">' +
                    '<div>' +
                    '<div class="hot-key-name">' + keyData.key + '</div>' +
                    '<div class="hot-key-stats">访问: ' + keyData.access_count + ' | 命中: ' + keyData.hit_count + ' | 命中率: ' + hitRate + '%</div>' +
                    '</div>' +
                    '<div style="text-align: right;">' +
                    '<div style="color: #007bff; font-weight: bold;">#' + (index + 1) + '</div>' +
                    '<div style="font-size: 0.8em; color: #666;">' + new Date(keyData.last_access).toLocaleString() + '</div>' +
                    '</div>' +
                    '</div>';
            });
            
            container.innerHTML = html;
        }

        function formatDuration(nanoseconds) {
            if (nanoseconds < 1000) {
                return nanoseconds + 'ns';
            } else if (nanoseconds < 1000000) {
                return (nanoseconds / 1000).toFixed(1) + 'μs';
            } else if (nanoseconds < 1000000000) {
                return (nanoseconds / 1000000).toFixed(1) + 'ms';
            } else {
                return (nanoseconds / 1000000000).toFixed(2) + 's';
            }
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function enableCollector() {
            fetch('/yyhertz/cache/enable', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('缓存指标收集已启用');
                    refreshMetrics();
                })
                .catch(error => {
                    console.error('启用失败:', error);
                    alert('启用失败');
                });
        }

        function disableCollector() {
            if (confirm('确定要禁用缓存指标收集吗？')) {
                fetch('/yyhertz/cache/disable', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('缓存指标收集已禁用');
                        refreshMetrics();
                    })
                    .catch(error => {
                        console.error('禁用失败:', error);
                        alert('禁用失败');
                    });
            }
        }

        function resetMetrics() {
            if (confirm('确定要重置所有缓存指标吗？')) {
                fetch('/yyhertz/cache/reset', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('缓存指标已重置');
                        refreshMetrics();
                    })
                    .catch(error => {
                        console.error('重置失败:', error);
                        alert('重置失败');
                    });
            }
        }

        // 页面加载时初始化
        window.onload = function() {
            loadCacheMetrics();
            // 每30秒自动刷新
            setInterval(loadCacheMetrics, 30000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}