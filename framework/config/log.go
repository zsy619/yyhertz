package config

import (
	"time"

	"github.com/spf13/viper"
)

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// LogFormat 日志格式
type LogFormat string

const (
	LogFormatJSON            LogFormat = "json"
	LogFormatText            LogFormat = "text"
	LogFormatBeego           LogFormat = "beego"
	LogFormatLog4Go          LogFormat = "log4go"
	LogFormatLogstash        LogFormat = "logstash"
	LogFormatSyslog          LogFormat = "syslog"
	LogFormatFluentd         LogFormat = "fluentd"
	LogFormatCloudWatch      LogFormat = "cloudwatch"
	LogFormatApplicationInsights LogFormat = "azure_insights"
)

// LogOutput 日志输出类型
type LogOutput string

const (
	LogOutputConsole      LogOutput = "console"
	LogOutputFile         LogOutput = "file"
	LogOutputSyslog       LogOutput = "syslog"
	LogOutputFluentd      LogOutput = "fluentd"
	LogOutputCloudWatch   LogOutput = "cloudwatch"
	LogOutputAzureInsights LogOutput = "azure_insights"
	LogOutputElasticsearch LogOutput = "elasticsearch"
	LogOutputKafka        LogOutput = "kafka"
)

// LogConfig 日志配置结构
type LogConfig struct {
	// 基础配置
	Level  LogLevel  `mapstructure:"level" yaml:"level" json:"level"`     // 日志级别
	Format LogFormat `mapstructure:"format" yaml:"format" json:"format"` // 日志格式

	// 输出配置
	EnableConsole bool   `mapstructure:"enable_console" yaml:"enable_console" json:"enable_console"` // 是否输出到控制台
	EnableFile    bool   `mapstructure:"enable_file" yaml:"enable_file" json:"enable_file"`          // 是否输出到文件
	FilePath      string `mapstructure:"file_path" yaml:"file_path" json:"file_path"`                // 日志文件路径
	MaxSize       int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"`                   // 单个日志文件最大大小(MB)
	MaxAge        int    `mapstructure:"max_age" yaml:"max_age" json:"max_age"`                      // 日志文件保留天数
	MaxBackups    int    `mapstructure:"max_backups" yaml:"max_backups" json:"max_backups"`          // 最大备份数量
	Compress      bool   `mapstructure:"compress" yaml:"compress" json:"compress"`                   // 是否压缩旧日志

	// 高级配置
	ShowCaller      bool   `mapstructure:"show_caller" yaml:"show_caller" json:"show_caller"`                // 是否显示调用位置
	ShowTimestamp   bool   `mapstructure:"show_timestamp" yaml:"show_timestamp" json:"show_timestamp"`    // 是否显示时间戳
	TimestampFormat string `mapstructure:"timestamp_format" yaml:"timestamp_format" json:"timestamp_format"` // 时间戳格式

	// 字段配置
	Fields map[string]any `mapstructure:"fields" yaml:"fields" json:"fields"` // 全局字段

	// 扩展输出配置
	Outputs      []string                 `mapstructure:"outputs" yaml:"outputs" json:"outputs"`                   // 启用的输出类型
	OutputConfig map[string]OutputConfig `mapstructure:"output_config" yaml:"output_config" json:"output_config"` // 各输出的配置
}

// OutputConfig 输出配置接口
type OutputConfig interface {
	Validate() error
}

// SyslogConfig Syslog输出配置
type SyslogConfig struct {
	Network  string `mapstructure:"network" yaml:"network" json:"network"`   // 网络类型：tcp, udp, unix
	Address  string `mapstructure:"address" yaml:"address" json:"address"`   // 地址
	Priority int    `mapstructure:"priority" yaml:"priority" json:"priority"` // 优先级
	Tag      string `mapstructure:"tag" yaml:"tag" json:"tag"`               // 标签
}

func (c SyslogConfig) Validate() error {
	return nil // 实现验证逻辑
}

// FluentdConfig Fluentd输出配置
type FluentdConfig struct {
	Host    string            `mapstructure:"host" yaml:"host" json:"host"`       // Fluentd主机
	Port    int               `mapstructure:"port" yaml:"port" json:"port"`       // Fluentd端口
	Tag     string            `mapstructure:"tag" yaml:"tag" json:"tag"`          // 标签
	Timeout time.Duration     `mapstructure:"timeout" yaml:"timeout" json:"timeout"` // 超时时间
	Extra   map[string]string `mapstructure:"extra" yaml:"extra" json:"extra"`    // 额外字段
}

func (c FluentdConfig) Validate() error {
	return nil
}

// CloudWatchConfig AWS CloudWatch输出配置
type CloudWatchConfig struct {
	Region          string `mapstructure:"region" yaml:"region" json:"region"`                         // AWS区域
	LogGroupName    string `mapstructure:"log_group_name" yaml:"log_group_name" json:"log_group_name"` // 日志组名
	LogStreamName   string `mapstructure:"log_stream_name" yaml:"log_stream_name" json:"log_stream_name"` // 日志流名
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id" json:"access_key_id"`    // 访问密钥ID
	SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key" json:"secret_access_key"` // 秘密访问密钥
}

func (c CloudWatchConfig) Validate() error {
	return nil
}

// AzureInsightsConfig Azure Application Insights输出配置
type AzureInsightsConfig struct {
	InstrumentationKey string            `mapstructure:"instrumentation_key" yaml:"instrumentation_key" json:"instrumentation_key"` // 仪器密钥
	Endpoint           string            `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`                                     // 端点
	Properties         map[string]string `mapstructure:"properties" yaml:"properties" json:"properties"`                              // 自定义属性
}

func (c AzureInsightsConfig) Validate() error {
	return nil
}

// ElasticsearchConfig Elasticsearch输出配置
type ElasticsearchConfig struct {
	URLs     []string `mapstructure:"urls" yaml:"urls" json:"urls"`          // Elasticsearch URLs
	Index    string   `mapstructure:"index" yaml:"index" json:"index"`       // 索引名
	Username string   `mapstructure:"username" yaml:"username" json:"username"` // 用户名
	Password string   `mapstructure:"password" yaml:"password" json:"password"` // 密码
}

func (c ElasticsearchConfig) Validate() error {
	return nil
}

// KafkaConfig Kafka输出配置
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers" yaml:"brokers" json:"brokers"` // Kafka brokers
	Topic   string   `mapstructure:"topic" yaml:"topic" json:"topic"`       // 主题
	ClientID string  `mapstructure:"client_id" yaml:"client_id" json:"client_id"` // 客户端ID
}

func (c KafkaConfig) Validate() error {
	return nil
}

// ============= ConfigInterface 实现 =============

// GetConfigName 获取配置名称
func (c LogConfig) GetConfigName() string {
	return LogConfigName
}

// SetDefaults 设置默认配置
func (c LogConfig) SetDefaults(v *viper.Viper) {
	// 基础配置
	v.SetDefault("level", "info")
	v.SetDefault("format", "beego")

	// 输出配置
	v.SetDefault("enable_console", true)
	v.SetDefault("enable_file", true)
	v.SetDefault("file_path", "./logs/app.log")
	v.SetDefault("max_size", 100)
	v.SetDefault("max_age", 7)
	v.SetDefault("max_backups", 10)
	v.SetDefault("compress", true)

	// 高级配置
	v.SetDefault("show_caller", true)
	v.SetDefault("show_timestamp", true)
	v.SetDefault("timestamp_format", "2006/01/02 15:04:05.000")

	// 字段配置
	v.SetDefault("fields", map[string]any{})

	// 扩展输出配置
	v.SetDefault("outputs", []string{"console", "file"})
	v.SetDefault("output_config", map[string]OutputConfig{})
}

// GenerateDefaultContent 生成默认配置文件内容
func (c LogConfig) GenerateDefaultContent() string {
	return `# 日志配置
# 支持日志级别: debug, info, warn, error, fatal, panic
level: "info"

# 日志格式: json, text, beego, log4go, logstash, syslog, fluentd, cloudwatch, azure_insights
# beego: Beego风格格式 [L] yyyy/mm/dd hh:mm:ss.sss [filename:line] message
# log4go: Log4go风格格式 [yyyy/mm/dd hh:mm:ss] [LEVEL] (filename:line) message
# logstash: Logstash JSON格式，包含@timestamp, @version, level, message等字段
# syslog: 标准Syslog格式 <priority>timestamp hostname tag: message
# fluentd: Fluentd JSON格式，适用于日志聚合
# cloudwatch: AWS CloudWatch格式
# azure_insights: Azure Application Insights格式
format: "beego"

# 输出配置
enable_console: true  # 是否输出到控制台
enable_file: true     # 是否输出到文件
file_path: "./logs/app.log"  # 日志文件路径

# 文件轮转配置
max_size: 100      # 单个日志文件最大大小(MB)
max_age: 7         # 日志文件保留天数
max_backups: 10    # 最大备份数量
compress: true     # 是否压缩旧日志

# 高级配置
show_caller: true      # 是否显示调用位置
show_timestamp: true   # 是否显示时间戳
timestamp_format: "2006/01/02 15:04:05.000"  # Beego风格时间格式

# 全局字段（可选）
fields:
  service: "yyhertz"
  version: "1.0.0"
  # 可以添加更多全局字段

# 输出目标配置
outputs:
  - "console"
  - "file"
  # - "syslog"
  # - "fluentd"
  # - "cloudwatch"
  # - "azure_insights"
  # - "elasticsearch"
  # - "kafka"

# 各输出目标的具体配置
output_config:
  syslog:
    network: "udp"
    address: "localhost:514"
    priority: 16  # local0.info
    tag: "yyhertz"
  
  fluentd:
    host: "localhost"
    port: 24224
    tag: "yyhertz.logs"
    timeout: "3s"
    extra:
      environment: "production"
  
  cloudwatch:
    region: "us-east-1"
    log_group_name: "/aws/yyhertz/application"
    log_stream_name: "yyhertz-instance-001"
    # access_key_id: "your-access-key"
    # secret_access_key: "your-secret-key"
  
  azure_insights:
    instrumentation_key: "your-instrumentation-key"
    endpoint: "https://dc.services.visualstudio.com/v2/track"
    properties:
      application: "yyhertz"
      environment: "production"
  
  elasticsearch:
    urls:
      - "http://localhost:9200"
    index: "yyhertz-logs"
    # username: "elastic"
    # password: "password"
  
  kafka:
    brokers:
      - "localhost:9092"
    topic: "yyhertz-logs"
    client_id: "yyhertz-logger"
`
}