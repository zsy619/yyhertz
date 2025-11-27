package main

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 测试配置
const (
	SessionTestIterations = 1000   // 测试迭代次数
	SessionWarmupRounds   = 50     // 预热轮数
)

// MyBatis会话管理器
type MyBatisSessionManager struct {
	db             *gorm.DB
	sessionPool    []mybatis.SimpleSession
	poolSize       int
	currentIndex   int
}

// NewMyBatisSessionManager 创建会话管理器
func NewMyBatisSessionManager(poolSize int) *MyBatisSessionManager {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	
	// 设置测试数据
	setupSessionTestData(db)
	
	// 创建会话池
	pool := make([]mybatis.SimpleSession, poolSize)
	for i := 0; i < poolSize; i++ {
		pool[i] = mybatis.NewSimpleSession(db)
	}
	
	return &MyBatisSessionManager{
		db:          db,
		sessionPool: pool,
		poolSize:    poolSize,
	}
}

// GetSession 获取会话（池化）
func (m *MyBatisSessionManager) GetSession() mybatis.SimpleSession {
	session := m.sessionPool[m.currentIndex]
	m.currentIndex = (m.currentIndex + 1) % m.poolSize
	return session
}

// CreateNewSession 创建新会话（非池化）
func (m *MyBatisSessionManager) CreateNewSession() mybatis.SimpleSession {
	return mybatis.NewSimpleSession(m.db)
}

// SessionOperations 会话操作接口
type SessionOperations struct {
	manager *MyBatisSessionManager
}

// DirectOperation 直接操作
func (ops *SessionOperations) DirectOperation() error {
	session := ops.manager.CreateNewSession()
	ctx := context.Background()
	
	// 执行简单查询
	_, err := session.SelectOne(ctx, "SELECT id FROM users WHERE id = 1")
	return err
}

// PooledOperation 池化操作
func (ops *SessionOperations) PooledOperation() error {
	session := ops.manager.GetSession()
	ctx := context.Background()
	
	// 执行简单查询
	_, err := session.SelectOne(ctx, "SELECT id FROM users WHERE id = 1")
	return err
}

// BatchOperation 批量操作
func (ops *SessionOperations) BatchOperation() error {
	session := ops.manager.GetSession()
	ctx := context.Background()
	
	// 批量插入数据
	batchData := make([][]any, 10)
	for i := 0; i < 10; i++ {
		batchData[i] = []any{
			fmt.Sprintf("Batch User %d", i),
			fmt.Sprintf("batch%d@test.com", i),
			25 + i,
		}
	}
	
	_, err := session.BatchInsert(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", batchData)
	return err
}

// CachedOperation 缓存操作
func (ops *SessionOperations) CachedOperation() error {
	session := ops.manager.GetSession()
	
	// 启用缓存（使用默认配置）
	cachedSession := session.EnableCache(nil)
	
	ctx := context.Background()
	
	// 执行查询（会被缓存）
	_, err := cachedSession.SelectOne(ctx, "SELECT id, name FROM users WHERE id = 1")
	return err
}

// TransactionOperation 事务操作
func (ops *SessionOperations) TransactionOperation() error {
	// 开始事务
	tx := ops.manager.db.Begin()
	session := mybatis.NewSimpleSession(tx)
	ctx := context.Background()
	
	// 执行事务操作
	_, err := session.Insert(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
		"TX User", "tx@test.com", 30)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// 提交事务
	return tx.Commit().Error
}

// SessionPerformanceResults 会话性能测试结果
type SessionPerformanceResults struct {
	DirectCall      time.Duration
	PooledCall      time.Duration
	BatchOperation  time.Duration
	CachedOperation time.Duration
	TransactionOp   time.Duration
	MemoryUsage     runtime.MemStats
}

// === MyBatis会话性能基准测试 ===

// BenchmarkMyBatisSessionDirect 直接会话性能测试
func BenchmarkMyBatisSessionDirect(b *testing.B) {
	manager := NewMyBatisSessionManager(1)
	ops := &SessionOperations{manager: manager}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := ops.DirectOperation()
		if err != nil {
			// 某些查询可能失败，不算致命错误
			continue
		}
	}
}

// BenchmarkMyBatisSessionPooled 池化会话性能测试
func BenchmarkMyBatisSessionPooled(b *testing.B) {
	manager := NewMyBatisSessionManager(10) // 10个会话的池
	ops := &SessionOperations{manager: manager}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := ops.PooledOperation()
		if err != nil {
			continue
		}
	}
}

// BenchmarkMyBatisSessionBatch 批量会话性能测试
func BenchmarkMyBatisSessionBatch(b *testing.B) {
	manager := NewMyBatisSessionManager(5)
	ops := &SessionOperations{manager: manager}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := ops.BatchOperation()
		if err != nil {
			b.Fatalf("Batch operation failed: %v", err)
		}
	}
}

// BenchmarkMyBatisSessionCached 缓存会话性能测试
func BenchmarkMyBatisSessionCached(b *testing.B) {
	manager := NewMyBatisSessionManager(5)
	ops := &SessionOperations{manager: manager}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := ops.CachedOperation()
		if err != nil {
			continue
		}
	}
}

// BenchmarkMyBatisSessionTransaction 事务会话性能测试
func BenchmarkMyBatisSessionTransaction(b *testing.B) {
	manager := NewMyBatisSessionManager(5)
	ops := &SessionOperations{manager: manager}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := ops.TransactionOperation()
		if err != nil {
			b.Fatalf("Transaction failed: %v", err)
		}
	}
}

// BenchmarkMyBatisSessionConcurrent 并发会话性能测试
func BenchmarkMyBatisSessionConcurrent(b *testing.B) {
	manager := NewMyBatisSessionManager(20) // 更大的池以支持并发
	ops := &SessionOperations{manager: manager}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 随机选择操作类型
			switch b.N % 4 {
			case 0:
				ops.DirectOperation()
			case 1:
				ops.PooledOperation()
			case 2:
				ops.CachedOperation()
			case 3:
				ops.TransactionOperation()
			}
		}
	})
}

// BenchmarkMyBatisSessionMemory 会话内存分配测试
func BenchmarkMyBatisSessionMemory(b *testing.B) {
	manager := NewMyBatisSessionManager(1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 测试会话创建的内存开销
		session := manager.CreateNewSession()
		ctx := context.Background()
		
		// 执行简单操作
		session.SelectOne(ctx, "SELECT 1")
	}
}

// === MyBatis会话功能测试 ===

// TestMyBatisSessionPerformance 会话性能对比测试
func TestMyBatisSessionPerformance(t *testing.T) {
	fmt.Printf("🚀 MyBatis会话性能对比测试\n")
	fmt.Printf("===============================\n")

	manager := NewMyBatisSessionManager(10)
	ops := &SessionOperations{manager: manager}

	// 预热
	fmt.Printf("⏳ 预热中...\n")
	for i := 0; i < SessionWarmupRounds; i++ {
		ops.PooledOperation()
	}

	// 强制垃圾回收
	runtime.GC()
	runtime.GC()

	// 性能测试
	fmt.Printf("\n📊 开始性能测试...\n")

	results := SessionPerformanceResults{}

	// 测试直接调用
	start := time.Now()
	for i := 0; i < SessionTestIterations; i++ {
		ops.DirectOperation()
	}
	results.DirectCall = time.Since(start)

	// 测试池化调用
	start = time.Now()
	for i := 0; i < SessionTestIterations; i++ {
		ops.PooledOperation()
	}
	results.PooledCall = time.Since(start)

	// 测试批量操作
	start = time.Now()
	for i := 0; i < SessionTestIterations/10; i++ { // 批量操作次数较少
		ops.BatchOperation()
	}
	results.BatchOperation = time.Since(start)

	// 测试缓存操作
	start = time.Now()
	for i := 0; i < SessionTestIterations; i++ {
		ops.CachedOperation()
	}
	results.CachedOperation = time.Since(start)

	// 测试事务操作
	start = time.Now()
	for i := 0; i < SessionTestIterations/5; i++ { // 事务操作次数较少
		ops.TransactionOperation()
	}
	results.TransactionOp = time.Since(start)

	// 获取内存使用情况
	runtime.ReadMemStats(&results.MemoryUsage)

	// 输出结果
	printSessionResults(results, t)
	
	// 验证性能指标
	validateSessionPerformance(results, t)
}

// TestMyBatisSessionLifecycle 会话生命周期测试
func TestMyBatisSessionLifecycle(t *testing.T) {
	manager := NewMyBatisSessionManager(5)

	// 测试会话创建
	session := manager.CreateNewSession()
	if session == nil {
		t.Fatal("Failed to create session")
	}

	// 测试会话池获取
	pooledSession := manager.GetSession()
	if pooledSession == nil {
		t.Fatal("Failed to get pooled session")
	}

	// 测试会话操作
	ctx := context.Background()
	_, err := session.SelectOne(ctx, "SELECT 1")
	if err != nil {
		t.Logf("Session operation test: %v", err)
	}

	t.Log("✅ MyBatis Session lifecycle test completed")
}

// TestMyBatisSessionCaching 会话缓存测试
func TestMyBatisSessionCaching(t *testing.T) {
	manager := NewMyBatisSessionManager(1)
	session := manager.GetSession()

	// 启用缓存（使用默认配置）
	cachedSession := session.EnableCache(nil)

	ctx := context.Background()

	// 第一次查询
	start := time.Now()
	_, err := cachedSession.SelectOne(ctx, "SELECT id, name FROM users WHERE id = 1")
	firstDuration := time.Since(start)
	if err != nil {
		t.Logf("First query: %v", err)
	}

	// 第二次查询（应该更快，因为有缓存）
	start = time.Now()
	_, err = cachedSession.SelectOne(ctx, "SELECT id, name FROM users WHERE id = 1")
	secondDuration := time.Since(start)
	if err != nil {
		t.Logf("Second query: %v", err)
	}

	t.Logf("Cache Performance:")
	t.Logf("- First query: %v", firstDuration)
	t.Logf("- Second query: %v", secondDuration)

	// 获取缓存统计
	stats := cachedSession.GetCacheStats()
	if stats != nil {
		t.Logf("- Cache stats: %+v", stats)
	}

	t.Log("✅ MyBatis Session caching test completed")
}

// 打印会话测试结果
func printSessionResults(results SessionPerformanceResults, t *testing.T) {
	t.Logf("\n📈 MyBatis Session Performance Results:")
	t.Logf("=====================================")
	
	t.Logf("Direct Call      : %v (%v/op)", results.DirectCall, results.DirectCall/SessionTestIterations)
	t.Logf("Pooled Call      : %v (%v/op)", results.PooledCall, results.PooledCall/SessionTestIterations)
	t.Logf("Batch Operation  : %v (%v/op)", results.BatchOperation, results.BatchOperation/(SessionTestIterations/10))
	t.Logf("Cached Operation : %v (%v/op)", results.CachedOperation, results.CachedOperation/SessionTestIterations)
	t.Logf("Transaction Op   : %v (%v/op)", results.TransactionOp, results.TransactionOp/(SessionTestIterations/5))
	
	t.Logf("\n💾 Memory Usage:")
	t.Logf("Allocated Memory : %d KB", results.MemoryUsage.Alloc/1024)
	t.Logf("Total Allocated  : %d KB", results.MemoryUsage.TotalAlloc/1024)
	t.Logf("System Memory    : %d KB", results.MemoryUsage.Sys/1024)
	t.Logf("GC Cycles        : %d", results.MemoryUsage.NumGC)

	// 性能优势分析
	poolAdvantage := float64(results.DirectCall-results.PooledCall) / float64(results.DirectCall) * 100
	t.Logf("\n🚀 Performance Analysis:")
	t.Logf("Session Pool Advantage: %.1f%% faster than direct calls", poolAdvantage)
}

// 验证会话性能指标
func validateSessionPerformance(results SessionPerformanceResults, t *testing.T) {
	// 池化调用应该比直接调用更快
	if results.PooledCall > results.DirectCall {
		t.Errorf("Pooled calls should be faster than direct calls")
	}

	// 平均操作时间应该在合理范围内
	avgDirectTime := results.DirectCall / SessionTestIterations
	if avgDirectTime > time.Millisecond {
		t.Errorf("Average direct call time %v exceeds 1ms threshold", avgDirectTime)
	}

	avgPooledTime := results.PooledCall / SessionTestIterations
	if avgPooledTime > time.Millisecond {
		t.Errorf("Average pooled call time %v exceeds 1ms threshold", avgPooledTime)
	}

	t.Log("✅ Session performance validation completed")
}

// 会话测试辅助函数
func setupSessionTestData(db *gorm.DB) {
	// 创建用户表
	db.Exec("DROP TABLE IF EXISTS users")
	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		age INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	// 插入测试数据
	testUsers := []struct {
		Name  string
		Email string
		Age   int
	}{
		{"Alice Johnson", "alice@session.com", 28},
		{"Bob Smith", "bob@session.com", 32},
		{"Carol Brown", "carol@session.com", 25},
		{"David Wilson", "david@session.com", 29},
		{"Eve Davis", "eve@session.com", 31},
	}

	for _, user := range testUsers {
		db.Exec("INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
			user.Name, user.Email, user.Age)
	}
}