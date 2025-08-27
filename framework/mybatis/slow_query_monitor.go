// Package mybatis 慢查询监控和告警系统
//
// 核心功能：
// 1. 多维度慢查询检测（执行时间、CPU使用、内存消耗）
// 2. 智能阈值自适应调整
// 3. 查询模式识别和分类
// 4. 实时告警和批量报告
// 5. 查询优化建议生成
package mybatis

import (
	"fmt"
	"hash/crc32"
	"log"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// SlowQueryMonitor 慢查询监控器
type SlowQueryMonitor struct {
	// 基础配置
	config              *SlowQueryConfig
	
	// 检测器组件
	thresholdDetector   *ThresholdDetector
	patternAnalyzer     *QueryPatternAnalyzer
	alertManager        *AlertManager
	optimizationEngine  *OptimizationEngine
	
	// 数据存储
	slowQueries         map[string]*SlowQueryRecord
	queryPatterns       map[uint32]*QueryPattern
	alertHistory        []AlertRecord
	
	// 统计信息
	stats               *SlowQueryStats
	
	// 并发控制
	mutex               sync.RWMutex
	isRunning           int32
	stopChan            chan struct{}
	
	// 实时监控
	realtimeBuffer      chan *QueryExecution
	batchProcessor      *BatchProcessor
}

// SlowQueryConfig 慢查询监控配置
type SlowQueryConfig struct {
	// 基础阈值配置
	SlowQueryThreshold      time.Duration `yaml:"slow_query_threshold" json:"slow_query_threshold"`         // 慢查询时间阈值
	VerySlowThreshold       time.Duration `yaml:"very_slow_threshold" json:"very_slow_threshold"`           // 极慢查询阈值
	CPUThreshold            float64       `yaml:"cpu_threshold" json:"cpu_threshold"`                       // CPU使用率阈值
	MemoryThreshold         int64         `yaml:"memory_threshold" json:"memory_threshold"`                 // 内存使用阈值(bytes)
	
	// 自适应配置
	EnableAdaptiveThreshold bool          `yaml:"enable_adaptive_threshold" json:"enable_adaptive_threshold"` // 启用自适应阈值
	AdaptiveInterval        time.Duration `yaml:"adaptive_interval" json:"adaptive_interval"`                // 自适应调整间隔
	ThresholdAdjustmentRate float64       `yaml:"threshold_adjustment_rate" json:"threshold_adjustment_rate"` // 阈值调整速率
	
	// 检测配置
	PatternDetectionEnabled bool          `yaml:"pattern_detection_enabled" json:"pattern_detection_enabled"` // 启用模式检测
	SimilarityThreshold     float64       `yaml:"similarity_threshold" json:"similarity_threshold"`           // 查询相似度阈值
	FrequencyThreshold      int           `yaml:"frequency_threshold" json:"frequency_threshold"`             // 频率阈值
	
	// 告警配置
	EnableInstantAlert      bool          `yaml:"enable_instant_alert" json:"enable_instant_alert"`           // 启用即时告警
	EnableBatchAlert        bool          `yaml:"enable_batch_alert" json:"enable_batch_alert"`               // 启用批量告警
	BatchAlertInterval      time.Duration `yaml:"batch_alert_interval" json:"batch_alert_interval"`           // 批量告警间隔
	MaxAlertFrequency       int           `yaml:"max_alert_frequency" json:"max_alert_frequency"`             // 最大告警频率
	
	// 存储配置
	MaxSlowQueryRecords     int           `yaml:"max_slow_query_records" json:"max_slow_query_records"`       // 最大慢查询记录数
	RecordRetentionPeriod   time.Duration `yaml:"record_retention_period" json:"record_retention_period"`     // 记录保留期
	
	// 优化建议配置
	EnableOptimizationTips  bool          `yaml:"enable_optimization_tips" json:"enable_optimization_tips"`   // 启用优化建议
	AnalysisDepth           int           `yaml:"analysis_depth" json:"analysis_depth"`                       // 分析深度
}

// SlowQueryRecord 慢查询记录
type SlowQueryRecord struct {
	// 基础信息
	ID              string            `json:"id"`
	SQL             string            `json:"sql"`
	NormalizedSQL   string            `json:"normalized_sql"`
	Parameters      []any             `json:"parameters"`
	
	// 性能指标
	Duration        time.Duration     `json:"duration"`
	CPUTime         time.Duration     `json:"cpu_time"`
	MemoryUsed      int64             `json:"memory_used"`
	RowsExamined    int64             `json:"rows_examined"`
	RowsReturned    int64             `json:"rows_returned"`
	
	// 上下文信息
	Timestamp       time.Time         `json:"timestamp"`
	DatabaseName    string            `json:"database_name"`
	SessionID       string            `json:"session_id"`
	UserID          string            `json:"user_id"`
	
	// 分析结果
	QueryType       string            `json:"query_type"`         // SELECT, INSERT, UPDATE, DELETE
	TableNames      []string          `json:"table_names"`        // 涉及的表名
	PatternID       uint32            `json:"pattern_id"`         // 查询模式ID
	Severity        string            `json:"severity"`           // NORMAL, SLOW, VERY_SLOW, CRITICAL
	
	// 优化建议
	OptimizationTips []string         `json:"optimization_tips"`
	IndexSuggestions []string         `json:"index_suggestions"`
}

// QueryExecution 查询执行信息
type QueryExecution struct {
	SQL           string
	Parameters    []any
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	CPUUsageBefore runtime.MemStats
	CPUUsageAfter  runtime.MemStats
	MemoryBefore   int64
	MemoryAfter    int64
	Context       map[string]any
}

// QueryPattern 查询模式
type QueryPattern struct {
	ID              uint32            `json:"id"`
	NormalizedSQL   string            `json:"normalized_sql"`
	Template        string            `json:"template"`
	Frequency       int64             `json:"frequency"`
	AvgDuration     time.Duration     `json:"avg_duration"`
	MaxDuration     time.Duration     `json:"max_duration"`
	MinDuration     time.Duration     `json:"min_duration"`
	TotalDuration   time.Duration     `json:"total_duration"`
	FirstSeen       time.Time         `json:"first_seen"`
	LastSeen        time.Time         `json:"last_seen"`
	TableNames      []string          `json:"table_names"`
	QueryType       string            `json:"query_type"`
	Examples        []string          `json:"examples"`           // 示例查询
}

// SlowQueryStats 慢查询统计信息
type SlowQueryStats struct {
	// 基础统计
	TotalQueries        int64             `json:"total_queries"`
	SlowQueries         int64             `json:"slow_queries"`
	VerySlowQueries     int64             `json:"very_slow_queries"`
	CriticalQueries     int64             `json:"critical_queries"`
	
	// 时间统计
	TotalSlowTime       time.Duration     `json:"total_slow_time"`
	AvgSlowQueryTime    time.Duration     `json:"avg_slow_query_time"`
	MaxSlowQueryTime    time.Duration     `json:"max_slow_query_time"`
	
	// 模式统计
	UniquePatterns      int               `json:"unique_patterns"`
	TopPatterns         []string          `json:"top_patterns"`
	
	// 告警统计
	TotalAlerts         int64             `json:"total_alerts"`
	InstantAlerts       int64             `json:"instant_alerts"`
	BatchAlerts         int64             `json:"batch_alerts"`
	LastAlertTime       time.Time         `json:"last_alert_time"`
	
	// 优化统计
	OptimizationTipsGenerated int64       `json:"optimization_tips_generated"`
	IndexSuggestionsGenerated int64       `json:"index_suggestions_generated"`
	
	// 时间分布
	HourlyDistribution  map[int]int64     `json:"hourly_distribution"`
	DayOfWeekDist      map[time.Weekday]int64 `json:"day_of_week_distribution"`
	
	mutex               sync.RWMutex
}

// ThresholdDetector 阈值检测器
type ThresholdDetector struct {
	config              *SlowQueryConfig
	adaptiveThresholds  map[string]time.Duration
	performanceHistory  []PerformanceSnapshot
	mutex               sync.RWMutex
}

// QueryPatternAnalyzer 查询模式分析器
type QueryPatternAnalyzer struct {
	patterns            map[uint32]*QueryPattern
	normalizer          *SQLNormalizer
	similarityCache     map[string]float64
	mutex               sync.RWMutex
}

// AlertManager 告警管理器
type AlertManager struct {
	config              *SlowQueryConfig
	alertChannels       []AlertChannel
	alertHistory        []AlertRecord
	alertFrequencyCount map[string]int
	lastAlertTime       map[string]time.Time
	mutex               sync.RWMutex
}

// OptimizationEngine 优化引擎
type OptimizationEngine struct {
	rules               []OptimizationRule
	indexAnalyzer       *IndexAnalyzer
	queryAnalyzer       *QueryAnalyzer
}

// PerformanceSnapshot 性能快照
type PerformanceSnapshot struct {
	Timestamp       time.Time     `json:"timestamp"`
	AvgQueryTime    time.Duration `json:"avg_query_time"`
	QueryCount      int64         `json:"query_count"`
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     int64         `json:"memory_usage"`
}

// AlertRecord 告警记录
type AlertRecord struct {
	ID              string            `json:"id"`
	Timestamp       time.Time         `json:"timestamp"`
	Type            string            `json:"type"`           // INSTANT, BATCH
	Severity        string            `json:"severity"`       // NORMAL, WARNING, CRITICAL
	Message         string            `json:"message"`
	SlowQueryID     string            `json:"slow_query_id"`
	Details         map[string]any    `json:"details"`
	Resolved        bool              `json:"resolved"`
	ResolvedAt      time.Time         `json:"resolved_at"`
}

// AlertChannel 告警通道接口
type AlertChannel interface {
	SendAlert(alert *AlertRecord) error
	GetChannelType() string
	IsEnabled() bool
}

// OptimizationRule 优化规则接口
type OptimizationRule interface {
	Analyze(record *SlowQueryRecord) []string
	GetRuleType() string
	Priority() int
}

// SQLNormalizer SQL标准化器
type SQLNormalizer struct {
	paramPattern    *regexp.Regexp
	numberPattern   *regexp.Regexp
	stringPattern   *regexp.Regexp
	whitespacePattern *regexp.Regexp
}

// IndexAnalyzer 索引分析器
type IndexAnalyzer struct {
	db              *gorm.DB
	indexCache      map[string][]IndexInfo
	mutex           sync.RWMutex
}

// QueryAnalyzer 查询分析器
type QueryAnalyzer struct {
	explainCache    map[string]*ExplainResult
	mutex           sync.RWMutex
}

// IndexInfo 索引信息
type IndexInfo struct {
	TableName   string   `json:"table_name"`
	IndexName   string   `json:"index_name"`
	Columns     []string `json:"columns"`
	IsUnique    bool     `json:"is_unique"`
	IsPrimary   bool     `json:"is_primary"`
}

// ExplainResult 执行计划结果
type ExplainResult struct {
	Rows        int64    `json:"rows"`
	Cost        float64  `json:"cost"`
	UsesIndex   bool     `json:"uses_index"`
	IndexNames  []string `json:"index_names"`
	TableScans  bool     `json:"table_scans"`
}

// BatchProcessor 批量处理器
type BatchProcessor struct {
	buffer          []*QueryExecution
	batchSize       int
	flushInterval   time.Duration
	processor       func([]*QueryExecution)
	mutex           sync.Mutex
	stopChan        chan struct{}
	isRunning       int32
}

// DefaultSlowQueryConfig 默认配置
func DefaultSlowQueryConfig() *SlowQueryConfig {
	return &SlowQueryConfig{
		SlowQueryThreshold:        100 * time.Millisecond,
		VerySlowThreshold:         500 * time.Millisecond,
		CPUThreshold:              80.0, // 80% CPU
		MemoryThreshold:           50 * 1024 * 1024, // 50MB
		EnableAdaptiveThreshold:   true,
		AdaptiveInterval:          5 * time.Minute,
		ThresholdAdjustmentRate:   0.1,
		PatternDetectionEnabled:   true,
		SimilarityThreshold:       0.85,
		FrequencyThreshold:        10,
		EnableInstantAlert:        true,
		EnableBatchAlert:          true,
		BatchAlertInterval:        10 * time.Minute,
		MaxAlertFrequency:         5,
		MaxSlowQueryRecords:       10000,
		RecordRetentionPeriod:     7 * 24 * time.Hour,
		EnableOptimizationTips:    true,
		AnalysisDepth:             3,
	}
}

// NewSlowQueryMonitor 创建慢查询监控器
func NewSlowQueryMonitor(config *SlowQueryConfig) *SlowQueryMonitor {
	if config == nil {
		config = DefaultSlowQueryConfig()
	}
	
	monitor := &SlowQueryMonitor{
		config:          config,
		slowQueries:     make(map[string]*SlowQueryRecord),
		queryPatterns:   make(map[uint32]*QueryPattern),
		alertHistory:    make([]AlertRecord, 0, 1000),
		stats:           &SlowQueryStats{
			HourlyDistribution: make(map[int]int64),
			DayOfWeekDist:     make(map[time.Weekday]int64),
		},
		stopChan:        make(chan struct{}),
		realtimeBuffer:  make(chan *QueryExecution, 1000),
	}
	
	// 初始化组件
	monitor.initComponents()
	
	return monitor
}

// initComponents 初始化组件
func (sqm *SlowQueryMonitor) initComponents() {
	// 初始化阈值检测器
	sqm.thresholdDetector = &ThresholdDetector{
		config:             sqm.config,
		adaptiveThresholds: make(map[string]time.Duration),
		performanceHistory: make([]PerformanceSnapshot, 0, 100),
	}
	
	// 初始化模式分析器
	sqm.patternAnalyzer = &QueryPatternAnalyzer{
		patterns:        make(map[uint32]*QueryPattern),
		normalizer:      NewSQLNormalizer(),
		similarityCache: make(map[string]float64),
	}
	
	// 初始化告警管理器
	sqm.alertManager = &AlertManager{
		config:              sqm.config,
		alertChannels:       make([]AlertChannel, 0),
		alertHistory:        make([]AlertRecord, 0, 1000),
		alertFrequencyCount: make(map[string]int),
		lastAlertTime:       make(map[string]time.Time),
	}
	
	// 初始化优化引擎
	sqm.optimizationEngine = &OptimizationEngine{
		rules:         []OptimizationRule{},
		indexAnalyzer: &IndexAnalyzer{indexCache: make(map[string][]IndexInfo)},
		queryAnalyzer: &QueryAnalyzer{explainCache: make(map[string]*ExplainResult)},
	}
	
	// 初始化批量处理器
	sqm.batchProcessor = &BatchProcessor{
		buffer:        make([]*QueryExecution, 0, 100),
		batchSize:     50,
		flushInterval: 30 * time.Second,
		processor:     sqm.processBatch,
		stopChan:      make(chan struct{}),
	}
}

// Start 启动慢查询监控
func (sqm *SlowQueryMonitor) Start() error {
	if !atomic.CompareAndSwapInt32(&sqm.isRunning, 0, 1) {
		return fmt.Errorf("slow query monitor is already running")
	}
	
	// 启动实时监控协程
	go sqm.realtimeMonitorLoop()
	
	// 启动批量处理器
	sqm.batchProcessor.Start()
	
	// 启动自适应阈值调整
	if sqm.config.EnableAdaptiveThreshold {
		go sqm.adaptiveThresholdLoop()
	}
	
	// 启动批量告警
	if sqm.config.EnableBatchAlert {
		go sqm.batchAlertLoop()
	}
	
	// 启动数据清理
	go sqm.cleanupLoop()
	
	log.Println("[SlowQueryMonitor] Slow query monitor started")
	return nil
}

// Stop 停止慢查询监控
func (sqm *SlowQueryMonitor) Stop() error {
	if !atomic.CompareAndSwapInt32(&sqm.isRunning, 1, 0) {
		return fmt.Errorf("slow query monitor is not running")
	}
	
	close(sqm.stopChan)
	sqm.batchProcessor.Stop()
	
	log.Println("[SlowQueryMonitor] Slow query monitor stopped")
	return nil
}

// RecordQuery 记录查询执行信息
func (sqm *SlowQueryMonitor) RecordQuery(execution *QueryExecution) {
	if atomic.LoadInt32(&sqm.isRunning) == 0 {
		return
	}
	
	// 发送到实时缓冲区
	select {
	case sqm.realtimeBuffer <- execution:
	default:
		// 缓冲区满了，直接处理
		go sqm.processQueryExecution(execution)
	}
}

// realtimeMonitorLoop 实时监控循环
func (sqm *SlowQueryMonitor) realtimeMonitorLoop() {
	for {
		select {
		case <-sqm.stopChan:
			return
		case execution := <-sqm.realtimeBuffer:
			sqm.processQueryExecution(execution)
		}
	}
}

// processQueryExecution 处理查询执行
func (sqm *SlowQueryMonitor) processQueryExecution(execution *QueryExecution) {
	// 更新基础统计
	atomic.AddInt64(&sqm.stats.TotalQueries, 1)
	
	// 检查是否为慢查询
	if !sqm.isSlowQuery(execution) {
		return
	}
	
	// 创建慢查询记录
	record := sqm.createSlowQueryRecord(execution)
	
	// 分析查询模式
	if sqm.config.PatternDetectionEnabled {
		sqm.analyzeQueryPattern(record)
	}
	
	// 生成优化建议
	if sqm.config.EnableOptimizationTips {
		sqm.generateOptimizationTips(record)
	}
	
	// 存储记录
	sqm.storeSlowQueryRecord(record)
	
	// 发送即时告警
	if sqm.config.EnableInstantAlert {
		sqm.sendInstantAlert(record)
	}
	
	// 更新统计
	sqm.updateStats(record)
}

// isSlowQuery 判断是否为慢查询
func (sqm *SlowQueryMonitor) isSlowQuery(execution *QueryExecution) bool {
	// 时间阈值检查
	threshold := sqm.thresholdDetector.getThreshold(execution.SQL)
	if execution.Duration >= threshold {
		return true
	}
	
	// CPU使用检查
	if sqm.config.CPUThreshold > 0 {
		// 简化的CPU检查逻辑
		cpuUsage := sqm.calculateCPUUsage(execution)
		if cpuUsage > sqm.config.CPUThreshold {
			return true
		}
	}
	
	// 内存使用检查
	if sqm.config.MemoryThreshold > 0 {
		memoryUsed := execution.MemoryAfter - execution.MemoryBefore
		if memoryUsed > sqm.config.MemoryThreshold {
			return true
		}
	}
	
	return false
}

// getThreshold 获取查询阈值
func (td *ThresholdDetector) getThreshold(sql string) time.Duration {
	if !td.config.EnableAdaptiveThreshold {
		return td.config.SlowQueryThreshold
	}
	
	td.mutex.RLock()
	defer td.mutex.RUnlock()
	
	// 标准化SQL作为键
	normalizedSQL := strings.ToUpper(strings.TrimSpace(sql))
	if threshold, exists := td.adaptiveThresholds[normalizedSQL]; exists {
		return threshold
	}
	
	return td.config.SlowQueryThreshold
}

// calculateCPUUsage 计算CPU使用率
func (sqm *SlowQueryMonitor) calculateCPUUsage(execution *QueryExecution) float64 {
	// 简化的CPU使用计算
	beforeTotalTime := time.Duration(execution.CPUUsageBefore.GCCPUFraction) * execution.Duration
	afterTotalTime := time.Duration(execution.CPUUsageAfter.GCCPUFraction) * execution.Duration
	
	if execution.Duration.Nanoseconds() == 0 {
		return 0
	}
	
	return float64(afterTotalTime-beforeTotalTime) / float64(execution.Duration) * 100
}

// createSlowQueryRecord 创建慢查询记录
func (sqm *SlowQueryMonitor) createSlowQueryRecord(execution *QueryExecution) *SlowQueryRecord {
	normalizedSQL := sqm.patternAnalyzer.normalizer.Normalize(execution.SQL)
	
	record := &SlowQueryRecord{
		ID:              sqm.generateRecordID(),
		SQL:             execution.SQL,
		NormalizedSQL:   normalizedSQL,
		Parameters:      execution.Parameters,
		Duration:        execution.Duration,
		MemoryUsed:      execution.MemoryAfter - execution.MemoryBefore,
		Timestamp:       execution.StartTime,
		QueryType:       sqm.extractQueryType(execution.SQL),
		TableNames:      sqm.extractTableNames(execution.SQL),
		Severity:        sqm.calculateSeverity(execution.Duration),
	}
	
	return record
}

// analyzeQueryPattern 分析查询模式
func (sqm *SlowQueryMonitor) analyzeQueryPattern(record *SlowQueryRecord) {
	patternID := sqm.patternAnalyzer.getOrCreatePattern(record)
	record.PatternID = patternID
	
	// 更新模式统计
	sqm.patternAnalyzer.updatePattern(patternID, record)
}

// generateOptimizationTips 生成优化建议
func (sqm *SlowQueryMonitor) generateOptimizationTips(record *SlowQueryRecord) {
	tips := make([]string, 0)
	indexSuggestions := make([]string, 0)
	
	// 应用优化规则
	for _, rule := range sqm.optimizationEngine.rules {
		ruleTips := rule.Analyze(record)
		tips = append(tips, ruleTips...)
	}
	
	// 生成索引建议
	indexSuggestions = sqm.optimizationEngine.generateIndexSuggestions(record)
	
	record.OptimizationTips = tips
	record.IndexSuggestions = indexSuggestions
	
	// 更新统计
	if len(tips) > 0 {
		atomic.AddInt64(&sqm.stats.OptimizationTipsGenerated, int64(len(tips)))
	}
	if len(indexSuggestions) > 0 {
		atomic.AddInt64(&sqm.stats.IndexSuggestionsGenerated, int64(len(indexSuggestions)))
	}
}

// storeSlowQueryRecord 存储慢查询记录
func (sqm *SlowQueryMonitor) storeSlowQueryRecord(record *SlowQueryRecord) {
	sqm.mutex.Lock()
	defer sqm.mutex.Unlock()
	
	// 检查存储上限
	if len(sqm.slowQueries) >= sqm.config.MaxSlowQueryRecords {
		sqm.evictOldestRecord()
	}
	
	sqm.slowQueries[record.ID] = record
}

// sendInstantAlert 发送即时告警
func (sqm *SlowQueryMonitor) sendInstantAlert(record *SlowQueryRecord) {
	if !sqm.shouldAlert(record) {
		return
	}
	
	alert := &AlertRecord{
		ID:          sqm.generateAlertID(),
		Timestamp:   time.Now(),
		Type:        "INSTANT",
		Severity:    record.Severity,
		Message:     fmt.Sprintf("Slow query detected: %s (Duration: %v)", record.QueryType, record.Duration),
		SlowQueryID: record.ID,
		Details:     map[string]any{
			"sql":           record.SQL,
			"duration":      record.Duration.String(),
			"memory_used":   record.MemoryUsed,
			"table_names":   record.TableNames,
		},
	}
	
	sqm.alertManager.sendAlert(alert)
}

// updateStats 更新统计信息
func (sqm *SlowQueryMonitor) updateStats(record *SlowQueryRecord) {
	sqm.stats.mutex.Lock()
	defer sqm.stats.mutex.Unlock()
	
	// 基础统计
	atomic.AddInt64(&sqm.stats.SlowQueries, 1)
	
	switch record.Severity {
	case "VERY_SLOW":
		atomic.AddInt64(&sqm.stats.VerySlowQueries, 1)
	case "CRITICAL":
		atomic.AddInt64(&sqm.stats.CriticalQueries, 1)
	}
	
	// 时间统计
	sqm.stats.TotalSlowTime += record.Duration
	if sqm.stats.SlowQueries > 0 {
		sqm.stats.AvgSlowQueryTime = sqm.stats.TotalSlowTime / time.Duration(sqm.stats.SlowQueries)
	}
	if record.Duration > sqm.stats.MaxSlowQueryTime {
		sqm.stats.MaxSlowQueryTime = record.Duration
	}
	
	// 时间分布统计
	hour := record.Timestamp.Hour()
	sqm.stats.HourlyDistribution[hour]++
	
	dayOfWeek := record.Timestamp.Weekday()
	sqm.stats.DayOfWeekDist[dayOfWeek]++
}

// 辅助方法实现

func (sqm *SlowQueryMonitor) generateRecordID() string {
	return fmt.Sprintf("slow_%d_%d", time.Now().UnixNano(), atomic.AddInt64(&sqm.stats.TotalQueries, 0))
}

func (sqm *SlowQueryMonitor) generateAlertID() string {
	return fmt.Sprintf("alert_%d_%d", time.Now().UnixNano(), atomic.AddInt64(&sqm.stats.TotalAlerts, 1))
}

func (sqm *SlowQueryMonitor) extractQueryType(sql string) string {
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))
	if strings.HasPrefix(upperSQL, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(upperSQL, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(upperSQL, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(upperSQL, "DELETE") {
		return "DELETE"
	}
	return "UNKNOWN"
}

func (sqm *SlowQueryMonitor) extractTableNames(sql string) []string {
	// 简化的表名提取逻辑
	tables := make([]string, 0)
	upperSQL := strings.ToUpper(sql)
	
	// 查找FROM子句
	fromIndex := strings.Index(upperSQL, " FROM ")
	if fromIndex > 0 {
		remaining := sql[fromIndex+6:]
		parts := strings.Fields(remaining)
		if len(parts) > 0 {
			tables = append(tables, strings.Trim(parts[0], ","))
		}
	}
	
	return tables
}

func (sqm *SlowQueryMonitor) calculateSeverity(duration time.Duration) string {
	if duration >= sqm.config.VerySlowThreshold*2 {
		return "CRITICAL"
	} else if duration >= sqm.config.VerySlowThreshold {
		return "VERY_SLOW"
	} else if duration >= sqm.config.SlowQueryThreshold {
		return "SLOW"
	}
	return "NORMAL"
}

func (sqm *SlowQueryMonitor) shouldAlert(record *SlowQueryRecord) bool {
	// 检查告警频率限制
	key := record.NormalizedSQL
	
	sqm.alertManager.mutex.Lock()
	defer sqm.alertManager.mutex.Unlock()
	
	now := time.Now()
	if lastTime, exists := sqm.alertManager.lastAlertTime[key]; exists {
		if now.Sub(lastTime) < time.Minute { // 1分钟内不重复告警
			return false
		}
	}
	
	count := sqm.alertManager.alertFrequencyCount[key]
	if count >= sqm.config.MaxAlertFrequency {
		return false
	}
	
	sqm.alertManager.lastAlertTime[key] = now
	sqm.alertManager.alertFrequencyCount[key]++
	
	return true
}

func (sqm *SlowQueryMonitor) evictOldestRecord() {
	var oldestID string
	var oldestTime time.Time = time.Now()
	
	for id, record := range sqm.slowQueries {
		if record.Timestamp.Before(oldestTime) {
			oldestTime = record.Timestamp
			oldestID = id
		}
	}
	
	if oldestID != "" {
		delete(sqm.slowQueries, oldestID)
	}
}

// 更多辅助方法和循环函数...

func (sqm *SlowQueryMonitor) adaptiveThresholdLoop() {
	ticker := time.NewTicker(sqm.config.AdaptiveInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sqm.stopChan:
			return
		case <-ticker.C:
			sqm.adjustAdaptiveThresholds()
		}
	}
}

func (sqm *SlowQueryMonitor) adjustAdaptiveThresholds() {
	// 自适应阈值调整逻辑
	sqm.thresholdDetector.mutex.Lock()
	defer sqm.thresholdDetector.mutex.Unlock()
	
	// 基于历史性能数据调整阈值
	// 实现省略...
}

func (sqm *SlowQueryMonitor) batchAlertLoop() {
	ticker := time.NewTicker(sqm.config.BatchAlertInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sqm.stopChan:
			return
		case <-ticker.C:
			sqm.sendBatchAlerts()
		}
	}
}

func (sqm *SlowQueryMonitor) sendBatchAlerts() {
	// 批量告警逻辑
	// 实现省略...
}

func (sqm *SlowQueryMonitor) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-sqm.stopChan:
			return
		case <-ticker.C:
			sqm.cleanupOldRecords()
		}
	}
}

func (sqm *SlowQueryMonitor) cleanupOldRecords() {
	sqm.mutex.Lock()
	defer sqm.mutex.Unlock()
	
	cutoff := time.Now().Add(-sqm.config.RecordRetentionPeriod)
	
	for id, record := range sqm.slowQueries {
		if record.Timestamp.Before(cutoff) {
			delete(sqm.slowQueries, id)
		}
	}
}

// GetSlowQueryStats 获取慢查询统计信息
func (sqm *SlowQueryMonitor) GetSlowQueryStats() *SlowQueryStats {
	sqm.stats.mutex.RLock()
	defer sqm.stats.mutex.RUnlock()
	
	// 返回统计信息副本
	statsCopy := *sqm.stats
	statsCopy.HourlyDistribution = make(map[int]int64)
	statsCopy.DayOfWeekDist = make(map[time.Weekday]int64)
	
	for k, v := range sqm.stats.HourlyDistribution {
		statsCopy.HourlyDistribution[k] = v
	}
	for k, v := range sqm.stats.DayOfWeekDist {
		statsCopy.DayOfWeekDist[k] = v
	}
	
	return &statsCopy
}

// GetSlowQueries 获取慢查询记录
func (sqm *SlowQueryMonitor) GetSlowQueries(limit int, severity string) []*SlowQueryRecord {
	sqm.mutex.RLock()
	defer sqm.mutex.RUnlock()
	
	records := make([]*SlowQueryRecord, 0, limit)
	
	for _, record := range sqm.slowQueries {
		if severity != "" && record.Severity != severity {
			continue
		}
		records = append(records, record)
		if len(records) >= limit {
			break
		}
	}
	
	// 按时间排序
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})
	
	return records
}

// GetQueryPatterns 获取查询模式
func (sqm *SlowQueryMonitor) GetQueryPatterns(limit int) []*QueryPattern {
	sqm.patternAnalyzer.mutex.RLock()
	defer sqm.patternAnalyzer.mutex.RUnlock()
	
	patterns := make([]*QueryPattern, 0, len(sqm.queryPatterns))
	for _, pattern := range sqm.queryPatterns {
		patterns = append(patterns, pattern)
	}
	
	// 按频率排序
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Frequency > patterns[j].Frequency
	})
	
	if limit > 0 && len(patterns) > limit {
		patterns = patterns[:limit]
	}
	
	return patterns
}

// NewSQLNormalizer 创建SQL标准化器
func NewSQLNormalizer() *SQLNormalizer {
	return &SQLNormalizer{
		paramPattern:      regexp.MustCompile(`\$\d+|\?`),
		numberPattern:     regexp.MustCompile(`\b\d+\b`),
		stringPattern:     regexp.MustCompile(`'[^']*'|"[^"]*"`),
		whitespacePattern: regexp.MustCompile(`\s+`),
	}
}

// Normalize 标准化SQL
func (sn *SQLNormalizer) Normalize(sql string) string {
	// 替换参数占位符
	normalized := sn.paramPattern.ReplaceAllString(sql, "?")
	
	// 替换数字字面量
	normalized = sn.numberPattern.ReplaceAllString(normalized, "?")
	
	// 替换字符串字面量
	normalized = sn.stringPattern.ReplaceAllString(normalized, "?")
	
	// 标准化空白字符
	normalized = sn.whitespacePattern.ReplaceAllString(strings.TrimSpace(normalized), " ")
	
	return strings.ToUpper(normalized)
}

// getOrCreatePattern 获取或创建查询模式
func (qpa *QueryPatternAnalyzer) getOrCreatePattern(record *SlowQueryRecord) uint32 {
	normalizedSQL := record.NormalizedSQL
	patternID := crc32.ChecksumIEEE([]byte(normalizedSQL))
	
	qpa.mutex.Lock()
	defer qpa.mutex.Unlock()
	
	if _, exists := qpa.patterns[patternID]; !exists {
		qpa.patterns[patternID] = &QueryPattern{
			ID:            patternID,
			NormalizedSQL: normalizedSQL,
			Template:      normalizedSQL,
			Frequency:     0,
			FirstSeen:     record.Timestamp,
			LastSeen:      record.Timestamp,
			TableNames:    record.TableNames,
			QueryType:     record.QueryType,
			Examples:      make([]string, 0, 5),
		}
	}
	
	return patternID
}

// updatePattern 更新查询模式
func (qpa *QueryPatternAnalyzer) updatePattern(patternID uint32, record *SlowQueryRecord) {
	qpa.mutex.Lock()
	defer qpa.mutex.Unlock()
	
	pattern := qpa.patterns[patternID]
	if pattern == nil {
		return
	}
	
	pattern.Frequency++
	pattern.LastSeen = record.Timestamp
	pattern.TotalDuration += record.Duration
	pattern.AvgDuration = pattern.TotalDuration / time.Duration(pattern.Frequency)
	
	if record.Duration > pattern.MaxDuration {
		pattern.MaxDuration = record.Duration
	}
	if pattern.MinDuration == 0 || record.Duration < pattern.MinDuration {
		pattern.MinDuration = record.Duration
	}
	
	// 添加示例（最多保留5个）
	if len(pattern.Examples) < 5 {
		pattern.Examples = append(pattern.Examples, record.SQL)
	}
}

// sendAlert 发送告警
func (am *AlertManager) sendAlert(alert *AlertRecord) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	// 存储告警记录
	am.alertHistory = append(am.alertHistory, *alert)
	
	// 发送到各个告警通道
	for _, channel := range am.alertChannels {
		if channel.IsEnabled() {
			go func(ch AlertChannel) {
				if err := ch.SendAlert(alert); err != nil {
					log.Printf("[AlertManager] Failed to send alert via %s: %v", ch.GetChannelType(), err)
				}
			}(channel)
		}
	}
}

// generateIndexSuggestions 生成索引建议
func (oe *OptimizationEngine) generateIndexSuggestions(record *SlowQueryRecord) []string {
	suggestions := make([]string, 0)
	
	// 基于查询类型生成建议
	switch record.QueryType {
	case "SELECT":
		suggestions = append(suggestions, oe.analyzeSelectQuery(record)...)
	case "UPDATE", "DELETE":
		suggestions = append(suggestions, oe.analyzeModifyQuery(record)...)
	}
	
	return suggestions
}

func (oe *OptimizationEngine) analyzeSelectQuery(record *SlowQueryRecord) []string {
	suggestions := make([]string, 0)
	
	// 分析WHERE条件
	if strings.Contains(strings.ToUpper(record.SQL), " WHERE ") {
		suggestions = append(suggestions, "Consider adding indexes on columns used in WHERE clause")
	}
	
	// 分析ORDER BY
	if strings.Contains(strings.ToUpper(record.SQL), " ORDER BY ") {
		suggestions = append(suggestions, "Consider adding indexes on columns used in ORDER BY clause")
	}
	
	// 分析JOIN
	if strings.Contains(strings.ToUpper(record.SQL), " JOIN ") {
		suggestions = append(suggestions, "Consider adding indexes on join columns")
	}
	
	return suggestions
}

func (oe *OptimizationEngine) analyzeModifyQuery(record *SlowQueryRecord) []string {
	suggestions := make([]string, 0)
	
	if strings.Contains(strings.ToUpper(record.SQL), " WHERE ") {
		suggestions = append(suggestions, "Consider adding indexes on columns used in WHERE clause for faster row location")
	}
	
	return suggestions
}

// BatchProcessor 方法实现

// Start 启动批量处理器
func (bp *BatchProcessor) Start() {
	if !atomic.CompareAndSwapInt32(&bp.isRunning, 0, 1) {
		return
	}
	
	go bp.flushLoop()
}

// Stop 停止批量处理器
func (bp *BatchProcessor) Stop() {
	if !atomic.CompareAndSwapInt32(&bp.isRunning, 1, 0) {
		return
	}
	
	close(bp.stopChan)
	
	// 处理剩余的批次
	bp.mutex.Lock()
	if len(bp.buffer) > 0 {
		bp.processor(bp.buffer)
		bp.buffer = bp.buffer[:0]
	}
	bp.mutex.Unlock()
}

// flushLoop 定期刷新循环
func (bp *BatchProcessor) flushLoop() {
	ticker := time.NewTicker(bp.flushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-bp.stopChan:
			return
		case <-ticker.C:
			bp.flush()
		}
	}
}

// flush 刷新缓冲区
func (bp *BatchProcessor) flush() {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	
	if len(bp.buffer) == 0 {
		return
	}
	
	// 复制缓冲区
	batch := make([]*QueryExecution, len(bp.buffer))
	copy(batch, bp.buffer)
	
	// 清空缓冲区
	bp.buffer = bp.buffer[:0]
	
	// 异步处理批次
	go bp.processor(batch)
}

// processBatch 处理批次
func (sqm *SlowQueryMonitor) processBatch(batch []*QueryExecution) {
	for _, execution := range batch {
		sqm.processQueryExecution(execution)
	}
}