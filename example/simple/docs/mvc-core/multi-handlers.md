# 🎯 多Handler类型系统详解

YYHertz v2.0引入了革命性的多Handler类型系统，提供7种不同的处理器类型，每种都针对特定场景进行了优化，同时使用增强的`mvcContext.Context`实现高性能并发处理。

## 📋 Handler类型概览

| Handler类型 | 函数签名 | 适用场景 | 性能特点 | 示例用途 |
|------------|----------|----------|----------|----------|
| **LightHandler** | `func()` | 健康检查、静态响应 | 0 allocs, 3.4ns/op | 监控端点、状态检查 |
| **SimpleHandler** | `func(context.Context)` | 简单业务逻辑 | 0 allocs, 3.0ns/op | 简单API、工具方法 |
| **DirectHandler** | `func(*mvcContext.Context)` | 直接控制 | 3 allocs, 1037ns/op | 自定义响应、高性能API |
| **ResponseHandler** | `func(*mvcContext.Context) any` | REST API | 4 allocs, 2686ns/op | JSON API、数据查询 |
| **AsyncHandler** | `func(*mvcContext.Context) <-chan any` | 异步处理 | 可配置超时 | 耗时操作、批处理 |
| **StreamHandler** | `func(*mvcContext.Context, chan<- []byte) error` | 流式传输 | 低延迟传输 | 大文件、实时数据 |
| **HandlerFunc** | `func(context.Context, *RequestContext)` | 传统方式 | 向后兼容 | 迁移项目 |

## 🚀 增强Context系统 (mvcContext.Context)

### 核心特性

YYHertz使用增强的`mvcContext.Context`替代传统的`*RequestContext`，提供：

#### 🏎️ 高性能优化
- **对象池化**: 减少GC压力，提升并发性能
- **原子操作**: 使用sync.Map和原子操作优化并发安全
- **零拷贝**: 高效的数据传递机制

#### 🔧 功能增强
- **类型安全**: 提供强类型的键值存储方法
- **批量操作**: 支持键值对的批量设置和获取
- **并发友好**: 高效的并发读写支持

```go
// 增强Context的典型用法
func handler(c *mvcContext.Context) {
    // 类型安全的数据存储
    c.SetTypedString("user_id", "12345")
    c.SetTypedInt("request_count", 1)
    c.SetTypedBool("authenticated", true)
    
    // 批量操作
    c.SetMultiple(map[string]any{
        "service": "user_api",
        "version": "2.0",
        "timestamp": time.Now(),
    })
    
    // 高效检索
    if userID, ok := c.GetTypedString("user_id"); ok {
        log.Printf("Processing request for user: %s", userID)
    }
    
    // 条件操作
    c.SetIfNotExists("session_id", generateSessionID())
    
    // 访问原始RequestContext（向后兼容）
    reqCtx := c.RequestContext()
    reqCtx.JSON(200, map[string]any{
        "message": "Enhanced context in action",
        "keys_count": c.KeysCount(),
    })
}
```

## 📝 详细Handler类型说明

### 1. LightHandler - 轻量级处理器

**最适用于**: 健康检查、监控端点、静态响应

```go
// 函数签名
type LightHandlerFunc = func()

// 使用示例
group.GETLight("/health", func() {
    // 无参数，最小开销
    // 自动返回200 OK状态
    log.Println("Health check performed")
})

group.GETLight("/ready", func() {
    // 适用于Kubernetes就绪探针
})
```

**性能特点**:
- **3.388 ns/op** - 极低延迟
- **0 allocs/op** - 零内存分配
- 适合高频调用的监控端点

### 2. SimpleHandler - 简单处理器

**最适用于**: 基础业务逻辑、需要上下文控制的场景

```go
// 函数签名
type SimpleHandlerFunc = func(context.Context)

// 使用示例
group.GETSimple("/ping", func(ctx context.Context) {
    // 支持上下文取消
    select {
    case <-ctx.Done():
        log.Println("Request cancelled")
        return
    default:
        log.Printf("Ping received at %s", time.Now().Format("15:04:05"))
    }
})

group.POSTSimple("/notify", func(ctx context.Context) {
    // 简单的通知逻辑
    timeout := time.After(5 * time.Second)
    select {
    case <-timeout:
        log.Println("Notification sent")
    case <-ctx.Done():
        log.Println("Notification cancelled")
    }
})
```

### 3. DirectHandler - 直接处理器

**最适用于**: 需要完全控制响应的场景、高性能API

```go
// 函数签名  
type DirectHandlerFunc = func(*mvcContext.Context)

// 使用示例
group.GETDirect("/metrics", func(c *mvcContext.Context) {
    // 使用增强Context功能
    c.Set("handler_type", "metrics")
    c.SetTypedString("service", "api_server")
    
    // 获取原始RequestContext进行响应控制
    reqCtx := c.RequestContext()
    reqCtx.SetContentType("text/plain")
    reqCtx.SetStatusCode(200)
    
    // 自定义响应格式
    metrics := fmt.Sprintf(`# HELP requests_total Total requests
# TYPE requests_total counter
requests_total{service="api_server",handler="metrics"} %d
context_keys_count %d
`, getRequestCount(), c.KeysCount())
    
    reqCtx.WriteString(metrics)
    log.Printf("Metrics exported with %d context keys", c.KeysCount())
})

group.POSTDirect("/upload", func(c *mvcContext.Context) {
    // 文件上传处理
    c.Set("upload_start", time.Now())
    
    reqCtx := c.RequestContext()
    fileHeader, err := reqCtx.FormFile("file")
    if err != nil {
        reqCtx.SetStatusCode(400)
        reqCtx.WriteString(`{"error":"No file uploaded"}`)
        return
    }
    
    // 保存文件逻辑...
    c.SetTypedString("filename", fileHeader.Filename)
    c.SetTypedInt("filesize", int(fileHeader.Size))
    
    reqCtx.SetStatusCode(201)
    reqCtx.WriteString(`{"success":true,"message":"File uploaded"}`)
})
```

### 4. ResponseHandler - 响应处理器

**最适用于**: REST API、JSON响应、标准CRUD操作

```go
// 函数签名
type ResponseHandlerFunc = func(*mvcContext.Context) any

// 使用示例
group.GETResponse("/users", func(c *mvcContext.Context) any {
    // 增强Context操作
    c.Set("operation", "list_users")
    c.SetTypedString("endpoint", "/users")
    
    // 模拟数据库查询
    users := getUsersFromDB()
    
    // 自动JSON序列化
    return map[string]any{
        "success": true,
        "data": users,
        "total": len(users),
        "timestamp": time.Now(),
        "context_keys": c.KeysCount(),
    }
})

group.POSTResponse("/users", func(c *mvcContext.Context) any {
    c.Set("operation", "create_user")
    
    // 获取请求数据
    reqCtx := c.RequestContext()
    var userData map[string]any
    if err := reqCtx.BindJSON(&userData); err != nil {
        c.RequestContext().SetStatusCode(400)
        return map[string]any{
            "success": false,
            "error": "Invalid JSON data",
        }
    }
    
    // 业务逻辑处理
    user := createUser(userData)
    c.SetTypedString("created_user_id", user.ID)
    
    return map[string]any{
        "success": true,
        "data": user,
        "message": "User created successfully",
    }
})
```

### 5. AsyncHandler - 异步处理器

**最适用于**: 耗时操作、批处理、后台任务

```go
// 函数签名
type AsyncHandlerFunc = func(*mvcContext.Context) <-chan any

// 使用示例
group.POSTAsync("/process-batch", func(c *mvcContext.Context) <-chan any {
    c.Set("operation", "batch_process")
    c.SetTypedString("batch_id", generateBatchID())
    
    resultChan := make(chan any, 1)
    
    go func() {
        defer close(resultChan)
        
        // 模拟批处理操作
        log.Printf("Starting batch processing...")
        c.SetTypedInt("processed_items", 0)
        
        for i := 1; i <= 100; i++ {
            time.Sleep(50 * time.Millisecond) // 模拟处理时间
            c.SetTypedInt("processed_items", i)
            
            if i%20 == 0 {
                log.Printf("Progress: %d/100 items processed", i)
            }
        }
        
        resultChan <- map[string]any{
            "success": true,
            "message": "Batch processing completed",
            "processed_items": c.GetTypedInt("processed_items"),
            "batch_id": c.GetTypedString("batch_id"),
            "duration": time.Since(time.Now()),
        }
    }()
    
    return resultChan
})

// 带超时控制的异步处理
group.POSTAsync("/analyze", func(c *mvcContext.Context) <-chan any {
    c.Set("operation", "data_analysis")
    resultChan := make(chan any, 1)
    
    go func() {
        defer close(resultChan)
        
        // 模拟数据分析（可能很耗时）
        analysisResult := performDataAnalysis()
        
        resultChan <- map[string]any{
            "success": true,
            "result": analysisResult,
            "analysis_type": "comprehensive",
        }
    }()
    
    return resultChan
})
```

### 6. StreamHandler - 流式处理器

**最适用于**: 大文件传输、实时数据流、日志流

```go
// 函数签名
type StreamHandlerFunc = func(*mvcContext.Context, chan<- []byte) error

// 使用示例
group.GETStream("/logs", func(c *mvcContext.Context, dataChan chan<- []byte) error {
    c.Set("operation", "stream_logs")
    c.SetTypedString("log_level", "info")
    
    // 设置流式响应头
    reqCtx := c.RequestContext()
    reqCtx.Header("Content-Type", "text/plain")
    reqCtx.Header("Cache-Control", "no-cache")
    
    log.Println("Starting log stream...")
    
    // 模拟实时日志流
    for i := 1; i <= 50; i++ {
        logEntry := fmt.Sprintf("[%s] INFO: Log entry %d\n", 
            time.Now().Format("15:04:05.000"), i)
        
        dataChan <- []byte(logEntry)
        c.SetTypedInt("streamed_lines", i)
        
        time.Sleep(200 * time.Millisecond)
    }
    
    log.Printf("Log stream completed. Streamed %d lines", 
        c.GetTypedInt("streamed_lines"))
    return nil
})

group.GETStream("/download", func(c *mvcContext.Context, dataChan chan<- []byte) error {
    c.Set("operation", "file_download")
    
    // 模拟大文件下载
    filename := c.RequestContext().Query("file")
    if filename == "" {
        return fmt.Errorf("filename parameter required")
    }
    
    c.SetTypedString("filename", filename)
    c.RequestContext().Header("Content-Disposition", 
        fmt.Sprintf("attachment; filename=%s", filename))
    
    // 分块传输大文件
    chunkSize := 64 * 1024 // 64KB chunks
    totalSize := 10 * 1024 * 1024 // 10MB file
    
    for sent := 0; sent < totalSize; sent += chunkSize {
        chunk := make([]byte, chunkSize)
        // 填充模拟数据
        for i := range chunk {
            chunk[i] = byte(i % 256)
        }
        
        dataChan <- chunk
        c.SetTypedInt("bytes_sent", sent+chunkSize)
        
        time.Sleep(10 * time.Millisecond) // 模拟网络延迟
    }
    
    return nil
})
```

## 🔧 路由注册方法

### HTTP方法支持

每种Handler类型都支持所有HTTP方法：

```go
group := mvc.CreateGroup("/api/v1")

// GET方法
group.GETLight("/health", lightHandler)
group.GETSimple("/ping", simpleHandler)  
group.GETDirect("/info", directHandler)
group.GETResponse("/users", responseHandler)
group.GETAsync("/report", asyncHandler)
group.GETStream("/logs", streamHandler)

// POST方法
group.POSTResponse("/users", createUserHandler)
group.POSTAsync("/process", processHandler)
group.POSTStream("/upload", uploadStreamHandler)

// PUT方法  
group.PUTResponse("/users/:id", updateUserHandler)

// DELETE方法
group.DELETEResponse("/users/:id", deleteUserHandler)

// PATCH方法
group.PATCHResponse("/users/:id", patchUserHandler)

// HEAD方法
group.HEADLight("/check", checkHandler)

// OPTIONS方法
group.OPTIONSLight("/cors", corsHandler)

// Any方法 - 支持所有HTTP方法
group.AnyResponse("/webhook", webhookHandler)
```

## ⚡ 性能对比

### 基准测试结果

```
BenchmarkAdapters/SimpleHandler-8     675M    2.977 ns/op    0 B/op    0 allocs/op
BenchmarkAdapters/LightHandler-8      1000M   3.388 ns/op    0 B/op    0 allocs/op  
BenchmarkAdapters/DirectHandler-8     1M      1037 ns/op     64 B/op   3 allocs/op
BenchmarkAdapters/ResponseHandler-8   1M      2686 ns/op     104 B/op  4 allocs/op
```

### 性能分析

1. **LightHandler** 和 **SimpleHandler** 零分配，适合高频调用
2. **DirectHandler** 使用对象池化，3次分配用于增强Context创建
3. **ResponseHandler** 包含JSON序列化开销，但仍保持高性能
4. **AsyncHandler** 和 **StreamHandler** 性能取决于具体业务逻辑

## 🛠️ 最佳实践

### 1. Handler选择原则

```go
// ✅ 正确选择
group.GETLight("/health", healthCheck)          // 监控端点用Light
group.GETResponse("/api/users", getUsers)       // REST API用Response  
group.POSTAsync("/api/process", processData)    // 耗时操作用Async
group.GETStream("/api/logs", streamLogs)        // 大数据传输用Stream

// ❌ 错误选择
group.GETAsync("/api/users", getUsers)          // 简单查询不需要异步
group.GETResponse("/health", healthCheck)       // 健康检查不需要JSON响应
```

### 2. 增强Context的有效使用

```go
func efficientHandler(c *mvcContext.Context) any {
    // ✅ 批量设置减少操作次数
    c.SetMultiple(map[string]any{
        "operation": "user_query",
        "start_time": time.Now(),
        "request_id": generateID(),
    })
    
    // ✅ 类型安全的操作
    c.SetTypedString("user_id", getUserID())
    c.SetTypedInt("page", 1)
    
    // ✅ 条件操作避免重复设置
    c.SetIfNotExists("default_limit", 20)
    
    // ✅ 高效检索和使用
    if userID, ok := c.GetTypedString("user_id"); ok {
        return processUser(userID)
    }
    
    return map[string]any{"error": "user_id required"}
}
```

### 3. 错误处理模式

```go
// ResponseHandler的错误处理
group.POSTResponse("/users", func(c *mvcContext.Context) any {
    defer func() {
        if r := recover(); r != nil {
            c.AddError(fmt.Errorf("panic: %v", r))
            c.RequestContext().SetStatusCode(500)
        }
    }()
    
    // 业务逻辑...
    if err := validateInput(); err != nil {
        c.RequestContext().SetStatusCode(400)
        return map[string]any{
            "success": false,
            "error": err.Error(),
        }
    }
    
    return map[string]any{"success": true}
})

// StreamHandler的错误处理
group.GETStream("/data", func(c *mvcContext.Context, dataChan chan<- []byte) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Stream handler panic: %v", r)
        }
    }()
    
    // 流式处理逻辑...
    if err := processStream(dataChan); err != nil {
        return fmt.Errorf("stream processing failed: %w", err)
    }
    
    return nil
})
```

## 🚀 迁移指南

### 从传统HandlerFunc迁移

```go
// 旧方式 (仍然支持)
app.RouterPrefix("/old", controller, "OldHandler", "GET:/old")
func (c *Controller) OldHandler() {
    // 传统控制器方法
}

// 新方式 - 推荐
group.GETResponse("/new", func(c *mvcContext.Context) any {
    // 使用增强Context和新Handler类型
    c.Set("migrated", true)
    return map[string]any{"message": "Using new handler system"}
})
```

### 向后兼容性

YYHertz完全支持现有的控制器和路由系统，可以渐进式迁移：

1. 现有代码继续工作
2. 新功能使用多Handler类型
3. 逐步迁移核心业务逻辑
4. 享受性能提升和新特性

---

**🎉 多Handler类型系统让您的应用更快、更灵活、更现代！**