package mvc

import (
	"testing"

	"github.com/zsy619/yyhertz/framework/util"
)

// TestGlobalAddFuncMapIntegration 测试全局静态方法AddFuncMap的集成功能
func TestGlobalAddFuncMapIntegration(t *testing.T) {
	// 使用全局静态方法添加函数
	AddFuncMap("globalTest", func(s string) string {
		return "global_" + s
	})
	
	AddFuncMap("containString", util.ContainString)

	// 验证函数是否添加到全局HertzApp中
	globalFuncs := GetGlobalFuncMap()
	if _, exists := globalFuncs["globalTest"]; !exists {
		t.Error("globalTest function was not added via global AddFuncMap")
	}

	if _, exists := globalFuncs["containString"]; !exists {
		t.Error("containString function was not added via global AddFuncMap")
	}

	// 验证函数列表
	funcNames := ListFuncMap()
	foundGlobalTest := false
	foundContainString := false
	
	for _, name := range funcNames {
		if name == "globalTest" {
			foundGlobalTest = true
		}
		if name == "containString" {
			foundContainString = true
		}
	}

	if !foundGlobalTest {
		t.Error("globalTest function not found in global function list")
	}
	if !foundContainString {
		t.Error("containString function not found in global function list")
	}

	// 测试移除函数
	RemoveFuncMap("globalTest")
	funcNamesAfterRemove := ListFuncMap()
	foundAfterRemove := false
	for _, name := range funcNamesAfterRemove {
		if name == "globalTest" {
			foundAfterRemove = true
		}
	}

	if foundAfterRemove {
		t.Error("globalTest function should have been removed from global function list")
	}
}

// TestGlobalFuncMapWithNilApp 测试当HertzApp为nil时的行为
func TestGlobalFuncMapWithNilApp(t *testing.T) {
	// 备份原始的HertzApp
	originalApp := HertzApp
	defer func() {
		HertzApp = originalApp
	}()

	// 设置HertzApp为nil
	HertzApp = nil

	// 调用全局方法不应该panic
	AddFuncMap("testNil", func() string { return "test" })
	RemoveFuncMap("testNil")
	
	funcMap := GetGlobalFuncMap()
	if funcMap == nil {
		t.Error("GetGlobalFuncMap should return non-nil map even when HertzApp is nil")
	}

	funcList := ListFuncMap()
	if funcList == nil {
		t.Error("ListFuncMap should return non-nil slice even when HertzApp is nil")
	}
}

// TestAddFuncMapUsageExample 测试AddFuncMap的实际使用示例
func TestAddFuncMapUsageExample(t *testing.T) {
	// 示例：添加用户自定义的模板函数
	AddFuncMap("containString", util.ContainString)
	AddFuncMap("upper", func(s string) string {
		return "UPPER_" + s
	})
	AddFuncMap("formatPrice", func(price float64) string {
		return "$" + util.FmtFloat2(price)
	})

	// 验证函数都已正确添加
	funcList := ListFuncMap()
	expectedFuncs := map[string]bool{
		"containString": false,
		"upper":         false,
		"formatPrice":   false,
	}

	for _, name := range funcList {
		if _, exists := expectedFuncs[name]; exists {
			expectedFuncs[name] = true
		}
	}

	for funcName, found := range expectedFuncs {
		if !found {
			t.Errorf("Function %s was not found in the global function list", funcName)
		}
	}

	// 验证可以获取到正确的全局函数映射
	globalFuncs := GetGlobalFuncMap()
	for funcName := range expectedFuncs {
		if _, exists := globalFuncs[funcName]; !exists {
			t.Errorf("Function %s was not found in the global function map", funcName)
		}
	}

	// 清理测试函数
	RemoveFuncMap("containString")
	RemoveFuncMap("upper")
	RemoveFuncMap("formatPrice")
}