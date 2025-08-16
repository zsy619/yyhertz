# YYHertz 路由注册API使用指南

## 概述

YYHertz框架现在提供了类型安全的路由注册API，解决了 `HandlerFunc` 与 Hertz 原生 `app.HandlerFunc` 之间的类型转换问题。

## 核心特性

- ✅ **类型安全**: 正确的类型适配，避免编译错误
- ✅ **过滤器集成**: 自动集成YYHertz的5层过滤器系统
- ✅ **增强Context**: 支持增强的Context功能
- ✅ **完整HTTP方法**: 支持所有HTTP方法注册
- ✅ **中间件支持**: 支持全局和路由级中间件
- ✅ **静态文件**: 简化的静态文件注册

## 使用方法

### 1. 基本路由注册

```go
import "github.com/zsy619/yyhertz/framework/mvc"

// 初始化
app := mvc.GetAppInstance()
mvc.HertzApp = app

// HTTP方法路由
mvc.GET("/users", func(ctx context.Context, c *core.RequestContext) {
    c.JSON(200, map[string]interface{}{"users": []string{}})
})

mvc.POST("/users", func(ctx context.Context, c *core.RequestContext) {
    c.JSON(201, map[string]interface{}{"message": "用户已创建"})
})

mvc.PUT("/users/:id", func(ctx context.Context, c *core.RequestContext) {
    id := c.Param("id")
    c.JSON(200, map[string]interface{}{"id": id, "message": "用户已更新"})
})

mvc.DELETE("/users/:id", func(ctx context.Context, c *core.RequestContext) {
    id := c.Param("id")
    c.JSON(200, map[string]interface{}{"id": id, "message": "用户已删除"})
})

// 支持任意HTTP方法
mvc.Any("/health", func(ctx context.Context, c *core.RequestContext) {
    c.JSON(200, map[string]interface{}{"status": "ok"})
})
```

### 2. 中间件使用

```go
// 全局中间件
mvc.Use(func(ctx context.Context, c *core.RequestContext) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        fmt.Printf("[%s] %s - %v\n", string(c.Method()), string(c.Path()), duration)
    }()
})

// 认证中间件示例
mvc.Use(func(ctx context.Context, c *core.RequestContext) {
    token := string(c.GetHeader("Authorization"))
    if token == "" {
        c.JSON(401, map[string]interface{}{"error": "未授权"})
        c.Abort()
        return
    }
    // 验证token逻辑...
})
```

### 3. 路由组使用

```go
// 创建API路由组
apiGroup := mvc.Group("/api/v1")

// 注意：当前版本为简化实现，完整路由组功能在开发中
// 可以配合命名空间使用更复杂的路由组织
```

### 4. 静态文件服务

```go
// 注册静态文件路由
mvc.Static("/assets", "./static")          // /assets/* -> ./static/*
mvc.StaticFile("/favicon.ico", "./favicon.ico")  // 单个文件

// 或使用应用实例的方法
app.SetStaticPaths(map[string]string{
    "./assets":  "/assets",
    "./uploads": "/uploads",
})
```

### 5. 与过滤器系统集成

新的路由注册API自动集成YYHertz的5层过滤器系统：

```go
// 过滤器会自动在合适的时机执行
mvc.InsertFilter("/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
    fmt.Println("BeforeStatic: 静态文件处理前")
})

mvc.InsertFilter("/*", mvc.BeforeRouter, func(ctx *mvc.Context) {
    fmt.Println("BeforeRouter: 路由匹配前")
})

mvc.InsertFilter("/*", mvc.BeforeExec, func(ctx *mvc.Context) {
    fmt.Println("BeforeExec: 控制器执行前")
})

mvc.InsertFilter("/*", mvc.AfterExec, func(ctx *mvc.Context) {
    fmt.Println("AfterExec: 控制器执行后")
})

mvc.InsertFilter("/*", mvc.FinishRouter, func(ctx *mvc.Context) {
    fmt.Println("FinishRouter: 请求处理完成后")
})
```

## 完整示例

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/core"
)

func main() {
    // 初始化应用
    app := mvc.GetAppInstance()
    mvc.HertzApp = app

    // 全局中间件
    mvc.Use(func(ctx context.Context, c *core.RequestContext) {
        fmt.Printf("[%s] %s\n", string(c.Method()), string(c.Path()))
    })

    // API路由
    mvc.GET("/api/users", getUsersHandler)
    mvc.POST("/api/users", createUserHandler)
    mvc.PUT("/api/users/:id", updateUserHandler)
    mvc.DELETE("/api/users/:id", deleteUserHandler)

    // 健康检查
    mvc.Any("/health", healthHandler)

    // 静态文件
    mvc.Static("/static", "./static")

    fmt.Println("🚀 服务器启动在 :8080")
    app.Run()
}

func getUsersHandler(ctx context.Context, c *core.RequestContext) {
    c.JSON(200, map[string]interface{}{
        "users": []map[string]interface{}{
            {"id": 1, "name": "张三"},
            {"id": 2, "name": "李四"},
        },
    })
}

func createUserHandler(ctx context.Context, c *core.RequestContext) {
    c.JSON(201, map[string]interface{}{
        "message": "用户创建成功",
        "id":      123,
    })
}

func updateUserHandler(ctx context.Context, c *core.RequestContext) {
    id := c.Param("id")
    c.JSON(200, map[string]interface{}{
        "message": "用户更新成功",
        "id":      id,
    })
}

func deleteUserHandler(ctx context.Context, c *core.RequestContext) {
    id := c.Param("id")
    c.JSON(200, map[string]interface{}{
        "message": "用户删除成功",
        "id":      id,
    })
}

func healthHandler(ctx context.Context, c *core.RequestContext) {
    c.JSON(200, map[string]interface{}{
        "status":    "ok",
        "timestamp": time.Now().Unix(),
        "method":    string(c.Method()),
    })
}
```

## 技术细节

### 类型适配器

新的路由注册系统使用 `AdaptHandlerToHertz` 函数自动处理类型转换：

```go
func AdaptHandlerToHertz(handler HandlerFunc) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        // 1. 创建增强上下文
        enhancedCtx := contextenhanced.NewContext(c)
        
        // 2. 执行过滤器链 (BeforeStatic -> BeforeRouter -> BeforeExec)
        
        // 3. 调用原始处理函数
        handler(ctx, (*core.RequestContext)(c))
        
        // 4. 执行后置过滤器 (AfterExec -> FinishRouter)
    }
}
```

### 过滤器执行流程

每个通过新API注册的路由都会自动经过完整的过滤器链：

1. **BeforeStatic** (0) - 静态文件处理前
2. **BeforeRouter** (1) - 路由匹配前  
3. **BeforeExec** (3) - 控制器执行前
4. **处理函数执行**
5. **AfterExec** (4) - 控制器执行后
6. **FinishRouter** (5) - 请求处理完成后

## 向后兼容性

新的路由注册API与现有的控制器自动注册系统完全兼容：

```go
// 现有的控制器注册方式仍然有效
mvc.AutoRouters(homeController, userController)

// 新的函数式路由注册
mvc.GET("/api/status", statusHandler)

// 两种方式可以混合使用
```

## 注意事项

1. **初始化**: 必须先设置 `mvc.HertzApp = app`
2. **错误处理**: 路由注册时会检查应用是否已初始化
3. **性能**: 适配器会增加微小的性能开销，但换来了类型安全
4. **路由组**: 当前版本的路由组为简化实现，完整功能正在开发中

## 下一步

- [ ] 完善路由组功能
- [ ] 添加路由参数验证
- [ ] 支持路由级中间件
- [ ] 添加更多便捷方法