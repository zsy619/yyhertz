package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthConfig(t *testing.T) {
	t.Run("创建认证配置", func(t *testing.T) {
		config := NewAuthConfig()
		
		assert.NotNil(t, config)
		assert.Equal(t, "/admin/login", config.LoginURI)
		assert.Equal(t, "/mobile/login", config.LoginMobileURI)
		assert.Equal(t, "/admin/logout", config.LogoutURI)
	})

	t.Run("初始化认证配置", func(t *testing.T) {
		// 设置测试配置
		Set("cas.enabled", "true")  
		Set("cas.url", "https://cas.example.com/")
		Set("appname", "Test App")
		Set("site.domain", "https://app.example.com")

		config := NewAuthConfig()
		config.InitAuthConfig()

		assert.True(t, config.CasEnabled)
		assert.Equal(t, "https://cas.example.com", config.CasHost)
		assert.Equal(t, "Test App", config.AppName)
		assert.Equal(t, "https://app.example.com", config.LocalDomain)
		
		// 验证构建的路径
		assert.Equal(t, "https://cas.example.com/cas/login", config.CasLoginPath)
		assert.Equal(t, "https://cas.example.com/cas/logout", config.CasLogoutPath)
		assert.Equal(t, "https://app.example.com/school/login", config.LoginPathOfSchool)
		assert.Equal(t, "https://app.example.com/admin/login", config.LoginPathOfAdmin)
	})

	t.Run("本地域名处理", func(t *testing.T) {
		config := NewAuthConfig()
		config.LocalDomain = "https://example.com"

		assert.Equal(t, "https://example.com", config.GetLocalDomain())
		assert.Equal(t, "https://example.com/", config.GetLocalDomainWithSlash())

		// 测试已有斜杠的情况
		config.LocalDomain = "https://example.com/"
		assert.Equal(t, "https://example.com/", config.GetLocalDomainWithSlash())
	})

	t.Run("CAS功能测试", func(t *testing.T) {
		config := NewAuthConfig()
		config.CasEnabled = true
		config.CasLoginPath = "https://cas.example.com/cas/login"
		config.CasLogoutPath = "https://cas.example.com/cas/logout"
		config.CasValidatePath = "https://cas.example.com/cas/validate"

		assert.True(t, config.IsCasEnabled())

		// 测试CAS登录URL
		loginURL := config.GetCasLoginURL("https://app.example.com/callback")
		assert.Equal(t, "https://cas.example.com/cas/login?service=https://app.example.com/callback", loginURL)

		// 测试CAS登出URL
		logoutURL := config.GetCasLogoutURL("https://app.example.com/")
		assert.Equal(t, "https://cas.example.com/cas/logout?service=https://app.example.com/", logoutURL)

		// 测试CAS验证URL
		validateURL := config.GetCasValidateURL("https://app.example.com/callback", "ST-123456")
		assert.Equal(t, "https://cas.example.com/cas/validate?service=https://app.example.com/callback&ticket=ST-123456", validateURL)
	})

	t.Run("CAS未启用时的行为", func(t *testing.T) {
		config := NewAuthConfig()
		config.CasEnabled = false
		
		assert.False(t, config.IsCasEnabled())
		assert.Equal(t, "", config.GetCasLoginURL("https://app.example.com/callback"))
		assert.Equal(t, "", config.GetCasLogoutURL("https://app.example.com/"))
		assert.Equal(t, "", config.GetCasValidateURL("https://app.example.com/callback", "ST-123456"))
	})
}

func TestGlobalAuthConfig(t *testing.T) {
	// 清理配置
	Clear()

	t.Run("全局认证配置", func(t *testing.T) {
		// 设置测试配置
		Set("appname", "Global Test App")
		Set("site.domain", "https://global.example.com")
		
		// 初始化全局配置
		InitAuth()
		
		config := GetAuthConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "Global Test App", config.AppName)
		assert.Equal(t, "https://global.example.com", config.LocalDomain)
	})

	t.Run("便捷函数测试", func(t *testing.T) {
		// 设置测试配置
		Set("appname", "Convenience Test")
		Set("site.domain", "https://convenience.example.com")
		
		InitAuth()
		
		assert.Equal(t, "Convenience Test", GetAppName())
		assert.Equal(t, "/admin/login", GetLoginURI())
		assert.Equal(t, "/admin/logout", GetLogoutURI())
		assert.Equal(t, "https://convenience.example.com", GetLocalDomain())
		assert.Equal(t, "https://convenience.example.com/", GetLocalDomainWithSlash())
	})

	t.Run("兼容性函数测试", func(t *testing.T) {
		Set("test.config", "test value")
		
		assert.Equal(t, "test value", C("test.config", "default"))
		assert.Equal(t, "default", C("nonexistent.config", "default"))
	})
}

func TestAuthConfigEdgeCases(t *testing.T) {
	t.Run("空配置处理", func(t *testing.T) {
		// 清理全局配置
		Clear()
		
		config := NewAuthConfig()
		config.CasHost = ""
		config.LocalDomain = ""
		
		config.InitAuthConfig()
		
		assert.Equal(t, "", config.CasLoginPath)
		assert.Equal(t, "", config.LoginPathOfSchool)
		assert.Equal(t, "", config.GetLocalDomainWithSlash())
	})

	t.Run("URL构建边界情况", func(t *testing.T) {
		config := NewAuthConfig()
		
		// CAS未启用时的URL构建
		config.CasEnabled = false
		assert.Equal(t, "", config.GetCasLoginURL("service"))
		
		// CAS路径为空时的URL构建
		config.CasEnabled = true
		config.CasLoginPath = ""
		assert.Equal(t, "", config.GetCasLoginURL("service"))
		
		// 登出URL没有service参数
		config.CasLogoutPath = "https://cas.example.com/logout"
		assert.Equal(t, "https://cas.example.com/logout", config.GetCasLogoutURL(""))
	})
}

// 基准测试
func BenchmarkAuthConfig(b *testing.B) {
	config := NewAuthConfig()
	config.CasEnabled = true
	config.CasLoginPath = "https://cas.example.com/cas/login"
	
	b.Run("GetCasLoginURL", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config.GetCasLoginURL("https://app.example.com/callback")
		}
	})
	
	b.Run("IsCasEnabled", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config.IsCasEnabled()
		}
	})
	
	b.Run("GetLocalDomainWithSlash", func(b *testing.B) {
		config.LocalDomain = "https://example.com"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config.GetLocalDomainWithSlash()
		}
	})
}