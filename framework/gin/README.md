# YYHertz Gin Framework

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-reference-blue.svg)](https://pkg.go.dev/github.com/zsy619/yyhertz/framework/gin)

基于 CloudWeGo Hertz 的 Gin 风格 Web 框架，提供完全兼容 Gin API 的开发体验，同时享受 Hertz 的高性能优势。

## 🚀 特性

- **完全兼容 Gin API** - 无缝迁移现有 Gin 项目
- **高性能底层** - 基于 CloudWeGo Hertz 引擎
- **统一上下文类型** - 解决类型冲突问题
- **中间件支持** - 内置日志、恢复等常用中间件
- **路由组支持** - 灵活的路由组织方式
- **多种数据绑定** - JSON、Query、URI 等数据绑定
- **多格式渲染** - JSON、HTML、String、Data 等渲染方式
- **静态文件服务** - 内置静态文件和目录服务

## 📦 安装

```bash
go get github.com/zsy619/yyhertz/framework/gin
```

## 🏃 快速开始

### 基本用法

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/gin"
)

func main() {
    // 创建默认引擎（包含 Logger 和 Recovery 中间件）
    r := gin.Default()

    // 定义路由
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello, YYHertz Gin!",
        })
    })

    // 启动服务器
    r.Run(":8080")
}
```

### 创建自定义引擎

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/gin"
)

func main() {
    // 创建空引擎
    r := gin.New()
    
    // 手动添加中间件
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    
    r.GET("/ping", func(c *gin.Context) {
        c.String(200, "pong")
    })
    
    r.Run()
}
```

## 🛣️ 路由

### HTTP 方法

支持所有标准 HTTP 方法：

```go
r.GET("/get", getHandler)
r.POST("/post", postHandler)
r.PUT("/put", putHandler)
r.DELETE("/delete", deleteHandler)
r.PATCH("/patch", patchHandler)
r.HEAD("/head", headHandler)
r.OPTIONS("/options", optionsHandler)

// 支持所有方法
r.Any("/any", anyHandler)

// 自定义方法
r.Handle("TRACE", "/trace", traceHandler)
```

### 路径参数

```go
// 单个参数
r.GET("/user/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"user_id": id})
})

// 多个参数
r.GET("/user/:id/book/:title", func(c *gin.Context) {
    userID := c.Param("id")
    title := c.Param("title")
    c.JSON(200, gin.H{
        "user_id": userID,
        "title": title,
    })
})

// 通配符参数
r.GET("/files/*filepath", func(c *gin.Context) {
    filepath := c.Param("filepath")
    c.String(200, "File path: %s", filepath)
})
```

### 查询参数

```go
r.GET("/search", func(c *gin.Context) {
    // 获取查询参数
    q := c.Query("q")                    // 获取 q 参数
    page := c.DefaultQuery("page", "1")  // 获取 page 参数，默认值为 "1"
    
    // 检查参数是否存在
    if value, exists := c.GetQuery("sort"); exists {
        // 处理 sort 参数
    }
    
    c.JSON(200, gin.H{
        "query": q,
        "page": page,
    })
})
```

## 📁 路由组

```go
// 创建路由组
v1 := r.Group("/api/v1")
{
    v1.GET("/users", getUsersHandler)
    v1.POST("/users", createUserHandler)
    v1.GET("/users/:id", getUserHandler)
}

// 带中间件的路由组
authorized := r.Group("/admin")
authorized.Use(AuthMiddleware())
{
    authorized.GET("/dashboard", dashboardHandler)
    authorized.POST("/settings", settingsHandler)
}

// 嵌套路由组
api := r.Group("/api")
{
    v1 := api.Group("/v1")
    {
        users := v1.Group("/users")
        {
            users.GET("/", listUsers)
            users.POST("/", createUser)
            users.GET("/:id", getUser)
            users.PUT("/:id", updateUser)
            users.DELETE("/:id", deleteUser)
        }
    }
}
```

## 🔧 中间件

### 使用内置中间件

```go
r := gin.New()

// 日志中间件
r.Use(gin.Logger())

// 恢复中间件
r.Use(gin.Recovery())

// 或使用默认中间件
r := gin.Default() // 已包含 Logger 和 Recovery
```

### 自定义中间件

```go
// 定义中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }
        
        // 验证 token...
        
        c.Next() // 继续执行下一个处理器
    }
}

// 使用中间件
r.Use(AuthMiddleware())

// 为特定路由使用中间件
r.GET("/protected", AuthMiddleware(), protectedHandler)
```

### 中间件控制流

```go
func MiddlewareExample() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        start := time.Now()
        
        // 设置变量
        c.Set("start_time", start)
        
        // 继续执行
        c.Next()
        
        // 后置处理
        latency := time.Since(start)
        status := c.Writer.Status()
        
        // 终止执行（不会执行后续中间件）
        // c.Abort()
        
        // 终止并返回错误
        // c.AbortWithStatus(500)
        // c.AbortWithStatusJSON(400, gin.H{"error": "Bad Request"})
    }
}
```

## 📄 数据绑定

### JSON 绑定

```go
type User struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"gte=0,lte=130"`
}

r.POST("/user", func(c *gin.Context) {
    var user User
    
    // 绑定 JSON（验证失败会返回 400 错误）
    if err := c.BindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 或使用 ShouldBindJSON（不会自动返回错误）
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"user": user})
})
```

### Query 参数绑定

```go
type SearchQuery struct {
    Query string `form:"q" binding:"required"`
    Page  int    `form:"page" binding:"min=1"`
    Size  int    `form:"size" binding:"min=1,max=100"`
}

r.GET("/search", func(c *gin.Context) {
    var query SearchQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"query": query})
})
```

### URI 参数绑定

```go
type UserID struct {
    ID int `uri:"id" binding:"required"`
}

r.GET("/user/:id", func(c *gin.Context) {
    var userID UserID
    if err := c.ShouldBindUri(&userID); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"user_id": userID.ID})
})
```

### 自动绑定

```go
r.POST("/user", func(c *gin.Context) {
    var user User
    
    // 根据 Content-Type 自动选择绑定方式
    if err := c.ShouldBind(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"user": user})
})
```

## 🎨 响应渲染

### JSON 响应

```go
r.GET("/json", func(c *gin.Context) {
    // 简单 JSON
    c.JSON(200, gin.H{
        "message": "Hello, World!",
        "status": "success",
    })
    
    // 结构体 JSON
    user := User{Name: "John", Email: "john@example.com"}
    c.JSON(200, user)
    
    // 数组 JSON
    users := []User{{Name: "John"}, {Name: "Jane"}}
    c.JSON(200, users)
})
```

### 字符串响应

```go
r.GET("/string", func(c *gin.Context) {
    c.String(200, "Hello, %s!", "World")
})
```

### HTML 响应

```go
r.GET("/html", func(c *gin.Context) {
    c.HTML(200, "index.html", gin.H{
        "title": "Home Page",
        "message": "Welcome!",
    })
})
```

### 原始数据响应

```go
r.GET("/data", func(c *gin.Context) {
    data := []byte("Raw binary data")
    c.Data(200, "application/octet-stream", data)
})
```

### 文件响应

```go
r.GET("/file", func(c *gin.Context) {
    c.File("./uploads/document.pdf")
})
```

### 设置响应头

```go
r.GET("/headers", func(c *gin.Context) {
    c.Header("X-Custom-Header", "Custom Value")
    c.Header("Cache-Control", "no-cache")
    c.JSON(200, gin.H{"message": "Success"})
})
```

## 📂 静态文件服务

### 单个静态文件

```go
// 访问 /favicon.ico 返回 ./static/favicon.ico
r.StaticFile("/favicon.ico", "./static/favicon.ico")
```

### 静态文件目录

```go
// 访问 /static/* 返回 ./assets/* 目录下的文件
r.Static("/static", "./assets")

// 访问 /assets/* 返回 ./public/* 目录下的文件
r.Static("/assets", "./public")
```

### 自定义文件系统

```go
// 使用自定义文件系统
r.StaticFS("/files", http.Dir("./uploads"))
```

## 🔥 上下文操作

### 存储和获取值

```go
r.GET("/context", func(c *gin.Context) {
    // 设置值
    c.Set("user_id", 123)
    c.Set("username", "john")
    
    // 获取值
    if userID, exists := c.Get("user_id"); exists {
        // 处理 user_id
    }
    
    // 必须获取值（不存在会 panic）
    username := c.MustGet("username").(string)
    
    // 获取特定类型的值
    userIDStr := c.GetString("user_id") // 返回空字符串如果不存在或类型不匹配
    
    c.JSON(200, gin.H{
        "user_id": userID,
        "username": username,
    })
})
```

### 获取请求信息

```go
r.POST("/info", func(c *gin.Context) {
    // 获取请求头
    userAgent := c.GetHeader("User-Agent")
    authorization := c.GetHeader("Authorization")
    
    // 获取客户端 IP
    clientIP := c.ClientIP()
    
    // 获取请求方法和路径
    method := c.Request.Method()
    path := c.Request.URI().Path()
    
    c.JSON(200, gin.H{
        "user_agent": userAgent,
        "client_ip": clientIP,
        "method": string(method),
        "path": string(path),
    })
})
```

## 🚨 错误处理

### 404 和 405 处理

```go
// 404 处理
r.NoRoute(func(c *gin.Context) {
    c.JSON(404, gin.H{"error": "Page not found"})
})

// 405 处理（方法不允许）
r.NoMethod(func(c *gin.Context) {
    c.JSON(405, gin.H{"error": "Method not allowed"})
})
```

### 错误收集

```go
r.GET("/error", func(c *gin.Context) {
    // 添加错误
    c.AbortWithError(400, errors.New("Something went wrong"))
    
    // 检查错误
    if len(c.Errors) > 0 {
        for _, err := range c.Errors {
            log.Println("Error:", err.Error())
        }
    }
})
```

## 🔧 高级用法

### 多端口启动

```go
// 默认端口 8080
r.Run()

// 指定端口
r.Run(":3000")

// 指定地址和端口
r.Run("0.0.0.0:8080")
```

### 优雅关闭

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"
    
    "github.com/zsy619/yyhertz/framework/gin"
)

func main() {
    r := gin.Default()
    
    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello"})
    })
    
    // 启动服务器
    go func() {
        if err := r.Run(":8080"); err != nil {
            log.Fatal("Server failed to start:", err)
        }
    }()
    
    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    
    log.Println("Shutting down server...")
    
    // 优雅关闭逻辑
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    log.Println("Server stopped")
}
```

## 🔍 性能监控

### 自定义日志格式

```go
func CustomLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URI().Path()
        raw := c.Request.URI().RawQuery()
        
        c.Next()
        
        latency := time.Since(start)
        clientIP := c.ClientIP()
        method := c.Request.Method()
        statusCode := c.Writer.Status()
        
        if raw != "" {
            path = path + "?" + raw
        }
        
        log.Printf("[%s] %s %s %d %v",
            clientIP,
            string(method),
            path,
            statusCode,
            latency,
        )
    }
}

r.Use(CustomLogger())
```

## 📚 API 参考

### Engine 方法

| 方法 | 描述 |
|------|------|
| `New()` | 创建新引擎 |
| `Default()` | 创建带默认中间件的引擎 |
| `Use(middleware...)` | 添加全局中间件 |
| `Group(path)` | 创建路由组 |
| `GET/POST/PUT/DELETE/...` | 注册路由 |
| `Handle(method, path, handlers...)` | 注册自定义方法路由 |
| `Any(path, handlers...)` | 注册所有方法路由 |
| `Static/StaticFile/StaticFS` | 静态文件服务 |
| `NoRoute/NoMethod` | 错误处理 |
| `Run(addr...)` | 启动服务器 |

### Context 方法

| 方法 | 描述 |
|------|------|
| `Param(key)` | 获取路径参数 |
| `Query(key)` | 获取查询参数 |
| `DefaultQuery(key, default)` | 获取查询参数（带默认值） |
| `GetQuery(key)` | 获取查询参数（返回是否存在） |
| `GetHeader(key)` | 获取请求头 |
| `Bind/BindJSON/ShouldBind...` | 数据绑定 |
| `JSON/String/HTML/Data` | 响应渲染 |
| `Set/Get/MustGet` | 上下文值操作 |
| `Next/Abort` | 中间件控制 |
| `Header(key, value)` | 设置响应头 |
| `Status(code)` | 设置状态码 |

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目基于 MIT 许可证开源。

## 🔗 相关链接

- [CloudWeGo Hertz](https://github.com/cloudwego/hertz)
- [Gin Framework](https://github.com/gin-gonic/gin)
- [Go 官方文档](https://golang.org/doc/)

---

**注意**: 本框架旨在提供 Gin 兼容的 API，同时利用 Hertz 的高性能特性。如果您正在从 Gin 迁移，大部分代码应该可以直接使用，只需要更改导入路径即可。