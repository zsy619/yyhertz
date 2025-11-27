// Package context provides framework compatibility layer.
//
// This file implements standard MVC framework compatibility adapters that allow
// seamless migration from various frameworks to YYHertz. The adapters maintain
// 100% API compatibility while leveraging YYHertz's modern, high-performance architecture.
//
// Key Features:
// - Zero-disruption migration from traditional MVC frameworks
// - Full session compatibility with standard session interfaces
// - Enhanced performance through YYHertz native optimizations
// - Unified Context architecture for better type safety
// - Extended functionality beyond traditional framework limitations
//
// Migration Benefits:
// - No code changes required for basic functionality
// - Significant performance improvements
// - Access to YYHertz's advanced features
// - Modern Go best practices and patterns
// - Seamless integration with YYHertz ecosystem
//
// NOTE: Most session and cookie functionality has been moved to the session package
// for better modularity. This file now provides compatibility proxies.
package context

import (
	"sync"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"gopkg.in/yaml.v3"

	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

// ============= 统一管理器接口定义 =============

// SessionStoreInterface Session存储接口
// 为了避免循环导入，定义接口而不直接引用session包
// 与session.Store接口完全兼容
type SessionStoreInterface interface {
	Get(key string) any
	Set(key string, value any)
	Delete(key string)
	Clear()
	GetID() string
	Destroy()
	Save() error
	Exists(key string) bool
	GetAll() map[string]any
}

// UnifiedManagerInterface 统一管理器接口
// 为了避免循环导入，定义接口而不直接引用unified包
type UnifiedManagerInterface interface {
	// Cookie操作
	SetCookie(ctx *Context, name, value string, options ...interface{})
	GetCookie(ctx *Context, name string) string
	DeleteCookie(ctx *Context, name string, path ...string)
	HasCookie(ctx *Context, name string) bool
	
	// Session操作
	SetSessionData(ctx *Context, key string, value interface{})
	GetSessionData(ctx *Context, key string) interface{}
	DeleteSessionData(ctx *Context, key string)
	
	// 上下文数据操作
	SetContextData(ctx *Context, key string, value interface{})
	GetContextData(ctx *Context, key string) interface{}
	
	// 状态检查
	IsInitialized() bool
}

// 全局统一管理器实例访问器
var (
	globalUnifiedManager UnifiedManagerInterface
	globalManagerMutex   sync.RWMutex
)

// SetGlobalUnifiedManager 设置全局统一管理器
// 这个函数将由unified包调用，以避免循环导入
func SetGlobalUnifiedManager(manager UnifiedManagerInterface) {
	globalManagerMutex.Lock()
	defer globalManagerMutex.Unlock()
	globalUnifiedManager = manager
}

// GetGlobalUnifiedManager 获取全局统一管理器
func GetGlobalUnifiedManager() UnifiedManagerInterface {
	globalManagerMutex.RLock()
	defer globalManagerMutex.RUnlock()
	return globalUnifiedManager
}

// InputData 标准MVC风格输入数据结构
// 现在优先使用统一管理器，向后兼容session包扩展
type InputData struct {
	ctx       *Context
	extension *session.ContextExtension // session包扩展
	unifiedManager UnifiedManagerInterface // 统一管理器接口
}

// OutputData 标准MVC风格输出数据结构
// 现在优先使用统一管理器，向后兼容session包扩展
type OutputData struct {
	ctx       *Context
	extension *session.ContextExtension // session包扩展
	unifiedManager UnifiedManagerInterface // 统一管理器接口
}

// ============= OutputData基础方法 =============

// Cookie 设置Cookie (Output兼容性方法)
func (o *OutputData) Cookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if o.ctx.request != nil {
		o.ctx.request.SetCookie(name, value, maxAge, path, domain, protocol.CookieSameSiteDefaultMode, secure, httpOnly)
	}
}

// SetStatus 设置响应状态码 (Output兼容性方法)
func (o *OutputData) SetStatus(code int) {
	if o.ctx.request != nil {
		o.ctx.request.SetStatusCode(code)
	}
}

// JSON 输出JSON响应 (Output兼容性方法)
func (o *OutputData) JSON(data any) error {
	if o.ctx.request != nil {
		o.ctx.request.JSON(http.StatusOK, data)
	}
	return nil
}

// Header 设置响应头 (Output兼容性方法)
func (o *OutputData) Header(key, value string) {
	if o.ctx.request != nil {
		o.ctx.request.Response.Header.Set(key, value)
	}
}

// ============= OutputData文件下载和服务方法 =============

// Download 文件下载 (beego兼容方法)
// 支持自定义文件名、Range请求、断点续传等高级功能
func (o *OutputData) Download(file string, filename ...string) error {
	if o.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(file)
	if err != nil {
		o.SetStatus(http.StatusNotFound)
		return fmt.Errorf("file not found: %s", file)
	}

	// 检查是否为目录
	if fileInfo.IsDir() {
		o.SetStatus(http.StatusForbidden)
		return fmt.Errorf("cannot download directory: %s", file)
	}

	// 确定文件名
	var displayName string
	if len(filename) > 0 && filename[0] != "" {
		displayName = filename[0]
	} else {
		displayName = filepath.Base(file)
	}

	// RFC 6266 兼容的 Content-Disposition
	escaped := url.PathEscape(displayName) // utf-8''<escaped>
	// 始终加引号，避免特殊字符问题
	var contentDisposition string
	if displayName == escaped {
		// 纯 ASCII
		contentDisposition = fmt.Sprintf("attachment; filename=\"%s\"", escaped)
	} else {
		// 同时提供 filename 与 filename*
		contentDisposition = fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", sanitizeToken(displayName), escaped)
	}

	// 设置下载响应头（不预设 Content-Length，交由底层处理以支持 Range）
	o.Header("Content-Disposition", contentDisposition)
	o.Header("Content-Description", "File Transfer")
	o.Header("Content-Type", "application/octet-stream")
	o.Header("Content-Transfer-Encoding", "binary")
	o.Header("Expires", "0")
	// 兼容 beego 的缓存控制
	o.Header("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
	o.Header("Pragma", "public")

	// Last-Modified 可由底层设置，这里显式设置也无妨
	o.Header("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))

	// 交给 hertz 文件服务（内部处理 Range/断点续传/长度）
	o.ctx.request.File(file)
	return nil
}

// ServeFile 静态文件服务 (beego兼容方法)
func (o *OutputData) ServeFile(file string) error {
	if o.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}

	// 检查文件是否存在
	_, err := os.Stat(file)
	if err != nil {
		o.SetStatus(http.StatusNotFound)
		return fmt.Errorf("file not found: %s", file)
	}

	// 自动检测MIME类型
	mimeType := mime.TypeByExtension(filepath.Ext(file))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	o.Header("Content-Type", mimeType)

	// 使用hertz的文件服务
	o.ctx.request.File(file)
	return nil
}

// ServeFileDownload 强制下载文件 (便捷方法)
func (o *OutputData) ServeFileDownload(file, displayName string) error {
	return o.Download(file, displayName)
}

// ============= OutputData格式化输出方法 =============

// XML 输出XML (beego兼容方法)
func (o *OutputData) XML(data any, hasIndent ...bool) error {
	if o.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}

	o.Header("Content-Type", "application/xml; charset=utf-8")

	var content []byte
	var err error
	if len(hasIndent) > 0 && hasIndent[0] {
		content, err = xml.MarshalIndent(data, "", "  ")
	} else {
		content, err = xml.Marshal(data)
	}

	if err != nil {
		o.SetStatus(http.StatusInternalServerError)
		return err
	}

	return o.Body(content)
}

// YAML 输出YAML (beego兼容方法)
func (o *OutputData) YAML(data any) error {
	if o.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}

	o.Header("Content-Type", "application/x-yaml; charset=utf-8")

	content, err := yaml.Marshal(data)
	if err != nil {
		o.SetStatus(http.StatusInternalServerError)
		return err
	}

	return o.Body(content)
}

// JSONP 输出JSONP (beego兼容方法)
func (o *OutputData) JSONP(data any, callback string, hasIndent ...bool) error {
	if o.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}

	// 如果没有提供callback，从query参数获取
	if callback == "" {
		callback = string(o.ctx.request.QueryArgs().Peek("callback"))
		if callback == "" {
			callback = string(o.ctx.request.QueryArgs().Peek("jsonp"))
		}
	}

	// 验证callback函数名（安全性检查）
	if callback != "" && !isValidJSONPCallback(callback) {
		o.SetStatus(http.StatusBadRequest)
		return fmt.Errorf("invalid callback function name")
	}

	o.Header("Content-Type", "application/javascript; charset=utf-8")

	var jsonContent []byte
	var err error
	if len(hasIndent) > 0 && hasIndent[0] {
		jsonContent, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonContent, err = json.Marshal(data)
	}

	if err != nil {
		o.SetStatus(http.StatusInternalServerError)
		return err
	}

	var content []byte
	if callback != "" {
		content = []byte(fmt.Sprintf("%s(%s);", callback, string(jsonContent)))
	} else {
		// 如果没有callback，返回普通JSON
		o.Header("Content-Type", "application/json; charset=utf-8")
		content = jsonContent
	}

	return o.Body(content)
}

// ServeFormatted 根据Accept头自动选择格式 (beego兼容方法)
func (o *OutputData) ServeFormatted(data any, hasIndent ...bool) error {
	if o.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}

	acceptHeader := string(o.ctx.request.Request.Header.Peek("Accept"))

	// 根据Accept头选择格式
	switch {
	case strings.Contains(acceptHeader, "application/xml"), strings.Contains(acceptHeader, "text/xml"):
		return o.XML(data, hasIndent...)
	case strings.Contains(acceptHeader, "application/x-yaml"), strings.Contains(acceptHeader, "text/yaml"):
		return o.YAML(data)
	case strings.Contains(acceptHeader, "application/javascript"):
		return o.JSONP(data, "", hasIndent...)
	default:
		// 默认返回JSON
		if len(hasIndent) > 0 && hasIndent[0] {
			o.ctx.request.IndentedJSON(http.StatusOK, data)
		} else {
			o.ctx.request.JSON(http.StatusOK, data)
		}
		return nil
	}
}

// ============= OutputData重定向方法 =============

// Redirect HTTP重定向 (beego兼容方法)
func (o *OutputData) Redirect(localurl string, code ...int) {
	if o.ctx.request == nil {
		return
	}

	statusCode := http.StatusFound // 默认302
	if len(code) > 0 {
		statusCode = code[0]
	}

	// 验证重定向状态码
	if !isRedirectCode(statusCode) {
		statusCode = http.StatusFound
	}

	o.ctx.request.Redirect(statusCode, []byte(localurl))
}

// RedirectTemp 临时重定向 (便捷方法)
func (o *OutputData) RedirectTemp(localurl string) {
	o.Redirect(localurl, http.StatusTemporaryRedirect) // 307
}

// RedirectPermanent 永久重定向 (便捷方法)
func (o *OutputData) RedirectPermanent(localurl string) {
	o.Redirect(localurl, http.StatusMovedPermanently) // 301
}

// ============= OutputData状态检查方法 =============

// IsRedirect 检查是否为重定向状态 (beego兼容方法)
func (o *OutputData) IsRedirect() bool {
	if o.ctx.request == nil {
		return false
	}
	code := o.ctx.request.Response.StatusCode()
	return isRedirectCode(code)
}

// IsForbidden 检查是否为403状态 (beego兼容方法)
func (o *OutputData) IsForbidden() bool {
	if o.ctx.request == nil {
		return false
	}
	return o.ctx.request.Response.StatusCode() == http.StatusForbidden
}

// IsNotFound 检查是否为404状态 (beego兼容方法)
func (o *OutputData) IsNotFound() bool {
	if o.ctx.request == nil {
		return false
	}
	return o.ctx.request.Response.StatusCode() == http.StatusNotFound
}

// IsSuccessful 检查是否为2xx状态 (beego兼容方法)
func (o *OutputData) IsSuccessful() bool {
	if o.ctx.request == nil {
		return false
	}
	code := o.ctx.request.Response.StatusCode()
	return code >= 200 && code < 300
}

// IsClientError 检查是否为4xx状态 (便捷方法)
func (o *OutputData) IsClientError() bool {
	if o.ctx.request == nil {
		return false
	}
	code := o.ctx.request.Response.StatusCode()
	return code >= 400 && code < 500
}

// IsServerError 检查是否为5xx状态 (便捷方法)
func (o *OutputData) IsServerError() bool {
	if o.ctx.request == nil {
		return false
	}
	code := o.ctx.request.Response.StatusCode()
	return code >= 500 && code < 600
}

// ============= OutputData内容处理方法 =============

// Body 输出原始内容 (beego兼容方法)
func (o *OutputData) Body(content []byte) error {
	if o.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}
	_, err := o.ctx.request.Write(content)
	return err
}

// WriteString 输出字符串 (便捷方法)
func (o *OutputData) WriteString(s string) error {
	return o.Body([]byte(s))
}

// Write 实现io.Writer接口
func (o *OutputData) Write(p []byte) (n int, err error) {
	if o.ctx.request == nil {
		return 0, fmt.Errorf("request context is nil")
	}
	return o.ctx.request.Write(p)
}

// Flush 强制刷新输出缓冲区
func (o *OutputData) Flush() {
	if o.ctx.request != nil {
		o.ctx.request.Flush()
	}
}

// ============= OutputData高级功能方法 =============

// SetContentType 设置Content-Type (便捷方法)
func (o *OutputData) SetContentType(contentType string) {
	o.Header("Content-Type", contentType)
}

// EnableCORS 启用CORS (便捷方法)
func (o *OutputData) EnableCORS(origin ...string) {
	if len(origin) > 0 {
		o.Header("Access-Control-Allow-Origin", origin[0])
	} else {
		o.Header("Access-Control-Allow-Origin", "*")
	}
	o.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	o.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
}

// SetCache 设置缓存控制 (便捷方法)
func (o *OutputData) SetCache(maxAge int) {
	if maxAge <= 0 {
		o.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		o.Header("Pragma", "no-cache")
		o.Header("Expires", "0")
	} else {
		o.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
		expires := time.Now().Add(time.Duration(maxAge) * time.Second)
		o.Header("Expires", expires.UTC().Format(http.TimeFormat))
	}
}

// SetETag 设置ETag (便捷方法)
func (o *OutputData) SetETag(etag string) {
	if etag != "" {
		o.Header("ETag", fmt.Sprintf(`"%s"`, etag))
	}
}

// SetContentSecurityPolicy 设置CSP (安全方法)
func (o *OutputData) SetContentSecurityPolicy(policy string) {
	o.Header("Content-Security-Policy", policy)
}

/*
Attachment 内存数据下载（便捷方法）
- content: 需要下载的字节数据
- name: 下载文件名
- contentType: 可选，默认 application/octet-stream
*/
func (o *OutputData) Attachment(content []byte, name string, contentType ...string) error {
	if o.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}
	ct := "application/octet-stream"
	if len(contentType) > 0 && strings.TrimSpace(contentType[0]) != "" {
		ct = contentType[0]
	}
	escaped := url.PathEscape(name)
	var cd string
	if name == escaped {
		cd = fmt.Sprintf("attachment; filename=\"%s\"", escaped)
	} else {
		cd = fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", sanitizeToken(name), escaped)
	}
	o.Header("Content-Type", ct)
	o.Header("Content-Disposition", cd)
	o.Header("Content-Description", "File Transfer")
	o.Header("Content-Transfer-Encoding", "binary")
	o.Header("Expires", "0")
	o.Header("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
	o.Header("Pragma", "public")
	return o.Body(content)
}

// ============= OutputData辅助函数 =============

func sanitizeToken(s string) string {
	// 简单清洗，移除换行与分号，避免注入
	s = strings.ReplaceAll(s, "\r", "_")
	s = strings.ReplaceAll(s, "\n", "_")
	s = strings.ReplaceAll(s, ";", "_")
	return s
}

// isValidJSONPCallback 验证JSONP回调函数名
func isValidJSONPCallback(callback string) bool {
	if callback == "" {
		return false
	}
	// 简单的JavaScript标识符验证
	for i, r := range callback {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == '$') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$' || r == '.') {
				return false
			}
		}
	}
	return true
}

// isRedirectCode 检查是否为重定向状态码
func isRedirectCode(code int) bool {
	return code == http.StatusMovedPermanently || // 301
		code == http.StatusFound || // 302
		code == http.StatusSeeOther || // 303
		code == http.StatusTemporaryRedirect || // 307
		code == http.StatusPermanentRedirect // 308
}

// ============= OutputData Session方法 =============

// getExtension 获取session扩展（延迟初始化）(OutputData版本)
// 现在优先使用统一管理器，如果不可用则回退到session包扩展
func (o *OutputData) getExtension() *session.ContextExtension {
	// 1. 优先尝试使用统一管理器
	if o.unifiedManager == nil {
		o.unifiedManager = GetGlobalUnifiedManager()
	}
	
	// 如果统一管理器可用，优先使用它
	if o.unifiedManager != nil && o.unifiedManager.IsInitialized() {
		// 统一管理器已可用，但我们仍需要extension来保持兼容性
		// 这里创建一个基于统一管理器的扩展适配器
		if o.extension == nil && o.ctx != nil && o.ctx.request != nil {
			o.extension = session.NewExtensionForHertzContext(o.ctx.request)
		}
		return o.extension
	}
	
	// 2. 回退到传统的session包扩展
	if o.extension == nil && o.ctx != nil && o.ctx.request != nil {
		// 创建session扩展
		o.extension = session.NewExtensionForHertzContext(o.ctx.request)
	}
	return o.extension
}

// SetSession 设置session数据 (Output兼容性方法) - 代理到session包
func (o *OutputData) SetSession(key string, value any) error {
	if ext := o.getExtension(); ext != nil {
		return ext.SetSession(key, value)
	}
	return nil
}

// GetSession 获取session数据 (Output兼容性方法) - 代理到session包
func (o *OutputData) GetSession(key string) any {
	if ext := o.getExtension(); ext != nil {
		return ext.GetSession(key)
	}
	return nil
}

// Session 获取或设置session数据的便捷方法 (OutputData)
// 用法：
//
//	value := ctx.Output.Session("adminId")          // 获取值
//	ctx.Output.Session("adminId", "12345")          // 设置值
func (o *OutputData) Session(key string, values ...any) any {
	if len(values) == 0 {
		// 获取模式
		return o.GetSession(key)
	} else {
		// 设置模式
		o.SetSession(key, values[0])
		return values[0]
	}
}

// ============= InputData基础方法 =============

// Scheme 获取请求协议
func (i *InputData) Scheme() string {
	if i.ctx.request != nil {
		return string(i.ctx.request.URI().Scheme())
	}
	return "http"
}

// Domain 获取请求域名 (Input兼容性方法)
func (i *InputData) Domain() string {
	if i.ctx.request != nil {
		return string(i.ctx.request.Host())
	}
	return ""
}

// Host 获取请求主机 (Input兼容性方法)
func (i *InputData) Host() string {
	return i.Domain()
}

// Method 获取请求方法 (Input兼容性方法)
func (i *InputData) Method() string {
	if i.ctx.request != nil {
		return string(i.ctx.request.Method())
	}
	return "GET"
}

// IP 获取客户端IP地址 (Input兼容性方法)
func (i *InputData) IP() string {
	if i.ctx.request != nil {
		return i.ctx.request.ClientIP()
	}
	return ""
}

// UserAgent 获取User-Agent (Input兼容性方法)
func (i *InputData) UserAgent() string {
	if i.ctx.request != nil {
		return string(i.ctx.request.Request.Header.Peek("User-Agent"))
	}
	return ""
}

// IsAjax 判断是否为Ajax请求 (Input兼容性方法)
func (i *InputData) IsAjax() bool {
	if i.ctx.request != nil {
		return string(i.ctx.request.Request.Header.Peek("X-Requested-With")) == "XMLHttpRequest"
	}
	return false
}

// URL 获取请求URL (Input兼容性方法)
func (i *InputData) URL() string {
	if i.ctx.request != nil {
		return i.ctx.request.URI().String()
	}
	return ""
}

// Is 检查请求方法 (Input兼容性方法)
func (i *InputData) Is(method string) bool {
	return i.Method() == method
}

// IsGet 检查是否为GET请求 (Input兼容性方法)
func (i *InputData) IsGet() bool {
	return i.Is("GET")
}

// IsPost 检查是否为POST请求 (Input兼容性方法)
func (i *InputData) IsPost() bool {
	return i.Is("POST")
}

// Query 获取查询参数 (Input兼容性方法)
func (i *InputData) Query(key string) string {
	if i.ctx.request != nil {
		return string(i.ctx.request.QueryArgs().Peek(key))
	}
	return ""
}

// Param 获取路径参数 (Input兼容性方法)
func (i *InputData) Param(key string) string {
	if i.ctx.request != nil {
		return i.ctx.request.Param(key)
	}
	return ""
}

// Header 获取请求头 (Input兼容性方法)
func (i *InputData) Header(key string) string {
	if i.ctx.request != nil {
		return string(i.ctx.request.GetHeader(key))
	}
	return ""
}

// Referer 获取来源页面 (Input兼容性方法)
func (i *InputData) Referer() string {
	return i.Header("Referer")
}

// Data 设置上下文数据 (Input兼容性方法)
func (i *InputData) Data(key string, val any) {
	if i.ctx != nil {
		i.ctx.Set(key, val)
	}
}

// GetData 获取上下文数据 (Input兼容性方法)
func (i *InputData) GetData(key string) any {
	if i.ctx != nil {
		if val, exists := i.ctx.Get(key); exists {
			return val
		}
	}
	return nil
}

// ============= BeegoInput 实用方法（移植） =============

// GetString 返回参数的字符串值，按 Query -> PostForm -> Path Param 顺序查找
func (i *InputData) GetString(key string, def ...string) string {
	if s, ok := i.getFirstParam(key); ok {
		return s
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// DefaultQuery 获取查询参数，若为空返回默认值
func (i *InputData) DefaultQuery(key, d string) string {
	v := i.Query(key)
	if v == "" {
		return d
	}
	return v
}

// FormValue 获取表单参数（application/x-www-form-urlencoded 或 multipart/form-data）
func (i *InputData) FormValue(key string) string {
	if i.ctx.request != nil {
		return string(i.ctx.request.PostArgs().Peek(key))
	}
	return ""
}

// GetStrings 获取同名参数的所有值（同时来自 Query 与 PostForm）
func (i *InputData) GetStrings(key string) []string {
	var res []string
	res = append(res, i.getQueryMulti(key)...)
	res = append(res, i.getPostMulti(key)...)
	// 如果没有多值，尝试路径参数
	if len(res) == 0 {
		if p := i.Param(key); p != "" {
			res = append(res, p)
		}
	}
	return res
}

// GetBool 获取布尔值，支持 true/false, 1/0, on/off, yes/no。若为空且提供默认值则返回默认值
func (i *InputData) GetBool(key string, def ...bool) (bool, error) {
	s, ok := i.getFirstParam(key)
	if !ok || strings.TrimSpace(s) == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return false, nil
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "t", "true", "on", "yes", "y":
		return true, nil
	case "0", "f", "false", "off", "no", "n":
		return false, nil
	default:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return false, err
		}
		return b, nil
	}
}

// GetInt 获取整数值（int64 表示），若为空且提供默认值则返回默认值
func (i *InputData) GetInt(key string, def ...int) (int64, error) {
	s, ok := i.getFirstParam(key)
	if !ok || strings.TrimSpace(s) == "" {
		if len(def) > 0 {
			return int64(def[0]), nil
		}
		return 0, nil
	}
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

// GetInt64 获取整数值（int64），若为空且提供默认值则返回默认值
func (i *InputData) GetInt64(key string, def ...int64) (int64, error) {
	s, ok := i.getFirstParam(key)
	if !ok || strings.TrimSpace(s) == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

// GetFloat 获取浮点值（float64），若为空且提供默认值则返回默认值
func (i *InputData) GetFloat(key string, def ...float64) (float64, error) {
	s, ok := i.getFirstParam(key)
	if !ok || strings.TrimSpace(s) == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

// Body 返回原始请求体（与 beego.Input.Body 类似）
func (i *InputData) Body() []byte {
	if i.ctx.request == nil {
		return nil
	}
	return i.ctx.request.Request.Body()
}

// BindJSON 解析 JSON 请求体到目标对象
func (i *InputData) BindJSON(v any) error {
	if i.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}
	b := i.Body()
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, v)
}

// GetFileHeader 获取上传文件头
func (i *InputData) GetFileHeader(key string) (*multipart.FileHeader, error) {
	if i.ctx.request == nil {
		return nil, fmt.Errorf("request context is nil")
	}
	return i.ctx.request.FormFile(key)
}

	// SaveUploadedFile 保存上传文件到指定路径
func (i *InputData) SaveUploadedFile(header *multipart.FileHeader, dst string) error {
	if i.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}
	return i.ctx.request.SaveUploadedFile(header, dst)
}

// GetFile 获取上传文件（beego.Input 兼容方法）
// 返回 multipart.File（调用方负责 Close）与 *multipart.FileHeader
func (i *InputData) GetFile(key string) (multipart.File, *multipart.FileHeader, error) {
	if i.ctx.request == nil {
		return nil, nil, fmt.Errorf("request context is nil")
	}
	fh, err := i.ctx.request.FormFile(key)
	if err != nil {
		return nil, nil, err
	}
	f, err := fh.Open()
	if err != nil {
		return nil, nil, err
	}
	return f, fh, nil
}

// SaveToFile 保存指定表单字段的上传文件到目标路径（beego.Input 兼容方法）
func (i *InputData) SaveToFile(form, saveTo string) error {
	if i.ctx.request == nil {
		return fmt.Errorf("request context is nil")
	}
	fh, err := i.ctx.request.FormFile(form)
	if err != nil {
		return err
	}
	return i.ctx.request.SaveUploadedFile(fh, saveTo)
}

// -------- 内部辅助 --------

// getFirstParam 获取 key 的首个值，按 Query -> PostForm -> Path Param 顺序
func (i *InputData) getFirstParam(key string) (string, bool) {
	if i.ctx.request == nil {
		return "", false
	}
	if v := string(i.ctx.request.QueryArgs().Peek(key)); v != "" {
		return v, true
	}
	if v := string(i.ctx.request.PostArgs().Peek(key)); v != "" {
		return v, true
	}
	if v := i.Param(key); v != "" {
		return v, true
	}
	return "", false
}

// getQueryMulti 收集 QueryString 中 key 的所有值
func (i *InputData) getQueryMulti(key string) []string {
	if i.ctx.request == nil {
		return nil
	}
	var res []string
	i.ctx.request.QueryArgs().VisitAll(func(k, v []byte) {
		if string(k) == key {
			res = append(res, string(v))
		}
	})
	return res
}

// getPostMulti 收集 PostForm 中 key 的所有值
func (i *InputData) getPostMulti(key string) []string {
	if i.ctx.request == nil {
		return nil
	}
	var res []string
	i.ctx.request.PostArgs().VisitAll(func(k, v []byte) {
		if string(k) == key {
			res = append(res, string(v))
		}
	})
	return res
}

// ============= Cookie方法代理 (代理到session包) =============

// getExtension 获取session扩展（延迟初始化）
// 现在优先使用统一管理器，如果不可用则回退到session包扩展
func (i *InputData) getExtension() *session.ContextExtension {
	// 1. 优先尝试使用统一管理器
	if i.unifiedManager == nil {
		i.unifiedManager = GetGlobalUnifiedManager()
	}
	
	// 如果统一管理器可用，优先使用它
	if i.unifiedManager != nil && i.unifiedManager.IsInitialized() {
		// 统一管理器已可用，但我们仍需要extension来保持兼容性
		// 这里创建一个基于统一管理器的扩展适配器
		if i.extension == nil && i.ctx != nil && i.ctx.request != nil {
			i.extension = session.NewExtensionForHertzContext(i.ctx.request)
		}
		return i.extension
	}
	
	// 2. 回退到传统的session包扩展
	if i.extension == nil && i.ctx != nil && i.ctx.request != nil {
		// 创建session扩展
		i.extension = session.NewExtensionForHertzContext(i.ctx.request)
	}
	return i.extension
}

// Cookie 获取Cookie (Input兼容性方法) - 代理到session包
func (i *InputData) Cookie(key string) string {
	if ext := i.getExtension(); ext != nil {
		return ext.GetCookie(key)
	}
	return ""
}

// SetCookie 设置Cookie (beego兼容方法) - 代理到session包
func (i *InputData) SetCookie(name, value string, others ...any) {
	if ext := i.getExtension(); ext != nil {
		ext.SetCookie(name, value, others...)
	}
}

// GetSecureCookie 获取安全Cookie (beego兼容方法) - 代理到session包
func (i *InputData) GetSecureCookie(secret, key string) (string, bool) {
	if ext := i.getExtension(); ext != nil {
		return ext.GetSecureCookie(secret, key)
	}
	return "", false
}

// SetSecureCookie 设置安全Cookie (beego兼容方法) - 代理到session包
func (i *InputData) SetSecureCookie(secret, name, value string, others ...any) {
	if ext := i.getExtension(); ext != nil {
		ext.SetSecureCookie(secret, name, value, others...)
	}
}

// DelCookie 删除Cookie (beego兼容方法) - 代理到session包
func (i *InputData) DelCookie(name string) {
	if ext := i.getExtension(); ext != nil {
		ext.DelCookie(name)
	}
}

// CookieExists 检查Cookie是否存在 - 代理到session包
func (i *InputData) CookieExists(key string) bool {
	if ext := i.getExtension(); ext != nil {
		return ext.CookieExists(key)
	}
	return false
}

// GetCookies 获取所有Cookie - 代理到session包
func (i *InputData) GetCookies() map[string]string {
	if ext := i.getExtension(); ext != nil && ext.Cookie != nil {
		return ext.Cookie.GetAll()
	}
	return make(map[string]string)
}

// ClearAllCookies 清除所有Cookie - 代理到session包
func (i *InputData) ClearAllCookies() {
	if ext := i.getExtension(); ext != nil && ext.Cookie != nil {
		ext.Cookie.Clear()
	}
}

// CookieCount 获取Cookie数量 - 代理到session包
func (i *InputData) CookieCount() int {
	if ext := i.getExtension(); ext != nil && ext.Cookie != nil {
		return ext.Cookie.Count()
	}
	return 0
}

// ValidateSecureCookie 验证安全Cookie但不返回值 - 代理到session包
func (i *InputData) ValidateSecureCookie(secret, key string) bool {
	if ext := i.getExtension(); ext != nil && ext.SecureCookie != nil {
		return ext.SecureCookie.Validate(secret, key)
	}
	return false
}

// ============= 高级Cookie功能代理 =============

// SetSecureCookieWithOptions 设置安全Cookie (增强版本) - 代理到session包
func (i *InputData) SetSecureCookieWithOptions(name, value string, options cookie.CookieSecurityOptions, others ...any) error {
	if ext := i.getExtension(); ext != nil && ext.SecureCookie != nil {
		return ext.SecureCookie.SetSecureWithOptions(name, value, options, others...)
	}
	return nil
}

// GetSecureCookieWithOptions 获取安全Cookie (增强版本) - 代理到session包
func (i *InputData) GetSecureCookieWithOptions(key string, options cookie.CookieSecurityOptions) (string, bool, error) {
	if ext := i.getExtension(); ext != nil && ext.SecureCookie != nil {
		return ext.SecureCookie.GetSecureWithOptions(key, options)
	}
	return "", false, nil
}

// ============= Session方法代理 (代理到session包) =============

// StartSession 启动Session (标准MVC风格接口) - 代理到session包
func (i *InputData) StartSession() *session.Adapter {
	if ext := i.getExtension(); ext != nil {
		return ext.StartSession()
	}
	return session.NewAdapter(nil, i.ctx)
}

// SetSession 设置session数据 (标准MVC兼容性方法) - 代理到session包
func (i *InputData) SetSession(key string, value any) error {
	if ext := i.getExtension(); ext != nil {
		return ext.SetSession(key, value)
	}
	return nil
}

// GetSession 获取session数据 (标准MVC兼容性方法) - 代理到session包
func (i *InputData) GetSession(key string) any {
	if ext := i.getExtension(); ext != nil {
		return ext.GetSession(key)
	}
	return nil
}

// DelSession 删除session数据 (标准MVC兼容性方法) - 代理到session包
func (i *InputData) DelSession(key string) error {
	if ext := i.getExtension(); ext != nil {
		return ext.DelSession(key)
	}
	return nil
}

// GetSessionID 获取session ID (beego兼容性方法) - 代理到session包
func (i *InputData) GetSessionID() string {
	if ext := i.getExtension(); ext != nil {
		return ext.GetSessionID()
	}
	return ""
}

// IsSessionStarted 检查session是否已启动 (beego兼容性方法) - 代理到session包
func (i *InputData) IsSessionStarted() bool {
	if ext := i.getExtension(); ext != nil {
		return ext.IsSessionStarted()
	}
	return false
}

// DestroySession 销毁session (标准MVC兼容性方法) - 代理到session包
func (i *InputData) DestroySession() {
	if ext := i.getExtension(); ext != nil {
		ext.DestroySession()
	}
}

// ClearSession 清空session数据 (标准MVC兼容性方法) - 代理到session包
func (i *InputData) ClearSession() {
	if ext := i.getExtension(); ext != nil {
		ext.ClearSession()
	}
}

// SaveSession 保存session数据 (beego兼容性方法) - 代理到session包
func (i *InputData) SaveSession() error {
	if ext := i.getExtension(); ext != nil {
		return ext.SaveSession()
	}
	return nil
}

// SessionRegenerateID 重新生成session ID (标准MVC兼容性方法) - 代理到session包
func (i *InputData) SessionRegenerateID() {
	if ext := i.getExtension(); ext != nil {
		ext.SessionRegenerateID()
	}
}

// Session 获取或设置session数据的便捷方法 (InputData)
// 用法：
//
//	value := ctx.Input.Session("adminId")          // 获取值
//	ctx.Input.Session("adminId", "12345")          // 设置值
func (i *InputData) Session(key string, values ...any) any {
	if len(values) == 0 {
		// 获取模式
		return i.GetSession(key)
	} else {
		// 设置模式
		i.SetSession(key, values[0])
		return values[0]
	}
}

// ============= 兼容性别名和附加功能 =============

// GetSessionExtension 获取session扩展对象 (提供对session包完整功能的访问)
func (i *InputData) GetSessionExtension() *session.ContextExtension {
	return i.getExtension()
}

// WithSessionExtension 使用自定义session扩展
func (i *InputData) WithSessionExtension(ext *session.ContextExtension) *InputData {
	i.extension = ext
	return i
}

// ============= 统一管理器便捷方法 (InputData) =============

// GetUnifiedManager 获取统一管理器实例
func (i *InputData) GetUnifiedManager() UnifiedManagerInterface {
	if i.unifiedManager == nil {
		i.unifiedManager = GetGlobalUnifiedManager()
	}
	return i.unifiedManager
}

// SetUnifiedCookie 使用统一管理器设置Cookie
// 优先使用统一管理器，如果不可用则回退到session包扩展
func (i *InputData) SetUnifiedCookie(name, value string, options ...any) {
	if i.unifiedManager == nil {
		i.unifiedManager = GetGlobalUnifiedManager()
	}
	
	if i.unifiedManager != nil && i.unifiedManager.IsInitialized() {
		// 使用统一管理器设置Cookie
		i.unifiedManager.SetCookie(i.ctx, name, value, options...)
	} else {
		// 回退到session包扩展
		i.SetCookie(name, value, options...)
	}
}

// GetUnifiedCookie 使用统一管理器获取Cookie
// 优先使用统一管理器，如果不可用则回退到session包扩展
func (i *InputData) GetUnifiedCookie(name string) string {
	if i.unifiedManager != nil && i.unifiedManager.IsInitialized() {
		// 使用统一管理器获取Cookie
		return i.unifiedManager.GetCookie(i.ctx, name)
	} else {
		// 回退到session包扩展
		return i.Cookie(name)
	}
}

// SetUnifiedSessionData 使用统一管理器设置Session数据
// 优先使用统一管理器，如果不可用则回退到session包扩展
func (i *InputData) SetUnifiedSessionData(key string, value any) error {
	if i.unifiedManager != nil && i.unifiedManager.IsInitialized() {
		// 使用统一管理器设置Session数据
		i.unifiedManager.SetSessionData(i.ctx, key, value)
		return nil
	} else {
		// 回退到session包扩展
		return i.SetSession(key, value)
	}
}

// GetUnifiedSessionData 使用统一管理器获取Session数据
// 优先使用统一管理器，如果不可用则回退到session包扩展
func (i *InputData) GetUnifiedSessionData(key string) any {
	if i.unifiedManager != nil && i.unifiedManager.IsInitialized() {
		// 使用统一管理器获取Session数据
		return i.unifiedManager.GetSessionData(i.ctx, key)
	} else {
		// 回退到session包扩展
		return i.GetSession(key)
	}
}

// SetUnifiedContextData 使用统一管理器设置上下文数据
func (i *InputData) SetUnifiedContextData(key string, value any) {
	if i.unifiedManager != nil && i.unifiedManager.IsInitialized() {
		i.unifiedManager.SetContextData(i.ctx, key, value)
	}
}

// GetUnifiedContextData 使用统一管理器获取上下文数据
func (i *InputData) GetUnifiedContextData(key string) any {
	if i.unifiedManager != nil && i.unifiedManager.IsInitialized() {
		return i.unifiedManager.GetContextData(i.ctx, key)
	}
	return nil
}

// ============= 向后兼容性类型别名 =============

// SessionStore 兼容性类型别名，指向session包的Adapter
type SessionStore = session.Adapter

// NewSessionStore 兼容性函数，创建session适配器
func NewSessionStore(store session.Store, ctx *Context) *SessionStore {
	return session.NewAdapter(store, ctx)
}

// CookieSecurityOptions 兼容性类型别名
type CookieSecurityOptions = cookie.CookieSecurityOptions

// ============= 初始化方法 =============

// Initialize 初始化InputData (兼容性方法)
// 用于在已创建的InputData实例上重新初始化Context
func (i *InputData) Initialize(c *app.RequestContext) {
	if i.ctx == nil {
		i.ctx = NewContext(c)
	} else {
		// 重置现有Context
		i.ctx.request = c
	}
	// 重置extension以便重新初始化
	i.extension = nil
	// 确保统一管理器可用
	if i.unifiedManager == nil {
		i.unifiedManager = GetGlobalUnifiedManager()
	}
}

// Initialize 初始化OutputData (兼容性方法)
// 用于在已创建的OutputData实例上重新初始化Context
func (o *OutputData) Initialize(c *app.RequestContext) {
	if o.ctx == nil {
		o.ctx = NewContext(c)
	} else {
		// 重置现有Context
		o.ctx.request = c
	}
	// 重置extension以便重新初始化
	o.extension = nil
	// 确保统一管理器可用
	if o.unifiedManager == nil {
		o.unifiedManager = GetGlobalUnifiedManager()
	}
}

// GetContext 获取关联的Context (InputData)
func (i *InputData) GetContext() *Context {
	return i.ctx
}

// GetContext 获取关联的Context (OutputData)
func (o *OutputData) GetContext() *Context {
	return o.ctx
}

// ============= 便利构造函数 =============

// NewInputData 创建InputData实例
func NewInputData(ctx *Context) *InputData {
	return &InputData{
		ctx: ctx,
		// extension 延迟初始化
		unifiedManager: GetGlobalUnifiedManager(), // 获取统一管理器实例
	}
}

// NewOutputData 创建OutputData实例
func NewOutputData(ctx *Context) *OutputData {
	return &OutputData{
		ctx: ctx,
		// extension 延迟初始化
		unifiedManager: GetGlobalUnifiedManager(), // 获取统一管理器实例
	}
}

// ============= 统一管理器便捷方法 (OutputData) =============

// GetUnifiedManager 获取统一管理器实例 (OutputData版本)
func (o *OutputData) GetUnifiedManager() UnifiedManagerInterface {
	if o.unifiedManager == nil {
		o.unifiedManager = GetGlobalUnifiedManager()
	}
	return o.unifiedManager
}

// SetUnifiedCookie 使用统一管理器设置Cookie (OutputData版本)
// 优先使用统一管理器，如果不可用则回退到其他方法
func (o *OutputData) SetUnifiedCookie(name, value string, options ...any) {
	if o.unifiedManager == nil {
		o.unifiedManager = GetGlobalUnifiedManager()
	}
	
	if o.unifiedManager != nil && o.unifiedManager.IsInitialized() {
		// 使用统一管理器设置Cookie
		o.unifiedManager.SetCookie(o.ctx, name, value, options...)
	} else {
		// 回退到传统方法（通过Context的Request设置）
		if o.ctx.request != nil {
			// 解析options参数
			maxAge := 0
			path := "/"
			domain := ""
			secure := false
			httpOnly := false
			
			// 简单的options处理（可以根据需要扩展）
			if len(options) > 0 {
				if maxAgeVal, ok := options[0].(int); ok {
					maxAge = maxAgeVal
				}
			}
			o.ctx.request.SetCookie(name, value, maxAge, path, domain, protocol.CookieSameSiteDefaultMode, secure, httpOnly)
		}
	}
}

// GetUnifiedCookie 使用统一管理器获取Cookie (OutputData版本)
// 优先使用统一管理器，如果不可用则回退到其他方法
func (o *OutputData) GetUnifiedCookie(name string) string {
	if o.unifiedManager != nil && o.unifiedManager.IsInitialized() {
		// 使用统一管理器获取Cookie
		return o.unifiedManager.GetCookie(o.ctx, name)
	} else {
		// 回退到从Request直接获取
		if o.ctx.request != nil {
			return string(o.ctx.request.Cookie(name))
		}
		return ""
	}
}

// SetUnifiedContextData 使用统一管理器设置上下文数据 (OutputData版本)
func (o *OutputData) SetUnifiedContextData(key string, value any) {
	if o.unifiedManager != nil && o.unifiedManager.IsInitialized() {
		o.unifiedManager.SetContextData(o.ctx, key, value)
	}
}

// GetUnifiedContextData 使用统一管理器获取上下文数据 (OutputData版本)
func (o *OutputData) GetUnifiedContextData(key string) any {
	if o.unifiedManager != nil && o.unifiedManager.IsInitialized() {
		return o.unifiedManager.GetContextData(o.ctx, key)
	}
	return nil
}

// ============= 文档和迁移说明 =============

/*
迁移说明:

1. 原有代码100%兼容，无需修改
2. 新功能（推荐）：使用统一管理器
   // 获取统一管理器
   manager := inputData.GetUnifiedManager()
   
   // 使用统一管理器方法（推荐）
   inputData.SetUnifiedCookie("auth", "token")
   token := inputData.GetUnifiedCookie("auth")
   inputData.SetUnifiedSessionData("user", userInfo)
   user := inputData.GetUnifiedSessionData("user")

3. 向后兼容：直接使用session包
   import "github.com/zsy619/yyhertz/framework/mvc/session"
   ext := session.NewContextExtension(ctx)

4. 初始化方法（兼容性）：
   inputData := &InputData{}
   outputData := &OutputData{}
   inputData.Initialize(hertzCtx)   // 初始化InputData
   outputData.Initialize(hertzCtx)  // 初始化OutputData
   ctx := inputData.GetContext()    // 获取关联的Context

5. 高级功能访问：
   // 使用统一管理器的高级功能（推荐）
   manager := inputData.GetUnifiedManager()
   token, err := manager.GenerateCSRFToken("123", "192.168.1.1")
   html, err := manager.RenderTemplate("user/profile", userData)
   
   // 或者使用session包的高级功能
   ext := inputData.GetSessionExtension()
   options := session.CookieSecurityOptions{...}
   ext.SecureCookie.SetSecureWithOptions(...)

6. 性能优化：
   - 统一管理器优先，提供更好的性能
   - session扩展延迟初始化，不使用时无性能开销
   - 自动管理器选择，避免重复初始化

7. 功能对照：
   // 新增统一管理器方法（推荐）
   InputData.SetUnifiedCookie() -> unified.Manager.SetCookie()
   InputData.GetUnifiedCookie() -> unified.Manager.GetCookie()
   InputData.SetUnifiedSessionData() -> unified.Manager.SetSessionData()
   InputData.GetUnifiedSessionData() -> unified.Manager.GetSessionData()
   
   // 向后兼容的session包方法
   InputData.Cookie() -> session.BaseCookie.Get()
   InputData.SetSecureCookie() -> session.SecureCookie.SetSecure()
   InputData.StartSession() -> session.ContextExtension.StartSession()

8. 新增方法：
   // 统一管理器方法
   - InputData.GetUnifiedManager() -> *unified.Manager
   - InputData.SetUnifiedCookie(name, value, ...options)
   - InputData.GetUnifiedCookie(name) -> string
   - InputData.SetUnifiedSessionData(key, value) -> error
   - InputData.GetUnifiedSessionData(key) -> any
   - InputData.SetUnifiedContextData(key, value)
   - InputData.GetUnifiedContextData(key) -> any
   - OutputData.GetUnifiedManager() -> *unified.Manager
   - OutputData.SetUnifiedCookie(name, value, ...options)
   - OutputData.GetUnifiedCookie(name) -> string
   - OutputData.SetUnifiedContextData(key, value)
   - OutputData.GetUnifiedContextData(key) -> any
   
   // 兼容性方法
   - InputData.Initialize(c *app.RequestContext)  // 初始化方法
   - OutputData.Initialize(c *app.RequestContext) // 初始化方法
   - InputData.GetContext() *Context              // 获取Context
   - OutputData.GetContext() *Context             // 获取Context

9. 便捷Session方法和统一管理器优化：
   InputData:
   - Session(key) -> GetSession(key)              // 获取session值
   - Session(key, value) -> SetSession(key, value) // 设置session值

   OutputData（新增完整session支持）:
   - Session(key) -> GetSession(key)              // 获取session值
   - Session(key, value) -> SetSession(key, value) // 设置session值
   - SetSession(key, value) error                 // 设置session数据
   - GetSession(key) any                  // 获取session数据

   使用示例（推荐使用统一管理器方法）：
   // 统一管理器方法（推荐）
   ctx.Input.SetUnifiedSessionData("adminId", "12345")     // 设置
   adminId := ctx.Input.GetUnifiedSessionData("adminId")   // 获取
   ctx.Input.SetUnifiedCookie("theme", "dark")           // 设置Cookie
   theme := ctx.Input.GetUnifiedCookie("theme")           // 获取Cookie
   
   // 向后兼容的便捷方法
   adminId := ctx.Input.Session("adminId")        // 获取
   ctx.Input.Session("adminId", "12345")          // 设置
   userId := ctx.Output.Session("userId")         // 获取
   ctx.Output.Session("userId", "67890")          // 设置

10. 智能管理器选择：
   - 优先使用统一管理器（如果已初始化）
   - 自动回退到session包扩展
   - 保证向后兼容性
   - 提供更好的性能和功能支持
*/
