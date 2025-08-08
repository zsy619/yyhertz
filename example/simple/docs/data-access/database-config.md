# ⚙️ 数据库配置

数据库配置是应用程序的核心组成部分。YYHertz MVC框架提供了灵活的数据库配置管理，支持多种数据库类型和连接模式。

## 🌟 核心特性

### ✨ 配置功能
- **🗄️ 多数据库支持** - MySQL、PostgreSQL、SQLite、SQL Server
- **🔄 连接池管理** - 智能连接池配置和监控
- **🎯 多环境配置** - 开发、测试、生产环境隔离
- **📊 读写分离** - 主从数据库配置支持
- **🔒 安全连接** - SSL/TLS加密连接支持

### 🎪 高级功能
- **📈 性能监控** - 连接状态和性能指标
- **🔄 故障切换** - 自动故障检测和切换
- **💾 连接缓存** - 连接复用和缓存机制
- **🔍 查询日志** - SQL查询日志记录

## 🚀 基础配置

### 1. 配置文件结构

```yaml
# config/database.yaml
database:
  # 默认数据库配置
  default:
    driver: "mysql"
    host: "localhost"
    port: 3306
    database: "myapp"
    username: "root"
    password: "password"
    charset: "utf8mb4"
    timezone: "Asia/Shanghai"
    
    # 连接池配置
    pool:
      max_idle_conns: 10      # 最大空闲连接数
      max_open_conns: 100     # 最大打开连接数
      conn_max_lifetime: "1h" # 连接最大生存时间
      conn_max_idle_time: "30m" # 连接最大空闲时间
    
    # GORM配置
    gorm:
      log_level: "warn"       # 日志级别：silent, error, warn, info
      slow_threshold: "200ms" # 慢查询阈值
      colorful: true          # 彩色日志
      ignore_record_not_found_error: true
      
    # SSL配置
    ssl:
      enabled: false
      cert_file: ""
      key_file: ""
      ca_file: ""
      skip_verify: false
  
  # 读写分离配置
  read_write_split:
    enabled: false
    write: "default"
    read: ["readonly1", "readonly2"]
    
  # 分库分表配置
  sharding:
    enabled: false
    rules: []
```

### 2. 环境特定配置

```yaml
# config/database.development.yaml
database:
  default:
    driver: "sqlite"
    dsn: "file:./data/development.db?cache=shared&mode=rwc"
    gorm:
      log_level: "info"
      colorful: true

# config/database.testing.yaml  
database:
  default:
    driver: "sqlite"
    dsn: ":memory:"
    gorm:
      log_level: "silent"

# config/database.production.yaml
database:
  default:
    driver: "mysql"
    host: "${DB_HOST}"
    port: ${DB_PORT:3306}
    database: "${DB_NAME}"
    username: "${DB_USER}"
    password: "${DB_PASSWORD}"
    pool:
      max_idle_conns: 20
      max_open_conns: 200
      conn_max_lifetime: "2h"
    ssl:
      enabled: true
      skip_verify: false
    gorm:
      log_level: "error"
      colorful: false
```

## 🔧 Go配置结构

### 1. 配置结构定义

```go
// config/database.go
package config

import (
    "time"
    "fmt"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
    Default         *ConnectionConfig            `yaml:"default"`
    Connections     map[string]*ConnectionConfig `yaml:"connections"`
    ReadWriteSplit  *ReadWriteSplitConfig        `yaml:"read_write_split"`
    Sharding        *ShardingConfig              `yaml:"sharding"`
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
    Driver   string      `yaml:"driver"`
    Host     string      `yaml:"host"`
    Port     int         `yaml:"port"`
    Database string      `yaml:"database"`
    Username string      `yaml:"username"`
    Password string      `yaml:"password"`
    Charset  string      `yaml:"charset"`
    Timezone string      `yaml:"timezone"`
    DSN      string      `yaml:"dsn"` // 完整DSN字符串
    
    Pool *PoolConfig `yaml:"pool"`
    GORM *GORMConfig `yaml:"gorm"`
    SSL  *SSLConfig  `yaml:"ssl"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
    MaxIdleConns    int           `yaml:"max_idle_conns"`
    MaxOpenConns    int           `yaml:"max_open_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
    ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// GORMConfig GORM配置
type GORMConfig struct {
    LogLevel                      string        `yaml:"log_level"`
    SlowThreshold                 time.Duration `yaml:"slow_threshold"`
    Colorful                      bool          `yaml:"colorful"`
    IgnoreRecordNotFoundError     bool          `yaml:"ignore_record_not_found_error"`
    ParameterizedQueries          bool          `yaml:"parameterized_queries"`
    PrepareStmt                   bool          `yaml:"prepare_stmt"`
    DisableForeignKeyConstraintWhenMigrating bool `yaml:"disable_foreign_key_constraint_when_migrating"`
}

// SSLConfig SSL配置
type SSLConfig struct {
    Enabled    bool   `yaml:"enabled"`
    CertFile   string `yaml:"cert_file"`
    KeyFile    string `yaml:"key_file"`
    CAFile     string `yaml:"ca_file"`
    SkipVerify bool   `yaml:"skip_verify"`
}

// ReadWriteSplitConfig 读写分离配置
type ReadWriteSplitConfig struct {
    Enabled bool     `yaml:"enabled"`
    Write   string   `yaml:"write"`
    Read    []string `yaml:"read"`
}

// ShardingConfig 分库分表配置
type ShardingConfig struct {
    Enabled bool           `yaml:"enabled"`
    Rules   []ShardingRule `yaml:"rules"`
}

type ShardingRule struct {
    Table    string `yaml:"table"`
    Database string `yaml:"database"`
    Strategy string `yaml:"strategy"`
}

// BuildDSN 构建DSN字符串
func (c *ConnectionConfig) BuildDSN() string {
    if c.DSN != "" {
        return c.DSN
    }
    
    switch c.Driver {
    case "mysql":
        return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
            c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset, c.Timezone)
    case "postgres":
        return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
            c.Host, c.Port, c.Username, c.Password, c.Database, c.Timezone)
    case "sqlite":
        return c.Database
    default:
        return ""
    }
}
```

### 2. 配置加载器

```go
// database/config_loader.go
package database

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "gopkg.in/yaml.v3"
    "github.com/zsy619/yyhertz/config"
)

type ConfigLoader struct {
    environment string
    configPath  string
}

func NewConfigLoader(environment string) *ConfigLoader {
    return &ConfigLoader{
        environment: environment,
        configPath:  "config",
    }
}

// LoadConfig 加载数据库配置
func (cl *ConfigLoader) LoadConfig() (*config.DatabaseConfig, error) {
    // 尝试加载环境特定配置
    envConfigFile := filepath.Join(cl.configPath, fmt.Sprintf("database.%s.yaml", cl.environment))
    if _, err := os.Stat(envConfigFile); err == nil {
        return cl.loadFromFile(envConfigFile)
    }
    
    // 回退到默认配置
    defaultConfigFile := filepath.Join(cl.configPath, "database.yaml")
    return cl.loadFromFile(defaultConfigFile)
}

func (cl *ConfigLoader) loadFromFile(filename string) (*config.DatabaseConfig, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
    }
    
    // 环境变量替换
    content := cl.expandEnvVars(string(data))
    
    var config config.DatabaseConfig
    if err := yaml.Unmarshal([]byte(content), &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
    }
    
    // 验证配置
    if err := cl.validateConfig(&config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    return &config, nil
}

// expandEnvVars 扩展环境变量
func (cl *ConfigLoader) expandEnvVars(content string) string {
    // 支持 ${VAR} 和 ${VAR:default} 格式
    for {
        start := strings.Index(content, "${")
        if start == -1 {
            break
        }
        
        end := strings.Index(content[start:], "}")
        if end == -1 {
            break
        }
        end += start
        
        // 提取变量名和默认值
        varExpr := content[start+2 : end]
        varName := varExpr
        defaultValue := ""
        
        if colonIndex := strings.Index(varExpr, ":"); colonIndex != -1 {
            varName = varExpr[:colonIndex]
            defaultValue = varExpr[colonIndex+1:]
        }
        
        // 获取环境变量值
        value := os.Getenv(varName)
        if value == "" {
            value = defaultValue
        }
        
        // 替换
        content = content[:start] + value + content[end+1:]
    }
    
    return content
}

// validateConfig 验证配置
func (cl *ConfigLoader) validateConfig(config *config.DatabaseConfig) error {
    if config.Default == nil {
        return fmt.Errorf("default database configuration is required")
    }
    
    if config.Default.Driver == "" {
        return fmt.Errorf("database driver is required")
    }
    
    supportedDrivers := []string{"mysql", "postgres", "sqlite", "sqlserver"}
    if !contains(supportedDrivers, config.Default.Driver) {
        return fmt.Errorf("unsupported database driver: %s", config.Default.Driver)
    }
    
    return nil
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

## 🔌 数据库连接管理

### 1. 连接管理器

```go
// database/manager.go
package database

import (
    "fmt"
    "log"
    "sync"
    "time"
    
    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/driver/sqlserver"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "github.com/zsy619/yyhertz/config"
)

type Manager struct {
    config      *config.DatabaseConfig
    connections map[string]*gorm.DB
    mutex       sync.RWMutex
}

var (
    defaultManager *Manager
    once           sync.Once
)

// GetManager 获取数据库管理器实例
func GetManager() *Manager {
    once.Do(func() {
        defaultManager = &Manager{
            connections: make(map[string]*gorm.DB),
        }
    })
    return defaultManager
}

// Initialize 初始化数据库管理器
func (m *Manager) Initialize(config *config.DatabaseConfig) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    m.config = config
    
    // 初始化默认连接
    if err := m.initConnection("default", config.Default); err != nil {
        return fmt.Errorf("failed to initialize default connection: %w", err)
    }
    
    // 初始化其他连接
    for name, connConfig := range config.Connections {
        if err := m.initConnection(name, connConfig); err != nil {
            log.Printf("Warning: failed to initialize connection %s: %v", name, err)
        }
    }
    
    log.Println("Database connections initialized successfully")
    return nil
}

// initConnection 初始化单个连接
func (m *Manager) initConnection(name string, config *config.ConnectionConfig) error {
    // 创建GORM配置
    gormConfig := m.buildGORMConfig(config.GORM)
    
    // 创建数据库驱动
    dialector, err := m.createDialector(config)
    if err != nil {
        return err
    }
    
    // 创建GORM实例
    db, err := gorm.Open(dialector, gormConfig)
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // 配置连接池
    if err := m.configurePool(db, config.Pool); err != nil {
        return fmt.Errorf("failed to configure connection pool: %w", err)
    }
    
    m.connections[name] = db
    return nil
}

// createDialector 创建数据库驱动
func (m *Manager) createDialector(config *config.ConnectionConfig) (gorm.Dialector, error) {
    dsn := config.BuildDSN()
    
    switch config.Driver {
    case "mysql":
        return mysql.Open(dsn), nil
    case "postgres":
        return postgres.Open(dsn), nil
    case "sqlite":
        return sqlite.Open(dsn), nil
    case "sqlserver":
        return sqlserver.Open(dsn), nil
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
    }
}

// buildGORMConfig 构建GORM配置
func (m *Manager) buildGORMConfig(config *config.GORMConfig) *gorm.Config {
    if config == nil {
        return &gorm.Config{}
    }
    
    // 配置日志级别
    var logLevel logger.LogLevel
    switch config.LogLevel {
    case "silent":
        logLevel = logger.Silent
    case "error":
        logLevel = logger.Error
    case "warn":
        logLevel = logger.Warn
    case "info":
        logLevel = logger.Info
    default:
        logLevel = logger.Warn
    }
    
    return &gorm.Config{
        Logger: logger.Default.LogMode(logLevel),
        NamingStrategy: schema.NamingStrategy{
            SingularTable: true,
        },
        DisableForeignKeyConstraintWhenMigrating: config.DisableForeignKeyConstraintWhenMigrating,
    }
}

// configurePool 配置连接池
func (m *Manager) configurePool(db *gorm.DB, config *config.PoolConfig) error {
    if config == nil {
        return nil
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    
    if config.MaxIdleConns > 0 {
        sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    }
    
    if config.MaxOpenConns > 0 {
        sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    }
    
    if config.ConnMaxLifetime > 0 {
        sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
    }
    
    if config.ConnMaxIdleTime > 0 {
        sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    }
    
    return nil
}

// GetConnection 获取数据库连接
func (m *Manager) GetConnection(name ...string) *gorm.DB {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    connectionName := "default"
    if len(name) > 0 && name[0] != "" {
        connectionName = name[0]
    }
    
    return m.connections[connectionName]
}

// GetWriteConnection 获取写连接
func (m *Manager) GetWriteConnection() *gorm.DB {
    if m.config.ReadWriteSplit != nil && m.config.ReadWriteSplit.Enabled {
        return m.GetConnection(m.config.ReadWriteSplit.Write)
    }
    return m.GetConnection()
}

// GetReadConnection 获取读连接
func (m *Manager) GetReadConnection() *gorm.DB {
    if m.config.ReadWriteSplit != nil && m.config.ReadWriteSplit.Enabled && len(m.config.ReadWriteSplit.Read) > 0 {
        // 简单轮询选择读连接
        readConnections := m.config.ReadWriteSplit.Read
        connectionName := readConnections[time.Now().Unix()%int64(len(readConnections))]
        return m.GetConnection(connectionName)
    }
    return m.GetConnection()
}
```

### 2. 便捷访问函数

```go
// database/db.go
package database

import "gorm.io/gorm"

// DB 获取默认数据库连接
func DB() *gorm.DB {
    return GetManager().GetConnection()
}

// WriteDB 获取写数据库连接
func WriteDB() *gorm.DB {
    return GetManager().GetWriteConnection()
}

// ReadDB 获取读数据库连接
func ReadDB() *gorm.DB {
    return GetManager().GetReadConnection()
}

// Connection 获取指定名称的数据库连接
func Connection(name string) *gorm.DB {
    return GetManager().GetConnection(name)
}
```

## 📊 健康检查和监控

### 1. 健康检查

```go
// database/health.go
package database

import (
    "context"
    "fmt"
    "time"
)

type HealthStatus struct {
    Connection string        `json:"connection"`
    Status     string        `json:"status"`
    Latency    time.Duration `json:"latency"`
    Error      string        `json:"error,omitempty"`
}

// CheckHealth 检查数据库健康状态
func CheckHealth() []HealthStatus {
    manager := GetManager()
    var results []HealthStatus
    
    manager.mutex.RLock()
    connections := make(map[string]*gorm.DB, len(manager.connections))
    for name, db := range manager.connections {
        connections[name] = db
    }
    manager.mutex.RUnlock()
    
    for name, db := range connections {
        status := checkSingleConnection(name, db)
        results = append(results, status)
    }
    
    return results
}

func checkSingleConnection(name string, db *gorm.DB) HealthStatus {
    start := time.Now()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    sqlDB, err := db.DB()
    if err != nil {
        return HealthStatus{
            Connection: name,
            Status:     "error",
            Latency:    time.Since(start),
            Error:      err.Error(),
        }
    }
    
    if err := sqlDB.PingContext(ctx); err != nil {
        return HealthStatus{
            Connection: name,
            Status:     "error",
            Latency:    time.Since(start),
            Error:      err.Error(),
        }
    }
    
    return HealthStatus{
        Connection: name,
        Status:     "healthy",
        Latency:    time.Since(start),
    }
}
```

### 2. 性能监控

```go
// database/metrics.go
package database

import (
    "database/sql"
    "time"
)

type ConnectionMetrics struct {
    Name            string        `json:"name"`
    OpenConnections int           `json:"open_connections"`
    InUse          int           `json:"in_use"`
    Idle           int           `json:"idle"`
    WaitCount      int64         `json:"wait_count"`
    WaitDuration   time.Duration `json:"wait_duration"`
    MaxIdleConns   int           `json:"max_idle_conns"`
    MaxOpenConns   int           `json:"max_open_conns"`
}

// GetMetrics 获取连接池指标
func GetMetrics() ([]ConnectionMetrics, error) {
    manager := GetManager()
    var metrics []ConnectionMetrics
    
    manager.mutex.RLock()
    defer manager.mutex.RUnlock()
    
    for name, db := range manager.connections {
        sqlDB, err := db.DB()
        if err != nil {
            continue
        }
        
        stats := sqlDB.Stats()
        metrics = append(metrics, ConnectionMetrics{
            Name:            name,
            OpenConnections: stats.OpenConnections,
            InUse:          stats.InUse,
            Idle:           stats.Idle,
            WaitCount:      stats.WaitCount,
            WaitDuration:   stats.WaitDuration,
            MaxIdleConns:   stats.MaxIdleConns,
            MaxOpenConns:   stats.MaxOpenConns,
        })
    }
    
    return metrics, nil
}

// StartMetricsCollection 启动指标收集
func StartMetricsCollection(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics, err := GetMetrics()
        if err != nil {
            continue
        }
        
        for _, metric := range metrics {
            // 记录指标到日志或监控系统
            logMetrics(metric)
        }
    }
}

func logMetrics(metric ConnectionMetrics) {
    // 可以集成Prometheus、InfluxDB等监控系统
    // prometheus.GaugeVec.WithLabelValues(metric.Name).Set(float64(metric.OpenConnections))
}
```

## 🔧 最佳实践

### 1. 连接池配置建议

```go
// 不同环境的推荐配置
var RecommendedConfigs = map[string]*config.PoolConfig{
    "development": {
        MaxIdleConns:    5,
        MaxOpenConns:    25,
        ConnMaxLifetime: 1 * time.Hour,
        ConnMaxIdleTime: 10 * time.Minute,
    },
    "testing": {
        MaxIdleConns:    2,
        MaxOpenConns:    10,
        ConnMaxLifetime: 30 * time.Minute,
        ConnMaxIdleTime: 5 * time.Minute,
    },
    "production": {
        MaxIdleConns:    25,
        MaxOpenConns:    100,
        ConnMaxLifetime: 2 * time.Hour,
        ConnMaxIdleTime: 30 * time.Minute,
    },
}
```

### 2. 配置安全建议

```yaml
# 生产环境安全配置
database:
  default:
    # 使用环境变量存储敏感信息
    username: "${DB_USER}"
    password: "${DB_PASSWORD}"
    
    # 启用SSL
    ssl:
      enabled: true
      skip_verify: false
      cert_file: "/path/to/client-cert.pem"
      key_file: "/path/to/client-key.pem"
      ca_file: "/path/to/ca-cert.pem"
    
    # 限制权限
    gorm:
      disable_foreign_key_constraint_when_migrating: true
      parameterized_queries: true
      prepare_stmt: true
```

### 3. 故障排除指南

```go
// 数据库连接故障诊断
func DiagnoseConnection(connectionName string) string {
    db := GetManager().GetConnection(connectionName)
    if db == nil {
        return fmt.Sprintf("Connection '%s' not found", connectionName)
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return fmt.Sprintf("Failed to get underlying sql.DB: %v", err)
    }
    
    // 检查连接
    if err := sqlDB.Ping(); err != nil {
        return fmt.Sprintf("Connection ping failed: %v", err)
    }
    
    // 检查统计信息
    stats := sqlDB.Stats()
    if stats.OpenConnections == 0 {
        return "No open connections"
    }
    
    if stats.InUse == stats.MaxOpenConns {
        return "Connection pool exhausted"
    }
    
    return "Connection is healthy"
}
```

## 🧪 测试支持

### 1. 测试数据库配置

```go
// database/testing.go
package database

import (
    "testing"
    "github.com/zsy619/yyhertz/config"
)

// SetupTestDB 设置测试数据库
func SetupTestDB(t *testing.T) {
    config := &config.DatabaseConfig{
        Default: &config.ConnectionConfig{
            Driver: "sqlite",
            DSN:    ":memory:",
            GORM: &config.GORMConfig{
                LogLevel: "silent",
            },
        },
    }
    
    if err := GetManager().Initialize(config); err != nil {
        t.Fatalf("Failed to setup test database: %v", err)
    }
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB() {
    // 清理连接
    manager := GetManager()
    manager.mutex.Lock()
    defer manager.mutex.Unlock()
    
    for _, db := range manager.connections {
        if sqlDB, err := db.DB(); err == nil {
            sqlDB.Close()
        }
    }
    
    manager.connections = make(map[string]*gorm.DB)
}
```

## 🔗 相关资源

- [GORM集成指南](./gorm.md)
- [事务管理](./transaction.md)
- [应用配置管理](../configuration/app-config.md)
- [性能优化建议](../dev-tools/performance.md)

---

> 💡 **提示**: 合理的数据库配置是应用性能的关键因素。建议根据实际负载和环境特点调整连接池参数。
