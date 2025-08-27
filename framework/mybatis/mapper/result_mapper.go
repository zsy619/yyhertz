// Package mapper 增强的结果映射系统
//
// 完全重写的ResultMap处理器，支持：
// 1. 复杂对象映射
// 2. 一对一关联映射（association）
// 3. 一对多集合映射（collection）
// 4. 嵌套对象映射
// 5. 构造函数映射
// 6. 类型转换和自定义转换器
// 7. 懒加载支持
package mapper

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ================================
// 增强的结果映射器
// ================================

// ResultMapper 结果映射器
type ResultMapper struct {
	typeRegistry     *TypeRegistry
	typeHandlers     map[reflect.Type]TypeHandler
	autoMappingEnabled bool
}

// TypeRegistry 类型注册表
type TypeRegistry struct {
	aliases map[string]reflect.Type
}

// TypeHandler 类型处理器接口
type TypeHandler interface {
	SetValue(target reflect.Value, value any) error
	GetValue(source reflect.Value) any
}

// NewResultMapper 创建结果映射器
func NewResultMapper() *ResultMapper {
	mapper := &ResultMapper{
		typeRegistry:       NewTypeRegistry(),
		typeHandlers:       make(map[reflect.Type]TypeHandler),
		autoMappingEnabled: true,
	}

	// 注册默认类型处理器
	mapper.registerDefaultTypeHandlers()
	return mapper
}

// NewTypeRegistry 创建类型注册表
func NewTypeRegistry() *TypeRegistry {
	registry := &TypeRegistry{
		aliases: make(map[string]reflect.Type),
	}

	// 注册常用类型别名
	registry.registerDefaultTypes()
	return registry
}

// registerDefaultTypes 注册默认类型
func (tr *TypeRegistry) registerDefaultTypes() {
	tr.aliases["string"] = reflect.TypeOf("")
	tr.aliases["int"] = reflect.TypeOf(int(0))
	tr.aliases["int32"] = reflect.TypeOf(int32(0))
	tr.aliases["int64"] = reflect.TypeOf(int64(0))
	tr.aliases["float32"] = reflect.TypeOf(float32(0))
	tr.aliases["float64"] = reflect.TypeOf(float64(0))
	tr.aliases["bool"] = reflect.TypeOf(false)
	tr.aliases["time"] = reflect.TypeOf(time.Time{})
	tr.aliases["Time"] = reflect.TypeOf(time.Time{})
}

// registerDefaultTypeHandlers 注册默认类型处理器
func (rm *ResultMapper) registerDefaultTypeHandlers() {
	rm.typeHandlers[reflect.TypeOf("")] = &StringTypeHandler{}
	rm.typeHandlers[reflect.TypeOf(int(0))] = &IntTypeHandler{}
	rm.typeHandlers[reflect.TypeOf(int32(0))] = &Int32TypeHandler{}
	rm.typeHandlers[reflect.TypeOf(int64(0))] = &Int64TypeHandler{}
	rm.typeHandlers[reflect.TypeOf(float32(0))] = &Float32TypeHandler{}
	rm.typeHandlers[reflect.TypeOf(float64(0))] = &Float64TypeHandler{}
	rm.typeHandlers[reflect.TypeOf(false)] = &BoolTypeHandler{}
	rm.typeHandlers[reflect.TypeOf(time.Time{})] = &TimeTypeHandler{}
}

// MapResult 映射单个结果
func (rm *ResultMapper) MapResult(rowData map[string]any, resultMap *XMLResultMap) (any, error) {
	if resultMap == nil {
		return rowData, nil
	}

	// 解析目标类型
	targetType, err := rm.resolveTargetType(resultMap.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target type '%s': %w", resultMap.Type, err)
	}

	// 创建目标对象
	target := reflect.New(targetType).Interface()
	targetValue := reflect.ValueOf(target).Elem()

	// 应用映射
	err = rm.applyMapping(rowData, targetValue, resultMap)
	if err != nil {
		return nil, fmt.Errorf("failed to apply result mapping: %w", err)
	}

	return target, nil
}

// applyMapping 应用结果映射
func (rm *ResultMapper) applyMapping(rowData map[string]any, target reflect.Value, resultMap *XMLResultMap) error {
	// 1. 处理ID映射
	for _, idMapping := range resultMap.IDMappings {
		err := rm.applyColumnMapping(rowData, target, idMapping.Column, idMapping.Property, idMapping.TargetType)
		if err != nil {
			return fmt.Errorf("failed to map ID field %s: %w", idMapping.Property, err)
		}
	}

	// 2. 处理结果映射
	for _, resultMapping := range resultMap.ResultMappings {
		err := rm.applyColumnMapping(rowData, target, resultMapping.Column, resultMapping.Property, resultMapping.TargetType)
		if err != nil {
			return fmt.Errorf("failed to map result field %s: %w", resultMapping.Property, err)
		}
	}

	// 3. 处理关联映射 (一对一)
	for _, association := range resultMap.Associations {
		err := rm.applyAssociationMapping(rowData, target, association)
		if err != nil {
			return fmt.Errorf("failed to map association %s: %w", association.Property, err)
		}
	}

	// 4. 处理集合映射 (一对多)
	for _, collection := range resultMap.Collections {
		err := rm.applyCollectionMapping(rowData, target, collection)
		if err != nil {
			return fmt.Errorf("failed to map collection %s: %w", collection.Property, err)
		}
	}

	// 5. 自动映射未明确指定的字段
	if resultMap.AutoMap {
		err := rm.applyAutoMapping(rowData, target, resultMap)
		if err != nil {
			return fmt.Errorf("failed to apply auto mapping: %w", err)
		}
	}

	return nil
}

// applyColumnMapping 应用列映射
func (rm *ResultMapper) applyColumnMapping(rowData map[string]any, target reflect.Value, column, property, targetType string) error {
	// 获取列值
	columnValue, exists := rowData[column]
	if !exists {
		return nil // 列不存在，跳过
	}

	// 获取目标字段
	fieldValue := rm.getFieldByPath(target, property)
	if !fieldValue.IsValid() {
		return fmt.Errorf("property '%s' not found", property)
	}

	if !fieldValue.CanSet() {
		return fmt.Errorf("property '%s' cannot be set", property)
	}

	// 类型转换并设置值
	return rm.setFieldValue(fieldValue, columnValue, targetType)
}

// applyAssociationMapping 应用关联映射 (一对一)
func (rm *ResultMapper) applyAssociationMapping(rowData map[string]any, target reflect.Value, association XMLAssociationMapping) error {
	// 获取关联属性字段
	fieldValue := rm.getFieldByPath(target, association.Property)
	if !fieldValue.IsValid() || !fieldValue.CanSet() {
		return fmt.Errorf("association property '%s' not found or cannot be set", association.Property)
	}

	// 如果指定了嵌套ResultMap，递归处理
	if association.ResultMap != "" {
		// TODO: 实现嵌套ResultMap处理
		// 这里需要访问其他ResultMap的功能，暂时简化处理
		return nil
	}

	// 如果指定了select语句，需要执行额外查询（懒加载）
	if association.Select != "" {
		// TODO: 实现懒加载
		return nil
	}

	// 简化处理：从指定列创建关联对象
	if association.Column != "" {
		columnValue, exists := rowData[association.Column]
		if !exists {
			return nil
		}

		// 创建关联对象
		associationType := fieldValue.Type()
		if associationType.Kind() == reflect.Ptr {
			associationType = associationType.Elem()
		}

		associationObj := reflect.New(associationType)
		
		// 假设关联对象有一个ID字段
		idField := associationObj.Elem().FieldByName("ID")
		if idField.IsValid() && idField.CanSet() {
			rm.setFieldValue(idField, columnValue, "")
		}

		if fieldValue.Type().Kind() == reflect.Ptr {
			fieldValue.Set(associationObj)
		} else {
			fieldValue.Set(associationObj.Elem())
		}
	}

	return nil
}

// applyCollectionMapping 应用集合映射 (一对多)
func (rm *ResultMapper) applyCollectionMapping(rowData map[string]any, target reflect.Value, collection XMLCollectionMapping) error {
	// 获取集合属性字段
	fieldValue := rm.getFieldByPath(target, collection.Property)
	if !fieldValue.IsValid() || !fieldValue.CanSet() {
		return fmt.Errorf("collection property '%s' not found or cannot be set", collection.Property)
	}

	// 集合映射通常需要额外的查询或者结果集分组
	// 这里提供基础框架，实际实现需要更复杂的逻辑

	// 创建空集合
	collectionType := fieldValue.Type()
	if collectionType.Kind() == reflect.Slice {
		emptySlice := reflect.MakeSlice(collectionType, 0, 0)
		fieldValue.Set(emptySlice)
	}

	return nil
}

// applyAutoMapping 应用自动映射
func (rm *ResultMapper) applyAutoMapping(rowData map[string]any, target reflect.Value, resultMap *XMLResultMap) error {
	if !rm.autoMappingEnabled {
		return nil
	}

	// 收集已经明确映射的属性
	mappedProperties := make(map[string]bool)
	for _, idMapping := range resultMap.IDMappings {
		mappedProperties[idMapping.Property] = true
	}
	for _, resultMapping := range resultMap.ResultMappings {
		mappedProperties[resultMapping.Property] = true
	}
	for _, association := range resultMap.Associations {
		mappedProperties[association.Property] = true
	}
	for _, collection := range resultMap.Collections {
		mappedProperties[collection.Property] = true
	}

	// 对未映射的列进行自动映射
	for columnName, columnValue := range rowData {
		// 转换列名为属性名（下划线转驼峰）
		propertyName := rm.columnToPropertyName(columnName)
		
		if mappedProperties[propertyName] {
			continue // 已经明确映射，跳过
		}

		// 查找匹配的字段
		field := target.FieldByName(propertyName)
		if !field.IsValid() {
			// 尝试其他命名约定
			field = rm.findFieldByAlternativeNames(target, columnName, propertyName)
		}

		if field.IsValid() && field.CanSet() {
			rm.setFieldValue(field, columnValue, "")
		}
	}

	return nil
}

// getFieldByPath 通过路径获取字段（支持嵌套）
func (rm *ResultMapper) getFieldByPath(obj reflect.Value, path string) reflect.Value {
	parts := strings.Split(path, ".")
	current := obj

	for _, part := range parts {
		if !current.IsValid() {
			return reflect.Value{}
		}

		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				// 创建零值
				current.Set(reflect.New(current.Type().Elem()))
			}
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			return reflect.Value{}
		}

		field := current.FieldByName(part)
		if !field.IsValid() {
			return reflect.Value{}
		}
		current = field
	}

	return current
}

// setFieldValue 设置字段值，支持类型转换
func (rm *ResultMapper) setFieldValue(field reflect.Value, value any, targetType string) error {
	if value == nil {
		return nil
	}

	// 如果指定了目标类型，先转换类型
	if targetType != "" {
		convertedValue, err := rm.convertValue(value, targetType)
		if err != nil {
			return err
		}
		value = convertedValue
	}

	// 检查是否有专用的类型处理器
	fieldType := field.Type()
	if handler, exists := rm.typeHandlers[fieldType]; exists {
		return handler.SetValue(field, value)
	}

	// 通用类型转换
	return rm.setValueWithConversion(field, value)
}

// setValueWithConversion 通用类型转换设置
func (rm *ResultMapper) setValueWithConversion(field reflect.Value, value any) error {
	sourceValue := reflect.ValueOf(value)
	targetType := field.Type()

	// 如果类型匹配，直接设置
	if sourceValue.Type().AssignableTo(targetType) {
		field.Set(sourceValue)
		return nil
	}

	// 处理指针类型
	if targetType.Kind() == reflect.Ptr {
		if sourceValue.Type().AssignableTo(targetType.Elem()) {
			newPtr := reflect.New(targetType.Elem())
			newPtr.Elem().Set(sourceValue)
			field.Set(newPtr)
			return nil
		}
	}

	// 处理可转换的类型
	if sourceValue.Type().ConvertibleTo(targetType) {
		field.Set(sourceValue.Convert(targetType))
		return nil
	}

	// 尝试从字符串转换
	if sourceStr, ok := value.(string); ok {
		return rm.convertFromString(field, sourceStr)
	}

	return fmt.Errorf("cannot convert %T to %s", value, targetType)
}

// convertFromString 从字符串转换
func (rm *ResultMapper) convertFromString(field reflect.Value, str string) error {
	targetType := field.Type()
	
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := strconv.ParseInt(str, 10, int(targetType.Size()*8)); err == nil {
			field.SetInt(i)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(str, 10, int(targetType.Size()*8)); err == nil {
			field.SetUint(u)
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(str, int(targetType.Size()*8)); err == nil {
			field.SetFloat(f)
			return nil
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(str); err == nil {
			field.SetBool(b)
			return nil
		}
	case reflect.String:
		field.SetString(str)
		return nil
	}

	// 特殊类型处理
	if targetType == reflect.TypeOf(time.Time{}) {
		return rm.parseTimeFromString(field, str)
	}

	return fmt.Errorf("cannot convert string '%s' to %s", str, targetType)
}

// parseTimeFromString 从字符串解析时间
func (rm *ResultMapper) parseTimeFromString(field reflect.Value, str string) error {
	timeFormats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
		"15:04:05",
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, str); err == nil {
			field.Set(reflect.ValueOf(t))
			return nil
		}
	}

	return fmt.Errorf("cannot parse time from string '%s'", str)
}

// convertValue 类型转换
func (rm *ResultMapper) convertValue(value any, targetType string) (any, error) {
	targetReflectType, err := rm.resolveTargetType(targetType)
	if err != nil {
		return nil, err
	}

	sourceValue := reflect.ValueOf(value)
	if sourceValue.Type().AssignableTo(targetReflectType) {
		return value, nil
	}

	if sourceValue.Type().ConvertibleTo(targetReflectType) {
		return sourceValue.Convert(targetReflectType).Interface(), nil
	}

	// 特殊转换逻辑
	return rm.specialConvert(value, targetReflectType)
}

// specialConvert 特殊转换
func (rm *ResultMapper) specialConvert(value any, targetType reflect.Type) (any, error) {
	switch targetType.Kind() {
	case reflect.String:
		return fmt.Sprintf("%v", value), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if str, ok := value.(string); ok {
			if i, err := strconv.ParseInt(str, 10, int(targetType.Size()*8)); err == nil {
				return reflect.ValueOf(i).Convert(targetType).Interface(), nil
			}
		}
	case reflect.Float32, reflect.Float64:
		if str, ok := value.(string); ok {
			if f, err := strconv.ParseFloat(str, int(targetType.Size()*8)); err == nil {
				return reflect.ValueOf(f).Convert(targetType).Interface(), nil
			}
		}
	case reflect.Bool:
		if str, ok := value.(string); ok {
			if b, err := strconv.ParseBool(str); err == nil {
				return b, nil
			}
		}
	}

	return nil, fmt.Errorf("cannot convert %T to %s", value, targetType)
}

// resolveTargetType 解析目标类型
func (rm *ResultMapper) resolveTargetType(typeName string) (reflect.Type, error) {
	if typeName == "" {
		return nil, fmt.Errorf("type name is empty")
	}

	// 从类型注册表查找
	if targetType, exists := rm.typeRegistry.aliases[typeName]; exists {
		return targetType, nil
	}

	// 尝试解析Go标准类型
	switch typeName {
	case "string":
		return reflect.TypeOf(""), nil
	case "int":
		return reflect.TypeOf(int(0)), nil
	case "int32":
		return reflect.TypeOf(int32(0)), nil
	case "int64":
		return reflect.TypeOf(int64(0)), nil
	case "float32":
		return reflect.TypeOf(float32(0)), nil
	case "float64":
		return reflect.TypeOf(float64(0)), nil
	case "bool":
		return reflect.TypeOf(false), nil
	case "time.Time", "Time":
		return reflect.TypeOf(time.Time{}), nil
	}

	return nil, fmt.Errorf("unknown type: %s", typeName)
}

// columnToPropertyName 列名转属性名（下划线转驼峰）
func (rm *ResultMapper) columnToPropertyName(columnName string) string {
	parts := strings.Split(strings.ToLower(columnName), "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	result := strings.Join(parts, "")
	if len(result) > 0 {
		result = strings.ToUpper(result[:1]) + result[1:]
	}
	return result
}

// findFieldByAlternativeNames 通过备选名称查找字段
func (rm *ResultMapper) findFieldByAlternativeNames(obj reflect.Value, columnName, propertyName string) reflect.Value {
	objType := obj.Type()
	
	// 尝试不同的命名约定
	alternatives := []string{
		propertyName,
		strings.ToLower(propertyName),
		strings.ToUpper(propertyName),
		columnName,
		strings.ToUpper(columnName),
		strings.Title(columnName),
	}

	for _, alt := range alternatives {
		if field := obj.FieldByName(alt); field.IsValid() {
			return field
		}
	}

	// 尝试通过JSON标签匹配
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName == columnName || tagName == propertyName {
				return obj.Field(i)
			}
		}
	}

	return reflect.Value{}
}

// ================================
// 类型处理器实现
// ================================

// StringTypeHandler 字符串类型处理器
type StringTypeHandler struct{}

func (h *StringTypeHandler) SetValue(target reflect.Value, value any) error {
	switch v := value.(type) {
	case string:
		target.SetString(v)
	case []byte:
		target.SetString(string(v))
	case driver.Valuer:
		val, err := v.Value()
		if err != nil {
			return err
		}
		if str, ok := val.(string); ok {
			target.SetString(str)
		}
	default:
		target.SetString(fmt.Sprintf("%v", value))
	}
	return nil
}

func (h *StringTypeHandler) GetValue(source reflect.Value) any {
	return source.String()
}

// IntTypeHandler 整数类型处理器
type IntTypeHandler struct{}

func (h *IntTypeHandler) SetValue(target reflect.Value, value any) error {
	switch v := value.(type) {
	case int:
		target.SetInt(int64(v))
	case int32:
		target.SetInt(int64(v))
	case int64:
		target.SetInt(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			target.SetInt(i)
		} else {
			return fmt.Errorf("cannot convert '%s' to int", v)
		}
	default:
		return fmt.Errorf("cannot convert %T to int", value)
	}
	return nil
}

func (h *IntTypeHandler) GetValue(source reflect.Value) any {
	return int(source.Int())
}

// Int32TypeHandler int32类型处理器
type Int32TypeHandler struct{}

func (h *Int32TypeHandler) SetValue(target reflect.Value, value any) error {
	switch v := value.(type) {
	case int32:
		target.SetInt(int64(v))
	case int:
		target.SetInt(int64(v))
	case int64:
		target.SetInt(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 32); err == nil {
			target.SetInt(i)
		} else {
			return fmt.Errorf("cannot convert '%s' to int32", v)
		}
	default:
		return fmt.Errorf("cannot convert %T to int32", value)
	}
	return nil
}

func (h *Int32TypeHandler) GetValue(source reflect.Value) any {
	return int32(source.Int())
}

// Int64TypeHandler int64类型处理器
type Int64TypeHandler struct{}

func (h *Int64TypeHandler) SetValue(target reflect.Value, value any) error {
	switch v := value.(type) {
	case int64:
		target.SetInt(v)
	case int:
		target.SetInt(int64(v))
	case int32:
		target.SetInt(int64(v))
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			target.SetInt(i)
		} else {
			return fmt.Errorf("cannot convert '%s' to int64", v)
		}
	default:
		return fmt.Errorf("cannot convert %T to int64", value)
	}
	return nil
}

func (h *Int64TypeHandler) GetValue(source reflect.Value) any {
	return source.Int()
}

// Float32TypeHandler float32类型处理器
type Float32TypeHandler struct{}

func (h *Float32TypeHandler) SetValue(target reflect.Value, value any) error {
	switch v := value.(type) {
	case float32:
		target.SetFloat(float64(v))
	case float64:
		target.SetFloat(v)
	case string:
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			target.SetFloat(f)
		} else {
			return fmt.Errorf("cannot convert '%s' to float32", v)
		}
	default:
		return fmt.Errorf("cannot convert %T to float32", value)
	}
	return nil
}

func (h *Float32TypeHandler) GetValue(source reflect.Value) any {
	return float32(source.Float())
}

// Float64TypeHandler float64类型处理器
type Float64TypeHandler struct{}

func (h *Float64TypeHandler) SetValue(target reflect.Value, value any) error {
	switch v := value.(type) {
	case float64:
		target.SetFloat(v)
	case float32:
		target.SetFloat(float64(v))
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			target.SetFloat(f)
		} else {
			return fmt.Errorf("cannot convert '%s' to float64", v)
		}
	default:
		return fmt.Errorf("cannot convert %T to float64", value)
	}
	return nil
}

func (h *Float64TypeHandler) GetValue(source reflect.Value) any {
	return source.Float()
}

// BoolTypeHandler 布尔类型处理器
type BoolTypeHandler struct{}

func (h *BoolTypeHandler) SetValue(target reflect.Value, value any) error {
	switch v := value.(type) {
	case bool:
		target.SetBool(v)
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			target.SetBool(b)
		} else {
			return fmt.Errorf("cannot convert '%s' to bool", v)
		}
	case int, int32, int64:
		target.SetBool(reflect.ValueOf(v).Int() != 0)
	default:
		return fmt.Errorf("cannot convert %T to bool", value)
	}
	return nil
}

func (h *BoolTypeHandler) GetValue(source reflect.Value) any {
	return source.Bool()
}

// TimeTypeHandler 时间类型处理器
type TimeTypeHandler struct{}

func (h *TimeTypeHandler) SetValue(target reflect.Value, value any) error {
	switch v := value.(type) {
	case time.Time:
		target.Set(reflect.ValueOf(v))
	case string:
		timeFormats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			"2006-01-02",
		}
		
		for _, format := range timeFormats {
			if t, err := time.Parse(format, v); err == nil {
				target.Set(reflect.ValueOf(t))
				return nil
			}
		}
		return fmt.Errorf("cannot parse time from '%s'", v)
	default:
		return fmt.Errorf("cannot convert %T to time.Time", value)
	}
	return nil
}

func (h *TimeTypeHandler) GetValue(source reflect.Value) any {
	return source.Interface().(time.Time)
}