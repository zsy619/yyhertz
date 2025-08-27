package mybatis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBatchOperationOptimizer(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建测试表
	err = db.Exec(`
		CREATE TABLE test_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			value TEXT NOT NULL
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 测试配置
	config := &BatchConfig{
		OptimalBatchSize:     100,
		MaxBatchSize:         1000,
		MinBatchSize:         10,
		MaxConcurrentBatch:   5,
		WorkerPoolSize:       10,
		PoolInitialSize:      100,
		PoolMaxSize:          1000,
		MemoryThreshold:      100 * 1024 * 1024, // 100MB
		EnableAdaptive:       true,
		AdaptiveInterval:     1 * time.Second,
		PerformanceWindow:    10,
		EnableMemoryMonitor:  true,
		MonitorInterval:      1 * time.Second,
	}

	optimizer := NewBatchOperationOptimizer(config)
	defer optimizer.Close()

	t.Run("SingleBatchOperation", func(t *testing.T) {
		sql := "INSERT INTO test_data (value) VALUES (?)"
		args := make([][]any, 0, 50)
		for i := 0; i < 50; i++ {
			args = append(args, []any{fmt.Sprintf("value-%d", i)})
		}

		affected, err := optimizer.OptimizedBatchInsert(context.Background(), db, sql, args)
		if err != nil {
			t.Fatalf("OptimizedBatchInsert failed: %v", err)
		}

		if affected != 50 {
			t.Errorf("Expected 50 affected rows, got %d", affected)
		}

		// 验证数据插入
		var count int64
		if err := db.Table("test_data").Count(&count).Error; err != nil {
			t.Fatalf("Failed to count rows: %v", err)
		}

		if count != 50 {
			t.Errorf("Expected 50 rows in table, got %d", count)
		}
	})

	t.Run("ConcurrentBatchOperations", func(t *testing.T) {
		sql := "INSERT INTO test_data (value) VALUES (?)"
		args := make([][]any, 0, 500)
		for i := 0; i < 500; i++ {
			args = append(args, []any{fmt.Sprintf("concurrent-value-%d", i)})
		}

		affected, err := optimizer.OptimizedBatchInsert(context.Background(), db, sql, args)
		if err != nil {
			t.Fatalf("OptimizedBatchInsert failed: %v", err)
		}

		if affected != 500 {
			t.Errorf("Expected 500 affected rows, got %d", affected)
		}

		// 验证数据插入
		var count int64
		if err := db.Table("test_data").Count(&count).Error; err != nil {
			t.Fatalf("Failed to count rows: %v", err)
		}

		if count != 550 { // 500 new + 50 from previous test
			t.Errorf("Expected 550 rows in table, got %d", count)
		}
	})

	t.Run("EmptyBatchOperation", func(t *testing.T) {
		sql := "INSERT INTO test_data (value) VALUES (?)"
		args := make([][]any, 0)

		affected, err := optimizer.OptimizedBatchInsert(context.Background(), db, sql, args)
		if err != nil {
			t.Fatalf("OptimizedBatchInsert failed: %v", err)
		}

		if affected != 0 {
			t.Errorf("Expected 0 affected rows for empty batch, got %d", affected)
		}
	})

	t.Run("MemoryPoolUsage", func(t *testing.T) {
		// 检查内存池命中率
		stats := optimizer.GetOptimizationStats()
		poolHits := stats["memory_stats"].(map[string]any)["pool_hits"].(int64)
		poolMisses := stats["memory_stats"].(map[string]any)["pool_misses"].(int64)

		if poolHits == 0 && poolMisses > 0 {
			t.Error("Expected some pool hits after multiple operations")
		}
	})

	t.Run("AdaptiveBatchSizeAdjustment", func(t *testing.T) {
		// 等待自适应调整
		time.Sleep(2 * time.Second)

		stats := optimizer.GetOptimizationStats()
		adaptiveAdjustments := stats["optimization_stats"].(map[string]any)["adaptive_adjustments"].(int64)

		if adaptiveAdjustments == 0 {
			t.Error("Expected adaptive adjustments to have occurred")
		}
	})

	t.Run("MemoryMonitoring", func(t *testing.T) {
		// 等待内存监控
		time.Sleep(2 * time.Second)

		stats := optimizer.GetOptimizationStats()
		peakUsage := stats["memory_stats"].(map[string]any)["peak_memory_usage"].(int64)

		if peakUsage == 0 {
			t.Error("Expected non-zero peak memory usage")
		}
	})
}