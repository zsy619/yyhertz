# 🔌 内置中间件

YYHertz MVC框架提供了丰富的内置中间件，涵盖了Web开发中常见的需求。这些中间件经过优化，可以直接使用，也可以作为自定义中间件的参考实现。

## 🌟 中间件概览

### 核心中间件
| 中间件 | 功能 | 使用场景 |
|--------|------|----------|
| Logger | 请求日志记录 | 调试、监控、审计 |
| Recovery | 异常恢复 | 防止应用崩溃 |
| CORS | 跨域请求支持 | API开发、前后端分离 |
| Security | 安全头设置 | 安全防护 |

### 功能中间件
| 中间件 | 功能 | 使用场景 |
|--------|------|----------|
| RateLimit | 请求频率限制 | 防止恶意攻击 |
| Cache | 响应缓存 | 性能优化 |
| Compress | 响应压缩 | 减少传输大小 |
| Auth | 身份认证 | 权限控制 |

## 📚 详细说明

### 1. Logger中间件

#### 功能特性
- **📝 详细日志记录** - 记录请求方法、路径、状态码、响应时间
- **🎨 彩色输出** - 支持控制台彩色日志
- **📄 文件输出** - 支持日志文件输出
- **🎯 自定义格式** - 可配置日志格式

#### 使用方法

```go
import "github.com/zsy619/yyhertz/framework/middleware"

// 使用默认配置
app.Use(middleware.LoggerMiddleware())

// 自定义配置
app.Use(middleware.LoggerMiddleware(middleware.LoggerConfig{
    // 自定义日志格式
    Format: "${time} | ${status} | ${latency} | ${ip} | ${method} ${path}",
    
    // 日志文件路径
    Output: "logs/access.log",
    
    // 是否启用彩色输出
    EnableColor: true,
    
    // 跳过特定路径
    SkipPaths: []string{"/health", "/metrics"},
}))
```

#### 日志格式变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `${time}` | 请求时间 | `2024-01-01 12:00:00` |
| `${status}` | HTTP状态码 | `200` |
| `${latency}` | 响应时间 | `15ms` |
| `${ip}` | 客户端IP | `192.168.1.1` |
| `${method}` | HTTP方法 | `GET` |
| `${path}` | 请求路径 | `/api/users` |
| `${user_agent}` | 用户代理 | `Mozilla/5.0...` |

### 2. Recovery中间件

#### 功能特性
- **🛡️ Panic恢复** - 捕获并恢复应用panic
- **📊 错误统计** - 记录错误发生次数
- **🔔 错误通知** - 支持错误通知机制
- **🐛 调试信息** - 开发环境显示详细错误信息

#### 使用方法

```go
// 使用默认配置
app.Use(middleware.RecoveryMiddleware())

// 自定义配置
app.Use(middleware.RecoveryMiddleware(middleware.RecoveryConfig{
    // 是否在响应中显示错误详情（仅开发环境）
    ShowErrorDetails: true,
    
    // 错误处理函数
    ErrorHandler: func(c context.Context, ctx *app.RequestContext, err error) {
        // 记录错误日志
        log.Printf("Panic recovered: %v", err)
        
        // 返回错误响应
        ctx.JSON(500, map[string]interface{}{
            "error": "Internal Server Error",
            "message": "服务器内部错误",
        })
    },
    
    // 错误通知函数
    NotifyFunc: func(err error, stack string) {
        // 发送错误通知（如邮件、钉钉等）
        sendErrorNotification(err, stack)
    },
}))
```

### 3. CORS中间件

#### 功能特性
- **🌍 跨域支持** - 完整的CORS协议支持
- **🔧 灵活配置** - 支持详细的CORS配置
- **🚀 预检请求** - 自动处理OPTIONS预检请求
- **🔒 安全控制** - 精确控制跨域权限

#### 使用方法

```go
// 使用默认配置（允许所有来源）
app.Use(middleware.CORSMiddleware())

// 自定义配置
app.Use(middleware.CORSMiddleware(middleware.CORSConfig{
    // 允许的来源
    AllowOrigins: []string{
        "https://example.com",
        "https://app.example.com",
    },
    
    // 允许的HTTP方法
    AllowMethods: []string{
        "GET", "POST", "PUT", "DELETE", "OPTIONS",
    },
    
    // 允许的请求头
    AllowHeaders: []string{
        "Origin", "Content-Type", "Accept", "Authorization",
        "X-Requested-With", "X-CSRF-Token",
    },
    
    // 暴露的响应头
    ExposeHeaders: []string{
        "Content-Length", "Content-Type",
    },
    
    // 是否允许携带凭证
    AllowCredentials: true,
    
    // 预检请求缓存时间
    MaxAge: 12 * time.Hour,
}))
```

#### 开发环境配置

```go
// 开发环境：允许所有来源
if config.IsDevelopment() {
    app.Use(middleware.CORSMiddleware(middleware.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"*"},
        AllowHeaders: []string{"*"},
    }))
}
```

### 4. Security中间件

#### 功能特性
- **🔒 安全头设置** - 设置常见的安全响应头
- **🛡️ XSS防护** - 防止跨站脚本攻击
- **🚫 内容嗅探防护** - 防止MIME类型嗅探
- **📋 内容安全策略** - CSP头设置

#### 使用方法

```go
app.Use(middleware.SecurityMiddleware(middleware.SecurityConfig{
    // X-Frame-Options
    FrameOptions: "SAMEORIGIN",
    
    // X-Content-Type-Options
    ContentTypeNosniff: true,
    
    // X-XSS-Protection
    XSSProtection: "1; mode=block",
    
    // Strict-Transport-Security
    HSTSMaxAge: 31536000, // 1年
    HSTSIncludeSubdomains: true,
    
    // Content-Security-Policy
    CSP: "default-src 'self'; script-src 'self' 'unsafe-inline';",
    
    // Referrer-Policy
    ReferrerPolicy: "strict-origin-when-cross-origin",
}))
```

### 5. RateLimit中间件

#### 功能特性
- **⏱️ 请求频率限制** - 基于时间窗口的限流
- **🎯 多维度限制** - 支持IP、用户、API等维度
- **💾 多种存储** - 内存、Redis存储支持
- **📊 限流统计** - 提供限流统计信息

#### 使用方法

```go
// 基于IP的限流
app.Use(middleware.RateLimitMiddleware(middleware.RateLimitConfig{
    // 时间窗口（秒）
    Window: 60,
    
    // 最大请求数
    Limit: 100,
    
    // 限流key生成函数
    KeyFunc: func(c context.Context, ctx *app.RequestContext) string {
        return ctx.ClientIP()
    },
    
    // 超限处理函数
    LimitReachedHandler: func(c context.Context, ctx *app.RequestContext) {
        ctx.JSON(429, map[string]interface{}{
            "error": "Too Many Requests",
            "message": "请求过于频繁，请稍后再试",
        })
    },
}))

// 基于用户的限流
app.Use(middleware.RateLimitMiddleware(middleware.RateLimitConfig{
    Window: 3600, // 1小时
    Limit: 1000,
    KeyFunc: func(c context.Context, ctx *app.RequestContext) string {
        userID := getUserID(ctx) // 从token或session获取用户ID
        return fmt.Sprintf("user:%s", userID)
    },
}))
```

### 6. Cache中间件

#### 功能特性
- **🚀 响应缓存** - 缓存GET请求的响应
- **🕒 TTL控制** - 灵活的缓存过期时间
- **🎯 条件缓存** - 基于条件的缓存策略
- **💾 多种后端** - 内存、Redis缓存支持

#### 使用方法

```go
// 缓存GET请求
app.Use(middleware.CacheMiddleware(middleware.CacheConfig{
    // 缓存TTL
    TTL: 5 * time.Minute,
    
    // 缓存条件
    CacheCondition: func(c context.Context, ctx *app.RequestContext) bool {
        // 只缓存GET请求
        return string(ctx.Method()) == "GET"
    },
    
    // 缓存key生成
    KeyFunc: func(c context.Context, ctx *app.RequestContext) string {
        return fmt.Sprintf("cache:%s:%s", 
            ctx.Method(), ctx.Request.URI().String())
    },
}))

// API接口缓存
apiGroup := app.Group("/api")
apiGroup.Use(middleware.CacheMiddleware(middleware.CacheConfig{
    TTL: 30 * time.Second,
    CacheCondition: func(c context.Context, ctx *app.RequestContext) bool {
        // 缓存状态码为200的响应
        return ctx.Response.StatusCode() == 200
    },
}))
```

### 7. Compress中间件

#### 功能特性
- **📦 响应压缩** - 支持gzip、brotli压缩
- **🎯 条件压缩** - 基于内容类型和大小的压缩
- **⚡ 性能优化** - 减少传输大小，提升加载速度
- **🔧 灵活配置** - 可配置压缩级别和类型

#### 使用方法

```go
app.Use(middleware.CompressMiddleware(middleware.CompressConfig{
    // 压缩级别（1-9）
    Level: 6,
    
    // 最小压缩大小
    MinLength: 1024,
    
    // 压缩的内容类型
    ContentTypes: []string{
        "text/html",
        "text/css", 
        "text/javascript",
        "application/json",
        "application/xml",
    },
    
    // 跳过压缩的路径
    SkipPaths: []string{
        "/api/upload",
        "/api/download",
    },
}))
```

### 8. Auth中间件

#### 功能特性
- **🔑 多种认证方式** - JWT、Session、Basic Auth
- **👥 权限控制** - 基于角色的访问控制
- **🚫 路径保护** - 灵活的路径保护规则
- **🔄 Token刷新** - 自动token刷新机制

#### 使用方法

```go
// JWT认证
app.Use(middleware.JWTAuthMiddleware(middleware.JWTConfig{
    // JWT密钥
    SecretKey: "your-secret-key",
    
    // Token来源
    TokenLookup: "header:Authorization,query:token,cookie:token",
    
    // Token前缀
    TokenPrefix: "Bearer ",
    
    // 跳过认证的路径
    SkipPaths: []string{
        "/api/auth/login",
        "/api/auth/register",
        "/api/public/*",
    },
    
    // 认证失败处理
    Unauthorized: func(c context.Context, ctx *app.RequestContext, err error) {
        ctx.JSON(401, map[string]interface{}{
            "error": "Unauthorized",
            "message": "请先登录",
        })
    },
}))

// Session认证
app.Use(middleware.SessionAuthMiddleware(middleware.SessionConfig{
    // Session存储
    Store: sessionStore,
    
    // Session名称
    SessionName: "session_id",
    
    // 用户key
    UserKey: "user_id",
    
    // 登录检查函数
    CheckLogin: func(c context.Context, ctx *app.RequestContext) bool {
        session := getSession(ctx)
        return session.Get("user_id") != nil
    },
}))
```

## 🔄 中间件组合使用

### 推荐配置组合

```go
func setupMiddlewares(app *hertz.Engine) {
    // 1. 基础中间件（必须）
    app.Use(
        middleware.RecoveryMiddleware(),  // 异常恢复
        middleware.LoggerMiddleware(),    // 日志记录
    )
    
    // 2. 安全中间件
    app.Use(
        middleware.SecurityMiddleware(), // 安全头
        middleware.CORSMiddleware(),     // 跨域支持
    )
    
    // 3. 性能中间件
    app.Use(
        middleware.CompressMiddleware(), // 响应压缩
        middleware.CacheMiddleware(),    // 响应缓存
    )
    
    // 4. 限流中间件
    app.Use(
        middleware.RateLimitMiddleware(), // 请求限流
    )
}
```

### 分组中间件配置

```go
// 公共API（无需认证）
publicAPI := app.Group("/api/public")
publicAPI.Use(
    middleware.RateLimitMiddleware(publicRateLimit),
    middleware.CacheMiddleware(publicCacheConfig),
)

// 用户API（需要认证）
userAPI := app.Group("/api/user")
userAPI.Use(
    middleware.JWTAuthMiddleware(jwtConfig),
    middleware.RateLimitMiddleware(userRateLimit),
)

// 管理API（需要管理员权限）
adminAPI := app.Group("/api/admin")
adminAPI.Use(
    middleware.JWTAuthMiddleware(jwtConfig),
    middleware.RBACMiddleware(adminRole),
    middleware.AuditLogMiddleware(),
)
```

## 🔧 配置最佳实践

### 1. 开发环境配置

```go
if config.IsDevelopment() {
    app.Use(
        middleware.LoggerMiddleware(middleware.LoggerConfig{
            EnableColor: true,
            Format: "${time} | ${status} | ${latency} | ${method} ${path}",
        }),
        middleware.CORSMiddleware(middleware.CORSConfig{
            AllowOrigins: []string{"*"},
            AllowMethods: []string{"*"},
            AllowHeaders: []string{"*"},
        }),
    )
}
```

### 2. 生产环境配置

```go
if config.IsProduction() {
    app.Use(
        middleware.LoggerMiddleware(middleware.LoggerConfig{
            Output: "logs/access.log",
            Format: "${time} | ${ip} | ${status} | ${latency} | ${method} ${path}",
        }),
        middleware.SecurityMiddleware(productionSecurityConfig),
        middleware.RateLimitMiddleware(productionRateLimit),
    )
}
```

### 3. 性能监控配置

```go
// 添加性能监控中间件
app.Use(
    middleware.MetricsMiddleware(), // 指标收集
    middleware.TracingMiddleware(), // 链路追踪
)

// 健康检查端点
app.GET("/health", healthCheckHandler)
app.GET("/metrics", metricsHandler)
```

## 📊 监控与调试

### 中间件性能监控

```go
// 中间件执行时间统计
type MiddlewareStats struct {
    Name        string
    TotalTime   time.Duration
    RequestCount int64
    ErrorCount   int64
}

// 监控中间件
func MonitoringMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        start := time.Now()
        
        ctx.Next(c)
        
        duration := time.Since(start)
        status := ctx.Response.StatusCode()
        
        // 记录统计信息
        recordMiddlewareStats(duration, status)
    }
}
```

### 调试信息输出

```go
// 调试中间件（仅开发环境）
func DebugMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        if config.IsDevelopment() {
            log.Printf("Request: %s %s", ctx.Method(), ctx.Path())
            log.Printf("Headers: %v", ctx.Request.Header)
        }
        
        ctx.Next(c)
        
        if config.IsDevelopment() {
            log.Printf("Response: %d", ctx.Response.StatusCode())
        }
    }
}
```

## 🛠️ 故障排除

### 常见问题

#### 1. 中间件执行顺序
```go
// 错误：Recovery中间件应该在最前面
app.Use(middleware.LoggerMiddleware())
app.Use(middleware.RecoveryMiddleware()) // 放在Logger后面无法捕获Logger的panic

// 正确：Recovery中间件在最前面
app.Use(middleware.RecoveryMiddleware())
app.Use(middleware.LoggerMiddleware())
```

#### 2. CORS配置问题
```go
// 检查预检请求是否正确处理
if string(ctx.Method()) == "OPTIONS" {
    log.Printf("OPTIONS request received for: %s", ctx.Path())
}
```

#### 3. 认证中间件问题
```go
// 确保跳过路径配置正确
SkipPaths: []string{
    "/api/auth/login",    // 登录接口
    "/api/auth/register", // 注册接口
    "/api/public/*",      // 公共接口
    "/health",            // 健康检查
}
```

## 📚 扩展阅读

- [自定义中间件开发](./custom.md)
- [中间件配置指南](./config.md)
- [性能优化建议](../dev-tools/performance.md)
- [安全防护最佳实践](../security/best-practices.md)

---

> 💡 **提示**: 合理使用内置中间件可以大幅提升开发效率和应用安全性。建议根据实际需求选择合适的中间件组合。
