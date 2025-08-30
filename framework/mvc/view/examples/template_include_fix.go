package main

import (
	"fmt"
	"log"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func maixxxn() {
	fmt.Printf("=== 测试 TemplateIncludeEngine 修复效果 ===\n")

	// 创建 TemplateIncludeEngine
	cfg := view.DefaultTemplateConfig()
	engine, err := view.NewTemplateIncludeEngine(cfg)
	if err != nil {
		log.Fatalf("❌ 创建TemplateIncludeEngine失败: %v", err)
	}

	fmt.Printf("✅ TemplateIncludeEngine创建成功\n")

	// 获取可用模板列表
	availableTemplates := engine.ListAvailableTemplates()
	fmt.Printf("📋 可用模板数量: %d\n", len(availableTemplates))

	if len(availableTemplates) > 0 {
		fmt.Printf("前10个可用模板:\n")
		for i, tmpl := range availableTemplates {
			if i >= 10 {
				break
			}
			fmt.Printf("  %d. %s\n", i+1, tmpl)
		}
	}

	// 测试关键的模板查找
	testTemplateName := "template/index.html"
	fmt.Printf("\n=== 测试关键模板查找 ===\n")
	fmt.Printf("请求模板: %s\n", testTemplateName)

	// 准备测试数据
	testData := map[string]any{
		"Title": "测试",
		"Stats": map[string]any{
			"total":      32,
			"pages":      22,
			"layouts":    7,
			"components": 3,
		},
		"Templates": []map[string]any{
			{"name": "test", "size": 100},
		},
	}

	// 尝试渲染
	result, err := engine.RenderTemplate(testTemplateName, testData)
	if err != nil {
		fmt.Printf("❌ 渲染失败: %v\n", err)
	} else {
		fmt.Printf("✅ 渲染成功! 结果长度: %d\n", len(result))
	}
}
