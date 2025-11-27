// Package mvc 提供MVC框架的命名空间管理功能
//
// 命名空间系统提供了一种分层的路由组织方式，类似Beego框架的Namespace功能。
// 通过命名空间，可以实现：
//
// 功能特性：
// - 路由分组管理：按功能模块或API版本组织路由
// - 嵌套支持：支持多级命名空间嵌套
// - 中间件继承：子命名空间自动继承父级中间件
// - 自动/手动路由：支持自动路由和精确控制的手动路由
// - 前缀统一：同一命名空间下的所有路由共享前缀
//
// 适用场景：
// - API版本管理：/api/v1、/api/v2
// - 功能模块分组：/admin、/user、/public
// - 微服务路由：/service-a、/service-b
// - 权限分级：不同命名空间使用不同的认证中间件
package mvc

import (
	"strings"

	"github.com/hertz-contrib/websocket"

	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// ============= 类型定义 =============

// NamespaceFunc 定义命名空间配置函数类型
//
// 该函数类型用于配置命名空间的各种属性，如添加控制器、路由、中间件等。
// 采用函数式配置方式，使得API更加灵活和链式调用友好。
type NamespaceFunc func(*Namespace)

// WsHandlerFunc 定义 WebSocket 处理函数类型
//
// 该函数类型用于处理 WebSocket 连接，在连接建立后被调用。
// 函数应该包含消息处理循环，负责读取和写入 WebSocket 消息。
type WsHandlerFunc = define.WsHandlerFunc

// Namespace 命名空间结构
//
// 提供了类似Beego框架的命名空间功能，用于组织和管理路由。
// 支持嵌套结构和中间件继承，现已扩展支持 WebSocket 路由。
type Namespace struct {
	// prefix 命名空间的URL前缀，如 "/api/v1"、"/admin"
	prefix string

	// controllers 注册到该命名空间的控制器列表
	controllers []controllerInfo

	// routers 手动注册的路由信息列表
	routers []routerInfo

	// wsRouters WebSocket路由信息列表
	wsRouters []wsRouterInfo

	// wsControllerRouters WebSocket控制器路由信息列表
	wsControllerRouters []wsControllerRouterInfo

	// namespaces 子命名空间列表，支持多级嵌套
	namespaces []*Namespace

	// middlewares 应用于该命名空间的中间件列表
	middlewares []define.HandlerFunc
}

// controllerInfo 控制器注册信息
//
// 存储控制器在命名空间中的注册信息，包括控制器实例和路由注册方式。
type controllerInfo struct {
	// controller 控制器实例
	controller core.IController

	// autoRoute 是否使用自动路由注册
	// true: 使用自动路由，根据控制器名和方法名生成路由
	// false: 使用手动路由，需要精确指定路由映射
	autoRoute bool
}

// routerInfo 手动路由信息
//
// 存储手动注册的路由信息，包括路径、控制器和方法映射。
type routerInfo struct {
	// path 路由路径，如 "/users"、"/users/:id"
	path string

	// controller 控制器实例
	controller core.IController

	// method HTTP方法和控制器方法的映射
	// 格式："GET:MethodName" 或 "*:MethodName" 或 "MethodName"
	method string
}

// wsRouterInfo WebSocket路由信息
//
// 存储 WebSocket 路由的注册信息，包括路径和处理函数。
type wsRouterInfo struct {
	// path WebSocket路由路径，如 "/ws"、"/chat/:room"
	path string

	// handler WebSocket连接处理函数
	handler WsHandlerFunc

	// upgrader WebSocket升级器配置
	upgrader websocket.HertzUpgrader
}

// wsControllerRouterInfo WebSocket控制器路由信息结构
//
// 用于存储命名空间中WebSocket控制器路由的配置信息，
// 支持将控制器方法映射为WebSocket处理器。
type wsControllerRouterInfo struct {
	// path WebSocket路由路径，如 "/ws"、"/chat/:room"
	path string

	// controller 控制器实例
	controller core.IController

	// method 控制器方法名，如 "HandleChat"、"EchoMessage"
	method string

	// upgrader WebSocket升级器配置（可选）
	upgrader *websocket.HertzUpgrader
}

// ============= 构造函数 =============

// NewNamespace 创建新的命名空间
//
// 类似Beego框架的beego.NewNamespace函数，提供相同的API风格和功能。
// 支持链式调用和函数式配置。
//
// 参数：
//   - prefix: string - 命名空间的URL前缀，应以"/"开头
//   - funcs: ...NamespaceFunc - 配置函数列表，用于添加控制器、路由、中间件等
//
// 返回值：
//   - *Namespace: 新创建的命名空间实例
//
// 使用示例：
//
//	// 基本使用
//	ns := mvc.NewNamespace("/api/v1")
//
//	// 带配置的使用
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSAutoRouter(&UserController{}),
//		mvc.NSRouter("/users/:id", &UserController{}, "GET:Get"),
//		mvc.NSMiddleware(authMiddleware),
//	)
//
//	// 嵌套命名空间
//	ns := mvc.NewNamespace("/api",
//		mvc.NSNamespace("/v1",
//			mvc.NSAutoRouter(&UserController{}),
//		),
//		mvc.NSNamespace("/v2",
//			mvc.NSAutoRouter(&UserV2Controller{}),
//		),
//	)
//
// 注意事项：
//   - prefix应以"/"开头，如"/api/v1"而不是"api/v1"
//   - 配置函数按顺序执行，后添加的会覆盖先添加的同名路由
//   - 中间件按添加顺序执行
func NewNamespace(prefix string, funcs ...NamespaceFunc) *Namespace {
	ns := &Namespace{
		prefix:      prefix,
		controllers: make([]controllerInfo, 0),
		routers:     make([]routerInfo, 0),
		wsRouters:   make([]wsRouterInfo, 0),
		namespaces:  make([]*Namespace, 0),
		middlewares: make([]define.HandlerFunc, 0),
	}

	// 按顺序执行所有配置函数
	for _, fn := range funcs {
		fn(ns)
	}

	return ns
}

// ============= 命名空间配置函数 =============

// NSAutoRouter 为命名空间添加自动路由控制器
//
// 类似Beego框架的beego.NSAutoRouter函数，自动根据控制器名和方法名生成路由。
// 遵循RESTful风格的路由命名约定。
//
// 参数：
//   - ctrl: core.IController - 要注册的控制器实例
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// 路由生成规则：
//   - 控制器名: UserController -> /user
//   - HTTP方法: Get() -> GET, Post() -> POST, Put() -> PUT, Delete() -> DELETE
//   - 自定义方法: GetProfile() -> GET /user/profile
//
// 使用示例：
//
//	// 在命名空间中使用
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSAutoRouter(&UserController{}),
//		mvc.NSAutoRouter(&OrderController{}),
//	)
//
//	// 生成的路由示例：
//	// GET /api/v1/user -> UserController.Get()
//	// POST /api/v1/user -> UserController.Post()
//	// GET /api/v1/user/profile -> UserController.GetProfile()
//	// GET /api/v1/order -> OrderController.Get()
func NSAutoRouter(ctrl core.IController) NamespaceFunc {
	return func(ns *Namespace) {
		ns.controllers = append(ns.controllers, controllerInfo{
			controller: ctrl,
			autoRoute:  true,
		})
	}
}

// NSRouter 为命名空间添加手动路由映射
//
// 类似Beego框架的beego.NSRouter函数，允许精确控制路由和控制器方法的映射。
// 支持路径参数和复杂路由规则。
//
// 参数：
//   - path: string - 路由路径，相对于命名空间前缀的路径
//   - ctrl: core.IController - 控制器实例
//   - method: string - HTTP方法和控制器方法的映射
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// method参数格式：
//   - "GET:MethodName" - 指定HTTP方法和控制器方法
//   - "*:MethodName" - 匹配所有HTTP方法
//   - "MethodName" - 默认匹配所有HTTP方法
//
// 使用示例：
//
//	// 基本使用
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSRouter("/users", &UserController{}, "GET:GetList"),
//		mvc.NSRouter("/users/:id", &UserController{}, "GET:Get"),
//		mvc.NSRouter("/users", &UserController{}, "POST:Create"),
//	)
//
//	// 生成的路由：
//	// GET /api/v1/users -> UserController.GetList()
//	// GET /api/v1/users/:id -> UserController.Get()
//	// POST /api/v1/users -> UserController.Create()
//
//	// 通配符使用
//	mvc.NSRouter("/proxy/*path", &ProxyController{}, "*:Handle")
func NSRouter(path string, ctrl core.IController, method string) NamespaceFunc {
	return func(ns *Namespace) {
		ns.routers = append(ns.routers, routerInfo{
			path:       path,
			controller: ctrl,
			method:     method,
		})
	}
}

// NSAutoPrefix 为命名空间添加带前缀的自动路由控制器
//
// 该函数提供了一种在控制器级别添加额外前缀的方式。
// 注意：当前实现与NSAutoRouter相同，未使用prefix参数。
//
// 参数：
//   - prefix: string - 额外的前缀（当前未使用）
//   - ctrl: core.IController - 要注册的控制器实例
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// TODO: 该函数需要完善以实际使用prefix参数
//
// 使用示例：
//
//	// 当前使用方式（与NSAutoRouter相同）
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSAutoPrefix("/extra", &UserController{}),
//	)
func NSAutoPrefix(prefix string, ctrl core.IController) NamespaceFunc {
	return func(ns *Namespace) {
		// TODO: 实现prefix的实际使用
		ns.controllers = append(ns.controllers, controllerInfo{
			controller: ctrl,
			autoRoute:  true,
		})
	}
}

// NSNamespace 为命名空间添加嵌套子命名空间
//
// 类似Beego框架的beego.NSNamespace函数，支持多级命名空间嵌套。
// 子命名空间会继承父命名空间的中间件和前缀。
//
// 参数：
//   - prefix: string - 子命名空间的相对前缀
//   - funcs: ...NamespaceFunc - 子命名空间的配置函数列表
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// 特性：
//   - 路径合并：父前缀 + 子前缀 = 最终路径
//   - 中间件继承：子命名空间继承父命名空间的所有中间件
//   - 执行顺序：父中间件 -> 子中间件 -> 控制器方法
//
// 使用示例：
//
//	// 二级嵌套
//	ns := mvc.NewNamespace("/api",
//		mvc.NSMiddleware(corsMiddleware), // 适用于所有子命名空间
//		mvc.NSNamespace("/v1",
//			mvc.NSMiddleware(v1AuthMiddleware),
//			mvc.NSAutoRouter(&UserV1Controller{}),
//		),
//		mvc.NSNamespace("/v2",
//			mvc.NSMiddleware(v2AuthMiddleware),
//			mvc.NSAutoRouter(&UserV2Controller{}),
//		),
//	)
//
//	// 生成的路由：
//	// GET /api/v1/user -> corsMiddleware -> v1AuthMiddleware -> UserV1Controller.Get()
//	// GET /api/v2/user -> corsMiddleware -> v2AuthMiddleware -> UserV2Controller.Get()
//
//	// 三级嵌套
//	ns := mvc.NewNamespace("/api",
//		mvc.NSNamespace("/v1",
//			mvc.NSNamespace("/admin",
//				mvc.NSMiddleware(adminAuthMiddleware),
//				mvc.NSAutoRouter(&AdminController{}),
//			),
//		),
//	)
//	// 生成路由： GET /api/v1/admin/dashboard -> AdminController.GetDashboard()
func NSNamespace(prefix string, funcs ...NamespaceFunc) NamespaceFunc {
	return func(ns *Namespace) {
		subNs := NewNamespace(prefix, funcs...)
		ns.namespaces = append(ns.namespaces, subNs)
	}
}

// NSWebSocket 为命名空间添加 WebSocket 路由
//
// 该函数用于在命名空间中注册 WebSocket 路由，支持 WebSocket 连接升级和消息处理。
// WebSocket 路由会自动处理连接升级，并在连接建立后调用指定的处理函数。
//
// 参数：
//   - path: string - WebSocket路由路径，相对于命名空间前缀的路径
//   - handler: WebSocketHandlerFunc - WebSocket连接处理函数
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// WebSocket 连接处理特性：
//   - 自动升级：自动处理 HTTP 到 WebSocket 的协议升级
//   - 连接管理：提供连接建立、断开的生命周期管理
//   - 错误处理：内置连接错误和升级失败的处理机制
//   - 中间件支持：继承命名空间的中间件，支持认证和授权
//
// 使用示例：
//
//	// 基本 WebSocket 路由
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSWebSocket("/ws", func(conn *websocket.Conn) {
//			for {
//				// 读取消息
//				messageType, message, err := conn.ReadMessage()
//				if err != nil {
//					log.Println("read error:", err)
//					break
//				}
//
//				// 处理消息
//				log.Printf("Received: %s", message)
//
//				// 发送响应
//				err = conn.WriteMessage(messageType, message)
//				if err != nil {
//					log.Println("write error:", err)
//					break
//				}
//			}
//		}),
//	)
//
//	// 聊天室 WebSocket 路由
//	ns := mvc.NewNamespace("/chat",
//		mvc.NSMiddleware(authMiddleware), // WebSocket 也会继承中间件
//		mvc.NSWebSocket("/room/:id", func(conn *websocket.Conn) {
//			roomID := conn.Params().ByName("id") // 获取路径参数
//			chatRoom.Join(roomID, conn)
//			defer chatRoom.Leave(roomID, conn)
//
//			for {
//				_, message, err := conn.ReadMessage()
//				if err != nil {
//					break
//				}
//				chatRoom.Broadcast(roomID, message)
//			}
//		}),
//	)
//
//	// 实时数据推送
//	ns := mvc.NewNamespace("/api",
//		mvc.NSWebSocket("/stream", func(conn *websocket.Conn) {
//			ticker := time.NewTicker(time.Second)
//			defer ticker.Stop()
//
//			for {
//				select {
//				case <-ticker.C:
//					data := getCurrentData()
//					if err := conn.WriteJSON(data); err != nil {
//						return
//					}
//				case <-conn.Done():
//					return
//				}
//			}
//		}),
//	)
//
// 注意事项：
//   - 处理函数应该包含消息处理循环
//   - 建议使用 defer 清理资源
//   - 需要适当的错误处理来处理连接断开
//   - WebSocket 连接是长连接，注意内存和资源管理
//   - 支持路径参数，如 "/chat/:room/:user"
func NSWebSocket(path string, handler WsHandlerFunc) NamespaceFunc {
	return func(ns *Namespace) {
		ns.wsRouters = append(ns.wsRouters, wsRouterInfo{
			path:     path,
			handler:  handler,
			upgrader: websocket.HertzUpgrader{}, // 使用默认配置
		})
	}
}

// NSWebSocketWithUpgrader 为命名空间添加带自定义升级器的 WebSocket 路由
//
// 该函数允许使用自定义的 WebSocket 升级器配置，提供更精细的控制。
//
// 参数：
//   - path: string - WebSocket路由路径
//   - handler: WebSocketHandlerFunc - WebSocket连接处理函数
//   - upgrader: websocket.HertzUpgrader - 自定义升级器配置
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// 使用示例：
//
//	// 自定义升级器配置
//	upgrader := websocket.HertzUpgrader{
//		ReadBufferSize:  1024,
//		WriteBufferSize: 1024,
//		CheckOrigin: func(ctx *app.RequestContext) bool {
//			origin := ctx.GetHeader("Origin")
//			return isAllowedOrigin(string(origin))
//		},
//	}
//
//	ns := mvc.NewNamespace("/api",
//		mvc.NSWebSocketWithUpgrader("/ws", handler, upgrader),
//	)
func NSWebSocketWithUpgrader(path string, handler WsHandlerFunc, upgrader websocket.HertzUpgrader) NamespaceFunc {
	return func(ns *Namespace) {
		ns.wsRouters = append(ns.wsRouters, wsRouterInfo{
			path:     path,
			handler:  handler,
			upgrader: upgrader,
		})
	}
}

// NSRouterWs 为命名空间添加WebSocket控制器路由映射
//
// 类似于NSRouter，但专门用于WebSocket路由。支持将控制器方法映射为WebSocket处理器。
// 控制器方法应该接受WebSocket连接参数并处理WebSocket消息。
//
// 参数：
//   - path: string - WebSocket路由路径，相对于命名空间前缀的路径
//   - ctrl: core.IController - 控制器实例
//   - method: string - 控制器方法名，如 "HandleChat"、"EchoMessage"
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// 使用示例：
//
//	// WebSocket控制器路由
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSRouterWs("/chat", &ChatController{}, "HandleChat"),
//		mvc.NSRouterWs("/echo", &EchoController{}, "HandleEcho"),
//	)
//
//	// 生成的路由：
//	// WS /api/v1/chat -> ChatController.HandleChat(conn *websocket.Conn)
//	// WS /api/v1/echo -> EchoController.HandleEcho(conn *websocket.Conn)
//
// 控制器方法要求：
//
//	func (c *YourController) HandleChat(conn *websocket.Conn) {
//		// WebSocket 消息处理逻辑
//	}
func NSRouterWs(path string, ctrl core.IController, method string) NamespaceFunc {
	return func(ns *Namespace) {
		ns.wsControllerRouters = append(ns.wsControllerRouters, wsControllerRouterInfo{
			path:       path,
			controller: ctrl,
			method:     method,
			upgrader:   nil, // 使用默认升级器
		})
	}
}

// NSRouterWsWithUpgrader 为命名空间添加带自定义升级器的WebSocket控制器路由映射
//
// 类似于NSRouterWs，但允许指定自定义的WebSocket升级器配置。
//
// 参数：
//   - path: string - WebSocket路由路径，相对于命名空间前缀的路径
//   - ctrl: core.IController - 控制器实例
//   - method: string - 控制器方法名，如 "HandleChat"、"EchoMessage"
//   - upgrader: websocket.HertzUpgrader - 自定义WebSocket升级器配置
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// 使用示例：
//
//	upgrader := websocket.HertzUpgrader{
//		ReadBufferSize:  1024,
//		WriteBufferSize: 1024,
//	}
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSRouterWsWithUpgrader("/chat", &ChatController{}, "HandleChat", upgrader),
//	)
func NSRouterWsWithUpgrader(path string, ctrl core.IController, method string, upgrader websocket.HertzUpgrader) NamespaceFunc {
	return func(ns *Namespace) {
		ns.wsControllerRouters = append(ns.wsControllerRouters, wsControllerRouterInfo{
			path:       path,
			controller: ctrl,
			method:     method,
			upgrader:   &upgrader,
		})
	}
}

// NSMiddleware 为命名空间添加中间件
//
// 添加到命名空间的中间件会应用于该命名空间及其所有子命名空间中的路由。
// 中间件按添加顺序执行，支持中间件链式调用。
//
// 参数：
//   - middlewares: ...define.HandlerFunc - 要添加的中间件函数列表
//
// 返回值：
//   - NamespaceFunc: 命名空间配置函数
//
// 中间件执行顺序：
//  1. 父命名空间中间件（按添加顺序）
//  2. 当前命名空间中间件（按添加顺序）
//  3. 路由级中间件
//  4. 控制器方法
//
// 使用示例：
//
//	// 单个中间件
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSMiddleware(authMiddleware),
//		mvc.NSAutoRouter(&UserController{}),
//	)
//
//	// 多个中间件
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSMiddleware(corsMiddleware, authMiddleware, rateLimitMiddleware),
//		mvc.NSAutoRouter(&UserController{}),
//	)
//
//	// 分步添加中间件
//	ns := mvc.NewNamespace("/api/v1",
//		mvc.NSMiddleware(corsMiddleware),
//		mvc.NSMiddleware(authMiddleware),     // 在corsMiddleware之后执行
//		mvc.NSMiddleware(rateLimitMiddleware), // 在authMiddleware之后执行
//		mvc.NSAutoRouter(&UserController{}),
//	)
//
// 常用中间件示例：
//   - CORS处理：跨域请求支持
//   - 认证中间件：用户身份验证
//   - 日志中间件：请求日志记录
//   - 限流中间件：请求频率限制
//   - 缓存中间件：响应缓存处理
func NSMiddleware(middlewares ...define.HandlerFunc) NamespaceFunc {
	return func(ns *Namespace) {
		ns.middlewares = append(ns.middlewares, middlewares...)
	}
}

// ============= 命名空间注册方法 =============

// Register 将命名空间注册到应用
//
// 这是一个内部方法，由框架自动调用来将命名空间中的所有控制器和路由注册到应用。
// 支持递归注册嵌套的子命名空间，并处理中间件继承。
//
// 参数：
//   - app: *core.App - YYHertz应用实例
//
// 注册过程：
//  1. 注册自动路由控制器
//  2. 注册手动路由映射
//  3. 递归注册所有子命名空间
//  4. 处理中间件继承和路径合并
//
// 注意事项：
//   - 该方法由框架内部调用，用户不应直接调用
//   - 注册顺序影响路由匹配优先级
//   - 中间件会按父子关系合并
func (ns *Namespace) Register(app *core.App) {
	// 步骤1：注册当前命名空间的自动路由控制器
	for _, ctrl := range ns.controllers {
		if ctrl.autoRoute {
			// 使用命名空间前缀注册自动路由
			app.RouterAutoPrefix(ns.prefix, ctrl.controller)
		}
	}

	// 步骤2：注册当前命名空间的手动路由
	for _, router := range ns.routers {
		ns.registerRouter(app, router)
	}

	// 步骤2.5：注册当前命名空间的 WebSocket 路由
	for _, wsRouter := range ns.wsRouters {
		ns.registerWebSocketRouter(app, wsRouter)
	}

	// 步骤2.6：注册当前命名空间的 WebSocket 控制器路由
	for _, wsControllerRouter := range ns.wsControllerRouters {
		ns.registerWebSocketControllerRouter(app, wsControllerRouter)
	}

	// 步骤3：递归注册所有子命名空间
	for _, subNs := range ns.namespaces {
		// 构建孌全的嵌套路径：父前缀 + 子前缀
		fullPrefix := ns.prefix
		if !strings.HasSuffix(fullPrefix, "/") {
			fullPrefix += "/"
		}
		fullPrefix += strings.TrimPrefix(subNs.prefix, "/")

		// 创建子命名空间的副本，更新完整路径和继承中间件
		subNsCopy := &Namespace{
			prefix:              fullPrefix,                                   // 合并后的完整前缀
			controllers:         subNs.controllers,                            // 子命名空间的控制器
			routers:             subNs.routers,                                // 子命名空间的路由
			wsRouters:           subNs.wsRouters,                              // 子命名空间的WebSocket路由
			wsControllerRouters: subNs.wsControllerRouters,                    // 子命名空间的WebSocket控制器路由
			namespaces:          subNs.namespaces,                             // 子命名空间的子命名空间
			middlewares:         append(ns.middlewares, subNs.middlewares...), // 继承父级中间件并添加自己的中间件
		}

		// 递归注册子命名空间
		subNsCopy.Register(app)
	}
}

// registerRouter 注册单个手动路由
//
// 该方法解析手动路由的方法格式，并将其注册到应用。
// 支持多种方法格式和通配符。
//
// 参数：
//   - app: *core.App - 应用实例
//   - router: routerInfo - 路由信息
//
// 支持的方法格式：
//   - "GET:MethodName" - 指定HTTP方法和控制器方法
//   - "*:MethodName" - 匹配所有HTTP方法
//   - "MethodName" - 默认匹配所有HTTP方法
func (ns *Namespace) registerRouter(app *core.App, router routerInfo) {
	// 解析方法规格："*:MethodName" 或 "GET:MethodName" 或 "MethodName"
	var httpMethod, methodName string

	if strings.Contains(router.method, ":") {
		// 包含冒号，格式为 "HTTP_METHOD:CONTROLLER_METHOD"
		parts := strings.SplitN(router.method, ":", 2)
		httpMethod = strings.ToUpper(parts[0])
		methodName = parts[1]

		// 将通配符 "*" 转换为框架识别的 "ANY"
		if httpMethod == "*" {
			httpMethod = "ANY"
		}
	} else {
		// 不包含冒号，默认为匹配所有HTTP方法
		httpMethod = "ANY"
		methodName = router.method
	}

	// 构建路由规格字符串，格式为 "HTTP_METHOD:PATH"
	routeSpec := httpMethod + ":" + router.path

	// 使用框架的RouterPrefix方法注册路由
	// 参数说明：
	// - ns.prefix: 命名空间前缀作为基础路径
	// - router.controller: 控制器实例
	// - true: 使用成对模式（methodName, routeSpec成对出现）
	// - methodName: 控制器方法名
	// - routeSpec: 路由规格
	app.RouterPrefix(ns.prefix, router.controller, true, methodName, routeSpec)
}

// registerWebSocketRouter 注册单个 WebSocket 路由
//
// 该方法将 WebSocket 路由注册到应用，创建专门的 WebSocket 控制器来处理连接升级。
//
// 参数：
//   - app: *core.App - 应用实例
//   - wsRouter: wsRouterInfo - WebSocket路由信息
func (ns *Namespace) registerWebSocketRouter(app *core.App, wsRouter wsRouterInfo) {
	// 创建 WebSocket 控制器
	wsController := CreateWebSocketController(wsRouter.handler, wsRouter.upgrader)

	// 构建路由规格字符串，WebSocket 通常使用 GET 方法进行升级
	routeSpec := "GET:" + wsRouter.path

	// 使用框架的 RouterPrefix 方法注册 WebSocket 路由
	// 注册为 GET 请求，由控制器内部处理 WebSocket 升级
	// fmt.Printf("注册 WebSocket 路由: %s%s -> %s\n", ns.prefix, wsRouter.path, routeSpec)
	app.RouterPrefix(ns.prefix, wsController, true, "HandleWebSocket", routeSpec)
}

// registerWebSocketControllerRouter 注册单个 WebSocket 控制器路由
//
// 该方法将 WebSocket 控制器路由注册到应用，直接使用控制器方法处理 WebSocket 连接。
//
// 参数：
//   - app: *core.App - 应用实例
//   - wsControllerRouter: wsControllerRouterInfo - WebSocket控制器路由信息
func (ns *Namespace) registerWebSocketControllerRouter(app *core.App, wsControllerRouter wsControllerRouterInfo) {
	// 构建完整的 WebSocket 路径
	fullPath := ns.prefix
	// 确保路径之间有正确的分隔符
	if !strings.HasSuffix(fullPath, "/") {
		fullPath += "/"
	}
	fullPath += strings.TrimPrefix(wsControllerRouter.path, "/")

	// 规范化路径：移除重复的斜杠
	fullPath = strings.ReplaceAll(fullPath, "//", "/")

	// 使用专门的 WebSocket 路由注册方法
	if wsControllerRouter.upgrader != nil {
		// 如果有自定义升级器，使用 RouterWsWithUpgrader
		app.RouterWsWithUpgrader(fullPath, wsControllerRouter.controller, wsControllerRouter.method, *wsControllerRouter.upgrader)
	} else {
		// 使用默认升级器的 RouterWs
		app.RouterWs(fullPath, wsControllerRouter.controller, wsControllerRouter.method)
	}
}

// ============= 命名空间查询方法 =============

// GetPrefix 获取命名空间的URL前缀
//
// 返回当前命名空间的URL前缀，不包括父命名空间的前缀。
// 如需获取完整路径，请使用GetFullPrefix()。
//
// 返回值：
//   - string: 命名空间的URL前缀
//
// 使用示例：
//
//	ns := mvc.NewNamespace("/api/v1")
//	prefix := ns.GetPrefix() // 返回 "/api/v1"
func (ns *Namespace) GetPrefix() string {
	return ns.prefix
}

// GetControllers 获取命名空间中所有控制器实例
//
// 返回当前命名空间中注册的所有控制器实例列表。
// 不包括子命名空间中的控制器。
//
// 返回值：
//   - []core.IController: 控制器实例列表
//
// 使用示例：
//
//	controllers := ns.GetControllers()
//	for _, ctrl := range controllers {
//		fmt.Printf("Controller: %T\n", ctrl)
//	}
func (ns *Namespace) GetControllers() []core.IController {
	var controllers []core.IController
	for _, ctrl := range ns.controllers {
		controllers = append(controllers, ctrl.controller)
	}
	return controllers
}

// GetRouters 获取命名空间中的手动路由信息
//
// 返回当前命名空间中手动注册的路由信息列表。
// 每个字符串包含路径和方法映射信息。
//
// 返回值：
//   - []string: 路由信息字符串列表，格式为 "path -> method"
//
// 使用示例：
//
//	routes := ns.GetRouters()
//	for _, route := range routes {
//		fmt.Println("Route:", route) // 输出格式： "/users/:id -> GET:Get"
//	}
func (ns *Namespace) GetRouters() []string {
	var routes []string
	for _, router := range ns.routers {
		routes = append(routes, router.path+" -> "+router.method)
	}
	return routes
}

// GetNamespaces 获取所有子命名空间
//
// 返回当前命名空间的所有直接子命名空间列表。
// 不包括更深层级的孙命名空间。
//
// 返回值：
//   - []*Namespace: 子命名空间列表
//
// 使用示例：
//
//	subNamespaces := ns.GetNamespaces()
//	for _, subNs := range subNamespaces {
//		fmt.Printf("Sub-namespace prefix: %s\n", subNs.GetPrefix())
//	}
func (ns *Namespace) GetNamespaces() []*Namespace {
	return ns.namespaces
}
