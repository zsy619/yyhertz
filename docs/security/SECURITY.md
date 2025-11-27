# YYHertz 安全配置指南

<div align="center">

🔒 **企业级安全配置** | 保护你的应用免受攻击

</div>

---

## 🛡️ 安全架构概览

### 多层防护体系
```
┌─────────────────── 网络层安全 ───────────────────┐
│  HTTPS/TLS  │  防火墙  │  DDoS防护  │  CDN     │
├─────────────────── 应用层安全 ───────────────────┤
│  CSRF保护   │  XSS防护  │  SQL注入防护│  认证   │
├─────────────────── 数据层安全 ───────────────────┤
│  数据加密   │  访问控制  │  审计日志   │  备份   │
└─────────────────── 基础设施安全 ─────────────────┘
```

---

## 🔐 身份认证与授权

### 1. JWT认证配置

```go
// JWT中间件配置
import "github.com/zsy619/yyhertz/framework/mvc/middleware"

func setupJWTAuth(app *mvc.Application) {
    jwtConfig := middleware.JWTConfig{
        SigningKey:    []byte(config.GetString("jwt.secret")),
        TokenLookup:   "header:Authorization,query:token,cookie:jwt",
        AuthScheme:    "Bearer",
        Expiration:    time.Hour * 24,
        RefreshTime:   time.Hour * 2,
        Issuer:        "YYHertz-App",
        Subject:       "user-auth",
        
        // 自定义密钥获取函数
        KeyFunc: func(t *jwt.Token) (interface{}, error) {
            if t.Method.Alg() != "HS256" {
                return nil, fmt.Errorf("unexpected jwt signing method: %v", t.Header["alg"])
            }
            return []byte(config.GetString("jwt.secret")), nil
        },
        
        // 跳过认证的路径
        Skipper: func(c *mvc.Context) bool {
            skipPaths := []string{"/login", "/register", "/health", "/public"}
            for _, path := range skipPaths {
                if strings.HasPrefix(c.Path(), path) {
                    return true
                }
            }
            return false
        },
        
        // 认证成功回调
        SuccessHandler: func(c *mvc.Context) {
            userID := c.Get("user_id").(string)
            // 记录登录日志
            logger.Info("User authenticated", "user_id", userID, "ip", c.ClientIP())
        },
        
        // 认证失败回调
        ErrorHandler: func(c *mvc.Context, err error) {
            logger.Warn("Authentication failed", "error", err.Error(), "ip", c.ClientIP())
            c.JSON(401, map[string]interface{}{
                "error":   "Unauthorized",
                "message": "Invalid or expired token",
                "code":    4001,
            })
        },
    }
    
    app.Use(middleware.JWTAuth(jwtConfig))
}
```

### 2. 多重认证支持

```go
// 多重认证策略
type AuthService struct {
    jwtService    *JWTService
    sessionService *SessionService
    apiKeyService  *APIKeyService
}

func (s *AuthService) AuthenticateRequest(c *mvc.Context) (*User, error) {
    // 1. 尝试JWT认证
    if token := extractJWTToken(c); token != "" {
        if user, err := s.jwtService.ValidateToken(token); err == nil {
            return user, nil
        }
    }
    
    // 2. 尝试会话认证
    if sessionID := extractSessionID(c); sessionID != "" {
        if user, err := s.sessionService.GetUser(sessionID); err == nil {
            return user, nil
        }
    }
    
    // 3. 尝试API Key认证
    if apiKey := extractAPIKey(c); apiKey != "" {
        if user, err := s.apiKeyService.ValidateAPIKey(apiKey); err == nil {
            return user, nil
        }
    }
    
    return nil, errors.New("authentication required")
}
```

### 3. 基于角色的访问控制 (RBAC)

```go
// 权限检查中间件
func RequirePermission(permission string) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        user := getCurrentUser(c)
        if user == nil {
            c.AbortWithJSON(401, map[string]string{"error": "未授权访问"})
            return
        }
        
        if !user.HasPermission(permission) {
            c.AbortWithJSON(403, map[string]string{"error": "权限不足"})
            return
        }
        
        c.Next()
    }
}

// 角色模型
type User struct {
    ID       uint     `json:"id"`
    Username string   `json:"username"`
    Roles    []Role   `json:"roles" gorm:"many2many:user_roles;"`
}

type Role struct {
    ID          uint         `json:"id"`
    Name        string       `json:"name"`
    Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

type Permission struct {
    ID       uint   `json:"id"`
    Name     string `json:"name"`
    Resource string `json:"resource"`
    Action   string `json:"action"`
}

// 权限检查方法
func (u *User) HasPermission(permissionName string) bool {
    for _, role := range u.Roles {
        for _, perm := range role.Permissions {
            if perm.Name == permissionName {
                return true
            }
        }
    }
    return false
}

// 使用权限中间件
app.GET("/admin/users", RequirePermission("user.list"), getUserList)
app.POST("/admin/users", RequirePermission("user.create"), createUser)
app.DELETE("/admin/users/:id", RequirePermission("user.delete"), deleteUser)
```

---

## 🚫 CSRF防护

### 1. CSRF中间件配置

```go
import "github.com/zsy619/yyhertz/framework/mvc/middleware"

func setupCSRFProtection(app *mvc.Application) {
    csrfConfig := middleware.CSRFConfig{
        TokenLookup:    "form:csrf_token,header:X-CSRF-Token,header:X-Requested-With",
        TokenLength:    32,
        TokenGenerator: generateSecureToken,
        CookieName:     "_csrf",
        CookieDomain:   config.GetString("app.domain"),
        CookieSecure:   config.GetBool("app.secure"),
        CookieHTTPOnly: true,
        CookieSameSite: http.SameSiteStrictMode,
        Expiration:     time.Hour * 12,
        
        // 跳过CSRF检查的路径
        Skipper: func(c *mvc.Context) bool {
            // GET、HEAD、OPTIONS请求跳过
            if c.Method() == "GET" || c.Method() == "HEAD" || c.Method() == "OPTIONS" {
                return true
            }
            
            // API路径跳过（使用其他认证方式）
            if strings.HasPrefix(c.Path(), "/api/") {
                return true
            }
            
            return false
        },
        
        ErrorHandler: func(c *mvc.Context, err error) {
            logger.Warn("CSRF token validation failed", 
                "ip", c.ClientIP(), 
                "path", c.Path(), 
                "error", err.Error())
                
            c.JSON(403, map[string]interface{}{
                "error":   "Forbidden", 
                "message": "CSRF token validation failed",
                "code":    4003,
            })
        },
    }
    
    app.Use(middleware.CSRF(csrfConfig))
}

// 安全令牌生成器
func generateSecureToken(length int) string {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        panic("failed to generate csrf token")
    }
    return hex.EncodeToString(bytes)
}
```

### 2. 表单CSRF保护

```html
<!-- 在表单中包含CSRF令牌 -->
<form method="POST" action="/users">
    <input type="hidden" name="csrf_token" value="{{csrf_token}}">
    
    <div>
        <label>用户名：</label>
        <input type="text" name="username" required>
    </div>
    
    <div>
        <label>邮箱：</label>
        <input type="email" name="email" required>
    </div>
    
    <button type="submit">创建用户</button>
</form>
```

### 3. AJAX请求CSRF保护

```javascript
// 获取CSRF令牌
function getCSRFToken() {
    return document.querySelector('meta[name="csrf-token"]').getAttribute('content');
}

// AJAX请求包含CSRF令牌
$.ajaxSetup({
    beforeSend: function(xhr, settings) {
        if (!/^(GET|HEAD|OPTIONS|TRACE)$/i.test(settings.type) && !this.crossDomain) {
            xhr.setRequestHeader("X-CSRF-Token", getCSRFToken());
        }
    }
});

// 或者使用fetch
fetch('/api/users', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': getCSRFToken()
    },
    body: JSON.stringify(userData)
});
```

---

## 🛡️ XSS防护

### 1. 输出转义

```go
// 自动HTML转义
func (c *BaseController) SafeHTML(content string) template.HTML {
    // 使用白名单策略
    p := bluemonday.UGCPolicy()
    cleaned := p.Sanitize(content)
    return template.HTML(cleaned)
}

// 模板中使用
func (c *PostController) GetDetail() {
    post := getPostByID(c.ParamInt("id"))
    
    c.HTML("post/detail.html", map[string]interface{}{
        "post": post,
        "content": c.SafeHTML(post.Content), // 安全的HTML输出
    })
}
```

### 2. 内容安全策略 (CSP)

```go
// CSP中间件
func CSPMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        csp := []string{
            "default-src 'self'",
            "script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net",
            "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
            "img-src 'self' data: https:",
            "font-src 'self' https://fonts.gstatic.com",
            "connect-src 'self'",
            "frame-ancestors 'none'",
            "base-uri 'self'",
            "form-action 'self'",
        }
        
        c.Header("Content-Security-Policy", strings.Join(csp, "; "))
        c.Next()
    }
}

app.Use(CSPMiddleware())
```

### 3. 输入验证和过滤

```go
import "github.com/go-playground/validator/v10"

// 输入验证结构体
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Bio      string `json:"bio" validate:"max=500"`
}

// 验证中间件
func ValidateInput() mvc.HandlerFunc {
    validate := validator.New()
    
    return func(c *mvc.Context) {
        var req CreateUserRequest
        
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, map[string]interface{}{
                "error": "Invalid JSON format",
                "details": err.Error(),
            })
            c.Abort()
            return
        }
        
        if err := validate.Struct(&req); err != nil {
            errors := make(map[string]string)
            for _, err := range err.(validator.ValidationErrors) {
                errors[err.Field()] = getValidationErrorMsg(err)
            }
            
            c.JSON(422, map[string]interface{}{
                "error": "Validation failed",
                "details": errors,
            })
            c.Abort()
            return
        }
        
        // 额外的安全过滤
        req.Username = sanitizeInput(req.Username)
        req.Bio = sanitizeInput(req.Bio)
        
        c.Set("validated_request", req)
        c.Next()
    }
}

// 输入清理函数
func sanitizeInput(input string) string {
    // 移除HTML标签
    re := regexp.MustCompile(`<[^>]*>`)
    cleaned := re.ReplaceAllString(input, "")
    
    // 移除危险字符
    cleaned = strings.ReplaceAll(cleaned, "<script", "")
    cleaned = strings.ReplaceAll(cleaned, "</script>", "")
    cleaned = strings.ReplaceAll(cleaned, "javascript:", "")
    
    return strings.TrimSpace(cleaned)
}
```

---

## 🗂️ SQL注入防护

### 1. 参数化查询

```go
// ❌ 错误做法：字符串拼接
func getUserByName(name string) (*User, error) {
    query := "SELECT * FROM users WHERE name = '" + name + "'"
    // 容易受到SQL注入攻击
}

// ✅ 正确做法：参数化查询
func getUserByName(name string) (*User, error) {
    var user User
    err := db.Where("name = ?", name).First(&user).Error
    return &user, err
}

// ✅ 原生SQL的安全写法
func getUserStats(userID int, startDate, endDate time.Time) ([]UserStat, error) {
    var stats []UserStat
    err := db.Raw(`
        SELECT u.id, u.name, COUNT(p.id) as post_count, AVG(p.rating) as avg_rating
        FROM users u 
        LEFT JOIN posts p ON u.id = p.user_id 
        WHERE u.id = ? AND p.created_at BETWEEN ? AND ?
        GROUP BY u.id, u.name
    `, userID, startDate, endDate).Scan(&stats).Error
    
    return stats, err
}
```

### 2. 输入验证和类型检查

```go
// 严格的数据类型验证
func (c *UserController) GetDetail() {
    // 验证ID参数
    idStr := c.Param("id")
    if idStr == "" {
        c.ErrorJSON(400, "Missing user ID")
        return
    }
    
    id, err := strconv.Atoi(idStr)
    if err != nil || id <= 0 {
        c.ErrorJSON(400, "Invalid user ID format")
        return
    }
    
    // 使用验证后的ID查询
    user, err := getUserByID(id)
    if err != nil {
        c.ErrorJSON(404, "User not found")
        return
    }
    
    c.JSON(user)
}

// 查询条件验证
func (c *UserController) GetList() {
    // 验证分页参数
    page := c.QueryInt("page")
    if page <= 0 {
        page = 1
    }
    if page > 1000 { // 防止过大的页码
        page = 1000
    }
    
    size := c.QueryInt("size")
    if size <= 0 || size > 100 { // 限制每页数量
        size = 20
    }
    
    // 验证排序参数
    sortField := c.Query("sort")
    allowedSorts := map[string]bool{
        "id": true, "name": true, "created_at": true, "updated_at": true,
    }
    if !allowedSorts[sortField] {
        sortField = "id"
    }
    
    users, total, err := getUserList(page, size, sortField)
    if err != nil {
        c.ErrorJSON(500, "Query failed")
        return
    }
    
    c.JSON(map[string]interface{}{
        "users": users,
        "total": total,
        "page":  page,
        "size":  size,
    })
}
```

---

## 🚦 安全头配置

### 1. 安全头中间件

```go
func SecurityHeadersMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 防止点击劫持
        c.Header("X-Frame-Options", "DENY")
        
        // 防止MIME类型嗅探
        c.Header("X-Content-Type-Options", "nosniff")
        
        // XSS防护
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // 强制HTTPS
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        // 推荐策略
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // 权限策略
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        
        c.Next()
    }
}

app.Use(SecurityHeadersMiddleware())
```

### 2. HTTPS重定向

```go
func HTTPSRedirectMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        if !c.Request.TLS && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
            if config.GetBool("app.force_https") {
                httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
                c.Redirect(301, httpsURL)
                c.Abort()
                return
            }
        }
        c.Next()
    }
}
```

---

## 🔍 安全日志和监控

### 1. 安全事件日志

```go
import "github.com/sirupsen/logrus"

// 安全日志器
type SecurityLogger struct {
    logger *logrus.Logger
}

func NewSecurityLogger() *SecurityLogger {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    
    // 安全日志单独存储
    file, err := os.OpenFile("logs/security.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        panic(err)
    }
    logger.SetOutput(file)
    
    return &SecurityLogger{logger: logger}
}

func (s *SecurityLogger) LogAuthFailure(ip, username, reason string) {
    s.logger.WithFields(logrus.Fields{
        "event":     "auth_failure",
        "ip":        ip,
        "username":  username,
        "reason":    reason,
        "timestamp": time.Now().Unix(),
    }).Warn("Authentication failed")
}

func (s *SecurityLogger) LogSuspiciousActivity(ip, path, reason string) {
    s.logger.WithFields(logrus.Fields{
        "event":     "suspicious_activity",
        "ip":        ip,
        "path":      path,
        "reason":    reason,
        "timestamp": time.Now().Unix(),
    }).Error("Suspicious activity detected")
}

// 使用安全日志
var securityLogger = NewSecurityLogger()

func authFailureHandler(c *mvc.Context, err error) {
    username := c.PostForm("username")
    securityLogger.LogAuthFailure(c.ClientIP(), username, err.Error())
    
    c.JSON(401, map[string]string{"error": "Authentication failed"})
}
```

### 2. 入侵检测

```go
// 简单的入侵检测系统
type IntrusionDetector struct {
    failedAttempts map[string]int
    blockedIPs     map[string]time.Time
    mutex          sync.RWMutex
}

func NewIntrusionDetector() *IntrusionDetector {
    id := &IntrusionDetector{
        failedAttempts: make(map[string]int),
        blockedIPs:     make(map[string]time.Time),
    }
    
    // 定期清理过期数据
    go id.cleanup()
    return id
}

func (id *IntrusionDetector) RecordFailedAttempt(ip string) {
    id.mutex.Lock()
    defer id.mutex.Unlock()
    
    id.failedAttempts[ip]++
    
    // 超过阈值则封禁
    if id.failedAttempts[ip] >= 5 {
        id.blockedIPs[ip] = time.Now().Add(time.Hour) // 封禁1小时
        delete(id.failedAttempts, ip)
        
        securityLogger.LogSuspiciousActivity(ip, "", "Too many failed attempts")
    }
}

func (id *IntrusionDetector) IsBlocked(ip string) bool {
    id.mutex.RLock()
    defer id.mutex.RUnlock()
    
    if blockedUntil, exists := id.blockedIPs[ip]; exists {
        if time.Now().Before(blockedUntil) {
            return true
        }
        // 过期则删除
        delete(id.blockedIPs, ip)
    }
    return false
}

func (id *IntrusionDetector) cleanup() {
    ticker := time.NewTicker(10 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        id.mutex.Lock()
        now := time.Now()
        
        // 清理过期的封禁
        for ip, blockedUntil := range id.blockedIPs {
            if now.After(blockedUntil) {
                delete(id.blockedIPs, ip)
            }
        }
        
        // 清理失败尝试记录（1小时后重置）
        for ip := range id.failedAttempts {
            if rand.Float32() < 0.1 { // 随机清理，避免内存泄漏
                delete(id.failedAttempts, ip)
            }
        }
        
        id.mutex.Unlock()
    }
}

// 入侵检测中间件
var detector = NewIntrusionDetector()

func IntrusionDetectionMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        ip := c.ClientIP()
        
        if detector.IsBlocked(ip) {
            c.JSON(429, map[string]string{
                "error": "IP temporarily blocked due to suspicious activity",
            })
            c.Abort()
            return
        }
        
        c.Next()
        
        // 记录失败的认证尝试
        if c.Writer.Status() == 401 {
            detector.RecordFailedAttempt(ip)
        }
    }
}
```

---

## 🔐 数据加密

### 1. 敏感数据加密

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

// 数据加密服务
type EncryptionService struct {
    key []byte
}

func NewEncryptionService(key []byte) *EncryptionService {
    if len(key) != 32 {
        panic("encryption key must be 32 bytes")
    }
    return &EncryptionService{key: key}
}

func (e *EncryptionService) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (e *EncryptionService) Decrypt(ciphertext string) (string, error) {
    data, err := base64.URLEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }
    
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}

// 敏感数据模型
type User struct {
    ID       uint   `gorm:"primarykey"`
    Username string
    Email    string
    Phone    string `gorm:"-"` // 不直接存储
    PhoneEnc string `gorm:"column:phone_encrypted"` // 加密存储
}

// 加密敏感字段
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Phone != "" {
        encrypted, err := encryptionService.Encrypt(u.Phone)
        if err != nil {
            return err
        }
        u.PhoneEnc = encrypted
        u.Phone = "" // 清空明文
    }
    return nil
}

// 解密敏感字段
func (u *User) AfterFind(tx *gorm.DB) error {
    if u.PhoneEnc != "" {
        decrypted, err := encryptionService.Decrypt(u.PhoneEnc)
        if err != nil {
            return err
        }
        u.Phone = decrypted
    }
    return nil
}
```

### 2. 密码哈希

```go
import "golang.org/x/crypto/bcrypt"

// 密码服务
type PasswordService struct{}

func NewPasswordService() *PasswordService {
    return &PasswordService{}
}

func (p *PasswordService) HashPassword(password string) (string, error) {
    // 使用高强度的cost值
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}

func (p *PasswordService) CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// 密码强度验证
func (p *PasswordService) ValidatePasswordStrength(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    var hasUpper, hasLower, hasNumber, hasSpecial bool
    
    for _, char := range password {
        switch {
        case 'A' <= char && char <= 'Z':
            hasUpper = true
        case 'a' <= char && char <= 'z':
            hasLower = true
        case '0' <= char && char <= '9':
            hasNumber = true
        default:
            hasSpecial = true
        }
    }
    
    if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
        return errors.New("password must contain uppercase, lowercase, number and special characters")
    }
    
    return nil
}
```

---

## ⚙️ 安全配置文件

### 1. 生产环境配置

```ini
# config/security.conf

# 基础安全设置
security.force_https = true
security.hsts_max_age = 31536000
security.frame_options = DENY
security.content_type_options = nosniff

# CSRF保护
csrf.enabled = true
csrf.token_length = 32
csrf.cookie_name = _csrf
csrf.cookie_secure = true
csrf.cookie_httponly = true

# 会话安全
session.secure = true
session.httponly = true
session.samesite = strict
session.timeout = 7200

# 加密配置
encryption.key = your-32-byte-encryption-key-here
jwt.secret = your-jwt-secret-key-here
jwt.expire_hours = 24

# 速率限制
rate_limit.enabled = true
rate_limit.requests_per_minute = 60
rate_limit.burst = 100

# IP封禁配置
ip_ban.enabled = true
ip_ban.max_attempts = 5
ip_ban.ban_duration = 3600

# 安全日志
security_log.enabled = true
security_log.file = logs/security.log
security_log.level = info
```

### 2. 环境变量配置

```bash
# .env.production
export YYH_ENCRYPTION_KEY="your-32-byte-encryption-key-here!!"
export YYH_JWT_SECRET="your-super-secret-jwt-key-change-this"
export YYH_DB_PASSWORD="your-secure-database-password"
export YYH_REDIS_PASSWORD="your-redis-password"

# 安全相关环境变量
export YYH_FORCE_HTTPS=true
export YYH_CSRF_ENABLED=true
export YYH_RATE_LIMIT_ENABLED=true
export YYH_SECURITY_LOG_ENABLED=true
```

---

## 🧪 安全测试

### 1. 安全测试用例

```go
func TestCSRFProtection(t *testing.T) {
    app := setupTestApp()
    
    // 测试没有CSRF令牌的POST请求
    w := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/users", strings.NewReader("name=test"))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    app.ServeHTTP(w, req)
    
    assert.Equal(t, 403, w.Code)
    
    // 测试有效的CSRF令牌
    token := generateCSRFToken()
    formData := fmt.Sprintf("name=test&csrf_token=%s", token)
    
    w = httptest.NewRecorder()
    req = httptest.NewRequest("POST", "/users", strings.NewReader(formData))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    app.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}

func TestSQLInjectionProtection(t *testing.T) {
    app := setupTestApp()
    
    // 测试SQL注入尝试
    maliciousInput := "'; DROP TABLE users; --"
    
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/users?name="+url.QueryEscape(maliciousInput), nil)
    
    app.ServeHTTP(w, req)
    
    // 应该返回安全的结果，不会执行恶意SQL
    assert.Equal(t, 200, w.Code)
    
    // 验证数据库表仍然存在
    var count int64
    db.Model(&User{}).Count(&count)
    assert.Greater(t, count, int64(0))
}
```

### 2. 安全扫描集成

```yaml
# .github/workflows/security-scan.yml
name: Security Scan

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Install gosec
      run: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    
    - name: Run gosec
      run: gosec ./...
      
    - name: Install nancy
      run: go install github.com/sonatypecommunity/nancy@latest
      
    - name: Check for vulnerabilities
      run: go list -json -m all | nancy sleuth
```

---

## 📋 安全检查清单

### 开发阶段
- [ ] 启用CSRF保护
- [ ] 配置XSS防护
- [ ] 实现输入验证
- [ ] 使用参数化查询
- [ ] 加密敏感数据
- [ ] 实现安全的身份认证
- [ ] 配置适当的权限控制

### 测试阶段
- [ ] 进行安全测试
- [ ] 运行安全扫描工具
- [ ] 测试SQL注入防护
- [ ] 验证CSRF保护
- [ ] 检查XSS防护效果
- [ ] 测试认证和授权

### 部署阶段
- [ ] 启用HTTPS
- [ ] 配置安全头
- [ ] 设置防火墙规则
- [ ] 启用安全日志
- [ ] 配置入侵检测
- [ ] 设置备份和恢复

### 运维阶段
- [ ] 监控安全日志
- [ ] 定期更新依赖
- [ ] 进行安全审计
- [ ] 备份加密密钥
- [ ] 更新安全配置
- [ ] 培训开发团队

---

<div align="center">

**🔒 安全是一个持续的过程，不是一次性的任务**

**遵循这些安全最佳实践，让你的YYHertz应用固若金汤！🛡️**

</div>