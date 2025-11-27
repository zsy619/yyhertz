package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestLayoutSectionsIntegration 完整集成测试
func TestLayoutSectionsIntegration(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "layout_sections_integration")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	viewsDir := filepath.Join(tmpDir, "views")
	layoutsDir := filepath.Join(viewsDir, "layouts")

	if err := os.MkdirAll(layoutsDir, 0755); err != nil {
		t.Fatalf("Failed to create layouts dir: %v", err)
	}

	// 创建完整的布局文件
	layoutContent := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    {{.HtmlHead}}
</head>
<body>
    <header>{{.Header}}</header>
    <nav>{{.Navigation}}</nav>
    <div class="container">
        <aside class="sidebar">{{.SideBar}}</aside>
        <main class="content">{{.LayoutContent}}</main>
    </div>
    <footer>{{.Footer}}</footer>
    {{.Scripts}}
</body>
</html>`

	layoutPath := filepath.Join(layoutsDir, "complete_layout.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Failed to write layout file: %v", err)
	}

	// 创建内容模板
	templateContent := `<article>
    <h1>{{.Title}}</h1>
    <p>作者: {{.Author}}</p>
    <div class="content">{{.Content}}</div>
    <time>发布时间: {{.PublishTime}}</time>
</article>`

	templatePath := filepath.Join(viewsDir, "article.html")
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
		"Title":       "技术文章",
		"Author":      "张开发者",
		"Content":     "这是一篇关于 LayoutSections 的技术文章内容...",
		"PublishTime": "2024-08-28 21:15:00",
	}

	// 定义完整的布局区块
	layoutSections := map[string]string{
		"HtmlHead": `
    <!-- SEO Meta -->
    <meta name="description" content="LayoutSections 技术文章演示">
    <meta name="keywords" content="YYHertz,模板引擎,Go,布局区块">
    <meta name="author" content="张开发者">
    
    <!-- CSS -->
    <style>
        body { font-family: "Microsoft YaHei", Arial, sans-serif; margin: 0; padding: 0; }
        .container { display: flex; max-width: 1200px; margin: 0 auto; padding: 20px; }
        .sidebar { width: 250px; background: #f8f9fa; padding: 20px; margin-right: 20px; }
        .content { flex: 1; background: #fff; padding: 20px; }
        header, footer { background: #343a40; color: white; padding: 15px 0; text-align: center; }
        nav { background: #6c757d; padding: 10px 0; }
        .nav-item { display: inline-block; margin: 0 15px; color: white; }
    </style>`,

		"Header": `
    <div>
        <h1>🌟 YYHertz 技术博客</h1>
        <p>专业的 Go 模板引擎解决方案</p>
    </div>`,

		"Navigation": `
    <div style="text-align: center;">
        <span class="nav-item">🏠 首页</span>
        <span class="nav-item">📚 文章</span>
        <span class="nav-item">🔧 工具</span>
        <span class="nav-item">📞 联系</span>
    </div>`,

		"SideBar": `
    <div class="sidebar-content">
        <h3>📖 文章目录</h3>
        <ul style="list-style: none; padding: 0;">
            <li style="padding: 5px 0;">• LayoutSections 介绍</li>
            <li style="padding: 5px 0;">• 使用方法</li>
            <li style="padding: 5px 0;">• 高级功能</li>
            <li style="padding: 5px 0;">• 性能优化</li>
        </ul>
        
        <h3>🏷️ 标签</h3>
        <div>
            <span style="background: #007bff; color: white; padding: 2px 8px; margin: 2px; display: inline-block; border-radius: 3px;">Go</span>
            <span style="background: #28a745; color: white; padding: 2px 8px; margin: 2px; display: inline-block; border-radius: 3px;">模板引擎</span>
            <span style="background: #ffc107; color: black; padding: 2px 8px; margin: 2px; display: inline-block; border-radius: 3px;">YYHertz</span>
        </div>
    </div>`,

		"Footer": `
    <div>
        <p>&copy; 2024 YYHertz 模板引擎 | LayoutSections 演示</p>
        <p><small>高性能 • 易使用 • 功能丰富</small></p>
    </div>`,

		"Scripts": `
    <script>
        // 页面统计
        console.log('📊 页面加载统计');
        console.log('- 模板引擎: YYHertz');
        console.log('- 渲染方式: LayoutSections');
        console.log('- 页面类型: 技术文章');
        
        // 添加交互功能
        document.addEventListener('DOMContentLoaded', function() {
            console.log('✅ 页面 DOM 加载完成');
            
            // 简单的页面访问统计
            let visitCount = localStorage.getItem('visitCount') || 0;
            visitCount++;
            localStorage.setItem('visitCount', visitCount);
            console.log('🔢 访问次数:', visitCount);
        });
        
        // 性能监控
        window.addEventListener('load', function() {
            const loadTime = performance.now();
            console.log('⚡ 页面加载时间:', Math.round(loadTime) + 'ms');
        });
    </script>`,
	}

	// 渲染测试
	result, err := engine.RenderWithLayoutSections("article", "layouts/complete_layout", testData, layoutSections)
	if err != nil {
		t.Fatalf("Failed to render with layout sections: %v", err)
	}

	t.Logf("完整渲染结果:\n%s", result)

	// 验证各个区块是否正确嵌入
	validations := []struct {
		name     string
		content  string
		expected bool
		desc     string
	}{
		// HTML Head 区块验证
		{"SEO Meta", `<meta name="description" content="LayoutSections 技术文章演示">`, true, "SEO元信息"},
		{"CSS Styles", `font-family: "Microsoft YaHei"`, true, "CSS样式"},

		// Header 区块验证
		{"Header Title", `🌟 YYHertz 技术博客`, true, "页面标题"},

		// Navigation 区块验证
		{"Navigation", `🏠 首页`, true, "导航菜单"},

		// SideBar 区块验证
		{"Sidebar Title", `📖 文章目录`, true, "侧边栏标题"},
		{"Sidebar Tags", `<span style="background: #007bff`, true, "侧边栏标签"},

		// Footer 区块验证
		{"Footer Copyright", `&copy; 2024 YYHertz 模板引擎`, true, "版权信息"},

		// Scripts 区块验证
		{"Scripts Console", `console.log('📊 页面加载统计')`, true, "JavaScript代码"},
		{"Performance Monitor", `performance.now()`, true, "性能监控"},

		// 内容模板验证
		{"Article Title", `<h1>技术文章</h1>`, true, "文章标题"},
		{"Author Info", `作者: 张开发者`, true, "作者信息"},
		{"Publish Time", `发布时间: 2024-08-28 21:15:00`, true, "发布时间"},

		// 验证占位符已被替换
		{"No HtmlHead Placeholder", `{{.HtmlHead}}`, false, "HtmlHead占位符已替换"},
		{"No SideBar Placeholder", `{{.SideBar}}`, false, "SideBar占位符已替换"},
		{"No Scripts Placeholder", `{{.Scripts}}`, false, "Scripts占位符已替换"},
		{"No Content Placeholder", `{{.LayoutContent}}`, false, "LayoutContent占位符已替换"},
	}

	for _, test := range validations {
		t.Run(test.name, func(t *testing.T) {
			contains := strings.Contains(result, test.content)
			if contains != test.expected {
				if test.expected {
					t.Errorf("❌ %s: 期望的内容 '%s' 未找到", test.desc, test.content)
				} else {
					t.Errorf("❌ %s: 不期望的内容 '%s' 仍然存在", test.desc, test.content)
				}
			} else {
				t.Logf("✅ %s: 验证通过", test.desc)
			}
		})
	}
}

// TestControllerStyleUsage 测试控制器风格的使用方式
func TestControllerStyleUsage(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "controller_style_test")
	defer os.RemoveAll(tmpDir)

	viewsDir := filepath.Join(tmpDir, "views")
	layoutsDir := filepath.Join(viewsDir, "layouts")
	os.MkdirAll(layoutsDir, 0755)

	// 简化的布局文件
	layoutContent := `<html>
<head>{{.HtmlHead}}</head>
<body>
<nav>{{.Navigation}}</nav>
<main>{{.LayoutContent}}</main>
{{.Scripts}}
</body>
</html>`

	layoutPath := filepath.Join(layoutsDir, "controller_layout.html")
	os.WriteFile(layoutPath, []byte(layoutContent), 0644)

	templateContent := `<div>Controller 演示: {{.Message}}</div>`
	templatePath := filepath.Join(viewsDir, "controller_demo.html")
	os.WriteFile(templatePath, []byte(templateContent), 0644)

	cfg := &config.TemplateConfig{}
	cfg.Paths.ViewPaths = []string{viewsDir}
	cfg.Paths.LayoutPath = layoutsDir
	cfg.Paths.Extension = ".html"
	cfg.Syntax.DelimLeft = "{{"
	cfg.Syntax.DelimRight = "}}"

	engine, _ := view.NewTemplateEngine(cfg)

	// 模拟控制器中的使用方式
	data := map[string]any{
		"Message": "Hello from Controller!",
	}

	// Controller 中定义 LayoutSections
	layoutSections := map[string]string{
		"HtmlHead":   `<title>Controller Demo</title><meta charset="utf-8">`,
		"Navigation": `<ul><li>Home</li><li>About</li></ul>`,
		"Scripts":    `<script>console.log('Controller demo loaded');</script>`,
	}

	result, err := engine.RenderWithLayoutSections("controller_demo", "layouts/controller_layout", data, layoutSections)
	if err != nil {
		t.Fatalf("Controller style render failed: %v", err)
	}

	// 验证 Controller 风格的渲染结果
	if !strings.Contains(result, "Controller Demo") {
		t.Error("Title from HtmlHead section not found")
	}

	if !strings.Contains(result, "<ul><li>Home</li>") {
		t.Error("Navigation section not found")
	}

	if !strings.Contains(result, "Hello from Controller!") {
		t.Error("Main content not found")
	}

	if !strings.Contains(result, "console.log('Controller demo loaded')") {
		t.Error("Scripts section not found")
	}

	t.Log("✅ Controller 风格的 LayoutSections 使用验证通过")
}
