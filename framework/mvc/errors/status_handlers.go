package errors

import (
	"fmt"
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 通用状态处理器 =============

// UniversalStatusHandler 通用状态处理器（替换20+个重复方法）
type UniversalStatusHandler struct {
	configManager ConfigManager
	renderers     []ErrorRenderer
}

// NewUniversalStatusHandler 创建通用状态处理器
func NewUniversalStatusHandler(configManager ConfigManager) *UniversalStatusHandler {
	return &UniversalStatusHandler{
		configManager: configManager,
		renderers:     []ErrorRenderer{},
	}
}

// HandleStatus 处理任意状态码（核心方法）
func (h *UniversalStatusHandler) HandleStatus(ctx *mvccontext.Context, statusCode int, err error) error {
	// 获取状态码配置
	config := h.configManager.GetStatusConfig(statusCode)
	
	// 构建错误上下文
	errorCtx := h.buildErrorContext(ctx, statusCode, config, err)
	
	// 选择合适的渲染器
	for _, renderer := range h.renderers {
		if renderer.CanRender(ctx) {
			return renderer.Render(ctx, errorCtx)
		}
	}
	
	// 默认渲染器（JSON）
	return h.renderDefaultJSON(ctx, errorCtx)
}

// GetStatusCodes 返回支持的状态码列表
func (h *UniversalStatusHandler) GetStatusCodes() []int {
	// 支持所有HTTP状态码
	return []int{} // 空表示支持所有
}

// AddRenderer 添加渲染器
func (h *UniversalStatusHandler) AddRenderer(renderer ErrorRenderer) {
	h.renderers = append(h.renderers, renderer)
}

// buildErrorContext 构建错误上下文
func (h *UniversalStatusHandler) buildErrorContext(ctx *mvccontext.Context, statusCode int, config *StatusConfig, err error) *ErrorContext {
	// 复用现有的ErrorContext结构
	errorCtx := &ErrorContext{
		StatusCode:    statusCode,
		StatusText:    config.Title,
		ErrorMessage:  config.Message,
		RequestPath:   string(ctx.Path()),
		RequestMethod: string(ctx.Method()),
		UserAgent:     string(ctx.UserAgent()),
		Details:       make(map[string]any, 4),
		Suggestions:   config.Suggestions,
	}
	
	// 添加请求ID（如果存在）
	if requestIDValue, exists := ctx.Get("request_id"); exists {
		if requestID, ok := requestIDValue.(string); ok {
			errorCtx.RequestID = requestID
		}
	}
	
	// 添加原始错误信息
	if err != nil {
		errorCtx.Details["original_error"] = err.Error()
	}
	
	// 添加配置中的额外信息
	if len(config.Recovery) > 0 {
		errorCtx.Details["recovery"] = config.Recovery
	}
	if len(config.Prevention) > 0 {
		errorCtx.Details["prevention"] = config.Prevention
	}
	
	return errorCtx
}

// renderDefaultJSON 默认JSON渲染
func (h *UniversalStatusHandler) renderDefaultJSON(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	response := map[string]any{
		"code":      errorCtx.StatusCode,
		"message":   errorCtx.ErrorMessage,
		"success":   false,
		"path":      errorCtx.RequestPath,
		"method":    errorCtx.RequestMethod,
		"timestamp": errorCtx.Timestamp.Unix(),
	}
	
	if len(errorCtx.Suggestions) > 0 {
		response["suggestions"] = errorCtx.Suggestions
	}
	
	if errorCtx.RequestID != "" {
		response["request_id"] = errorCtx.RequestID
	}
	
	ctx.JSON(errorCtx.StatusCode, response)
	return nil
}

// ============= 专门的状态处理器 =============

// CommonStatusHandler 常见状态码处理器
type CommonStatusHandler struct {
	handler *UniversalStatusHandler
	codes   []int
}

// NewCommonStatusHandler 创建常见状态码处理器
func NewCommonStatusHandler(configManager ConfigManager, codes []int) *CommonStatusHandler {
	return &CommonStatusHandler{
		handler: NewUniversalStatusHandler(configManager),
		codes:   codes,
	}
}

// HandleStatus 处理状态码
func (h *CommonStatusHandler) HandleStatus(ctx *mvccontext.Context, statusCode int, err error) error {
	return h.handler.HandleStatus(ctx, statusCode, err)
}

// GetStatusCodes 返回支持的状态码列表
func (h *CommonStatusHandler) GetStatusCodes() []int {
	return h.codes
}

// CanHandle 检查是否能处理该状态码
func (h *CommonStatusHandler) CanHandle(statusCode int) bool {
	for _, code := range h.codes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// ============= 预定义的处理器 =============

// Create4xxHandler 创建4xx错误处理器
func Create4xxHandler(configManager ConfigManager) *CommonStatusHandler {
	codes := []int{400, 401, 402, 403, 404, 405, 406, 408, 409, 410, 413, 415, 418, 422, 429}
	return NewCommonStatusHandler(configManager, codes)
}

// Create5xxHandler 创建5xx错误处理器
func Create5xxHandler(configManager ConfigManager) *CommonStatusHandler {
	codes := []int{500, 501, 502, 503, 504, 505}
	return NewCommonStatusHandler(configManager, codes)
}

// ============= 业务错误处理器 =============

// BusinessErrorStatusHandler 业务错误状态处理器
type BusinessErrorStatusHandler struct {
	configManager ConfigManager
}

// NewBusinessErrorStatusHandler 创建业务错误状态处理器
func NewBusinessErrorStatusHandler(configManager ConfigManager) *BusinessErrorStatusHandler {
	return &BusinessErrorStatusHandler{
		configManager: configManager,
	}
}

// HandleStatus 处理业务错误
func (h *BusinessErrorStatusHandler) HandleStatus(ctx *mvccontext.Context, statusCode int, err error) error {
	// 检查是否是业务错误
	if bizErr, ok := err.(*BusinessError); ok {
		response := map[string]any{
			"code":    bizErr.Code,
			"message": bizErr.Message,
			"success": false,
		}
		
		if bizErr.Data != nil {
			response["data"] = bizErr.Data
		}
		
		// 现有的BusinessError结构只有Code、Message、Data字段
		// 如果需要更多字段，可以在Data中传递
		
		ctx.JSON(statusCode, response)
		return nil
	}
	
	// 不是业务错误，使用通用处理
	handler := NewUniversalStatusHandler(h.configManager)
	return handler.HandleStatus(ctx, statusCode, err)
}

// GetStatusCodes 返回支持的状态码列表
func (h *BusinessErrorStatusHandler) GetStatusCodes() []int {
	return []int{} // 支持所有状态码
}

// ============= 智能内容协商处理器 =============

// ContentNegotiationHandler 内容协商处理器
type ContentNegotiationHandler struct {
	configManager ConfigManager
	htmlRenderer  ErrorRenderer
	jsonRenderer  ErrorRenderer
	xmlRenderer   ErrorRenderer
}

// NewContentNegotiationHandler 创建内容协商处理器
func NewContentNegotiationHandler(configManager ConfigManager) *ContentNegotiationHandler {
	return &ContentNegotiationHandler{
		configManager: configManager,
	}
}

// SetRenderers 设置渲染器
func (h *ContentNegotiationHandler) SetRenderers(html, json, xml ErrorRenderer) {
	h.htmlRenderer = html
	h.jsonRenderer = json
	h.xmlRenderer = xml
}

// HandleStatus 根据内容协商处理状态码
func (h *ContentNegotiationHandler) HandleStatus(ctx *mvccontext.Context, statusCode int, err error) error {
	config := h.configManager.GetStatusConfig(statusCode)
	errorCtx := h.buildErrorContext(ctx, statusCode, config, err)
	
	// 内容协商
	accept := string(ctx.GetHeader("Accept"))
	path := string(ctx.Path())
	
	// API请求优先返回JSON
	if h.isAPIRequest(path, accept) && h.jsonRenderer != nil {
		return h.jsonRenderer.Render(ctx, errorCtx)
	}
	
	// XML请求
	if h.isXMLRequest(accept) && h.xmlRenderer != nil {
		return h.xmlRenderer.Render(ctx, errorCtx)
	}
	
	// 默认HTML
	if h.htmlRenderer != nil {
		return h.htmlRenderer.Render(ctx, errorCtx)
	}
	
	// 最后回退到JSON
	return h.renderFallbackJSON(ctx, errorCtx)
}

// GetStatusCodes 返回支持的状态码列表
func (h *ContentNegotiationHandler) GetStatusCodes() []int {
	return []int{} // 支持所有状态码
}

// buildErrorContext 构建错误上下文
func (h *ContentNegotiationHandler) buildErrorContext(ctx *mvccontext.Context, statusCode int, config *StatusConfig, err error) *ErrorContext {
	errorCtx := &ErrorContext{
		StatusCode:    statusCode,
		StatusText:    config.Title,
		ErrorMessage:  config.Message,
		RequestPath:   string(ctx.Path()),
		RequestMethod: string(ctx.Method()),
		UserAgent:     string(ctx.UserAgent()),
		Details:       make(map[string]any, 4),
		Suggestions:   config.Suggestions,
	}
	
	// 添加请求ID
	if requestIDValue, exists := ctx.Get("request_id"); exists {
		if requestID, ok := requestIDValue.(string); ok {
			errorCtx.RequestID = requestID
		}
	}
	
	return errorCtx
}

// isAPIRequest 判断是否为API请求
func (h *ContentNegotiationHandler) isAPIRequest(path, accept string) bool {
	// 路径以 /api/ 开头
	if len(path) >= 4 && path[:4] == "/api" {
		return true
	}
	
	// Accept头包含JSON但不包含HTML
	return contains(accept, "application/json") && !contains(accept, "text/html")
}

// isXMLRequest 判断是否为XML请求
func (h *ContentNegotiationHandler) isXMLRequest(accept string) bool {
	return (contains(accept, "application/xml") || contains(accept, "text/xml")) && 
		   !contains(accept, "text/html")
}

// renderFallbackJSON 回退JSON渲染
func (h *ContentNegotiationHandler) renderFallbackJSON(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	response := map[string]any{
		"code":      errorCtx.StatusCode,
		"message":   errorCtx.ErrorMessage,
		"success":   false,
		"timestamp": errorCtx.Timestamp.Unix(),
	}
	
	ctx.JSON(errorCtx.StatusCode, response)
	return nil
}

// contains 检查字符串是否包含子字符串（简单实现）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   func() bool {
			   for i := 0; i <= len(s)-len(substr); i++ {
				   if s[i:i+len(substr)] == substr {
					   return true
				   }
			   }
			   return false
		   }()
}

// ============= 处理器工厂方法 =============

// CreateDefaultStatusHandlers 创建默认状态处理器集合
func CreateDefaultStatusHandlers(configManager ConfigManager) map[string]StatusHandler {
	handlers := make(map[string]StatusHandler)
	
	// 创建通用处理器
	handlers["universal"] = NewUniversalStatusHandler(configManager)
	
	// 创建分类处理器
	handlers["4xx"] = Create4xxHandler(configManager)
	handlers["5xx"] = Create5xxHandler(configManager)
	
	// 创建业务错误处理器
	handlers["business"] = NewBusinessErrorStatusHandler(configManager)
	
	// 创建内容协商处理器
	handlers["content_negotiation"] = NewContentNegotiationHandler(configManager)
	
	return handlers
}

// GetRecommendedHandler 获取推荐的处理器
func GetRecommendedHandler(configManager ConfigManager) StatusHandler {
	return NewContentNegotiationHandler(configManager)
}

// ============= 性能优化版本处理器 =============

// FastStatusHandler 高性能状态处理器（预编译配置）
type FastStatusHandler struct {
	statusConfigs map[int]*StatusConfig
	messages      map[int]string
	suggestions   map[int][]string
}

// NewFastStatusHandler 创建高性能状态处理器
func NewFastStatusHandler() *FastStatusHandler {
	handler := &FastStatusHandler{
		statusConfigs: make(map[int]*StatusConfig, 32),
		messages:      make(map[int]string, 32),
		suggestions:   make(map[int][]string, 32),
	}
	
	// 预编译常用状态码配置
	handler.precompileConfigs()
	
	return handler
}

// HandleStatus 高性能状态处理
func (h *FastStatusHandler) HandleStatus(ctx *mvccontext.Context, statusCode int, err error) error {
	// 直接从预编译的配置中获取
	message, exists := h.messages[statusCode]
	if !exists {
		message = fmt.Sprintf("HTTP %d Error", statusCode)
	}
	
	suggestions := h.suggestions[statusCode]
	
	// 构建最小化的响应
	response := map[string]any{
		"code":    statusCode,
		"message": message,
		"success": false,
	}
	
	if len(suggestions) > 0 {
		response["suggestions"] = suggestions
	}
	
	ctx.JSON(statusCode, response)
	return nil
}

// GetStatusCodes 返回支持的状态码列表
func (h *FastStatusHandler) GetStatusCodes() []int {
	codes := make([]int, 0, len(h.statusConfigs))
	for code := range h.statusConfigs {
		codes = append(codes, code)
	}
	return codes
}

// precompileConfigs 预编译配置
func (h *FastStatusHandler) precompileConfigs() {
	// 预编译常用状态码
	configs := map[int]struct{
		message     string
		suggestions []string
	}{
		400: {"请求参数错误", []string{"检查请求参数", "确认数据格式"}},
		401: {"未授权访问", []string{"请先登录", "检查登录状态"}},
		403: {"权限不足", []string{"联系管理员", "确认权限"}},
		404: {"页面未找到", []string{"检查URL", "返回首页"}},
		429: {"请求过多", []string{"降低频率", "稍后重试"}},
		500: {"服务器错误", []string{"稍后重试", "联系技术支持"}},
		502: {"网关错误", []string{"稍后重试", "检查网络"}},
		503: {"服务不可用", []string{"服务维护中", "稍后重试"}},
	}
	
	for code, config := range configs {
		h.messages[code] = config.message
		h.suggestions[code] = config.suggestions
	}
}