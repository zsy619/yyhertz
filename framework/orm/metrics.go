package orm

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	globalConfig "github.com/zsy619/yyhertz/framework/config"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	// 连接指标
	totalConnections      int64
	activeConnections     int64
	connectionErrors      int64
	connectionTimeouts    int64
	
	// 查询指标
	totalQueries          int64
	slowQueries           int64
	failedQueries         int64
	queryDurationTotal    int64 // 纳秒
	
	// 节点指标
	nodeMetrics           map[string]*NodeMetrics
	nodeMetricsMutex      sync.RWMutex
	
	// 时间窗口指标
	windowMetrics         *WindowMetrics
	
	// 配置
	slowQueryThreshold    time.Duration
	
	// 状态
	started               bool
	stopChan              chan struct{}
	mutex                 sync.RWMutex
}

// NodeMetrics 节点指标
type NodeMetrics struct {
	NodeID                string    `json:"node_id"`
	TotalConnections      int64     `json:"total_connections"`
	ActiveConnections     int64     `json:"active_connections"`
	TotalQueries          int64     `json:"total_queries"`
	SlowQueries           int64     `json:"slow_queries"`
	FailedQueries         int64     `json:"failed_queries"`
	ConnectionErrors      int64     `json:"connection_errors"`
	LastConnectionTime    time.Time `json:"last_connection_time"`
	LastQueryTime         time.Time `json:"last_query_time"`
	AverageResponseTime   time.Duration `json:"average_response_time"`
	HealthCheckCount      int64     `json:"health_check_count"`
	HealthCheckFailures   int64     `json:"health_check_failures"`
}

// WindowMetrics 时间窗口指标
type WindowMetrics struct {
	// 1分钟窗口
	OneMinute   *TimeWindowMetrics `json:"one_minute"`
	// 5分钟窗口
	FiveMinutes *TimeWindowMetrics `json:"five_minutes"`
	// 15分钟窗口
	FifteenMinutes *TimeWindowMetrics `json:"fifteen_minutes"`
	// 1小时窗口
	OneHour     *TimeWindowMetrics `json:"one_hour"`
}

// TimeWindowMetrics 时间窗口指标
type TimeWindowMetrics struct {
	StartTime             time.Time     `json:"start_time"`
	EndTime               time.Time     `json:"end_time"`
	TotalConnections      int64         `json:"total_connections"`
	TotalQueries          int64         `json:"total_queries"`
	SlowQueries           int64         `json:"slow_queries"`
	FailedQueries         int64         `json:"failed_queries"`
	AverageResponseTime   time.Duration `json:"average_response_time"`
	MaxResponseTime       time.Duration `json:"max_response_time"`
	MinResponseTime       time.Duration `json:"min_response_time"`
	QPS                   float64       `json:"qps"` // 每秒查询数
	ConnectionsPerSecond  float64       `json:"connections_per_second"`
}

// QueryMetrics 查询指标
type QueryMetrics struct {
	QueryID       string        `json:"query_id"`
	SQL           string        `json:"sql"`
	Duration      time.Duration `json:"duration"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	NodeID        string        `json:"node_id"`
	Success       bool          `json:"success"`
	Error         string        `json:"error,omitempty"`
	RowsAffected  int64         `json:"rows_affected"`
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		nodeMetrics:        make(map[string]*NodeMetrics),
		windowMetrics:      NewWindowMetrics(),
		slowQueryThreshold: time.Millisecond * 500,
		stopChan:          make(chan struct{}),
	}
}

// NewWindowMetrics 创建时间窗口指标
func NewWindowMetrics() *WindowMetrics {
	now := time.Now()
	return &WindowMetrics{
		OneMinute: &TimeWindowMetrics{
			StartTime:       now,
			MinResponseTime: time.Hour, // 初始化为很大的值
		},
		FiveMinutes: &TimeWindowMetrics{
			StartTime:       now,
			MinResponseTime: time.Hour,
		},
		FifteenMinutes: &TimeWindowMetrics{
			StartTime:       now,
			MinResponseTime: time.Hour,
		},
		OneHour: &TimeWindowMetrics{
			StartTime:       now,
			MinResponseTime: time.Hour,
		},
	}
}

// Start 启动指标收集
func (mc *MetricsCollector) Start() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if mc.started {
		return
	}
	
	mc.started = true
	
	// 启动指标收集goroutine
	go mc.collectMetrics()
	
	globalConfig.Info("Metrics collector started")
}

// Stop 停止指标收集
func (mc *MetricsCollector) Stop() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if !mc.started {
		return
	}
	
	mc.started = false
	close(mc.stopChan)
	
	globalConfig.Info("Metrics collector stopped")
}

// Close 关闭指标收集器
func (mc *MetricsCollector) Close() {
	mc.Stop()
}

// collectMetrics 收集指标
func (mc *MetricsCollector) collectMetrics() {
	ticker := time.NewTicker(time.Second * 10) // 每10秒收集一次
	defer ticker.Stop()
	
	for {
		select {
		case <-mc.stopChan:
			return
		case <-ticker.C:
			mc.updateWindowMetrics()
		}
	}
}

// RecordConnection 记录连接
func (mc *MetricsCollector) RecordConnection(nodeID, nodeType string) {
	atomic.AddInt64(&mc.totalConnections, 1)
	atomic.AddInt64(&mc.activeConnections, 1)
	
	mc.getOrCreateNodeMetrics(nodeID).recordConnection()
}

// RecordConnectionRelease 记录连接释放
func (mc *MetricsCollector) RecordConnectionRelease() {
	atomic.AddInt64(&mc.activeConnections, -1)
}

// RecordConnectionError 记录连接错误
func (mc *MetricsCollector) RecordConnectionError(nodeID string) {
	atomic.AddInt64(&mc.connectionErrors, 1)
	
	if nodeID != "" {
		mc.getOrCreateNodeMetrics(nodeID).recordConnectionError()
	}
}

// RecordQuery 记录查询
func (mc *MetricsCollector) RecordQuery(nodeID string, duration time.Duration, success bool) {
	atomic.AddInt64(&mc.totalQueries, 1)
	atomic.AddInt64(&mc.queryDurationTotal, int64(duration))
	
	if !success {
		atomic.AddInt64(&mc.failedQueries, 1)
	}
	
	if duration >= mc.slowQueryThreshold {
		atomic.AddInt64(&mc.slowQueries, 1)
	}
	
	if nodeID != "" {
		mc.getOrCreateNodeMetrics(nodeID).recordQuery(duration, success, mc.slowQueryThreshold)
	}
	
	// 更新时间窗口指标
	mc.updateTimeWindowQuery(duration, success)
}

// RecordQueryDetails 记录详细查询信息
func (mc *MetricsCollector) RecordQueryDetails(metrics *QueryMetrics) {
	mc.RecordQuery(metrics.NodeID, metrics.Duration, metrics.Success)
	
	// 可以在这里添加详细的查询日志记录
	if !metrics.Success {
		globalConfig.Warnf("Query failed on node %s: %s (duration: %v, error: %s)", 
			metrics.NodeID, metrics.SQL, metrics.Duration, metrics.Error)
	} else if metrics.Duration >= mc.slowQueryThreshold {
		globalConfig.Warnf("Slow query detected on node %s: %s (duration: %v)", 
			metrics.NodeID, metrics.SQL, metrics.Duration)
	}
}

// getOrCreateNodeMetrics 获取或创建节点指标
func (mc *MetricsCollector) getOrCreateNodeMetrics(nodeID string) *NodeMetrics {
	mc.nodeMetricsMutex.RLock()
	if metrics, exists := mc.nodeMetrics[nodeID]; exists {
		mc.nodeMetricsMutex.RUnlock()
		return metrics
	}
	mc.nodeMetricsMutex.RUnlock()
	
	mc.nodeMetricsMutex.Lock()
	defer mc.nodeMetricsMutex.Unlock()
	
	// 双重检查
	if metrics, exists := mc.nodeMetrics[nodeID]; exists {
		return metrics
	}
	
	metrics := &NodeMetrics{
		NodeID:              nodeID,
		LastConnectionTime:  time.Now(),
		LastQueryTime:       time.Now(),
		AverageResponseTime: 0,
	}
	
	mc.nodeMetrics[nodeID] = metrics
	return metrics
}

// recordConnection 记录节点连接
func (nm *NodeMetrics) recordConnection() {
	atomic.AddInt64(&nm.TotalConnections, 1)
	atomic.AddInt64(&nm.ActiveConnections, 1)
	nm.LastConnectionTime = time.Now()
}

// recordConnectionError 记录节点连接错误
func (nm *NodeMetrics) recordConnectionError() {
	atomic.AddInt64(&nm.ConnectionErrors, 1)
	atomic.AddInt64(&nm.ActiveConnections, -1)
}

// recordQuery 记录节点查询
func (nm *NodeMetrics) recordQuery(duration time.Duration, success bool, slowThreshold time.Duration) {
	atomic.AddInt64(&nm.TotalQueries, 1)
	
	if !success {
		atomic.AddInt64(&nm.FailedQueries, 1)
	}
	
	if duration >= slowThreshold {
		atomic.AddInt64(&nm.SlowQueries, 1)
	}
	
	nm.LastQueryTime = time.Now()
	
	// 更新平均响应时间（简化计算）
	totalQueries := atomic.LoadInt64(&nm.TotalQueries)
	if totalQueries > 0 {
		currentAvg := nm.AverageResponseTime
		nm.AverageResponseTime = time.Duration((int64(currentAvg)*(totalQueries-1) + int64(duration)) / totalQueries)
	}
}

// updateTimeWindowQuery 更新时间窗口查询指标
func (mc *MetricsCollector) updateTimeWindowQuery(duration time.Duration, success bool) {
	windows := []*TimeWindowMetrics{
		mc.windowMetrics.OneMinute,
		mc.windowMetrics.FiveMinutes,
		mc.windowMetrics.FifteenMinutes,
		mc.windowMetrics.OneHour,
	}
	
	for _, window := range windows {
		atomic.AddInt64(&window.TotalQueries, 1)
		
		if !success {
			atomic.AddInt64(&window.FailedQueries, 1)
		}
		
		if duration >= mc.slowQueryThreshold {
			atomic.AddInt64(&window.SlowQueries, 1)
		}
		
		// 更新响应时间统计
		if duration > window.MaxResponseTime {
			window.MaxResponseTime = duration
		}
		
		if duration < window.MinResponseTime {
			window.MinResponseTime = duration
		}
		
		// 更新平均响应时间
		totalQueries := atomic.LoadInt64(&window.TotalQueries)
		if totalQueries > 0 {
			currentAvg := window.AverageResponseTime
			window.AverageResponseTime = time.Duration((int64(currentAvg)*(totalQueries-1) + int64(duration)) / totalQueries)
		}
	}
}

// updateWindowMetrics 更新时间窗口指标
func (mc *MetricsCollector) updateWindowMetrics() {
	now := time.Now()
	
	// 检查并重置过期的时间窗口
	mc.checkAndResetWindow(mc.windowMetrics.OneMinute, now, time.Minute)
	mc.checkAndResetWindow(mc.windowMetrics.FiveMinutes, now, time.Minute*5)
	mc.checkAndResetWindow(mc.windowMetrics.FifteenMinutes, now, time.Minute*15)
	mc.checkAndResetWindow(mc.windowMetrics.OneHour, now, time.Hour)
	
	// 计算QPS等派生指标
	mc.calculateDerivedMetrics()
}

// checkAndResetWindow 检查并重置时间窗口
func (mc *MetricsCollector) checkAndResetWindow(window *TimeWindowMetrics, now time.Time, duration time.Duration) {
	if now.Sub(window.StartTime) >= duration {
		window.EndTime = now
		
		// 计算最终的派生指标
		elapsed := window.EndTime.Sub(window.StartTime)
		if elapsed > 0 {
			window.QPS = float64(window.TotalQueries) / elapsed.Seconds()
			window.ConnectionsPerSecond = float64(atomic.LoadInt64(&mc.totalConnections)) / elapsed.Seconds()
		}
		
		// 重置窗口
		*window = TimeWindowMetrics{
			StartTime:       now,
			MinResponseTime: time.Hour,
		}
	}
}

// calculateDerivedMetrics 计算派生指标
func (mc *MetricsCollector) calculateDerivedMetrics() {
	// 计算实时QPS（基于最近的查询）
	windows := []*TimeWindowMetrics{
		mc.windowMetrics.OneMinute,
		mc.windowMetrics.FiveMinutes,
		mc.windowMetrics.FifteenMinutes,
		mc.windowMetrics.OneHour,
	}
	
	for _, window := range windows {
		elapsed := time.Since(window.StartTime)
		if elapsed > 0 {
			window.QPS = float64(window.TotalQueries) / elapsed.Seconds()
		}
	}
}

// GetMetrics 获取所有指标
func (mc *MetricsCollector) GetMetrics() *MetricsSnapshot {
	mc.nodeMetricsMutex.RLock()
	nodeMetrics := make(map[string]*NodeMetrics)
	for k, v := range mc.nodeMetrics {
		// 创建副本
		nodeMetrics[k] = &NodeMetrics{
			NodeID:                v.NodeID,
			TotalConnections:      atomic.LoadInt64(&v.TotalConnections),
			ActiveConnections:     atomic.LoadInt64(&v.ActiveConnections),
			TotalQueries:          atomic.LoadInt64(&v.TotalQueries),
			SlowQueries:           atomic.LoadInt64(&v.SlowQueries),
			FailedQueries:         atomic.LoadInt64(&v.FailedQueries),
			ConnectionErrors:      atomic.LoadInt64(&v.ConnectionErrors),
			LastConnectionTime:    v.LastConnectionTime,
			LastQueryTime:         v.LastQueryTime,
			AverageResponseTime:   v.AverageResponseTime,
			HealthCheckCount:      atomic.LoadInt64(&v.HealthCheckCount),
			HealthCheckFailures:   atomic.LoadInt64(&v.HealthCheckFailures),
		}
	}
	mc.nodeMetricsMutex.RUnlock()
	
	return &MetricsSnapshot{
		Timestamp:           time.Now(),
		TotalConnections:    atomic.LoadInt64(&mc.totalConnections),
		ActiveConnections:   atomic.LoadInt64(&mc.activeConnections),
		ConnectionErrors:    atomic.LoadInt64(&mc.connectionErrors),
		ConnectionTimeouts:  atomic.LoadInt64(&mc.connectionTimeouts),
		TotalQueries:        atomic.LoadInt64(&mc.totalQueries),
		SlowQueries:         atomic.LoadInt64(&mc.slowQueries),
		FailedQueries:       atomic.LoadInt64(&mc.failedQueries),
		AverageQueryTime:    mc.calculateAverageQueryTime(),
		NodeMetrics:         nodeMetrics,
		WindowMetrics:       mc.copyWindowMetrics(),
	}
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	Timestamp           time.Time                `json:"timestamp"`
	TotalConnections    int64                    `json:"total_connections"`
	ActiveConnections   int64                    `json:"active_connections"`
	ConnectionErrors    int64                    `json:"connection_errors"`
	ConnectionTimeouts  int64                    `json:"connection_timeouts"`
	TotalQueries        int64                    `json:"total_queries"`
	SlowQueries         int64                    `json:"slow_queries"`
	FailedQueries       int64                    `json:"failed_queries"`
	AverageQueryTime    time.Duration            `json:"average_query_time"`
	NodeMetrics         map[string]*NodeMetrics  `json:"node_metrics"`
	WindowMetrics       *WindowMetrics           `json:"window_metrics"`
}

// calculateAverageQueryTime 计算平均查询时间
func (mc *MetricsCollector) calculateAverageQueryTime() time.Duration {
	totalQueries := atomic.LoadInt64(&mc.totalQueries)
	if totalQueries == 0 {
		return 0
	}
	
	totalDuration := atomic.LoadInt64(&mc.queryDurationTotal)
	return time.Duration(totalDuration / totalQueries)
}

// copyWindowMetrics 复制时间窗口指标
func (mc *MetricsCollector) copyWindowMetrics() *WindowMetrics {
	return &WindowMetrics{
		OneMinute:      mc.copyTimeWindowMetrics(mc.windowMetrics.OneMinute),
		FiveMinutes:    mc.copyTimeWindowMetrics(mc.windowMetrics.FiveMinutes),
		FifteenMinutes: mc.copyTimeWindowMetrics(mc.windowMetrics.FifteenMinutes),
		OneHour:        mc.copyTimeWindowMetrics(mc.windowMetrics.OneHour),
	}
}

// copyTimeWindowMetrics 复制时间窗口指标
func (mc *MetricsCollector) copyTimeWindowMetrics(original *TimeWindowMetrics) *TimeWindowMetrics {
	return &TimeWindowMetrics{
		StartTime:             original.StartTime,
		EndTime:               original.EndTime,
		TotalConnections:      atomic.LoadInt64(&original.TotalConnections),
		TotalQueries:          atomic.LoadInt64(&original.TotalQueries),
		SlowQueries:           atomic.LoadInt64(&original.SlowQueries),
		FailedQueries:         atomic.LoadInt64(&original.FailedQueries),
		AverageResponseTime:   original.AverageResponseTime,
		MaxResponseTime:       original.MaxResponseTime,
		MinResponseTime:       original.MinResponseTime,
		QPS:                   original.QPS,
		ConnectionsPerSecond:  original.ConnectionsPerSecond,
	}
}

// ToJSON 转换为JSON
func (ms *MetricsSnapshot) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ms, "", "  ")
}

// PrintMetrics 打印指标
func (ms *MetricsSnapshot) PrintMetrics() {
	globalConfig.Infof("=== Database Connection Pool Metrics ===")
	globalConfig.Infof("Timestamp: %s", ms.Timestamp.Format("2006-01-02 15:04:05"))
	globalConfig.Infof("Total Connections: %d", ms.TotalConnections)
	globalConfig.Infof("Active Connections: %d", ms.ActiveConnections)
	globalConfig.Infof("Connection Errors: %d", ms.ConnectionErrors)
	globalConfig.Infof("Total Queries: %d", ms.TotalQueries)
	globalConfig.Infof("Slow Queries: %d", ms.SlowQueries)
	globalConfig.Infof("Failed Queries: %d", ms.FailedQueries)
	globalConfig.Infof("Average Query Time: %v", ms.AverageQueryTime)
	
	globalConfig.Infof("=== Node Metrics ===")
	for nodeID, metrics := range ms.NodeMetrics {
		globalConfig.Infof("Node %s:", nodeID)
		globalConfig.Infof("  Total Connections: %d", metrics.TotalConnections)
		globalConfig.Infof("  Active Connections: %d", metrics.ActiveConnections)
		globalConfig.Infof("  Total Queries: %d", metrics.TotalQueries)
		globalConfig.Infof("  Slow Queries: %d", metrics.SlowQueries)
		globalConfig.Infof("  Failed Queries: %d", metrics.FailedQueries)
		globalConfig.Infof("  Average Response Time: %v", metrics.AverageResponseTime)
	}
	
	globalConfig.Infof("=== Window Metrics ===")
	globalConfig.Infof("1 Minute - QPS: %.2f, Queries: %d", 
		ms.WindowMetrics.OneMinute.QPS, ms.WindowMetrics.OneMinute.TotalQueries)
	globalConfig.Infof("5 Minutes - QPS: %.2f, Queries: %d", 
		ms.WindowMetrics.FiveMinutes.QPS, ms.WindowMetrics.FiveMinutes.TotalQueries)
	globalConfig.Infof("15 Minutes - QPS: %.2f, Queries: %d", 
		ms.WindowMetrics.FifteenMinutes.QPS, ms.WindowMetrics.FifteenMinutes.TotalQueries)
	globalConfig.Infof("1 Hour - QPS: %.2f, Queries: %d", 
		ms.WindowMetrics.OneHour.QPS, ms.WindowMetrics.OneHour.TotalQueries)
}

// ============= 全局指标收集器 =============

var (
	globalMetricsCollector *MetricsCollector
	metricsOnce            sync.Once
)

// GetGlobalMetricsCollector 获取全局指标收集器
func GetGlobalMetricsCollector() *MetricsCollector {
	metricsOnce.Do(func() {
		globalMetricsCollector = NewMetricsCollector()
		globalMetricsCollector.Start()
	})
	return globalMetricsCollector
}

// RecordGlobalConnection 记录全局连接
func RecordGlobalConnection(nodeID, nodeType string) {
	GetGlobalMetricsCollector().RecordConnection(nodeID, nodeType)
}

// RecordGlobalQuery 记录全局查询
func RecordGlobalQuery(nodeID string, duration time.Duration, success bool) {
	GetGlobalMetricsCollector().RecordQuery(nodeID, duration, success)
}

// GetGlobalMetrics 获取全局指标
func GetGlobalMetrics() *MetricsSnapshot {
	return GetGlobalMetricsCollector().GetMetrics()
}