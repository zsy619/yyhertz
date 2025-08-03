// Package orm 提供基于GORM的数据库ORM集成
//
// 这个包封装了GORM，提供了数据库连接管理、模型基类、事务管理等功能
// 类似于Beego的ORM功能，但基于现代化的GORM v2
package orm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zsy619/yyhertz/framework/config"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type         string `json:"type" yaml:"type"`                     // 数据库类型: mysql, postgres, sqlite
	Host         string `json:"host" yaml:"host"`                     // 主机地址
	Port         int    `json:"port" yaml:"port"`                     // 端口
	Username     string `json:"username" yaml:"username"`             // 用户名
	Password     string `json:"password" yaml:"password"`             // 密码
	Database     string `json:"database" yaml:"database"`             // 数据库名
	Charset      string `json:"charset" yaml:"charset"`               // 字符集
	Timezone     string `json:"timezone" yaml:"timezone"`             // 时区
	MaxIdleConns int    `json:"max_idle_conns" yaml:"max_idle_conns"` // 最大空闲连接数
	MaxOpenConns int    `json:"max_open_conns" yaml:"max_open_conns"` // 最大打开连接数
	MaxLifetime  int    `json:"max_lifetime" yaml:"max_lifetime"`     // 连接最大生存时间(秒)
	LogLevel     string `json:"log_level" yaml:"log_level"`           // 日志级别
	SlowQuery    int    `json:"slow_query" yaml:"slow_query"`         // 慢查询阈值(毫秒)
}

// DefaultDatabaseConfig 默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:         "sqlite",
		Host:         "localhost",
		Port:         3306,
		Username:     "root",
		Password:     "",
		Database:     "app.db",
		Charset:      "utf8mb4",
		Timezone:     "Asia/Shanghai",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  3600,
		LogLevel:     "info",
		SlowQuery:    1000,
	}
}

// ORM ORM管理器
type ORM struct {
	db               *gorm.DB
	config           *DatabaseConfig
	pool             ConnectionPool // 连接池
	mutex            sync.RWMutex
	migrator         gorm.Migrator
	logger           logger.Interface
	metricsCollector *MetricsCollector // 指标收集器
}

var (
	defaultORM *ORM
	once       sync.Once
	ormMutex   sync.Mutex
)

// GetDefaultORM 获取默认ORM实例
func GetDefaultORM() *ORM {
	once.Do(func() {
		ormMutex.Lock()
		defer ormMutex.Unlock()

		// 从配置管理器获取数据库配置
		dbConfig := DefaultDatabaseConfig()

		// 尝试从全局配置获取数据库配置
		if configManager := config.GetAppConfigManager(); configManager != nil {
			if appConfig, err := configManager.GetConfig(); err == nil {
				// 映射配置字段
				if appConfig.Database.Driver != "" {
					dbConfig.Type = appConfig.Database.Driver
				}
				if appConfig.Database.Host != "" {
					dbConfig.Host = appConfig.Database.Host
				}
				if appConfig.Database.Port > 0 {
					dbConfig.Port = appConfig.Database.Port
				}
				if appConfig.Database.Username != "" {
					dbConfig.Username = appConfig.Database.Username
				}
				if appConfig.Database.Password != "" {
					dbConfig.Password = appConfig.Database.Password
				}
				if appConfig.Database.Database != "" {
					dbConfig.Database = appConfig.Database.Database
				}
				if appConfig.Database.Charset != "" {
					dbConfig.Charset = appConfig.Database.Charset
				}
				if appConfig.Database.MaxIdle > 0 {
					dbConfig.MaxIdleConns = appConfig.Database.MaxIdle
				}
				if appConfig.Database.MaxOpen > 0 {
					dbConfig.MaxOpenConns = appConfig.Database.MaxOpen
				}
				if appConfig.Database.MaxLife > 0 {
					dbConfig.MaxLifetime = appConfig.Database.MaxLife
				}
			}
		}

		var err error
		defaultORM, err = NewORM(dbConfig)
		if err != nil {
			config.Fatalf("Failed to initialize default ORM: %v", err)
		}
	})
	return defaultORM
}

// NewORM 创建新的ORM实例
func NewORM(dbConfig *DatabaseConfig) (*ORM, error) {
	if dbConfig == nil {
		dbConfig = DefaultDatabaseConfig()
	}

	// 创建GORM配置
	gormConfig := &gorm.Config{
		Logger: newGormLogger(dbConfig.LogLevel, time.Duration(dbConfig.SlowQuery)*time.Millisecond),
	}

	// 根据数据库类型创建连接
	var db *gorm.DB
	var err error

	switch dbConfig.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
			dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port,
			dbConfig.Database, dbConfig.Charset, dbConfig.Timezone)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
			dbConfig.Host, dbConfig.Username, dbConfig.Password, dbConfig.Database, dbConfig.Port, dbConfig.Timezone)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dbConfig.Database), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
		sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime) * time.Second)
	}

	orm := &ORM{
		db:       db,
		config:   dbConfig,
		migrator: db.Migrator(),
		logger:   gormConfig.Logger,
	}

	config.Infof("ORM initialized successfully with %s database", dbConfig.Type)
	return orm, nil
}

// DB 获取底层GORM数据库实例
func (o *ORM) DB() *gorm.DB {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.db
}

// Config 获取数据库配置
func (o *ORM) Config() *DatabaseConfig {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.config
}

// Migrator 获取迁移器
func (o *ORM) Migrator() gorm.Migrator {
	return o.migrator
}

// Close 关闭数据库连接
func (o *ORM) Close() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if sqlDB, err := o.db.DB(); err == nil {
		return sqlDB.Close()
	}
	return nil
}

// Ping 检查数据库连接
func (o *ORM) Ping() error {
	if sqlDB, err := o.db.DB(); err == nil {
		return sqlDB.Ping()
	} else {
		return err
	}
}

// Transaction 执行事务
func (o *ORM) Transaction(fn func(tx *gorm.DB) error) error {
	return o.db.Transaction(fn)
}

// WithContext 使用上下文
func (o *ORM) WithContext(ctx context.Context) *gorm.DB {
	return o.db.WithContext(ctx)
}

// AutoMigrate 自动迁移模型
func (o *ORM) AutoMigrate(models ...any) error {
	config.Info("Starting database auto migration...")

	for _, model := range models {
		modelName := fmt.Sprintf("%T", model)
		config.Infof("Migrating model: %s", modelName)

		if err := o.db.AutoMigrate(model); err != nil {
			config.Errorf("Failed to migrate model %s: %v", modelName, err)
			return fmt.Errorf("failed to migrate model %s: %w", modelName, err)
		}

		config.Infof("Model %s migrated successfully", modelName)
	}

	config.Info("Database auto migration completed")
	return nil
}

// GetStats 获取数据库连接统计信息
func (o *ORM) GetStats() map[string]any {
	stats := make(map[string]any)

	if sqlDB, err := o.db.DB(); err == nil {
		dbStats := sqlDB.Stats()
		stats["max_open_connections"] = dbStats.MaxOpenConnections
		stats["open_connections"] = dbStats.OpenConnections
		stats["in_use"] = dbStats.InUse
		stats["idle"] = dbStats.Idle
		stats["wait_count"] = dbStats.WaitCount
		stats["wait_duration"] = dbStats.WaitDuration.String()
		stats["max_idle_closed"] = dbStats.MaxIdleClosed
		stats["max_idle_time_closed"] = dbStats.MaxIdleTimeClosed
		stats["max_lifetime_closed"] = dbStats.MaxLifetimeClosed
	}

	stats["database_type"] = o.config.Type
	stats["database_name"] = o.config.Database

	return stats
}

// newGormLogger 创建GORM日志器
func newGormLogger(level string, slowThreshold time.Duration) logger.Interface {
	var logLevel logger.LogLevel

	switch level {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Info
	}

	return logger.New(
		&gormLogWriter{}, // 使用自定义的日志写入器
		logger.Config{
			SlowThreshold:             slowThreshold,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

// gormLogWriter GORM日志写入器，适配框架日志系统
type gormLogWriter struct{}

// Printf 实现logger.Writer接口
func (w *gormLogWriter) Printf(format string, args ...any) {
	config.Infof(format, args...)
}

// ============= 便捷方法 =============

// Create 创建记录
func Create(value any) *gorm.DB {
	return GetDefaultORM().DB().Create(value)
}

// Save 保存记录
func Save(value any) *gorm.DB {
	return GetDefaultORM().DB().Save(value)
}

// First 查询第一条记录
func First(dest any, conds ...any) *gorm.DB {
	return GetDefaultORM().DB().First(dest, conds...)
}

// Find 查询多条记录
func Find(dest any, conds ...any) *gorm.DB {
	return GetDefaultORM().DB().Find(dest, conds...)
}

// Update 更新记录
func Update(column string, value any) *gorm.DB {
	return GetDefaultORM().DB().Update(column, value)
}

// Updates 批量更新记录
func Updates(values any) *gorm.DB {
	return GetDefaultORM().DB().Updates(values)
}

// Delete 删除记录
func Delete(value any, conds ...any) *gorm.DB {
	return GetDefaultORM().DB().Delete(value, conds...)
}

// Where 添加条件
func Where(query any, args ...any) *gorm.DB {
	return GetDefaultORM().DB().Where(query, args...)
}

// Raw 执行原生SQL
func Raw(sql string, values ...any) *gorm.DB {
	return GetDefaultORM().DB().Raw(sql, values...)
}

// Exec 执行SQL
func Exec(sql string, values ...any) *gorm.DB {
	return GetDefaultORM().DB().Exec(sql, values...)
}
