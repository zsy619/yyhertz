// Package mybatis SQL执行统计和性能监控
//
// 提供详细的SQL执行统计、慢查询检测和性能分析功能
package mybatis

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

// SQLStats SQL执行统计信息
type SQLStats struct {
	StatementID    string        `json:"statementId"`    // 语句ID
	SQL            string        `json:"sql"`            // SQL语句
	ExecuteCount   int64         `json:"executeCount"`   // 执行次数
	TotalTime      time.Duration `json:"totalTime"`      // 总执行时间
	MinTime        time.Duration `json:"minTime"`        // 最短执行时间
	MaxTime        time.Duration `json:"maxTime"`        // 最长执行时间
	AvgTime        time.Duration `json:"avgTime"`        // 平均执行时间
	ErrorCount     int64         `json:"errorCount"`     // 错误次数
	SlowQueryCount int64         `json:"slowQueryCount"` // 慢查询次数
	LastExecuteTime time.Time    `json:"lastExecuteTime"` // 最后执行时间
	DataSource     string        `json:"dataSource"`     // 数据源
	
	// 执行时间分布
	TimeDistribution map[string]int64 `json:"timeDistribution"` // 时间区间分布
}

// SlowQuery 慢查询记录
type SlowQuery struct {
	ID             string        `json:"id"`             // 唯一ID
	StatementID    string        `json:"statementId"`    // 语句ID
	SQL            string        `json:"sql"`            // SQL语句
	Parameters     []any         `json:"parameters"`     // 参数
	ExecuteTime    time.Duration `json:"executeTime"`    // 执行时间
	StartTime      time.Time     `json:"startTime"`      // 开始时间
	EndTime        time.Time     `json:"endTime"`        // 结束时间
	DataSource     string        `json:"dataSource"`     // 数据源
	ErrorMessage   string        `json:"errorMessage"`   // 错误信息（如果有）
	StackTrace     string        `json:"stackTrace"`     // 调用栈（如果需要）
}

// PerformanceConfig 性能监控配置
type PerformanceConfig struct {
	SlowQueryThreshold    time.Duration `json:"slowQueryThreshold"`    // 慢查询阈值
	EnableSlowQueryLog    bool          `json:"enableSlowQueryLog"`    // 是否启用慢查询日志
	MaxSlowQueryRecords   int           `json:"maxSlowQueryRecords"`   // 最大慢查询记录数
	EnableStatistics      bool          `json:"enableStatistics"`     // 是否启用统计
	MaxStatisticsRecords  int           `json:"maxStatisticsRecords"`  // 最大统计记录数
	StatisticsResetPeriod time.Duration `json:"statisticsResetPeriod"` // 统计重置周期
	EnableStackTrace      bool          `json:"enableStackTrace"`      // 是否收集调用栈
}

// DefaultPerformanceConfig 默认性能监控配置
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		SlowQueryThreshold:    1000 * time.Millisecond, // 1秒
		EnableSlowQueryLog:    true,
		MaxSlowQueryRecords:   1000,
		EnableStatistics:      true,
		MaxStatisticsRecords:  500,
		StatisticsResetPeriod: 24 * time.Hour, // 24小时重置
		EnableStackTrace:      false,
	}
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	config      *PerformanceConfig
	sqlStats    map[string]*SQLStats // statementId -> stats
	slowQueries []*SlowQuery         // 慢查询记录
	mutex       sync.RWMutex
	
	// 重置定时器
	resetTimer *time.Timer
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(config *PerformanceConfig) *PerformanceMonitor {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	monitor := &PerformanceMonitor{
		config:      config,
		sqlStats:    make(map[string]*SQLStats),
		slowQueries: make([]*SlowQuery, 0),
	}

	// 启动定期重置
	if config.StatisticsResetPeriod > 0 {
		monitor.resetTimer = time.AfterFunc(config.StatisticsResetPeriod, monitor.resetStatistics)
	}

	return monitor
}

// RecordExecution 记录SQL执行
func (pm *PerformanceMonitor) RecordExecution(
	statementId string,
	sql string,
	parameters []any,
	executeTime time.Duration,
	dataSource string,
	err error,
) {
	if !pm.config.EnableStatistics && !pm.config.EnableSlowQueryLog {
		return
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	now := time.Now()

	// 记录统计信息
	if pm.config.EnableStatistics {
		pm.recordStats(statementId, sql, executeTime, dataSource, err, now)
	}

	// 记录慢查询
	if pm.config.EnableSlowQueryLog && executeTime >= pm.config.SlowQueryThreshold {
		pm.recordSlowQuery(statementId, sql, parameters, executeTime, dataSource, err, now)
	}
}

// recordStats 记录统计信息
func (pm *PerformanceMonitor) recordStats(
	statementId string,
	sql string,
	executeTime time.Duration,
	dataSource string,
	err error,
	now time.Time,
) {
	stats, exists := pm.sqlStats[statementId]
	if !exists {
		stats = &SQLStats{
			StatementID:      statementId,
			SQL:             sql,
			DataSource:      dataSource,
			MinTime:         executeTime,
			MaxTime:         executeTime,
			TimeDistribution: make(map[string]int64),
		}
		pm.sqlStats[statementId] = stats
	}

	// 更新统计信息
	stats.ExecuteCount++
	stats.TotalTime += executeTime
	stats.LastExecuteTime = now

	// 更新最小最大时间
	if executeTime < stats.MinTime {
		stats.MinTime = executeTime
	}
	if executeTime > stats.MaxTime {
		stats.MaxTime = executeTime
	}

	// 计算平均时间
	stats.AvgTime = time.Duration(int64(stats.TotalTime) / stats.ExecuteCount)

	// 错误统计
	if err != nil {
		stats.ErrorCount++
	}

	// 慢查询统计
	if executeTime >= pm.config.SlowQueryThreshold {
		stats.SlowQueryCount++
	}

	// 时间分布统计
	pm.updateTimeDistribution(stats, executeTime)

	// 检查最大记录数限制
	if len(pm.sqlStats) > pm.config.MaxStatisticsRecords {
		pm.cleanupOldStats()
	}
}

// updateTimeDistribution 更新时间分布统计
func (pm *PerformanceMonitor) updateTimeDistribution(stats *SQLStats, executeTime time.Duration) {
	var bucket string
	millis := executeTime.Milliseconds()

	switch {
	case millis < 100:
		bucket = "0-100ms"
	case millis < 500:
		bucket = "100-500ms"
	case millis < 1000:
		bucket = "500ms-1s"
	case millis < 5000:
		bucket = "1s-5s"
	case millis < 10000:
		bucket = "5s-10s"
	default:
		bucket = "10s+"
	}

	stats.TimeDistribution[bucket]++
}

// recordSlowQuery 记录慢查询
func (pm *PerformanceMonitor) recordSlowQuery(
	statementId string,
	sql string,
	parameters []any,
	executeTime time.Duration,
	dataSource string,
	err error,
	now time.Time,
) {
	slowQuery := &SlowQuery{
		ID:          fmt.Sprintf("%d_%s", now.UnixNano(), statementId),
		StatementID: statementId,
		SQL:         sql,
		Parameters:  parameters,
		ExecuteTime: executeTime,
		StartTime:   now.Add(-executeTime),
		EndTime:     now,
		DataSource:  dataSource,
	}

	if err != nil {
		slowQuery.ErrorMessage = err.Error()
	}

	// 如果需要调用栈，收集调用栈信息
	if pm.config.EnableStackTrace {
		// 这里可以使用runtime.Stack()获取调用栈
		// slowQuery.StackTrace = string(runtime.Stack())
	}

	pm.slowQueries = append(pm.slowQueries, slowQuery)

	// 检查最大记录数限制
	if len(pm.slowQueries) > pm.config.MaxSlowQueryRecords {
		// 删除最旧的记录
		pm.slowQueries = pm.slowQueries[1:]
	}

	// 记录到日志
	log.Printf("[SLOW QUERY] %s executed in %v, SQL: %s", 
		statementId, executeTime, pm.shortenSQL(sql))
}

// shortenSQL 缩短SQL语句用于日志输出
func (pm *PerformanceMonitor) shortenSQL(sql string) string {
	const maxLen = 100
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "..."
}

// cleanupOldStats 清理旧的统计信息（保留执行次数最多的）
func (pm *PerformanceMonitor) cleanupOldStats() {
	type statsPair struct {
		key   string
		stats *SQLStats
	}

	pairs := make([]statsPair, 0, len(pm.sqlStats))
	for key, stats := range pm.sqlStats {
		pairs = append(pairs, statsPair{key, stats})
	}

	// 按执行次数排序
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].stats.ExecuteCount > pairs[j].stats.ExecuteCount
	})

	// 保留前80%的记录
	keepCount := int(float64(pm.config.MaxStatisticsRecords) * 0.8)
	if keepCount < len(pairs) {
		// 清理后20%的记录
		for i := keepCount; i < len(pairs); i++ {
			delete(pm.sqlStats, pairs[i].key)
		}
		log.Printf("[Performance] Cleaned up %d old statistics records", len(pairs)-keepCount)
	}
}

// GetStatistics 获取所有统计信息
func (pm *PerformanceMonitor) GetStatistics() map[string]*SQLStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]*SQLStats)
	for k, v := range pm.sqlStats {
		// 深拷贝
		statsCopy := *v
		statsCopy.TimeDistribution = make(map[string]int64)
		for tk, tv := range v.TimeDistribution {
			statsCopy.TimeDistribution[tk] = tv
		}
		result[k] = &statsCopy
	}
	return result
}

// GetSlowQueries 获取慢查询记录
func (pm *PerformanceMonitor) GetSlowQueries(limit int) []*SlowQuery {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if limit <= 0 || limit > len(pm.slowQueries) {
		limit = len(pm.slowQueries)
	}

	result := make([]*SlowQuery, limit)
	// 返回最新的记录
	start := len(pm.slowQueries) - limit
	for i := 0; i < limit; i++ {
		// 深拷贝
		slowQueryCopy := *pm.slowQueries[start+i]
		if slowQueryCopy.Parameters != nil {
			slowQueryCopy.Parameters = make([]any, len(pm.slowQueries[start+i].Parameters))
			copy(slowQueryCopy.Parameters, pm.slowQueries[start+i].Parameters)
		}
		result[i] = &slowQueryCopy
	}

	return result
}

// GetTopSlowQueries 获取最慢的查询
func (pm *PerformanceMonitor) GetTopSlowQueries(limit int) []*SlowQuery {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if len(pm.slowQueries) == 0 {
		return []*SlowQuery{}
	}

	// 复制切片进行排序
	queries := make([]*SlowQuery, len(pm.slowQueries))
	copy(queries, pm.slowQueries)

	// 按执行时间降序排序
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].ExecuteTime > queries[j].ExecuteTime
	})

	if limit <= 0 || limit > len(queries) {
		limit = len(queries)
	}

	return queries[:limit]
}

// GetStatisticsReport 获取统计报告
func (pm *PerformanceMonitor) GetStatisticsReport() *StatisticsReport {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	report := &StatisticsReport{
		TotalStatements:  len(pm.sqlStats),
		TotalSlowQueries: len(pm.slowQueries),
		GeneratedAt:      time.Now(),
	}

	var totalExecutions int64
	var totalTime time.Duration
	var totalErrors int64
	var totalSlowQueries int64

	for _, stats := range pm.sqlStats {
		totalExecutions += stats.ExecuteCount
		totalTime += stats.TotalTime
		totalErrors += stats.ErrorCount
		totalSlowQueries += stats.SlowQueryCount
	}

	report.TotalExecutions = totalExecutions
	report.TotalExecutionTime = totalTime
	report.TotalErrors = totalErrors
	report.TotalSlowQueryExecutions = totalSlowQueries

	if totalExecutions > 0 {
		report.AvgExecutionTime = time.Duration(int64(totalTime) / totalExecutions)
		report.ErrorRate = float64(totalErrors) / float64(totalExecutions)
		report.SlowQueryRate = float64(totalSlowQueries) / float64(totalExecutions)
	}

	return report
}

// StatisticsReport 统计报告
type StatisticsReport struct {
	TotalStatements         int           `json:"totalStatements"`
	TotalExecutions         int64         `json:"totalExecutions"`
	TotalExecutionTime      time.Duration `json:"totalExecutionTime"`
	AvgExecutionTime        time.Duration `json:"avgExecutionTime"`
	TotalErrors             int64         `json:"totalErrors"`
	ErrorRate               float64       `json:"errorRate"`
	TotalSlowQueries        int           `json:"totalSlowQueries"`
	TotalSlowQueryExecutions int64        `json:"totalSlowQueryExecutions"`
	SlowQueryRate           float64       `json:"slowQueryRate"`
	GeneratedAt             time.Time     `json:"generatedAt"`
}

// resetStatistics 重置统计信息
func (pm *PerformanceMonitor) resetStatistics() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 清空统计信息
	pm.sqlStats = make(map[string]*SQLStats)
	pm.slowQueries = make([]*SlowQuery, 0)

	log.Printf("[Performance] Statistics reset completed")

	// 设置下一次重置
	if pm.config.StatisticsResetPeriod > 0 {
		pm.resetTimer = time.AfterFunc(pm.config.StatisticsResetPeriod, pm.resetStatistics)
	}
}

// ClearStatistics 手动清空统计信息
func (pm *PerformanceMonitor) ClearStatistics() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.sqlStats = make(map[string]*SQLStats)
	pm.slowQueries = make([]*SlowQuery, 0)

	log.Printf("[Performance] Statistics manually cleared")
}

// UpdateConfig 更新配置
func (pm *PerformanceMonitor) UpdateConfig(config *PerformanceConfig) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.config = config

	// 重新设置重置定时器
	if pm.resetTimer != nil {
		pm.resetTimer.Stop()
		pm.resetTimer = nil
	}

	if config.StatisticsResetPeriod > 0 {
		pm.resetTimer = time.AfterFunc(config.StatisticsResetPeriod, pm.resetStatistics)
	}
}

// Close 关闭监控器
func (pm *PerformanceMonitor) Close() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.resetTimer != nil {
		pm.resetTimer.Stop()
		pm.resetTimer = nil
	}

	log.Printf("[Performance] Monitor closed")
}

// ExecutionContext 执行上下文，用于性能监控
type ExecutionContext struct {
	StatementID string
	SQL         string
	Parameters  []any
	DataSource  string
	StartTime   time.Time
	monitor     *PerformanceMonitor
}

// NewExecutionContext 创建执行上下文
func NewExecutionContext(monitor *PerformanceMonitor, statementId, sql, dataSource string, parameters []any) *ExecutionContext {
	return &ExecutionContext{
		StatementID: statementId,
		SQL:         sql,
		Parameters:  parameters,
		DataSource:  dataSource,
		StartTime:   time.Now(),
		monitor:     monitor,
	}
}

// Finish 完成执行并记录性能数据
func (ec *ExecutionContext) Finish(err error) {
	if ec.monitor == nil {
		return
	}

	executeTime := time.Since(ec.StartTime)
	ec.monitor.RecordExecution(
		ec.StatementID,
		ec.SQL,
		ec.Parameters,
		executeTime,
		ec.DataSource,
		err,
	)
}

// 性能监控相关的上下文key
type performanceContextKey string

const (
	PerformanceMonitorKey performanceContextKey = "performance_monitor"
	ExecutionContextKey   performanceContextKey = "execution_context"
)

// SetPerformanceMonitor 在上下文中设置性能监控器
func SetPerformanceMonitor(ctx context.Context, monitor *PerformanceMonitor) context.Context {
	return context.WithValue(ctx, PerformanceMonitorKey, monitor)
}

// GetPerformanceMonitor 从上下文中获取性能监控器
func GetPerformanceMonitor(ctx context.Context) *PerformanceMonitor {
	if monitor, ok := ctx.Value(PerformanceMonitorKey).(*PerformanceMonitor); ok {
		return monitor
	}
	return nil
}

// SetExecutionContext 在上下文中设置执行上下文
func SetExecutionContext(ctx context.Context, execCtx *ExecutionContext) context.Context {
	return context.WithValue(ctx, ExecutionContextKey, execCtx)
}

// GetExecutionContext 从上下文中获取执行上下文
func GetExecutionContext(ctx context.Context) *ExecutionContext {
	if execCtx, ok := ctx.Value(ExecutionContextKey).(*ExecutionContext); ok {
		return execCtx
	}
	return nil
}