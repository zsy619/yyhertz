package view

import (
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
)

// TestCSRFTokenAccess 测试CSRF token在模板中的访问方式
func TestCSRFTokenAccess(t *testing.T) {
	// 创建模板引擎
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"./../views"}

	engine, err := NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 测试数据
	testData := &RenderData{
		Data: map[string]interface{}{
			"username": "testuser",
		},
		CSRF: "test-csrf-token-12345",
	}

	// 测试1: 使用 .CSRF 字段访问
	t.Run("TestDirectCSRFFieldAccess", func(t *testing.T) {
		templateContent := `CSRF Token: {{.CSRF}}`
		tmpl, err := engine.createInlineTemplate("csrf_test", templateContent)
		if err != nil {
			t.Fatalf("创建内联模板失败: %v", err)
		}

		result, err := engine.executeTemplate(tmpl, testData)
		if err != nil {
			t.Errorf("执行模板失败: %v", err)
		}

		expected := "CSRF Token: test-csrf-token-12345"
		if result != expected {
			t.Errorf("期望: %s, 实际: %s", expected, result)
		}
	})

	// 测试2: 使用 .CsrfToken 字段访问
	t.Run("TestCSRFFieldAccess", func(t *testing.T) {
		templateContent := `CSRF Token: {{.CsrfToken}}`
		tmpl, err := engine.createInlineTemplate("csrf_field_test", templateContent)
		if err != nil {
			t.Fatalf("创建内联模板失败: %v", err)
		}

		// 需要先通过prepareRenderData处理数据，确保字段同步
		preparedData := engine.prepareRenderData(testData)

		result, err := engine.executeTemplate(tmpl, preparedData)
		if err != nil {
			t.Errorf("执行模板失败: %v", err)
		}

		expected := "CSRF Token: test-csrf-token-12345"
		if result != expected {
			t.Errorf("期望: %s, 实际: %s", expected, result)
		}
	})

	// 测试3: 使用 .Csrf_token() 方法访问（下划线命名）
	t.Run("TestCSRFUnderscoreMethodAccess", func(t *testing.T) {
		templateContent := `CSRF Token: {{.Csrf_token}}`
		tmpl, err := engine.createInlineTemplate("csrf_underscore_test", templateContent)
		if err != nil {
			t.Fatalf("创建内联模板失败: %v", err)
		}

		result, err := engine.executeTemplate(tmpl, testData)
		if err != nil {
			t.Errorf("执行模板失败: %v", err)
		}

		expected := "CSRF Token: test-csrf-token-12345"
		if result != expected {
			t.Errorf("期望: %s, 实际: %s", expected, result)
		}
	})

	// 测试4: 使用 csrf 模板函数访问
	t.Run("TestCSRFFunctionAccess", func(t *testing.T) {
		templateContent := `CSRF Token: {{csrf}}`
		tmpl, err := engine.createInlineTemplate("csrf_func_test", templateContent)
		if err != nil {
			t.Fatalf("创建内联模板失败: %v", err)
		}

		result, err := engine.executeTemplate(tmpl, testData)
		if err != nil {
			t.Errorf("执行模板失败: %v", err)
		}

		// 模板函数返回的是占位符，因为没有真实的请求上下文
		if !strings.Contains(result, "csrf-token-placeholder") {
			t.Errorf("期望包含 csrf-token-placeholder, 实际: %s", result)
		}
	})

	// 测试5: 使用 csrf_token 模板函数访问
	t.Run("TestCSRFTokenFunctionAccess", func(t *testing.T) {
		templateContent := `CSRF Token: {{csrf_token}}`
		tmpl, err := engine.createInlineTemplate("csrf_token_func_test", templateContent)
		if err != nil {
			t.Fatalf("创建内联模板失败: %v", err)
		}

		result, err := engine.executeTemplate(tmpl, testData)
		if err != nil {
			t.Errorf("执行模板失败: %v", err)
		}

		// 模板函数返回的是占位符，因为没有真实的请求上下文
		if !strings.Contains(result, "csrf-token-placeholder") {
			t.Errorf("期望包含 csrf-token-placeholder, 实际: %s", result)
		}
	})
}

// TestPrepareRenderDataCSRF 测试prepareRenderData方法中的CSRF token处理
func TestPrepareRenderDataCSRF(t *testing.T) {
	cfg := config.GlobalTemplate
	engine, err := NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 测试1: 传入普通数据时应该自动设置CSRF token
	t.Run("TestAutoSetCSRF", func(t *testing.T) {
		data := map[string]interface{}{
			"username": "testuser",
		}

		renderData := engine.prepareRenderData(data)

		if renderData.CSRF == "" {
			t.Error("期望自动设置CSRF token，但为空")
		}

		if renderData.Data == nil {
			t.Error("原始数据应该保持不变")
		}
	})

	// 测试2: 传入已有RenderData且CSRF为空时应该设置
	t.Run("TestSetCSRFForEmptyRenderData", func(t *testing.T) {
		data := &RenderData{
			Data: "test data",
			CSRF: "", // 空值
		}

		renderData := engine.prepareRenderData(data)

		if renderData.CSRF == "" {
			t.Error("期望设置CSRF token，但仍为空")
		}
	})

	// 测试3: 传入已有RenderData且CSRF不为空时应该保持原值
	t.Run("TestKeepExistingCSRF", func(t *testing.T) {
		originalCSRF := "existing-csrf-token"
		data := &RenderData{
			Data: "test data",
			CSRF: originalCSRF,
		}

		renderData := engine.prepareRenderData(data)

		if renderData.CSRF != originalCSRF {
			t.Errorf("期望保持原有CSRF token '%s'，实际为 '%s'", originalCSRF, renderData.CSRF)
		}
	})
}
