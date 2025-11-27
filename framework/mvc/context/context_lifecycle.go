package context

import (
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// ============= 重置和池化方法（优化版本） =============

// Reset 重置Context状态，准备复用
func (ctx *Context) Reset() {
	// 重置核心字段
	ctx.request = nil
	ctx.Context = nil
	ctx.params = ctx.params[:0]
	ctx.fullPath = ""

	// 清空sync.Map中的数据
	ctx.keys.Range(func(key, value any) bool {
		ctx.keys.Delete(key)
		return true
	})

	// 重置响应相关
	ctx.writer = nil

	// 重置中间件状态
	ctx.index = -1
	ctx.handlers = ctx.handlers[:0]
	atomic.StoreInt32(&ctx.aborted, 0) // 原子重置

	// 重置错误列表
	ctx.errMu.Lock()
	ctx.errors = ctx.errors[:0]
	ctx.errMu.Unlock()
}

// Release 释放Context到池中
func (ctx *Context) Release() {
	if ctx.pooled {
		defaultPool.Put(ctx)
		atomic.AddInt32(&poolSize, -1)
	}
}

func (ctx *Context) SetRequest(c *app.RequestContext) {
	ctx.request = c
}

// ============= 性能统计方法 =============

// Acquired 获取Context获取时间
func (ctx *Context) Acquired() time.Time {
	return ctx.acquired
}

// IsPooled 是否来自对象池
func (ctx *Context) IsPooled() bool {
	return ctx.pooled
}