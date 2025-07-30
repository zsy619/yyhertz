package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"go.opentelemetry.io/otel/trace"
	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/util"
)

func generateTraceID() string {
	ctx := context.Background()
	tracer := trace.NewNoopTracerProvider().Tracer("")
	ctx, _ = tracer.Start(ctx, "dummy-span")
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().TraceID().String()
}

// TracingMiddleware 链路追踪中间件 - 使用单例日志系统
func TracingMiddleware() Middleware {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		
		// 从 Header 中提取 TraceID，或生成新的
		traceID := string(c.GetHeader("X-Trace-ID"))
		if traceID == "" {
			traceID = generateTraceID()
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
		config.WithFields(map[string]any{
			"trace_id":   traceID,
			"request_id": requestID,
			"method":     string(c.Method()),
			"path":       string(c.Path()),
			"client_ip":  c.ClientIP(),
			"user_agent": string(c.UserAgent()),
			"start_time": start.Format(time.RFC3339),
		}).Info("Tracing: Request started")
		
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
			config.WithFields(endFields).Error("Tracing: Request completed with server error")
		} else if statusCode >= 400 {
			config.WithFields(endFields).Warn("Tracing: Request completed with client error")
		} else {
			config.WithFields(endFields).Info("Tracing: Request completed successfully")
		}
	}
}

// SimpleTracingMiddleware 简化的链路追踪中间件
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
		config.WithField("trace_id", traceID).Debug("Trace ID assigned")
		
		c.Next(ctx)
	}
}
