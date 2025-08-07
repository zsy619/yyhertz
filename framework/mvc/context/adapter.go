package context

import (
	"github.com/cloudwego/hertz/pkg/app"
)

// ContextAdapter 兼容性适配器
type ContextAdapter struct{}

// NewContextAdapter 创建新的上下文适配器 (兼容性API)
func NewContextAdapter() *ContextAdapter {
	return &ContextAdapter{}
}

// HertzToContext 将Hertz的RequestContext转换为Context (兼容性API)
func (a *ContextAdapter) HertzToContext(ctx *app.RequestContext) *Context {
	return NewContext(ctx)
}

// ContextToHertz 从Context获取Hertz的RequestContext (兼容性API)
func (a *ContextAdapter) ContextToHertz(ctx *Context) *app.RequestContext {
	if ctx == nil {
		return nil
	}
	return ctx.Request
}