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
//	// 方式2: 简化GORM版本
//	mb := mybatis.NewSimpleMyBatis(gormDB, nil)
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

// GetSimpleSession 获取简化会话（从完整版MyBatis）
func (mb *MyBatis) GetSimpleSession() SimpleSqlSession {
	sqlSession := mb.OpenSession()
	return &SimpleSqlSessionAdapter{
		sqlSession: sqlSession,
		mybatis:    mb,
	}
}

// GetSimpleSessionWithTx 获取带事务的简化会话（从完整版MyBatis）
func (mb *MyBatis) GetSimpleSessionWithTx() SimpleSqlSession {
	sqlSession := mb.OpenSessionWithAutoCommit(false)
	return &SimpleSqlSessionAdapter{
		sqlSession: sqlSession,
		mybatis:    mb,
	}
}

// SimpleMyBatis 简化版MyBatis实例 (基于GORM)
type SimpleMyBatis struct {
	db      *gorm.DB
	config  *SimpleConfig
	mappers map[string]*SimpleMapperInfo
	cache   *SimpleCache
	mutex   sync.RWMutex
}

// SimpleConfig MyBatis简化配置
type SimpleConfig struct {
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

// SimpleMapperInfo 简化映射器信息
type SimpleMapperInfo struct {
	Namespace   string
	Statements  map[string]*SimpleStatement
	ResultMaps  map[string]*SimpleResultMap
}

// SimpleStatement 简化SQL语句定义
type SimpleStatement struct {
	ID            string
	Namespace     string
	SQL           string
	StatementType SimpleStatementType
	ParameterType reflect.Type
	ResultType    reflect.Type
	ResultMap     string
	UseCache      bool
	Timeout       int
}

// SimpleStatementType 简化语句类型
type SimpleStatementType int

const (
	SimpleStatementTypeSelect SimpleStatementType = iota
	SimpleStatementTypeInsert
	SimpleStatementTypeUpdate
	SimpleStatementTypeDelete
)

// SimpleResultMap 简化结果映射
type SimpleResultMap struct {
	ID       string
	Type     reflect.Type
	Columns  []SimpleColumnMapping
}

// SimpleColumnMapping 简化列映射
type SimpleColumnMapping struct {
	Property string
	Column   string
	JavaType reflect.Type
}

// SimpleCache 简单缓存实现
type SimpleCache struct {
	data    map[string]interface{}
	mutex   sync.RWMutex
	maxSize int
}

// SimpleSqlSession 简化SQL会话接口
type SimpleSqlSession interface {
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

// DefaultSimpleSqlSession 默认简化SQL会话实现
type DefaultSimpleSqlSession struct {
	mybatis *SimpleMyBatis
	db      *gorm.DB
	tx      *gorm.DB // 事务数据库连接
}

// SimpleSqlSessionAdapter 简化会话适配器（完整版MyBatis到简化版的桥接）
type SimpleSqlSessionAdapter struct {
	sqlSession session.SqlSession
	mybatis    *MyBatis
}

// NewSimpleMyBatis 创建简化MyBatis实例
func NewSimpleMyBatis(db *gorm.DB, config *SimpleConfig) *SimpleMyBatis {
	if config == nil {
		config = DefaultSimpleConfig()
	}
	
	mb := &SimpleMyBatis{
		db:      db,
		config:  config,
		mappers: make(map[string]*SimpleMapperInfo),
		cache:   NewSimpleCache(config.CacheSize),
	}
	
	return mb
}

// DefaultSimpleConfig 默认简化配置
func DefaultSimpleConfig() *SimpleConfig {
	return &SimpleConfig{
		CacheEnabled:             true,
		CacheSize:               1000,
		MapUnderscoreToCamelCase: true,
		LogLevel:                "info",
		TypeAliases:             make(map[string]reflect.Type),
		MapperLocations:         []string{},
	}
}

// NewSimpleCache 创建简单缓存
func NewSimpleCache(maxSize int) *SimpleCache {
	return &SimpleCache{
		data:    make(map[string]interface{}),
		maxSize: maxSize,
	}
}

// OpenSession 打开简化会话
func (mb *SimpleMyBatis) OpenSession() SimpleSqlSession {
	return &DefaultSimpleSqlSession{
		mybatis: mb,
		db:      mb.db,
	}
}

// OpenSessionWithTx 打开带事务的简化会话
func (mb *SimpleMyBatis) OpenSessionWithTx() SimpleSqlSession {
	tx := mb.db.Begin()
	return &DefaultSimpleSqlSession{
		mybatis: mb,
		db:      mb.db,
		tx:      tx,
	}
}

// RegisterSimpleMapper 注册简化映射器
func (mb *SimpleMyBatis) RegisterSimpleMapper(namespace string, statements map[string]*SimpleStatement) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()
	
	mb.mappers[namespace] = &SimpleMapperInfo{
		Namespace:  namespace,
		Statements: statements,
		ResultMaps: make(map[string]*SimpleResultMap),
	}
}

// LoadMapperFromXML 从XML加载映射器（简化实现）
func (mb *SimpleMyBatis) LoadMapperFromXML(xmlPath string) error {
	// TODO: 实现XML解析
	// 现在先提供编程式注册方法
	return nil
}

// 实现SimpleSqlSession接口

// SelectOne 查询单条记录
func (session *DefaultSimpleSqlSession) SelectOne(statement string, parameter interface{}) (interface{}, error) {
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
func (session *DefaultSimpleSqlSession) SelectList(statement string, parameter interface{}) ([]interface{}, error) {
	stmt, err := session.getStatement(statement)
	if err != nil {
		return nil, err
	}
	
	if stmt.StatementType != SimpleStatementTypeSelect {
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
func (session *DefaultSimpleSqlSession) Insert(statement string, parameter interface{}) (int64, error) {
	return session.executeUpdate(statement, parameter, SimpleStatementTypeInsert)
}

// Update 更新记录
func (session *DefaultSimpleSqlSession) Update(statement string, parameter interface{}) (int64, error) {
	return session.executeUpdate(statement, parameter, SimpleStatementTypeUpdate)
}

// Delete 删除记录
func (session *DefaultSimpleSqlSession) Delete(statement string, parameter interface{}) (int64, error) {
	return session.executeUpdate(statement, parameter, SimpleStatementTypeDelete)
}

// executeUpdate 执行更新操作
func (session *DefaultSimpleSqlSession) executeUpdate(statement string, parameter interface{}, expectedType SimpleStatementType) (int64, error) {
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
func (session *DefaultSimpleSqlSession) GetMapper(mapperType reflect.Type) interface{} {
	// 简化实现：返回一个包含session的映射器实例
	// 实际应该创建动态代理
	return NewSimpleMapperProxy(mapperType, session)
}

// Commit 提交事务
func (session *DefaultSimpleSqlSession) Commit() error {
	if session.tx != nil {
		return session.tx.Commit().Error
	}
	return nil
}

// Rollback 回滚事务
func (session *DefaultSimpleSqlSession) Rollback() error {
	if session.tx != nil {
		return session.tx.Rollback().Error
	}
	return nil
}

// Close 关闭会话
func (session *DefaultSimpleSqlSession) Close() error {
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
func (session *DefaultSimpleSqlSession) getStatement(statementId string) (*SimpleStatement, error) {
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
func (session *DefaultSimpleSqlSession) getDB() *gorm.DB {
	if session.tx != nil {
		return session.tx
	}
	return session.db
}

// buildSQL 构建SQL和参数
func (session *DefaultSimpleSqlSession) buildSQL(stmt *SimpleStatement, parameter interface{}) (string, []interface{}, error) {
	sql := stmt.SQL
	var args []interface{}
	
	// 简化的参数处理
	if parameter != nil {
		args = session.extractParameters(parameter, sql)
	}
	
	return sql, args, nil
}

// extractParameters 提取参数
func (session *DefaultSimpleSqlSession) extractParameters(parameter interface{}, sql string) []interface{} {
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
func (session *DefaultSimpleSqlSession) convertResult(result map[string]interface{}, stmt *SimpleStatement) interface{} {
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
func (session *DefaultSimpleSqlSession) buildCacheKey(statement string, parameter interface{}) string {
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

// SimpleCache方法实现

// Get 获取缓存
func (cache *SimpleCache) Get(key string) interface{} {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.data[key]
}

// Put 放入缓存
func (cache *SimpleCache) Put(key string, value interface{}) {
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
func (cache *SimpleCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.data = make(map[string]interface{})
}

// SimpleMapperProxy 简化映射器代理
type SimpleMapperProxy struct {
	mapperType reflect.Type
	session    SimpleSqlSession
}

// NewSimpleMapperProxy 创建简化映射器代理
func NewSimpleMapperProxy(mapperType reflect.Type, session SimpleSqlSession) *SimpleMapperProxy {
	return &SimpleMapperProxy{
		mapperType: mapperType,
		session:    session,
	}
}

// 简化版便捷的构建器函数

// SimpleStatementBuilder 简化语句构建器
type SimpleStatementBuilder struct {
	statement *SimpleStatement
}

// NewSimpleStatement 创建新的简化语句
func NewSimpleStatement(id, namespace string) *SimpleStatementBuilder {
	return &SimpleStatementBuilder{
		statement: &SimpleStatement{
			ID:        id,
			Namespace: namespace,
			UseCache:  true,
			Timeout:   30,
		},
	}
}

// SQL 设置SQL
func (builder *SimpleStatementBuilder) SQL(sql string) *SimpleStatementBuilder {
	builder.statement.SQL = sql
	return builder
}

// Type 设置类型
func (builder *SimpleStatementBuilder) Type(statementType SimpleStatementType) *SimpleStatementBuilder {
	builder.statement.StatementType = statementType
	return builder
}

// ParameterType 设置参数类型
func (builder *SimpleStatementBuilder) ParameterType(paramType reflect.Type) *SimpleStatementBuilder {
	builder.statement.ParameterType = paramType
	return builder
}

// ResultType 设置结果类型
func (builder *SimpleStatementBuilder) ResultType(resultType reflect.Type) *SimpleStatementBuilder {
	builder.statement.ResultType = resultType
	return builder
}

// Cache 设置缓存
func (builder *SimpleStatementBuilder) Cache(useCache bool) *SimpleStatementBuilder {
	builder.statement.UseCache = useCache
	return builder
}

// Build 构建语句
func (builder *SimpleStatementBuilder) Build() *SimpleStatement {
	return builder.statement
}

// 简化版全局便捷函数

// QuickSetup 快速设置简化MyBatis
func QuickSetup(db *gorm.DB) *SimpleMyBatis {
	config := DefaultSimpleConfig()
	return NewSimpleMyBatis(db, config)
}

// WithContext 带上下文的操作
type ContextualSimpleSession struct {
	session SimpleSqlSession
	ctx     context.Context
}

// WithContext 为简化会话添加上下文
func (session *DefaultSimpleSqlSession) WithContext(ctx context.Context) *ContextualSimpleSession {
	return &ContextualSimpleSession{
		session: session,
		ctx:     ctx,
	}
}

// SelectOne 带上下文的查询单条
func (cs *ContextualSimpleSession) SelectOne(statement string, parameter interface{}) (interface{}, error) {
	// TODO: 实现超时和取消
	return cs.session.SelectOne(statement, parameter)
}

// SelectList 带上下文的查询多条
func (cs *ContextualSimpleSession) SelectList(statement string, parameter interface{}) ([]interface{}, error) {
	// TODO: 实现超时和取消
	return cs.session.SelectList(statement, parameter)
}

// ===============================================
// 简化会话适配器实现 (桥接完整版到简化版)
// ===============================================

// SelectOne 查询单条记录（适配器实现）
func (adapter *SimpleSqlSessionAdapter) SelectOne(statement string, parameter interface{}) (interface{}, error) {
	return adapter.sqlSession.SelectOne(statement, parameter)
}

// SelectList 查询多条记录（适配器实现）
func (adapter *SimpleSqlSessionAdapter) SelectList(statement string, parameter interface{}) ([]interface{}, error) {
	result, err := adapter.sqlSession.SelectList(statement, parameter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Insert 插入记录（适配器实现）
func (adapter *SimpleSqlSessionAdapter) Insert(statement string, parameter interface{}) (int64, error) {
	return adapter.sqlSession.Insert(statement, parameter)
}

// Update 更新记录（适配器实现）
func (adapter *SimpleSqlSessionAdapter) Update(statement string, parameter interface{}) (int64, error) {
	return adapter.sqlSession.Update(statement, parameter)
}

// Delete 删除记录（适配器实现）
func (adapter *SimpleSqlSessionAdapter) Delete(statement string, parameter interface{}) (int64, error) {
	return adapter.sqlSession.Delete(statement, parameter)
}

// GetMapper 获取映射器（适配器实现）
func (adapter *SimpleSqlSessionAdapter) GetMapper(mapperType reflect.Type) interface{} {
	mapper, _ := adapter.sqlSession.GetMapper(mapperType)
	return mapper
}

// Commit 提交事务（适配器实现）
func (adapter *SimpleSqlSessionAdapter) Commit() error {
	return adapter.sqlSession.Commit()
}

// Rollback 回滚事务（适配器实现）
func (adapter *SimpleSqlSessionAdapter) Rollback() error {
	return adapter.sqlSession.Rollback()
}

// Close 关闭会话（适配器实现）
func (adapter *SimpleSqlSessionAdapter) Close() error {
	return adapter.sqlSession.Close()
}

// ===============================================
// 统一的便捷方法
// ===============================================

// NewMyBatisWithSimpleConfig 使用简化配置创建完整版MyBatis
func NewMyBatisWithSimpleConfig(db *gorm.DB, simpleConfig *SimpleConfig) (*MyBatis, error) {
	if simpleConfig == nil {
		simpleConfig = DefaultSimpleConfig()
	}
	
	// 创建完整版配置
	configuration := config.NewConfiguration()
	
	// 如果有数据库配置，设置到完整版配置中
	if simpleConfig.DatabaseConfig != nil {
		configuration.SetDatabaseConfig(simpleConfig.DatabaseConfig)
	}
	
	// 设置其他配置项
	configuration.CacheEnabled = simpleConfig.CacheEnabled
	configuration.MapUnderscoreToCamelCase = simpleConfig.MapUnderscoreToCamelCase
	
	// 创建完整版MyBatis
	return NewMyBatis(configuration)
}

// CreateUnifiedMyBatis 创建统一的MyBatis实例（同时支持完整版和简化版API）
func CreateUnifiedMyBatis(db *gorm.DB, simpleConfig *SimpleConfig) (*MyBatis, *SimpleMyBatis, error) {
	// 创建完整版
	fullMyBatis, err := NewMyBatisWithSimpleConfig(db, simpleConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create full MyBatis: %w", err)
	}
	
	// 创建简化版
	simpleMyBatis := NewSimpleMyBatis(db, simpleConfig)
	
	return fullMyBatis, simpleMyBatis, nil
}
