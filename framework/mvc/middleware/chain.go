// Package middleware 增强的中间件系统
// 借鉴Gin框架的中间件设计，提供链式调用和Next()机制
package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// HandlerFunc 中间件处理函数类型
type HandlerFunc func(*Context)

// Context 增强的请求上下文，类似Gin的Context
type Context struct {
	*app.RequestContext
	handlers []HandlerFunc
	index    int8
	engine   *Engine
	Keys     map[string]any
	Errors   []error
}

// Engine 中间件引擎
type Engine struct {
	middleware []HandlerFunc
}

// NewEngine 创建新的中间件引擎
func NewEngine() *Engine {
	return &Engine{
		middleware: make([]HandlerFunc, 0),
	}
}

// Use 添加全局中间件
func (e *Engine) Use(middleware ...HandlerFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// Use 添加中间件到上下文
func (c *Context) Use(middleware HandlerFunc) {
	c.handlers = append(c.handlers, middleware)
}

// NewContext 创建新的上下文
func (e *Engine) NewContext(c *app.RequestContext) *Context {
	return &Context{
		RequestContext: c,
		handlers:       e.middleware,
		index:          -1,
		engine:         e,
		Keys:           make(map[string]any),
		Errors:         make([]error, 0),
	}
}

// Next 执行下一个中间件或处理器
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort 终止中间件链执行
func (c *Context) Abort() {
	c.index = 63 // 设置为最大值，阻止后续中间件执行
}

// AbortWithStatus 终止执行并设置状态码
func (c *Context) AbortWithStatus(code int) {
	c.SetStatusCode(code)
	c.Abort()
}

// AbortWithError 终止执行并记录错误
func (c *Context) AbortWithError(code int, err error) error {
	c.AbortWithStatus(code)
	return c.AddError(err)
}

// IsAborted 检查是否已终止
func (c *Context) IsAborted() bool {
	return c.index >= 63
}

// Set 设置键值对
func (c *Context) Set(key string, value any) {
	c.Keys[key] = value
}

// Get 获取值
func (c *Context) Get(key string) (any, bool) {
	value, exists := c.Keys[key]
	return value, exists
}

// MustGet 获取值，不存在则panic
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString 获取字符串值
func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool 获取布尔值
func (c *Context) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt 获取整数值
func (c *Context) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 获取64位整数值
func (c *Context) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetFloat64 获取浮点数值
func (c *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// AddError 添加错误到上下文
func (c *Context) AddError(err error) error {
	if err == nil {
		return nil
	}
	c.Errors = append(c.Errors, err)
	return err
}

// GetErrors 获取所有错误
func (c *Context) GetErrors() []error {
	return c.Errors
}

// HasErrors 检查是否有错误
func (c *Context) HasErrors() bool {
	return len(c.Errors) > 0
}

// LastError 获取最后一个错误
func (c *Context) LastError() error {
	if len(c.Errors) == 0 {
		return nil
	}
	return c.Errors[len(c.Errors)-1]
}

// Copy 复制上下文（用于异步处理）
func (c *Context) Copy() *Context {
	// 创建一个新的RequestContext
	newReqCtx := &app.RequestContext{}
	// 手动复制必要的字段
	*newReqCtx = *c.RequestContext

	copied := &Context{
		RequestContext: newReqCtx,
		handlers:       c.handlers,
		index:          63, // 复制的上下文不应该执行中间件
		engine:         c.engine,
		Keys:           make(map[string]any),
		Errors:         make([]error, len(c.Errors)),
	}

	// 复制Keys
	for k, v := range c.Keys {
		copied.Keys[k] = v
	}

	// 复制Errors
	copy(copied.Errors, c.Errors)

	return copied
}

// WithContext 设置context.Context
func (c *Context) WithContext(ctx context.Context) {
	// Hertz的RequestContext没有WithContext方法，直接设置上下文
	// 这个方法主要用于兼容，实际使用中可能需要其他方式处理
}
