package config

import (
	"github.com/spf13/viper"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	// 主数据库配置
	Primary struct {
		Driver                string `mapstructure:"driver" yaml:"driver" json:"driver"`                                           // mysql, postgres, sqlite, sqlserver
		DSN                   string `mapstructure:"dsn" yaml:"dsn" json:"dsn"`                                                    // 数据库连接字符串
		Host                  string `mapstructure:"host" yaml:"host" json:"host"`                                                 // 主机地址
		Port                  int    `mapstructure:"port" yaml:"port" json:"port"`                                                 // 端口
		Database              string `mapstructure:"database" yaml:"database" json:"database"`                                    // 数据库名
		Username              string `mapstructure:"username" yaml:"username" json:"username"`                                    // 用户名
		Password              string `mapstructure:"password" yaml:"password" json:"password"`                                    // 密码
		Charset               string `mapstructure:"charset" yaml:"charset" json:"charset"`                                       // 字符集
		Collation             string `mapstructure:"collation" yaml:"collation" json:"collation"`                                 // 排序规则
		Timezone              string `mapstructure:"timezone" yaml:"timezone" json:"timezone"`                                    // 时区
		MaxOpenConns          int    `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`                  // 最大打开连接数
		MaxIdleConns          int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`                  // 最大空闲连接数
		ConnMaxLifetime       string `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`         // 连接最大生存时间
		ConnMaxIdleTime       string `mapstructure:"conn_max_idle_time" yaml:"conn_max_idle_time" json:"conn_max_idle_time"`      // 连接最大空闲时间
		SlowQueryThreshold    string `mapstructure:"slow_query_threshold" yaml:"slow_query_threshold" json:"slow_query_threshold"` // 慢查询阈值
		LogLevel              string `mapstructure:"log_level" yaml:"log_level" json:"log_level"`                                 // 日志级别: silent, error, warn, info
		EnableMetrics         bool   `mapstructure:"enable_metrics" yaml:"enable_metrics" json:"enable_metrics"`                  // 启用性能监控
		EnableAutoMigration   bool   `mapstructure:"enable_auto_migration" yaml:"enable_auto_migration" json:"enable_auto_migration"` // 启用自动迁移
		MigrationTableName    string `mapstructure:"migration_table_name" yaml:"migration_table_name" json:"migration_table_name"` // 迁移表名
		SSLMode               string `mapstructure:"ssl_mode" yaml:"ssl_mode" json:"ssl_mode"`                                    // SSL模式: disable, require, verify-ca, verify-full
		SSLCert               string `mapstructure:"ssl_cert" yaml:"ssl_cert" json:"ssl_cert"`                                    // SSL证书路径
		SSLKey                string `mapstructure:"ssl_key" yaml:"ssl_key" json:"ssl_key"`                                       // SSL密钥路径
		SSLRootCert           string `mapstructure:"ssl_root_cert" yaml:"ssl_root_cert" json:"ssl_root_cert"`                     // SSL根证书路径
	} `mapstructure:"primary" yaml:"primary" json:"primary"`

	// 从数据库配置(读写分离)
	Replica struct {
		Enable                bool     `mapstructure:"enable" yaml:"enable" json:"enable"`                                          // 启用读写分离
		Hosts                 []string `mapstructure:"hosts" yaml:"hosts" json:"hosts"`                                             // 从库主机列表
		Driver                string   `mapstructure:"driver" yaml:"driver" json:"driver"`                                          // 数据库驱动
		Username              string   `mapstructure:"username" yaml:"username" json:"username"`                                   // 用户名
		Password              string   `mapstructure:"password" yaml:"password" json:"password"`                                   // 密码
		Database              string   `mapstructure:"database" yaml:"database" json:"database"`                                   // 数据库名
		MaxOpenConns          int      `mapstructure:"max_open_conns" yaml:"max_open_conns" json:"max_open_conns"`                 // 最大打开连接数
		MaxIdleConns          int      `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`                 // 最大空闲连接数
		ConnMaxLifetime       string   `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime" json:"conn_max_lifetime"`        // 连接最大生存时间
		LoadBalancingStrategy string   `mapstructure:"load_balancing_strategy" yaml:"load_balancing_strategy" json:"load_balancing_strategy"` // 负载均衡策略: round_robin, random, weighted
	} `mapstructure:"replica" yaml:"replica" json:"replica"`

	// GORM配置
	GORM struct {
		Enable                     bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                                        // 启用GORM
		DisableForeignKeyConstrain bool   `mapstructure:"disable_foreign_key_constrain" yaml:"disable_foreign_key_constrain" json:"disable_foreign_key_constrain"` // 禁用外键约束
		SkipDefaultTransaction     bool   `mapstructure:"skip_default_transaction" yaml:"skip_default_transaction" json:"skip_default_transaction"`  // 跳过默认事务
		FullSaveAssociations       bool   `mapstructure:"full_save_associations" yaml:"full_save_associations" json:"full_save_associations"`        // 完整保存关联
		DryRun                     bool   `mapstructure:"dry_run" yaml:"dry_run" json:"dry_run"`                                                     // 仅生成SQL不执行
		PrepareStmt                bool   `mapstructure:"prepare_stmt" yaml:"prepare_stmt" json:"prepare_stmt"`                                      // 启用预编译语句
		DisableNestedTransaction   bool   `mapstructure:"disable_nested_transaction" yaml:"disable_nested_transaction" json:"disable_nested_transaction"` // 禁用嵌套事务
		AllowGlobalUpdate          bool   `mapstructure:"allow_global_update" yaml:"allow_global_update" json:"allow_global_update"`                // 允许全局更新
		QueryFields                bool   `mapstructure:"query_fields" yaml:"query_fields" json:"query_fields"`                                      // 查询时选择字段
		CreateBatchSize            int    `mapstructure:"create_batch_size" yaml:"create_batch_size" json:"create_batch_size"`                       // 批量创建大小
		NamingStrategy             string `mapstructure:"naming_strategy" yaml:"naming_strategy" json:"naming_strategy"`                             // 命名策略: snake_case, camel_case
		TablePrefix                string `mapstructure:"table_prefix" yaml:"table_prefix" json:"table_prefix"`                                      // 表前缀
		SingularTable              bool   `mapstructure:"singular_table" yaml:"singular_table" json:"singular_table"`                               // 使用单数表名
	} `mapstructure:"gorm" yaml:"gorm" json:"gorm"`

	// MyBatis配置
	MyBatis struct {
		Enable          bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                   // 启用MyBatis
		ConfigFile      string `mapstructure:"config_file" yaml:"config_file" json:"config_file"`                   // MyBatis配置文件路径
		MapperLocations string `mapstructure:"mapper_locations" yaml:"mapper_locations" json:"mapper_locations"`    // Mapper文件位置
		TypeAliasesPath string `mapstructure:"type_aliases_path" yaml:"type_aliases_path" json:"type_aliases_path"`  // 类型别名路径
		CacheEnabled    bool   `mapstructure:"cache_enabled" yaml:"cache_enabled" json:"cache_enabled"`             // 启用缓存
		LazyLoading     bool   `mapstructure:"lazy_loading" yaml:"lazy_loading" json:"lazy_loading"`                // 延迟加载
		LogImpl         string `mapstructure:"log_impl" yaml:"log_impl" json:"log_impl"`                            // 日志实现: STDOUT_LOGGING, LOG4J, SLF4J
		MapUnderscoreMap bool  `mapstructure:"map_underscore_map" yaml:"map_underscore_map" json:"map_underscore_map"` // 下划线到驼峰映射
	} `mapstructure:"mybatis" yaml:"mybatis" json:"mybatis"`

	// 连接池配置
	Pool struct {
		Enable              bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                        // 启用连接池
		Type                string `mapstructure:"type" yaml:"type" json:"type"`                                              // 连接池类型: default, hikari, druid
		MaxActiveConns      int    `mapstructure:"max_active_conns" yaml:"max_active_conns" json:"max_active_conns"`          // 最大活跃连接数
		MaxIdleConns        int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`                // 最大空闲连接数
		MinIdleConns        int    `mapstructure:"min_idle_conns" yaml:"min_idle_conns" json:"min_idle_conns"`                // 最小空闲连接数
		MaxWaitTime         string `mapstructure:"max_wait_time" yaml:"max_wait_time" json:"max_wait_time"`                   // 最大等待时间
		TimeBetweenEviction string `mapstructure:"time_between_eviction" yaml:"time_between_eviction" json:"time_between_eviction"` // 连接回收间隔
		MinEvictableTime    string `mapstructure:"min_evictable_time" yaml:"min_evictable_time" json:"min_evictable_time"`    // 最小可回收空闲时间
		TestOnBorrow        bool   `mapstructure:"test_on_borrow" yaml:"test_on_borrow" json:"test_on_borrow"`                // 借用时测试
		TestOnReturn        bool   `mapstructure:"test_on_return" yaml:"test_on_return" json:"test_on_return"`                // 归还时测试
		TestWhileIdle       bool   `mapstructure:"test_while_idle" yaml:"test_while_idle" json:"test_while_idle"`             // 空闲时测试
		ValidationQuery     string `mapstructure:"validation_query" yaml:"validation_query" json:"validation_query"`          // 验证查询SQL
	} `mapstructure:"pool" yaml:"pool" json:"pool"`

	// 缓存配置
	Cache struct {
		Enable         bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                  // 启用查询缓存
		Type           string `mapstructure:"type" yaml:"type" json:"type"`                                        // 缓存类型: memory, redis, memcached
		TTL            string `mapstructure:"ttl" yaml:"ttl" json:"ttl"`                                           // 缓存生存时间
		MaxSize        int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"`                            // 最大缓存大小
		KeyPrefix      string `mapstructure:"key_prefix" yaml:"key_prefix" json:"key_prefix"`                      // 缓存键前缀
		RedisAddr      string `mapstructure:"redis_addr" yaml:"redis_addr" json:"redis_addr"`                      // Redis地址
		RedisPassword  string `mapstructure:"redis_password" yaml:"redis_password" json:"redis_password"`          // Redis密码
		RedisDB        int    `mapstructure:"redis_db" yaml:"redis_db" json:"redis_db"`                            // Redis数据库
		MemcachedAddrs string `mapstructure:"memcached_addrs" yaml:"memcached_addrs" json:"memcached_addrs"`       // Memcached地址列表
	} `mapstructure:"cache" yaml:"cache" json:"cache"`

	// 监控配置
	Monitoring struct {
		Enable           bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                        // 启用监控
		MetricsPath      string `mapstructure:"metrics_path" yaml:"metrics_path" json:"metrics_path"`                     // 监控指标路径
		SlowQueryLog     bool   `mapstructure:"slow_query_log" yaml:"slow_query_log" json:"slow_query_log"`               // 启用慢查询日志
		ConnectionEvents bool   `mapstructure:"connection_events" yaml:"connection_events" json:"connection_events"`       // 记录连接事件
		QueryEvents      bool   `mapstructure:"query_events" yaml:"query_events" json:"query_events"`                     // 记录查询事件
		ErrorEvents      bool   `mapstructure:"error_events" yaml:"error_events" json:"error_events"`                     // 记录错误事件
		StatsInterval    string `mapstructure:"stats_interval" yaml:"stats_interval" json:"stats_interval"`               // 统计间隔
		ExportFormat     string `mapstructure:"export_format" yaml:"export_format" json:"export_format"`                  // 导出格式: prometheus, json, text
	} `mapstructure:"monitoring" yaml:"monitoring" json:"monitoring"`

	// 迁移配置
	Migration struct {
		Enable        bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                // 启用数据库迁移
		Path          string `mapstructure:"path" yaml:"path" json:"path"`                                      // 迁移文件路径
		TableName     string `mapstructure:"table_name" yaml:"table_name" json:"table_name"`                   // 迁移记录表名
		AutoMigrate   bool   `mapstructure:"auto_migrate" yaml:"auto_migrate" json:"auto_migrate"`             // 自动迁移模型
		DropColumn    bool   `mapstructure:"drop_column" yaml:"drop_column" json:"drop_column"`                // 允许删除列
		DropTable     bool   `mapstructure:"drop_table" yaml:"drop_table" json:"drop_table"`                   // 允许删除表
		DropIndex     bool   `mapstructure:"drop_index" yaml:"drop_index" json:"drop_index"`                   // 允许删除索引
		AlterColumn   bool   `mapstructure:"alter_column" yaml:"alter_column" json:"alter_column"`             // 允许修改列
		CreateIndex   bool   `mapstructure:"create_index" yaml:"create_index" json:"create_index"`             // 允许创建索引
		RenameColumn bool   `mapstructure:"rename_column" yaml:"rename_column" json:"rename_column"`          // 允许重命名列
		RenameIndex   bool   `mapstructure:"rename_index" yaml:"rename_index" json:"rename_index"`             // 允许重命名索引
	} `mapstructure:"migration" yaml:"migration" json:"migration"`

	// 多租户配置
	MultiTenant struct {
		Enable        bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                // 启用多租户
		Strategy      string `mapstructure:"strategy" yaml:"strategy" json:"strategy"`                          // 策略: schema, database, discriminator
		TenantHeader  string `mapstructure:"tenant_header" yaml:"tenant_header" json:"tenant_header"`          // 租户请求头
		DefaultTenant string `mapstructure:"default_tenant" yaml:"default_tenant" json:"default_tenant"`       // 默认租户
		SchemaPrefix  string `mapstructure:"schema_prefix" yaml:"schema_prefix" json:"schema_prefix"`          // Schema前缀
		TableSuffix   string `mapstructure:"table_suffix" yaml:"table_suffix" json:"table_suffix"`             // 表后缀
	} `mapstructure:"multi_tenant" yaml:"multi_tenant" json:"multi_tenant"`

	// 开发配置
	Development struct {
		Enable      bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                            // 启用开发模式
		SeedData    bool   `mapstructure:"seed_data" yaml:"seed_data" json:"seed_data"`                   // 自动填充测试数据
		DropTables  bool   `mapstructure:"drop_tables" yaml:"drop_tables" json:"drop_tables"`             // 启动时删除所有表
		ShowSQL     bool   `mapstructure:"show_sql" yaml:"show_sql" json:"show_sql"`                      // 显示SQL语句
		ExplainPlan bool   `mapstructure:"explain_plan" yaml:"explain_plan" json:"explain_plan"`          // 显示查询计划
		MockData    string `mapstructure:"mock_data" yaml:"mock_data" json:"mock_data"`                   // 模拟数据配置文件
	} `mapstructure:"development" yaml:"development" json:"development"`
}

// GetConfigName 实现 ConfigInterface 接口
func (c DatabaseConfig) GetConfigName() string {
	return DatabaseConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c DatabaseConfig) SetDefaults(v *viper.Viper) {
	// 主数据库默认配置
	v.SetDefault("primary.driver", "mysql")
	v.SetDefault("primary.host", "localhost")
	v.SetDefault("primary.port", 3306)
	v.SetDefault("primary.database", "yyhertz")
	v.SetDefault("primary.username", "root")
	v.SetDefault("primary.password", "")
	v.SetDefault("primary.charset", "utf8mb4")
	v.SetDefault("primary.collation", "utf8mb4_unicode_ci")
	v.SetDefault("primary.timezone", "Local")
	v.SetDefault("primary.max_open_conns", 100)
	v.SetDefault("primary.max_idle_conns", 10)
	v.SetDefault("primary.conn_max_lifetime", "1h")
	v.SetDefault("primary.conn_max_idle_time", "30m")
	v.SetDefault("primary.slow_query_threshold", "200ms")
	v.SetDefault("primary.log_level", "warn")
	v.SetDefault("primary.enable_metrics", true)
	v.SetDefault("primary.enable_auto_migration", false)
	v.SetDefault("primary.migration_table_name", "schema_migrations")
	v.SetDefault("primary.ssl_mode", "disable")

	// 从数据库默认配置
	v.SetDefault("replica.enable", false)
	v.SetDefault("replica.driver", "mysql")
	v.SetDefault("replica.max_open_conns", 50)
	v.SetDefault("replica.max_idle_conns", 10)
	v.SetDefault("replica.conn_max_lifetime", "1h")
	v.SetDefault("replica.load_balancing_strategy", "round_robin")

	// GORM默认配置
	v.SetDefault("gorm.enable", true)
	v.SetDefault("gorm.disable_foreign_key_constrain", false)
	v.SetDefault("gorm.skip_default_transaction", false)
	v.SetDefault("gorm.full_save_associations", false)
	v.SetDefault("gorm.dry_run", false)
	v.SetDefault("gorm.prepare_stmt", true)
	v.SetDefault("gorm.disable_nested_transaction", false)
	v.SetDefault("gorm.allow_global_update", false)
	v.SetDefault("gorm.query_fields", true)
	v.SetDefault("gorm.create_batch_size", 1000)
	v.SetDefault("gorm.naming_strategy", "snake_case")
	v.SetDefault("gorm.table_prefix", "")
	v.SetDefault("gorm.singular_table", false)

	// MyBatis默认配置
	v.SetDefault("mybatis.enable", false)
	v.SetDefault("mybatis.config_file", "./config/mybatis-config.xml")
	v.SetDefault("mybatis.mapper_locations", "./mappers/*.xml")
	v.SetDefault("mybatis.type_aliases_path", "./models")
	v.SetDefault("mybatis.cache_enabled", true)
	v.SetDefault("mybatis.lazy_loading", false)
	v.SetDefault("mybatis.log_impl", "STDOUT_LOGGING")
	v.SetDefault("mybatis.map_underscore_map", true)

	// 连接池默认配置
	v.SetDefault("pool.enable", true)
	v.SetDefault("pool.type", "default")
	v.SetDefault("pool.max_active_conns", 100)
	v.SetDefault("pool.max_idle_conns", 10)
	v.SetDefault("pool.min_idle_conns", 5)
	v.SetDefault("pool.max_wait_time", "30s")
	v.SetDefault("pool.time_between_eviction", "30s")
	v.SetDefault("pool.min_evictable_time", "5m")
	v.SetDefault("pool.test_on_borrow", true)
	v.SetDefault("pool.test_on_return", false)
	v.SetDefault("pool.test_while_idle", true)
	v.SetDefault("pool.validation_query", "SELECT 1")

	// 缓存默认配置
	v.SetDefault("cache.enable", false)
	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.ttl", "1h")
	v.SetDefault("cache.max_size", 1000)
	v.SetDefault("cache.key_prefix", "yyhertz:db:")
	v.SetDefault("cache.redis_addr", "localhost:6379")
	v.SetDefault("cache.redis_password", "")
	v.SetDefault("cache.redis_db", 0)

	// 监控默认配置
	v.SetDefault("monitoring.enable", true)
	v.SetDefault("monitoring.metrics_path", "/metrics")
	v.SetDefault("monitoring.slow_query_log", true)
	v.SetDefault("monitoring.connection_events", true)
	v.SetDefault("monitoring.query_events", false)
	v.SetDefault("monitoring.error_events", true)
	v.SetDefault("monitoring.stats_interval", "30s")
	v.SetDefault("monitoring.export_format", "prometheus")

	// 迁移默认配置
	v.SetDefault("migration.enable", true)
	v.SetDefault("migration.path", "./migrations")
	v.SetDefault("migration.table_name", "schema_migrations")
	v.SetDefault("migration.auto_migrate", false)
	v.SetDefault("migration.drop_column", false)
	v.SetDefault("migration.drop_table", false)
	v.SetDefault("migration.drop_index", false)
	v.SetDefault("migration.alter_column", true)
	v.SetDefault("migration.create_index", true)
	v.SetDefault("migration.rename_column", true)
	v.SetDefault("migration.rename_index", true)

	// 多租户默认配置
	v.SetDefault("multi_tenant.enable", false)
	v.SetDefault("multi_tenant.strategy", "discriminator")
	v.SetDefault("multi_tenant.tenant_header", "X-Tenant-ID")
	v.SetDefault("multi_tenant.default_tenant", "default")
	v.SetDefault("multi_tenant.schema_prefix", "tenant_")
	v.SetDefault("multi_tenant.table_suffix", "")

	// 开发默认配置
	v.SetDefault("development.enable", false)
	v.SetDefault("development.seed_data", false)
	v.SetDefault("development.drop_tables", false)
	v.SetDefault("development.show_sql", false)
	v.SetDefault("development.explain_plan", false)
	v.SetDefault("development.mock_data", "./config/mock_data.yaml")
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c DatabaseConfig) GenerateDefaultContent() string {
	return `# YYHertz Database Configuration
# 数据库配置文件

# 主数据库配置
primary:
  driver: "mysql"                                 # 数据库驱动: mysql, postgres, sqlite, sqlserver
  dsn: ""                                        # 完整连接字符串(优先级高于单独配置)
  host: "localhost"                              # 数据库主机地址
  port: 3306                                     # 数据库端口
  database: "yyhertz"                            # 数据库名
  username: "root"                               # 用户名
  password: ""                                   # 密码
  charset: "utf8mb4"                             # 字符集
  collation: "utf8mb4_unicode_ci"                # 排序规则
  timezone: "Local"                              # 时区设置
  max_open_conns: 100                            # 最大打开连接数
  max_idle_conns: 10                             # 最大空闲连接数
  conn_max_lifetime: "1h"                        # 连接最大生存时间
  conn_max_idle_time: "30m"                      # 连接最大空闲时间
  slow_query_threshold: "200ms"                  # 慢查询阈值
  log_level: "warn"                              # 日志级别: silent, error, warn, info
  enable_metrics: true                           # 启用性能监控
  enable_auto_migration: false                   # 启用自动迁移
  migration_table_name: "schema_migrations"      # 迁移记录表名
  ssl_mode: "disable"                            # SSL模式: disable, require, verify-ca, verify-full
  ssl_cert: ""                                   # SSL证书路径
  ssl_key: ""                                    # SSL密钥路径
  ssl_root_cert: ""                              # SSL根证书路径

# 从数据库配置(读写分离)
replica:
  enable: false                                  # 启用读写分离
  hosts: ["localhost:3307"]                      # 从库主机列表
  driver: "mysql"                                # 数据库驱动
  username: "root"                               # 用户名
  password: ""                                   # 密码
  database: "yyhertz"                            # 数据库名
  max_open_conns: 50                             # 最大打开连接数
  max_idle_conns: 10                             # 最大空闲连接数
  conn_max_lifetime: "1h"                        # 连接最大生存时间
  load_balancing_strategy: "round_robin"         # 负载均衡策略: round_robin, random, weighted

# GORM配置
gorm:
  enable: true                                   # 启用GORM
  disable_foreign_key_constrain: false          # 禁用外键约束
  skip_default_transaction: false               # 跳过默认事务
  full_save_associations: false                 # 完整保存关联
  dry_run: false                                # 仅生成SQL不执行
  prepare_stmt: true                            # 启用预编译语句
  disable_nested_transaction: false             # 禁用嵌套事务
  allow_global_update: false                    # 允许全局更新
  query_fields: true                            # 查询时选择字段
  create_batch_size: 1000                       # 批量创建大小
  naming_strategy: "snake_case"                 # 命名策略: snake_case, camel_case
  table_prefix: ""                              # 表前缀
  singular_table: false                         # 使用单数表名

# MyBatis配置
mybatis:
  enable: false                                 # 启用MyBatis
  config_file: "./config/mybatis-config.xml"   # MyBatis配置文件路径
  mapper_locations: "./mappers/*.xml"           # Mapper XML文件位置
  type_aliases_path: "./models"                 # 类型别名路径
  cache_enabled: true                           # 启用缓存
  lazy_loading: false                           # 延迟加载
  log_impl: "STDOUT_LOGGING"                    # 日志实现: STDOUT_LOGGING, LOG4J, SLF4J
  map_underscore_map: true                      # 下划线到驼峰映射

# 连接池配置
pool:
  enable: true                                  # 启用连接池
  type: "default"                               # 连接池类型: default, hikari, druid
  max_active_conns: 100                         # 最大活跃连接数
  max_idle_conns: 10                            # 最大空闲连接数
  min_idle_conns: 5                             # 最小空闲连接数
  max_wait_time: "30s"                          # 最大等待时间
  time_between_eviction: "30s"                  # 连接回收间隔
  min_evictable_time: "5m"                      # 最小可回收空闲时间
  test_on_borrow: true                          # 借用时测试连接
  test_on_return: false                         # 归还时测试连接
  test_while_idle: true                         # 空闲时测试连接
  validation_query: "SELECT 1"                  # 验证查询SQL

# 缓存配置
cache:
  enable: false                                 # 启用查询缓存
  type: "memory"                                # 缓存类型: memory, redis, memcached
  ttl: "1h"                                     # 缓存生存时间
  max_size: 1000                                # 最大缓存大小
  key_prefix: "yyhertz:db:"                     # 缓存键前缀
  redis_addr: "localhost:6379"                  # Redis地址
  redis_password: ""                            # Redis密码
  redis_db: 0                                   # Redis数据库编号
  memcached_addrs: "localhost:11211"            # Memcached地址列表

# 监控配置
monitoring:
  enable: true                                  # 启用数据库监控
  metrics_path: "/metrics"                      # 监控指标路径
  slow_query_log: true                          # 启用慢查询日志
  connection_events: true                       # 记录连接事件
  query_events: false                           # 记录查询事件
  error_events: true                            # 记录错误事件
  stats_interval: "30s"                         # 统计信息收集间隔
  export_format: "prometheus"                   # 导出格式: prometheus, json, text

# 数据库迁移配置
migration:
  enable: true                                  # 启用数据库迁移
  path: "./migrations"                          # 迁移文件路径
  table_name: "schema_migrations"               # 迁移记录表名
  auto_migrate: false                           # 自动迁移模型
  drop_column: false                            # 允许删除列
  drop_table: false                             # 允许删除表
  drop_index: false                             # 允许删除索引
  alter_column: true                            # 允许修改列
  create_index: true                            # 允许创建索引
  rename_column: true                           # 允许重命名列
  rename_index: true                            # 允许重命名索引

# 多租户配置
multi_tenant:
  enable: false                                 # 启用多租户支持
  strategy: "discriminator"                     # 策略: schema, database, discriminator
  tenant_header: "X-Tenant-ID"                 # 租户请求头
  default_tenant: "default"                     # 默认租户
  schema_prefix: "tenant_"                      # Schema前缀
  table_suffix: ""                              # 表后缀

# 开发环境配置
development:
  enable: false                                 # 启用开发模式
  seed_data: false                              # 自动填充测试数据
  drop_tables: false                            # 启动时删除所有表
  show_sql: false                               # 显示SQL语句
  explain_plan: false                           # 显示查询执行计划
  mock_data: "./config/mock_data.yaml"          # 模拟数据配置文件

# 常用连接字符串示例:
# MySQL: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
# PostgreSQL: "host=localhost user=postgres password=password dbname=yyhertz port=5432 sslmode=disable TimeZone=Asia/Shanghai" 
# SQLite: "./database.db"
# SQL Server: "sqlserver://user:password@localhost:1433?database=yyhertz"
`
}