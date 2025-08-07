package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/errors"
	"github.com/zsy619/yyhertz/framework/response"
)

// RecoveryMiddleware 恢复中间件 - 捕获panic并恢复(参考FreeCar项目)
func RecoveryMiddleware() Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				// 使用单例日志系统记录详细的错误信息和堆栈
				stack := debug.Stack()

				// 获取请求信息用于结构化日志
				method := string(ctx.Method())
				path := string(ctx.Path())
				clientIP := ctx.ClientIP()
				userAgent := string(ctx.UserAgent())

				// 使用结构化日志记录panic信息
				config.WithFields(map[string]any{
					"error":      fmt.Sprintf("%v", err),
					"method":     method,
					"path":       path,
					"client_ip":  clientIP,
					"user_agent": userAgent,
					"stack":      string(stack),
				}).Error("PANIC recovered in middleware")

				// 返回标准错误响应
				response := response.BuildErrorResp(errors.ServiceError.WithMessage("Internal Server Error"))
				ctx.JSON(500, response)
				ctx.Abort()
			}
		}()

		ctx.Next(c)
	}
}

// RecoveryMiddlewareWithHandler 带自定义处理器的恢复中间件
func RecoveryMiddlewareWithHandler(handler func(c context.Context, ctx *app.RequestContext, err any)) Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				if handler != nil {
					handler(c, ctx, err)
				} else {
					// 默认处理 - 使用单例日志系统
					stack := debug.Stack()

					// 获取请求信息用于结构化日志
					method := string(ctx.Method())
					path := string(ctx.Path())
					clientIP := ctx.ClientIP()
					userAgent := string(ctx.UserAgent())

					// 使用结构化日志记录panic信息
					config.WithFields(map[string]any{
						"error":      fmt.Sprintf("%v", err),
						"method":     method,
						"path":       path,
						"client_ip":  clientIP,
						"user_agent": userAgent,
						"stack":      string(stack),
					}).Error("PANIC recovered in middleware with custom handler")

					response := response.BuildErrorResp(errors.ServiceError.WithMessage("Internal Server Error"))
					ctx.JSON(500, response)
					ctx.Abort()
				}
			}
		}()

		ctx.Next(c)
	}
}

// LoggingRecoveryHandler 记录日志的恢复处理器
func LoggingRecoveryHandler() func(c context.Context, ctx *app.RequestContext, err any) {
	return func(c context.Context, ctx *app.RequestContext, err any) {
		// 获取请求信息
		method := string(ctx.Method())
		path := string(ctx.Path())
		userAgent := string(ctx.UserAgent())
		clientIP := ctx.ClientIP()

		// 使用单例日志系统记录详细错误信息
		stack := debug.Stack()

		// 使用结构化日志记录详细的panic信息
		config.WithFields(map[string]any{
			"error":      fmt.Sprintf("%v", err),
			"method":     method,
			"path":       path,
			"client_ip":  clientIP,
			"user_agent": userAgent,
			"stack":      string(stack),
			"handler":    "LoggingRecoveryHandler",
		}).Error("PANIC recovered by logging recovery handler")

		// 返回错误响应
		response := response.BuildErrorResp(errors.ServiceError.WithMessage(fmt.Sprintf("Internal Server Error: %v", err)))
		ctx.JSON(500, response)
		ctx.Abort()
	}
}

// EnhancedRecoveryHandler 增强的恢复处理器 - 使用单例日志系统和更智能的错误分类
func EnhancedRecoveryHandler() func(c context.Context, ctx *app.RequestContext, err any) {
	return func(c context.Context, ctx *app.RequestContext, err any) {
		// 获取请求信息
		method := string(ctx.Method())
		path := string(ctx.Path())
		userAgent := string(ctx.UserAgent())
		clientIP := ctx.ClientIP()

		// 获取请求ID（如果存在）
		requestID := ctx.GetString("request_id")
		if requestID == "" {
			requestID = "unknown"
		}

		// 获取堆栈信息
		stack := debug.Stack()

		// 创建基础日志字段
		logFields := map[string]any{
			"error":      fmt.Sprintf("%v", err),
			"method":     method,
			"path":       path,
			"client_ip":  clientIP,
			"user_agent": userAgent,
			"request_id": requestID,
			"stack":      string(stack),
			"handler":    "EnhancedRecoveryHandler",
		}

		// 根据错误类型进行分类记录
		switch e := err.(type) {
		case *errors.ErrNo:
			// 业务错误
			logFields["error_type"] = "business_error"
			logFields["error_code"] = e.ErrCode
			config.WithFields(logFields).Warn("Business panic recovered")

			// 返回业务错误响应
			response := response.BuildErrorResp(e)
			ctx.JSON(400, response)

		case error:
			// 系统错误
			logFields["error_type"] = "system_error"
			config.WithFields(logFields).Error("System panic recovered")

			// 返回系统错误响应
			response := response.BuildErrorResp(errors.ServiceError.WithMessage("Internal Server Error"))
			ctx.JSON(500, response)

		default:
			// 未知错误
			logFields["error_type"] = "unknown_error"
			config.WithFields(logFields).Error("Unknown panic recovered")

			// 返回通用错误响应
			response := response.BuildErrorResp(errors.ServiceError.WithMessage("Internal Server Error"))
			ctx.JSON(500, response)
		}

		ctx.Abort()
	}
}

// RequestAwareRecoveryHandler 请求感知的恢复处理器 - 根据请求路径提供不同的错误处理
func RequestAwareRecoveryHandler() func(c context.Context, ctx *app.RequestContext, err any) {
	return func(c context.Context, ctx *app.RequestContext, err any) {
		// 获取请求信息
		method := string(ctx.Method())
		path := string(ctx.Path())
		userAgent := string(ctx.UserAgent())
		clientIP := ctx.ClientIP()
		requestID := ctx.GetString("request_id")

		if requestID == "" {
			requestID = "unknown"
		}

		stack := debug.Stack()

		// 判断是否为API请求
		isAPIRequest := len(path) >= 4 && path[:4] == "/api"

		// 创建日志字段
		logFields := map[string]any{
			"error":          fmt.Sprintf("%v", err),
			"method":         method,
			"path":           path,
			"client_ip":      clientIP,
			"user_agent":     userAgent,
			"request_id":     requestID,
			"stack":          string(stack),
			"handler":        "RequestAwareRecoveryHandler",
			"is_api_request": isAPIRequest,
		}

		// 记录错误日志
		if isAPIRequest {
			config.WithRequestID(requestID).WithFields(logFields).Error("API panic recovered")
		} else {
			config.WithFields(logFields).Error("Web panic recovered")
		}

		// 根据请求类型返回不同响应格式
		if isAPIRequest {
			// API请求返回JSON格式错误
			response := response.BuildErrorResp(errors.ServiceError.WithMessage("Internal Server Error"))
			ctx.JSON(500, response)
		} else {
			// Web请求返回HTML错误页面（简化版）
			ctx.HTML(500, "", map[string]any{
				"error": "Internal Server Error",
				"path":  path,
			})
		}

		ctx.Abort()
	}
}
