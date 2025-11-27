// Package mybatis Go风格的结构体标签和注解驱动系统
//
// 提供基于结构体标签的映射配置和自动SQL生成
package mybatis

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// TagNames 标签名称常量
const (
	TagColumn      = "column"      // 字段映射标签
	TagTable       = "table"       // 表名标签
	TagPrimaryKey  = "pk"          // 主键标签
	TagAutoIncr    = "auto_incr"   // 自增标签
	TagAssociation = "association" // 关联标签
	TagCollection  = "collection"  // 集合标签
	TagIgnore      = "ignore"      // 忽略标签
	TagValidation  = "validate"    // 验证标签
	TagCache       = "cache"       // 缓存标签
	TagLazy        = "lazy"        // 懒加载标签
)

// ColumnInfo 字段信息
type ColumnInfo struct {
	FieldName    string `json:"fieldName"`    // Go字段名
	ColumnName   string `json:"columnName"`   // 数据库列名
	IsPrimaryKey bool   `json:"isPrimaryKey"` // 是否主键
	IsAutoIncr   bool   `json:"isAutoIncr"`   // 是否自增
	DataType     string `json:"dataType"`     // 数据类型
	IsIgnored    bool   `json:"isIgnored"`    // 是否忽略
	DefaultValue any    `json:"defaultValue"` // 默认值
	Validation   string `json:"validation"`   // 验证规则
	CacheConfig  string `json:"cacheConfig"`  // 缓存配置
}

// TableInfo 表信息
type TableInfo struct {
	Name         string                 `json:"name"`         // 表名
	Columns      map[string]*ColumnInfo `json:"columns"`      // 字段映射
	PrimaryKeys  []string               `json:"primaryKeys"`  // 主键列表
	Associations []*AssociationInfo     `json:"associations"` // 关联信息
	Collections  []*CollectionInfo      `json:"collections"`  // 集合信息
	CacheEnabled bool                   `json:"cacheEnabled"` // 是否启用缓存
}

// AssociationInfo 关联信息
type AssociationInfo struct {
	PropertyName string `json:"propertyName"` // 属性名
	SelectSQL    string `json:"selectSQL"`    // 查询SQL
	Column       string `json:"column"`       // 关联列
	ForeignKey   string `json:"foreignKey"`   // 外键
	LazyLoad     bool   `json:"lazyLoad"`     // 是否懒加载
	CacheEnabled bool   `json:"cacheEnabled"` // 是否缓存
}

// CollectionInfo 集合信息
type CollectionInfo struct {
	PropertyName string `json:"propertyName"` // 属性名
	SelectSQL    string `json:"selectSQL"`    // 查询SQL
	Column       string `json:"column"`       // 关联列
	ForeignKey   string `json:"foreignKey"`   // 外键
	OfType       string `json:"ofType"`       // 元素类型
	LazyLoad     bool   `json:"lazyLoad"`     // 是否懒加载
	CacheEnabled bool   `json:"cacheEnabled"` // 是否缓存
}

// TagParser 标签解析器
type TagParser struct {
	mutex      sync.RWMutex
	tableCache map[reflect.Type]*TableInfo // 类型 -> 表信息缓存
}

// NewTagParser 创建标签解析器
func NewTagParser() *TagParser {
	return &TagParser{
		tableCache: make(map[reflect.Type]*TableInfo),
	}
}

// ParseStruct 解析结构体
func (p *TagParser) ParseStruct(obj any) (*TableInfo, error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Pointer {
		objType = objType.Elem()
	}

	if objType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", objType.Kind())
	}

	// 检查缓存
	p.mutex.RLock()
	if cached, exists := p.tableCache[objType]; exists {
		p.mutex.RUnlock()
		return cached, nil
	}
	p.mutex.RUnlock()

	// 解析表信息
	tableInfo := &TableInfo{
		Name:         p.parseTableName(objType),
		Columns:      make(map[string]*ColumnInfo),
		PrimaryKeys:  make([]string, 0),
		Associations: make([]*AssociationInfo, 0),
		Collections:  make([]*CollectionInfo, 0),
		CacheEnabled: p.parseCacheEnabled(objType),
	}

	// 解析字段
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		
		// 跳过未导出字段
		if !field.IsExported() {
			continue
		}

		err := p.parseField(field, tableInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %s: %w", field.Name, err)
		}
	}

	// 缓存结果
	p.mutex.Lock()
	p.tableCache[objType] = tableInfo
	p.mutex.Unlock()

	return tableInfo, nil
}

// parseTableName 解析表名
func (p *TagParser) parseTableName(objType reflect.Type) string {
	// 检查结构体标签
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if tableTag := field.Tag.Get(TagTable); tableTag != "" {
			return tableTag
		}
	}
	
	// 如果没有表标签，使用结构体名称的小写形式
	return strings.ToLower(objType.Name())
}

// parseCacheEnabled 解析缓存配置
func (p *TagParser) parseCacheEnabled(objType reflect.Type) bool {
	// 检查结构体级别的缓存标签
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if cacheTag := field.Tag.Get(TagCache); cacheTag != "" {
			enabled, _ := strconv.ParseBool(cacheTag)
			return enabled
		}
	}
	return false
}

// parseField 解析字段
func (p *TagParser) parseField(field reflect.StructField, tableInfo *TableInfo) error {
	// 检查是否忽略字段
	if ignoreTag := field.Tag.Get(TagIgnore); ignoreTag == "true" {
		return nil
	}

	// 解析关联字段
	if associationTag := field.Tag.Get(TagAssociation); associationTag != "" {
		return p.parseAssociation(field, associationTag, tableInfo)
	}

	// 解析集合字段
	if collectionTag := field.Tag.Get(TagCollection); collectionTag != "" {
		return p.parseCollection(field, collectionTag, tableInfo)
	}

	// 解析普通字段
	columnInfo := &ColumnInfo{
		FieldName:    field.Name,
		ColumnName:   p.parseColumnName(field),
		IsPrimaryKey: p.isPrimaryKey(field),
		IsAutoIncr:   p.isAutoIncrement(field),
		DataType:     field.Type.String(),
		IsIgnored:    false,
		DefaultValue: p.parseDefaultValue(field),
		Validation:   field.Tag.Get(TagValidation),
		CacheConfig:  field.Tag.Get(TagCache),
	}

	tableInfo.Columns[field.Name] = columnInfo

	// 添加到主键列表
	if columnInfo.IsPrimaryKey {
		tableInfo.PrimaryKeys = append(tableInfo.PrimaryKeys, columnInfo.ColumnName)
	}

	return nil
}

// parseColumnName 解析列名
func (p *TagParser) parseColumnName(field reflect.StructField) string {
	if columnTag := field.Tag.Get(TagColumn); columnTag != "" {
		return columnTag
	}
	// 默认使用字段名的小写形式
	return strings.ToLower(field.Name)
}

// isPrimaryKey 检查是否为主键
func (p *TagParser) isPrimaryKey(field reflect.StructField) bool {
	pkTag := field.Tag.Get(TagPrimaryKey)
	return pkTag == "true" || pkTag == "1"
}

// isAutoIncrement 检查是否为自增字段
func (p *TagParser) isAutoIncrement(field reflect.StructField) bool {
	autoIncrTag := field.Tag.Get(TagAutoIncr)
	return autoIncrTag == "true" || autoIncrTag == "1"
}

// parseDefaultValue 解析默认值
func (p *TagParser) parseDefaultValue(field reflect.StructField) any {
	defaultTag := field.Tag.Get("default")
	if defaultTag == "" {
		return nil
	}

	// 根据字段类型转换默认值
	switch field.Type.Kind() {
	case reflect.String:
		return defaultTag
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(defaultTag, 10, 64); err == nil {
			return val
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(defaultTag, 64); err == nil {
			return val
		}
	case reflect.Bool:
		if val, err := strconv.ParseBool(defaultTag); err == nil {
			return val
		}
	}
	
	return defaultTag
}

// parseAssociation 解析关联字段
func (p *TagParser) parseAssociation(field reflect.StructField, associationTag string, tableInfo *TableInfo) error {
	parts := strings.Split(associationTag, ",")
	association := &AssociationInfo{
		PropertyName: field.Name,
		LazyLoad:     field.Tag.Get(TagLazy) == "true",
		CacheEnabled: field.Tag.Get(TagCache) == "true",
	}

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "select":
			association.SelectSQL = value
		case "column":
			association.Column = value
		case "foreignKey":
			association.ForeignKey = value
		}
	}

	tableInfo.Associations = append(tableInfo.Associations, association)
	return nil
}

// parseCollection 解析集合字段
func (p *TagParser) parseCollection(field reflect.StructField, collectionTag string, tableInfo *TableInfo) error {
	parts := strings.Split(collectionTag, ",")
	collection := &CollectionInfo{
		PropertyName: field.Name,
		LazyLoad:     field.Tag.Get(TagLazy) == "true",
		CacheEnabled: field.Tag.Get(TagCache) == "true",
	}

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "select":
			collection.SelectSQL = value
		case "column":
			collection.Column = value
		case "foreignKey":
			collection.ForeignKey = value
		case "ofType":
			collection.OfType = value
		}
	}

	tableInfo.Collections = append(tableInfo.Collections, collection)
	return nil
}

// SQLGenerator SQL生成器
type SQLGenerator struct {
	parser *TagParser
}

// NewSQLGenerator 创建SQL生成器
func NewSQLGenerator() *SQLGenerator {
	return &SQLGenerator{
		parser: NewTagParser(),
	}
}

// GenerateInsertSQL 生成插入SQL
func (g *SQLGenerator) GenerateInsertSQL(obj any) (string, []any, error) {
	tableInfo, err := g.parser.ParseStruct(obj)
	if err != nil {
		return "", nil, err
	}

	var columns []string
	var placeholders []string
	var values []any

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Pointer {
		objValue = objValue.Elem()
	}

	for fieldName, columnInfo := range tableInfo.Columns {
		// 跳过自增主键
		if columnInfo.IsAutoIncr {
			continue
		}

		fieldValue := objValue.FieldByName(fieldName)
		if !fieldValue.IsValid() {
			continue
		}

		columns = append(columns, columnInfo.ColumnName)
		placeholders = append(placeholders, "?")
		values = append(values, fieldValue.Interface())
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableInfo.Name,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return sql, values, nil
}

// GenerateUpdateSQL 生成更新SQL
func (g *SQLGenerator) GenerateUpdateSQL(obj any) (string, []any, error) {
	tableInfo, err := g.parser.ParseStruct(obj)
	if err != nil {
		return "", nil, err
	}

	if len(tableInfo.PrimaryKeys) == 0 {
		return "", nil, fmt.Errorf("no primary key found for update")
	}

	var setParts []string
	var whereClause []string
	var values []any
	var whereValues []any

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Pointer {
		objValue = objValue.Elem()
	}

	for fieldName, columnInfo := range tableInfo.Columns {
		fieldValue := objValue.FieldByName(fieldName)
		if !fieldValue.IsValid() {
			continue
		}

		if columnInfo.IsPrimaryKey {
			whereClause = append(whereClause, columnInfo.ColumnName+" = ?")
			whereValues = append(whereValues, fieldValue.Interface())
		} else if !columnInfo.IsAutoIncr {
			setParts = append(setParts, columnInfo.ColumnName+" = ?")
			values = append(values, fieldValue.Interface())
		}
	}

	values = append(values, whereValues...)

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		tableInfo.Name,
		strings.Join(setParts, ", "),
		strings.Join(whereClause, " AND "))

	return sql, values, nil
}

// GenerateSelectSQL 生成查询SQL
func (g *SQLGenerator) GenerateSelectSQL(obj any, condition string) (string, error) {
	tableInfo, err := g.parser.ParseStruct(obj)
	if err != nil {
		return "", err
	}

	var columns []string
	for _, columnInfo := range tableInfo.Columns {
		columns = append(columns, columnInfo.ColumnName)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(columns, ", "),
		tableInfo.Name)

	if condition != "" {
		sql += " WHERE " + condition
	}

	return sql, nil
}

// GenerateDeleteSQL 生成删除SQL
func (g *SQLGenerator) GenerateDeleteSQL(obj any) (string, []any, error) {
	tableInfo, err := g.parser.ParseStruct(obj)
	if err != nil {
		return "", nil, err
	}

	if len(tableInfo.PrimaryKeys) == 0 {
		return "", nil, fmt.Errorf("no primary key found for delete")
	}

	var whereClause []string
	var values []any

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Pointer {
		objValue = objValue.Elem()
	}

	for fieldName, columnInfo := range tableInfo.Columns {
		if columnInfo.IsPrimaryKey {
			fieldValue := objValue.FieldByName(fieldName)
			if fieldValue.IsValid() {
				whereClause = append(whereClause, columnInfo.ColumnName+" = ?")
				values = append(values, fieldValue.Interface())
			}
		}
	}

	sql := fmt.Sprintf("DELETE FROM %s WHERE %s",
		tableInfo.Name,
		strings.Join(whereClause, " AND "))

	return sql, values, nil
}

// AnnotationDrivenSession 注解驱动的会话
type AnnotationDrivenSession struct {
	XMLSession
	sqlGenerator *SQLGenerator
	tagParser    *TagParser
}

// NewAnnotationDrivenSession 创建注解驱动会话
func NewAnnotationDrivenSession(session XMLSession) *AnnotationDrivenSession {
	return &AnnotationDrivenSession{
		XMLSession:   session,
		sqlGenerator: NewSQLGenerator(),
		tagParser:    NewTagParser(),
	}
}

// Insert 插入实体（基于注解）
func (s *AnnotationDrivenSession) Insert(ctx context.Context, entity any) (int64, error) {
	sql, args, err := s.sqlGenerator.GenerateInsertSQL(entity)
	if err != nil {
		return 0, fmt.Errorf("failed to generate insert SQL: %w", err)
	}

	return s.XMLSession.Insert(ctx, sql, args...)
}

// Update 更新实体（基于注解）
func (s *AnnotationDrivenSession) Update(ctx context.Context, entity any) (int64, error) {
	sql, args, err := s.sqlGenerator.GenerateUpdateSQL(entity)
	if err != nil {
		return 0, fmt.Errorf("failed to generate update SQL: %w", err)
	}

	return s.XMLSession.Update(ctx, sql, args...)
}

// Delete 删除实体（基于注解）
func (s *AnnotationDrivenSession) Delete(ctx context.Context, entity any) (int64, error) {
	sql, args, err := s.sqlGenerator.GenerateDeleteSQL(entity)
	if err != nil {
		return 0, fmt.Errorf("failed to generate delete SQL: %w", err)
	}

	return s.XMLSession.Delete(ctx, sql, args...)
}

// SelectByID 根据ID查询（基于注解）
func (s *AnnotationDrivenSession) SelectByID(ctx context.Context, entity any, id any) (any, error) {
	tableInfo, err := s.tagParser.ParseStruct(entity)
	if err != nil {
		return nil, err
	}

	if len(tableInfo.PrimaryKeys) == 0 {
		return nil, fmt.Errorf("no primary key found")
	}

	condition := tableInfo.PrimaryKeys[0] + " = ?"
	sql, err := s.sqlGenerator.GenerateSelectSQL(entity, condition)
	if err != nil {
		return nil, fmt.Errorf("failed to generate select SQL: %w", err)
	}

	return s.XMLSession.SelectOne(ctx, sql, id)
}

// SelectAll 查询所有记录（基于注解）
func (s *AnnotationDrivenSession) SelectAll(ctx context.Context, entity any) ([]any, error) {
	sql, err := s.sqlGenerator.GenerateSelectSQL(entity, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate select SQL: %w", err)
	}

	return s.XMLSession.SelectList(ctx, sql)
}

// SelectWhere 根据条件查询（基于注解）
func (s *AnnotationDrivenSession) SelectWhere(ctx context.Context, entity any, condition string, args ...any) ([]any, error) {
	sql, err := s.sqlGenerator.GenerateSelectSQL(entity, condition)
	if err != nil {
		return nil, fmt.Errorf("failed to generate select SQL: %w", err)
	}

	return s.XMLSession.SelectList(ctx, sql, args...)
}