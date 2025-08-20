package mvc

import (
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/security"
)

// TestCSRFConfigurationFromSession 测试CSRF配置是否正确从session配置加载
func TestCSRFConfigurationFromSession(t *testing.T) {
	// 重置全局管理器确保干净的测试环境
	ResetGlobalManagers()
	
	// 初始化全局管理器
	InitializeGlobalManagers()
	
	// 获取CSRF管理器
	manager := GetCSRFManager()
	if manager == nil {
		t.Fatal("❌ CSRF管理器初始化失败")
	}
	
	// 获取配置
	config := manager.GetConfig()
	if config == nil {
		t.Fatal("❌ 无法获取CSRF配置")
	}
	
	// 验证配置值是否来自session配置的默认值
	t.Logf("✅ CSRF配置加载成功:")
	t.Logf("  - Secret: %s", maskSecret(config.Secret))
	t.Logf("  - TokenLength: %d", config.TokenLength)
	t.Logf("  - ExpireTime: %d", config.ExpireTime)
	t.Logf("  - CookieName: %s", config.CookieName)
	t.Logf("  - HeaderName: %s", config.HeaderName)
	t.Logf("  - FormFieldName: %s", config.FormFieldName)
	
	// 验证配置值是否符合预期（来自session_config.go的默认值）
	expectedValues := map[string]string{
		"Secret":        "yyhertz-csrf-secret-key-default",
		"CookieName":    "_csrf_token",
		"HeaderName":    "X-CSRF-Token",
		"FormFieldName": "csrf_token",
	}
	
	if config.Secret != expectedValues["Secret"] {
		t.Errorf("❌ Secret配置不正确，期望: %s，实际: %s", expectedValues["Secret"], config.Secret)
	}
	
	if config.CookieName != expectedValues["CookieName"] {
		t.Errorf("❌ CookieName配置不正确，期望: %s，实际: %s", expectedValues["CookieName"], config.CookieName)
	}
	
	if config.HeaderName != expectedValues["HeaderName"] {
		t.Errorf("❌ HeaderName配置不正确，期望: %s，实际: %s", expectedValues["HeaderName"], config.HeaderName)
	}
	
	if config.FormFieldName != expectedValues["FormFieldName"] {
		t.Errorf("❌ FormFieldName配置不正确，期望: %s，实际: %s", expectedValues["FormFieldName"], config.FormFieldName)
	}
	
	if config.TokenLength != 32 {
		t.Errorf("❌ TokenLength配置不正确，期望: %d，实际: %d", 32, config.TokenLength)
	}
	
	if config.ExpireTime != 3600 {
		t.Errorf("❌ ExpireTime配置不正确，期望: %d，实际: %d", 3600, config.ExpireTime)
	}
	
	t.Log("✅ 所有配置值验证通过")
}

// TestCSRFConfigLoadFunction 直接测试LoadCSRFConfig函数
func TestCSRFConfigLoadFunction(t *testing.T) {
	// 直接调用LoadCSRFConfig函数
	config := security.LoadCSRFConfig()
	if config == nil {
		t.Fatal("❌ LoadCSRFConfig返回nil")
	}
	
	t.Logf("✅ LoadCSRFConfig函数测试通过:")
	t.Logf("  - Secret: %s", maskSecret(config.Secret))
	t.Logf("  - TokenLength: %d", config.TokenLength)
	t.Logf("  - ExpireTime: %d", config.ExpireTime)
	t.Logf("  - CookieName: %s", config.CookieName)
	t.Logf("  - HeaderName: %s", config.HeaderName)
	t.Logf("  - FormFieldName: %s", config.FormFieldName)
	
	// 验证基本配置值
	if config.Secret == "" {
		t.Error("❌ Secret不应为空")
	}
	
	if config.TokenLength <= 0 {
		t.Error("❌ TokenLength应大于0")
	}
	
	if config.ExpireTime <= 0 {
		t.Error("❌ ExpireTime应大于0")
	}
	
	if config.CookieName == "" {
		t.Error("❌ CookieName不应为空")
	}
	
	if config.HeaderName == "" {
		t.Error("❌ HeaderName不应为空")
	}
	
	if config.FormFieldName == "" {
		t.Error("❌ FormFieldName不应为空")
	}
	
	t.Log("✅ LoadCSRFConfig函数所有验证通过")
}

// maskSecret 用于在日志中隐藏敏感信息
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}