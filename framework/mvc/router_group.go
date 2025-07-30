package mvc

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
)

type RouterGroup struct {
	group *route.RouterGroup
}

func (a *App) Group(path string) *RouterGroup {
	return &RouterGroup{group: a.Hertz.Group(path)}
}

func (rg *RouterGroup) GET(path string, handler HandlerFunc) {
	rg.group.GET(path, AdaptHandler(handler))
}

func (rg *RouterGroup) POST(path string, handler HandlerFunc) {
	rg.group.POST(path, AdaptHandler(handler))
}

func (rg *RouterGroup) PUT(path string, handler HandlerFunc) {
	rg.group.PUT(path, AdaptHandler(handler))
}

func (rg *RouterGroup) DELETE(path string, handler HandlerFunc) {
	rg.group.DELETE(path, AdaptHandler(handler))
}

func (rg *RouterGroup) PATCH(path string, handler HandlerFunc) {
	rg.group.PATCH(path, AdaptHandler(handler))
}

func (rg *RouterGroup) HEAD(path string, handler HandlerFunc) {
	rg.group.HEAD(path, AdaptHandler(handler))
}

func (rg *RouterGroup) OPTIONS(path string, handler HandlerFunc) {
	rg.group.OPTIONS(path, AdaptHandler(handler))
}

func (rg *RouterGroup) Any(path string, handler HandlerFunc) {
	rg.group.Any(path, AdaptHandler(handler))
}

func (rg *RouterGroup) Handle(method, path string, handler HandlerFunc) {
	rg.group.Handle(method, path, AdaptHandler(handler))
}

func (rg *RouterGroup) Static(prefix, root string) {
	rg.group.Static(prefix, root)
}

func (rg *RouterGroup) StaticFile(path, file string) {
	rg.group.StaticFile(path, file)
}

func (rg *RouterGroup) StaticFS(prefix string, fs *app.FS) {
	rg.group.StaticFS(prefix, fs)
}

func (rg *RouterGroup) Use(middlewares ...HandlerFunc) {
	if len(middlewares) == 0 {
		return
	}
	// Convert HandlerFunc to app.HandlerFunc
	middlewaresAdapted := make([]app.HandlerFunc, len(middlewares))
	for i, middleware := range middlewares {
		middlewaresAdapted[i] = AdaptHandler(middleware)
	}
	rg.group.Use(middlewaresAdapted...)
}

func (rg *RouterGroup) Group(path string) *RouterGroup {
	return &RouterGroup{group: rg.group.Group(path)}
}
