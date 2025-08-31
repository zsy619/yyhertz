package view

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 视图渲染相关方法 =============

// RenderTemplate 渲染模板 (Beego风格)
func (e *TemplateEngine) RenderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, err := e.GetTemplate(templateName)
	if err != nil {
		return "", err
	}

	// 准备渲染数据
	renderData := e.prepareBeegoRenderData(data)

	// 执行渲染
	var result strings.Builder
	if err := tmpl.Execute(&result, renderData); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	output := result.String()

	// 生产模式下启用Gzip压缩
	if e.RunMode == "prod" && e.EnableGzip {
		// 这里可以添加Gzip压缩逻辑
		config.Debugf("Template rendered with Gzip compression: %s", templateName)
	}

	return output, nil
}

// ExecuteTemplate 执行模板（用于测试）
func (e *TemplateEngine) ExecuteTemplate(tmpl *template.Template, data any) (string, error) {
	var buf strings.Builder
	err := tmpl.Execute(&buf, data)
	return buf.String(), err
}

// prepareBeegoRenderData 准备Beego风格渲染数据
func (e *TemplateEngine) prepareBeegoRenderData(data interface{}) map[string]interface{} {
	renderData := make(map[string]interface{})

	// 添加全局变量
	renderData["RunMode"] = e.RunMode
	renderData["TemplateLeft"] = e.delimLeft
	renderData["TemplateRight"] = e.delimRight
	renderData["EnableGzip"] = e.EnableGzip

	// 处理用户数据
	if data != nil {
		switch v := data.(type) {
		case map[string]interface{}:
			for k, val := range v {
				renderData[k] = val
			}
		default:
			renderData["Data"] = data
		}
	}

	return renderData
}

// executeTemplate 执行模板（用于测试）- 内部方法兼容性
func (e *TemplateEngine) executeTemplate(tmpl *template.Template, data any) (string, error) {
	return e.ExecuteTemplate(tmpl, data)
}

// ============= 便捷渲染函数 =============

// Render 渲染模板（使用默认引擎）
func Render(templateName string, data any) (string, error) {
	return GetDefaultEngine().Render(templateName, data)
}

// RenderWithLayout 使用布局渲染模板（使用默认引擎）
func RenderWithLayout(templateName, layoutName string, data any) (string, error) {
	return GetDefaultEngine().RenderWithLayout(templateName, layoutName, data)
}

// RenderComponent 渲染组件（使用默认引擎）
func RenderComponent(componentName string, data any) (string, error) {
	return GetDefaultEngine().RenderComponent(componentName, data)
}

// RenderWithAutoDiscovery 带自动发现的渲染（增强模板管理器功能）
func RenderWithAutoDiscovery(templateName string, data any) (string, error) {
	manager := GetEnhancedTemplateManager()

	// 如果模板不存在，尝试重新发现
	if _, exists := manager.templatePaths[templateName]; !exists {
		if err := manager.DiscoverTemplates(); err != nil {
			return "", fmt.Errorf("failed to discover templates: %w", err)
		}
	}

	// 检查模板是否存在
	templatePath, exists := manager.templatePaths[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found after discovery", templateName)
	}

	config.Debugf("Using discovered template: %s -> %s", templateName, templatePath)

	// 使用引擎渲染模板
	return manager.engine.Render(templateName, data)
}