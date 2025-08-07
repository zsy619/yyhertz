package binding

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParameterValidator 参数验证器
type ParameterValidator struct {
	validators map[reflect.Type]ParameterValidatorFunc // 类型验证器映射
	rules      map[string]ValidationRule               // 验证规则映射
}

// ValidationRule 验证规则
type ValidationRule interface {
	Validate(value interface{}, param string) error
	Name() string
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string      // 字段名
	Tag     string      // 验证标签
	Value   interface{} // 实际值
	Param   string      // 验证参数
	Message string      // 错误消息
}

// NewParameterValidator 创建参数验证器
func NewParameterValidator() *ParameterValidator {
	validator := &ParameterValidator{
		validators: make(map[reflect.Type]ParameterValidatorFunc),
		rules:      make(map[string]ValidationRule),
	}

	// 注册内置验证规则
	validator.registerBuiltinRules()

	return validator
}

// registerBuiltinRules 注册内置验证规则
func (pv *ParameterValidator) registerBuiltinRules() {
	// 基础验证规则
	pv.rules["required"] = &RequiredRule{}
	pv.rules["min"] = &MinRule{}
	pv.rules["max"] = &MaxRule{}
	pv.rules["len"] = &LenRule{}
	pv.rules["email"] = &EmailRule{}
	pv.rules["url"] = &URLRule{}
	pv.rules["numeric"] = &NumericRule{}
	pv.rules["alpha"] = &AlphaRule{}
	pv.rules["alphanum"] = &AlphaNumRule{}
	pv.rules["regexp"] = &RegexpRule{}
	pv.rules["oneof"] = &OneOfRule{}
	pv.rules["range"] = &RangeRule{}
	pv.rules["datetime"] = &DateTimeRule{}
}

// GetValidator 获取参数验证器
func (pv *ParameterValidator) GetValidator(paramType reflect.Type) ParameterValidatorFunc {
	if validator, exists := pv.validators[paramType]; exists {
		return validator
	}

	// 返回通用验证器
	return func(value interface{}, param *ParamBinder) error {
		return pv.ValidateValue(value, param.Tags)
	}
}

// ValidateValue 验证值
func (pv *ParameterValidator) ValidateValue(value interface{}, tags map[string]string) error {
	for tag, param := range tags {
		if rule, exists := pv.rules[tag]; exists {
			if err := rule.Validate(value, param); err != nil {
				return &ValidationError{
					Tag:     tag,
					Value:   value,
					Param:   param,
					Message: err.Error(),
				}
			}
		}
	}
	return nil
}

// ValidateStruct 验证结构体
func (pv *ParameterValidator) ValidateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 跳过非导出字段
		if !fieldValue.CanInterface() {
			continue
		}

		// 解析验证标签
		validateTag := field.Tag.Get("validate")
		if validateTag == "" || validateTag == "-" {
			continue
		}

		// 验证字段
		if err := pv.validateField(field.Name, fieldValue.Interface(), validateTag); err != nil {
			if validationErr, ok := err.(*ValidationError); ok {
				validationErr.Field = field.Name
			}
			return err
		}
	}

	return nil
}

// validateField 验证字段
func (pv *ParameterValidator) validateField(fieldName string, value interface{}, validateTag string) error {
	rules := strings.Split(validateTag, ",")

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		// 解析规则和参数
		parts := strings.Split(rule, "=")
		ruleName := parts[0]
		param := ""
		if len(parts) > 1 {
			param = parts[1]
		}

		// 执行验证
		if validator, exists := pv.rules[ruleName]; exists {
			if err := validator.Validate(value, param); err != nil {
				return &ValidationError{
					Field:   fieldName,
					Tag:     ruleName,
					Value:   value,
					Param:   param,
					Message: err.Error(),
				}
			}
		}
	}

	return nil
}

// RegisterRule 注册验证规则
func (pv *ParameterValidator) RegisterRule(name string, rule ValidationRule) {
	pv.rules[name] = rule
}

// 内置验证规则实现

// RequiredRule 必填验证规则
type RequiredRule struct{}

func (r *RequiredRule) Name() string { return "required" }

func (r *RequiredRule) Validate(value interface{}, param string) error {
	if value == nil {
		return fmt.Errorf("field is required")
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return fmt.Errorf("field is required")
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if v.Len() == 0 {
			return fmt.Errorf("field is required")
		}
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return fmt.Errorf("field is required")
		}
	}

	return nil
}

// MinRule 最小值验证规则
type MinRule struct{}

func (r *MinRule) Name() string { return "min" }

func (r *MinRule) Validate(value interface{}, param string) error {
	min, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return fmt.Errorf("invalid min parameter: %s", param)
	}

	switch v := value.(type) {
	case string:
		if float64(len(v)) < min {
			return fmt.Errorf("field must be at least %s characters long", param)
		}
	case int:
		if float64(v) < min {
			return fmt.Errorf("field must be at least %s", param)
		}
	case int64:
		if float64(v) < min {
			return fmt.Errorf("field must be at least %s", param)
		}
	case float64:
		if v < min {
			return fmt.Errorf("field must be at least %s", param)
		}
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			if float64(rv.Len()) < min {
				return fmt.Errorf("field must have at least %s items", param)
			}
		}
	}

	return nil
}

// MaxRule 最大值验证规则
type MaxRule struct{}

func (r *MaxRule) Name() string { return "max" }

func (r *MaxRule) Validate(value interface{}, param string) error {
	max, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return fmt.Errorf("invalid max parameter: %s", param)
	}

	switch v := value.(type) {
	case string:
		if float64(len(v)) > max {
			return fmt.Errorf("field must be at most %s characters long", param)
		}
	case int:
		if float64(v) > max {
			return fmt.Errorf("field must be at most %s", param)
		}
	case int64:
		if float64(v) > max {
			return fmt.Errorf("field must be at most %s", param)
		}
	case float64:
		if v > max {
			return fmt.Errorf("field must be at most %s", param)
		}
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			if float64(rv.Len()) > max {
				return fmt.Errorf("field must have at most %s items", param)
			}
		}
	}

	return nil
}

// LenRule 长度验证规则
type LenRule struct{}

func (r *LenRule) Name() string { return "len" }

func (r *LenRule) Validate(value interface{}, param string) error {
	length, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("invalid len parameter: %s", param)
	}

	switch v := value.(type) {
	case string:
		if len(v) != length {
			return fmt.Errorf("field must be exactly %s characters long", param)
		}
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			if rv.Len() != length {
				return fmt.Errorf("field must have exactly %s items", param)
			}
		}
	}

	return nil
}

// EmailRule 邮箱验证规则
type EmailRule struct{}

func (r *EmailRule) Name() string { return "email" }

func (r *EmailRule) Validate(value interface{}, param string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("field must be a valid email address")
	}

	return nil
}

// URLRule URL验证规则
type URLRule struct{}

func (r *URLRule) Name() string { return "url" }

func (r *URLRule) Validate(value interface{}, param string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	urlRegex := regexp.MustCompile(`^(https?|ftp)://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(str) {
		return fmt.Errorf("field must be a valid URL")
	}

	return nil
}

// NumericRule 数字验证规则
type NumericRule struct{}

func (r *NumericRule) Name() string { return "numeric" }

func (r *NumericRule) Validate(value interface{}, param string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	if _, err := strconv.ParseFloat(str, 64); err != nil {
		return fmt.Errorf("field must be a valid number")
	}

	return nil
}

// AlphaRule 字母验证规则
type AlphaRule struct{}

func (r *AlphaRule) Name() string { return "alpha" }

func (r *AlphaRule) Validate(value interface{}, param string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(str) {
		return fmt.Errorf("field must contain only letters")
	}

	return nil
}

// AlphaNumRule 字母数字验证规则
type AlphaNumRule struct{}

func (r *AlphaNumRule) Name() string { return "alphanum" }

func (r *AlphaNumRule) Validate(value interface{}, param string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphaNumRegex.MatchString(str) {
		return fmt.Errorf("field must contain only letters and numbers")
	}

	return nil
}

// RegexpRule 正则表达式验证规则
type RegexpRule struct{}

func (r *RegexpRule) Name() string { return "regexp" }

func (r *RegexpRule) Validate(value interface{}, param string) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	regex, err := regexp.Compile(param)
	if err != nil {
		return fmt.Errorf("invalid regexp parameter: %s", param)
	}

	if !regex.MatchString(str) {
		return fmt.Errorf("field must match pattern %s", param)
	}

	return nil
}

// OneOfRule 枚举值验证规则
type OneOfRule struct{}

func (r *OneOfRule) Name() string { return "oneof" }

func (r *OneOfRule) Validate(value interface{}, param string) error {
	str := fmt.Sprintf("%v", value)
	values := strings.Split(param, " ")

	for _, v := range values {
		if str == v {
			return nil
		}
	}

	return fmt.Errorf("field must be one of [%s]", param)
}

// RangeRule 范围验证规则
type RangeRule struct{}

func (r *RangeRule) Name() string { return "range" }

func (r *RangeRule) Validate(value interface{}, param string) error {
	parts := strings.Split(param, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range parameter: %s", param)
	}

	min, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid range min: %s", parts[0])
	}

	max, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid range max: %s", parts[1])
	}

	var val float64
	switch v := value.(type) {
	case int:
		val = float64(v)
	case int64:
		val = float64(v)
	case float64:
		val = v
	case float32:
		val = float64(v)
	default:
		return fmt.Errorf("field must be a number")
	}

	if val < min || val > max {
		return fmt.Errorf("field must be between %s and %s", parts[0], parts[1])
	}

	return nil
}

// DateTimeRule 日期时间验证规则
type DateTimeRule struct{}

func (r *DateTimeRule) Name() string { return "datetime" }

func (r *DateTimeRule) Validate(value interface{}, param string) error {
	str, ok := value.(string)
	if !ok {
		if _, ok := value.(time.Time); ok {
			return nil
		}
		return fmt.Errorf("field must be a string or time.Time")
	}

	format := time.RFC3339
	if param != "" {
		format = param
	}

	if _, err := time.Parse(format, str); err != nil {
		return fmt.Errorf("field must be a valid datetime in format %s", format)
	}

	return nil
}

// Error 实现error接口
func (ve ValidationError) Error() string {
	if ve.Field != "" {
		return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
	}
	return fmt.Sprintf("validation failed: %s", ve.Message)
}

// CustomValidationRule 自定义验证规则接口
type CustomValidationRule interface {
	ValidationRule
	SetParams(params map[string]string)
}

// ConditionalValidationRule 条件验证规则接口
type ConditionalValidationRule interface {
	ValidationRule
	ShouldValidate(structValue reflect.Value, fieldName string) bool
}
