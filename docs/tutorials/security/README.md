# YYHertz 安全加固方案教程

<div align="center">

🛡️ **企业级安全防护** | 从基础防护到高级安全策略

</div>

---

## 📋 目录

- [安全架构概述](#安全架构概述)
- [身份认证安全](#身份认证安全)
- [数据传输安全](#数据传输安全)
- [输入验证与防护](#输入验证与防护)
- [会话管理安全](#会话管理安全)
- [API安全防护](#api安全防护)
- [数据库安全](#数据库安全)
- [日志与审计](#日志与审计)
- [部署安全](#部署安全)
- [安全测试](#安全测试)

---

## 🎯 安全架构概述

### YYHertz安全防护体系

```
                    安全防护层级架构
                           │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
    🌐 网络安全层      🔐 应用安全层      💾 数据安全层
        │                 │                 │
   ┌────▼────┐       ┌────▼────┐       ┌────▼────┐
   │防火墙    │       │身份认证  │       │数据加密  │
   │DDoS防护 │       │权限控制  │       │访问控制  │
   │WAF防护  │       │输入验证  │       │备份恢复  │
   └─────────┘       └─────────┘       └─────────┘
        │                 │                 │
   ┌────▼────┐       ┌────▼────┐       ┌────▼────┐
   │SSL/TLS  │       │CSRF防护 │       │审计日志  │
   │证书管理  │       │XSS防护  │       │合规检查  │
   │协议安全  │       │SQL注入  │       │隐私保护  │
   └─────────┘       └─────────┘       └─────────┘
```

### 安全配置管理

```go
package security

import (
    "crypto/rand"
    "crypto/sha256"
    "time"
    "encoding/hex"
    "golang.org/x/crypto/bcrypt"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
    // 身份认证配置
    Auth AuthConfig `json:"auth"`
    
    // 会话管理配置
    Session SessionConfig `json:"session"`
    
    // 加密配置
    Encryption EncryptionConfig `json:"encryption"`
    
    // 安全头配置
    Headers HeadersConfig `json:"headers"`
    
    // CSRF配置
    CSRF CSRFConfig `json:"csrf"`
    
    // 限流配置
    RateLimit RateLimitConfig `json:"rate_limit"`
    
    // 审计配置
    Audit AuditConfig `json:"audit"`
}

// AuthConfig 认证配置
type AuthConfig struct {
    JWTSecret         string        `json:"jwt_secret"`
    JWTExpiration     time.Duration `json:"jwt_expiration"`
    PasswordMinLength int           `json:"password_min_length"`
    PasswordPolicy    []string      `json:"password_policy"`
    MaxLoginAttempts  int           `json:"max_login_attempts"`
    LockoutDuration   time.Duration `json:"lockout_duration"`
}

// SessionConfig 会话配置
type SessionConfig struct {
    CookieName       string        `json:"cookie_name"`
    CookieSecure     bool          `json:"cookie_secure"`
    CookieHTTPOnly   bool          `json:"cookie_http_only"`
    CookieSameSite   string        `json:"cookie_same_site"`
    SessionTimeout   time.Duration `json:"session_timeout"`
    SessionKeyLength int           `json:"session_key_length"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
    Algorithm    string `json:"algorithm"`
    KeyLength    int    `json:"key_length"`
    SaltLength   int    `json:"salt_length"`
    Iterations   int    `json:"iterations"`
}

// HeadersConfig 安全头配置
type HeadersConfig struct {
    EnableHSTS            bool   `json:"enable_hsts"`
    HSTSMaxAge            int    `json:"hsts_max_age"`
    EnableCSP             bool   `json:"enable_csp"`
    CSPDirectives         string `json:"csp_directives"`
    EnableXFrameOptions   bool   `json:"enable_x_frame_options"`
    XFrameOptions         string `json:"x_frame_options"`
    EnableXContentType    bool   `json:"enable_x_content_type"`
    EnableXSSProtection   bool   `json:"enable_xss_protection"`
}

// SecurityManager 安全管理器
type SecurityManager struct {
    config        *SecurityConfig
    passwordHasher *PasswordHasher
    tokenManager  *TokenManager
    auditLogger   *AuditLogger
    rateLimiter   *RateLimiter
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager() *SecurityManager {
    config := &SecurityConfig{
        Auth: AuthConfig{
            JWTSecret:         generateSecretKey(),
            JWTExpiration:     24 * time.Hour,
            PasswordMinLength: 8,
            PasswordPolicy:    []string{"uppercase", "lowercase", "number", "special"},
            MaxLoginAttempts:  5,
            LockoutDuration:   30 * time.Minute,
        },
        Session: SessionConfig{
            CookieName:       "yyhertz_session",
            CookieSecure:     true,
            CookieHTTPOnly:   true,
            CookieSameSite:   "Strict",
            SessionTimeout:   2 * time.Hour,
            SessionKeyLength: 32,
        },
        Encryption: EncryptionConfig{
            Algorithm:  "AES-256-GCM",
            KeyLength:  32,
            SaltLength: 16,
            Iterations: 100000,
        },
        Headers: HeadersConfig{
            EnableHSTS:          true,
            HSTSMaxAge:          31536000,
            EnableCSP:           true,
            CSPDirectives:       "default-src 'self'; script-src 'self' 'unsafe-inline'",
            EnableXFrameOptions: true,
            XFrameOptions:       "DENY",
            EnableXContentType:  true,
            EnableXSSProtection: true,
        },
    }
    
    return &SecurityManager{
        config:        config,
        passwordHasher: NewPasswordHasher(config.Encryption),
        tokenManager:  NewTokenManager(config.Auth),
        auditLogger:   NewAuditLogger(config.Audit),
        rateLimiter:   NewRateLimiter(config.RateLimit),
    }
}

// generateSecretKey 生成密钥
func generateSecretKey() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}

// SecurityMiddleware 安全中间件
func SecurityMiddleware(manager *SecurityManager) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        // 设置安全头
        manager.setSecurityHeaders(c)
        
        // CSRF保护
        if err := manager.csrfProtection(c); err != nil {
            c.JSON(403, map[string]string{"error": "CSRF token invalid"})
            return
        }
        
        // 限流检查
        if !manager.rateLimiter.Allow(c.ClientIP()) {
            c.JSON(429, map[string]string{"error": "Rate limit exceeded"})
            return
        }
        
        // 记录请求
        manager.auditLogger.LogRequest(c)
        
        c.Next()
        
        // 记录响应
        manager.auditLogger.LogResponse(c)
    }
}

// setSecurityHeaders 设置安全头
func (m *SecurityManager) setSecurityHeaders(c *mvc.Context) {
    headers := m.config.Headers
    
    if headers.EnableHSTS && c.IsHTTPS() {
        c.Header("Strict-Transport-Security", 
            fmt.Sprintf("max-age=%d; includeSubDomains", headers.HSTSMaxAge))
    }
    
    if headers.EnableCSP {
        c.Header("Content-Security-Policy", headers.CSPDirectives)
    }
    
    if headers.EnableXFrameOptions {
        c.Header("X-Frame-Options", headers.XFrameOptions)
    }
    
    if headers.EnableXContentType {
        c.Header("X-Content-Type-Options", "nosniff")
    }
    
    if headers.EnableXSSProtection {
        c.Header("X-XSS-Protection", "1; mode=block")
    }
    
    // 移除敏感信息头
    c.Header("X-Powered-By", "")
    c.Header("Server", "")
}
```

---

## 🔐 身份认证安全

### 1. 多因素认证系统

```go
package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base32"
    "encoding/hex"
    "fmt"
    "time"
    
    "github.com/pquerna/otp"
    "github.com/pquerna/otp/totp"
    "golang.org/x/crypto/bcrypt"
)

// AuthenticationManager 认证管理器
type AuthenticationManager struct {
    config       *AuthConfig
    userStore    UserStore
    tokenStore   TokenStore
    attemptStore AttemptStore
    mfaProvider  *MFAProvider
}

// UserStore 用户存储接口
type UserStore interface {
    GetUser(username string) (*User, error)
    UpdateUser(user *User) error
    CreateUser(user *User) error
}

// TokenStore 令牌存储接口
type TokenStore interface {
    StoreToken(token string, userID uint, expiration time.Time) error
    ValidateToken(token string) (*TokenInfo, error)
    RevokeToken(token string) error
    CleanupExpiredTokens() error
}

// AttemptStore 登录尝试存储接口
type AttemptStore interface {
    RecordAttempt(identifier string, success bool) error
    GetAttempts(identifier string, duration time.Duration) (int, error)
    ClearAttempts(identifier string) error
}

// User 用户模型
type User struct {
    ID                uint      `json:"id"`
    Username          string    `json:"username"`
    Email             string    `json:"email"`
    PasswordHash      string    `json:"-"`
    Salt              string    `json:"-"`
    MFAEnabled        bool      `json:"mfa_enabled"`
    MFASecret         string    `json:"-"`
    AccountLocked     bool      `json:"account_locked"`
    LockExpiration    time.Time `json:"lock_expiration"`
    LastLogin         time.Time `json:"last_login"`
    FailedAttempts    int       `json:"failed_attempts"`
    PasswordChangedAt time.Time `json:"password_changed_at"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
}

// TokenInfo 令牌信息
type TokenInfo struct {
    UserID     uint      `json:"user_id"`
    Username   string    `json:"username"`
    IssuedAt   time.Time `json:"issued_at"`
    ExpiresAt  time.Time `json:"expires_at"`
    Scope      []string  `json:"scope"`
}

// LoginRequest 登录请求
type LoginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
    MFACode  string `json:"mfa_code"`
    RememberMe bool `json:"remember_me"`
}

// LoginResponse 登录响应
type LoginResponse struct {
    Success      bool      `json:"success"`
    Token        string    `json:"token,omitempty"`
    RefreshToken string    `json:"refresh_token,omitempty"`
    ExpiresAt    time.Time `json:"expires_at,omitempty"`
    MFARequired  bool      `json:"mfa_required"`
    Message      string    `json:"message"`
}

// NewAuthenticationManager 创建认证管理器
func NewAuthenticationManager(config *AuthConfig, userStore UserStore) *AuthenticationManager {
    return &AuthenticationManager{
        config:       config,
        userStore:    userStore,
        tokenStore:   NewRedisTokenStore(),
        attemptStore: NewRedisAttemptStore(),
        mfaProvider:  NewMFAProvider(),
    }
}

// Login 用户登录
func (am *AuthenticationManager) Login(req *LoginRequest) (*LoginResponse, error) {
    // 1. 检查登录尝试次数
    attempts, err := am.attemptStore.GetAttempts(req.Username, am.config.LockoutDuration)
    if err != nil {
        return nil, err
    }
    
    if attempts >= am.config.MaxLoginAttempts {
        return &LoginResponse{
            Success: false,
            Message: "Account temporarily locked due to too many failed attempts",
        }, nil
    }
    
    // 2. 获取用户信息
    user, err := am.userStore.GetUser(req.Username)
    if err != nil {
        am.attemptStore.RecordAttempt(req.Username, false)
        return &LoginResponse{
            Success: false,
            Message: "Invalid username or password",
        }, nil
    }
    
    // 3. 检查账户状态
    if user.AccountLocked && time.Now().Before(user.LockExpiration) {
        return &LoginResponse{
            Success: false,
            Message: "Account is locked",
        }, nil
    }
    
    // 4. 验证密码
    if !am.verifyPassword(req.Password, user.PasswordHash, user.Salt) {
        am.attemptStore.RecordAttempt(req.Username, false)
        user.FailedAttempts++
        
        // 达到最大尝试次数，锁定账户
        if user.FailedAttempts >= am.config.MaxLoginAttempts {
            user.AccountLocked = true
            user.LockExpiration = time.Now().Add(am.config.LockoutDuration)
        }
        
        am.userStore.UpdateUser(user)
        
        return &LoginResponse{
            Success: false,
            Message: "Invalid username or password",
        }, nil
    }
    
    // 5. 检查是否需要MFA
    if user.MFAEnabled {
        if req.MFACode == "" {
            return &LoginResponse{
                Success:     false,
                MFARequired: true,
                Message:     "MFA code required",
            }, nil
        }
        
        // 验证MFA代码
        if !am.mfaProvider.ValidateTOTP(user.MFASecret, req.MFACode) {
            am.attemptStore.RecordAttempt(req.Username, false)
            return &LoginResponse{
                Success: false,
                Message: "Invalid MFA code",
            }, nil
        }
    }
    
    // 6. 登录成功，生成令牌
    token, err := am.generateToken(user, req.RememberMe)
    if err != nil {
        return nil, err
    }
    
    // 7. 更新用户状态
    user.LastLogin = time.Now()
    user.FailedAttempts = 0
    user.AccountLocked = false
    am.userStore.UpdateUser(user)
    
    // 8. 清除失败尝试记录
    am.attemptStore.ClearAttempts(req.Username)
    
    return &LoginResponse{
        Success:   true,
        Token:     token,
        ExpiresAt: time.Now().Add(am.config.JWTExpiration),
        Message:   "Login successful",
    }, nil
}

// generateToken 生成JWT令牌
func (am *AuthenticationManager) generateToken(user *User, rememberMe bool) (string, error) {
    expiration := am.config.JWTExpiration
    if rememberMe {
        expiration = 30 * 24 * time.Hour // 30天
    }
    
    claims := jwt.MapClaims{
        "user_id":  user.ID,
        "username": user.Username,
        "email":    user.Email,
        "iat":      time.Now().Unix(),
        "exp":      time.Now().Add(expiration).Unix(),
        "scope":    []string{"read", "write"},
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(am.config.JWTSecret))
    if err != nil {
        return "", err
    }
    
    // 存储令牌信息
    am.tokenStore.StoreToken(tokenString, user.ID, time.Now().Add(expiration))
    
    return tokenString, nil
}

// verifyPassword 验证密码
func (am *AuthenticationManager) verifyPassword(password, hash, salt string) bool {
    // 使用salt重新计算hash
    saltedPassword := password + salt
    expectedHash := sha256.Sum256([]byte(saltedPassword))
    return hex.EncodeToString(expectedHash[:]) == hash
}

// MFAProvider MFA提供者
type MFAProvider struct {
    issuer string
}

// NewMFAProvider 创建MFA提供者
func NewMFAProvider() *MFAProvider {
    return &MFAProvider{
        issuer: "YYHertz App",
    }
}

// GenerateSecret 生成MFA密钥
func (mfa *MFAProvider) GenerateSecret(username string) (*otp.Key, error) {
    return totp.Generate(totp.GenerateOpts{
        Issuer:      mfa.issuer,
        AccountName: username,
        SecretSize:  32,
    })
}

// ValidateTOTP 验证TOTP代码
func (mfa *MFAProvider) ValidateTOTP(secret, code string) bool {
    return totp.Validate(code, secret)
}

// EnableMFA 启用MFA
func (am *AuthenticationManager) EnableMFA(userID uint) (*MFASetupResponse, error) {
    user, err := am.userStore.GetUser(fmt.Sprintf("%d", userID))
    if err != nil {
        return nil, err
    }
    
    // 生成MFA密钥
    key, err := am.mfaProvider.GenerateSecret(user.Username)
    if err != nil {
        return nil, err
    }
    
    // 更新用户MFA密钥（但还未启用）
    user.MFASecret = key.Secret()
    am.userStore.UpdateUser(user)
    
    return &MFASetupResponse{
        Secret:  key.Secret(),
        QRCode:  key.String(),
        Message: "Scan QR code with authenticator app",
    }, nil
}

// MFASetupResponse MFA设置响应
type MFASetupResponse struct {
    Secret  string `json:"secret"`
    QRCode  string `json:"qr_code"`
    Message string `json:"message"`
}

// ConfirmMFA 确认MFA设置
func (am *AuthenticationManager) ConfirmMFA(userID uint, code string) error {
    user, err := am.userStore.GetUser(fmt.Sprintf("%d", userID))
    if err != nil {
        return err
    }
    
    // 验证TOTP代码
    if !am.mfaProvider.ValidateTOTP(user.MFASecret, code) {
        return fmt.Errorf("invalid MFA code")
    }
    
    // 启用MFA
    user.MFAEnabled = true
    return am.userStore.UpdateUser(user)
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
    MinLength      int      `json:"min_length"`
    RequireUpper   bool     `json:"require_upper"`
    RequireLower   bool     `json:"require_lower"`
    RequireNumber  bool     `json:"require_number"`
    RequireSpecial bool     `json:"require_special"`
    ForbiddenWords []string `json:"forbidden_words"`
}

// ValidatePassword 验证密码强度
func (am *AuthenticationManager) ValidatePassword(password string) []string {
    var errors []string
    policy := am.getPasswordPolicy()
    
    if len(password) < policy.MinLength {
        errors = append(errors, fmt.Sprintf("Password must be at least %d characters", policy.MinLength))
    }
    
    if policy.RequireUpper && !containsUpper(password) {
        errors = append(errors, "Password must contain at least one uppercase letter")
    }
    
    if policy.RequireLower && !containsLower(password) {
        errors = append(errors, "Password must contain at least one lowercase letter")
    }
    
    if policy.RequireNumber && !containsNumber(password) {
        errors = append(errors, "Password must contain at least one number")
    }
    
    if policy.RequireSpecial && !containsSpecial(password) {
        errors = append(errors, "Password must contain at least one special character")
    }
    
    // 检查禁用词汇
    for _, word := range policy.ForbiddenWords {
        if strings.Contains(strings.ToLower(password), strings.ToLower(word)) {
            errors = append(errors, "Password contains forbidden words")
            break
        }
    }
    
    return errors
}

// HashPassword 哈希密码
func (am *AuthenticationManager) HashPassword(password string) (string, string, error) {
    // 生成随机盐值
    salt := make([]byte, 16)
    rand.Read(salt)
    saltStr := hex.EncodeToString(salt)
    
    // 计算密码哈希
    saltedPassword := password + saltStr
    hash := sha256.Sum256([]byte(saltedPassword))
    hashStr := hex.EncodeToString(hash[:])
    
    return hashStr, saltStr, nil
}

// getPasswordPolicy 获取密码策略
func (am *AuthenticationManager) getPasswordPolicy() *PasswordPolicy {
    return &PasswordPolicy{
        MinLength:      am.config.PasswordMinLength,
        RequireUpper:   contains(am.config.PasswordPolicy, "uppercase"),
        RequireLower:   contains(am.config.PasswordPolicy, "lowercase"),
        RequireNumber:  contains(am.config.PasswordPolicy, "number"),
        RequireSpecial: contains(am.config.PasswordPolicy, "special"),
        ForbiddenWords: []string{"password", "123456", "admin"},
    }
}

// 辅助函数
func containsUpper(s string) bool {
    for _, r := range s {
        if r >= 'A' && r <= 'Z' {
            return true
        }
    }
    return false
}

func containsLower(s string) bool {
    for _, r := range s {
        if r >= 'a' && r <= 'z' {
            return true
        }
    }
    return false
}

func containsNumber(s string) bool {
    for _, r := range s {
        if r >= '0' && r <= '9' {
            return true
        }
    }
    return false
}

func containsSpecial(s string) bool {
    special := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
    for _, r := range s {
        if strings.ContainsRune(special, r) {
            return true
        }
    }
    return false
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

---

## 🌐 数据传输安全

### 1. HTTPS与TLS配置

```go
package tls

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "net/http"
    "time"
)

// TLSManager TLS管理器
type TLSManager struct {
    config     *TLSConfig
    certStore  CertificateStore
    validator  *CertificateValidator
}

// TLSConfig TLS配置
type TLSConfig struct {
    CertFile         string        `json:"cert_file"`
    KeyFile          string        `json:"key_file"`
    CAFile           string        `json:"ca_file"`
    MinVersion       uint16        `json:"min_version"`
    MaxVersion       uint16        `json:"max_version"`
    CipherSuites     []uint16      `json:"cipher_suites"`
    CurvePreferences []tls.CurveID `json:"curve_preferences"`
    EnableSNI        bool          `json:"enable_sni"`
    EnableOCSP       bool          `json:"enable_ocsp"`
    CertRenewal      time.Duration `json:"cert_renewal"`
}

// CertificateStore 证书存储接口
type CertificateStore interface {
    StoreCertificate(domain string, cert *tls.Certificate) error
    GetCertificate(domain string) (*tls.Certificate, error)
    ListCertificates() ([]*CertificateInfo, error)
    DeleteCertificate(domain string) error
}

// CertificateInfo 证书信息
type CertificateInfo struct {
    Domain      string    `json:"domain"`
    Issuer      string    `json:"issuer"`
    Subject     string    `json:"subject"`
    NotBefore   time.Time `json:"not_before"`
    NotAfter    time.Time `json:"not_after"`
    IsExpired   bool      `json:"is_expired"`
    DaysToExpiry int      `json:"days_to_expiry"`
}

// NewTLSManager 创建TLS管理器
func NewTLSManager(config *TLSConfig) *TLSManager {
    if config == nil {
        config = &TLSConfig{
            MinVersion: tls.VersionTLS12,
            MaxVersion: tls.VersionTLS13,
            CipherSuites: []uint16{
                tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
                tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            },
            CurvePreferences: []tls.CurveID{
                tls.CurveP521,
                tls.CurveP384,
                tls.CurveP256,
            },
            EnableSNI:   true,
            EnableOCSP:  true,
            CertRenewal: 24 * time.Hour,
        }
    }
    
    return &TLSManager{
        config:    config,
        certStore: NewFileCertificateStore(),
        validator: NewCertificateValidator(),
    }
}

// CreateTLSConfig 创建TLS配置
func (tm *TLSManager) CreateTLSConfig() (*tls.Config, error) {
    config := &tls.Config{
        MinVersion:               tm.config.MinVersion,
        MaxVersion:               tm.config.MaxVersion,
        CipherSuites:             tm.config.CipherSuites,
        CurvePreferences:         tm.config.CurvePreferences,
        PreferServerCipherSuites: true,
        GetCertificate:           tm.getCertificate,
    }
    
    // 加载根CA证书
    if tm.config.CAFile != "" {
        caCert, err := ioutil.ReadFile(tm.config.CAFile)
        if err != nil {
            return nil, fmt.Errorf("failed to read CA file: %w", err)
        }
        
        caCertPool := x509.NewCertPool()
        caCertPool.AppendCertsFromPEM(caCert)
        config.ClientCAs = caCertPool
        config.ClientAuth = tls.RequireAndVerifyClientCert
    }
    
    return config, nil
}

// getCertificate 获取证书
func (tm *TLSManager) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
    domain := hello.ServerName
    if domain == "" {
        domain = "default"
    }
    
    // 从存储中获取证书
    cert, err := tm.certStore.GetCertificate(domain)
    if err != nil {
        return nil, fmt.Errorf("certificate not found for domain %s: %w", domain, err)
    }
    
    // 验证证书
    if err := tm.validator.ValidateCertificate(cert); err != nil {
        return nil, fmt.Errorf("certificate validation failed: %w", err)
    }
    
    return cert, nil
}

// CertificateValidator 证书验证器
type CertificateValidator struct {
    renewalThreshold time.Duration
}

// NewCertificateValidator 创建证书验证器
func NewCertificateValidator() *CertificateValidator {
    return &CertificateValidator{
        renewalThreshold: 30 * 24 * time.Hour, // 30天
    }
}

// ValidateCertificate 验证证书
func (cv *CertificateValidator) ValidateCertificate(cert *tls.Certificate) error {
    if len(cert.Certificate) == 0 {
        return fmt.Errorf("empty certificate chain")
    }
    
    // 解析证书
    x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
    if err != nil {
        return fmt.Errorf("failed to parse certificate: %w", err)
    }
    
    // 检查证书是否过期
    now := time.Now()
    if now.Before(x509Cert.NotBefore) {
        return fmt.Errorf("certificate not yet valid")
    }
    
    if now.After(x509Cert.NotAfter) {
        return fmt.Errorf("certificate has expired")
    }
    
    // 检查是否需要更新
    timeToExpiry := x509Cert.NotAfter.Sub(now)
    if timeToExpiry < cv.renewalThreshold {
        return fmt.Errorf("certificate expires soon: %v", timeToExpiry)
    }
    
    return nil
}

// AutoRenewalManager 自动续期管理器
type AutoRenewalManager struct {
    tlsManager *TLSManager
    interval   time.Duration
    stopCh     chan struct{}
}

// NewAutoRenewalManager 创建自动续期管理器
func NewAutoRenewalManager(tlsManager *TLSManager) *AutoRenewalManager {
    return &AutoRenewalManager{
        tlsManager: tlsManager,
        interval:   24 * time.Hour,
        stopCh:     make(chan struct{}),
    }
}

// Start 启动自动续期
func (arm *AutoRenewalManager) Start() {
    ticker := time.NewTicker(arm.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            arm.checkAndRenewCertificates()
        case <-arm.stopCh:
            return
        }
    }
}

// Stop 停止自动续期
func (arm *AutoRenewalManager) Stop() {
    close(arm.stopCh)
}

// checkAndRenewCertificates 检查并更新证书
func (arm *AutoRenewalManager) checkAndRenewCertificates() {
    certificates, err := arm.tlsManager.certStore.ListCertificates()
    if err != nil {
        log.Printf("Failed to list certificates: %v", err)
        return
    }
    
    for _, certInfo := range certificates {
        if certInfo.DaysToExpiry <= 30 { // 30天内过期
            log.Printf("Certificate for %s expires in %d days, attempting renewal", 
                certInfo.Domain, certInfo.DaysToExpiry)
            
            if err := arm.renewCertificate(certInfo.Domain); err != nil {
                log.Printf("Failed to renew certificate for %s: %v", certInfo.Domain, err)
            } else {
                log.Printf("Successfully renewed certificate for %s", certInfo.Domain)
            }
        }
    }
}

// renewCertificate 更新证书
func (arm *AutoRenewalManager) renewCertificate(domain string) error {
    // 这里应该实现具体的证书更新逻辑
    // 可以集成Let's Encrypt ACME客户端
    // 或者调用其他CA的API
    
    // 示例：使用ACME协议更新证书
    acmeClient := NewACMEClient()
    newCert, err := acmeClient.RenewCertificate(domain)
    if err != nil {
        return err
    }
    
    // 保存新证书
    return arm.tlsManager.certStore.StoreCertificate(domain, newCert)
}

// HTTPSRedirectMiddleware HTTPS重定向中间件
func HTTPSRedirectMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        if !c.IsHTTLS() {
            // 重定向到HTTPS
            httpsURL := "https://" + c.Request.Host + c.Request.URL.Path
            if c.Request.URL.RawQuery != "" {
                httpsURL += "?" + c.Request.URL.RawQuery
            }
            
            c.Redirect(301, httpsURL)
            return
        }
        
        c.Next()
    }
}

// HSTSMiddleware HSTS中间件
func HSTSMiddleware(maxAge int) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        if c.IsHTTPS() {
            c.Header("Strict-Transport-Security", 
                fmt.Sprintf("max-age=%d; includeSubDomains; preload", maxAge))
        }
        c.Next()
    }
}
```

### 2. 数据加密与解密

```go
package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "io"
    
    "golang.org/x/crypto/pbkdf2"
    "golang.org/x/crypto/scrypt"
)

// EncryptionManager 加密管理器
type EncryptionManager struct {
    config *EncryptionConfig
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
    Algorithm    string `json:"algorithm"`     // AES-256-GCM
    KeySize      int    `json:"key_size"`      // 32 bytes
    NonceSize    int    `json:"nonce_size"`    // 12 bytes for GCM
    SaltSize     int    `json:"salt_size"`     // 16 bytes
    Iterations   int    `json:"iterations"`    // PBKDF2 iterations
    DefaultKey   string `json:"default_key"`   // Default encryption key
}

// NewEncryptionManager 创建加密管理器
func NewEncryptionManager(config *EncryptionConfig) *EncryptionManager {
    if config == nil {
        config = &EncryptionConfig{
            Algorithm:  "AES-256-GCM",
            KeySize:    32,
            NonceSize:  12,
            SaltSize:   16,
            Iterations: 100000,
            DefaultKey: generateRandomKey(32),
        }
    }
    
    return &EncryptionManager{
        config: config,
    }
}

// Encrypt 加密数据
func (em *EncryptionManager) Encrypt(data []byte, password string) (string, error) {
    // 生成随机盐值
    salt := make([]byte, em.config.SaltSize)
    if _, err := rand.Read(salt); err != nil {
        return "", fmt.Errorf("failed to generate salt: %w", err)
    }
    
    // 派生密钥
    key := pbkdf2.Key([]byte(password), salt, em.config.Iterations, em.config.KeySize, sha256.New)
    
    // 创建AES加密器
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("failed to create cipher: %w", err)
    }
    
    // 使用GCM模式
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("failed to create GCM: %w", err)
    }
    
    // 生成随机nonce
    nonce := make([]byte, em.config.NonceSize)
    if _, err := rand.Read(nonce); err != nil {
        return "", fmt.Errorf("failed to generate nonce: %w", err)
    }
    
    // 加密数据
    ciphertext := gcm.Seal(nil, nonce, data, nil)
    
    // 组合: salt + nonce + ciphertext
    result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
    result = append(result, salt...)
    result = append(result, nonce...)
    result = append(result, ciphertext...)
    
    return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt 解密数据
func (em *EncryptionManager) Decrypt(encryptedData, password string) ([]byte, error) {
    // Base64解码
    data, err := base64.StdEncoding.DecodeString(encryptedData)
    if err != nil {
        return nil, fmt.Errorf("failed to decode base64: %w", err)
    }
    
    // 检查数据长度
    if len(data) < em.config.SaltSize+em.config.NonceSize {
        return nil, fmt.Errorf("encrypted data too short")
    }
    
    // 提取salt、nonce和密文
    salt := data[:em.config.SaltSize]
    nonce := data[em.config.SaltSize : em.config.SaltSize+em.config.NonceSize]
    ciphertext := data[em.config.SaltSize+em.config.NonceSize:]
    
    // 派生密钥
    key := pbkdf2.Key([]byte(password), salt, em.config.Iterations, em.config.KeySize, sha256.New)
    
    // 创建AES解密器
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, fmt.Errorf("failed to create cipher: %w", err)
    }
    
    // 使用GCM模式
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCM: %w", err)
    }
    
    // 解密数据
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt: %w", err)
    }
    
    return plaintext, nil
}

// EncryptString 加密字符串
func (em *EncryptionManager) EncryptString(text, password string) (string, error) {
    return em.Encrypt([]byte(text), password)
}

// DecryptString 解密字符串
func (em *EncryptionManager) DecryptString(encryptedText, password string) (string, error) {
    data, err := em.Decrypt(encryptedText, password)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

// FieldEncryption 字段加密器
type FieldEncryption struct {
    manager *EncryptionManager
    key     string
}

// NewFieldEncryption 创建字段加密器
func NewFieldEncryption(manager *EncryptionManager, key string) *FieldEncryption {
    return &FieldEncryption{
        manager: manager,
        key:     key,
    }
}

// EncryptField 加密字段
func (fe *FieldEncryption) EncryptField(value string) (string, error) {
    if value == "" {
        return "", nil
    }
    return fe.manager.EncryptString(value, fe.key)
}

// DecryptField 解密字段
func (fe *FieldEncryption) DecryptField(encryptedValue string) (string, error) {
    if encryptedValue == "" {
        return "", nil
    }
    return fe.manager.DecryptString(encryptedValue, fe.key)
}

// EncryptedModel 加密模型接口
type EncryptedModel interface {
    GetEncryptionFields() []string
    BeforeSave() error
    AfterFind() error
}

// ModelEncryption 模型加密器
type ModelEncryption struct {
    fieldEncryption *FieldEncryption
}

// NewModelEncryption 创建模型加密器
func NewModelEncryption(manager *EncryptionManager, key string) *ModelEncryption {
    return &ModelEncryption{
        fieldEncryption: NewFieldEncryption(manager, key),
    }
}

// EncryptModel 加密模型
func (me *ModelEncryption) EncryptModel(model EncryptedModel) error {
    fields := model.GetEncryptionFields()
    modelValue := reflect.ValueOf(model).Elem()
    
    for _, fieldName := range fields {
        field := modelValue.FieldByName(fieldName)
        if !field.IsValid() || !field.CanSet() {
            continue
        }
        
        if field.Kind() == reflect.String {
            originalValue := field.String()
            if originalValue != "" {
                encryptedValue, err := me.fieldEncryption.EncryptField(originalValue)
                if err != nil {
                    return fmt.Errorf("failed to encrypt field %s: %w", fieldName, err)
                }
                field.SetString(encryptedValue)
            }
        }
    }
    
    return nil
}

// DecryptModel 解密模型
func (me *ModelEncryption) DecryptModel(model EncryptedModel) error {
    fields := model.GetEncryptionFields()
    modelValue := reflect.ValueOf(model).Elem()
    
    for _, fieldName := range fields {
        field := modelValue.FieldByName(fieldName)
        if !field.IsValid() || !field.CanSet() {
            continue
        }
        
        if field.Kind() == reflect.String {
            encryptedValue := field.String()
            if encryptedValue != "" {
                decryptedValue, err := me.fieldEncryption.DecryptField(encryptedValue)
                if err != nil {
                    return fmt.Errorf("failed to decrypt field %s: %w", fieldName, err)
                }
                field.SetString(decryptedValue)
            }
        }
    }
    
    return nil
}

// 示例加密模型
type User struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`          // 需要加密
    Phone    string `json:"phone"`          // 需要加密
    SSN      string `json:"ssn"`            // 需要加密 - 社会保险号
}

// GetEncryptionFields 获取需要加密的字段
func (u *User) GetEncryptionFields() []string {
    return []string{"Email", "Phone", "SSN"}
}

// BeforeSave 保存前加密
func (u *User) BeforeSave() error {
    modelEncryption := GetGlobalModelEncryption()
    return modelEncryption.EncryptModel(u)
}

// AfterFind 查找后解密
func (u *User) AfterFind() error {
    modelEncryption := GetGlobalModelEncryption()
    return modelEncryption.DecryptModel(u)
}

var globalModelEncryption *ModelEncryption

// SetGlobalModelEncryption 设置全局模型加密器
func SetGlobalModelEncryption(me *ModelEncryption) {
    globalModelEncryption = me
}

// GetGlobalModelEncryption 获取全局模型加密器
func GetGlobalModelEncryption() *ModelEncryption {
    return globalModelEncryption
}

// generateRandomKey 生成随机密钥
func generateRandomKey(length int) string {
    bytes := make([]byte, length)
    rand.Read(bytes)
    return base64.StdEncoding.EncodeToString(bytes)
}

// KeyManager 密钥管理器
type KeyManager struct {
    keys map[string]string
    mutex sync.RWMutex
}

// NewKeyManager 创建密钥管理器
func NewKeyManager() *KeyManager {
    return &KeyManager{
        keys: make(map[string]string),
    }
}

// AddKey 添加密钥
func (km *KeyManager) AddKey(name, key string) {
    km.mutex.Lock()
    defer km.mutex.Unlock()
    km.keys[name] = key
}

// GetKey 获取密钥
func (km *KeyManager) GetKey(name string) (string, bool) {
    km.mutex.RLock()
    defer km.mutex.RUnlock()
    key, exists := km.keys[name]
    return key, exists
}

// RotateKey 轮换密钥
func (km *KeyManager) RotateKey(name string) string {
    newKey := generateRandomKey(32)
    km.AddKey(name, newKey)
    return newKey
}
```

---

## 🛡️ 输入验证与防护

### 1. XSS防护

```go
package security

import (
    "html"
    "net/url"
    "regexp"
    "strings"
)

// XSSProtection XSS防护器
type XSSProtection struct {
    whitelistTags map[string]bool
    whitelistAttrs map[string]bool
    patterns      []*regexp.Regexp
}

// NewXSSProtection 创建XSS防护器
func NewXSSProtection() *XSSProtection {
    xss := &XSSProtection{
        whitelistTags: map[string]bool{
            "b": true, "i": true, "u": true, "strong": true, "em": true,
            "p": true, "br": true, "div": true, "span": true,
        },
        whitelistAttrs: map[string]bool{
            "class": true, "id": true, "title": true,
        },
    }
    
    // 编译危险模式
    patterns := []string{
        `<script[^>]*>.*?</script>`,
        `javascript:`,
        `vbscript:`,
        `on\w+\s*=`,
        `<iframe[^>]*>.*?</iframe>`,
        `<object[^>]*>.*?</object>`,
        `<embed[^>]*>.*?</embed>`,
        `<form[^>]*>.*?</form>`,
        `<meta[^>]*>`,
        `<link[^>]*>`,
    }
    
    for _, pattern := range patterns {
        if regex, err := regexp.Compile(`(?i)` + pattern); err == nil {
            xss.patterns = append(xss.patterns, regex)
        }
    }
    
    return xss
}

// SanitizeHTML 清理HTML
func (xss *XSSProtection) SanitizeHTML(input string) string {
    // 1. HTML实体编码
    sanitized := html.EscapeString(input)
    
    // 2. 移除危险模式
    for _, pattern := range xss.patterns {
        sanitized = pattern.ReplaceAllString(sanitized, "")
    }
    
    // 3. 清理属性
    sanitized = xss.cleanAttributes(sanitized)
    
    return sanitized
}

// cleanAttributes 清理属性
func (xss *XSSProtection) cleanAttributes(input string) string {
    // 简化的属性清理实现
    // 移除所有事件处理器属性
    eventHandlers := []string{
        "onload", "onclick", "onmouseover", "onmouseout", "onmousemove",
        "onmousedown", "onmouseup", "onfocus", "onblur", "onchange",
        "onsubmit", "onreset", "onselect", "onkeydown", "onkeypress",
        "onkeyup", "onerror", "onabort",
    }
    
    for _, handler := range eventHandlers {
        re := regexp.MustCompile(`(?i)\s*` + handler + `\s*=\s*[^>\s]*`)
        input = re.ReplaceAllString(input, "")
    }
    
    return input
}

// FilterInput 过滤输入
func (xss *XSSProtection) FilterInput(input string) string {
    // URL解码
    decoded, _ := url.QueryUnescape(input)
    
    // HTML清理
    sanitized := xss.SanitizeHTML(decoded)
    
    // 移除控制字符
    sanitized = xss.removeControlChars(sanitized)
    
    return sanitized
}

// removeControlChars 移除控制字符
func (xss *XSSProtection) removeControlChars(input string) string {
    return regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(input, "")
}

// ValidateUserInput 验证用户输入
func (xss *XSSProtection) ValidateUserInput(input string) []string {
    var issues []string
    
    // 检查是否包含脚本标签
    if strings.Contains(strings.ToLower(input), "<script") {
        issues = append(issues, "Input contains script tags")
    }
    
    // 检查是否包含JavaScript
    if strings.Contains(strings.ToLower(input), "javascript:") {
        issues = append(issues, "Input contains JavaScript protocol")
    }
    
    // 检查是否包含事件处理器
    eventPattern := regexp.MustCompile(`(?i)on\w+\s*=`)
    if eventPattern.MatchString(input) {
        issues = append(issues, "Input contains event handlers")
    }
    
    return issues
}

// XSSMiddleware XSS防护中间件
func XSSMiddleware() mvc.HandlerFunc {
    xssProtection := NewXSSProtection()
    
    return func(c *mvc.Context) {
        // 清理查询参数
        for key, values := range c.Request.URL.Query() {
            for i, value := range values {
                values[i] = xssProtection.FilterInput(value)
            }
            c.Request.URL.Query()[key] = values
        }
        
        // 清理表单数据
        if c.Request.Method == "POST" {
            c.Request.ParseForm()
            for key, values := range c.Request.PostForm {
                for i, value := range values {
                    values[i] = xssProtection.FilterInput(value)
                }
                c.Request.PostForm[key] = values
            }
        }
        
        c.Next()
    }
}
```

### 2. SQL注入防护

```go
package security

import (
    "database/sql"
    "fmt"
    "regexp"
    "strings"
    
    "gorm.io/gorm"
)

// SQLInjectionProtection SQL注入防护器
type SQLInjectionProtection struct {
    dangerousPatterns []*regexp.Regexp
    keywordBlacklist  map[string]bool
}

// NewSQLInjectionProtection 创建SQL注入防护器
func NewSQLInjectionProtection() *SQLInjectionProtection {
    sip := &SQLInjectionProtection{
        keywordBlacklist: map[string]bool{
            "union":     true,
            "select":    true,
            "insert":    true,
            "update":    true,
            "delete":    true,
            "drop":      true,
            "create":    true,
            "alter":     true,
            "execute":   true,
            "exec":      true,
            "declare":   true,
            "script":    true,
            "xp_":       true,
            "sp_":       true,
        },
    }
    
    // 编译危险模式
    patterns := []string{
        `(\b(union|select|insert|update|delete|drop|create|alter)\b)`,
        `(\b(exec|execute|xp_|sp_)\b)`,
        `(--|\#|/\*|\*/)`,
        `(;|\||&)`,
        `('|"|`+"`"+`).*?(union|select|insert|update|delete)`,
        `\b(or|and)\s+\w+\s*[=><]`,
        `\b(true|false)\s*[=><]\s*(true|false)`,
        `1\s*=\s*1`,
        `''\s*=\s*''`,
        `benchmark\s*\(`,
        `sleep\s*\(`,
        `waitfor\s+delay`,
    }
    
    for _, pattern := range patterns {
        if regex, err := regexp.Compile(`(?i)` + pattern); err == nil {
            sip.dangerousPatterns = append(sip.dangerousPatterns, regex)
        }
    }
    
    return sip
}

// ValidateInput 验证输入
func (sip *SQLInjectionProtection) ValidateInput(input string) []string {
    var issues []string
    lowercaseInput := strings.ToLower(input)
    
    // 检查危险模式
    for _, pattern := range sip.dangerousPatterns {
        if pattern.MatchString(lowercaseInput) {
            issues = append(issues, fmt.Sprintf("Input matches dangerous pattern: %s", pattern.String()))
        }
    }
    
    // 检查黑名单关键词
    for keyword := range sip.keywordBlacklist {
        if strings.Contains(lowercaseInput, keyword) {
            issues = append(issues, fmt.Sprintf("Input contains blacklisted keyword: %s", keyword))
        }
    }
    
    return issues
}

// SanitizeInput 清理输入
func (sip *SQLInjectionProtection) SanitizeInput(input string) string {
    // 移除或转义危险字符
    sanitized := input
    
    // 转义单引号
    sanitized = strings.ReplaceAll(sanitized, "'", "''")
    
    // 移除注释符号
    sanitized = regexp.MustCompile(`--.*$`).ReplaceAllString(sanitized, "")
    sanitized = strings.ReplaceAll(sanitized, "/*", "")
    sanitized = strings.ReplaceAll(sanitized, "*/", "")
    
    // 移除分号（在某些情况下）
    sanitized = strings.ReplaceAll(sanitized, ";", "")
    
    return sanitized
}

// SafeQueryBuilder 安全查询构建器
type SafeQueryBuilder struct {
    db         *gorm.DB
    protection *SQLInjectionProtection
}

// NewSafeQueryBuilder 创建安全查询构建器
func NewSafeQueryBuilder(db *gorm.DB) *SafeQueryBuilder {
    return &SafeQueryBuilder{
        db:         db,
        protection: NewSQLInjectionProtection(),
    }
}

// SafeWhere 安全的WHERE条件
func (sqb *SafeQueryBuilder) SafeWhere(query *gorm.DB, condition string, args ...interface{}) *gorm.DB {
    // 验证条件字符串
    if issues := sqb.protection.ValidateInput(condition); len(issues) > 0 {
        // 记录安全警告
        log.Printf("SQL Injection attempt detected: %v", issues)
        return query.Where("1 = 0") // 返回空结果
    }
    
    // 验证参数
    for i, arg := range args {
        if str, ok := arg.(string); ok {
            if issues := sqb.protection.ValidateInput(str); len(issues) > 0 {
                log.Printf("SQL Injection attempt in parameter %d: %v", i, issues)
                args[i] = sqb.protection.SanitizeInput(str)
            }
        }
    }
    
    return query.Where(condition, args...)
}

// SafeRaw 安全的原始查询
func (sqb *SafeQueryBuilder) SafeRaw(query *gorm.DB, sql string, values ...interface{}) *gorm.DB {
    // 严格验证原始SQL
    if issues := sqb.protection.ValidateInput(sql); len(issues) > 0 {
        log.Printf("Dangerous raw SQL detected: %v", issues)
        return query.Where("1 = 0")
    }
    
    return query.Raw(sql, values...)
}

// ParameterizedQuery 参数化查询helper
type ParameterizedQuery struct {
    builder *SafeQueryBuilder
}

// NewParameterizedQuery 创建参数化查询
func NewParameterizedQuery(db *gorm.DB) *ParameterizedQuery {
    return &ParameterizedQuery{
        builder: NewSafeQueryBuilder(db),
    }
}

// QueryUsers 安全的用户查询
func (pq *ParameterizedQuery) QueryUsers(filters map[string]interface{}) ([]User, error) {
    var users []User
    query := pq.builder.db.Model(&User{})
    
    // 安全地应用过滤条件
    for field, value := range filters {
        // 白名单字段验证
        allowedFields := map[string]bool{
            "username": true,
            "email":    true,
            "status":   true,
            "role":     true,
        }
        
        if !allowedFields[field] {
            continue // 跳过不允许的字段
        }
        
        // 使用参数化查询
        condition := fmt.Sprintf("%s = ?", field)
        query = pq.builder.SafeWhere(query, condition, value)
    }
    
    err := query.Find(&users).Error
    return users, err
}

// QueryWithPagination 带分页的安全查询
func (pq *ParameterizedQuery) QueryWithPagination(model interface{}, page, pageSize int, orderBy string) error {
    // 验证排序字段
    allowedOrderFields := []string{"id", "created_at", "updated_at", "name", "username"}
    orderField := strings.ToLower(strings.Trim(orderBy, " "))
    
    validOrder := false
    for _, allowed := range allowedOrderFields {
        if orderField == allowed {
            validOrder = true
            break
        }
    }
    
    if !validOrder {
        orderBy = "id" // 默认排序
    }
    
    offset := (page - 1) * pageSize
    
    return pq.builder.db.Model(model).
        Order(orderBy).
        Offset(offset).
        Limit(pageSize).
        Find(model).Error
}

// DatabaseSecurityAuditor 数据库安全审计器
type DatabaseSecurityAuditor struct {
    protection   *SQLInjectionProtection
    logFile      string
    alertChannel chan SecurityAlert
}

// SecurityAlert 安全警报
type SecurityAlert struct {
    Type        string    `json:"type"`
    Message     string    `json:"message"`
    Query       string    `json:"query"`
    ClientIP    string    `json:"client_ip"`
    UserID      uint      `json:"user_id"`
    Timestamp   time.Time `json:"timestamp"`
    Severity    string    `json:"severity"`
}

// NewDatabaseSecurityAuditor 创建数据库安全审计器
func NewDatabaseSecurityAuditor() *DatabaseSecurityAuditor {
    return &DatabaseSecurityAuditor{
        protection:   NewSQLInjectionProtection(),
        logFile:      "/var/log/sql_security.log",
        alertChannel: make(chan SecurityAlert, 100),
    }
}

// AuditQuery 审计查询
func (dsa *DatabaseSecurityAuditor) AuditQuery(query string, clientIP string, userID uint) {
    issues := dsa.protection.ValidateInput(query)
    
    if len(issues) > 0 {
        alert := SecurityAlert{
            Type:      "SQL_INJECTION_ATTEMPT",
            Message:   strings.Join(issues, "; "),
            Query:     query,
            ClientIP:  clientIP,
            UserID:    userID,
            Timestamp: time.Now(),
            Severity:  "HIGH",
        }
        
        // 发送警报
        select {
        case dsa.alertChannel <- alert:
        default:
            // 通道已满，记录到日志
            log.Printf("Alert channel full, logging to file: %+v", alert)
        }
        
        // 记录到审计日志
        dsa.logSecurityEvent(alert)
    }
}

// logSecurityEvent 记录安全事件
func (dsa *DatabaseSecurityAuditor) logSecurityEvent(alert SecurityAlert) {
    // 实现日志记录逻辑
    logEntry := fmt.Sprintf("[%s] %s - %s from %s (User: %d)\n", 
        alert.Timestamp.Format(time.RFC3339), 
        alert.Type, 
        alert.Message, 
        alert.ClientIP, 
        alert.UserID)
    
    // 写入日志文件
    file, err := os.OpenFile(dsa.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Printf("Failed to open security log file: %v", err)
        return
    }
    defer file.Close()
    
    file.WriteString(logEntry)
}

// SQLSecurityMiddleware SQL安全中间件
func SQLSecurityMiddleware() mvc.HandlerFunc {
    auditor := NewDatabaseSecurityAuditor()
    
    return func(c *mvc.Context) {
        // 检查查询参数和表单数据中的SQL注入
        for _, values := range c.Request.URL.Query() {
            for _, value := range values {
                auditor.AuditQuery(value, c.ClientIP(), getCurrentUserID(c))
            }
        }
        
        if c.Request.Method == "POST" {
            c.Request.ParseForm()
            for _, values := range c.Request.PostForm {
                for _, value := range values {
                    auditor.AuditQuery(value, c.ClientIP(), getCurrentUserID(c))
                }
            }
        }
        
        c.Next()
    }
}

func getCurrentUserID(c *mvc.Context) uint {
    if userID, exists := c.Get("user_id"); exists {
        if id, ok := userID.(uint); ok {
            return id
        }
    }
    return 0
}
```

---

## 📝 总结

通过本教程，你已经掌握了YYHertz框架的全面安全防护技能：

### 🎯 核心安全技能
- **身份认证安全** - 多因素认证、密码策略、登录保护
- **数据传输安全** - HTTPS/TLS配置、证书管理、加密通信
- **输入验证防护** - XSS防护、SQL注入防护、数据清理
- **会话管理安全** - 安全会话配置、令牌管理
- **API安全防护** - CSRF防护、限流保护、权限控制
- **数据库安全** - 参数化查询、访问控制、审计日志

### 💡 安全最佳实践
- **纵深防御** - 多层安全防护机制
- **最小权限原则** - 用户和系统权限最小化
- **安全编码** - 安全的代码开发规范
- **定期审计** - 安全日志审计和监控
- **持续更新** - 安全补丁和配置更新

### 🚀 进阶安全方向
- **零信任架构** - 现代化安全架构设计
- **DevSecOps** - 安全集成到CI/CD流程
- **威胁建模** - 系统性安全风险分析
- **渗透测试** - 主动安全测试和评估

---

<div align="center">

**🛡️ 构建YYHertz安全防护体系，保护应用安全！**

**从基础防护到企业级安全，全方位安全解决方案！🔒**

</div>