package cookie

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/config"
)

// Options Cookie配置选项
type Options struct {
	MaxAge   int    // 过期时间（秒）
	Path     string // 路径
	Domain   string // 域名
	Secure   bool   // 是否仅HTTPS
	HttpOnly bool   // 是否仅HTTP（防XSS）
	SameSite string // SameSite策略: "Strict", "Lax", "None"
}

// DefaultOptions 默认Cookie配置
func DefaultOptions() *Options {
	return &Options{
		MaxAge:   3600,    // 1小时
		Path:     "/",     // 根路径
		Domain:   "",      // 当前域名
		Secure:   false,   // 开发环境通常不用HTTPS
		HttpOnly: true,    // 防XSS攻击
		SameSite: "Lax",   // 防CSRF攻击
	}
}

// Config Cookie全局配置
type Config struct {
	DefaultMaxAge int    // 默认过期时间
	DefaultPath   string // 默认路径
	DefaultDomain string // 默认域名
	DefaultSecure bool   // 默认是否仅HTTPS
	HttpOnly      bool   // 默认是否仅HTTP
	SameSite      string // 默认SameSite策略
}

// DefaultConfig 默认Cookie配置
func DefaultConfig() *Config {
	return &Config{
		DefaultMaxAge: 3600,
		DefaultPath:   "/",
		DefaultDomain: "",
		DefaultSecure: false,
		HttpOnly:      true,
		SameSite:      "Lax",
	}
}

// LoadFromConfig 从配置文件加载Cookie配置
func LoadFromConfig() *Config {
	// 获取session配置
	sessionConfig, err := config.LoadConfigWithGeneric[config.SessionConfig]("session")
	if err != nil {
		config.Warnf("Failed to load session config, using defaults: %v", err)
		return DefaultConfig()
	}

	cookieConfig := &Config{
		DefaultMaxAge: sessionConfig.Cookie.MaxAge,
		DefaultPath:   sessionConfig.Cookie.Path,
		DefaultDomain: sessionConfig.Cookie.Domain,
		DefaultSecure: sessionConfig.Cookie.Secure,
		HttpOnly:      sessionConfig.Cookie.HttpOnly,
		SameSite:      sessionConfig.Cookie.SameSite,
	}

	config.Infof("Cookie config loaded from session.yaml: maxAge=%d, path=%s, secure=%t", 
		cookieConfig.DefaultMaxAge, cookieConfig.DefaultPath, cookieConfig.DefaultSecure)
	
	return cookieConfig
}

// Helper Cookie辅助工具
type Helper struct {
	config *Config
}

// NewHelper 创建Cookie辅助工具
func NewHelper(config *Config) *Helper {
	if config == nil {
		config = DefaultConfig()
	}
	return &Helper{
		config: config,
	}
}

// NewHelperFromConfig 从配置文件创建Cookie辅助工具
func NewHelperFromConfig() *Helper {
	config := LoadFromConfig()
	return &Helper{
		config: config,
	}
}

// Set 设置Cookie（增强版）
func (h *Helper) Set(ctx *app.RequestContext, name, value string, options ...*Options) {
	// 使用默认配置或自定义配置
	var opts *Options
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = DefaultOptions()
	}

	// 构建cookie字符串
	cookie := fmt.Sprintf("%s=%s; Max-Age=%d; Path=%s", name, value, opts.MaxAge, opts.Path)
	
	if opts.Domain != "" {
		cookie += "; Domain=" + opts.Domain
	}
	if opts.Secure {
		cookie += "; Secure"
	}
	if opts.HttpOnly {
		cookie += "; HttpOnly"
	}
	if opts.SameSite != "" {
		cookie += "; SameSite=" + opts.SameSite
	}

	ctx.Header("Set-Cookie", cookie)
}

// SetSimple 设置简单Cookie（兼容原有接口）
func (h *Helper) SetSimple(ctx *app.RequestContext, name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	opts := &Options{
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: "Lax",
	}
	h.Set(ctx, name, value, opts)
}

// Get 获取Cookie
func (h *Helper) Get(ctx *app.RequestContext, name string) string {
	return string(ctx.Cookie(name))
}

// GetWithDefault 获取Cookie（带默认值）
func (h *Helper) GetWithDefault(ctx *app.RequestContext, name, defaultValue string) string {
	value := h.Get(ctx, name)
	if value == "" {
		return defaultValue
	}
	return value
}

// Delete 删除Cookie
func (h *Helper) Delete(ctx *app.RequestContext, name string, path ...string) {
	cookiePath := "/"
	if len(path) > 0 {
		cookiePath = path[0]
	}
	
	opts := &Options{
		MaxAge:   -1,
		Path:     cookiePath,
		HttpOnly: true,
	}
	h.Set(ctx, name, "", opts)
}

// SetSecure 设置安全Cookie（用于生产环境）
func (h *Helper) SetSecure(ctx *app.RequestContext, name, value string, maxAge int) {
	opts := &Options{
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   true,  // 仅HTTPS
		HttpOnly: true,  // 防XSS
		SameSite: "Strict", // 严格的CSRF保护
	}
	h.Set(ctx, name, value, opts)
}

// SetSession 设置会话Cookie（浏览器关闭时过期）
func (h *Helper) SetSession(ctx *app.RequestContext, name, value string) {
	opts := &Options{
		MaxAge:   0, // 会话Cookie
		Path:     "/",
		HttpOnly: true,
		SameSite: "Lax",
	}
	h.Set(ctx, name, value, opts)
}

// Has 检查Cookie是否存在
func (h *Helper) Has(ctx *app.RequestContext, name string) bool {
	return h.Get(ctx, name) != ""
}

// SetWithGlobalConfig 使用全局配置设置Cookie
func (h *Helper) SetWithGlobalConfig(ctx *app.RequestContext, name, value string, maxAge ...int) {
	age := h.config.DefaultMaxAge
	if len(maxAge) > 0 {
		age = maxAge[0]
	}

	cookie := fmt.Sprintf("%s=%s; Max-Age=%d; Path=%s; HttpOnly=%t; SameSite=%s",
		name, value, age, h.config.DefaultPath, 
		h.config.HttpOnly, h.config.SameSite)
	
	if h.config.DefaultDomain != "" {
		cookie += "; Domain=" + h.config.DefaultDomain
	}
	if h.config.DefaultSecure {
		cookie += "; Secure"
	}
	
	ctx.Header("Set-Cookie", cookie)
}

// DeleteWithGlobalConfig 使用全局配置删除Cookie
func (h *Helper) DeleteWithGlobalConfig(ctx *app.RequestContext, name string) {
	cookie := fmt.Sprintf("%s=; Max-Age=-1; Path=%s",
		name, h.config.DefaultPath)
	
	if h.config.DefaultDomain != "" {
		cookie += "; Domain=" + h.config.DefaultDomain
	}
	
	ctx.Header("Set-Cookie", cookie)
}