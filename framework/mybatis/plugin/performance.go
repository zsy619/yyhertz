// Package plugin 性能监控插件实现
//
// 提供SQL执行性能监控、慢查询检测、统计信息收集等功能
package plugin

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// PerformancePlugin 性能监控插件
type PerformancePlugin struct {
	*BasePlugin
	slowQueryThreshold time.Duration // 慢查询阈值
	enableMetrics      bool          // 是否启用指标收集
	metrics            *PerformanceMetrics
	slowQueries        *SlowQueryRecorder
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// 查询统计
	TotalQueries  int64         // 总查询数
	SlowQueries   int64         // 慢查询数
	FailedQueries int64         // 失败查询数
	TotalTime     time.Duration // 总执行时间
	MaxTime       time.Duration // 最大执行时间
	MinTime       time.Duration // 最小执行时间
	AvgTime       time.Duration // 平均执行时间

	// 并发统计
	ConcurrentQueries int64 // 当前并发查询数
	MaxConcurrent     int64 // 最大并发数

	// 时间分布
	TimeDistribution map[string]int64 // 时间分布统计

	mutex sync.RWMutex
}

// SlowQueryRecorder 慢查询记录器
type SlowQueryRecorder struct {
	records []SlowQueryRecord
	maxSize int
	mutex   sync.RWMutex
}

// SlowQueryRecord 慢查询记录
type SlowQueryRecord struct {
	SQL           string         // SQL语句
	Parameters    []any          // 参数
	ExecutionTime time.Duration  // 执行时间
	Timestamp     time.Time      // 执行时间戳
	Method        string         // 调用方法
	Error         error          // 错误信息
	Context       map[string]any // 上下文信息
}

// QueryExecutionInfo 查询执行信息
type QueryExecutionInfo struct {
	SQL           string
	Parameters    []any
	StartTime     time.Time
	EndTime       time.Time
	ExecutionTime time.Duration
	Success       bool
	Error         error
	RowsAffected  int64
	Method        string
}

// NewPerformancePlugin 创建性能监控插件
func NewPerformancePlugin() *PerformancePlugin {
	plugin := &PerformancePlugin{
		BasePlugin:         NewBasePlugin("performance", 2),
		slowQueryThreshold: 1000 * time.Millisecond, // 默认1秒
		enableMetrics:      true,
		metrics:            NewPerformanceMetrics(),
		slowQueries:        NewSlowQueryRecorder(100), // 最多记录100条慢查询
	}
	return plugin
}

// NewPerformanceMetrics 创建性能指标
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		MinTime:          time.Duration(^uint64(0) >> 1), // 最大值作为初始最小值
		TimeDistribution: make(map[string]int64),
	}
}

// NewSlowQueryRecorder 创建慢查询记录器
func NewSlowQueryRecorder(maxSize int) *SlowQueryRecorder {
	return &SlowQueryRecorder{
		records: make([]SlowQueryRecord, 0, maxSize),
		maxSize: maxSize,
	}
}

// Intercept 拦截方法调用
func (plugin *PerformancePlugin) Intercept(invocation *Invocation) (any, error) {
	if !plugin.enableMetrics {
		return invocation.Proceed()
	}

	// 开始监控
	startTime := time.Now()
	atomic.AddInt64(&plugin.metrics.ConcurrentQueries, 1)

	// 更新最大并发数
	current := atomic.LoadInt64(&plugin.metrics.ConcurrentQueries)
	for {
		max := atomic.LoadInt64(&plugin.metrics.MaxConcurrent)
		if current <= max || atomic.CompareAndSwapInt64(&plugin.metrics.MaxConcurrent, max, current) {
			break
		}
	}

	// 执行原方法
	result, err := invocation.Proceed()

	// 结束监控
	endTime := time.Now()
	executionTime := endTime.Sub(startTime)
	atomic.AddInt64(&plugin.metrics.ConcurrentQueries, -1)

	// 记录执行信息
	execInfo := &QueryExecutionInfo{
		SQL:           plugin.extractSQL(invocation),
		Parameters:    plugin.extractParameters(invocation),
		StartTime:     startTime,
		EndTime:       endTime,
		ExecutionTime: executionTime,
		Success:       err == nil,
		Error:         err,
		Method:        invocation.Method.Name,
	}

	// 更新统计信息
	plugin.updateMetrics(execInfo)

	// 记录慢查询
	if executionTime >= plugin.slowQueryThreshold {
		plugin.recordSlowQuery(execInfo)
	}

	return result, err
}

// Plugin 包装目标对象
func (plugin *PerformancePlugin) Plugin(target any) any {
	return target
}

// SetProperties 设置插件属性
func (plugin *PerformancePlugin) SetProperties(properties map[string]any) {
	plugin.BasePlugin.SetProperties(properties)

	if threshold := plugin.GetPropertyInt("slowQueryThreshold", 1000); threshold > 0 {
		plugin.slowQueryThreshold = time.Duration(threshold) * time.Millisecond
	}

	plugin.enableMetrics = plugin.GetPropertyBool("enableMetrics", true)
}

// extractSQL 提取SQL语句
func (plugin *PerformancePlugin) extractSQL(invocation *Invocation) string {
	// 从调用参数中提取SQL
	for _, arg := range invocation.Args {
		if sql, ok := arg.(string); ok && plugin.looksLikeSQL(sql) {
			return sql
		}
	}

	// 从属性中提取
	if sql, exists := invocation.Properties["sql"]; exists {
		if sqlStr, ok := sql.(string); ok {
			return sqlStr
		}
	}

	return fmt.Sprintf("Method: %s", invocation.Method.Name)
}

// extractParameters 提取参数
func (plugin *PerformancePlugin) extractParameters(invocation *Invocation) []any {
	params := make([]any, 0)

	for _, arg := range invocation.Args {
		// 跳过SQL字符串
		if sql, ok := arg.(string); ok && plugin.looksLikeSQL(sql) {
			continue
		}
		params = append(params, arg)
	}

	return params
}

// looksLikeSQL 判断字符串是否像SQL
func (plugin *PerformancePlugin) looksLikeSQL(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	return strings.HasPrefix(s, "SELECT") ||
		strings.HasPrefix(s, "INSERT") ||
		strings.HasPrefix(s, "UPDATE") ||
		strings.HasPrefix(s, "DELETE") ||
		strings.HasPrefix(s, "CREATE") ||
		strings.HasPrefix(s, "DROP") ||
		strings.HasPrefix(s, "ALTER")
}

// updateMetrics 更新统计信息
func (plugin *PerformancePlugin) updateMetrics(execInfo *QueryExecutionInfo) {
	plugin.metrics.mutex.Lock()
	defer plugin.metrics.mutex.Unlock()

	// 更新基本统计
	atomic.AddInt64(&plugin.metrics.TotalQueries, 1)
	plugin.metrics.TotalTime += execInfo.ExecutionTime

	if !execInfo.Success {
		atomic.AddInt64(&plugin.metrics.FailedQueries, 1)
	}

	// 更新最大最小时间
	if execInfo.ExecutionTime > plugin.metrics.MaxTime {
		plugin.metrics.MaxTime = execInfo.ExecutionTime
	}

	if execInfo.ExecutionTime < plugin.metrics.MinTime {
		plugin.metrics.MinTime = execInfo.ExecutionTime
	}

	// 计算平均时间
	totalQueries := atomic.LoadInt64(&plugin.metrics.TotalQueries)
	if totalQueries > 0 {
		plugin.metrics.AvgTime = plugin.metrics.TotalTime / time.Duration(totalQueries)
	}

	// 更新时间分布
	timeRange := plugin.getTimeRange(execInfo.ExecutionTime)
	plugin.metrics.TimeDistribution[timeRange]++
}

// getTimeRange 获取时间范围
func (plugin *PerformancePlugin) getTimeRange(duration time.Duration) string {
	ms := duration.Milliseconds()

	switch {
	case ms < 10:
		return "0-10ms"
	case ms < 50:
		return "10-50ms"
	case ms < 100:
		return "50-100ms"
	case ms < 500:
		return "100-500ms"
	case ms < 1000:
		return "500ms-1s"
	case ms < 5000:
		return "1-5s"
	case ms < 10000:
		return "5-10s"
	default:
		return "10s+"
	}
}

// recordSlowQuery 记录慢查询
func (plugin *PerformancePlugin) recordSlowQuery(execInfo *QueryExecutionInfo) {
	atomic.AddInt64(&plugin.metrics.SlowQueries, 1)

	record := SlowQueryRecord{
		SQL:           execInfo.SQL,
		Parameters:    execInfo.Parameters,
		ExecutionTime: execInfo.ExecutionTime,
		Timestamp:     execInfo.StartTime,
		Method:        execInfo.Method,
		Error:         execInfo.Error,
		Context:       make(map[string]any),
	}

	plugin.slowQueries.AddRecord(record)
}

// AddRecord 添加慢查询记录
func (recorder *SlowQueryRecorder) AddRecord(record SlowQueryRecord) {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()

	// 如果达到最大容量，移除最旧的记录
	if len(recorder.records) >= recorder.maxSize {
		recorder.records = recorder.records[1:]
	}

	recorder.records = append(recorder.records, record)
}

// GetRecords 获取慢查询记录
func (recorder *SlowQueryRecorder) GetRecords() []SlowQueryRecord {
	recorder.mutex.RLock()
	defer recorder.mutex.RUnlock()

	// 返回副本
	records := make([]SlowQueryRecord, len(recorder.records))
	copy(records, recorder.records)
	return records
}

// Clear 清空记录
func (recorder *SlowQueryRecorder) Clear() {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()

	recorder.records = recorder.records[:0]
}

// GetMetrics 获取性能指标
func (plugin *PerformancePlugin) GetMetrics() *PerformanceMetrics {
	plugin.metrics.mutex.RLock()
	defer plugin.metrics.mutex.RUnlock()

	// 返回副本
	metrics := &PerformanceMetrics{
		TotalQueries:      atomic.LoadInt64(&plugin.metrics.TotalQueries),
		SlowQueries:       atomic.LoadInt64(&plugin.metrics.SlowQueries),
		FailedQueries:     atomic.LoadInt64(&plugin.metrics.FailedQueries),
		TotalTime:         plugin.metrics.TotalTime,
		MaxTime:           plugin.metrics.MaxTime,
		MinTime:           plugin.metrics.MinTime,
		AvgTime:           plugin.metrics.AvgTime,
		ConcurrentQueries: atomic.LoadInt64(&plugin.metrics.ConcurrentQueries),
		MaxConcurrent:     atomic.LoadInt64(&plugin.metrics.MaxConcurrent),
		TimeDistribution:  make(map[string]int64),
	}

	// 复制时间分布
	for k, v := range plugin.metrics.TimeDistribution {
		metrics.TimeDistribution[k] = v
	}

	return metrics
}

// GetSlowQueries 获取慢查询记录
func (plugin *PerformancePlugin) GetSlowQueries() []SlowQueryRecord {
	return plugin.slowQueries.GetRecords()
}

// ResetMetrics 重置统计信息
func (plugin *PerformancePlugin) ResetMetrics() {
	plugin.metrics.mutex.Lock()
	defer plugin.metrics.mutex.Unlock()

	atomic.StoreInt64(&plugin.metrics.TotalQueries, 0)
	atomic.StoreInt64(&plugin.metrics.SlowQueries, 0)
	atomic.StoreInt64(&plugin.metrics.FailedQueries, 0)
	plugin.metrics.TotalTime = 0
	plugin.metrics.MaxTime = 0
	plugin.metrics.MinTime = time.Duration(^uint64(0) >> 1)
	plugin.metrics.AvgTime = 0
	atomic.StoreInt64(&plugin.metrics.MaxConcurrent, 0)

	// 清空时间分布
	for k := range plugin.metrics.TimeDistribution {
		delete(plugin.metrics.TimeDistribution, k)
	}

	// 清空慢查询记录
	plugin.slowQueries.Clear()
}

// GetPerformanceReport 获取性能报告
func (plugin *PerformancePlugin) GetPerformanceReport() map[string]any {
	metrics := plugin.GetMetrics()
	slowQueries := plugin.GetSlowQueries()

	report := map[string]any{
		"总查询数":   metrics.TotalQueries,
		"慢查询数":   metrics.SlowQueries,
		"失败查询数":  metrics.FailedQueries,
		"总执行时间":  metrics.TotalTime.String(),
		"最大执行时间": metrics.MaxTime.String(),
		"最小执行时间": metrics.MinTime.String(),
		"平均执行时间": metrics.AvgTime.String(),
		"当前并发数":  metrics.ConcurrentQueries,
		"最大并发数":  metrics.MaxConcurrent,
		"时间分布":   metrics.TimeDistribution,
		"慢查询记录数": len(slowQueries),
		"慢查询阈值":  plugin.slowQueryThreshold.String(),
	}

	if len(slowQueries) > 0 {
		report["最近慢查询"] = slowQueries[len(slowQueries)-1]
	}

	return report
}
