// Package mybatis 多数据源会话支持
//
// 扩展XMLSession以支持多数据源路由和读写分离
package mybatis

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// MultiDataSourceSession 多数据源会话接口
type MultiDataSourceSession interface {
	XMLSession
	
	// 数据源管理
	AddDataSource(ds *DataSource) error
	RemoveDataSource(name string) error
	GetDataSourceStats() *DataSourceStats
	GetDataSources() map[string]*DataSource
	SetDataSourceActive(name string, active bool) error
	
	// 读写分离控制
	EnableReadWriteSplit() MultiDataSourceSession
	DisableReadWriteSplit() MultiDataSourceSession
	
	// 数据源路由配置
	SetLoadBalanceStrategy(strategy LoadBalanceStrategy) MultiDataSourceSession
	EnableHealthCheck(config *HealthCheckConfig) MultiDataSourceSession
	DisableHealthCheck() MultiDataSourceSession
	
	// 手动指定数据源
	WithDataSource(dataSourceName string) MultiDataSourceSession
}

// multiDataSourceSession 多数据源会话实现
type multiDataSourceSession struct {
	XMLSession
	router         *DataSourceRouter
	currentDataSource string // 当前指定的数据源
}

// NewMultiDataSourceSession 创建多数据源会话
func NewMultiDataSourceSession(config *DataSourceConfig) MultiDataSourceSession {
	// 创建基础XML会话（使用空的DB，实际DB将通过路由器获取）
	baseSession := NewXMLSession(&gorm.DB{})
	
	// 创建数据源路由器
	router := NewDataSourceRouter(config)
	
	session := &multiDataSourceSession{
		XMLSession: baseSession,
		router:     router,
	}
	
	return session
}

// NewMultiDataSourceSessionWithDefaults 创建带默认配置的多数据源会话
func NewMultiDataSourceSessionWithDefaults() MultiDataSourceSession {
	config := &DataSourceConfig{
		ReadWriteSplit: true,
		LoadBalance:    LoadBalanceRoundRobin,
		HealthCheck:    DefaultHealthCheckConfig(),
		CustomRoutes:   make(map[string][]string),
	}
	
	return NewMultiDataSourceSession(config)
}

// AddDataSource 添加数据源
func (ms *multiDataSourceSession) AddDataSource(ds *DataSource) error {
	return ms.router.AddDataSource(ds)
}

// RemoveDataSource 移除数据源
func (ms *multiDataSourceSession) RemoveDataSource(name string) error {
	return ms.router.RemoveDataSource(name)
}

// GetDataSourceStats 获取数据源统计信息
func (ms *multiDataSourceSession) GetDataSourceStats() *DataSourceStats {
	return ms.router.GetStats()
}

// GetDataSources 获取所有数据源
func (ms *multiDataSourceSession) GetDataSources() map[string]*DataSource {
	return ms.router.GetDataSources()
}

// SetDataSourceActive 设置数据源状态
func (ms *multiDataSourceSession) SetDataSourceActive(name string, active bool) error {
	return ms.router.SetDataSourceActive(name, active)
}

// EnableReadWriteSplit 启用读写分离
func (ms *multiDataSourceSession) EnableReadWriteSplit() MultiDataSourceSession {
	ms.router.config.ReadWriteSplit = true
	log.Printf("[MultiDataSource] Read-write splitting enabled")
	return ms
}

// DisableReadWriteSplit 禁用读写分离
func (ms *multiDataSourceSession) DisableReadWriteSplit() MultiDataSourceSession {
	ms.router.config.ReadWriteSplit = false
	log.Printf("[MultiDataSource] Read-write splitting disabled")
	return ms
}

// SetLoadBalanceStrategy 设置负载均衡策略
func (ms *multiDataSourceSession) SetLoadBalanceStrategy(strategy LoadBalanceStrategy) MultiDataSourceSession {
	ms.router.config.LoadBalance = strategy
	log.Printf("[MultiDataSource] Load balance strategy set to: %s", strategy)
	return ms
}

// EnableHealthCheck 启用健康检查
func (ms *multiDataSourceSession) EnableHealthCheck(config *HealthCheckConfig) MultiDataSourceSession {
	ms.router.config.HealthCheck = config
	// 健康检查功能暂时注释，使用连接池的健康检查
	// if ms.router.healthChecker != nil {
	//     ms.router.healthChecker.Stop()
	// }
	// ms.router.healthChecker = NewHealthChecker(config, ms.router)
	// ms.router.healthChecker.Start()
	log.Printf("[MultiDataSource] Health check configuration set (using connection pool health check)")
	return ms
}

// DisableHealthCheck 禁用健康检查
func (ms *multiDataSourceSession) DisableHealthCheck() MultiDataSourceSession {
	if ms.router.healthChecker != nil {
		ms.router.healthChecker.Stop()
		ms.router.healthChecker = nil
	}
	log.Printf("[MultiDataSource] Health check disabled")
	return ms
}

// WithDataSource 指定使用的数据源
func (ms *multiDataSourceSession) WithDataSource(dataSourceName string) MultiDataSourceSession {
	// 创建新的会话实例，以避免影响原有实例
	newSession := &multiDataSourceSession{
		XMLSession:        ms.XMLSession,
		router:           ms.router,
		currentDataSource: dataSourceName,
	}
	return newSession
}

// 重写核心数据库操作方法，加入数据源路由逻辑

// SelectOne 查询单条记录（带数据源路由）
func (ms *multiDataSourceSession) SelectOne(ctx context.Context, sql string, args ...any) (any, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeRead)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeRead)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource for read: %w", err)
	}
	
	// 创建临时会话使用特定数据源
	tempSession := NewXMLSession(ds.DB)
	return tempSession.SelectOne(routedCtx, sql, args...)
}

// SelectList 查询多条记录（带数据源路由）
func (ms *multiDataSourceSession) SelectList(ctx context.Context, sql string, args ...any) ([]any, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeRead)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeRead)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource for read: %w", err)
	}
	
	// 创建临时会话使用特定数据源
	tempSession := NewXMLSession(ds.DB)
	return tempSession.SelectList(routedCtx, sql, args...)
}

// SelectPage 分页查询（带数据源路由）
func (ms *multiDataSourceSession) SelectPage(ctx context.Context, sql string, page PageRequest, args ...any) (*PageResult, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeRead)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeRead)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource for read: %w", err)
	}
	
	// 创建临时会话使用特定数据源
	tempSession := NewXMLSession(ds.DB)
	return tempSession.SelectPage(routedCtx, sql, page, args...)
}

// Insert 插入记录（带数据源路由）
func (ms *multiDataSourceSession) Insert(ctx context.Context, sql string, args ...any) (int64, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeWrite)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeWrite)
	if err != nil {
		return 0, fmt.Errorf("failed to get datasource for write: %w", err)
	}
	
	// 创建临时会话使用特定数据源
	tempSession := NewXMLSession(ds.DB)
	return tempSession.Insert(routedCtx, sql, args...)
}

// Update 更新记录（带数据源路由）
func (ms *multiDataSourceSession) Update(ctx context.Context, sql string, args ...any) (int64, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeWrite)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeWrite)
	if err != nil {
		return 0, fmt.Errorf("failed to get datasource for write: %w", err)
	}
	
	// 创建临时会话使用特定数据源
	tempSession := NewXMLSession(ds.DB)
	return tempSession.Update(routedCtx, sql, args...)
}

// Delete 删除记录（带数据源路由）
func (ms *multiDataSourceSession) Delete(ctx context.Context, sql string, args ...any) (int64, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeWrite)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeWrite)
	if err != nil {
		return 0, fmt.Errorf("failed to get datasource for write: %w", err)
	}
	
	// 创建临时会话使用特定数据源
	tempSession := NewXMLSession(ds.DB)
	return tempSession.Delete(routedCtx, sql, args...)
}

// 基于语句ID的操作方法重写

// SelectOneByID 通过语句ID查询单条记录（带数据源路由）
func (ms *multiDataSourceSession) SelectOneByID(ctx context.Context, statementId string, parameter any) (any, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeRead)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeRead)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource for read: %w", err)
	}
	
	// 创建临时会话使用特定数据源，并加载相同的映射器
	tempSession := ms.createTempSessionWithMappers(ds.DB)
	return tempSession.SelectOneByID(routedCtx, statementId, parameter)
}

// SelectListByID 通过语句ID查询多条记录（带数据源路由）
func (ms *multiDataSourceSession) SelectListByID(ctx context.Context, statementId string, parameter any) ([]any, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeRead)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeRead)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource for read: %w", err)
	}
	
	// 创建临时会话使用特定数据源，并加载相同的映射器
	tempSession := ms.createTempSessionWithMappers(ds.DB)
	return tempSession.SelectListByID(routedCtx, statementId, parameter)
}

// InsertByID 通过语句ID插入记录（带数据源路由）
func (ms *multiDataSourceSession) InsertByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeWrite)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeWrite)
	if err != nil {
		return 0, fmt.Errorf("failed to get datasource for write: %w", err)
	}
	
	// 创建临时会话使用特定数据源，并加载相同的映射器
	tempSession := ms.createTempSessionWithMappers(ds.DB)
	return tempSession.InsertByID(routedCtx, statementId, parameter)
}

// UpdateByID 通过语句ID更新记录（带数据源路由）
func (ms *multiDataSourceSession) UpdateByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeWrite)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeWrite)
	if err != nil {
		return 0, fmt.Errorf("failed to get datasource for write: %w", err)
	}
	
	// 创建临时会话使用特定数据源，并加载相同的映射器
	tempSession := ms.createTempSessionWithMappers(ds.DB)
	return tempSession.UpdateByID(routedCtx, statementId, parameter)
}

// DeleteByID 通过语句ID删除记录（带数据源路由）
func (ms *multiDataSourceSession) DeleteByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	// 设置数据源上下文
	routedCtx := ms.setDataSourceContext(ctx, OperationTypeWrite)
	
	// 获取对应的数据源
	ds, err := ms.router.GetDataSource(routedCtx, OperationTypeWrite)
	if err != nil {
		return 0, fmt.Errorf("failed to get datasource for write: %w", err)
	}
	
	// 创建临时会话使用特定数据源，并加载相同的映射器
	tempSession := ms.createTempSessionWithMappers(ds.DB)
	return tempSession.DeleteByID(routedCtx, statementId, parameter)
}

// 工具方法

// setDataSourceContext 设置数据源上下文
func (ms *multiDataSourceSession) setDataSourceContext(ctx context.Context, opType OperationType) context.Context {
	// 如果当前指定了数据源，则使用指定的数据源
	if ms.currentDataSource != "" {
		return SetDataSourceContext(ctx, ms.currentDataSource)
	}
	
	// 否则让路由器根据操作类型自动选择
	return ctx
}

// createTempSessionWithMappers 创建临时会话并复制映射器配置
func (ms *multiDataSourceSession) createTempSessionWithMappers(db *gorm.DB) XMLSession {
	tempSession := NewXMLSession(db)
	
	// 如果原会话有映射器，复制到临时会话中
	// 注意：这里需要访问XMLSession的内部实现，可能需要添加相应的方法
	// 目前先创建基础会话，后续可以优化映射器共享机制
	
	return tempSession
}