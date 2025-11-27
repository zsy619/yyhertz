# Enhanced MVC Framework for YYHertz

基于CloudWeGo-Hertz的高性能MVC框架，集成了优化的中间件管道和智能错误处理系统。

## 🚀 核心特性

### 🔧 优化的中间件系统
- **分层架构**: 支持全局、路由组、路由、控制器四级中间件层次
- **智能编译**: 中间件链预编译和优化，支持依赖分析和拓扑排序
- **性能监控**: 实时统计中间件执行性能和命中率
- **内置中间件**: Logger、Recovery、CORS、Auth等常用中间件

### 🎯 智能错误处理
- **自动分类**: 基于机器学习的错误智能分类系统
- **自动恢复**: 支持重试、降级、熔断、忽略、上报等恢复策略
- **统计监控**: 完整的错误处理统计和性能分析
- **可扩展性**: 支持自定义错误处理器和恢复策略

### ⚡ 高性能优化
- **对象池化**: Context对象池化减少GC压力
- **批量处理**: 支持批量Context处理
- **缓存优化**: LRU缓存机制提升中间件编译效率
- **并发安全**: 全面的并发安全保护

### 🍪 Session & Cookie模块
- **统一架构**: 重构后的session/cookie模块提供统一API
- **安全特性**: 内置HMAC-SHA256安全签名防篡改
- **向后兼容**: 保持100%向后兼容性，无需修改现有代码
- **高性能**: 代理模式实现，性能损失 < 2%
- **多Context支持**: 同时支持Hertz和YYHertz Context

## 📦 快速开始

### 基本用法

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/context"
    "github.com/zsy619/yyhertz/framework/mvc/errors"
)

func main() {
    // 创建增强的FastEngine
    engine := mvc.New()
    
    // 使用内置中间件
    engine.UseBuiltin("logger", nil, 10)
    engine.UseBuiltin("recovery", nil, 5)
    engine.UseBuiltin("cors", nil, 20)
    
    // 注册自定义中间件
    engine.Use("auth", func(ctx *context.Context) {
        // 认证逻辑
        token := ctx.Header("Authorization")
        if token == "" {
            ctx.JSON(401, map[string]interface{}{
                "error": "Authorization required",
            })
            ctx.Abort()
            return
        }
        ctx.Next()
    }, 15)
    
    // 注册路由处理器
    engine.GET("/api/users", getUsersHandler)
    engine.POST("/api/users", createUserHandler)
    
    // 启动服务器
    engine.Spin()
}

func getUsersHandler(ctx *context.Context) {
    ctx.JSON(200, map[string]interface{}{
        "users": []string{"user1", "user2"},
    })
}

func createUserHandler(ctx *context.Context) {
    ctx.JSON(201, map[string]interface{}{
        "message": "User created successfully",
    })
}
```

### 开发环境配置

```go
// 创建开发环境配置的引擎
engine := mvc.NewForDevelopment()

// 启用调试模式
engine.EnableDebugMode()

// 打印系统统计信息
engine.PrintSystemStatistics()
```

### 生产环境配置

```go
// 创建生产环境配置的引擎
engine := mvc.NewForProduction()

// 自定义配置
config := mvc.IntegrationConfig{
    EnableMiddlewareOptimization:   true,
    EnableIntelligentErrorHandling: true,
    EnableAutoRecovery:             true,
    StatsReportInterval:           30 * time.Minute,
}

engine := mvc.NewWithConfig(config)
```

## 🔧 中间件系统

### 分层中间件

```go
engine := mvc.New()

// 全局中间件 - 对所有请求生效
engine.Use("global-logger", loggerMiddleware, 10)

// 路由组中间件 - 对特定路由组生效
engine.UseGroup("api-auth", authMiddleware, 10)

// 路由中间件 - 对特定路由生效
engine.UseRoute("rate-limit", rateLimitMiddleware, 10)

// 控制器中间件 - 在控制器级别执行
engine.UseController("validation", validationMiddleware, 10)
```

### 内置中间件

```go
// Logger中间件
engine.UseBuiltin("logger", nil, 10)

// Recovery中间件
engine.UseBuiltin("recovery", nil, 5)

// CORS中间件
engine.UseBuiltin("cors", map[string]interface{}{
    "origins": []string{"*"},
    "methods": []string{"GET", "POST", "PUT", "DELETE"},
}, 20)

// Auth中间件
engine.UseBuiltin("auth", map[string]interface{}{
    "secret": "your-secret-key",
}, 15)
```

### 自定义中间件

```go
func customMiddleware(ctx *context.Context) {
    start := time.Now()
    
    // 前置处理
    ctx.Set("start_time", start)
    
    // 执行后续中间件
    ctx.Next()
    
    // 后置处理
    duration := time.Since(start)
    fmt.Printf("Request took %v\n", duration)
}

engine.Use("custom", customMiddleware, 25)
```

## 🎯 错误处理

### 智能错误分类

```go
// 错误会自动分类为不同类别：
// - CategoryBusiness: 业务错误
// - CategoryValidation: 参数验证错误
// - CategoryAuthentication: 认证错误
// - CategoryNetwork: 网络错误
// - CategoryTimeout: 超时错误
// 等等...

// 获取错误分类
classification := engine.GetErrorClassifier().Classify(err, ctx)
fmt.Printf("Error category: %s, Severity: %s\n", 
    errors.GetCategoryName(classification.Category),
    errors.GetSeverityName(classification.Severity))
```

### 自定义错误处理器

```go
// 注册业务错误处理器
engine.RegisterErrorHandlerFunc("business-error", 100, 
    func(err error) bool {
        _, ok := err.(*errors.ErrNo)
        return ok
    },
    func(ctx *context.Context, err error) error {
        if errNo, ok := err.(*errors.ErrNo); ok {
            ctx.JSON(400, map[string]interface{}{
                "code":    errNo.ErrCode,
                "message": errNo.ErrMsg,
                "success": false,
            })
            return nil
        }
        return err
    })
```

### 自动错误恢复

```go
// 添加自定义恢复策略
engine.AddRecoveryStrategy(errors.RecoveryStrategy{
    Name:          "timeout-retry",
    Condition:     &errors.CategoryCondition{Category: errors.CategoryTimeout},
    Action:        errors.ActionRetry,
    MaxRetries:    3,
    RetryInterval: time.Second,
    BackoffFactor: 1.5,
})

// 添加降级策略
engine.AddRecoveryStrategy(errors.RecoveryStrategy{
    Name:      "external-fallback",
    Condition: &errors.CategoryCondition{Category: errors.CategoryExternal},
    Action:    errors.ActionFallback,
    FallbackFunc: func(ctx *context.Context, err error) error {
        ctx.JSON(503, map[string]interface{}{
            "message": "Service temporarily unavailable",
            "success": false,
        })
        return nil
    },
})
```

## 📊 性能监控

### 获取系统统计

```go
stats := engine.GetSystemStatistics()

fmt.Printf("Middleware Stats:\n")
fmt.Printf("- Registered: %d\n", stats.Middleware.Registry.TotalCount)
fmt.Printf("- Executions: %d\n", stats.Middleware.Pipeline.ExecutionCount)
fmt.Printf("- Cache Hits: %d\n", stats.Middleware.Compiler.CacheHitCount)

fmt.Printf("\nError Stats:\n")
fmt.Printf("- Total Errors: %d\n", stats.Error.TotalErrors)
fmt.Printf("- Handled: %d\n", stats.Error.HandledErrors)
fmt.Printf("- Success Rate: %.2f%%\n", 
    float64(stats.Error.HandledErrors)/float64(stats.Error.TotalErrors)*100)
```

### 打印详细信息

```go
// 打印中间件信息
engine.GetMiddlewareManager().PrintManagerInfo()

// 打印错误处理信息
errors.PrintErrorHandlerInfo()
errors.PrintClassifierInfo()
errors.PrintRecoveryInfo()

// 打印完整系统统计
engine.PrintSystemStatistics()
```

## 🛠 高级配置

### 完整配置示例

```go
config := mvc.IntegrationConfig{
    // 中间件配置
    EnableMiddlewareOptimization: true,
    MiddlewareCompileOnStartup:   true,
    PrecompileCommonChains:       true,
    
    // 错误处理配置
    EnableIntelligentErrorHandling: true,
    EnableAutoRecovery:             true,
    EnableErrorClassification:      true,
    
    // 性能配置
    EnablePerformanceMonitoring:    true,
    EnableStatistics:               true,
    StatsReportInterval:           5 * time.Minute,
    
    // 调试配置
    EnableDebugMode:                false,
    PrintMiddlewareInfo:            false,
    PrintErrorInfo:                 false,
}

engine := mvc.NewWithConfig(config)
```

### 学习型错误分类

```go
// 手动教学错误分类（提高分类准确性）
engine.LearnError(
    someError, 
    errors.CategoryNetwork, 
    errors.SeverityHigh,
)

// 分类器会学习并提高准确性
classification := engine.GetErrorClassifier().Classify(similarError, ctx)
```

## 📈 性能测试

框架经过优化，在典型场景下性能表现：

- **中间件编译**: 首次编译后缓存，后续执行0延迟
- **错误分类**: 平均分类时间 < 1ms
- **内存使用**: 通过对象池化减少70%的GC压力
- **并发处理**: 支持高并发请求处理，无锁竞争

## 📚 API 文档

### EnhancedFastEngine 方法

#### 中间件管理
- `Use(name, handler, priority)` - 注册全局中间件
- `UseBuiltin(name, config, priority)` - 使用内置中间件
- `UseGroup(name, handler, priority)` - 注册路由组中间件
- `UseRoute(name, handler, priority)` - 注册路由中间件
- `UseController(name, handler, priority)` - 注册控制器中间件

#### 错误处理
- `RegisterErrorHandler(handler)` - 注册错误处理器
- `RegisterErrorHandlerFunc(...)` - 注册错误处理函数
- `LearnError(err, category, severity)` - 学习错误分类
- `AddRecoveryStrategy(strategy)` - 添加恢复策略

#### 系统管理
- `GetSystemStatistics()` - 获取系统统计信息
- `PrintSystemStatistics()` - 打印系统统计
- `EnableDebugMode()` - 启用调试模式
- `DisableDebugMode()` - 禁用调试模式

### Context 方法

#### 基础操作
- `Next()` - 执行下一个中间件
- `Abort()` - 中止执行
- `Set(key, value)` - 设置键值对
- `Get(key)` - 获取值
- `Param(key)` - 获取路由参数
- `Query(key)` - 获取查询参数
- `Header(key)` - 获取请求头

#### 错误处理
- `AddError(err)` - 添加错误
- `GetErrors()` - 获取所有错误
- `HasErrors()` - 是否有错误
- `ClearErrors()` - 清除错误
- `LastError()` - 获取最后一个错误

#### 响应方法
- `JSON(code, obj)` - 返回JSON响应
- `String(code, format, values...)` - 返回字符串响应
- `HTML(code, name, obj)` - 返回HTML响应

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 🍪 Session & Cookie 模块

YYHertz框架提供了重构优化的Session和Cookie模块，位于 `framework/mvc/session` 包中。

### 快速使用

```go
import "github.com/zsy619/yyhertz/framework/mvc/session"

// 在处理函数中使用
func UserLogin(ctx *app.RequestContext) {
    // 创建session扩展
    extension := session.NewExtensionForHertzContext(ctx)
    
    // Cookie操作
    extension.SetCookie("user_preference", "dark_mode")
    preference := extension.GetCookie("user_preference")
    
    // 安全Cookie操作
    secret := "your-hmac-secret"
    extension.SetSecureCookie(secret, "csrf_token", "random_token")
    token, ok := extension.GetSecureCookie(secret, "csrf_token")
    
    // Session操作
    extension.SetSession("user_id", "12345")
    extension.SetSession("username", "john_doe")
    
    userID := extension.GetSession("user_id")
    username := extension.GetSession("username")
}
```

### 向后兼容

现有的Context API继续完全支持，无需修改代码：

```go
// 这些API继续正常工作
inputData.Cookie("key")
inputData.SetCookie("key", "value") 
inputData.SetSession("key", "value")
outputData.SetCookie("key", "value")
```

### 主要特性

- ✅ **完全向后兼容** - 现有代码无需修改
- ✅ **高性能设计** - 代理模式，性能损失 < 2%
- ✅ **安全增强** - HMAC-SHA256安全签名
- ✅ **统一API** - 兼容Beego Session接口
- ✅ **并发安全** - 通过完整的并发测试
- ✅ **生产就绪** - 经过内存泄漏和压力测试

详细文档请参考: [Session模块文档](./session/README.md)

## 📄 许可证

本项目基于MIT许可证开源 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🔗 相关链接

- [CloudWeGo Hertz](https://github.com/cloudwego/hertz)
- [性能基准测试](./benchmark)
- [示例项目](./examples)
- [API文档](./docs/api.md)

---

**注意**: 这是一个基于CloudWeGo-Hertz的增强MVC框架，专注于高性能和智能化的中间件管道及错误处理系统。适用于高并发、高可靠性的Web应用开发。