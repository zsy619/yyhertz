// Package metrics 提供应用监控和指标收集
package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 指标收集器
type Metrics struct {
	counters map[string]*int64
	gauges   map[string]*int64
	histograms map[string]*Histogram
	mutex    sync.RWMutex
}

// Histogram 直方图
type Histogram struct {
	buckets []int64
	mutex   sync.RWMutex
}

// NewMetrics 创建指标收集器
func NewMetrics() *Metrics {
	return &Metrics{
		counters:   make(map[string]*int64),
		gauges:     make(map[string]*int64),
		histograms: make(map[string]*Histogram),
	}
}

// Counter 增加计数器
func (m *Metrics) Counter(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if counter, exists := m.counters[name]; exists {
		atomic.AddInt64(counter, 1)
	} else {
		counter := int64(1)
		m.counters[name] = &counter
	}
}

// Gauge 设置测量值
func (m *Metrics) Gauge(name string, value int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if gauge, exists := m.gauges[name]; exists {
		atomic.StoreInt64(gauge, value)
	} else {
		m.gauges[name] = &value
	}
}

// GetCounter 获取计数器值
func (m *Metrics) GetCounter(name string) int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if counter, exists := m.counters[name]; exists {
		return atomic.LoadInt64(counter)
	}
	return 0
}

// GetGauge 获取测量值
func (m *Metrics) GetGauge(name string) int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if gauge, exists := m.gauges[name]; exists {
		return atomic.LoadInt64(gauge)
	}
	return 0
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	metrics *Metrics
	running int32
	stop    chan struct{}
}

// NewSystemMetrics 创建系统指标收集器
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		metrics: NewMetrics(),
		stop:    make(chan struct{}),
	}
}

// Start 开始收集系统指标
func (sm *SystemMetrics) Start() {
	if atomic.LoadInt32(&sm.running) == 1 {
		return
	}
	
	atomic.StoreInt32(&sm.running, 1)
	go sm.collectLoop()
}

// Stop 停止收集
func (sm *SystemMetrics) Stop() {
	if atomic.LoadInt32(&sm.running) == 0 {
		return
	}
	
	atomic.StoreInt32(&sm.running, 0)
	close(sm.stop)
}

// collectLoop 收集循环
func (sm *SystemMetrics) collectLoop() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.stop:
			return
		case <-ticker.C:
			sm.collectSystemMetrics()
		}
	}
}

// collectSystemMetrics 收集系统指标
func (sm *SystemMetrics) collectSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	sm.metrics.Gauge("memory.alloc", int64(m.Alloc))
	sm.metrics.Gauge("memory.sys", int64(m.Sys))
	sm.metrics.Gauge("goroutines", int64(runtime.NumGoroutine()))
	sm.metrics.Gauge("gc.runs", int64(m.NumGC))
}

// GetMetrics 获取指标
func (sm *SystemMetrics) GetMetrics() *Metrics {
	return sm.metrics
}

// 全局实例
var globalMetrics = NewSystemMetrics()

// Start 启动全局指标收集
func Start() {
	globalMetrics.Start()
}

// Stop 停止全局指标收集
func Stop() {
	globalMetrics.Stop()
}

// Counter 全局计数器
func Counter(name string) {
	globalMetrics.metrics.Counter(name)
}

// Gauge 全局测量值
func Gauge(name string, value int64) {
	globalMetrics.metrics.Gauge(name, value)
}

// GetCounter 获取全局计数器
func GetCounter(name string) int64 {
	return globalMetrics.metrics.GetCounter(name)
}

// GetGauge 获取全局测量值
func GetGauge(name string) int64 {
	return globalMetrics.metrics.GetGauge(name)
}