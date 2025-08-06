// Package session 执行器实现
//
// 提供各种类型的SQL执行器，包括简单执行器、重用执行器、批处理执行器等
package session

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	
	"github.com/zsy619/yyhertz/framework/mybatis/config"
	"github.com/zsy619/yyhertz/framework/mybatis/cache"
)

// BaseExecutor 基础执行器
type BaseExecutor struct {
	configuration *config.Configuration
	db           any
	transaction  *Transaction
	localCache   cache.Cache
	closed       bool
	mutex        sync.RWMutex
}

// DefaultExecutor 默认执行器
type DefaultExecutor struct {
	*BaseExecutor
}

// ReuseExecutor 重用执行器
type ReuseExecutor struct {
	*BaseExecutor
	statementCache map[string]any
}

// BatchExecutor 批处理执行器
type BatchExecutor struct {
	*BaseExecutor
	statementList []any
	batchResultList []any
}

// CachingExecutor 缓存执行器
type CachingExecutor struct {
	delegate    Executor
	cache       cache.Cache
	flushCacheRequired bool
}

// Transaction 事务
type Transaction struct {
	db         *gorm.DB
	autoCommit bool
	timeout    *time.Duration
}

// DynamicSqlBuilder 动态SQL构建器 (从orm包移动到这里)
type DynamicSqlBuilder struct {
	paramIndex int
	parameters []any
}

// NewDefaultExecutor 创建默认执行器
func NewDefaultExecutor(configuration *config.Configuration, db any) *DefaultExecutor {
	baseExecutor := &BaseExecutor{
		configuration: configuration,
		db:           db,
		localCache:   cache.NewLruCache(cache.NewPerpetualCache("default"), 256), // 本地缓存
		closed:       false,
	}
	
	return &DefaultExecutor{
		BaseExecutor: baseExecutor,
	}
}

// NewReuseExecutor 创建重用执行器
func NewReuseExecutor(configuration *config.Configuration, db any) *ReuseExecutor {
	baseExecutor := &BaseExecutor{
		configuration: configuration,
		db:           db,
		localCache:   cache.NewLruCache(cache.NewPerpetualCache("reuse"), 256),
		closed:       false,
	}
	
	return &ReuseExecutor{
		BaseExecutor:   baseExecutor,
		statementCache: make(map[string]any),
	}
}

// NewBatchExecutor 创建批处理执行器
func NewBatchExecutor(configuration *config.Configuration, db any) *BatchExecutor {
	baseExecutor := &BaseExecutor{
		configuration: configuration,
		db:           db,
		localCache:   cache.NewLruCache(cache.NewPerpetualCache("batch"), 256),
		closed:       false,
	}
	
	return &BatchExecutor{
		BaseExecutor:    baseExecutor,
		statementList:   make([]any, 0),
		batchResultList: make([]any, 0),
	}
}

// NewCachingExecutor 创建缓存执行器
func NewCachingExecutor(delegate Executor, cache cache.Cache) *CachingExecutor {
	return &CachingExecutor{
		delegate: delegate,
		cache:    cache,
		flushCacheRequired: false,
	}
}

// Update 执行更新操作 (BaseExecutor)
func (executor *BaseExecutor) Update(ms *MappedStatement, parameter any) (int64, error) {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	
	if executor.closed {
		return 0, fmt.Errorf("executor is closed")
	}
	
	// 清除本地缓存
	executor.clearLocalCache()
	
	return executor.doUpdate(ms, parameter)
}

// Query 执行查询操作 (BaseExecutor)
func (executor *BaseExecutor) Query(ms *MappedStatement, parameter any, rowBounds *RowBounds, 
	resultHandler ResultHandler, cacheKey *CacheKey, boundSql *BoundSql) ([]any, error) {
	
	executor.mutex.RLock()
	defer executor.mutex.RUnlock()
	
	if executor.closed {
		return nil, fmt.Errorf("executor is closed")
	}
	
	// 检查本地缓存
	if cacheKey == nil {
		cacheKey = executor.CreateCacheKey(ms, parameter, rowBounds, boundSql)
	}
	
	return executor.queryFromDatabase(ms, parameter, rowBounds, resultHandler, cacheKey, boundSql)
}

// QueryCursor 执行游标查询 (BaseExecutor)
func (executor *BaseExecutor) QueryCursor(ms *MappedStatement, parameter any, rowBounds *RowBounds) (<-chan any, error) {
	boundSql := ms.SqlSource.GetBoundSql(parameter)
	return executor.doQueryCursor(ms, parameter, rowBounds, boundSql)
}

// Commit 提交事务 (BaseExecutor)
func (executor *BaseExecutor) Commit(required bool) error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	
	if executor.closed {
		return fmt.Errorf("executor is closed")
	}
	
	executor.clearLocalCache()
	
	if executor.transaction != nil && !executor.transaction.autoCommit {
		return executor.transaction.db.Commit().Error
	}
	return nil
}

// Rollback 回滚事务 (BaseExecutor)
func (executor *BaseExecutor) Rollback(required bool) error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	
	if executor.closed {
		return fmt.Errorf("executor is closed")
	}
	
	executor.clearLocalCache()
	
	if executor.transaction != nil && !executor.transaction.autoCommit {
		return executor.transaction.db.Rollback().Error
	}
	return nil
}

// Close 关闭执行器 (BaseExecutor)
func (executor *BaseExecutor) Close(forceRollback bool) error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	
	if !executor.closed {
		if forceRollback {
			executor.rollback(true)
		}
		executor.closed = true
	}
	return nil
}

// IsClosed 检查是否关闭 (BaseExecutor)
func (executor *BaseExecutor) IsClosed() bool {
	executor.mutex.RLock()
	defer executor.mutex.RUnlock()
	return executor.closed
}

// ClearLocalCache 清除本地缓存 (BaseExecutor)
func (executor *BaseExecutor) ClearLocalCache() {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	executor.clearLocalCache()
}

// CreateCacheKey 创建缓存键 (BaseExecutor)
func (executor *BaseExecutor) CreateCacheKey(ms *MappedStatement, parameterObject any, 
	rowBounds *RowBounds, boundSql *BoundSql) *CacheKey {
	
	cacheKey := &CacheKey{
		UpdateList: make([]any, 0),
		Count:      0,
	}
	
	cacheKey.UpdateList = append(cacheKey.UpdateList, ms.ID)
	cacheKey.UpdateList = append(cacheKey.UpdateList, rowBounds.Offset)
	cacheKey.UpdateList = append(cacheKey.UpdateList, rowBounds.Limit)
	cacheKey.UpdateList = append(cacheKey.UpdateList, boundSql.Sql)
	cacheKey.UpdateList = append(cacheKey.UpdateList, parameterObject)
	
	cacheKey.Count = len(cacheKey.UpdateList)
	return cacheKey
}

// IsCached 检查是否缓存 (BaseExecutor)
func (executor *BaseExecutor) IsCached(ms *MappedStatement, key *CacheKey) bool {
	executor.mutex.RLock()
	defer executor.mutex.RUnlock()
	
	keyStr := fmt.Sprintf("%v", key.UpdateList)
	_, exists := executor.localCache.Get(keyStr)
	return exists
}

// GetConnection 获取连接 (BaseExecutor)
func (executor *BaseExecutor) GetConnection() *gorm.DB {
	if db, ok := executor.db.(*gorm.DB); ok {
		return db
	}
	return nil
}

// SetExecutorWrapper 设置执行器包装器 (BaseExecutor)
func (executor *BaseExecutor) SetExecutorWrapper(wrapper ExecutorWrapper) {
	// 这里可以设置包装器
}

// 私有方法实现

// doUpdate 执行更新
func (executor *BaseExecutor) doUpdate(ms *MappedStatement, parameter any) (int64, error) {
	boundSql := ms.SqlSource.GetBoundSql(parameter)
	
	db := executor.GetConnection()
	if db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}
	
	// 构建SQL和参数
	sql, args := executor.buildSqlAndArgs(boundSql)
	
	// 执行SQL
	result := db.Exec(sql, args...)
	if result.Error != nil {
		return 0, result.Error
	}
	
	return result.RowsAffected, nil
}

// queryFromDatabase 从数据库查询
func (executor *BaseExecutor) queryFromDatabase(ms *MappedStatement, parameter any, rowBounds *RowBounds,
	resultHandler ResultHandler, cacheKey *CacheKey, boundSql *BoundSql) ([]any, error) {
	
	results, err := executor.doQuery(ms, parameter, rowBounds, resultHandler, boundSql)
	if err != nil {
		return nil, err
	}
	
	// 缓存结果
	if cacheKey != nil {
		executor.putToLocalCache(cacheKey, results)
	}
	
	return results, nil
}

// doQuery 执行查询
func (executor *BaseExecutor) doQuery(ms *MappedStatement, parameter any, rowBounds *RowBounds,
	resultHandler ResultHandler, boundSql *BoundSql) ([]any, error) {
	
	db := executor.GetConnection()
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	
	// 构建SQL和参数
	sql, args := executor.buildSqlAndArgs(boundSql)
	
	// 应用行边界
	if rowBounds.Limit > 0 {
		sql = fmt.Sprintf("%s LIMIT %d OFFSET %d", sql, rowBounds.Limit, rowBounds.Offset)
	}
	
	// 执行查询
	var results []map[string]any
	err := db.Raw(sql, args...).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	// 转换结果
	convertedResults := make([]any, len(results))
	for i, result := range results {
		convertedResults[i] = result
	}
	
	return convertedResults, nil
}

// doQueryCursor 执行游标查询
func (executor *BaseExecutor) doQueryCursor(ms *MappedStatement, parameter any, rowBounds *RowBounds, boundSql *BoundSql) (<-chan any, error) {
	ch := make(chan any, 100) // 缓冲通道
	
	go func() {
		defer close(ch)
		
		results, err := executor.doQuery(ms, parameter, rowBounds, nil, boundSql)
		if err != nil {
			return
		}
		
		for _, result := range results {
			ch <- result
		}
	}()
	
	return ch, nil
}

// buildSqlAndArgs 构建SQL和参数
func (executor *BaseExecutor) buildSqlAndArgs(boundSql *BoundSql) (string, []any) {
	sql := boundSql.Sql
	args := make([]any, 0)
	
	// 处理参数映射
	for _, paramMapping := range boundSql.ParameterMappings {
		value := executor.getParameterValue(boundSql.ParameterObject, paramMapping.Property)
		args = append(args, value)
	}
	
	return sql, args
}

// getParameterValue 获取参数值
func (executor *BaseExecutor) getParameterValue(parameterObject any, property string) any {
	if parameterObject == nil {
		return nil
	}
	
	// 简化实现，实际需要更复杂的参数处理
	if m, ok := parameterObject.(map[string]any); ok {
		return m[property]
	}
	
	return parameterObject
}

// clearLocalCache 清除本地缓存
func (executor *BaseExecutor) clearLocalCache() {
	if executor.localCache != nil {
		executor.localCache.Clear()
	}
}

// putToLocalCache 放入本地缓存
func (executor *BaseExecutor) putToLocalCache(cacheKey *CacheKey, results []any) {
	if executor.localCache != nil {
		keyStr := fmt.Sprintf("%v", cacheKey.UpdateList)
		executor.localCache.Put(keyStr, results)
	}
}

// rollback 回滚
func (executor *BaseExecutor) rollback(required bool) error {
	if executor.transaction != nil && !executor.transaction.autoCommit {
		return executor.transaction.db.Rollback().Error
	}
	return nil
}

// DefaultExecutor特有方法

// doUpdate 默认执行器的更新实现
func (executor *DefaultExecutor) doUpdate(ms *MappedStatement, parameter any) (int64, error) {
	return executor.BaseExecutor.doUpdate(ms, parameter)
}

// doQuery 默认执行器的查询实现
func (executor *DefaultExecutor) doQuery(ms *MappedStatement, parameter any, rowBounds *RowBounds,
	resultHandler ResultHandler, boundSql *BoundSql) ([]any, error) {
	
	return executor.BaseExecutor.doQuery(ms, parameter, rowBounds, resultHandler, boundSql)
}

// ReuseExecutor特有方法

// prepareStatement 重用执行器的预处理语句
func (executor *ReuseExecutor) prepareStatement(sql string) any {
	// 检查缓存
	if stmt, exists := executor.statementCache[sql]; exists {
		return stmt
	}
	
	// 创建新的预处理语句
	db := executor.GetConnection()
	if db == nil {
		return nil
	}
	
	// 这里应该创建真正的预处理语句
	newStmt := sql // 简化实现
	executor.statementCache[sql] = newStmt
	
	return newStmt
}

// BatchExecutor特有方法

// addBatch 批处理执行器添加批次
func (executor *BatchExecutor) addBatch(ms *MappedStatement, parameter any) {
	executor.statementList = append(executor.statementList, ms)
}

// doFlushStatements 批处理执行器刷新语句
func (executor *BatchExecutor) doFlushStatements() ([]any, error) {
	results := make([]any, len(executor.statementList))
	
	for i, _ := range executor.statementList {
		// 执行批处理
		results[i] = 1 // 模拟结果
	}
	
	// 清空批次
	executor.statementList = executor.statementList[:0]
	executor.batchResultList = executor.batchResultList[:0]
	
	return results, nil
}

// CachingExecutor特有方法

// Query 缓存执行器的查询实现
func (executor *CachingExecutor) Query(ms *MappedStatement, parameter any, rowBounds *RowBounds,
	resultHandler ResultHandler, cacheKey *CacheKey, boundSql *BoundSql) ([]any, error) {
	
	// 检查二级缓存
	if ms.UseCache && executor.cache != nil {
		if cacheKey == nil {
			cacheKey = executor.CreateCacheKey(ms, parameter, rowBounds, boundSql)
		}
		
		keyStr := fmt.Sprintf("%v", cacheKey.UpdateList)
		if cached, exists := executor.cache.Get(keyStr); exists {
			if results, ok := cached.([]any); ok {
				return results, nil
			}
		}
		
		// 从委托执行器查询
		results, err := executor.delegate.Query(ms, parameter, rowBounds, resultHandler, cacheKey, boundSql)
		if err != nil {
			return nil, err
		}
		
		// 缓存结果
		executor.cache.Put(keyStr, results)
		return results, nil
	}
	
	return executor.delegate.Query(ms, parameter, rowBounds, resultHandler, cacheKey, boundSql)
}

// Update 缓存执行器的更新实现
func (executor *CachingExecutor) Update(ms *MappedStatement, parameter any) (int64, error) {
	executor.flushCacheIfRequired(ms)
	return executor.delegate.Update(ms, parameter)
}

// flushCacheIfRequired 如果需要则刷新缓存
func (executor *CachingExecutor) flushCacheIfRequired(ms *MappedStatement) {
	if ms.FlushCacheRequired {
		executor.cache.Clear()
	}
}

// 委托方法实现
func (executor *CachingExecutor) QueryCursor(ms *MappedStatement, parameter any, rowBounds *RowBounds) (<-chan any, error) {
	return executor.delegate.QueryCursor(ms, parameter, rowBounds)
}

func (executor *CachingExecutor) Commit(required bool) error {
	return executor.delegate.Commit(required)
}

func (executor *CachingExecutor) Rollback(required bool) error {
	return executor.delegate.Rollback(required)
}

func (executor *CachingExecutor) Close(forceRollback bool) error {
	return executor.delegate.Close(forceRollback)
}

func (executor *CachingExecutor) IsClosed() bool {
	return executor.delegate.IsClosed()
}

func (executor *CachingExecutor) ClearLocalCache() {
	executor.delegate.ClearLocalCache()
}

func (executor *CachingExecutor) CreateCacheKey(ms *MappedStatement, parameterObject any, rowBounds *RowBounds, boundSql *BoundSql) *CacheKey {
	return executor.delegate.CreateCacheKey(ms, parameterObject, rowBounds, boundSql)
}

func (executor *CachingExecutor) IsCached(ms *MappedStatement, key *CacheKey) bool {
	return executor.delegate.IsCached(ms, key)
}

func (executor *CachingExecutor) GetConnection() *gorm.DB {
	return executor.delegate.GetConnection()
}

func (executor *CachingExecutor) SetExecutorWrapper(wrapper ExecutorWrapper) {
	executor.delegate.SetExecutorWrapper(wrapper)
}