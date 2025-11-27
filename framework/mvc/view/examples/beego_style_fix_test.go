package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestBeegoStyleTemplateWithLayout 测试基于Beego机制的模板与布局加载
func TestBeegoStyleTemplateWithLayout(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "beego_template_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建layouts目录
	layoutsDir := filepath.Join(tempDir, "layouts")
	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 创建测试布局文件
	layoutContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
{{template "content" .}}
</body>
</html>`
	layoutPath := filepath.Join(layoutsDir, "main.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	// 创建测试模板文件
	templateContent := `{{define "content"}}
<h1>Hello {{.Name}}!</h1>
<p>This is a test template with layout.</p>
{{end}}`
	templatePath := filepath.Join(tempDir, "test.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// 创建模板引擎配置
	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{tempDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.ComponentPath = ""
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = true
	cfg.Reload.Enabled = false
	cfg.Performance.EnableCompress = false
	cfg.Theme.Current = "default"
	cfg.Theme.Themes = make(map[string]*config.ThemeConfig)

	// 创建模板引擎
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// 布局已在创建引擎时自动加载
	// 验证布局是否成功加载
	layoutList := engine.GetLayoutList()
	if len(layoutList) == 0 {
		t.Fatalf("No layouts loaded, available layouts: %v", layoutList)
	}

	// 测试数据
	data := map[string]interface{}{
		"Title": "Test Page",
		"Name":  "Beego Style Template",
	}

	// 第一次渲染（应该创建并缓存模板）
	// 注意：布局名称应该包含目录路径
	result1, err := engine.RenderWithLayout("test", "layouts/main", data)
	if err != nil {
		t.Fatalf("First render failed: %v", err)
	}

	// 验证渲染结果
	expectedContent := []string{
		"<title>Test Page</title>",
		"<h1>Hello Beego Style Template!</h1>",
		"<p>This is a test template with layout.</p>",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(result1, expected) {
			t.Errorf("First render missing expected content: %s\nActual result:\n%s", expected, result1)
		}
	}

	// 第二次渲染（应该从缓存加载）
	result2, err := engine.RenderWithLayout("test", "layouts/main", data)
	if err != nil {
		t.Fatalf("Second render failed: %v", err)
	}

	// 验证两次渲染结果一致
	if result1 != result2 {
		t.Errorf("Cache consistency failed:\nFirst: %s\nSecond: %s", result1, result2)
	}

	// 验证缓存键是否正确匹配
	stats := engine.GetCacheStats()
	if templateCount, ok := stats["templates"].(int); !ok || templateCount == 0 {
		t.Errorf("Template should be cached, but cache stats show: %v", stats)
	}

	t.Logf("✅ Beego-style template rendering test passed!")
	t.Logf("Cache stats: %v", stats)
	t.Logf("First render result length: %d", len(result1))
	t.Logf("Second render result length: %d", len(result2))
}

// TestCacheKeyMatching 测试缓存键匹配机制
func TestCacheKeyMatching(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "cache_key_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建layouts目录
	layoutsDir := filepath.Join(tempDir, "layouts")
	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 创建测试布局文件
	layoutContent := `<html><body>{{template "content" .}}</body></html>`
	layoutPath := filepath.Join(layoutsDir, "simple.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	// 创建测试模板文件
	templateContent := `{{define "content"}}<h1>{{.Message}}</h1>{{end}}`
	templatePath := filepath.Join(tempDir, "message.html")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// 创建模板引擎配置
	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{tempDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = true
	cfg.Reload.Enabled = false

	// 创建模板引擎
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// 布局已在创建引擎时自动加载
	// 验证布局是否成功加载
	layoutList := engine.GetLayoutList()
	if len(layoutList) == 0 {
		t.Fatalf("No layouts loaded, available layouts: %v", layoutList)
	}

	// 测试数据
	data := map[string]interface{}{
		"Message": "Cache Key Test",
	}

	// 第一次渲染
	result1, err := engine.RenderWithLayout("message", "layouts/simple", data)
	if err != nil {
		t.Fatalf("First render failed: %v", err)
	}

	// 验证模板被正确缓存
	template, err := engine.GetTemplateWithLayout("message", "layouts/simple")
	if err != nil {
		t.Fatalf("Failed to get cached template: %v", err)
	}

	if template == nil {
		t.Fatalf("Template should not be nil")
	}

	// 验证模板名称与缓存键匹配（斜杠被替换为下划线）
	expectedTemplateName := "message_with_layouts_simple"
	if template.Name() != expectedTemplateName {
		t.Errorf("Template name mismatch. Expected: %s, Got: %s", expectedTemplateName, template.Name())
	}

	// 第二次渲染（应该命中缓存）
	result2, err := engine.RenderWithLayout("message", "layouts/simple", data)
	if err != nil {
		t.Fatalf("Second render failed: %v", err)
	}

	// 验证结果一致性
	if result1 != result2 {
		t.Errorf("Results should be identical:\nFirst: %s\nSecond: %s", result1, result2)
	}

	// 验证内容正确性
	if !strings.Contains(result1, "<h1>Cache Key Test</h1>") {
		t.Errorf("Result missing expected content: %s", result1)
	}

	t.Logf("✅ Cache key matching test passed!")
	t.Logf("Template name: %s", template.Name())
	t.Logf("Result: %s", result1)
}

// TestMultipleTemplatesWithSameLayout 测试多个模板使用同一布局的缓存隔离
func TestMultipleTemplatesWithSameLayout(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "multi_template_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建layouts目录
	layoutsDir := filepath.Join(tempDir, "layouts")
	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 创建共同布局
	layoutContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
<header>Common Header</header>
<main>{{template "content" .}}</main>
<footer>Common Footer</footer>
</body>
</html>`
	layoutPath := filepath.Join(layoutsDir, "common.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	// 创建第一个模板
	template1Content := `{{define "content"}}
<h1>Page 1: {{.Content}}</h1>
<p>This is the first page.</p>
{{end}}`
	template1Path := filepath.Join(tempDir, "page1.html")
	if err := os.WriteFile(template1Path, []byte(template1Content), 0644); err != nil {
		t.Fatalf("Failed to write template1 file: %v", err)
	}

	// 创建第二个模板
	template2Content := `{{define "content"}}
<h2>Page 2: {{.Content}}</h2>
<div>This is the second page with different content.</div>
{{end}}`
	template2Path := filepath.Join(tempDir, "page2.html")
	if err := os.WriteFile(template2Path, []byte(template2Content), 0644); err != nil {
		t.Fatalf("Failed to write template2 file: %v", err)
	}

	// 创建模板引擎
	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{tempDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = true

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}

	// 布局已在创建引擎时自动加载
	// 验证布局是否成功加载
	layoutList := engine.GetLayoutList()
	if len(layoutList) == 0 {
		t.Fatalf("No layouts loaded, available layouts: %v", layoutList)
	}

	// 渲染第一个模板
	data1 := map[string]interface{}{
		"Title":   "First Page",
		"Content": "Hello World",
	}
	result1, err := engine.RenderWithLayout("page1", "layouts/common", data1)
	if err != nil {
		t.Fatalf("Failed to render page1: %v", err)
	}

	// 渲染第二个模板
	data2 := map[string]interface{}{
		"Title":   "Second Page",
		"Content": "Goodbye World",
	}
	result2, err := engine.RenderWithLayout("page2", "layouts/common", data2)
	if err != nil {
		t.Fatalf("Failed to render page2: %v", err)
	}

	// 验证两个模板的结果是不同的
	if result1 == result2 {
		t.Errorf("Results should be different for different templates")
	}

	// 验证第一个模板的特定内容
	if !strings.Contains(result1, "<h1>Page 1: Hello World</h1>") {
		t.Errorf("Page1 missing expected h1 content: %s", result1)
	}
	if !strings.Contains(result1, "This is the first page") {
		t.Errorf("Page1 missing expected paragraph content: %s", result1)
	}

	// 验证第二个模板的特定内容
	if !strings.Contains(result2, "<h2>Page 2: Goodbye World</h2>") {
		t.Errorf("Page2 missing expected h2 content: %s", result2)
	}
	if !strings.Contains(result2, "This is the second page with different content") {
		t.Errorf("Page2 missing expected div content: %s", result2)
	}

	// 验证两个模板都包含共同的布局内容
	commonElements := []string{
		"<header>Common Header</header>",
		"<footer>Common Footer</footer>",
		"<title>First Page</title>",  // result1
		"<title>Second Page</title>", // result2
	}

	if !strings.Contains(result1, commonElements[0]) || !strings.Contains(result1, commonElements[1]) {
		t.Errorf("Page1 missing common layout elements: %s", result1)
	}
	if !strings.Contains(result2, commonElements[0]) || !strings.Contains(result2, commonElements[1]) {
		t.Errorf("Page2 missing common layout elements: %s", result2)
	}

	// 验证缓存统计
	stats := engine.GetCacheStats()
	if templateCount, ok := stats["templates"].(int); !ok || templateCount < 2 {
		t.Errorf("Should have at least 2 cached templates, got: %v", stats)
	}

	t.Logf("✅ Multiple templates with same layout test passed!")
	t.Logf("Cache stats: %v", stats)
	t.Logf("Page1 result length: %d", len(result1))
	t.Logf("Page2 result length: %d", len(result2))
}
