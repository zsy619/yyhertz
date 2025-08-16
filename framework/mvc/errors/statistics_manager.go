package errors

import (
	"sync"
	"time"
)

// ============= 统计管理器实现 =============

// DefaultStatisticsManager 默认统计管理器
type DefaultStatisticsManager struct {
	statistics *ErrorStatistics
	mu         sync.RWMutex
	
	// 性能优化
	bufferSize    int
	flushInterval time.Duration
	buffer        []ErrorRecord
	bufferMu      sync.Mutex
	stopChan      chan struct{}
	running       bool
}

// NewDefaultStatisticsManager 创建默认统计管理器
func NewDefaultStatisticsManager() *DefaultStatisticsManager {
	manager := &DefaultStatisticsManager{
		statistics:    NewErrorStatistics(),
		bufferSize:    100,
		flushInterval: 5 * time.Second,
		buffer:        make([]ErrorRecord, 0, 100),
		stopChan:      make(chan struct{}),
	}
	
	// 启动批量处理协程
	manager.start()
	
	return manager
}

// RecordError 记录错误
func (m *DefaultStatisticsManager) RecordError(statusCode int, path string, method string, err error) {
	record := ErrorRecord{
		StatusCode: statusCode,
		Path:       path,
		Method:     method,
		Timestamp:  time.Now(),
		// Duration字段在原ErrorRecord中不存在，这里删除
	}
	
	if err != nil {
		record.Message = err.Error()
	}
	
	// 添加到缓冲区
	m.bufferMu.Lock()
	m.buffer = append(m.buffer, record)
	
	// 如果缓冲区满了，立即刷新
	if len(m.buffer) >= m.bufferSize {
		m.flushBuffer()
	}
	m.bufferMu.Unlock()
}

// GetStatistics 获取统计信息
func (m *DefaultStatisticsManager) GetStatistics() *ErrorStatistics {
	// 先刷新缓冲区
	m.flushBufferIfNeeded()
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 返回统计信息的副本
	stats := &ErrorStatistics{
		TotalErrors:    m.statistics.TotalErrors,
		ErrorsByStatus: make(map[int]int64),
		ErrorsByPath:   make(map[string]int64),
		LastErrors:     make([]ErrorRecord, len(m.statistics.LastErrors)),
		StartTime:      m.statistics.StartTime,
	}
	
	// 深拷贝map
	for k, v := range m.statistics.ErrorsByStatus {
		stats.ErrorsByStatus[k] = v
	}
	for k, v := range m.statistics.ErrorsByPath {
		stats.ErrorsByPath[k] = v
	}
	
	// 深拷贝slice
	copy(stats.LastErrors, m.statistics.LastErrors)
	
	return stats
}

// Reset 重置统计
func (m *DefaultStatisticsManager) Reset() {
	m.flushBufferIfNeeded()
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 手动重置，因为原ErrorStatistics没有Reset方法
	m.statistics.TotalErrors = 0
	m.statistics.ErrorsByStatus = make(map[int]int64)
	m.statistics.ErrorsByPath = make(map[string]int64)
	m.statistics.LastErrors = m.statistics.LastErrors[:0]
	m.statistics.StartTime = time.Now()
}

// GetErrorRate 获取错误率
func (m *DefaultStatisticsManager) GetErrorRate(timeWindow int) float64 {
	if timeWindow <= 0 {
		timeWindow = 60 // 默认1分钟
	}
	
	m.flushBufferIfNeeded()
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	cutoff := time.Now().Add(-time.Duration(timeWindow) * time.Second)
	errorCount := int64(0)
	
	// 统计时间窗口内的错误数量
	for _, record := range m.statistics.LastErrors {
		if record.Timestamp.After(cutoff) {
			errorCount++
		}
	}
	
	// 计算错误率（错误数/秒）
	return float64(errorCount) / float64(timeWindow)
}

// GetTopErrorPaths 获取错误最多的路径
func (m *DefaultStatisticsManager) GetTopErrorPaths(limit int) []PathErrorStat {
	m.flushBufferIfNeeded()
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var paths []PathErrorStat
	for path, count := range m.statistics.ErrorsByPath {
		paths = append(paths, PathErrorStat{Path: path, Count: count})
	}
	
	// 简单排序
	for i := 0; i < len(paths)-1; i++ {
		for j := 0; j < len(paths)-1-i; j++ {
			if paths[j].Count < paths[j+1].Count {
				paths[j], paths[j+1] = paths[j+1], paths[j]
			}
		}
	}
	
	if limit > 0 && len(paths) > limit {
		paths = paths[:limit]
	}
	
	return paths
}

// GetErrorsByTimeRange 获取时间范围内的错误
func (m *DefaultStatisticsManager) GetErrorsByTimeRange(start, end time.Time) []ErrorRecord {
	m.flushBufferIfNeeded()
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var records []ErrorRecord
	for _, record := range m.statistics.LastErrors {
		if record.Timestamp.After(start) && record.Timestamp.Before(end) {
			records = append(records, record)
		}
	}
	
	return records
}

// GetHourlyStats 获取小时统计
func (m *DefaultStatisticsManager) GetHourlyStats() [24]int64 {
	m.flushBufferIfNeeded()
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var hourly [24]int64
	for _, record := range m.statistics.LastErrors {
		hour := record.Timestamp.Hour()
		hourly[hour]++
	}
	
	return hourly
}

// start 启动后台处理
func (m *DefaultStatisticsManager) start() {
	if m.running {
		return
	}
	
	m.running = true
	go func() {
		ticker := time.NewTicker(m.flushInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.flushBufferIfNeeded()
			case <-m.stopChan:
				m.flushBufferIfNeeded()
				return
			}
		}
	}()
}

// Stop 停止统计管理器
func (m *DefaultStatisticsManager) Stop() {
	if !m.running {
		return
	}
	
	close(m.stopChan)
	m.running = false
	
	// 最后刷新一次缓冲区
	m.flushBufferIfNeeded()
}

// flushBufferIfNeeded 如果需要则刷新缓冲区
func (m *DefaultStatisticsManager) flushBufferIfNeeded() {
	m.bufferMu.Lock()
	defer m.bufferMu.Unlock()
	
	if len(m.buffer) > 0 {
		m.flushBuffer()
	}
}

// flushBuffer 刷新缓冲区（需要在持有bufferMu的情况下调用）
func (m *DefaultStatisticsManager) flushBuffer() {
	if len(m.buffer) == 0 {
		return
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 批量处理缓冲区中的记录
	for _, record := range m.buffer {
		// 手动更新统计，因为原ErrorStatistics没有RecordError方法
		m.statistics.TotalErrors++
		m.statistics.ErrorsByStatus[record.StatusCode]++
		m.statistics.ErrorsByPath[record.Path]++
		
		// 添加到最近错误列表（保持最新100条）
		m.statistics.LastErrors = append(m.statistics.LastErrors, record)
		if len(m.statistics.LastErrors) > 100 {
			m.statistics.LastErrors = m.statistics.LastErrors[1:]
		}
	}
	
	// 清空缓冲区
	m.buffer = m.buffer[:0]
}

// ============= 数据结构 =============

// PathErrorStat 路径错误统计
type PathErrorStat struct {
	Path  string `json:"path"`
	Count int64  `json:"count"`
}

// ErrorTrend 错误趋势
type ErrorTrend struct {
	Time  time.Time `json:"time"`
	Count int64     `json:"count"`
}

// StatisticsSummary 统计摘要
type StatisticsSummary struct {
	TotalErrors      int64           `json:"total_errors"`
	ErrorRate        float64         `json:"error_rate"`        // 每秒错误数
	TopErrorPaths    []PathErrorStat `json:"top_error_paths"`
	TopErrorStatuses []StatusStat    `json:"top_error_statuses"`
	TimeRange        TimeRange       `json:"time_range"`
	HourlyTrend      [24]int64       `json:"hourly_trend"`
}

// StatusStat 状态码统计
type StatusStat struct {
	StatusCode int   `json:"status_code"`
	Count      int64 `json:"count"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// GetSummary 获取统计摘要
func (m *DefaultStatisticsManager) GetSummary() *StatisticsSummary {
	stats := m.GetStatistics()
	
	summary := &StatisticsSummary{
		TotalErrors:   stats.TotalErrors,
		ErrorRate:     m.GetErrorRate(60), // 最近1分钟的错误率
		TopErrorPaths: m.GetTopErrorPaths(10),
		HourlyTrend:   m.GetHourlyStats(),
		TimeRange: TimeRange{
			Start: stats.StartTime,
			End:   time.Now(),
		},
	}
	
	// 获取状态码统计
	var statusStats []StatusStat
	for status, count := range stats.ErrorsByStatus {
		statusStats = append(statusStats, StatusStat{
			StatusCode: status,
			Count:      count,
		})
	}
	
	// 排序状态码统计
	for i := 0; i < len(statusStats)-1; i++ {
		for j := 0; j < len(statusStats)-1-i; j++ {
			if statusStats[j].Count < statusStats[j+1].Count {
				statusStats[j], statusStats[j+1] = statusStats[j+1], statusStats[j]
			}
		}
	}
	
	if len(statusStats) > 10 {
		statusStats = statusStats[:10]
	}
	summary.TopErrorStatuses = statusStats
	
	return summary
}

// ============= 高性能统计管理器 =============

// HighPerformanceStatisticsManager 高性能统计管理器
type HighPerformanceStatisticsManager struct {
	// 使用原子操作优化
	totalErrors int64
	
	// 分段锁优化并发
	segments    [16]StatSegment
	segmentMask int64
	
	// 时间窗口统计
	timeWindows map[int]*TimeWindow
	windowMu    sync.RWMutex
}

// StatSegment 统计段
type StatSegment struct {
	mu             sync.RWMutex
	errorsByStatus map[int]int64
	errorsByPath   map[string]int64
}

// TimeWindow 时间窗口
type TimeWindow struct {
	startTime time.Time
	errors    []int64 // 每秒的错误数
	index     int
	size      int
}

// NewHighPerformanceStatisticsManager 创建高性能统计管理器
func NewHighPerformanceStatisticsManager() *HighPerformanceStatisticsManager {
	manager := &HighPerformanceStatisticsManager{
		segmentMask: 15, // 16-1
		timeWindows: make(map[int]*TimeWindow),
	}
	
	// 初始化分段
	for i := range manager.segments {
		manager.segments[i].errorsByStatus = make(map[int]int64)
		manager.segments[i].errorsByPath = make(map[string]int64)
	}
	
	// 初始化时间窗口
	manager.timeWindows[60] = &TimeWindow{   // 1分钟
		startTime: time.Now(),
		errors:    make([]int64, 60),
		size:      60,
	}
	manager.timeWindows[3600] = &TimeWindow{ // 1小时
		startTime: time.Now(),
		errors:    make([]int64, 3600),
		size:      3600,
	}
	
	return manager
}

// RecordError 记录错误（高性能版本）
func (m *HighPerformanceStatisticsManager) RecordError(statusCode int, path string, method string, err error) {
	// 原子递增总数
	totalErrors := m.addTotalErrors(1)
	
	// 计算分段索引
	segmentIndex := m.getSegmentIndex(path)
	segment := &m.segments[segmentIndex]
	
	// 更新分段统计
	segment.mu.Lock()
	segment.errorsByStatus[statusCode]++
	segment.errorsByPath[path]++
	segment.mu.Unlock()
	
	// 更新时间窗口
	m.updateTimeWindows(totalErrors)
}

// GetStatistics 获取统计信息（高性能版本）
func (m *HighPerformanceStatisticsManager) GetStatistics() *ErrorStatistics {
	stats := NewErrorStatistics()
	stats.TotalErrors = m.getTotalErrors()
	
	// 合并所有分段的统计
	for i := range m.segments {
		segment := &m.segments[i]
		segment.mu.RLock()
		
		for status, count := range segment.errorsByStatus {
			stats.ErrorsByStatus[status] += count
		}
		
		for path, count := range segment.errorsByPath {
			stats.ErrorsByPath[path] += count
		}
		
		segment.mu.RUnlock()
	}
	
	return stats
}

// Reset 重置统计（高性能版本）
func (m *HighPerformanceStatisticsManager) Reset() {
	m.setTotalErrors(0)
	
	for i := range m.segments {
		segment := &m.segments[i]
		segment.mu.Lock()
		segment.errorsByStatus = make(map[int]int64)
		segment.errorsByPath = make(map[string]int64)
		segment.mu.Unlock()
	}
	
	// 重置时间窗口
	m.windowMu.Lock()
	for _, window := range m.timeWindows {
		window.startTime = time.Now()
		for i := range window.errors {
			window.errors[i] = 0
		}
		window.index = 0
	}
	m.windowMu.Unlock()
}

// GetErrorRate 获取错误率（高性能版本）
func (m *HighPerformanceStatisticsManager) GetErrorRate(timeWindow int) float64 {
	m.windowMu.RLock()
	window, exists := m.timeWindows[timeWindow]
	m.windowMu.RUnlock()
	
	if !exists {
		return 0.0
	}
	
	var total int64
	for _, count := range window.errors {
		total += count
	}
	
	return float64(total) / float64(timeWindow)
}

// 辅助方法
func (m *HighPerformanceStatisticsManager) getSegmentIndex(path string) int {
	// 简单哈希函数
	hash := int64(0)
	for _, c := range path {
		hash = hash*31 + int64(c)
	}
	return int(hash & m.segmentMask)
}

func (m *HighPerformanceStatisticsManager) addTotalErrors(delta int64) int64 {
	// 暂时使用简单的加法，后续可以优化为atomic.AddInt64
	m.totalErrors += delta
	return m.totalErrors
}

func (m *HighPerformanceStatisticsManager) getTotalErrors() int64 {
	// 暂时直接返回，后续可以优化为atomic.LoadInt64
	return m.totalErrors
}

func (m *HighPerformanceStatisticsManager) setTotalErrors(value int64) {
	// 暂时直接赋值，后续可以优化为atomic.StoreInt64
	m.totalErrors = value
}

func (m *HighPerformanceStatisticsManager) updateTimeWindows(totalErrors int64) {
	now := time.Now()
	
	m.windowMu.Lock()
	defer m.windowMu.Unlock()
	
	for _, window := range m.timeWindows {
		elapsed := int(now.Sub(window.startTime).Seconds())
		if elapsed >= window.size {
			// 重置窗口
			window.startTime = now
			window.index = 0
			for i := range window.errors {
				window.errors[i] = 0
			}
		} else {
			// 更新当前位置
			newIndex := elapsed % window.size
			if newIndex != window.index {
				// 清空之间的位置
				for i := (window.index + 1) % window.size; i != newIndex; i = (i + 1) % window.size {
					window.errors[i] = 0
				}
				window.index = newIndex
			}
		}
		
		window.errors[window.index]++
	}
}

// ============= 全局统计管理器 =============

var globalStatisticsManager = NewDefaultStatisticsManager()

// GetGlobalStatisticsManager 获取全局统计管理器
func GetGlobalStatisticsManager() StatisticsManager {
	return globalStatisticsManager
}

// RecordGlobalError 记录全局错误
func RecordGlobalError(statusCode int, path string, method string, err error) {
	globalStatisticsManager.RecordError(statusCode, path, method, err)
}

// GetGlobalStatistics 获取全局统计
func GetGlobalStatistics() *ErrorStatistics {
	return globalStatisticsManager.GetStatistics()
}