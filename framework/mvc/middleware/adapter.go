package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// 函数签名适配器 - 实现两套中间件系统之间的转换

// BasicToMVC 将基础中间件转换为MVC中间件
func BasicToMVC(basicHandler Middleware) MiddlewareFunc {
	return func(ctx *mvccontext.EnhancedContext) {
		// 创建基础上下文
		c := context.Background()
		
		// 获取底层的 Hertz RequestContext
		hertzCtx := ctx.Request
		
		// 调用基础中间件
		basicHandler(c, hertzCtx)
	}
}

// HandlerFuncToMVC 将基础HandlerFunc转换为MVC中间件
func HandlerFuncToMVC(basicHandler HandlerFunc) MiddlewareFunc {
	return func(ctx *mvccontext.EnhancedContext) {
		// 创建基础中间件Context
		basicCtx := CreateBasicContext(ctx)
		
		// 调用基础处理器
		basicHandler(basicCtx)
		
		// 同步状态回MVC Context
		SyncContextState(basicCtx, ctx)
	}
}

// MVCToBasic 将MVC中间件转换为基础中间件（备用）
func MVCToBasic(mvcHandler MiddlewareFunc) Middleware {
	return func(c context.Context, hertzCtx *app.RequestContext) {
		// 创建MVC增强上下文
		enhancedCtx := mvccontext.NewContext(hertzCtx)
		
		// 调用MVC处理器
		mvcHandler(enhancedCtx)
	}
}

// CreateBasicContext 从MVC Context创建基础Context
func CreateBasicContext(mvcCtx *mvccontext.EnhancedContext) *Context {
	hertzCtx := mvcCtx.Request
	
	// 创建基础中间件引擎的Context
	engine := NewEngine()
	basicCtx := engine.NewContext(hertzCtx)
	
	// 同步现有数据
	for key, value := range mvcCtx.Keys {
		basicCtx.Set(key, value)
	}
	
	// 同步错误信息
	for _, err := range mvcCtx.GetErrors() {
		basicCtx.AddError(err)
	}
	
	return basicCtx
}

// SyncContextState 同步基础Context状态到MVC Context
func SyncContextState(basicCtx *Context, mvcCtx *mvccontext.EnhancedContext) {
	// 同步Keys
	for key, value := range basicCtx.Keys {
		mvcCtx.Set(key, value)
	}
	
	// 同步错误
	for _, err := range basicCtx.Errors {
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

// NewMiddlewareAdapter 创建中间件适配器
func NewMiddlewareAdapter(name string) *MiddlewareAdapter {
	return &MiddlewareAdapter{
		name:        name,
		basicEngine: NewEngine(),
		mvcManager:  NewMiddlewareManager(),
	}
}

// UseBasicMiddleware 在MVC系统中使用基础中间件
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

// UseBasicHandlerFunc 在MVC系统中使用基础HandlerFunc
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

// GetMVCManager 获取MVC管理器
func (adapter *MiddlewareAdapter) GetMVCManager() *MiddlewareManager {
	return adapter.mvcManager
}

// 全局适配器实例
var globalAdapter = NewMiddlewareAdapter("global")

// GetGlobalAdapter 获取全局适配器
func GetGlobalAdapter() *MiddlewareAdapter {
	return globalAdapter
}

// UseBasicInMVC 在MVC系统中使用基础中间件的便捷函数
func UseBasicInMVC(layer MiddlewareLayer, name string, handler Middleware, priority int) error {
	return globalAdapter.UseBasicMiddleware(layer, name, handler, priority)
}

// UseBasicHandlerInMVC 在MVC系统中使用基础HandlerFunc的便捷函数
func UseBasicHandlerInMVC(layer MiddlewareLayer, name string, handler HandlerFunc, priority int) error {
	return globalAdapter.UseBasicHandlerFunc(layer, name, handler, priority)
}