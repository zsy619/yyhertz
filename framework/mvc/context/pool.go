// Package context provides a unified context abstraction with object pooling support.
// 
// This file has been refactored for single responsibility. The original pool.go
// has been split into multiple files:
// - params.go: Route parameter handling
// - context.go: Core context structure and methods 
// - compat_beego.go: Beego-style compatibility support
// - response_writer.go: Response writer implementation
// - batch.go: Batch processing functionality
// - adapter.go: Compatibility adapters
// - pool.go: Object pool management (current file)
package context

import (
	"sync"
	"sync/atomic"
	"time"
)

// ContextPool Context对象池，用于减少内存分配
type ContextPool struct {
	pool    sync.Pool
	metrics PoolMetrics
}

// PoolMetrics 池化性能指标
type PoolMetrics struct {
	Gets    int64 // 获取次数
	Puts    int64 // 放回次数
	News    int64 // 新建次数
	Reuses  int64 // 复用次数
	MaxSize int32 // 最大池大小
}

// 全局Context池
var (
	defaultPool = NewContextPool()
	poolSize    = int32(0)
	maxPoolSize = int32(1000) // 最大池大小
)

// NewContextPool 创建新的Context池
func NewContextPool() *ContextPool {
	pool := &ContextPool{}
	
	pool.pool.New = func() interface{} {
		atomic.AddInt64(&pool.metrics.News, 1)
		ctx := &Context{
			Keys:     make(map[string]interface{}),
			index:    -1,
			pooled:   true,
			acquired: time.Now(),
		}
		// 初始化兼容性字段（这些字段会在NewContext中重新赋值，这里先设置为避免nil指针）
		ctx.Input = &InputData{ctx: ctx}
		ctx.Output = &OutputData{ctx: ctx}
		return ctx
	}
	
	return pool
}

// Get 从池中获取Context
func (pool *ContextPool) Get() *Context {
	atomic.AddInt64(&pool.metrics.Gets, 1)
	
	ctx := pool.pool.Get().(*Context)
	if ctx.Keys != nil {
		atomic.AddInt64(&pool.metrics.Reuses, 1)
	}
	
	ctx.acquired = time.Now()
	return ctx
}

// Put 将Context放回池中
func (pool *ContextPool) Put(ctx *Context) {
	if ctx == nil || !ctx.pooled {
		return
	}
	
	atomic.AddInt64(&pool.metrics.Puts, 1)
	
	// 检查池大小限制
	if atomic.LoadInt32(&poolSize) >= maxPoolSize {
		return
	}
	
	// 重置Context状态
	ctx.Reset()
	atomic.AddInt32(&poolSize, 1)
	pool.pool.Put(ctx)
}

// GetDefaultPool 获取默认池
func GetDefaultPool() *ContextPool {
	return defaultPool
}

// GetPoolMetrics 获取池性能指标
func GetPoolMetrics() PoolMetrics {
	return PoolMetrics{
		Gets:    atomic.LoadInt64(&defaultPool.metrics.Gets),
		Puts:    atomic.LoadInt64(&defaultPool.metrics.Puts),
		News:    atomic.LoadInt64(&defaultPool.metrics.News),
		Reuses:  atomic.LoadInt64(&defaultPool.metrics.Reuses),
		MaxSize: atomic.LoadInt32(&maxPoolSize),
	}
}

// SetMaxPoolSize 设置最大池大小
func SetMaxPoolSize(size int32) {
	atomic.StoreInt32(&maxPoolSize, size)
}

// GetCurrentPoolSize 获取当前池大小
func GetCurrentPoolSize() int32 {
	return atomic.LoadInt32(&poolSize)
}

