package config

import (
	"github.com/spf13/viper"
)

// SessionConfig Session配置结构
type SessionConfig struct {
	// Cookie配置
	Cookie struct {
		Name     string `mapstructure:"name" yaml:"name" json:"name"`                // Cookie名称
		Domain   string `mapstructure:"domain" yaml:"domain" json:"domain"`          // Cookie域名
		Path     string `mapstructure:"path" yaml:"path" json:"path"`                // Cookie路径
		MaxAge   int    `mapstructure:"max_age" yaml:"max_age" json:"max_age"`       // Cookie最大生存时间(秒)
		Secure   bool   `mapstructure:"secure" yaml:"secure" json:"secure"`          // 是否仅HTTPS传输
		HttpOnly bool   `mapstructure:"http_only" yaml:"http_only" json:"http_only"` // 是否HttpOnly
		SameSite string `mapstructure:"same_site" yaml:"same_site" json:"same_site"` // SameSite策略: Strict, Lax, None

		// 签名配置
		Sign struct {
			Enable bool   `mapstructure:"enable" yaml:"enable" json:"enable"` // 启用Cookie签名
			Secret string `mapstructure:"secret" yaml:"secret" json:"secret"` // 签名密钥
		} `mapstructure:"sign" yaml:"sign" json:"sign"`

		// 加密配置
		Encrypt struct {
			Enable    bool   `mapstructure:"enable" yaml:"enable" json:"enable"`             // 启用Cookie加密
			SecretKey string `mapstructure:"secret_key" yaml:"secret_key" json:"secret_key"` // 加密密钥
			Algorithm string `mapstructure:"algorithm" yaml:"algorithm" json:"algorithm"`    // 加密算法: AES256, AES128
		} `mapstructure:"encrypt" yaml:"encrypt" json:"encrypt"`

		// 压缩配置
		Compress struct {
			Enable    bool   `mapstructure:"enable" yaml:"enable" json:"enable"`          // 启用Cookie压缩
			Level     int    `mapstructure:"level" yaml:"level" json:"level"`             // 压缩级别 1-9
			Algorithm string `mapstructure:"algorithm" yaml:"algorithm" json:"algorithm"` // 压缩算法: gzip, deflate
		} `mapstructure:"compress" yaml:"compress" json:"compress"`
	} `mapstructure:"cookie" yaml:"cookie" json:"cookie"`

	// Session配置
	Session struct {
		Name   string `mapstructure:"name" yaml:"name" json:"name"`          // Session名称
		Secret string `mapstructure:"secret" yaml:"secret" json:"secret"`    // Session密钥
		MaxAge int    `mapstructure:"max_age" yaml:"max_age" json:"max_age"` // Session最大生存时间(秒)

		// Cookie配置 (Session Cookie)
		Cookie struct {
			Path     string `mapstructure:"path" yaml:"path" json:"path"`                // Session Cookie路径
			Domain   string `mapstructure:"domain" yaml:"domain" json:"domain"`          // Session Cookie域名
			Secure   bool   `mapstructure:"secure" yaml:"secure" json:"secure"`          // 是否仅HTTPS传输
			HttpOnly bool   `mapstructure:"http_only" yaml:"http_only" json:"http_only"` // 是否HttpOnly
			SameSite string `mapstructure:"same_site" yaml:"same_site" json:"same_site"` // SameSite策略
		} `mapstructure:"cookie" yaml:"cookie" json:"cookie"`

		// 存储配置
		Store struct {
			Type   string `mapstructure:"type" yaml:"type" json:"type"`       // 存储类型: cookie, memory, redis, file
			Prefix string `mapstructure:"prefix" yaml:"prefix" json:"prefix"` // 存储键前缀

			// Redis存储配置
			Redis struct {
				Addr     string `mapstructure:"addr" yaml:"addr" json:"addr"`                // Redis地址
				Password string `mapstructure:"password" yaml:"password" json:"password"`    // Redis密码
				DB       int    `mapstructure:"db" yaml:"db" json:"db"`                      // Redis数据库
				PoolSize int    `mapstructure:"pool_size" yaml:"pool_size" json:"pool_size"` // 连接池大小
			} `mapstructure:"redis" yaml:"redis" json:"redis"`

			// 文件存储配置
			File struct {
				Dir      string `mapstructure:"dir" yaml:"dir" json:"dir"`                   // 存储目录
				FileMode string `mapstructure:"file_mode" yaml:"file_mode" json:"file_mode"` // 文件权限
			} `mapstructure:"file" yaml:"file" json:"file"`

			// 内存存储配置
			Memory struct {
				MaxSize    int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"`          // 最大条目数
				GCInterval string `mapstructure:"gc_interval" yaml:"gc_interval" json:"gc_interval"` // 垃圾回收间隔
			} `mapstructure:"memory" yaml:"memory" json:"memory"`
		} `mapstructure:"store" yaml:"store" json:"store"`

		// 安全配置
		Security struct {
			RegenerateID       bool `mapstructure:"regenerate_id" yaml:"regenerate_id" json:"regenerate_id"`                   // 登录时重新生成Session ID
			CSRFProtection     bool `mapstructure:"csrf_protection" yaml:"csrf_protection" json:"csrf_protection"`             // CSRF保护
			EncryptData        bool `mapstructure:"encrypt_data" yaml:"encrypt_data" json:"encrypt_data"`                      // 加密Session数据
			SecureTransmission bool `mapstructure:"secure_transmission" yaml:"secure_transmission" json:"secure_transmission"` // 安全传输
		} `mapstructure:"security" yaml:"security" json:"security"`

		// 清理配置
		Cleanup struct {
			Enable    bool   `mapstructure:"enable" yaml:"enable" json:"enable"`             // 启用过期清理
			Interval  string `mapstructure:"interval" yaml:"interval" json:"interval"`       // 清理间隔
			BatchSize int    `mapstructure:"batch_size" yaml:"batch_size" json:"batch_size"` // 批量清理大小
		} `mapstructure:"cleanup" yaml:"cleanup" json:"cleanup"`
	} `mapstructure:"session" yaml:"session" json:"session"`

	// 中间件配置
	Middleware struct {
		// Cookie中间件配置
		Cookie struct {
			Enable           bool     `mapstructure:"enable" yaml:"enable" json:"enable"`                                     // 启用Cookie中间件
			TrustedProxies   []string `mapstructure:"trusted_proxies" yaml:"trusted_proxies" json:"trusted_proxies"`          // 信任的代理
			AllowedOrigins   []string `mapstructure:"allowed_origins" yaml:"allowed_origins" json:"allowed_origins"`          // 允许的来源
			BlockedUserAgent []string `mapstructure:"blocked_user_agent" yaml:"blocked_user_agent" json:"blocked_user_agent"` // 阻止的用户代理
		} `mapstructure:"cookie" yaml:"cookie" json:"cookie"`

		// Session中间件配置
		Session struct {
			Enable        bool     `mapstructure:"enable" yaml:"enable" json:"enable"`                         // 启用Session中间件
			SkipPaths     []string `mapstructure:"skip_paths" yaml:"skip_paths" json:"skip_paths"`             // 跳过的路径
			ErrorHandling string   `mapstructure:"error_handling" yaml:"error_handling" json:"error_handling"` // 错误处理: ignore, log, panic
			AutoSave      bool     `mapstructure:"auto_save" yaml:"auto_save" json:"auto_save"`                // 自动保存
			LazyLoading   bool     `mapstructure:"lazy_loading" yaml:"lazy_loading" json:"lazy_loading"`       // 延迟加载
		} `mapstructure:"session" yaml:"session" json:"session"`
	} `mapstructure:"middleware" yaml:"middleware" json:"middleware"`

	// 监控配置
	Monitoring struct {
		Enable         bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                            // 启用监控
		MetricsPath    string `mapstructure:"metrics_path" yaml:"metrics_path" json:"metrics_path"`          // 监控指标路径
		CollectCookie  bool   `mapstructure:"collect_cookie" yaml:"collect_cookie" json:"collect_cookie"`    // 收集Cookie指标
		CollectSession bool   `mapstructure:"collect_session" yaml:"collect_session" json:"collect_session"` // 收集Session指标
		ReportInterval string `mapstructure:"report_interval" yaml:"report_interval" json:"report_interval"` // 报告间隔
	} `mapstructure:"monitoring" yaml:"monitoring" json:"monitoring"`

	// 开发配置
	Development struct {
		Enable       bool `mapstructure:"enable" yaml:"enable" json:"enable"`                      // 启用开发模式
		ShowCookies  bool `mapstructure:"show_cookies" yaml:"show_cookies" json:"show_cookies"`    // 显示Cookie信息
		ShowSessions bool `mapstructure:"show_sessions" yaml:"show_sessions" json:"show_sessions"` // 显示Session信息
		DebugMode    bool `mapstructure:"debug_mode" yaml:"debug_mode" json:"debug_mode"`          // 调试模式
	} `mapstructure:"development" yaml:"development" json:"development"`
}

// GetConfigName 实现 ConfigInterface 接口
func (c SessionConfig) GetConfigName() string {
	return SessionConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c SessionConfig) SetDefaults(v *viper.Viper) {
	// Cookie默认配置
	v.SetDefault("cookie.name", "yyhertz_cookie")
	v.SetDefault("cookie.domain", "")
	v.SetDefault("cookie.path", "/")
	v.SetDefault("cookie.max_age", 86400) // 24小时
	v.SetDefault("cookie.secure", false)
	v.SetDefault("cookie.http_only", true)
	v.SetDefault("cookie.same_site", "Lax")

	// Cookie签名配置
	v.SetDefault("cookie.sign.enable", false)
	v.SetDefault("cookie.sign.secret", "your-cookie-sign-secret-change-me")

	// Cookie加密配置
	v.SetDefault("cookie.encrypt.enable", false)
	v.SetDefault("cookie.encrypt.secret_key", "your-32-char-secret-key-change-me")
	v.SetDefault("cookie.encrypt.algorithm", "AES256")

	// Cookie压缩配置
	v.SetDefault("cookie.compress.enable", false)
	v.SetDefault("cookie.compress.level", 6)
	v.SetDefault("cookie.compress.algorithm", "gzip")

	// Session默认配置
	v.SetDefault("session.name", "YYHERTZ_SESSION")
	v.SetDefault("session.secret", "your-session-secret-key-change-me")
	v.SetDefault("session.max_age", 3600) // 1小时

	// Session Cookie配置
	v.SetDefault("session.cookie.path", "/")
	v.SetDefault("session.cookie.domain", "")
	v.SetDefault("session.cookie.secure", false)
	v.SetDefault("session.cookie.http_only", true)
	v.SetDefault("session.cookie.same_site", "Lax")

	// Session存储配置
	v.SetDefault("session.store.type", "cookie")
	v.SetDefault("session.store.prefix", "session:")

	// Redis存储配置
	v.SetDefault("session.store.redis.addr", "localhost:6379")
	v.SetDefault("session.store.redis.password", "")
	v.SetDefault("session.store.redis.db", 0)
	v.SetDefault("session.store.redis.pool_size", 10)

	// 文件存储配置
	v.SetDefault("session.store.file.dir", "./sessions")
	v.SetDefault("session.store.file.file_mode", "0600")

	// 内存存储配置
	v.SetDefault("session.store.memory.max_size", 1000)
	v.SetDefault("session.store.memory.gc_interval", "10m")

	// Session安全配置
	v.SetDefault("session.security.regenerate_id", true)
	v.SetDefault("session.security.csrf_protection", false)
	v.SetDefault("session.security.encrypt_data", false)
	v.SetDefault("session.security.secure_transmission", false)

	// Session清理配置
	v.SetDefault("session.cleanup.enable", true)
	v.SetDefault("session.cleanup.interval", "1h")
	v.SetDefault("session.cleanup.batch_size", 100)

	// 中间件默认配置
	v.SetDefault("middleware.cookie.enable", true)
	v.SetDefault("middleware.cookie.trusted_proxies", []string{})
	v.SetDefault("middleware.cookie.allowed_origins", []string{"*"})
	v.SetDefault("middleware.cookie.blocked_user_agent", []string{})

	v.SetDefault("middleware.session.enable", true)
	v.SetDefault("middleware.session.skip_paths", []string{"/health", "/metrics"})
	v.SetDefault("middleware.session.error_handling", "log")
	v.SetDefault("middleware.session.auto_save", true)
	v.SetDefault("middleware.session.lazy_loading", false)

	// 监控默认配置
	v.SetDefault("monitoring.enable", false)
	v.SetDefault("monitoring.metrics_path", "/metrics")
	v.SetDefault("monitoring.collect_cookie", true)
	v.SetDefault("monitoring.collect_session", true)
	v.SetDefault("monitoring.report_interval", "30s")

	// 开发默认配置
	v.SetDefault("development.enable", false)
	v.SetDefault("development.show_cookies", false)
	v.SetDefault("development.show_sessions", false)
	v.SetDefault("development.debug_mode", false)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c SessionConfig) GenerateDefaultContent() string {
	return `# YYHertz Session Configuration
# Session配置文件

# Cookie配置
cookie:
  name: "yyhertz_cookie"                 # Cookie名称
  domain: ""                             # Cookie域名(为空表示当前域名)
  path: "/"                              # Cookie路径
  max_age: 86400                         # Cookie最大生存时间(秒) - 24小时
  secure: false                          # 是否仅HTTPS传输
  http_only: true                        # 是否HttpOnly(防XSS)
  same_site: "Lax"                       # SameSite策略: Strict, Lax, None
  
  # Cookie签名配置
  sign:
    enable: false                        # 启用Cookie签名
    secret: "your-cookie-sign-secret-change-me" # 签名密钥
  
  # Cookie加密配置
  encrypt:
    enable: false                        # 启用Cookie加密
    secret_key: "your-32-char-secret-key-change-me" # 加密密钥(32字符)
    algorithm: "AES256"                  # 加密算法: AES256, AES128
  
  # Cookie压缩配置
  compress:
    enable: false                        # 启用Cookie压缩
    level: 6                             # 压缩级别 1-9
    algorithm: "gzip"                    # 压缩算法: gzip, deflate

# Session配置
session:
  name: "YYHERTZ_SESSION"                # Session名称
  secret: "your-session-secret-key-change-me" # Session密钥
  max_age: 3600                          # Session最大生存时间(秒) - 1小时
  
  # Session Cookie配置
  cookie:
    path: "/"                            # Session Cookie路径
    domain: ""                           # Session Cookie域名
    secure: false                        # 是否仅HTTPS传输
    http_only: true                      # 是否HttpOnly
    same_site: "Lax"                     # SameSite策略
  
  # 存储配置
  store:
    type: "cookie"                       # 存储类型: cookie, memory, redis, file
    prefix: "session:"                   # 存储键前缀
    
    # Redis存储配置
    redis:
      addr: "localhost:6379"             # Redis地址
      password: ""                       # Redis密码
      db: 0                              # Redis数据库
      pool_size: 10                      # 连接池大小
    
    # 文件存储配置
    file:
      dir: "./sessions"                  # 存储目录
      file_mode: "0600"                  # 文件权限
    
    # 内存存储配置
    memory:
      max_size: 1000                     # 最大条目数
      gc_interval: "10m"                 # 垃圾回收间隔
  
  # 安全配置
  security:
    regenerate_id: true                  # 登录时重新生成Session ID
    csrf_protection: false               # CSRF保护
    encrypt_data: false                  # 加密Session数据
    secure_transmission: false           # 安全传输
  
  # 清理配置
  cleanup:
    enable: true                         # 启用过期清理
    interval: "1h"                       # 清理间隔
    batch_size: 100                      # 批量清理大小

# 中间件配置
middleware:
  # Cookie中间件配置
  cookie:
    enable: true                         # 启用Cookie中间件
    trusted_proxies: []                  # 信任的代理
    allowed_origins: ["*"]               # 允许的来源
    blocked_user_agent: []               # 阻止的用户代理
  
  # Session中间件配置  
  session:
    enable: true                         # 启用Session中间件
    skip_paths: ["/health", "/metrics"]  # 跳过的路径
    error_handling: "log"                # 错误处理: ignore, log, panic
    auto_save: true                      # 自动保存
    lazy_loading: false                  # 延迟加载

# 监控配置
monitoring:
  enable: false                          # 启用监控
  metrics_path: "/metrics"               # 监控指标路径
  collect_cookie: true                   # 收集Cookie指标
  collect_session: true                  # 收集Session指标
  report_interval: "30s"                 # 报告间隔

# 开发配置
development:
  enable: false                          # 启用开发模式
  show_cookies: false                    # 显示Cookie信息
  show_sessions: false                   # 显示Session信息
  debug_mode: false                      # 调试模式

# 使用示例:
# 1. Cookie设置:
#    cookie.name = "my_app_cookie"
#    cookie.max_age = 7 * 24 * 3600  // 7天
#    cookie.secure = true            // 生产环境建议启用
#
# 2. Session存储:
#    session.store.type = "redis"    // 生产环境建议使用Redis
#    session.store.redis.addr = "redis-cluster:6379"
#
# 3. 安全配置:
#    cookie.encrypt.enable = true    // 敏感数据建议加密
#    session.security.csrf_protection = true // 启用CSRF保护
`
}
