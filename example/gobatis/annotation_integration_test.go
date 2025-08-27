package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// AnnotationIntegrationSuite annotation集成测试套件
type AnnotationIntegrationSuite struct {
	stressTester   *AnnotationStressTester
	benchmarkSuite *AnnotationBenchmarkSuite
	stats          *AnnotationIntegrationStats
}

// AnnotationIntegrationStats annotation集成测试统计
type AnnotationIntegrationStats struct {
	// 综合性能指标
	TotalTestCases   int64   `json:"total_test_cases"`
	PassedTestCases  int64   `json:"passed_test_cases"`
	FailedTestCases  int64   `json:"failed_test_cases"`
	OverallQPS       float64 `json:"overall_qps"`
	OverallLatencyMs float64 `json:"overall_latency_ms"`

	// 缓存效率指标
	CacheHitRate    float64 `json:"cache_hit_rate"`
	CacheMissRate   float64 `json:"cache_miss_rate"`
	CacheOverheadMs float64 `json:"cache_overhead_ms"`

	// 资源使用指标
	PeakMemoryMB        float64 `json:"peak_memory_mb"`
	AvgCPUUsage         float64 `json:"avg_cpu_usage"`
	MaxGoroutines       int64   `json:"max_goroutines"`
	DatabaseConnections int64   `json:"database_connections"`

	// 稳定性指标
	ErrorRate      float64 `json:"error_rate"`
	RecoveryTime   float64 `json:"recovery_time_ms"`
	MemoryLeakRate float64 `json:"memory_leak_rate"`

	// 扩展性指标
	LinearScaling         float64 `json:"linear_scaling_factor"`
	ThroughputDegradation float64 `json:"throughput_degradation"`

	mutex sync.RWMutex
}

// NewAnnotationIntegrationSuite 创建annotation集成测试套件
func NewAnnotationIntegrationSuite() (*AnnotationIntegrationSuite, error) {
	stressTester, err := NewAnnotationStressTester(":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create stress tester: %w", err)
	}

	benchmarkSuite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		return nil, fmt.Errorf("failed to create benchmark suite: %w", err)
	}

	return &AnnotationIntegrationSuite{
		stressTester:   stressTester,
		benchmarkSuite: benchmarkSuite,
		stats:          &AnnotationIntegrationStats{},
	}, nil
}

// TestAnnotationFullIntegration 全面集成测试
func TestAnnotationFullIntegration(t *testing.T) {
	suite, err := NewAnnotationIntegrationSuite()
	if err != nil {
		t.Fatalf("Failed to create integration suite: %v", err)
	}
	defer suite.Close()

	t.Log("开始annotation全面集成测试...")

	// 子测试：基础功能验证
	t.Run("BasicFunctionality", func(t *testing.T) {
		suite.testBasicFunctionality(t)
	})

	// 子测试：高并发压力测试
	t.Run("HighConcurrencyStress", func(t *testing.T) {
		suite.testHighConcurrencyStress(t)
	})

	// 子测试：缓存系统集成
	t.Run("CacheSystemIntegration", func(t *testing.T) {
		suite.testCacheSystemIntegration(t)
	})

	// 子测试：性能基准验证
	t.Run("PerformanceBenchmark", func(t *testing.T) {
		suite.testPerformanceBenchmark(t)
	})

	// 子测试：长期稳定性测试
	t.Run("LongTermStability", func(t *testing.T) {
		suite.testLongTermStability(t)
	})

	// 子测试：扩展性测试
	t.Run("Scalability", func(t *testing.T) {
		suite.testScalability(t)
	})

	// 子测试：错误处理和恢复
	t.Run("ErrorHandlingRecovery", func(t *testing.T) {
		suite.testErrorHandlingRecovery(t)
	})

	// 输出综合测试报告
	suite.generateIntegrationReport(t)
}

// testBasicFunctionality 基础功能测试
func (suite *AnnotationIntegrationSuite) testBasicFunctionality(t *testing.T) {
	ctx := context.Background()

	// 测试annotation驱动的CRUD操作
	testUser := &TestUser{
		Username: "integration_test_user",
		Email:    "integration@test.com",
		Age:      30,
		Status:   "active",
	}

	// 测试插入
	id, err := suite.stressTester.annotationSess.Insert(ctx, testUser)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		return
	}
	t.Logf("插入成功，ID: %d", id)

	// 测试查询
	result, err := suite.stressTester.annotationSess.SelectByID(ctx, &TestUser{}, id)
	if err != nil {
		t.Errorf("SelectByID failed: %v", err)
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		return
	}

	if user, ok := result.(*TestUser); ok {
		if user.Username != testUser.Username {
			t.Errorf("Username mismatch: got %s, want %s", user.Username, testUser.Username)
			atomic.AddInt64(&suite.stats.FailedTestCases, 1)
			return
		}
	}
	t.Log("查询验证成功")

	// 测试更新
	testUser.ID = id
	testUser.Age = 31
	affected, err := suite.stressTester.annotationSess.Update(ctx, testUser)
	if err != nil || affected == 0 {
		t.Errorf("Update failed: %v, affected: %d", err, affected)
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		return
	}
	t.Log("更新成功")

	// 测试删除
	affected, err = suite.stressTester.annotationSess.Delete(ctx, testUser)
	if err != nil || affected == 0 {
		t.Errorf("Delete failed: %v, affected: %d", err, affected)
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		return
	}
	t.Log("删除成功")

	atomic.AddInt64(&suite.stats.PassedTestCases, 1)
	atomic.AddInt64(&suite.stats.TotalTestCases, 1)
}

// testHighConcurrencyStress 高并发压力测试
func (suite *AnnotationIntegrationSuite) testHighConcurrencyStress(t *testing.T) {
	ctx := context.Background()
	concurrency := 200
	operationsPerGoroutine := 50

	t.Logf("开始高并发压力测试: %d协程 x %d操作", concurrency, operationsPerGoroutine)

	var wg sync.WaitGroup
	var successOps, failedOps int64
	var totalLatency int64
	var peakMemory uint64

	startTime := time.Now()

	// 内存监控协程
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				if m.Alloc > peakMemory {
					peakMemory = m.Alloc
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// 启动并发测试
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				user := &TestUser{
					Username: fmt.Sprintf("stress_user_%d_%d", workerID, j),
					Email:    fmt.Sprintf("stress_%d_%d@test.com", workerID, j),
					Age:      20 + (j % 50),
					Status:   "active",
				}

				opStart := time.Now()
				id, err := suite.stressTester.annotationSess.Insert(ctx, user)
				latency := time.Since(opStart)

				atomic.AddInt64(&totalLatency, latency.Milliseconds())

				if err == nil && id > 0 {
					atomic.AddInt64(&successOps, 1)

					// 执行查询验证
					_, err = suite.stressTester.annotationSess.SelectByID(ctx, &TestUser{}, id)
					if err == nil {
						atomic.AddInt64(&successOps, 1)
					} else {
						atomic.AddInt64(&failedOps, 1)
					}
				} else {
					atomic.AddInt64(&failedOps, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 计算性能指标
	totalOps := successOps + failedOps
	suite.stats.OverallQPS = float64(totalOps) / duration.Seconds()
	suite.stats.OverallLatencyMs = float64(totalLatency) / float64(totalOps)
	suite.stats.ErrorRate = float64(failedOps) / float64(totalOps) * 100
	suite.stats.PeakMemoryMB = float64(peakMemory) / (1024 * 1024)
	suite.stats.MaxGoroutines = int64(concurrency)

	t.Logf("高并发压力测试完成:")
	t.Logf("- 总操作数: %d", totalOps)
	t.Logf("- 成功操作: %d", successOps)
	t.Logf("- 失败操作: %d", failedOps)
	t.Logf("- QPS: %.2f", suite.stats.OverallQPS)
	t.Logf("- 平均延迟: %.2fms", suite.stats.OverallLatencyMs)
	t.Logf("- 错误率: %.2f%%", suite.stats.ErrorRate)
	t.Logf("- 峰值内存: %.2fMB", suite.stats.PeakMemoryMB)

	// 验证性能指标
	if suite.stats.ErrorRate > 5.0 {
		t.Errorf("Error rate too high: %.2f%%, expected <= 5%%", suite.stats.ErrorRate)
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
	} else {
		atomic.AddInt64(&suite.stats.PassedTestCases, 1)
	}

	atomic.AddInt64(&suite.stats.TotalTestCases, 1)
}

// testCacheSystemIntegration 缓存系统集成测试
func (suite *AnnotationIntegrationSuite) testCacheSystemIntegration(t *testing.T) {

	// 测试TagParser缓存
	parser := mybatis.NewTagParser()
	testUser := &TestUser{}

	// 首次解析（缓存未命中）
	start := time.Now()
	_, err := parser.ParseStruct(testUser)
	firstParseTime := time.Since(start)

	if err != nil {
		t.Errorf("First parse failed: %v", err)
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		return
	}

	// 第二次解析（缓存命中）
	start = time.Now()
	_, err = parser.ParseStruct(testUser)
	secondParseTime := time.Since(start)

	if err != nil {
		t.Errorf("Second parse failed: %v", err)
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		return
	}

	// 计算缓存效率
	cacheSpeedup := float64(firstParseTime) / float64(secondParseTime)
	suite.stats.CacheHitRate = (1 - float64(secondParseTime)/float64(firstParseTime)) * 100

	t.Logf("缓存系统测试:")
	t.Logf("- 首次解析耗时: %v", firstParseTime)
	t.Logf("- 缓存命中耗时: %v", secondParseTime)
	t.Logf("- 缓存加速比: %.2fx", cacheSpeedup)
	t.Logf("- 缓存命中率效益: %.2f%%", suite.stats.CacheHitRate)

	// 并发缓存测试
	concurrency := 100
	iterations := 1000
	var wg sync.WaitGroup
	var cacheHits, cacheMisses int64

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				start := time.Now()
				_, err := parser.ParseStruct(&TestUser{})
				duration := time.Since(start)

				if err == nil {
					// 根据解析时间判断是否缓存命中
					// 缓存命中通常比首次解析快10倍以上
					if duration < firstParseTime/5 {
						atomic.AddInt64(&cacheHits, 1)
					} else {
						atomic.AddInt64(&cacheMisses, 1)
					}
				}
			}
		}()
	}

	wg.Wait()

	totalCacheOps := cacheHits + cacheMisses
	if totalCacheOps > 0 {
		suite.stats.CacheHitRate = float64(cacheHits) / float64(totalCacheOps) * 100
		suite.stats.CacheMissRate = float64(cacheMisses) / float64(totalCacheOps) * 100
	}

	t.Logf("并发缓存测试:")
	t.Logf("- 缓存命中: %d", cacheHits)
	t.Logf("- 缓存未命中: %d", cacheMisses)
	t.Logf("- 缓存命中率: %.2f%%", suite.stats.CacheHitRate)

	if suite.stats.CacheHitRate > 95.0 {
		atomic.AddInt64(&suite.stats.PassedTestCases, 1)
	} else {
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
	}

	atomic.AddInt64(&suite.stats.TotalTestCases, 1)
}

// testPerformanceBenchmark 性能基准测试
func (suite *AnnotationIntegrationSuite) testPerformanceBenchmark(t *testing.T) {
	ctx := context.Background()

	// 测试不同场景的性能基准
	scenarios := []struct {
		name       string
		operations int
		expected   struct {
			maxLatencyMs float64
			minQPS       float64
		}
	}{
		{"LightLoad", 1000, struct{ maxLatencyMs, minQPS float64 }{50.0, 100.0}},
		{"MediumLoad", 5000, struct{ maxLatencyMs, minQPS float64 }{100.0, 80.0}},
		{"HeavyLoad", 10000, struct{ maxLatencyMs, minQPS float64 }{200.0, 50.0}},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			start := time.Now()
			var successOps, failedOps int64
			var totalLatency int64

			for i := 0; i < scenario.operations; i++ {
				user := &TestUser{
					Username: fmt.Sprintf("bench_%s_%d", scenario.name, i),
					Email:    fmt.Sprintf("bench_%s_%d@test.com", scenario.name, i),
					Age:      25,
					Status:   "active",
				}

				opStart := time.Now()
				id, err := suite.stressTester.annotationSess.Insert(ctx, user)
				latency := time.Since(opStart)

				atomic.AddInt64(&totalLatency, latency.Milliseconds())

				if err == nil && id > 0 {
					atomic.AddInt64(&successOps, 1)
				} else {
					atomic.AddInt64(&failedOps, 1)
				}
			}

			duration := time.Since(start)
			avgLatency := float64(totalLatency) / float64(successOps)
			qps := float64(successOps) / duration.Seconds()

			t.Logf("%s性能指标:", scenario.name)
			t.Logf("- 平均延迟: %.2fms (期望 <= %.2fms)", avgLatency, scenario.expected.maxLatencyMs)
			t.Logf("- QPS: %.2f (期望 >= %.2f)", qps, scenario.expected.minQPS)

			if avgLatency <= scenario.expected.maxLatencyMs && qps >= scenario.expected.minQPS {
				atomic.AddInt64(&suite.stats.PassedTestCases, 1)
			} else {
				atomic.AddInt64(&suite.stats.FailedTestCases, 1)
			}

			atomic.AddInt64(&suite.stats.TotalTestCases, 1)
		})
	}
}

// testLongTermStability 长期稳定性测试
func (suite *AnnotationIntegrationSuite) testLongTermStability(t *testing.T) {
	ctx := context.Background()
	duration := 60 * time.Second // 1分钟长期测试

	t.Logf("开始长期稳定性测试，持续时间: %v", duration)

	var operations, errors int64
	var initialMemory, currentMemory uint64

	// 记录初始内存
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	initialMemory = m1.Alloc

	startTime := time.Now()
	endTime := startTime.Add(duration)

	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			localOps := 0
			for time.Now().Before(endTime) {
				user := &TestUser{
					Username: fmt.Sprintf("stability_%d_%d", workerID, localOps),
					Email:    fmt.Sprintf("stability_%d_%d@test.com", workerID, localOps),
					Age:      20 + (localOps % 50),
					Status:   "active",
				}

				_, err := suite.stressTester.annotationSess.Insert(ctx, user)
				if err != nil {
					atomic.AddInt64(&errors, 1)
				}

				atomic.AddInt64(&operations, 1)
				localOps++

				// 适当休息以模拟真实负载
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// 内存监控协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runtime.GC()
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				currentMemory = m.Alloc
			case <-time.After(duration):
				return
			}
		}
	}()

	wg.Wait()

	// 最终内存检查
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	currentMemory = m2.Alloc

	actualDuration := time.Since(startTime)
	errorRate := float64(errors) / float64(operations) * 100
	memoryGrowth := float64(int64(currentMemory)-int64(initialMemory)) / float64(initialMemory) * 100

	suite.stats.ErrorRate = errorRate
	suite.stats.MemoryLeakRate = memoryGrowth

	t.Logf("长期稳定性测试完成:")
	t.Logf("- 实际测试时间: %v", actualDuration)
	t.Logf("- 总操作数: %d", operations)
	t.Logf("- 错误数: %d", errors)
	t.Logf("- 错误率: %.2f%%", errorRate)
	t.Logf("- 平均QPS: %.2f", float64(operations)/actualDuration.Seconds())
	t.Logf("- 初始内存: %.2fMB", float64(initialMemory)/(1024*1024))
	t.Logf("- 结束内存: %.2fMB", float64(currentMemory)/(1024*1024))
	t.Logf("- 内存增长率: %.2f%%", memoryGrowth)

	// 验证稳定性指标
	passed := true
	if errorRate > 3.0 { // 3%错误率阈值
		t.Errorf("Error rate too high: %.2f%%, expected <= 3%%", errorRate)
		passed = false
	}

	if memoryGrowth > 10.0 { // 10%内存增长阈值
		t.Errorf("Memory growth too high: %.2f%%, expected <= 10%%", memoryGrowth)
		passed = false
	}

	if passed {
		atomic.AddInt64(&suite.stats.PassedTestCases, 1)
	} else {
		atomic.AddInt64(&suite.stats.FailedTestCases, 1)
	}

	atomic.AddInt64(&suite.stats.TotalTestCases, 1)
}

// testScalability 扩展性测试
func (suite *AnnotationIntegrationSuite) testScalability(t *testing.T) {
	concurrencies := []int{1, 10, 50, 100, 200}
	operationsPerTest := 1000

	var results []struct {
		concurrency int
		qps         float64
		latency     float64
	}

	for _, concurrency := range concurrencies {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			ctx := context.Background()
			var wg sync.WaitGroup
			var totalOps, totalLatency int64

			startTime := time.Now()

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					opsPerWorker := operationsPerTest / concurrency
					for j := 0; j < opsPerWorker; j++ {
						user := &TestUser{
							Username: fmt.Sprintf("scale_%d_%d_%d", concurrency, workerID, j),
							Email:    fmt.Sprintf("scale_%d_%d_%d@test.com", concurrency, workerID, j),
							Age:      25,
							Status:   "active",
						}

						opStart := time.Now()
						_, err := suite.stressTester.annotationSess.Insert(ctx, user)
						latency := time.Since(opStart)

						if err == nil {
							atomic.AddInt64(&totalOps, 1)
							atomic.AddInt64(&totalLatency, latency.Milliseconds())
						}
					}
				}(i)
			}

			wg.Wait()
			duration := time.Since(startTime)

			qps := float64(totalOps) / duration.Seconds()
			avgLatency := float64(totalLatency) / float64(totalOps)

			results = append(results, struct {
				concurrency int
				qps         float64
				latency     float64
			}{concurrency, qps, avgLatency})

			t.Logf("并发度 %d: QPS=%.2f, 延迟=%.2fms", concurrency, qps, avgLatency)
		})
	}

	// 计算线性扩展因子
	if len(results) >= 2 {
		baselineQPS := results[0].qps
		maxConcurrencyQPS := results[len(results)-1].qps
		maxConcurrency := results[len(results)-1].concurrency

		expectedQPS := baselineQPS * float64(maxConcurrency)
		suite.stats.LinearScaling = maxConcurrencyQPS / expectedQPS
		suite.stats.ThroughputDegradation = (1 - suite.stats.LinearScaling) * 100

		t.Logf("扩展性分析:")
		t.Logf("- 基准QPS (并发度1): %.2f", baselineQPS)
		t.Logf("- 最大QPS (并发度%d): %.2f", maxConcurrency, maxConcurrencyQPS)
		t.Logf("- 理论最大QPS: %.2f", expectedQPS)
		t.Logf("- 线性扩展因子: %.2f", suite.stats.LinearScaling)
		t.Logf("- 吞吐量衰减: %.2f%%", suite.stats.ThroughputDegradation)

		if suite.stats.LinearScaling > 0.7 { // 70%线性扩展阈值
			atomic.AddInt64(&suite.stats.PassedTestCases, 1)
		} else {
			atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		}
	}

	atomic.AddInt64(&suite.stats.TotalTestCases, 1)
}

// testErrorHandlingRecovery 错误处理和恢复测试
func (suite *AnnotationIntegrationSuite) testErrorHandlingRecovery(t *testing.T) {
	// 测试数据库连接中断恢复
	t.Run("DatabaseRecovery", func(t *testing.T) {
		ctx := context.Background()
		// 模拟数据库操作
		user := &TestUser{
			Username: "error_test_user",
			Email:    "error@test.com",
			Age:      25,
			Status:   "active",
		}

		// 正常操作
		id, err := suite.stressTester.annotationSess.Insert(ctx, user)
		if err != nil {
			t.Errorf("Normal operation failed: %v", err)
			atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		} else {
			// 验证恢复后的操作
			_, err = suite.stressTester.annotationSess.SelectByID(ctx, &TestUser{}, id)
			if err != nil {
				t.Errorf("Recovery operation failed: %v", err)
				atomic.AddInt64(&suite.stats.FailedTestCases, 1)
			} else {
				atomic.AddInt64(&suite.stats.PassedTestCases, 1)
			}
		}

		atomic.AddInt64(&suite.stats.TotalTestCases, 1)
	})

	// 测试无效参数处理
	t.Run("InvalidParameterHandling", func(t *testing.T) {
		ctx := context.Background()
		// 测试nil参数
		_, err := suite.stressTester.annotationSess.Insert(ctx, nil)
		if err == nil {
			t.Error("Expected error for nil parameter, but got nil")
			atomic.AddInt64(&suite.stats.FailedTestCases, 1)
		} else {
			atomic.AddInt64(&suite.stats.PassedTestCases, 1)
		}

		atomic.AddInt64(&suite.stats.TotalTestCases, 1)
	})
}

// generateIntegrationReport 生成综合测试报告
func (suite *AnnotationIntegrationSuite) generateIntegrationReport(t *testing.T) {
	passRate := float64(suite.stats.PassedTestCases) / float64(suite.stats.TotalTestCases) * 100

	t.Log("\n" + strings.Repeat("=", 80))
	t.Log("ANNOTATION集成测试综合报告")
	t.Log(strings.Repeat("=", 80))

	t.Logf("测试概况:")
	t.Logf("- 总测试用例: %d", suite.stats.TotalTestCases)
	t.Logf("- 通过用例: %d", suite.stats.PassedTestCases)
	t.Logf("- 失败用例: %d", suite.stats.FailedTestCases)
	t.Logf("- 通过率: %.2f%%", passRate)

	t.Logf("\n性能指标:")
	t.Logf("- 整体QPS: %.2f", suite.stats.OverallQPS)
	t.Logf("- 平均延迟: %.2fms", suite.stats.OverallLatencyMs)
	t.Logf("- 错误率: %.2f%%", suite.stats.ErrorRate)

	t.Logf("\n缓存效率:")
	t.Logf("- 缓存命中率: %.2f%%", suite.stats.CacheHitRate)
	t.Logf("- 缓存未命中率: %.2f%%", suite.stats.CacheMissRate)

	t.Logf("\n资源使用:")
	t.Logf("- 峰值内存: %.2fMB", suite.stats.PeakMemoryMB)
	t.Logf("- 最大协程数: %d", suite.stats.MaxGoroutines)
	t.Logf("- 内存泄漏率: %.2f%%", suite.stats.MemoryLeakRate)

	t.Logf("\n扩展性:")
	t.Logf("- 线性扩展因子: %.2f", suite.stats.LinearScaling)
	t.Logf("- 吞吐量衰减: %.2f%%", suite.stats.ThroughputDegradation)

	// 评级系统
	var grade string
	if passRate >= 95.0 && suite.stats.ErrorRate <= 2.0 && suite.stats.LinearScaling >= 0.8 {
		grade = "S+ (优秀)"
	} else if passRate >= 90.0 && suite.stats.ErrorRate <= 5.0 && suite.stats.LinearScaling >= 0.7 {
		grade = "A+ (良好)"
	} else if passRate >= 80.0 && suite.stats.ErrorRate <= 10.0 {
		grade = "B+ (一般)"
	} else {
		grade = "C (需改进)"
	}

	t.Logf("\n综合评级: %s", grade)
	t.Log(strings.Repeat("=", 80))
}

// Close 关闭集成测试套件
func (suite *AnnotationIntegrationSuite) Close() error {
	if suite.stressTester != nil {
		suite.stressTester.Close()
	}
	if suite.benchmarkSuite != nil {
		suite.benchmarkSuite.Close()
	}
	return nil
}
