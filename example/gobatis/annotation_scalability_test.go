package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestAnnotationScalability 可扩展性测试
func TestAnnotationScalability(t *testing.T) {
	fmt.Println("\n=== 📈 可扩展性测试 (Scalability) ===")

	// 创建测试器
	tester, err := NewAnnotationStressTester("/tmp/annotation_scalability.db")
	if err != nil {
		t.Fatalf("创建测试器失败: %v", err)
	}
	defer tester.Close()

	// 多个并发级别测试
	concurrencyLevels := []int{5, 10, 25, 50, 100}
	operationsPerGoroutine := 50
	testTimeout := 15 * time.Second

	results := make(map[int]ScalabilityResult)

	fmt.Printf("测试配置: 并发级别=%v, 每协程操作数=%d, 超时时间=%v\n", 
		concurrencyLevels, operationsPerGoroutine, testTimeout)

	for _, concurrency := range concurrencyLevels {
		fmt.Printf("\n🔄 测试并发级别: %d\n", concurrency)
		
		result := runScalabilityTest(t, tester, concurrency, operationsPerGoroutine, testTimeout)
		results[concurrency] = result

		fmt.Printf("   QPS: %.2f, 成功率: %.2f%%, 平均延迟: %.2fµs, 峰值内存: %.2fMB\n",
			result.QPS, result.SuccessRate, result.AvgLatency, result.PeakMemoryMB)
	}

	// 分析可扩展性指标
	fmt.Printf("\n📊 可扩展性分析结果:\n")
	
	baselineQPS := results[concurrencyLevels[0]].QPS
	baselineMemory := results[concurrencyLevels[0]].PeakMemoryMB
	
	var linearityScores []float64
	var efficiencyScores []float64
	
	for i, concurrency := range concurrencyLevels {
		result := results[concurrency]
		
		// 计算线性度 (实际QPS提升 vs 期望QPS提升)
		expectedQPS := baselineQPS * float64(concurrency) / float64(concurrencyLevels[0])
		linearity := result.QPS / expectedQPS * 100
		
		// 计算效率 (QPS/并发数)
		efficiency := result.QPS / float64(concurrency)
		
		// 计算内存扩展比例
		memoryScale := result.PeakMemoryMB / baselineMemory
		
		linearityScores = append(linearityScores, linearity)
		efficiencyScores = append(efficiencyScores, efficiency)
		
		fmt.Printf("   并发%d: 线性度=%.1f%%, 效率=%.1f ops/goroutine/s, 内存扩展=%.1fx\n",
			concurrency, linearity, efficiency, memoryScale)
		
		if i > 0 {
			// 检查是否存在性能衰减
			prevResult := results[concurrencyLevels[i-1]]
			qpsImprovement := (result.QPS - prevResult.QPS) / prevResult.QPS * 100
			
			if qpsImprovement < 10 && concurrency > concurrencyLevels[i-1]*2 {
				fmt.Printf("   ⚠️  检测到性能瓶颈: QPS提升仅%.1f%%\n", qpsImprovement)
			}
		}
	}
	
	// 计算整体可扩展性评分
	avgLinearity := average(linearityScores)
	avgEfficiency := average(efficiencyScores)
	maxConcurrency := concurrencyLevels[len(concurrencyLevels)-1]
	maxResult := results[maxConcurrency]
	
	fmt.Printf("\n🎯 综合可扩展性评估:\n")
	fmt.Printf("   平均线性度: %.1f%% (理想值: 100%%)\n", avgLinearity)
	fmt.Printf("   平均效率: %.1f ops/goroutine/s\n", avgEfficiency)
	fmt.Printf("   最大并发数: %d goroutines\n", maxConcurrency)
	fmt.Printf("   最高QPS: %.2f\n", maxResult.QPS)
	fmt.Printf("   最大并发成功率: %.2f%%\n", maxResult.SuccessRate)
	
	// 评级
	grade := "C"
	if avgLinearity >= 80 && maxResult.SuccessRate >= 95 && maxResult.QPS >= 2000 {
		grade = "A+"
	} else if avgLinearity >= 70 && maxResult.SuccessRate >= 90 && maxResult.QPS >= 1500 {
		grade = "A"
	} else if avgLinearity >= 60 && maxResult.SuccessRate >= 85 && maxResult.QPS >= 1000 {
		grade = "B+"
	} else if avgLinearity >= 50 && maxResult.SuccessRate >= 80 {
		grade = "B"
	}
	
	fmt.Printf("   可扩展性评级: %s\n", grade)
	
	// 验证基本要求
	if maxResult.SuccessRate < 80 {
		t.Errorf("最大并发成功率过低: %.2f%% (最低要求: 80%%)", maxResult.SuccessRate)
	}
	
	if avgLinearity < 50 {
		t.Errorf("平均线性度过低: %.1f%% (最低要求: 50%%)", avgLinearity)
	}

	fmt.Println("✅ 可扩展性测试完成")
}

// ScalabilityResult 可扩展性测试结果
type ScalabilityResult struct {
	Concurrency   int
	QPS          float64
	SuccessRate  float64
	AvgLatency   float64
	PeakMemoryMB float64
}

// runScalabilityTest 运行单个并发级别的测试
func runScalabilityTest(t *testing.T, tester *AnnotationStressTester, concurrency, operationsPerGoroutine int, timeout time.Duration) ScalabilityResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var (
		totalOps     int64
		successOps   int64
		failedOps    int64
		totalLatency int64
		peakMemory   uint64
	)

	// 内存监控
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

	startTime := time.Now()
	wg := sync.WaitGroup{}

	// 启动并发测试
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				user := &TestUser{
					Username: fmt.Sprintf("scale_user_%d_%d", workerID, j),
					Email:    fmt.Sprintf("scale_%d_%d@test.com", workerID, j),
					Age:      20 + (j % 50),
					Status:   "active",
				}

				atomic.AddInt64(&totalOps, 1)
				opStart := time.Now()

				id, err := tester.annotationSess.Insert(context.Background(), user)
				latency := time.Since(opStart)
				atomic.AddInt64(&totalLatency, latency.Microseconds())

				if err == nil && id > 0 {
					atomic.AddInt64(&successOps, 1)
				} else {
					atomic.AddInt64(&failedOps, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	totalOperations := atomic.LoadInt64(&totalOps)
	successful := atomic.LoadInt64(&successOps)
	avgLatency := float64(atomic.LoadInt64(&totalLatency)) / float64(totalOperations)

	return ScalabilityResult{
		Concurrency:   concurrency,
		QPS:          float64(totalOperations) / duration.Seconds(),
		SuccessRate:  float64(successful) / float64(totalOperations) * 100,
		AvgLatency:   avgLatency,
		PeakMemoryMB: float64(peakMemory) / (1024 * 1024),
	}
}

// average 计算平均值
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}