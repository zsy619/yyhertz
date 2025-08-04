package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/zsy619/yyhertz/framework/mvc"
)

type MarkdownController struct {
	mvc.BaseController
}

func (c *MarkdownController) GetMarkdown() {
	// 获取文件路径参数
	filePath := c.GetString("path")
	
	if filePath == "" {
		filePath = c.GetParam("path")
	}

	if filePath == "" {
		filePath = c.GetForm("path")
	}

	log.Printf("GetMarkdown called with path: '%s'", filePath)

	if filePath == "" {
		c.Error(400, "Missing file path parameter. Please use ?path=filename")
		return
	}

	// 安全检查：防止路径遍历攻击
	if strings.Contains(filePath, "..") {
		c.Error(403, "Invalid file path")
		return
	}

	// 构建完整文件路径 - 使用绝对路径
	fullPath := filepath.Join("/Volumes/E/JYW/YYHertz/example/simple/docs", filePath)
	if !strings.HasSuffix(fullPath, ".md") {
		fullPath += ".md"
	}

	// 读取Markdown文件
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		c.Error(404, fmt.Sprintf("File not found: %s", filePath))
		return
	}

	// 配置Goldmark解析器
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,            // GitHub Flavored Markdown
			extension.Table,          // 表格支持
			extension.Strikethrough,  // 删除线
			extension.TaskList,       // 任务列表
			extension.DefinitionList, // 定义列表
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // 硬换行
			html.WithXHTML(),     // XHTML兼容
			html.WithUnsafe(),    // 允许HTML标签
		),
	)

	// 解析Markdown为HTML
	var htmlBuf bytes.Buffer
	if err := md.Convert(content, &htmlBuf); err != nil {
		c.Error(500, "Failed to parse Markdown")
		return
	}

	// 渲染模板
	c.RenderHTML("markdown/markdown.html", map[string]interface{}{
		"Title":      strings.TrimSuffix(filepath.Base(filePath), ".md"),
		"Content":    template.HTML(htmlBuf.String()),
		"RawContent": string(content),
		"FilePath":   filePath,
	})
}

func (c *MarkdownController) ExportPDF() {
	// 尝试多种方式获取path参数
	filePath := c.GetString("path")

	if filePath == "" {
		filePath = c.GetParam("path")
	}

	if filePath == "" {
		filePath = c.GetForm("path")
	}
	if filePath == "" {
		filePath = c.GetRouteParam("path")
	}

	log.Printf("ExportPDF called with path: '%s' (query: '%s', param: '%s', form: '%s')",
		filePath, c.GetString("path"), c.GetParam("path"), c.GetForm("path"))

	if filePath == "" {
		c.Error(400, "Missing file path parameter. Please use ?path=filename")
		return
	}

	// 安全检查
	if strings.Contains(filePath, "..") {
		c.Error(403, "Invalid file path")
		return
	}

	// 构建完整文件路径 - 使用绝对路径
	fullPath := filepath.Join("/Volumes/E/JYW/YYHertz/example/simple/docs", filePath)
	if !strings.HasSuffix(fullPath, ".md") {
		fullPath += ".md"
	}

	// 读取Markdown文件
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		c.Error(404, fmt.Sprintf("File not found: %s", filePath))
		return
	}

	// 转换Markdown为HTML
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			extension.DefinitionList,
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	var htmlBuf bytes.Buffer
	if err := md.Convert(content, &htmlBuf); err != nil {
		c.Error(500, "Failed to parse Markdown")
		return
	}

	// 简单HTML格式用于演示（作为PDF的替代）
	fileName := strings.TrimSuffix(filepath.Base(filePath), ".md") + ".html"
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	// 返回HTML内容作为演示
	htmlDoc := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1, h2, h3, h4, h5, h6 { color: #333; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 5px; }
        code { background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
        table { border-collapse: collapse; width: 100%%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        blockquote { border-left: 4px solid #ddd; margin: 0; padding-left: 20px; }
    </style>
</head>
<body>
%s
</body>
</html>`, fileName, htmlBuf.String())

	c.String(htmlDoc)
}

// Debug 调试端点
func (c *MarkdownController) GetDebug() {
	log.Printf("Debug endpoint called")

	var result string

	result += fmt.Sprintf("GetString: '%s'\n", c.GetString("path"))
	result += fmt.Sprintf("GetParam: '%s'\n", c.GetParam("path"))
	result += fmt.Sprintf("GetForm: '%s'\n", c.GetForm("path"))

	c.String(result)
}

func (c *MarkdownController) GetList() {
	docsDir := "/Volumes/E/JYW/YYHertz/example/simple/docs"
	log.Printf("GetList called, looking in directory: %s", docsDir)

	// 获取所有Markdown文件
	files, err := filepath.Glob(filepath.Join(docsDir, "*.md"))
	log.Printf("Found files: %v, error: %v", files, err)
	if err != nil {
		c.Error(500, "Failed to list files")
		return
	}

	var fileList []map[string]interface{}
	for _, file := range files {
		baseName := filepath.Base(file)
		fileName := strings.TrimSuffix(baseName, ".md")

		// 获取文件信息
		info, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		// 读取前100个字符作为摘要
		summary := string(info)
		if len(summary) > 100 {
			summary = summary[:100] + "..."
		}

		fileList = append(fileList, map[string]interface{}{
			"Name":    fileName,
			"Path":    fileName,
			"Summary": summary,
		})
	}

	c.RenderHTML("markdown/list.html", map[string]interface{}{
		"Title":     "Markdown Files",
		"Files":     fileList,
		"FileCount": len(fileList),
	})
}
