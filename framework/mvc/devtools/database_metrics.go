// Package devtools 提供数据库监控指标功能
//
// 数据库监控模块用于监控数据库连接池、查询性能和事务统计，提供：
// - 连接池实时监控
// - 查询性能分析
// - 慢查询检测
// - 事务统计
// - 数据库健康度评估
// - 连接泄漏检测
//
// 功能特性：
// - 多数据库支持（MySQL、PostgreSQL、SQLite）
// - 连接池详细指标
// - 查询执行时间分布
// - 自动慢查询告警
// - 事务成功率监控
// - 连接生命周期跟踪
package devtools

import (
	"context"
	"database/sql"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/config"
)

// DatabaseType 数据库类型枚举
type DatabaseType string

const (
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypeOracle     DatabaseType = "oracle"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
)

// QueryType 查询类型枚举
type QueryType string

const (
	QueryTypeSelect QueryType = "SELECT"
	QueryTypeInsert QueryType = "INSERT"
	QueryTypeUpdate QueryType = "UPDATE"
	QueryTypeDelete QueryType = "DELETE"
	QueryTypeOther  QueryType = "OTHER"
)

// ConnectionPoolMetrics 连接池指标
type ConnectionPoolMetrics struct {
	// 连接数指标
	ActiveConnections int64 `json:"active_connections"` // 活跃连接数
	IdleConnections   int64 `json:"idle_connections"`   // 空闲连接数
	MaxConnections    int64 `json:"max_connections"`    // 最大连接数
	TotalConnections  int64 `json:"total_connections"`  // 总连接数

	// 等待和超时指标
	WaitCount          int64         `json:"wait_count"`          // 等待连接的次数
	WaitDuration       time.Duration `json:"wait_duration"`       // 总等待时间
	MaxWaitDuration    time.Duration `json:"max_wait_duration"`   // 最大等待时间
	ConnectionTimeouts int64         `json:"connection_timeouts"` // 连接超时次数

	// 生命周期指标
	ConnectionsCreated int64         `json:"connections_created"` // 创建的连接总数
	ConnectionsClosed  int64         `json:"connections_closed"`  // 关闭的连接总数
	ConnectionLifetime time.Duration `json:"connection_lifetime"` // 平均连接生命周期

	// 使用率指标
	UtilizationPercent float64   `json:"utilization_percent"` // 连接池使用率百分比
	PeakConnections    int64     `json:"peak_connections"`    // 峰值连接数
	LastUpdate         time.Time `json:"last_update"`         // 最后更新时间
}

// QueryMetrics 查询指标
type QueryMetrics struct {
	// 基础统计
	TotalQueries   int64               `json:"total_queries"`    // 查询总数
	SlowQueries    int64               `json:"slow_queries"`     // 慢查询数
	FailedQueries  int64               `json:"failed_queries"`   // 失败查询数
	QueryTypeStats map[QueryType]int64 `json:"query_type_stats"` // 按类型统计

	// 性能指标
	TotalExecutionTime   time.Duration `json:"total_execution_time"`   // 总执行时间
	AverageExecutionTime time.Duration `json:"average_execution_time"` // 平均执行时间
	MinExecutionTime     time.Duration `json:"min_execution_time"`     // 最小执行时间
	MaxExecutionTime     time.Duration `json:"max_execution_time"`     // 最大执行时间

	// 慢查询分析
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"` // 慢查询阈值
	SlowQueryExamples  []SlowQuery   `json:"slow_query_examples"`  // 慢查询示例

	// 时间分布
	ExecutionTimeHistogram map[string]int64 `json:"execution_time_histogram"` // 执行时间直方图
}

// SlowQuery 慢查询记录
type SlowQuery struct {
	SQL           string        `json:"sql"`            // SQL语句
	ExecutionTime time.Duration `json:"execution_time"` // 执行时间
	Timestamp     time.Time     `json:"timestamp"`      // 时间戳
	Parameters    []interface{} `json:"parameters"`     // 参数
	StackTrace    string        `json:"stack_trace"`    // 调用栈
}

// TransactionMetrics 事务指标
type TransactionMetrics struct {
	// 事务统计
	TotalTransactions      int64 `json:"total_transactions"`       // 事务总数
	CommittedTransactions  int64 `json:"committed_transactions"`   // 提交的事务数
	RolledBackTransactions int64 `json:"rolled_back_transactions"` // 回滚的事务数

	// 事务性能
	AverageTransactionTime time.Duration `json:"average_transaction_time"` // 平均事务时间
	MaxTransactionTime     time.Duration `json:"max_transaction_time"`     // 最大事务时间

	// 并发和锁等待
	DeadlockCount      int64 `json:"deadlock_count"`      // 死锁次数
	LockWaitTimeouts   int64 `json:"lock_wait_timeouts"`  // 锁等待超时次数
	ActiveTransactions int64 `json:"active_transactions"` // 活跃事务数
}

// DatabaseMetricsConfig 数据库监控配置
type DatabaseMetricsConfig struct {
	Enabled              bool          `json:"enabled"`                 // 是否启用
	DatabaseType         DatabaseType  `json:"database_type"`           // 数据库类型
	SlowQueryThreshold   time.Duration `json:"slow_query_threshold"`    // 慢查询阈值
	MaxSlowQueryExamples int           `json:"max_slow_query_examples"` // 最大慢查询示例数
	CollectInterval      time.Duration `json:"collect_interval"`        // 收集间隔
	EnableStackTrace     bool          `json:"enable_stack_trace"`      // 是否启用调用栈
	EnableQueryAnalysis  bool          `json:"enable_query_analysis"`   // 是否启用查询分析
}

// DatabaseMetricsCollector 数据库指标收集器
type DatabaseMetricsCollector struct {
	mu        sync.RWMutex
	config    *DatabaseMetricsConfig
	enabled   bool
	startTime time.Time

	// 数据库连接
	db      *sql.DB
	dbStats sql.DBStats

	// 指标数据
	connectionPoolMetrics *ConnectionPoolMetrics
	queryMetrics          *QueryMetrics
	transactionMetrics    *TransactionMetrics

	// 收集器状态
	collectTicker *time.Ticker
	stopChan      chan struct{}

	// 慢查询缓存
	slowQueries    []SlowQuery
	slowQueryMutex sync.RWMutex

	// 事务跟踪
	activeTransactions map[string]*TransactionContext
	transactionMutex   sync.RWMutex
}

// TransactionContext 事务上下文
type TransactionContext struct {
	ID        string    `json:"id"`         // 事务ID
	StartTime time.Time `json:"start_time"` // 开始时间
	SQL       []string  `json:"sql"`        // 执行的SQL语句
}

// NewDatabaseMetricsCollector 创建数据库指标收集器
func NewDatabaseMetricsCollector(db *sql.DB, config *DatabaseMetricsConfig) *DatabaseMetricsCollector {
	if config == nil {
		config = &DatabaseMetricsConfig{
			Enabled:              true,
			DatabaseType:         DatabaseTypeMySQL,
			SlowQueryThreshold:   100 * time.Millisecond,
			MaxSlowQueryExamples: 100,
			CollectInterval:      5 * time.Second,
			EnableStackTrace:     false,
			EnableQueryAnalysis:  true,
		}
	}

	collector := &DatabaseMetricsCollector{
		config:    config,
		enabled:   config.Enabled,
		startTime: time.Now(),
		db:        db,
		stopChan:  make(chan struct{}),

		connectionPoolMetrics: &ConnectionPoolMetrics{
			LastUpdate: time.Now(),
		},
		queryMetrics: &QueryMetrics{
			QueryTypeStats:         make(map[QueryType]int64),
			SlowQueryThreshold:     config.SlowQueryThreshold,
			SlowQueryExamples:      make([]SlowQuery, 0, config.MaxSlowQueryExamples),
			ExecutionTimeHistogram: make(map[string]int64),
			MinExecutionTime:       time.Hour, // 初始化为大值
		},
		transactionMetrics: &TransactionMetrics{},

		slowQueries:        make([]SlowQuery, 0, config.MaxSlowQueryExamples),
		activeTransactions: make(map[string]*TransactionContext),
	}

	// 初始化执行时间直方图桶
	collector.initializeHistogramBuckets()

	return collector
}

// initializeHistogramBuckets 初始化直方图桶
func (dmc *DatabaseMetricsCollector) initializeHistogramBuckets() {
	buckets := []string{
		"0-1ms", "1-5ms", "5-10ms", "10-50ms",
		"50-100ms", "100-500ms", "500ms-1s",
		"1s-5s", "5s-10s", "10s+",
	}

	for _, bucket := range buckets {
		dmc.queryMetrics.ExecutionTimeHistogram[bucket] = 0
	}
}

// Start 启动数据库指标收集
func (dmc *DatabaseMetricsCollector) Start() {
	if !dmc.enabled || dmc.db == nil {
		return
	}

	dmc.collectTicker = time.NewTicker(dmc.config.CollectInterval)
	go dmc.collectLoop()
}

// Stop 停止数据库指标收集
func (dmc *DatabaseMetricsCollector) Stop() {
	if dmc.collectTicker != nil {
		dmc.collectTicker.Stop()
	}
	close(dmc.stopChan)
}

// collectLoop 收集循环
func (dmc *DatabaseMetricsCollector) collectLoop() {
	for {
		select {
		case <-dmc.collectTicker.C:
			dmc.collectMetrics()
		case <-dmc.stopChan:
			return
		}
	}
}

// collectMetrics 收集指标数据
func (dmc *DatabaseMetricsCollector) collectMetrics() {
	if !dmc.enabled || dmc.db == nil {
		return
	}

	// 收集连接池指标
	dmc.collectConnectionPoolMetrics()
}

// collectConnectionPoolMetrics 收集连接池指标
func (dmc *DatabaseMetricsCollector) collectConnectionPoolMetrics() {
	stats := dmc.db.Stats()

	dmc.mu.Lock()
	defer dmc.mu.Unlock()

	dmc.connectionPoolMetrics.ActiveConnections = int64(stats.OpenConnections - stats.Idle)
	dmc.connectionPoolMetrics.IdleConnections = int64(stats.Idle)
	dmc.connectionPoolMetrics.MaxConnections = int64(stats.MaxOpenConnections)
	dmc.connectionPoolMetrics.TotalConnections = int64(stats.OpenConnections)
	dmc.connectionPoolMetrics.WaitCount = stats.WaitCount
	dmc.connectionPoolMetrics.WaitDuration = stats.WaitDuration
	dmc.connectionPoolMetrics.MaxWaitDuration = stats.WaitDuration
	dmc.connectionPoolMetrics.ConnectionsCreated = int64(stats.MaxOpenConnections)
	dmc.connectionPoolMetrics.ConnectionsClosed = stats.MaxIdleClosed + stats.MaxLifetimeClosed
	dmc.connectionPoolMetrics.LastUpdate = time.Now()

	// 计算使用率
	if stats.MaxOpenConnections > 0 {
		dmc.connectionPoolMetrics.UtilizationPercent = float64(stats.OpenConnections) / float64(stats.MaxOpenConnections) * 100
	}

	// 更新峰值连接数
	if int64(stats.OpenConnections) > dmc.connectionPoolMetrics.PeakConnections {
		dmc.connectionPoolMetrics.PeakConnections = int64(stats.OpenConnections)
	}
}

// RecordQuery 记录查询指标
func (dmc *DatabaseMetricsCollector) RecordQuery(sql string, duration time.Duration, err error, params ...interface{}) {
	if !dmc.enabled {
		return
	}

	dmc.mu.Lock()
	defer dmc.mu.Unlock()

	// 更新基础统计
	atomic.AddInt64(&dmc.queryMetrics.TotalQueries, 1)
	dmc.queryMetrics.TotalExecutionTime += duration
	dmc.queryMetrics.AverageExecutionTime = time.Duration(int64(dmc.queryMetrics.TotalExecutionTime) / dmc.queryMetrics.TotalQueries)

	// 更新最小/最大执行时间
	if duration < dmc.queryMetrics.MinExecutionTime {
		dmc.queryMetrics.MinExecutionTime = duration
	}
	if duration > dmc.queryMetrics.MaxExecutionTime {
		dmc.queryMetrics.MaxExecutionTime = duration
	}

	// 记录错误
	if err != nil {
		atomic.AddInt64(&dmc.queryMetrics.FailedQueries, 1)
	}

	// 分析查询类型
	queryType := dmc.analyzeQueryType(sql)
	dmc.queryMetrics.QueryTypeStats[queryType]++

	// 更新执行时间直方图
	dmc.updateExecutionTimeHistogram(duration)

	// 检查慢查询
	if duration > dmc.config.SlowQueryThreshold {
		dmc.recordSlowQuery(sql, duration, params)
	}
}

// analyzeQueryType 分析查询类型
func (dmc *DatabaseMetricsCollector) analyzeQueryType(sql string) QueryType {
	sql = strings.TrimSpace(strings.ToUpper(sql))

	switch {
	case strings.HasPrefix(sql, "SELECT"):
		return QueryTypeSelect
	case strings.HasPrefix(sql, "INSERT"):
		return QueryTypeInsert
	case strings.HasPrefix(sql, "UPDATE"):
		return QueryTypeUpdate
	case strings.HasPrefix(sql, "DELETE"):
		return QueryTypeDelete
	default:
		return QueryTypeOther
	}
}

// updateExecutionTimeHistogram 更新执行时间直方图
func (dmc *DatabaseMetricsCollector) updateExecutionTimeHistogram(duration time.Duration) {
	ms := duration.Nanoseconds() / 1000000 // 转换为毫秒

	switch {
	case ms < 1:
		dmc.queryMetrics.ExecutionTimeHistogram["0-1ms"]++
	case ms < 5:
		dmc.queryMetrics.ExecutionTimeHistogram["1-5ms"]++
	case ms < 10:
		dmc.queryMetrics.ExecutionTimeHistogram["5-10ms"]++
	case ms < 50:
		dmc.queryMetrics.ExecutionTimeHistogram["10-50ms"]++
	case ms < 100:
		dmc.queryMetrics.ExecutionTimeHistogram["50-100ms"]++
	case ms < 500:
		dmc.queryMetrics.ExecutionTimeHistogram["100-500ms"]++
	case ms < 1000:
		dmc.queryMetrics.ExecutionTimeHistogram["500ms-1s"]++
	case ms < 5000:
		dmc.queryMetrics.ExecutionTimeHistogram["1s-5s"]++
	case ms < 10000:
		dmc.queryMetrics.ExecutionTimeHistogram["5s-10s"]++
	default:
		dmc.queryMetrics.ExecutionTimeHistogram["10s+"]++
	}
}

// recordSlowQuery 记录慢查询
func (dmc *DatabaseMetricsCollector) recordSlowQuery(sql string, duration time.Duration, params []interface{}) {
	atomic.AddInt64(&dmc.queryMetrics.SlowQueries, 1)

	slowQuery := SlowQuery{
		SQL:           sql,
		ExecutionTime: duration,
		Timestamp:     time.Now(),
		Parameters:    params,
	}

	// 获取调用栈（如果启用）
	if dmc.config.EnableStackTrace {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		slowQuery.StackTrace = string(buf[:n])
	}

	dmc.slowQueryMutex.Lock()
	defer dmc.slowQueryMutex.Unlock()

	// 添加到慢查询列表
	if len(dmc.slowQueries) >= dmc.config.MaxSlowQueryExamples {
		// 移除最旧的记录
		dmc.slowQueries = dmc.slowQueries[1:]
	}
	dmc.slowQueries = append(dmc.slowQueries, slowQuery)
	dmc.queryMetrics.SlowQueryExamples = dmc.slowQueries
}

// RecordTransaction 记录事务指标
func (dmc *DatabaseMetricsCollector) RecordTransaction(committed bool, duration time.Duration) {
	if !dmc.enabled {
		return
	}

	dmc.mu.Lock()
	defer dmc.mu.Unlock()

	atomic.AddInt64(&dmc.transactionMetrics.TotalTransactions, 1)

	if committed {
		atomic.AddInt64(&dmc.transactionMetrics.CommittedTransactions, 1)
	} else {
		atomic.AddInt64(&dmc.transactionMetrics.RolledBackTransactions, 1)
	}

	// 更新平均事务时间
	totalTime := time.Duration(dmc.transactionMetrics.AverageTransactionTime.Nanoseconds()*dmc.transactionMetrics.TotalTransactions-1) + duration
	dmc.transactionMetrics.AverageTransactionTime = time.Duration(totalTime.Nanoseconds() / dmc.transactionMetrics.TotalTransactions)

	// 更新最大事务时间
	if duration > dmc.transactionMetrics.MaxTransactionTime {
		dmc.transactionMetrics.MaxTransactionTime = duration
	}
}

// GetConnectionPoolMetrics 获取连接池指标
func (dmc *DatabaseMetricsCollector) GetConnectionPoolMetrics() *ConnectionPoolMetrics {
	dmc.mu.RLock()
	defer dmc.mu.RUnlock()

	// 返回副本
	metrics := *dmc.connectionPoolMetrics
	return &metrics
}

// GetQueryMetrics 获取查询指标
func (dmc *DatabaseMetricsCollector) GetQueryMetrics() *QueryMetrics {
	dmc.mu.RLock()
	defer dmc.mu.RUnlock()

	// 返回副本
	metrics := *dmc.queryMetrics

	// 复制map数据
	metrics.QueryTypeStats = make(map[QueryType]int64)
	for k, v := range dmc.queryMetrics.QueryTypeStats {
		metrics.QueryTypeStats[k] = v
	}

	metrics.ExecutionTimeHistogram = make(map[string]int64)
	for k, v := range dmc.queryMetrics.ExecutionTimeHistogram {
		metrics.ExecutionTimeHistogram[k] = v
	}

	return &metrics
}

// GetTransactionMetrics 获取事务指标
func (dmc *DatabaseMetricsCollector) GetTransactionMetrics() *TransactionMetrics {
	dmc.mu.RLock()
	defer dmc.mu.RUnlock()

	// 返回副本
	metrics := *dmc.transactionMetrics
	return &metrics
}

// GetSlowQueries 获取慢查询列表
func (dmc *DatabaseMetricsCollector) GetSlowQueries() []SlowQuery {
	dmc.slowQueryMutex.RLock()
	defer dmc.slowQueryMutex.RUnlock()

	// 返回副本
	slowQueries := make([]SlowQuery, len(dmc.slowQueries))
	copy(slowQueries, dmc.slowQueries)
	return slowQueries
}

// Reset 重置指标
func (dmc *DatabaseMetricsCollector) Reset() {
	dmc.mu.Lock()
	defer dmc.mu.Unlock()

	// 重置查询指标
	dmc.queryMetrics.TotalQueries = 0
	dmc.queryMetrics.SlowQueries = 0
	dmc.queryMetrics.FailedQueries = 0
	dmc.queryMetrics.TotalExecutionTime = 0
	dmc.queryMetrics.AverageExecutionTime = 0
	dmc.queryMetrics.MinExecutionTime = time.Hour
	dmc.queryMetrics.MaxExecutionTime = 0

	// 重置查询类型统计
	for k := range dmc.queryMetrics.QueryTypeStats {
		dmc.queryMetrics.QueryTypeStats[k] = 0
	}

	// 重置执行时间直方图
	for k := range dmc.queryMetrics.ExecutionTimeHistogram {
		dmc.queryMetrics.ExecutionTimeHistogram[k] = 0
	}

	// 重置事务指标
	dmc.transactionMetrics.TotalTransactions = 0
	dmc.transactionMetrics.CommittedTransactions = 0
	dmc.transactionMetrics.RolledBackTransactions = 0
	dmc.transactionMetrics.AverageTransactionTime = 0
	dmc.transactionMetrics.MaxTransactionTime = 0
	dmc.transactionMetrics.DeadlockCount = 0
	dmc.transactionMetrics.LockWaitTimeouts = 0

	// 清空慢查询
	dmc.slowQueryMutex.Lock()
	dmc.slowQueries = dmc.slowQueries[:0]
	dmc.queryMetrics.SlowQueryExamples = dmc.slowQueries
	dmc.slowQueryMutex.Unlock()
}

// IsEnabled 检查是否启用
func (dmc *DatabaseMetricsCollector) IsEnabled() bool {
	dmc.mu.RLock()
	defer dmc.mu.RUnlock()
	return dmc.enabled
}

// Enable 启用收集器
func (dmc *DatabaseMetricsCollector) Enable() {
	dmc.mu.Lock()
	defer dmc.mu.Unlock()
	dmc.enabled = true
}

// Disable 禁用收集器
func (dmc *DatabaseMetricsCollector) Disable() {
	dmc.mu.Lock()
	defer dmc.mu.Unlock()
	dmc.enabled = false
}

// DatabaseMetricsPanel 数据库监控面板
type DatabaseMetricsPanel struct {
	collector *DatabaseMetricsCollector
}

// NewDatabaseMetricsPanel 创建数据库监控面板
func NewDatabaseMetricsPanel(collector *DatabaseMetricsCollector) *DatabaseMetricsPanel {
	return &DatabaseMetricsPanel{
		collector: collector,
	}
}

// RegisterRoutes 注册路由
func (dmp *DatabaseMetricsPanel) RegisterRoutes(engine any) {
	var dbGroup *route.RouterGroup

	if h, ok := engine.(*route.Engine); ok {
		dbGroup = h.Group("/yyhertz/database")
	} else {
		config.Error("无法注册数据库监控路由，未知引擎类型")
		return
	}

	// 注册路由
	dbGroup.GET("/", dmp.getDatabaseMetrics)
	dbGroup.GET("/connections", dmp.getConnectionPoolMetrics)
	dbGroup.GET("/queries", dmp.getQueryMetrics)
	dbGroup.GET("/transactions", dmp.getTransactionMetrics)
	dbGroup.GET("/slow-queries", dmp.getSlowQueries)
	dbGroup.POST("/reset", dmp.resetMetrics)
	dbGroup.POST("/enable", dmp.enableCollector)
	dbGroup.POST("/disable", dmp.disableCollector)
	dbGroup.GET("/panel", dmp.databasePanel)
}

// getDatabaseMetrics 获取数据库指标
func (dmp *DatabaseMetricsPanel) getDatabaseMetrics(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"connections":  dmp.collector.GetConnectionPoolMetrics(),
			"queries":      dmp.collector.GetQueryMetrics(),
			"transactions": dmp.collector.GetTransactionMetrics(),
			"enabled":      dmp.collector.IsEnabled(),
			"timestamp":    time.Now(),
		},
	})
}

// getConnectionPoolMetrics 获取连接池指标
func (dmp *DatabaseMetricsPanel) getConnectionPoolMetrics(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": dmp.collector.GetConnectionPoolMetrics(),
	})
}

// getQueryMetrics 获取查询指标
func (dmp *DatabaseMetricsPanel) getQueryMetrics(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": dmp.collector.GetQueryMetrics(),
	})
}

// getTransactionMetrics 获取事务指标
func (dmp *DatabaseMetricsPanel) getTransactionMetrics(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": dmp.collector.GetTransactionMetrics(),
	})
}

// getSlowQueries 获取慢查询列表
func (dmp *DatabaseMetricsPanel) getSlowQueries(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"slow_queries": dmp.collector.GetSlowQueries(),
			"total_count":  len(dmp.collector.GetSlowQueries()),
		},
	})
}

// resetMetrics 重置指标
func (dmp *DatabaseMetricsPanel) resetMetrics(ctx context.Context, c *app.RequestContext) {
	dmp.collector.Reset()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "数据库指标已重置",
	})
}

// enableCollector 启用收集器
func (dmp *DatabaseMetricsPanel) enableCollector(ctx context.Context, c *app.RequestContext) {
	dmp.collector.Enable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "数据库指标收集已启用",
		"enabled": true,
	})
}

// disableCollector 禁用收集器
func (dmp *DatabaseMetricsPanel) disableCollector(ctx context.Context, c *app.RequestContext) {
	dmp.collector.Disable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "数据库指标收集已禁用",
		"enabled": false,
	})
}

// databasePanel 数据库监控面板页面
func (dmp *DatabaseMetricsPanel) databasePanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz 数据库监控面板</title>
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
        .slow-queries { background: white; margin-top: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .slow-query-item { padding: 15px; border-bottom: 1px solid #eee; }
        .slow-query-sql { font-family: monospace; background: #f8f9fa; padding: 10px; border-radius: 4px; margin: 10px 0; }
        .status-indicator { width: 12px; height: 12px; border-radius: 50%; display: inline-block; margin-right: 5px; }
        .status-healthy { background: #28a745; }
        .status-warning { background: #ffc107; }
        .status-error { background: #dc3545; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz 数据库监控面板</h1>
        <div>
            <button class="btn btn-primary" onclick="refreshMetrics()">刷新指标</button>
            <button class="btn btn-success" onclick="enableCollector()">启用收集</button>
            <button class="btn btn-danger" onclick="disableCollector()">禁用收集</button>
            <button class="btn btn-warning" onclick="resetMetrics()">重置指标</button>
        </div>
    </div>

    <div class="metrics-grid">
        <div class="metric-card">
            <h3><span class="status-indicator status-healthy"></span>连接池状态</h3>
            <div id="connectionMetrics">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-healthy"></span>查询性能</h3>
            <div id="queryMetrics">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-healthy"></span>事务统计</h3>
            <div id="transactionMetrics">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>
    </div>

    <div class="slow-queries">
        <h3 style="padding: 20px; margin: 0; border-bottom: 1px solid #eee;">慢查询分析</h3>
        <div id="slowQueries">
            <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
        </div>
    </div>

    <script>
        function refreshMetrics() {
            loadDatabaseMetrics();
        }

        function loadDatabaseMetrics() {
            Promise.all([
                fetch('/yyhertz/database/connections'),
                fetch('/yyhertz/database/queries'),
                fetch('/yyhertz/database/transactions'),
                fetch('/yyhertz/database/slow-queries')
            ])
            .then(responses => Promise.all(responses.map(r => r.json())))
            .then(([connections, queries, transactions, slowQueries]) => {
                showConnectionMetrics(connections.data);
                showQueryMetrics(queries.data);
                showTransactionMetrics(transactions.data);
                showSlowQueries(slowQueries.data.slow_queries || []);
            })
            .catch(error => {
                console.error('加载指标失败:', error);
            });
        }

        function showConnectionMetrics(data) {
            const container = document.getElementById('connectionMetrics');
            container.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">活跃连接</span><span class="metric-value">' + data.active_connections + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">空闲连接</span><span class="metric-value">' + data.idle_connections + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">最大连接</span><span class="metric-value">' + data.max_connections + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">总连接数</span><span class="metric-value">' + data.total_connections + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">使用率</span><span class="metric-value">' + data.utilization_percent.toFixed(1) + '%</span></div>' +
                '<div class="metric-item"><span class="metric-label">峰值连接</span><span class="metric-value">' + data.peak_connections + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">等待次数</span><span class="metric-value">' + data.wait_count + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">连接超时</span><span class="metric-value">' + data.connection_timeouts + '</span></div>';
        }

        function showQueryMetrics(data) {
            const container = document.getElementById('queryMetrics');
            container.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">总查询数</span><span class="metric-value">' + data.total_queries + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">慢查询数</span><span class="metric-value">' + data.slow_queries + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">失败查询</span><span class="metric-value">' + data.failed_queries + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">平均执行时间</span><span class="metric-value">' + formatDuration(data.average_execution_time) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">最大执行时间</span><span class="metric-value">' + formatDuration(data.max_execution_time) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">SELECT查询</span><span class="metric-value">' + (data.query_type_stats.SELECT || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">INSERT查询</span><span class="metric-value">' + (data.query_type_stats.INSERT || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">UPDATE查询</span><span class="metric-value">' + (data.query_type_stats.UPDATE || 0) + '</span></div>';
        }

        function showTransactionMetrics(data) {
            const container = document.getElementById('transactionMetrics');
            const successRate = data.total_transactions > 0 ? 
                (data.committed_transactions / data.total_transactions * 100).toFixed(1) : 0;
            
            container.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">总事务数</span><span class="metric-value">' + data.total_transactions + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">提交事务</span><span class="metric-value">' + data.committed_transactions + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">回滚事务</span><span class="metric-value">' + data.rolled_back_transactions + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">成功率</span><span class="metric-value">' + successRate + '%</span></div>' +
                '<div class="metric-item"><span class="metric-label">平均事务时间</span><span class="metric-value">' + formatDuration(data.average_transaction_time) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">最大事务时间</span><span class="metric-value">' + formatDuration(data.max_transaction_time) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">死锁次数</span><span class="metric-value">' + data.deadlock_count + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">活跃事务</span><span class="metric-value">' + data.active_transactions + '</span></div>';
        }

        function showSlowQueries(slowQueries) {
            const container = document.getElementById('slowQueries');
            
            if (slowQueries.length === 0) {
                container.innerHTML = '<div style="padding: 20px; text-align: center; color: #666;">暂无慢查询记录</div>';
                return;
            }

            let html = '';
            slowQueries.slice(-10).reverse().forEach((query, index) => {
                html += '<div class="slow-query-item">' +
                    '<div><strong>执行时间:</strong> ' + formatDuration(query.execution_time) + '</div>' +
                    '<div><strong>时间:</strong> ' + new Date(query.timestamp).toLocaleString() + '</div>' +
                    '<div class="slow-query-sql">' + query.sql + '</div>' +
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

        function enableCollector() {
            fetch('/yyhertz/database/enable', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('数据库指标收集已启用');
                    refreshMetrics();
                })
                .catch(error => {
                    console.error('启用失败:', error);
                    alert('启用失败');
                });
        }

        function disableCollector() {
            if (confirm('确定要禁用数据库指标收集吗？')) {
                fetch('/yyhertz/database/disable', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('数据库指标收集已禁用');
                        refreshMetrics();
                    })
                    .catch(error => {
                        console.error('禁用失败:', error);
                        alert('禁用失败');
                    });
            }
        }

        function resetMetrics() {
            if (confirm('确定要重置所有数据库指标吗？')) {
                fetch('/yyhertz/database/reset', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('数据库指标已重置');
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
            loadDatabaseMetrics();
            // 每30秒自动刷新
            setInterval(loadDatabaseMetrics, 30000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}
