// Package context provides framework compatibility layer.
//
// This file implements standard MVC framework compatibility adapters that allow
// seamless migration from various frameworks to YYHertz. The adapters maintain
// 100% API compatibility while leveraging YYHertz's modern, high-performance architecture.
//
// Key Features:
// - Zero-disruption migration from traditional MVC frameworks
// - Full session compatibility with standard session interfaces
// - Enhanced performance through YYHertz native optimizations
// - Unified Context architecture for better type safety
// - Extended functionality beyond traditional framework limitations
//
// Migration Benefits:
// - No code changes required for basic functionality
// - Significant performance improvements
// - Access to YYHertz's advanced features
// - Modern Go best practices and patterns
// - Seamless integration with YYHertz ecosystem
//
// NOTE: Most session and cookie functionality has been moved to the session package
// for better modularity. This file now provides compatibility proxies.
package context

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

// InputData 标准MVC风格输入数据结构
// 现在通过session包提供的扩展来实现cookie和session功能
type InputData struct {
	ctx       *Context
	extension *session.ContextExtension // session包扩展
}

// OutputData 标准MVC风格输出数据结构
type OutputData struct {
	ctx       *Context
	extension *session.ContextExtension // session包扩展
}

// ============= OutputData方法 =============

// Cookie 设置Cookie (Output兼容性方法)
func (o *OutputData) Cookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if o.ctx.Request != nil {
		o.ctx.Request.SetCookie(name, value, maxAge, path, domain, protocol.CookieSameSiteDefaultMode, secure, httpOnly)
	}
}

// SetStatus 设置响应状态码 (Output兼容性方法)
func (o *OutputData) SetStatus(code int) {
	if o.ctx.Request != nil {
		o.ctx.Request.SetStatusCode(code)
	}
}

// JSON 输出JSON响应 (Output兼容性方法)
func (o *OutputData) JSON(data interface{}) error {
	if o.ctx.Request != nil {
		o.ctx.Request.JSON(200, data)
	}
	return nil
}

// Header 设置响应头 (Output兼容性方法)
func (o *OutputData) Header(key, value string) {
	if o.ctx.Request != nil {
		o.ctx.Request.Response.Header.Set(key, value)
	}
}

// ============= OutputData Session方法 =============

// getExtension 获取session扩展（延迟初始化）(OutputData版本)
func (o *OutputData) getExtension() *session.ContextExtension {
	if o.extension == nil && o.ctx != nil && o.ctx.Request != nil {
		// 创建session扩展
		o.extension = session.NewExtensionForHertzContext(o.ctx.Request)
	}
	return o.extension
}

// SetSession 设置session数据 (Output兼容性方法) - 代理到session包
func (o *OutputData) SetSession(key string, value interface{}) error {
	if ext := o.getExtension(); ext != nil {
		return ext.SetSession(key, value)
	}
	return nil
}

// GetSession 获取session数据 (Output兼容性方法) - 代理到session包
func (o *OutputData) GetSession(key string) interface{} {
	if ext := o.getExtension(); ext != nil {
		return ext.GetSession(key)
	}
	return nil
}

// Session 获取或设置session数据的便捷方法 (OutputData)
// 用法：
//   value := ctx.Output.Session("adminId")          // 获取值
//   ctx.Output.Session("adminId", "12345")          // 设置值
func (o *OutputData) Session(key string, values ...interface{}) interface{} {
	if len(values) == 0 {
		// 获取模式
		return o.GetSession(key)
	} else {
		// 设置模式
		o.SetSession(key, values[0])
		return values[0]
	}
}

// ============= InputData基础方法 =============

// Scheme 获取请求协议
func (i *InputData) Scheme() string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.URI().Scheme())
	}
	return "http"
}

// Domain 获取请求域名 (Input兼容性方法)
func (i *InputData) Domain() string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.Host())
	}
	return ""
}

// Host 获取请求主机 (Input兼容性方法)
func (i *InputData) Host() string {
	return i.Domain()
}

// Method 获取请求方法 (Input兼容性方法)
func (i *InputData) Method() string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.Method())
	}
	return "GET"
}

// IP 获取客户端IP地址 (Input兼容性方法)
func (i *InputData) IP() string {
	if i.ctx.Request != nil {
		return i.ctx.Request.ClientIP()
	}
	return ""
}

// UserAgent 获取User-Agent (Input兼容性方法)
func (i *InputData) UserAgent() string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.Request.Header.Peek("User-Agent"))
	}
	return ""
}

// IsAjax 判断是否为Ajax请求 (Input兼容性方法)
func (i *InputData) IsAjax() bool {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.Request.Header.Peek("X-Requested-With")) == "XMLHttpRequest"
	}
	return false
}

// URL 获取请求URL (Input兼容性方法)
func (i *InputData) URL() string {
	if i.ctx.Request != nil {
		return i.ctx.Request.URI().String()
	}
	return ""
}

// Is 检查请求方法 (Input兼容性方法)
func (i *InputData) Is(method string) bool {
	return i.Method() == method
}

// IsGet 检查是否为GET请求 (Input兼容性方法)
func (i *InputData) IsGet() bool {
	return i.Is("GET")
}

// IsPost 检查是否为POST请求 (Input兼容性方法)
func (i *InputData) IsPost() bool {
	return i.Is("POST")
}

// Query 获取查询参数 (Input兼容性方法)
func (i *InputData) Query(key string) string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.QueryArgs().Peek(key))
	}
	return ""
}

// Param 获取路径参数 (Input兼容性方法)
func (i *InputData) Param(key string) string {
	if i.ctx.Request != nil {
		return i.ctx.Request.Param(key)
	}
	return ""
}

// Header 获取请求头 (Input兼容性方法)
func (i *InputData) Header(key string) string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.GetHeader(key))
	}
	return ""
}


// Referer 获取来源页面 (Input兼容性方法)
func (i *InputData) Referer() string {
	return i.Header("Referer")
}

// Data 设置上下文数据 (Input兼容性方法)
func (i *InputData) Data(key string, val interface{}) {
	if i.ctx != nil {
		i.ctx.Set(key, val)
	}
}

// GetData 获取上下文数据 (Input兼容性方法)
func (i *InputData) GetData(key string) interface{} {
	if i.ctx != nil {
		if val, exists := i.ctx.Get(key); exists {
			return val
		}
	}
	return nil
}

// ============= Cookie方法代理 (代理到session包) =============

// getExtension 获取session扩展（延迟初始化）
func (i *InputData) getExtension() *session.ContextExtension {
	if i.extension == nil && i.ctx != nil && i.ctx.Request != nil {
		// 创建session扩展
		i.extension = session.NewExtensionForHertzContext(i.ctx.Request)
	}
	return i.extension
}

// Cookie 获取Cookie (Input兼容性方法) - 代理到session包
func (i *InputData) Cookie(key string) string {
	if ext := i.getExtension(); ext != nil {
		return ext.GetCookie(key)
	}
	return ""
}

// SetCookie 设置Cookie (beego兼容方法) - 代理到session包
func (i *InputData) SetCookie(name, value string, others ...interface{}) {
	if ext := i.getExtension(); ext != nil {
		ext.SetCookie(name, value, others...)
	}
}

// GetSecureCookie 获取安全Cookie (beego兼容方法) - 代理到session包
func (i *InputData) GetSecureCookie(secret, key string) (string, bool) {
	if ext := i.getExtension(); ext != nil {
		return ext.GetSecureCookie(secret, key)
	}
	return "", false
}

// SetSecureCookie 设置安全Cookie (beego兼容方法) - 代理到session包
func (i *InputData) SetSecureCookie(secret, name, value string, others ...interface{}) {
	if ext := i.getExtension(); ext != nil {
		ext.SetSecureCookie(secret, name, value, others...)
	}
}

// DelCookie 删除Cookie (beego兼容方法) - 代理到session包
func (i *InputData) DelCookie(name string) {
	if ext := i.getExtension(); ext != nil {
		ext.DelCookie(name)
	}
}

// CookieExists 检查Cookie是否存在 - 代理到session包
func (i *InputData) CookieExists(key string) bool {
	if ext := i.getExtension(); ext != nil {
		return ext.CookieExists(key)
	}
	return false
}

// GetCookies 获取所有Cookie - 代理到session包
func (i *InputData) GetCookies() map[string]string {
	if ext := i.getExtension(); ext != nil && ext.Cookie != nil {
		return ext.Cookie.GetAll()
	}
	return make(map[string]string)
}

// ClearAllCookies 清除所有Cookie - 代理到session包
func (i *InputData) ClearAllCookies() {
	if ext := i.getExtension(); ext != nil && ext.Cookie != nil {
		ext.Cookie.Clear()
	}
}

// CookieCount 获取Cookie数量 - 代理到session包
func (i *InputData) CookieCount() int {
	if ext := i.getExtension(); ext != nil && ext.Cookie != nil {
		return ext.Cookie.Count()
	}
	return 0
}

// ValidateSecureCookie 验证安全Cookie但不返回值 - 代理到session包
func (i *InputData) ValidateSecureCookie(secret, key string) bool {
	if ext := i.getExtension(); ext != nil && ext.SecureCookie != nil {
		return ext.SecureCookie.Validate(secret, key)
	}
	return false
}

// ============= 高级Cookie功能代理 =============

// SetSecureCookieWithOptions 设置安全Cookie (增强版本) - 代理到session包
func (i *InputData) SetSecureCookieWithOptions(name, value string, options session.CookieSecurityOptions, others ...interface{}) error {
	if ext := i.getExtension(); ext != nil && ext.SecureCookie != nil {
		return ext.SecureCookie.SetSecureWithOptions(name, value, options, others...)
	}
	return nil
}

// GetSecureCookieWithOptions 获取安全Cookie (增强版本) - 代理到session包
func (i *InputData) GetSecureCookieWithOptions(key string, options session.CookieSecurityOptions) (string, bool, error) {
	if ext := i.getExtension(); ext != nil && ext.SecureCookie != nil {
		return ext.SecureCookie.GetSecureWithOptions(key, options)
	}
	return "", false, nil
}

// ============= Session方法代理 (代理到session包) =============

// StartSession 启动Session (标准MVC风格接口) - 代理到session包
func (i *InputData) StartSession() *session.Adapter {
	if ext := i.getExtension(); ext != nil {
		return ext.StartSession()
	}
	return session.NewAdapter(nil, i.ctx)
}

// SetSession 设置session数据 (标准MVC兼容性方法) - 代理到session包
func (i *InputData) SetSession(key string, value interface{}) error {
	if ext := i.getExtension(); ext != nil {
		return ext.SetSession(key, value)
	}
	return nil
}

// GetSession 获取session数据 (标准MVC兼容性方法) - 代理到session包
func (i *InputData) GetSession(key string) interface{} {
	if ext := i.getExtension(); ext != nil {
		return ext.GetSession(key)
	}
	return nil
}

// DelSession 删除session数据 (标准MVC兼容性方法) - 代理到session包
func (i *InputData) DelSession(key string) error {
	if ext := i.getExtension(); ext != nil {
		return ext.DelSession(key)
	}
	return nil
}

// GetSessionID 获取session ID (beego兼容性方法) - 代理到session包
func (i *InputData) GetSessionID() string {
	if ext := i.getExtension(); ext != nil {
		return ext.GetSessionID()
	}
	return ""
}

// IsSessionStarted 检查session是否已启动 (beego兼容性方法) - 代理到session包
func (i *InputData) IsSessionStarted() bool {
	if ext := i.getExtension(); ext != nil {
		return ext.IsSessionStarted()
	}
	return false
}

// DestroySession 销毁session (标准MVC兼容性方法) - 代理到session包
func (i *InputData) DestroySession() {
	if ext := i.getExtension(); ext != nil {
		ext.DestroySession()
	}
}

// ClearSession 清空session数据 (标准MVC兼容性方法) - 代理到session包
func (i *InputData) ClearSession() {
	if ext := i.getExtension(); ext != nil {
		ext.ClearSession()
	}
}

// SaveSession 保存session数据 (beego兼容性方法) - 代理到session包
func (i *InputData) SaveSession() error {
	if ext := i.getExtension(); ext != nil {
		return ext.SaveSession()
	}
	return nil
}

// SessionRegenerateID 重新生成session ID (标准MVC兼容性方法) - 代理到session包
func (i *InputData) SessionRegenerateID() {
	if ext := i.getExtension(); ext != nil {
		ext.SessionRegenerateID()
	}
}

// Session 获取或设置session数据的便捷方法 (InputData)
// 用法：
//   value := ctx.Input.Session("adminId")          // 获取值
//   ctx.Input.Session("adminId", "12345")          // 设置值
func (i *InputData) Session(key string, values ...interface{}) interface{} {
	if len(values) == 0 {
		// 获取模式
		return i.GetSession(key)
	} else {
		// 设置模式
		i.SetSession(key, values[0])
		return values[0]
	}
}

// ============= 兼容性别名和附加功能 =============

// GetSessionExtension 获取session扩展对象 (提供对session包完整功能的访问)
func (i *InputData) GetSessionExtension() *session.ContextExtension {
	return i.getExtension()
}

// WithSessionExtension 使用自定义session扩展
func (i *InputData) WithSessionExtension(ext *session.ContextExtension) *InputData {
	i.extension = ext
	return i
}

// ============= 向后兼容性类型别名 =============

// SessionStore 兼容性类型别名，指向session包的Adapter
type SessionStore = session.Adapter

// NewSessionStore 兼容性函数，创建session适配器
func NewSessionStore(store session.Store, ctx *Context) *SessionStore {
	return session.NewAdapter(store, ctx)
}

// CookieSecurityOptions 兼容性类型别名
type CookieSecurityOptions = session.CookieSecurityOptions

// ============= 初始化方法 =============

// Initialize 初始化InputData (兼容性方法)
// 用于在已创建的InputData实例上重新初始化Context
func (i *InputData) Initialize(c *app.RequestContext) {
	if i.ctx == nil {
		i.ctx = &Context{
			Request:        c,
			RequestContext: c,
		}
	} else {
		i.ctx.Request = c
		i.ctx.RequestContext = c
	}
	// 重置extension以便重新初始化
	i.extension = nil
}

// Initialize 初始化OutputData (兼容性方法)
// 用于在已创建的OutputData实例上重新初始化Context
func (o *OutputData) Initialize(c *app.RequestContext) {
	if o.ctx == nil {
		o.ctx = &Context{
			Request:        c,
			RequestContext: c,
		}
	} else {
		o.ctx.Request = c
		o.ctx.RequestContext = c
	}
	// 重置extension以便重新初始化
	o.extension = nil
}

// GetContext 获取关联的Context (InputData)
func (i *InputData) GetContext() *Context {
	return i.ctx
}

// GetContext 获取关联的Context (OutputData)
func (o *OutputData) GetContext() *Context {
	return o.ctx
}

// ============= 便利构造函数 =============

// NewInputData 创建InputData实例
func NewInputData(ctx *Context) *InputData {
	return &InputData{
		ctx: ctx,
		// extension 延迟初始化
	}
}

// NewOutputData 创建OutputData实例
func NewOutputData(ctx *Context) *OutputData {
	return &OutputData{
		ctx: ctx,
	}
}

// ============= 文档和迁移说明 =============

/*
迁移说明:

1. 原有代码100%兼容，无需修改
2. 新功能建议直接使用session包：
   import "github.com/zsy619/yyhertz/framework/mvc/session"
   ext := session.NewContextExtension(ctx)

3. 初始化方法（兼容性）：
   inputData := &InputData{}
   outputData := &OutputData{}
   inputData.Initialize(hertzCtx)   // 初始化InputData
   outputData.Initialize(hertzCtx)  // 初始化OutputData
   ctx := inputData.GetContext()    // 获取关联的Context

4. 高级功能访问：
   ext := inputData.GetSessionExtension()
   options := session.CookieSecurityOptions{...}
   ext.SecureCookie.SetSecureWithOptions(...)

5. 性能优化：
   - session扩展延迟初始化，不使用时无性能开销
   - 直接使用session包可避免代理层开销

6. 功能对照：
   InputData.Cookie() -> session.BaseCookie.Get()
   InputData.SetSecureCookie() -> session.SecureCookie.SetSecure()
   InputData.StartSession() -> session.ContextExtension.StartSession()

7. 新增方法：
   - InputData.Initialize(c *app.RequestContext)  // 初始化方法
   - OutputData.Initialize(c *app.RequestContext) // 初始化方法
   - InputData.GetContext() *Context              // 获取Context
   - OutputData.GetContext() *Context             // 获取Context

8. 便捷Session方法（v2.0新增）：
   InputData:
   - Session(key) -> GetSession(key)              // 获取session值
   - Session(key, value) -> SetSession(key, value) // 设置session值
   
   OutputData（新增完整session支持）:
   - Session(key) -> GetSession(key)              // 获取session值
   - Session(key, value) -> SetSession(key, value) // 设置session值
   - SetSession(key, value) error                 // 设置session数据
   - GetSession(key) interface{}                  // 获取session数据
   
   使用示例：
   adminId := ctx.Input.Session("adminId")        // 获取
   ctx.Input.Session("adminId", "12345")          // 设置
   userId := ctx.Output.Session("userId")         // 获取
   ctx.Output.Session("userId", "67890")          // 设置
*/