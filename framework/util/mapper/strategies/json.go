package strategies

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// JSONMapper 基于JSON序列化的高性能映射器
type JSONMapper struct {
	mu        sync.RWMutex
	typeCache map[reflect.Type]bool // 缓存类型是否支持JSON映射
	checking  map[reflect.Type]bool // 正在检查的类型，防止递归
}

// NewJSONMapper 创建JSON映射器
func NewJSONMapper() *JSONMapper {
	return &JSONMapper{
		typeCache: make(map[reflect.Type]bool),
		checking:  make(map[reflect.Type]bool),
	}
}

// Map 使用JSON映射对象
func (jm *JSONMapper) Map(src, dst any, config interface{}) error {
	if src == nil {
		return nil
	}

	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	// 去除指针
	if srcType.Kind() == reflect.Ptr {
		srcType = srcType.Elem()
	}
	if dstType.Kind() == reflect.Ptr {
		dstType = dstType.Elem()
	}

	// 检查类型兼容性
	if !jm.isJSONMappable(srcType) || !jm.isJSONMappable(dstType) {
		return fmt.Errorf("types are not JSON mappable: %s -> %s", srcType, dstType)
	}

	// 使用JSON进行快速映射
	return jm.fastJSONMap(src, dst)
}

// fastJSONMap 快速JSON映射
func (jm *JSONMapper) fastJSONMap(src, dst any) error {
	// 序列化源对象
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("JSON marshal failed: %w", err)
	}

	// 反序列化到目标对象
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return nil
}

// isJSONMappable 检查类型是否支持JSON映射
func (jm *JSONMapper) isJSONMappable(t reflect.Type) bool {
	jm.mu.RLock()
	// 检查缓存
	if result, exists := jm.typeCache[t]; exists {
		jm.mu.RUnlock()
		return result
	}
	jm.mu.RUnlock()

	jm.mu.Lock()
	defer jm.mu.Unlock()

	// 再次检查缓存（双重检查）
	if result, exists := jm.typeCache[t]; exists {
		return result
	}

	// 检查是否正在检查此类型（防止无限递归）
	if jm.checking[t] {
		return true // 假设递归类型是可映射的
	}

	// 标记正在检查此类型
	jm.checking[t] = true

	result := jm.checkJSONMappable(t)

	// 移除检查标记并缓存结果
	delete(jm.checking, t)
	jm.typeCache[t] = result
	
	return result
}

// checkJSONMappable 检查类型是否可JSON映射
func (jm *JSONMapper) checkJSONMappable(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true

	case reflect.Struct:
		// 检查结构体字段，但要避免递归调用
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				// 对于结构体字段，进行简化检查，避免深度递归
				fieldType := field.Type
				if fieldType.Kind() == reflect.Ptr {
					fieldType = fieldType.Elem()
				}
				
				// 只检查基本类型和已知的安全类型
				if !jm.isBasicJSONType(fieldType) {
					return false
				}
			}
		}
		return true

	case reflect.Slice, reflect.Array:
		elemType := t.Elem()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		return jm.isBasicJSONType(elemType)

	case reflect.Map:
		return t.Key().Kind() == reflect.String && jm.isBasicJSONType(t.Elem())

	case reflect.Ptr:
		return jm.isBasicJSONType(t.Elem())

	case reflect.Interface:
		return true // interface{} 通常可以JSON映射

	default:
		return false
	}
}

// isBasicJSONType 检查是否为基本JSON类型，避免递归调用
func (jm *JSONMapper) isBasicJSONType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Struct:
		// 对于结构体，简单返回 true，假设大多数结构体都是可序列化的
		return true
	case reflect.Slice, reflect.Array:
		// 对于切片和数组，检查元素类型
		elemType := t.Elem()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		// 基本类型的切片是安全的
		switch elemType.Kind() {
		case reflect.Bool, reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return true
		case reflect.Struct:
			return true // 假设结构体切片也是可序列化的
		}
		return false
	case reflect.Map:
		// Map 类型
		return t.Key().Kind() == reflect.String
	case reflect.Interface:
		return true
	default:
		return false
	}
}