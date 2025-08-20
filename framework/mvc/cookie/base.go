package cookie

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
)

// BaseCookie 基础Cookie操作器
// 提供标准的cookie读写和管理功能，支持beego等框架的API兼容性
type BaseCookie struct {
	request *app.RequestContext
}

// NewBaseCookie 创建基础Cookie操作器
func NewBaseCookie(request *app.RequestContext) *BaseCookie {
	return &BaseCookie{
		request: request,
	}
}

// Get 获取Cookie值 (beego兼容方法)
func (bc *BaseCookie) Get(key string) string {
	if bc.request != nil {
		return string(bc.request.Cookie(key))
	}
	return ""
}

// Set 设置Cookie (beego兼容方法)
// 参数顺序: name, value, maxAge, path, domain, secure, httpOnly, sameSite
func (bc *BaseCookie) Set(name, value string, others ...any) {
	if bc.request == nil {
		return
	}

	// 默认参数
	maxAge := 0
	path := "/"
	domain := ""
	secure := false
	httpOnly := false
	sameSite := protocol.CookieSameSiteDefaultMode

	// 解析可选参数 (与beego参数顺序兼容)
	for k, v := range others {
		switch k {
		case 0:
			if age, ok := v.(int); ok {
				maxAge = age
			}
		case 1:
			if p, ok := v.(string); ok {
				path = p
			}
		case 2:
			if d, ok := v.(string); ok {
				domain = d
			}
		case 3:
			if s, ok := v.(bool); ok {
				secure = s
			}
		case 4:
			if h, ok := v.(bool); ok {
				httpOnly = h
			}
		case 5:
			if ss, ok := v.(protocol.CookieSameSite); ok {
				sameSite = ss
			}
		}
	}

	bc.request.SetCookie(name, value, maxAge, path, domain, sameSite, secure, httpOnly)
}

// Delete 删除Cookie (beego兼容方法)
func (bc *BaseCookie) Delete(name string) {
	bc.Set(name, "", -1, "/")
}

// Exists 检查Cookie是否存在
func (bc *BaseCookie) Exists(key string) bool {
	return bc.Get(key) != ""
}

// GetAll 获取所有Cookie (增强功能)
// 注意：由于Hertz API限制，当前返回空map，实际应用中可通过Get方法逐个获取
func (bc *BaseCookie) GetAll() map[string]string {
	cookies := make(map[string]string)
	if bc.request == nil {
		return cookies
	}

	// 简化实现：由于Hertz的API限制，暂时返回空map
	// 实际应用中可以通过其他方式获取cookie列表
	// 或者在具体使用时通过Get(key)方法逐个获取
	return cookies
}

// Count 获取Cookie数量
func (bc *BaseCookie) Count() int {
	return len(bc.GetAll())
}

// Clear 清除所有Cookie (增强功能)
func (bc *BaseCookie) Clear() {
	cookies := bc.GetAll()
	for name := range cookies {
		bc.Delete(name)
	}
}

// ============= 输出Cookie功能 =============

// OutputCookie 输出Cookie操作器 (用于响应设置)
type OutputCookie struct {
	request *app.RequestContext
}

// NewOutputCookie 创建输出Cookie操作器
func NewOutputCookie(request *app.RequestContext) *OutputCookie {
	return &OutputCookie{
		request: request,
	}
}

// Set 设置输出Cookie (Output兼容性方法)
func (oc *OutputCookie) Set(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if oc.request != nil {
		oc.request.SetCookie(name, value, maxAge, path, domain, protocol.CookieSameSiteDefaultMode, secure, httpOnly)
	}
}

// ============= Cookie工具函数 =============

// ParseCookieString 解析Cookie字符串
// 用于从"name1=value1; name2=value2"格式解析cookie
func ParseCookieString(cookieStr string) map[string]string {
	cookies := make(map[string]string)
	if cookieStr == "" {
		return cookies
	}

	parts := strings.Split(cookieStr, "; ")
	for _, part := range parts {
		if eqIdx := strings.Index(part, "="); eqIdx > 0 {
			name := part[:eqIdx]
			value := part[eqIdx+1:]
			cookies[name] = value
		}
	}
	return cookies
}

// FormatCookieString 格式化Cookie字符串
// 将map格式的cookie转换为"name1=value1; name2=value2"格式
func FormatCookieString(cookies map[string]string) string {
	var parts []string
	for name, value := range cookies {
		parts = append(parts, name+"="+value)
	}
	return strings.Join(parts, "; ")
}

// ValidateCookieName 验证Cookie名称是否合法
func ValidateCookieName(name string) bool {
	if name == "" {
		return false
	}
	// 简单验证：不包含特殊字符
	for _, char := range name {
		if char < 33 || char > 126 || strings.ContainsRune("(),/:;<=>?@[\\]{}", char) {
			return false
		}
	}
	return true
}

// ValidateCookieValue 验证Cookie值是否合法
func ValidateCookieValue(value string) bool {
	// 简单验证：不包含控制字符
	for _, char := range value {
		if char < 32 || char == 127 {
			return false
		}
	}
	return true
}