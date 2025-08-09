// Package main 数据库设置工具
//
// 提供数据库初始化、清理和管理功能
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseSetup 数据库设置工具
type DatabaseSetup struct {
	DSN        string
	DB         *gorm.DB
	TestDBName string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string // SQLite数据库文件路径
}

// DefaultDatabaseConfig 默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Path: ":memory:", // 使用内存数据库
	}
}

// BuildDSN 构建DSN字符串
func (config *DatabaseConfig) BuildDSN() string {
	return config.Path
}

// BuildSystemDSN 构建系统DSN（对SQLite来说和BuildDSN相同）
func (config *DatabaseConfig) BuildSystemDSN() string {
	return config.Path
}

// NewDatabaseSetup 创建数据库设置工具
func NewDatabaseSetup(config *DatabaseConfig) *DatabaseSetup {
	if config == nil {
		config = DefaultDatabaseConfig()
	}
	
	return &DatabaseSetup{
		DSN:        config.BuildDSN(),
		TestDBName: "sqlite_test", // SQLite测试数据库名
	}
}

// CreateTestDatabase 创建测试数据库（SQLite无需此操作）
func (setup *DatabaseSetup) CreateTestDatabase() error {
	// SQLite使用内存数据库，无需手动创建
	log.Printf("使用SQLite内存数据库")
	return nil
}

// DropTestDatabase 删除测试数据库（SQLite内存数据库会自动清理）
func (setup *DatabaseSetup) DropTestDatabase() error {
	// SQLite内存数据库会在连接关闭时自动清理
	log.Printf("SQLite内存数据库会自动清理")
	return nil
}

// ConnectToDatabase 连接到数据库
func (setup *DatabaseSetup) ConnectToDatabase() error {
	// 配置GORM日志
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	}
	
	db, err := gorm.Open(sqlite.Open(setup.DSN), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	setup.DB = db
	log.Printf("数据库连接成功: %s", setup.DSN)
	return nil
}

// CloseConnection 关闭数据库连接
func (setup *DatabaseSetup) CloseConnection() error {
	if setup.DB != nil {
		sqlDB, err := setup.DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		
		err = sqlDB.Close()
		if err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
		
		setup.DB = nil
		log.Println("数据库连接已关闭")
	}
	return nil
}

// CreateTables 创建所有表
func (setup *DatabaseSetup) CreateTables() error {
	if setup.DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	
	ctx := context.Background()
	
	// 表创建SQL列表
	tables := []struct {
		Name string
		SQL  string
	}{
		{"users", CreateUsersTableSQL},
		{"user_profiles", CreateUserProfilesTableSQL},
		{"user_roles", CreateUserRolesTableSQL},
		{"articles", CreateArticlesTableSQL},
		{"categories", CreateCategoriesTableSQL},
		{"user_article_views", CreateUserArticleViewsTableSQL},
	}
	
	// 创建表
	for _, table := range tables {
		if err := setup.DB.WithContext(ctx).Exec(table.SQL).Error; err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.Name, err)
		}
		log.Printf("表 '%s' 创建成功", table.Name)
	}
	
	return nil
}

// DropTables 删除所有表
func (setup *DatabaseSetup) DropTables() error {
	if setup.DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	
	ctx := context.Background()
	
	// 表删除顺序（考虑外键约束）
	tables := []string{
		"user_article_views",
		"articles", 
		"categories",
		"user_roles",
		"user_profiles",
		"users",
	}
	
	// 禁用外键检查
	if err := setup.DB.WithContext(ctx).Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		log.Printf("Warning: Failed to disable foreign key checks: %v", err)
	}
	
	// 删除表
	for _, tableName := range tables {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
		if err := setup.DB.WithContext(ctx).Exec(dropSQL).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", tableName, err)
		} else {
			log.Printf("表 '%s' 删除成功", tableName)
		}
	}
	
	// 重新启用外键检查
	if err := setup.DB.WithContext(ctx).Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		log.Printf("Warning: Failed to enable foreign key checks: %v", err)
	}
	
	return nil
}

// InsertTestData 插入测试数据
func (setup *DatabaseSetup) InsertTestData() error {
	if setup.DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	
	ctx := context.Background()
	
	// 测试数据SQL列表
	testData := []struct {
		Name string
		SQL  string
	}{
		{"test_users", InsertTestUsersSQL},
		{"test_profiles", InsertTestProfilesSQL},
		{"test_roles", InsertTestRolesSQL},
	}
	
	// 插入测试数据
	for _, data := range testData {
		if err := setup.DB.WithContext(ctx).Exec(data.SQL).Error; err != nil {
			return fmt.Errorf("failed to insert %s: %w", data.Name, err)
		}
		log.Printf("测试数据 '%s' 插入成功", data.Name)
	}
	
	return nil
}

// CreateStoredProcedures 创建存储过程和函数
func (setup *DatabaseSetup) CreateStoredProcedures() error {
	if setup.DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	
	ctx := context.Background()
	
	// 存储过程和函数列表
	procedures := []struct {
		Name string
		SQL  string
	}{
		{"GetUserStats", CreateUserStatsProcedureSQL},
		{"GetUsersByCustomFunction", CreateCustomFunctionSQL},
	}
	
	// 创建存储过程和函数
	for _, proc := range procedures {
		if err := setup.DB.WithContext(ctx).Exec(proc.SQL).Error; err != nil {
			log.Printf("Warning: Failed to create %s: %v", proc.Name, err)
		} else {
			log.Printf("存储过程/函数 '%s' 创建成功", proc.Name)
		}
	}
	
	return nil
}

// ClearTestData 清理测试数据
func (setup *DatabaseSetup) ClearTestData() error {
	if setup.DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	
	ctx := context.Background()
	
	// 清理数据的顺序（考虑外键约束）
	tables := []string{
		"user_article_views",
		"articles",
		"user_roles", 
		"user_profiles",
		"users",
		"categories",
	}
	
	// 禁用外键检查
	if err := setup.DB.WithContext(ctx).Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		log.Printf("Warning: Failed to disable foreign key checks: %v", err)
	}
	
	// 清理各表数据
	for _, tableName := range tables {
		truncateSQL := fmt.Sprintf("TRUNCATE TABLE `%s`", tableName)
		if err := setup.DB.WithContext(ctx).Exec(truncateSQL).Error; err != nil {
			// 如果TRUNCATE失败，尝试DELETE
			deleteSQL := fmt.Sprintf("DELETE FROM `%s`", tableName)
			if err2 := setup.DB.WithContext(ctx).Exec(deleteSQL).Error; err2 != nil {
				log.Printf("Warning: Failed to clear table %s: %v", tableName, err2)
			} else {
				log.Printf("表 '%s' 数据清理成功 (DELETE)", tableName)
			}
		} else {
			log.Printf("表 '%s' 数据清理成功 (TRUNCATE)", tableName)
		}
	}
	
	// 重新启用外键检查
	if err := setup.DB.WithContext(ctx).Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		log.Printf("Warning: Failed to enable foreign key checks: %v", err)
	}
	
	return nil
}

// ResetAutoIncrement 重置自增ID
func (setup *DatabaseSetup) ResetAutoIncrement() error {
	if setup.DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	
	ctx := context.Background()
	
	// 有自增ID的表
	tables := []string{
		"users",
		"user_roles", 
		"articles",
		"categories",
		"user_article_views",
	}
	
	// 重置自增ID
	for _, tableName := range tables {
		alterSQL := fmt.Sprintf("ALTER TABLE `%s` AUTO_INCREMENT = 1", tableName)
		if err := setup.DB.WithContext(ctx).Exec(alterSQL).Error; err != nil {
			log.Printf("Warning: Failed to reset auto increment for table %s: %v", tableName, err)
		} else {
			log.Printf("表 '%s' 自增ID重置成功", tableName)
		}
	}
	
	return nil
}

// SetupCompleteTestEnvironment 设置完整的测试环境
func (setup *DatabaseSetup) SetupCompleteTestEnvironment() error {
	log.Println("开始设置完整测试环境...")
	
	// 1. 创建测试数据库
	if err := setup.CreateTestDatabase(); err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}
	
	// 2. 连接到数据库
	if err := setup.ConnectToDatabase(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 3. 创建表
	if err := setup.CreateTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	
	// 4. 插入测试数据
	if err := setup.InsertTestData(); err != nil {
		return fmt.Errorf("failed to insert test data: %w", err)
	}
	
	// 5. 创建存储过程和函数
	if err := setup.CreateStoredProcedures(); err != nil {
		log.Printf("Warning: Some stored procedures/functions failed to create: %v", err)
	}
	
	log.Println("完整测试环境设置完成！")
	return nil
}

// TeardownCompleteTestEnvironment 清理完整的测试环境
func (setup *DatabaseSetup) TeardownCompleteTestEnvironment() error {
	log.Println("开始清理测试环境...")
	
	// 1. 清理测试数据
	if err := setup.ClearTestData(); err != nil {
		log.Printf("Warning: Failed to clear test data: %v", err)
	}
	
	// 2. 删除表
	if err := setup.DropTables(); err != nil {
		log.Printf("Warning: Failed to drop tables: %v", err)
	}
	
	// 3. 关闭连接
	if err := setup.CloseConnection(); err != nil {
		log.Printf("Warning: Failed to close connection: %v", err)
	}
	
	// 4. 删除测试数据库
	if err := setup.DropTestDatabase(); err != nil {
		log.Printf("Warning: Failed to drop test database: %v", err)
	}
	
	log.Println("测试环境清理完成！")
	return nil
}

// CheckDatabaseConnection 检查数据库连接
func (setup *DatabaseSetup) CheckDatabaseConnection() error {
	if setup.DB == nil {
		return fmt.Errorf("database connection is not established")
	}
	
	sqlDB, err := setup.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	log.Println("数据库连接正常")
	return nil
}

// GetTableInfo 获取表信息
func (setup *DatabaseSetup) GetTableInfo() (map[string]TableInfo, error) {
	if setup.DB == nil {
		return nil, fmt.Errorf("database connection is not established")
	}
	
	ctx := context.Background()
	
	// 查询表信息
	var tables []TableInfo
	query := `
		SELECT 
			TABLE_NAME as name,
			TABLE_ROWS as row_count,
			DATA_LENGTH as data_length,
			INDEX_LENGTH as index_length,
			CREATE_TIME as created_at
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? 
		ORDER BY TABLE_NAME
	`
	
	if err := setup.DB.WithContext(ctx).Raw(query, setup.TestDBName).Scan(&tables).Error; err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}
	
	// 转换为map
	result := make(map[string]TableInfo)
	for _, table := range tables {
		result[table.Name] = table
	}
	
	return result, nil
}

// TableInfo 表信息
type TableInfo struct {
	Name        string     `json:"name"`
	RowCount    int64      `json:"row_count"`
	DataLength  int64      `json:"data_length"`
	IndexLength int64      `json:"index_length"`
	CreatedAt   *time.Time `json:"created_at"`
}

// PrintDatabaseInfo 打印数据库信息
func (setup *DatabaseSetup) PrintDatabaseInfo() error {
	tableInfo, err := setup.GetTableInfo()
	if err != nil {
		return err
	}
	
	fmt.Println("\n=== 数据库信息 ===")
	fmt.Printf("数据库名称: %s\n", setup.TestDBName)
	fmt.Printf("连接字符串: %s\n", setup.DSN)
	fmt.Printf("表数量: %d\n\n", len(tableInfo))
	
	fmt.Println("表详细信息:")
	for tableName, info := range tableInfo {
		fmt.Printf("  - %s: %d 行, 数据大小: %d bytes, 索引大小: %d bytes\n", 
			tableName, info.RowCount, info.DataLength, info.IndexLength)
	}
	fmt.Println()
	
	return nil
}