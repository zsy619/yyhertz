// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zsy619/yyhertz/framework/config"
)

// LogLevel 日志级别
type LogLevel string

const (
	// LogLevelSilent 静默级别
	LogLevelSilent LogLevel = "silent"
	// LogLevelError 错误级别
	LogLevelError LogLevel = "error"
	// LogLevelWarn 警告级别
	LogLevelWarn LogLevel = "warn"
	// LogLevelInfo 信息级别
	LogLevelInfo LogLevel = "info"
	// LogLevelDebug 调试级别
	LogLevelDebug LogLevel = "debug"
	// LogLevelTrace 跟踪级别
	LogLevelTrace LogLevel = "trace"
)

// EnhancedLogger 增强的日志记录器
type EnhancedLogger struct {
	// 日志级别
	LogLevel LogLevel
	// 慢查询阈值
	SlowThreshold time.Duration
	// 是否忽略记录未找到错误
	IgnoreRecordNotFoundError bool
	// 是否记录参数值
	LogValues bool
	// 是否记录调用栈
	LogCallStack bool
	// 是否记录SQL语句
	LogSQL bool
	// 是否记录行数
	LogRows bool
	// 是否记录执行时间
	LogTime bool
	// 是否记录事务
	LogTransaction bool
	// 是否记录颜色
	Colorful bool
}

// NewEnhancedLogger 创建增强的日志记录器
func NewEnhancedLogger(level LogLevel, slowThreshold time.Duration) *EnhancedLogger {
	return &EnhancedLogger{
		LogLevel:                  level,
		SlowThreshold:             slowThreshold,
		IgnoreRecordNotFoundError: true,
		LogValues:                 true,
		LogCallStack:              true,
		LogSQL:                    true,
		LogRows:                   true,
		LogTime:                   true,
		LogTransaction:            true,
		Colorful:                  false,
	}
}

// LogMode 设置日志级别
func (l *EnhancedLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	switch level {
	case logger.Silent:
		newLogger.LogLevel = LogLevelSilent
	case logger.Error:
		newLogger.LogLevel = LogLevelError
	case logger.Warn:
		newLogger.LogLevel = LogLevelWarn
	case logger.Info:
		newLogger.LogLevel = LogLevelInfo
	}
	return &newLogger
}

// Info 记录信息日志
func (l *EnhancedLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel == LogLevelSilent {
		return
	}
	if l.LogLevel == LogLevelError || l.LogLevel == LogLevelWarn {
		return
	}
	config.Infof(msg, data...)
}

// Warn 记录警告日志
func (l *EnhancedLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel == LogLevelSilent {
		return
	}
	if l.LogLevel == LogLevelError {
		return
	}
	config.Warnf(msg, data...)
}

// Error 记录错误日志
func (l *EnhancedLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel == LogLevelSilent {
		return
	}
	config.Errorf(msg, data...)
}

// Trace 记录跟踪日志
func (l *EnhancedLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel == LogLevelSilent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 获取事务信息
	txID := ""
	if tx, ok := ctx.Value("tx_id").(string); ok {
		txID = tx
	}

	// 构建日志消息
	logMsg := ""
	if txID != "" && l.LogTransaction {
		logMsg += fmt.Sprintf("[TX:%s] ", txID)
	}

	if l.LogSQL {
		logMsg += sql
	}

	if l.LogRows {
		logMsg += fmt.Sprintf(" [%d rows affected]", rows)
	}

	if l.LogTime {
		logMsg += fmt.Sprintf(" [%.3fms]", float64(elapsed.Nanoseconds())/1e6)
	}

	// 根据不同情况记录日志
	switch {
	case err != nil && l.LogLevel >= LogLevelError && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		// 错误日志
		l.Error(ctx, "%s\n[ERROR] %v", logMsg, err)
	case elapsed > l.SlowThreshold && l.SlowThreshold > 0 && l.LogLevel >= LogLevelWarn:
		// 慢查询日志
		l.Warn(ctx, "%s\n[SLOW SQL] 超过 %v", logMsg, l.SlowThreshold)
	case l.LogLevel >= LogLevelInfo:
		// 普通查询日志
		l.Info(ctx, "%s", logMsg)
	case l.LogLevel >= LogLevelDebug:
		// 调试日志，包含更多信息
		l.Debug(ctx, "%s\n[DEBUG] %s", logMsg, getCallStack())
	case l.LogLevel >= LogLevelTrace:
		// 跟踪日志，包含完整调用栈
		config.Debugf("%s\n[TRACE] %s", logMsg, getFullCallStack())
	}
}

// Debug 记录调试日志
func (l *EnhancedLogger) Debug(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel == LogLevelSilent || l.LogLevel == LogLevelError || l.LogLevel == LogLevelWarn || l.LogLevel == LogLevelInfo {
		return
	}
	config.Debugf(msg, data...)
}

// getFullCallStack 获取完整调用栈
func getFullCallStack() string {
	var stack []string

	// 跳过当前函数和调用者函数
	for i := 2; i < 20; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		funcName := fn.Name()
		// 过滤掉runtime相关的函数
		if strings.Contains(funcName, "runtime.") ||
			strings.Contains(funcName, "syscall.") ||
			strings.Contains(funcName, "reflect.") {
			continue
		}

		// 保留完整的文件路径和函数名用于详细调试
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, funcName))
	}

	if len(stack) == 0 {
		return "无完整调用栈信息"
	}

	return strings.Join(stack, "\n  ")
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	// 日志级别
	Level LogLevel `json:"level" yaml:"level"`
	// 慢查询阈值（毫秒）
	SlowThreshold int `json:"slow_threshold" yaml:"slow_threshold"`
	// 是否忽略记录未找到错误
	IgnoreRecordNotFoundError bool `json:"ignore_record_not_found_error" yaml:"ignore_record_not_found_error"`
	// 是否记录参数值
	LogValues bool `json:"log_values" yaml:"log_values"`
	// 是否记录调用栈
	LogCallStack bool `json:"log_call_stack" yaml:"log_call_stack"`
	// 是否记录SQL语句
	LogSQL bool `json:"log_sql" yaml:"log_sql"`
	// 是否记录行数
	LogRows bool `json:"log_rows" yaml:"log_rows"`
	// 是否记录执行时间
	LogTime bool `json:"log_time" yaml:"log_time"`
	// 是否记录事务
	LogTransaction bool `json:"log_transaction" yaml:"log_transaction"`
	// 是否记录颜色
	Colorful bool `json:"colorful" yaml:"colorful"`
}

// DefaultLoggerConfig 默认日志配置
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:                     LogLevelInfo,
		SlowThreshold:             1000,
		IgnoreRecordNotFoundError: true,
		LogValues:                 true,
		LogCallStack:              true,
		LogSQL:                    true,
		LogRows:                   true,
		LogTime:                   true,
		LogTransaction:            true,
		Colorful:                  false,
	}
}

// NewLoggerFromConfig 从配置创建日志记录器
func NewLoggerFromConfig(config *LoggerConfig) *EnhancedLogger {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	logger := NewEnhancedLogger(config.Level, time.Duration(config.SlowThreshold)*time.Millisecond)
	logger.IgnoreRecordNotFoundError = config.IgnoreRecordNotFoundError
	logger.LogValues = config.LogValues
	logger.LogCallStack = config.LogCallStack
	logger.LogSQL = config.LogSQL
	logger.LogRows = config.LogRows
	logger.LogTime = config.LogTime
	logger.LogTransaction = config.LogTransaction
	logger.Colorful = config.Colorful

	return logger
}

// ToGormLogger 转换为GORM日志记录器
func (l *EnhancedLogger) ToGormLogger() logger.Interface {
	var level logger.LogLevel
	switch l.LogLevel {
	case LogLevelSilent:
		level = logger.Silent
	case LogLevelError:
		level = logger.Error
	case LogLevelWarn:
		level = logger.Warn
	case LogLevelInfo, LogLevelDebug, LogLevelTrace:
		level = logger.Info
	default:
		level = logger.Info
	}

	return logger.New(
		&enhancedLogWriter{logger: l},
		logger.Config{
			SlowThreshold:             l.SlowThreshold,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
			Colorful:                  l.Colorful,
		},
	)
}

// enhancedLogWriter 增强的日志写入器
type enhancedLogWriter struct {
	logger *EnhancedLogger
}

// Printf 实现logger.Writer接口
func (w *enhancedLogWriter) Printf(format string, args ...interface{}) {
	config.Infof(format, args...)
}

// 全局日志记录器
var (
	globalLogger *EnhancedLogger
	loggerOnce   sync.Once
)

// GetGlobalLogger 获取全局日志记录器
func GetGlobalLogger() *EnhancedLogger {
	loggerOnce.Do(func() {
		globalLogger = NewEnhancedLogger(LogLevelInfo, time.Second)
	})
	return globalLogger
}

// SetGlobalLogger 设置全局日志记录器
func SetGlobalLogger(logger *EnhancedLogger) {
	globalLogger = logger
}
