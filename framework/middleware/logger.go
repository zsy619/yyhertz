package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
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

// DefaultLoggerConfig 返回默认日志中间件配置
func DefaultLoggerConfig() *MiddlewareLoggerConfig {
	return &MiddlewareLoggerConfig{
		EnableRequestBody:  false,
		EnableResponseBody: false,
		SkipPaths:          []string{"/health", "/ping"},
		MaxBodySize:        1024, // 1KB
	}
}

// LoggerMiddleware 增强的请求日志中间件
func LoggerMiddleware() Middleware {
	return LoggerMiddlewareWithConfig(DefaultLoggerConfig())
}

// LoggerMiddlewareWithConfig 带配置的日志中间件
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

		config.WithFields(fields).Info("Request started")

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
		if statusCode >= 500 {
			config.WithFields(responseFields).Error("Request completed with server error")
		} else if statusCode >= 400 {
			config.WithFields(responseFields).Warn("Request completed with client error")
		} else {
			config.WithFields(responseFields).Info("Request completed successfully")
		}
	}
}

// AccessLogMiddleware 简化的访问日志中间件
func AccessLogMiddleware() Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()

		ctx.Next(c)

		duration := time.Since(start)

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

		ctx.Next(c)
	}
}
