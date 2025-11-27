// Package mybatis 提供MyBatis风格的ORM框架
//
// MyBatis-Go 是一个受Java MyBatis启发的轻量级ORM框架
// 核心特色：
// 1. SQL与代码分离（XML映射文件支持）
// 2. 强大的动态SQL构建
// 3. 灵活的结果映射
// 4. 类型安全的参数映射
// 5. 多级缓存机制
// 6. 简洁的Go风格API
//
// 使用示例:
//	// 基础使用
//	session := mybatis.NewSimpleSession(db)
//	user, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
//
//	// XML映射器使用
//	xmlSession := mybatis.NewXMLSession(db)
//	err = xmlSession.LoadMapperXML("user_mapper.xml")
//	users, err := xmlSession.SelectListByID(ctx, "UserMapper.selectByStatus", "active")
package mybatis

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MyBatis 框架主类 - 简化版本，作为工厂类使用
type MyBatis struct {
	db     *gorm.DB
	config *Config
}

// Config MyBatis配置
type Config struct {
	// 缓存配置
	CacheEnabled bool
	CacheSize    int

	// 映射配置
	MapUnderscoreToCamelCase bool

	// 调试配置
	EnableDebug bool
	DryRun      bool

	// 性能配置
	SlowQueryThreshold time.Duration

	// 映射器配置
	MapperLocations []string
}

// NewMyBatis 创建MyBatis实例
func NewMyBatis(db *gorm.DB, config *Config) *MyBatis {
	if config == nil {
		config = DefaultConfig()
	}

	return &MyBatis{
		db:     db,
		config: config,
	}
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		CacheEnabled:             true,
		CacheSize:                1000,
		MapUnderscoreToCamelCase: true,
		EnableDebug:              false,
		DryRun:                   false,
		SlowQueryThreshold:       100 * time.Millisecond,
		MapperLocations:          []string{},
	}
}

// NewSimpleSession 创建简化版会话
func (mb *MyBatis) NewSimpleSession() SimpleSession {
	session := NewSimpleSession(mb.db)

	// 应用配置
	if mb.config.EnableDebug {
		session = session.Debug(true)
	}
	if mb.config.DryRun {
		session = session.DryRun(true)
	}

	// 添加性能监控钩子
	if mb.config.SlowQueryThreshold > 0 {
		beforeHook, afterHook := PerformanceHook(mb.config.SlowQueryThreshold)
		session = session.AddBeforeHook(beforeHook).AddAfterHook(afterHook)
	}

	return session
}

// NewXMLSession 创建支持XML的会话
func (mb *MyBatis) NewXMLSession() XMLSession {
	session := NewXMLSession(mb.db)

	// 应用配置
	if mb.config.EnableDebug {
		session = session.Debug(true)
	}
	if mb.config.DryRun {
		session = session.DryRun(true)
	}

	// 添加性能监控钩子
	if mb.config.SlowQueryThreshold > 0 {
		beforeHook, afterHook := PerformanceHook(mb.config.SlowQueryThreshold)
		session = session.AddBeforeHook(beforeHook).AddAfterHook(afterHook)
	}

	// 自动加载映射器
	for _, location := range mb.config.MapperLocations {
		if err := session.LoadMapperXML(location); err != nil {
			// 记录错误但不中断
			fmt.Printf("Warning: failed to load mapper %s: %v\n", location, err)
		}
	}

	return session
}

// NewTransactionSession 创建支持事务的会话
func (mb *MyBatis) NewTransactionSession() *TransactionAwareSession {
	return NewTransactionAwareSession(mb.db)
}

// 全局便捷函数

var defaultMyBatis *MyBatis

// SetDefault 设置默认MyBatis实例
func SetDefault(mb *MyBatis) {
	defaultMyBatis = mb
}

// GetDefault 获取默认MyBatis实例
func GetDefault() *MyBatis {
	return defaultMyBatis
}

// 使用默认实例的便捷方法

// NewSimpleSessionDefault 使用默认配置创建简化会话
func NewSimpleSessionDefault(db *gorm.DB) SimpleSession {
	return NewMyBatis(db, DefaultConfig()).NewSimpleSession()
}

// NewXMLSessionDefault 使用默认配置创建XML会话
func NewXMLSessionDefault(db *gorm.DB) XMLSession {
	return NewMyBatis(db, DefaultConfig()).NewXMLSession()
}

// 版本信息
const (
	Version = "2.0.0"
	Name    = "MyBatis-Go"
)

// GetVersion 获取版本信息
func GetVersion() string {
	return Version
}

// GetName 获取框架名称
func GetName() string {
	return Name
}

// GetInfo 获取框架信息
func GetInfo() map[string]string {
	return map[string]string{
		"name":        Name,
		"version":     Version,
		"description": "轻量级MyBatis风格ORM框架",
		"features":    "XML映射、动态SQL、结果映射、缓存、事务管理",
	}
}
// NewXMLMapper 创建支持XML Mapper的会话（推荐用于复杂查询）
func NewXMLMapper(db *gorm.DB) XMLSession {
	return NewXMLSession(db)
}

// NewXMLMapperWithHooks 创建带钩子的XML Mapper会话
func NewXMLMapperWithHooks(db *gorm.DB, enableDebug bool) XMLSession {
	return NewXMLSessionWithHooks(db, enableDebug)
}
