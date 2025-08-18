// Package gin - Context高级方法实现
// 补全Context API，提供与Gin完全兼容的功能
package gin

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	
	"github.com/cloudwego/hertz/pkg/app"
)

// =============================================================================
// Cookie操作方法 (#26)
// =============================================================================

// SetCookie 设置HTTP Cookie
//
// 参数：
//   - name: Cookie名称
//   - value: Cookie值
//   - maxAge: 最大生存时间（秒），0表示会话Cookie，负数表示删除Cookie
//   - path: 路径
//   - domain: 域名
//   - secure: 是否仅HTTPS
//   - httpOnly: 是否仅HTTP（不允许JavaScript访问）
//
// 示例：
//
//	c.SetCookie("session_id", "abc123", 3600, "/", "example.com", false, true)
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	
	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	}
	
	// 设置Cookie头
	c.Header("Set-Cookie", cookie.String())
}

// Cookie 获取Cookie值
//
// 参数：
//   - name: Cookie名称
//
// 返回：
//   - string: Cookie值
//   - error: 错误信息
//
// 示例：
//
//	value, err := c.Cookie("session_id")
//	if err != nil {
//		// 处理Cookie不存在的情况
//	}
func (c *Context) Cookie(name string) (string, error) {
	cookieHeader := c.GetHeader("Cookie")
	if cookieHeader == "" {
		return "", http.ErrNoCookie
	}
	
	// 解析Cookie头
	cookies := parseCookies(cookieHeader)
	for _, cookie := range cookies {
		if cookie.Name == name {
			value, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				return cookie.Value, nil // 返回原始值
			}
			return value, nil
		}
	}
	
	return "", http.ErrNoCookie
}

// parseCookies 解析Cookie头
func parseCookies(cookieHeader string) []*http.Cookie {
	var cookies []*http.Cookie
	
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		eq := strings.Index(part, "=")
		if eq == -1 {
			continue
		}
		
		name := strings.TrimSpace(part[:eq])
		value := strings.TrimSpace(part[eq+1:])
		
		cookies = append(cookies, &http.Cookie{
			Name:  name,
			Value: value,
		})
	}
	
	return cookies
}

// =============================================================================
// 流式响应支持 (#27)
// =============================================================================

// Stream 流式响应
//
// 参数：
//   - step: 流处理函数，返回false停止流
//
// 示例：
//
//	c.Stream(func(w io.Writer) bool {
//		fmt.Fprintf(w, "data: %s\n\n", time.Now().Format(time.RFC3339))
//		return true // 继续流
//	})
func (c *Context) Stream(step func(w io.Writer) bool) {
	w := c.Writer()
	clientGone := w.CloseNotify()
	
	for {
		select {
		case <-clientGone:
			return
		default:
			if !step(w) {
				return
			}
			w.Flush()
		}
	}
}

// SSEvent 发送Server-Sent Events
//
// 参数：
//   - name: 事件名称
//   - message: 事件数据
//
// 示例：
//
//	c.Header("Content-Type", "text/event-stream")
//	c.SSEvent("message", "Hello World")
func (c *Context) SSEvent(name string, message any) {
	c.Render(-1, sse{
		Event: name,
		Data:  message,
	})
}

// sse Server-Sent Event渲染器
type sse struct {
	Event string
	Data  any
}

// WriteContentType 实现render.Render接口
func (s sse) WriteContentType(c *app.RequestContext) {
	c.Header("Content-Type", "text/event-stream")
}

// Render 实现SSE渲染
func (s sse) Render(c *app.RequestContext) error {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	
	if s.Event != "" {
		c.Write([]byte("event: " + s.Event + "\n"))
	}
	
	c.Write([]byte("data: "))
	switch v := s.Data.(type) {
	case string:
		c.Write([]byte(v))
	case []byte:
		c.Write(v)
	default:
		// 简化的JSON输出
		c.Write([]byte(fmt.Sprintf("%v", v)))
	}
	c.Write([]byte("\n\n"))
	
	return nil
}

// =============================================================================
// 文件操作增强 (#28)
// =============================================================================

// FileFromFS 从文件系统发送文件
//
// 参数：
//   - filepath: 文件路径
//   - fs: 文件系统
//
// 示例：
//
//	c.FileFromFS("assets/logo.png", http.Dir("./static"))
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.Request().URL.Path = old
	}(c.Request().URL.Path)
	
	c.Request().URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.Writer(), c.Request())
}

// FileAttachment 作为附件发送文件
//
// 参数：
//   - filepath: 文件路径
//   - filename: 下载时的文件名
//
// 示例：
//
//	c.FileAttachment("./uploads/report.pdf", "monthly_report.pdf")
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.Header("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	c.File(filepath)
}

// 注意：isASCII函数可能已在其他文件中定义

// =============================================================================
// 协商内容类型 (#29)
// =============================================================================

// Negotiate 内容协商
type Negotiate struct {
	Offered  []string
	HTMLName string
	HTMLData any
	JSONData any
	XMLData  any
	YAMLData any
	Data     any
}

// Negotiate 执行内容协商
//
// 参数：
//   - code: HTTP状态码
//   - config: 协商配置
//
// 示例：
//
//	c.Negotiate(200, gin.Negotiate{
//		Offered:  []string{"application/json", "application/xml"},
//		JSONData: gin.H{"message": "hello"},
//		XMLData:  gin.H{"message": "hello"},
//	})
func (c *Context) Negotiate(code int, config Negotiate) {
	switch c.NegotiateFormat(config.Offered...) {
	case MIMEJSON:
		data := chooseDataForContent(config.JSONData, config.Data)
		c.JSON(code, data)
		
	case MIMEHTML:
		data := chooseDataForContent(config.HTMLData, config.Data)
		c.HTML(code, config.HTMLName, data)
		
	case MIMEXML:
		data := chooseDataForContent(config.XMLData, config.Data)
		c.XML(code, data)
		
	case MIMEYAML:
		data := chooseDataForContent(config.YAMLData, config.Data)
		// 暂时使用JSON代替YAML
		c.JSON(code, data)
		
	default:
		c.AbortWithError(http.StatusNotAcceptable, errors.New("the accepted formats are not offered by the server"))
	}
}

// NegotiateFormat 协商格式
//
// 参数：
//   - offered: 提供的格式列表
//
// 返回：
//   - string: 选择的格式
//
// 示例：
//
//	format := c.NegotiateFormat("application/json", "application/xml")
func (c *Context) NegotiateFormat(offered ...string) string {
	if len(offered) == 0 {
		panic("you must provide at least one offer")
	}
	
	// 简化实现：直接解析Accept头
	accept := c.GetHeader("Accept")
	if accept == "" {
		return offered[0]
	}
	
	// 简单的匹配逻辑
	for _, offer := range offered {
		if strings.Contains(accept, offer) || strings.Contains(accept, "*/*") {
			return offer
		}
	}
	
	return offered[0]
}

// chooseDataForContent 选择内容数据
func chooseDataForContent(source, backup any) any {
	if source == nil {
		return backup
	}
	return source
}

// MIME类型常量
const (
	MIMEJSON = "application/json"
	MIMEHTML = "text/html"
	MIMEXML  = "application/xml"
	MIMEYAML = "application/x-yaml"
)

// =============================================================================
// 表单文件处理 (#30)
// =============================================================================

// FormFile 获取上传的文件
//
// 参数：
//   - name: 表单字段名
//
// 返回：
//   - *multipart.FileHeader: 文件头信息
//   - error: 错误信息
//
// 示例：
//
//	file, err := c.FormFile("upload")
//	if err != nil {
//		c.String(400, "文件上传失败: %s", err.Error())
//		return
//	}
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Request().MultipartForm == nil {
		if err := c.Request().ParseMultipartForm(c.engine.MaxMultipartMemory); err != nil {
			return nil, err
		}
	}
	
	f, fh, err := c.Request().FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	
	return fh, nil
}

// MultipartForm 获取multipart表单
//
// 返回：
//   - *multipart.Form: multipart表单
//   - error: 错误信息
//
// 示例：
//
//	form, err := c.MultipartForm()
//	if err != nil {
//		c.String(400, "解析表单失败: %s", err.Error())
//		return
//	}
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request().ParseMultipartForm(c.engine.MaxMultipartMemory)
	return c.Request().MultipartForm, err
}

// SaveUploadedFile 保存上传的文件
//
// 参数：
//   - file: 文件头信息
//   - dst: 目标路径
//
// 返回：
//   - error: 错误信息
//
// 示例：
//
//	file, _ := c.FormFile("upload")
//	err := c.SaveUploadedFile(file, "./uploads/" + file.Filename)
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	
	_, err = io.Copy(out, src)
	return err
}

// =============================================================================
// 请求体处理 (#31)
// =============================================================================

// GetRawData 获取原始请求体数据
//
// 返回：
//   - []byte: 请求体数据
//   - error: 错误信息
//
// 示例：
//
//	data, err := c.GetRawData()
//	if err != nil {
//		c.String(400, "读取请求体失败")
//		return
//	}
func (c *Context) GetRawData() ([]byte, error) {
	return io.ReadAll(c.Request().Body)
}

// ShouldBindBodyWith 使用指定绑定器绑定请求体
//
// 参数：
//   - obj: 目标对象
//   - bb: 绑定器
//
// 返回：
//   - error: 绑定错误
//
// 示例：
//
//	var user User
//	if err := c.ShouldBindBodyWith(&user, binding.JSON); err != nil {
//		c.JSON(400, gin.H{"error": err.Error()})
//		return
//	}
func (c *Context) ShouldBindBodyWith(obj any, bb BindingBody) (err error) {
	var body []byte
	if cb, ok := c.Get(BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	
	if body == nil {
		body, err = io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		c.Set(BodyBytesKey, body)
	}
	
	return bb.BindBody(body, obj)
}

// BindingBody 请求体绑定接口
type BindingBody interface {
	BindBody([]byte, any) error
}

// BodyBytesKey 请求体缓存键
const BodyBytesKey = "_gin-gonic/gin/bodybyteskey"

// =============================================================================
// 客户端信息获取 (#33)
// =============================================================================

// ClientIP 获取客户端IP地址
//
// 返回：
//   - string: 客户端IP地址
//
// 示例：
//
//	ip := c.ClientIP()
//	fmt.Printf("客户端IP: %s", ip)
func (c *Context) ClientIP() string {
	// 检查代理头
	if c.engine.ForwardedByClientIP {
		// 检查X-Forwarded-For头
		clientIP := c.requestHeader("X-Forwarded-For")
		clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
		if clientIP != "" {
			return clientIP
		}
		
		// 检查X-Real-IP头
		clientIP = strings.TrimSpace(c.requestHeader("X-Real-IP"))
		if clientIP != "" {
			return clientIP
		}
		
		// 检查其他代理头
		for _, header := range c.engine.RemoteIPHeaders {
			ip := c.requestHeader(header)
			if ip != "" {
				return strings.TrimSpace(strings.Split(ip, ",")[0])
			}
		}
	}
	
	// 从RemoteAddr获取
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request().RemoteAddr)); err == nil {
		return ip
	}
	
	return ""
}

// requestHeader 获取请求头（辅助方法）
func (c *Context) requestHeader(key string) string {
	return c.Request().Header.Get(key)
}

// ContentType 获取请求内容类型
//
// 返回：
//   - string: 内容类型
//
// 示例：
//
//	contentType := c.ContentType()
//	if contentType == "application/json" {
//		// 处理JSON请求
//	}
func (c *Context) ContentType() string {
	return filterFlags(c.requestHeader("Content-Type"))
}

// 注意：filterFlags函数可能已在其他文件中定义

// IsWebsocket 检查是否为WebSocket请求
//
// 返回：
//   - bool: 是否为WebSocket请求
//
// 示例：
//
//	if c.IsWebsocket() {
//		// 处理WebSocket升级
//	}
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.requestHeader("Connection")), "upgrade") &&
		strings.EqualFold(c.requestHeader("Upgrade"), "websocket") {
		return true
	}
	return false
}

// =============================================================================
// 响应写入器包装 (#34)
// =============================================================================

// Writer 返回ResponseWriter接口
//
// 返回：
//   - ResponseWriter: 响应写入器
//
// 示例：
//
//	w := c.Writer()
//	w.WriteHeader(200)
//	w.Write([]byte("Hello World"))
func (c *Context) Writer() ResponseWriter {
	return &responseWriterWrapper{c.RequestContext}
}

// ResponseWriter 响应写入器接口
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier
	
	// Size 返回已写入的字节数
	Size() int
	
	// Written 返回是否已写入响应
	Written() bool
	
	// WriteHeaderNow 强制写入状态码
	WriteHeaderNow()
	
	// Pusher 返回HTTP/2 推送器
	Pusher() http.Pusher
}

// responseWriterWrapper ResponseWriter包装器
type responseWriterWrapper struct {
	*app.RequestContext
}

// Header 实现http.ResponseWriter
func (w *responseWriterWrapper) Header() http.Header {
	header := make(http.Header)
	w.RequestContext.Response.Header.VisitAll(func(key, value []byte) {
		header[string(key)] = []string{string(value)}
	})
	return header
}

// Write 实现http.ResponseWriter
func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	w.RequestContext.Write(data)
	return len(data), nil
}

// WriteHeader 实现http.ResponseWriter
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.RequestContext.SetStatusCode(statusCode)
}

// Size 返回已写入的字节数
func (w *responseWriterWrapper) Size() int {
	return len(w.RequestContext.Response.Body())
}

// Written 返回是否已写入响应
func (w *responseWriterWrapper) Written() bool {
	return w.Size() > 0
}

// WriteHeaderNow 强制写入状态码
func (w *responseWriterWrapper) WriteHeaderNow() {
	// Hertz会自动写入状态码
}

// Pusher 返回HTTP/2推送器
func (w *responseWriterWrapper) Pusher() http.Pusher {
	return nil // Hertz暂不支持HTTP/2 Push
}

// Hijack 实现http.Hijacker
func (w *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijacking not supported")
}

// Flush 实现http.Flusher  
func (w *responseWriterWrapper) Flush() {
	w.RequestContext.Flush()
}

// CloseNotify 实现http.CloseNotifier
func (w *responseWriterWrapper) CloseNotify() <-chan bool {
	return make(<-chan bool) // 简化实现
}

// =============================================================================
// URL构建辅助方法 (#32)
// =============================================================================

// PostForm 获取POST表单值
//
// 参数：
//   - key: 表单字段名
//
// 返回：
//   - string: 表单值
//
// 示例：
//
//	username := c.PostForm("username")
func (c *Context) PostForm(key string) string {
	value, _ := c.GetPostForm(key)
	return value
}

// DefaultPostForm 获取POST表单值，带默认值
//
// 参数：
//   - key: 表单字段名
//   - defaultValue: 默认值
//
// 返回：
//   - string: 表单值或默认值
//
// 示例：
//
//	page := c.DefaultPostForm("page", "1")
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return defaultValue
}

// GetPostForm 获取POST表单值
//
// 参数：
//   - key: 表单字段名
//
// 返回：
//   - string: 表单值
//   - bool: 是否存在
//
// 示例：
//
//	if username, ok := c.GetPostForm("username"); ok {
//		// 处理用户名
//	}
func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// PostFormArray 获取POST表单数组值
//
// 参数：
//   - key: 表单字段名
//
// 返回：
//   - []string: 表单值数组
//
// 示例：
//
//	tags := c.PostFormArray("tags")
func (c *Context) PostFormArray(key string) []string {
	values, _ := c.GetPostFormArray(key)
	return values
}

// GetPostFormArray 获取POST表单数组值
//
// 参数：
//   - key: 表单字段名
//
// 返回：
//   - []string: 表单值数组
//   - bool: 是否存在
//
// 示例：
//
//	if tags, ok := c.GetPostFormArray("tags"); ok {
//		// 处理标签数组
//	}
func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	req := c.Request()
	if err := req.ParseMultipartForm(c.engine.MaxMultipartMemory); err != nil {
		if err := req.ParseForm(); err != nil {
			return []string{}, false
		}
	}
	
	if values := req.PostForm[key]; len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// =============================================================================
// 错误处理增强 (#35)
// =============================================================================

// Error 错误结构体
type Error struct {
	Err  error
	Type ErrorType
	Meta any
}

// ErrorType 错误类型
type ErrorType uint64

const (
	// ErrorTypeBind 绑定错误
	ErrorTypeBind ErrorType = 1 << 63
	// ErrorTypeRender 渲染错误  
	ErrorTypeRender ErrorType = 1 << 62
	// ErrorTypePrivate 私有错误
	ErrorTypePrivate ErrorType = 1 << 0
	// ErrorTypePublic 公开错误
	ErrorTypePublic ErrorType = 1 << 1
	// ErrorTypeAny 任意错误
	ErrorTypeAny ErrorType = 1<<64 - 1
	
	// 默认错误类型
	ErrorTypeNu = 2
)

// Error 实现error接口
func (msg *Error) Error() string {
	return msg.Err.Error()
}

// IsType 检查错误类型
func (msg *Error) IsType(typ ErrorType) bool {
	return (msg.Type & typ) > 0
}

// Unwrap 解包错误
func (msg *Error) Unwrap() error {
	return msg.Err
}

// SetType 设置错误类型
func (msg *Error) SetType(typ ErrorType) *Error {
	msg.Type = typ
	return msg
}

// SetMeta 设置错误元数据
func (msg *Error) SetMeta(data any) *Error {
	msg.Meta = data
	return msg
}

// JSON 返回错误的JSON表示
func (msg *Error) JSON() any {
	jsonData := H{}
	
	if msg.Meta != nil {
		value := reflect.ValueOf(msg.Meta)
		switch value.Kind() {
		case reflect.Struct:
			return msg.Meta
		case reflect.Map:
			for _, key := range value.MapKeys() {
				jsonData[key.String()] = value.MapIndex(key).Interface()
			}
		default:
			jsonData["meta"] = msg.Meta
		}
	}
	
	if _, exists := jsonData["error"]; !exists {
		jsonData["error"] = msg.Error()
	}
	
	return jsonData
}

// ErrorArray 错误数组类型
type ErrorArray []*Error

// String 返回错误数组的字符串表示
func (a ErrorArray) String() string {
	if len(a) == 0 {
		return ""
	}
	
	var buffer bytes.Buffer
	for i, msg := range a {
		fmt.Fprintf(&buffer, "Error #%02d: %s\n", i+1, msg.Err)
		if msg.Meta != nil {
			fmt.Fprintf(&buffer, "     Meta: %v\n", msg.Meta)
		}
	}
	return buffer.String()
}

// Errors 实现error接口
func (a ErrorArray) Errors() []string {
	if len(a) == 0 {
		return nil
	}
	
	errorStrings := make([]string, len(a))
	for i, err := range a {
		errorStrings[i] = err.Error()
	}
	return errorStrings
}

// JSON 返回错误数组的JSON表示
func (a ErrorArray) JSON() any {
	jsonData := make([]any, len(a))
	for i, err := range a {
		jsonData[i] = err.JSON()
	}
	return jsonData
}

// ByType 按类型过滤错误
func (a ErrorArray) ByType(typ ErrorType) ErrorArray {
	if len(a) == 0 {
		return nil
	}
	
	var result ErrorArray
	for _, err := range a {
		if err.IsType(typ) {
			result = append(result, err)
		}
	}
	return result
}

// Last 返回最后一个错误
func (a ErrorArray) Last() *Error {
	if length := len(a); length > 0 {
		return a[length-1]
	}
	return nil
}

// Error 添加错误到Context
//
// 参数：
//   - err: 错误对象
//
// 返回：
//   - *Error: 包装的错误对象
//
// 示例：
//
//	err := c.Error(errors.New("something went wrong"))
//	err.SetType(gin.ErrorTypePublic)
func (c *Context) Error(err error) *Error {
	if err == nil {
		panic("err is nil")
	}
	
	var parsedError *Error
	ok := false
	if parsedError, ok = err.(*Error); !ok {
		parsedError = &Error{
			Err:  err,
			Type: ErrorTypePrivate,
		}
	}
	
	c.Errors = append(c.Errors, parsedError)
	return parsedError
}

// AbortWithError 中止并添加错误
//
// 参数：
//   - code: HTTP状态码
//   - err: 错误对象
//
// 返回：
//   - *Error: 包装的错误对象
//
// 示例：
//
//	c.AbortWithError(500, errors.New("internal server error"))
func (c *Context) AbortWithError(code int, err error) *Error {
	c.AbortWithStatus(code)
	return c.Error(err)
}

