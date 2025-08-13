package context

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ============= 查询参数类型转换方法 =============

// QueryInt 获取查询参数并转为int
func (ctx *Context) QueryInt(key string) (int, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// QueryIntDefault 获取查询参数并转为int，带默认值
func (ctx *Context) QueryIntDefault(key string, defaultValue int) int {
	if value, ok := parseInt(ctx.Query(key)); ok {
		return value
	}
	return defaultValue
}

// QueryInt64 获取int64类型查询参数
func (ctx *Context) QueryInt64(key string) (int64, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

// QueryInt64Default 获取int64类型查询参数，带默认值
func (ctx *Context) QueryInt64Default(key string, defaultValue int64) int64 {
	if value, ok := parseInt64(ctx.Query(key)); ok {
		return value
	}
	return defaultValue
}

// QueryFloat64 获取float64类型查询参数
func (ctx *Context) QueryFloat64(key string) (float64, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

// QueryFloat64Default 获取float64类型查询参数，带默认值
func (ctx *Context) QueryFloat64Default(key string, defaultValue float64) float64 {
	if value, ok := parseFloat64(ctx.Query(key)); ok {
		return value
	}
	return defaultValue
}

// QueryBool 获取bool类型查询参数
func (ctx *Context) QueryBool(key string) (bool, error) {
	value := ctx.Query(key)
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

// QueryBoolDefault 获取bool类型查询参数，带默认值
func (ctx *Context) QueryBoolDefault(key string, defaultValue bool) bool {
	if value, ok := parseBool(ctx.Query(key)); ok {
		return value
	}
	return defaultValue
}

// ============= 表单参数类型转换方法 =============

// PostFormInt 获取int类型表单参数
func (ctx *Context) PostFormInt(key string) (int, error) {
	value := ctx.PostForm(key)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// PostFormIntDefault 获取int类型表单参数，带默认值
func (ctx *Context) PostFormIntDefault(key string, defaultValue int) int {
	if value, ok := parseInt(ctx.PostForm(key)); ok {
		return value
	}
	return defaultValue
}

// PostFormInt64 获取int64类型表单参数
func (ctx *Context) PostFormInt64(key string) (int64, error) {
	value := ctx.PostForm(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

// PostFormInt64Default 获取int64类型表单参数，带默认值
func (ctx *Context) PostFormInt64Default(key string, defaultValue int64) int64 {
	if value, ok := parseInt64(ctx.PostForm(key)); ok {
		return value
	}
	return defaultValue
}

// PostFormFloat64 获取float64类型表单参数
func (ctx *Context) PostFormFloat64(key string) (float64, error) {
	value := ctx.PostForm(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

// PostFormFloat64Default 获取float64类型表单参数，带默认值
func (ctx *Context) PostFormFloat64Default(key string, defaultValue float64) float64 {
	if value, ok := parseFloat64(ctx.PostForm(key)); ok {
		return value
	}
	return defaultValue
}

// PostFormBool 获取bool类型表单参数
func (ctx *Context) PostFormBool(key string) (bool, error) {
	value := ctx.PostForm(key)
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

// PostFormBoolDefault 获取bool类型表单参数，带默认值
func (ctx *Context) PostFormBoolDefault(key string, defaultValue bool) bool {
	if value, ok := parseBool(ctx.PostForm(key)); ok {
		return value
	}
	return defaultValue
}

// ============= 通用参数获取方法 =============

// GetParam 通用字符串参数获取 (路由参数 -> 查询参数 -> 表单参数)
func (ctx *Context) GetParam(key string) string {
	// 优先级：路由参数 > 查询参数 > 表单参数
	if value := ctx.Param(key); value != "" {
		return value
	}
	if value := ctx.Query(key); value != "" {
		return value
	}
	return ctx.PostForm(key)
}

// GetParamDefault 通用字符串参数获取，带默认值
func (ctx *Context) GetParamDefault(key, defaultValue string) string {
	return parseValueWithDefault(ctx.GetParam(key), defaultValue)
}

// GetParamInt 通用int参数获取
func (ctx *Context) GetParamInt(key string) (int, error) {
	value := ctx.GetParam(key)
	if value == "" {
		return 0, ErrParamNotFound
	}
	return strconv.Atoi(value)
}

// GetParamIntDefault 通用int参数获取，带默认值
func (ctx *Context) GetParamIntDefault(key string, defaultValue int) int {
	if value, ok := parseInt(ctx.GetParam(key)); ok {
		return value
	}
	return defaultValue
}

// GetParamFloat 通用float64参数获取
func (ctx *Context) GetParamFloat(key string) (float64, error) {
	value := ctx.GetParam(key)
	if value == "" {
		return 0, ErrParamNotFound
	}
	return strconv.ParseFloat(value, 64)
}

// GetParamFloatDefault 通用float64参数获取，带默认值
func (ctx *Context) GetParamFloatDefault(key string, defaultValue float64) float64 {
	if value, ok := parseFloat64(ctx.GetParam(key)); ok {
		return value
	}
	return defaultValue
}

// GetParamBool 通用bool参数获取
func (ctx *Context) GetParamBool(key string) (bool, error) {
	value := ctx.GetParam(key)
	if value == "" {
		return false, ErrParamNotFound
	}
	return strconv.ParseBool(value)
}

// GetParamBoolDefault 通用bool参数获取，带默认值
func (ctx *Context) GetParamBoolDefault(key string, defaultValue bool) bool {
	if value, ok := parseBool(ctx.GetParam(key)); ok {
		return value
	}
	return defaultValue
}

// ============= 参数存在性检查方法 =============

// HasQuery 检查查询参数是否存在
func (ctx *Context) HasQuery(key string) bool {
	return ctx.Query(key) != ""
}

// HasPostForm 检查表单参数是否存在
func (ctx *Context) HasPostForm(key string) bool {
	return ctx.PostForm(key) != ""
}

// HasParam 检查路由参数是否存在
func (ctx *Context) HasParam(key string) bool {
	return ctx.Param(key) != ""
}

// ============= 高级参数处理方法 =============

// GetForm 获取表单值（兼容beego风格）
func (ctx *Context) GetForm(key string) string {
	return ctx.FormValue(key)
}

// GetFormDefault 获取表单值，带默认值（兼容beego风格）
func (ctx *Context) GetFormDefault(key, defaultValue string) string {
	return ctx.FormValueDefault(key, defaultValue)
}

// ============= 增强Params处理（智能验证） =============

// ValidateRequired 验证必需参数
func (ctx *Context) ValidateRequired(keys ...string) error {
	var missingKeys []string

	for _, key := range keys {
		if !ctx.HasParam(key) && !ctx.HasQuery(key) && !ctx.HasPostForm(key) {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		return &ValidationError{
			Type:        "required",
			MissingKeys: missingKeys,
			Message:     "Required parameters missing: " + strings.Join(missingKeys, ", "),
		}
	}

	return nil
}

// ValidateInt 验证整数参数范围
func (ctx *Context) ValidateInt(key string, min, max int) error {
	value, err := ctx.GetParamInt(key)
	if err != nil {
		return &ValidationError{
			Type:    "type",
			Key:     key,
			Value:   ctx.GetParam(key),
			Message: "Parameter must be an integer",
		}
	}

	if value < min || value > max {
		return &ValidationError{
			Type:    "range",
			Key:     key,
			Value:   strconv.Itoa(value),
			Message: "Parameter must be between " + strconv.Itoa(min) + " and " + strconv.Itoa(max),
		}
	}

	return nil
}

// ValidateFloat 验证浮点数参数范围
func (ctx *Context) ValidateFloat(key string, min, max float64) error {
	value, err := ctx.GetParamFloat(key)
	if err != nil {
		return &ValidationError{
			Type:    "type",
			Key:     key,
			Value:   ctx.GetParam(key),
			Message: "Parameter must be a number",
		}
	}

	if value < min || value > max {
		return &ValidationError{
			Type:    "range",
			Key:     key,
			Value:   strconv.FormatFloat(value, 'f', -1, 64),
			Message: "Parameter must be between " + strconv.FormatFloat(min, 'f', -1, 64) + " and " + strconv.FormatFloat(max, 'f', -1, 64),
		}
	}

	return nil
}

// ValidateLength 验证字符串参数长度
func (ctx *Context) ValidateLength(key string, minLen, maxLen int) error {
	value := ctx.GetParam(key)
	length := len(value)

	if length < minLen || length > maxLen {
		return &ValidationError{
			Type:    "length",
			Key:     key,
			Value:   value,
			Message: "Parameter length must be between " + strconv.Itoa(minLen) + " and " + strconv.Itoa(maxLen),
		}
	}

	return nil
}

// ValidateEnum 验证枚举参数
func (ctx *Context) ValidateEnum(key string, allowedValues ...string) error {
	value := ctx.GetParam(key)
	if value == "" {
		return nil // 空值跳过验证
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return &ValidationError{
		Type:    "enum",
		Key:     key,
		Value:   value,
		Message: "Parameter must be one of: " + strings.Join(allowedValues, ", "),
	}
}

// ValidatePattern 验证正则表达式模式
func (ctx *Context) ValidatePattern(key, pattern, message string) error {
	value := ctx.GetParam(key)
	if value == "" {
		return nil // 空值跳过验证
	}

	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return &ValidationError{
			Type:    "pattern",
			Key:     key,
			Value:   value,
			Message: "Invalid pattern: " + err.Error(),
		}
	}

	if !matched {
		if message == "" {
			message = "Parameter format is invalid"
		}
		return &ValidationError{
			Type:    "pattern",
			Key:     key,
			Value:   value,
			Message: message,
		}
	}

	return nil
}

// ============= 增强Params处理（批量处理） =============

// GetAllParams 获取所有参数
func (ctx *Context) GetAllParams() *ParamCollection {
	collection := &ParamCollection{
		Route: make(map[string]string),
		Query: make(map[string]string),
		Form:  make(map[string]string),
	}

	// 路由参数
	for _, param := range ctx.params {
		collection.Route[param.Key] = param.Value
	}

	// 查询参数
	queryMap := ctx.QueryMap()
	for key, values := range queryMap {
		if len(values) > 0 {
			collection.Query[key] = values[0] // 取第一个值
		}
	}

	// 表单参数
	if ctx.ensureRequest() {
		ctx.request.PostArgs().VisitAll(func(key, value []byte) {
			collection.Form[safeStringConvert(key)] = safeStringConvert(value)
		})
	}

	return collection
}

// ValidateParams 批量验证参数
func (ctx *Context) ValidateParams(rules map[string]*ValidationRule) *ValidationErrors {
	errors := &ValidationErrors{
		Errors: make([]ValidationError, 0),
	}

	for key, rule := range rules {
		// 检查必需参数
		if rule.Required && !ctx.hasAnyParam(key) {
			errors.Add(ValidationError{
				Type:    "required",
				Key:     key,
				Message: "Parameter is required",
			})
			continue
		}

		// 如果参数不存在且非必需，跳过验证
		if !ctx.hasAnyParam(key) {
			continue
		}

		// 类型验证
		value := ctx.GetParam(key)
		if err := ctx.validateType(key, value, rule); err != nil {
			errors.Add(*err)
			continue
		}

		// 范围验证
		if err := ctx.validateRange(key, value, rule); err != nil {
			errors.Add(*err)
		}

		// 长度验证
		if err := ctx.validateStringLength(key, value, rule); err != nil {
			errors.Add(*err)
		}

		// 枚举验证
		if err := ctx.validateEnumValue(key, value, rule); err != nil {
			errors.Add(*err)
		}

		// 自定义验证
		if rule.CustomValidator != nil {
			if err := rule.CustomValidator(key, value); err != nil {
				errors.Add(ValidationError{
					Type:    "custom",
					Key:     key,
					Value:   value,
					Message: err.Error(),
				})
			}
		}
	}

	if len(errors.Errors) == 0 {
		return nil
	}

	return errors
}

// MapToStruct 将参数映射到结构体
func (ctx *Context) MapToStruct(dst any, options *MappingOptions) error {
	if options == nil {
		options = &MappingOptions{
			TagName:       "param",
			IgnoreUnknown: true,
			CaseSensitive: false,
		}
	}

	return mapParamsToStruct(ctx, dst, options)
}

// ============= 增强Params处理（格式化支持） =============

// ParseTime 解析时间参数
func (ctx *Context) ParseTime(key string, layout string) (time.Time, error) {
	value := ctx.GetParam(key)
	if value == "" {
		return time.Time{}, ErrParamNotFound
	}

	return time.Parse(layout, value)
}

// ParseTimeDefault 解析时间参数（带默认值）
func (ctx *Context) ParseTimeDefault(key string, layout string, defaultValue time.Time) time.Time {
	if t, err := ctx.ParseTime(key, layout); err == nil {
		return t
	}
	return defaultValue
}

// ParseJSON 解析JSON参数
func (ctx *Context) ParseJSON(key string, dst any) error {
	value := ctx.GetParam(key)
	if value == "" {
		return ErrParamNotFound
	}

	return json.Unmarshal([]byte(value), dst)
}

// ParseSlice 解析数组参数（逗号分隔）
func (ctx *Context) ParseSlice(key string, separator string) []string {
	value := ctx.GetParam(key)
	if value == "" {
		return nil
	}

	if separator == "" {
		separator = ","
	}

	parts := strings.Split(value, separator)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// ParseIntSlice 解析整数数组参数
func (ctx *Context) ParseIntSlice(key string, separator string) ([]int, error) {
	strSlice := ctx.ParseSlice(key, separator)
	if len(strSlice) == 0 {
		return nil, nil
	}

	result := make([]int, 0, len(strSlice))
	for _, str := range strSlice {
		if intVal, err := strconv.Atoi(str); err == nil {
			result = append(result, intVal)
		} else {
			return nil, &ValidationError{
				Type:    "type",
				Key:     key,
				Value:   str,
				Message: "Invalid integer in array: " + str,
			}
		}
	}

	return result, nil
}

// ============= 辅助类型定义 =============

// ValidationError 验证错误
type ValidationError struct {
	Type        string   `json:"type"`
	Key         string   `json:"key,omitempty"`
	Value       string   `json:"value,omitempty"`
	MissingKeys []string `json:"missing_keys,omitempty"`
	Message     string   `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidationErrors 验证错误集合
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Add 添加验证错误
func (ve *ValidationErrors) Add(err ValidationError) {
	ve.Errors = append(ve.Errors, err)
}

// Error 实现error接口
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "No validation errors"
	}

	messages := make([]string, len(ve.Errors))
	for i, err := range ve.Errors {
		messages[i] = err.Message
	}

	return "Validation failed: " + strings.Join(messages, "; ")
}

// HasErrors 是否有错误
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// ValidationRule 验证规则
type ValidationRule struct {
	Required        bool                          `json:"required"`
	Type            string                        `json:"type"` // string, int, float, bool
	MinInt          *int                          `json:"min_int,omitempty"`
	MaxInt          *int                          `json:"max_int,omitempty"`
	MinFloat        *float64                      `json:"min_float,omitempty"`
	MaxFloat        *float64                      `json:"max_float,omitempty"`
	MinLength       *int                          `json:"min_length,omitempty"`
	MaxLength       *int                          `json:"max_length,omitempty"`
	AllowedValues   []string                      `json:"allowed_values,omitempty"`
	Pattern         string                        `json:"pattern,omitempty"`
	CustomValidator func(key, value string) error `json:"-"`
}

// ParamCollection 参数集合
type ParamCollection struct {
	Route map[string]string `json:"route"`
	Query map[string]string `json:"query"`
	Form  map[string]string `json:"form"`
}

// Get 从集合中获取参数（按优先级：路由 > 查询 > 表单）
func (pc *ParamCollection) Get(key string) (string, bool) {
	if value, ok := pc.Route[key]; ok {
		return value, true
	}
	if value, ok := pc.Query[key]; ok {
		return value, true
	}
	if value, ok := pc.Form[key]; ok {
		return value, true
	}
	return "", false
}

// MappingOptions 映射选项
type MappingOptions struct {
	TagName       string `json:"tag_name"`
	IgnoreUnknown bool   `json:"ignore_unknown"`
	CaseSensitive bool   `json:"case_sensitive"`
}

// ============= 错误定义 =============

var (
	ErrParamNotFound = &ContextError{Code: "PARAM_NOT_FOUND", Message: "Parameter not found"}
)

// ============= 辅助函数 =============

// hasAnyParam 检查参数是否存在于任意位置
func (ctx *Context) hasAnyParam(key string) bool {
	return ctx.HasParam(key) || ctx.HasQuery(key) || ctx.HasPostForm(key)
}

// validateType 验证参数类型
func (ctx *Context) validateType(key, value string, rule *ValidationRule) *ValidationError {
	switch rule.Type {
	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return &ValidationError{
				Type:    "type",
				Key:     key,
				Value:   value,
				Message: "Parameter must be an integer",
			}
		}
	case "float":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return &ValidationError{
				Type:    "type",
				Key:     key,
				Value:   value,
				Message: "Parameter must be a number",
			}
		}
	case "bool":
		if _, err := strconv.ParseBool(value); err != nil {
			return &ValidationError{
				Type:    "type",
				Key:     key,
				Value:   value,
				Message: "Parameter must be a boolean",
			}
		}
	}
	return nil
}

// validateRange 验证参数范围
func (ctx *Context) validateRange(key, value string, rule *ValidationRule) *ValidationError {
	switch rule.Type {
	case "int":
		if intVal, err := strconv.Atoi(value); err == nil {
			if rule.MinInt != nil && intVal < *rule.MinInt {
				return &ValidationError{
					Type:    "range",
					Key:     key,
					Value:   value,
					Message: "Parameter must be at least " + strconv.Itoa(*rule.MinInt),
				}
			}
			if rule.MaxInt != nil && intVal > *rule.MaxInt {
				return &ValidationError{
					Type:    "range",
					Key:     key,
					Value:   value,
					Message: "Parameter must be at most " + strconv.Itoa(*rule.MaxInt),
				}
			}
		}
	case "float":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			if rule.MinFloat != nil && floatVal < *rule.MinFloat {
				return &ValidationError{
					Type:    "range",
					Key:     key,
					Value:   value,
					Message: "Parameter must be at least " + strconv.FormatFloat(*rule.MinFloat, 'f', -1, 64),
				}
			}
			if rule.MaxFloat != nil && floatVal > *rule.MaxFloat {
				return &ValidationError{
					Type:    "range",
					Key:     key,
					Value:   value,
					Message: "Parameter must be at most " + strconv.FormatFloat(*rule.MaxFloat, 'f', -1, 64),
				}
			}
		}
	}
	return nil
}

// validateStringLength 验证字符串长度
func (ctx *Context) validateStringLength(key, value string, rule *ValidationRule) *ValidationError {
	length := len(value)

	if rule.MinLength != nil && length < *rule.MinLength {
		return &ValidationError{
			Type:    "length",
			Key:     key,
			Value:   value,
			Message: "Parameter must be at least " + strconv.Itoa(*rule.MinLength) + " characters",
		}
	}

	if rule.MaxLength != nil && length > *rule.MaxLength {
		return &ValidationError{
			Type:    "length",
			Key:     key,
			Value:   value,
			Message: "Parameter must be at most " + strconv.Itoa(*rule.MaxLength) + " characters",
		}
	}

	return nil
}

// validateEnumValue 验证枚举值
func (ctx *Context) validateEnumValue(key, value string, rule *ValidationRule) *ValidationError {
	if len(rule.AllowedValues) == 0 {
		return nil
	}

	for _, allowed := range rule.AllowedValues {
		if value == allowed {
			return nil
		}
	}

	return &ValidationError{
		Type:    "enum",
		Key:     key,
		Value:   value,
		Message: "Parameter must be one of: " + strings.Join(rule.AllowedValues, ", "),
	}
}

// mapParamsToStruct 将参数映射到结构体
func mapParamsToStruct(ctx *Context, dst any, options *MappingOptions) error {
	// 这是一个简化实现，实际项目中应该使用更完整的反射映射
	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr || dstValue.Elem().Kind() != reflect.Struct {
		return &ValidationError{
			Type:    "mapping",
			Message: "Destination must be a pointer to struct",
		}
	}

	// 获取所有参数
	params := ctx.GetAllParams()

	// 映射到结构体字段
	structValue := dstValue.Elem()
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// 跳过未导出的字段
		if !field.CanSet() {
			continue
		}

		// 获取标签名
		tagName := fieldType.Tag.Get(options.TagName)
		if tagName == "" {
			if options.IgnoreUnknown {
				continue
			}
			tagName = fieldType.Name
		}

		// 获取参数值
		if value, ok := params.Get(tagName); ok && value != "" {
			// 设置字段值
			if err := setFieldValue(field, value); err != nil {
				return &ValidationError{
					Type:    "mapping",
					Key:     tagName,
					Value:   value,
					Message: "Failed to set field " + fieldType.Name + ": " + err.Error(),
				}
			}
		}
	}

	return nil
}

// setFieldValue 设置字段值
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(intVal)
		} else {
			return err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(uintVal)
		} else {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, field.Type().Bits()); err == nil {
			field.SetFloat(floatVal)
		} else {
			return err
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolVal)
		} else {
			return err
		}
	default:
		return &ValidationError{
			Type:    "mapping",
			Message: "Unsupported field type: " + field.Kind().String(),
		}
	}
	return nil
}
