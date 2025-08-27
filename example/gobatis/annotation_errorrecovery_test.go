package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestAnnotationErrorHandlingRecovery 错误处理恢复测试
func TestAnnotationErrorHandlingRecovery(t *testing.T) {
	fmt.Println("\n=== 🛠️ 错误处理恢复测试 (ErrorHandlingRecovery) ===")

	// 创建测试器
	tester, err := NewAnnotationStressTester("/tmp/annotation_recovery.db")
	if err != nil {
		t.Fatalf("创建测试器失败: %v", err)
	}
	defer tester.Close()

	fmt.Println("测试场景包括: 重复键错误, 并发错误处理, 超时恢复")

	// 错误恢复测试结果
	results := &ErrorRecoveryResults{
		TestStartTime: time.Now(),
	}

	// 1. 测试重复键错误处理
	testDuplicateKeyHandling(t, tester, results)

	// 2. 测试并发错误处理  
	testConcurrentErrorHandling(t, tester, results)

	// 3. 测试超时恢复
	testTimeoutRecovery(t, tester, results)

	// 计算综合评估
	results.TestEndTime = time.Now()
	results.TotalTestDuration = results.TestEndTime.Sub(results.TestStartTime)

	printErrorRecoveryResults(results)
	evaluateErrorRecoveryResults(t, results)

	fmt.Println("✅ 错误处理恢复测试完成")
}

// ErrorRecoveryResults 错误恢复测试结果
type ErrorRecoveryResults struct {
	TestStartTime       time.Time
	TestEndTime         time.Time
	TotalTestDuration   time.Duration
	
	// 重复键错误测试
	DuplicateKeyTests     int64
	DuplicateKeyErrors    int64
	DuplicateKeyRecovered int64
	
	// 并发错误测试
	ConcurrentOps       int64
	ConcurrentErrors    int64
	ConcurrentRecovered int64
	
	// 超时恢复测试
	TimeoutOps          int64
	TimeoutErrors       int64
	TimeoutRecovered    int64
}

// testDuplicateKeyHandling 测试重复键错误处理
func testDuplicateKeyHandling(t *testing.T, tester *AnnotationStressTester, results *ErrorRecoveryResults) {
	fmt.Printf("\n🔑 测试重复键错误处理...\n")

	// 先插入一个用户
	originalUser := &TestUser{
		Username: "duplicate_test_user",
		Email:    "duplicate@test.com",
		Age:      25,
		Status:   "active",
	}

	id, err := tester.annotationSess.Insert(context.Background(), originalUser)
	if err != nil {
		t.Errorf("插入原始用户失败: %v", err)
		return
	}

	atomic.AddInt64(&results.DuplicateKeyTests, 1)

	// 尝试插入重复用户 
	duplicateUser := &TestUser{
		Username: "duplicate_test_user", // 相同用户名
		Email:    "duplicate2@test.com",
		Age:      30,
		Status:   "active",
	}

	_, err = tester.annotationSess.Insert(context.Background(), duplicateUser)
	if err != nil {
		atomic.AddInt64(&results.DuplicateKeyErrors, 1)
		fmt.Printf("   ✅ 正确检测到重复键错误: %v\n", err)
		
		// 尝试用不同的用户名恢复
		duplicateUser.Username = "duplicate_test_user_fixed"
		_, err = tester.annotationSess.Insert(context.Background(), duplicateUser)
		if err == nil {
			atomic.AddInt64(&results.DuplicateKeyRecovered, 1)
			fmt.Printf("   ✅ 成功从重复键错误恢复\n")
		} else {
			fmt.Printf("   ❌ 恢复失败: %v\n", err)
		}
	} else {
		fmt.Printf("   ❌ 未能检测到重复键错误 (id: %d)\n", id)
	}
}

// testConcurrentErrorHandling 测试并发错误处理
func testConcurrentErrorHandling(t *testing.T, tester *AnnotationStressTester, results *ErrorRecoveryResults) {
	fmt.Printf("\n🔄 测试并发错误处理...\n")
	
	concurrency := 20
	operationsPerGoroutine := 10
	
	var wg sync.WaitGroup
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				atomic.AddInt64(&results.ConcurrentOps, 1)
				
				user := &TestUser{
					Username: fmt.Sprintf("concurrent_user_%d_%d", workerID, j),
					Email:    fmt.Sprintf("concurrent_%d_%d@test.com", workerID, j),
					Age:      20 + (j % 50),
					Status:   "active",
				}
				
				_, err := tester.annotationSess.Insert(context.Background(), user)
				if err != nil {
					atomic.AddInt64(&results.ConcurrentErrors, 1)
					
					// 尝试恢复 - 重试一次
					time.Sleep(10 * time.Millisecond)
					_, err = tester.annotationSess.Insert(context.Background(), user)
					if err == nil {
						atomic.AddInt64(&results.ConcurrentRecovered, 1)
					}
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	fmt.Printf("   📊 并发错误处理结果: 总操作=%d, 错误=%d, 恢复=%d\n",
		results.ConcurrentOps, results.ConcurrentErrors, results.ConcurrentRecovered)
}

// testTimeoutRecovery 测试超时恢复
func testTimeoutRecovery(t *testing.T, tester *AnnotationStressTester, results *ErrorRecoveryResults) {
	fmt.Printf("\n⏰ 测试超时恢复...\n")
	
	// 使用极短的超时来模拟超时情况
	shortTimeout := 1 * time.Microsecond
	
	for i := 0; i < 10; i++ {
		atomic.AddInt64(&results.TimeoutOps, 1)
		
		ctx, cancel := context.WithTimeout(context.Background(), shortTimeout)
		user := &TestUser{
			Username: fmt.Sprintf("timeout_user_%d", i),
			Email:    fmt.Sprintf("timeout_%d@test.com", i),
			Age:      25,
			Status:   "active",
		}
		
		_, err := tester.annotationSess.Insert(ctx, user)
		cancel()
		
		if err != nil {
			atomic.AddInt64(&results.TimeoutErrors, 1)
			
			// 使用正常超时重试
			_, err = tester.annotationSess.Insert(context.Background(), user)
			if err == nil {
				atomic.AddInt64(&results.TimeoutRecovered, 1)
			}
		}
	}
	
	fmt.Printf("   📊 超时恢复结果: 总操作=%d, 超时错误=%d, 恢复=%d\n",
		results.TimeoutOps, results.TimeoutErrors, results.TimeoutRecovered)
}

// printErrorRecoveryResults 打印错误恢复测试结果
func printErrorRecoveryResults(results *ErrorRecoveryResults) {
	fmt.Printf("\n📊 错误处理恢复测试综合结果:\n")
	fmt.Printf("   测试总时长: %v\n", results.TotalTestDuration)
	
	fmt.Printf("\n🔑 重复键错误处理:\n")
	fmt.Printf("   测试次数: %d\n", results.DuplicateKeyTests)
	fmt.Printf("   检测到错误: %d\n", results.DuplicateKeyErrors)
	fmt.Printf("   成功恢复: %d\n", results.DuplicateKeyRecovered)
	if results.DuplicateKeyErrors > 0 {
		fmt.Printf("   恢复率: %.1f%%\n", float64(results.DuplicateKeyRecovered)/float64(results.DuplicateKeyErrors)*100)
	}
	
	fmt.Printf("\n🔄 并发错误处理:\n")
	fmt.Printf("   总并发操作: %d\n", results.ConcurrentOps)
	fmt.Printf("   并发错误: %d\n", results.ConcurrentErrors)
	fmt.Printf("   并发恢复: %d\n", results.ConcurrentRecovered)
	if results.ConcurrentErrors > 0 {
		fmt.Printf("   并发恢复率: %.1f%%\n", float64(results.ConcurrentRecovered)/float64(results.ConcurrentErrors)*100)
	}
	
	fmt.Printf("\n⏰ 超时恢复:\n")
	fmt.Printf("   超时操作测试: %d\n", results.TimeoutOps)
	fmt.Printf("   超时错误: %d\n", results.TimeoutErrors)
	fmt.Printf("   超时恢复: %d\n", results.TimeoutRecovered)
	if results.TimeoutErrors > 0 {
		fmt.Printf("   超时恢复率: %.1f%%\n", float64(results.TimeoutRecovered)/float64(results.TimeoutErrors)*100)
	}
}

// evaluateErrorRecoveryResults 评估错误恢复测试结果
func evaluateErrorRecoveryResults(t *testing.T, results *ErrorRecoveryResults) {
	// 计算总体恢复率
	totalErrors := results.DuplicateKeyErrors + results.ConcurrentErrors + results.TimeoutErrors
	totalRecovered := results.DuplicateKeyRecovered + results.ConcurrentRecovered + results.TimeoutRecovered
	
	overallRecoveryRate := float64(0)
	if totalErrors > 0 {
		overallRecoveryRate = float64(totalRecovered) / float64(totalErrors) * 100
	}
	
	// 评级
	grade := "C"
	if overallRecoveryRate >= 90 {
		grade = "A+"
	} else if overallRecoveryRate >= 80 {
		grade = "A"
	} else if overallRecoveryRate >= 70 {
		grade = "B+"
	} else if overallRecoveryRate >= 60 {
		grade = "B"
	}
	
	fmt.Printf("\n🎯 错误恢复综合评估:\n")
	fmt.Printf("   总体恢复率: %.1f%%\n", overallRecoveryRate)
	fmt.Printf("   错误处理评级: %s\n", grade)
	
	// 验证基本要求
	if overallRecoveryRate < 50 {
		t.Errorf("总体恢复率过低: %.1f%% (最低要求: 50%%)", overallRecoveryRate)
	}
}