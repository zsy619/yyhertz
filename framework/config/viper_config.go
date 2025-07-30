package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const DefaultConfigName = "config"

// ViperConfigManager Viper配置管理器
type ViperConfigManager struct {
	viper       *viper.Viper
	configPaths []string
	configName  string
	configType  string
	envPrefix   string
	initialized bool
}

// AppConfig 应用程序主配置结构
type AppConfig struct {
	// 应用基础配置
	App struct {
		Name        string `mapstructure:"name" yaml:"name" json:"name"`
		Version     string `mapstructure:"version" yaml:"version" json:"version"`
		Environment string `mapstructure:"environment" yaml:"environment" json:"environment"` // dev, test, prod
		Debug       bool   `mapstructure:"debug" yaml:"debug" json:"debug"`
		Port        int    `mapstructure:"port" yaml:"port" json:"port"`
		Host        string `mapstructure:"host" yaml:"host" json:"host"`
		Timezone    string `mapstructure:"timezone" yaml:"timezone" json:"timezone"`
	} `mapstructure:"app" yaml:"app" json:"app"`

	// 数据库配置
	Database struct {
		Driver   string `mapstructure:"driver" yaml:"driver" json:"driver"`
		Host     string `mapstructure:"host" yaml:"host" json:"host"`
		Port     int    `mapstructure:"port" yaml:"port" json:"port"`
		Username string `mapstructure:"username" yaml:"username" json:"username"`
		Password string `mapstructure:"password" yaml:"password" json:"password"`
		Database string `mapstructure:"database" yaml:"database" json:"database"`
		Charset  string `mapstructure:"charset" yaml:"charset" json:"charset"`
		MaxIdle  int    `mapstructure:"max_idle" yaml:"max_idle" json:"max_idle"`
		MaxOpen  int    `mapstructure:"max_open" yaml:"max_open" json:"max_open"`
		MaxLife  int    `mapstructure:"max_life" yaml:"max_life" json:"max_life"` // 秒
		SSLMode  string `mapstructure:"ssl_mode" yaml:"ssl_mode" json:"ssl_mode"`
	} `mapstructure:"database" yaml:"database" json:"database"`

	// Redis配置
	Redis struct {
		Host        string `mapstructure:"host" yaml:"host" json:"host"`
		Port        int    `mapstructure:"port" yaml:"port" json:"port"`
		Password    string `mapstructure:"password" yaml:"password" json:"password"`
		Database    int    `mapstructure:"database" yaml:"database" json:"database"`
		MaxRetries  int    `mapstructure:"max_retries" yaml:"max_retries" json:"max_retries"`
		PoolSize    int    `mapstructure:"pool_size" yaml:"pool_size" json:"pool_size"`
		MinIdle     int    `mapstructure:"min_idle" yaml:"min_idle" json:"min_idle"`
		DialTimeout int    `mapstructure:"dial_timeout" yaml:"dial_timeout" json:"dial_timeout"` // 秒
		ReadTimeout int    `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"` // 秒
	} `mapstructure:"redis" yaml:"redis" json:"redis"`

	// 日志配置
	Log LogConfig `mapstructure:"log" yaml:"log" json:"log"`

	// TLS配置
	TLS TLSServerConfig `mapstructure:"tls" yaml:"tls" json:"tls"`

	// 中间件配置
	Middleware struct {
		// CORS配置
		CORS struct {
			Enable           bool     `mapstructure:"enable" yaml:"enable" json:"enable"`
			AllowOrigins     []string `mapstructure:"allow_origins" yaml:"allow_origins" json:"allow_origins"`
			AllowMethods     []string `mapstructure:"allow_methods" yaml:"allow_methods" json:"allow_methods"`
			AllowHeaders     []string `mapstructure:"allow_headers" yaml:"allow_headers" json:"allow_headers"`
			ExposeHeaders    []string `mapstructure:"expose_headers" yaml:"expose_headers" json:"expose_headers"`
			AllowCredentials bool     `mapstructure:"allow_credentials" yaml:"allow_credentials" json:"allow_credentials"`
			MaxAge           int      `mapstructure:"max_age" yaml:"max_age" json:"max_age"`
		} `mapstructure:"cors" yaml:"cors" json:"cors"`

		// 限流配置
		RateLimit struct {
			Enable   bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			Rate     int    `mapstructure:"rate" yaml:"rate" json:"rate"`             // 请求/秒
			Burst    int    `mapstructure:"burst" yaml:"burst" json:"burst"`          // 突发容量
			Strategy string `mapstructure:"strategy" yaml:"strategy" json:"strategy"` // token_bucket, sliding_window
		} `mapstructure:"rate_limit" yaml:"rate_limit" json:"rate_limit"`

		// 认证配置
		Auth struct {
			Enable     bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			JWTSecret  string `mapstructure:"jwt_secret" yaml:"jwt_secret" json:"jwt_secret"`
			TokenTTL   int    `mapstructure:"token_ttl" yaml:"token_ttl" json:"token_ttl"`       // 小时
			RefreshTTL int    `mapstructure:"refresh_ttl" yaml:"refresh_ttl" json:"refresh_ttl"` // 小时
		} `mapstructure:"auth" yaml:"auth" json:"auth"`
	} `mapstructure:"middleware" yaml:"middleware" json:"middleware"`

	// 外部服务配置
	Services struct {
		// Email服务
		Email struct {
			Provider string `mapstructure:"provider" yaml:"provider" json:"provider"` // smtp, sendgrid, ses
			Host     string `mapstructure:"host" yaml:"host" json:"host"`
			Port     int    `mapstructure:"port" yaml:"port" json:"port"`
			Username string `mapstructure:"username" yaml:"username" json:"username"`
			Password string `mapstructure:"password" yaml:"password" json:"password"`
			From     string `mapstructure:"from" yaml:"from" json:"from"`
		} `mapstructure:"email" yaml:"email" json:"email"`

		// 文件存储
		Storage struct {
			Provider  string `mapstructure:"provider" yaml:"provider" json:"provider"` // local, s3, oss
			LocalPath string `mapstructure:"local_path" yaml:"local_path" json:"local_path"`
			Bucket    string `mapstructure:"bucket" yaml:"bucket" json:"bucket"`
			Region    string `mapstructure:"region" yaml:"region" json:"region"`
			AccessKey string `mapstructure:"access_key" yaml:"access_key" json:"access_key"`
			SecretKey string `mapstructure:"secret_key" yaml:"secret_key" json:"secret_key"`
			CDNDomain string `mapstructure:"cdn_domain" yaml:"cdn_domain" json:"cdn_domain"`
		} `mapstructure:"storage" yaml:"storage" json:"storage"`
	} `mapstructure:"services" yaml:"services" json:"services"`

	// 监控配置
	Monitor struct {
		Enable   bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
		Endpoint string `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
		Interval int    `mapstructure:"interval" yaml:"interval" json:"interval"` // 秒
		Timeout  int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`    // 秒
	} `mapstructure:"monitor" yaml:"monitor" json:"monitor"`
}

// DefaultViperConfigManager 全局配置管理器实例
var defaultViperConfigManager *ViperConfigManager

var defaultAppConfig *AppConfig

// 全局配置管理器map实例
var viperConfigManagerMap sync.Map

// NewViperConfigManager 创建新的配置管理器
func NewViperConfigManager() *ViperConfigManager {
	return &ViperConfigManager{
		viper:       viper.New(),
		configPaths: []string{".", "./config", "/etc/yyhertz", "$HOME/.yyhertz"},
		configName:  DefaultConfigName,
		configType:  "yaml",
		envPrefix:   "YYHERTZ",
		initialized: false,
	}
}

// 根据configName创建新的配置管理器
func NewViperConfigManagerWithName(name string) *ViperConfigManager {
	return &ViperConfigManager{
		viper:       viper.New(),
		configPaths: []string{".", "./config", "/etc/yyhertz", "$HOME/.yyhertz"},
		configName:  name,
		configType:  "yaml",
		envPrefix:   "YYHERTZ",
		initialized: false,
	}
}

// GetViperConfigManager 获取全局配置管理器实例
func GetViperConfigManager() *ViperConfigManager {
	return GetViperConfigManagerWithName("config")
}

// 根据 configName获取配置管理器实例
func GetViperConfigManagerWithName(name string) *ViperConfigManager {
	if value, ok := viperConfigManagerMap.Load(name); ok {
		return value.(*ViperConfigManager)
	}

	manager := NewViperConfigManagerWithName(name)
	manager.Initialize()
	viperConfigManagerMap.Store(name, manager)
	return manager
}

// Initialize 初始化配置管理器
func (cm *ViperConfigManager) Initialize() error {
	if cm.initialized {
		return nil
	}

	// 设置配置文件名和类型
	cm.viper.SetConfigName(cm.configName)
	cm.viper.SetConfigType(cm.configType)

	// 添加配置文件搜索路径
	for _, path := range cm.configPaths {
		cm.viper.AddConfigPath(path)
	}

	// 设置环境变量前缀
	cm.viper.SetEnvPrefix(cm.envPrefix)
	cm.viper.AutomaticEnv()
	cm.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	cm.setDefaults()

	// 尝试读取配置文件
	if err := cm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置并创建示例配置文件
			GetGlobalLogger().WithFields(map[string]any{
				"config_name": cm.configName,
				"config_type": cm.configType,
				"paths":       cm.configPaths,
			}).Warn("配置文件未找到，使用默认配置" + cm.configName + "." + cm.configPaths[0] + "，并尝试创建示例配置文件")

			if err := cm.createDefaultConfigFile(); err != nil {
				GetGlobalLogger().WithFields(map[string]any{
					"error": err.Error(),
				}).Error("创建默认配置文件失败")
			}
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else {
		GetGlobalLogger().WithFields(map[string]any{
			"config_file": cm.viper.ConfigFileUsed(),
		}).Info("配置文件加载成功")
	}

	cm.initialized = true
	return nil
}

// setDefaults 设置默认配置值
func (cm *ViperConfigManager) setDefaults() {
	// 应用默认配置
	cm.viper.SetDefault("app.name", "YYHertz")
	cm.viper.SetDefault("app.version", "1.0.0")
	cm.viper.SetDefault("app.environment", "development")
	cm.viper.SetDefault("app.debug", true)
	cm.viper.SetDefault("app.port", 8888)
	cm.viper.SetDefault("app.host", "0.0.0.0")
	cm.viper.SetDefault("app.timezone", "Asia/Shanghai")

	// 数据库默认配置
	cm.viper.SetDefault("database.driver", "mysql")
	cm.viper.SetDefault("database.host", "127.0.0.1")
	cm.viper.SetDefault("database.port", 3306)
	cm.viper.SetDefault("database.username", "root")
	cm.viper.SetDefault("database.password", "")
	cm.viper.SetDefault("database.database", "yyhertz")
	cm.viper.SetDefault("database.charset", "utf8mb4")
	cm.viper.SetDefault("database.max_idle", 10)
	cm.viper.SetDefault("database.max_open", 100)
	cm.viper.SetDefault("database.max_life", 3600)
	cm.viper.SetDefault("database.ssl_mode", "disable")

	// Redis默认配置
	cm.viper.SetDefault("redis.host", "127.0.0.1")
	cm.viper.SetDefault("redis.port", 6379)
	cm.viper.SetDefault("redis.password", "")
	cm.viper.SetDefault("redis.database", 0)
	cm.viper.SetDefault("redis.max_retries", 3)
	cm.viper.SetDefault("redis.pool_size", 10)
	cm.viper.SetDefault("redis.min_idle", 2)
	cm.viper.SetDefault("redis.dial_timeout", 5)
	cm.viper.SetDefault("redis.read_timeout", 3)

	// 日志默认配置
	cm.viper.SetDefault("log.level", "info")
	cm.viper.SetDefault("log.format", "json")
	cm.viper.SetDefault("log.enable_console", true)
	cm.viper.SetDefault("log.enable_file", false)
	cm.viper.SetDefault("log.file_path", "./logs/app.log")
	cm.viper.SetDefault("log.max_size", 100)
	cm.viper.SetDefault("log.max_age", 7)
	cm.viper.SetDefault("log.max_backups", 10)
	cm.viper.SetDefault("log.compress", true)

	// TLS默认配置
	cm.viper.SetDefault("tls.enable", false)
	cm.viper.SetDefault("tls.cert_file", "")
	cm.viper.SetDefault("tls.key_file", "")
	cm.viper.SetDefault("tls.min_version", "1.2")
	cm.viper.SetDefault("tls.max_version", "1.3")

	// 中间件默认配置
	cm.viper.SetDefault("middleware.cors.enable", true)
	cm.viper.SetDefault("middleware.cors.allow_origins", []string{"*"})
	cm.viper.SetDefault("middleware.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	cm.viper.SetDefault("middleware.cors.allow_headers", []string{"*"})
	cm.viper.SetDefault("middleware.cors.allow_credentials", false)
	cm.viper.SetDefault("middleware.cors.max_age", 3600)

	cm.viper.SetDefault("middleware.rate_limit.enable", false)
	cm.viper.SetDefault("middleware.rate_limit.rate", 100)
	cm.viper.SetDefault("middleware.rate_limit.burst", 200)
	cm.viper.SetDefault("middleware.rate_limit.strategy", "token_bucket")

	cm.viper.SetDefault("middleware.auth.enable", false)
	cm.viper.SetDefault("middleware.auth.jwt_secret", "your-secret-key")
	cm.viper.SetDefault("middleware.auth.token_ttl", 24)
	cm.viper.SetDefault("middleware.auth.refresh_ttl", 168)

	// 监控默认配置
	cm.viper.SetDefault("monitor.enable", false)
	cm.viper.SetDefault("monitor.endpoint", "/metrics")
	cm.viper.SetDefault("monitor.interval", 30)
	cm.viper.SetDefault("monitor.timeout", 10)
}

// createDefaultConfigFile 创建默认配置文件
func (cm *ViperConfigManager) createDefaultConfigFile() error {
	// 使用第一个配置路径，或默认使用 ./config
	configDir := "./config"
	if len(cm.configPaths) > 0 {
		configDir = filepath.Join(cm.configPaths[0], "config")
	}
	configFile := filepath.Join(configDir, cm.configName+"."+cm.configType)

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 检查文件是否已存在
	if _, err := os.Stat(configFile); err == nil {
		return nil // 文件已存在，不需要创建
	}

	// 创建默认配置内容
	defaultConfig := cm.generateDefaultConfigContent()

	// 写入配置文件
	if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	GetGlobalLogger().WithFields(map[string]any{
		"config_file": configFile,
	}).Info("默认配置文件创建成功")

	return nil
}

// generateDefaultConfigContent 生成默认配置内容
func (cm *ViperConfigManager) generateDefaultConfigContent() string {
	return `# YYHertz Framework Configuration
# 配置文件格式: YAML

# 应用基础配置
app:
  name: "YYHertz"
  version: "1.0.0"
  environment: "development"  # development, testing, production
  debug: true
  port: 8888
  host: "0.0.0.0"
  timezone: "Asia/Shanghai"

# 数据库配置
database:
  driver: "mysql"
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: ""
  database: "yyhertz"
  charset: "utf8mb4"
  max_idle: 10
  max_open: 100
  max_life: 3600  # 秒
  ssl_mode: "disable"

# Redis配置
redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  database: 0
  max_retries: 3
  pool_size: 10
  min_idle: 2
  dial_timeout: 5  # 秒
  read_timeout: 3  # 秒

# 日志配置
log:
  level: "info"          # debug, info, warn, error, fatal, panic
  format: "json"         # json, text
  enable_console: true
  enable_file: false
  file_path: "./logs/app.log"
  max_size: 100          # MB
  max_age: 7            # 天
  max_backups: 10
  compress: true
  show_caller: true
  show_timestamp: true

# TLS配置
tls:
  enable: false
  cert_file: ""
  key_file: ""
  min_version: "1.2"
  max_version: "1.3"
  auto_reload: false
  reload_interval: 300

# 中间件配置
middleware:
  # CORS跨域配置
  cors:
    enable: true
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["*"]
    expose_headers: []
    allow_credentials: false
    max_age: 3600

  # 限流配置
  rate_limit:
    enable: false
    rate: 100              # 请求/秒
    burst: 200             # 突发容量
    strategy: "token_bucket"  # token_bucket, sliding_window

  # 认证配置
  auth:
    enable: false
    jwt_secret: "your-secret-key-change-me"
    token_ttl: 24          # 小时
    refresh_ttl: 168       # 小时

# 外部服务配置
services:
  # 邮件服务
  email:
    provider: "smtp"       # smtp, sendgrid, ses
    host: "smtp.gmail.com"
    port: 587
    username: ""
    password: ""
    from: "noreply@example.com"

  # 文件存储
  storage:
    provider: "local"      # local, s3, oss
    local_path: "./uploads"
    bucket: ""
    region: ""
    access_key: ""
    secret_key: ""
    cdn_domain: ""

# 监控配置
monitor:
  enable: false
  endpoint: "/metrics"
  interval: 30          # 秒
  timeout: 10           # 秒
`
}

// GetConfig 获取完整配置
func (cm *ViperConfigManager) GetConfig() (*AppConfig, error) {
	if !cm.initialized {
		if err := cm.Initialize(); err != nil {
			return nil, err
		}
	}

	var config AppConfig
	if err := cm.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &config, nil
}

// Get 获取配置值
func (cm *ViperConfigManager) Get(key string) any {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.Get(key)
}

// GetString 获取字符串配置值
func (cm *ViperConfigManager) GetString(key string) string {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.GetString(key)
}

// GetInt 获取整数配置值
func (cm *ViperConfigManager) GetInt(key string) int {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.GetInt(key)
}

// GetBool 获取布尔配置值
func (cm *ViperConfigManager) GetBool(key string) bool {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.GetBool(key)
}

// GetStringSlice 获取字符串数组配置值
func (cm *ViperConfigManager) GetStringSlice(key string) []string {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.GetStringSlice(key)
}

// GetDuration 获取时间间隔配置值
func (cm *ViperConfigManager) GetDuration(key string) time.Duration {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.GetDuration(key)
}

// Set 设置配置值
func (cm *ViperConfigManager) Set(key string, value any) {
	if !cm.initialized {
		cm.Initialize()
	}
	cm.viper.Set(key, value)
}

// IsSet 检查配置是否已设置
func (cm *ViperConfigManager) IsSet(key string) bool {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.IsSet(key)
}

// AllKeys 获取所有配置键
func (cm *ViperConfigManager) AllKeys() []string {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.AllKeys()
}

// WatchConfig 监听配置文件变化
func (cm *ViperConfigManager) WatchConfig() {
	if !cm.initialized {
		cm.Initialize()
	}

	cm.viper.WatchConfig()
	cm.viper.OnConfigChange(func(e fsnotify.Event) {
		GetGlobalLogger().WithFields(map[string]any{
			"file":      e.Name,
			"operation": e.Op.String(),
		}).Info("配置文件发生变化，重新加载")
	})
}

// WriteConfig 写入配置文件
func (cm *ViperConfigManager) WriteConfig() error {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.WriteConfig()
}

// WriteConfigAs 写入配置到指定文件
func (cm *ViperConfigManager) WriteConfigAs(filename string) error {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.WriteConfigAs(filename)
}

// SetConfigFile 设置配置文件路径
func (cm *ViperConfigManager) SetConfigFile(file string) {
	cm.viper.SetConfigFile(file)
}

// SetConfigName 设置配置文件名
func (cm *ViperConfigManager) SetConfigName(name string) {
	cm.configName = name
	cm.viper.SetConfigName(name)
}

// SetConfigType 设置配置文件类型
func (cm *ViperConfigManager) SetConfigType(configType string) {
	cm.configType = configType
	cm.viper.SetConfigType(configType)
}

// AddConfigPath 添加配置文件搜索路径
func (cm *ViperConfigManager) AddConfigPath(path string) {
	cm.configPaths = append(cm.configPaths, path)
	cm.viper.AddConfigPath(path)
}

// SetEnvPrefix 设置环境变量前缀
func (cm *ViperConfigManager) SetEnvPrefix(prefix string) {
	cm.envPrefix = prefix
	cm.viper.SetEnvPrefix(prefix)
}

// ConfigFileUsed 获取当前使用的配置文件路径
func (cm *ViperConfigManager) ConfigFileUsed() string {
	if !cm.initialized {
		cm.Initialize()
	}
	return cm.viper.ConfigFileUsed()
}

// 全局便捷函数

// GetGlobalConfig 获取全局配置
func GetGlobalConfig(configName ...string) (*AppConfig, error) {
	name := getConfigName(configName...)
	return GetViperConfigManagerWithName(name).GetConfig()
}

// GetConfigValue 获取全局配置值
func GetConfigValue(key string, configName ...string) any {
	name := getConfigName(configName...)
	return GetViperConfigManagerWithName(name).Get(key)
}

// GetConfigString 获取全局字符串配置值
func GetConfigString(key string, configName ...string) string {
	name := getConfigName(configName...)
	return GetViperConfigManagerWithName(name).GetString(key)
}

// GetConfigInt 获取全局整数配置值
func GetConfigInt(key string, configName ...string) int {
	name := getConfigName(configName...)
	return GetViperConfigManagerWithName(name).GetInt(key)
}

// GetConfigBool 获取全局布尔配置值
func GetConfigBool(key string, configName ...string) bool {
	name := getConfigName(configName...)
	return GetViperConfigManagerWithName(name).GetBool(key)
}

func getConfigName(configName ...string) string {
	if len(configName) > 0 {
		return configName[0]
	}
	return DefaultConfigName
}
