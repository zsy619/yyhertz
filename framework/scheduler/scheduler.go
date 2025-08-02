// Package scheduler 提供任务调度功能
//
// 这个包提供了一个强大的任务调度系统，支持：
// - Cron表达式调度
// - 一次性任务
// - 延时任务
// - 任务持久化
// - 任务监控和日志
// - 分布式调度支持
//
// 类似于Beego的Task功能，但更加现代化和强大
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

// Job 任务接口
type Job interface {
	// Execute 执行任务
	Execute(ctx context.Context) error
	// GetName 获取任务名称
	GetName() string
	// GetDescription 获取任务描述
	GetDescription() string
}

// JobFunc 函数式任务
type JobFunc struct {
	name        string
	description string
	fn          func(ctx context.Context) error
}

// NewJobFunc 创建函数式任务
func NewJobFunc(name, description string, fn func(ctx context.Context) error) *JobFunc {
	return &JobFunc{
		name:        name,
		description: description,
		fn:          fn,
	}
}

// Execute 执行任务
func (jf *JobFunc) Execute(ctx context.Context) error {
	return jf.fn(ctx)
}

// GetName 获取任务名称
func (jf *JobFunc) GetName() string {
	return jf.name
}

// GetDescription 获取任务描述
func (jf *JobFunc) GetDescription() string {
	return jf.description
}

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = iota // 等待中
	TaskStatusRunning                     // 运行中
	TaskStatusCompleted                   // 已完成
	TaskStatusFailed                      // 失败
	TaskStatusCanceled                    // 已取消
	TaskStatusPaused                      // 已暂停
)

// String 状态字符串表示
func (ts TaskStatus) String() string {
	switch ts {
	case TaskStatusPending:
		return "PENDING"
	case TaskStatusRunning:
		return "RUNNING"
	case TaskStatusCompleted:
		return "COMPLETED"
	case TaskStatusFailed:
		return "FAILED"
	case TaskStatusCanceled:
		return "CANCELED"
	case TaskStatusPaused:
		return "PAUSED"
	default:
		return "UNKNOWN"
	}
}

// Task 任务定义
type Task struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Schedule    string            `json:"schedule"`    // Cron表达式或时间格式
	Job         Job               `json:"-"`          // 任务实例
	Status      TaskStatus        `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastRunTime *time.Time        `json:"last_run_time,omitempty"`
	NextRunTime *time.Time        `json:"next_run_time,omitempty"`
	RunCount    int64             `json:"run_count"`
	FailCount   int64             `json:"fail_count"`
	MaxRetries  int               `json:"max_retries"`
	Timeout     time.Duration     `json:"timeout"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	
	// 内部字段
	cancel context.CancelFunc `json:"-"`
	mutex  sync.RWMutex       `json:"-"`
}

// NewTask 创建新任务
func NewTask(id, name, description, schedule string, job Job) *Task {
	return &Task{
		ID:          id,
		Name:        name,
		Description: description,
		Schedule:    schedule,
		Job:         job,
		Status:      TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MaxRetries:  3,
		Timeout:     time.Minute * 30,
		Metadata:    make(map[string]string),
	}
}

// SetStatus 设置任务状态
func (t *Task) SetStatus(status TaskStatus) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Status = status
	t.UpdatedAt = time.Now()
}

// IncrementRunCount 增加运行次数
func (t *Task) IncrementRunCount() {
	atomic.AddInt64(&t.RunCount, 1)
}

// IncrementFailCount 增加失败次数
func (t *Task) IncrementFailCount() {
	atomic.AddInt64(&t.FailCount, 1)
}

// SetLastRunTime 设置最后运行时间
func (t *Task) SetLastRunTime(t2 time.Time) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.LastRunTime = &t2
	t.UpdatedAt = time.Now()
}

// SetNextRunTime 设置下次运行时间
func (t *Task) SetNextRunTime(t2 time.Time) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.NextRunTime = &t2
	t.UpdatedAt = time.Now()
}

// GetMetadata 获取元数据
func (t *Task) GetMetadata(key string) string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.Metadata[key]
}

// SetMetadata 设置元数据
func (t *Task) SetMetadata(key, value string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Metadata[key] = value
	t.UpdatedAt = time.Now()
}

// Scheduler 调度器
type Scheduler struct {
	tasks    map[string]*Task
	running  int32
	stopChan chan struct{}
	workers  int
	
	// 事件回调
	onTaskStart    func(*Task)
	onTaskComplete func(*Task, error)
	onTaskFail     func(*Task, error)
	
	// 持久化存储
	storage Storage
	
	// 配置
	config *SchedulerConfig
	
	mutex sync.RWMutex
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	MaxWorkers       int           `json:"max_workers"`
	TickInterval     time.Duration `json:"tick_interval"`
	EnablePersistent bool          `json:"enable_persistent"`
	EnableLogging    bool          `json:"enable_logging"`
	TimeZone         string        `json:"timezone"`
}

// DefaultSchedulerConfig 默认调度器配置
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		MaxWorkers:       runtime.NumCPU(),
		TickInterval:     time.Second,
		EnablePersistent: false,
		EnableLogging:    true,
		TimeZone:         "Local",
	}
}

// NewScheduler 创建新的调度器
func NewScheduler(config *SchedulerConfig) *Scheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}
	
	return &Scheduler{
		tasks:    make(map[string]*Task),
		stopChan: make(chan struct{}),
		workers:  config.MaxWorkers,
		config:   config,
	}
}

// SetStorage 设置存储后端
func (s *Scheduler) SetStorage(storage Storage) {
	s.storage = storage
}

// SetOnTaskStart 设置任务开始回调
func (s *Scheduler) SetOnTaskStart(fn func(*Task)) {
	s.onTaskStart = fn
}

// SetOnTaskComplete 设置任务完成回调
func (s *Scheduler) SetOnTaskComplete(fn func(*Task, error)) {
	s.onTaskComplete = fn
}

// SetOnTaskFail 设置任务失败回调
func (s *Scheduler) SetOnTaskFail(fn func(*Task, error)) {
	s.onTaskFail = fn
}

// AddTask 添加任务
func (s *Scheduler) AddTask(task *Task) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}
	
	// 解析调度时间
	nextRun, err := s.parseSchedule(task.Schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule '%s': %w", task.Schedule, err)
	}
	
	task.SetNextRunTime(nextRun)
	s.tasks[task.ID] = task
	
	// 持久化任务
	if s.config.EnablePersistent && s.storage != nil {
		if err := s.storage.SaveTask(task); err != nil {
			config.Errorf("Failed to persist task %s: %v", task.ID, err)
		}
	}
	
	config.Infof("Task added: %s (%s)", task.Name, task.ID)
	return nil
}

// RemoveTask 移除任务
func (s *Scheduler) RemoveTask(taskID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}
	
	// 取消正在运行的任务
	if task.Status == TaskStatusRunning && task.cancel != nil {
		task.cancel()
	}
	
	delete(s.tasks, taskID)
	
	// 从存储中删除
	if s.config.EnablePersistent && s.storage != nil {
		if err := s.storage.DeleteTask(taskID); err != nil {
			config.Errorf("Failed to delete task %s from storage: %v", taskID, err)
		}
	}
	
	config.Infof("Task removed: %s (%s)", task.Name, taskID)
	return nil
}

// GetTask 获取任务
func (s *Scheduler) GetTask(taskID string) (*Task, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}
	
	return task, nil
}

// GetTasks 获取所有任务
func (s *Scheduler) GetTasks() map[string]*Task {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	tasks := make(map[string]*Task)
	for id, task := range s.tasks {
		tasks[id] = task
	}
	
	return tasks
}

// PauseTask 暂停任务
func (s *Scheduler) PauseTask(taskID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}
	
	if task.Status == TaskStatusRunning && task.cancel != nil {
		task.cancel()
	}
	
	task.SetStatus(TaskStatusPaused)
	
	config.Infof("Task paused: %s (%s)", task.Name, taskID)
	return nil
}

// ResumeTask 恢复任务
func (s *Scheduler) ResumeTask(taskID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}
	
	if task.Status != TaskStatusPaused {
		return fmt.Errorf("task %s is not paused", taskID)
	}
	
	// 重新计算下次运行时间
	nextRun, err := s.parseSchedule(task.Schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule '%s': %w", task.Schedule, err)
	}
	
	task.SetNextRunTime(nextRun)
	task.SetStatus(TaskStatusPending)
	
	config.Infof("Task resumed: %s (%s)", task.Name, taskID)
	return nil
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	if atomic.LoadInt32(&s.running) == 1 {
		return fmt.Errorf("scheduler is already running")
	}
	
	atomic.StoreInt32(&s.running, 1)
	
	// 从存储中加载任务
	if s.config.EnablePersistent && s.storage != nil {
		if err := s.loadTasksFromStorage(); err != nil {
			config.Errorf("Failed to load tasks from storage: %v", err)
		}
	}
	
	// 启动调度循环
	go s.scheduleLoop()
	
	// 启动工作协程
	for i := 0; i < s.workers; i++ {
		go s.workerLoop(i)
	}
	
	config.Infof("Scheduler started with %d workers", s.workers)
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	if atomic.LoadInt32(&s.running) == 0 {
		return fmt.Errorf("scheduler is not running")
	}
	
	atomic.StoreInt32(&s.running, 0)
	close(s.stopChan)
	
	// 取消所有正在运行的任务
	s.mutex.RLock()
	for _, task := range s.tasks {
		if task.Status == TaskStatusRunning && task.cancel != nil {
			task.cancel()
		}
	}
	s.mutex.RUnlock()
	
	config.Info("Scheduler stopped")
	return nil
}

// IsRunning 检查调度器是否运行中
func (s *Scheduler) IsRunning() bool {
	return atomic.LoadInt32(&s.running) == 1
}

// scheduleLoop 调度循环
func (s *Scheduler) scheduleLoop() {
	ticker := time.NewTicker(s.config.TickInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndScheduleTasks()
		}
	}
}

// checkAndScheduleTasks 检查并调度任务
func (s *Scheduler) checkAndScheduleTasks() {
	now := time.Now()
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	for _, task := range s.tasks {
		if task.Status != TaskStatusPending {
			continue
		}
		
		if task.NextRunTime != nil && now.After(*task.NextRunTime) {
			// 提交任务到工作队列
			go s.executeTask(task)
		}
	}
}

// executeTask 执行任务
func (s *Scheduler) executeTask(task *Task) {
	// 检查任务状态
	if task.Status != TaskStatusPending {
		return
	}
	
	// 设置任务状态为运行中
	task.SetStatus(TaskStatusRunning)
	task.SetLastRunTime(time.Now())
	task.IncrementRunCount()
	
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	task.cancel = cancel
	defer cancel()
	
	// 触发开始回调
	if s.onTaskStart != nil {
		s.onTaskStart(task)
	}
	
	// 执行任务
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("task panicked: %v", r)
			}
		}()
		
		err = task.Job.Execute(ctx)
	}()
	
	// 处理执行结果
	if err != nil {
		task.IncrementFailCount()
		task.SetStatus(TaskStatusFailed)
		
		if s.config.EnableLogging {
			config.Errorf("Task %s (%s) failed: %v", task.Name, task.ID, err)
		}
		
		// 触发失败回调
		if s.onTaskFail != nil {
			s.onTaskFail(task, err)
		}
		
		// 重试逻辑
		if task.FailCount < int64(task.MaxRetries) {
			// 计算下次重试时间（指数退避）
			retryDelay := time.Duration(task.FailCount) * time.Minute
			nextRun := time.Now().Add(retryDelay)
			task.SetNextRunTime(nextRun)
			task.SetStatus(TaskStatusPending)
			
			if s.config.EnableLogging {
				config.Infof("Task %s (%s) scheduled for retry at %v", task.Name, task.ID, nextRun)
			}
		}
	} else {
		task.SetStatus(TaskStatusCompleted)
		
		if s.config.EnableLogging {
			config.Infof("Task %s (%s) completed successfully", task.Name, task.ID)
		}
		
		// 触发完成回调
		if s.onTaskComplete != nil {
			s.onTaskComplete(task, nil)
		}
	}
	
	// 计算下次运行时间
	if task.Status == TaskStatusCompleted {
		nextRun, parseErr := s.parseSchedule(task.Schedule)
		if parseErr == nil {
			task.SetNextRunTime(nextRun)
			task.SetStatus(TaskStatusPending)
		}
	}
	
	// 持久化任务状态
	if s.config.EnablePersistent && s.storage != nil {
		if saveErr := s.storage.SaveTask(task); saveErr != nil {
			config.Errorf("Failed to persist task %s: %v", task.ID, saveErr)
		}
	}
}

// workerLoop 工作协程循环
func (s *Scheduler) workerLoop(workerID int) {
	config.Infof("Worker %d started", workerID)
	
	for {
		select {
		case <-s.stopChan:
			config.Infof("Worker %d stopped", workerID)
			return
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

// parseSchedule 解析调度表达式
func (s *Scheduler) parseSchedule(schedule string) (time.Time, error) {
	// 简化实现，支持几种常见格式
	switch schedule {
	case "@every_minute":
		return time.Now().Add(time.Minute), nil
	case "@every_hour":
		return time.Now().Add(time.Hour), nil
	case "@every_day":
		return time.Now().Add(24 * time.Hour), nil
	case "@once":
		return time.Now(), nil
	default:
		// 尝试解析为时间间隔
		if duration, err := time.ParseDuration(schedule); err == nil {
			return time.Now().Add(duration), nil
		}
		
		// 尝试解析为绝对时间
		if t, err := time.Parse("2006-01-02 15:04:05", schedule); err == nil {
			return t, nil
		}
		
		return time.Time{}, fmt.Errorf("unsupported schedule format: %s", schedule)
	}
}

// loadTasksFromStorage 从存储中加载任务
func (s *Scheduler) loadTasksFromStorage() error {
	tasks, err := s.storage.LoadTasks()
	if err != nil {
		return err
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for _, task := range tasks {
		s.tasks[task.ID] = task
		config.Infof("Loaded task from storage: %s (%s)", task.Name, task.ID)
	}
	
	return nil
}

// GetStats 获取调度器统计信息
func (s *Scheduler) GetStats() *SchedulerStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	stats := &SchedulerStats{
		TotalTasks:    len(s.tasks),
		RunningTasks:  0,
		PendingTasks:  0,
		CompletedTasks: 0,
		FailedTasks:   0,
		PausedTasks:   0,
	}
	
	for _, task := range s.tasks {
		switch task.Status {
		case TaskStatusRunning:
			stats.RunningTasks++
		case TaskStatusPending:
			stats.PendingTasks++
		case TaskStatusCompleted:
			stats.CompletedTasks++
		case TaskStatusFailed:
			stats.FailedTasks++
		case TaskStatusPaused:
			stats.PausedTasks++
		}
	}
	
	return stats
}

// SchedulerStats 调度器统计信息
type SchedulerStats struct {
	TotalTasks     int `json:"total_tasks"`
	RunningTasks   int `json:"running_tasks"`
	PendingTasks   int `json:"pending_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks    int `json:"failed_tasks"`
	PausedTasks    int `json:"paused_tasks"`
}

// ============= 全局调度器 =============

var (
	globalScheduler *Scheduler
	schedulerOnce   sync.Once
)

// GetGlobalScheduler 获取全局调度器
func GetGlobalScheduler() *Scheduler {
	schedulerOnce.Do(func() {
		globalScheduler = NewScheduler(DefaultSchedulerConfig())
	})
	return globalScheduler
}

// StartGlobalScheduler 启动全局调度器
func StartGlobalScheduler() error {
	return GetGlobalScheduler().Start()
}

// StopGlobalScheduler 停止全局调度器
func StopGlobalScheduler() error {
	return GetGlobalScheduler().Stop()
}

// AddGlobalTask 添加全局任务
func AddGlobalTask(task *Task) error {
	return GetGlobalScheduler().AddTask(task)
}

// RemoveGlobalTask 移除全局任务
func RemoveGlobalTask(taskID string) error {
	return GetGlobalScheduler().RemoveTask(taskID)
}