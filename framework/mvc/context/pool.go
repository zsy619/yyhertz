package context

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
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
	RequestContext *app.RequestContext // 兼容性字段别名
	Context context.Context
	
	// 路由相关
	Params   Params // 路由参数
	FullPath string // 完整路径

	// 请求数据
	Keys map[string]interface{} // 上下文键值对
	
	// 响应数据  
	Writer ResponseWriter
	ResponseWriter ResponseWriter // 兼容性字段别名
	
	// 兼容性字段 - 为了向后兼容Beego风格API
	Input  *InputData
	Output *OutputData
	
	// 内部状态
	index    int8           // 中间件索引
	handlers []HandlerFunc  // 处理器链
	mu       sync.RWMutex   // 读写锁
	aborted  bool           // 是否中止
	errors   []error        // 错误列表
	
	// 池化标识
	pooled   bool           // 是否来自池
	acquired time.Time      // 获取时间
}

// HandlerFunc 处理函数类型
type HandlerFunc func(*EnhancedContext)

// InputData Beego风格输入数据结构
type InputData struct {
	ctx *EnhancedContext
}

// OutputData Beego风格输出数据结构
type OutputData struct {
	ctx *EnhancedContext
}

// Cookie 设置Cookie (Output兼容性方法)
func (o *OutputData) Cookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if o.ctx.Request != nil {
		o.ctx.Request.SetCookie(name, value, maxAge, path, domain, protocol.CookieSameSiteDefaultMode, secure, httpOnly)
	}
}

// Header 设置响应头 (Output兼容性方法)
func (o *OutputData) Header(key, value string) {
	if o.ctx.Request != nil {
		o.ctx.Request.Response.Header.Set(key, value)
	}
}

// Status 设置状态码 (Output兼容性方法)
func (o *OutputData) Status(code int) {
	if o.ctx.Request != nil {
		o.ctx.Request.Response.SetStatusCode(code)
	}
}

// Body 设置响应体 (Output兼容性方法)
func (o *OutputData) Body(content []byte) error {
	if o.ctx.Request != nil {
		o.ctx.Request.Response.SetBody(content)
	}
	return nil
}

// JSON 设置JSON响应 (Output兼容性方法)
func (o *OutputData) JSON(data interface{}, hasIndent bool, coding ...bool) error {
	if o.ctx.Request != nil {
		o.ctx.Request.JSON(200, data)
	}
	return nil
}

// SetStatus 设置状态码 (Output兼容性方法，别名)
func (o *OutputData) SetStatus(code int) {
	o.Status(code)
}

// Param 获取路由参数 (Input兼容性方法)
func (i *InputData) Param(key string) string {
	return i.ctx.Params.ByName(key)
}

// Query 获取查询参数 (Input兼容性方法)
func (i *InputData) Query(key string) string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.QueryArgs().Peek(key))
	}
	return ""
}

// Header 获取请求头 (Input兼容性方法)
func (i *InputData) Header(key string) string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.GetHeader(key))
	}
	return ""
}

// Cookie 获取Cookie (Input兼容性方法)
func (i *InputData) Cookie(key string) string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.Cookie(key))
	}
	return ""
}

// Data 设置上下文数据 (Input兼容性方法)
func (i *InputData) Data(key string, val interface{}) {
	if i.ctx != nil {
		i.ctx.Keys[key] = val
	}
}

// RequestBody 获取请求体数据 (Input兼容性方法)
func (i *InputData) RequestBody() []byte {
	if i.ctx.Request != nil {
		body, _ := i.ctx.Request.Body()
		return body
	}
	return nil
}

// IP 获取客户端IP (Input兼容性方法)
func (i *InputData) IP() string {
	if i.ctx.Request != nil {
		return i.ctx.Request.ClientIP()
	}
	return ""
}

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
		// 初始化兼容性字段（这些字段会在NewContext中重新赋值，这里先设置为避免nil指针）
		ctx.Input = &InputData{ctx: ctx}
		ctx.Output = &OutputData{ctx: ctx}
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
	ctx.RequestContext = nil // 同时重置兼容性别名
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
	ctx.errors = ctx.errors[:0]
}

// NewContext 创建新的增强Context（使用池化）
func NewContext(c *app.RequestContext) *EnhancedContext {
	ctx := defaultPool.Get()
	ctx.Request = c
	ctx.RequestContext = c // 兼容性别名指向同一对象
	ctx.Context = context.Background()
	ctx.Writer = &responseWriter{RequestContext: c}
	ctx.ResponseWriter = ctx.Writer // 兼容性别名指向同一对象
	
	// 初始化Beego风格兼容性字段
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}
	
	return ctx
}

// NewContextWithContext 使用指定context创建增强Context
func NewContextWithContext(c *app.RequestContext, parent context.Context) *EnhancedContext {
	ctx := defaultPool.Get()
	ctx.Request = c
	ctx.RequestContext = c // 兼容性别名指向同一对象
	ctx.Context = parent
	ctx.Writer = &responseWriter{RequestContext: c}
	ctx.ResponseWriter = ctx.Writer // 兼容性别名指向同一对象
	
	// 初始化Beego风格兼容性字段
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}
	
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

// GetHeader 获取请求头 (兼容性别名)
func (ctx *EnhancedContext) GetHeader(key string) string {
	return ctx.Header(key)
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

// ============= 错误处理方法 =============

// AddError 添加错误
func (ctx *EnhancedContext) AddError(err error) {
	if err != nil {
		ctx.mu.Lock()
		ctx.errors = append(ctx.errors, err)
		ctx.mu.Unlock()
	}
}

// GetErrors 获取所有错误
func (ctx *EnhancedContext) GetErrors() []error {
	ctx.mu.RLock()
	errors := make([]error, len(ctx.errors))
	copy(errors, ctx.errors)
	ctx.mu.RUnlock()
	return errors
}

// HasErrors 是否有错误
func (ctx *EnhancedContext) HasErrors() bool {
	ctx.mu.RLock()
	hasErr := len(ctx.errors) > 0
	ctx.mu.RUnlock()
	return hasErr
}

// ClearErrors 清除所有错误
func (ctx *EnhancedContext) ClearErrors() {
	ctx.mu.Lock()
	ctx.errors = ctx.errors[:0]
	ctx.mu.Unlock()
}

// LastError 获取最后一个错误
func (ctx *EnhancedContext) LastError() error {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	
	if len(ctx.errors) == 0 {
		return nil
	}
	return ctx.errors[len(ctx.errors)-1]
}

// ============= 兼容性API - 来自原@framework/context =============
// 为了与原framework/context保持兼容性而提供的类型别名和适配器

// Context 兼容性别名 - 映射到EnhancedContext
type Context = EnhancedContext

// ContextAdapter 兼容性适配器
type ContextAdapter struct{}

// NewContextAdapter 创建新的上下文适配器 (兼容性API)
func NewContextAdapter() *ContextAdapter {
	return &ContextAdapter{}
}

// HertzToContext 将Hertz的RequestContext转换为Context (兼容性API)
func (a *ContextAdapter) HertzToContext(ctx *app.RequestContext) *EnhancedContext {
	return NewContext(ctx)
}

// ContextToHertz 从Context获取Hertz的RequestContext (兼容性API)
func (a *ContextAdapter) ContextToHertz(ctx *EnhancedContext) *app.RequestContext {
	if ctx == nil {
		return nil
	}
	return ctx.Request
}

// ============= 兼容性方法 - 为Context别名提供原框架期望的方法 =============

// AbortWithStatus 终止并设置状态码 (兼容性方法)
func (ctx *EnhancedContext) AbortWithStatus(code int) {
	ctx.JSON(code, map[string]string{"error": "Request aborted"})
	ctx.Abort()
}

// Write 写入响应数据 (兼容性方法)
func (ctx *EnhancedContext) Write(data []byte) (int, error) {
	return ctx.Writer.Write(data)
}

// SetHeader 设置响应头 (兼容性方法)
func (ctx *EnhancedContext) SetHeader(key, value string) {
	if ctx.Request != nil {
		ctx.Request.Response.Header.Set(key, value)
	}
}