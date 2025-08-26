package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// 这个文件提供了调度系统的完整使用示例

// ============= 基础调度示例 =============

// ExampleBasicScheduling 基础调度示例
func ExampleBasicScheduling() {
	config.Info("=== 基础调度示例 ===")

	// 创建调度器
	scheduler := NewScheduler(DefaultSchedulerConfig())

	// 创建简单任务
	job := NewJobFunc("hello", "打印问候", func(ctx context.Context) error {
		config.Info("Hello from scheduled task!")
		return nil
	})

	// 创建任务
	task := NewTask("task1", "问候任务", "每分钟执行一次", "@every_minute", job)

	// 添加任务到调度器
	if err := scheduler.AddTask(task); err != nil {
		config.Errorf("添加任务失败: %v", err)
		return
	}

	// 启动调度器
	if err := scheduler.Start(); err != nil {
		config.Errorf("启动调度器失败: %v", err)
		return
	}
	defer scheduler.Stop()

	config.Info("调度器已启动，等待任务执行...")

	// 等待一段时间观察任务执行
	time.Sleep(time.Second * 10)

	// 获取统计信息
	stats := scheduler.GetStats()
	config.Infof("调度器统计: 总任务=%d, 运行中=%d, 等待中=%d",
		stats.TotalTasks, stats.RunningTasks, stats.PendingTasks)
}

// ============= Cron表达式示例 =============

// ExampleCronScheduling Cron调度示例
func ExampleCronScheduling() {
	config.Info("=== Cron调度示例 ===")

	// 展示Cron表达式解析
	cronExamples := []string{
		"0 */5 * * * *",  // 每5分钟
		"0 0 9 * * 1-5",  // 工作日上午9点
		"0 30 8 * * MON", // 每周一上午8:30
		"0 0 0 1 * *",    // 每月1号午夜
	}

	for _, expr := range cronExamples {
		if cron, err := ParseCronExpression(expr); err == nil {
			nextTime := cron.NextTime(time.Now())
			config.Infof("Cron表达式: %s -> 下次执行: %s",
				expr, nextTime.Format("2006-01-02 15:04:05"))
		} else {
			config.Errorf("无效的Cron表达式: %s -> %v", expr, err)
		}
	}

	// 使用便捷函数调度Cron任务
	err := ScheduleCronJob("cron_task1", "定时报告", "生成每小时报告",
		"0 0 * * * *", // 每小时执行
		func(ctx context.Context) error {
			config.Info("正在生成每小时报告...")
			time.Sleep(time.Second) // 模拟工作
			config.Info("每小时报告生成完成")
			return nil
		})

	if err != nil {
		config.Errorf("调度Cron任务失败: %v", err)
		return
	}

	config.Info("Cron任务已调度")
}

// ============= 任务执行器示例 =============

// ExampleTaskExecution 任务执行示例
func ExampleTaskExecution() {
	config.Info("=== 任务执行示例 ===")

	// 创建执行器池
	executor := NewExecutorPool(DefaultExecutorConfig())

	// 启动执行器
	if err := executor.Start(); err != nil {
		config.Errorf("启动执行器失败: %v", err)
		return
	}
	defer executor.Stop()

	// 创建不同类型的任务
	tasks := []*Task{
		NewTask("fast_task", "快速任务", "立即执行", "@once",
			NewJobFunc("fast", "快速任务", func(ctx context.Context) error {
				config.Info("执行快速任务...")
				time.Sleep(time.Millisecond * 100)
				return nil
			})),

		NewTask("slow_task", "慢速任务", "立即执行", "@once",
			NewJobFunc("slow", "慢速任务", func(ctx context.Context) error {
				config.Info("执行慢速任务...")
				time.Sleep(time.Second * 2)
				return nil
			})),

		NewTask("error_task", "错误任务", "立即执行", "@once",
			NewJobFunc("error", "错误任务", func(ctx context.Context) error {
				config.Info("执行错误任务...")
				return fmt.Errorf("任务执行失败")
			})),
	}

	// 提交任务执行
	var executions []*TaskExecution
	for _, task := range tasks {
		execution, err := executor.Execute(task)
		if err != nil {
			config.Errorf("提交任务失败: %v", err)
			continue
		}
		executions = append(executions, execution)
		config.Infof("任务已提交: %s (执行ID: %s)", task.Name, execution.ExecutionID)
	}

	// 等待执行完成
	time.Sleep(time.Second * 5)

	// 检查执行结果
	for _, execution := range executions {
		config.Infof("任务 %s 执行状态: %s, 耗时: %v",
			execution.Task.Name, execution.Status.String(), execution.Duration)
		if execution.LastError != nil {
			config.Errorf("任务 %s 执行错误: %v", execution.Task.Name, execution.LastError)
		}
	}

	// 获取执行器统计
	stats := executor.GetStats()
	config.Infof("执行器统计: 总执行=%d, 成功=%d, 失败=%d",
		stats.TotalExecuted, stats.TotalSuccessful, stats.TotalFailed)
}

// ============= 任务持久化示例 =============

// ExampleTaskPersistence 任务持久化示例
func ExampleTaskPersistence() {
	config.Info("=== 任务持久化示例 ===")

	// 创建文件存储
	storage := NewFileStorage("./scheduler_test_data")
	defer storage.Close()

	// 创建调度器配置
	schedulerConfig := DefaultSchedulerConfig()
	schedulerConfig.EnablePersistent = true

	// 创建调度器
	scheduler := NewScheduler(schedulerConfig)
	scheduler.SetStorage(storage)

	// 创建持久化任务
	persistentJob := NewJobFunc("persistent", "持久化任务", func(ctx context.Context) error {
		config.Infof("持久化任务执行: %s", time.Now().Format("15:04:05"))
		return nil
	})

	task := NewTask("persistent_task", "持久化任务", "每30秒执行", "30s", persistentJob)
	task.SetMetadata("category", "system")
	task.SetMetadata("priority", "high")

	// 添加任务（会自动持久化）
	if err := scheduler.AddTask(task); err != nil {
		config.Errorf("添加持久化任务失败: %v", err)
		return
	}

	config.Info("持久化任务已添加并保存到存储")

	// 演示从存储加载任务
	loadedTasks, err := storage.LoadTasks()
	if err != nil {
		config.Errorf("从存储加载任务失败: %v", err)
		return
	}

	config.Infof("从存储加载了 %d 个任务", len(loadedTasks))
	for _, loadedTask := range loadedTasks {
		config.Infof("加载的任务: %s (%s) - %s",
			loadedTask.Name, loadedTask.ID, loadedTask.Schedule)
		config.Infof("  元数据: %+v", loadedTask.Metadata)
	}

	// 清理测试数据
	for _, loadedTask := range loadedTasks {
		storage.DeleteTask(loadedTask.ID)
	}
}

// ============= 任务监控示例 =============

// ExampleTaskMonitoring 任务监控示例
func ExampleTaskMonitoring() {
	config.Info("=== 任务监控示例 ===")

	// 创建监控器
	monitor := NewExecutionMonitor()

	// 添加默认告警规则
	for _, rule := range DefaultAlertRules() {
		monitor.AddAlertRule(rule)
	}

	// 订阅监控指标
	subscriber := &LoggingSubscriber{}
	monitor.Subscribe(subscriber)

	// 启动监控器
	if err := monitor.Start(); err != nil {
		config.Errorf("启动监控器失败: %v", err)
		return
	}
	defer monitor.Stop()

	// 创建高级执行器
	executor := NewAdvancedExecutor(DefaultExecutorConfig())
	executor.monitor = monitor

	if err := executor.Start(); err != nil {
		config.Errorf("启动高级执行器失败: %v", err)
		return
	}
	defer executor.Stop()

	// 注册节流策略
	throttleStrategy := NewThrottleStrategy(2, time.Second*5)
	executor.RegisterStrategy("throttle", throttleStrategy)

	// 创建监控任务
	monitoringTasks := []*Task{
		NewTask("success_task", "成功任务", "@once", "@once",
			NewJobFunc("success", "成功任务", func(ctx context.Context) error {
				time.Sleep(time.Millisecond * 500)
				return nil
			})),

		NewTask("failure_task", "失败任务", "@once", "@once",
			NewJobFunc("failure", "失败任务", func(ctx context.Context) error {
				time.Sleep(time.Millisecond * 200)
				return fmt.Errorf("模拟任务失败")
			})),
	}

	// 使用策略执行任务
	for i := 0; i < 5; i++ {
		for _, task := range monitoringTasks {
			// 修复：避免直接复制包含锁的Task结构体
			taskCopy := &Task{
				ID:          fmt.Sprintf("%s_%d", task.ID, i),
				Name:        task.Name,
				Description: task.Description,
				Schedule:    task.Schedule,
				Job:         task.Job,
				Status:      task.Status,
				CreatedAt:   task.CreatedAt,
				UpdatedAt:   time.Now(),
				LastRunTime: task.LastRunTime,
				NextRunTime: task.NextRunTime,
				RunCount:    task.RunCount,
				FailCount:   task.FailCount,
				MaxRetries:  task.MaxRetries,
				Timeout:     task.Timeout,
				Metadata:    make(map[string]string),
			}
			// 复制元数据
			for k, v := range task.Metadata {
				taskCopy.Metadata[k] = v
			}

			execution, err := executor.ExecuteWithStrategy(taskCopy, "throttle")
			if err != nil {
				config.Errorf("执行任务失败: %v", err)
				continue
			}

			config.Infof("任务已提交监控: %s", execution.ExecutionID)
		}

		time.Sleep(time.Second)
	}

	// 等待执行完成
	time.Sleep(time.Second * 3)

	// 获取监控指标
	metrics := monitor.GetMetrics()
	config.Infof("监控指标:")
	config.Infof("  总执行: %d", metrics.TotalExecutions)
	config.Infof("  成功: %d", metrics.SuccessfulExecutions)
	config.Infof("  失败: %d", metrics.FailedExecutions)
	config.Infof("  错误率: %.2f%%", metrics.ErrorRate)
	config.Infof("  平均执行时间: %v", metrics.AverageExecutionTime)
	config.Infof("  当前运行: %d", metrics.CurrentlyRunning)

	// 显示任务级别指标
	config.Info("任务级别指标:")
	for taskID, taskMetrics := range metrics.TaskMetrics {
		config.Infof("  任务 %s:", taskID)
		config.Infof("    总执行: %d", taskMetrics.TotalExecutions)
		config.Infof("    成功率: %.2f%%", taskMetrics.SuccessRate)
		config.Infof("    平均时间: %v", taskMetrics.AverageTime)
		config.Infof("    连续失败: %d", taskMetrics.ConsecutiveFails)
	}
}

// ============= 复杂业务场景示例 =============

// ExampleBusinessScenario 复杂业务场景示例
func ExampleBusinessScenario() {
	config.Info("=== 复杂业务场景示例 ===")

	// 模拟一个电商系统的后台任务调度场景

	// 1. 订单处理任务
	orderJob := NewJobFunc("order_processor", "订单处理", func(ctx context.Context) error {
		config.Info("处理待支付订单...")
		time.Sleep(time.Millisecond * 500)
		config.Info("订单处理完成")
		return nil
	})

	// 2. 库存同步任务
	inventoryJob := NewJobFunc("inventory_sync", "库存同步", func(ctx context.Context) error {
		config.Info("同步库存数据...")
		time.Sleep(time.Millisecond * 800)
		config.Info("库存同步完成")
		return nil
	})

	// 3. 数据备份任务
	backupJob := NewJobFunc("data_backup", "数据备份", func(ctx context.Context) error {
		config.Info("执行数据备份...")
		time.Sleep(time.Second * 2)
		config.Info("数据备份完成")
		return nil
	})

	// 4. 报表生成任务
	reportJob := NewJobFunc("report_generation", "报表生成", func(ctx context.Context) error {
		config.Info("生成销售报表...")
		time.Sleep(time.Second)
		config.Info("销售报表生成完成")
		return nil
	})

	// 创建任务
	tasks := []*Task{
		NewTask("order_task", "订单处理任务", "每2分钟处理一次订单", "0 */2 * * * *", orderJob),
		NewTask("inventory_task", "库存同步任务", "每小时同步库存", "0 0 * * * *", inventoryJob),
		NewTask("backup_task", "数据备份任务", "每天凌晨2点备份", "0 0 2 * * *", backupJob),
		NewTask("report_task", "报表生成任务", "每天上午8点生成报表", "0 0 8 * * *", reportJob),
	}

	// 创建完整的调度系统
	schedulerConfig := DefaultSchedulerConfig()
	schedulerConfig.EnablePersistent = true
	schedulerConfig.EnableLogging = true

	scheduler := NewScheduler(schedulerConfig)

	// 设置存储
	storage := NewFileStorage("./business_scheduler_data")
	scheduler.SetStorage(storage)
	defer storage.Close()

	// 设置回调
	scheduler.SetOnTaskStart(func(task *Task) {
		config.Infof("🚀 任务开始: %s", task.Name)
	})

	scheduler.SetOnTaskComplete(func(task *Task, err error) {
		if err != nil {
			config.Errorf("❌ 任务失败: %s - %v", task.Name, err)
		} else {
			config.Infof("✅ 任务完成: %s", task.Name)
		}
	})

	// 添加所有任务
	for _, task := range tasks {
		task.MaxRetries = 3
		task.Timeout = time.Minute * 5
		task.SetMetadata("environment", "production")
		task.SetMetadata("department", "backend")

		if err := scheduler.AddTask(task); err != nil {
			config.Errorf("添加任务失败: %v", err)
			continue
		}
	}

	// 启动调度器
	if err := scheduler.Start(); err != nil {
		config.Errorf("启动调度器失败: %v", err)
		return
	}
	defer scheduler.Stop()

	config.Info("业务调度系统已启动")

	// 手动触发一次报表生成（用于演示）
	reportTask, _ := scheduler.GetTask("report_task")
	if reportTask != nil {
		reportTask.SetNextRunTime(time.Now().Add(time.Second * 2))
		config.Info("已安排立即执行报表生成任务")
	}

	// 运行一段时间
	time.Sleep(time.Second * 10)

	// 获取最终统计
	stats := scheduler.GetStats()
	config.Infof("业务调度系统统计:")
	config.Infof("  总任务: %d", stats.TotalTasks)
	config.Infof("  运行中: %d", stats.RunningTasks)
	config.Infof("  等待中: %d", stats.PendingTasks)
	config.Infof("  已完成: %d", stats.CompletedTasks)
	config.Infof("  失败: %d", stats.FailedTasks)

	// 清理测试数据
	for _, task := range tasks {
		storage.DeleteTask(task.ID)
	}
}

// ============= 运行所有示例 =============

// RunAllExamples 运行所有调度系统示例
func RunAllExamples() error {
	config.Info("=== YYHertz 调度系统示例演示 ===")

	// 1. 基础调度示例
	config.Info("\n1. 基础调度功能演示:")
	ExampleBasicScheduling()

	// 2. Cron表达式示例
	config.Info("\n2. Cron表达式功能演示:")
	ExampleCronScheduling()

	// 3. 任务执行示例
	config.Info("\n3. 任务执行器演示:")
	ExampleTaskExecution()

	// 4. 任务持久化示例
	config.Info("\n4. 任务持久化演示:")
	ExampleTaskPersistence()

	// 5. 任务监控示例
	config.Info("\n5. 任务监控演示:")
	ExampleTaskMonitoring()

	// 6. 复杂业务场景示例
	config.Info("\n6. 复杂业务场景演示:")
	ExampleBusinessScenario()

	config.Info("\n=== 所有调度系统示例演示完成 ===")
	return nil
}

// ============= 便捷使用函数示例 =============

// DemoConvenienceFunctions 演示便捷函数
func DemoConvenienceFunctions() {
	config.Info("=== 便捷函数使用演示 ===")

	// 启动全局调度器
	if err := StartGlobalScheduler(); err != nil {
		config.Errorf("启动全局调度器失败: %v", err)
		return
	}
	defer StopGlobalScheduler()

	// 使用便捷函数调度不同类型的任务

	// 1. 延迟执行
	err := ScheduleAfter("delayed_task", "延迟任务", "5秒后执行",
		time.Second*5, func(ctx context.Context) error {
			config.Info("延迟任务执行完成")
			return nil
		})
	if err != nil {
		config.Errorf("调度延迟任务失败: %v", err)
	}

	// 2. 定时执行
	scheduleTime := time.Now().Add(time.Second * 3)
	err = ScheduleAt("timed_task", "定时任务", "在指定时间执行",
		scheduleTime, func(ctx context.Context) error {
			config.Info("定时任务执行完成")
			return nil
		})
	if err != nil {
		config.Errorf("调度定时任务失败: %v", err)
	}

	// 3. 周期执行
	err = ScheduleEvery("periodic_task", "周期任务", "每2秒执行",
		time.Second*2, func(ctx context.Context) error {
			config.Info("周期任务执行完成")
			return nil
		})
	if err != nil {
		config.Errorf("调度周期任务失败: %v", err)
	}

	// 4. Cron表达式
	err = ScheduleCronJob("cron_demo", "Cron演示", "每分钟执行",
		"0 * * * * *", func(ctx context.Context) error {
			config.Info("Cron任务执行完成")
			return nil
		})
	if err != nil {
		config.Errorf("调度Cron任务失败: %v", err)
	}

	config.Info("所有便捷函数演示任务已调度")

	// 等待任务执行
	time.Sleep(time.Second * 10)

	// 获取全局调度器状态
	globalScheduler := GetGlobalScheduler()
	stats := globalScheduler.GetStats()
	config.Infof("全局调度器统计: 总任务=%d, 运行中=%d",
		stats.TotalTasks, stats.RunningTasks)

	// 清理任务
	taskIDs := []string{"delayed_task", "timed_task", "periodic_task", "cron_demo"}
	for _, taskID := range taskIDs {
		RemoveGlobalTask(taskID)
	}
}

// ============= 性能测试示例 =============

// DemoPerformanceTest 性能测试演示
func DemoPerformanceTest() {
	config.Info("=== 性能测试演示 ===")

	// 创建高性能配置
	executorConfig := &ExecutorConfig{
		WorkerCount:    10,
		QueueSize:      5000,
		MaxRetries:     1,
		RetryDelay:     time.Second,
		ExecuteTimeout: time.Second * 10,
		EnableMetrics:  true,
		EnableRecovery: true,
	}

	executor := NewExecutorPool(executorConfig)

	if err := executor.Start(); err != nil {
		config.Errorf("启动执行器失败: %v", err)
		return
	}
	defer executor.Stop()

	// 创建性能测试任务
	taskCount := 100
	startTime := time.Now()

	config.Infof("开始性能测试: 提交 %d 个任务", taskCount)

	for i := 0; i < taskCount; i++ {
		taskID := fmt.Sprintf("perf_task_%d", i)
		job := NewJobFunc(taskID, "性能测试任务", func(ctx context.Context) error {
			// 模拟轻量级工作负载
			time.Sleep(time.Millisecond * 10)
			return nil
		})

		task := NewTask(taskID, "性能测试任务", "@once", "@once", job)

		_, err := executor.Execute(task)
		if err != nil {
			config.Errorf("提交任务 %s 失败: %v", taskID, err)
		}
	}

	submitTime := time.Since(startTime)
	config.Infof("任务提交完成，耗时: %v", submitTime)

	// 等待所有任务完成
	for {
		stats := executor.GetStats()
		if stats.TotalExecuted >= int64(taskCount) {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	totalTime := time.Since(startTime)
	stats := executor.GetStats()

	config.Infof("性能测试结果:")
	config.Infof("  任务数量: %d", taskCount)
	config.Infof("  总耗时: %v", totalTime)
	config.Infof("  提交耗时: %v", submitTime)
	config.Infof("  执行耗时: %v", totalTime-submitTime)
	config.Infof("  成功执行: %d", stats.TotalSuccessful)
	config.Infof("  失败执行: %d", stats.TotalFailed)
	config.Infof("  平均QPS: %.2f", float64(taskCount)/totalTime.Seconds())
	config.Infof("  工作协程: %d", stats.WorkerCount)
}
