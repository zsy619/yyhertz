// Package orm 提供多数据库驱动支持
package orm

import (
	"fmt"
	"strings"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	
	// 额外的数据库驱动
	// Oracle驱动需要额外安装: github.com/godror/godror
	// DB2驱动需要额外安装: github.com/ibmdb/go_ibm_db
	// 达梦驱动需要额外安装: github.com/dmdbms/go-dm
)

// DatabaseDriver 数据库驱动接口
type DatabaseDriver interface {
	// BuildDSN 构建数据源连接字符串
	BuildDSN(config *DatabaseConfig) string
	// GetDialector 获取GORM方言
	GetDialector(dsn string) gorm.Dialector
	// GetDefaultPort 获取默认端口
	GetDefaultPort() int
	// ValidateConfig 验证配置
	ValidateConfig(config *DatabaseConfig) error
}

// MySQLDriver MySQL驱动
type MySQLDriver struct{}

func (d *MySQLDriver) BuildDSN(config *DatabaseConfig) string {
	if config.Port == 0 {
		config.Port = d.GetDefaultPort()
	}
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		config.Username, config.Password, config.Host, config.Port,
		config.Database, config.Charset, config.Timezone)
	
	// 添加额外参数
	if config.SSLMode != "" {
		dsn += "&tls=" + config.SSLMode
	}
	
	return dsn
}

func (d *MySQLDriver) GetDialector(dsn string) gorm.Dialector {
	return mysql.Open(dsn)
}

func (d *MySQLDriver) GetDefaultPort() int {
	return 3306
}

func (d *MySQLDriver) ValidateConfig(config *DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("MySQL host is required")
	}
	if config.Database == "" {
		return fmt.Errorf("MySQL database name is required")
	}
	return nil
}

// PostgreSQLDriver PostgreSQL驱动
type PostgreSQLDriver struct{}

func (d *PostgreSQLDriver) BuildDSN(config *DatabaseConfig) string {
	if config.Port == 0 {
		config.Port = d.GetDefaultPort()
	}
	
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		config.Host, config.Port, config.Username, config.Password,
		config.Database, sslMode, config.Timezone)
	
	// 添加Schema支持
	if config.Schema != "" {
		dsn += " search_path=" + config.Schema
	}
	
	return dsn
}

func (d *PostgreSQLDriver) GetDialector(dsn string) gorm.Dialector {
	return postgres.Open(dsn)
}

func (d *PostgreSQLDriver) GetDefaultPort() int {
	return 5432
}

func (d *PostgreSQLDriver) ValidateConfig(config *DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("PostgreSQL host is required")
	}
	if config.Database == "" {
		return fmt.Errorf("PostgreSQL database name is required")
	}
	return nil
}

// SQLiteDriver SQLite驱动
type SQLiteDriver struct{}

func (d *SQLiteDriver) BuildDSN(config *DatabaseConfig) string {
	// SQLite直接使用文件路径
	if config.Database == "" {
		config.Database = "app.db"
	}
	return config.Database
}

func (d *SQLiteDriver) GetDialector(dsn string) gorm.Dialector {
	return sqlite.Open(dsn)
}

func (d *SQLiteDriver) GetDefaultPort() int {
	return 0 // SQLite不需要端口
}

func (d *SQLiteDriver) ValidateConfig(config *DatabaseConfig) error {
	// SQLite配置验证相对简单
	return nil
}

// SQLServerDriver SQL Server驱动
type SQLServerDriver struct{}

func (d *SQLServerDriver) BuildDSN(config *DatabaseConfig) string {
	if config.Port == 0 {
		config.Port = d.GetDefaultPort()
	}
	
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)
	
	// 添加额外参数
	if config.SSLMode != "" {
		if config.SSLMode == "disable" {
			dsn += "&encrypt=disable"
		} else {
			dsn += "&encrypt=true"
		}
	}
	
	return dsn
}

func (d *SQLServerDriver) GetDialector(dsn string) gorm.Dialector {
	return sqlserver.Open(dsn)
}

func (d *SQLServerDriver) GetDefaultPort() int {
	return 1433
}

func (d *SQLServerDriver) ValidateConfig(config *DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("SQL Server host is required")
	}
	if config.Database == "" {
		return fmt.Errorf("SQL Server database name is required")
	}
	return nil
}

// OracleDriver Oracle驱动（需要额外安装驱动）
type OracleDriver struct{}

func (d *OracleDriver) BuildDSN(config *DatabaseConfig) string {
	if config.Port == 0 {
		config.Port = d.GetDefaultPort()
	}
	
	// Oracle DSN格式: user/password@host:port/service_name
	dsn := fmt.Sprintf("%s/%s@%s:%d/%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)
	
	return dsn
}

func (d *OracleDriver) GetDialector(dsn string) gorm.Dialector {
	// 需要导入Oracle驱动
	// return oracle.Open(dsn)
	panic("Oracle driver not implemented, please install github.com/godror/godror")
}

func (d *OracleDriver) GetDefaultPort() int {
	return 1521
}

func (d *OracleDriver) ValidateConfig(config *DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("Oracle host is required")
	}
	if config.Database == "" {
		return fmt.Errorf("Oracle service name is required")
	}
	return nil
}

// DMDriver 达梦数据库驱动
type DMDriver struct{}

func (d *DMDriver) BuildDSN(config *DatabaseConfig) string {
	if config.Port == 0 {
		config.Port = d.GetDefaultPort()
	}
	
	// 达梦DSN格式
	dsn := fmt.Sprintf("dm://%s:%s@%s:%d",
		config.Username, config.Password, config.Host, config.Port)
	
	return dsn
}

func (d *DMDriver) GetDialector(dsn string) gorm.Dialector {
	// 需要导入达梦驱动
	// return dm.Open(dsn)
	panic("DM driver not implemented, please install github.com/dmdbms/go-dm")
}

func (d *DMDriver) GetDefaultPort() int {
	return 5236
}

func (d *DMDriver) ValidateConfig(config *DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("DM host is required")
	}
	return nil
}

// DB2Driver DB2驱动
type DB2Driver struct{}

func (d *DB2Driver) BuildDSN(config *DatabaseConfig) string {
	if config.Port == 0 {
		config.Port = d.GetDefaultPort()
	}
	
	// DB2 DSN格式
	dsn := fmt.Sprintf("HOSTNAME=%s;DATABASE=%s;PORT=%d;UID=%s;PWD=%s",
		config.Host, config.Database, config.Port, config.Username, config.Password)
	
	return dsn
}

func (d *DB2Driver) GetDialector(dsn string) gorm.Dialector {
	// 需要导入DB2驱动
	// return db2.Open(dsn)
	panic("DB2 driver not implemented, please install github.com/ibmdb/go_ibm_db")
}

func (d *DB2Driver) GetDefaultPort() int {
	return 50000
}

func (d *DB2Driver) ValidateConfig(config *DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("DB2 host is required")
	}
	if config.Database == "" {
		return fmt.Errorf("DB2 database name is required")
	}
	return nil
}

// DriverManager 驱动管理器
type DriverManager struct {
	drivers map[string]DatabaseDriver
	mutex   sync.RWMutex
}

// NewDriverManager 创建驱动管理器
func NewDriverManager() *DriverManager {
	dm := &DriverManager{
		drivers: make(map[string]DatabaseDriver),
	}
	
	// 注册内置驱动
	dm.RegisterDriver("mysql", &MySQLDriver{})
	dm.RegisterDriver("postgres", &PostgreSQLDriver{})
	dm.RegisterDriver("postgresql", &PostgreSQLDriver{}) // 别名
	dm.RegisterDriver("sqlite", &SQLiteDriver{})
	dm.RegisterDriver("sqlite3", &SQLiteDriver{}) // 别名
	dm.RegisterDriver("sqlserver", &SQLServerDriver{})
	dm.RegisterDriver("mssql", &SQLServerDriver{}) // 别名
	dm.RegisterDriver("oracle", &OracleDriver{})
	dm.RegisterDriver("dm", &DMDriver{})
	dm.RegisterDriver("dameng", &DMDriver{}) // 别名
	dm.RegisterDriver("db2", &DB2Driver{})
	
	return dm
}

// RegisterDriver 注册驱动
func (dm *DriverManager) RegisterDriver(name string, driver DatabaseDriver) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	
	dm.drivers[strings.ToLower(name)] = driver
}

// GetDriver 获取驱动
func (dm *DriverManager) GetDriver(name string) (DatabaseDriver, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	
	driver, exists := dm.drivers[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("unsupported database driver: %s", name)
	}
	
	return driver, nil
}

// ListDrivers 列出所有支持的驱动
func (dm *DriverManager) ListDrivers() []string {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	
	drivers := make([]string, 0, len(dm.drivers))
	for name := range dm.drivers {
		drivers = append(drivers, name)
	}
	
	return drivers
}

// 全局驱动管理器实例
var globalDriverManager = NewDriverManager()

// GetDriverManager 获取全局驱动管理器
func GetDriverManager() *DriverManager {
	return globalDriverManager
}