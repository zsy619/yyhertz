// Package session 提供MyBatis会话管理
//
// 实现SQL会话接口，提供数据库操作的统一入口
package session

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"gorm.io/gorm"
	
	"github.com/zsy619/yyhertz/framework/mybatis/config"
	"github.com/zsy619/yyhertz/framework/mybatis/cache"
)

// SqlSession SQL会话接口
type SqlSession interface {
	// SelectOne 查询单条记录
	SelectOne(statement string, parameter any) (any, error)
	
	// SelectList 查询多条记录
	SelectList(statement string, parameter any) ([]any, error)
	
	// SelectMap 查询返回Map
	SelectMap(statement string, parameter any) (map[string]any, error)
	
	// Insert 插入记录
	Insert(statement string, parameter any) (int64, error)
	
	// Update 更新记录
	Update(statement string, parameter any) (int64, error)
	
	// Delete 删除记录
	Delete(statement string, parameter any) (int64, error)
	
	// GetMapper 获取映射器
	GetMapper(mapperType reflect.Type) (any, error)
	
	// GetConfiguration 获取配置
	GetConfiguration() *config.Configuration
	
	// GetConnection 获取连接
	GetConnection() *gorm.DB
	
	// Commit 提交事务
	Commit() error
	
	// Rollback 回滚事务
	Rollback() error
	
	// Close 关闭会话
	Close() error
	
	// ClearCache 清除缓存
	ClearCache()
	
	// SelectCursor 查询游标 (Go中可用channel模拟)
	SelectCursor(statement string, parameter any) (<-chan any, error)
}

// DefaultSqlSession 默认SQL会话实现
type DefaultSqlSession struct {
	configuration *config.Configuration
	executor      Executor
	autoCommit    bool
	dirty         bool
	cursorList    []any
	mutex         sync.RWMutex
}

// Executor 执行器接口
type Executor interface {
	Update(ms *MappedStatement, parameter any) (int64, error)
	Query(ms *MappedStatement, parameter any, rowBounds *RowBounds, resultHandler ResultHandler, cacheKey *CacheKey, boundSql *BoundSql) ([]any, error)
	QueryCursor(ms *MappedStatement, parameter any, rowBounds *RowBounds) (<-chan any, error)
	Commit(required bool) error
	Rollback(required bool) error
	Close(forceRollback bool) error
	IsClosed() bool
	ClearLocalCache()
	CreateCacheKey(ms *MappedStatement, parameterObject any, rowBounds *RowBounds, boundSql *BoundSql) *CacheKey
	IsCached(ms *MappedStatement, key *CacheKey) bool
	GetConnection() *gorm.DB
	SetExecutorWrapper(wrapper ExecutorWrapper)
}

// MappedStatement 映射语句
type MappedStatement struct {
	ID              string                 // 语句ID
	Configuration   *config.Configuration  // 配置
	SqlSource       SqlSource              // SQL源
	Cache           cache.Cache            // 缓存
	ParameterMap    *ParameterMap          // 参数映射
	ResultMaps      []*ResultMap           // 结果映射
	StatementType   StatementType          // 语句类型
	SqlCommandType  config.SqlCommandType  // SQL命令类型
	FetchSize       *int                   // 获取大小
	Timeout         *time.Duration         // 超时时间
	KeyGenerator    KeyGenerator           // 主键生成器
	KeyProperty     string                 // 主键属性
	KeyColumn       string                 // 主键列
	DatabaseId      string                 // 数据库ID
	Lang            LanguageDriver         // 语言驱动
	ResultSetType   ResultSetType          // 结果集类型
	FlushCacheRequired bool                // 是否需要刷新缓存
	UseCache        bool                   // 是否使用缓存
	ResultOrdered   bool                   // 结果是否有序
}

// 创建转换函数，从config.MappedStatement转换为session.MappedStatement
func convertMappedStatement(configMS *config.MappedStatement) *MappedStatement {
	if configMS == nil {
		return nil
	}
	
	return &MappedStatement{
		ID:              configMS.ID,
		SqlSource:       &StaticSqlSource{SQL: configMS.SQL},
		StatementType:   StatementTypeUnknown, // 简化映射
		SqlCommandType:  config.SqlCommandTypeSelect, // 简化映射
		UseCache:        true,
		FlushCacheRequired: false,
	}
}

// SqlSource SQL源接口
type SqlSource interface {
	GetBoundSql(parameterObject any) *BoundSql
}

// StaticSqlSource 静态SQL源
type StaticSqlSource struct {
	SQL string
}

// GetBoundSql 获取绑定SQL
func (s *StaticSqlSource) GetBoundSql(parameterObject any) *BoundSql {
	return &BoundSql{
		Sql:                s.SQL,
		ParameterMappings:  make([]*ParameterMapping, 0),
		ParameterObject:    parameterObject,
		AdditionalParameters: make(map[string]any),
	}
}

// BoundSql 绑定SQL
type BoundSql struct {
	Sql                string
	ParameterMappings  []*ParameterMapping
	ParameterObject    any
	AdditionalParameters map[string]any
	MetaParameters     *MetaObject
}

// ParameterMapping 参数映射
type ParameterMapping struct {
	Property     string
	Mode         ParameterMode
	JavaType     reflect.Type
	JdbcType     string
	TypeHandler  config.TypeHandler
	Expression   string
}

// ParameterMode 参数模式
type ParameterMode int

const (
	ParameterModeIn ParameterMode = iota
	ParameterModeOut
	ParameterModeInOut
)

// ParameterMap 参数映射
type ParameterMap struct {
	ID                string
	Type              reflect.Type
	ParameterMappings []*ParameterMapping
}

// ResultMap 结果映射
type ResultMap struct {
	ID                string
	Type              reflect.Type
	ResultMappings    []*ResultMapping
	IdResultMappings  []*ResultMapping
	ConstructorResultMappings []*ResultMapping
	PropertyResultMappings    []*ResultMapping
	MappedColumns     []string
	MappedProperties  []string
	Discriminator     *Discriminator
	HasNestedResultMaps bool
	HasNestedQueries    bool
	AutoMapping         *bool
}

// ResultMapping 结果映射
type ResultMapping struct {
	Configuration *config.Configuration
	Property      string
	Column        string
	JavaType      reflect.Type
	JdbcType      string
	TypeHandler   config.TypeHandler
	NestedResultMapId string
	NestedQueryId     string
	NotNullColumns    []string
	ColumnPrefix      string
	Flags             []ResultFlag
	CompositeColumns  []string
	ResultSet         string
	ForeignColumn     string
	Lazy              bool
}

// ResultFlag 结果标记
type ResultFlag int

const (
	ResultFlagId ResultFlag = iota
	ResultFlagConstructor
)

// Discriminator 鉴别器
type Discriminator struct {
	ResultMapping *ResultMapping
	DiscriminatorMap map[string]string
}

// StatementType 语句类型
type StatementType int

const (
	StatementTypeUnknown StatementType = iota
	StatementTypeStatement
	StatementTypePrepared
	StatementTypeCallable
)

// ResultSetType 结果集类型
type ResultSetType int

const (
	ResultSetTypeDefault ResultSetType = iota
	ResultSetTypeForwardOnly
	ResultSetTypeScrollInsensitive
	ResultSetTypeScrollSensitive
)

// KeyGenerator 主键生成器接口
type KeyGenerator interface {
	ProcessBefore(executor Executor, ms *MappedStatement, stmt any, parameter any)
	ProcessAfter(executor Executor, ms *MappedStatement, stmt any, parameter any)
}

// LanguageDriver 语言驱动接口
type LanguageDriver interface {
	CreateParameterHandler(ms *MappedStatement, parameterObject any, boundSql *BoundSql) ParameterHandler
	CreateSqlSource(configuration *config.Configuration, script string, parameterType reflect.Type) SqlSource
}

// ParameterHandler 参数处理器接口
type ParameterHandler interface {
	GetParameterObject() any
	SetParameters(ps any) error
}

// ResultHandler 结果处理器接口
type ResultHandler interface {
	HandleResult(resultContext *ResultContext) error
}

// ResultContext 结果上下文
type ResultContext struct {
	ResultObject   any
	ResultCount    int
	Stopped        bool
}

// RowBounds 行边界
type RowBounds struct {
	Offset int
	Limit  int
}

// CacheKey 缓存键
type CacheKey struct {
	UpdateList []any
	HashCode   int
	CheckSum   int64
	Count      int
}

// ExecutorWrapper 执行器包装器
type ExecutorWrapper interface {
	Wrap(executor Executor) Executor
}

// MetaObject 元对象
type MetaObject struct {
	OriginalObject any
	ObjectWrapper  ObjectWrapper
}

// ObjectWrapper 对象包装器接口
type ObjectWrapper interface {
	Get(prop string) any
	Set(prop string, value any)
	FindProperty(name string, useCamelCaseMapping bool) string
	GetGetterNames() []string
	GetSetterNames() []string
	GetSetterType(name string) reflect.Type
	GetGetterType(name string) reflect.Type
	HasSetter(name string) bool
	HasGetter(name string) bool
	Instantiate() any
	IsCollection() bool
	Add(element any)
	AddAll(elements []any)
}

// NewDefaultSqlSession 创建默认SQL会话
func NewDefaultSqlSession(configuration *config.Configuration, executor Executor, autoCommit bool) *DefaultSqlSession {
	return &DefaultSqlSession{
		configuration: configuration,
		executor:      executor,
		autoCommit:    autoCommit,
		dirty:         false,
		cursorList:    make([]any, 0),
	}
}

// SelectOne 查询单条记录
func (session *DefaultSqlSession) SelectOne(statement string, parameter any) (any, error) {
	list, err := session.SelectList(statement, parameter)
	if err != nil {
		return nil, err
	}
	
	if len(list) == 1 {
		return list[0], nil
	} else if len(list) > 1 {
		return nil, fmt.Errorf("expected one result (or null) to be returned by selectOne(), but found: %d", len(list))
	}
	return nil, nil
}

// SelectList 查询多条记录
func (session *DefaultSqlSession) SelectList(statement string, parameter any) ([]any, error) {
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	
	configMS := session.configuration.GetMappedStatement(statement)
	if configMS == nil {
		return nil, fmt.Errorf("mapped statement not found: %s", statement)
	}
	
	ms := convertMappedStatement(configMS)
	return session.executor.Query(ms, parameter, &RowBounds{Offset: 0, Limit: -1}, nil, nil, nil)
}

// SelectMap 查询返回Map
func (session *DefaultSqlSession) SelectMap(statement string, parameter any) (map[string]any, error) {
	result, err := session.SelectOne(statement, parameter)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]any); ok {
		return resultMap, nil
	}
	
	return session.convertToMap(result), nil
}

// Insert 插入记录
func (session *DefaultSqlSession) Insert(statement string, parameter any) (int64, error) {
	return session.update(statement, parameter)
}

// Update 更新记录
func (session *DefaultSqlSession) Update(statement string, parameter any) (int64, error) {
	return session.update(statement, parameter)
}

// Delete 删除记录
func (session *DefaultSqlSession) Delete(statement string, parameter any) (int64, error) {
	return session.update(statement, parameter)
}

// update 执行更新操作
func (session *DefaultSqlSession) update(statement string, parameter any) (int64, error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	
	session.dirty = true
	
	configMS := session.configuration.GetMappedStatement(statement)
	if configMS == nil {
		return 0, fmt.Errorf("mapped statement not found: %s", statement)
	}
	
	ms := convertMappedStatement(configMS)
	return session.executor.Update(ms, parameter)
}

// GetMapper 获取映射器
func (session *DefaultSqlSession) GetMapper(mapperType reflect.Type) (any, error) {
	return session.configuration.GetMapperRegistry().GetMapper(mapperType, session)
}

// GetConfiguration 获取配置
func (session *DefaultSqlSession) GetConfiguration() *config.Configuration {
	return session.configuration
}

// GetConnection 获取连接
func (session *DefaultSqlSession) GetConnection() *gorm.DB {
	return session.executor.GetConnection()
}

// Commit 提交事务
func (session *DefaultSqlSession) Commit() error {
	return session.commit(false)
}

// commit 提交事务
func (session *DefaultSqlSession) commit(force bool) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	
	if session.dirty || force {
		session.dirty = false
		return session.executor.Commit(true)
	}
	return nil
}

// Rollback 回滚事务
func (session *DefaultSqlSession) Rollback() error {
	return session.rollback(false)
}

// rollback 回滚事务
func (session *DefaultSqlSession) rollback(force bool) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	
	if session.dirty || force {
		session.dirty = false
		return session.executor.Rollback(true)
	}
	return nil
}

// Close 关闭会话
func (session *DefaultSqlSession) Close() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	
	session.closeCursors()
	
	if session.dirty && !session.autoCommit {
		session.rollback(true)
	}
	
	return session.executor.Close(false)
}

// ClearCache 清除缓存
func (session *DefaultSqlSession) ClearCache() {
	session.executor.ClearLocalCache()
}

// SelectCursor 查询游标
func (session *DefaultSqlSession) SelectCursor(statement string, parameter any) (<-chan any, error) {
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	
	configMS := session.configuration.GetMappedStatement(statement)
	if configMS == nil {
		return nil, fmt.Errorf("mapped statement not found: %s", statement)
	}
	
	ms := convertMappedStatement(configMS)
	cursor, err := session.executor.QueryCursor(ms, parameter, &RowBounds{Offset: 0, Limit: -1})
	if err != nil {
		return nil, err
	}
	
	session.cursorList = append(session.cursorList, cursor)
	return cursor, nil
}

// closeCursors 关闭游标
func (session *DefaultSqlSession) closeCursors() {
	for _, cursor := range session.cursorList {
		if ch, ok := cursor.(<-chan any); ok {
			// 关闭channel
			go func() {
				for range ch {
					// 消费剩余数据
				}
			}()
		}
	}
	session.cursorList = session.cursorList[:0]
}

// convertToMap 转换为Map
func (session *DefaultSqlSession) convertToMap(obj any) map[string]any {
	result := make(map[string]any)
	if obj == nil {
		return result
	}
	
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		result["value"] = obj
		return result
	}
	
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			result[field.Name] = v.Field(i).Interface()
		}
	}
	
	return result
}

