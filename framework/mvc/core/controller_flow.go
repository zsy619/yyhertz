package core

import (
	"strconv"
)

// ============= 流程控制方法 =============

// StopRun 停止后续处理（Beego兼容）
func (c *BaseController) StopRun() {
	// 在增强Context中，我们可以使用Abort方法
	if c.Ctx != nil {
		c.Ctx.Abort()
	}
}

// Abort 终止请求处理（Beego兼容）
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

	c.Ctx.AbortWithStatus(statusCode)
}

// CustomAbort 自定义中止（Beego兼容）
func (c *BaseController) CustomAbort(status int, body string) {
	if c.Ctx == nil {
		return
	}
	c.Ctx.String(status, "%s", body)
	c.Ctx.Abort()
}