package mvc

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// contextKey 是用于在context.Context中存储增强Context的key类型
type contextKey struct{}

// mvcContextKey 是存储增强Context的key
var mvcContextKey = &contextKey{}

// ============= Context传递机制 =============

// WithContext 将增强Context注入到context.Context中
func WithContext(ctx context.Context, enhancedCtx *mvcContext.Context) context.Context {
	return context.WithValue(ctx, mvcContextKey, enhancedCtx)
}

// FromContext 从标准context.Context中获取增强Context
//
// 该函数安全地从标准context.Context中提取增强Context实例。
// 如果提取失败或类型不匹配，返回nil。
//
// 参数：
//   - ctx: context.Context - 包含增强Context的标准context
//
// 返回值：
//   - *mvcContext.Context: 增强Context实例，如果不存在则返回nil
//
// 使用示例：
//
//	func simpleHandler(ctx context.Context) {
//		enhancedCtx := mvc.FromContext(ctx)
//		if enhancedCtx == nil {
//			// 处理错误情况
//			return
//		}
//		enhancedCtx.JSON(200, map[string]string{"status": "ok"})
//	}
//
// 注意事项：
//   - 在SimpleHandlerFunc中使用时，需要检查返回值是否为nil
//   - 只有通过YYHertz路由注册的处理函数才能正确获取增强Context
func FromContext(ctx context.Context) *mvcContext.Context {
	if enhancedCtx, ok := ctx.Value(mvcContextKey).(*mvcContext.Context); ok {
		return enhancedCtx
	}
	return nil
}

// MustFromContext 强制从标准context.Context中获取增强Context
//
// 该函数与FromContext类似，但在无法获取增强Context时会触发panic。
// 适用于确保增强Context必须存在的场景。
//
// 参数：
//   - ctx: context.Context - 包含增强Context的标准context
//
// 返回值：
//   - *mvcContext.Context: 增强Context实例
//
// Panic情况：
//   - 当context中不存在增强Context时会触发panic
//
// 使用示例：
//
//	func criticalHandler(ctx context.Context) {
//		// 确保增强Context存在，否则程序崩溃
//		enhancedCtx := mvc.MustFromContext(ctx)
//		enhancedCtx.JSON(200, map[string]string{"status": "ok"})
//	}
//
// 注意事项：
//   - 仅在确保增强Context必须存在的情况下使用
//   - 建议在大多数情况下使用FromContext并检查返回值
//   - panic会中断请求处理，需要有适当的恢复机制
func MustFromContext(ctx context.Context) *mvcContext.Context {
	enhancedCtx := FromContext(ctx)
	if enhancedCtx == nil {
		panic("enhanced context not found in context.Context")
	}
	return enhancedCtx
}

// AdaptHandlerToHertz 将YYHertz HandlerFunc适配为Hertz app.HandlerFunc
func AdaptHandlerToHertz(handler define.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建增强的上下文
		enhancedCtx := mvcContext.NewContext(c)

		// 🔧 关键修复：复制路由参数从 Hertz RequestContext 到增强Context
		hertzParams := c.Params
		enhancedParams := make(mvcContext.Params, len(hertzParams))
		for i, param := range hertzParams {
			enhancedParams[i] = mvcContext.Param{
				Key:   param.Key,
				Value: param.Value,
			}
		}
		enhancedCtx.SetParams(enhancedParams)

		// 执行 BeforeStatic 过滤器
		if HertzApp != nil {
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeStatic)
			if enhancedCtx.IsAborted() {
				return
			}

			// 执行 BeforeRouter 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeRouter)
			if enhancedCtx.IsAborted() {
				return
			}

			// 执行 BeforeExec 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeExec)
			if enhancedCtx.IsAborted() {
				return
			}
		}

		// 调用原始处理函数
		handler(ctx, (*define.RequestContext)(c))

		// 执行 AfterExec 过滤器
		if HertzApp != nil {
			HertzApp.ExecuteFilters(enhancedCtx, core.AfterExec)

			// 执行 FinishRouter 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.FinishRouter)
		}
	}
}

// AdaptSimpleHandlerToHertz 将简化的处理函数适配为Hertz app.HandlerFunc
func AdaptSimpleHandlerToHertz(handler define.SimpleHandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建增强的上下文
		enhancedCtx := mvcContext.NewContext(c)

		// 🔧 关键修复：复制路由参数从 Hertz RequestContext 到增强Context
		hertzParams := c.Params
		enhancedParams := make(mvcContext.Params, len(hertzParams))
		for i, param := range hertzParams {
			enhancedParams[i] = mvcContext.Param{
				Key:   param.Key,
				Value: param.Value,
			}
		}
		enhancedCtx.SetParams(enhancedParams)

		// 执行 BeforeStatic 过滤器
		if HertzApp != nil {
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeStatic)
			if enhancedCtx.IsAborted() {
				return
			}

			// 执行 BeforeRouter 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeRouter)
			if enhancedCtx.IsAborted() {
				return
			}

			// 执行 BeforeExec 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeExec)
			if enhancedCtx.IsAborted() {
				return
			}
		}

		// 将增强Context注入到context.Context中
		ctxWithEnhanced := WithContext(ctx, enhancedCtx)

		// 调用简化的处理函数
		handler(ctxWithEnhanced)

		// 执行 AfterExec 过滤器
		if HertzApp != nil {
			HertzApp.ExecuteFilters(enhancedCtx, core.AfterExec)

			// 执行 FinishRouter 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.FinishRouter)
		}
	}
}

// AdaptDirectHandlerToHertz 将Direct处理函数适配为Hertz app.HandlerFunc
func AdaptDirectHandlerToHertz(handler define.DirectHandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建增强的上下文
		enhancedCtx := mvcContext.NewContext(c)

		// 🔧 关键修复：复制路由参数从 Hertz RequestContext 到增强Context
		// 将 Hertz 的路由参数复制到增强Context的Params字段
		hertzParams := c.Params
		enhancedParams := make(mvcContext.Params, len(hertzParams))
		for i, param := range hertzParams {
			enhancedParams[i] = mvcContext.Param{
				Key:   param.Key,
				Value: param.Value,
			}
		}
		enhancedCtx.SetParams(enhancedParams)

		// 执行 BeforeStatic 过滤器
		if HertzApp != nil {
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeStatic)
			if enhancedCtx.IsAborted() {
				return
			}

			// 执行 BeforeRouter 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeRouter)
			if enhancedCtx.IsAborted() {
				return
			}

			// 执行 BeforeExec 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.BeforeExec)
			if enhancedCtx.IsAborted() {
				return
			}
		}

		// 直接调用处理函数，传递增强Context
		handler(enhancedCtx)

		// 执行 AfterExec 过滤器
		if HertzApp != nil {
			HertzApp.ExecuteFilters(enhancedCtx, core.AfterExec)

			// 执行 FinishRouter 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.FinishRouter)
		}
	}
}

// AdaptHandlersToHertz 批量适配处理函数
func AdaptHandlersToHertz(handlers ...define.HandlerFunc) []app.HandlerFunc {
	hertzHandlers := make([]app.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hertzHandlers[i] = AdaptHandlerToHertz(handler)
	}
	return hertzHandlers
}

// AdaptSimpleHandlersToHertz 批量适配简化处理函数
func AdaptSimpleHandlersToHertz(handlers ...define.SimpleHandlerFunc) []app.HandlerFunc {
	hertzHandlers := make([]app.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hertzHandlers[i] = AdaptSimpleHandlerToHertz(handler)
	}
	return hertzHandlers
}

// AdaptDirectHandlersToHertz 批量适配Direct处理函数
func AdaptDirectHandlersToHertz(handlers ...define.DirectHandlerFunc) []app.HandlerFunc {
	hertzHandlers := make([]app.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hertzHandlers[i] = AdaptDirectHandlerToHertz(handler)
	}
	return hertzHandlers
}

// ============= HTTP路由注册方法 =============

// Any 注册任意HTTP方法的路由 (原有API，保持向后兼容)
func Any(relativePath string, handlers ...define.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.Any(relativePath, hertzHandlers...)
}

// AnySimple 注册任意HTTP方法的路由 (简化API，只接收context.Context)
func AnySimple(relativePath string, handlers ...define.SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.Any(relativePath, hertzHandlers...)
}

// AnyDirect 注册任意HTTP方法的路由 (直接API，直接接收增强Context)
func AnyDirect(relativePath string, handlers ...define.DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.Any(relativePath, hertzHandlers...)
}

// GET 注册GET路由 (原有API，保持向后兼容)
func GET(relativePath string, handlers ...define.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.GET(relativePath, hertzHandlers...)
}

// GETSimple 注册GET路由 (简化API，只接收context.Context)
func GETSimple(relativePath string, handlers ...define.SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.GET(relativePath, hertzHandlers...)
}

// GETDirect 注册GET路由 (直接API，直接接收增强Context)
func GETDirect(relativePath string, handlers ...define.DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.GET(relativePath, hertzHandlers...)
}

// POST 注册POST路由 (原有API，保持向后兼容)
func POST(relativePath string, handlers ...define.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.POST(relativePath, hertzHandlers...)
}

// POSTSimple 注册POST路由 (简化API，只接收context.Context)
func POSTSimple(relativePath string, handlers ...define.SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.POST(relativePath, hertzHandlers...)
}

// POSTDirect 注册POST路由 (直接API，直接接收增强Context)
func POSTDirect(relativePath string, handlers ...define.DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.POST(relativePath, hertzHandlers...)
}

// PUT 注册PUT路由 (原有API，保持向后兼容)
func PUT(relativePath string, handlers ...define.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.PUT(relativePath, hertzHandlers...)
}

// PUTSimple 注册PUT路由 (简化API，只接收context.Context)
func PUTSimple(relativePath string, handlers ...define.SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.PUT(relativePath, hertzHandlers...)
}

// PUTDirect 注册PUT路由 (直接API，直接接收增强Context)
func PUTDirect(relativePath string, handlers ...define.DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.PUT(relativePath, hertzHandlers...)
}

// DELETE 注册DELETE路由 (原有API，保持向后兼容)
func DELETE(relativePath string, handlers ...define.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.DELETE(relativePath, hertzHandlers...)
}

// DELETESimple 注册DELETE路由 (简化API，只接收context.Context)
func DELETESimple(relativePath string, handlers ...define.SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.DELETE(relativePath, hertzHandlers...)
}

// DELETEDirect 注册DELETE路由 (直接API，直接接收增强Context)
func DELETEDirect(relativePath string, handlers ...define.DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.DELETE(relativePath, hertzHandlers...)
}

// PATCH 注册PATCH路由 (原有API，保持向后兼容)
func PATCH(relativePath string, handlers ...define.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.PATCH(relativePath, hertzHandlers...)
}

// PATCHSimple 注册PATCH路由 (简化API，只接收context.Context)
func PATCHSimple(relativePath string, handlers ...define.SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.PATCH(relativePath, hertzHandlers...)
}

// PATCHDirect 注册PATCH路由 (直接API，直接接收增强Context)
func PATCHDirect(relativePath string, handlers ...define.DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.PATCH(relativePath, hertzHandlers...)
}

// HEAD 注册HEAD路由 (原有API，保持向后兼容)
func HEAD(relativePath string, handlers ...define.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.HEAD(relativePath, hertzHandlers...)
}

// HEADSimple 注册HEAD路由 (简化API，只接收context.Context)
func HEADSimple(relativePath string, handlers ...define.SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.HEAD(relativePath, hertzHandlers...)
}

// HEADDirect 注册HEAD路由 (直接API，直接接收增强Context)
func HEADDirect(relativePath string, handlers ...define.DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.HEAD(relativePath, hertzHandlers...)
}

// OPTIONS 注册OPTIONS路由 (原有API，保持向后兼容)
func OPTIONS(relativePath string, handlers ...define.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.OPTIONS(relativePath, hertzHandlers...)
}

// OPTIONSSimple 注册OPTIONS路由 (简化API，只接收context.Context)
func OPTIONSSimple(relativePath string, handlers ...define.SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.OPTIONS(relativePath, hertzHandlers...)
}

// OPTIONSDirect 注册OPTIONS路由 (直接API，直接接收增强Context)
func OPTIONSDirect(relativePath string, handlers ...define.DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.OPTIONS(relativePath, hertzHandlers...)
}

// ============= 中间件相关方法 =============

// Use 为全局应用添加中间件
func Use(middlewares ...define.HandlerFunc) {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	// 直接使用YYHertz的HandlerFunc，因为App.Use接受的就是HandlerFunc
	HertzApp.Use(middlewares...)
}

// ============= 辅助方法 =============

// Static 注册静态文件路由
func Static(relativePath, root string) {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	HertzApp.SetStaticPath(root, relativePath)
}

// StaticFile 注册单个静态文件路由
func StaticFile(relativePath, filepath string) {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	HertzApp.StaticFile(relativePath, filepath)
}

// ============= 路由组扩展方法 =============

// Group 创建子路由组（支持多处理器类型）
func Group(relativePath string) *RouterGroup {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	// 创建新的路由器实例
	router := NewRouter(HertzApp)

	// 创建支持多处理器的路由组
	return NewGroup(router, relativePath)
}
