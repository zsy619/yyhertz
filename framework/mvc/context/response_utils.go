package context

import (
	"strconv"
	"strings"
)

// ============= 响应便捷方法 =============

// SendStatus 仅发送HTTP状态码，不包含响应体
// 这是一个便捷方法，用于快速返回状态码响应
func (ctx *Context) SendStatus(code int) {
	if !ctx.ensureRequest() {
		return
	}
	
	ctx.Status(code)
	
	// 对于某些状态码，可以添加标准的状态文本
	switch code {
	case StatusNotFound:
		ctx.String(code, "Not Found")
	case StatusInternalServerError:
		ctx.String(code, "Internal Server Error")
	case StatusBadRequest:
		ctx.String(code, "Bad Request")
	case StatusUnauthorized:
		ctx.String(code, "Unauthorized")
	case StatusForbidden:
		ctx.String(code, "Forbidden")
	case StatusMethodNotAllowed:
		ctx.String(code, "Method Not Allowed")
	case StatusNotImplemented:
		ctx.String(code, "Not Implemented")
	case StatusBadGateway:
		ctx.String(code, "Bad Gateway")
	case StatusServiceUnavailable:
		ctx.String(code, "Service Unavailable")
	default:
		// 对于其他状态码，发送空响应
		ctx.String(code, "")
	}
}

// End 结束响应，不发送任何内容
// 这个方法用于结束响应而不发送任何数据，常用于204 No Content等场景
func (ctx *Context) End() {
	if !ctx.ensureRequest() {
		return
	}
	
	// 设置Content-Length为0
	ctx.SetHeader("Content-Length", "0")
	
	// 如果没有设置状态码，使用200
	if ctx.getStatusCode() == 0 {
		ctx.Status(StatusOK)
	}
}

// Location 设置Location头部，常用于重定向
func (ctx *Context) Location(url string) {
	if !ctx.ensureRequest() {
		return
	}
	
	ctx.SetHeader("Location", url)
}

// Vary 设置Vary头部，用于缓存控制
// Vary头指示哪些请求头会影响响应内容的选择
func (ctx *Context) Vary(fields ...string) {
	if !ctx.ensureRequest() || len(fields) == 0 {
		return
	}
	
	// 获取现有的Vary头
	existing := ctx.GetResponseHeader("Vary")
	
	// 合并新的字段
	var varyFields []string
	if existing != "" {
		varyFields = append(varyFields, existing)
	}
	
	for _, field := range fields {
		if field != "" {
			varyFields = append(varyFields, field)
		}
	}
	
	// 设置Vary头
	if len(varyFields) > 0 {
		ctx.SetHeader("Vary", strings.Join(varyFields, ", "))
	}
}

// Links 设置Link头部，用于指示相关资源的关系
func (ctx *Context) Links(links map[string]string) {
	if !ctx.ensureRequest() || len(links) == 0 {
		return
	}
	
	var linkParts []string
	for url, rel := range links {
		if url != "" && rel != "" {
			linkParts = append(linkParts, "<"+url+">; rel=\""+rel+"\"")
		}
	}
	
	if len(linkParts) > 0 {
		ctx.SetHeader("Link", strings.Join(linkParts, ", "))
	}
}

// Type 设置Content-Type头部（SetContentType的别名）
func (ctx *Context) Type(contentType string) {
	ctx.SetContentType(contentType)
}

// Append 向现有头部添加值，如果头部不存在则创建
func (ctx *Context) Append(name, value string) {
	if !ctx.ensureRequest() || name == "" || value == "" {
		return
	}
	
	existing := ctx.GetResponseHeader(name)
	if existing == "" {
		ctx.SetHeader(name, value)
	} else {
		ctx.SetHeader(name, existing+", "+value)
	}
}

// ============= 条件响应方法 =============

// Fresh 检查请求是否新鲜（即客户端缓存是否仍然有效）
// 基于Last-Modified和ETag进行检查
func (ctx *Context) Fresh() bool {
	if !ctx.ensureRequest() {
		return false
	}
	
	// 获取响应头中的缓存相关信息
	lastModified := ctx.GetResponseHeader("Last-Modified")
	etag := ctx.GetResponseHeader("ETag")
	
	// 如果没有缓存头，则认为不新鲜
	if lastModified == "" && etag == "" {
		return false
	}
	
	// 检查If-None-Match (ETag)
	if etag != "" {
		ifNoneMatch := ctx.Header("If-None-Match")
		if ifNoneMatch != "" {
			if ifNoneMatch == "*" || strings.Contains(ifNoneMatch, etag) {
				return true
			}
		}
	}
	
	// 检查If-Modified-Since
	if lastModified != "" {
		ifModifiedSince := ctx.Header("If-Modified-Since")
		if ifModifiedSince != "" && ifModifiedSince == lastModified {
			return true
		}
	}
	
	return false
}

// Stale 检查请求是否过期（Fresh的反义）
func (ctx *Context) Stale() bool {
	return !ctx.Fresh()
}

// ============= 重定向增强方法 =============

// RedirectBack 重定向到前一个页面，如果没有Referer则重定向到fallback
func (ctx *Context) RedirectBack(fallback string) {
	if !ctx.ensureRequest() {
		return
	}
	
	referer := ctx.Header("Referer")
	if referer == "" {
		referer = fallback
	}
	
	ctx.Redirect(StatusFound, referer)
}

// RedirectWithQuery 重定向并保持查询参数
func (ctx *Context) RedirectWithQuery(code int, url string) {
	if !ctx.ensureRequest() {
		return
	}
	
	// 获取当前请求的查询参数
	queryString := ""
	if ctx.ensureRequest() {
		uri := ctx.request.URI()
		queryString = safeStringConvert(uri.QueryString())
	}
	
	// 如果有查询参数，则添加到重定向URL
	if queryString != "" {
		separator := "?"
		if strings.Contains(url, "?") {
			separator = "&"
		}
		url = url + separator + queryString
	}
	
	ctx.Redirect(code, url)
}

// ============= 缓存控制方法 =============

// NoCache 设置不缓存头部
func (ctx *Context) NoCache() {
	if !ctx.ensureRequest() {
		return
	}
	
	ctx.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.SetHeader("Pragma", "no-cache")
	ctx.SetHeader("Expires", "0")
}

// MaxAge 设置缓存最大年龄（秒）
func (ctx *Context) MaxAge(seconds int) {
	if !ctx.ensureRequest() {
		return
	}
	
	cacheControl := "max-age=" + strconv.Itoa(seconds)
	ctx.SetHeader("Cache-Control", cacheControl)
}

// CacheControl 设置复杂的缓存控制指令
func (ctx *Context) CacheControl(directives map[string]interface{}) {
	if !ctx.ensureRequest() || len(directives) == 0 {
		return
	}
	
	var parts []string
	
	for directive, value := range directives {
		switch v := value.(type) {
		case bool:
			if v {
				parts = append(parts, directive)
			}
		case int:
			parts = append(parts, directive+"="+strconv.Itoa(v))
		case string:
			if v != "" {
				parts = append(parts, directive+"="+v)
			}
		}
	}
	
	if len(parts) > 0 {
		ctx.SetHeader("Cache-Control", strings.Join(parts, ", "))
	}
}

// ============= CORS支持方法 =============

// EnableCORS 启用基本的CORS支持
func (ctx *Context) EnableCORS(origins ...string) {
	if !ctx.ensureRequest() {
		return
	}
	
	origin := "*"
	if len(origins) > 0 && origins[0] != "" {
		origin = origins[0]
	}
	
	ctx.SetHeader("Access-Control-Allow-Origin", origin)
	ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// CORSHeaders 设置详细的CORS头部
func (ctx *Context) CORSHeaders(config map[string]string) {
	if !ctx.ensureRequest() || len(config) == 0 {
		return
	}
	
	for header, value := range config {
		if strings.HasPrefix(header, "Access-Control-") && value != "" {
			ctx.SetHeader(header, value)
		}
	}
}

// ============= 安全头部方法 =============

// SecurityHeaders 设置基本的安全响应头
func (ctx *Context) SecurityHeaders() {
	if !ctx.ensureRequest() {
		return
	}
	
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	ctx.SetHeader("X-Frame-Options", "DENY")
	ctx.SetHeader("X-XSS-Protection", "1; mode=block")
	ctx.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
}

// CSP 设置Content Security Policy
func (ctx *Context) CSP(policy string) {
	if !ctx.ensureRequest() || policy == "" {
		return
	}
	
	ctx.SetHeader("Content-Security-Policy", policy)
}

// HSTS 设置HTTP Strict Transport Security
func (ctx *Context) HSTS(maxAge int, includeSubdomains bool) {
	if !ctx.ensureRequest() {
		return
	}
	
	value := "max-age=" + strconv.Itoa(maxAge)
	if includeSubdomains {
		value += "; includeSubDomains"
	}
	
	ctx.SetHeader("Strict-Transport-Security", value)
}

// ============= 内容协商辅助方法 =============

// Format 根据Accept头返回不同格式的响应
func (ctx *Context) Format(responses map[string]func()) {
	if !ctx.ensureRequest() || len(responses) == 0 {
		return
	}
	
	accept := ctx.Header("Accept")
	
	// 简单的内容类型匹配
	for contentType, handler := range responses {
		if strings.Contains(accept, contentType) || accept == "*/*" {
			handler()
			return
		}
	}
	
	// 如果没有匹配，返回406 Not Acceptable
	ctx.SendStatus(StatusNotAcceptable)
}

// ============= 响应信息方法 =============

// HeaderSize 获取响应头大小（估算）
func (ctx *Context) HeaderSize() int {
	if !ctx.ensureRequest() {
		return 0
	}
	
	size := 0
	// 这里需要遍历所有响应头来计算大小
	// 由于Hertz的API限制，这里提供一个简化版本
	
	// 估算常见头部大小
	headers := []string{
		"Content-Type", "Content-Length", "Cache-Control",
		"Last-Modified", "ETag", "Set-Cookie",
	}
	
	for _, header := range headers {
		value := ctx.GetResponseHeader(header)
		if value != "" {
			size += len(header) + len(value) + 4 // header: value\r\n
		}
	}
	
	return size
}

// ResponseInfo 获取响应信息摘要
func (ctx *Context) ResponseInfo() map[string]interface{} {
	if !ctx.ensureRequest() {
		return nil
	}
	
	info := map[string]interface{}{
		"status_code":   ctx.getStatusCode(),
		"content_type":  ctx.GetResponseHeader("Content-Type"),
		"content_length": ctx.GetResponseHeader("Content-Length"),
		"cache_control": ctx.GetResponseHeader("Cache-Control"),
		"last_modified": ctx.GetResponseHeader("Last-Modified"),
		"etag":         ctx.GetResponseHeader("ETag"),
		"header_size":  ctx.HeaderSize(),
	}
	
	return info
}

// ============= 新增HTTP状态码常量 =============

const (
	StatusNotAcceptable = 406
)