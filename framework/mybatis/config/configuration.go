// Package config 提供MyBatis配置管理
//
// 负责管理MyBatis框架的配置信息，包括数据库连接、映射器、类型处理器等
package config

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// Configuration MyBatis核心配置
type Configuration struct {
	// 数据库配置
	DatabaseConfig *config.DatabaseConfig

	// 映射器配置
	MapperRegistry *MapperRegistry

	// 类型系统
	TypeAliasRegistry   *TypeAliasRegistry
	TypeHandlerRegistry *TypeHandlerRegistry

	// 缓存配置
	CacheEnabled       bool
	LocalCacheScope    LocalCacheScope
	DefaultCacheConfig *CacheConfig

	// 执行器配置
	DefaultExecutorType   ExecutorType
	LazyLoadingEnabled    bool
	AggressiveLazyLoading bool

	// 其他配置
	MultipleResultSetsEnabled        bool
	UseColumnLabel                   bool
	UseGeneratedKeys                 bool
	AutoMappingBehavior              AutoMappingBehavior
	AutoMappingUnknownColumnBehavior AutoMappingUnknownColumnBehavior
	DefaultStatementTimeout          *time.Duration
	DefaultFetchSize                 *int
	MapUnderscoreToCamelCase         bool
	CallSettersOnNulls               bool
	UseActualParamName               bool

	// 内部状态
	mutex sync.RWMutex
}

// LocalCacheScope 本地缓存作用域
type LocalCacheScope int

const (
	LocalCacheScopeSession LocalCacheScope = iota
	LocalCacheScopeStatement
)

// ExecutorType 执行器类型
type ExecutorType int

const (
	ExecutorTypeDefault ExecutorType = iota
	ExecutorTypeReuse
	ExecutorTypeBatch
)

// AutoMappingBehavior 自动映射行为
type AutoMappingBehavior int

const (
	AutoMappingBehaviorNone AutoMappingBehavior = iota
	AutoMappingBehaviorPartial
	AutoMappingBehaviorFull
)

// AutoMappingUnknownColumnBehavior 未知列自动映射行为
type AutoMappingUnknownColumnBehavior int

const (
	AutoMappingUnknownColumnBehaviorNone AutoMappingUnknownColumnBehavior = iota
	AutoMappingUnknownColumnBehaviorWarning
	AutoMappingUnknownColumnBehaviorFailing
)

// CacheConfig 缓存配置
type CacheConfig struct {
	Implementation string         `json:"implementation" yaml:"implementation"` // 缓存实现类型
	Enabled        bool           `json:"enabled" yaml:"enabled"`               // 是否启用
	Size           int            `json:"size" yaml:"size"`                     // 缓存大小
	FlushInterval  time.Duration  `json:"flush_interval" yaml:"flush_interval"` // 刷新间隔
	ReadWrite      bool           `json:"read_write" yaml:"read_write"`         // 是否读写
	Properties     map[string]any `json:"properties" yaml:"properties"`         // 其他属性
}

// MapperRegistry 映射器注册表
type MapperRegistry struct {
	knownMappers map[reflect.Type]*MapperProxyFactory
	mutex        sync.RWMutex
}

// MapperProxyFactory 映射器代理工厂
type MapperProxyFactory struct {
	mapperInterface reflect.Type
	methodCache     map[string]*MapperMethod
	mutex           sync.RWMutex
}

// TypeAliasRegistry 类型别名注册表
type TypeAliasRegistry struct {
	aliases map[string]reflect.Type
	mutex   sync.RWMutex
}

// TypeHandlerRegistry 类型处理器注册表
type TypeHandlerRegistry struct {
	// jdbcType -> javaType -> TypeHandler
	typeHandlerMap       map[string]map[reflect.Type]TypeHandler
	defaultTypeHandlers  map[reflect.Type]TypeHandler
	nullTypeHandlerTypes map[reflect.Type]TypeHandler
	mutex                sync.RWMutex
}

// TypeHandler 类型处理器接口
type TypeHandler interface {
	SetParameter(stmt any, i int, parameter any, jdbcType string) error
	GetResult(rs any, columnName string) (any, error)
	GetResultByIndex(rs any, columnIndex int) (any, error)
}

// MapperMethod 映射器方法
type MapperMethod struct {
	Command         *SqlCommand
	MethodSignature *MethodSignature
}

// SqlCommand SQL命令
type SqlCommand struct {
	Name string
	Type SqlCommandType
}

// SqlCommandType SQL命令类型
type SqlCommandType int

const (
	SqlCommandTypeUnknown SqlCommandType = iota
	SqlCommandTypeInsert
	SqlCommandTypeUpdate
	SqlCommandTypeDelete
	SqlCommandTypeSelect
	SqlCommandTypeFlush
)

// MethodSignature 方法签名
type MethodSignature struct {
	ReturnsMany        bool
	ReturnsMap         bool
	ReturnsVoid        bool
	ReturnsCursor      bool
	ReturnsOptional    bool
	MapKey             string
	ResultHandlerIndex *int
	RowBoundsIndex     *int
	ParamNameResolver  *ParamNameResolver
}

// ParamNameResolver 参数名解析器
type ParamNameResolver struct {
	names              []string
	hasParamAnnotation bool
}

// getDefaultDatabaseConfig 获取默认数据库配置
func getDefaultDatabaseConfig() *config.DatabaseConfig {
	cfg := &config.DatabaseConfig{}
	
	// 设置主数据库默认配置
	cfg.Primary.Driver = "mysql"
	cfg.Primary.Host = "localhost"
	cfg.Primary.Port = 3306
	cfg.Primary.Database = "yyhertz"
	cfg.Primary.Username = "root"
	cfg.Primary.Password = ""
	cfg.Primary.Charset = "utf8mb4"
	cfg.Primary.Collation = "utf8mb4_unicode_ci"
	cfg.Primary.Timezone = "Local"
	cfg.Primary.MaxOpenConns = 100
	cfg.Primary.MaxIdleConns = 10
	cfg.Primary.ConnMaxLifetime = "1h"
	cfg.Primary.ConnMaxIdleTime = "30m"
	cfg.Primary.SlowQueryThreshold = "200ms"
	cfg.Primary.LogLevel = "warn"
	cfg.Primary.EnableMetrics = true
	cfg.Primary.EnableAutoMigration = false
	cfg.Primary.MigrationTableName = "schema_migrations"
	cfg.Primary.SSLMode = "disable"
	
	// 设置GORM默认配置
	cfg.GORM.Enable = true
	cfg.GORM.DisableForeignKeyConstrain = false
	cfg.GORM.SkipDefaultTransaction = false
	cfg.GORM.FullSaveAssociations = false
	cfg.GORM.DryRun = false
	cfg.GORM.PrepareStmt = true
	cfg.GORM.DisableNestedTransaction = false
	cfg.GORM.AllowGlobalUpdate = false
	cfg.GORM.QueryFields = true
	cfg.GORM.CreateBatchSize = 1000
	cfg.GORM.NamingStrategy = "snake_case"
	cfg.GORM.TablePrefix = ""
	cfg.GORM.SingularTable = false
	
	// 设置MyBatis默认配置
	cfg.MyBatis.Enable = true
	cfg.MyBatis.ConfigFile = "./config/mybatis-config.xml"
	cfg.MyBatis.MapperLocations = "./mappers/*.xml"
	cfg.MyBatis.TypeAliasesPath = "./models"
	cfg.MyBatis.CacheEnabled = true
	cfg.MyBatis.LazyLoading = false
	cfg.MyBatis.LogImpl = "STDOUT_LOGGING"
	cfg.MyBatis.MapUnderscoreMap = true
	
	// 设置连接池默认配置
	cfg.Pool.Enable = true
	cfg.Pool.Type = "default"
	cfg.Pool.MaxActiveConns = 100
	cfg.Pool.MaxIdleConns = 10
	cfg.Pool.MinIdleConns = 5
	cfg.Pool.MaxWaitTime = "30s"
	cfg.Pool.TimeBetweenEviction = "30s"
	cfg.Pool.MinEvictableTime = "5m"
	cfg.Pool.TestOnBorrow = true
	cfg.Pool.TestOnReturn = false
	cfg.Pool.TestWhileIdle = true
	cfg.Pool.ValidationQuery = "SELECT 1"
	
	return cfg
}

// NewConfiguration 创建默认配置
func NewConfiguration() *Configuration {
	cfg := &Configuration{
		DatabaseConfig:      getDefaultDatabaseConfig(),
		MapperRegistry:      NewMapperRegistry(),
		TypeAliasRegistry:   NewTypeAliasRegistry(),
		TypeHandlerRegistry: NewTypeHandlerRegistry(),

		CacheEnabled:    true,
		LocalCacheScope: LocalCacheScopeSession,
		DefaultCacheConfig: &CacheConfig{
			Implementation: "LRU",
			Enabled:        true,
			Size:           1024,
			FlushInterval:  0,
			ReadWrite:      true,
			Properties:     make(map[string]any),
		},

		DefaultExecutorType:              ExecutorTypeDefault,
		LazyLoadingEnabled:               false,
		AggressiveLazyLoading:            false,
		MultipleResultSetsEnabled:        true,
		UseColumnLabel:                   true,
		UseGeneratedKeys:                 false,
		AutoMappingBehavior:              AutoMappingBehaviorPartial,
		AutoMappingUnknownColumnBehavior: AutoMappingUnknownColumnBehaviorNone,
		MapUnderscoreToCamelCase:         false,
		CallSettersOnNulls:               false,
		UseActualParamName:               false,
	}

	// 注册默认类型别名
	cfg.registerDefaultTypeAliases()
	// 注册默认类型处理器
	cfg.registerDefaultTypeHandlers()

	return cfg
}

// NewMapperRegistry 创建映射器注册表
func NewMapperRegistry() *MapperRegistry {
	return &MapperRegistry{
		knownMappers: make(map[reflect.Type]*MapperProxyFactory),
	}
}

// NewTypeAliasRegistry 创建类型别名注册表
func NewTypeAliasRegistry() *TypeAliasRegistry {
	return &TypeAliasRegistry{
		aliases: make(map[string]reflect.Type),
	}
}

// NewTypeHandlerRegistry 创建类型处理器注册表
func NewTypeHandlerRegistry() *TypeHandlerRegistry {
	return &TypeHandlerRegistry{
		typeHandlerMap:       make(map[string]map[reflect.Type]TypeHandler),
		defaultTypeHandlers:  make(map[reflect.Type]TypeHandler),
		nullTypeHandlerTypes: make(map[reflect.Type]TypeHandler),
	}
}

// RegisterMapper 注册映射器
func (mr *MapperRegistry) RegisterMapper(mapperType reflect.Type) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	if mapperType.Kind() != reflect.Interface {
		return fmt.Errorf("mapper must be interface, got %s", mapperType.Kind())
	}

	if _, exists := mr.knownMappers[mapperType]; exists {
		return fmt.Errorf("mapper %s already registered", mapperType.Name())
	}

	factory := &MapperProxyFactory{
		mapperInterface: mapperType,
		methodCache:     make(map[string]*MapperMethod),
	}

	mr.knownMappers[mapperType] = factory
	return nil
}

// GetMapper 获取映射器
func (mr *MapperRegistry) GetMapper(mapperType reflect.Type, sqlSession any) (any, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	factory, exists := mr.knownMappers[mapperType]
	if !exists {
		return nil, fmt.Errorf("mapper %s not registered", mapperType.Name())
	}

	return factory.NewInstance(sqlSession), nil
}

// HasMapper 检查是否有映射器
func (mr *MapperRegistry) HasMapper(mapperType reflect.Type) bool {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	_, exists := mr.knownMappers[mapperType]
	return exists
}

// GetMappers 获取所有映射器类型
func (mr *MapperRegistry) GetMappers() []reflect.Type {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	mappers := make([]reflect.Type, 0, len(mr.knownMappers))
	for mapperType := range mr.knownMappers {
		mappers = append(mappers, mapperType)
	}
	return mappers
}

// NewInstance 创建映射器实例
func (factory *MapperProxyFactory) NewInstance(sqlSession any) any {
	return NewMapperProxy(factory.mapperInterface, factory.methodCache, sqlSession)
}

// RegisterAlias 注册类型别名
func (tar *TypeAliasRegistry) RegisterAlias(alias string, value reflect.Type) {
	tar.mutex.Lock()
	defer tar.mutex.Unlock()

	tar.aliases[alias] = value
}

// ResolveAlias 解析类型别名
func (tar *TypeAliasRegistry) ResolveAlias(alias string) (reflect.Type, bool) {
	tar.mutex.RLock()
	defer tar.mutex.RUnlock()

	t, exists := tar.aliases[alias]
	return t, exists
}

// RegisterTypeHandler 注册类型处理器
func (thr *TypeHandlerRegistry) RegisterTypeHandler(javaType reflect.Type, handler TypeHandler) {
	thr.mutex.Lock()
	defer thr.mutex.Unlock()

	thr.defaultTypeHandlers[javaType] = handler
}

// RegisterTypeHandlerWithJdbcType 注册带JDBC类型的类型处理器
func (thr *TypeHandlerRegistry) RegisterTypeHandlerWithJdbcType(javaType reflect.Type, jdbcType string, handler TypeHandler) {
	thr.mutex.Lock()
	defer thr.mutex.Unlock()

	if thr.typeHandlerMap[jdbcType] == nil {
		thr.typeHandlerMap[jdbcType] = make(map[reflect.Type]TypeHandler)
	}
	thr.typeHandlerMap[jdbcType][javaType] = handler
}

// GetTypeHandler 获取类型处理器
func (thr *TypeHandlerRegistry) GetTypeHandler(javaType reflect.Type, jdbcType string) TypeHandler {
	thr.mutex.RLock()
	defer thr.mutex.RUnlock()

	// 先尝试精确匹配
	if jdbcType != "" {
		if handlers, exists := thr.typeHandlerMap[jdbcType]; exists {
			if handler, exists := handlers[javaType]; exists {
				return handler
			}
		}
	}

	// 尝试默认处理器
	if handler, exists := thr.defaultTypeHandlers[javaType]; exists {
		return handler
	}

	return nil
}

// registerDefaultTypeAliases 注册默认类型别名
func (c *Configuration) registerDefaultTypeAliases() {
	c.TypeAliasRegistry.RegisterAlias("string", reflect.TypeOf(""))
	c.TypeAliasRegistry.RegisterAlias("int", reflect.TypeOf(0))
	c.TypeAliasRegistry.RegisterAlias("int32", reflect.TypeOf(int32(0)))
	c.TypeAliasRegistry.RegisterAlias("int64", reflect.TypeOf(int64(0)))
	c.TypeAliasRegistry.RegisterAlias("float32", reflect.TypeOf(float32(0)))
	c.TypeAliasRegistry.RegisterAlias("float64", reflect.TypeOf(float64(0)))
	c.TypeAliasRegistry.RegisterAlias("bool", reflect.TypeOf(false))
	c.TypeAliasRegistry.RegisterAlias("time", reflect.TypeOf(time.Time{}))
	c.TypeAliasRegistry.RegisterAlias("bytes", reflect.TypeOf([]byte{}))
}

// registerDefaultTypeHandlers 注册默认类型处理器
func (c *Configuration) registerDefaultTypeHandlers() {
	// 这里会注册各种基本类型的处理器
	// 具体实现在后续的类型处理器文件中
}

// GetDatabaseConfig 获取数据库配置
func (c *Configuration) GetDatabaseConfig() *config.DatabaseConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.DatabaseConfig
}

// SetDatabaseConfig 设置数据库配置
func (c *Configuration) SetDatabaseConfig(config *config.DatabaseConfig) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.DatabaseConfig = config
}

// GetMapperRegistry 获取映射器注册表
func (c *Configuration) GetMapperRegistry() *MapperRegistry {
	return c.MapperRegistry
}

// GetTypeAliasRegistry 获取类型别名注册表
func (c *Configuration) GetTypeAliasRegistry() *TypeAliasRegistry {
	return c.TypeAliasRegistry
}

// GetTypeHandlerRegistry 获取类型处理器注册表
func (c *Configuration) GetTypeHandlerRegistry() *TypeHandlerRegistry {
	return c.TypeHandlerRegistry
}

// MappedStatementRegistry 映射语句注册表
type MappedStatementRegistry struct {
	statements map[string]*MappedStatement
	mutex      sync.RWMutex
}

// MappedStatement 映射语句
type MappedStatement struct {
	ID        string
	Namespace string
	SQL       string
	SqlType   StatementType
}

// StatementType 语句类型
type StatementType int

const (
	StatementTypeSelect StatementType = iota
	StatementTypeInsert
	StatementTypeUpdate
	StatementTypeDelete
)

// NewMappedStatementRegistry 创建映射语句注册表
func NewMappedStatementRegistry() *MappedStatementRegistry {
	return &MappedStatementRegistry{
		statements: make(map[string]*MappedStatement),
	}
}

// RegisterStatement 注册语句
func (r *MappedStatementRegistry) RegisterStatement(id string, stmt *MappedStatement) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.statements[id] = stmt
}

// GetStatement 获取语句
func (r *MappedStatementRegistry) GetStatement(id string) *MappedStatement {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.statements[id]
}

// 在Configuration中添加映射语句注册表
func (c *Configuration) GetMappedStatement(id string) *MappedStatement {
	// 这里简化实现，实际应该有专门的映射语句注册表
	return &MappedStatement{
		ID:        id,
		Namespace: "default",
		SQL:       "SELECT 1", // 简化的SQL
		SqlType:   StatementTypeSelect,
	}
}
