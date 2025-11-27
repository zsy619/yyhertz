// Package mybatis 数据源管理和读写分离
//
// 实现多数据源路由，支持读写分离和动态数据源切换
package mybatis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"gorm.io/gorm"
)

// DataSourceType 数据源类型
type DataSourceType string

const (
	DataSourceTypeMaster DataSourceType = "master" // 主数据源（写）
	DataSourceTypeSlave  DataSourceType = "slave"  // 从数据源（读）
	DataSourceTypeCustom DataSourceType = "custom" // 自定义数据源
)

// DataSource 数据源配置
type DataSource struct {
	Name     string         `json:"name"`     // 数据源名称
	Type     DataSourceType `json:"type"`     // 数据源类型
	Weight   int            `json:"weight"`   // 权重（用于负载均衡）
	DB       *gorm.DB       `json:"-"`        // 数据库连接
	IsActive bool           `json:"isActive"` // 是否活跃
}

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	DefaultDataSource string                 `json:"defaultDataSource"` // 默认数据源
	ReadWriteSplit    bool                   `json:"readWriteSplit"`    // 是否启用读写分离
	LoadBalance       LoadBalanceStrategy    `json:"loadBalance"`       // 负载均衡策略
	HealthCheck       *HealthCheckConfig     `json:"healthCheck"`       // 健康检查配置
	CustomRoutes      map[string][]string    `json:"customRoutes"`      // 自定义路由规则
}

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	LoadBalanceRoundRobin LoadBalanceStrategy = "round_robin" // 轮询
	LoadBalanceRandom     LoadBalanceStrategy = "random"      // 随机
	LoadBalanceWeighted   LoadBalanceStrategy = "weighted"    // 权重
)

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Interval    time.Duration `json:"interval"`    // 检查间隔
	Timeout     time.Duration `json:"timeout"`     // 超时时间
	MaxRetries  int           `json:"maxRetries"`  // 最大重试次数
	TestSQL     string        `json:"testSQL"`     // 测试SQL
	Enabled     bool          `json:"enabled"`     // 是否启用
}

// DefaultHealthCheckConfig 默认健康检查配置
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		Interval:   30 * time.Second,
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		TestSQL:    "SELECT 1",
		Enabled:    true,
	}
}

// DataSourceRouter 数据源路由器
type DataSourceRouter struct {
	config      *DataSourceConfig
	dataSources map[string]*DataSource // 所有数据源
	masters     []*DataSource          // 主数据源列表
	slaves      []*DataSource          // 从数据源列表
	custom      map[string][]*DataSource // 自定义数据源组
	
	// 负载均衡状态
	rrIndex     int           // 轮询索引
	mutex       sync.RWMutex  // 读写锁
	
	// 健康检查
	healthChecker *HealthChecker
	
	// 统计信息
	stats *DataSourceStats
}

// DataSourceStats 数据源统计
type DataSourceStats struct {
	TotalRequests     int64            `json:"totalRequests"`
	ReadRequests      int64            `json:"readRequests"`
	WriteRequests     int64            `json:"writeRequests"`
	CustomRequests    int64            `json:"customRequests"`
	DataSourceUsage   map[string]int64 `json:"dataSourceUsage"`
	FailedConnections int64            `json:"failedConnections"`
	mutex             sync.RWMutex
}

// NewDataSourceRouter 创建数据源路由器
func NewDataSourceRouter(config *DataSourceConfig) *DataSourceRouter {
	if config == nil {
		config = &DataSourceConfig{
			ReadWriteSplit: false,
			LoadBalance:    LoadBalanceRoundRobin,
			HealthCheck:    DefaultHealthCheckConfig(),
			CustomRoutes:   make(map[string][]string),
		}
	}

	router := &DataSourceRouter{
		config:      config,
		dataSources: make(map[string]*DataSource),
		masters:     make([]*DataSource, 0),
		slaves:      make([]*DataSource, 0),
		custom:      make(map[string][]*DataSource),
		stats: &DataSourceStats{
			DataSourceUsage: make(map[string]int64),
		},
	}

	// 健康检查功能暂时注释，使用连接池的健康检查
	// if config.HealthCheck != nil && config.HealthCheck.Enabled {
	//     router.healthChecker = NewHealthChecker(config.HealthCheck, router)
	//     router.healthChecker.Start()
	// }

	return router
}

// AddDataSource 添加数据源
func (r *DataSourceRouter) AddDataSource(ds *DataSource) error {
	if ds == nil || ds.DB == nil {
		return errors.New("datasource and db cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 添加到总列表
	r.dataSources[ds.Name] = ds

	// 根据类型分类
	switch ds.Type {
	case DataSourceTypeMaster:
		r.masters = append(r.masters, ds)
	case DataSourceTypeSlave:
		r.slaves = append(r.slaves, ds)
	case DataSourceTypeCustom:
		// 自定义数据源需要配置路由规则
	}

	// 初始化统计
	r.stats.mutex.Lock()
	r.stats.DataSourceUsage[ds.Name] = 0
	r.stats.mutex.Unlock()

	log.Printf("[DataSource] Added data source: %s (type: %s, weight: %d)", 
		ds.Name, ds.Type, ds.Weight)
	return nil
}

// RemoveDataSource 移除数据源
func (r *DataSourceRouter) RemoveDataSource(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	ds, exists := r.dataSources[name]
	if !exists {
		return fmt.Errorf("datasource not found: %s", name)
	}

	// 从分类列表中移除
	switch ds.Type {
	case DataSourceTypeMaster:
		r.masters = r.removeFromSlice(r.masters, name)
	case DataSourceTypeSlave:
		r.slaves = r.removeFromSlice(r.slaves, name)
	}

	// 从总列表中移除
	delete(r.dataSources, name)

	log.Printf("[DataSource] Removed data source: %s", name)
	return nil
}

// removeFromSlice 从切片中移除指定数据源
func (r *DataSourceRouter) removeFromSlice(slice []*DataSource, name string) []*DataSource {
	for i, ds := range slice {
		if ds.Name == name {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// GetDataSource 获取数据源（根据上下文和操作类型）
func (r *DataSourceRouter) GetDataSource(ctx context.Context, operationType OperationType) (*DataSource, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 更新统计
	r.updateStats(operationType)

	// 检查上下文中是否指定了特定数据源
	if dsName := GetDataSourceFromContext(ctx); dsName != "" {
		if ds, exists := r.dataSources[dsName]; exists && ds.IsActive {
			r.incrementUsage(ds.Name)
			return ds, nil
		}
	}

	// 根据操作类型选择数据源
	var candidates []*DataSource
	switch operationType {
	case OperationTypeRead:
		if r.config.ReadWriteSplit && len(r.slaves) > 0 {
			candidates = r.getActiveSources(r.slaves)
		} else {
			candidates = r.getActiveSources(r.masters)
		}
	case OperationTypeWrite:
		candidates = r.getActiveSources(r.masters)
	default:
		// 默认使用主数据源
		candidates = r.getActiveSources(r.masters)
	}

	if len(candidates) == 0 {
		r.stats.mutex.Lock()
		r.stats.FailedConnections++
		r.stats.mutex.Unlock()
		return nil, errors.New("no active datasource available")
	}

	// 根据负载均衡策略选择数据源
	selected := r.selectDataSource(candidates)
	r.incrementUsage(selected.Name)
	return selected, nil
}

// getActiveSources 获取活跃的数据源
func (r *DataSourceRouter) getActiveSources(sources []*DataSource) []*DataSource {
	active := make([]*DataSource, 0, len(sources))
	for _, ds := range sources {
		if ds.IsActive {
			active = append(active, ds)
		}
	}
	return active
}

// selectDataSource 根据负载均衡策略选择数据源
func (r *DataSourceRouter) selectDataSource(candidates []*DataSource) *DataSource {
	if len(candidates) == 1 {
		return candidates[0]
	}

	switch r.config.LoadBalance {
	case LoadBalanceRandom:
		return candidates[rand.Intn(len(candidates))]
	case LoadBalanceWeighted:
		return r.selectByWeight(candidates)
	case LoadBalanceRoundRobin:
		fallthrough
	default:
		r.rrIndex = (r.rrIndex + 1) % len(candidates)
		return candidates[r.rrIndex]
	}
}

// selectByWeight 根据权重选择数据源
func (r *DataSourceRouter) selectByWeight(candidates []*DataSource) *DataSource {
	totalWeight := 0
	for _, ds := range candidates {
		totalWeight += ds.Weight
	}

	if totalWeight == 0 {
		// 如果所有权重都是0，使用轮询
		r.rrIndex = (r.rrIndex + 1) % len(candidates)
		return candidates[r.rrIndex]
	}

	randomWeight := rand.Intn(totalWeight)
	currentWeight := 0

	for _, ds := range candidates {
		currentWeight += ds.Weight
		if currentWeight > randomWeight {
			return ds
		}
	}

	return candidates[0] // 备用返回
}

// updateStats 更新统计信息
func (r *DataSourceRouter) updateStats(operationType OperationType) {
	r.stats.mutex.Lock()
	defer r.stats.mutex.Unlock()

	r.stats.TotalRequests++
	switch operationType {
	case OperationTypeRead:
		r.stats.ReadRequests++
	case OperationTypeWrite:
		r.stats.WriteRequests++
	default:
		r.stats.CustomRequests++
	}
}

// incrementUsage 增加数据源使用计数
func (r *DataSourceRouter) incrementUsage(dsName string) {
	r.stats.mutex.Lock()
	defer r.stats.mutex.Unlock()
	r.stats.DataSourceUsage[dsName]++
}

// GetStats 获取统计信息
func (r *DataSourceRouter) GetStats() *DataSourceStats {
	r.stats.mutex.RLock()
	defer r.stats.mutex.RUnlock()

	// 深拷贝统计信息
	stats := &DataSourceStats{
		TotalRequests:     r.stats.TotalRequests,
		ReadRequests:      r.stats.ReadRequests,
		WriteRequests:     r.stats.WriteRequests,
		CustomRequests:    r.stats.CustomRequests,
		FailedConnections: r.stats.FailedConnections,
		DataSourceUsage:   make(map[string]int64),
	}

	for k, v := range r.stats.DataSourceUsage {
		stats.DataSourceUsage[k] = v
	}

	return stats
}

// GetDataSources 获取所有数据源信息
func (r *DataSourceRouter) GetDataSources() map[string]*DataSource {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[string]*DataSource)
	for k, v := range r.dataSources {
		result[k] = v
	}
	return result
}

// SetDataSourceActive 设置数据源状态
func (r *DataSourceRouter) SetDataSourceActive(name string, active bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	ds, exists := r.dataSources[name]
	if !exists {
		return fmt.Errorf("datasource not found: %s", name)
	}

	ds.IsActive = active
	log.Printf("[DataSource] Set data source %s active: %v", name, active)
	return nil
}

// Close 关闭路由器和清理资源
func (r *DataSourceRouter) Close() error {
	// 停止健康检查
	if r.healthChecker != nil {
		r.healthChecker.Stop()
	}

	// 关闭所有数据库连接
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for name, ds := range r.dataSources {
		if db, err := ds.DB.DB(); err == nil {
			if err := db.Close(); err != nil {
				log.Printf("[DataSource] Failed to close database %s: %v", name, err)
			}
		}
	}

	log.Printf("[DataSource] Router closed")
	return nil
}

// OperationType 操作类型
type OperationType string

const (
	OperationTypeRead  OperationType = "read"  // 读操作
	OperationTypeWrite OperationType = "write" // 写操作
)

// Context中数据源相关的key
type dataSourceContextKey string

const (
	DataSourceContextKey dataSourceContextKey = "datasource"
)

// SetDataSourceContext 在上下文中设置数据源
func SetDataSourceContext(ctx context.Context, dataSourceName string) context.Context {
	return context.WithValue(ctx, DataSourceContextKey, dataSourceName)
}

// GetDataSourceFromContext 从上下文中获取数据源名称
func GetDataSourceFromContext(ctx context.Context) string {
	if dsName, ok := ctx.Value(DataSourceContextKey).(string); ok {
		return dsName
	}
	return ""
}

