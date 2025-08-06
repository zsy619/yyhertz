// Package session SQL会话工厂实现
//
// 负责创建和管理SQL会话实例
package session

import (
	"sync"

	"github.com/zsy619/yyhertz/framework/mybatis/cache"
	"github.com/zsy619/yyhertz/framework/mybatis/config"
	"github.com/zsy619/yyhertz/framework/orm"
)

// SqlSessionFactory SQL会话工厂接口
type SqlSessionFactory interface {
	// OpenSession 打开新会话
	OpenSession() SqlSession

	// OpenSessionWithAutoCommit 打开带自动提交的会话
	OpenSessionWithAutoCommit(autoCommit bool) SqlSession

	// OpenSessionWithConnection 打开带连接的会话
	OpenSessionWithConnection(connection any) SqlSession

	// OpenSessionWithExecutorType 打开带执行器类型的会话
	OpenSessionWithExecutorType(execType config.ExecutorType) SqlSession

	// GetConfiguration 获取配置
	GetConfiguration() *config.Configuration
}

// DefaultSqlSessionFactory 默认SQL会话工厂
type DefaultSqlSessionFactory struct {
	configuration *config.Configuration
	orm           *orm.ORM
	cache         cache.Cache
	mutex         sync.RWMutex
}

// SqlSessionManager SQL会话管理器
type SqlSessionManager struct {
	factory         SqlSessionFactory
	localSqlSession *ThreadLocal
	mutex           sync.RWMutex
}

// ThreadLocal 线程本地存储 (Go中用Goroutine本地存储模拟)
type ThreadLocal struct {
	data  map[int64]SqlSession
	mutex sync.RWMutex
}

// ExecutorFactory 执行器工厂
type ExecutorFactory struct {
	configuration *config.Configuration
}

// NewDefaultSqlSessionFactory 创建默认SQL会话工厂
func NewDefaultSqlSessionFactory(configuration *config.Configuration) (SqlSessionFactory, error) {
	if configuration == nil {
		configuration = config.NewConfiguration()
	}

	// 创建ORM实例
	var ormInstance *orm.ORM
	var err error

	if configuration.GetDatabaseConfig() != nil {
		// 转换配置类型
		primaryConfig := configuration.GetDatabaseConfig().Primary
		dbConfig := &orm.DatabaseConfig{
			Type:         primaryConfig.Driver,
			Host:         primaryConfig.Host,
			Port:         primaryConfig.Port,
			Database:     primaryConfig.Database,
			Username:     primaryConfig.Username,
			Password:     primaryConfig.Password,
			Charset:      primaryConfig.Charset,
			Timezone:     primaryConfig.Timezone,
			MaxIdleConns: primaryConfig.MaxIdleConns,
			MaxOpenConns: primaryConfig.MaxOpenConns,
			LogLevel:     primaryConfig.LogLevel,
		}

		// 解析连接最大生存时间
		if primaryConfig.ConnMaxLifetime != "" {
			// 简化处理，默认设置为3600秒
			dbConfig.MaxLifetime = 3600
		}

		// 解析慢查询阈值
		if primaryConfig.SlowQueryThreshold != "" {
			// 简化处理，默认设置为200毫秒
			dbConfig.SlowQuery = 200
		}

		ormInstance, err = orm.NewORM(dbConfig)
		if err != nil {
			return nil, err
		}
	}

	// 创建缓存
	var c cache.Cache
	if configuration.CacheEnabled && configuration.DefaultCacheConfig != nil {
		perpetualCache := cache.NewPerpetualCache("factory")
		c = cache.NewLruCache(perpetualCache, configuration.DefaultCacheConfig.Size)
	}

	factory := &DefaultSqlSessionFactory{
		configuration: configuration,
		orm:           ormInstance,
		cache:         c,
	}

	return factory, nil
}

// OpenSession 打开新会话
func (factory *DefaultSqlSessionFactory) OpenSession() SqlSession {
	return factory.OpenSessionWithAutoCommit(false)
}

// OpenSessionWithAutoCommit 打开带自动提交的会话
func (factory *DefaultSqlSessionFactory) OpenSessionWithAutoCommit(autoCommit bool) SqlSession {
	return factory.OpenSessionWithExecutorTypeAndAutoCommit(config.ExecutorTypeDefault, autoCommit)
}

// OpenSessionWithConnection 打开带连接的会话
func (factory *DefaultSqlSessionFactory) OpenSessionWithConnection(connection any) SqlSession {
	// 这里可以传入自定义连接
	return factory.OpenSessionWithAutoCommit(false)
}

// OpenSessionWithExecutorType 打开带执行器类型的会话
func (factory *DefaultSqlSessionFactory) OpenSessionWithExecutorType(execType config.ExecutorType) SqlSession {
	return factory.OpenSessionWithExecutorTypeAndAutoCommit(execType, false)
}

// OpenSessionWithExecutorTypeAndAutoCommit 打开带执行器类型和自动提交的会话
func (factory *DefaultSqlSessionFactory) OpenSessionWithExecutorTypeAndAutoCommit(execType config.ExecutorType, autoCommit bool) SqlSession {
	executor := factory.createExecutor(execType, autoCommit)
	return NewDefaultSqlSession(factory.configuration, executor, autoCommit)
}

// GetConfiguration 获取配置
func (factory *DefaultSqlSessionFactory) GetConfiguration() *config.Configuration {
	factory.mutex.RLock()
	defer factory.mutex.RUnlock()
	return factory.configuration
}

// createExecutor 创建执行器
func (factory *DefaultSqlSessionFactory) createExecutor(execType config.ExecutorType, autoCommit bool) Executor {
	// 根据类型创建执行器
	var executor Executor

	switch execType {
	case config.ExecutorTypeReuse:
		executor = NewReuseExecutor(factory.configuration, factory.orm.DB())
	case config.ExecutorTypeBatch:
		executor = NewBatchExecutor(factory.configuration, factory.orm.DB())
	default:
		executor = NewDefaultExecutor(factory.configuration, factory.orm.DB())
	}

	// 应用插件
	executor = factory.applyPlugins(executor)

	// 包装缓存执行器
	if factory.configuration.CacheEnabled {
		executor = NewCachingExecutor(executor, factory.cache)
	}

	return executor
}

// applyPlugins 应用插件
func (factory *DefaultSqlSessionFactory) applyPlugins(executor Executor) Executor {
	// 这里可以应用各种插件
	// 插件系统在后续实现
	return executor
}

// NewSqlSessionManager 创建SQL会话管理器
func NewSqlSessionManager(factory SqlSessionFactory) *SqlSessionManager {
	return &SqlSessionManager{
		factory:         factory,
		localSqlSession: NewThreadLocal(),
	}
}

// StartManagedSession 启动托管会话
func (manager *SqlSessionManager) StartManagedSession() {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	session := manager.factory.OpenSession()
	manager.localSqlSession.Set(session)
}

// StartManagedSessionWithAutoCommit 启动带自动提交的托管会话
func (manager *SqlSessionManager) StartManagedSessionWithAutoCommit(autoCommit bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	session := manager.factory.OpenSessionWithAutoCommit(autoCommit)
	manager.localSqlSession.Set(session)
}

// StartManagedSessionWithExecutorType 启动带执行器类型的托管会话
func (manager *SqlSessionManager) StartManagedSessionWithExecutorType(execType config.ExecutorType) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	session := manager.factory.OpenSessionWithExecutorType(execType)
	manager.localSqlSession.Set(session)
}

// GetManagedSession 获取托管会话
func (manager *SqlSessionManager) GetManagedSession() SqlSession {
	return manager.localSqlSession.Get()
}

// CloseManagedSession 关闭托管会话
func (manager *SqlSessionManager) CloseManagedSession() error {
	session := manager.localSqlSession.Get()
	if session != nil {
		manager.localSqlSession.Remove()
		return session.Close()
	}
	return nil
}

// IsManagedSessionStarted 检查托管会话是否启动
func (manager *SqlSessionManager) IsManagedSessionStarted() bool {
	return manager.localSqlSession.Get() != nil
}

// NewThreadLocal 创建线程本地存储
func NewThreadLocal() *ThreadLocal {
	return &ThreadLocal{
		data: make(map[int64]SqlSession),
	}
}

// Set 设置值
func (tl *ThreadLocal) Set(session SqlSession) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	// 使用goroutine ID作为键 (简化实现)
	// 实际应该使用更可靠的goroutine本地存储
	goroutineID := getGoroutineID()
	tl.data[goroutineID] = session
}

// Get 获取值
func (tl *ThreadLocal) Get() SqlSession {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()

	goroutineID := getGoroutineID()
	return tl.data[goroutineID]
}

// Remove 移除值
func (tl *ThreadLocal) Remove() {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	goroutineID := getGoroutineID()
	delete(tl.data, goroutineID)
}

// getGoroutineID 获取Goroutine ID (简化实现)
func getGoroutineID() int64 {
	// 这是一个简化实现
	// 实际应该使用更可靠的方法获取goroutine ID
	// 或者使用context.Context来传递会话
	return 1
}

// NewExecutorFactory 创建执行器工厂
func NewExecutorFactory(configuration *config.Configuration) *ExecutorFactory {
	return &ExecutorFactory{
		configuration: configuration,
	}
}

// CreateExecutor 创建执行器
func (factory *ExecutorFactory) CreateExecutor(execType config.ExecutorType, db any) Executor {
	switch execType {
	case config.ExecutorTypeReuse:
		return NewReuseExecutor(factory.configuration, db)
	case config.ExecutorTypeBatch:
		return NewBatchExecutor(factory.configuration, db)
	default:
		return NewDefaultExecutor(factory.configuration, db)
	}
}

// SqlSessionTemplate SQL会话模板 (类似Spring的SqlSessionTemplate)
type SqlSessionTemplate struct {
	sqlSessionFactory   SqlSessionFactory
	executorType        config.ExecutorType
	exceptionTranslator ExceptionTranslator
}

// ExceptionTranslator 异常转换器
type ExceptionTranslator interface {
	Translate(task string, sql string, ex error) error
}

// NewSqlSessionTemplate 创建SQL会话模板
func NewSqlSessionTemplate(sqlSessionFactory SqlSessionFactory) *SqlSessionTemplate {
	return &SqlSessionTemplate{
		sqlSessionFactory: sqlSessionFactory,
		executorType:      config.ExecutorTypeDefault,
	}
}

// SetExecutorType 设置执行器类型
func (template *SqlSessionTemplate) SetExecutorType(executorType config.ExecutorType) {
	template.executorType = executorType
}

// SetExceptionTranslator 设置异常转换器
func (template *SqlSessionTemplate) SetExceptionTranslator(exceptionTranslator ExceptionTranslator) {
	template.exceptionTranslator = exceptionTranslator
}

// Execute 执行操作
func (template *SqlSessionTemplate) Execute(callback func(session SqlSession) (any, error)) (any, error) {
	session := template.getSqlSession()
	defer template.closeSqlSession(session)

	return callback(session)
}

// getSqlSession 获取SQL会话
func (template *SqlSessionTemplate) getSqlSession() SqlSession {
	return template.sqlSessionFactory.OpenSessionWithExecutorType(template.executorType)
}

// closeSqlSession 关闭SQL会话
func (template *SqlSessionTemplate) closeSqlSession(session SqlSession) {
	if session != nil {
		session.Close()
	}
}
