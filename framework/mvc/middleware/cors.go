package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/config"
)

// CORSMiddleware 跨域中间件 - 处理跨域请求
func CORSMiddleware() Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		// 获取请求信息
		requestOrigin := string(ctx.GetHeader("Origin"))
		clientIP := ctx.ClientIP()
		method := string(ctx.Method())
		path := string(ctx.Path())
		requestID := ctx.GetString("request_id")

		// 获取允许的源地址，默认为所有来源
		origin := "*" // 简化配置，实际应用中可以从配置文件读取

		// 使用单例日志系统记录CORS处理
		corsFields := map[string]any{
			"event":          "cors_request",
			"request_origin": requestOrigin,
			"allowed_origin": origin,
			"client_ip":      clientIP,
			"method":         method,
			"path":           path,
			"request_id":     requestID,
		}

		// 设置CORS头部(参考FreeCar项目)
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		ctx.Header("Access-Control-Allow-Headers", "AccessToken, Content-Type, Authorization, X-Requested-With")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Max-Age", "3600")

		// 处理预检请求
		if method == "OPTIONS" {
			corsFields["preflight"] = true
			go func() {
				config.WithFields(corsFields).Debug("CORS preflight request handled")
			}()
			ctx.Status(204)
			ctx.Abort()
			return
		}

		// 记录普通CORS请求
		config.WithFields(corsFields).Debug("CORS headers set for request")

		ctx.Next(c)
	}
}

// CORSMiddlewareWithConfig 带配置的跨域中间件
func CORSMiddlewareWithConfig(origins []string, methods []string, headers []string) Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		// 获取请求信息
		requestOrigin := string(ctx.GetHeader("Origin"))
		clientIP := ctx.ClientIP()
		method := string(ctx.Method())
		path := string(ctx.Path())
		requestID := ctx.GetString("request_id")

		// 检查请求来源
		allowedOrigin := "*"
		originAllowed := true

		// 如果指定了具体的来源，进行匹配
		if len(origins) > 0 && origins[0] != "*" {
			originAllowed = false
			for _, origin := range origins {
				if origin == requestOrigin {
					allowedOrigin = origin
					originAllowed = true
					break
				}
			}
			if !originAllowed && requestOrigin != "" {
				// 如果没有匹配到，禁止跨域
				allowedOrigin = "null"
			}
		}

		// 设置默认方法和头部
		if len(methods) == 0 {
			methods = []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"}
		}
		if len(headers) == 0 {
			headers = []string{"AccessToken", "Content-Type", "Authorization", "X-Requested-With"}
		}

		// 使用单例日志系统记录CORS配置处理
		corsFields := map[string]any{
			"event":           "cors_config_request",
			"request_origin":  requestOrigin,
			"allowed_origin":  allowedOrigin,
			"origin_allowed":  originAllowed,
			"client_ip":       clientIP,
			"method":          method,
			"path":            path,
			"request_id":      requestID,
			"allowed_methods": joinStrings(methods, ", "),
			"allowed_headers": joinStrings(headers, ", "),
		}

		// 如果来源不被允许，记录警告
		if !originAllowed {
			go func() {
				config.WithFields(corsFields).Warn("CORS request from disallowed origin")
			}()
		} else {
			go func() {
				config.WithFields(corsFields).Debug("CORS request from allowed origin")
			}()
		}

		// 设置CORS头部
		ctx.Header("Access-Control-Allow-Origin", allowedOrigin)
		ctx.Header("Access-Control-Allow-Methods", joinStrings(methods, ", "))
		ctx.Header("Access-Control-Allow-Headers", joinStrings(headers, ", "))
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Max-Age", "3600")

		// 处理预检请求
		if method == "OPTIONS" {
			corsFields["preflight"] = true
			go func() {
				config.WithFields(corsFields).Debug("CORS configured preflight request handled")
			}()

			ctx.Status(204)
			ctx.Abort()
			return
		}

		ctx.Next(c)
	}
}

// joinStrings 连接字符串数组
func joinStrings(arr []string, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	if len(arr) == 1 {
		return arr[0]
	}

	result := arr[0]
	for _, s := range arr[1:] {
		result += sep + s
	}
	return result
}
