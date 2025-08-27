package main

import (
	"html/template"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestTemplateFunctionManager 测试统一的模板函数管理器
func TestTemplateFunctionManager(t *testing.T) {
	// 测试获取全局函数管理器
	manager := view.GetGlobalFunctionManager()
	if manager == nil {
		t.Fatal("GetGlobalFunctionManager() returned nil")
	}

	// 测试内置函数是否正确加载
	builtinCount := view.GetBuiltinFunctionCount()
	if builtinCount == 0 {
		t.Error("No builtin functions loaded")
	}
	t.Logf("Loaded %d builtin functions", builtinCount)

	// 测试添加全局函数
	testFuncName := "testGlobalFunc"
	testFunc := func() string { return "test" }

	manager.AddGlobalFunction(testFuncName, testFunc)

	if !manager.HasFunction(testFuncName) {
		t.Errorf("Global function '%s' not found after adding", testFuncName)
	}

	// 测试函数来源
	source := manager.GetFunctionSource(testFuncName)
	if source != view.FunctionSourceGlobal {
		t.Errorf("Expected source to be global, got %s", source)
	}

	// 测试内置函数来源
	if view.HasBuiltinFunction("add") {
		source = manager.GetFunctionSource("add")
		if source != view.FunctionSourceBuiltin {
			t.Errorf("Expected 'add' to be builtin, got %s", source)
		}
	}

	// 测试获取合并函数
	engineFuncs := template.FuncMap{
		"engineFunc": func() string { return "engine" },
	}

	merged := manager.GetMergedFunctions(engineFuncs)
	if len(merged) == 0 {
		t.Error("Merged functions is empty")
	}

	// 验证所有函数都存在
	if _, exists := merged[testFuncName]; !exists {
		t.Error("Global function missing in merged functions")
	}

	if _, exists := merged["engineFunc"]; !exists {
		t.Error("Engine function missing in merged functions")
	}

	// 测试函数列表
	funcList := manager.ListFunctions()
	if len(funcList["global"]) == 0 {
		t.Error("No global functions in list")
	}

	if len(funcList["builtin"]) == 0 {
		t.Error("No builtin functions in list")
	}

	t.Logf("Global functions: %v", funcList["global"])
	t.Logf("Builtin function count: %d", len(funcList["builtin"]))
}

// TestBackwardCompatibility 测试向后兼容性
func TestBackwardCompatibility(t *testing.T) {
	// 测试旧的 AddGlobalFunction 接口
	oldTestFunc := func() string { return "old" }
	view.AddGlobalFunction("oldTestFunc", oldTestFunc)

	// 验证通过新接口可以访问
	manager := view.GetGlobalFunctionManager()
	if !manager.HasFunction("oldTestFunc") {
		t.Error("Function added via old interface not accessible via new interface")
	}

	// 测试 TemplateFuncs 变量仍然可用
	if len(view.TemplateFuncs) == 0 {
		t.Error("TemplateFuncs variable is empty")
	}

	// 测试内置函数仍然可以通过 TemplateFuncs 访问
	if _, exists := view.TemplateFuncs["add"]; !exists {
		t.Error("Builtin function 'add' not accessible via TemplateFuncs")
	}
}

// TestTemplateEngine 测试模板引擎的函数管理
func TestTemplateEngine(t *testing.T) {
	// 创建模板引擎
	engine, err := view.NewTemplateEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// 测试引擎函数数量
	count := engine.GetEngineFunctionCount()
	if count == 0 {
		t.Error("Engine has no functions")
	}
	t.Logf("Engine has %d functions", count)

	// 测试添加引擎函数
	engine.AddFunction("engineTestFunc", func() string { return "engine" })

	if !engine.HasEngineFunction("engineTestFunc") {
		t.Error("Engine function not found after adding")
	}

	// 验证函数数量增加
	newCount := engine.GetEngineFunctionCount()
	if newCount <= count {
		t.Error("Engine function count did not increase")
	}

	// 测试获取函数名称
	names := engine.GetFunctionNames()
	if len(names) == 0 {
		t.Error("Engine function names is empty")
	}

	foundEngineFunc := false
	for _, name := range names {
		if name == "engineTestFunc" {
			foundEngineFunc = true
			break
		}
	}
	if !foundEngineFunc {
		t.Error("engineTestFunc not found in function names")
	}
}

// TestFunctionPriority 测试函数优先级
func TestFunctionPriority(t *testing.T) {
	manager := view.GetGlobalFunctionManager()

	// 添加一个与内置函数同名的全局函数
	testFuncName := "add" // 与内置函数同名
	globalAdd := func() string { return "global_add" }

	manager.AddGlobalFunction(testFuncName, globalAdd)

	// 创建引擎函数映射
	engineFuncs := template.FuncMap{
		"add": func() string { return "engine_add" },
	}

	// 获取合并函数
	merged := manager.GetMergedFunctions(engineFuncs)

	// 验证引擎函数优先级最高
	if fn, exists := merged["add"]; exists {
		if result := fn.(func() string)(); result != "engine_add" {
			t.Errorf("Expected engine function to have priority, got %s", result)
		}
	} else {
		t.Error("add function not found in merged functions")
	}

	// 测试没有引擎函数时，全局函数优先级高于内置函数
	noEngineFuncs := template.FuncMap{}
	merged2 := manager.GetMergedFunctions(noEngineFuncs)

	if fn, exists := merged2["add"]; exists {
		if result := fn.(func() string)(); result != "global_add" {
			t.Errorf("Expected global function to override builtin, got %s", result)
		}
	}

	// 清理
	manager.RemoveGlobalFunction(testFuncName)
}
