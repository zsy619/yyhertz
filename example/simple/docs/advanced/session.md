# 会话管理

YYHertz 框架提供了强大而灵活的会话管理功能，支持多种存储后端和会话配置选项。

## 概述

会话管理是 Web 应用程序中的重要组件，用于在多个 HTTP 请求之间保持用户状态。YYHertz 提供了简单易用的会话 API，支持内存存储、Redis 存储等多种后端。

## 基本使用

### 启用会话中间件

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
    app := mvc.HertzApp
    
    // 使用默认的内存会话存储
    app.Use(middleware.SessionMiddleware())
    
    // 或者配置 Redis 会话存储
    app.Use(middleware.SessionMiddleware(middleware.SessionConfig{
        Store: "redis",
        RedisAddr: "localhost:6379",
        RedisPassword: "",
        RedisDB: 0,
    }))
    
    app.Run()
}
```

### 在控制器中使用会话

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
)

type UserController struct {
    mvc.Controller
}

func (c *UserController) Login() {
    // 获取会话
    session := c.GetSession()
    
    // 设置会话值
    session.Set("user_id", 12345)
    session.Set("username", "john_doe")
    session.Set("is_admin", true)
    
    // 保存会话
    if err := session.Save(); err != nil {
        c.JSON(500, map[string]string{"error": "Session save failed"})
        return
    }
    
    c.JSON(200, map[string]string{"status": "logged in"})
}

func (c *UserController) Profile() {
    session := c.GetSession()
    
    // 获取会话值
    userID := session.Get("user_id")
    username := session.Get("username")
    isAdmin := session.GetBool("is_admin")
    
    if userID == nil {
        c.JSON(401, map[string]string{"error": "Not logged in"})
        return
    }
    
    c.JSON(200, map[string]interface{}{
        "user_id": userID,
        "username": username,
        "is_admin": isAdmin,
    })
}

func (c *UserController) Logout() {
    session := c.GetSession()
    
    // 清除会话
    if err := session.Destroy(); err != nil {
        c.JSON(500, map[string]string{"error": "Logout failed"})
        return
    }
    
    c.JSON(200, map[string]string{"status": "logged out"})
}
```

## 会话配置

### 配置选项

```go
type SessionConfig struct {
    // 会话存储类型 ("memory", "redis")
    Store string `json:"store"`
    
    // Cookie 名称
    CookieName string `json:"cookie_name"`
    
    // 会话过期时间（秒）
    MaxAge int `json:"max_age"`
    
    // Cookie 路径
    Path string `json:"path"`
    
    // Cookie 域名
    Domain string `json:"domain"`
    
    // 是否仅限 HTTPS
    Secure bool `json:"secure"`
    
    // 是否仅限 HTTP（防止 XSS）
    HttpOnly bool `json:"http_only"`
    
    // SameSite 策略
    SameSite string `json:"same_site"`
    
    // Redis 配置
    RedisAddr     string `json:"redis_addr"`
    RedisPassword string `json:"redis_password"`
    RedisDB       int    `json:"redis_db"`
    RedisPrefix   string `json:"redis_prefix"`
}
```

### 使用配置文件

```yaml
# config.yaml
session:
  store: "redis"
  cookie_name: "session_id"
  max_age: 3600  # 1 hour
  path: "/"
  domain: ""
  secure: false
  http_only: true
  same_site: "lax"
  
  # Redis 配置
  redis_addr: "localhost:6379"
  redis_password: ""
  redis_db: 0
  redis_prefix: "session:"
```

```go
// 加载配置
config := LoadSessionConfig("config.yaml")
app.Use(middleware.SessionMiddleware(config))
```

## 会话存储

### 内存存储

内存存储适用于开发环境和单实例部署：

```go
app.Use(middleware.SessionMiddleware(middleware.SessionConfig{
    Store: "memory",
    MaxAge: 3600,
}))
```

**优点：**
- 简单快速
- 无外部依赖

**缺点：**
- 重启后会话丢失
- 不支持多实例部署

### Redis 存储

Redis 存储适用于生产环境和分布式部署：

```go
app.Use(middleware.SessionMiddleware(middleware.SessionConfig{
    Store: "redis",
    RedisAddr: "localhost:6379",
    RedisPassword: "your_password",
    RedisDB: 0,
    RedisPrefix: "app:session:",
    MaxAge: 7200, // 2 hours
}))
```

**优点：**
- 持久化存储
- 支持多实例部署
- 高性能

**缺点：**
- 需要 Redis 服务器
- 网络延迟

## 会话 API

### 基本操作

```go
session := c.GetSession()

// 设置值
session.Set("key", "value")
session.Set("count", 42)
session.Set("user", User{ID: 1, Name: "John"})

// 获取值
value := session.Get("key")                    // interface{}
count := session.GetInt("count")               // int
userName := session.GetString("user.name")     // string
userExists := session.Has("user")              // bool

// 删除值
session.Delete("key")

// 清除所有值
session.Clear()

// 销毁会话
session.Destroy()

// 保存会话
err := session.Save()
```

### 类型安全的获取方法

```go
// 获取不同类型的值
stringValue := session.GetString("name")           // 默认 ""
intValue := session.GetInt("age")                  // 默认 0
boolValue := session.GetBool("is_admin")           // 默认 false
floatValue := session.GetFloat64("price")          // 默认 0.0

// 带默认值的获取
name := session.GetStringDefault("name", "Guest")
age := session.GetIntDefault("age", 18)
```

### 会话元信息

```go
// 获取会话 ID
sessionID := session.ID()

// 检查会话是否为新会话
isNew := session.IsNew()

// 获取会话创建时间
createdAt := session.CreatedAt()

// 获取会话最后访问时间
lastAccess := session.LastAccess()

// 设置会话过期时间
session.SetMaxAge(7200) // 2 hours
```

## 安全考虑

### Cookie 安全

```go
middleware.SessionMiddleware(middleware.SessionConfig{
    // 仅限 HTTPS 传输
    Secure: true,
    
    // 防止 XSS 攻击
    HttpOnly: true,
    
    // SameSite 策略防止 CSRF
    SameSite: "strict",
    
    // 限制 Cookie 域名
    Domain: ".example.com",
})
```

### 会话固定攻击防护

```go
func (c *UserController) Login() {
    session := c.GetSession()
    
    // 登录成功后重新生成会话 ID
    if err := session.Regenerate(); err != nil {
        // 处理错误
        return
    }
    
    session.Set("user_id", userID)
    session.Save()
}
```

### 会话超时处理

```go
func (c *BaseController) RequireAuth() {
    session := c.GetSession()
    
    // 检查会话是否过期
    if session.IsExpired() {
        c.JSON(401, map[string]string{"error": "Session expired"})
        c.StopRun()
        return
    }
    
    // 更新最后访问时间
    session.Touch()
}
```

## 最佳实践

### 1. 会话数据最小化

```go
// 好的做法：只存储必要的标识符
session.Set("user_id", 12345)

// 避免：存储大量用户数据
// session.Set("user", fullUserObject) // 不推荐
```

### 2. 定期清理过期会话

```go
// 在应用启动时设置定时清理
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        sessionStore.CleanExpired()
    }
}()
```

### 3. 会话数据验证

```go
func (c *UserController) Profile() {
    session := c.GetSession()
    userID := session.GetInt("user_id")
    
    if userID == 0 {
        c.Redirect("/login")
        return
    }
    
    // 验证用户是否仍然有效
    user, err := c.userService.GetByID(userID)
    if err != nil || user == nil {
        session.Destroy()
        c.Redirect("/login")
        return
    }
    
    // 继续处理...
}
```

### 4. 错误处理

```go
func (c *UserController) UpdateProfile() {
    session := c.GetSession()
    
    // 设置会话数据
    session.Set("last_update", time.Now())
    
    // 始终检查保存错误
    if err := session.Save(); err != nil {
        c.Logger.Error("Failed to save session", "error", err)
        c.JSON(500, map[string]string{"error": "Internal server error"})
        return
    }
    
    c.JSON(200, map[string]string{"status": "updated"})
}
```

## 示例：完整的用户认证

```go
package controllers

import (
    "time"
    "github.com/zsy619/yyhertz/framework/mvc"
)

type AuthController struct {
    mvc.Controller
}

func (c *AuthController) Login() {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, map[string]string{"error": "Invalid request"})
        return
    }
    
    // 验证用户凭据
    user, err := c.authenticateUser(req.Username, req.Password)
    if err != nil {
        c.JSON(401, map[string]string{"error": "Invalid credentials"})
        return
    }
    
    // 创建会话
    session := c.GetSession()
    session.Set("user_id", user.ID)
    session.Set("username", user.Username)
    session.Set("role", user.Role)
    session.Set("login_time", time.Now())
    
    // 重新生成会话 ID 防止会话固定攻击
    if err := session.Regenerate(); err != nil {
        c.JSON(500, map[string]string{"error": "Session error"})
        return
    }
    
    if err := session.Save(); err != nil {
        c.JSON(500, map[string]string{"error": "Session save failed"})
        return
    }
    
    c.JSON(200, map[string]interface{}{
        "status": "success",
        "user": map[string]interface{}{
            "id": user.ID,
            "username": user.Username,
            "role": user.Role,
        },
    })
}

func (c *AuthController) Logout() {
    session := c.GetSession()
    
    if err := session.Destroy(); err != nil {
        c.JSON(500, map[string]string{"error": "Logout failed"})
        return
    }
    
    c.JSON(200, map[string]string{"status": "logged out"})
}

func (c *AuthController) GetCurrentUser() {
    session := c.GetSession()
    
    userID := session.GetInt("user_id")
    if userID == 0 {
        c.JSON(401, map[string]string{"error": "Not authenticated"})
        return
    }
    
    c.JSON(200, map[string]interface{}{
        "user_id": userID,
        "username": session.GetString("username"),
        "role": session.GetString("role"),
        "login_time": session.Get("login_time"),
    })
}
```

## 故障排除

### 常见问题

1. **会话数据丢失**
   - 检查 Cookie 设置
   - 确认会话存储服务状态
   - 验证会话过期时间

2. **Redis 连接失败**
   - 检查 Redis 服务器状态
   - 验证连接参数
   - 查看网络连接

3. **会话性能问题**
   - 减少会话数据大小
   - 使用 Redis 集群
   - 调整过期时间

### 调试技巧

```go
// 启用会话调试日志
middleware.SessionMiddleware(middleware.SessionConfig{
    Debug: true,
    Logger: logger,
})

// 在控制器中打印会话信息
func (c *BaseController) debugSession() {
    session := c.GetSession()
    c.Logger.Info("Session Debug",
        "id", session.ID(),
        "is_new", session.IsNew(),
        "data", session.All(),
    )
}
```

会话管理是构建现代 Web 应用程序的基础功能，YYHertz 提供的会话系统既简单易用又功能强大，能够满足各种应用场景的需求。
