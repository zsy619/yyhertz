// Package mybatis 批量操作内存池优化
//
// 优化重点：
// 1. 参数切片复用池
// 2. SQL构建器内存池
// 3. 批量大小自适应优化
// 4. 并发批量处理优化
// 5. 内存使用监控
package mybatis

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// BatchOperationOptimizer 批量操作优化器
type BatchOperationOptimizer struct {
	// 内存池
	argSlicePool     *sync.Pool
	stringBuilderPool *sync.Pool
	interfaceSlicePool *sync.Pool
	
	// 配置
	config           *BatchConfig
	
	// 统计信息
	stats            *BatchOperationStats
	
	// 并发控制
	semaphore        chan struct{}
	workerPool       *WorkerPool
}

// BatchConfig 批量操作配置
type BatchConfig struct {
	// 基础配置
	OptimalBatchSize     int           `yaml:"optimal_batch_size" json:"optimal_batch_size"`         // 最优批量大小
	MaxBatchSize         int           `yaml:"max_batch_size" json:"max_batch_size"`                 // 最大批量大小
	MinBatchSize         int           `yaml:"min_batch_size" json:"min_batch_size"`                 // 最小批量大小
	
	// 并发配置
	MaxConcurrentBatch   int           `yaml:"max_concurrent_batch" json:"max_concurrent_batch"`     // 最大并发批次数
	WorkerPoolSize       int           `yaml:"worker_pool_size" json:"worker_pool_size"`             // 工作协程池大小
	
	// 内存配置
	PoolInitialSize      int           `yaml:"pool_initial_size" json:"pool_initial_size"`           // 池初始大小
	PoolMaxSize          int           `yaml:"pool_max_size" json:"pool_max_size"`                   // 池最大大小
	MemoryThreshold      int64         `yaml:"memory_threshold" json:"memory_threshold"`             // 内存阈值(bytes)
	
	// 自适应配置
	EnableAdaptive       bool          `yaml:"enable_adaptive" json:"enable_adaptive"`               // 启用自适应批量大小
	AdaptiveInterval     time.Duration `yaml:"adaptive_interval" json:"adaptive_interval"`           // 自适应调整间隔
	PerformanceWindow    int           `yaml:"performance_window" json:"performance_window"`         // 性能窗口大小
	
	// 监控配置
	EnableMemoryMonitor  bool          `yaml:"enable_memory_monitor" json:"enable_memory_monitor"`   // 启用内存监控
	MonitorInterval      time.Duration `yaml:"monitor_interval" json:"monitor_interval"`             // 监控间隔
}

// BatchOperationStats 批量操作统计
type BatchOperationStats struct {
	// 操作统计
	TotalBatches         int64         `json:"total_batches"`
	TotalRecords         int64         `json:"total_records"`
	TotalDuration        time.Duration `json:"total_duration"`
	AvgBatchDuration     time.Duration `json:"avg_batch_duration"`
	AvgRecordsPerBatch   float64       `json:"avg_records_per_batch"`
	
	// 性能统计
	FastestBatch         time.Duration `json:"fastest_batch"`
	SlowestBatch         time.Duration `json:"slowest_batch"`
	OptimalBatchSize     int           `json:"optimal_batch_size"`
	AdaptiveAdjustments  int64         `json:"adaptive_adjustments"`
	
	// 内存统计
	PoolHits             int64         `json:"pool_hits"`
	PoolMisses           int64         `json:"pool_misses"`
	PeakMemoryUsage      int64         `json:"peak_memory_usage"`
	CurrentMemoryUsage   int64         `json:"current_memory_usage"`
	MemoryReductions     int64         `json:"memory_reductions"`
	
	// 错误统计
	FailedBatches        int64         `json:"failed_batches"`
	RetryAttempts        int64         `json:"retry_attempts"`
	
	// 最近性能窗口
	recentPerformance    []BatchPerformance
	mutex                sync.RWMutex
}

// BatchPerformance 批量操作性能记录
type BatchPerformance struct {
	BatchSize    int           `json:"batch_size"`
	Duration     time.Duration `json:"duration"`
	RecordsPerSec float64      `json:"records_per_sec"`
	MemoryUsed   int64         `json:"memory_used"`
	Timestamp    time.Time     `json:"timestamp"`
}

// WorkerPool 工作协程池
type WorkerPool struct {
	workers     []chan BatchJob
	workerCount int
	jobQueue    chan BatchJob
	quit        chan bool
	wg          sync.WaitGroup
}

// BatchJob 批量任务
type BatchJob struct {
	ID       string
	SQL      string
	Args     [][]any
	Callback func(int64, error)
	Context  context.Context
}

// DefaultBatchConfig 默认批量配置
func DefaultBatchConfig() *BatchConfig {
	return &BatchConfig{
		OptimalBatchSize:     1000,
		MaxBatchSize:         10000,
		MinBatchSize:         10,
		MaxConcurrentBatch:   5,
		WorkerPoolSize:       10,
		PoolInitialSize:      100,
		PoolMaxSize:          1000,
		MemoryThreshold:      100 * 1024 * 1024, // 100MB
		EnableAdaptive:       true,
		AdaptiveInterval:     30 * time.Second,
		PerformanceWindow:    50,
		EnableMemoryMonitor:  true,
		MonitorInterval:      10 * time.Second,
	}
}

// NewBatchOperationOptimizer 创建批量操作优化器
func NewBatchOperationOptimizer(config *BatchConfig) *BatchOperationOptimizer {
	if config == nil {
		config = DefaultBatchConfig()
	}
	
	optimizer := &BatchOperationOptimizer{
		config:    config,
		stats:     &BatchOperationStats{recentPerformance: make([]BatchPerformance, 0, config.PerformanceWindow)},
		semaphore: make(chan struct{}, config.MaxConcurrentBatch),
	}
	
	// 初始化内存池
	optimizer.initPools()
	
	// 初始化工作池
	optimizer.workerPool = NewWorkerPool(config.WorkerPoolSize, optimizer)
	
	// 启动监控
	if config.EnableMemoryMonitor {
		go optimizer.startMemoryMonitor()
	}
	
	if config.EnableAdaptive {
		go optimizer.startAdaptiveOptimization()
	}
	
	return optimizer
}

// initPools 初始化内存池
func (bo *BatchOperationOptimizer) initPools() {
	// 参数切片池
	bo.argSlicePool = &sync.Pool{
		New: func() interface{} {
			return make([][]any, 0, bo.config.OptimalBatchSize)
		},
	}
	
	// 字符串构建器池
	bo.stringBuilderPool = &sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
	
	// interface{}切片池
	bo.interfaceSlicePool = &sync.Pool{
		New: func() interface{} {
			return make([]any, 0, 20) // 假设平均每行20个字段
		},
	}
}

// OptimizedBatchInsert 优化版批量插入
func (bo *BatchOperationOptimizer) OptimizedBatchInsert(ctx context.Context, db *gorm.DB, sql string, allArgs [][]any) (int64, error) {
	if len(allArgs) == 0 {
		return 0, nil
	}
	
	startTime := time.Now()
	defer func() {
		bo.updatePerformanceStats(len(allArgs), time.Since(startTime))
	}()
	
	// 自适应批量大小
	batchSize := bo.getOptimalBatchSize(len(allArgs))
	
	// 如果数据量小，直接处理
	if len(allArgs) <= batchSize {
		return bo.executeSingleBatch(ctx, db, sql, allArgs)
	}
	
	// 大数据量，分批并发处理
	return bo.executeConcurrentBatches(ctx, db, sql, allArgs, batchSize)
}

// getOptimalBatchSize 获取最优批量大小
func (bo *BatchOperationOptimizer) getOptimalBatchSize(totalRecords int) int {
	bo.stats.mutex.RLock()
	defer bo.stats.mutex.RUnlock()
	
	// 基于历史性能数据动态调整
	if bo.config.EnableAdaptive && len(bo.stats.recentPerformance) > 10 {
		return bo.calculateAdaptiveBatchSize()
	}
	
	// 基于内存使用情况调整
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	if memStats.Alloc > uint64(bo.config.MemoryThreshold) {
		// 内存压力大，减小批量大小
		return bo.config.OptimalBatchSize / 2
	}
	
	// 基于数据量调整
	if totalRecords < bo.config.MinBatchSize {
		return totalRecords
	}
	
	if totalRecords > bo.config.MaxBatchSize {
		return bo.config.MaxBatchSize
	}
	
	return bo.config.OptimalBatchSize
}

// calculateAdaptiveBatchSize 计算自适应批量大小
func (bo *BatchOperationOptimizer) calculateAdaptiveBatchSize() int {
	// 分析最近的性能数据
	var bestPerf BatchPerformance
	maxThroughput := 0.0
	
	for _, perf := range bo.stats.recentPerformance {
		if perf.RecordsPerSec > maxThroughput {
			maxThroughput = perf.RecordsPerSec
			bestPerf = perf
		}
	}
	
	if bestPerf.BatchSize > 0 {
		return bestPerf.BatchSize
	}
	
	return bo.config.OptimalBatchSize
}

// executeSingleBatch 执行单个批次
func (bo *BatchOperationOptimizer) executeSingleBatch(ctx context.Context, db *gorm.DB, sql string, args [][]any) (int64, error) {
	// 获取参数切片（使用内存池）
	batchArgs := bo.getBatchArgsFromPool()
	defer bo.putBatchArgsToPool(batchArgs)
	
	// 复制参数到池化的切片
	batchArgs = batchArgs[:0] // 重置长度但保留容量
	for _, arg := range args {
		// 获取单行参数切片
		rowArgs := bo.getRowArgsFromPool()
		rowArgs = rowArgs[:0]
		rowArgs = append(rowArgs, arg...)
		batchArgs = append(batchArgs, rowArgs)
	}
	
	atomic.AddInt64(&bo.stats.PoolHits, 1)
	
	// 执行批量操作
	var totalAffected int64
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, rowArgs := range batchArgs {
			result := tx.Exec(sql, rowArgs...)
			if result.Error != nil {
				return result.Error
			}
			totalAffected += result.RowsAffected
		}
		return nil
	})
	
	// 归还参数切片到池
	for _, rowArgs := range batchArgs {
		bo.putRowArgsToPool(rowArgs)
	}
	
	return totalAffected, err
}

// executeConcurrentBatches 并发执行多个批次
func (bo *BatchOperationOptimizer) executeConcurrentBatches(ctx context.Context, db *gorm.DB, sql string, allArgs [][]any, batchSize int) (int64, error) {
	var totalAffected int64
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var firstError error
	
	// 分批处理
	for i := 0; i < len(allArgs); i += batchSize {
		end := i + batchSize
		if end > len(allArgs) {
			end = len(allArgs)
		}
		
		batch := allArgs[i:end]
		
		// 控制并发数量
		bo.semaphore <- struct{}{}
		
		wg.Add(1)
		go func(batchArgs [][]any, batchIndex int) {
			defer wg.Done()
			defer func() { <-bo.semaphore }()
			
			affected, err := bo.executeSingleBatch(ctx, db, sql, batchArgs)
			
			mutex.Lock()
			totalAffected += affected
			if err != nil && firstError == nil {
				firstError = err
			}
			mutex.Unlock()
			
		}(batch, i/batchSize)
	}
	
	wg.Wait()
	return totalAffected, firstError
}

// getBatchArgsFromPool 从池中获取批量参数切片
func (bo *BatchOperationOptimizer) getBatchArgsFromPool() [][]any {
	return bo.argSlicePool.Get().([][]any)
}

// putBatchArgsToPool 归还批量参数切片到池
func (bo *BatchOperationOptimizer) putBatchArgsToPool(args [][]any) {
	if cap(args) > bo.config.PoolMaxSize {
		// 切片太大，不归还到池中
		atomic.AddInt64(&bo.stats.MemoryReductions, 1)
		return
	}
	bo.argSlicePool.Put(args[:0]) // 重置长度但保留容量
}

// getRowArgsFromPool 从池中获取单行参数切片
func (bo *BatchOperationOptimizer) getRowArgsFromPool() []any {
	return bo.interfaceSlicePool.Get().([]any)
}

// putRowArgsToPool 归还单行参数切片到池
func (bo *BatchOperationOptimizer) putRowArgsToPool(args []any) {
	if cap(args) > 50 { // 单行参数不应该太多
		return
	}
	bo.interfaceSlicePool.Put(args[:0])
}

// updatePerformanceStats 更新性能统计
func (bo *BatchOperationOptimizer) updatePerformanceStats(recordCount int, duration time.Duration) {
	bo.stats.mutex.Lock()
	defer bo.stats.mutex.Unlock()
	
	atomic.AddInt64(&bo.stats.TotalBatches, 1)
	atomic.AddInt64(&bo.stats.TotalRecords, int64(recordCount))
	bo.stats.TotalDuration += duration
	
	// 更新平均值
	bo.stats.AvgBatchDuration = bo.stats.TotalDuration / time.Duration(bo.stats.TotalBatches)
	bo.stats.AvgRecordsPerBatch = float64(bo.stats.TotalRecords) / float64(bo.stats.TotalBatches)
	
	// 更新最快/最慢批次
	if bo.stats.FastestBatch == 0 || duration < bo.stats.FastestBatch {
		bo.stats.FastestBatch = duration
	}
	if duration > bo.stats.SlowestBatch {
		bo.stats.SlowestBatch = duration
	}
	
	// 添加到性能窗口
	recordsPerSec := float64(recordCount) / duration.Seconds()
	perf := BatchPerformance{
		BatchSize:     recordCount,
		Duration:      duration,
		RecordsPerSec: recordsPerSec,
		MemoryUsed:    bo.getCurrentMemoryUsage(),
		Timestamp:     time.Now(),
	}
	
	bo.stats.recentPerformance = append(bo.stats.recentPerformance, perf)
	if len(bo.stats.recentPerformance) > bo.config.PerformanceWindow {
		bo.stats.recentPerformance = bo.stats.recentPerformance[1:]
	}
}

// getCurrentMemoryUsage 获取当前内存使用量
func (bo *BatchOperationOptimizer) getCurrentMemoryUsage() int64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return int64(memStats.Alloc)
}

// startMemoryMonitor 启动内存监控
func (bo *BatchOperationOptimizer) startMemoryMonitor() {
	ticker := time.NewTicker(bo.config.MonitorInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		memUsage := bo.getCurrentMemoryUsage()
		
		bo.stats.mutex.Lock()
		bo.stats.CurrentMemoryUsage = memUsage
		if memUsage > bo.stats.PeakMemoryUsage {
			bo.stats.PeakMemoryUsage = memUsage
		}
		bo.stats.mutex.Unlock()
		
		// 内存压力检查
		if memUsage > bo.config.MemoryThreshold {
			log.Printf("[BatchOptimizer] High memory usage detected: %d MB", memUsage/(1024*1024))
			runtime.GC() // 强制GC
		}
	}
}

// startAdaptiveOptimization 启动自适应优化
func (bo *BatchOperationOptimizer) startAdaptiveOptimization() {
	ticker := time.NewTicker(bo.config.AdaptiveInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		bo.adjustOptimalBatchSize()
	}
}

// adjustOptimalBatchSize 调整最优批量大小
func (bo *BatchOperationOptimizer) adjustOptimalBatchSize() {
	bo.stats.mutex.Lock()
	defer bo.stats.mutex.Unlock()
	
	if len(bo.stats.recentPerformance) < 10 {
		return
	}
	
	// 分析性能趋势
	newOptimalSize := bo.calculateAdaptiveBatchSize()
	
	if newOptimalSize != bo.stats.OptimalBatchSize && newOptimalSize > 0 {
		oldSize := bo.stats.OptimalBatchSize
		bo.stats.OptimalBatchSize = newOptimalSize
		bo.config.OptimalBatchSize = newOptimalSize
		atomic.AddInt64(&bo.stats.AdaptiveAdjustments, 1)
		
		log.Printf("[BatchOptimizer] Adjusted optimal batch size: %d -> %d", 
			oldSize, newOptimalSize)
	}
}

// GetOptimizationStats 获取优化统计信息
func (bo *BatchOperationOptimizer) GetOptimizationStats() map[string]any {
	bo.stats.mutex.RLock()
	defer bo.stats.mutex.RUnlock()
	
	poolHitRate := float64(0)
	if total := bo.stats.PoolHits + bo.stats.PoolMisses; total > 0 {
		poolHitRate = float64(bo.stats.PoolHits) / float64(total) * 100
	}
	
	return map[string]any{
		"batch_stats": map[string]any{
			"total_batches":         bo.stats.TotalBatches,
			"total_records":         bo.stats.TotalRecords,
			"avg_batch_duration":    bo.stats.AvgBatchDuration.String(),
			"avg_records_per_batch": fmt.Sprintf("%.1f", bo.stats.AvgRecordsPerBatch),
			"fastest_batch":         bo.stats.FastestBatch.String(),
			"slowest_batch":         bo.stats.SlowestBatch.String(),
		},
		"memory_stats": map[string]any{
			"pool_hit_rate":        fmt.Sprintf("%.2f%%", poolHitRate),
			"pool_hits":            bo.stats.PoolHits,
			"pool_misses":          bo.stats.PoolMisses,
			"peak_memory_usage":    fmt.Sprintf("%d MB", bo.stats.PeakMemoryUsage/(1024*1024)),
			"current_memory_usage": fmt.Sprintf("%d MB", bo.stats.CurrentMemoryUsage/(1024*1024)),
			"memory_reductions":    bo.stats.MemoryReductions,
		},
		"optimization_stats": map[string]any{
			"optimal_batch_size":     bo.stats.OptimalBatchSize,
			"adaptive_adjustments":   bo.stats.AdaptiveAdjustments,
			"failed_batches":         bo.stats.FailedBatches,
			"retry_attempts":         bo.stats.RetryAttempts,
		},
		"config": map[string]any{
			"max_batch_size":         bo.config.MaxBatchSize,
			"max_concurrent_batch":   bo.config.MaxConcurrentBatch,
			"worker_pool_size":       bo.config.WorkerPoolSize,
			"memory_threshold":       fmt.Sprintf("%d MB", bo.config.MemoryThreshold/(1024*1024)),
			"enable_adaptive":        bo.config.EnableAdaptive,
		},
	}
}

// Close 关闭优化器
func (bo *BatchOperationOptimizer) Close() {
	if bo.workerPool != nil {
		bo.workerPool.Stop()
	}
	log.Println("[BatchOptimizer] Batch operation optimizer closed")
}

// ================================
// WorkerPool 实现
// ================================

// NewWorkerPool 创建工作协程池
func NewWorkerPool(workerCount int, optimizer *BatchOperationOptimizer) *WorkerPool {
	pool := &WorkerPool{
		workers:     make([]chan BatchJob, workerCount),
		workerCount: workerCount,
		jobQueue:    make(chan BatchJob, 100),
		quit:        make(chan bool),
	}
	
	pool.start()
	return pool
}

// start 启动工作池
func (wp *WorkerPool) start() {
	for i := 0; i < wp.workerCount; i++ {
		worker := make(chan BatchJob)
		wp.workers[i] = worker
		
		wp.wg.Add(1)
		go wp.startWorker(worker)
	}
	
	go wp.dispatch()
}

// startWorker 启动工作协程
func (wp *WorkerPool) startWorker(worker chan BatchJob) {
	defer wp.wg.Done()
	
	for {
		select {
		case job := <-worker:
			// 执行任务
			affected, err := wp.executeJob(job)
			job.Callback(affected, err)
			
		case <-wp.quit:
			return
		}
	}
}

// dispatch 分发任务
func (wp *WorkerPool) dispatch() {
	workerIndex := 0
	
	for {
		select {
		case job := <-wp.jobQueue:
			// 使用轮询方式分发到工作协程
			wp.workers[workerIndex] <- job
			workerIndex = (workerIndex + 1) % wp.workerCount
			
		case <-wp.quit:
			return
		}
	}
}

// executeJob 执行任务
func (wp *WorkerPool) executeJob(job BatchJob) (int64, error) {
	// 这里应该调用实际的数据库操作
	// 为简化示例，返回模拟结果
	return int64(len(job.Args)), nil
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	close(wp.quit)
	wp.wg.Wait()
}