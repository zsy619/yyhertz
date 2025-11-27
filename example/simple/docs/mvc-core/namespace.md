# 📁 命名空间

YYHertz的命名空间系统完全兼容Beego语法，提供了灵活而强大的路由组织方式。命名空间让您可以按功能模块、API版本或访问权限来组织路由，使大型应用的路由管理变得清晰有序。

## 🌟 核心概念

### 什么是命名空间？
命名空间是一种路由分组机制，它允许您：
- **分组相关路由** - 将相关功能的路由组织在一起
- **共享中间件** - 同组路由共享认证、日志等中间件  
- **嵌套管理** - 支持多层嵌套的复杂应用结构
- **版本控制** - 方便进行API版本管理

### 路由树结构
```
/api/v1/
├── /users/
│   ├── GET    /list      → UserController.GetList
│   ├── POST   /create    → UserController.PostCreate  
│   ├── PUT    /:id       → UserController.PutUpdate
│   └── DELETE /:id       → UserController.DeleteRemove
├── /products/
│   ├── GET    /          → ProductController.GetIndex
│   └── POST   /          → ProductController.PostCreate
└── /admin/
    ├── /users/           → 管理员用户路由
    └── /system/          → 系统管理路由
```

## 🏗️ 基础用法

### 简单命名空间
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
    app := mvc.HertzApp
    
    // 创建API命名空间
    nsAPI := mvc.NewNamespace("/api",
        // 自动路由注册
        mvc.NSAutoRouter(&UserController{}),
        mvc.NSAutoRouter(&ProductController{}),
    )
    
    // 注册命名空间到全局应用
    mvc.AddNamespace(nsAPI)
    
    app.Run(":8888")
}

// 用户控制器
type UserController struct {
    mvc.BaseController
}

func (c *UserController) GetList() {
    // GET /api/user/list
    c.JSON(map[string]any{"users": []string{"user1", "user2"}})
}

func (c *UserController) PostCreate() {
    // POST /api/user/create
    name := c.GetForm("name")
    c.JSON(map[string]any{"success": true, "user": name})
}
```

### 手动路由映射
```go
func main() {
    app := mvc.HertzApp
    
    // 创建API命名空间，使用手动路由
    nsAPI := mvc.NewNamespace("/api",
        // 指定HTTP方法和路由路径
        mvc.NSRouter("/users", &UserController{}, "GET:GetList"),
        mvc.NSRouter("/users", &UserController{}, "POST:CreateUser"), 
        mvc.NSRouter("/users/:id", &UserController{}, "GET:GetUser"),
        mvc.NSRouter("/users/:id", &UserController{}, "PUT:UpdateUser"),
        mvc.NSRouter("/users/:id", &UserController{}, "DELETE:DeleteUser"),
        
        // 支持所有HTTP方法
        mvc.NSRouter("/upload", &FileController{}, "*:HandleUpload"),
    )
    
    mvc.AddNamespace(nsAPI)
    app.Run(":8888")
}

type UserController struct {
    mvc.BaseController
}

// 获取用户列表
func (c *UserController) GetList() {
    // GET /api/users
    c.JSON(map[string]any{"users": []User{}})
}

// 创建用户 
func (c *UserController) CreateUser() {
    // POST /api/users
    c.JSON(map[string]any{"success": true})
}

// 获取单个用户
func (c *UserController) GetUser() {
    // GET /api/users/:id
    id := c.GetParam("id")
    c.JSON(map[string]any{"id": id})
}
```

## 🔗 嵌套命名空间

### 多层嵌套结构
```go
func main() {
    app := mvc.HertzApp
    
    // 创建主API命名空间
    nsAPI := mvc.NewNamespace("/api",
        // 添加全局API中间件
        mvc.NSBefore(middleware.CORS()),
        mvc.NSBefore(middleware.RateLimit(100, time.Minute)),
        
        // v1版本命名空间
        mvc.NSNamespace("/v1",
            mvc.NSBefore(middleware.Auth(middleware.AuthConfig{
                Strategy: middleware.AuthJWT,
            })),
            
            // 用户管理模块
            mvc.NSNamespace("/users",
                mvc.NSRouter("/", &UserController{}, "GET:GetList"),
                mvc.NSRouter("/", &UserController{}, "POST:Create"),
                mvc.NSRouter("/:id", &UserController{}, "GET:GetDetail"),
                mvc.NSRouter("/:id", &UserController{}, "PUT:Update"),
                mvc.NSRouter("/:id", &UserController{}, "DELETE:Remove"),
                
                // 用户相关的子资源
                mvc.NSNamespace("/:user_id/posts",
                    mvc.NSRouter("/", &PostController{}, "GET:GetUserPosts"),
                    mvc.NSRouter("/", &PostController{}, "POST:CreateUserPost"),
                ),
            ),
            
            // 产品管理模块
            mvc.NSNamespace("/products",
                mvc.NSAutoRouter(&ProductController{}),
                
                // 产品分类
                mvc.NSNamespace("/categories",
                    mvc.NSAutoRouter(&CategoryController{}),
                ),
            ),
        ),
        
        // v2版本命名空间
        mvc.NSNamespace("/v2",
            mvc.NSRouter("/users", &UserV2Controller{}, "*:HandleUser"),
            mvc.NSRouter("/products", &ProductV2Controller{}, "*:HandleProduct"),
        ),
        
        // 管理后台命名空间
        mvc.NSNamespace("/admin",
            mvc.NSBefore(middleware.BasicAuth(map[string]string{
                "admin": "secret123",
            })),
            
            mvc.NSNamespace("/system",
                mvc.NSAutoRouter(&SystemController{}),
            ),
            
            mvc.NSNamespace("/users",
                mvc.NSAutoRouter(&AdminUserController{}),
            ),
        ),
    )
    
    mvc.AddNamespace(nsAPI)
    app.Run(":8888")
}
```

生成的路由结构：
```
GET    /api/v1/users/              → UserController.GetList
POST   /api/v1/users/              → UserController.Create  
GET    /api/v1/users/:id           → UserController.GetDetail
PUT    /api/v1/users/:id           → UserController.Update
DELETE /api/v1/users/:id           → UserController.Remove
GET    /api/v1/users/:user_id/posts/ → PostController.GetUserPosts
POST   /api/v1/users/:user_id/posts/ → PostController.CreateUserPost

GET    /api/v1/products/           → ProductController.GetIndex
POST   /api/v1/products/create     → ProductController.PostCreate
GET    /api/v1/products/categories/ → CategoryController.GetIndex

*      /api/v2/users               → UserV2Controller.HandleUser  
*      /api/v2/products            → ProductV2Controller.HandleProduct

GET    /api/admin/system/          → SystemController.GetIndex
GET    /api/admin/users/           → AdminUserController.GetList
```

## 🛡️ 中间件集成

### 命名空间级别中间件
```go
func main() {
    app := mvc.HertzApp
    
    // 认证中间件
    authMiddleware := middleware.Auth(middleware.AuthConfig{
        Strategy:  middleware.AuthJWT,
        TokenKey:  "Authorization",
        SkipPaths: []string{"/api/v1/auth/login"},
    })
    
    // 日志中间件
    logMiddleware := middleware.Logger(middleware.LoggerConfig{
        Format: "[${time}] ${status} - ${method} ${path} (${latency})",
        Output: middleware.DefaultWriter,
    })
    
    // 限流中间件
    rateLimitMiddleware := middleware.RateLimit(1000, time.Hour)
    
    nsAPI := mvc.NewNamespace("/api",
        // 全局API中间件
        mvc.NSBefore(logMiddleware),
        mvc.NSBefore(middleware.CORS()),
        
        // v1版本命名空间
        mvc.NSNamespace("/v1",
            // v1版本专用中间件
            mvc.NSBefore(authMiddleware),
            mvc.NSBefore(rateLimitMiddleware),
            
            // 公开API (不需要认证)
            mvc.NSNamespace("/public",
                mvc.NSAutoRouter(&PublicController{}),
            ),
            
            // 用户API (需要认证)
            mvc.NSNamespace("/users",
                // 用户模块额外中间件
                mvc.NSBefore(middleware.Permission("user.read")),
                mvc.NSAutoRouter(&UserController{}),
            ),
            
            // 管理员API (需要管理员权限)
            mvc.NSNamespace("/admin",
                mvc.NSBefore(middleware.Permission("admin.manage")),
                mvc.NSAutoRouter(&AdminController{}),
            ),
        ),
    )
    
    mvc.AddNamespace(nsAPI)
    app.Run(":8888")
}
```

### 条件中间件
```go
func main() {
    app := mvc.HertzApp
    
    // 创建条件中间件
    conditionalAuth := func(c *mvc.Context) {
        // 只对非公开接口进行认证检查
        if strings.HasPrefix(c.Request.URI().Path(), "/api/v1/public/") {
            c.Next()
            return
        }
        
        // 执行认证检查
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, map[string]string{"error": "未授权访问"})
            return
        }
        
        // 验证token逻辑...
        c.Next()
    }
    
    nsAPI := mvc.NewNamespace("/api/v1",
        mvc.NSBefore(conditionalAuth),
        
        mvc.NSRouter("/public/health", &HealthController{}, "GET:Check"),
        mvc.NSRouter("/users", &UserController{}, "GET:List"),
        mvc.NSRouter("/admin/stats", &AdminController{}, "GET:Stats"),
    )
    
    mvc.AddNamespace(nsAPI)
    app.Run(":8888")
}
```

## 🔧 高级特性

### 动态命名空间注册
```go
package main

import (
    "fmt"
    "reflect"
    "github.com/zsy619/yyhertz/framework/mvc"
)

// 控制器注册器
type ControllerRegistry struct {
    controllers map[string]interface{}
    namespaces  map[string]*mvc.Namespace
}

func NewControllerRegistry() *ControllerRegistry {
    return &ControllerRegistry{
        controllers: make(map[string]interface{}),
        namespaces:  make(map[string]*mvc.Namespace),
    }
}

func (r *ControllerRegistry) Register(prefix string, controller interface{}) {
    controllerName := r.getControllerName(controller)
    r.controllers[controllerName] = controller
    
    // 动态创建命名空间
    ns := mvc.NewNamespace(prefix,
        mvc.NSAutoRouter(controller),
    )
    
    r.namespaces[prefix] = ns
    mvc.AddNamespace(ns)
}

func (r *ControllerRegistry) getControllerName(controller interface{}) string {
    t := reflect.TypeOf(controller)
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    return t.Name()
}

func main() {
    app := mvc.HertzApp
    registry := NewControllerRegistry()
    
    // 动态注册多个模块
    modules := map[string]interface{}{
        "/api/users":     &UserController{},
        "/api/products":  &ProductController{},
        "/api/orders":    &OrderController{},
        "/web/admin":     &AdminController{},
        "/web/dashboard": &DashboardController{},
    }
    
    for prefix, controller := range modules {
        registry.Register(prefix, controller)
    }
    
    app.Run(":8888")
}
```

### 命名空间中间件链
```go
type MiddlewareChain struct {
    middlewares []mvc.HandlerFunc
}

func NewMiddlewareChain() *MiddlewareChain {
    return &MiddlewareChain{
        middlewares: make([]mvc.HandlerFunc, 0),
    }
}

func (mc *MiddlewareChain) Use(middleware mvc.HandlerFunc) *MiddlewareChain {
    mc.middlewares = append(mc.middlewares, middleware)
    return mc
}

func (mc *MiddlewareChain) Build() []interface{} {
    chain := make([]interface{}, len(mc.middlewares))
    for i, mw := range mc.middlewares {
        chain[i] = mvc.NSBefore(mw)
    }
    return chain
}

func main() {
    app := mvc.HertzApp
    
    // 构建API中间件链
    apiChain := NewMiddlewareChain().
        Use(middleware.Logger()).
        Use(middleware.CORS()).
        Use(middleware.RateLimit(1000, time.Hour))
    
    // 构建认证中间件链  
    authChain := NewMiddlewareChain().
        Use(middleware.Auth(middleware.AuthConfig{
            Strategy: middleware.AuthJWT,
        })).
        Use(middleware.Permission("api.access"))
    
    // 创建命名空间时使用中间件链
    nsAPI := mvc.NewNamespace("/api",
        append(
            apiChain.Build(),
            mvc.NSNamespace("/v1",
                append(
                    authChain.Build(),
                    mvc.NSAutoRouter(&UserController{}),
                    mvc.NSAutoRouter(&ProductController{}),
                )...,
            ),
        )...,
    )
    
    mvc.AddNamespace(nsAPI)
    app.Run(":8888")
}
```

## 📋 路由方法完整列表

### 支持的HTTP方法映射
```go
// 单一HTTP方法
"GET:MethodName"     → GET请求映射
"POST:MethodName"    → POST请求映射  
"PUT:MethodName"     → PUT请求映射
"PATCH:MethodName"   → PATCH请求映射
"DELETE:MethodName"  → DELETE请求映射
"HEAD:MethodName"    → HEAD请求映射
"OPTIONS:MethodName" → OPTIONS请求映射

// 所有HTTP方法
"*:MethodName"       → 支持所有HTTP方法

// 多HTTP方法 (用逗号分隔)
"GET,POST:MethodName" → 支持GET和POST方法
"PUT,PATCH:MethodName" → 支持PUT和PATCH方法
```

### 路由参数和查询参数
```go
type UserController struct {
    mvc.BaseController
}

// GET /api/users/:id/posts/:post_id
func (c *UserController) GetUserPost() {
    userID := c.GetParam("id")        // 路径参数
    postID := c.GetParam("post_id")   // 路径参数
    page := c.GetQuery("page")        // 查询参数
    limit := c.GetQuery("limit")      // 查询参数
    
    c.JSON(map[string]any{
        "user_id":  userID,
        "post_id":  postID,
        "page":     page,
        "limit":    limit,
    })
}

func main() {
    nsAPI := mvc.NewNamespace("/api",
        mvc.NSRouter("/users/:id/posts/:post_id", &UserController{}, "GET:GetUserPost"),
    )
    
    mvc.AddNamespace(nsAPI)
    mvc.HertzApp.Run(":8888")
}

// 访问: GET /api/users/123/posts/456?page=1&limit=10
// 输出: {"user_id":"123", "post_id":"456", "page":"1", "limit":"10"}
```

## 🎯 实际应用案例

### 电商API设计
```go
func setupECommerceAPI() {
    app := mvc.HertzApp
    
    // 主API命名空间
    ecommerceAPI := mvc.NewNamespace("/api",
        mvc.NSBefore(middleware.Logger()),
        mvc.NSBefore(middleware.CORS()),
        mvc.NSBefore(middleware.RateLimit(2000, time.Hour)),
        
        // v1 API版本
        mvc.NSNamespace("/v1",
            // 公开API (无需认证)
            mvc.NSNamespace("/public",
                mvc.NSRouter("/products", &ProductController{}, "GET:GetPublicList"),
                mvc.NSRouter("/products/:id", &ProductController{}, "GET:GetPublicDetail"),
                mvc.NSRouter("/categories", &CategoryController{}, "GET:GetList"),
                mvc.NSRouter("/brands", &BrandController{}, "GET:GetList"),
            ),
            
            // 用户API (需要用户认证)
            mvc.NSNamespace("/user",
                mvc.NSBefore(middleware.Auth(middleware.AuthConfig{
                    Strategy: middleware.AuthJWT,
                    UserRole: "user",
                })),
                
                // 用户资料管理
                mvc.NSNamespace("/profile",
                    mvc.NSRouter("/", &UserController{}, "GET:GetProfile"),
                    mvc.NSRouter("/", &UserController{}, "PUT:UpdateProfile"),
                    mvc.NSRouter("/avatar", &UserController{}, "POST:UploadAvatar"),
                ),
                
                // 购物车管理
                mvc.NSNamespace("/cart",
                    mvc.NSAutoRouter(&CartController{}),
                ),
                
                // 订单管理
                mvc.NSNamespace("/orders",
                    mvc.NSRouter("/", &OrderController{}, "GET:GetUserOrders"),
                    mvc.NSRouter("/", &OrderController{}, "POST:CreateOrder"),
                    mvc.NSRouter("/:id", &OrderController{}, "GET:GetOrderDetail"),
                    mvc.NSRouter("/:id/cancel", &OrderController{}, "POST:CancelOrder"),
                    mvc.NSRouter("/:id/pay", &OrderController{}, "POST:PayOrder"),
                ),
                
                // 收货地址
                mvc.NSNamespace("/addresses",
                    mvc.NSAutoRouter(&AddressController{}),
                ),
            ),
            
            // 商家API (需要商家认证)
            mvc.NSNamespace("/merchant",
                mvc.NSBefore(middleware.Auth(middleware.AuthConfig{
                    Strategy: middleware.AuthJWT,
                    UserRole: "merchant",
                })),
                
                // 商品管理
                mvc.NSNamespace("/products",
                    mvc.NSAutoRouter(&MerchantProductController{}),
                ),
                
                // 订单管理
                mvc.NSNamespace("/orders",
                    mvc.NSAutoRouter(&MerchantOrderController{}),
                ),
                
                // 店铺管理
                mvc.NSNamespace("/shop",
                    mvc.NSAutoRouter(&ShopController{}),
                ),
            ),
            
            // 管理员API (需要管理员权限)
            mvc.NSNamespace("/admin",
                mvc.NSBefore(middleware.Auth(middleware.AuthConfig{
                    Strategy: middleware.AuthJWT,  
                    UserRole: "admin",
                })),
                
                // 用户管理
                mvc.NSNamespace("/users",
                    mvc.NSAutoRouter(&AdminUserController{}),
                ),
                
                // 商品管理
                mvc.NSNamespace("/products",
                    mvc.NSAutoRouter(&AdminProductController{}),
                ),
                
                // 订单管理
                mvc.NSNamespace("/orders",
                    mvc.NSAutoRouter(&AdminOrderController{}),
                ),
                
                // 系统设置
                mvc.NSNamespace("/system",
                    mvc.NSAutoRouter(&SystemController{}),
                ),
            ),
        ),
    )
    
    mvc.AddNamespace(ecommerceAPI)
}
```

### 多租户SaaS应用
```go
func setupSaaSAPI() {
    app := mvc.HertzApp
    
    saasAPI := mvc.NewNamespace("/api",
        mvc.NSBefore(middleware.Logger()),
        mvc.NSBefore(middleware.CORS()),
        
        // 租户识别中间件
        mvc.NSBefore(func(c *mvc.Context) {
            tenantID := c.GetHeader("X-Tenant-ID")
            if tenantID == "" {
                c.AbortWithStatusJSON(400, map[string]string{"error": "租户ID必需"})
                return
            }
            c.Set("tenant_id", tenantID)
            c.Next()
        }),
        
        // v1 API版本
        mvc.NSNamespace("/v1",
            mvc.NSBefore(middleware.Auth(middleware.AuthConfig{
                Strategy: middleware.AuthJWT,
            })),
            
            // 租户级别的路由
            mvc.NSNamespace("/tenant",
                // 租户管理
                mvc.NSNamespace("/management",
                    mvc.NSBefore(middleware.Permission("tenant.manage")),
                    mvc.NSAutoRouter(&TenantController{}),
                ),
                
                // 用户管理 (租户内用户)
                mvc.NSNamespace("/users",
                    mvc.NSBefore(middleware.Permission("user.manage")),
                    mvc.NSAutoRouter(&TenantUserController{}),
                ),
                
                // 应用数据
                mvc.NSNamespace("/data",
                    mvc.NSBefore(middleware.Permission("data.access")),
                    mvc.NSAutoRouter(&DataController{}),
                ),
                
                // 报表分析
                mvc.NSNamespace("/analytics",
                    mvc.NSBefore(middleware.Permission("analytics.view")),
                    mvc.NSAutoRouter(&AnalyticsController{}),
                ),
            ),
        ),
    )
    
    mvc.AddNamespace(saasAPI)
}
```

## 🚀 最佳实践

### 1. 命名规范
```
/api/v{version}          → API版本控制
/api/v1/public           → 公开接口
/api/v1/users            → 资源复数形式
/api/v1/users/:id        → 资源ID参数
/api/v1/users/:id/posts  → 嵌套资源
/web                     → Web页面路由
/admin                   → 管理后台路由
```

### 2. 中间件分层
```
Global Level    → 全局中间件 (日志、CORS等)
Namespace Level → 命名空间中间件 (认证、限流等)
Route Level     → 路由中间件 (权限、缓存等)
Controller Level → 控制器中间件 (参数验证等)
```

### 3. 版本管理策略
```go
// 通过URL版本控制
/api/v1/users  → Version 1.0
/api/v2/users  → Version 2.0

// 通过Header版本控制
versionMiddleware := func(c *mvc.Context) {
    version := c.GetHeader("API-Version")
    if version == "" {
        version = "v1"  // 默认版本
    }
    c.Set("api_version", version)
    c.Next()
}
```

### 4. 错误处理
```go
// 统一错误处理中间件
errorHandler := func(c *mvc.Context) {
    c.Next()
    
    // 检查是否有错误
    if len(c.Errors) > 0 {
        err := c.Errors.Last()
        c.JSON(map[string]any{
            "error": err.Error(),
            "code":  c.Writer.Status(),
            "path":  c.Request.URI().Path(),
        })
    }
}

nsAPI := mvc.NewNamespace("/api",
    mvc.NSBefore(errorHandler),
    // ... 其他路由
)
```

## 📖 下一步

现在您已经掌握了YYHertz命名空间的强大功能，建议继续学习：

1. 🔌 [中间件系统](/home/middleware-overview) - 深入了解中间件机制
2. 🛡️ [认证中间件](/home/builtin-middleware) - 实现用户认证和授权
3. 📊 [性能监控](/home/performance) - 监控命名空间性能
4. 🧪 [测试工具](/home/testing) - 编写命名空间测试

---

**🌟 命名空间是组织大型应用路由的最佳方式，掌握它将让您的应用架构更加清晰！**