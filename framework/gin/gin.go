// Package gin 提供Gin风格的Web框架API
//
// 本包基于CloudWeGo Hertz引擎，提供完全兼容Gin框架的API接口，
// 解决了原生Gin与Hertz的context类型冲突问题，让开发者可以
// 无缝迁移现有Gin项目，同时享受Hertz的高性能优势。
//
// 主要特性：
//   - 完全兼容Gin API
//   - 基于高性能Hertz引擎
//   - 统一的Context类型
//   - 支持中间件链式调用
//   - 支持路由组和嵌套路由
//   - 多种数据绑定和渲染方式
//   - 内置常用中间件
//
// 使用示例：
//
//	r := gin.Default()
//	r.GET("/ping", func(c *gin.Context) {
//		c.JSON(200, gin.H{"message": "pong"})
//	})
//	r.Run(":8080")
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

// HandlerFunc 定义Gin风格的处理函数类型
//
// 处理函数接收一个Context参数，用于处理HTTP请求和响应。
// 这与原生Gin的HandlerFunc保持完全兼容。
//
// 示例：
//
//	func handler(c *gin.Context) {
//		c.JSON(200, gin.H{"status": "ok"})
//	}
type HandlerFunc func(*Context)

// Context 表示Gin风格的请求上下文
//
// Context封装了HTTP请求和响应的所有信息，提供了处理请求、
// 参数绑定、响应渲染等功能。它包装了Hertz的RequestContext，
// 同时添加了Gin兼容的API和中间件支持。
//
// 主要功能：
//   - 路径参数和查询参数获取
//   - 请求数据绑定（JSON、Query、URI等）
//   - 响应渲染（JSON、HTML、String等）
//   - 中间件执行控制
//   - 键值对存储
//   - 错误收集
type Context struct {
	*app.RequestContext // Hertz原生请求上下文

	// 参数存储
	Params Params         // 路径参数（如：/user/:id中的id）
	Keys   map[string]any // 用户自定义的键值对存储
	Errors []error        // 请求处理过程中收集的错误

	// 中间件控制
	handlers []HandlerFunc // 当前请求的处理函数链
	index    int8          // 当前执行的处理函数索引

	// 引擎引用
	engine *Engine // 所属的引擎实例
}

// Params 表示路径参数的集合
//
// Params是Param的切片，用于存储从URL路径中解析出的参数。
// 例如：路由"/user/:id/book/:title"会产生两个参数。
type Params []Param

// Param 表示单个路径参数
//
// 包含参数名和参数值。例如：
// 路由"/user/:id"，请求"/user/123"会产生Param{Key: "id", Value: "123"}
type Param struct {
	Key   string // 参数名（如："id"）
	Value string // 参数值（如："123"）
}

// Get 根据参数名获取参数值
//
// 返回参数值和是否存在的布尔值。如果参数不存在，返回空字符串和false。
//
// 参数：
//   - name: 参数名
//
// 返回：
//   - string: 参数值
//   - bool: 是否存在该参数
func (ps Params) Get(name string) (string, bool) {
	for _, p := range ps {
		if p.Key == name {
			return p.Value, true
		}
	}
	return "", false
}

// ByName 根据参数名获取参数值
//
// 这是Get方法的简化版本，只返回参数值，不返回是否存在的标志。
// 如果参数不存在，返回空字符串。
//
// 参数：
//   - name: 参数名
//
// 返回：
//   - string: 参数值，不存在时返回空字符串
func (ps Params) ByName(name string) string {
	va, _ := ps.Get(name)
	return va
}

// Engine 表示Gin风格的Web框架引擎
//
// Engine是整个Web应用的核心，负责路由注册、中间件管理、
// 请求分发等功能。它包装了Hertz服务器，同时提供Gin兼容的API。
//
// 主要功能：
//   - HTTP路由注册和管理
//   - 全局中间件管理
//   - 静态文件服务
//   - 404/405错误处理
//   - 服务器启动和配置
type Engine struct {
	*server.Hertz  // Hertz服务器实例
	RouterGroup     // 根路由组，提供路由注册功能
	middleware []HandlerFunc // 全局中间件（已废弃，使用RouterGroup.handlers）
	noRoute    []HandlerFunc // 404错误处理器
	noMethod   []HandlerFunc // 405错误处理器（方法不允许）
}

// RouterGroup 表示路由组
//
// RouterGroup用于组织相关的路由，支持嵌套和中间件。
// 所有路由注册方法都定义在此结构体上。
//
// 特性：
//   - 支持路径前缀
//   - 支持组级中间件
//   - 支持嵌套路由组
//   - 提供所有HTTP方法的路由注册
type RouterGroup struct {
	handlers []HandlerFunc // 当前组的中间件列表
	basePath string        // 路由组的基础路径前缀
	engine   *Engine       // 所属的引擎实例
	root     bool          // 是否为根路由组
}

// IRoutes 定义路由注册接口
//
// 此接口定义了所有路由注册相关的方法，Engine和RouterGroup都实现了此接口。
// 这样设计可以让路由组和引擎具有相同的路由注册能力。
type IRoutes interface {
	// 中间件管理
	Use(...HandlerFunc) IRoutes // 添加中间件

	// HTTP方法路由
	Handle(string, string, ...HandlerFunc) IRoutes // 注册指定HTTP方法的路由
	Any(string, ...HandlerFunc) IRoutes            // 注册所有HTTP方法的路由
	GET(string, ...HandlerFunc) IRoutes            // 注册GET路由
	POST(string, ...HandlerFunc) IRoutes           // 注册POST路由
	DELETE(string, ...HandlerFunc) IRoutes         // 注册DELETE路由
	PATCH(string, ...HandlerFunc) IRoutes          // 注册PATCH路由
	PUT(string, ...HandlerFunc) IRoutes            // 注册PUT路由
	OPTIONS(string, ...HandlerFunc) IRoutes        // 注册OPTIONS路由
	HEAD(string, ...HandlerFunc) IRoutes           // 注册HEAD路由

	// 静态文件服务
	StaticFile(string, string) IRoutes             // 注册单个静态文件
	Static(string, string) IRoutes                 // 注册静态文件目录
	StaticFS(string, http.FileSystem) IRoutes      // 注册文件系统
}

// New 创建一个新的Gin引擎实例
//
// 创建一个空的引擎，不包含任何中间件。如果需要默认中间件，
// 请使用Default()函数。
//
// 返回：
//   - *Engine: 新的引擎实例
//
// 示例：
//
//	r := gin.New()
//	r.Use(gin.Logger())
//	r.Use(gin.Recovery())
func New() *Engine {
	// 初始化Engine结构体
	engine := &Engine{
		RouterGroup: RouterGroup{
			handlers: nil,   // 初始时没有中间件
			basePath: "/",   // 根路径
			root:     true,  // 标记为根路由组
		},
	}
	// 设置根路由组的引擎引用
	engine.RouterGroup.engine = engine
	// 创建底层Hertz服务器实例
	engine.Hertz = server.Default()
	return engine
}

// Default 创建带有默认中间件的Gin引擎
//
// 创建一个包含Logger和Recovery中间件的引擎实例。
// 这是最常用的创建方式，适合大多数应用场景。
//
// 默认包含的中间件：
//   - Logger: 请求日志记录
//   - Recovery: panic恢复处理
//
// 返回：
//   - *Engine: 配置了默认中间件的引擎实例
//
// 示例：
//
//	r := gin.Default()
//	r.GET("/ping", func(c *gin.Context) {
//		c.String(200, "pong")
//	})
func Default() *Engine {
	// 创建基础引擎
	engine := New()
	// 添加默认中间件
	engine.Use(Logger(), Recovery())
	return engine
}

// Use 为路由组添加中间件
//
// 中间件会按照添加顺序执行，适用于当前路由组及其子组的所有路由。
// 中间件函数可以在请求处理前后执行逻辑，通过调用c.Next()来继续执行链。
//
// 参数：
//   - middleware: 一个或多个中间件函数
//
// 返回：
//   - IRoutes: 返回路由接口，支持链式调用
//
// 示例：
//
//	r.Use(gin.Logger())
//	r.Use(AuthMiddleware(), CorsMiddleware())
func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.handlers = append(group.handlers, middleware...)
	return group.returnObj()
}

// Group 创建一个新的路由组
//
// 路由组用于组织相关的路由，支持路径前缀和组级中间件。
// 新建的路由组会继承父组的中间件。
//
// 参数：
//   - relativePath: 相对于父组的路径前缀
//   - handlers: 可选的组级中间件
//
// 返回：
//   - *RouterGroup: 新的路由组实例
//
// 示例：
//
//	v1 := r.Group("/api/v1")
//	auth := v1.Group("/auth", AuthMiddleware())
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

// Handle 注册指定HTTP方法的路由
//
// 这是所有HTTP方法路由注册的底层实现。它将Gin风格的处理函数
// 转换为Hertz处理函数，并注册到底层的Hertz引擎。
//
// 参数：
//   - httpMethod: HTTP方法（GET、POST、PUT等）
//   - relativePath: 相对路径
//   - handlers: 处理函数列表（包括中间件和最终处理器）
//
// 返回：
//   - IRoutes: 返回路由接口，支持链式调用
//
// 示例：
//
//	r.Handle("GET", "/users", middleware1, middleware2, handler)
func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes {
	// 计算绝对路径（组合父组路径和当前相对路径）
	absolutePath := group.calculateAbsolutePath(relativePath)
	// 合并组中间件和当前处理函数
	finalHandlers := group.combineHandlers(handlers)

	// 转换为Hertz处理函数
	// 这里是关键的适配层，将Gin风格的处理函数转换为Hertz可以理解的格式
	hertzHandler := func(ctx context.Context, req *app.RequestContext) {
		// 创建Gin Context并执行处理函数链
		ginCtx := group.engine.createContext(req, finalHandlers)
		ginCtx.Next() // 开始执行中间件和处理函数链
	}

	// 注册到底层Hertz路由系统
	// 根据HTTP方法选择对应的Hertz注册函数
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

// Next 执行下一个中间件或处理函数
//
// 在中间件中调用此方法可以继续执行后续的中间件和处理函数。
// 如果不调用Next()，处理链将在当前中间件停止。
//
// 示例：
//
//	func middleware(c *gin.Context) {
//		// 前置处理
//		log.Println("Before request")
//		c.Next() // 继续执行
//		// 后置处理
//		log.Println("After request")
//	}
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort 终止中间件链的执行
//
// 调用此方法会停止执行后续的中间件和处理函数。
// 但不会停止当前函数的执行。
//
// 示例：
//
//	func authMiddleware(c *gin.Context) {
//		if !isAuthenticated(c) {
//			c.Abort()
//			return
//		}
//		c.Next()
//	}
func (c *Context) Abort() {
	c.index = 63 // 设置为最大值以停止执行
}

// AbortWithStatus 终止执行并设置响应状态码
//
// 这是Abort()和Status()的组合方法，便于在一个调用中同时设置状态码和终止执行。
//
// 参数：
//   - code: HTTP状态码
//
// 示例：
//
//	if !authorized {
//		c.AbortWithStatus(401)
//		return
//	}
func (c *Context) AbortWithStatus(code int) {
	c.SetStatusCode(code)
	c.Abort()
}

// AbortWithStatusJSON 终止执行并返回JSON响应
//
// 这个方法结合了Abort()和JSON()，常用于在中间件中返回错误响应。
//
// 参数：
//   - code: HTTP状态码
//   - jsonObj: 要序列化为JSON的对象
//
// 示例：
//
//	if err != nil {
//		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
//		return
//	}
func (c *Context) AbortWithStatusJSON(code int, jsonObj any) {
	c.Abort()
	c.JSON(code, jsonObj)
}

// IsAborted 检查中间件链是否已被终止
//
// 返回当前请求的中间件链是否已经被Abort()终止。
//
// 返回：
//   - bool: true表示已终止，false表示正常
//
// 示例：
//
//	if c.IsAborted() {
//		return // 不继续处理
//	}
func (c *Context) IsAborted() bool {
	return c.index >= 63
}

// Set 在上下文中存储键值对
//
// 用于在中间件和处理函数之间传递数据。存储的数据只在当前请求的生命周期内有效。
//
// 参数：
//   - key: 键名
//   - value: 任意类型的值
//
// 示例：
//
//	c.Set("user_id", 12345)
//	c.Set("username", "john")
//	c.Set("start_time", time.Now())
func (c *Context) Set(key string, value any) {
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

// Get 从上下文中获取键值对
//
// 返回指定键的值和是否存在的标志。
//
// 参数：
//   - key: 键名
//
// 返回：
//   - value: 键对应的值（any类型）
//   - exists: 是否存在该键
//
// 示例：
//
//	if userID, exists := c.Get("user_id"); exists {
//		// 使用userID
//	}
func (c *Context) Get(key string) (value any, exists bool) {
	value, exists = c.Keys[key]
	return
}

// MustGet 从上下文中获取值（不存在时panic）
//
// 这是Get()的严格版本，如果键不存在会直接panic。
// 适用于你确信某个键一定存在的情况。
//
// 参数：
//   - key: 键名
//
// 返回：
//   - any: 键对应的值
//
// Panic：
//   - 当指定的键不存在时
//
// 示例：
//
//	userID := c.MustGet("user_id").(int)
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString 从上下文中获取字符串类型的值
//
// 这是一个类型安全的方法，如果键不存在或值不是字符串类型，
// 将返回空字符串。
//
// 参数：
//   - key: 键名
//
// 返回：
//   - string: 字符串值，不存在或类型不匹配时返回空字符串
//
// 示例：
//
//	username := c.GetString("username") // 安全获取字符串
func (c *Context) GetString(key string) string {
	if val, ok := c.Get(key); ok && val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// Param 获取URL路径参数
//
// 从路由定义中获取命名参数的值。例如，路由“/user/:id”中的“id”参数。
//
// 参数：
//   - key: 参数名（不包含冒号）
//
// 返回：
//   - string: 参数值，不存在时返回空字符串
//
// 示例：
//
//	// 路由：/user/:id
//	// 请求：/user/123
//	userID := c.Param("id") // 返回 "123"
func (c *Context) Param(key string) string {
	return c.Params.ByName(key)
}

// Query 获取URL查询参数
//
// 从URL查询字符串中获取指定参数的值。这是GetQuery()的简化版本。
//
// 参数：
//   - key: 查询参数名
//
// 返回：
//   - string: 参数值，不存在时返回空字符串
//
// 示例：
//
//	// URL：/search?q=golang&page=1
//	query := c.Query("q")    // 返回 "golang"
//	page := c.Query("page")  // 返回 "1"
func (c *Context) Query(key string) string {
	value, _ := c.GetQuery(key)
	return value
}

// DefaultQuery 获取URL查询参数，带默认值
//
// 当指定的查询参数不存在时，返回提供的默认值。
//
// 参数：
//   - key: 查询参数名
//   - defaultValue: 默认值
//
// 返回：
//   - string: 参数值或默认值
//
// 示例：
//
//	// URL：/search?q=golang
//	query := c.DefaultQuery("q", "default")     // 返回 "golang"
//	page := c.DefaultQuery("page", "1")        // 返回 "1" (默认值)
func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return defaultValue
}

// GetQuery 获取URL查询参数和存在标志
//
// 返回指定查询参数的值和是否存在的标志。
//
// 参数：
//   - key: 查询参数名
//
// 返回：
//   - string: 参数值
//   - bool: 是否存在该参数
//
// 示例：
//
//	if value, exists := c.GetQuery("optional"); exists {
//		// 处理可选参数
//	}
func (c *Context) GetQuery(key string) (string, bool) {
	return string(c.RequestContext.Query(key)), c.RequestContext.QueryArgs().Has(key)
}

// GetHeader 获取HTTP请求头
//
// 获取指定HTTP头的值。头名不区分大小写。
//
// 参数：
//   - key: HTTP头名
//
// 返回：
//   - string: 头值，不存在时返回空字符串
//
// 示例：
//
//	userAgent := c.GetHeader("User-Agent")
//	auth := c.GetHeader("Authorization")
//	contentType := c.GetHeader("Content-Type")
func (c *Context) GetHeader(key string) string {
	return string(c.Request.Header.Peek(key))
}

// ============= 绑定方法 =============

// Bind 自动选择绑定方式并绑定请求数据
//
// 根据HTTP方法和Content-Type自动选择适合的绑定器。
// 如果绑定失败，会自动设置400状态码并终止请求。
//
// 参数：
//   - obj: 要绑定数据的目标结构体指针
//
// 返回：
//   - error: 绑定错误，成功时为nil
//
// 示例：
//
//	var user User
//	if err := c.Bind(&user); err != nil {
//		// 处理错误（已自动设置400状态码）
//		return
//	}
func (c *Context) Bind(obj any) error {
	b := binding.Default(string(c.Request.Method()), string(c.Request.Header.ContentType()))
	return c.MustBindWith(obj, b)
}

// BindJSON 绑定JSON请求数据
//
// 将请求体中的JSON数据绑定到指定的结构体。
// 如果绑定失败，会自动设置400状态码并终止请求。
//
// 参数：
//   - obj: 要绑定数据的目标结构体指针
//
// 返回：
//   - error: 绑定错误，成功时为nil
//
// 示例：
//
//	var user User
//	if err := c.BindJSON(&user); err != nil {
//		return // 已自动处理错误
//	}
func (c *Context) BindJSON(obj any) error {
	return c.MustBindWith(obj, binding.JSON)
}

// ShouldBindJSON 尝试绑定JSON请求数据
//
// 与BindJSON类似，但不会自动设置错误状态码。
// 需要手动处理绑定错误。
//
// 参数：
//   - obj: 要绑定数据的目标结构体指针
//
// 返回：
//   - error: 绑定错误，成功时为nil
//
// 示例：
//
//	var user User
//	if err := c.ShouldBindJSON(&user); err != nil {
//		c.JSON(400, gin.H{"error": err.Error()})
//		return
//	}
func (c *Context) ShouldBindJSON(obj any) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindQuery 尝试绑定URL查询参数
//
// 将URL查询参数绑定到指定的结构体。不会自动设置错误状态码。
//
// 参数：
//   - obj: 要绑定数据的目标结构体指针，使用`form`标签
//
// 返回：
//   - error: 绑定错误，成功时为nil
//
// 示例：
//
//	type Query struct {
//		Page int `form:"page"`
//		Size int `form:"size"`
//	}
//	var q Query
//	if err := c.ShouldBindQuery(&q); err != nil {
//		c.JSON(400, gin.H{"error": err.Error()})
//	}
func (c *Context) ShouldBindQuery(obj any) error {
	return c.ShouldBindWith(obj, binding.Query)
}

// ShouldBindUri 尝试绑定URI路径参数
//
// 将URL路径中的命名参数绑定到指定的结构体。不会自动设置错误状态码。
//
// 参数：
//   - obj: 要绑定数据的目标结构体指针，使用`uri`标签
//
// 返回：
//   - error: 绑定错误，成功时为nil
//
// 示例：
//
//	// 路由：/user/:id/book/:title
//	type URI struct {
//		ID    int    `uri:"id"`
//		Title string `uri:"title"`
//	}
//	var uri URI
//	if err := c.ShouldBindUri(&uri); err != nil {
//		c.JSON(400, gin.H{"error": err.Error()})
//	}
func (c *Context) ShouldBindUri(obj any) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.Uri.BindUri(m, obj)
}

// ShouldBind 尝试自动选择绑定方式并绑定请求数据
//
// 根据HTTP方法和Content-Type自动选择适合的绑定器。
// 与Bind()类似，但不会自动设置错误状态码。
//
// 参数：
//   - obj: 要绑定数据的目标结构体指针
//
// 返回：
//   - error: 绑定错误，成功时为nil
//
// 示例：
//
//	var data interface{}
//	if err := c.ShouldBind(&data); err != nil {
//		c.JSON(400, gin.H{"error": err.Error()})
//		return
//	}
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

// combineHandlers 合并路由组中间件和当前处理函数
//
// 将路由组的中间件和新添加的处理函数合并为一个完整的处理函数链。
// 中间件会在处理函数之前执行。
func (group *RouterGroup) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	// 计算合并后的总长度
	finalSize := len(group.handlers) + len(handlers)
	// 创建新的切片存储合并结果
	mergedHandlers := make([]HandlerFunc, finalSize)
	// 先复制组中间件
	copy(mergedHandlers, group.handlers)
	// 再复制新添加的处理函数
	copy(mergedHandlers[len(group.handlers):], handlers)
	return mergedHandlers
}

// calculateAbsolutePath 计算绝对路径
//
// 将路由组的基础路径和相对路径组合成绝对路径。
func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

// returnObj 返回适当的路由接口实例
//
// 如果是根路由组，返回Engine实例；否则返回路由组实例。
// 这样设计可以实现链式调用。
func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.engine
	}
	return group
}

// createContext 创建Gin Context实例
//
// 将Hertz的RequestContext包装成Gin兼容的Context，
// 并初始化中间件处理状态。
func (engine *Engine) createContext(c *app.RequestContext, handlers []HandlerFunc) *Context {
	// 创建Gin Context实例
	ctx := &Context{
		RequestContext: c,                    // Hertz原生上下文
		handlers:       handlers,              // 处理函数链
		index:          -1,                    // 初始索引（下一个执行的是0）
		engine:         engine,                // 引擎引用
		Keys:           make(map[string]any),  // 用户数据存储
		Errors:         make([]error, 0),      // 错误收集
		Params:         make(Params, 0),       // 路径参数
	}
	return ctx
}

// joinPaths 连接路径段
//
// 将绝对路径和相对路径正确地连接起来，处理尾部斜杠。
func joinPaths(absolutePath, relativePath string) string {
	// 如果相对路径为空，直接返回绝对路径
	if relativePath == "" {
		return absolutePath
	}

	// 使用path.Join连接路径
	finalPath := path.Join(absolutePath, relativePath)
	// 如果原相对路径以斜杠结尾，保持这个特征
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

// lastChar 获取字符串的最后一个字符
//
// 返回字符串的最后一个字节，空字符串返回0。
func lastChar(str string) uint8 {
	if str == "" {
		return 0
	}
	return str[len(str)-1]
}

// resolveAddress 解析服务器监听地址
//
// 根据提供的参数解析出最终的监听地址。
// 无参数时默认使用:8080，一个参数时使用指定地址，
// 多个参数时panic。
func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		return ":8080"    // 默认端口
	case 1:
		return addr[0]    // 使用指定地址
	default:
		panic("too many parameters") // 不允许多个地址
	}
}

// ============= 内置中间件 =============

// Logger 返回一个请求日志中间件
//
// 记录每个请求的详细信息，包括：
//   - 请求时间
//   - HTTP状态码
//   - 处理耗时
//   - 客户端IP
//   - HTTP方法
//   - 请求路径
//
// 返回：
//   - HandlerFunc: 日志中间件函数
//
// 示例：
//
//	r := gin.New()
//	r.Use(gin.Logger()) // 添加日志中间件
//
// 输出格式：
//
//	[GIN] 2023/01/01 - 12:00:00 | 200 |    1.234567ms |  192.168.1.1 | GET     /api/users
func Logger() HandlerFunc {
	return func(c *Context) {
		// 记录请求开始时间
		start := time.Now()
		path := string(c.URI().Path())

		// 执行后续中间件和处理函数
		c.Next()

		// 计算处理耗时和收集响应信息
		latency := time.Since(start)
		statusCode := c.Response.StatusCode()
		method := string(c.Method())
		clientIP := c.ClientIP()

		// 格式化输出日志
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

// Recovery 返回一个panic恢复中间件
//
// 当处理函数中发生panic时，此中间件会捕获panic并恢复，
// 防止整个服务器崩溃。会自动返回500内部服务器错误。
//
// 功能：
//   - 捕获和恢复panic
//   - 记录panic信息
//   - 返回500状态码
//   - 保持服务器稳定运行
//
// 返回：
//   - HandlerFunc: 恢复中间件函数
//
// 示例：
//
//	r := gin.New()
//	r.Use(gin.Recovery()) // 添加panic恢复中间件
func Recovery() HandlerFunc {
	return func(c *Context) {
		// 使用defer捕获panic
		defer func() {
			if err := recover(); err != nil {
				// 记录panic信息
				fmt.Printf("[Recovery] panic recovered: %v\n", err)
				// 返回500错误并终止请求处理
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		// 执行后续中间件和处理函数
		c.Next()
	}
}
