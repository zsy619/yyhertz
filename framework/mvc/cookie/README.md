# YYHertz Cookie 模块

## 概述

YYHertz Cookie 模块提供了完整、安全、高性能的Cookie管理功能。从 `@framework/mvc/session` 独立出来后，Cookie功能现在更加专注和模块化，提供了从基础操作到高级安全功能的完整解决方案。

## 🚀 主要特性

### ✨ 核心功能
- **统一的Helper接口** - 简单易用的Cookie操作API
- **Beego风格兼容** - 完全兼容Beego Cookie API
- **安全Cookie支持** - 内置HMAC-SHA256签名防篡改
- **配置驱动** - 支持配置文件和程序化配置
- **多种操作模式** - 支持简单、安全、定制化三种操作模式

### 🔐 安全特性
- **HMAC-SHA256签名** - 防止Cookie篡改
- **时间戳验证** - 防重放攻击
- **Base64编码保护** - 内容安全传输
- **HTTPS强制要求** - 生产环境安全控制
- **过期时间验证** - 防止过期Cookie使用

### ⚡ 性能特性
- **延迟初始化** - 按需创建Cookie操作器
- **零拷贝设计** - 高效的内存使用
- **配置缓存** - 避免重复配置加载
- **并发安全** - 支持高并发环境

## 📁 模块架构

```
framework/mvc/cookie/
├── helper.go        # Cookie助手 - 统一接口
├── base.go          # 基础Cookie操作 (从session迁移)
├── security.go      # 安全Cookie功能 (从session迁移)
└── README.md        # 本文档
```

## 🔧 快速开始

### 基础Cookie操作

```go
import "github.com/zsy619/yyhertz/framework/mvc/cookie"

// 创建Cookie助手
helper := cookie.NewHelper(cookie.DefaultConfig())

// 设置Cookie
helper.Set(ctx, "user_preference", "dark_mode")

// 获取Cookie
value := helper.Get(ctx, "user_preference")
fmt.Printf("用户偏好: %s\n", value)

// 删除Cookie
helper.Delete(ctx, "user_preference")

// 检查Cookie是否存在
exists := helper.Has(ctx, "user_preference")
fmt.Printf("Cookie存在: %t\n", exists)
```

### 带选项的Cookie操作

```go
// 创建自定义选项
options := &cookie.Options{
    MaxAge:   3600 * 24,    // 24小时
    Path:     "/admin",     // 仅在/admin路径下有效
    Domain:   "example.com", // 仅在example.com域下有效
    Secure:   true,         // 仅HTTPS传输
    HttpOnly: true,         // 防XSS攻击
    SameSite: "Strict",     // 严格的CSRF保护
}

// 使用自定义选项设置Cookie
helper.Set(ctx, "admin_token", "secret_value", options)
```

### 安全Cookie操作

```go
// 设置安全Cookie (Beego兼容方式)
secret := "your-hmac-secret-key-32chars!!"
helper.SetSecureCookieValue(ctx, secret, "csrf_token", "random_token_value")

// 获取并验证安全Cookie
token, valid := helper.GetSecureCookieValue(ctx, secret, "csrf_token")
if valid {
    fmt.Printf("CSRF Token: %s\n", token)
} else {
    fmt.Println("Invalid or tampered cookie")
}

// 验证安全Cookie（不返回值）
if helper.ValidateSecureCookieValue(ctx, secret, "csrf_token") {
    fmt.Println("Cookie is valid and not tampered")
}
```

### 高级安全Cookie操作

```go
import "time"

// 创建安全选项
securityOptions := cookie.CookieSecurityOptions{
    Secret:         "your-256bit-secret-key-here!!",
    MaxAge:         time.Hour * 24,    // 24小时有效期
    ValidateExpiry: true,              // 验证过期时间
    RequireHTTPS:   true,             // 要求HTTPS环境
}

// 获取安全Cookie操作器
secureCookie := helper.GetSecureCookie(ctx)

// 设置带选项的安全Cookie
err := secureCookie.SetSecureWithOptions("sensitive_data", "important_value", securityOptions)
if err != nil {
    fmt.Printf("设置安全Cookie失败: %v\n", err)
}

// 获取带验证的安全Cookie
value, valid, err := secureCookie.GetSecureWithOptions("sensitive_data", securityOptions)
if err != nil {
    fmt.Printf("获取安全Cookie出错: %v\n", err)
} else if valid {
    fmt.Printf("安全数据: %s\n", value)
} else {
    fmt.Println("Cookie无效或已过期")
}
```

## 📋 配置管理

### 默认配置

```go
config := cookie.DefaultConfig()
fmt.Printf("默认过期时间: %d秒\n", config.DefaultMaxAge)  // 3600秒
fmt.Printf("默认路径: %s\n", config.DefaultPath)          // "/"
fmt.Printf("HttpOnly: %t\n", config.HttpOnly)            // true
fmt.Printf("SameSite: %s\n", config.SameSite)           // "Lax"
```

### 从配置文件加载

```go
// 从session.yaml配置文件加载
helper := cookie.NewHelperFromConfig()

// 或者手动加载配置
config := cookie.LoadFromConfig()
helper := cookie.NewHelper(config)
```

### 程序化配置

```go
config := &cookie.Config{
    DefaultMaxAge: 7200,        // 2小时
    DefaultPath:   "/api",      // API路径
    DefaultDomain: ".example.com", // 域和子域
    DefaultSecure: true,        // HTTPS only
    HttpOnly:      true,        // 防XSS
    SameSite:      "Strict",    // 严格CSRF保护
}

helper := cookie.NewHelper(config)
```

## 🔄 Beego兼容API

为了保持与Beego的兼容性，提供了完整的兼容接口：

### 基础兼容操作

```go
// 获取BaseCookie操作器
baseCookie := helper.GetBaseCookie(ctx)

// Beego风格的基础操作
baseCookie.Set("username", "john", 3600, "/", "", false, true)
username := baseCookie.Get("username")
exists := baseCookie.Exists("username")
baseCookie.Delete("username")

// 或者通过Helper的兼容方法
helper.SetBeegoStyle(ctx, "username", "john", 3600, "/", "", false, true)
username := helper.GetBeegoStyle(ctx, "username")
helper.DeleteBeegoStyle(ctx, "username")
```

### 安全Cookie兼容

```go
// 获取SecureCookie操作器
secureCookie := helper.GetSecureCookie(ctx)

// Beego风格的安全Cookie
secret := "your-secret-key"
secureCookie.SetSecure(secret, "secure_data", "sensitive_value", 3600, "/")
value, valid := secureCookie.GetSecure(secret, "secure_data")
isValid := secureCookie.Validate(secret, "secure_data")
```

## 🛠️ 工具函数

### Cookie字符串处理

```go
// 解析Cookie字符串
cookieStr := "name1=value1; name2=value2; name3=value3"
parsed := cookie.ParseCookieString(cookieStr)
fmt.Printf("解析出%d个cookie\n", len(parsed))

// 格式化Cookie映射为字符串
cookies := map[string]string{
    "session": "abc123",
    "theme":   "dark",
    "lang":    "zh-CN",
}
formatted := cookie.FormatCookieString(cookies)
fmt.Printf("格式化后: %s\n", formatted)
```

### Cookie验证

```go
// 验证Cookie名称
validNames := []string{"user_id", "session_token", "preference"}
invalidNames := []string{"invalid;name", "bad=name", "含中文"}

for _, name := range validNames {
    fmt.Printf("'%s' 有效: %t\n", name, cookie.ValidateCookieName(name))
}

for _, name := range invalidNames {
    fmt.Printf("'%s' 有效: %t\n", name, cookie.ValidateCookieName(name))
}

// 验证Cookie值
values := []string{"simple_value", "with spaces", "with\ncontrol"}
for _, value := range values {
    fmt.Printf("'%s' 有效: %t\n", value, cookie.ValidateCookieValue(value))
}
```

### 安全工具

```go
// 生成安全Cookie值
secret := "your-secret-key"
plainValue := "sensitive_data"
secureValue := cookie.GenerateSecureValue(secret, plainValue)
fmt.Printf("安全值: %s\n", secureValue)

// 解析安全Cookie值
parsedValue, valid := cookie.ParseSecureValue(secret, secureValue)
fmt.Printf("解析值: %s, 有效: %t\n", parsedValue, valid)

// 检查安全Cookie是否过期
isExpired := cookie.IsSecureCookieExpired(secureValue, time.Hour*24)
fmt.Printf("是否过期: %t\n", isExpired)

// 获取安全Cookie的时间戳
timestamp, err := cookie.GetSecureCookieTimestamp(secureValue)
if err == nil {
    fmt.Printf("创建时间: %s\n", timestamp.Format("2006-01-02 15:04:05"))
}
```

## 🔄 从Session包迁移

### 无缝迁移

如果你之前在session包中使用Cookie功能，现在可以无缝迁移：

```go
// 旧方式 (在session包中，仍然有效)
import "github.com/zsy619/yyhertz/framework/mvc/session"
extension := session.NewExtensionForHertzContext(ctx)
extension.SetCookie("key", "value")

// 新方式 (推荐，在cookie包中)
import "github.com/zsy619/yyhertz/framework/mvc/cookie"
helper := cookie.NewHelper(cookie.DefaultConfig())
helper.Set(ctx, "key", "value")
```

### 类型迁移

所有Cookie相关类型已迁移到cookie包：

- `session.BaseCookie` → `cookie.BaseCookie`
- `session.SecureCookie` → `cookie.SecureCookie`
- `session.OutputCookie` → `cookie.OutputCookie`
- `session.CookieSecurityOptions` → `cookie.CookieSecurityOptions`

### 函数迁移

所有Cookie相关函数已迁移：

- `session.NewBaseCookie()` → `cookie.NewBaseCookie()`
- `session.ParseCookieString()` → `cookie.ParseCookieString()`
- `session.GenerateSecureValue()` → `cookie.GenerateSecureValue()`
- 等等...

## 🧪 测试和验证

### 基本功能测试

```go
func TestCookieBasicOperations() {
    helper := cookie.NewHelper(cookie.DefaultConfig())
    ctx := // your RequestContext
    
    // 测试设置和获取
    helper.Set(ctx, "test", "value")
    value := helper.Get(ctx, "test")
    assert.Equal(t, "value", value)
    
    // 测试存在性检查
    exists := helper.Has(ctx, "test")
    assert.True(t, exists)
    
    // 测试删除
    helper.Delete(ctx, "test")
    exists = helper.Has(ctx, "test")
    assert.False(t, exists)
}
```

### 安全功能测试

```go
func TestSecureCookie() {
    helper := cookie.NewHelper(cookie.DefaultConfig())
    ctx := // your RequestContext
    secret := "test-secret-key"
    
    // 测试安全Cookie
    helper.SetSecureCookieValue(ctx, secret, "secure_test", "secure_value")
    value, valid := helper.GetSecureCookieValue(ctx, secret, "secure_test")
    
    assert.True(t, valid)
    assert.Equal(t, "secure_value", value)
    
    // 测试篡改检测
    // 篡改Cookie后应该验证失败
    valid = helper.ValidateSecureCookieValue(ctx, "wrong-secret", "secure_test")
    assert.False(t, valid)
}
```

## 📈 性能考量

### 最佳实践

1. **复用Helper实例**：避免重复创建Helper
2. **配置缓存**：使用`NewHelperFromConfig()`缓存配置
3. **延迟初始化**：BaseCookie和SecureCookie按需创建
4. **合理过期时间**：避免设置过长的过期时间

### 性能测试

```bash
# 运行基准测试
go test -bench=BenchmarkCookie ./framework/mvc/cookie
go test -bench=BenchmarkSecureCookie ./framework/mvc/cookie
```

## 🔒 安全建议

### 生产环境安全

1. **使用HTTPS**：生产环境务必启用HTTPS
2. **强密钥**：安全Cookie使用至少32字符的随机密钥
3. **合理过期**：根据业务需求设置合理的过期时间
4. **域限制**：限制Cookie的域和路径范围
5. **HttpOnly**：敏感Cookie启用HttpOnly防XSS
6. **SameSite**：根据需求设置适当的SameSite策略

### 密钥管理

```go
// 好的做法：从环境变量或配置文件获取密钥
secret := os.Getenv("COOKIE_SECRET")
if secret == "" {
    log.Fatal("COOKIE_SECRET环境变量未设置")
}

// 坏的做法：硬编码密钥
// secret := "hardcoded-secret"  // 不要这样做！
```

## 🎯 使用场景

### 1. 用户认证

```go
// 设置认证Cookie
authToken := generateAuthToken(userID)
helper.SetSecure(ctx, "auth_token", authToken, 3600*24*7) // 7天有效

// 验证认证Cookie
token := helper.Get(ctx, "auth_token")
if isValidAuthToken(token) {
    // 用户已认证
}
```

### 2. 用户偏好

```go
// 保存用户偏好
preferences := map[string]string{
    "theme": "dark",
    "lang":  "zh-CN",
    "timezone": "Asia/Shanghai",
}

for key, value := range preferences {
    helper.Set(ctx, "pref_"+key, value, &cookie.Options{
        MaxAge: 3600 * 24 * 365, // 1年
        Path:   "/",
    })
}
```

### 3. CSRF保护

```go
// 生成CSRF Token
csrfToken := generateCSRFToken()
helper.SetSecureCookieValue(ctx, secret, "csrf_token", csrfToken)

// 验证CSRF Token
token, valid := helper.GetSecureCookieValue(ctx, secret, "csrf_token")
if !valid || token != requestCSRFToken {
    return errors.New("CSRF token validation failed")
}
```

---

## 📖 更多资源

- **Session模块**: `@framework/mvc/session` - 会话管理功能
- **Context模块**: `@framework/mvc/context` - 上下文和兼容性接口  
- **全局架构**: 查看项目根目录的架构文档

**版本**: v1.0.0  
**更新时间**: 2025年8月19日  
**状态**: ✅ 生产就绪