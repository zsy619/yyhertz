# 🚀 YYHertz 错误处理快速开始

本指南将在5分钟内帮你掌握YYHertz错误处理系统的核心使用方法。

## 📋 准备工作

### 环境要求
- **Go版本**: Go 1.19+
- **YYHertz框架**: v1.4.0+
- **错误处理系统**: v2.0.0+

### 导入错误处理包

```go
import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/errors"
)
```

## ⚡ 一分钟快速启用

### 1. 最简单的设置方式

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/errors"
)

func main() {
    app := mvc.HertzApp
    
    // 🚀 一行代码启用错误处理
    errors.QuickSetup("development")
    
    app.Run()
}
```

这一行代码会自动：
- ✅ 注册默认的错误处理器（401, 403, 404, 500等）
- ✅ 启用错误分类和自动恢复
- ✅ 配置开发环境的详细错误信息
- ✅ 提供美观的HTML错误页面

### 2. 验证错误处理

启动应用后，访问一个不存在的路径：
```bash
curl http://localhost:8080/nonexistent
```

你会看到一个美观的404错误页面，包含：
- 📱 响应式设计
- 🎨 现代化UI界面
- 🔍 错误详情展示
- 🏠 返回首页链接

## 🎯 三分钟自定义配置

### 1. 环境相关配置

```go
func main() {
    app := mvc.HertzApp
    
    // 根据环境选择配置
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    switch env {
    case "production":
        // 生产环境：隐藏错误详情，启用监控
        errors.QuickSetup("production")
        errors.EnableMetrics(true)
        errors.Debug(false)
    case "staging":
        // 预发布环境：详细日志，启用恢复
        errors.QuickSetup("staging")
        errors.EnableRecovery(true)
        errors.Debug(true)
    default:
        // 开发环境：最详细的调试信息
        errors.QuickSetup("development")
        errors.Debug(true)
    }
    
    app.Run()
}
```

### 2. 注册自定义错误处理器

```go
func main() {
    app := mvc.HertzApp
    
    // 基础设置
    errors.QuickSetup("development")
    
    // 🎯 自定义404处理器
    errors.Register(404, errors.JSON(404, map[string]any{
        "code":    404,
        "message": "抱歉，您访问的资源不存在",
        "tips":    "请检查URL是否正确",
        "success": false,
    }))
    
    // 🛡️ 自定义401处理器
    errors.RegisterFunc(401, "custom-auth", 100,
        func(statusCode int, err error) bool {
            return statusCode == 401
        },
        func(ctx *errors.Context, statusCode int, err error) error {
            ctx.JSON(401, map[string]any{
                "code":    401,
                "message": "请先登录",
                "redirect": "/login",
                "success": false,
            })
            return nil
        },
    )
    
    app.Run()
}
```

### 3. 业务控制器中使用

```go
type UserController struct {
    core.BaseController
}

// 获取用户信息
func (c *UserController) GetUser() {
    userID := c.GetInt("id", 0)
    
    if userID == 0 {
        // 🚫 直接处理400错误
        errors.Handle(c.Ctx, 400, fmt.Errorf("用户ID不能为空"))
        return
    }
    
    user, err := getUserByID(userID)
    if err != nil {
        if errors.IsNotFound(err) {
            // 🔍 处理404错误
            errors.Handle(c.Ctx, 404, err)
            return
        }
        
        // ⚠️ 处理500错误
        errors.Handle(c.Ctx, 500, err)
        return
    }
    
    c.JSONSuccess("获取成功", user)
}

// 模拟业务函数
func getUserByID(id int) (map[string]any, error) {
    if id == 999 {
        return nil, fmt.Errorf("用户不存在")
    }
    return map[string]any{
        "id":   id,
        "name": "测试用户",
    }, nil
}
```

## 🔧 五分钟高级配置

### 1. 错误分类和自动恢复

```go
func setupAdvancedErrorHandling() {
    // 📊 启用错误分类
    classifier := errors.GetGlobalClassifier()
    
    // 🎯 自定义分类规则
    classifier.AddRule(func(err error) *errors.ErrorClassification {
        if strings.Contains(err.Error(), "database") {
            return &errors.ErrorClassification{
                Category: errors.CategoryDatabase,
                Severity: errors.SeverityHigh,
                Recoverable: true,
                RetryAfter: time.Second * 5,
            }
        }
        return nil
    })
    
    // 🔄 配置自动恢复
    recovery := errors.GetGlobalRecovery()
    recovery.SetRetryConfig(&errors.RecoveryConfig{
        MaxRetries:    3,
        RetryInterval: time.Second * 2,
        RetryableErrors: []int{500, 502, 503, 504},
        RetryableCategories: []errors.ErrorCategory{
            errors.CategoryNetwork,
            errors.CategoryTimeout,
            errors.CategoryDatabase,
        },
    })
    
    // 启用自动恢复
    errors.EnableRecovery(true)
}
```

### 2. 错误监控和指标

```go
func setupErrorMonitoring() {
    // 📈 启用指标收集
    errors.EnableMetrics(true)
    
    // 获取错误注册器
    registry := errors.GetGlobalErrorRegistry()
    
    // 📊 定期打印统计信息（可选）
    go func() {
        ticker := time.NewTicker(time.Minute * 5)
        for range ticker.C {
            stats := registry.GetStats()
            if metrics, ok := stats["metrics"].(map[string]any); ok {
                fmt.Printf("错误处理统计 - 总请求: %v, 已处理: %v, 未处理: %v\n", 
                    metrics["total_requests"], 
                    metrics["handled_errors"], 
                    metrics["unhandled_errors"])
            }
        }
    }()
}
```

### 3. 中间件集成

```go
func setupErrorMiddleware() {
    app := mvc.HertzApp
    
    // 📡 错误处理中间件
    app.Use(func(ctx *app.RequestContext) {
        defer func() {
            if r := recover(); r != nil {
                err := fmt.Errorf("panic recovered: %v", r)
                
                // 转换上下文
                mvcCtx := errors.NewContextFromRequest(ctx)
                if mvcCtx != nil {
                    errors.Handle(mvcCtx, 500, err)
                }
            }
        }()
        
        ctx.Next()
    })
}
```

## ✅ 完整示例

将所有配置整合到一个完整的应用中：

```go
package main

import (
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/errors"
    "github.com/zsy619/yyhertz/framework/mvc/core"
)

func main() {
    app := mvc.HertzApp
    
    // 🚀 基础设置
    setupBasicErrorHandling()
    
    // 🔧 高级配置
    setupAdvancedErrorHandling()
    
    // 📊 监控配置
    setupErrorMonitoring()
    
    // 📡 中间件集成
    setupErrorMiddleware()
    
    // 🎯 注册控制器
    app.NSNamespace("/api",
        app.NSRouter("/user", &UserController{}),
    )
    
    fmt.Println("🚀 YYHertz应用启动，错误处理系统已就绪")
    errors.PrintStats()
    
    app.Run()
}

func setupBasicErrorHandling() {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    // 基础设置
    errors.QuickSetup(env)
    
    // 自定义处理器
    errors.Register(404, errors.JSON(404, map[string]any{
        "code":    404,
        "message": "资源不存在",
        "success": false,
    }))
}

func setupAdvancedErrorHandling() {
    // 错误分类配置
    classifier := errors.GetGlobalClassifier()
    classifier.AddRule(func(err error) *errors.ErrorClassification {
        errMsg := strings.ToLower(err.Error())
        switch {
        case strings.Contains(errMsg, "timeout"):
            return &errors.ErrorClassification{
                Category: errors.CategoryTimeout,
                Severity: errors.SeverityMedium,
                Recoverable: true,
                RetryAfter: time.Second * 3,
            }
        case strings.Contains(errMsg, "permission"):
            return &errors.ErrorClassification{
                Category: errors.CategoryAuthorization,
                Severity: errors.SeverityLow,
                Recoverable: false,
            }
        default:
            return nil
        }
    })
    
    // 自动恢复配置
    errors.EnableRecovery(true)
}

func setupErrorMonitoring() {
    errors.EnableMetrics(true)
}

func setupErrorMiddleware() {
    // 中间件配置已在main中处理
}

type UserController struct {
    core.BaseController
}

func (c *UserController) GetUser() {
    id := c.GetInt("id", 0)
    
    if id == 0 {
        errors.Handle(c.Ctx, 400, fmt.Errorf("用户ID必须指定"))
        return
    }
    
    c.JSONSuccess("用户获取成功", map[string]any{
        "id":   id,
        "name": "测试用户",
    })
}
```

## 🎓 下一步学习

现在你已经掌握了错误处理系统的基础用法！建议继续学习：

1. **[默认处理器详解](default-handlers.md)** - 了解内置处理器的完整功能
2. **[自定义处理器](custom-handlers.md)** - 开发适合业务的错误处理逻辑
3. **[错误页面定制](error-pages.md)** - 创建个性化的错误页面
4. **[错误监控](monitoring.md)** - 建立完善的监控和告警体系

## 💡 常见问题

### Q: QuickSetup做了什么？
A: 自动注册了401、403、404、500等默认处理器，启用了错误分类和基础监控，配置了环境相关的错误显示详细程度。

### Q: 如何在生产环境隐藏错误详情？
A: 使用`errors.QuickSetup("production")`或手动设置`errors.Debug(false)`。

### Q: 可以同时注册多个404处理器吗？
A: 可以，系统会按优先级顺序尝试处理器，直到找到能处理的为止。

### Q: 错误处理会影响性能吗？
A: 影响很小。系统使用了缓存和池化技术，正常情况下性能开销可以忽略。

---

> 💡 **提示**: 建议先在开发环境测试各种错误情况，确保错误处理器按预期工作后再部署到生产环境。