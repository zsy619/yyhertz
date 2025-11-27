package core

import (
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
)

// IController 控制器接口定义（完全兼容Beego ControllerInterface）
//
// IController 定义了所有控制器必须实现的方法，完全兼容 Beego 的 ControllerInterface 规范。
// 这个接口确保了控制器的标准化和一致性，同时支持现代化的流程控制。
//
// 接口方法说明：
//
// 生命周期方法：
//   - Init(): 控制器初始化，设置基本属性
//   - Prepare(): 预处理，在具体方法执行前调用
//   - Finish(): 后处理，在具体方法执行后调用
//
// 控制器信息：
//   - GetControllerName(): 获取控制器名称
//   - GetActionName(): 获取当前动作名称
//
// 模板渲染：
//   - Render(): 渲染模板并输出
//
// 安全功能：
//   - XSRFToken(): 生成或获取XSRF令牌
//   - CheckXSRFCookie(): 验证XSRF令牌
//
// 路由映射：
//   - URLMapping(): 设置URL到方法的映射关系
//   - HandlerFunc(): 检查指定方法是否可处理请求
//
// 流程控制：
//   - ShouldStopExecution(): 检查是否应该停止后续执行
//
// 实现说明：
// BaseController 是此接口的标准实现，提供了完整的功能支持。
// 自定义控制器可以直接继承 BaseController 或实现此接口。
//
// 使用示例：
//
//	type MyController struct {
//		mvc.BaseController
//	}
//
//	func (c *MyController) Get() {
//		c.Data["json"] = map[string]any{"message": "Hello"}
//		c.ServeJSON() // 自动调用 StopRun()
//	}
//
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

	// 流程控制方法（执行状态检查）
	ShouldStopExecution() bool
}