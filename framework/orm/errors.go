// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"gorm.io/gorm"
)

// ErrorCode 错误代码
type ErrorCode string

const (
	// ErrCodeConnection 连接错误
	ErrCodeConnection ErrorCode = "CONNECTION_ERROR"
	// ErrCodeQuery 查询错误
	ErrCodeQuery ErrorCode = "QUERY_ERROR"
	// ErrCodeTransaction 事务错误
	ErrCodeTransaction ErrorCode = "TRANSACTION_ERROR"
	// ErrCodeValidation 验证错误
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// ErrCodeNotFound 记录未找到
	ErrCodeNotFound ErrorCode = "RECORD_NOT_FOUND"
	// ErrCodeDuplicate 重复记录
	ErrCodeDuplicate ErrorCode = "DUPLICATE_RECORD"
	// ErrCodeConstraint 约束错误
	ErrCodeConstraint ErrorCode = "CONSTRAINT_ERROR"
	// ErrCodeTimeout 超时错误
	ErrCodeTimeout ErrorCode = "TIMEOUT_ERROR"
	// ErrCodeCache 缓存错误
	ErrCodeCache ErrorCode = "CACHE_ERROR"
	// ErrCodeMigration 迁移错误
	ErrCodeMigration ErrorCode = "MIGRATION_ERROR"
	// ErrCodeUnknown 未知错误
	ErrCodeUnknown ErrorCode = "UNKNOWN_ERROR"
)

// ORMError ORM错误结构
type ORMError struct {
	// 错误代码
	Code ErrorCode `json:"code"`
	// 错误消息
	Message string `json:"message"`
	// 原始错误
	Cause error `json:"-"`
	// 堆栈信息
	Stack []StackFrame `json:"stack,omitempty"`
	// 上下文信息
	Context map[string]interface{} `json:"context,omitempty"`
	// SQL语句
	SQL string `json:"sql,omitempty"`
	// SQL参数
	Args []interface{} `json:"args,omitempty"`
	// 表名
	Table string `json:"table,omitempty"`
	// 操作类型
	Operation string `json:"operation,omitempty"`
}

// StackFrame 堆栈帧
type StackFrame struct {
	// 函数名
	Function string `json:"function"`
	// 文件路径
	File string `json:"file"`
	// 行号
	Line int `json:"line"`
}

// Error 实现error接口
func (e *ORMError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现errors.Unwrap接口
func (e *ORMError) Unwrap() error {
	return e.Cause
}

// Is 实现errors.Is接口
func (e *ORMError) Is(target error) bool {
	if target == nil {
		return false
	}

	// 检查错误代码
	if ormErr, ok := target.(*ORMError); ok {
		return e.Code == ormErr.Code
	}

	// 检查原始错误
	return errors.Is(e.Cause, target)
}

// WithContext 添加上下文信息
func (e *ORMError) WithContext(key string, value interface{}) *ORMError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSQL 添加SQL信息
func (e *ORMError) WithSQL(sql string, args ...interface{}) *ORMError {
	e.SQL = sql
	e.Args = args
	return e
}

// WithTable 添加表名信息
func (e *ORMError) WithTable(table string) *ORMError {
	e.Table = table
	return e
}

// WithOperation 添加操作类型信息
func (e *ORMError) WithOperation(operation string) *ORMError {
	e.Operation = operation
	return e
}

// NewORMError 创建ORM错误
func NewORMError(code ErrorCode, message string, cause error) *ORMError {
	return &ORMError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Stack:   captureStack(2), // 跳过当前函数和调用者
		Context: make(map[string]interface{}),
	}
}

// captureStack 捕获堆栈信息
func captureStack(skip int) []StackFrame {
	const maxStackDepth = 32
	var frames []StackFrame

	for i := skip; i < maxStackDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// 过滤掉runtime和系统相关的堆栈
		funcName := fn.Name()
		if strings.Contains(funcName, "runtime.") ||
			strings.Contains(funcName, "syscall.") ||
			strings.Contains(funcName, "reflect.") {
			continue
		}

		frames = append(frames, StackFrame{
			Function: funcName,
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// WrapError 包装错误
func WrapError(err error, code ErrorCode, message string) *ORMError {
	if err == nil {
		return nil
	}

	// 如果已经是ORMError，直接返回
	if ormErr, ok := err.(*ORMError); ok {
		return ormErr
	}

	return NewORMError(code, message, err)
}

// WrapGormError 包装GORM错误
func WrapGormError(err error) *ORMError {
	if err == nil {
		return nil
	}

	// 如果已经是ORMError，直接返回
	if ormErr, ok := err.(*ORMError); ok {
		return ormErr
	}

	// 根据GORM错误类型进行分类
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return NewORMError(ErrCodeNotFound, "记录未找到", err)
	case errors.Is(err, gorm.ErrInvalidTransaction):
		return NewORMError(ErrCodeTransaction, "无效的事务", err)
	case errors.Is(err, gorm.ErrNotImplemented):
		return NewORMError(ErrCodeQuery, "功能未实现", err)
	case errors.Is(err, gorm.ErrMissingWhereClause):
		return NewORMError(ErrCodeQuery, "缺少WHERE子句", err)
	case errors.Is(err, gorm.ErrUnsupportedRelation):
		return NewORMError(ErrCodeQuery, "不支持的关联关系", err)
	case errors.Is(err, gorm.ErrPrimaryKeyRequired):
		return NewORMError(ErrCodeValidation, "需要主键", err)
	case errors.Is(err, gorm.ErrModelValueRequired):
		return NewORMError(ErrCodeValidation, "需要模型值", err)
	case errors.Is(err, gorm.ErrInvalidData):
		return NewORMError(ErrCodeValidation, "无效的数据", err)
	case errors.Is(err, gorm.ErrUnsupportedDriver):
		return NewORMError(ErrCodeConnection, "不支持的数据库驱动", err)
	case errors.Is(err, gorm.ErrRegistered):
		return NewORMError(ErrCodeConnection, "已注册的回调", err)
	case errors.Is(err, gorm.ErrInvalidField):
		return NewORMError(ErrCodeValidation, "无效的字段", err)
	case errors.Is(err, gorm.ErrEmptySlice):
		return NewORMError(ErrCodeValidation, "空切片", err)
	case errors.Is(err, gorm.ErrDryRunModeUnsupported):
		return NewORMError(ErrCodeQuery, "不支持DryRun模式", err)
	case errors.Is(err, gorm.ErrInvalidDB):
		return NewORMError(ErrCodeConnection, "无效的数据库连接", err)
	case errors.Is(err, gorm.ErrInvalidValue):
		return NewORMError(ErrCodeValidation, "无效的值", err)
	case errors.Is(err, gorm.ErrInvalidValueOfLength):
		return NewORMError(ErrCodeValidation, "无效的值长度", err)
	default:
		// 检查错误消息中的关键词
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "connection"):
			return NewORMError(ErrCodeConnection, "数据库连接错误", err)
		case strings.Contains(errMsg, "timeout"):
			return NewORMError(ErrCodeTimeout, "操作超时", err)
		case strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique"):
			return NewORMError(ErrCodeDuplicate, "重复记录", err)
		case strings.Contains(errMsg, "constraint") || strings.Contains(errMsg, "foreign key"):
			return NewORMError(ErrCodeConstraint, "约束错误", err)
		default:
			return NewORMError(ErrCodeUnknown, "未知错误", err)
		}
	}
}

// ErrorHandler 错误处理器
type ErrorHandler struct {
	// 是否启用堆栈捕获
	EnableStackTrace bool
	// 是否记录SQL信息
	LogSQL bool
	// 错误回调函数
	OnError func(*ORMError)
}

// DefaultErrorHandler 默认错误处理器
var DefaultErrorHandler = &ErrorHandler{
	EnableStackTrace: true,
	LogSQL:           true,
	OnError:          nil,
}

// Handle 处理错误
func (h *ErrorHandler) Handle(err error) *ORMError {
	if err == nil {
		return nil
	}

	ormErr := WrapGormError(err)

	// 如果禁用堆栈跟踪，清空堆栈信息
	if !h.EnableStackTrace {
		ormErr.Stack = nil
	}

	// 调用错误回调
	if h.OnError != nil {
		h.OnError(ormErr)
	}

	return ormErr
}

// HandleWithContext 处理错误并添加上下文
func (h *ErrorHandler) HandleWithContext(err error, context map[string]interface{}) *ORMError {
	ormErr := h.Handle(err)
	if ormErr != nil && context != nil {
		for k, v := range context {
			ormErr.WithContext(k, v)
		}
	}
	return ormErr
}

// ============= 便捷函数 =============

// HandleError 处理错误
func HandleError(err error) *ORMError {
	return DefaultErrorHandler.Handle(err)
}

// HandleErrorWithContext 处理错误并添加上下文
func HandleErrorWithContext(err error, context map[string]interface{}) *ORMError {
	return DefaultErrorHandler.HandleWithContext(err, context)
}

// IsNotFound 检查是否为记录未找到错误
func IsNotFound(err error) bool {
	if ormErr, ok := err.(*ORMError); ok {
		return ormErr.Code == ErrCodeNotFound
	}
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsDuplicate 检查是否为重复记录错误
func IsDuplicate(err error) bool {
	if ormErr, ok := err.(*ORMError); ok {
		return ormErr.Code == ErrCodeDuplicate
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique")
}

// IsConstraint 检查是否为约束错误
func IsConstraint(err error) bool {
	if ormErr, ok := err.(*ORMError); ok {
		return ormErr.Code == ErrCodeConstraint
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "constraint") || strings.Contains(errMsg, "foreign key")
}

// IsTimeout 检查是否为超时错误
func IsTimeout(err error) bool {
	if ormErr, ok := err.(*ORMError); ok {
		return ormErr.Code == ErrCodeTimeout
	}
	return strings.Contains(err.Error(), "timeout")
}

// IsConnection 检查是否为连接错误
func IsConnection(err error) bool {
	if ormErr, ok := err.(*ORMError); ok {
		return ormErr.Code == ErrCodeConnection
	}
	return strings.Contains(err.Error(), "connection")
}

// ============= 错误恢复函数 =============

// RecoverFromPanic 从panic中恢复并转换为ORM错误
func RecoverFromPanic() *ORMError {
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		case string:
			err = errors.New(v)
		default:
			err = fmt.Errorf("panic: %v", v)
		}

		return NewORMError(ErrCodeUnknown, "发生panic", err)
	}
	return nil
}

// SafeExecute 安全执行函数，捕获panic并转换为错误
func SafeExecute(fn func() error) *ORMError {
	defer func() {
		if r := recover(); r != nil {
			// 这里无法直接返回错误，需要通过其他方式处理
			// 可以考虑使用channel或者全局变量
		}
	}()

	if err := fn(); err != nil {
		return HandleError(err)
	}

	return nil
}

// ============= 错误统计 =============

// ErrorStats 错误统计
type ErrorStats struct {
	// 总错误数
	Total int64 `json:"total"`
	// 按错误代码分组的统计
	ByCode map[ErrorCode]int64 `json:"by_code"`
	// 按表名分组的统计
	ByTable map[string]int64 `json:"by_table"`
	// 按操作类型分组的统计
	ByOperation map[string]int64 `json:"by_operation"`
}

// NewErrorStats 创建错误统计
func NewErrorStats() *ErrorStats {
	return &ErrorStats{
		ByCode:      make(map[ErrorCode]int64),
		ByTable:     make(map[string]int64),
		ByOperation: make(map[string]int64),
	}
}

// Record 记录错误
func (es *ErrorStats) Record(err *ORMError) {
	if err == nil {
		return
	}

	es.Total++
	es.ByCode[err.Code]++

	if err.Table != "" {
		es.ByTable[err.Table]++
	}

	if err.Operation != "" {
		es.ByOperation[err.Operation]++
	}
}

// Reset 重置统计
func (es *ErrorStats) Reset() {
	es.Total = 0
	es.ByCode = make(map[ErrorCode]int64)
	es.ByTable = make(map[string]int64)
	es.ByOperation = make(map[string]int64)
}

// GetTopErrors 获取最常见的错误
func (es *ErrorStats) GetTopErrors(limit int) []struct {
	Code  ErrorCode `json:"code"`
	Count int64     `json:"count"`
} {
	type errorCount struct {
		Code  ErrorCode
		Count int64
	}

	var errors []errorCount
	for code, count := range es.ByCode {
		errors = append(errors, errorCount{Code: code, Count: count})
	}

	// 简单排序（按计数降序）
	for i := 0; i < len(errors)-1; i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[i].Count < errors[j].Count {
				errors[i], errors[j] = errors[j], errors[i]
			}
		}
	}

	if limit > 0 && limit < len(errors) {
		errors = errors[:limit]
	}

	result := make([]struct {
		Code  ErrorCode `json:"code"`
		Count int64     `json:"count"`
	}, len(errors))

	for i, e := range errors {
		result[i].Code = e.Code
		result[i].Count = e.Count
	}

	return result
}

// 全局错误统计
var GlobalErrorStats = NewErrorStats()

// ============= 错误中间件 =============

// ErrorMiddleware 错误中间件
type ErrorMiddleware struct {
	handler *ErrorHandler
	stats   *ErrorStats
}

// NewErrorMiddleware 创建错误中间件
func NewErrorMiddleware(handler *ErrorHandler, stats *ErrorStats) *ErrorMiddleware {
	if handler == nil {
		handler = DefaultErrorHandler
	}
	if stats == nil {
		stats = GlobalErrorStats
	}

	return &ErrorMiddleware{
		handler: handler,
		stats:   stats,
	}
}

// Handle 处理错误
func (em *ErrorMiddleware) Handle(err error) *ORMError {
	ormErr := em.handler.Handle(err)
	if ormErr != nil {
		em.stats.Record(ormErr)
	}
	return ormErr
}

// HandleWithContext 处理错误并添加上下文
func (em *ErrorMiddleware) HandleWithContext(err error, context map[string]interface{}) *ORMError {
	ormErr := em.handler.HandleWithContext(err, context)
	if ormErr != nil {
		em.stats.Record(ormErr)
	}
	return ormErr
}

// 全局错误中间件
var GlobalErrorMiddleware = NewErrorMiddleware(DefaultErrorHandler, GlobalErrorStats)

// ============= 错误格式化 =============

// FormatError 格式化错误信息
func FormatError(err *ORMError, includeStack bool) string {
	if err == nil {
		return ""
	}

	var builder strings.Builder

	// 基本错误信息
	builder.WriteString(fmt.Sprintf("错误代码: %s\n", err.Code))
	builder.WriteString(fmt.Sprintf("错误消息: %s\n", err.Message))

	if err.Cause != nil {
		builder.WriteString(fmt.Sprintf("原始错误: %v\n", err.Cause))
	}

	// SQL信息
	if err.SQL != "" {
		builder.WriteString(fmt.Sprintf("SQL语句: %s\n", err.SQL))
		if len(err.Args) > 0 {
			builder.WriteString(fmt.Sprintf("SQL参数: %v\n", err.Args))
		}
	}

	// 表和操作信息
	if err.Table != "" {
		builder.WriteString(fmt.Sprintf("表名: %s\n", err.Table))
	}
	if err.Operation != "" {
		builder.WriteString(fmt.Sprintf("操作: %s\n", err.Operation))
	}

	// 上下文信息
	if len(err.Context) > 0 {
		builder.WriteString("上下文信息:\n")
		for k, v := range err.Context {
			builder.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	// 堆栈信息
	if includeStack && len(err.Stack) > 0 {
		builder.WriteString("堆栈跟踪:\n")
		for _, frame := range err.Stack {
			builder.WriteString(fmt.Sprintf("  %s\n", frame.Function))
			builder.WriteString(fmt.Sprintf("    %s:%d\n", frame.File, frame.Line))
		}
	}

	return builder.String()
}

// FormatErrorJSON 格式化错误为JSON
func FormatErrorJSON(ormErr *ORMError) ([]byte, error) {
	if ormErr == nil {
		return nil, nil
	}

	// 创建一个可序列化的错误结构
	type serializableError struct {
		Code      ErrorCode              `json:"code"`
		Message   string                 `json:"message"`
		Cause     string                 `json:"cause,omitempty"`
		Stack     []StackFrame           `json:"stack,omitempty"`
		Context   map[string]interface{} `json:"context,omitempty"`
		SQL       string                 `json:"sql,omitempty"`
		Args      []interface{}          `json:"args,omitempty"`
		Table     string                 `json:"table,omitempty"`
		Operation string                 `json:"operation,omitempty"`
	}

	serErr := serializableError{
		Code:      ormErr.Code,
		Message:   ormErr.Message,
		Stack:     ormErr.Stack,
		Context:   ormErr.Context,
		SQL:       ormErr.SQL,
		Args:      ormErr.Args,
		Table:     ormErr.Table,
		Operation: ormErr.Operation,
	}

	if ormErr.Cause != nil {
		serErr.Cause = ormErr.Cause.Error()
	}

	return fmt.Appendf(nil, "%+v", serErr), nil
}
