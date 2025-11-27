package mybatis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPerformanceMonitor(t *testing.T) {
	t.Run("PerformanceMonitor_Basic", func(t *testing.T) {
		config := &mybatis.PerformanceConfig{
			SlowQueryThreshold:   100 * time.Millisecond,
			EnableSlowQueryLog:   true,
			MaxSlowQueryRecords:  100,
			EnableStatistics:     true,
			MaxStatisticsRecords: 100,
		}
		
		monitor := mybatis.NewPerformanceMonitor(config)
		defer monitor.Close()

		// 模拟SQL执行记录
		monitor.RecordExecution(
			"test.selectUser",
			"SELECT * FROM users WHERE id = ?",
			[]any{1},
			50*time.Millisecond,
			"master",
			nil,
		)

		// 模拟慢查询
		monitor.RecordExecution(
			"test.selectSlowQuery",
			"SELECT * FROM large_table ORDER BY created_at",
			[]any{},
			200*time.Millisecond,
			"slave",
			nil,
		)

		// 获取统计信息
		stats := monitor.GetStatistics()
		if len(stats) != 2 {
			t.Errorf("Expected 2 statistics, got %d", len(stats))
		}

		// 验证快查询统计
		if userStats, exists := stats["test.selectUser"]; exists {
			if userStats.ExecuteCount != 1 {
				t.Errorf("Expected execute count 1, got %d", userStats.ExecuteCount)
			}
			if userStats.SlowQueryCount != 0 {
				t.Errorf("Expected 0 slow queries, got %d", userStats.SlowQueryCount)
			}
		} else {
			t.Error("User statistics not found")
		}

		// 验证慢查询统计
		if slowStats, exists := stats["test.selectSlowQuery"]; exists {
			if slowStats.ExecuteCount != 1 {
				t.Errorf("Expected execute count 1, got %d", slowStats.ExecuteCount)
			}
			if slowStats.SlowQueryCount != 1 {
				t.Errorf("Expected 1 slow query, got %d", slowStats.SlowQueryCount)
			}
		} else {
			t.Error("Slow query statistics not found")
		}

		// 获取慢查询记录
		slowQueries := monitor.GetSlowQueries(10)
		if len(slowQueries) != 1 {
			t.Errorf("Expected 1 slow query, got %d", len(slowQueries))
		}

		if len(slowQueries) > 0 {
			slowQuery := slowQueries[0]
			if slowQuery.StatementID != "test.selectSlowQuery" {
				t.Errorf("Expected statement ID 'test.selectSlowQuery', got %s", slowQuery.StatementID)
			}
			if slowQuery.ExecuteTime < 200*time.Millisecond {
				t.Errorf("Expected execution time >= 200ms, got %v", slowQuery.ExecuteTime)
			}
		}

		// 获取统计报告
		report := monitor.GetStatisticsReport()
		if report.TotalExecutions != 2 {
			t.Errorf("Expected 2 total executions, got %d", report.TotalExecutions)
		}
		if report.TotalSlowQueryExecutions != 1 {
			t.Errorf("Expected 1 slow query execution, got %d", report.TotalSlowQueryExecutions)
		}
		if report.SlowQueryRate != 0.5 {
			t.Errorf("Expected slow query rate 0.5, got %f", report.SlowQueryRate)
		}
	})

	t.Run("PerformanceMonitor_TimeDistribution", func(t *testing.T) {
		config := mybatis.DefaultPerformanceConfig()
		monitor := mybatis.NewPerformanceMonitor(config)
		defer monitor.Close()

		// 记录不同执行时间的查询
		testCases := []struct {
			duration     time.Duration
			expectedBucket string
		}{
			{50 * time.Millisecond, "0-100ms"},
			{200 * time.Millisecond, "100-500ms"},
			{800 * time.Millisecond, "500ms-1s"},
			{3 * time.Second, "1s-5s"},
			{8 * time.Second, "5s-10s"},
			{15 * time.Second, "10s+"},
		}

		for i, tc := range testCases {
			statementId := fmt.Sprintf("test.query_%d", i)
			monitor.RecordExecution(
				statementId,
				"SELECT * FROM test_table",
				[]any{},
				tc.duration,
				"master",
				nil,
			)
		}

		stats := monitor.GetStatistics()
		for i, tc := range testCases {
			statementId := fmt.Sprintf("test.query_%d", i)
			if stat, exists := stats[statementId]; exists {
				if count, bucketExists := stat.TimeDistribution[tc.expectedBucket]; !bucketExists || count != 1 {
					t.Errorf("Expected 1 count in bucket %s for query %d, got %d", tc.expectedBucket, i, count)
				}
			} else {
				t.Errorf("Statistics not found for query %d", i)
			}
		}
	})

	t.Run("PerformanceMonitor_ErrorTracking", func(t *testing.T) {
		config := mybatis.DefaultPerformanceConfig()
		monitor := mybatis.NewPerformanceMonitor(config)
		defer monitor.Close()

		// 记录成功和失败的执行
		monitor.RecordExecution("test.success", "SELECT 1", []any{}, 50*time.Millisecond, "master", nil)
		monitor.RecordExecution("test.error", "SELECT * FROM non_exist_table", []any{}, 10*time.Millisecond, "master", fmt.Errorf("table does not exist"))

		stats := monitor.GetStatistics()
		if successStats, exists := stats["test.success"]; exists {
			if successStats.ErrorCount != 0 {
				t.Errorf("Expected 0 errors for success query, got %d", successStats.ErrorCount)
			}
		}

		if errorStats, exists := stats["test.error"]; exists {
			if errorStats.ErrorCount != 1 {
				t.Errorf("Expected 1 error for error query, got %d", errorStats.ErrorCount)
			}
		}

		report := monitor.GetStatisticsReport()
		if report.TotalErrors != 1 {
			t.Errorf("Expected 1 total error, got %d", report.TotalErrors)
		}
		if report.ErrorRate != 0.5 {
			t.Errorf("Expected error rate 0.5, got %f", report.ErrorRate)
		}
	})
}

func TestSessionPerformanceIntegration(t *testing.T) {
	t.Run("Session_PerformanceMonitor", func(t *testing.T) {
		// 创建内存数据库
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}

		// 创建测试表
		err = db.Exec(`
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT NOT NULL
			);
			INSERT INTO users (name, email) VALUES 
				('User1', 'user1@example.com'),
				('User2', 'user2@example.com');
		`).Error
		if err != nil {
			t.Fatalf("Failed to setup test data: %v", err)
		}

		// 创建会话并启用性能监控
		config := &mybatis.PerformanceConfig{
			SlowQueryThreshold:   50 * time.Millisecond, // 设置较短的阈值用于测试
			EnableSlowQueryLog:   true,
			EnableStatistics:     true,
			MaxSlowQueryRecords:  100,
			MaxStatisticsRecords: 100,
		}
		session := mybatis.NewSessionWithPerformanceMonitoring(db, config)

		ctx := context.Background()

		// 执行一些查询
		results, err := session.SelectList(ctx, "SELECT * FROM users")
		if err != nil {
			t.Fatalf("Failed to execute query: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		// 执行单条查询
		result, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("Failed to execute single query: %v", err)
		}
		if result == nil {
			t.Error("Expected result, got nil")
		}

		// 执行插入操作
		affected, err := session.Insert(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "User3", "user3@example.com")
		if err != nil {
			t.Fatalf("Failed to execute insert: %v", err)
		}
		if affected != 1 {
			t.Errorf("Expected 1 affected row, got %d", affected)
		}

		// 获取性能统计
		stats := session.GetPerformanceStats()
		t.Logf("Performance stats: %d statements recorded", len(stats))

		// 获取性能报告
		report := session.GetPerformanceReport()
		if report.TotalExecutions < 3 {
			t.Errorf("Expected at least 3 executions, got %d", report.TotalExecutions)
		}

		t.Logf("Performance report: %d total executions, avg time: %v", 
			report.TotalExecutions, report.AvgExecutionTime)

		// 获取慢查询（如果有）
		slowQueries := session.GetSlowQueries(10)
		t.Logf("Slow queries: %d", len(slowQueries))
		for i, sq := range slowQueries {
			t.Logf("Slow query %d: %s (took %v)", i+1, sq.SQL, sq.ExecuteTime)
		}

		// 清空统计
		session = session.ClearPerformanceStats()
		clearedStats := session.GetPerformanceStats()
		if len(clearedStats) != 0 {
			t.Errorf("Expected 0 stats after clearing, got %d", len(clearedStats))
		}

		// 禁用性能监控
		session = session.DisablePerformanceMonitor()
	})
}