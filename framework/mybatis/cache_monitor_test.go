// Package mybatis 缓存监控系统测试
//
// 测试内容：
// 1. 多级缓存命中率监控测试
// 2. 缓存热点分析测试
// 3. 预热引擎功能测试
// 4. 实时指标收集测试
// 5. 缓存性能报告生成测试
package mybatis

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// TestCacheHitRateMonitor_BasicFunctionality 测试缓存监控基础功能
func TestCacheHitRateMonitor_BasicFunctionality(t *testing.T) {
	config := DefaultCacheMonitorConfig()
	config.MonitorInterval = 100 * time.Millisecond // 加快测试速度
	
	monitor := NewCacheHitRateMonitor(config)
	
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 模拟L1缓存访问
	testKeys := []string{
		"user:1", "user:2", "user:3", "order:100", "order:101",
		"product:50", "product:51", "session:abc", "session:def",
	}
	
	// 模拟缓存访问模式
	for i := 0; i < 100; i++ {
		key := testKeys[i%len(testKeys)]
		hit := (i%3) != 0 // 约67%的命中率
		duration := time.Duration(10+rand.Intn(20)) * time.Millisecond
		
		monitor.RecordCacheAccess("L1", key, hit, duration)
		
		if i%10 == 0 {
			time.Sleep(10 * time.Millisecond) // 模拟真实间隔
		}
	}
	
	// 等待监控器处理数据
	time.Sleep(200 * time.Millisecond)
	
	// 验证L1缓存统计
	l1Stats := monitor.GetLayerStats("L1")
	if l1Stats == nil {
		t.Fatalf("L1 cache stats should not be nil")
	}
	
	if l1Stats.TotalRequests != 100 {
		t.Errorf("Expected 100 total requests, got %d", l1Stats.TotalRequests)
	}
	
	if l1Stats.HitRate < 0.6 || l1Stats.HitRate > 0.7 {
		t.Errorf("Expected hit rate around 0.67, got %.3f", l1Stats.HitRate)
	}
	
	log.Printf("✓ Basic cache monitoring test passed - Hit Rate: %.2f%%", l1Stats.HitRate*100)
}

// TestCacheHitRateMonitor_MultiLayerCache 测试多层缓存监控
func TestCacheHitRateMonitor_MultiLayerCache(t *testing.T) {
	config := DefaultCacheMonitorConfig()
	config.MonitorInterval = 50 * time.Millisecond
	
	monitor := NewCacheHitRateMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 模拟L1和L2缓存的不同命中模式
	scenarios := []struct {
		layer   string
		hitRate float64
		requests int
	}{
		{"L1", 0.8, 200}, // L1缓存高命中率
		{"L2", 0.6, 150}, // L2缓存中等命中率
	}
	
	for _, scenario := range scenarios {
		for i := 0; i < scenario.requests; i++ {
			key := fmt.Sprintf("key_%s_%d", scenario.layer, i%20)
			hit := rand.Float64() < scenario.hitRate
			duration := time.Duration(5+rand.Intn(15)) * time.Millisecond
			
			monitor.RecordCacheAccess(scenario.layer, key, hit, duration)
		}
	}
	
	// 等待处理
	time.Sleep(100 * time.Millisecond)
	
	// 验证各层统计
	l1Stats := monitor.GetLayerStats("L1")
	l2Stats := monitor.GetLayerStats("L2")
	globalStats := monitor.GetGlobalStats()
	
	if l1Stats == nil || l2Stats == nil {
		t.Fatalf("Cache layer stats should not be nil")
	}
	
	// 验证L1缓存统计
	if abs(l1Stats.HitRate-0.8) > 0.1 {
		t.Errorf("L1 hit rate should be around 0.8, got %.3f", l1Stats.HitRate)
	}
	
	// 验证L2缓存统计
	if abs(l2Stats.HitRate-0.6) > 0.1 {
		t.Errorf("L2 hit rate should be around 0.6, got %.3f", l2Stats.HitRate)
	}
	
	// 验证全局统计
	if globalStats.OverallHitRate == 0 {
		t.Errorf("Global hit rate should not be zero")
	}
	
	log.Printf("✓ Multi-layer cache test passed - L1: %.2f%%, L2: %.2f%%, Global: %.2f%%",
		l1Stats.HitRate*100, l2Stats.HitRate*100, globalStats.OverallHitRate*100)
}

// TestCacheHitRateMonitor_HotspotAnalysis 测试热点分析功能
func TestCacheHitRateMonitor_HotspotAnalysis(t *testing.T) {
	config := DefaultCacheMonitorConfig()
	config.MonitorInterval = 50 * time.Millisecond
	config.HotspotThreshold = 0.1 // 10%访问比例认为是热点
	
	monitor := NewCacheHitRateMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 创建明显的热点访问模式
	hotspotKeys := []string{"user:popular_1", "user:popular_2"}
	normalKeys := make([]string, 20)
	for i := range normalKeys {
		normalKeys[i] = fmt.Sprintf("user:normal_%d", i)
	}
	
	// 模拟访问模式：热点键被频繁访问
	totalAccesses := 1000
	for i := 0; i < totalAccesses; i++ {
		var key string
		var hit bool
		
		if i < 600 { // 60%的访问集中在热点键
			key = hotspotKeys[i%len(hotspotKeys)]
			hit = true // 热点键有高命中率
		} else {
			key = normalKeys[(i-600)%len(normalKeys)]
			hit = rand.Float64() < 0.5 // 普通键中等命中率
		}
		
		duration := time.Duration(5+rand.Intn(10)) * time.Millisecond
		monitor.RecordCacheAccess("L1", key, hit, duration)
	}
	
	// 等待热点分析完成
	time.Sleep(200 * time.Millisecond)
	
	// 获取热点数据
	hotspots := monitor.GetHotspots()
	
	if len(hotspots) == 0 {
		t.Errorf("Should detect hotspots, but got none")
	}
	
	// 验证热点检测
	foundHotspot := false
	for _, hotspot := range hotspots {
		if hotspot.Key == "user:popular_1" || hotspot.Key == "user:popular_2" {
			foundHotspot = true
			if hotspot.AccessCount < 250 { // 应该有大量访问
				t.Errorf("Hotspot key should have high access count, got %d", hotspot.AccessCount)
			}
			if hotspot.HitRate < 0.9 { // 应该有高命中率
				t.Errorf("Hotspot key should have high hit rate, got %.3f", hotspot.HitRate)
			}
			break
		}
	}
	
	if !foundHotspot {
		t.Errorf("Should detect popular keys as hotspots")
	}
	
	log.Printf("✓ Hotspot analysis test passed - Detected %d hotspots", len(hotspots))
	if len(hotspots) > 0 {
		top := hotspots[0]
		log.Printf("   Top hotspot: %s (Access: %d, Hit Rate: %.2f%%, Score: %.2f)",
			top.Key, top.AccessCount, top.HitRate*100, top.Score)
	}
}

// TestCacheHitRateMonitor_ResultSetCaching 测试结果集缓存监控
func TestCacheHitRateMonitor_ResultSetCaching(t *testing.T) {
	config := DefaultCacheMonitorConfig()
	monitor := NewCacheHitRateMonitor(config)
	
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 模拟不同大小的结果集缓存
	testCases := []struct {
		query          string
		resultSize     int64
		hit            bool
		serializationTime time.Duration
	}{
		{"SELECT * FROM users WHERE id = ?", 1024, true, 2 * time.Millisecond},
		{"SELECT * FROM orders WHERE status = ?", 5120, false, 8 * time.Millisecond},
		{"SELECT * FROM products LIMIT 100", 51200, true, 25 * time.Millisecond},
		{"SELECT COUNT(*) FROM transactions", 128, true, 1 * time.Millisecond},
		{"SELECT * FROM large_table WHERE date > ?", 1024000, false, 150 * time.Millisecond},
	}
	
	for i, tc := range testCases {
		monitor.RecordResultSetCache(tc.query, tc.resultSize, tc.hit, tc.serializationTime)
		
		// 模拟查询间隔
		time.Sleep(10 * time.Millisecond)
		
		if i == 2 { // 中间检查一次
			if monitor.resultSetMonitor == nil {
				t.Fatalf("Result set monitor should not be nil")
			}
			
			stats := monitor.resultSetMonitor.Stats
			if stats.TotalQueries == 0 {
				t.Errorf("Should have recorded some queries by now")
			}
		}
	}
	
	// 验证结果集缓存统计
	if monitor.resultSetMonitor == nil {
		t.Fatalf("Result set monitor should not be nil")
	}
	
	stats := monitor.resultSetMonitor.Stats
	
	if stats.TotalQueries != int64(len(testCases)) {
		t.Errorf("Expected %d total queries, got %d", len(testCases), stats.TotalQueries)
	}
	
	if stats.MaxResultSize != 1024000 {
		t.Errorf("Expected max result size 1024000, got %d", stats.MaxResultSize)
	}
	
	if stats.AvgSerializationTime == 0 {
		t.Errorf("Average serialization time should not be zero")
	}
	
	// 检查大小分布
	if len(monitor.resultSetMonitor.SizeDistribution) == 0 {
		t.Errorf("Size distribution should not be empty")
	}
	
	log.Printf("✓ Result set caching test passed")
	log.Printf("   Total Queries: %d", stats.TotalQueries)
	log.Printf("   Cache Hits: %d", stats.CacheHits)
	log.Printf("   Avg Result Size: %d bytes", stats.AvgResultSize)
	log.Printf("   Avg Serialization Time: %v", stats.AvgSerializationTime)
}

// TestCacheHitRateMonitor_QueryCaching 测试查询缓存监控
func TestCacheHitRateMonitor_QueryCaching(t *testing.T) {
	config := DefaultCacheMonitorConfig()
	monitor := NewCacheHitRateMonitor(config)
	
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 模拟不同复杂度的查询
	testQueries := []struct {
		query      string
		params     []any
		complexity float64
		hit        bool
	}{
		{"SELECT * FROM users WHERE id = ?", []any{1}, 0.2, true},
		{"SELECT * FROM orders WHERE user_id = ? AND status = ?", []any{1, "active"}, 0.4, true},
		{"SELECT u.*, COUNT(o.id) FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.id", []any{}, 0.8, false},
		{"SELECT id FROM users WHERE email = ?", []any{"test@example.com"}, 0.3, true},
		{"SELECT * FROM complex_view WHERE created_at BETWEEN ? AND ? ORDER BY score DESC", []any{"2024-01-01", "2024-12-31"}, 0.9, false},
	}
	
	for _, tq := range testQueries {
		monitor.RecordQueryCache(tq.query, tq.params, tq.complexity, tq.hit)
		time.Sleep(5 * time.Millisecond)
	}
	
	// 验证查询缓存统计
	if monitor.queryMonitor == nil {
		t.Fatalf("Query monitor should not be nil")
	}
	
	stats := monitor.queryMonitor.Stats
	
	if stats.UniqueQueries != int64(len(testQueries)) {
		t.Errorf("Expected %d unique queries, got %d", len(testQueries), stats.UniqueQueries)
	}
	
	// 验证复杂度分类
	expectedSimple := int64(2) // 复杂度 < 0.5
	expectedComplex := int64(3) // 复杂度 >= 0.5
	
	if stats.SimpleQueries != expectedSimple {
		t.Errorf("Expected %d simple queries, got %d", expectedSimple, stats.SimpleQueries)
	}
	
	if stats.ComplexQueries != expectedComplex {
		t.Errorf("Expected %d complex queries, got %d", expectedComplex, stats.ComplexQueries)
	}
	
	// 验证缓存命中统计
	expectedHits := int64(3) // 有3个查询命中
	if stats.CacheHits != expectedHits {
		t.Errorf("Expected %d cache hits, got %d", expectedHits, stats.CacheHits)
	}
	
	log.Printf("✓ Query caching test passed")
	log.Printf("   Unique Queries: %d", stats.UniqueQueries)
	log.Printf("   Simple Queries: %d, Complex Queries: %d", stats.SimpleQueries, stats.ComplexQueries)
	log.Printf("   Cache Hits: %d, Cache Misses: %d", stats.CacheHits, stats.CacheMisses)
	log.Printf("   Cacheable Hit Rate: %.2f%%", stats.CacheableHitRate*100)
}

// TestCacheHitRateMonitor_RealtimeMetrics 测试实时指标
func TestCacheHitRateMonitor_RealtimeMetrics(t *testing.T) {
	config := DefaultCacheMonitorConfig()
	config.MonitorInterval = 100 * time.Millisecond // 快速更新
	
	monitor := NewCacheHitRateMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 模拟低命中率场景（触发告警）
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("test_key_%d", i%10)
		hit := i < 40 // 只有40%命中率
		duration := time.Duration(10+rand.Intn(10)) * time.Millisecond
		
		monitor.RecordCacheAccess("L1", key, hit, duration)
	}
	
	// 等待实时指标更新
	time.Sleep(200 * time.Millisecond)
	
	// 检查实时指标
	metrics := monitor.GetRealTimeMetrics()
	if metrics == nil {
		t.Fatalf("Realtime metrics should not be nil")
	}
	
	// 验证告警功能
	if metrics.AlertLevel == "NORMAL" {
		t.Errorf("Should trigger alert for low hit rate, but got %s", metrics.AlertLevel)
	}
	
	if metrics.AlertMessage == "" {
		t.Errorf("Should have alert message for low hit rate")
	}
	
	// 验证命中率指标
	if metrics.RecentHitRate > 0.5 {
		t.Errorf("Recent hit rate should be low, got %.3f", metrics.RecentHitRate)
	}
	
	log.Printf("✓ Realtime metrics test passed")
	log.Printf("   Recent Hit Rate: %.2f%%", metrics.RecentHitRate*100)
	log.Printf("   Alert Level: %s", metrics.AlertLevel)
	log.Printf("   Alert Message: %s", metrics.AlertMessage)
}

// TestCacheHitRateMonitor_ConcurrentAccess 测试并发访问
func TestCacheHitRateMonitor_ConcurrentAccess(t *testing.T) {
	config := DefaultCacheMonitorConfig()
	config.MonitorInterval = 50 * time.Millisecond
	
	monitor := NewCacheHitRateMonitor(config)
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 并发访问测试
	const numGoroutines = 10
	const accessesPerGoroutine = 100
	
	var wg sync.WaitGroup
	
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for i := 0; i < accessesPerGoroutine; i++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", goroutineID, i%20)
				hit := (i+goroutineID)%3 == 0 // 变化的命中模式
				duration := time.Duration(5+rand.Intn(10)) * time.Millisecond
				
				// 混合不同类型的记录
				switch i % 4 {
				case 0:
					monitor.RecordCacheAccess("L1", key, hit, duration)
				case 1:
					monitor.RecordCacheAccess("L2", key, hit, duration)
				case 2:
					monitor.RecordResultSetCache(
						fmt.Sprintf("SELECT * FROM table_%d WHERE id = ?", goroutineID),
						int64(1024+rand.Intn(10240)),
						hit,
						duration,
					)
				case 3:
					monitor.RecordQueryCache(
						fmt.Sprintf("SELECT count FROM table_%d", goroutineID),
						[]any{goroutineID},
						rand.Float64(),
						hit,
					)
				}
				
				// 随机短暂停顿
				if i%10 == 0 {
					time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				}
			}
		}(g)
	}
	
	// 等待所有goroutine完成
	wg.Wait()
	
	// 等待数据处理
	time.Sleep(200 * time.Millisecond)
	
	// 验证并发访问结果
	l1Stats := monitor.GetLayerStats("L1")
	l2Stats := monitor.GetLayerStats("L2")
	globalStats := monitor.GetGlobalStats()
	
	if l1Stats == nil || l2Stats == nil {
		t.Fatalf("Cache layer stats should not be nil after concurrent access")
	}
	
	totalExpectedRequests := int64(numGoroutines * accessesPerGoroutine / 2) // L1和L2各一半
	_ = totalExpectedRequests // 标记为已使用
	
	if l1Stats.TotalRequests == 0 || l2Stats.TotalRequests == 0 {
		t.Errorf("Should have recorded requests in both cache layers")
	}
	
	if globalStats.OverallHitRate < 0 || globalStats.OverallHitRate > 1 {
		t.Errorf("Global hit rate should be between 0 and 1, got %.3f", globalStats.OverallHitRate)
	}
	
	// 验证结果集和查询缓存统计
	if monitor.resultSetMonitor != nil {
		rsStats := monitor.resultSetMonitor.Stats
		if rsStats.TotalQueries == 0 {
			t.Errorf("Should have recorded result set cache queries")
		}
	}
	
	if monitor.queryMonitor != nil {
		qStats := monitor.queryMonitor.Stats
		if qStats.UniqueQueries == 0 {
			t.Errorf("Should have recorded query cache queries")
		}
	}
	
	log.Printf("✓ Concurrent access test passed")
	log.Printf("   L1 Requests: %d, L2 Requests: %d", l1Stats.TotalRequests, l2Stats.TotalRequests)
	log.Printf("   Global Hit Rate: %.2f%%", globalStats.OverallHitRate*100)
	
	// 检查热点检测是否正常工作
	hotspots := monitor.GetHotspots()
	log.Printf("   Detected %d hotspots during concurrent access", len(hotspots))
}

// TestCacheHitRateMonitor_PerformanceReport 测试性能报告生成
func TestCacheHitRateMonitor_PerformanceReport(t *testing.T) {
	config := DefaultCacheMonitorConfig()
	monitor := NewCacheHitRateMonitor(config)
	
	err := monitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	// 生成一些测试数据
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("report_key_%d", i%10)
		hit := i%3 != 0 // ~67% 命中率
		duration := time.Duration(8+rand.Intn(12)) * time.Millisecond
		
		monitor.RecordCacheAccess("L1", key, hit, duration)
		monitor.RecordCacheAccess("L2", key, i%4 != 0, duration) // ~75% 命中率
		
		// 添加结果集和查询缓存数据
		if i%5 == 0 {
			monitor.RecordResultSetCache(
				fmt.Sprintf("SELECT * FROM table WHERE id = %d", i),
				int64(2048+rand.Intn(8192)),
				hit,
				duration,
			)
			
			monitor.RecordQueryCache(
				fmt.Sprintf("SELECT count FROM table WHERE type = %d", i%3),
				[]any{i % 3},
				rand.Float64(),
				hit,
			)
		}
	}
	
	// 等待数据处理
	time.Sleep(100 * time.Millisecond)
	
	// 生成性能报告
	report := monitor.GenerateCacheReport()
	
	if report == nil {
		t.Fatalf("Performance report should not be nil")
	}
	
	// 验证报告内容
	if report.GlobalStats == nil {
		t.Errorf("Global stats should not be nil in report")
	}
	
	if report.RealTimeMetrics == nil {
		t.Errorf("Realtime metrics should not be nil in report")
	}
	
	if report.L1Stats == nil {
		t.Errorf("L1 stats should not be nil in report")
	}
	
	if report.L2Stats == nil {
		t.Errorf("L2 stats should not be nil in report")
	}
	
	if len(report.Hotspots) == 0 {
		log.Printf("   Note: No hotspots detected (may be normal for small dataset)")
	}
	
	if len(report.Recommendations) == 0 {
		log.Printf("   Note: No recommendations generated (may be normal for good performance)")
	}
	
	log.Printf("✓ Performance report test passed")
	log.Printf("   Report Timestamp: %s", report.Timestamp.Format("2006-01-02 15:04:05"))
	log.Printf("   Global Hit Rate: %.2f%%", report.GlobalStats.OverallHitRate*100)
	log.Printf("   L1 Hit Rate: %.2f%%, L2 Hit Rate: %.2f%%", 
		report.L1Stats.HitRate*100, report.L2Stats.HitRate*100)
	log.Printf("   Hotspots Count: %d", len(report.Hotspots))
	log.Printf("   Recommendations Count: %d", len(report.Recommendations))
	
	if len(report.Recommendations) > 0 {
		log.Printf("   First Recommendation: %s", report.Recommendations[0])
	}
}

// BenchmarkCacheHitRateMonitor_Performance 缓存监控器性能基准测试
func BenchmarkCacheHitRateMonitor_Performance(b *testing.B) {
	config := DefaultCacheMonitorConfig()
	config.MonitorInterval = 1 * time.Second // 降低监控频率以减少干扰
	
	monitor := NewCacheHitRateMonitor(config)
	err := monitor.Start()
	if err != nil {
		b.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer monitor.Stop()
	
	keys := make([]string, 100)
	for i := range keys {
		keys[i] = fmt.Sprintf("bench_key_%d", i)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := keys[rand.Intn(len(keys))]
			hit := rand.Float64() < 0.7 // 70% 命中率
			duration := time.Duration(5+rand.Intn(15)) * time.Millisecond
			
			monitor.RecordCacheAccess("L1", key, hit, duration)
		}
	})
}

// 辅助函数

// abs 计算浮点数绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}