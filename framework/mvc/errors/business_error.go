package errors

import (
	"time"
	
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// =============================================================================
// 模块：业务错误处理
// 职责：处理业务逻辑错误，支持自定义错误码和状态映射
// =============================================================================

// HandleBusinessError 处理业务错误
func HandleBusinessError(ctx *mvccontext.Context, bizErr *BusinessError, showDetailedError bool) error {
	// 业务错误通常使用200状态码，但在错误体中包含错误信息
	statusCode := 200
	if shouldUseErrorStatusForBusiness(bizErr.Code) {
		statusCode = getBusinessErrorStatusCode(bizErr.Code)
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
		Suggestions:   getBusinessErrorSuggestions(bizErr.Code),
	}

	// 添加业务错误的详细信息
	errorCtx.Details["business_code"] = bizErr.Code
	if bizErr.Data != nil {
		errorCtx.Details["business_data"] = bizErr.Data
	}

	// 业务错误优先返回JSON格式
	return RenderJSONError(ctx, errorCtx, showDetailedError)
}

// shouldUseErrorStatusForBusiness 判断业务错误是否应该使用HTTP错误状态码
func shouldUseErrorStatusForBusiness(code string) bool {
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
func getBusinessErrorStatusCode(code string) int {
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
func getBusinessErrorSuggestions(code string) []string {
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