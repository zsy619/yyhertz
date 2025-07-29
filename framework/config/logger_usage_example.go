package config

import (
	"time"
)

// SingletonLoggerUsageExample 展示单例日志管理器的使用方法
func SingletonLoggerUsageExample() {
	// ============= 1. 初始化全局日志 =============

	// 方式1: 使用默认配置
	InitGlobalLogger(DefaultLogConfig())

	// 方式2: 使用自定义配置
	customConfig := &LogConfig{
		Level:           LogLevelDebug,
		Format:          LogFormatJSON,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/singleton.log",
		MaxSize:         50,
		MaxAge:          7,
		MaxBackups:      5,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"service": "hertz-singleton-demo",
			"version": "2.0.0",
		},
	}
	InitGlobalLogger(customConfig)

	// ============= 2. 基础日志使用 =============

	// 直接使用全局函数（推荐方式）
	Debug("这是调试信息")
	Debugf("用户ID: %s 执行了操作", "user123")

	Info("应用启动成功")
	Infof("服务监听在端口: %d", 8080)

	Warn("内存使用率较高")
	Warnf("连接池使用率: %.2f%%", 85.5)

	Error("数据库连接失败")
	Errorf("Redis连接超时: %v", "5s")

	// ============= 3. 结构化日志使用 =============

	// 添加多个字段
	WithFields(map[string]any{
		"user_id":    "user123",
		"action":     "login",
		"ip":         "192.168.1.100",
		"user_agent": "Mozilla/5.0...",
	}).Info("用户登录成功")

	// 添加单个字段
	WithField("order_id", "order456").Info("订单创建成功")

	// 添加请求ID
	WithRequestID("req-abc-123").Error("处理请求时发生错误")

	// 添加用户ID
	WithUserID("user789").Warn("用户尝试访问受限资源")

	// ============= 4. 动态配置管理 =============

	// 获取全局日志管理器
	logger := GetGlobalLogger()

	// 动态更新日志级别
	logger.UpdateLevel(LogLevelError)
	Info("这条信息不会显示，因为当前级别是Error")
	Error("这条错误会显示")

	// 恢复调试级别
	logger.UpdateLevel(LogLevelDebug)
	Debug("现在又可以看到调试信息了")

	// 动态更新整个配置
	newConfig := &LogConfig{
		Level:         LogLevelInfo,
		Format:        LogFormatText,
		EnableConsole: true,
		EnableFile:    false, // 关闭文件输出
	}
	logger.UpdateConfig(newConfig)
	Info("配置已更新为文本格式，仅控制台输出")

	// ============= 5. 获取底层实例 =============

	// 获取hertz logger实例
	hertzLogger := logger.GetLogger()
	_ = hertzLogger // 可用于hertz框架集成

	// 获取原始logrus实例
	rawLogger := logger.GetRawLogger()
	rawLogger.WithField("custom", "value").Info("使用原始logrus功能")

	// 获取当前配置
	currentConfig := logger.GetConfig()
	Infof("当前日志级别: %s", currentConfig.Level)
}

// WebServiceLoggerExample 展示在Web服务中的使用
func WebServiceLoggerExample() {
	// 在应用启动时初始化
	logConfig := DevelopmentLogConfig()
	logConfig.Fields["service"] = "web-api"
	logConfig.Fields["instance"] = "web-01"

	InitGlobalLogger(logConfig)

	// 在请求处理中使用
	simulateWebRequest := func(requestID, userID, endpoint string) {
		// 记录请求开始
		WithFields(map[string]any{
			"request_id": requestID,
			"user_id":    userID,
			"endpoint":   endpoint,
			"method":     "GET",
		}).Info("处理请求开始")

		// 模拟业务逻辑
		WithRequestID(requestID).Debug("验证用户权限")
		WithRequestID(requestID).Debug("查询数据库")

		// 记录请求完成
		WithFields(map[string]any{
			"request_id":    requestID,
			"user_id":       userID,
			"endpoint":      endpoint,
			"response_time": "120ms",
			"status_code":   200,
		}).Info("请求处理完成")
	}

	// 模拟多个请求
	simulateWebRequest("req-001", "user123", "/api/users")
	simulateWebRequest("req-002", "user456", "/api/orders")
	simulateWebRequest("req-003", "user789", "/api/products")
}

// MicroserviceLoggerExample 展示在微服务中的使用
func MicroserviceLoggerExample() {
	// 每个微服务可以有自己的配置
	serviceConfig := ProductionLogConfig()
	serviceConfig.Fields = map[string]any{
		"service":    "user-service",
		"version":    "v1.2.3",
		"datacenter": "us-west-1",
		"pod":        "user-service-7d6f8b9c4d-xyz123",
	}

	InitGlobalLogger(serviceConfig)

	// 服务启动日志
	Info("微服务启动中...")

	// 依赖检查
	WithField("dependency", "database").Info("检查数据库连接")
	WithField("dependency", "redis").Info("检查Redis连接")
	WithField("dependency", "message-queue").Info("检查消息队列连接")

	// 服务就绪
	WithFields(map[string]any{
		"port":         8080,
		"health_check": "/health",
		"metrics":      "/metrics",
	}).Info("微服务启动完成")

	// 模拟服务间调用
	WithFields(map[string]any{
		"target_service": "order-service",
		"method":         "POST",
		"endpoint":       "/api/orders",
		"trace_id":       "trace-abc-123",
	}).Info("调用下游服务")
}

// PerformanceLoggerExample 展示高性能日志使用
func PerformanceLoggerExample() {
	// 高性能配置：只记录错误和致命错误
	perfConfig := HighPerformanceLogConfig()
	InitGlobalLogger(perfConfig)

	// 这些日志不会被记录（级别太低）
	Debug("调试信息")
	Info("普通信息")
	Warn("警告信息")

	// 这些日志会被记录
	Error("严重错误")
	WithField("error_code", "E001").Error("业务错误")

	// 在性能敏感的代码中，可以先检查日志级别
	logger := GetGlobalLogger()
	if logger.GetConfig().Level == LogLevelDebug {
		// 只有在调试模式下才执行复杂的日志构建
		expensiveDebugInfo := buildExpensiveDebugInfo()
		Debugf("复杂调试信息: %+v", expensiveDebugInfo)
	}
}

// buildExpensiveDebugInfo 模拟构建复杂调试信息的函数
func buildExpensiveDebugInfo() map[string]any {
	// 模拟一些耗时的调试信息构建
	return map[string]any{
		"memory_usage":  "150MB",
		"goroutines":    125,
		"gc_cycles":     45,
		"request_queue": []string{"req1", "req2", "req3"},
	}
}
