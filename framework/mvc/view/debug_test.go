package view

import (
	"bytes"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
)

// TestDebugTemplate 调试模板解析
func TestDebugTemplate(t *testing.T) {
	templatePath := "../../views/Login/Login.html"

	if err := DebugTemplate(templatePath); err != nil {
		t.Fatalf("调试模板失败: %v", err)
	}
}

// TestTemplateEngineDebug 调试模板引擎加载
func TestTemplateEngineDebug(t *testing.T) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"../../views"}
	cfg.Cache.EnableCache = true // 启用缓存来测试修复效果

	engine, err := NewTemplateEngine(cfg)
	if err != nil {
		t.Fatalf("创建模板引擎失败: %v", err)
	}

	// 加载模板
	tmpl, err := engine.loadTemplate("Login/Login")
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	t.Logf("返回的模板名称: %s", tmpl.Name())
	t.Logf("模板树是否为空: %t", tmpl.Tree == nil)
	if tmpl.Tree != nil {
		t.Logf("模板树根是否为空: %t", tmpl.Tree.Root == nil)
	}

	// 直接执行返回的模板
	testData := map[string]interface{}{
		"Data": "测试数据",
	}

	// 先测试直接执行模板
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, testData)
	if err != nil {
		t.Logf("直接执行模板失败: %v", err)
	} else {
		t.Logf("直接执行模板成功，长度: %d", buf.Len())
	}

	// 再通过引擎渲染
	html, err := engine.RenderWithLayout("Login/Login", "", testData)
	if err != nil {
		t.Errorf("引擎渲染失败: %v", err)
	} else {
		t.Logf("引擎渲染成功，长度: %d", len(html))
	}
}
