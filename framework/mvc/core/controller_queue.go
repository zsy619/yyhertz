package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 队列任务管理方法 =============

// QueueJob 队列任务结构
type QueueJob struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]any `json:"payload"`
	Priority    int                    `json:"priority"` // 1=高 5=中 10=低
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	Status      string                 `json:"status"` // pending, processing, completed, failed
	Error       string                 `json:"error,omitempty"`
	Queue       string                 `json:"queue"`
}

// JobHandler 任务处理器函数类型
type JobHandler func(*QueueJob) error

// jobHandlers 任务处理器注册表
var jobHandlers = make(map[string]JobHandler)

// queueStorage 队列存储（示例使用内存，实际应使用Redis等）
var queueStorage = make(map[string][]*QueueJob)

// ============= 任务分发方法 =============

// Dispatch 分发任务到队列
func (c *BaseController) Dispatch(jobType string, payload map[string]any, options ...JobOption) (*QueueJob, error) {
	job := &QueueJob{
		ID:          c.generateJobID(),
		Type:        jobType,
		Payload:     payload,
		Priority:    5, // 默认中等优先级
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
		Status:      "pending",
		Queue:       "default",
	}
	
	// 应用选项
	for _, option := range options {
		option(job)
	}
	
	// 验证任务
	if err := c.validateJob(job); err != nil {
		return nil, fmt.Errorf("invalid job: %v", err)
	}
	
	// 添加到队列
	if err := c.addJobToQueue(job); err != nil {
		return nil, fmt.Errorf("failed to add job to queue: %v", err)
	}
	
	config.Infof("Job dispatched: ID=%s, Type=%s, Queue=%s", job.ID, job.Type, job.Queue)
	return job, nil
}

// JobOption 任务选项函数类型
type JobOption func(*QueueJob)

// WithQueue 指定队列
func WithQueue(queue string) JobOption {
	return func(job *QueueJob) {
		job.Queue = queue
	}
}

// WithPriority 设置优先级
func WithPriority(priority int) JobOption {
	return func(job *QueueJob) {
		job.Priority = priority
	}
}

// WithDelay 延迟执行
func WithDelay(delay time.Duration) JobOption {
	return func(job *QueueJob) {
		job.ScheduledAt = time.Now().Add(delay)
	}
}

// WithMaxAttempts 设置最大重试次数
func WithMaxAttempts(attempts int) JobOption {
	return func(job *QueueJob) {
		job.MaxAttempts = attempts
	}
}

// ============= 快捷分发方法 =============

// DispatchNow 立即分发任务
func (c *BaseController) DispatchNow(jobType string, payload map[string]any) (*QueueJob, error) {
	return c.Dispatch(jobType, payload, WithPriority(1))
}

// DispatchLater 延迟分发任务
func (c *BaseController) DispatchLater(jobType string, payload map[string]any, delay time.Duration) (*QueueJob, error) {
	return c.Dispatch(jobType, payload, WithDelay(delay))
}

// DispatchToQueue 分发到指定队列
func (c *BaseController) DispatchToQueue(queue, jobType string, payload map[string]any) (*QueueJob, error) {
	return c.Dispatch(jobType, payload, WithQueue(queue))
}

// DispatchEmail 分发邮件任务
func (c *BaseController) DispatchEmail(to, subject, body string) (*QueueJob, error) {
	payload := map[string]any{
		"to":      to,
		"subject": subject,
		"body":    body,
	}
	return c.Dispatch("send_email", payload, WithQueue("mail"))
}

// DispatchNotification 分发通知任务
func (c *BaseController) DispatchNotification(userID int64, message string, notificationType string) (*QueueJob, error) {
	payload := map[string]any{
		"user_id": userID,
		"message": message,
		"type":    notificationType,
	}
	return c.Dispatch("send_notification", payload, WithQueue("notifications"))
}

// ============= 任务处理器注册 =============

// RegisterJobHandler 注册任务处理器
func (c *BaseController) RegisterJobHandler(jobType string, handler JobHandler) {
	jobHandlers[jobType] = handler
	config.Infof("Job handler registered for type: %s", jobType)
}

// GetJobHandler 获取任务处理器
func (c *BaseController) GetJobHandler(jobType string) (JobHandler, bool) {
	handler, exists := jobHandlers[jobType]
	return handler, exists
}

// ============= 队列处理方法 =============

// ProcessQueue 处理队列
func (c *BaseController) ProcessQueue(queueName string, maxJobs int) error {
	jobs := c.getJobsFromQueue(queueName, maxJobs)
	if len(jobs) == 0 {
		return nil
	}
	
	var processed, failed int
	
	for _, job := range jobs {
		if c.shouldProcessJob(job) {
			if err := c.processJob(job); err != nil {
				failed++
				config.Errorf("Job processing failed: ID=%s, Error=%v", job.ID, err)
			} else {
				processed++
				config.Debugf("Job processed successfully: ID=%s", job.ID)
			}
		}
	}
	
	config.Infof("Queue processing completed: Queue=%s, Processed=%d, Failed=%d", queueName, processed, failed)
	return nil
}

// processJob 处理单个任务
func (c *BaseController) processJob(job *QueueJob) error {
	// 更新任务状态
	job.Status = "processing"
	job.Attempts++
	
	start := time.Now()
	
	// 获取处理器
	handler, exists := c.GetJobHandler(job.Type)
	if !exists {
		return c.failJob(job, fmt.Errorf("no handler found for job type: %s", job.Type))
	}
	
	// 执行任务
	if err := handler(job); err != nil {
		duration := time.Since(start)
		
		// 检查是否还有重试机会
		if job.Attempts < job.MaxAttempts {
			// 重新加入队列等待重试
			job.Status = "pending"
			job.ScheduledAt = time.Now().Add(c.calculateRetryDelay(job.Attempts))
			c.updateJob(job)
			config.Warnf("Job retry scheduled: ID=%s, Attempt=%d/%d, Duration=%v", 
				job.ID, job.Attempts, job.MaxAttempts, duration)
			return nil
		} else {
			// 达到最大重试次数，标记为失败
			return c.failJob(job, err)
		}
	}
	
	// 任务成功完成
	return c.completeJob(job)
}

// completeJob 完成任务
func (c *BaseController) completeJob(job *QueueJob) error {
	now := time.Now()
	job.Status = "completed"
	job.ProcessedAt = &now
	job.Error = ""
	
	c.updateJob(job)
	c.recordJobMetrics(job, "completed")
	
	return nil
}

// failJob 失败任务
func (c *BaseController) failJob(job *QueueJob, err error) error {
	now := time.Now()
	job.Status = "failed"
	job.FailedAt = &now
	job.Error = err.Error()
	
	c.updateJob(job)
	c.recordJobMetrics(job, "failed")
	
	return err
}

// ============= 队列管理方法 =============

// GetQueueInfo 获取队列信息
func (c *BaseController) GetQueueInfo(queueName string) map[string]any {
	jobs := c.getAllJobsFromQueue(queueName)
	
	stats := map[string]int{
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
	}
	
	for _, job := range jobs {
		stats[job.Status]++
	}
	
	return map[string]any{
		"queue":        queueName,
		"total_jobs":   len(jobs),
		"pending":      stats["pending"],
		"processing":   stats["processing"],
		"completed":    stats["completed"],
		"failed":       stats["failed"],
		"success_rate": c.calculateSuccessRate(stats),
	}
}

// GetAllQueues 获取所有队列信息
func (c *BaseController) GetAllQueues() map[string]any {
	queues := make(map[string]any)
	
	for queueName := range queueStorage {
		queues[queueName] = c.GetQueueInfo(queueName)
	}
	
	return queues
}

// ClearQueue 清空队列
func (c *BaseController) ClearQueue(queueName string) error {
	queueStorage[queueName] = []*QueueJob{}
	config.Infof("Queue cleared: %s", queueName)
	return nil
}

// PurgeFailedJobs 清除失败的任务
func (c *BaseController) PurgeFailedJobs(queueName string) error {
	jobs := queueStorage[queueName]
	var remainingJobs []*QueueJob
	var purgedCount int
	
	for _, job := range jobs {
		if job.Status != "failed" {
			remainingJobs = append(remainingJobs, job)
		} else {
			purgedCount++
		}
	}
	
	queueStorage[queueName] = remainingJobs
	config.Infof("Purged %d failed jobs from queue: %s", purgedCount, queueName)
	
	return nil
}

// ============= 任务查询方法 =============

// GetJob 获取任务详情
func (c *BaseController) GetJob(jobID string) (*QueueJob, error) {
	for _, jobs := range queueStorage {
		for _, job := range jobs {
			if job.ID == jobID {
				return job, nil
			}
		}
	}
	return nil, fmt.Errorf("job not found: %s", jobID)
}

// GetJobsByType 根据类型获取任务
func (c *BaseController) GetJobsByType(jobType string) []*QueueJob {
	var result []*QueueJob
	
	for _, jobs := range queueStorage {
		for _, job := range jobs {
			if job.Type == jobType {
				result = append(result, job)
			}
		}
	}
	
	return result
}

// GetJobsByStatus 根据状态获取任务
func (c *BaseController) GetJobsByStatus(status string) []*QueueJob {
	var result []*QueueJob
	
	for _, jobs := range queueStorage {
		for _, job := range jobs {
			if job.Status == status {
				result = append(result, job)
			}
		}
	}
	
	return result
}

// ============= 内置任务处理器 =============

// InitDefaultJobHandlers 初始化默认任务处理器
func (c *BaseController) InitDefaultJobHandlers() {
	// 邮件发送任务
	c.RegisterJobHandler("send_email", func(job *QueueJob) error {
		to, _ := job.Payload["to"].(string)
		subject, _ := job.Payload["subject"].(string)
		body, _ := job.Payload["body"].(string)
		
		return c.SendSimpleMail(to, subject, body)
	})
	
	// 通知发送任务
	c.RegisterJobHandler("send_notification", func(job *QueueJob) error {
		userID, _ := job.Payload["user_id"].(int64)
		message, _ := job.Payload["message"].(string)
		notificationType, _ := job.Payload["type"].(string)
		
		config.Infof("Sending notification: UserID=%d, Type=%s, Message=%s", userID, notificationType, message)
		return nil
	})
	
	// 数据清理任务
	c.RegisterJobHandler("cleanup_data", func(job *QueueJob) error {
		tableName, _ := job.Payload["table"].(string)
		days, _ := job.Payload["days"].(int)
		
		config.Infof("Cleaning up data: Table=%s, Days=%d", tableName, days)
		// 这里应该执行实际的数据清理逻辑
		return nil
	})
	
	// 文件处理任务
	c.RegisterJobHandler("process_file", func(job *QueueJob) error {
		filePath, _ := job.Payload["file_path"].(string)
		operation, _ := job.Payload["operation"].(string)
		
		config.Infof("Processing file: Path=%s, Operation=%s", filePath, operation)
		// 这里应该执行实际的文件处理逻辑
		return nil
	})
}

// ============= 队列辅助方法 =============

// generateJobID 生成任务ID
func (c *BaseController) generateJobID() string {
	return fmt.Sprintf("job_%d_%s", time.Now().UnixNano(), c.GenerateSecureToken(8))
}

// validateJob 验证任务
func (c *BaseController) validateJob(job *QueueJob) error {
	if job.Type == "" {
		return fmt.Errorf("job type cannot be empty")
	}
	
	if job.MaxAttempts < 1 {
		return fmt.Errorf("max attempts must be at least 1")
	}
	
	if job.Priority < 1 || job.Priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}
	
	return nil
}

// addJobToQueue 添加任务到队列
func (c *BaseController) addJobToQueue(job *QueueJob) error {
	if queueStorage[job.Queue] == nil {
		queueStorage[job.Queue] = []*QueueJob{}
	}
	
	queueStorage[job.Queue] = append(queueStorage[job.Queue], job)
	return nil
}

// getJobsFromQueue 从队列获取任务
func (c *BaseController) getJobsFromQueue(queueName string, limit int) []*QueueJob {
	jobs := queueStorage[queueName]
	if jobs == nil {
		return []*QueueJob{}
	}
	
	var readyJobs []*QueueJob
	now := time.Now()
	
	for _, job := range jobs {
		if c.shouldProcessJob(job) && now.After(job.ScheduledAt) {
			readyJobs = append(readyJobs, job)
			if len(readyJobs) >= limit {
				break
			}
		}
	}
	
	return readyJobs
}

// getAllJobsFromQueue 获取队列中的所有任务
func (c *BaseController) getAllJobsFromQueue(queueName string) []*QueueJob {
	jobs := queueStorage[queueName]
	if jobs == nil {
		return []*QueueJob{}
	}
	
	// 返回副本
	result := make([]*QueueJob, len(jobs))
	copy(result, jobs)
	return result
}

// shouldProcessJob 检查是否应该处理任务
func (c *BaseController) shouldProcessJob(job *QueueJob) bool {
	return job.Status == "pending" && job.Attempts < job.MaxAttempts
}

// updateJob 更新任务
func (c *BaseController) updateJob(job *QueueJob) {
	// 在实际实现中，这里应该更新数据库或Redis中的任务状态
	config.Debugf("Job updated: ID=%s, Status=%s, Attempts=%d", job.ID, job.Status, job.Attempts)
}

// calculateRetryDelay 计算重试延迟
func (c *BaseController) calculateRetryDelay(attempt int) time.Duration {
	// 指数退避算法
	delay := time.Duration(attempt*attempt) * time.Minute
	if delay > 30*time.Minute {
		delay = 30 * time.Minute
	}
	return delay
}

// calculateSuccessRate 计算成功率
func (c *BaseController) calculateSuccessRate(stats map[string]int) float64 {
	total := stats["completed"] + stats["failed"]
	if total == 0 {
		return 0
	}
	return float64(stats["completed"]) / float64(total) * 100
}

// ============= 队列监控方法 =============

// recordJobMetrics 记录任务指标
func (c *BaseController) recordJobMetrics(job *QueueJob, status string) {
	metrics := map[string]any{
		"job_id":    job.ID,
		"job_type":  job.Type,
		"queue":     job.Queue,
		"status":    status,
		"attempts":  job.Attempts,
		"priority":  job.Priority,
		"created_at": job.CreatedAt,
	}
	
	if job.ProcessedAt != nil {
		duration := job.ProcessedAt.Sub(job.CreatedAt)
		metrics["duration"] = duration.String()
	}
	
	config.Infof("Job metrics: %+v", metrics)
}

// GetQueueMetrics 获取队列指标
func (c *BaseController) GetQueueMetrics() map[string]any {
	metrics := map[string]any{
		"total_queues":     len(queueStorage),
		"total_jobs":       0,
		"job_types":        make(map[string]int),
		"queue_sizes":      make(map[string]int),
		"status_breakdown": make(map[string]int),
	}
	
	for queueName, jobs := range queueStorage {
		metrics["total_jobs"] = metrics["total_jobs"].(int) + len(jobs)
		metrics["queue_sizes"].(map[string]int)[queueName] = len(jobs)
		
		for _, job := range jobs {
			// 统计任务类型
			jobTypes := metrics["job_types"].(map[string]int)
			jobTypes[job.Type]++
			
			// 统计状态
			statusBreakdown := metrics["status_breakdown"].(map[string]int)
			statusBreakdown[job.Status]++
		}
	}
	
	return metrics
}

// ExportQueueData 导出队列数据
func (c *BaseController) ExportQueueData(queueName string) ([]byte, error) {
	jobs := c.getAllJobsFromQueue(queueName)
	return json.MarshalIndent(jobs, "", "  ")
}

// ImportQueueData 导入队列数据
func (c *BaseController) ImportQueueData(queueName string, data []byte) error {
	var jobs []*QueueJob
	if err := json.Unmarshal(data, &jobs); err != nil {
		return fmt.Errorf("failed to unmarshal queue data: %v", err)
	}
	
	queueStorage[queueName] = jobs
	config.Infof("Imported %d jobs to queue: %s", len(jobs), queueName)
	
	return nil
}