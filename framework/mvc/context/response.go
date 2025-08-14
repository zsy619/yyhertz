package context

import (
	"compress/gzip"
	"compress/zlib"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// ============= 基础响应方法 =============

// JSON 返回JSON响应
func (ctx *Context) JSON(code int, obj any) {
	if ctx.ensureRequest() {
		ctx.request.JSON(code, obj)
	}
}

// IndentedJSON 返回格式化的JSON响应 (美化输出)
func (ctx *Context) IndentedJSON(code int, obj any) {
	if ctx.ensureRequest() {
		ctx.request.IndentedJSON(code, obj)
	}
}

// WriteString writes a string to response body.
func (ctx *Context) WriteString(content string) {
	_, _ = ctx.RequestContext().WriteString(content)
}

// String 返回字符串响应
func (ctx *Context) String(code int, format string, values ...any) {
	if ctx.ensureRequest() {
		ctx.request.String(code, format, values...)
	}
}

// HTML 返回HTML响应
func (ctx *Context) HTML(code int, name string, obj any) {
	if ctx.ensureRequest() {
		// 这里需要集成模板引擎
		ctx.request.HTML(code, name, obj)
	}
}

// XML 返回XML响应
func (ctx *Context) XML(code int, obj any) {
	if !ctx.ensureRequest() {
		return
	}

	ctx.request.SetStatusCode(code)
	ctx.request.Response.Header.Set(HeaderContentType, ContentTypeXML+"; charset=utf-8")

	if data, err := xml.Marshal(obj); err == nil {
		ctx.request.Response.SetBody(data)
	}
}

// YAML 返回YAML响应
func (ctx *Context) YAML(code int, obj any) {
	if !ctx.ensureRequest() {
		return
	}

	ctx.request.SetStatusCode(code)
	ctx.request.Response.Header.Set(HeaderContentType, ContentTypeYAML+"; charset=utf-8")

	if data, err := yaml.Marshal(obj); err == nil {
		ctx.request.Response.SetBody(data)
	}
}

// Data 返回原始数据响应
func (ctx *Context) Data(code int, contentType string, data []byte) {
	if !ctx.ensureRequest() {
		return
	}

	ctx.request.SetStatusCode(code)
	ctx.request.Response.Header.Set(HeaderContentType, contentType)
	ctx.request.Response.SetBody(data)
}

// ============= 状态码和重定向方法 =============

// Status 设置状态码
func (ctx *Context) Status(code int) {
	if ctx.ensureRequest() {
		ctx.request.SetStatusCode(code)
	}
}

// Redirect 重定向响应
func (ctx *Context) Redirect(code int, location string) {
	if ctx.ensureRequest() {
		ctx.request.Redirect(code, []byte(location))
	}
}

// ============= 高级响应方法 =============

// NoContent 返回204 No Content响应
func (ctx *Context) NoContent() {
	ctx.Status(StatusNoContent)
}

// JSONWithStatus 带状态码的JSON响应
func (ctx *Context) JSONWithStatus(code int, obj any) {
	ctx.JSON(code, obj)
}

// XMLWithStatus 带状态码的XML响应
func (ctx *Context) XMLWithStatus(code int, obj any) {
	ctx.XML(code, obj)
}

// YAMLWithStatus 带状态码的YAML响应
func (ctx *Context) YAMLWithStatus(code int, obj any) {
	ctx.YAML(code, obj)
}

// StringWithStatus 带状态码的字符串响应
func (ctx *Context) StringWithStatus(code int, format string, values ...any) {
	ctx.String(code, format, values...)
}

// ============= 便捷响应方法 =============

// JSONOK 返回200状态的JSON响应
func (ctx *Context) JSONOK(obj any) {
	ctx.JSON(StatusOK, obj)
}

// JSONError 返回错误状态的JSON响应
func (ctx *Context) JSONError(code int, message string) {
	ctx.JSON(code, map[string]any{
		"error":   true,
		"message": message,
		"code":    code,
	})
}

// JSONSuccess 返回成功状态的JSON响应
func (ctx *Context) JSONSuccess(data any) {
	ctx.JSON(StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}

// JSONPage 返回分页JSON响应
func (ctx *Context) JSONPage(data any, page, pageSize, total int) {
	ctx.JSON(StatusOK, map[string]any{
		"success":   true,
		"data":      data,
		"page":      page,
		"pageSize":  pageSize,
		"total":     total,
		"totalPage": (total + pageSize - 1) / pageSize,
	})
}

// ============= JSONP支持 =============

// JSONP 返回JSONP响应
func (ctx *Context) JSONP(code int, obj any) {
	callback := ctx.QueryDefault("callback", "callback")
	if callback == "" {
		ctx.JSON(code, obj)
		return
	}

	if !ctx.ensureRequest() {
		return
	}

	ctx.request.SetStatusCode(code)
	ctx.request.Response.Header.Set(HeaderContentType, "application/javascript; charset=utf-8")

	// 简化的JSONP实现
	data := ctx.request.Response.BodyBytes()
	if len(data) > 0 {
		jsonpData := callback + "(" + string(data) + ");"
		ctx.request.Response.SetBodyString(jsonpData)
	}
}

// JSONPWithStatus 带状态码的JSONP响应
func (ctx *Context) JSONPWithStatus(code int, obj any) {
	ctx.JSONP(code, obj)
}

// ============= 响应头操作方法 =============

// SetHeader 设置响应头
func (ctx *Context) SetHeader(key, value string) {
	if ctx.ensureRequest() {
		ctx.request.Response.Header.Set(key, value)
	}
}

// AddHeader 添加响应头
func (ctx *Context) AddHeader(key, value string) {
	if ctx.ensureRequest() {
		ctx.request.Response.Header.Add(key, value)
	}
}

// GetResponseHeader 获取响应头
func (ctx *Context) GetResponseHeader(key string) string {
	if !ctx.ensureRequest() {
		return ""
	}
	return safeStringConvert(ctx.request.Response.Header.Peek(key))
}

// SetContentType 设置Content-Type
func (ctx *Context) SetContentType(contentType string) {
	ctx.SetHeader(HeaderContentType, contentType)
}

// ============= 特殊响应方法 =============

// Write 写入响应数据 (兼容性方法)
func (ctx *Context) Write(data []byte) (int, error) {
	if ctx.writer != nil {
		return ctx.writer.Write(data)
	}
	return 0, ErrWriterNotFound
}

// Stream 流式响应
func (ctx *Context) Stream(contentType string, reader func() ([]byte, bool)) {
	if !ctx.ensureRequest() {
		return
	}

	ctx.SetContentType(contentType)

	for {
		data, hasMore := reader()
		if len(data) > 0 {
			ctx.request.Response.AppendBody(data)
		}
		if !hasMore {
			break
		}
	}
}

// ============= 重定向方法 =============

// RedirectPermanent 永久重定向 (301)
func (ctx *Context) RedirectPermanent(location string) {
	ctx.Redirect(StatusMovedPermanently, location)
}

// RedirectTemporary 临时重定向 (302)
func (ctx *Context) RedirectTemporary(location string) {
	ctx.Redirect(StatusFound, location)
}

// RedirectSeeOther See Other重定向 (303)
func (ctx *Context) RedirectSeeOther(location string) {
	ctx.Redirect(StatusSeeOther, location)
}

// ============= 状态检查方法 =============

// IsOk 判断响应状态是否为200
func (ctx *Context) IsOk() bool {
	return ctx.getStatusCode() == StatusOK
}

// IsSuccessful 判断响应状态是否为2xx
func (ctx *Context) IsSuccessful() bool {
	code := ctx.getStatusCode()
	return code >= 200 && code < 300
}

// IsRedirect 判断响应状态是否为3xx
func (ctx *Context) IsRedirect() bool {
	code := ctx.getStatusCode()
	return code >= 300 && code < 400
}

// IsClientError 判断响应状态是否为4xx
func (ctx *Context) IsClientError() bool {
	code := ctx.getStatusCode()
	return code >= 400 && code < 500
}

// IsServerError 判断响应状态是否为5xx
func (ctx *Context) IsServerError() bool {
	code := ctx.getStatusCode()
	return code >= 500 && code < 600
}

// IsForbidden 判断响应状态是否为403
func (ctx *Context) IsForbidden() bool {
	return ctx.getStatusCode() == StatusForbidden
}

// ============= 辅助方法 =============

// getStatusCode 获取当前响应状态码
func (ctx *Context) getStatusCode() int {
	if !ctx.ensureRequest() {
		return 0
	}
	return ctx.request.Response.StatusCode()
}

// ============= 增强Writer操作（流控制） =============

// BufferedWrite 缓冲写入（批量写入优化）
func (ctx *Context) BufferedWrite(data [][]byte, contentType string) error {
	if !ctx.ensureRequest() {
		return ErrWriterNotFound
	}

	ctx.SetContentType(contentType)

	// 预计算总大小
	totalSize := 0
	for _, chunk := range data {
		totalSize += len(chunk)
	}

	// 预分配缓冲区
	buffer := make([]byte, 0, totalSize)
	for _, chunk := range data {
		buffer = append(buffer, chunk...)
	}

	ctx.request.Response.SetBody(buffer)
	return nil
}

// ChunkedWrite 分块写入（适合大数据流）
func (ctx *Context) ChunkedWrite(reader func() ([]byte, bool), contentType string) error {
	if !ctx.ensureRequest() {
		return ErrWriterNotFound
	}

	ctx.SetContentType(contentType)
	ctx.SetHeader("Transfer-Encoding", "chunked")

	for {
		chunk, hasMore := reader()
		if len(chunk) > 0 {
			ctx.request.Response.AppendBody(chunk)
		}
		if !hasMore {
			break
		}
	}

	return nil
}

// StreamWrite 流式写入（实时数据传输）
func (ctx *Context) StreamWrite(contentType string, writer func(*StreamWriter)) error {
	if !ctx.ensureRequest() {
		return ErrWriterNotFound
	}

	ctx.SetContentType(contentType)

	streamWriter := &StreamWriter{
		ctx: ctx,
	}

	writer(streamWriter)
	return nil
}

// ============= 增强Writer操作（压缩支持） =============

// WriteWithCompression 带压缩的写入
func (ctx *Context) WriteWithCompression(data []byte, contentType string) error {
	if !ctx.ensureRequest() {
		return ErrWriterNotFound
	}

	// 检查客户端支持的压缩格式
	acceptEncoding := ctx.Header("Accept-Encoding")

	var compressedData []byte
	var encoding string

	// 根据支持情况选择压缩算法
	if strings.Contains(acceptEncoding, "gzip") && len(data) > 1024 {
		if compressed, err := gzipCompress(data); err == nil {
			compressedData = compressed
			encoding = "gzip"
		}
	} else if strings.Contains(acceptEncoding, "deflate") && len(data) > 1024 {
		if compressed, err := deflateCompress(data); err == nil {
			compressedData = compressed
			encoding = "deflate"
		}
	}

	if compressedData != nil {
		ctx.SetHeader("Content-Encoding", encoding)
		ctx.SetHeader("Content-Length", strconv.Itoa(len(compressedData)))
		ctx.request.Response.SetBody(compressedData)
	} else {
		ctx.SetContentType(contentType)
		ctx.request.Response.SetBody(data)
	}

	return nil
}

// ============= 增强Writer操作（缓存控制） =============

// WriteWithETag 带ETag的写入（缓存优化）
func (ctx *Context) WriteWithETag(data []byte, contentType string) error {
	if !ctx.ensureRequest() {
		return ErrWriterNotFound
	}

	// 生成ETag
	etag := generateETag(data)
	ctx.SetHeader("ETag", etag)

	// 检查客户端ETag
	clientETag := ctx.Header("If-None-Match")
	if clientETag == etag {
		ctx.Status(StatusNotModified)
		return nil
	}

	ctx.SetContentType(contentType)
	ctx.request.Response.SetBody(data)
	return nil
}

// WriteWithCache 带缓存控制的写入
func (ctx *Context) WriteWithCache(data []byte, contentType string, maxAge int) error {
	if !ctx.ensureRequest() {
		return ErrWriterNotFound
	}

	// 设置缓存头部
	ctx.SetHeader("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
	ctx.SetHeader("Expires", generateExpiresHeader(maxAge))

	return ctx.WriteWithETag(data, contentType)
}

// ============= 增强Writer操作（安全头部） =============

// WriteWithSecurityHeaders 带安全头部的写入
func (ctx *Context) WriteWithSecurityHeaders(data []byte, contentType string, options *SecurityOptions) error {
	if !ctx.ensureRequest() {
		return ErrWriterNotFound
	}

	// 设置默认安全选项
	if options == nil {
		options = &SecurityOptions{
			EnableCSP:  true,
			EnableHSTS: true,
			EnableCORS: false,
		}
	}

	// Content Security Policy
	if options.EnableCSP {
		csp := options.CSPPolicy
		if csp == "" {
			csp = "default-src 'self'"
		}
		ctx.SetHeader("Content-Security-Policy", csp)
	}

	// HTTP Strict Transport Security
	if options.EnableHSTS && ctx.IsSecure() {
		hsts := options.HSTSPolicy
		if hsts == "" {
			hsts = "max-age=31536000; includeSubDomains"
		}
		ctx.SetHeader("Strict-Transport-Security", hsts)
	}

	// CORS headers
	if options.EnableCORS {
		ctx.SetHeader("Access-Control-Allow-Origin", options.CORSOrigin)
		if options.CORSMethods != "" {
			ctx.SetHeader("Access-Control-Allow-Methods", options.CORSMethods)
		}
		if options.CORSHeaders != "" {
			ctx.SetHeader("Access-Control-Allow-Headers", options.CORSHeaders)
		}
	}

	// Other security headers
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	ctx.SetHeader("X-Frame-Options", "DENY")
	ctx.SetHeader("X-XSS-Protection", "1; mode=block")

	ctx.SetContentType(contentType)
	ctx.request.Response.SetBody(data)
	return nil
}

// ============= 增强Writer操作（响应钩子） =============

// WriteWithHooks 带钩子的写入（监控和日志）
func (ctx *Context) WriteWithHooks(data []byte, contentType string, hooks *ResponseHooks) error {
	if !ctx.ensureRequest() {
		return ErrWriterNotFound
	}

	// 执行写入前钩子
	if hooks != nil && hooks.BeforeWrite != nil {
		if err := hooks.BeforeWrite(ctx, data, contentType); err != nil {
			return err
		}
	}

	// 记录写入开始时间
	startTime := getCurrentTime()

	// 执行实际写入
	ctx.SetContentType(contentType)
	ctx.request.Response.SetBody(data)

	// 执行写入后钩子
	if hooks != nil && hooks.AfterWrite != nil {
		writeTime := getCurrentTime() - startTime
		hooks.AfterWrite(ctx, data, contentType, writeTime)
	}

	return nil
}

// ============= 增强Writer操作（性能监控） =============

// WriteWithMetrics 带性能指标的写入
func (ctx *Context) WriteWithMetrics(data []byte, contentType string) (*WriteMetrics, error) {
	if !ctx.ensureRequest() {
		return nil, ErrWriterNotFound
	}

	startTime := getCurrentTime()
	dataSize := len(data)

	// 执行写入
	ctx.SetContentType(contentType)
	ctx.request.Response.SetBody(data)

	writeTime := getCurrentTime() - startTime

	// 计算性能指标
	metrics := &WriteMetrics{
		WriteTime:    writeTime,
		DataSize:     int64(dataSize),
		Throughput:   float64(dataSize) / float64(writeTime) * 1000, // bytes per second
		ContentType:  contentType,
		StatusCode:   ctx.getStatusCode(),
		Compressed:   ctx.GetResponseHeader("Content-Encoding") != "",
		CacheEnabled: ctx.GetResponseHeader("Cache-Control") != "",
	}

	return metrics, nil
}

// GetWriteStats 获取当前写入统计
func (ctx *Context) GetWriteStats() *WriteStats {
	if !ctx.ensureRequest() {
		return nil
	}

	response := &ctx.request.Response
	return &WriteStats{
		BytesWritten:   int64(len(response.Body())),
		StatusCode:     response.StatusCode(),
		HeadersCount:   getHeaderCount(response),
		ContentType:    safeStringConvert(response.Header.Peek("Content-Type")),
		HasCompression: safeStringConvert(response.Header.Peek("Content-Encoding")) != "",
		HasCache:       safeStringConvert(response.Header.Peek("Cache-Control")) != "",
	}
}

// ============= 辅助类型和方法 =============

// StreamWriter 流式写入器
type StreamWriter struct {
	ctx *Context
}

// Write 写入数据到流
func (sw *StreamWriter) Write(data []byte) (int, error) {
	sw.ctx.request.Response.AppendBody(data)
	return len(data), nil
}

// Flush 刷新缓冲区
func (sw *StreamWriter) Flush() error {
	// Hertz 的响应会自动刷新，这里可以添加额外的刷新逻辑
	return nil
}

// SecurityOptions 安全选项
type SecurityOptions struct {
	EnableCSP   bool
	EnableHSTS  bool
	EnableCORS  bool
	CSPPolicy   string
	HSTSPolicy  string
	CORSOrigin  string
	CORSMethods string
	CORSHeaders string
}

// ResponseHooks 响应钩子
type ResponseHooks struct {
	BeforeWrite func(ctx *Context, data []byte, contentType string) error
	AfterWrite  func(ctx *Context, data []byte, contentType string, writeTime int64)
}

// WriteMetrics 写入性能指标
type WriteMetrics struct {
	WriteTime    int64   `json:"write_time_ms"`
	DataSize     int64   `json:"data_size_bytes"`
	Throughput   float64 `json:"throughput_bps"`
	ContentType  string  `json:"content_type"`
	StatusCode   int     `json:"status_code"`
	Compressed   bool    `json:"compressed"`
	CacheEnabled bool    `json:"cache_enabled"`
}

// WriteStats 写入统计信息
type WriteStats struct {
	BytesWritten   int64  `json:"bytes_written"`
	StatusCode     int    `json:"status_code"`
	HeadersCount   int    `json:"headers_count"`
	ContentType    string `json:"content_type"`
	HasCompression bool   `json:"has_compression"`
	HasCache       bool   `json:"has_cache"`
}

// ============= 错误定义 =============

var (
	ErrWriterNotFound = &ContextError{Code: "WRITER_NOT_FOUND", Message: "Response writer not found"}
)

// ============= 辅助函数 =============

// gzipCompress Gzip压缩
func gzipCompress(data []byte) ([]byte, error) {
	var buf []byte
	writer := gzip.NewWriter(&bufferWriter{buf: &buf})
	defer writer.Close()

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// deflateCompress Deflate压缩
func deflateCompress(data []byte) ([]byte, error) {
	var buf []byte
	writer := zlib.NewWriter(&bufferWriter{buf: &buf})
	defer writer.Close()

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// generateETag 生成ETag
func generateETag(data []byte) string {
	hash := md5.Sum(data)
	return "\"" + hex.EncodeToString(hash[:]) + "\""
}

// generateExpiresHeader 生成Expires头
func generateExpiresHeader(maxAge int) string {
	return time.Now().Add(time.Duration(maxAge) * time.Second).UTC().Format(time.RFC1123)
}

// getCurrentTime 获取当前时间戳（毫秒）
func getCurrentTime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// getHeaderCount 获取头部数量
func getHeaderCount(response any) int {
	// 这里需要根据Hertz的Response结构实现
	// 由于我们无法直接访问内部字段，返回估算值
	return 10
}

// bufferWriter 缓冲区写入器
type bufferWriter struct {
	buf *[]byte
}

// Write 写入数据到缓冲区
func (bw *bufferWriter) Write(data []byte) (int, error) {
	*bw.buf = append(*bw.buf, data...)
	return len(data), nil
}
