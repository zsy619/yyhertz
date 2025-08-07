package middleware

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// Middleware 中间件函数类型定义
type Middleware func(c context.Context, ctx *app.RequestContext)

// LoggerConfig Logger中间件配置
type LoggerConfig struct {
	// 输出目标
	Output io.Writer

	// 是否跳过路径
	SkipPaths []string

	// 时间格式
	TimeFormat string

	// 自定义格式化函数
	Formatter LogFormatter
}

// LogFormatter 日志格式化函数
type LogFormatter func(param LogFormatterParams) string

// LogFormatterParams 日志格式化参数
type LogFormatterParams struct {
	Request      *Context
	TimeStamp    time.Time
	StatusCode   int
	Latency      time.Duration
	ClientIP     string
	Method       string
	Path         string
	ErrorMessage string
}

// DefaultBuiltinLoggerConfig 默认内置Logger配置
func DefaultBuiltinLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Output:     os.Stdout,
		TimeFormat: time.RFC3339,
		Formatter:  nil, // 将在builtin.go中定义默认格式化器
	}
}