package core

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
)

// ============= XSRF/CSRF安全方法 =============

// XSRFToken 生成XSRF令牌（Beego ControllerInterface兼容）
func (c *BaseController) XSRFToken() string {
	if c.xsrfToken != "" {
		return c.xsrfToken
	}
	
	// 生成新的XSRF令牌
	token := c.generateXSRFToken()
	c.xsrfToken = token
	
	// 设置XSRF Cookie
	c.SetCookie("_xsrf", token, &cookie.Options{
		MaxAge:   c.XSRFExpire,
		HttpOnly: false, // XSRF token需要被JavaScript访问
		Path:     "/",
		SameSite: "Strict",
	})
	
	// 将令牌添加到模板数据中
	if c.Data == nil {
		c.Data = make(map[string]any)
	}
	c.Data["XSRFToken"] = token
	
	return token
}

// CheckXSRFCookie 检查XSRF令牌（Beego ControllerInterface兼容）
func (c *BaseController) CheckXSRFCookie() bool {
	if !c.checkXSRFCookie {
		return true // 如果未启用XSRF检查，直接返回true
	}
	
	if c.Ctx == nil {
		return false
	}
	
	// 获取请求中的令牌
	var requestToken string
	
	// 1. 从Header中获取（优先级最高）
	if token := c.GetHeader("X-Xsrftoken"); token != "" {
		requestToken = token
	} else if token := c.GetHeader("X-CsrfToken"); token != "" {
		requestToken = token
	} else if token := c.GetHeader("X-CSRF-Token"); token != "" {
		requestToken = token
	} else {
		// 2. 从表单字段中获取
		requestToken = c.GetString("_xsrf")
	}
	
	if requestToken == "" {
		return false
	}
	
	// 获取Cookie中的令牌
	cookieToken := c.GetCookie("_xsrf")
	if cookieToken == "" {
		return false
	}
	
	// 验证令牌
	return c.validateXSRFToken(requestToken, cookieToken)
}

// EnableXSRF 启用XSRF保护
func (c *BaseController) EnableXSRF(expire ...int) {
	c.checkXSRFCookie = true
	if len(expire) > 0 {
		c.XSRFExpire = expire[0]
	}
}

// DisableXSRF 禁用XSRF保护
func (c *BaseController) DisableXSRF() {
	c.checkXSRFCookie = false
}

// ============= XSRF令牌相关的辅助方法 =============

// generateXSRFToken 生成XSRF令牌
func (c *BaseController) generateXSRFToken() string {
	// 生成随机盐值
	salt := make([]byte, 32)
	rand.Read(salt)
	
	// 获取用户标识（可以是Session ID或IP地址）
	var userID string
	if sessionID := c.GetSessionID(); sessionID != "" {
		userID = sessionID
	} else {
		userID = c.GetClientIP()
	}
	
	// 添加时间戳
	timestamp := time.Now().Unix()
	
	// 创建令牌数据
	data := fmt.Sprintf("%s:%d:%s", userID, timestamp, base64.StdEncoding.EncodeToString(salt))
	
	// 使用HMAC签名
	h := hmac.New(sha256.New, []byte(c.getXSRFSecret()))
	h.Write([]byte(data))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// 组合最终令牌
	token := fmt.Sprintf("%s:%s", data, signature)
	return base64.StdEncoding.EncodeToString([]byte(token))
}

// validateXSRFToken 验证XSRF令牌
func (c *BaseController) validateXSRFToken(requestToken, cookieToken string) bool {
	// 简化验证：检查令牌是否匹配
	if requestToken != cookieToken {
		return false
	}
	
	// 解码令牌
	tokenBytes, err := base64.StdEncoding.DecodeString(requestToken)
	if err != nil {
		return false
	}
	
	tokenParts := strings.Split(string(tokenBytes), ":")
	if len(tokenParts) != 4 {
		return false
	}
	
	// 检查时间戳是否过期
	timestamp, err := strconv.ParseInt(tokenParts[1], 10, 64)
	if err != nil {
		return false
	}
	
	if time.Now().Unix()-timestamp > int64(c.XSRFExpire) {
		return false
	}
	
	// 重新计算签名验证
	data := strings.Join(tokenParts[:3], ":")
	h := hmac.New(sha256.New, []byte(c.getXSRFSecret()))
	h.Write([]byte(data))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	return hmac.Equal([]byte(tokenParts[3]), []byte(expectedSignature))
}

// getXSRFSecret 获取XSRF密钥
func (c *BaseController) getXSRFSecret() string {
	// 尝试从配置中获取
	if secret := config.GetConfigString("security.xsrf_secret"); secret != "" {
		return secret
	}
	
	// 默认密钥（生产环境应该配置自定义密钥）
	return "yyhertz-default-xsrf-secret-key"
}