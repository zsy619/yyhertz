package core

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 数据验证方法 =============

// ValidateRequired 验证必填字段
func (c *BaseController) ValidateRequired(value any, fieldName string) error {
	if value == nil {
		return fmt.Errorf("%s is required", fieldName)
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
	case int, int8, int16, int32, int64:
		// 数字类型不检查空值
	case float32, float64:
		// 浮点类型不检查空值
	default:
		// 使用反射检查其他类型
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr && val.IsNil() {
			return fmt.Errorf("%s is required", fieldName)
		}
	}

	return nil
}

// ValidateLength 验证字符串长度
func (c *BaseController) ValidateLength(value string, min, max int, fieldName string) error {
	length := len(value)

	if min > 0 && length < min {
		return fmt.Errorf("%s must be at least %d characters long", fieldName, min)
	}

	if max > 0 && length > max {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, max)
	}

	return nil
}

// ValidateRange 验证数值范围
func (c *BaseController) ValidateRange(value any, min, max float64, fieldName string) error {
	var numValue float64
	var err error

	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int8:
		numValue = float64(v)
	case int16:
		numValue = float64(v)
	case int32:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	case string:
		numValue, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("%s must be a valid number", fieldName)
		}
	default:
		return fmt.Errorf("%s must be a number", fieldName)
	}

	if numValue < min {
		return fmt.Errorf("%s must be at least %.2f", fieldName, min)
	}

	if numValue > max {
		return fmt.Errorf("%s must not exceed %.2f", fieldName, max)
	}

	return nil
}

// ValidatePattern 验证正则表达式模式
func (c *BaseController) ValidatePattern(value, pattern, fieldName string) error {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return fmt.Errorf("invalid pattern for %s", fieldName)
	}

	if !matched {
		return fmt.Errorf("%s format is invalid", fieldName)
	}

	return nil
}

// ValidateIn 验证值是否在指定列表中
func (c *BaseController) ValidateIn(value any, allowed []any, fieldName string) error {
	for _, allowedValue := range allowed {
		if reflect.DeepEqual(value, allowedValue) {
			return nil
		}
	}

	return fmt.Errorf("%s is not a valid value", fieldName)
}

// ============= 常用验证规则 =============

// ValidateEmailFormat 验证邮箱格式
func (c *BaseController) ValidateEmailFormat(email, fieldName string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return c.ValidatePattern(email, pattern, fieldName)
}

// ValidatePhoneFormat 验证手机号格式（中国大陆）
func (c *BaseController) ValidatePhoneFormat(phone, fieldName string) error {
	pattern := `^1[3-9]\d{9}$`
	return c.ValidatePattern(phone, pattern, fieldName)
}

// ValidateIDCardFormat 验证身份证号格式（中国大陆）
func (c *BaseController) ValidateIDCardFormat(idCard, fieldName string) error {
	pattern := `^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`
	return c.ValidatePattern(idCard, pattern, fieldName)
}

// ValidateURLFormat 验证URL格式
func (c *BaseController) ValidateURLFormat(url, fieldName string) error {
	pattern := `^https?://[^\s/$.?#].[^\s]*$`
	return c.ValidatePattern(url, pattern, fieldName)
}

// ValidateIPFormat 验证IP地址格式
func (c *BaseController) ValidateIPFormat(ip, fieldName string) error {
	pattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	return c.ValidatePattern(ip, pattern, fieldName)
}

// ValidatePasswordStrength 验证密码强度
func (c *BaseController) ValidatePasswordStrength(password, fieldName string) error {
	if len(password) < 8 {
		return fmt.Errorf("%s must be at least 8 characters long", fieldName)
	}

	// 检查是否包含数字
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 检查是否包含小写字母
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// 检查是否包含大写字母
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// 检查是否包含特殊字符
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	strengthCount := 0
	if hasNumber {
		strengthCount++
	}
	if hasLower {
		strengthCount++
	}
	if hasUpper {
		strengthCount++
	}
	if hasSpecial {
		strengthCount++
	}

	if strengthCount < 3 {
		return fmt.Errorf("%s must contain at least 3 of: numbers, lowercase letters, uppercase letters, special characters", fieldName)
	}

	return nil
}

// ============= 批量验证方法 =============

// ValidationRule 验证规则结构
type ValidationRule struct {
	Field   string         // 字段名
	Value   any            // 字段值
	Rules   []string       // 验证规则列表
	Message string         // 自定义错误消息
	Params  map[string]any // 规则参数
}

// ValidateBatch 批量验证
func (c *BaseController) ValidateBatch(rules []ValidationRule) []error {
	var errors []error

	for _, rule := range rules {
		for _, ruleName := range rule.Rules {
			if err := c.applyRule(rule, ruleName); err != nil {
				// 使用自定义消息或默认错误消息
				if rule.Message != "" {
					errors = append(errors, fmt.Errorf("%s", rule.Message))
				} else {
					errors = append(errors, err)
				}
				break // 一个字段遇到错误就停止验证该字段的其他规则
			}
		}
	}

	return errors
}

// applyRule 应用验证规则
func (c *BaseController) applyRule(rule ValidationRule, ruleName string) error {
	switch ruleName {
	case "required":
		return c.ValidateRequired(rule.Value, rule.Field)

	case "email":
		if str, ok := rule.Value.(string); ok {
			return c.ValidateEmailFormat(str, rule.Field)
		}
		return fmt.Errorf("%s must be a string for email validation", rule.Field)

	case "phone":
		if str, ok := rule.Value.(string); ok {
			return c.ValidatePhoneFormat(str, rule.Field)
		}
		return fmt.Errorf("%s must be a string for phone validation", rule.Field)

	case "url":
		if str, ok := rule.Value.(string); ok {
			return c.ValidateURLFormat(str, rule.Field)
		}
		return fmt.Errorf("%s must be a string for URL validation", rule.Field)

	case "ip":
		if str, ok := rule.Value.(string); ok {
			return c.ValidateIPFormat(str, rule.Field)
		}
		return fmt.Errorf("%s must be a string for IP validation", rule.Field)

	case "password":
		if str, ok := rule.Value.(string); ok {
			return c.ValidatePasswordStrength(str, rule.Field)
		}
		return fmt.Errorf("%s must be a string for password validation", rule.Field)

	case "length":
		if str, ok := rule.Value.(string); ok {
			min := c.getIntParam(rule.Params, "min", 0)
			max := c.getIntParam(rule.Params, "max", 0)
			return c.ValidateLength(str, min, max, rule.Field)
		}
		return fmt.Errorf("%s must be a string for length validation", rule.Field)

	case "range":
		min := c.getFloatParam(rule.Params, "min", 0)
		max := c.getFloatParam(rule.Params, "max", 0)
		return c.ValidateRange(rule.Value, min, max, rule.Field)

	case "pattern":
		if str, ok := rule.Value.(string); ok {
			pattern := c.getStringParam(rule.Params, "pattern", "")
			if pattern == "" {
				return fmt.Errorf("pattern parameter is required for pattern validation")
			}
			return c.ValidatePattern(str, pattern, rule.Field)
		}
		return fmt.Errorf("%s must be a string for pattern validation", rule.Field)

	default:
		return fmt.Errorf("unknown validation rule: %s", ruleName)
	}
}

// ============= 参数辅助方法 =============

// getIntParam 获取整数参数
func (c *BaseController) getIntParam(params map[string]any, key string, defaultValue int) int {
	if params == nil {
		return defaultValue
	}

	if value, exists := params[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
		if strValue, ok := value.(string); ok {
			if intValue, err := strconv.Atoi(strValue); err == nil {
				return intValue
			}
		}
	}

	return defaultValue
}

// getFloatParam 获取浮点数参数
func (c *BaseController) getFloatParam(params map[string]any, key string, defaultValue float64) float64 {
	if params == nil {
		return defaultValue
	}

	if value, exists := params[key]; exists {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
		if intValue, ok := value.(int); ok {
			return float64(intValue)
		}
		if strValue, ok := value.(string); ok {
			if floatValue, err := strconv.ParseFloat(strValue, 64); err == nil {
				return floatValue
			}
		}
	}

	return defaultValue
}

// getStringParam 获取字符串参数
func (c *BaseController) getStringParam(params map[string]any, key string, defaultValue string) string {
	if params == nil {
		return defaultValue
	}

	if value, exists := params[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}

	return defaultValue
}

// ============= 表单验证辅助方法 =============

// ValidateForm 验证表单数据
func (c *BaseController) ValidateForm(validationMap map[string][]string) map[string]string {
	errors := make(map[string]string)

	for field, rules := range validationMap {
		value := c.GetForm(field)

		for _, rule := range rules {
			var err error

			// 解析规则和参数
			parts := strings.Split(rule, ":")
			ruleName := parts[0]

			switch ruleName {
			case "required":
				err = c.ValidateRequired(value, field)
			case "email":
				if value != "" {
					err = c.ValidateEmailFormat(value, field)
				}
			case "phone":
				if value != "" {
					err = c.ValidatePhoneFormat(value, field)
				}
			case "url":
				if value != "" {
					err = c.ValidateURLFormat(value, field)
				}
			case "min":
				if len(parts) > 1 {
					if min, parseErr := strconv.Atoi(parts[1]); parseErr == nil {
						err = c.ValidateLength(value, min, 0, field)
					}
				}
			case "max":
				if len(parts) > 1 {
					if max, parseErr := strconv.Atoi(parts[1]); parseErr == nil {
						err = c.ValidateLength(value, 0, max, field)
					}
				}
			}

			if err != nil {
				errors[field] = err.Error()
				break // 一个字段遇到错误就停止验证该字段的其他规则
			}
		}
	}

	return errors
}

// HasValidationErrors 检查是否有验证错误
func (c *BaseController) HasValidationErrors(errors map[string]string) bool {
	return len(errors) > 0
}

// GetFirstValidationError 获取第一个验证错误
func (c *BaseController) GetFirstValidationError(errors map[string]string) string {
	for _, error := range errors {
		return error
	}
	return ""
}

// SetValidationErrors 设置验证错误到模板数据
func (c *BaseController) SetValidationErrors(errors map[string]string) {
	c.SetData("ValidationErrors", errors)
	c.SetData("HasErrors", len(errors) > 0)
}

// ============= 自定义验证器支持 =============

// ValidatorFunc 自定义验证器函数类型
type ValidatorFunc func(value any, params map[string]any) error

// customValidators 自定义验证器映射
var customValidators = make(map[string]ValidatorFunc)

// RegisterValidator 注册自定义验证器
func (c *BaseController) RegisterValidator(name string, validator ValidatorFunc) {
	customValidators[name] = validator
}

// ApplyCustomValidator 应用自定义验证器
func (c *BaseController) ApplyCustomValidator(name string, value any, params map[string]any) error {
	if validator, exists := customValidators[name]; exists {
		return validator(value, params)
	}
	return fmt.Errorf("custom validator '%s' not found", name)
}

// ============= 数据类型验证 =============

// ValidateDateTime 验证日期时间格式
func (c *BaseController) ValidateDateTime(value, layout, fieldName string) error {
	_, err := time.Parse(layout, value)
	if err != nil {
		return fmt.Errorf("%s must be a valid datetime in format %s", fieldName, layout)
	}
	return nil
}

// ValidateJSON 验证JSON格式
func (c *BaseController) ValidateJSON(value, fieldName string) error {
	// 简单的JSON格式检查
	value = strings.TrimSpace(value)
	if !((strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")) ||
		(strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]"))) {
		return fmt.Errorf("%s must be valid JSON", fieldName)
	}
	return nil
}

// ValidateNumeric 验证是否为数值
func (c *BaseController) ValidateNumeric(value, fieldName string) error {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("%s must be a valid number", fieldName)
	}
	return nil
}

// ValidateInteger 验证是否为整数
func (c *BaseController) ValidateInteger(value, fieldName string) error {
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%s must be a valid integer", fieldName)
	}
	return nil
}

// ValidateBoolean 验证是否为布尔值
func (c *BaseController) ValidateBoolean(value, fieldName string) error {
	value = strings.ToLower(strings.TrimSpace(value))
	validBooleans := []string{"true", "false", "1", "0", "yes", "no", "on", "off"}

	for _, valid := range validBooleans {
		if value == valid {
			return nil
		}
	}

	return fmt.Errorf("%s must be a valid boolean value", fieldName)
}

// ============= 验证结果处理 =============

// ValidationResult 验证结果结构
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  map[string]string `json:"errors"`
	Data    map[string]any    `json:"data"`
}

// CreateValidationResult 创建验证结果
func (c *BaseController) CreateValidationResult(errors map[string]string, data map[string]any) *ValidationResult {
	return &ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
		Data:    data,
	}
}

// ReturnValidationResult 返回验证结果JSON响应
func (c *BaseController) ReturnValidationResult(result *ValidationResult) {
	if result.IsValid {
		c.JSONSuccess("Validation passed", result.Data)
	} else {
		c.JSONError("Validation failed")
		c.SetData("errors", result.Errors)
	}
}

// LogValidationError 记录验证错误
func (c *BaseController) LogValidationError(field, rule, value string, err error) {
	config.Warnf("Validation failed - Field: %s, Rule: %s, Value: %s, Error: %v", field, rule, value, err)
}
