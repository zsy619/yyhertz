package middleware

import (
	"context"
	"hertz-controller/framework/config"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORSMiddleware 跨域中间件 - 处理跨域请求
func CORSMiddleware() Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		// 获取允许的源地址，默认为所有来源
		origin := config.Get("cors.origin", "*")
		if origin == "" {
			origin = "*"
		}
		
		// 设置CORS头部(参考FreeCar项目)
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		ctx.Header("Access-Control-Allow-Headers", "AccessToken, Content-Type, Authorization, X-Requested-With")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Max-Age", "3600")
		
		// 处理预检请求
		if string(ctx.Method()) == "OPTIONS" {
			ctx.Status(204)
			ctx.Abort()
			return
		}
		
		ctx.Next(c)
	}
}

// CORSMiddlewareWithConfig 带配置的跨域中间件
func CORSMiddlewareWithConfig(origins []string, methods []string, headers []string) Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		// 检查请求来源
		requestOrigin := string(ctx.GetHeader("Origin"))
		allowedOrigin := "*"
		
		// 如果指定了具体的来源，进行匹配
		if len(origins) > 0 && origins[0] != "*" {
			for _, origin := range origins {
				if origin == requestOrigin {
					allowedOrigin = origin
					break
				}
			}
			if allowedOrigin == "*" && requestOrigin != "" {
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
		
		// 设置CORS头部
		ctx.Header("Access-Control-Allow-Origin", allowedOrigin)
		ctx.Header("Access-Control-Allow-Methods", joinStrings(methods, ", "))
		ctx.Header("Access-Control-Allow-Headers", joinStrings(headers, ", "))
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Max-Age", "3600")
		
		// 处理预检请求
		if string(ctx.Method()) == "OPTIONS" {
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