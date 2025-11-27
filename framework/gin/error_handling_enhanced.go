// Package gin - 增强错误处理系统
// 提供强大的错误处理、恢复和监控功能
package gin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

// =============================================================================
// 增强错误类型定义
// =============================================================================

// ErrorCode 错误码类型
type ErrorCode string

const (
	// 通用错误码
	ErrorCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrorCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodeTimeout      ErrorCode = "TIMEOUT"
	ErrorCodeTooManyReq   ErrorCode = "TOO_MANY_REQUESTS"
	
	// 业务错误码
	ErrorCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrorCodeBinding      ErrorCode = "BINDING_ERROR"
	ErrorCodeDatabase     ErrorCode = "DATABASE_ERROR"
	ErrorCodeNetwork      ErrorCode = "NETWORK_ERROR"
	ErrorCodeAuth         ErrorCode = "AUTH_ERROR"
)

// EnhancedError 增强错误结构
type EnhancedError struct {
	Code      ErrorCode   `json:"code"`
	Message   string      `json:"message"`
	Details   string      `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Cause     error       `json:"-"`
	Stack     []StackFrame `json:"stack,omitempty"`
	Context   H           `json:"context,omitempty"`
	Retryable bool        `json:"retryable"`
	HTTPCode  int         `json:"-"`
}

// StackFrame 堆栈帧信息
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Error 实现error接口
func (e *EnhancedError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现errors.Unwrap
func (e *EnhancedError) Unwrap() error {
	return e.Cause
}

// WithCode 设置错误码
func (e *EnhancedError) WithCode(code ErrorCode) *EnhancedError {
	e.Code = code
	return e
}

// WithMessage 设置错误消息
func (e *EnhancedError) WithMessage(message string) *EnhancedError {
	e.Message = message
	return e
}

// WithDetails 设置错误详情
func (e *EnhancedError) WithDetails(details string) *EnhancedError {
	e.Details = details
	return e
}

// WithRequestID 设置请求ID
func (e *EnhancedError) WithRequestID(requestID string) *EnhancedError {
	e.RequestID = requestID
	return e
}

// WithUserID 设置用户ID
func (e *EnhancedError) WithUserID(userID string) *EnhancedError {
	e.UserID = userID
	return e
}

// WithContext 设置上下文信息
func (e *EnhancedError) WithContext(key string, value any) *EnhancedError {
	if e.Context == nil {
		e.Context = make(H)
	}
	e.Context[key] = value
	return e
}

// WithHTTPCode 设置HTTP状态码
func (e *EnhancedError) WithHTTPCode(code int) *EnhancedError {
	e.HTTPCode = code
	return e
}

// WithRetryable 设置是否可重试
func (e *EnhancedError) WithRetryable(retryable bool) *EnhancedError {
	e.Retryable = retryable
	return e
}

// ToJSON 转换为JSON格式
func (e *EnhancedError) ToJSON() H {
	result := H{
		"code":      e.Code,
		"message":   e.Message,
		"timestamp": e.Timestamp.Format(time.RFC3339),
		"retryable": e.Retryable,
	}
	
	if e.Details != "" {
		result["details"] = e.Details
	}
	
	if e.RequestID != "" {
		result["request_id"] = e.RequestID
	}
	
	if e.UserID != "" {
		result["user_id"] = e.UserID
	}
	
	if len(e.Context) > 0 {
		result["context"] = e.Context
	}
	
	if len(e.Stack) > 0 {
		result["stack"] = e.Stack
	}
	
	return result
}

// =============================================================================
// 错误构造函数
// =============================================================================

// NewError 创建新的增强错误
func NewError(code ErrorCode, message string) *EnhancedError {
	return &EnhancedError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Stack:     captureStack(2), // 跳过当前函数和调用者
		HTTPCode:  errorCodeToHTTP(code),
		Retryable: isRetryableError(code),
	}
}

// WrapError 包装现有错误
func WrapError(err error, code ErrorCode, message string) *EnhancedError {
	if err == nil {
		return nil
	}
	
	// 如果已经是EnhancedError，直接返回
	if enhanced, ok := err.(*EnhancedError); ok {
		if code != "" {
			enhanced.Code = code
		}
		if message != "" {
			enhanced.Message = message
		}
		return enhanced
	}
	
	return &EnhancedError{
		Code:      code,
		Message:   message,
		Details:   err.Error(),
		Timestamp: time.Now(),
		Cause:     err,
		Stack:     captureStack(2),
		HTTPCode:  errorCodeToHTTP(code),
		Retryable: isRetryableError(code),
	}
}

// 错误码到HTTP状态码的映射
func errorCodeToHTTP(code ErrorCode) int {
	switch code {
	case ErrorCodeBadRequest, ErrorCodeValidation, ErrorCodeBinding:
		return http.StatusBadRequest
	case ErrorCodeUnauthorized, ErrorCodeAuth:
		return http.StatusUnauthorized
	case ErrorCodeForbidden:
		return http.StatusForbidden
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeTimeout:
		return http.StatusRequestTimeout
	case ErrorCodeTooManyReq:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// 判断错误是否可重试
func isRetryableError(code ErrorCode) bool {
	switch code {
	case ErrorCodeTimeout, ErrorCodeNetwork, ErrorCodeInternal:
		return true
	default:
		return false
	}
}

// 捕获堆栈信息
func captureStack(skip int) []StackFrame {
	var frames []StackFrame
	for i := skip; i < skip+10; i++ { // 最多捕获10层
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}
	return frames
}

// =============================================================================
// 错误处理器
// =============================================================================

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	Handle(c *Context, err error)
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	Logger ErrorLogger
}

// ErrorLogger 错误处理日志接口
type ErrorLogger interface {
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Info(msg string, fields ...any)
}

// Handle 处理错误
func (h *DefaultErrorHandler) Handle(c *Context, err error) {
	enhanced := ensureEnhancedError(err)
	
	// 设置请求上下文信息
	if enhanced.RequestID == "" {
		enhanced.RequestID = getRequestID(c)
	}
	
	// 添加请求上下文
	enhanced.WithContext("method", c.Request().Method)
	enhanced.WithContext("path", c.Request().URL.Path)
	enhanced.WithContext("user_agent", c.GetHeader("User-Agent"))
	enhanced.WithContext("ip", c.ClientIP())
	
	// 记录日志
	if h.Logger != nil {
		if enhanced.HTTPCode >= 500 {
			h.Logger.Error("Request error", 
				"error", enhanced.Error(),
				"code", enhanced.Code,
				"request_id", enhanced.RequestID,
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
			)
		} else {
			h.Logger.Warn("Client error",
				"error", enhanced.Error(),
				"code", enhanced.Code,
				"request_id", enhanced.RequestID,
			)
		}
	}
	
	// 返回错误响应
	if !c.Writer().Written() {
		c.JSON(enhanced.HTTPCode, enhanced.ToJSON())
	}
}

// 确保错误是EnhancedError类型
func ensureEnhancedError(err error) *EnhancedError {
	if enhanced, ok := err.(*EnhancedError); ok {
		return enhanced
	}
	return WrapError(err, ErrorCodeInternal, "Internal server error")
}

// 获取请求ID
func getRequestID(c *Context) string {
	// 尝试从Header获取
	if rid := c.GetHeader("X-Request-ID"); rid != "" {
		return rid
	}
	if rid := c.GetHeader("X-Trace-ID"); rid != "" {
		return rid
	}
	
	// 尝试从Context获取
	if rid, exists := c.Get("request_id"); exists {
		if ridStr, ok := rid.(string); ok {
			return ridStr
		}
	}
	
	// 生成新的请求ID
	return generateRequestID()
}

// 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

// =============================================================================
// 错误恢复中间件
// =============================================================================

// RecoveryConfig 错误恢复中间件配置
type RecoveryConfig struct {
	Skipper         func(*Context) bool
	ErrorHandler    ErrorHandler
	Logger          ErrorLogger
	EnableStackAll  bool // 是否记录完整堆栈
	EnableRequestID bool // 是否生成请求ID
	PanicHandler    func(*Context, any) // 自定义panic处理
}

// DefaultRecoveryConfig 默认恢复配置
var DefaultRecoveryConfig = RecoveryConfig{
	Skipper:         func(*Context) bool { return false },
	ErrorHandler:    &DefaultErrorHandler{},
	EnableStackAll:  false,
	EnableRequestID: true,
}

// RecoveryWithConfig 带配置的恢复中间件
func RecoveryWithConfig(config RecoveryConfig) HandlerFunc {
	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultRecoveryConfig.ErrorHandler
	}
	
	return func(c *Context) {
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}
		
		// 生成请求ID
		if config.EnableRequestID {
			if requestID := getRequestID(c); requestID != "" {
				c.Set("request_id", requestID)
				c.Header("X-Request-ID", requestID)
			}
		}
		
		defer func() {
			if r := recover(); r != nil {
				var err error
				
				// 处理panic
				if config.PanicHandler != nil {
					config.PanicHandler(c, r)
					return
				}
				
				// 转换panic为错误
				switch x := r.(type) {
				case string:
					err = NewError(ErrorCodeInternal, x)
				case error:
					err = WrapError(x, ErrorCodeInternal, "Panic recovered")
				default:
					err = NewError(ErrorCodeInternal, fmt.Sprintf("Panic recovered: %v", x))
				}
				
				// 如果启用完整堆栈，记录详细信息
				if config.EnableStackAll {
					stack := make([]byte, 4096)
					length := runtime.Stack(stack, true)
					if enhanced, ok := err.(*EnhancedError); ok {
						enhanced.WithDetails(string(stack[:length]))
					}
				}
				
				// 处理错误
				config.ErrorHandler.Handle(c, err)
				c.Abort()
			}
		}()
		
		c.Next()
		
		// 处理请求过程中收集的错误
		if len(c.Errors) > 0 {
			lastError := c.Errors[len(c.Errors)-1]
			config.ErrorHandler.Handle(c, lastError)
		}
	}
}

// ErrorRecovery 默认恢复中间件
func ErrorRecovery() HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig)
}

// =============================================================================
// 错误监控和统计
// =============================================================================

// ErrorMonitor 错误监控器
type ErrorMonitor struct {
	mu     sync.RWMutex
	stats  map[ErrorCode]*ErrorStats
	logger ErrorLogger
}

// ErrorStats 错误统计
type ErrorStats struct {
	Code         ErrorCode     `json:"code"`
	Count        int64         `json:"count"`
	LastOccurred time.Time     `json:"last_occurred"`
	FirstSeen    time.Time     `json:"first_seen"`
	AvgDuration  time.Duration `json:"avg_duration"`
	Examples     []*EnhancedError `json:"examples,omitempty"`
}

// NewErrorMonitor 创建错误监控器
func NewErrorMonitor() *ErrorMonitor {
	return &ErrorMonitor{
		stats: make(map[ErrorCode]*ErrorStats),
	}
}

// Record 记录错误
func (m *ErrorMonitor) Record(err *EnhancedError) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	stats, exists := m.stats[err.Code]
	if !exists {
		stats = &ErrorStats{
			Code:      err.Code,
			FirstSeen: err.Timestamp,
			Examples:  make([]*EnhancedError, 0, 5),
		}
		m.stats[err.Code] = stats
	}
	
	stats.Count++
	stats.LastOccurred = err.Timestamp
	
	// 保留最新的5个错误示例
	if len(stats.Examples) < 5 {
		stats.Examples = append(stats.Examples, err)
	} else {
		stats.Examples = append(stats.Examples[1:], err)
	}
}

// GetStats 获取错误统计
func (m *ErrorMonitor) GetStats() map[ErrorCode]*ErrorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[ErrorCode]*ErrorStats)
	for code, stats := range m.stats {
		result[code] = &ErrorStats{
			Code:         stats.Code,
			Count:        stats.Count,
			LastOccurred: stats.LastOccurred,
			FirstSeen:    stats.FirstSeen,
			AvgDuration:  stats.AvgDuration,
			Examples:     append([]*EnhancedError(nil), stats.Examples...),
		}
	}
	return result
}

// Reset 重置统计
func (m *ErrorMonitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stats = make(map[ErrorCode]*ErrorStats)
}

// 全局错误监控器
var globalErrorMonitor = NewErrorMonitor()

// GetGlobalErrorStats 获取全局错误统计
func GetGlobalErrorStats() map[ErrorCode]*ErrorStats {
	return globalErrorMonitor.GetStats()
}

// =============================================================================
// 错误处理中间件
// =============================================================================

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware(handler ErrorHandler) HandlerFunc {
	if handler == nil {
		handler = &DefaultErrorHandler{}
	}
	
	return func(c *Context) {
		c.Next()
		
		// 检查是否有错误需要处理
		if len(c.Errors) > 0 && !c.Writer().Written() {
			lastError := c.Errors[len(c.Errors)-1]
			handler.Handle(c, lastError)
		}
	}
}

// =============================================================================
// 上下文错误处理方法扩展
// =============================================================================

// AbortWithEnhancedError 使用增强错误中止请求
func (c *Context) AbortWithEnhancedError(err *EnhancedError) {
	if err.RequestID == "" {
		err.RequestID = getRequestID(c)
	}
	
	c.Error(err)
	c.AbortWithStatus(err.HTTPCode)
	
	// 记录到全局监控器
	globalErrorMonitor.Record(err)
}

// NewErrorAndAbort 创建错误并中止请求
func (c *Context) NewErrorAndAbort(code ErrorCode, message string) {
	err := NewError(code, message).WithRequestID(getRequestID(c))
	c.AbortWithEnhancedError(err)
}

// WrapErrorAndAbort 包装错误并中止请求
func (c *Context) WrapErrorAndAbort(err error, code ErrorCode, message string) {
	enhanced := WrapError(err, code, message).WithRequestID(getRequestID(c))
	c.AbortWithEnhancedError(enhanced)
}

// HandleError 处理错误（不中止请求）
func (c *Context) HandleError(err error) {
	if err == nil {
		return
	}
	
	enhanced := ensureEnhancedError(err)
	if enhanced.RequestID == "" {
		enhanced.RequestID = getRequestID(c)
	}
	
	c.Error(enhanced)
	globalErrorMonitor.Record(enhanced)
}

// =============================================================================
// 工具函数
// =============================================================================

// IsEnhancedError 检查是否为增强错误
func IsEnhancedError(err error) bool {
	_, ok := err.(*EnhancedError)
	return ok
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	if enhanced, ok := err.(*EnhancedError); ok {
		return enhanced.Code
	}
	return ErrorCodeInternal
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	if enhanced, ok := err.(*EnhancedError); ok {
		return enhanced.Retryable
	}
	return false
}

// =============================================================================
// 错误响应格式化
// =============================================================================

// ErrorResponse 标准错误响应格式
type ErrorResponse struct {
	Success   bool   `json:"success"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// FormatErrorResponse 格式化错误响应
func FormatErrorResponse(err *EnhancedError) ErrorResponse {
	return ErrorResponse{
		Success:   false,
		Code:      string(err.Code),
		Message:   err.Message,
		Details:   err.Details,
		RequestID: err.RequestID,
		Timestamp: err.Timestamp.Format(time.RFC3339),
	}
}

// JSONErrorHandler JSON格式错误处理器
type JSONErrorHandler struct {
	Logger         ErrorLogger
	ShowStack      bool
	ShowDetails    bool
	CustomFormat   func(*EnhancedError) any
}

// Handle 处理错误
func (h *JSONErrorHandler) Handle(c *Context, err error) {
	enhanced := ensureEnhancedError(err)
	
	if enhanced.RequestID == "" {
		enhanced.RequestID = getRequestID(c)
	}
	
	// 记录日志
	if h.Logger != nil {
		if enhanced.HTTPCode >= 500 {
			h.Logger.Error("Request error", "error", enhanced.Error())
		}
	}
	
	// 返回响应
	if !c.Writer().Written() {
		var response any
		
		if h.CustomFormat != nil {
			response = h.CustomFormat(enhanced)
		} else {
			resp := FormatErrorResponse(enhanced)
			
			if h.ShowDetails && enhanced.Details != "" {
				resp.Details = enhanced.Details
			}
			
			response = resp
		}
		
		c.JSON(enhanced.HTTPCode, response)
	}
}

// =============================================================================
// 超时处理
// =============================================================================

// TimeoutHandler 超时处理中间件
func TimeoutHandler(timeout time.Duration) HandlerFunc {
	return func(c *Context) {
		ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
		defer cancel()
		
		// 替换请求上下文
		c.Request().WithContext(ctx)
		
		done := make(chan struct{})
		go func() {
			defer close(done)
			c.Next()
		}()
		
		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			if ctx.Err() == context.DeadlineExceeded {
				err := NewError(ErrorCodeTimeout, "Request timeout")
				c.AbortWithEnhancedError(err)
			}
		}
	}
}

// =============================================================================
// 默认日志实现
// =============================================================================

// DefaultErrorLogger 默认错误日志实现
type DefaultErrorLogger struct{}

// Error 记录错误日志
func (l *DefaultErrorLogger) Error(msg string, fields ...any) {
	log.Printf("[ERROR] %s %v", msg, fields)
}

// Warn 记录警告日志
func (l *DefaultErrorLogger) Warn(msg string, fields ...any) {
	log.Printf("[WARN] %s %v", msg, fields)
}

// Info 记录信息日志  
func (l *DefaultErrorLogger) Info(msg string, fields ...any) {
	log.Printf("[INFO] %s %v", msg, fields)
}

// 设置默认日志
func init() {
	if DefaultRecoveryConfig.Logger == nil {
		DefaultRecoveryConfig.Logger = &DefaultErrorLogger{}
	}
}