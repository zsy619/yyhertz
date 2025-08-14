package devtools

import (
	"database/sql"
	"fmt"
	"time"

	logconfig "github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc"
)

// DevToolsConfig 开发工具配置
type DevToolsConfig struct {
	// 基础配置
	Enabled     bool   `json:"enabled"`     // 是否启用
	Environment string `json:"environment"` // 环境：development/staging/production

	// 功能开关
	EnableDebug       bool `json:"enable_debug"`        // 启用调试中间件
	EnableHealthCheck bool `json:"enable_health_check"` // 启用健康检查
	EnableQPS         bool `json:"enable_qps"`          // 启用QPS监控
	EnableMetrics     bool `json:"enable_metrics"`      // 启用指标收集(包含性能和运行时监控)
	EnableProfiler    bool `json:"enable_profiler"`     // 启用性能分析
	EnableRateLimit   bool `json:"enable_rate_limit"`   // 启用限流
	EnableHotReload   bool `json:"enable_hot_reload"`   // 启用热重载
	EnableCache       bool `json:"enable_cache"`        // 启用缓存监控
	EnableDatabase    bool `json:"enable_database"`     // 启用数据库监控
	EnableSecurity    bool `json:"enable_security"`     // 启用安全监控

	// 数据库连接（用于健康检查）
	Database *sql.DB `json:"-"`

	// Redis连接信息（用于健康检查）
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`

	// 外部服务URL（用于健康检查）
	ExternalServices []string `json:"external_services"`

	// 限流规则
	RateLimitRules []*RateLimit `json:"rate_limit_rules"`

	// 白名单IP
	WhitelistIPs []string `json:"whitelist_ips"`
}

// DefaultDevToolsConfig 默认开发工具配置
func DefaultDevToolsConfig() *DevToolsConfig {
	return &DevToolsConfig{
		Enabled:           true,
		Environment:       "development",
		EnableDebug:       true,
		EnableHealthCheck: true,
		EnableQPS:         true,
		EnableMetrics:     true,
		EnableProfiler:    true,
		EnableRateLimit:   true, // 默认启用限流
		EnableHotReload:   true,
		EnableCache:       true, // 默认启用缓存监控
		EnableDatabase:    true, // 默认启用数据库监控
		EnableSecurity:    true, // 默认启用安全监控
		ExternalServices:  []string{},
		RateLimitRules: []*RateLimit{
			{
				Rate:      100,
				Burst:     200,
				Strategy:  LimitStrategyTokenBucket,
				Dimension: LimitDimensionGlobal,
				Enabled:   true,
			},
		},
		WhitelistIPs: []string{"127.0.0.1", "::1"},
	}
}

// SetupDevTools 设置开发工具
func SetupDevTools(app *mvc.App, config *DevToolsConfig) error {
	if config == nil {
		config = DefaultDevToolsConfig()
	}

	if !config.Enabled {
		return nil
	}

	// 1. 设置调试中间件
	var debugMiddleware *DebugMiddleware
	var debugPanel *DebugPanel
	if config.EnableDebug {
		debugMiddleware = NewDebugMiddleware()
		debugPanel = NewDebugPanel(debugMiddleware)
		app.Use(debugMiddleware.Handler())
	}

	// 2. 设置健康检查
	var healthMiddleware *HealthCheckMiddleware
	var healthPanel *HealthCheckPanel
	if config.EnableHealthCheck {
		healthConfig := &HealthCheckConfig{
			Enabled:        true,
			CheckInterval:  30 * time.Second,
			DefaultTimeout: 5 * time.Second,
			CacheExpiry:    10 * time.Second,
			Version:        "1.0.0",
			Environment:    config.Environment,
		}
		healthMiddleware = NewHealthCheckMiddleware(healthConfig)
		healthPanel = NewHealthCheckPanel(healthMiddleware)

		// 添加系统检查器
		healthMiddleware.AddChecker(&SystemResourceChecker{})

		// 添加数据库检查器
		if config.Database != nil {
			dbChecker := CreateMySQLChecker("database", config.Database)
			healthMiddleware.AddChecker(dbChecker)
		}

		// 添加Redis检查器
		if config.RedisAddr != "" {
			var redisChecker *RedisChecker
			if config.RedisPassword != "" {
				redisChecker = NewRedisChecker("redis", config.RedisAddr,
					WithRedisPassword(config.RedisPassword))
			} else {
				redisChecker = CreateRedisChecker("redis", config.RedisAddr)
			}
			healthMiddleware.AddChecker(redisChecker)
		}

		// 添加外部服务检查器
		for i, serviceURL := range config.ExternalServices {
			serviceChecker := CreateHTTPChecker(
				fmt.Sprintf("external_service_%d", i+1),
				serviceURL,
			)
			healthMiddleware.AddChecker(serviceChecker)
		}

		// 添加磁盘检查器
		diskChecker := CreateDiskChecker("disk_space", "/")
		healthMiddleware.AddChecker(diskChecker)

		app.Use(healthMiddleware.Handler())
	}

	// 3. 性能监控功能已集成到Metrics收集器中，无需单独设置

	// 4. 设置QPS监控
	var qpsMonitor *QPSMonitor
	var qpsPanel *QPSPanel
	if config.EnableQPS {
		qpsConfig := &QPSConfig{
			Enabled:             true,
			WindowSize:          60 * time.Second,
			MaxHistorySize:      1000,
			QPSThreshold:        1000,
			ConcurrentThreshold: 1000,
		}
		qpsMonitor = NewQPSMonitor(qpsConfig)
		qpsPanel = NewQPSPanel(qpsMonitor)
		app.Use(qpsMonitor.Handler())
	}

	// 5. 设置指标收集
	var metricsCollector *MetricsCollector
	var metricsPanel *MetricsPanel
	if config.EnableMetrics {
		metricsConfig := &MetricsConfig{
			Enabled:       true,
			Namespace:     "yyhertz",
			Subsystem:     "http",
			IncludeSystem: true,
		}
		metricsCollector = NewMetricsCollector(metricsConfig)
		metricsPanel = NewMetricsPanel(metricsCollector)
		app.Use(metricsCollector.Handler())

		// 启动性能监控收集(集成了原performance_monitor功能)
		metricsCollector.StartPerformanceCollection()

		// 启动运行时监控收集(集成了原runtime_metrics功能)
		metricsCollector.StartRuntimeCollection()
	}

	// 6. 设置性能分析
	var profiler *Profiler
	var profilePanel *ProfilePanel
	if config.EnableProfiler {
		profilerConfig := &ProfilerConfig{
			Enabled:         true,
			MaxProfiles:     10,
			AutoProfile:     false,
			ProfileInterval: 5 * time.Minute,
		}
		profiler = NewProfiler(profilerConfig)
		profilePanel = NewProfilePanel(profiler)
		app.Use(profiler.Handler())
	}

	// 7. 设置限流
	var rateLimiter *RateLimiter
	var rateLimitPanel *RateLimitPanel
	if config.EnableRateLimit {
		rateLimiterConfig := &RateLimiterConfig{
			Enabled:   true,
			Rules:     config.RateLimitRules,
			Whitelist: config.WhitelistIPs,
		}
		rateLimiter = NewRateLimiter(rateLimiterConfig)
		rateLimitPanel = NewRateLimitPanel(rateLimiter)
		app.Use(rateLimiter.Handler())
	}

	// 8. 设置热重载
	var hotReloader *HotReloadServer

	// 9. 设置缓存监控
	var cacheCollector *CacheMetricsCollector
	var cachePanel *CacheMetricsPanel

	// 10. 设置数据库监控
	var databaseCollector *DatabaseMetricsCollector
	var databasePanel *DatabaseMetricsPanel

	// 11. 设置安全监控
	var securityCollector *SecurityMetricsCollector
	var securityPanel *SecurityMetricsPanel

	if config.EnableCache {
		cacheConfig := &CacheMetricsConfig{
			Enabled:               true,
			CacheType:             CacheTypeRedis,
			CollectInterval:       5 * time.Second,
			HotKeyThreshold:       100,
			MaxHotKeys:            20,
			EnableKeyTracking:     true,
			EnableLatencyTracking: true,
		}
		cacheCollector = NewCacheMetricsCollector(cacheConfig)
		cachePanel = NewCacheMetricsPanel(cacheCollector)
		cacheCollector.Start()
	}

	if config.EnableDatabase && config.Database != nil {
		dbConfig := &DatabaseMetricsConfig{
			Enabled:              true,
			DatabaseType:         DatabaseTypeMySQL,
			SlowQueryThreshold:   100 * time.Millisecond,
			MaxSlowQueryExamples: 100,
			CollectInterval:      5 * time.Second,
			EnableStackTrace:     false,
			EnableQueryAnalysis:  true,
		}
		databaseCollector = NewDatabaseMetricsCollector(config.Database, dbConfig)
		databasePanel = NewDatabaseMetricsPanel(databaseCollector)
		databaseCollector.Start()
	}

	if config.EnableSecurity {
		securityConfig := &SecurityMetricsConfig{
			Enabled:                     true,
			EnableSQLInjectionDetection: true,
			EnableXSSDetection:          true,
			EnableBruteForceDetection:   true,
			MaxFailedLogins:             5,
			BruteForceWindow:            15 * time.Minute,
			IPLockDuration:              1 * time.Hour,
			UserLockDuration:            30 * time.Minute,
			MaxSecurityEvents:           10000,
			MaxIPThreatInfo:             5000,
			CleanupInterval:             1 * time.Hour,
			EnableAlerts:                true,
			AlertThreshold:              10,
		}
		securityCollector = NewSecurityMetricsCollector(securityConfig)
		securityPanel = NewSecurityMetricsPanel(securityCollector)
		app.Use(securityCollector.Handler())
		securityCollector.Start()
	}

	// 8. 设置热重载
	if config.EnableHotReload && config.Environment == "development" {
		hotReloadConfig := DefaultHotReloadConfig()
		hotReloadConfig.OnReload = func() error {
			logconfig.Debug("执行热重载...")
			return nil
		}

		var err error
		hotReloader, err = NewHotReloadServer(app, hotReloadConfig)
		if err != nil {
			logconfig.Warnf("热重载初始化失败: %v", err)
		} else {
			go func() {
				if err := hotReloader.Run(); err != nil {
					logconfig.Errorf("热重载服务器错误: %v", err)
				}
			}()
		}
	}

	// 注册所有面板路由
	engine := app.Engine
	if debugPanel != nil {
		debugPanel.RegisterRoutes(engine)
	}
	if healthPanel != nil {
		healthPanel.RegisterRoutes(engine)
	}
	// 性能监控路由已集成到metrics面板中
	if qpsPanel != nil {
		qpsPanel.RegisterRoutes(engine)
	}
	if metricsPanel != nil {
		metricsPanel.RegisterRoutes(engine)
	}
	if profilePanel != nil {
		profilePanel.RegisterRoutes(engine)
	}
	if rateLimitPanel != nil {
		rateLimitPanel.RegisterRoutes(engine)
	}
	if cachePanel != nil {
		cachePanel.RegisterRoutes(engine)
	}
	if databasePanel != nil {
		databasePanel.RegisterRoutes(engine)
	}
	if securityPanel != nil {
		securityPanel.RegisterRoutes(engine)
	}

	// 输出启用的功能
	logconfig.Debug("YYHertz 开发工具已启用:")
	if config.EnableDebug {
		logconfig.Debug("- 调试面板: http://localhost/yyhertz/debug/panel")
	}
	if config.EnableHealthCheck {
		logconfig.Debug("- 健康检查: http://localhost/yyhertz/health/panel")
	}
	if config.EnableMetrics {
		logconfig.Debug("- 性能监控: http://localhost/yyhertz/metrics/performance")
		logconfig.Debug("- 运行时监控: http://localhost/yyhertz/metrics/runtime")
	}
	if config.EnableQPS {
		logconfig.Debug("- QPS监控: http://localhost/yyhertz/qps/panel")
	}
	if config.EnableMetrics {
		logconfig.Debug("- 指标收集: http://localhost/yyhertz/metrics/panel")
		logconfig.Debug("- Prometheus: http://localhost/yyhertz/metrics/prometheus")
	}
	if config.EnableProfiler {
		logconfig.Debug("- 性能分析: http://localhost/yyhertz/profile/panel")
		logconfig.Debug("- PProf: http://localhost/yyhertz/profile/debug/pprof/")
	}
	if config.EnableRateLimit {
		logconfig.Debug("- 限流管理: http://localhost/yyhertz/ratelimit/panel")
	}
	if config.EnableCache {
		logconfig.Debug("- 缓存监控: http://localhost/yyhertz/cache/panel")
	}
	if config.EnableDatabase {
		logconfig.Debug("- 数据库监控: http://localhost/yyhertz/database/panel")
	}
	if config.EnableSecurity {
		logconfig.Debug("- 安全监控: http://localhost/yyhertz/security/panel")
	}
	if config.EnableHotReload && config.Environment == "development" {
		logconfig.Debug("- 热重载: 已启用文件监控")
	}

	return nil
}

// SetupBasicDevTools 设置基础开发工具（简化版本）
func SetupBasicDevTools(app *mvc.App) error {
	config := &DevToolsConfig{
		Enabled:           true,
		Environment:       "development",
		EnableDebug:       true,
		EnableHealthCheck: true,
		EnableQPS:         false,
		EnableMetrics:     false,
		EnableProfiler:    false,
		EnableRateLimit:   false,
		EnableHotReload:   true,
		EnableCache:       true,  // 基础配置也启用缓存监控
		EnableDatabase:    false, // 基础配置不启用数据库监控（需要数据库连接）
		EnableSecurity:    true,  // 基础配置启用安全监控
	}

	return SetupDevTools(app, config)
}

// SetupProductionDevTools 设置生产环境开发工具
func SetupProductionDevTools(app *mvc.App, db *sql.DB, redisAddr string) error {
	config := &DevToolsConfig{
		Enabled:           true,
		Environment:       "production",
		EnableDebug:       false, // 生产环境不启用调试
		EnableHealthCheck: true,
		EnableQPS:         true,
		EnableMetrics:     true,
		EnableProfiler:    true,
		EnableRateLimit:   true,  // 生产环境启用限流
		EnableHotReload:   false, // 生产环境不启用热重载
		EnableCache:       true,  // 生产环境启用缓存监控
		EnableDatabase:    true,  // 生产环境启用数据库监控
		EnableSecurity:    true,  // 生产环境启用安全监控
		Database:          db,
		RedisAddr:         redisAddr,
		RateLimitRules: []*RateLimit{
			// 全局限流：每秒1000请求
			{
				Rate:      1000,
				Burst:     2000,
				Strategy:  LimitStrategyTokenBucket,
				Dimension: LimitDimensionGlobal,
				Enabled:   true,
			},
			// IP限流：每秒100请求
			{
				Rate:      100,
				Burst:     200,
				Strategy:  LimitStrategyTokenBucket,
				Dimension: LimitDimensionIP,
				Enabled:   true,
			},
		},
		WhitelistIPs: []string{"127.0.0.1", "::1"},
	}

	return SetupDevTools(app, config)
}

// isDevelopment 检查是否为开发环境
func isDevelopment() bool {
	// 这里可以根据环境变量或配置文件判断
	// 简单示例，实际项目中应该有更完善的环境判断
	return true
}
