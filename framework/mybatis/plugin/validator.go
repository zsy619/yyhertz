// Package plugin 参数验证插件实现
//
// 提供输入参数验证功能，支持多种验证规则和自定义验证器
package plugin

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidatorPlugin 参数验证插件
type ValidatorPlugin struct {
	*BasePlugin
	enableValidation bool
	validators       map[string]Validator
	rules            map[string][]ValidationRule
}

// Validator 验证器接口
type Validator interface {
	Validate(value any, rule ValidationRule) error
	GetName() string
}

// ValidationRule 验证规则
type ValidationRule struct {
	Field     string         // 字段名
	Type      string         // 验证类型
	Required  bool           // 是否必需
	Message   string         // 错误消息
	Params    map[string]any // 验证参数
	Condition func(any) bool // 验证条件
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string // 字段名
	Message string // 错误消息
	Value   any    // 字段值
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool              // 是否有效
	Errors []ValidationError // 错误列表
}

// NewValidatorPlugin 创建参数验证插件
func NewValidatorPlugin() *ValidatorPlugin {
	plugin := &ValidatorPlugin{
		BasePlugin:       NewBasePlugin("validator", 5),
		enableValidation: true,
		validators:       make(map[string]Validator),
		rules:            make(map[string][]ValidationRule),
	}

	// 注册默认验证器
	plugin.registerDefaultValidators()

	return plugin
}

// Intercept 拦截方法调用
func (plugin *ValidatorPlugin) Intercept(invocation *Invocation) (any, error) {
	if !plugin.enableValidation {
		return invocation.Proceed()
	}

	// 验证参数
	if err := plugin.validateParameters(invocation); err != nil {
		return nil, err
	}

	// 执行原方法
	return invocation.Proceed()
}

// Plugin 包装目标对象
func (plugin *ValidatorPlugin) Plugin(target any) any {
	return target
}

// SetProperties 设置插件属性
func (plugin *ValidatorPlugin) SetProperties(properties map[string]any) {
	plugin.BasePlugin.SetProperties(properties)

	plugin.enableValidation = plugin.GetPropertyBool("enableValidation", true)
}

// validateParameters 验证参数
func (plugin *ValidatorPlugin) validateParameters(invocation *Invocation) error {
	methodName := invocation.Method.Name
	rules, exists := plugin.rules[methodName]
	if !exists || len(rules) == 0 {
		return nil // 没有验证规则
	}

	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// 验证每个参数
	for _, arg := range invocation.Args {
		plugin.validateArgument(arg, rules, result)
	}

	if !result.Valid {
		return plugin.createValidationError(result.Errors)
	}

	return nil
}

// validateArgument 验证单个参数
func (plugin *ValidatorPlugin) validateArgument(arg any, rules []ValidationRule, result *ValidationResult) {
	if arg == nil {
		return
	}

	// 如果是结构体，验证字段
	if plugin.isStruct(arg) {
		plugin.validateStruct(arg, rules, result)
		return
	}

	// 如果是map，验证键值
	if argMap, ok := arg.(map[string]any); ok {
		plugin.validateMap(argMap, rules, result)
		return
	}

	// 其他类型的简单验证
	plugin.validateSimpleValue(arg, rules, result)
}

// validateStruct 验证结构体
func (plugin *ValidatorPlugin) validateStruct(obj any, rules []ValidationRule, result *ValidationResult) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		// 查找对应的验证规则
		for _, rule := range rules {
			if plugin.matchField(field.Name, rule.Field) {
				plugin.validateFieldValue(field.Name, fieldValue.Interface(), rule, result)
			}
		}
	}
}

// validateMap 验证Map
func (plugin *ValidatorPlugin) validateMap(argMap map[string]any, rules []ValidationRule, result *ValidationResult) {
	for key, value := range argMap {
		for _, rule := range rules {
			if plugin.matchField(key, rule.Field) {
				plugin.validateFieldValue(key, value, rule, result)
			}
		}
	}
}

// validateSimpleValue 验证简单值
func (plugin *ValidatorPlugin) validateSimpleValue(value any, rules []ValidationRule, result *ValidationResult) {
	for _, rule := range rules {
		if rule.Field == "" || rule.Field == "*" {
			plugin.validateFieldValue("value", value, rule, result)
		}
	}
}

// validateFieldValue 验证字段值
func (plugin *ValidatorPlugin) validateFieldValue(fieldName string, value any, rule ValidationRule, result *ValidationResult) {
	// 检查条件
	if rule.Condition != nil && !rule.Condition(value) {
		return
	}

	// 检查必需字段
	if rule.Required && plugin.isEmpty(value) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Message: plugin.getErrorMessage(rule, "字段不能为空"),
			Value:   value,
		})
		return
	}

	// 如果值为空且不是必需字段，跳过验证
	if plugin.isEmpty(value) && !rule.Required {
		return
	}

	// 执行具体验证
	validator, exists := plugin.validators[rule.Type]
	if !exists {
		return // 没有对应的验证器
	}

	if err := validator.Validate(value, rule); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Message: plugin.getErrorMessage(rule, err.Error()),
			Value:   value,
		})
	}
}

// matchField 匹配字段名
func (plugin *ValidatorPlugin) matchField(fieldName, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}

	// 支持通配符匹配
	if strings.Contains(pattern, "*") {
		matched, _ := regexp.MatchString(strings.ReplaceAll(pattern, "*", ".*"), fieldName)
		return matched
	}

	return strings.EqualFold(fieldName, pattern)
}

// isEmpty 检查值是否为空
func (plugin *ValidatorPlugin) isEmpty(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// isStruct 检查是否为结构体
func (plugin *ValidatorPlugin) isStruct(obj any) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Kind() == reflect.Struct
}

// getErrorMessage 获取错误消息
func (plugin *ValidatorPlugin) getErrorMessage(rule ValidationRule, defaultMsg string) string {
	if rule.Message != "" {
		return rule.Message
	}
	return defaultMsg
}

// createValidationError 创建验证错误
func (plugin *ValidatorPlugin) createValidationError(errors []ValidationError) error {
	var messages []string
	for _, err := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return fmt.Errorf("参数验证失败: %s", strings.Join(messages, "; "))
}

// registerDefaultValidators 注册默认验证器
func (plugin *ValidatorPlugin) registerDefaultValidators() {
	plugin.RegisterValidator(&RequiredValidator{})
	plugin.RegisterValidator(&LengthValidator{})
	plugin.RegisterValidator(&RangeValidator{})
	plugin.RegisterValidator(&PatternValidator{})
	plugin.RegisterValidator(&EmailValidator{})
	plugin.RegisterValidator(&PhoneValidator{})
	plugin.RegisterValidator(&NumberValidator{})
}

// RegisterValidator 注册验证器
func (plugin *ValidatorPlugin) RegisterValidator(validator Validator) {
	plugin.validators[validator.GetName()] = validator
}

// AddRule 添加验证规则
func (plugin *ValidatorPlugin) AddRule(methodName string, rule ValidationRule) {
	if plugin.rules[methodName] == nil {
		plugin.rules[methodName] = make([]ValidationRule, 0)
	}
	plugin.rules[methodName] = append(plugin.rules[methodName], rule)
}

// AddRules 批量添加验证规则
func (plugin *ValidatorPlugin) AddRules(methodName string, rules []ValidationRule) {
	for _, rule := range rules {
		plugin.AddRule(methodName, rule)
	}
}

// 内置验证器实现

// RequiredValidator 必需验证器
type RequiredValidator struct{}

func (v *RequiredValidator) GetName() string { return "required" }

func (v *RequiredValidator) Validate(value any, rule ValidationRule) error {
	if value == nil {
		return fmt.Errorf("值不能为空")
	}

	if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
		return fmt.Errorf("字符串不能为空")
	}

	return nil
}

// LengthValidator 长度验证器
type LengthValidator struct{}

func (v *LengthValidator) GetName() string { return "length" }

func (v *LengthValidator) Validate(value any, rule ValidationRule) error {
	str := fmt.Sprintf("%v", value)
	length := len(str)

	if minLen, exists := rule.Params["min"]; exists {
		if min, ok := minLen.(int); ok && length < min {
			return fmt.Errorf("长度不能少于%d个字符", min)
		}
	}

	if maxLen, exists := rule.Params["max"]; exists {
		if max, ok := maxLen.(int); ok && length > max {
			return fmt.Errorf("长度不能超过%d个字符", max)
		}
	}

	return nil
}

// RangeValidator 范围验证器
type RangeValidator struct{}

func (v *RangeValidator) GetName() string { return "range" }

func (v *RangeValidator) Validate(value any, rule ValidationRule) error {
	num, err := v.toFloat64(value)
	if err != nil {
		return fmt.Errorf("值必须是数字")
	}

	if minVal, exists := rule.Params["min"]; exists {
		if min, ok := minVal.(float64); ok && num < min {
			return fmt.Errorf("值不能小于%.2f", min)
		}
	}

	if maxVal, exists := rule.Params["max"]; exists {
		if max, ok := maxVal.(float64); ok && num > max {
			return fmt.Errorf("值不能大于%.2f", max)
		}
	}

	return nil
}

func (v *RangeValidator) toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("无法转换为数字")
	}
}

// PatternValidator 正则表达式验证器
type PatternValidator struct{}

func (v *PatternValidator) GetName() string { return "pattern" }

func (v *PatternValidator) Validate(value any, rule ValidationRule) error {
	str := fmt.Sprintf("%v", value)

	pattern, exists := rule.Params["pattern"]
	if !exists {
		return fmt.Errorf("缺少正则表达式模式")
	}

	patternStr, ok := pattern.(string)
	if !ok {
		return fmt.Errorf("正则表达式模式必须是字符串")
	}

	matched, err := regexp.MatchString(patternStr, str)
	if err != nil {
		return fmt.Errorf("正则表达式错误: %v", err)
	}

	if !matched {
		return fmt.Errorf("格式不正确")
	}

	return nil
}

// EmailValidator 邮箱验证器
type EmailValidator struct{}

func (v *EmailValidator) GetName() string { return "email" }

func (v *EmailValidator) Validate(value any, rule ValidationRule) error {
	str := fmt.Sprintf("%v", value)

	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailPattern, str)
	if err != nil {
		return fmt.Errorf("邮箱验证错误: %v", err)
	}

	if !matched {
		return fmt.Errorf("邮箱格式不正确")
	}

	return nil
}

// PhoneValidator 手机号验证器
type PhoneValidator struct{}

func (v *PhoneValidator) GetName() string { return "phone" }

func (v *PhoneValidator) Validate(value any, rule ValidationRule) error {
	str := fmt.Sprintf("%v", value)

	// 中国手机号格式
	phonePattern := `^1[3-9]\d{9}$`
	matched, err := regexp.MatchString(phonePattern, str)
	if err != nil {
		return fmt.Errorf("手机号验证错误: %v", err)
	}

	if !matched {
		return fmt.Errorf("手机号格式不正确")
	}

	return nil
}

// NumberValidator 数字验证器
type NumberValidator struct{}

func (v *NumberValidator) GetName() string { return "number" }

func (v *NumberValidator) Validate(value any, rule ValidationRule) error {
	switch value.(type) {
	case int, int32, int64, float32, float64:
		return nil
	case string:
		str := value.(string)
		if _, err := strconv.ParseFloat(str, 64); err != nil {
			return fmt.Errorf("必须是有效的数字")
		}
		return nil
	default:
		return fmt.Errorf("必须是数字类型")
	}
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	Enabled bool                        `json:"enabled" yaml:"enabled"`
	Rules   map[string][]ValidationRule `json:"rules" yaml:"rules"`
}

// NewValidationConfig 创建默认验证配置
func NewValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		Enabled: true,
		Rules:   make(map[string][]ValidationRule),
	}
}
