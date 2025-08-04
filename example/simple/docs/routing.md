# 🛣️ 路由系统

YYHertz提供了灵活而强大的路由系统，支持RESTful路由、参数绑定、路由分组等特性。

## 基础路由

### 手动路由注册

```go
func main() {
    app := mvc.HertzApp
    
    // 基础路由
    app.RouterPrefix("/", homeController, "GetIndex", "GET:/")
    app.RouterPrefix("/about", homeController, "GetAbout", "GET:/about")
    
    // 带参数的路由
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

## 路由参数

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

## 最佳实践

### 1. 路由组织结构

```go
// 推荐的路由组织方式
func setupRoutes(app *mvc.App) {
    // 静态路由
    setupStaticRoutes(app)
    
    // 前端路由
    setupWebRoutes(app)
    
    // API路由
    setupAPIRoutes(app)
    
    // 管理后台路由
    setupAdminRoutes(app)
}

func setupAPIRoutes(app *mvc.App) {
    api := mvc.NewNamespace("/api",
        mvc.NSBefore(middleware.APIMiddleware()),
        
        // V1 API
        mvc.NSNamespace("/v1", setupAPIV1Routes()),
        
        // V2 API
        mvc.NSNamespace("/v2", setupAPIV2Routes()),
    )
    
    mvc.AddNamespace(api)
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

合理的路由设计是构建可维护、可扩展Web应用的基础。记住保持路由结构清晰、语义明确！