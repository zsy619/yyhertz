package examples

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"

	"github.com/zsy619/yyhertz/framework/config"
)

// 演示如何使用泛型配置管理器

func ExampleAppConfig() {
	fmt.Println("=== 应用配置示例 ===")

	// 方式1：使用便捷函数获取完整配置
	appConfig, err := config.GetAppConfig()
	if err != nil {
		log.Printf("获取应用配置失败: %v", err)
		return
	}

	fmt.Printf("应用名称: %s\n", appConfig.App.Name)
	fmt.Printf("应用端口: %d\n", appConfig.App.Port)
	fmt.Printf("数据库主机: %s\n", appConfig.Database.Host)
	fmt.Printf("数据库端口: %d\n", appConfig.Database.Port)

	// 方式2：使用配置管理器获取单个值
	manager := config.GetAppConfigManager()
	appName := manager.GetString("app.name")
	dbHost := manager.GetString("database.host")
	debugMode := manager.GetBool("app.debug")

	fmt.Printf("管理器方式 - 应用名称: %s\n", appName)
	fmt.Printf("管理器方式 - 数据库主机: %s\n", dbHost)
	fmt.Printf("管理器方式 - 调试模式: %v\n", debugMode)

	// 方式3：使用泛型便捷函数
	corsEnabled := config.GetConfigBool(config.AppConfig{}, "middleware.cors.enable")
	redisPort := config.GetConfigInt(config.AppConfig{}, "redis.port")

	fmt.Printf("泛型方式 - CORS启用: %v\n", corsEnabled)
	fmt.Printf("泛型方式 - Redis端口: %d\n", redisPort)
}

func ExampleTemplateConfig() {
	fmt.Println("\n=== 模板配置示例 ===")

	// 方式1：使用便捷函数获取完整配置
	templateConfig, err := config.GetTemplateConfig()
	if err != nil {
		log.Printf("获取模板配置失败: %v", err)
		return
	}

	fmt.Printf("模板引擎类型: %s\n", templateConfig.Engine.Type)
	fmt.Printf("模板目录: %s\n", templateConfig.Engine.Directory)
	fmt.Printf("模板扩展名: %s\n", templateConfig.Engine.Extension)
	fmt.Printf("启用缓存: %v\n", templateConfig.Cache.Enable)

	// 方式2：使用配置管理器
	manager := config.GetTemplateConfigManager()
	staticRoot := manager.GetString("static.root")
	liveReload := manager.GetBool("development.live_reload")

	fmt.Printf("静态文件根目录: %s\n", staticRoot)
	fmt.Printf("实时重载: %v\n", liveReload)

	// 方式3：使用泛型便捷函数
	cacheType := config.GetConfigString(config.TemplateConfig{}, "cache.type")
	compressHtml := config.GetConfigBool(config.TemplateConfig{}, "render.compress_html")

	fmt.Printf("缓存类型: %s\n", cacheType)
	fmt.Printf("压缩HTML: %v\n", compressHtml)
}

// 自定义配置示例
type CustomDatabaseConfig struct {
	Host         string `mapstructure:"host" yaml:"host"`
	Port         int    `mapstructure:"port" yaml:"port"`
	Username     string `mapstructure:"username" yaml:"username"`
	Password     string `mapstructure:"password" yaml:"password"`
	Database     string `mapstructure:"database" yaml:"database"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`
}

func (c CustomDatabaseConfig) GetConfigName() string {
	return "custom_database"
}

func (c CustomDatabaseConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault("host", "localhost")
	v.SetDefault("port", 3306)
	v.SetDefault("username", "root")
	v.SetDefault("password", "")
	v.SetDefault("database", "myapp")
	v.SetDefault("max_idle_conns", 10)
	v.SetDefault("max_open_conns", 100)
}

func (c CustomDatabaseConfig) GenerateDefaultContent() string {
	return `# 自定义数据库配置
host: "localhost"
port: 3306
username: "root"
password: ""
database: "myapp"
max_idle_conns: 10
max_open_conns: 100
`
}

func ExampleAuthConfig() {
	fmt.Println("\n=== 认证配置示例 ===")

	// 方式1：使用便捷函数获取完整配置
	authConfig, err := config.GetAuthConfig()
	if err != nil {
		log.Printf("获取认证配置失败: %v", err)
		return
	}

	fmt.Printf("CAS启用状态: %v\n", authConfig.CAS.Enabled)
	fmt.Printf("CAS服务器: %s\n", authConfig.CAS.Host)
	fmt.Printf("JWT密钥: %s\n", authConfig.JWT.Secret)
	fmt.Printf("Session名称: %s\n", authConfig.Session.Name)
	fmt.Printf("应用名称: %s\n", authConfig.Application.Name)

	// 方式2：使用配置管理器
	manager := config.GetAuthConfigManager()
	casVersion := manager.GetString("cas.version")
	loginPath := manager.GetString("login_paths.school.path")
	authEnabled := manager.GetBool("authorization.enable")

	fmt.Printf("CAS版本: %s\n", casVersion)
	fmt.Printf("学校登录路径: %s\n", loginPath)
	fmt.Printf("权限控制启用: %v\n", authEnabled)

	// 方式3：使用泛型便捷函数
	jwtTTL := config.GetConfigInt(config.AuthConfig{}, "jwt.token_ttl")
	passwordMinLength := config.GetConfigInt(config.AuthConfig{}, "security.password_policy.min_length")

	fmt.Printf("JWT Token生存时间: %d小时\n", jwtTTL)
	fmt.Printf("密码最小长度: %d\n", passwordMinLength)
}

func ExampleLogConfig() {
	fmt.Println("\n=== 日志配置示例 ===")

	// 方式1：使用便捷函数获取完整配置
	logConfig, err := config.GetLogConfig()
	if err != nil {
		log.Printf("获取日志配置失败: %v", err)
		return
	}

	fmt.Printf("日志级别: %s\n", logConfig.Level)
	fmt.Printf("日志格式: %s\n", logConfig.Format)
	fmt.Printf("启用控制台输出: %v\n", logConfig.EnableConsole)
	fmt.Printf("启用文件输出: %v\n", logConfig.EnableFile)
	fmt.Printf("日志文件路径: %s\n", logConfig.FilePath)
	fmt.Printf("最大文件大小: %d MB\n", logConfig.MaxSize)
	fmt.Printf("文件保留天数: %d天\n", logConfig.MaxAge)

	// 方式2：使用配置管理器
	manager := config.GetLogConfigManager()
	timestampFormat := manager.GetString("timestamp_format")
	showCaller := manager.GetBool("show_caller")
	compress := manager.GetBool("compress")

	fmt.Printf("时间戳格式: %s\n", timestampFormat)
	fmt.Printf("显示调用位置: %v\n", showCaller)
	fmt.Printf("压缩旧日志: %v\n", compress)

	// 方式3：使用泛型便捷函数
	level := config.GetConfigString(config.LogConfig{}, "level")
	enableFile := config.GetConfigBool(config.LogConfig{}, "enable_file")
	maxBackups := config.GetConfigInt(config.LogConfig{}, "max_backups")

	fmt.Printf("泛型方式 - 日志级别: %s\n", level)
	fmt.Printf("泛型方式 - 启用文件: %v\n", enableFile)
	fmt.Printf("泛型方式 - 最大备份数: %d\n", maxBackups)

	// 方式4：使用便捷函数
	logLevel := config.GetLogConfigString("level")
	logFormat := config.GetLogConfigString("format")
	consoleEnabled := config.GetLogConfigBool("enable_console")

	fmt.Printf("便捷函数 - 日志级别: %s\n", logLevel)
	fmt.Printf("便捷函数 - 日志格式: %s\n", logFormat)
	fmt.Printf("便捷函数 - 控制台输出: %v\n", consoleEnabled)
}

func ExampleCustomConfig() {
	fmt.Println("\n=== 自定义配置示例 ===")

	// 使用自定义配置
	manager := config.GetViperConfigManager(CustomDatabaseConfig{})

	// 获取配置值
	host := manager.GetString("host")
	port := manager.GetInt("port")
	username := manager.GetString("username")
	maxIdle := manager.GetInt("max_idle_conns")

	fmt.Printf("数据库主机: %s\n", host)
	fmt.Printf("数据库端口: %d\n", port)
	fmt.Printf("用户名: %s\n", username)
	fmt.Printf("最大空闲连接: %d\n", maxIdle)

	// 设置配置值
	manager.Set("host", "192.168.1.100")
	manager.Set("port", 3307)

	// 获取完整配置
	dbConfig, err := manager.GetConfig()
	if err != nil {
		log.Printf("获取数据库配置失败: %v", err)
		return
	}

	fmt.Printf("完整配置: %+v\n", dbConfig)
}

func ExampleConfigWatching() {
	fmt.Println("\n=== 配置监听示例 ===")

	// 监听应用配置变化
	config.WatchConfig(config.AppConfig{})
	fmt.Println("开始监听应用配置文件变化...")

	// 监听模板配置变化
	config.WatchConfig(config.TemplateConfig{})
	fmt.Println("开始监听模板配置文件变化...")

	// 监听认证配置变化
	config.WatchConfig(config.AuthConfig{})
	fmt.Println("开始监听认证配置文件变化...")

	// 监听日志配置变化
	config.WatchConfig(config.LogConfig{})
	fmt.Println("开始监听日志配置文件变化...")

	// 监听自定义配置变化
	config.WatchConfig(CustomDatabaseConfig{})
	fmt.Println("开始监听自定义数据库配置文件变化...")
}

func ExampleDynamicConfiguration() {
	fmt.Println("\n=== 动态配置示例 ===")

	// 动态设置应用配置
	config.SetConfigValue(config.AppConfig{}, "app.debug", false)
	config.SetConfigValue(config.AppConfig{}, "app.port", 9999)
	config.SetConfigValue(config.AppConfig{}, "database.host", "production-db.example.com")

	// 验证设置的值
	debug := config.GetConfigBool(config.AppConfig{}, "app.debug")
	port := config.GetConfigInt(config.AppConfig{}, "app.port")
	dbHost := config.GetConfigString(config.AppConfig{}, "database.host")

	fmt.Printf("动态设置后 - 调试模式: %v\n", debug)
	fmt.Printf("动态设置后 - 应用端口: %d\n", port)
	fmt.Printf("动态设置后 - 数据库主机: %s\n", dbHost)

	// 动态设置模板配置
	config.SetConfigValue(config.TemplateConfig{}, "engine.type", "pug")
	config.SetConfigValue(config.TemplateConfig{}, "cache.enable", false)
	config.SetConfigValue(config.TemplateConfig{}, "development.live_reload", true)

	engineType := config.GetConfigString(config.TemplateConfig{}, "engine.type")
	cacheEnabled := config.GetConfigBool(config.TemplateConfig{}, "cache.enable")
	liveReload := config.GetConfigBool(config.TemplateConfig{}, "development.live_reload")

	fmt.Printf("动态设置后 - 模板引擎: %s\n", engineType)
	fmt.Printf("动态设置后 - 缓存启用: %v\n", cacheEnabled)
	fmt.Printf("动态设置后 - 实时重载: %v\n", liveReload)

	// 动态设置认证配置
	config.SetConfigValue(config.AuthConfig{}, "cas.enabled", true)
	config.SetConfigValue(config.AuthConfig{}, "cas.host", "https://sso.production.com")
	config.SetConfigValue(config.AuthConfig{}, "jwt.token_ttl", 72)
	config.SetConfigValue(config.AuthConfig{}, "application.environment", "production")

	casEnabled := config.GetConfigBool(config.AuthConfig{}, "cas.enabled")
	casHost := config.GetConfigString(config.AuthConfig{}, "cas.host")
	jwtTTL := config.GetConfigInt(config.AuthConfig{}, "jwt.token_ttl")
	authEnv := config.GetConfigString(config.AuthConfig{}, "application.environment")

	fmt.Printf("动态设置后 - CAS启用: %v\n", casEnabled)
	fmt.Printf("动态设置后 - CAS主机: %s\n", casHost)
	fmt.Printf("动态设置后 - JWT生存时间: %d小时\n", jwtTTL)
	fmt.Printf("动态设置后 - 认证环境: %s\n", authEnv)

	// 动态设置日志配置
	config.SetConfigValue(config.LogConfig{}, "level", "debug")
	config.SetConfigValue(config.LogConfig{}, "format", "text")
	config.SetConfigValue(config.LogConfig{}, "enable_console", true)
	config.SetConfigValue(config.LogConfig{}, "max_size", 200)

	logLevel := config.GetConfigString(config.LogConfig{}, "level")
	logFormat := config.GetConfigString(config.LogConfig{}, "format")
	enableConsole := config.GetConfigBool(config.LogConfig{}, "enable_console")
	maxSize := config.GetConfigInt(config.LogConfig{}, "max_size")

	fmt.Printf("动态设置后 - 日志级别: %s\n", logLevel)
	fmt.Printf("动态设置后 - 日志格式: %s\n", logFormat)
	fmt.Printf("动态设置后 - 控制台输出: %v\n", enableConsole)
	fmt.Printf("动态设置后 - 最大文件大小: %d MB\n", maxSize)
}

// ExampleGetStringSlice 演示GetConfigStringSlice的使用
func ExampleGetStringSlice() {
	fmt.Println("\n=== GetConfigStringSlice 示例 ===")

	// 设置一些字符串切片配置值用于测试
	config.SetConfigValue(config.AppConfig{}, "middleware.cors.allowed_origins", []string{
		"http://localhost:3000",
		"https://example.com",
		"https://api.example.com",
	})

	config.SetConfigValue(config.AppConfig{}, "security.allowed_hosts", []string{
		"localhost",
		"127.0.0.1",
		"example.com",
	})

	// 使用泛型GetConfigStringSlice获取字符串切片
	corsOrigins := config.GetConfigStringSlice(config.AppConfig{}, "middleware.cors.allowed_origins")
	allowedHosts := config.GetConfigStringSlice(config.AppConfig{}, "security.allowed_hosts")
	
	// 使用便捷函数GetAppConfigStringSlice
	emptySlice := config.GetAppConfigStringSlice("non.existent.key")

	fmt.Printf("CORS允许的源: %v\n", corsOrigins)
	fmt.Printf("安全允许的主机: %v\n", allowedHosts)
	fmt.Printf("不存在的键: %v (长度: %d)\n", emptySlice, len(emptySlice))

	// 演示对不同配置类型的使用
	config.SetConfigValue(config.TemplateConfig{}, "engine.supported_extensions", []string{
		".html",
		".htm", 
		".tmpl",
		".tpl",
	})

	extensions := config.GetConfigStringSlice(config.TemplateConfig{}, "engine.supported_extensions")
	fmt.Printf("模板支持的扩展名: %v\n", extensions)

	// 设置认证配置的字符串切片
	config.SetConfigValue(config.AuthConfig{}, "oauth.scopes", []string{
		"read",
		"write", 
		"admin",
	})

	scopes := config.GetConfigStringSlice(config.AuthConfig{}, "oauth.scopes")
	fmt.Printf("OAuth作用域: %v\n", scopes)

	// 演示其他类型的切片
	config.SetConfigValue(config.AppConfig{}, "performance.timeout_values", []int{
		30, 60, 120, 300,
	})
	timeouts := config.GetConfigIntSlice(config.AppConfig{}, "performance.timeout_values")
	fmt.Printf("超时值列表: %v\n", timeouts)

	config.SetConfigValue(config.AppConfig{}, "feature.flags", []bool{
		true, false, true, false,
	})
	flags := config.GetConfigBoolSlice(config.AppConfig{}, "feature.flags")
	fmt.Printf("功能标志: %v\n", flags)

	config.SetConfigValue(config.AppConfig{}, "metrics.ratios", []float64{
		0.1, 0.25, 0.5, 0.75, 1.0,
	})
	ratios := config.GetConfigFloat64Slice(config.AppConfig{}, "metrics.ratios")
	fmt.Printf("指标比率: %v\n", ratios)

	config.SetConfigValue(config.AppConfig{}, "cache.durations", []string{
		"1m", "5m", "15m", "1h",
	})
	durations := config.GetConfigDurationSlice(config.AppConfig{}, "cache.durations")
	fmt.Printf("缓存持续时间: %v\n", durations)
}

// ExampleConfigWithDefaults 演示带默认值的配置获取
func ExampleConfigWithDefaults() {
	fmt.Println("\n=== 配置默认值示例 ===")

	// 演示直接方式 - 带默认值的配置获取
	fmt.Println("1. 直接方式 - 带默认值")
	
	// 获取不存在的配置，使用默认值
	appName := config.GetConfigString(config.AppConfig{}, "app.unknown_name", "DefaultApp")
	port := config.GetConfigInt(config.AppConfig{}, "app.unknown_port", 8080)
	debugMode := config.GetConfigBool(config.AppConfig{}, "app.unknown_debug", true)
	
	fmt.Printf("应用名称: %s (默认值)\n", appName)
	fmt.Printf("端口: %d (默认值)\n", port)
	fmt.Printf("调试模式: %v (默认值)\n", debugMode)
	
	// 演示切片默认值
	defaultHosts := []string{"localhost", "127.0.0.1"}
	hosts := config.GetConfigStringSlice(config.AppConfig{}, "security.unknown_hosts", defaultHosts)
	fmt.Printf("允许的主机: %v (默认值)\n", hosts)
	
	defaultPorts := []int{80, 443, 8080}
	ports := config.GetConfigIntSlice(config.AppConfig{}, "server.unknown_ports", defaultPorts)
	fmt.Printf("监听端口: %v (默认值)\n", ports)
	
	// 演示反射方式 - 带默认值的配置获取
	fmt.Println("\n2. 反射方式 - 带默认值")
	
	// 使用反射方式获取配置
	reflectAppName := config.GetConfigStringWithDefaults[config.AppConfig]("app.reflect_name", "ReflectApp")
	reflectPort := config.GetConfigIntWithDefaults[config.AppConfig]("app.reflect_port", 9090)
	reflectDebug := config.GetConfigBoolWithDefaults[config.AppConfig]("app.reflect_debug", false)
	
	fmt.Printf("反射 - 应用名称: %s (默认值)\n", reflectAppName)
	fmt.Printf("反射 - 端口: %d (默认值)\n", reflectPort)
	fmt.Printf("反射 - 调试模式: %v (默认值)\n", reflectDebug)
	
	// 反射方式的切片默认值
	reflectHosts := config.GetConfigStringSliceWithDefaults[config.AppConfig]("security.reflect_hosts", defaultHosts)
	fmt.Printf("反射 - 允许的主机: %v (默认值)\n", reflectHosts)
	
	reflectPorts := config.GetConfigIntSliceWithDefaults[config.AppConfig]("server.reflect_ports", defaultPorts)
	fmt.Printf("反射 - 监听端口: %v (默认值)\n", reflectPorts)
	
	// 演示浮点数和时间间隔的默认值
	ratio := config.GetConfigFloat64WithDefaults[config.AppConfig]("performance.unknown_ratio", 0.85)
	fmt.Printf("反射 - 性能比率: %.2f (默认值)\n", ratio)
	
	timeout := config.GetConfigDurationWithDefaults[config.AppConfig]("performance.unknown_timeout", 30*time.Second)
	fmt.Printf("反射 - 超时时间: %v (默认值)\n", timeout)
	
	// 演示已存在配置值的获取
	fmt.Println("\n3. 已存在配置值的获取")
	
	// 设置一些配置值
	config.SetConfigValue(config.AppConfig{}, "test.existing_string", "ExistingValue")
	config.SetConfigValue(config.AppConfig{}, "test.existing_int", 42)
	config.SetConfigValue(config.AppConfig{}, "test.existing_bool", true)
	
	// 获取已存在的配置（不会使用默认值）
	existingString := config.GetConfigString(config.AppConfig{}, "test.existing_string", "DefaultValue")
	existingInt := config.GetConfigInt(config.AppConfig{}, "test.existing_int", 999)
	existingBool := config.GetConfigBool(config.AppConfig{}, "test.existing_bool", false)
	
	fmt.Printf("已存在字符串: %s (实际值)\n", existingString)
	fmt.Printf("已存在整数: %d (实际值)\n", existingInt)
	fmt.Printf("已存在布尔值: %v (实际值)\n", existingBool)
}

// RunAllExamples 运行所有示例
func RunAllExamples() {
	fmt.Println("泛型配置管理器使用示例")
	fmt.Println("========================================")

	ExampleAppConfig()
	ExampleTemplateConfig()
	ExampleAuthConfig()
	ExampleLogConfig()
	ExampleCustomConfig()
	ExampleConfigWatching()
	ExampleDynamicConfiguration()
	ExampleGetStringSlice()
	ExampleConfigWithDefaults()

	fmt.Println("\n========================================")
	fmt.Println("所有示例执行完成!")
}
