package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// LoggerMiddleware 日志中间件 - 记录请求日志和响应时间
func LoggerMiddleware() Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()
		method := string(ctx.Method())
		path := string(ctx.Path())
		
		fmt.Printf("[%s] %s %s", time.Now().Format("2006-01-02 15:04:05"), method, path)
		
		ctx.Next(c)
		
		latency := time.Since(start)
		status := ctx.Response.StatusCode()
		fmt.Printf(" - %d - %v\n", status, latency)
	}
}