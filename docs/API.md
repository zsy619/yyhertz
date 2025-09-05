# YYHertz API 参考手册

<div align="center">

📋 **完整的YYHertz API文档和示例** | 企业级接口开发指南

</div>

---

## 目录

- [核心API概述](#核心api概述)
- [应用程序API](#应用程序api)
- [控制器API](#控制器api)
- [上下文API](#上下文api)
- [中间件API](#中间件api)
- [路由API](#路由api)
- [ORM API](#orm-api)
- [模板API](#模板api)
- [配置API](#配置api)
- [工具包API](#工具包api)
- [错误处理](#错误处理)

---

## 🎯 核心API概述

### 快速导入
```go
import (
    "github.com/zsy619/yyhertz/framework/mvc"           // MVC核心
    "github.com/zsy619/yyhertz/framework/mvc/middleware" // 中间件
    "github.com/zsy619/yyhertz/framework/orm"           // ORM操作
    "github.com/zsy619/yyhertz/framework/config"        // 配置管理
    "github.com/zsy619/yyhertz/framework/pkg/xstring"   // 字符串工具
    "github.com/zsy619/yyhertz/framework/pkg/xdate"     // 日期工具
)
```

### 架构层次
```
Application Layer (应用层)
    ├── mvc.HertzApp                    // 应用实例
    ├── mvc.BaseController             // 控制器基类
    └── mvc.Context                    // 上下文对象

Middleware Layer (中间件层)
    ├── middleware.Recovery()          // 异常恢复
    ├── middleware.Logger()            // 访问日志
    ├── middleware.CORS()             // 跨域处理
    └── middleware.Auth()             // 身份认证

Data Layer (数据层)
    ├── orm.InitDB()                  // 数据库初始化
    ├── orm.GetDB()                   // 获取数据库连接
    └── orm.SmartSelector             // 智能ORM选择器
```

---

## 🌐 应用程序API

### mvc.HertzApp

**全局应用实例，提供框架的核心功能**

#### 基本方法

```go
// 获取应用实例
app := mvc.HertzApp

// 启动服务器
app.Run(addr string) error
app.RunTLS(addr, certFile, keyFile string) error
```

**示例：**
```go
func main() {
    app := mvc.HertzApp
    
    // HTTP服务
    app.Run(":8888")
    
    // HTTPS服务
    app.RunTLS(":8443", "cert.pem", "key.pem")
}
```

#### 路由注册

```go
// HTTP方法路由
app.GET(path string, handlers ...mvc.HandlerFunc)
app.POST(path string, handlers ...mvc.HandlerFunc)
app.PUT(path string, handlers ...mvc.HandlerFunc)
app.DELETE(path string, handlers ...mvc.HandlerFunc)
app.PATCH(path string, handlers ...mvc.HandlerFunc)
app.HEAD(path string, handlers ...mvc.HandlerFunc)
app.OPTIONS(path string, handlers ...mvc.HandlerFunc)

// 任意方法路由
app.Any(path string, handlers ...mvc.HandlerFunc)

// 自动路由注册
app.AutoRouters(controllers ...mvc.ControllerInterface)

// 静态文件服务
app.Static(relativePath, root string)
app.StaticFile(relativePath, filepath string)
```

**示例：**
```go
// 基础路由
app.GET("/", func(c *mvc.Context) {
    c.JSON(map[string]any{"message": "Hello YYHertz"})
})

// 带参数路由
app.GET("/users/:id", func(c *mvc.Context) {
    id := c.Param("id")
    c.JSON(map[string]any{"user_id": id})
})

// 自动路由
app.AutoRouters(&UserController{}, &ProductController{})

// 静态文件
app.Static("/static", "./static")
app.StaticFile("/favicon.ico", "./assets/favicon.ico")
```

#### 中间件管理

```go
// 全局中间件
app.Use(middleware ...mvc.HandlerFunc)

// 路由组中间件
group := app.Group("/api", middleware.Auth())
```

**示例：**
```go
// 全局中间件
app.Use(
    middleware.Recovery(),
    middleware.Logger(),
    middleware.CORS(),
)

// API路由组
apiGroup := app.Group("/api/v1", middleware.Auth())
{
    apiGroup.GET("/users", userHandler)
    apiGroup.POST("/users", createUserHandler)
}
```

---

## 🎮 控制器API

### mvc.BaseController

**控制器基类，提供请求处理的核心方法**

#### 请求数据获取

```go
// 路径参数
c.Param(key string) string
c.ParamInt(key string) (int, error)
c.ParamInt64(key string) (int64, error)

// 查询参数
c.Query(key string) string
c.QueryDefault(key, defaultValue string) string
c.QueryInt(key string) (int, error)
c.QueryBool(key string) (bool, error)
c.QueryArray(key string) []string

// 表单数据
c.GetForm(key string) string
c.GetFormDefault(key, defaultValue string) string
c.GetFormArray(key string) []string
c.GetFormFile(name string) (*multipart.FileHeader, error)
c.GetFormFiles(name string) ([]*multipart.FileHeader, error)

// 请求头
c.GetHeader(key string) string
c.GetHeaderDefault(key, defaultValue string) string

// Cookie
c.GetCookie(name string) (string, error)

// 请求体绑定
c.BindJSON(obj interface{}) error
c.BindXML(obj interface{}) error
c.BindForm(obj interface{}) error
c.BindQuery(obj interface{}) error
```

**示例：**
```go
type UserController struct {
    mvc.BaseController
}

func (c *UserController) GetDetail() {
    // 获取路径参数
    id, err := c.ParamInt("id")
    if err != nil {
        c.ErrorJSON(400, "Invalid user ID")
        return
    }
    
    // 获取查询参数
    fields := c.QueryDefault("fields", "id,name,email")
    
    // 获取请求头
    token := c.GetHeader("Authorization")
    
    user := getUserByID(id)
    c.JSON(user)
}

func (c *UserController) PostCreate() {
    var user User
    
    // 绑定JSON数据
    if err := c.BindJSON(&user); err != nil {
        c.ErrorJSON(400, "Invalid JSON data")
        return
    }
    
    // 创建用户
    if err := createUser(&user); err != nil {
        c.ErrorJSON(500, "Failed to create user")
        return
    }
    
    c.JSON(user)
}
```

#### 响应数据输出

```go
// JSON响应
c.JSON(data interface{})
c.JSONIndented(data interface{}, indent string)
c.SuccessJSON(data interface{})
c.ErrorJSON(code int, message string)

// HTML响应
c.HTML(templateName string, data interface{})
c.HTMLWithLayout(templateName, layoutName string, data interface{})

// 文本响应
c.String(format string, values ...interface{})
c.Text(text string)

// 文件响应
c.File(filepath string)
c.FileAttachment(filepath, filename string)
c.Data(contentType string, data []byte)

// 重定向
c.Redirect(code int, location string)
c.RedirectTo(location string)

// 状态码设置
c.SetStatus(code int)
c.SetHeader(key, value string)
c.SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
```

**示例：**
```go
func (c *UserController) GetList() {
    users := getAllUsers()
    
    // 成功响应
    c.SuccessJSON(map[string]any{
        "users": users,
        "total": len(users),
    })
}

func (c *UserController) GetProfile() {
    user := getCurrentUser()
    
    // HTML响应
    c.HTMLWithLayout("user/profile.html", "layouts/main.html", map[string]any{
        "user": user,
        "title": "用户资料",
    })
}

func (c *FileController) GetDownload() {
    filepath := c.Query("file")
    
    // 文件下载
    c.FileAttachment(filepath, "download.pdf")
}
```

#### 控制器生命周期

```go
// 控制器接口
type ControllerInterface interface {
    Prepare()                    // 前置处理
    Finish()                     // 后置处理
}

// 实现示例
func (c *UserController) Prepare() {
    // 身份验证
    if !c.IsAuthenticated() {
        c.ErrorJSON(401, "Unauthorized")
        c.StopRun()
        return
    }
}

func (c *UserController) Finish() {
    // 记录访问日志
    c.LogAccess()
}
```

---

## 📦 上下文API

### mvc.Context

**请求上下文对象，封装了请求和响应的所有信息**

#### 请求信息

```go
// 基本信息
c.Method() string                    // HTTP方法
c.Path() string                      // 请求路径
c.URL() *url.URL                     // 完整URL
c.Host() string                      // 主机名
c.RemoteAddr() string                // 客户端地址
c.UserAgent() string                 // User-Agent
c.Referer() string                   // 来源页面

// 请求体
c.Body() []byte                      // 原始请求体
c.ContentLength() int64              // 内容长度
c.ContentType() string               // 内容类型
```

**示例：**
```go
func requestInfoHandler(c *mvc.Context) {
    info := map[string]any{
        "method":       c.Method(),
        "path":         c.Path(),
        "host":         c.Host(),
        "remote_addr":  c.RemoteAddr(),
        "user_agent":   c.UserAgent(),
        "content_type": c.ContentType(),
    }
    
    c.JSON(info)
}
```

#### 状态管理

```go
// 上下文存储
c.Set(key string, value interface{})
c.Get(key string) (interface{}, bool)
c.MustGet(key string) interface{}

// 中断处理
c.Abort()
c.AbortWithStatus(code int)
c.AbortWithJSON(code int, data interface{})

// 处理状态
c.IsAborted() bool
c.Next()
```

**示例：**
```go
// 中间件中设置用户信息
func AuthMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        token := c.GetHeader("Authorization")
        user := validateToken(token)
        
        if user == nil {
            c.AbortWithJSON(401, map[string]string{
                "error": "Unauthorized",
            })
            return
        }
        
        c.Set("user", user)
        c.Next()
    }
}

// 控制器中获取用户信息
func (ctrl *UserController) GetProfile() {
    user := ctrl.MustGet("user").(*User)
    ctrl.JSON(user)
}
```

---

## 🔧 中间件API

### 内置中间件

#### Recovery中间件
```go
// 异常恢复中间件
middleware.Recovery() mvc.HandlerFunc

// 自定义恢复处理
middleware.RecoveryWithWriter(out io.Writer, recovery ...RecoveryFunc) mvc.HandlerFunc
```

**示例：**
```go
app.Use(middleware.Recovery())

// 自定义恢复处理
app.Use(middleware.RecoveryWithWriter(os.Stdout, func(c *mvc.Context, recovered interface{}) {
    if err, ok := recovered.(string); ok {
        c.JSON(500, map[string]string{
            "error": fmt.Sprintf("Internal server error: %s", err),
        })
    }
    c.AbortWithStatus(500)
}))
```

#### Logger中间件
```go
// 访问日志中间件
middleware.Logger() mvc.HandlerFunc

// 自定义日志格式
middleware.LoggerWithFormatter(f LogFormatter) mvc.HandlerFunc
middleware.LoggerWithWriter(out io.Writer) mvc.HandlerFunc
```

**示例：**
```go
app.Use(middleware.Logger())

// 自定义日志格式
app.Use(middleware.LoggerWithFormatter(func(param LogFormatterParams) string {
    return fmt.Sprintf("%s - [%s] \"%s %s\" %d %d \"%s\" \"%s\" %v\n",
        param.ClientIP,
        param.TimeStamp.Format("2006/01/02 - 15:04:05"),
        param.Method,
        param.Path,
        param.StatusCode,
        param.BodySize,
        param.Referer,
        param.UserAgent,
        param.Latency,
    )
}))
```

#### CORS中间件
```go
// 跨域中间件
middleware.CORS() mvc.HandlerFunc

// 自定义CORS配置
middleware.CORSWithConfig(config CORSConfig) mvc.HandlerFunc
```

**示例：**
```go
// 默认配置
app.Use(middleware.CORS())

// 自定义配置
app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins:     []string{"https://example.com", "https://api.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))
```

#### 认证中间件
```go
// JWT认证
middleware.JWTAuth(config JWTConfig) mvc.HandlerFunc

// Basic认证
middleware.BasicAuth(accounts map[string]string) mvc.HandlerFunc

// API Key认证
middleware.APIKeyAuth(config APIKeyConfig) mvc.HandlerFunc
```

**示例：**
```go
// JWT认证
app.Use(middleware.JWTAuth(middleware.JWTConfig{
    SigningKey:    []byte("your-secret-key"),
    TokenLookup:   "header:Authorization",
    AuthScheme:    "Bearer",
    Skipper: func(c *mvc.Context) bool {
        return c.Path() == "/login" || c.Path() == "/register"
    },
}))

// Basic认证
app.Use(middleware.BasicAuth(map[string]string{
    "admin": "password",
    "user":  "userpass",
}))
```

#### 限流中间件
```go
// 基于IP的限流
middleware.RateLimit(rate int, window time.Duration) mvc.HandlerFunc

// 自定义限流配置
middleware.RateLimitWithConfig(config RateLimitConfig) mvc.HandlerFunc
```

**示例：**
```go
// 每分钟100次请求
app.Use(middleware.RateLimit(100, time.Minute))

// 自定义限流
app.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
    Rate:   1000,
    Window: time.Hour,
    KeyFunc: func(c *mvc.Context) string {
        return c.GetHeader("X-API-Key")
    },
    ErrorHandler: func(c *mvc.Context) {
        c.JSON(429, map[string]string{
            "error": "Rate limit exceeded",
        })
    },
}))
```

### 自定义中间件

```go
// 中间件函数签名
type HandlerFunc func(*Context)

// 自定义中间件示例
func CustomMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 前置处理
        start := time.Now()
        
        // 处理请求
        c.Next()
        
        // 后置处理
        duration := time.Since(start)
        log.Printf("Request %s %s took %v", c.Method(), c.Path(), duration)
    }
}

// 使用中间件
app.Use(CustomMiddleware())
```

---

## 🛣️ 路由API

### 路由注册

```go
// 基础路由
app.Handle(method, path string, handlers ...mvc.HandlerFunc)
app.GET(path string, handlers ...mvc.HandlerFunc)
app.POST(path string, handlers ...mvc.HandlerFunc)
// ... 其他HTTP方法

// 路由组
group := app.Group(relativePath string, handlers ...mvc.HandlerFunc)
```

**示例：**
```go
// 基础路由
app.GET("/", homeHandler)
app.POST("/users", createUserHandler)

// RESTful路由
userRoutes := app.Group("/users")
{
    userRoutes.GET("", getUserList)           // GET /users
    userRoutes.GET("/:id", getUserDetail)     // GET /users/123
    userRoutes.POST("", createUser)           // POST /users
    userRoutes.PUT("/:id", updateUser)        // PUT /users/123
    userRoutes.DELETE("/:id", deleteUser)     // DELETE /users/123
}
```

### 路径参数

```go
// 命名参数
app.GET("/users/:id", handler)              // 匹配 /users/123
app.GET("/users/:id/posts/:postId", handler) // 匹配 /users/123/posts/456

// 通配符
app.GET("/static/*filepath", handler)        // 匹配 /static/css/style.css
```

**示例：**
```go
func userDetailHandler(c *mvc.Context) {
    userID := c.Param("id")
    
    user, err := getUserByID(userID)
    if err != nil {
        c.JSON(404, map[string]string{"error": "User not found"})
        return
    }
    
    c.JSON(user)
}
```

### 自动路由

```go
// 控制器自动路由
app.AutoRouters(controllers ...mvc.ControllerInterface)

// 命名空间路由
ns := mvc.NewNamespace("/api/v1")
ns.AutoRouter(&UserController{})
ns.Router("/auth/login", &AuthController{}, "POST:Login")
mvc.AddNamespace(ns)
```

**示例：**
```go
type UserController struct {
    mvc.BaseController
}

func (c *UserController) GetIndex() {}     // GET /user/index
func (c *UserController) PostCreate() {}   // POST /user/create
func (c *UserController) PutUpdate() {}    // PUT /user/update
func (c *UserController) DeleteRemove() {} // DELETE /user/remove

// 注册自动路由
app.AutoRouters(&UserController{})
```

---

## 🗄️ ORM API

### 数据库初始化

```go
// 初始化数据库连接
orm.InitDB(driver, dsn string) (*gorm.DB, error)
orm.InitDBWithConfig(config *orm.Config) (*gorm.DB, error)

// 获取数据库实例
orm.GetDB() *gorm.DB

// 多数据库连接
orm.InitNamedDB(name, driver, dsn string) (*gorm.DB, error)
orm.GetNamedDB(name string) *gorm.DB
```

**示例：**
```go
// 初始化数据库
db, err := orm.InitDB("mysql", "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local")
if err != nil {
    panic("Failed to connect to database")
}

// 自动迁移
db.AutoMigrate(&User{}, &Product{}, &Order{})

// 在控制器中使用
func (c *UserController) GetList() {
    var users []User
    db := orm.GetDB()
    db.Find(&users)
    c.JSON(users)
}
```

### 智能ORM选择器

```go
// 创建智能选择器
selector := orm.NewSmartSelector()

// CRUD操作（自动选择最优引擎）
selector.Create(value interface{}) error
selector.Find(dest interface{}, conditions ...interface{}) error
selector.Update(value interface{}) error
selector.Delete(value interface{}) error

// 复杂查询（使用MyBatis引擎）
selector.ExecuteComplexQuery(queryID string, params map[string]interface{}) ([]map[string]interface{}, error)
```

**示例：**
```go
func (c *UserController) PostCreate() {
    var user User
    c.BindJSON(&user)
    
    selector := orm.NewSmartSelector()
    
    // 简单创建操作 - 自动选择GORM引擎
    if err := selector.Create(&user); err != nil {
        c.ErrorJSON(500, "Failed to create user")
        return
    }
    
    c.SuccessJSON(user)
}

func (c *ReportController) GetUserStats() {
    selector := orm.NewSmartSelector()
    
    // 复杂统计查询 - 自动选择MyBatis引擎
    stats, err := selector.ExecuteComplexQuery("userStatsQuery", map[string]interface{}{
        "startDate": c.Query("start_date"),
        "endDate":   c.Query("end_date"),
        "region":    c.Query("region"),
    })
    
    if err != nil {
        c.ErrorJSON(500, "Query failed")
        return
    }
    
    c.SuccessJSON(stats)
}
```

### GORM集成

```go
// 标准GORM操作
db := orm.GetDB()

// 查询
db.First(&user, 1)                          // 根据主键查询
db.Find(&users)                             // 查询所有
db.Where("name = ?", "john").First(&user)   // 条件查询
db.Where("age > ?", 18).Find(&users)        // 批量查询

// 创建
db.Create(&user)                            // 创建单条
db.CreateInBatches(users, 100)              // 批量创建

// 更新
db.Save(&user)                              // 更新所有字段
db.Model(&user).Updates(User{Name: "hello", Age: 18})  // 更新指定字段

// 删除
db.Delete(&user, 1)                         // 软删除
db.Unscoped().Delete(&user)                 // 硬删除
```

### 事务处理

```go
// 手动事务
tx := orm.GetDB().Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&profile).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()

// 自动事务
err := orm.GetDB().Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }
    
    return nil
})
```

---

## 🎨 模板API

### 模板引擎

```go
// 初始化模板引擎
engine := view.NewTemplateEngine(config view.Config)

// 渲染模板
engine.Render(templateName string, data interface{}) (string, error)

// 设置模板函数
engine.SetFuncMap(funcMap template.FuncMap)

// 解析模板
engine.ParseTemplates(pattern string) error
```

**示例：**
```go
// 配置模板引擎
config := view.Config{
    TemplateDir:    "views",
    TemplateExt:    ".html",
    EnableCache:    true,
    EnableReload:   false,
}

engine := view.NewTemplateEngine(config)

// 在控制器中使用
func (c *HomeController) GetIndex() {
    data := map[string]interface{}{
        "title": "首页",
        "user":  getCurrentUser(),
        "posts": getLatestPosts(),
    }
    
    c.HTML("home/index.html", data)
}
```

### 模板函数

YYHertz提供了150+个模板函数，涵盖字符串处理、日期格式化、数学运算等：

```go
// 字符串处理函数
{{substr "hello world" 0 5}}        // "hello"
{{upper "hello"}}                   // "HELLO"
{{lower "WORLD"}}                   // "world"

// 日期时间函数
{{dateformat .CreatedAt "2006-01-02"}}  // "2024-09-05"
{{timeago .UpdatedAt}}                   // "2小时前"
{{now | dateformat "15:04:05"}}         // "14:30:25"

// 数学函数
{{add 10 20}}                       // 30
{{sub 50 20}}                       // 30
{{mul 6 7}}                         // 42
{{div 100 4}}                       // 25

// 逻辑判断
{{if equal .Status "active"}}活跃{{end}}
{{if gt .Age 18}}成年人{{else}}未成年{{end}}

// 模板包含
{{include "partials/header.html" .}}
{{templatefunc "components/user-card.html" .User}}
```

---

## ⚙️ 配置API

### 配置管理

```go
// 加载配置文件
config.LoadConfig(filename string) error
config.LoadConfigFromBytes(data []byte) error

// 获取配置值
config.GetString(key string) string
config.GetInt(key string) int
config.GetInt64(key string) int64
config.GetFloat64(key string) float64
config.GetBool(key string) bool
config.GetStringSlice(key string) []string

// 默认值
config.GetStringDefault(key, defaultValue string) string
config.GetIntDefault(key string, defaultValue int) int
```

**示例：**
```go
// 加载配置
config.LoadConfig("config/app.conf")

// 获取配置
appName := config.GetString("app.name")
port := config.GetInt("app.port")
debug := config.GetBool("app.debug")

// 数据库配置
dbHost := config.GetString("db.host")
dbPort := config.GetInt("db.port")
dbName := config.GetString("db.name")

dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
    config.GetString("db.user"),
    config.GetString("db.password"),
    dbHost,
    dbPort,
    dbName,
)

orm.InitDB("mysql", dsn)
```

### 应用配置示例

```ini
# config/app.conf

# 应用基础配置
app.name = YYHertz App
app.version = 2.0.0
app.debug = true
app.host = localhost
app.port = 8888

# 数据库配置
db.driver = mysql
db.host = localhost
db.port = 3306
db.name = app_db
db.user = root
db.password = password
db.charset = utf8mb4
db.max_idle_conns = 10
db.max_open_conns = 100

# Redis配置
redis.host = localhost
redis.port = 6379
redis.password = 
redis.db = 0
redis.pool_size = 10

# 日志配置
log.level = info
log.file = logs/app.log
log.max_size = 100
log.max_backups = 10
log.max_age = 30

# JWT配置
jwt.secret = your-jwt-secret-key
jwt.expire_hours = 24
jwt.issuer = YYHertz

# 上传配置
upload.path = uploads
upload.max_size = 10485760
upload.allowed_exts = .jpg,.jpeg,.png,.gif,.pdf,.doc,.docx
```

---

## 🧰 工具包API

### 字符串工具 (xstring)

```go
import "github.com/zsy619/yyhertz/framework/pkg/xstring"

// 字符串处理
xstring.CapitalizeFirst(s string) string        // 首字母大写
xstring.Substr(s string, start, length int) string  // 子字符串
xstring.Contains(s, substr string) bool         // 包含检查
xstring.HasPrefix(s, prefix string) bool        // 前缀检查
xstring.HasSuffix(s, suffix string) bool        // 后缀检查
xstring.TrimSlash(s string) string              // 移除首尾斜杠
```

**示例：**
```go
// 字符串处理
text := "hello world"
result := xstring.CapitalizeFirst(text)  // "Hello world"

// 中文支持的子字符串
content := "这是一个测试字符串"
part := xstring.Substr(content, 0, 5)    // "这是一个测"

// URL路径处理  
path := "/api/users/"
cleaned := xstring.TrimSlash(path)       // "api/users"
```

### 日期工具 (xdate)

```go
import "github.com/zsy619/yyhertz/framework/pkg/xdate"

// 日期格式化
xdate.Format(t time.Time, layout string) string
xdate.TimeAgo(t time.Time) string               // 相对时间
xdate.TimeSince(t time.Time) string             // 距离现在
xdate.TimeUntil(t time.Time) string             // 到某时间

// 日期解析
xdate.Parse(value string) (time.Time, error)
xdate.ParseInLocation(value, layout, location string) (time.Time, error)
```

**示例：**
```go
now := time.Now()

// 格式化
formatted := xdate.Format(now, "Y-m-d H:i:s")   // "2024-09-05 15:30:25"

// 相对时间
past := now.Add(-2 * time.Hour)
ago := xdate.TimeAgo(past)                      // "2小时前"

// 解析日期
date, err := xdate.Parse("2024-09-05 15:30:25")
```

### 数学工具 (xmath)

```go
import "github.com/zsy619/yyhertz/framework/pkg/xmath"

// 基本运算
xmath.Add(a, b interface{}) interface{}         // 加法
xmath.Sub(a, b interface{}) interface{}         // 减法  
xmath.Mul(a, b interface{}) interface{}         // 乘法
xmath.Div(a, b interface{}) interface{}         // 除法
xmath.Mod(a, b interface{}) interface{}         // 取模

// 类型转换
xmath.ToInt(value interface{}) int
xmath.ToInt64(value interface{}) int64
xmath.ToFloat64(value interface{}) float64
xmath.ToBool(value interface{}) bool

// 数学函数
xmath.Round(f float64, precision int) float64   // 四舍五入
xmath.Ceil(f float64) float64                   // 向上取整
xmath.Floor(f float64) float64                  // 向下取整
xmath.Abs(n interface{}) interface{}            // 绝对值
```

**示例：**
```go
// 动态类型运算
result := xmath.Add(10, 20)        // 30
sum := xmath.Add(10.5, 20.3)       // 30.8

// 类型转换
str := "123"
num := xmath.ToInt(str)            // 123

// 精度控制
price := 123.456
rounded := xmath.Round(price, 2)    // 123.46
```

### 加密工具 (xcrypto)

```go
import "github.com/zsy619/yyhertz/framework/pkg/xcrypto"

// 哈希算法
xcrypto.MD5(data string) string
xcrypto.SHA1(data string) string
xcrypto.SHA256(data string) string

// 对称加密
xcrypto.AESEncrypt(plaintext, key []byte) ([]byte, error)
xcrypto.AESDecrypt(ciphertext, key []byte) ([]byte, error)

// Base64编码
xcrypto.Base64Encode(data []byte) string
xcrypto.Base64Decode(encoded string) ([]byte, error)

// 密码哈希
xcrypto.HashPassword(password string) (string, error)
xcrypto.CheckPassword(hashedPassword, password string) bool
```

**示例：**
```go
// 密码哈希
password := "userpassword"
hashed, err := xcrypto.HashPassword(password)

// 验证密码
isValid := xcrypto.CheckPassword(hashed, password)

// 数据加密
key := []byte("your-32-byte-key-here-for-aes!!")
plaintext := []byte("sensitive data")
encrypted, err := xcrypto.AESEncrypt(plaintext, key)

// 生成签名
signature := xcrypto.SHA256("data" + "secret_key")
```

---

## ❌ 错误处理

### 错误类型

```go
// 框架错误类型
type Error struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

// HTTP错误
type HTTPError struct {
    Code     int         `json:"code"`
    Message  string      `json:"message"`
    Internal error       `json:"-"`
}
```

### 错误处理中间件

```go
func ErrorHandler() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        c.Next()
        
        // 检查是否有错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            switch e := err.Err.(type) {
            case *HTTPError:
                c.JSON(e.Code, map[string]interface{}{
                    "error": e.Message,
                    "code":  e.Code,
                })
            default:
                c.JSON(500, map[string]interface{}{
                    "error": "Internal Server Error",
                    "code":  500,
                })
            }
        }
    }
}
```

### 统一错误响应

```go
// 错误响应结构
type ErrorResponse struct {
    Success   bool        `json:"success"`
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    Timestamp int64       `json:"timestamp"`
    RequestID string      `json:"request_id,omitempty"`
}

// 控制器错误处理
func (c *BaseController) HandleError(err error, code int) {
    response := ErrorResponse{
        Success:   false,
        Code:      code,
        Message:   err.Error(),
        Timestamp: time.Now().Unix(),
        RequestID: c.GetHeader("X-Request-ID"),
    }
    
    c.SetStatus(code)
    c.JSON(response)
    c.Abort()
}

// 使用示例
func (c *UserController) GetDetail() {
    id, err := c.ParamInt("id")
    if err != nil {
        c.HandleError(errors.New("Invalid user ID"), 400)
        return
    }
    
    user, err := getUserByID(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.HandleError(err, 404)
        } else {
            c.HandleError(err, 500)
        }
        return
    }
    
    c.SuccessJSON(user)
}
```

### 错误码定义

```go
const (
    // 成功
    CodeSuccess = 200
    
    // 客户端错误 4xx
    CodeBadRequest          = 400
    CodeUnauthorized        = 401
    CodeForbidden          = 403
    CodeNotFound           = 404
    CodeMethodNotAllowed    = 405
    CodeValidationError     = 422
    CodeTooManyRequests     = 429
    
    // 服务端错误 5xx
    CodeInternalError       = 500
    CodeServiceUnavailable  = 503
    CodeDatabaseError       = 1001
    CodeCacheError         = 1002
    CodeExternalAPIError   = 1003
)

var errorMessages = map[int]string{
    CodeBadRequest:          "请求参数错误",
    CodeUnauthorized:        "未授权访问",
    CodeForbidden:          "访问被禁止", 
    CodeNotFound:           "资源不存在",
    CodeValidationError:    "数据验证失败",
    CodeInternalError:      "内部服务器错误",
    CodeDatabaseError:      "数据库操作失败",
}

func GetErrorMessage(code int) string {
    if msg, exists := errorMessages[code]; exists {
        return msg
    }
    return "未知错误"
}
```

---

## 📋 API使用最佳实践

### 1. 控制器设计
```go
// ✅ 好的实践
type UserController struct {
    mvc.BaseController
    userService *service.UserService
}

func (c *UserController) Prepare() {
    c.userService = service.NewUserService()
}

func (c *UserController) GetList() {
    // 参数验证
    page, _ := c.QueryInt("page")
    if page <= 0 {
        page = 1
    }
    
    // 业务逻辑
    users, total, err := c.userService.GetUserList(page, 20)
    if err != nil {
        c.HandleError(err, 500)
        return
    }
    
    // 统一响应格式
    c.SuccessJSON(map[string]interface{}{
        "users": users,
        "total": total,
        "page":  page,
    })
}
```

### 2. 中间件使用
```go
// ✅ 中间件顺序很重要
app.Use(
    middleware.Recovery(),      // 第一个：异常恢复
    middleware.Logger(),        // 第二个：日志记录
    middleware.CORS(),          // 第三个：跨域处理
    middleware.RateLimit(),     // 第四个：限流
    middleware.Auth(),          // 最后：身份验证
)
```

### 3. 错误处理
```go
// ✅ 统一错误处理
func (c *BaseController) HandleBusinessError(err error) {
    switch {
    case errors.Is(err, ErrUserNotFound):
        c.ErrorJSON(404, "用户不存在")
    case errors.Is(err, ErrInvalidCredentials):
        c.ErrorJSON(401, "用户名或密码错误")
    case errors.Is(err, ErrValidationFailed):
        c.ErrorJSON(422, err.Error())
    default:
        c.ErrorJSON(500, "服务内部错误")
    }
}
```

### 4. 数据库操作
```go
// ✅ 使用智能选择器
func (s *UserService) GetUserStats() ([]UserStat, error) {
    selector := orm.NewSmartSelector()
    
    // 复杂统计查询自动选择MyBatis
    return selector.ExecuteComplexQuery("getUserStats", nil)
}

// ✅ 使用事务
func (s *UserService) CreateUserWithProfile(user *User, profile *Profile) error {
    return orm.GetDB().Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        profile.UserID = user.ID
        return tx.Create(profile).Error
    })
}
```

---

<div align="center">

**📖 这份API参考手册涵盖了YYHertz框架的核心功能**  
**更多详细示例请参考：[完整示例项目](../example/)**

</div>