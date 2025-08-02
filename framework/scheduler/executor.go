package scheduler

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ExecutorPool 执行器池
type ExecutorPool struct {
	workers     int
	taskQueue   chan *TaskExecution
	workerGroup sync.WaitGroup
	stopChan    chan struct{}
	running     int32

	// 统计信息
	totalExecuted   int64
	totalSuccessful int64
	totalFailed     int64
	totalCanceled   int64

	// 配置
	config *ExecutorConfig

	// 回调函数
	onBeforeExecute func(*TaskExecution)
	onAfterExecute  func(*TaskExecution, error)
	onPanic         func(*TaskExecution, any)

	mutex sync.RWMutex
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	WorkerCount    int           `json:"worker_count"`    // 工作协程数量
	QueueSize      int           `json:"queue_size"`      // 任务队列大小
	MaxRetries     int           `json:"max_retries"`     // 最大重试次数
	RetryDelay     time.Duration `json:"retry_delay"`     // 重试延迟
	ExecuteTimeout time.Duration `json:"execute_timeout"` // 执行超时时间
	EnableMetrics  bool          `json:"enable_metrics"`  // 启用指标收集
	EnableRecovery bool          `json:"enable_recovery"` // 启用panic恢复
}

// DefaultExecutorConfig 默认执行器配置
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		WorkerCount:    runtime.NumCPU(),
		QueueSize:      1000,
		MaxRetries:     3,
		RetryDelay:     time.Second * 5,
		ExecuteTimeout: time.Minute * 30,
		EnableMetrics:  true,
		EnableRecovery: true,
	}
}

// TaskExecution 任务执行上下文
type TaskExecution struct {
	Task        *Task
	Context     context.Context
	CancelFunc  context.CancelFunc // 重命名避免与Cancel方法冲突
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	RetryCount  int
	LastError   error
	Status      ExecutionStatus
	WorkerID    int
	ExecutionID string
	Metadata    map[string]any

	mutex sync.RWMutex
}

// ExecutionStatus 执行状态
type ExecutionStatus int

const (
	ExecutionStatusPending ExecutionStatus = iota
	ExecutionStatusRunning
	ExecutionStatusCompleted
	ExecutionStatusFailed
	ExecutionStatusCanceled
	ExecutionStatusRetrying
)

// String 执行状态字符串
func (es ExecutionStatus) String() string {
	switch es {
	case ExecutionStatusPending:
		return "PENDING"
	case ExecutionStatusRunning:
		return "RUNNING"
	case ExecutionStatusCompleted:
		return "COMPLETED"
	case ExecutionStatusFailed:
		return "FAILED"
	case ExecutionStatusCanceled:
		return "CANCELED"
	case ExecutionStatusRetrying:
		return "RETRYING"
	default:
		return "UNKNOWN"
	}
}

// NewExecutorPool 创建执行器池
func NewExecutorPool(config *ExecutorConfig) *ExecutorPool {
	if config == nil {
		config = DefaultExecutorConfig()
	}

	return &ExecutorPool{
		workers:   config.WorkerCount,
		taskQueue: make(chan *TaskExecution, config.QueueSize),
		stopChan:  make(chan struct{}),
		config:    config,
	}
}

// SetOnBeforeExecute 设置执行前回调
func (ep *ExecutorPool) SetOnBeforeExecute(fn func(*TaskExecution)) {
	ep.onBeforeExecute = fn
}

// SetOnAfterExecute 设置执行后回调
func (ep *ExecutorPool) SetOnAfterExecute(fn func(*TaskExecution, error)) {
	ep.onAfterExecute = fn
}

// SetOnPanic 设置panic回调
func (ep *ExecutorPool) SetOnPanic(fn func(*TaskExecution, any)) {
	ep.onPanic = fn
}

// Start 启动执行器池
func (ep *ExecutorPool) Start() error {
	if atomic.LoadInt32(&ep.running) == 1 {
		return fmt.Errorf("executor pool is already running")
	}

	atomic.StoreInt32(&ep.running, 1)

	// 启动工作协程
	for i := 0; i < ep.workers; i++ {
		ep.workerGroup.Add(1)
		go ep.worker(i)
	}

	config.Infof("Executor pool started with %d workers", ep.workers)
	return nil
}

// Stop 停止执行器池
func (ep *ExecutorPool) Stop() error {
	if atomic.LoadInt32(&ep.running) == 0 {
		return fmt.Errorf("executor pool is not running")
	}

	atomic.StoreInt32(&ep.running, 0)
	close(ep.stopChan)

	// 等待所有工作协程结束
	ep.workerGroup.Wait()

	config.Info("Executor pool stopped")
	return nil
}

// IsRunning 检查是否运行中
func (ep *ExecutorPool) IsRunning() bool {
	return atomic.LoadInt32(&ep.running) == 1
}

// Execute 提交任务执行
func (ep *ExecutorPool) Execute(task *Task) (*TaskExecution, error) {
	if !ep.IsRunning() {
		return nil, fmt.Errorf("executor pool is not running")
	}

	// 创建执行上下文
	ctx, cancel := context.WithTimeout(context.Background(), ep.config.ExecuteTimeout)

	execution := &TaskExecution{
		Task:        task,
		Context:     ctx,
		CancelFunc:  cancel, // 使用重命名后的字段
		StartTime:   time.Now(),
		Status:      ExecutionStatusPending,
		ExecutionID: generateExecutionID(),
		Metadata:    make(map[string]any),
	}

	// 提交到队列
	select {
	case ep.taskQueue <- execution:
		return execution, nil
	default:
		cancel()
		return nil, fmt.Errorf("task queue is full")
	}
}

// ExecuteSync 同步执行任务
func (ep *ExecutorPool) ExecuteSync(task *Task) error {
	execution, err := ep.Execute(task)
	if err != nil {
		return err
	}

	// 等待执行完成
	for execution.Status == ExecutionStatusPending || execution.Status == ExecutionStatusRunning || execution.Status == ExecutionStatusRetrying {
		time.Sleep(time.Millisecond * 100)
	}

	return execution.LastError
}

// worker 工作协程
func (ep *ExecutorPool) worker(workerID int) {
	defer ep.workerGroup.Done()

	config.Infof("Worker %d started", workerID)

	for {
		select {
		case <-ep.stopChan:
			config.Infof("Worker %d stopped", workerID)
			return

		case execution := <-ep.taskQueue:
			ep.executeTask(execution, workerID)
		}
	}
}

// executeTask 执行任务
func (ep *ExecutorPool) executeTask(execution *TaskExecution, workerID int) {
	execution.WorkerID = workerID
	execution.Status = ExecutionStatusRunning
	execution.StartTime = time.Now()

	// 执行前回调
	if ep.onBeforeExecute != nil {
		ep.onBeforeExecute(execution)
	}

	var err error

	// 执行任务（带panic恢复）
	if ep.config.EnableRecovery {
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("task panicked: %v", r)

					// panic回调
					if ep.onPanic != nil {
						ep.onPanic(execution, r)
					}

					// 记录stack trace
					stack := make([]byte, 4096)
					n := runtime.Stack(stack, false)
					config.Errorf("Task %s panic: %v\nStack:\n%s",
						execution.Task.ID, r, stack[:n])
				}
			}()

			err = execution.Task.Job.Execute(execution.Context)
		}()
	} else {
		err = execution.Task.Job.Execute(execution.Context)
	}

	// 更新执行结果
	execution.EndTime = time.Now()
	execution.Duration = execution.EndTime.Sub(execution.StartTime)
	execution.LastError = err

	// 处理执行结果
	if err != nil {
		execution.Status = ExecutionStatusFailed
		execution.RetryCount++

		atomic.AddInt64(&ep.totalFailed, 1)

		// 重试逻辑
		if execution.RetryCount < ep.config.MaxRetries {
			execution.Status = ExecutionStatusRetrying

			// 延迟重试
			go func() {
				time.Sleep(ep.config.RetryDelay)

				// 重新提交任务
				select {
				case ep.taskQueue <- execution:
					config.Infof("Task %s retry %d/%d scheduled",
						execution.Task.ID, execution.RetryCount, ep.config.MaxRetries)
				default:
					execution.Status = ExecutionStatusFailed
					config.Errorf("Failed to schedule retry for task %s: queue full", execution.Task.ID)
				}
			}()
		}
	} else {
		execution.Status = ExecutionStatusCompleted
		atomic.AddInt64(&ep.totalSuccessful, 1)
	}

	atomic.AddInt64(&ep.totalExecuted, 1)

	// 执行后回调
	if ep.onAfterExecute != nil {
		ep.onAfterExecute(execution, err)
	}

	// 记录执行日志
	if execution.Status == ExecutionStatusCompleted {
		config.Infof("Task %s completed in %v by worker %d",
			execution.Task.ID, execution.Duration, workerID)
	} else if execution.Status == ExecutionStatusFailed {
		config.Errorf("Task %s failed after %d retries in %v by worker %d: %v",
			execution.Task.ID, execution.RetryCount, execution.Duration, workerID, err)
	}
}

// GetStats 获取执行器统计信息
func (ep *ExecutorPool) GetStats() *ExecutorStats {
	return &ExecutorStats{
		WorkerCount:     ep.workers,
		QueueSize:       len(ep.taskQueue),
		QueueCapacity:   cap(ep.taskQueue),
		TotalExecuted:   atomic.LoadInt64(&ep.totalExecuted),
		TotalSuccessful: atomic.LoadInt64(&ep.totalSuccessful),
		TotalFailed:     atomic.LoadInt64(&ep.totalFailed),
		TotalCanceled:   atomic.LoadInt64(&ep.totalCanceled),
		IsRunning:       ep.IsRunning(),
	}
}

// ExecutorStats 执行器统计信息
type ExecutorStats struct {
	WorkerCount     int   `json:"worker_count"`
	QueueSize       int   `json:"queue_size"`
	QueueCapacity   int   `json:"queue_capacity"`
	TotalExecuted   int64 `json:"total_executed"`
	TotalSuccessful int64 `json:"total_successful"`
	TotalFailed     int64 `json:"total_failed"`
	TotalCanceled   int64 `json:"total_canceled"`
	IsRunning       bool  `json:"is_running"`
}

// ============= 任务执行上下文方法 =============

// SetMetadata 设置执行元数据
func (te *TaskExecution) SetMetadata(key string, value any) {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	te.Metadata[key] = value
}

// GetMetadata 获取执行元数据
func (te *TaskExecution) GetMetadata(key string) (any, bool) {
	te.mutex.RLock()
	defer te.mutex.RUnlock()
	value, exists := te.Metadata[key]
	return value, exists
}

// Cancel 取消执行
func (te *TaskExecution) Cancel() {
	te.mutex.Lock()
	defer te.mutex.Unlock()

	if te.CancelFunc != nil {
		te.CancelFunc() // 调用正确的取消函数
	}
	te.Status = ExecutionStatusCanceled
}

// IsCompleted 检查是否完成
func (te *TaskExecution) IsCompleted() bool {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	return te.Status == ExecutionStatusCompleted ||
		te.Status == ExecutionStatusFailed ||
		te.Status == ExecutionStatusCanceled
}

// ============= 高级执行器 =============

// AdvancedExecutor 高级执行器
type AdvancedExecutor struct {
	pool         *ExecutorPool
	scheduler    *Scheduler
	executions   map[string]*TaskExecution
	executionsMx sync.RWMutex

	// 执行策略
	strategies map[string]ExecutionStrategy

	// 监控
	monitor *ExecutionMonitor
}

// ExecutionStrategy 执行策略接口
type ExecutionStrategy interface {
	ShouldExecute(task *Task, context ExecutionContext) bool
	OnExecute(execution *TaskExecution)
	OnComplete(execution *TaskExecution, err error)
	OnRetry(execution *TaskExecution, retryCount int)
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	CurrentTime     time.Time
	LastExecution   *TaskExecution
	QueueSize       int
	RunningTasks    int
	SystemLoad      float64
	AvailableMemory int64
}

// NewAdvancedExecutor 创建高级执行器
func NewAdvancedExecutor(config *ExecutorConfig) *AdvancedExecutor {
	pool := NewExecutorPool(config)

	ae := &AdvancedExecutor{
		pool:       pool,
		executions: make(map[string]*TaskExecution),
		strategies: make(map[string]ExecutionStrategy),
		monitor:    NewExecutionMonitor(),
	}

	// 设置执行回调
	pool.SetOnBeforeExecute(ae.onBeforeExecute)
	pool.SetOnAfterExecute(ae.onAfterExecute)
	pool.SetOnPanic(ae.onPanic)

	return ae
}

// RegisterStrategy 注册执行策略
func (ae *AdvancedExecutor) RegisterStrategy(name string, strategy ExecutionStrategy) {
	ae.strategies[name] = strategy
}

// Start 启动高级执行器
func (ae *AdvancedExecutor) Start() error {
	if err := ae.pool.Start(); err != nil {
		return err
	}

	ae.monitor.Start()
	return nil
}

// Stop 停止高级执行器
func (ae *AdvancedExecutor) Stop() error {
	ae.monitor.Stop()
	return ae.pool.Stop()
}

// ExecuteWithStrategy 使用策略执行任务
func (ae *AdvancedExecutor) ExecuteWithStrategy(task *Task, strategyName string) (*TaskExecution, error) {
	strategy, exists := ae.strategies[strategyName]
	if !exists {
		return nil, fmt.Errorf("execution strategy '%s' not found", strategyName)
	}

	// 构建执行上下文
	context := ae.buildExecutionContext(task)

	// 检查是否应该执行
	if !strategy.ShouldExecute(task, context) {
		return nil, fmt.Errorf("execution strategy '%s' rejected task execution", strategyName)
	}

	// 执行任务
	execution, err := ae.pool.Execute(task)
	if err != nil {
		return nil, err
	}

	// 记录执行
	ae.executionsMx.Lock()
	ae.executions[execution.ExecutionID] = execution
	ae.executionsMx.Unlock()

	// 调用策略回调
	strategy.OnExecute(execution)

	return execution, nil
}

// buildExecutionContext 构建执行上下文
func (ae *AdvancedExecutor) buildExecutionContext(task *Task) ExecutionContext {
	stats := ae.pool.GetStats()

	return ExecutionContext{
		CurrentTime:     time.Now(),
		LastExecution:   ae.getLastExecution(task.ID),
		QueueSize:       stats.QueueSize,
		RunningTasks:    stats.QueueSize, // 简化实现
		SystemLoad:      ae.monitor.GetSystemLoad(),
		AvailableMemory: ae.monitor.GetAvailableMemory(),
	}
}

// getLastExecution 获取最后一次执行
func (ae *AdvancedExecutor) getLastExecution(taskID string) *TaskExecution {
	ae.executionsMx.RLock()
	defer ae.executionsMx.RUnlock()

	var lastExec *TaskExecution
	for _, exec := range ae.executions {
		if exec.Task.ID == taskID {
			if lastExec == nil || exec.StartTime.After(lastExec.StartTime) {
				lastExec = exec
			}
		}
	}

	return lastExec
}

// 回调方法
func (ae *AdvancedExecutor) onBeforeExecute(execution *TaskExecution) {
	ae.monitor.RecordExecutionStart(execution)
}

func (ae *AdvancedExecutor) onAfterExecute(execution *TaskExecution, err error) {
	ae.monitor.RecordExecutionEnd(execution, err)

	// 调用策略回调
	for _, strategy := range ae.strategies {
		strategy.OnComplete(execution, err)
	}
}

func (ae *AdvancedExecutor) onPanic(execution *TaskExecution, panicValue any) {
	ae.monitor.RecordPanic(execution, panicValue)
}

// ============= 预定义执行策略 =============

// ThrottleStrategy 节流策略
type ThrottleStrategy struct {
	MaxConcurrent int
	MinInterval   time.Duration
	lastExecution map[string]time.Time
	mutex         sync.RWMutex
}

// NewThrottleStrategy 创建节流策略
func NewThrottleStrategy(maxConcurrent int, minInterval time.Duration) *ThrottleStrategy {
	return &ThrottleStrategy{
		MaxConcurrent: maxConcurrent,
		MinInterval:   minInterval,
		lastExecution: make(map[string]time.Time),
	}
}

// ShouldExecute 检查是否应该执行
func (ts *ThrottleStrategy) ShouldExecute(task *Task, context ExecutionContext) bool {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// 检查并发限制
	if context.RunningTasks >= ts.MaxConcurrent {
		return false
	}

	// 检查间隔限制
	if lastTime, exists := ts.lastExecution[task.ID]; exists {
		if time.Since(lastTime) < ts.MinInterval {
			return false
		}
	}

	return true
}

// OnExecute 执行回调
func (ts *ThrottleStrategy) OnExecute(execution *TaskExecution) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	ts.lastExecution[execution.Task.ID] = execution.StartTime
}

// OnComplete 完成回调
func (ts *ThrottleStrategy) OnComplete(execution *TaskExecution, err error) {
	// 节流策略不需要处理完成事件
}

// OnRetry 重试回调
func (ts *ThrottleStrategy) OnRetry(execution *TaskExecution, retryCount int) {
	// 节流策略不需要处理重试事件
}

// ============= 辅助函数 =============

var executionIDCounter int64

// generateExecutionID 生成执行ID
func generateExecutionID() string {
	id := atomic.AddInt64(&executionIDCounter, 1)
	return fmt.Sprintf("exec_%d_%d", time.Now().Unix(), id)
}
