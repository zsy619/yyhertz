package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager_AppConfig(t *testing.T) {
	t.Run("获取应用配置", func(t *testing.T) {
		config, err := GetAppConfig()
		require.NoError(t, err)
		require.NotNil(t, config)

		// 检查默认值
		assert.Equal(t, "YYHertz", config.App.Name)
		assert.Equal(t, 8888, config.App.Port)
		// 注释掉数据库配置测试，因为已从AppConfig中移除
		// assert.Equal(t, "mysql", config.Database.Driver)
		// assert.Equal(t, "127.0.0.1", config.Database.Host)
	})

	t.Run("使用泛型便捷函数", func(t *testing.T) {
		appName := GetConfigString(AppConfig{}, "app.name")
		assert.Equal(t, "YYHertz", appName)

		appPort := GetConfigInt(AppConfig{}, "app.port")
		assert.Equal(t, 8888, appPort)

		appDebug := GetConfigBool(AppConfig{}, "app.debug")
		assert.True(t, appDebug)
	})

	t.Run("设置配置值", func(t *testing.T) {
		// 设置新值
		SetConfigValue(AppConfig{}, "app.name", "TestApp")
		SetConfigValue(AppConfig{}, "app.port", 9999)

		// 验证设置的值
		appName := GetConfigString(AppConfig{}, "app.name")
		assert.Equal(t, "TestApp", appName)

		appPort := GetConfigInt(AppConfig{}, "app.port")
		assert.Equal(t, 9999, appPort)
	})
}

func TestConfigManager_TemplateConfig(t *testing.T) {
	t.Run("获取模板配置", func(t *testing.T) {
		config, err := GetTemplateConfig()
		require.NoError(t, err)
		require.NotNil(t, config)

		// 检查默认值
		assert.Equal(t, ".html", config.Paths.Extension)
		assert.Contains(t, config.Paths.ViewPaths, "./views")
		assert.Equal(t, "./views/layouts", config.Paths.LayoutPath)
		assert.True(t, config.Cache.EnableCache)
	})

	t.Run("使用泛型便捷函数", func(t *testing.T) {
		extension := GetConfigString(TemplateConfig{}, "extension")
		assert.Equal(t, ".html", extension)

		layoutPath := GetConfigString(TemplateConfig{}, "layout_path")
		assert.Equal(t, "./views/layouts", layoutPath)

		cacheEnabled := GetConfigBool(TemplateConfig{}, "enable_cache")
		assert.True(t, cacheEnabled)
	})

	t.Run("设置模板配置值", func(t *testing.T) {
		// 设置新值
		SetConfigValue(TemplateConfig{}, "extension", ".pug")
		SetConfigValue(TemplateConfig{}, "enable_cache", false)

		// 验证设置的值
		extension := GetConfigString(TemplateConfig{}, "extension")
		assert.Equal(t, ".pug", extension)

		cacheEnabled := GetConfigBool(TemplateConfig{}, "enable_cache")
		assert.False(t, cacheEnabled)
	})
}

func TestConfigManager_SingletonBehavior(t *testing.T) {
	t.Run("配置管理器单例行为", func(t *testing.T) {
		manager1 := GetAppConfigManager()
		manager2 := GetAppConfigManager()

		// 应该是同一个实例
		assert.Equal(t, manager1, manager2)

		templateManager1 := GetTemplateConfigManager()
		templateManager2 := GetTemplateConfigManager()

		// 应该是同一个实例
		assert.Equal(t, templateManager1, templateManager2)

		// 不同类型的管理器应该是不同的实例
		assert.NotEqual(t, manager1, templateManager1)
	})
}

// 自定义配置示例
type TestConfig struct {
	Server struct {
		Host string `mapstructure:"host" yaml:"host"`
		Port int    `mapstructure:"port" yaml:"port"`
	} `mapstructure:"server" yaml:"server"`

	Features struct {
		EnableCache bool `mapstructure:"enable_cache" yaml:"enable_cache"`
		EnableLog   bool `mapstructure:"enable_log" yaml:"enable_log"`
	} `mapstructure:"features" yaml:"features"`
}

func (c TestConfig) GetConfigName() string {
	return "test"
}

func (c TestConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("features.enable_cache", true)
	v.SetDefault("features.enable_log", false)
}

func (c TestConfig) GenerateDefaultContent() string {
	return `server:
  host: "localhost"
  port: 8080

features:
  enable_cache: true
  enable_log: false
`
}

func TestConfigManager_CustomConfig(t *testing.T) {
	t.Run("自定义配置测试", func(t *testing.T) {
		// 获取自定义配置管理器
		manager := GetViperConfigManager(TestConfig{})
		require.NotNil(t, manager)

		// 获取配置值
		host := manager.GetString("server.host")
		assert.Equal(t, "localhost", host)

		port := manager.GetInt("server.port")
		assert.Equal(t, 8080, port)

		cacheEnabled := manager.GetBool("features.enable_cache")
		assert.True(t, cacheEnabled)

		// 设置新值
		manager.Set("server.port", 9000)
		manager.Set("features.enable_log", true)

		// 验证设置的值
		newPort := manager.GetInt("server.port")
		assert.Equal(t, 9000, newPort)

		logEnabled := manager.GetBool("features.enable_log")
		assert.True(t, logEnabled)

		// 获取完整配置
		config, err := manager.GetConfig()
		require.NoError(t, err)
		assert.Equal(t, "localhost", config.Server.Host)
		assert.Equal(t, 9000, config.Server.Port)
		assert.True(t, config.Features.EnableCache)
		assert.True(t, config.Features.EnableLog)
	})
}
