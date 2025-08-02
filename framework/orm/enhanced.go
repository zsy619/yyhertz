package orm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// EnhancedORM 增强的ORM管理器，支持连接池
type EnhancedORM struct {
	*ORM
	pool             ConnectionPool
	metricsCollector *MetricsCollector
}

// NewEnhancedORM 创建增强的ORM实例
func NewEnhancedORM(dbConfig *DatabaseConfig, poolConfig *PoolConfig) (*EnhancedORM, error) {
	if dbConfig == nil {
		dbConfig = DefaultDatabaseConfig()
	}

	if poolConfig == nil {
		poolConfig = DefaultPoolConfig()
	}

	// 创建数据库节点
	node := &DatabaseNode{
		ID:       "primary",
		Config:   dbConfig,
		Weight:   10,
		IsMaster: true,
	}

	// 创建连接池
	pool, err := NewMultiNodePool(poolConfig, []*DatabaseNode{node})
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 获取主连接用于初始化
	ctx := context.Background()
	db, err := pool.GetMasterConnection(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to get initial connection: %w", err)
	}

	// 创建基础ORM
	baseORM := &ORM{
		db:       db,
		config:   dbConfig,
		migrator: db.Migrator(),
		// logger:   &gormLogWriter{},
	}

	// 创建指标收集器
	metricsCollector := NewMetricsCollector()
	metricsCollector.Start()

	enhancedORM := &EnhancedORM{
		ORM:              baseORM,
		pool:             pool,
		metricsCollector: metricsCollector,
	}

	config.Info("Enhanced ORM with connection pool initialized")
	return enhancedORM, nil
}

// NewEnhancedORMWithMasterSlave 创建主从架构的增强ORM
func NewEnhancedORMWithMasterSlave(masterConfig, slaveConfig *DatabaseConfig, poolConfig *PoolConfig) (*EnhancedORM, error) {
	if poolConfig == nil {
		poolConfig = DefaultPoolConfig()
	}

	// 创建主从连接池
	pool, err := CreateMasterSlavePool(masterConfig, slaveConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create master-slave pool: %w", err)
	}

	// 获取主连接用于初始化
	ctx := context.Background()
	db, err := pool.GetMasterConnection(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to get master connection: %w", err)
	}

	// 创建基础ORM
	baseORM := &ORM{
		db:       db,
		config:   masterConfig,
		migrator: db.Migrator(),
		// logger:   &gormLogWriter{},
	}

	// 创建指标收集器
	metricsCollector := NewMetricsCollector()
	metricsCollector.Start()

	enhancedORM := &EnhancedORM{
		ORM:              baseORM,
		pool:             pool,
		metricsCollector: metricsCollector,
	}

	config.Info("Enhanced ORM with master-slave pool initialized")
	return enhancedORM, nil
}

// GetMasterDB 获取主库连接
func (eo *EnhancedORM) GetMasterDB(ctx context.Context) (*gorm.DB, error) {
	db, err := eo.pool.GetMasterConnection(ctx)

	// 记录指标
	if eo.metricsCollector != nil {
		eo.metricsCollector.RecordConnection("master", "master")
		if err != nil {
			eo.metricsCollector.RecordConnectionError("master")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get master connection: %w", err)
	}

	// 添加查询钩子来记录指标
	return eo.addMetricsHooks(db, "master"), nil
}

// GetSlaveDB 获取从库连接
func (eo *EnhancedORM) GetSlaveDB(ctx context.Context) (*gorm.DB, error) {
	db, err := eo.pool.GetSlaveConnection(ctx)

	// 记录指标
	if eo.metricsCollector != nil {
		eo.metricsCollector.RecordConnection("slave", "slave")
		if err != nil {
			eo.metricsCollector.RecordConnectionError("slave")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get slave connection: %w", err)
	}

	// 添加查询钩子来记录指标
	return eo.addMetricsHooks(db, "slave"), nil
}

// GetReadDB 获取读连接（优先从库）
func (eo *EnhancedORM) GetReadDB(ctx context.Context) (*gorm.DB, error) {
	return eo.GetSlaveDB(ctx)
}

// GetWriteDB 获取写连接（主库）
func (eo *EnhancedORM) GetWriteDB(ctx context.Context) (*gorm.DB, error) {
	return eo.GetMasterDB(ctx)
}

// addMetricsHooks 添加指标钩子
func (eo *EnhancedORM) addMetricsHooks(db *gorm.DB, nodeType string) *gorm.DB {
	if eo.metricsCollector == nil {
		return db
	}

	// 添加查询前钩子
	db.Callback().Query().Before("gorm:query").Register("metrics:before_query", func(db *gorm.DB) {
		db.Set("metrics_start_time", time.Now())
		db.Set("metrics_node_type", nodeType)
	})

	// 添加查询后钩子
	db.Callback().Query().After("gorm:query").Register("metrics:after_query", func(db *gorm.DB) {
		eo.recordQueryMetrics(db, "query")
	})

	// 添加创建钩子
	db.Callback().Create().After("gorm:create").Register("metrics:after_create", func(db *gorm.DB) {
		eo.recordQueryMetrics(db, "create")
	})

	// 添加更新钩子
	db.Callback().Update().After("gorm:update").Register("metrics:after_update", func(db *gorm.DB) {
		eo.recordQueryMetrics(db, "update")
	})

	// 添加删除钩子
	db.Callback().Delete().After("gorm:delete").Register("metrics:after_delete", func(db *gorm.DB) {
		eo.recordQueryMetrics(db, "delete")
	})

	return db
}

// recordQueryMetrics 记录查询指标
func (eo *EnhancedORM) recordQueryMetrics(db *gorm.DB, operation string) {
	startTime, exists := db.Get("metrics_start_time")
	if !exists {
		return
	}

	start, ok := startTime.(time.Time)
	if !ok {
		return
	}

	duration := time.Since(start)
	success := db.Error == nil

	nodeType, _ := db.Get("metrics_node_type")
	nodeTypeStr, _ := nodeType.(string)

	// 记录基础指标
	eo.metricsCollector.RecordQuery(nodeTypeStr, duration, success)

	// 记录详细指标
	queryMetrics := &QueryMetrics{
		QueryID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		SQL:          db.Statement.SQL.String(),
		Duration:     duration,
		StartTime:    start,
		EndTime:      time.Now(),
		NodeID:       nodeTypeStr,
		Success:      success,
		RowsAffected: db.RowsAffected,
	}

	if !success && db.Error != nil {
		queryMetrics.Error = db.Error.Error()
	}

	eo.metricsCollector.RecordQueryDetails(queryMetrics)
}

// WithTimeout 设置超时上下文
func (eo *EnhancedORM) WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// Transaction 执行事务（使用主库）
func (eo *EnhancedORM) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	db, err := eo.GetWriteDB(ctx)
	if err != nil {
		return err
	}

	return db.Transaction(fn)
}

// ReadTransaction 只读事务（使用从库）
func (eo *EnhancedORM) ReadTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	db, err := eo.GetReadDB(ctx)
	if err != nil {
		return err
	}

	// 注意：这里实际上不是真正的事务，只是为了API一致性
	return fn(db)
}

// GetPoolStats 获取连接池统计信息
func (eo *EnhancedORM) GetPoolStats() *PoolStats {
	return eo.pool.Stats()
}

// GetMetrics 获取指标快照
func (eo *EnhancedORM) GetMetrics() *MetricsSnapshot {
	if eo.metricsCollector == nil {
		return nil
	}
	return eo.metricsCollector.GetMetrics()
}

// HealthCheck 执行健康检查
func (eo *EnhancedORM) HealthCheck(ctx context.Context) error {
	return eo.pool.HealthCheck(ctx)
}

// Close 关闭增强ORM
func (eo *EnhancedORM) Close() error {
	var errors []error

	// 停止指标收集
	if eo.metricsCollector != nil {
		eo.metricsCollector.Stop()
	}

	// 关闭连接池
	if eo.pool != nil {
		if err := eo.pool.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close pool: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing enhanced ORM: %v", errors)
	}

	config.Info("Enhanced ORM closed")
	return nil
}

// ============= 便捷的CRUD操作 =============

// Create 创建记录（使用主库）
func (eo *EnhancedORM) Create(ctx context.Context, value any) error {
	db, err := eo.GetWriteDB(ctx)
	if err != nil {
		return err
	}
	return db.Create(value).Error
}

// Find 查询记录（使用从库）
func (eo *EnhancedORM) Find(ctx context.Context, dest any, conds ...any) error {
	db, err := eo.GetReadDB(ctx)
	if err != nil {
		return err
	}
	return db.Find(dest, conds...).Error
}

// First 查询第一条记录（使用从库）
func (eo *EnhancedORM) First(ctx context.Context, dest any, conds ...any) error {
	db, err := eo.GetReadDB(ctx)
	if err != nil {
		return err
	}
	return db.First(dest, conds...).Error
}

// Update 更新记录（使用主库）
func (eo *EnhancedORM) Update(ctx context.Context, model any, column string, value any) error {
	db, err := eo.GetWriteDB(ctx)
	if err != nil {
		return err
	}
	return db.Model(model).Update(column, value).Error
}

// Updates 批量更新（使用主库）
func (eo *EnhancedORM) Updates(ctx context.Context, model any, values any) error {
	db, err := eo.GetWriteDB(ctx)
	if err != nil {
		return err
	}
	return db.Model(model).Updates(values).Error
}

// Delete 删除记录（使用主库）
func (eo *EnhancedORM) Delete(ctx context.Context, value any, conds ...any) error {
	db, err := eo.GetWriteDB(ctx)
	if err != nil {
		return err
	}
	return db.Delete(value, conds...).Error
}

// Count 计数（使用从库）
func (eo *EnhancedORM) Count(ctx context.Context, model any) (int64, error) {
	db, err := eo.GetReadDB(ctx)
	if err != nil {
		return 0, err
	}

	var count int64
	err = db.Model(model).Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在（使用从库）
func (eo *EnhancedORM) Exists(ctx context.Context, model any, conds ...any) (bool, error) {
	count, err := eo.Count(ctx, model)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ============= 全局增强ORM实例 =============

var (
	globalEnhancedORM *EnhancedORM
	enhancedOnce      sync.Once
	enhancedMutex     sync.Mutex
)

// GetGlobalEnhancedORM 获取全局增强ORM实例
func GetGlobalEnhancedORM() *EnhancedORM {
	enhancedOnce.Do(func() {
		enhancedMutex.Lock()
		defer enhancedMutex.Unlock()

		// 从配置管理器获取数据库配置
		dbConfig := DefaultDatabaseConfig()
		poolConfig := DefaultPoolConfig()

		var err error
		globalEnhancedORM, err = NewEnhancedORM(dbConfig, poolConfig)
		if err != nil {
			config.Fatalf("Failed to initialize global enhanced ORM: %v", err)
		}
	})
	return globalEnhancedORM
}

// SetGlobalEnhancedORM 设置全局增强ORM实例
func SetGlobalEnhancedORM(orm *EnhancedORM) {
	enhancedMutex.Lock()
	defer enhancedMutex.Unlock()

	// 关闭现有实例
	if globalEnhancedORM != nil {
		globalEnhancedORM.Close()
	}

	globalEnhancedORM = orm
}

// ============= 便捷函数（使用全局实例）=============

// CreateWithPool 使用连接池创建记录
func CreateWithPool(ctx context.Context, value any) error {
	return GetGlobalEnhancedORM().Create(ctx, value)
}

// FindWithPool 使用连接池查询记录
func FindWithPool(ctx context.Context, dest any, conds ...any) error {
	return GetGlobalEnhancedORM().Find(ctx, dest, conds...)
}

// FirstWithPool 使用连接池查询第一条记录
func FirstWithPool(ctx context.Context, dest any, conds ...any) error {
	return GetGlobalEnhancedORM().First(ctx, dest, conds...)
}

// UpdateWithPool 使用连接池更新记录
func UpdateWithPool(ctx context.Context, model any, column string, value any) error {
	return GetGlobalEnhancedORM().Update(ctx, model, column, value)
}

// DeleteWithPool 使用连接池删除记录
func DeleteWithPool(ctx context.Context, value any, conds ...any) error {
	return GetGlobalEnhancedORM().Delete(ctx, value, conds...)
}

// GetPoolMetrics 获取连接池指标
func GetPoolMetrics() *MetricsSnapshot {
	return GetGlobalEnhancedORM().GetMetrics()
}

// PrintPoolMetrics 打印连接池指标
func PrintPoolMetrics() {
	if metrics := GetPoolMetrics(); metrics != nil {
		metrics.PrintMetrics()
	}
}
