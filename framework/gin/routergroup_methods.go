// Package gin - RouterGroup 方法实现
// RouterGroup相关的所有方法实现，包括路由注册、中间件管理、静态文件服务等

package gin

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// =============================================================================
// 中间件管理方法
// =============================================================================

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
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

// =============================================================================
// 路由组管理方法
// =============================================================================

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
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

// =============================================================================
// HTTP方法路由注册
// =============================================================================

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
//
// GET方法通常用于获取资源数据。
//
// 参数：
//   - relativePath: 相对路径
//   - handlers: 处理函数列表
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.GET("/users", getUsersHandler)
//	r.GET("/users/:id", getUserHandler)
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("GET", relativePath, handlers...)
}

// POST 注册POST路由
//
// POST方法通常用于创建新资源。
//
// 参数：
//   - relativePath: 相对路径
//   - handlers: 处理函数列表
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.POST("/users", createUserHandler)
func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("POST", relativePath, handlers...)
}

// DELETE 注册DELETE路由
//
// DELETE方法通常用于删除资源。
//
// 参数：
//   - relativePath: 相对路径
//   - handlers: 处理函数列表
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.DELETE("/users/:id", deleteUserHandler)
func (group *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("DELETE", relativePath, handlers...)
}

// PATCH 注册PATCH路由
//
// PATCH方法通常用于部分更新资源。
//
// 参数：
//   - relativePath: 相对路径
//   - handlers: 处理函数列表
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.PATCH("/users/:id", updateUserHandler)
func (group *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("PATCH", relativePath, handlers...)
}

// PUT 注册PUT路由
//
// PUT方法通常用于完整更新资源。
//
// 参数：
//   - relativePath: 相对路径
//   - handlers: 处理函数列表
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.PUT("/users/:id", replaceUserHandler)
func (group *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("PUT", relativePath, handlers...)
}

// OPTIONS 注册OPTIONS路由
//
// OPTIONS方法通常用于CORS预检请求。
//
// 参数：
//   - relativePath: 相对路径
//   - handlers: 处理函数列表
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.OPTIONS("/api/*path", corsHandler)
func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("OPTIONS", relativePath, handlers...)
}

// HEAD 注册HEAD路由
//
// HEAD方法类似GET，但只返回响应头。
//
// 参数：
//   - relativePath: 相对路径
//   - handlers: 处理函数列表
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.HEAD("/users/:id", checkUserExistsHandler)
func (group *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.Handle("HEAD", relativePath, handlers...)
}

// Any 注册所有HTTP方法路由
//
// 为指定路径注册所有常用的HTTP方法路由。
//
// 参数：
//   - relativePath: 相对路径
//   - handlers: 处理函数列表
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.Any("/webhook", webhookHandler)
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

// =============================================================================
// 静态文件服务方法
// =============================================================================

// StaticFile 注册静态文件路由
//
// 为单个静态文件创建路由，支持GET和HEAD方法。
//
// 参数：
//   - relativePath: URL路径
//   - filepath: 文件系统路径
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.StaticFile("/favicon.ico", "./assets/favicon.ico")
func (group *RouterGroup) StaticFile(relativePath, filepath string) IRoutes {
	handler := func(c *Context) {
		c.File(filepath)
	}
	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
	return group.returnObj()
}

// Static 注册静态文件目录路由
//
// 为文件目录创建静态文件服务，这是StaticFS的便捷方法。
//
// 参数：
//   - relativePath: URL路径前缀
//   - root: 文件系统根目录路径
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.Static("/assets", "./public")
//	r.Static("/uploads", "./storage/uploads")
func (group *RouterGroup) Static(relativePath, root string) IRoutes {
	return group.StaticFS(relativePath, http.Dir(root))
}

// StaticFS 注册文件系统路由
//
// 使用自定义的文件系统提供静态文件服务。
//
// 参数：
//   - relativePath: URL路径前缀
//   - fs: 文件系统接口
//
// 返回：
//   - IRoutes: 支持链式调用
//
// 示例：
//
//	r.StaticFS("/static", http.Dir("./assets"))
//	r.StaticFS("/files", myCustomFileSystem)
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

// =============================================================================
// 辅助方法
// =============================================================================

// combineHandlers 合并处理函数列表
//
// 将路由组的中间件和新的处理函数合并为一个统一的处理函数列表。
// 组中间件会在新处理函数之前执行。
//
// 参数：
//   - handlers: 新的处理函数列表
//
// 返回：
//   - []HandlerFunc: 合并后的处理函数列表
func (group *RouterGroup) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	// 计算合并后的总长度
	finalSize := len(group.Handlers) + len(handlers)
	// 创建新的切片存储合并结果
	mergedHandlers := make([]HandlerFunc, finalSize)
	// 先复制组中间件
	copy(mergedHandlers, group.Handlers)
	// 再复制新添加的处理函数
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

// calculateAbsolutePath 计算绝对路径
//
// 将路由组的基础路径和相对路径组合成绝对路径。
//
// 参数：
//   - relativePath: 相对路径
//
// 返回：
//   - string: 绝对路径
func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

// returnObj 返回适当的路由接口实例
//
// 如果是根路由组，返回Engine实例；否则返回路由组实例。
// 这样设计可以实现链式调用。
//
// 返回：
//   - IRoutes: 路由接口实例
func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.engine
	}
	return group
}
