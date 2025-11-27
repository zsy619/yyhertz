package mvc

import (
	"fmt"

	"github.com/zsy619/yyhertz/framework/mvc/security"
)

// ============= 全局CSRF Token API =============

// GenerateCSRFToken 生成CSRF token（全局方法）
//
// 这是一个全局方法，用于生成CSRF token。它使用全局CSRF管理器，
// 适用于在控制器、中间件或其他任何地方需要生成CSRF token的场景。
//
// 参数：
//   - userID: 用户标识（可以是session ID、用户ID等）
//   - clientIP: 客户端IP地址
//
// 返回：
//   - *security.CSRFToken: 生成的CSRF token对象
//   - error: 错误信息
//
// 示例：
//
//	token, err := mvc.GenerateCSRFToken("user123", "192.168.1.1")
//	if err != nil {
//	    log.Printf("生成CSRF token失败: %v", err)
//	    return
//	}
//	// 使用token.Value作为CSRF token值
//	ctx.Set("csrf_token", token.Value)
func GenerateCSRFToken(userID, clientIP string) (*security.CSRFToken, error) {
	manager := GetCSRFManager()
	if manager == nil {
		return nil, fmt.Errorf("CSRF manager is not initialized")
	}
	
	return manager.GenerateToken(userID, clientIP)
}

// ValidateCSRFToken 验证CSRF token（全局方法）
//
// 这是一个全局方法，用于验证CSRF token。它使用全局CSRF管理器，
// 适用于在控制器、中间件或其他任何地方需要验证CSRF token的场景。
//
// 参数：
//   - tokenValue: 要验证的CSRF token值
//   - userID: 用户标识（应与生成时使用的相同）
//   - clientIP: 客户端IP地址（应与生成时使用的相同）
//
// 返回：
//   - bool: 验证是否成功
//   - error: 错误信息
//
// 示例：
//
//	isValid, err := mvc.ValidateCSRFToken(tokenValue, "user123", "192.168.1.1")
//	if err != nil {
//	    log.Printf("验证CSRF token失败: %v", err)
//	    return false
//	}
//	if !isValid {
//	    log.Println("CSRF token验证失败")
//	    return false
//	}
func ValidateCSRFToken(tokenValue, userID, clientIP string) (bool, error) {
	manager := GetCSRFManager()
	if manager == nil {
		return false, fmt.Errorf("CSRF manager is not initialized")
	}
	
	return manager.ValidateToken(tokenValue, userID, clientIP)
}

// GenerateSimpleCSRFToken 生成简单的CSRF token（全局方法）
//
// 这是一个全局方法，用于生成简单的CSRF token。这种token不依赖用户ID和IP，
// 主要用于模板引擎等需要占位符token的场景。
//
// 返回：
//   - string: 生成的简单CSRF token值
//
// 示例：
//
//	token := mvc.GenerateSimpleCSRFToken()
//	// 在模板数据中设置CSRF token
//	templateData := map[string]interface{}{
//	    "csrf_token": token,
//	}
func GenerateSimpleCSRFToken() string {
	manager := GetCSRFManager()
	if manager == nil {
		return "csrf-placeholder-manager-not-initialized"
	}
	
	return manager.GenerateSimpleToken()
}

// GetCSRFConfig 获取CSRF配置（全局方法）
//
// 返回当前全局CSRF管理器的配置。
//
// 返回：
//   - *security.CSRFConfig: CSRF配置对象的副本
//   - error: 错误信息
//
// 示例：
//
//	config, err := mvc.GetCSRFConfig()
//	if err != nil {
//	    log.Printf("获取CSRF配置失败: %v", err)
//	    return
//	}
//	log.Printf("CSRF cookie name: %s", config.CookieName)
func GetCSRFConfig() (*security.CSRFConfig, error) {
	manager := GetCSRFManager()
	if manager == nil {
		return nil, fmt.Errorf("CSRF manager is not initialized")
	}
	
	return manager.GetConfig(), nil
}

// UpdateCSRFConfig 更新CSRF配置（全局方法）
//
// 更新全局CSRF管理器的配置。
//
// 参数：
//   - config: 新的CSRF配置
//
// 示例：
//
//	config := security.DefaultCSRFConfig()
//	config.Secret = "my-secret-key"
//	config.ExpireTime = 7200 // 2小时
//	mvc.UpdateCSRFConfig(config)
func UpdateCSRFConfig(config *security.CSRFConfig) {
	manager := GetCSRFManager()
	if manager != nil {
		manager.UpdateConfig(config)
	}
}

// SetCSRFSkipCheckFunc 设置CSRF跳过检查的函数（全局方法）
//
// 设置一个函数来决定在什么情况下跳过CSRF检查。
//
// 参数：
//   - fn: 跳过检查的函数，接收userID和clientIP，返回是否跳过检查
//
// 示例：
//
//	mvc.SetCSRFSkipCheckFunc(func(userID, clientIP string) bool {
//	    // 对于管理员用户跳过CSRF检查
//	    return strings.HasPrefix(userID, "admin_")
//	})
func SetCSRFSkipCheckFunc(fn func(userID, clientIP string) bool) {
	manager := GetCSRFManager()
	if manager != nil {
		manager.SetSkipCheckFunc(fn)
	}
}

// ============= 便捷方法 =============

// IsCSRFProtectionEnabled 检查CSRF保护是否启用
//
// 返回：
//   - bool: CSRF保护是否启用
func IsCSRFProtectionEnabled() bool {
	manager := GetCSRFManager()
	return manager != nil
}

// GetCSRFTokenName 获取CSRF token的表单字段名
//
// 返回：
//   - string: 表单字段名
func GetCSRFTokenName() string {
	config, err := GetCSRFConfig()
	if err != nil {
		return "csrf_token" // 默认名称
	}
	return config.FormFieldName
}

// GetCSRFCookieName 获取CSRF token的Cookie名
//
// 返回：
//   - string: Cookie名称
func GetCSRFCookieName() string {
	config, err := GetCSRFConfig()
	if err != nil {
		return "_csrf_token" // 默认名称
	}
	return config.CookieName
}

// GetCSRFHeaderName 获取CSRF token的请求头名
//
// 返回：
//   - string: 请求头名称
func GetCSRFHeaderName() string {
	config, err := GetCSRFConfig()
	if err != nil {
		return "X-CSRF-Token" // 默认名称
	}
	return config.HeaderName
}