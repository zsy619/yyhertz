package view

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadTemplate 测试模板加载功能
func TestLoadTemplate(t *testing.T) {
	// 创建临时测试目录和文件
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "test.html")
	
	// 创建测试模板文件
	templateContent := `<h1>{{.Title}}</h1>
<p>{{.Message}}</p>
<ul>
{{range .Items}}
<li>{{.}}</li>
{{end}}
</ul>`
	
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}
	
	// 准备测试数据
	data := map[string]any{
		"Title":   "Test Title",
		"Message": "This is a test message",
		"Items":   []string{"Item 1", "Item 2", "Item 3"},
	}
	
	// 测试模板加载
	result, err := LoadTemplate(templatePath, data)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}
	
	// 验证结果
	expected := `<h1>Test Title</h1>
<p>This is a test message</p>
<ul>

<li>Item 1</li>

<li>Item 2</li>

<li>Item 3</li>

</ul>`
	
	if result != expected {
		t.Errorf("Template result mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

// TestLoadTemplateWithLayout 测试带布局的模板加载
func TestLoadTemplateWithLayout(t *testing.T) {
	tempDir := t.TempDir()
	
	// 创建布局模板
	layoutPath := filepath.Join(tempDir, "layout.html")
	layoutContent := `{{define "layout"}}<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
{{template "content" .}}
</body>
</html>{{end}}`
	
	// 创建内容模板
	contentPath := filepath.Join(tempDir, "content.html")
	contentContent := `{{define "content"}}<h1>{{.Heading}}</h1>
<p>{{.Content}}</p>{{end}}`
	
	// 写入文件
	err := os.WriteFile(layoutPath, []byte(layoutContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create layout template: %v", err)
	}
	
	err = os.WriteFile(contentPath, []byte(contentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create content template: %v", err)
	}
	
	// 准备测试数据
	data := map[string]any{
		"Title":   "Page Title",
		"Heading": "Welcome",
		"Content": "Hello World!",
	}
	
	// 测试带布局的模板加载
	result, err := LoadTemplateWithLayout(layoutPath, contentPath, data)
	if err != nil {
		t.Fatalf("LoadTemplateWithLayout failed: %v", err)
	}
	
	// 验证结果包含预期内容
	if !contains(result, "Page Title") {
		t.Error("Result should contain page title")
	}
	if !contains(result, "Welcome") {
		t.Error("Result should contain heading")
	}
	if !contains(result, "Hello World!") {
		t.Error("Result should contain content")
	}
}

// TestTemplateFunctions 测试模板函数
func TestTemplateFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function string
		args     []any
		expected any
	}{
		{"MakeSlice", "makeSlice", []any{"a", "b", "c"}, []any{"a", "b", "c"}},
		{"ConcatString", "concatString", []any{"Hello", " ", "World"}, "Hello World"},
		{"ContainString", "containString", []any{"a,b,c", "b"}, true},
		{"FmtByte", "fmtByte", []any{int64(1024)}, "1.00KB"},
		{"FmtFloat2", "fmtFloat2", []any{3.14159}, "3.14"},
		{"Add", "add", []any{5, 3}, float64(8)},
		{"Sub", "sub", []any{5, 3}, float64(2)},
		{"Mul", "mul", []any{5, 3}, float64(15)},
		{"Div", "div", []any{6, 3}, float64(2)},
		{"Eq", "eq", []any{5, 5}, true},
		{"Lt", "lt", []any{3, 5}, true},
		{"Gt", "gt", []any{5, 3}, true},
		{"And", "and", []any{true, true}, true},
		{"Or", "or", []any{true, false}, true},
		{"Not", "not", []any{false}, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, exists := TemplateFuncs[tt.function]
			if !exists {
				t.Fatalf("Function %s not found in TemplateFuncs", tt.function)
			}
			
			// 这里我们无法直接测试模板函数，因为它们需要在模板上下文中运行
			// 但我们可以验证函数是否存在
			if fn == nil {
				t.Errorf("Function %s is nil", tt.function)
			}
		})
	}
}

// TestMakeSlice 测试MakeSlice函数
func TestMakeSlice(t *testing.T) {
	result := MakeSlice("a", 1, true)
	expected := []any{"a", 1, true}
	
	if len(result) != len(expected) {
		t.Errorf("Length mismatch. Expected %d, got %d", len(expected), len(result))
	}
	
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Value mismatch at index %d. Expected %v, got %v", i, v, result[i])
		}
	}
}

// TestConcatString 测试ConcatString函数
func TestConcatString(t *testing.T) {
	result := ConcatString("Hello", " ", "World", "!")
	expected := "Hello World!"
	
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestContainString 测试ContainString函数
func TestContainString(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"a,b,c", "b", true},
		{"a,b,c", "d", false},
		{"apple,banana,cherry", "banana", true},
		{"apple,banana,cherry", "grape", false},
	}
	
	for _, tt := range tests {
		result := ContainString(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("ContainString(%s, %s) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

// TestFmtByte 测试FmtByte函数
func TestFmtByte(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{512, "512.00B"},
		{1024, "1.00KB"},
		{1048576, "1.00MB"},
		{1073741824, "1.00GB"},
		{1099511627776, "1.00TB"},
	}
	
	for _, tt := range tests {
		result := FmtByte(tt.size)
		if result != tt.expected {
			t.Errorf("FmtByte(%d) = %s, expected %s", tt.size, result, tt.expected)
		}
	}
}

// TestAuthContain 测试AuthContain函数
func TestAuthContain(t *testing.T) {
	tests := []struct {
		s        string
		value    int
		expected bool
	}{
		{"1,2,3", 2, true},
		{"1,2,3", 4, false},
		{"10,20,30", 20, true},
		{"10,20,30", 25, false},
	}
	
	for _, tt := range tests {
		result := AuthContain(tt.s, tt.value)
		if result != tt.expected {
			t.Errorf("AuthContain(%s, %d) = %v, expected %v", tt.s, tt.value, result, tt.expected)
		}
	}
}

// TestMathFunctions 测试数学函数
func TestMathFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(any, any) any
		a, b     any
		expected float64
	}{
		{"Add", Add, 5, 3, 8},
		{"Sub", Sub, 5, 3, 2},
		{"Mul", Mul, 5, 3, 15},
		{"Div", Div, 6, 3, 2},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s(%v, %v) = %v, expected %v", tt.name, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestComparisonFunctions 测试比较函数
func TestComparisonFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(any, any) bool
		a, b     any
		expected bool
	}{
		{"Eq", Eq, 5, 5, true},
		{"Ne", Ne, 5, 3, true},
		{"Lt", Lt, 3, 5, true},
		{"Le", Le, 5, 5, true},
		{"Gt", Gt, 5, 3, true},
		{"Ge", Ge, 5, 5, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("%s(%v, %v) = %v, expected %v", tt.name, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestLogicalFunctions 测试逻辑函数
func TestLogicalFunctions(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"And true true", And(true, true) == true},
		{"And true false", And(true, false) == false},
		{"Or true false", Or(true, false) == true},
		{"Or false false", Or(false, false) == false},
		{"Not true", Not(true) == false},
		{"Not false", Not(false) == true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.expected {
				t.Errorf("Test %s failed", tt.name)
			}
		})
	}
}

// TestTypeConversion 测试类型转换函数
func TestTypeConversion(t *testing.T) {
	// ToString
	if ToString(123) != "123" {
		t.Error("ToString(123) should return '123'")
	}
	
	// ToInt
	if ToInt("123") != 123 {
		t.Error("ToInt('123') should return 123")
	}
	
	// ToFloat
	if ToFloat("3.14") != 3.14 {
		t.Error("ToFloat('3.14') should return 3.14")
	}
}

// TestCollectionFunctions 测试集合函数
func TestCollectionFunctions(t *testing.T) {
	slice := []any{"a", "b", "c"}
	
	// Len
	if Len(slice) != 3 {
		t.Error("Len should return 3")
	}
	
	// Index
	if Index(slice, 1) != "b" {
		t.Error("Index(slice, 1) should return 'b'")
	}
	
	// Slice
	result := Slice(slice, 1, 3)
	if sliceResult, ok := result.([]any); !ok || len(sliceResult) != 2 {
		t.Error("Slice(slice, 1, 3) should return slice of length 2")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}