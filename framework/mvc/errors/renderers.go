package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 渲染器实现 =============

// JSONRenderer JSON错误渲染器
type JSONRenderer struct {
	showDetails bool
	prettyPrint bool
}

// NewJSONRenderer 创建JSON渲染器
func NewJSONRenderer(showDetails, prettyPrint bool) *JSONRenderer {
	return &JSONRenderer{
		showDetails: showDetails,
		prettyPrint: prettyPrint,
	}
}

// Render 渲染JSON响应
func (r *JSONRenderer) Render(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	response := map[string]any{
		"code":      errorCtx.StatusCode,
		"message":   errorCtx.ErrorMessage,
		"success":   false,
		"path":      errorCtx.RequestPath,
		"method":    errorCtx.RequestMethod,
		"timestamp": errorCtx.Timestamp.Unix(),
	}
	
	if r.showDetails {
		if len(errorCtx.Suggestions) > 0 {
			response["suggestions"] = errorCtx.Suggestions
		}
		
		if errorCtx.RequestID != "" {
			response["request_id"] = errorCtx.RequestID
		}
		
		if len(errorCtx.Details) > 0 {
			response["details"] = errorCtx.Details
		}
	}
	
	if r.prettyPrint {
		ctx.SetContentType("application/json; charset=utf-8")
		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return err
		}
		ctx.String(errorCtx.StatusCode, string(data))
	} else {
		ctx.JSON(errorCtx.StatusCode, response)
	}
	
	return nil
}

// CanRender 检查是否能渲染
func (r *JSONRenderer) CanRender(ctx *mvccontext.Context) bool {
	accept := string(ctx.GetHeader("Accept"))
	path := string(ctx.Path())
	
	// API路径优先使用JSON
	if strings.HasPrefix(path, "/api/") {
		return true
	}
	
	// Accept头明确要求JSON
	return strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html")
}

// ContentType 返回内容类型
func (r *JSONRenderer) ContentType() string {
	return "application/json"
}

// ============= HTML渲染器 =============

// HTMLRenderer HTML错误页面渲染器
type HTMLRenderer struct {
	templateManager TemplateManager
	templateName    string
	showDetails     bool
	customTitle     string
	supportEmail    string
	supportPhone    string
}

// NewHTMLRenderer 创建HTML渲染器
func NewHTMLRenderer(templateManager TemplateManager) *HTMLRenderer {
	return &HTMLRenderer{
		templateManager: templateManager,
		templateName:    "error.html",
		showDetails:     true,
		customTitle:     "YYHertz Framework",
	}
}

// SetTemplate 设置模板名称
func (r *HTMLRenderer) SetTemplate(templateName string) {
	r.templateName = templateName
}

// SetShowDetails 设置是否显示详情
func (r *HTMLRenderer) SetShowDetails(show bool) {
	r.showDetails = show
}

// SetSupportInfo 设置支持信息
func (r *HTMLRenderer) SetSupportInfo(email, phone string) {
	r.supportEmail = email
	r.supportPhone = phone
}

// Render 渲染HTML页面
func (r *HTMLRenderer) Render(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	// 创建模板数据
	templateData := NewTemplateData(errorCtx)
	templateData.SetFrameworkInfo(r.customTitle, "错误页面")
	templateData.SetSupportInfo(r.supportEmail, r.supportPhone)
	templateData.ShowDebug = r.showDetails
	
	// 从配置获取图标
	config := GetStatusConfig(errorCtx.StatusCode)
	templateData.SetConfig(config)
	
	// 渲染模板
	html, err := r.templateManager.RenderTemplate(r.templateName, templateData)
	if err != nil {
		// 模板渲染失败，使用简单的HTML
		return r.renderFallbackHTML(ctx, errorCtx)
	}
	
	ctx.SetContentType("text/html; charset=utf-8")
	ctx.String(errorCtx.StatusCode, html)
	return nil
}

// CanRender 检查是否能渲染
func (r *HTMLRenderer) CanRender(ctx *mvccontext.Context) bool {
	accept := string(ctx.GetHeader("Accept"))
	path := string(ctx.Path())
	
	// 非API路径且Accept包含HTML
	if !strings.HasPrefix(path, "/api/") && strings.Contains(accept, "text/html") {
		return true
	}
	
	// 默认情况下，如果没有明确要求JSON，则使用HTML
	return !strings.Contains(accept, "application/json")
}

// ContentType 返回内容类型
func (r *HTMLRenderer) ContentType() string {
	return "text/html"
}

// renderFallbackHTML 渲染后备HTML
func (r *HTMLRenderer) renderFallbackHTML(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>错误 %d</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .error { background: #f8f9fa; padding: 30px; border-radius: 8px; border-left: 5px solid #dc3545; }
        h1 { color: #dc3545; }
        .btn { padding: 10px 20px; margin: 10px; text-decoration: none; background: #007bff; color: white; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="error">
        <h1>%d %s</h1>
        <p>%s</p>
        <p><strong>路径:</strong> %s</p>
        <p><strong>时间:</strong> %s</p>
        <div>
            <a href="javascript:history.back()" class="btn">返回上页</a>
            <a href="/" class="btn">返回首页</a>
        </div>
    </div>
</body>
</html>`,
		errorCtx.StatusCode,
		errorCtx.StatusCode, errorCtx.StatusText,
		errorCtx.ErrorMessage,
		errorCtx.RequestPath,
		errorCtx.Timestamp.Format("2006-01-02 15:04:05"),
	)
	
	ctx.SetContentType("text/html; charset=utf-8")
	ctx.String(errorCtx.StatusCode, html)
	return nil
}

// ============= XML渲染器 =============

// XMLRenderer XML错误渲染器
type XMLRenderer struct {
	prettyPrint bool
}

// NewXMLRenderer 创建XML渲染器
func NewXMLRenderer(prettyPrint bool) *XMLRenderer {
	return &XMLRenderer{
		prettyPrint: prettyPrint,
	}
}

// Render 渲染XML响应
func (r *XMLRenderer) Render(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<error>
    <code>%d</code>
    <title>%s</title>
    <message>%s</message>
    <path>%s</path>
    <method>%s</method>
    <timestamp>%s</timestamp>`,
		errorCtx.StatusCode,
		errorCtx.StatusText,
		errorCtx.ErrorMessage,
		errorCtx.RequestPath,
		errorCtx.RequestMethod,
		errorCtx.Timestamp.Format("2006-01-02T15:04:05Z"),
	)
	
	if errorCtx.RequestID != "" {
		xml += fmt.Sprintf("\n    <request_id>%s</request_id>", errorCtx.RequestID)
	}
	
	if len(errorCtx.Suggestions) > 0 {
		xml += "\n    <suggestions>"
		for _, suggestion := range errorCtx.Suggestions {
			xml += fmt.Sprintf("\n        <suggestion>%s</suggestion>", suggestion)
		}
		xml += "\n    </suggestions>"
	}
	
	xml += "\n</error>"
	
	ctx.SetContentType("application/xml; charset=utf-8")
	ctx.String(errorCtx.StatusCode, xml)
	return nil
}

// CanRender 检查是否能渲染
func (r *XMLRenderer) CanRender(ctx *mvccontext.Context) bool {
	accept := string(ctx.GetHeader("Accept"))
	
	// 明确要求XML且不要HTML
	return (strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml")) &&
		!strings.Contains(accept, "text/html")
}

// ContentType 返回内容类型
func (r *XMLRenderer) ContentType() string {
	return "application/xml"
}

// ============= 纯文本渲染器 =============

// TextRenderer 纯文本错误渲染器
type TextRenderer struct {
	showDetails bool
}

// NewTextRenderer 创建文本渲染器
func NewTextRenderer(showDetails bool) *TextRenderer {
	return &TextRenderer{
		showDetails: showDetails,
	}
}

// Render 渲染文本响应
func (r *TextRenderer) Render(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	text := fmt.Sprintf("Error %d: %s\n\nMessage: %s\nPath: %s\nMethod: %s\nTime: %s",
		errorCtx.StatusCode,
		errorCtx.StatusText,
		errorCtx.ErrorMessage,
		errorCtx.RequestPath,
		errorCtx.RequestMethod,
		errorCtx.Timestamp.Format("2006-01-02 15:04:05"),
	)
	
	if r.showDetails {
		if errorCtx.RequestID != "" {
			text += fmt.Sprintf("\nRequest ID: %s", errorCtx.RequestID)
		}
		
		if len(errorCtx.Suggestions) > 0 {
			text += "\n\nSuggestions:"
			for i, suggestion := range errorCtx.Suggestions {
				text += fmt.Sprintf("\n%d. %s", i+1, suggestion)
			}
		}
	}
	
	ctx.SetContentType("text/plain; charset=utf-8")
	ctx.String(errorCtx.StatusCode, text)
	return nil
}

// CanRender 检查是否能渲染
func (r *TextRenderer) CanRender(ctx *mvccontext.Context) bool {
	accept := string(ctx.GetHeader("Accept"))
	return strings.Contains(accept, "text/plain")
}

// ContentType 返回内容类型
func (r *TextRenderer) ContentType() string {
	return "text/plain"
}

// ============= 渲染器工厂 =============

// DefaultRendererFactory 默认渲染器工厂
type DefaultRendererFactory struct {
	templateManager TemplateManager
}

// NewDefaultRendererFactory 创建默认渲染器工厂
func NewDefaultRendererFactory(templateManager TemplateManager) *DefaultRendererFactory {
	return &DefaultRendererFactory{
		templateManager: templateManager,
	}
}

// CreateRenderer 创建渲染器
func (f *DefaultRendererFactory) CreateRenderer(contentType string) ErrorRenderer {
	switch contentType {
	case "application/json":
		return NewJSONRenderer(true, false)
	case "text/html":
		return NewHTMLRenderer(f.templateManager)
	case "application/xml", "text/xml":
		return NewXMLRenderer(false)
	case "text/plain":
		return NewTextRenderer(true)
	default:
		// 默认返回JSON渲染器
		return NewJSONRenderer(true, false)
	}
}

// GetAvailableRenderers 获取可用渲染器列表
func (f *DefaultRendererFactory) GetAvailableRenderers() []string {
	return []string{
		"application/json",
		"text/html",
		"application/xml",
		"text/xml",
		"text/plain",
	}
}

// ============= 多渲染器 =============

// MultiRenderer 多渲染器（根据内容协商自动选择）
type MultiRenderer struct {
	renderers       []ErrorRenderer
	defaultRenderer ErrorRenderer
}

// NewMultiRenderer 创建多渲染器
func NewMultiRenderer() *MultiRenderer {
	return &MultiRenderer{
		renderers: make([]ErrorRenderer, 0, 4),
	}
}

// AddRenderer 添加渲染器
func (m *MultiRenderer) AddRenderer(renderer ErrorRenderer) {
	m.renderers = append(m.renderers, renderer)
}

// SetDefaultRenderer 设置默认渲染器
func (m *MultiRenderer) SetDefaultRenderer(renderer ErrorRenderer) {
	m.defaultRenderer = renderer
}

// Render 渲染响应
func (m *MultiRenderer) Render(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	// 尝试找到匹配的渲染器
	for _, renderer := range m.renderers {
		if renderer.CanRender(ctx) {
			return renderer.Render(ctx, errorCtx)
		}
	}
	
	// 使用默认渲染器
	if m.defaultRenderer != nil {
		return m.defaultRenderer.Render(ctx, errorCtx)
	}
	
	// 最后的后备方案
	fallbackRenderer := NewJSONRenderer(true, false)
	return fallbackRenderer.Render(ctx, errorCtx)
}

// CanRender 检查是否能渲染
func (m *MultiRenderer) CanRender(ctx *mvccontext.Context) bool {
	// 多渲染器总是可以渲染
	return true
}

// ContentType 返回内容类型
func (m *MultiRenderer) ContentType() string {
	// 返回默认内容类型
	if m.defaultRenderer != nil {
		return m.defaultRenderer.ContentType()
	}
	return "application/json"
}

// ============= 全局渲染器工厂 =============

var globalRendererFactory = NewDefaultRendererFactory(GetGlobalTemplateManager())

// GetGlobalRendererFactory 获取全局渲染器工厂
func GetGlobalRendererFactory() RendererFactory {
	return globalRendererFactory
}

// CreateDefaultMultiRenderer 创建默认多渲染器
func CreateDefaultMultiRenderer() *MultiRenderer {
	multi := NewMultiRenderer()
	
	// 添加各种渲染器
	multi.AddRenderer(NewJSONRenderer(true, false))
	multi.AddRenderer(NewHTMLRenderer(GetGlobalTemplateManager()))
	multi.AddRenderer(NewXMLRenderer(false))
	multi.AddRenderer(NewTextRenderer(true))
	
	// 设置JSON为默认渲染器
	multi.SetDefaultRenderer(NewJSONRenderer(true, false))
	
	return multi
}