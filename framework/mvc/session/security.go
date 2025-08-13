package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// CookieSecurityOptions 安全Cookie配置选项
type CookieSecurityOptions struct {
	Secret          string        // HMAC密钥
	MaxAge          time.Duration // Cookie最大有效期
	ValidateExpiry  bool          // 是否验证时间戳过期
	RequireHTTPS    bool          // 是否要求HTTPS
}

// SecureCookie 安全Cookie操作器
// 提供HMAC-SHA256签名验证的安全cookie功能，防止篡改和重放攻击
type SecureCookie struct {
	*BaseCookie
}

// NewSecureCookie 创建安全Cookie操作器
func NewSecureCookie(request *app.RequestContext) *SecureCookie {
	return &SecureCookie{
		BaseCookie: NewBaseCookie(request),
	}
}

// GetSecure 获取安全Cookie (beego兼容方法)
// 使用HMAC-SHA256验证cookie完整性，防止篡改
func (sc *SecureCookie) GetSecure(secret, key string) (string, bool) {
	val := sc.Get(key)
	if val == "" {
		return "", false
	}

	parts := strings.SplitN(val, "|", 3)
	if len(parts) != 3 {
		return "", false
	}

	vs := parts[0]
	timestamp := parts[1]
	sig := parts[2]

	// 验证HMAC签名
	h := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return "", false
	}

	// 解码base64值
	res, err := base64.URLEncoding.DecodeString(vs)
	if err != nil {
		return "", false
	}
	return string(res), true
}

// SetSecure 设置安全Cookie (beego兼容方法)
// 使用HMAC-SHA256签名保护cookie完整性
func (sc *SecureCookie) SetSecure(secret, name, value string, others ...any) {
	// Base64编码值
	vs := base64.URLEncoding.EncodeToString([]byte(value))
	// 添加时间戳
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	// 生成HMAC签名
	h := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	// 组合成最终cookie值
	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	
	// 设置cookie
	sc.Set(name, cookie, others...)
}

// GetSecureWithOptions 获取安全Cookie (增强版本)
// 支持时间戳验证和更严格的安全检查
func (sc *SecureCookie) GetSecureWithOptions(key string, options CookieSecurityOptions) (string, bool, error) {
	if options.RequireHTTPS && sc.request != nil {
		// 检查是否为HTTPS - 通过URI scheme或X-Forwarded-Proto header
		uri := sc.request.URI()
		scheme := string(uri.Scheme())
		isHTTPS := scheme == "https"
		
		// 检查代理头
		if !isHTTPS {
			forwardedProto := string(sc.request.GetHeader("X-Forwarded-Proto"))
			isHTTPS = forwardedProto == "https"
		}
		
		if !isHTTPS {
			return "", false, fmt.Errorf("secure cookie requires HTTPS connection")
		}
	}

	val := sc.Get(key)
	if val == "" {
		return "", false, nil
	}

	parts := strings.SplitN(val, "|", 4)
	if len(parts) != 4 {
		// 兼容老格式 (3部分)
		if len(parts) == 3 {
			value, valid := sc.GetSecure(options.Secret, key)
			if !valid {
				return "", false, fmt.Errorf("invalid secure cookie")
			}
			return value, true, nil
		}
		return "", false, fmt.Errorf("invalid secure cookie format")
	}

	vs := parts[0]
	timestamp := parts[1]
	expiry := parts[2]
	sig := parts[3]

	// 验证HMAC签名
	h := hmac.New(sha256.New, []byte(options.Secret))
	fmt.Fprintf(h, "%s%s%s", vs, timestamp, expiry)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return "", false, fmt.Errorf("invalid cookie signature")
	}

	// 验证过期时间
	if options.ValidateExpiry && expiry != "0" {
		expiryTime, err := strconv.ParseInt(expiry, 10, 64)
		if err != nil {
			return "", false, fmt.Errorf("invalid expiry timestamp")
		}
		if time.Now().UnixNano() > expiryTime {
			return "", false, fmt.Errorf("cookie expired")
		}
	}

	// 验证最大生存时间
	if options.MaxAge > 0 {
		timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return "", false, fmt.Errorf("invalid timestamp")
		}
		createTime := time.Unix(0, timestampInt)
		if time.Since(createTime) > options.MaxAge {
			return "", false, fmt.Errorf("cookie too old")
		}
	}

	// 解码base64值
	res, err := base64.URLEncoding.DecodeString(vs)
	if err != nil {
		return "", false, fmt.Errorf("failed to decode cookie value: %w", err)
	}
	
	return string(res), true, nil
}

// SetSecureWithOptions 设置安全Cookie (增强版本)
// 提供更细粒度的安全控制和时间戳验证
func (sc *SecureCookie) SetSecureWithOptions(name, value string, options CookieSecurityOptions, others ...any) error {
	if options.RequireHTTPS && sc.request != nil {
		// 检查是否为HTTPS - 通过URI scheme或X-Forwarded-Proto header
		uri := sc.request.URI()
		scheme := string(uri.Scheme())
		isHTTPS := scheme == "https"
		
		// 检查代理头
		if !isHTTPS {
			forwardedProto := string(sc.request.GetHeader("X-Forwarded-Proto"))
			isHTTPS = forwardedProto == "https"
		}
		
		if !isHTTPS {
			return fmt.Errorf("secure cookie requires HTTPS connection")
		}
	}

	// Base64编码值
	vs := base64.URLEncoding.EncodeToString([]byte(value))
	// 添加时间戳和过期时间
	now := time.Now()
	timestamp := strconv.FormatInt(now.UnixNano(), 10)
	
	var expiry string
	if options.MaxAge > 0 {
		expiry = strconv.FormatInt(now.Add(options.MaxAge).UnixNano(), 10)
	} else {
		expiry = "0" // 不设置过期时间
	}
	
	// 生成HMAC签名 (包含过期时间)
	h := hmac.New(sha256.New, []byte(options.Secret))
	fmt.Fprintf(h, "%s%s%s", vs, timestamp, expiry)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	
	// 组合成最终cookie值: value|timestamp|expiry|signature
	cookie := strings.Join([]string{vs, timestamp, expiry, sig}, "|")
	
	// 设置cookie (强制启用安全选项)
	secureOthers := make([]any, len(others))
	copy(secureOthers, others)
	
	// 如果需要HTTPS，强制设置secure标志
	if options.RequireHTTPS && len(secureOthers) >= 4 {
		secureOthers[3] = true // secure = true
	}
	if len(secureOthers) >= 5 {
		secureOthers[4] = true // httpOnly = true
	}
	
	sc.Set(name, cookie, secureOthers...)
	return nil
}

// Validate 验证安全Cookie但不返回值 (增强功能)
func (sc *SecureCookie) Validate(secret, key string) bool {
	_, valid := sc.GetSecure(secret, key)
	return valid
}

// ValidateWithOptions 使用选项验证安全Cookie
func (sc *SecureCookie) ValidateWithOptions(key string, options CookieSecurityOptions) bool {
	_, valid, err := sc.GetSecureWithOptions(key, options)
	return valid && err == nil
}

// ============= 安全工具函数 =============

// GenerateSecureValue 生成安全的cookie值
func GenerateSecureValue(secret, value string) string {
	vs := base64.URLEncoding.EncodeToString([]byte(value))
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	h := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	return strings.Join([]string{vs, timestamp, sig}, "|")
}

// ParseSecureValue 解析安全cookie值
func ParseSecureValue(secret, cookieValue string) (string, bool) {
	parts := strings.SplitN(cookieValue, "|", 3)
	if len(parts) != 3 {
		return "", false
	}

	vs := parts[0]
	timestamp := parts[1]
	sig := parts[2]

	// 验证HMAC签名
	h := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return "", false
	}

	// 解码base64值
	res, err := base64.URLEncoding.DecodeString(vs)
	if err != nil {
		return "", false
	}
	return string(res), true
}

// IsSecureCookieExpired 检查安全cookie是否过期
func IsSecureCookieExpired(cookieValue string, maxAge time.Duration) bool {
	parts := strings.SplitN(cookieValue, "|", 4)
	if len(parts) < 3 {
		return true
	}

	timestamp := parts[1]
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return true
	}

	createTime := time.Unix(0, timestampInt)
	return time.Since(createTime) > maxAge
}

// GetSecureCookieTimestamp 获取安全cookie的创建时间戳
func GetSecureCookieTimestamp(cookieValue string) (time.Time, error) {
	parts := strings.SplitN(cookieValue, "|", 4)
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid cookie format")
	}

	timestamp := parts[1]
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	return time.Unix(0, timestampInt), nil
}