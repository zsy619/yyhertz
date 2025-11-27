// Package mybatis 实时缓存监控系统
//
// 功能特性：
// 1. 多级缓存命中率实时监控
// 2. 缓存性能热点分析
// 3. 智能缓存预热策略
// 4. 缓存失效模式识别
// 5. 缓存容量优化建议
package mybatis

import (
	"container/ring"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CacheHitRateMonitor 缓存命中率监控器
type CacheHitRateMonitor struct {
	// 基础配置
	config *CacheMonitorConfig

	// 缓存层监控器
	l1Monitor        *CacheLayerMonitor     // 一级缓存（Session级别）
	l2Monitor        *CacheLayerMonitor     // 二级缓存（应用级别）
	resultSetMonitor *ResultSetCacheMonitor // 结果集缓存
	queryMonitor     *QueryCacheMonitor     // 查询缓存

	// 实时指标
	globalStats     *GlobalCacheStats
	realtimeMetrics *RealtimeCacheMetrics

	// 热点分析
	hotspotAnalyzer *CacheHotspotAnalyzer

	// 预热策略
	preWarmEngine *CachePrewarmEngine

	// 状态管理
	isRunning int32
	stopChan  chan struct{}
	mutex     sync.RWMutex
}

// CacheMonitorConfig 缓存监控配置
type CacheMonitorConfig struct {
	// 基础配置
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	MonitorInterval time.Duration `yaml:"monitor_interval" json:"monitor_interval"`
	SampleWindow    time.Duration `yaml:"sample_window" json:"sample_window"`

	// 监控层级
	MonitorL1Cache    bool `yaml:"monitor_l1_cache" json:"monitor_l1_cache"`
	MonitorL2Cache    bool `yaml:"monitor_l2_cache" json:"monitor_l2_cache"`
	MonitorResultSet  bool `yaml:"monitor_result_set" json:"monitor_result_set"`
	MonitorQueryCache bool `yaml:"monitor_query_cache" json:"monitor_query_cache"`

	// 热点分析配置
	HotspotThreshold  float64 `yaml:"hotspot_threshold" json:"hotspot_threshold"`
	HotspotWindowSize int     `yaml:"hotspot_window_size" json:"hotspot_window_size"`

	// 预警阈值
	LowHitRateThreshold   float64 `yaml:"low_hit_rate_threshold" json:"low_hit_rate_threshold"`
	HighMissRateThreshold float64 `yaml:"high_miss_rate_threshold" json:"high_miss_rate_threshold"`

	// 预热配置
	EnablePrewarm    bool    `yaml:"enable_prewarm" json:"enable_prewarm"`
	PrewarmThreshold int     `yaml:"prewarm_threshold" json:"prewarm_threshold"`
	PrewarmRatio     float64 `yaml:"prewarm_ratio" json:"prewarm_ratio"`

	// 报告配置
	ReportInterval       time.Duration `yaml:"report_interval" json:"report_interval"`
	EnableDetailedReport bool          `yaml:"enable_detailed_report" json:"enable_detailed_report"`
}

// CacheLayerMonitor 缓存层监控器
type CacheLayerMonitor struct {
	LayerName      string                      `json:"layer_name"`
	Stats          *CacheLayerStats            `json:"stats"`
	TimeSeriesData *ring.Ring                  `json:"-"` // 时间序列数据
	KeyPatterns    map[string]*KeyPatternStats `json:"key_patterns"`
	mutex          sync.RWMutex
}

// CacheLayerStats 缓存层统计信息
type CacheLayerStats struct {
	// 基础指标
	TotalRequests int64 `json:"total_requests"`
	CacheHits     int64 `json:"cache_hits"`
	CacheMisses   int64 `json:"cache_misses"`

	// 命中率指标
	HitRate  float64 `json:"hit_rate"`
	MissRate float64 `json:"miss_rate"`

	// 性能指标
	AvgHitTime  time.Duration `json:"avg_hit_time"`
	AvgMissTime time.Duration `json:"avg_miss_time"`

	// 容量指标
	CacheSize        int64   `json:"cache_size"`
	MaxCacheSize     int64   `json:"max_cache_size"`
	CacheUtilization float64 `json:"cache_utilization"`

	// 操作指标
	CacheUpdates     int64 `json:"cache_updates"`
	CacheEvictions   int64 `json:"cache_evictions"`
	CacheExpirations int64 `json:"cache_expirations"`

	// 时间戳
	LastUpdated time.Time `json:"last_updated"`
	StartTime   time.Time `json:"start_time"`
}

// ResultSetCacheMonitor 结果集缓存监控器
type ResultSetCacheMonitor struct {
	Stats            *ResultSetCacheStats `json:"stats"`
	QueryAnalyzer    *QueryResultAnalyzer `json:"-"`
	SizeDistribution map[string]int64     `json:"size_distribution"`
	mutex            sync.RWMutex
}

// ResultSetCacheStats 结果集缓存统计
type ResultSetCacheStats struct {
	// 基础指标
	TotalQueries  int64 `json:"total_queries"`
	CachedResults int64 `json:"cached_results"`
	CacheHits     int64 `json:"cache_hits"`

	// 大小指标
	AvgResultSize   int64 `json:"avg_result_size"`
	MaxResultSize   int64 `json:"max_result_size"`
	TotalResultSize int64 `json:"total_result_size"`

	// 性能指标
	AvgSerializationTime   time.Duration `json:"avg_serialization_time"`
	AvgDeserializationTime time.Duration `json:"avg_deserialization_time"`

	// 有效性指标
	InvalidationCount int64 `json:"invalidation_count"`
	DirtyReadCount    int64 `json:"dirty_read_count"`
}

// QueryCacheMonitor 查询缓存监控器
type QueryCacheMonitor struct {
	Stats                   *QueryCacheStats          `json:"stats"`
	QueryComplexityAnalyzer *QueryComplexityAnalyzer  `json:"-"`
	ParameterAnalyzer       *ParameterPatternAnalyzer `json:"-"`
	mutex                   sync.RWMutex
}

// QueryCacheStats 查询缓存统计
type QueryCacheStats struct {
	// 基础指标
	UniqueQueries     int64 `json:"unique_queries"`
	ParametricQueries int64 `json:"parametric_queries"`
	CacheableQueries  int64 `json:"cacheable_queries"`
	CacheHits         int64 `json:"cache_hits"`   // 修复：添加缺失的字段
	CacheMisses       int64 `json:"cache_misses"` // 修复：添加缺失的字段

	// 复杂度分析
	SimpleQueries  int64 `json:"simple_queries"`
	ComplexQueries int64 `json:"complex_queries"`

	// 参数分析
	HighVariabilityQueries int64 `json:"high_variability_queries"`
	LowVariabilityQueries  int64 `json:"low_variability_queries"`

	// 性能影响
	CacheableHitRate float64 `json:"cacheable_hit_rate"`
	NonCacheableRate float64 `json:"non_cacheable_rate"`
}

// GlobalCacheStats 全局缓存统计
type GlobalCacheStats struct {
	// 整体指标
	OverallHitRate  float64 `json:"overall_hit_rate"`
	OverallMissRate float64 `json:"overall_miss_rate"`
	TotalCacheSize  int64   `json:"total_cache_size"`

	// 层级对比
	L1HitRate        float64 `json:"l1_hit_rate"`
	L2HitRate        float64 `json:"l2_hit_rate"`
	ResultSetHitRate float64 `json:"result_set_hit_rate"`
	QueryHitRate     float64 `json:"query_hit_rate"`

	// 性能影响
	CacheSpeedupRatio  float64 `json:"cache_speedup_ratio"`
	CacheOverheadRatio float64 `json:"cache_overhead_ratio"`

	// 资源使用
	MemoryUsage int64   `json:"memory_usage"`
	CPUOverhead float64 `json:"cpu_overhead"`

	mutex sync.RWMutex
}

// RealtimeCacheMetrics 实时缓存指标
type RealtimeCacheMetrics struct {
	// 实时指标（最近1分钟）
	RecentHitRate        float64 `json:"recent_hit_rate"`
	RecentMissRate       float64 `json:"recent_miss_rate"`
	RecentRequestsPerSec float64 `json:"recent_requests_per_sec"`

	// 趋势分析
	HitRateTrend  string `json:"hit_rate_trend"` // UP, DOWN, STABLE
	MissRateTrend string `json:"miss_rate_trend"`
	RequestTrend  string `json:"request_trend"`

	// 预测指标
	PredictedHitRate  float64 `json:"predicted_hit_rate"`
	PredictedMissRate float64 `json:"predicted_miss_rate"`

	// 告警状态
	AlertLevel   string `json:"alert_level"` // NORMAL, WARNING, CRITICAL
	AlertMessage string `json:"alert_message"`

	LastUpdated time.Time `json:"last_updated"`
	mutex       sync.RWMutex
}

// KeyPatternStats 键模式统计
type KeyPatternStats struct {
	Pattern            string    `json:"pattern"`
	Count              int64     `json:"count"`
	HitRate            float64   `json:"hit_rate"`
	AvgAccessFrequency float64   `json:"avg_access_frequency"`
	LastAccessTime     time.Time `json:"last_access_time"`
}

// CacheHotspotAnalyzer 缓存热点分析器
type CacheHotspotAnalyzer struct {
	config         *CacheMonitorConfig
	hotspots       map[string]*HotspotData
	analysisWindow *ring.Ring
	mutex          sync.RWMutex
}

// HotspotData 热点数据
type HotspotData struct {
	Key            string    `json:"key"`
	AccessCount    int64     `json:"access_count"`
	HitRate        float64   `json:"hit_rate"`
	LastAccessTime time.Time `json:"last_access_time"`
	AccessPattern  string    `json:"access_pattern"`
	Score          float64   `json:"score"`
}

// CachePrewarmEngine 缓存预热引擎
type CachePrewarmEngine struct {
	config            *CacheMonitorConfig
	prewarmCandidates []*PrewarmCandidate
	prewarmHistory    map[string]*PrewarmResult
	mutex             sync.RWMutex
}

// PrewarmCandidate 预热候选项
type PrewarmCandidate struct {
	QueryPattern      string        `json:"query_pattern"`
	Parameters        []any         `json:"parameters"`
	Priority          int           `json:"priority"`
	EstimatedBenefit  float64       `json:"estimated_benefit"`
	LastExecutionTime time.Duration `json:"last_execution_time"`
}

// PrewarmResult 预热结果
type PrewarmResult struct {
	QueryPattern       string        `json:"query_pattern"`
	PrewarmTime        time.Time     `json:"prewarm_time"`
	Success            bool          `json:"success"`
	ExecutionTime      time.Duration `json:"execution_time"`
	HitRateImprovement float64       `json:"hit_rate_improvement"`
}

// QueryResultAnalyzer 查询结果分析器
type QueryResultAnalyzer struct {
	resultSizes        []int64
	serializationTimes []time.Duration
	complexityScores   []float64
	mutex              sync.RWMutex
}

// QueryComplexityAnalyzer 查询复杂度分析器
type QueryComplexityAnalyzer struct {
	complexityPatterns map[string]float64
	mutex              sync.RWMutex
}

// ParameterPatternAnalyzer 参数模式分析器
type ParameterPatternAnalyzer struct {
	parameterPatterns map[string]*ParameterPattern
	mutex             sync.RWMutex
}

// ParameterPattern 参数模式
type ParameterPattern struct {
	Pattern          string  `json:"pattern"`
	Variability      float64 `json:"variability"`
	UniqueValues     int     `json:"unique_values"`
	TotalOccurrences int     `json:"total_occurrences"`
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	HitRate      float64   `json:"hit_rate"`
	MissRate     float64   `json:"miss_rate"`
	RequestCount int64     `json:"request_count"`
}

// NewCacheHitRateMonitor 创建缓存命中率监控器
func NewCacheHitRateMonitor(config *CacheMonitorConfig) *CacheHitRateMonitor {
	if config == nil {
		config = DefaultCacheMonitorConfig()
	}

	monitor := &CacheHitRateMonitor{
		config:          config,
		globalStats:     &GlobalCacheStats{},
		realtimeMetrics: &RealtimeCacheMetrics{},
		stopChan:        make(chan struct{}),
	}

	// 初始化各层监控器
	if config.MonitorL1Cache {
		monitor.l1Monitor = NewCacheLayerMonitor("L1_CACHE")
	}

	if config.MonitorL2Cache {
		monitor.l2Monitor = NewCacheLayerMonitor("L2_CACHE")
	}

	if config.MonitorResultSet {
		monitor.resultSetMonitor = &ResultSetCacheMonitor{
			Stats:            &ResultSetCacheStats{},
			QueryAnalyzer:    &QueryResultAnalyzer{},
			SizeDistribution: make(map[string]int64),
		}
	}

	if config.MonitorQueryCache {
		monitor.queryMonitor = &QueryCacheMonitor{
			Stats: &QueryCacheStats{},
			QueryComplexityAnalyzer: &QueryComplexityAnalyzer{
				complexityPatterns: make(map[string]float64),
			},
			ParameterAnalyzer: &ParameterPatternAnalyzer{
				parameterPatterns: make(map[string]*ParameterPattern),
			},
		}
	}

	// 初始化热点分析器
	monitor.hotspotAnalyzer = &CacheHotspotAnalyzer{
		config:         config,
		hotspots:       make(map[string]*HotspotData),
		analysisWindow: ring.New(config.HotspotWindowSize),
	}

	// 初始化预热引擎
	if config.EnablePrewarm {
		monitor.preWarmEngine = &CachePrewarmEngine{
			config:            config,
			prewarmCandidates: make([]*PrewarmCandidate, 0),
			prewarmHistory:    make(map[string]*PrewarmResult),
		}
	}

	return monitor
}

// DefaultCacheMonitorConfig 默认缓存监控配置
func DefaultCacheMonitorConfig() *CacheMonitorConfig {
	return &CacheMonitorConfig{
		Enabled:               true,
		MonitorInterval:       10 * time.Second,
		SampleWindow:          5 * time.Minute,
		MonitorL1Cache:        true,
		MonitorL2Cache:        true,
		MonitorResultSet:      true,
		MonitorQueryCache:     true,
		HotspotThreshold:      0.1, // 10%的访问比例认为是热点
		HotspotWindowSize:     100,
		LowHitRateThreshold:   0.6, // 60%命中率以下告警
		HighMissRateThreshold: 0.4, // 40%失效率以上告警
		EnablePrewarm:         true,
		PrewarmThreshold:      10,
		PrewarmRatio:          0.2,
		ReportInterval:        1 * time.Minute,
		EnableDetailedReport:  true,
	}
}

// NewCacheLayerMonitor 创建缓存层监控器
func NewCacheLayerMonitor(layerName string) *CacheLayerMonitor {
	return &CacheLayerMonitor{
		LayerName:      layerName,
		Stats:          &CacheLayerStats{StartTime: time.Now()},
		TimeSeriesData: ring.New(300), // 5分钟的历史数据（每秒一个点）
		KeyPatterns:    make(map[string]*KeyPatternStats),
	}
}

// Start 启动缓存监控器
func (chrm *CacheHitRateMonitor) Start() error {
	if !atomic.CompareAndSwapInt32(&chrm.isRunning, 0, 1) {
		return fmt.Errorf("cache monitor is already running")
	}

	if !chrm.config.Enabled {
		return fmt.Errorf("cache monitor is disabled")
	}

	// 启动实时监控协程
	go chrm.realtimeMonitorLoop()

	// 启动热点分析协程
	go chrm.hotspotAnalysisLoop()

	// 启动预热引擎协程
	if chrm.config.EnablePrewarm && chrm.preWarmEngine != nil {
		go chrm.prewarmEngineLoop()
	}

	// 启动报告生成协程
	if chrm.config.ReportInterval > 0 {
		go chrm.reportLoop()
	}

	log.Println("[CacheHitRateMonitor] Cache hit rate monitor started")
	return nil
}

// Stop 停止缓存监控器
func (chrm *CacheHitRateMonitor) Stop() error {
	if !atomic.CompareAndSwapInt32(&chrm.isRunning, 1, 0) {
		return fmt.Errorf("cache monitor is not running")
	}

	close(chrm.stopChan)
	log.Println("[CacheHitRateMonitor] Cache hit rate monitor stopped")
	return nil
}

// RecordCacheAccess 记录缓存访问
func (chrm *CacheHitRateMonitor) RecordCacheAccess(layer string, key string, hit bool, duration time.Duration) {
	if atomic.LoadInt32(&chrm.isRunning) == 0 {
		return
	}

	// 根据缓存层记录访问
	switch layer {
	case "L1":
		if chrm.l1Monitor != nil {
			chrm.recordLayerAccess(chrm.l1Monitor, key, hit, duration)
		}
	case "L2":
		if chrm.l2Monitor != nil {
			chrm.recordLayerAccess(chrm.l2Monitor, key, hit, duration)
		}
	}

	// 更新全局统计
	chrm.updateGlobalStats(hit, duration)

	// 更新热点分析
	chrm.updateHotspotAnalysis(key, hit)
}

// RecordResultSetCache 记录结果集缓存访问
func (chrm *CacheHitRateMonitor) RecordResultSetCache(query string, resultSize int64, hit bool, serializationTime time.Duration) {
	if atomic.LoadInt32(&chrm.isRunning) == 0 || chrm.resultSetMonitor == nil {
		return
	}

	chrm.resultSetMonitor.mutex.Lock()
	defer chrm.resultSetMonitor.mutex.Unlock()

	stats := chrm.resultSetMonitor.Stats
	stats.TotalQueries++

	if hit {
		stats.CacheHits++
	} else {
		stats.CachedResults++
	}

	// 更新结果大小统计
	if resultSize > stats.MaxResultSize {
		stats.MaxResultSize = resultSize
	}

	stats.TotalResultSize += resultSize
	if stats.TotalQueries > 0 {
		stats.AvgResultSize = stats.TotalResultSize / stats.TotalQueries
	}

	// 更新序列化时间
	if serializationTime > 0 {
		if stats.AvgSerializationTime == 0 {
			stats.AvgSerializationTime = serializationTime
		} else {
			stats.AvgSerializationTime = (stats.AvgSerializationTime + serializationTime) / 2
		}
	}

	// 分析结果大小分布
	chrm.analyzeResultSizeDistribution(resultSize)
}

// RecordQueryCache 记录查询缓存
func (chrm *CacheHitRateMonitor) RecordQueryCache(query string, params []any, complexity float64, hit bool) {
	if atomic.LoadInt32(&chrm.isRunning) == 0 || chrm.queryMonitor == nil {
		return
	}

	chrm.queryMonitor.mutex.Lock()
	defer chrm.queryMonitor.mutex.Unlock()

	stats := chrm.queryMonitor.Stats
	stats.UniqueQueries++

	// 记录缓存命中/失效
	if hit {
		stats.CacheHits++
	} else {
		stats.CacheMisses++
	}

	// 分析查询复杂度
	if complexity < 0.5 {
		stats.SimpleQueries++
	} else {
		stats.ComplexQueries++
	}

	// 分析参数模式
	paramPattern := chrm.analyzeParameterPattern(params)
	chrm.updateParameterAnalysis(paramPattern, params)

	// 评估缓存能力
	if chrm.isQueryCacheable(query, params) {
		stats.CacheableQueries++
		if hit {
			// 更新可缓存查询命中率
			if stats.CacheableQueries > 0 {
				stats.CacheableHitRate = float64(stats.CacheHits) / float64(stats.CacheableQueries)
			}
		}
	}
}

// GetRealTimeMetrics 获取实时指标
func (chrm *CacheHitRateMonitor) GetRealTimeMetrics() *RealtimeCacheMetrics {
	chrm.realtimeMetrics.mutex.RLock()
	defer chrm.realtimeMetrics.mutex.RUnlock()

	metricsCopy := *chrm.realtimeMetrics
	return &metricsCopy
}

// GetGlobalStats 获取全局统计
func (chrm *CacheHitRateMonitor) GetGlobalStats() *GlobalCacheStats {
	chrm.globalStats.mutex.RLock()
	defer chrm.globalStats.mutex.RUnlock()

	statsCopy := *chrm.globalStats
	return &statsCopy
}

// GetLayerStats 获取缓存层统计
func (chrm *CacheHitRateMonitor) GetLayerStats(layer string) *CacheLayerStats {
	switch layer {
	case "L1":
		if chrm.l1Monitor != nil {
			chrm.l1Monitor.mutex.RLock()
			defer chrm.l1Monitor.mutex.RUnlock()
			statsCopy := *chrm.l1Monitor.Stats
			return &statsCopy
		}
	case "L2":
		if chrm.l2Monitor != nil {
			chrm.l2Monitor.mutex.RLock()
			defer chrm.l2Monitor.mutex.RUnlock()
			statsCopy := *chrm.l2Monitor.Stats
			return &statsCopy
		}
	}
	return nil
}

// GetHotspots 获取热点数据
func (chrm *CacheHitRateMonitor) GetHotspots() []*HotspotData {
	chrm.hotspotAnalyzer.mutex.RLock()
	defer chrm.hotspotAnalyzer.mutex.RUnlock()

	hotspots := make([]*HotspotData, 0, len(chrm.hotspotAnalyzer.hotspots))
	for _, hotspot := range chrm.hotspotAnalyzer.hotspots {
		hotspotCopy := *hotspot
		hotspots = append(hotspots, &hotspotCopy)
	}

	// 按分数排序
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].Score > hotspots[j].Score
	})

	return hotspots
}

// GenerateCacheReport 生成缓存报告
func (chrm *CacheHitRateMonitor) GenerateCacheReport() *CachePerformanceReport {
	report := &CachePerformanceReport{
		Timestamp:       time.Now(),
		GlobalStats:     chrm.GetGlobalStats(),
		RealTimeMetrics: chrm.GetRealTimeMetrics(),
		Hotspots:        chrm.GetHotspots(),
		Recommendations: chrm.generateRecommendations(),
	}

	// 添加各层统计
	if chrm.l1Monitor != nil {
		report.L1Stats = chrm.GetLayerStats("L1")
	}
	if chrm.l2Monitor != nil {
		report.L2Stats = chrm.GetLayerStats("L2")
	}

	// 添加结果集缓存统计
	if chrm.resultSetMonitor != nil {
		chrm.resultSetMonitor.mutex.RLock()
		report.ResultSetStats = chrm.resultSetMonitor.Stats
		chrm.resultSetMonitor.mutex.RUnlock()
	}

	// 添加查询缓存统计
	if chrm.queryMonitor != nil {
		chrm.queryMonitor.mutex.RLock()
		report.QueryStats = chrm.queryMonitor.Stats
		chrm.queryMonitor.mutex.RUnlock()
	}

	return report
}

// CachePerformanceReport 缓存性能报告
type CachePerformanceReport struct {
	Timestamp       time.Time             `json:"timestamp"`
	GlobalStats     *GlobalCacheStats     `json:"global_stats"`
	RealTimeMetrics *RealtimeCacheMetrics `json:"realtime_metrics"`
	L1Stats         *CacheLayerStats      `json:"l1_stats,omitempty"`
	L2Stats         *CacheLayerStats      `json:"l2_stats,omitempty"`
	ResultSetStats  *ResultSetCacheStats  `json:"result_set_stats,omitempty"`
	QueryStats      *QueryCacheStats      `json:"query_stats,omitempty"`
	Hotspots        []*HotspotData        `json:"hotspots"`
	Recommendations []string              `json:"recommendations"`
}

// 内部方法实现

// recordLayerAccess 记录缓存层访问
func (chrm *CacheHitRateMonitor) recordLayerAccess(monitor *CacheLayerMonitor, key string, hit bool, duration time.Duration) {
	monitor.mutex.Lock()
	defer monitor.mutex.Unlock()

	stats := monitor.Stats
	stats.TotalRequests++

	if hit {
		stats.CacheHits++
		// 更新命中时间统计
		if stats.AvgHitTime == 0 {
			stats.AvgHitTime = duration
		} else {
			stats.AvgHitTime = (stats.AvgHitTime + duration) / 2
		}
	} else {
		stats.CacheMisses++
		// 更新失效时间统计
		if stats.AvgMissTime == 0 {
			stats.AvgMissTime = duration
		} else {
			stats.AvgMissTime = (stats.AvgMissTime + duration) / 2
		}
	}

	// 更新命中率
	if stats.TotalRequests > 0 {
		stats.HitRate = float64(stats.CacheHits) / float64(stats.TotalRequests)
		stats.MissRate = float64(stats.CacheMisses) / float64(stats.TotalRequests)
	}

	stats.LastUpdated = time.Now()

	// 记录时间序列数据
	chrm.recordTimeSeriesData(monitor, stats.HitRate, stats.MissRate, 1)

	// 更新键模式统计
	chrm.updateKeyPatternStats(monitor, key, hit)
}

// updateGlobalStats 更新全局统计
func (chrm *CacheHitRateMonitor) updateGlobalStats(hit bool, duration time.Duration) {
	chrm.globalStats.mutex.Lock()
	defer chrm.globalStats.mutex.Unlock()

	// 这里应该基于各层的统计计算全局统计
	// 简化实现，实际应该加权平均
	if chrm.l1Monitor != nil && chrm.l2Monitor != nil {
		l1Stats := chrm.l1Monitor.Stats
		l2Stats := chrm.l2Monitor.Stats

		totalRequests := l1Stats.TotalRequests + l2Stats.TotalRequests
		if totalRequests > 0 {
			totalHits := l1Stats.CacheHits + l2Stats.CacheHits
			chrm.globalStats.OverallHitRate = float64(totalHits) / float64(totalRequests)
			chrm.globalStats.OverallMissRate = 1.0 - chrm.globalStats.OverallHitRate
		}

		chrm.globalStats.L1HitRate = l1Stats.HitRate
		chrm.globalStats.L2HitRate = l2Stats.HitRate
	}
}

// updateHotspotAnalysis 更新热点分析
func (chrm *CacheHitRateMonitor) updateHotspotAnalysis(key string, hit bool) {
	chrm.hotspotAnalyzer.mutex.Lock()
	defer chrm.hotspotAnalyzer.mutex.Unlock()

	hotspot, exists := chrm.hotspotAnalyzer.hotspots[key]
	if !exists {
		hotspot = &HotspotData{
			Key:            key,
			LastAccessTime: time.Now(),
		}
		chrm.hotspotAnalyzer.hotspots[key] = hotspot
	}

	hotspot.AccessCount++
	hotspot.LastAccessTime = time.Now()

	// 更新命中率
	if hit {
		hotspot.HitRate = (hotspot.HitRate*float64(hotspot.AccessCount-1) + 1.0) / float64(hotspot.AccessCount)
	} else {
		hotspot.HitRate = (hotspot.HitRate * float64(hotspot.AccessCount-1)) / float64(hotspot.AccessCount)
	}

	// 计算热点分数（基于访问频率和命中率）
	hotspot.Score = float64(hotspot.AccessCount) * (hotspot.HitRate + 0.1) // +0.1避免为0
}

// 更多内部方法...

func (chrm *CacheHitRateMonitor) realtimeMonitorLoop() {
	ticker := time.NewTicker(chrm.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-chrm.stopChan:
			return
		case <-ticker.C:
			chrm.updateRealtimeMetrics()
		}
	}
}

func (chrm *CacheHitRateMonitor) updateRealtimeMetrics() {
	chrm.realtimeMetrics.mutex.Lock()
	defer chrm.realtimeMetrics.mutex.Unlock()

	// 计算最近的命中率指标
	globalStats := chrm.GetGlobalStats()
	chrm.realtimeMetrics.RecentHitRate = globalStats.OverallHitRate
	chrm.realtimeMetrics.RecentMissRate = globalStats.OverallMissRate

	// 分析趋势
	chrm.realtimeMetrics.HitRateTrend = chrm.analyzeTrend("hit_rate")
	chrm.realtimeMetrics.MissRateTrend = chrm.analyzeTrend("miss_rate")

	// 检查告警条件
	chrm.checkAlerts()

	chrm.realtimeMetrics.LastUpdated = time.Now()
}

func (chrm *CacheHitRateMonitor) analyzeTrend(metric string) string {
	// 简化的趋势分析实现
	// 实际应该基于历史数据进行趋势分析
	return "STABLE"
}

func (chrm *CacheHitRateMonitor) checkAlerts() {
	globalStats := chrm.GetGlobalStats()

	if globalStats.OverallHitRate < chrm.config.LowHitRateThreshold {
		chrm.realtimeMetrics.AlertLevel = "WARNING"
		chrm.realtimeMetrics.AlertMessage = fmt.Sprintf("Low cache hit rate detected: %.2f%%", globalStats.OverallHitRate*100)
	} else if globalStats.OverallMissRate > chrm.config.HighMissRateThreshold {
		chrm.realtimeMetrics.AlertLevel = "CRITICAL"
		chrm.realtimeMetrics.AlertMessage = fmt.Sprintf("High cache miss rate detected: %.2f%%", globalStats.OverallMissRate*100)
	} else {
		chrm.realtimeMetrics.AlertLevel = "NORMAL"
		chrm.realtimeMetrics.AlertMessage = ""
	}
}

func (chrm *CacheHitRateMonitor) hotspotAnalysisLoop() {
	ticker := time.NewTicker(chrm.config.MonitorInterval * 2) // 较低频率的热点分析
	defer ticker.Stop()

	for {
		select {
		case <-chrm.stopChan:
			return
		case <-ticker.C:
			chrm.performHotspotAnalysis()
		}
	}
}

func (chrm *CacheHitRateMonitor) performHotspotAnalysis() {
	// 实现热点分析逻辑
	// 识别高频访问但低命中率的键
	// 识别应该预热的查询模式
}

func (chrm *CacheHitRateMonitor) prewarmEngineLoop() {
	ticker := time.NewTicker(chrm.config.MonitorInterval * 5) // 预热检查频率更低
	defer ticker.Stop()

	for {
		select {
		case <-chrm.stopChan:
			return
		case <-ticker.C:
			chrm.evaluatePrewarmCandidates()
		}
	}
}

func (chrm *CacheHitRateMonitor) evaluatePrewarmCandidates() {
	// 实现预热候选评估逻辑
	// 基于热点分析结果和历史性能数据
	// 决定哪些查询需要预热
}

func (chrm *CacheHitRateMonitor) reportLoop() {
	ticker := time.NewTicker(chrm.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-chrm.stopChan:
			return
		case <-ticker.C:
			chrm.logCacheReport()
		}
	}
}

func (chrm *CacheHitRateMonitor) logCacheReport() {
	report := chrm.GenerateCacheReport()

	log.Printf("=== Cache Performance Report ===")
	log.Printf("Overall Hit Rate: %.2f%%", report.GlobalStats.OverallHitRate*100)
	log.Printf("Overall Miss Rate: %.2f%%", report.GlobalStats.OverallMissRate*100)

	if report.L1Stats != nil {
		log.Printf("L1 Cache Hit Rate: %.2f%%", report.L1Stats.HitRate*100)
	}
	if report.L2Stats != nil {
		log.Printf("L2 Cache Hit Rate: %.2f%%", report.L2Stats.HitRate*100)
	}

	log.Printf("Active Hotspots: %d", len(report.Hotspots))
	log.Printf("Alert Level: %s", report.RealTimeMetrics.AlertLevel)

	if len(report.Recommendations) > 0 {
		log.Printf("Recommendations:")
		for _, rec := range report.Recommendations {
			log.Printf("  - %s", rec)
		}
	}
}

// 辅助方法

func (chrm *CacheHitRateMonitor) recordTimeSeriesData(monitor *CacheLayerMonitor, hitRate, missRate float64, requestCount int64) {
	point := &TimeSeriesPoint{
		Timestamp:    time.Now(),
		HitRate:      hitRate,
		MissRate:     missRate,
		RequestCount: requestCount,
	}

	monitor.TimeSeriesData.Value = point
	monitor.TimeSeriesData = monitor.TimeSeriesData.Next()
}

func (chrm *CacheHitRateMonitor) updateKeyPatternStats(monitor *CacheLayerMonitor, key string, hit bool) {
	pattern := chrm.extractKeyPattern(key)

	patternStats, exists := monitor.KeyPatterns[pattern]
	if !exists {
		patternStats = &KeyPatternStats{
			Pattern:        pattern,
			LastAccessTime: time.Now(),
		}
		monitor.KeyPatterns[pattern] = patternStats
	}

	patternStats.Count++
	patternStats.LastAccessTime = time.Now()

	// 更新命中率
	if hit {
		patternStats.HitRate = (patternStats.HitRate*float64(patternStats.Count-1) + 1.0) / float64(patternStats.Count)
	} else {
		patternStats.HitRate = (patternStats.HitRate * float64(patternStats.Count-1)) / float64(patternStats.Count)
	}
}

func (chrm *CacheHitRateMonitor) extractKeyPattern(key string) string {
	// 简化的键模式提取
	// 实际应该使用更复杂的模式识别算法
	if strings.Contains(key, ":") {
		parts := strings.Split(key, ":")
		if len(parts) > 0 {
			return parts[0] + ":*"
		}
	}
	return "unknown"
}

func (chrm *CacheHitRateMonitor) analyzeResultSizeDistribution(size int64) {
	chrm.resultSetMonitor.mutex.Lock()
	defer chrm.resultSetMonitor.mutex.Unlock()

	// 按大小范围分类
	sizeRange := chrm.getSizeRange(size)
	chrm.resultSetMonitor.SizeDistribution[sizeRange]++
}

func (chrm *CacheHitRateMonitor) getSizeRange(size int64) string {
	if size < 1024 {
		return "< 1KB"
	} else if size < 10*1024 {
		return "1KB - 10KB"
	} else if size < 100*1024 {
		return "10KB - 100KB"
	} else if size < 1024*1024 {
		return "100KB - 1MB"
	} else {
		return "> 1MB"
	}
}

func (chrm *CacheHitRateMonitor) analyzeParameterPattern(params []any) string {
	// 简化的参数模式分析
	return fmt.Sprintf("params_%d", len(params))
}

func (chrm *CacheHitRateMonitor) updateParameterAnalysis(pattern string, params []any) {
	analyzer := chrm.queryMonitor.ParameterAnalyzer

	paramPattern, exists := analyzer.parameterPatterns[pattern]
	if !exists {
		paramPattern = &ParameterPattern{
			Pattern: pattern,
		}
		analyzer.parameterPatterns[pattern] = paramPattern
	}

	paramPattern.TotalOccurrences++
	// 这里应该实现更复杂的变异性分析
}

func (chrm *CacheHitRateMonitor) isQueryCacheable(query string, params []any) bool {
	// 简化的缓存能力评估
	upperQuery := strings.ToUpper(query)

	// 检查是否包含非确定性函数
	nonCacheableFunctions := []string{"NOW()", "RAND()", "UUID()"}
	for _, fn := range nonCacheableFunctions {
		if strings.Contains(upperQuery, fn) {
			return false
		}
	}

	return true
}

func (chrm *CacheHitRateMonitor) generateRecommendations() []string {
	recommendations := make([]string, 0)

	globalStats := chrm.GetGlobalStats()

	// 基于命中率生成建议
	if globalStats.OverallHitRate < 0.5 {
		recommendations = append(recommendations, "Consider increasing cache size or optimizing cache key strategies")
	}

	if globalStats.OverallMissRate > 0.5 {
		recommendations = append(recommendations, "High miss rate detected - review cache eviction policies")
	}

	// 基于热点分析生成建议
	hotspots := chrm.GetHotspots()
	if len(hotspots) > 0 {
		topHotspot := hotspots[0]
		if topHotspot.HitRate < 0.3 {
			recommendations = append(recommendations,
				fmt.Sprintf("Consider prewarming hotspot key pattern: %s", topHotspot.Key))
		}
	}

	return recommendations
}
