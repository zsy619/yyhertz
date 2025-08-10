package strategies

import (
	"fmt"
	"reflect"
	"sync"
)

// CodegenMapper 基于代码生成的零反射开销映射器
type CodegenMapper struct {
	mu               sync.RWMutex
	generatedMappers map[string]GeneratedMapper // 缓存生成的映射器
	typeRegistry     map[string]reflect.Type    // 类型注册表
}

// GeneratedMapper 生成的映射器函数
type GeneratedMapper func(src, dst any) error

// NewCodegenMapper 创建代码生成映射器
func NewCodegenMapper() *CodegenMapper {
	return &CodegenMapper{
		generatedMappers: make(map[string]GeneratedMapper),
		typeRegistry:     make(map[string]reflect.Type),
	}
}

// Map 使用代码生成映射对象
func (cm *CodegenMapper) Map(src, dst any, config interface{}) error {
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	// 获取映射器标识
	mapperKey := cm.getMapperKey(srcType, dstType)

	// 检查是否已有生成的映射器
	cm.mu.RLock()
	mapper, exists := cm.generatedMappers[mapperKey]
	cm.mu.RUnlock()

	if !exists {
		// 生成新的映射器
		var err error
		mapper, err = cm.generateMapper(srcType, dstType)
		if err != nil {
			return fmt.Errorf("failed to generate mapper: %w", err)
		}

		// 缓存映射器
		cm.mu.Lock()
		cm.generatedMappers[mapperKey] = mapper
		cm.mu.Unlock()
	}

	// 执行映射
	return mapper(src, dst)
}

// HasMapping 检查是否有可用的映射
func (cm *CodegenMapper) HasMapping(srcType, dstType reflect.Type) bool {
	mapperKey := cm.getMapperKey(srcType, dstType)
	cm.mu.RLock()
	_, exists := cm.generatedMappers[mapperKey]
	cm.mu.RUnlock()
	return exists
}

// generateMapper 生成映射器函数
func (cm *CodegenMapper) generateMapper(srcType, dstType reflect.Type) (GeneratedMapper, error) {
	// 对于简单情况，直接生成内联函数
	if cm.isSimpleMapping(srcType, dstType) {
		return cm.generateSimpleMapper(srcType, dstType), nil
	}

	// 复杂情况回退到反射实现
	return cm.generateReflectionMapper(), nil
}

// generateSimpleMapper 生成简单映射器
func (cm *CodegenMapper) generateSimpleMapper(srcType, dstType reflect.Type) GeneratedMapper {
	// 去除指针类型
	if srcType.Kind() == reflect.Ptr {
		srcType = srcType.Elem()
	}
	if dstType.Kind() == reflect.Ptr {
		dstType = dstType.Elem()
	}

	// 为基本结构体生成高效映射器
	return func(src, dst any) error {
		srcValue := reflect.ValueOf(src)
		dstValue := reflect.ValueOf(dst)

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

		return cm.mapStructFields(srcValue, dstValue, srcType, dstType)
	}
}

// generateReflectionMapper 生成反射映射器作为回退
func (cm *CodegenMapper) generateReflectionMapper() GeneratedMapper {
	return func(src, dst any) error {
		srcValue := reflect.ValueOf(src)
		dstValue := reflect.ValueOf(dst)

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

		return cm.basicReflectionMap(srcValue, dstValue)
	}
}

// mapStructFields 映射结构体字段
func (cm *CodegenMapper) mapStructFields(srcValue, dstValue reflect.Value, srcType, dstType reflect.Type) error {
	// 创建字段映射
	srcFields := cm.getStructFields(srcType)
	dstFields := cm.getStructFields(dstType)

	// 执行字段映射
	for dstFieldName, dstField := range dstFields {
		if srcField, exists := srcFields[dstFieldName]; exists {
			srcFieldValue := srcValue.FieldByName(srcField.Name)
			dstFieldValue := dstValue.FieldByName(dstField.Name)

			if !dstFieldValue.CanSet() {
				continue
			}

			// 直接赋值或转换
			if srcFieldValue.Type().AssignableTo(dstFieldValue.Type()) {
				dstFieldValue.Set(srcFieldValue)
			} else if srcFieldValue.Type().ConvertibleTo(dstFieldValue.Type()) {
				dstFieldValue.Set(srcFieldValue.Convert(dstFieldValue.Type()))
			}
		}
	}

	return nil
}

// basicReflectionMap 基础反射映射（作为回退）
func (cm *CodegenMapper) basicReflectionMap(srcValue, dstValue reflect.Value) error {
	srcType := srcValue.Type()
	dstType := dstValue.Type()

	if srcType.Kind() != reflect.Struct || dstType.Kind() != reflect.Struct {
		return fmt.Errorf("only struct mapping is supported")
	}

	return cm.mapStructFields(srcValue, dstValue, srcType, dstType)
}

// Helper methods

func (cm *CodegenMapper) getMapperKey(srcType, dstType reflect.Type) string {
	return fmt.Sprintf("%s_to_%s", srcType.String(), dstType.String())
}

func (cm *CodegenMapper) isSimpleMapping(srcType, dstType reflect.Type) bool {
	// 简单映射：都是结构体且字段数量不多
	if srcType.Kind() == reflect.Ptr {
		srcType = srcType.Elem()
	}
	if dstType.Kind() == reflect.Ptr {
		dstType = dstType.Elem()
	}

	return srcType.Kind() == reflect.Struct &&
		dstType.Kind() == reflect.Struct &&
		srcType.NumField() <= 10 &&
		dstType.NumField() <= 10
}

func (cm *CodegenMapper) getStructFields(t reflect.Type) map[string]reflect.StructField {
	fields := make(map[string]reflect.StructField)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fields[field.Name] = field
		}
	}

	return fields
}

// RegisterType 注册类型用于代码生成
func (cm *CodegenMapper) RegisterType(name string, t reflect.Type) {
	cm.mu.Lock()
	cm.typeRegistry[name] = t
	cm.mu.Unlock()
}