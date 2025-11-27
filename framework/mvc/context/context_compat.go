package context

import "github.com/cloudwego/hertz/pkg/app"

// ============= 兼容性访问器方法 =============
// 这些方法提供向后兼容性，允许现有代码无缝迁移

// Request 获取Request对象
func (ctx *Context) Request() *app.RequestContext {
	return ctx.request
}

// Writer 获取Writer对象
func (ctx *Context) Writer() ResponseWriter {
	return ctx.writer
}

// ResponseWriter 获取ResponseWriter (兼容性方法)
func (ctx *Context) ResponseWriter() ResponseWriter {
	return ctx.writer
}

// ============= 调试和监控方法 =============

// KeysCount 获取存储的键值对数量
func (ctx *Context) KeysCount() int {
	count := 0
	ctx.keys.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// ListKeys 列出所有键
func (ctx *Context) ListKeys() []string {
	var keys []string
	ctx.keys.Range(func(key, value any) bool {
		if k, ok := key.(string); ok {
			keys = append(keys, k)
		}
		return true
	})
	return keys
}