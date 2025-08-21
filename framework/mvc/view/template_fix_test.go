package view

import (
	"testing"
)

// TestTemplateLoading 测试模板加载功能
func TestTemplateLoading(t *testing.T) {
	// 启用调试模式（如果可用）
	// config.SetGlobalLogLevel("debug")

	// 创建模板引擎
	cfg := DefaultTemplateConfig()
	cfg.Paths.ViewPaths = []string{"../../views"} // 指向我们创建的views目录

	engine, err := NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 测试加载Login模板
	t.Run("LoadLoginTemplate", func(t *testing.T) {
		tmpl, err := engine.getTemplate("Login/Login")
		if err != nil {
			t.Errorf("加载Login模板失败: %v", err)
			return
		}

		if tmpl == nil {
			t.Error("模板为空")
			return
		}

		t.Logf("成功加载模板: %s", tmpl.Name())
	})

	// 测试渲染Login模板
	t.Run("RenderLoginTemplate", func(t *testing.T) {
		testData := map[string]interface{}{
			"username": "testuser",
			"message":  "欢迎使用系统",
		}

		html, err := engine.RenderWithLayout("Login/Login", "", testData)
		if err != nil {
			t.Errorf("渲染Login模板失败: %v", err)
			return
		}

		if len(html) == 0 {
			t.Error("渲染结果为空")
			return
		}

		t.Logf("渲染成功，HTML长度: %d", len(html))

		// 检查是否包含预期内容
		if !containsStr(html, "用户登录") {
			t.Error("渲染结果不包含预期的标题")
		}

		if !containsStr(html, "用户名") {
			t.Error("渲染结果不包含用户名字段")
		}
	})

	// 测试加载不存在的模板
	t.Run("LoadNonExistentTemplate", func(t *testing.T) {
		_, err := engine.getTemplate("NonExistent/Template")
		if err == nil {
			t.Error("期望加载不存在的模板时返回错误")
		} else {
			t.Logf("正确处理不存在的模板: %v", err)
		}
	})
}

// containsStr 检查字符串是否包含子字符串
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr) >= 0
}

// findSubstr 查找子字符串位置
func findSubstr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestTemplatePathResolution 测试模板路径解析
func TestTemplatePathResolution(t *testing.T) {
	cfg := DefaultTemplateConfig()
	cfg.Paths.ViewPaths = []string{"../../views", "./views", "templates"}

	engine, err := NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 测试查找文件功能
	t.Run("FindExistingFile", func(t *testing.T) {
		path, err := engine.findTemplateFile("Login/Login")
		if err != nil {
			t.Errorf("查找已存在文件失败: %v", err)
		} else {
			t.Logf("找到模板文件: %s", path)
		}
	})

	t.Run("FindNonExistentFile", func(t *testing.T) {
		_, err := engine.findTemplateFile("NonExistent/File")
		if err == nil {
			t.Error("期望查找不存在文件时返回错误")
		} else {
			t.Logf("正确处理不存在文件: %v", err)
		}
	})
}
