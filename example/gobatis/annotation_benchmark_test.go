package main

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// AnnotationBenchmarkSuite annotation基准测试套件
type AnnotationBenchmarkSuite struct {
	tester      *AnnotationStressTester
	tagParser   *mybatis.TagParser
	sqlGen      *mybatis.SQLGenerator
	warmupUsers []*TestUser
}

// NewAnnotationBenchmarkSuite 创建annotation基准测试套件
func NewAnnotationBenchmarkSuite() (*AnnotationBenchmarkSuite, error) {
	tester, err := NewAnnotationStressTester(":memory:")
	if err != nil {
		return nil, err
	}
	
	suite := &AnnotationBenchmarkSuite{
		tester:      tester,
		tagParser:   mybatis.NewTagParser(),
		sqlGen:      mybatis.NewSQLGenerator(),
		warmupUsers: make([]*TestUser, 1000),
	}
	
	// 预热数据
	suite.warmupData()
	
	return suite, nil
}

// warmupData 预热测试数据
func (suite *AnnotationBenchmarkSuite) warmupData() {
	for i := 0; i < 1000; i++ {
		suite.warmupUsers[i] = &TestUser{
			ID:       int64(i + 1),
			Username: fmt.Sprintf("benchmark_user_%d", i),
			Email:    fmt.Sprintf("benchmark_%d@test.com", i),
			Age:      20 + (i % 50),
			Status:   "active",
		}
	}
}

// BenchmarkTagParserStructParsing 基准测试: 结构体解析性能
func BenchmarkTagParserStructParsing(b *testing.B) {
	suite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		b.Fatalf("Failed to create benchmark suite: %v", err)
	}
	defer suite.tester.Close()
	
	user := &TestUser{}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	b.Run("ParseStruct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := suite.tagParser.ParseStruct(user)
			if err != nil {
				b.Fatalf("ParseStruct failed: %v", err)
			}
		}
	})
	
	b.Run("ParseStruct_Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := suite.tagParser.ParseStruct(user)
				if err != nil {
					b.Fatalf("ParseStruct failed: %v", err)
				}
			}
		})
	})
}

// BenchmarkTagParserCaching 基准测试: 缓存性能
func BenchmarkTagParserCaching(b *testing.B) {
	suite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		b.Fatalf("Failed to create benchmark suite: %v", err)
	}
	defer suite.tester.Close()
	
	// 预热缓存
	suite.tagParser.ParseStruct(&TestUser{})
	suite.tagParser.ParseStruct(&TestUserProfile{})
	suite.tagParser.ParseStruct(&TestOrder{})
	
	b.ReportAllocs()
	b.ResetTimer()
	
	b.Run("CacheHit_Single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := suite.tagParser.ParseStruct(&TestUser{})
			if err != nil {
				b.Fatalf("ParseStruct failed: %v", err)
			}
		}
	})
	
	b.Run("CacheHit_Mixed", func(b *testing.B) {
		entities := []any{&TestUser{}, &TestUserProfile{}, &TestOrder{}}
		for i := 0; i < b.N; i++ {
			entity := entities[i%len(entities)]
			_, err := suite.tagParser.ParseStruct(entity)
			if err != nil {
				b.Fatalf("ParseStruct failed: %v", err)
			}
		}
	})
	
	b.Run("CacheHit_Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			entities := []any{&TestUser{}, &TestUserProfile{}, &TestOrder{}}
			i := 0
			for pb.Next() {
				entity := entities[i%len(entities)]
				_, err := suite.tagParser.ParseStruct(entity)
				if err != nil {
					b.Fatalf("ParseStruct failed: %v", err)
				}
				i++
			}
		})
	})
}

// BenchmarkSQLGeneration 基准测试: SQL生成性能
func BenchmarkSQLGeneration(b *testing.B) {
	suite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		b.Fatalf("Failed to create benchmark suite: %v", err)
	}
	defer suite.tester.Close()
	
	user := suite.warmupUsers[0]
	
	b.ReportAllocs()
	b.ResetTimer()
	
	b.Run("GenerateInsertSQL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := suite.sqlGen.GenerateInsertSQL(user)
			if err != nil {
				b.Fatalf("GenerateInsertSQL failed: %v", err)
			}
		}
	})
	
	b.Run("GenerateUpdateSQL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := suite.sqlGen.GenerateUpdateSQL(user)
			if err != nil {
				b.Fatalf("GenerateUpdateSQL failed: %v", err)
			}
		}
	})
	
	b.Run("GenerateSelectSQL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := suite.sqlGen.GenerateSelectSQL(user, "id = ?")
			if err != nil {
				b.Fatalf("GenerateSelectSQL failed: %v", err)
			}
		}
	})
	
	b.Run("GenerateDeleteSQL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := suite.sqlGen.GenerateDeleteSQL(user)
			if err != nil {
				b.Fatalf("GenerateDeleteSQL failed: %v", err)
			}
		}
	})
	
	b.Run("GenerateSQL_Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				switch i % 4 {
				case 0:
					suite.sqlGen.GenerateInsertSQL(user)
				case 1:
					suite.sqlGen.GenerateUpdateSQL(user)
				case 2:
					suite.sqlGen.GenerateSelectSQL(user, "id = ?")
				case 3:
					suite.sqlGen.GenerateDeleteSQL(user)
				}
				i++
			}
		})
	})
}

// BenchmarkReflectionOperations 基准测试: 反射操作性能
func BenchmarkReflectionOperations(b *testing.B) {
	suite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		b.Fatalf("Failed to create benchmark suite: %v", err)
	}
	defer suite.tester.Close()
	
	user := suite.warmupUsers[0]
	userType := reflect.TypeOf(user).Elem()
	userValue := reflect.ValueOf(user).Elem()
	
	b.ReportAllocs()
	b.ResetTimer()
	
	b.Run("TypeOf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = reflect.TypeOf(user)
		}
	})
	
	b.Run("ValueOf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = reflect.ValueOf(user)
		}
	})
	
	b.Run("NumField", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = userType.NumField()
		}
	})
	
	b.Run("Field", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = userType.Field(0)
		}
	})
	
	b.Run("FieldByName", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = userValue.FieldByName("Username")
		}
	})
	
	b.Run("TagGet", func(b *testing.B) {
		field := userType.Field(1) // Username field
		for i := 0; i < b.N; i++ {
			_ = field.Tag.Get("column")
		}
	})
	
	b.Run("Interface", func(b *testing.B) {
		field := userValue.Field(1) // Username field
		for i := 0; i < b.N; i++ {
			_ = field.Interface()
		}
	})
}

// BenchmarkAnnotationVsDirectSQL 基准测试: annotation驱动 vs 直接SQL对比
func BenchmarkAnnotationVsDirectSQL(b *testing.B) {
	suite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		b.Fatalf("Failed to create benchmark suite: %v", err)
	}
	defer suite.tester.Close()
	
	ctx := context.Background()
	
	// 预编译的直接SQL
	directInsertSQL := "INSERT INTO test_users (username, email, age, status) VALUES (?, ?, ?, ?)"
	directSelectSQL := "SELECT id, username, email, age, status FROM test_users WHERE id = ?"
	
	b.ReportAllocs()
	b.ResetTimer()
	
	b.Run("Annotation_Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testUser := &TestUser{
				Username: fmt.Sprintf("bench_user_%d", i),
				Email:    fmt.Sprintf("bench_%d@test.com", i),
				Age:      25,
				Status:   "active",
			}
			_, err := suite.tester.annotationSess.Insert(ctx, testUser)
			if err != nil {
				b.Fatalf("Annotation Insert failed: %v", err)
			}
		}
	})
	
	b.Run("Direct_Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			username := fmt.Sprintf("bench_user_direct_%d", i)
			email := fmt.Sprintf("bench_direct_%d@test.com", i)
			_, err := suite.tester.db.ExecContext(ctx, directInsertSQL, username, email, 25, "active")
			if err != nil {
				b.Fatalf("Direct Insert failed: %v", err)
			}
		}
	})
	
	b.Run("Annotation_Select", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := suite.tester.annotationSess.SelectByID(ctx, &TestUser{}, int64(i%100+1))
			if err != nil {
				// 忽略not found错误
				continue
			}
		}
	})
	
	b.Run("Direct_Select", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var user TestUser
			row := suite.tester.db.QueryRowContext(ctx, directSelectSQL, int64(i%100+1))
			row.Scan(&user.ID, &user.Username, &user.Email, &user.Age, &user.Status)
		}
	})
}

// BenchmarkConcurrentAnnotationOperations 基准测试: 并发annotation操作
func BenchmarkConcurrentAnnotationOperations(b *testing.B) {
	suite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		b.Fatalf("Failed to create benchmark suite: %v", err)
	}
	defer suite.tester.Close()
	
	ctx := context.Background()
	
	b.ReportAllocs()
	b.ResetTimer()
	
	concurrencies := []int{1, 10, 50, 100, 200}
	
	for _, concurrency := range concurrencies {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					user := &TestUser{
						Username: fmt.Sprintf("concurrent_user_%d_%d", concurrency, i),
						Email:    fmt.Sprintf("concurrent_%d_%d@test.com", concurrency, i),
						Age:      20 + (i % 50),
						Status:   "active",
					}
					
					// 混合操作测试
					switch i % 3 {
					case 0: // 插入
						suite.tester.annotationSess.Insert(ctx, user)
					case 1: // 查询
						suite.tester.annotationSess.SelectByID(ctx, &TestUser{}, int64(i%100+1))
					case 2: // 结构体解析
						suite.tagParser.ParseStruct(user)
					}
					i++
				}
			})
		})
	}
}

// BenchmarkMemoryUsage 基准测试: 内存使用情况
func BenchmarkMemoryUsage(b *testing.B) {
	suite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		b.Fatalf("Failed to create benchmark suite: %v", err)
	}
	defer suite.tester.Close()
	
	b.ReportAllocs()
	
	// 预热
	for i := 0; i < 1000; i++ {
		suite.tagParser.ParseStruct(&TestUser{})
	}
	
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	b.ResetTimer()
	
	b.Run("MemoryUsage_TagParser", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			user := &TestUser{
				Username: fmt.Sprintf("mem_test_%d", i),
				Email:    fmt.Sprintf("mem_%d@test.com", i),
				Age:      25,
			}
			suite.tagParser.ParseStruct(user)
		}
	})
	
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	
	b.Logf("Memory allocation per operation: %d bytes", (m2.TotalAlloc-m1.TotalAlloc)/uint64(b.N))
	b.Logf("Heap objects: %d", m2.HeapObjects)
	b.Logf("GC cycles: %d", m2.NumGC-m1.NumGC)
}

// BenchmarkLongRunningStability 基准测试: 长期运行稳定性
func BenchmarkLongRunningStability(b *testing.B) {
	suite, err := NewAnnotationBenchmarkSuite()
	if err != nil {
		b.Fatalf("Failed to create benchmark suite: %v", err)
	}
	defer suite.tester.Close()
	
	ctx := context.Background()
	duration := 30 * time.Second // 30秒长期测试
	
	b.Logf("开始长期稳定性测试，持续时间: %v", duration)
	
	var operations int64
	var errors int64
	var maxMemory uint64
	
	startTime := time.Now()
	endTime := startTime.Add(duration)
	
	var wg sync.WaitGroup
	
	// 启动多个工作协程
	for i := 0; i < 10; i++ {
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
				
				// 执行混合操作
				_, err := suite.tester.annotationSess.Insert(ctx, user)
				if err != nil {
					atomic.AddInt64(&errors, 1)
				}
				
				atomic.AddInt64(&operations, 1)
				localOps++
				
				// 定期检查内存使用
				if localOps%100 == 0 {
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					if m.Alloc > maxMemory {
						maxMemory = m.Alloc
					}
				}
			}
		}(i)
	}
	
	// 内存监控协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				if m.Alloc > maxMemory {
					maxMemory = m.Alloc
				}
			case <-time.After(duration):
				return
			}
		}
	}()
	
	wg.Wait()
	
	actualDuration := time.Since(startTime)
	
	b.Logf("长期稳定性测试完成:")
	b.Logf("- 实际测试时间: %v", actualDuration)
	b.Logf("- 总操作数: %d", operations)
	b.Logf("- 错误数: %d", errors)
	b.Logf("- 错误率: %.2f%%", float64(errors)/float64(operations)*100)
	b.Logf("- 平均QPS: %.2f", float64(operations)/actualDuration.Seconds())
	b.Logf("- 最大内存使用: %.2fMB", float64(maxMemory)/(1024*1024))
	
	// 验证稳定性指标
	errorRate := float64(errors) / float64(operations) * 100
	if errorRate > 5.0 { // 5%错误率阈值
		b.Errorf("Error rate too high: %.2f%%, expected <= 5%%", errorRate)
	}
	
	maxMemoryMB := float64(maxMemory) / (1024 * 1024)
	if maxMemoryMB > 100.0 { // 100MB内存阈值
		b.Errorf("Memory usage too high: %.2fMB, expected <= 100MB", maxMemoryMB)
	}
}

// Close 关闭基准测试套件
func (suite *AnnotationBenchmarkSuite) Close() error {
	return suite.tester.Close()
}