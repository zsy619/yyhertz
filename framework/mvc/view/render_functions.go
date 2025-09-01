package view

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"
)

// ============= 模板函数实现 =============

// includeTemplate 包含其他模板
func (e *TemplateEngine) includeTemplate(templateName string, data any) template.HTML {
	content, err := e.Render(templateName, data)
	if err != nil {
		return template.HTML(fmt.Sprintf("<!-- Include error: %s -->", err.Error()))
	}
	return template.HTML(content)
}

// renderComponent 渲染组件
func (e *TemplateEngine) renderComponent(componentName string, data any) template.HTML {
	content, err := e.RenderComponent(componentName, data)
	if err != nil {
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
	// 尝试通过CSRF提供者获取token
	provider := GetCSRFTokenProvider()
	if provider != nil {
		return provider.GenerateSimpleToken()
	}

	// 如果没有提供者，返回占位符
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
	for i := range result {
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