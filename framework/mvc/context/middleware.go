package context

import (
	"sync/atomic"
	"time"
)

// ============= 中间件执行方法 =============

// Next 执行下一个中间件
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		if !ctx.IsAborted() {
			ctx.handlers[ctx.index](ctx)
		}
		ctx.index++
	}
}

// Abort 中止执行
func (ctx *Context) Abort() {
	atomic.StoreInt32(&ctx.aborted, 1)
}

// IsAborted 是否已中止
func (ctx *Context) IsAborted() bool {
	return atomic.LoadInt32(&ctx.aborted) != 0
}

// AbortWithStatus 终止并设置状态码
func (ctx *Context) AbortWithStatus(code int) {
	ctx.Status(code)
	ctx.Abort()
}

// AbortWithStatusJSON 终止并返回JSON错误
func (ctx *Context) AbortWithStatusJSON(code int, jsonObj any) {
	ctx.Abort()
	ctx.JSON(code, jsonObj)
}

// AbortWithError 终止执行并记录错误
func (ctx *Context) AbortWithError(code int, err error) error {
	ctx.AbortWithStatus(code)
	if err != nil {
		ctx.AddError(err)
	}
	return err
}

// ============= 处理器链管理方法 =============

// SetHandlers 设置处理器链
func (ctx *Context) SetHandlers(handlers []HandlerFunc) {
	ctx.handlers = handlers
	ctx.index = -1
}

// GetHandlers 获取处理器链
func (ctx *Context) GetHandlers() []HandlerFunc {
	return ctx.handlers
}

// HandlerCount 获取处理器数量
func (ctx *Context) HandlerCount() int {
	return len(ctx.handlers)
}

// CurrentHandler 获取当前处理器索引
func (ctx *Context) CurrentHandler() int {
	return int(ctx.index)
}

// ============= 上下文超时控制方法 =============

// Done 获取请求完成通道 (用于超时控制)
func (ctx *Context) Done() <-chan struct{} {
	if ctx.Context != nil {
		return ctx.Context.Done()
	}
	// 如果没有父Context，返回一个永不关闭的通道
	return make(<-chan struct{})
}

// Deadline 获取截止时间
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	if ctx.Context != nil {
		return ctx.Context.Deadline()
	}
	return
}

// Err 获取上下文错误
func (ctx *Context) Err() error {
	if ctx.Context != nil {
		return ctx.Context.Err()
	}
	return nil
}

// Value 获取上下文值
func (ctx *Context) Value(key any) any {
	if ctx.Context != nil {
		return ctx.Context.Value(key)
	}
	return nil
}
