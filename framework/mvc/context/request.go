package context

import (
	"net/url"
	"strconv"
	"strings"
)

// ============= 基础请求信息方法 =============

// Method 获取HTTP方法 (GET/POST等)
func (ctx *Context) Method() string {
	if !ctx.ensureRequest() {
		return ""
	}
	return safeStringConvert(ctx.request.Method())
}

// Path 获取请求路径
func (ctx *Context) Path() string {
	if !ctx.ensureRequest() {
		return ""
	}
	return safeStringConvert(ctx.request.URI().Path())
}

// Host 获取请求主机名
func (ctx *Context) Host() string {
	if !ctx.ensureRequest() {
		return ""
	}
	
	// 首先尝试从Host()方法获取
	if host := safeStringConvert(ctx.request.Host()); host != "" {
		return host
	}
	
	// 如果为空，尝试从Host头获取
	return ctx.Header("Host")
}

// URI 获取完整请求URI
func (ctx *Context) URI() string {
	if !ctx.ensureRequest() {
		return ""
	}
	return ctx.request.URI().String()
}

// Header 获取请求头
func (ctx *Context) Header(key string) string {
	if !ctx.ensureRequest() {
		return ""
	}
	return safeStringConvert(ctx.request.GetHeader(key))
}

// GetHeader 获取请求头 (兼容性别名)
func (ctx *Context) GetHeader(key string) string {
	return ctx.Header(key)
}

// ClientIP 获取客户端真实IP地址
func (ctx *Context) ClientIP() string {
	if !ctx.ensureRequest() {
		return ""
	}
	return ctx.request.ClientIP()
}

// UserAgent 获取User-Agent
func (ctx *Context) UserAgent() string {
	return ctx.Header(HeaderUserAgent)
}

// Referer 获取Referer
func (ctx *Context) Referer() string {
	return ctx.Header(HeaderReferer)
}

// ContentType 获取Content-Type
func (ctx *Context) ContentType() string {
	return ctx.Header(HeaderContentType)
}

// ============= 请求头增强方法 =============

// HeaderContains 检查请求头是否包含指定值
func (ctx *Context) HeaderContains(key, value string) bool {
	if !ctx.ensureRequest() {
		return false
	}
	headerValue := ctx.Header(key)
	return strings.Contains(strings.ToLower(headerValue), strings.ToLower(value))
}

// ============= HTTP方法判断 =============

// IsGet 判断是否为GET请求
func (ctx *Context) IsGet() bool {
	return ctx.Method() == MethodGet
}

// IsPost 判断是否为POST请求
func (ctx *Context) IsPost() bool {
	return ctx.Method() == MethodPost
}

// IsPut 判断是否为PUT请求
func (ctx *Context) IsPut() bool {
	return ctx.Method() == MethodPut
}

// IsDelete 判断是否为DELETE请求
func (ctx *Context) IsDelete() bool {
	return ctx.Method() == MethodDelete
}

// IsPatch 判断是否为PATCH请求
func (ctx *Context) IsPatch() bool {
	return ctx.Method() == MethodPatch
}

// IsHead 判断是否为HEAD请求
func (ctx *Context) IsHead() bool {
	return ctx.Method() == MethodHead
}

// IsOptions 判断是否为OPTIONS请求
func (ctx *Context) IsOptions() bool {
	return ctx.Method() == MethodOptions
}

// IsMethod 判断是否为指定的HTTP方法
func (ctx *Context) IsMethod(method string) bool {
	return strings.EqualFold(ctx.Method(), method)
}

// ============= 内容类型判断 =============

// IsJSON 判断请求内容是否为JSON
func (ctx *Context) IsJSON() bool {
	return ctx.HeaderContains(HeaderContentType, ContentTypeJSON)
}

// IsXML 判断请求内容是否为XML
func (ctx *Context) IsXML() bool {
	contentType := ctx.ContentType()
	return strings.Contains(strings.ToLower(contentType), "xml")
}

// IsForm 判断是否为表单请求
func (ctx *Context) IsForm() bool {
	return ctx.HeaderContains(HeaderContentType, ContentTypeForm)
}

// IsMultipart 判断是否为multipart表单
func (ctx *Context) IsMultipart() bool {
	return ctx.HeaderContains(HeaderContentType, ContentTypeMultipart)
}

// IsUpload 判断是否为文件上传请求
func (ctx *Context) IsUpload() bool {
	return ctx.IsMultipart()
}

// ============= 特殊请求类型判断 =============

// IsAjax 判断是否为Ajax请求
func (ctx *Context) IsAjax() bool {
	return ctx.HeaderContains(HeaderXRequestedWith, "XMLHttpRequest")
}

// IsSecure 判断是否为HTTPS请求
func (ctx *Context) IsSecure() bool {
	if !ctx.ensureRequest() {
		return false
	}
	return safeStringConvert(ctx.request.URI().Scheme()) == "https"
}

// IsWebsocket 判断是否为WebSocket请求
func (ctx *Context) IsWebsocket() bool {
	return ctx.HeaderContains(HeaderUpgrade, "websocket") &&
		ctx.HeaderContains(HeaderConnection, "upgrade")
}

// ============= 查询参数方法 =============

// Query 获取查询参数
func (ctx *Context) Query(key string) string {
	if !ctx.ensureRequest() {
		return ""
	}
	return safeStringConvert(ctx.request.QueryArgs().Peek(key))
}

// QueryDefault 带默认值的查询参数
func (ctx *Context) QueryDefault(key, defaultValue string) string {
	return parseValueWithDefault(ctx.Query(key), defaultValue)
}

// QueryAll 获取所有同名查询参数
func (ctx *Context) QueryAll(key string) []string {
	if !ctx.ensureRequest() {
		return nil
	}

	values := getStringSlice()
	defer putStringSlice(values)

	result := make([]string, 0, 4) // 预分配4个位置
	ctx.request.QueryArgs().VisitAll(func(k, v []byte) {
		if safeStringConvert(k) == key {
			result = append(result, safeStringConvert(v))
		}
	})

	if len(result) == 0 {
		return nil
	}
	return result
}

// QueryMap 获取查询参数映射，返回url.Values
func (ctx *Context) QueryMap() url.Values {
	if !ctx.ensureRequest() {
		return make(url.Values)
	}

	values := make(url.Values)
	ctx.request.QueryArgs().VisitAll(func(k, v []byte) {
		key := safeStringConvert(k)
		value := safeStringConvert(v)
		values.Add(key, value)
	})
	return values
}

// ============= 表单参数方法 =============

// PostForm 获取POST表单参数
func (ctx *Context) PostForm(key string) string {
	if !ctx.ensureRequest() {
		return ""
	}
	return safeStringConvert(ctx.request.PostArgs().Peek(key))
}

// FormValue 获取表单值 (GET/POST通用)
func (ctx *Context) FormValue(key string) string {
	if !ctx.ensureRequest() {
		return ""
	}

	// 先尝试POST参数
	if value := safeStringConvert(ctx.request.PostArgs().Peek(key)); value != "" {
		return value
	}

	// 再尝试GET参数
	return safeStringConvert(ctx.request.QueryArgs().Peek(key))
}

// FormValueDefault 带默认值的表单参数
func (ctx *Context) FormValueDefault(key, defaultValue string) string {
	return parseValueWithDefault(ctx.FormValue(key), defaultValue)
}

// ============= 路由参数方法 =============

// Param 获取路由参数
func (ctx *Context) Param(key string) string {
	return ctx.params.ByName(key)
}

// ============= 原始请求体方法 =============

// RawBody 获取原始请求体
func (ctx *Context) RawBody() ([]byte, error) {
	if !ctx.ensureRequest() {
		return nil, ErrRequestNotFound
	}

	body, err := ctx.request.Body()
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, ErrEmptyBody
	}

	// 复制一份数据，避免原始数据被修改
	result := make([]byte, len(body))
	copy(result, body)
	return result, nil
}

// GetRawData 获取原始请求数据 (兼容gin命名)
func (ctx *Context) GetRawData() ([]byte, error) {
	return ctx.RawBody()
}

// ============= 增强Request操作（请求体处理） =============

// BodyReader 获取请求体读取器
func (ctx *Context) BodyReader() (*BodyReader, error) {
	if !ctx.ensureRequest() {
		return nil, ErrRequestNotFound
	}

	body, err := ctx.request.Body()
	if err != nil {
		return nil, err
	}

	return &BodyReader{
		data:   body,
		offset: 0,
		size:   len(body),
	}, nil
}

// BodySize 获取请求体大小
func (ctx *Context) BodySize() int64 {
	if !ctx.ensureRequest() {
		return 0
	}

	body, err := ctx.request.Body()
	if err != nil {
		return 0
	}
	return int64(len(body))
}

// HasRequestBody 检查是否有请求体
func (ctx *Context) HasRequestBody() bool {
	return ctx.BodySize() > 0
}

// ============= 增强Request操作（Cookie增强） =============

// SetEnhancedCookie 设置增强型安全Cookie（HttpOnly + Secure + SameSite）
func (ctx *Context) SetEnhancedCookie(name, value string, maxAge int, path, domain string, sameSite string) {
	if !ctx.ensureRequest() {
		return
	}

	cookie := &EnhancedCookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		HttpOnly: true,
		Secure:   ctx.IsSecure(),
		SameSite: sameSite,
	}

	ctx.SetEnhancedCookieObject(cookie)
}

// GetEnhancedCookie 获取Cookie值（增强版）
func (ctx *Context) GetEnhancedCookie(name string) (string, error) {
	if !ctx.ensureRequest() {
		return "", ErrRequestNotFound
	}

	value := safeStringConvert(ctx.request.Cookie(name))
	if value == "" {
		return "", ErrCookieNotFound
	}
	return value, nil
}

// GetEnhancedCookieDefault 获取Cookie值（带默认值）
func (ctx *Context) GetEnhancedCookieDefault(name, defaultValue string) string {
	value, err := ctx.GetEnhancedCookie(name)
	if err != nil {
		return defaultValue
	}
	return value
}

// ListCookies 列出所有Cookie
func (ctx *Context) ListCookies() map[string]string {
	if !ctx.ensureRequest() {
		return make(map[string]string)
	}

	cookies := make(map[string]string)
	ctx.request.VisitAllCookie(func(key, value []byte) {
		cookies[safeStringConvert(key)] = safeStringConvert(value)
	})
	return cookies
}

// ============= 增强Request操作（头部增强） =============

// SetRequestHeader 设置请求头（在代理场景中使用）
func (ctx *Context) SetRequestHeader(key, value string) {
	if ctx.ensureRequest() {
		ctx.request.Request.Header.Set(key, value)
	}
}

// HeaderMap 获取所有请求头的映射
func (ctx *Context) HeaderMap() map[string][]string {
	if !ctx.ensureRequest() {
		return make(map[string][]string)
	}

	headers := make(map[string][]string)
	ctx.request.Request.Header.VisitAll(func(key, value []byte) {
		k := safeStringConvert(key)
		v := safeStringConvert(value)
		headers[k] = append(headers[k], v)
	})
	return headers
}

// HeaderExists 检查请求头是否存在
func (ctx *Context) HeaderExists(key string) bool {
	return ctx.Header(key) != ""
}

// HeadersContaining 获取包含指定关键词的所有头部
func (ctx *Context) HeadersContaining(keyword string) map[string]string {
	result := make(map[string]string)
	headerMap := ctx.HeaderMap()

	for key, values := range headerMap {
		if strings.Contains(strings.ToLower(key), strings.ToLower(keyword)) {
			if len(values) > 0 {
				result[key] = values[0] // 取第一个值
			}
		}
	}
	return result
}

// ============= 增强Request操作（请求验证） =============

// ValidateContentType 验证Content-Type
func (ctx *Context) ValidateContentType(expectedTypes ...string) bool {
	contentType := strings.ToLower(ctx.ContentType())
	for _, expected := range expectedTypes {
		if strings.Contains(contentType, strings.ToLower(expected)) {
			return true
		}
	}
	return false
}

// ValidateMethod 验证HTTP方法
func (ctx *Context) ValidateMethod(allowedMethods ...string) bool {
	currentMethod := ctx.Method()
	for _, allowed := range allowedMethods {
		if strings.EqualFold(currentMethod, allowed) {
			return true
		}
	}
	return false
}

// ValidateOrigin 验证Origin头部
func (ctx *Context) ValidateOrigin(allowedOrigins ...string) bool {
	origin := ctx.Header("Origin")
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed || allowed == "*" {
			return true
		}
	}
	return false
}

// ============= 增强Request操作（性能优化） =============

// GetRequestWithCache 带缓存的获取request对象
func (ctx *Context) GetRequestWithCache() any {
	return ctx.request // 直接返回，因为已经被缓存在Context中
}

// RequestInfo 获取请求的详细信息（用于调试和监控）
func (ctx *Context) RequestInfo() *RequestInfo {
	if !ctx.ensureRequest() {
		return nil
	}

	return &RequestInfo{
		Method:      ctx.Method(),
		Path:        ctx.Path(),
		Host:        ctx.Host(),
		UserAgent:   ctx.UserAgent(),
		ContentType: ctx.ContentType(),
		BodySize:    ctx.BodySize(),
		IsSecure:    ctx.IsSecure(),
		IsAjax:      ctx.IsAjax(),
		ClientIP:    ctx.ClientIP(),
		Referer:     ctx.Referer(),
	}
}

// ============= 辅助类型定义 =============

// BodyReader 请求体读取器
type BodyReader struct {
	data   []byte
	offset int
	size   int
}

// Read 读取数据
func (br *BodyReader) Read(p []byte) (n int, err error) {
	if br.offset >= br.size {
		return 0, nil
	}

	n = copy(p, br.data[br.offset:])
	br.offset += n
	return n, nil
}

// Seek 定位读取位置
func (br *BodyReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0: // 从头开始
		br.offset = int(offset)
	case 1: // 从当前位置
		br.offset += int(offset)
	case 2: // 从末尾开始
		br.offset = br.size + int(offset)
	}

	if br.offset < 0 {
		br.offset = 0
	}
	if br.offset > br.size {
		br.offset = br.size
	}

	return int64(br.offset), nil
}

// EnhancedCookie 增强的Cookie结构
type EnhancedCookie struct {
	Name     string
	Value    string
	MaxAge   int
	Path     string
	Domain   string
	HttpOnly bool
	Secure   bool
	SameSite string
}

// SetEnhancedCookieObject 设置Cookie（增强版）
func (ctx *Context) SetEnhancedCookieObject(cookie *EnhancedCookie) {
	if !ctx.ensureRequest() {
		return
	}

	// 构建cookie字符串
	cookieStr := cookie.Name + "=" + cookie.Value

	if cookie.MaxAge > 0 {
		cookieStr += "; Max-Age=" + strconv.Itoa(cookie.MaxAge)
	}

	if cookie.Path != "" {
		cookieStr += "; Path=" + cookie.Path
	}

	if cookie.Domain != "" {
		cookieStr += "; Domain=" + cookie.Domain
	}

	if cookie.HttpOnly {
		cookieStr += "; HttpOnly"
	}

	if cookie.Secure {
		cookieStr += "; Secure"
	}

	if cookie.SameSite != "" {
		cookieStr += "; SameSite=" + cookie.SameSite
	}

	ctx.request.Response.Header.Add("Set-Cookie", cookieStr)
}

// RequestInfo 请求信息结构体
type RequestInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Host        string `json:"host"`
	UserAgent   string `json:"user_agent"`
	ContentType string `json:"content_type"`
	BodySize    int64  `json:"body_size"`
	IsSecure    bool   `json:"is_secure"`
	IsAjax      bool   `json:"is_ajax"`
	ClientIP    string `json:"client_ip"`
	Referer     string `json:"referer"`
}

// ============= 错误定义 =============

// 请求处理相关的错误
var (
	ErrRequestNotFound = &ContextError{Code: "REQUEST_NOT_FOUND", Message: "Request context not found"}
	ErrEmptyBody       = &ContextError{Code: "EMPTY_BODY", Message: "Request body is empty"}
	ErrCookieNotFound  = &ContextError{Code: "COOKIE_NOT_FOUND", Message: "Cookie not found"}
)

// ContextError 上下文相关错误类型
type ContextError struct {
	Code    string
	Message string
	Cause   error
}

func (e *ContextError) Error() string {
	return e.Message
}
