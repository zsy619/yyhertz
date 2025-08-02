package core

import (
	"fmt"
	"strings"

	"github.com/zsy619/yyhertz/framework/mvc/cookie"
)

// ============= Cookie操作方法（委托给Helper） =============

// SetCookie 设置Cookie
func (c *BaseController) SetCookie(name, value string, options ...*cookie.Options) {
	if c.Ctx == nil {
		return
	}
	c.cookieHelper.Set(c.Ctx.RequestContext, name, value, options...)
}

// GetCookie 获取Cookie
func (c *BaseController) GetCookie(name string) string {
	if c.Ctx == nil {
		return ""
	}
	return c.cookieHelper.Get(c.Ctx.RequestContext, name)
}

// DeleteCookie 删除Cookie
func (c *BaseController) DeleteCookie(name string, path ...string) {
	if c.Ctx == nil {
		return
	}
	c.cookieHelper.Delete(c.Ctx.RequestContext, name, path...)
}

// HasCookie 检查Cookie是否存在
func (c *BaseController) HasCookie(name string) bool {
	if c.Ctx == nil {
		return false
	}
	return c.cookieHelper.Has(c.Ctx.RequestContext, name)
}

// SetSecureCookie 设置安全Cookie（Beego兼容）
func (c *BaseController) SetSecureCookie(secret, name, value string, others ...any) {
	// 简化实现，实际使用中可以集成更复杂的加密逻辑
	options := &cookie.Options{
		MaxAge:   3600, // 默认1小时
		HttpOnly: true,
		Secure:   true,
	}

	if len(others) > 0 {
		if maxAge, ok := others[0].(int); ok {
			options.MaxAge = maxAge
		}
	}

	// 这里可以添加加密逻辑
	encryptedValue := c.encryptCookieValue(secret, value)
	c.SetCookie(name, encryptedValue, options)
}

// GetSecureCookie 获取安全Cookie（Beego兼容）
func (c *BaseController) GetSecureCookie(secret, name string) (string, bool) {
	encryptedValue := c.GetCookie(name)
	if encryptedValue == "" {
		return "", false
	}

	// 这里可以添加解密逻辑
	value, ok := c.decryptCookieValue(secret, encryptedValue)
	return value, ok
}

// encryptCookieValue 加密Cookie值（简化实现）
func (c *BaseController) encryptCookieValue(secret, value string) string {
	// 简化实现：实际项目中应使用更安全的加密算法
	return fmt.Sprintf("%s:%s", secret, value)
}

// decryptCookieValue 解密Cookie值（简化实现）
func (c *BaseController) decryptCookieValue(secret, encryptedValue string) (string, bool) {
	// 简化实现：实际项目中应使用对应的解密算法
	parts := strings.SplitN(encryptedValue, ":", 2)
	if len(parts) != 2 || parts[0] != secret {
		return "", false
	}
	return parts[1], true
}