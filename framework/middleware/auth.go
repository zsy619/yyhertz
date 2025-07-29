package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// AuthMiddleware 认证中间件 - 进行身份验证
func AuthMiddleware(skipPaths ...string) Middleware {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}
	
	return func(c context.Context, ctx *app.RequestContext) {
		path := string(ctx.Path())
		
		if skipMap[path] {
			return
		}
		
		token := ctx.GetHeader("Authorization")
		if string(token) == "" {
			ctx.JSON(401, map[string]string{
				"error": "未授权访问",
			})
			ctx.Abort()
			return
		}
		
		if string(token) != "Bearer valid-token" {
			ctx.JSON(401, map[string]string{
				"error": "无效的令牌",
			})
			ctx.Abort()
			return
		}
	}
}