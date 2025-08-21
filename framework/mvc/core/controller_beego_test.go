package core

import (
	"testing"

	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/define"
)

// TestController 测试控制器
type TestController struct {
	*BaseController
	ExecutedSteps []string // 记录执行步骤
}

func (t *TestController) TestServeJSON() {
	t.ExecutedSteps = append(t.ExecutedSteps, "Before ServeJSON")
	t.Data["json"] = map[string]any{"message": "test"}
	t.ServeJSON()
	t.ExecutedSteps = append(t.ExecutedSteps, "After ServeJSON") // 这行不应该被执行
}

func (t *TestController) TestStopRun() {
	t.ExecutedSteps = append(t.ExecutedSteps, "Before StopRun")
	t.StopRun()
	t.ExecutedSteps = append(t.ExecutedSteps, "After StopRun") // 这行不应该被执行
}

func (t *TestController) Finish() {
	t.ExecutedSteps = append(t.ExecutedSteps, "Finish called") // 这应该不会被执行
}

// Test_BeegoStyleStopRun_ServeJSON 测试ServeJSON自动调用StopRun
func Test_BeegoStyleStopRun_ServeJSON(t *testing.T) {
	// 创建测试控制器
	testCtrl := &TestController{
		BaseController: NewBaseController(),
		ExecutedSteps:  []string{},
	}

	// 模拟初始化上下文
	c := &define.RequestContext{}
	enhancedCtx := contextenhanced.NewContext(c)
	testCtrl.Init(enhancedCtx, "TestController", "TestServeJSON", nil)

	// 测试ServeJSON是否能正确停止执行
	defer func() {
		if r := recover(); r != nil {
			if r == ErrAbort {
				// 这是预期的行为
				t.Log("✓ ServeJSON 正确触发了 ErrAbort panic")
			} else {
				t.Errorf("意外的panic: %v", r)
			}
		} else {
			t.Error("ServeJSON 应该触发 ErrAbort panic")
		}
	}()

	// 调用方法
	testCtrl.TestServeJSON()

	// 检查执行步骤
	expectedSteps := []string{"Before ServeJSON"}
	if len(testCtrl.ExecutedSteps) != len(expectedSteps) {
		t.Errorf("执行步骤数量不匹配，期望 %d，实际 %d", len(expectedSteps), len(testCtrl.ExecutedSteps))
	}

	for i, step := range expectedSteps {
		if i >= len(testCtrl.ExecutedSteps) || testCtrl.ExecutedSteps[i] != step {
			t.Errorf("步骤 %d 不匹配，期望 '%s'，实际 '%s'", i, step, testCtrl.ExecutedSteps[i])
		}
	}

	// 如果到这里，说明没有panic，这是不对的
	t.Error("ServeJSON 应该触发panic停止执行")
}

// Test_BeegoStyleStopRun_Manual 测试手动调用StopRun
func Test_BeegoStyleStopRun_Manual(t *testing.T) {
	// 创建测试控制器
	testCtrl := &TestController{
		BaseController: NewBaseController(),
		ExecutedSteps:  []string{},
	}

	// 模拟初始化上下文
	c := &define.RequestContext{}
	enhancedCtx := contextenhanced.NewContext(c)
	testCtrl.Init(enhancedCtx, "TestController", "TestStopRun", nil)

	// 测试手动调用StopRun是否能正确停止执行
	defer func() {
		if r := recover(); r != nil {
			if r == ErrAbort {
				// 这是预期的行为
				t.Log("✓ 手动调用 StopRun 正确触发了 ErrAbort panic")
			} else {
				t.Errorf("意外的panic: %v", r)
			}
		} else {
			t.Error("手动调用 StopRun 应该触发 ErrAbort panic")
		}
	}()

	// 调用方法
	testCtrl.TestStopRun()

	// 检查执行步骤
	expectedSteps := []string{"Before StopRun"}
	if len(testCtrl.ExecutedSteps) != len(expectedSteps) {
		t.Errorf("执行步骤数量不匹配，期望 %d，实际 %d", len(expectedSteps), len(testCtrl.ExecutedSteps))
	}

	// 如果到这里，说明没有panic，这是不对的
	t.Error("手动调用 StopRun 应该触发panic停止执行")
}

// Test_BeegoStyleStopRun_ErrAbortType 测试ErrAbort类型
func Test_BeegoStyleStopRun_ErrAbortType(t *testing.T) {
	if ErrAbort == nil {
		t.Error("ErrAbort 不应该为 nil")
	}

	if ErrAbort.Error() != "user stop run" {
		t.Errorf("ErrAbort 错误消息不匹配，期望 'user stop run'，实际 '%s'", ErrAbort.Error())
	}

	t.Log("✓ ErrAbort 类型检查通过")
}