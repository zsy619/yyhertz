package main

import (
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func TestCheckDependencies(t *testing.T) {
	t.Log("🔧 调试模板名称精确匹配问题")

	// 获取统一的模板引擎实例
	engine := view.GetUnifiedEngine()
	if engine == nil {
		t.Fatal("❌ 无法获取模板引擎实例")
	}

	// 测试不同的模板名称查找方式
	testNames := []string{
		"template/index.html", // 完整路径+扩展名（TemplateController使用的）
		"template/index",      // 路径+无扩展名（调试脚本看到的）
		"index.html",          // 只有文件名+扩展名
		"index",               // 只有文件名无扩展名
	}

	t.Log("🔍 测试不同模板名称的查找结果：")

	for _, templateName := range testNames {
		t.Logf("\n--- 测试模板名称: %s ---", templateName)

		// 1. 测试 FindTemplateFile
		templatePath, err := engine.FindTemplateFile(templateName)
		if err != nil {
			t.Logf("❌ FindTemplateFile失败: %v\n", err)
		} else {
			t.Logf("✅ FindTemplateFile成功: %s\n", templatePath)
		}

		// 2. 测试 GetTemplate
		tmpl, err := engine.GetTemplate(templateName)
		if err != nil {
			t.Logf("❌ GetTemplate失败: %v\n", err)
		} else {
			t.Logf("✅ GetTemplate成功, 模板名称: %s\n", tmpl.Name())

			// 3. 测试模板执行
			testData := map[string]interface{}{
				"Title": "测试标题",
				"Stats": map[string]interface{}{
					"total":      32,
					"layouts":    7,
					"components": 3,
					"pages":      22,
				},
				"Templates": []map[string]interface{}{
					{
						"name":     "test.html",
						"path":     "/test/path",
						"size":     1024,
						"modified": "2024-08-29 21:00:00",
						"type":     "page",
					},
				},
			}

			result, execErr := engine.Render(templateName, testData)
			if execErr != nil {
				t.Logf("❌ 模板执行失败: %v\n", execErr)
			} else {
				t.Logf("✅ 模板执行成功，输出长度: %d chars\n", len(result))
			}
		}
	}

	// 4. 获取所有已缓存的模板名称列表
	t.Log("\n📋 引擎中所有缓存的模板名称:")
	templateList := engine.GetTemplateList()
	for i, name := range templateList {
		t.Logf("  %d. %s\n", i+1, name)
	}

	t.Log("🔧 调试完成")
}
