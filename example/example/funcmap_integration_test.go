package example

import (
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/pkg/xfmt"
	xstrings "github.com/zsy619/yyhertz/framework/pkg/xstring"
)

// TestGlobalAddFuncMapIntegration 测试全局静态方法mvc.AddFuncMap的集成功能
func TestGlobalAddFuncMapIntegration(t *testing.T) {
	// 使用全局静态方法添加函数
	mvc.AddFuncMap("globalTest", func(s string) string {
		return "global_" + s
	})

	mvc.AddFuncMap("containString", xstrings.ContainsCommaStr)

	// 验证函数是否添加到全局mvc.HertzApp中
	globalFuncs := mvc.GetGlobalFuncMap()
	if _, exists := globalFuncs["globalTest"]; !exists {
		t.Error("globalTest function was not added via global mvc.AddFuncMap")
	}

	if _, exists := globalFuncs["containString"]; !exists {
		t.Error("containString function was not added via global mvc.AddFuncMap")
	}

	// 验证函数列表
	funcNames := mvc.ListFuncMap()
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
	mvc.RemoveFuncMap("globalTest")
	funcNamesAfterRemove := mvc.ListFuncMap()
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

// TestGlobalFuncMapWithNilApp 测试当mvc.HertzApp为nil时的行为
func TestGlobalFuncMapWithNilApp(t *testing.T) {
	// 备份原始的mvc.HertzApp
	originalApp := mvc.HertzApp
	defer func() {
		mvc.HertzApp = originalApp
	}()

	// 设置mvc.HertzApp为nil
	mvc.HertzApp = nil

	// 调用全局方法不应该panic
	mvc.AddFuncMap("testNil", func() string { return "test" })
	mvc.RemoveFuncMap("testNil")

	funcMap := mvc.GetGlobalFuncMap()
	if funcMap == nil {
		t.Error("mvc.GetGlobalFuncMap should return non-nil map even when mvc.HertzApp is nil")
	}

	funcList := mvc.ListFuncMap()
	if funcList == nil {
		t.Error("mvc.ListFuncMap should return non-nil slice even when mvc.HertzApp is nil")
	}
}

// TestAddFuncMapUsageExample 测试mvc.AddFuncMap的实际使用示例
func TestAddFuncMapUsageExample(t *testing.T) {
	// 示例：添加用户自定义的模板函数
	mvc.AddFuncMap("containString", xstrings.ContainsCommaStr)
	mvc.AddFuncMap("upper", func(s string) string {
		return "UPPER_" + s
	})
	mvc.AddFuncMap("formatPrice", func(price float64) string {
		return "$" + xfmt.FmtFloat2(price)
	})

	// 验证函数都已正确添加
	funcList := mvc.ListFuncMap()
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
	globalFuncs := mvc.GetGlobalFuncMap()
	for funcName := range expectedFuncs {
		if _, exists := globalFuncs[funcName]; !exists {
			t.Errorf("Function %s was not found in the global function map", funcName)
		}
	}

	// 清理测试函数
	mvc.RemoveFuncMap("containString")
	mvc.RemoveFuncMap("upper")
	mvc.RemoveFuncMap("formatPrice")
}
