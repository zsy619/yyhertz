# InsertFilter 过滤器功能

YYHertz MVC 框架提供了强大的 `InsertFilter` 功能，允许开发者在请求处理的5个关键位置插入自定义过滤器，实现灵活的请求拦截和处理。

## 🎯 功能特点

- **5个执行位置**: 支持在请求处理生命周期的5个关键点插入过滤器
- **自动模式匹配**: 框架自动进行pattern匹配，过滤器无需手动判断路径
- **通配符支持**: 支持 `*` 通配符模式匹配，灵活控制过滤器作用范围
- **线程安全**: 完全线程安全的设计，支持并发操作
- **动态管理**: 支持运行时动态添加和移除过滤器
- **优先级控制**: 按插入顺序执行，确保可预测的执行顺序

## 📚 API 参考

### 过滤器位置常量

```go
const (
    BeforeStatic = 0  // 静态文件处理前
    BeforeRouter = 1  // 路由匹配前
    BeforeExec   = 2  // 控制器执行前
    AfterExec    = 3  // 控制器执行后
    FinishRouter = 4  // 请求处理完成后
)
```

### 核心方法

```go
// 插入过滤器
mvc.InsertFilter(pattern string, position int, filter FilterFunc, params ...bool)

// 移除过滤器
mvc.RemoveFilter(pattern string, position int) bool

// 列出指定位置的过滤器
mvc.ListFilters(position int) []*FilterPattern

// 获取所有过滤器
mvc.GetAllFilters() map[int][]*FilterPattern
```

### FilterFunc 类型

```go
type FilterFunc func(*mvc.Context)
```

## 🚀 快速开始

### 1. 基本用法

```go
package main

import (
    "fmt"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/context"
)

func main() {
    // 认证过滤器 - 只对 /api/* 路径执行，框架自动匹配
    authFilter := func(ctx *context.Context) {
        // 无需检查路径！框架已确保只在 /api/* 时执行
        token := ctx.Header("Authorization")
        if token == "" {
            ctx.JSON(401, map[string]string{"error": "Unauthorized"})
            ctx.Abort()
            return
        }
        // 验证token逻辑...
        ctx.Set("user_id", "authenticated_user")
    }

    // 日志过滤器 - 对所有路径执行
    logFilter := func(ctx *context.Context) {
        // 无需判断路径，框架保证pattern匹配后才执行
        path := string(ctx.Request.Path())
        method := string(ctx.Request.Method())
        fmt.Printf("[LOG] %s %s\n", method, path)
    }

    // 管理员过滤器 - 只对管理员路径执行
    adminFilter := func(ctx *context.Context) {
        // 框架已经匹配了 /admin/* pattern，直接执行业务逻辑
        role := ctx.Header("X-Admin-Role")
        if role != "admin" {
            ctx.JSON(403, map[string]string{"error": "Admin required"})
            ctx.Abort()
            return
        }
    }

    // 插入过滤器 - 框架自动进行pattern匹配
    mvc.InsertFilter("/api/*", mvc.BeforeRouter, authFilter)     // 只对API路径
    mvc.InsertFilter("/*", mvc.BeforeExec, logFilter)           // 对所有路径
    mvc.InsertFilter("/admin/*", mvc.BeforeExec, adminFilter)   // 只对管理员路径

    // 启动应用
    app := mvc.HertzApp
    app.AutoRouters(&YourController{})
    app.Run(":8080")
}
```

### 2. 自动Pattern匹配的优势

**🎯 重要特性：框架自动处理pattern匹配，过滤器函数无需手动判断路径！**

```go
// ❌ 旧的方式：需要在过滤器内部判断路径
func oldStyleFilter(ctx *context.Context) {
    path := string(ctx.Request.Path())
    if strings.HasPrefix(path, "/api/") {  // 手动判断路径
        // 执行过滤逻辑
    }
}

// ✅ YYHertz方式：框架自动匹配，过滤器专注业务逻辑
func newStyleFilter(ctx *context.Context) {
    // 无需判断路径！框架已保证只在匹配时执行
    // 直接执行业务逻辑
    token := ctx.Header("Authorization")
    if token == "" {
        ctx.JSON(401, map[string]string{"error": "Unauthorized"})
        ctx.Abort()
    }
}

// 使用时指定pattern，框架自动匹配
mvc.InsertFilter("/api/*", mvc.BeforeRouter, newStyleFilter)
```

### 3. 模式匹配示例

```go
// 精确匹配 - 只对 /api/users 路径执行
mvc.InsertFilter("/api/users", mvc.BeforeRouter, authFilter)

// 前缀匹配 - 只对 /api/ 开头的路径执行
mvc.InsertFilter("/api/*", mvc.BeforeRouter, authFilter)

// 后缀匹配 - 只对 .json 结尾的路径执行
mvc.InsertFilter("*.json", mvc.AfterExec, jsonResponseFilter)

// 全局匹配 - 对所有路径执行
mvc.InsertFilter("*", mvc.BeforeRouter, globalFilter)

// 中间通配符匹配 - 只对 /api/任意/users 模式执行
mvc.InsertFilter("/api/*/users", mvc.BeforeExec, userFilter)
```

## 📖 详细用法示例

### 1. 认证和权限控制

```go
// JWT认证过滤器 - 框架确保只对匹配的路径执行
func jwtAuthFilter(ctx *context.Context) {
    // 🎯 无需检查路径！框架已经保证只在 /api/* 时执行
    token := ctx.Header("Authorization")
    if token == "" {
        ctx.JSON(401, map[string]interface{}{
            "error": "Missing authorization token",
            "code":  401,
        })
        ctx.Abort()
        return
    }

    // 验证JWT token
    claims, err := validateJWT(token)
    if err != nil {
        ctx.JSON(401, map[string]interface{}{
            "error": "Invalid token", 
            "code":  401,
        })
        ctx.Abort()
        return
    }

    // 将用户信息存储到上下文
    ctx.Set("user_id", claims.UserID)
    ctx.Set("user_role", claims.Role)
}

// 管理员权限过滤器 - 只对管理员API执行
func adminAuthFilter(ctx *context.Context) {
    // 🎯 框架已匹配 /api/admin/* pattern，直接检查权限
    role, exists := ctx.Get("user_role")
    if !exists || role != "admin" {
        ctx.JSON(403, map[string]interface{}{
            "error": "Admin access required",
            "code":  403,
        })
        ctx.Abort()
        return
    }
}

// 使用认证过滤器 - 指定精确pattern
mvc.InsertFilter("/api/*", mvc.BeforeRouter, jwtAuthFilter)         // 只对API路径
mvc.InsertFilter("/api/admin/*", mvc.BeforeExec, adminAuthFilter)   // 只对管理员API
```

### 2. 请求日志和监控

```go
// 请求开始时间记录
func requestStartFilter(ctx *context.Context) {
    ctx.Set("start_time", time.Now())
    
    // 生成请求ID
    requestID := generateRequestID()
    ctx.Set("request_id", requestID)
    ctx.Request.Header("X-Request-ID", requestID)
}

// 请求完成日志
func requestCompleteFilter(ctx *context.Context) {
    startTime, _ := ctx.Get("start_time")
    if startTime != nil {
        duration := time.Since(startTime.(time.Time))
        
        requestID, _ := ctx.Get("request_id")
        path := string(ctx.Request.Path())
        method := string(ctx.Request.Method())
        status := ctx.Writer.Status()
        
        // 记录访问日志
        logEntry := map[string]interface{}{
            "request_id": requestID,
            "method":     method,
            "path":       path,
            "status":     status,
            "duration":   duration.Milliseconds(),
            "timestamp":  time.Now().Unix(),
        }
        
        // 发送到日志系统
        logger.Info("HTTP Request", logEntry)
        
        // 性能监控
        if duration > time.Second {
            logger.Warn("Slow Request", logEntry)
        }
    }
}

mvc.InsertFilter("/*", mvc.BeforeStatic, requestStartFilter)
mvc.InsertFilter("/*", mvc.FinishRouter, requestCompleteFilter)
```

### 3. 限流和防护

```go
// 简单的限流过滤器
func rateLimitFilter(ctx *context.Context) {
    clientIP := ctx.Request.ClientIP()
    
    // 检查IP限流
    if !checkRateLimit(clientIP) {
        ctx.JSON(429, map[string]interface{}{
            "error": "Too many requests",
            "code":  429,
        })
        ctx.Abort()
        return
    }
    
    // 更新请求计数
    updateRequestCount(clientIP)
}

// CORS处理过滤器
func corsFilter(ctx *context.Context) {
    origin := ctx.Header("Origin")
    if origin != "" {
        ctx.SetHeader("Access-Control-Allow-Origin", origin)
        ctx.SetHeader("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type,Authorization")
        ctx.SetHeader("Access-Control-Allow-Credentials", "true")
    }
    
    // 处理预检请求
    if string(ctx.Request.Method()) == "OPTIONS" {
        ctx.JSON(204, nil)
        ctx.Abort()
        return
    }
}

mvc.InsertFilter("/*", mvc.BeforeRouter, corsFilter)
mvc.InsertFilter("/api/*", mvc.BeforeRouter, rateLimitFilter)
```

### 4. 数据验证和转换

```go
// 请求体大小限制
func bodySizeLimitFilter(ctx *context.Context) {
    contentLength := ctx.Request.Header("Content-Length")
    if contentLength != "" {
        size, err := strconv.ParseInt(contentLength, 10, 64)
        if err == nil && size > 10*1024*1024 { // 10MB限制
            ctx.JSON(413, map[string]interface{}{
                "error": "Request body too large",
                "code":  413,
            })
            ctx.Abort()
            return
        }
    }
}

// API版本验证
func apiVersionFilter(ctx *context.Context) {
    version := ctx.Header("X-API-Version")
    if version == "" {
        version = "v1" // 默认版本
    }
    
    supportedVersions := []string{"v1", "v2"}
    supported := false
    for _, v := range supportedVersions {
        if v == version {
            supported = true
            break
        }
    }
    
    if !supported {
        ctx.JSON(400, map[string]interface{}{
            "error": "Unsupported API version",
            "code":  400,
        })
        ctx.Abort()
        return
    }
    
    ctx.Set("api_version", version)
}

mvc.InsertFilter("/api/*", mvc.BeforeRouter, bodySizeLimitFilter)
mvc.InsertFilter("/api/*", mvc.BeforeRouter, apiVersionFilter)
```

### 5. 响应处理和缓存

```go
// 响应压缩过滤器
func compressionFilter(ctx *context.Context) {
    acceptEncoding := ctx.Header("Accept-Encoding")
    if strings.Contains(acceptEncoding, "gzip") {
        ctx.SetHeader("Content-Encoding", "gzip")
        ctx.Set("use_compression", true)
    }
}

// 缓存控制过滤器
func cacheControlFilter(ctx *context.Context) {
    path := string(ctx.Request.Path())
    
    // 静态资源缓存
    if strings.HasPrefix(path, "/static/") {
        ctx.SetHeader("Cache-Control", "max-age=86400") // 1天
    } else if strings.HasPrefix(path, "/api/") {
        ctx.SetHeader("Cache-Control", "no-cache")
    }
}

// 安全头设置
func securityHeadersFilter(ctx *context.Context) {
    ctx.SetHeader("X-Content-Type-Options", "nosniff")
    ctx.SetHeader("X-Frame-Options", "DENY")
    ctx.SetHeader("X-XSS-Protection", "1; mode=block")
    ctx.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
}

mvc.InsertFilter("/*", mvc.AfterExec, compressionFilter)
mvc.InsertFilter("/*", mvc.AfterExec, cacheControlFilter)
mvc.InsertFilter("/*", mvc.FinishRouter, securityHeadersFilter)
```

## 🛠️ 高级用法

### 1. 动态过滤器管理

```go
// 根据配置动态添加过滤器
func setupDynamicFilters() {
    config := loadAppConfig()
    
    // 根据环境启用不同的过滤器
    if config.Environment == "development" {
        // 开发环境：启用详细日志
        mvc.InsertFilter("/*", mvc.BeforeExec, debugLogFilter)
    } else if config.Environment == "production" {
        // 生产环境：启用性能监控
        mvc.InsertFilter("/*", mvc.BeforeExec, performanceMonitorFilter)
    }
    
    // 根据功能开关启用过滤器
    if config.Features.EnableRateLimit {
        mvc.InsertFilter("/api/*", mvc.BeforeRouter, rateLimitFilter)
    }
    
    if config.Features.EnableAuth {
        mvc.InsertFilter("/api/*", mvc.BeforeRouter, authFilter)
    }
}

// 运行时添加和移除过滤器
func addTemporaryFilter(pattern string, duration time.Duration) {
    tempFilter := func(ctx *context.Context) {
        // 临时过滤器逻辑
    }
    
    mvc.InsertFilter(pattern, mvc.BeforeRouter, tempFilter)
    
    // 定时移除
    time.AfterFunc(duration, func() {
        mvc.RemoveFilter(pattern, mvc.BeforeRouter)
    })
}
```

### 2. 过滤器链管理

```go
// 获取和分析过滤器状态
func analyzeFilters() {
    allFilters := mvc.GetAllFilters()
    
    for position, filters := range allFilters {
        positionName := getPositionName(position)
        fmt.Printf("Position %s: %d filters\n", positionName, len(filters))
        
        for i, filter := range filters {
            fmt.Printf("  %d. Pattern: %s, Enabled: %v\n", 
                i+1, filter.Pattern, filter.Enabled)
        }
    }
}

// 过滤器性能统计
func filterPerformanceStats() {
    positions := []int{
        mvc.BeforeStatic, mvc.BeforeRouter, mvc.BeforeExec,
        mvc.AfterExec, mvc.FinishRouter,
    }
    
    for _, position := range positions {
        filters := mvc.ListFilters(position)
        fmt.Printf("Position %d has %d active filters\n", position, len(filters))
    }
}

func getPositionName(position int) string {
    names := map[int]string{
        mvc.BeforeStatic: "BeforeStatic",
        mvc.BeforeRouter: "BeforeRouter",
        mvc.BeforeExec:   "BeforeExec",
        mvc.AfterExec:    "AfterExec",
        mvc.FinishRouter: "FinishRouter",
    }
    return names[position]
}
```

### 3. 条件过滤器

```go
// 基于条件的过滤器包装器
func conditionalFilter(condition func(*context.Context) bool, filter FilterFunc) FilterFunc {
    return func(ctx *context.Context) {
        if condition(ctx) {
            filter(ctx)
        }
    }
}

// 使用示例
func setupConditionalFilters() {
    // 只在工作时间执行的过滤器
    businessHoursCondition := func(ctx *context.Context) bool {
        now := time.Now()
        hour := now.Hour()
        return hour >= 9 && hour <= 17 // 9AM - 5PM
    }
    
    businessHoursFilter := conditionalFilter(businessHoursCondition, func(ctx *context.Context) {
        ctx.SetHeader("X-Business-Hours", "true")
    })
    
    mvc.InsertFilter("/*", mvc.BeforeExec, businessHoursFilter)
    
    // 只对特定用户类型执行的过滤器
    premiumUserCondition := func(ctx *context.Context) bool {
        userType, exists := ctx.Get("user_type")
        return exists && userType == "premium"
    }
    
    premiumFeatureFilter := conditionalFilter(premiumUserCondition, func(ctx *context.Context) {
        ctx.Set("premium_features_enabled", true)
    })
    
    mvc.InsertFilter("/api/*", mvc.BeforeExec, premiumFeatureFilter)
}
```

## ⚠️ 注意事项

### 1. 过滤器执行顺序

过滤器按以下顺序执行：
1. **BeforeStatic** - 在静态文件处理前
2. **BeforeRouter** - 在路由匹配前
3. **BeforeExec** - 在控制器执行前  
4. **AfterExec** - 在控制器执行后
5. **FinishRouter** - 在请求处理完成后

同一位置的过滤器按插入顺序执行。

### 2. 性能考虑

```go
// ✅ 推荐：轻量级过滤器
func efficientFilter(ctx *context.Context) {
    // 快速检查和处理
    if quickCheck(ctx) {
        // 简单操作
    }
}

// ❌ 避免：耗时操作
func heavyFilter(ctx *context.Context) {
    // 避免数据库查询
    // 避免外部API调用
    // 避免复杂计算
}
```

### 3. 错误处理

```go
func safeFilter(ctx *context.Context) {
    defer func() {
        if r := recover(); r != nil {
            // 记录错误但不中断请求
            logger.Error("Filter panic", r)
        }
    }()
    
    // 过滤器逻辑
}
```

### 4. 模式匹配规则

- `*` - 匹配任意路径
- `/api/*` - 匹配以 `/api/` 开头的路径
- `*.json` - 匹配以 `.json` 结尾的路径  
- `/api/*/users` - 匹配 `/api/任意/users` 模式
- `/exact` - 精确匹配 `/exact` 路径

## 🧪 测试

```go
func TestCustomFilter(t *testing.T) {
    // 插入测试过滤器
    var executed bool
    testFilter := func(ctx *context.Context) {
        executed = true
    }
    
    mvc.InsertFilter("/test/*", mvc.BeforeRouter, testFilter)
    
    // 验证过滤器被插入
    filters := mvc.ListFilters(mvc.BeforeRouter)
    if len(filters) != 1 {
        t.Errorf("Expected 1 filter, got %d", len(filters))
    }
    
    // 清理
    mvc.RemoveFilter("/test/*", mvc.BeforeRouter)
}
```

## 📚 更多资源

- [YYHertz MVC 框架文档](../../README.md)
- [中间件系统文档](../middleware/README.md)
- [上下文系统文档](../context/README.md)

---

通过 `InsertFilter` 功能，您可以轻松实现灵活的请求拦截和处理，构建强大的Web应用！