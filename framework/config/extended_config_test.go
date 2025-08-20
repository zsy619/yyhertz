package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthConfig_ConfigManager(t *testing.T) {
	t.Run("获取认证配置", func(t *testing.T) {
		config, err := GetAuthConfig()
		require.NoError(t, err)
		require.NotNil(t, config)

		// 检查默认值
		assert.False(t, config.CAS.Enabled)
		assert.Equal(t, "https://cas.example.com", config.CAS.Host)
		assert.Equal(t, "/cas/login", config.CAS.LoginPath)
		assert.Equal(t, "3.0", config.CAS.Version)

		assert.Equal(t, "/login/school", config.LoginPaths.School.Path)
		assert.Equal(t, "/login/admin", config.LoginPaths.Admin.Path)
		assert.Equal(t, "/login", config.LoginPaths.Common.LoginURI)

		assert.Equal(t, "your-jwt-secret-key-change-me", config.JWT.Secret)
		assert.Equal(t, 24, config.JWT.TokenTTL)
		assert.Equal(t, "HS256", config.JWT.SigningMethod)

		assert.Equal(t, "YYHERTZ_SESSION", config.Session.Name)
		assert.Equal(t, 3600, config.Session.MaxAge)
		assert.True(t, config.Session.HttpOnly)

		assert.True(t, config.Authorization.Enable)
		assert.Equal(t, "user", config.Authorization.DefaultRole)
		assert.Contains(t, config.Authorization.AdminRoles, "admin")

		assert.Equal(t, 8, config.Security.PasswordPolicy.MinLength)
		assert.True(t, config.Security.PasswordPolicy.RequireUppercase)
		assert.Equal(t, 5, config.Security.LoginAttempts.MaxAttempts)

		assert.Equal(t, "YYHertz Auth", config.Application.Name)
		assert.Equal(t, "localhost", config.Application.LocalDomain)
		assert.True(t, config.Application.Debug)

		assert.True(t, config.Logging.Enable)
		assert.Equal(t, "info", config.Logging.Level)
		assert.Equal(t, "./logs/auth.log", config.Logging.LogFile)
	})

	t.Run("使用泛型便捷函数 - AuthConfig", func(t *testing.T) {
		casEnabled := GetConfigBool(AuthConfig{}, "cas.enabled")
		assert.False(t, casEnabled)

		casHost := GetConfigString(AuthConfig{}, "cas.host")
		assert.Equal(t, "https://cas.example.com", casHost)

		jwtTTL := GetConfigInt(AuthConfig{}, "jwt.token_ttl")
		assert.Equal(t, 24, jwtTTL)

		sessionMaxAge := GetConfigInt(AuthConfig{}, "session.max_age")
		assert.Equal(t, 3600, sessionMaxAge)

		authEnable := GetConfigBool(AuthConfig{}, "authorization.enable")
		assert.True(t, authEnable)

		passwordMinLength := GetConfigInt(AuthConfig{}, "security.password_policy.min_length")
		assert.Equal(t, 8, passwordMinLength)

		appName := GetConfigString(AuthConfig{}, "application.name")
		assert.Equal(t, "YYHertz Auth", appName)
	})

	t.Run("设置认证配置值", func(t *testing.T) {
		// 设置新值
		SetConfigValue(AuthConfig{}, "cas.enabled", true)
		SetConfigValue(AuthConfig{}, "cas.host", "https://cas.mycompany.com")
		SetConfigValue(AuthConfig{}, "jwt.token_ttl", 48)
		SetConfigValue(AuthConfig{}, "application.name", "My Auth Service")

		// 验证设置的值
		casEnabled := GetConfigBool(AuthConfig{}, "cas.enabled")
		assert.True(t, casEnabled)

		casHost := GetConfigString(AuthConfig{}, "cas.host")
		assert.Equal(t, "https://cas.mycompany.com", casHost)

		jwtTTL := GetConfigInt(AuthConfig{}, "jwt.token_ttl")
		assert.Equal(t, 48, jwtTTL)

		appName := GetConfigString(AuthConfig{}, "application.name")
		assert.Equal(t, "My Auth Service", appName)
	})

	t.Run("使用认证配置管理器", func(t *testing.T) {
		manager := GetAuthConfigManager()
		require.NotNil(t, manager)

		// 获取配置值
		casVersion := manager.GetString("cas.version")
		assert.Equal(t, "3.0", casVersion)

		loginPath := manager.GetString("login_paths.school.path")
		assert.Equal(t, "/login/school", loginPath)

		adminRoles := manager.GetStringSlice("authorization.admin_roles")
		assert.Contains(t, adminRoles, "admin")
		assert.Contains(t, adminRoles, "superadmin")

		// 设置配置值
		manager.Set("cas.enabled", true)
		manager.Set("application.environment", "production")

		// 验证设置的值
		casEnabled := manager.GetBool("cas.enabled")
		assert.True(t, casEnabled)

		environment := manager.GetString("application.environment")
		assert.Equal(t, "production", environment)
	})
}

func TestTemplateConfig_ConfigManager(t *testing.T) {
	t.Run("获取模板配置", func(t *testing.T) {
		config, err := GetTemplateConfig()
		require.NoError(t, err)
		require.NotNil(t, config)

		// 检查默认值
		assert.Equal(t, ".html", config.Extension)
		assert.Contains(t, config.ViewPaths, "./views")
		assert.Equal(t, "./views/layouts", config.LayoutPath)
		assert.True(t, config.EnableReload)
		assert.True(t, config.EnableCache)
		assert.Equal(t, "./views/components", config.ComponentPath)
		assert.Equal(t, "default", config.CurrentTheme)
		assert.NotNil(t, config.Themes)
		assert.Contains(t, config.Themes, "default")
	})

	t.Run("使用泛型便捷函数 - TemplateConfig", func(t *testing.T) {
		extension := GetConfigString(TemplateConfig{}, "extension")
		assert.Equal(t, ".html", extension)

		layoutPath := GetConfigString(TemplateConfig{}, "layout_path")
		assert.Equal(t, "./views/layouts", layoutPath)

		reloadEnabled := GetConfigBool(TemplateConfig{}, "enable_reload")
		assert.True(t, reloadEnabled)

		cacheEnabled := GetConfigBool(TemplateConfig{}, "enable_cache")
		assert.True(t, cacheEnabled)

		componentPath := GetConfigString(TemplateConfig{}, "component_path")
		assert.Equal(t, "./views/components", componentPath)

		currentTheme := GetConfigString(TemplateConfig{}, "current_theme")
		assert.Equal(t, "default", currentTheme)
	})

	t.Run("设置模板配置值", func(t *testing.T) {
		// 设置新值
		SetConfigValue(TemplateConfig{}, "extension", ".pug")
		SetConfigValue(TemplateConfig{}, "layout_path", "./templates/layouts")
		SetConfigValue(TemplateConfig{}, "enable_cache", false)
		SetConfigValue(TemplateConfig{}, "current_theme", "custom")

		// 验证设置的值
		extension := GetConfigString(TemplateConfig{}, "extension")
		assert.Equal(t, ".pug", extension)

		layoutPath := GetConfigString(TemplateConfig{}, "layout_path")
		assert.Equal(t, "./templates/layouts", layoutPath)

		cacheEnabled := GetConfigBool(TemplateConfig{}, "enable_cache")
		assert.False(t, cacheEnabled)

		theme := GetConfigString(TemplateConfig{}, "current_theme")
		assert.Equal(t, "custom", theme)
	})

	t.Run("使用模板配置管理器", func(t *testing.T) {
		manager := GetTemplateConfigManager()
		require.NotNil(t, manager)

		// 获取配置值
		extension := manager.GetString("engine.extension")
		assert.Equal(t, ".html", extension)

		htmlEscape := manager.GetBool("render.html_escape")
		assert.True(t, htmlEscape)

		cacheType := manager.GetString("cache.type")
		assert.Equal(t, "memory", cacheType)

		staticCompress := manager.GetBool("static.compress")
		assert.True(t, staticCompress)

		// 设置配置值
		manager.Set("extension", ".hbs")
		manager.Set("enable_cache", false)
		manager.Set("current_theme", "custom")

		// 验证设置的值
		extension = manager.GetString("extension")
		assert.Equal(t, ".hbs", extension)

		cacheEnabled := manager.GetBool("enable_cache")
		assert.False(t, cacheEnabled)

		theme := manager.GetString("current_theme")
		assert.Equal(t, "custom", theme)
	})
}

func TestExtendedConfig_SingletonBehavior(t *testing.T) {
	t.Run("扩展配置管理器单例行为", func(t *testing.T) {
		// 认证配置管理器
		authManager1 := GetAuthConfigManager()
		authManager2 := GetAuthConfigManager()
		assert.Equal(t, authManager1, authManager2)

		// 模板配置管理器
		templateManager1 := GetTemplateConfigManager()
		templateManager2 := GetTemplateConfigManager()
		assert.Equal(t, templateManager1, templateManager2)

		// 应用配置管理器
		appManager1 := GetAppConfigManager()
		appManager2 := GetAppConfigManager()
		assert.Equal(t, appManager1, appManager2)

		// 不同类型的管理器应该是不同的实例
		assert.NotEqual(t, authManager1, templateManager1)
		assert.NotEqual(t, authManager1, appManager1)
		assert.NotEqual(t, templateManager1, appManager1)
	})
}

func TestExtendedConfig_CrossConfigAccess(t *testing.T) {
	t.Run("跨配置类型访问", func(t *testing.T) {
		// 设置不同配置类型的值
		SetConfigValue(AppConfig{}, "app.name", "MyApp")
		SetConfigValue(AuthConfig{}, "application.name", "MyAuthService")
		SetConfigValue(TemplateConfig{}, "extension", ".vue")

		// 验证每个配置类型都维持独立的值
		appName := GetConfigString(AppConfig{}, "app.name")
		assert.Equal(t, "MyApp", appName)

		authAppName := GetConfigString(AuthConfig{}, "application.name")
		assert.Equal(t, "MyAuthService", authAppName)

		templateType := GetConfigString(TemplateConfig{}, "engine.type")
		assert.Equal(t, "vue", templateType)

		// 确保配置之间没有相互影响
		assert.NotEqual(t, appName, authAppName)
	})
}

func TestExtendedConfig_ComplexDataTypes(t *testing.T) {
	t.Run("复杂数据类型处理", func(t *testing.T) {
		authManager := GetAuthConfigManager()

		// 测试字符串数组
		adminRoles := authManager.GetStringSlice("authorization.admin_roles")
		assert.Contains(t, adminRoles, "admin")
		assert.Contains(t, adminRoles, "superadmin")

		guestPaths := authManager.GetStringSlice("authorization.guest_paths")
		assert.Contains(t, guestPaths, "/")
		assert.Contains(t, guestPaths, "/login")

		templateManager := GetTemplateConfigManager()

		// 测试字符串数组
		indexFiles := templateManager.GetStringSlice("static.index")
		assert.Contains(t, indexFiles, "index.html")
		assert.Contains(t, indexFiles, "index.htm")

		watchPatterns := templateManager.GetStringSlice("development.watch_patterns")
		assert.Contains(t, watchPatterns, "*.html")
		assert.Contains(t, watchPatterns, "*.css")
		assert.Contains(t, watchPatterns, "*.js")
	})
}
