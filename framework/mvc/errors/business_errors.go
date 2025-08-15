// Package errors 提供业务错误码的便捷创建方法
package errors

import "fmt"

// ============= 业务错误码常量 =============

const (
	// 认证相关错误
	AUTH_FAILED           = "AUTH_FAILED"           // 认证失败
	TOKEN_EXPIRED         = "TOKEN_EXPIRED"         // Token过期
	INVALID_CREDENTIALS   = "INVALID_CREDENTIALS"   // 无效凭据
	
	// 权限相关错误
	PERMISSION_DENIED     = "PERMISSION_DENIED"     // 权限不足
	ACCESS_DENIED         = "ACCESS_DENIED"         // 访问被拒绝
	INSUFFICIENT_PRIVILEGES = "INSUFFICIENT_PRIVILEGES" // 权限不足
	
	// 资源相关错误
	RESOURCE_NOT_FOUND    = "RESOURCE_NOT_FOUND"    // 资源不存在
	RESOURCE_LOCKED       = "RESOURCE_LOCKED"       // 资源被锁定
	RESOURCE_CONFLICT     = "RESOURCE_CONFLICT"     // 资源冲突
	RESOURCE_EXPIRED      = "RESOURCE_EXPIRED"      // 资源过期
	
	// 数据验证错误
	VALIDATION_ERROR      = "VALIDATION_ERROR"      // 数据验证错误
	INVALID_PARAMETER     = "INVALID_PARAMETER"     // 无效参数
	MISSING_PARAMETER     = "MISSING_PARAMETER"     // 缺失参数
	PARAMETER_TOO_LONG    = "PARAMETER_TOO_LONG"    // 参数过长
	PARAMETER_TOO_SHORT   = "PARAMETER_TOO_SHORT"   // 参数过短
	
	// 业务流程错误
	BUSINESS_LOGIC_ERROR  = "BUSINESS_LOGIC_ERROR"  // 业务逻辑错误
	OPERATION_NOT_ALLOWED = "OPERATION_NOT_ALLOWED" // 操作不被允许
	PRECONDITION_FAILED   = "PRECONDITION_FAILED"   // 前置条件失败
	WORKFLOW_ERROR        = "WORKFLOW_ERROR"        // 工作流错误
	
	// 限制相关错误
	RATE_LIMITED          = "RATE_LIMITED"          // 频率限制
	QUOTA_EXCEEDED        = "QUOTA_EXCEEDED"        // 配额超限
	CONCURRENT_LIMIT      = "CONCURRENT_LIMIT"      // 并发限制
	
	// 系统相关错误
	SYSTEM_ERROR          = "SYSTEM_ERROR"          // 系统错误
	DATABASE_ERROR        = "DATABASE_ERROR"        // 数据库错误
	NETWORK_ERROR         = "NETWORK_ERROR"         // 网络错误
	TIMEOUT_ERROR         = "TIMEOUT_ERROR"         // 超时错误
	SERVICE_UNAVAILABLE   = "SERVICE_UNAVAILABLE"   // 服务不可用
	
	// 第三方服务错误
	EXTERNAL_SERVICE_ERROR = "EXTERNAL_SERVICE_ERROR" // 外部服务错误
	API_CALL_FAILED       = "API_CALL_FAILED"        // API调用失败
	PAYMENT_FAILED        = "PAYMENT_FAILED"         // 支付失败
)

// ============= 业务错误构造器 =============

// NewBusinessError 创建业务错误
func NewBusinessError(code, message string, data ...any) *BusinessError {
	bizErr := &BusinessError{
		Code:    code,
		Message: message,
	}
	
	if len(data) > 0 {
		bizErr.Data = data[0]
	}
	
	return bizErr
}

// NewBusinessErrorf 创建格式化消息的业务错误
func NewBusinessErrorf(code, format string, args ...any) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// ============= 认证相关错误构造器 =============

// AuthFailed 认证失败错误
func AuthFailed(message ...string) *BusinessError {
	msg := "认证失败"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(AUTH_FAILED, msg)
}

// TokenExpired Token过期错误
func TokenExpired(message ...string) *BusinessError {
	msg := "访问令牌已过期"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(TOKEN_EXPIRED, msg)
}

// InvalidCredentials 无效凭据错误
func InvalidCredentials(message ...string) *BusinessError {
	msg := "用户名或密码错误"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(INVALID_CREDENTIALS, msg)
}

// ============= 权限相关错误构造器 =============

// PermissionDenied 权限不足错误
func PermissionDenied(message ...string) *BusinessError {
	msg := "权限不足，无法访问该资源"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(PERMISSION_DENIED, msg)
}

// AccessDenied 访问被拒绝错误
func AccessDenied(message ...string) *BusinessError {
	msg := "访问被拒绝"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(ACCESS_DENIED, msg)
}

// ============= 资源相关错误构造器 =============

// ResourceNotFound 资源不存在错误
func ResourceNotFound(resourceType string, message ...string) *BusinessError {
	msg := fmt.Sprintf("%s不存在", resourceType)
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(RESOURCE_NOT_FOUND, msg, map[string]string{
		"resource_type": resourceType,
	})
}

// ResourceLocked 资源被锁定错误
func ResourceLocked(resourceType string, message ...string) *BusinessError {
	msg := fmt.Sprintf("%s正在被其他用户使用", resourceType)
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(RESOURCE_LOCKED, msg, map[string]string{
		"resource_type": resourceType,
	})
}

// ResourceConflict 资源冲突错误
func ResourceConflict(message ...string) *BusinessError {
	msg := "资源状态冲突"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(RESOURCE_CONFLICT, msg)
}

// ============= 数据验证错误构造器 =============

// ValidationError 数据验证错误
func ValidationError(field string, message ...string) *BusinessError {
	msg := fmt.Sprintf("字段%s验证失败", field)
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(VALIDATION_ERROR, msg, map[string]string{
		"field": field,
	})
}

// InvalidParameter 无效参数错误
func InvalidParameter(param string, message ...string) *BusinessError {
	msg := fmt.Sprintf("参数%s无效", param)
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(INVALID_PARAMETER, msg, map[string]string{
		"parameter": param,
	})
}

// MissingParameter 缺失参数错误
func MissingParameter(param string) *BusinessError {
	return NewBusinessError(MISSING_PARAMETER, 
		fmt.Sprintf("缺少必需参数: %s", param), 
		map[string]string{"parameter": param})
}

// ParameterTooLong 参数过长错误
func ParameterTooLong(param string, maxLength int) *BusinessError {
	return NewBusinessError(PARAMETER_TOO_LONG,
		fmt.Sprintf("参数%s过长，最大长度为%d", param, maxLength),
		map[string]any{
			"parameter":  param,
			"max_length": maxLength,
		})
}

// ParameterTooShort 参数过短错误
func ParameterTooShort(param string, minLength int) *BusinessError {
	return NewBusinessError(PARAMETER_TOO_SHORT,
		fmt.Sprintf("参数%s过短，最小长度为%d", param, minLength),
		map[string]any{
			"parameter":  param,
			"min_length": minLength,
		})
}

// ============= 业务流程错误构造器 =============

// BusinessLogicError 业务逻辑错误
func BusinessLogicError(message string) *BusinessError {
	return NewBusinessError(BUSINESS_LOGIC_ERROR, message)
}

// OperationNotAllowed 操作不被允许错误
func OperationNotAllowed(operation string, reason ...string) *BusinessError {
	msg := fmt.Sprintf("操作%s不被允许", operation)
	if len(reason) > 0 {
		msg += ": " + reason[0]
	}
	return NewBusinessError(OPERATION_NOT_ALLOWED, msg, map[string]string{
		"operation": operation,
	})
}

// PreconditionFailed 前置条件失败错误
func PreconditionFailed(condition string) *BusinessError {
	return NewBusinessError(PRECONDITION_FAILED,
		fmt.Sprintf("前置条件不满足: %s", condition),
		map[string]string{"condition": condition})
}

// ============= 限制相关错误构造器 =============

// RateLimited 频率限制错误
func RateLimited(limit int, window string) *BusinessError {
	return NewBusinessError(RATE_LIMITED,
		fmt.Sprintf("请求过于频繁，%s内最多允许%d次请求", window, limit),
		map[string]any{
			"limit":  limit,
			"window": window,
		})
}

// QuotaExceeded 配额超限错误
func QuotaExceeded(quotaType string, limit int) *BusinessError {
	return NewBusinessError(QUOTA_EXCEEDED,
		fmt.Sprintf("%s配额已用完，当前限制为%d", quotaType, limit),
		map[string]any{
			"quota_type": quotaType,
			"limit":      limit,
		})
}

// ============= 系统相关错误构造器 =============

// SystemError 系统错误
func SystemError(message ...string) *BusinessError {
	msg := "系统内部错误"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(SYSTEM_ERROR, msg)
}

// DatabaseError 数据库错误
func DatabaseError(operation string, message ...string) *BusinessError {
	msg := fmt.Sprintf("数据库%s操作失败", operation)
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(DATABASE_ERROR, msg, map[string]string{
		"operation": operation,
	})
}

// NetworkError 网络错误
func NetworkError(message ...string) *BusinessError {
	msg := "网络连接错误"
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(NETWORK_ERROR, msg)
}

// TimeoutError 超时错误
func TimeoutError(operation string, timeout int) *BusinessError {
	return NewBusinessError(TIMEOUT_ERROR,
		fmt.Sprintf("%s操作超时，超时时间为%d秒", operation, timeout),
		map[string]any{
			"operation": operation,
			"timeout":   timeout,
		})
}

// ServiceUnavailable 服务不可用错误
func ServiceUnavailable(service string, message ...string) *BusinessError {
	msg := fmt.Sprintf("%s服务暂时不可用", service)
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(SERVICE_UNAVAILABLE, msg, map[string]string{
		"service": service,
	})
}

// ============= 第三方服务错误构造器 =============

// ExternalServiceError 外部服务错误
func ExternalServiceError(service string, message ...string) *BusinessError {
	msg := fmt.Sprintf("外部服务%s调用失败", service)
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(EXTERNAL_SERVICE_ERROR, msg, map[string]string{
		"service": service,
	})
}

// APICallFailed API调用失败错误
func APICallFailed(api string, statusCode int, message ...string) *BusinessError {
	msg := fmt.Sprintf("API %s 调用失败 (HTTP %d)", api, statusCode)
	if len(message) > 0 {
		msg = message[0]
	}
	return NewBusinessError(API_CALL_FAILED, msg, map[string]any{
		"api":         api,
		"status_code": statusCode,
	})
}

// PaymentFailed 支付失败错误
func PaymentFailed(reason string, data ...any) *BusinessError {
	bizErr := NewBusinessError(PAYMENT_FAILED, fmt.Sprintf("支付失败: %s", reason))
	if len(data) > 0 {
		bizErr.Data = data[0]
	}
	return bizErr
}

// ============= 快速判断方法 =============

// IsAuthError 判断是否为认证错误
func IsAuthError(err error) bool {
	if bizErr, ok := err.(*BusinessError); ok {
		return bizErr.Code == AUTH_FAILED || 
		       bizErr.Code == TOKEN_EXPIRED || 
		       bizErr.Code == INVALID_CREDENTIALS
	}
	return false
}

// IsPermissionError 判断是否为权限错误  
func IsPermissionError(err error) bool {
	if bizErr, ok := err.(*BusinessError); ok {
		return bizErr.Code == PERMISSION_DENIED || 
		       bizErr.Code == ACCESS_DENIED ||
		       bizErr.Code == INSUFFICIENT_PRIVILEGES
	}
	return false
}

// IsValidationError 判断是否为验证错误
func IsValidationError(err error) bool {
	if bizErr, ok := err.(*BusinessError); ok {
		return bizErr.Code == VALIDATION_ERROR ||
		       bizErr.Code == INVALID_PARAMETER ||
		       bizErr.Code == MISSING_PARAMETER ||
		       bizErr.Code == PARAMETER_TOO_LONG ||
		       bizErr.Code == PARAMETER_TOO_SHORT
	}
	return false
}

// IsSystemError 判断是否为系统错误
func IsSystemError(err error) bool {
	if bizErr, ok := err.(*BusinessError); ok {
		return bizErr.Code == SYSTEM_ERROR ||
		       bizErr.Code == DATABASE_ERROR ||
		       bizErr.Code == NETWORK_ERROR ||
		       bizErr.Code == TIMEOUT_ERROR ||
		       bizErr.Code == SERVICE_UNAVAILABLE
	}
	return false
}