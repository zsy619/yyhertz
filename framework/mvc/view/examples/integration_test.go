package main

import (
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestRealTemplateCSRFAccess 测试真实模板中的CSRF访问
func TestRealTemplateCSRFAccess(t *testing.T) {
	// 创建模板引擎
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 测试使用带有CSRF token的LoginWithCSRF模板
	t.Run("TestLoginWithCSRFTemplate", func(t *testing.T) {
		testData := &view.RenderData{
			Data: map[string]interface{}{
				"username": "testuser",
				"message":  "欢迎登录系统",
			},
			CSRF: "real-csrf-token-12345",
		}

		html, err := engine.RenderWithLayout("Login/LoginWithCSRF", "", testData)
		if err != nil {
			t.Errorf("渲染LoginWithCSRF模板失败: %v", err)
			return
		}

		if len(html) == 0 {
			t.Error("渲染结果为空")
			return
		}

		// 验证基本内容
		expectedContents := []string{
			"用户登录",
			"用户名:",
			"密码:",
			"CSRF Token 测试:",
		}

		for _, content := range expectedContents {
			if !strings.Contains(html, content) {
				t.Errorf("渲染结果不包含预期内容: %s", content)
			}
		}

		// 验证CSRF token的不同访问方式
		csrfTests := []struct {
			name     string
			expected string
		}{
			{"直接访问 .CSRF", "real-csrf-token-12345"},
			{"方法访问 .CsrfToken", "real-csrf-token-12345"},
			{"下划线方法 .Csrf_token", "real-csrf-token-12345"},
			{"模板函数 csrf", "csrf-token-placeholder"},       // 模板函数返回占位符
			{"模板函数 csrf_token", "csrf-token-placeholder"}, // 模板函数返回占位符
		}

		for _, test := range csrfTests {
			if !strings.Contains(html, test.expected) {
				t.Errorf("%s: 期望包含 '%s'，但在渲染结果中未找到", test.name, test.expected)
				// 输出部分HTML用于调试
				lines := strings.Split(html, "\n")
				for i, line := range lines {
					if strings.Contains(line, "CSRF") || strings.Contains(line, "csrf") {
						t.Logf("第%d行: %s", i+1, strings.TrimSpace(line))
					}
				}
			}
		}

		// 验证隐藏的CSRF字段
		expectedHiddenField := `<input type="hidden" name="csrf_token" value="real-csrf-token-12345"`
		if !strings.Contains(html, expectedHiddenField) {
			t.Error("期望包含CSRF隐藏字段，但未找到")
		}

		t.Logf("模板渲染成功，HTML长度: %d", len(html))
	})
}

// TestCSRFTokenErrorRecovery 测试CSRF token错误恢复
func TestCSRFTokenErrorRecovery(t *testing.T) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 测试：当传入的数据不包含CSRF时，系统应该自动添加
	t.Run("TestAutoRecoveryFromNilCSRF", func(t *testing.T) {
		// 传入普通map数据，没有CSRF信息
		basicData := map[string]interface{}{
			"username": "testuser",
		}

		html, err := engine.RenderWithLayout("Login/LoginWithCSRF", "", basicData)
		if err != nil {
			t.Errorf("渲染模板失败: %v", err)
			return
		}

		// 应该能正常渲染，并且包含占位符CSRF token
		if !strings.Contains(html, "csrf-token-placeholder") {
			t.Error("期望系统自动填充CSRF token占位符")
		}

		t.Log("成功从无CSRF数据中恢复并渲染")
	})

	// 测试：当RenderData的CSRF为空时，系统应该填充占位符
	t.Run("TestAutoRecoveryFromEmptyCSRF", func(t *testing.T) {
		emptyCSRFData := &view.RenderData{
			Data: map[string]interface{}{
				"username": "testuser",
			},
			CSRF: "", // 空的CSRF
		}

		html, err := engine.RenderWithLayout("Login/LoginWithCSRF", "", emptyCSRFData)
		if err != nil {
			t.Errorf("渲染模板失败: %v", err)
			return
		}

		// 应该能正常渲染，并且包含占位符CSRF token
		if !strings.Contains(html, "csrf-token-placeholder") {
			t.Error("期望系统自动填充空CSRF为占位符")
		}

		t.Log("成功从空CSRF中恢复并渲染")
	})
}
