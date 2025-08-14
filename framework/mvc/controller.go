// Package mvc 提供MVC模式的控制器和路由管理功能
//
// 本包是YYHertz框架的MVC层核心，提供：
// 1. BaseController - 基础控制器，包含请求处理、响应、会话等功能
// 2. 自动路由注册 - 根据控制器方法名自动生成路由
// 3. 手动路由注册 - 精确控制路由映射
// 4. 命名空间支持 - 类似Beego的命名空间功能
//
// 设计原则：
// - 保持向后兼容性
// - 简化路由配置
// - 支持RESTful风格
// - 高性能路由匹配
//
// 基本用法示例：
//
//	// 1. 定义控制器
//	type UserController struct {
//		mvc.BaseController
//	}
//
//	func (c *UserController) Get() {
//		c.Ctx.JSON(200, map[string]string{"message": "Hello"})
//	}
//
//	// 2. 注册路由
//	mvc.RouterAuto(&UserController{})  // 自动路由: GET /user
//
//	// 3. 手动路由
//	mvc.Router(&UserController{}, false, "GET:/api/users|Get")
//
// 高级用法：
//
//	// 批量注册
//	mvc.RouterAuto(&UserController{}, &OrderController{})
//
//	// 带前缀注册
//	mvc.RouterAutoPrefix("/api/v1", &UserController{})
//
//	// 命名空间
//	ns := mvc.NewNamespace("/api",
//		mvc.NSRouter("/users", &UserController{}),
//	)
//	mvc.AddNamespace(ns)
package mvc

// 重新导出BaseController，保持向后兼容
import (
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// BaseController MVC基础控制器类型别名
//
// BaseController是所有控制器的基类，提供了完整的Web开发功能：
//
// 核心功能：
//   - HTTP请求处理 (GET, POST, PUT, DELETE等)
//   - 响应管理 (JSON, HTML, File等)
//   - 会话管理 (Session)
//   - 参数获取和验证
//   - 国际化支持
//   - 错误处理
//
// 使用方式：
//
//	type MyController struct {
//		mvc.BaseController
//	}
//
//	func (c *MyController) Get() {
//		// 处理GET请求
//		name := c.GetString("name", "World")
//		c.Ctx.JSON(200, map[string]string{
//			"message": "Hello " + name,
//		})
//	}
//
// 支持的HTTP方法：
//   - Get() - 处理GET请求
//   - Post() - 处理POST请求
//   - Put() - 处理PUT请求
//   - Delete() - 处理DELETE请求
//   - Head() - 处理HEAD请求
//   - Options() - 处理OPTIONS请求
//   - Patch() - 处理PATCH请求
//
// 更多功能请参考 core.BaseController 文档
type BaseController = core.BaseController

// NewBaseController 创建新的BaseController实例
//
// 这是BaseController的标准构造函数，会初始化控制器的所有功能。
//
// 返回值：
//   - *BaseController: 新创建的控制器实例
//
// 使用示例：
//
//	controller := mvc.NewBaseController()
//	// controller 现在可以使用所有BaseController功能
//
// 注意事项：
//   - 通常不需要手动调用此函数
//   - 框架会在路由匹配时自动创建控制器实例
//   - 如需手动创建，请确保正确设置Context
var NewBaseController = core.NewBaseController

// NewBaseControllerWithContext 使用指定Context创建BaseController实例
//
// 这个构造函数允许你传入一个预先配置的Context来创建控制器。
// 主要用于测试或特殊场景下的控制器实例化。
//
// 参数：
//   - ctx: context.Context - 用于创建控制器的上下文
//
// 返回值：
//   - *BaseController: 带有指定Context的控制器实例
//
// 使用示例：
//
//	// 测试场景
//	testCtx := context.NewContext(mockRequestContext)
//	controller := mvc.NewBaseControllerWithContext(testCtx)
//
//	// 现在可以测试控制器方法
//	controller.Get()
//
// 适用场景：
//   - 单元测试
//   - 自定义Context配置
//   - 特殊的控制器初始化需求
var NewBaseControllerWithContext = core.NewBaseControllerWithContext

// ============= 自动路由注册函数 =============

// RouterAuto 批量自动注册多个控制器路由
//
// 此函数会根据控制器的类型名和方法名自动生成RESTful风格的路由。
// 路由生成规则：
//   - 控制器名: UserController -> /user
//   - HTTP方法: Get() -> GET, Post() -> POST, Put() -> PUT, Delete() -> DELETE
//   - 自定义方法: GetProfile() -> GET /user/profile
//
// 参数：
//   - controllers: ...IController - 要注册的控制器实例列表
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 使用示例：
//
//	// 定义控制器
//	type UserController struct {
//		mvc.BaseController
//	}
//
//	func (c *UserController) Get() {
//		// 处理 GET /user
//	}
//
//	func (c *UserController) Post() {
//		// 处理 POST /user
//	}
//
//	func (c *UserController) GetProfile() {
//		// 处理 GET /user/profile
//	}
//
//	type OrderController struct {
//		mvc.BaseController
//	}
//
//	func (c *OrderController) Get() {
//		// 处理 GET /order
//	}
//
//	// 批量注册
//	mvc.RouterAuto(&UserController{}, &OrderController{})
//
// 生成的路由：
//   - GET /user -> UserController.Get()
//   - POST /user -> UserController.Post()
//   - GET /user/profile -> UserController.GetProfile()
//   - GET /order -> OrderController.Get()
//
// 注意事项：
//   - 控制器名会自动转换为小写并移除Controller后缀
//   - 方法名会根据HTTP动词前缀自动路由
//   - 不支持路径参数，如需要请使用手动路由
func RouterAuto(controllers ...IController) *App {
	return HertzApp.RouterAuto(controllers...)
}

// RouterAutoPrefix 批量自动注册控制器路由，使用指定路径前缀
//
// 与RouterAuto类似，但为所有生成的路由添加统一前缀。
// 常用于API版本控制或模块分组。
//
// 参数：
//   - prefix: string - 路由前缀，如 "/api/v1"
//   - ctrls: ...IController - 要注册的控制器实例列表
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 使用示例：
//
//	type UserController struct {
//		mvc.BaseController
//	}
//
//	func (c *UserController) Get() {
//		// 处理请求
//	}
//
//	func (c *UserController) GetProfile() {
//		// 处理个人资料请求
//	}
//
//	// 带前缀注册
//	mvc.RouterAutoPrefix("/api/v1", &UserController{})
//
// 生成的路由：
//   - GET /api/v1/user -> UserController.Get()
//   - GET /api/v1/user/profile -> UserController.GetProfile()
//
// 前缀规则：
//   - 前缀会自动添加到所有生成的路由前
//   - 前缀应以"/"开头，框架会自动处理重复的"/"
//   - 支持多级路径前缀，如 "/api/v1/admin"
//
// 最佳实践：
//   - 用于API版本管理: "/api/v1", "/api/v2"
//   - 用于模块分组: "/admin", "/public"
//   - 保持前缀简洁，避免过深的嵌套
func RouterAutoPrefix(prefix string, ctrls ...IController) *App {
	return HertzApp.RouterAutoPrefix(prefix, ctrls...)
}

// ============= 手动路由注册函数 =============

// Router 手动注册控制器路由
//
// 允许精确控制路由映射，支持路径参数、复杂路由规则等高级功能。
// 当自动路由无法满足需求时，使用手动路由来获得完全的控制权。
//
// 参数：
//   - ctrl: IController - 控制器实例
//   - routePair: bool - 路由配对模式标识
//   - true: 使用成对模式，路由规则按 "方法名", "HTTP_METHOD:PATH" 成对出现
//   - false: 使用单一模式，路由规则格式为 "HTTP_METHOD:PATH|MethodName"
//   - routes: ...string - 路由规则列表
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 两种路由格式：
//
// 1. 成对模式 (routePair = true)：
//   - "MethodName", "GET:/users/:id" - 方法名和路由成对
//   - "Get", "GET:/users/:id" - Get方法处理GET /users/:id
//   - "PostCreate", "POST:/users" - PostCreate方法处理POST /users
//
// 2. 单一模式 (routePair = false)：
//   - "GET:/users/:id|Get" - HTTP方法:路径|方法名
//   - "POST:/users|PostCreate" - 指定方法名映射
//   - "GET:/users" - 默认映射到Get()方法
//
// 使用示例：
//
//	type UserController struct {
//		mvc.BaseController
//	}
//
//	func (c *UserController) Get() {
//		// 处理获取用户信息
//		id := c.Ctx.Param("id")
//		c.Ctx.JSON(200, map[string]string{"user_id": id})
//	}
//
//	func (c *UserController) PostCreate() {
//		// 处理创建用户
//		name := c.GetString("name")
//		c.Ctx.JSON(200, map[string]string{"created": name})
//	}
//
//	func (c *UserController) GetSearch() {
//		// 处理搜索用户
//		keyword := c.GetString("q")
//		c.Ctx.JSON(200, map[string]string{"search": keyword})
//	}
//
//	// 方式1: 成对模式 (推荐)
//	mvc.Router(&UserController{}, true,
//		"Get", "GET:/users/:id",           // 方法名, 路由
//		"PostCreate", "POST:/users",       // 方法名, 路由
//		"GetSearch", "GET:/users/search",  // 方法名, 路由
//	)
//
//	// 方式2: 单一模式
//	mvc.Router(&UserController{}, false,
//		"GET:/users/:id|Get",              // 路由|方法名
//		"POST:/users|PostCreate",          // 路由|方法名
//		"GET:/users/search|GetSearch",     // 路由|方法名
//	)
//
// 生成的路由：
//   - GET /users/:id -> UserController.Get()
//   - POST /users -> UserController.PostCreate()
//   - GET /users/search -> UserController.GetSearch()
//
// 高级特性：
//   - 支持路径参数: :id, :name, :category
//   - 支持通配符: *path, *filepath
//   - 支持正则表达式约束
//   - 灵活的方法名映射
//
// 选择建议：
//   - 成对模式 (true): 代码清晰，IDE友好，推荐使用
//   - 单一模式 (false): 配置紧凑，适合简单映射
//
// 适用场景：
//   - 需要路径参数的API
//   - 复杂的URL结构
//   - RESTful API设计
//   - 与前端路由对接
func Router(ctrl IController, routePair bool, routes ...string) *App {
	return HertzApp.Router(ctrl, routePair, routes...)
}

// RouterPrefix 手动注册控制器路由，使用指定前缀
//
// 在Router的基础上添加路径前缀，
// 适用于需要前缀且要精确控制路由的场景。
//
// 参数：
//   - prefix: string - 路由前缀，如 "/api/v1"
//   - ctrl: IController - 控制器实例
//   - routePair: bool - 路由配对模式标识 (同Router)
//   - routes: ...string - 路由规则列表
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 使用示例：
//
//	type UserController struct {
//		mvc.BaseController
//	}
//
//	func (c *UserController) Get() {
//		id := c.Ctx.Param("id")
//		// 获取用户逻辑...
//		c.Ctx.JSON(200, map[string]interface{}{
//			"id":   id,
//			"name": "John Doe",
//		})
//	}
//
//	func (c *UserController) PostCreate() {
//		// 创建用户逻辑...
//		c.Ctx.JSON(201, map[string]string{"status": "created"})
//	}
//
//	func (c *UserController) GetFollowers() {
//		id := c.Ctx.Param("id")
//		// 获取关注者逻辑...
//		c.Ctx.JSON(200, []string{"follower1", "follower2"})
//	}
//
//	// 方式1: 带前缀的单一模式
//	mvc.RouterPrefix("/api/v1", &UserController{}, false,
//		"GET:/users/:id|Get",                       // -> /api/v1/users/:id
//		"POST:/users|PostCreate",                   // -> /api/v1/users
//		"GET:/users/:id/followers|GetFollowers",    // -> /api/v1/users/:id/followers
//	)
//
//	// 方式2: 带前缀的成对模式
//	mvc.RouterPrefix("/api/v1", &UserController{}, true,
//		"Get", "GET:/users/:id",                    // -> /api/v1/users/:id
//		"PostCreate", "POST:/users",                // -> /api/v1/users
//		"GetFollowers", "GET:/users/:id/followers", // -> /api/v1/users/:id/followers
//	)
//
// 生成的路由：
//   - GET /api/v1/users/:id -> UserController.Get()
//   - POST /api/v1/users -> UserController.PostCreate()
//   - GET /api/v1/users/:id/followers -> UserController.GetFollowers()
//
// 使用场景：
//   - API版本化管理
//   - 模块化路由组织
//   - 复杂应用的路由分层
//   - 微服务路由前缀
//
// 最佳实践：
//   - 保持前缀简洁
//   - 使用语义化的前缀
//   - 避免过深的嵌套层级
func RouterPrefix(prefix string, ctrl IController, routePair bool, routes ...string) *App {
	return HertzApp.RouterPrefix(prefix, ctrl, routePair, routes...)
}

// RouterMap 手动注册控制器路由（使用map格式）
//
// 提供一种更简洁的路由配置方式，使用map[string]string格式定义路由规则。
// 相比Router函数的字符串数组方式，RouterMap更直观易读。
//
// 参数：
//   - ctrl: IController - 控制器实例
//   - routes: map[string]string - 路由规则映射
//     • key: HTTP方法:路径格式，如 "GET:/users/:id"
//     • value: 控制器方法名，如 "Get"
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 路由格式说明：
//   - key格式: "HTTP_METHOD:PATH" 
//   - value格式: "MethodName"
//   - 支持路径参数: ":id", ":name" 等
//   - 支持通配符: "*path", "*file" 等
//
// 使用示例：
//
//	type UserController struct {
//		mvc.BaseController
//	}
//	
//	func (c *UserController) GetList() {
//		// 获取用户列表
//		users := []map[string]interface{}{
//			{"id": 1, "name": "Alice"},
//			{"id": 2, "name": "Bob"},
//		}
//		c.Ctx.JSON(200, users)
//	}
//	
//	func (c *UserController) Get() {
//		// 获取单个用户
//		id := c.Ctx.Param("id")
//		c.Ctx.JSON(200, map[string]interface{}{
//			"id": id, "name": "User " + id,
//		})
//	}
//	
//	func (c *UserController) PostCreate() {
//		// 创建用户
//		name := c.GetString("name")
//		c.Ctx.JSON(201, map[string]interface{}{
//			"id": 123, "name": name, "status": "created",
//		})
//	}
//	
//	func (c *UserController) PutUpdate() {
//		// 更新用户
//		id := c.Ctx.Param("id")
//		name := c.GetString("name")
//		c.Ctx.JSON(200, map[string]interface{}{
//			"id": id, "name": name, "status": "updated",
//		})
//	}
//	
//	func (c *UserController) DeleteRemove() {
//		// 删除用户
//		id := c.Ctx.Param("id")
//		c.Ctx.JSON(200, map[string]interface{}{
//			"id": id, "status": "deleted",
//		})
//	}
//	
//	// 使用Map格式注册路由
//	mvc.RouterMap(&UserController{}, map[string]string{
//		"GET:/users":     "GetList",     // 获取用户列表
//		"GET:/users/:id": "Get",         // 获取单个用户
//		"POST:/users":    "PostCreate",  // 创建用户
//		"PUT:/users/:id": "PutUpdate",   // 更新用户
//		"DELETE:/users/:id": "DeleteRemove", // 删除用户
//	})
//
// 生成的路由：
//   - GET /users -> UserController.GetList()
//   - GET /users/:id -> UserController.Get()
//   - POST /users -> UserController.PostCreate()
//   - PUT /users/:id -> UserController.PutUpdate()
//   - DELETE /users/:id -> UserController.DeleteRemove()
//
// 与其他方式对比：
//   - 相比Router: 配置更直观，键值对映射清晰
//   - 相比RouterAuto: 支持路径参数和自定义路径
//   - 适合中等复杂度的路由配置
//
// 适用场景：
//   - RESTful API设计
//   - 需要路径参数的场景
//   - 希望配置简洁易读的项目
//   - 路由数量适中的应用
//
// 最佳实践：
//   - 保持路由规则简洁明了
//   - 使用语义化的方法名
//   - 遵循RESTful命名约定
//   - 适当使用路径参数而非查询参数
func RouterMap(ctrl IController, routes map[string]string) *App {
	return HertzApp.RouterMap(ctrl, routes)
}

// RouterPrefixMap 手动注册控制器路由（使用map格式，带前缀）
//
// 在RouterMap的基础上添加路径前缀，结合了Map格式的简洁性和前缀的组织性。
// 非常适合API版本管理和模块化路由组织。
//
// 参数：
//   - prefix: string - 路由前缀，如 "/api/v1", "/admin"
//   - ctrl: IController - 控制器实例  
//   - routes: map[string]string - 路由规则映射
//
// 返回值：
//   - *App: 应用实例，支持链式调用
//
// 使用示例：
//
//	type UserController struct {
//		mvc.BaseController
//	}
//	
//	func (c *UserController) GetProfile() {
//		// 获取用户个人资料
//		id := c.Ctx.Param("id")
//		c.Ctx.JSON(200, map[string]interface{}{
//			"id": id,
//			"name": "John Doe",
//			"email": "john@example.com",
//			"profile": map[string]interface{}{
//				"bio": "Software Developer",
//				"location": "San Francisco",
//			},
//		})
//	}
//	
//	func (c *UserController) PutProfile() {
//		// 更新用户个人资料
//		id := c.Ctx.Param("id")
//		name := c.GetString("name")
//		email := c.GetString("email")
//		
//		c.Ctx.JSON(200, map[string]interface{}{
//			"id": id,
//			"name": name,
//			"email": email,
//			"status": "profile updated",
//		})
//	}
//	
//	func (c *UserController) GetFollowers() {
//		// 获取用户关注者列表
//		id := c.Ctx.Param("id")
//		page := c.GetString("page", "1")
//		
//		c.Ctx.JSON(200, map[string]interface{}{
//			"user_id": id,
//			"page": page,
//			"followers": []map[string]interface{}{
//				{"id": 2, "name": "Alice", "avatar": "avatar1.jpg"},
//				{"id": 3, "name": "Bob", "avatar": "avatar2.jpg"},
//			},
//			"total": 234,
//		})
//	}
//	
//	func (c *UserController) PostFollow() {
//		// 关注用户
//		id := c.Ctx.Param("id")
//		followerId := c.GetString("follower_id")
//		
//		c.Ctx.JSON(200, map[string]interface{}{
//			"user_id": id,
//			"follower_id": followerId,
//			"status": "followed",
//		})
//	}
//	
//	// API v1 用户相关路由
//	mvc.RouterPrefixMap("/api/v1", &UserController{}, map[string]string{
//		"GET:/users/:id/profile":    "GetProfile",    // 获取个人资料
//		"PUT:/users/:id/profile":    "PutProfile",    // 更新个人资料  
//		"GET:/users/:id/followers":  "GetFollowers",  // 获取关注者
//		"POST:/users/:id/follow":    "PostFollow",    // 关注用户
//	})
//	
//	// API v2 版本可以使用相同控制器但不同路径
//	mvc.RouterPrefixMap("/api/v2", &UserController{}, map[string]string{
//		"GET:/user/:id":         "GetProfile",    // v2简化路径
//		"PATCH:/user/:id":       "PutProfile",    // v2使用PATCH
//		"GET:/user/:id/fans":    "GetFollowers",  // v2改名fans
//	})
//
// 生成的路由：
//   - GET /api/v1/users/:id/profile -> UserController.GetProfile()
//   - PUT /api/v1/users/:id/profile -> UserController.PutProfile()
//   - GET /api/v1/users/:id/followers -> UserController.GetFollowers()
//   - POST /api/v1/users/:id/follow -> UserController.PostFollow()
//   - GET /api/v2/user/:id -> UserController.GetProfile()
//   - PATCH /api/v2/user/:id -> UserController.PutProfile()
//   - GET /api/v2/user/:id/fans -> UserController.GetFollowers()
//
// 高级用法示例：
//
//	// 管理员模块
//	type AdminController struct {
//		mvc.BaseController
//	}
//	
//	func (c *AdminController) GetDashboard() {
//		c.Ctx.JSON(200, map[string]interface{}{
//			"stats": map[string]int{
//				"users": 1234, "orders": 5678, "revenue": 99999,
//			},
//		})
//	}
//	
//	func (c *AdminController) GetUsers() {
//		page := c.GetString("page", "1")
//		c.Ctx.JSON(200, map[string]interface{}{
//			"page": page, "users": []string{"user1", "user2"},
//		})
//	}
//	
//	// 管理后台路由
//	mvc.RouterPrefixMap("/admin", &AdminController{}, map[string]string{
//		"GET:/dashboard": "GetDashboard", // 仪表盘
//		"GET:/users":     "GetUsers",     // 用户管理
//	})
//
// 与其他方式对比：
//   - 相比RouterMap: 增加了前缀支持，便于版本和模块管理
//   - 相比RouterPrefix: Map格式更直观，配置更简洁
//   - 结合了前缀组织性和Map格式的可读性
//
// 适用场景：
//   - API版本管理 (/api/v1, /api/v2)
//   - 模块化路由组织 (/admin, /public, /mobile)
//   - 微服务路由前缀
//   - 多租户应用 (/tenant1, /tenant2)
//   - 需要路径参数的复杂API
//
// 最佳实践：
//   - 前缀保持简洁语义化
//   - 合理使用版本号前缀
//   - 避免过深的路径嵌套
//   - 保持同一前缀下的路由风格统一
//   - 使用RESTful设计原则
func RouterPrefixMap(prefix string, ctrl IController, routes map[string]string) *App {
	return HertzApp.RouterPrefixMap(prefix, ctrl, routes)
}

// ============= 命名空间支持 =============

// AddNamespace 添加命名空间到全局应用
//
// 命名空间提供了一种组织和管理路由的高级方式，
// 类似于Beego的命名空间功能，支持嵌套、中间件、过滤器等。
//
// 参数：
//   - ns: ...*Namespace - 要添加的命名空间列表
//
// 命名空间的优势：
//   - 路由组织: 将相关路由分组管理
//   - 中间件作用域: 为特定路由组应用中间件
//   - 嵌套结构: 支持多级命名空间嵌套
//   - 版本管理: 便于API版本控制
//
// 使用示例：
//
//	// 定义控制器
//	type UserController struct {
//		mvc.BaseController
//	}
//
//	func (c *UserController) Get() {
//		c.Ctx.JSON(200, map[string]string{"action": "get_user"})
//	}
//
//	type AdminController struct {
//		mvc.BaseController
//	}
//
//	func (c *AdminController) Get() {
//		c.Ctx.JSON(200, map[string]string{"action": "admin_panel"})
//	}
//
//	// 创建API v1命名空间
//	apiV1 := mvc.NewNamespace("/api/v1",
//		// 添加中间件
//		mvc.NSBefore(authMiddleware),
//
//		// 添加路由
//		mvc.NSRouter("/users", &UserController{}),
//
//		// 嵌套命名空间
//		mvc.NSNamespace("/admin",
//			mvc.NSBefore(adminMiddleware),
//			mvc.NSRouter("/", &AdminController{}),
//		),
//	)
//
//	// 创建API v2命名空间
//	apiV2 := mvc.NewNamespace("/api/v2",
//		mvc.NSRouter("/users", &UserV2Controller{}),
//	)
//
//	// 添加到应用
//	mvc.AddNamespace(apiV1, apiV2)
//
// 生成的路由结构：
//   - GET /api/v1/users -> UserController.Get() (with authMiddleware)
//   - GET /api/v1/admin/ -> AdminController.Get() (with authMiddleware + adminMiddleware)
//   - GET /api/v2/users -> UserV2Controller.Get()
//
// 高级特性：
//   - 支持中间件链
//   - 支持条件路由
//   - 支持过滤器
//   - 支持嵌套命名空间
//
// 适用场景：
//   - 大型应用的路由组织
//   - API版本管理
//   - 权限分级控制
//   - 微服务架构
//   - 模块化开发
//
// 注意事项：
//   - 命名空间注册顺序影响路由匹配优先级
//   - 中间件会按照命名空间层级顺序执行
//   - 避免命名空间路径冲突
func AddNamespace(ns ...*Namespace) {
	if HertzApp != nil {
		for _, n := range ns {
			n.Register(HertzApp)
		}
	}
}

// ============= 性能和最佳实践指南 =============

/*
路由注册方式性能对比：

1. 自动路由 (RouterAuto系列)
   - 优点: 开发效率高，代码简洁，自动RESTful风格
   - 缺点: 路由灵活性较低，不支持路径参数
   - 性能: 中等 (需要反射解析控制器名和方法名)
   - 适用: 快速原型开发，简单CRUD操作

2. 手动路由 (Router系列)
   - 优点: 完全控制路由规则，支持路径参数和复杂路径
   - 缺点: 需要手动维护路由配置
   - 性能: 最高 (直接路由映射，无反射开销)
   - 适用: 生产环境，复杂API设计

3. 命名空间 (Namespace)
   - 优点: 最佳的组织性，支持中间件作用域，适合大型应用
   - 缺点: 配置复杂度较高
   - 性能: 高 (优化的路由树结构)
   - 适用: 大型应用，API版本管理，模块化架构

最佳实践建议：

1. 项目初期
   - 使用RouterAuto快速搭建原型
   - 保持控制器结构简单清晰
   - 遵循RESTful设计原则

2. 项目成熟期
   - 逐步迁移到Router获得更好的性能
   - 使用Namespace组织复杂的路由结构
   - 添加适当的中间件和错误处理

3. 生产环境
   - 优先使用Router和Namespace
   - 避免在热路径上使用反射
   - 合理设计URL结构，便于缓存和CDN

4. 错误处理
   - 在控制器中添加统一的错误处理
   - 使用中间件进行全局错误捕获
   - 实现优雅的错误响应格式

5. 安全性
   - 在命名空间级别添加认证中间件
   - 对敏感操作进行权限验证
   - 使用HTTPS和适当的安全头

示例：生产环境推荐配置

	// API v1 - 使用命名空间组织
	apiV1 := mvc.NewNamespace("/api/v1",
		mvc.NSBefore(corsMiddleware),
		mvc.NSBefore(authMiddleware),
		mvc.NSBefore(rateLimitMiddleware),

		// 用户模块
		mvc.NSNamespace("/users",
			mvc.NSManualRouter(&UserController{}, false,
				"GET:/|GetList",
				"GET:/:id|Get",
				"POST:/|PostCreate",
				"PUT:/:id|PutUpdate",
				"DELETE:/:id|DeleteRemove",
			),
		),

		// 管理员模块
		mvc.NSNamespace("/admin",
			mvc.NSBefore(adminAuthMiddleware),
			mvc.NSManualRouter(&AdminController{}, false,
				"GET:/dashboard|GetDashboard",
				"GET:/users|GetUsers",
			),
		),
	)

	mvc.AddNamespace(apiV1)

注意事项：
- 控制器实例在每次请求时会被重用，确保线程安全
- 避免在控制器中存储状态，使用Context传递数据
- 合理使用缓存减少数据库查询
- 监控路由性能，及时优化热点路径
*/
