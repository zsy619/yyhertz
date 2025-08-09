// Package mapper XML映射文件解析器
//
// 支持完整的MyBatis mapper.xml文件解析，包括：
// 1. select/insert/update/delete语句解析
// 2. resultMap复杂结果映射
// 3. 动态SQL标签解析
// 4. 参数类型处理
package mapper

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MapperXML mapper.xml文件的根结构
type MapperXML struct {
	XMLName    xml.Name       `xml:"mapper"`
	Namespace  string         `xml:"namespace,attr"`
	Selects    []SelectXML    `xml:"select"`
	Inserts    []InsertXML    `xml:"insert"`
	Updates    []UpdateXML    `xml:"update"`
	Deletes    []DeleteXML    `xml:"delete"`
	ResultMaps []ResultMapXML `xml:"resultMap"`
	SQLs       []SQLXML       `xml:"sql"`
}

// SelectXML select语句XML结构
type SelectXML struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr,omitempty"`
	ResultType    string `xml:"resultType,attr,omitempty"`
	ResultMap     string `xml:"resultMap,attr,omitempty"`
	UseCache      string `xml:"useCache,attr,omitempty"`
	Timeout       string `xml:"timeout,attr,omitempty"`
	Content       string `xml:",innerxml"`
}

// InsertXML insert语句XML结构
type InsertXML struct {
	ID               string `xml:"id,attr"`
	ParameterType    string `xml:"parameterType,attr,omitempty"`
	KeyProperty      string `xml:"keyProperty,attr,omitempty"`
	KeyColumn        string `xml:"keyColumn,attr,omitempty"`
	UseGeneratedKeys string `xml:"useGeneratedKeys,attr,omitempty"`
	Content          string `xml:",innerxml"`
}

// UpdateXML update语句XML结构
type UpdateXML struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr,omitempty"`
	Timeout       string `xml:"timeout,attr,omitempty"`
	Content       string `xml:",innerxml"`
}

// DeleteXML delete语句XML结构
type DeleteXML struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr,omitempty"`
	Timeout       string `xml:"timeout,attr,omitempty"`
	Content       string `xml:",innerxml"`
}

// ResultMapXML resultMap XML结构
type ResultMapXML struct {
	ID          string               `xml:"id,attr"`
	Type        string               `xml:"type,attr"`
	Extends     string               `xml:"extends,attr,omitempty"`
	AutoMap     string               `xml:"autoMapping,attr,omitempty"`
	IDs         []IDMapping          `xml:"id"`
	Results     []ResultMapping      `xml:"result"`
	Constructor []ConstructorMapping `xml:"constructor"`
	Collection  []CollectionMapping  `xml:"collection"`
	Association []AssociationMapping `xml:"association"`
}

// IDMapping ID字段映射
type IDMapping struct {
	Property string `xml:"property,attr"`
	Column   string `xml:"column,attr"`
	JavaType string `xml:"javaType,attr,omitempty"`
	JdbcType string `xml:"jdbcType,attr,omitempty"`
}

// ResultMapping 结果字段映射
type ResultMapping struct {
	Property string `xml:"property,attr"`
	Column   string `xml:"column,attr"`
	JavaType string `xml:"javaType,attr,omitempty"`
	JdbcType string `xml:"jdbcType,attr,omitempty"`
}

// ConstructorMapping 构造函数映射
type ConstructorMapping struct {
	Args []ConstructorArg `xml:"arg"`
}

// ConstructorArg 构造函数参数
type ConstructorArg struct {
	Column   string `xml:"column,attr,omitempty"`
	JavaType string `xml:"javaType,attr,omitempty"`
	JdbcType string `xml:"jdbcType,attr,omitempty"`
	Select   string `xml:"select,attr,omitempty"`
}

// CollectionMapping 集合映射
type CollectionMapping struct {
	Property  string `xml:"property,attr"`
	OfType    string `xml:"ofType,attr,omitempty"`
	Column    string `xml:"column,attr,omitempty"`
	Select    string `xml:"select,attr,omitempty"`
	ResultMap string `xml:"resultMap,attr,omitempty"`
}

// AssociationMapping 关联映射
type AssociationMapping struct {
	Property  string `xml:"property,attr"`
	JavaType  string `xml:"javaType,attr,omitempty"`
	Column    string `xml:"column,attr,omitempty"`
	Select    string `xml:"select,attr,omitempty"`
	ResultMap string `xml:"resultMap,attr,omitempty"`
}

// SQLXML 可重用SQL片段
type SQLXML struct {
	ID      string `xml:"id,attr"`
	Content string `xml:",innerxml"`
}

// MapperXMLParser XML解析器
type MapperXMLParser struct {
	namespace      string
	statements     map[string]*XMLMappedStatement
	resultMaps     map[string]*XMLResultMap
	sqlFragments   map[string]string
	dynamicBuilder *DynamicSqlBuilder
}

// XMLMappedStatement XML解析后的语句
type XMLMappedStatement struct {
	ID               string
	Namespace        string
	StatementType    StatementType
	ParameterType    string
	ResultType       string
	ResultMap        string
	SQL              string
	UseCache         bool
	Timeout          int
	KeyProperty      string
	KeyColumn        string
	UseGeneratedKeys bool
}

// XMLResultMap XML解析后的结果映射
type XMLResultMap struct {
	ID             string
	Type           string
	Namespace      string
	Extends        string
	AutoMap        bool
	IDMappings     []XMLColumnMapping
	ResultMappings []XMLColumnMapping
	Constructors   []XMLConstructorMapping
	Collections    []XMLCollectionMapping
	Associations   []XMLAssociationMapping
}

// XMLColumnMapping 列映射
type XMLColumnMapping struct {
	Property string
	Column   string
	JavaType string
	JdbcType string
}

// XMLConstructorMapping 构造函数映射
type XMLConstructorMapping struct {
	Args []XMLConstructorArg
}

// XMLConstructorArg 构造函数参数
type XMLConstructorArg struct {
	Column   string
	JavaType string
	JdbcType string
	Select   string
}

// XMLCollectionMapping 集合映射
type XMLCollectionMapping struct {
	Property  string
	OfType    string
	Column    string
	Select    string
	ResultMap string
}

// XMLAssociationMapping 关联映射
type XMLAssociationMapping struct {
	Property  string
	JavaType  string
	Column    string
	Select    string
	ResultMap string
}

// StatementType 语句类型
type StatementType int

const (
	StatementTypeSelect StatementType = iota
	StatementTypeInsert
	StatementTypeUpdate
	StatementTypeDelete
)

func (st StatementType) String() string {
	switch st {
	case StatementTypeSelect:
		return "SELECT"
	case StatementTypeInsert:
		return "INSERT"
	case StatementTypeUpdate:
		return "UPDATE"
	case StatementTypeDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// NewMapperXMLParser 创建XML解析器
func NewMapperXMLParser() *MapperXMLParser {
	return &MapperXMLParser{
		statements:     make(map[string]*XMLMappedStatement),
		resultMaps:     make(map[string]*XMLResultMap),
		sqlFragments:   make(map[string]string),
		dynamicBuilder: NewDynamicSqlBuilder(),
	}
}

// ParseXMLFile 解析XML文件
func (parser *MapperXMLParser) ParseXMLFile(xmlPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		return fmt.Errorf("mapper XML file not found: %s", xmlPath)
	}

	// 读取XML文件
	file, err := os.Open(xmlPath)
	if err != nil {
		return fmt.Errorf("failed to open XML file: %w", err)
	}
	defer file.Close()

	return parser.ParseXMLReader(file)
}

// ParseXMLReader 从Reader解析XML
func (parser *MapperXMLParser) ParseXMLReader(reader io.Reader) error {
	var mapperXML MapperXML

	// 解析XML
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&mapperXML); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	// 设置命名空间
	parser.namespace = mapperXML.Namespace

	// 解析SQL片段
	if err := parser.parseSQLFragments(mapperXML.SQLs); err != nil {
		return fmt.Errorf("failed to parse SQL fragments: %w", err)
	}

	// 解析ResultMap
	if err := parser.parseResultMaps(mapperXML.ResultMaps); err != nil {
		return fmt.Errorf("failed to parse ResultMaps: %w", err)
	}

	// 解析语句
	if err := parser.parseStatements(mapperXML); err != nil {
		return fmt.Errorf("failed to parse statements: %w", err)
	}

	return nil
}

// parseSQLFragments 解析SQL片段
func (parser *MapperXMLParser) parseSQLFragments(sqls []SQLXML) error {
	for _, sql := range sqls {
		fragmentKey := parser.namespace + "." + sql.ID
		parser.sqlFragments[fragmentKey] = strings.TrimSpace(sql.Content)
	}
	return nil
}

// parseResultMaps 解析ResultMap
func (parser *MapperXMLParser) parseResultMaps(resultMaps []ResultMapXML) error {
	for _, rm := range resultMaps {
		resultMap := &XMLResultMap{
			ID:        rm.ID,
			Type:      rm.Type,
			Namespace: parser.namespace,
			Extends:   rm.Extends,
			AutoMap:   rm.AutoMap == "true",
		}

		// 解析ID映射
		for _, idMapping := range rm.IDs {
			resultMap.IDMappings = append(resultMap.IDMappings, XMLColumnMapping{
				Property: idMapping.Property,
				Column:   idMapping.Column,
				JavaType: idMapping.JavaType,
				JdbcType: idMapping.JdbcType,
			})
		}

		// 解析结果映射
		for _, resultMapping := range rm.Results {
			resultMap.ResultMappings = append(resultMap.ResultMappings, XMLColumnMapping{
				Property: resultMapping.Property,
				Column:   resultMapping.Column,
				JavaType: resultMapping.JavaType,
				JdbcType: resultMapping.JdbcType,
			})
		}

		// 解析构造函数映射
		for _, constructor := range rm.Constructor {
			var args []XMLConstructorArg
			for _, arg := range constructor.Args {
				args = append(args, XMLConstructorArg{
					Column:   arg.Column,
					JavaType: arg.JavaType,
					JdbcType: arg.JdbcType,
					Select:   arg.Select,
				})
			}
			resultMap.Constructors = append(resultMap.Constructors, XMLConstructorMapping{Args: args})
		}

		// 解析集合映射
		for _, collection := range rm.Collection {
			resultMap.Collections = append(resultMap.Collections, XMLCollectionMapping{
				Property:  collection.Property,
				OfType:    collection.OfType,
				Column:    collection.Column,
				Select:    collection.Select,
				ResultMap: collection.ResultMap,
			})
		}

		// 解析关联映射
		for _, association := range rm.Association {
			resultMap.Associations = append(resultMap.Associations, XMLAssociationMapping{
				Property:  association.Property,
				JavaType:  association.JavaType,
				Column:    association.Column,
				Select:    association.Select,
				ResultMap: association.ResultMap,
			})
		}

		// 存储ResultMap
		resultMapKey := parser.namespace + "." + rm.ID
		parser.resultMaps[resultMapKey] = resultMap
	}
	return nil
}

// parseStatements 解析所有语句
func (parser *MapperXMLParser) parseStatements(mapperXML MapperXML) error {
	// 解析SELECT语句
	for _, selectStmt := range mapperXML.Selects {
		stmt, err := parser.parseSelectStatement(selectStmt)
		if err != nil {
			return fmt.Errorf("failed to parse select statement %s: %w", selectStmt.ID, err)
		}
		statementKey := parser.namespace + "." + selectStmt.ID
		parser.statements[statementKey] = stmt
	}

	// 解析INSERT语句
	for _, insertStmt := range mapperXML.Inserts {
		stmt, err := parser.parseInsertStatement(insertStmt)
		if err != nil {
			return fmt.Errorf("failed to parse insert statement %s: %w", insertStmt.ID, err)
		}
		statementKey := parser.namespace + "." + insertStmt.ID
		parser.statements[statementKey] = stmt
	}

	// 解析UPDATE语句
	for _, updateStmt := range mapperXML.Updates {
		stmt, err := parser.parseUpdateStatement(updateStmt)
		if err != nil {
			return fmt.Errorf("failed to parse update statement %s: %w", updateStmt.ID, err)
		}
		statementKey := parser.namespace + "." + updateStmt.ID
		parser.statements[statementKey] = stmt
	}

	// 解析DELETE语句
	for _, deleteStmt := range mapperXML.Deletes {
		stmt, err := parser.parseDeleteStatement(deleteStmt)
		if err != nil {
			return fmt.Errorf("failed to parse delete statement %s: %w", deleteStmt.ID, err)
		}
		statementKey := parser.namespace + "." + deleteStmt.ID
		parser.statements[statementKey] = stmt
	}

	return nil
}

// parseSelectStatement 解析SELECT语句
func (parser *MapperXMLParser) parseSelectStatement(selectXML SelectXML) (*XMLMappedStatement, error) {
	sql, err := parser.processSQLContent(selectXML.Content)
	if err != nil {
		return nil, err
	}

	stmt := &XMLMappedStatement{
		ID:            selectXML.ID,
		Namespace:     parser.namespace,
		StatementType: StatementTypeSelect,
		ParameterType: selectXML.ParameterType,
		ResultType:    selectXML.ResultType,
		ResultMap:     selectXML.ResultMap,
		SQL:           sql,
		UseCache:      selectXML.UseCache == "true" || selectXML.UseCache == "",
	}

	// 解析timeout
	if selectXML.Timeout != "" {
		// 这里可以添加timeout解析逻辑
	}

	return stmt, nil
}

// parseInsertStatement 解析INSERT语句
func (parser *MapperXMLParser) parseInsertStatement(insertXML InsertXML) (*XMLMappedStatement, error) {
	sql, err := parser.processSQLContent(insertXML.Content)
	if err != nil {
		return nil, err
	}

	stmt := &XMLMappedStatement{
		ID:               insertXML.ID,
		Namespace:        parser.namespace,
		StatementType:    StatementTypeInsert,
		ParameterType:    insertXML.ParameterType,
		SQL:              sql,
		KeyProperty:      insertXML.KeyProperty,
		KeyColumn:        insertXML.KeyColumn,
		UseGeneratedKeys: insertXML.UseGeneratedKeys == "true",
	}

	return stmt, nil
}

// parseUpdateStatement 解析UPDATE语句
func (parser *MapperXMLParser) parseUpdateStatement(updateXML UpdateXML) (*XMLMappedStatement, error) {
	sql, err := parser.processSQLContent(updateXML.Content)
	if err != nil {
		return nil, err
	}

	stmt := &XMLMappedStatement{
		ID:            updateXML.ID,
		Namespace:     parser.namespace,
		StatementType: StatementTypeUpdate,
		ParameterType: updateXML.ParameterType,
		SQL:           sql,
	}

	return stmt, nil
}

// parseDeleteStatement 解析DELETE语句
func (parser *MapperXMLParser) parseDeleteStatement(deleteXML DeleteXML) (*XMLMappedStatement, error) {
	sql, err := parser.processSQLContent(deleteXML.Content)
	if err != nil {
		return nil, err
	}

	stmt := &XMLMappedStatement{
		ID:            deleteXML.ID,
		Namespace:     parser.namespace,
		StatementType: StatementTypeDelete,
		ParameterType: deleteXML.ParameterType,
		SQL:           sql,
	}

	return stmt, nil
}

// processSQLContent 处理SQL内容，包括include和动态SQL
func (parser *MapperXMLParser) processSQLContent(content string) (string, error) {
	content = strings.TrimSpace(content)

	// 处理include标签
	content = parser.processIncludes(content)

	// 如果包含动态SQL标签，保持原样，留给动态SQL解析器处理
	if parser.containsDynamicSQL(content) {
		return content, nil
	}

	return content, nil
}

// processIncludes 处理include标签
func (parser *MapperXMLParser) processIncludes(content string) string {
	// 简单的include处理，查找<include refid="xxx"/>并替换
	// 这里可以用更复杂的XML解析，简化实现使用字符串替换

	// includePattern := `<include\s+refid="([^"]+)"\s*/>`
	// TODO: 实现include标签的处理
	return content // 简化实现，实际需要正则替换
}

// containsDynamicSQL 检查是否包含动态SQL标签
func (parser *MapperXMLParser) containsDynamicSQL(content string) bool {
	dynamicTags := []string{"<if", "<where", "<set", "<choose", "<foreach", "<trim", "<bind"}
	for _, tag := range dynamicTags {
		if strings.Contains(content, tag) {
			return true
		}
	}
	return false
}

// GetStatement 获取语句
func (parser *MapperXMLParser) GetStatement(statementId string) *XMLMappedStatement {
	return parser.statements[statementId]
}

// GetAllStatements 获取所有语句
func (parser *MapperXMLParser) GetAllStatements() map[string]*XMLMappedStatement {
	result := make(map[string]*XMLMappedStatement)
	for k, v := range parser.statements {
		result[k] = v
	}
	return result
}

// GetResultMap 获取ResultMap
func (parser *MapperXMLParser) GetResultMap(resultMapId string) *XMLResultMap {
	return parser.resultMaps[resultMapId]
}

// GetAllResultMaps 获取所有ResultMap
func (parser *MapperXMLParser) GetAllResultMaps() map[string]*XMLResultMap {
	result := make(map[string]*XMLResultMap)
	for k, v := range parser.resultMaps {
		result[k] = v
	}
	return result
}

// GetNamespace 获取命名空间
func (parser *MapperXMLParser) GetNamespace() string {
	return parser.namespace
}

// LoadMapperDirectory 批量加载目录下的所有mapper.xml文件
func LoadMapperDirectory(dirPath string) (map[string]*MapperXMLParser, error) {
	parsers := make(map[string]*MapperXMLParser)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理.xml文件
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".xml") {
			return nil
		}

		parser := NewMapperXMLParser()
		if err := parser.ParseXMLFile(path); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		parsers[parser.GetNamespace()] = parser
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load mapper directory: %w", err)
	}

	return parsers, nil
}
