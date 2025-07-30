package config

import (
	"io"
	"os"
	"path/filepath"
	"time"

	hertzlogrus "github.com/hertz-contrib/logger/logrus"
	"github.com/sirupsen/logrus"
)

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// LogFormat 日志格式
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

// LogConfig 日志配置结构
type LogConfig struct {
	// 基础配置
	Level  LogLevel  `json:"level" yaml:"level"`   // 日志级别
	Format LogFormat `json:"format" yaml:"format"` // 日志格式

	// 输出配置
	EnableConsole bool   `json:"enable_console" yaml:"enable_console"` // 是否输出到控制台
	EnableFile    bool   `json:"enable_file" yaml:"enable_file"`       // 是否输出到文件
	FilePath      string `json:"file_path" yaml:"file_path"`           // 日志文件路径
	MaxSize       int    `json:"max_size" yaml:"max_size"`             // 单个日志文件最大大小(MB)
	MaxAge        int    `json:"max_age" yaml:"max_age"`               // 日志文件保留天数
	MaxBackups    int    `json:"max_backups" yaml:"max_backups"`       // 最大备份数量
	Compress      bool   `json:"compress" yaml:"compress"`             // 是否压缩旧日志

	// 高级配置
	ShowCaller      bool   `json:"show_caller" yaml:"show_caller"`           // 是否显示调用位置
	ShowTimestamp   bool   `json:"show_timestamp" yaml:"show_timestamp"`     // 是否显示时间戳
	TimestampFormat string `json:"timestamp_format" yaml:"timestamp_format"` // 时间戳格式

	// 字段配置
	Fields map[string]any `json:"fields" yaml:"fields"` // 全局字段
}

// DefaultLogConfig 返回默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatJSON,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/app.log",
		MaxSize:         100,
		MaxAge:          7,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields:          make(map[string]any),
	}
}

// CreateLogger 根据配置创建logrus logger
func (cfg *LogConfig) CreateLogger() *hertzlogrus.Logger {
	logger := hertzlogrus.NewLogger()
	logrusLogger := logger.Logger()

	// 设置日志级别
	switch cfg.Level {
	case LogLevelDebug:
		logrusLogger.SetLevel(logrus.DebugLevel)
	case LogLevelInfo:
		logrusLogger.SetLevel(logrus.InfoLevel)
	case LogLevelWarn:
		logrusLogger.SetLevel(logrus.WarnLevel)
	case LogLevelError:
		logrusLogger.SetLevel(logrus.ErrorLevel)
	case LogLevelFatal:
		logrusLogger.SetLevel(logrus.FatalLevel)
	case LogLevelPanic:
		logrusLogger.SetLevel(logrus.PanicLevel)
	default:
		logrusLogger.SetLevel(logrus.InfoLevel)
	}

	// 设置日志格式
	if cfg.Format == LogFormatJSON {
		formatter := &logrus.JSONFormatter{
			TimestampFormat: cfg.TimestampFormat,
		}
		logrusLogger.SetFormatter(formatter)
	} else {
		formatter := &logrus.TextFormatter{
			FullTimestamp:   cfg.ShowTimestamp,
			TimestampFormat: cfg.TimestampFormat,
			DisableColors:   false,
		}
		logrusLogger.SetFormatter(formatter)
	}

	// 设置是否显示调用位置
	logrusLogger.SetReportCaller(cfg.ShowCaller)

	// 配置输出
	var writers []io.Writer

	if cfg.EnableConsole {
		writers = append(writers, os.Stdout)
	}

	if cfg.EnableFile && cfg.FilePath != "" {
		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			logrusLogger.Errorf("Failed to create log directory: %v", err)
		} else {
			// 创建或打开日志文件
			file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				logrusLogger.Errorf("Failed to open log file: %v", err)
			} else {
				writers = append(writers, file)
			}
		}
	}

	if len(writers) > 0 {
		if len(writers) == 1 {
			logrusLogger.SetOutput(writers[0])
		} else {
			logrusLogger.SetOutput(io.MultiWriter(writers...))
		}
	}

	// 设置全局字段
	if len(cfg.Fields) > 0 {
		entry := logrusLogger.WithFields(logrus.Fields(cfg.Fields))
		// 注意：我们不能直接替换logger，而是通过hertz-logrus的方式添加字段
		// 这里暂时跳过全局字段设置，可以在使用时添加
		_ = entry
	}

	return logger
}

// LoggerWithRequestID 为logger添加请求ID
func LoggerWithRequestID(logger *logrus.Logger, requestID string) *logrus.Entry {
	return logger.WithField("request_id", requestID)
}

// LoggerWithUserID 为logger添加用户ID
func LoggerWithUserID(logger *logrus.Logger, userID string) *logrus.Entry {
	return logger.WithField("user_id", userID)
}

// LoggerWithFields 为logger添加多个字段
func LoggerWithFields(logger *logrus.Logger, fields map[string]any) *logrus.Entry {
	return logger.WithFields(logrus.Fields(fields))
}

// ============= 配置便捷方法 =============

// DevelopmentLogConfig 开发环境日志配置
func DevelopmentLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelDebug,
		Format:          LogFormatText,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/dev.log",
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
}

// ProductionLogConfig 生产环境日志配置
func ProductionLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatJSON,
		EnableConsole:   false,
		EnableFile:      true,
		FilePath:        "./logs/prod.log",
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
}

// TestLogConfig 测试环境日志配置
func TestLogConfig() *LogConfig {
	return &LogConfig{
		Level:         LogLevelWarn,
		Format:        LogFormatText,
		EnableConsole: true,
		EnableFile:    false,
		ShowCaller:    false,
		ShowTimestamp: false,
		Fields:        map[string]any{},
	}
}

// HighPerformanceLogConfig 高性能日志配置（最小日志）
func HighPerformanceLogConfig() *LogConfig {
	return &LogConfig{
		Level:         LogLevelError,
		Format:        LogFormatJSON,
		EnableConsole: false,
		EnableFile:    true,
		FilePath:      "./logs/error.log",
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
}

// UpdateConfigLevel 更新配置的日志级别
func (cfg *LogConfig) UpdateConfigLevel(level LogLevel) *LogConfig {
	newConfig := *cfg // 复制配置
	newConfig.Level = level
	return &newConfig
}

// UpdateConfigFormat 更新配置的日志格式
func (cfg *LogConfig) UpdateConfigFormat(format LogFormat) *LogConfig {
	newConfig := *cfg // 复制配置
	newConfig.Format = format
	return &newConfig
}

// AddConfigFields 向配置添加字段
func (cfg *LogConfig) AddConfigFields(fields map[string]any) *LogConfig {
	newConfig := *cfg // 复制配置
	if newConfig.Fields == nil {
		newConfig.Fields = make(map[string]any)
	}
	for k, v := range fields {
		newConfig.Fields[k] = v
	}
	return &newConfig
}
