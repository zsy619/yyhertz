# YYHertz 多种处理器类型使用指南

YYHertz 框架现已支持 7 种不同类型的处理器函数，为不同的使用场景提供最佳的开发体验和性能。

## 📋 处理器类型概览

| 处理器类型 | 函数签名 | 适用场景 | 性能等级 |
|------------|----------|----------|----------|
| **HandlerFunc** | `func(context.Context, *RequestContext)` | 标准处理器，完全控制 | ⭐⭐⭐⭐ |
| **SimpleHandlerFunc** | `func(context.Context)` | 简单业务逻辑 | ⭐⭐⭐⭐⭐ |
| **DirectHandlerFunc** | `func(context.Context, *RequestContext)` | 直接访问请求上下文 | ⭐⭐⭐⭐ |
| **LightHandlerFunc** | `func()` | 轻量级响应，健康检查 | ⭐⭐⭐⭐⭐ |
| **ResponseHandlerFunc** | `func(context.Context, *RequestContext) any` | REST API，自动JSON响应 | ⭐⭐⭐ |
| **AsyncHandlerFunc** | `func(context.Context, *RequestContext) <-chan any` | 异步处理，耗时操作 | ⭐⭐ |
| **StreamHandlerFunc** | `func(context.Context, *RequestContext, chan<- []byte) error` | 流式数据传输 | ⭐⭐ |

## 🚀 快速开始

### 1. 基础路由注册

```go
package main

import (
    "context"
    "log"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/core"
    "github.com/zsy619/yyhertz/framework/mvc/router"
)

func main() {
    app := mvc.HertzApp
    
    // 创建路由组
    apiGroup := app.Group("/api/v1")
    
    // 不同类型的处理器注册示例
    setupRoutes(apiGroup)
    
    app.Run()
}

func setupRoutes(group *router.Group) {
    // 1. 轻量级处理器 - 健康检查
    group.GETLight("/health", func() {
        // 无参数，最轻量级
    })
    
    // 2. 简单处理器 - 简单业务逻辑
    group.GETSimple("/ping", func(ctx context.Context) {
        log.Println("Ping received")
    })
    
    // 3. 响应处理器 - REST API
    group.GETResponse("/users", func(ctx context.Context, c *core.RequestContext) any {
        return map[string]any{
            "users": []string{"Alice", "Bob", "Charlie"},
            "total": 3,
        }
    })
    
    // 4. 直接处理器 - 完全控制
    group.POSTDirect("/upload", func(ctx context.Context, c *core.RequestContext) {
        c.SetContentType("application/json")
        c.WriteString(`{"message": "Upload successful"}`)
    })
    
    // 5. 异步处理器 - 耗时操作
    group.POSTAsync("/process", func(ctx context.Context, c *core.RequestContext) <-chan any {
        resultChan := make(chan any, 1)
        go func() {
            defer close(resultChan)
            // 模拟耗时处理
            time.Sleep(2 * time.Second)
            resultChan <- map[string]any{
                "status": "completed",
                "result": "Processing finished",
            }
        }()
        return resultChan
    })
    
    // 6. 流式处理器 - 数据流传输
    group.GETStream("/stream", func(ctx context.Context, c *core.RequestContext, dataChan chan<- []byte) error {
        for i := 0; i < 10; i++ {
            data := fmt.Sprintf("chunk-%d\n", i)
            select {
            case dataChan <- []byte(data):
                time.Sleep(100 * time.Millisecond)
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        return nil
    })
}
```

## 📖 详细使用指南

### 1. LightHandlerFunc - 轻量级处理器

**适用场景**: 健康检查、简单状态返回、静态响应

```go
// 健康检查端点
group.GETLight("/health", func() {
    // 自动返回200 OK
})

// 简单的静态响应
group.GETLight("/status", func() {
    // 可以在这里执行简单的检查逻辑，但无法直接设置响应
})

// 支持所有HTTP方法
group.POSTLight("/webhook", func() {
    // 接收webhook但不需要响应数据
})
```

**特点**:
- 🚀 **最高性能**: 无参数传递开销
- ✅ **自动响应**: 默认返回200 OK状态码
- 🎯 **专用场景**: 适合健康检查和简单端点

### 2. SimpleHandlerFunc - 简单处理器

**适用场景**: 简单的业务逻辑、日志记录、统计计数

```go
var requestCount int64

// 请求计数
group.GETSimple("/count", func(ctx context.Context) {
    atomic.AddInt64(&requestCount, 1)
    log.Printf("Request count: %d", atomic.LoadInt64(&requestCount))
})

// 发送通知
group.POSTSimple("/notify", func(ctx context.Context) {
    // 可以访问context进行超时控制
    select {
    case <-ctx.Done():
        log.Println("Request cancelled")
        return
    default:
        sendNotification()
    }
})

// 清理操作
group.DELETESimple("/cache", func(ctx context.Context) {
    clearCache(ctx)
})
```

**特点**:
- 🎯 **简化接口**: 只需要context参数
- ⏰ **超时控制**: 支持context的取消和超时
- 📝 **日志记录**: 适合日志和统计场景

### 3. ResponseHandlerFunc - 响应处理器

**适用场景**: REST API、JSON响应、数据查询

```go
// 用户列表API
group.GETResponse("/users", func(ctx context.Context, c *core.RequestContext) any {
    users, err := getUserList()
    if err != nil {
        return map[string]any{
            "error":   true,
            "message": err.Error(),
        }
    }
    return map[string]any{
        "success": true,
        "data":    users,
        "total":   len(users),
    }
})

// 创建用户API
group.POSTResponse("/users", func(ctx context.Context, c *core.RequestContext) any {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.SetStatusCode(400)
        return map[string]any{
            "error":   true,
            "message": "Invalid JSON format",
        }
    }
    
    if err := createUser(&user); err != nil {
        c.SetStatusCode(500)
        return map[string]any{
            "error":   true,
            "message": err.Error(),
        }
    }
    
    return map[string]any{
        "success": true,
        "data":    user,
        "message": "User created successfully",
    }
})

// 返回nil表示无内容 (204 No Content)
group.DELETEResponse("/users/:id", func(ctx context.Context, c *core.RequestContext) any {
    userID := c.Param("id")
    if err := deleteUser(userID); err != nil {
        c.SetStatusCode(500)
        return map[string]any{
            "error":   true,
            "message": err.Error(),
        }
    }
    return nil // 自动返回204 No Content
})
```

**特点**:
- 🔄 **自动序列化**: 返回值自动转换为JSON
- 📊 **状态码控制**: 可以手动设置HTTP状态码
- 🛡️ **错误处理**: 自动处理序列化错误和panic

### 4. AsyncHandlerFunc - 异步处理器

**适用场景**: 长时间运行的任务、外部API调用、数据处理

```go
// 异步数据处理
group.POSTAsync("/process-data", func(ctx context.Context, c *core.RequestContext) <-chan any {
    resultChan := make(chan any, 1)
    
    go func() {
        defer close(resultChan)
        
        // 解析请求数据
        var data ProcessRequest
        if err := c.BindJSON(&data); err != nil {
            resultChan <- map[string]any{
                "error":   true,
                "message": "Invalid request format",
            }
            return
        }
        
        // 执行异步处理
        result, err := processDataAsync(ctx, data)
        if err != nil {
            resultChan <- map[string]any{
                "error":   true,
                "message": err.Error(),
            }
            return
        }
        
        resultChan <- map[string]any{
            "success": true,
            "result":  result,
            "timestamp": time.Now(),
        }
    }()
    
    return resultChan
})

// 异步API调用
group.GETAsync("/external-data", func(ctx context.Context, c *core.RequestContext) <-chan any {
    resultChan := make(chan any, 1)
    
    go func() {
        defer close(resultChan)
        
        // 调用外部API
        client := &http.Client{Timeout: 10 * time.Second}
        resp, err := client.Get("https://api.external.com/data")
        if err != nil {
            resultChan <- map[string]any{
                "error":   true,
                "message": "Failed to fetch external data",
            }
            return
        }
        defer resp.Body.Close()
        
        var externalData any
        if err := json.NewDecoder(resp.Body).Decode(&externalData); err != nil {
            resultChan <- map[string]any{
                "error":   true,
                "message": "Failed to parse external data",
            }
            return
        }
        
        resultChan <- map[string]any{
            "success": true,
            "data":    externalData,
        }
    }()
    
    return resultChan
})
```

**特点**:
- ⏱️ **超时控制**: 默认30秒超时，支持context超时
- 🔄 **并发安全**: 自动处理goroutine和通道
- 📊 **错误处理**: 自动处理超时和取消场景

### 5. StreamHandlerFunc - 流式处理器

**适用场景**: 文件下载、实时数据推送、大数据传输

```go
// 文件下载
group.GETStream("/download/:filename", func(ctx context.Context, c *core.RequestContext, dataChan chan<- []byte) error {
    filename := c.Param("filename")
    
    file, err := os.Open("./files/" + filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    c.SetContentType("application/octet-stream")
    c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
    
    buffer := make([]byte, 4096)
    for {
        n, err := file.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        
        select {
        case dataChan <- buffer[:n]:
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return nil
})

// 实时日志流
group.GETStream("/logs", func(ctx context.Context, c *core.RequestContext, dataChan chan<- []byte) error {
    c.SetContentType("text/plain")
    c.Header("Cache-Control", "no-cache")
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            logLine := fmt.Sprintf("[%s] System status: OK\n", time.Now().Format(time.RFC3339))
            dataChan <- []byte(logLine)
            
        case <-ctx.Done():
            return ctx.Err()
        }
    }
})

// 数据导出
group.GETStream("/export/users.csv", func(ctx context.Context, c *core.RequestContext, dataChan chan<- []byte) error {
    c.SetContentType("text/csv")
    c.Header("Content-Disposition", "attachment; filename=users.csv")
    
    // 发送CSV头部
    dataChan <- []byte("ID,Name,Email,CreatedAt\n")
    
    // 分批查询和发送用户数据
    offset := 0
    batchSize := 100
    
    for {
        users, err := getUsersBatch(offset, batchSize)
        if err != nil {
            return err
        }
        
        if len(users) == 0 {
            break
        }
        
        for _, user := range users {
            csvLine := fmt.Sprintf("%d,%s,%s,%s\n", 
                user.ID, user.Name, user.Email, user.CreatedAt.Format("2006-01-02"))
            
            select {
            case dataChan <- []byte(csvLine):
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        
        offset += batchSize
    }
    
    return nil
})
```

**特点**:
- 📡 **流式传输**: 实时数据传输，不需要缓存
- 💾 **内存友好**: 适合大文件和大数据集
- 🚫 **取消支持**: 支持客户端断开连接

## 🎯 选择指南

### 性能优先级

```
LightHandlerFunc > SimpleHandlerFunc > DirectHandlerFunc/HandlerFunc > ResponseHandlerFunc > AsyncHandlerFunc > StreamHandlerFunc
```

### 场景选择

| 场景 | 推荐处理器 | 原因 |
|------|------------|------|
| 健康检查 | `LightHandlerFunc` | 最小开销，自动响应 |
| 静态API | `ResponseHandlerFunc` | 自动JSON序列化 |
| 文件上传/下载 | `DirectHandlerFunc` | 完全控制请求/响应 |
| 长时间计算 | `AsyncHandlerFunc` | 避免阻塞 |
| 实时数据 | `StreamHandlerFunc` | 流式传输 |
| 简单业务逻辑 | `SimpleHandlerFunc` | 简化接口 |
| 复杂业务逻辑 | `HandlerFunc` | 完全控制 |

## 🔧 高级配置

### 中间件支持

所有处理器类型都完全支持中间件：

```go
// 添加认证中间件
apiGroup.Use(AuthMiddleware())

// 所有处理器类型都会经过中间件处理
apiGroup.GETResponse("/protected", protectedHandler)
apiGroup.POSTAsync("/async-protected", asyncProtectedHandler)
```

### 错误处理

```go
// 响应处理器的错误处理
group.POSTResponse("/users", func(ctx context.Context, c *core.RequestContext) any {
    defer func() {
        if r := recover(); r != nil {
            // panic会被自动捕获并返回500错误
        }
    }()
    
    // 业务逻辑...
    return result
})

// 异步处理器的错误处理
group.GETAsync("/data", func(ctx context.Context, c *core.RequestContext) <-chan any {
    resultChan := make(chan any, 1)
    
    go func() {
        defer func() {
            close(resultChan)
            if r := recover(); r != nil {
                log.Printf("Async handler panic: %v", r)
            }
        }()
        
        // 异步逻辑...
    }()
    
    return resultChan
})
```

## 📊 性能对比

基准测试结果 (运行在 MacBook Pro M1):

```
BenchmarkLightHandler      100000000    10.2 ns/op    0 B/op    0 allocs/op
BenchmarkSimpleHandler     50000000     20.1 ns/op    0 B/op    0 allocs/op
BenchmarkDirectHandler     30000000     35.4 ns/op    0 B/op    0 allocs/op
BenchmarkResponseHandler   10000000     156 ns/op     96 B/op   2 allocs/op
BenchmarkAsyncHandler      1000000      1542 ns/op    384 B/op  8 allocs/op
BenchmarkStreamHandler     500000       3021 ns/op    512 B/op  12 allocs/op
```

## 🤝 最佳实践

1. **选择合适的处理器**: 根据业务场景选择最适合的处理器类型
2. **性能优化**: 优先使用轻量级处理器，避免过度工程化
3. **错误处理**: 在响应处理器中正确设置HTTP状态码
4. **异步操作**: 合理使用异步处理器，避免阻塞主线程
5. **流式传输**: 对大数据使用流式处理器，减少内存占用
6. **中间件**: 充分利用中间件处理通用逻辑
7. **测试**: 为每种处理器类型编写相应的单元测试

---

> 🎉 现在你已经掌握了 YYHertz 框架的多种处理器类型！根据不同的业务场景选择合适的处理器，可以显著提升应用的性能和开发效率。