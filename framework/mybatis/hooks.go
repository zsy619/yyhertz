// Package mybatis 提供常用钩子函数
//
// 使用Go函数式编程风格，避免Java式的过度抽象
package mybatis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// 事务上下文键
type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
	TxKey        contextKey = "transaction"
)

// PerformanceHook 性能监控钩子 - 记录慢查询
func PerformanceHook(slowThreshold time.Duration) (BeforeHook, AfterHook) {
	beforeHook := func(ctx context.Context, sql string, args []interface{}) error {
		// 可以在这里记录查询开始时间，但我们在AfterHook中使用传入的duration
		return nil
	}
	
	afterHook := func(ctx context.Context, result interface{}, duration time.Duration, err error) {
		if duration > slowThreshold {
			userID := getContextValue(ctx, UserIDKey, "unknown")
			log.Printf("[SLOW QUERY] User:%s Duration:%v Error:%v", userID, duration, err)
		}
	}
	
	return beforeHook, afterHook
}

// AuditHook 审计钩子 - 记录数据操作
func AuditHook() BeforeHook {
	return func(ctx context.Context, sql string, args []interface{}) error {
		userID := getContextValue(ctx, UserIDKey, "anonymous")
		requestID := getContextValue(ctx, RequestIDKey, "")
		
		// 检查是否是写操作
		if isWriteOperation(sql) {
			log.Printf("[AUDIT] User:%s Request:%s SQL:%s", userID, requestID, sql)
		}
		return nil
	}
}

// SecurityHook 安全检查钩子 - 防止SQL注入
func SecurityHook() BeforeHook {
	return func(ctx context.Context, sql string, args []interface{}) error {
		// 简单的SQL注入检查
		if containsSQLInjectionPatterns(sql) {
			userID := getContextValue(ctx, UserIDKey, "unknown")
			log.Printf("[SECURITY ALERT] Potential SQL injection from User:%s SQL:%s", userID, sql)
			return fmt.Errorf("potential SQL injection detected")
		}
		return nil
	}
}

// MetricsHook 指标收集钩子 - 收集执行统计
func MetricsHook(collector MetricsCollector) (BeforeHook, AfterHook) {
	beforeHook := func(ctx context.Context, sql string, args []interface{}) error {
		collector.IncrementQueryCount(getOperationType(sql))
		return nil
	}
	
	afterHook := func(ctx context.Context, result interface{}, duration time.Duration, err error) {
		collector.RecordQueryDuration(duration)
		if err != nil {
			collector.IncrementErrorCount()
		}
	}
	
	return beforeHook, afterHook
}

// TransactionHook 事务追踪钩子
func TransactionHook() BeforeHook {
	return func(ctx context.Context, sql string, args []interface{}) error {
		if tx := getContextValue(ctx, TxKey, nil); tx != nil {
			log.Printf("[TRANSACTION] In transaction, SQL: %s", sql)
		}
		return nil
	}
}

// CacheHook 缓存钩子 - 简单的查询结果缓存
func CacheHook(cache Cache) (BeforeHook, AfterHook) {
	beforeHook := func(ctx context.Context, sql string, args []interface{}) error {
		// 在这里可以检查缓存，但由于钩子的限制，我们主要在after中处理
		return nil
	}
	
	afterHook := func(ctx context.Context, result interface{}, duration time.Duration, err error) {
		if err == nil && isSelectOperation(sql(ctx)) {
			cacheKey := generateCacheKey(sql(ctx), getArgs(ctx))
			cache.Set(cacheKey, result, 5*time.Minute) // 缓存5分钟
		}
	}
	
	return beforeHook, afterHook
}

// DebugHook 调试信息钩子 - 输出详细调试信息
func DebugHook() (BeforeHook, AfterHook) {
	beforeHook := func(ctx context.Context, sql string, args []interface{}) error {
		log.Printf("[DEBUG] Executing SQL: %s with args: %+v", sql, args)
		return nil
	}
	
	afterHook := func(ctx context.Context, result interface{}, duration time.Duration, err error) {
		if err != nil {
			log.Printf("[DEBUG] Query failed in %v: %v", duration, err)
		} else {
			resultCount := getResultCount(result)
			log.Printf("[DEBUG] Query completed in %v, returned %d results", duration, resultCount)
		}
	}
	
	return beforeHook, afterHook
}

// 辅助函数

// getContextValue 安全获取context值
func getContextValue(ctx context.Context, key contextKey, defaultValue interface{}) interface{} {
	if value := ctx.Value(key); value != nil {
		return value
	}
	return defaultValue
}

// isWriteOperation 检查是否是写操作
func isWriteOperation(sql string) bool {
	sql = normalizeSQL(sql)
	return startsWith(sql, "INSERT") || startsWith(sql, "UPDATE") || startsWith(sql, "DELETE")
}

// isSelectOperation 检查是否是查询操作
func isSelectOperation(sql string) bool {
	return startsWith(normalizeSQL(sql), "SELECT")
}

// getOperationType 获取操作类型
func getOperationType(sql string) string {
	sql = normalizeSQL(sql)
	if startsWith(sql, "SELECT") {
		return "SELECT"
	} else if startsWith(sql, "INSERT") {
		return "INSERT"
	} else if startsWith(sql, "UPDATE") {
		return "UPDATE"
	} else if startsWith(sql, "DELETE") {
		return "DELETE"
	}
	return "OTHER"
}

// containsSQLInjectionPatterns 简单的SQL注入检查
func containsSQLInjectionPatterns(sql string) bool {
	sql = normalizeSQL(sql)
	patterns := []string{
		"'; DROP TABLE",
		"'; DELETE FROM",
		"UNION SELECT",
		"' OR '1'='1",
		"' OR 1=1",
	}
	
	for _, pattern := range patterns {
		if contains(sql, pattern) {
			return true
		}
	}
	return false
}

// getResultCount 获取结果数量
func getResultCount(result interface{}) int {
	switch r := result.(type) {
	case []interface{}:
		return len(r)
	case *PageResult:
		return len(r.Items)
	case nil:
		return 0
	default:
		return 1
	}
}

// generateCacheKey 生成缓存键
func generateCacheKey(sql string, args []interface{}) string {
	return fmt.Sprintf("sql:%s:args:%+v", sql, args)
}

// 工具函数

func normalizeSQL(sql string) string {
	return strings.ToUpper(strings.TrimSpace(sql))
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// sql 和 getArgs 是从 context 中获取 SQL 和参数的辅助函数
// 这里简化实现，实际使用中需要在 context 中存储这些信息
func sql(ctx context.Context) string {
	if s, ok := ctx.Value("current_sql").(string); ok {
		return s
	}
	return ""
}

func getArgs(ctx context.Context) []interface{} {
	if args, ok := ctx.Value("current_args").([]interface{}); ok {
		return args
	}
	return nil
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	IncrementQueryCount(operation string)
	RecordQueryDuration(duration time.Duration)
	IncrementErrorCount()
}

// Cache 简单缓存接口
type Cache interface {
	Set(key string, value interface{}, duration time.Duration)
	Get(key string) (interface{}, bool)
	Delete(key string)
}

// SimpleMetricsCollector 简单指标收集器实现
type SimpleMetricsCollector struct {
	QueryCounts    map[string]int64
	TotalDuration  time.Duration
	ErrorCount     int64
}

func NewSimpleMetricsCollector() *SimpleMetricsCollector {
	return &SimpleMetricsCollector{
		QueryCounts: make(map[string]int64),
	}
}

func (c *SimpleMetricsCollector) IncrementQueryCount(operation string) {
	c.QueryCounts[operation]++
}

func (c *SimpleMetricsCollector) RecordQueryDuration(duration time.Duration) {
	c.TotalDuration += duration
}

func (c *SimpleMetricsCollector) IncrementErrorCount() {
	c.ErrorCount++
}

// GetStats 获取统计信息
func (c *SimpleMetricsCollector) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"query_counts":   c.QueryCounts,
		"total_duration": c.TotalDuration,
		"error_count":    c.ErrorCount,
	}
}

// SimpleCache 简单内存缓存实现
type SimpleCache struct {
	data map[string]cacheItem
}

type cacheItem struct {
	value  interface{}
	expiry time.Time
}

func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		data: make(map[string]cacheItem),
	}
}

func (c *SimpleCache) Set(key string, value interface{}, duration time.Duration) {
	c.data[key] = cacheItem{
		value:  value,
		expiry: time.Now().Add(duration),
	}
}

func (c *SimpleCache) Get(key string) (interface{}, bool) {
	item, exists := c.data[key]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(item.expiry) {
		delete(c.data, key)
		return nil, false
	}
	
	return item.value, true
}

func (c *SimpleCache) Delete(key string) {
	delete(c.data, key)
}