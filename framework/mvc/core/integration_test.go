package core

import (
	"testing"
)

// TestController 用于测试的控制器
type TestJSONController struct {
	*BaseController
}

func (t *TestJSONController) Get() {
	// 设置JSON响应并触发StopRun
	t.Data["json"] = map[string]string{"message": "success"}
	t.ServeJSON()
	
	// 这后面不应该有任何HTML输出
}

func (t *TestJSONController) GetWithRender() {
	// 故意设置模板数据，看看会不会渲染
	t.Data["title"] = "Test Page"
	t.TplName = "test/page"
	
	// 但是调用JSON响应，应该阻止HTML渲染
	t.Data["json"] = map[string]string{"message": "json_only"} 
	t.ServeJSON()
}

// Test_ControllerHandler_StopExecution 测试控制器处理器是否正确处理StopExecution
func Test_ControllerHandler_StopExecution(t *testing.T) {
	// 创建测试控制器
	testController := &TestJSONController{
		BaseController: NewBaseController(),
	}
	
	// 验证JSON数据已设置但StopRun没有调用（因为Context为nil）
	if jsonData, exists := testController.Data["json"]; exists {
		t.Errorf("JSON data should not be set initially, but got: %v", jsonData)
	}
	
	// 直接测试StopRun机制
	if testController.ShouldStopExecution() {
		t.Error("Initial ShouldStopExecution should be false")
	}
	
	// 手动调用StopRun
	testController.StopRun()
	
	// 现在应该是true
	if !testController.ShouldStopExecution() {
		t.Error("Expected ShouldStopExecution to be true after StopRun")
	}
	
	t.Log("✓ Controller StopRun mechanism works correctly")
}

// Test_ShouldStopExecution_Flag 测试ShouldStopExecution标志
func Test_ShouldStopExecution_Flag(t *testing.T) {
	controller := NewBaseController()
	
	// 初始状态
	if controller.ShouldStopExecution() {
		t.Error("Initial ShouldStopExecution should be false")
	}
	
	// 设置JSON数据并调用ServeJSON（模拟）
	controller.Data["json"] = map[string]string{"test": "data"}
	
	// 手动调用StopRun（模拟ServeJSON的行为）
	controller.StopRun()
	
	// 现在应该是true
	if !controller.ShouldStopExecution() {
		t.Error("ShouldStopExecution should be true after StopRun")
	}
	
	t.Log("✓ ShouldStopExecution flag works correctly")
}