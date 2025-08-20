package session

import (
	"github.com/cloudwego/hertz/pkg/app"
	
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
)

// ContextExtension Context的session/cookie扩展
// 为各种Context类型提供统一的session和cookie操作接口
type ContextExtension struct {
	Cookie         *cookie.BaseCookie     // 基础cookie操作
	SecureCookie   *cookie.SecureCookie   // 安全cookie操作
	OutputCookie   *cookie.OutputCookie   // 输出cookie操作
	SessionMgr     *SessionManager        // session管理器
	context        any                    // 关联的上下文
	currentSession *Adapter               // 当前session适配器
}

// NewExtensionForHertzContext 为Hertz RequestContext创建扩展
func NewExtensionForHertzContext(ctx *app.RequestContext) *ContextExtension {
	return &ContextExtension{
		Cookie:       cookie.NewBaseCookie(ctx),
		SecureCookie: cookie.NewSecureCookie(ctx),
		OutputCookie: cookie.NewOutputCookie(ctx),
		SessionMgr:   NewSessionManager(nil), // 使用默认配置避免测试环境配置问题
		context:      ctx,
	}
}

// NewExtensionForYYHertzContext 为YYHertz Context创建扩展
// 注意：这里使用any来避免循环导入，实际使用时需要类型断言
func NewExtensionForYYHertzContext(ctx any) *ContextExtension {
	// 尝试从YYHertz Context获取Hertz RequestContext
	var hertzCtx *app.RequestContext

	// 这里需要根据实际的YYHertz Context结构进行适配
	// 暂时使用any避免循环导入问题
	if yyCtx, ok := ctx.(interface{ GetRequestContext() *app.RequestContext }); ok {
		hertzCtx = yyCtx.GetRequestContext()
	}

	extension := &ContextExtension{
		SessionMgr: NewSessionManager(nil), // 使用默认配置
		context:    ctx,
	}

	if hertzCtx != nil {
		extension.Cookie = cookie.NewBaseCookie(hertzCtx)
		extension.SecureCookie = cookie.NewSecureCookie(hertzCtx)
		extension.OutputCookie = cookie.NewOutputCookie(hertzCtx)
	}

	return extension
}

// ============= Session便利方法 =============

// StartSession 启动Session (标准MVC风格接口)
// 返回标准框架兼容的session adapter，支持所有标准MVC session API
func (ce *ContextExtension) StartSession() *Adapter {
	if !ce.SessionMgr.IsEnabled() {
		return NewAdapter(nil, ce.context)
	}

	// 尝试从上下文获取现有session
	if sessionID := ce.getSessionIDFromContext(); sessionID != "" {
		adapter := ce.SessionMgr.GetSession(ce.context, sessionID)
		ce.currentSession = adapter
		return adapter
	}

	// 创建新的session
	adapter := ce.SessionMgr.CreateSession(ce.context)
	ce.currentSession = adapter

	// 设置session cookie
	if ce.Cookie != nil && adapter.SessionID() != "" {
		config := ce.SessionMgr.GetConfig()
		ce.Cookie.Set(config.CookieName, adapter.SessionID(),
			config.MaxAge, config.CookiePath, config.CookieDomain,
			config.Secure, config.HttpOnly)
	}

	return adapter
}

// SetSession 设置session数据 (标准MVC兼容性方法)
func (ce *ContextExtension) SetSession(key string, value any) error {
	adapter := ce.getOrCreateSession()
	return adapter.Set(key, value)
}

// GetSession 获取session数据 (标准MVC兼容性方法)
func (ce *ContextExtension) GetSession(key string) any {
	adapter := ce.getOrCreateSession()
	return adapter.Get(key)
}

// DelSession 删除session数据 (标准MVC兼容性方法)
func (ce *ContextExtension) DelSession(key string) error {
	adapter := ce.getOrCreateSession()
	return adapter.Delete(key)
}

// GetSessionID 获取session ID (beego兼容性方法)
func (ce *ContextExtension) GetSessionID() string {
	if ce.currentSession != nil {
		return ce.currentSession.SessionID()
	}
	return ce.getSessionIDFromContext()
}

// IsSessionStarted 检查session是否已启动 (beego兼容性方法)
func (ce *ContextExtension) IsSessionStarted() bool {
	return ce.currentSession != nil && ce.currentSession.IsStarted()
}

// DestroySession 销毁session (标准MVC兼容性方法)
func (ce *ContextExtension) DestroySession() {
	if ce.currentSession != nil {
		ce.currentSession.Destroy()
		ce.currentSession = nil
	}

	// 删除session cookie
	if ce.Cookie != nil {
		config := ce.SessionMgr.GetConfig()
		ce.Cookie.Delete(config.CookieName)
	}

	// 清除上下文中的session引用
	ce.clearSessionFromContext()
}

// ClearSession 清空session数据 (标准MVC兼容性方法，别名)
func (ce *ContextExtension) ClearSession() {
	if ce.currentSession != nil {
		ce.currentSession.Flush()
	}
}

// SaveSession 保存session数据 (beego兼容性方法)
func (ce *ContextExtension) SaveSession() error {
	if ce.currentSession != nil {
		return ce.currentSession.Save()
	}
	return nil
}

// SessionRegenerateID 重新生成session ID (标准MVC兼容性方法)
func (ce *ContextExtension) SessionRegenerateID() {
	// 这个功能需要session manager支持，暂时保持为空实现
	// 在实际使用中，应该通过session manager的SessionRegenerate方法实现
	if ce.currentSession != nil {
		// 销毁旧session
		oldSessionID := ce.currentSession.SessionID()
		ce.currentSession.Destroy()

		// 创建新session
		ce.currentSession = ce.SessionMgr.CreateSession(ce.context)

		// 更新cookie
		if ce.Cookie != nil && ce.currentSession.SessionID() != "" {
			config := ce.SessionMgr.GetConfig()
			ce.Cookie.Set(config.CookieName, ce.currentSession.SessionID(),
				config.MaxAge, config.CookiePath, config.CookieDomain,
				config.Secure, config.HttpOnly)
		}

		// 可选：记录日志
		_ = oldSessionID // 避免未使用变量警告
	}
}

// ============= Cookie便利方法 =============

// GetCookie 获取Cookie (beego兼容方法)
func (ce *ContextExtension) GetCookie(key string) string {
	if ce.Cookie != nil {
		return ce.Cookie.Get(key)
	}
	return ""
}

// SetCookie 设置Cookie (beego兼容方法)
func (ce *ContextExtension) SetCookie(name, value string, others ...any) {
	if ce.Cookie != nil {
		ce.Cookie.Set(name, value, others...)
	}
}

// GetSecureCookie 获取安全Cookie (beego兼容方法)
func (ce *ContextExtension) GetSecureCookie(secret, key string) (string, bool) {
	if ce.SecureCookie != nil {
		return ce.SecureCookie.GetSecure(secret, key)
	}
	return "", false
}

// SetSecureCookie 设置安全Cookie (beego兼容方法)
func (ce *ContextExtension) SetSecureCookie(secret, name, value string, others ...any) {
	if ce.SecureCookie != nil {
		ce.SecureCookie.SetSecure(secret, name, value, others...)
	}
}

// DelCookie 删除Cookie
func (ce *ContextExtension) DelCookie(name string) {
	if ce.Cookie != nil {
		ce.Cookie.Delete(name)
	}
}

// CookieExists 检查Cookie是否存在
func (ce *ContextExtension) CookieExists(key string) bool {
	if ce.Cookie != nil {
		return ce.Cookie.Exists(key)
	}
	return false
}

// ============= 高级功能 =============

// GetSessionAdapter 获取当前session适配器
func (ce *ContextExtension) GetSessionAdapter() *Adapter {
	return ce.currentSession
}

// SetSessionAdapter 设置session适配器
func (ce *ContextExtension) SetSessionAdapter(adapter *Adapter) {
	ce.currentSession = adapter
}

// GetSessionManager 获取session管理器
func (ce *ContextExtension) GetSessionManager() *SessionManager {
	return ce.SessionMgr
}

// ============= 内部方法 =============

// getOrCreateSession 获取或创建session
func (ce *ContextExtension) getOrCreateSession() *Adapter {
	if ce.currentSession == nil {
		ce.currentSession = ce.StartSession()
	}
	return ce.currentSession
}

// getSessionIDFromContext 从上下文获取session ID
func (ce *ContextExtension) getSessionIDFromContext() string {
	if ce.Cookie != nil {
		config := ce.SessionMgr.GetConfig()
		return ce.Cookie.Get(config.CookieName)
	}
	return ""
}

// clearSessionFromContext 清除上下文中的session引用
func (ce *ContextExtension) clearSessionFromContext() {
	// 这里需要根据具体的上下文类型进行处理
	// 当前简化实现，主要是清理内存引用
	ce.currentSession = nil
}

// ============= 全局便利函数 =============

// NewContextExtension 创建上下文扩展（自动检测类型）
func NewContextExtension(ctx any) *ContextExtension {
	// 尝试类型断言
	if hertzCtx, ok := ctx.(*app.RequestContext); ok {
		return NewExtensionForHertzContext(hertzCtx)
	}

	// 默认使用YYHertz Context扩展
	return NewExtensionForYYHertzContext(ctx)
}

// GetExtension 从上下文获取或创建扩展
// 这是一个便利函数，可以在中间件中使用
func GetExtension(ctx any) *ContextExtension {
	// 这里可以实现缓存逻辑，避免重复创建扩展
	// 当前简化实现，每次都创建新的扩展
	return NewContextExtension(ctx)
}
