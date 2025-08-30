package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 直接从 TemplateController 复制的 scanTemplates 方法
func scanTemplates2(dir string) []map[string]any {
	var templates []map[string]any

	fmt.Printf("🔍 开始扫描模板目录: %s\n", dir)
	fmt.Printf("📁 绝对路径: %s\n", func() string {
		if filepath.IsAbs(dir) {
			return dir
		}
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, dir)
	}())

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("⚠️  路径访问错误 %s: %v\n", path, err)
			return nil // 忽略错误继续扫描
		}

		fmt.Printf("🚶 正在访问: %s (目录: %t)\n", path, info.IsDir())

		if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".tpl")) {
			relPath, _ := filepath.Rel(dir, path)
			templates = append(templates, map[string]any{
				"name":     relPath,
				"path":     path,
				"size":     info.Size(),
				"modified": info.ModTime().Format("2006-01-02 15:04:05"),
				"type":     getTemplateType2(relPath),
			})
			fmt.Printf("  ✅ 添加模板: %s\n", relPath)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("❌ 扫描过程出错: %v\n", err)
	}

	fmt.Printf("📊 最终结果: %d 个模板文件\n", len(templates))
	return templates
}

func getTemplateType2(relPath string) string {
	if strings.HasPrefix(relPath, "layouts/") {
		return "layout"
	} else if strings.HasPrefix(relPath, "components/") {
		return "component"
	} else {
		return "page"
	}
}

func TestExatScan(t *testing.T) {
	fmt.Printf("=== TemplateController.scanTemplates 精确测试 ===\n")

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ 获取工作目录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 当前工作目录: %s\n", cwd)

	// 测试与 TemplateController 完全相同的扫描逻辑
	viewsDir := "./views"
	fmt.Printf("📁 扫描目录: %s\n", viewsDir)

	// 检查目录是否存在
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		fmt.Printf("❌ views目录不存在: %s\n", viewsDir)
		fmt.Printf("完整路径: %s\n", filepath.Join(cwd, "views"))
		return
	}

	fmt.Printf("✅ views目录存在\n")

	// 🔍 重点：使用与 TemplateController 完全相同的扫描方法
	templates := scanTemplates2(viewsDir)

	fmt.Printf("📈 最终统计:\n")
	fmt.Printf("  总模板数: %d\n", len(templates))

	if len(templates) == 0 {
		fmt.Printf("❌ 扫描到0个模板！这就是问题所在！\n")

		// 进行更详细的诊断
		fmt.Printf("\n=== 详细诊断 ===\n")

		// 直接检查 views 目录内容
		files, err := os.ReadDir(viewsDir)
		if err != nil {
			fmt.Printf("❌ 无法读取views目录: %v\n", err)
		} else {
			fmt.Printf("📁 views目录包含 %d 个项目:\n", len(files))
			for i, file := range files {
				if i < 10 { // 只显示前10个
					fmt.Printf("  - %s (目录: %t)\n", file.Name(), file.IsDir())
				}
			}
		}

		// 测试 filepath.Walk 是否工作
		fmt.Printf("\n=== 测试 filepath.Walk ===\n")
		walkCount := 0
		filepath.Walk(viewsDir, func(path string, info os.FileInfo, err error) error {
			walkCount++
			if walkCount <= 5 {
				fmt.Printf("  访问: %s\n", path)
			}
			return nil
		})
		fmt.Printf("Walk 访问了 %d 个路径\n", walkCount)
	} else {
		fmt.Printf("✅ 扫描成功！找到模板: %s\n", templates[0]["name"])
	}
}
