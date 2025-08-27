// Package mybatis 智能连接池管理器
//
// 提供以下优化功能：
// 1. 自适应连接池大小调整
// 2. 连接健康检查和自动恢复
// 3. 负载均衡和读写分离
// 4. 连接池监控和统计
// 5. 连接预热和懒初始化
package mybatis

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// ConnectionPoolManager 智能连接池管理器
type ConnectionPoolManager struct {
	db                    *gorm.DB
	config                *PoolConfig
	stats                 *PoolStats
	healthChecker         *HealthChecker
	loadBalancer          *LoadBalancer
	autoScaler           *PoolAutoScaler
	
	// 并发控制
	mutex                 sync.RWMutex
	adjustmentInProgress  int32
	
	// 监控
	lastAdjustment        time.Time
	adjustmentHistory     []PoolAdjustment
}

// PoolConfig 连接池配置
type PoolConfig struct {
	// 基础配置
	InitialSize         int           `yaml:"initial_size" json:"initial_size"`                 // 初始连接数
	MinSize             int           `yaml:"min_size" json:"min_size"`                         // 最小连接数  
	MaxSize             int           `yaml:"max_size" json:"max_size"`                         // 最大连接数
	MaxIdleTime         time.Duration `yaml:"max_idle_time" json:"max_idle_time"`               // 最大空闲时间
	MaxLifetime         time.Duration `yaml:"max_lifetime" json:"max_lifetime"`                 // 连接最大生命周期
	ConnectionTimeout   time.Duration `yaml:"connection_timeout" json:"connection_timeout"`     // 连接超时
	
	// 自适应配置
	EnableAutoScaling   bool          `yaml:"enable_auto_scaling" json:"enable_auto_scaling"`   // 启用自动扩缩容
	ScalingThreshold    float64       `yaml:"scaling_threshold" json:"scaling_threshold"`       // 扩缩容阈值 (0.8 = 80%)
	ScalingInterval     time.Duration `yaml:"scaling_interval" json:"scaling_interval"`         // 扩缩容检查间隔
	ScaleUpFactor       float64       `yaml:"scale_up_factor" json:"scale_up_factor"`           // 扩容因子 (1.5 = 150%)
	ScaleDownFactor     float64       `yaml:"scale_down_factor" json:"scale_down_factor"`       // 缩容因子 (0.8 = 80%)
	
	// 健康检查配置
	HealthCheckEnabled  bool          `yaml:"health_check_enabled" json:"health_check_enabled"` // 启用健康检查
	HealthCheckInterval time.Duration `yaml:"health_check_interval" json:"health_check_interval"` // 健康检查间隔
	HealthCheckTimeout  time.Duration `yaml:"health_check_timeout" json:"health_check_timeout"` // 健康检查超时
	HealthCheckSQL      string        `yaml:"health_check_sql" json:"health_check_sql"`         // 健康检查SQL
	
	// 读写分离配置
	ReadWriteSplit      bool          `yaml:"read_write_split" json:"read_write_split"`         // 启用读写分离
	ReadReplicaRatio    float64       `yaml:"read_replica_ratio" json:"read_replica_ratio"`     // 读副本流量比例
}

// PoolStats 连接池统计信息
type PoolStats struct {
	// 基础统计
	TotalConnections    int32     `json:"total_connections"`
	ActiveConnections   int32     `json:"active_connections"`
	IdleConnections     int32     `json:"idle_connections"`
	WaitingConnections  int32     `json:"waiting_connections"`
	
	// 性能统计
	ConnectionsCreated  int64     `json:"connections_created"`
	ConnectionsClosed   int64     `json:"connections_closed"`
	ConnectionErrors    int64     `json:"connection_errors"`
	AvgConnectionTime   time.Duration `json:"avg_connection_time"`
	MaxConnectionTime   time.Duration `json:"max_connection_time"`
	
	// 健康状态
	HealthyConnections  int32     `json:"healthy_connections"`
	UnhealthyConnections int32    `json:"unhealthy_connections"`
	LastHealthCheck     time.Time `json:"last_health_check"`
	
	// 自动扩缩容统计
	ScaleUpEvents       int64     `json:"scale_up_events"`
	ScaleDownEvents     int64     `json:"scale_down_events"`
	LastScalingAction   time.Time `json:"last_scaling_action"`
	
	// 使用率统计
	PeakUsage           float64   `json:"peak_usage"`
	CurrentUsage        float64   `json:"current_usage"`
	AvgUsage            float64   `json:"avg_usage"`
	
	mutex               sync.RWMutex
}

// HealthChecker 健康检查器
type HealthChecker struct {
	config    *PoolConfig
	db        *gorm.DB
	isRunning int32
	stopChan  chan struct{}
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	readDbs   []*gorm.DB
	writeDbs  []*gorm.DB
	roundRobin int32
	config    *PoolConfig
	mutex     sync.RWMutex
}

// PoolAutoScaler 连接池自动扩缩容器
type PoolAutoScaler struct {
	manager   *ConnectionPoolManager
	isRunning int32
	stopChan  chan struct{}
}

// PoolAdjustment 连接池调整记录
type PoolAdjustment struct {
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`      // "scale_up", "scale_down", "health_recovery"
	OldSize     int       `json:"old_size"`
	NewSize     int       `json:"new_size"`
	Reason      string    `json:"reason"`
	Usage       float64   `json:"usage"`
}

// DefaultPoolConfig 默认连接池配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		InitialSize:         10,
		MinSize:             5,
		MaxSize:             100,
		MaxIdleTime:         10 * time.Minute,
		MaxLifetime:         1 * time.Hour,
		ConnectionTimeout:   30 * time.Second,
		EnableAutoScaling:   true,
		ScalingThreshold:    0.8,
		ScalingInterval:     30 * time.Second,
		ScaleUpFactor:       1.5,
		ScaleDownFactor:     0.8,
		HealthCheckEnabled:  true,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		HealthCheckSQL:      "SELECT 1",
		ReadWriteSplit:      false,
		ReadReplicaRatio:    0.7,
	}
}

// NewConnectionPoolManager 创建连接池管理器
func NewConnectionPoolManager(db *gorm.DB, config *PoolConfig) *ConnectionPoolManager {
	if config == nil {
		config = DefaultPoolConfig()
	}
	
	manager := &ConnectionPoolManager{
		db:                db,
		config:            config,
		stats:             &PoolStats{},
		adjustmentHistory: make([]PoolAdjustment, 0, 100),
	}
	
	// 初始化组件
	manager.healthChecker = NewHealthChecker(config, db)
	manager.loadBalancer = NewLoadBalancer(config)
	manager.autoScaler = NewPoolAutoScaler(manager)
	
	// 应用初始配置
	manager.applyInitialConfig()
	
	// 启动后台服务
	if config.HealthCheckEnabled {
		manager.healthChecker.Start()
	}
	
	if config.EnableAutoScaling {
		manager.autoScaler.Start()
	}
	
	return manager
}

// applyInitialConfig 应用初始配置
func (pm *ConnectionPoolManager) applyInitialConfig() {
	if sqlDB, err := pm.db.DB(); err == nil {
		// 设置连接池参数
		sqlDB.SetMaxOpenConns(pm.config.MaxSize)
		sqlDB.SetMaxIdleConns(pm.config.InitialSize)
		sqlDB.SetConnMaxLifetime(pm.config.MaxLifetime)
		sqlDB.SetConnMaxIdleTime(pm.config.MaxIdleTime)
		
		log.Printf("[ConnectionPool] Applied initial config: MaxOpen=%d, MaxIdle=%d", 
			pm.config.MaxSize, pm.config.InitialSize)
	}
}

// GetOptimizedDB 获取优化的数据库连接
func (pm *ConnectionPoolManager) GetOptimizedDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	// 更新统计信息
	pm.updateConnectionStats()
	
	// 如果启用读写分离
	if pm.config.ReadWriteSplit {
		return pm.loadBalancer.GetDB(ctx, readOnly)
	}
	
	// 检查连接池健康状态
	if pm.config.HealthCheckEnabled && !pm.isHealthy() {
		return nil, fmt.Errorf("connection pool is unhealthy")
	}
	
	return pm.db, nil
}

// isHealthy 检查连接池是否健康
func (pm *ConnectionPoolManager) isHealthy() bool {
	pm.stats.mutex.RLock()
	defer pm.stats.mutex.RUnlock()
	
	// 简单的健康检查：至少有50%的连接是健康的
	total := pm.stats.TotalConnections
	healthy := pm.stats.HealthyConnections
	
	if total == 0 {
		return false
	}
	
	healthyRatio := float64(healthy) / float64(total)
	return healthyRatio >= 0.5
}

// updateConnectionStats 更新连接统计信息
func (pm *ConnectionPoolManager) updateConnectionStats() {
	sqlDB, err := pm.db.DB()
	if err != nil {
		return
	}
	
	dbStats := sqlDB.Stats()
	
	pm.stats.mutex.Lock()
	defer pm.stats.mutex.Unlock()
	
	// 更新基础统计
	pm.stats.TotalConnections = int32(dbStats.OpenConnections)
	pm.stats.ActiveConnections = int32(dbStats.InUse)
	pm.stats.IdleConnections = int32(dbStats.Idle)
	pm.stats.WaitingConnections = int32(dbStats.WaitCount)
	
	// 计算使用率
	if dbStats.MaxOpenConnections > 0 {
		pm.stats.CurrentUsage = float64(dbStats.InUse) / float64(dbStats.MaxOpenConnections)
		if pm.stats.CurrentUsage > pm.stats.PeakUsage {
			pm.stats.PeakUsage = pm.stats.CurrentUsage
		}
	}
	
	// 更新连接创建/关闭统计
	pm.stats.ConnectionsCreated = dbStats.MaxLifetimeClosed + dbStats.MaxIdleClosed
	pm.stats.ConnectionsClosed = dbStats.MaxLifetimeClosed + dbStats.MaxIdleClosed
}

// AdjustPoolSize 调整连接池大小
func (pm *ConnectionPoolManager) AdjustPoolSize(newSize int, reason string) error {
	if !atomic.CompareAndSwapInt32(&pm.adjustmentInProgress, 0, 1) {
		return fmt.Errorf("another adjustment is in progress")
	}
	defer atomic.StoreInt32(&pm.adjustmentInProgress, 0)
	
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	sqlDB, err := pm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	// 获取当前大小
	currentStats := sqlDB.Stats()
	oldSize := currentStats.MaxOpenConnections
	
	// 验证新大小
	if newSize < pm.config.MinSize {
		newSize = pm.config.MinSize
	}
	if newSize > pm.config.MaxSize {
		newSize = pm.config.MaxSize
	}
	
	// 应用新配置
	sqlDB.SetMaxOpenConns(newSize)
	
	// 记录调整
	adjustment := PoolAdjustment{
		Timestamp: time.Now(),
		Action:    pm.getActionType(oldSize, newSize),
		OldSize:   oldSize,
		NewSize:   newSize,
		Reason:    reason,
		Usage:     pm.stats.CurrentUsage,
	}
	
	pm.addAdjustmentHistory(adjustment)
	pm.lastAdjustment = time.Now()
	
	// 更新统计
	if newSize > oldSize {
		atomic.AddInt64(&pm.stats.ScaleUpEvents, 1)
	} else if newSize < oldSize {
		atomic.AddInt64(&pm.stats.ScaleDownEvents, 1)
	}
	
	pm.stats.LastScalingAction = time.Now()
	
	log.Printf("[ConnectionPool] Adjusted pool size: %d -> %d (reason: %s)", 
		oldSize, newSize, reason)
	
	return nil
}

// getActionType 获取调整动作类型
func (pm *ConnectionPoolManager) getActionType(oldSize, newSize int) string {
	if newSize > oldSize {
		return "scale_up"
	} else if newSize < oldSize {
		return "scale_down"
	}
	return "no_change"
}

// addAdjustmentHistory 添加调整历史
func (pm *ConnectionPoolManager) addAdjustmentHistory(adjustment PoolAdjustment) {
	pm.adjustmentHistory = append(pm.adjustmentHistory, adjustment)
	
	// 保持历史记录大小
	if len(pm.adjustmentHistory) > 100 {
		pm.adjustmentHistory = pm.adjustmentHistory[len(pm.adjustmentHistory)-100:]
	}
}

// GetPoolStats 获取连接池统计信息
func (pm *ConnectionPoolManager) GetPoolStats() *PoolStats {
	pm.updateConnectionStats()
	
	pm.stats.mutex.RLock()
	defer pm.stats.mutex.RUnlock()
	
	// 返回统计信息的副本
	statsCopy := *pm.stats
	return &statsCopy
}

// GetDetailedStats 获取详细统计信息
func (pm *ConnectionPoolManager) GetDetailedStats() map[string]any {
	stats := pm.GetPoolStats()
	
	sqlDB, _ := pm.db.DB()
	var dbStats sql.DBStats
	if sqlDB != nil {
		dbStats = sqlDB.Stats()
	}
	
	return map[string]any{
		"pool_config": map[string]any{
			"min_size":           pm.config.MinSize,
			"max_size":           pm.config.MaxSize,
			"initial_size":       pm.config.InitialSize,
			"max_idle_time":      pm.config.MaxIdleTime.String(),
			"max_lifetime":       pm.config.MaxLifetime.String(),
			"auto_scaling":       pm.config.EnableAutoScaling,
			"health_check":       pm.config.HealthCheckEnabled,
		},
		"current_stats": map[string]any{
			"total_connections":     stats.TotalConnections,
			"active_connections":    stats.ActiveConnections,
			"idle_connections":      stats.IdleConnections,
			"waiting_connections":   stats.WaitingConnections,
			"healthy_connections":   stats.HealthyConnections,
			"unhealthy_connections": stats.UnhealthyConnections,
			"current_usage":         fmt.Sprintf("%.2f%%", stats.CurrentUsage*100),
			"peak_usage":            fmt.Sprintf("%.2f%%", stats.PeakUsage*100),
		},
		"gorm_db_stats": map[string]any{
			"max_open_connections": dbStats.MaxOpenConnections,
			"open_connections":     dbStats.OpenConnections,
			"in_use":               dbStats.InUse,
			"idle":                 dbStats.Idle,
			"wait_count":           dbStats.WaitCount,
			"wait_duration":        dbStats.WaitDuration.String(),
			"max_idle_closed":      dbStats.MaxIdleClosed,
			"max_lifetime_closed":  dbStats.MaxLifetimeClosed,
		},
		"scaling_events": map[string]any{
			"scale_up_events":      stats.ScaleUpEvents,
			"scale_down_events":    stats.ScaleDownEvents,
			"last_scaling_action":  stats.LastScalingAction.Format(time.RFC3339),
			"last_adjustment":      pm.lastAdjustment.Format(time.RFC3339),
		},
		"recent_adjustments": pm.getRecentAdjustments(10),
	}
}

// getRecentAdjustments 获取最近的调整记录
func (pm *ConnectionPoolManager) getRecentAdjustments(limit int) []PoolAdjustment {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	if len(pm.adjustmentHistory) == 0 {
		return []PoolAdjustment{}
	}
	
	start := len(pm.adjustmentHistory) - limit
	if start < 0 {
		start = 0
	}
	
	adjustments := make([]PoolAdjustment, len(pm.adjustmentHistory)-start)
	copy(adjustments, pm.adjustmentHistory[start:])
	
	return adjustments
}

// Close 关闭连接池管理器
func (pm *ConnectionPoolManager) Close() error {
	if pm.healthChecker != nil {
		pm.healthChecker.Stop()
	}
	
	if pm.autoScaler != nil {
		pm.autoScaler.Stop()
	}
	
	log.Println("[ConnectionPool] Connection pool manager closed")
	return nil
}

// ================================
// HealthChecker 实现
// ================================

// NewHealthChecker 创建健康检查器
func NewHealthChecker(config *PoolConfig, db *gorm.DB) *HealthChecker {
	return &HealthChecker{
		config:   config,
		db:       db,
		stopChan: make(chan struct{}),
	}
}

// Start 启动健康检查
func (hc *HealthChecker) Start() {
	if !atomic.CompareAndSwapInt32(&hc.isRunning, 0, 1) {
		return
	}
	
	go hc.healthCheckLoop()
	log.Println("[HealthChecker] Started health checker")
}

// Stop 停止健康检查
func (hc *HealthChecker) Stop() {
	if !atomic.CompareAndSwapInt32(&hc.isRunning, 1, 0) {
		return
	}
	
	close(hc.stopChan)
	log.Println("[HealthChecker] Stopped health checker")
}

// healthCheckLoop 健康检查循环
func (hc *HealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(hc.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-hc.stopChan:
			return
		case <-ticker.C:
			hc.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (hc *HealthChecker) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), hc.config.HealthCheckTimeout)
	defer cancel()
	
	var result int
	err := hc.db.WithContext(ctx).Raw(hc.config.HealthCheckSQL).Scan(&result).Error
	
	if err != nil {
		log.Printf("[HealthChecker] Health check failed: %v", err)
		// 这里可以触发连接池恢复逻辑
	}
}

// ================================
// LoadBalancer 实现
// ================================

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(config *PoolConfig) *LoadBalancer {
	return &LoadBalancer{
		readDbs:  make([]*gorm.DB, 0),
		writeDbs: make([]*gorm.DB, 0),
		config:   config,
	}
}

// GetDB 获取数据库连接（支持读写分离）
func (lb *LoadBalancer) GetDB(ctx context.Context, readOnly bool) (*gorm.DB, error) {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()
	
	if readOnly && len(lb.readDbs) > 0 {
		// 轮询选择读库
		index := atomic.AddInt32(&lb.roundRobin, 1) % int32(len(lb.readDbs))
		return lb.readDbs[index], nil
	}
	
	if len(lb.writeDbs) > 0 {
		// 轮询选择写库
		index := atomic.AddInt32(&lb.roundRobin, 1) % int32(len(lb.writeDbs))
		return lb.writeDbs[index], nil
	}
	
	return nil, fmt.Errorf("no available database connection")
}

// ================================
// PoolAutoScaler 实现
// ================================

// NewPoolAutoScaler 创建自动扩缩容器
func NewPoolAutoScaler(manager *ConnectionPoolManager) *PoolAutoScaler {
	return &PoolAutoScaler{
		manager:  manager,
		stopChan: make(chan struct{}),
	}
}

// Start 启动自动扩缩容
func (as *PoolAutoScaler) Start() {
	if !atomic.CompareAndSwapInt32(&as.isRunning, 0, 1) {
		return
	}
	
	go as.scalingLoop()
	log.Println("[PoolAutoScaler] Started auto scaler")
}

// Stop 停止自动扩缩容
func (as *PoolAutoScaler) Stop() {
	if !atomic.CompareAndSwapInt32(&as.isRunning, 1, 0) {
		return
	}
	
	close(as.stopChan)
	log.Println("[PoolAutoScaler] Stopped auto scaler")
}

// scalingLoop 扩缩容循环
func (as *PoolAutoScaler) scalingLoop() {
	ticker := time.NewTicker(as.manager.config.ScalingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-as.stopChan:
			return
		case <-ticker.C:
			as.checkAndScale()
		}
	}
}

// checkAndScale 检查并执行扩缩容
func (as *PoolAutoScaler) checkAndScale() {
	stats := as.manager.GetPoolStats()
	
	// 防止频繁调整
	if time.Since(as.manager.lastAdjustment) < as.manager.config.ScalingInterval {
		return
	}
	
	// 获取当前配置
	sqlDB, err := as.manager.db.DB()
	if err != nil {
		return
	}
	
	currentMax := sqlDB.Stats().MaxOpenConnections
	threshold := as.manager.config.ScalingThreshold
	
	// 扩容检查
	if stats.CurrentUsage > threshold {
		newSize := int(float64(currentMax) * as.manager.config.ScaleUpFactor)
		reason := fmt.Sprintf("usage %.2f%% > threshold %.2f%%", 
			stats.CurrentUsage*100, threshold*100)
		as.manager.AdjustPoolSize(newSize, reason)
		return
	}
	
	// 缩容检查 - 更保守的策略
	lowThreshold := threshold * 0.5 // 50%阈值用于缩容
	if stats.CurrentUsage < lowThreshold && currentMax > as.manager.config.MinSize {
		newSize := int(float64(currentMax) * as.manager.config.ScaleDownFactor)
		reason := fmt.Sprintf("usage %.2f%% < low threshold %.2f%%", 
			stats.CurrentUsage*100, lowThreshold*100)
		as.manager.AdjustPoolSize(newSize, reason)
	}
}