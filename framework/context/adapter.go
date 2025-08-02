// Package context 适配层，提供兼容性支持
package context

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
)

// ContextAdapter 上下文适配器
type ContextAdapter struct{}

// NewContextAdapter 创建新的上下文适配器
func NewContextAdapter() *ContextAdapter {
	return &ContextAdapter{}
}

// HertzToContext 将Hertz的RequestContext转换为Context
func (a *ContextAdapter) HertzToContext(ctx *app.RequestContext) *Context {
	return NewContext(ctx)
}

// ContextToHertz 从Context获取Hertz的RequestContext
func (a *ContextAdapter) ContextToHertz(ctx *Context) *app.RequestContext {
	if ctx == nil {
		return nil
	}
	return ctx.RequestContext
}

// ============= 兼容性函数 =============

// ConvertToBeego 兼容性函数，实际返回统一的Context（保持向后兼容）
func ConvertToBeego(ctx *app.RequestContext) *Context {
	return NewContext(ctx)
}

// ConvertToHertz 从Context获取Hertz的RequestContext
func ConvertToHertz(ctx *Context) *app.RequestContext {
	if ctx == nil {
		return nil
	}
	return ctx.RequestContext
}

// ============= 中间件适配 =============

// HertzHandlerFunc Hertz处理函数类型
type HertzHandlerFunc func(context.Context, *app.RequestContext)

// ContextHandlerFunc Context处理函数类型
type ContextHandlerFunc func(*Context)

// AdaptHertzToContext 将Hertz处理函数适配为Context处理函数
func AdaptHertzToContext(hertzHandler HertzHandlerFunc) ContextHandlerFunc {
	return func(ctx *Context) {
		hertzHandler(context.Background(), ctx.RequestContext)
	}
}

// AdaptContextToHertz 将Context处理函数适配为Hertz处理函数
func AdaptContextToHertz(contextHandler ContextHandlerFunc) HertzHandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		enhancedCtx := NewContext(c)
		contextHandler(enhancedCtx)
	}
}

// ============= 中间件适配器 =============

// ContextMiddleware Context中间件函数类型
type ContextMiddleware func(*Context)

// AdaptMiddleware 适配中间件
func AdaptMiddleware(middleware ContextMiddleware) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		enhancedCtx := NewContext(c)
		middleware(enhancedCtx)
	}
}

// ============= 控制器适配 =============

// ContextControllerHandler Context控制器处理函数类型
type ContextControllerHandler func(*Context)

// AdaptController 适配控制器处理函数
func AdaptController(handler ContextControllerHandler) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		enhancedCtx := NewContext(c)
		handler(enhancedCtx)
	}
}

// ============= 上下文管理 =============

// GetContextFromHertz 从Hertz上下文中获取或创建Context
func GetContextFromHertz(ctx *app.RequestContext) *Context {
	// 尝试从Keys中获取已存在的Context
	if val, exists := ctx.Get("enhanced_context"); exists {
		if enhancedCtx, ok := val.(*Context); ok {
			return enhancedCtx
		}
	}

	// 如果不存在，创建新的Context并存储
	enhancedCtx := NewContext(ctx)
	ctx.Set("enhanced_context", enhancedCtx)
	return enhancedCtx
}

// SetContextToHertz 将Context设置到Hertz上下文中
func SetContextToHertz(hertzCtx *app.RequestContext, enhancedCtx *Context) {
	hertzCtx.Set("enhanced_context", enhancedCtx)
}

// ============= 响应复制 =============

// CopyResponseToHertz 将Context中的响应信息复制到Hertz Context
func CopyResponseToHertz(enhancedCtx *Context, hertzCtx *app.RequestContext) {
	if enhancedCtx == nil || enhancedCtx.Output == nil {
		return
	}

	// 复制状态码
	if enhancedCtx.Output.Status != 0 {
		hertzCtx.SetStatusCode(enhancedCtx.Output.Status)
	}

	// 复制响应头
	for key, value := range enhancedCtx.Output.headers {
		hertzCtx.Response.Header.Set(key, value)
	}
}

// ============= 功能适配函数 =============

// AdaptSession 适配Session功能
func AdaptSession(enhancedCtx *Context, sessionKey string) any {
	// 这里需要根据实际的session实现来适配
	if store, exists := enhancedCtx.Get("session"); exists {
		if s, ok := store.(interface{ Get(string) any }); ok {
			return s.Get(sessionKey)
		}
	}
	return nil
}

// AdaptTemplate 适配模板渲染功能
func AdaptTemplate(enhancedCtx *Context, templateName string, data any) error {
	// 这里需要根据实际的模板引擎来适配
	// 暂时返回未实现错误
	return errors.New("template rendering not implemented")
}

// AdaptError 适配错误处理功能
func AdaptError(enhancedCtx *Context, err error, code int) {
	enhancedCtx.Output.SetStatus(code)
	enhancedCtx.Output.JSON(map[string]any{
		"error": err.Error(),
		"code":  code,
	}, false, true)
}

// AdaptFileUpload 适配文件上传功能
func AdaptFileUpload(enhancedCtx *Context, fieldName string) ([]byte, string, error) {
	// 这里需要根据Hertz的文件上传API来实现
	// 暂时返回未实现错误
	return nil, "", errors.New("file upload not implemented")
}

// AdaptSecurity 适配安全功能
func AdaptSecurity(enhancedCtx *Context) {
	// 设置安全相关的响应头
	enhancedCtx.Output.Header("X-Content-Type-Options", "nosniff")
	enhancedCtx.Output.Header("X-Frame-Options", "DENY")
	enhancedCtx.Output.Header("X-XSS-Protection", "1; mode=block")
}
