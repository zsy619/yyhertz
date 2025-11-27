package main

import (
	"path/filepath"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func TestDebugTemplateName(t *testing.T) {
	t.Logf("🔧 调试模板名称解析问题")

	// 获取统一的模板引擎实例
	engine := view.GetUnifiedEngine()
	if engine == nil {
		t.Fatal("❌ 无法获取模板引擎实例")
	}

	// 测试模板加载
	templateName := "template/index.html"
	t.Logf("🔍 测试加载模板: %s", templateName)

	// 先测试FindTemplateFile
	templatePath, err := engine.FindTemplateFile(templateName)
	if err != nil {
		t.Logf("❌ 查找模板文件失败: %v", err)
		return
	}
	t.Logf("📁 找到模板路径: %s", templatePath)
	t.Logf("📄 文件basename: %s", filepath.Base(templatePath))

	tmpl, err := engine.GetTemplate(templateName)
	if err != nil {
		t.Logf("❌ 模板加载失败: %v", err)
	} else {
		t.Logf("✅ 模板加载成功")
		t.Logf("📋 模板名称: %s", tmpl.Name())

		// 显示模板的所有子模板
		templates := tmpl.Templates()
		t.Logf("📋 可用的模板数量: %d", len(templates))
		for i, tt := range templates {
			t.Logf("  %d. 模板名: %s", i+1, tt.Name())
		}
	}

	t.Log("🔧 调试完成")
}
