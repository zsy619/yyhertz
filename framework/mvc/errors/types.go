package errors

import (
	"fmt"
	"sync"
	"time"
	
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// =============================================================================
// 模块：核心数据类型定义
// 职责：定义所有错误处理相关的数据结构和接口
// =============================================================================

// ErrorContext 错误上下文信息
type ErrorContext struct {
	StatusCode    int            `json:"status_code"`
	StatusText    string         `json:"status_text"`
	ErrorMessage  string         `json:"error_message"`
	RequestPath   string         `json:"request_path"`
	RequestMethod string         `json:"request_method"`
	UserAgent     string         `json:"user_agent"`
	Timestamp     time.Time      `json:"timestamp"`
	RequestID     string         `json:"request_id,omitempty"`
	Details       map[string]any `json:"details,omitempty"`
	Suggestions   []string       `json:"suggestions,omitempty"`
}

// ErrorStatistics 错误统计信息
type ErrorStatistics struct {
	TotalErrors    int64            `json:"total_errors"`
	ErrorsByStatus map[int]int64    `json:"errors_by_status"`
	ErrorsByPath   map[string]int64 `json:"errors_by_path"`
	LastErrors     []ErrorRecord    `json:"last_errors"`
	StartTime      time.Time        `json:"start_time"`
	mu             sync.RWMutex
}

// ErrorRecord 错误记录
type ErrorRecord struct {
	StatusCode int       `json:"status_code"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	UserAgent  string    `json:"user_agent,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
}

// ErrorHook 错误处理钩子函数类型
type ErrorHook func(ctx *mvccontext.Context, statusCode int, err error) error

// I18nMessage 国际化消息结构
type I18nMessage struct {
	ZhCN string `json:"zh_cn"`
	EnUS string `json:"en_us"`
}

// ErrorControllerConfig 错误控制器配置结构
type ErrorControllerConfig struct {
	ShowDetailedError bool   `json:"show_detailed_error"` // 是否显示详细错误信息
	Language          string `json:"language"`            // 语言设置
	CustomTitle       string `json:"custom_title"`        // 自定义页面标题
	SupportEmail      string `json:"support_email"`       // 支持邮箱
	SupportPhone      string `json:"support_phone"`       // 支持电话
	EnableDebugInfo   bool   `json:"enable_debug_info"`   // 是否启用调试信息
}

// BusinessError 业务错误结构
type BusinessError struct {
	Code    string `json:"code"`    // 业务错误码
	Message string `json:"message"` // 错误消息
	Data    any    `json:"data"`    // 附加数据
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewErrorStatistics 创建错误统计实例
func NewErrorStatistics() *ErrorStatistics {
	return &ErrorStatistics{
		TotalErrors:    0,
		ErrorsByStatus: make(map[int]int64),
		ErrorsByPath:   make(map[string]int64),
		LastErrors:     make([]ErrorRecord, 0),
		StartTime:      time.Now(),
	}
}