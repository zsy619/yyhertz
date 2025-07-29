package config

import (
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
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
}

// InitGlobalLogger 初始化全局日志实例
func InitGlobalLogger(config *LogConfig) *LoggerManager {
	loggerOnce.Do(func() {
		globalLogger = &LoggerManager{}
		globalLogger.updateLogger(config)
	})
	return globalLogger
}

// GetGlobalLogger 获取全局日志实例
func GetGlobalLogger() *LoggerManager {
	if globalLogger == nil {
		// 如果未初始化，使用默认配置初始化
		return InitGlobalLogger(DefaultLogConfig())
	}
	return globalLogger
}

// updateLogger 更新日志实例（内部方法）
func (lm *LoggerManager) updateLogger(config *LogConfig) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	lm.config = config
	lm.logger = config.CreateLogger()
	lm.rawLogger = lm.logger.Logger()

	// 设置为hertz的全局日志
	hlog.SetLogger(lm.logger)
}

// UpdateConfig 动态更新日志配置
func (lm *LoggerManager) UpdateConfig(config *LogConfig) {
	lm.updateLogger(config)
}

// UpdateLevel 动态更新日志级别
func (lm *LoggerManager) UpdateLevel(level LogLevel) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	newConfig := *lm.config
	newConfig.Level = level
	lm.updateLogger(&newConfig)
}

// GetLogger 获取hertz logger实例
func (lm *LoggerManager) GetLogger() *hertzlogrus.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return lm.logger
}

// GetRawLogger 获取原始logrus实例
func (lm *LoggerManager) GetRawLogger() *logrus.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return lm.rawLogger
}

// GetConfig 获取当前日志配置
func (lm *LoggerManager) GetConfig() *LogConfig {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return lm.config
}

// ============= 便捷日志方法 =============

// Debug 记录调试级别日志
func (lm *LoggerManager) Debug(args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Debug(args...)
}

// Debugf 格式化记录调试级别日志
func (lm *LoggerManager) Debugf(format string, args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Debugf(format, args...)
}

// Info 记录信息级别日志
func (lm *LoggerManager) Info(args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Info(args...)
}

// Infof 格式化记录信息级别日志
func (lm *LoggerManager) Infof(format string, args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Infof(format, args...)
}

// Warn 记录警告级别日志
func (lm *LoggerManager) Warn(args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Warn(args...)
}

// Warnf 格式化记录警告级别日志
func (lm *LoggerManager) Warnf(format string, args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Warnf(format, args...)
}

// Error 记录错误级别日志
func (lm *LoggerManager) Error(args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Error(args...)
}

// Errorf 格式化记录错误级别日志
func (lm *LoggerManager) Errorf(format string, args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Errorf(format, args...)
}

// Fatal 记录致命错误级别日志
func (lm *LoggerManager) Fatal(args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Fatal(args...)
}

// Fatalf 格式化记录致命错误级别日志
func (lm *LoggerManager) Fatalf(format string, args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Fatalf(format, args...)
}

// Panic 记录恐慌级别日志
func (lm *LoggerManager) Panic(args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Panic(args...)
}

// Panicf 格式化记录恐慌级别日志
func (lm *LoggerManager) Panicf(format string, args ...any) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	lm.rawLogger.Panicf(format, args...)
}

// ============= 结构化日志方法 =============

// WithFields 添加字段
func (lm *LoggerManager) WithFields(fields map[string]any) *logrus.Entry {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return lm.rawLogger.WithFields(logrus.Fields(fields))
}

// WithField 添加单个字段
func (lm *LoggerManager) WithField(key string, value any) *logrus.Entry {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	return lm.rawLogger.WithField(key, value)
}

// WithRequestID 添加请求ID字段
func (lm *LoggerManager) WithRequestID(requestID string) *logrus.Entry {
	return lm.WithField("request_id", requestID)
}

// WithUserID 添加用户ID字段
func (lm *LoggerManager) WithUserID(userID string) *logrus.Entry {
	return lm.WithField("user_id", userID)
}

// WithContext 添加上下文字段
func (lm *LoggerManager) WithContext(ctx map[string]any) *logrus.Entry {
	return lm.WithFields(ctx)
}

// ============= 全局快捷方法 =============

// Debug 全局调试日志
func Debug(args ...any) {
	GetGlobalLogger().Debug(args...)
}

// Debugf 全局格式化调试日志
func Debugf(format string, args ...any) {
	GetGlobalLogger().Debugf(format, args...)
}

// Info 全局信息日志
func Info(args ...any) {
	GetGlobalLogger().Info(args...)
}

// Infof 全局格式化信息日志
func Infof(format string, args ...any) {
	GetGlobalLogger().Infof(format, args...)
}

// Warn 全局警告日志
func Warn(args ...any) {
	GetGlobalLogger().Warn(args...)
}

// Warnf 全局格式化警告日志
func Warnf(format string, args ...any) {
	GetGlobalLogger().Warnf(format, args...)
}

// Error 全局错误日志
func Error(args ...any) {
	GetGlobalLogger().Error(args...)
}

// Errorf 全局格式化错误日志
func Errorf(format string, args ...any) {
	GetGlobalLogger().Errorf(format, args...)
}

// Fatal 全局致命错误日志
func Fatal(args ...any) {
	GetGlobalLogger().Fatal(args...)
}

// Fatalf 全局格式化致命错误日志
func Fatalf(format string, args ...any) {
	GetGlobalLogger().Fatalf(format, args...)
}

// Panic 全局恐慌日志
func Panic(args ...any) {
	GetGlobalLogger().Panic(args...)
}

// Panicf 全局格式化恐慌日志
func Panicf(format string, args ...any) {
	GetGlobalLogger().Panicf(format, args...)
}

// WithFields 全局添加字段
func WithFields(fields map[string]any) *logrus.Entry {
	return GetGlobalLogger().WithFields(fields)
}

// WithField 全局添加单个字段
func WithField(key string, value any) *logrus.Entry {
	return GetGlobalLogger().WithField(key, value)
}

// WithRequestID 全局添加请求ID
func WithRequestID(requestID string) *logrus.Entry {
	return GetGlobalLogger().WithRequestID(requestID)
}

// WithUserID 全局添加用户ID
func WithUserID(userID string) *logrus.Entry {
	return GetGlobalLogger().WithUserID(userID)
}
