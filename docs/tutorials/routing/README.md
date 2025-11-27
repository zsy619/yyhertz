# YYHertz 路由系统详解

<div align="center">

🛣️ **智能路由系统完全指南** | 从自动路由到高级配置

</div>

---

## 📋 目录

- [路由基础概念](#路由基础概念)
- [自动路由系统](#自动路由系统)
- [手动路由配置](#手动路由配置)
- [路由组和中间件](#路由组和中间件)
- [命名空间路由](#命名空间路由)
- [高级路由特性](#高级路由特性)

---

## 🎯 路由基础概念

### 什么是路由？
路由是URL路径与处理函数之间的映射关系，决定了用户访问特定URL时应该执行哪个控制器的哪个方法。

### YYHertz路由特性
- **🤖 自动路由**: 基于约定优于配置的自动路由注册
- **🎛️ 手动路由**: 灵活的手动路由配置
- **📦 路由组**: 支持路由分组和中间件
- **🏷️ 命名空间**: Beego风格的命名空间路由
- **🔀 RESTful**: 完整支持RESTful API设计

---

## 🤖 自动路由系统

### 1. 基本自动路由

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

// 用户控制器
type UserController struct {
    mvc.BaseController
}

// 自动路由规则: HTTP方法 + 函数名 = 路由路径
func (c *UserController) GetIndex() {
    // GET /user/index
    c.JSON(map[string]any{
        "message": "用户首页",
        "method":  "GET",
    })
}

func (c *UserController) PostCreate() {
    // POST /user/create
    name := c.GetForm("name")
    email := c.GetForm("email")
    
    c.JSON(map[string]any{
        "message": "创建用户",
        "data": map[string]string{
            "name":  name,
            "email": email,
        },
    })
}

func (c *UserController) PutUpdate() {
    // PUT /user/update
    c.JSON(map[string]any{
        "message": "更新用户",
        "method":  "PUT",
    })
}

func (c *UserController) DeleteRemove() {
    // DELETE /user/remove
    c.JSON(map[string]any{
        "message": "删除用户",
        "method":  "DELETE",
    })
}

func main() {
    app := mvc.HertzApp
    
    // 注册自动路由
    app.AutoRouters(&UserController{})
    
    app.Run(":8888")
}
```

### 2. 自动路由命名规则

```go
// 控制器命名规则
type ProductController struct {
    mvc.BaseController
}

// 支持的HTTP方法前缀
func (c *ProductController) GetIndex() {}     // GET  /product/index
func (c *ProductController) GetList() {}      // GET  /product/list  
func (c *ProductController) GetDetail() {}    // GET  /product/detail
func (c *ProductController) PostCreate() {}   // POST /product/create
func (c *ProductController) PostStore() {}    // POST /product/store
func (c *ProductController) PutUpdate() {}    // PUT  /product/update
func (c *ProductController) PatchModify() {}  // PATCH /product/modify
func (c *ProductController) DeleteRemove() {} // DELETE /product/remove
func (c *ProductController) DeleteDestroy() {} // DELETE /product/destroy
func (c *ProductController) HeadCheck() {}    // HEAD /product/check
func (c *ProductController) OptionsInfo() {}  // OPTIONS /product/info

// 特殊方法名处理
func (c *ProductController) GetUserInfo() {}  // GET /product/user-info (驼峰转连字符)
func (c *ProductController) PostBatchImport() {} // POST /product/batch-import
```

### 3. 参数路由

```go
type ArticleController struct {
    mvc.BaseController
}

// 带参数的路由方法
func (c *ArticleController) GetDetail() {
    // 自动路由会注册为 GET /article/detail
    // 但通常我们希望是 GET /article/:id
    // 这需要配合手动路由或参数获取
    
    id := c.GetParam("id")  // 从URL参数获取
    if id == "" {
        id = c.GetQuery("id")  // 从查询参数获取
    }
    
    c.JSON(map[string]any{
        "article_id": id,
        "title":      "文章标题",
        "content":    "文章内容",
    })
}

// 在main.go中可以结合手动路由
func main() {
    app := mvc.HertzApp
    
    // 自动路由
    app.AutoRouters(&ArticleController{})
    
    // 补充参数路由
    controller := &ArticleController{}
    app.GET("/article/:id", controller.GetDetail)
    
    app.Run(":8888")
}
```

### 4. 多控制器注册

```go
func main() {
    app := mvc.HertzApp
    
    // 批量注册多个控制器
    app.AutoRouters(
        &UserController{},      // /user/*
        &ProductController{},   // /product/*
        &ArticleController{},   // /article/*
        &OrderController{},     // /order/*
    )
    
    app.Run(":8888")
}
```

---

## 🎛️ 手动路由配置

### 1. 基础手动路由

```go
func main() {
    app := mvc.HertzApp
    
    // 基础路由
    app.GET("/", func(c *mvc.Context) {
        c.JSON(map[string]any{
            "message": "欢迎使用YYHertz",
            "version": "2.0.0",
        })
    })
    
    // 带参数的路由
    app.GET("/users/:id", func(c *mvc.Context) {
        id := c.Param("id")
        c.JSON(map[string]any{
            "user_id": id,
        })
    })
    
    // 通配符路由
    app.GET("/static/*filepath", func(c *mvc.Context) {
        filepath := c.Param("filepath")
        c.JSON(map[string]any{
            "file_path": filepath,
        })
    })
    
    // 多种HTTP方法
    app.POST("/users", createUser)
    app.PUT("/users/:id", updateUser) 
    app.DELETE("/users/:id", deleteUser)
    
    // Any方法支持所有HTTP方法
    app.Any("/ping", func(c *mvc.Context) {
        c.JSON(map[string]any{
            "message": "pong",
            "method":  c.Request.Method,
        })
    })
    
    app.Run(":8888")
}

// 处理函数
func createUser(c *mvc.Context) {
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    
    if err := c.BindJSON(&user); err != nil {
        c.JSON(400, map[string]any{"error": "Invalid JSON"})
        return
    }
    
    // 创建用户逻辑
    c.JSON(201, map[string]any{
        "message": "用户创建成功",
        "user":    user,
    })
}

func updateUser(c *mvc.Context) {
    id := c.Param("id")
    
    var updateData map[string]interface{}
    if err := c.BindJSON(&updateData); err != nil {
        c.JSON(400, map[string]any{"error": "Invalid JSON"})
        return
    }
    
    c.JSON(200, map[string]any{
        "message": "用户更新成功",
        "user_id": id,
        "data":    updateData,
    })
}

func deleteUser(c *mvc.Context) {
    id := c.Param("id")
    
    // 删除用户逻辑
    c.JSON(200, map[string]any{
        "message": "用户删除成功",
        "user_id": id,
    })
}
```

### 2. 控制器方法绑定

```go
func main() {
    app := mvc.HertzApp
    
    // 实例化控制器
    userController := &UserController{}
    
    // 手动绑定控制器方法
    app.GET("/api/users", userController.GetList)
    app.GET("/api/users/:id", userController.GetDetail)
    app.POST("/api/users", userController.PostCreate)
    app.PUT("/api/users/:id", userController.PutUpdate)
    app.DELETE("/api/users/:id", userController.DeleteRemove)
    
    app.Run(":8888")
}
```

### 3. 路由参数处理

```go
// 各种路由参数模式
func setupRoutes(app *mvc.Application) {
    // 1. 路径参数
    app.GET("/users/:id", func(c *mvc.Context) {
        id := c.Param("id")  // 获取路径参数
        c.JSON(map[string]any{"user_id": id})
    })
    
    // 2. 多个路径参数
    app.GET("/users/:userId/posts/:postId", func(c *mvc.Context) {
        userId := c.Param("userId")
        postId := c.Param("postId")
        c.JSON(map[string]any{
            "user_id": userId,
            "post_id": postId,
        })
    })
    
    // 3. 可选参数
    app.GET("/articles/*category", func(c *mvc.Context) {
        category := c.Param("category")
        if category == "" {
            category = "all"
        }
        c.JSON(map[string]any{"category": category})
    })
    
    // 4. 查询参数
    app.GET("/search", func(c *mvc.Context) {
        keyword := c.Query("q")              // 单个查询参数
        page := c.DefaultQuery("page", "1")  // 带默认值的查询参数
        tags := c.QueryArray("tags")         // 数组查询参数
        
        c.JSON(map[string]any{
            "keyword": keyword,
            "page":    page,
            "tags":    tags,
        })
    })
    
    // 5. 表单参数
    app.POST("/form", func(c *mvc.Context) {
        name := c.PostForm("name")
        email := c.PostForm("email")
        
        c.JSON(map[string]any{
            "name":  name,
            "email": email,
        })
    })
}
```

---

## 📦 路由组和中间件

### 1. 基础路由组

```go
func main() {
    app := mvc.HertzApp
    
    // 创建API路由组
    apiV1 := app.Group("/api/v1")
    {
        // 用户相关路由
        users := apiV1.Group("/users")
        {
            users.GET("", getUserList)
            users.GET("/:id", getUserDetail)
            users.POST("", createUser)
            users.PUT("/:id", updateUser)
            users.DELETE("/:id", deleteUser)
        }
        
        // 产品相关路由
        products := apiV1.Group("/products")
        {
            products.GET("", getProductList)
            products.GET("/:id", getProductDetail)
            products.POST("", createProduct)
        }
    }
    
    // 管理员路由组
    admin := app.Group("/admin")
    {
        admin.GET("/dashboard", adminDashboard)
        admin.GET("/users", adminUserList)
        admin.GET("/settings", adminSettings)
    }
    
    app.Run(":8888")
}
```

### 2. 路由组中间件

```go
import "github.com/zsy619/yyhertz/framework/mvc/middleware"

func main() {
    app := mvc.HertzApp
    
    // 全局中间件
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    
    // API路由组 - 需要认证
    apiV1 := app.Group("/api/v1", middleware.Auth())
    {
        apiV1.GET("/profile", getProfile)
        apiV1.PUT("/profile", updateProfile)
        
        // 管理员子组 - 需要管理员权限
        admin := apiV1.Group("/admin", middleware.AdminRequired())
        {
            admin.GET("/users", adminGetUsers)
            admin.DELETE("/users/:id", adminDeleteUser)
        }
    }
    
    // 公开API路由组 - 有限流
    public := app.Group("/public", middleware.RateLimit(100, 60))
    {
        public.GET("/articles", getPublicArticles)
        public.GET("/articles/:id", getPublicArticleDetail)
    }
    
    app.Run(":8888")
}

// 自定义中间件示例
func AdminRequired() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 检查用户是否为管理员
        user := getCurrentUser(c)
        if user == nil || !user.IsAdmin {
            c.JSON(403, map[string]any{
                "error": "需要管理员权限",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 3. 路由级中间件

```go
func main() {
    app := mvc.HertzApp
    
    // 单个路由使用中间件
    app.GET("/protected", 
        middleware.Auth(),           // 认证中间件
        middleware.RateLimit(10, 60), // 限流中间件
        protectedHandler,            // 处理函数
    )
    
    // 多个中间件组合
    app.POST("/sensitive", 
        middleware.Auth(),
        middleware.CSRF(),
        middleware.AdminRequired(),
        sensitiveHandler,
    )
    
    app.Run(":8888")
}

func protectedHandler(c *mvc.Context) {
    user := c.MustGet("user").(User)
    c.JSON(map[string]any{
        "message": "这是受保护的资源",
        "user":    user.Username,
    })
}

func sensitiveHandler(c *mvc.Context) {
    c.JSON(map[string]any{
        "message": "敏感操作完成",
    })
}
```

---

## 🏷️ 命名空间路由

### 1. Beego风格命名空间

```go
import "github.com/zsy619/yyhertz/framework/mvc"

func main() {
    app := mvc.HertzApp
    
    // 创建API v1命名空间
    nsV1 := mvc.NewNamespace("/api/v1").
        Middleware(middleware.Auth()).  // 命名空间中间件
        AutoRouter(&UserController{}).  // 自动路由
        Router("/custom", &CustomController{}, "GET:GetData;POST:PostData") // 自定义路由
    
    // 创建API v2命名空间
    nsV2 := mvc.NewNamespace("/api/v2").
        Middleware(middleware.Auth(), middleware.RateLimit(200, 60)).
        AutoRouter(&UserV2Controller{}).
        Router("/upload", &FileController{}, "POST:Upload")
    
    // 管理员命名空间
    nsAdmin := mvc.NewNamespace("/admin").
        Middleware(middleware.Auth(), middleware.AdminRequired()).
        AutoRouter(&AdminController{}).
        Router("/stats", &StatsController{}, "GET:GetStats")
    
    // 注册命名空间
    mvc.AddNamespace(nsV1, nsV2, nsAdmin)
    
    app.Run(":8888")
}
```

### 2. 复杂命名空间配置

```go
func setupNamespaces() {
    // API命名空间 - 支持多版本
    apiNS := mvc.NewNamespace("/api")
    
    // v1 子命名空间
    v1NS := mvc.NewNamespace("/v1").
        Middleware(middleware.APIKeyAuth()).  // API Key认证
        AutoRouter(&UserController{}).
        Router("/auth/login", &AuthController{}, "POST:Login").
        Router("/auth/logout", &AuthController{}, "POST:Logout")
    
    // v2 子命名空间  
    v2NS := mvc.NewNamespace("/v2").
        Middleware(middleware.JWTAuth()).     // JWT认证
        AutoRouter(&UserV2Controller{}).
        Router("/auth/token", &AuthV2Controller{}, "POST:GetToken")
    
    // 嵌套命名空间
    apiNS.Namespace(v1NS, v2NS)
    
    // 移动端专用命名空间
    mobileNS := mvc.NewNamespace("/mobile").
        Middleware(
            middleware.MobileUserAgent(),   // 移动端检测
            middleware.RateLimit(50, 60),  // 移动端限流
        ).
        AutoRouter(&MobileController{})
    
    // Web端命名空间
    webNS := mvc.NewNamespace("/web").
        Middleware(middleware.SessionAuth()). // Session认证
        AutoRouter(&WebController{})
    
    mvc.AddNamespace(apiNS, mobileNS, webNS)
}
```

---

## 🚀 高级路由特性

### 1. 路由条件和约束

```go
func main() {
    app := mvc.HertzApp
    
    // 主机名约束
    app.GET("/api/*", func(c *mvc.Context) {
        if c.Request.Host != "api.example.com" {
            c.JSON(404, map[string]any{"error": "Not Found"})
            return
        }
        c.JSON(map[string]any{"message": "API endpoint"})
    })
    
    // 请求头约束
    app.GET("/json", func(c *mvc.Context) {
        if c.GetHeader("Accept") != "application/json" {
            c.JSON(406, map[string]any{"error": "Only JSON accepted"})
            return
        }
        c.JSON(map[string]any{"data": "JSON response"})
    })
    
    // 参数验证约束
    app.GET("/users/:id", func(c *mvc.Context) {
        id := c.Param("id")
        
        // 验证ID格式
        if userID, err := strconv.Atoi(id); err != nil || userID <= 0 {
            c.JSON(400, map[string]any{"error": "Invalid user ID"})
            return
        }
        
        c.JSON(map[string]any{"user_id": id})
    })
    
    app.Run(":8888")
}
```

### 2. 子域名路由

```go
func setupSubdomainRoutes() {
    app := mvc.HertzApp
    
    // 主域名路由
    app.GET("/", func(c *mvc.Context) {
        c.JSON(map[string]any{"domain": "main"})
    })
    
    // 子域名路由处理
    app.Use(func(c *mvc.Context) {
        host := c.Request.Host
        
        switch {
        case strings.HasPrefix(host, "api."):
            // API子域名
            c.Set("subdomain", "api")
        case strings.HasPrefix(host, "admin."):
            // 管理员子域名
            c.Set("subdomain", "admin")
        case strings.HasPrefix(host, "mobile."):
            // 移动端子域名
            c.Set("subdomain", "mobile")
        default:
            c.Set("subdomain", "www")
        }
        
        c.Next()
    })
    
    // 根据子域名分发
    app.GET("/dashboard", func(c *mvc.Context) {
        subdomain := c.GetString("subdomain")
        
        switch subdomain {
        case "api":
            c.JSON(map[string]any{"type": "API Dashboard"})
        case "admin":
            c.JSON(map[string]any{"type": "Admin Dashboard"})
        default:
            c.JSON(map[string]any{"type": "User Dashboard"})
        }
    })
}
```

### 3. 动态路由

```go
// 动态路由配置
type RouteConfig struct {
    Path       string            `json:"path"`
    Method     string            `json:"method"`
    Controller string            `json:"controller"`
    Action     string            `json:"action"`
    Middleware []string          `json:"middleware"`
    Params     map[string]string `json:"params"`
}

func setupDynamicRoutes(app *mvc.Application) {
    // 从配置文件或数据库加载路由配置
    routes := loadRouteConfig()
    
    for _, route := range routes {
        registerDynamicRoute(app, route)
    }
}

func registerDynamicRoute(app *mvc.Application, config RouteConfig) {
    // 动态创建处理函数
    handler := func(c *mvc.Context) {
        // 根据配置调用对应的控制器和方法
        result := callControllerAction(config.Controller, config.Action, c)
        c.JSON(result)
    }
    
    // 注册路由
    switch config.Method {
    case "GET":
        app.GET(config.Path, handler)
    case "POST":
        app.POST(config.Path, handler)
    case "PUT":
        app.PUT(config.Path, handler)
    case "DELETE":
        app.DELETE(config.Path, handler)
    }
}

func loadRouteConfig() []RouteConfig {
    // 从JSON文件加载配置
    return []RouteConfig{
        {
            Path:       "/dynamic/users",
            Method:     "GET",
            Controller: "UserController",
            Action:     "GetList",
            Middleware: []string{"auth"},
        },
        {
            Path:       "/dynamic/users/:id",
            Method:     "GET", 
            Controller: "UserController",
            Action:     "GetDetail",
            Middleware: []string{"auth"},
        },
    }
}
```

### 4. 路由缓存和优化

```go
// 路由性能优化
func optimizeRoutes(app *mvc.Application) {
    // 1. 静态路由优先
    app.GET("/health", healthCheck)      // 静态路由，查找最快
    app.GET("/metrics", metricsHandler)  // 静态路由
    
    // 2. 参数路由其次  
    app.GET("/users/:id", getUserDetail)         // 单参数路由
    app.GET("/users/:id/posts/:pid", getPost)    // 多参数路由
    
    // 3. 通配符路由最后
    app.GET("/static/*filepath", serveStatic)    // 通配符路由，查找最慢
    
    // 4. 路由预编译
    app.PrecompileRoutes()  // 预编译路由正则表达式
}

// 路由统计中间件
func RouteStatsMiddleware() mvc.HandlerFunc {
    routeStats := make(map[string]*RouteMetrics)
    var mutex sync.RWMutex
    
    return func(c *mvc.Context) {
        start := time.Now()
        path := c.FullPath()
        
        c.Next()
        
        duration := time.Since(start)
        
        // 更新路由统计
        mutex.Lock()
        if stats, exists := routeStats[path]; exists {
            stats.Count++
            stats.TotalTime += duration
            stats.AvgTime = stats.TotalTime / time.Duration(stats.Count)
        } else {
            routeStats[path] = &RouteMetrics{
                Path:      path,
                Count:     1,
                TotalTime: duration,
                AvgTime:   duration,
            }
        }
        mutex.Unlock()
    }
}

type RouteMetrics struct {
    Path      string
    Count     int64
    TotalTime time.Duration
    AvgTime   time.Duration
}
```

---

## 📊 路由最佳实践

### 1. RESTful API设计

```go
func setupRESTfulRoutes(app *mvc.Application) {
    // 用户资源的RESTful路由
    users := app.Group("/api/users")
    {
        users.GET("", getUserList)           // GET    /api/users          - 获取用户列表
        users.POST("", createUser)           // POST   /api/users          - 创建用户
        users.GET("/:id", getUserDetail)     // GET    /api/users/:id      - 获取用户详情
        users.PUT("/:id", updateUser)        // PUT    /api/users/:id      - 完整更新用户
        users.PATCH("/:id", patchUser)       // PATCH  /api/users/:id      - 部分更新用户
        users.DELETE("/:id", deleteUser)     // DELETE /api/users/:id      - 删除用户
    }
    
    // 嵌套资源路由
    posts := users.Group("/:userId/posts")
    {
        posts.GET("", getUserPosts)          // GET    /api/users/:userId/posts      - 获取用户文章
        posts.POST("", createUserPost)       // POST   /api/users/:userId/posts      - 创建用户文章
        posts.GET("/:postId", getUserPost)   // GET    /api/users/:userId/posts/:postId - 获取特定文章
        posts.PUT("/:postId", updateUserPost)    // PUT    /api/users/:userId/posts/:postId - 更新文章
        posts.DELETE("/:postId", deleteUserPost) // DELETE /api/users/:userId/posts/:postId - 删除文章
    }
}
```

### 2. 版本控制策略

```go
// 方式1: URL路径版本控制
func setupVersionRoutes(app *mvc.Application) {
    // API v1
    v1 := app.Group("/api/v1")
    {
        v1.GET("/users", getUsersV1)
        v1.POST("/users", createUserV1)
    }
    
    // API v2
    v2 := app.Group("/api/v2")
    {
        v2.GET("/users", getUsersV2)
        v2.POST("/users", createUserV2)
    }
    
    // 默认版本(最新版本)
    app.Group("/api").
        GET("/users", getUsersV2).  // 指向最新版本
        POST("/users", createUserV2)
}

// 方式2: Header版本控制
func VersionMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        version := c.GetHeader("API-Version")
        if version == "" {
            version = "v2" // 默认版本
        }
        
        c.Set("api_version", version)
        c.Next()
    }
}

func versionedHandler(c *mvc.Context) {
    version := c.GetString("api_version")
    
    switch version {
    case "v1":
        handleV1Request(c)
    case "v2":
        handleV2Request(c)
    default:
        c.JSON(400, map[string]any{
            "error": "Unsupported API version",
        })
    }
}
```

### 3. 错误处理和状态码

```go
// 统一的错误处理
func handleAPIError(c *mvc.Context, err error) {
    switch e := err.(type) {
    case *ValidationError:
        c.JSON(400, map[string]any{
            "error":   "Validation failed",
            "details": e.Details,
        })
    case *NotFoundError:
        c.JSON(404, map[string]any{
            "error": "Resource not found",
        })
    case *UnauthorizedError:
        c.JSON(401, map[string]any{
            "error": "Unauthorized",
        })
    case *ForbiddenError:
        c.JSON(403, map[string]any{
            "error": "Forbidden",
        })
    default:
        c.JSON(500, map[string]any{
            "error": "Internal server error",
        })
    }
}

// 标准的API响应格式
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     string      `json:"error,omitempty"`
    Code      int         `json:"code"`
    Message   string      `json:"message,omitempty"`
    Timestamp int64       `json:"timestamp"`
}

func successResponse(c *mvc.Context, data interface{}) {
    c.JSON(200, APIResponse{
        Success:   true,
        Data:      data,
        Code:      200,
        Timestamp: time.Now().Unix(),
    })
}

func errorResponse(c *mvc.Context, code int, message string) {
    c.JSON(code, APIResponse{
        Success:   false,
        Error:     message,
        Code:      code,
        Timestamp: time.Now().Unix(),
    })
}
```

---

<div align="center">

**🛣️ 掌握YYHertz路由系统，让URL设计更加优雅！**

**合理的路由设计是API成功的基础 🎯**

</div>