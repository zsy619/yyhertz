package util

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Pool 协程池结构
type Pool struct {
	queue chan int
	wg    *sync.WaitGroup
}

// NewPool 创建协程池
func NewPool(size int) *Pool {
	if size <= 0 {
		size = runtime.NumCPU()
	}
	return &Pool{
		queue: make(chan int, size),
		wg:    &sync.WaitGroup{},
	}
}

// Add 添加任务到池中
func (p *Pool) Add(delta int) {
	for i := 0; i < delta; i++ { // delta > 0
		p.queue <- 1
	}
	for i := 0; i > delta; i-- { // delta < 0
		<-p.queue
	}
	p.wg.Add(delta)
}

// Done 标记任务完成
func (p *Pool) Done() {
	<-p.queue
	p.wg.Done()
}

// Wait 等待所有任务完成
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Size 获取池大小
func (p *Pool) Size() int {
	return cap(p.queue)
}

// Running 获取正在运行的任务数
func (p *Pool) Running() int {
	return len(p.queue)
}

// Available 获取可用的槽位数
func (p *Pool) Available() int {
	return cap(p.queue) - len(p.queue)
}

// WorkerPool 工作池结构
type WorkerPool struct {
	workerCount int
	taskQueue   chan func()
	quit        chan bool
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workerCount int, queueSize int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}
	if queueSize <= 0 {
		queueSize = workerCount * 2
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan func(), queueSize),
		quit:        make(chan bool),
		ctx:         ctx,
		cancel:      cancel,
	}

	return pool
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case task := <-wp.taskQueue:
			if task != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("Worker %d panic: %v\n", id, r)
						}
					}()
					task()
				}()
			}
		case <-wp.quit:
			fmt.Printf("Worker %d stopping\n", id)
			return
		case <-wp.ctx.Done():
			fmt.Printf("Worker %d cancelled\n", id)
			return
		}
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(task func()) error {
	select {
	case wp.taskQueue <- task:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is closed")
	default:
		return fmt.Errorf("task queue is full")
	}
}

// SubmitWithTimeout 带超时的任务提交
func (wp *WorkerPool) SubmitWithTimeout(task func(), timeout time.Duration) error {
	select {
	case wp.taskQueue <- task:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is closed")
	case <-time.After(timeout):
		return fmt.Errorf("submit timeout")
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.quit)
	wp.wg.Wait()
	close(wp.taskQueue)
}

// Size 获取工作池大小
func (wp *WorkerPool) Size() int {
	return wp.workerCount
}

// QueueSize 获取任务队列大小
func (wp *WorkerPool) QueueSize() int {
	return cap(wp.taskQueue)
}

// QueueLength 获取当前队列长度
func (wp *WorkerPool) QueueLength() int {
	return len(wp.taskQueue)
}

// IsRunning 检查工作池是否在运行
func (wp *WorkerPool) IsRunning() bool {
	select {
	case <-wp.ctx.Done():
		return false
	default:
		return true
	}
}

// TaskResult 任务结果结构
type TaskResult struct {
	Result any
	Error  error
}

// Task 任务结构
type Task struct {
	ID     string
	Func   func() (any, error)
	Result chan TaskResult
}

// NewTask 创建新任务
func NewTask(id string, fn func() (any, error)) *Task {
	return &Task{
		ID:     id,
		Func:   fn,
		Result: make(chan TaskResult, 1),
	}
}

// Execute 执行任务
func (t *Task) Execute() {
	defer close(t.Result)

	result, err := t.Func()
	t.Result <- TaskResult{
		Result: result,
		Error:  err,
	}
}

// Wait 等待任务完成
func (t *Task) Wait() TaskResult {
	return <-t.Result
}

// WaitWithTimeout 带超时等待任务完成
func (t *Task) WaitWithTimeout(timeout time.Duration) (TaskResult, error) {
	select {
	case result := <-t.Result:
		return result, nil
	case <-time.After(timeout):
		return TaskResult{}, fmt.Errorf("task timeout")
	}
}

// ResultPool 结果池，用于处理有返回值的任务
type ResultPool struct {
	*WorkerPool
	results sync.Map
}

// NewResultPool 创建结果池
func NewResultPool(workerCount int, queueSize int) *ResultPool {
	return &ResultPool{
		WorkerPool: NewWorkerPool(workerCount, queueSize),
	}
}

// SubmitTask 提交有返回值的任务
func (rp *ResultPool) SubmitTask(task *Task) error {
	return rp.Submit(func() {
		task.Execute()
		rp.results.Store(task.ID, task)
	})
}

// GetResult 获取任务结果
func (rp *ResultPool) GetResult(taskID string) (*Task, bool) {
	if value, ok := rp.results.Load(taskID); ok {
		if task, ok := value.(*Task); ok {
			rp.results.Delete(taskID) // 获取后删除
			return task, true
		}
	}
	return nil, false
}

// GetAllResults 获取所有结果
func (rp *ResultPool) GetAllResults() map[string]*Task {
	results := make(map[string]*Task)
	rp.results.Range(func(key, value any) bool {
		if taskID, ok := key.(string); ok {
			if task, ok := value.(*Task); ok {
				results[taskID] = task
			}
		}
		return true
	})

	// 清空结果
	rp.results = sync.Map{}

	return results
}

// BatchProcessor 批处理器
type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
	processor func([]any) error
	buffer    []any
	mutex     sync.Mutex
	timer     *time.Timer
}

// NewBatchProcessor 创建批处理器
func NewBatchProcessor(batchSize int, timeout time.Duration, processor func([]any) error) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		timeout:   timeout,
		processor: processor,
		buffer:    make([]any, 0, batchSize),
	}
}

// Add 添加数据到批处理器
func (bp *BatchProcessor) Add(item any) error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.buffer = append(bp.buffer, item)

	// 如果是第一个元素，启动定时器
	if len(bp.buffer) == 1 {
		bp.timer = time.AfterFunc(bp.timeout, func() {
			bp.flush()
		})
	}

	// 如果达到批大小，立即处理
	if len(bp.buffer) >= bp.batchSize {
		bp.stopTimer()
		return bp.processBatch()
	}

	return nil
}

// Flush 强制处理缓冲区中的所有数据
func (bp *BatchProcessor) Flush() error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.stopTimer()
	return bp.processBatch()
}

// flush 内部刷新方法
func (bp *BatchProcessor) flush() {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.processBatch()
}

// stopTimer 停止定时器
func (bp *BatchProcessor) stopTimer() {
	if bp.timer != nil {
		bp.timer.Stop()
		bp.timer = nil
	}
}

// processBatch 处理批数据
func (bp *BatchProcessor) processBatch() error {
	if len(bp.buffer) == 0 {
		return nil
	}

	batch := make([]any, len(bp.buffer))
	copy(batch, bp.buffer)
	bp.buffer = bp.buffer[:0] // 清空缓冲区

	return bp.processor(batch)
}

// 全局默认池
var (
	DefaultPool       *Pool
	DefaultWorkerPool *WorkerPool
)

// 初始化默认池
func init() {
	DefaultPool = NewPool(runtime.NumCPU())
	DefaultWorkerPool = NewWorkerPool(runtime.NumCPU(), runtime.NumCPU()*2)
	DefaultWorkerPool.Start()
}

// 便捷函数
func Submit(task func()) error {
	return DefaultWorkerPool.Submit(task)
}

func SubmitWithTimeout(task func(), timeout time.Duration) error {
	return DefaultWorkerPool.SubmitWithTimeout(task, timeout)
}
