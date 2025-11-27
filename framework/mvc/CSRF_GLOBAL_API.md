# 全局CSRF Token API 使用指南

本文档介绍如何使用 @framework/mvc 提供的全局CSRF token功能。

## 概述

全局CSRF管理器提供了统一的CSRF token生成和验证功能，支持：
- 安全的CSRF token生成和验证
- 与模板引擎的无缝集成
- 灵活的配置选项
- 自定义跳过检查规则

## 快速开始

### 1. 初始化

在应用启动时初始化全局管理器：

```go
import "github.com/zsy619/yyhertz/framework/mvc"

func init() {
    // 初始化所有全局管理器（包括CSRF）
    mvc.InitializeGlobalManagers()
}
```

### 2. 生成CSRF Token

#### 生成完整的CSRF Token

```go
import "github.com/zsy619/yyhertz/framework/mvc"

// 在控制器中生成CSRF token
func (c *Controller) Login() {
    userID := c.GetSessionID()  // 或其他用户标识
    clientIP := c.GetClientIP()
    
    token, err := mvc.GenerateCSRFToken(userID, clientIP)
    if err != nil {
        c.Error500("生成CSRF token失败")
        return
    }
    
    // 设置Cookie
    c.SetCookie("_csrf_token", token.Value, &cookie.Options{
        MaxAge:   3600,
        HttpOnly: false, // CSRF token需要被JavaScript访问
        Path:     "/",
        SameSite: "Strict",
    })
    
    // 添加到模板数据
    c.Data["csrf_token"] = token.Value
    c.TplName = "login.html"
}
```

#### 生成简单的CSRF Token（推荐用于模板）

```go
// 生成简单token，不依赖用户ID和IP
token := mvc.GenerateSimpleCSRFToken()
c.Data["csrf_token"] = token
```

### 3. 验证CSRF Token

```go
func (c *Controller) LoginPost() {
    // 从表单或请求头获取token
    tokenValue := c.GetString("csrf_token")
    if tokenValue == "" {
        tokenValue = c.Ctx.Request.Header.Get("X-CSRF-Token")
    }
    
    userID := c.GetSessionID()
    clientIP := c.GetClientIP()
    
    // 验证token
    isValid, err := mvc.ValidateCSRFToken(tokenValue, userID, clientIP)
    if err != nil || !isValid {
        c.Error403("CSRF token验证失败")
        return
    }
    
    // 处理登录逻辑
    // ...
}
```

## 模板中使用

### 1. 在HTML表单中

```html
<form method="POST" action="/login">
    <!-- 使用隐藏字段 -->
    <input type="hidden" name="csrf_token" value="{{.CsrfToken}}" />
    
    <!-- 或者使用模板函数 -->
    <input type="hidden" name="csrf_token" value="{{csrf_token}}" />
    
    <input type="text" name="username" required>
    <input type="password" name="password" required>
    <button type="submit">登录</button>
</form>
```

### 2. 在JavaScript中

```html
<script>
// 从模板数据获取token
const csrfToken = "{{.CsrfToken}}";

// AJAX请求中使用
fetch('/api/data', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': csrfToken
    },
    body: JSON.stringify({data: 'example'})
});
</script>
```

## 配置管理

### 1. 获取当前配置

```go
config, err := mvc.GetCSRFConfig()
if err != nil {
    log.Printf("获取CSRF配置失败: %v", err)
    return
}

fmt.Printf("Cookie名称: %s\n", config.CookieName)
fmt.Printf("过期时间: %d秒\n", config.ExpireTime)
```

### 2. 更新配置

```go
import "github.com/zsy619/yyhertz/framework/mvc/security"

// 创建自定义配置
customConfig := &security.CSRFConfig{
    Secret:        "your-secret-key-here",
    TokenLength:   64,
    ExpireTime:    7200, // 2小时
    CookieName:    "_my_csrf_token",
    HeaderName:    "X-My-CSRF-Token",
    FormFieldName: "my_csrf_token",
}

// 应用配置
mvc.UpdateCSRFConfig(customConfig)
```

### 3. 便捷配置方法

```go
// 获取配置的字段名
tokenName := mvc.GetCSRFTokenName()     // "csrf_token"
cookieName := mvc.GetCSRFCookieName()   // "_csrf_token"
headerName := mvc.GetCSRFHeaderName()   // "X-CSRF-Token"

// 检查保护状态
if mvc.IsCSRFProtectionEnabled() {
    fmt.Println("CSRF保护已启用")
}
```

## 高级功能

### 1. 自定义跳过检查规则

```go
// 设置跳过检查函数
mvc.SetCSRFSkipCheckFunc(func(userID, clientIP string) bool {
    // 对管理员用户跳过CSRF检查
    if strings.HasPrefix(userID, "admin_") {
        return true
    }
    
    // 对内网IP跳过检查
    if strings.HasPrefix(clientIP, "192.168.") {
        return true
    }
    
    return false
})
```

### 2. 中间件集成

```go
import "github.com/cloudwego/hertz/pkg/app"

// CSRF验证中间件
func CSRFMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        // 只对POST、PUT、DELETE请求验证CSRF
        method := string(ctx.Method())
        if method == "GET" || method == "HEAD" || method == "OPTIONS" {
            ctx.Next(c)
            return
        }
        
        // 获取token
        tokenValue := string(ctx.PostForm("csrf_token"))
        if tokenValue == "" {
            tokenValue = string(ctx.Request.Header.Get("X-CSRF-Token"))
        }
        
        // 获取用户信息（需要根据实际情况调整）
        userID := getUserID(ctx)
        clientIP := getClientIP(ctx)
        
        // 验证token
        isValid, err := mvc.ValidateCSRFToken(tokenValue, userID, clientIP)
        if err != nil || !isValid {
            ctx.JSON(403, map[string]string{
                "error": "CSRF token验证失败",
            })
            ctx.Abort()
            return
        }
        
        ctx.Next(c)
    }
}
```

## 最佳实践

### 1. 安全建议

- **总是验证POST/PUT/DELETE请求的CSRF token**
- **使用HTTPS来保护token传输**
- **定期轮换密钥**
- **设置合适的过期时间**

### 2. 性能优化

- **使用简单token用于模板渲染**
- **缓存配置信息**
- **避免频繁生成token**

### 3. 错误处理

```go
// 统一的CSRF错误处理
func handleCSRFError(c *Controller, err error) {
    c.Data["error"] = "安全验证失败，请刷新页面重试"
    c.TplName = "error.html"
    c.Render()
}
```

## API参考

### 全局方法

- `mvc.GenerateCSRFToken(userID, clientIP string) (*security.CSRFToken, error)`
- `mvc.ValidateCSRFToken(tokenValue, userID, clientIP string) (bool, error)`
- `mvc.GenerateSimpleCSRFToken() string`
- `mvc.GetCSRFConfig() (*security.CSRFConfig, error)`
- `mvc.UpdateCSRFConfig(config *security.CSRFConfig)`
- `mvc.SetCSRFSkipCheckFunc(fn func(userID, clientIP string) bool)`

### 便捷方法

- `mvc.IsCSRFProtectionEnabled() bool`
- `mvc.GetCSRFTokenName() string`
- `mvc.GetCSRFCookieName() string`
- `mvc.GetCSRFHeaderName() string`

### 模板函数

- `{{csrf}}` - 生成简单CSRF token
- `{{csrf_token}}` - 生成简单CSRF token（别名）

### 模板字段

- `{{.CSRF}}` - 直接访问CSRF字段
- `{{.CsrfToken}}` - 访问CsrfToken字段（推荐）
- `{{.Csrf_token}}` - 调用Csrf_token方法

## 故障排除

### 常见问题

1. **Token验证失败**
   - 检查用户ID和IP是否一致
   - 验证token是否过期
   - 确认密钥配置正确

2. **模板中token为空**
   - 确保已初始化全局管理器
   - 检查模板数据是否正确设置
   - 验证模板语法是否正确

3. **循环导入错误**
   - 本系统通过接口和适配器模式解决了循环导入问题
   - 确保按照文档示例正确导入包

## 示例项目

完整的示例可以参考测试文件：
- `csrf_integration_test.go` - 集成测试示例
- `csrf_api.go` - API使用示例

通过这些示例，你可以了解如何在实际项目中使用全局CSRF功能。