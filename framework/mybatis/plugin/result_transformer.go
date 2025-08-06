// Package plugin 结果转换插件实现
//
// 提供查询结果自动转换功能，支持多种数据格式转换和自定义转换器
package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ResultTransformerPlugin 结果转换插件
type ResultTransformerPlugin struct {
	*BasePlugin
	enableTransform bool
	transformers    map[string]ResultTransformer
	rules           map[string]TransformRule
}

// ResultTransformer 结果转换器接口
type ResultTransformer interface {
	Transform(value any, rule TransformRule) (any, error)
	GetName() string
	CanTransform(fromType, toType reflect.Type) bool
}

// TransformRule 转换规则
type TransformRule struct {
	FromType   string         // 源类型
	ToType     string         // 目标类型
	Method     string         // 转换方法
	Parameters map[string]any // 转换参数
	Condition  func(any) bool // 转换条件
}

// TransformResult 转换结果
type TransformResult struct {
	Success     bool   // 是否成功
	Value       any    // 转换后的值
	Error       error  // 错误信息
	Transformer string // 使用的转换器
}

// NewResultTransformerPlugin 创建结果转换插件
func NewResultTransformerPlugin() *ResultTransformerPlugin {
	plugin := &ResultTransformerPlugin{
		BasePlugin:      NewBasePlugin("result_transformer", 6),
		enableTransform: true,
		transformers:    make(map[string]ResultTransformer),
		rules:           make(map[string]TransformRule),
	}

	// 注册默认转换器
	plugin.registerDefaultTransformers()

	return plugin
}

// Intercept 拦截方法调用
func (plugin *ResultTransformerPlugin) Intercept(invocation *Invocation) (any, error) {
	// 执行原方法
	result, err := invocation.Proceed()
	if err != nil {
		return result, err
	}

	if !plugin.enableTransform || result == nil {
		return result, err
	}

	// 转换结果
	transformedResult := plugin.transformResult(invocation.Method.Name, result)
	return transformedResult, err
}

// Plugin 包装目标对象
func (plugin *ResultTransformerPlugin) Plugin(target any) any {
	return target
}

// SetProperties 设置插件属性
func (plugin *ResultTransformerPlugin) SetProperties(properties map[string]any) {
	plugin.BasePlugin.SetProperties(properties)

	plugin.enableTransform = plugin.GetPropertyBool("enableTransform", true)
}

// transformResult 转换结果
func (plugin *ResultTransformerPlugin) transformResult(methodName string, result any) any {
	rule, exists := plugin.rules[methodName]
	if !exists {
		return result // 没有转换规则
	}

	// 检查转换条件
	if rule.Condition != nil && !rule.Condition(result) {
		return result
	}

	// 执行转换
	transformer, exists := plugin.transformers[rule.Method]
	if !exists {
		return result // 没有对应的转换器
	}

	transformed, err := transformer.Transform(result, rule)
	if err != nil {
		// 转换失败，返回原结果
		return result
	}

	return transformed
}

// registerDefaultTransformers 注册默认转换器
func (plugin *ResultTransformerPlugin) registerDefaultTransformers() {
	plugin.RegisterTransformer(&MapTransformer{})
	plugin.RegisterTransformer(&JsonTransformer{})
	plugin.RegisterTransformer(&StringTransformer{})
	plugin.RegisterTransformer(&NumberTransformer{})
	plugin.RegisterTransformer(&TimeTransformer{})
	plugin.RegisterTransformer(&CamelCaseTransformer{})
	plugin.RegisterTransformer(&SnakeCaseTransformer{})
}

// RegisterTransformer 注册转换器
func (plugin *ResultTransformerPlugin) RegisterTransformer(transformer ResultTransformer) {
	plugin.transformers[transformer.GetName()] = transformer
}

// AddRule 添加转换规则
func (plugin *ResultTransformerPlugin) AddRule(methodName string, rule TransformRule) {
	plugin.rules[methodName] = rule
}

// 内置转换器实现

// MapTransformer Map转换器
type MapTransformer struct{}

func (t *MapTransformer) GetName() string { return "map" }

func (t *MapTransformer) CanTransform(fromType, toType reflect.Type) bool {
	return fromType.Kind() == reflect.Struct || fromType.Kind() == reflect.Map
}

func (t *MapTransformer) Transform(value any, rule TransformRule) (any, error) {
	if value == nil {
		return nil, nil
	}

	// 如果已经是map，直接返回
	if m, ok := value.(map[string]any); ok {
		return m, nil
	}

	// 如果是切片，转换每个元素
	if slice, ok := value.([]any); ok {
		result := make([]map[string]any, len(slice))
		for i, item := range slice {
			if itemMap, err := t.structToMap(item); err == nil {
				result[i] = itemMap
			}
		}
		return result, nil
	}

	// 结构体转map
	return t.structToMap(value)
}

func (t *MapTransformer) structToMap(obj any) (map[string]any, error) {
	if obj == nil {
		return nil, nil
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("不是结构体类型")
	}

	result := make(map[string]any)
	structType := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		// 获取字段名（支持json tag）
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if tagName := strings.Split(jsonTag, ",")[0]; tagName != "" && tagName != "-" {
				fieldName = tagName
			}
		}

		result[fieldName] = fieldValue.Interface()
	}

	return result, nil
}

// JsonTransformer JSON转换器
type JsonTransformer struct{}

func (t *JsonTransformer) GetName() string { return "json" }

func (t *JsonTransformer) CanTransform(fromType, toType reflect.Type) bool {
	return true // JSON可以转换任何类型
}

func (t *JsonTransformer) Transform(value any, rule TransformRule) (any, error) {
	if value == nil {
		return nil, nil
	}

	// 转换为JSON字符串
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 如果目标类型是字符串，直接返回
	if rule.ToType == "string" {
		return string(jsonBytes), nil
	}

	// 否则返回格式化的JSON
	var result any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("JSON反序列化失败: %v", err)
	}

	return result, nil
}

// StringTransformer 字符串转换器
type StringTransformer struct{}

func (t *StringTransformer) GetName() string { return "string" }

func (t *StringTransformer) CanTransform(fromType, toType reflect.Type) bool {
	return toType.Kind() == reflect.String
}

func (t *StringTransformer) Transform(value any, rule TransformRule) (any, error) {
	if value == nil {
		return "", nil
	}

	return fmt.Sprintf("%v", value), nil
}

// NumberTransformer 数字转换器
type NumberTransformer struct{}

func (t *NumberTransformer) GetName() string { return "number" }

func (t *NumberTransformer) CanTransform(fromType, toType reflect.Type) bool {
	return t.isNumericType(toType)
}

func (t *NumberTransformer) Transform(value any, rule TransformRule) (any, error) {
	if value == nil {
		return 0, nil
	}

	str := fmt.Sprintf("%v", value)

	switch rule.ToType {
	case "int", "int32":
		if i, err := strconv.Atoi(str); err == nil {
			return i, nil
		}
	case "int64":
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i, nil
		}
	case "float32":
		if f, err := strconv.ParseFloat(str, 32); err == nil {
			return float32(f), nil
		}
	case "float64":
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f, nil
		}
	}

	return nil, fmt.Errorf("无法转换为数字类型: %s", rule.ToType)
}

func (t *NumberTransformer) isNumericType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// TimeTransformer 时间转换器
type TimeTransformer struct{}

func (t *TimeTransformer) GetName() string { return "time" }

func (t *TimeTransformer) CanTransform(fromType, toType reflect.Type) bool {
	return toType == reflect.TypeOf(time.Time{}) || fromType == reflect.TypeOf(time.Time{})
}

func (t *TimeTransformer) Transform(value any, rule TransformRule) (any, error) {
	if value == nil {
		return time.Time{}, nil
	}

	// 时间转字符串
	if t, ok := value.(time.Time); ok {
		format := "2006-01-02 15:04:05"
		if f, exists := rule.Parameters["format"]; exists {
			if formatStr, ok := f.(string); ok {
				format = formatStr
			}
		}
		return t.Format(format), nil
	}

	// 字符串转时间
	if str, ok := value.(string); ok {
		format := "2006-01-02 15:04:05"
		if f, exists := rule.Parameters["format"]; exists {
			if formatStr, ok := f.(string); ok {
				format = formatStr
			}
		}
		return time.Parse(format, str)
	}

	return nil, fmt.Errorf("无法转换时间类型")
}

// CamelCaseTransformer 驼峰命名转换器
type CamelCaseTransformer struct{}

func (t *CamelCaseTransformer) GetName() string { return "camelCase" }

func (t *CamelCaseTransformer) CanTransform(fromType, toType reflect.Type) bool {
	return fromType.Kind() == reflect.Map || fromType.Kind() == reflect.Struct
}

func (t *CamelCaseTransformer) Transform(value any, rule TransformRule) (any, error) {
	if value == nil {
		return nil, nil
	}

	// 处理map
	if m, ok := value.(map[string]any); ok {
		result := make(map[string]any)
		for k, v := range m {
			camelKey := t.toCamelCase(k)
			result[camelKey] = v
		}
		return result, nil
	}

	// 处理切片
	if slice, ok := value.([]any); ok {
		result := make([]any, len(slice))
		for i, item := range slice {
			if transformed, err := t.Transform(item, rule); err == nil {
				result[i] = transformed
			} else {
				result[i] = item
			}
		}
		return result, nil
	}

	return value, nil
}

func (t *CamelCaseTransformer) toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) <= 1 {
		return s
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return result
}

// SnakeCaseTransformer 下划线命名转换器
type SnakeCaseTransformer struct{}

func (t *SnakeCaseTransformer) GetName() string { return "snakeCase" }

func (t *SnakeCaseTransformer) CanTransform(fromType, toType reflect.Type) bool {
	return fromType.Kind() == reflect.Map || fromType.Kind() == reflect.Struct
}

func (t *SnakeCaseTransformer) Transform(value any, rule TransformRule) (any, error) {
	if value == nil {
		return nil, nil
	}

	// 处理map
	if m, ok := value.(map[string]any); ok {
		result := make(map[string]any)
		for k, v := range m {
			snakeKey := t.toSnakeCase(k)
			result[snakeKey] = v
		}
		return result, nil
	}

	// 处理切片
	if slice, ok := value.([]any); ok {
		result := make([]any, len(slice))
		for i, item := range slice {
			if transformed, err := t.Transform(item, rule); err == nil {
				result[i] = transformed
			} else {
				result[i] = item
			}
		}
		return result, nil
	}

	return value, nil
}

func (t *SnakeCaseTransformer) toSnakeCase(s string) string {
	var result strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// TransformConfig 转换配置
type TransformConfig struct {
	Enabled bool                     `json:"enabled" yaml:"enabled"`
	Rules   map[string]TransformRule `json:"rules" yaml:"rules"`
}

// NewTransformConfig 创建默认转换配置
func NewTransformConfig() *TransformConfig {
	return &TransformConfig{
		Enabled: true,
		Rules:   make(map[string]TransformRule),
	}
}
