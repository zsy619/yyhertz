package errors

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ============= 配置管理器实现 =============

// DefaultConfigManager 默认配置管理器
type DefaultConfigManager struct {
	statusConfigs   map[int]*StatusConfig
	retryableErrors map[int]bool
	i18nMessages    map[string]map[string]string // language -> key -> message
	mu              sync.RWMutex
}

// NewDefaultConfigManager 创建默认配置管理器
func NewDefaultConfigManager() *DefaultConfigManager {
	manager := &DefaultConfigManager{
		statusConfigs:   make(map[int]*StatusConfig, 32),
		retryableErrors: make(map[int]bool, 16),
		i18nMessages:    make(map[string]map[string]string, 4),
		mu:              sync.RWMutex{},
	}
	
	// 初始化默认配置
	manager.initDefaultConfigs()
	manager.initDefaultMessages()
	
	return manager
}

// GetStatusConfig 获取状态码配置
func (m *DefaultConfigManager) GetStatusConfig(statusCode int) *StatusConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if config, exists := m.statusConfigs[statusCode]; exists {
		return config
	}
	
	// 返回默认配置
	return m.getDefaultConfig(statusCode)
}

// SetStatusConfig 设置状态码配置
func (m *DefaultConfigManager) SetStatusConfig(statusCode int, config *StatusConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.statusConfigs[statusCode] = config
}

// LoadConfig 从JSON加载配置
func (m *DefaultConfigManager) LoadConfig(configData []byte) error {
	var config struct {
		StatusConfigs   map[string]*StatusConfig `json:"status_configs"`
		RetryableErrors map[string]bool          `json:"retryable_errors"`
		I18nMessages    map[string]map[string]string `json:"i18n_messages"`
	}
	
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 加载状态码配置
	if config.StatusConfigs != nil {
		for statusStr, statusConfig := range config.StatusConfigs {
			var statusCode int
			if _, err := fmt.Sscanf(statusStr, "%d", &statusCode); err == nil {
				m.statusConfigs[statusCode] = statusConfig
			}
		}
	}
	
	// 加载可重试错误配置
	if config.RetryableErrors != nil {
		for statusStr, retryable := range config.RetryableErrors {
			var statusCode int
			if _, err := fmt.Sscanf(statusStr, "%d", &statusCode); err == nil {
				m.retryableErrors[statusCode] = retryable
			}
		}
	}
	
	// 加载国际化消息
	if config.I18nMessages != nil {
		for lang, messages := range config.I18nMessages {
			m.i18nMessages[lang] = messages
		}
	}
	
	return nil
}

// GetRetryableErrors 获取可重试错误列表
func (m *DefaultConfigManager) GetRetryableErrors() map[int]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[int]bool, len(m.retryableErrors))
	for k, v := range m.retryableErrors {
		result[k] = v
	}
	return result
}

// IsRetryable 检查错误是否可重试
func (m *DefaultConfigManager) IsRetryable(statusCode int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.retryableErrors[statusCode]
}

// SetRetryable 设置错误是否可重试
func (m *DefaultConfigManager) SetRetryable(statusCode int, retryable bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.retryableErrors[statusCode] = retryable
}

// GetMessage 获取本地化消息
func (m *DefaultConfigManager) GetMessage(key string, language string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if langMessages, exists := m.i18nMessages[language]; exists {
		if message, exists := langMessages[key]; exists {
			return message
		}
	}
	
	// 回退到中文
	if language != "zh-CN" {
		if langMessages, exists := m.i18nMessages["zh-CN"]; exists {
			if message, exists := langMessages[key]; exists {
				return message
			}
		}
	}
	
	return key // 最后回退到key本身
}

// SetMessage 设置本地化消息
func (m *DefaultConfigManager) SetMessage(language, key, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.i18nMessages[language] == nil {
		m.i18nMessages[language] = make(map[string]string)
	}
	m.i18nMessages[language][key] = message
}

// ============= 默认配置初始化 =============

// initDefaultConfigs 初始化默认状态码配置
func (m *DefaultConfigManager) initDefaultConfigs() {
	// 4xx 客户端错误
	m.statusConfigs[400] = &StatusConfig{
		StatusCode:  400,
		Title:       "错误请求",
		Message:     "请求参数错误或格式不正确",
		Icon:        "❓",
		Suggestions: []string{"检查请求参数是否正确", "确认数据格式符合API要求", "查看API文档了解正确的请求格式"},
		Recovery:    []string{"1. 检查请求参数格式和类型", "2. 验证必需参数是否缺失", "3. 确认请求体格式正确（JSON/XML等）"},
		Prevention:  []string{"💡 使用API文档验证请求格式", "💡 实施客户端数据验证", "💡 使用类型安全的API客户端"},
		Retryable:   false,
		ShowDetails: true,
		LogLevel:    "warn",
	}
	
	m.statusConfigs[401] = &StatusConfig{
		StatusCode:  401,
		Title:       "未授权访问",
		Message:     "您需要登录才能访问此资源",
		Icon:        "🔐",
		Suggestions: []string{"请先登录您的账户", "检查您的登录凭证是否过期", "联系管理员确认您的访问权限"},
		Recovery:    []string{"1. 获取有效的访问令牌或登录凭证", "2. 检查令牌是否已过期", "3. 确认权限范围是否足够"},
		Prevention:  []string{"💡 实施令牌自动刷新机制", "💡 监控令牌过期时间", "💡 使用安全的认证流程"},
		Retryable:   false,
		ShowDetails: true,
		LogLevel:    "warn",
	}
	
	m.statusConfigs[403] = &StatusConfig{
		StatusCode:  403,
		Title:       "禁止访问",
		Message:     "您没有权限访问此资源",
		Icon:        "🚫",
		Suggestions: []string{"联系管理员申请相关权限", "确认您的账户状态是否正常", "检查是否访问了受限制的资源"},
		Recovery:    []string{"1. 联系管理员申请必要权限", "2. 检查用户角色和权限设置", "3. 确认访问的资源确实需要当前权限"},
		Prevention:  []string{"💡 实施基于角色的访问控制", "💡 定期审查用户权限", "💡 使用最小权限原则"},
		Retryable:   false,
		ShowDetails: true,
		LogLevel:    "warn",
	}
	
	m.statusConfigs[404] = &StatusConfig{
		StatusCode:  404,
		Title:       "页面未找到",
		Message:     "请检查URL地址是否正确",
		Icon:        "🔍",
		Suggestions: []string{"检查URL拼写是否正确", "尝试返回首页重新导航", "清除浏览器缓存后重试"},
		Recovery:    []string{"1. 验证URL路径拼写是否正确", "2. 检查资源是否已被移动或删除", "3. 确认API版本是否正确"},
		Prevention:  []string{"💡 使用静态分析检查链接", "💡 实施链接检查工具", "💡 建立资源重定向机制"},
		Retryable:   false,
		ShowDetails: true,
		LogLevel:    "info",
	}
	
	m.statusConfigs[429] = &StatusConfig{
		StatusCode:  429,
		Title:       "请求过多",
		Message:     "请求过于频繁，请稍后重试",
		Icon:        "🚦",
		Suggestions: []string{"请降低请求频率", "等待一段时间后重试", "考虑使用缓存减少重复请求"},
		Recovery:    []string{"1. 实施指数退避重试策略", "2. 减少并发请求数量", "3. 考虑请求缓存机制"},
		Prevention:  []string{"💡 实施客户端限流机制", "💡 使用请求队列管理", "💡 监控API使用模式"},
		Retryable:   true,
		ShowDetails: true,
		LogLevel:    "warn",
	}
	
	// 5xx 服务器错误
	m.statusConfigs[500] = &StatusConfig{
		StatusCode:  500,
		Title:       "服务器内部错误",
		Message:     "服务器遇到了意外情况",
		Icon:        "⚠️",
		Suggestions: []string{"请稍后重试", "如果问题持续存在，请联系技术支持", "您也可以尝试刷新页面"},
		Recovery:    []string{"1. 稍后重试请求", "2. 检查请求是否触发了服务器bug", "3. 联系技术支持并提供错误上下文"},
		Prevention:  []string{"💡 增加服务器监控和告警", "💡 实施断路器模式", "💡 建立容错和降级机制"},
		Retryable:   true,
		ShowDetails: false,
		LogLevel:    "error",
	}
	
	m.statusConfigs[502] = &StatusConfig{
		StatusCode:  502,
		Title:       "网关错误",
		Message:     "网关错误",
		Icon:        "🌐",
		Suggestions: []string{"请稍后重试", "检查网络连接是否正常", "如果问题持续存在，请联系技术支持"},
		Retryable:   true,
		ShowDetails: false,
		LogLevel:    "error",
	}
	
	m.statusConfigs[503] = &StatusConfig{
		StatusCode:  503,
		Title:       "服务不可用",
		Message:     "服务暂时不可用",
		Icon:        "🔧",
		Suggestions: []string{"服务正在维护中，请稍后重试", "关注官方公告了解维护时间", "使用其他可用的服务入口"},
		Retryable:   true,
		ShowDetails: false,
		LogLevel:    "warn",
	}
}

// initDefaultMessages 初始化默认国际化消息
func (m *DefaultConfigManager) initDefaultMessages() {
	// 中文消息
	m.i18nMessages["zh-CN"] = map[string]string{
		"error.title":              "错误页面",
		"error.back":               "返回上页",
		"error.home":               "返回首页",
		"error.retry":              "重试",
		"error.suggestions":        "解决建议",
		"error.recovery":           "恢复指令",
		"error.prevention":         "预防措施",
		"error.contact.email":      "邮箱支持",
		"error.contact.phone":      "电话支持",
		"error.debug.info":         "调试信息",
		"error.request.path":       "请求路径",
		"error.request.method":     "请求方法",
		"error.timestamp":          "时间",
		"framework.name":           "YYHertz Framework",
	}
	
	// 英文消息
	m.i18nMessages["en-US"] = map[string]string{
		"error.title":              "Error Page",
		"error.back":               "Go Back",
		"error.home":               "Home",
		"error.retry":              "Retry",
		"error.suggestions":        "Suggestions",
		"error.recovery":           "Recovery Instructions",
		"error.prevention":         "Prevention Tips",
		"error.contact.email":      "Email Support",
		"error.contact.phone":      "Phone Support",
		"error.debug.info":         "Debug Information",
		"error.request.path":       "Request Path",
		"error.request.method":     "Request Method",
		"error.timestamp":          "Timestamp",
		"framework.name":           "YYHertz Framework",
	}
}

// getDefaultConfig 获取默认配置
func (m *DefaultConfigManager) getDefaultConfig(statusCode int) *StatusConfig {
	title := fmt.Sprintf("HTTP %d", statusCode)
	message := "An error occurred"
	icon := "❌"
	
	// 根据状态码类别设置默认值
	switch {
	case statusCode >= 400 && statusCode < 500:
		message = "客户端请求错误"
		icon = "⚠️"
	case statusCode >= 500:
		message = "服务器内部错误"
		icon = "🔥"
	}
	
	return &StatusConfig{
		StatusCode:  statusCode,
		Title:       title,
		Message:     message,
		Icon:        icon,
		Suggestions: []string{"请稍后重试", "如果问题持续存在，请联系技术支持"},
		Recovery:    []string{"1. 查看相关文档了解错误原因", "2. 检查网络连接状态", "3. 如问题持续请联系技术支持"},
		Prevention:  []string{"💡 监控应用性能和错误率", "💡 建立完善的日志记录", "💡 定期进行系统健康检查"},
		Retryable:   statusCode >= 500, // 5xx错误默认可重试
		ShowDetails: statusCode < 500,  // 4xx错误显示详情
		LogLevel:    "error",
	}
}

// ============= 全局配置管理器实例 =============

var globalConfigManager = NewDefaultConfigManager()

// GetGlobalConfigManager 获取全局配置管理器
func GetGlobalConfigManager() ConfigManager {
	return globalConfigManager
}

// GetStatusConfig 获取状态码配置（全局方法）
func GetStatusConfig(statusCode int) *StatusConfig {
	return globalConfigManager.GetStatusConfig(statusCode)
}

// IsRetryableError 检查错误是否可重试（全局方法）
func IsRetryableError(statusCode int) bool {
	return globalConfigManager.IsRetryable(statusCode)
}

// GetLocalizedMessage 获取本地化消息（全局方法）
func GetLocalizedMessage(key string, language string) string {
	return globalConfigManager.GetMessage(key, language)
}