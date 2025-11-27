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

// TestAnnotationLongTermStability 长期稳定性测试
func TestAnnotationLongTermStability(t *testing.T) {
	fmt.Println("\n=== 🏃‍♂️ 长期稳定性测试 (LongTermStability) ===")

	// 创建测试器
	tester, err := NewAnnotationStressTester("/tmp/annotation_longterm.db")
	if err != nil {
		t.Fatalf("创建测试器失败: %v", err)
	}
	defer tester.Close()

	// 测试配置
	testDuration := 30 * time.Second  // 缩短测试时间到30秒
	concurrency := 20                 // 减少并发数
	batchSize := 100                  // 批量大小

	fmt.Printf("测试配置: 持续时间=%v, 并发数=%d, 批量大小=%d\n", 
		testDuration, concurrency, batchSize)

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var (
		totalOps    int64
		successOps  int64
		failedOps   int64
		totalLatency int64
		peakMemory  uint64
	)

	// 内存监控协程
	go func() {
		ticker := time.NewTicker(1 * time.Second)
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

	// 启动并发工作协程
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			batchCounter := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// 执行批量操作
					for j := 0; j < batchSize; j++ {
						user := &TestUser{
							Username: fmt.Sprintf("longterm_user_%d_%d", workerID, batchCounter*batchSize+j),
							Email:    fmt.Sprintf("longterm_%d_%d@test.com", workerID, batchCounter*batchSize+j),
							Age:      20 + ((batchCounter*batchSize + j) % 50),
							Status:   "active",
						}

						atomic.AddInt64(&totalOps, 1)
						opStart := time.Now()

						id, err := tester.annotationSess.Insert(context.Background(), user)
						latency := time.Since(opStart)
						atomic.AddInt64(&totalLatency, latency.Microseconds())

						if err == nil && id > 0 {
							atomic.AddInt64(&successOps, 1)

							// 执行查询验证
							_, err = tester.annotationSess.SelectByID(context.Background(), &TestUser{}, id)
							if err == nil {
								atomic.AddInt64(&successOps, 1)
							} else {
								atomic.AddInt64(&failedOps, 1)
							}
						} else {
							atomic.AddInt64(&failedOps, 1)
						}
						atomic.AddInt64(&totalOps, 1)
					}
					batchCounter++

					// 短暂休息避免过载
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 计算统计数据
	totalOperations := atomic.LoadInt64(&totalOps)
	successful := atomic.LoadInt64(&successOps)
	failed := atomic.LoadInt64(&failedOps)
	avgLatency := float64(atomic.LoadInt64(&totalLatency)) / float64(totalOperations)

	successRate := float64(successful) / float64(totalOperations) * 100
	qps := float64(totalOperations) / duration.Seconds()

	fmt.Printf("\n📊 长期稳定性测试结果:\n")
	fmt.Printf("   测试持续时间: %v\n", duration)
	fmt.Printf("   总操作数: %d\n", totalOperations)
	fmt.Printf("   成功操作: %d\n", successful)
	fmt.Printf("   失败操作: %d\n", failed)
	fmt.Printf("   成功率: %.2f%%\n", successRate)
	fmt.Printf("   平均QPS: %.2f\n", qps)
	fmt.Printf("   平均延迟: %.2fµs\n", avgLatency)
	fmt.Printf("   峰值内存: %.2fMB\n", float64(peakMemory)/(1024*1024))

	// 评估测试结果
	grade := "C"
	if successRate >= 98 && qps >= 1000 {
		grade = "A+"
	} else if successRate >= 95 && qps >= 800 {
		grade = "A"
	} else if successRate >= 90 && qps >= 500 {
		grade = "B+"
	} else if successRate >= 80 {
		grade = "B"
	}

	fmt.Printf("   性能评级: %s\n", grade)

	// 验证基本要求
	if successRate < 85 {
		t.Errorf("成功率过低: %.2f%% (最低要求: 85%%)", successRate)
	}

	if qps < 300 {
		t.Errorf("QPS过低: %.2f (最低要求: 300)", qps)
	}

	fmt.Println("✅ 长期稳定性测试完成")
}