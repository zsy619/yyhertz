package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// RateLimitMiddleware 限流中间件 - 限制请求频率
func RateLimitMiddleware(maxRequests int, duration time.Duration) Middleware {
	requests := make(map[string][]time.Time)
	
	return func(c context.Context, ctx *app.RequestContext) {
		clientIP := ctx.ClientIP()
		now := time.Now()
		
		if times, exists := requests[clientIP]; exists {
			validTimes := make([]time.Time, 0)
			for _, t := range times {
				if now.Sub(t) < duration {
					validTimes = append(validTimes, t)
				}
			}
			
			if len(validTimes) >= maxRequests {
				ctx.JSON(429, map[string]string{
					"error": "请求过于频繁",
				})
				ctx.Abort()
				return
			}
			
			requests[clientIP] = append(validTimes, now)
		} else {
			requests[clientIP] = []time.Time{now}
		}
	}
}