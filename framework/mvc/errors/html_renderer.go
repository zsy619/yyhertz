package errors

import (
	"fmt"
)

// =============================================================================
// 模块：HTML渲染引擎
// 职责：生成美观的HTML错误页面，包含图标、样式、建议等
// =============================================================================

// generateErrorHTML 生成错误页面HTML
func GenerateErrorHTML(errorCtx *ErrorContext, customTitle, supportEmail, supportPhone string, enableDebug bool) string {
	// 获取状态对应的颜色和图标
	statusClass := getStatusClass(errorCtx.StatusCode)
	if statusClass == "" {
		statusClass = "status-error"
	}

	statusIcon := getStatusIcon(errorCtx.StatusCode)
	if statusIcon == "" {
		statusIcon = "❌"
	}

	// 构建建议列表HTML
	suggestionsHTML := buildSuggestionsHTML(errorCtx.Suggestions)

	htmlData, htmlErr := TplFS.ReadFile("templates/error.html")
	if htmlErr != nil {
		fmt.Println("Error reading error template:", htmlErr.Error())
		return ""
	}

	// 构建调试信息HTML
	debugInfoHTML := buildDebugInfoHTML(errorCtx, enableDebug)
	errorPageTemplate := string(htmlData)

	// 正确的参数传递顺序
	fmt.Println(
		suggestionsHTML,
		debugInfoHTML,
		getSupportInfo(supportEmail, supportPhone))
	return fmt.Sprintf(errorPageTemplate, errorCtx.StatusCode, errorCtx.StatusCode, errorCtx.StatusCode)
	// return fmt.Sprintf(errorPageTemplate,
	// 	errorCtx.StatusCode,
	// 	errorCtx.StatusCode,
	// 	errorCtx.StatusCode,
	// 	suggestionsHTML,
	// 	debugInfoHTML,
	// 	getSupportInfo(supportEmail, supportPhone),
	// )
	// return fmt.Sprintf(errorPageTemplate,
	// 	customTitle,                    // 1. 页面标题
	// 	"YYHertz Framework",            // 2. header h1
	// 	statusClass,                    // 3. 状态CSS类
	// 	statusIcon,                     // 4. 状态图标
	// 	errorCtx.StatusCode,            // 5. 状态码
	// 	errorCtx.StatusText,            // 6. 状态文本
	// 	errorCtx.StatusText,            // 7. 简短状态描述
	// 	errorCtx.ErrorMessage,          // 8. 详细错误消息
	// 	errorCtx.RequestPath,           // 9. 请求路径
	// 	errorCtx.RequestMethod,         // 10. 请求方法
	// 	errorCtx.Timestamp.Format("2006-01-02 15:04:05"), // 11. 时间戳
	// 	suggestionsHTML,                // 12. 建议列表
	// 	debugInfoHTML,                  // 13. 调试信息
	// 	getSupportInfo(supportEmail, supportPhone), // 14. 支持信息
	// 	errorCtx.StatusCode,            // 15. JavaScript中的状态码1
	// 	errorCtx.RequestPath,           // 16. JavaScript中的请求路径
	// 	errorCtx.RequestMethod,         // 17. JavaScript中的请求方法
	// 	errorCtx.StatusCode,            // 18. JavaScript中的状态码2
	// 	errorCtx.StatusCode,            // 19. JavaScript中的状态码3
	// )
}

// getStatusClass 获取状态对应的CSS类
func getStatusClass(statusCode int) string {
	switch {
	case statusCode == 400:
		return "status-400"
	case statusCode == 401:
		return "status-401"
	case statusCode == 402:
		return "status-402"
	case statusCode == 403:
		return "status-403"
	case statusCode == 404:
		return "status-404"
	case statusCode == 405:
		return "status-405"
	case statusCode == 406:
		return "status-406"
	case statusCode == 408:
		return "status-408"
	case statusCode == 409:
		return "status-409"
	case statusCode == 410:
		return "status-410"
	case statusCode == 413:
		return "status-413"
	case statusCode == 415:
		return "status-415"
	case statusCode == 418:
		return "status-418"
	case statusCode == 422:
		return "status-422"
	case statusCode == 429:
		return "status-429"
	case statusCode == 500:
		return "status-500"
	case statusCode == 501:
		return "status-501"
	case statusCode == 502:
		return "status-502"
	case statusCode == 503:
		return "status-503"
	case statusCode == 504:
		return "status-504"
	case statusCode == 505:
		return "status-505"
	case statusCode >= 400 && statusCode < 500:
		return "status-4xx"
	case statusCode >= 500:
		return "status-5xx"
	default:
		return "status-error"
	}
}

// getStatusIcon 获取状态对应的图标
func getStatusIcon(statusCode int) string {
	switch statusCode {
	case 400:
		return "❓" // Bad Request
	case 401:
		return "🔐" // Unauthorized
	case 402:
		return "💳" // Payment Required
	case 403:
		return "🚫" // Forbidden
	case 404:
		return "🔍" // Not Found
	case 405:
		return "🚷" // Method Not Allowed
	case 406:
		return "📋" // Not Acceptable
	case 408:
		return "⏰" // Request Timeout
	case 409:
		return "⚔️" // Conflict
	case 410:
		return "📱" // Gone
	case 413:
		return "📦" // Payload Too Large
	case 415:
		return "📄" // Unsupported Media Type
	case 418:
		return "🫖" // I'm a teapot
	case 422:
		return "📝" // Unprocessable Entity
	case 429:
		return "🚦" // Too Many Requests
	case 500:
		return "⚠️" // Internal Server Error
	case 501:
		return "🚧" // Not Implemented
	case 502:
		return "🌐" // Bad Gateway
	case 503:
		return "🔧" // Service Unavailable
	case 504:
		return "⏳" // Gateway Timeout
	case 505:
		return "📡" // HTTP Version Not Supported
	default:
		return "❌" // Generic Error
	}
}

// buildSuggestionsHTML 构建建议列表HTML
func buildSuggestionsHTML(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}

	html := "<div><h3>解决建议</h3><ul class=\"suggestions\">"
	for _, suggestion := range suggestions {
		html += fmt.Sprintf("<li>%s</li>", suggestion)
	}
	html += "</ul></div>"
	return html
}

// buildDebugInfoHTML 构建调试信息HTML
func buildDebugInfoHTML(errorCtx *ErrorContext, enableDebug bool) string {
	if !enableDebug || len(errorCtx.Details) == 0 {
		return ""
	}

	html := `<div class="debug-info">
		<h3>调试信息 <span class="debug-label">开发环境</span></h3>
		<div class="debug-grid">`

	for key, value := range errorCtx.Details {
		html += fmt.Sprintf(`
			<div class="debug-item">
				<div class="debug-key">%s</div>
				<div class="debug-value">%v</div>
			</div>`, key, value)
	}

	html += "</div></div>"
	return html
}

// getSupportInfo 获取支持信息
func getSupportInfo(supportEmail, supportPhone string) string {
	if supportEmail == "" && supportPhone == "" {
		return ""
	}

	html := `<div class="support-info">
		<h3>需要帮助？</h3>
		<div class="contact-info">`

	if supportEmail != "" {
		html += fmt.Sprintf(`<div class="contact-item">
			<span class="contact-icon">📧</span>
			<a href="mailto:%s">%s</a>
		</div>`, supportEmail, supportEmail)
	}

	if supportPhone != "" {
		html += fmt.Sprintf(`<div class="contact-item">
			<span class="contact-icon">📞</span>
			<a href="tel:%s">%s</a>
		</div>`, supportPhone, supportPhone)
	}

	html += "</div></div>"
	return html
}
