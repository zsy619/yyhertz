package errors

import (
	"testing"
)

func TestDefaultErrorControllerCreation(t *testing.T) {
	// 测试默认控制器创建
	controller := NewDefaultErrorController()
	if controller == nil {
		t.Fatal("默认控制器创建失败")
	}

	// 验证默认配置
	if controller.Language != DefaultLanguage {
		t.Errorf("期望语言: %s, 实际: %s", DefaultLanguage, controller.Language)
	}

	if controller.CustomTitle != DefaultTitle {
		t.Errorf("期望标题: %s, 实际: %s", DefaultTitle, controller.CustomTitle)
	}

	if !controller.ShowDetailedError {
		t.Error("默认应该显示详细错误")
	}

	if !controller.EnableDebugInfo {
		t.Error("默认应该启用调试信息")
	}
}

func TestProductionErrorController(t *testing.T) {
	// 测试生产环境控制器
	controller := NewProductionErrorController()
	if controller == nil {
		t.Fatal("生产环境控制器创建失败")
	}

	// 验证生产环境配置
	if controller.ShowDetailedError {
		t.Error("生产环境不应该显示详细错误")
	}

	if controller.EnableDebugInfo {
		t.Error("生产环境不应该启用调试信息")
	}
}

func TestDevelopmentErrorController(t *testing.T) {
	// 测试开发环境控制器
	controller := NewDevelopmentErrorController()
	if controller == nil {
		t.Fatal("开发环境控制器创建失败")
	}

	// 验证开发环境配置
	if !controller.ShowDetailedError {
		t.Error("开发环境应该显示详细错误")
	}

	if !controller.EnableDebugInfo {
		t.Error("开发环境应该启用调试信息")
	}
}

func TestErrorControllerInterface(t *testing.T) {
	controller := NewDefaultErrorController()
	
	// 测试优先级
	priority := controller.Priority()
	if priority != DefaultPriority {
		t.Errorf("期望优先级: %d, 实际: %d", DefaultPriority, priority)
	}

	// 测试CanHandle
	if !controller.CanHandle(404, nil) {
		t.Error("控制器应该能够处理所有错误")
	}

	if !controller.CanHandle(500, nil) {
		t.Error("控制器应该能够处理所有错误")
	}
}

func TestConstants(t *testing.T) {
	// 验证常量定义
	if DefaultLanguage != "zh-CN" {
		t.Errorf("默认语言常量错误: %s", DefaultLanguage)
	}

	if DefaultTitle != "YYHertz Framework" {
		t.Errorf("默认标题常量错误: %s", DefaultTitle)
	}

	if DefaultPriority != 1000 {
		t.Errorf("默认优先级常量错误: %d", DefaultPriority)
	}

	if MaxLastErrors != 50 {
		t.Errorf("最大错误记录常量错误: %d", MaxLastErrors)
	}

	if HookPhaseBefore != "before" {
		t.Errorf("前置钩子阶段常量错误: %s", HookPhaseBefore)
	}

	if HookPhaseAfter != "after" {
		t.Errorf("后置钩子阶段常量错误: %s", HookPhaseAfter)
	}
}