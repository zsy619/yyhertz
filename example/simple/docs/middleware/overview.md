# 🔌 中间件概览

YYHertz v2.0 引入了革命性的**统一中间件系统**，将原有的分散式中间件架构整合为**4层智能架构**，实现了60%的性能提升和100%的向后兼容性。

## 🌟 统一架构优势

### 架构对比

#### ❌ 旧版本架构
```
@framework/middleware  (独立模块)
@framework/mvc/middleware  (MVC专用)
├── 重复实现           ┌── 性能冗余
├── 兼容性问题         ├── 内存浪费  
├── 维护成本高         └── 编译缓存缺失
```

#### ✅ v2.0统一架构
```
@framework/mvc/middleware  (统一模块)
├── 🎯 4层架构设计        ┌── ⚡ 智能编译优化
├── 🚀 性能提升60%        ├── 💾 性能缓存95%+
├── 🔄 100%向后兼容       ├── 🧠 死代码消除
└── 🛠️ 自动适配器        └── 📊 实时监控
```

## 🏗️ 4层中间件架构

### 架构层次图
```
┌─────────────────────────────────────────┐
│           Global Middleware             │
│     全局级 - 影响所有请求                │
├─────────────────────────────────────────┤
│           Group Middleware              │
│     分组级 - 影响路由组/命名空间          │
├─────────────────────────────────────────┤
│           Route Middleware              │
│     路由级 - 影响特定路由                │
├─────────────────────────────────────────┤
│         Controller Middleware           │
│     控制器级 - 方法级别的中间件           │
└─────────────────────────────────────────┘
```

### 执行顺序
```
Request  →  Global  →  Group  →  Route  →  Controller  →  Handler
         ↑                                               ↓
Response ←  Global  ←  Group  ←  Route  ←  Controller  ←  Handler
```

## ⚡ 智能编译系统

### 编译优化流程
```go
// 中间件编译过程
type MiddlewareCompiler struct {
    cache       *sync.Map          // 编译缓存
    optimizer   *CodeOptimizer     // 代码优化器
    eliminator  *DeadCodeEliminator // 死代码消除器
    monitor     *PerformanceMonitor // 性能监控器
}

// 智能编译示例
func (c *MiddlewareCompiler) Compile(middlewares []Middleware) CompiledChain {
    // 1. 检查缓存
    if cached := c.cache.Load(getChainHash(middlewares)); cached != nil {
        c.monitor.CacheHit()
        return cached.(CompiledChain)
    }
    
    // 2. 分析依赖
    dependencies := c.analyzer.AnalyzeDependencies(middlewares)
    
    // 3. 优化顺序
    optimized := c.optimizer.OptimizeOrder(middlewares, dependencies)
    
    // 4. 消除死代码
    eliminated := c.eliminator.RemoveDeadCode(optimized)
    
    // 5. 生成执行链
    compiled := c.generateChain(eliminated)
    
    // 6. 缓存结果
    c.cache.Store(getChainHash(middlewares), compiled)
    
    c.monitor.CompileComplete()
    return compiled
}
```

### 性能对比
```bash
# 基准测试结果
BenchmarkOldMiddleware-8     2000000    650 ns/op   128 B/op    3 allocs/op
BenchmarkNewMiddleware-8     5000000    240 ns/op    48 B/op    1 allocs/op

# 性能提升
响应时间: ↓63% (650ns → 240ns)
内存分配: ↓62% (128B → 48B)  
GC次数:  ↓67% (3 → 1 allocs)
缓存命中: ↑95%+ (几乎无重复编译)
```

## 🔧 基础用法

### 全局中间件
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
    app := mvc.HertzApp
    
    // 全局中间件 - 所有请求都会执行
    app.Use(
        middleware.Recovery(),        // 异常恢复 + 智能错误追踪
        middleware.Logger(),          // 结构化日志 + 性能监控  
        middleware.CORS(),            // 完整CORS策略
        middleware.Compress(),        // 智能响应压缩
        middleware.RateLimit(1000, time.Hour), // 智能限流
    )
    
    // 注册控制器
    app.AutoRouters(&HomeController{})
    
    app.Run(":8888")
}
```

### 分组中间件
```go
func main() {
    app := mvc.HertzApp
    
    // 全局中间件
    app.Use(middleware.Logger(), middleware.Recovery())
    
    // API分组中间件
    api := app.Group("/api")
    api.Use(
        middleware.Auth(middleware.AuthConfig{
            Strategy: middleware.AuthJWT,
        }),
        middleware.RateLimit(500, time.Hour),
        middleware.Metrics(), // API指标收集
    )
    
    // v1版本中间件
    v1 := api.Group("/v1")
    v1.Use(middleware.APIVersion("1.0"))
    
    // v2版本中间件
    v2 := api.Group("/v2") 
    v2.Use(middleware.APIVersion("2.0"))
    
    app.Run(":8888")
}
```

### 路由级中间件
```go
func setupRoutes(app *mvc.Application) {
    // 单个路由中间件
    app.GET("/sensitive", middleware.Auth(), handleSensitive)
    
    // 多个路由中间件
    app.POST("/upload", 
        middleware.Auth(),
        middleware.RateLimit(10, time.Minute),
        middleware.FileSize(50*1024*1024), // 50MB限制
        handleUpload,
    )
    
    // 路由组中间件
    protected := app.Group("/protected")
    protected.Use(
        middleware.Auth(),
        middleware.Permission("protected.access"),
    )
    protected.GET("/data", handleProtectedData)
    protected.POST("/action", handleProtectedAction)
}
```

### 控制器级中间件
```go
type UserController struct {
    mvc.BaseController
}

// 控制器级中间件 - 在控制器方法中使用
func (c *UserController) GetProfile() {
    // 方法级认证检查
    if !c.IsAuthenticated() {
        c.Error(401, "请先登录")
        return
    }
    
    // 方法级权限检查
    if !c.HasPermission("user.profile.read") {
        c.Error(403, "权限不足")
        return
    }
    
    // 业务逻辑
    user := c.GetCurrentUser()
    c.JSON(user)
}

// 使用中间件注解 (规划中功能)
// @Middleware(Auth, Permission("user.profile.read"))
func (c *UserController) GetSensitiveData() {
    c.JSON(map[string]any{"data": "sensitive"})
}
```

## 🛠️ 内置中间件完整列表

### 核心中间件
```go
// 异常恢复 - 增强版
middleware.Recovery()
middleware.Recovery(middleware.RecoveryConfig{
    EnableStackTrace: true,
    LogLevel:        "error",
    CustomHandler:   customRecoveryHandler,
})

// 智能日志 - 结构化
middleware.Logger()
middleware.Logger(middleware.LoggerConfig{
    Format: "[${time}] ${status} - ${method} ${path} (${latency})",
    Output: middleware.LoggerOutputFile("./logs/access.log"),
    EnableMetrics: true,  // 启用性能指标
    SanitizeHeaders: []string{"Authorization"}, // 敏感头脱敏
})

// 跨域支持 - 完整策略
middleware.CORS()
middleware.CORS(middleware.CORSConfig{
    AllowOrigins:     []string{"*"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:          12 * time.Hour,
})
```

### 安全中间件
```go
// 认证中间件 - 多策略
middleware.Auth(middleware.AuthConfig{
    Strategy:  middleware.AuthJWT,
    TokenKey:  "Authorization", 
    UserKey:   "user",
    SkipPaths: []string{"/login", "/register"},
    OnError:   customAuthErrorHandler,
})

// Basic认证
middleware.BasicAuth(map[string]string{
    "admin": "secret123",
    "user":  "password",
})

// 权限控制
middleware.Permission("user.read")
middleware.Permission(middleware.PermissionConfig{
    Required: []string{"user.read", "user.write"},
    Mode:     middleware.PermissionModeAny, // Any, All
})

// 限流中间件 - 智能算法
middleware.RateLimit(100, time.Minute)
middleware.RateLimit(middleware.RateLimitConfig{
    Max:        100,
    Duration:   time.Minute,
    Algorithm:  middleware.TokenBucket, // TokenBucket, SlidingWindow
    KeyFunc:    middleware.RateLimitByIP,
    OnExceeded: customRateLimitHandler,
})
```

### 性能中间件
```go
// 响应压缩 - 智能协商
middleware.Compress()
middleware.Compress(middleware.CompressConfig{
    Level:     middleware.BestCompression,
    Types:     []string{"text/html", "application/json"},
    MinLength: 1024, // 最小压缩大小
})

// 缓存中间件
middleware.Cache(5 * time.Minute)
middleware.Cache(middleware.CacheConfig{
    TTL:        5 * time.Minute,
    KeyFunc:    middleware.CacheKeyByPath,
    Store:      middleware.NewRedisStore("localhost:6379"),
    Condition:  middleware.CacheOnlyGET,
})

// 超时控制 - 渐进式取消
middleware.Timeout(30 * time.Second)
middleware.Timeout(middleware.TimeoutConfig{
    Duration:     30 * time.Second,
    ErrorHandler: timeoutErrorHandler,
    CleanupFunc:  timeoutCleanup,
})
```

### 监控中间件
```go
// 性能指标 - Prometheus兼容
middleware.Metrics()
middleware.Metrics(middleware.MetricsConfig{
    Path:      "/metrics",
    Namespace: "yyhertz",
    Subsystem: "http",
    Labels:    []string{"method", "path", "status"},
})

// 链路追踪 - OpenTelemetry
middleware.Tracing()
middleware.Tracing(middleware.TracingConfig{
    ServiceName:    "yyhertz-app",
    ServiceVersion: "2.0.0",
    Endpoint:       "http://jaeger:14268/api/traces",
    SampleRate:     0.1, // 10%采样率
})

// 健康检查
middleware.HealthCheck("/health")
middleware.HealthCheck(middleware.HealthConfig{
    Path: "/health",
    Checks: []middleware.HealthChecker{
        middleware.DatabaseHealthCheck,
        middleware.RedisHealthCheck,
        middleware.CustomHealthCheck(func() error {
            // 自定义健康检查逻辑
            return nil
        }),
    },
})
```

## 📊 性能监控与分析

### 实时性能监控
```go
package main

import (
    "log"
    "time"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
    app := mvc.HertzApp
    
    // 启用性能监控
    monitor := middleware.NewPerformanceMonitor(middleware.MonitorConfig{
        ReportInterval: 10 * time.Second,
        MetricsPath:    "/internal/metrics",
        EnableProfile:  true,
    })
    
    app.Use(monitor.Middleware())
    
    // 启动监控报告
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            stats := monitor.GetStats()
            log.Printf("中间件性能报告: %+v", stats)
        }
    }()
    
    app.Run(":8888")
}

// 性能统计结构
type PerformanceStats struct {
    // 编译缓存统计
    CacheHitRate        float64       `json:"cache_hit_rate"`
    TotalCompilations   int64         `json:"total_compilations"`
    CacheHits          int64         `json:"cache_hits"`
    CacheMisses        int64         `json:"cache_misses"`
    
    // 执行性能统计  
    AverageLatency     time.Duration `json:"average_latency"`
    P50Latency         time.Duration `json:"p50_latency"`
    P95Latency         time.Duration `json:"p95_latency"`
    P99Latency         time.Duration `json:"p99_latency"`
    
    // 内存使用统计
    MemoryUsage        int64         `json:"memory_usage"`
    MemorySaved        int64         `json:"memory_saved"`
    GCReductions       int64         `json:"gc_reductions"`
    
    // 错误统计
    TotalErrors        int64         `json:"total_errors"`
    RecoveredPanics    int64         `json:"recovered_panics"`
    TimeoutErrors      int64         `json:"timeout_errors"`
}
```

### Prometheus集成
```go
// 暴露Prometheus指标
func setupMetrics(app *mvc.Application) {
    // 自定义指标
    httpDuration := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "yyhertz",
            Subsystem: "middleware",
            Name:      "request_duration_seconds",
            Help:      "HTTP request latencies in seconds.",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequests := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "yyhertz",
            Subsystem: "middleware", 
            Name:      "requests_total",
            Help:      "Total number of HTTP requests.",
        },
        []string{"method", "path", "status"},
    )
    
    // 注册指标
    prometheus.MustRegister(httpDuration, httpRequests)
    
    // 中间件指标收集
    app.Use(func(c *mvc.Context) {
        start := time.Now()
        path := c.Request.URI().Path()
        method := string(c.Request.Header.Method())
        
        c.Next()
        
        status := strconv.Itoa(c.Response.StatusCode())
        duration := time.Since(start).Seconds()
        
        httpDuration.WithLabelValues(method, path, status).Observe(duration)
        httpRequests.WithLabelValues(method, path, status).Inc()
    })
    
    // 暴露指标端点
    app.GET("/metrics", func(c *mvc.Context) {
        handler := promhttp.Handler()
        handler.ServeHTTP(c.Response, c.Request)
    })
}
```

## 🔄 兼容性适配器

### 自动适配机制
```go
// 旧版本中间件自动适配
// 框架内部自动处理，用户无需修改代码

// 旧写法 - 仍然有效
import "github.com/zsy619/yyhertz/framework/middleware" // 自动重定向
app.Use(middleware.RecoveryMiddleware()) // 自动适配到 Recovery()
app.Use(middleware.LoggerMiddleware())   // 自动适配到 Logger()
app.Use(middleware.CORSMiddleware())     // 自动适配到 CORS()

// 新写法 - 推荐使用
import "github.com/zsy619/yyhertz/framework/mvc/middleware"
app.Use(middleware.Recovery()) // 原生统一API
app.Use(middleware.Logger())   // 更好的性能
app.Use(middleware.CORS())     // 更多配置选项
```

### 迁移指南
```go
// ===== 迁移前 (v1.x) =====
import (
    oldMw "github.com/zsy619/yyhertz/framework/middleware"
    mvcMw "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

app.Use(
    oldMw.RecoveryMiddleware(),
    mvcMw.LoggerMiddleware(),  // 混用不同模块
    oldMw.CORSMiddleware(),
)

// ===== 迁移后 (v2.0) =====
import "github.com/zsy619/yyhertz/framework/mvc/middleware"

app.Use(
    middleware.Recovery(),  // 统一模块
    middleware.Logger(),    // 更好性能
    middleware.CORS(),      // 更多功能
)
```

## 🎯 最佳实践

### 1. 中间件顺序优化
```go
// 推荐的中间件顺序
app.Use(
    middleware.Recovery(),     // 1. 异常恢复 (最外层)
    middleware.Logger(),       // 2. 日志记录
    middleware.Metrics(),      // 3. 性能指标
    middleware.CORS(),         // 4. 跨域处理
    middleware.Compress(),     // 5. 响应压缩
    middleware.RateLimit(),    // 6. 限流控制
    middleware.Auth(),         // 7. 身份认证
    middleware.Permission(),   // 8. 权限控制
    middleware.Cache(),        // 9. 缓存处理 (最内层)
)
```

### 2. 条件中间件
```go
// 基于环境的条件中间件
func conditionalMiddleware() mvc.HandlerFunc {
    if config.IsProduction() {
        return middleware.RateLimit(1000, time.Hour)
    }
    return func(c *mvc.Context) { c.Next() } // 开发环境跳过限流
}

// 基于路径的条件中间件  
func pathBasedAuth() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        path := c.Request.URI().Path()
        
        // 公开路径跳过认证
        if isPublicPath(path) {
            c.Next()
            return
        }
        
        // 执行认证检查
        middleware.Auth()(c)
    }
}
```

### 3. 错误处理策略
```go
// 统一错误处理中间件
func unifiedErrorHandler() mvc.HandlerFunc {
    return mvc.CustomRecovery(func(c *mvc.Context, recovered interface{}) {
        var err error
        
        switch x := recovered.(type) {
        case string:
            err = errors.New(x)
        case error:
            err = x
        default:
            err = errors.New("unknown error")
        }
        
        // 记录错误
        log.Error("Middleware Error", log.Fields{
            "error":  err.Error(),
            "path":   c.Request.URI().Path(),
            "method": string(c.Request.Header.Method()),
        })
        
        // 返回统一错误格式
        c.AbortWithStatusJSON(500, map[string]any{
            "error":     "Internal Server Error",
            "message":   err.Error(),
            "timestamp": time.Now().Unix(),
            "path":      c.Request.URI().Path(),
        })
    })
}
```

## 📖 下一步学习

现在您已经了解了YYHertz统一中间件系统的强大功能，建议继续学习：

1. 🛡️ [内置中间件](/home/builtin-middleware) - 掌握所有内置中间件的使用
2. 🔧 [自定义中间件](/home/custom-middleware) - 学习编写自己的中间件
3. ⚙️ [中间件配置](/home/middleware-config) - 深入了解配置选项
4. 📊 [性能监控](/home/performance) - 监控中间件性能表现

---

**🚀 统一中间件系统是YYHertz v2.0的核心创新，让您的应用性能和开发效率都得到质的提升！**