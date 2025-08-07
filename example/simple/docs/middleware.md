# ⚡ 中间件系统

中间件是在请求处理过程中执行的函数，可以用于认证、日志记录、CORS处理、限流等功能。

## 中间件概念

### 中间件执行流程

```
请求 → 中间件1 → 中间件2 → 控制器 → 中间件2 → 中间件1 → 响应
```

中间件采用洋葱模型，先进后出的执行顺序。

### 中间件接口

```go
type MiddlewareFunc func(c *app.RequestContext)

// 或者返回中间件函数的工厂函数
type MiddlewareFactory func(...interface{}) MiddlewareFunc
```

## 内置中间件

### 恢复中间件

```go
import "github.com/zsy619/yyhertz/framework/mvc/middleware"

func main() {
    app := mvc.HertzApp
    
    // 恢复中间件 - 捕获panic并恢复
    app.Use(middleware.RecoveryMiddleware())
    
    app.Run(":8080")
}
```

### 日志中间件

```go
func main() {
    app := mvc.HertzApp
    
    // 日志中间件 - 记录请求日志
    app.Use(middleware.LoggerMiddleware())
    
    // 自定义日志格式
    app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
        Format: "[${time_rfc3339}] ${status} ${method} ${path} ${latency}\n",
        Output: os.Stdout,
    }))
    
    app.Run(":8080")
}
```

### CORS中间件

```go
func main() {
    app := mvc.HertzApp
    
    // CORS中间件 - 处理跨域请求
    app.Use(middleware.CORSMiddleware())
    
    // 自定义CORS配置
    app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins:     []string{"http://localhost:3000", "https://yourdomain.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
    
    app.Run(":8080")
}
```

### 限流中间件

```go
func main() {
    app := mvc.HertzApp
    
    // 限流中间件 - 100次/分钟
    app.Use(middleware.RateLimitMiddleware(100, time.Minute))
    
    // 按IP限流
    app.Use(middleware.RateLimitByIP(50, time.Minute))
    
    // 按用户限流
    app.Use(middleware.RateLimitByUser(1000, time.Hour))
    
    app.Run(":8080")
}
```

## 自定义中间件

### 基础中间件

```go
func AuthMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        token := string(ctx.GetHeader("Authorization"))
        
        if token == "" {
            ctx.JSON(401, map[string]string{
                "error": "需要认证token",
            })
            ctx.Abort()
            return
        }
        
        // 验证token
        user, err := validateToken(token)
        if err != nil {
            ctx.JSON(401, map[string]string{
                "error": "无效的token",
            })
            ctx.Abort()
            return
        }
        
        // 将用户信息存储到上下文中
        ctx.Set("user", user)
        ctx.Next(c)
    }
}
```

### 带配置的中间件

```go
type AuthConfig struct {
    TokenHeader    string
    SkipPaths      []string
    UnauthorizedHandler func(*app.RequestContext)
}

func AuthWithConfig(config AuthConfig) app.HandlerFunc {
    // 设置默认值
    if config.TokenHeader == "" {
        config.TokenHeader = "Authorization"
    }
    
    if config.UnauthorizedHandler == nil {
        config.UnauthorizedHandler = func(ctx *app.RequestContext) {
            ctx.JSON(401, map[string]string{"error": "Unauthorized"})
        }
    }
    
    return func(c context.Context, ctx *app.RequestContext) {
        path := string(ctx.Path())
        
        // 跳过指定路径
        for _, skipPath := range config.SkipPaths {
            if path == skipPath {
                ctx.Next(c)
                return
            }
        }
        
        token := string(ctx.GetHeader(config.TokenHeader))
        if token == "" {
            config.UnauthorizedHandler(ctx)
            ctx.Abort()
            return
        }
        
        // 验证逻辑...
        ctx.Next(c)
    }
}

// 使用
func main() {
    app := mvc.HertzApp
    
    app.Use(AuthWithConfig(AuthConfig{
        TokenHeader: "X-API-Token",
        SkipPaths: []string{"/login", "/register", "/health"},
        UnauthorizedHandler: func(ctx *app.RequestContext) {
            ctx.JSON(401, map[string]interface{}{
                "code": 401,
                "message": "请先登录",
            })
        },
    }))
    
    app.Run(":8080")
}
```

## 中间件应用场景

### 认证中间件

```go
func JWTMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        tokenString := extractToken(ctx)
        if tokenString == "" {
            respondUnauthorized(ctx, "Token缺失")
            return
        }
        
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
            }
            return jwtSecret, nil
        })
        
        if err != nil || !token.Valid {
            respondUnauthorized(ctx, "无效的token")
            return
        }
        
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            ctx.Set("user_id", claims["user_id"])
            ctx.Set("username", claims["username"])
            ctx.Set("role", claims["role"])
        }
        
        ctx.Next(c)
    }
}

func extractToken(ctx *app.RequestContext) string {
    // 从Header获取
    bearer := string(ctx.GetHeader("Authorization"))
    if len(bearer) > 7 && bearer[:7] == "Bearer " {
        return bearer[7:]
    }
    
    // 从查询参数获取
    return string(ctx.QueryArgs().Peek("token"))
}
```

### 权限中间件

```go
func RequireRole(roles ...string) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        userRole, exists := ctx.Get("role")
        if !exists {
            ctx.JSON(403, map[string]string{"error": "未找到用户角色"})
            ctx.Abort()
            return
        }
        
        roleStr := userRole.(string)
        for _, role := range roles {
            if roleStr == role {
                ctx.Next(c)
                return
            }
        }
        
        ctx.JSON(403, map[string]string{"error": "权限不足"})
        ctx.Abort()
    }
}

// 使用方式
nsAdmin := mvc.NewNamespace("/admin",
    mvc.NSBefore(JWTMiddleware()),
    mvc.NSBefore(RequireRole("admin", "super_admin")),
    
    mvc.NSAutoRouter(&controllers.AdminController{}),
)
```

### 请求验证中间件

```go
func ValidateJSON() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        if string(ctx.Method()) == "POST" || string(ctx.Method()) == "PUT" {
            contentType := string(ctx.GetHeader("Content-Type"))
            if !strings.Contains(contentType, "application/json") {
                ctx.JSON(400, map[string]string{
                    "error": "Content-Type必须为application/json",
                })
                ctx.Abort()
                return
            }
            
            // 验证JSON格式
            var jsonData interface{}
            if err := ctx.BindJSON(&jsonData); err != nil {
                ctx.JSON(400, map[string]string{
                    "error": "无效的JSON格式",
                })
                ctx.Abort()
                return
            }
        }
        
        ctx.Next(c)
    }
}
```

### 缓存中间件

```go
func CacheMiddleware(duration time.Duration) app.HandlerFunc {
    cache := make(map[string]CacheItem)
    mutex := sync.RWMutex{}
    
    return func(c context.Context, ctx *app.RequestContext) {
        // 只缓存GET请求
        if string(ctx.Method()) != "GET" {
            ctx.Next(c)
            return
        }
        
        key := generateCacheKey(ctx)
        
        mutex.RLock()
        item, exists := cache[key]
        mutex.RUnlock()
        
        if exists && time.Now().Before(item.ExpireAt) {
            ctx.Data(item.StatusCode, item.ContentType, item.Body)
            return
        }
        
        // 创建响应写入器来捕获响应
        writer := &CacheWriter{
            ResponseWriter: ctx.Response.BodyWriter(),
            StatusCode:     200,
        }
        
        ctx.Next(c)
        
        // 缓存响应
        if writer.StatusCode == 200 {
            mutex.Lock()
            cache[key] = CacheItem{
                Body:        writer.Body.Bytes(),
                ContentType: string(ctx.Response.Header.ContentType()),
                StatusCode:  writer.StatusCode,
                ExpireAt:    time.Now().Add(duration),
            }
            mutex.Unlock()
        }
    }
}

type CacheItem struct {
    Body        []byte
    ContentType string
    StatusCode  int
    ExpireAt    time.Time
}
```

## 中间件最佳实践

### 1. 中间件顺序

```go
func main() {
    app := mvc.HertzApp
    
    // 推荐的中间件顺序
    app.Use(middleware.RecoveryMiddleware())        // 1. 异常恢复
    app.Use(middleware.LoggerMiddleware())          // 2. 日志记录
    app.Use(middleware.CORSMiddleware())            // 3. CORS处理
    app.Use(middleware.RateLimitMiddleware(100, time.Minute)) // 4. 限流
    app.Use(SecurityHeadersMiddleware())            // 5. 安全头
    app.Use(AuthMiddleware())                       // 6. 认证
    app.Use(PermissionMiddleware())                 // 7. 权限检查
    
    app.Run(":8080")
}
```

### 2. 错误处理

```go
func SafeMiddleware(next app.HandlerFunc) app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("中间件异常: %v", err)
                ctx.JSON(500, map[string]string{
                    "error": "服务器内部错误",
                })
                ctx.Abort()
            }
        }()
        
        next(c, ctx)
    }
}
```

### 3. 性能优化

```go
func OptimizedMiddleware() app.HandlerFunc {
    // 预编译正则表达式
    skipPaths := regexp.MustCompile(`^/(health|metrics|favicon\.ico)$`)
    
    return func(c context.Context, ctx *app.RequestContext) {
        path := string(ctx.Path())
        
        // 快速跳过静态资源
        if skipPaths.MatchString(path) {
            ctx.Next(c)
            return
        }
        
        // 其他逻辑...
        ctx.Next(c)
    }
}
```

### 4. 可配置中间件

```go
type MiddlewareConfig struct {
    Skipper    func(*app.RequestContext) bool
    BeforeFunc func(*app.RequestContext) error
    AfterFunc  func(*app.RequestContext) error
}

func ConfigurableMiddleware(config MiddlewareConfig) app.HandlerFunc {
    if config.Skipper == nil {
        config.Skipper = func(*app.RequestContext) bool { return false }
    }
    
    return func(c context.Context, ctx *app.RequestContext) {
        if config.Skipper(ctx) {
            ctx.Next(c)
            return
        }
        
        if config.BeforeFunc != nil {
            if err := config.BeforeFunc(ctx); err != nil {
                ctx.JSON(400, map[string]string{"error": err.Error()})
                ctx.Abort()
                return
            }
        }
        
        ctx.Next(c)
        
        if config.AfterFunc != nil {
            config.AfterFunc(ctx)
        }
    }
}
```

## 中间件调试

### 中间件执行追踪

```go
func TracingMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        start := time.Now()
        path := string(ctx.Path())
        method := string(ctx.Method())
        
        log.Printf("[TRACE] 开始执行: %s %s", method, path)
        
        ctx.Next(c)
        
        duration := time.Since(start)
        status := ctx.Response.StatusCode()
        
        log.Printf("[TRACE] 完成执行: %s %s %d %v", 
            method, path, status, duration)
    }
}
```

### 中间件性能监控

```go
func PerformanceMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        start := time.Now()
        
        ctx.Next(c)
        
        duration := time.Since(start)
        
        // 记录慢请求
        if duration > 500*time.Millisecond {
            log.Printf("[SLOW] %s %s took %v", 
                ctx.Method(), ctx.Path(), duration)
        }
        
        // 发送到监控系统
        metrics.RecordRequestDuration(
            string(ctx.Method()),
            string(ctx.Path()),
            duration,
        )
    }
}
```

---

中间件是构建强大Web应用的重要工具，合理使用中间件可以让代码更加模块化和可维护！