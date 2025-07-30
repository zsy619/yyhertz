package config

import (
	"fmt"
	"log"
)

// ConfigExample 配置使用示例
func ConfigExample() {
	fmt.Println("=== Viper配置管理示例 ===")
	
	// 1. 基本配置管理器使用
	fmt.Println("\n1. 基本配置管理器使用:")
	cm := NewViperConfigManager()
	
	// 初始化配置
	if err := cm.Initialize(); err != nil {
		log.Fatal("配置初始化失败:", err)
	}
	
	// 获取基本配置值
	fmt.Printf("应用名称: %s\n", cm.GetString("app.name"))
	fmt.Printf("应用版本: %s\n", cm.GetString("app.version"))
	fmt.Printf("服务端口: %d\n", cm.GetInt("app.port"))
	fmt.Printf("调试模式: %v\n", cm.GetBool("app.debug"))
	
	// 2. 获取完整配置结构
	fmt.Println("\n2. 获取完整配置结构:")
	config, err := cm.GetConfig()
	if err != nil {
		log.Fatal("获取配置失败:", err)
	}
	
	fmt.Printf("应用配置:\n")
	fmt.Printf("  名称: %s\n", config.App.Name)
	fmt.Printf("  版本: %s\n", config.App.Version)
	fmt.Printf("  环境: %s\n", config.App.Environment)
	fmt.Printf("  端口: %d\n", config.App.Port)
	fmt.Printf("  调试: %v\n", config.App.Debug)
	
	fmt.Printf("数据库配置:\n")
	fmt.Printf("  驱动: %s\n", config.Database.Driver)
	fmt.Printf("  主机: %s\n", config.Database.Host)
	fmt.Printf("  端口: %d\n", config.Database.Port)
	fmt.Printf("  数据库: %s\n", config.Database.Database)
	
	fmt.Printf("Redis配置:\n")
	fmt.Printf("  主机: %s\n", config.Redis.Host)
	fmt.Printf("  端口: %d\n", config.Redis.Port)
	fmt.Printf("  数据库: %d\n", config.Redis.Database)
	
	// 3. 动态设置和获取配置
	fmt.Println("\n3. 动态设置和获取配置:")
	cm.Set("custom.api_key", "your-api-key-here")
	cm.Set("custom.timeout", 30)
	cm.Set("custom.enabled", true)
	
	fmt.Printf("API密钥: %s\n", cm.GetString("custom.api_key"))
	fmt.Printf("超时时间: %d\n", cm.GetInt("custom.timeout"))
	fmt.Printf("是否启用: %v\n", cm.GetBool("custom.enabled"))
	
	// 4. 检查配置是否存在
	fmt.Println("\n4. 检查配置是否存在:")
	keys := []string{"app.name", "nonexistent.key", "custom.api_key"}
	for _, key := range keys {
		if cm.IsSet(key) {
			fmt.Printf("✅ %s 存在: %v\n", key, cm.Get(key))
		} else {
			fmt.Printf("❌ %s 不存在\n", key)
		}
	}
	
	// 5. 获取所有配置键
	fmt.Println("\n5. 所有配置键（前10个）:")
	allKeys := cm.AllKeys()
	for i, key := range allKeys {
		if i >= 10 {
			fmt.Printf("... 还有 %d 个配置项\n", len(allKeys)-10)
			break
		}
		fmt.Printf("  %s = %v\n", key, cm.Get(key))
	}
	
	// 6. 使用全局便捷函数
	fmt.Println("\n6. 使用全局便捷函数:")
	globalConfig, err := GetGlobalConfig()
	if err != nil {
		log.Fatal("获取全局配置失败:", err)
	}
	
	fmt.Printf("全局配置 - 应用名: %s\n", globalConfig.App.Name)
	fmt.Printf("全局配置 - 端口: %s\n", GetConfigString("app.port"))
	fmt.Printf("全局配置 - 调试: %v\n", GetConfigBool("app.debug"))
	
	// 7. 中间件配置示例
	fmt.Println("\n7. 中间件配置:")
	fmt.Printf("CORS启用: %v\n", config.Middleware.CORS.Enable)
	fmt.Printf("允许源: %v\n", config.Middleware.CORS.AllowOrigins)
	fmt.Printf("限流启用: %v\n", config.Middleware.RateLimit.Enable)
	fmt.Printf("限流策略: %s\n", config.Middleware.RateLimit.Strategy)
	
	// 8. 服务配置示例
	fmt.Println("\n8. 服务配置:")
	fmt.Printf("邮件提供商: %s\n", config.Services.Email.Provider)
	fmt.Printf("存储提供商: %s\n", config.Services.Storage.Provider)
	fmt.Printf("监控启用: %v\n", config.Monitor.Enable)
	
	fmt.Println("\n=== 配置示例完成 ===")
}

// DatabaseConfigExample 数据库配置使用示例
func DatabaseConfigExample() {
	fmt.Println("=== 数据库配置示例 ===")
	
	config, err := GetGlobalConfig()
	if err != nil {
		log.Fatal("获取配置失败:", err)
	}
	
	// 构建数据库连接字符串
	var dsn string
	switch config.Database.Driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
			config.Database.Charset,
		)
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Database.Host,
			config.Database.Port,
			config.Database.Username,
			config.Database.Password,
			config.Database.Database,
			config.Database.SSLMode,
		)
	default:
		log.Fatal("不支持的数据库驱动:", config.Database.Driver)
	}
	
	fmt.Printf("数据库连接字符串: %s\n", dsn)
	fmt.Printf("最大空闲连接: %d\n", config.Database.MaxIdle)
	fmt.Printf("最大打开连接: %d\n", config.Database.MaxOpen)
	fmt.Printf("连接最大生命周期: %d秒\n", config.Database.MaxLife)
}

// RedisConfigExample Redis配置使用示例
func RedisConfigExample() {
	fmt.Println("=== Redis配置示例 ===")
	
	config, err := GetGlobalConfig()
	if err != nil {
		log.Fatal("获取配置失败:", err)
	}
	
	// Redis连接配置
	fmt.Printf("Redis地址: %s:%d\n", config.Redis.Host, config.Redis.Port)
	fmt.Printf("数据库: %d\n", config.Redis.Database)
	fmt.Printf("连接池大小: %d\n", config.Redis.PoolSize)
	fmt.Printf("最小空闲连接: %d\n", config.Redis.MinIdle)
	fmt.Printf("最大重试次数: %d\n", config.Redis.MaxRetries)
	
	if config.Redis.Password != "" {
		fmt.Println("需要密码认证")
	} else {
		fmt.Println("无需密码认证")
	}
}

// LogConfigExample 日志配置使用示例
func LogConfigExample() {
	fmt.Println("=== 日志配置示例 ===")
	
	config, err := GetGlobalConfig()
	if err != nil {
		log.Fatal("获取配置失败:", err)
	}
	
	fmt.Printf("日志级别: %s\n", config.Log.Level)
	fmt.Printf("日志格式: %s\n", config.Log.Format)
	fmt.Printf("控制台输出: %v\n", config.Log.EnableConsole)
	fmt.Printf("文件输出: %v\n", config.Log.EnableFile)
	
	if config.Log.EnableFile {
		fmt.Printf("日志文件: %s\n", config.Log.FilePath)
		fmt.Printf("最大大小: %dMB\n", config.Log.MaxSize)
		fmt.Printf("保留天数: %d天\n", config.Log.MaxAge)
		fmt.Printf("最大备份: %d个\n", config.Log.MaxBackups)
		fmt.Printf("压缩备份: %v\n", config.Log.Compress)
	}
}

// TLSConfigExample TLS配置使用示例
func TLSConfigExample() {
	fmt.Println("=== TLS配置示例 ===")
	
	config, err := GetGlobalConfig()
	if err != nil {
		log.Fatal("获取配置失败:", err)
	}
	
	fmt.Printf("TLS启用: %v\n", config.TLS.Enable)
	
	if config.TLS.Enable {
		fmt.Printf("证书文件: %s\n", config.TLS.CertFile)
		fmt.Printf("私钥文件: %s\n", config.TLS.KeyFile)
		fmt.Printf("最小版本: %s\n", config.TLS.MinVersion)
		fmt.Printf("最大版本: %s\n", config.TLS.MaxVersion)
		fmt.Printf("自动重载: %v\n", config.TLS.AutoReload)
		
		if config.TLS.AutoReload {
			fmt.Printf("重载间隔: %d秒\n", config.TLS.ReloadInterval)
		}
	}
}

// MiddlewareConfigExample 中间件配置使用示例
func MiddlewareConfigExample() {
	fmt.Println("=== 中间件配置示例 ===")
	
	config, err := GetGlobalConfig()
	if err != nil {
		log.Fatal("获取配置失败:", err)
	}
	
	// CORS配置
	fmt.Println("CORS配置:")
	fmt.Printf("  启用: %v\n", config.Middleware.CORS.Enable)
	if config.Middleware.CORS.Enable {
		fmt.Printf("  允许源: %v\n", config.Middleware.CORS.AllowOrigins)
		fmt.Printf("  允许方法: %v\n", config.Middleware.CORS.AllowMethods)
		fmt.Printf("  允许头: %v\n", config.Middleware.CORS.AllowHeaders)
		fmt.Printf("  允许凭证: %v\n", config.Middleware.CORS.AllowCredentials)
		fmt.Printf("  最大年龄: %d秒\n", config.Middleware.CORS.MaxAge)
	}
	
	// 限流配置
	fmt.Println("限流配置:")
	fmt.Printf("  启用: %v\n", config.Middleware.RateLimit.Enable)
	if config.Middleware.RateLimit.Enable {
		fmt.Printf("  速率: %d/秒\n", config.Middleware.RateLimit.Rate)
		fmt.Printf("  突发: %d\n", config.Middleware.RateLimit.Burst)
		fmt.Printf("  策略: %s\n", config.Middleware.RateLimit.Strategy)
	}
	
	// 认证配置
	fmt.Println("认证配置:")
	fmt.Printf("  启用: %v\n", config.Middleware.Auth.Enable)
	if config.Middleware.Auth.Enable {
		fmt.Printf("  Token TTL: %d小时\n", config.Middleware.Auth.TokenTTL)
		fmt.Printf("  刷新 TTL: %d小时\n", config.Middleware.Auth.RefreshTTL)
	}
}

// EnvironmentConfigExample 环境变量配置示例
func EnvironmentConfigExample() {
	fmt.Println("=== 环境变量配置示例 ===")
	
	// 显示如何使用环境变量覆盖配置
	fmt.Println("支持的环境变量前缀: YYHERTZ_")
	fmt.Println("示例环境变量:")
	fmt.Println("  YYHERTZ_APP_NAME=MyApp")
	fmt.Println("  YYHERTZ_APP_PORT=9000")
	fmt.Println("  YYHERTZ_APP_DEBUG=false")
	fmt.Println("  YYHERTZ_DATABASE_HOST=prod-db.example.com")
	fmt.Println("  YYHERTZ_DATABASE_PASSWORD=secret")
	fmt.Println("  YYHERTZ_REDIS_HOST=redis.example.com")
	fmt.Println("  YYHERTZ_LOG_LEVEL=error")
	
	// 获取当前配置值（可能被环境变量覆盖）
	fmt.Println("\n当前有效配置:")
	fmt.Printf("  应用名: %s\n", GetConfigString("app.name"))
	fmt.Printf("  端口: %d\n", GetConfigInt("app.port"))
	fmt.Printf("  数据库主机: %s\n", GetConfigString("database.host"))
	fmt.Printf("  Redis主机: %s\n", GetConfigString("redis.host"))
	fmt.Printf("  日志级别: %s\n", GetConfigString("log.level"))
}