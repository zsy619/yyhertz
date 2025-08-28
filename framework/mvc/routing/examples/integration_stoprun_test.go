package main

import (
	"reflect"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// MockController 模拟控制器
type MockController struct {
	*core.BaseController
	finishCalled bool
}

func (mc *MockController) TestAction() {
	// 设置JSON数据并调用StopRun
	mc.Data["json"] = map[string]any{"message": "test"}
	mc.StopRun() // 模拟ServeJSON调用StopRun
}

func (mc *MockController) Finish() {
	mc.finishCalled = true
	mc.BaseController.Finish()
}

// Test_RouteInfo_StopExecution_Integration 测试路由信息与停止执行的集成
func Test_RouteInfo_StopExecution_Integration(t *testing.T) {
	// 创建模拟控制器
	controller := &MockController{
		BaseController: core.NewBaseController(),
	}

	// 模拟调用控制器方法
	method := reflect.ValueOf(controller).MethodByName("TestAction")
	if !method.IsValid() {
		t.Fatal("TestAction method not found")
	}

	// 调用方法 - 这会触发StopRun
	method.Call([]reflect.Value{})

	// 验证ShouldStopExecution状态
	if !controller.ShouldStopExecution() {
		t.Error("Expected ShouldStopExecution to be true after TestAction calls StopRun")
	}

	// 模拟路由处理器的检查逻辑
	if iController, ok := interface{}(controller).(core.IController); ok {
		if iController.ShouldStopExecution() {
			t.Log("✓ Route handler would correctly skip further processing")
			// 这里应该跳过Finish调用
			return
		}
	}

	// 如果到这里，说明检查失败
	controller.Finish()
	t.Error("Route handler should have stopped execution but didn't")
}

// Test_Multiple_StopRun_Calls 测试多次调用StopRun
func Test_Multiple_StopRun_Calls(t *testing.T) {
	controller := core.NewBaseController()

	// 初始状态
	if controller.ShouldStopExecution() {
		t.Error("Initial state should be false")
	}

	// 第一次调用
	controller.StopRun()
	if !controller.ShouldStopExecution() {
		t.Error("Should be true after first StopRun")
	}

	// 多次调用应该保持状态
	controller.StopRun()
	controller.StopRun()
	if !controller.ShouldStopExecution() {
		t.Error("Should remain true after multiple StopRun calls")
	}

	t.Log("✓ Multiple StopRun calls handled correctly")
}

// Test_StopExecution_Interface_Compliance 测试接口符合性
func Test_StopExecution_Interface_Compliance(t *testing.T) {
	controller := core.NewBaseController()

	// 确保实现了IController接口
	var iController core.IController = controller

	// 测试接口方法
	if iController.ShouldStopExecution() {
		t.Error("Initial state should be false")
	}

	// 调用StopRun（通过具体类型）
	controller.StopRun()

	// 通过接口检查状态
	if !iController.ShouldStopExecution() {
		t.Error("Interface method should return true after StopRun")
	}

	t.Log("✓ IController interface compliance verified")
}
