package config

import (
	"io"
	"sync"

	hertzlogrus "github.com/hertz-contrib/logger/logrus"
	"github.com/sirupsen/logrus"
)

var (
	// 全局日志实例
	globalLogger *LoggerManager
	// 初始化锁
	loggerOnce sync.Once
	// 运行时锁
	loggerMutex sync.RWMutex
)

// LoggerManager 全局日志管理器
type LoggerManager struct {
	logger    *hertzlogrus.Logger
	config    *LogConfig
	rawLogger *logrus.Logger
	writers   []OutputWriter
}

// InitGlobalLogger 初始化全局日志实例
func InitGlobalLogger(config *LogConfig) *LoggerManager {
	loggerOnce.Do(func() {
		globalLogger = &LoggerManager{}
		globalLogger.updateLogger(config)
	})
	return globalLogger
}

// ResetGlobalLogger 重置全局日志实例（主要用于测试）
func ResetGlobalLogger(config *LogConfig) *LoggerManager {
	loggerMutex.Lock()
	
	// 关闭旧的全局日志器（在持有锁的情况下直接操作，避免调用Close()方法再次获取锁）
	if globalLogger != nil && globalLogger.writers != nil {
		for _, writer := range globalLogger.writers {
			if writer != nil {
				writer.Close()
			}
		}
		globalLogger.writers = nil
	}
	
	// 重置 sync.Once
	loggerOnce = sync.Once{}
	globalLogger = nil
	
	loggerMutex.Unlock()
	
	// 重新初始化
	return InitGlobalLogger(config)
}

// GetGlobalLogger 获取全局日志实例
func GetGlobalLogger() *LoggerManager {
	if globalLogger == nil {
		// 如果未初始化，使用默认配置初始化
		logConfig, err := GetLogConfig()
		if err != nil {
			logConfig = DefaultLogConfig()
		}
		return InitGlobalLogger(logConfig)
	}
	return globalLogger
}

// updateLogger 更新日志实例（内部方法）
func (lm *LoggerManager) updateLogger(config *LogConfig) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// 关闭旧的写入器
	if lm.writers != nil {
		for _, writer := range lm.writers {
			if writer != nil {
				writer.Close()
			}
		}
	}

	// 创建新的Hertz日志器
	lm.logger = hertzlogrus.NewLogger()
	lm.rawLogger = lm.logger.Logger()
	lm.config = config

	// 设置日志级别
	switch config.Level {
	case LogLevelDebug:
		lm.rawLogger.SetLevel(logrus.DebugLevel)
	case LogLevelInfo:
		lm.rawLogger.SetLevel(logrus.InfoLevel)
	case LogLevelWarn:
		lm.rawLogger.SetLevel(logrus.WarnLevel)
	case LogLevelError:
		lm.rawLogger.SetLevel(logrus.ErrorLevel)
	case LogLevelFatal:
		lm.rawLogger.SetLevel(logrus.FatalLevel)
	case LogLevelPanic:
		lm.rawLogger.SetLevel(logrus.PanicLevel)
	default:
		lm.rawLogger.SetLevel(logrus.InfoLevel)
	}

	// 设置格式化器
	formatter := GetFormatter(config.Format, config)
	lm.rawLogger.SetFormatter(formatter)

	// 设置调用位置显示
	lm.rawLogger.SetReportCaller(config.ShowCaller)

	// 创建输出写入器
	writers, err := CreateOutputWriters(config)
	if err != nil {
		lm.rawLogger.Errorf("创建输出写入器失败: %v", err)
		writers = []OutputWriter{NewConsoleWriter()} // 降级到控制台输出
	}
	lm.writers = writers

	// 设置输出
	if len(writers) == 1 {
		lm.rawLogger.SetOutput(writers[0])
	} else if len(writers) > 1 {
		multiWriter := NewMultiWriter(writers...)
		lm.rawLogger.SetOutput(multiWriter)
	}

	// 设置钩子
	if err := SetupLoggerHooks(lm.rawLogger, config); err != nil {
		lm.rawLogger.Errorf("设置日志钩子失败: %v", err)
	}

	// 设置全局字段
	// 注意：这里不能直接修改rawLogger，因为WithFields返回的是Entry而不是Logger
	// 全局字段将在实际使用时通过GetRawLogger().WithFields()添加
}

// UpdateConfig 更新日志配置
func (lm *LoggerManager) UpdateConfig(config *LogConfig) {
	lm.updateLogger(config)
}

// UpdateLevel 动态更新日志级别
func (lm *LoggerManager) UpdateLevel(level LogLevel) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// 更新配置中的级别
	if lm.config != nil {
		lm.config.Level = level
	}

	// 更新logrus logger的级别
	if lm.rawLogger != nil {
		switch level {
		case LogLevelDebug:
			lm.rawLogger.SetLevel(logrus.DebugLevel)
		case LogLevelInfo:
			lm.rawLogger.SetLevel(logrus.InfoLevel)
		case LogLevelWarn:
			lm.rawLogger.SetLevel(logrus.WarnLevel)
		case LogLevelError:
			lm.rawLogger.SetLevel(logrus.ErrorLevel)
		case LogLevelFatal:
			lm.rawLogger.SetLevel(logrus.FatalLevel)
		case LogLevelPanic:
			lm.rawLogger.SetLevel(logrus.PanicLevel)
		default:
			lm.rawLogger.SetLevel(logrus.InfoLevel)
		}
	}
}

// UpdateFormat 动态更新日志格式
func (lm *LoggerManager) UpdateFormat(format LogFormat) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// 更新配置中的格式
	if lm.config != nil {
		lm.config.Format = format
	}

	// 更新logrus logger的格式化器
	if lm.rawLogger != nil {
		formatter := GetFormatter(format, lm.config)
		lm.rawLogger.SetFormatter(formatter)
	}
}

// GetLevel 获取当前日志级别
func (lm *LoggerManager) GetLevel() LogLevel {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	
	if lm.config != nil {
		return lm.config.Level
	}
	return LogLevelInfo // 默认级别
}

// GetFormat 获取当前日志格式
func (lm *LoggerManager) GetFormat() LogFormat {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	
	if lm.config != nil {
		return lm.config.Format
	}
	return LogFormatBeego // 默认格式
}

// GetLogger 获取Hertz日志器
func (lm *LoggerManager) GetLogger() *hertzlogrus.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return lm.logger
}

// GetRawLogger 获取原始logrus日志器
func (lm *LoggerManager) GetRawLogger() *logrus.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return lm.rawLogger
}

// GetConfig 获取当前配置
func (lm *LoggerManager) GetConfig() *LogConfig {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return lm.config
}

// Close 关闭日志管理器
func (lm *LoggerManager) Close() error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if lm.writers != nil {
		for _, writer := range lm.writers {
			if writer != nil {
				writer.Close()
			}
		}
		lm.writers = nil
	}
	return nil
}

// WithFields 添加字段
func (lm *LoggerManager) WithFields(fields map[string]any) *logrus.Entry {
	return lm.rawLogger.WithFields(logrus.Fields(fields))
}

// WithField 添加单个字段
func (lm *LoggerManager) WithField(key string, value any) *logrus.Entry {
	return lm.rawLogger.WithField(key, value)
}

// WithError 添加错误字段
func (lm *LoggerManager) WithError(err error) *logrus.Entry {
	return lm.rawLogger.WithError(err)
}

// Debug 调试日志
func (lm *LoggerManager) Debug(args ...any) {
	lm.rawLogger.Debug(args...)
}

// Debugf 格式化调试日志
func (lm *LoggerManager) Debugf(format string, args ...any) {
	lm.rawLogger.Debugf(format, args...)
}

// Info 信息日志
func (lm *LoggerManager) Info(args ...any) {
	lm.rawLogger.Info(args...)
}

// Infof 格式化信息日志
func (lm *LoggerManager) Infof(format string, args ...any) {
	lm.rawLogger.Infof(format, args...)
}

// Warn 警告日志
func (lm *LoggerManager) Warn(args ...any) {
	lm.rawLogger.Warn(args...)
}

// Warnf 格式化警告日志
func (lm *LoggerManager) Warnf(format string, args ...any) {
	lm.rawLogger.Warnf(format, args...)
}

// Error 错误日志
func (lm *LoggerManager) Error(args ...any) {
	lm.rawLogger.Error(args...)
}

// Errorf 格式化错误日志
func (lm *LoggerManager) Errorf(format string, args ...any) {
	lm.rawLogger.Errorf(format, args...)
}

// Fatal 致命错误日志
func (lm *LoggerManager) Fatal(args ...any) {
	lm.rawLogger.Fatal(args...)
}

// Fatalf 格式化致命错误日志
func (lm *LoggerManager) Fatalf(format string, args ...any) {
	lm.rawLogger.Fatalf(format, args...)
}

// Panic panic日志
func (lm *LoggerManager) Panic(args ...any) {
	lm.rawLogger.Panic(args...)
}

// Panicf 格式化panic日志
func (lm *LoggerManager) Panicf(format string, args ...any) {
	lm.rawLogger.Panicf(format, args...)
}

// CreateLogger 根据配置创建logrus logger（保留原有方法以兼容）
func (cfg *LogConfig) CreateLogger() *hertzlogrus.Logger {
	// 创建Hertz日志器
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

	// 设置格式化器
	formatter := GetFormatter(cfg.Format, cfg)
	logrusLogger.SetFormatter(formatter)

	// 设置是否显示调用位置
	logrusLogger.SetReportCaller(cfg.ShowCaller)

	// 创建输出写入器
	writers, err := CreateOutputWriters(cfg)
	if err != nil {
		logrusLogger.Errorf("创建输出写入器失败: %v", err)
		writers = []OutputWriter{NewConsoleWriter()} // 降级到控制台输出
	}

	// 设置输出
	if len(writers) == 1 {
		logrusLogger.SetOutput(writers[0])
	} else if len(writers) > 1 {
		multiWriter := NewMultiWriter(writers...)
		logrusLogger.SetOutput(multiWriter)
	} else {
		// 如果没有写入器，设置为丢弃输出
		logrusLogger.SetOutput(io.Discard)
	}

	// 设置钩子
	if err := SetupLoggerHooks(logrusLogger, cfg); err != nil {
		logrusLogger.Errorf("设置日志钩子失败: %v", err)
	}

	return logger
}