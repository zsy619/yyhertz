package main

import (
	"os"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func TestTemplateContent(t *testing.T) {
	t.Log("🔧 调试模板内容和执行问题")

	// 获取统一的模板引擎实例
	engine := view.GetUnifiedEngine()
	if engine == nil {
		t.Fatal("❌ 无法获取模板引擎实例")
	}

	templateName := "template/index.html"
	t.Logf("🔍 调试模板: %s", templateName)

	// 1. 检查模板文件内容
	templatePath, err := engine.FindTemplateFile(templateName)
	if err != nil {
		t.Logf("❌ 查找模板文件失败: %v", err)
		return
	}
	t.Logf("📁 模板文件路径: %s", templatePath)

	// 读取模板内容
	content, err := os.ReadFile(templatePath)
	if err != nil {
		t.Logf("❌ 读取模板内容失败: %v", err)
		return
	}
	t.Logf("📄 模板内容长度: %d 字节", len(content))
	t.Logf("📄 模板内容前200字符:\n%s", string(content[:min(len(content), 200)]))

	// 2. 测试模板加载
	tmpl, err := engine.GetTemplate(templateName)
	if err != nil {
		t.Logf("❌ 模板加载失败: %v", err)
		return
	}
	t.Logf("✅ 模板加载成功, 模板名称: %s", tmpl.Name())

	// 3. 检查模板结构
	templates := tmpl.Templates()
	t.Logf("📋 可用的子模板数量: %d", len(templates))
	for i, tt := range templates {
		hasValidTree := tt.Tree != nil && tt.Tree.Root != nil
		t.Logf("  %d. 子模板名: %s, 有效树: %v", i+1, tt.Name(), hasValidTree)
	}

	// 4. 测试简单数据渲染
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
		"now": "2024-08-29 21:00:00", // 简化时间数据
	}

	t.Logf("🎯 开始测试模板渲染...")
	result, err := engine.Render(templateName, testData)
	if err != nil {
		t.Logf("❌ 模板渲染失败: %v", err)

		// 检查是否是特定函数缺失导致的问题
		if containsString(err.Error(), "formatFileSize") {
			t.Logf("🔧 检测到formatFileSize函数问题，尝试修复数据...")
			// 移除可能导致问题的formatFileSize调用
			for _, template := range testData["Templates"].([]map[string]interface{}) {
				template["size"] = "1.0 KB" // 直接提供格式化后的字符串
			}

			result, err = engine.Render(templateName, testData)
			if err != nil {
				t.Logf("❌ 修复尝试失败: %v", err)
			} else {
				t.Logf("✅ 修复成功！渲染结果长度: %d 字符", len(result))
			}
		}
	} else {
		t.Logf("✅ 模板渲染成功！输出长度: %d 字符", len(result))
		if len(result) < 500 {
			t.Logf("📄 渲染结果:\n%s", result)
		} else {
			t.Logf("📄 渲染结果前500字符:\n%s", result[:500])
		}
	}

	t.Log("🔧 调试完成")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (len(needle) == 0 || haystack[:len(needle)] == needle ||
		haystack[len(haystack)-len(needle):] == needle ||
		findInString(haystack, needle))
}

func findInString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
