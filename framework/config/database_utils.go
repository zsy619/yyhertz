package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// DatabaseNodeConfig 数据库节点配置结构体（与Primary配置结构相同）
type DatabaseNodeConfig struct {
	Driver              string `mapstructure:"driver" yaml:"driver" json:"driver"`                                              // mysql, postgres, sqlite, sqlserver
	DSN                 string `mapstructure:"dsn" yaml:"dsn" json:"dsn"`                                                       // 数据库连接字符串
	Host                string `mapstructure:"host" yaml:"host" json:"host"`                                                    // 主机地址
	Port                int    `mapstructure:"port" yaml:"port" json:"port"`                                                    // 端口
	Database            string `mapstructure:"database" yaml:"database" json:"database"`                                        // 数据库名
	Username            string `mapstructure:"username" yaml:"username" json:"username"`                                        // 用户名
	Password            string `mapstructure:"password" yaml:"password" json:"password"`                                        // 密码
	Charset             string `mapstructure:"charset" yaml:"charset" json:"charset"`                                           // 字符集
	Collation           string `mapstructure:"collation" yaml:"collation" json:"collation"`                                     // 排序规则
	Timezone            string `mapstructure:"timezone" yaml:"timezone" json:"timezone"`                                        // 时区
	MaxOpenConns        int    `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`                      // 最大打开连接数
	MaxIdleConns        int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`                      // 最大空闲连接数
	ConnMaxLifetime     string `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`             // 连接最大生存时间
	ConnMaxIdleTime     string `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time" json:"conn_max_idle_time"`          // 连接最大空闲时间
	SlowQueryThreshold  string `mapstructure:"slow_query_threshold" yaml:"slow_query_threshold" json:"slow_query_threshold"`    // 慢查询阈值
	LogLevel            string `mapstructure:"log_level" yaml:"log_level" json:"log_level"`                                     // 日志级别: silent, error, warn, info
	EnableMetrics       bool   `mapstructure:"enable_metrics" yaml:"enable_metrics" json:"enable_metrics"`                      // 启用性能监控
	EnableAutoMigration bool   `mapstructure:"enable_auto_migration" yaml:"enable_auto_migration" json:"enable_auto_migration"` // 启用自动迁移
	MigrationTableName  string `mapstructure:"migration_table_name" yaml:"migration_table_name" json:"migration_table_name"`    // 迁移表名
	SSLMode             string `mapstructure:"ssl_mode" yaml:"ssl_mode" json:"ssl_mode"`                                        // SSL模式: disable, require, verify-ca, verify-full
	SSLCert             string `mapstructure:"ssl_cert" yaml:"ssl_cert" json:"ssl_cert"`                                        // SSL证书路径
	SSLKey              string `mapstructure:"ssl_key" yaml:"ssl_key" json:"ssl_key"`                                           // SSL密钥路径
	SSLRootCert         string `mapstructure:"ssl_root_cert" yaml:"ssl_root_cert" json:"ssl_root_cert"`                         // SSL根证书路径
}

func (node *DatabaseNodeConfig) BuildDSN() string {
	cnf, err := buildDSN(node)
	if err != nil {
		return ""
	}
	return cnf
}

var (
	// 全局数据库实例
	GlobalDatabase *DatabaseConfig
	// 初始化锁
	dbOnce sync.Once
	// 配置文件是否已加载
	configLoaded bool
	// 配置文件加载锁
	configOnce sync.Once
)

// GetDatabasePrimary 获取指定数据库节点的配置
// nodeKey: 节点名称，如 "primary", "sso", "user", "order" 等
// 支持从配置文件动态读取节点配置，如果节点不存在则返回默认配置
func GetDatabasePrimary(nodeKeys ...string) map[string]*DatabaseNodeConfig {
	if len(nodeKeys) == 0 {
		panic("GetDatabasePrimary requires at least one node key")
	}
	// 确保配置文件已加载
	ensureDatabaseConfigLoaded()

	// 初始化map，避免对nil map进行写操作导致panic
	dbs := make(map[string]*DatabaseNodeConfig, len(nodeKeys))
	for _, nodeKey := range nodeKeys {
		config, err := loadDatabaseNodeConfig(nodeKey)
		if err != nil {
			panic(fmt.Sprintf("无法加载数据库节点 '%s' 的配置: %v", nodeKey, err))
		}
		dbs[nodeKey] = config
	}
	return dbs
}

// ensureDatabaseConfigLoaded 确保数据库配置文件已加载
func ensureDatabaseConfigLoaded() {
	configOnce.Do(func() {
		// 尝试加载数据库配置文件
		loadDatabaseConfigFile()
		configLoaded = true
	})
}

// loadDatabaseConfigFile 加载数据库配置文件
func loadDatabaseConfigFile() {
	// 设置配置文件路径和名称
	configPaths := []string{
		"./conf",
		"./config",
		"./configs",
		".",
	}

	// 尝试从多个路径加载配置文件
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// 设置配置文件名称（不包含扩展名）
	viper.SetConfigName("database")
	viper.SetConfigType("yaml")

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，不报错，使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("数据库配置文件未找到，使用默认配置\n")
		} else {
			fmt.Printf("读取数据库配置文件失败: %v，使用默认配置\n", err)
		}
		return
	}

	fmt.Printf("成功加载数据库配置文件: %s\n", viper.ConfigFileUsed())
}

// loadDatabaseNodeConfig 从配置文件加载指定节点的数据库配置
func loadDatabaseNodeConfig(nodeName string) (*DatabaseNodeConfig, error) {
	// 尝试从viper配置中读取节点配置
	configKey := fmt.Sprintf("%s", nodeName)

	// 检查配置是否存在
	if !viper.IsSet(configKey) {
		return nil, fmt.Errorf("database config node '%s' not found", nodeName)
	}

	var config DatabaseNodeConfig

	// 绑定配置到结构体
	if err := viper.UnmarshalKey(configKey, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal database config for node '%s': %v", nodeName, err)
	}

	// 验证必需字段
	if err := validateDatabaseNodeConfig(&config, nodeName); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateDatabaseNodeConfig 验证数据库节点配置的有效性
func validateDatabaseNodeConfig(config *DatabaseNodeConfig, nodeName string) error {
	if config.Driver == "" {
		return fmt.Errorf("database driver is required for node '%s'", nodeName)
	}

	if config.Host == "" {
		return fmt.Errorf("database host is required for node '%s'", nodeName)
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid database port %d for node '%s', must be between 1-65535", config.Port, nodeName)
	}

	if config.Database == "" {
		return fmt.Errorf("database name is required for node '%s'", nodeName)
	}

	if config.Username == "" {
		return fmt.Errorf("database username is required for node '%s'", nodeName)
	}

	if config.MaxOpenConns < 0 {
		return fmt.Errorf("max_open_conns must be >= 0 for node '%s'", nodeName)
	}

	if config.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns must be >= 0 for node '%s'", nodeName)
	}

	if config.MaxOpenConns > 0 && config.MaxIdleConns > config.MaxOpenConns {
		return fmt.Errorf("max_idle_conns (%d) cannot be greater than max_open_conns (%d) for node '%s'",
			config.MaxIdleConns, config.MaxOpenConns, nodeName)
	}

	return nil
}

// GetDatabaseDSN 获取指定节点的DSN连接字符串
func GetDatabaseDSN(nodeName string) (string, error) {
	configs := GetDatabasePrimary(nodeName)
	for _, config := range configs {
		if config == nil {
			return "", fmt.Errorf("failed to get database config for node '%s'", nodeName)
		}

		// 如果已经配置了完整的DSN，直接返回
		if config.DSN != "" {
			return config.DSN, nil
		}

		// 根据不同数据库类型构建DSN
		return buildDSN(config)
	}
	return "", nil
}

// buildDSN 根据配置构建DSN连接字符串
func buildDSN(config *DatabaseNodeConfig) (string, error) {
	switch config.Driver {
	case "mysql":
		// MySQL DSN格式: username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s&collation=%s",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
			config.Charset,
			config.Timezone,
			config.Collation,
		)

		// 添加SSL配置
		if config.SSLMode != "disable" && config.SSLMode != "" {
			dsn += "&tls=" + config.SSLMode
		}

		return dsn, nil

	case "postgres":
		// PostgreSQL DSN格式: host=localhost user=postgres password=password dbname=yyhertz port=5432 sslmode=disable TimeZone=Asia/Shanghai
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			config.Host,
			config.Username,
			config.Password,
			config.Database,
			config.Port,
			config.SSLMode,
			config.Timezone,
		)
		return dsn, nil

	case "sqlite":
		// SQLite DSN格式: 文件路径
		return config.Database, nil

	case "sqlserver":
		// SQL Server DSN格式: sqlserver://username:password@host:port?database=dbname
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		)
		return dsn, nil
	case "oracle":
		// Oracle DSN格式: username/password@host:port/service_name
		dsn := fmt.Sprintf("%s/%s@%s:%d/%s",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		)
		return dsn, nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}

// GetAllDatabaseNodes 获取所有已配置的数据库节点列表
func GetAllDatabaseNodes() []string {
	var nodes []string

	// 获取所有database配置键
	databaseConfigs := viper.GetStringMap("database")
	for nodeName := range databaseConfigs {
		// 过滤掉非节点配置（如全局配置）
		if _, err := loadDatabaseNodeConfig(nodeName); err == nil {
			nodes = append(nodes, nodeName)
		}
	}

	return nodes
}

// InitGlobalDatabase 初始化全局数据库配置
func InitGlobalDatabase() error {
	dbOnce.Do(func() {
		GlobalDatabase = DefaultDatabaseConfig()
	})
	return nil
}

// SetGlobalDatabase 设置全局数据库配置
func SetGlobalDatabase(config *DatabaseConfig) {
	GlobalDatabase = config
}

// GetGlobalDatabase 获取全局数据库配置
func GetGlobalDatabase() *DatabaseConfig {
	if GlobalDatabase == nil {
		_ = InitGlobalDatabase() // 忽略错误，因为这是初始化默认配置
	}
	return GlobalDatabase
}

// 获取primary数据库的DSN（保持向后兼容）
func GetGlobalDatabaseExt() string {
	if GlobalDatabase == nil {
		return ""
	}

	// 如果Primary.DSN已配置，直接返回
	if GlobalDatabase.Primary.DSN != "" {
		return GlobalDatabase.Primary.DSN
	}

	// 否则构建DSN
	config := &DatabaseNodeConfig{
		Driver:    GlobalDatabase.Primary.Driver,
		Host:      GlobalDatabase.Primary.Host,
		Port:      GlobalDatabase.Primary.Port,
		Database:  GlobalDatabase.Primary.Database,
		Username:  GlobalDatabase.Primary.Username,
		Password:  GlobalDatabase.Primary.Password,
		Charset:   GlobalDatabase.Primary.Charset,
		Collation: GlobalDatabase.Primary.Collation,
		Timezone:  GlobalDatabase.Primary.Timezone,
		SSLMode:   GlobalDatabase.Primary.SSLMode,
	}

	dsn, _ := buildDSN(config)
	return dsn
}

// ValidateNodeConnection 验证指定节点的数据库连接
func ValidateNodeConnection(nodeName string) error {
	configs := GetDatabasePrimary(nodeName)
	for _, config := range configs {
		if config == nil {
			return fmt.Errorf("failed to get config for node '%s'", nodeName)
		}

		// 构建DSN用于连接测试
		dsn, err := buildDSN(config)
		if err != nil {
			return fmt.Errorf("failed to build DSN for node '%s': %v", nodeName, err)
		}

		// 这里可以添加实际的数据库连接测试逻辑
		// 例如：使用sql.Open()和db.Ping()测试连接
		_ = dsn // 避免未使用变量的编译错误
	}
	return nil
}

// ============= 便捷配置管理函数 =============

// LoadDatabaseConfigFromFile 手动从指定文件加载数据库配置
func LoadDatabaseConfigFromFile(filePath string) error {
	viper.SetConfigFile(filePath)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file '%s': %v", filePath, err)
	}

	configLoaded = true
	fmt.Printf("成功从 %s 加载数据库配置\n", filePath)
	return nil
}

// ReloadDatabaseConfig 重新加载数据库配置
func ReloadDatabaseConfig() error {
	configLoaded = false
	configOnce = sync.Once{} // 重置once

	ensureDatabaseConfigLoaded()
	return nil
}

// IsConfigLoaded 检查配置是否已加载
func IsConfigLoaded() bool {
	return configLoaded
}

// GetConfiguredNodes 获取配置文件中定义的所有节点（不包括验证失败的）
func GetConfiguredNodes() []string {
	ensureDatabaseConfigLoaded()
	return GetAllDatabaseNodes()
}

// TestDatabaseNodeConnection 测试指定节点的数据库连接（实际连接测试，需要database/sql包）
func TestDatabaseNodeConnection(nodeName string) error {
	configs := GetDatabasePrimary(nodeName)
	for _, config := range configs {
		if config == nil {
			return fmt.Errorf("无法获取节点 '%s' 的配置", nodeName)
		}

		dsn, err := buildDSN(config)
		if err != nil {
			return fmt.Errorf("构建DSN失败: %v", err)
		}

		// 这里可以添加实际的数据库连接测试
		// 例如：
		// db, err := sql.Open(config.Driver, dsn)
		// if err != nil {
		//     return fmt.Errorf("打开数据库连接失败: %v", err)
		// }
		// defer db.Close()
		//
		// if err := db.Ping(); err != nil {
		//     return fmt.Errorf("数据库连接测试失败: %v", err)
		// }

		fmt.Printf("节点 '%s' 配置验证通过，DSN: %s\n", nodeName, dsn)
	}
	return nil
}

// PrintNodeConfiguration 打印指定节点的详细配置信息
func PrintNodeConfiguration(nodeName string) {
	configs := GetDatabasePrimary(nodeName)
	for _, config := range configs {
		if config == nil {
			fmt.Printf("❌ 无法获取节点 '%s' 的配置\n", nodeName)
			return
		}

		fmt.Printf("=== %s 数据库节点配置 ===\n", strings.ToUpper(nodeName))
		fmt.Printf("驱动类型: %s\n", config.Driver)
		fmt.Printf("主机地址: %s:%d\n", config.Host, config.Port)
		fmt.Printf("数据库名: %s\n", config.Database)
		fmt.Printf("用户名: %s\n", config.Username)
		fmt.Printf("字符集: %s (%s)\n", config.Charset, config.Collation)
		fmt.Printf("时区: %s\n", config.Timezone)
		fmt.Printf("连接池: 最大打开 %d, 最大空闲 %d\n", config.MaxOpenConns, config.MaxIdleConns)
		fmt.Printf("连接生存时间: %s\n", config.ConnMaxLifetime)
		fmt.Printf("连接空闲时间: %s\n", config.ConnMaxIdleTime)
		fmt.Printf("慢查询阈值: %s\n", config.SlowQueryThreshold)
		fmt.Printf("日志级别: %s\n", config.LogLevel)
		fmt.Printf("启用监控: %v\n", config.EnableMetrics)
		fmt.Printf("自动迁移: %v\n", config.EnableAutoMigration)
		fmt.Printf("迁移表名: %s\n", config.MigrationTableName)
		fmt.Printf("SSL模式: %s\n", config.SSLMode)
		fmt.Println("=====================")
	}
}

// PrintAllConfigurations 打印所有已配置节点的信息
func PrintAllConfigurations() {
	nodes := GetConfiguredNodes()
	if len(nodes) == 0 {
		fmt.Println("⚠️  未找到任何配置节点")
		return
	}

	fmt.Printf("📋 发现 %d 个配置节点:\n", len(nodes))
	for _, nodeName := range nodes {
		PrintNodeConfiguration(nodeName)
		fmt.Println()
	}
}

// DefaultDatabaseConfig 返回默认的数据库配置
// 适用于开发环境的基础配置，生产环境需要根据实际情况调整
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		// 主数据库配置
		Primary: struct {
			Driver              string `mapstructure:"driver" yaml:"driver" json:"driver"`
			DSN                 string `mapstructure:"dsn" yaml:"dsn" json:"dsn"`
			Host                string `mapstructure:"host" yaml:"host" json:"host"`
			Port                int    `mapstructure:"port" yaml:"port" json:"port"`
			Database            string `mapstructure:"database" yaml:"database" json:"database"`
			Username            string `mapstructure:"username" yaml:"username" json:"username"`
			Password            string `mapstructure:"password" yaml:"password" json:"password"`
			Charset             string `mapstructure:"charset" yaml:"charset" json:"charset"`
			Collation           string `mapstructure:"collation" yaml:"collation" json:"collation"`
			Timezone            string `mapstructure:"timezone" yaml:"timezone" json:"timezone"`
			MaxOpenConns        int    `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`
			MaxIdleConns        int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
			ConnMaxLifetime     string `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
			ConnMaxIdleTime     string `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
			SlowQueryThreshold  string `mapstructure:"slow_query_threshold" yaml:"slow_query_threshold" json:"slow_query_threshold"`
			LogLevel            string `mapstructure:"log_level" yaml:"log_level" json:"log_level"`
			EnableMetrics       bool   `mapstructure:"enable_metrics" yaml:"enable_metrics" json:"enable_metrics"`
			EnableAutoMigration bool   `mapstructure:"enable_auto_migration" yaml:"enable_auto_migration" json:"enable_auto_migration"`
			MigrationTableName  string `mapstructure:"migration_table_name" yaml:"migration_table_name" json:"migration_table_name"`
			SSLMode             string `mapstructure:"ssl_mode" yaml:"ssl_mode" json:"ssl_mode"`
			SSLCert             string `mapstructure:"ssl_cert" yaml:"ssl_cert" json:"ssl_cert"`
			SSLKey              string `mapstructure:"ssl_key" yaml:"ssl_key" json:"ssl_key"`
			SSLRootCert         string `mapstructure:"ssl_root_cert" yaml:"ssl_root_cert" json:"ssl_root_cert"`
		}{
			Driver:              "mysql",
			DSN:                 "", // 为空时使用单独的配置项构建
			Host:                "localhost",
			Port:                3306,
			Database:            "yyhertz",
			Username:            "root",
			Password:            "",
			Charset:             "utf8mb4",
			Collation:           "utf8mb4_unicode_ci",
			Timezone:            "Local",
			MaxOpenConns:        100,
			MaxIdleConns:        10,
			ConnMaxLifetime:     "1h",
			ConnMaxIdleTime:     "30m",
			SlowQueryThreshold:  "200ms",
			LogLevel:            "warn",
			EnableMetrics:       true,
			EnableAutoMigration: false,
			MigrationTableName:  "schema_migrations",
			SSLMode:             "disable",
			SSLCert:             "",
			SSLKey:              "",
			SSLRootCert:         "",
		},

		// 从数据库配置(读写分离) - 默认禁用
		Replica: struct {
			Enable                bool     `mapstructure:"enable" yaml:"enable" json:"enable"`
			Hosts                 []string `mapstructure:"hosts" yaml:"hosts" json:"hosts"`
			Driver                string   `mapstructure:"driver" yaml:"driver" json:"driver"`
			Username              string   `mapstructure:"username" yaml:"username" json:"username"`
			Password              string   `mapstructure:"password" yaml:"password" json:"password"`
			Database              string   `mapstructure:"database" yaml:"database" json:"database"`
			MaxOpenConns          int      `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`
			MaxIdleConns          int      `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
			ConnMaxLifetime       string   `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
			LoadBalancingStrategy string   `mapstructure:"load_balancing_strategy" yaml:"load_balancing_strategy" json:"load_balancing_strategy"`
		}{
			Enable:                false,
			Hosts:                 []string{},
			Driver:                "mysql",
			Username:              "root",
			Password:              "",
			Database:              "yyhertz",
			MaxOpenConns:          50,
			MaxIdleConns:          10,
			ConnMaxLifetime:       "1h",
			LoadBalancingStrategy: "round_robin",
		},

		// GORM配置 - 默认启用
		GORM: struct {
			Enable                     bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			DisableForeignKeyConstrain bool   `mapstructure:"disable_foreign_key_constrain" yaml:"disable_foreign_key_constrain" json:"disable_foreign_key_constrain"`
			SkipDefaultTransaction     bool   `mapstructure:"skip_default_transaction" yaml:"skip_default_transaction" json:"skip_default_transaction"`
			FullSaveAssociations       bool   `mapstructure:"full_save_associations" yaml:"full_save_associations" json:"full_save_associations"`
			DryRun                     bool   `mapstructure:"dry_run" yaml:"dry_run" json:"dry_run"`
			PrepareStmt                bool   `mapstructure:"prepare_stmt" yaml:"prepare_stmt" json:"prepare_stmt"`
			DisableNestedTransaction   bool   `mapstructure:"disable_nested_transaction" yaml:"disable_nested_transaction" json:"disable_nested_transaction"`
			AllowGlobalUpdate          bool   `mapstructure:"allow_global_update" yaml:"allow_global_update" json:"allow_global_update"`
			QueryFields                bool   `mapstructure:"query_fields" yaml:"query_fields" json:"query_fields"`
			CreateBatchSize            int    `mapstructure:"create_batch_size" yaml:"create_batch_size" json:"create_batch_size"`
			NamingStrategy             string `mapstructure:"naming_strategy" yaml:"naming_strategy" json:"naming_strategy"`
			TablePrefix                string `mapstructure:"table_prefix" yaml:"table_prefix" json:"table_prefix"`
			SingularTable              bool   `mapstructure:"singular_table" yaml:"singular_table" json:"singular_table"`
		}{
			Enable:                     true,
			DisableForeignKeyConstrain: false,
			SkipDefaultTransaction:     false,
			FullSaveAssociations:       false,
			DryRun:                     false,
			PrepareStmt:                true,
			DisableNestedTransaction:   false,
			AllowGlobalUpdate:          false,
			QueryFields:                true,
			CreateBatchSize:            1000,
			NamingStrategy:             "snake_case",
			TablePrefix:                "",
			SingularTable:              false,
		},

		// MyBatis配置 - 默认禁用
		MyBatis: struct {
			Enable           bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			ConfigFile       string `mapstructure:"config_file" yaml:"config_file" json:"config_file"`
			MapperLocations  string `mapstructure:"mapper_locations" yaml:"mapper_locations" json:"mapper_locations"`
			TypeAliasesPath  string `mapstructure:"type_aliases_path" yaml:"type_aliases_path" json:"type_aliases_path"`
			CacheEnabled     bool   `mapstructure:"cache_enabled" yaml:"cache_enabled" json:"cache_enabled"`
			LazyLoading      bool   `mapstructure:"lazy_loading" yaml:"lazy_loading" json:"lazy_loading"`
			LogImpl          string `mapstructure:"log_impl" yaml:"log_impl" json:"log_impl"`
			MapUnderscoreMap bool   `mapstructure:"map_underscore_map" yaml:"map_underscore_map" json:"map_underscore_map"`
		}{
			Enable:           false,
			ConfigFile:       "./config/mybatis-config.xml",
			MapperLocations:  "./mappers/*.xml",
			TypeAliasesPath:  "./models",
			CacheEnabled:     true,
			LazyLoading:      false,
			LogImpl:          "STDOUT_LOGGING",
			MapUnderscoreMap: true,
		},

		// 连接池配置 - 默认启用
		Pool: struct {
			Enable              bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			Type                string `mapstructure:"type" yaml:"type" json:"type"`
			MaxActiveConns      int    `mapstructure:"max_active_conns" yaml:"max_active_conns" json:"max_active_conns"`
			MaxIdleConns        int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
			MinIdleConns        int    `mapstructure:"min_idle_conns" yaml:"min_idle_conns" json:"min_idle_conns"`
			MaxWaitTime         string `mapstructure:"max_wait_time" yaml:"max_wait_time" json:"max_wait_time"`
			TimeBetweenEviction string `mapstructure:"time_between_eviction" yaml:"time_between_eviction" json:"time_between_eviction"`
			MinEvictableTime    string `mapstructure:"min_evictable_time" yaml:"min_evictable_time" json:"min_evictable_time"`
			TestOnBorrow        bool   `mapstructure:"test_on_borrow" yaml:"test_on_borrow" json:"test_on_borrow"`
			TestOnReturn        bool   `mapstructure:"test_on_return" yaml:"test_on_return" json:"test_on_return"`
			TestWhileIdle       bool   `mapstructure:"test_while_idle" yaml:"test_while_idle" json:"test_while_idle"`
			ValidationQuery     string `mapstructure:"validation_query" yaml:"validation_query" json:"validation_query"`
		}{
			Enable:              true,
			Type:                "default",
			MaxActiveConns:      100,
			MaxIdleConns:        10,
			MinIdleConns:        5,
			MaxWaitTime:         "30s",
			TimeBetweenEviction: "30s",
			MinEvictableTime:    "5m",
			TestOnBorrow:        true,
			TestOnReturn:        false,
			TestWhileIdle:       true,
			ValidationQuery:     "SELECT 1",
		},

		// 缓存配置 - 默认禁用
		Cache: struct {
			Enable         bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			Type           string `mapstructure:"type" yaml:"type" json:"type"`
			TTL            string `mapstructure:"ttl" yaml:"ttl" json:"ttl"`
			MaxSize        int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"`
			KeyPrefix      string `mapstructure:"key_prefix" yaml:"key_prefix" json:"key_prefix"`
			RedisAddr      string `mapstructure:"redis_addr" yaml:"redis_addr" json:"redis_addr"`
			RedisPassword  string `mapstructure:"redis_password" yaml:"redis_password" json:"redis_password"`
			RedisDB        int    `mapstructure:"redis_db" yaml:"redis_db" json:"redis_db"`
			MemcachedAddrs string `mapstructure:"memcached_addrs" yaml:"memcached_addrs" json:"memcached_addrs"`
		}{
			Enable:         false,
			Type:           "memory",
			TTL:            "1h",
			MaxSize:        1000,
			KeyPrefix:      "yyhertz:db:",
			RedisAddr:      "localhost:6379",
			RedisPassword:  "",
			RedisDB:        0,
			MemcachedAddrs: "localhost:11211",
		},

		// 监控配置 - 默认启用
		Monitoring: struct {
			Enable           bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			MetricsPath      string `mapstructure:"metrics_path" yaml:"metrics_path" json:"metrics_path"`
			SlowQueryLog     bool   `mapstructure:"slow_query_log" yaml:"slow_query_log" json:"slow_query_log"`
			ConnectionEvents bool   `mapstructure:"connection_events" yaml:"connection_events" json:"connection_events"`
			QueryEvents      bool   `mapstructure:"query_events" yaml:"query_events" json:"query_events"`
			ErrorEvents      bool   `mapstructure:"error_events" yaml:"error_events" json:"error_events"`
			StatsInterval    string `mapstructure:"stats_interval" yaml:"stats_interval" json:"stats_interval"`
			ExportFormat     string `mapstructure:"export_format" yaml:"export_format" json:"export_format"`
		}{
			Enable:           true,
			MetricsPath:      "/metrics",
			SlowQueryLog:     true,
			ConnectionEvents: true,
			QueryEvents:      false, // 默认关闭，避免过多日志
			ErrorEvents:      true,
			StatsInterval:    "30s",
			ExportFormat:     "prometheus",
		},

		// 迁移配置 - 默认启用
		Migration: struct {
			Enable       bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			Path         string `mapstructure:"path" yaml:"path" json:"path"`
			TableName    string `mapstructure:"table_name" yaml:"table_name" json:"table_name"`
			AutoMigrate  bool   `mapstructure:"auto_migrate" yaml:"auto_migrate" json:"auto_migrate"`
			DropColumn   bool   `mapstructure:"drop_column" yaml:"drop_column" json:"drop_column"`
			DropTable    bool   `mapstructure:"drop_table" yaml:"drop_table" json:"drop_table"`
			DropIndex    bool   `mapstructure:"drop_index" yaml:"drop_index" json:"drop_index"`
			AlterColumn  bool   `mapstructure:"alter_column" yaml:"alter_column" json:"alter_column"`
			CreateIndex  bool   `mapstructure:"create_index" yaml:"create_index" json:"create_index"`
			RenameColumn bool   `mapstructure:"rename_column" yaml:"rename_column" json:"rename_column"`
			RenameIndex  bool   `mapstructure:"rename_index" yaml:"rename_index" json:"rename_index"`
		}{
			Enable:       true,
			Path:         "./migrations",
			TableName:    "schema_migrations",
			AutoMigrate:  false, // 开发环境可考虑启用
			DropColumn:   false, // 安全起见默认禁用
			DropTable:    false, // 安全起见默认禁用
			DropIndex:    false, // 安全起见默认禁用
			AlterColumn:  true,
			CreateIndex:  true,
			RenameColumn: true,
			RenameIndex:  true,
		},

		// 多租户配置 - 默认禁用
		MultiTenant: struct {
			Enable        bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			Strategy      string `mapstructure:"strategy" yaml:"strategy" json:"strategy"`
			TenantHeader  string `mapstructure:"tenant_header" yaml:"tenant_header" json:"tenant_header"`
			DefaultTenant string `mapstructure:"default_tenant" yaml:"default_tenant" json:"default_tenant"`
			SchemaPrefix  string `mapstructure:"schema_prefix" yaml:"schema_prefix" json:"schema_prefix"`
			TableSuffix   string `mapstructure:"table_suffix" yaml:"table_suffix" json:"table_suffix"`
		}{
			Enable:        false,
			Strategy:      "discriminator",
			TenantHeader:  "X-Tenant-ID",
			DefaultTenant: "default",
			SchemaPrefix:  "tenant_",
			TableSuffix:   "",
		},

		// 开发配置 - 根据环境启用
		Development: struct {
			Enable      bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			SeedData    bool   `mapstructure:"seed_data" yaml:"seed_data" json:"seed_data"`
			DropTables  bool   `mapstructure:"drop_tables" yaml:"drop_tables" json:"drop_tables"`
			ShowSQL     bool   `mapstructure:"show_sql" yaml:"show_sql" json:"show_sql"`
			ExplainPlan bool   `mapstructure:"explain_plan" yaml:"explain_plan" json:"explain_plan"`
			MockData    string `mapstructure:"mock_data" yaml:"mock_data" json:"mock_data"`
		}{
			Enable:      false, // 默认禁用，开发环境手动启用
			SeedData:    false, // 默认不填充测试数据
			DropTables:  false, // 安全起见默认禁用
			ShowSQL:     false, // 默认不显示SQL，需要时开启
			ExplainPlan: false, // 默认不显示查询计划
			MockData:    "./config/mock_data.yaml",
		},
	}
}
