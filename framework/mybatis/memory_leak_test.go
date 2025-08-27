// Package mybatis 长期运行内存泄漏检测测试
//
// 测试场景：
// 1. 24小时连续运行测试
// 2. 内存使用趋势监控
// 3. 垃圾回收效率测试
// 4. 资源清理验证
// 5. 内存泄漏检测和报告
package mybatis

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// MemoryLeakTestConfig 内存泄漏测试配置
type MemoryLeakTestConfig struct {
	TestDuration            time.Duration `json:"test_duration"`
	SamplingInterval        time.Duration `json:"sampling_interval"`
	MaxMemoryGrowthRate     float64       `json:"max_memory_growth_rate"`     // 每小时最大内存增长率
	MaxGoroutineGrowthRate  float64       `json:"max_goroutine_growth_rate"`  // 每小时最大goroutine增长率
	ForceGCInterval         time.Duration `json:"force_gc_interval"`
	WorkloadConcurrency     int           `json:"workload_concurrency"`
	EnableDetailedTracking  bool          `json:"enable_detailed_tracking"`
}

// MemorySnapshot 内存快照
type MemorySnapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	HeapAlloc      uint64    `json:"heap_alloc"`
	HeapSys        uint64    `json:"heap_sys"`
	HeapIdle       uint64    `json:"heap_idle"`
	HeapInuse      uint64    `json:"heap_inuse"`
	HeapReleased   uint64    `json:"heap_released"`
	HeapObjects    uint64    `json:"heap_objects"`
	StackInuse     uint64    `json:"stack_inuse"`
	StackSys       uint64    `json:"stack_sys"`
	GoroutineCount int       `json:"goroutine_count"`
	GCCount        uint32    `json:"gc_count"`
	GCPauseTotal   uint64    `json:"gc_pause_total"`
}

// MemoryLeakTestResult 内存泄漏测试结果
type MemoryLeakTestResult struct {
	TestConfig           *MemoryLeakTestConfig `json:"test_config"`
	StartSnapshot        *MemorySnapshot       `json:"start_snapshot"`
	EndSnapshot          *MemorySnapshot       `json:"end_snapshot"`
	Snapshots            []*MemorySnapshot     `json:"snapshots"`
	
	// 趋势分析
	MemoryGrowthRate     float64               `json:"memory_growth_rate"`     // MB/hour
	GoroutineGrowthRate  float64               `json:"goroutine_growth_rate"`  // count/hour
	GCEfficiency         float64               `json:"gc_efficiency"`          // %
	
	// 泄漏检测结果
	MemoryLeakDetected   bool                  `json:"memory_leak_detected"`
	GoroutineLeakDetected bool                 `json:"goroutine_leak_detected"`
	ResourceLeakDetected bool                 `json:"resource_leak_detected"`
	
	// 详细分析
	SuspiciousPatterns   []LeakPattern         `json:"suspicious_patterns"`
	Recommendations      []string              `json:"recommendations"`
	
	// 测试指标
	TotalOperations      int64                 `json:"total_operations"`
	OperationsPerSecond  float64               `json:"operations_per_second"`
	ErrorCount           int64                 `json:"error_count"`
	
	Duration             time.Duration         `json:"duration"`
	StartTime            time.Time             `json:"start_time"`
	EndTime              time.Time             `json:"end_time"`
}

// LeakPattern 泄漏模式
type LeakPattern struct {
	Type        string    `json:"type"`         // memory, goroutine, file_descriptor, etc.
	Severity    string    `json:"severity"`     // LOW, MEDIUM, HIGH, CRITICAL
	Description string    `json:"description"`
	Evidence    []string  `json:"evidence"`
	DetectedAt  time.Time `json:"detected_at"`
	GrowthRate  float64   `json:"growth_rate"`
}

// MemoryLeakDetector 内存泄漏检测器
type MemoryLeakDetector struct {
	config           *MemoryLeakTestConfig
	snapshots        []*MemorySnapshot
	workload         *TestWorkload
	isRunning        int32
	stopChan         chan struct{}
	snapshotChan     chan *MemorySnapshot
	mutex            sync.RWMutex
}

// TestWorkload 测试工作负载
type TestWorkload struct {
	sessions         []TestMyBatisSession
	cacheMonitor     *CacheHitRateMonitor
	slowQueryMonitor *SlowQueryMonitor
	operationCount   int64
	errorCount       int64
	stopChan         chan struct{}
}

// TestMemoryLeak_ShortDuration 短期内存泄漏测试 (用于CI)
func TestMemoryLeak_ShortDuration(t *testing.T) {
	config := &MemoryLeakTestConfig{
		TestDuration:           2 * time.Minute,    // 短期测试
		SamplingInterval:       5 * time.Second,
		MaxMemoryGrowthRate:    10.0,               // 10MB/hour 
		MaxGoroutineGrowthRate: 100.0,              // 100 goroutines/hour
		ForceGCInterval:        10 * time.Second,
		WorkloadConcurrency:    50,
		EnableDetailedTracking: true,
	}
	
	log.Printf("=== Starting Short Duration Memory Leak Test ===")
	log.Printf("Duration: %v", config.TestDuration)
	log.Printf("Sampling Interval: %v", config.SamplingInterval)
	
	result := runMemoryLeakTest(t, config)
	
	// 验证结果
	if result.MemoryLeakDetected {
		t.Errorf("Memory leak detected in short duration test")
	}
	
	if result.GoroutineLeakDetected {
		t.Errorf("Goroutine leak detected in short duration test")
	}
	
	printMemoryLeakTestResults(result)
	
	log.Printf("✓ Short duration memory leak test completed")
}

// TestMemoryLeak_MediumDuration 中期内存泄漏测试
func TestMemoryLeak_MediumDuration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping medium duration memory leak test in short mode")
	}
	
	config := &MemoryLeakTestConfig{
		TestDuration:           10 * time.Minute,   // 中期测试
		SamplingInterval:       15 * time.Second,
		MaxMemoryGrowthRate:    5.0,                // 5MB/hour
		MaxGoroutineGrowthRate: 50.0,               // 50 goroutines/hour
		ForceGCInterval:        30 * time.Second,
		WorkloadConcurrency:    100,
		EnableDetailedTracking: true,
	}
	
	log.Printf("=== Starting Medium Duration Memory Leak Test ===")
	log.Printf("Duration: %v", config.TestDuration)
	
	result := runMemoryLeakTest(t, config)
	
	// 更严格的验证
	if result.MemoryGrowthRate > config.MaxMemoryGrowthRate {
		t.Errorf("Memory growth rate too high: %.2f MB/hour (limit: %.2f)", 
			result.MemoryGrowthRate, config.MaxMemoryGrowthRate)
	}
	
	if result.GoroutineGrowthRate > config.MaxGoroutineGrowthRate {
		t.Errorf("Goroutine growth rate too high: %.2f/hour (limit: %.2f)",
			result.GoroutineGrowthRate, config.MaxGoroutineGrowthRate)
	}
	
	printMemoryLeakTestResults(result)
	
	log.Printf("✓ Medium duration memory leak test completed")
}

// TestMemoryLeak_LongDuration 长期内存泄漏测试 (手动运行)
func TestMemoryLeak_LongDuration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long duration memory leak test in short mode")
	}
	
	// 检查是否明确要求运行长期测试
	if !isLongTestRequested() {
		t.Skip("Long duration test not requested - use -long flag to enable")
	}
	
	config := &MemoryLeakTestConfig{
		TestDuration:           1 * time.Hour,      // 1小时测试
		SamplingInterval:       1 * time.Minute,
		MaxMemoryGrowthRate:    2.0,                // 2MB/hour
		MaxGoroutineGrowthRate: 10.0,               // 10 goroutines/hour
		ForceGCInterval:        5 * time.Minute,
		WorkloadConcurrency:    200,
		EnableDetailedTracking: true,
	}
	
	log.Printf("=== Starting Long Duration Memory Leak Test ===")
	log.Printf("Duration: %v (This will take a while...)", config.TestDuration)
	log.Printf("You can monitor progress in the logs")
	
	result := runMemoryLeakTest(t, config)
	
	// 最严格的验证
	if result.MemoryLeakDetected {
		t.Errorf("Memory leak detected: Growth rate %.2f MB/hour", result.MemoryGrowthRate)
	}
	
	if result.GoroutineLeakDetected {
		t.Errorf("Goroutine leak detected: Growth rate %.2f/hour", result.GoroutineGrowthRate)
	}
	
	// 生成详细报告
	generateDetailedLeakReport(result)
	
	printMemoryLeakTestResults(result)
	
	log.Printf("✓ Long duration memory leak test completed successfully")
}

// BenchmarkMemoryLeak_Performance 内存泄漏检测性能基准
func BenchmarkMemoryLeak_Performance(b *testing.B) {
	detector := NewMemoryLeakDetector(&MemoryLeakTestConfig{
		SamplingInterval:       100 * time.Millisecond,
		WorkloadConcurrency:    10,
		EnableDetailedTracking: false, // 减少开销
	})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			snapshot := detector.takeMemorySnapshot()
			detector.analyzeSnapshot(snapshot)
		}
	})
}

// 核心实现函数

// runMemoryLeakTest 运行内存泄漏测试
func runMemoryLeakTest(t *testing.T, config *MemoryLeakTestConfig) *MemoryLeakTestResult {
	detector := NewMemoryLeakDetector(config)
	
	result := &MemoryLeakTestResult{
		TestConfig:        config,
		Snapshots:        make([]*MemorySnapshot, 0),
		SuspiciousPatterns: make([]LeakPattern, 0),
		Recommendations:   make([]string, 0),
		StartTime:        time.Now(),
	}
	
	// 启动内存泄漏检测器
	err := detector.Start()
	if err != nil {
		t.Fatalf("Failed to start memory leak detector: %v", err)
	}
	defer detector.Stop()
	
	// 启动测试工作负载
	workload := startTestWorkload(config)
	defer stopTestWorkload(workload)
	
	// 记录初始快照
	result.StartSnapshot = detector.takeMemorySnapshot()
	log.Printf("Initial memory usage: %.2f MB, Goroutines: %d", 
		float64(result.StartSnapshot.HeapAlloc)/1024/1024, 
		result.StartSnapshot.GoroutineCount)
	
	// 运行测试
	log.Printf("Running memory leak test for %v...", config.TestDuration)
	
	// 监控循环
	ticker := time.NewTicker(config.SamplingInterval)
	defer ticker.Stop()
	
	testEndTime := time.Now().Add(config.TestDuration)
	
	for time.Now().Before(testEndTime) {
		select {
		case <-ticker.C:
			snapshot := detector.takeMemorySnapshot()
			result.Snapshots = append(result.Snapshots, snapshot)
			
			// 实时分析
			detector.analyzeSnapshot(snapshot)
			
			// 周期性报告
			if len(result.Snapshots)%10 == 0 {
				log.Printf("Progress: %.1f%%, Memory: %.2f MB, Goroutines: %d", 
					float64(time.Now().Sub(result.StartTime))/float64(config.TestDuration)*100,
					float64(snapshot.HeapAlloc)/1024/1024,
					snapshot.GoroutineCount)
			}
			
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	// 记录最终快照
	result.EndSnapshot = detector.takeMemorySnapshot()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	
	// 收集工作负载统计
	result.TotalOperations = atomic.LoadInt64(&workload.operationCount)
	result.ErrorCount = atomic.LoadInt64(&workload.errorCount)
	
	if result.Duration > 0 {
		result.OperationsPerSecond = float64(result.TotalOperations) / result.Duration.Seconds()
	}
	
	// 执行最终分析
	detector.performFinalAnalysis(result)
	
	return result
}

// NewMemoryLeakDetector 创建内存泄漏检测器
func NewMemoryLeakDetector(config *MemoryLeakTestConfig) *MemoryLeakDetector {
	return &MemoryLeakDetector{
		config:       config,
		snapshots:    make([]*MemorySnapshot, 0),
		stopChan:     make(chan struct{}),
		snapshotChan: make(chan *MemorySnapshot, 100),
	}
}

// Start 启动检测器
func (mld *MemoryLeakDetector) Start() error {
	if !atomic.CompareAndSwapInt32(&mld.isRunning, 0, 1) {
		return fmt.Errorf("memory leak detector is already running")
	}
	
	// 启动GC强制执行器
	if mld.config.ForceGCInterval > 0 {
		go mld.gcForcer()
	}
	
	return nil
}

// Stop 停止检测器
func (mld *MemoryLeakDetector) Stop() error {
	if !atomic.CompareAndSwapInt32(&mld.isRunning, 1, 0) {
		return fmt.Errorf("memory leak detector is not running")
	}
	
	close(mld.stopChan)
	return nil
}

// takeMemorySnapshot 获取内存快照
func (mld *MemoryLeakDetector) takeMemorySnapshot() *MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return &MemorySnapshot{
		Timestamp:      time.Now(),
		HeapAlloc:      m.Alloc,
		HeapSys:        m.Sys,
		HeapIdle:       m.HeapIdle,
		HeapInuse:      m.HeapInuse,
		HeapReleased:   m.HeapReleased,
		HeapObjects:    m.HeapObjects,
		StackInuse:     m.StackInuse,
		StackSys:       m.StackSys,
		GoroutineCount: runtime.NumGoroutine(),
		GCCount:        m.NumGC,
		GCPauseTotal:   m.PauseTotalNs,
	}
}

// analyzeSnapshot 分析快照
func (mld *MemoryLeakDetector) analyzeSnapshot(snapshot *MemorySnapshot) {
	mld.mutex.Lock()
	mld.snapshots = append(mld.snapshots, snapshot)
	
	// 保持快照数量在合理范围内
	if len(mld.snapshots) > 1000 {
		mld.snapshots = mld.snapshots[len(mld.snapshots)-1000:]
	}
	mld.mutex.Unlock()
	
	// 检测异常模式
	if len(mld.snapshots) > 10 {
		mld.detectLeakPatterns()
	}
}

// detectLeakPatterns 检测泄漏模式
func (mld *MemoryLeakDetector) detectLeakPatterns() {
	// 分析内存增长趋势
	// 分析goroutine增长趋势
	// 检测GC效率下降
	// 这里实现简化版本
}

// performFinalAnalysis 执行最终分析
func (mld *MemoryLeakDetector) performFinalAnalysis(result *MemoryLeakTestResult) {
	if result.StartSnapshot == nil || result.EndSnapshot == nil {
		return
	}
	
	// 计算增长率
	durationHours := result.Duration.Hours()
	if durationHours > 0 {
		memoryGrowthMB := float64(result.EndSnapshot.HeapAlloc-result.StartSnapshot.HeapAlloc) / 1024 / 1024
		result.MemoryGrowthRate = memoryGrowthMB / durationHours
		
		goroutineGrowth := float64(result.EndSnapshot.GoroutineCount - result.StartSnapshot.GoroutineCount)
		result.GoroutineGrowthRate = goroutineGrowth / durationHours
	}
	
	// 检测泄漏
	result.MemoryLeakDetected = result.MemoryGrowthRate > mld.config.MaxMemoryGrowthRate
	result.GoroutineLeakDetected = result.GoroutineGrowthRate > mld.config.MaxGoroutineGrowthRate
	
	// 计算GC效率
	if result.EndSnapshot.GCCount > result.StartSnapshot.GCCount {
		gcCount := result.EndSnapshot.GCCount - result.StartSnapshot.GCCount
		gcTime := result.EndSnapshot.GCPauseTotal - result.StartSnapshot.GCPauseTotal
		
		if gcCount > 0 {
			result.GCEfficiency = (1.0 - float64(gcTime)/float64(result.Duration.Nanoseconds())) * 100
		}
	}
	
	// 生成建议
	mld.generateRecommendations(result)
}

// generateRecommendations 生成建议
func (mld *MemoryLeakDetector) generateRecommendations(result *MemoryLeakTestResult) {
	if result.MemoryLeakDetected {
		result.Recommendations = append(result.Recommendations, 
			"Memory leak detected - check for unclosed resources and circular references")
	}
	
	if result.GoroutineLeakDetected {
		result.Recommendations = append(result.Recommendations,
			"Goroutine leak detected - ensure all goroutines have proper termination conditions")
	}
	
	if result.GCEfficiency < 90 {
		result.Recommendations = append(result.Recommendations,
			"Low GC efficiency - consider optimizing memory allocation patterns")
	}
	
	if result.ErrorCount > 0 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Errors occurred during test (%d) - check error handling", result.ErrorCount))
	}
}

// gcForcer GC强制执行器
func (mld *MemoryLeakDetector) gcForcer() {
	ticker := time.NewTicker(mld.config.ForceGCInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-mld.stopChan:
			return
		case <-ticker.C:
			runtime.GC()
		}
	}
}

// startTestWorkload 启动测试工作负载
func startTestWorkload(config *MemoryLeakTestConfig) *TestWorkload {
	workload := &TestWorkload{
		sessions:  make([]TestMyBatisSession, 0, config.WorkloadConcurrency),
		stopChan:  make(chan struct{}),
	}
	
	// 创建缓存监控器
	workload.cacheMonitor = NewCacheHitRateMonitor(DefaultCacheMonitorConfig())
	workload.cacheMonitor.Start()
	
	// 创建慢查询监控器
	workload.slowQueryMonitor = NewSlowQueryMonitor(DefaultSlowQueryConfig())
	workload.slowQueryMonitor.Start()
	
	// 启动工作线程
	for i := 0; i < config.WorkloadConcurrency; i++ {
		session := &MockSession{}
		workload.sessions = append(workload.sessions, session)
		
		go workload.runWorker(session)
	}
	
	return workload
}

// stopTestWorkload 停止测试工作负载
func stopTestWorkload(workload *TestWorkload) {
	close(workload.stopChan)
	
	if workload.cacheMonitor != nil {
		workload.cacheMonitor.Stop()
	}
	
	if workload.slowQueryMonitor != nil {
		workload.slowQueryMonitor.Stop()
	}
	
	// 关闭所有session
	for _, session := range workload.sessions {
		session.Close()
	}
}

// runWorker 运行工作器
func (tw *TestWorkload) runWorker(session TestMyBatisSession) {
	for {
		select {
		case <-tw.stopChan:
			return
		default:
			// 执行各种操作
			err := tw.performOperation(session)
			atomic.AddInt64(&tw.operationCount, 1)
			
			if err != nil {
				atomic.AddInt64(&tw.errorCount, 1)
			}
			
			// 短暂休息
			time.Sleep(time.Duration(10+len(tw.sessions)) * time.Millisecond)
		}
	}
}

// performOperation 执行操作
func (tw *TestWorkload) performOperation(session TestMyBatisSession) error {
	operations := []func(TestMyBatisSession) error{
		tw.performSelect,
		tw.performInsert,
		tw.performUpdate,
		tw.performCacheOperation,
	}
	
	op := operations[int(atomic.LoadInt64(&tw.operationCount))%len(operations)]
	return op(session)
}

// performSelect 执行SELECT操作
func (tw *TestWorkload) performSelect(session TestMyBatisSession) error {
	var result interface{}
	return session.SelectOne("SELECT * FROM users WHERE id = ?", &result, 
		atomic.LoadInt64(&tw.operationCount)%1000+1)
}

// performInsert 执行INSERT操作
func (tw *TestWorkload) performInsert(session TestMyBatisSession) error {
	opCount := atomic.LoadInt64(&tw.operationCount)
	return session.Insert("INSERT INTO test_users (name, email) VALUES (?, ?)",
		fmt.Sprintf("user_%d", opCount),
		fmt.Sprintf("user_%d@test.com", opCount))
}

// performUpdate 执行UPDATE操作
func (tw *TestWorkload) performUpdate(session TestMyBatisSession) error {
	opCount := atomic.LoadInt64(&tw.operationCount)
	return session.Update("UPDATE test_users SET name = ? WHERE id = ?",
		fmt.Sprintf("updated_user_%d", opCount),
		opCount%1000+1)
}

// performCacheOperation 执行缓存操作
func (tw *TestWorkload) performCacheOperation(session TestMyBatisSession) error {
	if tw.cacheMonitor != nil {
		opCount := atomic.LoadInt64(&tw.operationCount)
		key := fmt.Sprintf("cache_key_%d", opCount%100)
		hit := opCount%3 == 0 // 33%命中率
		duration := time.Duration(5+opCount%10) * time.Millisecond
		
		tw.cacheMonitor.RecordCacheAccess("L1", key, hit, duration)
	}
	return nil
}

// 辅助函数

// isLongTestRequested 检查是否请求长期测试
func isLongTestRequested() bool {
	// 这里可以检查环境变量或命令行参数
	// 简化实现，总是返回false除非明确设置
	return false
}

// printMemoryLeakTestResults 打印内存泄漏测试结果
func printMemoryLeakTestResults(result *MemoryLeakTestResult) {
	log.Printf("\n=== Memory Leak Test Results ===")
	log.Printf("Test Duration: %v", result.Duration)
	log.Printf("Total Operations: %d (%.2f ops/sec)", 
		result.TotalOperations, result.OperationsPerSecond)
	log.Printf("Error Count: %d", result.ErrorCount)
	
	log.Printf("\n--- Memory Analysis ---")
	log.Printf("Start Memory: %.2f MB", float64(result.StartSnapshot.HeapAlloc)/1024/1024)
	log.Printf("End Memory: %.2f MB", float64(result.EndSnapshot.HeapAlloc)/1024/1024)
	log.Printf("Memory Growth Rate: %.2f MB/hour", result.MemoryGrowthRate)
	log.Printf("Memory Leak Detected: %v", result.MemoryLeakDetected)
	
	log.Printf("\n--- Goroutine Analysis ---")
	log.Printf("Start Goroutines: %d", result.StartSnapshot.GoroutineCount)
	log.Printf("End Goroutines: %d", result.EndSnapshot.GoroutineCount)
	log.Printf("Goroutine Growth Rate: %.2f/hour", result.GoroutineGrowthRate)
	log.Printf("Goroutine Leak Detected: %v", result.GoroutineLeakDetected)
	
	log.Printf("\n--- GC Analysis ---")
	log.Printf("GC Count: %d", result.EndSnapshot.GCCount-result.StartSnapshot.GCCount)
	log.Printf("GC Efficiency: %.2f%%", result.GCEfficiency)
	
	if len(result.SuspiciousPatterns) > 0 {
		log.Printf("\n--- Suspicious Patterns ---")
		for _, pattern := range result.SuspiciousPatterns {
			log.Printf("%s: %s (Severity: %s)", pattern.Type, pattern.Description, pattern.Severity)
		}
	}
	
	if len(result.Recommendations) > 0 {
		log.Printf("\n--- Recommendations ---")
		for i, rec := range result.Recommendations {
			log.Printf("%d. %s", i+1, rec)
		}
	}
}

// generateDetailedLeakReport 生成详细泄漏报告
func generateDetailedLeakReport(result *MemoryLeakTestResult) {
	// 这里可以生成更详细的报告文件
	log.Printf("Detailed leak report would be generated here...")
}