# YYHertz 路由API进化指南

## 🎉 API进化历程：三代路由API

YYHertz框架经历了三代API演进，每一代都在简化开发体验的同时保持功能完整性：

1. **第一代：原始API** - 功能完整但相对复杂
2. **第二代：简化API** - 使用 `func(context.Context)` 简化参数  
3. **第三代：直接API** - 直接传递 `*contextenhanced.Context`，最简洁

## 🔄 三代API对比

### 第一代：原始API（保持向后兼容）
```go
mvc.GET("/users", func(ctx context.Context, c *core.RequestContext) {
    enhancedCtx := contextenhanced.NewContext((*app.RequestContext)(c))
    enhancedCtx.JSON(200, map[string]interface{}{"users": []string{}})
})
```

### 第二代：简化API 
```go
mvc.SimpleGET("/users", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    c.JSON(200, map[string]interface{}{"users": []string{}})
})
```

### 第三代：直接API ✨（推荐）
```go
mvc.DirectGET("/users", func(c *contextenhanced.Context) {
    c.JSON(200, map[string]interface{}{"users": []string{}})
})
```

## 🚀 核心优势

### 第三代Direct API的优势
1. **极致简洁**: 直接使用增强Context，无需额外调用
2. **零开销**: 没有context.Value的查找开销
3. **类型安全**: 直接传递具体类型，编译时检查
4. **API一致性**: 与其他主流框架(Gin、Echo)风格类似
5. **完整功能**: 保留YYHertz所有高级特性

### 第二代Simple API的优势  
1. **更符合Go标准**: 使用标准的 `context.Context` 接口
2. **API更简洁**: 只需要一个参数，减少认知负担
3. **类型安全**: 避免手动类型转换
4. **完整功能**: 通过 `mvc.FromContext(ctx)` 获取完整的增强Context

### 向后兼容承诺
5. **向后兼容**: 三代API可以在同一应用中并存，可以渐进式迁移

## 📋 完整API列表

### HTTP方法路由

```go
// 第三代: 直接API (推荐) ⭐
mvc.DirectGET(path, handlers...)     // GET请求
mvc.DirectPOST(path, handlers...)    // POST请求
mvc.DirectPUT(path, handlers...)     // PUT请求
mvc.DirectDELETE(path, handlers...)  // DELETE请求
mvc.DirectPATCH(path, handlers...)   // PATCH请求
mvc.DirectHEAD(path, handlers...)    // HEAD请求
mvc.DirectOPTIONS(path, handlers...) // OPTIONS请求
mvc.DirectAny(path, handlers...)     // 任意HTTP方法

// 第二代: 简化API
mvc.SimpleGET(path, handlers...)     // GET请求
mvc.SimplePOST(path, handlers...)    // POST请求
mvc.SimplePUT(path, handlers...)     // PUT请求
mvc.SimpleDELETE(path, handlers...)  // DELETE请求
mvc.SimplePATCH(path, handlers...)   // PATCH请求
mvc.SimpleHEAD(path, handlers...)    // HEAD请求
mvc.SimpleOPTIONS(path, handlers...) // OPTIONS请求
mvc.SimpleAny(path, handlers...)     // 任意HTTP方法

// 第一代: 原有API (向后兼容)
mvc.GET(path, handlers...)           // 传统方式
mvc.POST(path, handlers...)          // 传统方式
// ... 其他方法
```

### Context获取API

```go
// 从 context.Context 获取增强Context
c := mvc.FromContext(ctx)           // 安全获取，可能返回nil
c := mvc.MustFromContext(ctx)       // 强制获取，不存在时panic
```

## 💡 使用示例

### 🌟 Direct API 示例（推荐）

#### 1. 基础路由注册

```go
func main() {
    app := mvc.GetAppInstance()
    mvc.HertzApp = app

    // 用户列表 - 最简洁的写法
    mvc.DirectGET("/api/users", func(c *contextenhanced.Context) {
        c.JSON(200, map[string]interface{}{
            "users": []map[string]interface{}{
                {"id": 1, "name": "张三"},
                {"id": 2, "name": "李四"},
            },
        })
    })

    // 创建用户
    mvc.DirectPOST("/api/users", func(c *contextenhanced.Context) {
        c.JSON(201, map[string]interface{}{
            "message": "用户创建成功",
            "id":      123,
        })
    })

    app.Run()
}
```

#### 2. 路由参数和查询参数

```go
// 获取路由参数
mvc.DirectGET("/api/users/:id", func(c *contextenhanced.Context) {
    id := c.Param("id")
    
    c.JSON(200, map[string]interface{}{
        "user": map[string]interface{}{
            "id":   id,
            "name": "用户" + id,
        },
    })
})

// 获取查询参数
mvc.DirectGET("/api/search", func(c *contextenhanced.Context) {
    keyword := c.Query("q")
    limit := c.Query("limit", "10") // 带默认值
    
    c.JSON(200, map[string]interface{}{
        "keyword": keyword,
        "limit":   limit,
        "results": []string{"结果1", "结果2"},
    })
})
```

#### 3. 请求体和表单处理

```go
mvc.DirectPOST("/api/login", func(c *contextenhanced.Context) {
    // 获取表单数据
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    // 或者解析JSON
    var jsonData map[string]interface{}
    if err := c.BindJSON(&jsonData); err == nil {
        username = jsonData["username"].(string)
        password = jsonData["password"].(string)
    }
    
    c.JSON(200, map[string]interface{}{
        "message": "登录成功",
        "username": username,
    })
})
```

#### 4. 中间件链式调用

```go
// 认证中间件
authMiddleware := func(c *contextenhanced.Context) {
    token := c.GetHeader("Authorization")
    if token == "" {
        c.JSON(401, map[string]interface{}{
            "error": "未授权访问",
        })
        c.Abort()
        return
    }
    
    // 设置用户信息
    c.Set("user_id", "12345")
}

// 日志中间件
logMiddleware := func(c *contextenhanced.Context) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        fmt.Printf("Request: %s %s - %v\n", 
            string(c.Method()), string(c.Path()), duration)
    }()
}

// 受保护的路由
mvc.DirectGET("/api/profile", logMiddleware, authMiddleware, 
    func(c *contextenhanced.Context) {
        userID := c.GetString("user_id")
        c.JSON(200, map[string]interface{}{
            "message": "用户资料",
            "user_id": userID,
        })
    })
```

### 📚 Simple API 示例

### 1. 基础路由注册

```go
func main() {
    app := mvc.GetAppInstance()
    mvc.HertzApp = app

    // 用户列表
    mvc.SimpleGET("/api/users", func(ctx context.Context) {
        c := mvc.FromContext(ctx)
        c.JSON(200, map[string]interface{}{
            "users": []map[string]interface{}{
                {"id": 1, "name": "张三"},
                {"id": 2, "name": "李四"},
            },
        })
    })

    // 创建用户
    mvc.SimplePOST("/api/users", func(ctx context.Context) {
        c := mvc.FromContext(ctx)
        c.JSON(201, map[string]interface{}{
            "message": "用户创建成功",
            "id":      123,
        })
    })

    app.Run()
}
```

### 2. 路由参数处理

```go
// 获取路由参数
mvc.SimpleGET("/api/users/:id", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    id := c.Param("id")
    
    c.JSON(200, map[string]interface{}{
        "user": map[string]interface{}{
            "id":   id,
            "name": "用户" + id,
        },
    })
})

// 获取查询参数
mvc.SimpleGET("/api/search", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    keyword := c.Query("q")
    limit := c.Query("limit")
    
    c.JSON(200, map[string]interface{}{
        "keyword": keyword,
        "limit":   limit,
        "results": []string{"结果1", "结果2"},
    })
})
```

### 3. 请求头和表单处理

```go
mvc.SimplePOST("/api/login", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    
    // 获取请求头
    contentType := c.Header("Content-Type")
    authorization := c.Header("Authorization")
    
    // 获取表单数据
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    c.JSON(200, map[string]interface{}{
        "message": "登录处理",
        "headers": map[string]string{
            "content_type":  contentType,
            "authorization": authorization,
        },
        "form": map[string]string{
            "username": username,
            "password": password,
        },
    })
})
```

### 4. 错误处理和中断

```go
mvc.SimpleGET("/api/protected", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    
    // 权限检查
    token := c.Header("Authorization")
    if token == "" {
        c.JSON(401, map[string]interface{}{
            "error": "未授权访问",
            "code":  401,
        })
        c.Abort() // 中断请求处理
        return
    }
    
    // 正常处理
    c.JSON(200, map[string]interface{}{
        "message": "访问成功",
        "data":    "受保护的数据",
    })
})
```

### 5. 上下文值存储

```go
mvc.SimpleGET("/api/context", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    
    // 设置值
    c.Set("request_id", "req_123456")
    c.Set("user_role", "admin")
    
    // 获取值
    requestID, exists := c.Get("request_id")
    if !exists {
        c.JSON(500, map[string]interface{}{"error": "上下文值丢失"})
        return
    }
    
    userRole := c.MustGet("user_role") // 强制获取，不存在会panic
    
    c.JSON(200, map[string]interface{}{
        "request_id": requestID,
        "user_role":  userRole,
        "message":    "上下文处理成功",
    })
})
```

### 6. 多个处理器链式调用

```go
// 中间件函数
func authMiddleware(ctx context.Context) {
    c := mvc.FromContext(ctx)
    
    token := c.Header("Authorization")
    if token == "" {
        c.JSON(401, map[string]interface{}{"error": "未授权"})
        c.Abort()
        return
    }
    
    // 验证token并设置用户信息
    c.Set("user_id", "123")
    c.Next() // 继续执行下一个处理器
}

func logMiddleware(ctx context.Context) {
    c := mvc.FromContext(ctx)
    start := time.Now()
    
    defer func() {
        duration := time.Since(start)
        method := string(c.Request.Method())
        path := string(c.Request.Path())
        fmt.Printf("[%s] %s - %v\n", method, path, duration)
    }()
    
    c.Next()
}

// 注册带中间件的路由
mvc.SimpleGET("/api/admin/users", logMiddleware, authMiddleware, func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    userID := c.MustGet("user_id")
    
    c.JSON(200, map[string]interface{}{
        "message": "管理员用户列表",
        "user_id": userID,
        "users":   []string{"admin1", "admin2"},
    })
})
```

### 7. 响应类型处理

```go
mvc.SimpleGET("/api/response-types", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    
    responseType := c.Query("type")
    
    switch responseType {
    case "json":
        c.JSON(200, map[string]interface{}{
            "type": "json",
            "data": "JSON响应",
        })
    case "string":
        c.String(200, "字符串响应")
    case "html":
        c.HTML(200, "template.html", map[string]interface{}{
            "title": "HTML响应",
        })
    default:
        c.JSON(400, map[string]interface{}{
            "error": "不支持的响应类型",
        })
    }
})
```

## 🔧 技术实现细节

### Context传递机制

```go
// 内部实现原理
type contextKey struct{}
var enhancedContextKey = &contextKey{}

// 注入增强Context
func WithContext(ctx context.Context, enhancedCtx *Context) context.Context {
    return context.WithValue(ctx, enhancedContextKey, enhancedCtx)
}

// 获取增强Context
func FromContext(ctx context.Context) *Context {
    if enhancedCtx, ok := ctx.Value(enhancedContextKey).(*Context); ok {
        return enhancedCtx
    }
    return nil
}
```

### 适配器函数

```go
func AdaptSimpleHandlerToHertz(handler SimpleHandlerFunc) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        // 1. 创建增强Context
        enhancedCtx := contextenhanced.NewContext(c)
        
        // 2. 执行过滤器链 (BeforeStatic -> BeforeRouter -> BeforeExec)
        
        // 3. 注入增强Context到context.Context
        ctxWithEnhanced := WithContext(ctx, enhancedCtx)
        
        // 4. 调用处理函数
        handler(ctxWithEnhanced)
        
        // 5. 执行后置过滤器 (AfterExec -> FinishRouter)
    }
}
```

## 🎯 迁移指南

### 推荐的API选择策略

```go
// 新项目: 直接使用Direct API (推荐)
mvc.DirectGET("/api/users", func(c *contextenhanced.Context) {
    c.JSON(200, userData)
})

// 现有项目: 可选择渐进式迁移
// 第一步: 迁移到Simple API
mvc.SimpleGET("/api/users", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    c.JSON(200, userData)
})

// 第二步: 进一步迁移到Direct API
mvc.DirectGET("/api/users", func(c *contextenhanced.Context) {
    c.JSON(200, userData)
})
```

### 三步迁移路径

#### 从原始API → Simple API
```go
// 步骤1: 更新函数签名
// 从:
func oldHandler(ctx context.Context, c *core.RequestContext) {
    enhancedCtx := contextenhanced.NewContext((*app.RequestContext)(c))
    enhancedCtx.JSON(200, data)
}

// 改为:
func simpleHandler(ctx context.Context) {
    c := mvc.FromContext(ctx)
    c.JSON(200, data)
}

// 步骤2: 更新路由注册
// 从: mvc.GET("/path", oldHandler)
// 改为: mvc.SimpleGET("/path", simpleHandler)
```

#### 从Simple API → Direct API
```go
// 步骤1: 更新函数签名
// 从:
func simpleHandler(ctx context.Context) {
    c := mvc.FromContext(ctx)
    c.JSON(200, data)
}

// 改为:
func directHandler(c *contextenhanced.Context) {
    c.JSON(200, data)
}

// 步骤2: 更新路由注册
// 从: mvc.SimpleGET("/path", simpleHandler)
// 改为: mvc.DirectGET("/path", directHandler)
```

### 渐进式迁移策略

1. **新项目推荐Direct API**: 所有新开发的功能使用 `DirectXXX` 方法，获得最佳开发体验
2. **现有项目混合使用**: 新功能用Direct API，现有功能保持不变或选择性迁移
3. **分阶段迁移**: 原始API → Simple API → Direct API，每一步都是可选的
4. **三代并存**: 三代API可以在同一应用中完美并存，无任何冲突

## 📊 性能对比

| 特性 | 原始API | Simple API | Direct API | 说明 |
|------|---------|------------|------------|------|
| 参数数量 | 2个 | 1个 | 1个 | Direct API最简洁 |
| 类型转换 | 需要 | 无需 | 无需 | Direct/Simple API更安全 |
| 内存开销 | 基线 | +context.Value | 基线 | Direct API零开销 |
| 查找开销 | 无 | context.Value查找 | 无 | Direct API无查找开销 |
| 编译检查 | ✅ | ✅ | ✅ | 都支持编译时类型检查 |
| 过滤器集成 | ✅ | ✅ | ✅ | 完全相同 |
| 功能完整性 | ✅ | ✅ | ✅ | 完全相同 |
| 开发体验 | 一般 | 良好 | 极佳 | Direct API最直观 |

## 🤝 与其他框架对比

```go
// Gin风格
router.GET("/users", func(c *gin.Context) {
    c.JSON(200, gin.H{"users": []string{}})
})

// Echo风格  
e.GET("/users", func(c echo.Context) error {
    return c.JSON(200, map[string]interface{}{"users": []string{}})
})

// Fiber风格
app.Get("/users", func(c *fiber.Ctx) error {
    return c.JSON(map[string]interface{}{"users": []string{}})
})

// YYHertz Direct风格 ⭐ (推荐)
mvc.DirectGET("/users", func(c *contextenhanced.Context) {
    c.JSON(200, map[string]interface{}{"users": []string{}})
})

// YYHertz Simple风格
mvc.SimpleGET("/users", func(ctx context.Context) {
    c := mvc.FromContext(ctx)
    c.JSON(200, map[string]interface{}{"users": []string{}})
})
```

## 🌟 API演进总结

YYHertz框架的API演进体现了对开发者体验的持续优化：

1. **第一代（原始API）**: 功能强大，但需要手动处理Context转换
2. **第二代（Simple API）**: 引入context.Context标准，简化参数传递  
3. **第三代（Direct API）**: 达到与主流框架相当的简洁性，同时保持强大功能

## ✅ 最佳实践

1. **新项目优先使用Direct API** - 获得最佳开发体验
2. **统一项目内API风格** - 避免混用，保持代码一致性  
3. **合理使用中间件** - 充分利用链式调用特性
4. **错误处理要统一** - 建立项目统一的错误处理模式
5. **保持向后兼容性** - 老项目可渐进式迁移，无强制要求

## 🎉 结语

YYHertz的三代API设计展现了框架的演进思路：**在保持强大功能的同时，不断简化开发者体验**。

- **功能性**: 三代API功能完全相同，都支持完整的YYHertz特性
- **兼容性**: 三代API可以完美并存，支持渐进式升级  
- **简洁性**: Direct API达到业界最简洁的路由注册体验

选择适合你项目的API风格，享受高效的Web开发体验！