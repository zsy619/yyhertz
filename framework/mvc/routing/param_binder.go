package routing

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"

	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ConverterFunc 类型转换函数
type ConverterFunc func(value string, targetType reflect.Type) (reflect.Value, error)

// ParamExtractor 参数提取器接口
type ParamExtractor interface {
	Extract(paramInfo *ParamInfo, c *app.RequestContext) (string, error)
}

// ParamBinder 参数绑定器（统一从comment包提取和增强）
type ParamBinder struct {
	// 方法缓存，避免重复反射
	methodCache map[string]reflect.Method
	// 类型转换器映射
	converters map[reflect.Kind]ConverterFunc
	// 参数提取器映射
	extractors map[ParamSource]ParamExtractor
	// 缓存锁
	cacheMutex sync.RWMutex
	// 参数切片池
	argsPool sync.Pool
}

// NewParamBinder 创建参数绑定器
func NewParamBinder() *ParamBinder {
	pb := &ParamBinder{
		methodCache: make(map[string]reflect.Method),
		converters:  make(map[reflect.Kind]ConverterFunc),
		extractors:  make(map[ParamSource]ParamExtractor),
	}

	// 初始化参数切片池
	pb.argsPool = sync.Pool{
		New: func() interface{} {
			return make([]reflect.Value, 0, 8) // 预分配8个参数的空间
		},
	}

	// 注册默认类型转换器
	pb.registerDefaultConverters()
	// 注册默认参数提取器
	pb.registerDefaultExtractors()

	return pb
}

// registerDefaultConverters 注册默认类型转换器
func (pb *ParamBinder) registerDefaultConverters() {
	// 字符串类型
	pb.converters[reflect.String] = func(value string, _ reflect.Type) (reflect.Value, error) {
		return reflect.ValueOf(value), nil
	}

	// 布尔类型
	pb.converters[reflect.Bool] = func(value string, _ reflect.Type) (reflect.Value, error) {
		if value == "" {
			return reflect.ValueOf(false), nil
		}
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to bool: %w", value, err)
		}
		return reflect.ValueOf(boolVal), nil
	}

	// 整数类型转换器工厂
	intConverter := func(bitSize int) ConverterFunc {
		return func(value string, targetType reflect.Type) (reflect.Value, error) {
			if value == "" {
				return reflect.Zero(targetType), nil
			}
			intVal, err := strconv.ParseInt(value, 10, bitSize)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert '%s' to %s: %w", value, targetType.Kind(), err)
			}
			// 根据目标类型返回相应的值
			switch targetType.Kind() {
			case reflect.Int:
				return reflect.ValueOf(int(intVal)), nil
			case reflect.Int8:
				return reflect.ValueOf(int8(intVal)), nil
			case reflect.Int16:
				return reflect.ValueOf(int16(intVal)), nil
			case reflect.Int32:
				return reflect.ValueOf(int32(intVal)), nil
			case reflect.Int64:
				return reflect.ValueOf(intVal), nil
			default:
				return reflect.Zero(targetType), fmt.Errorf("unsupported int type: %s", targetType.Kind())
			}
		}
	}

	// 无符号整数类型转换器工厂
	uintConverter := func(bitSize int) ConverterFunc {
		return func(value string, targetType reflect.Type) (reflect.Value, error) {
			if value == "" {
				return reflect.Zero(targetType), nil
			}
			uintVal, err := strconv.ParseUint(value, 10, bitSize)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert '%s' to %s: %w", value, targetType.Kind(), err)
			}
			// 根据目标类型返回相应的值
			switch targetType.Kind() {
			case reflect.Uint:
				return reflect.ValueOf(uint(uintVal)), nil
			case reflect.Uint8:
				return reflect.ValueOf(uint8(uintVal)), nil
			case reflect.Uint16:
				return reflect.ValueOf(uint16(uintVal)), nil
			case reflect.Uint32:
				return reflect.ValueOf(uint32(uintVal)), nil
			case reflect.Uint64:
				return reflect.ValueOf(uintVal), nil
			default:
				return reflect.Zero(targetType), fmt.Errorf("unsupported uint type: %s", targetType.Kind())
			}
		}
	}

	// 浮点类型转换器工厂
	floatConverter := func(bitSize int) ConverterFunc {
		return func(value string, targetType reflect.Type) (reflect.Value, error) {
			if value == "" {
				return reflect.Zero(targetType), nil
			}
			floatVal, err := strconv.ParseFloat(value, bitSize)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert '%s' to %s: %w", value, targetType.Kind(), err)
			}
			// 根据目标类型返回相应的值
			switch targetType.Kind() {
			case reflect.Float32:
				return reflect.ValueOf(float32(floatVal)), nil
			case reflect.Float64:
				return reflect.ValueOf(floatVal), nil
			default:
				return reflect.Zero(targetType), fmt.Errorf("unsupported float type: %s", targetType.Kind())
			}
		}
	}

	// 注册所有数值类型转换器
	pb.converters[reflect.Int] = intConverter(0)
	pb.converters[reflect.Int8] = intConverter(8)
	pb.converters[reflect.Int16] = intConverter(16)
	pb.converters[reflect.Int32] = intConverter(32)
	pb.converters[reflect.Int64] = intConverter(64)

	pb.converters[reflect.Uint] = uintConverter(0)
	pb.converters[reflect.Uint8] = uintConverter(8)
	pb.converters[reflect.Uint16] = uintConverter(16)
	pb.converters[reflect.Uint32] = uintConverter(32)
	pb.converters[reflect.Uint64] = uintConverter(64)

	pb.converters[reflect.Float32] = floatConverter(32)
	pb.converters[reflect.Float64] = floatConverter(64)

	// 切片类型转换器
	pb.converters[reflect.Slice] = func(value string, targetType reflect.Type) (reflect.Value, error) {
		if targetType.Elem().Kind() == reflect.String {
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
	}
}

// registerDefaultExtractors 注册默认参数提取器
func (pb *ParamBinder) registerDefaultExtractors() {
	pb.extractors[ParamSourcePath] = &PathExtractor{}
	pb.extractors[ParamSourceQuery] = &QueryExtractor{}
	pb.extractors[ParamSourceHeader] = &HeaderExtractor{}
	pb.extractors[ParamSourceCookie] = &CookieExtractor{}
	pb.extractors[ParamSourceForm] = &FormExtractor{}
}

// PathExtractor 路径参数提取器
type PathExtractor struct{}

func (pe *PathExtractor) Extract(paramInfo *ParamInfo, c *app.RequestContext) (string, error) {
	return c.Param(paramInfo.Name), nil
}

// QueryExtractor 查询参数提取器
type QueryExtractor struct{}

func (qe *QueryExtractor) Extract(paramInfo *ParamInfo, c *app.RequestContext) (string, error) {
	value := c.Query(paramInfo.Name)
	if value == "" && paramInfo.DefaultValue != "" {
		value = paramInfo.DefaultValue
	}
	return value, nil
}

// HeaderExtractor 请求头参数提取器
type HeaderExtractor struct{}

func (he *HeaderExtractor) Extract(paramInfo *ParamInfo, c *app.RequestContext) (string, error) {
	value := string(c.GetHeader(paramInfo.Name))
	if value == "" && paramInfo.DefaultValue != "" {
		value = paramInfo.DefaultValue
	}
	return value, nil
}

// CookieExtractor Cookie参数提取器
type CookieExtractor struct{}

func (ce *CookieExtractor) Extract(paramInfo *ParamInfo, c *app.RequestContext) (string, error) {
	value := string(c.Cookie(paramInfo.Name))
	if value == "" && paramInfo.DefaultValue != "" {
		value = paramInfo.DefaultValue
	}
	return value, nil
}

// FormExtractor 表单参数提取器
type FormExtractor struct{}

func (fe *FormExtractor) Extract(paramInfo *ParamInfo, c *app.RequestContext) (string, error) {
	value := string(c.FormValue(paramInfo.Name))
	if value == "" && paramInfo.DefaultValue != "" {
		value = paramInfo.DefaultValue
	}
	return value, nil
}

// PrepareMethodArgs 准备方法参数（从comment包提取）
func (pb *ParamBinder) PrepareMethodArgs(methodInfo *MethodInfo, c *app.RequestContext, controllerValue reflect.Value) ([]reflect.Value, error) {
	// 生成缓存key
	cacheKey := fmt.Sprintf("%s.%s", controllerValue.Type().String(), methodInfo.MethodName)
	
	// 尝试从缓存中获取方法信息
	pb.cacheMutex.RLock()
	method, exists := pb.methodCache[cacheKey]
	pb.cacheMutex.RUnlock()
	
	if !exists {
		// 缓存未命中，进行反射调用
		methodValue := controllerValue.MethodByName(methodInfo.MethodName)
		if !methodValue.IsValid() {
			return nil, fmt.Errorf("method %s not found in controller", methodInfo.MethodName)
		}
		
		// 从方法值获取方法类型
		method = reflect.Method{
			Type: methodValue.Type(),
		}
		
		// 写入缓存
		pb.cacheMutex.Lock()
		pb.methodCache[cacheKey] = method
		pb.cacheMutex.Unlock()
	}
	
	methodType := method.Type
	paramCount := methodType.NumIn()
	
	// 参数长度边界检查
	if paramCount == 0 {
		return []reflect.Value{}, nil
	}
	
	// 从对象池获取参数切片
	args := pb.argsPool.Get().([]reflect.Value)
	defer func() {
		// 清空切片并归还到对象池
		args = args[:0]
		pb.argsPool.Put(args)
	}()
	
	// 确保切片容量足够
	if cap(args) < paramCount {
		args = make([]reflect.Value, 0, paramCount)
	}
	args = args[:paramCount]

	// 处理参数
	for i := 0; i < paramCount; i++ {
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
			return nil, fmt.Errorf("failed to get parameter %d (%s): %w", i, paramType.String(), err)
		}

		args[i] = paramValue
	}

	// 复制结果以避免对象池的影响
	result := make([]reflect.Value, len(args))
	copy(result, args)
	
	return result, nil
}

// GetParamValueFromInfo 根据参数信息获取参数值（从comment包提取和增强）
func (pb *ParamBinder) GetParamValueFromInfo(paramInfo *ParamInfo, paramType reflect.Type, c *app.RequestContext) (reflect.Value, error) {
	// 对于请求体参数，直接使用特殊处理
	if paramInfo.Source == ParamSourceBody {
		return pb.ParseBodyParam(paramType, c)
	}
	
	// 使用参数提取器获取参数值
	extractor, exists := pb.extractors[paramInfo.Source]
	if !exists {
		return pb.InferParamValue(paramType, c)
	}
	
	value, err := extractor.Extract(paramInfo, c)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to extract parameter '%s' from %s: %w", 
			paramInfo.Name, paramInfo.Source, err)
	}
	
	// 使用新的类型转换系统
	return pb.ConvertValue(value, paramType)
}

// InferParamValue 自动推断参数值（从comment包提取和增强）
func (pb *ParamBinder) InferParamValue(paramType reflect.Type, c *app.RequestContext) (reflect.Value, error) {
	switch {
	case IsContextType(paramType):
		// Context类型
		ctx := contextenhanced.NewContext(c)
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
	// 使用转换器映射进行快速转换
	if converter, exists := pb.converters[targetType.Kind()]; exists {
		return converter(value, targetType)
	}
	
	// 处理指针类型
	if targetType.Kind() == reflect.Ptr {
		if value == "" {
			// 空值返回nil指针
			return reflect.Zero(targetType), nil
		}
		// 递归转换元素类型
		elemValue, err := pb.ConvertValue(value, targetType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		// 创建指针
		ptrValue := reflect.New(targetType.Elem())
		ptrValue.Elem().Set(elemValue)
		return ptrValue, nil
	}
	
	// 处理接口类型
	if targetType.Kind() == reflect.Interface && targetType.NumMethod() == 0 {
		// 空接口类型，返回字符串
		return reflect.ValueOf(value), nil
	}
	
	// 尝试自定义类型转换（如果目标类型实现了特定接口）
	if customValue, err := pb.tryCustomConversion(value, targetType); err == nil {
		return customValue, nil
	}
	
	// 默认返回零值
	return reflect.Zero(targetType), nil
}

// tryCustomConversion 尝试自定义类型转换
func (pb *ParamBinder) tryCustomConversion(value string, targetType reflect.Type) (reflect.Value, error) {
	// 检查是否实现了 encoding.TextUnmarshaler 接口
	if targetType.Implements(reflect.TypeOf((*interface {
		UnmarshalText([]byte) error
	})(nil)).Elem()) {
		// 创建类型实例
		instance := reflect.New(targetType)
		// 调用 UnmarshalText 方法
		method := instance.MethodByName("UnmarshalText")
		if method.IsValid() {
			results := method.Call([]reflect.Value{reflect.ValueOf([]byte(value))})
			if len(results) > 0 && !results[0].IsNil() {
				return reflect.Value{}, results[0].Interface().(error)
			}
			return instance.Elem(), nil
		}
	}
	
	// 检查是否为时间类型等其他常见自定义类型
	switch targetType.String() {
	case "time.Time":
		// 这里可以添加时间解析逻辑
		// 为简化，暂时返回错误
		return reflect.Value{}, fmt.Errorf("time.Time conversion not implemented")
	}
	
	return reflect.Value{}, fmt.Errorf("no custom conversion available for type %s", targetType.String())
}

// ValidateParams 验证参数（新增功能）
func (pb *ParamBinder) ValidateParams(params []*ParamInfo, c *app.RequestContext) error {
	var validationErrors []string
	
	for _, param := range params {
		if param.Required {
			var value string
			var err error

			// 使用提取器获取参数值
			if param.Source == ParamSourceBody {
				// 请求体参数的验证需要特殊处理，这里简单跳过
				continue
			}
			
			if extractor, exists := pb.extractors[param.Source]; exists {
				value, err = extractor.Extract(param, c)
				if err != nil {
					validationErrors = append(validationErrors, 
						fmt.Sprintf("failed to extract required parameter '%s' from %s: %v", 
							param.Name, param.Source, err))
					continue
				}
			} else {
				// 兼容旧的验证方式
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
				default:
					validationErrors = append(validationErrors, 
						fmt.Sprintf("unsupported parameter source: %s for parameter '%s'", 
							param.Source, param.Name))
					continue
				}
			}

			// 验证必填参数
			if value == "" && param.DefaultValue == "" {
				validationErrors = append(validationErrors, 
					fmt.Sprintf("required parameter '%s' is missing from %s", 
						param.Name, param.Source))
			}
		}
	}

	// 如果有验证错误，返回聚合错误
	if len(validationErrors) > 0 {
		return &RouteError{
			Type:    ErrorTypeInvalidParam,
			Message: fmt.Sprintf("parameter validation failed: %s", strings.Join(validationErrors, "; ")),
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

// RegisterConverter 注册自定义类型转换器
func (pb *ParamBinder) RegisterConverter(kind reflect.Kind, converter ConverterFunc) {
	pb.converters[kind] = converter
}

// RegisterExtractor 注册自定义参数提取器
func (pb *ParamBinder) RegisterExtractor(source ParamSource, extractor ParamExtractor) {
	pb.extractors[source] = extractor
}

// ClearCache 清空缓存（用于内存回收或重置）
func (pb *ParamBinder) ClearCache() {
	pb.cacheMutex.Lock()
	defer pb.cacheMutex.Unlock()
	
	// 清空方法缓存
	for k := range pb.methodCache {
		delete(pb.methodCache, k)
	}
}

// GetCacheStats 获取缓存统计信息（用于监控）
func (pb *ParamBinder) GetCacheStats() map[string]int {
	pb.cacheMutex.RLock()
	defer pb.cacheMutex.RUnlock()
	
	return map[string]int{
		"method_cache_size": len(pb.methodCache),
		"converters_count":  len(pb.converters),
		"extractors_count":  len(pb.extractors),
	}
}
