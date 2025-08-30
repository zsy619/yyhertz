package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func TestEngineLoad(t *testing.T) {
	fmt.Printf("=== 测试模板引擎加载 template/index.html ===\n")

	// 获取统一引擎
	engine := view.GetUnifiedEngine()
	if engine == nil {
		fmt.Printf("❌ 无法获取模板引擎\n")
		t.Fatal("❌ 无法获取模板引擎")
	}

	fmt.Printf("✅ 获得模板引擎实例\n")

	// 检查当前工作目录
	cwd, _ := os.Getwd()
	fmt.Printf("工作目录: %s\n", cwd)

	// 检查目标模板文件是否存在
	templateFile := "views/template/index.html"
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		fmt.Printf("❌ 模板文件不存在: %s\n", templateFile)
		t.Fatal("❌ 模板文件不存在")
	}
	fmt.Printf("✅ 模板文件存在: %s\n", templateFile)

	// 尝试通过引擎的 Render 方法加载模板
	templateName := "template/index.html"
	fmt.Printf("尝试渲染模板: %s\n", templateName)

	// 准备测试数据
	testData := map[string]any{
		"Title": "测试标题",
		"Templates": []map[string]any{
			{"name": "test.html", "size": 100, "type": "page"},
		},
		"Stats": map[string]any{
			"total":      1,
			"pages":      1,
			"layouts":    0,
			"components": 0,
		},
	}

	// 尝试渲染
	result, err := engine.Render(templateName, testData)
	if err != nil {
		fmt.Printf("❌ 渲染失败: %v\n", err)

		// 尝试获取引擎的缓存统计来调试
		stats := engine.GetCacheStats()
		fmt.Printf("缓存统计: %v\n", stats)

		// 尝试直接加载模板文件进行诊断
		fmt.Printf("\n=== 尝试手动解析模板文件 ===\n")

		content, readErr := os.ReadFile(templateFile)
		if readErr != nil {
			fmt.Printf("❌ 读取文件失败: %v\n", readErr)
		} else {
			fmt.Printf("✅ 文件大小: %d bytes\n", len(content))
			fmt.Printf("前100个字符: %s...\n", string(content[:min(100, len(content))]))
		}

	} else {
		fmt.Printf("✅ 渲染成功! 结果长度: %d\n", len(result))
	}
}
