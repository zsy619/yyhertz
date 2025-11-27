package define

import (
	"context"

	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
)

type Middleware func(c context.Context, ctx *RequestContext)

type MiddlewareFunc func(*mvcContext.Context)
