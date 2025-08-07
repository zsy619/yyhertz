package context

import (
	"github.com/cloudwego/hertz/pkg/protocol"
)

// InputData Beego风格输入数据结构
type InputData struct {
	ctx *Context
}

// OutputData Beego风格输出数据结构
type OutputData struct {
	ctx *Context
}

// Cookie 设置Cookie (Output兼容性方法)
func (o *OutputData) Cookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if o.ctx.Request != nil {
		o.ctx.Request.SetCookie(name, value, maxAge, path, domain, protocol.CookieSameSiteDefaultMode, secure, httpOnly)
	}
}

// Header 设置响应头 (Output兼容性方法)
func (o *OutputData) Header(key, value string) {
	if o.ctx.Request != nil {
		o.ctx.Request.Response.Header.Set(key, value)
	}
}

// Status 设置状态码 (Output兼容性方法)
func (o *OutputData) Status(code int) {
	if o.ctx.Request != nil {
		o.ctx.Request.Response.SetStatusCode(code)
	}
}

// Body 设置响应体 (Output兼容性方法)
func (o *OutputData) Body(content []byte) error {
	if o.ctx.Request != nil {
		o.ctx.Request.Response.SetBody(content)
	}
	return nil
}

// JSON 设置JSON响应 (Output兼容性方法)
func (o *OutputData) JSON(data interface{}, hasIndent bool, coding ...bool) error {
	if o.ctx.Request != nil {
		o.ctx.Request.JSON(200, data)
	}
	return nil
}

// SetStatus 设置状态码 (Output兼容性方法，别名)
func (o *OutputData) SetStatus(code int) {
	o.Status(code)
}

// Param 获取路由参数 (Input兼容性方法)
func (i *InputData) Param(key string) string {
	return i.ctx.Params.ByName(key)
}

// Query 获取查询参数 (Input兼容性方法)
func (i *InputData) Query(key string) string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.QueryArgs().Peek(key))
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

// Cookie 获取Cookie (Input兼容性方法)
func (i *InputData) Cookie(key string) string {
	if i.ctx.Request != nil {
		return string(i.ctx.Request.Cookie(key))
	}
	return ""
}

// Data 设置上下文数据 (Input兼容性方法)
func (i *InputData) Data(key string, val interface{}) {
	if i.ctx != nil {
		i.ctx.Keys[key] = val
	}
}

// RequestBody 获取请求体数据 (Input兼容性方法)
func (i *InputData) RequestBody() []byte {
	if i.ctx.Request != nil {
		body, _ := i.ctx.Request.Body()
		return body
	}
	return nil
}

// IP 获取客户端IP (Input兼容性方法)
func (i *InputData) IP() string {
	if i.ctx.Request != nil {
		return i.ctx.Request.ClientIP()
	}
	return ""
}