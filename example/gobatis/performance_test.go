// Package gobatis_test 性能测试和压力测试
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// setupInMemoryDatabase 设置内存数据库
func setupInMemoryDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
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
		return nil, err
	}
	
	// 自动迁移表结构
	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, err
	}
	
	return db, nil
}

// BenchmarkSimpleSession 简化版Session性能基准测试
func BenchmarkSimpleSession(b *testing.B) {
	// 跳过如果没有数据库连接  
	db, err := setupInMemoryDatabase()
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}

	// 创建简化版session
	session := mybatis.NewSimpleSession(db).Debug(false) // 关闭debug提升性能
	ctx := context.Background()

	// 准备测试数据
	setupBenchmarkData(b, session, ctx)

	b.ResetTimer() // 重置计时器，不计算准备时间

	b.Run("SelectOne", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			id := rand.Intn(1000) + 1
			_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", id)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SelectList", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			status := []string{"active", "inactive", "pending"}[rand.Intn(3)]
			_, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 10", status)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			timestamp := time.Now().UnixNano()
			name := fmt.Sprintf("BenchUser_%d_%d", timestamp, i)
			email := fmt.Sprintf("bench_%d_%d@test.com", timestamp, i)
			_, err := session.Insert(ctx, 
				"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
				name, email, 25, "active")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Update", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			id := rand.Intn(1000) + 1
			newAge := rand.Intn(80) + 18
			_, err := session.Update(ctx, "UPDATE users SET age = ? WHERE id = ?", newAge, id)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkXMLSession XML映射器性能基准测试
func BenchmarkXMLSession(b *testing.B) {
	db, err := setupInMemoryDatabase()
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// 创建XML session
	session := mybatis.NewXMLMapper(db)
	session.Debug(false)
	ctx := context.Background()

	// 加载XML映射
	err = session.LoadMapperXMLFromString(getUserMapperXML())
	if err != nil {
		b.Fatal(err)
	}

	// 准备测试数据
	setupBenchmarkData(b, session, ctx)

	b.ResetTimer()

	b.Run("XMLSelectById", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			id := rand.Intn(1000) + 1
			_, err := session.SelectOneByID(ctx, "UserMapper.selectById", id)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("XMLSelectList", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			query := UserQuery{
				Status: []string{"active", "inactive", "pending"}[rand.Intn(3)],
				AgeMin: 20,
				AgeMax: 60,
			}
			_, err := session.SelectListByID(ctx, "UserMapper.selectByCondition", query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkDryRun DryRun模式性能测试
func BenchmarkDryRun(b *testing.B) {
	db, err := setupInMemoryDatabase()
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	session := mybatis.NewSimpleSession(db).DryRun(true).Debug(false)
	ctx := context.Background()

	b.ResetTimer()

	b.Run("DryRunSelect", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", i)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("DryRunInsert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := session.Insert(ctx,
				"INSERT INTO users (name, email) VALUES (?, ?)",
				fmt.Sprintf("User_%d", i), fmt.Sprintf("user_%d@test.com", i))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkPagination 分页查询性能测试
func BenchmarkPagination(b *testing.B) {
	db, err := setupInMemoryDatabase()
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备大量测试数据
	setupLargeDataset(b, session, ctx, 10000)

	b.ResetTimer()

	b.Run("SmallPage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			page := mybatis.PageRequest{
				Page: rand.Intn(100) + 1,
				Size: 10,
			}
			_, err := session.SelectPage(ctx, "SELECT * FROM users WHERE status = 'active'", page)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MediumPage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			page := mybatis.PageRequest{
				Page: rand.Intn(20) + 1,
				Size: 50,
			}
			_, err := session.SelectPage(ctx, "SELECT * FROM users WHERE status = 'active'", page)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("LargePage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			page := mybatis.PageRequest{
				Page: rand.Intn(5) + 1,
				Size: 200,
			}
			_, err := session.SelectPage(ctx, "SELECT * FROM users WHERE status = 'active'", page)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// TestConcurrentAccess 并发访问压力测试
func TestConcurrentAccess(t *testing.T) {
	db, err := setupInMemoryDatabase()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备测试数据
	setupBenchmarkData(t, session, ctx)

	tests := []struct {
		name        string
		goroutines  int
		operations  int
		operation   func(session mybatis.SimpleSession, ctx context.Context, id int) error
	}{
		{
			name:       "ConcurrentRead",
			goroutines: 50,
			operations: 100,
			operation: func(session mybatis.SimpleSession, ctx context.Context, id int) error {
				_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", id%1000+1)
				return err
			},
		},
		{
			name:       "ConcurrentWrite",
			goroutines: 20,
			operations: 50,
			operation: func(session mybatis.SimpleSession, ctx context.Context, id int) error {
				name := fmt.Sprintf("ConcurrentUser_%d", id)
				email := fmt.Sprintf("concurrent_%d@test.com", id)
				_, err := session.Insert(ctx,
					"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
					name, email, 25, "active")
				return err
			},
		},
		{
			name:       "MixedOperations",
			goroutines: 30,
			operations: 100,
			operation: func(session mybatis.SimpleSession, ctx context.Context, id int) error {
				if id%3 == 0 {
					// 读操作
					_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", id%1000+1)
					return err
				} else if id%3 == 1 {
					// 写操作
					name := fmt.Sprintf("MixedUser_%d", id)
					email := fmt.Sprintf("mixed_%d@test.com", id)
					_, err := session.Insert(ctx,
						"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
						name, email, 25, "active")
					return err
				} else {
					// 更新操作
					newAge := rand.Intn(80) + 18
					_, err := session.Update(ctx, "UPDATE users SET age = ? WHERE id = ?", newAge, id%1000+1)
					return err
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			errors := make(chan error, tt.goroutines*tt.operations)
			start := time.Now()

			// 启动多个goroutine
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()
					
					// 每个goroutine执行多次操作
					for j := 0; j < tt.operations; j++ {
						if err := tt.operation(session, ctx, goroutineID*tt.operations+j); err != nil {
							errors <- fmt.Errorf("goroutine %d operation %d: %w", goroutineID, j, err)
							return
						}
					}
				}(i)
			}

			wg.Wait()
			close(errors)
			
			duration := time.Since(start)
			totalOps := tt.goroutines * tt.operations

			// 检查错误
			errorCount := 0
			for err := range errors {
				t.Errorf("Concurrent operation error: %v", err)
				errorCount++
			}

			if errorCount == 0 {
				opsPerSecond := float64(totalOps) / duration.Seconds()
				t.Logf("%s: %d operations in %v (%.2f ops/sec)",
					tt.name, totalOps, duration, opsPerSecond)
			} else {
				t.Errorf("%s: %d errors out of %d operations", tt.name, errorCount, totalOps)
			}
		})
	}
}

// TestMemoryUsage 内存使用测试
func TestMemoryUsage(t *testing.T) {
	db, err := setupInMemoryDatabase()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 记录初始内存
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// 执行大量操作
	const operations = 10000
	for i := 0; i < operations; i++ {
		if i%4 == 0 {
			_, _ = session.SelectList(ctx, "SELECT * FROM users LIMIT 10")
		} else if i%4 == 1 {
			name := fmt.Sprintf("MemUser_%d", i)
			email := fmt.Sprintf("mem_%d@test.com", i)
			_, _ = session.Insert(ctx,
				"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
				name, email, 25, "active")
		} else if i%4 == 2 {
			_, _ = session.Update(ctx, "UPDATE users SET status = 'updated' WHERE id = ?", i%100+1)
		} else {
			_, _ = session.SelectOne(ctx, "SELECT COUNT(*) FROM users")
		}

		// 每1000次操作检查一次内存
		if i%1000 == 999 {
			runtime.GC()
			runtime.ReadMemStats(&m2)
			memUsed := (m2.Alloc - m1.Alloc) / 1024 / 1024 // MB
			t.Logf("After %d operations: Memory used: %d MB", i+1, memUsed)
		}
	}

	// 最终内存使用
	runtime.GC()
	runtime.ReadMemStats(&m2)
	memUsed := (m2.Alloc - m1.Alloc) / 1024 / 1024 // MB
	memPerOp := float64(m2.Alloc-m1.Alloc) / float64(operations) // bytes per operation

	t.Logf("Total memory used: %d MB", memUsed)
	t.Logf("Memory per operation: %.2f bytes", memPerOp)

	// 检查内存泄漏（简单检查）
	if memPerOp > 1000 { // 超过1KB每操作可能有问题
		t.Errorf("Potential memory leak detected: %.2f bytes per operation", memPerOp)
	}
}

// TestLongRunning 长时间运行稳定性测试
func TestLongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long running test in short mode")
	}

	db, err := setupInMemoryDatabase()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 添加性能监控钩子
	var totalOperations int64
	var totalDuration time.Duration
	var slowQueries int64
	var mu sync.Mutex

	session.AddAfterHook(func(ctx context.Context, result interface{}, duration time.Duration, err error) {
		mu.Lock()
		defer mu.Unlock()
		
		totalOperations++
		totalDuration += duration
		
		if duration > 100*time.Millisecond {
			slowQueries++
		}
	})

	// 运行5分钟的连续操作
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	operationCounter := 0
	startTime := time.Now()

	for {
		select {
		case <-timeout:
			// 超时，结束测试
			goto TestComplete
		case <-ticker.C:
			// 执行操作
			operationType := operationCounter % 4
			switch operationType {
			case 0:
				_, _ = session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", rand.Intn(1000)+1)
			case 1:
				_, _ = session.SelectList(ctx, "SELECT * FROM users WHERE status = 'active' LIMIT 5")
			case 2:
				name := fmt.Sprintf("LongRunUser_%d", operationCounter)
				email := fmt.Sprintf("longrun_%d@test.com", operationCounter)
				_, _ = session.Insert(ctx,
					"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
					name, email, rand.Intn(60)+18, "active")
			case 3:
				_, _ = session.Update(ctx, "UPDATE users SET age = ? WHERE id = ?",
					rand.Intn(80)+18, rand.Intn(1000)+1)
			}
			operationCounter++
		}
	}

TestComplete:
	totalTime := time.Since(startTime)
	
	mu.Lock()
	avgDuration := totalDuration / time.Duration(totalOperations)
	slowQueryRate := float64(slowQueries) / float64(totalOperations) * 100
	mu.Unlock()

	opsPerSecond := float64(operationCounter) / totalTime.Seconds()

	t.Logf("Long running test completed:")
	t.Logf("  Total time: %v", totalTime)
	t.Logf("  Total operations: %d", operationCounter)
	t.Logf("  Operations per second: %.2f", opsPerSecond)
	t.Logf("  Average operation duration: %v", avgDuration)
	t.Logf("  Slow query rate: %.2f%%", slowQueryRate)

	// 性能指标检查
	if opsPerSecond < 100 {
		t.Errorf("Performance degradation: only %.2f ops/sec", opsPerSecond)
	}

	if slowQueryRate > 5.0 {
		t.Errorf("Too many slow queries: %.2f%%", slowQueryRate)
	}
}

// setupBenchmarkData 准备基准测试数据
func setupBenchmarkData(tb testing.TB, session interface{}, ctx context.Context) {
	var simpleSession mybatis.SimpleSession
	
	// 类型断言
	switch s := session.(type) {
	case mybatis.SimpleSession:
		simpleSession = s
	case mybatis.XMLSession:
		simpleSession = s  // XMLSession继承了SimpleSession
	default:
		tb.Fatal("Unsupported session type")
	}

	// 创建测试表 (SQLite兼容语法)
	_, err := simpleSession.Update(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			age INTEGER,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		tb.Fatal("Failed to create users table:", err)
	}

	// 添加索引
	_, _ = simpleSession.Update(ctx, "CREATE INDEX IF NOT EXISTS idx_status ON users(status)")
	_, _ = simpleSession.Update(ctx, "CREATE INDEX IF NOT EXISTS idx_age ON users(age)")

	// 插入基础测试数据
	for i := 1; i <= 1000; i++ {
		name := fmt.Sprintf("BenchUser_%d", i)
		email := fmt.Sprintf("bench_%d@test.com", i)
		age := rand.Intn(60) + 18
		status := []string{"active", "inactive", "pending"}[rand.Intn(3)]
		
		_, err := simpleSession.Insert(ctx,
			"INSERT OR IGNORE INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			name, email, age, status)
		if err != nil {
			tb.Logf("Warning: Failed to insert benchmark data %d: %v", i, err)
		}
	}
}

// setupLargeDataset 准备大数据集
func setupLargeDataset(tb testing.TB, session mybatis.SimpleSession, ctx context.Context, count int) {
	tb.Logf("Setting up large dataset with %d records...", count)
	
	// 批量插入
	batchSize := 100
	for i := 0; i < count; i += batchSize {
		end := i + batchSize
		if end > count {
			end = count
		}

		// 构建批量插入SQL
		values := make([]string, 0, end-i)
		args := make([]interface{}, 0, (end-i)*4)
		
		for j := i; j < end; j++ {
			values = append(values, "(?, ?, ?, ?)")
			args = append(args,
				fmt.Sprintf("LargeDataUser_%d", j),
				fmt.Sprintf("large_%d@test.com", j),
				rand.Intn(60)+18,
				"active")
		}

		sql := fmt.Sprintf("INSERT OR IGNORE INTO users (name, email, age, status) VALUES %s",
			strings.Join(values, ", "))
		
		_, err := session.Insert(ctx, sql, args...)
		if err != nil {
			tb.Logf("Warning: Failed to insert large dataset batch %d-%d: %v", i, end-1, err)
		}

		if i%1000 == 0 {
			tb.Logf("Inserted %d/%d records", i, count)
		}
	}
	tb.Logf("Large dataset setup completed")
}

// 使用complete_example.go中定义的getUserMapperXML函数和models.go中定义的UserQuery结构体