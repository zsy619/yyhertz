// Package mybatis 提供MyBatis风格的ORM框架
//
// MyBatis-Go 是一个受MyBatis启发的Golang ORM框架，集成了完整框架版本和简化GORM版本
// 特色功能：
// 1. SQL映射和注解支持
// 2. 动态SQL构建 (支持if、where、foreach等标签)
// 3. 结果映射和类型转换
// 4. 多级缓存机制 (一级缓存、二级缓存)
// 5. 事务管理
// 6. 插件系统
// 7. 映射器代理
// 8. 批处理支持
// 9. GORM集成支持
//
// 使用示例:
//
//	// 方式1: 完整框架版本
//	config := mybatis.NewConfiguration()
//	config.SetDatabaseConfig(dbConfig)
//	factory, _ := mybatis.NewSqlSessionFactory(config)
//	session := factory.OpenSession()
//	defer session.Close()
//
//	// 方式2: GORM集成版本
//	mb := mybatis.NewMyBatisGorm(gormDB, nil)
//	session := mb.OpenSession()
//	defer session.Close()
//
//	// 执行数据库操作
//	user, err := session.SelectOne("UserMapper.selectById", 1)
package mybatis

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	frameworkConfig "github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mybatis/cache"
	"github.com/zsy619/yyhertz/framework/mybatis/config"
	"github.com/zsy619/yyhertz/framework/mybatis/mapper"
	"github.com/zsy619/yyhertz/framework/mybatis/session"
)

// MyBatis MyBatis框架主类
type MyBatis struct {
	configuration     *config.Configuration
	sqlSessionFactory session.SqlSessionFactory
}

// Builder MyBatis构建器
type Builder struct {
	config *config.Configuration
	error  error
}

// NewMyBatis 创建MyBatis实例
func NewMyBatis(configuration *config.Configuration) (*MyBatis, error) {
	if configuration == nil {
		configuration = config.NewConfiguration()
	}

	sqlSessionFactory, err := session.NewDefaultSqlSessionFactory(configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create sql session factory: %w", err)
	}

	return &MyBatis{
		configuration:     configuration,
		sqlSessionFactory: sqlSessionFactory,
	}, nil
}

// NewBuilder 创建MyBatis构建器
func NewBuilder() *Builder {
	return &Builder{
		config: config.NewConfiguration(),
	}
}

// DatabaseConfig 设置数据库配置
func (b *Builder) DatabaseConfig(dbConfig *frameworkConfig.DatabaseConfig) *Builder {
	if b.error != nil {
		return b
	}

	b.config.SetDatabaseConfig(dbConfig)
	return b
}

// CacheEnabled 设置是否启用缓存
func (b *Builder) CacheEnabled(enabled bool) *Builder {
	if b.error != nil {
		return b
	}

	b.config.CacheEnabled = enabled
	return b
}

// LazyLoading 设置是否启用延迟加载
func (b *Builder) LazyLoading(enabled bool) *Builder {
	if b.error != nil {
		return b
	}

	b.config.LazyLoadingEnabled = enabled
	return b
}

// AutoMapping 设置自动映射行为
func (b *Builder) AutoMapping(behavior config.AutoMappingBehavior) *Builder {
	if b.error != nil {
		return b
	}

	b.config.AutoMappingBehavior = behavior
	return b
}

// MapUnderscoreToCamelCase 设置下划线转驼峰
func (b *Builder) MapUnderscoreToCamelCase(enabled bool) *Builder {
	if b.error != nil {
		return b
	}

	b.config.MapUnderscoreToCamelCase = enabled
	return b
}

// RegisterMapper 注册映射器
func (b *Builder) RegisterMapper(mapperType reflect.Type) *Builder {
	if b.error != nil {
		return b
	}

	err := b.config.GetMapperRegistry().RegisterMapper(mapperType)
	if err != nil {
		b.error = fmt.Errorf("failed to register mapper %s: %w", mapperType.Name(), err)
	}
	return b
}

// RegisterTypeAlias 注册类型别名
func (b *Builder) RegisterTypeAlias(alias string, valueType reflect.Type) *Builder {
	if b.error != nil {
		return b
	}

	b.config.GetTypeAliasRegistry().RegisterAlias(alias, valueType)
	return b
}

// RegisterTypeHandler 注册类型处理器
func (b *Builder) RegisterTypeHandler(javaType reflect.Type, handler config.TypeHandler) *Builder {
	if b.error != nil {
		return b
	}

	b.config.GetTypeHandlerRegistry().RegisterTypeHandler(javaType, handler)
	return b
}

// Build 构建MyBatis实例
func (b *Builder) Build() (*MyBatis, error) {
	if b.error != nil {
		return nil, b.error
	}

	return NewMyBatis(b.config)
}

// GetConfiguration 获取配置
func (mb *MyBatis) GetConfiguration() *config.Configuration {
	return mb.configuration
}

// GetSqlSessionFactory 获取SQL会话工厂
func (mb *MyBatis) GetSqlSessionFactory() session.SqlSessionFactory {
	return mb.sqlSessionFactory
}

// OpenSession 打开新会话
func (mb *MyBatis) OpenSession() session.SqlSession {
	return mb.sqlSessionFactory.OpenSession()
}

// OpenSessionWithAutoCommit 打开带自动提交的会话
func (mb *MyBatis) OpenSessionWithAutoCommit(autoCommit bool) session.SqlSession {
	return mb.sqlSessionFactory.OpenSessionWithAutoCommit(autoCommit)
}

// OpenSessionWithExecutorType 打开带执行器类型的会话
func (mb *MyBatis) OpenSessionWithExecutorType(execType config.ExecutorType) session.SqlSession {
	return mb.sqlSessionFactory.OpenSessionWithExecutorType(execType)
}

// Execute 执行数据库操作 (使用会话模板)
func (mb *MyBatis) Execute(callback func(session session.SqlSession) (any, error)) (any, error) {
	sqlSession := mb.OpenSession()
	defer sqlSession.Close()

	return callback(sqlSession)
}

// ExecuteWithTransaction 在事务中执行数据库操作
func (mb *MyBatis) ExecuteWithTransaction(callback func(session session.SqlSession) error) error {
	sqlSession := mb.OpenSession()
	defer sqlSession.Close()

	err := callback(sqlSession)
	if err != nil {
		sqlSession.Rollback()
		return err
	}

	return sqlSession.Commit()
}

// 便捷方法

// SelectOne 查询单条记录
func (mb *MyBatis) SelectOne(statement string, parameter any) (any, error) {
	return mb.Execute(func(session session.SqlSession) (any, error) {
		return session.SelectOne(statement, parameter)
	})
}

// SelectList 查询多条记录
func (mb *MyBatis) SelectList(statement string, parameter any) ([]any, error) {
	result, err := mb.Execute(func(session session.SqlSession) (any, error) {
		return session.SelectList(statement, parameter)
	})
	if err != nil {
		return nil, err
	}

	if list, ok := result.([]any); ok {
		return list, nil
	}

	return nil, fmt.Errorf("result is not a list")
}

// Insert 插入记录
func (mb *MyBatis) Insert(statement string, parameter any) (int64, error) {
	result, err := mb.Execute(func(session session.SqlSession) (any, error) {
		return session.Insert(statement, parameter)
	})
	if err != nil {
		return 0, err
	}

	if count, ok := result.(int64); ok {
		return count, nil
	}

	return 0, fmt.Errorf("result is not int64")
}

// Update 更新记录
func (mb *MyBatis) Update(statement string, parameter any) (int64, error) {
	result, err := mb.Execute(func(session session.SqlSession) (any, error) {
		return session.Update(statement, parameter)
	})
	if err != nil {
		return 0, err
	}

	if count, ok := result.(int64); ok {
		return count, nil
	}

	return 0, fmt.Errorf("result is not int64")
}

// Delete 删除记录
func (mb *MyBatis) Delete(statement string, parameter any) (int64, error) {
	result, err := mb.Execute(func(session session.SqlSession) (any, error) {
		return session.Delete(statement, parameter)
	})
	if err != nil {
		return 0, err
	}

	if count, ok := result.(int64); ok {
		return count, nil
	}

	return 0, fmt.Errorf("result is not int64")
}

// GetMapper 获取映射器
func (mb *MyBatis) GetMapper(mapperType reflect.Type) (any, error) {
	return mb.Execute(func(session session.SqlSession) (any, error) {
		return session.GetMapper(mapperType)
	})
}

// WithSession 使用指定会话执行操作
func (mb *MyBatis) WithSession(sqlSession session.SqlSession, callback func() error) error {
	return callback()
}

// 全局函数 (兼容性)

var defaultMyBatis *MyBatis

// SetDefaultMyBatis 设置默认MyBatis实例
func SetDefaultMyBatis(mb *MyBatis) {
	defaultMyBatis = mb
}

// GetDefaultMyBatis 获取默认MyBatis实例
func GetDefaultMyBatis() *MyBatis {
	return defaultMyBatis
}

// DefaultSelectOne 使用默认实例查询单条记录
func DefaultSelectOne(statement string, parameter any) (any, error) {
	if defaultMyBatis == nil {
		return nil, fmt.Errorf("default MyBatis instance not set")
	}
	return defaultMyBatis.SelectOne(statement, parameter)
}

// DefaultSelectList 使用默认实例查询多条记录
func DefaultSelectList(statement string, parameter any) ([]any, error) {
	if defaultMyBatis == nil {
		return nil, fmt.Errorf("default MyBatis instance not set")
	}
	return defaultMyBatis.SelectList(statement, parameter)
}

// DefaultInsert 使用默认实例插入记录
func DefaultInsert(statement string, parameter any) (int64, error) {
	if defaultMyBatis == nil {
		return 0, fmt.Errorf("default MyBatis instance not set")
	}
	return defaultMyBatis.Insert(statement, parameter)
}

// DefaultUpdate 使用默认实例更新记录
func DefaultUpdate(statement string, parameter any) (int64, error) {
	if defaultMyBatis == nil {
		return 0, fmt.Errorf("default MyBatis instance not set")
	}
	return defaultMyBatis.Update(statement, parameter)
}

// DefaultDelete 使用默认实例删除记录
func DefaultDelete(statement string, parameter any) (int64, error) {
	if defaultMyBatis == nil {
		return 0, fmt.Errorf("default MyBatis instance not set")
	}
	return defaultMyBatis.Delete(statement, parameter)
}

// 工厂方法

// NewConfiguration 创建新配置
func NewConfiguration() *config.Configuration {
	return config.NewConfiguration()
}

// NewSqlSessionFactory 创建SQL会话工厂
func NewSqlSessionFactory(configuration *config.Configuration) (session.SqlSessionFactory, error) {
	return session.NewDefaultSqlSessionFactory(configuration)
}

// NewLRUCache 创建LRU缓存
func NewLRUCache(capacity int) cache.Cache {
	return cache.NewLruCache(cache.NewPerpetualCache("default"), capacity)
}

// NewDynamicSqlBuilder 创建动态SQL构建器
func NewDynamicSqlBuilder() *mapper.DynamicSqlBuilder {
	return mapper.NewDynamicSqlBuilder()
}

// 版本信息
const (
	Version = "1.0.0"
	Name    = "MyBatis-Go"
)

// GetVersion 获取版本信息
func GetVersion() string {
	return Version
}

// GetName 获取框架名称
func GetName() string {
	return Name
}

// GetInfo 获取框架信息
func GetInfo() map[string]string {
	return map[string]string{
		"name":        Name,
		"version":     Version,
		"description": "MyBatis风格的Golang ORM框架",
		"features":    "SQL映射、动态SQL、缓存、事务、插件系统、GORM集成",
	}
}

// ===============================================
// 集成简化版MyBatis实现 (基于GORM)
// ===============================================

// 为完整版MyBatis添加简化API支持

// GetGormSession 获取GORM会话（从完整版MyBatis）
func (mb *MyBatis) GetGormSession() SqlSession {
	sqlSession := mb.OpenSession()
	return &SqlSessionAdapter{
		sqlSession: sqlSession,
		mybatis:    mb,
	}
}

// GetGormSessionWithTx 获取带事务的GORM会话（从完整版MyBatis）
func (mb *MyBatis) GetGormSessionWithTx() SqlSession {
	sqlSession := mb.OpenSessionWithAutoCommit(false)
	return &SqlSessionAdapter{
		sqlSession: sqlSession,
		mybatis:    mb,
	}
}

// MyBatisGorm GORM集成版MyBatis实例
type MyBatisGorm struct {
	db      *gorm.DB
	config  *GormConfig
	mappers map[string]*MapperInfo
	cache   *LegacyCache
	mutex   sync.RWMutex
}

// GormConfig MyBatis GORM集成配置
type GormConfig struct {
	// 数据库配置
	DatabaseConfig *frameworkConfig.DatabaseConfig
	
	// 映射器配置
	MapperLocations []string // XML映射文件路径
	TypeAliases     map[string]reflect.Type
	
	// 缓存配置
	CacheEnabled    bool
	CacheSize       int
	
	// 其他配置
	MapUnderscoreToCamelCase bool
	LogLevel                 string
}

// MapperInfo 映射器信息
type MapperInfo struct {
	Namespace   string
	Statements  map[string]*Statement
	ResultMaps  map[string]*ResultMap
}

// Statement SQL语句定义
type Statement struct {
	ID            string
	Namespace     string
	SQL           string
	StatementType StatementType
	ParameterType reflect.Type
	ResultType    reflect.Type
	ResultMap     string
	UseCache      bool
	Timeout       int
}

// StatementType 语句类型
type StatementType int

const (
	StatementTypeSelect StatementType = iota
	StatementTypeInsert
	StatementTypeUpdate
	StatementTypeDelete
)

// ResultMap 结果映射
type ResultMap struct {
	ID       string
	Type     reflect.Type
	Columns  []ColumnMapping
}

// ColumnMapping 列映射
type ColumnMapping struct {
	Property string
	Column   string
	JavaType reflect.Type
}

// LegacyCache 缓存实现（保持向后兼容）
type LegacyCache struct {
	data    map[string]interface{}
	mutex   sync.RWMutex
	maxSize int
}

// SqlSession SQL会话接口
type SqlSession interface {
	SelectOne(statement string, parameter interface{}) (interface{}, error)
	SelectList(statement string, parameter interface{}) ([]interface{}, error)
	Insert(statement string, parameter interface{}) (int64, error)
	Update(statement string, parameter interface{}) (int64, error)
	Delete(statement string, parameter interface{}) (int64, error)
	GetMapper(mapperType reflect.Type) interface{}
	Commit() error
	Rollback() error
	Close() error
}

// DefaultSqlSession 默认SQL会话实现
type DefaultSqlSession struct {
	mybatis *MyBatisGorm
	db      *gorm.DB
	tx      *gorm.DB // 事务数据库连接
}

// SqlSessionAdapter 会话适配器（完整版MyBatis到GORM版的桥接）
type SqlSessionAdapter struct {
	sqlSession session.SqlSession
	mybatis    *MyBatis
}

// NewMyBatisGorm 创建GORM集成版MyBatis实例
func NewMyBatisGorm(db *gorm.DB, config *GormConfig) *MyBatisGorm {
	if config == nil {
		config = DefaultGormConfig()
	}
	
	mb := &MyBatisGorm{
		db:      db,
		config:  config,
		mappers: make(map[string]*MapperInfo),
		cache:   NewLegacyCache(config.CacheSize),
	}
	
	return mb
}

// DefaultGormConfig 默认GORM集成配置
func DefaultGormConfig() *GormConfig {
	return &GormConfig{
		CacheEnabled:             true,
		CacheSize:               1000,
		MapUnderscoreToCamelCase: true,
		LogLevel:                "info",
		TypeAliases:             make(map[string]reflect.Type),
		MapperLocations:         []string{},
	}
}

// NewLegacyCache 创建缓存
func NewLegacyCache(maxSize int) *LegacyCache {
	return &LegacyCache{
		data:    make(map[string]interface{}),
		maxSize: maxSize,
	}
}

// OpenSession 打开会话
func (mb *MyBatisGorm) OpenSession() SqlSession {
	return &DefaultSqlSession{
		mybatis: mb,
		db:      mb.db,
	}
}

// OpenSessionWithTx 打开带事务的会话
func (mb *MyBatisGorm) OpenSessionWithTx() SqlSession {
	tx := mb.db.Begin()
	return &DefaultSqlSession{
		mybatis: mb,
		db:      mb.db,
		tx:      tx,
	}
}

// RegisterMapper 注册映射器
func (mb *MyBatisGorm) RegisterMapper(namespace string, statements map[string]*Statement) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()
	
	mb.mappers[namespace] = &MapperInfo{
		Namespace:  namespace,
		Statements: statements,
		ResultMaps: make(map[string]*ResultMap),
	}
}

// LoadMapperFromXML 从XML加载映射器
func (mb *MyBatisGorm) LoadMapperFromXML(xmlPath string) error {
	// TODO: 实现XML解析
	// 现在先提供编程式注册方法
	return nil
}

// 实现SqlSession接口

// SelectOne 查询单条记录
func (session *DefaultSqlSession) SelectOne(statement string, parameter interface{}) (interface{}, error) {
	results, err := session.SelectList(statement, parameter)
	if err != nil {
		return nil, err
	}
	
	if len(results) == 0 {
		return nil, nil
	}
	
	if len(results) > 1 {
		return nil, fmt.Errorf("expected one result but got %d", len(results))
	}
	
	return results[0], nil
}

// SelectList 查询多条记录
func (session *DefaultSqlSession) SelectList(statement string, parameter interface{}) ([]interface{}, error) {
	stmt, err := session.getStatement(statement)
	if err != nil {
		return nil, err
	}
	
	if stmt.StatementType != StatementTypeSelect {
		return nil, fmt.Errorf("statement %s is not a select statement", statement)
	}
	
	// 检查缓存
	if stmt.UseCache && session.mybatis.config.CacheEnabled {
		cacheKey := session.buildCacheKey(statement, parameter)
		if cached := session.mybatis.cache.Get(cacheKey); cached != nil {
			return cached.([]interface{}), nil
		}
	}
	
	// 构建SQL和参数
	sql, args, err := session.buildSQL(stmt, parameter)
	if err != nil {
		return nil, err
	}
	
	// 执行查询
	db := session.getDB()
	var results []map[string]interface{}
	err = db.Raw(sql, args...).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	
	// 转换结果
	convertedResults := make([]interface{}, len(results))
	for i, result := range results {
		converted := session.convertResult(result, stmt)
		convertedResults[i] = converted
	}
	
	// 缓存结果
	if stmt.UseCache && session.mybatis.config.CacheEnabled {
		cacheKey := session.buildCacheKey(statement, parameter)
		session.mybatis.cache.Put(cacheKey, convertedResults)
	}
	
	return convertedResults, nil
}

// Insert 插入记录
func (session *DefaultSqlSession) Insert(statement string, parameter interface{}) (int64, error) {
	return session.executeUpdate(statement, parameter, StatementTypeInsert)
}

// Update 更新记录
func (session *DefaultSqlSession) Update(statement string, parameter interface{}) (int64, error) {
	return session.executeUpdate(statement, parameter, StatementTypeUpdate)
}

// Delete 删除记录
func (session *DefaultSqlSession) Delete(statement string, parameter interface{}) (int64, error) {
	return session.executeUpdate(statement, parameter, StatementTypeDelete)
}

// executeUpdate 执行更新操作
func (session *DefaultSqlSession) executeUpdate(statement string, parameter interface{}, expectedType StatementType) (int64, error) {
	stmt, err := session.getStatement(statement)
	if err != nil {
		return 0, err
	}
	
	if stmt.StatementType != expectedType {
		return 0, fmt.Errorf("statement %s type mismatch", statement)
	}
	
	// 构建SQL和参数
	sql, args, err := session.buildSQL(stmt, parameter)
	if err != nil {
		return 0, err
	}
	
	// 执行更新
	db := session.getDB()
	result := db.Exec(sql, args...)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to execute update: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}

// GetMapper 获取映射器代理
func (session *DefaultSqlSession) GetMapper(mapperType reflect.Type) interface{} {
	// 简化实现：返回一个包含session的映射器实例
	// 实际应该创建动态代理
	return NewMapperProxy(mapperType, session)
}

// Commit 提交事务
func (session *DefaultSqlSession) Commit() error {
	if session.tx != nil {
		return session.tx.Commit().Error
	}
	return nil
}

// Rollback 回滚事务
func (session *DefaultSqlSession) Rollback() error {
	if session.tx != nil {
		return session.tx.Rollback().Error
	}
	return nil
}

// Close 关闭会话
func (session *DefaultSqlSession) Close() error {
	if session.tx != nil {
		// 如果事务还没有提交或回滚，则回滚
		if err := session.tx.Rollback().Error; err != nil && err != gorm.ErrInvalidTransaction {
			return err
		}
	}
	return nil
}

// 辅助方法

// getStatement 获取语句定义
func (session *DefaultSqlSession) getStatement(statementId string) (*Statement, error) {
	parts := strings.SplitN(statementId, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid statement id: %s", statementId)
	}
	
	namespace := parts[0]
	statementName := parts[1]
	
	session.mybatis.mutex.RLock()
	mapperInfo, exists := session.mybatis.mappers[namespace]
	session.mybatis.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("namespace not found: %s", namespace)
	}
	
	statement, exists := mapperInfo.Statements[statementName]
	if !exists {
		return nil, fmt.Errorf("statement not found: %s", statementId)
	}
	
	return statement, nil
}

// getDB 获取数据库连接
func (session *DefaultSqlSession) getDB() *gorm.DB {
	if session.tx != nil {
		return session.tx
	}
	return session.db
}

// buildSQL 构建SQL和参数
func (session *DefaultSqlSession) buildSQL(stmt *Statement, parameter interface{}) (string, []interface{}, error) {
	sql := stmt.SQL
	var args []interface{}
	
	// 简化的参数处理
	if parameter != nil {
		args = session.extractParameters(parameter, sql)
	}
	
	return sql, args, nil
}

// extractParameters 提取参数
func (session *DefaultSqlSession) extractParameters(parameter interface{}, sql string) []interface{} {
	// 计算SQL中的参数占位符数量
	paramCount := strings.Count(sql, "?")
	
	if paramCount == 0 {
		return []interface{}{}
	}
	
	paramValue := reflect.ValueOf(parameter)
	paramType := reflect.TypeOf(parameter)
	
	// 处理指针
	if paramType.Kind() == reflect.Ptr {
		paramValue = paramValue.Elem()
		paramType = paramType.Elem()
	}
	
	var args []interface{}
	
	switch paramType.Kind() {
	case reflect.Struct:
		// 结构体：提取字段值
		for i := 0; i < paramType.NumField() && len(args) < paramCount; i++ {
			field := paramType.Field(i)
			if field.IsExported() {
				fieldValue := paramValue.Field(i)
				args = append(args, fieldValue.Interface())
			}
		}
	case reflect.Map:
		// Map：提取值（按key排序）
		for _, key := range paramValue.MapKeys() {
			if len(args) >= paramCount {
				break
			}
			args = append(args, paramValue.MapIndex(key).Interface())
		}
	case reflect.Slice:
		// 切片：提取元素
		for i := 0; i < paramValue.Len() && len(args) < paramCount; i++ {
			args = append(args, paramValue.Index(i).Interface())
		}
	default:
		// 基本类型：重复使用
		for i := 0; i < paramCount; i++ {
			args = append(args, parameter)
		}
	}
	
	// 如果参数不够，用nil填充
	for len(args) < paramCount {
		args = append(args, nil)
	}
	
	return args
}

// convertResult 转换查询结果
func (session *DefaultSqlSession) convertResult(result map[string]interface{}, stmt *Statement) interface{} {
	if !session.mybatis.config.MapUnderscoreToCamelCase {
		return result
	}
	
	// 下划线转驼峰
	converted := make(map[string]interface{})
	for key, value := range result {
		camelKey := underscoreToCamelCase(key)
		converted[camelKey] = value
	}
	
	return converted
}

// buildCacheKey 构建缓存键
func (session *DefaultSqlSession) buildCacheKey(statement string, parameter interface{}) string {
	return fmt.Sprintf("%s:%v", statement, parameter)
}

// underscoreToCamelCase 下划线转驼峰
func underscoreToCamelCase(name string) string {
	parts := strings.Split(strings.ToLower(name), "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// LegacyCache方法实现

// Get 获取缓存
func (cache *LegacyCache) Get(key string) interface{} {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.data[key]
}

// Put 放入缓存
func (cache *LegacyCache) Put(key string, value interface{}) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	
	// 简单的LRU实现：如果超过最大大小，清除一半
	if len(cache.data) >= cache.maxSize {
		for k := range cache.data {
			delete(cache.data, k)
			if len(cache.data) <= cache.maxSize/2 {
				break
			}
		}
	}
	
	cache.data[key] = value
}

// Clear 清空缓存
func (cache *LegacyCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.data = make(map[string]interface{})
}

// MapperProxy 映射器代理
type MapperProxy struct {
	mapperType reflect.Type
	session    SqlSession
}

// NewMapperProxy 创建映射器代理
func NewMapperProxy(mapperType reflect.Type, session SqlSession) *MapperProxy {
	return &MapperProxy{
		mapperType: mapperType,
		session:    session,
	}
}

// 便捷的构建器函数

// StatementBuilder 语句构建器
type StatementBuilder struct {
	statement *Statement
}

// NewStatement 创建新的语句
func NewStatement(id, namespace string) *StatementBuilder {
	return &StatementBuilder{
		statement: &Statement{
			ID:        id,
			Namespace: namespace,
			UseCache:  true,
			Timeout:   30,
		},
	}
}

// SQL 设置SQL
func (builder *StatementBuilder) SQL(sql string) *StatementBuilder {
	builder.statement.SQL = sql
	return builder
}

// Type 设置类型
func (builder *StatementBuilder) Type(statementType StatementType) *StatementBuilder {
	builder.statement.StatementType = statementType
	return builder
}

// ParameterType 设置参数类型
func (builder *StatementBuilder) ParameterType(paramType reflect.Type) *StatementBuilder {
	builder.statement.ParameterType = paramType
	return builder
}

// ResultType 设置结果类型
func (builder *StatementBuilder) ResultType(resultType reflect.Type) *StatementBuilder {
	builder.statement.ResultType = resultType
	return builder
}

// Cache 设置缓存
func (builder *StatementBuilder) Cache(useCache bool) *StatementBuilder {
	builder.statement.UseCache = useCache
	return builder
}

// Build 构建语句
func (builder *StatementBuilder) Build() *Statement {
	return builder.statement
}

// 简化版全局便捷函数

// QuickSetup 快速设置MyBatis GORM集成版
func QuickSetup(db *gorm.DB) *MyBatisGorm {
	config := DefaultGormConfig()
	return NewMyBatisGorm(db, config)
}

// WithContext 带上下文的操作
type ContextualSession struct {
	session SqlSession
	ctx     context.Context
}

// WithContext 为会话添加上下文
func (session *DefaultSqlSession) WithContext(ctx context.Context) *ContextualSession {
	return &ContextualSession{
		session: session,
		ctx:     ctx,
	}
}

// SelectOne 带上下文的查询单条
func (cs *ContextualSession) SelectOne(statement string, parameter interface{}) (interface{}, error) {
	// TODO: 实现超时和取消
	return cs.session.SelectOne(statement, parameter)
}

// SelectList 带上下文的查询多条
func (cs *ContextualSession) SelectList(statement string, parameter interface{}) ([]interface{}, error) {
	// TODO: 实现超时和取消
	return cs.session.SelectList(statement, parameter)
}

// ===============================================
// 简化会话适配器实现 (桥接完整版到简化版)
// ===============================================

// SelectOne 查询单条记录（适配器实现）
func (adapter *SqlSessionAdapter) SelectOne(statement string, parameter interface{}) (interface{}, error) {
	return adapter.sqlSession.SelectOne(statement, parameter)
}

// SelectList 查询多条记录（适配器实现）
func (adapter *SqlSessionAdapter) SelectList(statement string, parameter interface{}) ([]interface{}, error) {
	result, err := adapter.sqlSession.SelectList(statement, parameter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Insert 插入记录（适配器实现）
func (adapter *SqlSessionAdapter) Insert(statement string, parameter interface{}) (int64, error) {
	return adapter.sqlSession.Insert(statement, parameter)
}

// Update 更新记录（适配器实现）
func (adapter *SqlSessionAdapter) Update(statement string, parameter interface{}) (int64, error) {
	return adapter.sqlSession.Update(statement, parameter)
}

// Delete 删除记录（适配器实现）
func (adapter *SqlSessionAdapter) Delete(statement string, parameter interface{}) (int64, error) {
	return adapter.sqlSession.Delete(statement, parameter)
}

// GetMapper 获取映射器（适配器实现）
func (adapter *SqlSessionAdapter) GetMapper(mapperType reflect.Type) interface{} {
	mapper, _ := adapter.sqlSession.GetMapper(mapperType)
	return mapper
}

// Commit 提交事务（适配器实现）
func (adapter *SqlSessionAdapter) Commit() error {
	return adapter.sqlSession.Commit()
}

// Rollback 回滚事务（适配器实现）
func (adapter *SqlSessionAdapter) Rollback() error {
	return adapter.sqlSession.Rollback()
}

// Close 关闭会话（适配器实现）
func (adapter *SqlSessionAdapter) Close() error {
	return adapter.sqlSession.Close()
}

// ===============================================
// 统一的便捷方法
// ===============================================

// NewMyBatisWithGormConfig 使用GORM配置创建完整版MyBatis
func NewMyBatisWithGormConfig(db *gorm.DB, gormConfig *GormConfig) (*MyBatis, error) {
	if gormConfig == nil {
		gormConfig = DefaultGormConfig()
	}
	
	// 创建完整版配置
	configuration := config.NewConfiguration()
	
	// 如果有数据库配置，设置到完整版配置中
	if gormConfig.DatabaseConfig != nil {
		configuration.SetDatabaseConfig(gormConfig.DatabaseConfig)
	}
	
	// 设置其他配置项
	configuration.CacheEnabled = gormConfig.CacheEnabled
	configuration.MapUnderscoreToCamelCase = gormConfig.MapUnderscoreToCamelCase
	
	// 创建完整版MyBatis
	return NewMyBatis(configuration)
}

// CreateUnifiedMyBatis 创建统一的MyBatis实例（同时支持完整版和GORM集成版API）
func CreateUnifiedMyBatis(db *gorm.DB, gormConfig *GormConfig) (*MyBatis, *MyBatisGorm, error) {
	// 创建完整版
	fullMyBatis, err := NewMyBatisWithGormConfig(db, gormConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create full MyBatis: %w", err)
	}
	
	// 创建GORM集成版
	gormMyBatis := NewMyBatisGorm(db, gormConfig)
	
	return fullMyBatis, gormMyBatis, nil
}

// ===============================================
// 新增：简化版API（推荐使用）
// ===============================================

// NewSimple 创建简化版MyBatis会话（推荐使用）
func NewSimple(db *gorm.DB) SimpleSession {
	return NewSimpleSession(db)
}

// NewSimpleWithHooks 创建带常用钩子的简化版会话
func NewSimpleWithHooks(db *gorm.DB, enableDebug bool) SimpleSession {
	session := NewSimpleSession(db).Debug(enableDebug)
	
	// 添加常用钩子
	session = session.AddBeforeHook(AuditHook())
	
	// 添加性能监控钩子（100ms慢查询阈值）
	beforeHook, afterHook := PerformanceHook(100 * time.Millisecond)
	session = session.AddBeforeHook(beforeHook).AddAfterHook(afterHook)
	
	if enableDebug {
		debugBefore, debugAfter := DebugHook()
		session = session.AddBeforeHook(debugBefore).AddAfterHook(debugAfter)
	}
	
	return session
}

// NewTransactionSession 创建支持事务的会话
func NewTransactionSession(db *gorm.DB) *TransactionAwareSession {
	return NewTransactionAwareSession(db)
}

// NewXMLMapper 创建支持XML Mapper的会话（推荐用于复杂查询）
func NewXMLMapper(db *gorm.DB) XMLSession {
	return NewXMLSession(db)
}

// NewXMLMapperWithHooks 创建带钩子的XML Mapper会话
func NewXMLMapperWithHooks(db *gorm.DB, enableDebug bool) XMLSession {
	return NewXMLSessionWithHooks(db, enableDebug)
}
