# YYHertz 用户认证系统教程

<div align="center">

🔐 **完整的用户认证解决方案** | JWT + Session + OAuth2 全覆盖

</div>

---

## 📋 目录

- [认证系统概述](#认证系统概述)
- [JWT认证实现](#jwt认证实现)
- [Session会话管理](#session会话管理)
- [OAuth2社交登录](#oauth2社交登录)
- [权限控制系统](#权限控制系统)
- [多因素认证](#多因素认证)
- [安全最佳实践](#安全最佳实践)
- [实战案例](#实战案例)

---

## 🎯 认证系统概述

### 认证方案对比

| 认证方式 | 适用场景 | 优点 | 缺点 |
|---------|---------|------|------|
| Session + Cookie | 传统Web应用 | 服务端可控、安全性高 | 状态存储、扩展性差 |
| JWT Token | API接口、微服务 | 无状态、跨域友好 | Token泄露风险、难以撤销 |
| OAuth2 | 第三方登录 | 标准协议、用户体验好 | 实现复杂、依赖第三方 |

### 系统架构设计

```
┌─────────────────── 认证层 (Authentication Layer) ───────────────────┐
│  JWT中间件  │  Session中间件  │  OAuth2中间件  │  多因素认证        │
├─────────────────── 授权层 (Authorization Layer) ────────────────────┤
│  RBAC权限控制  │  API权限管理  │  资源访问控制  │  权限缓存策略    │
├─────────────────── 用户管理层 (User Management) ─────────────────────┤
│  用户注册  │  密码管理  │  用户资料  │  账户状态管理  │  安全日志  │
└─────────────────── 数据存储层 (Data Storage) ────────────────────────┘
```

---

## 🔑 JWT认证实现

### 1. JWT工具类

```go
package auth

import (
    "time"
    "crypto/rand"
    "encoding/base64"
    "github.com/golang-jwt/jwt/v4"
    "github.com/zsy619/yyhertz/framework/config"
)

// JWT配置
type JWTConfig struct {
    SecretKey       string
    AccessTokenTTL  time.Duration
    RefreshTokenTTL time.Duration
    Issuer          string
}

// JWT Claims
type JWTClaims struct {
    UserID      uint   `json:"user_id"`
    Username    string `json:"username"`
    Email       string `json:"email"`
    Role        string `json:"role"`
    Permissions []string `json:"permissions,omitempty"`
    jwt.RegisteredClaims
}

// JWT管理器
type JWTManager struct {
    config *JWTConfig
}

func NewJWTManager() *JWTManager {
    return &JWTManager{
        config: &JWTConfig{
            SecretKey:       config.GetString("jwt.secret"),
            AccessTokenTTL:  config.GetDuration("jwt.access_token_ttl", 15*time.Minute),
            RefreshTokenTTL: config.GetDuration("jwt.refresh_token_ttl", 7*24*time.Hour),
            Issuer:          config.GetString("app.name", "YYHertz"),
        },
    }
}

// 生成访问令牌
func (j *JWTManager) GenerateAccessToken(user *User) (string, error) {
    now := time.Now()
    claims := &JWTClaims{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
        Role:     user.GetRole(),
        Permissions: user.GetPermissions(),
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    j.config.Issuer,
            Subject:   fmt.Sprintf("user:%d", user.ID),
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(j.config.AccessTokenTTL)),
            NotBefore: jwt.NewNumericDate(now),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(j.config.SecretKey))
}

// 生成刷新令牌
func (j *JWTManager) GenerateRefreshToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

// 验证访问令牌
func (j *JWTManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(j.config.SecretKey), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid token")
}
```

### 2. JWT认证中间件

```go
// JWT认证中间件
func JWTAuthMiddleware() mvc.HandlerFunc {
    jwtManager := NewJWTManager()
    
    return func(c *mvc.Context) {
        token := extractToken(c)
        if token == "" {
            c.JSON(401, map[string]interface{}{
                "error": "Missing access token",
                "code":  "TOKEN_MISSING",
            })
            c.Abort()
            return
        }
        
        claims, err := jwtManager.ValidateAccessToken(token)
        if err != nil {
            // 判断错误类型
            if ve, ok := err.(*jwt.ValidationError); ok {
                if ve.Errors&jwt.ValidationErrorExpired != 0 {
                    c.JSON(401, map[string]interface{}{
                        "error": "Token expired",
                        "code":  "TOKEN_EXPIRED",
                    })
                } else {
                    c.JSON(401, map[string]interface{}{
                        "error": "Invalid token",
                        "code":  "TOKEN_INVALID",
                    })
                }
            } else {
                c.JSON(401, map[string]interface{}{
                    "error": "Token validation failed",
                    "code":  "TOKEN_VALIDATION_FAILED",
                })
            }
            c.Abort()
            return
        }
        
        // 检查用户状态
        user, err := getUserByID(claims.UserID)
        if err != nil {
            c.JSON(401, map[string]interface{}{
                "error": "User not found",
                "code":  "USER_NOT_FOUND",
            })
            c.Abort()
            return
        }
        
        if !user.IsActive() {
            c.JSON(403, map[string]interface{}{
                "error": "User account is disabled",
                "code":  "USER_DISABLED",
            })
            c.Abort()
            return
        }
        
        // 设置用户信息到上下文
        c.Set("current_user", user)
        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)
        c.Set("user_permissions", claims.Permissions)
        c.Set("token_claims", claims)
        
        c.Next()
    }
}

// 提取Token的多种方式
func extractToken(c *mvc.Context) string {
    // 1. Authorization Header (Bearer Token)
    authHeader := c.GetHeader("Authorization")
    if strings.HasPrefix(authHeader, "Bearer ") {
        return strings.TrimPrefix(authHeader, "Bearer ")
    }
    
    // 2. Query Parameter
    if token := c.Query("access_token"); token != "" {
        return token
    }
    
    // 3. Cookie
    if token, err := c.Cookie("access_token"); err == nil {
        return token
    }
    
    // 4. Custom Header
    if token := c.GetHeader("X-Access-Token"); token != "" {
        return token
    }
    
    return ""
}
```

### 3. JWT认证控制器

```go
package controllers

type AuthController struct {
    mvc.BaseController
    jwtManager *JWTManager
}

// POST /api/auth/login - 用户登录
func (c *AuthController) PostLogin() {
    var req LoginRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "Invalid request format")
        return
    }
    
    // 参数验证
    if req.Username == "" || req.Password == "" {
        c.ErrorJSON(400, "Username and password are required")
        return
    }
    
    // 用户验证
    user, err := c.authService.ValidateCredentials(req.Username, req.Password)
    if err != nil {
        // 记录登录失败
        c.authService.RecordLoginAttempt(req.Username, false, c.ClientIP())
        
        c.ErrorJSON(401, "Invalid credentials")
        return
    }
    
    // 检查账户状态
    if !user.IsActive() {
        c.ErrorJSON(403, "Account is disabled")
        return
    }
    
    // 生成JWT令牌
    accessToken, err := c.jwtManager.GenerateAccessToken(user)
    if err != nil {
        c.ErrorJSON(500, "Failed to generate access token")
        return
    }
    
    refreshToken, err := c.jwtManager.GenerateRefreshToken()
    if err != nil {
        c.ErrorJSON(500, "Failed to generate refresh token")
        return
    }
    
    // 保存refresh token到数据库
    tokenRecord := &RefreshToken{
        UserID:    user.ID,
        Token:     refreshToken,
        ExpiresAt: time.Now().Add(c.jwtManager.config.RefreshTokenTTL),
        CreatedIP: c.ClientIP(),
        UserAgent: c.GetHeader("User-Agent"),
    }
    
    if err := c.authService.SaveRefreshToken(tokenRecord); err != nil {
        c.ErrorJSON(500, "Failed to save refresh token")
        return
    }
    
    // 记录成功登录
    c.authService.RecordLoginAttempt(req.Username, true, c.ClientIP())
    c.authService.UpdateLastLogin(user.ID, c.ClientIP())
    
    // 返回认证结果
    c.JSON(map[string]interface{}{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "token_type":    "Bearer",
        "expires_in":    int(c.jwtManager.config.AccessTokenTTL.Seconds()),
        "user": user.ToPublicResponse(),
    })
}

// POST /api/auth/refresh - 刷新访问令牌
func (c *AuthController) PostRefresh() {
    var req RefreshTokenRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "Invalid request format")
        return
    }
    
    // 验证refresh token
    tokenRecord, err := c.authService.ValidateRefreshToken(req.RefreshToken)
    if err != nil {
        c.ErrorJSON(401, "Invalid refresh token")
        return
    }
    
    // 检查用户状态
    user, err := c.userService.GetUserByID(tokenRecord.UserID)
    if err != nil {
        c.ErrorJSON(401, "User not found")
        return
    }
    
    if !user.IsActive() {
        c.ErrorJSON(403, "User account is disabled")
        return
    }
    
    // 生成新的访问令牌
    newAccessToken, err := c.jwtManager.GenerateAccessToken(user)
    if err != nil {
        c.ErrorJSON(500, "Failed to generate new access token")
        return
    }
    
    // 可选：生成新的refresh token（Token Rotation）
    var newRefreshToken string
    if c.shouldRotateRefreshToken() {
        newRefreshToken, err = c.jwtManager.GenerateRefreshToken()
        if err != nil {
            c.ErrorJSON(500, "Failed to generate new refresh token")
            return
        }
        
        // 更新refresh token记录
        tokenRecord.Token = newRefreshToken
        tokenRecord.UpdatedAt = time.Now()
        c.authService.UpdateRefreshToken(tokenRecord)
    }
    
    response := map[string]interface{}{
        "access_token": newAccessToken,
        "token_type":   "Bearer",
        "expires_in":   int(c.jwtManager.config.AccessTokenTTL.Seconds()),
    }
    
    if newRefreshToken != "" {
        response["refresh_token"] = newRefreshToken
    }
    
    c.JSON(response)
}

// POST /api/auth/logout - 用户登出
func (c *AuthController) PostLogout() {
    var req LogoutRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "Invalid request format")
        return
    }
    
    // 将refresh token加入黑名单
    if req.RefreshToken != "" {
        c.authService.RevokeRefreshToken(req.RefreshToken)
    }
    
    // 将当前access token加入黑名单（可选）
    if token := extractToken(c.Context); token != "" {
        c.authService.BlacklistAccessToken(token)
    }
    
    c.JSON(map[string]interface{}{
        "message": "Logged out successfully",
    })
}

// GET /api/auth/me - 获取当前用户信息
func (c *AuthController) GetMe() {
    user := c.MustGet("current_user").(*User)
    c.JSON(user.ToPublicResponse())
}
```

---

## 🍪 Session会话管理

### 1. Session存储实现

```go
package session

import (
    "time"
    "encoding/json"
    "crypto/rand"
    "encoding/base64"
)

// Session接口
type SessionStore interface {
    Get(sessionID string) (*SessionData, error)
    Set(sessionID string, data *SessionData) error
    Delete(sessionID string) error
    Regenerate(oldID string) (string, error)
    Cleanup() error
}

// Session数据结构
type SessionData struct {
    UserID      uint                   `json:"user_id,omitempty"`
    Username    string                 `json:"username,omitempty"`
    IsLoggedIn  bool                   `json:"is_logged_in"`
    Data        map[string]interface{} `json:"data"`
    CreatedAt   time.Time              `json:"created_at"`
    LastAccess  time.Time              `json:"last_access"`
    ExpiresAt   time.Time              `json:"expires_at"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent"`
}

// Redis Session存储
type RedisSessionStore struct {
    client     *redis.Client
    keyPrefix  string
    maxAge     time.Duration
}

func NewRedisSessionStore(client *redis.Client) *RedisSessionStore {
    return &RedisSessionStore{
        client:    client,
        keyPrefix: "session:",
        maxAge:    24 * time.Hour, // 默认24小时
    }
}

func (s *RedisSessionStore) Get(sessionID string) (*SessionData, error) {
    key := s.keyPrefix + sessionID
    data, err := s.client.Get(context.Background(), key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, ErrSessionNotFound
        }
        return nil, err
    }
    
    var sessionData SessionData
    if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
        return nil, err
    }
    
    // 检查是否过期
    if time.Now().After(sessionData.ExpiresAt) {
        s.Delete(sessionID)
        return nil, ErrSessionExpired
    }
    
    // 更新最后访问时间
    sessionData.LastAccess = time.Now()
    s.Set(sessionID, &sessionData)
    
    return &sessionData, nil
}

func (s *RedisSessionStore) Set(sessionID string, data *SessionData) error {
    key := s.keyPrefix + sessionID
    
    // 设置过期时间
    if data.ExpiresAt.IsZero() {
        data.ExpiresAt = time.Now().Add(s.maxAge)
    }
    
    jsonData, err := json.Marshal(data)
    if err != nil {
        return err
    }
    
    return s.client.Set(context.Background(), key, jsonData, s.maxAge).Err()
}

func (s *RedisSessionStore) Delete(sessionID string) error {
    key := s.keyPrefix + sessionID
    return s.client.Del(context.Background(), key).Err()
}

func (s *RedisSessionStore) Regenerate(oldID string) (string, error) {
    // 获取旧session数据
    oldData, err := s.Get(oldID)
    if err != nil {
        return "", err
    }
    
    // 生成新session ID
    newID, err := generateSessionID()
    if err != nil {
        return "", err
    }
    
    // 保存到新session
    if err := s.Set(newID, oldData); err != nil {
        return "", err
    }
    
    // 删除旧session
    s.Delete(oldID)
    
    return newID, nil
}

// 生成Session ID
func generateSessionID() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}
```

### 2. Session中间件

```go
// Session配置
type SessionConfig struct {
    Store      SessionStore
    CookieName string
    Domain     string
    Path       string
    MaxAge     time.Duration
    Secure     bool
    HttpOnly   bool
    SameSite   http.SameSite
}

// Session中间件
func SessionMiddleware(config SessionConfig) mvc.HandlerFunc {
    if config.CookieName == "" {
        config.CookieName = "session_id"
    }
    if config.Path == "" {
        config.Path = "/"
    }
    if config.MaxAge == 0 {
        config.MaxAge = 24 * time.Hour
    }
    
    return func(c *mvc.Context) {
        // 获取session ID
        sessionID, err := c.Cookie(config.CookieName)
        if err != nil || sessionID == "" {
            // 创建新session
            sessionID, err = generateSessionID()
            if err != nil {
                c.JSON(500, map[string]interface{}{
                    "error": "Failed to generate session",
                })
                c.Abort()
                return
            }
            
            // 设置cookie
            c.SetCookie(&http.Cookie{
                Name:     config.CookieName,
                Value:    sessionID,
                Domain:   config.Domain,
                Path:     config.Path,
                MaxAge:   int(config.MaxAge.Seconds()),
                Secure:   config.Secure,
                HttpOnly: config.HttpOnly,
                SameSite: config.SameSite,
            })
            
            // 创建新session数据
            sessionData := &SessionData{
                IsLoggedIn: false,
                Data:       make(map[string]interface{}),
                CreatedAt:  time.Now(),
                LastAccess: time.Now(),
                ExpiresAt:  time.Now().Add(config.MaxAge),
                IPAddress:  c.ClientIP(),
                UserAgent:  c.GetHeader("User-Agent"),
            }
            
            config.Store.Set(sessionID, sessionData)
            c.Set("session", sessionData)
            c.Set("session_id", sessionID)
        } else {
            // 获取已存在的session
            sessionData, err := config.Store.Get(sessionID)
            if err != nil {
                // Session不存在或已过期，创建新session
                sessionID, _ = generateSessionID()
                sessionData = &SessionData{
                    IsLoggedIn: false,
                    Data:       make(map[string]interface{}),
                    CreatedAt:  time.Now(),
                    LastAccess: time.Now(),
                    ExpiresAt:  time.Now().Add(config.MaxAge),
                    IPAddress:  c.ClientIP(),
                    UserAgent:  c.GetHeader("User-Agent"),
                }
                
                // 更新cookie
                c.SetCookie(&http.Cookie{
                    Name:     config.CookieName,
                    Value:    sessionID,
                    Domain:   config.Domain,
                    Path:     config.Path,
                    MaxAge:   int(config.MaxAge.Seconds()),
                    Secure:   config.Secure,
                    HttpOnly: config.HttpOnly,
                    SameSite: config.SameSite,
                })
            }
            
            c.Set("session", sessionData)
            c.Set("session_id", sessionID)
        }
        
        c.Next()
        
        // 请求处理完成后保存session
        if sessionData := c.Get("session"); sessionData != nil {
            config.Store.Set(sessionID, sessionData.(*SessionData))
        }
    }
}
```

### 3. Session认证控制器

```go
// Session认证控制器
type SessionAuthController struct {
    mvc.BaseController
}

// POST /auth/login - Session登录
func (c *SessionAuthController) PostLogin() {
    var req LoginRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "Invalid request format")
        return
    }
    
    // 验证用户凭证
    user, err := c.authService.ValidateCredentials(req.Username, req.Password)
    if err != nil {
        c.ErrorJSON(401, "Invalid credentials")
        return
    }
    
    // 获取session
    session := c.MustGet("session").(*SessionData)
    sessionID := c.GetString("session_id")
    
    // 登录成功，regenerate session ID防止session fixation攻击
    newSessionID, err := c.sessionStore.Regenerate(sessionID)
    if err != nil {
        c.ErrorJSON(500, "Session regeneration failed")
        return
    }
    
    // 更新session数据
    session.UserID = user.ID
    session.Username = user.Username
    session.IsLoggedIn = true
    session.Data["role"] = user.GetRole()
    session.Data["permissions"] = user.GetPermissions()
    session.LastAccess = time.Now()
    
    // 更新cookie
    c.SetCookie(&http.Cookie{
        Name:     "session_id",
        Value:    newSessionID,
        Path:     "/",
        MaxAge:   int(24 * time.Hour.Seconds()),
        HttpOnly: true,
        Secure:   c.IsHTTPS(),
        SameSite: http.SameSiteStrictMode,
    })
    
    // 记录登录日志
    c.authService.RecordLoginAttempt(req.Username, true, c.ClientIP())
    c.authService.UpdateLastLogin(user.ID, c.ClientIP())
    
    c.JSON(map[string]interface{}{
        "message": "Login successful",
        "user":    user.ToPublicResponse(),
    })
}

// POST /auth/logout - Session登出
func (c *SessionAuthController) PostLogout() {
    sessionID := c.GetString("session_id")
    session := c.MustGet("session").(*SessionData)
    
    // 清除session数据
    session.UserID = 0
    session.Username = ""
    session.IsLoggedIn = false
    session.Data = make(map[string]interface{})
    
    // 或者直接删除session
    c.sessionStore.Delete(sessionID)
    
    // 清除cookie
    c.SetCookie(&http.Cookie{
        Name:     "session_id",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        HttpOnly: true,
    })
    
    c.JSON(map[string]interface{}{
        "message": "Logout successful",
    })
}

// GET /auth/check - 检查认证状态
func (c *SessionAuthController) GetCheck() {
    session := c.MustGet("session").(*SessionData)
    
    if !session.IsLoggedIn {
        c.JSON(401, map[string]interface{}{
            "authenticated": false,
            "message":       "Not authenticated",
        })
        return
    }
    
    // 获取用户信息
    user, err := c.userService.GetUserByID(session.UserID)
    if err != nil {
        c.JSON(401, map[string]interface{}{
            "authenticated": false,
            "message":       "User not found",
        })
        return
    }
    
    c.JSON(map[string]interface{}{
        "authenticated": true,
        "user":          user.ToPublicResponse(),
        "session": map[string]interface{}{
            "created_at":  session.CreatedAt,
            "last_access": session.LastAccess,
            "expires_at":  session.ExpiresAt,
        },
    })
}

// Session认证中间件
func SessionAuthRequired() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        session := c.MustGet("session").(*SessionData)
        
        if !session.IsLoggedIn {
            c.JSON(401, map[string]interface{}{
                "error": "Authentication required",
            })
            c.Abort()
            return
        }
        
        // 验证用户是否仍然存在且激活
        user, err := getUserByID(session.UserID)
        if err != nil || !user.IsActive() {
            // 用户不存在或已停用，清除session
            session.IsLoggedIn = false
            session.UserID = 0
            
            c.JSON(401, map[string]interface{}{
                "error": "User account is disabled",
            })
            c.Abort()
            return
        }
        
        c.Set("current_user", user)
        c.Next()
    }
}
```

---

## 🌐 OAuth2社交登录

### 1. OAuth2客户端实现

```go
package oauth2

import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "golang.org/x/oauth2/github"
)

// OAuth2配置管理器
type OAuth2Manager struct {
    providers map[string]*oauth2.Config
}

func NewOAuth2Manager() *OAuth2Manager {
    manager := &OAuth2Manager{
        providers: make(map[string]*oauth2.Config),
    }
    
    // 配置Google OAuth2
    manager.providers["google"] = &oauth2.Config{
        ClientID:     config.GetString("oauth2.google.client_id"),
        ClientSecret: config.GetString("oauth2.google.client_secret"),
        RedirectURL:  config.GetString("oauth2.google.redirect_url"),
        Scopes: []string{
            "https://www.googleapis.com/auth/userinfo.email",
            "https://www.googleapis.com/auth/userinfo.profile",
        },
        Endpoint: google.Endpoint,
    }
    
    // 配置GitHub OAuth2
    manager.providers["github"] = &oauth2.Config{
        ClientID:     config.GetString("oauth2.github.client_id"),
        ClientSecret: config.GetString("oauth2.github.client_secret"),
        RedirectURL:  config.GetString("oauth2.github.redirect_url"),
        Scopes:       []string{"user:email"},
        Endpoint:     github.Endpoint,
    }
    
    // 配置微信OAuth2
    manager.providers["wechat"] = &oauth2.Config{
        ClientID:     config.GetString("oauth2.wechat.app_id"),
        ClientSecret: config.GetString("oauth2.wechat.app_secret"),
        RedirectURL:  config.GetString("oauth2.wechat.redirect_url"),
        Scopes:       []string{"snsapi_userinfo"},
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://open.weixin.qq.com/connect/qrconnect",
            TokenURL: "https://api.weixin.qq.com/sns/oauth2/access_token",
        },
    }
    
    return manager
}

// 获取授权URL
func (m *OAuth2Manager) GetAuthURL(provider, state string) (string, error) {
    config, exists := m.providers[provider]
    if !exists {
        return "", fmt.Errorf("unsupported OAuth2 provider: %s", provider)
    }
    
    return config.AuthCodeURL(state), nil
}

// 交换授权码获取用户信息
func (m *OAuth2Manager) ExchangeCodeForUser(provider, code, state string) (*OAuth2UserInfo, error) {
    config, exists := m.providers[provider]
    if !exists {
        return nil, fmt.Errorf("unsupported OAuth2 provider: %s", provider)
    }
    
    // 验证state参数防止CSRF攻击
    if !m.validateState(state) {
        return nil, fmt.Errorf("invalid state parameter")
    }
    
    // 交换授权码获取访问令牌
    token, err := config.Exchange(context.Background(), code)
    if err != nil {
        return nil, err
    }
    
    // 根据提供商获取用户信息
    switch provider {
    case "google":
        return m.getGoogleUserInfo(token)
    case "github":
        return m.getGitHubUserInfo(token)
    case "wechat":
        return m.getWeChatUserInfo(token)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", provider)
    }
}

// OAuth2用户信息结构
type OAuth2UserInfo struct {
    Provider     string `json:"provider"`
    ProviderID   string `json:"provider_id"`
    Email        string `json:"email"`
    Name         string `json:"name"`
    Avatar       string `json:"avatar"`
    Username     string `json:"username"`
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

// 获取Google用户信息
func (m *OAuth2Manager) getGoogleUserInfo(token *oauth2.Token) (*OAuth2UserInfo, error) {
    client := m.providers["google"].Client(context.Background(), token)
    
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var googleUser struct {
        ID      string `json:"id"`
        Email   string `json:"email"`
        Name    string `json:"name"`
        Picture string `json:"picture"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
        return nil, err
    }
    
    return &OAuth2UserInfo{
        Provider:     "google",
        ProviderID:   googleUser.ID,
        Email:        googleUser.Email,
        Name:         googleUser.Name,
        Avatar:       googleUser.Picture,
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
    }, nil
}

// 获取GitHub用户信息
func (m *OAuth2Manager) getGitHubUserInfo(token *oauth2.Token) (*OAuth2UserInfo, error) {
    client := m.providers["github"].Client(context.Background(), token)
    
    // 获取用户基本信息
    resp, err := client.Get("https://api.github.com/user")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var githubUser struct {
        ID       int    `json:"id"`
        Login    string `json:"login"`
        Name     string `json:"name"`
        Email    string `json:"email"`
        AvatarURL string `json:"avatar_url"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
        return nil, err
    }
    
    // 如果公开邮箱为空，获取主邮箱
    if githubUser.Email == "" {
        emailResp, err := client.Get("https://api.github.com/user/emails")
        if err == nil {
            defer emailResp.Body.Close()
            
            var emails []struct {
                Email   string `json:"email"`
                Primary bool   `json:"primary"`
            }
            
            if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
                for _, email := range emails {
                    if email.Primary {
                        githubUser.Email = email.Email
                        break
                    }
                }
            }
        }
    }
    
    return &OAuth2UserInfo{
        Provider:     "github",
        ProviderID:   fmt.Sprintf("%d", githubUser.ID),
        Email:        githubUser.Email,
        Name:         githubUser.Name,
        Username:     githubUser.Login,
        Avatar:       githubUser.AvatarURL,
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
    }, nil
}
```

### 2. OAuth2认证控制器

```go
// OAuth2认证控制器
type OAuth2Controller struct {
    mvc.BaseController
    oauth2Manager *OAuth2Manager
}

// GET /auth/oauth2/{provider} - 获取OAuth2授权URL
func (c *OAuth2Controller) GetAuth() {
    provider := c.Param("provider")
    if provider == "" {
        c.ErrorJSON(400, "Provider is required")
        return
    }
    
    // 生成state参数防止CSRF
    state := c.generateState()
    c.saveState(state) // 保存到session或缓存
    
    // 获取授权URL
    authURL, err := c.oauth2Manager.GetAuthURL(provider, state)
    if err != nil {
        c.ErrorJSON(400, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "auth_url": authURL,
        "state":    state,
    })
}

// GET /auth/oauth2/{provider}/callback - OAuth2回调处理
func (c *OAuth2Controller) GetCallback() {
    provider := c.Param("provider")
    code := c.Query("code")
    state := c.Query("state")
    errorParam := c.Query("error")
    
    // 检查是否有错误
    if errorParam != "" {
        c.ErrorJSON(400, fmt.Sprintf("OAuth2 error: %s", errorParam))
        return
    }
    
    if code == "" {
        c.ErrorJSON(400, "Authorization code is required")
        return
    }
    
    // 验证state参数
    if !c.validateState(state) {
        c.ErrorJSON(400, "Invalid state parameter")
        return
    }
    
    // 交换授权码获取用户信息
    userInfo, err := c.oauth2Manager.ExchangeCodeForUser(provider, code, state)
    if err != nil {
        c.ErrorJSON(500, fmt.Sprintf("Failed to get user info: %v", err))
        return
    }
    
    // 查找或创建用户
    user, isNewUser, err := c.findOrCreateOAuth2User(userInfo)
    if err != nil {
        c.ErrorJSON(500, fmt.Sprintf("Failed to process user: %v", err))
        return
    }
    
    // 生成认证令牌
    jwtManager := NewJWTManager()
    accessToken, err := jwtManager.GenerateAccessToken(user)
    if err != nil {
        c.ErrorJSON(500, "Failed to generate access token")
        return
    }
    
    refreshToken, err := jwtManager.GenerateRefreshToken()
    if err != nil {
        c.ErrorJSON(500, "Failed to generate refresh token")
        return
    }
    
    // 保存refresh token
    tokenRecord := &RefreshToken{
        UserID:    user.ID,
        Token:     refreshToken,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
        CreatedIP: c.ClientIP(),
        UserAgent: c.GetHeader("User-Agent"),
    }
    c.authService.SaveRefreshToken(tokenRecord)
    
    // 记录OAuth2登录
    c.authService.RecordOAuth2Login(user.ID, provider, userInfo.ProviderID)
    
    response := map[string]interface{}{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "token_type":    "Bearer",
        "expires_in":    int(15 * time.Minute.Seconds()),
        "user":          user.ToPublicResponse(),
        "is_new_user":   isNewUser,
    }
    
    // 如果是Web应用，可以重定向到前端
    if c.Query("redirect_type") == "web" {
        // 将token信息编码到URL参数或使用其他安全方式传递
        redirectURL := fmt.Sprintf("%s?token=%s", 
            config.GetString("oauth2.success_redirect_url"), 
            accessToken)
        c.Redirect(redirectURL, 302)
        return
    }
    
    c.JSON(response)
}

// 查找或创建OAuth2用户
func (c *OAuth2Controller) findOrCreateOAuth2User(userInfo *OAuth2UserInfo) (*User, bool, error) {
    // 首先通过OAuth2提供商ID查找
    socialAccount, err := c.authService.GetSocialAccount(userInfo.Provider, userInfo.ProviderID)
    if err == nil && socialAccount != nil {
        // 找到已关联的社交账户
        user, err := c.userService.GetUserByID(socialAccount.UserID)
        if err != nil {
            return nil, false, err
        }
        
        // 更新社交账户信息
        socialAccount.AccessToken = userInfo.AccessToken
        socialAccount.RefreshToken = userInfo.RefreshToken
        socialAccount.UpdatedAt = time.Now()
        c.authService.UpdateSocialAccount(socialAccount)
        
        return user, false, nil
    }
    
    // 通过邮箱查找已存在用户
    if userInfo.Email != "" {
        existingUser, err := c.userService.GetUserByEmail(userInfo.Email)
        if err == nil && existingUser != nil {
            // 用户已存在，创建社交账户关联
            socialAccount := &SocialAccount{
                UserID:       existingUser.ID,
                Provider:     userInfo.Provider,
                ProviderID:   userInfo.ProviderID,
                AccessToken:  userInfo.AccessToken,
                RefreshToken: userInfo.RefreshToken,
                CreatedAt:    time.Now(),
                UpdatedAt:    time.Now(),
            }
            
            if err := c.authService.CreateSocialAccount(socialAccount); err != nil {
                return nil, false, err
            }
            
            return existingUser, false, nil
        }
    }
    
    // 创建新用户
    username := userInfo.Username
    if username == "" {
        username = c.generateUsernameFromEmail(userInfo.Email)
    }
    
    newUser := &User{
        Username:  c.ensureUniqueUsername(username),
        Email:     userInfo.Email,
        Name:      userInfo.Name,
        Avatar:    userInfo.Avatar,
        Status:    UserStatusActive,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    if err := c.userService.CreateUser(newUser); err != nil {
        return nil, false, err
    }
    
    // 创建社交账户关联
    socialAccount = &SocialAccount{
        UserID:       newUser.ID,
        Provider:     userInfo.Provider,
        ProviderID:   userInfo.ProviderID,
        AccessToken:  userInfo.AccessToken,
        RefreshToken: userInfo.RefreshToken,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
    
    if err := c.authService.CreateSocialAccount(socialAccount); err != nil {
        return nil, false, err
    }
    
    return newUser, true, nil
}

// 社交账户模型
type SocialAccount struct {
    ID           uint      `gorm:"primarykey"`
    UserID       uint      `gorm:"not null;index"`
    Provider     string    `gorm:"size:50;not null"`
    ProviderID   string    `gorm:"size:100;not null"`
    AccessToken  string    `gorm:"type:text"`
    RefreshToken string    `gorm:"type:text"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    
    User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// 唯一索引
func (SocialAccount) TableName() string {
    return "social_accounts"
}
```

---

## 🛡️ 权限控制系统

### 1. RBAC权限模型

```go
// 角色模型
type Role struct {
    ID          uint         `gorm:"primarykey" json:"id"`
    Name        string       `gorm:"size:50;uniqueIndex;not null" json:"name"`
    DisplayName string       `gorm:"size:100" json:"display_name"`
    Description string       `gorm:"type:text" json:"description"`
    IsSystem    bool         `gorm:"default:false" json:"is_system"` // 系统角色不可删除
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
    
    // 关联关系
    Users       []User       `gorm:"many2many:user_roles" json:"users,omitempty"`
    Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

// 权限模型
type Permission struct {
    ID          uint      `gorm:"primarykey" json:"id"`
    Name        string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
    DisplayName string    `gorm:"size:100" json:"display_name"`
    Description string    `gorm:"type:text" json:"description"`
    Resource    string    `gorm:"size:50;not null;index" json:"resource"` // 资源名称
    Action      string    `gorm:"size:50;not null;index" json:"action"`   // 操作类型
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // 关联关系
    Roles []Role `gorm:"many2many:role_permissions" json:"roles,omitempty"`
}

// 用户角色关联
type UserRole struct {
    UserID    uint      `gorm:"primarykey"`
    RoleID    uint      `gorm:"primarykey"`
    CreatedAt time.Time
    
    User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
    Role Role `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// 角色权限关联
type RolePermission struct {
    RoleID       uint `gorm:"primarykey"`
    PermissionID uint `gorm:"primarykey"`
    
    Role       Role       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
    Permission Permission `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// 用户权限扩展方法
func (u *User) GetRoles() []Role {
    var roles []Role
    db := orm.GetDB()
    db.Model(u).Association("Roles").Find(&roles)
    return roles
}

func (u *User) GetPermissions() []string {
    var permissions []string
    roles := u.GetRoles()
    
    permissionSet := make(map[string]bool)
    for _, role := range roles {
        var rolePerms []Permission
        db := orm.GetDB()
        db.Model(&role).Association("Permissions").Find(&rolePerms)
        
        for _, perm := range rolePerms {
            permissionSet[perm.Name] = true
        }
    }
    
    for perm := range permissionSet {
        permissions = append(permissions, perm)
    }
    
    return permissions
}

func (u *User) HasRole(roleName string) bool {
    roles := u.GetRoles()
    for _, role := range roles {
        if role.Name == roleName {
            return true
        }
    }
    return false
}

func (u *User) HasPermission(permissionName string) bool {
    permissions := u.GetPermissions()
    for _, perm := range permissions {
        if perm == permissionName {
            return true
        }
    }
    return false
}

func (u *User) HasAnyPermission(permissionNames ...string) bool {
    userPermissions := u.GetPermissions()
    userPermSet := make(map[string]bool)
    for _, perm := range userPermissions {
        userPermSet[perm] = true
    }
    
    for _, perm := range permissionNames {
        if userPermSet[perm] {
            return true
        }
    }
    return false
}
```

### 2. 权限检查中间件

```go
// 角色权限中间件
func RequireRoles(roles ...string) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        user := getCurrentUser(c)
        if user == nil {
            c.JSON(401, map[string]interface{}{
                "error": "Authentication required",
            })
            c.Abort()
            return
        }
        
        // 检查是否有任一所需角色
        hasRole := false
        for _, requiredRole := range roles {
            if user.HasRole(requiredRole) {
                hasRole = true
                break
            }
        }
        
        if !hasRole {
            c.JSON(403, map[string]interface{}{
                "error":          "Insufficient privileges",
                "required_roles": roles,
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 权限中间件
func RequirePermissions(permissions ...string) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        user := getCurrentUser(c)
        if user == nil {
            c.JSON(401, map[string]interface{}{
                "error": "Authentication required",
            })
            c.Abort()
            return
        }
        
        // 检查是否有任一所需权限
        if !user.HasAnyPermission(permissions...) {
            c.JSON(403, map[string]interface{}{
                "error":               "Insufficient privileges",
                "required_permissions": permissions,
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 资源所有者权限中间件
func RequireOwnershipOr(permissions ...string) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        user := getCurrentUser(c)
        if user == nil {
            c.JSON(401, map[string]interface{}{
                "error": "Authentication required",
            })
            c.Abort()
            return
        }
        
        // 获取资源所有者ID
        resourceOwnerID := c.getResourceOwnerID() // 需要实现此方法
        
        // 如果是资源所有者或有管理权限，则允许访问
        if user.ID == resourceOwnerID || user.HasAnyPermission(permissions...) {
            c.Next()
            return
        }
        
        c.JSON(403, map[string]interface{}{
            "error": "Access denied",
        })
        c.Abort()
    }
}

// 获取当前用户
func getCurrentUser(c *mvc.Context) *User {
    if user := c.Get("current_user"); user != nil {
        return user.(*User)
    }
    return nil
}
```

### 3. 权限管理API

```go
// 权限管理控制器
type PermissionController struct {
    mvc.BaseController
}

// GET /api/permissions - 获取权限列表
func (c *PermissionController) GetIndex() {
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    resource := c.GetString("resource")
    
    filters := &PermissionFilters{
        Resource: resource,
    }
    
    permissions, total, err := c.permissionService.GetPermissionList(filters, page, pageSize)
    if err != nil {
        c.ErrorJSON(500, "Failed to get permissions")
        return
    }
    
    meta := &Meta{
        Page:       page,
        PageSize:   pageSize,
        Total:      total,
        TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
    }
    
    c.SuccessWithPagination(permissions, meta)
}

// POST /api/permissions - 创建权限
func (c *PermissionController) PostStore() {
    var req CreatePermissionRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "Invalid request format")
        return
    }
    
    if errors := validator.Validate(req); len(errors) > 0 {
        c.ValidationErrorJSON(errors)
        return
    }
    
    permission, err := c.permissionService.CreatePermission(&req)
    if err != nil {
        c.ErrorJSON(500, "Failed to create permission")
        return
    }
    
    c.Success(permission)
}

// GET /api/roles - 获取角色列表
func (c *RoleController) GetIndex() {
    roles, err := c.roleService.GetAllRoles()
    if err != nil {
        c.ErrorJSON(500, "Failed to get roles")
        return
    }
    
    c.Success(roles)
}

// POST /api/roles - 创建角色
func (c *RoleController) PostStore() {
    var req CreateRoleRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "Invalid request format")
        return
    }
    
    role, err := c.roleService.CreateRole(&req)
    if err != nil {
        c.ErrorJSON(500, "Failed to create role")
        return
    }
    
    c.Success(role)
}

// PUT /api/roles/:id/permissions - 分配权限给角色
func (c *RoleController) PutPermissions() {
    roleID, err := c.GetInt("id")
    if err != nil {
        c.ErrorJSON(400, "Invalid role ID")
        return
    }
    
    var req AssignPermissionsRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "Invalid request format")
        return
    }
    
    err = c.roleService.AssignPermissions(uint(roleID), req.PermissionIDs)
    if err != nil {
        c.ErrorJSON(500, "Failed to assign permissions")
        return
    }
    
    c.Success(map[string]interface{}{
        "message": "Permissions assigned successfully",
    })
}

// PUT /api/users/:id/roles - 分配角色给用户
func (c *UserController) PutRoles() {
    userID, err := c.GetInt("id")
    if err != nil {
        c.ErrorJSON(400, "Invalid user ID")
        return
    }
    
    var req AssignRolesRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "Invalid request format")
        return
    }
    
    err = c.userService.AssignRoles(uint(userID), req.RoleIDs)
    if err != nil {
        c.ErrorJSON(500, "Failed to assign roles")
        return
    }
    
    c.Success(map[string]interface{}{
        "message": "Roles assigned successfully",
    })
}
```

---

<div align="center">

**🔐 安全的认证系统是现代应用的基石！**

**选择合适的认证方案，构建可靠的权限体系 🛡️**

</div>