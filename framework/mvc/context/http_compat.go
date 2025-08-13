package context

import (
	"net/http"
	"net/url"
	"strings"
)

// ============= HTTP标准兼容性方法 =============

// HTTPRequest 将Hertz RequestContext转换为标准http.Request
// 用于与第三方库兼容，需要标准http.Request的场景
func (ctx *Context) HTTPRequest() *http.Request {
	if !ctx.ensureRequest() {
		return nil
	}

	// 构建请求URL
	requestURL := ctx.parseURL()

	// 获取请求体
	body, _ := ctx.request.Body()
	bodyReader := strings.NewReader(string(body))

	// 创建http.Request
	req, err := http.NewRequest(safeStringConvert(ctx.request.Method()), requestURL.String(), bodyReader)
	if err != nil {
		return nil
	}

	// 复制请求头
	ctx.request.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Add(safeStringConvert(key), safeStringConvert(value))
	})

	// 设置Host
	if host := safeStringConvert(ctx.request.Host()); host != "" {
		req.Host = host
	}

	// 设置RemoteAddr (客户端地址)
	if clientIP := ctx.request.ClientIP(); clientIP != "" {
		req.RemoteAddr = clientIP + ":0" // 端口使用0作为默认值
	}

	return req
}

// parseURL 构建完整的请求URL
func (ctx *Context) parseURL() *url.URL {
	if !ctx.ensureRequest() {
		return &url.URL{}
	}

	uri := ctx.request.URI()

	// 确定协议
	scheme := "http"
	if ctx.IsSecure() {
		scheme = "https"
	}

	// 构建URL
	return &url.URL{
		Scheme:   scheme,
		Host:     safeStringConvert(ctx.request.Host()),
		Path:     safeStringConvert(uri.Path()),
		RawQuery: safeStringConvert(uri.QueryString()),
	}
}