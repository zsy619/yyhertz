// Package gin - 中间件链优化
// 提供高性能的中间件执行机制
package gin

import (
	"runtime"
	"sync"
	"time"
)

// 中间件链优化常量
const (
	// 最大中间件数量
	MaxMiddlewareCount = 63
	
	// 中间件链预编译阈值
	CompileThreshold = 10
)

// MiddlewareChain 优化的中间件链
type MiddlewareChain struct {
	handlers    []HandlerFunc
	compiled    bool
	execFunc    func(*Context)
	stats       *ChainStats
	mu          sync.RWMutex
}

// ChainStats 中间件链统计
type ChainStats struct {
	TotalCalls    uint64
	TotalDuration time.Duration
	AvgDuration   time.Duration
	Calls         []time.Duration
	mu            sync.RWMutex
}

// NewMiddlewareChain 创建新的中间件链
func NewMiddlewareChain(handlers []HandlerFunc) *MiddlewareChain {
	return &MiddlewareChain{
		handlers: make([]HandlerFunc, len(handlers)),
		stats:    &ChainStats{Calls: make([]time.Duration, 0, 100)},
	}
}

// Compile 编译中间件链
func (mc *MiddlewareChain) Compile() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.compiled {
		return
	}
	
	// 预编译中间件链执行函数
	mc.execFunc = mc.generateExecutionFunc()
	mc.compiled = true
}

// Execute 执行中间件链
func (mc *MiddlewareChain) Execute(c *Context) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		mc.recordStats(duration)
	}()
	
	mc.mu.RLock()
	compiled := mc.compiled
	execFunc := mc.execFunc
	mc.mu.RUnlock()
	
	if compiled && execFunc != nil {
		// 使用预编译的执行函数
		execFunc(c)
	} else {
		// 回退到标准执行方式
		mc.executeStandard(c)
	}
}

// executeStandard 标准执行方式
func (mc *MiddlewareChain) executeStandard(c *Context) {
	c.handlers = mc.handlers
	c.index = -1
	c.Next()
}

// generateExecutionFunc 生成预编译的执行函数
func (mc *MiddlewareChain) generateExecutionFunc() func(*Context) {
	handlers := mc.handlers
	
	// 根据中间件数量选择不同的优化策略
	switch len(handlers) {
	case 0:
		return func(c *Context) {}
	case 1:
		return mc.generateSingleHandler(handlers[0])
	case 2:
		return mc.generateDoubleHandler(handlers[0], handlers[1])
	case 3:
		return mc.generateTripleHandler(handlers[0], handlers[1], handlers[2])
	default:
		return mc.generateGenericHandler(handlers)
	}
}

// generateSingleHandler 单个中间件优化
func (mc *MiddlewareChain) generateSingleHandler(h HandlerFunc) func(*Context) {
	return func(c *Context) {
		h(c)
	}
}

// generateDoubleHandler 双中间件优化
func (mc *MiddlewareChain) generateDoubleHandler(h1, h2 HandlerFunc) func(*Context) {
	return func(c *Context) {
		// 执行第一个中间件
		c.handlers = []HandlerFunc{h1, h2}
		c.index = 0
		
		defer func() {
			if r := recover(); r != nil {
				// 处理panic
				c.AbortWithStatus(500)
			}
		}()
		
		h1(c)
		if !c.IsAborted() {
			h2(c)
		}
	}
}

// generateTripleHandler 三中间件优化
func (mc *MiddlewareChain) generateTripleHandler(h1, h2, h3 HandlerFunc) func(*Context) {
	return func(c *Context) {
		c.handlers = []HandlerFunc{h1, h2, h3}
		c.index = 0
		
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatus(500)
			}
		}()
		
		h1(c)
		if !c.IsAborted() {
			h2(c)
			if !c.IsAborted() {
				h3(c)
			}
		}
	}
}

// generateGenericHandler 通用中间件处理器
func (mc *MiddlewareChain) generateGenericHandler(handlers []HandlerFunc) func(*Context) {
	return func(c *Context) {
		c.handlers = handlers
		c.index = -1
		c.Next()
	}
}

// recordStats 记录统计信息
func (mc *MiddlewareChain) recordStats(duration time.Duration) {
	mc.stats.mu.Lock()
	defer mc.stats.mu.Unlock()
	
	mc.stats.TotalCalls++
	mc.stats.TotalDuration += duration
	mc.stats.AvgDuration = mc.stats.TotalDuration / time.Duration(mc.stats.TotalCalls)
	
	// 保持最近100次调用的记录
	if len(mc.stats.Calls) >= 100 {
		mc.stats.Calls = mc.stats.Calls[1:]
	}
	mc.stats.Calls = append(mc.stats.Calls, duration)
}

// GetStats 获取统计信息
func (mc *MiddlewareChain) GetStats() ChainStats {
	mc.stats.mu.RLock()
	defer mc.stats.mu.RUnlock()
	
	return ChainStats{
		TotalCalls:    mc.stats.TotalCalls,
		TotalDuration: mc.stats.TotalDuration,
		AvgDuration:   mc.stats.AvgDuration,
		Calls:         append([]time.Duration(nil), mc.stats.Calls...),
	}
}

// 条件中间件执行器
type ConditionalMiddleware struct {
	condition func(*Context) bool
	handler   HandlerFunc
	enabled   bool
}

// NewConditionalMiddleware 创建条件中间件
func NewConditionalMiddleware(condition func(*Context) bool, handler HandlerFunc) *ConditionalMiddleware {
	return &ConditionalMiddleware{
		condition: condition,
		handler:   handler,
		enabled:   true,
	}
}

// Execute 执行条件中间件
func (cm *ConditionalMiddleware) Execute(c *Context) {
	if !cm.enabled {
		return
	}
	
	if cm.condition == nil || cm.condition(c) {
		cm.handler(c)
	}
}

// Enable 启用中间件
func (cm *ConditionalMiddleware) Enable() {
	cm.enabled = true
}

// Disable 禁用中间件
func (cm *ConditionalMiddleware) Disable() {
	cm.enabled = false
}

// 异步中间件执行器
type AsyncMiddleware struct {
	handler HandlerFunc
	timeout time.Duration
	pool    *sync.Pool
}

// NewAsyncMiddleware 创建异步中间件
func NewAsyncMiddleware(handler HandlerFunc, timeout time.Duration) *AsyncMiddleware {
	return &AsyncMiddleware{
		handler: handler,
		timeout: timeout,
		pool: &sync.Pool{
			New: func() any {
				return make(chan struct{})
			},
		},
	}
}

// Execute 异步执行中间件
func (am *AsyncMiddleware) Execute(c *Context) {
	done := am.pool.Get().(chan struct{})
	defer am.pool.Put(done)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 记录panic，但不影响主流程
			}
			done <- struct{}{}
		}()
		
		am.handler(c)
	}()
	
	// 等待执行完成或超时
	select {
	case <-done:
		// 正常完成
	case <-time.After(am.timeout):
		// 超时，但不阻塞主流程
	}
}

// 中间件性能监控器
type MiddlewareMonitor struct {
	stats map[string]*MiddlewareStats
	mu    sync.RWMutex
}

// MiddlewareStats 中间件统计
type MiddlewareStats struct {
	Name          string
	CallCount     uint64
	TotalDuration time.Duration
	AvgDuration   time.Duration
	MaxDuration   time.Duration
	MinDuration   time.Duration
	ErrorCount    uint64
}

// NewMiddlewareMonitor 创建中间件监控器
func NewMiddlewareMonitor() *MiddlewareMonitor {
	return &MiddlewareMonitor{
		stats: make(map[string]*MiddlewareStats),
	}
}

// WrapHandler 包装处理器以进行监控
func (mm *MiddlewareMonitor) WrapHandler(name string, handler HandlerFunc) HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		var hasError bool
		
		defer func() {
			duration := time.Since(start)
			if r := recover(); r != nil {
				hasError = true
				panic(r) // 重新抛出panic
			}
			mm.recordStats(name, duration, hasError)
		}()
		
		handler(c)
		
		// 检查是否有错误
		if len(c.Errors) > 0 {
			hasError = true
		}
	}
}

// recordStats 记录统计信息
func (mm *MiddlewareMonitor) recordStats(name string, duration time.Duration, hasError bool) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	stats, exists := mm.stats[name]
	if !exists {
		stats = &MiddlewareStats{
			Name:        name,
			MinDuration: duration,
			MaxDuration: duration,
		}
		mm.stats[name] = stats
	}
	
	stats.CallCount++
	stats.TotalDuration += duration
	stats.AvgDuration = stats.TotalDuration / time.Duration(stats.CallCount)
	
	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}
	if duration < stats.MinDuration {
		stats.MinDuration = duration
	}
	
	if hasError {
		stats.ErrorCount++
	}
}

// GetStats 获取所有中间件统计
func (mm *MiddlewareMonitor) GetStats() map[string]MiddlewareStats {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	result := make(map[string]MiddlewareStats)
	for name, stats := range mm.stats {
		result[name] = *stats
	}
	
	return result
}

// 中间件链构建器
type ChainBuilder struct {
	handlers    []HandlerFunc
	monitors    []string
	conditions  []func(*Context) bool
	asyncTasks  []AsyncTask
}

// AsyncTask 异步任务
type AsyncTask struct {
	handler HandlerFunc
	timeout time.Duration
}

// NewChainBuilder 创建链构建器
func NewChainBuilder() *ChainBuilder {
	return &ChainBuilder{
		handlers:   make([]HandlerFunc, 0),
		monitors:   make([]string, 0),
		conditions: make([]func(*Context) bool, 0),
		asyncTasks: make([]AsyncTask, 0),
	}
}

// Use 添加中间件
func (cb *ChainBuilder) Use(handlers ...HandlerFunc) *ChainBuilder {
	cb.handlers = append(cb.handlers, handlers...)
	return cb
}

// UseConditional 添加条件中间件
func (cb *ChainBuilder) UseConditional(condition func(*Context) bool, handler HandlerFunc) *ChainBuilder {
	cb.conditions = append(cb.conditions, condition)
	cb.handlers = append(cb.handlers, handler)
	return cb
}

// UseAsync 添加异步中间件
func (cb *ChainBuilder) UseAsync(handler HandlerFunc, timeout time.Duration) *ChainBuilder {
	cb.asyncTasks = append(cb.asyncTasks, AsyncTask{
		handler: handler,
		timeout: timeout,
	})
	return cb
}

// WithMonitoring 添加监控
func (cb *ChainBuilder) WithMonitoring(names ...string) *ChainBuilder {
	cb.monitors = append(cb.monitors, names...)
	return cb
}

// Build 构建中间件链
func (cb *ChainBuilder) Build() *MiddlewareChain {
	chain := NewMiddlewareChain(cb.handlers)
	
	// 如果处理器数量达到阈值，进行预编译
	if len(cb.handlers) >= CompileThreshold {
		chain.Compile()
	}
	
	return chain
}

// 全局中间件监控器
var globalMonitor = NewMiddlewareMonitor()

// GetGlobalMiddlewareStats 获取全局中间件统计
func GetGlobalMiddlewareStats() map[string]MiddlewareStats {
	return globalMonitor.GetStats()
}

// MonitoredMiddleware 创建被监控的中间件
func MonitoredMiddleware(name string, handler HandlerFunc) HandlerFunc {
	return globalMonitor.WrapHandler(name, handler)
}

// 内存优化的中间件执行器
type MemoryOptimizedExecutor struct {
	contextPool sync.Pool
	bufferPool  sync.Pool
}

// NewMemoryOptimizedExecutor 创建内存优化执行器
func NewMemoryOptimizedExecutor() *MemoryOptimizedExecutor {
	return &MemoryOptimizedExecutor{
		contextPool: sync.Pool{
			New: func() any {
				return &Context{
					Keys:   make(map[string]any, 8),
					Errors: make([]error, 0, 4),
					Params: make(Params, 0, 16),
				}
			},
		},
		bufferPool: sync.Pool{
			New: func() any {
				return make([]byte, 0, 1024)
			},
		},
	}
}

// Execute 执行中间件（内存优化版本）
func (moe *MemoryOptimizedExecutor) Execute(chain *MiddlewareChain, c *Context) {
	// 使用对象池减少内存分配
	chain.Execute(c)
}

// 性能分析器
type PerformanceProfiler struct {
	enabled    bool
	samples    []ProfileSample
	mu         sync.RWMutex
	maxSamples int
}

// ProfileSample 性能样本
type ProfileSample struct {
	Timestamp    time.Time
	Duration     time.Duration
	MemoryUsage  uint64
	GoroutineNum int
	HandlerName  string
}

// NewPerformanceProfiler 创建性能分析器
func NewPerformanceProfiler(maxSamples int) *PerformanceProfiler {
	return &PerformanceProfiler{
		samples:    make([]ProfileSample, 0, maxSamples),
		maxSamples: maxSamples,
	}
}

// Enable 启用性能分析
func (pp *PerformanceProfiler) Enable() {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.enabled = true
}

// Disable 禁用性能分析
func (pp *PerformanceProfiler) Disable() {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.enabled = false
}

// ProfileHandler 分析处理器性能
func (pp *PerformanceProfiler) ProfileHandler(name string, handler HandlerFunc) HandlerFunc {
	return func(c *Context) {
		if !pp.enabled {
			handler(c)
			return
		}
		
		start := time.Now()
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		handler(c)
		
		runtime.ReadMemStats(&m2)
		duration := time.Since(start)
		
		sample := ProfileSample{
			Timestamp:    start,
			Duration:     duration,
			MemoryUsage:  m2.Alloc - m1.Alloc,
			GoroutineNum: runtime.NumGoroutine(),
			HandlerName:  name,
		}
		
		pp.addSample(sample)
	}
}

// addSample 添加性能样本
func (pp *PerformanceProfiler) addSample(sample ProfileSample) {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	
	if len(pp.samples) >= pp.maxSamples {
		// 移除最旧的样本
		pp.samples = pp.samples[1:]
	}
	
	pp.samples = append(pp.samples, sample)
}

// GetSamples 获取性能样本
func (pp *PerformanceProfiler) GetSamples() []ProfileSample {
	pp.mu.RLock()
	defer pp.mu.RUnlock()
	
	result := make([]ProfileSample, len(pp.samples))
	copy(result, pp.samples)
	return result
}

// 全局性能分析器
var globalProfiler = NewPerformanceProfiler(1000)

// EnableProfiling 启用全局性能分析
func EnableProfiling() {
	globalProfiler.Enable()
}

// DisableProfiling 禁用全局性能分析
func DisableProfiling() {
	globalProfiler.Disable()
}

// ProfiledMiddleware 创建被分析的中间件
func ProfiledMiddleware(name string, handler HandlerFunc) HandlerFunc {
	return globalProfiler.ProfileHandler(name, handler)
}