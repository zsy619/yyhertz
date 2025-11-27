package context

import (
	"net/http"
)

// ============= Cookie兼容性方法 (beego风格) =============

// GetCookie 获取Cookie (beego兼容方法)
func (ctx *Context) GetCookie(key string) string {
	return ctx.Input.Cookie(key)
}

// SetCookie 设置Cookie (beego兼容方法)
func (ctx *Context) SetCookie(name, value string, others ...any) {
	ctx.Input.SetCookie(name, value, others...)
}

// GetSecureCookie 获取安全Cookie (beego兼容方法)
func (ctx *Context) GetSecureCookie(secret, key string) (string, bool) {
	return ctx.Input.GetSecureCookie(secret, key)
}

// SetSecureCookie 设置安全Cookie (beego兼容方法)
func (ctx *Context) SetSecureCookie(secret, name, value string, others ...any) {
	ctx.Input.SetSecureCookie(secret, name, value, others...)
}

// ============= Cookie增强方法 =============

// Cookie 获取Cookie值 (简化版)
func (ctx *Context) Cookie(name string) (string, error) {
	if !ctx.ensureRequest() {
		return "", ErrRequestNotFound
	}

	value := safeStringConvert(ctx.request.Cookie(name))
	if value == "" {
		return "", http.ErrNoCookie
	}
	return value, nil
}

// SetSameSiteCookie 设置带SameSite属性的Cookie
func (ctx *Context) SetSameSiteCookie(name, value string, maxAge int, path, domain string, sameSite http.SameSite, secure, httpOnly bool) {
	if !ctx.ensureRequest() {
		return
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	}

	ctx.request.Response.Header.Add("Set-Cookie", cookie.String())
}

// DeleteCookie 删除Cookie
func (ctx *Context) DeleteCookie(name string, path ...string) {
	cookiePath := "/"
	if len(path) > 0 {
		cookiePath = path[0]
	}

	ctx.SetSameSiteCookie(name, "", -1, cookiePath, "", http.SameSiteDefaultMode, false, false)
}

// HasCookie 检查Cookie是否存在
func (ctx *Context) HasCookie(name string) bool {
	_, err := ctx.Cookie(name)
	return err == nil
}

// ============= 数据操作兼容性方法 =============

// SetData 设置数据 (beego兼容)
func (ctx *Context) SetData(key string, value any) {
	ctx.Set(key, value)
}

// GetData 获取数据 (beego兼容)
func (ctx *Context) GetData(key string) (any, bool) {
	return ctx.Get(key)
}

// DelData 删除数据 (beego兼容)
func (ctx *Context) DelData(key string) {
	ctx.keys.Delete(key)
}

// ============= 错误处理兼容性方法 =============

// CustomAbort 自定义中止 (beego兼容)
func (ctx *Context) CustomAbort(status int, body string) {
	ctx.AbortWithStatus(status)
	if body != "" {
		ctx.String(status, body)
	}
}

// StopRun 停止运行 (beego兼容)
func (ctx *Context) StopRun() {
	ctx.Abort()
}

// Error 返回错误响应 (简化版)
func (ctx *Context) Error(code int, message string) {
	ctx.JSONError(code, message)
}

// ============= 兼容性访问器方法 =============
// 注意：这些方法已在context.go中定义，这里不再重复定义

// ============= 其他兼容性方法 =============

// Ctx 获取Context自身 (beego兼容)
func (ctx *Context) Ctx() *Context {
	return ctx
}

// Controller 兼容性方法，返回nil
func (ctx *Context) Controller() any {
	return nil
}

// Handler 兼容性方法，返回nil
func (ctx *Context) Handler() any {
	return nil
}
