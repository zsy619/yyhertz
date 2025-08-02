// Package middleware 内置中间件集合
// 提供常用的中间件实现，类似Gin的内置中间件
package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

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
		Formatter:  defaultLogFormatter,
	}
}

// Logger 日志中间件
func Logger() HandlerFunc {
	return LoggerWithConfig(DefaultBuiltinLoggerConfig())
}

// LoggerWithConfig 使用配置的Logger中间件
func LoggerWithConfig(conf LoggerConfig) HandlerFunc {
	notlogged := conf.SkipPaths

	out := conf.Output
	if out == nil {
		out = os.Stdout
	}

	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}

	return func(c *Context) {
		// 检查是否跳过
		path := c.URI().Path()
		for _, skip := range notlogged {
			if strings.Contains(string(path), skip) {
				c.Next()
				return
			}
		}

		// 记录开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 获取状态码
		statusCode := c.Response.StatusCode()

		// 获取错误信息
		errorMessage := ""
		if len(c.Errors) > 0 {
			errorMessage = c.Errors[len(c.Errors)-1].Error()
		}

		// 格式化日志
		params := LogFormatterParams{
			Request:      c,
			TimeStamp:    time.Now(),
			StatusCode:   statusCode,
			Latency:      latency,
			ClientIP:     c.ClientIP(),
			Method:       string(c.Method()),
			Path:         string(path),
			ErrorMessage: errorMessage,
		}

		fmt.Fprint(out, formatter(params))
	}
}

// defaultLogFormatter 默认日志格式化函数
func defaultLogFormatter(param LogFormatterParams) string {
	var statusColor, methodColor, resetColor string

	statusColor = getColorByStatus(param.StatusCode)
	methodColor = getColorByMethod(param.Method)
	resetColor = "\033[0m"

	return fmt.Sprintf("%s[YYHertz]%s %v |%s %3d %s| %13v | %15s |%s %-7s %s %s",
		"\033[90m", resetColor, param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
	)
}

// getColorByStatus 根据状态码获取颜色
func getColorByStatus(code int) string {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return "\033[97;42m" // green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return "\033[90;47m" // white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return "\033[90;43m" // yellow
	default:
		return "\033[97;41m" // red
	}
}

// getColorByMethod 根据HTTP方法获取颜色
func getColorByMethod(method string) string {
	switch method {
	case "GET":
		return "\033[97;44m" // blue
	case "POST":
		return "\033[97;42m" // green
	case "PUT":
		return "\033[97;43m" // yellow
	case "DELETE":
		return "\033[97;41m" // red
	case "PATCH":
		return "\033[97;42m" // green
	case "HEAD":
		return "\033[97;45m" // magenta
	case "OPTIONS":
		return "\033[90;47m" // white
	default:
		return "\033[0m" // reset
	}
}

// Recovery 恢复中间件
func Recovery() HandlerFunc {
	return RecoveryWithWriter(os.Stderr)
}

// RecoveryWithWriter 使用指定Writer的恢复中间件
func RecoveryWithWriter(out io.Writer) HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)

				// 记录错误
				fmt.Fprintf(out, "[Recovery] %s panic recovered:\n%s\n%s\n",
					time.Now().Format("2006/01/02 - 15:04:05"), err, stack[:length])

				// 添加到错误列表
				c.AddError(fmt.Errorf("panic recovered: %v", err))

				// 返回500错误
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// NoRoute 处理404路由
func NoRoute() HandlerFunc {
	return func(c *Context) {
		c.SetStatusCode(http.StatusNotFound)
		c.SetContentType("application/json")
		c.SetBodyString(`{"error":"route not found","path":"` + string(c.URI().Path()) + `"}`)
	}
}

// NoMethod 处理405方法不允许
func NoMethod() HandlerFunc {
	return func(c *Context) {
		c.SetStatusCode(http.StatusMethodNotAllowed)
		c.SetContentType("application/json")
		c.SetBodyString(`{"error":"method not allowed","method":"` + string(c.Method()) + `"}`)
	}
}

// BasicAuth 基础认证中间件
func BasicAuth(accounts map[string]string) HandlerFunc {
	return func(c *Context) {
		user, password, hasAuth := c.Request.BasicAuth()
		if hasAuth {
			if expectedPassword, ok := accounts[user]; ok && expectedPassword == password {
				c.Set("user", user)
				c.Next()
				return
			}
		}

		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// RequestID 生成请求ID中间件
func RequestID() HandlerFunc {
	return func(c *Context) {
		requestID := generateRequestID()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) HandlerFunc {
	return func(c *Context) {
		finish := make(chan struct{})
		panicChan := make(chan any, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			finish <- struct{}{}
		}()

		select {
		case p := <-panicChan:
			panic(p)
		case <-finish:
			return
		case <-time.After(timeout):
			c.AbortWithStatus(http.StatusRequestTimeout)
			return
		}
	}
}

// Secure 安全头中间件
func Secure() HandlerFunc {
	return func(c *Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000")
		c.Next()
	}
}

// GZip 压缩中间件
func GZip() HandlerFunc {
	return func(c *Context) {
		// 检查客户端是否支持gzip
		acceptEncoding := string(c.Request.Header.Peek("Accept-Encoding"))
		if strings.Contains(acceptEncoding, "gzip") {
			c.Header("Content-Encoding", "gzip")
		}
		c.Next()
	}
}
