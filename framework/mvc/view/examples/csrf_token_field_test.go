package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestCsrfTokenFieldAccess 测试CsrfToken字段的访问方式
func TestCsrfTokenFieldAccess(t *testing.T) {
	// 创建模板引擎
	cfg := config.GlobalTemplate
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 测试数据
	testData := &view.RenderData{
		Data: map[string]interface{}{
			"username": "testuser",
		},
		CSRF:      "test-csrf-123",
		CsrfToken: "test-csrf-456", // 不同的值用于区分
	}

	// 测试1: 直接访问 .CsrfToken 字段
	t.Run("TestDirectCsrfTokenFieldAccess", func(t *testing.T) {
		templateContent := `CSRF Token: {{.CsrfToken}}`
		tmpl, err := engine.CreateInlineTemplate("csrf_field_test", templateContent)
		if err != nil {
			t.Fatalf("创建内联模板失败: %v", err)
		}

		result, err := engine.ExecuteTemplate(tmpl, testData)
		if err != nil {
			t.Errorf("执行模板失败: %v", err)
		}

		expected := "CSRF Token: test-csrf-456"
		if result != expected {
			t.Errorf("期望: %s, 实际: %s", expected, result)
		}
	})

	// 测试2: 验证prepareRenderData自动同步两个字段
	t.Run("TestPrepareRenderDataSyncFields", func(t *testing.T) {
		// 传入只有CSRF字段的数据
		inputData := &view.RenderData{
			Data: "test data",
			CSRF: "original-csrf-token",
			// CsrfToken 留空
		}

		prepared := engine.PrepareRenderData(inputData)

		if prepared.CSRF != prepared.CsrfToken {
			t.Errorf("CSRF字段与CsrfToken字段不同步: CSRF=%s, CsrfToken=%s",
				prepared.CSRF, prepared.CsrfToken)
		}

		if prepared.CsrfToken != "original-csrf-token" {
			t.Errorf("期望CsrfToken为 'original-csrf-token', 实际为 '%s'", prepared.CsrfToken)
		}
	})

	// 测试3: 验证从普通数据创建RenderData时的字段同步
	t.Run("TestCreateFromPlainDataSyncFields", func(t *testing.T) {
		plainData := map[string]interface{}{
			"username": "testuser",
		}

		prepared := engine.PrepareRenderData(plainData)

		if prepared.CSRF != prepared.CsrfToken {
			t.Errorf("从普通数据创建时CSRF字段不同步: CSRF=%s, CsrfToken=%s",
				prepared.CSRF, prepared.CsrfToken)
		}

		if prepared.CsrfToken == "" {
			t.Error("期望CsrfToken不为空")
		}
	})
}

// TestRealTemplateWithCsrfTokenField 测试真实模板中使用CsrfToken字段
func TestRealTemplateWithCsrfTokenField(t *testing.T) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 创建一个使用 .CsrfToken 的测试模板
	testTemplateContent := `
<!DOCTYPE html>
<html>
<head><title>CSRF Field Test</title></head>
<body>
    <form method="POST">
        <input type="hidden" name="csrf_token" value="{{.CsrfToken}}" />
        <input type="text" name="username" />
        <button type="submit">Submit</button>
    </form>
    
    <!-- 测试不同的访问方式 -->
    <div style="display: none;">
        <p>CSRF字段: {{.CSRF}}</p>
        <p>CsrfToken字段: {{.CsrfToken}}</p>
        <p>Csrf_token方法: {{.Csrf_token}}</p>
        <p>csrf_token函数: {{csrf_token}}</p>
    </div>
</body>
</html>`

	tmpl, err := engine.CreateInlineTemplate("csrf_field_real_test", testTemplateContent)
	if err != nil {
		t.Fatalf("创建模板失败: %v", err)
	}

	testData := &view.RenderData{
		CSRF:      "real-csrf-token-789",
		CsrfToken: "should-be-overridden", // 这应该被prepareRenderData同步
		Data: map[string]interface{}{
			"username": "testuser",
		},
	}

	// 通过prepareRenderData处理数据
	preparedData := engine.PrepareRenderData(testData)

	result, err := engine.ExecuteTemplate(tmpl, preparedData)
	if err != nil {
		t.Errorf("执行模板失败: %v", err)
		return
	}

	// 验证结果包含正确的CSRF token
	expectedToken := "real-csrf-token-789"
	if !strings.Contains(result, fmt.Sprintf(`value="%s"`, expectedToken)) {
		t.Errorf("期望结果包含 value=\"%s\"，但未找到", expectedToken)
	}

	// 验证两个字段都有相同的值
	if !strings.Contains(result, fmt.Sprintf("CSRF字段: %s", expectedToken)) {
		t.Error("CSRF字段值不正确")
	}

	if !strings.Contains(result, fmt.Sprintf("CsrfToken字段: %s", expectedToken)) {
		t.Error("CsrfToken字段值不正确")
	}

	t.Logf("模板渲染成功，HTML长度: %d", len(result))
}

// TestAllCsrfAccessMethods 测试所有CSRF访问方式的兼容性
func TestAllCsrfAccessMethods(t *testing.T) {
	cfg := config.GlobalTemplate
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 准备数据
	plainData := map[string]interface{}{
		"username": "testuser",
	}
	renderData := engine.PrepareRenderData(plainData)

	// 测试各种访问方式
	testCases := []struct {
		name        string
		template    string
		shouldWork  bool
		description string
	}{
		{"CSRF字段", "{{.CSRF}}", true, "直接访问CSRF字段"},
		{"CsrfToken字段", "{{.CsrfToken}}", true, "访问新增的CsrfToken字段"},
		{"Csrf_token方法", "{{.Csrf_token}}", true, "调用Csrf_token方法"},
		{"GetCSRFToken方法", "{{.GetCSRFToken}}", true, "调用GetCSRFToken方法"},
		{"csrf函数", "{{csrf}}", true, "使用csrf模板函数"},
		{"csrf_token函数", "{{csrf_token}}", true, "使用csrf_token模板函数"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := engine.CreateInlineTemplate(tc.name, tc.template)
			if err != nil {
				if tc.shouldWork {
					t.Errorf("创建模板失败: %v", err)
				}
				return
			}

			result, err := engine.ExecuteTemplate(tmpl, renderData)
			if err != nil {
				if tc.shouldWork {
					t.Errorf("执行模板失败 (%s): %v", tc.description, err)
				}
			} else {
				if tc.shouldWork && result == "" {
					t.Errorf("期望有结果但为空 (%s)", tc.description)
				}
				t.Logf("%s 成功: %s", tc.description, result)
			}
		})
	}
}
