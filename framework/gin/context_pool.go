// Package gin - Context对象池和零分配优化
// 提供Context对象的池化管理，减少GC压力，提升性能
package gin

import (
	"sync"
	"github.com/cloudwego/hertz/pkg/app"
)

// ContextPool Context对象池
type ContextPool struct {
	pool sync.Pool
}

// NewContextPool 创建新的Context池
func NewContextPool() *ContextPool {
	return &ContextPool{
		pool: sync.Pool{
			New: func() any {
				return &Context{
					// 预分配slice容量以减少后续分配
					Keys:   make(map[string]any, 8),  // 预分配8个键值对容量
					Errors: make([]error, 0, 4),      // 预分配4个错误容量
					Params: make(Params, 0, 16),      // 预分配16个参数容量
				}
			},
		},
	}
}

// Get 从池中获取Context
func (p *ContextPool) Get() *Context {
	ctx := p.pool.Get().(*Context)
	// 重置Context状态
	ctx.reset()
	return ctx
}

// Put 将Context归还给池
func (p *ContextPool) Put(ctx *Context) {
	// 清理敏感数据
	ctx.cleanup()
	p.pool.Put(ctx)
}

// Context重置方法
func (c *Context) reset() {
	c.RequestContext = nil
	c.handlers = nil
	c.index = -1
	c.engine = nil
	
	// 重置但保持容量
	if c.Keys != nil {
		for k := range c.Keys {
			delete(c.Keys, k)
		}
	}
	c.Errors = c.Errors[:0]
	c.Params = c.Params[:0]
}

// Context清理方法
func (c *Context) cleanup() {
	// 清理可能的敏感数据
	c.RequestContext = nil
	c.handlers = nil
	c.engine = nil
	
	// 避免持有大量内存
	if c.Keys != nil && len(c.Keys) > 64 {
		c.Keys = make(map[string]any, 8)
	}
	if cap(c.Errors) > 32 {
		c.Errors = make([]error, 0, 4)
	}
	if cap(c.Params) > 64 {
		c.Params = make(Params, 0, 16)
	}
}

// 全局Context池实例
var defaultContextPool = NewContextPool()

// 零分配路由处理器
type ZeroAllocHandler struct {
	// 预分配的处理器函数池
	handlerPool sync.Pool
	
	// 参数池
	paramPool sync.Pool
}

// NewZeroAllocHandler 创建零分配处理器
func NewZeroAllocHandler() *ZeroAllocHandler {
	return &ZeroAllocHandler{
		handlerPool: sync.Pool{
			New: func() any {
				return make([]HandlerFunc, 0, 16) // 预分配16个处理函数容量
			},
		},
		paramPool: sync.Pool{
			New: func() any {
				return make(Params, 0, 16) // 预分配16个参数容量
			},
		},
	}
}

// GetHandlers 获取处理器切片
func (h *ZeroAllocHandler) GetHandlers() []HandlerFunc {
	handlers := h.handlerPool.Get().([]HandlerFunc)
	return handlers[:0] // 重置长度但保持容量
}

// PutHandlers 归还处理器切片
func (h *ZeroAllocHandler) PutHandlers(handlers []HandlerFunc) {
	if cap(handlers) > 64 {
		// 如果容量过大，不归还给池
		return
	}
	h.handlerPool.Put(handlers)
}

// GetParams 获取参数切片
func (h *ZeroAllocHandler) GetParams() Params {
	params := h.paramPool.Get().(Params)
	return params[:0] // 重置长度但保持容量
}

// PutParams 归还参数切片
func (h *ZeroAllocHandler) PutParams(params Params) {
	if cap(params) > 64 {
		// 如果容量过大，不归还给池
		return
	}
	h.paramPool.Put(params)
}

// 全局零分配处理器实例
var defaultZeroAllocHandler = NewZeroAllocHandler()

// Engine的零分配方法增强
func (engine *Engine) acquireContext(c *app.RequestContext, handlers []HandlerFunc) *Context {
	// 从池中获取Context
	ctx := defaultContextPool.Get()
	
	// 设置Context属性
	ctx.RequestContext = c
	ctx.handlers = handlers
	ctx.index = -1
	ctx.engine = engine
	
	// 从Hertz RequestContext中提取路由参数（零分配）
	if len(c.Params) > 0 {
		// 确保参数切片有足够容量
		if cap(ctx.Params) < len(c.Params) {
			ctx.Params = make(Params, len(c.Params))
		} else {
			ctx.Params = ctx.Params[:len(c.Params)]
		}
		
		// 复制参数（避免分配）
		for i, param := range c.Params {
			ctx.Params[i] = Param{
				Key:   param.Key,
				Value: param.Value,
			}
		}
	}
	
	return ctx
}

// Engine的Context归还方法
func (engine *Engine) releaseContext(ctx *Context) {
	defaultContextPool.Put(ctx)
}

// 优化的路由处理函数
func (engine *Engine) handleHTTPRequestOptimized(c *app.RequestContext) {
	// 使用高性能路由引擎处理请求
	if engine.router != nil {
		// 创建优化的Context
		ctx := engine.acquireContext(c, nil)
		defer engine.releaseContext(ctx)
		
		// 使用新的路由引擎处理
		engine.router.handleHTTPRequest(ctx)
	} else {
		// 回退到原有处理方式
		_ = engine.createContext(c, nil)
		// 处理请求...
	}
}

// 预热池子，在应用启动时调用
func (engine *Engine) WarmupPools() {
	// 预创建一些Context对象
	contexts := make([]*Context, 10)
	for i := range contexts {
		contexts[i] = defaultContextPool.Get()
	}
	
	// 归还给池
	for _, ctx := range contexts {
		defaultContextPool.Put(ctx)
	}
	
	// 预创建一些处理器切片
	handlers := make([][]HandlerFunc, 5)
	for i := range handlers {
		handlers[i] = defaultZeroAllocHandler.GetHandlers()
	}
	
	// 归还给池
	for _, h := range handlers {
		defaultZeroAllocHandler.PutHandlers(h)
	}
}

// 性能统计
type PoolStats struct {
	ContextPoolHits   uint64 // Context池命中次数
	ContextPoolMisses uint64 // Context池缺失次数
	HandlerPoolHits   uint64 // Handler池命中次数
	HandlerPoolMisses uint64 // Handler池缺失次数
	ParamPoolHits     uint64 // Param池命中次数
	ParamPoolMisses   uint64 // Param池缺失次数
}

// GetPoolStats 获取池统计信息
func GetPoolStats() *PoolStats {
	// 这里应该从实际的计数器获取统计信息
	// 为了简化，返回空统计
	return &PoolStats{}
}

// 内存使用优化
type MemoryOptimizer struct {
	// 最大内存使用量（字节）
	MaxMemoryUsage uint64
	
	// 当前内存使用量
	CurrentMemoryUsage uint64
	
	// GC触发阈值
	GCThreshold uint64
}

// NewMemoryOptimizer 创建内存优化器
func NewMemoryOptimizer() *MemoryOptimizer {
	return &MemoryOptimizer{
		MaxMemoryUsage: 1024 * 1024 * 100, // 100MB
		GCThreshold:    1024 * 1024 * 50,  // 50MB
	}
}

// CheckMemoryUsage 检查内存使用情况
func (m *MemoryOptimizer) CheckMemoryUsage() {
	// 这里应该检查实际的内存使用情况
	// 如果超过阈值，可以主动清理池中的对象
	if m.CurrentMemoryUsage > m.GCThreshold {
		m.cleanupPools()
	}
}

// cleanupPools 清理池中的对象
func (m *MemoryOptimizer) cleanupPools() {
	// 强制清理池中的部分对象以释放内存
	// 这里可以实现更复杂的清理策略
}

// 零分配字符串操作
type StringPool struct {
	pool sync.Pool
}

// NewStringPool 创建字符串池
func NewStringPool() *StringPool {
	return &StringPool{
		pool: sync.Pool{
			New: func() any {
				return make([]byte, 0, 256) // 预分配256字节
			},
		},
	}
}

// GetBuffer 获取字节缓冲区
func (s *StringPool) GetBuffer() []byte {
	buf := s.pool.Get().([]byte)
	return buf[:0] // 重置长度
}

// PutBuffer 归还字节缓冲区
func (s *StringPool) PutBuffer(buf []byte) {
	if cap(buf) > 4096 {
		// 如果缓冲区太大，不归还给池
		return
	}
	s.pool.Put(buf)
}

// 全局字符串池
var defaultStringPool = NewStringPool()