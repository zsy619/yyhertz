// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// SortDirection 排序方向
type SortDirection string

const (
	// SortAsc 升序
	SortAsc SortDirection = "ASC"
	// SortDesc 降序
	SortDesc SortDirection = "DESC"
)

// SortField 排序字段
type SortField struct {
	// 字段名
	Field string `json:"field"`
	// 排序方向
	Direction SortDirection `json:"direction"`
	// 是否为表达式
	IsExpression bool `json:"is_expression"`
	// 空值处理
	NullsHandling string `json:"nulls_handling"` // NULLS FIRST, NULLS LAST
}

// NewSortField 创建排序字段
func NewSortField(field string, direction SortDirection) *SortField {
	return &SortField{
		Field:     field,
		Direction: direction,
	}
}

// NewSortFieldAsc 创建升序排序字段
func NewSortFieldAsc(field string) *SortField {
	return NewSortField(field, SortAsc)
}

// NewSortFieldDesc 创建降序排序字段
func NewSortFieldDesc(field string) *SortField {
	return NewSortField(field, SortDesc)
}

// WithNullsFirst 设置空值优先
func (sf *SortField) WithNullsFirst() *SortField {
	sf.NullsHandling = "NULLS FIRST"
	return sf
}

// WithNullsLast 设置空值最后
func (sf *SortField) WithNullsLast() *SortField {
	sf.NullsHandling = "NULLS LAST"
	return sf
}

// AsExpression 设置为表达式
func (sf *SortField) AsExpression() *SortField {
	sf.IsExpression = true
	return sf
}

// ToString 转换为字符串
func (sf *SortField) ToString() string {
	var parts []string

	if sf.IsExpression {
		parts = append(parts, fmt.Sprintf("(%s)", sf.Field))
	} else {
		parts = append(parts, sf.Field)
	}

	parts = append(parts, string(sf.Direction))

	if sf.NullsHandling != "" {
		parts = append(parts, sf.NullsHandling)
	}

	return strings.Join(parts, " ")
}

// Sorter 排序器
type Sorter struct {
	fields []*SortField
	db     *gorm.DB
}

// NewSorter 创建排序器
func NewSorter(db *gorm.DB) *Sorter {
	return &Sorter{
		fields: make([]*SortField, 0),
		db:     db,
	}
}

// OrderBy 添加排序字段
func (s *Sorter) OrderBy(field string, direction SortDirection) *Sorter {
	s.fields = append(s.fields, NewSortField(field, direction))
	return s
}

// OrderByAsc 添加升序排序字段
func (s *Sorter) OrderByAsc(field string) *Sorter {
	return s.OrderBy(field, SortAsc)
}

// OrderByDesc 添加降序排序字段
func (s *Sorter) OrderByDesc(field string) *Sorter {
	return s.OrderBy(field, SortDesc)
}

// OrderByField 添加排序字段对象
func (s *Sorter) OrderByField(field *SortField) *Sorter {
	s.fields = append(s.fields, field)
	return s
}

// OrderByFields 添加多个排序字段
func (s *Sorter) OrderByFields(fields ...*SortField) *Sorter {
	s.fields = append(s.fields, fields...)
	return s
}

// OrderByExpression 添加表达式排序
func (s *Sorter) OrderByExpression(expression string, direction SortDirection) *Sorter {
	field := NewSortField(expression, direction).AsExpression()
	s.fields = append(s.fields, field)
	return s
}

// OrderByRaw 添加原始排序语句
func (s *Sorter) OrderByRaw(raw string) *Sorter {
	field := &SortField{
		Field:        raw,
		Direction:    "", // 原始语句不需要方向
		IsExpression: true,
	}
	s.fields = append(s.fields, field)
	return s
}

// Clear 清空排序字段
func (s *Sorter) Clear() *Sorter {
	s.fields = make([]*SortField, 0)
	return s
}

// GetFields 获取排序字段
func (s *Sorter) GetFields() []*SortField {
	return s.fields
}

// IsEmpty 检查是否为空
func (s *Sorter) IsEmpty() bool {
	return len(s.fields) == 0
}

// Count 获取排序字段数量
func (s *Sorter) Count() int {
	return len(s.fields)
}

// ToStringSlice 转换为字符串切片
func (s *Sorter) ToStringSlice() []string {
	result := make([]string, len(s.fields))
	for i, field := range s.fields {
		result[i] = field.ToString()
	}
	return result
}

// ToString 转换为字符串
func (s *Sorter) ToString() string {
	if s.IsEmpty() {
		return ""
	}
	return strings.Join(s.ToStringSlice(), ", ")
}

// Apply 应用排序到GORM查询
func (s *Sorter) Apply(db *gorm.DB) *gorm.DB {
	if s.IsEmpty() {
		return db
	}

	for _, field := range s.fields {
		if field.Direction == "" {
			// 原始排序语句
			db = db.Order(field.Field)
		} else {
			db = db.Order(field.ToString())
		}
	}

	return db
}

// ApplyToQuery 应用排序到查询
func (s *Sorter) ApplyToQuery() *gorm.DB {
	return s.Apply(s.db)
}

// Clone 克隆排序器
func (s *Sorter) Clone() *Sorter {
	newSorter := &Sorter{
		fields: make([]*SortField, len(s.fields)),
		db:     s.db,
	}

	for i, field := range s.fields {
		newSorter.fields[i] = &SortField{
			Field:         field.Field,
			Direction:     field.Direction,
			IsExpression:  field.IsExpression,
			NullsHandling: field.NullsHandling,
		}
	}

	return newSorter
}

// Merge 合并其他排序器
func (s *Sorter) Merge(other *Sorter) *Sorter {
	if other != nil {
		s.fields = append(s.fields, other.fields...)
	}
	return s
}

// Prepend 在前面添加排序字段
func (s *Sorter) Prepend(field *SortField) *Sorter {
	s.fields = append([]*SortField{field}, s.fields...)
	return s
}

// PrependAsc 在前面添加升序排序字段
func (s *Sorter) PrependAsc(field string) *Sorter {
	return s.Prepend(NewSortFieldAsc(field))
}

// PrependDesc 在前面添加降序排序字段
func (s *Sorter) PrependDesc(field string) *Sorter {
	return s.Prepend(NewSortFieldDesc(field))
}

// Remove 移除指定字段的排序
func (s *Sorter) Remove(field string) *Sorter {
	newFields := make([]*SortField, 0, len(s.fields))
	for _, f := range s.fields {
		if f.Field != field {
			newFields = append(newFields, f)
		}
	}
	s.fields = newFields
	return s
}

// Replace 替换指定字段的排序
func (s *Sorter) Replace(field string, newField *SortField) *Sorter {
	for i, f := range s.fields {
		if f.Field == field {
			s.fields[i] = newField
			return s
		}
	}
	// 如果没找到，则添加
	s.fields = append(s.fields, newField)
	return s
}

// ReplaceOrAdd 替换或添加排序字段
func (s *Sorter) ReplaceOrAdd(field string, direction SortDirection) *Sorter {
	return s.Replace(field, NewSortField(field, direction))
}

// ============= 预定义排序器 =============

// CommonSorters 常用排序器
type CommonSorters struct{}

// NewCommonSorters 创建常用排序器
func NewCommonSorters() *CommonSorters {
	return &CommonSorters{}
}

// ByID 按ID排序
func (cs *CommonSorters) ByID(direction SortDirection) *SortField {
	return NewSortField("id", direction)
}

// ByIDDesc 按ID降序
func (cs *CommonSorters) ByIDDesc() *SortField {
	return cs.ByID(SortDesc)
}

// ByIDASC 按ID升序
func (cs *CommonSorters) ByIDASC() *SortField {
	return cs.ByID(SortAsc)
}

// ByCreatedAt 按创建时间排序
func (cs *CommonSorters) ByCreatedAt(direction SortDirection) *SortField {
	return NewSortField("created_at", direction)
}

// ByCreatedAtDesc 按创建时间降序
func (cs *CommonSorters) ByCreatedAtDesc() *SortField {
	return cs.ByCreatedAt(SortDesc)
}

// ByCreatedAtAsc 按创建时间升序
func (cs *CommonSorters) ByCreatedAtAsc() *SortField {
	return cs.ByCreatedAt(SortAsc)
}

// ByUpdatedAt 按更新时间排序
func (cs *CommonSorters) ByUpdatedAt(direction SortDirection) *SortField {
	return NewSortField("updated_at", direction)
}

// ByUpdatedAtDesc 按更新时间降序
func (cs *CommonSorters) ByUpdatedAtDesc() *SortField {
	return cs.ByUpdatedAt(SortDesc)
}

// ByUpdatedAtAsc 按更新时间升序
func (cs *CommonSorters) ByUpdatedAtAsc() *SortField {
	return cs.ByUpdatedAt(SortAsc)
}

// ByName 按名称排序
func (cs *CommonSorters) ByName(direction SortDirection) *SortField {
	return NewSortField("name", direction)
}

// ByNameAsc 按名称升序
func (cs *CommonSorters) ByNameAsc() *SortField {
	return cs.ByName(SortAsc)
}

// ByNameDesc 按名称降序
func (cs *CommonSorters) ByNameDesc() *SortField {
	return cs.ByName(SortDesc)
}

// ============= 排序解析器 =============

// SortParser 排序解析器
type SortParser struct {
	// 默认排序字段
	DefaultField string
	// 默认排序方向
	DefaultDirection SortDirection
	// 允许的排序字段
	AllowedFields []string
	// 字段映射
	FieldMapping map[string]string
}

// NewSortParser 创建排序解析器
func NewSortParser() *SortParser {
	return &SortParser{
		DefaultField:     "id",
		DefaultDirection: SortAsc,
		AllowedFields:    []string{},
		FieldMapping:     make(map[string]string),
	}
}

// WithDefaultField 设置默认排序字段
func (sp *SortParser) WithDefaultField(field string) *SortParser {
	sp.DefaultField = field
	return sp
}

// WithDefaultDirection 设置默认排序方向
func (sp *SortParser) WithDefaultDirection(direction SortDirection) *SortParser {
	sp.DefaultDirection = direction
	return sp
}

// WithAllowedFields 设置允许的排序字段
func (sp *SortParser) WithAllowedFields(fields ...string) *SortParser {
	sp.AllowedFields = fields
	return sp
}

// WithFieldMapping 设置字段映射
func (sp *SortParser) WithFieldMapping(mapping map[string]string) *SortParser {
	sp.FieldMapping = mapping
	return sp
}

// AddFieldMapping 添加字段映射
func (sp *SortParser) AddFieldMapping(from, to string) *SortParser {
	if sp.FieldMapping == nil {
		sp.FieldMapping = make(map[string]string)
	}
	sp.FieldMapping[from] = to
	return sp
}

// ParseSortString 解析排序字符串
func (sp *SortParser) ParseSortString(sortStr string) *Sorter {
	sorter := NewSorter(nil)

	if sortStr == "" {
		// 使用默认排序
		sorter.OrderBy(sp.DefaultField, sp.DefaultDirection)
		return sorter
	}

	// 分割多个排序字段
	parts := strings.Split(sortStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		field, direction := sp.parseSortPart(part)
		if field != "" {
			sorter.OrderBy(field, direction)
		}
	}

	// 如果没有解析到任何字段，使用默认排序
	if sorter.IsEmpty() {
		sorter.OrderBy(sp.DefaultField, sp.DefaultDirection)
	}

	return sorter
}

// parseSortPart 解析单个排序部分
func (sp *SortParser) parseSortPart(part string) (string, SortDirection) {
	var field string
	var direction SortDirection = SortAsc

	// 检查是否以-开头（降序）
	if strings.HasPrefix(part, "-") {
		direction = SortDesc
		field = strings.TrimPrefix(part, "-")
	} else if strings.HasPrefix(part, "+") {
		direction = SortAsc
		field = strings.TrimPrefix(part, "+")
	} else {
		// 检查是否包含空格分隔的方向
		spaceParts := strings.Fields(part)
		if len(spaceParts) >= 2 {
			field = spaceParts[0]
			dirStr := strings.ToUpper(spaceParts[1])
			if dirStr == "DESC" || dirStr == "DESCENDING" {
				direction = SortDesc
			}
		} else {
			field = part
		}
	}

	// 应用字段映射
	if mappedField, exists := sp.FieldMapping[field]; exists {
		field = mappedField
	}

	// 检查字段是否允许
	if len(sp.AllowedFields) > 0 {
		allowed := false
		for _, allowedField := range sp.AllowedFields {
			if field == allowedField {
				allowed = true
				break
			}
		}
		if !allowed {
			config.Warnf("排序字段 %s 不被允许", field)
			return "", SortAsc
		}
	}

	return field, direction
}

// ============= 便捷函数 =============

// NewSorterWithDB 使用指定数据库创建排序器
func NewSorterWithDB(db *gorm.DB) *Sorter {
	return NewSorter(db)
}

// NewSorterWithDefault 使用默认ORM创建排序器
func NewSorterWithDefault() *Sorter {
	return NewSorter(GetDefaultORM().DB())
}

// ParseSort 解析排序字符串
func ParseSort(sortStr string) *Sorter {
	parser := NewSortParser()
	return parser.ParseSortString(sortStr)
}

// ParseSortWithFields 解析排序字符串并限制字段
func ParseSortWithFields(sortStr string, allowedFields ...string) *Sorter {
	parser := NewSortParser().WithAllowedFields(allowedFields...)
	return parser.ParseSortString(sortStr)
}

// ApplySort 应用排序到查询
func ApplySort(db *gorm.DB, sortStr string) *gorm.DB {
	sorter := ParseSort(sortStr)
	return sorter.Apply(db)
}

// ApplySortWithFields 应用排序到查询并限制字段
func ApplySortWithFields(db *gorm.DB, sortStr string, allowedFields ...string) *gorm.DB {
	sorter := ParseSortWithFields(sortStr, allowedFields...)
	return sorter.Apply(db)
}

// ============= 全局排序器 =============

var (
	// 全局常用排序器
	GlobalCommonSorters = NewCommonSorters()
	// 全局排序解析器
	GlobalSortParser = NewSortParser()
)

// SetGlobalSortParser 设置全局排序解析器
func SetGlobalSortParser(parser *SortParser) {
	GlobalSortParser = parser
}

// GetGlobalSortParser 获取全局排序解析器
func GetGlobalSortParser() *SortParser {
	return GlobalSortParser
}
