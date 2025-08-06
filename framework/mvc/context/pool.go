package context

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// Params 路由参数
type Params []Param

// Param 单个参数
type Param struct {
	Key   string
	Value string
}

// ByName 根据名称获取参数值
func (ps Params) ByName(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

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

// EnhancedContext 增强的上下文，支持对象池化
type EnhancedContext struct {
	// 核心上下文
	Request *app.RequestContext
	Context context.Context
	
	// 路由相关
	Params   Params // 路由参数
	FullPath string // 完整路径

	// 请求数据
	Keys map[string]interface{} // 上下文键值对
	
	// 响应数据  
	Writer ResponseWriter
	
	// 内部状态
	index    int8           // 中间件索引
	handlers []HandlerFunc  // 处理器链
	mu       sync.RWMutex   // 读写锁
	aborted  bool           // 是否中止
	
	// 池化标识
	pooled   bool           // 是否来自池
	acquired time.Time      // 获取时间
}

// HandlerFunc 处理函数类型
type HandlerFunc func(*EnhancedContext)

// ResponseWriter 响应写入器接口
type ResponseWriter interface {
	Header() map[string]string
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
	Status() int
	Size() int
	Written() bool
}

// responseWriter 响应写入器实现
type responseWriter struct {
	RequestContext *app.RequestContext
	status         int
	size           int
	written        bool
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
		ctx := &EnhancedContext{
			Keys:     make(map[string]interface{}),
			index:    -1,
			pooled:   true,
			acquired: time.Now(),
		}
		return ctx
	}
	
	return pool
}

// Get 从池中获取Context
func (pool *ContextPool) Get() *EnhancedContext {
	atomic.AddInt64(&pool.metrics.Gets, 1)
	
	ctx := pool.pool.Get().(*EnhancedContext)
	if ctx.Keys != nil {
		atomic.AddInt64(&pool.metrics.Reuses, 1)
	}
	
	ctx.acquired = time.Now()
	return ctx
}

// Put 将Context放回池中
func (pool *ContextPool) Put(ctx *EnhancedContext) {
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

// Reset 重置Context状态，准备复用
func (ctx *EnhancedContext) Reset() {
	ctx.Request = nil
	ctx.Context = nil
	ctx.Params = ctx.Params[:0]
	ctx.FullPath = ""
	
	// 清空Keys但保留底层数组
	for k := range ctx.Keys {
		delete(ctx.Keys, k)
	}
	
	ctx.Writer = nil
	ctx.index = -1
	ctx.handlers = ctx.handlers[:0]
	ctx.aborted = false
}

// NewContext 创建新的增强Context（使用池化）
func NewContext(c *app.RequestContext) *EnhancedContext {
	ctx := defaultPool.Get()
	ctx.Request = c
	ctx.Context = context.Background()
	ctx.Writer = &responseWriter{RequestContext: c}
	return ctx
}

// NewContextWithContext 使用指定context创建增强Context
func NewContextWithContext(c *app.RequestContext, parent context.Context) *EnhancedContext {
	ctx := defaultPool.Get()
	ctx.Request = c
	ctx.Context = parent
	ctx.Writer = &responseWriter{RequestContext: c}
	return ctx
}

// Release 释放Context到池中
func (ctx *EnhancedContext) Release() {
	if ctx.pooled {
		defaultPool.Put(ctx)
		atomic.AddInt32(&poolSize, -1)
	}
}

// ============= Context核心方法 =============

// Next 执行下一个中间件
func (ctx *EnhancedContext) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		if !ctx.aborted {
			ctx.handlers[ctx.index](ctx)
		}
		ctx.index++
	}
}

// Abort 中止执行
func (ctx *EnhancedContext) Abort() {
	ctx.aborted = true
}

// IsAborted 是否已中止
func (ctx *EnhancedContext) IsAborted() bool {
	return ctx.aborted
}

// Set 设置键值对
func (ctx *EnhancedContext) Set(key string, value interface{}) {
	ctx.mu.Lock()
	ctx.Keys[key] = value
	ctx.mu.Unlock()
}

// Get 获取值
func (ctx *EnhancedContext) Get(key string) (interface{}, bool) {
	ctx.mu.RLock()
	value, exists := ctx.Keys[key]
	ctx.mu.RUnlock()
	return value, exists
}

// MustGet 必须获取值
func (ctx *EnhancedContext) MustGet(key string) interface{} {
	if value, exists := ctx.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// Param 获取路由参数
func (ctx *EnhancedContext) Param(key string) string {
	return ctx.Params.ByName(key)
}

// Query 获取查询参数
func (ctx *EnhancedContext) Query(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.QueryArgs().Peek(key))
}

// PostForm 获取POST表单参数
func (ctx *EnhancedContext) PostForm(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.PostArgs().Peek(key))
}

// Header 获取请求头
func (ctx *EnhancedContext) Header(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.GetHeader(key))
}

// ============= 响应方法 =============

// JSON 返回JSON响应
func (ctx *EnhancedContext) JSON(code int, obj interface{}) {
	if ctx.Request != nil {
		ctx.Request.JSON(code, obj)
	}
}

// String 返回字符串响应
func (ctx *EnhancedContext) String(code int, format string, values ...interface{}) {
	if ctx.Request != nil {
		ctx.Request.String(code, format, values...)
	}
}

// HTML 返回HTML响应
func (ctx *EnhancedContext) HTML(code int, name string, obj interface{}) {
	if ctx.Request != nil {
		// 这里需要集成模板引擎
		ctx.Request.HTML(code, name, obj)
	}
}

// ============= ResponseWriter实现 =============

func (w *responseWriter) Header() map[string]string {
	headers := make(map[string]string)
	if w.RequestContext != nil {
		w.RequestContext.Response.Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})
	}
	return headers
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if w.RequestContext != nil {
		w.size += len(data)
		w.written = true
		return w.RequestContext.Write(data)
	}
	return 0, nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.RequestContext != nil && !w.written {
		w.status = statusCode
		w.RequestContext.SetStatusCode(statusCode)
	}
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.written
}

// ============= 池化统计和配置 =============

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

// ============= 批量操作优化 =============

// BatchContexts 批量Context处理器
type BatchContexts struct {
	contexts []*EnhancedContext
	size     int
}

// NewBatchContexts 创建批量处理器
func NewBatchContexts(size int) *BatchContexts {
	return &BatchContexts{
		contexts: make([]*EnhancedContext, size),
		size:     0,
	}
}

// Add 添加Context到批处理
func (batch *BatchContexts) Add(ctx *EnhancedContext) {
	if batch.size < len(batch.contexts) {
		batch.contexts[batch.size] = ctx
		batch.size++
	}
}

// Release 批量释放Context
func (batch *BatchContexts) Release() {
	for i := 0; i < batch.size; i++ {
		if ctx := batch.contexts[i]; ctx != nil {
			ctx.Release()
			batch.contexts[i] = nil
		}
	}
	batch.size = 0
}

// ForEach 遍历所有Context
func (batch *BatchContexts) ForEach(fn func(*EnhancedContext)) {
	for i := 0; i < batch.size; i++ {
		if ctx := batch.contexts[i]; ctx != nil {
			fn(ctx)
		}
	}
}

// SetHandlers 设置处理器链
func (ctx *EnhancedContext) SetHandlers(handlers []HandlerFunc) {
	ctx.handlers = handlers
	ctx.index = -1
}