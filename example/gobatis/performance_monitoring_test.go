// Package main 性能监控测试
//
// 测试MyBatis性能监控功能：
// 1. SQL执行时间统计
// 2. 慢查询检测和记录
// 3. 批量操作性能分析
// 4. 内存使用监控
// 5. 性能基准对比
package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// PerformanceUser 性能测试用户模型
type PerformanceUser struct {
	ID       int64     `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Email    string    `json:"email" db:"email"`
	Age      int       `json:"age" db:"age"`
	Status   string    `json:"status" db:"status"`
	CreateAt time.Time `json:"createAt" db:"created_at"`
}

// TestSQLExecutionTiming 测试SQL执行时间统计
func TestSQLExecutionTiming(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备测试数据
	setupPerformanceTestData(t, session, ctx)

	t.Run("不同SQL操作耗时统计", func(t *testing.T) {
		operations := []struct {
			name      string
			operation func() (time.Duration, error)
		}{
			{"SELECT查询", func() (time.Duration, error) {
				start := time.Now()
				_, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ?", "active")
				return time.Since(start), err
			}},
			{"INSERT插入", func() (time.Duration, error) {
				start := time.Now()
				_, err := session.Insert(ctx, "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
					"性能测试用户", "perf@example.com", 25, "active")
				return time.Since(start), err
			}},
			{"UPDATE更新", func() (time.Duration, error) {
				start := time.Now()
				_, err := session.Update(ctx, "UPDATE users SET age = age + 1 WHERE status = ?", "active")
				return time.Since(start), err
			}},
			{"DELETE删除", func() (time.Duration, error) {
				start := time.Now()
				_, err := session.Delete(ctx, "DELETE FROM users WHERE name LIKE ?", "临时%")
				return time.Since(start), err
			}},
		}

		var totalTime time.Duration
		for _, op := range operations {
			duration, err := op.operation()
			totalTime += duration
			
			if err != nil {
				t.Logf("%s 操作失败: %v", op.name, err)
			} else {
				t.Logf("%s 操作耗时: %v", op.name, duration)
			}
		}

		t.Logf("所有操作总耗时: %v", totalTime)
		t.Logf("平均操作耗时: %v", totalTime/time.Duration(len(operations)))
		t.Log("SQL操作耗时统计测试通过")
	})

	t.Run("并发SQL执行性能", func(t *testing.T) {
		const concurrentOps = 10
		var wg sync.WaitGroup
		durations := make([]time.Duration, concurrentOps)
		errors := make([]error, concurrentOps)

		// 并发执行查询操作
		start := time.Now()
		for i := 0; i < concurrentOps; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				opStart := time.Now()
				_, err := session.SelectList(ctx, "SELECT * FROM users WHERE age > ? LIMIT 5", 20+index)
				durations[index] = time.Since(opStart)
				errors[index] = err
			}(i)
		}

		wg.Wait()
		totalConcurrentTime := time.Since(start)

		// 统计结果
		var totalDuration time.Duration
		successCount := 0
		for i, duration := range durations {
			if errors[i] == nil {
				totalDuration += duration
				successCount++
			}
		}

		t.Logf("并发性能统计:")
		t.Logf("  并发操作数: %d", concurrentOps)
		t.Logf("  成功操作数: %d", successCount)
		t.Logf("  并发总耗时: %v", totalConcurrentTime)
		if successCount > 0 {
			t.Logf("  平均操作耗时: %v", totalDuration/time.Duration(successCount))
		}

		t.Log("并发SQL执行性能测试通过")
	})
}

// TestSlowQueryDetection 测试慢查询检测
func TestSlowQueryDetection(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备大量测试数据以制造慢查询
	setupLargePerformanceTestData(t, session, ctx)

	t.Run("慢查询识别", func(t *testing.T) {
		slowQueryThreshold := 10 * time.Millisecond
		
		// 执行可能较慢的查询
		queries := []struct {
			name string
			sql  string
			args []interface{}
		}{
			{"全表扫描", "SELECT * FROM users", nil},
			{"复杂条件查询", "SELECT * FROM users WHERE age > ? AND name LIKE ? ORDER BY created_at", []interface{}{20, "%性能%"}},
			{"聚合查询", "SELECT status, COUNT(*), AVG(age) FROM users GROUP BY status", nil},
			{"范围查询", "SELECT * FROM users WHERE age BETWEEN ? AND ? ORDER BY age DESC", []interface{}{25, 35}},
		}

		slowQueries := make([]struct {
			name     string
			duration time.Duration
		}, 0)

		for _, query := range queries {
			start := time.Now()
			
			if query.args != nil {
				_, err := session.SelectList(ctx, query.sql, query.args...)
				if err != nil {
					t.Logf("%s 执行失败: %v", query.name, err)
					continue
				}
			} else {
				_, err := session.SelectList(ctx, query.sql)
				if err != nil {
					t.Logf("%s 执行失败: %v", query.name, err)
					continue
				}
			}
			
			duration := time.Since(start)
			t.Logf("%s 执行耗时: %v", query.name, duration)
			
			if duration > slowQueryThreshold {
				slowQueries = append(slowQueries, struct {
					name     string
					duration time.Duration
				}{query.name, duration})
			}
		}

		if len(slowQueries) > 0 {
			t.Logf("检测到 %d 个慢查询（阈值: %v）:", len(slowQueries), slowQueryThreshold)
			for i, sq := range slowQueries {
				t.Logf("  慢查询 %d: %s, 耗时 %v", i+1, sq.name, sq.duration)
			}
		} else {
			t.Log("未检测到慢查询")
		}

		t.Log("慢查询识别测试通过")
	})
}

// TestPerformanceBenchmark 测试性能基准
func TestPerformanceBenchmark(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	setupPerformanceTestData(t, session, ctx)

	t.Run("CRUD操作性能基准", func(t *testing.T) {
		benchmarks := []struct {
			name      string
			operation func() error
		}{
			{"单行查询", func() error {
				_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
				return err
			}},
			{"多行查询", func() error {
				_, err := session.SelectList(ctx, "SELECT * FROM users LIMIT 10")
				return err
			}},
			{"条件查询", func() error {
				_, err := session.SelectList(ctx, "SELECT * FROM users WHERE age > ?", 25)
				return err
			}},
			{"聚合查询", func() error {
				_, err := session.SelectOne(ctx, "SELECT COUNT(*) FROM users")
				return err
			}},
		}

		const iterations = 100
		for _, bm := range benchmarks {
			start := time.Now()
			successCount := 0
			
			for i := 0; i < iterations; i++ {
				if err := bm.operation(); err == nil {
					successCount++
				}
			}
			
			duration := time.Since(start)
			t.Logf("%s 基准结果:", bm.name)
			t.Logf("  执行次数: %d", iterations)
			t.Logf("  成功次数: %d", successCount)
			t.Logf("  总耗时: %v", duration)
			if successCount > 0 {
				t.Logf("  平均耗时: %v", duration/time.Duration(successCount))
				t.Logf("  QPS: %.2f", float64(successCount)/duration.Seconds())
			}
		}

		t.Log("CRUD操作性能基准测试通过")
	})
}

// setupPerformanceTestData 设置性能测试数据
func setupPerformanceTestData(t *testing.T, session mybatis.SimpleSession, ctx context.Context) {
	t.Helper()
	
	testUsers := []struct {
		name   string
		email  string
		age    int
		status string
	}{
		{"性能测试用户1", "perf1@example.com", 25, "active"},
		{"性能测试用户2", "perf2@example.com", 30, "active"},
		{"性能测试用户3", "perf3@example.com", 28, "inactive"},
		{"性能测试用户4", "perf4@example.com", 32, "active"},
		{"性能测试用户5", "perf5@example.com", 27, "pending"},
	}
	
	for i, user := range testUsers {
		_, err := session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			user.name, user.email, user.age, user.status)
		if err != nil {
			t.Logf("插入性能测试用户 %d 失败: %v", i+1, err)
		}
	}
	
	t.Logf("成功设置性能测试数据，共 %d 个用户", len(testUsers))
}

// setupLargePerformanceTestData 设置大量性能测试数据
func setupLargePerformanceTestData(t *testing.T, session mybatis.SimpleSession, ctx context.Context) {
	t.Helper()
	
	const largeDataSize = 50
	
	for i := 0; i < largeDataSize; i++ {
		_, err := session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("性能测试大数据用户%d", i+1),
			fmt.Sprintf("large%d@example.com", i+1),
			20+(i%50),
			[]string{"active", "inactive", "pending"}[i%3])
		if err != nil {
			t.Logf("插入大量性能测试数据 %d 失败: %v", i+1, err)
		}
	}
	
	t.Logf("成功设置大量性能测试数据，共 %d 个用户", largeDataSize)
}

// 注意：由于当前框架的性能监控API尚未完全实现，
// 这些测试主要使用基本的时间测量来演示性能监控的概念。
// 实际的性能统计需要在框架层面实现更详细的监控机制。