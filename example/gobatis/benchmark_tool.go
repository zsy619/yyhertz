// Package main 性能基准测试工具
//
// 提供专业的数据库操作性能测试工具，包括：
// 1. 吞吐量测试
// 2. 并发性能测试
// 3. 内存使用监控
// 4. 延迟分析
// 5. 压力测试
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// BenchmarkConfig 基准测试配置
type BenchmarkConfig struct {
	DatabasePath     string
	ConcurrentUsers  int
	TestDuration     time.Duration
	WarmupDuration   time.Duration
	OperationMix     OperationMix
	DataSetSize      int
	ReportInterval   time.Duration
	EnableProfiling  bool
}

// OperationMix 操作混合比例
type OperationMix struct {
	ReadPercent   int // 读操作百分比
	WritePercent  int // 写操作百分比
	UpdatePercent int // 更新操作百分比
	DeletePercent int // 删除操作百分比
}

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	TotalOperations   int64
	SuccessOperations int64
	FailedOperations  int64
	TotalDuration     time.Duration
	AvgLatency        time.Duration
	MinLatency        time.Duration
	MaxLatency        time.Duration
	P50Latency        time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	ThroughputOPS     float64
	MemoryUsageMB     float64
	CPUUsage          float64
}

// LatencyRecord 延迟记录
type LatencyRecord struct {
	Latency   time.Duration
	Timestamp time.Time
	Operation string
	Success   bool
}

// BenchmarkTool 基准测试工具
type BenchmarkTool struct {
	config          BenchmarkConfig
	db              *gorm.DB
	simpleSession   mybatis.SimpleSession
	xmlSession      mybatis.XMLSession
	latencyRecords  []LatencyRecord
	latencyMutex    sync.Mutex
	operationCount  int64
	successCount    int64
	failedCount     int64
	startTime       time.Time
	stopChan        chan struct{}
}

// NewBenchmarkTool 创建基准测试工具
func NewBenchmarkTool(config BenchmarkConfig) (*BenchmarkTool, error) {
	// 创建数据库连接
	db, err := gorm.Open(sqlite.Open(config.DatabasePath), &gorm.Config{
		Logger: logger.New(
			log.Default(),
			logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Error, // 只记录错误，减少日志开销
				Colorful:      false,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 针对并发测试优化连接池
	sqlDB.SetMaxOpenConns(config.ConcurrentUsers * 2)
	sqlDB.SetMaxIdleConns(config.ConcurrentUsers)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 创建会话
	simpleSession := mybatis.NewSimpleSession(db)
	xmlSession := mybatis.NewXMLMapper(db)

	// 加载XML映射
	err = xmlSession.LoadMapperXMLFromString(getBenchmarkMapperXML())
	if err != nil {
		return nil, fmt.Errorf("failed to load XML mapper: %w", err)
	}

	return &BenchmarkTool{
		config:         config,
		db:             db,
		simpleSession:  simpleSession,
		xmlSession:     xmlSession,
		latencyRecords: make([]LatencyRecord, 0, 100000), // 预分配空间
		stopChan:       make(chan struct{}),
	}, nil
}

// RunBenchmark 运行基准测试
func (bt *BenchmarkTool) RunBenchmark() (*BenchmarkResult, error) {
	fmt.Printf("🚀 开始基准测试...\n")
	fmt.Printf("配置: %d并发用户, %v测试时长, %v预热时长\n",
		bt.config.ConcurrentUsers, bt.config.TestDuration, bt.config.WarmupDuration)

	// 1. 准备测试数据
	if err := bt.setupTestData(); err != nil {
		return nil, fmt.Errorf("setup test data failed: %w", err)
	}

	// 2. 预热
	if bt.config.WarmupDuration > 0 {
		fmt.Printf("🔥 预热阶段 (%v)...\n", bt.config.WarmupDuration)
		if err := bt.warmup(); err != nil {
			return nil, fmt.Errorf("warmup failed: %w", err)
		}
	}

	// 3. 重置计数器和记录
	atomic.StoreInt64(&bt.operationCount, 0)
	atomic.StoreInt64(&bt.successCount, 0)
	atomic.StoreInt64(&bt.failedCount, 0)
	bt.latencyRecords = bt.latencyRecords[:0]

	// 4. 启动监控
	go bt.startMonitoring()

	// 5. 开始基准测试
	bt.startTime = time.Now()
	if err := bt.runConcurrentTest(); err != nil {
		return nil, fmt.Errorf("concurrent test failed: %w", err)
	}

	// 6. 计算结果
	return bt.calculateResult(), nil
}

// setupTestData 准备测试数据
func (bt *BenchmarkTool) setupTestData() error {
	fmt.Printf("📊 准备测试数据 (%d条记录)...\n", bt.config.DataSetSize)

	// 创建表
	err := bt.db.AutoMigrate(&User{})
	if err != nil {
		return err
	}

	// 清空现有数据
	bt.db.Exec("DELETE FROM users")

	// 批量插入测试数据
	batchSize := 1000
	for i := 0; i < bt.config.DataSetSize; i += batchSize {
		end := i + batchSize
		if end > bt.config.DataSetSize {
			end = bt.config.DataSetSize
		}

		users := make([]*User, end-i)
		for j := 0; j < end-i; j++ {
			idx := i + j
			users[j] = &User{
				Name:   fmt.Sprintf("BenchUser_%d", idx),
				Email:  fmt.Sprintf("bench_%d@test.com", idx),
				Age:    rand.Intn(50) + 20,
				Status: []string{"active", "inactive", "pending"}[rand.Intn(3)],
				Phone:  fmt.Sprintf("1380013%04d", idx),
			}
		}

		if err := bt.db.CreateInBatches(users, batchSize).Error; err != nil {
			return fmt.Errorf("batch insert failed: %w", err)
		}

		if (i+batchSize)%5000 == 0 {
			fmt.Printf("已插入 %d/%d 条记录\n", i+batchSize, bt.config.DataSetSize)
		}
	}

	fmt.Printf("✅ 测试数据准备完成\n")
	return nil
}

// warmup 预热
func (bt *BenchmarkTool) warmup() error {
	ctx, cancel := context.WithTimeout(context.Background(), bt.config.WarmupDuration)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < bt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bt.runWorker(ctx, true) // warmup mode
		}()
	}

	wg.Wait()
	fmt.Printf("✅ 预热完成\n")
	return nil
}

// runConcurrentTest 运行并发测试
func (bt *BenchmarkTool) runConcurrentTest() error {
	fmt.Printf("⚡ 开始并发测试...\n")

	ctx, cancel := context.WithTimeout(context.Background(), bt.config.TestDuration)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < bt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			bt.runWorker(ctx, false) // benchmark mode
		}(i)
	}

	wg.Wait()
	close(bt.stopChan)
	fmt.Printf("✅ 并发测试完成\n")
	return nil
}

// runWorker 工作协程
func (bt *BenchmarkTool) runWorker(ctx context.Context, warmupMode bool) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			bt.executeRandomOperation(ctx, warmupMode)
		}
	}
}

// executeRandomOperation 执行随机操作
func (bt *BenchmarkTool) executeRandomOperation(ctx context.Context, warmupMode bool) {
	// 根据配置的操作混合比例选择操作
	rand := rand.Intn(100)
	var operation string
	var err error
	
	start := time.Now()
	
	switch {
	case rand < bt.config.OperationMix.ReadPercent:
		operation = "READ"
		err = bt.executeRead(ctx)
	case rand < bt.config.OperationMix.ReadPercent+bt.config.OperationMix.WritePercent:
		operation = "WRITE"
		err = bt.executeWrite(ctx)
	case rand < bt.config.OperationMix.ReadPercent+bt.config.OperationMix.WritePercent+bt.config.OperationMix.UpdatePercent:
		operation = "UPDATE"
		err = bt.executeUpdate(ctx)
	default:
		operation = "DELETE"
		err = bt.executeDelete(ctx)
	}
	
	latency := time.Since(start)
	
	if !warmupMode {
		// 记录延迟
		bt.latencyMutex.Lock()
		bt.latencyRecords = append(bt.latencyRecords, LatencyRecord{
			Latency:   latency,
			Timestamp: time.Now(),
			Operation: operation,
			Success:   err == nil,
		})
		bt.latencyMutex.Unlock()
		
		// 更新计数器
		atomic.AddInt64(&bt.operationCount, 1)
		if err == nil {
			atomic.AddInt64(&bt.successCount, 1)
		} else {
			atomic.AddInt64(&bt.failedCount, 1)
		}
	}
}

// executeRead 执行读操作
func (bt *BenchmarkTool) executeRead(ctx context.Context) error {
	operations := []func(context.Context) error{
		bt.readById,
		bt.readByStatus,
		bt.readWithPagination,
		bt.readCount,
	}
	
	return operations[rand.Intn(len(operations))](ctx)
}

// executeWrite 执行写操作
func (bt *BenchmarkTool) executeWrite(ctx context.Context) error {
	return bt.insertUser(ctx)
}

// executeUpdate 执行更新操作
func (bt *BenchmarkTool) executeUpdate(ctx context.Context) error {
	return bt.updateUser(ctx)
}

// executeDelete 执行删除操作
func (bt *BenchmarkTool) executeDelete(ctx context.Context) error {
	return bt.deleteUser(ctx)
}

// 具体操作方法
func (bt *BenchmarkTool) readById(ctx context.Context) error {
	id := rand.Intn(bt.config.DataSetSize) + 1
	_, err := bt.simpleSession.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", id)
	return err
}

func (bt *BenchmarkTool) readByStatus(ctx context.Context) error {
	status := []string{"active", "inactive", "pending"}[rand.Intn(3)]
	_, err := bt.simpleSession.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 10", status)
	return err
}

func (bt *BenchmarkTool) readWithPagination(ctx context.Context) error {
	page := rand.Intn(100) + 1
	_, err := bt.simpleSession.SelectPage(ctx, "SELECT * FROM users WHERE status = 'active'", mybatis.PageRequest{
		Page: page,
		Size: 20,
	})
	return err
}

func (bt *BenchmarkTool) readCount(ctx context.Context) error {
	_, err := bt.simpleSession.SelectOne(ctx, "SELECT COUNT(*) FROM users WHERE age > ?", rand.Intn(50)+20)
	return err
}

func (bt *BenchmarkTool) insertUser(ctx context.Context) error {
	id := rand.Intn(1000000) + bt.config.DataSetSize
	name := fmt.Sprintf("TestUser_%d", id)
	email := fmt.Sprintf("test_%d@bench.com", id)
	age := rand.Intn(50) + 20
	status := []string{"active", "inactive", "pending"}[rand.Intn(3)]
	
	_, err := bt.simpleSession.Insert(ctx,
		"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
		name, email, age, status)
	return err
}

func (bt *BenchmarkTool) updateUser(ctx context.Context) error {
	id := rand.Intn(bt.config.DataSetSize) + 1
	newAge := rand.Intn(50) + 20
	_, err := bt.simpleSession.Update(ctx, "UPDATE users SET age = ? WHERE id = ?", newAge, id)
	return err
}

func (bt *BenchmarkTool) deleteUser(ctx context.Context) error {
	id := rand.Intn(bt.config.DataSetSize) + 1
	_, err := bt.simpleSession.Delete(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

// startMonitoring 启动监控
func (bt *BenchmarkTool) startMonitoring() {
	ticker := time.NewTicker(bt.config.ReportInterval)
	defer ticker.Stop()

	var lastOpCount int64
	lastTime := bt.startTime

	for {
		select {
		case <-bt.stopChan:
			return
		case now := <-ticker.C:
			currentOpCount := atomic.LoadInt64(&bt.operationCount)
			successCount := atomic.LoadInt64(&bt.successCount)
			failedCount := atomic.LoadInt64(&bt.failedCount)

			duration := now.Sub(lastTime)
			opsInInterval := currentOpCount - lastOpCount
			currentThroughput := float64(opsInInterval) / duration.Seconds()

			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memoryMB := float64(m.Alloc) / 1024 / 1024

			fmt.Printf("📊 [%v] 操作数: %d, 成功: %d, 失败: %d, 吞吐量: %.2f ops/s, 内存: %.2f MB\n",
				now.Sub(bt.startTime).Truncate(time.Second),
				currentOpCount, successCount, failedCount,
				currentThroughput, memoryMB)

			lastOpCount = currentOpCount
			lastTime = now
		}
	}
}

// calculateResult 计算测试结果
func (bt *BenchmarkTool) calculateResult() *BenchmarkResult {
	totalOps := atomic.LoadInt64(&bt.operationCount)
	successOps := atomic.LoadInt64(&bt.successCount)
	failedOps := atomic.LoadInt64(&bt.failedCount)
	duration := time.Since(bt.startTime)

	// 计算延迟统计
	bt.latencyMutex.Lock()
	latencies := make([]time.Duration, len(bt.latencyRecords))
	for i, record := range bt.latencyRecords {
		latencies[i] = record.Latency
	}
	bt.latencyMutex.Unlock()

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var avgLatency, minLatency, maxLatency, p50, p95, p99 time.Duration
	if len(latencies) > 0 {
		var totalLatency time.Duration
		for _, lat := range latencies {
			totalLatency += lat
		}
		avgLatency = totalLatency / time.Duration(len(latencies))
		minLatency = latencies[0]
		maxLatency = latencies[len(latencies)-1]
		p50 = latencies[len(latencies)*50/100]
		p95 = latencies[len(latencies)*95/100]
		p99 = latencies[len(latencies)*99/100]
	}

	// 获取内存使用
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryMB := float64(m.Alloc) / 1024 / 1024

	throughput := float64(totalOps) / duration.Seconds()

	return &BenchmarkResult{
		TotalOperations:   totalOps,
		SuccessOperations: successOps,
		FailedOperations:  failedOps,
		TotalDuration:     duration,
		AvgLatency:        avgLatency,
		MinLatency:        minLatency,
		MaxLatency:        maxLatency,
		P50Latency:        p50,
		P95Latency:        p95,
		P99Latency:        p99,
		ThroughputOPS:     throughput,
		MemoryUsageMB:     memoryMB,
	}
}

// PrintResult 打印测试结果
func (bt *BenchmarkTool) PrintResult(result *BenchmarkResult) {
	fmt.Print("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("🎯 基准测试结果报告\n")
	fmt.Print(strings.Repeat("=", 80) + "\n")
	
	fmt.Printf("📈 基础指标:\n")
	fmt.Printf("  总操作数:     %d\n", result.TotalOperations)
	fmt.Printf("  成功操作数:   %d (%.2f%%)\n", result.SuccessOperations, 
		float64(result.SuccessOperations)/float64(result.TotalOperations)*100)
	fmt.Printf("  失败操作数:   %d (%.2f%%)\n", result.FailedOperations,
		float64(result.FailedOperations)/float64(result.TotalOperations)*100)
	fmt.Printf("  测试时长:     %v\n", result.TotalDuration)
	fmt.Printf("  吞吐量:       %.2f 操作/秒\n", result.ThroughputOPS)
	
	fmt.Printf("\n⏱️ 延迟统计:\n")
	fmt.Printf("  平均延迟:     %v\n", result.AvgLatency)
	fmt.Printf("  最小延迟:     %v\n", result.MinLatency)
	fmt.Printf("  最大延迟:     %v\n", result.MaxLatency)
	fmt.Printf("  P50延迟:      %v\n", result.P50Latency)
	fmt.Printf("  P95延迟:      %v\n", result.P95Latency)
	fmt.Printf("  P99延迟:      %v\n", result.P99Latency)
	
	fmt.Printf("\n💾 资源使用:\n")
	fmt.Printf("  内存使用:     %.2f MB\n", result.MemoryUsageMB)
	
	// 性能评级
	fmt.Printf("\n🏆 性能评级:\n")
	if result.ThroughputOPS > 10000 {
		fmt.Printf("  吞吐量评级:   🥇 优秀 (>10000 ops/s)\n")
	} else if result.ThroughputOPS > 5000 {
		fmt.Printf("  吞吐量评级:   🥈 良好 (>5000 ops/s)\n")
	} else if result.ThroughputOPS > 1000 {
		fmt.Printf("  吞吐量评级:   🥉 一般 (>1000 ops/s)\n")
	} else {
		fmt.Printf("  吞吐量评级:   ⚠️ 需要优化 (<1000 ops/s)\n")
	}
	
	if result.P95Latency < 10*time.Millisecond {
		fmt.Printf("  延迟评级:     🥇 优秀 (P95<10ms)\n")
	} else if result.P95Latency < 50*time.Millisecond {
		fmt.Printf("  延迟评级:     🥈 良好 (P95<50ms)\n")
	} else if result.P95Latency < 100*time.Millisecond {
		fmt.Printf("  延迟评级:     🥉 一般 (P95<100ms)\n")
	} else {
		fmt.Printf("  延迟评级:     ⚠️ 需要优化 (P95>100ms)\n")
	}
	
	fmt.Print(strings.Repeat("=", 80) + "\n")
}

// getBenchmarkMapperXML 获取基准测试用的XML映射
func getBenchmarkMapperXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="UserMapper">
    <select id="selectById" parameterType="int" resultType="map">
        SELECT * FROM users WHERE id = #{id}
    </select>
    
    <select id="selectByStatus" parameterType="string" resultType="map">
        SELECT * FROM users WHERE status = #{status} LIMIT 10
    </select>
</mapper>`
}

// 主函数示例 - 如何运行基准测试
func runBenchmarkExample() {
	config := BenchmarkConfig{
		DatabasePath:     "benchmark.db",
		ConcurrentUsers:  50,
		TestDuration:     2 * time.Minute,
		WarmupDuration:   30 * time.Second,
		DataSetSize:      10000,
		ReportInterval:   10 * time.Second,
		EnableProfiling:  true,
		OperationMix: OperationMix{
			ReadPercent:   70, // 70% 读操作
			WritePercent:  15, // 15% 写操作
			UpdatePercent: 10, // 10% 更新操作
			DeletePercent: 5,  // 5% 删除操作
		},
	}

	tool, err := NewBenchmarkTool(config)
	if err != nil {
		log.Fatal("创建基准测试工具失败:", err)
	}

	result, err := tool.RunBenchmark()
	if err != nil {
		log.Fatal("运行基准测试失败:", err)
	}

	tool.PrintResult(result)
}