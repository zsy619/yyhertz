package context

import (
	stdcontext "context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// ============= 基础类型定义 =============

// Param 路由参数结构
type Param struct {
	Key   string
	Value string
}

// Params 路由参数列表
type Params []Param

// ByName 通过名称获取参数值
func (ps Params) ByName(name string) string {
	for _, param := range ps {
		if param.Key == name {
			return param.Value
		}
	}
	return ""
}

// Get 获取参数值（别名方法）
func (ps Params) Get(name string) (string, bool) {
	for _, param := range ps {
		if param.Key == name {
			return param.Value, true
		}
	}
	return "", false
}

// 注意：ResponseWriter接口在response_writer.go中已定义

// HandlerFunc 处理函数类型
type HandlerFunc func(*Context)

// Context 增强的上下文，支持对象池化和高性能并发访问
//
// 设计原则：
// 1. 高性能：使用sync.Map和原子操作优化并发性能
// 2. 向后兼容：保持与旧版本API的完全兼容
// 3. 模块化：功能按模块拆分，便于维护
// 4. 类型安全：减少运行时错误
//
// 性能特性：
// - 支持对象池化，减少GC压力
// - 原子操作替代粗粒度锁
// - 高效的并发数据存储
type Context struct {
	// ============= 核心字段（优化后） =============

	// 核心上下文 - 私有字段，通过方法访问以确保安全性
	request *app.RequestContext
	writer  ResponseWriter
	Context stdcontext.Context

	// 路由相关
	params   Params // 路由参数
	fullPath string // 完整路径

	// 高性能并发数据存储 - 使用sync.Map替代map+mutex
	keys sync.Map // 上下文键值对，支持高并发访问

	// 兼容性字段 - 为了向后兼容传统MVC风格API
	Input  *InputData
	Output *OutputData

	// ============= 中间件状态（原子操作优化） =============

	index    int8          // 中间件索引
	handlers []HandlerFunc // 处理器链
	aborted  int32         // 是否中止 - 使用原子操作优化

	// ============= 错误处理（细粒度锁优化） =============

	errors []error    // 错误列表
	errMu  sync.Mutex // 专用错误锁，细粒度控制

	// ============= 池化相关 =============

	pooled   bool      // 是否来自池
	acquired time.Time // 获取时间
}

// ============= 构造函数（优化版本） =============

// NewContext 创建新的增强Context（使用池化）
func NewContext(c *app.RequestContext) *Context {
	return NewContextWithContext(c, stdcontext.Background())
}

// NewContextWithContext 使用指定context创建增强Context
func NewContextWithContext(c *app.RequestContext, parent stdcontext.Context) *Context {
	ctx := defaultPool.Get()
	ctx.request = c
	ctx.Context = parent
	ctx.writer = &responseWriter{RequestContext: c}

	// 初始化兼容性字段
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}

	return ctx
}

// ============= 重置和池化方法（优化版本） =============

// Reset 重置Context状态，准备复用
func (ctx *Context) Reset() {
	// 重置核心字段
	ctx.request = nil
	ctx.Context = nil
	ctx.params = ctx.params[:0]
	ctx.fullPath = ""

	// 清空sync.Map中的数据
	ctx.keys.Range(func(key, value any) bool {
		ctx.keys.Delete(key)
		return true
	})

	// 重置响应相关
	ctx.writer = nil

	// 重置中间件状态
	ctx.index = -1
	ctx.handlers = ctx.handlers[:0]
	atomic.StoreInt32(&ctx.aborted, 0) // 原子重置

	// 重置错误列表
	ctx.errMu.Lock()
	ctx.errors = ctx.errors[:0]
	ctx.errMu.Unlock()
}

// Release 释放Context到池中
func (ctx *Context) Release() {
	if ctx.pooled {
		defaultPool.Put(ctx)
		atomic.AddInt32(&poolSize, -1)
	}
}

// ============= 核心数据访问方法（优化版本） =============

// Set 设置键值对 - 使用sync.Map优化并发性能
func (ctx *Context) Set(key string, value any) {
	ctx.keys.Store(key, value)
}

// Get 获取值 - 使用sync.Map优化并发性能
func (ctx *Context) Get(key string) (any, bool) {
	return ctx.keys.Load(key)
}

// MustGet 必须获取值
func (ctx *Context) MustGet(key string) any {
	if value, exists := ctx.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// ============= 增强Keys操作（批量操作） =============

// SetMultiple 批量设置键值对
func (ctx *Context) SetMultiple(pairs map[string]any) {
	for key, value := range pairs {
		ctx.keys.Store(key, value)
	}
}

// GetMultiple 批量获取值
func (ctx *Context) GetMultiple(keys []string) map[string]any {
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, exists := ctx.keys.Load(key); exists {
			result[key] = value
		}
	}
	return result
}

// DeleteMultiple 批量删除键
func (ctx *Context) DeleteMultiple(keys []string) {
	for _, key := range keys {
		ctx.keys.Delete(key)
	}
}

// ============= 增强Keys操作（类型安全） =============

// SetTypedString 设置字符串类型值（类型安全）
func (ctx *Context) SetTypedString(key string, value string) {
	ctx.keys.Store(key, value)
}

// GetTypedString 获取字符串类型值（类型安全）
func (ctx *Context) GetTypedString(key string) (string, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// SetTypedInt 设置整数类型值（类型安全）
func (ctx *Context) SetTypedInt(key string, value int) {
	ctx.keys.Store(key, value)
}

// GetTypedInt 获取整数类型值（类型安全）
func (ctx *Context) GetTypedInt(key string) (int, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return 0, false
	}
	intVal, ok := value.(int)
	return intVal, ok
}

// SetTypedBool 设置布尔类型值（类型安全）
func (ctx *Context) SetTypedBool(key string, value bool) {
	ctx.keys.Store(key, value)
}

// GetTypedBool 获取布尔类型值（类型安全）
func (ctx *Context) GetTypedBool(key string) (bool, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return false, false
	}
	boolVal, ok := value.(bool)
	return boolVal, ok
}

// SetFloat64 设置浮点数类型值（类型安全）
func (ctx *Context) SetFloat64(key string, value float64) {
	ctx.keys.Store(key, value)
}

// GetFloat64 获取浮点数类型值（类型安全）
func (ctx *Context) GetFloat64(key string) (float64, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return 0, false
	}
	floatVal, ok := value.(float64)
	return floatVal, ok
}

// ============= 增强Keys操作（条件操作） =============

// SetIfNotExists 仅当键不存在时设置值
func (ctx *Context) SetIfNotExists(key string, value any) bool {
	_, loaded := ctx.keys.LoadOrStore(key, value)
	return !loaded // 返回true表示成功设置，false表示键已存在
}

// GetOrSet 获取值，如果不存在则设置默认值
func (ctx *Context) GetOrSet(key string, defaultValue any) any {
	actual, _ := ctx.keys.LoadOrStore(key, defaultValue)
	return actual
}

// CompareAndSwap 比较并交换值（原子操作）
func (ctx *Context) CompareAndSwap(key string, oldValue, newValue any) bool {
	return ctx.keys.CompareAndSwap(key, oldValue, newValue)
}

// Delete 删除键
func (ctx *Context) Delete(key string) {
	ctx.keys.Delete(key)
}

// Exists 检查键是否存在
func (ctx *Context) Exists(key string) bool {
	_, exists := ctx.keys.Load(key)
	return exists
}

// ============= 增强Keys操作（遍历和过滤） =============

// ForEach 遍历所有键值对
func (ctx *Context) ForEach(fn func(key string, value any) bool) {
	ctx.keys.Range(func(k, v any) bool {
		key, ok := k.(string)
		if !ok {
			return true // 跳过非字符串键
		}
		return fn(key, v)
	})
}

// Filter 过滤键值对，返回满足条件的键列表
func (ctx *Context) Filter(predicate func(key string, value any) bool) []string {
	var result []string
	ctx.keys.Range(func(k, v any) bool {
		key, ok := k.(string)
		if !ok {
			return true
		}
		if predicate(key, v) {
			result = append(result, key)
		}
		return true
	})
	return result
}

// MapKeys 对所有键值对执行映射操作
func (ctx *Context) MapKeys(mapper func(key string, value any) any) map[string]any {
	result := make(map[string]any)
	ctx.keys.Range(func(k, v any) bool {
		key, ok := k.(string)
		if !ok {
			return true
		}
		result[key] = mapper(key, v)
		return true
	})
	return result
}

// Clear 清空所有键值对
func (ctx *Context) Clear() {
	ctx.keys.Range(func(key, value any) bool {
		ctx.keys.Delete(key)
		return true
	})
}

// ============= 错误处理方法（细粒度锁优化） =============

// AddError 添加错误 - 使用专用锁优化
func (ctx *Context) AddError(err error) {
	if err != nil {
		ctx.errMu.Lock()
		ctx.errors = append(ctx.errors, err)
		ctx.errMu.Unlock()
	}
}

// GetErrors 获取所有错误
func (ctx *Context) GetErrors() []error {
	ctx.errMu.Lock()
	errors := make([]error, len(ctx.errors))
	copy(errors, ctx.errors)
	ctx.errMu.Unlock()
	return errors
}

// HasErrors 是否有错误
func (ctx *Context) HasErrors() bool {
	ctx.errMu.Lock()
	hasErr := len(ctx.errors) > 0
	ctx.errMu.Unlock()
	return hasErr
}

// ClearErrors 清除所有错误
func (ctx *Context) ClearErrors() {
	ctx.errMu.Lock()
	ctx.errors = ctx.errors[:0]
	ctx.errMu.Unlock()
}

// LastError 获取最后一个错误
func (ctx *Context) LastError() error {
	ctx.errMu.Lock()
	defer ctx.errMu.Unlock()

	if len(ctx.errors) == 0 {
		return nil
	}
	return ctx.errors[len(ctx.errors)-1]
}

// ============= 兼容性访问器方法 =============
// 这些方法提供向后兼容性，允许现有代码无缝迁移

// Request 获取Request对象
func (ctx *Context) Request() *app.RequestContext {
	return ctx.request
}

// RequestContext 获取RequestContext (兼容性方法)  
func (ctx *Context) RequestContext() *app.RequestContext {
	return ctx.request
}

// Writer 获取Writer对象
func (ctx *Context) Writer() ResponseWriter {
	return ctx.writer
}

// ResponseWriter 获取ResponseWriter (兼容性方法)
func (ctx *Context) ResponseWriter() ResponseWriter {
	return ctx.writer
}

// Params 获取路由参数
func (ctx *Context) Params() Params {
	return ctx.params
}

// SetParams 设置路由参数
func (ctx *Context) SetParams(params Params) {
	ctx.params = params
}

// ParamKeys 获取所有路由参数的键名
func (ctx *Context) ParamKeys() []string {
	if len(ctx.params) == 0 {
		return nil
	}
	
	keys := make([]string, len(ctx.params))
	for i, param := range ctx.params {
		keys[i] = param.Key
	}
	return keys
}

// ParamMap 将路由参数转换为map[string]string
func (ctx *Context) ParamMap() map[string]string {
	if len(ctx.params) == 0 {
		return make(map[string]string)
	}
	
	paramMap := make(map[string]string, len(ctx.params))
	for _, param := range ctx.params {
		paramMap[param.Key] = param.Value
	}
	return paramMap
}

// ParamValues 获取所有路由参数的值
func (ctx *Context) ParamValues() []string {
	if len(ctx.params) == 0 {
		return nil
	}
	
	values := make([]string, len(ctx.params))
	for i, param := range ctx.params {
		values[i] = param.Value
	}
	return values
}

// FullPath 获取完整路径
func (ctx *Context) FullPath() string {
	return ctx.fullPath
}

// SetFullPath 设置完整路径
func (ctx *Context) SetFullPath(path string) {
	ctx.fullPath = path
}

// ============= 性能统计方法 =============

// Acquired 获取Context获取时间
func (ctx *Context) Acquired() time.Time {
	return ctx.acquired
}

// IsPooled 是否来自对象池
func (ctx *Context) IsPooled() bool {
	return ctx.pooled
}

// ============= 调试和监控方法 =============

// KeysCount 获取存储的键值对数量
func (ctx *Context) KeysCount() int {
	count := 0
	ctx.keys.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// ListKeys 列出所有键
func (ctx *Context) ListKeys() []string {
	var keys []string
	ctx.keys.Range(func(key, value any) bool {
		if k, ok := key.(string); ok {
			keys = append(keys, k)
		}
		return true
	})
	return keys
}

// ============= 类型断言辅助方法 =============

// GetString 获取字符串类型值
func (ctx *Context) GetStringValue(key string) (string, bool) {
	value, exists := ctx.Get(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// GetInt 获取整数类型值
func (ctx *Context) GetIntValue(key string) (int, bool) {
	value, exists := ctx.Get(key)
	if !exists {
		return 0, false
	}
	intVal, ok := value.(int)
	return intVal, ok
}

// GetBool 获取布尔类型值
func (ctx *Context) GetBoolValue(key string) (bool, bool) {
	value, exists := ctx.Get(key)
	if !exists {
		return false, false
	}
	boolVal, ok := value.(bool)
	return boolVal, ok
}

// ============= Writer Copy功能 =============

// Copy 从io.Reader复制数据到Context的Writer，类似io.Copy功能
// 这个方法提供了类似io.Copy(ctx.Writer(), reader)的功能
// 返回写入的字节数和可能的错误
func (ctx *Context) Copy(reader io.Reader) (int64, error) {
	if !ctx.ensureRequest() {
		return 0, ErrRequestNotFound
	}

	if reader == nil {
		return 0, &ContextError{
			Code:    "NIL_READER",
			Message: "Reader cannot be nil",
		}
	}

	// 使用缓冲区进行高效复制
	buf := make([]byte, 32*1024) // 32KB缓冲区，平衡性能和内存使用
	var totalWritten int64

	for {
		// 从reader读取数据
		bytesRead, readErr := reader.Read(buf)
		if bytesRead > 0 {
			// 写入到response
			if _, writeErr := ctx.writer.Write(buf[:bytesRead]); writeErr != nil {
				return totalWritten, &ContextError{
					Code:    "WRITE_ERROR",
					Message: "Failed to write to response: " + writeErr.Error(),
					Cause:   writeErr,
				}
			}
			totalWritten += int64(bytesRead)
		}

		// 检查读取错误
		if readErr != nil {
			if readErr == io.EOF {
				// 正常结束
				break
			}
			return totalWritten, &ContextError{
				Code:    "READ_ERROR", 
				Message: "Failed to read from source: " + readErr.Error(),
				Cause:   readErr,
			}
		}
	}

	return totalWritten, nil
}

// CopyBuffer 使用指定缓冲区从io.Reader复制数据到Context的Writer
// 允许用户自定义缓冲区大小以优化性能
func (ctx *Context) CopyBuffer(reader io.Reader, buf []byte) (int64, error) {
	if !ctx.ensureRequest() {
		return 0, ErrRequestNotFound
	}

	if reader == nil {
		return 0, &ContextError{
			Code:    "NIL_READER",
			Message: "Reader cannot be nil",
		}
	}

	if buf == nil || len(buf) == 0 {
		// 如果没有提供缓冲区，使用默认的Copy方法
		return ctx.Copy(reader)
	}

	var totalWritten int64

	for {
		// 从reader读取数据到指定缓冲区
		bytesRead, readErr := reader.Read(buf)
		if bytesRead > 0 {
			// 写入到response
			if _, writeErr := ctx.writer.Write(buf[:bytesRead]); writeErr != nil {
				return totalWritten, &ContextError{
					Code:    "WRITE_ERROR",
					Message: "Failed to write to response: " + writeErr.Error(),
					Cause:   writeErr,
				}
			}
			totalWritten += int64(bytesRead)
		}

		// 检查读取错误
		if readErr != nil {
			if readErr == io.EOF {
				// 正常结束
				break
			}
			return totalWritten, &ContextError{
				Code:    "READ_ERROR",
				Message: "Failed to read from source: " + readErr.Error(),
				Cause:   readErr,
			}
		}
	}

	return totalWritten, nil
}

// CopyWithContentType 复制数据并设置Content-Type
// 这是一个便捷方法，将Copy和设置Content-Type结合
func (ctx *Context) CopyWithContentType(reader io.Reader, contentType string) (int64, error) {
	// 设置Content-Type
	ctx.SetContentType(contentType)
	
	// 执行复制
	return ctx.Copy(reader)
}

// StreamCopy 流式复制，支持实时写入（适合大文件或流数据）
// 每次写入后会刷新缓冲区，适合需要实时传输的场景
func (ctx *Context) StreamCopy(reader io.Reader) (int64, error) {
	if !ctx.ensureRequest() {
		return 0, ErrRequestNotFound
	}

	if reader == nil {
		return 0, &ContextError{
			Code:    "NIL_READER",
			Message: "Reader cannot be nil",
		}
	}

	// 对于流式传输，使用较小的缓冲区以减少延迟
	buf := make([]byte, 8*1024) // 8KB缓冲区
	var totalWritten int64

	for {
		// 从reader读取数据
		bytesRead, readErr := reader.Read(buf)
		if bytesRead > 0 {
			// 写入到response
			if _, writeErr := ctx.writer.Write(buf[:bytesRead]); writeErr != nil {
				return totalWritten, &ContextError{
					Code:    "WRITE_ERROR",
					Message: "Failed to write to response: " + writeErr.Error(),
					Cause:   writeErr,
				}
			}
			totalWritten += int64(bytesRead)
			
			// 流式传输时，每次写入后尝试刷新（如果writer支持）
			if flusher, ok := ctx.writer.(interface{ Flush() error }); ok {
				flusher.Flush()
			}
		}

		// 检查读取错误
		if readErr != nil {
			if readErr == io.EOF {
				// 正常结束
				break
			}
			return totalWritten, &ContextError{
				Code:    "READ_ERROR",
				Message: "Failed to read from source: " + readErr.Error(), 
				Cause:   readErr,
			}
		}
	}

	return totalWritten, nil
}
