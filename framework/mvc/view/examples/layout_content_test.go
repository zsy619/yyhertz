package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestLayoutContentEmbedding 测试 {{.LayoutContent}} 布局内容嵌入功能
func TestLayoutContentEmbedding(t *testing.T) {
	// 创建临时目录结构
	tmpDir, err := os.MkdirTemp("", "layout_content_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建视图和布局目录
	viewsDir := filepath.Join(tmpDir, "views")
	layoutsDir := filepath.Join(viewsDir, "layouts")

	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 创建测试布局文件（使用 {{.LayoutContent}} 占位符）
	layoutContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<header>Layout Header</header>
<main>
{{.LayoutContent}}
</main>
<footer>Layout Footer</footer>
</body>
</html>`

	layoutPath := filepath.Join(layoutsDir, "test_layout.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	// 创建测试模板文件
	templateContent := `<h1>Hello {{.Name}}!</h1>
<p>This is the content template.</p>
<p>Current time: {{.Time}}</p>`

	templatePath := filepath.Join(viewsDir, "test_template.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// 创建模板引擎配置
	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{viewsDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = true
	cfg.Performance.EnableCompress = false

	// 初始化模板引擎
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// 准备测试数据
	testData := map[string]any{
		"Title": "Layout Content Test",
		"Name":  "World",
		"Time":  "2024-01-01 12:00:00",
	}

	// 测试使用布局渲染
	result, err := engine.RenderWithLayout("test_template", "layouts/test_layout", testData)
	if err != nil {
		t.Fatalf("Failed to render with layout: %v", err)
	}

	// 验证结果
	t.Logf("Rendered result:\n%s", result)

	// 检查布局结构是否正确嵌入
	if !strings.Contains(result, "<header>Layout Header</header>") {
		t.Error("Layout header not found in result")
	}

	if !strings.Contains(result, "<footer>Layout Footer</footer>") {
		t.Error("Layout footer not found in result")
	}

	// 检查内容是否正确嵌入
	if !strings.Contains(result, "<h1>Hello World!</h1>") {
		t.Error("Template content not properly embedded")
	}

	if !strings.Contains(result, "<p>This is the content template.</p>") {
		t.Error("Template paragraph not found")
	}

	// 检查标题是否正确渲染
	if !strings.Contains(result, "<title>Layout Content Test</title>") {
		t.Error("Title not properly rendered in layout")
	}

	// 确保 {{.LayoutContent}} 占位符被完全替换
	if strings.Contains(result, "{{.LayoutContent}}") {
		t.Error("{{.LayoutContent}} placeholder was not replaced")
	}
}

// TestMultipleLayoutContentPlaceholders 测试多种布局内容占位符格式
func TestMultipleLayoutContentPlaceholders(t *testing.T) {

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "multi_placeholder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	viewsDir := filepath.Join(tmpDir, "views")
	layoutsDir := filepath.Join(viewsDir, "layouts")

	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 测试不同的占位符格式
	testCases := []struct {
		name        string
		placeholder string
	}{
		{"LayoutContent", "{{.LayoutContent}}"},
		{"Content", "{{.Content}}"},
		{"SpacedLayoutContent", "{{ .LayoutContent }}"},
		{"SpacedContent", "{{ .Content }}"},
		{"BeegoTemplate", "{{template \"content\" .}}"},
		{"RailsYield", "{{yield}}"},
	}

	templateContent := `<div>Template Content</div>`
	templatePath := filepath.Join(viewsDir, "multi_test.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{viewsDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = false // 禁用缓存以便测试多个布局

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建使用特定占位符的布局
			layoutContent := `<html><body><header>Header</header>` + tc.placeholder + `<footer>Footer</footer></body></html>`

			layoutPath := filepath.Join(layoutsDir, tc.name+"_layout.html")
			if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
				t.Fatalf("Failed to write layout file: %v", err)
			}

			// 渲染测试
			result, err := engine.RenderWithLayout("multi_test", tc.name+"_layout", map[string]any{})
			if err != nil {
				t.Errorf("Failed to render with %s: %v", tc.placeholder, err)
				return
			}

			// 验证内容是否正确嵌入
			if !strings.Contains(result, "<div>Template Content</div>") {
				t.Errorf("Template content not found for placeholder %s", tc.placeholder)
			}

			if !strings.Contains(result, "<header>Header</header>") {
				t.Errorf("Layout header not found for placeholder %s", tc.placeholder)
			}

			if !strings.Contains(result, "<footer>Footer</footer>") {
				t.Errorf("Layout footer not found for placeholder %s", tc.placeholder)
			}

			// 确保占位符被替换（除了Beego模板语法，它会被转换）
			if tc.name != "BeegoTemplate" && strings.Contains(result, tc.placeholder) {
				t.Errorf("Placeholder %s was not replaced in result", tc.placeholder)
			}

			t.Logf("Successfully tested placeholder: %s", tc.placeholder)
		})
	}
}

// TestLayoutContentWithBeegoFunctions 测试布局内容与Beego函数的集成
func TestLayoutContentWithBeegoFunctions(t *testing.T) {

	tmpDir, err := os.MkdirTemp("", "beego_functions_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	viewsDir := filepath.Join(tmpDir, "views")
	layoutsDir := filepath.Join(viewsDir, "layouts")

	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 创建使用Beego函数的布局
	layoutContent := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title | str2html}}</title>
</head>
<body>
    <nav>{{dateformat .Now "2006-01-02 15:04:05"}}</nav>
    <main>{{.LayoutContent}}</main>
    <footer>{{.Message | substr 0 50}}</footer>
</body>
</html>`

	layoutPath := filepath.Join(layoutsDir, "beego_layout.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	// 创建使用Beego函数的模板
	templateContent := `<h1>{{.Name | toupper}}</h1>
<p>Price: {{.Price | fmtFloat2}}</p>
<div>{{.Description | nl2br}}</div>`

	templatePath := filepath.Join(viewsDir, "beego_template.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{viewsDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = true

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// 准备测试数据
	testData := map[string]any{
		"Title":       "Beego Template Test",
		"Name":        "hello world",
		"Price":       123.456,
		"Description": "Line 1\nLine 2\nLine 3",
		"Message":     "This is a very long message that should be truncated by the substr function",
		"Now":         "2024-01-01T12:00:00Z",
	}

	result, err := engine.RenderWithLayout("beego_template", "beego_layout", testData)
	if err != nil {
		t.Fatalf("Failed to render with Beego functions: %v", err)
	}

	t.Logf("Beego functions result:\n%s", result)

	// 验证Beego函数是否正常工作
	if !strings.Contains(result, "HELLO WORLD") {
		t.Error("toupper function not working")
	}

	if !strings.Contains(result, "123.46") {
		t.Error("fmtFloat2 function not working")
	}

	if !strings.Contains(result, "2024-01-01 12:00:00") {
		t.Error("dateformat function not working")
	}

	// 验证布局内容嵌入正常
	if strings.Contains(result, "{{.LayoutContent}}") {
		t.Error("LayoutContent placeholder not replaced")
	}
}
