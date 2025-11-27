package main

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func TestDeepDebug(t *testing.T) {
	t.Log("🔧 深度调试模板解析和执行")

	// 获取统一的模板引擎实例
	engine := view.GetUnifiedEngine()
	if engine == nil {
		t.Fatal("❌ 无法获取模板引擎实例")
	}

	templateName := "template/index.html"
	t.Logf("🔍 深度分析模板: %s", templateName)

	// 1. 直接读取和分析模板文件
	templatePath, err := engine.FindTemplateFile(templateName)
	if err != nil {
		t.Logf("❌ 查找模板文件失败: %v", err)
		return
	}
	t.Logf("📁 模板文件路径: %s", templatePath)

	// 读取原始模板内容
	content, err := os.ReadFile(templatePath)
	if err != nil {
		t.Logf("❌ 读取模板内容失败: %v", err)
		return
	}

	templateContent := string(content)
	t.Logf("📄 模板内容长度: %d 字节", len(templateContent))
	t.Logf("📄 模板是否包含formatFileSize: %v", strings.Contains(templateContent, "formatFileSize"))
	t.Logf("📄 模板是否包含dateformat: %v", strings.Contains(templateContent, "dateformat"))

	// 2. 手动创建和解析模板，模拟LoadTemplate的过程
	t.Log("\n🔧 手动模拟LoadTemplate过程:")

	// 获取管理器和函数
	manager := view.GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(make(template.FuncMap))
	t.Logf("📋 可用函数数量: %d", len(mergedFuncs))

	// 检查关键函数
	criticalFuncs := []string{"formatFileSize", "dateformat", "now"}
	for _, funcName := range criticalFuncs {
		if _, exists := mergedFuncs[funcName]; exists {
			t.Logf("✅ 关键函数 '%s' 可用", funcName)
		} else {
			t.Logf("❌ 关键函数 '%s' 缺失", funcName)
		}
	}

	// 3. 按照LoadTemplate中的逻辑创建模板
	templateBaseName := filepath.Base(templatePath)
	t.Logf("🎯 模板基础名称: %s", templateBaseName)

	// 创建新模板实例
	tmpl := template.New(templateBaseName).
		Delims("{{", "}}").
		Funcs(mergedFuncs)

	t.Logf("📝 创建的模板名称: %s", tmpl.Name())

	// 解析文件
	parsedTmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		t.Logf("❌ 解析模板文件失败: %v", err)
		return
	}

	t.Logf("✅ 模板解析成功")

	// 4. 检查解析后的模板结构
	templates := parsedTmpl.Templates()
	t.Logf("📋 解析后的子模板数量: %d", len(templates))

	var executableTemplate *template.Template
	for i, tt := range templates {
		hasValidTree := tt.Tree != nil && tt.Tree.Root != nil
		t.Logf("  %d. 子模板: %s, 有效树: %v", i+1, tt.Name(), hasValidTree)

		// 寻找可执行的主模板
		if tt.Name() == templateBaseName && hasValidTree {
			executableTemplate = tt
			t.Logf("  ✅ 找到可执行的主模板: %s", tt.Name())
		}
	}

	if executableTemplate == nil {
		t.Logf("❌ 未找到可执行的主模板")
		// 尝试使用第一个有效的模板
		for _, tt := range templates {
			if tt.Tree != nil && tt.Tree.Root != nil {
				executableTemplate = tt
				t.Logf("📝 使用备选模板: %s", tt.Name())
				break
			}
		}
	}

	if executableTemplate == nil {
		t.Logf("❌ 没有任何可执行的模板")
		return
	}

	// 5. 测试模板执行
	t.Log("\n🎯 测试模板执行:")

	// 简化测试数据，避免复杂的函数调用
	testData := map[string]interface{}{
		"Title": "简单测试",
		"Stats": map[string]interface{}{
			"total":      10,
			"layouts":    2,
			"components": 1,
			"pages":      7,
		},
		"Templates": []map[string]interface{}{
			{
				"name":     "test.html",
				"path":     "/test/path.html",
				"size":     1024, // 数字，避免formatFileSize问题
				"modified": "2024-08-29 21:00:00",
				"type":     "page",
			},
		},
	}

	// 先检查模板内容，看是否需要特殊处理
	if strings.Contains(templateContent, "formatFileSize .size") {
		t.Logf("🔧 检测到formatFileSize调用，修改测试数据")
		// 替换为已格式化的字符串
		for _, template := range testData["Templates"].([]map[string]interface{}) {
			template["size"] = "1.0 KB"
		}
	}

	// 添加时间函数需要的数据
	if strings.Contains(templateContent, "dateformat now") {
		t.Logf("🔧 检测到dateformat now调用，添加时间数据")
		testData["now"] = "2024-08-29 21:00:00"
	}

	var buf strings.Builder
	err = executableTemplate.Execute(&buf, testData)
	if err != nil {
		t.Logf("❌ 模板执行失败: %v", err)

		// 如果是函数相关错误，尝试更简化的数据
		if strings.Contains(err.Error(), "function") {
			t.Logf("🔧 尝试使用最简化数据...")
			simpleData := map[string]interface{}{
				"Title": "最简测试",
			}

			var simpleBuf strings.Builder
			err2 := executableTemplate.Execute(&simpleBuf, simpleData)
			if err2 != nil {
				t.Logf("❌ 简化数据测试也失败: %v", err2)
			} else {
				result := simpleBuf.String()
				t.Logf("✅ 简化数据测试成功，输出长度: %d", len(result))
				if len(result) > 0 && len(result) < 200 {
					t.Logf("📄 简化输出: %s", result)
				}
			}
		}
	} else {
		result := buf.String()
		t.Logf("✅ 模板执行成功！输出长度: %d", len(result))
		if len(result) < 500 {
			t.Logf("📄 完整输出:\n%s", result)
		} else {
			t.Logf("📄 输出前200字符:\n%s", result[:200])
		}
	}

	t.Log("\n🔧 深度调试完成")
}
