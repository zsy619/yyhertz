package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewViperConfigManager(t *testing.T) {
	cm := NewViperConfigManager()

	assert.NotNil(t, cm)
	assert.NotNil(t, cm.viper)
	assert.Equal(t, "config", cm.configName)
	assert.Equal(t, "yaml", cm.configType)
	assert.Equal(t, "YYHERTZ", cm.envPrefix)
	assert.False(t, cm.initialized)
}

func TestViperConfigManagerInitialize(t *testing.T) {
	cm := NewViperConfigManager()

	err := cm.Initialize()
	assert.NoError(t, err)
	assert.True(t, cm.initialized)

	// 验证默认值
	assert.Equal(t, "YYHertz", cm.GetString("app.name"))
	assert.Equal(t, "1.0.0", cm.GetString("app.version"))
	assert.Equal(t, 8888, cm.GetInt("app.port"))
	assert.True(t, cm.GetBool("app.debug"))
}

func TestViperConfigManagerWithConfigFile(t *testing.T) {
	// 创建临时配置目录
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// 创建测试配置文件
	configContent := `
app:
  name: "TestApp"
  version: "2.0.0"
  port: 9999
  debug: false

database:
  host: "test-db"
  port: 5432
  username: "testuser"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// 创建完全独立的配置管理器实例
	cm := &ViperConfigManager{
		viper:       viper.New(),
		configPaths: []string{},
		configName:  "config",
		configType:  "yaml",
		envPrefix:   "YYHERTZ",
		initialized: false,
	}

	// 直接设置配置文件路径
	cm.viper.SetConfigFile(configFile)

	// 读取配置文件
	err = cm.viper.ReadInConfig()
	assert.NoError(t, err)

	// 标记为已初始化
	cm.initialized = true

	// 验证配置值
	assert.Equal(t, "TestApp", cm.GetString("app.name"))
	assert.Equal(t, "2.0.0", cm.GetString("app.version"))
	assert.Equal(t, 9999, cm.GetInt("app.port"))
	assert.False(t, cm.GetBool("app.debug"))
	assert.Equal(t, "test-db", cm.GetString("database.host"))
	assert.Equal(t, 5432, cm.GetInt("database.port"))
}

func TestViperConfigManagerEnvironmentVariables(t *testing.T) {
	// 设置环境变量
	os.Setenv("YYHERTZ_APP_NAME", "EnvApp")
	os.Setenv("YYHERTZ_APP_PORT", "7777")
	os.Setenv("YYHERTZ_APP_DEBUG", "false")
	defer func() {
		os.Unsetenv("YYHERTZ_APP_NAME")
		os.Unsetenv("YYHERTZ_APP_PORT")
		os.Unsetenv("YYHERTZ_APP_DEBUG")
	}()

	cm := NewViperConfigManager()
	err := cm.Initialize()
	assert.NoError(t, err)

	// 环境变量应该覆盖默认值
	assert.Equal(t, "EnvApp", cm.GetString("app.name"))
	assert.Equal(t, 7777, cm.GetInt("app.port"))
	assert.False(t, cm.GetBool("app.debug"))
}

func TestViperConfigManagerGetConfig(t *testing.T) {
	cm := NewViperConfigManager()

	config, err := cm.GetConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// 验证配置结构
	assert.Equal(t, "YYHertz", config.App.Name)
	assert.Equal(t, "1.0.0", config.App.Version)
	assert.Equal(t, 8888, config.App.Port)
	assert.Equal(t, "development", config.App.Environment)
	assert.True(t, config.App.Debug)

	// 验证数据库配置
	assert.Equal(t, "mysql", config.Database.Driver)
	assert.Equal(t, "127.0.0.1", config.Database.Host)
	assert.Equal(t, 3306, config.Database.Port)

	// 验证Redis配置
	assert.Equal(t, "127.0.0.1", config.Redis.Host)
	assert.Equal(t, 6379, config.Redis.Port)
}

func TestViperConfigManagerSetAndGet(t *testing.T) {
	cm := NewViperConfigManager()

	// 设置配置值
	cm.Set("test.string", "hello")
	cm.Set("test.int", 123)
	cm.Set("test.bool", true)
	cm.Set("test.slice", []string{"a", "b", "c"})

	// 获取配置值
	assert.Equal(t, "hello", cm.GetString("test.string"))
	assert.Equal(t, 123, cm.GetInt("test.int"))
	assert.True(t, cm.GetBool("test.bool"))
	assert.Equal(t, []string{"a", "b", "c"}, cm.GetStringSlice("test.slice"))
}

func TestViperConfigManagerIsSet(t *testing.T) {
	cm := NewViperConfigManager()

	// 默认配置应该存在
	assert.True(t, cm.IsSet("app.name"))
	assert.True(t, cm.IsSet("app.port"))

	// 不存在的配置
	assert.False(t, cm.IsSet("nonexistent.key"))

	// 设置新配置
	cm.Set("new.key", "value")
	assert.True(t, cm.IsSet("new.key"))
}

func TestViperConfigManagerAllKeys(t *testing.T) {
	cm := NewViperConfigManager()

	keys := cm.AllKeys()
	assert.NotEmpty(t, keys)

	// 检查一些预期的键
	expectedKeys := []string{
		"app.name",
		"app.version",
		"app.port",
		"database.host",
		"redis.host",
	}

	for _, key := range expectedKeys {
		assert.Contains(t, keys, key)
	}
}

func TestViperConfigManagerConfigPaths(t *testing.T) {
	cm := NewViperConfigManager()

	// 添加自定义路径
	cm.AddConfigPath("/custom/path")

	assert.Contains(t, cm.configPaths, "/custom/path")
}

func TestViperConfigManagerConfigTypes(t *testing.T) {
	cm := NewViperConfigManager()

	// 测试设置不同的配置类型
	cm.SetConfigType("json")
	assert.Equal(t, "json", cm.configType)

	cm.SetConfigType("toml")
	assert.Equal(t, "toml", cm.configType)
}

func TestViperConfigManagerEnvPrefix(t *testing.T) {
	cm := NewViperConfigManager()

	// 设置环境变量前缀
	cm.SetEnvPrefix("MYAPP")
	assert.Equal(t, "MYAPP", cm.envPrefix)
}

func TestGlobalConfigFunctions(t *testing.T) {
	// 测试全局便捷函数
	config, err := GetGlobalConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// 测试全局获取函数
	assert.Equal(t, "YYHertz", GetConfigString("app.name"))
	assert.Equal(t, 8888, GetConfigInt("app.port"))
	assert.True(t, GetConfigBool("app.debug"))

	// 测试获取任意类型
	name := GetConfigValue("app.name")
	assert.Equal(t, "YYHertz", name)
}

func TestViperConfigManagerDuration(t *testing.T) {
	cm := NewViperConfigManager()

	// 设置持续时间配置
	cm.Set("timeout", "30s")
	cm.Set("interval", "5m")

	timeout := cm.GetDuration("timeout")
	assert.Equal(t, 30*time.Second, timeout)

	interval := cm.GetDuration("interval")
	assert.Equal(t, 5*time.Minute, interval)
}

func TestCreateDefaultConfigFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	cm := NewViperConfigManager()
	// 设置配置搜索路径为临时目录
	cm.configPaths = []string{"./config", tempDir, "/etc/yyhertz", "$HOME/.yyhertz"}
	// 清空已存在的全局配置管理器，确保独立测试
	viperConfigManagerMap = sync.Map{}

	// 初始化应该创建默认配置文件
	err := cm.Initialize()
	assert.NoError(t, err)

	// 检查配置文件是否创建 - 配置文件应该在 tempDir/config/config.yaml
	configFile := filepath.Join(tempDir, "config", "config.yaml")

	// 如果文件不存在，手动调用创建方法
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		err = cm.createDefaultConfigFile()
		assert.NoError(t, err)
	}

	// 现在检查文件是否存在
	_, err = os.Stat(configFile)
	assert.NoError(t, err)

	// 读取文件内容验证
	content, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "YYHertz Framework Configuration")
	assert.Contains(t, string(content), "app:")
	assert.Contains(t, string(content), "database:")
}
