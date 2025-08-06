// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// JSONB PostgreSQL JSONB类型
type JSONB map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONB) Scan(value interface{}) error {
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
		return fmt.Errorf("无法将 %T 转换为 JSONB", value)
	}

	return json.Unmarshal(bytes, j)
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (JSONB) GormDataType() string {
	return "jsonb"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (JSONB) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "JSONB"
	case "mysql":
		return "JSON"
	case "sqlite":
		return "TEXT"
	default:
		return "TEXT"
	}
}

// PostgreSQLEnum PostgreSQL枚举类型
type PostgreSQLEnum string

// Value 实现driver.Valuer接口
func (e PostgreSQLEnum) Value() (driver.Value, error) {
	return string(e), nil
}

// Scan 实现sql.Scanner接口
func (e *PostgreSQLEnum) Scan(value interface{}) error {
	if value == nil {
		*e = ""
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*e = PostgreSQLEnum(v)
	case string:
		*e = PostgreSQLEnum(v)
	default:
		return fmt.Errorf("无法将 %T 转换为 PostgreSQLEnum", value)
	}

	return nil
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (PostgreSQLEnum) GormDataType() string {
	return "enum"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (PostgreSQLEnum) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		// 从字段标签中获取枚举类型名称
		if enumType, ok := field.Tag.Lookup("enum_type"); ok {
			return enumType
		}
		return "TEXT" // 默认使用TEXT类型
	case "mysql":
		// MySQL使用ENUM语法
		if enumValues, ok := field.Tag.Lookup("enum_values"); ok {
			return fmt.Sprintf("ENUM(%s)", enumValues)
		}
		return "VARCHAR(255)"
	default:
		return "VARCHAR(255)"
	}
}

// Vector PostgreSQL向量类型（用于向量搜索）
type Vector []float64

// Value 实现driver.Valuer接口
func (v Vector) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	
	// 转换为PostgreSQL向量格式: [1.0,2.0,3.0]
	strValues := make([]string, len(v))
	for i, val := range v {
		strValues[i] = strconv.FormatFloat(val, 'f', -1, 64)
	}
	
	return "[" + strings.Join(strValues, ",") + "]", nil
}

// Scan 实现sql.Scanner接口
func (v *Vector) Scan(value interface{}) error {
	if value == nil {
		*v = nil
		return nil
	}

	var str string
	switch val := value.(type) {
	case []byte:
		str = string(val)
	case string:
		str = val
	default:
		return fmt.Errorf("无法将 %T 转换为 Vector", value)
	}

	// 解析PostgreSQL向量格式: [1.0,2.0,3.0]
	str = strings.Trim(str, "[]")
	if str == "" {
		*v = Vector{}
		return nil
	}

	parts := strings.Split(str, ",")
	result := make(Vector, len(parts))
	
	for i, part := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return fmt.Errorf("解析向量值失败: %s", part)
		}
		result[i] = val
	}
	
	*v = result
	return nil
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (Vector) GormDataType() string {
	return "vector"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (Vector) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		// 从字段标签中获取向量维度
		if dimension, ok := field.Tag.Lookup("vector_dimension"); ok {
			return fmt.Sprintf("VECTOR(%s)", dimension)
		}
		return "VECTOR(1536)" // 默认1536维（OpenAI embedding维度）
	default:
		return "TEXT" // 其他数据库使用TEXT存储
	}
}

// UUID PostgreSQL UUID类型
type UUID string

// Value 实现driver.Valuer接口
func (u UUID) Value() (driver.Value, error) {
	if u == "" {
		return nil, nil
	}
	return string(u), nil
}

// Scan 实现sql.Scanner接口
func (u *UUID) Scan(value interface{}) error {
	if value == nil {
		*u = ""
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*u = UUID(v)
	case string:
		*u = UUID(v)
	default:
		return fmt.Errorf("无法将 %T 转换为 UUID", value)
	}

	return nil
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (UUID) GormDataType() string {
	return "uuid"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (UUID) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "UUID"
	case "mysql":
		return "CHAR(36)"
	case "sqlite":
		return "TEXT"
	default:
		return "VARCHAR(36)"
	}
}

// INET PostgreSQL网络地址类型
type INET string

// Value 实现driver.Valuer接口
func (i INET) Value() (driver.Value, error) {
	if i == "" {
		return nil, nil
	}
	return string(i), nil
}

// Scan 实现sql.Scanner接口
func (i *INET) Scan(value interface{}) error {
	if value == nil {
		*i = ""
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*i = INET(v)
	case string:
		*i = INET(v)
	default:
		return fmt.Errorf("无法将 %T 转换为 INET", value)
	}

	return nil
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (INET) GormDataType() string {
	return "inet"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (INET) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "INET"
	default:
		return "VARCHAR(45)" // 支持IPv6的最大长度
	}
}

// CIDR PostgreSQL网络地址类型（带子网掩码）
type CIDR string

// Value 实现driver.Valuer接口
func (c CIDR) Value() (driver.Value, error) {
	if c == "" {
		return nil, nil
	}
	return string(c), nil
}

// Scan 实现sql.Scanner接口
func (c *CIDR) Scan(value interface{}) error {
	if value == nil {
		*c = ""
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*c = CIDR(v)
	case string:
		*c = CIDR(v)
	default:
		return fmt.Errorf("无法将 %T 转换为 CIDR", value)
	}

	return nil
}

// GormDataType 实现schema.GormDataTypeInterface接口
func (CIDR) GormDataType() string {
	return "cidr"
}

// GormDBDataType 实现schema.GormDBDataTypeInterface接口
func (CIDR) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "CIDR"
	default:
		return "VARCHAR(45)" // 支持IPv6的最大长度
	}
}

// ============= PostgreSQL特定工具函数 =============

// NewJSONB 创建JSONB类型
func NewJSONB(data map[string]interface{}) JSONB {
	return JSONB(data)
}

// NewPostgreSQLEnum 创建PostgreSQL枚举类型
func NewPostgreSQLEnum(value string) PostgreSQLEnum {
	return PostgreSQLEnum(value)
}

// NewVector 创建向量类型
func NewVector(values []float64) Vector {
	return Vector(values)
}

// NewUUID 创建UUID类型
func NewUUID(value string) UUID {
	return UUID(value)
}

// NewINET 创建INET类型
func NewINET(value string) INET {
	return INET(value)
}

// NewCIDR 创建CIDR类型
func NewCIDR(value string) CIDR {
	return CIDR(value)
}

// ============= 向量操作函数 =============

// Distance 计算两个向量之间的欧几里得距离
func (v Vector) Distance(other Vector) (float64, error) {
	if len(v) != len(other) {
		return 0, fmt.Errorf("向量维度不匹配: %d vs %d", len(v), len(other))
	}
	
	var sum float64
	for i := range v {
		diff := v[i] - other[i]
		sum += diff * diff
	}
	
	return sum, nil // 返回平方距离，避免开方运算
}

// DotProduct 计算两个向量的点积
func (v Vector) DotProduct(other Vector) (float64, error) {
	if len(v) != len(other) {
		return 0, fmt.Errorf("向量维度不匹配: %d vs %d", len(v), len(other))
	}
	
	var sum float64
	for i := range v {
		sum += v[i] * other[i]
	}
	
	return sum, nil
}

// CosineSimilarity 计算两个向量的余弦相似度
func (v Vector) CosineSimilarity(other Vector) (float64, error) {
	if len(v) != len(other) {
		return 0, fmt.Errorf("向量维度不匹配: %d vs %d", len(v), len(other))
	}
	
	var dotProduct, normA, normB float64
	
	for i := range v {
		dotProduct += v[i] * other[i]
		normA += v[i] * v[i]
		normB += other[i] * other[i]
	}
	
	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("零向量无法计算余弦相似度")
	}
	
	return dotProduct / (normA * normB), nil
}

// Normalize 向量归一化
func (v Vector) Normalize() Vector {
	var norm float64
	for _, val := range v {
		norm += val * val
	}
	
	if norm == 0 {
		return v
	}
	
	result := make(Vector, len(v))
	for i, val := range v {
		result[i] = val / norm
	}
	
	return result
}

// ============= 枚举类型辅助函数 =============

// EnumDefinition 枚举定义
type EnumDefinition struct {
	Name   string
	Values []string
}

// CreateEnumType 创建PostgreSQL枚举类型
func CreateEnumType(db *gorm.DB, enumDef EnumDefinition) error {
	if db.Dialector.Name() != "postgres" {
		return fmt.Errorf("枚举类型仅支持PostgreSQL")
	}
	
	// 检查枚举类型是否已存在
	var exists bool
	err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_type WHERE typname = ?)", enumDef.Name).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("检查枚举类型失败: %w", err)
	}
	
	if exists {
		return nil // 枚举类型已存在
	}
	
	// 创建枚举类型
	values := make([]string, len(enumDef.Values))
	for i, v := range enumDef.Values {
		values[i] = fmt.Sprintf("'%s'", v)
	}
	
	sql := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", enumDef.Name, strings.Join(values, ", "))
	return db.Exec(sql).Error
}

// DropEnumType 删除PostgreSQL枚举类型
func DropEnumType(db *gorm.DB, enumName string) error {
	if db.Dialector.Name() != "postgres" {
		return fmt.Errorf("枚举类型仅支持PostgreSQL")
	}
	
	sql := fmt.Sprintf("DROP TYPE IF EXISTS %s", enumName)
	return db.Exec(sql).Error
}

// AddEnumValue 添加枚举值
func AddEnumValue(db *gorm.DB, enumName, value string) error {
	if db.Dialector.Name() != "postgres" {
		return fmt.Errorf("枚举类型仅支持PostgreSQL")
	}
	
	sql := fmt.Sprintf("ALTER TYPE %s ADD VALUE '%s'", enumName, value)
	return db.Exec(sql).Error
}

// GetEnumValues 获取枚举值列表
func GetEnumValues(db *gorm.DB, enumName string) ([]string, error) {
	if db.Dialector.Name() != "postgres" {
		return nil, fmt.Errorf("枚举类型仅支持PostgreSQL")
	}
	
	var values []string
	err := db.Raw(`
		SELECT enumlabel 
		FROM pg_enum 
		WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = ?) 
		ORDER BY enumsortorder
	`, enumName).Scan(&values).Error
	
	return values, err
}

// ============= 向量搜索辅助函数 =============

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	ID       interface{} `json:"id"`
	Distance float64     `json:"distance"`
	Data     interface{} `json:"data"`
}

// VectorSearch 向量相似度搜索
func VectorSearch(db *gorm.DB, tableName, vectorColumn string, queryVector Vector, limit int) ([]VectorSearchResult, error) {
	if db.Dialector.Name() != "postgres" {
		return nil, fmt.Errorf("向量搜索仅支持PostgreSQL")
	}
	
	// 构建向量搜索SQL
	vectorStr, _ := queryVector.Value()
	sql := fmt.Sprintf(`
		SELECT id, %s <-> '%s' AS distance, *
		FROM %s
		ORDER BY %s <-> '%s'
		LIMIT %d
	`, vectorColumn, vectorStr, tableName, vectorColumn, vectorStr, limit)
	
	var results []VectorSearchResult
	err := db.Raw(sql).Scan(&results).Error
	return results, err
}

// CreateVectorIndex 创建向量索引
func CreateVectorIndex(db *gorm.DB, tableName, vectorColumn string, indexType string) error {
	if db.Dialector.Name() != "postgres" {
		return fmt.Errorf("向量索引仅支持PostgreSQL")
	}
	
	if indexType == "" {
		indexType = "ivfflat" // 默认使用ivfflat索引
	}
	
	indexName := fmt.Sprintf("idx_%s_%s_%s", tableName, vectorColumn, indexType)
	sql := fmt.Sprintf("CREATE INDEX %s ON %s USING %s (%s)", indexName, tableName, indexType, vectorColumn)
	
	return db.Exec(sql).Error
}

// ============= UUID辅助函数 =============

// GenerateUUID 生成UUID
func GenerateUUID(db *gorm.DB) (UUID, error) {
	if db.Dialector.Name() != "postgres" {
		return "", fmt.Errorf("UUID生成仅支持PostgreSQL")
	}
	
	var uuid string
	err := db.Raw("SELECT gen_random_uuid()").Scan(&uuid).Error
	return UUID(uuid), err
}

// IsValidUUID 验证UUID格式
func IsValidUUID(uuid string) bool {
	// 简单的UUID格式验证
	if len(uuid) != 36 {
		return false
	}
	
	// 检查连字符位置
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		return false
	}
	
	return true
}