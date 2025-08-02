package config

import (
	"fmt"
	"reflect"
	"strings"
)

// convertInterfaceSliceToBool 将 []any 转换为 []bool
func convertInterfaceSliceToBool(ifaceSlice []any) ([]bool, error) {
	result := make([]bool, len(ifaceSlice))
	for i, val := range ifaceSlice {
		switch v := val.(type) {
		case bool:
			result[i] = v
		case string:
			b, err := parseBool(v)
			if err != nil {
				return nil, fmt.Errorf("element %d: %v", i, err)
			}
			result[i] = b
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			// 非零值视为 true
			result[i] = reflect.ValueOf(v).Int() != 0
		case float32, float64:
			result[i] = reflect.ValueOf(v).Float() != 0
		default:
			return nil, fmt.Errorf("unsupported type for element %d: %T", i, val)
		}
	}
	return result, nil
}

// parseBoolSliceFromString 从字符串解析布尔切片
func parseBoolSliceFromString(str string) ([]bool, error) {
	// 去除方括号（如果有）
	str = strings.Trim(str, "[]")
	if str == "" {
		return []bool{}, nil
	}

	// 分割字符串
	parts := strings.Split(str, ",")
	result := make([]bool, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		b, err := parseBool(part)
		if err != nil {
			return nil, fmt.Errorf("element %d: %v", i, err)
		}
		result[i] = b
	}

	return result, nil
}

// parseBool 解析各种形式的布尔值
func parseBool(str string) (bool, error) {
	switch strings.ToLower(str) {
	case "1", "t", "true", "y", "yes", "on":
		return true, nil
	case "0", "f", "false", "n", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", str)
	}
}

// convertToBoolSlice 使用反射转换任意类型到布尔切片
func convertToBoolSlice(value any) ([]bool, error) {
	rv := reflect.ValueOf(value)

	// 处理切片类型
	if rv.Kind() == reflect.Slice {
		length := rv.Len()
		result := make([]bool, length)

		for i := 0; i < length; i++ {
			elem := rv.Index(i).Interface()
			switch v := elem.(type) {
			case bool:
				result[i] = v
			case string:
				b, err := parseBool(v)
				if err != nil {
					return nil, fmt.Errorf("element %d: %v", i, err)
				}
				result[i] = b
			case int, int8, int16, int32, int64:
				result[i] = reflect.ValueOf(v).Int() != 0
			case uint, uint8, uint16, uint32, uint64:
				result[i] = reflect.ValueOf(v).Uint() != 0
			case float32, float64:
				result[i] = reflect.ValueOf(v).Float() != 0
			default:
				return nil, fmt.Errorf("unsupported slice element type: %T", elem)
			}
		}
		return result, nil
	}

	// 处理单个布尔值（转换为单元素切片）
	if b, ok := value.(bool); ok {
		return []bool{b}, nil
	}

	return nil, fmt.Errorf("unsupported type: %T", value)
}

func convertToIntSlice(value any) ([]int, error) {
	rv := reflect.ValueOf(value)

	// 处理切片类型
	if rv.Kind() == reflect.Slice {
		length := rv.Len()
		result := make([]int, length)

		for i := 0; i < length; i++ {
			elem := rv.Index(i).Interface()
			switch v := elem.(type) {
			case int:
				result[i] = v
			case int8:
				result[i] = int(v)
			case int16:
				result[i] = int(v)
			case int32:
				result[i] = int(v)
			case int64:
				result[i] = int(v)
			default:
				return nil, fmt.Errorf("unsupported slice element type: %T", elem)
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("unsupported type: %T", value)
}
