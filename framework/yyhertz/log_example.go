package yyhertz

import (
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// 示例：创建开发环境日志配置
func ExampleDevelopmentLogConfig() *App {
	logConfig := &config.LogConfig{
		Level:           config.LogLevelDebug,
		Format:          config.LogFormatText,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "logs/dev.log",
		MaxSize:         50,
		MaxAge:          3,
		MaxBackups:      5,
		Compress:        false,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		Fields: map[string]any{
			"env":     "development",
			"service": "yyhertz",
		},
	}

	return NewAppWithLogConfig(logConfig)
}

// 示例：创建生产环境日志配置
func ExampleProductionLogConfig() *App {
	logConfig := &config.LogConfig{
		Level:           config.LogLevelInfo,
		Format:          config.LogFormatJSON,
		EnableConsole:   false,
		EnableFile:      true,
		FilePath:        "logs/prod.log",
		MaxSize:         100,
		MaxAge:          30,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      false,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"env":     "production",
			"service": "yyhertz",
			"version": "1.0.0",
		},
	}

	return NewAppWithLogConfig(logConfig)
}

// 示例：创建测试环境日志配置
func ExampleTestLogConfig() *App {
	logConfig := &config.LogConfig{
		Level:         config.LogLevelWarn,
		Format:        config.LogFormatText,
		EnableConsole: true,
		EnableFile:    false,
		ShowCaller:    false,
		ShowTimestamp: false,
		Fields:        map[string]any{},
	}

	return NewAppWithLogConfig(logConfig)
}

// 示例：创建高性能日志配置（最小日志）
func ExampleHighPerformanceLogConfig() *App {
	logConfig := &config.LogConfig{
		Level:         config.LogLevelError,
		Format:        config.LogFormatJSON,
		EnableConsole: false,
		EnableFile:    true,
		FilePath:      "logs/error.log",
		MaxSize:       200,
		MaxAge:        7,
		MaxBackups:    3,
		Compress:      true,
		ShowCaller:    true,
		ShowTimestamp: true,
		Fields: map[string]any{
			"mode": "high-performance",
		},
	}

	return NewAppWithLogConfig(logConfig)
}
