# YYHertz 中间件开发指南

<div align="center">

🔧 **中间件系统完全指南** | 从使用到自定义开发

</div>

---

## 📋 目录

- [中间件基础概念](#中间件基础概念)
- [内置中间件使用](#内置中间件使用)
- [自定义中间件开发](#自定义中间件开发)
- [中间件执行顺序](#中间件执行顺序)
- [高级中间件技巧](#高级中间件技巧)
- [中间件最佳实践](#中间件最佳实践)

---

## 🎯 中间件基础概念

### 什么是中间件？
中间件是位于HTTP请求和响应处理之间的组件，可以在请求到达处理函数之前或响应返回客户端之前执行特定的逻辑。

### YYHertz中间件特性
- **🔗 链式执行**: 中间件按顺序执行，形成处理链
- **⚡ 高性能**: 智能编译优化，零性能损失
- **🎯 灵活配置**: 支持全局、路由组、单路由级别配置
- **🛡️ 丰富内置**: 提供20+常用中间件开箱即用

### 中间件执行流程
```
请求 → 中间件1 → 中间件2 → 中间件3 → 处理函数 → 中间件3 → 中间件2 → 中间件1 → 响应
```

---

## 🏗️ 内置中间件使用

### 1. 异常恢复中间件

```go
import "github.com/zsy619/yyhertz/framework/mvc/middleware"

func main() {
    app := mvc.HertzApp
    
    // 基础异常恢复
    app.Use(middleware.Recovery())
    
    // 自定义异常恢复
    app.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
        EnableStackTrace: true,  // 启用堆栈追踪
        EnablePrintStack: false, // 不打印堆栈到控制台
        RecoveryHandler: func(c *mvc.Context, recovered interface{}) {
            // 自定义恢复处理逻辑
            logger.Error("Panic recovered", "error", recovered)
            
            c.JSON(500, map[string]any{
                "error":   "Internal server error",
                "code":    500,
                "message": "服务器内部错误",
            })
        },
    }))
    
    app.Run(":8888")
}
```

### 2. 日志中间件

```go
func main() {
    app := mvc.HertzApp
    
    // 基础日志中间件
    app.Use(middleware.Logger())
    
    // 自定义日志格式
    app.Use(middleware.LoggerWithFormatter(func(param middleware.LogFormatterParams) string {
        return fmt.Sprintf("[%s] %s %s %d %v \"%s\" \"%s\"\n",
            param.TimeStamp.Format("2006/01/02 - 15:04:05"),
            param.Method,
            param.Path,
            param.StatusCode,
            param.Latency,
            param.Request.UserAgent(),
            param.ErrorMessage,
        )
    }))
    
    // 输出到文件
    logFile, _ := os.OpenFile("logs/access.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    app.Use(middleware.LoggerWithWriter(logFile))
    
    app.Run(":8888")
}
```

### 3. CORS跨域中间件

```go
func main() {
    app := mvc.HertzApp
    
    // 默认CORS配置
    app.Use(middleware.CORS())
    
    // 自定义CORS配置
    app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{
            "https://example.com", 
            "https://app.example.com",
        },
        AllowMethods: []string{
            "GET", "POST", "PUT", "DELETE", "OPTIONS",
        },
        AllowHeaders: []string{
            "Origin", "Content-Length", "Content-Type", 
            "Authorization", "X-Requested-With",
        },
        ExposeHeaders: []string{
            "Content-Length", "Content-Type",
        },
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
    
    app.Run(":8888")
}
```

### 4. 认证中间件

```go
// JWT认证中间件
func setupJWTAuth(app *mvc.Application) {
    jwtConfig := middleware.JWTConfig{
        SigningKey:  []byte("your-secret-key"),
        TokenLookup: "header:Authorization,query:token,cookie:jwt",
        AuthScheme:  "Bearer",
        
        // 跳过认证的路径
        Skipper: func(c *mvc.Context) bool {
            skipPaths := []string{"/login", "/register", "/health"}
            for _, path := range skipPaths {
                if strings.HasPrefix(c.Path(), path) {
                    return true
                }
            }
            return false
        },
        
        // 认证成功回调
        SuccessHandler: func(c *mvc.Context) {
            userID := c.Get("user_id").(string)
            logger.Info("User authenticated", "user_id", userID)
        },
        
        // 认证失败回调
        ErrorHandler: func(c *mvc.Context, err error) {
            c.JSON(401, map[string]any{
                "error":   "Unauthorized",
                "message": "Token无效或已过期",
            })
        },
    }
    
    app.Use(middleware.JWTAuth(jwtConfig))
}

// Basic认证中间件
func setupBasicAuth(app *mvc.Application) {
    accounts := map[string]string{
        "admin":  "admin123",
        "user":   "user123",
        "guest":  "guest123",
    }
    
    app.Use(middleware.BasicAuth(accounts))
}
```

### 5. 限流中间件

```go
func main() {
    app := mvc.HertzApp
    
    // 简单限流 - 每分钟100次请求
    app.Use(middleware.RateLimit(100, time.Minute))
    
    // 高级限流配置
    app.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
        Rate:   1000,          // 每小时1000次
        Window: time.Hour,     // 时间窗口
        
        // 自定义键函数 - 基于用户ID限流
        KeyFunc: func(c *mvc.Context) string {
            userID := c.GetHeader("X-User-ID")
            if userID == "" {
                return c.ClientIP() // 没有用户ID则使用IP
            }
            return "user:" + userID
        },
        
        // 限流触发时的处理
        ErrorHandler: func(c *mvc.Context) {
            c.JSON(429, map[string]any{
                "error":   "Rate limit exceeded",
                "message": "请求过于频繁，请稍后再试",
            })
        },
        
        // 自定义存储 (默认使用内存存储)
        Store: middleware.NewRedisStore(redisClient),
    }))
    
    app.Run(":8888")
}
```

### 6. 压缩中间件

```go
func main() {
    app := mvc.HertzApp
    
    // 启用Gzip压缩
    app.Use(middleware.Gzip(middleware.DefaultCompression))
    
    // 自定义压缩配置
    app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
        Level: middleware.BestCompression,  // 最高压缩率
        MinLength: 1024,                    // 最小压缩长度
        ExcludedExtensions: []string{       // 排除的文件扩展名
            ".png", ".jpg", ".gif", ".zip", ".pdf",
        },
        ExcludedPaths: []string{            // 排除的路径
            "/api/upload", "/api/download",
        },
    }))
    
    app.Run(":8888")
}
```

---

## 🛠️ 自定义中间件开发

### 1. 基础中间件结构

```go
// 中间件函数签名
type HandlerFunc func(*Context)

// 简单的中间件示例
func SimpleMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 前置处理
        fmt.Println("请求开始处理")
        start := time.Now()
        
        // 调用下一个中间件或处理函数
        c.Next()
        
        // 后置处理
        duration := time.Since(start)
        fmt.Printf("请求处理完成，耗时: %v\n", duration)
    }
}

// 使用中间件
func main() {
    app := mvc.HertzApp
    app.Use(SimpleMiddleware())
    app.Run(":8888")
}
```

### 2. 带配置的中间件

```go
// 中间件配置结构
type RequestIDConfig struct {
    Header    string                     // 响应头名称
    Generator func() string              // ID生成器
    Skipper   func(*mvc.Context) bool    // 跳过条件
}

// 默认配置
var DefaultRequestIDConfig = RequestIDConfig{
    Header: "X-Request-ID",
    Generator: func() string {
        return uuid.New().String()
    },
    Skipper: nil,
}

// 请求ID中间件
func RequestID() mvc.HandlerFunc {
    return RequestIDWithConfig(DefaultRequestIDConfig)
}

func RequestIDWithConfig(config RequestIDConfig) mvc.HandlerFunc {
    // 配置初始化
    if config.Header == "" {
        config.Header = DefaultRequestIDConfig.Header
    }
    if config.Generator == nil {
        config.Generator = DefaultRequestIDConfig.Generator
    }
    
    return func(c *mvc.Context) {
        // 跳过检查
        if config.Skipper != nil && config.Skipper(c) {
            c.Next()
            return
        }
        
        // 检查是否已有请求ID
        requestID := c.GetHeader(config.Header)
        if requestID == "" {
            requestID = config.Generator()
        }
        
        // 设置请求ID
        c.Set("request_id", requestID)
        c.Header(config.Header, requestID)
        
        c.Next()
    }
}

// 使用示例
func main() {
    app := mvc.HertzApp
    
    // 使用默认配置
    app.Use(RequestID())
    
    // 使用自定义配置
    app.Use(RequestIDWithConfig(RequestIDConfig{
        Header: "X-Trace-ID",
        Generator: func() string {
            return fmt.Sprintf("trace-%d", time.Now().UnixNano())
        },
    }))
    
    app.Run(":8888")
}
```

### 3. 数据库连接中间件

```go
import (
    "database/sql"
    "gorm.io/gorm"
)

// 数据库中间件配置
type DBConfig struct {
    DB      *gorm.DB
    Key     string                    // 上下文键名
    Skipper func(*mvc.Context) bool   // 跳过条件
}

// 数据库连接中间件
func DBMiddleware(db *gorm.DB) mvc.HandlerFunc {
    return DBMiddlewareWithConfig(DBConfig{
        DB:  db,
        Key: "db",
    })
}

func DBMiddlewareWithConfig(config DBConfig) mvc.HandlerFunc {
    if config.Key == "" {
        config.Key = "db"
    }
    
    return func(c *mvc.Context) {
        if config.Skipper != nil && config.Skipper(c) {
            c.Next()
            return
        }
        
        // 为每个请求创建新的数据库会话
        session := config.DB.Session(&gorm.Session{})
        
        // 设置到上下文
        c.Set(config.Key, session)
        
        c.Next()
        
        // 清理资源（如果需要）
        // session.Close() // GORM不需要手动关闭
    }
}

// 使用示例
func main() {
    app := mvc.HertzApp
    
    // 初始化数据库
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }
    
    // 应用数据库中间件
    app.Use(DBMiddleware(db))
    
    app.GET("/users", func(c *mvc.Context) {
        // 从上下文获取数据库连接
        db := c.MustGet("db").(*gorm.DB)
        
        var users []User
        db.Find(&users)
        
        c.JSON(users)
    })
    
    app.Run(":8888")
}
```

### 4. 缓存中间件

```go
import (
    "crypto/md5"
    "encoding/hex"
    "time"
)

// 缓存配置
type CacheConfig struct {
    Store     CacheStore                // 缓存存储
    KeyFunc   func(*mvc.Context) string // 缓存键生成函数
    TTL       time.Duration             // 过期时间
    OnlyGET   bool                      // 只缓存GET请求
    Skipper   func(*mvc.Context) bool   // 跳过条件
}

// 缓存存储接口
type CacheStore interface {
    Get(key string) ([]byte, error)
    Set(key string, value []byte, ttl time.Duration) error
    Delete(key string) error
}

// 内存缓存实现
type MemoryCache struct {
    data map[string]cacheItem
    mu   sync.RWMutex
}

type cacheItem struct {
    value  []byte
    expire time.Time
}

func NewMemoryCache() *MemoryCache {
    cache := &MemoryCache{
        data: make(map[string]cacheItem),
    }
    
    // 启动清理协程
    go cache.cleanup()
    
    return cache
}

func (m *MemoryCache) Get(key string) ([]byte, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    item, exists := m.data[key]
    if !exists || time.Now().After(item.expire) {
        return nil, errors.New("cache miss")
    }
    
    return item.value, nil
}

func (m *MemoryCache) Set(key string, value []byte, ttl time.Duration) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.data[key] = cacheItem{
        value:  value,
        expire: time.Now().Add(ttl),
    }
    
    return nil
}

func (m *MemoryCache) Delete(key string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    delete(m.data, key)
    return nil
}

func (m *MemoryCache) cleanup() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        m.mu.Lock()
        now := time.Now()
        for key, item := range m.data {
            if now.After(item.expire) {
                delete(m.data, key)
            }
        }
        m.mu.Unlock()
    }
}

// 缓存中间件
func Cache(store CacheStore, ttl time.Duration) mvc.HandlerFunc {
    return CacheWithConfig(CacheConfig{
        Store: store,
        TTL:   ttl,
        KeyFunc: func(c *mvc.Context) string {
            // 默认键生成：方法+路径+查询参数
            h := md5.New()
            h.Write([]byte(c.Request.Method + c.Request.URL.String()))
            return hex.EncodeToString(h.Sum(nil))
        },
        OnlyGET: true,
    })
}

func CacheWithConfig(config CacheConfig) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 跳过检查
        if config.Skipper != nil && config.Skipper(c) {
            c.Next()
            return
        }
        
        // 只缓存GET请求
        if config.OnlyGET && c.Request.Method != "GET" {
            c.Next()
            return
        }
        
        // 生成缓存键
        key := config.KeyFunc(c)
        
        // 尝试从缓存获取
        if data, err := config.Store.Get(key); err == nil {
            c.Data(200, "application/json", data)
            return
        }
        
        // 缓存未命中，继续处理
        writer := &responseWriter{
            ResponseWriter: c.Writer,
            body:          &bytes.Buffer{},
        }
        c.Writer = writer
        
        c.Next()
        
        // 如果响应成功，缓存结果
        if writer.status >= 200 && writer.status < 300 {
            config.Store.Set(key, writer.body.Bytes(), config.TTL)
        }
    }
}

// 响应写入器包装
type responseWriter struct {
    http.ResponseWriter
    body   *bytes.Buffer
    status int
}

func (w *responseWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(data []byte) (int, error) {
    w.body.Write(data)
    return w.ResponseWriter.Write(data)
}

// 使用示例
func main() {
    app := mvc.HertzApp
    
    // 创建缓存存储
    cache := NewMemoryCache()
    
    // 应用缓存中间件
    app.Use(Cache(cache, 5*time.Minute))
    
    app.GET("/api/data", func(c *mvc.Context) {
        // 模拟耗时操作
        time.Sleep(100 * time.Millisecond)
        
        c.JSON(map[string]any{
            "data":      "some expensive data",
            "timestamp": time.Now().Unix(),
        })
    })
    
    app.Run(":8888")
}
```

---

## ⚡ 中间件执行顺序

### 1. 全局中间件顺序

```go
func main() {
    app := mvc.HertzApp
    
    // 中间件执行顺序：按注册顺序执行
    app.Use(middleware.Recovery())      // 1. 异常恢复（应该最先）
    app.Use(middleware.Logger())        // 2. 访问日志
    app.Use(middleware.CORS())          // 3. 跨域处理
    app.Use(middleware.RateLimit())     // 4. 限流控制
    app.Use(middleware.Auth())          // 5. 身份认证（应该最后）
    
    app.GET("/test", func(c *mvc.Context) {
        c.JSON(map[string]any{"message": "success"})
    })
    
    app.Run(":8888")
}
```

### 2. 中间件执行流程

```go
// 演示中间件执行流程
func DemoMiddleware(name string) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        fmt.Printf("[%s] 前置处理\n", name)
        
        c.Next() // 调用下一个中间件或处理函数
        
        fmt.Printf("[%s] 后置处理\n", name)
    }
}

func main() {
    app := mvc.HertzApp
    
    app.Use(DemoMiddleware("Middleware1"))
    app.Use(DemoMiddleware("Middleware2"))
    app.Use(DemoMiddleware("Middleware3"))
    
    app.GET("/demo", func(c *mvc.Context) {
        fmt.Println("[Handler] 处理请求")
        c.JSON(map[string]any{"message": "demo"})
    })
    
    app.Run(":8888")
}

// 访问 /demo 时的输出:
// [Middleware1] 前置处理
// [Middleware2] 前置处理  
// [Middleware3] 前置处理
// [Handler] 处理请求
// [Middleware3] 后置处理
// [Middleware2] 后置处理
// [Middleware1] 后置处理
```

### 3. 中断中间件链

```go
// 带条件中断的中间件
func AuthMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        token := c.GetHeader("Authorization")
        
        if token == "" {
            c.JSON(401, map[string]any{
                "error": "Missing authorization token",
            })
            c.Abort() // 中断中间件链，不会执行后续中间件和处理函数
            return
        }
        
        // 验证token
        if !validateToken(token) {
            c.JSON(401, map[string]any{
                "error": "Invalid token",
            })
            c.Abort()
            return
        }
        
        // 验证通过，继续执行
        c.Set("user_id", getUserIDFromToken(token))
        c.Next()
    }
}

func main() {
    app := mvc.HertzApp
    
    app.Use(middleware.Logger())
    app.Use(AuthMiddleware())      // 如果认证失败，会中断执行
    app.Use(middleware.RateLimit()) // 认证失败时不会执行到这里
    
    app.GET("/protected", func(c *mvc.Context) {
        userID := c.GetString("user_id")
        c.JSON(map[string]any{
            "message": "Protected resource",
            "user_id": userID,
        })
    })
    
    app.Run(":8888")
}
```

---

## 🎓 高级中间件技巧

### 1. 条件中间件

```go
// 基于条件的中间件装饰器
func ConditionalMiddleware(condition func(*mvc.Context) bool, middleware mvc.HandlerFunc) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        if condition(c) {
            middleware(c)
        } else {
            c.Next()
        }
    }
}

// 使用示例
func main() {
    app := mvc.HertzApp
    
    // 只对API路径应用认证
    app.Use(ConditionalMiddleware(
        func(c *mvc.Context) bool {
            return strings.HasPrefix(c.Request.URL.Path, "/api")
        },
        middleware.Auth(),
    ))
    
    // 只对管理员路径应用CSRF保护
    app.Use(ConditionalMiddleware(
        func(c *mvc.Context) bool {
            return strings.HasPrefix(c.Request.URL.Path, "/admin")
        },
        middleware.CSRF(),
    ))
    
    app.Run(":8888")
}
```

### 2. 中间件组合

```go
// 中间件组合器
func CombineMiddlewares(middlewares ...mvc.HandlerFunc) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 创建一个新的中间件链
        for i, middleware := range middlewares {
            if i == len(middlewares)-1 {
                // 最后一个中间件，直接调用Next
                middleware(c)
            } else {
                // 包装下一个中间件
                next := c.Next
                c.Next = func() {
                    middlewares[i+1](c)
                    c.Next = next
                }
                middleware(c)
                break
            }
        }
    }
}

// 预定义中间件组合
var (
    // API中间件组
    APIMiddlewares = CombineMiddlewares(
        middleware.Recovery(),
        middleware.Logger(),
        middleware.CORS(),
        middleware.RateLimit(100, time.Minute),
        middleware.Auth(),
    )
    
    // Web中间件组
    WebMiddlewares = CombineMiddlewares(
        middleware.Recovery(),
        middleware.Logger(),
        middleware.Session(),
        middleware.CSRF(),
    )
)

func main() {
    app := mvc.HertzApp
    
    // API路由使用API中间件组
    apiGroup := app.Group("/api/v1", APIMiddlewares)
    {
        apiGroup.GET("/users", getUsersAPI)
        apiGroup.POST("/users", createUserAPI)
    }
    
    // Web路由使用Web中间件组
    webGroup := app.Group("/web", WebMiddlewares)
    {
        webGroup.GET("/dashboard", dashboardWeb)
        webGroup.POST("/profile", updateProfileWeb)
    }
    
    app.Run(":8888")
}
```

### 3. 异步中间件

```go
// 异步处理中间件
func AsyncMiddleware(asyncFunc func(*mvc.Context)) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 复制上下文数据
        contextData := make(map[string]interface{})
        c.Keys().Range(func(key, value interface{}) bool {
            if k, ok := key.(string); ok {
                contextData[k] = value
            }
            return true
        })
        
        // 异步执行
        go func() {
            // 创建新的上下文用于异步处理
            asyncCtx := &mvc.Context{
                Request: c.Request.Clone(context.Background()),
            }
            
            // 恢复上下文数据
            for k, v := range contextData {
                asyncCtx.Set(k, v)
            }
            
            asyncFunc(asyncCtx)
        }()
        
        c.Next()
    }
}

// 异步日志中间件
func AsyncLogMiddleware() mvc.HandlerFunc {
    logChannel := make(chan LogEntry, 1000)
    
    // 启动日志处理协程
    go func() {
        for logEntry := range logChannel {
            // 批量写入日志
            writeLogToDB(logEntry)
        }
    }()
    
    return func(c *mvc.Context) {
        start := time.Now()
        
        c.Next()
        
        // 异步记录日志
        select {
        case logChannel <- LogEntry{
            Method:    c.Request.Method,
            Path:      c.Request.URL.Path,
            Status:    c.Writer.Status(),
            Duration:  time.Since(start),
            UserAgent: c.Request.UserAgent(),
            IP:        c.ClientIP(),
            Timestamp: time.Now(),
        }:
        default:
            // 日志通道满了，丢弃日志（或者可以选择阻塞）
        }
    }
}

type LogEntry struct {
    Method    string
    Path      string
    Status    int
    Duration  time.Duration
    UserAgent string
    IP        string
    Timestamp time.Time
}
```

---

## 💡 中间件最佳实践

### 1. 中间件设计原则

```go
// ✅ 好的中间件设计
func GoodMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 1. 快速失败 - 尽早返回错误
        if !validateRequest(c) {
            c.JSON(400, map[string]any{"error": "Invalid request"})
            c.Abort()
            return
        }
        
        // 2. 最小化工作 - 只做必要的处理
        processMinimalLogic(c)
        
        // 3. 清理资源
        defer cleanup()
        
        c.Next()
    }
}

// ❌ 避免的做法
func BadMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 避免：耗时操作
        expensiveOperation()
        
        // 避免：忘记调用Next()
        // c.Next() // 忘记调用会导致请求中断
        
        // 避免：不必要的内存分配
        largeData := make([]byte, 1024*1024)
        processLargeData(largeData)
        
        c.Next()
    }
}
```

### 2. 错误处理最佳实践

```go
// 统一错误处理中间件
func ErrorHandlerMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 记录panic错误
                logger.Error("Panic occurred", "error", err, "stack", debug.Stack())
                
                c.JSON(500, map[string]any{
                    "error": "Internal server error",
                    "code":  500,
                })
                c.Abort()
            }
        }()
        
        c.Next()
        
        // 检查是否有错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            handleError(c, err.Err)
        }
    }
}

func handleError(c *mvc.Context, err error) {
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
    default:
        logger.Error("Unhandled error", "error", err)
        c.JSON(500, map[string]any{
            "error": "Internal server error",
        })
    }
}
```

### 3. 性能优化建议

```go
// 高性能中间件实现
func HighPerformanceMiddleware() mvc.HandlerFunc {
    // 1. 预分配资源
    pool := sync.Pool{
        New: func() interface{} {
            return make([]byte, 1024)
        },
    }
    
    return func(c *mvc.Context) {
        // 2. 使用对象池
        buffer := pool.Get().([]byte)
        defer pool.Put(buffer[:0])
        
        // 3. 避免字符串拼接
        var builder strings.Builder
        builder.Grow(256) // 预分配容量
        
        // 4. 使用快速路径
        if c.Request.Method == "GET" && isCacheableRequest(c) {
            handleCachedRequest(c)
            return
        }
        
        c.Next()
    }
}

// 缓存友好的中间件
func CacheFriendlyMiddleware() mvc.HandlerFunc {
    // 预计算常用值
    commonHeaders := map[string]string{
        "X-Content-Type-Options": "nosniff",
        "X-Frame-Options":        "DENY",
        "X-XSS-Protection":       "1; mode=block",
    }
    
    return func(c *mvc.Context) {
        // 批量设置头部
        for key, value := range commonHeaders {
            c.Header(key, value)
        }
        
        c.Next()
    }
}
```

### 4. 测试中间件

```go
// 中间件测试示例
func TestAuthMiddleware(t *testing.T) {
    // 设置测试环境
    app := mvc.NewTest()
    app.Use(AuthMiddleware())
    
    app.GET("/protected", func(c *mvc.Context) {
        c.JSON(200, map[string]any{"message": "success"})
    })
    
    // 测试用例1：无token
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/protected", nil)
    app.ServeHTTP(w, req)
    
    assert.Equal(t, 401, w.Code)
    
    // 测试用例2：有效token
    w = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    app.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    
    // 测试用例3：无效token
    w = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    app.ServeHTTP(w, req)
    
    assert.Equal(t, 401, w.Code)
}

// 中间件基准测试
func BenchmarkMiddleware(b *testing.B) {
    app := mvc.NewTest()
    app.Use(YourMiddleware())
    
    app.GET("/test", func(c *mvc.Context) {
        c.JSON(200, map[string]any{"ok": true})
    })
    
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/test", nil)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        app.ServeHTTP(w, req)
    }
}
```

---

<div align="center">

**🔧 掌握中间件开发，让你的应用更强大！**

**合理使用中间件是构建高质量Web应用的关键 ⚡**

</div>