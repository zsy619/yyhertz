package view

import (
	"sync"
)

// CSRFTokenProvider CSRF token提供者接口
//
// 此接口用于解决循环导入问题，允许view包获取CSRF token
// 而不需要直接依赖mvc包
type CSRFTokenProvider interface {
	// GenerateSimpleToken 生成简单的CSRF token
	GenerateSimpleToken() string
}

// 全局CSRF token提供者
var (
	globalCSRFProvider CSRFTokenProvider
	csrfProviderMutex  sync.RWMutex
)

// SetCSRFTokenProvider 设置CSRF token提供者
//
// 此方法由mvc包调用，用于注册CSRF token提供者
func SetCSRFTokenProvider(provider CSRFTokenProvider) {
	csrfProviderMutex.Lock()
	defer csrfProviderMutex.Unlock()
	globalCSRFProvider = provider
}

// GetCSRFTokenProvider 获取CSRF token提供者
func GetCSRFTokenProvider() CSRFTokenProvider {
	csrfProviderMutex.RLock()
	defer csrfProviderMutex.RUnlock()
	return globalCSRFProvider
}

// RenderData 渲染数据结构
type RenderData struct {
	Data           any               `json:"data"`
	Meta           *MetaData         `json:"meta,omitempty"`
	Flash          *FlashData        `json:"flash,omitempty"`
	CSRF           string            `json:"csrf,omitempty"`
	CsrfToken      string            `json:"csrf_token,omitempty"` // 添加驼峰命名字段，用于模板中的{{.CsrfToken}}访问
	Theme          string            `json:"theme,omitempty"`
	User           any               `json:"user,omitempty"`
	Request        *RequestData      `json:"request,omitempty"`
	LayoutSections map[string]string `json:"layout_sections,omitempty"` // 布局区块内容
}

// GetCSRFToken 获取CSRF令牌（别名方法）
func (r *RenderData) GetCSRFToken() string {
	return r.CSRF
}

// Csrf_token 获取CSRF令牌（下划线命名）
func (r *RenderData) Csrf_token() string {
	return r.CSRF
}

// MetaData 页面元数据
type MetaData struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    string            `json:"keywords"`
	Author      string            `json:"author"`
	Canonical   string            `json:"canonical"`
	Image       string            `json:"image"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// FlashData 闪存消息
type FlashData struct {
	Success []string `json:"success,omitempty"`
	Error   []string `json:"error,omitempty"`
	Warning []string `json:"warning,omitempty"`
	Info    []string `json:"info,omitempty"`
}

// RequestData 请求信息
type RequestData struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Query     string `json:"query"`
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
	Timestamp int64  `json:"timestamp"`
}