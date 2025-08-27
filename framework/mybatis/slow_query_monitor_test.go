// Package mybatis 慢查询监控系统测试
//
// 测试内容：
// 1. 慢查询检测功能测试
// 2. 告警通道测试
// 3. 优化规则引擎测试
// 4. 模式分析测试
// 5. 性能统计测试
package mybatis

import (
	"runtime"
	"strings"
	"testing"
	"time"
	"log"
	"fmt"
)

// TestSlowQueryMonitor_BasicDetection 测试基础慢查询检测
func TestSlowQueryMonitor_BasicDetection(t *testing.T) {
	config := DefaultSlowQueryConfig()
	config.SlowQueryThreshold = 50 * time.Millisecond
	
	monitor := NewSlowQueryMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 模拟正常查询
	normalExecution := &QueryExecution{
		SQL:        "SELECT id FROM users WHERE status = ?",
		Parameters: []any{"active"},
		StartTime:  time.Now(),
		Duration:   30 * time.Millisecond,
		Context:    map[string]any{"session_id": "test_session_1"},
	}
	normalExecution.EndTime = normalExecution.StartTime.Add(normalExecution.Duration)
	
	monitor.RecordQuery(normalExecution)
	
	// 模拟慢查询
	slowExecution := &QueryExecution{
		SQL:        "SELECT * FROM orders o JOIN users u ON o.user_id = u.id WHERE o.created_at > ?",
		Parameters: []any{"2024-01-01"},
		StartTime:  time.Now(),
		Duration:   200 * time.Millisecond,
		MemoryBefore: 1024 * 1024,     // 1MB
		MemoryAfter:  5 * 1024 * 1024, // 5MB
		Context:    map[string]any{"session_id": "test_session_2"},
	}
	slowExecution.EndTime = slowExecution.StartTime.Add(slowExecution.Duration)
	
	monitor.RecordQuery(slowExecution)
	
	// 等待处理完成
	time.Sleep(100 * time.Millisecond)
	
	// 验证统计信息
	stats := monitor.GetSlowQueryStats()
	if stats.TotalQueries < 2 {
		t.Errorf("Expected at least 2 total queries, got %d", stats.TotalQueries)
	}
	if stats.SlowQueries < 1 {
		t.Errorf("Expected at least 1 slow query, got %d", stats.SlowQueries)
	}
	
	// 验证慢查询记录
	slowQueries := monitor.GetSlowQueries(10, "")
	if len(slowQueries) < 1 {
		t.Errorf("Expected at least 1 slow query record, got %d", len(slowQueries))
	}
	
	// 验证第一个慢查询记录
	if len(slowQueries) > 0 {
		record := slowQueries[0]
		if record.QueryType != "SELECT" {
			t.Errorf("Expected query type SELECT, got %s", record.QueryType)
		}
		if len(record.TableNames) == 0 {
			t.Errorf("Expected table names to be extracted")
		}
		if record.Duration < config.SlowQueryThreshold {
			t.Errorf("Expected duration >= %v, got %v", config.SlowQueryThreshold, record.Duration)
		}
	}
	
	log.Printf("✓ Basic slow query detection test passed")
}

// TestSlowQueryMonitor_PatternAnalysis 测试查询模式分析
func TestSlowQueryMonitor_PatternAnalysis(t *testing.T) {
	config := DefaultSlowQueryConfig()
	config.SlowQueryThreshold = 10 * time.Millisecond
	config.PatternDetectionEnabled = true
	
	monitor := NewSlowQueryMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 创建相似的查询模式
	queries := []struct {
		sql    string
		params []any
	}{
		{"SELECT * FROM users WHERE id = ?", []any{1}},
		{"SELECT * FROM users WHERE id = ?", []any{2}},
		{"SELECT * FROM users WHERE id = ?", []any{3}},
		{"SELECT name, email FROM users WHERE status = ?", []any{"active"}},
		{"SELECT name, email FROM users WHERE status = ?", []any{"inactive"}},
	}
	
	for i, query := range queries {
		execution := &QueryExecution{
			SQL:        query.sql,
			Parameters: query.params,
			StartTime:  time.Now(),
			Duration:   time.Duration(20+i*10) * time.Millisecond,
			Context:    map[string]any{"session_id": fmt.Sprintf("session_%d", i)},
		}
		execution.EndTime = execution.StartTime.Add(execution.Duration)
		
		monitor.RecordQuery(execution)
		time.Sleep(10 * time.Millisecond) // 避免时间戳重复
	}
	
	// 等待处理完成
	time.Sleep(200 * time.Millisecond)
	
	// 验证模式分析
	patterns := monitor.GetQueryPatterns(10)
	if len(patterns) == 0 {
		t.Errorf("Expected query patterns to be detected, got 0")
	}
	
	// 应该有两个不同的模式
	expectedPatterns := 2
	if len(patterns) < expectedPatterns {
		t.Errorf("Expected at least %d patterns, got %d", expectedPatterns, len(patterns))
	}
	
	// 验证第一个模式（应该是最频繁的）
	if len(patterns) > 0 {
		topPattern := patterns[0]
		if topPattern.Frequency < 3 {
			t.Errorf("Expected top pattern frequency >= 3, got %d", topPattern.Frequency)
		}
		if topPattern.QueryType != "SELECT" {
			t.Errorf("Expected query type SELECT, got %s", topPattern.QueryType)
		}
	}
	
	log.Printf("✓ Query pattern analysis test passed - detected %d patterns", len(patterns))
}

// TestSlowQueryMonitor_AlertChannels 测试告警通道
func TestSlowQueryMonitor_AlertChannels(t *testing.T) {
	// 测试日志告警通道
	t.Run("LogAlertChannel", func(t *testing.T) {
		channel := NewLogAlertChannel("", "INFO")
		
		alert := &AlertRecord{
			ID:        "test_alert_1",
			Timestamp: time.Now(),
			Type:      "INSTANT",
			Severity:  "SLOW",
			Message:   "Test slow query alert",
			Details: map[string]any{
				"sql":      "SELECT * FROM test_table",
				"duration": "150ms",
			},
		}
		
		err := channel.SendAlert(alert)
		if err != nil {
			t.Errorf("Failed to send log alert: %v", err)
		}
		
		if channel.GetChannelType() != "LOG" {
			t.Errorf("Expected channel type LOG, got %s", channel.GetChannelType())
		}
		
		if !channel.IsEnabled() {
			t.Errorf("Expected channel to be enabled")
		}
	})
	
	// 测试控制台告警通道
	t.Run("ConsoleAlertChannel", func(t *testing.T) {
		channel := NewConsoleAlertChannel(true)
		
		alert := &AlertRecord{
			ID:        "test_alert_2",
			Timestamp: time.Now(),
			Type:      "INSTANT",
			Severity:  "CRITICAL",
			Message:   "Critical slow query detected",
			Details: map[string]any{
				"sql":      "SELECT * FROM large_table WHERE complex_condition",
				"duration": "2.5s",
			},
		}
		
		err := channel.SendAlert(alert)
		if err != nil {
			t.Errorf("Failed to send console alert: %v", err)
		}
		
		if channel.GetChannelType() != "CONSOLE" {
			t.Errorf("Expected channel type CONSOLE, got %s", channel.GetChannelType())
		}
	})
	
	// 测试Webhook告警通道
	t.Run("WebhookAlertChannel", func(t *testing.T) {
		config := &WebhookConfig{
			URL:        "https://httpbin.org/post", // 测试endpoint
			Method:     "POST",
			Headers:    map[string]string{"Content-Type": "application/json"},
			Timeout:    5 * time.Second,
			RetryCount: 1,
			RetryDelay: 1 * time.Second,
		}
		
		channel := NewWebhookAlertChannel(config)
		
		alert := &AlertRecord{
			ID:        "test_alert_3",
			Timestamp: time.Now(),
			Type:      "INSTANT",
			Severity:  "WARNING",
			Message:   "Webhook test alert",
			Details: map[string]any{
				"sql":      "UPDATE users SET last_login = NOW()",
				"duration": "800ms",
			},
		}
		
		// 注意：这个测试可能因为网络问题失败，在实际环境中应该使用mock
		err := channel.SendAlert(alert)
		if err != nil {
			t.Logf("Webhook alert failed (expected in test environment): %v", err)
		}
		
		if channel.GetChannelType() != "WEBHOOK" {
			t.Errorf("Expected channel type WEBHOOK, got %s", channel.GetChannelType())
		}
	})
	
	log.Printf("✓ Alert channels test completed")
}

// TestSlowQueryMonitor_OptimizationRules 测试优化规则
func TestSlowQueryMonitor_OptimizationRules(t *testing.T) {
	engine := NewOptimizationRuleEngine()
	
	// 测试SELECT * 优化规则
	t.Run("SelectStarOptimization", func(t *testing.T) {
		record := &SlowQueryRecord{
			ID:           "test_1",
			SQL:          "SELECT * FROM users WHERE status = 'active'",
			NormalizedSQL: "SELECT * FROM users WHERE status = ?",
			Duration:     150 * time.Millisecond,
			QueryType:    "SELECT",
			TableNames:   []string{"users"},
			Severity:     "SLOW",
		}
		
		suggestions := engine.AnalyzeQuery(record)
		
		found := false
		for _, suggestion := range suggestions {
			if strings.Contains(strings.ToLower(suggestion), "select *") {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Expected SELECT * optimization suggestion")
		}
	})
	
	// 测试索引建议规则
	t.Run("IndexSuggestion", func(t *testing.T) {
		record := &SlowQueryRecord{
			ID:           "test_2",
			SQL:          "SELECT name, email FROM users WHERE created_at > '2024-01-01' ORDER BY created_at",
			NormalizedSQL: "SELECT name, email FROM users WHERE created_at > ? ORDER BY created_at",
			Duration:     300 * time.Millisecond,
			QueryType:    "SELECT",
			TableNames:   []string{"users"},
			Severity:     "VERY_SLOW",
		}
		
		suggestions := engine.AnalyzeQuery(record)
		
		foundIndex := false
		foundOrderBy := false
		for _, suggestion := range suggestions {
			lower := strings.ToLower(suggestion)
			if strings.Contains(lower, "index") && strings.Contains(lower, "where") {
				foundIndex = true
			}
			if strings.Contains(lower, "index") && strings.Contains(lower, "order by") {
				foundOrderBy = true
			}
		}
		
		if !foundIndex {
			t.Errorf("Expected WHERE clause index suggestion")
		}
		if !foundOrderBy {
			t.Errorf("Expected ORDER BY index suggestion")
		}
	})
	
	// 测试JOIN优化建议
	t.Run("JoinOptimization", func(t *testing.T) {
		record := &SlowQueryRecord{
			ID:           "test_3",
			SQL:          "SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id WHERE o.status = 'completed'",
			NormalizedSQL: "SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id WHERE o.status = ?",
			Duration:     500 * time.Millisecond,
			QueryType:    "SELECT",
			TableNames:   []string{"users", "orders"},
			Severity:     "VERY_SLOW",
		}
		
		suggestions := engine.AnalyzeQuery(record)
		
		foundJoin := false
		for _, suggestion := range suggestions {
			if strings.Contains(strings.ToLower(suggestion), "join") {
				foundJoin = true
				break
			}
		}
		
		if !foundJoin {
			t.Errorf("Expected JOIN optimization suggestion")
		}
	})
	
	// 测试性能规则
	t.Run("PerformanceRule", func(t *testing.T) {
		record := &SlowQueryRecord{
			ID:           "test_4",
			SQL:          "SELECT COUNT(*) FROM large_table WHERE complex_condition LIKE '%pattern%'",
			NormalizedSQL: "SELECT COUNT(*) FROM large_table WHERE complex_condition LIKE ?",
			Duration:     2 * time.Second,
			MemoryUsed:   100 * 1024 * 1024, // 100MB
			QueryType:    "SELECT",
			TableNames:   []string{"large_table"},
			Severity:     "CRITICAL",
		}
		
		suggestions := engine.AnalyzeQuery(record)
		
		foundPerformance := false
		for _, suggestion := range suggestions {
			lower := strings.ToLower(suggestion)
			if strings.Contains(lower, "time") || strings.Contains(lower, "memory") {
				foundPerformance = true
				break
			}
		}
		
		if !foundPerformance {
			t.Errorf("Expected performance optimization suggestion")
		}
	})
	
	log.Printf("✓ Optimization rules test passed with %d suggestions generated", len(engine.rules))
}

// TestSlowQueryMonitor_AdaptiveThreshold 测试自适应阈值
func TestSlowQueryMonitor_AdaptiveThreshold(t *testing.T) {
	config := DefaultSlowQueryConfig()
	config.SlowQueryThreshold = 100 * time.Millisecond
	config.EnableAdaptiveThreshold = true
	config.AdaptiveInterval = 1 * time.Second
	
	monitor := NewSlowQueryMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 模拟一系列查询以建立基线
	baselineQueries := []time.Duration{
		50 * time.Millisecond,
		80 * time.Millisecond,
		120 * time.Millisecond,
		90 * time.Millisecond,
		110 * time.Millisecond,
	}
	
	for i, duration := range baselineQueries {
		execution := &QueryExecution{
			SQL:        "SELECT * FROM baseline_table WHERE id = ?",
			Parameters: []any{i},
			StartTime:  time.Now(),
			Duration:   duration,
			Context:    map[string]any{"session_id": fmt.Sprintf("baseline_%d", i)},
		}
		execution.EndTime = execution.StartTime.Add(execution.Duration)
		
		monitor.RecordQuery(execution)
		time.Sleep(10 * time.Millisecond)
	}
	
	// 等待自适应调整
	time.Sleep(1200 * time.Millisecond)
	
	// 验证阈值调整（这里主要是验证功能不报错）
	stats := monitor.GetSlowQueryStats()
	if stats.TotalQueries == 0 {
		t.Errorf("Expected queries to be recorded for adaptive threshold testing")
	}
	
	log.Printf("✓ Adaptive threshold test completed - processed %d queries", stats.TotalQueries)
}

// TestSlowQueryMonitor_ConcurrentAccess 测试并发访问
func TestSlowQueryMonitor_ConcurrentAccess(t *testing.T) {
	config := DefaultSlowQueryConfig()
	config.SlowQueryThreshold = 50 * time.Millisecond
	
	monitor := NewSlowQueryMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 并发执行查询记录
	const numGoroutines = 10
	const queriesPerGoroutine = 20
	
	done := make(chan bool, numGoroutines)
	
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			
			for i := 0; i < queriesPerGoroutine; i++ {
				execution := &QueryExecution{
					SQL:        fmt.Sprintf("SELECT * FROM table_%d WHERE id = ?", goroutineID),
					Parameters: []any{i},
					StartTime:  time.Now(),
					Duration:   time.Duration(60+i*5) * time.Millisecond,
					Context:    map[string]any{
						"session_id":    fmt.Sprintf("concurrent_%d_%d", goroutineID, i),
						"goroutine_id":  goroutineID,
					},
				}
				execution.EndTime = execution.StartTime.Add(execution.Duration)
				
				monitor.RecordQuery(execution)
				
				// 随机延迟模拟真实场景
				time.Sleep(time.Duration(i%10) * time.Millisecond)
			}
		}(g)
	}
	
	// 等待所有goroutine完成
	for g := 0; g < numGoroutines; g++ {
		<-done
	}
	
	// 等待处理完成
	time.Sleep(500 * time.Millisecond)
	
	// 验证结果
	stats := monitor.GetSlowQueryStats()
	expectedTotal := int64(numGoroutines * queriesPerGoroutine)
	
	if stats.TotalQueries < expectedTotal {
		t.Errorf("Expected at least %d total queries, got %d", expectedTotal, stats.TotalQueries)
	}
	
	if stats.SlowQueries == 0 {
		t.Errorf("Expected some slow queries to be detected")
	}
	
	// 验证模式分析
	patterns := monitor.GetQueryPatterns(20)
	if len(patterns) == 0 {
		t.Errorf("Expected query patterns to be detected")
	}
	
	log.Printf("✓ Concurrent access test passed - processed %d queries, detected %d patterns", 
		stats.TotalQueries, len(patterns))
}

// TestSlowQueryMonitor_MemoryUsage 测试内存使用
func TestSlowQueryMonitor_MemoryUsage(t *testing.T) {
	// 记录初始内存
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	config := DefaultSlowQueryConfig()
	config.SlowQueryThreshold = 10 * time.Millisecond
	config.MaxSlowQueryRecords = 1000 // 限制记录数量
	
	monitor := NewSlowQueryMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 生成大量查询记录
	const numQueries = 2000
	
	for i := 0; i < numQueries; i++ {
		execution := &QueryExecution{
			SQL:        fmt.Sprintf("SELECT * FROM memory_test_table_%d WHERE column_%d = ?", i%10, i%5),
			Parameters: []any{i},
			StartTime:  time.Now(),
			Duration:   time.Duration(20+i%100) * time.Millisecond,
			Context:    map[string]any{
				"session_id": fmt.Sprintf("memory_test_%d", i),
				"data":      strings.Repeat("x", 100), // 添加一些数据
			},
		}
		execution.EndTime = execution.StartTime.Add(execution.Duration)
		
		monitor.RecordQuery(execution)
		
		if i%500 == 0 {
			runtime.GC() // 定期垃圾回收
		}
	}
	
	// 等待处理完成
	time.Sleep(1 * time.Second)
	
	// 检查内存使用
	runtime.ReadMemStats(&m2)
	memoryIncrease := m2.Alloc - m1.Alloc
	
	// 验证记录数量不超过限制
	stats := monitor.GetSlowQueryStats()
	slowQueries := monitor.GetSlowQueries(2000, "")
	
	if len(slowQueries) > config.MaxSlowQueryRecords {
		t.Errorf("Expected slow query records <= %d, got %d", config.MaxSlowQueryRecords, len(slowQueries))
	}
	
	if stats.TotalQueries == 0 {
		t.Errorf("Expected some queries to be tracked for memory test")
	}
	
	log.Printf("✓ Memory usage test completed - Memory increase: %d bytes, Records: %d", 
		memoryIncrease, len(slowQueries))
	
	// 内存增长应该合理（这里设置一个宽松的限制）
	maxExpectedMemory := uint64(50 * 1024 * 1024) // 50MB
	if memoryIncrease > maxExpectedMemory {
		t.Errorf("Memory usage too high: %d bytes (max expected: %d bytes)", 
			memoryIncrease, maxExpectedMemory)
	}
}

// TestSlowQueryMonitor_Integration 集成测试
func TestSlowQueryMonitor_Integration(t *testing.T) {
	// 创建完整配置的监控器
	config := DefaultSlowQueryConfig()
	config.SlowQueryThreshold = 30 * time.Millisecond
	config.VerySlowThreshold = 100 * time.Millisecond
	config.EnableInstantAlert = true
	config.EnableBatchAlert = true
	config.BatchAlertInterval = 2 * time.Second
	config.PatternDetectionEnabled = true
	config.EnableOptimizationTips = true
	
	monitor := NewSlowQueryMonitor(config)
	
	// 添加告警通道
	logChannel := NewLogAlertChannel("", "INFO")
	consoleChannel := NewConsoleAlertChannel(true)
	
	monitor.alertManager.alertChannels = append(monitor.alertManager.alertChannels, logChannel, consoleChannel)
	
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 模拟各种类型的查询
	testScenarios := []struct {
		name     string
		sql      string
		params   []any
		duration time.Duration
		memory   int64
	}{
		{
			name:     "Fast Query",
			sql:      "SELECT id FROM users WHERE id = ?",
			params:   []any{1},
			duration: 20 * time.Millisecond,
			memory:   1024,
		},
		{
			name:     "Slow SELECT *",
			sql:      "SELECT * FROM users WHERE status = ?",
			params:   []any{"active"},
			duration: 80 * time.Millisecond,
			memory:   10240,
		},
		{
			name:     "Very Slow JOIN",
			sql:      "SELECT u.*, o.* FROM users u JOIN orders o ON u.id = o.user_id WHERE o.created_at > ?",
			params:   []any{"2024-01-01"},
			duration: 250 * time.Millisecond,
			memory:   102400,
		},
		{
			name:     "Critical Query",
			sql:      "SELECT COUNT(*) FROM large_table WHERE complex_condition LIKE ?",
			params:   []any{"%pattern%"},
			duration: 800 * time.Millisecond,
			memory:   1048576,
		},
		{
			name:     "Batch INSERT",
			sql:      "INSERT INTO logs (message, created_at) VALUES (?, NOW())",
			params:   []any{"Log message"},
			duration: 60 * time.Millisecond,
			memory:   2048,
		},
	}
	
	// 执行测试场景
	for i, scenario := range testScenarios {
		execution := &QueryExecution{
			SQL:          scenario.sql,
			Parameters:   scenario.params,
			StartTime:    time.Now(),
			Duration:     scenario.duration,
			MemoryBefore: 1024 * 1024,
			MemoryAfter:  1024*1024 + scenario.memory,
			Context: map[string]any{
				"session_id": fmt.Sprintf("integration_test_%d", i),
				"scenario":   scenario.name,
			},
		}
		execution.EndTime = execution.StartTime.Add(execution.Duration)
		
		monitor.RecordQuery(execution)
		time.Sleep(50 * time.Millisecond) // 间隔执行
	}
	
	// 等待所有处理完成
	time.Sleep(3 * time.Second)
	
	// 验证整体结果
	stats := monitor.GetSlowQueryStats()
	
	// 验证基础统计
	if stats.TotalQueries != int64(len(testScenarios)) {
		t.Errorf("Expected %d total queries, got %d", len(testScenarios), stats.TotalQueries)
	}
	
	// 应该检测到慢查询
	if stats.SlowQueries == 0 {
		t.Errorf("Expected slow queries to be detected")
	}
	
	// 验证严重程度分类
	if stats.VerySlowQueries == 0 {
		t.Errorf("Expected very slow queries to be detected")
	}
	
	if stats.CriticalQueries == 0 {
		t.Errorf("Expected critical queries to be detected")
	}
	
	// 验证模式检测
	patterns := monitor.GetQueryPatterns(10)
	if len(patterns) == 0 {
		t.Errorf("Expected query patterns to be detected")
	}
	
	// 验证优化建议
	slowQueries := monitor.GetSlowQueries(10, "")
	foundOptimizations := false
	for _, record := range slowQueries {
		if len(record.OptimizationTips) > 0 {
			foundOptimizations = true
			break
		}
	}
	
	if !foundOptimizations {
		t.Errorf("Expected optimization tips to be generated")
	}
	
	// 打印详细结果
	log.Printf("=== Integration Test Results ===")
	log.Printf("Total Queries: %d", stats.TotalQueries)
	log.Printf("Slow Queries: %d", stats.SlowQueries)
	log.Printf("Very Slow Queries: %d", stats.VerySlowQueries)
	log.Printf("Critical Queries: %d", stats.CriticalQueries)
	log.Printf("Query Patterns: %d", len(patterns))
	log.Printf("Average Slow Query Time: %v", stats.AvgSlowQueryTime)
	log.Printf("Max Slow Query Time: %v", stats.MaxSlowQueryTime)
	log.Printf("Total Alerts: %d", stats.TotalAlerts)
	
	if len(patterns) > 0 {
		log.Printf("Top Query Pattern: %s (Frequency: %d)", 
			patterns[0].NormalizedSQL, patterns[0].Frequency)
	}
	
	log.Printf("✓ Integration test completed successfully")
}

// BenchmarkSlowQueryMonitor_Performance 性能基准测试
func BenchmarkSlowQueryMonitor_Performance(b *testing.B) {
	config := DefaultSlowQueryConfig()
	config.SlowQueryThreshold = 1 * time.Millisecond
	
	monitor := NewSlowQueryMonitor(config)
	err := monitor.Start()
	if err != nil {
		b.Fatalf("Failed to start monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 准备测试查询
	testQuery := &QueryExecution{
		SQL:        "SELECT * FROM benchmark_table WHERE id = ?",
		Parameters: []any{1},
		StartTime:  time.Now(),
		Duration:   10 * time.Millisecond,
		Context:    map[string]any{"benchmark": true},
	}
	testQuery.EndTime = testQuery.StartTime.Add(testQuery.Duration)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// 基准测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			monitor.RecordQuery(testQuery)
		}
	})
}