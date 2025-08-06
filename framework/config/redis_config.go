package config

import (
	"github.com/spf13/viper"
)

// RedisConfig Redis配置结构
type RedisConfig struct {
	// 基础连接配置
	Primary struct {
		Host     string `mapstructure:"host" yaml:"host" json:"host"`             // Redis主机地址
		Port     int    `mapstructure:"port" yaml:"port" json:"port"`             // Redis端口
		Password string `mapstructure:"password" yaml:"password" json:"password"` // Redis密码
		Database int    `mapstructure:"database" yaml:"database" json:"database"` // Redis数据库索引
		Username string `mapstructure:"username" yaml:"username" json:"username"` // Redis用户名(Redis 6.0+)
	} `mapstructure:"primary" yaml:"primary" json:"primary"`

	// 连接池配置
	Pool struct {
		MaxIdle        int    `mapstructure:"max_idle" yaml:"max_idle" json:"max_idle"`                         // 最大空闲连接数
		MaxActive      int    `mapstructure:"max_active" yaml:"max_active" json:"max_active"`                   // 最大活跃连接数
		IdleTimeout    string `mapstructure:"idle_timeout" yaml:"idle_timeout" json:"idle_timeout"`             // 空闲连接超时时间
		MaxConnTimeout string `mapstructure:"max_conn_timeout" yaml:"max_conn_timeout" json:"max_conn_timeout"` // 最大连接超时时间
		TestOnBorrow   bool   `mapstructure:"test_on_borrow" yaml:"test_on_borrow" json:"test_on_borrow"`       // 获取连接时测试
		Wait           bool   `mapstructure:"wait" yaml:"wait" json:"wait"`                                     // 连接池满时是否等待
	} `mapstructure:"pool" yaml:"pool" json:"pool"`

	// 集群配置
	Cluster struct {
		Enable         bool     `mapstructure:"enable" yaml:"enable" json:"enable"`                               // 启用Redis集群
		Nodes          []string `mapstructure:"nodes" yaml:"nodes" json:"nodes"`                                  // 集群节点列表
		Password       string   `mapstructure:"password" yaml:"password" json:"password"`                         // 集群密码
		ReadOnly       bool     `mapstructure:"read_only" yaml:"read_only" json:"read_only"`                      // 只读模式
		RouteByLatency bool     `mapstructure:"route_by_latency" yaml:"route_by_latency" json:"route_by_latency"` // 按延迟路由
		RouteRandomly  bool     `mapstructure:"route_randomly" yaml:"route_randomly" json:"route_randomly"`       // 随机路由
	} `mapstructure:"cluster" yaml:"cluster" json:"cluster"`

	// 哨兵配置
	Sentinel struct {
		Enable       bool     `mapstructure:"enable" yaml:"enable" json:"enable"`                      // 启用Redis哨兵
		Addresses    []string `mapstructure:"addresses" yaml:"addresses" json:"addresses"`             // 哨兵地址列表
		MasterName   string   `mapstructure:"master_name" yaml:"master_name" json:"master_name"`       // 主节点名称
		Password     string   `mapstructure:"password" yaml:"password" json:"password"`                // 哨兵密码
		Database     int      `mapstructure:"database" yaml:"database" json:"database"`                // 数据库索引
		DialTimeout  string   `mapstructure:"dial_timeout" yaml:"dial_timeout" json:"dial_timeout"`    // 连接超时
		ReadTimeout  string   `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`    // 读取超时
		WriteTimeout string   `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"` // 写入超时
	} `mapstructure:"sentinel" yaml:"sentinel" json:"sentinel"`

	// 性能配置
	Performance struct {
		DialTimeout        string `mapstructure:"dial_timeout" yaml:"dial_timeout" json:"dial_timeout"`                         // 连接超时
		ReadTimeout        string `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`                         // 读取超时
		WriteTimeout       string `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"`                      // 写入超时
		PoolTimeout        string `mapstructure:"pool_timeout" yaml:"pool_timeout" json:"pool_timeout"`                         // 连接池获取超时
		IdleCheckFrequency string `mapstructure:"idle_check_frequency" yaml:"idle_check_frequency" json:"idle_check_frequency"` // 空闲检查频率
		MaxRetries         int    `mapstructure:"max_retries" yaml:"max_retries" json:"max_retries"`                            // 最大重试次数
		MinRetryBackoff    string `mapstructure:"min_retry_backoff" yaml:"min_retry_backoff" json:"min_retry_backoff"`          // 最小重试间隔
		MaxRetryBackoff    string `mapstructure:"max_retry_backoff" yaml:"max_retry_backoff" json:"max_retry_backoff"`          // 最大重试间隔
	} `mapstructure:"performance" yaml:"performance" json:"performance"`

	// TLS配置
	TLS struct {
		Enable             bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                                           // 启用TLS
		CertFile           string `mapstructure:"cert_file" yaml:"cert_file" json:"cert_file"`                                  // 证书文件路径
		KeyFile            string `mapstructure:"key_file" yaml:"key_file" json:"key_file"`                                     // 私钥文件路径
		CAFile             string `mapstructure:"ca_file" yaml:"ca_file" json:"ca_file"`                                        // CA证书文件路径
		InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify" yaml:"insecure_skip_verify" json:"insecure_skip_verify"` // 跳过证书验证
		ServerName         string `mapstructure:"server_name" yaml:"server_name" json:"server_name"`                            // 服务器名称
	} `mapstructure:"tls" yaml:"tls" json:"tls"`

	// 监控配置
	Monitoring struct {
		Enable        bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                            // 启用监控
		MetricsPath   string `mapstructure:"metrics_path" yaml:"metrics_path" json:"metrics_path"`          // 监控指标路径
		SlowLogEnable bool   `mapstructure:"slow_log_enable" yaml:"slow_log_enable" json:"slow_log_enable"` // 启用慢日志
		SlowThreshold string `mapstructure:"slow_threshold" yaml:"slow_threshold" json:"slow_threshold"`    // 慢查询阈值
		StatsPeriod   string `mapstructure:"stats_period" yaml:"stats_period" json:"stats_period"`          // 统计周期
	} `mapstructure:"monitoring" yaml:"monitoring" json:"monitoring"`

	// 缓存配置
	Cache struct {
		DefaultTTL   string `mapstructure:"default_ttl" yaml:"default_ttl" json:"default_ttl"`          // 默认TTL
		KeyPrefix    string `mapstructure:"key_prefix" yaml:"key_prefix" json:"key_prefix"`             // 键前缀
		Serializer   string `mapstructure:"serializer" yaml:"serializer" json:"serializer"`             // 序列化器: json, msgpack, gob
		Compression  bool   `mapstructure:"compression" yaml:"compression" json:"compression"`          // 启用压缩
		MaxKeyLength int    `mapstructure:"max_key_length" yaml:"max_key_length" json:"max_key_length"` // 最大键长度
		MaxValueSize int    `mapstructure:"max_value_size" yaml:"max_value_size" json:"max_value_size"` // 最大值大小(字节)
	} `mapstructure:"cache" yaml:"cache" json:"cache"`

	// Session存储配置
	Session struct {
		KeyPrefix    string `mapstructure:"key_prefix" yaml:"key_prefix" json:"key_prefix"`          // Session键前缀
		DefaultTTL   string `mapstructure:"default_ttl" yaml:"default_ttl" json:"default_ttl"`       // 默认过期时间
		Serializer   string `mapstructure:"serializer" yaml:"serializer" json:"serializer"`          // 序列化方式
		GCInterval   string `mapstructure:"gc_interval" yaml:"gc_interval" json:"gc_interval"`       // 垃圾回收间隔
		MaxSessions  int    `mapstructure:"max_sessions" yaml:"max_sessions" json:"max_sessions"`    // 最大Session数量
		CleanupBatch int    `mapstructure:"cleanup_batch" yaml:"cleanup_batch" json:"cleanup_batch"` // 清理批次大小
	} `mapstructure:"session" yaml:"session" json:"session"`

	// 限流配置
	RateLimit struct {
		Enable        bool   `mapstructure:"enable" yaml:"enable" json:"enable"`                         // 启用限流
		KeyPrefix     string `mapstructure:"key_prefix" yaml:"key_prefix" json:"key_prefix"`             // 限流键前缀
		DefaultWindow string `mapstructure:"default_window" yaml:"default_window" json:"default_window"` // 默认时间窗口
		DefaultLimit  int    `mapstructure:"default_limit" yaml:"default_limit" json:"default_limit"`    // 默认限制数
		CleanupPeriod string `mapstructure:"cleanup_period" yaml:"cleanup_period" json:"cleanup_period"` // 清理周期
	} `mapstructure:"rate_limit" yaml:"rate_limit" json:"rate_limit"`

	// 开发配置
	Development struct {
		Enable      bool `mapstructure:"enable" yaml:"enable" json:"enable"`                   // 启用开发模式
		LogCommands bool `mapstructure:"log_commands" yaml:"log_commands" json:"log_commands"` // 记录Redis命令
		LogResults  bool `mapstructure:"log_results" yaml:"log_results" json:"log_results"`    // 记录查询结果
		DebugMode   bool `mapstructure:"debug_mode" yaml:"debug_mode" json:"debug_mode"`       // 调试模式
	} `mapstructure:"development" yaml:"development" json:"development"`
}

// GetConfigName 实现 ConfigInterface 接口
func (c RedisConfig) GetConfigName() string {
	return RedisConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c RedisConfig) SetDefaults(v *viper.Viper) {
	// 基础连接默认配置
	v.SetDefault("primary.host", "localhost")
	v.SetDefault("primary.port", 6379)
	v.SetDefault("primary.password", "")
	v.SetDefault("primary.database", 0)
	v.SetDefault("primary.username", "")

	// 连接池默认配置
	v.SetDefault("pool.max_idle", 10)
	v.SetDefault("pool.max_active", 100)
	v.SetDefault("pool.idle_timeout", "300s")
	v.SetDefault("pool.max_conn_timeout", "1s")
	v.SetDefault("pool.test_on_borrow", true)
	v.SetDefault("pool.wait", true)

	// 集群默认配置
	v.SetDefault("cluster.enable", false)
	v.SetDefault("cluster.nodes", []string{})
	v.SetDefault("cluster.password", "")
	v.SetDefault("cluster.read_only", false)
	v.SetDefault("cluster.route_by_latency", false)
	v.SetDefault("cluster.route_randomly", false)

	// 哨兵默认配置
	v.SetDefault("sentinel.enable", false)
	v.SetDefault("sentinel.addresses", []string{})
	v.SetDefault("sentinel.master_name", "mymaster")
	v.SetDefault("sentinel.password", "")
	v.SetDefault("sentinel.database", 0)
	v.SetDefault("sentinel.dial_timeout", "5s")
	v.SetDefault("sentinel.read_timeout", "3s")
	v.SetDefault("sentinel.write_timeout", "3s")

	// 性能默认配置
	v.SetDefault("performance.dial_timeout", "5s")
	v.SetDefault("performance.read_timeout", "3s")
	v.SetDefault("performance.write_timeout", "3s")
	v.SetDefault("performance.pool_timeout", "4s")
	v.SetDefault("performance.idle_check_frequency", "60s")
	v.SetDefault("performance.max_retries", 3)
	v.SetDefault("performance.min_retry_backoff", "8ms")
	v.SetDefault("performance.max_retry_backoff", "512ms")

	// TLS默认配置
	v.SetDefault("tls.enable", false)
	v.SetDefault("tls.cert_file", "")
	v.SetDefault("tls.key_file", "")
	v.SetDefault("tls.ca_file", "")
	v.SetDefault("tls.insecure_skip_verify", false)
	v.SetDefault("tls.server_name", "")

	// 监控默认配置
	v.SetDefault("monitoring.enable", true)
	v.SetDefault("monitoring.metrics_path", "/metrics")
	v.SetDefault("monitoring.slow_log_enable", true)
	v.SetDefault("monitoring.slow_threshold", "100ms")
	v.SetDefault("monitoring.stats_period", "10s")

	// 缓存默认配置
	v.SetDefault("cache.default_ttl", "1h")
	v.SetDefault("cache.key_prefix", "yyhertz:cache:")
	v.SetDefault("cache.serializer", "json")
	v.SetDefault("cache.compression", false)
	v.SetDefault("cache.max_key_length", 250)
	v.SetDefault("cache.max_value_size", 1048576) // 1MB

	// Session存储默认配置
	v.SetDefault("session.key_prefix", "yyhertz:session:")
	v.SetDefault("session.default_ttl", "24h")
	v.SetDefault("session.serializer", "json")
	v.SetDefault("session.gc_interval", "10m")
	v.SetDefault("session.max_sessions", 10000)
	v.SetDefault("session.cleanup_batch", 100)

	// 限流默认配置
	v.SetDefault("rate_limit.enable", false)
	v.SetDefault("rate_limit.key_prefix", "yyhertz:rate_limit:")
	v.SetDefault("rate_limit.default_window", "60s")
	v.SetDefault("rate_limit.default_limit", 100)
	v.SetDefault("rate_limit.cleanup_period", "5m")

	// 开发默认配置
	v.SetDefault("development.enable", false)
	v.SetDefault("development.log_commands", false)
	v.SetDefault("development.log_results", false)
	v.SetDefault("development.debug_mode", false)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c RedisConfig) GenerateDefaultContent() string {
	return `# YYHertz Redis Configuration
# Redis配置文件

# 基础连接配置
primary:
  host: "localhost"                          # Redis主机地址
  port: 6379                                 # Redis端口
  password: ""                               # Redis密码
  database: 0                                # Redis数据库索引(0-15)
  username: ""                               # Redis用户名(Redis 6.0+支持)

# 连接池配置
pool:
  max_idle: 10                               # 最大空闲连接数
  max_active: 100                            # 最大活跃连接数  
  idle_timeout: "300s"                       # 空闲连接超时时间
  max_conn_timeout: "1s"                     # 最大连接超时时间
  test_on_borrow: true                       # 获取连接时测试连接
  wait: true                                 # 连接池满时是否等待

# Redis集群配置
cluster:
  enable: false                              # 启用Redis集群模式
  nodes: []                                  # 集群节点列表
    # - "localhost:7001"
    # - "localhost:7002"
    # - "localhost:7003"
  password: ""                               # 集群密码
  read_only: false                           # 只读模式
  route_by_latency: false                    # 按延迟路由
  route_randomly: false                      # 随机路由

# Redis哨兵配置
sentinel:
  enable: false                              # 启用Redis哨兵模式
  addresses: []                              # 哨兵地址列表
    # - "localhost:26379"
    # - "localhost:26380"
    # - "localhost:26381"
  master_name: "mymaster"                    # 主节点名称
  password: ""                               # 哨兵密码
  database: 0                                # 数据库索引
  dial_timeout: "5s"                         # 连接超时
  read_timeout: "3s"                         # 读取超时
  write_timeout: "3s"                        # 写入超时

# 性能配置
performance:
  dial_timeout: "5s"                         # 连接超时
  read_timeout: "3s"                         # 读取超时
  write_timeout: "3s"                        # 写入超时
  pool_timeout: "4s"                         # 连接池获取超时
  idle_check_frequency: "60s"                # 空闲连接检查频率
  max_retries: 3                             # 最大重试次数
  min_retry_backoff: "8ms"                   # 最小重试间隔
  max_retry_backoff: "512ms"                 # 最大重试间隔

# TLS配置
tls:
  enable: false                              # 启用TLS加密
  cert_file: ""                              # 客户端证书文件路径
  key_file: ""                               # 客户端私钥文件路径
  ca_file: ""                                # CA证书文件路径
  insecure_skip_verify: false                # 跳过证书验证(不建议生产环境使用)
  server_name: ""                            # 服务器名称

# 监控配置
monitoring:
  enable: true                               # 启用监控
  metrics_path: "/metrics"                   # 监控指标路径
  slow_log_enable: true                      # 启用慢日志
  slow_threshold: "100ms"                    # 慢查询阈值
  stats_period: "10s"                        # 统计周期

# 缓存配置
cache:
  default_ttl: "1h"                          # 默认过期时间
  key_prefix: "yyhertz:cache:"               # 缓存键前缀
  serializer: "json"                         # 序列化器: json, msgpack, gob
  compression: false                         # 启用压缩(适用于大值)
  max_key_length: 250                        # 最大键长度
  max_value_size: 1048576                    # 最大值大小(字节) - 1MB

# Session存储配置  
session:
  key_prefix: "yyhertz:session:"             # Session键前缀
  default_ttl: "24h"                         # 默认过期时间
  serializer: "json"                         # 序列化方式: json, gob, msgpack
  gc_interval: "10m"                         # 垃圾回收间隔
  max_sessions: 10000                        # 最大Session数量
  cleanup_batch: 100                         # 批量清理大小

# 限流配置
rate_limit:
  enable: false                              # 启用限流功能
  key_prefix: "yyhertz:rate_limit:"          # 限流键前缀
  default_window: "60s"                      # 默认时间窗口
  default_limit: 100                         # 默认限制数量
  cleanup_period: "5m"                       # 过期数据清理周期

# 开发配置
development:
  enable: false                              # 启用开发模式
  log_commands: false                        # 记录Redis命令到日志
  log_results: false                         # 记录查询结果到日志
  debug_mode: false                          # 调试模式

# 使用示例:
# 1. 基础使用:
#    primary.host = "127.0.0.1"
#    primary.port = 6379
#    primary.password = "your_redis_password"
#
# 2. 集群模式:
#    cluster.enable = true
#    cluster.nodes = ["127.0.0.1:7001", "127.0.0.1:7002", "127.0.0.1:7003"]
#
# 3. 哨兵模式:
#    sentinel.enable = true
#    sentinel.addresses = ["127.0.0.1:26379", "127.0.0.1:26380"]
#    sentinel.master_name = "mymaster"
#
# 4. TLS加密:
#    tls.enable = true
#    tls.cert_file = "/path/to/client.crt"
#    tls.key_file = "/path/to/client.key"
#    tls.ca_file = "/path/to/ca.crt"
#
# 5. 性能优化:
#    pool.max_active = 200
#    performance.read_timeout = "1s"
#    performance.write_timeout = "1s"
`
}
