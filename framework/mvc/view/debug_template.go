package view

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
)

// DebugTemplate 调试模板解析问题
func DebugTemplate(templatePath string) error {
	fmt.Printf("=== 调试模板: %s ===\n", templatePath)
	
	// 检查文件是否存在
	info, err := os.Stat(templatePath)
	if err != nil {
		fmt.Printf("文件状态错误: %v\n", err)
		return err
	}
	fmt.Printf("文件大小: %d bytes\n", info.Size())
	
	// 读取文件内容
	content, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("读取文件错误: %v\n", err)
		return err
	}
	fmt.Printf("文件内容长度: %d\n", len(content))
	fmt.Printf("文件内容预览 (前100字符):\n%s\n", string(content[:min(100, len(content))]))
	
	// 尝试解析模板
	tmpl := template.New("test-template")
	parsedTmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		fmt.Printf("解析模板错误: %v\n", err)
		return err
	}
	
	fmt.Printf("解析成功!\n")
	fmt.Printf("模板名称: %s\n", parsedTmpl.Name())
	
	// 检查模板树
	if parsedTmpl.Tree == nil {
		fmt.Printf("⚠️  模板树为空\n")
	} else if parsedTmpl.Tree.Root == nil {
		fmt.Printf("⚠️  模板树根节点为空\n")
	} else {
		fmt.Printf("✅ 模板树正常\n")
	}
	
	// 尝试查找所有关联的模板
	templates := parsedTmpl.Templates()
	fmt.Printf("关联模板数量: %d\n", len(templates))
	for i, t := range templates {
		fmt.Printf("  %d: %s\n", i, t.Name())
	}
	
	// 尝试执行模板
	testData := map[string]interface{}{
		"Data": "测试数据",
	}
	
	var buf bytes.Buffer
	if err := parsedTmpl.Execute(&buf, testData); err != nil {
		fmt.Printf("❌ 执行模板错误: %v\n", err)
		
		// 尝试通过名称执行
		for _, t := range templates {
			fmt.Printf("尝试执行模板: %s\n", t.Name())
			buf.Reset()
			if err := parsedTmpl.ExecuteTemplate(&buf, t.Name(), testData); err != nil {
				fmt.Printf("  ❌ 执行失败: %v\n", err)
			} else {
				fmt.Printf("  ✅ 执行成功, 结果长度: %d\n", buf.Len())
			}
		}
	} else {
		fmt.Printf("✅ 执行成功, 结果长度: %d\n", buf.Len())
	}
	
	return nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}