package unified

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
)

// SSOConfig SSO配置
type SSOConfig struct {
	// 基础配置
	Enabled         bool     `json:"enabled" yaml:"enabled"`                   // 是否启用SSO
	Secret          string   `json:"secret" yaml:"secret"`                     // 签名密钥
	TokenExpireTime int64    `json:"token_expire_time" yaml:"token_expire_time"` // Token过期时间（秒）
	
	// Cookie配置
	CookieName       string `json:"cookie_name" yaml:"cookie_name"`             // SSO Token Cookie名称
	CookieDomain     string `json:"cookie_domain" yaml:"cookie_domain"`         // Cookie域名
	CookiePath       string `json:"cookie_path" yaml:"cookie_path"`             // Cookie路径
	CookieSecure     bool   `json:"cookie_secure" yaml:"cookie_secure"`         // 是否仅HTTPS
	CookieHttpOnly   bool   `json:"cookie_http_only" yaml:"cookie_http_only"`   // 是否HttpOnly
	
	// 记住我功能
	RememberCookieName string `json:"remember_cookie_name" yaml:"remember_cookie_name"` // 记住我Cookie名称
	RememberExpireTime int64  `json:"remember_expire_time" yaml:"remember_expire_time"` // 记住我过期时间（秒）
	
	// 排除路径（这些路径不需要SSO验证）
	ExcludePaths []string `json:"exclude_paths" yaml:"exclude_paths"`
	
	// 登录页面
	LoginURL    string `json:"login_url" yaml:"login_url"`       // 登录页面URL
	LogoutURL   string `json:"logout_url" yaml:"logout_url"`     // 登出页面URL
	
	// 自定义验证函数
	TokenValidator   func(token *SSOToken) (*UserInfo, error) `json:"-" yaml:"-"`
	RememberValidator func(rememberToken string) (*UserInfo, error) `json:"-" yaml:"-"`
}

// DefaultSSOConfig 默认SSO配置
func DefaultSSOConfig() *SSOConfig {
	return &SSOConfig{
		Enabled:            true,
		Secret:             "default-sso-secret-change-me",
		TokenExpireTime:    3600, // 1小时
		CookieName:         "sso_token",
		CookieDomain:       "",
		CookiePath:         "/",
		CookieSecure:       false,
		CookieHttpOnly:     true,
		RememberCookieName: "remember_token",
		RememberExpireTime: 7 * 24 * 3600, // 7天
		ExcludePaths: []string{
			"/login", "/register", "/forgot-password",
			"/api/public/*", "/static/*", "/health",
		},
		LoginURL:  "/login",
		LogoutURL: "/logout",
	}
}

// SSOToken SSO令牌结构
type SSOToken struct {
	UserID    string    `json:"user_id"`    // 用户ID
	Username  string    `json:"username"`   // 用户名
	Email     string    `json:"email"`      // 邮箱
	Roles     []string  `json:"roles"`      // 用户角色
	IssuedAt  time.Time `json:"issued_at"`  // 签发时间
	ExpiresAt time.Time `json:"expires_at"` // 过期时间
	ClientIP  string    `json:"client_ip"`  // 客户端IP
	UserAgent string    `json:"user_agent"` // 用户代理
}

// UserInfo 用户信息结构
type UserInfo struct {
	ID       string                 `json:"id"`       // 用户ID
	Username string                 `json:"username"` // 用户名
	Email    string                 `json:"email"`    // 邮箱
	Roles    []string               `json:"roles"`    // 用户角色
	Profile  map[string]interface{} `json:"profile"`  // 用户资料
}

// 全局SSO配置
var globalSSOConfig *SSOConfig

// SetSSOConfig 设置全局SSO配置
func SetSSOConfig(config *SSOConfig) {
	globalSSOConfig = config
}

// GetSSOConfig 获取全局SSO配置
func GetSSOConfig() *SSOConfig {
	if globalSSOConfig == nil {
		globalSSOConfig = DefaultSSOConfig()
	}
	return globalSSOConfig
}

// FilterSSO SSO过滤器实现
//
// 这是一个全局SSO过滤器，用于检查用户的SSO状态。
// 它会检查Session中的SSO令牌，如果无效则尝试从Cookie中恢复会话。
//
// 执行流程：
// 1. 检查路径是否需要SSO验证（排除路径跳过）
// 2. 从Session中获取SSO令牌并验证
// 3. 如果Session无效，尝试从Cookie中恢复
// 4. 如果都无效，根据配置决定是否重定向到登录页
//
// 参数：
//   - ctx: 请求上下文
//
// 返回：
//   - FilterResult: 过滤结果
//   - error: 执行错误
var FilterSSO FilterFunc = func(ctx *context.Context) (FilterResult, error) {
	config := GetSSOConfig()
	
	// 检查SSO是否启用
	if !config.Enabled {
		return FilterContinue, nil
	}
	
	// 获取请求路径
	path := string(ctx.Request().URI().Path())
	
	// 检查是否在排除路径中
	if isExcludedPath(config, path) {
		return FilterContinue, nil
	}
	
	// 获取统一管理器
	manager := GetManager()
	
	// 1. 首先检查Session中的SSO令牌
	if userInfo := checkSessionSSO(manager, ctx); userInfo != nil {
		// Session中有有效的用户信息，设置到上下文
		manager.SetContextData(ctx, "user", userInfo)
		manager.SetContextData(ctx, "authenticated", true)
		return FilterContinue, nil
	}
	
	// 2. 检查Cookie中的SSO令牌
	if userInfo := checkCookieSSO(manager, ctx, config); userInfo != nil {
		// Cookie中有有效令牌，恢复到Session并设置到上下文
		manager.SetSessionData(ctx, "user", userInfo)
		manager.SetSessionData(ctx, "sso_token", time.Now().Unix())
		manager.SetContextData(ctx, "user", userInfo)
		manager.SetContextData(ctx, "authenticated", true)
		return FilterContinue, nil
	}
	
	// 3. 检查记住我功能
	if userInfo := checkRememberMe(manager, ctx, config); userInfo != nil {
		// 记住我令牌有效，恢复会话
		manager.SetSessionData(ctx, "user", userInfo)
		manager.SetSessionData(ctx, "sso_token", time.Now().Unix())
		manager.SetContextData(ctx, "user", userInfo)
		manager.SetContextData(ctx, "authenticated", true)
		
		// 重新设置SSO Cookie
		if err := setSSOCookie(manager, ctx, config, userInfo); err != nil {
			// 记录错误但不阻断请求
			fmt.Printf("Warning: Failed to set SSO cookie: %v\n", err)
		}
		
		return FilterContinue, nil
	}
	
	// 4. 所有验证都失败，用户未认证
	manager.SetContextData(ctx, "authenticated", false)
	
	// 对于API请求，返回401错误而不是重定向
	if isAPIRequest(path) {
		ctx.Request().SetStatusCode(401)
		ctx.Request().Response.SetBodyString(`{"error":"authentication required","code":401}`)
		return FilterStop, nil
	}
	
	// 对于页面请求，重定向到登录页
	if config.LoginURL != "" {
		// 保存原始请求URL
		originalURL := string(ctx.Request().URI().RequestURI())
		loginURL := fmt.Sprintf("%s?redirect=%s", config.LoginURL, originalURL)
		
		ctx.Request().Response.Header.Set("Location", loginURL)
		ctx.Request().SetStatusCode(302)
		return FilterStop, nil
	}
	
	// 没有配置登录页，直接返回403
	ctx.Request().SetStatusCode(403)
	ctx.Request().Response.SetBodyString("Access Denied")
	return FilterStop, nil
}

// isExcludedPath 检查路径是否在排除列表中
func isExcludedPath(config *SSOConfig, path string) bool {
	for _, excludePath := range config.ExcludePaths {
		if matchPath(excludePath, path) {
			return true
		}
	}
	return false
}

// matchPath 路径匹配（支持通配符）
func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	
	// 支持 * 通配符
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	
	return false
}

// checkSessionSSO 检查Session中的SSO状态
func checkSessionSSO(manager *Manager, ctx *context.Context) *UserInfo {
	// 检查Session中是否有用户信息
	userInfo := manager.GetSessionData(ctx, "user")
	if userInfo == nil {
		return nil
	}
	
	// 检查SSO令牌时间戳
	ssoTokenTime := manager.GetSessionData(ctx, "sso_token")
	if ssoTokenTime == nil {
		return nil
	}
	
	// 验证令牌是否过期
	tokenTime, ok := ssoTokenTime.(int64)
	if !ok {
		return nil
	}
	
	config := GetSSOConfig()
	if time.Now().Unix()-tokenTime > config.TokenExpireTime {
		// 令牌已过期，清除Session
		manager.DeleteSessionData(ctx, "user")
		manager.DeleteSessionData(ctx, "sso_token")
		return nil
	}
	
	// 尝试转换为UserInfo
	if user, ok := userInfo.(*UserInfo); ok {
		return user
	}
	
	return nil
}

// checkCookieSSO 检查Cookie中的SSO令牌
func checkCookieSSO(manager *Manager, ctx *context.Context, config *SSOConfig) *UserInfo {
	// 获取SSO Cookie
	tokenStr := manager.GetCookie(ctx, config.CookieName)
	if tokenStr == "" {
		return nil
	}
	
	// 解析和验证令牌
	token, err := parseSSOToken(tokenStr, config.Secret)
	if err != nil {
		return nil
	}
	
	// 检查令牌是否过期
	if time.Now().After(token.ExpiresAt) {
		// 令牌已过期，删除Cookie
		manager.DeleteCookie(ctx, config.CookieName, config.CookiePath)
		return nil
	}
	
	// 验证客户端IP（可选）
	clientIP := getClientIP(ctx)
	if token.ClientIP != "" && token.ClientIP != clientIP {
		// IP不匹配，可能是安全问题
		return nil
	}
	
	// 如果有自定义验证函数，使用它
	if config.TokenValidator != nil {
		userInfo, err := config.TokenValidator(token)
		if err != nil {
			return nil
		}
		return userInfo
	}
	
	// 默认验证：转换为UserInfo
	return &UserInfo{
		ID:       token.UserID,
		Username: token.Username,
		Email:    token.Email,
		Roles:    token.Roles,
		Profile:  make(map[string]interface{}),
	}
}

// checkRememberMe 检查记住我功能
func checkRememberMe(manager *Manager, ctx *context.Context, config *SSOConfig) *UserInfo {
	// 获取记住我Cookie
	rememberToken := manager.GetCookie(ctx, config.RememberCookieName)
	if rememberToken == "" {
		return nil
	}
	
	// 如果有自定义验证函数，使用它
	if config.RememberValidator != nil {
		userInfo, err := config.RememberValidator(rememberToken)
		if err != nil {
			// 验证失败，删除无效的记住我Cookie
			manager.DeleteCookie(ctx, config.RememberCookieName, config.CookiePath)
			return nil
		}
		return userInfo
	}
	
	// 默认的记住我验证（简单的base64解码）
	// 生产环境应该使用更安全的方法
	data, err := base64.URLEncoding.DecodeString(rememberToken)
	if err != nil {
		return nil
	}
	
	var userInfo UserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil
	}
	
	return &userInfo
}

// isAPIRequest 检查是否是API请求
func isAPIRequest(path string) bool {
	return strings.HasPrefix(path, "/api/") || 
		   strings.HasPrefix(path, "/v1/") ||
		   strings.HasPrefix(path, "/v2/")
}

// setSSOCookie 设置SSO Cookie
func setSSOCookie(manager *Manager, ctx *context.Context, config *SSOConfig, userInfo *UserInfo) error {
	// 创建SSO令牌
	token := &SSOToken{
		UserID:    userInfo.ID,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		Roles:     userInfo.Roles,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(config.TokenExpireTime) * time.Second),
		ClientIP:  getClientIP(ctx),
		UserAgent: string(ctx.Request().Request.Header.UserAgent()),
	}
	
	// 生成令牌字符串
	tokenStr, err := generateSSOToken(token, config.Secret)
	if err != nil {
		return err
	}
	
	// 设置Cookie选项
	options := &cookie.Options{
		MaxAge:   int(config.TokenExpireTime),
		Domain:   config.CookieDomain,
		Path:     config.CookiePath,
		Secure:   config.CookieSecure,
		HttpOnly: config.CookieHttpOnly,
	}
	
	// 设置Cookie
	manager.SetCookie(ctx, config.CookieName, tokenStr, options)
	
	return nil
}

// generateSSOToken 生成SSO令牌字符串
func generateSSOToken(token *SSOToken, secret string) (string, error) {
	// 序列化令牌
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	
	// Base64编码
	payload := base64.URLEncoding.EncodeToString(data)
	
	// 生成签名
	signature := generateSignature(payload, secret)
	
	// 组合令牌：payload.signature
	return fmt.Sprintf("%s.%s", payload, signature), nil
}

// parseSSOToken 解析SSO令牌
func parseSSOToken(tokenStr, secret string) (*SSOToken, error) {
	// 分割令牌
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}
	
	payload, signature := parts[0], parts[1]
	
	// 验证签名
	expectedSignature := generateSignature(payload, secret)
	if signature != expectedSignature {
		return nil, fmt.Errorf("invalid token signature")
	}
	
	// 解码payload
	data, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}
	
	// 反序列化令牌
	var token SSOToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	
	return &token, nil
}

// generateSignature 生成HMAC-SHA256签名
func generateSignature(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// getClientIP 获取客户端IP地址
func getClientIP(ctx *context.Context) string {
	// 尝试从X-Forwarded-For头获取
	if xff := string(ctx.Request().Request.Header.Peek("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	
	// 尝试从X-Real-IP头获取
	if xri := string(ctx.Request().Request.Header.Peek("X-Real-IP")); xri != "" {
		return strings.TrimSpace(xri)
	}
	
	// 从RemoteAddr获取
	return ctx.Request().RemoteAddr().String()
}


// ============= SSO辅助函数 =============

// LoginUser 用户登录（设置SSO状态）
//
// 在用户成功登录后调用此函数，设置SSO会话和Cookie。
//
// 参数：
//   - manager: 统一管理器
//   - ctx: 请求上下文
//   - userInfo: 用户信息
//   - rememberMe: 是否记住我
//
// 返回：
//   - error: 设置错误
func LoginUser(manager *Manager, ctx *context.Context, userInfo *UserInfo, rememberMe bool) error {
	config := GetSSOConfig()
	
	// 设置Session
	manager.SetSessionData(ctx, "user", userInfo)
	manager.SetSessionData(ctx, "sso_token", time.Now().Unix())
	
	// 设置上下文
	manager.SetContextData(ctx, "user", userInfo)
	manager.SetContextData(ctx, "authenticated", true)
	
	// 设置SSO Cookie
	if err := setSSOCookie(manager, ctx, config, userInfo); err != nil {
		return fmt.Errorf("failed to set SSO cookie: %w", err)
	}
	
	// 如果启用记住我功能
	if rememberMe {
		if err := setRememberMeCookie(manager, ctx, config, userInfo); err != nil {
			// 记住我Cookie设置失败不应该阻断登录
			fmt.Printf("Warning: Failed to set remember me cookie: %v\n", err)
		}
	}
	
	return nil
}

// LogoutUser 用户登出（清除SSO状态）
//
// 在用户登出时调用此函数，清除所有SSO相关数据。
//
// 参数：
//   - manager: 统一管理器
//   - ctx: 请求上下文
func LogoutUser(manager *Manager, ctx *context.Context) {
	config := GetSSOConfig()
	
	// 清除Session
	manager.DeleteSessionData(ctx, "user")
	manager.DeleteSessionData(ctx, "sso_token")
	
	// 清除上下文
	manager.SetContextData(ctx, "user", nil)
	manager.SetContextData(ctx, "authenticated", false)
	
	// 删除Cookie
	manager.DeleteCookie(ctx, config.CookieName, config.CookiePath)
	manager.DeleteCookie(ctx, config.RememberCookieName, config.CookiePath)
}

// GetCurrentUser 获取当前用户信息
//
// 从上下文中获取当前认证用户的信息。
//
// 参数：
//   - manager: 统一管理器
//   - ctx: 请求上下文
//
// 返回：
//   - *UserInfo: 用户信息，如果未认证返回nil
func GetCurrentUser(manager *Manager, ctx *context.Context) *UserInfo {
	if user := manager.GetContextData(ctx, "user"); user != nil {
		if userInfo, ok := user.(*UserInfo); ok {
			return userInfo
		}
	}
	return nil
}

// IsAuthenticated 检查用户是否已认证
//
// 参数：
//   - manager: 统一管理器
//   - ctx: 请求上下文
//
// 返回：
//   - bool: 是否已认证
func IsAuthenticated(manager *Manager, ctx *context.Context) bool {
	if authenticated := manager.GetContextData(ctx, "authenticated"); authenticated != nil {
		if auth, ok := authenticated.(bool); ok {
			return auth
		}
	}
	return false
}

// setRememberMeCookie 设置记住我Cookie
func setRememberMeCookie(manager *Manager, ctx *context.Context, config *SSOConfig, userInfo *UserInfo) error {
	// 简单的记住我实现（生产环境应该使用更安全的方法）
	data, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	
	rememberToken := base64.URLEncoding.EncodeToString(data)
	
	// 设置记住我Cookie（长期有效）
	options := &cookie.Options{
		MaxAge:   int(config.RememberExpireTime),
		Domain:   config.CookieDomain,
		Path:     config.CookiePath,
		Secure:   config.CookieSecure,
		HttpOnly: config.CookieHttpOnly,
	}
	
	manager.SetCookie(ctx, config.RememberCookieName, rememberToken, options)
	
	return nil
}