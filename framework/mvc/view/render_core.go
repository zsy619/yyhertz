package view

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/zsy619/yyhertz/framework/config"
)

// Render 渲染模板
func (e *TemplateEngine) Render(templateName string, data any) (string, error) {
	return e.RenderWithLayout(templateName, "", data)
}

// RenderWithLayout 使用布局渲染模板
func (e *TemplateEngine) RenderWithLayout(templateName, layoutName string, data any) (string, error) {
	// 准备渲染数据
	renderData := e.prepareRenderData(data)

	var tmpl *template.Template
	var err error

	if layoutName != "" {
		// 使用布局渲染
		tmpl, err = e.GetTemplateWithLayout(templateName, layoutName)
	} else {
		// 直接渲染模板
		tmpl, err = e.GetTemplate(templateName)
	}

	if err != nil {
		return "", fmt.Errorf("template loading error: %w", err)
	}

	// 渲染模板 - 增强调试信息
	var buf bytes.Buffer
	config.Debugf("🎯 开始执行模板: %s (布局: %s)", templateName, layoutName)
	config.Debugf("📦 渲染数据类型: %T, 数据内容预览: %+v", renderData.Data, renderData.Data)

	if err := tmpl.Execute(&buf, renderData.Data); err != nil {
		// 提供更详细的错误信息
		errorDetails := fmt.Sprintf(`
❌ 模板执行失败详情:
  - 模板名称: %s
  - 布局名称: %s  
  - 模板路径: %v
  - 错误类型: %T
  - 错误内容: %v
  - 数据类型: %T
  - 可用函数数量: %d`,
			templateName, layoutName, e.viewPaths, err, err, renderData.Data, len(e.funcMap))

		return "", fmt.Errorf("template execution error: %s", errorDetails)
	}

	result := buf.String()

	// 如果启用压缩，移除多余空白
	if e.enableCompress {
		result = e.compressHTML(result)
	}

	return result, nil
}

// RenderWithLayoutSections 使用布局和区块渲染模板
func (e *TemplateEngine) RenderWithLayoutSections(templateName, layoutName string, data any, layoutSections map[string]string) (string, error) {
	// 准备渲染数据
	renderData := e.prepareRenderData(data)

	// 添加布局区块到渲染数据中
	if renderData.LayoutSections == nil {
		renderData.LayoutSections = make(map[string]string)
	}

	// 合并传入的布局区块
	for key, value := range layoutSections {
		renderData.LayoutSections[key] = value
	}

	var tmpl *template.Template
	var err error

	if layoutName != "" {
		// 使用布局渲染（支持布局区块）
		tmpl, err = e.GetTemplateWithLayoutSections(templateName, layoutName, layoutSections)
	} else {
		// 直接渲染模板
		tmpl, err = e.GetTemplate(templateName)
	}

	if err != nil {
		return "", fmt.Errorf("template loading error: %w", err)
	}

	// 渲染模板 - 增强调试信息 (布局区块版本)
	var buf bytes.Buffer
	config.Debugf("🎯 开始执行布局区块模板: %s (布局: %s, 区块数: %d)", templateName, layoutName, len(layoutSections))
	config.Debugf("📦 渲染数据类型: %T, 数据内容预览: %+v", renderData.Data, renderData.Data)

	if err := tmpl.Execute(&buf, renderData.Data); err != nil {
		// 提供更详细的错误信息
		sectionNames := make([]string, 0, len(layoutSections))
		for name := range layoutSections {
			sectionNames = append(sectionNames, name)
		}

		errorDetails := fmt.Sprintf(`
❌ 布局区块模板执行失败详情:
  - 模板名称: %s
  - 布局名称: %s
  - 布局区块: %v  
  - 模板路径: %v
  - 错误类型: %T
  - 错误内容: %v
  - 数据类型: %T
  - 可用函数数量: %d`,
			templateName, layoutName, sectionNames, e.viewPaths, err, err, renderData.Data, len(e.funcMap))

		return "", fmt.Errorf("template execution error: %s", errorDetails)
	}

	result := buf.String()

	// 如果启用压缩，移除多余空白
	if e.enableCompress {
		result = e.compressHTML(result)
	}

	return result, nil
}

// prepareRenderData 准备渲染数据
func (e *TemplateEngine) prepareRenderData(data any) *RenderData {
	renderData := &RenderData{
		Data:  data,
		Theme: e.currentTheme,
	}

	// 如果数据已经是RenderData类型，直接使用
	if rd, ok := data.(*RenderData); ok {
		renderData = rd
		if renderData.Theme == "" {
			renderData.Theme = e.currentTheme
		}
		// 确保CSRF token已设置，如果为空则从模板引擎获取
		if renderData.CSRF == "" {
			renderData.CSRF = e.getCSRFToken()
		}
		// 同时设置CsrfToken字段，确保与CSRF字段同步
		renderData.CsrfToken = renderData.CSRF
	} else {
		// 为新的RenderData设置默认的CSRF token
		csrfToken := e.getCSRFToken()
		renderData.CSRF = csrfToken
		renderData.CsrfToken = csrfToken // 同时设置两个字段
	}

	return renderData
}

// PrepareRenderData 公开的渲染数据准备方法（用于测试和外部调用）
func (e *TemplateEngine) PrepareRenderData(data any) *RenderData {
	return e.prepareRenderData(data)
}

// GetStats 获取模板引擎统计信息
func (e *TemplateEngine) GetStats() map[string]any {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	return map[string]any{
		"templates_count":  len(e.templates),
		"layouts_count":    len(e.layouts),
		"components_count": len(e.components),
		"current_theme":    e.currentTheme,
		"available_themes": e.getAvailableThemesFromEngine(),
		"cache_enabled":    e.enableCache,
		"reload_enabled":   e.enableReload,
		"compress_enabled": e.enableCompress,
		"view_paths":       e.viewPaths,
		"layout_path":      e.layoutPath,
		"component_path":   e.componentPath,
	}
}

// getAvailableThemesFromEngine 获取可用主题列表（内部方法，避免与engine_theme.go冲突）
func (e *TemplateEngine) getAvailableThemesFromEngine() []string {
	themes := make([]string, 0, len(e.themes))
	for name := range e.themes {
		themes = append(themes, name)
	}
	return themes
}