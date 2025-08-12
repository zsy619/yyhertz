package core

import (
	"html/template"
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/util"
)

// TestAppAddFuncMap 测试应用级别的AddFuncMap功能
func TestAppAddFuncMap(t *testing.T) {
	// 创建测试应用
	app := NewApp()

	// 测试添加自定义函数
	testFunc := func(s string) string {
		return strings.ToUpper(s)
	}

	// 添加函数
	app.AddFuncMap("toUpper", testFunc)
	app.AddFuncMap("containString", util.ContainString)

	// 验证函数是否成功添加
	globalFuncs := app.GetGlobalFuncMap()
	if _, exists := globalFuncs["toUpper"]; !exists {
		t.Error("toUpper function was not added to global funcMap")
	}

	if _, exists := globalFuncs["containString"]; !exists {
		t.Error("containString function was not added to global funcMap")
	}

	// 测试函数列表
	funcNames := app.ListFuncMap()
	foundUpper := false
	foundContain := false
	for _, name := range funcNames {
		if name == "toUpper" {
			foundUpper = true
		}
		if name == "containString" {
			foundContain = true
		}
	}

	if !foundUpper {
		t.Error("toUpper function not found in function list")
	}
	if !foundContain {
		t.Error("containString function not found in function list")
	}

	// 测试移除函数
	app.RemoveFuncMap("toUpper")
	globalFuncsAfterRemove := app.GetGlobalFuncMap()
	if _, exists := globalFuncsAfterRemove["toUpper"]; exists {
		t.Error("toUpper function should have been removed from global funcMap")
	}
}

// TestAddFuncMapWithTemplate 测试AddFuncMap与模板引擎的集成
func TestAddFuncMapWithTemplate(t *testing.T) {
	app := NewApp()

	// 添加测试函数
	app.AddFuncMap("reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})

	// 创建一个简单的模板来测试函数
	tmpl := template.New("test").Funcs(app.GetGlobalFuncMap())
	
	templateStr := `{{reverse "hello"}}`
	tmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	var result strings.Builder
	err = tmpl.Execute(&result, nil)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	expected := "olleh"
	if result.String() != expected {
		t.Errorf("Expected %s, got %s", expected, result.String())
	}
}

// TestAddFuncMapContainString 测试util.ContainString函数
func TestAddFuncMapContainString(t *testing.T) {
	app := NewApp()

	// 添加util.ContainString函数
	app.AddFuncMap("containString", util.ContainString)

	// 创建模板来测试containString函数
	tmpl := template.New("test").Funcs(app.GetGlobalFuncMap())
	
	templateStr := `{{if containString .Tags "important"}}Important{{else}}Normal{{end}}`
	tmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// 测试包含情况
	data1 := map[string]string{"Tags": "urgent,important,priority"}
	var result1 strings.Builder
	err = tmpl.Execute(&result1, data1)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if result1.String() != "Important" {
		t.Errorf("Expected 'Important', got '%s'", result1.String())
	}

	// 测试不包含情况
	data2 := map[string]string{"Tags": "normal,regular,standard"}
	var result2 strings.Builder
	err = tmpl.Execute(&result2, data2)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if result2.String() != "Normal" {
		t.Errorf("Expected 'Normal', got '%s'", result2.String())
	}
}

// TestAddFuncMapThreadSafety 测试AddFuncMap的线程安全性
func TestAddFuncMapThreadSafety(t *testing.T) {
	app := NewApp()

	// 并发添加和移除函数
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			funcName := "func" + string(rune('A'+i%26))
			app.AddFuncMap(funcName, func(s string) string { return s })
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			funcName := "func" + string(rune('A'+i%26))
			app.RemoveFuncMap(funcName)
		}
		done <- true
	}()

	// 等待两个协程完成
	<-done
	<-done

	// 验证操作完成后仍然可以正常访问
	funcList := app.ListFuncMap()
	if funcList == nil {
		t.Error("Function list should not be nil after concurrent operations")
	}
}

// BenchmarkAddFuncMap 基准测试AddFuncMap性能
func BenchmarkAddFuncMap(b *testing.B) {
	app := NewApp()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		funcName := "benchFunc"
		app.AddFuncMap(funcName, func(s string) string { return s })
		app.RemoveFuncMap(funcName)
	}
}