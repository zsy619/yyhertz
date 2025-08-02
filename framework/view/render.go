package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// RenderData 渲染数据结构
type RenderData struct {
	Data    any          `json:"data"`
	Meta    *MetaData    `json:"meta,omitempty"`
	Flash   *FlashData   `json:"flash,omitempty"`
	CSRF    string       `json:"csrf,omitempty"`
	Theme   string       `json:"theme,omitempty"`
	User    any          `json:"user,omitempty"`
	Request *RequestData `json:"request,omitempty"`
}

// MetaData 页面元数据
type MetaData struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    string            `json:"keywords"`
	Author      string            `json:"author"`
	Canonical   string            `json:"canonical"`
	Image       string            `json:"image"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// FlashData 闪存消息
type FlashData struct {
	Success []string `json:"success,omitempty"`
	Error   []string `json:"error,omitempty"`
	Warning []string `json:"warning,omitempty"`
	Info    []string `json:"info,omitempty"`
}

// RequestData 请求信息
type RequestData struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Query     string `json:"query"`
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
	Timestamp int64  `json:"timestamp"`
}

// Render 渲染模板
func (e *TemplateEngine) Render(templateName string, data any) (string, error) {
	return e.RenderWithLayout(templateName, "", data)
}

// RenderWithLayout 使用布局渲染模板
func (e *TemplateEngine) RenderWithLayout(templateName, layoutName string, data any) (string, error) {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	// 准备渲染数据
	renderData := e.prepareRenderData(data)

	var tmpl *template.Template
	var err error

	if layoutName != "" {
		// 使用布局渲染
		tmpl, err = e.getTemplateWithLayout(templateName, layoutName)
	} else {
		// 直接渲染模板
		tmpl, err = e.getTemplate(templateName)
	}

	if err != nil {
		return "", fmt.Errorf("template loading error: %w", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, renderData); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	result := buf.String()

	// 如果启用压缩，移除多余空白
	if e.enableCompress {
		result = e.compressHTML(result)
	}

	return result, nil
}

// RenderComponent 渲染组件
func (e *TemplateEngine) RenderComponent(componentName string, data any) (string, error) {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	component, exists := e.components[componentName]
	if !exists {
		return "", fmt.Errorf("component '%s' not found", componentName)
	}

	renderData := e.prepareRenderData(data)

	var buf bytes.Buffer
	if err := component.Execute(&buf, renderData); err != nil {
		return "", fmt.Errorf("component execution error: %w", err)
	}

	return buf.String(), nil
}

// getTemplate 获取模板
func (e *TemplateEngine) getTemplate(templateName string) (*template.Template, error) {
	// 从缓存获取
	if e.enableCache {
		if tmpl, exists := e.templates[templateName]; exists {
			return tmpl, nil
		}
	}

	// 动态加载模板
	tmpl, err := e.loadTemplate(templateName)
	if err != nil {
		return nil, err
	}

	// 缓存模板
	if e.enableCache {
		e.templates[templateName] = tmpl
	}

	return tmpl, nil
}

// getTemplateWithLayout 获取带布局的模板
func (e *TemplateEngine) getTemplateWithLayout(templateName, layoutName string) (*template.Template, error) {
	cacheKey := fmt.Sprintf("%s@%s", templateName, layoutName)

	// 从缓存获取
	if e.enableCache {
		if tmpl, exists := e.templates[cacheKey]; exists {
			return tmpl, nil
		}
	}

	// 获取布局模板
	layout, exists := e.layouts[layoutName]
	if !exists {
		return nil, fmt.Errorf("layout '%s' not found", layoutName)
	}

	// 克隆布局模板
	tmpl, err := layout.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone layout: %w", err)
	}

	// 加载并解析内容模板
	contentPath, err := e.findTemplateFile(templateName)
	if err != nil {
		return nil, err
	}

	if _, err := tmpl.ParseFiles(contentPath); err != nil {
		return nil, fmt.Errorf("failed to parse content template: %w", err)
	}

	// 缓存组合模板
	if e.enableCache {
		e.templates[cacheKey] = tmpl
	}

	return tmpl, nil
}

// loadTemplate 动态加载模板
func (e *TemplateEngine) loadTemplate(templateName string) (*template.Template, error) {
	templatePath, err := e.findTemplateFile(templateName)
	if err != nil {
		return nil, err
	}

	tmpl := template.New(templateName).
		Delims(e.delimLeft, e.delimRight).
		Funcs(e.funcMap)

	if _, err := tmpl.ParseFiles(templatePath); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// findTemplateFile 查找模板文件
func (e *TemplateEngine) findTemplateFile(templateName string) (string, error) {
	// 确保模板名有扩展名
	if !strings.HasSuffix(templateName, e.extension) {
		templateName += e.extension
	}

	// 在所有视图路径中搜索
	for _, viewPath := range e.viewPaths {
		templatePath := fmt.Sprintf("%s/%s", viewPath, templateName)

		// 检查文件是否存在（这里简化处理，实际应该用os.Stat）
		// 为了避免依赖os包，我们暂时返回第一个路径
		return templatePath, nil
	}

	return "", fmt.Errorf("template file '%s' not found in view paths: %v", templateName, e.viewPaths)
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
	}

	return renderData
}

// compressHTML 压缩HTML
func (e *TemplateEngine) compressHTML(html string) string {
	// 简单的HTML压缩：移除多余的空白字符
	lines := strings.Split(html, "\n")
	compressed := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			compressed = append(compressed, trimmed)
		}
	}

	return strings.Join(compressed, "\n")
}

// ============= 模板函数实现 =============

// includeTemplate 包含其他模板
func (e *TemplateEngine) includeTemplate(templateName string, data any) template.HTML {
	content, err := e.Render(templateName, data)
	if err != nil {
		config.Errorf("Include template error: %v", err)
		return template.HTML(fmt.Sprintf("<!-- Include error: %s -->", err.Error()))
	}
	return template.HTML(content)
}

// renderComponent 渲染组件
func (e *TemplateEngine) renderComponent(componentName string, data any) template.HTML {
	content, err := e.RenderComponent(componentName, data)
	if err != nil {
		config.Errorf("Component render error: %v", err)
		return template.HTML(fmt.Sprintf("<!-- Component error: %s -->", err.Error()))
	}
	return template.HTML(content)
}

// getThemeVariable 获取主题变量
func (e *TemplateEngine) getThemeVariable(key string) string {
	if theme, exists := e.themes[e.currentTheme]; exists {
		if value, exists := theme.Variables[key]; exists {
			return value
		}
	}
	return ""
}

// getAssetURL 获取资源URL
func (e *TemplateEngine) getAssetURL(path string) string {
	if theme, exists := e.themes[e.currentTheme]; exists {
		return fmt.Sprintf("/%s/%s", theme.StaticPath, strings.TrimPrefix(path, "/"))
	}
	return fmt.Sprintf("/static/%s", strings.TrimPrefix(path, "/"))
}

// buildURL 构建URL
func (e *TemplateEngine) buildURL(path string, params ...any) string {
	url := strings.TrimPrefix(path, "/")

	// 如果有参数，构建查询字符串
	if len(params) > 0 {
		query := make([]string, 0, len(params)/2)
		for i := 0; i < len(params)-1; i += 2 {
			key := fmt.Sprintf("%v", params[i])
			value := fmt.Sprintf("%v", params[i+1])
			query = append(query, fmt.Sprintf("%s=%s", key, value))
		}
		if len(query) > 0 {
			url += "?" + strings.Join(query, "&")
		}
	}

	return "/" + url
}

// getCSRFToken 获取CSRF令牌
func (e *TemplateEngine) getCSRFToken() string {
	// 这里应该从请求上下文中获取CSRF令牌
	// 暂时返回占位符
	return "csrf-token-placeholder"
}

// getFlashMessage 获取Flash消息
func (e *TemplateEngine) getFlashMessage(msgType string) []string {
	// 这里应该从会话中获取Flash消息
	// 暂时返回空切片
	return []string{}
}

// truncateString 截断字符串
func (e *TemplateEngine) truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// renderMarkdown 渲染Markdown（简化版）
func (e *TemplateEngine) renderMarkdown(content string) template.HTML {
	// 这里应该使用真正的Markdown渲染器
	// 暂时进行简单的转换
	html := strings.ReplaceAll(content, "\n", "<br>")
	html = strings.ReplaceAll(html, "**", "<strong>")
	html = strings.ReplaceAll(html, "**", "</strong>")
	return template.HTML(html)
}

// toJSON 转换为JSON
func (e *TemplateEngine) toJSON(data any) template.JS {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return template.JS("{}")
	}
	return template.JS(jsonData)
}

// safeHTML 安全HTML
func (e *TemplateEngine) safeHTML(content string) template.HTML {
	return template.HTML(content)
}

// createDict 创建字典
func (e *TemplateEngine) createDict(values ...any) map[string]any {
	dict := make(map[string]any)
	for i := 0; i < len(values)-1; i += 2 {
		key := fmt.Sprintf("%v", values[i])
		dict[key] = values[i+1]
	}
	return dict
}

// createSlice 创建切片
func (e *TemplateEngine) createSlice(values ...any) []any {
	return values
}

// createRange 创建数字范围
func (e *TemplateEngine) createRange(start, end int) []int {
	if start > end {
		return []int{}
	}

	result := make([]int, end-start+1)
	for i := 0; i < len(result); i++ {
		result[i] = start + i
	}
	return result
}

// formatDate 格式化日期
func (e *TemplateEngine) formatDate(date any, format string) string {
	var t time.Time

	switch v := date.(type) {
	case time.Time:
		t = v
	case *time.Time:
		if v != nil {
			t = *v
		} else {
			return ""
		}
	case int64:
		t = time.Unix(v, 0)
	case string:
		// 尝试解析字符串日期
		if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			t = parsed
		} else {
			return v
		}
	default:
		return fmt.Sprintf("%v", date)
	}

	// 格式化映射
	switch format {
	case "date":
		return t.Format("2006-01-02")
	case "datetime":
		return t.Format("2006-01-02 15:04:05")
	case "time":
		return t.Format("15:04:05")
	case "iso":
		return t.Format(time.RFC3339)
	case "rfc":
		return t.Format(time.RFC822)
	default:
		return t.Format(format)
	}
}

// formatCurrency 格式化货币
func (e *TemplateEngine) formatCurrency(amount any, currency string) string {
	var value float64

	switch v := amount.(type) {
	case float64:
		value = v
	case float32:
		value = float64(v)
	case int:
		value = float64(v)
	case int64:
		value = float64(v)
	default:
		return fmt.Sprintf("%v", amount)
	}

	switch currency {
	case "CNY", "RMB", "¥":
		return fmt.Sprintf("¥%.2f", value)
	case "USD", "$":
		return fmt.Sprintf("$%.2f", value)
	case "EUR", "€":
		return fmt.Sprintf("€%.2f", value)
	default:
		return fmt.Sprintf("%.2f %s", value, currency)
	}
}

// formatFileSize 格式化文件大小
func (e *TemplateEngine) formatFileSize(size any) string {
	var bytes int64

	switch v := size.(type) {
	case int64:
		bytes = v
	case int:
		bytes = int64(v)
	case float64:
		bytes = int64(v)
	default:
		return fmt.Sprintf("%v", size)
	}

	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ============= 辅助方法 =============

// GetTemplateList 获取模板列表
func (e *TemplateEngine) GetTemplateList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	templates := make([]string, 0, len(e.templates))
	for name := range e.templates {
		templates = append(templates, name)
	}
	return templates
}

// GetLayoutList 获取布局列表
func (e *TemplateEngine) GetLayoutList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	layouts := make([]string, 0, len(e.layouts))
	for name := range e.layouts {
		layouts = append(layouts, name)
	}
	return layouts
}

// GetComponentList 获取组件列表
func (e *TemplateEngine) GetComponentList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	components := make([]string, 0, len(e.components))
	for name := range e.components {
		components = append(components, name)
	}
	return components
}

// ClearCache 清除模板缓存
func (e *TemplateEngine) ClearCache() {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)

	config.Info("Template cache cleared")
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
		"available_themes": e.GetAvailableThemes(),
		"cache_enabled":    e.enableCache,
		"reload_enabled":   e.enableReload,
		"compress_enabled": e.enableCompress,
		"view_paths":       e.viewPaths,
		"layout_path":      e.layoutPath,
		"component_path":   e.componentPath,
	}
}
