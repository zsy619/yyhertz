# 🛠️ 自定义中间件

自定义中间件是扩展YYHertz MVC框架功能的核心方式。通过编写自定义中间件，您可以实现特定的业务逻辑、增强安全性、优化性能等。

## 🎯 中间件基础

### 中间件函数签名

```go
type HandlerFunc func(c context.Context, ctx *app.RequestContext)

// 中间件函数返回HandlerFunc
type MiddlewareFunc func() HandlerFunc

// 带配置的中间件函数
type ConfigurableMiddleware func(config Config) HandlerFunc
```

### 基本中间件结构

```go
func MyMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        // 前置处理：在请求处理前执行
        // 例如：验证、预处理等
        
        // 调用下一个中间件或处理器
        ctx.Next(c)
        
        // 后置处理：在请求处理后执行
        // 例如：清理、统计等
    }
}
```

## 🚀 实战示例

### 1. 简单的日志中间件

```go
package middleware

import (
    "context"
    "log"
    "time"
    
    "github.com/cloudwego/hertz/pkg/app"
)

// SimpleLogger 简单日志中间件
func SimpleLogger() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        start := time.Now()
        path := string(ctx.Path())
        method := string(ctx.Method())
        
        // 记录请求开始
        log.Printf("Started %s %s", method, path)
        
        // 执行下一个处理器
        ctx.Next(c)
        
        // 记录请求结束
        latency := time.Since(start)
        status := ctx.Response.StatusCode()
        
        log.Printf("Completed %s %s - %d in %v", 
            method, path, status, latency)
    }
}
```

### 2. 请求ID生成中间件

```go
import (
    "crypto/rand"
    "encoding/hex"
    "context"
)

// RequestID 为每个请求生成唯一ID
func RequestID() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        // 生成请求ID
        requestID := generateRequestID()
        
        // 设置到响应头
        ctx.Response.Header.Set("X-Request-ID", requestID)
        
        // 设置到上下文，供后续处理器使用
        ctx.Set("request_id", requestID)
        
        ctx.Next(c)
    }
}

func generateRequestID() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
```

### 3. 响应时间统计中间件

```go
import (
    "context"
    "strconv"
    "time"
)

// ResponseTime 响应时间统计中间件
func ResponseTime() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        start := time.Now()
        
        ctx.Next(c)
        
        // 计算响应时间
        duration := time.Since(start)
        
        // 设置响应头
        ctx.Response.Header.Set("X-Response-Time", 
            strconv.FormatInt(duration.Nanoseconds()/1000000, 10)+"ms")
        
        // 记录慢请求
        if duration > 1*time.Second {
            log.Printf("Slow request: %s %s took %v", 
                ctx.Method(), ctx.Path(), duration)
        }
    }
}
```

### 4. API版本控制中间件

```go
import (
    "fmt"
    "net/http"
    "strings"
)

type VersionConfig struct {
    // 支持的版本列表
    SupportedVersions []string
    // 默认版本
    DefaultVersion string
    // 版本头名称
    VersionHeader string
}

// APIVersion API版本控制中间件
func APIVersion(config VersionConfig) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        // 从请求头获取版本
        version := string(ctx.Request.Header.Get(config.VersionHeader))
        
        // 从URL路径获取版本（如 /v1/users）
        if version == "" {
            path := string(ctx.Path())
            if strings.HasPrefix(path, "/v") {
                parts := strings.Split(path, "/")
                if len(parts) > 1 {
                    version = parts[1]
                }
            }
        }
        
        // 使用默认版本
        if version == "" {
            version = config.DefaultVersion
        }
        
        // 验证版本是否支持
        if !isVersionSupported(version, config.SupportedVersions) {
            ctx.JSON(http.StatusBadRequest, map[string]interface{}{
                "error": "Unsupported API version",
                "supported_versions": config.SupportedVersions,
            })
            ctx.Abort()
            return
        }
        
        // 设置版本到上下文
        ctx.Set("api_version", version)
        ctx.Response.Header.Set("API-Version", version)
        
        ctx.Next(c)
    }
}

func isVersionSupported(version string, supported []string) bool {
    for _, v := range supported {
        if v == version {
            return true
        }
    }
    return false
}
```

### 5. 请求大小限制中间件

```go
import (
    "fmt"
    "net/http"
)

type SizeLimitConfig struct {
    // 最大请求体大小（字节）
    MaxSize int64
    // 错误消息
    ErrorMessage string
}

// SizeLimit 请求大小限制中间件
func SizeLimit(config SizeLimitConfig) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        // 检查Content-Length头
        contentLength := ctx.Request.Header.ContentLength()
        
        if contentLength > config.MaxSize {
            ctx.JSON(http.StatusRequestEntityTooLarge, map[string]interface{}{
                "error": "Request entity too large",
                "message": config.ErrorMessage,
                "max_size": config.MaxSize,
            })
            ctx.Abort()
            return
        }
        
        ctx.Next(c)
    }
}
```

### 6. IP白名单中间件

```go
import (
    "net"
    "net/http"
)

type IPWhitelistConfig struct {
    // 允许的IP列表
    AllowedIPs []string
    // 允许的IP网段
    AllowedCIDRs []*net.IPNet
    // 错误消息
    ErrorMessage string
}

// IPWhitelist IP白名单中间件
func IPWhitelist(config IPWhitelistConfig) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        clientIP := ctx.ClientIP()
        
        // 检查IP是否在白名单中
        if !isIPAllowed(clientIP, config) {
            ctx.JSON(http.StatusForbidden, map[string]interface{}{
                "error": "Access denied",
                "message": config.ErrorMessage,
            })
            ctx.Abort()
            return
        }
        
        ctx.Next(c)
    }
}

func isIPAllowed(ip string, config IPWhitelistConfig) bool {
    // 检查精确IP匹配
    for _, allowedIP := range config.AllowedIPs {
        if ip == allowedIP {
            return true
        }
    }
    
    // 检查CIDR网段匹配
    clientIP := net.ParseIP(ip)
    if clientIP != nil {
        for _, cidr := range config.AllowedCIDRs {
            if cidr.Contains(clientIP) {
                return true
            }
        }
    }
    
    return false
}
```

### 7. 缓存中间件

```go
import (
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "sync"
    "time"
)

type CacheItem struct {
    Data      []byte
    Headers   map[string]string
    ExpiresAt time.Time
}

type CacheConfig struct {
    // 缓存TTL
    TTL time.Duration
    // 缓存大小限制
    MaxSize int
    // 缓存键生成函数
    KeyFunc func(ctx *app.RequestContext) string
}

// MemoryCache 内存缓存中间件
func MemoryCache(config CacheConfig) app.HandlerFunc {
    cache := make(map[string]*CacheItem)
    mutex := sync.RWMutex{}
    
    return func(c context.Context, ctx *app.RequestContext) {
        // 只缓存GET请求
        if string(ctx.Method()) != "GET" {
            ctx.Next(c)
            return
        }
        
        // 生成缓存键
        key := config.KeyFunc(ctx)
        if key == "" {
            key = generateCacheKey(ctx)
        }
        
        // 检查缓存
        mutex.RLock()
        item, exists := cache[key]
        mutex.RUnlock()
        
        if exists && time.Now().Before(item.ExpiresAt) {
            // 缓存命中，返回缓存数据
            for k, v := range item.Headers {
                ctx.Response.Header.Set(k, v)
            }
            ctx.Response.Header.Set("X-Cache", "HIT")
            ctx.Write(item.Data)
            return
        }
        
        // 缓存未命中，执行请求
        ctx.Next(c)
        
        // 缓存响应（仅缓存200状态码）
        if ctx.Response.StatusCode() == 200 {
            mutex.Lock()
            defer mutex.Unlock()
            
            // 检查缓存大小限制
            if len(cache) >= config.MaxSize {
                // 简单的LRU：删除一个过期项
                for k, v := range cache {
                    if time.Now().After(v.ExpiresAt) {
                        delete(cache, k)
                        break
                    }
                }
            }
            
            // 复制响应头
            headers := make(map[string]string)
            ctx.Response.Header.VisitAll(func(key, value []byte) {
                headers[string(key)] = string(value)
            })
            
            // 存储到缓存
            cache[key] = &CacheItem{
                Data:      ctx.Response.Body(),
                Headers:   headers,
                ExpiresAt: time.Now().Add(config.TTL),
            }
            
            ctx.Response.Header.Set("X-Cache", "MISS")
        }
    }
}

func generateCacheKey(ctx *app.RequestContext) string {
    data := fmt.Sprintf("%s:%s", ctx.Method(), ctx.Request.URI().String())
    hash := md5.Sum([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

## 🔧 高级中间件模式

### 1. 中间件工厂模式

```go
// MiddlewareFactory 中间件工厂
type MiddlewareFactory struct {
    config Config
    logger Logger
}

func NewMiddlewareFactory(config Config, logger Logger) *MiddlewareFactory {
    return &MiddlewareFactory{
        config: config,
        logger: logger,
    }
}

// CreateRateLimit 创建限流中间件
func (f *MiddlewareFactory) CreateRateLimit(opts RateLimitOptions) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        // 使用工厂配置和选项创建限流逻辑
        // ...
    }
}

// CreateAuth 创建认证中间件
func (f *MiddlewareFactory) CreateAuth(opts AuthOptions) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        // 使用工厂配置和选项创建认证逻辑
        // ...
    }
}
```

### 2. 中间件链式调用

```go
// MiddlewareChain 中间件链
type MiddlewareChain struct {
    middlewares []app.HandlerFunc
}

func NewMiddlewareChain() *MiddlewareChain {
    return &MiddlewareChain{
        middlewares: make([]app.HandlerFunc, 0),
    }
}

func (mc *MiddlewareChain) Use(middleware app.HandlerFunc) *MiddlewareChain {
    mc.middlewares = append(mc.middlewares, middleware)
    return mc
}

func (mc *MiddlewareChain) Handler() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        // 执行所有中间件
        for _, middleware := range mc.middlewares {
            middleware(c, ctx)
            // 检查是否被中断
            if ctx.IsAborted() {
                return
            }
        }
        ctx.Next(c)
    }
}

// 使用示例
chain := NewMiddlewareChain().
    Use(LoggerMiddleware()).
    Use(AuthMiddleware()).
    Use(RateLimitMiddleware())

app.Use(chain.Handler())
```

### 3. 条件中间件

```go
// ConditionalMiddleware 条件中间件
func ConditionalMiddleware(condition func(*app.RequestContext) bool, middleware app.HandlerFunc) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        if condition(ctx) {
            middleware(c, ctx)
        } else {
            ctx.Next(c)
        }
    }
}

// 使用示例
app.Use(ConditionalMiddleware(
    func(ctx *app.RequestContext) bool {
        // 只对API路径应用认证
        return strings.HasPrefix(string(ctx.Path()), "/api/")
    },
    AuthMiddleware(),
))
```

### 4. 异步中间件

```go
import (
    "sync"
)

// AsyncMiddleware 异步中间件
func AsyncMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        var wg sync.WaitGroup
        
        // 异步任务1
        wg.Add(1)
        go func() {
            defer wg.Done()
            // 异步记录日志
            asyncLogRequest(ctx)
        }()
        
        // 异步任务2  
        wg.Add(1)
        go func() {
            defer wg.Done()
            // 异步统计
            asyncRecordMetrics(ctx)
        }()
        
        // 继续处理请求
        ctx.Next(c)
        
        // 可选：等待异步任务完成
        // wg.Wait()
    }
}
```

## 📊 中间件测试

### 单元测试示例

```go
package middleware_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/common/test/assert"
    "github.com/cloudwego/hertz/pkg/common/ut"
)

func TestResponseTime(t *testing.T) {
    h := server.Default()
    h.Use(ResponseTime())
    
    h.GET("/test", func(c context.Context, ctx *app.RequestContext) {
        // 模拟处理时间
        time.Sleep(100 * time.Millisecond)
        ctx.String(200, "OK")
    })
    
    // 创建测试请求
    w := ut.PerformRequest(h.Engine, "GET", "/test", nil)
    
    // 验证响应
    assert.DeepEqual(t, 200, w.Code)
    assert.NotNil(t, w.Header().Get("X-Response-Time"))
}

func TestIPWhitelist(t *testing.T) {
    config := IPWhitelistConfig{
        AllowedIPs: []string{"127.0.0.1", "192.168.1.1"},
        ErrorMessage: "Access denied",
    }
    
    h := server.Default()
    h.Use(IPWhitelist(config))
    
    h.GET("/test", func(c context.Context, ctx *app.RequestContext) {
        ctx.String(200, "OK")
    })
    
    // 测试允许的IP
    req := ut.NewRequest("GET", "/test", nil)
    req.Header.Set("X-Forwarded-For", "127.0.0.1")
    w := ut.PerformRequest(h.Engine, req)
    assert.DeepEqual(t, 200, w.Code)
    
    // 测试不允许的IP
    req = ut.NewRequest("GET", "/test", nil)
    req.Header.Set("X-Forwarded-For", "192.168.2.1")
    w = ut.PerformRequest(h.Engine, req)
    assert.DeepEqual(t, 403, w.Code)
}
```

### 基准测试

```go
func BenchmarkResponseTime(b *testing.B) {
    h := server.Default()
    h.Use(ResponseTime())
    
    h.GET("/test", func(c context.Context, ctx *app.RequestContext) {
        ctx.String(200, "OK")
    })
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := ut.PerformRequest(h.Engine, "GET", "/test", nil)
        _ = w
    }
}
```

## 🛠️ 中间件调试

### 调试中间件

```go
// DebugMiddleware 调试中间件
func DebugMiddleware(enabled bool) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        if !enabled {
            ctx.Next(c)
            return
        }
        
        start := time.Now()
        
        // 打印请求信息
        log.Printf("[DEBUG] Request: %s %s", ctx.Method(), ctx.Path())
        log.Printf("[DEBUG] Headers: %v", ctx.Request.Header)
        
        ctx.Next(c)
        
        // 打印响应信息
        duration := time.Since(start)
        log.Printf("[DEBUG] Response: %d in %v", ctx.Response.StatusCode(), duration)
    }
}
```

### 性能监控

```go
// MetricsMiddleware 性能监控中间件
func MetricsMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        start := time.Now()
        path := string(ctx.Path())
        method := string(ctx.Method())
        
        ctx.Next(c)
        
        duration := time.Since(start)
        status := ctx.Response.StatusCode()
        
        // 记录指标
        recordMetrics(method, path, status, duration)
    }
}

func recordMetrics(method, path string, status int, duration time.Duration) {
    // 发送到监控系统（如Prometheus）
    // prometheus.CounterVec.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
    // prometheus.HistogramVec.WithLabelValues(method, path).Observe(duration.Seconds())
}
```

## 📚 最佳实践

### 1. 中间件设计原则

- **单一职责**: 每个中间件只负责一个功能
- **无状态**: 避免在中间件中保存状态
- **性能优先**: 避免阻塞操作，考虑异步处理
- **可配置**: 提供配置选项以适应不同场景

### 2. 错误处理

```go
func SafeMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Middleware panic recovered: %v", r)
                ctx.JSON(500, map[string]interface{}{
                    "error": "Internal server error",
                })
                ctx.Abort()
            }
        }()
        
        ctx.Next(c)
    }
}
```

### 3. 配置验证

```go
func ValidateConfig(config MiddlewareConfig) error {
    if config.Timeout <= 0 {
        return errors.New("timeout must be positive")
    }
    if len(config.AllowedHosts) == 0 {
        return errors.New("allowed hosts cannot be empty")
    }
    return nil
}
```

## 🔗 相关资源

- [中间件概览](./overview.md)
- [内置中间件](./builtin.md)
- [中间件配置](./config.md)
- [性能优化指南](../dev-tools/performance.md)

---

> 💡 **提示**: 编写中间件时要注意性能影响，特别是在高并发场景下。建议进行充分的测试和性能评估。
