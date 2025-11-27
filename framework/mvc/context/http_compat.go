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

// ============= HTTP ResponseWriter 适配器 =============

// httpResponseWriterAdapter 将YYHertz的ResponseWriter适配为标准http.ResponseWriter
type httpResponseWriterAdapter struct {
	ctx    *Context
	header http.Header // 缓存header
	synced bool        // 是否已同步
}

// Header 返回http.Header类型的响应头
func (w *httpResponseWriterAdapter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
		w.syncFromHertz()
		w.synced = true
	}
	return w.header
}

// syncFromHertz 从Hertz同步header到本地缓存
func (w *httpResponseWriterAdapter) syncFromHertz() {
	if !w.ctx.ensureRequest() {
		return
	}

	w.ctx.request.Response.Header.VisitAll(func(key, value []byte) {
		keyStr := safeStringConvert(key)
		valueStr := safeStringConvert(value)
		w.header[keyStr] = append(w.header[keyStr], valueStr)
	})
}

// syncToHertz 将本地缓存同步到Hertz (在Write或WriteHeader前调用)
func (w *httpResponseWriterAdapter) syncToHertz() {
	if !w.ctx.ensureRequest() || !w.synced {
		return
	}

	// 清空Hertz的现有header
	keys := make([]string, 0)
	w.ctx.request.Response.Header.VisitAll(func(key, value []byte) {
		keys = append(keys, safeStringConvert(key))
	})
	for _, key := range keys {
		w.ctx.request.Response.Header.Del(key)
	}

	// 从本地缓存设置到Hertz
	for key, values := range w.header {
		for _, value := range values {
			w.ctx.request.Response.Header.Add(key, value)
		}
	}
}

// Write 写入响应数据
func (w *httpResponseWriterAdapter) Write(data []byte) (int, error) {
	if !w.ctx.ensureRequest() {
		return 0, ErrRequestNotFound
	}

	// 在写入数据前同步header
	w.syncToHertz()

	return w.ctx.writer.Write(data)
}

// WriteHeader 设置HTTP状态码
func (w *httpResponseWriterAdapter) WriteHeader(statusCode int) {
	if !w.ctx.ensureRequest() {
		return
	}

	// 在设置状态码前同步header
	w.syncToHertz()

	w.ctx.writer.WriteHeader(statusCode)
}

// HTTPResponseWriter 返回标准的http.ResponseWriter接口
// 用于与需要标准http.ResponseWriter的第三方库兼容
func (ctx *Context) HTTPResponseWriter() http.ResponseWriter {
	return &httpResponseWriterAdapter{ctx: ctx}
}

// HTTPHeader 返回标准的http.Header类型
// 用于与需要标准http.Header的第三方库兼容
// 返回的Header与Hertz ResponseHeader保持同步
func (ctx *Context) HTTPHeader() http.Header {
	if !ctx.ensureRequest() {
		return make(http.Header)
	}
	
	// 创建标准http.Header并从Hertz同步数据
	header := make(http.Header)
	ctx.request.Response.Header.VisitAll(func(key, value []byte) {
		keyStr := safeStringConvert(key)
		valueStr := safeStringConvert(value)
		header.Add(keyStr, valueStr)
	})
	
	return header
}

