package errors

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ============= 错误处理统一配置 =============

// ErrorConfig 统一错误处理配置
// 整合了所有错误处理子系统的配置项
type ErrorConfig struct {
	// ============= 基础配置 =============
	
	// ShowDetailedError 是否显示详细错误信息（开发环境true，生产环境false）
	ShowDetailedError bool `json:"show_detailed_error" yaml:"show_detailed_error"`

	// LogErrors 是否记录错误日志
	LogErrors bool `json:"log_errors" yaml:"log_errors"`

	// ErrorLogLevel 错误日志级别
	ErrorLogLevel string `json:"error_log_level" yaml:"error_log_level"`

	// EnableRecovery 是否启用panic恢复
	EnableRecovery bool `json:"enable_recovery" yaml:"enable_recovery"`

	// ============= 处理器分发器配置 =============
	
	// Dispatcher 错误分发器配置
	Dispatcher DispatcherConfig `json:"dispatcher" yaml:"dispatcher"`

	// ============= 注册器配置 =============
	
	// Registry 错误注册器配置
	Registry RegistryConfig `json:"registry" yaml:"registry"`

	// ============= 智能分类器配置 =============
	
	// Classifier 智能分类器配置
	Classifier ClassifierConfig `json:"classifier" yaml:"classifier"`

	// ============= 自动恢复系统配置 =============
	
	// Recovery 自动恢复配置
	Recovery RecoveryConfig `json:"recovery" yaml:"recovery"`

	// ============= 模板和页面配置 =============
	
	// DefaultErrorPage 默认错误页面模板路径
	DefaultErrorPage string `json:"default_error_page" yaml:"default_error_page"`

	// CustomErrorPages 自定义状态码错误页面映射
	CustomErrorPages map[int]string `json:"custom_error_pages" yaml:"custom_error_pages"`

	// TemplateEngine 模板引擎类型 ("html", "pongo2", "jet")
	TemplateEngine string `json:"template_engine" yaml:"template_engine"`

	// ============= 响应和内容协商配置 =============
	
	// ContentNegotiation 是否启用内容协商
	ContentNegotiation bool `json:"content_negotiation" yaml:"content_negotiation"`

	// DefaultResponseFormat 默认响应格式 ("json", "xml", "html")
	DefaultResponseFormat string `json:"default_response_format" yaml:"default_response_format"`

	// SupportedFormats 支持的响应格式列表
	SupportedFormats []string `json:"supported_formats" yaml:"supported_formats"`

	// ============= 监控和指标配置 =============
	
	// EnableErrorMetrics 是否启用错误监控指标
	EnableErrorMetrics bool `json:"enable_error_metrics" yaml:"enable_error_metrics"`

	// MetricsCollectionInterval 指标收集间隔
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval" yaml:"metrics_collection_interval"`

	// EnableTracing 是否启用链路跟踪
	EnableTracing bool `json:"enable_tracing" yaml:"enable_tracing"`

	// ============= 性能配置 =============
	
	// MaxStackTraceSize 最大堆栈跟踪大小
	MaxStackTraceSize int `json:"max_stack_trace_size" yaml:"max_stack_trace_size"`

	// HandlerTimeout 单个错误处理器的超时时间
	HandlerTimeout time.Duration `json:"handler_timeout" yaml:"handler_timeout"`

	// MaxConcurrentHandling 最大并发错误处理数
	MaxConcurrentHandling int `json:"max_concurrent_handling" yaml:"max_concurrent_handling"`

	// ============= 安全配置 =============
	
	// SanitizeErrorMessages 是否清理错误信息中的敏感数据
	SanitizeErrorMessages bool `json:"sanitize_error_messages" yaml:"sanitize_error_messages"`

	// AllowedErrorFields 允许暴露的错误字段列表
	AllowedErrorFields []string `json:"allowed_error_fields" yaml:"allowed_error_fields"`

	// MaskSensitiveData 是否遮蔽敏感数据
	MaskSensitiveData bool `json:"mask_sensitive_data" yaml:"mask_sensitive_data"`

	// ============= 开发和调试配置 =============
	
	// DebugMode 是否启用调试模式
	DebugMode bool `json:"debug_mode" yaml:"debug_mode"`

	// EnableProfiler 是否启用性能分析
	EnableProfiler bool `json:"enable_profiler" yaml:"enable_profiler"`

	// VerboseLogging 是否启用详细日志
	VerboseLogging bool `json:"verbose_logging" yaml:"verbose_logging"`
}

// DefaultErrorConfig 返回默认错误处理配置
func DefaultErrorConfig() *ErrorConfig {
	return &ErrorConfig{
		// 基础配置
		ShowDetailedError: true,
		LogErrors:         true,
		ErrorLogLevel:     "error",
		EnableRecovery:    true,

		// 分发器配置
		Dispatcher: DefaultDispatcherConfig(),

		// 注册器配置
		Registry: *DefaultRegistryConfig(),

		// 分类器配置
		Classifier: DefaultClassifierConfig(),

		// 恢复系统配置
		Recovery: DefaultRecoveryConfig(),

		// 模板和页面配置
		DefaultErrorPage:   "",
		CustomErrorPages:   make(map[int]string),
		TemplateEngine:     "html",

		// 响应和内容协商配置
		ContentNegotiation:    true,
		DefaultResponseFormat: "json",
		SupportedFormats:      []string{"json", "xml", "html"},

		// 监控和指标配置
		EnableErrorMetrics:        true,
		MetricsCollectionInterval: time.Minute,
		EnableTracing:             false,

		// 性能配置
		MaxStackTraceSize:     4096,
		HandlerTimeout:        30 * time.Second,
		MaxConcurrentHandling: 100,

		// 安全配置
		SanitizeErrorMessages: true,
		AllowedErrorFields:    []string{"code", "message", "timestamp", "path", "method"},
		MaskSensitiveData:     true,

		// 开发和调试配置
		DebugMode:       false,
		EnableProfiler:  false,
		VerboseLogging:  false,
	}
}

// DevelopmentErrorConfig 返回开发环境错误配置
func DevelopmentErrorConfig() *ErrorConfig {
	config := DefaultErrorConfig()
	
	// 开发环境特殊设置
	config.ShowDetailedError = true
	config.DebugMode = true
	config.VerboseLogging = true
	config.EnableProfiler = true
	config.SanitizeErrorMessages = false
	config.MaskSensitiveData = false
	config.AllowedErrorFields = []string{"code", "message", "error", "stack_trace", "timestamp", "path", "method", "params", "headers"}

	// 开发环境允许更详细的错误信息
	config.Dispatcher.EnableStackTrace = true
	config.Dispatcher.EnableStatistics = true
	config.Registry.EnableMetrics = true
	config.Registry.ShowDetailedError = true
	config.Classifier.EnableLearning = true
	config.Recovery.EnableAutoRecovery = true

	return config
}

// ProductionErrorConfig 返回生产环境错误配置
func ProductionErrorConfig() *ErrorConfig {
	config := DefaultErrorConfig()
	
	// 生产环境特殊设置
	config.ShowDetailedError = false
	config.DebugMode = false
	config.VerboseLogging = false
	config.EnableProfiler = false
	config.SanitizeErrorMessages = true
	config.MaskSensitiveData = true
	config.ErrorLogLevel = "warn"

	// 生产环境性能优化
	config.Dispatcher.EnableStackTrace = false
	config.Dispatcher.EnableStatistics = true
	config.Registry.EnableMetrics = true
	config.Registry.ShowDetailedError = false
	config.Classifier.EnableLearning = false  // 生产环境关闭学习功能
	config.Recovery.EnableAutoRecovery = true

	// 更严格的超时设置
	config.HandlerTimeout = 10 * time.Second
	config.Dispatcher.RetryInterval = 500 * time.Millisecond

	return config
}

// TestErrorConfig 返回测试环境错误配置
func TestErrorConfig() *ErrorConfig {
	config := DefaultErrorConfig()
	
	// 测试环境特殊设置
	config.ShowDetailedError = true
	config.DebugMode = true
	config.LogErrors = false  // 测试时通常不需要日志
	config.EnableErrorMetrics = false  // 测试时关闭指标收集
	config.VerboseLogging = false

	// 测试环境快速失败
	config.HandlerTimeout = 5 * time.Second
	config.Dispatcher.MaxRetries = 1
	config.Dispatcher.RetryInterval = 100 * time.Millisecond
	config.Recovery.EnableAutoRecovery = false  // 测试时关闭自动恢复

	return config
}

// ============= 配置验证和管理 =============

// Validate 验证配置的有效性
func (c *ErrorConfig) Validate() error {
	// 验证基础配置
	if c.ErrorLogLevel != "" {
		validLevels := map[string]bool{
			"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
		}
		if !validLevels[c.ErrorLogLevel] {
			return fmt.Errorf("invalid error log level: %s", c.ErrorLogLevel)
		}
	}

	// 验证模板引擎
	if c.TemplateEngine != "" {
		validEngines := map[string]bool{
			"html": true, "pongo2": true, "jet": true,
		}
		if !validEngines[c.TemplateEngine] {
			return fmt.Errorf("invalid template engine: %s", c.TemplateEngine)
		}
	}

	// 验证响应格式
	if c.DefaultResponseFormat != "" {
		validFormats := map[string]bool{
			"json": true, "xml": true, "html": true, "yaml": true, "text": true,
		}
		if !validFormats[c.DefaultResponseFormat] {
			return fmt.Errorf("invalid default response format: %s", c.DefaultResponseFormat)
		}
	}

	// 验证支持的格式列表
	for _, format := range c.SupportedFormats {
		validFormats := map[string]bool{
			"json": true, "xml": true, "html": true, "yaml": true, "text": true,
		}
		if !validFormats[format] {
			return fmt.Errorf("invalid supported format: %s", format)
		}
	}

	// 验证超时配置
	if c.HandlerTimeout <= 0 {
		return fmt.Errorf("handler timeout must be positive, got: %v", c.HandlerTimeout)
	}

	if c.MetricsCollectionInterval <= 0 {
		return fmt.Errorf("metrics collection interval must be positive, got: %v", c.MetricsCollectionInterval)
	}

	// 验证性能配置
	if c.MaxStackTraceSize <= 0 {
		return fmt.Errorf("max stack trace size must be positive, got: %d", c.MaxStackTraceSize)
	}

	if c.MaxConcurrentHandling <= 0 {
		return fmt.Errorf("max concurrent handling must be positive, got: %d", c.MaxConcurrentHandling)
	}

	// 验证子配置
	if err := c.Registry.Validate(); err != nil {
		return fmt.Errorf("registry config validation failed: %w", err)
	}

	if err := c.Dispatcher.Validate(); err != nil {
		return fmt.Errorf("dispatcher config validation failed: %w", err)
	}

	return nil
}

// Validate 验证注册器配置
func (c *RegistryConfig) Validate() error {
	if c.FallbackChainLength <= 0 {
		return fmt.Errorf("fallback chain length must be positive, got: %d", c.FallbackChainLength)
	}

	if c.HandlerTimeout <= 0 {
		return fmt.Errorf("handler timeout must be positive, got: %v", c.HandlerTimeout)
	}

	return nil
}

// Validate 验证分发器配置
func (c *DispatcherConfig) Validate() error {
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative, got: %d", c.MaxRetries)
	}

	if c.RetryInterval < 0 {
		return fmt.Errorf("retry interval must be non-negative, got: %v", c.RetryInterval)
	}

	if c.CircuitBreakerThreshold <= 0 {
		return fmt.Errorf("circuit breaker threshold must be positive, got: %d", c.CircuitBreakerThreshold)
	}

	return nil
}

// Clone 深拷贝配置
func (c *ErrorConfig) Clone() *ErrorConfig {
	// 使用JSON序列化/反序列化进行深拷贝
	data, _ := json.Marshal(c)
	var cloned ErrorConfig
	json.Unmarshal(data, &cloned)
	return &cloned
}

// Merge 合并配置（other的非零值会覆盖当前配置）
func (c *ErrorConfig) Merge(other *ErrorConfig) *ErrorConfig {
	if other == nil {
		return c
	}

	merged := c.Clone()

	// 合并基础配置
	if other.ShowDetailedError {
		merged.ShowDetailedError = other.ShowDetailedError
	}
	if other.LogErrors {
		merged.LogErrors = other.LogErrors
	}
	if other.ErrorLogLevel != "" {
		merged.ErrorLogLevel = other.ErrorLogLevel
	}
	if other.EnableRecovery {
		merged.EnableRecovery = other.EnableRecovery
	}

	// 合并模板配置
	if other.DefaultErrorPage != "" {
		merged.DefaultErrorPage = other.DefaultErrorPage
	}
	if len(other.CustomErrorPages) > 0 {
		if merged.CustomErrorPages == nil {
			merged.CustomErrorPages = make(map[int]string)
		}
		for k, v := range other.CustomErrorPages {
			merged.CustomErrorPages[k] = v
		}
	}
	if other.TemplateEngine != "" {
		merged.TemplateEngine = other.TemplateEngine
	}

	// 合并响应配置
	if other.ContentNegotiation {
		merged.ContentNegotiation = other.ContentNegotiation
	}
	if other.DefaultResponseFormat != "" {
		merged.DefaultResponseFormat = other.DefaultResponseFormat
	}
	if len(other.SupportedFormats) > 0 {
		merged.SupportedFormats = append([]string{}, other.SupportedFormats...)
	}

	// 合并性能配置
	if other.HandlerTimeout > 0 {
		merged.HandlerTimeout = other.HandlerTimeout
	}
	if other.MaxConcurrentHandling > 0 {
		merged.MaxConcurrentHandling = other.MaxConcurrentHandling
	}
	if other.MaxStackTraceSize > 0 {
		merged.MaxStackTraceSize = other.MaxStackTraceSize
	}

	// 合并安全配置
	if other.SanitizeErrorMessages {
		merged.SanitizeErrorMessages = other.SanitizeErrorMessages
	}
	if other.MaskSensitiveData {
		merged.MaskSensitiveData = other.MaskSensitiveData
	}
	if len(other.AllowedErrorFields) > 0 {
		merged.AllowedErrorFields = append([]string{}, other.AllowedErrorFields...)
	}

	return merged
}

// ToMap 将配置转换为map格式（用于序列化）
func (c *ErrorConfig) ToMap() map[string]any {
	result := make(map[string]any)

	// 基础配置
	result["show_detailed_error"] = c.ShowDetailedError
	result["log_errors"] = c.LogErrors
	result["error_log_level"] = c.ErrorLogLevel
	result["enable_recovery"] = c.EnableRecovery

	// 模板配置
	result["default_error_page"] = c.DefaultErrorPage
	result["custom_error_pages"] = c.CustomErrorPages
	result["template_engine"] = c.TemplateEngine

	// 响应配置
	result["content_negotiation"] = c.ContentNegotiation
	result["default_response_format"] = c.DefaultResponseFormat
	result["supported_formats"] = c.SupportedFormats

	// 监控配置
	result["enable_error_metrics"] = c.EnableErrorMetrics
	result["metrics_collection_interval"] = c.MetricsCollectionInterval.String()
	result["enable_tracing"] = c.EnableTracing

	// 性能配置
	result["max_stack_trace_size"] = c.MaxStackTraceSize
	result["handler_timeout"] = c.HandlerTimeout.String()
	result["max_concurrent_handling"] = c.MaxConcurrentHandling

	// 安全配置
	result["sanitize_error_messages"] = c.SanitizeErrorMessages
	result["allowed_error_fields"] = c.AllowedErrorFields
	result["mask_sensitive_data"] = c.MaskSensitiveData

	// 调试配置
	result["debug_mode"] = c.DebugMode
	result["enable_profiler"] = c.EnableProfiler
	result["verbose_logging"] = c.VerboseLogging

	return result
}

// ============= 环境配置工厂 =============

// ConfigurationFactory 配置工厂
type ConfigurationFactory struct {
	environment string
	custom      *ErrorConfig
}

// NewConfigurationFactory 创建配置工厂
func NewConfigurationFactory(environment string) *ConfigurationFactory {
	return &ConfigurationFactory{
		environment: environment,
	}
}

// SetCustomConfig 设置自定义配置
func (f *ConfigurationFactory) SetCustomConfig(config *ErrorConfig) *ConfigurationFactory {
	f.custom = config
	return f
}

// Build 构建最终配置
func (f *ConfigurationFactory) Build() (*ErrorConfig, error) {
	var baseConfig *ErrorConfig

	// 根据环境选择基础配置
	switch f.environment {
	case "development", "dev":
		baseConfig = DevelopmentErrorConfig()
	case "production", "prod":
		baseConfig = ProductionErrorConfig()
	case "test", "testing":
		baseConfig = TestErrorConfig()
	default:
		baseConfig = DefaultErrorConfig()
	}

	// 合并自定义配置
	if f.custom != nil {
		baseConfig = baseConfig.Merge(f.custom)
	}

	// 验证最终配置
	if err := baseConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return baseConfig, nil
}

// ============= 全局配置管理 =============

var (
	globalErrorConfig *ErrorConfig
	configMutex       sync.RWMutex
)

// SetGlobalErrorConfig 设置全局错误配置
func SetGlobalErrorConfig(config *ErrorConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	configMutex.Lock()
	defer configMutex.Unlock()
	globalErrorConfig = config
	return nil
}

// GetGlobalErrorConfig 获取全局错误配置
func GetGlobalErrorConfig() *ErrorConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalErrorConfig == nil {
		// 如果没有设置全局配置，返回默认配置
		return DefaultErrorConfig()
	}

	return globalErrorConfig.Clone()
}

// UpdateGlobalErrorConfig 更新全局错误配置
func UpdateGlobalErrorConfig(updater func(*ErrorConfig) *ErrorConfig) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	if globalErrorConfig == nil {
		globalErrorConfig = DefaultErrorConfig()
	}

	updated := updater(globalErrorConfig.Clone())
	if err := updated.Validate(); err != nil {
		return err
	}

	globalErrorConfig = updated
	return nil
}

// ResetGlobalErrorConfig 重置全局错误配置为默认值
func ResetGlobalErrorConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalErrorConfig = DefaultErrorConfig()
}

// 初始化默认全局配置
func init() {
	globalErrorConfig = DefaultErrorConfig()
}