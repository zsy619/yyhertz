package mvc

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// contextKey 是用于在context.Context中存储增强Context的key类型
type contextKey struct{}

// enhancedContextKey 是存储增强Context的key
var enhancedContextKey = &contextKey{}

// SimpleHandlerFunc 简化的处理函数类型，只接收context.Context
type SimpleHandlerFunc func(context.Context)

// DirectHandlerFunc 直接接收增强Context的处理函数类型
type DirectHandlerFunc func(*contextenhanced.Context)

// ============= Context传递机制 =============

// WithContext 将增强Context注入到context.Context中
func WithContext(ctx context.Context, enhancedCtx *contextenhanced.Context) context.Context {
	return context.WithValue(ctx, enhancedContextKey, enhancedCtx)
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
//   - *contextenhanced.Context: 增强Context实例，如果不存在则返回nil
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
func FromContext(ctx context.Context) *contextenhanced.Context {
	if enhancedCtx, ok := ctx.Value(enhancedContextKey).(*contextenhanced.Context); ok {
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
//   - *contextenhanced.Context: 增强Context实例
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
func MustFromContext(ctx context.Context) *contextenhanced.Context {
	enhancedCtx := FromContext(ctx)
	if enhancedCtx == nil {
		panic("enhanced context not found in context.Context")
	}
	return enhancedCtx
}

// AdaptHandlerToHertz 将YYHertz HandlerFunc适配为Hertz app.HandlerFunc
func AdaptHandlerToHertz(handler core.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建增强的上下文
		enhancedCtx := contextenhanced.NewContext(c)

		// 🔧 关键修复：复制路由参数从 Hertz RequestContext 到增强Context
		hertzParams := c.Params
		enhancedParams := make(contextenhanced.Params, len(hertzParams))
		for i, param := range hertzParams {
			enhancedParams[i] = contextenhanced.Param{
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
		handler(ctx, (*core.RequestContext)(c))

		// 执行 AfterExec 过滤器
		if HertzApp != nil {
			HertzApp.ExecuteFilters(enhancedCtx, core.AfterExec)

			// 执行 FinishRouter 过滤器
			HertzApp.ExecuteFilters(enhancedCtx, core.FinishRouter)
		}
	}
}

// AdaptSimpleHandlerToHertz 将简化的处理函数适配为Hertz app.HandlerFunc
func AdaptSimpleHandlerToHertz(handler SimpleHandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建增强的上下文
		enhancedCtx := contextenhanced.NewContext(c)

		// 🔧 关键修复：复制路由参数从 Hertz RequestContext 到增强Context
		hertzParams := c.Params
		enhancedParams := make(contextenhanced.Params, len(hertzParams))
		for i, param := range hertzParams {
			enhancedParams[i] = contextenhanced.Param{
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
func AdaptDirectHandlerToHertz(handler DirectHandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建增强的上下文
		enhancedCtx := contextenhanced.NewContext(c)

		// 🔧 关键修复：复制路由参数从 Hertz RequestContext 到增强Context
		// 将 Hertz 的路由参数复制到增强Context的Params字段
		hertzParams := c.Params
		enhancedParams := make(contextenhanced.Params, len(hertzParams))
		for i, param := range hertzParams {
			enhancedParams[i] = contextenhanced.Param{
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
func AdaptHandlersToHertz(handlers ...core.HandlerFunc) []app.HandlerFunc {
	hertzHandlers := make([]app.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hertzHandlers[i] = AdaptHandlerToHertz(handler)
	}
	return hertzHandlers
}

// AdaptSimpleHandlersToHertz 批量适配简化处理函数
func AdaptSimpleHandlersToHertz(handlers ...SimpleHandlerFunc) []app.HandlerFunc {
	hertzHandlers := make([]app.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hertzHandlers[i] = AdaptSimpleHandlerToHertz(handler)
	}
	return hertzHandlers
}

// AdaptDirectHandlersToHertz 批量适配Direct处理函数
func AdaptDirectHandlersToHertz(handlers ...DirectHandlerFunc) []app.HandlerFunc {
	hertzHandlers := make([]app.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hertzHandlers[i] = AdaptDirectHandlerToHertz(handler)
	}
	return hertzHandlers
}

// ============= HTTP路由注册方法 =============

// Any 注册任意HTTP方法的路由 (原有API，保持向后兼容)
func Any(relativePath string, handlers ...core.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.Any(relativePath, hertzHandlers...)
}

// SimpleAny 注册任意HTTP方法的路由 (简化API，只接收context.Context)
func SimpleAny(relativePath string, handlers ...SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.Any(relativePath, hertzHandlers...)
}

// DirectAny 注册任意HTTP方法的路由 (直接API，直接接收增强Context)
func DirectAny(relativePath string, handlers ...DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.Any(relativePath, hertzHandlers...)
}

// GET 注册GET路由 (原有API，保持向后兼容)
func GET(relativePath string, handlers ...core.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.GET(relativePath, hertzHandlers...)
}

// SimpleGET 注册GET路由 (简化API，只接收context.Context)
func SimpleGET(relativePath string, handlers ...SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.GET(relativePath, hertzHandlers...)
}

// DirectGET 注册GET路由 (直接API，直接接收增强Context)
func DirectGET(relativePath string, handlers ...DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.GET(relativePath, hertzHandlers...)
}

// POST 注册POST路由 (原有API，保持向后兼容)
func POST(relativePath string, handlers ...core.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.POST(relativePath, hertzHandlers...)
}

// SimplePOST 注册POST路由 (简化API，只接收context.Context)
func SimplePOST(relativePath string, handlers ...SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.POST(relativePath, hertzHandlers...)
}

// DirectPOST 注册POST路由 (直接API，直接接收增强Context)
func DirectPOST(relativePath string, handlers ...DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.POST(relativePath, hertzHandlers...)
}

// PUT 注册PUT路由 (原有API，保持向后兼容)
func PUT(relativePath string, handlers ...core.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.PUT(relativePath, hertzHandlers...)
}

// SimplePUT 注册PUT路由 (简化API，只接收context.Context)
func SimplePUT(relativePath string, handlers ...SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.PUT(relativePath, hertzHandlers...)
}

// DirectPUT 注册PUT路由 (直接API，直接接收增强Context)
func DirectPUT(relativePath string, handlers ...DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.PUT(relativePath, hertzHandlers...)
}

// DELETE 注册DELETE路由 (原有API，保持向后兼容)
func DELETE(relativePath string, handlers ...core.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.DELETE(relativePath, hertzHandlers...)
}

// SimpleDELETE 注册DELETE路由 (简化API，只接收context.Context)
func SimpleDELETE(relativePath string, handlers ...SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.DELETE(relativePath, hertzHandlers...)
}

// DirectDELETE 注册DELETE路由 (直接API，直接接收增强Context)
func DirectDELETE(relativePath string, handlers ...DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.DELETE(relativePath, hertzHandlers...)
}

// PATCH 注册PATCH路由 (原有API，保持向后兼容)
func PATCH(relativePath string, handlers ...core.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.PATCH(relativePath, hertzHandlers...)
}

// SimplePATCH 注册PATCH路由 (简化API，只接收context.Context)
func SimplePATCH(relativePath string, handlers ...SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.PATCH(relativePath, hertzHandlers...)
}

// DirectPATCH 注册PATCH路由 (直接API，直接接收增强Context)
func DirectPATCH(relativePath string, handlers ...DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.PATCH(relativePath, hertzHandlers...)
}

// HEAD 注册HEAD路由 (原有API，保持向后兼容)
func HEAD(relativePath string, handlers ...core.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.HEAD(relativePath, hertzHandlers...)
}

// SimpleHEAD 注册HEAD路由 (简化API，只接收context.Context)
func SimpleHEAD(relativePath string, handlers ...SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.HEAD(relativePath, hertzHandlers...)
}

// DirectHEAD 注册HEAD路由 (直接API，直接接收增强Context)
func DirectHEAD(relativePath string, handlers ...DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.HEAD(relativePath, hertzHandlers...)
}

// OPTIONS 注册OPTIONS路由 (原有API，保持向后兼容)
func OPTIONS(relativePath string, handlers ...core.HandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptHandlersToHertz(handlers...)
	return HertzApp.OPTIONS(relativePath, hertzHandlers...)
}

// SimpleOPTIONS 注册OPTIONS路由 (简化API，只接收context.Context)
func SimpleOPTIONS(relativePath string, handlers ...SimpleHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptSimpleHandlersToHertz(handlers...)
	return HertzApp.OPTIONS(relativePath, hertzHandlers...)
}

// DirectOPTIONS 注册OPTIONS路由 (直接API，直接接收增强Context)
func DirectOPTIONS(relativePath string, handlers ...DirectHandlerFunc) route.IRoutes {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	hertzHandlers := AdaptDirectHandlersToHertz(handlers...)
	return HertzApp.OPTIONS(relativePath, hertzHandlers...)
}

// ============= 中间件相关方法 =============

// Use 为全局应用添加中间件
func Use(middlewares ...core.HandlerFunc) {
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

// Group 创建子路由组（简化版实现）
func Group(relativePath string) *RouterGroup {
	if HertzApp == nil {
		panic("HertzApp is not initialized")
	}

	// 创建一个简单的路由组，由于RouterGroup已在其他地方定义，
	// 这里直接返回一个兼容的类型
	// 注意：这个实现需要根据实际的RouterGroup结构来调整
	return &RouterGroup{}
}
