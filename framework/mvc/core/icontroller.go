package core

import (
	contextenhanced "github.com/zsy619/yyhertz/framework/context"
)

// 控制器接口定义（完全兼容Beego ControllerInterface）
type IController interface {
	// 生命周期方法（符合Beego ControllerInterface规范）
	Init(ct *contextenhanced.Context, controllerName, actionName string, app any)
	Prepare()
	Finish()

	// Controller名称相关方法
	GetControllerName() string
	GetActionName() string

	// Beego兼容的渲染方法
	Render() error

	// XSRF/CSRF安全方法（Beego兼容）
	XSRFToken() string
	CheckXSRFCookie() bool

	// 控制器方法处理（Beego兼容）
	HandlerFunc(fn string) bool

	// URL映射注册（Beego兼容）
	URLMapping()
}