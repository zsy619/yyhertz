package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestTemplateCacheKeyManager 测试缓存键管理器
func TestTemplateCacheKeyManager(t *testing.T) {
	ckm := view.DefaultCacheKeyManager

	// 测试生成模板键
	templateKey := ckm.GenerateTemplateKey("index.html")
	if templateKey != "index.html" {
		t.Errorf("Expected template key 'index.html', got '%s'", templateKey)
	}

	// 测试生成布局键
	layoutKey := ckm.GenerateLayoutKey("index.html", "app.html")
	if layoutKey != "index.html@app.html" {
		t.Errorf("Expected layout key 'index.html@app.html', got '%s'", layoutKey)
	}

	// 测试解析布局键
	templateName, layoutName := ckm.ParseLayoutKey(layoutKey)
	if templateName != "index.html" || layoutName != "app.html" {
		t.Errorf("Expected parsed template 'index.html' and layout 'app.html', got '%s' and '%s'", templateName, layoutName)
	}

	// 测试是否是布局键
	if !ckm.IsLayoutKey(layoutKey) {
		t.Errorf("Expected '%s' to be a layout key", layoutKey)
	}

	if ckm.IsLayoutKey(templateKey) {
		t.Errorf("Expected '%s' to NOT be a layout key", templateKey)
	}

	// 测试键验证
	if err := ckm.ValidateKey(""); err == nil {
		t.Error("Expected error for empty key")
	}

	if err := ckm.ValidateKey("key with spaces"); err == nil {
		t.Error("Expected error for key with spaces")
	}

	if err := ckm.ValidateKey("valid_key"); err != nil {
		t.Errorf("Expected no error for valid key, got: %v", err)
	}
}

// TestOptimizedTemplateWithLayout 测试优化的模板布局组合
func TestOptimizedTemplateWithLayout(t *testing.T) {
	// 创建临时目录和文件
	tempDir, err := os.MkdirTemp("", "template_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建布局文件
	layoutsDir := filepath.Join(tempDir, "layouts")
	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	layoutContent := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>{{template "content" .}}</body>
</html>`

	layoutFile := filepath.Join(layoutsDir, "test.html")
	if err := os.WriteFile(layoutFile, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	// 创建内容模板
	templateContent := `{{define "content"}}
<h1>{{.Message}}</h1>
{{end}}`

	templateFile := filepath.Join(tempDir, "test_content.html")
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
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
	defer engine.Close()

	// 测试缓存键生成和使用
	cacheKey := view.DefaultCacheKeyManager.GenerateLayoutKey("test_content.html", "test.html")
	t.Logf("Generated cache key: %s", cacheKey)

	// 测试模板渲染
	data := map[string]any{
		"Title":   "Test Page",
		"Message": "Hello, World!",
	}

	// 第一次渲染（应该从磁盘加载）
	result1, err := engine.RenderWithLayout("test_content.html", "test.html", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if !strings.Contains(result1, "Hello, World!") {
		t.Errorf("Expected result to contain 'Hello, World!', got: %s", result1)
	}

	if !strings.Contains(result1, "Test Page") {
		t.Errorf("Expected result to contain 'Test Page', got: %s", result1)
	}

	// 第二次渲染（应该从缓存加载）
	result2, err := engine.RenderWithLayout("test_content.html", "test.html", data)
	if err != nil {
		t.Fatalf("Failed to render template from cache: %v", err)
	}

	if result1 != result2 {
		t.Error("Expected same result when rendering from cache")
	}

	// 验证缓存中是否有正确的模板
	cachedTemplate, exists := engine.GetTemplate("test_content.html")

	if exists != nil {
		t.Errorf("Expected template to be cached with key: %s", cacheKey)
	}

	if cachedTemplate == nil {
		t.Error("Expected cached template to not be nil")
	}

	// 测试缓存统计
	stats := engine.GetEnhancedCacheStats()
	if stats.TotalCount == 0 {
		t.Error("Expected non-zero total count in cache stats")
	}

	if stats.KeyTypes["layout_keys"] == 0 {
		t.Error("Expected non-zero layout keys count")
	}

	t.Logf("Cache stats: %+v", stats)
}

// TestTemplatePreloader 测试模板预加载器
func TestTemplatePreloader(t *testing.T) {
	// 创建临时目录和文件
	tempDir, err := os.MkdirTemp("", "preloader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试模板文件
	templates := map[string]string{
		"index.html": `<h1>{{.Title}}</h1>`,
		"about.html": `<p>{{.Content}}</p>`,
		"error.html": `<div>{{.Error}}</div>`,
	}

	for name, content := range templates {
		templateFile := filepath.Join(tempDir, name)
		if err := os.WriteFile(templateFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template file %s: %v", name, err)
		}
	}

	// 创建模板引擎配置
	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{tempDir}
	cfg.Paths.LayoutPath = ""
	cfg.Paths.ComponentPath = ""
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"
	cfg.Cache.EnableCache = true
	cfg.Reload.Enabled = false
	cfg.Performance.EnableCompress = false
	cfg.Theme.Current = "default"
	cfg.Theme.Themes = make(map[string]*config.ThemeConfig)

	// 创建模板引擎和预加载器
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create template engine: %v", err)
	}
	defer engine.Close()

	preloader := view.NewTemplatePreloader(engine)

	// 添加模板到预加载列表
	preloader.AddPreloadTemplate("index.html")
	preloader.AddPreloadTemplate("about.html")
	preloader.AddPreloadTemplate("error.html")

	// 测试预加载统计（预加载前）
	stats := preloader.GetPreloadStats()
	if stats["preload_list_count"] != 3 {
		t.Errorf("Expected 3 templates in preload list, got %v", stats["preload_list_count"])
	}

	// 执行预加载
	if err := preloader.PreloadAll(); err != nil {
		t.Fatalf("Failed to preload templates: %v", err)
	}

	// 验证模板是否被加载到缓存中
	cacheCount := engine.GetTemplateCount()
	if cacheCount < 3 {
		t.Errorf("Expected at least 3 templates in cache, got %d", cacheCount)
	}

	// 测试渲染预加载的模板
	data := map[string]any{
		"Title":   "Preloaded Page",
		"Content": "This template was preloaded",
		"Error":   "No error",
	}

	for templateName := range templates {
		result, err := engine.Render(templateName, data)
		if err != nil {
			t.Errorf("Failed to render preloaded template %s: %v", templateName, err)
		}

		if result == "" {
			t.Errorf("Expected non-empty result for template %s", templateName)
		}

		t.Logf("Rendered %s: %s", templateName, result)
	}
}

// TestCachePerformance 测试缓存性能
func TestCachePerformance(t *testing.T) {
	// 创建临时目录和文件
	tempDir, err := os.MkdirTemp("", "performance_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建简单的模板文件
	templateContent := `<h1>{{.Title}}</h1><p>{{.Message}}</p>`
	templateFile := filepath.Join(tempDir, "performance.html")
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template file: %v", err)
	}

	// 创建模板引擎配置
	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{tempDir}
	cfg.Paths.LayoutPath = ""
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
	defer engine.Close()

	data := map[string]any{
		"Title":   "Performance Test",
		"Message": "Testing cache performance",
	}

	// 第一次渲染（冷缓存）
	start := time.Now()
	_, err = engine.Render("performance.html", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}
	coldTime := time.Since(start)

	// 多次渲染（热缓存）
	const iterations = 100
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err = engine.Render("performance.html", data)
		if err != nil {
			t.Fatalf("Failed to render template on iteration %d: %v", i, err)
		}
	}
	hotTime := time.Since(start)
	avgHotTime := hotTime / iterations

	t.Logf("Cold cache render time: %v", coldTime)
	t.Logf("Hot cache render time (avg of %d): %v", iterations, avgHotTime)
	t.Logf("Performance improvement: %.2fx", float64(coldTime)/float64(avgHotTime))

	// 验证缓存加速效果
	if avgHotTime > coldTime {
		t.Error("Expected hot cache to be faster than cold cache")
	}

	// 测试缓存性能分析
	analysis := engine.AnalyzeCachePerformance()
	t.Logf("Cache analysis: %+v", analysis)

	// 获取缓存统计
	stats := engine.GetEnhancedCacheStats()
	if stats.TotalCount == 0 {
		t.Error("Expected non-zero cache count")
	}

	if stats.MemoryEstimate == 0 {
		t.Error("Expected non-zero memory estimate")
	}
}

// TestCacheKeyConsistency 测试缓存键一致性
func TestCacheKeyConsistency(t *testing.T) {
	ckm := view.DefaultCacheKeyManager

	testCases := []struct {
		templateName string
		layoutName   string
		expectedKey  string
	}{
		{"index.html", "app.html", "index.html@app.html"},
		{"user/profile.html", "admin.html", "user/profile.html@admin.html"},
		{"simple.html", "", "simple.html"},
		{"test.html", "layout.html", "test.html@layout.html"},
	}

	for _, tc := range testCases {
		// 测试生成键
		generatedKey := ckm.GenerateLayoutKey(tc.templateName, tc.layoutName)
		if generatedKey != tc.expectedKey {
			t.Errorf("Expected key '%s', got '%s' for template '%s' and layout '%s'",
				tc.expectedKey, generatedKey, tc.templateName, tc.layoutName)
		}

		// 测试解析键
		if tc.layoutName != "" { // 只有布局键才需要解析
			parsedTemplate, parsedLayout := ckm.ParseLayoutKey(generatedKey)
			if parsedTemplate != tc.templateName || parsedLayout != tc.layoutName {
				t.Errorf("Key parsing failed for '%s': expected template '%s' and layout '%s', got '%s' and '%s'",
					generatedKey, tc.templateName, tc.layoutName, parsedTemplate, parsedLayout)
			}
		}

		// 测试键验证
		if err := ckm.ValidateKey(generatedKey); err != nil {
			t.Errorf("Key validation failed for '%s': %v", generatedKey, err)
		}
	}
}
