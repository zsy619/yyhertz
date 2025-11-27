package core

import (
	"testing"
)

// Test_BaseController_ImplementsIController 测试BaseController完全实现IController接口
func Test_BaseController_ImplementsIController(t *testing.T) {
	// 创建BaseController实例
	controller := NewBaseController()
	
	// 确保可以转换为IController接口
	var iController IController = controller
	
	// 测试所有接口方法是否可调用（不测试具体功能，只测试接口完整性）
	
	// 生命周期方法
	controller.Prepare()
	controller.Finish()
	
	// 控制器信息方法
	controllerName := iController.GetControllerName()
	_ = controllerName // 使用变量避免编译警告
	
	actionName := iController.GetActionName()
	_ = actionName
	
	// 模板渲染方法（需要上下文，这里只测试能调用）
	err := iController.Render()
	_ = err // 可能会出错，这里只关心接口可调用
	
	// 安全方法
	xsrfToken := iController.XSRFToken()
	_ = xsrfToken
	
	checkResult := iController.CheckXSRFCookie()
	_ = checkResult
	
	// 路由映射方法
	iController.URLMapping()
	
	handlerResult := iController.HandlerFunc("TestMethod")
	_ = handlerResult
	
	// 流程控制方法
	shouldStop := iController.ShouldStopExecution()
	if shouldStop {
		t.Error("初始状态下ShouldStopExecution应该返回false")
	}
	
	t.Log("✓ BaseController 完全实现了 IController 接口")
}

// Test_BaseController_InterfaceMethodsWork 测试接口方法的基本功能
func Test_BaseController_InterfaceMethodsWork(t *testing.T) {
	controller := NewBaseController()
	
	// 测试StopRun机制
	if controller.ShouldStopExecution() {
		t.Error("初始状态应该是false")
	}
	
	controller.StopRun()
	
	if !controller.ShouldStopExecution() {
		t.Error("调用StopRun后应该是true")
	}
	
	// 重置状态
	controller.ResetExecutionState()
	
	if controller.ShouldStopExecution() {
		t.Error("重置后应该是false")
	}
	
	// 测试控制器名称获取
	controller.ControllerName = "TestController"
	controller.ActionName = "TestAction"
	
	if controller.GetControllerName() == "" {
		t.Error("应该能获取控制器名称")
	}
	
	actionName := controller.GetActionName()
	if actionName == "" {
		t.Error("应该能获取动作名称")
	}
	
	// 测试URL映射
	controller.URLMapping() // 不应该panic
	
	// 测试处理器函数检查
	result := controller.HandlerFunc("NonExistentMethod")
	if result {
		t.Error("不存在的方法应该返回false")
	}
	
	t.Log("✓ BaseController 接口方法基本功能正常")
}

// Test_BaseController_TypeSafety 测试类型安全
func Test_BaseController_TypeSafety(t *testing.T) {
	controller := NewBaseController()
	
	// 确保可以作为IController使用
	var iController IController = controller
	
	// 确保可以调用BaseController特有的方法
	controller.SetContextData("test", "value")
	value := controller.GetContextData("test")
	_ = value
	
	// 确保接口方法调用不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("接口方法调用不应该panic: %v", r)
		}
	}()
	
	iController.GetControllerName()
	iController.GetActionName()
	iController.XSRFToken()
	iController.CheckXSRFCookie()
	iController.HandlerFunc("test")
	iController.URLMapping()
	iController.ShouldStopExecution()
	
	t.Log("✓ BaseController 类型安全检查通过")
}