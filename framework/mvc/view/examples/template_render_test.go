package main

import (
	"log"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func TestTemplateRender(t *testing.T) {
	log.Println("🧪 测试模板渲染功能")

	// 获取统一的模板引擎实例
	engine := view.GetUnifiedEngine()
	if engine == nil {
		t.Fatal("❌ 无法获取模板引擎实例")
	}

	// 准备与TemplateController.Index()相同的测试数据
	testData := map[string]any{
		"Title": "模板管理 - YYHertz 演示",
		"Templates": []map[string]any{
			{
				"name":     "template/index.html",
				"path":     "/test/path.html",
				"size":     1024,
				"modified": "2024-08-29 21:00:00",
				"type":     "page",
			},
			{
				"name":     "layouts/app.html",
				"path":     "/layouts/app.html",
				"size":     2048,
				"modified": "2024-08-29 20:00:00",
				"type":     "layout",
			},
		},
		"Stats": map[string]any{
			"total":      2,
			"layouts":    1,
			"components": 0,
			"pages":      1,
		},
		"now": time.Now(),
	}

	t.Logf("🎯 开始测试模板渲染...")
	result, err := engine.Render("template/index.html", testData)
	if err != nil {
		t.Logf("❌ 模板渲染失败: %v", err)
		return
	}

	t.Logf("✅ 模板渲染成功！输出长度: %d 字符", len(result))
	if len(result) < 1000 {
		t.Logf("📄 渲染结果:\n%s", result)
	} else {
		t.Logf("📄 渲染结果前500字符:\n%s...", result[:500])
	}

	t.Log("🧪 测试完成")
}
