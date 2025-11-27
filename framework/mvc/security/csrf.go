package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// generateCSRFSecureKey 动态生成CSRF安全密钥
func generateCSRFSecureKey() string {
	// 生成48字节随机数据，base64编码后约64字符
	bytes := make([]byte, 48)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用备用方案
		return "YYHertz-CSRF-Fallback-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// CSRFConfig CSRF token配置
type CSRFConfig struct {
	// Secret 用于签名的密钥
	Secret string `json:"secret" yaml:"secret"`

	// TokenLength Token长度（字节）
	TokenLength int `json:"token_length" yaml:"token_length"`

	// ExpireTime 过期时间（秒）
	ExpireTime int64 `json:"expire_time" yaml:"expire_time"`

	// CookieName CSRF token的Cookie名称
	CookieName string `json:"cookie_name" yaml:"cookie_name"`

	// HeaderName 请求头中的CSRF token名称
	HeaderName string `json:"header_name" yaml:"header_name"`

	// FormFieldName 表单字段中的CSRF token名称
	FormFieldName string `json:"form_field_name" yaml:"form_field_name"`

	// SkipCheckFunc 跳过检查的函数（可选）
	SkipCheckFunc func(userID, clientIP string) bool `json:"-" yaml:"-"`
}

// DefaultCSRFConfig 默认CSRF配置
func DefaultCSRFConfig() *CSRFConfig {
	return &CSRFConfig{
		Secret:        generateCSRFSecureKey(), // 动态生成CSRF密钥
		TokenLength:   32,
		ExpireTime:    3600, // 1小时
		CookieName:    "_csrf_token",
		HeaderName:    "X-CSRF-Token",
		FormFieldName: "csrf_token",
	}
}

// LoadCSRFConfig 从配置文件加载CSRF配置
func LoadCSRFConfig() *CSRFConfig {
	// 先获取默认配置
	cfg := DefaultCSRFConfig()

	// 尝试从session配置中读取CSRF设置
	sessionConfigPtr, err := config.GetSessionConfig()
	if err != nil {
		// 如果加载失败，返回默认配置
		return cfg
	}

	// 将session配置中的CSRF设置转换为CSRFConfig
	return convertSessionCSRFToCSRFConfig(sessionConfigPtr, cfg)
}

// convertSessionCSRFToCSRFConfig 将SessionConfig中的CSRF配置转换为CSRFConfig
func convertSessionCSRFToCSRFConfig(sessionConfig *config.SessionConfig, defaultConfig *CSRFConfig) *CSRFConfig {
	csrfConfig := &CSRFConfig{
		Secret:        sessionConfig.Session.Security.Csrf.Secret,
		TokenLength:   sessionConfig.Session.Security.Csrf.TokenLength,
		ExpireTime:    sessionConfig.Session.Security.Csrf.ExpireTime,
		CookieName:    sessionConfig.Session.Security.Csrf.CookieName,
		HeaderName:    sessionConfig.Session.Security.Csrf.HeaderName,
		FormFieldName: sessionConfig.Session.Security.Csrf.FormFieldName,
	}

	// 如果某些值为空，使用默认值
	if csrfConfig.Secret == "" {
		csrfConfig.Secret = defaultConfig.Secret
	}
	if csrfConfig.TokenLength == 0 {
		csrfConfig.TokenLength = defaultConfig.TokenLength
	}
	if csrfConfig.ExpireTime == 0 {
		csrfConfig.ExpireTime = defaultConfig.ExpireTime
	}
	if csrfConfig.CookieName == "" {
		csrfConfig.CookieName = defaultConfig.CookieName
	}
	if csrfConfig.HeaderName == "" {
		csrfConfig.HeaderName = defaultConfig.HeaderName
	}
	if csrfConfig.FormFieldName == "" {
		csrfConfig.FormFieldName = defaultConfig.FormFieldName
	}

	return csrfConfig
}

// CSRFManager CSRF token管理器
type CSRFManager struct {
	config *CSRFConfig
	mutex  sync.RWMutex
}

// NewCSRFManager 创建新的CSRF管理器
func NewCSRFManager(config *CSRFConfig) *CSRFManager {
	if config == nil {
		config = LoadCSRFConfig()
	}
	return &CSRFManager{
		config: config,
	}
}

// CSRFToken CSRF token结构
type CSRFToken struct {
	Value     string    `json:"value"`
	UserID    string    `json:"user_id"`
	ClientIP  string    `json:"client_ip"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GenerateToken 生成CSRF token
func (m *CSRFManager) GenerateToken(userID, clientIP string) (*CSRFToken, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 检查是否需要跳过
	if m.config.SkipCheckFunc != nil && m.config.SkipCheckFunc(userID, clientIP) {
		return &CSRFToken{
			Value:     "skip-csrf-check",
			UserID:    userID,
			ClientIP:  clientIP,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Duration(m.config.ExpireTime) * time.Second),
		}, nil
	}

	// 生成随机盐值
	salt := make([]byte, m.config.TokenLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}

	// 添加时间戳
	now := time.Now()
	timestamp := now.Unix()

	// 创建令牌数据
	data := fmt.Sprintf("%s:%s:%d:%s",
		userID,
		clientIP,
		timestamp,
		base64.StdEncoding.EncodeToString(salt))

	// 使用HMAC签名
	h := hmac.New(sha256.New, []byte(m.config.Secret))
	h.Write([]byte(data))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 组合最终令牌
	tokenValue := fmt.Sprintf("%s:%s", data, signature)
	encodedToken := base64.StdEncoding.EncodeToString([]byte(tokenValue))

	return &CSRFToken{
		Value:     encodedToken,
		UserID:    userID,
		ClientIP:  clientIP,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(m.config.ExpireTime) * time.Second),
	}, nil
}

// ValidateToken 验证CSRF token
func (m *CSRFManager) ValidateToken(tokenValue, userID, clientIP string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 检查是否需要跳过
	if m.config.SkipCheckFunc != nil && m.config.SkipCheckFunc(userID, clientIP) {
		return true, nil
	}

	if tokenValue == "" {
		return false, fmt.Errorf("CSRF token is empty")
	}

	// 特殊值处理
	if tokenValue == "skip-csrf-check" {
		return true, nil
	}

	// 解码令牌
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenValue)
	if err != nil {
		return false, fmt.Errorf("failed to decode CSRF token: %w", err)
	}

	tokenParts := strings.Split(string(tokenBytes), ":")
	if len(tokenParts) != 5 {
		return false, fmt.Errorf("invalid CSRF token format")
	}

	tokenUserID := tokenParts[0]
	tokenClientIP := tokenParts[1]
	timestampStr := tokenParts[2]
	salt := tokenParts[3]
	providedSignature := tokenParts[4]

	// 验证用户ID和客户端IP
	if tokenUserID != userID || tokenClientIP != clientIP {
		return false, fmt.Errorf("CSRF token validation failed: user or IP mismatch")
	}

	// 检查时间戳是否过期
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp in CSRF token: %w", err)
	}

	if time.Now().Unix()-timestamp > m.config.ExpireTime {
		return false, fmt.Errorf("CSRF token has expired")
	}

	// 重新计算签名
	data := fmt.Sprintf("%s:%s:%s:%s", tokenUserID, tokenClientIP, timestampStr, salt)
	h := hmac.New(sha256.New, []byte(m.config.Secret))
	h.Write([]byte(data))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 比较签名
	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return false, fmt.Errorf("CSRF token signature validation failed")
	}

	return true, nil
}

// GenerateSimpleToken 生成简单的CSRF token（用于与现有系统兼容）
func (m *CSRFManager) GenerateSimpleToken() string {
	// 生成简单的随机token，用于模板引擎等场景
	salt := make([]byte, 16)
	rand.Read(salt)

	timestamp := time.Now().Unix()
	data := fmt.Sprintf("simple:%d:%s", timestamp, base64.StdEncoding.EncodeToString(salt))

	h := hmac.New(sha256.New, []byte(m.config.Secret))
	h.Write([]byte(data))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	token := fmt.Sprintf("%s:%s", data, signature)
	return base64.StdEncoding.EncodeToString([]byte(token))
}

// GetConfig 获取配置
func (m *CSRFManager) GetConfig() *CSRFConfig {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 返回配置的副本
	configCopy := *m.config
	return &configCopy
}

// UpdateConfig 更新配置
func (m *CSRFManager) UpdateConfig(config *CSRFConfig) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if config != nil {
		m.config = config
	}
}

// SetSkipCheckFunc 设置跳过检查的函数
func (m *CSRFManager) SetSkipCheckFunc(fn func(userID, clientIP string) bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.config.SkipCheckFunc = fn
}
