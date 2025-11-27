# 任务调度

YYHertz 框架提供了功能强大的任务调度系统，支持定时任务、延迟任务、重复任务等多种调度模式，适用于数据清理、报告生成、消息推送等场景。

## 概述

任务调度是现代 Web 应用程序的重要组件。YYHertz 的调度系统基于 Go 的并发特性，提供：

- Cron 表达式定时任务
- 延迟任务执行
- 重复任务调度
- 任务队列管理
- 任务状态监控
- 分布式任务调度
- 任务失败重试

## 基本使用

### 初始化调度器

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/scheduler"
)

func main() {
    app := mvc.HertzApp
    
    // 创建调度器
    sched := scheduler.New(scheduler.Config{
        MaxWorkers: 10,
        QueueSize:  1000,
    })
    
    // 启动调度器
    sched.Start()
    
    // 注册全局调度器
    scheduler.SetGlobalScheduler(sched)
    
    // 应用关闭时停止调度器
    defer sched.Stop()
    
    app.Run()
}
```

### 定义和注册任务

```go
package tasks

import (
    "context"
    "fmt"
    "time"
    "github.com/zsy619/yyhertz/framework/mvc/scheduler"
)

// 定义任务
type EmailTask struct {
    To      string
    Subject string
    Body    string
}

// 实现 Task 接口
func (t *EmailTask) Execute(ctx context.Context) error {
    // 发送邮件逻辑
    fmt.Printf("Sending email to %s: %s\n", t.To, t.Subject)
    
    // 模拟发送时间
    time.Sleep(2 * time.Second)
    
    return nil
}

func (t *EmailTask) GetID() string {
    return fmt.Sprintf("email_%s_%d", t.To, time.Now().Unix())
}

func (t *EmailTask) GetRetryCount() int {
    return 3 // 最多重试3次
}

// 数据清理任务
type CleanupTask struct {
    TableName string
    Days      int
}

func (t *CleanupTask) Execute(ctx context.Context) error {
    fmt.Printf("Cleaning up %s table, removing data older than %d days\n", 
        t.TableName, t.Days)
    
    // 数据库清理逻辑
    // db.Exec("DELETE FROM ? WHERE created_at < ?", t.TableName, cutoffDate)
    
    return nil
}

func (t *CleanupTask) GetID() string {
    return fmt.Sprintf("cleanup_%s", t.TableName)
}

func (t *CleanupTask) GetRetryCount() int {
    return 1
}

// 报告生成任务
type ReportTask struct {
    ReportType string
    UserID     int
    Format     string
}

func (t *ReportTask) Execute(ctx context.Context) error {
    fmt.Printf("Generating %s report for user %d in %s format\n", 
        t.ReportType, t.UserID, t.Format)
    
    // 报告生成逻辑
    // 1. 查询数据
    // 2. 生成报告
    // 3. 发送通知
    
    return nil
}

func (t *ReportTask) GetID() string {
    return fmt.Sprintf("report_%s_%d_%d", t.ReportType, t.UserID, time.Now().Unix())
}

func (t *ReportTask) GetRetryCount() int {
    return 2
}
```

## 任务调度模式

### Cron 定时任务

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc/scheduler"
)

func setupCronJobs() {
    sched := scheduler.GetGlobalScheduler()
    
    // 每天凌晨 2 点执行数据清理
    sched.AddCronJob("0 2 * * *", &CleanupTask{
        TableName: "logs",
        Days:      30,
    })
    
    // 每周一上午 9 点生成周报
    sched.AddCronJob("0 9 * * 1", &ReportTask{
        ReportType: "weekly",
        UserID:     0, // 系统报告
        Format:     "pdf",
    })
    
    // 每小时检查任务状态
    sched.AddCronJob("0 * * * *", &HealthCheckTask{})
    
    // 每分钟发送待发邮件
    sched.AddCronJob("* * * * *", &EmailQueueTask{})
}

// Cron 表达式说明
/*
格式: 分 时 日 月 周
* * * * *  每分钟执行
0 * * * *  每小时执行
0 9 * * *  每天 9 点执行
0 9 * * 1  每周一 9 点执行
0 2 1 * *  每月 1 号 2 点执行
0 0 1 1 *  每年 1 月 1 号执行
*/
```

### 延迟任务

```go
func scheduleDelayedTasks() {
    sched := scheduler.GetGlobalScheduler()
    
    // 5 分钟后发送邮件
    sched.AddDelayedJob(5*time.Minute, &EmailTask{
        To:      "user@example.com",
        Subject: "Welcome!",
        Body:    "Welcome to our platform!",
    })
    
    // 1 小时后执行清理
    sched.AddDelayedJob(1*time.Hour, &CleanupTask{
        TableName: "temp_files",
        Days:      1,
    })
    
    // 24 小时后生成报告
    sched.AddDelayedJob(24*time.Hour, &ReportTask{
        ReportType: "daily",
        UserID:     123,
        Format:     "excel",
    })
}
```

### 重复任务

```go
func scheduleRecurringTasks() {
    sched := scheduler.GetGlobalScheduler()
    
    // 每 30 秒执行一次健康检查
    jobID := sched.AddRecurringJob(30*time.Second, &HealthCheckTask{
        ServiceName: "database",
    })
    
    // 每 5 分钟处理消息队列
    sched.AddRecurringJob(5*time.Minute, &MessageProcessTask{
        QueueName: "notifications",
        BatchSize: 100,
    })
    
    // 稍后可以取消任务
    go func() {
        time.Sleep(10 * time.Minute)
        sched.CancelJob(jobID)
    }()
}
```

### 即时任务

```go
func scheduleImmediateTasks(c *gin.Context) {
    sched := scheduler.GetGlobalScheduler()
    
    // 立即执行任务
    err := sched.AddImmediateJob(&EmailTask{
        To:      "admin@example.com",
        Subject: "Urgent: System Alert",
        Body:    "System requires immediate attention",
    })
    
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to schedule task"})
        return
    }
    
    c.JSON(200, gin.H{"status": "Task scheduled"})
}
```

## 任务队列

### 队列配置

```go
type QueueConfig struct {
    Name        string        `json:"name"`
    MaxSize     int           `json:"max_size"`
    Workers     int           `json:"workers"`
    Timeout     time.Duration `json:"timeout"`
    RetryPolicy RetryPolicy   `json:"retry_policy"`
}

type RetryPolicy struct {
    MaxRetries int           `json:"max_retries"`
    Delay      time.Duration `json:"delay"`
    Backoff    string        `json:"backoff"` // "fixed", "exponential", "linear"
}

// 创建不同优先级的队列
func setupQueues() {
    sched := scheduler.GetGlobalScheduler()
    
    // 高优先级队列
    sched.CreateQueue("high_priority", QueueConfig{
        Name:    "high_priority",
        MaxSize: 100,
        Workers: 5,
        Timeout: 30 * time.Second,
        RetryPolicy: RetryPolicy{
            MaxRetries: 3,
            Delay:      5 * time.Second,
            Backoff:    "exponential",
        },
    })
    
    // 低优先级队列
    sched.CreateQueue("low_priority", QueueConfig{
        Name:    "low_priority",
        MaxSize: 1000,
        Workers: 2,
        Timeout: 5 * time.Minute,
        RetryPolicy: RetryPolicy{
            MaxRetries: 1,
            Delay:      10 * time.Second,
            Backoff:    "fixed",
        },
    })
}
```

### 使用不同队列

```go
func enqueueToSpecificQueue() {
    sched := scheduler.GetGlobalScheduler()
    
    // 高优先级任务
    sched.EnqueueToQueue("high_priority", &EmailTask{
        To:      "vip@example.com",
        Subject: "VIP Notification",
        Body:    "Important message for VIP user",
    })
    
    // 低优先级任务
    sched.EnqueueToQueue("low_priority", &ReportTask{
        ReportType: "monthly",
        UserID:     456,
        Format:     "csv",
    })
    
    // 默认队列
    sched.Enqueue(&CleanupTask{
        TableName: "cache",
        Days:      7,
    })
}
```

## 任务监控

### 任务状态

```go
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
    TaskStatusRetrying  TaskStatus = "retrying"
    TaskStatusCancelled TaskStatus = "cancelled"
)

type TaskInfo struct {
    ID          string        `json:"id"`
    Type        string        `json:"type"`
    Status      TaskStatus    `json:"status"`
    CreatedAt   time.Time     `json:"created_at"`
    StartedAt   *time.Time    `json:"started_at"`
    CompletedAt *time.Time    `json:"completed_at"`
    Duration    time.Duration `json:"duration"`
    Error       string        `json:"error,omitempty"`
    RetryCount  int           `json:"retry_count"`
    QueueName   string        `json:"queue_name"`
}
```

### 监控 API

```go
type SchedulerController struct {
    mvc.Controller
}

// 获取任务统计
func (c *SchedulerController) GetStats() {
    sched := scheduler.GetGlobalScheduler()
    stats := sched.GetStats()
    
    c.JSON(200, stats)
}

// 获取任务列表
func (c *SchedulerController) GetTasks() {
    sched := scheduler.GetGlobalScheduler()
    
    status := c.GetString("status")
    queue := c.GetString("queue")
    limit := c.GetInt("limit", 100)
    offset := c.GetInt("offset", 0)
    
    tasks := sched.GetTasks(scheduler.TaskFilter{
        Status: TaskStatus(status),
        Queue:  queue,
        Limit:  limit,
        Offset: offset,
    })
    
    c.JSON(200, gin.H{
        "tasks": tasks,
        "total": len(tasks),
    })
}

// 获取特定任务详情
func (c *SchedulerController) GetTask() {
    taskID := c.GetString("id")
    
    sched := scheduler.GetGlobalScheduler()
    task, exists := sched.GetTask(taskID)
    
    if !exists {
        c.JSON(404, gin.H{"error": "Task not found"})
        return
    }
    
    c.JSON(200, task)
}

// 取消任务
func (c *SchedulerController) CancelTask() {
    taskID := c.GetString("id")
    
    sched := scheduler.GetGlobalScheduler()
    err := sched.CancelJob(taskID)
    
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"status": "cancelled"})
}

// 重试失败任务
func (c *SchedulerController) RetryTask() {
    taskID := c.GetString("id")
    
    sched := scheduler.GetGlobalScheduler()
    err := sched.RetryTask(taskID)
    
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"status": "retrying"})
}
```

### 监控面板

```go
// 创建监控面板路由
func setupMonitoringRoutes(app *mvc.App) {
    schedulerController := &SchedulerController{}
    
    app.RouterPrefix("/admin/scheduler", schedulerController, "GetDashboard", "GET:/")
    app.RouterPrefix("/admin/scheduler", schedulerController, "GetStats", "GET:/stats")
    app.RouterPrefix("/admin/scheduler", schedulerController, "GetTasks", "GET:/tasks")
    app.RouterPrefix("/admin/scheduler", schedulerController, "GetTask", "GET:/tasks/:id")
    app.RouterPrefix("/admin/scheduler", schedulerController, "CancelTask", "DELETE:/tasks/:id")
    app.RouterPrefix("/admin/scheduler", schedulerController, "RetryTask", "POST:/tasks/:id/retry")
}

func (c *SchedulerController) GetDashboard() {
    sched := scheduler.GetGlobalScheduler()
    
    stats := sched.GetStats()
    recentTasks := sched.GetTasks(scheduler.TaskFilter{
        Limit: 10,
    })
    
    c.Data["Stats"] = stats
    c.Data["RecentTasks"] = recentTasks
    
    c.TplName = "admin/scheduler/dashboard.html"
}
```

## 分布式调度

### Redis 分布式锁

```go
package scheduler

import (
    "context"
    "time"
    "github.com/go-redis/redis/v8"
)

type DistributedScheduler struct {
    *Scheduler
    redis  *redis.Client
    nodeID string
}

func NewDistributedScheduler(config Config, redisClient *redis.Client) *DistributedScheduler {
    return &DistributedScheduler{
        Scheduler: New(config),
        redis:     redisClient,
        nodeID:    generateNodeID(),
    }
}

// 分布式任务执行
func (ds *DistributedScheduler) executeWithLock(task Task) error {
    lockKey := fmt.Sprintf("scheduler:lock:%s", task.GetID())
    lockValue := ds.nodeID
    lockTTL := 30 * time.Minute
    
    // 尝试获取分布式锁
    acquired, err := ds.acquireLock(lockKey, lockValue, lockTTL)
    if err != nil {
        return err
    }
    
    if !acquired {
        // 其他节点正在执行此任务
        return nil
    }
    
    defer ds.releaseLock(lockKey, lockValue)
    
    // 执行任务
    return task.Execute(context.Background())
}

func (ds *DistributedScheduler) acquireLock(key, value string, ttl time.Duration) (bool, error) {
    result, err := ds.redis.SetNX(context.Background(), key, value, ttl).Result()
    return result, err
}

func (ds *DistributedScheduler) releaseLock(key, value string) error {
    script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
    
    _, err := ds.redis.Eval(context.Background(), script, []string{key}, value).Result()
    return err
}
```

### 任务状态同步

```go
// 分布式任务状态管理
func (ds *DistributedScheduler) syncTaskStatus(taskID string, status TaskStatus) error {
    key := fmt.Sprintf("scheduler:task:%s", taskID)
    
    taskInfo := TaskInfo{
        ID:        taskID,
        Status:    status,
        UpdatedAt: time.Now(),
        NodeID:    ds.nodeID,
    }
    
    data, err := json.Marshal(taskInfo)
    if err != nil {
        return err
    }
    
    return ds.redis.Set(context.Background(), key, data, 24*time.Hour).Err()
}

// 获取分布式任务状态
func (ds *DistributedScheduler) getDistributedTaskStatus(taskID string) (*TaskInfo, error) {
    key := fmt.Sprintf("scheduler:task:%s", taskID)
    
    data, err := ds.redis.Get(context.Background(), key).Result()
    if err != nil {
        return nil, err
    }
    
    var taskInfo TaskInfo
    err = json.Unmarshal([]byte(data), &taskInfo)
    return &taskInfo, err
}
```

## 任务持久化

### 数据库持久化

```go
type TaskRepository struct {
    db *gorm.DB
}

type TaskRecord struct {
    ID          string    `gorm:"primaryKey"`
    Type        string    `gorm:"index"`
    Payload     string    `gorm:"type:text"`
    Status      string    `gorm:"index"`
    ScheduledAt time.Time `gorm:"index"`
    StartedAt   *time.Time
    CompletedAt *time.Time
    Error       string    `gorm:"type:text"`
    RetryCount  int       `gorm:"default:0"`
    MaxRetries  int       `gorm:"default:3"`
    QueueName   string    `gorm:"index"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (r *TaskRepository) SaveTask(task Task, scheduledAt time.Time) error {
    payload, err := json.Marshal(task)
    if err != nil {
        return err
    }
    
    record := TaskRecord{
        ID:          task.GetID(),
        Type:        reflect.TypeOf(task).String(),
        Payload:     string(payload),
        Status:      string(TaskStatusPending),
        ScheduledAt: scheduledAt,
        MaxRetries:  task.GetRetryCount(),
        CreatedAt:   time.Now(),
    }
    
    return r.db.Create(&record).Error
}

func (r *TaskRepository) UpdateTaskStatus(taskID string, status TaskStatus, err error) error {
    updates := map[string]interface{}{
        "status":     string(status),
        "updated_at": time.Now(),
    }
    
    if status == TaskStatusRunning {
        updates["started_at"] = time.Now()
    } else if status == TaskStatusCompleted || status == TaskStatusFailed {
        updates["completed_at"] = time.Now()
    }
    
    if err != nil {
        updates["error"] = err.Error()
        updates["retry_count"] = gorm.Expr("retry_count + 1")
    }
    
    return r.db.Model(&TaskRecord{}).Where("id = ?", taskID).Updates(updates).Error
}

// 恢复未完成的任务
func (r *TaskRepository) RecoverTasks() ([]TaskRecord, error) {
    var tasks []TaskRecord
    err := r.db.Where("status IN ?", []string{
        string(TaskStatusPending),
        string(TaskStatusRunning),
    }).Find(&tasks).Error
    
    return tasks, err
}
```

## 最佳实践

### 1. 任务设计原则

```go
// 好的任务设计
type GoodTask struct {
    // 任务参数应该是可序列化的
    UserID    int    `json:"user_id"`
    Email     string `json:"email"`
    Template  string `json:"template"`
    
    // 避免在任务中存储复杂对象
    // user *User // 不推荐
}

func (t *GoodTask) Execute(ctx context.Context) error {
    // 任务应该是幂等的
    if t.isAlreadyProcessed() {
        return nil
    }
    
    // 任务应该处理超时
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        return t.doWork()
    }
}

func (t *GoodTask) isAlreadyProcessed() bool {
    // 检查任务是否已经处理过
    return false
}
```

### 2. 错误处理和重试

```go
type RobustTask struct {
    ID         string
    maxRetries int
    retryDelay time.Duration
}

func (t *RobustTask) Execute(ctx context.Context) error {
    var lastErr error
    
    for attempt := 0; attempt <= t.maxRetries; attempt++ {
        if attempt > 0 {
            // 指数退避
            delay := time.Duration(attempt) * t.retryDelay
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
            }
        }
        
        err := t.doWork(ctx)
        if err == nil {
            return nil
        }
        
        // 判断是否应该重试
        if !t.shouldRetry(err) {
            return err
        }
        
        lastErr = err
        t.logRetryAttempt(attempt, err)
    }
    
    return fmt.Errorf("task failed after %d attempts: %w", t.maxRetries, lastErr)
}

func (t *RobustTask) shouldRetry(err error) bool {
    // 根据错误类型决定是否重试
    if errors.Is(err, ErrNetworkTimeout) {
        return true
    }
    if errors.Is(err, ErrValidationFailed) {
        return false // 验证错误不应该重试
    }
    return true
}
```

### 3. 任务性能优化

```go
// 批量处理任务
type BatchTask struct {
    Items     []Item
    BatchSize int
}

func (t *BatchTask) Execute(ctx context.Context) error {
    for i := 0; i < len(t.Items); i += t.BatchSize {
        end := i + t.BatchSize
        if end > len(t.Items) {
            end = len(t.Items)
        }
        
        batch := t.Items[i:end]
        if err := t.processBatch(ctx, batch); err != nil {
            return fmt.Errorf("failed to process batch %d-%d: %w", i, end, err)
        }
        
        // 检查上下文取消
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
    }
    
    return nil
}

// 并发处理
func (t *BatchTask) processBatch(ctx context.Context, items []Item) error {
    semaphore := make(chan struct{}, 10) // 限制并发数
    errors := make(chan error, len(items))
    
    for _, item := range items {
        go func(item Item) {
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            errors <- t.processItem(ctx, item)
        }(item)
    }
    
    // 收集错误
    var errs []error
    for i := 0; i < len(items); i++ {
        if err := <-errors; err != nil {
            errs = append(errs, err)
        }
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("batch processing failed: %v", errs)
    }
    
    return nil
}
```

### 4. 监控和告警

```go
// 任务监控
type TaskMonitor struct {
    prometheus.Counter
    duration prometheus.Histogram
    errors   prometheus.Counter
}

func (m *TaskMonitor) RecordExecution(task Task, duration time.Duration, err error) {
    labels := prometheus.Labels{
        "task_type": reflect.TypeOf(task).Name(),
        "queue":     task.GetQueueName(),
    }
    
    m.Counter.With(labels).Inc()
    m.duration.With(labels).Observe(duration.Seconds())
    
    if err != nil {
        errorLabels := prometheus.Labels{
            "task_type": reflect.TypeOf(task).Name(),
            "error":     err.Error(),
        }
        m.errors.With(errorLabels).Inc()
    }
}

// 告警规则
func setupAlerting() {
    // 任务失败率过高告警
    // 任务执行时间过长告警
    // 队列积压告警
}
```

YYHertz 的任务调度系统提供了完整的任务管理解决方案，从简单的定时任务到复杂的分布式调度，能够满足各种应用场景的需求，提高系统的自动化程度和运维效率。
