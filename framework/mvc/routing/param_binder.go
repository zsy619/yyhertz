package routing

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ParamBinder 参数绑定器（统一从comment包提取和增强）
type ParamBinder struct {
	// 可以添加配置选项
}

// NewParamBinder 创建参数绑定器
func NewParamBinder() *ParamBinder {
	return &ParamBinder{}
}

// PrepareMethodArgs 准备方法参数（从comment包提取）
func (pb *ParamBinder) PrepareMethodArgs(methodInfo *MethodInfo, c *app.RequestContext, controllerValue reflect.Value) ([]reflect.Value, error) {
	method := controllerValue.MethodByName(methodInfo.MethodName)
	methodType := method.Type()
	
	args := make([]reflect.Value, methodType.NumIn())
	
	// 处理参数
	for i := 0; i < methodType.NumIn(); i++ {
		paramType := methodType.In(i)
		
		var paramValue reflect.Value
		var err error
		
		// 根据参数信息获取参数值
		if i < len(methodInfo.Params) {
			paramInfo := methodInfo.Params[i]
			paramValue, err = pb.GetParamValueFromInfo(paramInfo, paramType, c)
		} else {
			// 如果没有参数信息，尝试自动推断
			paramValue, err = pb.InferParamValue(paramType, c)
		}
		
		if err != nil {
			return nil, fmt.Errorf("failed to get parameter %d: %w", i, err)
		}
		
		args[i] = paramValue
	}
	
	return args, nil
}

// GetParamValueFromInfo 根据参数信息获取参数值（从comment包提取和增强）
func (pb *ParamBinder) GetParamValueFromInfo(paramInfo *ParamInfo, paramType reflect.Type, c *app.RequestContext) (reflect.Value, error) {
	switch paramInfo.Source {
	case ParamSourcePath:
		// 路径参数
		value := c.Param(paramInfo.Name)
		return pb.ConvertValue(value, paramType)
		
	case ParamSourceQuery:
		// 查询参数
		value := c.Query(paramInfo.Name)
		if value == "" && paramInfo.DefaultValue != "" {
			value = paramInfo.DefaultValue
		}
		return pb.ConvertValue(value, paramType)
		
	case ParamSourceBody:
		// 请求体参数
		return pb.ParseBodyParam(paramType, c)
		
	case ParamSourceHeader:
		// 请求头参数
		value := c.GetHeader(paramInfo.Name)
		if string(value) == "" && paramInfo.DefaultValue != "" {
			value = []byte(paramInfo.DefaultValue)
		}
		return pb.ConvertValue(string(value), paramType)
		
	case ParamSourceCookie:
		// Cookie参数
		value := string(c.Cookie(paramInfo.Name))
		if value == "" && paramInfo.DefaultValue != "" {
			value = paramInfo.DefaultValue
		}
		return pb.ConvertValue(value, paramType)
		
	case ParamSourceForm:
		// 表单参数
		value := string(c.FormValue(paramInfo.Name))
		if value == "" && paramInfo.DefaultValue != "" {
			value = paramInfo.DefaultValue
		}
		return pb.ConvertValue(value, paramType)
		
	default:
		return pb.InferParamValue(paramType, c)
	}
}

// InferParamValue 自动推断参数值（从comment包提取和增强）
func (pb *ParamBinder) InferParamValue(paramType reflect.Type, c *app.RequestContext) (reflect.Value, error) {
	switch {
	case IsContextType(paramType):
		// Context类型
		ctx := &contextenhanced.Context{RequestContext: c}
		return reflect.ValueOf(ctx), nil
		
	case IsStructType(paramType):
		// 结构体类型，解析为请求体
		return pb.ParseBodyParam(paramType, c)
		
	case IsStringType(paramType):
		// 字符串类型，返回空字符串
		return reflect.ValueOf(""), nil
		
	default:
		// 创建零值
		return reflect.Zero(paramType), nil
	}
}

// ParseBodyParam 解析请求体参数（从comment包提取和增强）
func (pb *ParamBinder) ParseBodyParam(paramType reflect.Type, c *app.RequestContext) (reflect.Value, error) {
	// 创建参数类型的实例
	if paramType.Kind() == reflect.Ptr {
		// 指针类型
		elemType := paramType.Elem()
		elemValue := reflect.New(elemType)
		
		// 绑定JSON数据
		err := c.BindAndValidate(elemValue.Interface())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to bind request body: %w", err)
		}
		
		return elemValue, nil
	} else {
		// 值类型
		paramValue := reflect.New(paramType)
		
		// 绑定JSON数据
		err := c.BindAndValidate(paramValue.Interface())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to bind request body: %w", err)
		}
		
		return paramValue.Elem(), nil
	}
}

// ConvertValue 转换值类型（增强版本）
func (pb *ParamBinder) ConvertValue(value string, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil
		
	case reflect.Int:
		if value == "" {
			return reflect.ValueOf(0), nil
		}
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to int: %w", value, err)
		}
		return reflect.ValueOf(intVal), nil
		
	case reflect.Int8:
		if value == "" {
			return reflect.ValueOf(int8(0)), nil
		}
		intVal, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to int8: %w", value, err)
		}
		return reflect.ValueOf(int8(intVal)), nil
		
	case reflect.Int16:
		if value == "" {
			return reflect.ValueOf(int16(0)), nil
		}
		intVal, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to int16: %w", value, err)
		}
		return reflect.ValueOf(int16(intVal)), nil
		
	case reflect.Int32:
		if value == "" {
			return reflect.ValueOf(int32(0)), nil
		}
		intVal, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to int32: %w", value, err)
		}
		return reflect.ValueOf(int32(intVal)), nil
		
	case reflect.Int64:
		if value == "" {
			return reflect.ValueOf(int64(0)), nil
		}
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to int64: %w", value, err)
		}
		return reflect.ValueOf(intVal), nil
		
	case reflect.Uint:
		if value == "" {
			return reflect.ValueOf(uint(0)), nil
		}
		uintVal, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to uint: %w", value, err)
		}
		return reflect.ValueOf(uint(uintVal)), nil
		
	case reflect.Uint8:
		if value == "" {
			return reflect.ValueOf(uint8(0)), nil
		}
		uintVal, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to uint8: %w", value, err)
		}
		return reflect.ValueOf(uint8(uintVal)), nil
		
	case reflect.Uint16:
		if value == "" {
			return reflect.ValueOf(uint16(0)), nil
		}
		uintVal, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to uint16: %w", value, err)
		}
		return reflect.ValueOf(uint16(uintVal)), nil
		
	case reflect.Uint32:
		if value == "" {
			return reflect.ValueOf(uint32(0)), nil
		}
		uintVal, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to uint32: %w", value, err)
		}
		return reflect.ValueOf(uint32(uintVal)), nil
		
	case reflect.Uint64:
		if value == "" {
			return reflect.ValueOf(uint64(0)), nil
		}
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to uint64: %w", value, err)
		}
		return reflect.ValueOf(uintVal), nil
		
	case reflect.Float32:
		if value == "" {
			return reflect.ValueOf(float32(0)), nil
		}
		floatVal, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to float32: %w", value, err)
		}
		return reflect.ValueOf(float32(floatVal)), nil
		
	case reflect.Float64:
		if value == "" {
			return reflect.ValueOf(float64(0)), nil
		}
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to float64: %w", value, err)
		}
		return reflect.ValueOf(floatVal), nil
		
	case reflect.Bool:
		if value == "" {
			return reflect.ValueOf(false), nil
		}
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to bool: %w", value, err)
		}
		return reflect.ValueOf(boolVal), nil
		
	case reflect.Slice:
		// 处理切片类型（例如 []string）
		if targetType.Elem().Kind() == reflect.String {
			// 字符串切片，按逗号分割
			if value == "" {
				return reflect.ValueOf([]string{}), nil
			}
			parts := strings.Split(value, ",")
			for i, part := range parts {
				parts[i] = strings.TrimSpace(part)
			}
			return reflect.ValueOf(parts), nil
		}
		return reflect.Zero(targetType), nil
		
	default:
		return reflect.Zero(targetType), nil
	}
}

// ValidateParams 验证参数（新增功能）
func (pb *ParamBinder) ValidateParams(params []*ParamInfo, c *app.RequestContext) error {
	for _, param := range params {
		if param.Required {
			var value string
			
			switch param.Source {
			case ParamSourcePath:
				value = c.Param(param.Name)
			case ParamSourceQuery:
				value = c.Query(param.Name)
			case ParamSourceHeader:
				value = string(c.GetHeader(param.Name))
			case ParamSourceCookie:
				value = string(c.Cookie(param.Name))
			case ParamSourceForm:
				value = string(c.FormValue(param.Name))
			case ParamSourceBody:
				// 请求体参数的验证需要特殊处理
				continue
			}
			
			if value == "" && param.DefaultValue == "" {
				return &RouteError{
					Type:    ErrorTypeInvalidParam,
					Message: fmt.Sprintf("required parameter '%s' is missing", param.Name),
				}
			}
		}
	}
	
	return nil
}

// GetParamInfo 从方法签名中提取参数信息（辅助函数）
func (pb *ParamBinder) GetParamInfo(methodType reflect.Type) []*ParamInfo {
	var params []*ParamInfo
	
	// 跳过receiver参数（第0个参数）
	for i := 0; i < methodType.NumIn(); i++ {
		paramType := methodType.In(i)
		
		var paramInfo *ParamInfo
		
		// 根据类型推断参数来源
		switch {
		case IsContextType(paramType):
			// Context参数，跳过
			continue
		case IsStructType(paramType):
			// 结构体类型，通常是请求体
			paramInfo = NewBodyParam(false)
		case IsStringType(paramType):
			// 字符串类型，默认为查询参数
			paramInfo = NewQueryParam(fmt.Sprintf("param%d", i), "", false)
		default:
			// 其他类型，默认为查询参数
			paramInfo = NewQueryParam(fmt.Sprintf("param%d", i), "", false)
		}
		
		if paramInfo != nil {
			params = append(params, paramInfo)
		}
	}
	
	return params
}