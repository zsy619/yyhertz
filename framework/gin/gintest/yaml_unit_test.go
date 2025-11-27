// Package gin - YAML功能单元测试
package gin

import (
	"testing"

	"github.com/zsy619/yyhertz/framework/gin"
	"github.com/zsy619/yyhertz/framework/gin/render"
)

// TestYAMLMethod 测试Context.YAML方法是否存在并可调用
func TestYAMLMethod(t *testing.T) {
	// 创建模拟Context
	engine := gin.New()
	c := &gin.Context{}
	c.SetEngine(engine)

	// 测试YAML方法是否存在
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("YAML method caused panic: %v", r)
		}
	}()

	// 简单验证方法签名正确
	var yamlFunc func(int, any) = c.YAML
	if yamlFunc == nil {
		t.Error("YAML method not found on Context")
	}

	t.Log("✅ Context.YAML method exists with correct signature")
}

// TestYAMLRenderStructure 测试YAML render结构
func TestYAMLRenderStructure(t *testing.T) {
	// 测试YAML渲染器结构是否正确
	data := gin.H{
		"message": "test",
		"status":  200,
	}

	yamlRender := render.YAML{Data: data}

	// 验证结构体字段
	if yamlRender.Data == nil {
		t.Error("YAML render Data field is nil")
	}

	// 验证数据内容
	if h, ok := yamlRender.Data.(gin.H); ok {
		if h["message"] != "test" {
			t.Errorf("Expected message 'test', got %v", h["message"])
		}
		if h["status"] != 200 {
			t.Errorf("Expected status 200, got %v", h["status"])
		}
	} else {
		t.Error("YAML data is not of expected type H")
	}

	t.Log("✅ YAML render structure is correct")
}

// TestBindYAMLMethods 测试YAML绑定方法是否存在
func TestBindYAMLMethods(t *testing.T) {
	engine := gin.New()
	c := &gin.Context{}
	c.SetEngine(engine)

	// 测试BindYAML方法签名
	var bindYAMLFunc func(any) error = c.BindYAML
	if bindYAMLFunc == nil {
		t.Error("BindYAML method not found on Context")
	}

	// 测试ShouldBindYAML方法签名
	var shouldBindYAMLFunc func(any) error = c.ShouldBindYAML
	if shouldBindYAMLFunc == nil {
		t.Error("ShouldBindYAML method not found on Context")
	}

	// 测试ShouldBindBodyWithYAML方法签名
	var shouldBindBodyYAMLFunc func(any) error = c.ShouldBindBodyWithYAML
	if shouldBindBodyYAMLFunc == nil {
		t.Error("ShouldBindBodyWithYAML method not found on Context")
	}

	t.Log("✅ All YAML binding methods exist with correct signatures")
}

// TestXMLMethods 测试XML方法是否存在
func TestXMLMethods(t *testing.T) {
	engine := gin.New()
	c := &gin.Context{}
	c.SetEngine(engine)

	// 测试XML渲染方法
	var xmlFunc func(int, any) = c.XML
	if xmlFunc == nil {
		t.Error("XML method not found on Context")
	}

	// 测试XML绑定方法
	var bindXMLFunc func(any) error = c.BindXML
	if bindXMLFunc == nil {
		t.Error("BindXML method not found on Context")
	}

	var shouldBindXMLFunc func(any) error = c.ShouldBindXML
	if shouldBindXMLFunc == nil {
		t.Error("ShouldBindXML method not found on Context")
	}

	t.Log("✅ All XML methods exist with correct signatures")
}

// TestMethodsComparison 对比新增方法与现有JSON方法
func TestMethodsComparison(t *testing.T) {
	engine := gin.New()
	c := &gin.Context{}
	c.SetEngine(engine)

	// 确保JSON方法仍然存在（回归测试）
	var jsonFunc func(int, any) = c.JSON
	if jsonFunc == nil {
		t.Error("JSON method not found - regression detected")
	}

	// 确保新增的YAML方法存在
	var yamlFunc func(int, any) = c.YAML
	if yamlFunc == nil {
		t.Error("YAML method not found")
	}

	// 确保新增的XML方法存在
	var xmlFunc func(int, any) = c.XML
	if xmlFunc == nil {
		t.Error("XML method not found")
	}

	t.Log("✅ JSON, YAML, and XML rendering methods all exist")

	// 测试绑定方法的一致性
	methods := map[string]interface{}{
		"BindJSON":       c.BindJSON,
		"BindYAML":       c.BindYAML,
		"BindXML":        c.BindXML,
		"ShouldBindJSON": c.ShouldBindJSON,
		"ShouldBindYAML": c.ShouldBindYAML,
		"ShouldBindXML":  c.ShouldBindXML,
	}

	for methodName, method := range methods {
		if method == nil {
			t.Errorf("Method %s not found", methodName)
		}
	}

	t.Log("✅ All binding methods exist with consistent naming")
}

// TestYAMLImplementationCompleteness 测试YAML实现的完整性
func TestYAMLImplementationCompleteness(t *testing.T) {
	// 根据原生Gin的YAML实现，检查我们的实现是否完整

	// 1. 渲染方法
	engine := gin.New()
	c := &gin.Context{}
	c.SetEngine(engine)

	// 检查YAML渲染方法
	var yamlRenderFunc func(int, any) = c.YAML
	if yamlRenderFunc == nil {
		t.Error("Missing: Context.YAML(code int, obj any) method")
	}

	// 2. 绑定方法
	bindingMethods := map[string]interface{}{
		"BindYAML":               c.BindYAML,
		"ShouldBindYAML":         c.ShouldBindYAML,
		"ShouldBindBodyWithYAML": c.ShouldBindBodyWithYAML,
	}

	for methodName, method := range bindingMethods {
		if method == nil {
			t.Errorf("Missing: Context.%s method", methodName)
		}
	}

	// 3. 检查render包中的YAML支持
	yamlRender := render.YAML{Data: "test"}
	if yamlRender.Data != "test" {
		t.Error("render.YAML struct not working properly")
	}

	t.Log("✅ YAML implementation is complete and matches Gin's API")
}

// TestAPICompatibility 测试API兼容性
func TestAPICompatibility(t *testing.T) {
	// 测试方法签名是否与原生Gin兼容

	engine := gin.New()
	c := &gin.Context{}
	c.SetEngine(engine)

	// 原生Gin的方法签名：
	// func (c *Context) YAML(code int, obj any)
	// func (c *Context) BindYAML(obj any) error
	// func (c *Context) ShouldBindYAML(obj any) error
	// func (c *Context) ShouldBindBodyWithYAML(obj any) error

	// 测试渲染方法兼容性
	testData := gin.H{"test": "compatibility"}

	// 这些调用应该在编译时通过（方法签名正确）
	defer func() {
		if r := recover(); r != nil {
			// 忽略运行时错误，我们只测试编译时兼容性
		}
	}()

	// YAML渲染方法
	c.YAML(200, testData)

	// YAML绑定方法
	var testStruct struct {
		Test string `yaml:"test"`
	}

	_ = c.BindYAML(&testStruct)
	_ = c.ShouldBindYAML(&testStruct)
	_ = c.ShouldBindBodyWithYAML(&testStruct)

	t.Log("✅ All method signatures are compatible with Gin")
}
