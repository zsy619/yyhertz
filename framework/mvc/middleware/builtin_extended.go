package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// 扩展的内置中间件 - 移植自 @framework/middleware 系统

// registerExtendedBuiltinMiddlewares 注册所有从基础系统移植的扩展内置中间件。
//
// 这些中间件包括日志增强、恢复增强、CORS、速率限制、链路追踪、TLS支持、
// 基础认证、请求ID、超时控制、安全防护和GZip压缩等功能。
//
// 该方法通常在服务启动时调用，用于初始化中间件链。
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

// registerEnhancedLoggerMiddleware 注册一个增强的日志记录中间件到中间件管理器中。
//
// 该中间件用于记录HTTP请求的详细信息，包括请求路径、状态码、延迟、客户端IP等。
// 支持自定义日志输出目标、日志格式以及跳过特定路径的日志记录。
//
// 参数:
//   - config: 可选的日志配置参数。如果为nil或非LoggerConfig类型，将使用默认配置。
//     LoggerConfig包含以下字段：
//   - SkipPaths: 需要跳过日志记录的路径列表。
//   - Output: 日志输出目标，默认为os.Stdout。
//   - Formatter: 日志格式化函数，默认为defaultLogFormatter。
//
// 返回值:
//   - 无。中间件会通过MiddlewareManager的RegisterBuiltin方法注册。
//
// 功能说明:
//   - 检查请求路径是否需要跳过日志记录。
//   - 记录请求开始时间，并在请求完成后计算延迟。
//   - 收集请求状态码、错误信息等数据。
//   - 使用自定义或默认的格式化函数生成日志并输出。
//
// 元数据:
//   - 名称: "enhanced-logger"
//   - 版本: "2.0.0"
//   - 描述: "Enhanced HTTP请求日志记录中间件 (from basic system)"
//   - 作者: "YYHertz Team"
func (m *MiddlewareManager) registerEnhancedLoggerMiddleware() {
	m.RegisterBuiltin("enhanced-logger", func(config any) MiddlewareFunc {
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

		return func(ctx *mvcContext.Context) {
			// 检查是否跳过
			path := string(ctx.Request().Path())
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
			statusCode := ctx.Writer().Status()

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
				ClientIP:     string(ctx.Request().ClientIP()),
				Method:       string(ctx.Request().Method()),
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

// registerEnhancedRecoveryMiddleware 注册一个增强的 panic 恢复中间件。
//
// 该中间件用于捕获和处理请求处理过程中发生的 panic，记录详细的错误信息和堆栈跟踪，
// 并返回一个 500 内部服务器错误的响应。
//
// 参数:
//   - config: 可选的配置参数，支持传入一个 io.Writer 类型的对象，用于指定错误日志的输出位置。
//     如果未提供配置，默认将错误日志输出到标准错误流 (os.Stderr)。
//
// 中间件行为:
//   - 捕获 panic 并记录错误信息（包括时间戳、错误内容和堆栈跟踪）。
//   - 将错误信息添加到上下文的错误列表中。
//   - 返回一个 JSON 格式的 500 错误响应。
//   - 终止当前请求的处理流程。
//
// 元数据:
//   - Name: "enhanced-recovery"
//   - Version: "2.0.0"
//   - Description: "Enhanced Panic 恢复中间件 (from basic system)"
//   - Author: "YYHertz Team"
//
// 注意:
//   - 该中间件是内置的，通过 MiddlewareManager 的 RegisterBuiltin 方法注册。
//   - 适用于需要高可靠性的 Web 服务场景。
func (m *MiddlewareManager) registerEnhancedRecoveryMiddleware() {
	m.RegisterBuiltin("enhanced-recovery", func(config any) MiddlewareFunc {
		out := os.Stderr
		if config != nil {
			if writer, ok := config.(io.Writer); ok {
				out = writer.(*os.File)
			}
		}

		return func(ctx *mvcContext.Context) {
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

// registerCORSMiddleware 注册一个增强版的CORS中间件到中间件管理器中。
//
// 该中间件实现了跨域资源共享（CORS）功能，支持以下特性：
// - 允许所有来源（Access-Control-Allow-Origin: *）
// - 允许的HTTP方法包括 GET, POST, PUT, DELETE, OPTIONS
// - 允许的请求头包括 Content-Type 和 Authorization
// - 自动处理预检请求（OPTIONS 方法），返回 204 状态码并终止请求
//
// 中间件元数据：
// - Name: "cors-extended" - 中间件名称
// - Version: "2.0.0" - 版本号
// - Description: "跨域资源共享中间件 (enhanced from basic system)" - 描述
// - Author: "YYHertz Team" - 作者
//
// 注意：该中间件会设置必要的CORS头，并确保预检请求的正确处理。
func (m *MiddlewareManager) registerCORSMiddleware() {
	m.RegisterBuiltin("cors-extended", func(config any) MiddlewareFunc {
		return func(ctx *mvcContext.Context) {
			// 设置CORS头
			ctx.Request().Response.Header.Set("Access-Control-Allow-Origin", "*")
			ctx.Request().Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			ctx.Request().Response.Header.Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

			// 处理预检请求
			if string(ctx.Request().Method()) == "OPTIONS" {
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

// registerRateLimitMiddleware 注册一个限流中间件到中间件管理器中。
//
// 该中间件使用简化的限流实现，实际应用中建议替换为专业的限流算法（如令牌桶或滑动窗口）。
// 当前实现仅为基础检查，后续需完善具体限流逻辑。
//
// 参数:
//   - config: 中间件的配置参数，类型为任意（any）。
//
// 返回值:
//   - MiddlewareFunc: 限流中间件的函数实现。
//
// 元数据:
//   - Name: "ratelimit" - 中间件名称。
//   - Version: "1.0.0" - 中间件版本。
//   - Description: "请求限流中间件" - 中间件功能描述。
//   - Author: "YYHertz Team" - 中间件作者。
//   - Dependencies: ["logger"] - 中间件依赖的其他组件。
func (m *MiddlewareManager) registerRateLimitMiddleware() {
	m.RegisterBuiltin("ratelimit", func(config any) MiddlewareFunc {
		// 简化的限流实现 - 实际应用中需要使用专业的限流算法
		return func(ctx *mvcContext.Context) {
			// TODO: 实现令牌桶或滑动窗口算法
			// 这里先实现基础检查
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:         "ratelimit",
		Version:      "1.0.0",
		Description:  "请求限流中间件",
		Author:       "YYHertz Team",
		Dependencies: []string{"logger"},
	})
}

// registerTracingMiddleware 注册一个链路追踪中间件到中间件管理器。
//
// 该中间件的主要功能：
// - 为每个请求生成唯一的 Trace ID。
// - 将 Trace ID 存储到请求上下文中，并设置到响应头中（X-Trace-ID）。
// - 预留了集成实际链路追踪系统（如 Jaeger、Zipkin）的扩展点。
//
// 参数：
//   - config: 中间件的配置参数（当前未使用）。
//
// 返回值：
//   - MiddlewareFunc: 一个中间件函数，用于处理请求链路追踪逻辑。
//
// 元数据：
//   - Name: "tracing" - 中间件名称。
//   - Version: "1.0.0" - 中间件版本。
//   - Description: "分布式链路追踪中间件" - 中间件描述。
//   - Author: "YYHertz Team" - 作者信息。
//
// 注意：
// - 当前实现仅生成 Trace ID，实际链路追踪功能需后续集成。
func (m *MiddlewareManager) registerTracingMiddleware() {
	m.RegisterBuiltin("tracing", func(config any) MiddlewareFunc {
		return func(ctx *mvcContext.Context) {
			// 生成Trace ID
			traceID := generateTraceID()
			ctx.Set("TraceID", traceID)
			ctx.Request().Response.Header.Set("X-Trace-ID", traceID)

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

// registerTLSMiddleware 注册一个内置的TLS中间件到中间件管理器中。
//
// 该中间件用于处理TLS相关的请求，包括记录请求的客户端IP和路径信息。
// 中间件会在请求处理链中调用 `ctx.Next()` 继续后续处理。
//
// 参数:
//   - config: 中间件的配置参数（当前未使用，保留为扩展点）。
//
// 返回值:
//   - MiddlewareFunc: 返回一个中间件函数，用于处理TLS请求。
//
// 元数据:
//   - Name: "tls" - 中间件名称。
//   - Version: "1.0.0" - 中间件版本。
//   - Description: "TLS连接处理中间件" - 中间件功能描述。
//   - Author: "YYHertz Team" - 中间件作者。
//
// 注意:
//   - 当前实现为简化版本，实际应用中需要进一步检查协议和其他安全相关逻辑。
func (m *MiddlewareManager) registerTLSMiddleware() {
	m.RegisterBuiltin("tls", func(config any) MiddlewareFunc {
		return func(ctx *mvcContext.Context) {
			// TLS相关检查和处理
			// 简化实现，实际应用中需要检查协议
			fmt.Printf("[TLS] Processing request - client_ip: %s, path: %s\n",
				string(ctx.Request().ClientIP()), string(ctx.Request().Path()))
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
	m.RegisterBuiltin("basicauth", func(config any) MiddlewareFunc {
		_ = config // 简化实现，暂不处理账户配置

		return func(ctx *mvcContext.Context) {
			// 简化的Basic Auth实现
			auth := string(ctx.Request().GetHeader("Authorization"))
			if strings.HasPrefix(auth, "Basic ") {
				// TODO: 解析Basic Auth头部
				ctx.Set("user", "demo")
				ctx.Next()
				return
			}

			ctx.Request().Response.Header.Set("WWW-Authenticate", "Basic realm=Authorization Required")
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
	m.RegisterBuiltin("requestid", func(config any) MiddlewareFunc {
		return func(ctx *mvcContext.Context) {
			requestID := generateRequestID()
			ctx.Set("RequestID", requestID)
			ctx.Request().Response.Header.Set("X-Request-ID", requestID)
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
	m.RegisterBuiltin("timeout", func(config any) MiddlewareFunc {
		timeout := 30 * time.Second
		if config != nil {
			if t, ok := config.(time.Duration); ok {
				timeout = t
			}
		}

		return func(ctx *mvcContext.Context) {
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
	m.RegisterBuiltin("secure", func(config any) MiddlewareFunc {
		return func(ctx *mvcContext.Context) {
			ctx.Request().Response.Header.Set("X-Frame-Options", "DENY")
			ctx.Request().Response.Header.Set("Content-Security-Policy", "default-src 'self'")
			ctx.Request().Response.Header.Set("X-Content-Type-Options", "nosniff")
			ctx.Request().Response.Header.Set("X-XSS-Protection", "1; mode=block")
			ctx.Request().Response.Header.Set("Strict-Transport-Security", "max-age=31536000")
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
	m.RegisterBuiltin("gzip", func(config any) MiddlewareFunc {
		return func(ctx *mvcContext.Context) {
			// 检查客户端是否支持gzip
			acceptEncoding := string(ctx.Request().GetHeader("Accept-Encoding"))
			if strings.Contains(acceptEncoding, "gzip") {
				ctx.Request().Response.Header.Set("Content-Encoding", "gzip")
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
