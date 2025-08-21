package core

import (
	"html/template"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestAppTemplateIntegration 测试 App 与 view 引擎的模板函数集成
func TestAppTemplateIntegration(t *testing.T) {
	// 创建 App 实例
	app := NewApp()
	if app == nil {
		t.Fatal("Failed to create App instance")
	}

	// 测试添加函数
	testFuncName := "appTestFunc"
	testFunc := func() string { return "app_test" }
	
	app.AddFuncMap(testFuncName, testFunc)
	
	// 验证函数是否添加成功
	if !app.HasFuncMap(testFuncName) {
		t.Errorf("Function '%s' not found after adding", testFuncName)
	}
	
	// 验证函数也存在于 view 引擎中
	if !view.HasTemplateFunction(testFuncName) {
		t.Error("Function not found in view engine")
	}
	
	// 测试获取函数来源
	source := app.GetFuncMapSource(testFuncName)
	if source != "global" {
		t.Errorf("Expected source to be 'global', got '%s'", source)
	}
	
	// 测试获取全局函数映射
	globalFuncs := app.GetGlobalFuncMap()
	if _, exists := globalFuncs[testFuncName]; !exists {
		t.Error("Function not found in global func map")
	}
	
	// 测试函数列表
	funcNames := app.ListFuncMap()
	found := false
	for _, name := range funcNames {
		if name == testFuncName {
			found = true
			break
		}
	}
	if !found {
		t.Error("Function not found in function list")
	}
	
	// 测试移除函数
	app.RemoveFuncMap(testFuncName)
	if app.HasFuncMap(testFuncName) {
		t.Error("Function still exists after removal")
	}
	
	// 验证函数也从 view 引擎中移除
	if view.HasTemplateFunction(testFuncName) {
		t.Error("Function still exists in view engine after removal")
	}
}

// TestAppFuncMapGroup 测试批量函数注册
func TestAppFuncMapGroup(t *testing.T) {
	app := NewApp()
	
	// 准备函数组
	testGroup := template.FuncMap{
		"groupFunc1": func() string { return "group1" },
		"groupFunc2": func() string { return "group2" },
		"groupFunc3": func() string { return "group3" },
	}
	
	// 记录添加前的函数数量
	countBefore := app.GetFuncMapCount()
	
	// 添加函数组
	app.AddFuncMapGroup("testGroup", testGroup)
	
	// 验证函数数量增加
	countAfter := app.GetFuncMapCount()
	if countAfter != countBefore+3 {
		t.Errorf("Expected function count to increase by 3, got %d -> %d", countBefore, countAfter)
	}
	
	// 验证每个函数都存在
	for name := range testGroup {
		if !app.HasFuncMap(name) {
			t.Errorf("Group function '%s' not found", name)
		}
	}
	
	// 测试按来源分类的函数列表
	funcsBySource := app.ListFuncMapBySource()
	globalFuncs := funcsBySource["global"]
	
	// 验证组函数都在全局函数列表中
	for name := range testGroup {
		found := false
		for _, globalName := range globalFuncs {
			if globalName == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Group function '%s' not found in global functions", name)
		}
	}
	
	// 清理 - 移除测试函数
	for name := range testGroup {
		app.RemoveFuncMap(name)
	}
}

// TestAppTemplateBackwardCompatibility 测试向后兼容性
func TestAppTemplateBackwardCompatibility(t *testing.T) {
	app := NewApp()
	
	// 测试旧的方法接口是否正常工作
	testFunc := func() string { return "backward_test" }
	app.AddFuncMap("backwardTestFunc", testFunc)
	
	// 使用新方法验证
	if !app.HasFuncMap("backwardTestFunc") {
		t.Error("Backward compatibility test failed")
	}
	
	// 测试 GetGlobalFuncMap 返回正确的数据
	globalMap := app.GetGlobalFuncMap()
	if _, exists := globalMap["backwardTestFunc"]; !exists {
		t.Error("Function not found via GetGlobalFuncMap")
	}
	
	// 测试 ListFuncMap 包含新函数
	allFuncs := app.ListFuncMap()
	found := false
	for _, name := range allFuncs {
		if name == "backwardTestFunc" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Function not found via ListFuncMap")
	}
	
	// 清理
	app.RemoveFuncMap("backwardTestFunc")
}

// TestAppTemplateIntegrationWithViewEngine 测试与 view 引擎的完全集成
func TestAppTemplateIntegrationWithViewEngine(t *testing.T) {
	app := NewApp()
	
	// 在 App 层添加函数
	app.AddFuncMap("appLayerFunc", func() string { return "app_layer" })
	
	// 在 view 层直接添加函数
	view.AddGlobalFunction("viewLayerFunc", func() string { return "view_layer" })
	
	// 验证 App 层可以看到 view 层的函数
	if !app.HasFuncMap("viewLayerFunc") {
		t.Error("App cannot see view layer function")
	}
	
	// 验证两个函数都存在于统一管理器中
	manager := view.GetTemplateFunctionManager()
	if !manager.HasFunction("appLayerFunc") {
		t.Error("App layer function not found in manager")
	}
	if !manager.HasFunction("viewLayerFunc") {
		t.Error("View layer function not found in manager")
	}
	
	// 验证函数数量一致性
	appCount := app.GetFuncMapCount()
	managerFuncs := manager.ListFunctions()
	managerCount := len(managerFuncs["builtin"]) + len(managerFuncs["global"])
	
	if appCount != managerCount {
		t.Errorf("Function count mismatch: App=%d, Manager=%d", appCount, managerCount)
	}
	
	// 清理
	app.RemoveFuncMap("appLayerFunc")
	view.RemoveGlobalFunction("viewLayerFunc")
}

// TestAppEnhancedMethods 测试新增的增强方法
func TestAppEnhancedMethods(t *testing.T) {
	app := NewApp()
	
	// 测试函数数量统计
	initialCount := app.GetFuncMapCount()
	if initialCount == 0 {
		t.Error("No builtin functions found")
	}
	t.Logf("Initial function count: %d", initialCount)
	
	// 测试按来源分类列表
	funcsBySource := app.ListFuncMapBySource()
	if len(funcsBySource["builtin"]) == 0 {
		t.Error("No builtin functions in categorized list")
	}
	t.Logf("Builtin functions: %d", len(funcsBySource["builtin"]))
	t.Logf("Global functions: %d", len(funcsBySource["global"]))
	
	// 测试函数来源查询
	if len(funcsBySource["builtin"]) > 0 {
		firstBuiltin := funcsBySource["builtin"][0]
		source := app.GetFuncMapSource(firstBuiltin)
		if source != "builtin" {
			t.Errorf("Expected '%s' to be builtin, got '%s'", firstBuiltin, source)
		}
	}
	
	// 测试添加函数后的数量变化
	app.AddFuncMap("enhancedTestFunc", func() string { return "enhanced" })
	newCount := app.GetFuncMapCount()
	if newCount != initialCount+1 {
		t.Errorf("Function count should be %d, got %d", initialCount+1, newCount)
	}
	
	// 清理
	app.RemoveFuncMap("enhancedTestFunc")
}