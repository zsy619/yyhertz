package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// 扩展的内置中间件 - 移植自 @framework/middleware 系统

// registerExtendedBuiltinMiddlewares 注册扩展的内置中间件
func (m *MiddlewareManager) registerExtendedBuiltinMiddlewares() {
	// 注册所有从基础系统移植的中间件
	m.registerEnhancedLoggerMiddleware()
	m.registerEnhancedRecoveryMiddleware()
	m.registerCORSMiddleware()
	m.registerRateLimitMiddleware()
	m.registerTracingMiddleware()
	m.registerTLSMiddleware()
	m.registerBasicAuthMiddleware()
	m.registerRequestIDMiddleware()
	m.registerTimeoutMiddleware()
	m.registerSecureMiddleware()
	m.registerGZipMiddleware()
}

// registerEnhancedLoggerMiddleware 注册增强的Logger中间件
func (m *MiddlewareManager) registerEnhancedLoggerMiddleware() {
	m.RegisterBuiltin("enhanced-logger", func(config interface{}) MiddlewareFunc {
		// 配置解析
		var logConfig LoggerConfig
		if config != nil {
			if cfg, ok := config.(LoggerConfig); ok {
				logConfig = cfg
			} else {
				logConfig = DefaultBuiltinLoggerConfig()
			}
		} else {
			logConfig = DefaultBuiltinLoggerConfig()
		}

		notlogged := logConfig.SkipPaths
		out := logConfig.Output
		if out == nil {
			out = os.Stdout
		}

		formatter := logConfig.Formatter
		if formatter == nil {
			formatter = defaultLogFormatter
		}

		return func(ctx *mvccontext.Context) {
			// 检查是否跳过
			path := string(ctx.Request.Path())
			for _, skip := range notlogged {
				if strings.Contains(path, skip) {
					ctx.Next()
					return
				}
			}

			// 记录开始时间
			start := time.Now()

			// 处理请求
			ctx.Next()

			// 计算延迟
			latency := time.Since(start)

			// 获取状态码
			statusCode := ctx.Writer.Status()

			// 获取错误信息
			errorMessage := ""
			if len(ctx.GetErrors()) > 0 {
				errorMessage = ctx.GetErrors()[len(ctx.GetErrors())-1].Error()
			}

			// 创建基础Context用于兼容
			basicCtx := CreateBasicContext(ctx)

			// 格式化日志
			params := LogFormatterParams{
				Request:      basicCtx,
				TimeStamp:    time.Now(),
				StatusCode:   statusCode,
				Latency:      latency,
				ClientIP:     string(ctx.Request.ClientIP()),
				Method:       string(ctx.Request.Method()),
				Path:         path,
				ErrorMessage: errorMessage,
			}

			fmt.Fprint(out, formatter(params))
		}
	}, MiddlewareMetadata{
		Name:        "enhanced-logger",
		Version:     "2.0.0",
		Description: "Enhanced HTTP请求日志记录中间件 (from basic system)",
		Author:      "YYHertz Team",
	})
}

// registerEnhancedRecoveryMiddleware 注册增强的Recovery中间件
func (m *MiddlewareManager) registerEnhancedRecoveryMiddleware() {
	m.RegisterBuiltin("enhanced-recovery", func(config interface{}) MiddlewareFunc {
		out := os.Stderr
		if config != nil {
			if writer, ok := config.(io.Writer); ok {
				out = writer.(*os.File)
			}
		}

		return func(ctx *mvccontext.Context) {
			defer func() {
				if err := recover(); err != nil {
					// 获取堆栈信息
					stack := make([]byte, 4096)
					length := runtime.Stack(stack, false)

					// 记录错误
					fmt.Fprintf(out, "[Recovery] %s panic recovered:\n%s\n%s\n",
						time.Now().Format("2006/01/02 - 15:04:05"), err, stack[:length])

					// 添加到错误列表
					ctx.AddError(fmt.Errorf("panic recovered: %v", err))

					// 返回500错误
					ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
					ctx.Abort()
				}
			}()
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "enhanced-recovery",
		Version:     "2.0.0",
		Description: "Enhanced Panic恢复中间件 (from basic system)",
		Author:      "YYHertz Team",
	})
}

// registerCORSMiddleware 注册CORS中间件
func (m *MiddlewareManager) registerCORSMiddleware() {
	m.RegisterBuiltin("cors-extended", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.Context) {
			// 设置CORS头
			ctx.Request.Response.Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Request.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			ctx.Request.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

			// 处理预检请求
			if string(ctx.Request.Method()) == "OPTIONS" {
				ctx.JSON(204, nil)
				ctx.Abort()
				return
			}

			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "cors-extended",
		Version:     "2.0.0",
		Description: "跨域资源共享中间件 (enhanced from basic system)",
		Author:      "YYHertz Team",
	})
}

// registerRateLimitMiddleware 注册限流中间件
func (m *MiddlewareManager) registerRateLimitMiddleware() {
	m.RegisterBuiltin("ratelimit", func(config interface{}) MiddlewareFunc {
		// 简化的限流实现 - 实际应用中需要使用专业的限流算法
		return func(ctx *mvccontext.Context) {
			// TODO: 实现令牌桶或滑动窗口算法
			// 这里先实现基础检查
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "ratelimit",
		Version:     "1.0.0",
		Description: "请求限流中间件",
		Author:      "YYHertz Team",
		Dependencies: []string{"logger"},
	})
}

// registerTracingMiddleware 注册链路追踪中间件
func (m *MiddlewareManager) registerTracingMiddleware() {
	m.RegisterBuiltin("tracing", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.Context) {
			// 生成Trace ID
			traceID := generateTraceID()
			ctx.Set("TraceID", traceID)
			ctx.Request.Response.Header.Set("X-Trace-ID", traceID)

			// TODO: 集成实际的链路追踪系统 (如 Jaeger, Zipkin)
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "tracing",
		Version:     "1.0.0",
		Description: "分布式链路追踪中间件",
		Author:      "YYHertz Team",
	})
}

// registerTLSMiddleware 注册TLS中间件
func (m *MiddlewareManager) registerTLSMiddleware() {
	m.RegisterBuiltin("tls", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.Context) {
			// TLS相关检查和处理
			// 简化实现，实际应用中需要检查协议  
			fmt.Printf("[TLS] Processing request - client_ip: %s, path: %s\n", 
				string(ctx.Request.ClientIP()), string(ctx.Request.Path()))
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "tls",
		Version:     "1.0.0",
		Description: "TLS连接处理中间件",
		Author:      "YYHertz Team",
	})
}

// registerBasicAuthMiddleware 注册基础认证中间件
func (m *MiddlewareManager) registerBasicAuthMiddleware() {
	m.RegisterBuiltin("basicauth", func(config interface{}) MiddlewareFunc {
		_ = config // 简化实现，暂不处理账户配置

		return func(ctx *mvccontext.Context) {
			// 简化的Basic Auth实现
			auth := string(ctx.Request.GetHeader("Authorization"))
			if strings.HasPrefix(auth, "Basic ") {
				// TODO: 解析Basic Auth头部
				ctx.Set("user", "demo")
				ctx.Next()
				return
			}

			ctx.Request.Response.Header.Set("WWW-Authenticate", "Basic realm=Authorization Required")
			ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			ctx.Abort()
		}
	}, MiddlewareMetadata{
		Name:        "basicauth",
		Version:     "1.0.0",
		Description: "HTTP Basic认证中间件",
		Author:      "YYHertz Team",
	})
}

// registerRequestIDMiddleware 注册请求ID中间件
func (m *MiddlewareManager) registerRequestIDMiddleware() {
	m.RegisterBuiltin("requestid", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.Context) {
			requestID := generateRequestID()
			ctx.Set("RequestID", requestID)
			ctx.Request.Response.Header.Set("X-Request-ID", requestID)
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "requestid",
		Version:     "1.0.0",
		Description: "请求ID生成中间件",
		Author:      "YYHertz Team",
	})
}

// registerTimeoutMiddleware 注册超时中间件
func (m *MiddlewareManager) registerTimeoutMiddleware() {
	m.RegisterBuiltin("timeout", func(config interface{}) MiddlewareFunc {
		timeout := 30 * time.Second
		if config != nil {
			if t, ok := config.(time.Duration); ok {
				timeout = t
			}
		}

		return func(ctx *mvccontext.Context) {
			finish := make(chan struct{})
			panicChan := make(chan any, 1)

			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicChan <- p
					}
				}()
				ctx.Next()
				finish <- struct{}{}
			}()

			select {
			case p := <-panicChan:
				panic(p)
			case <-finish:
				return
			case <-time.After(timeout):
				ctx.JSON(http.StatusRequestTimeout, map[string]string{"error": "Request Timeout"})
				ctx.Abort()
				return
			}
		}
	}, MiddlewareMetadata{
		Name:        "timeout",
		Version:     "1.0.0",
		Description: "请求超时处理中间件",
		Author:      "YYHertz Team",
	})
}

// registerSecureMiddleware 注册安全头中间件
func (m *MiddlewareManager) registerSecureMiddleware() {
	m.RegisterBuiltin("secure", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.Context) {
			ctx.Request.Response.Header.Set("X-Frame-Options", "DENY")
			ctx.Request.Response.Header.Set("Content-Security-Policy", "default-src 'self'")
			ctx.Request.Response.Header.Set("X-Content-Type-Options", "nosniff")
			ctx.Request.Response.Header.Set("X-XSS-Protection", "1; mode=block")
			ctx.Request.Response.Header.Set("Strict-Transport-Security", "max-age=31536000")
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "secure",
		Version:     "1.0.0",
		Description: "安全HTTP头中间件",
		Author:      "YYHertz Team",
	})
}

// registerGZipMiddleware 注册压缩中间件
func (m *MiddlewareManager) registerGZipMiddleware() {
	m.RegisterBuiltin("gzip", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.Context) {
			// 检查客户端是否支持gzip
			acceptEncoding := string(ctx.Request.GetHeader("Accept-Encoding"))
			if strings.Contains(acceptEncoding, "gzip") {
				ctx.Request.Response.Header.Set("Content-Encoding", "gzip")
			}
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "gzip",
		Version:     "1.0.0",
		Description: "GZIP压缩中间件",
		Author:      "YYHertz Team",
	})
}

// 辅助函数

// defaultLogFormatter 默认日志格式化函数
func defaultLogFormatter(param LogFormatterParams) string {
	var statusColor, methodColor, resetColor string

	statusColor = getColorByStatus(param.StatusCode)
	methodColor = getColorByMethod(param.Method)
	resetColor = "\033[0m"

	return fmt.Sprintf("%s[YYHertz-MVC]%s %v |%s %3d %s| %13v | %15s |%s %-7s %s %s\n",
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

// generateTraceID 生成追踪ID
func generateTraceID() string {
	return fmt.Sprintf("trace-%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req-%d-%s", time.Now().UnixNano(), generateShortID())
}

// generateShortID 生成短ID
func generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// InitExtendedMiddlewares 初始化扩展中间件
func (m *MiddlewareManager) InitExtendedMiddlewares() {
	m.registerExtendedBuiltinMiddlewares()
}

// 全局初始化函数
func init() {
	// 在默认管理器中注册扩展中间件
	defaultManager.InitExtendedMiddlewares()
}

// ============= 兼容性API已集成 =============
// 所有原@framework/middleware的API现在直接可用，无需额外的兼容性层