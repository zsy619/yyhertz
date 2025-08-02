// Package validation 提供强大的数据验证功能
//
// 这个包提供了类似Beego的验证功能，包括：
// - 内置验证规则
// - 自定义验证器
// - 表单数据绑定和验证
// - 国际化错误消息
// - 链式验证调用
package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
)

// Validator 验证器结构
type Validator struct {
	errors   []ValidationError
	mutex    sync.RWMutex
	funcs    map[string]ValidatorFunc
	messages map[string]string
	locale   string
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   any    `json:"value"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

// ValidatorFunc 验证器函数类型
type ValidatorFunc func(value any, param string) bool

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidationRule 验证规则
type ValidationRule struct {
	Field     string
	Tag       string
	Param     string
	Message   string
	Required  bool
	Validator ValidatorFunc
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	Locale         string            `json:"locale" yaml:"locale"`
	StopOnFirst    bool              `json:"stop_on_first" yaml:"stop_on_first"`
	CustomMessages map[string]string `json:"custom_messages" yaml:"custom_messages"`
	TagName        string            `json:"tag_name" yaml:"tag_name"`
}

var (
	defaultValidator *Validator
	validatorOnce    sync.Once
	validatorMutex   sync.Mutex
)

// GetDefaultValidator 获取默认验证器
func GetDefaultValidator() *Validator {
	validatorOnce.Do(func() {
		validatorMutex.Lock()
		defer validatorMutex.Unlock()

		defaultValidator = NewValidator(&ValidationConfig{
			Locale:      "zh-CN",
			StopOnFirst: false,
			TagName:     "validate",
		})
	})
	return defaultValidator
}

// NewValidator 创建新的验证器
func NewValidator(cfg *ValidationConfig) *Validator {
	if cfg == nil {
		cfg = &ValidationConfig{
			Locale:      "zh-CN",
			StopOnFirst: false,
			TagName:     "validate",
		}
	}

	v := &Validator{
		errors:   make([]ValidationError, 0),
		funcs:    make(map[string]ValidatorFunc),
		messages: make(map[string]string),
		locale:   cfg.Locale,
	}

	// 注册内置验证器
	v.registerBuiltinValidators()

	// 加载默认错误消息
	v.loadDefaultMessages()

	// 加载自定义消息
	if cfg.CustomMessages != nil {
		for key, msg := range cfg.CustomMessages {
			v.messages[key] = msg
		}
	}

	config.Infof("Validator initialized with locale: %s", v.locale)
	return v
}

// RegisterValidator 注册自定义验证器
func (v *Validator) RegisterValidator(tag string, fn ValidatorFunc) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.funcs[tag] = fn
	config.Debugf("Registered custom validator: %s", tag)
}

// SetMessage 设置错误消息
func (v *Validator) SetMessage(tag, message string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.messages[tag] = message
}

// Validate 验证单个值
func (v *Validator) Validate(value any, rules string) *ValidationResult {
	return v.ValidateField("", value, rules)
}

// ValidateField 验证字段
func (v *Validator) ValidateField(field string, value any, rules string) *ValidationResult {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	// 清空之前的错误
	v.errors = v.errors[:0]

	if rules == "" {
		return &ValidationResult{Valid: true}
	}

	// 解析验证规则
	ruleList := v.parseRules(field, rules)

	// 执行验证
	for _, rule := range ruleList {
		if !v.executeRule(value, rule) {
			v.errors = append(v.errors, ValidationError{
				Field:   rule.Field,
				Tag:     rule.Tag,
				Value:   value,
				Message: v.getErrorMessage(rule.Tag, rule.Field, rule.Param),
				Param:   rule.Param,
			})
		}
	}

	return &ValidationResult{
		Valid:  len(v.errors) == 0,
		Errors: append([]ValidationError(nil), v.errors...),
	}
}

// ValidateStruct 验证结构体
func (v *Validator) ValidateStruct(obj any) *ValidationResult {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	// 清空之前的错误
	v.errors = v.errors[:0]

	objValue := reflect.ValueOf(obj)
	objType := reflect.TypeOf(obj)

	// 处理指针
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
		objType = objType.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		v.errors = append(v.errors, ValidationError{
			Field:   "",
			Tag:     "struct",
			Value:   obj,
			Message: "待验证对象必须是结构体",
		})
		return &ValidationResult{
			Valid:  false,
			Errors: append([]ValidationError(nil), v.errors...),
		}
	}

	// 遍历结构体字段
	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i)

		// 跳过私有字段
		if !fieldValue.CanInterface() {
			continue
		}

		// 获取验证标签
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// 获取字段名
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if parts := strings.Split(jsonTag, ","); parts[0] != "" {
				fieldName = parts[0]
			}
		}

		// 验证字段
		result := v.ValidateField(fieldName, fieldValue.Interface(), validateTag)
		if !result.Valid {
			v.errors = append(v.errors, result.Errors...)
		}
	}

	return &ValidationResult{
		Valid:  len(v.errors) == 0,
		Errors: append([]ValidationError(nil), v.errors...),
	}
}

// parseRules 解析验证规则
func (v *Validator) parseRules(field, rules string) []ValidationRule {
	var ruleList []ValidationRule

	// 分割多个规则
	ruleParts := strings.Split(rules, "|")

	for _, rulePart := range ruleParts {
		rulePart = strings.TrimSpace(rulePart)
		if rulePart == "" {
			continue
		}

		// 解析规则和参数
		var tag, param string
		if idx := strings.Index(rulePart, ":"); idx != -1 {
			tag = strings.TrimSpace(rulePart[:idx])
			param = strings.TrimSpace(rulePart[idx+1:])
		} else {
			tag = rulePart
		}

		// 检查是否为必填
		required := tag == "required"

		// 获取验证器函数
		validatorFunc := v.funcs[tag]
		if validatorFunc == nil {
			config.Warnf("Unknown validator tag: %s", tag)
			continue
		}

		ruleList = append(ruleList, ValidationRule{
			Field:     field,
			Tag:       tag,
			Param:     param,
			Required:  required,
			Validator: validatorFunc,
		})
	}

	return ruleList
}

// executeRule 执行验证规则
func (v *Validator) executeRule(value any, rule ValidationRule) bool {
	// 处理nil值
	if value == nil {
		return !rule.Required // 如果不是必填，nil值通过验证
	}

	// 处理空值
	if v.isEmpty(value) {
		return !rule.Required // 如果不是必填，空值通过验证
	}

	// 执行验证器函数
	return rule.Validator(value, rule.Param)
}

// isEmpty 检查值是否为空
func (v *Validator) isEmpty(value any) bool {
	if value == nil {
		return true
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	default:
		return false
	}
}

// getErrorMessage 获取错误消息
func (v *Validator) getErrorMessage(tag, field, param string) string {
	// 优先使用字段特定消息
	if msg, exists := v.messages[field+"."+tag]; exists {
		return v.formatMessage(msg, field, param)
	}

	// 使用通用消息
	if msg, exists := v.messages[tag]; exists {
		return v.formatMessage(msg, field, param)
	}

	// 默认消息
	return fmt.Sprintf("字段 %s 验证失败: %s", field, tag)
}

// formatMessage 格式化错误消息
func (v *Validator) formatMessage(message, field, param string) string {
	message = strings.ReplaceAll(message, "{field}", field)
	message = strings.ReplaceAll(message, "{param}", param)
	return message
}

// Clear 清空验证错误
func (v *Validator) Clear() {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.errors = v.errors[:0]
}

// HasErrors 检查是否有验证错误
func (v *Validator) HasErrors() bool {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	return len(v.errors) > 0
}

// GetErrors 获取所有验证错误
func (v *Validator) GetErrors() []ValidationError {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	return append([]ValidationError(nil), v.errors...)
}

// GetErrorMessages 获取错误消息列表
func (v *Validator) GetErrorMessages() []string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	messages := make([]string, len(v.errors))
	for i, err := range v.errors {
		messages[i] = err.Message
	}
	return messages
}

// FirstError 获取第一个错误
func (v *Validator) FirstError() *ValidationError {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	if len(v.errors) > 0 {
		return &v.errors[0]
	}
	return nil
}

// ============= 内置验证器 =============

// registerBuiltinValidators 注册内置验证器
func (v *Validator) registerBuiltinValidators() {
	// 基础验证器
	v.funcs["required"] = v.validateRequired
	v.funcs["min"] = v.validateMin
	v.funcs["max"] = v.validateMax
	v.funcs["range"] = v.validateRange
	v.funcs["minLength"] = v.validateMinLength
	v.funcs["maxLength"] = v.validateMaxLength
	v.funcs["length"] = v.validateLength

	// 格式验证器
	v.funcs["email"] = v.validateEmail
	v.funcs["url"] = v.validateURL
	v.funcs["alpha"] = v.validateAlpha
	v.funcs["alphaNum"] = v.validateAlphaNum
	v.funcs["numeric"] = v.validateNumeric
	v.funcs["integer"] = v.validateInteger
	v.funcs["decimal"] = v.validateDecimal
	v.funcs["ip"] = v.validateIP
	v.funcs["ipv4"] = v.validateIPv4
	v.funcs["ipv6"] = v.validateIPv6
	v.funcs["mac"] = v.validateMAC
	v.funcs["uuid"] = v.validateUUID

	// 日期时间验证器
	v.funcs["date"] = v.validateDate
	v.funcs["datetime"] = v.validateDateTime
	v.funcs["time"] = v.validateTime
	v.funcs["before"] = v.validateBefore
	v.funcs["after"] = v.validateAfter

	// 字符串验证器
	v.funcs["regex"] = v.validateRegex
	v.funcs["contains"] = v.validateContains
	v.funcs["startsWith"] = v.validateStartsWith
	v.funcs["endsWith"] = v.validateEndsWith
	v.funcs["in"] = v.validateIn
	v.funcs["notIn"] = v.validateNotIn

	// 数字验证器
	v.funcs["positive"] = v.validatePositive
	v.funcs["negative"] = v.validateNegative
	v.funcs["nonZero"] = v.validateNonZero

	// 中国特有验证器
	v.funcs["mobile"] = v.validateMobile
	v.funcs["phone"] = v.validatePhone
	v.funcs["idCard"] = v.validateIDCard
	v.funcs["zipCode"] = v.validateZipCode

	config.Debug("Built-in validators registered")
}

// ============= 基础验证器实现 =============

// validateRequired 必填验证
func (v *Validator) validateRequired(value any, param string) bool {
	return !v.isEmpty(value)
}

// validateMin 最小值验证
func (v *Validator) validateMin(value any, param string) bool {
	minVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return false
	}

	val := v.toFloat64(value)
	if val == nil {
		return false
	}

	return *val >= minVal
}

// validateMax 最大值验证
func (v *Validator) validateMax(value any, param string) bool {
	maxVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return false
	}

	val := v.toFloat64(value)
	if val == nil {
		return false
	}

	return *val <= maxVal
}

// validateRange 范围验证
func (v *Validator) validateRange(value any, param string) bool {
	parts := strings.Split(param, ",")
	if len(parts) != 2 {
		return false
	}

	min, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	max, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return false
	}

	val := v.toFloat64(value)
	if val == nil {
		return false
	}

	return *val >= min && *val <= max
}

// validateMinLength 最小长度验证
func (v *Validator) validateMinLength(value any, param string) bool {
	minLen, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	length := v.getLength(value)
	return length >= minLen
}

// validateMaxLength 最大长度验证
func (v *Validator) validateMaxLength(value any, param string) bool {
	maxLen, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	length := v.getLength(value)
	return length <= maxLen
}

// validateLength 长度验证
func (v *Validator) validateLength(value any, param string) bool {
	expectedLen, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	length := v.getLength(value)
	return length == expectedLen
}

// ============= 辅助方法 =============

// toFloat64 转换为float64
func (v *Validator) toFloat64(value any) *float64 {
	var val float64
	var err error

	switch v := value.(type) {
	case int:
		val = float64(v)
	case int8:
		val = float64(v)
	case int16:
		val = float64(v)
	case int32:
		val = float64(v)
	case int64:
		val = float64(v)
	case uint:
		val = float64(v)
	case uint8:
		val = float64(v)
	case uint16:
		val = float64(v)
	case uint32:
		val = float64(v)
	case uint64:
		val = float64(v)
	case float32:
		val = float64(v)
	case float64:
		val = v
	case string:
		val, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil
		}
	default:
		return nil
	}

	return &val
}

// getLength 获取长度
func (v *Validator) getLength(value any) int {
	if value == nil {
		return 0
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		return len([]rune(val.String())) // 使用rune计算中文字符长度
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len()
	default:
		return 0
	}
}

// ============= 便捷函数 =============

// Validate 验证单个值（使用默认验证器）
func Validate(value any, rules string) *ValidationResult {
	return GetDefaultValidator().Validate(value, rules)
}

// ValidateField 验证字段（使用默认验证器）
func ValidateField(field string, value any, rules string) *ValidationResult {
	return GetDefaultValidator().ValidateField(field, value, rules)
}

// ValidateStruct 验证结构体（使用默认验证器）
func ValidateStruct(obj any) *ValidationResult {
	return GetDefaultValidator().ValidateStruct(obj)
}

// RegisterValidator 注册验证器（使用默认验证器）
func RegisterValidator(tag string, fn ValidatorFunc) {
	GetDefaultValidator().RegisterValidator(tag, fn)
}

// SetMessage 设置错误消息（使用默认验证器）
func SetMessage(tag, message string) {
	GetDefaultValidator().SetMessage(tag, message)
}

// ============= 错误处理 =============

// Error 实现error接口
func (ve ValidationError) Error() string {
	return ve.Message
}

// String 字符串表示
func (ve ValidationError) String() string {
	return fmt.Sprintf("Field: %s, Tag: %s, Message: %s", ve.Field, ve.Tag, ve.Message)
}

// ValidationErrors 验证错误集合
type ValidationErrors []ValidationError

// Error 实现error接口
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	messages := make([]string, len(ve))
	for i, err := range ve {
		messages[i] = err.Message
	}
	return strings.Join(messages, "; ")
}

// First 获取第一个错误
func (ve ValidationErrors) First() *ValidationError {
	if len(ve) > 0 {
		return &ve[0]
	}
	return nil
}

// ByField 按字段获取错误
func (ve ValidationErrors) ByField(field string) []ValidationError {
	var errors []ValidationError
	for _, err := range ve {
		if err.Field == field {
			errors = append(errors, err)
		}
	}
	return errors
}

// ToMap 转换为字段-错误映射
func (ve ValidationErrors) ToMap() map[string][]string {
	result := make(map[string][]string)
	for _, err := range ve {
		result[err.Field] = append(result[err.Field], err.Message)
	}
	return result
}
