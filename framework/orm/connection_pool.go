// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// ReadWriteConfig 读写分离配置
type ReadWriteConfig struct {
	// 主库配置
	Master *DatabaseConfig `json:"master" yaml:"master"`
	// 从库配置列表
	Slaves []*DatabaseConfig `json:"slaves" yaml:"slaves"`
	// 负载均衡策略: round_robin, random, weight
	LoadBalanceStrategy string `json:"load_balance_strategy" yaml:"load_balance_strategy"`
	// 故障转移配置
	FailoverEnabled bool `json:"failover_enabled" yaml:"failover_enabled"`
	// 重试次数
	RetryAttempts int `json:"retry_attempts" yaml:"retry_attempts"`
	// 重试间隔
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
}

// DefaultReadWriteConfig 默认读写分离配置
func DefaultReadWriteConfig() *ReadWriteConfig {
	return &ReadWriteConfig{
		Master:              DefaultDatabaseConfig(),
		Slaves:              []*DatabaseConfig{},
		LoadBalanceStrategy: "round_robin",
		FailoverEnabled:     true,
		RetryAttempts:       3,
		RetryInterval:       time.Second,
	}
}

// ConnectionPoolManager 连接池管理器
type ConnectionPoolManager struct {
	// 主库连接池
	masterPool *gorm.DB
	// 从库连接池列表
	slavePools []*gorm.DB
	// 读写分离配置
	config *ReadWriteConfig
	// 连接池配置
	poolConfig *PoolConfig
	// 负载均衡器
	loadBalancer LoadBalancer
	// 指标收集器
	metricsCollector *MetricsCollector
	// 互斥锁
	mutex sync.RWMutex
	// 健康检查定时器
	healthCheckTicker *time.Ticker
	// 健康检查停止信号
	healthCheckStop chan struct{}
}

// NewConnectionPoolManager 创建连接池管理器
func NewConnectionPoolManager(rwConfig *ReadWriteConfig, poolConfig *PoolConfig) (*ConnectionPoolManager, error) {
	if rwConfig == nil {
		rwConfig = DefaultReadWriteConfig()
	}

	if poolConfig == nil {
		poolConfig = DefaultPoolConfig()
	}

	// 创建主库连接
	masterDB, err := createDBConnection(rwConfig.Master, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("创建主库连接失败: %w", err)
	}

	// 创建连接池管理器
	cpm := &ConnectionPoolManager{
		masterPool:       masterDB,
		slavePools:       make([]*gorm.DB, 0, len(rwConfig.Slaves)),
		config:           rwConfig,
		poolConfig:       poolConfig,
		metricsCollector: NewMetricsCollector(),
		healthCheckStop:  make(chan struct{}),
	}

	// 创建从库连接
	for i, slaveConfig := range rwConfig.Slaves {
		slaveDB, err := createDBConnection(slaveConfig, poolConfig)
		if err != nil {
			config.Warnf("创建从库%d连接失败: %v", i+1, err)
			continue
		}
		cpm.slavePools = append(cpm.slavePools, slaveDB)
	}

	// 创建负载均衡器
	cpm.loadBalancer = createLoadBalancer(rwConfig.LoadBalanceStrategy)

	// 启动指标收集
	cpm.metricsCollector.Start()

	// 启动健康检查
	if poolConfig.HealthCheckEnabled {
		cpm.startHealthCheck(poolConfig.HealthCheckInterval)
	}

	config.Infof("连接池管理器初始化成功，主库: 1, 从库: %d", len(cpm.slavePools))
	return cpm, nil
}

// createDBConnection 创建数据库连接
func createDBConnection(dbConfig *DatabaseConfig, poolConfig *PoolConfig) (*gorm.DB, error) {
	// 创建GORM配置
	gormConfig := &gorm.Config{
		Logger: newGormLogger(dbConfig.LogLevel, time.Duration(dbConfig.SlowQuery)*time.Millisecond),
	}

	// 根据数据库类型创建连接
	var db *gorm.DB
	var err error

	switch dbConfig.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port,
			dbConfig.Database, dbConfig.Charset, dbConfig.Timezone)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
			dbConfig.Host, dbConfig.Username, dbConfig.Password, dbConfig.Database, dbConfig.Port, dbConfig.Timezone)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dbConfig.Database), gormConfig)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbConfig.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
		sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)
	}

	return db, nil
}

// GetMaster 获取主库连接
func (cpm *ConnectionPoolManager) GetMaster() *gorm.DB {
	cpm.mutex.RLock()
	defer cpm.mutex.RUnlock()

	// 记录指标
	cpm.metricsCollector.RecordConnection("master", "master")

	return cpm.masterPool
}

// GetSlave 获取从库连接
func (cpm *ConnectionPoolManager) GetSlave() *gorm.DB {
	cpm.mutex.RLock()
	defer cpm.mutex.RUnlock()

	// 如果没有从库，返回主库
	if len(cpm.slavePools) == 0 {
		cpm.metricsCollector.RecordConnection("master", "master")
		return cpm.masterPool
	}

	// 使用负载均衡器选择从库
	index := cpm.loadBalancer.Next(len(cpm.slavePools))

	// 记录指标
	cpm.metricsCollector.RecordConnection(fmt.Sprintf("slave-%d", index), "slave")

	return cpm.slavePools[index]
}

// GetReadDB 获取读库连接（优先从库）
func (cpm *ConnectionPoolManager) GetReadDB() *gorm.DB {
	return cpm.GetSlave()
}

// GetWriteDB 获取写库连接（主库）
func (cpm *ConnectionPoolManager) GetWriteDB() *gorm.DB {
	return cpm.GetMaster()
}

// WithContext 使用上下文
func (cpm *ConnectionPoolManager) WithContext(ctx context.Context) *ConnectionPoolManager {
	return cpm
}

// Transaction 执行事务（使用主库）
func (cpm *ConnectionPoolManager) Transaction(fn func(tx *gorm.DB) error) error {
	return cpm.GetMaster().Transaction(fn)
}

// TransactionWithContext 使用上下文执行事务
func (cpm *ConnectionPoolManager) TransactionWithContext(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return cpm.GetMaster().WithContext(ctx).Transaction(fn)
}

// Close 关闭连接池
func (cpm *ConnectionPoolManager) Close() error {
	cpm.mutex.Lock()
	defer cpm.mutex.Unlock()

	// 停止健康检查
	if cpm.healthCheckTicker != nil {
		cpm.healthCheckTicker.Stop()
		close(cpm.healthCheckStop)
	}

	// 停止指标收集
	if cpm.metricsCollector != nil {
		cpm.metricsCollector.Stop()
	}

	// 关闭主库连接
	if sqlDB, err := cpm.masterPool.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			config.Errorf("关闭主库连接失败: %v", err)
		}
	}

	// 关闭从库连接
	for i, slave := range cpm.slavePools {
		if sqlDB, err := slave.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				config.Errorf("关闭从库%d连接失败: %v", i+1, err)
			}
		}
	}

	config.Info("连接池管理器已关闭")
	return nil
}

// startHealthCheck 启动健康检查
func (cpm *ConnectionPoolManager) startHealthCheck(interval time.Duration) {
	cpm.healthCheckTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-cpm.healthCheckTicker.C:
				cpm.performHealthCheck()
			case <-cpm.healthCheckStop:
				return
			}
		}
	}()

	config.Info("连接池健康检查已启动")
}

// performHealthCheck 执行健康检查
func (cpm *ConnectionPoolManager) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), cpm.poolConfig.HealthCheckTimeout)
	defer cancel()

	// 检查主库
	if err := cpm.checkConnection(ctx, cpm.masterPool, "master"); err != nil {
		config.Errorf("主库健康检查失败: %v", err)
	}

	// 检查从库
	for i, slave := range cpm.slavePools {
		if err := cpm.checkConnection(ctx, slave, fmt.Sprintf("slave-%d", i)); err != nil {
			config.Errorf("从库%d健康检查失败: %v", i, err)
		}
	}
}

// checkConnection 检查连接
func (cpm *ConnectionPoolManager) checkConnection(ctx context.Context, db *gorm.DB, name string) error {
	var result int
	err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error

	if err != nil {
		cpm.metricsCollector.RecordConnectionError(name)
		return err
	}

	return nil
}

// GetStats 获取连接池统计信息
func (cpm *ConnectionPoolManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取主库统计信息
	if sqlDB, err := cpm.masterPool.DB(); err == nil {
		masterStats := sqlDB.Stats()
		stats["master"] = map[string]interface{}{
			"max_open_connections": masterStats.MaxOpenConnections,
			"open_connections":     masterStats.OpenConnections,
			"in_use":               masterStats.InUse,
			"idle":                 masterStats.Idle,
			"wait_count":           masterStats.WaitCount,
			"wait_duration":        masterStats.WaitDuration.String(),
			"max_idle_closed":      masterStats.MaxIdleClosed,
			"max_lifetime_closed":  masterStats.MaxLifetimeClosed,
		}
	}

	// 获取从库统计信息
	slaves := make([]map[string]interface{}, 0, len(cpm.slavePools))
	for i, slave := range cpm.slavePools {
		if sqlDB, err := slave.DB(); err == nil {
			slaveStats := sqlDB.Stats()
			slaves = append(slaves, map[string]interface{}{
				"id":                   i,
				"max_open_connections": slaveStats.MaxOpenConnections,
				"open_connections":     slaveStats.OpenConnections,
				"in_use":               slaveStats.InUse,
				"idle":                 slaveStats.Idle,
				"wait_count":           slaveStats.WaitCount,
				"wait_duration":        slaveStats.WaitDuration.String(),
				"max_idle_closed":      slaveStats.MaxIdleClosed,
				"max_lifetime_closed":  slaveStats.MaxLifetimeClosed,
			})
		}
	}
	stats["slaves"] = slaves

	// 添加指标收集器的统计信息
	if metrics := cpm.metricsCollector.GetMetrics(); metrics != nil {
		stats["metrics"] = map[string]interface{}{
			"total_connections":  metrics.TotalConnections,
			"active_connections": metrics.ActiveConnections,
			"connection_errors":  metrics.ConnectionErrors,
			"total_queries":      metrics.TotalQueries,
			"slow_queries":       metrics.SlowQueries,
			"failed_queries":     metrics.FailedQueries,
			"average_query_time": metrics.AverageQueryTime.String(),
		}
	}

	return stats
}

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// Next 获取下一个节点索引
	Next(n int) int
}

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	counter int64
	mutex   sync.Mutex
}

// Next 获取下一个节点索引
func (rb *RoundRobinBalancer) Next(n int) int {
	if n <= 0 {
		return 0
	}

	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.counter++
	return int(rb.counter % int64(n))
}

// RandomBalancer 随机负载均衡器
type RandomBalancer struct{}

// Next 获取下一个节点索引
func (rb *RandomBalancer) Next(n int) int {
	if n <= 0 {
		return 0
	}

	return int(time.Now().UnixNano() % int64(n))
}

// WeightedBalancer 加权负载均衡器
type WeightedBalancer struct {
	weights []int
	total   int
	current int
	mutex   sync.Mutex
}

// NewWeightedBalancer 创建加权负载均衡器
func NewWeightedBalancer(weights []int) *WeightedBalancer {
	total := 0
	for _, w := range weights {
		total += w
	}

	return &WeightedBalancer{
		weights: weights,
		total:   total,
	}
}

// Next 获取下一个节点索引
func (wb *WeightedBalancer) Next(n int) int {
	if n <= 0 || len(wb.weights) == 0 {
		return 0
	}

	wb.mutex.Lock()
	defer wb.mutex.Unlock()

	// 如果权重数组长度小于n，使用轮询
	if len(wb.weights) < n {
		wb.current = (wb.current + 1) % n
		return wb.current
	}

	// 使用权重选择
	wb.current = (wb.current + 1) % wb.total

	// 根据权重选择节点
	sum := 0
	for i, w := range wb.weights[:n] {
		sum += w
		if wb.current < sum {
			return i
		}
	}

	return 0
}

// createLoadBalancer 创建负载均衡器
func createLoadBalancer(strategy string) LoadBalancer {
	switch strategy {
	case "random":
		return &RandomBalancer{}
	case "weight":
		return NewWeightedBalancer([]int{1, 1}) // 默认权重
	default:
		return &RoundRobinBalancer{}
	}
}

// 全局连接池管理器
var (
	globalPoolManager *ConnectionPoolManager
	poolManagerOnce   sync.Once
)

// GetGlobalConnectionPoolManager 获取全局连接池管理器
func GetGlobalConnectionPoolManager() *ConnectionPoolManager {
	poolManagerOnce.Do(func() {
		// 从配置中获取读写分离配置
		rwConfig := DefaultReadWriteConfig()
		poolConfig := DefaultPoolConfig()

		// 尝试从全局配置获取
		if configManager := config.GetDatabaseConfigManager(); configManager != nil {
			if appConfig, err := configManager.GetConfig(); err == nil {
				// 配置主库
				if appConfig.Primary.Driver != "" {
					rwConfig.Master.Type = appConfig.Primary.Driver
				}
				if appConfig.Primary.Host != "" {
					rwConfig.Master.Host = appConfig.Primary.Host
				}
				if appConfig.Primary.Port > 0 {
					rwConfig.Master.Port = appConfig.Primary.Port
				}
				if appConfig.Primary.Username != "" {
					rwConfig.Master.Username = appConfig.Primary.Username
				}
				if appConfig.Primary.Password != "" {
					rwConfig.Master.Password = appConfig.Primary.Password
				}
				if appConfig.Primary.Database != "" {
					rwConfig.Master.Database = appConfig.Primary.Database
				}

				// 配置从库（如果有）
				// 注意：这里需要根据实际配置结构调整
				// 暂时使用简单配置，后续可以扩展
				// 这里假设只有一个从库，与主库配置相同但使用不同端口
				slaveConfig := DefaultDatabaseConfig()
				slaveConfig.Type = rwConfig.Master.Type
				slaveConfig.Host = rwConfig.Master.Host
				slaveConfig.Port = rwConfig.Master.Port + 1 // 默认从库端口为主库端口+1
				slaveConfig.Username = rwConfig.Master.Username
				slaveConfig.Password = rwConfig.Master.Password
				slaveConfig.Database = rwConfig.Master.Database

				// 只有在主库配置有效时才添加从库
				if rwConfig.Master.Host != "" && rwConfig.Master.Port > 0 {
					rwConfig.Slaves = append(rwConfig.Slaves, slaveConfig)
				}
			}
		}

		var err error
		globalPoolManager, err = NewConnectionPoolManager(rwConfig, poolConfig)
		if err != nil {
			config.Fatalf("初始化全局连接池管理器失败: %v", err)
		}
	})

	return globalPoolManager
}

// SetGlobalConnectionPoolManager 设置全局连接池管理器
func SetGlobalConnectionPoolManager(cpm *ConnectionPoolManager) {
	if globalPoolManager != nil {
		globalPoolManager.Close()
	}

	globalPoolManager = cpm
}

// ============= 便捷函数 =============

// GetMasterDB 获取主库连接
func GetMasterDB() *gorm.DB {
	return GetGlobalConnectionPoolManager().GetMaster()
}

// GetSlaveDB 获取从库连接
func GetSlaveDB() *gorm.DB {
	return GetGlobalConnectionPoolManager().GetSlave()
}

// GetReadDB 获取读库连接
func GetReadDB() *gorm.DB {
	return GetGlobalConnectionPoolManager().GetReadDB()
}

// GetWriteDB 获取写库连接
func GetWriteDB() *gorm.DB {
	return GetGlobalConnectionPoolManager().GetWriteDB()
}

// TransactionWithPool 使用连接池执行事务
func TransactionWithPool(fn func(tx *gorm.DB) error) error {
	return GetGlobalConnectionPoolManager().Transaction(fn)
}
