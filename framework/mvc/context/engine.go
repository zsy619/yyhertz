package context

import "github.com/cloudwego/hertz/pkg/app"

type Engine struct {
	middleware []HandlerFunc
}

func NewEngine() *Engine {
	return &Engine{
		middleware: make([]HandlerFunc, 0),
	}
}

func (e *Engine) Middleware() []HandlerFunc {
	return e.middleware
}

func (e *Engine) SetMiddleware(middleware ...HandlerFunc) {
	e.middleware = middleware
}

// Use 添加全局中间件
func (e *Engine) Use(middleware ...HandlerFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// Use 添加中间件到上下文
func (e *Engine) UseContext(middleware ...HandlerFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// Use 添加中间件到上下文
func (c *Context) Use(middleware HandlerFunc) {
	c.handlers = append(c.handlers, middleware)
}

// NewContext 创建新的上下文
func (e *Engine) NewContext(c *app.RequestContext) *Context {
	ctx := &Context{
		handlers: e.Middleware(),
		index:    -1,
		engine:   e,
		errors:   make([]error, 0),
	}
	ctx.request = c

	return ctx
}
