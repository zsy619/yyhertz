package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FormValidator 表单验证器
type FormValidator struct {
	validator *Validator
	data      map[string]any
	errors    map[string][]string
}

// FormData 表单数据接口
type FormData interface {
	Get(key string) string
	GetAll(key string) []string
	Has(key string) bool
	Keys() []string
}

// MapFormData map类型表单数据
type MapFormData map[string]any

// Get 获取值
func (m MapFormData) Get(key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// GetAll 获取所有值
func (m MapFormData) GetAll(key string) []string {
	if val, exists := m[key]; exists {
		if slice, ok := val.([]string); ok {
			return slice
		}
		if str, ok := val.(string); ok {
			return []string{str}
		}
		return []string{fmt.Sprintf("%v", val)}
	}
	return []string{}
}

// Has 检查是否存在
func (m MapFormData) Has(key string) bool {
	_, exists := m[key]
	return exists
}

// Keys 获取所有键
func (m MapFormData) Keys() []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// NewFormValidator 创建表单验证器
func NewFormValidator(data FormData) *FormValidator {
	fv := &FormValidator{
		validator: GetDefaultValidator(),
		data:      make(map[string]any),
		errors:    make(map[string][]string),
	}

	// 转换表单数据
	for _, key := range data.Keys() {
		values := data.GetAll(key)
		if len(values) == 1 {
			fv.data[key] = values[0]
		} else {
			fv.data[key] = values
		}
	}

	return fv
}

// ValidateField 验证单个字段
func (fv *FormValidator) ValidateField(field, rules string) *FormValidator {
	value, exists := fv.data[field]
	if !exists {
		value = ""
	}

	result := fv.validator.ValidateField(field, value, rules)
	if !result.Valid {
		for _, err := range result.Errors {
			fv.errors[err.Field] = append(fv.errors[err.Field], err.Message)
		}
	}

	return fv
}

// ValidateFields 批量验证字段
func (fv *FormValidator) ValidateFields(rules map[string]string) *FormValidator {
	for field, rule := range rules {
		fv.ValidateField(field, rule)
	}
	return fv
}

// BindAndValidate 绑定数据并验证
func (fv *FormValidator) BindAndValidate(target any, rules map[string]string) error {
	// 首先验证
	fv.ValidateFields(rules)

	// 如果有验证错误，返回错误
	if fv.HasErrors() {
		return fmt.Errorf("validation failed: %v", fv.errors)
	}

	// 绑定数据到目标结构体
	return fv.bindToStruct(target)
}

// bindToStruct 绑定数据到结构体
func (fv *FormValidator) bindToStruct(target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetType := targetValue.Type()

	for i := 0; i < targetValue.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// 获取字段名
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if parts := strings.Split(jsonTag, ","); parts[0] != "" {
				fieldName = parts[0]
			}
		}
		if formTag := field.Tag.Get("form"); formTag != "" {
			fieldName = formTag
		}

		// 获取表单数据
		data, exists := fv.data[fieldName]
		if !exists {
			continue
		}

		// 设置字段值
		if err := fv.setFieldValue(fieldValue, data); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldName, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值
func (fv *FormValidator) setFieldValue(fieldValue reflect.Value, data any) error {
	dataStr := fmt.Sprintf("%v", data)

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(dataStr)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(dataStr, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(val)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(dataStr, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(val)

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(dataStr, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(val)

	case reflect.Bool:
		val, err := strconv.ParseBool(dataStr)
		if err != nil {
			return err
		}
		fieldValue.SetBool(val)

	case reflect.Slice:
		return fv.setSliceValue(fieldValue, data)

	case reflect.Struct:
		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			return fv.setTimeValue(fieldValue, dataStr)
		}
		return fmt.Errorf("unsupported struct type: %s", fieldValue.Type())

	case reflect.Ptr:
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		return fv.setFieldValue(fieldValue.Elem(), data)

	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}

	return nil
}

// setSliceValue 设置切片值
func (fv *FormValidator) setSliceValue(fieldValue reflect.Value, data any) error {
	var values []string

	switch v := data.(type) {
	case []string:
		values = v
	case string:
		values = []string{v}
	default:
		values = []string{fmt.Sprintf("%v", v)}
	}

	slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

	for i, val := range values {
		elemValue := slice.Index(i)
		if err := fv.setFieldValue(elemValue, val); err != nil {
			return err
		}
	}

	fieldValue.Set(slice)
	return nil
}

// setTimeValue 设置时间值
func (fv *FormValidator) setTimeValue(fieldValue reflect.Value, dataStr string) error {
	// 尝试不同的时间格式
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
		"15:04:05",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dataStr); err == nil {
			fieldValue.Set(reflect.ValueOf(t))
			return nil
		}
	}

	return fmt.Errorf("unable to parse time: %s", dataStr)
}

// HasErrors 检查是否有错误
func (fv *FormValidator) HasErrors() bool {
	return len(fv.errors) > 0
}

// GetErrors 获取所有错误
func (fv *FormValidator) GetErrors() map[string][]string {
	return fv.errors
}

// GetFieldErrors 获取字段错误
func (fv *FormValidator) GetFieldErrors(field string) []string {
	return fv.errors[field]
}

// FirstError 获取第一个错误
func (fv *FormValidator) FirstError() string {
	for _, errors := range fv.errors {
		if len(errors) > 0 {
			return errors[0]
		}
	}
	return ""
}

// ToJSON 转换为JSON
func (fv *FormValidator) ToJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"valid":  !fv.HasErrors(),
		"errors": fv.errors,
	})
}

// ============= 结构体验证绑定 =============

// StructValidator 结构体验证器
type StructValidator struct {
	validator *Validator
	tagName   string
}

// NewStructValidator 创建结构体验证器
func NewStructValidator() *StructValidator {
	return &StructValidator{
		validator: GetDefaultValidator(),
		tagName:   "validate",
	}
}

// SetTagName 设置标签名
func (sv *StructValidator) SetTagName(tagName string) *StructValidator {
	sv.tagName = tagName
	return sv
}

// ValidateStruct 验证结构体
func (sv *StructValidator) ValidateStruct(obj any) *ValidationResult {
	return sv.validator.ValidateStruct(obj)
}

// BindJSON 从JSON绑定并验证
func (sv *StructValidator) BindJSON(jsonData []byte, target any) error {
	// 解析JSON
	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("JSON解析失败: %w", err)
	}

	// 验证结构体
	result := sv.ValidateStruct(target)
	if !result.Valid {
		return ValidationErrors(result.Errors)
	}

	return nil
}

// BindForm 从表单绑定并验证
func (sv *StructValidator) BindForm(data FormData, target any) error {
	fv := NewFormValidator(data)

	// 解析结构体标签获取验证规则
	rules := sv.extractValidationRules(target)

	return fv.BindAndValidate(target, rules)
}

// extractValidationRules 提取验证规则
func (sv *StructValidator) extractValidationRules(obj any) map[string]string {
	rules := make(map[string]string)

	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	if objType.Kind() != reflect.Struct {
		return rules
	}

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)

		// 获取字段名
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if parts := strings.Split(jsonTag, ","); parts[0] != "" {
				fieldName = parts[0]
			}
		}
		if formTag := field.Tag.Get("form"); formTag != "" {
			fieldName = formTag
		}

		// 获取验证规则
		if validateTag := field.Tag.Get(sv.tagName); validateTag != "" {
			rules[fieldName] = validateTag
		}
	}

	return rules
}

// ============= 验证规则构建器 =============

// RuleBuilder 验证规则构建器
type RuleBuilder struct {
	rules []string
}

// NewRuleBuilder 创建规则构建器
func NewRuleBuilder() *RuleBuilder {
	return &RuleBuilder{
		rules: make([]string, 0),
	}
}

// Required 必填
func (rb *RuleBuilder) Required() *RuleBuilder {
	rb.rules = append(rb.rules, "required")
	return rb
}

// Min 最小值
func (rb *RuleBuilder) Min(value float64) *RuleBuilder {
	rb.rules = append(rb.rules, fmt.Sprintf("min:%.2f", value))
	return rb
}

// Max 最大值
func (rb *RuleBuilder) Max(value float64) *RuleBuilder {
	rb.rules = append(rb.rules, fmt.Sprintf("max:%.2f", value))
	return rb
}

// Range 范围
func (rb *RuleBuilder) Range(min, max float64) *RuleBuilder {
	rb.rules = append(rb.rules, fmt.Sprintf("range:%.2f,%.2f", min, max))
	return rb
}

// MinLength 最小长度
func (rb *RuleBuilder) MinLength(length int) *RuleBuilder {
	rb.rules = append(rb.rules, fmt.Sprintf("minLength:%d", length))
	return rb
}

// MaxLength 最大长度
func (rb *RuleBuilder) MaxLength(length int) *RuleBuilder {
	rb.rules = append(rb.rules, fmt.Sprintf("maxLength:%d", length))
	return rb
}

// Length 长度
func (rb *RuleBuilder) Length(length int) *RuleBuilder {
	rb.rules = append(rb.rules, fmt.Sprintf("length:%d", length))
	return rb
}

// Email 邮箱
func (rb *RuleBuilder) Email() *RuleBuilder {
	rb.rules = append(rb.rules, "email")
	return rb
}

// URL URL地址
func (rb *RuleBuilder) URL() *RuleBuilder {
	rb.rules = append(rb.rules, "url")
	return rb
}

// Mobile 手机号
func (rb *RuleBuilder) Mobile() *RuleBuilder {
	rb.rules = append(rb.rules, "mobile")
	return rb
}

// Regex 正则表达式
func (rb *RuleBuilder) Regex(pattern string) *RuleBuilder {
	rb.rules = append(rb.rules, fmt.Sprintf("regex:%s", pattern))
	return rb
}

// In 枚举
func (rb *RuleBuilder) In(values ...string) *RuleBuilder {
	rb.rules = append(rb.rules, fmt.Sprintf("in:%s", strings.Join(values, ",")))
	return rb
}

// Build 构建规则字符串
func (rb *RuleBuilder) Build() string {
	return strings.Join(rb.rules, "|")
}

// ============= 便捷函数 =============

// ValidateForm 验证表单数据
func ValidateForm(data FormData, rules map[string]string) (*FormValidator, error) {
	fv := NewFormValidator(data)
	fv.ValidateFields(rules)

	if fv.HasErrors() {
		return fv, fmt.Errorf("表单验证失败")
	}

	return fv, nil
}

// BindJSON 绑定JSON数据
func BindJSON(jsonData []byte, target any) error {
	return NewStructValidator().BindJSON(jsonData, target)
}

// BindForm 绑定表单数据
func BindForm(data FormData, target any) error {
	return NewStructValidator().BindForm(data, target)
}

// Rules 创建规则构建器（链式调用）
func Rules() *RuleBuilder {
	return NewRuleBuilder()
}
