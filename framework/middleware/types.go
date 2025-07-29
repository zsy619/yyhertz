package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// Middleware 中间件函数类型定义
type Middleware func(c context.Context, ctx *app.RequestContext)