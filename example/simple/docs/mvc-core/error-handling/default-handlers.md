# 🎯 YYHertz 默认错误处理器详解

YYHertz框架内置了一套完整的默认错误处理器，为常见HTTP错误状态码提供了开箱即用的处理能力。

## 📋 默认处理器概览

### 🔄 自动注册的处理器

当你调用`errors.QuickSetup()`时，系统会自动注册以下默认处理器：

| 状态码 | 处理器名称 | 功能描述 | 优先级 |
|--------|-----------|----------|--------|
| **401** | `default-401` | 未授权访问处理 | 1000 |
| **403** | `default-403` | 禁止访问处理 | 1000 |
| **404** | `default-404` | 资源未找到处理 | 1000 |
| **500** | `default-500` | 服务器内部错误处理 | 1000 |

### 🎨 DefaultErrorController

除了基础处理器外，系统还提供了`DefaultErrorController`，它能够：
- 🖥️ 生成美观的HTML错误页面
- 📱 支持响应式设计
- 🌍 多语言支持（中文/英文）
- 🎯 智能内容协商（HTML/JSON/XML）

## 🔧 默认处理器详解

### 1. 401 Unauthorized 处理器

```go
// 注册源码（精简版）
registry.defaultHandlers[401] = &FuncErrorHandler{
    name:     "default-401",
    priority: 1000,
    canHandle: func(err error) bool { return true },
    handleFunc: func(ctx *Context, statusCode int, err error) error {
        ctx.JSON(401, map[string]any{
            "code":    401,
            "message": "Unauthorized",
            "success": false,
        })
        return nil
    },
}
```

#### 特性说明
- **触发场景**: 用户未登录或Token无效
- **默认响应**: JSON格式的标准错误信息
- **可重写**: 支持注册自定义401处理器覆盖

#### 使用示例
```go
// 在控制器中触发401错误
func (c *UserController) GetProfile() {
    token := c.GetHeader("Authorization")
    if token == "" {
        errors.Handle(c.Ctx, 401, fmt.Errorf("缺少认证token"))
        return
    }
    
    // 验证token逻辑...
}
```

### 2. 403 Forbidden 处理器

```go
// 默认403处理逻辑
registry.defaultHandlers[403] = &FuncErrorHandler{
    name:     "default-403",
    priority: 1000,
    handleFunc: func(ctx *Context, statusCode int, err error) error {
        ctx.JSON(403, map[string]any{
            "code":    403,
            "message": "Forbidden",
            "success": false,
        })
        return nil
    },
}
```

#### 特性说明
- **触发场景**: 用户已认证但权限不足
- **应用场景**: 角色权限控制、资源访问限制
- **扩展性**: 可添加权限提示和申请流程

#### 使用示例
```go
func (c *AdminController) DeleteUser() {
    currentUser := c.GetCurrentUser()
    if !currentUser.HasPermission("delete_user") {
        errors.Handle(c.Ctx, 403, fmt.Errorf("您没有删除用户的权限"))
        return
    }
    
    // 删除用户逻辑...
}
```

### 3. 404 Not Found 处理器

```go
// 默认404处理逻辑
registry.defaultHandlers[404] = &FuncErrorHandler{
    name:     "default-404",
    priority: 1000,
    handleFunc: func(ctx *Context, statusCode int, err error) error {
        ctx.JSON(404, map[string]any{
            "code":    404,
            "message": "Not Found",
            "success": false,
        })
        return nil
    },
}
```

#### 特性说明
- **触发场景**: 路由不匹配、资源不存在
- **自动触发**: 框架路由系统自动调用
- **可定制**: 支持自定义404页面

#### 使用示例
```go
func (c *ArticleController) GetArticle() {
    id := c.GetInt("id", 0)
    
    article, err := articleService.GetByID(id)
    if err != nil {
        if errors.IsNotFound(err) {
            errors.Handle(c.Ctx, 404, fmt.Errorf("文章ID %d 不存在", id))
            return
        }
        errors.Handle(c.Ctx, 500, err)
        return
    }
    
    c.JSONSuccess("获取成功", article)
}
```

### 4. 500 Internal Server Error 处理器

```go
// 默认500处理逻辑（支持详细错误）
registry.defaultHandlers[500] = &FuncErrorHandler{
    name:     "default-500",
    priority: 1000,
    handleFunc: func(ctx *Context, statusCode int, err error) error {
        response := map[string]any{
            "code":    500,
            "message": "Internal Server Error",
            "success": false,
        }
        if config.ShowDetailedError && err != nil {
            response["error"] = err.Error()
        }
        ctx.JSON(500, response)
        return nil
    },
}
```

#### 特性说明
- **触发场景**: 系统异常、业务逻辑错误
- **错误详情**: 开发环境显示详细错误，生产环境隐藏
- **日志记录**: 自动记录错误堆栈信息

#### 使用示例
```go
func (c *OrderController) CreateOrder() {
    order := &Order{}
    if err := c.BindJSON(order); err != nil {
        errors.Handle(c.Ctx, 400, err)
        return
    }
    
    // 业务逻辑可能抛出异常
    if err := orderService.Create(order); err != nil {
        errors.Handle(c.Ctx, 500, fmt.Errorf("创建订单失败: %v", err))
        return
    }
    
    c.JSONSuccess("订单创建成功", order)
}
```

## 🎨 DefaultErrorController 详解

### HTML错误页面特性

DefaultErrorController 提供了功能强大的HTML错误页面：

#### 1. 响应式设计
```html
<!-- 页面自适应不同设备 -->
<meta name="viewport" content="width=device-width, initial-scale=1.0">

<!-- 移动端优化的CSS -->
@media (max-width: 768px) {
    .error-container {
        padding: 20px;
        font-size: 14px;
    }
}
```

#### 2. 现代化UI界面
- **Bootstrap 5.3.0** - 现代化组件库
- **渐进式动画** - CSS过渡效果
- **图标支持** - 状态对应的图标
- **色彩体系** - 基于错误类型的配色

#### 3. 错误详情展示
```html
<!-- 错误信息展示区域 -->
<div class="error-details">
    <h1 class="error-code">{{.StatusCode}}</h1>
    <p class="error-message">{{.StatusText}}</p>
    <div class="error-meta">
        <p><strong>请求路径:</strong> {{.RequestPath}}</p>
        <p><strong>请求方法:</strong> {{.RequestMethod}}</p>
        <p><strong>时间戳:</strong> {{.Timestamp}}</p>
    </div>
</div>
```

#### 4. 交互功能
```javascript
// 自动刷新功能
function autoRefresh() {
    setTimeout(function() {
        window.location.reload();
    }, 30000); // 30秒后刷新
}

// 错误报告功能  
function reportError() {
    const errorData = {
        statusCode: {{.StatusCode}},
        path: '{{.RequestPath}}',
        method: '{{.RequestMethod}}',
        userAgent: navigator.userAgent,
        timestamp: new Date().toISOString()
    };
    
    // 发送错误报告到后端
    fetch('/api/error-report', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(errorData)
    });
}
```

### 内容协商机制

DefaultErrorController 支持智能内容协商：

```go
func (c *DefaultErrorController) Handle(ctx *Context, statusCode int, err error) error {
    // 根据Accept头判断响应类型
    accept := string(ctx.GetHeader("Accept"))
    
    switch {
    case strings.Contains(accept, "application/json"):
        return c.handleJSON(ctx, statusCode, err)
    case strings.Contains(accept, "application/xml"):
        return c.handleXML(ctx, statusCode, err)
    default:
        return c.handleHTML(ctx, statusCode, err)
    }
}
```

## ⚙️ 配置和自定义

### 1. 环境相关配置

```go
// 开发环境配置
func setupDevelopmentHandlers() {
    errors.QuickSetup("development")
    
    // 开发环境特殊配置
    config := errors.GetConfig()
    config.ShowDetailedError = true
    config.VerboseLogging = true
    config.DebugMode = true
    errors.SetConfig(config)
}

// 生产环境配置
func setupProductionHandlers() {
    errors.QuickSetup("production")
    
    // 生产环境安全配置
    config := errors.GetConfig()
    config.ShowDetailedError = false
    config.VerboseLogging = false
    config.DebugMode = false
    errors.SetConfig(config)
}
```

### 2. 替换默认处理器

```go
// 替换默认404处理器
func setupCustom404() {
    // 注册自定义404处理器，优先级更高
    errors.RegisterFunc(404, "custom-404", 50,
        func(statusCode int, err error) bool {
            return statusCode == 404
        },
        func(ctx *errors.Context, statusCode int, err error) error {
            // 自定义404响应
            response := map[string]any{
                "code":    404,
                "message": "抱歉，您访问的页面不存在",
                "tips":    "请检查URL是否正确，或返回首页",
                "redirect": "/",
                "success": false,
            }
            
            ctx.JSON(404, response)
            return nil
        },
    )
}
```

### 3. 扩展默认处理器

```go
// 添加更多状态码的默认处理器
func registerAdditionalHandlers() {
    registry := errors.GetGlobalErrorRegistry()
    
    // 400 Bad Request
    registry.RegisterHandlerFunc(400, "default-400", 1000,
        func(statusCode int, err error) bool { return statusCode == 400 },
        func(ctx *errors.Context, statusCode int, err error) error {
            ctx.JSON(400, map[string]any{
                "code":    400,
                "message": "请求参数错误",
                "success": false,
            })
            return nil
        },
    )
    
    // 502 Bad Gateway
    registry.RegisterHandlerFunc(502, "default-502", 1000,
        func(statusCode int, err error) bool { return statusCode == 502 },
        func(ctx *errors.Context, statusCode int, err error) error {
            ctx.JSON(502, map[string]any{
                "code":    502,
                "message": "网关错误，请稍后重试",
                "success": false,
            })
            return nil
        },
    )
}
```

## 🔄 回退机制

默认处理器支持智能回退：

```go
// 默认回退规则（registry.go中的设置）
func setupDefaultFallbacks() {
    // 5xx错误回退到500
    for code := 501; code <= 511; code++ {
        fallbacks[code] = []int{500}
    }
    
    // 4xx错误的回退规则
    fallbacks[402] = []int{400}       // Payment Required -> Bad Request
    fallbacks[405] = []int{404, 400}  // Method Not Allowed -> Not Found -> Bad Request
    fallbacks[410] = []int{404, 400}  // Gone -> Not Found -> Bad Request
}
```

### 回退触发示例

```go
func testFallback() {
    // 当405错误没有专门的处理器时
    // 系统会依次尝试：405 -> 404 -> 400 -> 通用处理器
    
    errors.Handle(ctx, 405, fmt.Errorf("方法不被支持"))
    // 1. 查找405处理器 - 不存在
    // 2. 查找404处理器 - 存在，使用404处理器
    // 3. 返回404错误页面，但状态码仍为405
}
```

## 📊 性能和监控

### 处理器性能指标

```go
// 获取默认处理器的性能统计
func getHandlerMetrics() {
    registry := errors.GetGlobalErrorRegistry()
    metrics := registry.GetMetrics()
    
    // 查看各处理器的调用情况
    for handlerName, perf := range metrics.HandlerPerformance {
        if strings.HasPrefix(handlerName, "default-") {
            fmt.Printf("处理器: %s\n", handlerName)
            fmt.Printf("  调用次数: %d\n", perf.CallCount)
            fmt.Printf("  成功次数: %d\n", perf.SuccessCount)
            fmt.Printf("  平均耗时: %v\n", perf.AverageTime)
            fmt.Printf("  最后调用: %v\n", perf.LastCalled)
        }
    }
}
```

### 监控集成

```go
// 监控默认处理器的使用情况
func monitorDefaultHandlers() {
    registry := errors.GetGlobalErrorRegistry()
    
    // 定期输出统计信息
    go func() {
        ticker := time.NewTicker(time.Minute * 10)
        defer ticker.Stop()
        
        for range ticker.C {
            stats := registry.GetStats()
            if metrics, ok := stats["metrics"].(map[string]any); ok {
                statusCodes := metrics["status_code_counts"].(map[int]int64)
                
                fmt.Println("=== 默认处理器使用统计 ===")
                for _, code := range []int{401, 403, 404, 500} {
                    count := statusCodes[code]
                    fmt.Printf("%d错误: %d次\n", code, count)
                }
            }
        }
    }()
}
```

## 🚀 最佳实践

### 1. 合理使用默认处理器
```go
// ✅ 推荐：让默认处理器处理标准场景
func (c *UserController) GetUser() {
    user, err := userService.GetByID(id)
    if err != nil {
        if err == ErrUserNotFound {
            // 直接使用默认404处理器
            errors.Handle(c.Ctx, 404, err)
            return
        }
        // 直接使用默认500处理器
        errors.Handle(c.Ctx, 500, err) 
        return
    }
    
    c.JSONSuccess("获取成功", user)
}
```

### 2. 保持处理器一致性
```go
// ✅ 推荐：在应用启动时统一配置
func initializeErrorHandlers() {
    // 基础配置
    errors.QuickSetup(os.Getenv("APP_ENV"))
    
    // 统一的错误响应格式
    errors.Register(400, errors.JSON(400, map[string]any{
        "code":    400,
        "message": "请求参数错误",
        "success": false,
    }))
}
```

### 3. 测试默认处理器
```go
// 测试默认处理器是否按预期工作
func TestDefaultHandlers(t *testing.T) {
    app := setupTestApp()
    
    // 测试404处理器
    resp := app.Get("/nonexistent")
    assert.Equal(t, 404, resp.StatusCode)
    
    var result map[string]any
    json.Unmarshal(resp.Body, &result)
    assert.Equal(t, 404, result["code"])
    assert.Equal(t, "Not Found", result["message"])
    assert.Equal(t, false, result["success"])
}
```

## 📚 相关文档

- **[快速开始](quick-start.md)** - 了解如何快速启用默认处理器
- **[自定义处理器](custom-handlers.md)** - 学习如何自定义错误处理逻辑
- **[错误页面定制](error-pages.md)** - 定制DefaultErrorController的HTML页面
- **[最佳实践](best-practices.md)** - 默认处理器使用的最佳实践

---

> 💡 **提示**: 默认处理器为大多数场景提供了合理的默认行为，建议先使用默认处理器，然后根据具体需求进行定制。