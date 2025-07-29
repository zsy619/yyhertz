package middleware

import (
	"context"
	"fmt"
	"hertz-controller/framework/types"
	"hertz-controller/framework/util"
	"log"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
)

// RecoveryMiddleware 恢复中间件 - 捕获panic并恢复(参考FreeCar项目)
func RecoveryMiddleware() Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				// 记录详细的错误信息和堆栈
				stack := debug.Stack()
				log.Printf("PANIC: %v\nStack: %s", err, string(stack))
				
				// 返回标准错误响应
				response := util.BuildErrorResp(types.ServiceError.WithMessage("Internal Server Error"))
				ctx.JSON(500, response)
				ctx.Abort()
			}
		}()
		
		ctx.Next(c)
	}
}

// RecoveryMiddlewareWithHandler 带自定义处理器的恢复中间件
func RecoveryMiddlewareWithHandler(handler func(c context.Context, ctx *app.RequestContext, err interface{})) Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				if handler != nil {
					handler(c, ctx, err)
				} else {
					// 默认处理
					stack := debug.Stack()
					log.Printf("PANIC: %v\nStack: %s", err, string(stack))
					
					response := util.BuildErrorResp(types.ServiceError.WithMessage("Internal Server Error"))
					ctx.JSON(500, response)
					ctx.Abort()
				}
			}
		}()
		
		ctx.Next(c)
	}
}

// LoggingRecoveryHandler 记录日志的恢复处理器
func LoggingRecoveryHandler() func(c context.Context, ctx *app.RequestContext, err interface{}) {
	return func(c context.Context, ctx *app.RequestContext, err interface{}) {
		// 获取请求信息
		method := string(ctx.Method())
		path := string(ctx.Path())
		userAgent := string(ctx.UserAgent())
		clientIP := ctx.ClientIP()
		
		// 记录详细错误信息
		stack := debug.Stack()
		log.Printf("PANIC RECOVERED:\n"+
			"Error: %v\n"+
			"Method: %s\n"+
			"Path: %s\n"+
			"Client IP: %s\n"+
			"User Agent: %s\n"+
			"Stack Trace:\n%s",
			err, method, path, clientIP, userAgent, string(stack))
		
		// 返回错误响应
		response := util.BuildErrorResp(types.ServiceError.WithMessage(fmt.Sprintf("Internal Server Error: %v", err)))
		ctx.JSON(500, response)
		ctx.Abort()
	}
}