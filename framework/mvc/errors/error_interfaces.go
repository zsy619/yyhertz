package errors

import (
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 核心数据结构 =============

// StatusConfig 状态码配置
type StatusConfig struct {
	// 基础信息
	StatusCode int    `json:"status_code"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Icon       string `json:"icon"`
	
	// 建议和指导
	Suggestions    []string `json:"suggestions"`
	Recovery       []string `json:"recovery"`
	Prevention     []string `json:"prevention"`
	
	// 行为配置
	Retryable      bool `json:"retryable"`
	ShowDetails    bool `json:"show_details"`
	LogLevel       string `json:"log_level"`
	
	// 模板配置
	Template       string `json:"template"`
	CustomCSS      string `json:"custom_css"`
}

// 注意：ErrorControllerConfig已在default_controller.go中定义，这里不重复定义

// ============= 核心接口定义 =============
// 注意：ErrorHandler接口已在handler.go中定义，这里不重复定义

// StatusHandler 状态码处理器接口
type StatusHandler interface {
	// HandleStatus 处理特定状态码
	HandleStatus(ctx *mvccontext.Context, statusCode int, err error) error
	
	// GetStatusCodes 返回支持的状态码列表
	GetStatusCodes() []int
}

// ErrorRenderer 错误渲染器接口
type ErrorRenderer interface {
	// Render 渲染错误响应
	Render(ctx *mvccontext.Context, errorCtx *ErrorContext) error
	
	// CanRender 检查是否能渲染该请求类型
	CanRender(ctx *mvccontext.Context) bool
	
	// ContentType 返回内容类型
	ContentType() string
}

// TemplateManager 模板管理器接口
type TemplateManager interface {
	// RenderTemplate 渲染模板
	RenderTemplate(name string, data interface{}) (string, error)
	
	// LoadTemplate 加载模板
	LoadTemplate(name string, content string) error
	
	// ReloadTemplates 重新加载所有模板
	ReloadTemplates() error
}

// I18nManager 国际化管理器接口
type I18nManager interface {
	// GetMessage 获取本地化消息
	GetMessage(key string, language string) string
	
	// GetMessages 获取所有消息
	GetMessages(language string) map[string]string
	
	// LoadMessages 加载消息
	LoadMessages(language string, messages map[string]string) error
}

// StatisticsManager 统计管理器接口
type StatisticsManager interface {
	// RecordError 记录错误
	RecordError(statusCode int, path string, method string, err error)
	
	// GetStatistics 获取统计信息
	GetStatistics() *ErrorStatistics
	
	// Reset 重置统计
	Reset()
	
	// GetErrorRate 获取错误率
	GetErrorRate(timeWindow int) float64
}

// 注意：ErrorHook类型已在default_controller.go中定义，这里不重复定义

// HookManager 钩子管理器接口
type HookManager interface {
	// AddHook 添加钩子
	AddHook(phase string, hook ErrorHook)
	
	// RemoveHooks 移除钩子
	RemoveHooks(phase string)
	
	// ExecuteHooks 执行钩子
	ExecuteHooks(phase string, ctx *mvccontext.Context, statusCode int, err error) error
}

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// GetStatusConfig 获取状态码配置
	GetStatusConfig(statusCode int) *StatusConfig
	
	// SetStatusConfig 设置状态码配置
	SetStatusConfig(statusCode int, config *StatusConfig)
	
	// LoadConfig 加载配置
	LoadConfig(configData []byte) error
	
	// GetRetryableErrors 获取可重试错误列表
	GetRetryableErrors() map[int]bool
}

// ============= 工厂接口 =============

// ErrorControllerFactory 错误控制器工厂接口
type ErrorControllerFactory interface {
	// CreateController 创建错误控制器
	CreateController(env string) ErrorHandler
	
	// CreateWithConfig 使用配置创建控制器
	CreateWithConfig(config *ErrorControllerConfig) ErrorHandler
}

// RendererFactory 渲染器工厂接口
type RendererFactory interface {
	// CreateRenderer 创建渲染器
	CreateRenderer(contentType string) ErrorRenderer
	
	// GetAvailableRenderers 获取可用渲染器列表
	GetAvailableRenderers() []string
}