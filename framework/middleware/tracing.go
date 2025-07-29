package middleware

import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"go.opentelemetry.io/otel/trace"
)

func generateTraceID() string {
	ctx := context.Background()
	tracer := trace.NewNoopTracerProvider().Tracer("")
	ctx, _ = tracer.Start(ctx, "dummy-span")
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().TraceID().String()
}

// 假设有一个链路追踪中间件函数
func TracingMiddleware() Middleware {
	return func(ctx context.Context, c *app.RequestContext) {
		// 从 Header 中提取 TraceID，或生成新的
		traceID := string(c.GetHeader("X-Trace-ID"))
		if traceID == "" {
			traceID = generateTraceID() // 自行实现生成逻辑，如 UUID
		}
		// 将 TraceID 放入上下文，便于后续使用
		ctx = context.WithValue(ctx, "traceID", traceID)
		// 可在此处上报开始时间、调用链信息等
		c.Set("traceID", traceID) // 也可以存到 Hertz 的上下文中

		// 打印或记录日志（可选）
		log.Printf("[Tracing] 开始处理请求，TraceID: %s", traceID)

		// 请求结束后可记录状态、耗时等（可选）
		log.Printf("[Tracing] 请求结束，TraceID: %s", traceID)
	}
}
