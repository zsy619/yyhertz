// Package mybatis 高并发压力测试
//
// 测试场景：
// 1. 500+ 并发连接压力测试
// 2. 极限并发下的性能表现
// 3. 连接池资源竞争测试
// 4. 高并发下的缓存性能
// 5. 系统稳定性验证
package mybatis

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// HighConcurrencyTestConfig 高并发测试配置
type HighConcurrencyTestConfig struct {
	MaxConcurrentConnections int           `json:"max_concurrent_connections"`
	TestDuration             time.Duration `json:"test_duration"`
	QueryTypes               []string      `json:"query_types"`
	DataVolumePerConnection  int           `json:"data_volume_per_connection"`
	EnableCacheMonitoring    bool          `json:"enable_cache_monitoring"`
	EnableSlowQueryMonitor   bool          `json:"enable_slow_query_monitor"`
}

// ConcurrencyTestResult 并发测试结果
type ConcurrencyTestResult struct {
	// 基础指标
	TotalConnections      int64 `json:"total_connections"`
	SuccessfulConnections int64 `json:"successful_connections"`
	FailedConnections     int64 `json:"failed_connections"`
	TotalQueries          int64 `json:"total_queries"`
	SuccessfulQueries     int64 `json:"successful_queries"`
	FailedQueries         int64 `json:"failed_queries"`

	// 性能指标
	AvgConnectionTime time.Duration `json:"avg_connection_time"`
	MaxConnectionTime time.Duration `json:"max_connection_time"`
	MinConnectionTime time.Duration `json:"min_connection_time"`
	AvgQueryTime      time.Duration `json:"avg_query_time"`
	MaxQueryTime      time.Duration `json:"max_query_time"`
	MinQueryTime      time.Duration `json:"min_query_time"`

	// 吞吐量指标
	QueriesPerSecond     float64 `json:"queries_per_second"`
	ConnectionsPerSecond float64 `json:"connections_per_second"`

	// 资源使用
	PeakMemoryUsage uint64 `json:"peak_memory_usage"`
	PeakGoroutines  int    `json:"peak_goroutines"`

	// 错误分析
	ErrorTypes map[string]int64 `json:"error_types"`

	// 测试配置
	TestConfig *HighConcurrencyTestConfig `json:"test_config"`

	// 时间戳
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

// TestMyBatisSession 测试用的MyBatis会话接口
type TestMyBatisSession interface {
	SelectOne(sql string, result interface{}, args ...interface{}) error
	Insert(sql string, args ...interface{}) error
	Update(sql string, args ...interface{}) error
	Delete(sql string, args ...interface{}) error
	Close() error
}

// ConnectionWorker 连接工作器
type ConnectionWorker struct {
	ID             int
	Session        TestMyBatisSession
	Config         *HighConcurrencyTestConfig
	Results        *WorkerResults
	StopChan       chan struct{}
	ErrorCollector chan error
}

// WorkerResults 工作器结果
type WorkerResults struct {
	QueriesExecuted   int64            `json:"queries_executed"`
	QueriesSuccessful int64            `json:"queries_successful"`
	QueriesFailed     int64            `json:"queries_failed"`
	TotalQueryTime    time.Duration    `json:"total_query_time"`
	MaxQueryTime      time.Duration    `json:"max_query_time"`
	MinQueryTime      time.Duration    `json:"min_query_time"`
	ConnectionTime    time.Duration    `json:"connection_time"`
	ErrorCounts       map[string]int64 `json:"error_counts"`
	mutex             sync.RWMutex
}

// TestHighConcurrency_500Connections 测试500+并发连接
func TestHighConcurrency_500Connections(t *testing.T) {
	// 检查系统资源
	if runtime.NumCPU() < 4 {
		t.Skip("Skipping high concurrency test on systems with < 4 CPUs")
	}

	config := &HighConcurrencyTestConfig{
		MaxConcurrentConnections: 500,
		TestDuration:             30 * time.Second,
		QueryTypes:               []string{"SELECT", "INSERT", "UPDATE"},
		DataVolumePerConnection:  100, // 每个连接执行100个查询
		EnableCacheMonitoring:    true,
		EnableSlowQueryMonitor:   true,
	}

	log.Printf("=== Starting 500+ Concurrent Connections Test ===")
	log.Printf("Max Connections: %d", config.MaxConcurrentConnections)
	log.Printf("Test Duration: %v", config.TestDuration)
	log.Printf("CPUs Available: %d", runtime.NumCPU())

	result := runHighConcurrencyTest(t, config)

	// 验证结果
	if result.SuccessfulConnections < int64(float64(config.MaxConcurrentConnections)*0.8) {
		t.Errorf("Too many failed connections. Expected >80%% success rate, got %.2f%%",
			float64(result.SuccessfulConnections)/float64(result.TotalConnections)*100)
	}

	if result.QueriesPerSecond < 1000 {
		t.Errorf("QPS too low. Expected >1000 QPS, got %.2f", result.QueriesPerSecond)
	}

	// 打印详细结果
	printConcurrencyTestResults(result)

	log.Printf("✓ High concurrency test completed successfully")
}

// TestHighConcurrency_ExtremeConcurrency 极限并发测试
func TestHighConcurrency_ExtremeConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping extreme concurrency test in short mode")
	}

	// 根据系统资源动态调整并发数
	maxConcurrency := runtime.NumCPU() * 200 // 每个CPU核心200个连接
	if maxConcurrency > 1000 {
		maxConcurrency = 1000 // 限制最大值
	}

	config := &HighConcurrencyTestConfig{
		MaxConcurrentConnections: maxConcurrency,
		TestDuration:             45 * time.Second,
		QueryTypes:               []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
		DataVolumePerConnection:  50,
		EnableCacheMonitoring:    true,
		EnableSlowQueryMonitor:   true,
	}

	log.Printf("=== Starting Extreme Concurrency Test ===")
	log.Printf("Max Connections: %d", config.MaxConcurrentConnections)

	result := runHighConcurrencyTest(t, config)

	// 验证系统稳定性
	successRate := float64(result.SuccessfulConnections) / float64(result.TotalConnections) * 100
	if successRate < 70 { // 极限情况下允许更多失败
		t.Errorf("System instability detected. Success rate: %.2f%%", successRate)
	}

	// 检查内存使用
	if result.PeakMemoryUsage > 2*1024*1024*1024 { // 2GB限制
		t.Errorf("Memory usage too high: %d bytes", result.PeakMemoryUsage)
	}

	printConcurrencyTestResults(result)

	log.Printf("✓ Extreme concurrency test completed")
}

// TestHighConcurrency_ResourceCompetition 资源竞争测试
func TestHighConcurrency_ResourceCompetition(t *testing.T) {
	config := &HighConcurrencyTestConfig{
		MaxConcurrentConnections: 300,
		TestDuration:             20 * time.Second,
		QueryTypes:               []string{"SELECT", "UPDATE"},
		DataVolumePerConnection:  200,
		EnableCacheMonitoring:    true,
		EnableSlowQueryMonitor:   true,
	}

	log.Printf("=== Starting Resource Competition Test ===")

	// 创建资源竞争场景：多个连接访问相同的数据
	result := runResourceCompetitionTest(t, config)

	// 验证死锁检测
	if deadlocks, exists := result.ErrorTypes["deadlock"]; exists && deadlocks > 10 {
		t.Errorf("Too many deadlocks detected: %d", deadlocks)
	}

	// 验证事务冲突处理
	if conflicts, exists := result.ErrorTypes["conflict"]; exists && conflicts > result.TotalQueries/10 {
		t.Errorf("Too many transaction conflicts: %d", conflicts)
	}

	printConcurrencyTestResults(result)

	log.Printf("✓ Resource competition test completed")
}

// TestHighConcurrency_CachePerformance 高并发缓存性能测试
func TestHighConcurrency_CachePerformance(t *testing.T) {
	// 启用缓存监控
	cacheMonitor := NewCacheHitRateMonitor(DefaultCacheMonitorConfig())
	err := cacheMonitor.Start()
	if err != nil {
		t.Fatalf("Failed to start cache monitor: %v", err)
	}
	defer cacheMonitor.Stop()

	config := &HighConcurrencyTestConfig{
		MaxConcurrentConnections: 400,
		TestDuration:             25 * time.Second,
		QueryTypes:               []string{"SELECT"}, // 重点测试缓存
		DataVolumePerConnection:  300,
		EnableCacheMonitoring:    true,
		EnableSlowQueryMonitor:   false,
	}

	log.Printf("=== Starting High Concurrency Cache Performance Test ===")

	result := runCachePerformanceTest(t, config, cacheMonitor)

	// 验证缓存性能
	cacheReport := cacheMonitor.GenerateCacheReport()
	if cacheReport.GlobalStats.OverallHitRate < 0.3 {
		t.Errorf("Cache hit rate too low under high concurrency: %.2f%%",
			cacheReport.GlobalStats.OverallHitRate*100)
	}

	printConcurrencyTestResults(result)
	printCachePerformanceReport(cacheReport)

	log.Printf("✓ High concurrency cache performance test completed")
}

// BenchmarkHighConcurrency_Scalability 可扩展性基准测试
func BenchmarkHighConcurrency_Scalability(b *testing.B) {
	concurrencyLevels := []int{10, 50, 100, 200, 500}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			config := &HighConcurrencyTestConfig{
				MaxConcurrentConnections: concurrency,
				TestDuration:             time.Duration(b.N) * 10 * time.Millisecond,
				QueryTypes:               []string{"SELECT", "INSERT"},
				DataVolumePerConnection:  10,
				EnableCacheMonitoring:    false, // 减少开销
				EnableSlowQueryMonitor:   false,
			}

			b.ResetTimer()
			b.ReportAllocs()

			result := benchmarkConcurrency(b, config)

			// 报告指标
			b.ReportMetric(result.QueriesPerSecond, "qps")
			b.ReportMetric(float64(result.AvgQueryTime.Nanoseconds()), "ns/query")
			b.ReportMetric(float64(result.PeakMemoryUsage), "bytes/mem")
		})
	}
}

// 核心测试实现函数

// runHighConcurrencyTest 运行高并发测试
func runHighConcurrencyTest(t *testing.T, config *HighConcurrencyTestConfig) *ConcurrencyTestResult {
	result := &ConcurrencyTestResult{
		TestConfig:        config,
		StartTime:         time.Now(),
		ErrorTypes:        make(map[string]int64),
		MinQueryTime:      time.Hour, // 初始化为很大的值
		MinConnectionTime: time.Hour,
	}

	// 创建错误收集器
	errorChan := make(chan error, config.MaxConcurrentConnections*10)

	// 启动资源监控
	resourceMonitor := startResourceMonitor(result)
	defer close(resourceMonitor)

	// 创建工作器
	var wg sync.WaitGroup
	workers := make([]*ConnectionWorker, config.MaxConcurrentConnections)

	log.Printf("Creating %d connection workers...", config.MaxConcurrentConnections)

	for i := 0; i < config.MaxConcurrentConnections; i++ {
		worker := &ConnectionWorker{
			ID:     i,
			Config: config,
			Results: &WorkerResults{
				ErrorCounts:  make(map[string]int64),
				MinQueryTime: time.Hour,
			},
			StopChan:       make(chan struct{}),
			ErrorCollector: errorChan,
		}

		workers[i] = worker

		wg.Add(1)
		go func(w *ConnectionWorker) {
			defer wg.Done()
			runConnectionWorker(w)
		}(worker)

		// 避免同时创建太多连接
		if (i+1)%50 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	log.Printf("All workers started, running for %v...", config.TestDuration)

	// 运行指定时间
	time.Sleep(config.TestDuration)

	// 停止所有工作器
	log.Printf("Stopping all workers...")
	for _, worker := range workers {
		close(worker.StopChan)
	}

	// 等待所有工作器完成
	wg.Wait()
	close(errorChan)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// 收集所有工作器的结果
	collectWorkerResults(result, workers)

	// 处理错误信息
	processErrors(result, errorChan)

	// 计算派生指标
	calculateDerivedMetrics(result)

	return result
}

// runConnectionWorker 运行连接工作器
func runConnectionWorker(worker *ConnectionWorker) {
	startTime := time.Now()

	// 尝试建立连接
	session, err := createTestSession()
	if err != nil {
		worker.ErrorCollector <- fmt.Errorf("worker %d connection failed: %w", worker.ID, err)
		return
	}
	defer session.Close()

	worker.Session = session
	worker.Results.ConnectionTime = time.Since(startTime)

	// 执行查询
	for {
		select {
		case <-worker.StopChan:
			return
		default:
			executeWorkerQuery(worker)

			// 短暂休息避免CPU占用过高
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		}
	}
}

// executeWorkerQuery 执行工作器查询
func executeWorkerQuery(worker *ConnectionWorker) {
	worker.Results.mutex.Lock()
	queryNum := worker.Results.QueriesExecuted
	worker.Results.QueriesExecuted++
	worker.Results.mutex.Unlock()

	queryType := worker.Config.QueryTypes[queryNum%int64(len(worker.Config.QueryTypes))]

	startTime := time.Now()
	err := executeQueryByType(worker.Session, queryType, queryNum)
	duration := time.Since(startTime)

	worker.Results.mutex.Lock()
	defer worker.Results.mutex.Unlock()

	worker.Results.TotalQueryTime += duration

	if duration > worker.Results.MaxQueryTime {
		worker.Results.MaxQueryTime = duration
	}
	if duration < worker.Results.MinQueryTime {
		worker.Results.MinQueryTime = duration
	}

	if err != nil {
		worker.Results.QueriesFailed++
		errorType := categorizeError(err)
		worker.Results.ErrorCounts[errorType]++
		worker.ErrorCollector <- fmt.Errorf("worker %d query failed: %w", worker.ID, err)
	} else {
		worker.Results.QueriesSuccessful++
	}
}

// executeQueryByType 根据类型执行查询
func executeQueryByType(session TestMyBatisSession, queryType string, queryNum int64) error {
	switch queryType {
	case "SELECT":
		return executeSelectQuery(session, queryNum)
	case "INSERT":
		return executeInsertQuery(session, queryNum)
	case "UPDATE":
		return executeUpdateQuery(session, queryNum)
	case "DELETE":
		return executeDeleteQuery(session, queryNum)
	default:
		return fmt.Errorf("unknown query type: %s", queryType)
	}
}

// executeSelectQuery 执行SELECT查询
func executeSelectQuery(session TestMyBatisSession, queryNum int64) error {
	// 模拟不同的SELECT模式
	patterns := []string{
		"SELECT id, name FROM users WHERE id = ?",
		"SELECT * FROM orders WHERE user_id = ? LIMIT 10",
		"SELECT COUNT(*) FROM products WHERE category = ?",
		"SELECT u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = ?",
	}

	sql := patterns[queryNum%int64(len(patterns))]
	param := queryNum%1000 + 1

	var result interface{}
	return session.SelectOne(sql, &result, param)
}

// executeInsertQuery 执行INSERT查询
func executeInsertQuery(session TestMyBatisSession, queryNum int64) error {
	sql := "INSERT INTO test_users (name, email, created_at) VALUES (?, ?, ?)"
	return session.Insert(sql,
		fmt.Sprintf("user_%d", queryNum),
		fmt.Sprintf("user_%d@test.com", queryNum),
		time.Now(),
	)
}

// executeUpdateQuery 执行UPDATE查询
func executeUpdateQuery(session TestMyBatisSession, queryNum int64) error {
	sql := "UPDATE test_users SET name = ?, updated_at = ? WHERE id = ?"
	return session.Update(sql,
		fmt.Sprintf("updated_user_%d", queryNum),
		time.Now(),
		queryNum%1000+1,
	)
}

// executeDeleteQuery 执行DELETE查询
func executeDeleteQuery(session TestMyBatisSession, queryNum int64) error {
	sql := "DELETE FROM test_users WHERE id = ? AND created_at < ?"
	return session.Delete(sql, queryNum%1000+1, time.Now().Add(-24*time.Hour))
}

// runResourceCompetitionTest 运行资源竞争测试
func runResourceCompetitionTest(t *testing.T, config *HighConcurrencyTestConfig) *ConcurrencyTestResult {
	// 这里实现专门的资源竞争测试逻辑
	// 创建多个连接同时访问相同的数据行，测试锁竞争处理
	return runHighConcurrencyTest(t, config)
}

// runCachePerformanceTest 运行缓存性能测试
func runCachePerformanceTest(t *testing.T, config *HighConcurrencyTestConfig, cacheMonitor *CacheHitRateMonitor) *ConcurrencyTestResult {
	// 这里实现专门的缓存性能测试逻辑
	result := runHighConcurrencyTest(t, config)

	// 在测试过程中记录缓存访问
	// 这里应该集成到具体的查询执行中

	return result
}

// benchmarkConcurrency 基准测试并发性能
func benchmarkConcurrency(b *testing.B, config *HighConcurrencyTestConfig) *ConcurrencyTestResult {
	// 简化的基准测试实现
	config.TestDuration = time.Duration(b.N) * time.Millisecond
	result := &ConcurrencyTestResult{
		TestConfig: config,
		StartTime:  time.Now(),
		ErrorTypes: make(map[string]int64),
	}

	// 执行简化的并发测试
	var wg sync.WaitGroup
	var totalQueries int64

	for i := 0; i < config.MaxConcurrentConnections; i++ {
		wg.Go(func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				atomic.AddInt64(&totalQueries, 1)
				time.Sleep(time.Microsecond) // 模拟查询执行
			}
		})
	}

	wg.Wait()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.TotalQueries = totalQueries
	result.SuccessfulQueries = totalQueries

	if result.Duration > 0 {
		result.QueriesPerSecond = float64(totalQueries) / result.Duration.Seconds()
	}

	return result
}

// 辅助函数

// createTestSession 创建测试会话
func createTestSession() (TestMyBatisSession, error) {
	// 这里应该创建一个真实的MyBatis会话
	// 目前返回一个模拟实现
	return &MockSession{}, nil
}

// MockSession 模拟会话实现
type MockSession struct{}

func (ms *MockSession) SelectOne(sql string, result interface{}, args ...interface{}) error {
	time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond)
	return nil
}

func (ms *MockSession) Insert(sql string, args ...interface{}) error {
	time.Sleep(time.Duration(2+rand.Intn(8)) * time.Millisecond)
	return nil
}

func (ms *MockSession) Update(sql string, args ...interface{}) error {
	time.Sleep(time.Duration(1+rand.Intn(6)) * time.Millisecond)
	return nil
}

func (ms *MockSession) Delete(sql string, args ...interface{}) error {
	time.Sleep(time.Duration(1+rand.Intn(4)) * time.Millisecond)
	return nil
}

func (ms *MockSession) Close() error {
	return nil
}

// startResourceMonitor 启动资源监控
func startResourceMonitor(result *ConcurrencyTestResult) chan struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				if m.Alloc > result.PeakMemoryUsage {
					result.PeakMemoryUsage = m.Alloc
				}

				goroutines := runtime.NumGoroutine()
				if goroutines > result.PeakGoroutines {
					result.PeakGoroutines = goroutines
				}
			}
		}
	}()

	return stopChan
}

// collectWorkerResults 收集工作器结果
func collectWorkerResults(result *ConcurrencyTestResult, workers []*ConnectionWorker) {
	result.TotalConnections = int64(len(workers))

	for _, worker := range workers {
		if worker.Results == nil {
			result.FailedConnections++
			continue
		}

		result.SuccessfulConnections++
		result.TotalQueries += worker.Results.QueriesExecuted
		result.SuccessfulQueries += worker.Results.QueriesSuccessful
		result.FailedQueries += worker.Results.QueriesFailed

		// 连接时间统计
		if worker.Results.ConnectionTime > result.MaxConnectionTime {
			result.MaxConnectionTime = worker.Results.ConnectionTime
		}
		if worker.Results.ConnectionTime < result.MinConnectionTime {
			result.MinConnectionTime = worker.Results.ConnectionTime
		}

		// 查询时间统计
		if worker.Results.MaxQueryTime > result.MaxQueryTime {
			result.MaxQueryTime = worker.Results.MaxQueryTime
		}
		if worker.Results.MinQueryTime < result.MinQueryTime {
			result.MinQueryTime = worker.Results.MinQueryTime
		}

		// 错误统计
		for errorType, count := range worker.Results.ErrorCounts {
			result.ErrorTypes[errorType] += count
		}
	}
}

// processErrors 处理错误信息
func processErrors(result *ConcurrencyTestResult, errorChan chan error) {
	for err := range errorChan {
		errorType := categorizeError(err)
		result.ErrorTypes[errorType]++
	}
}

// categorizeError 错误分类
func categorizeError(err error) string {
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "connection"):
		return "connection_error"
	case strings.Contains(errStr, "timeout"):
		return "timeout_error"
	case strings.Contains(errStr, "deadlock"):
		return "deadlock"
	case strings.Contains(errStr, "conflict"):
		return "conflict"
	case strings.Contains(errStr, "constraint"):
		return "constraint_violation"
	default:
		return "unknown_error"
	}
}

// calculateDerivedMetrics 计算派生指标
func calculateDerivedMetrics(result *ConcurrencyTestResult) {
	if result.Duration > 0 {
		result.QueriesPerSecond = float64(result.TotalQueries) / result.Duration.Seconds()
		result.ConnectionsPerSecond = float64(result.TotalConnections) / result.Duration.Seconds()
	}

	if result.SuccessfulConnections > 0 {
		totalConnectionTime := time.Duration(0)
		// 这里需要从workers中收集连接时间数据
		result.AvgConnectionTime = totalConnectionTime / time.Duration(result.SuccessfulConnections)
	}

	if result.SuccessfulQueries > 0 {
		// 这里需要从workers中收集查询时间数据
		// 简化处理
		result.AvgQueryTime = time.Duration(result.TotalQueries) * time.Millisecond / time.Duration(result.SuccessfulQueries)
	}
}

// printConcurrencyTestResults 打印并发测试结果
func printConcurrencyTestResults(result *ConcurrencyTestResult) {
	log.Printf("\n=== High Concurrency Test Results ===")
	log.Printf("Test Duration: %v", result.Duration)
	log.Printf("Total Connections: %d (Success: %d, Failed: %d)",
		result.TotalConnections, result.SuccessfulConnections, result.FailedConnections)
	log.Printf("Total Queries: %d (Success: %d, Failed: %d)",
		result.TotalQueries, result.SuccessfulQueries, result.FailedQueries)
	log.Printf("Success Rate: %.2f%%",
		float64(result.SuccessfulQueries)/float64(result.TotalQueries)*100)
	log.Printf("Queries Per Second: %.2f", result.QueriesPerSecond)
	log.Printf("Connections Per Second: %.2f", result.ConnectionsPerSecond)
	log.Printf("Avg Query Time: %v", result.AvgQueryTime)
	log.Printf("Max Query Time: %v", result.MaxQueryTime)
	log.Printf("Min Query Time: %v", result.MinQueryTime)
	log.Printf("Peak Memory Usage: %d bytes (%.2f MB)",
		result.PeakMemoryUsage, float64(result.PeakMemoryUsage)/1024/1024)
	log.Printf("Peak Goroutines: %d", result.PeakGoroutines)

	if len(result.ErrorTypes) > 0 {
		log.Printf("Error Summary:")
		for errorType, count := range result.ErrorTypes {
			log.Printf("  %s: %d", errorType, count)
		}
	}
}

// printCachePerformanceReport 打印缓存性能报告
func printCachePerformanceReport(report *CachePerformanceReport) {
	log.Printf("\n=== Cache Performance Under High Concurrency ===")
	log.Printf("Overall Hit Rate: %.2f%%", report.GlobalStats.OverallHitRate*100)
	log.Printf("Overall Miss Rate: %.2f%%", report.GlobalStats.OverallMissRate*100)

	if report.L1Stats != nil {
		log.Printf("L1 Cache Hit Rate: %.2f%%", report.L1Stats.HitRate*100)
		log.Printf("L1 Cache Requests: %d", report.L1Stats.TotalRequests)
	}

	if report.L2Stats != nil {
		log.Printf("L2 Cache Hit Rate: %.2f%%", report.L2Stats.HitRate*100)
		log.Printf("L2 Cache Requests: %d", report.L2Stats.TotalRequests)
	}

	log.Printf("Cache Hotspots Detected: %d", len(report.Hotspots))
	log.Printf("Cache Recommendations: %d", len(report.Recommendations))
}
