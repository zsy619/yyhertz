package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// MVCConfigName MVC配置名称常量
const MVCConfigName = "mvc"

// MVCConfig MVC框架配置结构
type MVCConfig struct {
	// 中间件配置
	Middleware struct {
		EnableOptimization    bool          `mapstructure:"enable_optimization" yaml:"enable_optimization" json:"enable_optimization"`
		CompileOnStartup      bool          `mapstructure:"compile_on_startup" yaml:"compile_on_startup" json:"compile_on_startup"`
		PrecompileChains      bool          `mapstructure:"precompile_chains" yaml:"precompile_chains" json:"precompile_chains"`
		CacheSize             int           `mapstructure:"cache_size" yaml:"cache_size" json:"cache_size"`
		CompileTimeout        time.Duration `mapstructure:"compile_timeout" yaml:"compile_timeout" json:"compile_timeout"`
		MaxConcurrency        int           `mapstructure:"max_concurrency" yaml:"max_concurrency" json:"max_concurrency"`
		EnableDependencyAnalysis bool       `mapstructure:"enable_dependency_analysis" yaml:"enable_dependency_analysis" json:"enable_dependency_analysis"`
	} `mapstructure:"middleware" yaml:"middleware" json:"middleware"`

	// 错误处理配置
	ErrorHandling struct {
		EnableIntelligent       bool          `mapstructure:"enable_intelligent" yaml:"enable_intelligent" json:"enable_intelligent"`
		EnableAutoRecovery      bool          `mapstructure:"enable_auto_recovery" yaml:"enable_auto_recovery" json:"enable_auto_recovery"`
		EnableClassification    bool          `mapstructure:"enable_classification" yaml:"enable_classification" json:"enable_classification"`
		EnableCircuitBreaker    bool          `mapstructure:"enable_circuit_breaker" yaml:"enable_circuit_breaker" json:"enable_circuit_breaker"`
		RetryMaxCount           int           `mapstructure:"retry_max_count" yaml:"retry_max_count" json:"retry_max_count"`
		RetryInterval           time.Duration `mapstructure:"retry_interval" yaml:"retry_interval" json:"retry_interval"`
		CircuitBreakerThreshold int           `mapstructure:"circuit_breaker_threshold" yaml:"circuit_breaker_threshold" json:"circuit_breaker_threshold"`
		CircuitBreakerTimeout   time.Duration `mapstructure:"circuit_breaker_timeout" yaml:"circuit_breaker_timeout" json:"circuit_breaker_timeout"`
		EnableStackTrace        bool          `mapstructure:"enable_stack_trace" yaml:"enable_stack_trace" json:"enable_stack_trace"`
	} `mapstructure:"error_handling" yaml:"error_handling" json:"error_handling"`

	// 性能监控配置
	Performance struct {
		EnableMonitoring        bool          `mapstructure:"enable_monitoring" yaml:"enable_monitoring" json:"enable_monitoring"`
		EnableStatistics        bool          `mapstructure:"enable_statistics" yaml:"enable_statistics" json:"enable_statistics"`
		StatsReportInterval     time.Duration `mapstructure:"stats_report_interval" yaml:"stats_report_interval" json:"stats_report_interval"`
		MetricsEndpoint         string        `mapstructure:"metrics_endpoint" yaml:"metrics_endpoint" json:"metrics_endpoint"`
		EnableProfiler          bool          `mapstructure:"enable_profiler" yaml:"enable_profiler" json:"enable_profiler"`
		ProfilerEndpoint        string        `mapstructure:"profiler_endpoint" yaml:"profiler_endpoint" json:"profiler_endpoint"`
		HealthCheckInterval     time.Duration `mapstructure:"health_check_interval" yaml:"health_check_interval" json:"health_check_interval"`
		MaxConcurrentRecoveries int           `mapstructure:"max_concurrent_recoveries" yaml:"max_concurrent_recoveries" json:"max_concurrent_recoveries"`
	} `mapstructure:"performance" yaml:"performance" json:"performance"`

	// 调试配置
	Debug struct {
		EnableMode          bool   `mapstructure:"enable_mode" yaml:"enable_mode" json:"enable_mode"`
		PrintMiddleware     bool   `mapstructure:"print_middleware" yaml:"print_middleware" json:"print_middleware"`
		PrintError          bool   `mapstructure:"print_error" yaml:"print_error" json:"print_error"`
		PrintStatistics     bool   `mapstructure:"print_statistics" yaml:"print_statistics" json:"print_statistics"`
		LogLevel            string `mapstructure:"log_level" yaml:"log_level" json:"log_level"`
		EnableTrace         bool   `mapstructure:"enable_trace" yaml:"enable_trace" json:"enable_trace"`
		ShowRequestDetails  bool   `mapstructure:"show_request_details" yaml:"show_request_details" json:"show_request_details"`
	} `mapstructure:"debug" yaml:"debug" json:"debug"`

	// 上下文配置
	Context struct {
		EnablePooling       bool          `mapstructure:"enable_pooling" yaml:"enable_pooling" json:"enable_pooling"`
		PoolSize            int           `mapstructure:"pool_size" yaml:"pool_size" json:"pool_size"`
		MaxPoolSize         int           `mapstructure:"max_pool_size" yaml:"max_pool_size" json:"max_pool_size"`
		PoolTimeout         time.Duration `mapstructure:"pool_timeout" yaml:"pool_timeout" json:"pool_timeout"`
		EnableBatchRelease  bool          `mapstructure:"enable_batch_release" yaml:"enable_batch_release" json:"enable_batch_release"`
		BatchSize           int           `mapstructure:"batch_size" yaml:"batch_size" json:"batch_size"`
	} `mapstructure:"context" yaml:"context" json:"context"`

	// 路由配置
	Router struct {
		EnableCaching       bool          `mapstructure:"enable_caching" yaml:"enable_caching" json:"enable_caching"`
		CacheSize           int           `mapstructure:"cache_size" yaml:"cache_size" json:"cache_size"`
		CacheTimeout        time.Duration `mapstructure:"cache_timeout" yaml:"cache_timeout" json:"cache_timeout"`
		EnableCompression   bool          `mapstructure:"enable_compression" yaml:"enable_compression" json:"enable_compression"`
		MaxParamCount       int           `mapstructure:"max_param_count" yaml:"max_param_count" json:"max_param_count"`
		EnableRegexOptim    bool          `mapstructure:"enable_regex_optim" yaml:"enable_regex_optim" json:"enable_regex_optim"`
	} `mapstructure:"router" yaml:"router" json:"router"`
}

// GetConfigName 实现 ConfigInterface 接口
func (c MVCConfig) GetConfigName() string {
	return MVCConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c MVCConfig) SetDefaults(v *viper.Viper) {
	// 中间件默认配置
	v.SetDefault("middleware.enable_optimization", true)
	v.SetDefault("middleware.compile_on_startup", true)
	v.SetDefault("middleware.precompile_chains", true)
	v.SetDefault("middleware.cache_size", 500)
	v.SetDefault("middleware.compile_timeout", "30s")
	v.SetDefault("middleware.max_concurrency", 100)
	v.SetDefault("middleware.enable_dependency_analysis", true)

	// 错误处理默认配置
	v.SetDefault("error_handling.enable_intelligent", true)
	v.SetDefault("error_handling.enable_auto_recovery", true)
	v.SetDefault("error_handling.enable_classification", true)
	v.SetDefault("error_handling.enable_circuit_breaker", true)
	v.SetDefault("error_handling.retry_max_count", 3)
	v.SetDefault("error_handling.retry_interval", "1s")
	v.SetDefault("error_handling.circuit_breaker_threshold", 10)
	v.SetDefault("error_handling.circuit_breaker_timeout", "30s")
	v.SetDefault("error_handling.enable_stack_trace", true)

	// 性能监控默认配置
	v.SetDefault("performance.enable_monitoring", true)
	v.SetDefault("performance.enable_statistics", true)
	v.SetDefault("performance.stats_report_interval", "5m")
	v.SetDefault("performance.metrics_endpoint", "/metrics")
	v.SetDefault("performance.enable_profiler", false)
	v.SetDefault("performance.profiler_endpoint", "/debug/pprof")
	v.SetDefault("performance.health_check_interval", "1m")
	v.SetDefault("performance.max_concurrent_recoveries", 100)

	// 调试默认配置
	v.SetDefault("debug.enable_mode", false)
	v.SetDefault("debug.print_middleware", false)
	v.SetDefault("debug.print_error", false)
	v.SetDefault("debug.print_statistics", false)
	v.SetDefault("debug.log_level", "info")
	v.SetDefault("debug.enable_trace", false)
	v.SetDefault("debug.show_request_details", false)

	// 上下文默认配置
	v.SetDefault("context.enable_pooling", true)
	v.SetDefault("context.pool_size", 1000)
	v.SetDefault("context.max_pool_size", 10000)
	v.SetDefault("context.pool_timeout", "10s")
	v.SetDefault("context.enable_batch_release", true)
	v.SetDefault("context.batch_size", 100)

	// 路由默认配置
	v.SetDefault("router.enable_caching", true)
	v.SetDefault("router.cache_size", 1000)
	v.SetDefault("router.cache_timeout", "1h")
	v.SetDefault("router.enable_compression", true)
	v.SetDefault("router.max_param_count", 50)
	v.SetDefault("router.enable_regex_optim", true)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c MVCConfig) GenerateDefaultContent() string {
	return `# YYHertz MVC Framework Configuration
# MVC框架高级配置

# 中间件配置
middleware:
  enable_optimization: true        # 启用中间件优化
  compile_on_startup: true        # 启动时预编译中间件
  precompile_chains: true         # 预编译常用中间件链
  cache_size: 500                 # 中间件缓存大小
  compile_timeout: "30s"          # 编译超时时间
  max_concurrency: 100            # 最大并发处理数
  enable_dependency_analysis: true # 启用依赖分析

# 错误处理配置
error_handling:
  enable_intelligent: true        # 启用智能错误处理
  enable_auto_recovery: true      # 启用自动恢复
  enable_classification: true     # 启用错误分类
  enable_circuit_breaker: true    # 启用熔断器
  retry_max_count: 3              # 最大重试次数
  retry_interval: "1s"            # 重试间隔
  circuit_breaker_threshold: 10   # 熔断器阈值
  circuit_breaker_timeout: "30s"  # 熔断器超时时间
  enable_stack_trace: true        # 启用堆栈跟踪

# 性能监控配置
performance:
  enable_monitoring: true         # 启用性能监控
  enable_statistics: true         # 启用统计信息
  stats_report_interval: "5m"     # 统计报告间隔
  metrics_endpoint: "/metrics"    # 监控指标端点
  enable_profiler: false          # 启用性能分析器
  profiler_endpoint: "/debug/pprof" # 性能分析端点
  health_check_interval: "1m"     # 健康检查间隔
  max_concurrent_recoveries: 100  # 最大并发恢复数

# 调试配置
debug:
  enable_mode: false              # 启用调试模式
  print_middleware: false         # 打印中间件信息
  print_error: false              # 打印错误信息
  print_statistics: false         # 打印统计信息
  log_level: "info"               # 日志级别 (debug, info, warn, error)
  enable_trace: false             # 启用调用链跟踪
  show_request_details: false     # 显示请求详细信息

# 上下文配置
context:
  enable_pooling: true            # 启用对象池化
  pool_size: 1000                 # 池大小
  max_pool_size: 10000            # 最大池大小
  pool_timeout: "10s"             # 池超时时间
  enable_batch_release: true      # 启用批量释放
  batch_size: 100                 # 批处理大小

# 路由配置
router:
  enable_caching: true            # 启用路由缓存
  cache_size: 1000                # 路由缓存大小
  cache_timeout: "1h"             # 缓存超时时间
  enable_compression: true        # 启用压缩
  max_param_count: 50             # 最大参数数量
  enable_regex_optim: true        # 启用正则优化

# 环境特定配置示例:
# development:
#   debug:
#     enable_mode: true
#     print_middleware: true
#     print_error: true
#     log_level: "debug"
#
# production:
#   debug:
#     enable_mode: false
#     log_level: "warn"
#   performance:
#     stats_report_interval: "30m"
#
# testing:
#   debug:
#     log_level: "debug"
#   error_handling:
#     retry_max_count: 1
`
}

// GetDevelopmentConfig 获取开发环境配置
func GetDevelopmentConfig() MVCConfig {
	config := GetDefaultMVCConfig()
	
	// 开发环境特定配置
	config.Debug.EnableMode = true
	config.Debug.PrintMiddleware = true
	config.Debug.PrintError = true
	config.Debug.PrintStatistics = true
	config.Debug.LogLevel = "debug"
	config.Debug.EnableTrace = true
	config.Debug.ShowRequestDetails = true
	
	config.Performance.StatsReportInterval = time.Minute
	
	return config
}

// GetProductionConfig 获取生产环境配置
func GetProductionConfig() MVCConfig {
	config := GetDefaultMVCConfig()
	
	// 生产环境特定配置
	config.Debug.EnableMode = false
	config.Debug.PrintMiddleware = false
	config.Debug.PrintError = false
	config.Debug.PrintStatistics = false
	config.Debug.LogLevel = "warn"
	config.Debug.EnableTrace = false
	config.Debug.ShowRequestDetails = false
	
	config.Performance.StatsReportInterval = 30 * time.Minute
	config.Performance.EnableProfiler = false
	
	// 生产环境优化配置
	config.Middleware.CacheSize = 1000
	config.Context.PoolSize = 2000
	config.Router.CacheSize = 2000
	
	return config
}

// GetTestingConfig 获取测试环境配置
func GetTestingConfig() MVCConfig {
	config := GetDefaultMVCConfig()
	
	// 测试环境特定配置
	config.Debug.LogLevel = "debug"
	config.ErrorHandling.RetryMaxCount = 1
	config.Performance.StatsReportInterval = 10 * time.Second
	
	// 测试环境快速配置
	config.Middleware.CompileTimeout = 5 * time.Second
	config.Context.PoolTimeout = 1 * time.Second
	
	return config
}

// GetDefaultMVCConfig 获取默认MVC配置
func GetDefaultMVCConfig() MVCConfig {
	config := MVCConfig{}
	
	// 中间件默认配置
	config.Middleware.EnableOptimization = true
	config.Middleware.CompileOnStartup = true
	config.Middleware.PrecompileChains = true
	config.Middleware.CacheSize = 500
	config.Middleware.CompileTimeout = 30 * time.Second
	config.Middleware.MaxConcurrency = 100
	config.Middleware.EnableDependencyAnalysis = true

	// 错误处理默认配置
	config.ErrorHandling.EnableIntelligent = true
	config.ErrorHandling.EnableAutoRecovery = true
	config.ErrorHandling.EnableClassification = true
	config.ErrorHandling.EnableCircuitBreaker = true
	config.ErrorHandling.RetryMaxCount = 3
	config.ErrorHandling.RetryInterval = time.Second
	config.ErrorHandling.CircuitBreakerThreshold = 10
	config.ErrorHandling.CircuitBreakerTimeout = 30 * time.Second
	config.ErrorHandling.EnableStackTrace = true

	// 性能监控默认配置
	config.Performance.EnableMonitoring = true
	config.Performance.EnableStatistics = true
	config.Performance.StatsReportInterval = 5 * time.Minute
	config.Performance.MetricsEndpoint = "/metrics"
	config.Performance.EnableProfiler = false
	config.Performance.ProfilerEndpoint = "/debug/pprof"
	config.Performance.HealthCheckInterval = time.Minute
	config.Performance.MaxConcurrentRecoveries = 100

	// 调试默认配置
	config.Debug.EnableMode = false
	config.Debug.PrintMiddleware = false
	config.Debug.PrintError = false
	config.Debug.PrintStatistics = false
	config.Debug.LogLevel = "info"
	config.Debug.EnableTrace = false
	config.Debug.ShowRequestDetails = false

	// 上下文默认配置
	config.Context.EnablePooling = true
	config.Context.PoolSize = 1000
	config.Context.MaxPoolSize = 10000
	config.Context.PoolTimeout = 10 * time.Second
	config.Context.EnableBatchRelease = true
	config.Context.BatchSize = 100

	// 路由默认配置
	config.Router.EnableCaching = true
	config.Router.CacheSize = 1000
	config.Router.CacheTimeout = time.Hour
	config.Router.EnableCompression = true
	config.Router.MaxParamCount = 50
	config.Router.EnableRegexOptim = true
	
	return config
}

// IsDebugMode 是否为调试模式
func (c *MVCConfig) IsDebugMode() bool {
	return c.Debug.EnableMode
}

// IsProductionMode 是否为生产模式
func (c *MVCConfig) IsProductionMode() bool {
	return !c.Debug.EnableMode && c.Debug.LogLevel != "debug"
}

// GetRetryConfig 获取重试配置
func (c *MVCConfig) GetRetryConfig() (maxCount int, interval time.Duration) {
	return c.ErrorHandling.RetryMaxCount, c.ErrorHandling.RetryInterval
}

// GetCircuitBreakerConfig 获取熔断器配置
func (c *MVCConfig) GetCircuitBreakerConfig() (threshold int, timeout time.Duration) {
	return c.ErrorHandling.CircuitBreakerThreshold, c.ErrorHandling.CircuitBreakerTimeout
}

// GetPoolConfig 获取池配置
func (c *MVCConfig) GetPoolConfig() (poolSize int, maxPoolSize int, timeout time.Duration) {
	return c.Context.PoolSize, c.Context.MaxPoolSize, c.Context.PoolTimeout
}

// Validate 验证配置的合法性
func (c *MVCConfig) Validate() error {
	if c.Middleware.CacheSize <= 0 {
		return fmt.Errorf("middleware cache size must be positive")
	}
	
	if c.ErrorHandling.RetryMaxCount < 0 {
		return fmt.Errorf("retry max count cannot be negative")
	}
	
	if c.Performance.MaxConcurrentRecoveries <= 0 {
		return fmt.Errorf("max concurrent recoveries must be positive")
	}
	
	if c.Context.PoolSize <= 0 {
		return fmt.Errorf("context pool size must be positive")
	}
	
	return nil
}