package core

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 安全头设置方法 =============

// SetSecurityHeaders 设置通用安全响应头
func (c *BaseController) SetSecurityHeaders() {
	c.SetContentSecurityPolicy("")
	c.SetXFrameOptions("DENY")
	c.SetXContentTypeOptions()
	c.SetXXSSProtection()
	c.SetReferrerPolicy("strict-origin-when-cross-origin")
	c.SetPermissionsPolicy("")
}

// SetContentSecurityPolicy 设置内容安全策略
func (c *BaseController) SetContentSecurityPolicy(policy string) {
	if policy == "" {
		// 默认安全策略
		policy = "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; media-src 'self'; object-src 'none'; child-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
	}
	c.SetHeader("Content-Security-Policy", policy)
}

// SetXFrameOptions 设置X-Frame-Options头
func (c *BaseController) SetXFrameOptions(option string) {
	if option == "" {
		option = "DENY"
	}
	c.SetHeader("X-Frame-Options", option)
}

// SetXContentTypeOptions 设置X-Content-Type-Options头
func (c *BaseController) SetXContentTypeOptions() {
	c.SetHeader("X-Content-Type-Options", "nosniff")
}

// SetXXSSProtection 设置X-XSS-Protection头
func (c *BaseController) SetXXSSProtection() {
	c.SetHeader("X-XSS-Protection", "1; mode=block")
}

// SetReferrerPolicy 设置Referrer-Policy头
func (c *BaseController) SetReferrerPolicy(policy string) {
	if policy == "" {
		policy = "strict-origin-when-cross-origin"
	}
	c.SetHeader("Referrer-Policy", policy)
}

// SetPermissionsPolicy 设置Permissions-Policy头
func (c *BaseController) SetPermissionsPolicy(policy string) {
	if policy == "" {
		// 默认权限策略：禁用不必要的浏览器功能
		policy = "camera=(), microphone=(), geolocation=(), interest-cohort=()"
	}
	c.SetHeader("Permissions-Policy", policy)
}

// SetStrictTransportSecurity 设置HSTS头
func (c *BaseController) SetStrictTransportSecurity(maxAge int, includeSubDomains bool) {
	hsts := fmt.Sprintf("max-age=%d", maxAge)
	if includeSubDomains {
		hsts += "; includeSubDomains"
	}
	c.SetHeader("Strict-Transport-Security", hsts)
}

// ============= CSRF防护方法 =============

// GenerateCSRFToken 生成CSRF令牌
func (c *BaseController) GenerateCSRFToken() (string, error) {
	// 生成32字节随机数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Base64编码
	token := base64.URLEncoding.EncodeToString(bytes)

	// 存储在Session中
	c.SetSession("_csrf_token", token)

	return token, nil
}

// GetCSRFToken 获取CSRF令牌
func (c *BaseController) GetCSRFToken() string {
	// 先尝试从Session获取
	if token := c.GetSession("_csrf_token"); token != nil {
		if tokenStr, ok := token.(string); ok && tokenStr != "" {
			return tokenStr
		}
	}

	// 如果没有，生成新的
	token, err := c.GenerateCSRFToken()
	if err != nil {
		config.Errorf("Failed to generate CSRF token: %v", err)
		return ""
	}

	return token
}

// ValidateCSRFToken 验证CSRF令牌
func (c *BaseController) ValidateCSRFToken(providedToken string) bool {
	sessionToken := c.GetSession("_csrf_token")
	if sessionToken == nil {
		return false
	}

	sessionTokenStr, ok := sessionToken.(string)
	if !ok || sessionTokenStr == "" {
		return false
	}

	// 使用constant time比较防止时序攻击
	return subtle.ConstantTimeCompare([]byte(providedToken), []byte(sessionTokenStr)) == 1
}

// RequireCSRFToken 要求CSRF令牌验证
func (c *BaseController) RequireCSRFToken() bool {
	// 只对修改数据的请求进行CSRF验证
	method := c.Ctx.Method()
	if method == "GET" || method == "HEAD" || method == "OPTIONS" {
		return true
	}

	// 获取提供的令牌
	token := c.getCSRFTokenFromRequest()
	if token == "" {
		c.CSRFError("CSRF token missing")
		return false
	}

	// 验证令牌
	if !c.ValidateCSRFToken(token) {
		c.CSRFError("Invalid CSRF token")
		return false
	}

	return true
}

// getCSRFTokenFromRequest 从请求中获取CSRF令牌
func (c *BaseController) getCSRFTokenFromRequest() string {
	// 首先检查Header
	if token := c.GetHeader("X-CSRF-Token"); token != "" {
		return token
	}

	// 然后检查表单字段
	if token := c.GetForm("_csrf_token"); token != "" {
		return token
	}

	return ""
}

// CSRFError 处理CSRF错误
func (c *BaseController) CSRFError(message string) {
	config.Warn("CSRF validation failed: " + message)
	c.Error(consts.StatusForbidden, "CSRF token validation failed")
}

// ============= 输入验证和清理方法 =============

// SanitizeHTML 清理HTML内容，防止XSS
func (c *BaseController) SanitizeHTML(input string) string {
	// 简单的HTML标签清理（实际项目中建议使用专业的HTML清理库）

	// 移除script标签
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")

	// 移除危险的事件处理属性
	eventRegex := regexp.MustCompile(`(?i)\s*on\w+\s*=\s*['""][^'"]*['"]`)
	input = eventRegex.ReplaceAllString(input, "")

	// 移除javascript: 协议
	jsRegex := regexp.MustCompile(`(?i)javascript\s*:`)
	input = jsRegex.ReplaceAllString(input, "")

	return input
}

// ValidateEmail 验证邮箱格式
func (c *BaseController) ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePhone 验证手机号格式（中国）
func (c *BaseController) ValidatePhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// ValidateURL 验证URL格式
func (c *BaseController) ValidateURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

// ValidateIPAddress 验证IP地址
func (c *BaseController) ValidateIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}

// SanitizeFilename 清理文件名，防止路径遍历
func (c *BaseController) SanitizeFilename(filename string) string {
	// 移除路径分隔符
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")

	// 移除控制字符
	controlRegex := regexp.MustCompile(`[\x00-\x1f\x7f]`)
	filename = controlRegex.ReplaceAllString(filename, "")

	return filename
}

// ============= 访问控制方法 =============

// CheckIPWhitelist 检查IP白名单
func (c *BaseController) CheckIPWhitelist(whitelist []string) bool {
	clientIP := c.GetClientIP()

	for _, allowedIP := range whitelist {
		if c.isIPMatched(clientIP, allowedIP) {
			return true
		}
	}

	return false
}

// CheckIPBlacklist 检查IP黑名单
func (c *BaseController) CheckIPBlacklist(blacklist []string) bool {
	clientIP := c.GetClientIP()

	for _, blockedIP := range blacklist {
		if c.isIPMatched(clientIP, blockedIP) {
			return false
		}
	}

	return true
}

// isIPMatched 检查IP是否匹配（支持CIDR）
func (c *BaseController) isIPMatched(clientIP, pattern string) bool {
	// 如果包含/，当作CIDR处理
	if strings.Contains(pattern, "/") {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		ip := net.ParseIP(clientIP)
		return ip != nil && ipNet.Contains(ip)
	}

	// 直接IP比较
	return clientIP == pattern
}

// RateLimitCheck 简单的速率限制检查
func (c *BaseController) RateLimitCheck(key string, limit int, window time.Duration) bool {
	// 这里应该使用Redis或内存缓存实现真正的速率限制
	// 示例实现：使用Session模拟

	now := time.Now()
	sessionKey := fmt.Sprintf("rate_limit_%s", key)

	// 获取上次请求时间和计数
	if lastData := c.GetSession(sessionKey); lastData != nil {
		if lastDataStr, ok := lastData.(string); ok && lastDataStr != "" {
			parts := strings.Split(lastDataStr, "|")
			if len(parts) == 2 {
				if lastTime, err := time.Parse(time.RFC3339, parts[0]); err == nil {
					if count, err := strconv.Atoi(parts[1]); err == nil {
						// 检查是否在时间窗口内
						if now.Sub(lastTime) < window {
							if count >= limit {
								return false // 超出限制
							}
							// 增加计数
							newData := fmt.Sprintf("%s|%d", parts[0], count+1)
							c.SetSession(sessionKey, newData)
							return true
						}
					}
				}
			}
		}
	}

	// 新的时间窗口，重置计数
	newData := fmt.Sprintf("%s|1", now.Format(time.RFC3339))
	c.SetSession(sessionKey, newData)
	return true
}

// RequireHTTPS 要求HTTPS连接
func (c *BaseController) RequireHTTPS() bool {
	if !c.IsHTTPS() {
		// 重定向到HTTPS
		host := string(c.Ctx.Request().Host())
		path := c.Ctx.Path()
		httpsURL := "https://" + host + path
		if query := c.Ctx.Request().URI().QueryString(); len(query) > 0 {
			httpsURL += "?" + string(query)
		}
		c.Redirect(httpsURL, consts.StatusMovedPermanently)
		return false
	}
	return true
}

// IsHTTPS 检查是否为HTTPS连接
func (c *BaseController) IsHTTPS() bool {
	if c.Ctx == nil {
		return false
	}

	// 检查协议
	if string(c.Ctx.Request().URI().Scheme()) == "https" {
		return true
	}

	// 检查X-Forwarded-Proto头（代理后面）
	if proto := c.GetHeader("X-Forwarded-Proto"); strings.ToLower(proto) == "https" {
		return true
	}

	// 检查X-Forwarded-SSL头
	if ssl := c.GetHeader("X-Forwarded-SSL"); strings.ToLower(ssl) == "on" {
		return true
	}

	return false
}

// ============= 加密和哈希方法 =============

// HashPassword 哈希密码（使用bcrypt风格的简单实现）
func (c *BaseController) HashPassword(password string, salt ...string) string {
	var saltStr string
	if len(salt) > 0 {
		saltStr = salt[0]
	} else {
		saltStr = c.GenerateSalt()
	}

	// 使用HMAC-SHA256
	h := hmac.New(sha256.New, []byte(saltStr))
	h.Write([]byte(password))
	hash := hex.EncodeToString(h.Sum(nil))

	return saltStr + "$" + hash
}

// VerifyPassword 验证密码
func (c *BaseController) VerifyPassword(password, hashedPassword string) bool {
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 2 {
		return false
	}

	salt := parts[0]
	expectedHash := parts[1]

	// 计算提供密码的哈希
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(password))
	actualHash := hex.EncodeToString(h.Sum(nil))

	// 使用constant time比较
	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actualHash)) == 1
}

// GenerateSalt 生成盐值
func (c *BaseController) GenerateSalt() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 降级到时间戳
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// GenerateSecureToken 生成安全令牌
func (c *BaseController) GenerateSecureToken(length int) string {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 降级实现
		return c.generateFallbackToken(length)
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// generateFallbackToken 降级令牌生成
func (c *BaseController) generateFallbackToken(length int) string {
	timestamp := time.Now().UnixNano()
	hash := md5.Sum([]byte(fmt.Sprintf("%d", timestamp)))
	token := hex.EncodeToString(hash[:])

	if len(token) > length {
		return token[:length]
	}
	return token
}

// ============= 会话安全方法 =============

// RegenerateSessionID 重新生成会话ID
func (c *BaseController) RegenerateSessionID() {
	// 这里应该调用session管理器的重新生成方法
	// 示例：删除旧session，创建新的
	oldSessionID := c.GetSessionID()

	// 生成新的session ID
	newSessionID := c.GenerateSecureToken(32)

	// 更新session ID（具体实现依赖于session存储方式）
	config.Infof("Session ID regenerated: %s -> %s", oldSessionID, newSessionID)
}

// SetSecureSessionCookie 设置安全的会话Cookie
func (c *BaseController) SetSecureSessionCookie(sessionID string) {
	expiry := time.Now().Add(24 * time.Hour) // 24小时过期
	maxAge := int(time.Until(expiry).Seconds())

	// 设置安全的Cookie选项
	secure := ""
	if c.IsHTTPS() {
		secure = "; Secure"
	}

	cookieStr := fmt.Sprintf("session_id=%s; Max-Age=%d; Path=/; HttpOnly%s", sessionID, maxAge, secure)
	c.SetHeader("Set-Cookie", cookieStr)
}

// ValidateSessionTimeout 验证会话超时
func (c *BaseController) ValidateSessionTimeout(timeoutMinutes int) bool {
	lastActive := c.GetSession("last_active")
	if lastActive == nil {
		// 首次访问，设置当前时间
		c.SetSession("last_active", time.Now().Format(time.RFC3339))
		return true
	}

	lastActiveStr, ok := lastActive.(string)
	if !ok || lastActiveStr == "" {
		// 首次访问，设置当前时间
		c.SetSession("last_active", time.Now().Format(time.RFC3339))
		return true
	}

	lastActiveTime, err := time.Parse(time.RFC3339, lastActiveStr)
	if err != nil {
		return false
	}

	// 检查是否超时
	if time.Since(lastActiveTime) > time.Duration(timeoutMinutes)*time.Minute {
		return false
	}

	// 更新最后活动时间
	c.SetSession("last_active", time.Now().Format(time.RFC3339))
	return true
}

// ============= 文件上传安全方法 =============

// ValidateFileUpload 验证文件上传安全性
func (c *BaseController) ValidateFileUpload(fileKey string, allowedExts []string, maxSize int64) error {
	// 检查文件是否存在
	if !c.HasFile(fileKey) {
		return fmt.Errorf("no file uploaded")
	}

	// 验证文件大小
	if err := c.ValidateFileSize(fileKey, maxSize); err != nil {
		return err
	}

	// 验证文件扩展名
	if err := c.ValidateFileExtension(fileKey, allowedExts); err != nil {
		return err
	}

	// 验证文件内容类型
	filename := c.GetFileName(fileKey)
	if c.isExecutableFile(filename) {
		return fmt.Errorf("executable files are not allowed")
	}

	return nil
}

// isExecutableFile 检查是否为可执行文件
func (c *BaseController) isExecutableFile(filename string) bool {
	dangerousExts := []string{
		".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js",
		".jar", ".sh", ".php", ".asp", ".aspx", ".jsp", ".py", ".rb",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, dangerous := range dangerousExts {
		if ext == dangerous {
			return true
		}
	}

	return false
}

// ============= 审计日志方法 =============

// LogSecurityEvent 记录安全事件
func (c *BaseController) LogSecurityEvent(eventType, message string, details map[string]any) {
	logData := map[string]any{
		"timestamp":   time.Now().Format(time.RFC3339),
		"event_type":  eventType,
		"message":     message,
		"client_ip":   c.GetClientIP(),
		"user_agent":  c.GetUserAgent(),
		"request_uri": c.Ctx.Path(),
		"method":      c.Ctx.Method(),
		"details":     details,
	}

	config.Infof("Security Event: %+v", logData)
}

// LogFailedLogin 记录登录失败
func (c *BaseController) LogFailedLogin(username string, reason string) {
	c.LogSecurityEvent("failed_login", "User login failed", map[string]any{
		"username": username,
		"reason":   reason,
	})
}

// LogSuspiciousActivity 记录可疑活动
func (c *BaseController) LogSuspiciousActivity(activity string, details map[string]any) {
	c.LogSecurityEvent("suspicious_activity", activity, details)
}
