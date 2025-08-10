package main

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/util/mapper"
)

// 性能测试套件

// TestData 测试数据结构
type TestData struct {
	SimpleStruct  *SimpleTestStruct
	ComplexStruct *ComplexTestStruct
	LargeSlice    []*SimpleTestStruct
}

type SimpleTestStruct struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Score     float64   `json:"score"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type ComplexTestStruct struct {
	ID       int                        `json:"id"`
	Name     string                     `json:"name"`
	Profile  Profile                    `json:"profile"`
	Tags     []string                   `json:"tags"`
	Metadata map[string]interface{}     `json:"metadata"`
	Items    []Item                     `json:"items"`
	Settings *Settings                  `json:"settings"`
	History  []map[string]interface{}   `json:"history"`
}

type Profile struct {
	Avatar    string            `json:"avatar"`
	Bio       string            `json:"bio"`
	Links     []string          `json:"links"`
	Contacts  map[string]string `json:"contacts"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type Item struct {
	ID          string             `json:"id"`
	Type        string             `json:"type"`
	Properties  map[string]interface{} `json:"properties"`
	SubItems    []SubItem          `json:"sub_items"`
	CreatedAt   time.Time          `json:"created_at"`
}

type SubItem struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Tags  []string    `json:"tags"`
}

type Settings struct {
	Theme         string            `json:"theme"`
	Language      string            `json:"language"`
	Preferences   map[string]bool   `json:"preferences"`
	Limits        map[string]int    `json:"limits"`
	LastModified  time.Time         `json:"last_modified"`
}

// 目标结构体
type SimpleTarget struct {
	ID        int       `json:"id"`
	FullName  string    `json:"name"`
	EmailAddr string    `json:"email"`
	Years     int       `json:"age"`
	Rating    float64   `json:"score"`
	IsActive  bool      `json:"active"`
	Timestamp time.Time `json:"created_at"`
}

type ComplexTarget struct {
	ID       int                        `json:"id"`
	Title    string                     `json:"name"`
	Profile  Profile                    `json:"profile"`
	Labels   []string                   `json:"tags"`
	Data     map[string]interface{}     `json:"metadata"`
	Products []Item                     `json:"items"`
	Config   *Settings                  `json:"settings"`
	Records  []map[string]interface{}   `json:"history"`
}

// 生成测试数据
func generateTestData() *TestData {
	now := time.Now()
	
	return &TestData{
		SimpleStruct: &SimpleTestStruct{
			ID:        12345,
			Name:      "测试用户",
			Email:     "test@example.com",
			Age:       30,
			Score:     95.5,
			Active:    true,
			CreatedAt: now,
		},
		ComplexStruct: &ComplexTestStruct{
			ID:   67890,
			Name: "复杂用户",
			Profile: Profile{
				Avatar: "https://example.com/avatar.jpg",
				Bio:    "这是一个复杂的用户资料，包含了各种信息",
				Links:  []string{"https://github.com", "https://twitter.com"},
				Contacts: map[string]string{
					"email":  "complex@example.com",
					"phone":  "+86-138-0000-0000",
					"wechat": "complex_user",
				},
				UpdatedAt: now.Add(-1 * time.Hour),
			},
			Tags: []string{"开发者", "架构师", "技术专家", "团队负责人"},
			Metadata: map[string]interface{}{
				"department":     "技术部",
				"level":         "P8",
				"years_exp":     10,
				"team_size":     15,
				"projects":      []string{"项目A", "项目B", "项目C"},
				"skills":        []string{"Go", "Python", "Java", "Kubernetes", "Docker"},
				"certifications": map[string]interface{}{
					"aws":   true,
					"azure": false,
					"gcp":   true,
				},
			},
			Items: []Item{
				{
					ID:   "item-001",
					Type: "document",
					Properties: map[string]interface{}{
						"size":     1024000,
						"format":   "PDF",
						"pages":    50,
						"editable": true,
					},
					SubItems: []SubItem{
						{Name: "章节1", Value: "介绍", Tags: []string{"intro", "基础"}},
						{Name: "章节2", Value: "实践", Tags: []string{"practice", "高级"}},
					},
					CreatedAt: now.Add(-24 * time.Hour),
				},
				{
					ID:   "item-002",
					Type: "image",
					Properties: map[string]interface{}{
						"width":  1920,
						"height": 1080,
						"format": "PNG",
						"size":   2048000,
					},
					SubItems: []SubItem{
						{Name: "缩略图", Value: "thumbnail.png", Tags: []string{"thumb"}},
					},
					CreatedAt: now.Add(-2 * time.Hour),
				},
			},
			Settings: &Settings{
				Theme:    "dark",
				Language: "zh-CN",
				Preferences: map[string]bool{
					"email_notifications": true,
					"sms_notifications":   false,
					"push_notifications":  true,
					"auto_save":          true,
				},
				Limits: map[string]int{
					"max_files":     1000,
					"max_size_mb":   5120,
					"max_duration":  3600,
				},
				LastModified: now.Add(-30 * time.Minute),
			},
			History: []map[string]interface{}{
				{
					"action":    "login",
					"timestamp": now.Add(-8 * time.Hour),
					"ip":        "192.168.1.100",
					"device":    "Chrome/114.0",
				},
				{
					"action":    "update_profile",
					"timestamp": now.Add(-6 * time.Hour),
					"changes":   []string{"avatar", "bio"},
				},
				{
					"action":    "create_item",
					"timestamp": now.Add(-4 * time.Hour),
					"item_id":   "item-001",
					"item_type": "document",
				},
			},
		},
		LargeSlice: make([]*SimpleTestStruct, 1000),
	}
}

// 初始化大切片数据
func initLargeSliceData(data *TestData) {
	for i := range data.LargeSlice {
		data.LargeSlice[i] = &SimpleTestStruct{
			ID:        i + 1,
			Name:      fmt.Sprintf("用户%d", i+1),
			Email:     fmt.Sprintf("user%d@example.com", i+1),
			Age:       20 + i%50,
			Score:     60.0 + float64(i%40),
			Active:    i%3 != 0,
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}
}

// 性能基准测试

func BenchmarkYYHertzMapper_Simple_Auto(b *testing.B) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyAuto))
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var target SimpleTarget
		err := mapperInstance.Map(data.SimpleStruct, &target)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYYHertzMapper_Simple_Reflection(b *testing.B) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyReflection))
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var target SimpleTarget
		err := mapperInstance.Map(data.SimpleStruct, &target)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYYHertzMapper_Simple_JSON(b *testing.B) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyJSON))
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var target SimpleTarget
		err := mapperInstance.Map(data.SimpleStruct, &target)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYYHertzMapper_Complex_Auto(b *testing.B) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyAuto))
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var target ComplexTarget
		err := mapperInstance.Map(data.ComplexStruct, &target)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYYHertzMapper_Complex_Reflection(b *testing.B) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyReflection))
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var target ComplexTarget
		err := mapperInstance.Map(data.ComplexStruct, &target)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYYHertzMapper_Complex_JSON(b *testing.B) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyJSON))
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var target ComplexTarget
		err := mapperInstance.Map(data.ComplexStruct, &target)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYYHertzMapper_SliceMapping(b *testing.B) {
	data := generateTestData()
	initLargeSliceData(data)
	
	mapperInstance := mapper.NewMapper()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var targets []SimpleTarget
		err := mapperInstance.MapSlice(data.LargeSlice, &targets)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 对比基准测试 - 纯JSON方案
func BenchmarkPureJSON_Simple(b *testing.B) {
	data := generateTestData()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		jsonData, err := json.Marshal(data.SimpleStruct)
		if err != nil {
			b.Fatal(err)
		}
		
		var target SimpleTarget
		err = json.Unmarshal(jsonData, &target)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPureJSON_Complex(b *testing.B) {
	data := generateTestData()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		jsonData, err := json.Marshal(data.ComplexStruct)
		if err != nil {
			b.Fatal(err)
		}
		
		var target ComplexTarget
		err = json.Unmarshal(jsonData, &target)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 手动映射基准测试（理论性能上限）
func BenchmarkManualMapping_Simple(b *testing.B) {
	data := generateTestData()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		target := SimpleTarget{
			ID:        data.SimpleStruct.ID,
			FullName:  data.SimpleStruct.Name,
			EmailAddr: data.SimpleStruct.Email,
			Years:     data.SimpleStruct.Age,
			Rating:    data.SimpleStruct.Score,
			IsActive:  data.SimpleStruct.Active,
			Timestamp: data.SimpleStruct.CreatedAt,
		}
		_ = target
	}
}

func BenchmarkManualMapping_Complex(b *testing.B) {
	data := generateTestData()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		target := ComplexTarget{
			ID:       data.ComplexStruct.ID,
			Title:    data.ComplexStruct.Name,
			Profile:  data.ComplexStruct.Profile,
			Labels:   data.ComplexStruct.Tags,
			Data:     data.ComplexStruct.Metadata,
			Products: data.ComplexStruct.Items,
			Config:   data.ComplexStruct.Settings,
			Records:  data.ComplexStruct.History,
		}
		_ = target
	}
}

// 并发性能测试
func BenchmarkYYHertzMapper_Concurrent(b *testing.B) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var target SimpleTarget
			err := mapperInstance.Map(data.SimpleStruct, &target)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 内存使用测试
func TestMemoryUsage(t *testing.T) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper()
	
	// 获取初始内存状态
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	// 执行大量映射操作
	iterations := 10000
	for i := 0; i < iterations; i++ {
		var target SimpleTarget
		err := mapperInstance.Map(data.SimpleStruct, &target)
		if err != nil {
			t.Fatal(err)
		}
	}
	
	// 获取最终内存状态
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	
	allocatedMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
	avgAllocPerOp := float64(m2.TotalAlloc-m1.TotalAlloc) / float64(iterations)
	
	t.Logf("内存使用统计:")
	t.Logf("  总分配: %.2f MB", allocatedMB)
	t.Logf("  每次操作平均分配: %.2f bytes", avgAllocPerOp)
	t.Logf("  GC次数: %d", m2.NumGC-m1.NumGC)
	
	// 获取缓存统计
	stats := mapperInstance.GetStats()
	t.Logf("  缓存命中率: %.2f%%", 
		float64(stats.CacheHits)/float64(stats.CacheHits+stats.CacheMisses)*100)
}

// 压力测试
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}
	
	data := generateTestData()
	initLargeSliceData(data)
	
	mapperInstance := mapper.NewMapper()
	
	// 并发压力测试
	concurrency := 100
	iterations := 1000
	
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)
	
	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < iterations; j++ {
				var target SimpleTarget
				err := mapperInstance.Map(data.SimpleStruct, &target)
				if err != nil {
					errors <- fmt.Errorf("worker %d iteration %d failed: %v", workerID, j, err)
					return
				}
			}
		}(i)
	}
	
	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(errors)
	}()
	
	// 检查错误
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
	}
	
	duration := time.Since(start)
	totalOps := concurrency * iterations
	opsPerSec := float64(totalOps) / duration.Seconds()
	
	t.Logf("压力测试结果:")
	t.Logf("  并发数: %d", concurrency)
	t.Logf("  每个goroutine操作数: %d", iterations)
	t.Logf("  总操作数: %d", totalOps)
	t.Logf("  总耗时: %v", duration)
	t.Logf("  吞吐量: %.2f ops/sec", opsPerSec)
	t.Logf("  错误数: %d", errorCount)
	
	// 获取最终统计
	stats := mapperInstance.GetStats()
	t.Logf("  成功率: %.2f%%", float64(stats.SuccessfulMaps)/float64(stats.TotalMaps)*100)
	t.Logf("  缓存命中率: %.2f%%", 
		float64(stats.CacheHits)/float64(stats.CacheHits+stats.CacheMisses)*100)
	
	if errorCount > totalOps/100 { // 错误率超过1%
		t.Errorf("压力测试失败率过高: %d/%d (%.2f%%)", 
			errorCount, totalOps, float64(errorCount)/float64(totalOps)*100)
	}
}

// 长期稳定性测试
func TestLongTermStability(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过长期稳定性测试")
	}
	
	data := generateTestData()
	mapperInstance := mapper.NewMapper()
	
	// 运行30秒的持续测试
	duration := 30 * time.Second
	start := time.Now()
	count := 0
	errors := 0
	
	for time.Since(start) < duration {
		var target SimpleTarget
		err := mapperInstance.Map(data.SimpleStruct, &target)
		if err != nil {
			errors++
		}
		count++
		
		// 每1000次操作检查一次内存
		if count%1000 == 0 {
			runtime.GC()
		}
	}
	
	actualDuration := time.Since(start)
	opsPerSec := float64(count) / actualDuration.Seconds()
	
	t.Logf("长期稳定性测试结果:")
	t.Logf("  运行时间: %v", actualDuration)
	t.Logf("  总操作数: %d", count)
	t.Logf("  错误数: %d", errors)
	t.Logf("  平均吞吐量: %.2f ops/sec", opsPerSec)
	t.Logf("  错误率: %.6f%%", float64(errors)/float64(count)*100)
	
	if errors > count/1000 { // 错误率超过0.1%
		t.Errorf("长期稳定性测试失败率过高: %d/%d", errors, count)
	}
}

// 性能回归检测
func TestPerformanceRegression(t *testing.T) {
	data := generateTestData()
	mapperInstance := mapper.NewMapper()
	
	// 基准性能阈值（这些数值需要根据实际情况调整）
	thresholds := map[string]time.Duration{
		"SimpleMapping": 5 * time.Microsecond,   // 简单映射应在5微秒内完成
		"ComplexMapping": 50 * time.Microsecond, // 复杂映射应在50微秒内完成
	}
	
	// 测试简单映射性能
	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		var target SimpleTarget
		err := mapperInstance.Map(data.SimpleStruct, &target)
		if err != nil {
			t.Fatal(err)
		}
	}
	simpleAvg := time.Since(start) / time.Duration(iterations)
	
	// 测试复杂映射性能
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var target ComplexTarget
		err := mapperInstance.Map(data.ComplexStruct, &target)
		if err != nil {
			t.Fatal(err)
		}
	}
	complexAvg := time.Since(start) / time.Duration(iterations)
	
	t.Logf("性能回归检测结果:")
	t.Logf("  简单映射平均耗时: %v (阈值: %v)", simpleAvg, thresholds["SimpleMapping"])
	t.Logf("  复杂映射平均耗时: %v (阈值: %v)", complexAvg, thresholds["ComplexMapping"])
	
	// 检查是否超过阈值
	if simpleAvg > thresholds["SimpleMapping"] {
		t.Errorf("简单映射性能回归: %v > %v", simpleAvg, thresholds["SimpleMapping"])
	}
	
	if complexAvg > thresholds["ComplexMapping"] {
		t.Errorf("复杂映射性能回归: %v > %v", complexAvg, thresholds["ComplexMapping"])
	}
}

// 运行完整的性能报告
func TestFullPerformanceReport(t *testing.T) {
	log.Println("=== YYHertz 高性能对象映射库 - 完整性能报告 ===")
	
	data := generateTestData()
	initLargeSliceData(data)
	
	strategies := []struct {
		name     string
		strategy mapper.MappingStrategy
	}{
		{"自动选择", mapper.StrategyAuto},
		{"反射映射", mapper.StrategyReflection},
		{"JSON映射", mapper.StrategyJSON},
	}
	
	iterations := 10000
	
	log.Printf("\n1. 简单结构体映射性能测试 (%d次迭代):", iterations)
	log.Printf("%-12s | %-12s | %-15s | %-12s", "策略", "总耗时", "平均耗时", "吞吐量")
	log.Printf("-----------|-------------|----------------|------------")
	
	for _, s := range strategies {
		mapperInstance := mapper.NewMapper(mapper.WithStrategy(s.strategy))
		
		// 预热
		var warmup SimpleTarget
		mapperInstance.Map(data.SimpleStruct, &warmup)
		
		start := time.Now()
		for i := 0; i < iterations; i++ {
			var target SimpleTarget
			err := mapperInstance.Map(data.SimpleStruct, &target)
			if err != nil && s.strategy != mapper.StrategyCodegen {
				t.Errorf("%s策略失败: %v", s.name, err)
				break
			}
		}
		duration := time.Since(start)
		avgTime := duration / time.Duration(iterations)
		opsPerSec := int64(iterations) * int64(time.Second) / int64(duration)
		
		log.Printf("%-12s | %-12v | %-15v | %8d ops/s", 
			s.name, duration, avgTime, opsPerSec)
	}
	
	log.Printf("\n2. 复杂结构体映射性能测试 (%d次迭代):", iterations/10)
	log.Printf("%-12s | %-12s | %-15s | %-12s", "策略", "总耗时", "平均耗时", "吞吐量")
	log.Printf("-----------|-------------|----------------|------------")
	
	for _, s := range strategies {
		mapperInstance := mapper.NewMapper(mapper.WithStrategy(s.strategy))
		
		// 预热
		var warmup ComplexTarget
		mapperInstance.Map(data.ComplexStruct, &warmup)
		
		smallIterations := iterations / 10
		start := time.Now()
		for i := 0; i < smallIterations; i++ {
			var target ComplexTarget
			err := mapperInstance.Map(data.ComplexStruct, &target)
			if err != nil && s.strategy != mapper.StrategyCodegen {
				t.Errorf("%s策略失败: %v", s.name, err)
				break
			}
		}
		duration := time.Since(start)
		avgTime := duration / time.Duration(smallIterations)
		opsPerSec := int64(smallIterations) * int64(time.Second) / int64(duration)
		
		log.Printf("%-12s | %-12v | %-15v | %8d ops/s", 
			s.name, duration, avgTime, opsPerSec)
	}
	
	// 批量映射测试
	log.Printf("\n3. 批量映射性能测试 (1000条记录):")
	mapperInstance := mapper.NewMapper()
	start := time.Now()
	var targets []SimpleTarget
	err := mapperInstance.MapSlice(data.LargeSlice, &targets)
	duration := time.Since(start)
	
	if err != nil {
		t.Errorf("批量映射失败: %v", err)
	} else {
		recordsPerSec := int64(len(data.LargeSlice)) * int64(time.Second) / int64(duration)
		log.Printf("批量映射: %v (%.2f records/ms, %d records/sec)", 
			duration, float64(len(data.LargeSlice))/duration.Seconds()*1000, recordsPerSec)
	}
	
	// 缓存统计
	log.Printf("\n4. 缓存统计信息:")
	stats := mapperInstance.GetStats()
	log.Printf("总映射次数: %d", stats.TotalMaps)
	log.Printf("成功映射: %d", stats.SuccessfulMaps)
	log.Printf("失败映射: %d", stats.FailedMaps)
	log.Printf("缓存命中: %d", stats.CacheHits)
	log.Printf("缓存未命中: %d", stats.CacheMisses)
	
	if stats.TotalMaps > 0 {
		successRate := float64(stats.SuccessfulMaps) / float64(stats.TotalMaps) * 100
		log.Printf("成功率: %.2f%%", successRate)
	}
	
	if stats.CacheHits+stats.CacheMisses > 0 {
		hitRate := float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses) * 100
		log.Printf("缓存命中率: %.2f%%", hitRate)
	}
	
	log.Printf("\n=== 性能报告完成 ===")
}