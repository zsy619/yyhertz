package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestLayoutSections 测试布局区块功能
func TestLayoutSections(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "layout_sections_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	viewsDir := filepath.Join(tmpDir, "views")
	layoutsDir := filepath.Join(viewsDir, "layouts")

	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 创建支持布局区块的布局文件
	layoutContent := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    {{.HtmlHead}}
</head>
<body>
    <nav>{{.Navigation}}</nav>
    <aside>{{.SideBar}}</aside>
    <main>{{.LayoutContent}}</main>
    <footer>{{.Footer}}</footer>
    {{.Scripts}}
</body>
</html>`

	layoutPath := filepath.Join(layoutsDir, "sections_layout.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	// 创建内容模板文件
	templateContent := `<h1>{{.PageTitle}}</h1>
<p>这是页面的主要内容。</p>
<p>用户: {{.User}}</p>`

	templatePath := filepath.Join(viewsDir, "sections_test.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// 创建模板引擎
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
		"Title":     "布局区块测试",
		"PageTitle": "测试页面",
		"User":      "张三",
	}

	// 准备布局区块
	layoutSections := map[string]string{
		"HtmlHead": `
<meta name="description" content="布局区块测试页面">
<style>
    body { font-family: Arial, sans-serif; }
    .highlight { color: red; }
</style>`,
		"Navigation": `
<ul class="nav">
    <li><a href="/">首页</a></li>
    <li><a href="/about">关于</a></li>
    <li><a href="/contact">联系</a></li>
</ul>`,
		"SideBar": `
<div class="sidebar">
    <h3>侧边栏</h3>
    <ul>
        <li>链接1</li>
        <li>链接2</li>
        <li>链接3</li>
    </ul>
</div>`,
		"Footer": `
<p>&copy; 2024 测试公司</p>`,
		"Scripts": `
<script>
    console.log('页面脚本已加载');
    document.addEventListener('DOMContentLoaded', function() {
        console.log('DOM 加载完成');
    });
</script>`,
	}

	// 使用布局区块渲染
	result, err := engine.RenderWithLayoutSections("sections_test", "layouts/sections_layout", testData, layoutSections)
	if err != nil {
		t.Fatalf("Failed to render with layout sections: %v", err)
	}

	t.Logf("Rendered result:\n%s", result)

	// 验证布局区块是否正确嵌入
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"HtmlHead meta tag", `<meta name="description"`, true},
		{"HtmlHead style", `body { font-family: Arial`, true},
		{"Navigation", `<ul class="nav">`, true},
		{"Navigation links", `<a href="/">首页</a>`, true},
		{"SideBar", `<div class="sidebar">`, true},
		{"SideBar content", `<h3>侧边栏</h3>`, true},
		{"Footer", `&copy; 2024 测试公司`, true},
		{"Scripts", `console.log('页面脚本已加载')`, true},
		{"Main content", `<h1>测试页面</h1>`, true},
		{"User data", `用户: 张三`, true},
		{"Title", `<title>布局区块测试</title>`, true},
		{"No placeholder left", `{{.HtmlHead}}`, false},
		{"No placeholder left", `{{.SideBar}}`, false},
		{"No placeholder left", `{{.Scripts}}`, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contains := strings.Contains(result, test.content)
			if contains != test.expected {
				if test.expected {
					t.Errorf("Expected content '%s' not found in result", test.content)
				} else {
					t.Errorf("Unexpected content '%s' found in result", test.content)
				}
			}
		})
	}
}

// TestLayoutSectionsWithEmptyValues 测试空值布局区块
func TestLayoutSectionsWithEmptyValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "layout_sections_empty_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	viewsDir := filepath.Join(tmpDir, "views")
	layoutsDir := filepath.Join(viewsDir, "layouts")

	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 布局文件含有多个区块占位符
	layoutContent := `<html>
<head>{{.HtmlHead}}</head>
<body>
<nav>{{.Navigation}}</nav>
<div>{{.Content}}</div>
<aside>{{.SideBar}}</aside>
<footer>{{.Footer}}</footer>
{{.Scripts}}
</body>
</html>`

	layoutPath := filepath.Join(layoutsDir, "empty_sections_layout.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	templateContent := `<p>Main content here</p>`
	templatePath := filepath.Join(viewsDir, "empty_sections_test.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{viewsDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = false

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// 只提供部分布局区块，其他保持空白
	layoutSections := map[string]string{
		"Navigation": `<ul><li>Home</li></ul>`,
		"Footer":     `<p>Copyright 2024</p>`,
		// HtmlHead, SideBar, Scripts 未定义，应该被替换为空字符串
	}

	result, err := engine.RenderWithLayoutSections("empty_sections_test", "layouts/empty_sections_layout", map[string]any{}, layoutSections)
	if err != nil {
		t.Fatalf("Failed to render with partial layout sections: %v", err)
	}

	t.Logf("Rendered result with empty sections:\n%s", result)

	// 验证已定义的区块存在
	if !strings.Contains(result, "<ul><li>Home</li></ul>") {
		t.Error("Navigation section not found")
	}

	if !strings.Contains(result, "<p>Copyright 2024</p>") {
		t.Error("Footer section not found")
	}

	// 验证未定义的区块占位符被清空
	emptyPlaceholders := []string{"{{.HtmlHead}}", "{{.SideBar}}", "{{.Scripts}}"}
	for _, placeholder := range emptyPlaceholders {
		if strings.Contains(result, placeholder) {
			t.Errorf("Empty placeholder '%s' was not replaced", placeholder)
		}
	}
}

// TestLayoutSectionsHashing 测试布局区块哈希功能
func TestLayoutSectionsHashing(t *testing.T) {
	cfg := &config.TemplateConfig{}
	engine, _ := view.NewTemplateEngine(cfg)

	// 测试相同的布局区块生成相同的哈希
	sections1 := map[string]string{
		"HtmlHead": "<meta charset='utf-8'>",
		"Scripts":  "<script>console.log('test');</script>",
	}

	sections2 := map[string]string{
		"Scripts":  "<script>console.log('test');</script>",
		"HtmlHead": "<meta charset='utf-8'>", // 顺序不同
	}

	hash1 := engine.GenerateSectionsHash(sections1)
	hash2 := engine.GenerateSectionsHash(sections2)

	if hash1 != hash2 {
		t.Errorf("Expected same hash for same sections, got %s and %s", hash1, hash2)
	}

	// 测试不同的布局区块生成不同的哈希
	sections3 := map[string]string{
		"HtmlHead": "<meta charset='utf-8'>",
		"Scripts":  "<script>console.log('different');</script>", // 内容不同
	}

	hash3 := engine.GenerateSectionsHash(sections3)
	if hash1 == hash3 {
		t.Errorf("Expected different hash for different sections, got same hash: %s", hash1)
	}

	// 测试空区块
	emptyHash := engine.GenerateSectionsHash(nil)
	if emptyHash != "empty" {
		t.Errorf("Expected 'empty' hash for nil sections, got %s", emptyHash)
	}

	emptyHash2 := engine.GenerateSectionsHash(map[string]string{})
	if emptyHash2 != "empty" {
		t.Errorf("Expected 'empty' hash for empty sections, got %s", emptyHash2)
	}
}

// BenchmarkLayoutSections 布局区块性能测试
func BenchmarkLayoutSections(b *testing.B) {
	// 创建测试环境
	tmpDir, _ := os.MkdirTemp("", "layout_sections_bench")
	defer os.RemoveAll(tmpDir)

	viewsDir := filepath.Join(tmpDir, "views")
	layoutsDir := filepath.Join(viewsDir, "layouts")
	os.MkdirAll(layoutsDir, 0755)

	// 简单的布局文件
	layoutContent := `<html><head>{{.HtmlHead}}</head><body>{{.LayoutContent}}{{.Scripts}}</body></html>`
	layoutPath := filepath.Join(layoutsDir, "bench_layout.html")
	os.WriteFile(layoutPath, []byte(layoutContent), 0644)

	templateContent := `<p>{{.Message}}</p>`
	templatePath := filepath.Join(viewsDir, "bench_template.html")
	os.WriteFile(templatePath, []byte(templateContent), 0644)

	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{viewsDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = true

	engine, _ := view.NewTemplateEngine(cfg)

	layoutSections := map[string]string{
		"HtmlHead": `<meta charset="utf-8"><title>Benchmark</title>`,
		"Scripts":  `<script>console.log("benchmark");</script>`,
	}

	data := map[string]any{"Message": "Hello Benchmark"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := engine.RenderWithLayoutSections("bench_template", "layouts/bench_layout", data, layoutSections)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}
