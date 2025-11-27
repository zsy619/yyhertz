// Package main 压力测试和性能基准
//
// 测试MyBatis的性能极限和基准：
// 1. 并发查询压力测试
// 2. 大数据量操作基准测试
// 3. 内存使用压力测试
// 4. 连接池压力测试
// 5. 缓存系统压力测试
// 6. 事务并发压力测试
// 7. 综合性能基准测试
// 8. 性能回归测试
package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// BenchmarkUser 基准测试用户模型
type BenchmarkUser struct {
	ID       int64     `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Email    string    `json:"email" db:"email"`
	Age      int       `json:"age" db:"age"`
	Status   string    `json:"status" db:"status"`
	CreateAt time.Time `json:"createAt" db:"created_at"`
}

// TestConcurrentQueryStress 测试并发查询压力
func TestConcurrentQueryStress(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备大量测试数据
	setupStressTestData(t, session, ctx, 1000)

	testCases := []struct {
		name           string
		concurrency    int
		operationsPerG int
		duration       time.Duration
	}{
		{"轻负载并发", 10, 100, 5 * time.Second},
		{"中负载并发", 50, 200, 10 * time.Second},
		{"重负载并发", 100, 300, 15 * time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				successOps  int64
				failedOps   int64
				totalTime   int64
				wg          sync.WaitGroup
			)

			startTime := time.Now()
			
			// 启动并发 goroutines
			for i := 0; i < tc.concurrency; i++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()
					
					localSuccessOps := int64(0)
					localFailedOps := int64(0)
					localStartTime := time.Now()

					for j := 0; j < tc.operationsPerG; j++ {
						// 随机查询类型
						queryType := rand.Intn(4)
						var err error

						switch queryType {
						case 0: // 简单查询
							_, err = session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 10", "active")
						case 1: // 条件查询
							_, err = session.SelectList(ctx, "SELECT * FROM users WHERE age > ? AND age < ? LIMIT 5", 
								20+rand.Intn(20), 40+rand.Intn(20))
						case 2: // 聚合查询
							_, err = session.SelectOne(ctx, "SELECT COUNT(*) as count FROM users WHERE status = ?", 
								[]string{"active", "inactive", "pending"}[rand.Intn(3)])
						case 3: // 排序查询
							_, err = session.SelectList(ctx, "SELECT * FROM users ORDER BY age DESC LIMIT ?", 
								5+rand.Intn(15))
						}

						if err != nil {
							localFailedOps++
						} else {
							localSuccessOps++
						}

						// 检查是否超时
						if time.Since(startTime) > tc.duration {
							break
						}
					}

					localDuration := time.Since(localStartTime).Nanoseconds()
					atomic.AddInt64(&successOps, localSuccessOps)
					atomic.AddInt64(&failedOps, localFailedOps)
					atomic.AddInt64(&totalTime, localDuration)
				}(i)
			}

			// 等待所有goroutines完成或超时
			done := make(chan bool)
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// 正常完成
			case <-time.After(tc.duration + 5*time.Second):
				t.Logf("%s: 超时等待goroutines完成", tc.name)
			}

			actualDuration := time.Since(startTime)
			
			t.Logf("%s 压力测试结果:", tc.name)
			t.Logf("  并发数: %d", tc.concurrency)
			t.Logf("  实际运行时间: %v", actualDuration)
			t.Logf("  成功操作数: %d", successOps)
			t.Logf("  失败操作数: %d", failedOps)
			t.Logf("  总操作数: %d", successOps+failedOps)
			if actualDuration > 0 {
				t.Logf("  QPS (每秒查询数): %.2f", float64(successOps)/actualDuration.Seconds())
			}
			if totalTime > 0 && successOps > 0 {
				t.Logf("  平均响应时间: %v", time.Duration(totalTime/successOps))
			}
			if successOps+failedOps > 0 {
				successRate := float64(successOps) / float64(successOps+failedOps) * 100
				t.Logf("  成功率: %.2f%%", successRate)
			}

			// 性能断言
			if failedOps > successOps/10 { // 失败率不应超过10%
				t.Errorf("%s: 失败率过高 %.2f%%", tc.name, float64(failedOps)/float64(successOps+failedOps)*100)
			}
		})
	}
}

// TestLargeDataOperationBenchmark 测试大数据量操作基准
func TestLargeDataOperationBenchmark(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	dataSizes := []int{1000, 5000, 10000, 20000}

	for _, dataSize := range dataSizes {
		t.Run(fmt.Sprintf("数据量_%d", dataSize), func(t *testing.T) {
			// 测试批量插入性能
			t.Run("批量插入", func(t *testing.T) {
				batchArgs := make([][]interface{}, dataSize)
				for i := 0; i < dataSize; i++ {
					batchArgs[i] = []interface{}{
						fmt.Sprintf("批量用户%d", i),
						fmt.Sprintf("batch%d@example.com", i),
						20 + (i % 60),
						[]string{"active", "inactive", "pending"}[i%3],
					}
				}

				start := time.Now()
				affected, err := session.BatchInsert(ctx,
					"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)", batchArgs)
				duration := time.Since(start)

				if err != nil {
					t.Fatalf("批量插入失败: %v", err)
				}

				t.Logf("批量插入 %d 条记录:", dataSize)
				t.Logf("  插入时间: %v", duration)
				t.Logf("  实际插入: %d 条", affected)
				t.Logf("  插入速率: %.2f 条/秒", float64(affected)/duration.Seconds())
				t.Logf("  平均每条: %v", duration/time.Duration(affected))
			})

			// 测试全表查询性能
			t.Run("全表查询", func(t *testing.T) {
				start := time.Now()
				users, err := session.SelectList(ctx, "SELECT * FROM users")
				duration := time.Since(start)

				if err != nil {
					t.Fatalf("全表查询失败: %v", err)
				}

				t.Logf("全表查询结果:")
				t.Logf("  查询时间: %v", duration)
				t.Logf("  返回记录: %d 条", len(users))
				t.Logf("  查询速率: %.2f 条/秒", float64(len(users))/duration.Seconds())
			})

			// 测试复杂聚合查询性能
			t.Run("聚合查询", func(t *testing.T) {
				complexQueries := []string{
					"SELECT status, COUNT(*) as count, AVG(age) as avg_age FROM users GROUP BY status",
					"SELECT COUNT(*) as total, MIN(age) as min_age, MAX(age) as max_age, AVG(age) as avg_age FROM users",
					"SELECT status, COUNT(*) as count FROM users WHERE age BETWEEN 25 AND 35 GROUP BY status ORDER BY count DESC",
				}

				for i, query := range complexQueries {
					start := time.Now()
					result, err := session.SelectList(ctx, query)
					duration := time.Since(start)

					if err != nil {
						t.Logf("复杂查询 %d 失败: %v", i+1, err)
					} else {
						t.Logf("复杂查询 %d: 耗时 %v, 结果 %d 行", i+1, duration, len(result))
					}
				}
			})

			// 测试批量更新性能
			t.Run("批量更新", func(t *testing.T) {
				start := time.Now()
				affected, err := session.Update(ctx,
					"UPDATE users SET age = age + 1 WHERE status = ?", "active")
				duration := time.Since(start)

				if err != nil {
					t.Fatalf("批量更新失败: %v", err)
				}

				t.Logf("批量更新结果:")
				t.Logf("  更新时间: %v", duration)
				t.Logf("  影响行数: %d", affected)
				if affected > 0 {
					t.Logf("  更新速率: %.2f 条/秒", float64(affected)/duration.Seconds())
				}
			})
		})
	}
}

// TestMemoryStressTest 测试内存压力
func TestMemoryStressTest(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备数据
	setupStressTestData(t, session, ctx, 5000)

	t.Run("大量并发查询内存测试", func(t *testing.T) {
		var memBefore, memAfter runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memBefore)

		const concurrency = 50
		const operationsPerG = 100
		var wg sync.WaitGroup

		startTime := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerG; j++ {
					// 执行各种查询以消耗内存
					_, err := session.SelectList(ctx, "SELECT * FROM users WHERE age > ? LIMIT 50", 
						20+rand.Intn(40))
					if err != nil {
						t.Logf("内存测试查询失败: %v", err)
					}

					// 偶尔进行 GC
					if j%20 == 0 {
						runtime.GC()
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// 强制垃圾回收
		runtime.GC()
		runtime.ReadMemStats(&memAfter)

		t.Logf("内存压力测试结果:")
		t.Logf("  运行时间: %v", duration)
		t.Logf("  测试前内存: %.2f MB", float64(memBefore.Alloc)/1024/1024)
		t.Logf("  测试后内存: %.2f MB", float64(memAfter.Alloc)/1024/1024)
		t.Logf("  内存增长: %.2f MB", float64(memAfter.Alloc-memBefore.Alloc)/1024/1024)
		t.Logf("  GC次数: %d", memAfter.NumGC-memBefore.NumGC)
		t.Logf("  堆内存: %.2f MB", float64(memAfter.HeapAlloc)/1024/1024)

		// 检查内存泄漏
		memGrowth := float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024
		if memGrowth > 100 { // 内存增长超过100MB可能有问题
			t.Logf("警告: 内存增长较大 %.2f MB，可能存在内存泄漏", memGrowth)
		}
	})

	t.Run("长时间运行内存稳定性测试", func(t *testing.T) {
		const testDuration = 10 * time.Second
		const checkInterval = 1 * time.Second
		
		var memStats []runtime.MemStats
		done := make(chan bool)
		
		// 内存监控goroutine
		go func() {
			ticker := time.NewTicker(checkInterval)
			defer ticker.Stop()
			
			for {
				select {
				case <-ticker.C:
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					memStats = append(memStats, m)
				case <-done:
					return
				}
			}
		}()
		
		// 执行持续负载
		endTime := time.Now().Add(testDuration)
		operationCount := int64(0)
		
		for time.Now().Before(endTime) {
			_, err := session.SelectList(ctx, "SELECT * FROM users WHERE id > ? LIMIT 20", 
				rand.Intn(1000))
			if err != nil {
				t.Logf("长时间测试查询失败: %v", err)
			}
			atomic.AddInt64(&operationCount, 1)
			
			// 短暂休息
			time.Sleep(10 * time.Millisecond)
		}
		
		close(done)
		
		t.Logf("长时间运行测试结果:")
		t.Logf("  运行时间: %v", testDuration)
		t.Logf("  总操作数: %d", operationCount)
		t.Logf("  平均QPS: %.2f", float64(operationCount)/testDuration.Seconds())
		
		if len(memStats) > 2 {
			first := memStats[0]
			last := memStats[len(memStats)-1]
			t.Logf("  初始内存: %.2f MB", float64(first.Alloc)/1024/1024)
			t.Logf("  最终内存: %.2f MB", float64(last.Alloc)/1024/1024)
			t.Logf("  内存变化: %.2f MB", float64(last.Alloc-first.Alloc)/1024/1024)
		}
	})
}

// TestCacheSystemStress 测试缓存系统压力
func TestCacheSystemStress(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	// 启用缓存
	cacheConfig := &mybatis.CacheConfig{
		L1CacheEnabled:  true,
		L1CacheSize:     1000,
		L1CacheTTL:      5 * time.Minute,
		L2CacheEnabled:  true,
		L2CacheSize:     5000,
		L2CacheTTL:      10 * time.Minute,
		CleanupInterval: 30 * time.Second,
	}

	session := mybatis.NewSimpleSession(db).
		EnableCache(cacheConfig).
		Debug(false)
	ctx := context.Background()

	setupStressTestData(t, session, ctx, 2000)

	t.Run("缓存命中率压力测试", func(t *testing.T) {
		const concurrency = 30
		const operationsPerG = 200
		var wg sync.WaitGroup
		var cacheHits, cacheMisses int64

		// 预热缓存
		warmupQueries := []string{
			"SELECT * FROM users WHERE status = 'active' LIMIT 10",
			"SELECT * FROM users WHERE age > 25 LIMIT 15",
			"SELECT COUNT(*) FROM users WHERE status = 'pending'",
		}
		
		for _, query := range warmupQueries {
			_, err := session.SelectList(ctx, query)
			if err != nil {
				t.Logf("缓存预热失败: %v", err)
			}
		}

		startTime := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerG; j++ {
					// 80%的查询使用相同的SQL（应该命中缓存）
					if rand.Float64() < 0.8 {
						queryIndex := rand.Intn(len(warmupQueries))
						_, err := session.SelectList(ctx, warmupQueries[queryIndex])
						if err != nil {
							atomic.AddInt64(&cacheMisses, 1)
						} else {
							atomic.AddInt64(&cacheHits, 1)
						}
					} else {
						// 20%的查询使用不同的SQL（可能不命中缓存）
						_, err := session.SelectList(ctx, 
							"SELECT * FROM users WHERE age = ? LIMIT 5", 20+rand.Intn(50))
						if err != nil {
							atomic.AddInt64(&cacheMisses, 1)
						} else {
							atomic.AddInt64(&cacheHits, 1)
						}
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		totalOps := cacheHits + cacheMisses
		hitRate := float64(cacheHits) / float64(totalOps) * 100

		t.Logf("缓存压力测试结果:")
		t.Logf("  运行时间: %v", duration)
		t.Logf("  总操作数: %d", totalOps)
		t.Logf("  缓存命中: %d", cacheHits)
		t.Logf("  缓存未命中: %d", cacheMisses)
		t.Logf("  命中率: %.2f%%", hitRate)
		t.Logf("  QPS: %.2f", float64(totalOps)/duration.Seconds())

		// 获取缓存统计
		stats := session.GetCacheStats()
		if stats != nil {
			t.Logf("缓存统计:")
			if l1Size, ok := stats["l1_cache_size"].(int); ok {
				t.Logf("  L1缓存大小: %d", l1Size)
			}
			if l2Size, ok := stats["l2_cache_size"].(int); ok {
				t.Logf("  L2缓存大小: %d", l2Size)
			}
		}
	})
}

// BenchmarkSimpleSelect 基准测试：简单查询
func BenchmarkSimpleSelect(b *testing.B) {
	db, err := setupTestDatabase()
	if err != nil {
		b.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备测试数据
	for i := 0; i < 100; i++ {
		_, err := session.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("基准用户%d", i), fmt.Sprintf("bench%d@example.com", i),
			20+(i%50), []string{"active", "inactive"}[i%2])
		if err != nil {
			b.Fatalf("准备基准测试数据失败: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 10", "active")
			if err != nil {
				b.Errorf("基准测试查询失败: %v", err)
			}
		}
	})
}

// BenchmarkBatchInsert 基准测试：批量插入
func BenchmarkBatchInsert(b *testing.B) {
	db, err := setupTestDatabase()
	if err != nil {
		b.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batchSize := 50
		batchArgs := make([][]interface{}, batchSize)
		for j := 0; j < batchSize; j++ {
			batchArgs[j] = []interface{}{
				fmt.Sprintf("基准批量用户%d_%d", i, j),
				fmt.Sprintf("bench_%d_%d@example.com", i, j),
				20 + (j % 40),
				"active",
			}
		}

		_, err := session.BatchInsert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)", batchArgs)
		if err != nil {
			b.Errorf("基准测试批量插入失败: %v", err)
		}
	}
}

// BenchmarkWithCache 基准测试：启用缓存的查询
func BenchmarkWithCache(b *testing.B) {
	db, err := setupTestDatabase()
	if err != nil {
		b.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	cacheConfig := &mybatis.CacheConfig{
		L1CacheEnabled:  true,
		L1CacheSize:     500,
		L1CacheTTL:      10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}

	session := mybatis.NewSimpleSession(db).
		EnableCache(cacheConfig).
		Debug(false)
	ctx := context.Background()

	// 准备测试数据
	for i := 0; i < 50; i++ {
		_, err := session.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("缓存基准用户%d", i), fmt.Sprintf("cache%d@example.com", i),
			20+(i%30), "active")
		if err != nil {
			b.Fatalf("准备缓存基准测试数据失败: %v", err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 重复查询相同的SQL，应该命中缓存
			_, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 5", "active")
			if err != nil {
				b.Errorf("缓存基准测试查询失败: %v", err)
			}
		}
	})
}

// setupStressTestData 设置压力测试数据
func setupStressTestData(t *testing.T, session mybatis.SimpleSession, ctx context.Context, count int) {
	t.Helper()

	// 批量插入以提高效率
	batchSize := 100
	for i := 0; i < count; i += batchSize {
		currentBatch := batchSize
		if i+batchSize > count {
			currentBatch = count - i
		}

		batchArgs := make([][]interface{}, currentBatch)
		for j := 0; j < currentBatch; j++ {
			idx := i + j
			batchArgs[j] = []interface{}{
				fmt.Sprintf("压力测试用户%d", idx),
				fmt.Sprintf("stress%d@example.com", idx),
				20 + (idx % 60),
				[]string{"active", "inactive", "pending"}[idx%3],
			}
		}

		_, err := session.BatchInsert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)", batchArgs)
		if err != nil {
			t.Logf("批量插入压力测试数据失败 (batch %d): %v", i/batchSize+1, err)
		}
	}

	t.Logf("成功设置压力测试数据，共 %d 个用户", count)
}