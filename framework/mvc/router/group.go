package router

import (
	"context"
	"strings"

	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/define"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
)

// Group 路由组结构体
//
// Group 提供了分层的路由管理能力，支持：
// - 路由前缀分组管理
// - 中间件链继承和层叠
// - 嵌套路由组支持
// - 7种不同类型的处理器注册
// - 所有HTTP方法的完整支持
type Group struct {
	router     *Router                  // 路由器实例，负责实际的路由注册
	prefix     string                   // 路由组的URL前缀，如 "/api/v1"
	middleware []middleware.HandlerFunc // 中间件链，在所有路由上执行
	parent     *Group                   // 父路由组引用，用于实现嵌套和继承
}

// NewGroup 创建新的路由组
//
// 参数:
//   - router: 路由器实例，用于注册实际的路由
//   - prefix: 路由前缀，如 "/api", "/v1", "/admin" 等
//
// 返回值:
//   - *Group: 新创建的路由组实例
//
// 示例:
//
//	apiGroup := NewGroup(router, "/api/v1")
//	adminGroup := NewGroup(router, "/admin")
func NewGroup(router *Router, prefix string) *Group {
	return &Group{
		router: router,
		prefix: prefix,
		parent: nil,
	}
}

// Group 创建子路由组
//
// 在当前路由组的基础上创建一个子组，子组会继承父组的所有中间件和前缀。
//
// 参数:
//   - prefix: 子组的相对前缀，会自动与父组前缀合并
//
// 返回值:
//   - *Group: 新的子路由组实例
//
// 示例:
//
//	apiGroup := group.Group("/api")    // 前缀: /api
//	v1Group := apiGroup.Group("/v1")  // 前缀: /api/v1
//	userGroup := v1Group.Group("/users") // 前缀: /api/v1/users
func (g *Group) Group(prefix string) *Group {
	newPrefix := g.prefix
	if newPrefix != "/" {
		newPrefix += "/" + prefix
	} else {
		newPrefix = prefix
	}

	// 清理重复的斜杠
	for strings.Contains(newPrefix, "//") {
		newPrefix = strings.ReplaceAll(newPrefix, "//", "/")
	}

	return &Group{
		router:     g.router,
		prefix:     newPrefix,
		middleware: g.middleware, // 继承父组的中间件
		parent:     g,            // 设置父组引用
	}
}

// Use 添加中间件到当前路由组
//
// 添加的中间件将在所有通过该组注册的路由上执行，
// 包括子组中的路由。中间件的执行顺序为添加顺序。
//
// 参数:
//   - middleware: 一个或多个中间件函数
//
// 示例:
//
//	group.Use(authMiddleware, logMiddleware)
//	group.Use(corsMiddleware)
func (g *Group) Use(middleware ...middleware.HandlerFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// RegisterController 在路由组中注册控制器
//
// 将控制器注册到指定路径，路径会自动加上组前缀。
// 控制器的所有方法都会被自动注册为路由。
//
// 参数:
//   - path: 控制器的相对路径
//   - ctrl: 控制器实例，必须实现 IController 接口
//
// 示例:
//
//	userGroup.RegisterController("/profile", &UserController{})
//	// 生成路由: /api/v1/users/profile/[action]
func (g *Group) RegisterController(path string, ctrl core.IController) {
	fullPath := g.prefix + path
	g.router.RegisterController(fullPath, ctrl)
}

// GetFullPrefix 获取包含所有父级前缀的完整路径
//
// 递归向上遍历父组，构建完整的路径前缀。
// 主要用于调试和日志输出。
//
// 返回值:
//   - string: 完整的路径前缀，包含所有父级前缀
//
// 示例:
//
//	如果组层次为 root("/api") -> v1("/v1") -> users("/users")
//	则 users组的 GetFullPrefix() 返回 "/api/v1/users"
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
//
// 递归向上收集所有父组的中间件，按照层次从根到叶子的顺序组装。
// 主要用于调试和中间件链分析。
//
// 返回值:
//   - []middleware.HandlerFunc: 完整的中间件链，包含所有父级中间件
//
// 示例:
//
//	如果 root组有[auth], v1组有[cors], users组有[validate]
//	则 users组的 GetAllMiddleware() 返回 [auth, cors, validate]
func (g *Group) GetAllMiddleware() []middleware.HandlerFunc {
	var allMiddleware []middleware.HandlerFunc
	if g.parent != nil {
		allMiddleware = append(allMiddleware, g.parent.GetAllMiddleware()...)
	}
	allMiddleware = append(allMiddleware, g.middleware...)
	return allMiddleware
}

// ===== 基础路由注册方法 =====
//
// 以下方法使用标准的 HandlerFunc 类型，提供完整的请求处理能力
// HandlerFunc 签名: func(context.Context, *RequestContext)

// GET 注册标准GET处理器
//
// 适用场景: 复杂的业务逻辑、需要完整请求上下文的数据查询
// 示例: 复杂的数据查询、需要中间件支持的操作、有超时控制的请求
func (g *Group) GET(path string, handler define.HandlerFunc) {
	g.addRoute("GET", path, handler)
}

// POST 注册标准POST处理器
//
// 适用场景: 复杂的数据创建、需要事务支持的操作、复杂的业务逻辑
func (g *Group) POST(path string, handler define.HandlerFunc) {
	g.addRoute("POST", path, handler)
}

// PUT 注册标准PUT处理器
//
// 适用场景: 复杂的资源更新、需要并发控制的修改操作
func (g *Group) PUT(path string, handler define.HandlerFunc) {
	g.addRoute("PUT", path, handler)
}

// DELETE 注册标准DELETE处理器
//
// 适用场景: 复杂的删除逻辑、需要权限检查的删除操作、有级联关系的删除
func (g *Group) DELETE(path string, handler define.HandlerFunc) {
	g.addRoute("DELETE", path, handler)
}

// PATCH 注册标准PATCH处理器
//
// 适用场景: 复杂的部分更新、需要合并策略的修改、有条件的部分更新
func (g *Group) PATCH(path string, handler define.HandlerFunc) {
	g.addRoute("PATCH", path, handler)
}

// HEAD 注册标准HEAD处理器
//
// 适用场景: 复杂的资源存在性检查、需要权限验证的元数据获取
func (g *Group) HEAD(path string, handler define.HandlerFunc) {
	g.addRoute("HEAD", path, handler)
}

// OPTIONS 注册标准OPTIONS处理器
//
// 适用场景: 复杂的CORS处理、动态的API能力描述、需要权限检查的预检处理
func (g *Group) OPTIONS(path string, handler define.HandlerFunc) {
	g.addRoute("OPTIONS", path, handler)
}

// Any 注册标准万能处理器（支持所有HTTP方法）
//
// 适用场景: 复杂的通用接口、需要根据方法分发的复杂逻辑、通用的中间件处理
func (g *Group) Any(path string, handler define.HandlerFunc) {
	g.addRoute("ANY", path, handler)
}

// addRoute 添加路由（内部方法）
//
// 统一的路由注册方法，被所有具体的HTTP方法注册函数调用。
// 负责处理路由前缀合并、中间件包装和最终注册。
//
// 参数:
//   - method: HTTP方法（GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, ANY）
//   - path: 相对路径，会与组前缀合并
//   - handler: 已适配的标准处理函数
func (g *Group) addRoute(method, path string, handler define.HandlerFunc) {
	fullPath := g.prefix + path

	// 如果有中间件，需要包装处理函数
	if len(g.middleware) > 0 {
		handler = g.wrapWithMiddleware(handler)
	}

	g.router.registerRoute(method, fullPath, handler)
}

// wrapWithMiddleware 使用中间件包装处理函数（内部方法）
//
// 将路由组的中间件链包装到处理函数周围，实现中间件的自动执行。
// 中间件按照添加顺序执行，在原始处理函数之前运行。
//
// 参数:
//   - handler: 原始的处理函数
//
// 返回值:
//   - define.HandlerFunc: 包装后的处理函数，包含中间件执行逻辑
func (g *Group) wrapWithMiddleware(handler define.HandlerFunc) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		// 创建中间件上下文
		middlewareCtx := &middleware.Context{}
		middlewareCtx.SetRequest(c)

		// 顺序执行中间件链
		for _, mw := range g.middleware {
			mw(middlewareCtx)
		}
		// 执行原始处理函数
		handler(ctx, c)
	}
}

// ===== 扩展的多处理器路由注册方法 =====
//
// YYHertz框架提供7种不同的处理器类型，以满足不同场景的性能和功能需求：
//
// 1. LightHandlerFunc   - 最轻量级，无参数，适用于健康检查
// 2. SimpleHandlerFunc  - 简单逻辑，只需context，适用于日志记录
// 3. DirectHandlerFunc  - 直接访问，只需RequestContext，适用于快速响应
// 4. HandlerFunc        - 标准处理器，完整访问，适用于复杂业务逻辑
// 5. ResponseHandlerFunc - 自动JSON，返回任意数据，适用于REST API
// 6. AsyncHandlerFunc   - 异步处理，返回channel，适用于耗时操作
// 7. StreamHandlerFunc  - 流式传输，支持大数据，适用于文件传输

// ===== GET 方法的多种处理器支持 =====

// GETSimple 注册简单GET处理器
//
// 适用场景:
//   - 简单的业务逻辑处理
//   - 需要context进行超时控制
//   - 日志记录和统计计数
//   - 触发器和通知发送
//
// 示例:
//
//	group.GETSimple("/ping", func(ctx context.Context) {
//	    log.Println("收到ping请求")
//	    // 自动返回200 OK
//	})
func (g *Group) GETSimple(path string, handler define.SimpleHandlerFunc) {
	g.addRoute("GET", path, adaptSimpleHandler(handler))
}

// GETDirect 注册直接GET处理器
//
// 适用场景:
//   - 需要直接访问请求上下文
//   - 自定义响应格式和状态码
//   - 简洁的请求处理逻辑
//   - 性能敏感的快速响应
//
// 示例:
//
//	group.GETDirect("/info", func(c *define.RequestContext) {
//	    c.SetContentType("application/json")
//	    c.WriteString(`{"status":"ok"}`)
//	})
func (g *Group) GETDirect(path string, handler define.DirectHandlerFunc) {
	g.addRoute("GET", path, adaptDirectHandler(handler))
}

// GETLight 注册轻量级GET处理器
//
// 适用场景:
//   - 健康检查端点
//   - 探活和监控检测
//   - 简单的状态返回
//   - 最高性能要求的场景
//
// 示例:
//
//	group.GETLight("/health", func() {
//	    // 自动返回200 OK，无需任何处理
//	})
func (g *Group) GETLight(path string, handler define.LightHandlerFunc) {
	g.addRoute("GET", path, adaptLightHandler(handler))
}

// GETResponse 注册响应GET处理器
//
// 适用场景:
//   - REST API数据接口
//   - JSON格式的响应数据
//   - 自动序列化返回值
//   - 标准的CRUD操作
//
// 示例:
//
//	group.GETResponse("/users", func(c *define.RequestContext) any {
//	    users := getUserList()
//	    return map[string]any{
//	        "success": true,
//	        "data": users,
//	    }
//	})
func (g *Group) GETResponse(path string, handler define.ResponseHandlerFunc) {
	g.addRoute("GET", path, adaptResponseHandler(handler))
}

// GETAsync 注册异步GET处理器
//
// 适用场景:
//   - 耗时的数据查询
//   - 需要等待外部服务响应
//   - 异步任务处理
//   - 支持超时控制的操作
//
// 示例:
//
//	group.GETAsync("/report", func(c *define.RequestContext) <-chan any {
//	    resultChan := make(chan any, 1)
//	    go func() {
//	        defer close(resultChan)
//	        report := generateReport() // 耗时操作
//	        resultChan <- report
//	    }()
//	    return resultChan
//	})
func (g *Group) GETAsync(path string, handler define.AsyncHandlerFunc) {
	g.addRoute("GET", path, adaptAsyncHandler(handler))
}

// GETStream 注册流式GET处理器
//
// 适用场景:
//   - 大文件下载
//   - 实时数据流传输
//   - 分块数据传输
//   - 减少内存占用的数据传输
//
// 示例:
//
//	group.GETStream("/download", func(c *define.RequestContext, dataChan chan<- []byte) error {
//	    file, err := os.Open("large-file.zip")
//	    if err != nil {
//	        return err
//	    }
//	    defer file.Close()
//
//	    buffer := make([]byte, 1024)
//	    for {
//	        n, err := file.Read(buffer)
//	        if n > 0 {
//	            dataChan <- buffer[:n]
//	        }
//	        if err != nil {
//	            break
//	        }
//	    }
//	    return nil
//	})
func (g *Group) GETStream(path string, handler define.StreamHandlerFunc) {
	g.addRoute("GET", path, adaptStreamHandler(handler))
}

// ===== POST 方法的多种处理器支持 =====

// POSTSimple 注册简单POST处理器
//
// 适用场景: 简单的数据提交、事件触发、通知发送
// 示例: 接收webhook通知、触发定时任务、记录操作日志
func (g *Group) POSTSimple(path string, handler define.SimpleHandlerFunc) {
	g.addRoute("POST", path, adaptSimpleHandler(handler))
}

// POSTDirect 注册直接POST处理器
//
// 适用场景: 需要完全控制请求解析和响应格式的数据提交
// 示例: 自定义格式的数据上传、特殊协议的API接口
func (g *Group) POSTDirect(path string, handler define.DirectHandlerFunc) {
	g.addRoute("POST", path, adaptDirectHandler(handler))
}

// POSTLight 注册轻量级POST处理器
//
// 适用场景: 简单的状态更新、开关切换、快速确认操作
// 示例: 点赞操作、状态切换、简单的配置更新
func (g *Group) POSTLight(path string, handler define.LightHandlerFunc) {
	g.addRoute("POST", path, adaptLightHandler(handler))
}

// POSTResponse 注册响应POST处理器
//
// 适用场景: 标准的数据创建、表单提交、REST API的CREATE操作
// 示例: 用户注册、商品创建、订单提交
func (g *Group) POSTResponse(path string, handler define.ResponseHandlerFunc) {
	g.addRoute("POST", path, adaptResponseHandler(handler))
}

// POSTAsync 注册异步POST处理器
//
// 适用场景: 耗时的数据处理、批量导入、后台任务提交
// 示例: 文件上传处理、数据分析任务、邮件发送队列
func (g *Group) POSTAsync(path string, handler define.AsyncHandlerFunc) {
	g.addRoute("POST", path, adaptAsyncHandler(handler))
}

// POSTStream 注册流式POST处理器
//
// 适用场景: 大文件上传、实时数据接收、分块数据处理
// 示例: 视频上传、日志收集、实时监控数据接收
func (g *Group) POSTStream(path string, handler define.StreamHandlerFunc) {
	g.addRoute("POST", path, adaptStreamHandler(handler))
}

// ===== PUT 方法的多种处理器支持 =====

// PUTSimple 注册简单PUT处理器
//
// 适用场景: 简单的资源更新、状态修改、配置变更
// 示例: 更新用户状态、修改系统配置、切换功能开关
func (g *Group) PUTSimple(path string, handler define.SimpleHandlerFunc) {
	g.addRoute("PUT", path, adaptSimpleHandler(handler))
}

// PUTDirect 注册直接PUT处理器
//
// 适用场景: 需要精确控制更新逻辑的资源修改
// 示例: 复杂的数据更新、自定义验证逻辑、特殊格式处理
func (g *Group) PUTDirect(path string, handler define.DirectHandlerFunc) {
	g.addRoute("PUT", path, adaptDirectHandler(handler))
}

// PUTLight 注册轻量级PUT处理器
//
// 适用场景: 快速的状态切换、简单的标记更新
// 示例: 已读/未读标记、启用/禁用状态、简单的计数器更新
func (g *Group) PUTLight(path string, handler define.LightHandlerFunc) {
	g.addRoute("PUT", path, adaptLightHandler(handler))
}

// PUTResponse 注册响应PUT处理器
//
// 适用场景: 标准的资源更新、REST API的UPDATE操作
// 示例: 用户信息更新、商品信息修改、订单状态变更
func (g *Group) PUTResponse(path string, handler define.ResponseHandlerFunc) {
	g.addRoute("PUT", path, adaptResponseHandler(handler))
}

// PUTAsync 注册异步PUT处理器
//
// 适用场景: 耗时的数据更新、批量修改、复杂的业务逻辑处理
// 示例: 批量数据更新、复杂的计算更新、需要外部API调用的更新
func (g *Group) PUTAsync(path string, handler define.AsyncHandlerFunc) {
	g.addRoute("PUT", path, adaptAsyncHandler(handler))
}

// PUTStream 注册流式PUT处理器
//
// 适用场景: 大数据量的资源替换、分块数据更新
// 示例: 大文件替换、数据库备份还原、批量数据导入更新
func (g *Group) PUTStream(path string, handler define.StreamHandlerFunc) {
	g.addRoute("PUT", path, adaptStreamHandler(handler))
}

// ===== DELETE 方法的多种处理器支持 =====

// DELETESimple 注册简单DELETE处理器
//
// 适用场景: 简单的资源删除、记录清除、临时数据清理
// 示例: 删除缓存、清理临时文件、移除简单标记
func (g *Group) DELETESimple(path string, handler define.SimpleHandlerFunc) {
	g.addRoute("DELETE", path, adaptSimpleHandler(handler))
}

// DELETEDirect 注册直接DELETE处理器
//
// 适用场景: 需要精确控制删除逻辑的资源移除
// 示例: 复杂的级联删除、有条件的删除操作、自定义删除验证
func (g *Group) DELETEDirect(path string, handler define.DirectHandlerFunc) {
	g.addRoute("DELETE", path, adaptDirectHandler(handler))
}

// DELETELight 注册轻量级DELETE处理器
//
// 适用场景: 快速的标记删除、简单的状态重置
// 示例: 软删除标记、快速清除操作、简单的重置功能
func (g *Group) DELETELight(path string, handler define.LightHandlerFunc) {
	g.addRoute("DELETE", path, adaptLightHandler(handler))
}

// DELETEResponse 注册响应DELETE处理器
//
// 适用场景: 标准的资源删除、REST API的DELETE操作
// 示例: 用户删除、商品下架、订单取消
func (g *Group) DELETEResponse(path string, handler define.ResponseHandlerFunc) {
	g.addRoute("DELETE", path, adaptResponseHandler(handler))
}

// DELETEAsync 注册异步DELETE处理器
//
// 适用场景: 耗时的删除操作、批量删除、复杂的清理任务
// 示例: 大量数据删除、文件系统清理、复杂的业务逻辑删除
func (g *Group) DELETEAsync(path string, handler define.AsyncHandlerFunc) {
	g.addRoute("DELETE", path, adaptAsyncHandler(handler))
}

// DELETEStream 注册流式DELETE处理器
//
// 适用场景: 大规模数据删除、分批删除操作、实时删除反馈
// 示例: 批量文件删除、大表数据清理、分块删除操作
func (g *Group) DELETEStream(path string, handler define.StreamHandlerFunc) {
	g.addRoute("DELETE", path, adaptStreamHandler(handler))
}

// ===== PATCH 方法的多种处理器支持 =====

// PATCHSimple 注册简单PATCH处理器
//
// 适用场景: 简单的部分字段更新、单一属性修改
// 示例: 更新用户昵称、修改商品价格、调整配置项
func (g *Group) PATCHSimple(path string, handler define.SimpleHandlerFunc) {
	g.addRoute("PATCH", path, adaptSimpleHandler(handler))
}

// PATCHDirect 注册直接PATCH处理器
//
// 适用场景: 需要精确控制的部分更新操作
// 示例: 复杂的字段验证更新、有条件的部分修改
func (g *Group) PATCHDirect(path string, handler define.DirectHandlerFunc) {
	g.addRoute("PATCH", path, adaptDirectHandler(handler))
}

// PATCHLight 注册轻量级PATCH处理器
//
// 适用场景: 快速的单字段更新、简单的增量修改
// 示例: 计数器增减、状态位切换、时间戳更新
func (g *Group) PATCHLight(path string, handler define.LightHandlerFunc) {
	g.addRoute("PATCH", path, adaptLightHandler(handler))
}

// PATCHResponse 注册响应PATCH处理器
//
// 适用场景: 标准的部分资源更新、REST API的PATCH操作
// 示例: 用户资料部分更新、商品信息局部修改、设置项调整
func (g *Group) PATCHResponse(path string, handler define.ResponseHandlerFunc) {
	g.addRoute("PATCH", path, adaptResponseHandler(handler))
}

// PATCHAsync 注册异步PATCH处理器
//
// 适用场景: 需要复杂计算的部分更新、耗时的增量修改
// 示例: 数据重新计算更新、外部API同步更新、复杂业务逻辑的部分修改
func (g *Group) PATCHAsync(path string, handler define.AsyncHandlerFunc) {
	g.addRoute("PATCH", path, adaptAsyncHandler(handler))
}

// PATCHStream 注册流式PATCH处理器
//
// 适用场景: 大数据的增量更新、分块部分修改
// 示例: 大文件的部分替换、数据库的增量更新、分批字段修改
func (g *Group) PATCHStream(path string, handler define.StreamHandlerFunc) {
	g.addRoute("PATCH", path, adaptStreamHandler(handler))
}

// ===== HEAD 方法的多种处理器支持 =====

// HEADSimple 注册简单HEAD处理器
//
// 适用场景: 简单的资源存在性检查、基础元数据获取
// 示例: 检查文件是否存在、验证用户权限、简单的状态检查
func (g *Group) HEADSimple(path string, handler define.SimpleHandlerFunc) {
	g.addRoute("HEAD", path, adaptSimpleHandler(handler))
}

// HEADDirect 注册直接HEAD处理器
//
// 适用场景: 需要设置详细响应头的资源检查
// 示例: 文件信息检查、自定义元数据返回、复杂的头部设置
func (g *Group) HEADDirect(path string, handler define.DirectHandlerFunc) {
	g.addRoute("HEAD", path, adaptDirectHandler(handler))
}

// HEADLight 注册轻量级HEAD处理器
//
// 适用场景: 快速的存在性检查、最小开销的资源验证
// 示例: 快速可用性检查、简单的资源存在确认
func (g *Group) HEADLight(path string, handler define.LightHandlerFunc) {
	g.addRoute("HEAD", path, adaptLightHandler(handler))
}

// HEADResponse 注册响应HEAD处理器
//
// 适用场景: 标准的资源元数据检查、REST API的HEAD操作
// 示例: 获取资源信息、检查资源状态、元数据查询
func (g *Group) HEADResponse(path string, handler define.ResponseHandlerFunc) {
	g.addRoute("HEAD", path, adaptResponseHandler(handler))
}

// HEADAsync 注册异步HEAD处理器
//
// 适用场景: 需要复杂检查的资源验证、耗时的元数据获取
// 示例: 远程资源检查、复杂权限验证、需要计算的元数据
func (g *Group) HEADAsync(path string, handler define.AsyncHandlerFunc) {
	g.addRoute("HEAD", path, adaptAsyncHandler(handler))
}

// HEADStream 注册流式HEAD处理器
//
// 适用场景: 大文件的元数据检查、分块资源的信息获取
// 示例: 大文件信息检查、分块上传状态验证、流式资源元数据
func (g *Group) HEADStream(path string, handler define.StreamHandlerFunc) {
	g.addRoute("HEAD", path, adaptStreamHandler(handler))
}

// ===== OPTIONS 方法的多种处理器支持 =====

// OPTIONSSimple 注册简单OPTIONS处理器
//
// 适用场景: 简单的CORS预检处理、基础的方法支持查询
// 示例: 标准CORS响应、简单的API能力查询
func (g *Group) OPTIONSSimple(path string, handler define.SimpleHandlerFunc) {
	g.addRoute("OPTIONS", path, adaptSimpleHandler(handler))
}

// OPTIONSDirect 注册直接OPTIONS处理器
//
// 适用场景: 需要自定义CORS头部、复杂的预检处理
// 示例: 动态CORS配置、自定义访问控制、详细的方法支持说明
func (g *Group) OPTIONSDirect(path string, handler define.DirectHandlerFunc) {
	g.addRoute("OPTIONS", path, adaptDirectHandler(handler))
}

// OPTIONSLight 注册轻量级OPTIONS处理器
//
// 适用场景: 快速的CORS响应、最小开销的预检处理
// 示例: 简单的跨域支持、快速的方法检查
func (g *Group) OPTIONSLight(path string, handler define.LightHandlerFunc) {
	g.addRoute("OPTIONS", path, adaptLightHandler(handler))
}

// OPTIONSResponse 注册响应OPTIONS处理器
//
// 适用场景: 标准的API能力描述、详细的方法支持信息
// 示例: API文档响应、支持的方法和参数说明、服务能力描述
func (g *Group) OPTIONSResponse(path string, handler define.ResponseHandlerFunc) {
	g.addRoute("OPTIONS", path, adaptResponseHandler(handler))
}

// OPTIONSAsync 注册异步OPTIONS处理器
//
// 适用场景: 需要复杂计算的API能力查询、动态权限检查
// 示例: 复杂的权限检查、动态API能力计算、外部服务依赖的能力查询
func (g *Group) OPTIONSAsync(path string, handler define.AsyncHandlerFunc) {
	g.addRoute("OPTIONS", path, adaptAsyncHandler(handler))
}

// OPTIONSStream 注册流式OPTIONS处理器
//
// 适用场景: 大量API的能力描述、分块的服务发现
// 示例: 微服务能力发现、大量接口的批量描述、分块的API文档
func (g *Group) OPTIONSStream(path string, handler define.StreamHandlerFunc) {
	g.addRoute("OPTIONS", path, adaptStreamHandler(handler))
}

// ===== Any 方法的多种处理器支持 =====

// AnySimple 注册简单Any处理器（支持所有HTTP方法）
//
// 适用场景: 通用的请求处理、不区分HTTP方法的简单逻辑
// 示例: 通用的日志记录、统计计数、简单的回调处理
func (g *Group) AnySimple(path string, handler define.SimpleHandlerFunc) {
	g.addRoute("ANY", path, adaptSimpleHandler(handler))
}

// AnyDirect 注册直接Any处理器（支持所有HTTP方法）
//
// 适用场景: 需要根据HTTP方法进行不同处理的通用接口
// 示例: 通用的资源处理器、根据方法分发的处理逻辑
func (g *Group) AnyDirect(path string, handler define.DirectHandlerFunc) {
	g.addRoute("ANY", path, adaptDirectHandler(handler))
}

// AnyLight 注册轻量级Any处理器（支持所有HTTP方法）
//
// 适用场景: 最轻量级的通用处理、快速响应的万能接口
// 示例: 通用的健康检查、简单的状态确认、快速响应接口
func (g *Group) AnyLight(path string, handler define.LightHandlerFunc) {
	g.addRoute("ANY", path, adaptLightHandler(handler))
}

// AnyResponse 注册响应Any处理器（支持所有HTTP方法）
//
// 适用场景: 通用的REST风格接口、支持多种操作的资源接口
// 示例: 通用资源API、支持CRUD操作的统一接口、灵活的数据处理接口
func (g *Group) AnyResponse(path string, handler define.ResponseHandlerFunc) {
	g.addRoute("ANY", path, adaptResponseHandler(handler))
}

// AnyAsync 注册异步Any处理器（支持所有HTTP方法）
//
// 适用场景: 通用的异步任务处理、支持多种触发方式的异步操作
// 示例: 通用任务队列接口、支持多种方式的异步处理、灵活的后台任务
func (g *Group) AnyAsync(path string, handler define.AsyncHandlerFunc) {
	g.addRoute("ANY", path, adaptAsyncHandler(handler))
}

// AnyStream 注册流式Any处理器（支持所有HTTP方法）
//
// 适用场景: 通用的流式数据处理、支持多种操作的数据流接口
// 示例: 通用的数据流接口、支持上传下载的统一处理、灵活的数据传输
func (g *Group) AnyStream(path string, handler define.StreamHandlerFunc) {
	g.addRoute("ANY", path, adaptStreamHandler(handler))
}
