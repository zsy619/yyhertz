package core

import (
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/response"
)

// ============= JSON响应方法 =============

// JSON 返回JSON格式的数据
func (c *BaseController) JSON(data any) {
	c.JSONWithStatus(consts.StatusOK, data)
}

// JSONWithStatus 返回指定状态码的JSON数据
func (c *BaseController) JSONWithStatus(status int, data any) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return JSON")
		return
	}
	c.Ctx.JSON(status, data)
}

// JSONSuccess 返回成功的JSON响应
func (c *BaseController) JSONSuccess(message string, data any) {
	c.JSON(response.Success(message, data))
}

// JSONError 返回错误的JSON响应
func (c *BaseController) JSONError(message string) {
	c.JSON(response.Error(message))
}

// JSONPage 返回分页JSON响应
func (c *BaseController) JSONPage(message string, data any, count int64) {
	c.JSON(response.SuccessPage(message, data, count))
}

// JSONOK 返回成功响应（200）
func (c *BaseController) JSONOK(message string, data any) {
	c.JSONStatus(200, 0, message, data)
}

// JSONStatus 返回指定状态码的JSON响应
func (c *BaseController) JSONStatus(status int, code int, message string, data any) {
	response := map[string]any{
		"code":    code,
		"message": message,
		"data":    data,
	}
	c.JSONWithStatus(status, response)
}

// ============= 字符串响应方法 =============

// String 返回字符串响应
func (c *BaseController) String(s string) {
	c.StringWithStatus(consts.StatusOK, s)
}

// StringWithStatus 返回指定状态码的字符串响应
func (c *BaseController) StringWithStatus(status int, s string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return string")
		return
	}
	c.Ctx.String(status, "%s", s)
}

// ============= 重定向和错误方法 =============

// Redirect 重定向
func (c *BaseController) Redirect(url string, code ...int) {
	statusCode := consts.StatusFound
	if len(code) > 0 {
		statusCode = code[0]
	}
	c.Ctx.RequestContext.Redirect(statusCode, []byte(url))
}

// Error 返回错误响应
func (c *BaseController) Error(code int, msg string) {
	c.Ctx.String(code, "%s", msg)
}

// ============= 响应头和原始数据方法 =============

// SetHeader 设置响应头
func (c *BaseController) SetHeader(key, value string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to set header")
		return
	}
	c.Ctx.SetHeader(key, value)
}

// Write 写入原始字节数据
func (c *BaseController) Write(data []byte) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to write data")
		return
	}
	c.Ctx.Write(data)
}