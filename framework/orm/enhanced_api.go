// Package orm 提供增强的API接口
//
// 基于GORM最佳实践，提供更简单、方便、快速的数据库操作方法
package orm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 简化的CRUD操作接口 =============

// SimpleCRUD 简化的CRUD操作
type SimpleCRUD struct {
	db *gorm.DB
}

// NewSimpleCRUD 创建简化的CRUD操作实例
func NewSimpleCRUD(db *gorm.DB) *SimpleCRUD {
	if db == nil {
		db = GetDefaultORM().DB()
	}
	return &SimpleCRUD{db: db}
}

// Create 创建记录
func (c *SimpleCRUD) Create(value interface{}) error {
	return c.db.Create(value).Error
}

// Save 保存记录（创建或更新）
func (c *SimpleCRUD) Save(value interface{}) error {
	return c.db.Save(value).Error
}

// Find 查询记录
func (c *SimpleCRUD) Find(dest interface{}, conds ...interface{}) error {
	return c.db.Find(dest, conds...).Error
}

// First 查询第一条记录
func (c *SimpleCRUD) First(dest interface{}, conds ...interface{}) error {
	return c.db.First(dest, conds...).Error
}

// Update 更新记录
func (c *SimpleCRUD) Update(model interface{}, column string, value interface{}) error {
	return c.db.Model(model).Update(column, value).Error
}

// Updates 批量更新记录
func (c *SimpleCRUD) Updates(model interface{}, values interface{}) error {
	return c.db.Model(model).Updates(values).Error
}

// Delete 删除记录
func (c *SimpleCRUD) Delete(value interface{}, conds ...interface{}) error {
	return c.db.Delete(value, conds...).Error
}

// Count 统计记录数
func (c *SimpleCRUD) Count(model interface{}) (int64, error) {
	var count int64
	err := c.db.Model(model).Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在
func (c *SimpleCRUD) Exists(model interface{}, conds ...interface{}) (bool, error) {
	var count int64
	err := c.db.Model(model).Where(conds[0], conds[1:]...).Count(&count).Error
	return count > 0, err
}

// Paginate 分页查询
func (c *SimpleCRUD) Paginate(model interface{}, page, pageSize int, dest interface{}) (int64, error) {
	var total int64
	
	// 先统计总数
	if err := c.db.Model(model).Count(&total).Error; err != nil {
		return 0, err
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	err := c.db.Model(model).Offset(offset).Limit(pageSize).Find(dest).Error
	return total, err
}

// Transaction 执行事务
func (c *SimpleCRUD) Transaction(fn func(*SimpleCRUD) error) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		txCRUD := &SimpleCRUD{db: tx}
		return fn(txCRUD)
	})
}

// ============= 增强的查询方法 =============

// WhereIn IN查询
func (c *SimpleCRUD) WhereIn(column string, values interface{}) *SimpleCRUD {
	return &SimpleCRUD{db: c.db.Where(fmt.Sprintf("%s IN ?", column), values)}
}

// WhereNotIn NOT IN查询
func (c *SimpleCRUD) WhereNotIn(column string, values interface{}) *SimpleCRUD {
	return &SimpleCRUD{db: c.db.Where(fmt.Sprintf("%s NOT IN ?", column), values)}
}

// WhereBetween BETWEEN查询
func (c *SimpleCRUD) WhereBetween(column string, start, end interface{}) *SimpleCRUD {
	return &SimpleCRUD{db: c.db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", column), start, end)}
}

// WhereNull NULL查询
func (c *SimpleCRUD) WhereNull(column string) *SimpleCRUD {
	return &SimpleCRUD{db: c.db.Where(fmt.Sprintf("%s IS NULL", column))}
}

// WhereNotNull NOT NULL查询
func (c *SimpleCRUD) WhereNotNull(column string) *SimpleCRUD {
	return &SimpleCRUD{db: c.db.Where(fmt.Sprintf("%s IS NOT NULL", column))}
}

// WhereLike LIKE查询
func (c *SimpleCRUD) WhereLike(column string, pattern string) *SimpleCRUD {
	return &SimpleCRUD{db: c.db.Where(fmt.Sprintf("%s LIKE ?", column), "%"+pattern+"%")}
}

// Where 基本条件查询
func (c *SimpleCRUD) Where(query interface{}, args ...interface{}) *SimpleCRUD {
	return &SimpleCRUD{db: c.db.Where(query, args...)}
}

// OrderBy 排序
func (c *SimpleCRUD) OrderBy(column string, direction ...string) *SimpleCRUD {
	dir := "ASC"
	if len(direction) > 0 && strings.ToUpper(direction[0]) == "DESC" {
		dir = "DESC"
	}
	return &SimpleCRUD{db: c.db.Order(fmt.Sprintf("%s %s", column, dir))}
}

// Limit 限制
func (c *SimpleCRUD) Limit(limit int) *SimpleCRUD {
	return &SimpleCRUD{db: c.db.Limit(limit)}
}

// ============= 错误处理增强 =============

// SafeExecute 安全执行数据库操作，带错误重试
func (c *SimpleCRUD) SafeExecute(operation func() error, maxRetries int) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// 检查是否为可重试的错误
		if !isRetryableDBError(err) {
			return err
		}
		
		// 等待一段时间后重试
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}
	
	return lastErr
}

// isRetryableDBError 检查是否为可重试的数据库错误
func isRetryableDBError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	retryableErrors := []string{
		"connection",
		"timeout",
		"deadlock",
		"lock",
	}
	
	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}
	
	return false
}

// ============= 模型辅助方法 =============

// ModelHelper 模型辅助类
type ModelHelper struct {
	db *gorm.DB
}

// NewModelHelper 创建模型辅助类
func NewModelHelper(db *gorm.DB) *ModelHelper {
	if db == nil {
		db = GetDefaultORM().DB()
	}
	return &ModelHelper{db: db}
}

// ValidateModel 验证模型
func (mh *ModelHelper) ValidateModel(model interface{}) error {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if !v.IsValid() {
		return errors.New("invalid model")
	}
	
	return nil
}

// GetTableName 获取表名
func (mh *ModelHelper) GetTableName(model interface{}) string {
	if tableNamer, ok := model.(interface{ TableName() string }); ok {
		return tableNamer.TableName()
	}
	
	stmt := &gorm.Statement{DB: mh.db}
	stmt.Parse(model)
	return stmt.Schema.Table
}

// CloneModel 克隆模型
func (mh *ModelHelper) CloneModel(src interface{}) interface{} {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	
	cloneVal := reflect.New(srcVal.Type())
	cloneVal.Elem().Set(srcVal)
	
	// 重置ID字段
	if idField := cloneVal.Elem().FieldByName("ID"); idField.IsValid() && idField.CanSet() {
		idField.Set(reflect.Zero(idField.Type()))
	}
	
	return cloneVal.Interface()
}

// ============= 性能优化方法 =============

// BatchCreate 批量创建
func (c *SimpleCRUD) BatchCreate(records interface{}, batchSize int) error {
	return c.db.CreateInBatches(records, batchSize).Error
}

// BulkUpdate 批量更新
func (c *SimpleCRUD) BulkUpdate(model interface{}, updates map[string]interface{}, conditions string, args ...interface{}) error {
	return c.db.Model(model).Where(conditions, args...).Updates(updates).Error
}

// BulkDelete 批量删除
func (c *SimpleCRUD) BulkDelete(model interface{}, conditions string, args ...interface{}) error {
	return c.db.Where(conditions, args...).Delete(model).Error
}

// ============= 缓存相关方法 =============

// WithCache 带缓存的查询（简化版）
type CachedQuery struct {
	*SimpleCRUD
	cacheKey string
	cacheTTL time.Duration
}

// NewCachedQuery 创建缓存查询
func NewCachedQuery(db *gorm.DB) *CachedQuery {
	return &CachedQuery{
		SimpleCRUD: NewSimpleCRUD(db),
		cacheTTL:   5 * time.Minute,
	}
}

// WithCacheKey 设置缓存键
func (cq *CachedQuery) WithCacheKey(key string) *CachedQuery {
	cq.cacheKey = key
	return cq
}

// WithTTL 设置缓存时间
func (cq *CachedQuery) WithTTL(ttl time.Duration) *CachedQuery {
	cq.cacheTTL = ttl
	return cq
}

// ============= 连接管理增强 =============

// ConnectionStats 连接统计
type ConnectionStats struct {
	MaxOpenConnections     int
	OpenConnections        int
	InUse                  int
	Idle                   int
	WaitCount              int64
	WaitDuration           time.Duration
	MaxIdleClosed          int64
	MaxIdleTimeClosed      int64
	MaxLifetimeClosed      int64
}

// GetConnectionStats 获取连接统计
func GetConnectionStats() (*ConnectionStats, error) {
	orm := GetDefaultORM()
	sqlDB, err := orm.DB().DB()
	if err != nil {
		return nil, err
	}
	
	stats := sqlDB.Stats()
	return &ConnectionStats{
		MaxOpenConnections:    stats.MaxOpenConnections,
		OpenConnections:       stats.OpenConnections,
		InUse:                 stats.InUse,
		Idle:                  stats.Idle,
		WaitCount:             stats.WaitCount,
		WaitDuration:          stats.WaitDuration,
		MaxIdleClosed:         stats.MaxIdleClosed,
		MaxIdleTimeClosed:     stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:     stats.MaxLifetimeClosed,
	}, nil
}

// PrintConnectionStats 打印连接统计
func PrintConnectionStats() {
	stats, err := GetConnectionStats()
	if err != nil {
		config.Errorf("Failed to get connection stats: %v", err)
		return
	}
	
	fmt.Printf("\n=== Database Connection Stats ===\n")
	fmt.Printf("Max Open Connections: %d\n", stats.MaxOpenConnections)
	fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
	fmt.Printf("In Use: %d\n", stats.InUse)
	fmt.Printf("Idle: %d\n", stats.Idle)
	fmt.Printf("Wait Count: %d\n", stats.WaitCount)
	fmt.Printf("Wait Duration: %v\n", stats.WaitDuration)
	fmt.Printf("================================\n\n")
}

// ============= 全局便捷函数 =============

// GetSimpleCRUD 获取默认SimpleCRUD实例
func GetSimpleCRUD() *SimpleCRUD {
	return NewSimpleCRUD(GetDefaultORM().DB())
}

// GetModelHelper 获取默认模型助手
func GetModelHelper() *ModelHelper {
	return NewModelHelper(GetDefaultORM().DB())
}

// QuickCreate 快速创建
func QuickCreate(value interface{}) error {
	return GetSimpleCRUD().Create(value)
}

// QuickFind 快速查询
func QuickFind(dest interface{}, conds ...interface{}) error {
	return GetSimpleCRUD().Find(dest, conds...)
}

// QuickFirst 快速查询第一条
func QuickFirst(dest interface{}, conds ...interface{}) error {
	return GetSimpleCRUD().First(dest, conds...)
}

// QuickUpdate 快速更新
func QuickUpdate(model interface{}, column string, value interface{}) error {
	return GetSimpleCRUD().Update(model, column, value)
}

// QuickDelete 快速删除
func QuickDelete(value interface{}, conds ...interface{}) error {
	return GetSimpleCRUD().Delete(value, conds...)
}

// QuickCount 快速计数
func QuickCount(model interface{}) (int64, error) {
	return GetSimpleCRUD().Count(model)
}

// QuickExists 快速检查存在
func QuickExists(model interface{}, conds ...interface{}) (bool, error) {
	return GetSimpleCRUD().Exists(model, conds...)
}

// QuickPaginate 快速分页
func QuickPaginate(model interface{}, page, pageSize int, dest interface{}) (int64, error) {
	return GetSimpleCRUD().Paginate(model, page, pageSize, dest)
}

// QuickTransaction 快速事务
func QuickTransaction(fn func(*SimpleCRUD) error) error {
	return GetSimpleCRUD().Transaction(fn)
}

// ============= 性能监控 =============

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	slowQueryThreshold time.Duration
	queryCount         int64
	slowQueryCount     int64
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(slowQueryThreshold time.Duration) *PerformanceMonitor {
	return &PerformanceMonitor{
		slowQueryThreshold: slowQueryThreshold,
	}
}

// RecordQuery 记录查询
func (pm *PerformanceMonitor) RecordQuery(duration time.Duration) {
	pm.queryCount++
	if duration > pm.slowQueryThreshold {
		pm.slowQueryCount++
		config.Warnf("Slow query detected: %v", duration)
	}
}

// GetStats 获取统计信息
func (pm *PerformanceMonitor) GetStats() map[string]interface{} {
	slowQueryRate := float64(0)
	if pm.queryCount > 0 {
		slowQueryRate = float64(pm.slowQueryCount) / float64(pm.queryCount) * 100
	}
	
	return map[string]interface{}{
		"total_queries":     pm.queryCount,
		"slow_queries":      pm.slowQueryCount,
		"slow_query_rate":   slowQueryRate,
		"threshold":         pm.slowQueryThreshold,
	}
}

// ============= 调试和工具方法 =============

// DebugMode 调试模式
func DebugMode(enabled bool) *SimpleCRUD {
	db := GetDefaultORM().DB()
	if enabled {
		db = db.Debug()
	}
	return NewSimpleCRUD(db)
}

// DryRun 预览SQL
func DryRun() *SimpleCRUD {
	db := GetDefaultORM().DB().Session(&gorm.Session{DryRun: true})
	return NewSimpleCRUD(db)
}

// ShowSQL 显示SQL语句
func ShowSQL(operation func(*SimpleCRUD)) {
	crud := DebugMode(true)
	operation(crud)
}

// ============= 健康检查 =============

// HealthCheck 健康检查
func HealthCheck() error {
	return GetDefaultORM().Ping()
}

// HealthCheckWithTimeout 带超时的健康检查
func HealthCheckWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	done := make(chan error, 1)
	go func() {
		done <- HealthCheck()
	}()
	
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}