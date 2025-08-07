package binding

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// TypeConverter 类型转换器
type TypeConverter struct {
	converters map[reflect.Type]TypeConverterFunc // 类型转换器映射
	customConverters map[reflect.Type]TypeConverterFunc // 自定义转换器
}

// NewTypeConverter 创建类型转换器
func NewTypeConverter() *TypeConverter {
	tc := &TypeConverter{
		converters:       make(map[reflect.Type]TypeConverterFunc),
		customConverters: make(map[reflect.Type]TypeConverterFunc),
	}

	// 注册内置转换器
	tc.registerBuiltinConverters()

	return tc
}

// registerBuiltinConverters 注册内置转换器
func (tc *TypeConverter) registerBuiltinConverters() {
	// 字符串转换器
	tc.converters[reflect.TypeOf("")] = tc.toString
	
	// 整数转换器
	tc.converters[reflect.TypeOf(int(0))] = tc.toInt
	tc.converters[reflect.TypeOf(int8(0))] = tc.toInt8
	tc.converters[reflect.TypeOf(int16(0))] = tc.toInt16
	tc.converters[reflect.TypeOf(int32(0))] = tc.toInt32
	tc.converters[reflect.TypeOf(int64(0))] = tc.toInt64
	
	// 无符号整数转换器
	tc.converters[reflect.TypeOf(uint(0))] = tc.toUint
	tc.converters[reflect.TypeOf(uint8(0))] = tc.toUint8
	tc.converters[reflect.TypeOf(uint16(0))] = tc.toUint16
	tc.converters[reflect.TypeOf(uint32(0))] = tc.toUint32
	tc.converters[reflect.TypeOf(uint64(0))] = tc.toUint64
	
	// 浮点数转换器
	tc.converters[reflect.TypeOf(float32(0))] = tc.toFloat32
	tc.converters[reflect.TypeOf(float64(0))] = tc.toFloat64
	
	// 布尔值转换器
	tc.converters[reflect.TypeOf(bool(false))] = tc.toBool
	
	// 时间转换器
	tc.converters[reflect.TypeOf(time.Time{})] = tc.toTime
	
	// 字节切片转换器
	tc.converters[reflect.TypeOf([]byte{})] = tc.toBytes
}

// Convert 转换值到目标类型
func (tc *TypeConverter) Convert(value interface{}, targetType reflect.Type) (interface{}, error) {
	if value == nil {
		return reflect.Zero(targetType).Interface(), nil
	}

	// 如果类型已经匹配，直接返回
	valueType := reflect.TypeOf(value)
	if valueType == targetType {
		return value, nil
	}

	// 处理指针类型
	if targetType.Kind() == reflect.Ptr {
		elemType := targetType.Elem()
		convertedValue, err := tc.Convert(value, elemType)
		if err != nil {
			return nil, err
		}
		
		// 创建指针
		ptrValue := reflect.New(elemType)
		ptrValue.Elem().Set(reflect.ValueOf(convertedValue))
		return ptrValue.Interface(), nil
	}

	// 处理切片类型
	if targetType.Kind() == reflect.Slice {
		return tc.convertToSlice(value, targetType)
	}

	// 处理数组类型
	if targetType.Kind() == reflect.Array {
		return tc.convertToArray(value, targetType)
	}

	// 处理结构体类型
	if targetType.Kind() == reflect.Struct {
		return tc.convertToStruct(value, targetType)
	}

	// 查找自定义转换器
	if converter, exists := tc.customConverters[targetType]; exists {
		return converter(value, targetType)
	}

	// 查找内置转换器
	if converter, exists := tc.converters[targetType]; exists {
		return converter(value, targetType)
	}

	// 尝试反射转换
	return tc.reflectConvert(value, targetType)
}

// GetConverter 获取转换器函数
func (tc *TypeConverter) GetConverter(targetType reflect.Type) TypeConverterFunc {
	if converter, exists := tc.customConverters[targetType]; exists {
		return converter
	}
	if converter, exists := tc.converters[targetType]; exists {
		return converter
	}
	
	// 返回通用转换器
	return func(value interface{}, targetType reflect.Type) (interface{}, error) {
		return tc.Convert(value, targetType)
	}
}

// RegisterConverter 注册自定义转换器
func (tc *TypeConverter) RegisterConverter(targetType reflect.Type, converter TypeConverterFunc) {
	tc.customConverters[targetType] = converter
}

// 内置转换器实现

// toString 转换为字符串
func (tc *TypeConverter) toString(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// toInt 转换为int
func (tc *TypeConverter) toInt(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return 0, nil
		}
		return strconv.Atoi(v)
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// toInt8 转换为int8
func (tc *TypeConverter) toInt8(value interface{}, targetType reflect.Type) (interface{}, error) {
	intVal, err := tc.toInt(value, reflect.TypeOf(int(0)))
	if err != nil {
		return nil, err
	}
	return int8(intVal.(int)), nil
}

// toInt16 转换为int16
func (tc *TypeConverter) toInt16(value interface{}, targetType reflect.Type) (interface{}, error) {
	intVal, err := tc.toInt(value, reflect.TypeOf(int(0)))
	if err != nil {
		return nil, err
	}
	return int16(intVal.(int)), nil
}

// toInt32 转换为int32
func (tc *TypeConverter) toInt32(value interface{}, targetType reflect.Type) (interface{}, error) {
	intVal, err := tc.toInt(value, reflect.TypeOf(int(0)))
	if err != nil {
		return nil, err
	}
	return int32(intVal.(int)), nil
}

// toInt64 转换为int64
func (tc *TypeConverter) toInt64(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return int64(0), nil
		}
		return strconv.ParseInt(v, 10, 64)
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case bool:
		if v {
			return int64(1), nil
		}
		return int64(0), nil
	default:
		return int64(0), fmt.Errorf("cannot convert %T to int64", value)
	}
}

// toUint 转换为uint
func (tc *TypeConverter) toUint(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return uint(0), nil
		}
		val, err := strconv.ParseUint(v, 10, 32)
		return uint(val), err
	case int:
		if v < 0 {
			return uint(0), fmt.Errorf("cannot convert negative int to uint")
		}
		return uint(v), nil
	case uint:
		return v, nil
	case uint64:
		return uint(v), nil
	case float64:
		if v < 0 {
			return uint(0), fmt.Errorf("cannot convert negative float to uint")
		}
		return uint(v), nil
	default:
		return uint(0), fmt.Errorf("cannot convert %T to uint", value)
	}
}

// toUint8 转换为uint8
func (tc *TypeConverter) toUint8(value interface{}, targetType reflect.Type) (interface{}, error) {
	uintVal, err := tc.toUint(value, reflect.TypeOf(uint(0)))
	if err != nil {
		return nil, err
	}
	return uint8(uintVal.(uint)), nil
}

// toUint16 转换为uint16
func (tc *TypeConverter) toUint16(value interface{}, targetType reflect.Type) (interface{}, error) {
	uintVal, err := tc.toUint(value, reflect.TypeOf(uint(0)))
	if err != nil {
		return nil, err
	}
	return uint16(uintVal.(uint)), nil
}

// toUint32 转换为uint32
func (tc *TypeConverter) toUint32(value interface{}, targetType reflect.Type) (interface{}, error) {
	uintVal, err := tc.toUint(value, reflect.TypeOf(uint(0)))
	if err != nil {
		return nil, err
	}
	return uint32(uintVal.(uint)), nil
}

// toUint64 转换为uint64
func (tc *TypeConverter) toUint64(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return uint64(0), nil
		}
		return strconv.ParseUint(v, 10, 64)
	case int:
		if v < 0 {
			return uint64(0), fmt.Errorf("cannot convert negative int to uint64")
		}
		return uint64(v), nil
	case uint64:
		return v, nil
	case float64:
		if v < 0 {
			return uint64(0), fmt.Errorf("cannot convert negative float to uint64")
		}
		return uint64(v), nil
	default:
		return uint64(0), fmt.Errorf("cannot convert %T to uint64", value)
	}
}

// toFloat32 转换为float32
func (tc *TypeConverter) toFloat32(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return float32(0), nil
		}
		val, err := strconv.ParseFloat(v, 32)
		return float32(val), err
	case int:
		return float32(v), nil
	case int64:
		return float32(v), nil
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	default:
		return float32(0), fmt.Errorf("cannot convert %T to float32", value)
	}
}

// toFloat64 转换为float64
func (tc *TypeConverter) toFloat64(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return float64(0), nil
		}
		return strconv.ParseFloat(v, 64)
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return float64(0), fmt.Errorf("cannot convert %T to float64", value)
	}
}

// toBool 转换为布尔值
func (tc *TypeConverter) toBool(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		switch strings.ToLower(v) {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off", "":
			return false, nil
		default:
			return strconv.ParseBool(v)
		}
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case float64:
		return v != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// toTime 转换为时间
func (tc *TypeConverter) toTime(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return time.Time{}, nil
		}
		// 尝试多种时间格式
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
			"15:04:05",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse time: %s", v)
	case int64:
		return time.Unix(v, 0), nil
	case time.Time:
		return v, nil
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time.Time", value)
	}
}

// toBytes 转换为字节切片
func (tc *TypeConverter) toBytes(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []byte", value)
	}
}

// convertToSlice 转换为切片
func (tc *TypeConverter) convertToSlice(value interface{}, targetType reflect.Type) (interface{}, error) {
	elemType := targetType.Elem()
	
	switch v := value.(type) {
	case string:
		// 字符串分割为切片
		parts := strings.Split(v, ",")
		slice := reflect.MakeSlice(targetType, len(parts), len(parts))
		for i, part := range parts {
			convertedValue, err := tc.Convert(strings.TrimSpace(part), elemType)
			if err != nil {
				return nil, err
			}
			slice.Index(i).Set(reflect.ValueOf(convertedValue))
		}
		return slice.Interface(), nil
	case []interface{}:
		// interface{}切片转换
		slice := reflect.MakeSlice(targetType, len(v), len(v))
		for i, item := range v {
			convertedValue, err := tc.Convert(item, elemType)
			if err != nil {
				return nil, err
			}
			slice.Index(i).Set(reflect.ValueOf(convertedValue))
		}
		return slice.Interface(), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to slice", value)
	}
}

// convertToArray 转换为数组
func (tc *TypeConverter) convertToArray(value interface{}, targetType reflect.Type) (interface{}, error) {
	elemType := targetType.Elem()
	arrayLen := targetType.Len()
	
	switch v := value.(type) {
	case string:
		parts := strings.Split(v, ",")
		if len(parts) > arrayLen {
			parts = parts[:arrayLen]
		}
		
		array := reflect.New(targetType).Elem()
		for i, part := range parts {
			if i >= arrayLen {
				break
			}
			convertedValue, err := tc.Convert(strings.TrimSpace(part), elemType)
			if err != nil {
				return nil, err
			}
			array.Index(i).Set(reflect.ValueOf(convertedValue))
		}
		return array.Interface(), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to array", value)
	}
}

// convertToStruct 转换为结构体
func (tc *TypeConverter) convertToStruct(value interface{}, targetType reflect.Type) (interface{}, error) {
	// 这里可以实现更复杂的结构体转换逻辑
	// 暂时只处理简单的情况
	if reflect.TypeOf(value) == targetType {
		return value, nil
	}
	
	return nil, fmt.Errorf("cannot convert %T to struct %s", value, targetType.Name())
}

// reflectConvert 反射转换
func (tc *TypeConverter) reflectConvert(value interface{}, targetType reflect.Type) (interface{}, error) {
	valueReflect := reflect.ValueOf(value)
	
	// 检查是否可以直接转换
	if valueReflect.Type().ConvertibleTo(targetType) {
		return valueReflect.Convert(targetType).Interface(), nil
	}
	
	return nil, fmt.Errorf("cannot convert %T to %s", value, targetType.Name())
}