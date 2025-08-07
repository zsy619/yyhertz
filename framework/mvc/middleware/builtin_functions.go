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