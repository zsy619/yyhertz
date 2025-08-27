// Package main 数据库设置工具
//
// 提供数据库初始化、清理和管理功能
package main

import (
	"log"

	"gorm.io/gorm"
)

// DatabaseSetup 数据库设置工具
type DatabaseSetup struct {
	DSN        string
	DB         *gorm.DB
	TestDBName string
}

// NewDatabaseSetup 创建数据库设置实例
func NewDatabaseSetup(dsn string) *DatabaseSetup {
	return &DatabaseSetup{
		DSN:        dsn,
		TestDBName: "test_gobatis",
	}
}

// Connect 连接到数据库
func (ds *DatabaseSetup) Connect() error {
	config := DefaultConfig()
	config.Database.DSN = ds.DSN
	
	db, err := InitializeDatabase(&config.Database)
	if err != nil {
		return err
	}
	
	ds.DB = db
	return nil
}

// Close 关闭数据库连接
func (ds *DatabaseSetup) Close() {
	if ds.DB != nil {
		if sqlDB, err := ds.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
}

// Reset 重置数据库（清理所有数据）
func (ds *DatabaseSetup) Reset() error {
	if ds.DB == nil {
		return nil
	}
	
	return CleanupDatabase(ds.DB)
}

// IsConnected 检查是否已连接
func (ds *DatabaseSetup) IsConnected() bool {
	if ds.DB == nil {
		return false
	}
	
	sqlDB, err := ds.DB.DB()
	if err != nil {
		return false
	}
	
	err = sqlDB.Ping()
	return err == nil
}

// LogStatus 记录数据库状态
func (ds *DatabaseSetup) LogStatus() {
	if ds.IsConnected() {
		log.Printf("数据库连接正常: %s", ds.DSN)
	} else {
		log.Printf("数据库连接异常: %s", ds.DSN)
	}
}