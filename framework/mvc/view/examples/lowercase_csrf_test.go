package main

import (
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestLowercaseCsrfTokenAccess 测试小写的.csrf_token访问是否仍然有问题
func TestLowercaseCsrfTokenAccess(t *testing.T) {
	cfg := view.DefaultTemplateConfig()
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 准备数据
	testData := &view.RenderData{
		CSRF:      "test-csrf-token-123",
		CsrfToken: "test-csrf-token-456", // 不同值用于区分
		Data:      "test data",
	}

	// 测试小写的 .csrf_token（这应该仍然不工作）
	t.Run("TestLowercaseCsrfTokenStillFails", func(t *testing.T) {
		templateContent := `{{.csrf_token}}` // 全小写
		tmpl, err := engine.CreateInlineTemplate("lowercase_test", templateContent)
		if err != nil {
			t.Fatalf("创建模板失败: %v", err)
		}

		_, err = engine.ExecuteTemplate(tmpl, testData)
		if err == nil {
			t.Error("期望 .csrf_token (全小写) 仍然会失败，但却成功了")
		} else {
			t.Logf("正确：.csrf_token (全小写) 仍然失败: %v", err)
		}
	})

	// 测试大写的 .CsrfToken（这应该工作）
	t.Run("TestCamelCaseCsrfTokenWorks", func(t *testing.T) {
		templateContent := `{{.CsrfToken}}` // 驼峰命名
		tmpl, err := engine.CreateInlineTemplate("camelcase_test", templateContent)
		if err != nil {
			t.Fatalf("创建模板失败: %v", err)
		}

		result, err := engine.ExecuteTemplate(tmpl, testData)
		if err != nil {
			t.Errorf("期望 .CsrfToken 成功，但失败了: %v", err)
		} else {
			expected := "test-csrf-token-456"
			if result != expected {
				t.Errorf("期望结果: %s, 实际结果: %s", expected, result)
			}
			t.Logf("正确：.CsrfToken 成功返回: %s", result)
		}
	})

	// 给出解决方案建议
	t.Run("TestSolutionSuggestions", func(t *testing.T) {
		solutions := []struct {
			name     string
			template string
			works    bool
		}{
			{"使用.CsrfToken字段", "{{.CsrfToken}}", true},
			{"使用.CSRF字段", "{{.CSRF}}", true},
			{"使用.Csrf_token方法", "{{.Csrf_token}}", true},
			{"使用csrf_token函数", "{{csrf_token}}", true},
			{"使用原始.csrf_token", "{{.csrf_token}}", false},
		}

		for _, solution := range solutions {
			tmpl, err := engine.CreateInlineTemplate(solution.name, solution.template)
			if err != nil {
				if solution.works {
					t.Errorf("%s: 创建模板失败: %v", solution.name, err)
				}
				continue
			}

			result, err := engine.ExecuteTemplate(tmpl, testData)
			if solution.works {
				if err != nil {
					t.Errorf("%s: 期望成功但失败了: %v", solution.name, err)
				} else {
					t.Logf("✅ %s 成功: %s", solution.name, result)
				}
			} else {
				if err == nil {
					t.Errorf("%s: 期望失败但成功了: %s", solution.name, result)
				} else {
					t.Logf("❌ %s 正确地失败了: %v", solution.name, err)
				}
			}
		}
	})
}

// TestPracticalUsageExample 演示实际使用场景
func TestPracticalUsageExample(t *testing.T) {
	cfg := view.DefaultTemplateConfig()
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 模拟真实的使用场景
	t.Run("TestLoginFormExample", func(t *testing.T) {
		// 这是用户可能会写的模板内容
		loginFormTemplate := `
<form method="POST" action="/login">
    <input type="hidden" name="csrf_token" value="{{.CsrfToken}}" />
    <input type="text" name="username" value="{{.Data.username}}" />
    <button type="submit">登录</button>
</form>
`

		// 准备数据 - 模拟从控制器传来的数据
		plainData := map[string]interface{}{
			"username": "johndoe",
			"message":  "欢迎登录",
		}

		// 通过prepareRenderData处理（这会自动设置CSRF token）
		renderData := engine.PrepareRenderData(plainData)

		tmpl, err := engine.CreateInlineTemplate("login_form", loginFormTemplate)
		if err != nil {
			t.Fatalf("创建登录表单模板失败: %v", err)
		}

		result, err := engine.ExecuteTemplate(tmpl, renderData)
		if err != nil {
			t.Errorf("渲染登录表单失败: %v", err)
		} else {
			t.Logf("✅ 登录表单渲染成功，长度: %d", len(result))

			// 验证包含CSRF token
			if renderData.CsrfToken != "" &&
				renderData.CSRF == renderData.CsrfToken {
				t.Log("✅ CSRF token字段同步正确")
			} else {
				t.Error("❌ CSRF token字段同步失败")
			}
		}
	})
}
