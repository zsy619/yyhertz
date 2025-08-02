package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
)

// MessageManager 消息管理器
type MessageManager struct {
	messages map[string]map[string]string // locale -> tag -> message
	mutex    sync.RWMutex
	fallback string // 默认语言
}

// NewMessageManager 创建消息管理器
func NewMessageManager(fallback string) *MessageManager {
	mm := &MessageManager{
		messages: make(map[string]map[string]string),
		fallback: fallback,
	}
	
	// 加载内置消息
	mm.loadBuiltinMessages()
	
	return mm
}

// loadBuiltinMessages 加载内置消息
func (mm *MessageManager) loadBuiltinMessages() {
	// 中文消息
	zhCN := map[string]string{
		"required":    "{field}是必填项",
		"min":         "{field}的值不能小于{param}",
		"max":         "{field}的值不能大于{param}",
		"range":       "{field}的值必须在{param}范围内",
		"minLength":   "{field}的长度不能少于{param}个字符",
		"maxLength":   "{field}的长度不能超过{param}个字符",
		"length":      "{field}的长度必须为{param}个字符",
		"email":       "{field}必须是有效的邮箱地址",
		"url":         "{field}必须是有效的URL地址",
		"alpha":       "{field}只能包含字母",
		"alphaNum":    "{field}只能包含字母和数字",
		"numeric":     "{field}必须是数字",
		"integer":     "{field}必须是整数",
		"decimal":     "{field}必须是小数",
		"ip":          "{field}必须是有效的IP地址",
		"ipv4":        "{field}必须是有效的IPv4地址",
		"ipv6":        "{field}必须是有效的IPv6地址",
		"mac":         "{field}必须是有效的MAC地址",
		"uuid":        "{field}必须是有效的UUID",
		"date":        "{field}必须是有效的日期",
		"datetime":    "{field}必须是有效的日期时间",
		"time":        "{field}必须是有效的时间",
		"before":      "{field}必须早于{param}",
		"after":       "{field}必须晚于{param}",
		"regex":       "{field}格式不正确",
		"contains":    "{field}必须包含{param}",
		"startsWith":  "{field}必须以{param}开头",
		"endsWith":    "{field}必须以{param}结尾",
		"in":          "{field}的值必须是{param}中的一个",
		"notIn":       "{field}的值不能是{param}中的一个",
		"positive":    "{field}必须是正数",
		"negative":    "{field}必须是负数",
		"nonZero":     "{field}不能为零",
		"mobile":      "{field}必须是有效的手机号码",
		"phone":       "{field}必须是有效的固定电话号码",
		"idCard":      "{field}必须是有效的身份证号码",
		"zipCode":     "{field}必须是有效的邮政编码",
		"card":        "{field}必须是有效的银行卡号",
		"json":        "{field}必须是有效的JSON格式",
		"base64":      "{field}必须是有效的Base64格式",
		"hexColor":    "{field}必须是有效的十六进制颜色值",
	}
	
	// 英文消息
	enUS := map[string]string{
		"required":    "{field} is required",
		"min":         "{field} must be at least {param}",
		"max":         "{field} must be at most {param}",
		"range":       "{field} must be between {param}",
		"minLength":   "{field} must be at least {param} characters long",
		"maxLength":   "{field} must be at most {param} characters long",
		"length":      "{field} must be exactly {param} characters long",
		"email":       "{field} must be a valid email address",
		"url":         "{field} must be a valid URL",
		"alpha":       "{field} must contain only letters",
		"alphaNum":    "{field} must contain only letters and numbers",
		"numeric":     "{field} must be a number",
		"integer":     "{field} must be an integer",
		"decimal":     "{field} must be a decimal number",
		"ip":          "{field} must be a valid IP address",
		"ipv4":        "{field} must be a valid IPv4 address",
		"ipv6":        "{field} must be a valid IPv6 address",
		"mac":         "{field} must be a valid MAC address",
		"uuid":        "{field} must be a valid UUID",
		"date":        "{field} must be a valid date",
		"datetime":    "{field} must be a valid datetime",
		"time":        "{field} must be a valid time",
		"before":      "{field} must be before {param}",
		"after":       "{field} must be after {param}",
		"regex":       "{field} format is invalid",
		"contains":    "{field} must contain {param}",
		"startsWith":  "{field} must start with {param}",
		"endsWith":    "{field} must end with {param}",
		"in":          "{field} must be one of {param}",
		"notIn":       "{field} must not be one of {param}",
		"positive":    "{field} must be positive",
		"negative":    "{field} must be negative",
		"nonZero":     "{field} must not be zero",
		"mobile":      "{field} must be a valid mobile number",
		"phone":       "{field} must be a valid phone number",
		"idCard":      "{field} must be a valid ID card number",
		"zipCode":     "{field} must be a valid zip code",
		"card":        "{field} must be a valid credit card number",
		"json":        "{field} must be a valid JSON",
		"base64":      "{field} must be a valid Base64 string",
		"hexColor":    "{field} must be a valid hex color",
	}
	
	mm.messages["zh-CN"] = zhCN
	mm.messages["en-US"] = enUS
}

// SetMessages 设置语言消息
func (mm *MessageManager) SetMessages(locale string, messages map[string]string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	if mm.messages[locale] == nil {
		mm.messages[locale] = make(map[string]string)
	}
	
	for tag, message := range messages {
		mm.messages[locale][tag] = message
	}
}

// SetMessage 设置单个消息
func (mm *MessageManager) SetMessage(locale, tag, message string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	
	if mm.messages[locale] == nil {
		mm.messages[locale] = make(map[string]string)
	}
	
	mm.messages[locale][tag] = message
}

// GetMessage 获取消息
func (mm *MessageManager) GetMessage(locale, tag string) string {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	// 尝试获取指定语言的消息
	if messages, exists := mm.messages[locale]; exists {
		if message, exists := messages[tag]; exists {
			return message
		}
	}
	
	// 回退到默认语言
	if locale != mm.fallback {
		if messages, exists := mm.messages[mm.fallback]; exists {
			if message, exists := messages[tag]; exists {
				return message
			}
		}
	}
	
	// 返回默认消息
	return fmt.Sprintf("Validation failed for %s", tag)
}

// LoadFromFile 从文件加载消息
func (mm *MessageManager) LoadFromFile(locale, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read message file: %w", err)
	}
	
	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to parse message file: %w", err)
	}
	
	mm.SetMessages(locale, messages)
	config.Infof("Loaded validation messages for locale: %s", locale)
	
	return nil
}

// LoadFromDirectory 从目录加载所有消息文件
func (mm *MessageManager) LoadFromDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			// 从文件名提取语言代码
			filename := strings.TrimSuffix(info.Name(), ".json")
			locale := filename
			
			// 尝试加载消息
			if err := mm.LoadFromFile(locale, path); err != nil {
				config.Warnf("Failed to load message file %s: %v", path, err)
			}
		}
		
		return nil
	})
}

// SaveToFile 保存消息到文件
func (mm *MessageManager) SaveToFile(locale, filePath string) error {
	mm.mutex.RLock()
	messages, exists := mm.messages[locale]
	mm.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("locale %s not found", locale)
	}
	
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write message file: %w", err)
	}
	
	return nil
}

// GetSupportedLocales 获取支持的语言列表
func (mm *MessageManager) GetSupportedLocales() []string {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	
	locales := make([]string, 0, len(mm.messages))
	for locale := range mm.messages {
		locales = append(locales, locale)
	}
	
	return locales
}

// ============= 验证器消息扩展 =============

// loadDefaultMessages 加载默认错误消息
func (v *Validator) loadDefaultMessages() {
	messageManager := NewMessageManager("zh-CN")
	
	// 获取当前语言的消息
	if messages, exists := messageManager.messages[v.locale]; exists {
		v.mutex.Lock()
		for tag, message := range messages {
			v.messages[tag] = message
		}
		v.mutex.Unlock()
	} else {
		config.Warnf("Messages for locale %s not found, using fallback", v.locale)
		// 使用回退语言
		if messages, exists := messageManager.messages[messageManager.fallback]; exists {
			v.mutex.Lock()
			for tag, message := range messages {
				v.messages[tag] = message
			}
			v.mutex.Unlock()
		}
	}
}

// SetLocale 设置语言
func (v *Validator) SetLocale(locale string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	v.locale = locale
	
	// 重新加载消息
	v.loadDefaultMessages()
	
	config.Debugf("Validator locale changed to: %s", locale)
}

// GetLocale 获取当前语言
func (v *Validator) GetLocale() string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	
	return v.locale
}

// LoadMessages 从文件加载消息
func (v *Validator) LoadMessages(filePath string) error {
	messageManager := NewMessageManager(v.locale)
	
	if err := messageManager.LoadFromFile(v.locale, filePath); err != nil {
		return err
	}
	
	// 更新验证器消息
	if messages, exists := messageManager.messages[v.locale]; exists {
		v.mutex.Lock()
		for tag, message := range messages {
			v.messages[tag] = message
		}
		v.mutex.Unlock()
	}
	
	return nil
}

// ============= 消息模板系统 =============

// MessageTemplate 消息模板
type MessageTemplate struct {
	template string
	params   map[string]string
}

// NewMessageTemplate 创建消息模板
func NewMessageTemplate(template string) *MessageTemplate {
	return &MessageTemplate{
		template: template,
		params:   make(map[string]string),
	}
}

// SetParam 设置参数
func (mt *MessageTemplate) SetParam(key, value string) *MessageTemplate {
	mt.params[key] = value
	return mt
}

// Render 渲染消息
func (mt *MessageTemplate) Render() string {
	message := mt.template
	
	for key, value := range mt.params {
		placeholder := "{" + key + "}"
		message = strings.ReplaceAll(message, placeholder, value)
	}
	
	return message
}

// ============= 自定义消息示例 =============

// ExampleCustomMessages 示例自定义消息
func ExampleCustomMessages() map[string]string {
	return map[string]string{
		// 用户相关
		"username.required":    "用户名不能为空",
		"username.minLength":   "用户名至少需要3个字符",
		"username.maxLength":   "用户名不能超过20个字符",
		"username.alphaNum":    "用户名只能包含字母和数字",
		
		// 密码相关
		"password.required":    "密码不能为空",
		"password.minLength":   "密码至少需要6个字符",
		"password.maxLength":   "密码不能超过100个字符",
		
		// 邮箱相关
		"email.required":       "邮箱地址不能为空",
		"email.email":          "请输入有效的邮箱地址",
		"email.maxLength":      "邮箱地址不能超过255个字符",
		
		// 手机号相关
		"mobile.required":      "手机号码不能为空",
		"mobile.mobile":        "请输入有效的手机号码",
		
		// 年龄相关
		"age.required":         "年龄不能为空",
		"age.integer":          "年龄必须是整数",
		"age.min":              "年龄不能小于18岁",
		"age.max":              "年龄不能大于120岁",
		
		// 金额相关
		"amount.required":      "金额不能为空",
		"amount.decimal":       "金额必须是有效的数字",
		"amount.positive":      "金额必须大于0",
	}
}

// ============= 便捷函数 =============

var globalMessageManager *MessageManager

// GetMessageManager 获取全局消息管理器
func GetMessageManager() *MessageManager {
	if globalMessageManager == nil {
		globalMessageManager = NewMessageManager("zh-CN")
	}
	return globalMessageManager
}

// SetGlobalMessage 设置全局消息
func SetGlobalMessage(locale, tag, message string) {
	GetMessageManager().SetMessage(locale, tag, message)
}

// GetGlobalMessage 获取全局消息
func GetGlobalMessage(locale, tag string) string {
	return GetMessageManager().GetMessage(locale, tag)
}

// LoadGlobalMessages 加载全局消息
func LoadGlobalMessages(locale, filePath string) error {
	return GetMessageManager().LoadFromFile(locale, filePath)
}

// SetValidatorLocale 设置验证器语言
func SetValidatorLocale(locale string) {
	GetDefaultValidator().SetLocale(locale)
}