package util

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
	nanos := atomic.LoadInt64(&m.lastCall)
	return time.Unix(0, nanos)
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

// recordMetric 记录指标 - 使用原子操作优化
func (pm *PerformanceMonitor) recordMetric(name string, duration time.Duration) {
	durationNanos := duration.Nanoseconds()
	now := time.Now().UnixNano()

	pm.mutex.RLock()
	metric, exists := pm.metrics[name]
	pm.mutex.RUnlock()

	if !exists {
		// 需要创建新metric，使用写锁
		pm.mutex.Lock()
		// 双重检查
		if metric, exists = pm.metrics[name]; !exists {
			metric = &MetricData{
				minNanos: durationNanos,
				maxNanos: durationNanos,
			}
			pm.metrics[name] = metric
		}
		pm.mutex.Unlock()
	}

	// 使用原子操作更新，无锁
	atomic.AddInt64(&metric.count, 1)
	atomic.AddInt64(&metric.totalNanos, durationNanos)
	atomic.StoreInt64(&metric.lastCall, now)

	// 原子更新min/max值
	for {
		oldMin := atomic.LoadInt64(&metric.minNanos)
		if durationNanos >= oldMin {
			break
		}
		if atomic.CompareAndSwapInt64(&metric.minNanos, oldMin, durationNanos) {
			break
		}
	}

	for {
		oldMax := atomic.LoadInt64(&metric.maxNanos)
		if durationNanos <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&metric.maxNanos, oldMax, durationNanos) {
			break
		}
	}
}

// GetMetrics 获取所有指标
func (pm *PerformanceMonitor) GetMetrics() map[string]*MetricData {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]*MetricData)
	for k, v := range pm.metrics {
		// 返回原始metric的引用，因为所有访问都通过方法进行
		result[k] = v
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

	// 返回原始metric的引用
	return metric, true
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

// BytePool 字节池，使用固定大小池减少锁竞争
type BytePool struct {
	// 预定义常用大小的池，避免map查找和读写锁
	pool1K    sync.Pool // 1KB
	pool2K    sync.Pool // 2KB
	pool4K    sync.Pool // 4KB
	pool8K    sync.Pool // 8KB
	pool16K   sync.Pool // 16KB
	pool32K   sync.Pool // 32KB
	pool64K   sync.Pool // 64KB
	poolLarge sync.Pool // 大于64K的缓冲区
}

// byteSliceWrapper 包装[]byte切片以避免sync.Pool的额外分配
type byteSliceWrapper struct {
	buf []byte
}

// NewBytePool 创建字节池
func NewBytePool() *BytePool {
	bp := &BytePool{}

	// 初始化各个大小的池
	bp.pool1K.New = func() any { return &byteSliceWrapper{buf: make([]byte, 1024)} }
	bp.pool2K.New = func() any { return &byteSliceWrapper{buf: make([]byte, 2048)} }
	bp.pool4K.New = func() any { return &byteSliceWrapper{buf: make([]byte, 4096)} }
	bp.pool8K.New = func() any { return &byteSliceWrapper{buf: make([]byte, 8192)} }
	bp.pool16K.New = func() any { return &byteSliceWrapper{buf: make([]byte, 16384)} }
	bp.pool32K.New = func() any { return &byteSliceWrapper{buf: make([]byte, 32768)} }
	bp.pool64K.New = func() any { return &byteSliceWrapper{buf: make([]byte, 65536)} }
	bp.poolLarge.New = func() any { return &byteSliceWrapper{buf: make([]byte, 0, 131072)} } // 128K初始容量

	return bp
}

// Get 获取指定大小的字节切片 - 无锁快速路径
func (bp *BytePool) Get(size int) []byte {
	var wrapper *byteSliceWrapper
	switch {
	case size <= 1024:
		wrapper = bp.pool1K.Get().(*byteSliceWrapper)
		wrapper.buf = wrapper.buf[:size]
		return wrapper.buf
	case size <= 2048:
		wrapper = bp.pool2K.Get().(*byteSliceWrapper)
		wrapper.buf = wrapper.buf[:size]
		return wrapper.buf
	case size <= 4096:
		wrapper = bp.pool4K.Get().(*byteSliceWrapper)
		wrapper.buf = wrapper.buf[:size]
		return wrapper.buf
	case size <= 8192:
		wrapper = bp.pool8K.Get().(*byteSliceWrapper)
		wrapper.buf = wrapper.buf[:size]
		return wrapper.buf
	case size <= 16384:
		wrapper = bp.pool16K.Get().(*byteSliceWrapper)
		wrapper.buf = wrapper.buf[:size]
		return wrapper.buf
	case size <= 32768:
		wrapper = bp.pool32K.Get().(*byteSliceWrapper)
		wrapper.buf = wrapper.buf[:size]
		return wrapper.buf
	case size <= 65536:
		wrapper = bp.pool64K.Get().(*byteSliceWrapper)
		wrapper.buf = wrapper.buf[:size]
		return wrapper.buf
	default:
		// 大缓冲区处理
		wrapper = bp.poolLarge.Get().(*byteSliceWrapper)
		if cap(wrapper.buf) < size {
			wrapper.buf = make([]byte, size)
		} else {
			wrapper.buf = wrapper.buf[:size]
		}
		return wrapper.buf
	}
}

// Put 归还字节切片 - 无锁快速路径
func (bp *BytePool) Put(buf []byte) {
	if buf == nil {
		return
	}

	capacity := cap(buf)

	// 清零切片内容（安全考虑）
	for i := range buf[:] {
		buf[i] = 0
	}

	// 根据容量归还到对应的池
	switch {
	case capacity == 1024:
		bp.pool1K.Put(&byteSliceWrapper{buf: buf})
	case capacity == 2048:
		bp.pool2K.Put(&byteSliceWrapper{buf: buf})
	case capacity == 4096:
		bp.pool4K.Put(&byteSliceWrapper{buf: buf})
	case capacity == 8192:
		bp.pool8K.Put(&byteSliceWrapper{buf: buf})
	case capacity == 16384:
		bp.pool16K.Put(&byteSliceWrapper{buf: buf})
	case capacity == 32768:
		bp.pool32K.Put(&byteSliceWrapper{buf: buf})
	case capacity == 65536:
		bp.pool64K.Put(&byteSliceWrapper{buf: buf})
	case capacity >= 131072:
		// 只回收大缓冲区，避免内存泄漏
		bp.poolLarge.Put(&byteSliceWrapper{buf: buf[:0]})
	}
	// 其他大小的缓冲区直接丢弃，让GC处理
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
	tokens     int64
	maxTokens  int64
	refillRate int64
	lastRefill time.Time
	mutex      sync.Mutex
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
	state            CBState
	failures         int64
	lastFailTime     time.Time
	mutex            sync.RWMutex
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
		state:            CBStateClosed,
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
	NumGC      uint32   `json:"numGC"`
	PauseTotal uint64   `json:"pauseTotal"`
	PauseNs    []uint64 `json:"pauseNs"`
	LastGC     uint64   `json:"lastGC"`
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
		"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"heap_alloc_mb":  float64(m.HeapAlloc) / 1024 / 1024,
		"heap_sys_mb":    float64(m.HeapSys) / 1024 / 1024,
	}
}

// 全局性能监控器实例
var (
	GlobalMonitor    = NewPerformanceMonitor()
	GlobalBytePool   = NewBytePool()
	GlobalStringPool = NewStringPool()
)
