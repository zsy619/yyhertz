package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/util"
)

// MiddlewareLoggerConfig 日志中间件配置
type MiddlewareLoggerConfig struct {
	EnableRequestBody  bool     // 是否记录请求体
	EnableResponseBody bool     // 是否记录响应体
	SkipPaths          []string // 跳过记录的路径
	MaxBodySize        int      // 最大记录的Body大小
}

// DefaultLoggerConfig 返回一个默认的中间件日志配置。
// 配置包括：
//   - EnableRequestBody: 是否启用请求体日志记录，默认为 false。
//   - EnableResponseBody: 是否启用响应体日志记录，默认为 false。
//   - SkipPaths: 需要跳过日志记录的路径列表，默认包含 "/health" 和 "/ping"。
//   - MaxBodySize: 日志记录的最大请求/响应体大小（单位：字节），默认为 1024（1KB）。
func DefaultLoggerConfig() *MiddlewareLoggerConfig {
	return &MiddlewareLoggerConfig{
		EnableRequestBody:  false,
		EnableResponseBody: false,
		SkipPaths:          []string{"/health", "/ping"},
		MaxBodySize:        1024, // 1KB
	}
}

// LoggerMiddleware 返回一个默认配置的日志中间件。
// 该中间件用于记录请求和响应的基本信息，如请求方法、路径、状态码和耗时。
// 返回的中间件可以直接用于HTTP服务器的中间件链中。
func LoggerMiddleware() Middleware {
	return LoggerMiddlewareWithConfig(DefaultLoggerConfig())
}

// LoggerMiddlewareWithConfig 返回一个中间件函数，用于记录 HTTP 请求和响应的详细信息。
//
// 参数:
//
//	logConfig *MiddlewareLoggerConfig: 日志配置，包含是否启用请求/响应体记录、最大记录体大小、跳过的路径等配置项。
//
// 返回值:
//
//	Middleware: 一个中间件函数，用于处理请求并记录日志。
//
// 功能描述:
//  1. 检查请求路径是否在跳过列表中，如果是则直接跳过日志记录。
//  2. 为请求生成唯一请求ID，并记录请求开始时的基本信息（如请求方法、路径、用户代理、客户端IP等）。
//  3. 如果启用，记录请求体内容（超过最大记录体大小时会截断并记录大小）。
//  4. 继续处理请求，并在请求完成后记录响应状态码、处理时间等信息。
//  5. 如果启用，记录响应体内容（超过最大记录体大小时会截断并记录大小）。
//  6. 根据响应状态码选择日志级别（错误、警告或信息）。
//
// 示例:
//
//	配置日志中间件并应用到路由中：
//	```
//	logConfig := &MiddlewareLoggerConfig{
//	    EnableRequestBody:  true,
//	    EnableResponseBody: true,
//	    MaxBodySize:        1024,
//	    SkipPaths:         []string{"/health"},
//	}
//	router.Use(LoggerMiddlewareWithConfig(logConfig))
//	```
//
// 注意事项:
//   - 日志记录的性能开销取决于是否启用请求/响应体记录以及记录体大小。
//   - 跳过的路径应避免包含敏感信息。
func LoggerMiddlewareWithConfig(logConfig *MiddlewareLoggerConfig) Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()
		path := string(ctx.Path())

		// 检查是否跳过此路径
		for _, skipPath := range logConfig.SkipPaths {
			if path == skipPath {
				ctx.Next(c)
				return
			}
		}

		// 生成请求ID
		requestID := util.ShortID()
		ctx.Set("request_id", requestID)

		// 记录请求开始
		fields := map[string]any{
			"request_id": requestID,
			"method":     string(ctx.Method()),
			"path":       path,
			"user_agent": string(ctx.UserAgent()),
			"client_ip":  ctx.ClientIP(),
			"timestamp":  start.Format(time.RFC3339),
		}

		// 记录请求体（如果启用）
		if logConfig.EnableRequestBody && ctx.Request.Body() != nil {
			bodySize := len(ctx.Request.Body())
			if bodySize > 0 && bodySize <= logConfig.MaxBodySize {
				fields["request_body"] = string(ctx.Request.Body())
			} else if bodySize > logConfig.MaxBodySize {
				fields["request_body_size"] = bodySize
				fields["request_body_truncated"] = true
			}
		}

		// 安全地记录请求开始日志
		safeLogRequest(fields)

		// 继续处理请求
		ctx.Next(c)

		// 计算处理时间
		duration := time.Since(start)
		statusCode := ctx.Response.StatusCode()

		// 准备响应日志字段
		responseFields := map[string]any{
			"timestamp":   time.Now().Format(time.RFC3339),
			"status_code": statusCode,
			"path":        path,
			"method":      string(ctx.Method()),
			"request_id":  requestID,
			"duration_ms": duration.Milliseconds(),
			"duration":    duration.String(),
		}

		// 记录响应体（如果启用）
		if logConfig.EnableResponseBody {
			responseBody := ctx.Response.Body()
			if len(responseBody) > 0 && len(responseBody) <= logConfig.MaxBodySize {
				responseFields["response_body"] = string(responseBody)
			} else if len(responseBody) > logConfig.MaxBodySize {
				responseFields["response_body_size"] = len(responseBody)
				responseFields["response_body_truncated"] = true
			}
		}

		// 根据状态码选择日志级别使用单例日志系统
		// 使用安全的日志记录方式，防止panic
		safeLogWithFields(responseFields, statusCode)
	}
}

// safeLogWithFields 安全地记录带字段的日志，防止panic
func safeLogWithFields(fields map[string]any, statusCode int) {
	defer func() {
		if r := recover(); r != nil {
			// 如果发生panic，使用基础日志记录
			config.Errorf("Logger middleware panic recovered: %v, status: %d", r, statusCode)
		}
	}()

	// 验证和清理字段
	safeFields := make(map[string]any)
	for k, v := range fields {
		if k == "" {
			continue // 跳过空键名
		}
		
		// 处理可能有问题的值
		switch val := v.(type) {
		case nil:
			safeFields[k] = "<nil>"
		case string:
			// 限制字符串长度防止日志过大
			if len(val) > 10000 {
				safeFields[k] = val[:10000] + "...[truncated]"
			} else {
				safeFields[k] = val
			}
		case []byte:
			// 将字节数组转换为安全的字符串
			if len(val) > 1000 {
				safeFields[k] = string(val[:1000]) + "...[truncated]"
			} else {
				safeFields[k] = string(val)
			}
		default:
			// 其他类型直接使用，但如果出问题会被recover捕获
			safeFields[k] = val
		}
	}

	// 根据状态码选择日志级别
	message := getLogMessage(statusCode)
	
	// 尝试使用WithFields，如果失败则使用fallback
	if logEntry := safeCreateLogEntry(safeFields); logEntry != nil {
		if statusCode >= 500 {
			logEntry.Error(message)
		} else if statusCode >= 400 {
			logEntry.Warn(message)
		} else {
			logEntry.Info(message)
		}
	} else {
		// Fallback: 使用简单的日志记录
		config.Infof("%s (status: %d)", message, statusCode)
	}
}

// safeCreateLogEntry 安全地创建带字段的日志条目
func safeCreateLogEntry(fields map[string]any) *logrus.Entry {
	defer func() {
		if recover() != nil {
			// 如果创建日志条目时发生panic，返回nil让调用方使用fallback
		}
	}()
	
	// 尝试创建带字段的日志条目
	return config.WithFields(fields)
}

// getLogMessage 根据状态码获取日志消息
func getLogMessage(statusCode int) string {
	if statusCode >= 500 {
		return "Request completed with server error"
	} else if statusCode >= 400 {
		return "Request completed with client error"
	} else {
		return "Request completed successfully"
	}
}

// safeLogRequest 安全地记录请求开始日志
func safeLogRequest(fields map[string]any) {
	defer func() {
		if r := recover(); r != nil {
			// 如果发生panic，使用基础日志记录
			config.Errorf("Logger middleware panic recovered in request start: %v", r)
		}
	}()

	// 验证和清理字段
	safeFields := make(map[string]any)
	for k, v := range fields {
		if k == "" {
			continue // 跳过空键名
		}
		
		// 处理可能有问题的值
		switch val := v.(type) {
		case nil:
			safeFields[k] = "<nil>"
		case string:
			// 限制字符串长度防止日志过大
			if len(val) > 10000 {
				safeFields[k] = val[:10000] + "...[truncated]"
			} else {
				safeFields[k] = val
			}
		case []byte:
			// 将字节数组转换为安全的字符串
			if len(val) > 1000 {
				safeFields[k] = string(val[:1000]) + "...[truncated]"
			} else {
				safeFields[k] = string(val)
			}
		default:
			// 其他类型直接使用，但如果出问题会被recover捕获
			safeFields[k] = val
		}
	}

	// 尝试使用WithFields，如果失败则使用fallback
	if logEntry := safeCreateLogEntry(safeFields); logEntry != nil {
		logEntry.Info("Request started")
	} else {
		// Fallback: 使用简单的日志记录
		config.Info("Request started")
	}
}

// AccessLogMiddleware 返回一个中间件函数，用于记录HTTP请求的访问日志。
// 该中间件会在请求处理前后记录以下信息：
//   - 请求方法（如GET、POST等）
//   - 请求路径
//   - 响应状态码
//   - 请求处理耗时（包括毫秒和字符串格式）
//   - 客户端IP地址
//
// 日志通过单例日志系统异步记录，避免阻塞请求处理流程。
// 注意：此中间件会调用两次 ctx.Next(c)，确保在请求处理前后都能执行日志记录逻辑。
func AccessLogMiddleware() Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()

		ctx.Next(c)

		duration := time.Since(start)
		go func() {
			// 使用单例日志系统记录访问日志
			config.WithFields(map[string]any{
				"type":        "access",
				"method":      string(ctx.Method()),
				"path":        string(ctx.Path()),
				"status_code": ctx.Response.StatusCode(),
				"duration":    duration.String(),
				"duration_ms": duration.Milliseconds(),
				"client_ip":   ctx.ClientIP(),
			}).Info("Access log")
		}()

		ctx.Next(c)
	}
}
