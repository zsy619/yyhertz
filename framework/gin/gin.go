// Package gin 提供Gin风格的API
// 统一context类型，解决类型冲突问题
package gin

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/zsy619/yyhertz/framework/binding"
	"github.com/zsy619/yyhertz/framework/render"
)

// HandlerFunc Gin风格的处理函数
type HandlerFunc func(*Context)

// Context Gin风格的上下文
type Context struct {
	*app.RequestContext

	// 参数存储
	Params Params
	Keys   map[string]any
	Errors []error

	// 中间件
	handlers []HandlerFunc
	index    int8

	// 引擎引用
	engine *Engine
}

// Params 路由参数
type Params []Param

// Param 单个路由参数
type Param struct {
	Key   string
	Value string
}

// Get 获取参数值
func (ps Params) Get(name string) (string, bool) {
	for _, p := range ps {
		if p.Key == name {
			return p.Value, true
		}
	}
	return "", false
}

// ByName 根据名称获取参数值
func (ps Params) ByName(name string) string {
	va, _ := ps.Get(name)
	return va
}

// Engine Gin风格的引擎
type Engine struct {
	*server.Hertz
	RouterGroup
	middleware []HandlerFunc
	noRoute    []HandlerFunc
	noMethod   []HandlerFunc
}

// RouterGroup 路由组
type RouterGroup struct {
	handlers []HandlerFunc
	basePath string
	engine   *Engine
	root     bool
}

// IRoutes 路由接口
type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	Any(string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
	PATCH(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
	OPTIONS(string, ...HandlerFunc) IRoutes
	HEAD(string, ...HandlerFunc) IRoutes

	StaticFile(string, string) IRoutes
	Static(string, string) IRoutes
	StaticFS(string, http.FileSystem) IRoutes
}

// New 创建新的Gin引擎
func New() *Engine {
	engine := &Engine{
		RouterGroup: RouterGroup{
			handlers: nil,
			basePath: "/",
			root:     true,
		},
	}
	engine.RouterGroup.engine = engine
	engine.Hertz = server.Default()
	return engine
}

// Default 创建带默认中间件的引擎
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

// Use 添加中间件
func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.handlers = append(group.handlers, middleware...)
	return group.returnObj()
}

// Group 创建路由组
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

// Handle 处理路由
func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	finalHandlers := group.combineHandlers(handlers)

	// 转换为Hertz处理函数
	hertzHandler := func(ctx context.Context, req *app.RequestContext) {
		ginCtx := group.engine.createContext(req, finalHandlers)
		ginCtx.Next()
	}

	// 注册到Hertz路由
	switch httpMethod {
	case "GET":
		group.engine.Hertz.GET(absolutePath, hertzHandler)
	case "POST":
		group.engine.Hertz.POST(absolutePath, hertzHandler)
	case "PUT":
		group.engine.Hertz.PUT(absolutePath, hertzHandler)
	case "DELETE":
		group.engine.Hertz.DELETE(absolutePath, hertzHandler)
	case "PATCH":
		group.engine.Hertz.PATCH(absolutePath, hertzHandler)
	case "HEAD":
		group.engine.Hertz.HEAD(absolutePath, hertzHandler)
	case "OPTIONS":
		group.engine.Hertz.OPTIONS(absolutePath, hertzHandler)
	default:
		group.engine.Hertz.Handle(httpMethod, absolutePath, hertzHandler)
	}

	return group.returnObj()
}

// GET 注册GET路由
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("GET", relativePath, handlers...)
}

// POST 注册POST路由
func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("POST", relativePath, handlers...)
}

// DELETE 注册DELETE路由
func (group *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("DELETE", relativePath, handlers...)
}

// PATCH 注册PATCH路由
func (group *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("PATCH", relativePath, handlers...)
}

// PUT 注册PUT路由
func (group *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("PUT", relativePath, handlers...)
}

// OPTIONS 注册OPTIONS路由
func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("OPTIONS", relativePath, handlers...)
}

// HEAD 注册HEAD路由
func (group *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("HEAD", relativePath, handlers...)
}

// Any 注册所有HTTP方法路由
func (group *RouterGroup) Any(relativePath string, handlers ...HandlerFunc) IRoutes {
	group.Handle("GET", relativePath, handlers...)
	group.Handle("POST", relativePath, handlers...)
	group.Handle("PUT", relativePath, handlers...)
	group.Handle("PATCH", relativePath, handlers...)
	group.Handle("HEAD", relativePath, handlers...)
	group.Handle("OPTIONS", relativePath, handlers...)
	group.Handle("DELETE", relativePath, handlers...)
	return group.returnObj()
}

// StaticFile 注册静态文件路由
func (group *RouterGroup) StaticFile(relativePath, filepath string) IRoutes {
	handler := func(c *Context) {
		c.File(filepath)
	}
	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
	return group.returnObj()
}

// Static 注册静态文件目录路由
func (group *RouterGroup) Static(relativePath, root string) IRoutes {
	return group.StaticFS(relativePath, http.Dir(root))
}

// StaticFS 注册文件系统路由
func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters cannot be used when serving a static folder")
	}
	handler := func(c *Context) {
		file := c.Param("filepath")
		if file == "" {
			file = "/"
		}
		c.File(file)
	}
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
	group.HEAD(urlPattern, handler)
	return group.returnObj()
}

// NoRoute 设置404处理器
func (engine *Engine) NoRoute(handlers ...HandlerFunc) {
	engine.noRoute = handlers
}

// NoMethod 设置405处理器
func (engine *Engine) NoMethod(handlers ...HandlerFunc) {
	engine.noMethod = handlers
}

// Run 启动服务器
func (engine *Engine) Run(addr ...string) error {
	address := resolveAddress(addr)
	fmt.Printf("Gin server listening on %s\n", address)
	engine.Hertz.Spin()
	return nil
}

// ============= Context 方法 =============

// Next 执行下一个中间件
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort 终止执行
func (c *Context) Abort() {
	c.index = 63
}

// AbortWithStatus 终止并设置状态码
func (c *Context) AbortWithStatus(code int) {
	c.SetStatusCode(code)
	c.Abort()
}

// AbortWithStatusJSON 终止并返回JSON错误
func (c *Context) AbortWithStatusJSON(code int, jsonObj any) {
	c.Abort()
	c.JSON(code, jsonObj)
}

// IsAborted 检查是否已终止
func (c *Context) IsAborted() bool {
	return c.index >= 63
}

// Set 设置值
func (c *Context) Set(key string, value any) {
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

// Get 获取值
func (c *Context) Get(key string) (value any, exists bool) {
	value, exists = c.Keys[key]
	return
}

// MustGet 必须获取值
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString 获取字符串值
func (c *Context) GetString(key string) string {
	if val, ok := c.Get(key); ok && val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// Param 获取路径参数
func (c *Context) Param(key string) string {
	return c.Params.ByName(key)
}

// Query 获取查询参数
func (c *Context) Query(key string) string {
	value, _ := c.GetQuery(key)
	return value
}

// DefaultQuery 获取查询参数，带默认值
func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return defaultValue
}

// GetQuery 获取查询参数
func (c *Context) GetQuery(key string) (string, bool) {
	return string(c.RequestContext.Query(key)), c.RequestContext.QueryArgs().Has(key)
}

// GetHeader 获取请求头
func (c *Context) GetHeader(key string) string {
	return string(c.Request.Header.Peek(key))
}

// ============= 绑定方法 =============

// Bind 自动绑定
func (c *Context) Bind(obj any) error {
	b := binding.Default(string(c.Request.Method()), string(c.Request.Header.ContentType()))
	return c.MustBindWith(obj, b)
}

// BindJSON 绑定JSON
func (c *Context) BindJSON(obj any) error {
	return c.MustBindWith(obj, binding.JSON)
}

// ShouldBindJSON 应该绑定JSON
func (c *Context) ShouldBindJSON(obj any) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindQuery 应该绑定查询参数
func (c *Context) ShouldBindQuery(obj any) error {
	return c.ShouldBindWith(obj, binding.Query)
}

// ShouldBindUri 应该绑定URI参数
func (c *Context) ShouldBindUri(obj any) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.Uri.BindUri(m, obj)
}

// ShouldBind 应该绑定
func (c *Context) ShouldBind(obj any) error {
	b := binding.Default(string(c.Request.Method()), string(c.Request.Header.ContentType()))
	return c.ShouldBindWith(obj, b)
}

// ShouldBindWith 应该绑定（使用指定绑定器）
func (c *Context) ShouldBindWith(obj any, b binding.Binding) error {
	return b.Bind(c.RequestContext, obj)
}

// MustBindWith 必须绑定
func (c *Context) MustBindWith(obj any, b binding.Binding) error {
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return err
	}
	return nil
}

// AbortWithError 终止并添加错误
func (c *Context) AbortWithError(code int, err error) {
	c.AbortWithStatus(code)
	c.Errors = append(c.Errors, err)
}

// ============= 渲染方法 =============

// JSON 渲染JSON
func (c *Context) JSON(code int, obj any) {
	c.Render(code, render.JSON{Data: obj})
}

// String 渲染字符串
func (c *Context) String(code int, format string, values ...any) {
	c.Render(code, render.String{Format: format, Data: values})
}

// HTML 渲染HTML
func (c *Context) HTML(code int, name string, obj any) {
	c.SetStatusCode(code)
	c.SetContentType("text/html; charset=utf-8")
	c.SetBodyString(fmt.Sprintf("<html><body>HTML rendering: %s</body></html>", name))
}

// Data 渲染原始数据
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, render.Data{ContentType: contentType, Data: data})
}

// Render 使用渲染器渲染
func (c *Context) Render(code int, r render.Render) {
	c.SetStatusCode(code)

	if err := r.Render(c.RequestContext); err != nil {
		panic(err)
	}
}

// File 发送文件
func (c *Context) File(filepath string) {
	// 这里需要适配Hertz的文件发送
	c.SetBodyString("File: " + filepath)
}

// Header 设置响应头
func (c *Context) Header(key, value string) {
	c.Response.Header.Set(key, value)
}

// Status 设置状态码
func (c *Context) Status(code int) {
	c.SetStatusCode(code)
}

// ============= 辅助方法 =============

func (group *RouterGroup) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	finalSize := len(group.handlers) + len(handlers)
	mergedHandlers := make([]HandlerFunc, finalSize)
	copy(mergedHandlers, group.handlers)
	copy(mergedHandlers[len(group.handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.engine
	}
	return group
}

func (engine *Engine) createContext(c *app.RequestContext, handlers []HandlerFunc) *Context {
	ctx := &Context{
		RequestContext: c,
		handlers:       handlers,
		index:          -1,
		engine:         engine,
		Keys:           make(map[string]any),
		Errors:         make([]error, 0),
		Params:         make(Params, 0),
	}
	return ctx
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

func lastChar(str string) uint8 {
	if str == "" {
		return 0
	}
	return str[len(str)-1]
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

// ============= 内置中间件 =============

// Logger 日志中间件
func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		path := string(c.URI().Path())

		c.Next()

		latency := time.Since(start)
		statusCode := c.Response.StatusCode()
		method := string(c.Method())
		clientIP := c.ClientIP()

		fmt.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %s\n",
			start.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// Recovery 恢复中间件
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[Recovery] panic recovered: %v\n", err)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
