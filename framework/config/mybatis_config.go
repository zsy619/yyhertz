package config

import (
	"github.com/spf13/viper"
)

// MyBatisConfig MyBatis配置结构
type MyBatisConfig struct {
	// 基础配置
	Basic struct {
		Enable             bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                           // 是否启用MyBatis
		ConfigFile         string `mapstructure:"config_file" yaml:"config_file" json:"config_file"`                            // MyBatis配置文件路径
		MapperLocations    string `mapstructure:"mapper_locations" yaml:"mapper_locations" json:"mapper_locations"`             // Mapper文件位置
		TypeAliasesPackage string `mapstructure:"type_aliases_package" yaml:"type_aliases_package" json:"type_aliases_package"` // 类型别名包
	} `mapstructure:"basic" yaml:"basic" json:"basic"`

	// 缓存配置
	Cache struct {
		Enable    bool   `mapstructure:"enable" yaml:"enable" json:"enable"`             // 是否启用缓存
		Type      string `mapstructure:"type" yaml:"type" json:"type"`                   // 缓存类型: memory, redis
		TTL       int    `mapstructure:"ttl" yaml:"ttl" json:"ttl"`                      // 缓存生存时间(秒)
		MaxSize   int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"`       // 最大缓存条目数
		RedisAddr string `mapstructure:"redis_addr" yaml:"redis_addr" json:"redis_addr"` // Redis地址
		RedisDB   int    `mapstructure:"redis_db" yaml:"redis_db" json:"redis_db"`       // Redis数据库
	} `mapstructure:"cache" yaml:"cache" json:"cache"`

	// 日志配置
	Logging struct {
		Enable    bool   `mapstructure:"enable" yaml:"enable" json:"enable"`             // 是否启用SQL日志
		Level     string `mapstructure:"level" yaml:"level" json:"level"`                // 日志级别: debug, info, warn, error
		ShowSQL   bool   `mapstructure:"show_sql" yaml:"show_sql" json:"show_sql"`       // 是否显示SQL语句
		SlowQuery int    `mapstructure:"slow_query" yaml:"slow_query" json:"slow_query"` // 慢查询阈值(毫秒)
	} `mapstructure:"logging" yaml:"logging" json:"logging"`

	// 性能配置
	Performance struct {
		LazyLoading         bool   `mapstructure:"lazy_loading" yaml:"lazy_loading" json:"lazy_loading"`                            // 延迟加载
		MultipleResultSets  bool   `mapstructure:"multiple_result_sets" yaml:"multiple_result_sets" json:"multiple_result_sets"`    // 多结果集
		UseColumnLabel      bool   `mapstructure:"use_column_label" yaml:"use_column_label" json:"use_column_label"`                // 使用列标签
		UseGeneratedKeys    bool   `mapstructure:"use_generated_keys" yaml:"use_generated_keys" json:"use_generated_keys"`          // 使用生成的键
		AutoMappingBehavior string `mapstructure:"auto_mapping_behavior" yaml:"auto_mapping_behavior" json:"auto_mapping_behavior"` // 自动映射行为: NONE, PARTIAL, FULL
	} `mapstructure:"performance" yaml:"performance" json:"performance"`

	// 事务配置
	Transaction struct {
		DefaultTimeout int    `mapstructure:"default_timeout" yaml:"default_timeout" json:"default_timeout"` // 默认事务超时时间(秒)
		IsolationLevel string `mapstructure:"isolation_level" yaml:"isolation_level" json:"isolation_level"` // 事务隔离级别
		AutoCommit     bool   `mapstructure:"auto_commit" yaml:"auto_commit" json:"auto_commit"`             // 自动提交
	} `mapstructure:"transaction" yaml:"transaction" json:"transaction"`

	// 插件配置
	Plugins struct {
		Interceptors []string `mapstructure:"interceptors" yaml:"interceptors" json:"interceptors"` // 拦截器列表
		PageHelper   struct {
			Enable                  bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                                          // 启用分页插件
			HelperDialect           string `mapstructure:"helper_dialect" yaml:"helper_dialect" json:"helper_dialect"`                                  // 分页方言
			ReasonableEnabled       bool   `mapstructure:"reasonable_enabled" yaml:"reasonable_enabled" json:"reasonable_enabled"`                      // 合理化参数
			SupportMethodsArguments bool   `mapstructure:"support_methods_arguments" yaml:"support_methods_arguments" json:"support_methods_arguments"` // 支持方法参数
		} `mapstructure:"page_helper" yaml:"page_helper" json:"page_helper"`
	} `mapstructure:"plugins" yaml:"plugins" json:"plugins"`
}

// GetConfigName 实现 ConfigInterface 接口
func (c MyBatisConfig) GetConfigName() string {
	return MyBatisConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c MyBatisConfig) SetDefaults(v *viper.Viper) {
	// 基础配置默认值
	v.SetDefault("basic.enable", true)
	v.SetDefault("basic.config_file", "./cnf/mybatis-config.xml")
	v.SetDefault("basic.mapper_locations", "./mappers/*.xml")
	v.SetDefault("basic.type_aliases_package", "")

	// 缓存配置默认值
	v.SetDefault("cache.enable", false)
	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.ttl", 3600)
	v.SetDefault("cache.max_size", 1000)
	v.SetDefault("cache.redis_addr", "localhost:6379")
	v.SetDefault("cache.redis_db", 0)

	// 日志配置默认值
	v.SetDefault("logging.enable", true)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.show_sql", false)
	v.SetDefault("logging.slow_query", 1000)

	// 性能配置默认值
	v.SetDefault("performance.lazy_loading", true)
	v.SetDefault("performance.multiple_result_sets", true)
	v.SetDefault("performance.use_column_label", true)
	v.SetDefault("performance.use_generated_keys", false)
	v.SetDefault("performance.auto_mapping_behavior", "PARTIAL")

	// 事务配置默认值
	v.SetDefault("transaction.default_timeout", 30)
	v.SetDefault("transaction.isolation_level", "READ_COMMITTED")
	v.SetDefault("transaction.auto_commit", true)

	// 插件配置默认值
	v.SetDefault("plugins.interceptors", []string{})
	v.SetDefault("plugins.page_helper.enable", false)
	v.SetDefault("plugins.page_helper.helper_dialect", "mysql")
	v.SetDefault("plugins.page_helper.reasonable_enabled", true)
	v.SetDefault("plugins.page_helper.support_methods_arguments", true)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c MyBatisConfig) GenerateDefaultContent() string {
	return `# YYHertz MyBatis Configuration
# MyBatis ORM框架配置文件

# 基础配置
basic:
  enable: true                                    # 是否启用MyBatis
  config_file: "./config/mybatis-config.xml"     # MyBatis配置文件路径
  mapper_locations: "./mappers/*.xml"            # Mapper文件位置
  type_aliases_package: ""                       # 类型别名包

# 缓存配置
cache:
  enable: false                                 # 是否启用缓存
  type: "memory"                                # 缓存类型: memory, redis
  ttl: 3600                                     # 缓存生存时间(秒)
  max_size: 1000                                # 最大缓存条目数
  redis_addr: "localhost:6379"                  # Redis地址
  redis_db: 0                                   # Redis数据库

# 日志配置
logging:
  enable: true                                  # 是否启用SQL日志
  level: "info"                                 # 日志级别: debug, info, warn, error
  show_sql: false                               # 是否显示SQL语句
  slow_query: 1000                              # 慢查询阈值(毫秒)

# 性能配置
performance:
  lazy_loading: true                            # 延迟加载
  multiple_result_sets: true                    # 多结果集
  use_column_label: true                        # 使用列标签
  use_generated_keys: false                     # 使用生成的键
  auto_mapping_behavior: "PARTIAL"              # 自动映射行为: NONE, PARTIAL, FULL

# 事务配置
transaction:
  default_timeout: 30                           # 默认事务超时时间(秒)
  isolation_level: "READ_COMMITTED"             # 事务隔离级别
  auto_commit: true                             # 自动提交

# 插件配置
plugins:
  interceptors: []                              # 拦截器列表
  page_helper:                                  # 分页插件配置
    enable: false                               # 启用分页插件
    helper_dialect: "mysql"                     # 分页方言
    reasonable_enabled: true                    # 合理化参数
    support_methods_arguments: true             # 支持方法参数
`
}
