package errors

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zsy619/yyhertz/framework/errors"
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	Handle(ctx *mvccontext.EnhancedContext, err error) error
	CanHandle(err error) bool
	Priority() int
}

// ErrorHandlerFunc 错误处理函数类型
type ErrorHandlerFunc func(ctx *mvccontext.EnhancedContext, err error) error

// ErrorContext 错误上下文
type ErrorContext struct {
	Original    error                          // 原始错误
	Request     *mvccontext.EnhancedContext    // 请求上下文
	Handled     bool                           // 是否已处理
	HandlerName string                         // 处理器名称
	Timestamp   time.Time                      // 错误时间
	StackTrace  string                         // 堆栈跟踪
	Metadata    map[string]interface{}         // 附加元数据
}

// ErrorDispatcher 错误分发器
type ErrorDispatcher struct {
	handlers    []ErrorHandler     // 注册的错误处理器
	fallback    ErrorHandlerFunc   // 兜底处理器
	mu          sync.RWMutex       // 读写锁
	stats       DispatcherStats    // 统计信息
	config      DispatcherConfig   // 配置
}

// DispatcherStats 分发器统计
type DispatcherStats struct {
	TotalErrors     int64 // 总错误数
	HandledErrors   int64 // 已处理错误数
	UnhandledErrors int64 // 未处理错误数
	HandlerStats    map[string]*HandlerStats // 各处理器统计
}

// HandlerStats 处理器统计
type HandlerStats struct {
	HandledCount int64         // 处理次数
	ErrorCount   int64         // 处理失败次数
	TotalTime    time.Duration // 总处理时间
	AverageTime  time.Duration // 平均处理时间
	LastHandled  time.Time     // 最后处理时间
}

// DispatcherConfig 分发器配置
type DispatcherConfig struct {
	EnablePanicRecovery    bool          // 启用panic恢复
	EnableStackTrace       bool          // 启用堆栈跟踪
	EnableStatistics       bool          // 启用统计
	MaxRetries             int           // 最大重试次数
	RetryInterval          time.Duration // 重试间隔
	EnableCircuitBreaker   bool          // 启用熔断器
	CircuitBreakerThreshold int          // 熔断器阈值
}

// NewErrorDispatcher 创建错误分发器
func NewErrorDispatcher() *ErrorDispatcher {
	dispatcher := &ErrorDispatcher{
		handlers: make([]ErrorHandler, 0),
		config:   DefaultDispatcherConfig(),
		stats: DispatcherStats{
			HandlerStats: make(map[string]*HandlerStats),
		},
	}
	
	// 设置默认兜底处理器
	dispatcher.SetFallbackHandler(DefaultFallbackHandler)
	
	return dispatcher
}

// DefaultDispatcherConfig 默认分发器配置
func DefaultDispatcherConfig() DispatcherConfig {
	return DispatcherConfig{
		EnablePanicRecovery:     true,
		EnableStackTrace:        true,
		EnableStatistics:        true,
		MaxRetries:             3,
		RetryInterval:          time.Second,
		EnableCircuitBreaker:   true,
		CircuitBreakerThreshold: 10,
	}
}

// RegisterHandler 注册错误处理器
func (d *ErrorDispatcher) RegisterHandler(handler ErrorHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// 插入到正确的位置（按优先级排序）
	inserted := false
	for i, h := range d.handlers {
		if handler.Priority() > h.Priority() {
			// 插入到当前位置
			d.handlers = append(d.handlers[:i+1], d.handlers[i:]...)
			d.handlers[i] = handler
			inserted = true
			break
		}
	}
	
	if !inserted {
		d.handlers = append(d.handlers, handler)
	}
	
	// 初始化统计信息
	if d.config.EnableStatistics {
		handlerName := fmt.Sprintf("%T", handler)
		d.stats.HandlerStats[handlerName] = &HandlerStats{}
	}
}

// RegisterHandlerFunc 注册错误处理函数
func (d *ErrorDispatcher) RegisterHandlerFunc(name string, priority int, canHandle func(error) bool, handleFunc ErrorHandlerFunc) {
	handler := &FuncErrorHandler{
		name:       name,
		priority:   priority,
		canHandle:  canHandle,
		handleFunc: handleFunc,
	}
	d.RegisterHandler(handler)
}

// SetFallbackHandler 设置兜底处理器
func (d *ErrorDispatcher) SetFallbackHandler(handler ErrorHandlerFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.fallback = handler
}

// Dispatch 分发错误处理
func (d *ErrorDispatcher) Dispatch(ctx *mvccontext.EnhancedContext, err error) error {
	if err == nil {
		return nil
	}
	
	atomic.AddInt64(&d.stats.TotalErrors, 1)
	
	// 创建错误上下文
	errorCtx := &ErrorContext{
		Original:  err,
		Request:   ctx,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	// 添加堆栈跟踪
	if d.config.EnableStackTrace {
		errorCtx.StackTrace = getStackTrace()
	}
	
	// 使用defer处理panic恢复
	if d.config.EnablePanicRecovery {
		defer func() {
			if r := recover(); r != nil {
				// 处理器内部发生panic，使用兜底处理器
				if d.fallback != nil {
					d.fallback(ctx, fmt.Errorf("error handler panic: %v, original error: %v", r, err))
				}
			}
		}()
	}
	
	// 尝试使用注册的处理器
	d.mu.RLock()
	handlers := make([]ErrorHandler, len(d.handlers))
	copy(handlers, d.handlers)
	d.mu.RUnlock()
	
	for _, handler := range handlers {
		if handler.CanHandle(err) {
			handlerName := fmt.Sprintf("%T", handler)
			start := time.Now()
			
			handleErr := d.handleWithRetry(handler, ctx, err)
			
			// 更新统计信息
			if d.config.EnableStatistics {
				d.updateHandlerStats(handlerName, time.Since(start), handleErr)
			}
			
			if handleErr == nil {
				errorCtx.Handled = true
				errorCtx.HandlerName = handlerName
				atomic.AddInt64(&d.stats.HandledErrors, 1)
				return nil
			}
		}
	}
	
	// 没有处理器能处理该错误，使用兜底处理器
	atomic.AddInt64(&d.stats.UnhandledErrors, 1)
	
	if d.fallback != nil {
		return d.fallback(ctx, err)
	}
	
	return err
}

// handleWithRetry 带重试的错误处理
func (d *ErrorDispatcher) handleWithRetry(handler ErrorHandler, ctx *mvccontext.EnhancedContext, err error) error {
	var lastErr error
	
	for i := 0; i <= d.config.MaxRetries; i++ {
		if i > 0 {
			// 重试前等待
			time.Sleep(d.config.RetryInterval)
		}
		
		lastErr = handler.Handle(ctx, err)
		if lastErr == nil {
			return nil
		}
	}
	
	return lastErr
}

// updateHandlerStats 更新处理器统计信息
func (d *ErrorDispatcher) updateHandlerStats(handlerName string, duration time.Duration, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	stats, exists := d.stats.HandlerStats[handlerName]
	if !exists {
		stats = &HandlerStats{}
		d.stats.HandlerStats[handlerName] = stats
	}
	
	if err == nil {
		atomic.AddInt64(&stats.HandledCount, 1)
	} else {
		atomic.AddInt64(&stats.ErrorCount, 1)
	}
	
	stats.TotalTime += duration
	totalCount := atomic.LoadInt64(&stats.HandledCount) + atomic.LoadInt64(&stats.ErrorCount)
	if totalCount > 0 {
		stats.AverageTime = stats.TotalTime / time.Duration(totalCount)
	}
	stats.LastHandled = time.Now()
}

// GetStatistics 获取统计信息
func (d *ErrorDispatcher) GetStatistics() DispatcherStats {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	// 深拷贝统计信息
	result := DispatcherStats{
		TotalErrors:     atomic.LoadInt64(&d.stats.TotalErrors),
		HandledErrors:   atomic.LoadInt64(&d.stats.HandledErrors),
		UnhandledErrors: atomic.LoadInt64(&d.stats.UnhandledErrors),
		HandlerStats:    make(map[string]*HandlerStats),
	}
	
	for name, stats := range d.stats.HandlerStats {
		result.HandlerStats[name] = &HandlerStats{
			HandledCount: atomic.LoadInt64(&stats.HandledCount),
			ErrorCount:   atomic.LoadInt64(&stats.ErrorCount),
			TotalTime:    stats.TotalTime,
			AverageTime:  stats.AverageTime,
			LastHandled:  stats.LastHandled,
		}
	}
	
	return result
}

// FuncErrorHandler 函数式错误处理器
type FuncErrorHandler struct {
	name       string
	priority   int
	canHandle  func(error) bool
	handleFunc ErrorHandlerFunc
}

func (h *FuncErrorHandler) Handle(ctx *mvccontext.EnhancedContext, err error) error {
	return h.handleFunc(ctx, err)
}

func (h *FuncErrorHandler) CanHandle(err error) bool {
	return h.canHandle(err)
}

func (h *FuncErrorHandler) Priority() int {
	return h.priority
}

// 内置错误处理器

// BusinessErrorHandler 业务错误处理器
type BusinessErrorHandler struct{}

func (h *BusinessErrorHandler) Handle(ctx *mvccontext.EnhancedContext, err error) error {
	if errNo, ok := err.(*errors.ErrNo); ok {
		ctx.JSON(400, map[string]interface{}{
			"code":    errNo.ErrCode,
			"message": errNo.ErrMsg,
			"success": false,
		})
		return nil
	}
	return err
}

func (h *BusinessErrorHandler) CanHandle(err error) bool {
	_, ok := err.(*errors.ErrNo)
	return ok
}

func (h *BusinessErrorHandler) Priority() int {
	return 100 // 高优先级
}

// SystemErrorHandler 系统错误处理器
type SystemErrorHandler struct{}

func (h *SystemErrorHandler) Handle(ctx *mvccontext.EnhancedContext, err error) error {
	ctx.JSON(500, map[string]interface{}{
		"code":    500,
		"message": "Internal Server Error",
		"success": false,
	})
	return nil
}

func (h *SystemErrorHandler) CanHandle(err error) bool {
	// 处理非业务错误
	_, isBusiness := err.(*errors.ErrNo)
	return !isBusiness
}

func (h *SystemErrorHandler) Priority() int {
	return 50 // 中等优先级
}

// DefaultFallbackHandler 默认兜底处理器
func DefaultFallbackHandler(ctx *mvccontext.EnhancedContext, err error) error {
	ctx.JSON(500, map[string]interface{}{
		"code":    500,
		"message": "Unknown Error",
		"error":   err.Error(),
		"success": false,
	})
	return nil
}

// 全局错误分发器
var globalDispatcher = NewErrorDispatcher()

func init() {
	// 注册内置错误处理器
	globalDispatcher.RegisterHandler(&BusinessErrorHandler{})
	globalDispatcher.RegisterHandler(&SystemErrorHandler{})
}

// GetGlobalDispatcher 获取全局错误分发器
func GetGlobalDispatcher() *ErrorDispatcher {
	return globalDispatcher
}

// DispatchError 分发错误（全局方法）
func DispatchError(ctx *mvccontext.EnhancedContext, err error) error {
	return globalDispatcher.Dispatch(ctx, err)
}

// RegisterGlobalHandler 注册全局错误处理器
func RegisterGlobalHandler(handler ErrorHandler) {
	globalDispatcher.RegisterHandler(handler)
}

// RegisterGlobalHandlerFunc 注册全局错误处理函数
func RegisterGlobalHandlerFunc(name string, priority int, canHandle func(error) bool, handleFunc ErrorHandlerFunc) {
	globalDispatcher.RegisterHandlerFunc(name, priority, canHandle, handleFunc)
}

// getStackTrace 获取堆栈跟踪信息
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// PrintErrorHandlerInfo 打印错误处理器信息
func PrintErrorHandlerInfo() {
	stats := globalDispatcher.GetStatistics()
	
	fmt.Println("=== Error Handler Statistics ===")
	fmt.Printf("Total Errors: %d\n", stats.TotalErrors)
	fmt.Printf("Handled Errors: %d\n", stats.HandledErrors)
	fmt.Printf("Unhandled Errors: %d\n", stats.UnhandledErrors)
	
	if stats.TotalErrors > 0 {
		handledRate := float64(stats.HandledErrors) / float64(stats.TotalErrors) * 100
		fmt.Printf("Handled Rate: %.2f%%\n", handledRate)
	}
	
	fmt.Println("\nHandler Statistics:")
	for name, handlerStats := range stats.HandlerStats {
		total := handlerStats.HandledCount + handlerStats.ErrorCount
		if total > 0 {
			successRate := float64(handlerStats.HandledCount) / float64(total) * 100
			fmt.Printf("  %s: %d total (%d success, %d failed), %.2f%% success rate, avg: %v\n",
				name, total, handlerStats.HandledCount, handlerStats.ErrorCount, 
				successRate, handlerStats.AverageTime)
		}
	}
}