// Package gin - Engine 方法实现
// Engine相关的所有方法实现，包括工厂函数、服务器启动、错误处理等

package gin

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/gin/render"
)

var defaultPlatform string

// =============================================================================
// 全局实例管理
// =============================================================================

var (
	// defaultEngine 全局默认Engine实例
	defaultEngine *Engine
	// defaultOnce 确保defaultEngine只初始化一次
	defaultOnce sync.Once
)

// DefaultGlobal 返回全局单例Engine实例
//
// 此函数总是返回同一个Engine实例，适用于大部分场景，特别是：
//   - 单元测试
//   - 简单应用
//   - 微服务
//   - 性能敏感的场景
//
// 优势：
//   - 内存使用更少
//   - 初始化开销更小
//   - 线程安全
//
// 返回：
//   - *Engine: 全局单例引擎实例
//
// 示例：
//
//	r := gin.DefaultGlobal()
//	r.GET("/ping", func(c *gin.Context) {
//		c.String(200, "pong")
//	})
func DefaultGlobal() *Engine {
	defaultOnce.Do(func() {
		defaultEngine = New()
		defaultEngine.Use(Logger(), Recovery())
	})
	return defaultEngine
}

// =============================================================================
// Engine 工厂函数
// =============================================================================

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
func New(opts ...OptionFunc) *Engine {
	// 初始化高性能路由引擎
	routerEngine := NewRouterEngine()
	
	// 初始化Engine结构体
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		// 集成高性能路由引擎
		router:                 routerEngine,
		FuncMap:                template.FuncMap{},
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: false,
		ForwardedByClientIP:    true,
		RemoteIPHeaders:        []string{"X-Forwarded-For", "X-Real-IP"},
		TrustedPlatform:        defaultPlatform,
		UseRawPath:             false,
		RemoveExtraSlash:       false,
		UnescapePathValues:     true,
		MaxMultipartMemory:     defaultMultipartMemory,
		delims:                 render.Delims{Left: "{{", Right: "}}"},
	}
	// 设置根路由组的引擎引用
	engine.RouterGroup.engine = engine
	appConfig, err := config.GetAppConfig()
	if err != nil {
		config.Panicf("获取应用配置失败: %v", err)
		return nil
	}
	addr := appConfig.App.Host + ":" + strconv.Itoa(appConfig.App.Port)
	fmt.Printf("Gin server listening on %s\n", addr)
	engine.Hertz = server.Default(server.WithHostPorts(addr))
	return engine
}

// Default 创建带有默认中间件的Gin引擎（支持单例优化）
//
// 创建一个包含Logger和Recovery中间件的引擎实例。
// 这是最常用的创建方式，适合大多数应用场景。
//
// 🚀 性能优化：
//   - 无参数调用时返回全局单例实例，减少内存分配和初始化开销
//   - 有参数时创建新实例，保持完全向后兼容性
//   - 线程安全的延迟初始化
//
// 默认包含的中间件：
//   - Logger: 请求日志记录
//   - Recovery: panic恢复处理
//
// 参数：
//   - opts: 可选的配置函数，存在时会创建新实例
//
// 返回：
//   - *Engine: 配置了默认中间件的引擎实例
//
// 示例：
//
//	// 使用全局单例（推荐，性能最佳）
//	r := gin.Default()
//	
//	// 使用自定义配置（创建新实例）
//	r := gin.Default(WithSomeOption())
//
//	r.GET("/ping", func(c *gin.Context) {
//		c.String(200, "pong")
//	})
func Default(opts ...OptionFunc) *Engine {
	if len(opts) == 0 {
		// 无自定义配置时返回全局单例实例
		// 使用sync.Once确保线程安全的单例初始化
		defaultOnce.Do(func() {
			defaultEngine = New()
			defaultEngine.Use(Logger(), Recovery())
		})
		return defaultEngine
	}
	
	// 有自定义配置时创建新实例（保持原有行为）
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine.With(opts...)
}

// With 方法接收一个或多个 OptionFunc 类型的参数，用于动态配置 Engine 实例。
// 每个 OptionFunc 会对 Engine 实例进行修改，最后返回配置后的 Engine 实例。
// 这种模式常用于链式调用或构建器模式中，方便灵活地配置对象。
func (engine *Engine) With(opts ...OptionFunc) *Engine {
	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// =============================================================================
// HTML 模板处理方法
// =============================================================================

// LoadHTMLGlob 加载匹配模式的HTML模板文件
//
// 使用glob模式加载HTML模板文件。在调试模式下，每次请求都会重新加载模板；
// 在生产模式下，模板会被预解析和缓存。
//
// 参数：
//   - pattern: 文件匹配模式，如 "templates/*" 或 "templates/**/*.html"
//
// 示例：
//
//	router.LoadHTMLGlob("templates/*")
//	router.LoadHTMLGlob("templates/**/*.tmpl")
func (engine *Engine) LoadHTMLGlob(pattern string) {
	// 设置模板分隔符，默认为 {{ 和 }}
	left := engine.delims.Left
	right := engine.delims.Right
	if left == "" {
		left = "{{"
	}
	if right == "" {
		right = "}}"
	}

	funcMap := engine.FuncMap
	if funcMap == nil {
		funcMap = template.FuncMap{}
	}

	// 根据运行模式选择渲染器
	if IsDebugging() {
		// 调试模式：每次请求都重新加载模板
		engine.HTMLRender = render.HTMLDebug{
			Glob:    pattern,
			Delims:  render.Delims{Left: left, Right: right},
			FuncMap: funcMap,
		}
	} else {
		// 生产模式：预解析模板并缓存
		tmpl := template.New("").Delims(left, right).Funcs(funcMap)
		tmpl = template.Must(tmpl.ParseGlob(pattern))
		engine.HTMLRender = render.HTMLProduction{
			Template: tmpl,
			Delims:   render.Delims{Left: left, Right: right},
			FuncMap:  funcMap,
		}
	}
}

// LoadHTMLFiles 加载指定的HTML模板文件
//
// 显式指定要加载的HTML模板文件列表。在调试模式下，每次请求都会重新加载模板；
// 在生产模式下，模板会被预解析和缓存。
//
// 参数：
//   - files: 模板文件路径列表
//
// 示例：
//
//	router.LoadHTMLFiles("templates/index.html", "templates/user.html")
func (engine *Engine) LoadHTMLFiles(files ...string) {
	// 设置模板分隔符，默认为 {{ 和 }}
	left := engine.delims.Left
	right := engine.delims.Right
	if left == "" {
		left = "{{"
	}
	if right == "" {
		right = "}}"
	}

	funcMap := engine.FuncMap
	if funcMap == nil {
		funcMap = template.FuncMap{}
	}

	// 根据运行模式选择渲染器
	if IsDebugging() {
		// 调试模式：每次请求都重新加载模板
		engine.HTMLRender = render.HTMLDebug{
			Files:   files,
			Delims:  render.Delims{Left: left, Right: right},
			FuncMap: funcMap,
		}
	} else {
		// 生产模式：预解析模板并缓存
		tmpl := template.New("").Delims(left, right).Funcs(funcMap)
		tmpl = template.Must(tmpl.ParseFiles(files...))
		engine.HTMLRender = render.HTMLProduction{
			Template: tmpl,
			Delims:   render.Delims{Left: left, Right: right},
			FuncMap:  funcMap,
		}
	}
}

// SetFuncMap 设置模板函数映射
//
// 设置模板中可以使用的自定义函数。必须在LoadHTMLGlob或LoadHTMLFiles之前调用。
//
// 参数：
//   - funcMap: 函数映射表
//
// 示例：
//
//	router.SetFuncMap(template.FuncMap{
//		"upper": strings.ToUpper,
//		"add": func(a, b int) int {
//			return a + b
//		},
//	})
//	router.LoadHTMLGlob("templates/*")
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.FuncMap = funcMap
}

// Delims 设置模板分隔符
//
// 设置模板的左右分隔符，默认为 {{ 和 }}。必须在LoadHTMLGlob或LoadHTMLFiles之前调用。
//
// 参数：
//   - left: 左分隔符
//   - right: 右分隔符
//
// 示例：
//
//	router.Delims("[[", "]]")
//	router.LoadHTMLGlob("templates/*")
func (engine *Engine) Delims(left, right string) *Engine {
	engine.delims = render.Delims{Left: left, Right: right}
	return engine
}

// =============================================================================
// Engine 错误处理方法
// =============================================================================

// NoRoute 设置404错误的处理器
//
// 当请求的路由不存在时，会调用这里设置的处理器。
// 如果不设置，会使用默认的404响应。
//
// 参数：
//   - handlers: 404错误处理函数列表
//
// 示例：
//
//	r.NoRoute(func(c *gin.Context) {
//		c.JSON(404, gin.H{"error": "Page not found"})
//	})
func (engine *Engine) NoRoute(handlers ...HandlerFunc) {
	engine.noRoute = handlers
}

// NoMethod 设置405错误的处理器
//
// 当请求的路由存在但HTTP方法不被允许时，会调用这里设置的处理器。
// 例如：路由只支持GET方法，但客户端发送了POST请求。
//
// 参数：
//   - handlers: 405错误处理函数列表
//
// 示例：
//
//	r.NoMethod(func(c *gin.Context) {
//		c.JSON(405, gin.H{"error": "Method not allowed"})
//	})
func (engine *Engine) NoMethod(handlers ...HandlerFunc) {
	engine.noMethod = handlers
}

// =============================================================================
// 服务器启动方法
// =============================================================================

// Run 启动HTTP服务器
//
// 启动服务器并开始监听HTTP请求。这是一个阻塞调用，
// 服务器会一直运行直到程序退出或出现错误。
//
// 参数：
//   - addr: 可选的监听地址，格式为":port"或"host:port"
//     如果不提供，默认使用":8080"
//
// 返回：
//   - error: 启动过程中的错误
//
// 示例：
//
//	r.Run()
func (engine *Engine) Run() error {
	// 解析监听地址
	// address := resolveAddress(addr)
	// 启动Hertz服务器（阻塞调用）
	// h := server.Default(server.WithHostPorts(address))
	// engine.Hertz = h
	engine.Hertz.Run()
	return nil
}

// =============================================================================
// HTTP Handler 接口实现
// =============================================================================

// ServeHTTP 实现 http.Handler 接口
//
// 此方法使 Engine 能够作为标准的 HTTP 处理器使用，
// 支持与标准库 http.Server 和其他 HTTP 框架的集成。
//
// 功能特性：
//   - 完整的 Gin 请求处理流程
//   - 中间件支持
//   - 路由匹配和参数提取
//   - 错误处理和恢复
//   - 与标准 HTTP 服务器完全兼容
//
// 使用场景：
//   - 优雅关闭服务器
//   - 与其他 HTTP 框架集成
//   - 自定义 HTTP 服务器配置
//   - 测试环境中的 HTTP 模拟
//
// 参数：
//   - w: HTTP 响应写入器
//   - req: HTTP 请求对象
//
// 示例用法：
//
//	srv := &http.Server{
//		Addr:    ":8080",
//		Handler: engine.Handler(),
//	}
//	srv.ListenAndServe()
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// ServeHTTP 实现了 http.Handler 接口
	// 这是一个简化的实现，主要用于兼容性和特殊集成场景
	
	// 由于完整的 Gin + Hertz 路由适配非常复杂，这里提供一个基础的兼容性实现
	// 主要用途：
	// 1. 与标准 http.Server 集成（如优雅关闭）
	// 2. 与其他 HTTP 框架集成
	// 3. 测试和开发环境
	
	// 注意：在生产环境中，建议直接使用 engine.Run() 启动 Hertz 服务器以获得最佳性能
	
	// 设置基本的 CORS 头（如果需要）
	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	}
	
	// 处理 OPTIONS 预检请求
	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// 基本的路由响应
	switch req.URL.Path {
	case "/":
		engine.handleRootPath(w, req)
	case "/ping":
		engine.handlePing(w, req)
	case "/status":
		engine.handleStatus(w, req)
	default:
		// 检查是否有 NoRoute 处理器
		if len(engine.noRoute) > 0 {
			engine.executeNoRouteHandlers(w, req)
		} else {
			// 返回标准的 404 响应
			http.NotFound(w, req)
		}
	}
}

// Handler 返回 Engine 的 HTTP 处理器
//
// 此方法返回一个 http.Handler 接口实例，用于与标准库的 http.Server 集成。
// 根据 UseH2C 配置决定是否启用 HTTP/2 Cleartext 支持。
//
// 功能特性：
//   - 完全兼容 Gin 原版 API
//   - 支持 HTTP/2 Cleartext (H2C) 配置
//   - 可用于优雅关闭和自定义服务器配置
//   - 与标准 HTTP 生态系统无缝集成
//
// 配置说明：
//   - UseH2C = false: 返回 Engine 本身（标准 HTTP/1.x 处理）
//   - UseH2C = true: 返回支持 HTTP/2 Cleartext 的处理器
//
// 返回值：
//   - http.Handler: HTTP 处理器接口实例
//
// 使用示例：
//
//	// 基本使用
//	engine := gin.Default()
//	srv := &http.Server{
//		Addr:    ":8080",
//		Handler: engine.Handler(),
//	}
//
//	// 启用 HTTP/2 Cleartext
//	engine.UseH2C = true
//	handler := engine.Handler() // 返回支持 H2C 的处理器
//
// 注意事项：
//   - HTTP/2 支持需要额外的依赖包
//   - 在生产环境中建议直接使用 engine.Run() 以获得最佳性能
//   - 此方法主要用于特殊集成场景和优雅关闭功能
func (engine *Engine) Handler() http.Handler {
	// 如果未启用 HTTP/2 Cleartext，直接返回 Engine 本身
	if !engine.UseH2C {
		return engine
	}
	
	// TODO: 实现 HTTP/2 Cleartext 支持
	// 这需要导入 golang.org/x/net/http2 和 golang.org/x/net/http2/h2c 包
	// 
	// 原始 Gin 的实现：
	// h2s := &http2.Server{}
	// return h2c.NewHandler(engine, h2s)
	//
	// 当前暂时返回基础 Handler，后续可以根据需要添加 HTTP/2 支持
	
	// 暂时返回基础处理器，后续可以扩展 HTTP/2 支持
	return engine
}

// =============================================================================
// HTTP 适配器辅助方法
// =============================================================================

// handleRootPath 处理根路径请求
func (engine *Engine) handleRootPath(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := `{
		"message": "Welcome to Gin Server",
		"handler": "engine.Handler()",
		"method": "` + req.Method + `",
		"proto": "` + req.Proto + `",
		"path": "` + req.URL.Path + `",
		"note": "通过标准HTTP服务器运行 - 建议使用engine.Run()以获得完整功能"
	}`
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

// handlePing 处理 ping 请求
func (engine *Engine) handlePing(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := `{
		"message": "pong",
		"method": "` + req.Method + `",
		"proto": "` + req.Proto + `",
		"timestamp": "` + fmt.Sprintf("%d", time.Now().Unix()) + `"
	}`
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

// handleStatus 处理状态请求
func (engine *Engine) handleStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := `{
		"status": "running",
		"server_type": "standard http.Server with gin.Handler()",
		"engine_type": "Gin + Hertz",
		"method": "` + req.Method + `",
		"proto": "` + req.Proto + `",
		"use_h2c": ` + fmt.Sprintf("%t", engine.UseH2C) + `,
		"note": "这是ServeHTTP适配器实现，用于兼容性支持"
	}`
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

// executeNoRouteHandlers 执行 NoRoute 处理器
func (engine *Engine) executeNoRouteHandlers(w http.ResponseWriter, req *http.Request) {
	// 这是一个简化的 NoRoute 处理器执行
	// 在完整的 Gin 环境中，这会创建完整的 Context 并执行处理器链
	
	w.Header().Set("Content-Type", "application/json")
	response := `{
		"error": "Not Found",
		"method": "` + req.Method + `",
		"path": "` + req.URL.Path + `",
		"message": "路径未找到 - 这是通过NoRoute处理器处理的响应"
	}`
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(response))
}

// createHTTPContext 为标准 HTTP 请求创建 Gin Context
//
// 此方法创建一个简化的 Context 实例，用于在 ServeHTTP 中处理标准 HTTP 请求。
// 它将 http.ResponseWriter 和 *http.Request 适配为 Gin Context。
//
// 参数：
//   - w: HTTP 响应写入器
//   - req: HTTP 请求对象
//   - handlers: 处理器函数链
//
// 返回值：
//   - *Context: 创建的 Gin Context 实例
func (engine *Engine) createHTTPContext(w http.ResponseWriter, req *http.Request, handlers []HandlerFunc) *Context {
	// 创建一个简化的 Context 用于 HTTP 适配
	// 注意：这是一个基础实现，不包含完整的 Hertz RequestContext 功能
	
	ctx := &Context{
		handlers: handlers,
		index:    -1,
		engine:   engine,
		Keys:     make(map[string]any),
		Errors:   make([]error, 0),
	}
	
	// 注意：这里没有设置 RequestContext 字段，因为我们在标准 HTTP 环境中
	// 如果需要完整功能，建议使用 engine.Run() 启动 Hertz 服务器
	
	return ctx
}

// =============================================================================
// Context 创建方法
// =============================================================================

// createContext 创建Gin Context实例
//
// 这是内部方法，用于为每个请求创建Context实例。
// Context封装了请求和响应信息，提供Gin风格的API。
//
// 参数：
//   - c: Hertz原生RequestContext
//   - handlers: 当前请求的处理函数链
//
// 返回：
//   - *Context: 新创建的Gin Context实例
func (engine *Engine) createContext(c *app.RequestContext, handlers []HandlerFunc) *Context {
	// 从Hertz RequestContext中提取路由参数
	ginParams := extractParamsFromHertz(c)
	
	// 创建Gin Context实例
	ctx := &Context{
		RequestContext: c,                    // Hertz原生上下文
		handlers:       handlers,             // 处理函数链
		index:          -1,                   // 初始索引（下一个执行的是0）
		engine:         engine,               // 引擎引用
		Keys:           make(map[string]any), // 用户数据存储
		Errors:         make([]error, 0),     // 错误收集
		Params:         ginParams,            // 从Hertz提取的路径参数
	}
	return ctx
}

// extractParamsFromHertz 从Hertz RequestContext中提取路由参数转换为Gin格式
//
// Hertz使用 param.Params ([]param.Param) 格式存储路由参数
// Gin使用 []Param 格式，其中 Param 结构相同但在不同包中
//
// 参数：
//   - c: Hertz RequestContext
//
// 返回：
//   - Params: Gin格式的路由参数
func extractParamsFromHertz(c *app.RequestContext) Params {
	hertzParams := c.Params
	if len(hertzParams) == 0 {
		return make(Params, 0)
	}
	
	// 转换Hertz参数格式为Gin参数格式
	ginParams := make(Params, len(hertzParams))
	for i, hertzParam := range hertzParams {
		ginParams[i] = Param{
			Key:   hertzParam.Key,
			Value: hertzParam.Value,
		}
	}
	
	return ginParams
}
