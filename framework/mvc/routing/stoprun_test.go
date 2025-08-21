package routing

import (
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// Test_StopRun_Mechanism 测试StopRun机制
func Test_StopRun_Mechanism(t *testing.T) {
	// 创建测试控制器
	controller := core.NewBaseController()
	
	// 初始状态应该是false
	if controller.ShouldStopExecution() {
		t.Error("Initial ShouldStopExecution should be false")
	}
	
	// 调用StopRun
	controller.StopRun()
	
	// 现在应该是true
	if !controller.ShouldStopExecution() {
		t.Error("ShouldStopExecution should be true after calling StopRun")
	}
	
	t.Log("✓ StopRun mechanism works correctly")
}

// Test_ResetExecutionState 测试重置执行状态
func Test_ResetExecutionState(t *testing.T) {
	controller := core.NewBaseController()
	
	// 设置停止状态
	controller.StopRun()
	if !controller.ShouldStopExecution() {
		t.Error("ShouldStopExecution should be true after StopRun")
	}
	
	// 重置状态
	controller.ResetExecutionState()
	if controller.ShouldStopExecution() {
		t.Error("ShouldStopExecution should be false after reset")
	}
	
	t.Log("✓ ResetExecutionState works correctly")
}