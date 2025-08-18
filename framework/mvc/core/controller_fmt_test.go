package core

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

// 创建测试用的BaseController
func createTestController() *BaseController {
	// 创建简单的Controller，不需要复杂的上下文设置
	return NewBaseController()
}

// 简单的断言辅助函数
func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func assertNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func assertNotNil(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func assertTrue(t *testing.T, condition bool) {
	t.Helper()
	if !condition {
		t.Error("Expected condition to be true")
	}
}

func assertContains(t *testing.T, str, substr string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("Expected '%s' to contain '%s'", str, substr)
	}
}

// ============= Print系列方法测试 =============

func TestPrint(t *testing.T) {
	c := createTestController()

	// 测试Print方法存在性和兼容性
	// 由于直接调用fmt.Print，我们主要验证方法签名
	var printFunc func(...any) (int, error) = c.Print
	if printFunc == nil {
		t.Error("Print method not found on BaseController")
	}
}

func TestPrintf(t *testing.T) {
	c := createTestController()

	// 测试Printf方法存在性和兼容性
	var printfFunc func(string, ...any) (int, error) = c.Printf
	if printfFunc == nil {
		t.Error("Printf method not found on BaseController")
	}
}

func TestPrintln(t *testing.T) {
	c := createTestController()

	// 测试Println方法存在性和兼容性
	var printlnFunc func(...any) (int, error) = c.Println
	if printlnFunc == nil {
		t.Error("Println method not found on BaseController")
	}
}

// ============= Sprint系列方法测试 =============

func TestSprint(t *testing.T) {
	c := createTestController()

	// 测试Sprint方法
	result := c.Sprint("Hello", " ", "World")
	expected := "Hello World"
	assertEqual(t, expected, result)

	// 测试数字Sprint
	result = c.Sprint(123, 456)
	expected = "123 456"
	assertEqual(t, expected, result)
}

func TestSprintf(t *testing.T) {
	c := createTestController()

	// 测试格式化字符串
	result := c.Sprintf("用户ID: %d, 姓名: %s", 123, "张三")
	expected := "用户ID: 123, 姓名: 张三"
	assertEqual(t, expected, result)

	// 测试浮点数格式化
	result = c.Sprintf("进度: %.1f%%", 85.7)
	expected = "进度: 85.7%"
	assertEqual(t, expected, result)
}

func TestSprintln(t *testing.T) {
	c := createTestController()

	// 测试带换行的字符串格式化
	result := c.Sprintln("调试信息:", "test")
	expected := "调试信息: test\n"
	assertEqual(t, expected, result)
}

// ============= Fprint系列方法测试 =============

func TestFprint(t *testing.T) {
	c := createTestController()
	var buf bytes.Buffer

	// 测试Fprint功能
	n, err := c.Fprint(&buf, "Hello", " ", "World")
	assertNil(t, err)
	assertTrue(t, n > 0)

	result := buf.String()
	expected := "Hello World"
	assertEqual(t, expected, result)
}

func TestFprintf(t *testing.T) {
	c := createTestController()
	var buf bytes.Buffer

	// 测试格式化输出到Writer
	n, err := c.Fprintf(&buf, "用户ID: %d, 姓名: %s", 123, "张三")
	assertNil(t, err)
	assertTrue(t, n > 0)

	result := buf.String()
	expected := "用户ID: 123, 姓名: 张三"
	assertEqual(t, expected, result)
}

func TestFprintln(t *testing.T) {
	c := createTestController()
	var buf bytes.Buffer

	// 测试带换行的Writer输出
	n, err := c.Fprintln(&buf, "调试信息:", "test")
	assertNil(t, err)
	assertTrue(t, n > 0)

	result := buf.String()
	expected := "调试信息: test\n"
	assertEqual(t, expected, result)
}

// ============= Scan系列方法测试 =============

func TestScan(t *testing.T) {
	c := createTestController()

	// 测试Scan方法存在性和兼容性
	var scanFunc func(...any) (int, error) = c.Scan
	if scanFunc == nil {
		t.Error("Scan method not found on BaseController")
	}
}

func TestScanf(t *testing.T) {
	c := createTestController()

	// 测试Scanf方法存在性和兼容性
	var scanfFunc func(string, ...any) (int, error) = c.Scanf
	if scanfFunc == nil {
		t.Error("Scanf method not found on BaseController")
	}
}

func TestScanln(t *testing.T) {
	c := createTestController()

	// 测试Scanln方法存在性和兼容性
	var scanlnFunc func(...any) (int, error) = c.Scanln
	if scanlnFunc == nil {
		t.Error("Scanln method not found on BaseController")
	}
}

// ============= Sscan系列方法测试 =============

func TestSscan(t *testing.T) {
	c := createTestController()

	var name string
	var age int
	n, err := c.Sscan("张三 25", &name, &age)

	assertNil(t, err)
	assertEqual(t, 2, n)
	assertEqual(t, "张三", name)
	assertEqual(t, 25, age)
}

func TestSscanf(t *testing.T) {
	c := createTestController()

	var name string
	var age int
	// 使用简单的格式化字符串，避免复杂解析
	n, err := c.Sscanf("张三 25", "%s %d", &name, &age)

	assertNil(t, err)
	assertEqual(t, 2, n)
	assertEqual(t, "张三", name)
	assertEqual(t, 25, age)
}

func TestSscanln(t *testing.T) {
	c := createTestController()

	var name string
	var age int
	n, err := c.Sscanln("张三 25\n其他内容", &name, &age)

	assertNil(t, err)
	assertEqual(t, 2, n)
	assertEqual(t, "张三", name)
	assertEqual(t, 25, age)
}

// ============= FormatOutput增强测试 =============

func TestFormatOutput(t *testing.T) {
	c := createTestController()

	// 测试FormatOutput方法与fmt.Sprintf的兼容性
	testData := []any{"hello", 123, 3.14, true, []int{1, 2, 3}}

	for _, data := range testData {
		controllerResult := c.FormatOutput(data)
		fmtResult := fmt.Sprintf("%v", data)
		assertEqual(t, fmtResult, controllerResult)
	}

	// 测试nil值
	result := c.FormatOutput(nil)
	expected := fmt.Sprintf("%v", nil)
	assertEqual(t, expected, result)
}

// ============= 兼容性测试 =============

func TestFmtCompatibility(t *testing.T) {
	c := createTestController()

	// 测试与标准fmt包的兼容性
	testData := []interface{}{"hello", 123, 3.14, true}

	// Sprint系列应该与fmt包产生相同结果
	for _, data := range testData {
		controllerResult := c.Sprint(data)
		fmtResult := fmt.Sprint(data)
		assertEqual(t, fmtResult, controllerResult)
	}

	// Sprintf测试
	format := "数据: %v, 类型: %T"
	for _, data := range testData {
		controllerResult := c.Sprintf(format, data, data)
		fmtResult := fmt.Sprintf(format, data, data)
		assertEqual(t, fmtResult, controllerResult)
	}
}

func TestFprintCompatibility(t *testing.T) {
	c := createTestController()

	testData := "Hello World"

	// 测试Fprint兼容性
	var buf1, buf2 bytes.Buffer
	n1, err1 := c.Fprint(&buf1, testData)
	n2, err2 := fmt.Fprint(&buf2, testData)

	assertEqual(t, err2, err1)
	assertEqual(t, n2, n1)
	assertEqual(t, buf2.String(), buf1.String())
}

func TestSscanCompatibility(t *testing.T) {
	c := createTestController()

	testStr := "张三 25 true"
	var name1, name2 string
	var age1, age2 int
	var flag1, flag2 bool

	// 测试Sscan兼容性
	n1, err1 := c.Sscan(testStr, &name1, &age1, &flag1)
	n2, err2 := fmt.Sscan(testStr, &name2, &age2, &flag2)

	assertEqual(t, err2, err1)
	assertEqual(t, n2, n1)
	assertEqual(t, name2, name1)
	assertEqual(t, age2, age1)
	assertEqual(t, flag2, flag1)
}

// ============= API存在性验证测试 =============

func TestAllMethodsExist(t *testing.T) {
	c := createTestController()

	// 验证所有fmt兼容方法都存在且签名正确
	
	// Print系列
	var _ func(...any) (int, error) = c.Print
	var _ func(string, ...any) (int, error) = c.Printf
	var _ func(...any) (int, error) = c.Println

	// Sprint系列
	var _ func(...any) string = c.Sprint
	var _ func(string, ...any) string = c.Sprintf
	var _ func(...any) string = c.Sprintln

	// Fprint系列
	var _ func(io.Writer, ...any) (int, error) = c.Fprint
	var _ func(io.Writer, string, ...any) (int, error) = c.Fprintf
	var _ func(io.Writer, ...any) (int, error) = c.Fprintln

	// Scan系列
	var _ func(...any) (int, error) = c.Scan
	var _ func(string, ...any) (int, error) = c.Scanf
	var _ func(...any) (int, error) = c.Scanln

	// Sscan系列
	var _ func(string, ...any) (int, error) = c.Sscan
	var _ func(string, string, ...any) (int, error) = c.Sscanf
	var _ func(string, ...any) (int, error) = c.Sscanln

	// FormatOutput
	var _ func(any) string = c.FormatOutput

	t.Log("✅ All fmt-compatible methods exist with correct signatures")
}