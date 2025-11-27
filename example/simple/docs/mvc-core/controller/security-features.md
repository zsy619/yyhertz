# 🔒 安全特性详解

YYHertz控制器内置了多层安全防护机制，包括XSRF保护、安全Cookie、输入验证等功能。

## 🛡️ XSRF/CSRF保护

### 核心安全方法 (6个)

| 方法 | 说明 | 示例 |
|------|------|------|
| `XSRFToken()` | 生成XSRF令牌 | `token := c.XSRFToken()` |
| `CheckXSRFCookie()` | 检查XSRF令牌 | `valid := c.CheckXSRFCookie()` |
| `EnableXSRF(expire...)` | 启用XSRF保护 | `c.EnableXSRF(3600)` |
| `DisableXSRF()` | 禁用XSRF保护 | `c.DisableXSRF()` |
| `SetSecureCookie(...)` | 设置安全Cookie | `c.SetSecureCookie(key, name, value)` |
| `GetSecureCookie(...)` | 获取安全Cookie | `value, ok := c.GetSecureCookie(key, name)` |

### XSRF防护实现

```go
func (c *BaseController) enableXSRFProtection() {
    // 🛡️ 为POST/PUT/DELETE请求启用XSRF保护
    if c.needsXSRFProtection() {
        c.EnableXSRF(3600) // 1小时有效期
        
        if !c.CheckXSRFCookie() {
            c.LogWarn("XSRF令牌验证失败", map[string]any{
                "path":   c.Ctx.Path(),
                "method": c.Ctx.Method(),
                "ip":     c.GetClientIP(),
            })
            
            if c.IsAjax() {
                c.JSONForbidden("CSRF令牌验证失败")
            } else {
                c.Error(403, "请求被拒绝")
            }
            c.StopRun()
            return
        }
    }
    
    // 📝 为所有页面提供XSRF令牌
    if !c.IsAjax() {
        xsrfToken := c.XSRFToken()
        c.SetData("XSRFToken", xsrfToken)
        c.SetData("CSRFToken", xsrfToken) // 兼容性
    }
}

func (c *BaseController) needsXSRFProtection() bool {
    method := c.Ctx.Method()
    return method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"
}
```

### 前端XSRF令牌使用

```html
<!-- HTML表单中使用XSRF令牌 -->
<form method="POST" action="/users/create">
    <input type="hidden" name="_xsrf" value="{{.XSRFToken}}">
    <input type="text" name="name" placeholder="用户名">
    <button type="submit">创建用户</button>
</form>

<!-- Ajax请求中使用XSRF令牌 -->
<script>
function createUser(userData) {
    $.ajaxSetup({
        beforeSend: function(xhr) {
            xhr.setRequestHeader('X-Xsrftoken', '{{.XSRFToken}}');
            // 或者使用其他格式的头
            xhr.setRequestHeader('X-CSRF-Token', '{{.XSRFToken}}');
        }
    });
    
    $.post('/api/users', userData)
        .done(function(response) {
            console.log('用户创建成功');
        })
        .fail(function(xhr) {
            if (xhr.status === 403) {
                alert('安全验证失败，请刷新页面重试');
                location.reload();
            }
        });
}
</script>
```

## 🔐 输入安全验证

### 数据验证和过滤

```go
type UserCreateRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=50,alpha"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*"`
    Phone    string `json:"phone" validate:"omitempty,phone"`
    Website  string `json:"website" validate:"omitempty,url"`
    Bio      string `json:"bio" validate:"max=500"`
}

func (c *UserController) PostCreate() {
    var req UserCreateRequest
    
    // 📝 绑定和验证数据
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("JSON格式错误: " + err.Error())
        return
    }
    
    if err := c.ValidateStruct(&req); err != nil {
        c.JSONBadRequest("数据验证失败: " + err.Error())
        return
    }
    
    // 🛡️ 额外安全检查
    if err := c.securityValidation(&req); err != nil {
        c.JSONBadRequest("安全检查失败: " + err.Error())
        return
    }
    
    // 🧹 数据清理
    req.Name = c.sanitizeString(req.Name)
    req.Bio = c.sanitizeHTML(req.Bio)
    
    // 💾 创建用户
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        c.JSONInternalServerError("创建失败")
        return
    }
    
    c.JSONSuccess("创建成功", user)
}

// 安全验证
func (c *BaseController) securityValidation(req *UserCreateRequest) error {
    // 🚫 检查恶意字符
    maliciousPatterns := []string{
        "<script", "javascript:", "on[a-z]+\\s*=", // XSS
        "union.*select", "drop.*table", "exec.*sp_", // SQL注入
        "../", "..\\\\", // 路径遍历
    }
    
    content := fmt.Sprintf("%s %s %s", req.Name, req.Email, req.Bio)
    for _, pattern := range maliciousPatterns {
        if matched, _ := regexp.MatchString("(?i)"+pattern, content); matched {
            return fmt.Errorf("检测到潜在的恶意输入")
        }
    }
    
    return nil
}

// 字符串清理
func (c *BaseController) sanitizeString(input string) string {
    // 移除控制字符
    input = strings.Map(func(r rune) rune {
        if r < 32 && r != '\n' && r != '\r' && r != '\t' {
            return -1
        }
        return r
    }, input)
    
    // 限制长度
    if len(input) > 200 {
        input = input[:200]
    }
    
    return strings.TrimSpace(input)
}
```

## 🔒 访问控制

### 权限验证系统

```go
type Permission struct {
    Resource string `json:"resource"`
    Action   string `json:"action"`
    UserID   int    `json:"user_id"`
}

func (c *BaseController) CheckPermission(resource, action string) bool {
    userID := c.GetCurrentUserID()
    if userID == 0 {
        return false
    }
    
    // 🔍 检查超级管理员
    if c.IsSuperAdmin(userID) {
        return true
    }
    
    // 🔍 检查具体权限
    permission := Permission{
        Resource: resource,
        Action:   action,
        UserID:   userID,
    }
    
    return c.permissionService.HasPermission(permission)
}

// 权限中间件
func (c *BaseController) RequirePermission(resource, action string) {
    if !c.CheckPermission(resource, action) {
        c.LogWarn("权限检查失败", map[string]any{
            "user_id":  c.GetCurrentUserID(),
            "resource": resource,
            "action":   action,
            "path":     c.Ctx.Path(),
        })
        
        if c.IsAjax() {
            c.JSONForbidden("权限不足")
        } else {
            c.Error(403, "访问被拒绝")
        }
        c.StopRun()
        return
    }
}

// 使用示例
func (c *AdminController) DeleteUser() {
    c.RequirePermission("users", "delete")
    
    userID := c.GetParam("id")
    // ... 删除逻辑
}
```

## 🌐 请求限流

### API限流实现

```go
type RateLimit struct {
    requests map[string][]time.Time
    mutex    sync.RWMutex
    limit    int
    window   time.Duration
}

func (c *BaseController) checkRateLimit() bool {
    // 🆔 生成客户端标识
    clientID := c.getClientIdentifier()
    
    // 🔍 检查请求频率
    if !c.rateLimiter.Allow(clientID) {
        c.LogWarn("请求频率超限", map[string]any{
            "client_id": clientID,
            "path":      c.Ctx.Path(),
            "ip":        c.GetClientIP(),
        })
        
        c.Ctx.Header("Retry-After", "60")
        c.JSONError("请求过于频繁，请稍后再试", nil, 429)
        c.StopRun()
        return false
    }
    
    return true
}

func (c *BaseController) getClientIdentifier() string {
    // 🔑 优先使用用户ID
    if userID := c.GetCurrentUserID(); userID > 0 {
        return fmt.Sprintf("user:%d", userID)
    }
    
    // 🌐 使用IP地址
    ip := c.GetClientIP()
    
    // 🔍 考虑User-Agent防止同IP不同客户端
    userAgent := c.GetUserAgent()
    hash := c.hashString(ip + userAgent)
    
    return fmt.Sprintf("ip:%s", hash[:8])
}
```

## 🔍 安全日志记录

### 安全事件监控

```go
type SecurityEvent struct {
    Type        string    `json:"type"`
    Level       string    `json:"level"`
    UserID      int       `json:"user_id,omitempty"`
    IP          string    `json:"ip"`
    UserAgent   string    `json:"user_agent"`
    Path        string    `json:"path"`
    Method      string    `json:"method"`
    Description string    `json:"description"`
    Timestamp   time.Time `json:"timestamp"`
    Extra       map[string]any `json:"extra,omitempty"`
}

func (c *BaseController) logSecurityEvent(eventType, level, description string, extra map[string]any) {
    event := SecurityEvent{
        Type:        eventType,
        Level:       level,
        UserID:      c.GetCurrentUserID(),
        IP:          c.GetClientIP(),
        UserAgent:   c.GetUserAgent(),
        Path:        c.Ctx.Path(),
        Method:      c.Ctx.Method(),
        Description: description,
        Timestamp:   time.Now(),
        Extra:       extra,
    }
    
    // 📝 记录到安全日志
    c.securityLogger.Log(event)
    
    // 🚨 高风险事件实时告警
    if level == "high" || level == "critical" {
        c.alertManager.SendAlert(event)
    }
}

// 常见安全事件记录
func (c *BaseController) recordLoginAttempt(success bool, username string) {
    eventType := "login_success"
    level := "info"
    description := fmt.Sprintf("用户 %s 登录成功", username)
    
    if !success {
        eventType = "login_failure"
        level = "warning"
        description = fmt.Sprintf("用户 %s 登录失败", username)
    }
    
    c.logSecurityEvent(eventType, level, description, map[string]any{
        "username": username,
        "success":  success,
    })
}
```

## 🔧 安全配置

### 安全头设置

```go
func (c *BaseController) setSecurityHeaders() {
    // 🛡️ XSS保护
    c.Ctx.Header("X-XSS-Protection", "1; mode=block")
    
    // 🔒 内容类型嗅探保护
    c.Ctx.Header("X-Content-Type-Options", "nosniff")
    
    // 🖼️ 点击劫持保护
    c.Ctx.Header("X-Frame-Options", "DENY")
    
    // 🔐 HTTPS强制
    if c.app.Config.Environment == "production" {
        c.Ctx.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    }
    
    // 🌐 内容安全策略
    csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
    c.Ctx.Header("Content-Security-Policy", csp)
    
    // 🔍 引用策略
    c.Ctx.Header("Referrer-Policy", "strict-origin-when-cross-origin")
}
```

## ❓ 安全最佳实践

**Q: 如何防止SQL注入攻击？**
A: 使用参数化查询、ORM映射、输入验证等多层防护。

**Q: 密码如何安全存储？**
A: 使用bcrypt、scrypt等强哈希算法，加盐存储。

**Q: 文件上传安全如何保证？**
A: 检查文件类型、大小、重命名文件、隔离存储。

**Q: API接口如何防止暴力破解？**
A: 实施请求限流、验证码验证、账户锁定策略。