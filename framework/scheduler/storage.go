package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// Storage 存储接口
type Storage interface {
	// SaveTask 保存任务
	SaveTask(task *Task) error
	// LoadTask 加载任务
	LoadTask(taskID string) (*Task, error)
	// LoadTasks 加载所有任务
	LoadTasks() ([]*Task, error)
	// DeleteTask 删除任务
	DeleteTask(taskID string) error
	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(taskID string, status TaskStatus) error
	// SaveExecution 保存执行记录
	SaveExecution(execution *TaskExecution) error
	// LoadExecutions 加载执行记录
	LoadExecutions(taskID string, limit int) ([]*TaskExecution, error)
	// Close 关闭存储
	Close() error
}

// ============= 文件存储实现 =============

// FileStorage 文件存储
type FileStorage struct {
	basePath      string
	tasksDir      string
	executionsDir string
	mutex         sync.RWMutex
}

// NewFileStorage 创建文件存储
func NewFileStorage(basePath string) *FileStorage {
	tasksDir := filepath.Join(basePath, "tasks")
	executionsDir := filepath.Join(basePath, "executions")

	// 创建目录
	os.MkdirAll(tasksDir, 0755)
	os.MkdirAll(executionsDir, 0755)

	return &FileStorage{
		basePath:      basePath,
		tasksDir:      tasksDir,
		executionsDir: executionsDir,
	}
}

// SaveTask 保存任务
func (fs *FileStorage) SaveTask(task *Task) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// 创建可序列化的任务数据
	taskData := &SerializableTask{
		ID:          task.ID,
		Name:        task.Name,
		Description: task.Description,
		Schedule:    task.Schedule,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		LastRunTime: task.LastRunTime,
		NextRunTime: task.NextRunTime,
		RunCount:    task.RunCount,
		FailCount:   task.FailCount,
		MaxRetries:  task.MaxRetries,
		Timeout:     task.Timeout,
		Metadata:    task.Metadata,
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(taskData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 写入文件
	filename := filepath.Join(fs.tasksDir, task.ID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// LoadTask 加载任务
func (fs *FileStorage) LoadTask(taskID string) (*Task, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filename := filepath.Join(fs.tasksDir, taskID+".json")

	// 读取文件
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("task %s not found", taskID)
		}
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	// 反序列化
	var taskData SerializableTask
	if err := json.Unmarshal(data, &taskData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// 转换为Task对象（注意：Job字段需要在调度器中重新设置）
	task := &Task{
		ID:          taskData.ID,
		Name:        taskData.Name,
		Description: taskData.Description,
		Schedule:    taskData.Schedule,
		Status:      taskData.Status,
		CreatedAt:   taskData.CreatedAt,
		UpdatedAt:   taskData.UpdatedAt,
		LastRunTime: taskData.LastRunTime,
		NextRunTime: taskData.NextRunTime,
		RunCount:    taskData.RunCount,
		FailCount:   taskData.FailCount,
		MaxRetries:  taskData.MaxRetries,
		Timeout:     taskData.Timeout,
		Metadata:    taskData.Metadata,
	}

	return task, nil
}

// LoadTasks 加载所有任务
func (fs *FileStorage) LoadTasks() ([]*Task, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	// 读取任务目录
	files, err := os.ReadDir(fs.tasksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	var tasks []*Task
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			taskID := file.Name()[:len(file.Name())-5] // 去掉.json后缀

			task, err := fs.LoadTask(taskID)
			if err != nil {
				config.Errorf("Failed to load task %s: %v", taskID, err)
				continue
			}

			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// DeleteTask 删除任务
func (fs *FileStorage) DeleteTask(taskID string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	filename := filepath.Join(fs.tasksDir, taskID+".json")

	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("task %s not found", taskID)
		}
		return fmt.Errorf("failed to delete task file: %w", err)
	}

	return nil
}

// UpdateTaskStatus 更新任务状态
func (fs *FileStorage) UpdateTaskStatus(taskID string, status TaskStatus) error {
	// 加载任务
	task, err := fs.LoadTask(taskID)
	if err != nil {
		return err
	}

	// 更新状态
	task.Status = status
	task.UpdatedAt = time.Now()

	// 保存任务
	return fs.SaveTask(task)
}

// SaveExecution 保存执行记录
func (fs *FileStorage) SaveExecution(execution *TaskExecution) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// 创建可序列化的执行数据
	execData := &SerializableExecution{
		ExecutionID: execution.ExecutionID,
		TaskID:      execution.Task.ID,
		TaskName:    execution.Task.Name,
		StartTime:   execution.StartTime,
		EndTime:     execution.EndTime,
		Duration:    execution.Duration,
		Status:      execution.Status,
		RetryCount:  execution.RetryCount,
		WorkerID:    execution.WorkerID,
		Metadata:    execution.Metadata,
	}

	if execution.LastError != nil {
		execData.Error = execution.LastError.Error()
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(execData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal execution: %w", err)
	}

	// 创建任务执行目录
	taskExecDir := filepath.Join(fs.executionsDir, execution.Task.ID)
	if err := os.MkdirAll(taskExecDir, 0755); err != nil {
		return fmt.Errorf("failed to create execution directory: %w", err)
	}

	// 写入文件
	filename := filepath.Join(taskExecDir, execution.ExecutionID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write execution file: %w", err)
	}

	return nil
}

// LoadExecutions 加载执行记录
func (fs *FileStorage) LoadExecutions(taskID string, limit int) ([]*TaskExecution, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	taskExecDir := filepath.Join(fs.executionsDir, taskID)

	// 检查目录是否存在
	if _, err := os.Stat(taskExecDir); os.IsNotExist(err) {
		return []*TaskExecution{}, nil
	}

	// 读取执行记录目录
	files, err := os.ReadDir(taskExecDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read executions directory: %w", err)
	}

	var executions []*TaskExecution
	count := 0

	// 按文件修改时间倒序排列
	for i := len(files) - 1; i >= 0 && (limit <= 0 || count < limit); i-- {
		file := files[i]
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			execID := file.Name()[:len(file.Name())-5] // 去掉.json后缀

			execution, err := fs.loadExecution(taskID, execID)
			if err != nil {
				config.Errorf("Failed to load execution %s: %v", execID, err)
				continue
			}

			executions = append(executions, execution)
			count++
		}
	}

	return executions, nil
}

// loadExecution 加载单个执行记录
func (fs *FileStorage) loadExecution(taskID, executionID string) (*TaskExecution, error) {
	filename := filepath.Join(fs.executionsDir, taskID, executionID+".json")

	// 读取文件
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read execution file: %w", err)
	}

	// 反序列化
	var execData SerializableExecution
	if err := json.Unmarshal(data, &execData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution: %w", err)
	}

	// 转换为TaskExecution对象
	execution := &TaskExecution{
		ExecutionID: execData.ExecutionID,
		StartTime:   execData.StartTime,
		EndTime:     execData.EndTime,
		Duration:    execData.Duration,
		Status:      execData.Status,
		RetryCount:  execData.RetryCount,
		WorkerID:    execData.WorkerID,
		Metadata:    execData.Metadata,
	}

	if execData.Error != "" {
		execution.LastError = fmt.Errorf("%s", execData.Error)
	}

	return execution, nil
}

// Close 关闭存储
func (fs *FileStorage) Close() error {
	// 文件存储不需要特殊的关闭操作
	return nil
}

// ============= 序列化数据结构 =============

// SerializableTask 可序列化的任务
type SerializableTask struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Schedule    string            `json:"schedule"`
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
}

// SerializableExecution 可序列化的执行记录
type SerializableExecution struct {
	ExecutionID string          `json:"execution_id"`
	TaskID      string          `json:"task_id"`
	TaskName    string          `json:"task_name"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	Duration    time.Duration   `json:"duration"`
	Status      ExecutionStatus `json:"status"`
	RetryCount  int             `json:"retry_count"`
	WorkerID    int             `json:"worker_id"`
	Error       string          `json:"error,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// ============= 内存存储实现 =============

// MemoryStorage 内存存储（用于测试）
type MemoryStorage struct {
	tasks      map[string]*Task
	executions map[string][]*TaskExecution
	mutex      sync.RWMutex
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		tasks:      make(map[string]*Task),
		executions: make(map[string][]*TaskExecution),
	}
}

// SaveTask 保存任务
func (ms *MemoryStorage) SaveTask(task *Task) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 深拷贝任务
	taskCopy := *task
	taskCopy.Metadata = make(map[string]string)
	for k, v := range task.Metadata {
		taskCopy.Metadata[k] = v
	}

	ms.tasks[task.ID] = &taskCopy
	return nil
}

// LoadTask 加载任务
func (ms *MemoryStorage) LoadTask(taskID string) (*Task, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	task, exists := ms.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	// 返回深拷贝
	taskCopy := *task
	taskCopy.Metadata = make(map[string]string)
	for k, v := range task.Metadata {
		taskCopy.Metadata[k] = v
	}

	return &taskCopy, nil
}

// LoadTasks 加载所有任务
func (ms *MemoryStorage) LoadTasks() ([]*Task, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	tasks := make([]*Task, 0, len(ms.tasks))
	for _, task := range ms.tasks {
		// 深拷贝
		taskCopy := *task
		taskCopy.Metadata = make(map[string]string)
		for k, v := range task.Metadata {
			taskCopy.Metadata[k] = v
		}
		tasks = append(tasks, &taskCopy)
	}

	return tasks, nil
}

// DeleteTask 删除任务
func (ms *MemoryStorage) DeleteTask(taskID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.tasks[taskID]; !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	delete(ms.tasks, taskID)
	delete(ms.executions, taskID)
	return nil
}

// UpdateTaskStatus 更新任务状态
func (ms *MemoryStorage) UpdateTaskStatus(taskID string, status TaskStatus) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	task, exists := ms.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	task.Status = status
	task.UpdatedAt = time.Now()
	return nil
}

// SaveExecution 保存执行记录
func (ms *MemoryStorage) SaveExecution(execution *TaskExecution) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 深拷贝执行记录
	execCopy := *execution
	execCopy.Metadata = make(map[string]any)
	for k, v := range execution.Metadata {
		execCopy.Metadata[k] = v
	}

	taskID := execution.Task.ID
	ms.executions[taskID] = append(ms.executions[taskID], &execCopy)

	return nil
}

// LoadExecutions 加载执行记录
func (ms *MemoryStorage) LoadExecutions(taskID string, limit int) ([]*TaskExecution, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	executions, exists := ms.executions[taskID]
	if !exists {
		return []*TaskExecution{}, nil
	}

	// 返回最近的执行记录
	start := 0
	if limit > 0 && len(executions) > limit {
		start = len(executions) - limit
	}

	result := make([]*TaskExecution, len(executions)-start)
	for i, exec := range executions[start:] {
		// 深拷贝
		execCopy := *exec
		execCopy.Metadata = make(map[string]any)
		for k, v := range exec.Metadata {
			execCopy.Metadata[k] = v
		}
		result[i] = &execCopy
	}

	return result, nil
}

// Close 关闭存储
func (ms *MemoryStorage) Close() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.tasks = nil
	ms.executions = nil
	return nil
}

// ============= 存储工厂 =============

// StorageType 存储类型
type StorageType string

const (
	StorageTypeFile   StorageType = "file"
	StorageTypeMemory StorageType = "memory"
)

// StorageConfig 存储配置
type StorageConfig struct {
	Type     StorageType `json:"type"`
	FilePath string      `json:"file_path,omitempty"`
}

// CreateStorage 创建存储实例
func CreateStorage(config *StorageConfig) (Storage, error) {
	switch config.Type {
	case StorageTypeFile:
		if config.FilePath == "" {
			config.FilePath = "./scheduler_data"
		}
		return NewFileStorage(config.FilePath), nil
	case StorageTypeMemory:
		return NewMemoryStorage(), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// ============= 存储迁移工具 =============

// MigrateStorage 存储迁移
func MigrateStorage(from, to Storage) error {
	// 加载所有任务
	tasks, err := from.LoadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks from source storage: %w", err)
	}

	// 保存到目标存储
	for _, task := range tasks {
		if err := to.SaveTask(task); err != nil {
			return fmt.Errorf("failed to save task %s to target storage: %w", task.ID, err)
		}

		// 迁移执行记录
		executions, err := from.LoadExecutions(task.ID, 0)
		if err != nil {
			config.Errorf("Failed to load executions for task %s: %v", task.ID, err)
			continue
		}

		for _, execution := range executions {
			if err := to.SaveExecution(execution); err != nil {
				config.Errorf("Failed to save execution %s: %v", execution.ExecutionID, err)
			}
		}
	}

	config.Infof("Successfully migrated %d tasks", len(tasks))
	return nil
}

// BackupStorage 备份存储
func BackupStorage(storage Storage, backupPath string) error {
	// 创建文件存储作为备份
	backupStorage := NewFileStorage(backupPath)
	defer backupStorage.Close()

	return MigrateStorage(storage, backupStorage)
}

// RestoreStorage 恢复存储
func RestoreStorage(storage Storage, backupPath string) error {
	// 从备份文件存储中恢复
	backupStorage := NewFileStorage(backupPath)
	defer backupStorage.Close()

	return MigrateStorage(backupStorage, storage)
}
