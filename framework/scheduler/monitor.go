package scheduler

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ExecutionMonitor 执行监控器
type ExecutionMonitor struct {
	metrics     *MonitorMetrics
	running     int32
	stopChan    chan struct{}
	alertRules  []AlertRule
	subscribers []MetricsSubscriber

	// 系统监控
	systemMonitor *SystemMonitor

	mutex sync.RWMutex
}

// MonitorMetrics 监控指标
type MonitorMetrics struct {
	// 执行统计
	TotalExecutions      int64 `json:"total_executions"`
	SuccessfulExecutions int64 `json:"successful_executions"`
	FailedExecutions     int64 `json:"failed_executions"`
	CanceledExecutions   int64 `json:"canceled_executions"`

	// 性能统计
	TotalExecutionTime   time.Duration `json:"total_execution_time"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	MinExecutionTime     time.Duration `json:"min_execution_time"`

	// 当前状态
	CurrentlyRunning int32 `json:"currently_running"`
	QueueSize        int32 `json:"queue_size"`

	// 错误统计
	ErrorRate    float64 `json:"error_rate"`
	PanicCount   int64   `json:"panic_count"`
	TimeoutCount int64   `json:"timeout_count"`

	// 时间窗口统计
	LastHour *TimeWindowMetrics `json:"last_hour"`
	LastDay  *TimeWindowMetrics `json:"last_day"`

	// 任务级别统计
	TaskMetrics map[string]*TaskMetrics `json:"task_metrics"`

	// 更新时间
	LastUpdated time.Time `json:"last_updated"`

	mutex sync.RWMutex
}

// TimeWindowMetrics 时间窗口指标
type TimeWindowMetrics struct {
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Executions      int64         `json:"executions"`
	Successes       int64         `json:"successes"`
	Failures        int64         `json:"failures"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	Throughput      float64       `json:"throughput"` // 每秒执行数
}

// TaskMetrics 任务级别指标
type TaskMetrics struct {
	TaskID           string        `json:"task_id"`
	TaskName         string        `json:"task_name"`
	TotalExecutions  int64         `json:"total_executions"`
	SuccessfulRuns   int64         `json:"successful_runs"`
	FailedRuns       int64         `json:"failed_runs"`
	AverageTime      time.Duration `json:"average_time"`
	LastExecution    time.Time     `json:"last_execution"`
	NextExecution    time.Time     `json:"next_execution"`
	ConsecutiveFails int64         `json:"consecutive_fails"`
	SuccessRate      float64       `json:"success_rate"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name      string
	Condition func(*MonitorMetrics) bool
	Message   string
	Severity  AlertSeverity
	Cooldown  time.Duration
	LastFired time.Time
	Enabled   bool
}

// AlertSeverity 告警级别
type AlertSeverity int

const (
	AlertSeverityInfo AlertSeverity = iota
	AlertSeverityWarning
	AlertSeverityError
	AlertSeverityCritical
)

// String 告警级别字符串
func (as AlertSeverity) String() string {
	switch as {
	case AlertSeverityInfo:
		return "INFO"
	case AlertSeverityWarning:
		return "WARNING"
	case AlertSeverityError:
		return "ERROR"
	case AlertSeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// MetricsSubscriber 指标订阅者
type MetricsSubscriber interface {
	OnMetricsUpdate(metrics *MonitorMetrics)
	OnAlert(alert *Alert)
}

// Alert 告警信息
type Alert struct {
	RuleName  string          `json:"rule_name"`
	Message   string          `json:"message"`
	Severity  AlertSeverity   `json:"severity"`
	Timestamp time.Time       `json:"timestamp"`
	Metrics   *MonitorMetrics `json:"metrics"`
}

// NewExecutionMonitor 创建执行监控器
func NewExecutionMonitor() *ExecutionMonitor {
	return &ExecutionMonitor{
		metrics: &MonitorMetrics{
			TaskMetrics: make(map[string]*TaskMetrics),
			LastHour:    &TimeWindowMetrics{StartTime: time.Now().Add(-time.Hour)},
			LastDay:     &TimeWindowMetrics{StartTime: time.Now().Add(-24 * time.Hour)},
			LastUpdated: time.Now(),
		},
		stopChan:      make(chan struct{}),
		alertRules:    make([]AlertRule, 0),
		subscribers:   make([]MetricsSubscriber, 0),
		systemMonitor: NewSystemMonitor(),
	}
}

// Start 启动监控器
func (em *ExecutionMonitor) Start() error {
	if atomic.LoadInt32(&em.running) == 1 {
		return fmt.Errorf("execution monitor is already running")
	}

	atomic.StoreInt32(&em.running, 1)

	// 启动系统监控
	em.systemMonitor.Start()

	// 启动监控循环
	go em.monitorLoop()

	config.Info("Execution monitor started")
	return nil
}

// Stop 停止监控器
func (em *ExecutionMonitor) Stop() error {
	if atomic.LoadInt32(&em.running) == 0 {
		return fmt.Errorf("execution monitor is not running")
	}

	atomic.StoreInt32(&em.running, 0)
	close(em.stopChan)

	// 停止系统监控
	em.systemMonitor.Stop()

	config.Info("Execution monitor stopped")
	return nil
}

// IsRunning 检查是否运行中
func (em *ExecutionMonitor) IsRunning() bool {
	return atomic.LoadInt32(&em.running) == 1
}

// RecordExecutionStart 记录执行开始
func (em *ExecutionMonitor) RecordExecutionStart(execution *TaskExecution) {
	atomic.AddInt32(&em.metrics.CurrentlyRunning, 1)
	atomic.AddInt64(&em.metrics.TotalExecutions, 1)

	// 更新任务级别指标
	em.updateTaskMetricsStart(execution)

	// 通知订阅者
	em.notifySubscribers()
}

// RecordExecutionEnd 记录执行结束
func (em *ExecutionMonitor) RecordExecutionEnd(execution *TaskExecution, err error) {
	atomic.AddInt32(&em.metrics.CurrentlyRunning, -1)

	// 更新执行统计
	if err != nil {
		atomic.AddInt64(&em.metrics.FailedExecutions, 1)
	} else {
		atomic.AddInt64(&em.metrics.SuccessfulExecutions, 1)
	}

	// 更新性能统计
	em.updatePerformanceMetrics(execution)

	// 更新任务级别指标
	em.updateTaskMetricsEnd(execution, err)

	// 更新时间窗口指标
	em.updateTimeWindowMetrics(execution, err)

	// 检查告警规则
	em.checkAlertRules()

	// 通知订阅者
	em.notifySubscribers()
}

// RecordPanic 记录panic
func (em *ExecutionMonitor) RecordPanic(execution *TaskExecution, panicValue any) {
	atomic.AddInt64(&em.metrics.PanicCount, 1)
	config.Errorf("Task %s panicked: %v", execution.Task.ID, panicValue)
}

// RecordTimeout 记录超时
func (em *ExecutionMonitor) RecordTimeout(execution *TaskExecution) {
	atomic.AddInt64(&em.metrics.TimeoutCount, 1)
	config.Warnf("Task %s timed out after %v", execution.Task.ID, execution.Duration)
}

// updatePerformanceMetrics 更新性能指标
func (em *ExecutionMonitor) updatePerformanceMetrics(execution *TaskExecution) {
	em.metrics.mutex.Lock()
	defer em.metrics.mutex.Unlock()

	duration := execution.Duration

	// 更新总执行时间
	em.metrics.TotalExecutionTime += duration

	// 更新最大执行时间
	if duration > em.metrics.MaxExecutionTime {
		em.metrics.MaxExecutionTime = duration
	}

	// 更新最小执行时间
	if em.metrics.MinExecutionTime == 0 || duration < em.metrics.MinExecutionTime {
		em.metrics.MinExecutionTime = duration
	}

	// 更新平均执行时间
	totalExecs := atomic.LoadInt64(&em.metrics.TotalExecutions)
	if totalExecs > 0 {
		em.metrics.AverageExecutionTime = em.metrics.TotalExecutionTime / time.Duration(totalExecs)
	}

	// 更新错误率
	failedExecs := atomic.LoadInt64(&em.metrics.FailedExecutions)
	if totalExecs > 0 {
		em.metrics.ErrorRate = float64(failedExecs) / float64(totalExecs) * 100
	}

	em.metrics.LastUpdated = time.Now()
}

// updateTaskMetricsStart 更新任务开始指标
func (em *ExecutionMonitor) updateTaskMetricsStart(execution *TaskExecution) {
	em.metrics.mutex.Lock()
	defer em.metrics.mutex.Unlock()

	taskID := execution.Task.ID
	taskMetrics, exists := em.metrics.TaskMetrics[taskID]
	if !exists {
		taskMetrics = &TaskMetrics{
			TaskID:   taskID,
			TaskName: execution.Task.Name,
		}
		em.metrics.TaskMetrics[taskID] = taskMetrics
	}

	taskMetrics.TotalExecutions++
	taskMetrics.LastExecution = execution.StartTime

	// 更新下次执行时间
	if execution.Task.NextRunTime != nil {
		taskMetrics.NextExecution = *execution.Task.NextRunTime
	}
}

// updateTaskMetricsEnd 更新任务结束指标
func (em *ExecutionMonitor) updateTaskMetricsEnd(execution *TaskExecution, err error) {
	em.metrics.mutex.Lock()
	defer em.metrics.mutex.Unlock()

	taskID := execution.Task.ID
	taskMetrics := em.metrics.TaskMetrics[taskID]
	if taskMetrics == nil {
		return
	}

	if err != nil {
		taskMetrics.FailedRuns++
		taskMetrics.ConsecutiveFails++
	} else {
		taskMetrics.SuccessfulRuns++
		taskMetrics.ConsecutiveFails = 0
	}

	// 更新成功率
	if taskMetrics.TotalExecutions > 0 {
		taskMetrics.SuccessRate = float64(taskMetrics.SuccessfulRuns) / float64(taskMetrics.TotalExecutions) * 100
	}

	// 更新平均执行时间
	if taskMetrics.TotalExecutions > 0 {
		// 简化计算，实际应该使用滑动平均
		taskMetrics.AverageTime = (taskMetrics.AverageTime*time.Duration(taskMetrics.TotalExecutions-1) + execution.Duration) / time.Duration(taskMetrics.TotalExecutions)
	}
}

// updateTimeWindowMetrics 更新时间窗口指标
func (em *ExecutionMonitor) updateTimeWindowMetrics(execution *TaskExecution, err error) {
	em.metrics.mutex.Lock()
	defer em.metrics.mutex.Unlock()

	now := time.Now()

	// 更新小时窗口
	if now.Sub(em.metrics.LastHour.StartTime) >= time.Hour {
		em.metrics.LastHour = &TimeWindowMetrics{
			StartTime: now.Add(-time.Hour),
			EndTime:   now,
		}
	}
	em.updateWindowMetrics(em.metrics.LastHour, execution, err)

	// 更新日窗口
	if now.Sub(em.metrics.LastDay.StartTime) >= 24*time.Hour {
		em.metrics.LastDay = &TimeWindowMetrics{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		}
	}
	em.updateWindowMetrics(em.metrics.LastDay, execution, err)
}

// updateWindowMetrics 更新窗口指标
func (em *ExecutionMonitor) updateWindowMetrics(window *TimeWindowMetrics, execution *TaskExecution, err error) {
	window.Executions++
	window.TotalDuration += execution.Duration
	window.EndTime = time.Now()

	if err != nil {
		window.Failures++
	} else {
		window.Successes++
	}

	// 更新平均时间
	if window.Executions > 0 {
		window.AverageDuration = window.TotalDuration / time.Duration(window.Executions)
	}

	// 更新吞吐量
	duration := window.EndTime.Sub(window.StartTime)
	if duration > 0 {
		window.Throughput = float64(window.Executions) / duration.Seconds()
	}
}

// AddAlertRule 添加告警规则
func (em *ExecutionMonitor) AddAlertRule(rule AlertRule) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	em.alertRules = append(em.alertRules, rule)
}

// Subscribe 订阅指标更新
func (em *ExecutionMonitor) Subscribe(subscriber MetricsSubscriber) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	em.subscribers = append(em.subscribers, subscriber)
}

// checkAlertRules 检查告警规则
func (em *ExecutionMonitor) checkAlertRules() {
	em.mutex.RLock()
	rules := make([]AlertRule, len(em.alertRules))
	copy(rules, em.alertRules)
	em.mutex.RUnlock()

	now := time.Now()

	for i, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// 检查冷却期
		if now.Sub(rule.LastFired) < rule.Cooldown {
			continue
		}

		// 检查条件
		if rule.Condition(em.metrics) {
			alert := &Alert{
				RuleName:  rule.Name,
				Message:   rule.Message,
				Severity:  rule.Severity,
				Timestamp: now,
				Metrics:   em.GetMetrics(),
			}

			// 更新最后触发时间
			em.mutex.Lock()
			em.alertRules[i].LastFired = now
			em.mutex.Unlock()

			// 通知订阅者
			em.notifyAlert(alert)

			// 记录告警日志
			em.logAlert(alert)
		}
	}
}

// notifySubscribers 通知订阅者
func (em *ExecutionMonitor) notifySubscribers() {
	em.mutex.RLock()
	subscribers := make([]MetricsSubscriber, len(em.subscribers))
	copy(subscribers, em.subscribers)
	em.mutex.RUnlock()

	metrics := em.GetMetrics()

	for _, subscriber := range subscribers {
		go func(sub MetricsSubscriber) {
			defer func() {
				if r := recover(); r != nil {
					config.Errorf("Metrics subscriber panicked: %v", r)
				}
			}()
			sub.OnMetricsUpdate(metrics)
		}(subscriber)
	}
}

// notifyAlert 通知告警
func (em *ExecutionMonitor) notifyAlert(alert *Alert) {
	em.mutex.RLock()
	subscribers := make([]MetricsSubscriber, len(em.subscribers))
	copy(subscribers, em.subscribers)
	em.mutex.RUnlock()

	for _, subscriber := range subscribers {
		go func(sub MetricsSubscriber) {
			defer func() {
				if r := recover(); r != nil {
					config.Errorf("Alert subscriber panicked: %v", r)
				}
			}()
			sub.OnAlert(alert)
		}(subscriber)
	}
}

// logAlert 记录告警日志
func (em *ExecutionMonitor) logAlert(alert *Alert) {
	switch alert.Severity {
	case AlertSeverityInfo:
		config.Infof("[ALERT:%s] %s", alert.RuleName, alert.Message)
	case AlertSeverityWarning:
		config.Warnf("[ALERT:%s] %s", alert.RuleName, alert.Message)
	case AlertSeverityError, AlertSeverityCritical:
		config.Errorf("[ALERT:%s] %s", alert.RuleName, alert.Message)
	}
}

// monitorLoop 监控循环
func (em *ExecutionMonitor) monitorLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-em.stopChan:
			return
		case <-ticker.C:
			em.checkAlertRules()
			em.notifySubscribers()
		}
	}
}

// GetMetrics 获取当前指标
func (em *ExecutionMonitor) GetMetrics() *MonitorMetrics {
	em.metrics.mutex.RLock()
	defer em.metrics.mutex.RUnlock()

	// 深拷贝指标
	metrics := &MonitorMetrics{
		TotalExecutions:      atomic.LoadInt64(&em.metrics.TotalExecutions),
		SuccessfulExecutions: atomic.LoadInt64(&em.metrics.SuccessfulExecutions),
		FailedExecutions:     atomic.LoadInt64(&em.metrics.FailedExecutions),
		CanceledExecutions:   atomic.LoadInt64(&em.metrics.CanceledExecutions),
		TotalExecutionTime:   em.metrics.TotalExecutionTime,
		AverageExecutionTime: em.metrics.AverageExecutionTime,
		MaxExecutionTime:     em.metrics.MaxExecutionTime,
		MinExecutionTime:     em.metrics.MinExecutionTime,
		CurrentlyRunning:     atomic.LoadInt32(&em.metrics.CurrentlyRunning),
		QueueSize:            atomic.LoadInt32(&em.metrics.QueueSize),
		ErrorRate:            em.metrics.ErrorRate,
		PanicCount:           atomic.LoadInt64(&em.metrics.PanicCount),
		TimeoutCount:         atomic.LoadInt64(&em.metrics.TimeoutCount),
		LastUpdated:          em.metrics.LastUpdated,
		TaskMetrics:          make(map[string]*TaskMetrics),
	}

	// 拷贝任务指标
	for k, v := range em.metrics.TaskMetrics {
		taskMetrics := *v
		metrics.TaskMetrics[k] = &taskMetrics
	}

	// 拷贝时间窗口指标
	if em.metrics.LastHour != nil {
		hour := *em.metrics.LastHour
		metrics.LastHour = &hour
	}
	if em.metrics.LastDay != nil {
		day := *em.metrics.LastDay
		metrics.LastDay = &day
	}

	return metrics
}

// GetSystemLoad 获取系统负载
func (em *ExecutionMonitor) GetSystemLoad() float64 {
	return em.systemMonitor.GetCPUUsage()
}

// GetAvailableMemory 获取可用内存
func (em *ExecutionMonitor) GetAvailableMemory() int64 {
	return em.systemMonitor.GetAvailableMemory()
}

// ============= 系统监控器 =============

// SystemMonitor 系统监控器
type SystemMonitor struct {
	cpuUsage        float64
	memoryUsage     int64
	availableMemory int64
	goroutineCount  int

	running  int32
	stopChan chan struct{}
	mutex    sync.RWMutex
}

// NewSystemMonitor 创建系统监控器
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		stopChan: make(chan struct{}),
	}
}

// Start 启动系统监控
func (sm *SystemMonitor) Start() {
	if atomic.LoadInt32(&sm.running) == 1 {
		return
	}

	atomic.StoreInt32(&sm.running, 1)
	go sm.monitorLoop()
}

// Stop 停止系统监控
func (sm *SystemMonitor) Stop() {
	if atomic.LoadInt32(&sm.running) == 0 {
		return
	}

	atomic.StoreInt32(&sm.running, 0)
	close(sm.stopChan)
}

// monitorLoop 系统监控循环
func (sm *SystemMonitor) monitorLoop() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.updateSystemMetrics()
		}
	}
}

// updateSystemMetrics 更新系统指标
func (sm *SystemMonitor) updateSystemMetrics() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 更新Goroutine数量
	sm.goroutineCount = runtime.NumGoroutine()

	// 更新内存使用（简化实现）
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	sm.memoryUsage = int64(m.Alloc)
	sm.availableMemory = int64(m.Sys - m.Alloc)

	// CPU使用率需要更复杂的实现，这里简化处理
	sm.cpuUsage = float64(sm.goroutineCount) / float64(runtime.NumCPU()) * 10
	if sm.cpuUsage > 100 {
		sm.cpuUsage = 100
	}
}

// GetCPUUsage 获取CPU使用率
func (sm *SystemMonitor) GetCPUUsage() float64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.cpuUsage
}

// GetMemoryUsage 获取内存使用
func (sm *SystemMonitor) GetMemoryUsage() int64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.memoryUsage
}

// GetAvailableMemory 获取可用内存
func (sm *SystemMonitor) GetAvailableMemory() int64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.availableMemory
}

// GetGoroutineCount 获取Goroutine数量
func (sm *SystemMonitor) GetGoroutineCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.goroutineCount
}

// ============= 预定义告警规则 =============

// DefaultAlertRules 默认告警规则
func DefaultAlertRules() []AlertRule {
	return []AlertRule{
		{
			Name: "HighErrorRate",
			Condition: func(metrics *MonitorMetrics) bool {
				return metrics.ErrorRate > 50.0 // 错误率超过50%
			},
			Message:  "任务执行错误率过高",
			Severity: AlertSeverityWarning,
			Cooldown: time.Minute * 5,
			Enabled:  true,
		},
		{
			Name: "HighPanicCount",
			Condition: func(metrics *MonitorMetrics) bool {
				return metrics.PanicCount > 10 // Panic次数超过10次
			},
			Message:  "任务执行Panic次数过多",
			Severity: AlertSeverityError,
			Cooldown: time.Minute * 10,
			Enabled:  true,
		},
		{
			Name: "LongExecutionTime",
			Condition: func(metrics *MonitorMetrics) bool {
				return metrics.AverageExecutionTime > time.Minute*10 // 平均执行时间超过10分钟
			},
			Message:  "任务平均执行时间过长",
			Severity: AlertSeverityWarning,
			Cooldown: time.Minute * 15,
			Enabled:  true,
		},
		{
			Name: "NoExecutions",
			Condition: func(metrics *MonitorMetrics) bool {
				return metrics.LastHour != nil && metrics.LastHour.Executions == 0 // 最近一小时无执行
			},
			Message:  "最近一小时内无任务执行",
			Severity: AlertSeverityWarning,
			Cooldown: time.Hour,
			Enabled:  true,
		},
	}
}

// ============= 指标订阅者实现示例 =============

// LoggingSubscriber 日志订阅者
type LoggingSubscriber struct{}

// OnMetricsUpdate 指标更新回调
func (ls *LoggingSubscriber) OnMetricsUpdate(metrics *MonitorMetrics) {
	config.Debugf("Metrics updated: Total=%d, Success=%d, Failed=%d, Running=%d",
		metrics.TotalExecutions, metrics.SuccessfulExecutions,
		metrics.FailedExecutions, metrics.CurrentlyRunning)
}

// OnAlert 告警回调
func (ls *LoggingSubscriber) OnAlert(alert *Alert) {
	config.Warnf("Alert triggered: %s - %s [%s]",
		alert.RuleName, alert.Message, alert.Severity.String())
}
