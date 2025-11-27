package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/config"
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
				go func() {
					// 使用单例日志系统记录限流事件
					config.WithFields(map[string]any{
						"client_ip":     clientIP,
						"max_requests":  maxRequests,
						"duration":      duration.String(),
						"current_count": len(validTimes),
						"path":          string(ctx.Path()),
						"method":        string(ctx.Method()),
						"user_agent":    string(ctx.UserAgent()),
					}).Warn("Rate limit exceeded")
				}()

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

		ctx.Next(c)
	}
}

// EnhancedRateLimitMiddleware 增强的限流中间件 - 支持不同路径的不同限制
func EnhancedRateLimitMiddleware(maxRequests int, duration time.Duration) Middleware {
	requests := make(map[string][]time.Time)

	return func(c context.Context, ctx *app.RequestContext) {
		clientIP := ctx.ClientIP()
		path := string(ctx.Path())
		method := string(ctx.Method())
		now := time.Now()

		// 为每个IP创建唯一键
		key := clientIP + ":" + method + ":" + path

		if times, exists := requests[key]; exists {
			validTimes := make([]time.Time, 0)
			for _, t := range times {
				if now.Sub(t) < duration {
					validTimes = append(validTimes, t)
				}
			}

			if len(validTimes) >= maxRequests {
				// 获取请求ID用于追踪
				requestID := ctx.GetString("request_id")

				go func() {
					// 使用单例日志系统记录详细的限流事件
					config.WithFields(map[string]any{
						"event":         "rate_limit_exceeded",
						"client_ip":     clientIP,
						"path":          path,
						"method":        method,
						"max_requests":  maxRequests,
						"duration":      duration.String(),
						"current_count": len(validTimes),
						"user_agent":    string(ctx.UserAgent()),
						"request_id":    requestID,
						"retry_after":   duration.Seconds(),
					}).Warn("Enhanced rate limit exceeded")
				}()

				// 返回更详细的错误信息
				ctx.JSON(429, map[string]any{
					"error":       "Rate limit exceeded",
					"message":     "请求过于频繁，请稍后再试",
					"retry_after": duration.Seconds(),
					"limit":       maxRequests,
					"window":      duration.String(),
				})
				ctx.Abort()
				return
			}

			requests[key] = append(validTimes, now)
		} else {
			requests[key] = []time.Time{now}

			go func() {
				// 记录新IP的首次访问
				config.WithFields(map[string]any{
					"event":     "first_request",
					"client_ip": clientIP,
					"path":      path,
					"method":    method,
				}).Debug("First request from client")
			}()
		}

		ctx.Next(c)
	}
}

// IPBasedRateLimitMiddleware 基于IP的限流中间件 - 只按IP限制，不区分路径
func IPBasedRateLimitMiddleware(maxRequests int, duration time.Duration) Middleware {
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
				go func() {
					config.WithFields(map[string]any{
						"event":         "ip_rate_limit_exceeded",
						"client_ip":     clientIP,
						"max_requests":  maxRequests,
						"duration":      duration.String(),
						"current_count": len(validTimes),
						"path":          string(ctx.Path()),
						"method":        string(ctx.Method()),
					}).Warn("IP-based rate limit exceeded")
				}()

				ctx.JSON(429, map[string]any{
					"error":   "IP rate limit exceeded",
					"message": "您的IP访问过于频繁，请稍后再试",
				})
				ctx.Abort()
				return
			}

			requests[clientIP] = append(validTimes, now)
		} else {
			requests[clientIP] = []time.Time{now}
		}

		ctx.Next(c)
	}
}
