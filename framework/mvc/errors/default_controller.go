package errors

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 默认错误控制器 =============

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

// ErrorContext 错误上下文信息
type ErrorContext struct {
	StatusCode    int            `json:"status_code"`
	StatusText    string         `json:"status_text"`
	ErrorMessage  string         `json:"error_message"`
	RequestPath   string         `json:"request_path"`
	RequestMethod string         `json:"request_method"`
	UserAgent     string         `json:"user_agent"`
	Timestamp     time.Time      `json:"timestamp"`
	RequestID     string         `json:"request_id,omitempty"`
	Details       map[string]any `json:"details,omitempty"`
	Suggestions   []string       `json:"suggestions,omitempty"`
}

// ErrorStatistics 错误统计信息
type ErrorStatistics struct {
	TotalErrors    int64            `json:"total_errors"`
	ErrorsByStatus map[int]int64    `json:"errors_by_status"`
	ErrorsByPath   map[string]int64 `json:"errors_by_path"`
	LastErrors     []ErrorRecord    `json:"last_errors"`
	StartTime      time.Time        `json:"start_time"`
	mu             sync.RWMutex
}

// ErrorRecord 错误记录
type ErrorRecord struct {
	StatusCode int       `json:"status_code"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	UserAgent  string    `json:"user_agent,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
}

// ErrorHook 错误处理钩子函数类型
type ErrorHook func(ctx *mvccontext.Context, statusCode int, err error) error

// I18nMessage 国际化消息结构
type I18nMessage struct {
	ZhCN string `json:"zh_cn"`
	EnUS string `json:"en_us"`
}

// NewDefaultErrorController 创建默认错误控制器实例
func NewDefaultErrorController() *DefaultErrorController {
	controller := &DefaultErrorController{
		ShowDetailedError: true,    // 默认显示详细错误信息
		Language:          "zh-CN", // 默认中文
		CustomTitle:       "YYHertz Framework",
		SupportEmail:      "",
		SupportPhone:      "",
		EnableDebugInfo:   true, // 默认启用调试信息

		// 新增功能初始化
		EnableErrorLogging: true,
		ErrorLogger:        log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
		ErrorStatistics:    NewErrorStatistics(),
		CustomTemplates:    make(map[int]string),
		ErrorHooks:         make(map[string][]ErrorHook),
		RetryableErrors:    initRetryableErrors(),
	}

	// 初始化默认错误钩子
	controller.initDefaultHooks()

	return controller
}

// NewErrorStatistics 创建错误统计实例
func NewErrorStatistics() *ErrorStatistics {
	return &ErrorStatistics{
		TotalErrors:    0,
		ErrorsByStatus: make(map[int]int64),
		ErrorsByPath:   make(map[string]int64),
		LastErrors:     make([]ErrorRecord, 0),
		StartTime:      time.Now(),
	}
}

// initRetryableErrors 初始化可重试的错误类型
func initRetryableErrors() map[int]bool {
	return map[int]bool{
		408: true, // Request Timeout
		429: true, // Too Many Requests
		500: true, // Internal Server Error
		502: true, // Bad Gateway
		503: true, // Service Unavailable
		504: true, // Gateway Timeout
	}
}

// initDefaultHooks 初始化默认错误钩子
func (c *DefaultErrorController) initDefaultHooks() {
	// 添加日志记录钩子
	c.ErrorHooks["before"] = []ErrorHook{
		c.logErrorHook,
		c.statisticsHook,
	}

	// 添加清理钩子
	c.ErrorHooks["after"] = []ErrorHook{
		c.cleanupHook,
	}
}

// NewProductionErrorController 创建生产环境错误控制器
func NewProductionErrorController() *DefaultErrorController {
	controller := &DefaultErrorController{
		ShowDetailedError: false, // 生产环境不显示详细错误
		Language:          "zh-CN",
		CustomTitle:       "YYHertz Framework",
		SupportEmail:      "",
		SupportPhone:      "",
		EnableDebugInfo:   false, // 生产环境不显示调试信息

		// 生产环境配置
		EnableErrorLogging: true,
		ErrorLogger:        log.New(os.Stderr, "[PROD-ERROR] ", log.LstdFlags),
		ErrorStatistics:    NewErrorStatistics(),
		CustomTemplates:    make(map[int]string),
		ErrorHooks:         make(map[string][]ErrorHook),
		RetryableErrors:    initRetryableErrors(),
	}

	controller.initDefaultHooks()
	return controller
}

// NewDevelopmentErrorController 创建开发环境错误控制器
func NewDevelopmentErrorController() *DefaultErrorController {
	controller := &DefaultErrorController{
		ShowDetailedError: true, // 开发环境显示详细错误
		Language:          "zh-CN",
		CustomTitle:       "YYHertz Framework - 开发环境",
		SupportEmail:      "",
		SupportPhone:      "",
		EnableDebugInfo:   true, // 开发环境显示调试信息

		// 开发环境配置
		EnableErrorLogging: true,
		ErrorLogger:        log.New(os.Stdout, "[DEV-ERROR] ", log.LstdFlags|log.Lshortfile),
		ErrorStatistics:    NewErrorStatistics(),
		CustomTemplates:    make(map[int]string),
		ErrorHooks:         make(map[string][]ErrorHook),
		RetryableErrors:    initRetryableErrors(),
	}

	controller.initDefaultHooks()
	return controller
}

// ============= 实现ErrorHandler接口 =============

// Handle 处理错误（实现ErrorHandler接口）
func (c *DefaultErrorController) Handle(ctx *mvccontext.Context, statusCode int, err error) error {
	// 执行前置钩子
	if hooks, exists := c.ErrorHooks["before"]; exists {
		for _, hook := range hooks {
			if hookErr := hook(ctx, statusCode, err); hookErr != nil {
				// 钩子执行失败时记录日志但继续处理
				c.logError("Hook execution failed", hookErr)
			}
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
	if hooks, exists := c.ErrorHooks["after"]; exists {
		for _, hook := range hooks {
			if hookErr := hook(ctx, statusCode, err); hookErr != nil {
				c.logError("After hook execution failed", hookErr)
			}
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
	return 1000 // 默认控制器优先级最低
}

// ============= 核心错误处理方法 =============

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

// ============= 新增专用错误处理函数 =============

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
	message := c.getStatusMessage(statusCode)
	if err != nil {
		message = err.Error()
	}

	// 智能生成建议
	suggestions := c.generateSmartSuggestions(statusCode, ctx)

	return c.handleError(ctx, statusCode, statusText, message, suggestions, err)
}

// ============= 错误分类和增强功能 =============

// getErrorCategory 获取错误类别
func (c *DefaultErrorController) getErrorCategory(statusCode int) string {
	switch {
	case statusCode >= 400 && statusCode < 410:
		return "客户端请求错误"
	case statusCode >= 410 && statusCode < 420:
		return "资源状态错误"
	case statusCode >= 420 && statusCode < 430:
		return "客户端数据错误"
	case statusCode >= 430 && statusCode < 500:
		return "客户端限制错误"
	case statusCode >= 500 && statusCode < 510:
		return "服务器执行错误"
	case statusCode >= 510 && statusCode < 520:
		return "服务器配置错误"
	default:
		return "未知错误类别"
	}
}

// isRetryableError 判断错误是否可重试（增强版）
func (c *DefaultErrorController) isRetryableError(statusCode int) bool {
	// 基础可重试检查
	if c.IsRetryable(statusCode) {
		return true
	}

	// 额外的智能判断
	switch statusCode {
	case 408, 429: // 超时和频率限制总是可重试
		return true
	case 500, 502, 503, 504: // 服务器临时错误
		return true
	case 507: // 存储空间不足
		return true
	default:
		return false
	}
}

// generateContextualAdvice 生成上下文相关的建议
func (c *DefaultErrorController) generateContextualAdvice(statusCode int, ctx *mvccontext.Context) []string {
	advice := []string{}

	// 基于请求路径的建议
	path := string(ctx.Path())
	method := string(ctx.Method())

	if strings.HasPrefix(path, "/api/") {
		advice = append(advice, "API请求失败，请检查API版本和端点是否正确")
		if statusCode == 401 {
			advice = append(advice, "API请求需要有效的认证令牌")
		}
	}

	if strings.HasPrefix(path, "/admin/") {
		advice = append(advice, "管理功能需要管理员权限")
	}

	// 基于请求方法的建议
	switch method {
	case "POST":
		if statusCode == 413 {
			advice = append(advice, "POST请求数据过大，考虑分块上传")
		}
	case "GET":
		if statusCode == 405 {
			advice = append(advice, "该端点可能只支持POST或PUT方法")
		}
	}

	// 基于用户代理的建议
	userAgent := string(ctx.UserAgent())
	if strings.Contains(userAgent, "Mobile") && statusCode == 406 {
		advice = append(advice, "移动端请求可能需要特定的Accept头")
	}

	return advice
}

// getRecoveryInstructions 获取恢复指令
func (c *DefaultErrorController) getRecoveryInstructions(statusCode int) []string {
	switch statusCode {
	case 400:
		return []string{
			"1. 检查请求参数格式和类型",
			"2. 验证必需参数是否缺失",
			"3. 确认请求体格式正确（JSON/XML等）",
		}
	case 401:
		return []string{
			"1. 获取有效的访问令牌或登录凭证",
			"2. 检查令牌是否已过期",
			"3. 确认权限范围是否足够",
		}
	case 403:
		return []string{
			"1. 联系管理员申请必要权限",
			"2. 检查用户角色和权限设置",
			"3. 确认访问的资源确实需要当前权限",
		}
	case 404:
		return []string{
			"1. 验证URL路径拼写是否正确",
			"2. 检查资源是否已被移动或删除",
			"3. 确认API版本是否正确",
		}
	case 429:
		return []string{
			"1. 实施指数退避重试策略",
			"2. 减少并发请求数量",
			"3. 考虑请求缓存机制",
		}
	case 500:
		return []string{
			"1. 稍后重试请求",
			"2. 检查请求是否触发了服务器bug",
			"3. 联系技术支持并提供错误上下文",
		}
	default:
		return []string{
			"1. 查看相关文档了解错误原因",
			"2. 检查网络连接状态",
			"3. 如问题持续请联系技术支持",
		}
	}
}

// getPrevention获取预防措施建议
func (c *DefaultErrorController) getPreventionTips(statusCode int) []string {
	switch statusCode {
	case 400:
		return []string{
			"💡 使用API文档验证请求格式",
			"💡 实施客户端数据验证",
			"💡 使用类型安全的API客户端",
		}
	case 401:
		return []string{
			"💡 实施令牌自动刷新机制",
			"💡 监控令牌过期时间",
			"💡 使用安全的认证流程",
		}
	case 403:
		return []string{
			"💡 实施基于角色的访问控制",
			"💡 定期审查用户权限",
			"💡 使用最小权限原则",
		}
	case 429:
		return []string{
			"💡 实施客户端限流机制",
			"💡 使用请求队列管理",
			"💡 监控API使用模式",
		}
	case 500:
		return []string{
			"💡 增加服务器监控和告警",
			"💡 实施断路器模式",
			"💡 建立容错和降级机制",
		}
	default:
		return []string{
			"💡 监控应用性能和错误率",
			"💡 建立完善的日志记录",
			"💡 定期进行系统健康检查",
		}
	}
}

// ============= 核心处理逻辑 =============

// handleError 核心错误处理逻辑
func (c *DefaultErrorController) handleError(ctx *mvccontext.Context, statusCode int, title, message string, suggestions []string, err error) error {
	// 构建错误上下文
	errorCtx := c.buildErrorContext(ctx, statusCode, title, message, suggestions, err)

	// 判断请求类型并选择合适的响应格式
	if c.isAPIRequest(ctx) {
		return c.renderJSONError(ctx, errorCtx)
	}

	// 检查是否请求XML格式
	if c.isXMLRequest(ctx) {
		return c.renderXMLError(ctx, errorCtx)
	}

	// 默认渲染HTML页面
	return c.renderHTMLError(ctx, errorCtx)
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

// isAPIRequest 判断是否为API请求
func (c *DefaultErrorController) isAPIRequest(ctx *mvccontext.Context) bool {
	path := string(ctx.Path())
	accept := string(ctx.GetHeader("Accept"))

	// 路径以 /api/ 开头的明确是API请求
	if strings.HasPrefix(path, "/api/") {
		return true
	}

	// Accept头明确表示只要JSON，不要HTML
	if (strings.Contains(accept, "application/json") ||
		strings.Contains(accept, "application/vnd.api+json")) &&
		!strings.Contains(accept, "text/html") {
		return true
	}

	return false
}

// isXMLRequest 判断是否请求XML格式
func (c *DefaultErrorController) isXMLRequest(ctx *mvccontext.Context) bool {
	accept := string(ctx.GetHeader("Accept"))

	// 如果同时包含HTML和XML，优先选择HTML
	if strings.Contains(accept, "text/html") {
		return false
	}

	// 只有明确请求XML且不包含HTML时才返回XML
	return (strings.Contains(accept, "application/xml") ||
		strings.Contains(accept, "text/xml")) &&
		!strings.Contains(accept, "text/html")
}

// renderJSONError 渲染JSON格式错误响应
func (c *DefaultErrorController) renderJSONError(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	response := map[string]any{
		"code":      errorCtx.StatusCode,
		"message":   errorCtx.ErrorMessage,
		"success":   false,
		"path":      errorCtx.RequestPath,
		"method":    errorCtx.RequestMethod,
		"timestamp": errorCtx.Timestamp.Unix(),
	}

	// 添加额外的上下文信息
	if c.ShowDetailedError {
		response["details"] = errorCtx.Details
		response["suggestions"] = errorCtx.Suggestions
		response["request_id"] = errorCtx.RequestID
	}

	ctx.JSON(errorCtx.StatusCode, response)
	return nil
}

// renderXMLError 渲染XML格式错误响应
func (c *DefaultErrorController) renderXMLError(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	xmlResponse := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<error>
    <code>%d</code>
    <title>%s</title>
    <message>%s</message>
    <path>%s</path>
    <method>%s</method>
    <timestamp>%s</timestamp>
    <request_id>%s</request_id>
</error>`,
		errorCtx.StatusCode,
		errorCtx.StatusText,
		errorCtx.ErrorMessage,
		errorCtx.RequestPath,
		errorCtx.RequestMethod,
		errorCtx.Timestamp.Format(time.RFC3339),
		errorCtx.RequestID,
	)

	ctx.SetContentType("application/xml")
	ctx.String(errorCtx.StatusCode, xmlResponse)
	return nil
}

// renderHTMLError 渲染HTML格式错误页面
func (c *DefaultErrorController) renderHTMLError(ctx *mvccontext.Context, errorCtx *ErrorContext) error {
	// 生成HTML页面
	html := c.generateErrorHTML(errorCtx)

	// 使用Data方法直接输出HTML内容
	ctx.Data(errorCtx.StatusCode, "text/html; charset=utf-8", []byte(html))
	return nil
}

// ============= HTML模板生成 =============

// generateErrorHTML 生成错误页面HTML
func (c *DefaultErrorController) generateErrorHTML(errorCtx *ErrorContext) string {
	// 获取状态对应的颜色和图标
	statusClass := c.getStatusClass(errorCtx.StatusCode)
	if statusClass == "" {
		statusClass = "status-error"
	}

	statusIcon := c.getStatusIcon(errorCtx.StatusCode)
	if statusIcon == "" {
		statusIcon = "❌"
	}

	// 构建建议列表HTML
	suggestionsHTML := c.buildSuggestionsHTML(errorCtx.Suggestions)

	// 构建调试信息HTML
	debugInfoHTML := c.buildDebugInfoHTML(errorCtx)

	// 正确的参数传递顺序
	return fmt.Sprintf(errorPageTemplate,
		c.CustomTitle,          // 1. 页面标题
		"YYHertz Framework",    // 2. header h1
		statusClass,            // 3. 状态CSS类
		statusIcon,             // 4. 状态图标
		errorCtx.StatusCode,    // 5. 状态码
		errorCtx.StatusText,    // 6. 状态文本
		errorCtx.StatusText,    // 7. 简短状态描述
		errorCtx.ErrorMessage,  // 8. 详细错误消息
		errorCtx.RequestPath,   // 9. 请求路径
		errorCtx.RequestMethod, // 10. 请求方法
		errorCtx.Timestamp.Format("2006-01-02 15:04:05"), // 11. 时间戳
		suggestionsHTML,        // 12. 建议列表
		debugInfoHTML,          // 13. 调试信息
		c.getSupportInfo(),     // 14. 支持信息
		errorCtx.StatusCode,    // 15. JavaScript中的状态码1
		errorCtx.RequestPath,   // 16. JavaScript中的请求路径
		errorCtx.RequestMethod, // 17. JavaScript中的请求方法
		errorCtx.StatusCode,    // 18. JavaScript中的状态码2
		errorCtx.StatusCode,    // 19. JavaScript中的状态码3
	)
}

// getStatusClass 获取状态对应的CSS类
func (c *DefaultErrorController) getStatusClass(statusCode int) string {
	switch {
	case statusCode == 400:
		return "status-400"
	case statusCode == 401:
		return "status-401"
	case statusCode == 402:
		return "status-402"
	case statusCode == 403:
		return "status-403"
	case statusCode == 404:
		return "status-404"
	case statusCode == 405:
		return "status-405"
	case statusCode == 406:
		return "status-406"
	case statusCode == 408:
		return "status-408"
	case statusCode == 409:
		return "status-409"
	case statusCode == 410:
		return "status-410"
	case statusCode == 413:
		return "status-413"
	case statusCode == 415:
		return "status-415"
	case statusCode == 418:
		return "status-418"
	case statusCode == 422:
		return "status-422"
	case statusCode == 429:
		return "status-429"
	case statusCode == 500:
		return "status-500"
	case statusCode == 501:
		return "status-501"
	case statusCode == 502:
		return "status-502"
	case statusCode == 503:
		return "status-503"
	case statusCode == 504:
		return "status-504"
	case statusCode == 505:
		return "status-505"
	case statusCode >= 400 && statusCode < 500:
		return "status-4xx"
	case statusCode >= 500:
		return "status-5xx"
	default:
		return "status-error"
	}
}

// getStatusIcon 获取状态对应的图标
func (c *DefaultErrorController) getStatusIcon(statusCode int) string {
	switch statusCode {
	case 400:
		return "❓" // Bad Request
	case 401:
		return "🔐" // Unauthorized
	case 402:
		return "💳" // Payment Required
	case 403:
		return "🚫" // Forbidden
	case 404:
		return "🔍" // Not Found
	case 405:
		return "🚷" // Method Not Allowed
	case 406:
		return "📋" // Not Acceptable
	case 408:
		return "⏰" // Request Timeout
	case 409:
		return "⚔️" // Conflict
	case 410:
		return "📱" // Gone
	case 413:
		return "📦" // Payload Too Large
	case 415:
		return "📄" // Unsupported Media Type
	case 418:
		return "🫖" // I'm a teapot
	case 422:
		return "📝" // Unprocessable Entity
	case 429:
		return "🚦" // Too Many Requests
	case 500:
		return "⚠️" // Internal Server Error
	case 501:
		return "🚧" // Not Implemented
	case 502:
		return "🌐" // Bad Gateway
	case 503:
		return "🔧" // Service Unavailable
	case 504:
		return "⏳" // Gateway Timeout
	case 505:
		return "📡" // HTTP Version Not Supported
	default:
		return "❌" // Generic Error
	}
}

// buildSuggestionsHTML 构建建议列表HTML
func (c *DefaultErrorController) buildSuggestionsHTML(suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}

	html := "<div><h3>解决建议</h3><ul class=\"suggestions\">"
	for _, suggestion := range suggestions {
		html += fmt.Sprintf("<li>%s</li>", suggestion)
	}
	html += "</ul></div>"
	return html
}

// buildDebugInfoHTML 构建调试信息HTML
func (c *DefaultErrorController) buildDebugInfoHTML(errorCtx *ErrorContext) string {
	if !c.EnableDebugInfo || len(errorCtx.Details) == 0 {
		return ""
	}

	html := `<div class="debug-info">
		<h3>调试信息 <span class="debug-label">开发环境</span></h3>
		<div class="debug-grid">`

	for key, value := range errorCtx.Details {
		html += fmt.Sprintf(`
			<div class="debug-item">
				<div class="debug-key">%s</div>
				<div class="debug-value">%v</div>
			</div>`, key, value)
	}

	html += "</div></div>"
	return html
}

// getSupportInfo 获取支持信息
func (c *DefaultErrorController) getSupportInfo() string {
	if c.SupportEmail == "" && c.SupportPhone == "" {
		return ""
	}

	html := `<div class="support-info">
		<h3>需要帮助？</h3>
		<div class="contact-info">`

	if c.SupportEmail != "" {
		html += fmt.Sprintf(`<div class="contact-item">
			<span class="contact-icon">📧</span>
			<a href="mailto:%s">%s</a>
		</div>`, c.SupportEmail, c.SupportEmail)
	}

	if c.SupportPhone != "" {
		html += fmt.Sprintf(`<div class="contact-item">
			<span class="contact-icon">📞</span>
			<a href="tel:%s">%s</a>
		</div>`, c.SupportPhone, c.SupportPhone)
	}

	html += "</div></div>"
	return html
}

// ============= HTML页面模板 =============

// errorPageTemplate HTML页面模板 - 简化版用于调试
const errorPageTemplateDebug = `<!DOCTYPE html>
<html>
<head><title>DEBUG: %s</title></head>
<body>
<h1>Header: %s</h1>
<div class="error-card %s">
<div class="status-icon">%s</div>
<h2>%d %s</h2>
<p>%s</p>
<p>Main message: %s</p>
<div>Request Path: %s</div>
<div>Request Method: %s</div>
<div>Timestamp: %s</div>
<div>%s</div>
<div>%s</div>
<div>%s</div>
JS: %d %d %s %s %d
</body>
</html>`

// errorPageTemplate HTML页面模板
const errorPageTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s - 错误页面</title>
	<style>
		body { 
			font-family: Arial, sans-serif; 
			margin: 0; 
			padding: 20px; 
			background: #f5f5f5; 
			line-height: 1.6; 
		}
		
		.container { 
			max-width: 800px; 
			margin: 0 auto; 
		}
		
		.header { 
			background: white; 
			padding: 30px; 
			border-radius: 8px; 
			margin-bottom: 20px; 
			box-shadow: 0 2px 4px rgba(0,0,0,0.1); 
			text-align: center; 
		}
		
		.error-card { 
			background: white; 
			padding: 30px; 
			border-radius: 8px; 
			margin-bottom: 20px; 
			box-shadow: 0 2px 4px rgba(0,0,0,0.1); 
		}
		
		/* 状态码特定样式 */
		.status-400 { border-left: 5px solid #ffc107; }  /* Bad Request */
		.status-401 { border-left: 5px solid #fd7e14; }  /* Unauthorized */
		.status-402 { border-left: 5px solid #20c997; }  /* Payment Required */
		.status-403 { border-left: 5px solid #dc3545; }  /* Forbidden */
		.status-404 { border-left: 5px solid #007bff; }  /* Not Found */
		.status-405 { border-left: 5px solid #e83e8c; }  /* Method Not Allowed */
		.status-406 { border-left: 5px solid #6610f2; }  /* Not Acceptable */
		.status-408 { border-left: 5px solid #fd7e14; }  /* Request Timeout */
		.status-409 { border-left: 5px solid #dc3545; }  /* Conflict */
		.status-410 { border-left: 5px solid #6c757d; }  /* Gone */
		.status-413 { border-left: 5px solid #ffc107; }  /* Payload Too Large */
		.status-415 { border-left: 5px solid #e83e8c; }  /* Unsupported Media */
		.status-418 { border-left: 5px solid #17a2b8; }  /* I'm a teapot */
		.status-422 { border-left: 5px solid #ffc107; }  /* Unprocessable Entity */
		.status-429 { border-left: 5px solid #fd7e14; }  /* Too Many Requests */
		.status-500 { border-left: 5px solid #6f42c1; }  /* Internal Server Error */
		.status-501 { border-left: 5px solid #6c757d; }  /* Not Implemented */
		.status-502 { border-left: 5px solid #dc3545; }  /* Bad Gateway */
		.status-503 { border-left: 5px solid #ffc107; }  /* Service Unavailable */
		.status-504 { border-left: 5px solid #fd7e14; }  /* Gateway Timeout */
		.status-505 { border-left: 5px solid #6c757d; }  /* HTTP Version Not Supported */
		.status-4xx { border-left: 5px solid #ffc107; }  /* Generic 4xx */
		.status-5xx { border-left: 5px solid #dc3545; }  /* Generic 5xx */
		.status-error { border-left: 5px solid #6c757d; } /* Generic Error */
		
		.error-header { 
			display: flex; 
			align-items: center; 
			margin-bottom: 20px; 
			flex-wrap: wrap; 
		}
		
		.status-icon { 
			font-size: 48px; 
			margin-right: 20px; 
		}
		
		.status-info h1 { 
			margin: 0; 
			color: #333; 
			font-size: 2.5em; 
		}
		
		.status-info p { 
			margin: 5px 0 0 0; 
			color: #666; 
			font-size: 1.2em; 
		}
		
		.error-details { 
			background: #f8f9fa; 
			padding: 20px; 
			border-radius: 6px; 
			margin: 20px 0; 
		}
		
		.error-details h3 { 
			margin-top: 0; 
			color: #333; 
		}
		
		.detail-item { 
			display: flex; 
			margin-bottom: 10px; 
		}
		
		.detail-label { 
			font-weight: bold; 
			width: 100px; 
			color: #555; 
		}
		
		.detail-value { 
			flex: 1; 
			color: #333; 
		}
		
		.suggestions { 
			list-style: none; 
			padding: 0; 
		}
		
		.suggestions li { 
			background: #e3f2fd; 
			padding: 12px 16px; 
			margin: 8px 0; 
			border-radius: 4px; 
			border-left: 4px solid #2196f3; 
		}
		
		.actions { 
			text-align: center; 
			margin: 30px 0; 
		}
		
		.btn { 
			padding: 12px 24px; 
			margin: 0 8px; 
			border: none; 
			border-radius: 4px; 
			cursor: pointer; 
			font-size: 16px; 
			text-decoration: none; 
			display: inline-block; 
			transition: background-color 0.3s; 
		}
		
		.btn-primary { 
			background: #007bff; 
			color: white; 
		}
		
		.btn-primary:hover { 
			background: #0056b3; 
		}
		
		.btn-secondary { 
			background: #6c757d; 
			color: white; 
		}
		
		.btn-secondary:hover { 
			background: #545b62; 
		}
		
		.btn-retry { 
			background: #17a2b8; 
			color: white; 
			position: relative;
			overflow: hidden;
		}
		
		.btn-retry:hover { 
			background: #138496; 
		}
		
		.btn-retry:disabled { 
			background: #6c757d; 
			cursor: not-allowed; 
		}
		
		.btn-warning { 
			background: #ffc107; 
			color: #212529; 
		}
		
		.btn-warning:hover { 
			background: #e0a800; 
		}
		
		.retry-progress {
			position: absolute;
			bottom: 0;
			left: 0;
			height: 3px;
			background: rgba(255,255,255,0.3);
			transition: width 1s linear;
		}
		
		.notification {
			position: fixed;
			top: 20px;
			right: 20px;
			padding: 15px 20px;
			border-radius: 4px;
			color: white;
			font-weight: bold;
			z-index: 1000;
			opacity: 0;
			transform: translateX(100%%);
			transition: all 0.3s ease;
		}
		
		.notification.show {
			opacity: 1;
			transform: translateX(0);
		}
		
		.notification.success {
			background: #28a745;
		}
		
		.notification.error {
			background: #dc3545;
		}
		
		.notification.info {
			background: #17a2b8;
		}
		
		.debug-info { 
			background: #fff3cd; 
			border: 1px solid #ffeaa7; 
			border-radius: 6px; 
			padding: 20px; 
			margin: 20px 0; 
		}
		
		.debug-label { 
			background: #fd7e14; 
			color: white; 
			padding: 2px 8px; 
			border-radius: 12px; 
			font-size: 12px; 
			font-weight: normal; 
		}
		
		.debug-grid { 
			display: grid; 
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); 
			gap: 15px; 
			margin-top: 15px; 
		}
		
		.debug-item { 
			background: white; 
			padding: 12px; 
			border-radius: 4px; 
		}
		
		.debug-key { 
			font-weight: bold; 
			color: #495057; 
			font-size: 12px; 
			text-transform: uppercase; 
			margin-bottom: 4px; 
		}
		
		.debug-value { 
			color: #007bff; 
			font-weight: bold; 
		}
		
		.support-info { 
			background: #d4edda; 
			border: 1px solid #c3e6cb; 
			border-radius: 6px; 
			padding: 20px; 
			margin: 20px 0; 
		}
		
		.contact-info { 
			display: flex; 
			flex-wrap: wrap; 
			gap: 20px; 
			margin-top: 15px; 
		}
		
		.contact-item { 
			display: flex; 
			align-items: center; 
		}
		
		.contact-icon { 
			margin-right: 8px; 
			font-size: 16px; 
		}
		
		.contact-item a { 
			color: #28a745; 
			text-decoration: none; 
			font-weight: bold; 
		}
		
		.contact-item a:hover { 
			text-decoration: underline; 
		}
		
		.footer { 
			text-align: center; 
			color: #6c757d; 
			padding: 20px; 
			font-size: 14px; 
		}
		
		/* 响应式设计 */
		@media (max-width: 768px) {
			body { padding: 10px; }
			.header, .error-card { padding: 20px; }
			.error-header { flex-direction: column; text-align: center; }
			.status-icon { margin-right: 0; margin-bottom: 10px; }
			.status-info h1 { font-size: 2em; }
			.detail-item { flex-direction: column; }
			.detail-label { width: auto; margin-bottom: 4px; }
			.actions { margin: 20px 0; }
			.btn { margin: 4px; padding: 10px 20px; }
			.debug-grid { grid-template-columns: 1fr; }
			.contact-info { flex-direction: column; gap: 10px; }
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>%s</h1>
			<p>应用遇到了一个问题，请查看下面的详细信息</p>
		</div>
		
		<div class="error-card %s">
			<div class="error-header">
				<div class="status-icon">%s</div>
				<div class="status-info">
					<h1>%d %s</h1>
					<p>%s</p>
				</div>
			</div>
			
			<div class="error-message">
				<p style="font-size: 1.1em; color: #333; margin-bottom: 20px;">%s</p>
			</div>
			
			<div class="error-details">
				<h3>🔍 请求详情</h3>
				<div class="detail-item">
					<div class="detail-label">请求路径:</div>
					<div class="detail-value">%s</div>
				</div>
				<div class="detail-item">
					<div class="detail-label">请求方法:</div>
					<div class="detail-value">%s</div>
				</div>
				<div class="detail-item">
					<div class="detail-label">时间:</div>
					<div class="detail-value">%s</div>
				</div>
			</div>
			
			%s
			
			<div class="actions">
				<a href="javascript:history.back()" class="btn btn-secondary">返回上页</a>
				<a href="/" class="btn btn-primary">返回首页</a>
				<button onclick="retryRequest()" class="btn btn-retry" id="retryBtn">
					重试
					<div class="retry-progress" id="retryProgress"></div>
				</button>
			</div>
			
			%s
			
			%s
		</div>
		
		<div class="footer">
			<p>&copy; YYHertz Framework. 技术支持团队随时为您服务。</p>
		</div>
	</div>
	
	<div id="notification" class="notification"></div>
	
	<script>
		// 基本页面信息
		const pageInfo = {
			statusCode: %d,
			path: '%s',
			method: '%s',
			timestamp: new Date().toISOString(),
			userAgent: navigator.userAgent,
			referrer: document.referrer || 'None'
		};
		
		// 页面加载完成后的初始化
		document.addEventListener('DOMContentLoaded', function() {
			// 添加页面加载动画
			document.body.style.opacity = '0';
			document.body.style.transition = 'opacity 0.5s ease-in-out';
			setTimeout(() => {
				document.body.style.opacity = '1';
			}, 100);
			
			// 显示页面信息（开发模式）
			console.group('🔧 页面错误信息');
			console.log('状态码:', pageInfo.statusCode);
			console.log('请求路径:', pageInfo.path);
			console.log('请求方法:', pageInfo.method);
			console.log('时间戳:', pageInfo.timestamp);
			console.log('用户代理:', pageInfo.userAgent);
			console.log('来源页面:', pageInfo.referrer);
			console.groupEnd();
		});
		
		// 重试请求功能
		function retryRequest() {
			const retryBtn = document.getElementById('retryBtn');
			const progressBar = document.getElementById('retryProgress');
			
			if (retryBtn.disabled) return;
			
			retryBtn.disabled = true;
			retryBtn.textContent = '重试中...';
			progressBar.style.width = '0%%';
			
			// 模拟进度
			let progress = 0;
			const progressInterval = setInterval(() => {
				progress += 10;
				progressBar.style.width = progress + '%%';
				
				if (progress >= 100) {
					clearInterval(progressInterval);
					// 重新加载当前页面
					window.location.reload();
				}
			}, 100);
		}
		
		// 显示通知
		function showNotification(message, type = 'info') {
			const notification = document.getElementById('notification');
			notification.textContent = message;
			notification.className = 'notification ' + type + ' show';
			
			setTimeout(() => {
				notification.classList.remove('show');
			}, 3000);
		}
		
		// 键盘快捷键支持
		document.addEventListener('keydown', function(e) {
			if (e.ctrlKey || e.metaKey) {
				switch(e.key) {
					case 'r':
						e.preventDefault();
						retryRequest();
						break;
					case 'h':
						e.preventDefault();
						window.location.href = '/';
						break;
				}
			}
			
			if (e.key === 'Escape') {
				history.back();
			}
		});
		
		// 监控状态码用于分析
		if (window.gtag) {
			gtag('event', 'error_page_view', {
				'error_code': %d,
				'error_path': pageInfo.path,
				'error_method': pageInfo.method
			});
		}
		
		// 开发环境调试信息
		if (location.hostname === 'localhost' || location.hostname === '127.0.0.1') {
			const statusCode = %d;
			if (statusCode >= 500) {
				console.warn('🚨 服务器错误 - 请检查服务器日志');
			} else if (statusCode >= 400) {
				console.info('ℹ️ 客户端错误 - 请检查请求参数');
			}
		}
	</script>
</body>
</html>`

// ============= 配置结构 =============

// ErrorControllerConfig 错误控制器配置结构
type ErrorControllerConfig struct {
	ShowDetailedError bool   `json:"show_detailed_error"` // 是否显示详细错误信息
	Language          string `json:"language"`            // 语言设置
	CustomTitle       string `json:"custom_title"`        // 自定义页面标题
	SupportEmail      string `json:"support_email"`       // 支持邮箱
	SupportPhone      string `json:"support_phone"`       // 支持电话
	EnableDebugInfo   bool   `json:"enable_debug_info"`   // 是否启用调试信息
}

// DefaultErrorControllerConfig 返回默认错误控制器配置
func DefaultErrorControllerConfig() ErrorControllerConfig {
	return ErrorControllerConfig{
		ShowDetailedError: true,
		Language:          "zh-CN",
		CustomTitle:       "YYHertz Framework",
		SupportEmail:      "",
		SupportPhone:      "",
		EnableDebugInfo:   true,
	}
}

// SetErrorControllerConfig 配置错误控制器行为
func SetErrorControllerConfig(controller *DefaultErrorController, config ErrorControllerConfig) {
	if controller == nil {
		return
	}

	controller.ShowDetailedError = config.ShowDetailedError
	controller.Language = config.Language
	controller.CustomTitle = config.CustomTitle
	controller.SupportEmail = config.SupportEmail
	controller.SupportPhone = config.SupportPhone
	controller.EnableDebugInfo = config.EnableDebugInfo
}

// EnableErrorDebugging 开启/关闭错误调试模式
func EnableErrorDebugging(controller *DefaultErrorController, enabled bool) {
	if controller == nil {
		return
	}

	controller.EnableDebugInfo = enabled
	controller.ShowDetailedError = enabled
}

// ============= 便捷工厂函数 =============

// CreateDefaultErrorHandlers 创建一套默认错误处理器
func CreateDefaultErrorHandlers() map[int]ErrorHandler {
	handlers := make(map[int]ErrorHandler)

	// 创建默认控制器
	controller := NewDefaultErrorController()

	// 为常见状态码创建专用处理器
	handlers[401] = controller
	handlers[403] = controller
	handlers[404] = controller
	handlers[500] = controller

	return handlers
}

// RegisterDefaultHandlers 注册默认错误处理器到注册器
func RegisterDefaultHandlers(registry *ErrorRegistry) error {
	if registry == nil {
		return fmt.Errorf("registry cannot be nil")
	}

	handlers := CreateDefaultErrorHandlers()
	for statusCode, handler := range handlers {
		if err := registry.RegisterHandler(statusCode, handler); err != nil {
			return fmt.Errorf("failed to register handler for status %d: %w", statusCode, err)
		}
	}

	return nil
}

// QuickSetupDefaultHandlers 快速设置默认错误处理器
func QuickSetupDefaultHandlers(registry *ErrorRegistry, env string) error {
	if registry == nil {
		return fmt.Errorf("registry cannot be nil")
	}

	var controller *DefaultErrorController

	// 根据环境创建相应的控制器
	switch env {
	case "production", "prod":
		controller = NewProductionErrorController()
	case "development", "dev":
		controller = NewDevelopmentErrorController()
	default:
		controller = NewDefaultErrorController()
	}

	// 注册更多的错误状态码
	statusCodes := []int{400, 401, 402, 403, 404, 405, 406, 408, 409, 410, 413, 415, 418, 422, 429, 500, 501, 502, 503, 504, 505}
	for _, statusCode := range statusCodes {
		if err := registry.RegisterHandler(statusCode, controller); err != nil {
			return fmt.Errorf("failed to register handler for status %d: %w", statusCode, err)
		}
	}

	return nil
}

// ============= 智能错误处理辅助方法 =============

// getStatusMessage 获取状态码对应的中文描述
func (c *DefaultErrorController) getStatusMessage(statusCode int) string {
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

// generateSmartSuggestions 智能生成解决建议
func (c *DefaultErrorController) generateSmartSuggestions(statusCode int, ctx *mvccontext.Context) []string {
	switch statusCode {
	case 400:
		return []string{
			"检查请求参数是否正确",
			"确认数据格式符合API要求",
			"查看API文档了解正确的请求格式",
		}
	case 401:
		return []string{
			"请先登录您的账户",
			"检查您的登录状态是否过期",
			"确认API密钥或访问令牌是否有效",
		}
	case 402:
		return []string{
			"请联系管理员升级您的账户",
			"查看付费计划了解更多详情",
			"确认您的订阅状态",
		}
	case 403:
		return []string{
			"请联系管理员申请相应权限",
			"确认您的账户具有访问此资源的权限",
			"检查您是否登录了正确的账户",
		}
	case 404:
		return []string{
			"检查URL拼写是否正确",
			"尝试返回首页重新导航",
			"清除浏览器缓存后重试",
		}
	case 405:
		return []string{
			"检查请求方法（GET、POST、PUT、DELETE）是否正确",
			"查看API文档了解支持的HTTP方法",
			"尝试使用不同的请求方法",
		}
	case 408:
		return []string{
			"检查网络连接是否稳定",
			"尝试减少请求数据量",
			"稍后重试请求",
		}
	case 409:
		return []string{
			"检查是否有资源冲突",
			"确认操作是否已经执行过",
			"刷新页面获取最新状态后重试",
		}
	case 413:
		return []string{
			"减少上传文件的大小",
			"分批上传大型数据",
			"联系管理员提高上传限制",
		}
	case 415:
		return []string{
			"检查文件格式是否被支持",
			"尝试使用标准的媒体类型",
			"查看支持的文件格式列表",
		}
	case 418:
		return []string{
			"这是一个彩蛋错误，恭喜你发现了！",
			"尝试使用咖啡机而不是茶壶",
			"RFC 2324 - 超文本咖啡壶控制协议",
		}
	case 422:
		return []string{
			"检查提交的数据是否符合业务规则",
			"确认必填字段都已填写",
			"验证数据格式和内容是否正确",
		}
	case 429:
		return []string{
			"请降低请求频率",
			"等待一段时间后重试",
			"考虑使用缓存减少重复请求",
		}
	case 500:
		return []string{
			"请稍后重试",
			"如果问题持续存在，请联系技术支持",
			"您也可以尝试刷新页面",
		}
	case 501:
		return []string{
			"尝试使用其他可用的功能",
			"联系技术支持了解功能开发计划",
			"查看API文档了解已实现的功能",
		}
	case 502:
		return []string{
			"请稍后重试",
			"检查网络连接是否正常",
			"如果问题持续存在，请联系技术支持",
		}
	case 503:
		return []string{
			"服务正在维护中，请稍后重试",
			"关注官方公告了解维护时间",
			"使用其他可用的服务入口",
		}
	case 504:
		return []string{
			"请稍后重试",
			"检查网络连接是否稳定",
			"如果问题持续存在，请联系技术支持",
		}
	case 505:
		return []string{
			"升级您的客户端版本",
			"使用标准的HTTP协议版本",
			"联系技术支持获取兼容性信息",
		}
	default:
		// 默认建议
		if statusCode >= 400 && statusCode < 500 {
			return []string{
				"检查请求是否正确",
				"查看相关文档了解正确的使用方法",
				"如需帮助请联系技术支持",
			}
		} else if statusCode >= 500 {
			return []string{
				"请稍后重试",
				"检查网络连接是否正常",
				"如果问题持续存在，请联系技术支持",
			}
		}
		return []string{
			"请稍后重试",
			"如果问题持续存在，请联系技术支持",
		}
	}
}

// ============= 错误钩子函数 =============

// logErrorHook 错误日志记录钩子
func (c *DefaultErrorController) logErrorHook(ctx *mvccontext.Context, statusCode int, err error) error {
	if !c.EnableErrorLogging || c.ErrorLogger == nil {
		return nil
	}

	requestID := ""
	if val, exists := ctx.Get("request_id"); exists {
		if id, ok := val.(string); ok {
			requestID = id
		}
	}

	logMsg := fmt.Sprintf("Status: %d, Path: %s, Method: %s, RequestID: %s, Error: %v",
		statusCode, string(ctx.Path()), string(ctx.Method()), requestID, err)

	c.ErrorLogger.Println(logMsg)
	return nil
}

// statisticsHook 错误统计钩子
func (c *DefaultErrorController) statisticsHook(ctx *mvccontext.Context, statusCode int, err error) error {
	if c.ErrorStatistics == nil {
		return nil
	}

	c.ErrorStatistics.mu.Lock()
	defer c.ErrorStatistics.mu.Unlock()

	// 更新统计信息
	c.ErrorStatistics.TotalErrors++
	c.ErrorStatistics.ErrorsByStatus[statusCode]++

	path := string(ctx.Path())
	c.ErrorStatistics.ErrorsByPath[path]++

	// 添加到最近错误列表（保留最近50个）
	record := ErrorRecord{
		StatusCode: statusCode,
		Path:       path,
		Method:     string(ctx.Method()),
		Message:    err.Error(),
		Timestamp:  time.Now(),
		UserAgent:  string(ctx.UserAgent()),
	}

	if val, exists := ctx.Get("request_id"); exists {
		if id, ok := val.(string); ok {
			record.RequestID = id
		}
	}

	c.ErrorStatistics.LastErrors = append(c.ErrorStatistics.LastErrors, record)
	if len(c.ErrorStatistics.LastErrors) > 50 {
		c.ErrorStatistics.LastErrors = c.ErrorStatistics.LastErrors[1:]
	}

	return nil
}

// cleanupHook 清理钩子
func (c *DefaultErrorController) cleanupHook(ctx *mvccontext.Context, statusCode int, err error) error {
	// 可以在这里进行资源清理，比如关闭数据库连接等
	// 当前实现为空，留给用户自定义
	return nil
}

// ============= 新增公共方法 =============

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
	if c.ErrorStatistics == nil {
		return nil
	}

	c.ErrorStatistics.mu.RLock()
	defer c.ErrorStatistics.mu.RUnlock()

	// 返回副本以避免并发问题
	stats := &ErrorStatistics{
		TotalErrors:    c.ErrorStatistics.TotalErrors,
		ErrorsByStatus: make(map[int]int64),
		ErrorsByPath:   make(map[string]int64),
		LastErrors:     make([]ErrorRecord, len(c.ErrorStatistics.LastErrors)),
		StartTime:      c.ErrorStatistics.StartTime,
	}

	for k, v := range c.ErrorStatistics.ErrorsByStatus {
		stats.ErrorsByStatus[k] = v
	}

	for k, v := range c.ErrorStatistics.ErrorsByPath {
		stats.ErrorsByPath[k] = v
	}

	copy(stats.LastErrors, c.ErrorStatistics.LastErrors)

	return stats
}

// ResetStatistics 重置错误统计信息
func (c *DefaultErrorController) ResetStatistics() {
	if c.ErrorStatistics == nil {
		return
	}

	c.ErrorStatistics.mu.Lock()
	defer c.ErrorStatistics.mu.Unlock()

	c.ErrorStatistics.TotalErrors = 0
	c.ErrorStatistics.ErrorsByStatus = make(map[int]int64)
	c.ErrorStatistics.ErrorsByPath = make(map[string]int64)
	c.ErrorStatistics.LastErrors = make([]ErrorRecord, 0)
	c.ErrorStatistics.StartTime = time.Now()
}

// IsRetryable 检查错误是否可重试
func (c *DefaultErrorController) IsRetryable(statusCode int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.RetryableErrors[statusCode]
}

// SetRetryable 设置错误是否可重试
func (c *DefaultErrorController) SetRetryable(statusCode int, retryable bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.RetryableErrors[statusCode] = retryable
}

// logError 内部日志记录方法
func (c *DefaultErrorController) logError(prefix string, err error) {
	if c.EnableErrorLogging && c.ErrorLogger != nil {
		c.ErrorLogger.Printf("%s: %v", prefix, err)
	}
}

// ============= 多语言支持 =============

// getLocalizedMessage 获取本地化消息
func (c *DefaultErrorController) getLocalizedMessage(key string, messages map[string]I18nMessage) string {
	if msg, exists := messages[key]; exists {
		switch c.Language {
		case "en-US", "en":
			if msg.EnUS != "" {
				return msg.EnUS
			}
		case "zh-CN", "zh":
			if msg.ZhCN != "" {
				return msg.ZhCN
			}
		}
	}
	return key // 如果没有找到对应的翻译，返回key本身
}

// ============= 业务错误码支持 =============

// BusinessError 业务错误结构
type BusinessError struct {
	Code    string `json:"code"`    // 业务错误码
	Message string `json:"message"` // 错误消息
	Data    any    `json:"data"`    // 附加数据
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// HandleBusinessError 处理业务错误
func (c *DefaultErrorController) HandleBusinessError(ctx *mvccontext.Context, bizErr *BusinessError) error {
	// 业务错误通常使用200状态码，但在错误体中包含错误信息
	statusCode := 200
	if c.shouldUseErrorStatusForBusiness(bizErr.Code) {
		statusCode = c.getBusinessErrorStatusCode(bizErr.Code)
	}

	// 构建错误上下文
	errorCtx := &ErrorContext{
		StatusCode:    statusCode,
		StatusText:    "业务处理错误",
		ErrorMessage:  bizErr.Message,
		RequestPath:   string(ctx.Path()),
		RequestMethod: string(ctx.Method()),
		UserAgent:     string(ctx.UserAgent()),
		Timestamp:     time.Now(),
		Details:       make(map[string]any),
		Suggestions:   c.getBusinessErrorSuggestions(bizErr.Code),
	}

	// 添加业务错误的详细信息
	errorCtx.Details["business_code"] = bizErr.Code
	if bizErr.Data != nil {
		errorCtx.Details["business_data"] = bizErr.Data
	}

	// 业务错误优先返回JSON格式
	return c.renderJSONError(ctx, errorCtx)
}

// shouldUseErrorStatusForBusiness 判断业务错误是否应该使用HTTP错误状态码
func (c *DefaultErrorController) shouldUseErrorStatusForBusiness(code string) bool {
	// 某些严重的业务错误应该使用HTTP错误状态码
	severeCodes := []string{
		"AUTH_FAILED",        // 认证失败
		"PERMISSION_DENIED",  // 权限不足
		"RESOURCE_NOT_FOUND", // 资源不存在
		"RATE_LIMITED",       // 频率限制
		"SYSTEM_ERROR",       // 系统错误
	}

	for _, severe := range severeCodes {
		if code == severe {
			return true
		}
	}
	return false
}

// getBusinessErrorStatusCode 获取业务错误对应的HTTP状态码
func (c *DefaultErrorController) getBusinessErrorStatusCode(code string) int {
	switch code {
	case "AUTH_FAILED":
		return 401
	case "PERMISSION_DENIED":
		return 403
	case "RESOURCE_NOT_FOUND":
		return 404
	case "RATE_LIMITED":
		return 429
	case "SYSTEM_ERROR":
		return 500
	default:
		return 400
	}
}

// getBusinessErrorSuggestions 获取业务错误的建议
func (c *DefaultErrorController) getBusinessErrorSuggestions(code string) []string {
	switch code {
	case "AUTH_FAILED":
		return []string{
			"请检查您的登录凭据",
			"确认用户名和密码是否正确",
			"如果忘记密码，请使用密码重置功能",
		}
	case "PERMISSION_DENIED":
		return []string{
			"请联系管理员申请相应权限",
			"确认您的账户角色和权限设置",
			"检查是否登录了正确的账户",
		}
	case "RESOURCE_NOT_FOUND":
		return []string{
			"确认请求的资源是否存在",
			"检查资源ID或标识符是否正确",
			"资源可能已被删除或移动",
		}
	case "RATE_LIMITED":
		return []string{
			"请降低请求频率",
			"等待一段时间后重试",
			"考虑升级账户以获得更高的请求限制",
		}
	case "VALIDATION_ERROR":
		return []string{
			"检查输入数据的格式和内容",
			"确认必填字段都已正确填写",
			"参考数据验证规则进行修正",
		}
	case "SYSTEM_ERROR":
		return []string{
			"系统暂时出现问题，请稍后重试",
			"如果问题持续存在，请联系技术支持",
			"您可以尝试使用其他功能",
		}
	default:
		return []string{
			"请检查操作是否正确",
			"如需帮助请联系客服",
			"您可以尝试重新操作",
		}
	}
}
