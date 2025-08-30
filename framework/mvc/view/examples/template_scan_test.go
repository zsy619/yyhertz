package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateScan(t *testing.T) {
	fmt.Printf("=== 模板扫描路径分析 ===\n")

	// 1. 检查当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ 获取工作目录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 当前工作目录: %s\n", cwd)

	// 2. 测试 TemplateController 中使用的相对路径
	viewsDir := "./views"
	fmt.Printf("✅ 使用的views路径: %s\n", viewsDir)

	// 3. 检查views目录是否存在
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		fmt.Printf("❌ views目录不存在: %s\n", viewsDir)
		return
	}
	fmt.Printf("✅ views目录存在\n")

	// 4. 扫描模板文件 (模拟TemplateController.scanTemplates)
	var templates []map[string]any

	err = filepath.Walk(viewsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("⚠️  访问路径出错 %s: %v\n", path, err)
			return nil // 忽略错误继续扫描
		}

		fmt.Printf("📁 扫描: %s (目录: %t)\n", path, info.IsDir())

		if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".tpl")) {
			relPath, _ := filepath.Rel(viewsDir, path)
			templates = append(templates, map[string]any{
				"name":     relPath,
				"path":     path,
				"size":     info.Size(),
				"modified": info.ModTime().Format("2006-01-02 15:04:05"),
			})
			fmt.Printf("  ✅ 找到模板: %s -> %s (%d bytes)\n", relPath, path, info.Size())
		}
		return nil
	})

	if err != nil {
		fmt.Printf("❌ 模板扫描失败: %v\n", err)
		return
	}

	fmt.Printf("\n=== 扫描结果 ===\n")
	fmt.Printf("总共找到 %d 个模板文件\n", len(templates))

	for i, tmpl := range templates {
		if i < 10 { // 只显示前10个
			fmt.Printf("%d. %s (%s)\n", i+1, tmpl["name"], tmpl["path"])
		}
	}

	if len(templates) > 10 {
		fmt.Printf("... 还有 %d 个模板\n", len(templates)-10)
	}

	// 5. 检查特定的template/index.html文件
	indexPath := "./views/template/index.html"
	fmt.Printf("\n=== 特定文件检查 ===\n")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Printf("❌ template/index.html 不存在: %s\n", indexPath)
	} else {
		fmt.Printf("✅ template/index.html 存在: %s\n", indexPath)
		// 读取文件头部内容
		content, err := os.ReadFile(indexPath)
		if err != nil {
			fmt.Printf("❌ 读取文件失败: %v\n", err)
		} else {
			lines := strings.Split(string(content), "\n")
			fmt.Printf("文件内容预览 (前5行):\n")
			for i, line := range lines {
				if i >= 5 {
					break
				}
				fmt.Printf("  %d: %s\n", i+1, line)
			}
		}
	}

	fmt.Printf("\n=== 分析完成 ===\n")
}
