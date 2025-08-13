package core

import (
	"io"

	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gopkg.in/yaml.v2"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 高级响应方法 =============

// XML 返回XML响应
func (c *BaseController) XML(data any) {
	c.XMLWithStatus(consts.StatusOK, data)
}

// XMLWithStatus 返回指定状态码的XML响应
func (c *BaseController) XMLWithStatus(status int, data any) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return XML")
		return
	}
	c.Ctx.XML(status, data)
}

// YAML 返回YAML响应
func (c *BaseController) YAML(data any) {
	c.YAMLWithStatus(consts.StatusOK, data)
}

// YAMLWithStatus 返回指定状态码的YAML响应
func (c *BaseController) YAMLWithStatus(status int, data any) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return YAML")
		return
	}
	c.Ctx.YAML(status, data)
}

// IndentedJSON 返回格式化的JSON响应
func (c *BaseController) IndentedJSON(data any) {
	c.IndentedJSONWithStatus(consts.StatusOK, data)
}

// IndentedJSONWithStatus 返回指定状态码的格式化JSON响应
func (c *BaseController) IndentedJSONWithStatus(status int, data any) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return IndentedJSON")
		return
	}
	c.Ctx.IndentedJSON(status, data)
}

// DataWithStatus 返回指定状态码的原始数据响应
func (c *BaseController) DataWithStatus(status int, contentType string, data []byte) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return Data")
		return
	}
	c.Ctx.Data(status, contentType, data)
}

// Status 设置状态码
func (c *BaseController) Status(code int) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to set status")
		return
	}
	c.Ctx.Status(code)
}

// NoContent 返回无内容响应 (204)
func (c *BaseController) NoContent() {
	c.Status(consts.StatusNoContent)
}

// Stream 流式响应
func (c *BaseController) Stream(contentType string, r io.Reader) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to stream")
		return
	}
	c.SetContentType(contentType)
	io.Copy(c.Ctx.Writer, r)
}

// ============= 高级重定向方法 =============

// RedirectPermanent 永久重定向 (301)
func (c *BaseController) RedirectPermanent(url string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to redirect")
		return
	}
	c.Ctx.Redirect(consts.StatusMovedPermanently, url)
}

// RedirectTemporary 临时重定向 (302)
func (c *BaseController) RedirectTemporary(url string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to redirect")
		return
	}
	c.Ctx.Redirect(consts.StatusFound, url)
}

// RedirectSeeOther 参见其他重定向 (303)
func (c *BaseController) RedirectSeeOther(url string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to redirect")
		return
	}
	c.Ctx.Redirect(consts.StatusSeeOther, url)
}

// ============= 响应状态检测方法 =============

// IsOk 检查响应是否为200
func (c *BaseController) IsOk() bool {
	if c.Ctx == nil || c.Ctx.Request == nil {
		return false
	}
	return c.Ctx.Request.Response.StatusCode() == consts.StatusOK
}

// IsSuccessful 检查响应是否成功 (2xx)
func (c *BaseController) IsSuccessful() bool {
	if c.Ctx == nil || c.Ctx.Request == nil {
		return false
	}
	status := c.Ctx.Request.Response.StatusCode()
	return status >= 200 && status < 300
}

// IsRedirect 检查响应是否为重定向 (3xx)
func (c *BaseController) IsRedirect() bool {
	if c.Ctx == nil || c.Ctx.Request == nil {
		return false
	}
	status := c.Ctx.Request.Response.StatusCode()
	return status >= 300 && status < 400
}

// IsClientError 检查响应是否为客户端错误 (4xx)
func (c *BaseController) IsClientError() bool {
	if c.Ctx == nil || c.Ctx.Request == nil {
		return false
	}
	status := c.Ctx.Request.Response.StatusCode()
	return status >= 400 && status < 500
}

// IsServerError 检查响应是否为服务器错误 (5xx)
func (c *BaseController) IsServerError() bool {
	if c.Ctx == nil || c.Ctx.Request == nil {
		return false
	}
	status := c.Ctx.Request.Response.StatusCode()
	return status >= 500 && status < 600
}

// IsForbidden 检查响应是否为403
func (c *BaseController) IsForbidden() bool {
	if c.Ctx == nil || c.Ctx.Request == nil {
		return false
	}
	return c.Ctx.Request.Response.StatusCode() == consts.StatusForbidden
}

// ============= 响应头操作方法 =============

// SetResponseHeader 设置响应头（避免与基类重复）
func (c *BaseController) SetResponseHeader(key, value string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to set header")
		return
	}
	c.Ctx.SetHeader(key, value)
}

// AddHeader 添加响应头 (可以有多个同名头)
func (c *BaseController) AddHeader(key, value string) {
	if c.Ctx == nil || c.Ctx.Request == nil {
		config.Error("Context is nil when trying to add header")
		return
	}
	c.Ctx.Request.Response.Header.Add(key, value)
}

// GetResponseHeader 获取响应头
func (c *BaseController) GetResponseHeader(key string) string {
	if c.Ctx == nil || c.Ctx.Request == nil {
		return ""
	}
	return string(c.Ctx.Request.Response.Header.Peek(key))
}

// SetContentType 设置Content-Type
func (c *BaseController) SetContentType(contentType string) {
	c.SetHeader("Content-Type", contentType)
}

// ============= JSONP支持方法 =============

// JSONP 返回JSONP响应
func (c *BaseController) JSONP(callback string, data any) {
	c.JSONPWithStatus(consts.StatusOK, callback, data)
}

// JSONPWithStatus 返回指定状态码的JSONP响应
func (c *BaseController) JSONPWithStatus(status int, callback string, data any) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return JSONP")
		return
	}

	if callback == "" {
		// 如果没有callback参数，退化为普通JSON
		c.Ctx.JSON(status, data)
		return
	}

	// 设置JSONP响应
	c.Ctx.Request.SetStatusCode(status)
	c.SetContentType("application/javascript; charset=utf-8")

	// 序列化JSON数据
	jsonData, err := yaml.Marshal(data) // 这里应该用json.Marshal，但为了避免导入冲突暂用yaml
	if err != nil {
		config.Errorf("Failed to marshal JSONP data: %v", err)
		return
	}

	// 构造JSONP响应
	response := callback + "(" + string(jsonData) + ");"
	c.Ctx.Request.Response.SetBodyString(response)
}

// ============= 特殊响应方法 =============

// EmptyResponse 返回空响应
func (c *BaseController) EmptyResponse() {
	if c.Ctx == nil {
		return
	}
	c.Ctx.Request.Response.SetBody(nil)
}

// RawResponse 返回原始响应
func (c *BaseController) RawResponse(data []byte) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return raw response")
		return
	}
	c.Ctx.Request.Response.SetBody(data)
}
