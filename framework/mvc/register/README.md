# ControllerRegister 实现

## 概述

这个实现提供了完整的 ControllerRegister 功能，包括：

- 完整的控制器注册和路由匹配
- 生命周期管理（Init、Prepare、Finish）
- 过滤器和中间件支持
- RESTful路由支持
- 参数路由和通配符路由
- 控制器实例池化
- 性能监控和统计

## 核心组件

### 1. ControllerRegister

主要的控制器注册器，负责：
- 路由注册和管理
- 请求分发和处理
- 过滤器执行
- 控制器生命周期管理

```go
// 创建注册器
cr := register.NewControllerRegister()

// 注册控制器
cr.Add("/user", &UserController{})
cr.AddAuto(&ProductController{})
cr.AddAutoPrefix("/api", &APIController{})
```

### 2. 路由类型

#### 固定路由
```go
cr.Add("/user", &UserController{})
```

#### 参数路由
```go
cr.Add("/user/:id", &UserController{})        // /user/123
cr.Add("/posts/:category/:id", controller)    // /posts/tech/123
```

#### 通配符路由
```go
cr.Add("/files/*", &FileController{})         // /files/docs/readme.txt
```

#### 正则路由
```go
cr.Add("/user/:id(\\d+)", &UserController{})  // 只匹配数字ID
```

### 3. HTTP方法支持

支持所有标准HTTP方法：

```go
// 控制器方法命名规则
func (c *UserController) GetIndex()    {} // GET /user/
func (c *UserController) GetShow()     {} // GET /user/:id
func (c *UserController) PostCreate()  {} // POST /user/
func (c *UserController) PutUpdate()   {} // PUT /user/:id
func (c *UserController) DeleteRemove(){} // DELETE /user/:id
```

### 4. 函数路由

除了控制器路由，还支持函数路由：

```go
cr.Get("/ping", func(ctx *contextenhanced.Context) {
    ctx.Output.JSON(map[string]interface{}{
        "message": "pong",
    }, false, true)
})

cr.Post("/echo", echoHandler)
cr.Any("/health", healthHandler)
```

## 过滤器系统

### 过滤器位置

```go
const (
    BeforeStatic = iota // 静态文件之前
    BeforeRouter        // 路由之前
    BeforeExec          // 执行之前
    AfterExec           // 执行之后
    FinishRouter        // 路由完成
)
```

### 内置过滤器

1. **LoggingFilter** - 请求日志记录
2. **CORSFilter** - CORS跨域处理
3. **AuthFilter** - 认证过滤器
4. **SecurityFilter** - 安全过滤器
5. **RateLimitFilter** - 限流过滤器
6. **CompressFilter** - 压缩过滤器

### 使用过滤器

```go
// 全局过滤器
cr.InsertFilter("/*", register.BeforeRouter, register.LoggingFilter)

// 路径特定过滤器
cr.InsertFilter("/api/*", register.BeforeExec, register.AuthFilter)

// 自定义过滤器
customFilter := func(ctx *contextenhanced.Context, chain *register.FilterChain) {
    // 前置处理
    chain.Next(ctx)
    // 后置处理
}
cr.InsertFilter("/admin/*", register.BeforeExec, customFilter)
```

## 控制器生命周期

### 1. 控制器接口

```go
type IController interface {
    Init(ct *contextenhanced.Context, controllerName, actionName string, app interface{})
    Prepare()
    Finish()
    GetControllerName() string
    GetActionName() string
}
```

### 2. 生命周期流程

1. **创建控制器实例** - 从对象池获取或新建
2. **Init()** - 初始化控制器，设置Context
3. **Prepare()** - 预处理，可以进行权限检查等
4. **业务方法执行** - 执行具体的GetIndex、PostCreate等方法
5. **Finish()** - 后置处理，清理资源等

### 3. 基础控制器

```go
type UserController struct {
    core.BaseController
}

func (c *UserController) Init(ct *contextenhanced.Context, controllerName, actionName string, app interface{}) {
    c.BaseController.Init(ct, controllerName, actionName, app)
    // 自定义初始化逻辑
}

func (c *UserController) Prepare() {
    // 权限检查
    if !c.checkAuth() {
        c.Context.Output.SetStatus(401)
        return
    }
}

func (c *UserController) GetIndex() {
    // 业务逻辑
    users := c.getUserList()
    c.Context.Output.JSON(users, true, true)
}
```

## Context 使用

### 输入处理 (Input)

```go
// 获取请求参数
id := ctx.Input.Param("id")              // 路径参数
page := ctx.Input.Query("page")          // 查询参数
token := ctx.Input.Header("Authorization") // 请求头
cookie := ctx.Input.Cookie("session_id") // Cookie

// 获取请求体
var data map[string]interface{}
ctx.Input.JSON(&data)

// 获取客户端信息
ip := ctx.Input.IP()
userAgent := ctx.Input.UserAgent()
isAjax := ctx.Input.IsAjax()
```

### 输出处理 (Output)

```go
// JSON响应
ctx.Output.JSON(data, true, true) // data, hasIndent, encoding

// 设置状态码
ctx.Output.SetStatus(201)

// 设置响应头
ctx.Output.Header("Content-Type", "application/json")

// 设置Cookie
ctx.Output.Cookie("session_id", "abc123", 3600, "/", "", false, true)

// 文件下载
ctx.Output.Download("file.pdf", "report.pdf")
```

## 性能特性

### 1. 对象池化

控制器实例使用对象池来减少GC压力：

```go
// 内部实现了sync.Pool
cr.pool.New = func() interface{} {
    return make(map[string]interface{})
}
```

### 2. 路由缓存

- 固定路由使用map快速查找
- 正则路由按长度排序优先匹配
- 编译时预处理路由表达式

### 3. 并发安全

- 读写锁保护路由表
- 线程安全的过滤器执行
- 无锁的请求处理路径

## 监控和调试

### 统计信息

```go
// 获取请求计数
count := cr.GetRequestCount()

// 获取路由数量
routeCount := cr.GetRouteCount()

// 列出所有路由
routes := cr.ListRoutes()
```

### 调试端点

```go
// 添加调试路由
h.GET("/debug/routes", func(ctx context.Context, c *app.RequestContext) {
    c.JSON(200, map[string]interface{}{
        "routes":        cr.ListRoutes(),
        "request_count": cr.GetRequestCount(),
        "route_count":   cr.GetRouteCount(),
    })
})
```

## 集成示例

### 与Hertz集成

```go
func main() {
    // 创建Hertz应用
    h := server.Default()
    
    // 创建ControllerRegister
    cr := register.NewControllerRegister()
    
    // 注册路由
    cr.Add("/user", &UserController{})
    
    // 集成到Hertz
    h.Any("/*path", func(ctx context.Context, c *app.RequestContext) {
        cr.ServeHTTP(ctx, c)
    })
    
    h.Spin()
}
```

### 批量注册

```go
func RegisterControllers(cr *register.ControllerRegister) {
    controllers := map[string]core.IController{
        "/user":    &UserController{},
        "/product": &ProductController{},
        "/order":   &OrderController{},
    }
    
    for path, controller := range controllers {
        cr.Add(path, controller)
    }
}
```

## 最佳实践

### 1. 控制器设计

- 继承 `core.BaseController`
- 实现必要的生命周期方法
- 方法命名遵循HTTP动词前缀规则
- 使用Context进行输入输出处理

### 2. 路由设计

- 使用RESTful路由风格
- 合理使用参数路由和通配符路由
- 避免过于复杂的正则表达式路由

### 3. 过滤器使用

- 全局过滤器处理通用逻辑
- 路径特定过滤器处理特殊需求
- 过滤器链要考虑执行顺序

### 4. 性能优化

- 利用对象池减少内存分配
- 合理设计路由层次结构
- 使用过滤器进行缓存和限流

## 错误处理

### 内置错误处理

- 404 Not Found - 路由不匹配
- 405 Method Not Allowed - HTTP方法不支持
- 500 Internal Server Error - 服务器内部错误

### 自定义错误处理

```go
// 在过滤器中处理错误
func ErrorHandlerFilter(ctx *contextenhanced.Context, chain *register.FilterChain) {
    defer func() {
        if err := recover(); err != nil {
            ctx.Output.SetStatus(500)
            ctx.Output.JSON(map[string]interface{}{
                "error": "Internal Server Error",
                "code":  500,
            }, false, true)
        }
    }()
    
    chain.Next(ctx)
}
```

## 扩展能力

### 自定义ControllerRegister

```go
type CustomControllerRegister struct {
    *register.ControllerRegister
    customFeatures map[string]interface{}
}

func (ccr *CustomControllerRegister) AddCustomRoute(pattern string, handler interface{}) {
    // 自定义路由处理逻辑
}
```

### 中间件适配

```go
// 适配第三方中间件
func AdaptMiddleware(middleware ThirdPartyMiddleware) register.FilterFunc {
    return func(ctx *contextenhanced.Context, chain *register.FilterChain) {
        // 适配逻辑
        middleware.Process(ctx)
        chain.Next(ctx)
    }
}
```

这个实现提供了与Beego完全兼容的ControllerRegister功能，同时结合了Hertz的高性能特性，为开发者提供了熟悉且强大的Web开发体验。