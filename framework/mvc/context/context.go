package context

import (
	"context"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"gopkg.in/yaml.v2"
)

// HandlerFunc 处理函数类型
type HandlerFunc func(*Context)

// Context 增强的上下文，支持对象池化
type Context struct {
	// 核心上下文
	Request *app.RequestContext
	RequestContext *app.RequestContext // 兼容性字段别名
	Context context.Context
	
	// 路由相关
	Params   Params // 路由参数
	FullPath string // 完整路径

	// 请求数据
	Keys map[string]interface{} // 上下文键值对
	
	// 响应数据  
	Writer ResponseWriter
	ResponseWriter ResponseWriter // 兼容性字段别名
	
	// 兼容性字段 - 为了向后兼容传络MVC风格API
	Input  *InputData
	Output *OutputData
	
	// 内部状态
	index    int8           // 中间件索引
	handlers []HandlerFunc  // 处理器链
	mu       sync.RWMutex   // 读写锁
	aborted  bool           // 是否中止
	errors   []error        // 错误列表
	
	// 池化标识
	pooled   bool           // 是否来自池
	acquired time.Time      // 获取时间
}


// Reset 重置Context状态，准备复用
func (ctx *Context) Reset() {
	ctx.Request = nil
	ctx.RequestContext = nil // 同时重置兼容性别名
	ctx.Context = nil
	ctx.Params = ctx.Params[:0]
	ctx.FullPath = ""
	
	// 清空Keys但保留底层数组
	for k := range ctx.Keys {
		delete(ctx.Keys, k)
	}
	
	ctx.Writer = nil
	ctx.index = -1
	ctx.handlers = ctx.handlers[:0]
	ctx.aborted = false
	ctx.errors = ctx.errors[:0]
}

// NewContext 创建新的增强Context（使用池化）
func NewContext(c *app.RequestContext) *Context {
	ctx := defaultPool.Get()
	ctx.Request = c
	ctx.RequestContext = c // 兼容性别名指向同一对象
	ctx.Context = context.Background()
	ctx.Writer = &responseWriter{RequestContext: c}
	ctx.ResponseWriter = ctx.Writer // 兼容性别名指向同一对象
	
	// 初始化传络MVC风格兼容性字段
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}
	
	return ctx
}

// NewContextWithContext 使用指定context创建增强Context
func NewContextWithContext(c *app.RequestContext, parent context.Context) *Context {
	ctx := defaultPool.Get()
	ctx.Request = c
	ctx.RequestContext = c // 兼容性别名指向同一对象
	ctx.Context = parent
	ctx.Writer = &responseWriter{RequestContext: c}
	ctx.ResponseWriter = ctx.Writer // 兼容性别名指向同一对象
	
	// 初始化传络MVC风格兼容性字段
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}
	
	return ctx
}

// Release 释放Context到池中
func (ctx *Context) Release() {
	if ctx.pooled {
		defaultPool.Put(ctx)
		atomic.AddInt32(&poolSize, -1)
	}
}

// ============= Context核心方法 =============

// Next 执行下一个中间件
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		if !ctx.aborted {
			ctx.handlers[ctx.index](ctx)
		}
		ctx.index++
	}
}

// Abort 中止执行
func (ctx *Context) Abort() {
	ctx.aborted = true
}

// IsAborted 是否已中止
func (ctx *Context) IsAborted() bool {
	return ctx.aborted
}

// Set 设置键值对
func (ctx *Context) Set(key string, value interface{}) {
	ctx.mu.Lock()
	ctx.Keys[key] = value
	ctx.mu.Unlock()
}

// Get 获取值
func (ctx *Context) Get(key string) (interface{}, bool) {
	ctx.mu.RLock()
	value, exists := ctx.Keys[key]
	ctx.mu.RUnlock()
	return value, exists
}

// MustGet 必须获取值
func (ctx *Context) MustGet(key string) interface{} {
	if value, exists := ctx.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// Param 获取路由参数
func (ctx *Context) Param(key string) string {
	return ctx.Params.ByName(key)
}

// Query 获取查询参数
func (ctx *Context) Query(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.QueryArgs().Peek(key))
}

// PostForm 获取POST表单参数
func (ctx *Context) PostForm(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.PostArgs().Peek(key))
}

// Header 获取请求头
func (ctx *Context) Header(key string) string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.GetHeader(key))
}

// GetHeader 获取请求头 (兼容性别名)
func (ctx *Context) GetHeader(key string) string {
	return ctx.Header(key)
}

// ============= 增强请求信息方法 =============

// Method 获取HTTP方法 (GET/POST等)
func (ctx *Context) Method() string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.Method())
}

// Path 获取请求路径
func (ctx *Context) Path() string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.URI().Path())
}

// Host 获取请求主机名
func (ctx *Context) Host() string {
	if ctx.Request == nil {
		return ""
	}
	return string(ctx.Request.Host())
}

// URI 获取完整请求URI
func (ctx *Context) URI() string {
	if ctx.Request == nil {
		return ""
	}
	return ctx.Request.URI().String()
}

// HeaderContains 检查请求头是否包含指定值
func (ctx *Context) HeaderContains(key, value string) bool {
	if ctx.Request == nil {
		return false
	}
	headerValue := string(ctx.Request.GetHeader(key))
	return strings.Contains(strings.ToLower(headerValue), strings.ToLower(value))
}

// ============= 增强查询参数方法 =============

// QueryDefault 带默认值的查询参数
func (ctx *Context) QueryDefault(key, defaultValue string) string {
	if value := ctx.Query(key); value != "" {
		return value
	}
	return defaultValue
}

// QueryAll 获取所有同名查询参数
func (ctx *Context) QueryAll(key string) []string {
	if ctx.Request == nil {
		return nil
	}
	
	var values []string
	ctx.Request.QueryArgs().VisitAll(func(k, v []byte) {
		if string(k) == key {
			values = append(values, string(v))
		}
	})
	return values
}

// QueryMap 获取查询参数映射，返回url.Values
func (ctx *Context) QueryMap() url.Values {
	if ctx.Request == nil {
		return make(url.Values)
	}
	
	values := make(url.Values)
	ctx.Request.QueryArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		value := string(v)
		values.Add(key, value)
	})
	return values
}

// QueryInt 获取查询参数并转为int
func (ctx *Context) QueryInt(key string) (int, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// QueryIntDefault 获取查询参数并转为int，带默认值
func (ctx *Context) QueryIntDefault(key string, defaultValue int) int {
	if value, err := ctx.QueryInt(key); err == nil && value != 0 {
		return value
	}
	return defaultValue
}

// ============= 增强表单处理方法 =============

// FormValue 获取表单值 (GET/POST通用)
func (ctx *Context) FormValue(key string) string {
	if ctx.Request == nil {
		return ""
	}
	
	// 先尝试POST参数
	if value := string(ctx.Request.PostArgs().Peek(key)); value != "" {
		return value
	}
	
	// 再尝试GET参数
	return string(ctx.Request.QueryArgs().Peek(key))
}

// FormValueDefault 带默认值的表单参数
func (ctx *Context) FormValueDefault(key, defaultValue string) string {
	if value := ctx.FormValue(key); value != "" {
		return value
	}
	return defaultValue
}

// FormFile 获取上传的文件
func (ctx *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if ctx.Request == nil {
		return nil, io.EOF
	}
	
	return ctx.Request.FormFile(name)
}

// ParseForm 解析表单数据
func (ctx *Context) ParseForm() error {
	if ctx.Request == nil {
		return io.EOF
	}
	
	// Hertz的RequestContext会自动解析表单数据
	// 这里主要是为了兼容性
	return nil
}

// ParseMultipartForm 解析多部分表单
func (ctx *Context) ParseMultipartForm(maxMemory int64) error {
	if ctx.Request == nil {
		return io.EOF
	}
	
	// Hertz的RequestContext会自动处理multipart表单
	// 这里主要是为了兼容性
	return nil
}

// ============= 数据绑定方法 =============

// BindJSON 解析JSON请求体
func (ctx *Context) BindJSON(obj interface{}) error {
	if ctx.Request == nil {
		return io.EOF
	}
	
	return ctx.Request.BindJSON(obj)
}

// BindXML 解析XML请求体
func (ctx *Context) BindXML(obj interface{}) error {
	if ctx.Request == nil {
		return io.EOF
	}
	
	body, err := ctx.Request.Body()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return io.EOF
	}
	
	return xml.Unmarshal(body, obj)
}

// RawBody 获取原始请求体
func (ctx *Context) RawBody() ([]byte, error) {
	if ctx.Request == nil {
		return nil, io.EOF
	}
	
	body, err := ctx.Request.Body()
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, io.EOF
	}
	
	// 复制一份数据，避免原始数据被修改
	result := make([]byte, len(body))
	copy(result, body)
	return result, nil
}

// ============= 辅助判断方法 =============

// IsGet 判断是否为GET请求
func (ctx *Context) IsGet() bool {
	return ctx.Method() == "GET"
}

// IsPost 判断是否为POST请求
func (ctx *Context) IsPost() bool {
	return ctx.Method() == "POST"
}

// IsPut 判断是否为PUT请求
func (ctx *Context) IsPut() bool {
	return ctx.Method() == "PUT"
}

// IsDelete 判断是否为DELETE请求
func (ctx *Context) IsDelete() bool {
	return ctx.Method() == "DELETE"
}

// IsAjax 判断是否为Ajax请求
func (ctx *Context) IsAjax() bool {
	return ctx.HeaderContains("X-Requested-With", "XMLHttpRequest")
}

// ContentType 获取Content-Type
func (ctx *Context) ContentType() string {
	return ctx.Header("Content-Type")
}

// ============= 响应方法 =============

// JSON 返回JSON响应
func (ctx *Context) JSON(code int, obj interface{}) {
	if ctx.Request != nil {
		ctx.Request.JSON(code, obj)
	}
}

// String 返回字符串响应
func (ctx *Context) String(code int, format string, values ...interface{}) {
	if ctx.Request != nil {
		ctx.Request.String(code, format, values...)
	}
}

// HTML 返回HTML响应
func (ctx *Context) HTML(code int, name string, obj interface{}) {
	if ctx.Request != nil {
		// 这里需要集成模板引擎
		ctx.Request.HTML(code, name, obj)
	}
}

// ============= 增强响应方法 =============

// IndentedJSON 返回格式化的JSON响应 (美化输出)
func (ctx *Context) IndentedJSON(code int, obj interface{}) {
	if ctx.Request != nil {
		ctx.Request.IndentedJSON(code, obj)
	}
}

// XML 返回XML响应
func (ctx *Context) XML(code int, obj interface{}) {
	if ctx.Request != nil {
		ctx.Request.SetStatusCode(code)
		ctx.Request.Response.Header.Set("Content-Type", "application/xml; charset=utf-8")
		
		if data, err := xml.Marshal(obj); err == nil {
			ctx.Request.Response.SetBody(data)
		}
	}
}

// YAML 返回YAML响应
func (ctx *Context) YAML(code int, obj interface{}) {
	if ctx.Request != nil {
		ctx.Request.SetStatusCode(code)
		ctx.Request.Response.Header.Set("Content-Type", "application/x-yaml; charset=utf-8")
		
		if data, err := yaml.Marshal(obj); err == nil {
			ctx.Request.Response.SetBody(data)
		}
	}
}

// Data 返回原始数据响应
func (ctx *Context) Data(code int, contentType string, data []byte) {
	if ctx.Request != nil {
		ctx.Request.SetStatusCode(code)
		ctx.Request.Response.Header.Set("Content-Type", contentType)
		ctx.Request.Response.SetBody(data)
	}
}

// Redirect 重定向响应
func (ctx *Context) Redirect(code int, location string) {
	if ctx.Request != nil {
		ctx.Request.Redirect(code, []byte(location))
	}
}

// Status 设置状态码
func (ctx *Context) Status(code int) {
	if ctx.Request != nil {
		ctx.Request.SetStatusCode(code)
	}
}

// ============= 客户端信息方法 =============

// ClientIP 获取客户端真实IP地址
func (ctx *Context) ClientIP() string {
	if ctx.Request == nil {
		return ""
	}
	return ctx.Request.ClientIP()
}

// UserAgent 获取User-Agent
func (ctx *Context) UserAgent() string {
	return ctx.Header("User-Agent")
}

// Referer 获取Referer
func (ctx *Context) Referer() string {
	return ctx.Header("Referer")
}

// ============= 参数绑定增强方法 =============

// Bind 智能绑定 (根据Content-Type自动选择)
func (ctx *Context) Bind(obj interface{}) error {
	if ctx.Request == nil {
		return io.EOF
	}
	
	contentType := ctx.ContentType()
	
	if strings.Contains(contentType, "application/json") {
		return ctx.BindJSON(obj)
	} else if strings.Contains(contentType, "application/xml") {
		return ctx.BindXML(obj)
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") || 
			  strings.Contains(contentType, "multipart/form-data") {
		return ctx.Request.Bind(obj)
	}
	
	// 默认尝试JSON绑定
	return ctx.BindJSON(obj)
}

// ShouldBind 安全绑定 (不会在失败时终止)
func (ctx *Context) ShouldBind(obj interface{}) error {
	return ctx.Bind(obj)
}

// ShouldBindJSON 安全JSON绑定
func (ctx *Context) ShouldBindJSON(obj interface{}) error {
	return ctx.BindJSON(obj)
}

// ShouldBindQuery 绑定查询参数到结构体
func (ctx *Context) ShouldBindQuery(obj interface{}) error {
	if ctx.Request == nil {
		return io.EOF
	}
	return ctx.Request.BindQuery(obj)
}

// ============= 文件处理增强方法 =============

// SaveUploadedFile 保存上传文件到指定路径
func (ctx *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	if file == nil {
		return io.EOF
	}
	
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	
	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	
	// 复制文件内容
	_, err = io.Copy(out, src)
	return err
}

// FileAttachment 发送文件作为附件下载
func (ctx *Context) FileAttachment(filepath, filename string) {
	if ctx.Request == nil {
		return
	}
	
	if filename != "" {
		ctx.Request.Response.Header.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	}
	
	ctx.Request.File(filepath)
}

// ============= 查询和表单参数类型转换方法 =============

// QueryInt64 获取int64类型查询参数
func (ctx *Context) QueryInt64(key string) (int64, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

// QueryFloat64 获取float64类型查询参数
func (ctx *Context) QueryFloat64(key string) (float64, error) {
	value := ctx.Query(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

// QueryBool 获取bool类型查询参数
func (ctx *Context) QueryBool(key string) (bool, error) {
	value := ctx.Query(key)
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

// PostFormInt 获取int类型表单参数
func (ctx *Context) PostFormInt(key string) (int, error) {
	value := ctx.PostForm(key)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// PostFormFloat64 获取float64类型表单参数
func (ctx *Context) PostFormFloat64(key string) (float64, error) {
	value := ctx.PostForm(key)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

// PostFormBool 获取bool类型表单参数
func (ctx *Context) PostFormBool(key string) (bool, error) {
	value := ctx.PostForm(key)
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

// ============= Cookie增强方法 =============

// SetSameSiteCookie 设置带SameSite属性的Cookie
func (ctx *Context) SetSameSiteCookie(name, value string, maxAge int, path, domain string, sameSite http.SameSite, secure, httpOnly bool) {
	if ctx.Request == nil {
		return
	}
	
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	}
	
	ctx.Request.Response.Header.Add("Set-Cookie", cookie.String())
}

// Cookie 获取Cookie值 (简化版)
func (ctx *Context) Cookie(name string) (string, error) {
	if ctx.Request == nil {
		return "", io.EOF
	}
	
	value := string(ctx.Request.Cookie(name))
	if value == "" {
		return "", http.ErrNoCookie
	}
	return value, nil
}

// ============= 安全和验证方法 =============

// GetRawData 获取原始请求数据 (兼容gin命名)
func (ctx *Context) GetRawData() ([]byte, error) {
	return ctx.RawBody()
}

// IsSecure 判断是否为HTTPS请求
func (ctx *Context) IsSecure() bool {
	if ctx.Request == nil {
		return false
	}
	return string(ctx.Request.URI().Scheme()) == "https"
}

// IsWebsocket 判断是否为WebSocket请求
func (ctx *Context) IsWebsocket() bool {
	return ctx.HeaderContains("Upgrade", "websocket") && 
		   ctx.HeaderContains("Connection", "upgrade")
}

// ============= 中间件和流控制方法 =============

// AbortWithStatusJSON 终止并返回JSON错误
func (ctx *Context) AbortWithStatusJSON(code int, jsonObj interface{}) {
	ctx.Abort()
	ctx.JSON(code, jsonObj)
}

// Done 获取请求完成通道 (用于超时控制)
func (ctx *Context) Done() <-chan struct{} {
	if ctx.Context != nil {
		return ctx.Context.Done()
	}
	// 如果没有父Context，返回一个永不关闭的通道
	return make(<-chan struct{})
}

// SetHandlers 设置处理器链
func (ctx *Context) SetHandlers(handlers []HandlerFunc) {
	ctx.handlers = handlers
	ctx.index = -1
}

// ============= 错误处理方法 =============

// AddError 添加错误
func (ctx *Context) AddError(err error) {
	if err != nil {
		ctx.mu.Lock()
		ctx.errors = append(ctx.errors, err)
		ctx.mu.Unlock()
	}
}

// GetErrors 获取所有错误
func (ctx *Context) GetErrors() []error {
	ctx.mu.RLock()
	errors := make([]error, len(ctx.errors))
	copy(errors, ctx.errors)
	ctx.mu.RUnlock()
	return errors
}

// HasErrors 是否有错误
func (ctx *Context) HasErrors() bool {
	ctx.mu.RLock()
	hasErr := len(ctx.errors) > 0
	ctx.mu.RUnlock()
	return hasErr
}

// ClearErrors 清除所有错误
func (ctx *Context) ClearErrors() {
	ctx.mu.Lock()
	ctx.errors = ctx.errors[:0]
	ctx.mu.Unlock()
}

// LastError 获取最后一个错误
func (ctx *Context) LastError() error {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	
	if len(ctx.errors) == 0 {
		return nil
	}
	return ctx.errors[len(ctx.errors)-1]
}

// ============= 兼容性方法 =============

// AbortWithStatus 终止并设置状态码 (兼容性方法)
func (ctx *Context) AbortWithStatus(code int) {
	ctx.JSON(code, map[string]string{"error": "Request aborted"})
	ctx.Abort()
}

// Write 写入响应数据 (兼容性方法)
func (ctx *Context) Write(data []byte) (int, error) {
	return ctx.Writer.Write(data)
}

// SetHeader 设置响应头 (兼容性方法)
func (ctx *Context) SetHeader(key, value string) {
	if ctx.Request != nil {
		ctx.Request.Response.Header.Set(key, value)
	}
}

// ============= Cookie方法 (beego兼容性) =============

// GetCookie 获取Cookie (beego兼容方法)
func (ctx *Context) GetCookie(key string) string {
	return ctx.Input.Cookie(key)
}

// SetCookie 设置Cookie (beego兼容方法)
func (ctx *Context) SetCookie(name, value string, others ...interface{}) {
	ctx.Input.SetCookie(name, value, others...)
}

// GetSecureCookie 获取安全Cookie (beego兼容方法)
func (ctx *Context) GetSecureCookie(secret, key string) (string, bool) {
	return ctx.Input.GetSecureCookie(secret, key)
}

// SetSecureCookie 设置安全Cookie (beego兼容方法)
func (ctx *Context) SetSecureCookie(secret, name, value string, others ...interface{}) {
	ctx.Input.SetSecureCookie(secret, name, value, others...)
}