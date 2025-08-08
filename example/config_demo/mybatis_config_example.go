package main

import (
	"fmt"
	"log"

	"github.com/zsy619/yyhertz/framework/config"
)

// ConfigExample 演示如何使用MyBatis配置
func ConfigExample() {
	fmt.Println("=== MyBatis 配置使用示例 ===")

	// 方法1: 直接获取完整配置
	mybatisConfig, err := config.GetMyBatisConfig()
	if err != nil {
		log.Fatalf("获取MyBatis配置失败: %v", err)
	}

	fmt.Printf("MyBatis启用状态: %t\n", mybatisConfig.Basic.Enable)
	fmt.Printf("配置文件路径: %s\n", mybatisConfig.Basic.ConfigFile)
	fmt.Printf("Mapper文件位置: %s\n", mybatisConfig.Basic.MapperLocations)
	fmt.Printf("缓存启用: %t\n", mybatisConfig.Cache.Enable)
	fmt.Printf("日志级别: %s\n", mybatisConfig.Logging.Level)

	// 方法2: 使用配置管理器获取特定值
	manager := config.GetMyBatisConfigManager()

	fmt.Println("\n=== 使用配置管理器获取特定值 ===")
	fmt.Printf("基础启用状态: %t\n", manager.GetBool("basic.enable"))
	fmt.Printf("数据源URL: %s\n", manager.GetString("datasource.url"))
	fmt.Printf("连接池最大空闲连接: %d\n", manager.GetInt("pool.max_idle_conns"))
	fmt.Printf("缓存TTL: %d秒\n", manager.GetInt("cache.ttl"))
	fmt.Printf("慢查询阈值: %d毫秒\n", manager.GetInt("logging.slow_query"))
	fmt.Printf("事务超时: %d秒\n", manager.GetInt("transaction.default_timeout"))

	// 获取拦截器列表
	interceptors := manager.GetStringSlice("plugins.interceptors")
	fmt.Printf("拦截器列表: %v\n", interceptors)

	// 方法3: 使用泛型配置函数
	fmt.Println("\n=== 使用泛型配置函数 ===")
	enable := config.GetConfigBool(config.MyBatisConfig{}, "basic.enable", false)
	fmt.Printf("MyBatis启用(带默认值): %t\n", enable)

	driver := config.GetConfigString(config.MyBatisConfig{}, "datasource.driver", "postgresql")
	fmt.Printf("数据库驱动(带默认值): %s\n", driver)

	maxConns := config.GetConfigInt(config.MyBatisConfig{}, "pool.max_open_conns", 20)
	fmt.Printf("最大连接数(带默认值): %d\n", maxConns)

	fmt.Println("\n=== 配置文件信息 ===")
	fmt.Printf("当前使用的配置文件: %s\n", manager.ConfigFileUsed())
}

// CacheConfigExample 演示缓存配置的使用
func CacheConfigExample() {
	fmt.Println("\n=== 缓存配置示例 ===")

	config, err := config.GetMyBatisConfig()
	if err != nil {
		log.Printf("获取配置失败: %v", err)
		return
	}

	if config.Cache.Enable {
		fmt.Printf("缓存已启用:\n")
		fmt.Printf("  类型: %s\n", config.Cache.Type)
		fmt.Printf("  TTL: %d秒\n", config.Cache.TTL)
		fmt.Printf("  最大条目数: %d\n", config.Cache.MaxSize)

		if config.Cache.Type == "redis" {
			fmt.Printf("  Redis配置:\n")
			fmt.Printf("    地址: %s\n", config.Cache.RedisAddr)
			fmt.Printf("    数据库: %d\n", config.Cache.RedisDB)
		}
	} else {
		fmt.Println("缓存未启用")
	}
}

// LoggingConfigExample 演示日志配置的使用
func LoggingConfigExample() {
	fmt.Println("\n=== 日志配置示例 ===")

	config, err := config.GetMyBatisConfig()
	if err != nil {
		log.Printf("获取配置失败: %v", err)
		return
	}

	if config.Logging.Enable {
		fmt.Printf("SQL日志已启用:\n")
		fmt.Printf("  日志级别: %s\n", config.Logging.Level)
		fmt.Printf("  显示SQL: %t\n", config.Logging.ShowSQL)
		fmt.Printf("  慢查询阈值: %d毫秒\n", config.Logging.SlowQuery)
	} else {
		fmt.Println("SQL日志未启用")
	}
}

func main() {
	ConfigExample()
	CacheConfigExample()
	LoggingConfigExample()
}
