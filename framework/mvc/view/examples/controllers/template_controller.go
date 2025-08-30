package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TemplateController 专门处理模板渲染的控制器
type TemplateController struct {
	mvc.BaseController
}

// Index 模板管理首页
func (c *TemplateController) Index() {
	// 🔍 调试：收集调试信息，直接在HTTP响应中显示
	cwd, _ := os.Getwd()
	debugInfo := []string{
		"=== TemplateController.Index 调试 ===",
		fmt.Sprintf("当前工作目录: %s", cwd),
	}

	// 扫描所有模板文件
	viewsDir := "./views"
	debugInfo = append(debugInfo, fmt.Sprintf("扫描目录: %s", viewsDir))
	debugInfo = append(debugInfo, fmt.Sprintf("绝对路径: %s", filepath.Join(cwd, "views")))

	// 🔍 调试：检查目录是否存在
	if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
		debugInfo = append(debugInfo, fmt.Sprintf("❌ views目录不存在: %s", viewsDir))
		// 尝试其他可能的路径
		alternativePaths := []string{"../views", "../../views", "./framework/mvc/view/examples/views"}
		for _, altPath := range alternativePaths {
			if _, err := os.Stat(altPath); err == nil {
				debugInfo = append(debugInfo, fmt.Sprintf("✅ 找到替代路径: %s", altPath))
				viewsDir = altPath
				break
			}
		}
	} else {
		debugInfo = append(debugInfo, "✅ views目录存在")
	}

	templates := c.scanTemplates(viewsDir)
	debugInfo = append(debugInfo, fmt.Sprintf("扫描结果: %d 个模板文件", len(templates)))

	// 🔍 关键调试：检查前几个模板的详细信息
	if len(templates) > 0 {
		debugInfo = append(debugInfo, "前5个模板:")
		for i, tmpl := range templates {
			if i >= 5 {
				break
			}
			debugInfo = append(debugInfo, fmt.Sprintf("  %d. %s (%s)", i+1, tmpl["name"], tmpl["path"]))
		}
	} else {
		debugInfo = append(debugInfo, "⚠️ 扫描结果为空!")
		// 如果扫描为空，进行额外诊断
		if files, err := os.ReadDir(viewsDir); err == nil {
			debugInfo = append(debugInfo, fmt.Sprintf("但目录确实包含 %d 个项目", len(files)))
		}
	}

	// 按类型统计模板
	pageCount := c.countTemplates(templates, "")
	layoutCount := c.countTemplates(templates, "layouts/")
	componentCount := c.countTemplates(templates, "components/")

	debugInfo = append(debugInfo, fmt.Sprintf("统计结果: 页面%d, 布局%d, 组件%d, 总计%d", pageCount, layoutCount, componentCount, len(templates)))

	// 如果扫描到0个模板，直接返回调试信息而不尝试渲染
	if len(templates) == 0 {
		c.SetData("Title", "调试信息 - 模板扫描失败")
		c.SetData("DebugInfo", debugInfo)
		c.SetData("ErrorMessage", "扫描到0个模板文件，可能是工作目录或路径配置问题")

		// 尝试直接返回调试页面
		debugHTML := "<html><head><title>调试信息</title></head><body><h1>模板扫描调试</h1><ul>"
		for _, info := range debugInfo {
			debugHTML += "<li>" + info + "</li>"
		}
		debugHTML += "</ul></body></html>"

		c.Ctx.String(200, debugHTML)
		return
	}

	// 设置模板数据
	c.SetData("Title", "模板管理 - YYHertz 演示")
	c.SetData("Templates", templates)
	c.SetData("PageCount", pageCount)
	c.SetData("LayoutCount", layoutCount)
	c.SetData("ComponentCount", componentCount)
	c.SetData("TotalCount", len(templates))
	c.SetData("ViewsDir", viewsDir)
	c.SetData("DebugInfo", debugInfo) // 添加调试信息

	// 创建统计数据结构，匹配模板期望的 .Stats 对象
	stats := map[string]any{
		"total":      len(templates),
		"layouts":    layoutCount,
		"components": componentCount,
		"pages":      pageCount,
	}
	c.SetData("Stats", stats)

	// 添加当前时间，供模板中的 now 和 dateformat 函数使用
	c.SetData("now", time.Now())

	c.Layout = ""
	c.TplName = "template/index.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString(err.Error())
	}
}

// Show 查看特定模板
func (c *TemplateController) Show() {
	templateName := c.Ctx.Param("name")
	if templateName == "" {
		c.Error(400, "模板名称不能为空")
		return
	}

	// 安全检查，防止路径遍历攻击
	templateName = strings.ReplaceAll(templateName, "..", "")
	templateName = strings.TrimPrefix(templateName, "/")

	templatePath := filepath.Join("./views", templateName)

	// 检查文件是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		c.Error(404, fmt.Sprintf("模板文件不存在: %s", templateName))
		return
	}

	// 读取模板内容
	content, err := os.ReadFile(templatePath)
	if err != nil {
		c.Error(500, fmt.Sprintf("读取模板文件失败: %v", err))
		return
	}

	c.SetData("Title", fmt.Sprintf("模板预览 - %s", templateName))
	c.SetData("TemplateName", templateName)
	c.SetData("TemplateContent", string(content))
	c.SetData("TemplateSize", len(content))

	// 获取文件信息
	fileInfo, _ := os.Stat(templatePath)
	c.SetData("LastModified", fileInfo.ModTime().Format("2006-01-02 15:04:05"))

	c.Layout = ""
	c.TplName = "template/show.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString(err.Error())
	}
}

// Preview 实时模板预览
func (c *TemplateController) Preview() {
	templateName := c.GetString("template")
	templateContent := c.GetString("content")

	if templateName == "" || templateContent == "" {
		c.Ctx.JSON(400, map[string]string{
			"error": "模板名称和内容不能为空",
		})
		return
	}

	// 创建模板引擎
	api := view.GetUnifiedAPI()

	// 解析数据（这里简化处理，实际应该解析JSON）
	data := map[string]any{
		"Title":     "预览测试",
		"Message":   "这是一个模板预览测试",
		"Timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 渲染模板
	result, err := api.RenderString(templateContent, data)
	if err != nil {
		c.Ctx.JSON(500, map[string]any{
			"error": fmt.Sprintf("模板渲染失败: %v", err),
		})
		return
	}

	c.Ctx.JSON(200, map[string]any{
		"status": "success",
		"result": result,
		"data":   data,
	})
}

// scanTemplates 扫描模板文件
func (c *TemplateController) scanTemplates(dir string) []map[string]any {
	var templates []map[string]any

	fmt.Printf("🔍 开始扫描模板目录: %s\n", dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("⚠️  路径访问错误 %s: %v\n", path, err)
			return nil // 忽略错误继续扫描
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".tpl")) {
			relPath, _ := filepath.Rel(dir, path)
			templates = append(templates, map[string]any{
				"name":     relPath,
				"path":     path,
				"size":     info.Size(),
				"modified": info.ModTime().Format("2006-01-02 15:04:05"),
				"type":     c.getTemplateType(relPath),
			})
			fmt.Printf("  ✅ 添加模板: %s\n", relPath)
		}
		return nil
	})

	if err != nil {
		// 如果扫描失败，返回空列表
		return templates
	}

	return templates
}

// countTemplates 计算特定类型的模板数量
func (c *TemplateController) countTemplates(templates []map[string]any, prefix string) int {
	count := 0
	for _, template := range templates {
		name := template["name"].(string)
		if prefix == "" {
			// 计算非layouts和components的模板
			if !strings.HasPrefix(name, "layouts/") && !strings.HasPrefix(name, "components/") {
				count++
			}
		} else {
			if strings.HasPrefix(name, prefix) {
				count++
			}
		}
	}
	return count
}

// getTemplateType 获取模板类型
func (c *TemplateController) getTemplateType(templateName string) string {
	if strings.HasPrefix(templateName, "layouts/") {
		return "layout"
	} else if strings.HasPrefix(templateName, "components/") {
		return "component"
	} else {
		return "page"
	}
}
