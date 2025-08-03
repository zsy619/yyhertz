package config

import (
	"github.com/spf13/viper"
)

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

// GetConfigName 实现 ConfigInterface 接口
func (c AppConfig) GetConfigName() string {
	return AppConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c AppConfig) SetDefaults(v *viper.Viper) {
	// 应用默认配置
	v.SetDefault("app.name", "YYHertz")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.port", 8888)
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.timezone", "Asia/Shanghai")

	// 数据库默认配置
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.username", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.database", "yyhertz")
	v.SetDefault("database.charset", "utf8mb4")
	v.SetDefault("database.max_idle", 10)
	v.SetDefault("database.max_open", 100)
	v.SetDefault("database.max_life", 3600)
	v.SetDefault("database.ssl_mode", "disable")

	// Redis默认配置
	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.max_retries", 3)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle", 2)
	v.SetDefault("redis.dial_timeout", 5)
	v.SetDefault("redis.read_timeout", 3)

	// 日志默认配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.enable_console", true)
	v.SetDefault("log.enable_file", false)
	v.SetDefault("log.file_path", "./logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_age", 7)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.compress", true)

	// TLS默认配置
	v.SetDefault("tls.enable", false)
	v.SetDefault("tls.cert_file", "")
	v.SetDefault("tls.key_file", "")
	v.SetDefault("tls.min_version", "1.2")
	v.SetDefault("tls.max_version", "1.3")

	// 中间件默认配置
	v.SetDefault("middleware.cors.enable", true)
	v.SetDefault("middleware.cors.allow_origins", []string{"*"})
	v.SetDefault("middleware.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("middleware.cors.allow_headers", []string{"*"})
	v.SetDefault("middleware.cors.allow_credentials", false)
	v.SetDefault("middleware.cors.max_age", 3600)

	v.SetDefault("middleware.rate_limit.enable", false)
	v.SetDefault("middleware.rate_limit.rate", 100)
	v.SetDefault("middleware.rate_limit.burst", 200)
	v.SetDefault("middleware.rate_limit.strategy", "token_bucket")

	v.SetDefault("middleware.auth.enable", false)
	v.SetDefault("middleware.auth.jwt_secret", "your-secret-key")
	v.SetDefault("middleware.auth.token_ttl", 24)
	v.SetDefault("middleware.auth.refresh_ttl", 168)

	// 监控默认配置
	v.SetDefault("monitor.enable", false)
	v.SetDefault("monitor.endpoint", "/metrics")
	v.SetDefault("monitor.interval", 30)
	v.SetDefault("monitor.timeout", 10)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c AppConfig) GenerateDefaultContent() string {
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
