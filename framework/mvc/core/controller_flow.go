package core

import (
	"errors"
	"strconv"
)

// ErrAbort 用户停止运行错误（完全兼容Beego）
var ErrAbort = errors.New("user stop run")

// ============= 流程控制方法 =============

// StopRun 停止当前请求的执行，完全兼容beego风格
// 
// 该方法可以在任意位置调用，停止后续的控制器方法执行。
// 完全按照beego方式实现：使用panic(ErrAbort)来停止执行。
//
// 使用示例：
//
//	if !c.isAuthenticated() {
//		c.Data["json"] = map[string]any{"error": "Unauthorized"}
//		c.ServeJSON()
//		c.StopRun()  // 停止后续执行
//	}
//
func (c *BaseController) StopRun() {
	panic(ErrAbort)
}

// Abort 中止请求并设置HTTP状态码，兼容beego风格
//
// 该方法中止请求执行，并设置指定的HTTP状态码。
// 为了兼容beego，同时支持字符串和整数参数。
//
// 使用示例：
//
//	c.Abort("404")  // 字符串参数
//	c.Abort(404)    // 整数参数(重载方法)
//
func (c *BaseController) Abort(code string) {
	if c.Ctx == nil {
		return
	}

	statusCode := 500
	switch code {
	case "404":
		statusCode = 404
	case "403":
		statusCode = 403
	case "401":
		statusCode = 401
	default:
		if c, err := strconv.Atoi(code); err == nil {
			statusCode = c
		}
	}

	c.Ctx.Status(statusCode)
	c.StopRun()
}

// AbortWithStatus 中止请求并设置HTTP状态码（整数参数版本）
func (c *BaseController) AbortWithStatus(status int) {
	if c.Ctx != nil {
		c.Ctx.Status(status)
	}
	c.StopRun()
}

// CustomAbort 自定义中止请求，兼容beego风格
//
// 该方法中止请求执行，设置指定的HTTP状态码和响应内容。
//
// 使用示例：
//
//	c.CustomAbort(403, "Access denied")
//
func (c *BaseController) CustomAbort(status int, body string) {
	if c.Ctx != nil {
		c.Ctx.Status(status)
		c.Ctx.WriteString(body)
	}
	c.StopRun()
}