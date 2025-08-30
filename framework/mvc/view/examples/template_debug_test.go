package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func TestTemplateDebug(t *testing.T) {
	// 测试使用模板引擎加载模板
	templatePath := "/Volumes/E/JYW/YYHertz/framework/mvc/view/examples/views/template/index.html"

	// 检查文件是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		fmt.Printf("❌ 模板文件不存在: %s\n", templatePath)
		return
	}
	fmt.Printf("✅ 模板文件存在: %s\n", templatePath)

	// 使用我们的模板引擎来测试
	fmt.Printf("\n=== 使用模板引擎测试 ===\n")

	// 获取统一的模板引擎实例
	engine := view.GetUnifiedEngine()
	if engine == nil {
		fmt.Printf("❌ 无法获取模板引擎实例\n")
		return
	}
	fmt.Printf("✅ 获取到模板引擎实例\n")

	// 尝试加载模板
	tmpl, err := engine.LoadTemplate("template/index.html")
	if err != nil {
		fmt.Printf("❌ 加载模板失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 模板加载成功\n")

	// 检查模板是否为空
	if tmpl == nil {
		fmt.Printf("❌ 模板为空\n")
		return
	}

	// 尝试获取模板定义的名称
	tmplNames := tmpl.DefinedTemplates()
	fmt.Printf("✅ 模板定义的名称: %s\n", tmplNames)

	// 测试模板执行
	fmt.Printf("\n=== 测试模板执行 ===\n")
	testData := map[string]interface{}{
		"Templates": []map[string]interface{}{
			{"Name": "test.html", "Size": 1024, "ModTime": "2025-08-29"},
		},
		"Stats": map[string]interface{}{
			"total":      1,
			"layouts":    0,
			"components": 0,
			"pages":      1,
		},
		"now": "2025-08-29T15:30:00Z",
	}

	// 尝试执行模板
	err = tmpl.Execute(os.Stdout, testData)
	if err != nil {
		fmt.Printf("❌ 模板执行失败: %v\n", err)
		return
	}
	fmt.Printf("\n✅ 模板执行成功\n")
}
