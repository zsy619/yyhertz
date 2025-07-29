package util

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics   map[string]*MetricData
	mutex     sync.RWMutex
	startTime time.Time
}

// MetricData 指标数据
type MetricData struct {
	Count     int64         `json:"count"`
	Total     time.Duration `json:"total"`
	Average   time.Duration `json:"average"`
	Min       time.Duration `json:"min"`
	Max       time.Duration `json:"max"`
	LastCall  time.Time     `json:"lastCall"`
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:   make(map[string]*MetricData),
		startTime: time.Now(),
	}
}

// TrackExecution 跟踪方法执行时间
func (pm *PerformanceMonitor) TrackExecution(name string, fn func()) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		pm.recordMetric(name, duration)
	}()
	fn()
}

// StartTracking 开始跟踪
func (pm *PerformanceMonitor) StartTracking(name string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		pm.recordMetric(name, duration)
	}
}

// recordMetric 记录指标
func (pm *PerformanceMonitor) recordMetric(name string, duration time.Duration) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	metric, exists := pm.metrics[name]
	if !exists {
		metric = &MetricData{
			Min: duration,
			Max: duration,
		}
		pm.metrics[name] = metric
	}
	
	metric.Count++
	metric.Total += duration
	metric.Average = metric.Total / time.Duration(metric.Count)
	metric.LastCall = time.Now()
	
	if duration < metric.Min {
		metric.Min = duration
	}
	if duration > metric.Max {
		metric.Max = duration
	}
}

// GetMetrics 获取所有指标
func (pm *PerformanceMonitor) GetMetrics() map[string]*MetricData {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	result := make(map[string]*MetricData)
	for k, v := range pm.metrics {
		result[k] = &MetricData{
			Count:    v.Count,
			Total:    v.Total,
			Average:  v.Average,
			Min:      v.Min,
			Max:      v.Max,
			LastCall: v.LastCall,
		}
	}
	return result
}

// GetMetric 获取单个指标
func (pm *PerformanceMonitor) GetMetric(name string) (*MetricData, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	metric, exists := pm.metrics[name]
	if !exists {
		return nil, false
	}
	
	return &MetricData{
		Count:    metric.Count,
		Total:    metric.Total,
		Average:  metric.Average,
		Min:      metric.Min,
		Max:      metric.Max,
		LastCall: metric.LastCall,
	}, true
}

// Reset 重置所有指标
func (pm *PerformanceMonitor) Reset() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	pm.metrics = make(map[string]*MetricData)
	pm.startTime = time.Now()
}

// GetUptime 获取运行时间
func (pm *PerformanceMonitor) GetUptime() time.Duration {
	return time.Since(pm.startTime)
}

// MemoryPool 内存池
type MemoryPool[T any] struct {
	pool sync.Pool
	new  func() *T
}

// NewMemoryPool 创建内存池
func NewMemoryPool[T any](newFn func() *T) *MemoryPool[T] {
	return &MemoryPool[T]{
		pool: sync.Pool{
			New: func() any {
				if newFn != nil {
					return newFn()
				}
				var zero T
				return &zero
			},
		},
		new: newFn,
	}
}

// Get 从池中获取对象
func (mp *MemoryPool[T]) Get() *T {
	return mp.pool.Get().(*T)
}

// Put 将对象放回池中
func (mp *MemoryPool[T]) Put(obj *T) {
	// 可以在这里重置对象状态
	mp.pool.Put(obj)
}

// BytePool 字节池，用于减少[]byte分配
type BytePool struct {
	pools map[int]*sync.Pool
	mutex sync.RWMutex
}

// NewBytePool 创建字节池
func NewBytePool() *BytePool {
	return &BytePool{
		pools: make(map[int]*sync.Pool),
	}
}

// Get 获取指定大小的字节切片
func (bp *BytePool) Get(size int) []byte {
	bp.mutex.RLock()
	pool, exists := bp.pools[size]
	bp.mutex.RUnlock()
	
	if !exists {
		bp.mutex.Lock()
		// 双重检查
		if pool, exists = bp.pools[size]; !exists {
			pool = &sync.Pool{
				New: func() any {
					return make([]byte, size)
				},
			}
			bp.pools[size] = pool
		}
		bp.mutex.Unlock()
	}
	
	return pool.Get().([]byte)
}

// Put 归还字节切片
func (bp *BytePool) Put(buf []byte) {
	if buf == nil {
		return
	}
	
	size := cap(buf)
	bp.mutex.RLock()
	pool, exists := bp.pools[size]
	bp.mutex.RUnlock()
	
	if exists {
		// 清零切片
		for i := range buf[:len(buf)] {
			buf[i] = 0
		}
		buf = buf[:cap(buf)]
		pool.Put(buf)
	}
}

// StringPool 字符串构建器池
type StringPool struct {
	pool sync.Pool
}

// NewStringPool 创建字符串池
func NewStringPool() *StringPool {
	return &StringPool{
		pool: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
	}
}

// Get 获取字符串构建器
func (sp *StringPool) Get() *strings.Builder {
	return sp.pool.Get().(*strings.Builder)
}

// Put 归还字符串构建器
func (sp *StringPool) Put(sb *strings.Builder) {
	sb.Reset()
	sp.pool.Put(sb)
}

// RateLimiter 令牌桶限流器
type RateLimiter struct {
	tokens    int64
	maxTokens int64
	refillRate int64
	lastRefill time.Time
	mutex     sync.Mutex
}

// NewRateLimiter 创建限流器
func NewRateLimiter(maxTokens, refillRate int64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许通过
func (rl *RateLimiter) Allow() bool {
	return rl.AllowN(1)
}

// AllowN 检查是否允许N个请求通过
func (rl *RateLimiter) AllowN(n int64) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	
	// 添加令牌
	if elapsed > 0 {
		tokensToAdd := int64(elapsed.Seconds()) * rl.refillRate
		rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}
	
	// 检查是否有足够的令牌
	if rl.tokens >= n {
		rl.tokens -= n
		return true
	}
	
	return false
}

// GetTokens 获取当前令牌数
func (rl *RateLimiter) GetTokens() int64 {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return rl.tokens
}

// min 获取最小值
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	failureThreshold int64
	resetTimeout     time.Duration
	state           CBState
	failures        int64
	lastFailTime    time.Time
	mutex           sync.RWMutex
}

// CBState 熔断器状态
type CBState int

const (
	CBStateClosed CBState = iota
	CBStateOpen
	CBStateHalfOpen
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(failureThreshold int64, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:           CBStateClosed,
	}
}

// Call 执行调用
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker is open")
	}
	
	err := fn()
	cb.recordResult(err == nil)
	return err
}

// allowRequest 是否允许请求
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	now := time.Now()
	
	switch cb.state {
	case CBStateClosed:
		return true
	case CBStateOpen:
		if now.Sub(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CBStateHalfOpen
			return true
		}
		return false
	case CBStateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult 记录结果
func (cb *CircuitBreaker) recordResult(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if success {
		cb.failures = 0
		if cb.state == CBStateHalfOpen {
			cb.state = CBStateClosed
		}
	} else {
		cb.failures++
		cb.lastFailTime = time.Now()
		
		if cb.failures >= cb.failureThreshold {
			cb.state = CBStateOpen
		}
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CBState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// SystemInfo 系统信息
type SystemInfo struct {
	GoVersion    string `json:"goVersion"`
	NumCPU       int    `json:"numCPU"`
	NumGoroutine int    `json:"numGoroutine"`
	
	// 内存信息
	MemStats MemoryStats `json:"memStats"`
	
	// GC信息
	GCStats GCStats `json:"gcStats"`
}

// MemoryStats 内存统计
type MemoryStats struct {
	Alloc        uint64 `json:"alloc"`        // 当前分配的内存
	TotalAlloc   uint64 `json:"totalAlloc"`   // 总分配的内存
	Sys          uint64 `json:"sys"`          // 系统分配的内存
	NumGC        uint32 `json:"numGC"`        // GC次数
	HeapAlloc    uint64 `json:"heapAlloc"`    // 堆分配的内存
	HeapSys      uint64 `json:"heapSys"`      // 堆系统内存
	HeapInuse    uint64 `json:"heapInuse"`    // 堆正在使用的内存
	HeapReleased uint64 `json:"heapReleased"` // 堆释放的内存
}

// GCStats GC统计
type GCStats struct {
	NumGC        uint32  `json:"numGC"`
	PauseTotal   uint64  `json:"pauseTotal"`
	PauseNs      []uint64 `json:"pauseNs"`
	LastGC       uint64  `json:"lastGC"`
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() *SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return &SystemInfo{
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemStats: MemoryStats{
			Alloc:        m.Alloc,
			TotalAlloc:   m.TotalAlloc,
			Sys:          m.Sys,
			NumGC:        m.NumGC,
			HeapAlloc:    m.HeapAlloc,
			HeapSys:      m.HeapSys,
			HeapInuse:    m.HeapInuse,
			HeapReleased: m.HeapReleased,
		},
		GCStats: GCStats{
			NumGC:      m.NumGC,
			PauseTotal: m.PauseTotalNs,
			PauseNs:    m.PauseNs[:],
			LastGC:     m.LastGC,
		},
	}
}

// ForceGC 强制垃圾回收
func ForceGC() {
	runtime.GC()
}

// GetMemoryUsage 获取内存使用情况（MB）
func GetMemoryUsage() map[string]float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]float64{
		"alloc_mb":      float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":        float64(m.Sys) / 1024 / 1024,
		"heap_alloc_mb": float64(m.HeapAlloc) / 1024 / 1024,
		"heap_sys_mb":   float64(m.HeapSys) / 1024 / 1024,
	}
}

// 全局性能监控器实例
var (
	GlobalMonitor = NewPerformanceMonitor()
	GlobalBytePool = NewBytePool()
	GlobalStringPool = NewStringPool()
)