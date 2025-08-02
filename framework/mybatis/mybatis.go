// Package mybatis 提供MyBatis风格的ORM框架
//
// MyBatis-Go 是一个受MyBatis启发的Golang ORM框架
// 特色功能：
// 1. SQL映射和注解支持
// 2. 动态SQL构建 (支持if、where、foreach等标签)
// 3. 结果映射和类型转换
// 4. 多级缓存机制 (一级缓存、二级缓存)
// 5. 事务管理
// 6. 插件系统
// 7. 映射器代理
// 8. 批处理支持
//
// 使用示例:
//   // 1. 创建配置
//   config := mybatis.NewConfiguration()
//   config.SetDatabaseConfig(dbConfig)
//
//   // 2. 创建会话工厂
//   factory, _ := mybatis.NewSqlSessionFactory(config)
//
//   // 3. 获取会话和映射器
//   session := factory.OpenSession()
//   defer session.Close()
//   mapper := session.GetMapper(reflect.TypeOf((*UserMapper)(nil)).Elem())
//
//   // 4. 执行数据库操作
//   user, err := mapper.SelectById(1)
package mybatis

import (
	"fmt"
	"reflect"
	
	"github.com/zsy619/yyhertz/framework/mybatis/config"
	"github.com/zsy619/yyhertz/framework/mybatis/session"
	"github.com/zsy619/yyhertz/framework/mybatis/cache"
	"github.com/zsy619/yyhertz/framework/mybatis/mapper"
	"github.com/zsy619/yyhertz/framework/orm"
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
func (b *Builder) DatabaseConfig(dbConfig *orm.DatabaseConfig) *Builder {
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
		"name":    Name,
		"version": Version,
		"description": "MyBatis风格的Golang ORM框架",
		"features": "SQL映射、动态SQL、缓存、事务、插件系统",
	}
}