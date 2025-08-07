package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// MiddlewareUnifiedConfigName 统一中间件配置名称常量
const MiddlewareUnifiedConfigName = "middleware_unified"

// MiddlewareMode 中间件模式
type MiddlewareMode string

const (
	BasicMode    MiddlewareMode = "basic"    // 基础模式
	AdvancedMode MiddlewareMode = "advanced" // 高级模式
	AutoMode     MiddlewareMode = "auto"     // 自动模式
)

// MiddlewareUnifiedConfig 统一中间件配置
type MiddlewareUnifiedConfig struct {
	// 全局配置
	Global struct {
		Mode               MiddlewareMode `mapstructure:"mode" yaml:"mode" json:"mode"`
		EnableStatistics   bool           `mapstructure:"enable_statistics" yaml:"enable_statistics" json:"enable_statistics"`
		EnableMonitoring   bool           `mapstructure:"enable_monitoring" yaml:"enable_monitoring" json:"enable_monitoring"`
		EnableDebug        bool           `mapstructure:"enable_debug" yaml:"enable_debug" json:"enable_debug"`
	} `mapstructure:"global" yaml:"global" json:"global"`

	// 基础模式配置
	Basic struct {
		EnableChain        bool     `mapstructure:"enable_chain" yaml:"enable_chain" json:"enable_chain"`
		DefaultMiddlewares []string `mapstructure:"default_middlewares" yaml:"default_middlewares" json:"default_middlewares"`
		MaxHandlers        int      `mapstructure:"max_handlers" yaml:"max_handlers" json:"max_handlers"`
	} `mapstructure:"basic" yaml:"basic" json:"basic"`

	// 高级模式配置 (MVC)
	Advanced struct {
		EnableOptimization       bool          `mapstructure:"enable_optimization" yaml:"enable_optimization" json:"enable_optimization"`
		EnableCompilation        bool          `mapstructure:"enable_compilation" yaml:"enable_compilation" json:"enable_compilation"`
		EnableLayeredArchitecture bool         `mapstructure:"enable_layered_architecture" yaml:"enable_layered_architecture" json:"enable_layered_architecture"`
		EnableDependencyAnalysis bool          `mapstructure:"enable_dependency_analysis" yaml:"enable_dependency_analysis" json:"enable_dependency_analysis"`
		CacheSize               int           `mapstructure:"cache_size" yaml:"cache_size" json:"cache_size"`
		CacheExpireTime         time.Duration `mapstructure:"cache_expire_time" yaml:"cache_expire_time" json:"cache_expire_time"`
		CompileTimeout          time.Duration `mapstructure:"compile_timeout" yaml:"compile_timeout" json:"compile_timeout"`
		MaxConcurrency          int           `mapstructure:"max_concurrency" yaml:"max_concurrency" json:"max_concurrency"`
	} `mapstructure:"advanced" yaml:"advanced" json:"advanced"`

	// 自动模式配置
	Auto struct {
		EnableAutoSwitch      bool          `mapstructure:"enable_auto_switch" yaml:"enable_auto_switch" json:"enable_auto_switch"`
		RequestThreshold      int           `mapstructure:"request_threshold" yaml:"request_threshold" json:"request_threshold"`
		ResponseTimeLimit     time.Duration `mapstructure:"response_time_limit" yaml:"response_time_limit" json:"response_time_limit"`
		SwitchCheckInterval   time.Duration `mapstructure:"switch_check_interval" yaml:"switch_check_interval" json:"switch_check_interval"`
		PerformanceThreshold  float64       `mapstructure:"performance_threshold" yaml:"performance_threshold" json:"performance_threshold"`
		EnableUpgrade         bool          `mapstructure:"enable_upgrade" yaml:"enable_upgrade" json:"enable_upgrade"`
		EnableDowngrade       bool          `mapstructure:"enable_downgrade" yaml:"enable_downgrade" json:"enable_downgrade"`
	} `mapstructure:"auto" yaml:"auto" json:"auto"`

	// 内置中间件配置
	Builtin struct {
		Logger struct {
			Enable        bool     `mapstructure:"enable" yaml:"enable" json:"enable"`
			Mode         string   `mapstructure:"mode" yaml:"mode" json:"mode"` // enhanced, basic
			Format       string   `mapstructure:"format" yaml:"format" json:"format"`
			TimeFormat   string   `mapstructure:"time_format" yaml:"time_format" json:"time_format"`
			SkipPaths    []string `mapstructure:"skip_paths" yaml:"skip_paths" json:"skip_paths"`
			EnableColors bool     `mapstructure:"enable_colors" yaml:"enable_colors" json:"enable_colors"`
		} `mapstructure:"logger" yaml:"logger" json:"logger"`

		Recovery struct {
			Enable      bool `mapstructure:"enable" yaml:"enable" json:"enable"`
			Mode        string `mapstructure:"mode" yaml:"mode" json:"mode"` // enhanced, basic
			EnableStack bool `mapstructure:"enable_stack" yaml:"enable_stack" json:"enable_stack"`
		} `mapstructure:"recovery" yaml:"recovery" json:"recovery"`

		CORS struct {
			Enable           bool     `mapstructure:"enable" yaml:"enable" json:"enable"`
			AllowOrigins     []string `mapstructure:"allow_origins" yaml:"allow_origins" json:"allow_origins"`
			AllowMethods     []string `mapstructure:"allow_methods" yaml:"allow_methods" json:"allow_methods"`
			AllowHeaders     []string `mapstructure:"allow_headers" yaml:"allow_headers" json:"allow_headers"`
			ExposeHeaders    []string `mapstructure:"expose_headers" yaml:"expose_headers" json:"expose_headers"`
			AllowCredentials bool     `mapstructure:"allow_credentials" yaml:"allow_credentials" json:"allow_credentials"`
			MaxAge           int      `mapstructure:"max_age" yaml:"max_age" json:"max_age"`
		} `mapstructure:"cors" yaml:"cors" json:"cors"`

		Auth struct {
			Enable      bool     `mapstructure:"enable" yaml:"enable" json:"enable"`
			Type        string   `mapstructure:"type" yaml:"type" json:"type"` // basic, jwt, custom
			SkipPaths   []string `mapstructure:"skip_paths" yaml:"skip_paths" json:"skip_paths"`
			SecretKey   string   `mapstructure:"secret_key" yaml:"secret_key" json:"secret_key"`
			TokenHeader string   `mapstructure:"token_header" yaml:"token_header" json:"token_header"`
		} `mapstructure:"auth" yaml:"auth" json:"auth"`

		RateLimit struct {
			Enable   bool          `mapstructure:"enable" yaml:"enable" json:"enable"`
			Rate     int           `mapstructure:"rate" yaml:"rate" json:"rate"`
			Burst    int           `mapstructure:"burst" yaml:"burst" json:"burst"`
			Window   time.Duration `mapstructure:"window" yaml:"window" json:"window"`
			Strategy string        `mapstructure:"strategy" yaml:"strategy" json:"strategy"` // token_bucket, sliding_window
		} `mapstructure:"rate_limit" yaml:"rate_limit" json:"rate_limit"`

		Tracing struct {
			Enable     bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			Provider   string `mapstructure:"provider" yaml:"provider" json:"provider"` // jaeger, zipkin, custom
			Endpoint   string `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
			SampleRate float64 `mapstructure:"sample_rate" yaml:"sample_rate" json:"sample_rate"`
		} `mapstructure:"tracing" yaml:"tracing" json:"tracing"`

		RequestID struct {
			Enable    bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			Header    string `mapstructure:"header" yaml:"header" json:"header"`
			Generator string `mapstructure:"generator" yaml:"generator" json:"generator"` // uuid, timestamp, custom
		} `mapstructure:"request_id" yaml:"request_id" json:"request_id"`

		Timeout struct {
			Enable  bool          `mapstructure:"enable" yaml:"enable" json:"enable"`
			Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
		} `mapstructure:"timeout" yaml:"timeout" json:"timeout"`

		Secure struct {
			Enable                bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			ContentSecurityPolicy string `mapstructure:"content_security_policy" yaml:"content_security_policy" json:"content_security_policy"`
			EnableHSTS            bool   `mapstructure:"enable_hsts" yaml:"enable_hsts" json:"enable_hsts"`
			EnableFrameOptions    bool   `mapstructure:"enable_frame_options" yaml:"enable_frame_options" json:"enable_frame_options"`
		} `mapstructure:"secure" yaml:"secure" json:"secure"`

		GZip struct {
			Enable      bool     `mapstructure:"enable" yaml:"enable" json:"enable"`
			Level       int      `mapstructure:"level" yaml:"level" json:"level"` // 1-9
			MinSize     int      `mapstructure:"min_size" yaml:"min_size" json:"min_size"`
			ExcludePaths []string `mapstructure:"exclude_paths" yaml:"exclude_paths" json:"exclude_paths"`
		} `mapstructure:"gzip" yaml:"gzip" json:"gzip"`
	} `mapstructure:"builtin" yaml:"builtin" json:"builtin"`

	// 自定义中间件配置
	Custom map[string]interface{} `mapstructure:"custom" yaml:"custom" json:"custom"`
}

// GetConfigName 实现 ConfigInterface 接口
func (c MiddlewareUnifiedConfig) GetConfigName() string {
	return MiddlewareUnifiedConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c MiddlewareUnifiedConfig) SetDefaults(v *viper.Viper) {
	// 全局默认配置
	v.SetDefault("global.mode", string(AutoMode))
	v.SetDefault("global.enable_statistics", true)
	v.SetDefault("global.enable_monitoring", true)
	v.SetDefault("global.enable_debug", false)

	// 基础模式默认配置
	v.SetDefault("basic.enable_chain", true)
	v.SetDefault("basic.default_middlewares", []string{"logger", "recovery"})
	v.SetDefault("basic.max_handlers", 64)

	// 高级模式默认配置
	v.SetDefault("advanced.enable_optimization", true)
	v.SetDefault("advanced.enable_compilation", true)
	v.SetDefault("advanced.enable_layered_architecture", true)
	v.SetDefault("advanced.enable_dependency_analysis", true)
	v.SetDefault("advanced.cache_size", 500)
	v.SetDefault("advanced.cache_expire_time", "30m")
	v.SetDefault("advanced.compile_timeout", "30s")
	v.SetDefault("advanced.max_concurrency", 100)

	// 自动模式默认配置
	v.SetDefault("auto.enable_auto_switch", true)
	v.SetDefault("auto.request_threshold", 1000)
	v.SetDefault("auto.response_time_limit", "100ms")
	v.SetDefault("auto.switch_check_interval", "5m")
	v.SetDefault("auto.performance_threshold", 0.8)
	v.SetDefault("auto.enable_upgrade", true)
	v.SetDefault("auto.enable_downgrade", true)

	// 内置中间件默认配置
	// Logger
	v.SetDefault("builtin.logger.enable", true)
	v.SetDefault("builtin.logger.mode", "enhanced")
	v.SetDefault("builtin.logger.format", "default")
	v.SetDefault("builtin.logger.time_format", "2006/01/02 - 15:04:05")
	v.SetDefault("builtin.logger.skip_paths", []string{"/health", "/metrics"})
	v.SetDefault("builtin.logger.enable_colors", true)

	// Recovery
	v.SetDefault("builtin.recovery.enable", true)
	v.SetDefault("builtin.recovery.mode", "enhanced")
	v.SetDefault("builtin.recovery.enable_stack", true)

	// CORS
	v.SetDefault("builtin.cors.enable", false)
	v.SetDefault("builtin.cors.allow_origins", []string{"*"})
	v.SetDefault("builtin.cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("builtin.cors.allow_headers", []string{"*"})
	v.SetDefault("builtin.cors.allow_credentials", false)
	v.SetDefault("builtin.cors.max_age", 3600)

	// Auth
	v.SetDefault("builtin.auth.enable", false)
	v.SetDefault("builtin.auth.type", "basic")
	v.SetDefault("builtin.auth.skip_paths", []string{"/health", "/ping"})
	v.SetDefault("builtin.auth.token_header", "Authorization")

	// Rate Limit
	v.SetDefault("builtin.rate_limit.enable", false)
	v.SetDefault("builtin.rate_limit.rate", 100)
	v.SetDefault("builtin.rate_limit.burst", 200)
	v.SetDefault("builtin.rate_limit.window", "1m")
	v.SetDefault("builtin.rate_limit.strategy", "token_bucket")

	// Tracing
	v.SetDefault("builtin.tracing.enable", false)
	v.SetDefault("builtin.tracing.provider", "jaeger")
	v.SetDefault("builtin.tracing.sample_rate", 0.1)

	// Request ID
	v.SetDefault("builtin.request_id.enable", true)
	v.SetDefault("builtin.request_id.header", "X-Request-ID")
	v.SetDefault("builtin.request_id.generator", "timestamp")

	// Timeout
	v.SetDefault("builtin.timeout.enable", false)
	v.SetDefault("builtin.timeout.timeout", "30s")

	// Secure
	v.SetDefault("builtin.secure.enable", false)
	v.SetDefault("builtin.secure.content_security_policy", "default-src 'self'")
	v.SetDefault("builtin.secure.enable_hsts", true)
	v.SetDefault("builtin.secure.enable_frame_options", true)

	// GZip
	v.SetDefault("builtin.gzip.enable", false)
	v.SetDefault("builtin.gzip.level", 6)
	v.SetDefault("builtin.gzip.min_size", 1024)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c MiddlewareUnifiedConfig) GenerateDefaultContent() string {
	return `# YYHertz Unified Middleware Configuration
# 统一中间件系统配置

# 全局配置
global:
  mode: "auto"                    # 中间件模式: basic, advanced, auto
  enable_statistics: true         # 启用统计
  enable_monitoring: true         # 启用监控
  enable_debug: false            # 启用调试模式

# 基础模式配置
basic:
  enable_chain: true             # 启用中间件链
  default_middlewares:           # 默认中间件
    - "logger"
    - "recovery"
  max_handlers: 64               # 最大处理器数量

# 高级模式配置 (MVC)
advanced:
  enable_optimization: true      # 启用优化
  enable_compilation: true       # 启用编译
  enable_layered_architecture: true # 启用分层架构
  enable_dependency_analysis: true  # 启用依赖分析
  cache_size: 500               # 缓存大小
  cache_expire_time: "30m"      # 缓存过期时间
  compile_timeout: "30s"        # 编译超时
  max_concurrency: 100          # 最大并发数

# 自动模式配置
auto:
  enable_auto_switch: true       # 启用自动切换
  request_threshold: 1000        # 请求量阈值
  response_time_limit: "100ms"   # 响应时间限制
  switch_check_interval: "5m"    # 检查间隔
  performance_threshold: 0.8     # 性能阈值
  enable_upgrade: true           # 允许升级到高级模式
  enable_downgrade: true         # 允许降级到基础模式

# 内置中间件配置
builtin:
  # 日志中间件
  logger:
    enable: true
    mode: "enhanced"             # enhanced, basic
    format: "default"            # 日志格式
    time_format: "2006/01/02 - 15:04:05"
    skip_paths:                  # 跳过的路径
      - "/health"
      - "/metrics"
    enable_colors: true          # 启用颜色

  # 恢复中间件
  recovery:
    enable: true
    mode: "enhanced"             # enhanced, basic
    enable_stack: true           # 启用堆栈跟踪

  # CORS中间件
  cors:
    enable: false
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["*"]
    expose_headers: []
    allow_credentials: false
    max_age: 3600

  # 认证中间件
  auth:
    enable: false
    type: "basic"                # basic, jwt, custom
    skip_paths:
      - "/health"
      - "/ping"
    secret_key: ""
    token_header: "Authorization"

  # 限流中间件
  rate_limit:
    enable: false
    rate: 100                    # 请求/秒
    burst: 200                   # 突发容量
    window: "1m"                 # 时间窗口
    strategy: "token_bucket"     # token_bucket, sliding_window

  # 链路追踪中间件
  tracing:
    enable: false
    provider: "jaeger"           # jaeger, zipkin, custom
    endpoint: ""
    sample_rate: 0.1             # 采样率

  # 请求ID中间件
  request_id:
    enable: true
    header: "X-Request-ID"
    generator: "timestamp"       # uuid, timestamp, custom

  # 超时中间件
  timeout:
    enable: false
    timeout: "30s"

  # 安全头中间件
  secure:
    enable: false
    content_security_policy: "default-src 'self'"
    enable_hsts: true
    enable_frame_options: true

  # GZIP压缩中间件
  gzip:
    enable: false
    level: 6                     # 压缩级别 1-9
    min_size: 1024              # 最小压缩大小 (bytes)
    exclude_paths: []           # 排除的路径

# 自定义中间件配置
custom: {}

# 环境特定配置示例:
# development:
#   global:
#     enable_debug: true
#   builtin:
#     logger:
#       enable_colors: true
#
# production:
#   global:
#     enable_debug: false
#   advanced:
#     cache_size: 1000
#   builtin:
#     logger:
#       enable_colors: false
#
# testing:
#   global:
#     mode: "basic"
#   builtin:
#     logger:
#       skip_paths: ["*"]
`
}

// GetMode 获取中间件模式
func (c *MiddlewareUnifiedConfig) GetMode() MiddlewareMode {
	return MiddlewareMode(c.Global.Mode)
}

// IsBasicMode 是否为基础模式
func (c *MiddlewareUnifiedConfig) IsBasicMode() bool {
	return c.GetMode() == BasicMode
}

// IsAdvancedMode 是否为高级模式
func (c *MiddlewareUnifiedConfig) IsAdvancedMode() bool {
	return c.GetMode() == AdvancedMode
}

// IsAutoMode 是否为自动模式
func (c *MiddlewareUnifiedConfig) IsAutoMode() bool {
	return c.GetMode() == AutoMode
}

// GetBuiltinConfig 获取内置中间件配置
func (c *MiddlewareUnifiedConfig) GetBuiltinConfig(name string) interface{} {
	switch name {
	case "logger":
		return c.Builtin.Logger
	case "recovery":
		return c.Builtin.Recovery
	case "cors":
		return c.Builtin.CORS
	case "auth":
		return c.Builtin.Auth
	case "rate_limit":
		return c.Builtin.RateLimit
	case "tracing":
		return c.Builtin.Tracing
	case "request_id":
		return c.Builtin.RequestID
	case "timeout":
		return c.Builtin.Timeout
	case "secure":
		return c.Builtin.Secure
	case "gzip":
		return c.Builtin.GZip
	default:
		return nil
	}
}

// Validate 验证配置的合法性
func (c *MiddlewareUnifiedConfig) Validate() error {
	// 验证模式
	mode := c.GetMode()
	if mode != BasicMode && mode != AdvancedMode && mode != AutoMode {
		return fmt.Errorf("invalid middleware mode: %s", mode)
	}

	// 验证高级模式配置
	if c.IsAdvancedMode() || c.IsAutoMode() {
		if c.Advanced.CacheSize <= 0 {
			return fmt.Errorf("advanced.cache_size must be positive")
		}
		if c.Advanced.MaxConcurrency <= 0 {
			return fmt.Errorf("advanced.max_concurrency must be positive")
		}
	}

	// 验证自动模式配置
	if c.IsAutoMode() {
		if c.Auto.RequestThreshold <= 0 {
			return fmt.Errorf("auto.request_threshold must be positive")
		}
		if c.Auto.PerformanceThreshold <= 0 || c.Auto.PerformanceThreshold > 1 {
			return fmt.Errorf("auto.performance_threshold must be between 0 and 1")
		}
	}

	return nil
}