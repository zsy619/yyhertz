// Package main 快速性能测试
package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// BenchmarkQuickPerformance 快速性能基准测试
func BenchmarkQuickPerformance(b *testing.B) {
	// 初始化测试数据库
	db, err := InitializeTestDatabase()
	if err != nil {
		b.Fatalf("初始化测试数据库失败: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	session := mybatis.NewSimpleSession(db)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := session.Insert(ctx,
				"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
				fmt.Sprintf("BenchUser_%d", i), fmt.Sprintf("bench_%d@test.com", i), 25, "active")
			if err != nil {
				b.Fatalf("插入失败: %v", err)
			}
		}
	})

	// 先插入一些数据用于查询
	for i := 0; i < 100; i++ {
		session.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("QueryUser_%d", i), fmt.Sprintf("query_%d@test.com", i), 25, "active")
	}

	b.Run("Select", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			id := (i % 100) + 1
			_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", id)
			if err != nil {
				b.Logf("查询失败: %v", err)
			}
		}
	})

	b.Run("SelectList", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 10", "active")
			if err != nil {
				b.Logf("列表查询失败: %v", err)
			}
		}
	})
}

// TestQuickPerformanceReport 生成快速性能报告
func TestQuickPerformanceReport(t *testing.T) {
	// 初始化测试数据库
	db, err := InitializeTestDatabase()
	if err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	session := mybatis.NewSimpleSession(db)
	ctx := context.Background()

	// 性能测试参数
	const (
		insertCount = 1000
		selectCount = 1000
		warmupCount = 100
	)

	t.Log("=== 快速性能测试报告 ===")

	// 预热
	for i := 0; i < warmupCount; i++ {
		session.Insert(ctx, "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("Warmup_%d", i), fmt.Sprintf("warmup_%d@test.com", i), 25, "active")
	}

	// 测试插入性能
	t.Run("插入性能测试", func(t *testing.T) {
		start := time.Now()
		successCount := 0

		for i := 0; i < insertCount; i++ {
			_, err := session.Insert(ctx,
				"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
				fmt.Sprintf("PerfUser_%d", i), fmt.Sprintf("perf_%d@test.com", i), 25, "active")
			if err == nil {
				successCount++
			}
		}

		duration := time.Since(start)
		opsPerSec := float64(successCount) / duration.Seconds()

		t.Logf("插入性能结果:")
		t.Logf("  总操作数: %d", insertCount)
		t.Logf("  成功操作数: %d", successCount)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  平均每个操作: %v", duration/time.Duration(successCount))
		t.Logf("  吞吐量: %.2f ops/s", opsPerSec)
	})

	// 测试查询性能
	t.Run("查询性能测试", func(t *testing.T) {
		start := time.Now()
		successCount := 0

		for i := 0; i < selectCount; i++ {
			id := (i % 100) + 1
			_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", id)
			if err == nil {
				successCount++
			}
		}

		duration := time.Since(start)
		opsPerSec := float64(successCount) / duration.Seconds()

		t.Logf("查询性能结果:")
		t.Logf("  总操作数: %d", selectCount)
		t.Logf("  成功操作数: %d", successCount)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  平均每个操作: %v", duration/time.Duration(successCount))
		t.Logf("  吞吐量: %.2f ops/s", opsPerSec)
	})

	// 测试列表查询性能
	t.Run("列表查询性能测试", func(t *testing.T) {
		start := time.Now()
		successCount := 0
		totalRecords := 0

		for i := 0; i < selectCount; i++ {
			results, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 10", "active")
			if err == nil {
				successCount++
				totalRecords += len(results)
			}
		}

		duration := time.Since(start)
		opsPerSec := float64(successCount) / duration.Seconds()

		t.Logf("列表查询性能结果:")
		t.Logf("  总操作数: %d", selectCount)
		t.Logf("  成功操作数: %d", successCount)
		t.Logf("  查询到的记录总数: %d", totalRecords)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  平均每个操作: %v", duration/time.Duration(successCount))
		t.Logf("  吞吐量: %.2f ops/s", opsPerSec)
	})
}