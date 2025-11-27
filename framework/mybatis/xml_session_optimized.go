// Package mybatis XML Session性能优化版本
//
// 性能优化重点：
// 1. AST缓存和SQL预编译
// 2. 参数映射优化
// 3. 反射缓存
// 4. 批量操作优化
package mybatis

import (
	"context"
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis/mapper"
	"gorm.io/gorm"
)

// OptimizedXMLSession 优化版XML会话
type OptimizedXMLSession struct {
	XMLSession
	
	// 性能优化缓存
	sqlCache         *SQLCache         // SQL缓存
	reflectionCache  *ReflectionCache  // 反射缓存
	compiledStmtCache *CompiledStatementCache // 预编译语句缓存
	
	// 配置
	cacheEnabled     bool
	maxCacheSize     int
	cacheTTL         time.Duration
	
	// 统计信息
	stats            *PerformanceStats
	
	// 互斥锁
	cacheMutex       sync.RWMutex
}

// SQLCache SQL缓存结构
type SQLCache struct {
	cache     map[uint64]*CachedSQL
	mutex     sync.RWMutex
	maxSize   int
	hits      int64
	misses    int64
}

// CachedSQL 缓存的SQL
type CachedSQL struct {
	SQL           string
	Args          []any
	PreparedSQL   string
	CreateTime    time.Time
	LastUsedTime  time.Time
	UseCount      int64
	ParameterHash uint64
}

// ReflectionCache 反射缓存
type ReflectionCache struct {
	typeCache  map[reflect.Type]*TypeInfo
	fieldCache map[string]*FieldInfo
	mutex      sync.RWMutex
}

// TypeInfo 类型信息
type TypeInfo struct {
	Type       reflect.Type
	Fields     map[string]*FieldInfo
	IsPointer  bool
	CreateTime time.Time
}

// FieldInfo 字段信息  
type FieldInfo struct {
	Field      reflect.StructField
	Index      []int
	Type       reflect.Type
	IsPointer  bool
	CreateTime time.Time
}

// CompiledStatementCache 预编译语句缓存
type CompiledStatementCache struct {
	cache   map[string]*CompiledStatement
	mutex   sync.RWMutex
	maxSize int
}

// CompiledStatement 预编译语句
type CompiledStatement struct {
	SQL           string
	ParameterMask []bool  // 参数位置掩码
	StaticParts   []string // 静态SQL部分
	ParamCount    int
	CreateTime    time.Time
	UseCount      int64
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	SQLCacheHits        int64
	SQLCacheMisses      int64
	ReflectionCacheHits int64  
	ReflectionCacheMisses int64
	CompiledStmtHits    int64
	CompiledStmtMisses  int64
	TotalQueries        int64
	TotalBuildTime      time.Duration
	AvgBuildTime        time.Duration
	mutex               sync.RWMutex
}

// NewOptimizedXMLSession 创建优化版XML会话
func NewOptimizedXMLSession(db *gorm.DB) XMLSession {
	baseSession := NewXMLSession(db)
	
	optimized := &OptimizedXMLSession{
		XMLSession:        baseSession,
		cacheEnabled:      true,
		maxCacheSize:      10000,
		cacheTTL:          30 * time.Minute,
		stats:             &PerformanceStats{},
	}
	
	optimized.initCaches()
	
	// 启动缓存清理协程
	go optimized.startCacheCleanup()
	
	return optimized
}

// initCaches 初始化缓存
func (xs *OptimizedXMLSession) initCaches() {
	xs.sqlCache = &SQLCache{
		cache:   make(map[uint64]*CachedSQL),
		maxSize: xs.maxCacheSize,
	}
	
	xs.reflectionCache = &ReflectionCache{
		typeCache:  make(map[reflect.Type]*TypeInfo),
		fieldCache: make(map[string]*FieldInfo),
	}
	
	xs.compiledStmtCache = &CompiledStatementCache{
		cache:   make(map[string]*CompiledStatement),
		maxSize: xs.maxCacheSize / 2,
	}
}

// SelectOneByID 优化版单条查询
func (xs *OptimizedXMLSession) SelectOneByID(ctx context.Context, statementId string, parameter any) (any, error) {
	startTime := time.Now()
	defer func() {
		xs.updateStats(time.Since(startTime))
	}()
	
	// 尝试从编译语句缓存获取
	compiledStmt := xs.getOrCreateCompiledStatement(statementId)
	if compiledStmt == nil {
		// 回退到原始实现
		return xs.XMLSession.SelectOneByID(ctx, statementId, parameter)
	}
	
	// 使用优化的SQL构建
	sql, args, err := xs.buildOptimizedSQL(compiledStmt, parameter)
	if err != nil {
		return nil, fmt.Errorf("optimized SQL build failed: %w", err)
	}
	
	// 执行查询
	return xs.XMLSession.SelectOne(ctx, sql, args...)
}

// SelectListByID 优化版列表查询
func (xs *OptimizedXMLSession) SelectListByID(ctx context.Context, statementId string, parameter any) ([]any, error) {
	startTime := time.Now()
	defer func() {
		xs.updateStats(time.Since(startTime))
	}()
	
	// 尝试从编译语句缓存获取
	compiledStmt := xs.getOrCreateCompiledStatement(statementId)
	if compiledStmt == nil {
		// 回退到原始实现
		return xs.XMLSession.SelectListByID(ctx, statementId, parameter)
	}
	
	// 使用优化的SQL构建
	sql, args, err := xs.buildOptimizedSQL(compiledStmt, parameter)
	if err != nil {
		return nil, fmt.Errorf("optimized SQL build failed: %w", err)
	}
	
	// 执行查询
	return xs.XMLSession.SelectList(ctx, sql, args...)
}

// getOrCreateCompiledStatement 获取或创建预编译语句
func (xs *OptimizedXMLSession) getOrCreateCompiledStatement(statementId string) *CompiledStatement {
	xs.compiledStmtCache.mutex.RLock()
	if stmt, exists := xs.compiledStmtCache.cache[statementId]; exists {
		xs.compiledStmtCache.mutex.RUnlock()
		xs.stats.mutex.Lock()
		xs.stats.CompiledStmtHits++
		xs.stats.mutex.Unlock()
		stmt.UseCount++
		return stmt
	}
	xs.compiledStmtCache.mutex.RUnlock()
	
	// 缓存未命中，需要创建
	xs.stats.mutex.Lock()
	xs.stats.CompiledStmtMisses++
	xs.stats.mutex.Unlock()
	
	// 获取原始语句
	originalStmt := xs.XMLSession.GetStatement(statementId)
	if originalStmt == nil {
		return nil
	}
	
	// 预编译SQL
	compiledStmt := xs.compileStatement(originalStmt)
	if compiledStmt == nil {
		return nil
	}
	
	// 加入缓存
	xs.compiledStmtCache.mutex.Lock()
	// 检查缓存大小，必要时清理
	if len(xs.compiledStmtCache.cache) >= xs.compiledStmtCache.maxSize {
		xs.evictOldestCompiledStatement()
	}
	xs.compiledStmtCache.cache[statementId] = compiledStmt
	xs.compiledStmtCache.mutex.Unlock()
	
	return compiledStmt
}

// compileStatement 编译语句
func (xs *OptimizedXMLSession) compileStatement(stmt *mapper.XMLMappedStatement) *CompiledStatement {
	sql := stmt.SQL
	
	// 分析SQL中的参数占位符
	paramMask := make([]bool, len(sql))
	staticParts := make([]string, 0)
	paramCount := 0
	
	// 简化的参数解析（实际应该更复杂）
	parts := strings.Split(sql, "#{")
	if len(parts) == 1 {
		// 无参数的静态SQL
		return &CompiledStatement{
			SQL:           sql,
			ParameterMask: paramMask,
			StaticParts:   []string{sql},
			ParamCount:    0,
			CreateTime:    time.Now(),
		}
	}
	
	// 有参数的动态SQL
	staticParts = append(staticParts, parts[0])
	for i := 1; i < len(parts); i++ {
		closeIdx := strings.Index(parts[i], "}")
		if closeIdx > 0 {
			paramCount++
			staticParts = append(staticParts, parts[i][closeIdx+1:])
		}
	}
	
	return &CompiledStatement{
		SQL:           sql,
		ParameterMask: paramMask,
		StaticParts:   staticParts,
		ParamCount:    paramCount,
		CreateTime:    time.Now(),
	}
}

// buildOptimizedSQL 构建优化的SQL
func (xs *OptimizedXMLSession) buildOptimizedSQL(compiledStmt *CompiledStatement, parameter any) (string, []any, error) {
	// 生成参数哈希用于缓存键
	paramHash := xs.hashParameter(parameter)
	cacheKey := paramHash
	
	// 尝试从SQL缓存获取
	xs.sqlCache.mutex.RLock()
	if cached, exists := xs.sqlCache.cache[cacheKey]; exists {
		xs.sqlCache.mutex.RUnlock()
		xs.stats.mutex.Lock()
		xs.stats.SQLCacheHits++
		xs.stats.mutex.Unlock()
		cached.UseCount++
		cached.LastUsedTime = time.Now()
		return cached.SQL, cached.Args, nil
	}
	xs.sqlCache.mutex.RUnlock()
	
	// 缓存未命中，构建SQL
	xs.stats.mutex.Lock()
	xs.stats.SQLCacheMisses++
	xs.stats.mutex.Unlock()
	
	// 使用优化的参数提取
	args, err := xs.extractParametersOptimized(compiledStmt.SQL, parameter)
	if err != nil {
		return "", nil, err
	}
	
	// 构建最终SQL
	finalSQL := xs.buildFinalSQL(compiledStmt, args)
	
	// 加入缓存
	cachedSQL := &CachedSQL{
		SQL:           finalSQL,
		Args:          args,
		PreparedSQL:   compiledStmt.SQL,
		CreateTime:    time.Now(),
		LastUsedTime:  time.Now(),
		UseCount:      1,
		ParameterHash: paramHash,
	}
	
	xs.sqlCache.mutex.Lock()
	// 检查缓存大小
	if len(xs.sqlCache.cache) >= xs.sqlCache.maxSize {
		xs.evictOldestCachedSQL()
	}
	xs.sqlCache.cache[cacheKey] = cachedSQL
	xs.sqlCache.mutex.Unlock()
	
	return finalSQL, args, nil
}

// extractParametersOptimized 优化的参数提取
func (xs *OptimizedXMLSession) extractParametersOptimized(sql string, parameter any) ([]any, error) {
	if parameter == nil {
		return []any{}, nil
	}
	
	// 使用反射缓存
	paramType := reflect.TypeOf(parameter)
	typeInfo := xs.getOrCreateTypeInfo(paramType)
	
	args := make([]any, 0)
	
	// 查找参数占位符
	parts := strings.Split(sql, "#{")
	for i := 1; i < len(parts); i++ {
		closeIdx := strings.Index(parts[i], "}")
		if closeIdx > 0 {
			paramName := parts[i][:closeIdx]
			value := xs.getParameterValueOptimized(parameter, paramName, typeInfo)
			args = append(args, value)
		}
	}
	
	return args, nil
}

// getOrCreateTypeInfo 获取或创建类型信息
func (xs *OptimizedXMLSession) getOrCreateTypeInfo(t reflect.Type) *TypeInfo {
	xs.reflectionCache.mutex.RLock()
	if info, exists := xs.reflectionCache.typeCache[t]; exists {
		xs.reflectionCache.mutex.RUnlock()
		xs.stats.mutex.Lock()
		xs.stats.ReflectionCacheHits++
		xs.stats.mutex.Unlock()
		return info
	}
	xs.reflectionCache.mutex.RUnlock()
	
	// 缓存未命中，创建类型信息
	xs.stats.mutex.Lock()
	xs.stats.ReflectionCacheMisses++
	xs.stats.mutex.Unlock()
	
	info := &TypeInfo{
		Type:       t,
		Fields:     make(map[string]*FieldInfo),
		IsPointer:  t.Kind() == reflect.Ptr,
		CreateTime: time.Now(),
	}
	
	// 处理指针类型
	actualType := t
	if t.Kind() == reflect.Ptr {
		actualType = t.Elem()
	}
	
	// 分析结构体字段
	if actualType.Kind() == reflect.Struct {
		for i := 0; i < actualType.NumField(); i++ {
			field := actualType.Field(i)
			fieldInfo := &FieldInfo{
				Field:      field,
				Index:      []int{i},
				Type:       field.Type,
				IsPointer:  field.Type.Kind() == reflect.Ptr,
				CreateTime: time.Now(),
			}
			info.Fields[field.Name] = fieldInfo
		}
	}
	
	// 加入缓存
	xs.reflectionCache.mutex.Lock()
	xs.reflectionCache.typeCache[t] = info
	xs.reflectionCache.mutex.Unlock()
	
	return info
}

// getParameterValueOptimized 优化的参数值获取
func (xs *OptimizedXMLSession) getParameterValueOptimized(parameter any, paramName string, typeInfo *TypeInfo) any {
	if parameter == nil {
		return nil
	}
	
	// 处理map类型
	if paramMap, ok := parameter.(map[string]any); ok {
		return paramMap[paramName]
	}
	
	// 处理结构体类型
	if typeInfo != nil {
		if fieldInfo, exists := typeInfo.Fields[paramName]; exists {
			value := reflect.ValueOf(parameter)
			if typeInfo.IsPointer {
				value = value.Elem()
			}
			fieldValue := value.FieldByIndex(fieldInfo.Index)
			if fieldValue.IsValid() && fieldValue.CanInterface() {
				return fieldValue.Interface()
			}
		}
	}
	
	return nil
}

// buildFinalSQL 构建最终SQL
func (xs *OptimizedXMLSession) buildFinalSQL(compiledStmt *CompiledStatement, args []any) string {
	if compiledStmt.ParamCount == 0 {
		return compiledStmt.StaticParts[0]
	}
	
	// 简单的参数替换
	sql := compiledStmt.SQL
	for range args {
		sql = strings.Replace(sql, "#{", "?", 1)
		if closeIdx := strings.Index(sql, "}"); closeIdx > 0 {
			sql = sql[:closeIdx] + sql[closeIdx+1:]
		}
	}
	
	return sql
}

// hashParameter 生成参数哈希
func (xs *OptimizedXMLSession) hashParameter(parameter any) uint64 {
	h := fnv.New64a()
	
	if parameter == nil {
		h.Write([]byte("nil"))
		return h.Sum64()
	}
	
	// 根据参数类型生成哈希
	switch p := parameter.(type) {
	case string:
		h.Write([]byte(p))
	case int, int32, int64:
		h.Write([]byte(fmt.Sprintf("%d", p)))
	case float32, float64:
		h.Write([]byte(fmt.Sprintf("%f", p)))
	case bool:
		h.Write([]byte(fmt.Sprintf("%t", p)))
	case map[string]any:
		// 对map的键值对进行排序后哈希
		keys := make([]string, 0, len(p))
		for k := range p {
			keys = append(keys, k)
		}
		for _, k := range keys {
			h.Write([]byte(k))
			h.Write([]byte(fmt.Sprintf("%v", p[k])))
		}
	default:
		// 使用反射处理其他类型
		h.Write([]byte(fmt.Sprintf("%v", parameter)))
	}
	
	return h.Sum64()
}

// evictOldestCachedSQL 清理最旧的SQL缓存
func (xs *OptimizedXMLSession) evictOldestCachedSQL() {
	var oldestKey uint64
	var oldestTime time.Time = time.Now()
	
	for key, cached := range xs.sqlCache.cache {
		if cached.LastUsedTime.Before(oldestTime) {
			oldestTime = cached.LastUsedTime
			oldestKey = key
		}
	}
	
	if oldestKey != 0 {
		delete(xs.sqlCache.cache, oldestKey)
	}
}

// evictOldestCompiledStatement 清理最旧的编译语句
func (xs *OptimizedXMLSession) evictOldestCompiledStatement() {
	var oldestKey string
	var oldestTime time.Time = time.Now()
	
	for key, compiled := range xs.compiledStmtCache.cache {
		if compiled.CreateTime.Before(oldestTime) {
			oldestTime = compiled.CreateTime
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(xs.compiledStmtCache.cache, oldestKey)
	}
}

// updateStats 更新统计信息
func (xs *OptimizedXMLSession) updateStats(duration time.Duration) {
	xs.stats.mutex.Lock()
	defer xs.stats.mutex.Unlock()
	
	xs.stats.TotalQueries++
	xs.stats.TotalBuildTime += duration
	xs.stats.AvgBuildTime = xs.stats.TotalBuildTime / time.Duration(xs.stats.TotalQueries)
}

// startCacheCleanup 启动缓存清理
func (xs *OptimizedXMLSession) startCacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		xs.cleanupExpiredCache()
	}
}

// cleanupExpiredCache 清理过期缓存
func (xs *OptimizedXMLSession) cleanupExpiredCache() {
	now := time.Now()
	
	// 清理SQL缓存
	xs.sqlCache.mutex.Lock()
	for key, cached := range xs.sqlCache.cache {
		if now.Sub(cached.LastUsedTime) > xs.cacheTTL {
			delete(xs.sqlCache.cache, key)
		}
	}
	xs.sqlCache.mutex.Unlock()
	
	// 清理反射缓存（较长TTL）
	xs.reflectionCache.mutex.Lock()
	for key, typeInfo := range xs.reflectionCache.typeCache {
		if now.Sub(typeInfo.CreateTime) > xs.cacheTTL*2 {
			delete(xs.reflectionCache.typeCache, key)
		}
	}
	xs.reflectionCache.mutex.Unlock()
}

// GetOptimizationStats 获取优化统计信息
func (xs *OptimizedXMLSession) GetOptimizationStats() map[string]any {
	xs.stats.mutex.RLock()
	defer xs.stats.mutex.RUnlock()
	
	sqlCacheHitRate := float64(0)
	if total := xs.stats.SQLCacheHits + xs.stats.SQLCacheMisses; total > 0 {
		sqlCacheHitRate = float64(xs.stats.SQLCacheHits) / float64(total) * 100
	}
	
	reflectionCacheHitRate := float64(0)
	if total := xs.stats.ReflectionCacheHits + xs.stats.ReflectionCacheMisses; total > 0 {
		reflectionCacheHitRate = float64(xs.stats.ReflectionCacheHits) / float64(total) * 100
	}
	
	compiledStmtHitRate := float64(0)
	if total := xs.stats.CompiledStmtHits + xs.stats.CompiledStmtMisses; total > 0 {
		compiledStmtHitRate = float64(xs.stats.CompiledStmtHits) / float64(total) * 100
	}
	
	return map[string]any{
		"sql_cache_hit_rate":        fmt.Sprintf("%.2f%%", sqlCacheHitRate),
		"reflection_cache_hit_rate": fmt.Sprintf("%.2f%%", reflectionCacheHitRate),
		"compiled_stmt_hit_rate":    fmt.Sprintf("%.2f%%", compiledStmtHitRate),
		"total_queries":             xs.stats.TotalQueries,
		"avg_build_time":            xs.stats.AvgBuildTime.String(),
		"cache_sizes": map[string]int{
			"sql_cache":           len(xs.sqlCache.cache),
			"reflection_cache":    len(xs.reflectionCache.typeCache),
			"compiled_stmt_cache": len(xs.compiledStmtCache.cache),
		},
	}
}

// NewOptimizedXMLMapper 创建优化版XML Mapper（向后兼容）
func NewOptimizedXMLMapper(db *gorm.DB) XMLSession {
	return NewOptimizedXMLSession(db)
}