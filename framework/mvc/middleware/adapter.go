package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// 函数签名适配器 - 实现两套中间件系统之间的转换

// BasicToMVC 将基础中间件转换为符合 MVC 上下文的中间件函数。
//
// 参数:
//
//	basicHandler: 基础中间件函数，接受标准上下文和 Hertz 请求。
//
// 返回值:
//
//	MiddlewareFunc: 适配后的中间件函数，适用于 MVC 上下文。
//
// 功能说明:
//  1. 创建一个基础上下文。
//  2. 从 MVC 上下文中提取底层的 Hertz 请求。
//  3. 调用传入的基础中间件函数，传递基础上下文和 Hertz 请求。
//
// 适用场景:
//
//	用于在 MVC 框架中集成标准中间件逻辑。
func BasicToMVC(basicHandler Middleware) MiddlewareFunc {
	return func(ctx *mvcContext.Context) {
		// 创建基础上下文
		c := context.Background()

		// 获取底层的 Hertz Request
		hertzCtx := ctx.Request()

		// 调用基础中间件
		basicHandler(c, hertzCtx)
	}
}

// HandlerFuncToMVC 将基础的 HandlerFunc 转换为 MVC 中间件函数。
//
// 该函数接收一个基础的 HandlerFunc，并返回一个 MiddlewareFunc，用于在 MVC 上下文中执行。
// 转换后的中间件会：
//  1. 创建一个基础的中间件 Context。
//  2. 调用传入的基础处理器。
//  3. 将状态同步回 MVC Context。
//
// 参数:
//
//	basicHandler: 基础处理器函数，用于处理请求。
//
// 返回值:
//
//	MiddlewareFunc: 转换后的 MVC 中间件函数。
func HandlerFuncToMVC(basicHandler HandlerFunc) MiddlewareFunc {
	return func(ctx *mvcContext.Context) {
		// 创建基础中间件Context
		basicCtx := CreateBasicContext(ctx)

		// 调用基础处理器
		basicHandler(basicCtx)

		// 同步状态回MVC Context
		SyncContextState(basicCtx, ctx)
	}
}

// MVCToBasic 将MVC中间件转换为基础中间件（备用）
//
// MVCToBasic 将 MVC 风格的中间件函数转换为标准的 Hertz 中间件。
//
// 参数:
//
//	mvcHandler: 一个 MVC 风格的中间件函数，接收 mvcContext.Context 作为参数。
//
// 返回值:
//
//	一个标准的 Hertz 中间件函数，接收 context.Context 和 *app.RequestContext 作为参数。
//
// 功能:
//  1. 创建一个 MVC 增强上下文 (mvcContext.Context)。
//  2. 调用传入的 MVC 处理器 (mvcHandler) 并传递增强上下文。
//
// 适用场景:
//
//	用于将 MVC 风格的中间件集成到 Hertz 框架的标准中间件链中。
func MVCToBasic(mvcHandler MiddlewareFunc) Middleware {
	return func(c context.Context, hertzCtx *app.RequestContext) {
		// 创建MVC增强上下文
		enhancedCtx := mvcContext.NewContext(hertzCtx)

		// 调用MVC处理器
		mvcHandler(enhancedCtx)
	}
}

// CreateBasicContext 从MVC Context创建基础Context
func CreateBasicContext(mvcCtx *mvcContext.Context) *Context {
	hertzCtx := mvcCtx.Request()

	// 创建基础中间件引擎的Context
	engine := mvcContext.NewEngine()
	basicCtx := engine.NewContext(hertzCtx)

	// 同步现有数据
	maps := mvcCtx.ParamMap()
	for key, value := range maps {
		basicCtx.Set(key, value)
	}

	// 同步错误信息
	for _, err := range mvcCtx.GetErrors() {
		basicCtx.AddError(err)
	}

	return basicCtx
}

// SyncContextState 同步基础Context状态到MVC Context
func SyncContextState(basicCtx *Context, mvcCtx *mvcContext.Context) {
	// 同步Keys
	keys := basicCtx.GetKeys()
	keys.Range(func(k, v any) bool {
		if ks, ok := k.(string); ok {
			mvcCtx.Set(ks, v)
		}
		return true
	})

	// 同步错误
	errs := basicCtx.GetErrors()
	for _, err := range errs {
		mvcCtx.AddError(err)
	}

	// 同步状态
	if basicCtx.IsAborted() {
		mvcCtx.Abort()
	}
}

// MiddlewareAdapter 中间件适配器结构
type MiddlewareAdapter struct {
	name        string
	basicEngine *Engine
	mvcManager  *MiddlewareManager
}

// NewMiddlewareAdapter 创建一个新的 MiddlewareAdapter 实例。
//
// 参数:
//
//	name: 中间件适配器的名称，用于标识和管理。
//
// 返回值:
//
//	*MiddlewareAdapter: 返回一个初始化好的 MiddlewareAdapter 指针，包含基础引擎和中间件管理器。
func NewMiddlewareAdapter(name string) *MiddlewareAdapter {
	return &MiddlewareAdapter{
		name:        name,
		basicEngine: mvcContext.NewEngine(),
		mvcManager:  NewMiddlewareManager(),
	}
}

// UseBasicMiddleware 将基础中间件转换为MVC中间件并注册到MVC系统中。
//
// 参数:
//   - layer: 中间件所属的层级。
//   - name: 中间件的名称。
//   - handler: 基础中间件处理函数。
//   - priority: 中间件的优先级。
//
// 返回值:
//   - error: 如果注册或使用中间件时发生错误，返回错误信息；否则返回nil。
//
// 说明:
//   - 该函数会将基础中间件转换为MVC中间件，并将其注册到MVC系统中。
//   - 注册成功后，会在指定的层级和优先级下使用该中间件。
func (adapter *MiddlewareAdapter) UseBasicMiddleware(layer MiddlewareLayer, name string, handler Middleware, priority int) error {
	// 转换为MVC中间件
	mvcHandler := BasicToMVC(handler)

	// 注册到MVC系统
	err := adapter.mvcManager.RegisterCustom(name, mvcHandler, MiddlewareMetadata{
		Name:        name,
		Description: "Converted from basic middleware",
		Author:      "Adapter",
	})
	if err != nil {
		return err
	}

	// 使用中间件
	return adapter.mvcManager.UseCustom(layer, name, priority)
}

// UseBasicHandlerFunc 将基础的 HandlerFunc 转换为 MVC 中间件并注册到 MVC 系统中。
//
// 参数:
//   - layer: 中间件所属的层级（MiddlewareLayer 类型）。
//   - name: 中间件的名称，用于标识和引用。
//   - handler: 基础的 HandlerFunc 函数，将被转换为 MVC 中间件。
//   - priority: 中间件的优先级，数值越小优先级越高。
//
// 返回值:
//   - error: 如果注册或使用中间件过程中发生错误，返回错误信息；否则返回 nil。
//
// 功能说明:
//  1. 将基础的 HandlerFunc 转换为 MVC 中间件格式。
//  2. 注册到 MVC 系统中，并附带元数据（名称、描述、作者）。
//  3. 在指定的层级和优先级下使用该中间件。
//
// 注意:
//   - 如果注册或使用过程中发生错误，函数会立即返回错误。
func (adapter *MiddlewareAdapter) UseBasicHandlerFunc(layer MiddlewareLayer, name string, handler HandlerFunc, priority int) error {
	// 转换为MVC中间件
	mvcHandler := HandlerFuncToMVC(handler)

	// 注册到MVC系统
	err := adapter.mvcManager.RegisterCustom(name, mvcHandler, MiddlewareMetadata{
		Name:        name,
		Description: "Converted from basic HandlerFunc",
		Author:      "Adapter",
	})
	if err != nil {
		return err
	}

	// 使用中间件
	return adapter.mvcManager.UseCustom(layer, name, priority)
}

// GetMVCManager 返回适配器中的 MVC 管理器实例。
// 该方法用于获取当前适配器关联的 MiddlewareManager 对象，通常用于外部调用或进一步操作。
func (adapter *MiddlewareAdapter) GetMVCManager() *MiddlewareManager {
	return adapter.mvcManager
}

// 全局适配器实例
var globalAdapter = NewMiddlewareAdapter("global")

// GetGlobalAdapter 获取全局适配器
func GetGlobalAdapter() *MiddlewareAdapter {
	return globalAdapter
}

// UseBasicInMVC 注册一个基础的中间件到指定的 MVC 层中。
//
// 参数:
//
//	layer: 中间件所属的 MVC 层（如 Controller、Service 等）。
//	name: 中间件的名称，用于标识和调试。
//	handler: 中间件的处理函数，类型为 Middleware。
//	priority: 中间件的优先级，数值越小优先级越高。
//
// 返回:
//
//	error: 如果注册失败，返回错误信息；否则返回 nil。
//
// 示例:
//
//	err := UseBasicInMVC(ControllerLayer, "auth", AuthMiddleware, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
func UseBasicInMVC(layer MiddlewareLayer, name string, handler Middleware, priority int) error {
	return globalAdapter.UseBasicMiddleware(layer, name, handler, priority)
}

// UseBasicHandlerInMVC 注册一个基本的处理器函数到指定的中间件层。
// 参数：
//   - layer: 中间件层（MiddlewareLayer），指定处理器所属的层级。
//   - name: 处理器的名称，用于标识。
//   - handler: 处理器函数（HandlerFunc），实际执行的逻辑。
//   - priority: 优先级，数值越小优先级越高。
//
// 返回值：
//   - error: 如果注册失败，返回错误信息；否则返回 nil。
func UseBasicHandlerInMVC(layer MiddlewareLayer, name string, handler HandlerFunc, priority int) error {
	return globalAdapter.UseBasicHandlerFunc(layer, name, handler, priority)
}
