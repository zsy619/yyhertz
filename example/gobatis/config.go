// Package main MyBatis配置和初始化
//
// 提供MyBatis框架的配置管理：
// 1. 数据源配置
// 2. 缓存配置 
// 3. 性能监控配置
// 4. 日志配置
// 5. 连接池配置
// 6. 事务配置
// 7. XML映射配置
// 8. 初始化工具函数
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `json:"driver" yaml:"driver"`             // mysql, postgres, sqlite
	DSN             string `json:"dsn" yaml:"dsn"`                   // 数据源连接字符串
	MaxOpenConns    int    `json:"maxOpenConns" yaml:"maxOpenConns"` // 最大打开连接数
	MaxIdleConns    int    `json:"maxIdleConns" yaml:"maxIdleConns"` // 最大空闲连接数
	ConnMaxLifetime int    `json:"connMaxLifetime" yaml:"connMaxLifetime"` // 连接最大生存时间(秒)
}

// CacheConfigOptions 缓存配置选项
type CacheConfigOptions struct {
	L1Enabled       bool          `json:"l1Enabled" yaml:"l1Enabled"`             // 启用一级缓存
	L1Size          int           `json:"l1Size" yaml:"l1Size"`                   // 一级缓存大小
	L1TTL           time.Duration `json:"l1TTL" yaml:"l1TTL"`                     // 一级缓存TTL
	L2Enabled       bool          `json:"l2Enabled" yaml:"l2Enabled"`             // 启用二级缓存
	L2Size          int           `json:"l2Size" yaml:"l2Size"`                   // 二级缓存大小
	L2TTL           time.Duration `json:"l2TTL" yaml:"l2TTL"`                     // 二级缓存TTL
	CleanupInterval time.Duration `json:"cleanupInterval" yaml:"cleanupInterval"` // 清理间隔
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `json:"level" yaml:"level"`       // debug, info, warn, error
	ShowSQL  bool   `json:"showSQL" yaml:"showSQL"`   // 显示SQL语句
	SlowTime int    `json:"slowTime" yaml:"slowTime"` // 慢查询阈值(毫秒)
}

// MyBatisConfig MyBatis完整配置
type MyBatisConfig struct {
	Database DatabaseConfig     `json:"database" yaml:"database"`
	Cache    CacheConfigOptions `json:"cache" yaml:"cache"`
	Log      LogConfig          `json:"log" yaml:"log"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *MyBatisConfig {
	return &MyBatisConfig{
		Database: DatabaseConfig{
			Driver:          "sqlite",
			DSN:             ":memory:",
			MaxOpenConns:    100,
			MaxIdleConns:    10,
			ConnMaxLifetime: 3600, // 1小时
		},
		Cache: CacheConfigOptions{
			L1Enabled:       true,
			L1Size:          1000,
			L1TTL:           10 * time.Minute,
			L2Enabled:       false,
			L2Size:          5000,
			L2TTL:           30 * time.Minute,
			CleanupInterval: 5 * time.Minute,
		},
		Log: LogConfig{
			Level:    "info",
			ShowSQL:  true,
			SlowTime: 200, // 200ms
		},
	}
}

// MySQLConfig 返回MySQL配置模板
func MySQLConfig(host, user, password, dbname string) *MyBatisConfig {
	config := DefaultConfig()
	config.Database.Driver = "mysql"
	config.Database.DSN = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, dbname)
	return config
}

// PostgreSQLConfig 返回PostgreSQL配置模板
func PostgreSQLConfig(host, user, password, dbname string, port int) *MyBatisConfig {
	config := DefaultConfig()
	config.Database.Driver = "postgres"
	config.Database.DSN = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		host, user, password, dbname, port)
	return config
}

// SQLiteConfig 返回SQLite配置模板
func SQLiteConfig(filepath string) *MyBatisConfig {
	config := DefaultConfig()
	config.Database.Driver = "sqlite"
	config.Database.DSN = filepath
	return config
}

// InitializeDatabase 根据配置初始化数据库连接
func InitializeDatabase(config *DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.Driver {
	case "mysql":
		dialector = mysql.Open(config.DSN)
	case "postgres":
		dialector = postgres.Open(config.DSN)
	case "sqlite":
		dialector = sqlite.Open(config.DSN)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Driver)
	}

	// 配置日志级别
	logLevel := logger.Info
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	return db, nil
}

// CreateSimpleSession 创建SimpleSession
func CreateSimpleSession(config *MyBatisConfig) (mybatis.SimpleSession, error) {
	db, err := InitializeDatabase(&config.Database)
	if err != nil {
		return nil, err
	}

	session := mybatis.NewSimpleSession(db)

	// 配置缓存
	if config.Cache.L1Enabled || config.Cache.L2Enabled {
		cacheConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  config.Cache.L1Enabled,
			L1CacheSize:     config.Cache.L1Size,
			L1CacheTTL:      config.Cache.L1TTL,
			L2CacheEnabled:  config.Cache.L2Enabled,
			L2CacheSize:     config.Cache.L2Size,
			L2CacheTTL:      config.Cache.L2TTL,
			CleanupInterval: config.Cache.CleanupInterval,
		}
		session = session.EnableCache(cacheConfig)
	}

	// 配置调试模式
	debug := config.Log.ShowSQL && config.Log.Level == "debug"
	session = session.Debug(debug)

	return session, nil
}

// CreateXMLSession 创建XMLSession
func CreateXMLSession(config *MyBatisConfig) (mybatis.XMLSession, error) {
	db, err := InitializeDatabase(&config.Database)
	if err != nil {
		return nil, err
	}

	session := mybatis.NewXMLMapper(db)

	// 配置缓存
	if config.Cache.L1Enabled || config.Cache.L2Enabled {
		cacheConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  config.Cache.L1Enabled,
			L1CacheSize:     config.Cache.L1Size,
			L1CacheTTL:      config.Cache.L1TTL,
			L2CacheEnabled:  config.Cache.L2Enabled,
			L2CacheSize:     config.Cache.L2Size,
			L2CacheTTL:      config.Cache.L2TTL,
			CleanupInterval: config.Cache.CleanupInterval,
		}
		session = session.EnableCache(cacheConfig)
	}

	// 配置调试模式
	debug := config.Log.ShowSQL && config.Log.Level == "debug"
	session = session.Debug(debug)

	return session, nil
}

// InitializeTestDatabase 初始化测试数据库（所有测试文件共用）
func InitializeTestDatabase() (*gorm.DB, error) {
	config := DefaultConfig()
	
	// 使用内存SQLite数据库进行测试
	config.Database.Driver = "sqlite"
	config.Database.DSN = ":memory:"
	
	db, err := InitializeDatabase(&config.Database)
	if err != nil {
		return nil, err
	}

	// 创建测试表结构
	err = createTestTables(db)
	if err != nil {
		return nil, fmt.Errorf("创建测试表失败: %w", err)
	}

	return db, nil
}

// createTestTables 创建测试用的数据库表
func createTestTables(db *gorm.DB) error {
	// 用户表
	err := db.Exec(`
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    age INTEGER DEFAULT 0,
    status TEXT DEFAULT 'active',
    department_id INTEGER DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`).Error
	if err != nil {
		return fmt.Errorf("创建users表失败: %w", err)
	}

	// 用户资料表
	err = db.Exec(`
CREATE TABLE IF NOT EXISTS user_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    avatar TEXT DEFAULT '',
    bio TEXT DEFAULT '',
    phone TEXT DEFAULT '',
    address TEXT DEFAULT '',
    FOREIGN KEY (user_id) REFERENCES users(id)
);`).Error
	if err != nil {
		return fmt.Errorf("创建user_profiles表失败: %w", err)
	}

	// 部门表
	err = db.Exec(`
CREATE TABLE IF NOT EXISTS departments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    code TEXT UNIQUE NOT NULL,
    manager TEXT DEFAULT ''
);`).Error
	if err != nil {
		return fmt.Errorf("创建departments表失败: %w", err)
	}

	// 订单表
	err = db.Exec(`
CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    order_no TEXT UNIQUE NOT NULL,
    amount DECIMAL(10,2) DEFAULT 0.00,
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);`).Error
	if err != nil {
		return fmt.Errorf("创建orders表失败: %w", err)
	}

	// 订单项目表
	err = db.Exec(`
CREATE TABLE IF NOT EXISTS order_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    product_name TEXT NOT NULL,
    quantity INTEGER DEFAULT 1,
    price DECIMAL(10,2) DEFAULT 0.00,
    FOREIGN KEY (order_id) REFERENCES orders(id)
);`).Error
	if err != nil {
		return fmt.Errorf("创建order_items表失败: %w", err)
	}

	log.Println("成功创建所有测试表")
	return nil
}

// CleanupDatabase 清理测试数据库
func CleanupDatabase(db *gorm.DB) error {
	tables := []string{"order_items", "orders", "user_profiles", "departments", "users"}
	
	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		if err != nil {
			log.Printf("清理表 %s 失败: %v", table, err)
		}
	}
	
	return nil
}

// ValidateConfig 验证配置的有效性
func ValidateConfig(config *MyBatisConfig) error {
	if config.Database.DSN == "" {
		return fmt.Errorf("数据库连接字符串不能为空")
	}
	
	if config.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("最大连接数必须大于0")
	}
	
	if config.Database.MaxIdleConns <= 0 {
		return fmt.Errorf("最大空闲连接数必须大于0")
	}
	
	if config.Cache.L1Size <= 0 && config.Cache.L1Enabled {
		return fmt.Errorf("启用一级缓存时，缓存大小必须大于0")
	}
	
	if config.Cache.L2Size <= 0 && config.Cache.L2Enabled {
		return fmt.Errorf("启用二级缓存时，缓存大小必须大于0")
	}
	
	return nil
}

// LogConfigInfo 输出配置信息
func LogConfigInfo(config *MyBatisConfig) {
	log.Printf("MyBatis配置信息:")
	log.Printf("  数据库驱动: %s", config.Database.Driver)
	log.Printf("  最大连接数: %d", config.Database.MaxOpenConns)
	log.Printf("  最大空闲连接数: %d", config.Database.MaxIdleConns)
	log.Printf("  连接最大生存时间: %ds", config.Database.ConnMaxLifetime)
	log.Printf("  一级缓存: %v (大小: %d, TTL: %v)", config.Cache.L1Enabled, config.Cache.L1Size, config.Cache.L1TTL)
	log.Printf("  二级缓存: %v (大小: %d, TTL: %v)", config.Cache.L2Enabled, config.Cache.L2Size, config.Cache.L2TTL)
	log.Printf("  日志级别: %s", config.Log.Level)
	log.Printf("  显示SQL: %v", config.Log.ShowSQL)
	log.Printf("  慢查询阈值: %dms", config.Log.SlowTime)
}