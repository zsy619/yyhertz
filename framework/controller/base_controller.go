package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"hertz-controller/framework/types"
)

// BaseController 基础控制器结构
// 提供了标准的HTTP响应方法、模板渲染、参数绑定等功能
type BaseController struct {
	Ctx        *RequestContext       // HTTP请求上下文
	ViewPath   string                // 视图文件路径
	LayoutPath string                // 布局文件路径
	Data       map[string]any        // 模板数据
	logger     *log.Logger           // 日志记录器
}

// NewBaseController 创建新的基础控制器实例
func NewBaseController() *BaseController {
	return &BaseController{
		Data:       make(map[string]any),
		ViewPath:   "views",
		LayoutPath: "views/layout",
		logger:     log.Default(),
	}
}

// NewBaseControllerWithContext 使用指定上下文创建控制器
func NewBaseControllerWithContext(ctx *RequestContext) *BaseController {
	c := NewBaseController()
	c.Ctx = ctx
	return c
}

// Init 初始化控制器，确保数据结构正确初始化
func (c *BaseController) Init() {
	if c.Data == nil {
		c.Data = make(map[string]any)
	}
	if c.logger == nil {
		c.logger = log.Default()
	}
}

// Prepare 预处理方法，在处理请求前调用
func (c *BaseController) Prepare() {
	// 默认实现为空，子类可以重写
}

// Finish 后处理方法，在处理请求后调用
func (c *BaseController) Finish() {
	// 默认实现为空，子类可以重写
}

// JSON 返回JSON格式的数据
func (c *BaseController) JSON(data any) {
	c.JSONWithStatus(consts.StatusOK, data)
}

// JSONWithStatus 返回指定状态码的JSON数据
func (c *BaseController) JSONWithStatus(status int, data any) {
	if c.Ctx == nil {
		c.logger.Printf("Error: Context is nil when trying to return JSON")
		return
	}
	c.Ctx.JSON(status, data)
}

// JSONSuccess 返回成功的JSON响应
func (c *BaseController) JSONSuccess(message string, data any) {
	c.JSON(types.Success(message, data))
}

// JSONError 返回错误的JSON响应
func (c *BaseController) JSONError(message string) {
	c.JSON(types.Error(message))
}

// JSONPage 返回分页JSON响应
func (c *BaseController) JSONPage(message string, data any, count int64) {
	c.JSON(types.SuccessPage(message, data, count))
}

// String 返回字符串响应
func (c *BaseController) String(s string) {
	c.StringWithStatus(consts.StatusOK, s)
}

// StringWithStatus 返回指定状态码的字符串响应
func (c *BaseController) StringWithStatus(status int, s string) {
	if c.Ctx == nil {
		c.logger.Printf("Error: Context is nil when trying to return string")
		return
	}
	c.Ctx.String(status, s)
}

func (c *BaseController) Render(viewName string, data ...map[string]any) {
	if len(data) > 0 {
		for k, v := range data[0] {
			c.Data[k] = v
		}
	}
	
	c.RenderWithLayout(viewName, "")
}

func (c *BaseController) RenderWithLayout(viewName, layoutName string) {
	if layoutName == "" {
		layoutName = "layout.html"
	}
	
	layoutPath := filepath.Join(c.LayoutPath, layoutName)
	viewPath := filepath.Join(c.ViewPath, viewName)
	
	tmpl, err := template.ParseFiles(layoutPath, viewPath)
	if err != nil {
		c.Error(500, "模板解析错误: "+err.Error())
		return
	}
	
	c.Ctx.Header("Content-Type", "text/html; charset=utf-8")
	err = tmpl.ExecuteTemplate(c.Ctx, "layout", c.Data)
	if err != nil {
		c.Error(500, "模板渲染错误: "+err.Error())
		return
	}
}

func (c *BaseController) RenderHTML(viewName string, data ...map[string]any) {
	if len(data) > 0 {
		for k, v := range data[0] {
			c.Data[k] = v
		}
	}
	
	viewPath := filepath.Join(c.ViewPath, viewName)
	tmpl, err := template.ParseFiles(viewPath)
	if err != nil {
		c.Error(500, "模板解析错误: "+err.Error())
		return
	}
	
	c.Ctx.Header("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(c.Ctx, c.Data)
	if err != nil {
		c.Error(500, "模板渲染错误: "+err.Error())
		return
	}
}

func (c *BaseController) Redirect(url string, code ...int) {
	statusCode := consts.StatusFound
	if len(code) > 0 {
		statusCode = code[0]
	}
	c.Ctx.Redirect(statusCode, []byte(url))
}

func (c *BaseController) Error(code int, msg string) {
	c.Ctx.String(code, msg)
}

func (c *BaseController) SetData(key string, value any) {
	c.Data[key] = value
}

func (c *BaseController) GetString(key string, def ...string) string {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
	if val := c.Ctx.Query(key); val != "" {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func (c *BaseController) GetInt(key string, def ...int) int {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

func (c *BaseController) GetForm(key string, def ...string) string {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
	if val := c.Ctx.PostForm(key); val != "" {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// HTTP方法判断
func (c *BaseController) IsPost() bool {
	if c.Ctx == nil {
		return false
	}
	return string(c.Ctx.Method()) == "POST"
}

func (c *BaseController) IsGet() bool {
	if c.Ctx == nil {
		return false
	}
	return string(c.Ctx.Method()) == "GET"
}

// 获取去除空格的字符串
func (c *BaseController) GetStringTrim(key string, def ...string) string {
	val := c.GetString(key, def...)
	return strings.TrimSpace(val)
}

// 获取安全字符串(防XSS)
func (c *BaseController) GetSafeString(key string, def ...string) string {
	val := c.GetString(key, def...)
	// 简单的XSS防护
	val = strings.ReplaceAll(val, "<", "&lt;")
	val = strings.ReplaceAll(val, ">", "&gt;")
	val = strings.ReplaceAll(val, "\"", "&quot;")
	val = strings.ReplaceAll(val, "'", "&#39;")
	return val
}

// 获取客户端IP地址
func (c *BaseController) GetClientIP() string {
	if c.Ctx == nil {
		return ""
	}
	// 尝试从X-Forwarded-For获取真实IP
	if xff := c.Ctx.GetHeader("X-Forwarded-For"); len(xff) > 0 {
		xffStr := string(xff)
		ips := strings.Split(xffStr, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}
	
	// 尝试从X-Real-IP获取
	if xri := c.Ctx.GetHeader("X-Real-IP"); len(xri) > 0 {
		return string(xri)
	}
	
	// 从RemoteAddr获取
	remoteAddr := c.Ctx.RemoteAddr().String()
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// 获取POST JSON数据
func (c *BaseController) GetPostJSON() map[string]any {
	if c.Ctx == nil {
		return make(map[string]any)
	}
	body, err := c.Ctx.Body()
	if err != nil {
		return make(map[string]any)
	}
	
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return make(map[string]any)
	}
	return data
}

// 获取JSON中的特定字段
func (c *BaseController) GetJSON(key string) any {
	data := c.GetPostJSON()
	return data[key]
}

// 分页信息获取
func (c *BaseController) GetPageInfo() (page, pageSize int) {
	return c.GetPageInfoByParam("page", "limit", 20)
}

func (c *BaseController) GetPageInfoDefault(pageSizeDefault int) (page, pageSize int) {
	if pageSizeDefault <= 0 {
		pageSizeDefault = 20
	}
	return c.GetPageInfoByParam("page", "limit", pageSizeDefault)
}

func (c *BaseController) GetPageInfoByParam(pageParam, pageSizeParam string, pageSizeDefault int) (page, pageSize int) {
	if pageParam == "" {
		pageParam = "page"
	}
	if pageSizeParam == "" {
		pageSizeParam = "limit"
	}
	
	// 如果没有Context，返回默认值
	if c.Ctx == nil {
		page = 1
		pageSize = pageSizeDefault
	} else {
		page = c.GetInt(pageParam, 1)
		pageSize = c.GetInt(pageSizeParam, pageSizeDefault)
	}
	
	// 确保页码和页大小的合理性
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = pageSizeDefault
	}
	if pageSize > 1000 { // 防止页大小过大
		pageSize = 1000
	}
	return
}

// 时间戳生成
func (c *BaseController) CreateTime() int64 {
	rt := time.Now().Format("20060102150405")
	ok, _ := strconv.ParseInt(rt, 10, 64)
	return ok
}

func (c *BaseController) CreateDate() int {
	rt := time.Now().Format("20060102")
	ok, _ := strconv.Atoi(rt)
	return ok
}

func (c *BaseController) UpdateTime() int64 {
	return c.CreateTime()
}

func (c *BaseController) UpdateDate() int {
	return c.CreateDate()
}

// 批量设置数据
func (c *BaseController) SetDatas(data map[string]any) {
	for k, v := range data {
		c.Data[k] = v
	}
}

// 响应结构体定义
type JSONResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type JSONResponsePage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Count   int64  `json:"count"`
}

// === 增强的参数绑定方法 ===

// GetParam 获取路径参数
func (c *BaseController) GetParam(key string) string {
	if c.Ctx == nil {
		return ""
	}
	return c.Ctx.Param(key)
}

// GetPostForm 获取POST表单参数 
func (c *BaseController) GetPostForm(key string) string {
	if c.Ctx == nil {
		return ""
	}
	return c.Ctx.PostForm(key)
}

// GetInt64 获取int64类型参数
func (c *BaseController) GetInt64(key string, def ...int64) int64 {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetFloat 获取float64类型参数
func (c *BaseController) GetFloat(key string, def ...float64) float64 {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetBool 获取bool类型参数
func (c *BaseController) GetBool(key string, def ...bool) bool {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return false
	}
	if val := c.Ctx.Query(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
		// 支持更多的bool表示
		val = strings.ToLower(val)
		return val == "1" || val == "yes" || val == "on"
	}
	if len(def) > 0 {
		return def[0]
	}
	return false
}

// BindJSON 绑定JSON请求体到结构体
func (c *BaseController) BindJSON(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.BindJSON(obj)
}

// BindQuery 绑定查询参数到结构体
func (c *BaseController) BindQuery(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.BindQuery(obj)
}

// BindForm 绑定表单参数到结构体
func (c *BaseController) BindForm(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.BindForm(obj)
}

// === HTTP方法扩展 ===

// IsPut 判断是否为PUT请求
func (c *BaseController) IsPut() bool {
	if c.Ctx == nil {
		return false
	}
	return string(c.Ctx.Method()) == "PUT"
}

// IsDelete 判断是否为DELETE请求
func (c *BaseController) IsDelete() bool {
	if c.Ctx == nil {
		return false
	}
	return string(c.Ctx.Method()) == "DELETE"
}

// IsPatch 判断是否为PATCH请求
func (c *BaseController) IsPatch() bool {
	if c.Ctx == nil {
		return false
	}
	return string(c.Ctx.Method()) == "PATCH"
}

// IsAjax 判断是否为Ajax请求
func (c *BaseController) IsAjax() bool {
	if c.Ctx == nil {
		return false
	}
	xreq := c.Ctx.GetHeader("X-Requested-With")
	return string(xreq) == "XMLHttpRequest"
}

// === 文件上传处理 ===

// GetFile 获取上传的文件
func (c *BaseController) GetFile(key string) (*multipart.FileHeader, error) {
	if c.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	return c.Ctx.FormFile(key)
}

// GetFiles 获取多个上传的文件
func (c *BaseController) GetFiles(key string) ([]*multipart.FileHeader, error) {
	if c.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	form, err := c.Ctx.MultipartForm()
	if err != nil {
		return nil, err
	}
	return form.File[key], nil
}

// SaveFile 保存上传的文件
func (c *BaseController) SaveFile(file *multipart.FileHeader, dst string) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.SaveUploadedFile(file, dst)
}

// === 响应增强方法 ===

// JSONStatus 返回指定状态码的JSON响应
func (c *BaseController) JSONStatus(status int, code int, message string, data any) {
	response := map[string]any{
		"code":    code,
		"message": message,
		"data":    data,
	}
	c.JSONWithStatus(status, response)
}

// JSONOK 返回成功响应（200）
func (c *BaseController) JSONOK(message string, data any) {
	c.JSONStatus(200, 0, message, data)
}

// JSONBadRequest 返回400错误
func (c *BaseController) JSONBadRequest(message string) {
	c.JSONStatus(400, 400, message, nil)
}

// JSONUnauthorized 返回401错误
func (c *BaseController) JSONUnauthorized(message string) {
	c.JSONStatus(401, 401, message, nil)
}

// JSONForbidden 返回403错误
func (c *BaseController) JSONForbidden(message string) {
	c.JSONStatus(403, 403, message, nil)
}

// JSONNotFound 返回404错误
func (c *BaseController) JSONNotFound(message string) {
	c.JSONStatus(404, 404, message, nil)
}

// JSONInternalError 返回500错误
func (c *BaseController) JSONInternalError(message string) {
	c.JSONStatus(500, 500, message, nil)
}

// === 验证辅助方法 ===

// ValidateRequired 验证必填字段
func (c *BaseController) ValidateRequired(fields map[string]string) []string {
	var errors []string
	for field, value := range fields {
		if strings.TrimSpace(value) == "" {
			errors = append(errors, fmt.Sprintf("%s不能为空", field))
		}
	}
	return errors
}

// ValidateEmail 验证邮箱格式
func (c *BaseController) ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePhone 验证手机号格式
func (c *BaseController) ValidatePhone(phone string) bool {
	if phone == "" {
		return false
	}
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// === Cookie操作 ===

// SetCookie 设置Cookie
func (c *BaseController) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if c.Ctx == nil {
		return
	}
	cookie := fmt.Sprintf("%s=%s; Max-Age=%d; Path=%s", name, value, maxAge, path)
	if domain != "" {
		cookie += "; Domain=" + domain
	}
	if secure {
		cookie += "; Secure"
	}
	if httpOnly {
		cookie += "; HttpOnly"
	}
	c.Ctx.Header("Set-Cookie", cookie)
}

// GetCookie 获取Cookie
func (c *BaseController) GetCookie(name string) string {
	if c.Ctx == nil {
		return ""
	}
	return string(c.Ctx.Cookie(name))
}

// === 会话辅助方法 ===

// SetSession 设置会话数据（需要配合session中间件使用）
func (c *BaseController) SetSession(key string, value any) {
	if c.Data == nil {
		c.Data = make(map[string]any)
	}
	c.Data["session_"+key] = value
}

// GetSession 获取会话数据
func (c *BaseController) GetSession(key string) any {
	if c.Data == nil {
		return nil
	}
	return c.Data["session_"+key]
}

// === 调试和日志方法 ===

// LogInfo 记录信息日志
func (c *BaseController) LogInfo(format string, args ...any) {
	if c.logger != nil {
		c.logger.Printf("[INFO] "+format, args...)
	}
}

// LogError 记录错误日志
func (c *BaseController) LogError(format string, args ...any) {
	if c.logger != nil {
		c.logger.Printf("[ERROR] "+format, args...)
	}
}

// LogDebug 记录调试日志
func (c *BaseController) LogDebug(format string, args ...any) {
	if c.logger != nil {
		c.logger.Printf("[DEBUG] "+format, args...)
	}
}

// DumpRequest 输出请求信息（调试用）
func (c *BaseController) DumpRequest() map[string]any {
	if c.Ctx == nil {
		return map[string]any{"error": "context is nil"}
	}
	
	return map[string]any{
		"method":     string(c.Ctx.Method()),
		"path":       string(c.Ctx.Path()),
		"query":      string(c.Ctx.URI().QueryString()),
		"user_agent": string(c.Ctx.UserAgent()),
		"client_ip":  c.GetClientIP(),
		"headers":    c.getHeaders(),
	}
}

// getHeaders 获取所有请求头（私有方法）
func (c *BaseController) getHeaders() map[string]string {
	headers := make(map[string]string)
	if c.Ctx == nil {
		return headers
	}
	
	// 这里需要根据Hertz的实际API来实现
	// 暂时返回空map，实际项目中需要补充
	return headers
}
