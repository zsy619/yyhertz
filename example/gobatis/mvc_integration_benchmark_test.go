package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MyBatis集成测试套件
type MyBatisIntegrationSuite struct {
	db           *gorm.DB
	simpleSession mybatis.SimpleSession
	xmlSession   mybatis.XMLSession
	multiSession mybatis.MultiDataSourceSession
}

// NewMyBatisIntegrationSuite 创建集成测试套件
func NewMyBatisIntegrationSuite() *MyBatisIntegrationSuite {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	setupIntegrationTestData(db)

	return &MyBatisIntegrationSuite{
		db:           db,
		simpleSession: mybatis.NewSimpleSession(db),
		xmlSession:   mybatis.NewXMLSession(db),
	}
}

// MyBatisWorkload 工作负载类型
type MyBatisWorkload int

const (
	SimpleQuery MyBatisWorkload = iota
	ComplexQuery
	BatchInsert
	Transaction
	CachedQuery
)

// WorkloadResult 工作负载结果
type WorkloadResult struct {
	WorkloadType MyBatisWorkload
	Operations   int
	Duration     time.Duration
	ErrorCount   int
	AvgLatency   time.Duration
	MaxLatency   time.Duration
	MinLatency   time.Duration
}

// ExecuteWorkload 执行工作负载
func (suite *MyBatisIntegrationSuite) ExecuteWorkload(workloadType MyBatisWorkload, operations int) WorkloadResult {
	result := WorkloadResult{
		WorkloadType: workloadType,
		Operations:   operations,
		MaxLatency:   0,
		MinLatency:   time.Hour, // 初始化为很大的值
	}

	start := time.Now()
	var totalLatency time.Duration

	for i := 0; i < operations; i++ {
		opStart := time.Now()
		
		var err error
		switch workloadType {
		case SimpleQuery:
			err = suite.executeSimpleQuery()
		case ComplexQuery:
			err = suite.executeComplexQuery()
		case BatchInsert:
			err = suite.executeBatchInsert()
		case Transaction:
			err = suite.executeTransaction()
		case CachedQuery:
			err = suite.executeCachedQuery()
		}

		opDuration := time.Since(opStart)
		totalLatency += opDuration

		if opDuration > result.MaxLatency {
			result.MaxLatency = opDuration
		}
		if opDuration < result.MinLatency {
			result.MinLatency = opDuration
		}

		if err != nil {
			result.ErrorCount++
		}
	}

	result.Duration = time.Since(start)
	result.AvgLatency = totalLatency / time.Duration(operations)

	return result
}

// 各种工作负载实现
func (suite *MyBatisIntegrationSuite) executeSimpleQuery() error {
	ctx := context.Background()
	_, err := suite.simpleSession.SelectOne(ctx, "SELECT id, name FROM users WHERE id = 1")
	return err
}

func (suite *MyBatisIntegrationSuite) executeComplexQuery() error {
	ctx := context.Background()
	_, err := suite.simpleSession.SelectList(ctx, 
		"SELECT u.id, u.name, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.id, u.name")
	return err
}

func (suite *MyBatisIntegrationSuite) executeBatchInsert() error {
	ctx := context.Background()
	batchData := [][]any{
		{"Batch User 1", "batch1@test.com", 25},
		{"Batch User 2", "batch2@test.com", 26},
		{"Batch User 3", "batch3@test.com", 27},
	}
	_, err := suite.simpleSession.BatchInsert(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", batchData)
	return err
}

func (suite *MyBatisIntegrationSuite) executeTransaction() error {
	tx := suite.db.Begin()
	txSession := mybatis.NewSimpleSession(tx)
	ctx := context.Background()

	// 执行多个操作
	_, err := txSession.Insert(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
		"TX User", "tx@test.com", 30)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = txSession.Update(ctx, "UPDATE users SET age = age + 1 WHERE id = 1")
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (suite *MyBatisIntegrationSuite) executeCachedQuery() error {
	// 使用简单会话，缓存功能可能需要配置
	ctx := context.Background()
	_, err := suite.simpleSession.SelectOne(ctx, "SELECT id, name FROM users WHERE id = 1")
	return err
}

// === MyBatis集成性能基准测试 ===

// BenchmarkMyBatisIntegrationSimple 简单查询集成测试
func BenchmarkMyBatisIntegrationSimple(b *testing.B) {
	suite := NewMyBatisIntegrationSuite()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suite.executeSimpleQuery()
	}
}

// BenchmarkMyBatisIntegrationComplex 复杂查询集成测试
func BenchmarkMyBatisIntegrationComplex(b *testing.B) {
	suite := NewMyBatisIntegrationSuite()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suite.executeComplexQuery()
	}
}

// BenchmarkMyBatisIntegrationBatch 批量操作集成测试
func BenchmarkMyBatisIntegrationBatch(b *testing.B) {
	suite := NewMyBatisIntegrationSuite()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suite.executeBatchInsert()
	}
}

// BenchmarkMyBatisIntegrationTransaction 事务集成测试
func BenchmarkMyBatisIntegrationTransaction(b *testing.B) {
	suite := NewMyBatisIntegrationSuite()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		suite.executeTransaction()
	}
}

// BenchmarkMyBatisIntegrationMixed 混合工作负载测试
func BenchmarkMyBatisIntegrationMixed(b *testing.B) {
	suite := NewMyBatisIntegrationSuite()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 按比例执行不同类型的操作
		switch i % 5 {
		case 0, 1: // 40% 简单查询
			suite.executeSimpleQuery()
		case 2: // 20% 复杂查询
			suite.executeComplexQuery()
		case 3: // 20% 批量插入
			suite.executeBatchInsert()
		case 4: // 20% 事务操作
			suite.executeTransaction()
		}
	}
}

// BenchmarkMyBatisIntegrationConcurrent 并发集成测试
func BenchmarkMyBatisIntegrationConcurrent(b *testing.B) {
	suite := NewMyBatisIntegrationSuite()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 并发执行各种操作
			switch b.N % 4 {
			case 0:
				suite.executeSimpleQuery()
			case 1:
				suite.executeComplexQuery()
			case 2:
				suite.executeBatchInsert()
			case 3:
				suite.executeTransaction()
			}
		}
	})
}

// BenchmarkMyBatisStressTest 压力测试
func BenchmarkMyBatisStressTest(b *testing.B) {
	suite := NewMyBatisIntegrationSuite()
	concurrency := runtime.NumCPU() * 4 // 4倍CPU核心数的并发

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	operations := b.N / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				suite.executeSimpleQuery()
			}
		}()
	}

	wg.Wait()
}

// === MyBatis集成功能测试 ===

// TestMyBatisIntegrationWorkloads 工作负载集成测试
func TestMyBatisIntegrationWorkloads(t *testing.T) {
	suite := NewMyBatisIntegrationSuite()
	
	fmt.Printf("🚀 MyBatis集成工作负载测试\n")
	fmt.Printf("==========================\n")

	// 测试各种工作负载
	workloads := []struct {
		name       string
		workload   MyBatisWorkload
		operations int
	}{
		{"简单查询", SimpleQuery, 1000},
		{"复杂查询", ComplexQuery, 100},
		{"批量插入", BatchInsert, 50},
		{"事务操作", Transaction, 100},
		{"缓存查询", CachedQuery, 1000},
	}

	results := make([]WorkloadResult, len(workloads))

	for i, wl := range workloads {
		t.Logf("执行 %s 工作负载...", wl.name)
		results[i] = suite.ExecuteWorkload(wl.workload, wl.operations)
		
		t.Logf("✅ %s 完成:", wl.name)
		t.Logf("  操作数: %d", results[i].Operations)
		t.Logf("  总耗时: %v", results[i].Duration)
		t.Logf("  平均延迟: %v", results[i].AvgLatency)
		t.Logf("  最大延迟: %v", results[i].MaxLatency)
		t.Logf("  错误数: %d", results[i].ErrorCount)
		t.Logf("  QPS: %.2f", float64(results[i].Operations)/results[i].Duration.Seconds())
	}

	// 分析结果
	analyzeWorkloadResults(results, t)
}

// TestMyBatisIntegrationStability 稳定性测试
func TestMyBatisIntegrationStability(t *testing.T) {
	suite := NewMyBatisIntegrationSuite()
	
	t.Log("🔄 MyBatis稳定性测试开始...")
	
	// 长时间运行测试
	duration := 30 * time.Second // 30秒稳定性测试
	end := time.Now().Add(duration)
	
	operationCount := 0
	errorCount := 0
	
	for time.Now().Before(end) {
		err := suite.executeSimpleQuery()
		operationCount++
		if err != nil {
			errorCount++
		}
		
		// 每1000次操作检查一次内存
		if operationCount%1000 == 0 {
			runtime.GC()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			t.Logf("操作数: %d, 内存使用: %d KB", operationCount, m.Alloc/1024)
		}
	}
	
	errorRate := float64(errorCount) / float64(operationCount) * 100
	qps := float64(operationCount) / duration.Seconds()
	
	t.Logf("📊 稳定性测试结果:")
	t.Logf("  测试时长: %v", duration)
	t.Logf("  总操作数: %d", operationCount)
	t.Logf("  错误数: %d", errorCount)
	t.Logf("  错误率: %.2f%%", errorRate)
	t.Logf("  QPS: %.2f", qps)
	
	// 验证稳定性指标
	if errorRate > 1.0 {
		t.Errorf("错误率 %.2f%% 超过阈值 1%%", errorRate)
	}
	
	if qps < 100 {
		t.Errorf("QPS %.2f 低于最低要求 100", qps)
	}
	
	t.Log("✅ MyBatis稳定性测试完成")
}

// TestMyBatisIntegrationConcurrency 并发测试
func TestMyBatisIntegrationConcurrency(t *testing.T) {
	suite := NewMyBatisIntegrationSuite()
	
	concurrency := runtime.NumCPU() * 2
	operationsPerGoroutine := 100
	
	t.Logf("🔀 MyBatis并发测试开始 (并发数: %d)", concurrency)
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	totalErrors := 0
	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			errors := 0
			for j := 0; j < operationsPerGoroutine; j++ {
				if err := suite.executeSimpleQuery(); err != nil {
					errors++
				}
			}
			
			mu.Lock()
			totalErrors += errors
			mu.Unlock()
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(start)
	
	totalOperations := concurrency * operationsPerGoroutine
	errorRate := float64(totalErrors) / float64(totalOperations) * 100
	qps := float64(totalOperations) / duration.Seconds()
	
	t.Logf("📊 并发测试结果:")
	t.Logf("  并发数: %d", concurrency)
	t.Logf("  总操作数: %d", totalOperations)
	t.Logf("  总耗时: %v", duration)
	t.Logf("  错误数: %d", totalErrors)
	t.Logf("  错误率: %.2f%%", errorRate)
	t.Logf("  并发QPS: %.2f", qps)
	
	// 验证并发性能
	if errorRate > 2.0 {
		t.Errorf("并发错误率 %.2f%% 超过阈值 2%%", errorRate)
	}
	
	expectedMinQPS := float64(concurrency) * 50 // 每个并发至少50 QPS
	if qps < expectedMinQPS {
		t.Errorf("并发QPS %.2f 低于预期 %.2f", qps, expectedMinQPS)
	}
	
	t.Log("✅ MyBatis并发测试完成")
}

// 分析工作负载结果
func analyzeWorkloadResults(results []WorkloadResult, t *testing.T) {
	t.Log("\n📈 工作负载分析:")
	t.Log("================")
	
	// 计算总体统计
	totalOps := 0
	totalDuration := time.Duration(0)
	totalErrors := 0
	
	for _, result := range results {
		totalOps += result.Operations
		totalDuration += result.Duration
		totalErrors += result.ErrorCount
	}
	
	avgQPS := float64(totalOps) / totalDuration.Seconds()
	overallErrorRate := float64(totalErrors) / float64(totalOps) * 100
	
	t.Logf("总操作数: %d", totalOps)
	t.Logf("总耗时: %v", totalDuration)
	t.Logf("平均QPS: %.2f", avgQPS)
	t.Logf("总体错误率: %.2f%%", overallErrorRate)
	
	// 找出性能最佳和最差的工作负载
	var bestQPS, worstQPS float64
	var bestWorkload, worstWorkload string
	
	workloadNames := []string{"简单查询", "复杂查询", "批量插入", "事务操作", "缓存查询"}
	
	for i, result := range results {
		qps := float64(result.Operations) / result.Duration.Seconds()
		
		if i == 0 || qps > bestQPS {
			bestQPS = qps
			bestWorkload = workloadNames[i]
		}
		
		if i == 0 || qps < worstQPS {
			worstQPS = qps
			worstWorkload = workloadNames[i]
		}
	}
	
	t.Logf("\n🏆 性能分析:")
	t.Logf("最佳工作负载: %s (QPS: %.2f)", bestWorkload, bestQPS)
	t.Logf("最慢工作负载: %s (QPS: %.2f)", worstWorkload, worstQPS)
	t.Logf("性能差异: %.1fx", bestQPS/worstQPS)
}

// 集成测试辅助函数
func setupIntegrationTestData(db *gorm.DB) {
	// 创建用户表
	db.Exec("DROP TABLE IF EXISTS users")
	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		age INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	
	// 创建订单表（用于复杂查询）
	db.Exec("DROP TABLE IF EXISTS orders")
	db.Exec(`CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		amount DECIMAL(10,2) NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id)
	)`)
	
	// 插入测试数据
	users := []struct {
		Name  string
		Email string
		Age   int
	}{
		{"Integration User 1", "int1@test.com", 25},
		{"Integration User 2", "int2@test.com", 30},
		{"Integration User 3", "int3@test.com", 35},
		{"Integration User 4", "int4@test.com", 28},
		{"Integration User 5", "int5@test.com", 32},
	}
	
	for _, user := range users {
		result := db.Exec("INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
			user.Name, user.Email, user.Age)
		if result.Error == nil {
			userID := result.Statement.Context.Value("last_insert_id")
			// 为每个用户创建一些订单
			for i := 0; i < 3; i++ {
				db.Exec("INSERT INTO orders (user_id, amount, status) VALUES (?, ?, ?)",
					userID, float64((i+1)*100), "completed")
			}
		}
	}
}