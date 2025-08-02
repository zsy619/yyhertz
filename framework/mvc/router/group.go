package router

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/middleware"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// Group 路由组，增强以支持命名空间功能
type Group struct {
	router     *Router
	prefix     string
	middleware []middleware.HandlerFunc // 中间件链
	parent     *Group                   // 父路由组引用，支持嵌套
}

// NewGroup 创建路由组
func NewGroup(router *Router, prefix string) *Group {
	return &Group{
		router: router,
		prefix: prefix,
		parent: nil,
	}
}

// Group 创建子路由组
func (g *Group) Group(prefix string) *Group {
	newPrefix := g.prefix
	if newPrefix != "/" {
		newPrefix += prefix
	} else {
		newPrefix = prefix
	}

	return &Group{
		router:     g.router,
		prefix:     newPrefix,
		middleware: g.middleware, // 继承父组的中间件
		parent:     g,            // 设置父组引用
	}
}

// Use 添加中间件
func (g *Group) Use(middleware ...middleware.HandlerFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// RegisterController 在路由组中注册控制器
func (g *Group) RegisterController(path string, ctrl core.IController) {
	fullPath := g.prefix + path
	g.router.RegisterController(fullPath, ctrl)
}

// GetFullPrefix 获取包含所有父级前缀的完整路径
func (g *Group) GetFullPrefix() string {
	if g.parent == nil {
		return g.prefix
	}
	parentPrefix := g.parent.GetFullPrefix()
	if parentPrefix == "/" {
		return g.prefix
	}
	return parentPrefix + g.prefix
}

// GetAllMiddleware 获取包含所有父级中间件的完整中间件链
func (g *Group) GetAllMiddleware() []middleware.HandlerFunc {
	var allMiddleware []middleware.HandlerFunc
	if g.parent != nil {
		allMiddleware = append(allMiddleware, g.parent.GetAllMiddleware()...)
	}
	allMiddleware = append(allMiddleware, g.middleware...)
	return allMiddleware
}

// 便捷方法用于注册单个路由
func (g *Group) GET(path string, handler core.HandlerFunc) {
	g.addRoute("GET", path, handler)
}

func (g *Group) POST(path string, handler core.HandlerFunc) {
	g.addRoute("POST", path, handler)
}

func (g *Group) PUT(path string, handler core.HandlerFunc) {
	g.addRoute("PUT", path, handler)
}

func (g *Group) DELETE(path string, handler core.HandlerFunc) {
	g.addRoute("DELETE", path, handler)
}

func (g *Group) PATCH(path string, handler core.HandlerFunc) {
	g.addRoute("PATCH", path, handler)
}

func (g *Group) HEAD(path string, handler core.HandlerFunc) {
	g.addRoute("HEAD", path, handler)
}

func (g *Group) OPTIONS(path string, handler core.HandlerFunc) {
	g.addRoute("OPTIONS", path, handler)
}

func (g *Group) Any(path string, handler core.HandlerFunc) {
	g.addRoute("ANY", path, handler)
}

// addRoute 添加路由（内部方法）
func (g *Group) addRoute(method, path string, handler core.HandlerFunc) {
	fullPath := g.prefix + path

	// 如果有中间件，需要包装处理函数
	if len(g.middleware) > 0 {
		handler = g.wrapWithMiddleware(handler)
	}

	g.router.registerRoute(method, fullPath, handler)
}

// wrapWithMiddleware 使用中间件包装处理函数
func (g *Group) wrapWithMiddleware(handler core.HandlerFunc) core.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 创建中间件Context
		middlewareCtx := &middleware.Context{
			RequestContext: c,
		}
		
		// 执行中间件链
		for _, mw := range g.middleware {
			mw(middlewareCtx)
		}
		// 执行原始处理函数
		handler(ctx, c)
	}
}
