// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// JSON 自定义JSON类型
type JSON map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("无法将 %T 转换为 JSON", value)
	}

	return json.Unmarshal(bytes, j)
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (JSON) GormDataType() string {
	return "json"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (JSON) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlite":
		return "TEXT"
	default:
		return "TEXT"
	}
}

// JSONArray 自定义JSON数组类型
type JSONArray []interface{}

// Value 实现driver.Valuer接口
func (ja JSONArray) Value() (driver.Value, error) {
	if ja == nil {
		return nil, nil
	}
	return json.Marshal(ja)
}

// Scan 实现sql.Scanner接口
func (ja *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*ja = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("无法将 %T 转换为 JSONArray", value)
	}

	return json.Unmarshal(bytes, ja)
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (JSONArray) GormDataType() string {
	return "json"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (JSONArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlite":
		return "TEXT"
	default:
		return "TEXT"
	}
}

// StringArray 字符串数组类型
type StringArray []string

// Value 实现driver.Valuer接口
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}

	switch len(sa) {
	case 0:
		return "{}", nil
	default:
		return "{" + strings.Join(sa, ",") + "}", nil
	}
}

// Scan 实现sql.Scanner接口
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("无法将 %T 转换为 StringArray", value)
	}

	// 解析PostgreSQL数组格式 {item1,item2,item3}
	str = strings.Trim(str, "{}")
	if str == "" {
		*sa = StringArray{}
		return nil
	}

	*sa = StringArray(strings.Split(str, ","))
	return nil
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (StringArray) GormDataType() string {
	return "text[]"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (StringArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "TEXT[]"
	case "mysql":
		return "JSON"
	case "sqlite":
		return "TEXT"
	default:
		return "TEXT"
	}
}

// IntArray 整数数组类型
type IntArray []int

// Value 实现driver.Valuer接口
func (ia IntArray) Value() (driver.Value, error) {
	if ia == nil {
		return nil, nil
	}

	if len(ia) == 0 {
		return "{}", nil
	}

	strArray := make([]string, len(ia))
	for i, v := range ia {
		strArray[i] = fmt.Sprintf("%d", v)
	}

	return "{" + strings.Join(strArray, ",") + "}", nil
}

// Scan 实现sql.Scanner接口
func (ia *IntArray) Scan(value interface{}) error {
	if value == nil {
		*ia = nil
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("无法将 %T 转换为 IntArray", value)
	}

	// 解析PostgreSQL数组格式 {item1,item2,item3}
	str = strings.Trim(str, "{}")
	if str == "" {
		*ia = IntArray{}
		return nil
	}

	parts := strings.Split(str, ",")
	result := make(IntArray, len(parts))

	for i, part := range parts {
		var val int
		if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &val); err != nil {
			return fmt.Errorf("解析整数失败: %s", part)
		}
		result[i] = val
	}

	*ia = result
	return nil
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (IntArray) GormDataType() string {
	return "integer[]"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (IntArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "INTEGER[]"
	case "mysql":
		return "JSON"
	case "sqlite":
		return "TEXT"
	default:
		return "TEXT"
	}
}

// BLOB 二进制大对象类型
type BLOB []byte

// Value 实现driver.Valuer接口
func (b BLOB) Value() (driver.Value, error) {
	if b == nil {
		return nil, nil
	}
	return []byte(b), nil
}

// Scan 实现sql.Scanner接口
func (b *BLOB) Scan(value interface{}) error {
	if value == nil {
		*b = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*b = BLOB(v)
	case string:
		*b = BLOB([]byte(v))
	default:
		return fmt.Errorf("无法将 %T 转换为 BLOB", value)
	}

	return nil
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (BLOB) GormDataType() string {
	return "blob"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (BLOB) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "LONGBLOB"
	case "postgres":
		return "BYTEA"
	case "sqlite":
		return "BLOB"
	default:
		return "BLOB"
	}
}

// GenericJSON 泛型JSON类型
type GenericJSON[T any] struct {
	Data T
}

// Value 实现driver.Valuer接口
func (gj GenericJSON[T]) Value() (driver.Value, error) {
	return json.Marshal(gj.Data)
}

// Scan 实现sql.Scanner接口
func (gj *GenericJSON[T]) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("无法将 %T 转换为 GenericJSON", value)
	}

	return json.Unmarshal(bytes, &gj.Data)
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (GenericJSON[T]) GormDataType() string {
	return "json"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (GenericJSON[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlite":
		return "TEXT"
	default:
		return "TEXT"
	}
}

// ============= 数据类型工具函数 =============

// NewJSON 创建JSON类型
func NewJSON(data map[string]interface{}) JSON {
	return JSON(data)
}

// NewJSONArray 创建JSON数组类型
func NewJSONArray(data []interface{}) JSONArray {
	return JSONArray(data)
}

// NewStringArray 创建字符串数组类型
func NewStringArray(data []string) StringArray {
	return StringArray(data)
}

// NewIntArray 创建整数数组类型
func NewIntArray(data []int) IntArray {
	return IntArray(data)
}

// NewBLOB 创建BLOB类型
func NewBLOB(data []byte) BLOB {
	return BLOB(data)
}

// NewGenericJSON 创建泛型JSON类型
func NewGenericJSON[T any](data T) GenericJSON[T] {
	return GenericJSON[T]{Data: data}
}

// ============= 数据类型转换函数 =============

// ToJSON 转换为JSON类型
func ToJSON(v interface{}) (JSON, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		return JSON(val), nil
	case JSON:
		return val, nil
	default:
		// 尝试通过JSON序列化/反序列化转换
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("转换为JSON失败: %w", err)
		}

		var result JSON
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return nil, fmt.Errorf("解析JSON失败: %w", err)
		}

		return result, nil
	}
}

// ToJSONArray 转换为JSON数组类型
func ToJSONArray(v interface{}) (JSONArray, error) {
	switch val := v.(type) {
	case []interface{}:
		return JSONArray(val), nil
	case JSONArray:
		return val, nil
	default:
		// 使用反射检查是否为切片
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Slice {
			return nil, fmt.Errorf("值不是切片类型")
		}

		result := make(JSONArray, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}

		return result, nil
	}
}

// ToStringArray 转换为字符串数组类型
func ToStringArray(v interface{}) (StringArray, error) {
	switch val := v.(type) {
	case []string:
		return StringArray(val), nil
	case StringArray:
		return val, nil
	default:
		// 使用反射检查是否为切片
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Slice {
			return nil, fmt.Errorf("值不是切片类型")
		}

		result := make(StringArray, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = fmt.Sprintf("%v", rv.Index(i).Interface())
		}

		return result, nil
	}
}

// ToIntArray 转换为整数数组类型
func ToIntArray(v interface{}) (IntArray, error) {
	switch val := v.(type) {
	case []int:
		return IntArray(val), nil
	case IntArray:
		return val, nil
	default:
		// 使用反射检查是否为切片
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Slice {
			return nil, fmt.Errorf("值不是切片类型")
		}

		result := make(IntArray, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			switch elem.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				result[i] = int(elem.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				result[i] = int(elem.Uint())
			case reflect.Float32, reflect.Float64:
				result[i] = int(elem.Float())
			default:
				return nil, fmt.Errorf("无法将 %v 转换为整数", elem.Interface())
			}
		}

		return result, nil
	}
}

// ToBLOB 转换为BLOB类型
func ToBLOB(v interface{}) (BLOB, error) {
	switch val := v.(type) {
	case []byte:
		return BLOB(val), nil
	case BLOB:
		return val, nil
	case string:
		return BLOB([]byte(val)), nil
	default:
		return nil, fmt.Errorf("无法将 %T 转换为 BLOB", v)
	}
}

// ============= 数据类型验证函数 =============

// IsValidJSON 验证是否为有效的JSON
func IsValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

// IsValidJSONString 验证字符串是否为有效的JSON
func IsValidJSONString(str string) bool {
	return IsValidJSON([]byte(str))
}

// ============= 数据类型注册函数 =============

// RegisterCustomDataTypes 注册自定义数据类型
func RegisterCustomDataTypes(db *gorm.DB) error {
	// 注册JSON类型
	if err := db.Callback().Create().Before("gorm:create").Register("json_create", jsonCreateCallback); err != nil {
		return fmt.Errorf("注册JSON创建回调失败: %w", err)
	}

	if err := db.Callback().Update().Before("gorm:update").Register("json_update", jsonUpdateCallback); err != nil {
		return fmt.Errorf("注册JSON更新回调失败: %w", err)
	}

	return nil
}

// jsonCreateCallback JSON创建回调
func jsonCreateCallback(db *gorm.DB) {
	// 在这里可以添加JSON字段的特殊处理逻辑
}

// jsonUpdateCallback JSON更新回调
func jsonUpdateCallback(db *gorm.DB) {
	// 在这里可以添加JSON字段的特殊处理逻辑
}
