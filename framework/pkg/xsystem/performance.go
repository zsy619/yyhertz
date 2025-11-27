package xsystem

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics   map[string]*MetricData
	mutex     sync.RWMutex
	startTime time.Time
}

// MetricData 指标数据 - 使用原子操作优化
type MetricData struct {
	count      int64 // 使用atomic操作
	totalNanos int64 // 总时间纳秒，使用atomic操作
	minNanos   int64 // 最小时间纳秒，使用atomic操作
	maxNanos   int64 // 最大时间纳秒，使用atomic操作
	lastCall   int64 // 最后调用时间戳，使用atomic操作
}

// 提供线程安全的访问方法
func (m *MetricData) Count() int64 {
	return atomic.LoadInt64(&m.count)
}

func (m *MetricData) Total() time.Duration {
	return time.Duration(atomic.LoadInt64(&m.totalNanos))
}

func (m *MetricData) Average() time.Duration {
	count := atomic.LoadInt64(&m.count)
	if count == 0 {
		return 0
	}
	total := atomic.LoadInt64(&m.totalNanos)
	return time.Duration(total / count)
}

func (m *MetricData) Min() time.Duration {
	return time.Duration(atomic.LoadInt64(&m.minNanos))
}

func (m *MetricData) Max() time.Duration {
	return time.Duration(atomic.LoadInt64(&m.maxNanos))
}

func (m *MetricData) LastCall() time.Time {
	timestamp := atomic.LoadInt64(&m.lastCall)
	if timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(0, timestamp)
}

// TimeMeasurement 时间测量结构
type TimeMeasurement struct {
	startTime time.Time
	monitor   *PerformanceMonitor
	name      string
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:   make(map[string]*MetricData),
		startTime: time.Now(),
	}
}

// StartMeasurement 开始测量性能
func (pm *PerformanceMonitor) StartMeasurement(name string) *TimeMeasurement {
	return &TimeMeasurement{
		startTime: time.Now(),
		monitor:   pm,
		name:      name,
	}
}

// End 结束测量
func (tm *TimeMeasurement) End() time.Duration {
	duration := time.Since(tm.startTime)
	tm.monitor.recordMetric(tm.name, duration)
	return duration
}

// recordMetric 记录指标数据
func (pm *PerformanceMonitor) recordMetric(name string, duration time.Duration) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	metric, exists := pm.metrics[name]
	if !exists {
		metric = &MetricData{}
		pm.metrics[name] = metric
	}

	nanos := duration.Nanoseconds()
	now := time.Now().UnixNano()

	// 原子操作更新统计数据
	atomic.AddInt64(&metric.count, 1)
	atomic.AddInt64(&metric.totalNanos, nanos)
	atomic.StoreInt64(&metric.lastCall, now)

	// 更新最小值
	for {
		oldMin := atomic.LoadInt64(&metric.minNanos)
		if oldMin == 0 || nanos < oldMin {
			if atomic.CompareAndSwapInt64(&metric.minNanos, oldMin, nanos) {
				break
			}
		} else {
			break
		}
	}

	// 更新最大值
	for {
		oldMax := atomic.LoadInt64(&metric.maxNanos)
		if nanos > oldMax {
			if atomic.CompareAndSwapInt64(&metric.maxNanos, oldMax, nanos) {
				break
			}
		} else {
			break
		}
	}
}

// GetMetric 获取指标数据
func (pm *PerformanceMonitor) GetMetric(name string) *MetricData {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.metrics[name]
}

// GetAllMetrics 获取所有指标数据
func (pm *PerformanceMonitor) GetAllMetrics() map[string]*MetricData {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	result := make(map[string]*MetricData)
	for name, metric := range pm.metrics {
		result[name] = metric
	}
	return result
}

// Clear 清除所有指标
func (pm *PerformanceMonitor) Clear() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.metrics = make(map[string]*MetricData)
}

// Report 生成性能报告
func (pm *PerformanceMonitor) Report() string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if len(pm.metrics) == 0 {
		return "No performance metrics collected"
	}

	var report strings.Builder
	report.WriteString("Performance Report\n")
	report.WriteString("==================\n\n")

	uptime := time.Since(pm.startTime)
	report.WriteString(fmt.Sprintf("Monitor uptime: %v\n", uptime))
	report.WriteString(fmt.Sprintf("Total metrics: %d\n\n", len(pm.metrics)))

	// 表头
	report.WriteString(fmt.Sprintf("%-30s %10s %15s %15s %15s %15s\n",
		"Metric", "Count", "Total", "Average", "Min", "Max"))
	report.WriteString(strings.Repeat("-", 100) + "\n")

	// 数据行
	for name, metric := range pm.metrics {
		report.WriteString(fmt.Sprintf("%-30s %10d %15v %15v %15v %15v\n",
			name,
			metric.Count(),
			metric.Total(),
			metric.Average(),
			metric.Min(),
			metric.Max()))
	}

	return report.String()
}

// GetMemoryUsage 获取内存使用情况
func GetMemoryUsage() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// FormatMemoryUsage 格式化内存使用情况
func FormatMemoryUsage(m runtime.MemStats) string {
	return fmt.Sprintf(
		"Alloc: %d KB, TotalAlloc: %d KB, Sys: %d KB, NumGC: %d",
		bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

// GetGoroutineCount 获取当前goroutine数量
func GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

// GetCPUCount 获取CPU核心数
func GetCPUCount() int {
	return runtime.NumCPU()
}

// ForceGC 强制垃圾回收
func ForceGC() {
	runtime.GC()
}

// CPUProfile CPU性能分析
type CPUProfile struct {
	startTime time.Time
	samples   []CPUSample
	mutex     sync.Mutex
}

// CPUSample CPU采样数据
type CPUSample struct {
	Timestamp   time.Time
	Goroutines  int
	HeapAlloc   uint64
	HeapSys     uint64
	StackInuse  uint64
	StackSys    uint64
}

// NewCPUProfile 创建CPU性能分析器
func NewCPUProfile() *CPUProfile {
	return &CPUProfile{
		startTime: time.Now(),
		samples:   make([]CPUSample, 0),
	}
}

// Sample 采样当前状态
func (cp *CPUProfile) Sample() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	sample := CPUSample{
		Timestamp:   time.Now(),
		Goroutines:  runtime.NumGoroutine(),
		HeapAlloc:   m.HeapAlloc,
		HeapSys:     m.HeapSys,
		StackInuse:  m.StackInuse,
		StackSys:    m.StackSys,
	}

	cp.samples = append(cp.samples, sample)
}

// GetSamples 获取所有采样数据
func (cp *CPUProfile) GetSamples() []CPUSample {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	result := make([]CPUSample, len(cp.samples))
	copy(result, cp.samples)
	return result
}

// Clear 清除采样数据
func (cp *CPUProfile) Clear() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	cp.samples = cp.samples[:0]
}

// 全局性能监控器实例
var DefaultPerformanceMonitor = NewPerformanceMonitor()

// StartMeasurement 使用默认监控器开始测量
func StartMeasurement(name string) *TimeMeasurement {
	return DefaultPerformanceMonitor.StartMeasurement(name)
}

// GetMetric 使用默认监控器获取指标
func GetMetric(name string) *MetricData {
	return DefaultPerformanceMonitor.GetMetric(name)
}

// GetPerformanceReport 获取默认监控器的性能报告
func GetPerformanceReport() string {
	return DefaultPerformanceMonitor.Report()
}

// ClearMetrics 清除默认监控器的指标
func ClearMetrics() {
	DefaultPerformanceMonitor.Clear()
}