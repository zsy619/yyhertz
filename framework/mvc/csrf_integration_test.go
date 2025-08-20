package mvc

import (
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc/security"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestCSRFIntegration 测试CSRF token与系统的集成
func TestCSRFIntegration(t *testing.T) {
	// 重置状态
	ResetGlobalManagers()
	
	// 测试初始化
	t.Run("TestInitialization", func(t *testing.T) {
		// 初始化全局管理器
		InitializeGlobalManagers()
		
		// 验证CSRF管理器已初始化
		manager := GetCSRFManager()
		if manager == nil {
			t.Fatal("CSRF管理器未初始化")
		}
		
		// 验证view包中的CSRF提供者已设置
		provider := view.GetCSRFTokenProvider()
		if provider == nil {
			t.Fatal("view包中的CSRF提供者未设置")
		}
		
		t.Log("✅ CSRF集成初始化成功")
	})
	
	// 测试全局API
	t.Run("TestGlobalAPI", func(t *testing.T) {
		// 测试生成token
		token, err := GenerateCSRFToken("test_user", "127.0.0.1")
		if err != nil {
			t.Fatalf("生成CSRF token失败: %v", err)
		}
		
		if token.Value == "" {
			t.Error("生成的CSRF token为空")
		}
		
		if token.UserID != "test_user" {
			t.Errorf("期望用户ID: test_user, 实际: %s", token.UserID)
		}
		
		// 测试验证token
		isValid, err := ValidateCSRFToken(token.Value, "test_user", "127.0.0.1")
		if err != nil {
			t.Fatalf("验证CSRF token失败: %v", err)
		}
		
		if !isValid {
			t.Error("CSRF token验证失败")
		}
		
		// 测试简单token生成
		simpleToken := GenerateSimpleCSRFToken()
		if simpleToken == "" {
			t.Error("生成的简单CSRF token为空")
		}
		
		t.Logf("✅ 全局API测试通过 - token: %s", simpleToken[:20]+"...")
	})
	
	// 测试与view包的集成
	t.Run("TestViewIntegration", func(t *testing.T) {
		// 创建模板引擎
		cfg := view.DefaultTemplateConfig()
		engine, err := view.NewTemplateEngine(cfg)
		if err != nil {
			t.Fatalf("创建模板引擎失败: %v", err)
		}
		
		// 测试模板函数
		testTemplate := `CSRF Token: {{csrf_token}}`
		tmpl, err := engine.CreateInlineTemplate("csrf_test", testTemplate)
		if err != nil {
			t.Fatalf("创建模板失败: %v", err)
		}
		
		result, err := engine.ExecuteTemplate(tmpl, nil)
		if err != nil {
			t.Fatalf("执行模板失败: %v", err)
		}
		
		if !strings.Contains(result, "CSRF Token:") {
			t.Error("模板结果不包含CSRF Token")
		}
		
		if strings.Contains(result, "csrf-token-placeholder") {
			t.Error("模板仍然使用占位符，集成可能失败")
		}
		
		t.Logf("✅ view包集成测试通过 - 结果: %s", result)
	})
	
	// 测试配置
	t.Run("TestConfiguration", func(t *testing.T) {
		// 测试获取配置
		config, err := GetCSRFConfig()
		if err != nil {
			t.Fatalf("获取CSRF配置失败: %v", err)
		}
		
		if config.Secret == "" {
			t.Error("CSRF配置中的secret为空")
		}
		
		if config.CookieName == "" {
			t.Error("CSRF配置中的cookie名称为空")
		}
		
		// 测试便捷方法
		tokenName := GetCSRFTokenName()
		if tokenName == "" {
			t.Error("获取的CSRF token名称为空")
		}
		
		cookieName := GetCSRFCookieName()
		if cookieName == "" {
			t.Error("获取的CSRF cookie名称为空")
		}
		
		headerName := GetCSRFHeaderName()
		if headerName == "" {
			t.Error("获取的CSRF header名称为空")
		}
		
		// 测试保护状态
		if !IsCSRFProtectionEnabled() {
			t.Error("CSRF保护应该已启用")
		}
		
		t.Logf("✅ 配置测试通过 - token名: %s, cookie名: %s, header名: %s", 
			tokenName, cookieName, headerName)
	})
	
	// 测试自定义配置
	t.Run("TestCustomConfiguration", func(t *testing.T) {
		// 创建自定义配置
		customConfig := &security.CSRFConfig{
			Secret:        "custom-secret-key",
			TokenLength:   64,
			ExpireTime:    7200,
			CookieName:    "_custom_csrf",
			HeaderName:    "X-Custom-CSRF-Token",
			FormFieldName: "custom_csrf_token",
		}
		
		// 更新配置
		UpdateCSRFConfig(customConfig)
		
		// 验证配置已更新
		config, err := GetCSRFConfig()
		if err != nil {
			t.Fatalf("获取更新后的配置失败: %v", err)
		}
		
		if config.Secret != "custom-secret-key" {
			t.Errorf("期望secret: custom-secret-key, 实际: %s", config.Secret)
		}
		
		if config.TokenLength != 64 {
			t.Errorf("期望token长度: 64, 实际: %d", config.TokenLength)
		}
		
		t.Log("✅ 自定义配置测试通过")
	})
	
	// 测试跳过检查功能
	t.Run("TestSkipCheckFunction", func(t *testing.T) {
		// 设置跳过检查函数
		SetCSRFSkipCheckFunc(func(userID, clientIP string) bool {
			return strings.HasPrefix(userID, "admin_")
		})
		
		// 测试管理员用户（应该跳过）
		token, err := GenerateCSRFToken("admin_test", "127.0.0.1")
		if err != nil {
			t.Fatalf("生成管理员CSRF token失败: %v", err)
		}
		
		isValid, err := ValidateCSRFToken(token.Value, "admin_test", "127.0.0.1")
		if err != nil {
			t.Fatalf("验证管理员CSRF token失败: %v", err)
		}
		
		if !isValid {
			t.Error("管理员CSRF token验证应该成功")
		}
		
		// 测试普通用户
		userToken, err := GenerateCSRFToken("user_test", "127.0.0.1")
		if err != nil {
			t.Fatalf("生成用户CSRF token失败: %v", err)
		}
		
		userValid, err := ValidateCSRFToken(userToken.Value, "user_test", "127.0.0.1")
		if err != nil {
			t.Fatalf("验证用户CSRF token失败: %v", err)
		}
		
		if !userValid {
			t.Error("用户CSRF token验证应该成功")
		}
		
		t.Log("✅ 跳过检查功能测试通过")
	})
}

// TestCSRFManagerStatus 测试CSRF管理器状态
func TestCSRFManagerStatus(t *testing.T) {
	// 重置并初始化
	ResetGlobalManagers()
	InitializeGlobalManagers()
	
	// 获取状态
	status := GetGlobalManagersStatus()
	
	// 验证CSRF状态
	if !status["csrf"] {
		t.Error("CSRF管理器状态应该为true")
	}
	
	if !status["all"] {
		t.Error("所有管理器状态应该为true")
	}
	
	t.Logf("✅ 管理器状态: %+v", status)
}

// BenchmarkCSRFTokenGeneration 基准测试CSRF token生成
func BenchmarkCSRFTokenGeneration(b *testing.B) {
	// 初始化
	ResetGlobalManagers()
	InitializeGlobalManagers()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := GenerateCSRFToken("benchmark_user", "127.0.0.1")
		if err != nil {
			b.Fatalf("生成CSRF token失败: %v", err)
		}
	}
}

// BenchmarkSimpleCSRFTokenGeneration 基准测试简单CSRF token生成
func BenchmarkSimpleCSRFTokenGeneration(b *testing.B) {
	// 初始化
	ResetGlobalManagers()
	InitializeGlobalManagers()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		token := GenerateSimpleCSRFToken()
		if token == "" {
			b.Fatal("生成的简单CSRF token为空")
		}
	}
}