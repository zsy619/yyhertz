// Package context 增强的上下文处理
// 整合Gin风格的上下文处理能力，并继承Beego的Context设计
package context

import (
	"encoding/json"
	"fmt"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/binding"
	"github.com/zsy619/yyhertz/framework/render"
)

// Context 增强的上下文处理（统一的Context设计）
type Context struct {
	*app.RequestContext // 嵌入Hertz的RequestContext

	// ============= 输入输出处理 =============
	Input          *Input          // 输入处理器
	Output         *Output         // 输出处理器
	ResponseWriter *ResponseWriter // 响应写入器

	// ============= 扩展属性 =============
	// 参数存储
	Params   Params
	Keys     map[string]any
	Errors   errorMsgs
	Accepted []string

	// 引擎引用
	engine *Engine

	// 中间件
	handlers []HandlerFunc
	index    int8

	// 查询缓存
	queryCache url.Values
	formCache  url.Values

	// 采样器
	sampler func() bool
}

// Input 输入处理器
type Input struct {
	context *Context
	data    map[any]any // 存储解析后的数据
}

// Output 输出处理器
type Output struct {
	context *Context
	Status  int               // HTTP状态码
	headers map[string]string // 响应头
}

// ResponseWriter 响应写入器（适配Hertz）
type ResponseWriter struct {
	context *app.RequestContext
	status  int
	size    int
}

// NewContext 创建增强的Context
func NewContext(ctx *app.RequestContext) *Context {
	enhancedCtx := &Context{
		RequestContext: ctx,
		Keys:           make(map[string]any),
		queryCache:     make(url.Values),
		formCache:      make(url.Values),
	}

	// 初始化Input和Output
	enhancedCtx.Input = &Input{
		context: enhancedCtx,
		data:    make(map[any]any),
	}
	enhancedCtx.Output = &Output{
		context: enhancedCtx,
		headers: make(map[string]string),
	}
	enhancedCtx.ResponseWriter = &ResponseWriter{
		context: ctx,
	}

	return enhancedCtx
}

// ConvertToContext 从Hertz的RequestContext转换为增强的Context（保持向后兼容）
func ConvertToContext(ctx *app.RequestContext) *Context {
	return NewContext(ctx)
}

// ============= Input 方法 =============

// Header 获取请求头
func (input *Input) Header(key string) string {
	return string(input.context.RequestContext.Request.Header.Peek(key))
}

// Query 获取查询参数
func (input *Input) Query(key string) string {
	return input.context.RequestContext.Query(key)
}

// Param 获取路径参数
func (input *Input) Param(key string) string {
	return input.context.RequestContext.Param(key)
}

// Cookie 获取Cookie
func (input *Input) Cookie(key string) string {
	return string(input.context.RequestContext.Cookie(key))
}

// Session 获取Session值
func (input *Input) Session(key any) any {
	// 这里需要与你的session系统集成
	if store, exists := input.context.Get("session"); exists {
		// 假设session实现了Get方法
		if s, ok := store.(interface{ Get(any) any }); ok {
			return s.Get(key)
		}
	}
	return nil
}

// Body 获取请求体
func (input *Input) Body() ([]byte, error) {
	return input.context.RequestContext.Request.Body(), nil
}

// JSON 解析JSON到指定对象
func (input *Input) JSON(obj any) error {
	body, err := input.Body()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, obj)
}

// Bind 数据绑定（继承原有的绑定功能）
func (input *Input) Bind(obj any) error {
	b := binding.Default(string(input.context.RequestContext.Request.Method()), input.ContentType())
	return b.Bind(input.context.RequestContext, obj)
}

// ContentType 获取内容类型
func (input *Input) ContentType() string {
	return string(input.context.RequestContext.Request.Header.ContentType())
}

// IsAjax 判断是否为Ajax请求
func (input *Input) IsAjax() bool {
	return input.Header("X-Requested-With") == "XMLHttpRequest"
}

// IsSecure 判断是否为HTTPS请求
func (input *Input) IsSecure() bool {
	return string(input.context.RequestContext.URI().Scheme()) == "https"
}

// IsWebsocket 判断是否为WebSocket请求
func (input *Input) IsWebsocket() bool {
	return strings.ToLower(input.Header("Connection")) == "upgrade" &&
		strings.ToLower(input.Header("Upgrade")) == "websocket"
}

// IsUpload 判断是否为文件上传请求
func (input *Input) IsUpload() bool {
	return strings.Contains(input.ContentType(), "multipart/form-data")
}

// IP 获取客户端IP
func (input *Input) IP() string {
	return input.context.RequestContext.ClientIP()
}

// Proxy 获取代理信息
func (input *Input) Proxy() []string {
	if ips := input.Header("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}

// Refer 获取来源页面
func (input *Input) Refer() string {
	return input.Header("Referer")
}

// SubDomains 获取子域名
func (input *Input) SubDomains() []string {
	host := input.Header("Host")
	if host == "" {
		return []string{}
	}

	parts := strings.Split(host, ".")
	if len(parts) <= 2 {
		return []string{}
	}

	return parts[:len(parts)-2]
}

// Port 获取端口
func (input *Input) Port() int {
	host := input.Header("Host")
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		if len(parts) > 1 {
			if port, err := strconv.Atoi(parts[1]); err == nil {
				return port
			}
		}
	}

	// 默认端口
	if input.IsSecure() {
		return 443
	}
	return 80
}

// UserAgent 获取用户代理
func (input *Input) UserAgent() string {
	return input.Header("User-Agent")
}

// Data 设置数据
func (input *Input) Data(key, val any) {
	input.data[key] = val
}

// GetData 获取数据
func (input *Input) GetData(key any) any {
	return input.data[key]
}

// ============= Output 方法 =============

// Header 设置响应头
func (output *Output) Header(key, val string) {
	output.headers[key] = val
	output.context.RequestContext.Response.Header.Set(key, val)
}

// Body 设置响应体
func (output *Output) Body(content []byte) error {
	output.context.RequestContext.Write(content)
	return nil
}

// Cookie 设置Cookie
func (output *Output) Cookie(name, value string, others ...any) {
	var b strings.Builder
	fmt.Fprintf(&b, "%s=%s", name, value)

	// 处理其他Cookie属性
	for i, other := range others {
		switch i {
		case 0: // MaxAge
			if maxAge, ok := other.(int); ok {
				fmt.Fprintf(&b, "; Max-Age=%d", maxAge)
			}
		case 1: // Path
			if path, ok := other.(string); ok && path != "" {
				fmt.Fprintf(&b, "; Path=%s", path)
			}
		case 2: // Domain
			if domain, ok := other.(string); ok && domain != "" {
				fmt.Fprintf(&b, "; Domain=%s", domain)
			}
		case 3: // Secure
			if secure, ok := other.(bool); ok && secure {
				fmt.Fprintf(&b, "; Secure")
			}
		case 4: // HttpOnly
			if httpOnly, ok := other.(bool); ok && httpOnly {
				fmt.Fprintf(&b, "; HttpOnly")
			}
		case 5: // SameSite
			if sameSite, ok := other.(string); ok && sameSite != "" {
				fmt.Fprintf(&b, "; SameSite=%s", sameSite)
			}
		}
	}

	output.context.RequestContext.Response.Header.Set("Set-Cookie", b.String())
}

// JSON 输出JSON
func (output *Output) JSON(data any, hasIndent bool, encoding bool) error {
	output.Header("Content-Type", "application/json; charset=utf-8")

	var content []byte
	var err error

	if hasIndent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = json.Marshal(data)
	}

	if err != nil {
		return err
	}

	return output.Body(content)
}

// JSONP 输出JSONP
func (output *Output) JSONP(data any, hasIndent bool) error {
	callback := output.context.Query("callback")
	if callback == "" {
		return output.JSON(data, hasIndent, true)
	}

	output.Header("Content-Type", "application/javascript; charset=utf-8")

	var content []byte
	var err error

	if hasIndent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = json.Marshal(data)
	}

	if err != nil {
		return err
	}

	callback_content := fmt.Sprintf("%s(%s);", callback, string(content))
	return output.Body([]byte(callback_content))
}

// XML 输出XML
func (output *Output) XML(data any, hasIndent bool) error {
	output.Header("Content-Type", "application/xml; charset=utf-8")
	// 这里需要实现XML序列化，暂时简化
	return output.Body([]byte(fmt.Sprintf("<xml>%v</xml>", data)))
}

// Download 文件下载
func (output *Output) Download(file string, filename ...string) {
	output.Header("Content-Description", "File Transfer")
	output.Header("Content-Type", "application/octet-stream")
	if len(filename) > 0 && filename[0] != "" {
		output.Header("Content-Disposition", "attachment; filename="+filename[0])
	} else {
		output.Header("Content-Disposition", "attachment; filename="+file)
	}
	output.Header("Content-Transfer-Encoding", "binary")
	output.Header("Expires", "0")
	output.Header("Cache-Control", "must-revalidate")
	output.Header("Pragma", "public")

	// 这里需要实现文件读取和输出，暂时简化
	output.Body([]byte("File download not implemented"))
}

// ContentType 设置内容类型
func (output *Output) ContentType(ext string) {
	var contentType string
	switch ext {
	case "json":
		contentType = "application/json"
	case "xml":
		contentType = "application/xml"
	case "html":
		contentType = "text/html"
	case "text":
		contentType = "text/plain"
	default:
		contentType = ext
	}
	output.Header("Content-Type", contentType+"; charset=utf-8")
}

// SetStatus 设置HTTP状态码
func (output *Output) SetStatus(status int) {
	output.Status = status
	output.context.RequestContext.SetStatusCode(status)
}

// Session 设置Session值
func (output *Output) Session(name any, value any) {
	// 这里需要与你的session系统集成
	if store, exists := output.context.Get("session"); exists {
		// 假设session实现了Set方法
		if s, ok := store.(interface {
			Set(any, any)
		}); ok {
			s.Set(name, value)
		}
	}
}

// ============= ResponseWriter 方法 =============

func (w *ResponseWriter) Header() http.Header {
	// 适配Hertz的Header到标准库的Header
	header := make(http.Header)
	w.context.Response.Header.VisitAll(func(key, value []byte) {
		header.Set(string(key), string(value))
	})
	return header
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	w.size += len(data)
	w.context.Write(data)
	return len(data), nil
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.context.SetStatusCode(statusCode)
}

func (w *ResponseWriter) Size() int {
	return w.size
}

func (w *ResponseWriter) Status() int {
	return w.status
}

func (w *ResponseWriter) Written() bool {
	return w.size > 0
}

// ============= Context 的方法 =============

// Reset 重置Context
func (c *Context) Reset() {
	c.Keys = make(map[string]any)
	c.Errors = nil
	c.Accepted = nil
	c.queryCache = make(url.Values)
	c.formCache = make(url.Values)
	c.index = -1
}

// Copy 复制Context
func (c *Context) Copy() *Context {
	copied := &Context{
		RequestContext: c.RequestContext,
		Keys:           make(map[string]any),
		Params:         make(Params, len(c.Params)),
		engine:         c.engine,
		sampler:        c.sampler,
	}

	// 复制Keys
	for k, v := range c.Keys {
		copied.Keys[k] = v
	}

	// 复制Params
	copy(copied.Params, c.Params)

	// 重新创建Input和Output
	copied.Input = &Input{
		context: copied,
		data:    make(map[any]any),
	}
	copied.Output = &Output{
		context: copied,
		headers: make(map[string]string),
	}
	copied.ResponseWriter = &ResponseWriter{
		context: c.RequestContext,
	}

	return copied
}

// ============= 类型定义 =============

// HandlerFunc 处理函数类型
type HandlerFunc func(*Context)

// Engine 引擎接口
type Engine interface {
	// 可以定义引擎需要的方法
}

// Params 路由参数
type Params []Param

// Param 单个路由参数
type Param struct {
	Key   string
	Value string
}

// Get 获取参数值
func (ps Params) Get(name string) (string, bool) {
	for _, p := range ps {
		if p.Key == name {
			return p.Value, true
		}
	}
	return "", false
}

// ByName 根据名称获取参数值
func (ps Params) ByName(name string) string {
	va, _ := ps.Get(name)
	return va
}

// errorMsg 错误消息
type errorMsg struct {
	Err  error
	Type errorType
	Meta any
}

type errorType uint64

const (
	ErrorTypeBind    errorType = 1 << 63
	ErrorTypeRender  errorType = 1 << 62
	ErrorTypePrivate errorType = 1 << 0
	ErrorTypePublic  errorType = 1 << 1
	ErrorTypeAny     errorType = 1<<64 - 1
)

type errorMsgs []*errorMsg

// String 实现Stringer接口
func (msg *errorMsg) String() string {
	return msg.Err.Error()
}

// Error 实现error接口
func (msg *errorMsg) Error() string {
	return msg.String()
}

// ============= 中间件方法 =============

// Next 执行下一个中间件
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// IsAborted 检查是否已终止
func (c *Context) IsAborted() bool {
	return c.index >= int8(len(c.handlers))
}

// Abort 终止执行
func (c *Context) Abort() {
	c.index = math.MaxInt8 >> 1
}

// AbortWithStatus 终止并设置状态码
func (c *Context) AbortWithStatus(code int) {
	c.SetStatusCode(code)
	c.Abort()
}

// AbortWithStatusJSON 终止并返回JSON错误
func (c *Context) AbortWithStatusJSON(code int, jsonObj any) {
	c.Abort()
	c.JSON(code, jsonObj)
}

// AbortWithError 终止并添加错误
func (c *Context) AbortWithError(code int, err error) *errorMsg {
	c.AbortWithStatus(code)
	return c.Error(err)
}

// ============= 错误处理 =============

// Error 添加错误
func (c *Context) Error(err error) *errorMsg {
	if err == nil {
		panic("err is nil")
	}

	var parsedError *errorMsg
	ok := false
	if parsedError, ok = err.(*errorMsg); !ok {
		parsedError = &errorMsg{
			Err:  err,
			Type: ErrorTypePrivate,
		}
	}

	c.Errors = append(c.Errors, parsedError)
	return parsedError
}

// ============= 参数获取 =============

// Param 获取路径参数
func (c *Context) Param(key string) string {
	return c.Params.ByName(key)
}

// Query 获取查询参数
func (c *Context) Query(key string) string {
	value, _ := c.GetQuery(key)
	return value
}

// DefaultQuery 获取查询参数，带默认值
func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return defaultValue
}

// GetQuery 获取查询参数
func (c *Context) GetQuery(key string) (string, bool) {
	if values, ok := c.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// QueryArray 获取查询参数数组
func (c *Context) QueryArray(key string) []string {
	values, _ := c.GetQueryArray(key)
	return values
}

// GetQueryArray 获取查询参数数组
func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	if values, ok := c.queryCache[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// QueryMap 获取查询参数映射
func (c *Context) QueryMap(key string) map[string]string {
	dicts, _ := c.GetQueryMap(key)
	return dicts
}

// GetQueryMap 获取查询参数映射
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

// PostForm 获取表单参数
func (c *Context) PostForm(key string) string {
	value, _ := c.GetPostForm(key)
	return value
}

// DefaultPostForm 获取表单参数，带默认值
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	if value, ok := c.GetPostForm(key); ok {
		return value
	}
	return defaultValue
}

// GetPostForm 获取表单参数
func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// PostFormArray 获取表单参数数组
func (c *Context) PostFormArray(key string) []string {
	values, _ := c.GetPostFormArray(key)
	return values
}

// GetPostFormArray 获取表单参数数组
func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.initFormCache()
	if values, ok := c.formCache[key]; ok && len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// PostFormMap 获取表单参数映射
func (c *Context) PostFormMap(key string) map[string]string {
	dicts, _ := c.GetPostFormMap(key)
	return dicts
}

// GetPostFormMap 获取表单参数映射
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initFormCache()
	return c.get(c.formCache, key)
}

// FormFile 获取上传文件
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	// Hertz的FormFile API与标准库不同，需要适配
	// 这里返回一个简单的实现，实际项目中需要根据Hertz的API调整
	return nil, fmt.Errorf("FormFile not implemented for Hertz")
}

// MultipartForm 获取多部分表单
func (c *Context) MultipartForm() (*multipart.Form, error) {
	// Hertz的MultipartForm API与标准库不同，需要适配
	return nil, fmt.Errorf("MultipartForm not implemented for Hertz")
}

// SaveUploadedFile 保存上传的文件
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	// 这里需要实现文件保存逻辑
	// 由于Hertz的API与标准库不同，需要适配
	return fmt.Errorf("SaveUploadedFile not implemented for Hertz")
}

// ============= 绑定方法 =============

// Bind 自动绑定
func (c *Context) Bind(obj any) error {
	b := binding.Default(string(c.RequestContext.Method()), c.ContentType())
	return c.MustBindWith(obj, b)
}

// BindJSON 绑定JSON
func (c *Context) BindJSON(obj any) error {
	return c.MustBindWith(obj, binding.JSON)
}

// BindXML 绑定XML
func (c *Context) BindXML(obj any) error {
	return c.MustBindWith(obj, binding.XML)
}

// BindQuery 绑定查询参数
func (c *Context) BindQuery(obj any) error {
	return c.MustBindWith(obj, binding.Query)
}

// BindYAML 绑定YAML
func (c *Context) BindYAML(obj any) error {
	return c.MustBindWith(obj, binding.YAML)
}

// BindHeader 绑定请求头
func (c *Context) BindHeader(obj any) error {
	return c.MustBindWith(obj, binding.Header)
}

// BindUri 绑定URI参数
func (c *Context) BindUri(obj any) error {
	if err := c.bindUri(obj); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind)
		return err
	}
	return nil
}

// MustBindWith 必须绑定
func (c *Context) MustBindWith(obj any, b binding.Binding) error {
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind)
		return err
	}
	return nil
}

// ShouldBind 应该绑定
func (c *Context) ShouldBind(obj any) error {
	b := binding.Default(string(c.RequestContext.Method()), c.ContentType())
	return c.ShouldBindWith(obj, b)
}

// ShouldBindJSON 应该绑定JSON
func (c *Context) ShouldBindJSON(obj any) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindXML 应该绑定XML
func (c *Context) ShouldBindXML(obj any) error {
	return c.ShouldBindWith(obj, binding.XML)
}

// ShouldBindQuery 应该绑定查询参数
func (c *Context) ShouldBindQuery(obj any) error {
	return c.ShouldBindWith(obj, binding.Query)
}

// ShouldBindYAML 应该绑定YAML
func (c *Context) ShouldBindYAML(obj any) error {
	return c.ShouldBindWith(obj, binding.YAML)
}

// ShouldBindHeader 应该绑定请求头
func (c *Context) ShouldBindHeader(obj any) error {
	return c.ShouldBindWith(obj, binding.Header)
}

// ShouldBindUri 应该绑定URI参数
func (c *Context) ShouldBindUri(obj any) error {
	return c.bindUri(obj)
}

// ShouldBindWith 应该绑定（使用指定绑定器）
func (c *Context) ShouldBindWith(obj any, b binding.Binding) error {
	return b.Bind(c.RequestContext, obj)
}

// ============= 渲染方法 =============

// JSON 渲染JSON
func (c *Context) JSON(code int, obj any) {
	c.Render(code, render.JSON{Data: obj})
}

// IndentedJSON 渲染带缩进的JSON
func (c *Context) IndentedJSON(code int, obj any) {
	c.Render(code, render.IndentedJSON{Data: obj})
}

// SecureJSON 渲染安全JSON
func (c *Context) SecureJSON(code int, obj any) {
	c.Render(code, render.SecureJSON{Prefix: "while(1);", Data: obj})
}

// JSONP 渲染JSONP
func (c *Context) JSONP(code int, obj any) {
	callback := c.DefaultQuery("callback", "")
	if callback == "" {
		c.Render(code, render.JSON{Data: obj})
		return
	}
	c.Render(code, render.JsonpJSON{Callback: callback, Data: obj})
}

// XML 渲染XML
func (c *Context) XML(code int, obj any) {
	c.Render(code, render.XML{Data: obj})
}

// YAML 渲染YAML
func (c *Context) YAML(code int, obj any) {
	c.Render(code, render.YAML{Data: obj})
}

// String 渲染字符串
func (c *Context) String(code int, format string, values ...any) {
	c.Render(code, render.String{Format: format, Data: values})
}

// HTML 渲染HTML
func (c *Context) HTML(code int, name string, obj any) {
	// 这里需要集成模板引擎
	c.SetStatusCode(code)
	c.SetContentType("text/html; charset=utf-8")
	c.SetBodyString(fmt.Sprintf("<html><body>HTML rendering not implemented: %s</body></html>", name))
}

// Data 渲染原始数据
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, render.Data{ContentType: contentType, Data: data})
}

// DataFromReader 从Reader渲染数据
func (c *Context) DataFromReader(code int, contentLength int64, contentType string, reader func(*app.RequestContext), extraHeaders map[string]string) {
	c.Render(code, render.Reader{
		Headers:       extraHeaders,
		ContentType:   contentType,
		ContentLength: contentLength,
		Reader:        reader,
	})
}

// Redirect 重定向
func (c *Context) Redirect(code int, location string) {
	c.Render(-1, render.Redirect{Code: code, Location: location})
}

// Render 使用渲染器渲染
func (c *Context) Render(code int, r render.Render) {
	c.SetStatusCode(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.RequestContext)
		return
	}

	if err := r.Render(c.RequestContext); err != nil {
		panic(err)
	}
}

// ============= 辅助方法 =============

// Set 设置值
func (c *Context) Set(key string, value any) {
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

// Get 获取值
func (c *Context) Get(key string) (value any, exists bool) {
	value, exists = c.Keys[key]
	return
}

// MustGet 必须获取值
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString 获取字符串值
func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool 获取布尔值
func (c *Context) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt 获取整数值
func (c *Context) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 获取64位整数值
func (c *Context) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetFloat64 获取浮点数值
func (c *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime 获取时间值
func (c *Context) GetTime(key string) (t time.Time) {
	if val, ok := c.Get(key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration 获取持续时间值
func (c *Context) GetDuration(key string) (d time.Duration) {
	if val, ok := c.Get(key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

// GetStringSlice 获取字符串切片值
func (c *Context) GetStringSlice(key string) (ss []string) {
	if val, ok := c.Get(key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap 获取字符串映射值
func (c *Context) GetStringMap(key string) (sm map[string]any) {
	if val, ok := c.Get(key); ok && val != nil {
		sm, _ = val.(map[string]any)
	}
	return
}

// GetStringMapString 获取字符串到字符串的映射值
func (c *Context) GetStringMapString(key string) (sms map[string]string) {
	if val, ok := c.Get(key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice 获取字符串到字符串切片的映射值
func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
	if val, ok := c.Get(key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}

// ClientIP 获取客户端IP
func (c *Context) ClientIP() string {
	return c.RequestContext.ClientIP()
}

// ContentType 获取内容类型
func (c *Context) ContentType() string {
	return string(c.RequestContext.ContentType())
}

// IsWebsocket 检查是否为WebSocket请求
func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") &&
		strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
		return true
	}
	return false
}

// GetHeader 获取请求头
func (c *Context) GetHeader(key string) string {
	return string(c.RequestContext.GetHeader(key))
}

// GetRawData 获取原始数据
func (c *Context) GetRawData() ([]byte, error) {
	return c.RequestContext.Request.Body(), nil
}

// SetSampler 设置采样器
func (c *Context) SetSampler(sampler func() bool) {
	c.sampler = sampler
}

// 私有方法

func (c *Context) initQueryCache() {
	if c.queryCache == nil {
		c.queryCache = make(url.Values)
		c.RequestContext.QueryArgs().VisitAll(func(key, value []byte) {
			c.queryCache.Add(string(key), string(value))
		})
	}
}

func (c *Context) initFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		// 这里需要解析表单数据
		// 由于Hertz的API不同，需要适配
	}
}

func (c *Context) get(cache url.Values, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, v := range cache {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts, exist
}

func (c *Context) bindUri(obj any) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.Uri.BindUri(m, obj)
}

// SetType 设置错误类型
func (msg *errorMsg) SetType(flags errorType) *errorMsg {
	msg.Type = flags
	return msg
}

// SetMeta 设置错误元数据
func (msg *errorMsg) SetMeta(data any) *errorMsg {
	msg.Meta = data
	return msg
}

// bodyAllowedForStatus 检查状态码是否允许响应体
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}
