// Package mybatis XML Mapper会话支持
//
// 扩展SimpleSession以支持XML Mapper文件，提供与Java MyBatis兼容的使用体验
package mybatis

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"github.com/zsy619/yyhertz/framework/mybatis/mapper"
)

// XMLSession XML支持的会话接口
type XMLSession interface {
	SimpleSession
	
	// XML映射器管理
	LoadMapperXML(xmlPath string) error
	LoadMapperXMLFromString(xmlContent string) error
	LoadMapperDirectory(dirPath string) error
	
	// 通过语句ID执行（MyBatis风格）
	SelectOneByID(ctx context.Context, statementId string, parameter interface{}) (interface{}, error)
	SelectListByID(ctx context.Context, statementId string, parameter interface{}) ([]interface{}, error)
	SelectPageByID(ctx context.Context, statementId string, parameter interface{}, page PageRequest) (*PageResult, error)
	InsertByID(ctx context.Context, statementId string, parameter interface{}) (int64, error)
	UpdateByID(ctx context.Context, statementId string, parameter interface{}) (int64, error)
	DeleteByID(ctx context.Context, statementId string, parameter interface{}) (int64, error)
	
	// Mapper信息查询
	GetStatement(statementId string) *mapper.XMLMappedStatement
	GetResultMap(resultMapId string) *mapper.XMLResultMap
	GetNamespaces() []string
	GetStatementIds(namespace string) []string
}

// xmlSession XML会话实现
type xmlSession struct {
	SimpleSession
	parsers       map[string]*mapper.MapperXMLParser  // namespace -> parser
	dynamicBuilder *mapper.DynamicSqlBuilder
}

// NewXMLSession 创建支持XML的会话
func NewXMLSession(db *gorm.DB) XMLSession {
	return &xmlSession{
		SimpleSession:  NewSimpleSession(db),
		parsers:        make(map[string]*mapper.MapperXMLParser),
		dynamicBuilder: mapper.NewDynamicSqlBuilder(),
	}
}

// NewXMLSessionWithHooks 创建带钩子的XML会话
func NewXMLSessionWithHooks(db *gorm.DB, enableDebug bool) XMLSession {
	return &xmlSession{
		SimpleSession:  NewSimpleWithHooks(db, enableDebug),
		parsers:        make(map[string]*mapper.MapperXMLParser),
		dynamicBuilder: mapper.NewDynamicSqlBuilder(),
	}
}

// LoadMapperXML 加载XML映射文件
func (xs *xmlSession) LoadMapperXML(xmlPath string) error {
	parser := mapper.NewMapperXMLParser()
	if err := parser.ParseXMLFile(xmlPath); err != nil {
		return fmt.Errorf("failed to load mapper XML %s: %w", xmlPath, err)
	}
	
	namespace := parser.GetNamespace()
	if namespace == "" {
		return fmt.Errorf("mapper XML file must have a namespace: %s", xmlPath)
	}
	
	xs.parsers[namespace] = parser
	log.Printf("[XML Mapper] Loaded mapper: %s with %d statements", namespace, len(parser.GetAllStatements()))
	
	return nil
}

// LoadMapperXMLFromString 从字符串加载XML映射
func (xs *xmlSession) LoadMapperXMLFromString(xmlContent string) error {
	parser := mapper.NewMapperXMLParser()
	if err := parser.ParseXMLReader(strings.NewReader(xmlContent)); err != nil {
		return fmt.Errorf("failed to parse mapper XML content: %w", err)
	}
	
	namespace := parser.GetNamespace()
	if namespace == "" {
		return fmt.Errorf("mapper XML content must have a namespace")
	}
	
	xs.parsers[namespace] = parser
	log.Printf("[XML Mapper] Loaded mapper from string: %s with %d statements", namespace, len(parser.GetAllStatements()))
	
	return nil
}

// LoadMapperDirectory 批量加载目录下的映射文件
func (xs *xmlSession) LoadMapperDirectory(dirPath string) error {
	parsers, err := mapper.LoadMapperDirectory(dirPath)
	if err != nil {
		return fmt.Errorf("failed to load mapper directory %s: %w", dirPath, err)
	}
	
	for namespace, parser := range parsers {
		xs.parsers[namespace] = parser
		log.Printf("[XML Mapper] Loaded mapper: %s with %d statements", namespace, len(parser.GetAllStatements()))
	}
	
	log.Printf("[XML Mapper] Total loaded %d mappers from directory: %s", len(parsers), dirPath)
	return nil
}

// SelectOneByID 通过语句ID查询单条记录
func (xs *xmlSession) SelectOneByID(ctx context.Context, statementId string, parameter interface{}) (interface{}, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return nil, fmt.Errorf("statement not found: %s", statementId)
	}
	
	if stmt.StatementType != mapper.StatementTypeSelect {
		return nil, fmt.Errorf("statement %s is not a SELECT statement", statementId)
	}
	
	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}
	
	// 执行查询
	result, err := xs.SimpleSession.SelectOne(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	
	// 应用ResultMap映射（如果有的话）
	if stmt.ResultMap != "" {
		return xs.applyResultMap(result, stmt.ResultMap)
	}
	
	return result, nil
}

// SelectListByID 通过语句ID查询多条记录
func (xs *xmlSession) SelectListByID(ctx context.Context, statementId string, parameter interface{}) ([]interface{}, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return nil, fmt.Errorf("statement not found: %s", statementId)
	}
	
	if stmt.StatementType != mapper.StatementTypeSelect {
		return nil, fmt.Errorf("statement %s is not a SELECT statement", statementId)
	}
	
	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}
	
	// 执行查询
	results, err := xs.SimpleSession.SelectList(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	
	// 应用ResultMap映射（如果有的话）
	if stmt.ResultMap != "" {
		mappedResults := make([]interface{}, len(results))
		for i, result := range results {
			mapped, err := xs.applyResultMap(result, stmt.ResultMap)
			if err != nil {
				return nil, fmt.Errorf("failed to apply result map: %w", err)
			}
			mappedResults[i] = mapped
		}
		return mappedResults, nil
	}
	
	return results, nil
}

// SelectPageByID 通过语句ID分页查询
func (xs *xmlSession) SelectPageByID(ctx context.Context, statementId string, parameter interface{}, page PageRequest) (*PageResult, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return nil, fmt.Errorf("statement not found: %s", statementId)
	}
	
	if stmt.StatementType != mapper.StatementTypeSelect {
		return nil, fmt.Errorf("statement %s is not a SELECT statement", statementId)
	}
	
	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}
	
	// 执行分页查询
	pageResult, err := xs.SimpleSession.SelectPage(ctx, sql, page, args...)
	if err != nil {
		return nil, err
	}
	
	// 应用ResultMap映射（如果有的话）
	if stmt.ResultMap != "" {
		mappedItems := make([]interface{}, len(pageResult.Items))
		for i, item := range pageResult.Items {
			mapped, err := xs.applyResultMap(item, stmt.ResultMap)
			if err != nil {
				return nil, fmt.Errorf("failed to apply result map: %w", err)
			}
			mappedItems[i] = mapped
		}
		pageResult.Items = mappedItems
	}
	
	return pageResult, nil
}

// InsertByID 通过语句ID插入记录
func (xs *xmlSession) InsertByID(ctx context.Context, statementId string, parameter interface{}) (int64, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return 0, fmt.Errorf("statement not found: %s", statementId)
	}
	
	if stmt.StatementType != mapper.StatementTypeInsert {
		return 0, fmt.Errorf("statement %s is not an INSERT statement", statementId)
	}
	
	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return 0, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}
	
	return xs.SimpleSession.Insert(ctx, sql, args...)
}

// UpdateByID 通过语句ID更新记录
func (xs *xmlSession) UpdateByID(ctx context.Context, statementId string, parameter interface{}) (int64, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return 0, fmt.Errorf("statement not found: %s", statementId)
	}
	
	if stmt.StatementType != mapper.StatementTypeUpdate {
		return 0, fmt.Errorf("statement %s is not an UPDATE statement", statementId)
	}
	
	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return 0, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}
	
	return xs.SimpleSession.Update(ctx, sql, args...)
}

// DeleteByID 通过语句ID删除记录
func (xs *xmlSession) DeleteByID(ctx context.Context, statementId string, parameter interface{}) (int64, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return 0, fmt.Errorf("statement not found: %s", statementId)
	}
	
	if stmt.StatementType != mapper.StatementTypeDelete {
		return 0, fmt.Errorf("statement %s is not a DELETE statement", statementId)
	}
	
	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return 0, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}
	
	return xs.SimpleSession.Delete(ctx, sql, args...)
}

// GetStatement 获取语句定义
func (xs *xmlSession) GetStatement(statementId string) *mapper.XMLMappedStatement {
	return xs.getStatementByID(statementId)
}

// GetResultMap 获取ResultMap定义
func (xs *xmlSession) GetResultMap(resultMapId string) *mapper.XMLResultMap {
	return xs.getResultMapByID(resultMapId)
}

// GetNamespaces 获取所有已加载的命名空间
func (xs *xmlSession) GetNamespaces() []string {
	namespaces := make([]string, 0, len(xs.parsers))
	for namespace := range xs.parsers {
		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

// GetStatementIds 获取指定命名空间下的所有语句ID
func (xs *xmlSession) GetStatementIds(namespace string) []string {
	parser, exists := xs.parsers[namespace]
	if !exists {
		return []string{}
	}
	
	statements := parser.GetAllStatements()
	ids := make([]string, 0, len(statements))
	for statementId := range statements {
		ids = append(ids, statementId)
	}
	return ids
}

// 内部方法

// getStatementByID 通过ID获取语句
func (xs *xmlSession) getStatementByID(statementId string) *mapper.XMLMappedStatement {
	parts := strings.SplitN(statementId, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	
	namespace := parts[0]
	parser, exists := xs.parsers[namespace]
	if !exists {
		return nil
	}
	
	return parser.GetStatement(statementId)
}

// getResultMapByID 通过ID获取ResultMap
func (xs *xmlSession) getResultMapByID(resultMapId string) *mapper.XMLResultMap {
	parts := strings.SplitN(resultMapId, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	
	namespace := parts[0]
	parser, exists := xs.parsers[namespace]
	if !exists {
		return nil
	}
	
	return parser.GetResultMap(resultMapId)
}

// buildSQL 构建最终的SQL语句
func (xs *xmlSession) buildSQL(stmt *mapper.XMLMappedStatement, parameter interface{}) (string, []interface{}, error) {
	sql := stmt.SQL
	
	// 检查是否包含动态SQL
	if xs.containsDynamicSQL(sql) {
		// 使用动态SQL构建器
		builtSQL, args, err := xs.dynamicBuilder.Build(sql, parameter)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build dynamic SQL: %w", err)
		}
		return builtSQL, args, nil
	} else {
		// 静态SQL，直接处理参数占位符
		builtSQL, args, err := xs.processStaticSQL(sql, parameter)
		if err != nil {
			return "", nil, fmt.Errorf("failed to process static SQL: %w", err)
		}
		return builtSQL, args, nil
	}
}

// containsDynamicSQL 检查是否包含动态SQL标签
func (xs *xmlSession) containsDynamicSQL(sql string) bool {
	dynamicTags := []string{"<if", "<where", "<set", "<choose", "<foreach", "<trim", "<bind"}
	for _, tag := range dynamicTags {
		if strings.Contains(sql, tag) {
			return true
		}
	}
	return false
}

// processStaticSQL 处理静态SQL的参数占位符
func (xs *xmlSession) processStaticSQL(sql string, parameter interface{}) (string, []interface{}, error) {
	args := make([]interface{}, 0)
	
	// 简单的#{param}占位符替换
	result := sql
	
	// 查找所有#{xxx}占位符
	paramPattern := `#\{([^}]+)\}`
	matches := regexp.MustCompile(paramPattern).FindAllStringSubmatch(sql, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			paramName := match[1]
			value := xs.getParameterValue(parameter, paramName)
			args = append(args, value)
		}
	}
	
	// 将所有#{xxx}替换为?
	result = regexp.MustCompile(paramPattern).ReplaceAllString(result, "?")
	
	return result, args, nil
}

// getParameterValue 从参数对象中获取指定属性的值
func (xs *xmlSession) getParameterValue(parameter interface{}, propertyPath string) interface{} {
	if parameter == nil {
		return nil
	}
	
	// 如果是map类型
	if paramMap, ok := parameter.(map[string]interface{}); ok {
		return paramMap[propertyPath]
	}
	
	// 使用反射获取结构体字段值
	return xs.getFieldValue(parameter, propertyPath)
}

// getFieldValue 使用反射获取字段值
func (xs *xmlSession) getFieldValue(obj interface{}, fieldPath string) interface{} {
	if obj == nil {
		return nil
	}
	
	parts := strings.Split(fieldPath, ".")
	current := obj
	
	for _, part := range parts {
		if current == nil {
			return nil
		}
		
		v := reflect.ValueOf(current)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		
		if v.Kind() != reflect.Struct {
			return nil
		}
		
		field := v.FieldByName(part)
		if !field.IsValid() || !field.CanInterface() {
			return nil
		}
		
		current = field.Interface()
	}
	
	return current
}

// applyResultMap 应用ResultMap映射
func (xs *xmlSession) applyResultMap(result interface{}, resultMapId string) (interface{}, error) {
	resultMap := xs.getResultMapByID(resultMapId)
	if resultMap == nil {
		return result, nil // 如果没有找到ResultMap，直接返回原结果
	}
	
	// 简化实现：只处理基本的列映射
	if resultData, ok := result.(map[string]interface{}); ok {
		mappedResult := make(map[string]interface{})
		
		// 应用ID映射
		for _, idMapping := range resultMap.IDMappings {
			if value, exists := resultData[idMapping.Column]; exists {
				mappedResult[idMapping.Property] = xs.convertValue(value, idMapping.JavaType)
			}
		}
		
		// 应用结果映射
		for _, mapping := range resultMap.ResultMappings {
			if value, exists := resultData[mapping.Column]; exists {
				mappedResult[mapping.Property] = xs.convertValue(value, mapping.JavaType)
			}
		}
		
		// 如果启用了自动映射，复制未明确映射的字段
		if resultMap.AutoMap {
			for column, value := range resultData {
				if _, exists := mappedResult[column]; !exists {
					mappedResult[column] = value
				}
			}
		}
		
		return mappedResult, nil
	}
	
	return result, nil
}

// convertValue 根据JavaType转换值
func (xs *xmlSession) convertValue(value interface{}, javaType string) interface{} {
	if javaType == "" || value == nil {
		return value
	}
	
	switch strings.ToLower(javaType) {
	case "string", "java.lang.string":
		return fmt.Sprintf("%v", value)
	case "int", "integer", "java.lang.integer":
		if str, ok := value.(string); ok {
			if i, err := strconv.Atoi(str); err == nil {
				return i
			}
		}
		return value
	case "long", "java.lang.long":
		if str, ok := value.(string); ok {
			if l, err := strconv.ParseInt(str, 10, 64); err == nil {
				return l
			}
		}
		return value
	case "double", "java.lang.double":
		if str, ok := value.(string); ok {
			if d, err := strconv.ParseFloat(str, 64); err == nil {
				return d
			}
		}
		return value
	case "boolean", "java.lang.boolean":
		if str, ok := value.(string); ok {
			if b, err := strconv.ParseBool(str); err == nil {
				return b
			}
		}
		return value
	case "date", "java.util.date":
		if str, ok := value.(string); ok {
			// 尝试解析常见的日期格式
			formats := []string{
				"2006-01-02 15:04:05",
				"2006-01-02",
				"15:04:05",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, str); err == nil {
					return t
				}
			}
		}
		return value
	default:
		return value
	}
}