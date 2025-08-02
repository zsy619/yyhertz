// Package mybat XML映射器加载器
//
// 提供从XML文件加载MyBatis映射配置的功能
package mybat

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/zsy619/yyhertz/framework/mybatis/config"
)

// XMLMapperLoader XML映射器加载器
type XMLMapperLoader struct {
	basePath string
	config   *config.Configuration
}

// NewXMLMapperLoader 创建XML映射器加载器
func NewXMLMapperLoader(basePath string, config *config.Configuration) *XMLMapperLoader {
	return &XMLMapperLoader{
		basePath: basePath,
		config:   config,
	}
}

// Mapper XML映射器定义
type Mapper struct {
	XMLName   xml.Name `xml:"mapper"`
	Namespace string   `xml:"namespace,attr"`
	
	// 结果映射
	ResultMaps []ResultMap `xml:"resultMap"`
	
	// SQL片段
	SqlFragments []SqlFragment `xml:"sql"`
	
	// 查询语句
	Selects []Select `xml:"select"`
	
	// 插入语句
	Inserts []Insert `xml:"insert"`
	
	// 更新语句
	Updates []Update `xml:"update"`
	
	// 删除语句
	Deletes []Delete `xml:"delete"`
}

// ResultMap 结果映射
type ResultMap struct {
	ID   string `xml:"id,attr"`
	Type string `xml:"type,attr"`
	
	// ID映射
	IDMappings []IDMapping `xml:"id"`
	
	// 结果映射
	Results []Result `xml:"result"`
	
	// 关联映射
	Associations []Association `xml:"association"`
	
	// 集合映射
	Collections []Collection `xml:"collection"`
}

// IDMapping ID映射
type IDMapping struct {
	Column   string `xml:"column,attr"`
	Property string `xml:"property,attr"`
	JdbcType string `xml:"jdbcType,attr"`
}

// Result 结果映射
type Result struct {
	Column   string `xml:"column,attr"`
	Property string `xml:"property,attr"`
	JdbcType string `xml:"jdbcType,attr"`
}

// Association 关联映射
type Association struct {
	Property string `xml:"property,attr"`
	JavaType string `xml:"javaType,attr"`
	
	// 嵌套ID映射
	IDMappings []IDMapping `xml:"id"`
	
	// 嵌套结果映射
	Results []Result `xml:"result"`
}

// Collection 集合映射
type Collection struct {
	Property string `xml:"property,attr"`
	OfType   string `xml:"ofType,attr"`
	
	// 嵌套ID映射
	IDMappings []IDMapping `xml:"id"`
	
	// 嵌套结果映射
	Results []Result `xml:"result"`
}

// SqlFragment SQL片段
type SqlFragment struct {
	ID      string `xml:"id,attr"`
	Content string `xml:",innerxml"`
}

// Select 查询语句
type Select struct {
	ID              string `xml:"id,attr"`
	ParameterType   string `xml:"parameterType,attr"`
	ResultType      string `xml:"resultType,attr"`
	ResultMap       string `xml:"resultMap,attr"`
	StatementType   string `xml:"statementType,attr"`
	UseCache        string `xml:"useCache,attr"`
	FlushCache      string `xml:"flushCache,attr"`
	Timeout         string `xml:"timeout,attr"`
	FetchSize       string `xml:"fetchSize,attr"`
	ResultSetType   string `xml:"resultSetType,attr"`
	Content         string `xml:",innerxml"`
}

// Insert 插入语句
type Insert struct {
	ID                 string `xml:"id,attr"`
	ParameterType      string `xml:"parameterType,attr"`
	UseGeneratedKeys   string `xml:"useGeneratedKeys,attr"`
	KeyProperty        string `xml:"keyProperty,attr"`
	KeyColumn          string `xml:"keyColumn,attr"`
	StatementType      string `xml:"statementType,attr"`
	FlushCache         string `xml:"flushCache,attr"`
	Timeout            string `xml:"timeout,attr"`
	Content            string `xml:",innerxml"`
}

// Update 更新语句
type Update struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr"`
	StatementType string `xml:"statementType,attr"`
	FlushCache    string `xml:"flushCache,attr"`
	Timeout       string `xml:"timeout,attr"`
	Content       string `xml:",innerxml"`
}

// Delete 删除语句
type Delete struct {
	ID            string `xml:"id,attr"`
	ParameterType string `xml:"parameterType,attr"`
	StatementType string `xml:"statementType,attr"`
	FlushCache    string `xml:"flushCache,attr"`
	Timeout       string `xml:"timeout,attr"`
	Content       string `xml:",innerxml"`
}

// LoadMapper 从XML文件加载映射器
func (loader *XMLMapperLoader) LoadMapper(filename string) (*Mapper, error) {
	filePath := filepath.Join(loader.basePath, filename)
	
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapper file %s: %w", filePath, err)
	}
	
	var mapper Mapper
	if err := xml.Unmarshal(data, &mapper); err != nil {
		return nil, fmt.Errorf("failed to parse mapper XML %s: %w", filePath, err)
	}
	
	return &mapper, nil
}

// LoadAllMappers 加载所有映射器文件
func (loader *XMLMapperLoader) LoadAllMappers() ([]*Mapper, error) {
	pattern := filepath.Join(loader.basePath, "*.xml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find mapper files: %w", err)
	}
	
	var mappers []*Mapper
	for _, file := range files {
		filename := filepath.Base(file)
		mapper, err := loader.LoadMapper(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load mapper %s: %w", filename, err)
		}
		mappers = append(mappers, mapper)
	}
	
	return mappers, nil
}

// RegisterMapper 注册映射器到配置中
func (loader *XMLMapperLoader) RegisterMapper(mapper *Mapper) error {
	// 注册结果映射
	for _, resultMap := range mapper.ResultMaps {
		if err := loader.registerResultMap(mapper.Namespace, &resultMap); err != nil {
			return fmt.Errorf("failed to register result map %s: %w", resultMap.ID, err)
		}
	}
	
	// 注册SQL片段
	for _, sqlFragment := range mapper.SqlFragments {
		if err := loader.registerSqlFragment(mapper.Namespace, &sqlFragment); err != nil {
			return fmt.Errorf("failed to register SQL fragment %s: %w", sqlFragment.ID, err)
		}
	}
	
	// 注册语句
	for _, stmt := range mapper.Selects {
		if err := loader.registerStatement(mapper.Namespace, stmt.ID, &stmt, config.SqlCommandTypeSelect); err != nil {
			return fmt.Errorf("failed to register select statement %s: %w", stmt.ID, err)
		}
	}
	
	for _, stmt := range mapper.Inserts {
		if err := loader.registerStatement(mapper.Namespace, stmt.ID, &stmt, config.SqlCommandTypeInsert); err != nil {
			return fmt.Errorf("failed to register insert statement %s: %w", stmt.ID, err)
		}
	}
	
	for _, stmt := range mapper.Updates {
		if err := loader.registerStatement(mapper.Namespace, stmt.ID, &stmt, config.SqlCommandTypeUpdate); err != nil {
			return fmt.Errorf("failed to register update statement %s: %w", stmt.ID, err)
		}
	}
	
	for _, stmt := range mapper.Deletes {
		if err := loader.registerStatement(mapper.Namespace, stmt.ID, &stmt, config.SqlCommandTypeDelete); err != nil {
			return fmt.Errorf("failed to register delete statement %s: %w", stmt.ID, err)
		}
	}
	
	return nil
}

// registerResultMap 注册结果映射
func (loader *XMLMapperLoader) registerResultMap(namespace string, resultMap *ResultMap) error {
	// 这里实现结果映射的注册逻辑
	// 转换XML结果映射为框架内部的结果映射对象
	return nil
}

// registerSqlFragment 注册SQL片段
func (loader *XMLMapperLoader) registerSqlFragment(namespace string, fragment *SqlFragment) error {
	// 这里实现SQL片段的注册逻辑
	// 将SQL片段存储到配置中，供其他语句引用
	return nil
}

// registerStatement 注册语句
func (loader *XMLMapperLoader) registerStatement(namespace string, id string, stmt interface{}, cmdType config.SqlCommandType) error {
	statementId := namespace + "." + id
	
	var sql string
	
	switch s := stmt.(type) {
	case *Select:
		sql = loader.processSQL(s.Content)
	case *Insert:
		sql = loader.processSQL(s.Content)
	case *Update:
		sql = loader.processSQL(s.Content)
	case *Delete:
		sql = loader.processSQL(s.Content)
	}
	
	// 简化实现 - 仅打印语句信息（演示用）
	fmt.Printf("注册语句: %s, SQL: %s\n", statementId, sql)
	
	return nil
}

// processSQL 处理SQL内容
func (loader *XMLMapperLoader) processSQL(content string) string {
	// 清理XML内容
	sql := strings.TrimSpace(content)
	
	// 处理include标签
	sql = loader.processIncludes(sql)
	
	// 处理动态SQL标签
	sql = loader.processDynamicSQL(sql)
	
	return sql
}

// processIncludes 处理include标签
func (loader *XMLMapperLoader) processIncludes(sql string) string {
	// 这里实现include标签的处理逻辑
	// 例如: <include refid="Base_Column_List" />
	// 需要替换为实际的SQL片段内容
	return sql
}

// processDynamicSQL 处理动态SQL标签
func (loader *XMLMapperLoader) processDynamicSQL(sql string) string {
	// 这里实现动态SQL标签的处理逻辑
	// 例如: <if>, <where>, <foreach>, <choose> 等标签
	// 将XML标签转换为框架内部的动态SQL表示
	return sql
}

// XMLMapperConfig XML映射器配置
type XMLMapperConfig struct {
	ConfigFile   string            // 主配置文件路径
	MapperDir    string            // 映射器文件目录
	Properties   map[string]string // 属性配置
	Environment  string            // 环境标识
}

// XMLConfigLoader XML配置加载器
type XMLConfigLoader struct {
	configPath string
}

// NewXMLConfigLoader 创建XML配置加载器
func NewXMLConfigLoader(configPath string) *XMLConfigLoader {
	return &XMLConfigLoader{
		configPath: configPath,
	}
}

// LoadConfiguration 加载MyBatis配置
func (loader *XMLConfigLoader) LoadConfiguration() (*config.Configuration, error) {
	// 创建基础配置
	cfg := config.NewConfiguration()
	
	// 从XML文件加载配置
	if err := loader.loadFromXML(cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration from XML: %w", err)
	}
	
	return cfg, nil
}

// loadFromXML 从XML文件加载配置
func (loader *XMLConfigLoader) loadFromXML(cfg *config.Configuration) error {
	data, err := ioutil.ReadFile(loader.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// 解析XML配置
	// 这里实现完整的MyBatis配置解析逻辑
	// 包括environments, typeAliases, mappers等
	
	fmt.Printf("Loaded MyBatis configuration from: %s\n", loader.configPath)
	fmt.Printf("Configuration size: %d bytes\n", len(data))
	
	return nil
}

// LoadPropertiesFile 加载属性文件
func LoadPropertiesFile(filename string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read properties file: %w", err)
	}
	
	properties := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			properties[key] = value
		}
	}
	
	return properties, nil
}

// ConfigurationBuilder 配置构建器
type ConfigurationBuilder struct {
	xmlLoader    *XMLConfigLoader
	mapperLoader *XMLMapperLoader
	properties   map[string]string
}

// NewConfigurationBuilder 创建配置构建器
func NewConfigurationBuilder(configFile, mapperDir string) *ConfigurationBuilder {
	return &ConfigurationBuilder{
		xmlLoader:    NewXMLConfigLoader(configFile),
		mapperLoader: NewXMLMapperLoader(mapperDir, nil),
		properties:   make(map[string]string),
	}
}

// LoadProperties 加载属性文件
func (builder *ConfigurationBuilder) LoadProperties(filename string) error {
	props, err := LoadPropertiesFile(filename)
	if err != nil {
		return err
	}
	
	// 合并属性
	for k, v := range props {
		builder.properties[k] = v
	}
	
	return nil
}

// Build 构建配置
func (builder *ConfigurationBuilder) Build() (*config.Configuration, error) {
	// 加载基础配置
	cfg, err := builder.xmlLoader.LoadConfiguration()
	if err != nil {
		return nil, err
	}
	
	// 设置映射器加载器的配置
	builder.mapperLoader.config = cfg
	
	// 加载所有映射器
	mappers, err := builder.mapperLoader.LoadAllMappers()
	if err != nil {
		return nil, err
	}
	
	// 注册映射器
	for _, mapper := range mappers {
		if err := builder.mapperLoader.RegisterMapper(mapper); err != nil {
			return nil, fmt.Errorf("failed to register mapper %s: %w", mapper.Namespace, err)
		}
	}
	
	// 应用属性配置
	builder.applyProperties(cfg)
	
	return cfg, nil
}

// applyProperties 应用属性配置
func (builder *ConfigurationBuilder) applyProperties(cfg *config.Configuration) {
	// 根据属性配置设置框架参数
	if val, ok := builder.properties["mybatis.config.mapUnderscoreToCamelCase"]; ok && val == "true" {
		cfg.MapUnderscoreToCamelCase = true
	}
	
	if val, ok := builder.properties["mybatis.config.lazyLoadingEnabled"]; ok && val == "true" {
		cfg.LazyLoadingEnabled = true
	}
	
	if val, ok := builder.properties["mybatis.config.cacheEnabled"]; ok && val == "true" {
		cfg.CacheEnabled = true
	}
}