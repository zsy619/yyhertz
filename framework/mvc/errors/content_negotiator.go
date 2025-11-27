package errors

import (
	"fmt"
	"strings"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// =============================================================================
// 模块：内容协商和渲染
// 职责：根据请求类型选择合适的响应格式（HTML/JSON/XML）
// =============================================================================

// isAPIRequest 判断是否为API请求
func IsAPIRequest(ctx *mvccontext.Context) bool {
	path := string(ctx.Path())
	accept := string(ctx.GetHeader("Accept"))

	// 路径以 /api/ 开头的明确是API请求
	if strings.HasPrefix(path, "/api/") {
		return true
	}

	// Accept头明确表示只要JSON，不要HTML
	if (strings.Contains(accept, "application/json") ||
		strings.Contains(accept, "application/vnd.api+json")) &&
		!strings.Contains(accept, "text/html") {
		return true
	}

	return false
}

// isXMLRequest 判断是否请求XML格式
func IsXMLRequest(ctx *mvccontext.Context) bool {
	accept := string(ctx.GetHeader("Accept"))

	// 如果同时包含HTML和XML，优先选择HTML
	if strings.Contains(accept, "text/html") {
		return false
	}

	// 只有明确请求XML且不包含HTML时才返回XML
	return (strings.Contains(accept, "application/xml") ||
		strings.Contains(accept, "text/xml")) &&
		!strings.Contains(accept, "text/html")
}

// renderJSONError 渲染JSON格式错误响应
func RenderJSONError(ctx *mvccontext.Context, errorCtx *ErrorContext, showDetailedError bool) error {
	response := map[string]any{
		"code":      errorCtx.StatusCode,
		"message":   errorCtx.ErrorMessage,
		"success":   false,
		"path":      errorCtx.RequestPath,
		"method":    errorCtx.RequestMethod,
		"timestamp": errorCtx.Timestamp.Unix(),
	}

	// 添加额外的上下文信息
	if showDetailedError {
		response["details"] = errorCtx.Details
		response["suggestions"] = errorCtx.Suggestions
		response["request_id"] = errorCtx.RequestID
	}

	ctx.JSON(errorCtx.StatusCode, response)
	return nil
}

// renderXMLError 渲染XML格式错误响应
func RenderXMLError(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	xmlData, xmlErr := TplFS.ReadFile("templates/error.xml")
	if xmlErr != nil {
		fmt.Println("Error reading error template:", xmlErr.Error())
		return xmlErr
	}
	xmlTemplate := string(xmlData)
	xmlResponse := fmt.Sprintf(xmlTemplate,
		errorCtx.StatusCode,
		errorCtx.StatusText,
		errorCtx.ErrorMessage,
		errorCtx.RequestPath,
		errorCtx.RequestMethod,
		errorCtx.Timestamp.Format(time.RFC3339),
		errorCtx.RequestID,
	)

	ctx.SetContentType("application/xml")
	ctx.String(errorCtx.StatusCode, xmlResponse)
	return nil
}

// renderHTMLError 渲染HTML格式错误页面
func RenderHTMLError(ctx *mvccontext.Context, errorCtx *ErrorContext, customTitle, supportEmail, supportPhone string, enableDebug bool) error {
	// 生成HTML页面
	html := GenerateErrorHTML(errorCtx, customTitle, supportEmail, supportPhone, enableDebug)

	// 使用Data方法直接输出HTML内容
	ctx.Data(errorCtx.StatusCode, "text/html; charset=utf-8", []byte(html))
	return nil
}

// HandleErrorResponse 根据内容协商处理错误响应
func HandleErrorResponse(ctx *mvccontext.Context, errorCtx *ErrorContext, showDetailedError bool, customTitle, supportEmail, supportPhone string, enableDebug bool) error {
	// 判断请求类型并选择合适的响应格式
	if IsAPIRequest(ctx) {
		return RenderJSONError(ctx, errorCtx, showDetailedError)
	}

	// 检查是否请求XML格式
	if IsXMLRequest(ctx) {
		return RenderXMLError(ctx, errorCtx)
	}

	// 默认渲染HTML页面
	return RenderHTMLError(ctx, errorCtx, customTitle, supportEmail, supportPhone, enableDebug)
}