package binding

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// ParameterBinder 参数绑定器
type ParameterBinder struct {
	methodType    reflect.Type              // 方法类型
	paramBinders  []ParamBinder            // 参数绑定器列表
	typeConverter *TypeConverter           // 类型转换器
	validator     *ParameterValidator      // 参数验证器
}

// ParamBinder 单个参数绑定器
type ParamBinder struct {
	Name        string                    // 参数名
	Type        reflect.Type              // 参数类型
	Index       int                       // 参数索引
	Source      ParameterSource           // 参数来源
	Required    bool                      // 是否必需
	DefaultValue interface{}              // 默认值
	Converter   TypeConverterFunc         // 类型转换函数
	Validator   ParameterValidatorFunc    // 参数验证函数
	Tags        map[string]string         // 标签信息
}

// ParameterSource 参数来源枚举
type ParameterSource int

const (
	SourceQuery  ParameterSource = iota // 查询参数
	SourcePath                         // 路径参数
	SourceForm                         // 表单参数
	SourceJSON                         // JSON体参数
	SourceHeader                       // 请求头参数
	SourceCookie                       // Cookie参数
	SourceContext                      // 上下文参数
	SourceFile                         // 文件参数
)

// TypeConverterFunc 类型转换函数
type TypeConverterFunc func(value interface{}, targetType reflect.Type) (interface{}, error)

// ParameterValidatorFunc 参数验证函数
type ParameterValidatorFunc func(value interface{}, param *ParamBinder) error

// BindingResult 绑定结果
type BindingResult struct {
	Values []interface{}         // 绑定的值
	Errors []ParameterError     // 绑定错误
}

// ParameterError 参数错误
type ParameterError struct {
	Parameter string // 参数名
	Message   string // 错误消息
	Code      string // 错误码
	Value     interface{} // 原始值
}

// NewParameterBinder 创建参数绑定器
func NewParameterBinder(methodType reflect.Type) (*ParameterBinder, error) {
	if methodType.NumIn() < 1 {
		return nil, fmt.Errorf("method must have at least one parameter (receiver)")
	}

	binder := &ParameterBinder{
		methodType:    methodType,
		paramBinders:  make([]ParamBinder, 0),
		typeConverter: NewTypeConverter(),
		validator:     NewParameterValidator(),
	}

	// 分析方法参数
	if err := binder.analyzeParameters(); err != nil {
		return nil, fmt.Errorf("failed to analyze parameters: %w", err)
	}

	return binder, nil
}

// analyzeParameters 分析方法参数
func (pb *ParameterBinder) analyzeParameters() error {
	// 跳过第一个参数（接收者）
	for i := 1; i < pb.methodType.NumIn(); i++ {
		paramType := pb.methodType.In(i)
		
		// 创建参数绑定器
		paramBinder := ParamBinder{
			Name:      fmt.Sprintf("param%d", i),
			Type:      paramType,
			Index:     i,
			Source:    pb.inferParameterSource(paramType),
			Required:  true,
			Tags:      make(map[string]string),
		}

		// 设置类型转换器
		paramBinder.Converter = pb.typeConverter.GetConverter(paramType)
		
		// 设置参数验证器
		paramBinder.Validator = pb.validator.GetValidator(paramType)

		pb.paramBinders = append(pb.paramBinders, paramBinder)
	}

	return nil
}

// inferParameterSource 推断参数来源
func (pb *ParameterBinder) inferParameterSource(paramType reflect.Type) ParameterSource {
	// 根据类型推断参数来源
	switch paramType.Kind() {
	case reflect.String:
		return SourceQuery
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return SourceQuery
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return SourceQuery
	case reflect.Float32, reflect.Float64:
		return SourceQuery
	case reflect.Bool:
		return SourceQuery
	case reflect.Struct:
		if paramType == reflect.TypeOf(time.Time{}) {
			return SourceQuery
		}
		return SourceJSON // 结构体默认从JSON绑定
	case reflect.Ptr:
		return pb.inferParameterSource(paramType.Elem())
	case reflect.Slice, reflect.Array:
		elemType := paramType.Elem()
		if elemType.Kind() == reflect.Uint8 { // []byte
			return SourceJSON
		}
		return SourceQuery
	case reflect.Map:
		return SourceJSON
	default:
		return SourceQuery
	}
}

// BindParameters 绑定参数
func (pb *ParameterBinder) BindParameters(ctx *context.Context) ([]interface{}, error) {
	// 使用适配器
	adapter := NewContextAdapter(ctx)
	
	result := &BindingResult{
		Values: make([]interface{}, len(pb.paramBinders)),
		Errors: make([]ParameterError, 0),
	}

	// 绑定每个参数
	for i, paramBinder := range pb.paramBinders {
		value, err := pb.bindParameter(adapter, &paramBinder)
		if err != nil {
			result.Errors = append(result.Errors, ParameterError{
				Parameter: paramBinder.Name,
				Message:   err.Error(),
				Code:      "BINDING_ERROR",
			})
			continue
		}

		result.Values[i] = value
	}

	// 检查是否有错误
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("parameter binding failed: %v", result.Errors)
	}

	return result.Values, nil
}

// bindParameter 绑定单个参数
func (pb *ParameterBinder) bindParameter(adapter *ContextAdapter, param *ParamBinder) (interface{}, error) {
	// 获取原始值
	rawValue, err := pb.extractRawValue(adapter, param)
	if err != nil {
		return nil, err
	}

	// 处理空值和默认值
	if rawValue == nil || rawValue == "" {
		if param.Required {
			return nil, fmt.Errorf("parameter %s is required", param.Name)
		}
		if param.DefaultValue != nil {
			return param.DefaultValue, nil
		}
		return pb.getZeroValue(param.Type), nil
	}

	// 类型转换
	convertedValue, err := pb.convertValue(rawValue, param)
	if err != nil {
		return nil, fmt.Errorf("failed to convert parameter %s: %w", param.Name, err)
	}

	// 参数验证
	if param.Validator != nil {
		if err := param.Validator(convertedValue, param); err != nil {
			return nil, fmt.Errorf("parameter %s validation failed: %w", param.Name, err)
		}
	}

	return convertedValue, nil
}

// extractRawValue 提取原始值
func (pb *ParameterBinder) extractRawValue(adapter *ContextAdapter, param *ParamBinder) (interface{}, error) {
	switch param.Source {
	case SourceQuery:
		return adapter.Query(param.Name), nil
	case SourcePath:
		return adapter.Param(param.Name), nil
	case SourceForm:
		return adapter.FormValue(param.Name), nil
	case SourceJSON:
		return pb.extractJSONValue(adapter, param)
	case SourceHeader:
		return adapter.ContextHelpers.GetHeader(adapter.ctx, param.Name), nil
	case SourceCookie:
		cookie := adapter.Cookie(param.Name)
		if cookie == "" {
			return nil, nil
		}
		return cookie, nil
	case SourceContext:
		if adapter.ctx.Keys != nil {
			return adapter.ctx.Keys[param.Name], nil
		}
		return nil, nil
	case SourceFile:
		return adapter.FormFile(param.Name)
	default:
		return nil, fmt.Errorf("unsupported parameter source: %d", param.Source)
	}
}

// extractJSONValue 提取JSON值
func (pb *ParameterBinder) extractJSONValue(adapter *ContextAdapter, param *ParamBinder) (interface{}, error) {
	// 获取请求体
	body, err := adapter.GetRawData()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	if len(body) == 0 {
		return nil, nil
	}

	// 根据参数类型解析JSON
	if param.Type.Kind() == reflect.Struct || 
	   (param.Type.Kind() == reflect.Ptr && param.Type.Elem().Kind() == reflect.Struct) {
		// 解析整个结构体
		valuePtr := reflect.New(param.Type)
		if err := json.Unmarshal(body, valuePtr.Interface()); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return valuePtr.Elem().Interface(), nil
	} else {
		// 解析特定字段
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return data[param.Name], nil
	}
}

// convertValue 转换值
func (pb *ParameterBinder) convertValue(rawValue interface{}, param *ParamBinder) (interface{}, error) {
	if param.Converter != nil {
		return param.Converter(rawValue, param.Type)
	}
	return pb.typeConverter.Convert(rawValue, param.Type)
}

// getZeroValue 获取零值
func (pb *ParameterBinder) getZeroValue(t reflect.Type) interface{} {
	return reflect.Zero(t).Interface()
}

// BindToStruct 绑定到结构体
func (pb *ParameterBinder) BindToStruct(ctx *context.Context, target interface{}) error {
	adapter := NewContextAdapter(ctx)
	
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetType := targetValue.Type().Elem()
	structValue := targetValue.Elem()

	// 遍历结构体字段
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := structValue.Field(i)

		// 跳过非导出字段
		if !fieldValue.CanSet() {
			continue
		}

		// 解析字段标签
		paramName := pb.parseFieldTag(field)
		if paramName == "-" {
			continue // 跳过忽略的字段
		}
		if paramName == "" {
			paramName = strings.ToLower(field.Name)
		}

		// 创建临时参数绑定器
		paramBinder := ParamBinder{
			Name:   paramName,
			Type:   field.Type,
			Source: pb.inferParameterSource(field.Type),
		}

		// 绑定字段值
		value, err := pb.bindParameter(adapter, &paramBinder)
		if err != nil {
			// 可选字段的错误可以忽略
			if !pb.isRequiredField(field) {
				continue
			}
			return fmt.Errorf("failed to bind field %s: %w", field.Name, err)
		}

		// 设置字段值
		if value != nil {
			fieldValue.Set(reflect.ValueOf(value))
		}
	}

	return nil
}

// parseFieldTag 解析字段标签
func (pb *ParameterBinder) parseFieldTag(field reflect.StructField) string {
	// 优先使用 json 标签
	if tag := field.Tag.Get("json"); tag != "" {
		if tag == "-" {
			return "-"
		}
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	}

	// 使用 form 标签
	if tag := field.Tag.Get("form"); tag != "" {
		if tag == "-" {
			return "-"
		}
		return tag
	}

	// 使用 query 标签
	if tag := field.Tag.Get("query"); tag != "" {
		if tag == "-" {
			return "-"
		}
		return tag
	}

	return ""
}

// isRequiredField 判断字段是否必需
func (pb *ParameterBinder) isRequiredField(field reflect.StructField) bool {
	// 检查 validate 标签
	if tag := field.Tag.Get("validate"); tag != "" {
		return strings.Contains(tag, "required")
	}
	
	// 检查 binding 标签
	if tag := field.Tag.Get("binding"); tag != "" {
		return strings.Contains(tag, "required")
	}

	return false
}

// ShouldBindQuery 从查询参数绑定
func (pb *ParameterBinder) ShouldBindQuery(ctx *context.Context, target interface{}) error {
	adapter := NewContextAdapter(ctx)
	return pb.bindFromSource(adapter, target, SourceQuery)
}

// ShouldBindJSON 从JSON体绑定
func (pb *ParameterBinder) ShouldBindJSON(ctx *context.Context, target interface{}) error {
	adapter := NewContextAdapter(ctx)
	body, err := adapter.GetRawData()
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	if len(body) == 0 {
		return fmt.Errorf("empty request body")
	}

	return json.Unmarshal(body, target)
}

// ShouldBindForm 从表单绑定
func (pb *ParameterBinder) ShouldBindForm(ctx *context.Context, target interface{}) error {
	adapter := NewContextAdapter(ctx)
	return pb.bindFromSource(adapter, target, SourceForm)
}

// bindFromSource 从指定来源绑定
func (pb *ParameterBinder) bindFromSource(adapter *ContextAdapter, target interface{}, source ParameterSource) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetType := targetValue.Type().Elem()
	structValue := targetValue.Elem()

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := structValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		paramName := pb.parseFieldTag(field)
		if paramName == "-" {
			continue
		}
		if paramName == "" {
			paramName = strings.ToLower(field.Name)
		}

		var rawValue interface{}
		var err error

		// 根据来源获取值
		switch source {
		case SourceQuery:
			rawValue = adapter.Query(paramName)
		case SourceForm:
			rawValue = adapter.FormValue(paramName)
		default:
			continue
		}

		if rawValue == nil || rawValue == "" {
			continue
		}

		// 类型转换
		convertedValue, err := pb.typeConverter.Convert(rawValue, field.Type)
		if err != nil {
			return fmt.Errorf("failed to convert field %s: %w", field.Name, err)
		}

		if convertedValue != nil {
			fieldValue.Set(reflect.ValueOf(convertedValue))
		}
	}

	return nil
}

// ValidateParameters 验证参数
func (pb *ParameterBinder) ValidateParameters(values []interface{}) error {
	for i, value := range values {
		if i < len(pb.paramBinders) {
			param := &pb.paramBinders[i]
			if param.Validator != nil {
				if err := param.Validator(value, param); err != nil {
					return fmt.Errorf("parameter %s validation failed: %w", param.Name, err)
				}
			}
		}
	}
	return nil
}

// MustBind 必须绑定（如果失败会panic）
func (pb *ParameterBinder) MustBind(ctx *context.Context, target interface{}) {
	if err := pb.BindToStruct(ctx, target); err != nil {
		panic(fmt.Sprintf("binding failed: %v", err))
	}
}

// Error 实现error接口
func (pe ParameterError) Error() string {
	return fmt.Sprintf("parameter %s: %s (code: %s)", pe.Parameter, pe.Message, pe.Code)
}