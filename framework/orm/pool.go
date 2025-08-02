package orm

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	globalConfig "github.com/zsy619/yyhertz/framework/config"
)

// ConnectionPool 连接池接口
type ConnectionPool interface {
	// 获取连接
	GetConnection(ctx context.Context) (*gorm.DB, error)
	// 获取主库连接
	GetMasterConnection(ctx context.Context) (*gorm.DB, error)
	// 获取从库连接
	GetSlaveConnection(ctx context.Context) (*gorm.DB, error)
	// 释放连接
	ReleaseConnection(db *gorm.DB)
	// 关闭连接池
	Close() error
	// 获取连接池统计信息
	Stats() *PoolStats
	// 健康检查
	HealthCheck(ctx context.Context) error
}

// PoolConfig 连接池配置
type PoolConfig struct {
	// 基础配置
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`       // 最大空闲连接数
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`       // 最大打开连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"` // 连接最大生存时间
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"` // 连接最大空闲时间
	
	// 高级配置
	EnableLoadBalance bool          `json:"enable_load_balance" yaml:"enable_load_balance"` // 启用负载均衡
	LoadBalanceMethod string        `json:"load_balance_method" yaml:"load_balance_method"` // 负载均衡方法: round_robin, random, weight
	FailoverEnabled   bool          `json:"failover_enabled" yaml:"failover_enabled"`       // 启用故障转移
	RetryAttempts     int           `json:"retry_attempts" yaml:"retry_attempts"`           // 重试次数
	RetryInterval     time.Duration `json:"retry_interval" yaml:"retry_interval"`           // 重试间隔
	
	// 健康检查
	HealthCheckEnabled  bool          `json:"health_check_enabled" yaml:"health_check_enabled"`   // 启用健康检查
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"` // 健康检查间隔
	HealthCheckTimeout  time.Duration `json:"health_check_timeout" yaml:"health_check_timeout"`   // 健康检查超时
	
	// 监控配置
	MetricsEnabled     bool          `json:"metrics_enabled" yaml:"metrics_enabled"`         // 启用指标收集
	SlowQueryThreshold time.Duration `json:"slow_query_threshold" yaml:"slow_query_threshold"` // 慢查询阈值
}

// DefaultPoolConfig 默认连接池配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxIdleConns:        10,
		MaxOpenConns:        100,
		ConnMaxLifetime:     time.Hour,
		ConnMaxIdleTime:     time.Minute * 30,
		EnableLoadBalance:   false,
		LoadBalanceMethod:   "round_robin",
		FailoverEnabled:     true,
		RetryAttempts:       3,
		RetryInterval:       time.Second,
		HealthCheckEnabled:  true,
		HealthCheckInterval: time.Minute * 5,
		HealthCheckTimeout:  time.Second * 30,
		MetricsEnabled:      true,
		SlowQueryThreshold:  time.Millisecond * 500,
	}
}

// PoolStats 连接池统计信息
type PoolStats struct {
	MaxOpenConnections int `json:"max_open_connections"` // 最大打开连接数
	OpenConnections    int `json:"open_connections"`     // 当前打开连接数
	InUse              int `json:"in_use"`               // 正在使用的连接数
	Idle               int `json:"idle"`                 // 空闲连接数
	
	// 扩展统计
	TotalConnections    int64         `json:"total_connections"`     // 总连接数
	TotalQueries        int64         `json:"total_queries"`         // 总查询数
	SlowQueries         int64         `json:"slow_queries"`          // 慢查询数
	FailedConnections   int64         `json:"failed_connections"`    // 失败连接数
	AverageResponseTime time.Duration `json:"average_response_time"` // 平均响应时间
	
	// 健康状态
	HealthyNodes   int               `json:"healthy_nodes"`   // 健康节点数
	UnhealthyNodes int               `json:"unhealthy_nodes"` // 不健康节点数
	LastHealthCheck time.Time        `json:"last_health_check"` // 最后健康检查时间
}

// DatabaseNode 数据库节点
type DatabaseNode struct {
	ID       string          `json:"id"`       // 节点ID
	Config   *DatabaseConfig `json:"config"`   // 数据库配置
	DB       *gorm.DB        `json:"-"`        // 数据库连接
	Weight   int             `json:"weight"`   // 权重(用于负载均衡)
	IsMaster bool            `json:"is_master"` // 是否为主库
	IsHealthy bool           `json:"is_healthy"` // 是否健康
	
	// 统计信息
	ConnectionCount int64     `json:"connection_count"` // 连接数
	QueryCount      int64     `json:"query_count"`      // 查询数
	ErrorCount      int64     `json:"error_count"`      // 错误数
	LastUsed        time.Time `json:"last_used"`        // 最后使用时间
	LastHealthCheck time.Time `json:"last_health_check"` // 最后健康检查时间
}

// MultiNodePool 多节点连接池
type MultiNodePool struct {
	config    *PoolConfig
	nodes     []*DatabaseNode
	masterNodes []*DatabaseNode
	slaveNodes  []*DatabaseNode
	
	// 负载均衡
	rrIndex    int64 // round robin 索引
	randSource *rand.Rand
	
	// 状态管理
	closed    bool
	mutex     sync.RWMutex
	stats     *PoolStats
	statsMutex sync.RWMutex
	
	// 健康检查
	healthCheckCancel context.CancelFunc
	
	// 监控
	metricsCollector *MetricsCollector
}

// NewMultiNodePool 创建多节点连接池
func NewMultiNodePool(config *PoolConfig, nodes []*DatabaseNode) (*MultiNodePool, error) {
	if config == nil {
		config = DefaultPoolConfig()
	}
	
	if len(nodes) == 0 {
		return nil, errors.New("至少需要一个数据库节点")
	}
	
	pool := &MultiNodePool{
		config:     config,
		nodes:      make([]*DatabaseNode, 0, len(nodes)),
		masterNodes: make([]*DatabaseNode, 0),
		slaveNodes:  make([]*DatabaseNode, 0),
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
		stats:      &PoolStats{},
	}
	
	// 初始化节点
	for _, node := range nodes {
		if err := pool.addNode(node); err != nil {
			globalConfig.Errorf("Failed to add database node %s: %v", node.ID, err)
			continue
		}
	}
	
	if len(pool.nodes) == 0 {
		return nil, errors.New("没有可用的数据库节点")
	}
	
	// 初始化监控
	if config.MetricsEnabled {
		pool.metricsCollector = NewMetricsCollector()
	}
	
	// 启动健康检查
	if config.HealthCheckEnabled {
		pool.startHealthCheck()
	}
	
	globalConfig.Infof("Multi-node connection pool initialized with %d nodes", len(pool.nodes))
	return pool, nil
}

// addNode 添加节点
func (p *MultiNodePool) addNode(node *DatabaseNode) error {
	// 创建数据库连接
	db, err := p.createConnection(node.Config)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}
	
	node.DB = db
	node.IsHealthy = true
	node.LastHealthCheck = time.Now()
	
	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(p.config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(p.config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(p.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(p.config.ConnMaxIdleTime)
	
	// 添加到节点列表
	p.nodes = append(p.nodes, node)
	
	// 分类节点
	if node.IsMaster {
		p.masterNodes = append(p.masterNodes, node)
	} else {
		p.slaveNodes = append(p.slaveNodes, node)
	}
	
	globalConfig.Infof("Added database node: %s (master: %t)", node.ID, node.IsMaster)
	return nil
}

// createConnection 创建数据库连接
func (p *MultiNodePool) createConnection(dbConfig *DatabaseConfig) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector
	
	switch dbConfig.Type {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port,
			dbConfig.Database, dbConfig.Charset, dbConfig.Timezone)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
			dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password,
			dbConfig.Database, dbConfig.Timezone)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dbConfig.Database)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}
	
	// GORM配置
	gormConfig := &gorm.Config{
		// Logger: &gormLogWriter{},
	}
	
	return gorm.Open(dialector, gormConfig)
}

// GetConnection 获取连接
func (p *MultiNodePool) GetConnection(ctx context.Context) (*gorm.DB, error) {
	if p.closed {
		return nil, errors.New("connection pool is closed")
	}
	
	// 优先从主库获取连接
	if len(p.masterNodes) > 0 {
		return p.GetMasterConnection(ctx)
	}
	
	// 从从库获取连接
	return p.GetSlaveConnection(ctx)
}

// GetMasterConnection 获取主库连接
func (p *MultiNodePool) GetMasterConnection(ctx context.Context) (*gorm.DB, error) {
	return p.getConnectionFromNodes(ctx, p.masterNodes, "master")
}

// GetSlaveConnection 获取从库连接
func (p *MultiNodePool) GetSlaveConnection(ctx context.Context) (*gorm.DB, error) {
	// 如果没有从库，使用主库
	if len(p.slaveNodes) == 0 {
		return p.GetMasterConnection(ctx)
	}
	
	return p.getConnectionFromNodes(ctx, p.slaveNodes, "slave")
}

// getConnectionFromNodes 从节点列表获取连接
func (p *MultiNodePool) getConnectionFromNodes(ctx context.Context, nodes []*DatabaseNode, nodeType string) (*gorm.DB, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available %s nodes", nodeType)
	}
	
	var selectedNode *DatabaseNode
	
	// 根据负载均衡策略选择节点
	switch p.config.LoadBalanceMethod {
	case "round_robin":
		selectedNode = p.selectNodeRoundRobin(nodes)
	case "random":
		selectedNode = p.selectNodeRandom(nodes)
	case "weight":
		selectedNode = p.selectNodeByWeight(nodes)
	default:
		selectedNode = p.selectNodeRoundRobin(nodes)
	}
	
	if selectedNode == nil || !selectedNode.IsHealthy {
		return nil, fmt.Errorf("no healthy %s nodes available", nodeType)
	}
	
	// 更新统计信息
	atomic.AddInt64(&selectedNode.ConnectionCount, 1)
	atomic.AddInt64(&selectedNode.QueryCount, 1)
	selectedNode.LastUsed = time.Now()
	
	// 记录指标
	if p.metricsCollector != nil {
		p.metricsCollector.RecordConnection(selectedNode.ID, nodeType)
	}
	
	return selectedNode.DB.WithContext(ctx), nil
}

// selectNodeRoundRobin 轮询选择节点
func (p *MultiNodePool) selectNodeRoundRobin(nodes []*DatabaseNode) *DatabaseNode {
	if len(nodes) == 0 {
		return nil
	}
	
	index := atomic.AddInt64(&p.rrIndex, 1) % int64(len(nodes))
	return nodes[index]
}

// selectNodeRandom 随机选择节点
func (p *MultiNodePool) selectNodeRandom(nodes []*DatabaseNode) *DatabaseNode {
	if len(nodes) == 0 {
		return nil
	}
	
	p.mutex.RLock()
	index := p.randSource.Intn(len(nodes))
	p.mutex.RUnlock()
	
	return nodes[index]
}

// selectNodeByWeight 按权重选择节点
func (p *MultiNodePool) selectNodeByWeight(nodes []*DatabaseNode) *DatabaseNode {
	if len(nodes) == 0 {
		return nil
	}
	
	// 计算总权重
	totalWeight := 0
	for _, node := range nodes {
		if node.IsHealthy {
			totalWeight += node.Weight
		}
	}
	
	if totalWeight == 0 {
		return p.selectNodeRandom(nodes)
	}
	
	// 按权重随机选择
	p.mutex.RLock()
	randomWeight := p.randSource.Intn(totalWeight)
	p.mutex.RUnlock()
	
	currentWeight := 0
	for _, node := range nodes {
		if node.IsHealthy {
			currentWeight += node.Weight
			if currentWeight > randomWeight {
				return node
			}
		}
	}
	
	return nodes[0] // 兜底
}

// ReleaseConnection 释放连接
func (p *MultiNodePool) ReleaseConnection(db *gorm.DB) {
	// 在GORM中，连接是自动管理的，这里主要是更新统计信息
	if p.metricsCollector != nil {
		p.metricsCollector.RecordConnectionRelease()
	}
}

// HealthCheck 健康检查
func (p *MultiNodePool) HealthCheck(ctx context.Context) error {
	p.mutex.RLock()
	nodes := make([]*DatabaseNode, len(p.nodes))
	copy(nodes, p.nodes)
	p.mutex.RUnlock()
	
	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(n *DatabaseNode) {
			defer wg.Done()
			p.checkNodeHealth(ctx, n)
		}(node)
	}
	
	wg.Wait()
	
	// 更新统计信息
	p.updateHealthStats()
	
	return nil
}

// checkNodeHealth 检查单个节点健康状态
func (p *MultiNodePool) checkNodeHealth(ctx context.Context, node *DatabaseNode) {
	timeoutCtx, cancel := context.WithTimeout(ctx, p.config.HealthCheckTimeout)
	defer cancel()
	
	// 执行健康检查查询
	var result int
	err := node.DB.WithContext(timeoutCtx).Raw("SELECT 1").Scan(&result).Error
	
	wasHealthy := node.IsHealthy
	node.IsHealthy = (err == nil && result == 1)
	node.LastHealthCheck = time.Now()
	
	if wasHealthy != node.IsHealthy {
		if node.IsHealthy {
			globalConfig.Infof("Database node %s is now healthy", node.ID)
		} else {
			globalConfig.Errorf("Database node %s is unhealthy: %v", node.ID, err)
			atomic.AddInt64(&node.ErrorCount, 1)
		}
	}
}

// startHealthCheck 启动健康检查
func (p *MultiNodePool) startHealthCheck() {
	ctx, cancel := context.WithCancel(context.Background())
	p.healthCheckCancel = cancel
	
	go func() {
		ticker := time.NewTicker(p.config.HealthCheckInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := p.HealthCheck(ctx); err != nil {
					globalConfig.Errorf("Health check failed: %v", err)
				}
			}
		}
	}()
	
	globalConfig.Info("Health check started")
}

// updateHealthStats 更新健康统计信息
func (p *MultiNodePool) updateHealthStats() {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()
	
	healthyCount := 0
	unhealthyCount := 0
	
	for _, node := range p.nodes {
		if node.IsHealthy {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}
	
	p.stats.HealthyNodes = healthyCount
	p.stats.UnhealthyNodes = unhealthyCount
	p.stats.LastHealthCheck = time.Now()
}

// Stats 获取连接池统计信息
func (p *MultiNodePool) Stats() *PoolStats {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()
	
	// 收集所有节点的统计信息
	var totalConnections, totalQueries, slowQueries, failedConnections int64
	var totalResponseTime time.Duration
	
	for _, node := range p.nodes {
		totalConnections += atomic.LoadInt64(&node.ConnectionCount)
		totalQueries += atomic.LoadInt64(&node.QueryCount)
		failedConnections += atomic.LoadInt64(&node.ErrorCount)
	}
	
	// 计算平均响应时间
	if totalQueries > 0 {
		totalResponseTime = totalResponseTime / time.Duration(totalQueries)
	}
	
	// 获取第一个健康节点的连接池统计信息
	for _, node := range p.nodes {
		if node.IsHealthy {
			if sqlDB, err := node.DB.DB(); err == nil {
				stats := sqlDB.Stats()
				p.stats.MaxOpenConnections = stats.MaxOpenConnections
				p.stats.OpenConnections = stats.OpenConnections
				p.stats.InUse = stats.InUse
				p.stats.Idle = stats.Idle
				break
			}
		}
	}
	
	p.stats.TotalConnections = totalConnections
	p.stats.TotalQueries = totalQueries
	p.stats.SlowQueries = slowQueries
	p.stats.FailedConnections = failedConnections
	p.stats.AverageResponseTime = totalResponseTime
	
	// 复制统计信息返回
	statsCopy := *p.stats
	return &statsCopy
}

// Close 关闭连接池
func (p *MultiNodePool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.closed {
		return nil
	}
	
	p.closed = true
	
	// 停止健康检查
	if p.healthCheckCancel != nil {
		p.healthCheckCancel()
	}
	
	// 关闭所有连接
	var errors []error
	for _, node := range p.nodes {
		if sqlDB, err := node.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close node %s: %w", node.ID, err))
			}
		}
	}
	
	// 关闭监控
	if p.metricsCollector != nil {
		p.metricsCollector.Close()
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errors closing pool: %v", errors)
	}
	
	globalConfig.Info("Multi-node connection pool closed")
	return nil
}

// ============= 便捷函数 =============

// CreateMasterSlavePool 创建主从连接池
func CreateMasterSlavePool(masterConfig, slaveConfig *DatabaseConfig) (*MultiNodePool, error) {
	nodes := []*DatabaseNode{
		{
			ID:       "master",
			Config:   masterConfig,
			Weight:   10,
			IsMaster: true,
		},
		{
			ID:       "slave",
			Config:   slaveConfig,
			Weight:   5,
			IsMaster: false,
		},
	}
	
	poolConfig := DefaultPoolConfig()
	poolConfig.EnableLoadBalance = true
	poolConfig.LoadBalanceMethod = "weight"
	
	return NewMultiNodePool(poolConfig, nodes)
}

// CreateClusterPool 创建集群连接池
func CreateClusterPool(configs []*DatabaseConfig) (*MultiNodePool, error) {
	nodes := make([]*DatabaseNode, len(configs))
	
	for i, cfg := range configs {
		nodes[i] = &DatabaseNode{
			ID:       fmt.Sprintf("node-%d", i+1),
			Config:   cfg,
			Weight:   1,
			IsMaster: i == 0, // 第一个节点作为主库
		}
	}
	
	poolConfig := DefaultPoolConfig()
	poolConfig.EnableLoadBalance = true
	poolConfig.LoadBalanceMethod = "round_robin"
	
	return NewMultiNodePool(poolConfig, nodes)
}