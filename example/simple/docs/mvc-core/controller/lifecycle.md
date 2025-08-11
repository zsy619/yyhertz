# 🔄 控制器生命周期管理

YYHertz控制器拥有完整的生命周期管理机制，从初始化到销毁的每个阶段都可以进行精确控制。

## 📋 生命周期概览

### 核心生命周期方法

```go
// 1️⃣ 初始化控制器
func (c *BaseController) Init(ct *context.Context, controllerName, actionName string, app any)

// 2️⃣ 预处理（可重写）
func (c *BaseController) Prepare()

// 3️⃣ 执行业务逻辑（自动路由到具体方法）
// GetUser(), PostCreate(), PutUpdate(), DeleteUser() 等

// 4️⃣ 后处理（可重写）
func (c *BaseController) Finish()

// 5️⃣ 资源清理
func (c *BaseController) Destroy() error
```

### 生命周期流程图

```
    请求到达
        ↓
    Init() - 初始化控制器实例
        ↓
    Prepare() - 预处理逻辑
        ↓
    执行具体Action方法 (Get/Post/Put/Delete等)
        ↓
    Finish() - 后处理逻辑
        ↓
    Destroy() - 资源清理
        ↓
    响应返回
```

## 🔧 详细方法说明

### 1. Init() - 控制器初始化

**方法签名:**
```go
func (c *BaseController) Init(ct *context.Context, controllerName, actionName string, app any)
```

**功能:**
- 设置上下文对象
- 初始化控制器名称和动作名称
- 注入应用实例
- 初始化辅助组件 (Cookie、Session、模板引擎等)

**自动调用:** 框架自动调用，无需手动执行

### 2. Prepare() - 请求预处理

**方法签名:**
```go
func (c *BaseController) Prepare()
```

**功能:**
- 身份验证和权限检查
- 设置通用数据和变量
- 启用安全防护机制
- 记录请求日志

**重写示例:**
```go
func (c *UserController) Prepare() {
    // 调用父类方法
    c.BaseController.Prepare()
    
    // 身份验证
    if !c.IsAuthenticated() {
        c.Redirect("/login")
        c.StopRun()
        return
    }
    
    // 设置通用数据
    c.SetData("CurrentUser", c.GetSession("user"))
    c.SetData("CSRFToken", c.XSRFToken())
}
```

### 3. Action执行 - 业务逻辑处理

**自动路由规则:**
- `GetIndex()` → GET /
- `GetUser()` → GET /user
- `PostCreate()` → POST /create
- `PutUpdate()` → PUT /update
- `DeleteUser()` → DELETE /user

**业务方法示例:**
```go
func (c *UserController) GetUser() {
    userID := c.GetParam("id")
    user, err := c.userService.GetByID(userID)
    if err != nil {
        c.JSONError("用户不存在")
        return
    }
    c.JSONSuccess("获取成功", user)
}
```

### 4. Finish() - 请求后处理

**方法签名:**
```go
func (c *BaseController) Finish()
```

**功能:**
- 记录性能统计
- 清理敏感数据
- 执行审计日志
- 资源释放

**重写示例:**
```go
func (c *UserController) Finish() {
    // 记录响应时间
    if startTime := c.GetData("start_time"); startTime != nil {
        duration := time.Since(startTime.(time.Time))
        c.LogInfo("请求耗时", map[string]any{
            "duration": duration.String(),
            "path":     c.Ctx.Path(),
        })
    }
    
    // 清理敏感数据
    c.DelData("password")
    c.DelData("secret_key")
    
    // 调用父类方法
    c.BaseController.Finish()
}
```

### 5. Destroy() - 资源清理

**方法签名:**
```go
func (c *BaseController) Destroy() error
```

**功能:**
- 关闭数据库连接
- 释放文件句柄
- 清理缓存数据
- 归还对象池资源

**自动调用:** 框架自动调用，确保资源释放

## 🎯 生命周期最佳实践

### 完整示例 - 用户控制器

```go
package controllers

import (
    "time"
    "github.com/zsy619/yyhertz/framework/mvc/core"
)

type UserController struct {
    core.BaseController
    userService *UserService
    startTime   time.Time
}

// 预处理 - 每个请求前执行
func (c *UserController) Prepare() {
    c.startTime = time.Now()
    c.SetData("start_time", c.startTime)
    
    // 🔐 身份验证检查
    if c.needsAuth() && !c.IsAuthenticated() {
        if c.IsAjax() {
            c.JSONError("请先登录")
        } else {
            c.Redirect("/login")
        }
        c.StopRun()
        return
    }
    
    // 📊 设置通用数据
    if user := c.GetCurrentUser(); user != nil {
        c.SetData("CurrentUser", user)
        c.SetData("IsLoggedIn", true)
    }
    
    // 🛡️ 启用XSRF保护（POST/PUT/DELETE请求）
    if c.needsXSRF() {
        c.EnableXSRF(3600) // 1小时有效期
        if !c.CheckXSRFCookie() {
            c.JSONError("CSRF令牌验证失败")
            c.StopRun()
            return
        }
    }
    
    // 📝 记录请求信息
    c.LogInfo("处理用户请求", map[string]any{
        "method":     c.Ctx.Method(),
        "path":       c.Ctx.Path(),
        "ip":         c.GetClientIP(),
        "user_agent": c.GetHeader("User-Agent"),
    })
}

// 业务方法示例
func (c *UserController) GetProfile() {
    userID := c.GetParam("id")
    if userID == "" {
        c.JSONError("用户ID不能为空")
        return
    }
    
    user, err := c.userService.GetProfile(userID)
    if err != nil {
        c.JSONError("获取用户信息失败: " + err.Error())
        return
    }
    
    c.JSONSuccess("获取成功", user)
}

func (c *UserController) PostUpdate() {
    userID := c.GetParam("id")
    name := c.GetString("name")
    email := c.GetString("email")
    
    // 参数验证
    if name == "" || email == "" {
        c.JSONError("姓名和邮箱不能为空")
        return
    }
    
    // 更新用户信息
    err := c.userService.UpdateProfile(userID, name, email)
    if err != nil {
        c.JSONError("更新失败: " + err.Error())
        return
    }
    
    c.JSONSuccess("更新成功", nil)
}

// 后处理 - 每个请求后执行
func (c *UserController) Finish() {
    // ⏱️ 记录性能指标
    duration := time.Since(c.startTime)
    statusCode := c.Ctx.Response.StatusCode()
    
    c.LogInfo("请求完成", map[string]any{
        "duration":    duration.String(),
        "status_code": statusCode,
        "path":        c.Ctx.Path(),
        "method":      c.Ctx.Method(),
    })
    
    // 🧹 清理敏感数据
    sensitiveKeys := []string{"password", "token", "secret", "private_key"}
    for _, key := range sensitiveKeys {
        c.DelData(key)
    }
    
    // 📊 性能监控（慢请求告警）
    if duration > time.Second {
        c.LogWarn("慢请求检测", map[string]any{
            "duration": duration.String(),
            "path":     c.Ctx.Path(),
            "method":   c.Ctx.Method(),
        })
    }
    
    // 🔄 调用父类方法进行基础清理
    c.BaseController.Finish()
}

// 辅助方法
func (c *UserController) needsAuth() bool {
    // 需要身份验证的路径
    authPaths := []string{"/profile", "/update", "/delete"}
    currentPath := c.Ctx.Path()
    
    for _, path := range authPaths {
        if strings.Contains(currentPath, path) {
            return true
        }
    }
    return false
}

func (c *UserController) needsXSRF() bool {
    method := c.Ctx.Method()
    return method == "POST" || method == "PUT" || method == "DELETE"
}
```

### 错误处理最佳实践

```go
func (c *UserController) Prepare() {
    // 统一错误恢复
    defer func() {
        if r := recover(); r != nil {
            c.LogError("Prepare阶段发生panic", map[string]any{
                "error": r,
                "path":  c.Ctx.Path(),
            })
            c.JSONError("服务器内部错误")
            c.StopRun()
        }
    }()
    
    // 原有逻辑...
}
```

## 🔍 调试和监控

### 生命周期日志记录

```go
// 在每个生命周期方法中添加日志
func (c *UserController) Prepare() {
    c.LogDebug("进入Prepare阶段", map[string]any{
        "controller": c.ControllerName,
        "action":     c.ActionName,
    })
    
    // ... 处理逻辑
    
    c.LogDebug("Prepare阶段完成", nil)
}
```

### 性能监控

```go
func (c *UserController) Finish() {
    // 记录详细性能指标
    metrics := map[string]any{
        "duration":     time.Since(c.startTime),
        "memory_used":  runtime.MemStats{}.Alloc,
        "goroutines":   runtime.NumGoroutine(),
        "status_code":  c.Ctx.Response.StatusCode(),
    }
    
    c.LogInfo("性能指标", metrics)
}
```

## ❓ 常见问题

**Q: 什么时候应该重写Prepare方法？**
A: 当需要进行身份验证、权限检查、设置通用数据或启用安全防护时。

**Q: Finish方法是否总是会执行？**
A: 是的，即使在Action方法中发生异常，Finish方法也会被执行，类似于finally块。

**Q: 可以在Prepare中调用StopRun吗？**
A: 可以，StopRun会终止后续的Action执行，但Finish方法仍会被调用。

**Q: 如何在生命周期中传递数据？**
A: 使用SetData/GetData方法在不同生命周期阶段之间传递数据。