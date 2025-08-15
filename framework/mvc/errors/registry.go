package errors

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 错误处理注册器 =============

// ErrorRegistry 错误处理注册器，管理各种错误处理器
// 整合了原有的错误处理注册机制和新的统一ErrorDispatcher
type ErrorRegistry struct {
	// handlers 按状态码分组的错误处理器列表
	handlers map[int][]ErrorHandler

	// fallbacks 错误回退映射，例如 503 -> [500, 400]
	fallbacks map[int][]int

	// dispatcher 统一错误分发器
	dispatcher *ErrorDispatcher

	// config 错误处理配置
	config *RegistryConfig

	// metrics 错误处理指标
	metrics *RegistryMetrics

	// mutex 读写锁，保证并发安全
	mutex sync.RWMutex

	// defaultHandlers 默认错误处理器映射
	defaultHandlers map[int]ErrorHandler

	// globalHandlers 全局错误处理器列表（不分状态码）
	globalHandlers []ErrorHandler
}

// RegistryConfig 注册器配置
type RegistryConfig struct {
	// EnableFallback 启用错误回退机制
	EnableFallback bool `json:"enable_fallback" yaml:"enable_fallback"`

	// EnableMetrics 启用错误指标收集
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`

	// EnableLogging 启用错误日志记录
	EnableLogging bool `json:"enable_logging" yaml:"enable_logging"`

	// FallbackChainLength 最大回退链长度
	FallbackChainLength int `json:"fallback_chain_length" yaml:"fallback_chain_length"`

	// HandlerTimeout 单个处理器超时时间
	HandlerTimeout time.Duration `json:"handler_timeout" yaml:"handler_timeout"`

	// EnablePanicRecovery 启用panic恢复
	EnablePanicRecovery bool `json:"enable_panic_recovery" yaml:"enable_panic_recovery"`

	// ShowDetailedError 是否显示详细错误信息
	ShowDetailedError bool `json:"show_detailed_error" yaml:"show_detailed_error"`

	// ContentNegotiation 启用内容协商
	ContentNegotiation bool `json:"content_negotiation" yaml:"content_negotiation"`
}

// RegistryMetrics 注册器指标
type RegistryMetrics struct {
	TotalRequests       int64                       `json:"total_requests"`       // 总请求数
	HandledErrors       int64                       `json:"handled_errors"`       // 已处理错误数
	UnhandledErrors     int64                       `json:"unhandled_errors"`     // 未处理错误数
	FallbackUsed        int64                       `json:"fallback_used"`        // 使用回退的次数
	StatusCodeCounts    map[int]int64               `json:"status_code_counts"`   // 各状态码统计
	HandlerTypeCounts   map[string]int64            `json:"handler_type_counts"`  // 处理器类型统计
	AverageHandleTime   time.Duration               `json:"average_handle_time"`  // 平均处理时间
	HandlerPerformance  map[string]*HandlerMetrics  `json:"handler_performance"`  // 处理器性能指标
	LastUpdated         time.Time                   `json:"last_updated"`         // 最后更新时间
	mu                  sync.RWMutex                `json:"-"`                    // 保护指标的锁
}

// HandlerMetrics 处理器指标
type HandlerMetrics struct {
	CallCount       int64         `json:"call_count"`       // 调用次数
	SuccessCount    int64         `json:"success_count"`    // 成功次数
	FailureCount    int64         `json:"failure_count"`    // 失败次数
	AverageTime     time.Duration `json:"average_time"`     // 平均耗时
	TotalTime       time.Duration `json:"total_time"`       // 总耗时
	LastCalled      time.Time     `json:"last_called"`      // 最后调用时间
	MaxTime         time.Duration `json:"max_time"`         // 最大耗时
	MinTime         time.Duration `json:"min_time"`         // 最小耗时
}

// DefaultRegistryConfig 默认注册器配置
func DefaultRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		EnableFallback:      true,
		EnableMetrics:       true,
		EnableLogging:       true,
		FallbackChainLength: 3,
		HandlerTimeout:      30 * time.Second,
		EnablePanicRecovery: true,
		ShowDetailedError:   true,
		ContentNegotiation:  true,
	}
}

// NewErrorRegistry 创建新的错误处理注册器
func NewErrorRegistry(config *RegistryConfig) *ErrorRegistry {
	if config == nil {
		config = DefaultRegistryConfig()
	}

	registry := &ErrorRegistry{
		handlers:        make(map[int][]ErrorHandler),
		fallbacks:       make(map[int][]int),
		dispatcher:      NewErrorDispatcher(),
		config:          config,
		metrics:         NewRegistryMetrics(),
		defaultHandlers: make(map[int]ErrorHandler),
		globalHandlers:  make([]ErrorHandler, 0),
	}

	// 注册默认错误处理器
	registry.registerDefaultHandlers()

	// 设置默认回退规则
	registry.setupDefaultFallbacks()

	return registry
}

// NewRegistryMetrics 创建新的注册器指标
func NewRegistryMetrics() *RegistryMetrics {
	return &RegistryMetrics{
		StatusCodeCounts:   make(map[int]int64),
		HandlerTypeCounts:  make(map[string]int64),
		HandlerPerformance: make(map[string]*HandlerMetrics),
		LastUpdated:        time.Now(),
	}
}

// RegisterHandler 注册错误处理器到指定状态码
func (r *ErrorRegistry) RegisterHandler(statusCode int, handler ErrorHandler) error {
	if handler == nil {
		return fmt.Errorf("error handler cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 初始化状态码对应的处理器列表
	if r.handlers[statusCode] == nil {
		r.handlers[statusCode] = make([]ErrorHandler, 0)
	}

	// 添加处理器
	r.handlers[statusCode] = append(r.handlers[statusCode], handler)

	// 按优先级排序（数值越小优先级越高）
	sort.Slice(r.handlers[statusCode], func(i, j int) bool {
		return r.handlers[statusCode][i].Priority() < r.handlers[statusCode][j].Priority()
	})

	// 同时注册到全局分发器
	r.dispatcher.RegisterHandler(handler)

	// 初始化处理器指标
	if r.config.EnableMetrics {
		handlerName := fmt.Sprintf("%T", handler)
		r.metrics.HandlerPerformance[handlerName] = &HandlerMetrics{
			LastCalled: time.Now(),
			MinTime:    time.Hour, // 初始化为很大的值
		}
	}

	return nil
}

// RegisterHandlerFunc 注册错误处理函数
func (r *ErrorRegistry) RegisterHandlerFunc(statusCode int, name string, priority int, canHandle func(int, error) bool, handleFunc ErrorHandlerFunc) error {
	handler := &FuncErrorHandler{
		name:       name,
		priority:   priority,
		canHandle:  func(err error) bool { return canHandle(statusCode, err) },
		handleFunc: handleFunc,
	}
	return r.RegisterHandler(statusCode, handler)
}

// RegisterGlobalHandler 注册全局错误处理器（不限状态码）
func (r *ErrorRegistry) RegisterGlobalHandler(handler ErrorHandler) error {
	if handler == nil {
		return fmt.Errorf("global error handler cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 添加到全局处理器列表
	r.globalHandlers = append(r.globalHandlers, handler)

	// 按优先级排序
	sort.Slice(r.globalHandlers, func(i, j int) bool {
		return r.globalHandlers[i].Priority() < r.globalHandlers[j].Priority()
	})

	// 同时注册到全局分发器
	r.dispatcher.RegisterHandler(handler)

	return nil
}

// UnregisterHandler 取消注册指定状态码的错误处理器
func (r *ErrorRegistry) UnregisterHandler(statusCode int, handler ErrorHandler) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	handlers, exists := r.handlers[statusCode]
	if !exists {
		return false
	}

	// 查找并移除指定处理器
	for i, h := range handlers {
		if h == handler {
			// 移除处理器
			r.handlers[statusCode] = append(handlers[:i], handlers[i+1:]...)
			return true
		}
	}

	return false
}

// ClearHandlers 清除指定状态码的所有错误处理器
func (r *ErrorRegistry) ClearHandlers(statusCode int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.handlers, statusCode)
}

// SetFallback 设置错误回退规则
func (r *ErrorRegistry) SetFallback(fromStatusCode int, toStatusCodes ...int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.fallbacks[fromStatusCode] = toStatusCodes
}

// HandleError 处理错误，这是核心处理逻辑
func (r *ErrorRegistry) HandleError(ctx *mvccontext.Context, statusCode int, err error) error {
	start := time.Now()
	
	// 更新错误指标
	if r.config.EnableMetrics {
		r.updateMetrics(statusCode, start)
	}

	// 首先尝试使用状态码特定的处理器
	if handlerErr := r.tryHandleByStatusCode(ctx, statusCode, err); handlerErr == nil {
		r.recordSuccess(statusCode, time.Since(start))
		return nil
	}

	// 然后尝试全局处理器
	if handlerErr := r.tryHandleByGlobal(ctx, statusCode, err); handlerErr == nil {
		r.recordSuccess(statusCode, time.Since(start))
		return nil
	}

	// 使用统一错误分发器
	if handlerErr := r.dispatcher.Dispatch(ctx, statusCode, err); handlerErr == nil {
		r.recordSuccess(statusCode, time.Since(start))
		return nil
	}

	// 如果启用了回退机制，尝试回退处理
	if r.config.EnableFallback {
		if fallbackErr := r.tryFallbackHandling(ctx, statusCode, err); fallbackErr == nil {
			r.recordFallback(statusCode, time.Since(start))
			return nil
		}
	}

	// 最后使用默认处理器
	if defaultErr := r.useDefaultHandler(ctx, statusCode, err); defaultErr == nil {
		r.recordDefault(statusCode, time.Since(start))
		return nil
	}

	// 记录未处理的错误
	r.recordUnhandled(statusCode, time.Since(start))
	return fmt.Errorf("no handler could process error for status code %d: %v", statusCode, err)
}

// tryHandleByStatusCode 尝试使用状态码特定的处理器
func (r *ErrorRegistry) tryHandleByStatusCode(ctx *mvccontext.Context, statusCode int, err error) error {
	r.mutex.RLock()
	handlers, exists := r.handlers[statusCode]
	r.mutex.RUnlock()

	if !exists || len(handlers) == 0 {
		return fmt.Errorf("no handlers registered for status code %d", statusCode)
	}

	// 按优先级尝试处理器
	for _, handler := range handlers {
		if !handler.CanHandle(statusCode, err) {
			continue
		}

		handlerStart := time.Now()
		handleErr := r.executeHandler(handler, ctx, statusCode, err)
		r.recordHandlerPerformance(handler, time.Since(handlerStart), handleErr == nil)

		if handleErr == nil {
			return nil
		}
	}

	return fmt.Errorf("all status-specific handlers failed for status code %d", statusCode)
}

// tryHandleByGlobal 尝试使用全局处理器
func (r *ErrorRegistry) tryHandleByGlobal(ctx *mvccontext.Context, statusCode int, err error) error {
	r.mutex.RLock()
	globalHandlers := make([]ErrorHandler, len(r.globalHandlers))
	copy(globalHandlers, r.globalHandlers)
	r.mutex.RUnlock()

	if len(globalHandlers) == 0 {
		return fmt.Errorf("no global handlers registered")
	}

	// 按优先级尝试全局处理器
	for _, handler := range globalHandlers {
		if !handler.CanHandle(statusCode, err) {
			continue
		}

		handlerStart := time.Now()
		handleErr := r.executeHandler(handler, ctx, statusCode, err)
		r.recordHandlerPerformance(handler, time.Since(handlerStart), handleErr == nil)

		if handleErr == nil {
			return nil
		}
	}

	return fmt.Errorf("all global handlers failed for status code %d", statusCode)
}

// executeHandler 执行处理器（带panic恢复）
func (r *ErrorRegistry) executeHandler(handler ErrorHandler, ctx *mvccontext.Context, statusCode int, err error) (result error) {
	if r.config.EnablePanicRecovery {
		defer func() {
			if r := recover(); r != nil {
				result = fmt.Errorf("handler panic: %v", r)
			}
		}()
	}

	// 设置超时
	if r.config.HandlerTimeout > 0 {
		done := make(chan error, 1)
		go func() {
			done <- handler.Handle(ctx, statusCode, err)
		}()

		select {
		case result = <-done:
			return result
		case <-time.After(r.config.HandlerTimeout):
			return fmt.Errorf("handler timeout after %v", r.config.HandlerTimeout)
		}
	}

	return handler.Handle(ctx, statusCode, err)
}

// tryFallbackHandling 尝试回退处理
func (r *ErrorRegistry) tryFallbackHandling(ctx *mvccontext.Context, statusCode int, err error) error {
	r.mutex.RLock()
	fallbackCodes, exists := r.fallbacks[statusCode]
	r.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no fallback defined for status code %d", statusCode)
	}

	// 防止无限递归的回退链长度限制
	fallbackChain := make([]int, 0, r.config.FallbackChainLength)
	fallbackChain = append(fallbackChain, statusCode)

	return r.tryFallbackChain(ctx, fallbackCodes, err, fallbackChain)
}

// tryFallbackChain 尝试回退链
func (r *ErrorRegistry) tryFallbackChain(ctx *mvccontext.Context, fallbackCodes []int, err error, fallbackChain []int) error {
	if len(fallbackChain) >= r.config.FallbackChainLength {
		return fmt.Errorf("fallback chain too long, max: %d", r.config.FallbackChainLength)
	}

	// 尝试每个回退状态码
	for _, fallbackCode := range fallbackCodes {
		// 检查是否会造成循环
		for _, chainCode := range fallbackChain {
			if chainCode == fallbackCode {
				continue // 跳过会造成循环的回退码
			}
		}

		if fallbackErr := r.tryHandleByStatusCode(ctx, fallbackCode, err); fallbackErr == nil {
			return nil
		}

		// 如果当前回退码也有回退规则，继续尝试
		r.mutex.RLock()
		nextFallbacks, hasNext := r.fallbacks[fallbackCode]
		r.mutex.RUnlock()

		if hasNext {
			newChain := make([]int, len(fallbackChain)+1)
			copy(newChain, fallbackChain)
			newChain[len(fallbackChain)] = fallbackCode

			if fallbackErr := r.tryFallbackChain(ctx, nextFallbacks, err, newChain); fallbackErr == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("all fallback handlers failed")
}

// useDefaultHandler 使用默认处理器
func (r *ErrorRegistry) useDefaultHandler(ctx *mvccontext.Context, statusCode int, err error) error {
	r.mutex.RLock()
	defaultHandler, exists := r.defaultHandlers[statusCode]
	r.mutex.RUnlock()

	if exists {
		return r.executeHandler(defaultHandler, ctx, statusCode, err)
	}

	// 最终回退：通用默认处理
	return r.handleGenericError(ctx, statusCode, err)
}

// handleGenericError 通用错误处理，最后的回退
func (r *ErrorRegistry) handleGenericError(ctx *mvccontext.Context, statusCode int, err error) error {
	// 构建响应数据
	response := map[string]any{
		"code":    statusCode,
		"message": http.StatusText(statusCode),
		"success": false,
	}

	// 如果启用详细错误信息
	if r.config.ShowDetailedError && err != nil {
		response["error"] = err.Error()
		response["timestamp"] = time.Now().Unix()
	}

	// 内容协商
	if r.config.ContentNegotiation {
		// 这里可以根据Accept头决定响应格式
		// 现在简单返回JSON
	}

	ctx.JSON(statusCode, response)
	return nil
}

// registerDefaultHandlers 注册默认错误处理器
func (r *ErrorRegistry) registerDefaultHandlers() {
	// 401 Unauthorized
	r.defaultHandlers[401] = &FuncErrorHandler{
		name:     "default-401",
		priority: 1000,
		canHandle: func(err error) bool { return true },
		handleFunc: func(ctx *mvccontext.Context, statusCode int, err error) error {
			ctx.JSON(401, map[string]any{
				"code":    401,
				"message": "Unauthorized",
				"success": false,
			})
			return nil
		},
	}

	// 403 Forbidden
	r.defaultHandlers[403] = &FuncErrorHandler{
		name:     "default-403",
		priority: 1000,
		canHandle: func(err error) bool { return true },
		handleFunc: func(ctx *mvccontext.Context, statusCode int, err error) error {
			ctx.JSON(403, map[string]any{
				"code":    403,
				"message": "Forbidden",
				"success": false,
			})
			return nil
		},
	}

	// 404 Not Found
	r.defaultHandlers[404] = &FuncErrorHandler{
		name:     "default-404",
		priority: 1000,
		canHandle: func(err error) bool { return true },
		handleFunc: func(ctx *mvccontext.Context, statusCode int, err error) error {
			ctx.JSON(404, map[string]any{
				"code":    404,
				"message": "Not Found",
				"success": false,
			})
			return nil
		},
	}

	// 500 Internal Server Error
	r.defaultHandlers[500] = &FuncErrorHandler{
		name:     "default-500",
		priority: 1000,
		canHandle: func(err error) bool { return true },
		handleFunc: func(ctx *mvccontext.Context, statusCode int, err error) error {
			response := map[string]any{
				"code":    500,
				"message": "Internal Server Error",
				"success": false,
			}
			if r.config.ShowDetailedError && err != nil {
				response["error"] = err.Error()
			}
			ctx.JSON(500, response)
			return nil
		},
	}
}

// setupDefaultFallbacks 设置默认回退规则
func (r *ErrorRegistry) setupDefaultFallbacks() {
	// 5xx错误回退到500
	for code := 501; code <= 511; code++ {
		r.fallbacks[code] = []int{500}
	}

	// 4xx错误的回退规则
	r.fallbacks[402] = []int{400}
	r.fallbacks[405] = []int{404, 400}
	r.fallbacks[406] = []int{400}
	r.fallbacks[407] = []int{401, 400}
	r.fallbacks[408] = []int{400}
	r.fallbacks[409] = []int{400}
	r.fallbacks[410] = []int{404, 400}
	r.fallbacks[411] = []int{400}
	r.fallbacks[412] = []int{400}
	r.fallbacks[413] = []int{400}
	r.fallbacks[414] = []int{400}
	r.fallbacks[415] = []int{400}
	r.fallbacks[416] = []int{400}
	r.fallbacks[417] = []int{400}
	r.fallbacks[418] = []int{400} // I'm a teapot
	r.fallbacks[421] = []int{400}
	r.fallbacks[422] = []int{400}
	r.fallbacks[423] = []int{403, 400}
	r.fallbacks[424] = []int{400}
	r.fallbacks[425] = []int{400}
	r.fallbacks[426] = []int{400}
	r.fallbacks[428] = []int{400}
	r.fallbacks[429] = []int{400}
	r.fallbacks[431] = []int{400}
	r.fallbacks[451] = []int{403, 400}
}

// 指标记录方法

// updateMetrics 更新基础指标
func (r *ErrorRegistry) updateMetrics(statusCode int, startTime time.Time) {
	if !r.config.EnableMetrics {
		return
	}

	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.TotalRequests++
	r.metrics.StatusCodeCounts[statusCode]++
	r.metrics.LastUpdated = time.Now()
}

// recordSuccess 记录成功处理
func (r *ErrorRegistry) recordSuccess(statusCode int, duration time.Duration) {
	if !r.config.EnableMetrics {
		return
	}

	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.HandledErrors++
	r.updateAverageHandleTime(duration)
}

// recordFallback 记录回退使用
func (r *ErrorRegistry) recordFallback(statusCode int, duration time.Duration) {
	if !r.config.EnableMetrics {
		return
	}

	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.FallbackUsed++
	r.metrics.HandledErrors++
	r.updateAverageHandleTime(duration)
}

// recordDefault 记录默认处理器使用
func (r *ErrorRegistry) recordDefault(statusCode int, duration time.Duration) {
	if !r.config.EnableMetrics {
		return
	}

	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.HandledErrors++
	r.updateAverageHandleTime(duration)
}

// recordUnhandled 记录未处理错误
func (r *ErrorRegistry) recordUnhandled(statusCode int, duration time.Duration) {
	if !r.config.EnableMetrics {
		return
	}

	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.UnhandledErrors++
	r.updateAverageHandleTime(duration)
}

// recordHandlerPerformance 记录处理器性能
func (r *ErrorRegistry) recordHandlerPerformance(handler ErrorHandler, duration time.Duration, success bool) {
	if !r.config.EnableMetrics {
		return
	}

	handlerName := fmt.Sprintf("%T", handler)

	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	metrics, exists := r.metrics.HandlerPerformance[handlerName]
	if !exists {
		metrics = &HandlerMetrics{
			MinTime: time.Hour, // 初始化为很大的值
		}
		r.metrics.HandlerPerformance[handlerName] = metrics
	}

	metrics.CallCount++
	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	metrics.TotalTime += duration
	metrics.AverageTime = metrics.TotalTime / time.Duration(metrics.CallCount)
	metrics.LastCalled = time.Now()

	if duration > metrics.MaxTime {
		metrics.MaxTime = duration
	}
	if duration < metrics.MinTime {
		metrics.MinTime = duration
	}
}

// updateAverageHandleTime 更新平均处理时间
func (r *ErrorRegistry) updateAverageHandleTime(duration time.Duration) {
	totalHandled := r.metrics.HandledErrors + r.metrics.UnhandledErrors
	if totalHandled > 0 {
		r.metrics.AverageHandleTime = (r.metrics.AverageHandleTime*time.Duration(totalHandled-1) + duration) / time.Duration(totalHandled)
	}
}

// 查询和管理方法

// GetHandlers 获取指定状态码的所有处理器
func (r *ErrorRegistry) GetHandlers(statusCode int) []ErrorHandler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	handlers, exists := r.handlers[statusCode]
	if !exists {
		return nil
	}

	// 返回副本以避免外部修改
	result := make([]ErrorHandler, len(handlers))
	copy(result, handlers)
	return result
}

// GetAllHandlers 获取所有注册的处理器
func (r *ErrorRegistry) GetAllHandlers() map[int][]ErrorHandler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[int][]ErrorHandler)
	for statusCode, handlers := range r.handlers {
		handlersCopy := make([]ErrorHandler, len(handlers))
		copy(handlersCopy, handlers)
		result[statusCode] = handlersCopy
	}
	return result
}

// GetGlobalHandlers 获取全局处理器
func (r *ErrorRegistry) GetGlobalHandlers() []ErrorHandler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make([]ErrorHandler, len(r.globalHandlers))
	copy(result, r.globalHandlers)
	return result
}

// GetDispatcher 获取错误分发器
func (r *ErrorRegistry) GetDispatcher() *ErrorDispatcher {
	return r.dispatcher
}

// GetConfig 获取注册器配置
func (r *ErrorRegistry) GetConfig() *RegistryConfig {
	return r.config
}

// UpdateConfig 更新注册器配置
func (r *ErrorRegistry) UpdateConfig(config *RegistryConfig) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if config != nil {
		r.config = config
	}
}

// GetMetrics 获取错误处理指标
func (r *ErrorRegistry) GetMetrics() *RegistryMetrics {
	r.metrics.mu.RLock()
	defer r.metrics.mu.RUnlock()

	// 深拷贝指标数据
	metrics := &RegistryMetrics{
		TotalRequests:      r.metrics.TotalRequests,
		HandledErrors:      r.metrics.HandledErrors,
		UnhandledErrors:    r.metrics.UnhandledErrors,
		FallbackUsed:       r.metrics.FallbackUsed,
		AverageHandleTime:  r.metrics.AverageHandleTime,
		LastUpdated:        r.metrics.LastUpdated,
		StatusCodeCounts:   make(map[int]int64),
		HandlerTypeCounts:  make(map[string]int64),
		HandlerPerformance: make(map[string]*HandlerMetrics),
	}

	for k, v := range r.metrics.StatusCodeCounts {
		metrics.StatusCodeCounts[k] = v
	}

	for k, v := range r.metrics.HandlerTypeCounts {
		metrics.HandlerTypeCounts[k] = v
	}

	for k, v := range r.metrics.HandlerPerformance {
		metrics.HandlerPerformance[k] = &HandlerMetrics{
			CallCount:    v.CallCount,
			SuccessCount: v.SuccessCount,
			FailureCount: v.FailureCount,
			AverageTime:  v.AverageTime,
			TotalTime:    v.TotalTime,
			LastCalled:   v.LastCalled,
			MaxTime:      v.MaxTime,
			MinTime:      v.MinTime,
		}
	}

	return metrics
}

// ResetMetrics 重置指标数据
func (r *ErrorRegistry) ResetMetrics() {
	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.TotalRequests = 0
	r.metrics.HandledErrors = 0
	r.metrics.UnhandledErrors = 0
	r.metrics.FallbackUsed = 0
	r.metrics.AverageHandleTime = 0
	r.metrics.StatusCodeCounts = make(map[int]int64)
	r.metrics.HandlerTypeCounts = make(map[string]int64)
	r.metrics.HandlerPerformance = make(map[string]*HandlerMetrics)
	r.metrics.LastUpdated = time.Now()
}

// GetStats 获取统计信息
func (r *ErrorRegistry) GetStats() map[string]any {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := make(map[string]any)

	// 基本统计
	stats["total_status_codes"] = len(r.handlers)
	totalHandlers := 0
	for _, handlers := range r.handlers {
		totalHandlers += len(handlers)
	}
	stats["total_handlers"] = totalHandlers
	stats["total_global_handlers"] = len(r.globalHandlers)
	stats["total_fallback_rules"] = len(r.fallbacks)

	// 配置信息
	stats["config"] = map[string]any{
		"enable_fallback":       r.config.EnableFallback,
		"enable_metrics":        r.config.EnableMetrics,
		"enable_logging":        r.config.EnableLogging,
		"fallback_chain_length": r.config.FallbackChainLength,
		"handler_timeout":       r.config.HandlerTimeout.String(),
		"enable_panic_recovery": r.config.EnablePanicRecovery,
		"show_detailed_error":   r.config.ShowDetailedError,
		"content_negotiation":   r.config.ContentNegotiation,
	}

	// 指标信息
	if r.config.EnableMetrics {
		metrics := r.GetMetrics()
		stats["metrics"] = map[string]any{
			"total_requests":       metrics.TotalRequests,
			"handled_errors":       metrics.HandledErrors,
			"unhandled_errors":     metrics.UnhandledErrors,
			"fallback_used":        metrics.FallbackUsed,
			"average_handle_time":  metrics.AverageHandleTime.String(),
			"status_code_counts":   metrics.StatusCodeCounts,
			"handler_type_counts":  metrics.HandlerTypeCounts,
			"handler_performance":  metrics.HandlerPerformance,
			"last_updated":         metrics.LastUpdated,
		}
	}

	return stats
}

// Reset 重置注册器
func (r *ErrorRegistry) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 清除所有自定义处理器
	r.handlers = make(map[int][]ErrorHandler)
	r.globalHandlers = make([]ErrorHandler, 0)

	// 重新注册默认处理器
	r.registerDefaultHandlers()

	// 重置指标
	if r.config.EnableMetrics {
		r.ResetMetrics()
	}
}

// 全局错误注册器
var globalErrorRegistry = NewErrorRegistry(nil)

// GetGlobalErrorRegistry 获取全局错误注册器
func GetGlobalErrorRegistry() *ErrorRegistry {
	return globalErrorRegistry
}

// RegisterErrorHandler 注册全局错误处理器
func RegisterErrorHandler(statusCode int, handler ErrorHandler) error {
	return globalErrorRegistry.RegisterHandler(statusCode, handler)
}

// RegisterErrorHandlerFunc 注册全局错误处理函数
func RegisterErrorHandlerFunc(statusCode int, name string, priority int, canHandle func(int, error) bool, handleFunc ErrorHandlerFunc) error {
	return globalErrorRegistry.RegisterHandlerFunc(statusCode, name, priority, canHandle, handleFunc)
}

// HandleGlobalError 处理全局错误
func HandleGlobalError(ctx *mvccontext.Context, statusCode int, err error) error {
	return globalErrorRegistry.HandleError(ctx, statusCode, err)
}

// PrintRegistryInfo 打印注册器信息
func PrintRegistryInfo() {
	stats := globalErrorRegistry.GetStats()

	fmt.Println("=== Error Registry Statistics ===")
	fmt.Printf("Total Status Codes: %v\n", stats["total_status_codes"])
	fmt.Printf("Total Handlers: %v\n", stats["total_handlers"])
	fmt.Printf("Total Global Handlers: %v\n", stats["total_global_handlers"])
	fmt.Printf("Total Fallback Rules: %v\n", stats["total_fallback_rules"])

	if config, ok := stats["config"].(map[string]any); ok {
		fmt.Println("\nConfiguration:")
		for key, value := range config {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	if metrics, ok := stats["metrics"].(map[string]any); ok {
		fmt.Println("\nMetrics:")
		for key, value := range metrics {
			if key != "status_code_counts" && key != "handler_type_counts" && key != "handler_performance" {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}
}