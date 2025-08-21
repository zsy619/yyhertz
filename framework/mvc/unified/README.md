# 统一管理器 (Unified Manager)

统一管理器提供Cookie、Session、CSRF、模板、SSO功能的统一管理，通过统一的API接口减少重复代码，提高系统性能和可维护性。

## 核心特性

- **统一管理**: 集中管理Cookie、Session、CSRF、模板功能
- **高性能**: 使用单例模式和优化的数据结构
- **类型安全**: 支持类型安全的数据操作
- **过滤器支持**: 内置过滤器链机制，支持优先级排序
- **SSO集成**: 完整的单点登录(SSO)解决方案
- **中间件适配**: 提供中间件适配器，易于集成到现有框架

## 快速开始

### 基本使用

```go
// 获取统一管理器实例
manager := unified.GetManager()

// 使用Cookie功能
manager.SetCookie(ctx, "user_pref", "theme=dark")
value := manager.GetCookie(ctx, "user_pref")

// 使用Session功能
manager.SetSessionData(ctx, "user", userInfo)
user := manager.GetSessionData(ctx, "user")

// 使用上下文数据
manager.SetContextData(ctx, "authenticated", true)
isAuth := manager.GetContextData(ctx, "authenticated")

// 生成CSRF令牌
token, err := manager.GenerateCSRFToken("123", "192.168.1.1")

// 渲染模板
html, err := manager.RenderTemplate("user/profile", userData)
```

### 在BaseController中使用

```go
// 获取统一管理器（如果需要高级功能）
manager := controller.GetUnifiedManager()

// 便捷的上下文数据操作
controller.SetContextData("user_id", 123)
userID := controller.GetContextData("user_id")

// 类型安全的数据获取
userID, ok := controller.GetTypedContextData("user_id", 0)

// Cookie操作 - 统一接口，智能选择管理器
// SetCookie会自动优先使用统一管理器，然后回退到其他管理器
controller.SetCookie("session_id", sessionID)         // 智能选择最佳Cookie管理器
sessionID := controller.GetCookie("session_id")       // 从相同的Cookie管理器获取数据
controller.DeleteCookie("temp_cookie")               // 删除特定Cookie
hasAuth := controller.HasCookie("auth_token")         // 检查Cookie是否存在

// 安全Cookie操作（加密存储）
controller.SetSecureCookie("secret", "user_data", userData)
userData, ok := controller.GetSecureCookie("secret", "user_data")

// Session操作 - 统一接口，智能选择管理器
// SetSession会自动优先使用统一管理器，然后回退到其他管理器
controller.SetSession("user", userInfo)      // 智能选择最佳Session管理器
user := controller.GetSession("user")        // 从相同的Session存储获取数据
sessionID := controller.GetSessionID()      // 获取Session ID
controller.DeleteSession("temp_data")       // 删除特定数据
controller.DestroySession()                 // 清空整个Session

// SSO功能
err := controller.LoginUser(userInfo, true) // 登录并记住我
controller.LogoutUser() // 登出
user := controller.GetCurrentUser() // 获取当前用户
isAuthenticated := controller.IsUserAuthenticated() // 检查认证状态
```

#### Cookie和Session操作的智能选择机制

统一管理器提供了智能的Cookie和Session管理，所有操作方法会按以下优先级自动选择最佳的管理器：

**Cookie管理器选择顺序**：
1. **统一管理器的Cookie助手**：如果已初始化，优先使用（推荐）
2. **全局Cookie管理器**：如果可用，作为备选
3. **本地Cookie管理器**：最终回退选项

**Session管理器选择顺序**：
1. **缓存检查**：避免重复获取Session存储
2. **统一管理器的Session管理器**：如果已初始化，优先使用（推荐）
3. **全局Session管理器**：如果可用，作为备选
4. **本地Session管理器**：最终回退选项

这种设计确保了：
- **单一接口**：只需使用标准方法如`SetCookie/GetCookie`、`SetSession/GetSession`
- **性能优化**：Session存储会缓存在Context中，避免重复获取
- **向后兼容**：现有代码无需任何修改
- **自动优化**：系统自动选择最佳的管理器
- **透明升级**：应用自动享受统一管理器的性能和功能优势

## 过滤器系统

### 添加过滤器

```go
manager := unified.GetManager()

// 添加SSO过滤器
err := manager.AddFilter(&unified.Filter{
    Name: "sso",
    Pattern: "/*", // 匹配所有路径
    FilterFunc: unified.FilterSSO,
    Priority: 10, // 优先级（数值越小越先执行）
    Enabled: true,
    Description: "Single Sign-On authentication filter",
})

// 添加自定义过滤器
err := manager.AddFilter(&unified.Filter{
    Name: "api_auth",
    Pattern: "/api/*", // 只匹配API路径
    FilterFunc: func(ctx *context.Context) (unified.FilterResult, error) {
        // 自定义验证逻辑
        if !isValidAPIKey(ctx) {
            ctx.Request().SetStatusCode(401)
            return unified.FilterStop, nil
        }
        return unified.FilterContinue, nil
    },
    Priority: 20,
    Enabled: true,
})
```

### 执行过滤器

```go
// 手动执行过滤器链
path := "/api/users"
result, err := manager.ExecuteFilters(ctx, path)

switch result {
case unified.FilterContinue:
    // 继续处理请求
case unified.FilterStop:
    // 请求被拦截，不继续处理
case unified.FilterSkip:
    // 跳过当前处理，但继续执行
}
```

## SSO (单点登录)

### SSO配置

```go
config := &unified.SSOConfig{
    Enabled: true,
    Secret: "your-sso-secret-key",
    TokenExpireTime: 3600, // 1小时
    CookieName: "sso_token",
    CookiePath: "/",
    CookieSecure: false, // 生产环境建议设为true
    CookieHttpOnly: true,
    RememberCookieName: "remember_token",
    RememberExpireTime: 7 * 24 * 3600, // 7天
    ExcludePaths: []string{
        "/login", "/register", "/api/public/*",
    },
    LoginURL: "/login",
    LogoutURL: "/logout",
}

// 设置全局SSO配置
unified.SetSSOConfig(config)
```

### 用户登录和登出

```go
// 用户登录
userInfo := &unified.UserInfo{
    ID: "123",
    Username: "john",
    Email: "john@example.com",
    Roles: []string{"user", "admin"},
    Profile: make(map[string]interface{}),
}

// 登录用户（支持记住我功能）
err := unified.LoginUser(manager, ctx, userInfo, true)

// 登出用户
unified.LogoutUser(manager, ctx)

// 获取当前用户
currentUser := unified.GetCurrentUser(manager, ctx)

// 检查用户是否已认证
isAuthenticated := unified.IsAuthenticated(manager, ctx)
```

## 中间件集成

### 创建中间件

```go
// 创建中间件适配器
adapter := unified.NewMiddlewareAdapter()

// 转换为通用中间件
middleware := adapter.ToMiddlewareFunc()

// 使用中间件
middleware(ctx, func() {
    // 继续执行后续逻辑
})

// 创建SSO中间件
ssoMiddleware := unified.CreateGlobalSSO(nil) // 使用默认配置
```

### 带配置的中间件

```go
config := &unified.MiddlewareConfig{
    HandleError: true,
    LogErrors: true,
    SkipPaths: []string{"/health", "/metrics"},
}

middleware := adapter.ToMiddlewareFuncWithConfig(config)
```

### 链式中间件

```go
// 创建可链式调用的中间件
chainMiddleware := unified.NewChainableMiddleware([]string{
    "auth", "csrf", "sso",
})

handler := chainMiddleware.ToMiddlewareHandler()
```

## 高级功能

### 自定义验证器

```go
config := unified.GetSSOConfig()

// 自定义Token验证器
config.TokenValidator = func(token *unified.SSOToken) (*unified.UserInfo, error) {
    // 从数据库验证用户信息
    user, err := getUserFromDatabase(token.UserID)
    if err != nil {
        return nil, err
    }
    
    return &unified.UserInfo{
        ID: user.ID,
        Username: user.Username,
        Email: user.Email,
        Roles: user.Roles,
    }, nil
}

// 自定义记住我验证器
config.RememberValidator = func(rememberToken string) (*unified.UserInfo, error) {
    // 验证记住我令牌
    return validateRememberToken(rememberToken)
}
```

### 错误处理

```go
// 自定义错误处理器
config.ErrorHandler = func(ctx *context.Context, err error) {
    // 记录错误日志
    log.Printf("Filter error: %v", err)
    
    // 返回自定义错误响应
    ctx.Request().SetStatusCode(500)
    ctx.Request().Response.SetBodyString("Custom error response")
}
```

## 最佳实践

1. **单例使用**: 始终使用 `unified.GetManager()` 获取全局实例
2. **安全配置**: 生产环境中配置适当的Cookie安全选项
3. **路径匹配**: 合理配置过滤器的路径模式，避免不必要的执行
4. **优先级管理**: 设置合理的过滤器优先级，确保执行顺序正确
5. **错误处理**: 实现适当的错误处理和日志记录机制

## 性能优化

- 使用对象池化减少GC压力
- 并发安全的数据结构支持高并发访问
- 懒加载组件，按需初始化
- 高效的路径匹配算法

## 安全考虑

- CSRF保护机制
- 安全的Cookie配置选项
- IP地址验证
- 令牌签名验证
- 自动过期和清理机制

## 迁移指南

### 从旧版本升级

如果您之前使用了`SetUnifiedSessionData`、`GetUnifiedSessionData`、`SetUnifiedCookie`、`GetUnifiedCookie`等方法，请按以下方式迁移：

#### Session操作迁移
```go
// 旧代码（已废弃）
controller.SetUnifiedSessionData("user", userInfo)
user := controller.GetUnifiedSessionData("user")

// 新代码（推荐，智能选择管理器）
controller.SetSession("user", userInfo)
user := controller.GetSession("user")
```

#### Cookie操作迁移
```go
// 旧代码（已废弃）
controller.SetUnifiedCookie("auth", token)
token := controller.GetUnifiedCookie("auth")

// 新代码（推荐，智能选择管理器）
controller.SetCookie("auth", token)
token := controller.GetCookie("auth")
```

### 升级优势

- **零配置升级**：现有代码自动享受统一管理器的性能优势
- **API简化**：统一的接口，减少学习成本
- **性能提升**：智能管理器选择和缓存机制
- **功能增强**：完整的SSO、CSRF、过滤器支持

通过统一管理器，您可以大大简化Web应用的开发，提高代码的可维护性和系统性能。