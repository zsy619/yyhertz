package yyhertz

import (
	"github.com/zsy619/yyhertz/framework/config"
)

// LogUsageExample 展示各种日志使用方法的示例
func LogUsageExample() {
	// 1. 使用不同环境的预设配置

	// 开发环境
	devApp := NewAppWithLogConfig(config.DevelopmentLogConfig())
	devApp.LogDebug("Development mode: detailed debugging info")

	// 生产环境
	prodApp := NewAppWithLogConfig(config.ProductionLogConfig())
	prodApp.LogInfo("Production mode: general information")

	// 测试环境
	testApp := NewAppWithLogConfig(config.TestLogConfig())
	testApp.LogWarn("Test mode: warning level")

	// 高性能环境
	perfApp := NewAppWithLogConfig(config.HighPerformanceLogConfig())
	perfApp.LogError("High performance mode: errors only")

	// 2. 动态修改日志配置

	app := NewApp()

	// 动态更新日志级别
	app.UpdateLogLevel(config.LogLevelDebug)
	app.LogDebug("Now debug level is enabled")

	// 获取当前配置并修改
	currentConfig := app.GetLogConfig()
	newConfig := currentConfig.UpdateConfigLevel(config.LogLevelError)
	app.SetLogConfig(newConfig)

	// 3. 使用结构化日志

	// 带字段的日志
	app.LogWithFields(config.LogLevelInfo, "User operation", map[string]any{
		"user_id":   "user123",
		"action":    "login",
		"ip":        "192.168.1.100",
		"timestamp": "2024-01-01T12:00:00Z",
	})

	// 带请求ID的日志
	app.LogWithRequestID(config.LogLevelInfo, "Processing request", "req-abc-123")

	// 带用户ID的日志
	app.LogWithUserID(config.LogLevelWarn, "User exceeded rate limit", "user456")

	// 4. 使用配置便捷方法

	// 添加全局字段
	configWithFields := config.DefaultLogConfig().AddConfigFields(map[string]any{
		"service":     "my-service",
		"version":     "2.0.0",
		"environment": "staging",
	})

	appWithFields := NewAppWithLogConfig(configWithFields)
	appWithFields.LogInfo("Service started with additional context")

	// 5. 在控制器中使用上下文日志

	// 这通常在实际的HTTP处理器中使用
	// func (c *MyController) HandleRequest() {
	//     logger := c.app.GetLoggerWithContext(c.Ctx)
	//     // logger可以用于记录带上下文信息的日志
	// }
}

// CustomLogConfigExample 展示如何创建自定义日志配置
func CustomLogConfigExample() {
	// 创建自定义配置
	customConfig := &config.LogConfig{
		Level:           config.LogLevelInfo,
		Format:          config.LogFormatJSON,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/custom.log",
		MaxSize:         75,
		MaxAge:          14,
		MaxBackups:      7,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		Fields: map[string]any{
			"application": "hertz-mvc",
			"module":      "custom",
			"datacenter":  "us-west-1",
		},
	}

	app := NewAppWithLogConfig(customConfig)

	// 使用自定义配置的应用
	app.LogInfo("Custom configuration is working")
	app.LogWithFields(config.LogLevelInfo, "Custom log with additional fields", map[string]any{
		"custom_field": "custom_value",
		"processed":    true,
	})
}

// LogLevelDemoExample 展示所有日志级别的使用
func LogLevelDemoExample() {
	app := NewAppWithLogConfig(config.DevelopmentLogConfig())

	// 标准日志方法
	app.LogDebug("Debug: Detailed information for diagnosing problems")
	app.LogInfo("Info: General information about application flow")
	app.LogWarn("Warn: Something unexpected happened, but application can continue")
	app.LogError("Error: A serious problem occurred")

	// 注意：LogFatal和LogPanic会终止程序，在示例中谨慎使用
	// app.LogFatal("Fatal: Application cannot continue and will exit")
	// app.LogPanic("Panic: Critical error that causes panic")

	// 使用LogWithFields的各级别示例
	fields := map[string]any{
		"component": "demo",
		"operation": "level_test",
	}

	app.LogWithFields(config.LogLevelDebug, "Debug with fields", fields)
	app.LogWithFields(config.LogLevelInfo, "Info with fields", fields)
	app.LogWithFields(config.LogLevelWarn, "Warning with fields", fields)
	app.LogWithFields(config.LogLevelError, "Error with fields", fields)
}
