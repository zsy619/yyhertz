package middleware

import (
	"io"
	"os"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// Middleware 中间件函数类型定义
type Middleware = define.Middleware

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

// LogFormatter 是一个函数类型，用于定义日志的格式化逻辑。
// 它接收一个 LogFormatterParams 参数，返回格式化后的日志字符串。
// 通常用于自定义日志输出的格式，例如时间戳、日志级别、消息内容等。
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

// DefaultBuiltinLoggerConfig 返回一个默认的内置日志记录器配置。
// 该配置包括标准输出作为日志输出目标，使用 RFC3339 时间格式，
// 并且未设置格式化器（格式化器将在 builtin.go 中定义默认实现）。
// 适用于需要快速初始化日志记录器的场景。
func DefaultBuiltinLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Output:     os.Stdout,
		TimeFormat: time.RFC3339,
		Formatter:  nil, // 将在builtin.go中定义默认格式化器
	}
}
