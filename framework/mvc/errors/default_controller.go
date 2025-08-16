package errors

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// =============================================================================
// 模块：DefaultErrorController 核心实现
// 职责：默认错误控制器的核心功能，专注于错误处理流程和状态码分发
// =============================================================================

// DefaultErrorController 默认错误控制器
// 提供美观的错误页面展示、多格式响应支持、智能内容协商等功能
type DefaultErrorController struct {
	// 配置选项
	ShowDetailedError bool   // 是否显示详细错误信息（开发环境建议true）
	Language          string // 语言设置 ("zh-CN", "en-US")
	CustomTitle       string // 自定义页面标题
	SupportEmail      string // 支持邮箱
	SupportPhone      string // 支持电话
	EnableDebugInfo   bool   // 是否启用调试信息

	// 新增功能配置
	EnableErrorLogging bool                   // 是否启用错误日志记录
	ErrorLogger        *log.Logger            // 自定义错误日志记录器
	ErrorStatistics    *ErrorStatistics       // 错误统计信息
	CustomTemplates    map[int]string         // 自定义错误页面模板
	ErrorHooks         map[string][]ErrorHook // 错误处理钩子
	RetryableErrors    map[int]bool           // 可重试的错误类型

	// 线程安全保护
	mu sync.RWMutex
}

// NewDefaultErrorController 创建默认错误控制器实例
func NewDefaultErrorController() *DefaultErrorController {
	controller := &DefaultErrorController{
		ShowDetailedError: true,            // 默认显示详细错误信息
		Language:          DefaultLanguage, // 默认中文
		CustomTitle:       DefaultTitle,
		SupportEmail:      "",
		SupportPhone:      "",
		EnableDebugInfo:   true, // 默认启用调试信息

		// 新增功能初始化
		EnableErrorLogging: true,
		ErrorLogger:        log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
		ErrorStatistics:    NewErrorStatistics(),
		CustomTemplates:    make(map[int]string),
		ErrorHooks:         make(map[string][]ErrorHook),
		RetryableErrors:    InitRetryableErrors(),
	}

	// 初始化默认钩子
	controller.initDefaultHooks()

	return controller
}

// initDefaultHooks 初始化默认错误钩子
func (c *DefaultErrorController) initDefaultHooks() {
	// 添加日志记录钩子
	c.ErrorHooks[HookPhaseBefore] = []ErrorHook{
		c.logErrorHook,
		c.statisticsHook,
	}

	// 添加清理钩子
	c.ErrorHooks[HookPhaseAfter] = []ErrorHook{
		c.cleanupHook,
	}
}

// =============================================================================
// ErrorHandler 接口实现
// =============================================================================

// Handle 处理错误（实现ErrorHandler接口）
func (c *DefaultErrorController) Handle(ctx *mvccontext.Context, statusCode int, err error) error {
	// 执行前置钩子
	if hooks, exists := c.ErrorHooks[HookPhaseBefore]; exists {
		if hookErr := ExecuteHooks(hooks, ctx, statusCode, err, c.ErrorLogger); hookErr != nil {
			// 钩子执行失败时记录日志但继续处理
			logError("Before hook failed", hookErr, c.ErrorLogger)
		}
	}

	var handleErr error

	// 根据状态码选择处理方法
	switch statusCode {
	// 4xx 客户端错误
	case 400:
		handleErr = c.handle400(ctx, err)
	case 401:
		handleErr = c.handle401(ctx, err)
	case 402:
		handleErr = c.handle402(ctx, err)
	case 403:
		handleErr = c.handle403(ctx, err)
	case 404:
		handleErr = c.handle404(ctx, err)
	case 405:
		handleErr = c.handle405(ctx, err)
	case 406:
		handleErr = c.handle406(ctx, err)
	case 408:
		handleErr = c.handle408(ctx, err)
	case 409:
		handleErr = c.handle409(ctx, err)
	case 410:
		handleErr = c.handle410(ctx, err)
	case 413:
		handleErr = c.handle413(ctx, err)
	case 415:
		handleErr = c.handle415(ctx, err)
	case 418:
		handleErr = c.handle418(ctx, err)
	case 422:
		handleErr = c.handle422(ctx, err)
	case 429:
		handleErr = c.handle429(ctx, err)
	// 5xx 服务器错误
	case 500:
		handleErr = c.handle500(ctx, err)
	case 501:
		handleErr = c.handle501(ctx, err)
	case 502:
		handleErr = c.handle502(ctx, err)
	case 503:
		handleErr = c.handle503(ctx, err)
	case 504:
		handleErr = c.handle504(ctx, err)
	case 505:
		handleErr = c.handle505(ctx, err)
	default:
		handleErr = c.handleGeneric(ctx, statusCode, err)
	}

	// 执行后置钩子
	if hooks, exists := c.ErrorHooks[HookPhaseAfter]; exists {
		if hookErr := ExecuteHooks(hooks, ctx, statusCode, err, c.ErrorLogger); hookErr != nil {
			logError("After hook execution failed", hookErr, c.ErrorLogger)
		}
	}

	return handleErr
}

// CanHandle 检查是否能处理该错误
func (c *DefaultErrorController) CanHandle(statusCode int, err error) bool {
	// 默认控制器可以处理所有错误
	return true
}

// Priority 返回处理器优先级
func (c *DefaultErrorController) Priority() int {
	return DefaultPriority // 默认控制器优先级最低
}

// =============================================================================
// 具体状态码处理器
// =============================================================================

// handle401 处理401未授权访问错误
func (c *DefaultErrorController) handle401(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 401, "未授权访问", "您需要登录才能访问此资源", []string{
		"请先登录您的账户",
		"检查您的登录凭证是否过期",
		"联系管理员确认您的访问权限",
	}, err)
}

// handle403 处理403禁止访问错误
func (c *DefaultErrorController) handle403(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 403, "禁止访问", "您没有权限访问此资源", []string{
		"联系管理员申请相关权限",
		"确认您的账户状态是否正常",
		"检查是否访问了受限制的资源",
	}, err)
}

// handle404 处理404页面未找到错误
func (c *DefaultErrorController) handle404(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 404, "页面未找到", "请检查URL地址是否正确", []string{
		"检查URL拼写是否正确",
		"尝试返回首页重新导航",
		"清除浏览器缓存后重试",
	}, err)
}

// handle500 处理500服务器内部错误
func (c *DefaultErrorController) handle500(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 500, "服务器内部错误", "服务器遇到了意外情况", []string{
		"请稍后重试",
		"如果问题持续存在，请联系技术支持",
		"您也可以尝试刷新页面",
	}, err)
}

// handle400 处理400错误请求错误
func (c *DefaultErrorController) handle400(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 400, "错误请求", "请求参数错误或格式不正确", []string{
		"检查请求参数是否正确",
		"确认数据格式符合API要求",
		"查看API文档了解正确的请求格式",
		"检查Content-Type头是否正确设置",
	}, err)
}

// handle402 处理402需要付费错误
func (c *DefaultErrorController) handle402(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 402, "需要付费", "需要付费才能访问此资源", []string{
		"请联系管理员升级您的账户",
		"查看付费计划了解更多详情",
		"确认您的订阅状态",
		"联系销售团队获取报价信息",
	}, err)
}

// handle405 处理405方法不允许错误
func (c *DefaultErrorController) handle405(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 405, "方法不允许", "请求方法不被允许", []string{
		"检查请求方法（GET、POST、PUT、DELETE）是否正确",
		"查看API文档了解支持的HTTP方法",
		"尝试使用不同的请求方法",
		"确认端点是否支持当前的HTTP动词",
	}, err)
}

// handle406 处理406不可接受错误
func (c *DefaultErrorController) handle406(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 406, "不可接受", "服务器无法生成客户端可接受的内容", []string{
		"检查Accept头是否设置正确",
		"确认服务器支持您请求的内容类型",
		"尝试修改Accept头为支持的格式",
		"查看API文档了解支持的响应格式",
	}, err)
}

// handle408 处理408请求超时错误
func (c *DefaultErrorController) handle408(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 408, "请求超时", "请求超时，请重试", []string{
		"检查网络连接是否稳定",
		"尝试减少请求数据量",
		"稍后重试请求",
		"考虑分批处理大量数据",
		"联系网络管理员检查网络状况",
	}, err)
}

// handle409 处理409冲突错误
func (c *DefaultErrorController) handle409(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 409, "请求冲突", "请求与服务器当前状态冲突", []string{
		"检查是否有资源冲突",
		"确认操作是否已经执行过",
		"刷新页面获取最新状态后重试",
		"检查并发操作是否导致冲突",
		"使用乐观锁或版本控制机制",
	}, err)
}

// handle410 处理410资源已删除错误
func (c *DefaultErrorController) handle410(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 410, "资源已删除", "请求的资源已不再可用", []string{
		"确认请求的资源是否已被永久删除",
		"检查是否使用了过期的链接或书签",
		"联系管理员确认资源状态",
		"查找资源的新位置或替代方案",
		"更新您的书签或缓存的链接",
	}, err)
}

// handle413 处理413请求实体过大错误
func (c *DefaultErrorController) handle413(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 413, "请求实体过大", "请求数据过大", []string{
		"减少上传文件的大小",
		"分批上传大型数据",
		"联系管理员提高上传限制",
		"压缩文件后再上传",
		"检查服务器的最大请求大小配置",
	}, err)
}

// handle415 处理415不支持的媒体类型错误
func (c *DefaultErrorController) handle415(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 415, "不支持的媒体类型", "不支持的媒体类型", []string{
		"检查文件格式是否被支持",
		"尝试使用标准的媒体类型",
		"查看支持的文件格式列表",
		"确认Content-Type头设置正确",
		"转换文件为支持的格式",
	}, err)
}

// handle418 处理418我是茶壶错误（RFC 2324彩蛋）
func (c *DefaultErrorController) handle418(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 418, "我是茶壶", "我是一个茶壶，不能泡咖啡", []string{
		"这是一个彩蛋错误，恭喜你发现了！",
		"尝试使用咖啡机而不是茶壶",
		"RFC 2324 - 超文本咖啡壶控制协议",
		"这个状态码通常用于测试或幽默目的",
		"检查是否误触发了测试端点",
	}, err)
}

// handle422 处理422无法处理的实体错误
func (c *DefaultErrorController) handle422(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 422, "无法处理的实体", "请求格式正确但语义错误", []string{
		"检查提交的数据是否符合业务规则",
		"确认必填字段都已填写",
		"验证数据格式和内容是否正确",
		"检查字段值是否在允许的范围内",
		"确认关联数据的完整性和一致性",
	}, err)
}

// handle429 处理429请求过多错误
func (c *DefaultErrorController) handle429(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 429, "请求过多", "请求过于频繁，请稍后重试", []string{
		"请降低请求频率",
		"等待一段时间后重试",
		"考虑使用缓存减少重复请求",
		"检查是否有重复或批量请求",
		"实现请求去重和防抖机制",
	}, err)
}

// handle501 处理501未实现错误
func (c *DefaultErrorController) handle501(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 501, "未实现", "服务器不支持此功能", []string{
		"尝试使用其他可用的功能",
		"联系技术支持了解功能开发计划",
		"查看API文档了解已实现的功能",
		"检查是否使用了正确的API版本",
		"寻找替代的实现方案",
	}, err)
}

// handle502 处理502网关错误
func (c *DefaultErrorController) handle502(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 502, "网关错误", "网关错误", []string{
		"请稍后重试",
		"检查网络连接是否正常",
		"如果问题持续存在，请联系技术支持",
		"可能是上游服务器的问题",
		"检查代理或负载均衡器配置",
	}, err)
}

// handle503 处理503服务不可用错误
func (c *DefaultErrorController) handle503(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 503, "服务不可用", "服务暂时不可用", []string{
		"服务正在维护中，请稍后重试",
		"关注官方公告了解维护时间",
		"使用其他可用的服务入口",
		"检查服务器是否过载",
		"联系技术支持获取维护信息",
	}, err)
}

// handle504 处理504网关超时错误
func (c *DefaultErrorController) handle504(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 504, "网关超时", "网关超时", []string{
		"请稍后重试",
		"检查网络连接是否稳定",
		"如果问题持续存在，请联系技术支持",
		"可能是后端服务响应太慢",
		"考虑优化请求或分解复杂操作",
	}, err)
}

// handle505 处理505HTTP版本不支持错误
func (c *DefaultErrorController) handle505(ctx *mvccontext.Context, err error) error {
	return c.handleError(ctx, 505, "HTTP版本不支持", "不支持的HTTP版本", []string{
		"升级您的客户端版本",
		"使用标准的HTTP协议版本",
		"联系技术支持获取兼容性信息",
		"检查客户端HTTP版本设置",
		"确认服务器支持的HTTP版本",
	}, err)
}

// handleGeneric 处理通用错误
func (c *DefaultErrorController) handleGeneric(ctx *mvccontext.Context, statusCode int, err error) error {
	statusText := http.StatusText(statusCode)
	message := getStatusMessage(statusCode)
	if err != nil {
		message = err.Error()
	}

	// 智能生成建议
	suggestions := GenerateSmartSuggestions(statusCode, ctx)

	return c.handleError(ctx, statusCode, statusText, message, suggestions, err)
}

// =============================================================================
// 核心处理逻辑
// =============================================================================

// handleError 核心错误处理逻辑
func (c *DefaultErrorController) handleError(ctx *mvccontext.Context, statusCode int, title, message string, suggestions []string, err error) error {
	// 构建错误上下文
	errorCtx := c.buildErrorContext(ctx, statusCode, title, message, suggestions, err)

	// 使用内容协商处理错误响应
	return HandleErrorResponse(ctx, errorCtx, c.ShowDetailedError, c.CustomTitle, c.SupportEmail, c.SupportPhone, c.EnableDebugInfo)
}

// buildErrorContext 构建错误上下文信息
func (c *DefaultErrorController) buildErrorContext(ctx *mvccontext.Context, statusCode int, title, message string, suggestions []string, err error) *ErrorContext {
	errorCtx := &ErrorContext{
		StatusCode:    statusCode,
		StatusText:    title,
		ErrorMessage:  message,
		RequestPath:   string(ctx.Path()),
		RequestMethod: string(ctx.Method()),
		UserAgent:     string(ctx.UserAgent()),
		Timestamp:     time.Now(),
		Details:       make(map[string]any),
		Suggestions:   suggestions,
	}

	// 添加请求ID（如果存在）
	if requestIDValue, exists := ctx.Get("request_id"); exists {
		if requestID, ok := requestIDValue.(string); ok {
			errorCtx.RequestID = requestID
		}
	}

	// 开发环境添加调试信息
	if c.EnableDebugInfo {
		errorCtx.Details["goroutines"] = runtime.NumGoroutine()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		errorCtx.Details["memory_usage_mb"] = float64(m.Alloc) / 1024 / 1024
		errorCtx.Details["gc_count"] = m.NumGC

		// 添加原始错误信息
		if err != nil && c.ShowDetailedError {
			errorCtx.Details["original_error"] = err.Error()
		}
	}

	return errorCtx
}

// getStatusMessage 获取状态码对应的中文描述
func getStatusMessage(statusCode int) string {
	switch statusCode {
	case 400:
		return "请求参数错误或格式不正确"
	case 401:
		return "需要身份验证才能访问此资源"
	case 402:
		return "需要付费才能访问此资源"
	case 403:
		return "您没有权限访问此资源"
	case 404:
		return "请检查URL地址是否正确"
	case 405:
		return "请求方法不被允许"
	case 406:
		return "服务器无法生成客户端可接受的内容"
	case 408:
		return "请求超时，请重试"
	case 409:
		return "请求与服务器当前状态冲突"
	case 410:
		return "请求的资源已不再可用"
	case 413:
		return "请求数据过大"
	case 415:
		return "不支持的媒体类型"
	case 418:
		return "我是一个茶壶，不能泡咖啡"
	case 422:
		return "请求格式正确但语义错误"
	case 429:
		return "请求过于频繁，请稍后重试"
	case 500:
		return "服务器遇到了意外情况"
	case 501:
		return "服务器不支持此功能"
	case 502:
		return "网关错误"
	case 503:
		return "服务暂时不可用"
	case 504:
		return "网关超时"
	case 505:
		return "不支持的HTTP版本"
	default:
		if statusCode >= 400 && statusCode < 500 {
			return "客户端请求错误"
		} else if statusCode >= 500 {
			return "服务器内部错误"
		}
		return "发生了一个错误"
	}
}

// =============================================================================
// 钩子方法（兼容性保留）
// =============================================================================

// logErrorHook 错误日志记录钩子
func (c *DefaultErrorController) logErrorHook(ctx *mvccontext.Context, statusCode int, err error) error {
	return LogErrorHook(ctx, statusCode, err, c.EnableErrorLogging, c.ErrorLogger)
}

// statisticsHook 错误统计钩子
func (c *DefaultErrorController) statisticsHook(ctx *mvccontext.Context, statusCode int, err error) error {
	return StatisticsHook(ctx, statusCode, err, c.ErrorStatistics)
}

// cleanupHook 清理钩子
func (c *DefaultErrorController) cleanupHook(ctx *mvccontext.Context, statusCode int, err error) error {
	return CleanupHook(ctx, statusCode, err)
}

// =============================================================================
// 公共API方法
// =============================================================================

// AddHook 添加错误处理钩子
func (c *DefaultErrorController) AddHook(phase string, hook ErrorHook) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ErrorHooks[phase] == nil {
		c.ErrorHooks[phase] = make([]ErrorHook, 0)
	}
	c.ErrorHooks[phase] = append(c.ErrorHooks[phase], hook)
}

// RemoveHooks 移除指定阶段的所有钩子
func (c *DefaultErrorController) RemoveHooks(phase string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.ErrorHooks, phase)
}

// SetCustomTemplate 设置自定义错误页面模板
func (c *DefaultErrorController) SetCustomTemplate(statusCode int, template string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.CustomTemplates[statusCode] = template
}

// GetErrorStatistics 获取错误统计信息
func (c *DefaultErrorController) GetErrorStatistics() *ErrorStatistics {
	return GetErrorStatistics(c.ErrorStatistics)
}

// ResetStatistics 重置错误统计信息
func (c *DefaultErrorController) ResetStatistics() {
	ResetStatistics(c.ErrorStatistics)
}

// IsRetryable 检查错误是否可重试
func (c *DefaultErrorController) IsRetryable(statusCode int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return IsRetryable(statusCode, c.RetryableErrors)
}

// SetRetryable 设置错误是否可重试
func (c *DefaultErrorController) SetRetryable(statusCode int, retryable bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	SetRetryable(statusCode, retryable, c.RetryableErrors)
}

// =============================================================================
// 多语言支持
// =============================================================================

// getLocalizedMessage 获取本地化消息
func (c *DefaultErrorController) getLocalizedMessage(key string, messages map[string]I18nMessage) string {
	if msg, exists := messages[key]; exists {
		switch c.Language {
		case LanguageEnUS, LanguageEn:
			if msg.EnUS != "" {
				return msg.EnUS
			}
		case LanguageZhCN, LanguageZh:
			if msg.ZhCN != "" {
				return msg.ZhCN
			}
		}
	}
	return key // 如果没有找到对应的翻译，返回key本身
}

// =============================================================================
// 业务错误处理兼容性方法
// =============================================================================

// HandleBusinessError 处理业务错误
func (c *DefaultErrorController) HandleBusinessError(ctx *mvccontext.Context, bizErr *BusinessError) error {
	return HandleBusinessError(ctx, bizErr, c.ShowDetailedError)
}