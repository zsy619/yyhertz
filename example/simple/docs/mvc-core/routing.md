# 🛣️ 路由系统

YYHertz提供了革命性的路由系统，支持传统的控制器模式和全新的多Handler类型系统，让您可以选择最适合的开发方式。

## 🌟 路由系统特性

### 🎯 双模式支持
- **传统控制器模式** - 完全兼容Beego风格的MVC开发
- **多Handler类型系统** - 7种专门优化的Handler类型
- **混合使用** - 两种模式可以在同一应用中并存

### ⚡ 性能优化
- **零拷贝路由匹配** - 高效的路径解析
- **对象池化** - Context复用减少GC压力
- **智能编译优化** - 路由表自动优化

## 🎯 多Handler类型路由系统 (推荐)

### 路由组创建

```go
func main() {
    app := mvc.HertzApp
    
    // 创建API路由组
    apiGroup := mvc.CreateGroup("/api/v1")
    
    // 创建管理员路由组
    adminGroup := mvc.CreateGroup("/admin")
    
    // 设置路由
    setupAPIRoutes(apiGroup)
    setupAdminRoutes(adminGroup)
    
    app.Run(":8080")
}
```

### 7种Handler类型路由

```go
func setupAPIRoutes(group *router.Group) {
    // 1. LightHandler - 零开销健康检查
    group.GETLight("/health", func() {
        // 自动返回200 OK，无响应体
    })
    
    // 2. SimpleHandler - 简单业务逻辑
    group.GETSimple("/ping", func(ctx context.Context) {
        log.Printf("Ping at %s", time.Now().Format("15:04:05"))
    })
    
    // 3. DirectHandler - 完全控制响应
    group.GETDirect("/info", func(c *mvcContext.Context) {
        c.Set("handler_type", "direct")
        reqCtx := c.RequestContext()
        reqCtx.JSON(200, map[string]any{
            "message": "Direct handler response",
            "timestamp": time.Now(),
        })
    })
    
    // 4. ResponseHandler - 自动JSON序列化
    group.GETResponse("/users", func(c *mvcContext.Context) any {
        return map[string]any{
            "users": []User{},
            "total": 0,
        }
    })
    
    // 5. AsyncHandler - 异步处理
    group.POSTAsync("/process", func(c *mvcContext.Context) <-chan any {
        result := make(chan any, 1)
        go func() {
            defer close(result)
            // 异步处理逻辑
            time.Sleep(2 * time.Second)
            result <- map[string]any{"status": "completed"}
        }()
        return result
    })
    
    // 6. StreamHandler - 流式数据传输  
    group.GETStream("/logs", func(c *mvcContext.Context, dataChan chan<- []byte) error {
        for i := 1; i <= 10; i++ {
            data := fmt.Sprintf("Log entry %d\n", i)
            dataChan <- []byte(data)
            time.Sleep(100 * time.Millisecond)
        }
        return nil
    })
    
    // 7. 支持所有HTTP方法
    group.AnyResponse("/webhook", func(c *mvcContext.Context) any {
        method := c.RequestContext().Method()
        return map[string]any{
            "method": string(method),
            "message": "Webhook received",
        }
    })
}
```

## 🏛️ 传统控制器路由

### 手动路由注册

```go
func main() {
    app := mvc.HertzApp
    
    // 创建控制器实例
    homeController := &controllers.HomeController{}
    userController := &controllers.UserController{}
    
    // 手动注册路由
    app.RouterPrefix("/", homeController, "GetIndex", "GET:/")
    app.RouterPrefix("/about", homeController, "GetAbout", "GET:/about")
    
    // RESTful路由
    app.RouterPrefix("/user", userController, "GetUser", "GET:/:id")
    app.RouterPrefix("/user", userController, "PostUser", "POST:/")
    app.RouterPrefix("/user", userController, "PutUser", "PUT:/:id")
    app.RouterPrefix("/user", userController, "DeleteUser", "DELETE:/:id")
    
    app.Run(":8080")
}
```

### 自动路由注册

```go
func main() {
    app := mvc.HertzApp
    
    // 自动注册所有符合命名规则的方法
    app.AutoRouters(
        &controllers.HomeController{},
        &controllers.UserController{},
        &controllers.AdminController{},
    )
    
    app.Run(":8080")
}
```

## 🔧 多Handler类型参数处理

### URL参数处理

多Handler类型系统通过增强的`mvcContext.Context`提供强大的参数处理能力：

```go
// DirectHandler - 直接访问请求参数
group.GETDirect("/user/:id", func(c *mvcContext.Context) {
    // 从增强Context获取URL参数
    reqCtx := c.RequestContext()
    id := string(reqCtx.Param("id"))
    
    // 存储到增强Context
    c.SetTypedString("user_id", id)
    
    // 业务逻辑
    user := getUserById(id)
    
    // 响应
    reqCtx.JSON(200, map[string]any{
        "user": user,
        "request_id": c.GetTypedString("user_id"),
    })
})

// ResponseHandler - 自动参数提取和验证
group.GETResponse("/posts/:category/:id", func(c *mvcContext.Context) any {
    reqCtx := c.RequestContext()
    category := string(reqCtx.Param("category"))
    id := string(reqCtx.Param("id"))
    
    // 参数验证和存储
    if category == "" || id == "" {
        c.RequestContext().SetStatusCode(400)
        return map[string]any{"error": "Missing parameters"}
    }
    
    c.SetTypedString("category", category)
    c.SetTypedString("post_id", id)
    
    post := getPostByCategoryAndId(category, id)
    return map[string]any{
        "success": true,
        "data": post,
        "params": map[string]any{
            "category": c.GetTypedString("category"),
            "post_id": c.GetTypedString("post_id"),
        },
    }
})
```

### 查询参数处理

```go
// ResponseHandler - 高级查询参数处理
group.GETResponse("/users", func(c *mvcContext.Context) any {
    reqCtx := c.RequestContext()
    
    // 获取查询参数 - URL: /users?page=1&size=10&search=john&active=true
    page := getIntParam(reqCtx, "page", 1)           // 默认值1
    size := getIntParam(reqCtx, "size", 10)          // 默认值10
    search := string(reqCtx.Query("search"))         // 默认空字符串
    active := getBoolParam(reqCtx, "active", true)   // 默认true
    
    // 存储到增强Context
    c.SetTypedInt("page", page)
    c.SetTypedInt("size", size)
    c.SetTypedString("search", search)
    c.SetTypedBool("active", active)
    
    // 参数验证
    if page < 1 || size < 1 || size > 100 {
        return map[string]any{
            "error": "Invalid pagination parameters",
            "params": map[string]any{
                "page": page,
                "size": size,
            },
        }
    }
    
    // 业务逻辑
    users, total := getUserList(page, size, search, active)
    
    return map[string]any{
        "success": true,
        "data": users,
        "pagination": map[string]any{
            "page": c.GetTypedInt("page"),
            "size": c.GetTypedInt("size"),
            "total": total,
        },
        "filters": map[string]any{
            "search": c.GetTypedString("search"),
            "active": c.GetTypedBool("active"),
        },
    }
})

// 辅助函数
func getIntParam(ctx *app.RequestContext, key string, defaultValue int) int {
    if str := string(ctx.Query(key)); str != "" {
        if val, err := strconv.Atoi(str); err == nil {
            return val
        }
    }
    return defaultValue
}

func getBoolParam(ctx *app.RequestContext, key string, defaultValue bool) bool {
    if str := string(ctx.Query(key)); str != "" {
        if val, err := strconv.ParseBool(str); err == nil {
            return val
        }
    }
    return defaultValue
}
```

### 请求体数据处理

```go
// ResponseHandler - JSON数据处理
group.POSTResponse("/users", func(c *mvcContext.Context) any {
    c.SetTypedString("operation", "create_user")
    
    reqCtx := c.RequestContext()
    
    // 解析JSON数据
    var userData struct {
        Name     string `json:"name" binding:"required"`
        Email    string `json:"email" binding:"required,email"`
        Age      int    `json:"age" binding:"gte=0,lte=150"`
        IsActive bool   `json:"is_active"`
    }
    
    if err := reqCtx.BindJSON(&userData); err != nil {
        c.AddError(err)
        reqCtx.SetStatusCode(400)
        return map[string]any{
            "success": false,
            "error": "Invalid JSON data",
            "details": err.Error(),
        }
    }
    
    // 存储解析后的数据到增强Context
    c.SetTypedString("user_name", userData.Name)
    c.SetTypedString("user_email", userData.Email)
    c.SetTypedInt("user_age", userData.Age)
    c.SetTypedBool("user_active", userData.IsActive)
    
    // 业务逻辑
    user := createUser(userData)
    
    return map[string]any{
        "success": true,
        "data": user,
        "operation": c.GetTypedString("operation"),
    }
})

// StreamHandler - 文件上传处理
group.POSTStream("/upload", func(c *mvcContext.Context, dataChan chan<- []byte) error {
    c.SetTypedString("operation", "file_upload")
    
    reqCtx := c.RequestContext()
    fileHeader, err := reqCtx.FormFile("file")
    if err != nil {
        return fmt.Errorf("no file uploaded: %w", err)
    }
    
    c.SetTypedString("filename", fileHeader.Filename)
    c.SetTypedInt("filesize", int(fileHeader.Size))
    
    // 打开上传的文件
    file, err := fileHeader.Open()
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()
    
    // 流式处理文件
    buffer := make([]byte, 32*1024) // 32KB buffer
    for {
        n, err := file.Read(buffer)
        if n > 0 {
            // 处理文件块（可以是保存到磁盘、上传到云存储等）
            processedData := processFileChunk(buffer[:n])
            dataChan <- processedData
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("file read error: %w", err)
        }
    }
    
    log.Printf("File upload completed: %s (%d bytes)", 
        c.GetTypedString("filename"), c.GetTypedInt("filesize"))
    return nil
})
```

## 🏛️ 传统控制器参数处理

### URL参数

```go
// 路由: GET /user/:id
func (c *UserController) GetUser() {
    id := c.GetParam("id")          // 获取路径参数
    userInfo := getUserById(id)
    c.JSON(userInfo)
}

// 路由: GET /posts/:category/:id
func (c *PostController) GetPost() {
    category := c.GetParam("category")
    id := c.GetParam("id")
    
    post := getPostByCategoryAndId(category, id)
    c.JSON(post)
}
```

### 查询参数

```go
// URL: /users?page=1&size=10&search=john
func (c *UserController) GetUsers() {
    page := c.GetInt("page")           // 默认为0
    size := c.GetInt("size")           // 默认为0
    search := c.GetString("search")    // 默认为空字符串
    
    // 设置默认值
    if page <= 0 {
        page = 1
    }
    if size <= 0 {
        size = 10
    }
    
    users := getUserList(page, size, search)
    c.JSON(users)
}
```

### 表单参数

```go
func (c *UserController) PostCreate() {
    name := c.GetForm("name")
    email := c.GetForm("email")
    age := c.GetInt("age")
    
    // 创建用户
    user := createUser(name, email, age)
    c.JSON(user)
}
```

## RESTful路由

### 标准RESTful模式

```go
type UserController struct {
    mvc.BaseController
}

// GET /user - 获取用户列表
func (c *UserController) GetIndex() {
    users := getAllUsers()
    c.JSON(users)
}

// GET /user/:id - 获取单个用户
func (c *UserController) GetShow() {
    id := c.GetParam("id")
    user := getUserById(id)
    c.JSON(user)
}

// POST /user - 创建用户
func (c *UserController) PostCreate() {
    var user User
    c.BindJSON(&user)
    
    createdUser := createUser(user)
    c.JSON(createdUser)
}

// PUT /user/:id - 更新用户
func (c *UserController) PutUpdate() {
    id := c.GetParam("id")
    var user User
    c.BindJSON(&user)
    
    updatedUser := updateUser(id, user)
    c.JSON(updatedUser)
}

// DELETE /user/:id - 删除用户
func (c *UserController) DeleteDestroy() {
    id := c.GetParam("id")
    deleteUser(id)
    c.JSON(map[string]string{"message": "删除成功"})
}
```

### 资源路由映射

| HTTP方法 | URL路径 | 控制器方法 | 说明 |
|----------|---------|------------|------|
| GET | /user | GetIndex | 列表页面 |
| GET | /user/create | GetCreate | 创建表单页面 |
| POST | /user | PostStore | 保存新资源 |
| GET | /user/:id | GetShow | 显示资源 |
| GET | /user/:id/edit | GetEdit | 编辑表单页面 |
| PUT/PATCH | /user/:id | PutUpdate | 更新资源 |
| DELETE | /user/:id | DeleteDestroy | 删除资源 |

## 路由分组

### Namespace路由

```go
func main() {
    app := mvc.HertzApp
    
    // API v1 命名空间
    nsApiV1 := mvc.NewNamespace("/api/v1",
        mvc.NSAutoRouter(&controllers.UserController{}),
        mvc.NSAutoRouter(&controllers.PostController{}),
        
        // 手动路由
        mvc.NSRouter("/auth/login", &controllers.AuthController{}, "POST:Login"),
        mvc.NSRouter("/auth/logout", &controllers.AuthController{}, "POST:Logout"),
    )
    
    // API v2 命名空间
    nsApiV2 := mvc.NewNamespace("/api/v2",
        mvc.NSNamespace("/users",
            mvc.NSAutoRouter(&controllers.V2UserController{}),
            mvc.NSRouter("/profile", &controllers.V2UserController{}, "GET:GetProfile"),
        ),
    )
    
    // 添加命名空间
    mvc.AddNamespace(nsApiV1)
    mvc.AddNamespace(nsApiV2)
    
    app.Run(":8080")
}
```

### 嵌套命名空间

```go
nsAdmin := mvc.NewNamespace("/admin",
    // 用户管理
    mvc.NSNamespace("/users",
        mvc.NSRouter("/", &controllers.AdminUserController{}, "GET:GetIndex"),
        mvc.NSRouter("/:id", &controllers.AdminUserController{}, "GET:GetShow"),
        mvc.NSRouter("/:id", &controllers.AdminUserController{}, "PUT:PutUpdate"),
        mvc.NSRouter("/:id", &controllers.AdminUserController{}, "DELETE:DeleteDestroy"),
    ),
    
    // 文章管理
    mvc.NSNamespace("/posts",
        mvc.NSAutoRouter(&controllers.AdminPostController{}),
        
        // 文章分类管理
        mvc.NSNamespace("/categories",
            mvc.NSAutoRouter(&controllers.AdminCategoryController{}),
        ),
    ),
    
    // 系统设置
    mvc.NSNamespace("/system",
        mvc.NSRouter("/config", &controllers.AdminSystemController{}, "GET:GetConfig"),
        mvc.NSRouter("/config", &controllers.AdminSystemController{}, "POST:PostSaveConfig"),
        mvc.NSRouter("/logs", &controllers.AdminSystemController{}, "GET:GetLogs"),
    ),
)
```

## 路由中间件

### 全局中间件

```go
func main() {
    app := mvc.HertzApp
    
    // 全局中间件
    app.Use(
        middleware.RecoveryMiddleware(),    // 异常恢复
        middleware.LoggerMiddleware(),      // 日志记录
        middleware.CORSMiddleware(),        // 跨域处理
        middleware.RateLimitMiddleware(100, time.Minute), // 限流
    )
    
    app.AutoRouters(&controllers.HomeController{})
    app.Run(":8080")
}
```

### 路由组中间件

```go
nsApi := mvc.NewNamespace("/api",
    // 在命名空间级别应用中间件
    mvc.NSBefore(middleware.AuthMiddleware()),
    mvc.NSBefore(middleware.JSONMiddleware()),
    
    mvc.NSAutoRouter(&controllers.UserController{}),
    mvc.NSAutoRouter(&controllers.PostController{}),
)
```

### 控制器级中间件

```go
type UserController struct {
    mvc.BaseController
}

func (c *UserController) Prepare() {
    // 在每个方法执行前运行
    if !c.isAuthenticated() {
        c.Error(401, "需要登录")
        return
    }
}

func (c *UserController) Finish() {
    // 在每个方法执行后运行
    c.LogRequest()
}
```

## 高级路由特性

### 路由条件

```go
// 只接受JSON请求
app.RouterPrefix("/api/user", userController, "PostCreate", "POST:/",
    mvc.WithCondition(func(ctx *gin.Context) bool {
        return ctx.GetHeader("Content-Type") == "application/json"
    }),
)

// 基于用户代理的路由
app.RouterPrefix("/mobile", mobileController, "GetIndex", "GET:/",
    mvc.WithCondition(func(ctx *gin.Context) bool {
        ua := ctx.GetHeader("User-Agent")
        return strings.Contains(ua, "Mobile")
    }),
)
```

### 路由缓存

```go
func (c *PostController) GetPost() {
    id := c.GetParam("id")
    
    // 检查缓存
    cacheKey := fmt.Sprintf("post_%s", id)
    if cached := c.GetCache(cacheKey); cached != nil {
        c.JSON(cached)
        return
    }
    
    // 获取数据
    post := getPostById(id)
    
    // 设置缓存（5分钟）
    c.SetCache(cacheKey, post, 5*time.Minute)
    
    c.JSON(post)
}
```

### 路由重定向

```go
func (c *HomeController) GetOldPage() {
    c.Redirect(301, "/new-page")
}

func (c *UserController) GetProfile() {
    if !c.isAuthenticated() {
        c.Redirect(302, "/login?redirect=/profile")
        return
    }
    
    // 显示用户资料
    c.RenderHTML("user/profile.html")
}
```

## 路由调试

### 打印所有路由

```go
func main() {
    app := mvc.HertzApp
    
    // 注册路由...
    
    // 开发环境下打印路由信息
    if os.Getenv("DEBUG") == "true" {
        app.PrintRoutes()
    }
    
    app.Run(":8080")
}
```

### 路由性能监控

```go
func RoutePerformanceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        method := c.Request.Method
        path := c.Request.URL.Path
        status := c.Writer.Status()
        
        log.Printf("[%s] %s %s %d %v", 
            method, path, c.ClientIP(), status, duration)
    }
}
```

## 🚀 最佳实践

### 1. Handler类型选择指南

选择合适的Handler类型可以显著提升应用性能：

```go
// ✅ 正确的Handler类型选择
func setupRoutes() {
    group := mvc.CreateGroup("/api/v1")
    
    // 监控和健康检查 - 使用LightHandler (最小开销)
    group.GETLight("/health", func() {})
    group.GETLight("/ready", func() {})
    
    // 简单操作 - 使用SimpleHandler
    group.GETSimple("/ping", func(ctx context.Context) {
        log.Printf("Ping at %s", time.Now().Format("15:04:05"))
    })
    
    // 需要完全控制的API - 使用DirectHandler
    group.GETDirect("/metrics", func(c *mvcContext.Context) {
        // 自定义响应格式，如Prometheus metrics
        reqCtx := c.RequestContext()
        reqCtx.SetContentType("text/plain")
        reqCtx.WriteString(generateMetrics())
    })
    
    // REST API - 使用ResponseHandler (自动JSON)
    group.GETResponse("/users", getUsersHandler)
    group.POSTResponse("/users", createUserHandler)
    
    // 耗时操作 - 使用AsyncHandler
    group.POSTAsync("/reports", generateReportHandler)
    group.POSTAsync("/batch", processBatchHandler)
    
    // 大数据传输 - 使用StreamHandler
    group.GETStream("/logs", streamLogsHandler)
    group.POSTStream("/upload", uploadHandler)
    
    // 需要处理多种HTTP方法 - 使用Any*
    group.AnyResponse("/webhook", webhookHandler)
}

// ❌ 错误的Handler类型选择
// group.GETAsync("/users", getUsersHandler)     // 简单查询不需要异步
// group.GETResponse("/health", healthHandler)   // 健康检查不需要JSON响应
// group.GETStream("/ping", pingHandler)         // Ping不需要流式传输
```

### 2. 增强Context最佳实践

```go
func efficientContextUsage(c *mvcContext.Context) any {
    // ✅ 批量操作减少调用次数
    c.SetMultiple(map[string]any{
        "operation": "user_management",
        "timestamp": time.Now(),
        "request_id": generateRequestID(),
        "version": "2.0",
    })
    
    // ✅ 类型安全的操作
    c.SetTypedString("user_id", "12345")
    c.SetTypedInt("page", 1)
    c.SetTypedBool("is_admin", false)
    
    // ✅ 条件操作避免重复设置
    c.SetIfNotExists("default_limit", 20)
    c.SetIfNotExists("sort_order", "desc")
    
    // ✅ 高效数据检索
    if userID, ok := c.GetTypedString("user_id"); ok {
        // 使用类型安全的方法
        page, _ := c.GetTypedInt("page")
        isAdmin, _ := c.GetTypedBool("is_admin")
        
        return processUserRequest(userID, page, isAdmin)
    }
    
    // ✅ 错误处理
    c.AddError(fmt.Errorf("user_id is required"))
    return map[string]any{
        "success": false,
        "error": "Missing user_id parameter",
    }
}

// ❌ 低效的Context使用
func inefficientContextUsage(c *mvcContext.Context) any {
    // 一个个设置，调用次数多
    c.Set("operation", "user_management")
    c.Set("timestamp", time.Now())
    c.Set("request_id", generateRequestID())
    
    // 不使用类型安全的方法
    c.Set("page", 1)
    if page, exists := c.Get("page"); exists {
        // 需要手动类型断言，容易出错
        if pageInt, ok := page.(int); ok {
            // 使用pageInt
        }
    }
}
```

### 3. 路由组织结构

```go
// 推荐的现代路由组织方式
func setupRoutes(app *mvc.App) {
    // 1. 多Handler类型API路由 (推荐)
    setupMultiHandlerRoutes(app)
    
    // 2. 传统控制器路由 (兼容现有项目)
    setupLegacyControllerRoutes(app)
    
    // 3. 混合使用 (渐进式迁移)
    setupHybridRoutes(app)
}

func setupMultiHandlerRoutes(app *mvc.App) {
    // 公开API - 无认证
    publicAPI := mvc.CreateGroup("/api/public")
    publicAPI.GETLight("/health", healthCheck)
    publicAPI.GETResponse("/status", systemStatus)
    
    // 认证API - 需要token
    authAPI := mvc.CreateGroup("/api/v1")
    // 可以添加认证中间件
    
    authAPI.GETResponse("/profile", getProfile)
    authAPI.POSTResponse("/profile", updateProfile)
    authAPI.DELETEResponse("/profile", deleteProfile)
    
    // 管理员API - 需要管理员权限
    adminAPI := mvc.CreateGroup("/api/admin")
    adminAPI.GETResponse("/users", listUsers)
    adminAPI.POSTAsync("/batch-operations", batchOperations)
    adminAPI.GETStream("/audit-logs", streamAuditLogs)
    
    // 实时API - WebSocket和流式传输
    realtimeAPI := mvc.CreateGroup("/api/realtime")
    realtimeAPI.GETStream("/events", streamEvents)
    realtimeAPI.POSTStream("/upload", handleUpload)
}

func setupLegacyControllerRoutes(app *mvc.App) {
    // 传统控制器方式 (保持向后兼容)
    homeController := &controllers.HomeController{}
    userController := &controllers.UserController{}
    
    // 自动路由注册
    app.AutoRouters(homeController, userController)
    
    // 手动路由注册
    app.RouterPrefix("/legacy", userController, "GetLegacyAPI", "GET:/api")
}
```

### 4. 性能优化技巧

```go
// 高性能Handler实现
func optimizedHandler(c *mvcContext.Context) any {
    // 1. 预分配map容量
    response := make(map[string]any, 8) // 预估需要8个key
    
    // 2. 重用Context数据，减少重复计算
    if !c.Exists("computed_data") {
        data := expensiveComputation()
        c.Set("computed_data", data)
    }
    
    // 3. 批量获取避免多次调用
    contextData := c.GetMultiple([]string{
        "user_id", "request_id", "timestamp", "computed_data",
    })
    
    // 4. 条件设置避免不必要的操作
    c.SetIfNotExists("cached_result", func() any {
        return computeIfNeeded()
    }())
    
    return response
}

// 流式Handler的缓冲优化
func optimizedStreamHandler(c *mvcContext.Context, dataChan chan<- []byte) error {
    // 使用合适大小的缓冲区
    buffer := make([]byte, 64*1024) // 64KB
    
    // 批量写入减少系统调用
    var accumulated []byte
    batchSize := 1024 * 1024 // 1MB batch
    
    for {
        data := readData()
        if data == nil {
            break
        }
        
        accumulated = append(accumulated, data...)
        
        if len(accumulated) >= batchSize {
            dataChan <- accumulated
            accumulated = accumulated[:0] // 重用slice
        }
    }
    
    // 发送剩余数据
    if len(accumulated) > 0 {
        dataChan <- accumulated
    }
    
    return nil
}
```

### 5. 错误处理模式

```go
// 统一错误处理中间件
func errorHandlingMiddleware() {
    // 可以添加到路由组级别的错误处理
}

// ResponseHandler错误处理
func safeResponseHandler(c *mvcContext.Context) any {
    defer func() {
        if r := recover(); r != nil {
            c.AddError(fmt.Errorf("handler panic: %v", r))
            c.RequestContext().SetStatusCode(500)
            log.Printf("Handler panic: %v", r)
        }
    }()
    
    // 业务逻辑...
    if err := validateInput(); err != nil {
        c.AddError(err)
        c.RequestContext().SetStatusCode(400)
        return map[string]any{
            "success": false,
            "error": err.Error(),
            "code": "VALIDATION_ERROR",
        }
    }
    
    return map[string]any{"success": true}
}

// AsyncHandler错误处理
func safeAsyncHandler(c *mvcContext.Context) <-chan any {
    resultChan := make(chan any, 1)
    
    go func() {
        defer func() {
            if r := recover(); r != nil {
                c.AddError(fmt.Errorf("async handler panic: %v", r))
                resultChan <- map[string]any{
                    "success": false,
                    "error": "Internal processing error",
                }
            }
            close(resultChan)
        }()
        
        // 异步业务逻辑...
        result := processAsync()
        resultChan <- result
    }()
    
    return resultChan
}
```

### 6. 测试策略

```go
// Handler单元测试
func TestResponseHandler(t *testing.T) {
    // 创建测试用的增强Context
    reqCtx := &app.RequestContext{}
    enhancedCtx := mvcContext.NewContextWithContext(reqCtx, context.Background())
    defer enhancedCtx.Release()
    
    // 测试Handler
    result := getUsersHandler(enhancedCtx)
    
    // 验证结果
    assert.NotNil(t, result)
    
    // 验证Context状态
    assert.True(t, enhancedCtx.Exists("operation"))
    
    // 验证增强Context的类型安全方法
    if userCount, ok := enhancedCtx.GetTypedInt("user_count"); ok {
        assert.GreaterOrEqual(t, userCount, 0)
    }
}

// 集成测试
func TestAPIIntegration(t *testing.T) {
    // 测试多Handler类型的集成
    tests := []struct {
        method   string
        path     string
        handler  string
        expected int
    }{
        {"GET", "/api/v1/health", "LightHandler", 200},
        {"GET", "/api/v1/users", "ResponseHandler", 200},
        {"POST", "/api/v1/process", "AsyncHandler", 200},
        {"GET", "/api/v1/logs", "StreamHandler", 200},
    }
    
    for _, test := range tests {
        // 执行测试...
    }
}
```

### 2. 参数验证

```go
func (c *UserController) GetUser() {
    id := c.GetParam("id")
    
    // 验证参数
    if id == "" {
        c.Error(400, "用户ID不能为空")
        return
    }
    
    if !isValidUUID(id) {
        c.Error(400, "无效的用户ID格式")
        return
    }
    
    // 业务逻辑...
}
```

### 3. 错误处理

```go
func (c *BaseController) HandleRouteError(err error) {
    switch e := err.(type) {
    case *RouteNotFoundError:
        c.Error(404, "页面不存在")
    case *MethodNotAllowedError:
        c.Error(405, "方法不被允许")
    case *ParameterError:
        c.Error(400, e.Message)
    default:
        c.Error(500, "服务器内部错误")
    }
}
```

---

## 🎯 总结

YYHertz的双模式路由系统为您提供了极大的灵活性：

### 🚀 多Handler类型系统 (推荐新项目)
- **7种专门优化的Handler类型**
- **增强Context系统** 提供高性能并发支持
- **类型安全的数据操作** 减少运行时错误
- **对象池化** 显著提升性能

### 🏛️ 传统控制器系统 (兼容现有项目)  
- **完全向后兼容** Beego风格开发
- **自动路由注册** 快速开发
- **命名空间支持** 复杂应用结构

### 🔄 渐进式迁移
- 两种模式可以在同一应用中并存
- 支持逐步迁移到新的Handler类型系统
- 保护现有投资的同时享受新特性

**选择建议**：
- 🌟 **新项目**: 优先使用多Handler类型系统
- 🔧 **现有项目**: 可以混合使用，逐步迁移核心API
- 📈 **高性能要求**: 多Handler类型 + 增强Context

合理的路由设计是构建高性能、可维护Web应用的基础。选择适合的Handler类型，让您的应用更快、更现代！ 🎉