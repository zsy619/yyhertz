package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/util"
)

// GenerateOpenTelemetryTraceID 生成一个 OpenTelemetry 格式的 Trace ID。
//
// 返回:
//
//	string: 生成的 Trace ID 字符串。
//
// 说明:
//
//	该函数通过创建一个虚拟的 OpenTelemetry Span 上下文，
//	并从中提取 Trace ID 来生成一个符合 OpenTelemetry 规范的 Trace ID。
//	主要用于测试或需要模拟 Trace ID 的场景。
func GenerateOpenTelemetryTraceID() string {
	ctx := context.Background()
	tracer := noop.NewTracerProvider().Tracer("default")
	ctx, _ = tracer.Start(ctx, "dummy-span")
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().TraceID().String()
}

// TracingMiddleware 返回一个中间件，用于追踪 HTTP 请求的生命周期。
//
// 该中间件的主要功能包括：
// - 从请求头中提取或生成唯一的 TraceID。
// - 生成或复用请求的 request_id。
// - 将 TraceID 注入到请求上下文中，便于后续使用。
// - 记录请求开始时的关键信息（如方法、路径、客户端 IP 等）。
// - 记录请求完成时的状态码、处理时间等信息，并根据状态码选择日志级别。
//
// 日志级别规则：
// - 状态码 >= 500：记录为错误（Error）。
// - 状态码 >= 400：记录为警告（Warn）。
// - 其他状态码：记录为信息（Info）。
//
// 返回值：
//   - Middleware：一个符合中间件签名的函数，用于处理请求追踪。
func TracingMiddleware() Middleware {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		// 从 Header 中提取 TraceID，或生成新的
		traceID := string(c.GetHeader("X-Trace-ID"))
		if traceID == "" {
			traceID = GenerateOpenTelemetryTraceID()
		}

		// 如果没有request_id，也生成一个
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = util.ShortID()
			c.Set("request_id", requestID)
		}

		// 将 TraceID 放入上下文，便于后续使用
		ctx = context.WithValue(ctx, "traceID", traceID)
		c.Set("traceID", traceID)

		// 使用单例日志系统记录追踪开始
		go func() {
			config.WithFields(map[string]any{
				"trace_id":   traceID,
				"request_id": requestID,
				"method":     string(c.Method()),
				"path":       string(c.Path()),
				"client_ip":  c.ClientIP(),
				"user_agent": string(c.UserAgent()),
				"start_time": start.Format(time.RFC3339),
			}).Info("Tracing: Request started")
		}()

		// 处理请求
		c.Next(ctx)

		// 计算处理时间
		duration := time.Since(start)
		statusCode := c.Response.StatusCode()

		// 使用单例日志系统记录追踪结束
		endFields := map[string]any{
			"trace_id":    traceID,
			"request_id":  requestID,
			"method":      string(c.Method()),
			"path":        string(c.Path()),
			"status_code": statusCode,
			"duration":    duration.String(),
			"duration_ms": duration.Milliseconds(),
			"end_time":    time.Now().Format(time.RFC3339),
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			go func() {
				config.WithFields(endFields).Error("Tracing: Request completed with server error")
			}()
		} else if statusCode >= 400 {
			go func() {
				config.WithFields(endFields).Warn("Tracing: Request completed with client error")
			}()
		} else {
			go func() {
				config.WithFields(endFields).Info("Tracing: Request completed successfully")
			}()
		}
	}
}

// SimpleTracingMiddleware 返回一个中间件函数，用于为请求添加简单的追踪功能。
//
// 功能说明：
// 1. 从请求头 "X-Trace-ID" 中获取 TraceID，如果不存在则生成一个简短的唯一 ID。
// 2. 将 TraceID 设置到请求上下文中，支持两种键名："traceID" 和 "trace_id"（兼容性）。
// 3. 异步记录 TraceID 到日志中，便于后续追踪。
// 4. 调用 c.Next(ctx) 继续处理后续中间件或请求。
//
// 注意：
// - 该中间件适用于需要简单追踪请求的场景，不包含复杂的分布式追踪逻辑。
// - 日志记录是异步的，不会阻塞请求处理流程。
func SimpleTracingMiddleware() Middleware {
	return func(ctx context.Context, c *app.RequestContext) {
		// 生成或获取TraceID
		traceID := string(c.GetHeader("X-Trace-ID"))
		if traceID == "" {
			traceID = util.ShortID() // 使用更简单的ID生成
		}

		// 设置到上下文
		c.Set("traceID", traceID)
		c.Set("trace_id", traceID) // 兼容性

		// 记录追踪信息
		go func() {
			config.WithField("trace_id", traceID).Debug("Trace ID assigned")
		}()

		c.Next(ctx)
	}
}
