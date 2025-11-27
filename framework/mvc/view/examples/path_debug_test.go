package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPath(t *testing.T) {
	fmt.Printf("=== 路径调试分析 ===\n")

	// 1. 检查当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ 获取工作目录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 当前工作目录: %s\n", cwd)

	// 2. 检查 ./views 目录
	viewsDir := "./views"
	fmt.Printf("\n=== 检查views目录 ===\n")
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		fmt.Printf("❌ views目录不存在: %s\n", viewsDir)
		fmt.Printf("绝对路径应该是: %s\n", filepath.Join(cwd, "views"))
	} else {
		fmt.Printf("✅ views目录存在: %s\n", viewsDir)

		// 扫描views目录内容
		fmt.Printf("\n=== views目录内容 ===\n")
		err := filepath.Walk(viewsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("⚠️  访问路径出错 %s: %v\n", path, err)
				return nil
			}

			if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".tpl")) {
				relPath, _ := filepath.Rel(viewsDir, path)
				fmt.Printf("  📄 找到模板: %s -> %s\n", relPath, path)
			}
			return nil
		})

		if err != nil {
			fmt.Printf("❌ 扫描views目录失败: %v\n", err)
		}
	}

	// 3. 检查当前目录的所有内容
	fmt.Printf("\n=== 当前目录内容 ===\n")
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Printf("❌ 读取当前目录失败: %v\n", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("📁 目录: %s\n", file.Name())
		} else {
			fmt.Printf("📄 文件: %s\n", file.Name())
		}
	}

	// 4. 特别检查template相关目录
	fmt.Printf("\n=== 寻找template相关目录 ===\n")
	possibleDirs := []string{"views", "templates", "template", "tpl"}
	for _, dir := range possibleDirs {
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("✅ 找到目录: %s\n", dir)
			// 快速扫描内容
			subFiles, err := os.ReadDir(dir)
			if err == nil {
				templateCount := 0
				for _, subFile := range subFiles {
					if strings.HasSuffix(subFile.Name(), ".html") || strings.HasSuffix(subFile.Name(), ".tpl") {
						templateCount++
					}
				}
				fmt.Printf("   包含 %d 个模板文件\n", templateCount)
			}
		} else {
			fmt.Printf("❌ 目录不存在: %s\n", dir)
		}
	}

	fmt.Printf("\n=== 分析完成 ===\n")
}
