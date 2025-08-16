package errors

import (
	"strings"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// =============================================================================
// 模块：智能建议生成器
// 职责：根据状态码和上下文生成智能化的解决建议
// =============================================================================

// getErrorCategory 获取错误类别
func GetErrorCategory(statusCode int) string {
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

// generateContextualAdvice 生成上下文相关的建议
func GenerateContextualAdvice(statusCode int, ctx *mvccontext.Context) []string {
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
func GetRecoveryInstructions(statusCode int) []string {
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

// getPreventionTips 获取预防措施建议
func GetPreventionTips(statusCode int) []string {
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

// generateSmartSuggestions 智能生成解决建议
func GenerateSmartSuggestions(statusCode int, ctx *mvccontext.Context) []string {
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