package strategies

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// TypeConverter 类型转换器函数
type TypeConverter func(src any) (any, error)

// ReflectionMapper 基于反射的高性能映射器
type ReflectionMapper struct {
	mu           sync.RWMutex
	fieldCache   map[string]*FieldMapping
	structCache  map[reflect.Type]*StructInfo
	converterMap map[reflect.Type]TypeConverter
}

// FieldMapping 字段映射信息
type FieldMapping struct {
	SrcField    reflect.StructField
	DstField    reflect.StructField
	SrcIndex    []int
	DstIndex    []int
	NeedConvert bool
	Converter   TypeConverter
}

// StructInfo 结构体信息缓存
type StructInfo struct {
	Type         reflect.Type
	Fields       []reflect.StructField
	FieldMap     map[string]int
	TagMap       map[string]string
	HasConverter bool
}

// NewReflectionMapper 创建反射映射器
func NewReflectionMapper() *ReflectionMapper {
	return &ReflectionMapper{
		fieldCache:   make(map[string]*FieldMapping),
		structCache:  make(map[reflect.Type]*StructInfo),
		converterMap: make(map[reflect.Type]TypeConverter),
	}
}

// Map 使用反射映射对象
func (rm *ReflectionMapper) Map(src, dst any, config interface{}) error {
	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst)

	// 处理指针
	if srcValue.Kind() == reflect.Ptr {
		if srcValue.IsNil() {
			return nil
		}
		srcValue = srcValue.Elem()
	}

	if dstValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}
	dstValue = dstValue.Elem()

	// 检查是否可设置
	if !dstValue.CanSet() {
		return fmt.Errorf("destination is not settable")
	}

	return rm.mapValue(srcValue, dstValue, 0)
}

// mapValue 映射值
func (rm *ReflectionMapper) mapValue(srcValue, dstValue reflect.Value, depth int) error {
	if depth > 10 { // 默认最大深度
		return fmt.Errorf("maximum mapping depth exceeded")
	}

	srcType := srcValue.Type()
	dstType := dstValue.Type()

	// 处理不同类型
	switch {
	case srcType == dstType:
		// 相同类型直接赋值
		return rm.assignValue(srcValue, dstValue)
	case srcType.Kind() == dstType.Kind() && srcType.Kind() != reflect.Struct:
		// 相同基本类型直接赋值
		return rm.assignValue(srcValue, dstValue)
	case srcType.Kind() == reflect.Struct && dstType.Kind() == reflect.Struct:
		// 结构体到结构体映射
		return rm.mapStruct(srcValue, dstValue, depth+1)
	case srcType.Kind() == reflect.Slice && dstType.Kind() == reflect.Slice:
		// 切片映射
		return rm.mapSlice(srcValue, dstValue, depth+1)
	case srcType.Kind() == reflect.Map && dstType.Kind() == reflect.Map:
		// Map映射
		return rm.mapMap(srcValue, dstValue, depth+1)
	default:
		// 尝试类型转换
		return rm.convertType(srcValue, dstValue)
	}
}

// mapStruct 映射结构体
func (rm *ReflectionMapper) mapStruct(srcValue, dstValue reflect.Value, depth int) error {
	srcType := srcValue.Type()
	dstType := dstValue.Type()

	// 获取结构体信息
	srcInfo := rm.getStructInfo(srcType, "json")
	dstInfo := rm.getStructInfo(dstType, "json")

	// 创建字段映射
	mappings := rm.createFieldMappings(srcInfo, dstInfo)

	// 执行字段映射
	for _, mapping := range mappings {
		srcFieldValue := srcValue.FieldByIndex(mapping.SrcIndex)
		dstFieldValue := dstValue.FieldByIndex(mapping.DstIndex)

		// 执行字段映射
		if mapping.NeedConvert && mapping.Converter != nil {
			converted, err := mapping.Converter(srcFieldValue.Interface())
			if err != nil {
				return fmt.Errorf("field conversion failed for %s: %w", mapping.SrcField.Name, err)
			}
			dstFieldValue.Set(reflect.ValueOf(converted))
		} else {
			if err := rm.mapValue(srcFieldValue, dstFieldValue, depth); err != nil {
				return fmt.Errorf("field mapping failed for %s: %w", mapping.SrcField.Name, err)
			}
		}
	}

	return nil
}

// getStructInfo 获取结构体信息
func (rm *ReflectionMapper) getStructInfo(t reflect.Type, tagName string) *StructInfo {
	rm.mu.RLock()
	if info, exists := rm.structCache[t]; exists {
		rm.mu.RUnlock()
		return info
	}
	rm.mu.RUnlock()

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 双重检查
	if info, exists := rm.structCache[t]; exists {
		return info
	}

	info := &StructInfo{
		Type:     t,
		Fields:   make([]reflect.StructField, 0, t.NumField()),
		FieldMap: make(map[string]int),
		TagMap:   make(map[string]string),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		info.Fields = append(info.Fields, field)
		info.FieldMap[field.Name] = len(info.Fields) - 1

		// 处理标签
		if tag := field.Tag.Get(tagName); tag != "" {
			// 解析标签，支持 json:"name,omitempty" 格式
			tagParts := strings.Split(tag, ",")
			if tagParts[0] != "" && tagParts[0] != "-" {
				info.TagMap[tagParts[0]] = field.Name
			}
		}
	}

	rm.structCache[t] = info
	return info
}

// createFieldMappings 创建字段映射
func (rm *ReflectionMapper) createFieldMappings(srcInfo, dstInfo *StructInfo) []*FieldMapping {
	var mappings []*FieldMapping

	for dstFieldName, dstIndex := range dstInfo.FieldMap {
		dstField := dstInfo.Fields[dstIndex]

		// 查找源字段
		var srcField reflect.StructField
		var srcIndex int
		var found bool

		// 按字段名查找
		if srcIdx, exists := srcInfo.FieldMap[dstFieldName]; exists {
			srcField = srcInfo.Fields[srcIdx]
			srcIndex = srcIdx
			found = true
		}

		if found {
			mapping := &FieldMapping{
				SrcField: srcField,
				DstField: dstField,
				SrcIndex: []int{srcIndex},
				DstIndex: []int{dstIndex},
			}

			// 检查是否需要类型转换
			if srcField.Type != dstField.Type {
				mapping.NeedConvert = true
				// 这里可以添加自定义转换器逻辑
			}

			mappings = append(mappings, mapping)
		}
	}

	return mappings
}

// assignValue 赋值
func (rm *ReflectionMapper) assignValue(srcValue, dstValue reflect.Value) error {
	if !dstValue.CanSet() {
		return fmt.Errorf("destination field is not settable")
	}

	if srcValue.Type().AssignableTo(dstValue.Type()) {
		dstValue.Set(srcValue)
		return nil
	}

	if srcValue.Type().ConvertibleTo(dstValue.Type()) {
		dstValue.Set(srcValue.Convert(dstValue.Type()))
		return nil
	}

	return fmt.Errorf("cannot assign %s to %s", srcValue.Type(), dstValue.Type())
}

// convertType 类型转换
func (rm *ReflectionMapper) convertType(srcValue, dstValue reflect.Value) error {
	if srcValue.Type().ConvertibleTo(dstValue.Type()) {
		dstValue.Set(srcValue.Convert(dstValue.Type()))
		return nil
	}

	return fmt.Errorf("unsupported type conversion from %s to %s", srcValue.Type(), dstValue.Type())
}

// mapSlice 映射切片
func (rm *ReflectionMapper) mapSlice(srcValue, dstValue reflect.Value, depth int) error {
	srcLen := srcValue.Len()
	if srcLen == 0 {
		return nil
	}

	// 创建目标切片
	dstSlice := reflect.MakeSlice(dstValue.Type(), srcLen, srcLen)

	// 逐个映射元素
	for i := 0; i < srcLen; i++ {
		srcElem := srcValue.Index(i)
		dstElem := dstSlice.Index(i)

		if err := rm.mapValue(srcElem, dstElem, depth); err != nil {
			return fmt.Errorf("error mapping slice element at index %d: %w", i, err)
		}
	}

	dstValue.Set(dstSlice)
	return nil
}

// mapMap 映射Map
func (rm *ReflectionMapper) mapMap(srcValue, dstValue reflect.Value, depth int) error {
	if srcValue.Len() == 0 {
		return nil
	}

	// 创建目标Map
	dstMap := reflect.MakeMap(dstValue.Type())

	// 映射每个键值对
	for _, key := range srcValue.MapKeys() {
		srcVal := srcValue.MapIndex(key)
		dstVal := reflect.New(dstValue.Type().Elem()).Elem()

		if err := rm.mapValue(srcVal, dstVal, depth); err != nil {
			return fmt.Errorf("map value mapping failed for key %v: %w", key.Interface(), err)
		}

		dstMap.SetMapIndex(key, dstVal)
	}

	dstValue.Set(dstMap)
	return nil
}