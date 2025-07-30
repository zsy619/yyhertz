# YYHertz Viper 配置管理系统使用指南

## 概述

YYHertz 框架集成了强大的 Viper 配置管理系统，支持多种数据源、环境变量覆盖、配置文件监听等企业级功能。本指南详细介绍了如何使用这个配置系统。

## 🚀 主要特性

### ✨ 功能特性

- **多数据源支持**：支持 YAML、JSON、TOML 等格式
- **分层配置**：默认值 → 配置文件 → 环境变量 → 手动设置
- **环境变量映射**：自动支持 `YYHERTZ_` 前缀的环境变量
- **配置文件监听**：支持配置文件热重载
- **路径搜索**：支持多个配置文件搜索路径
- **类型安全**：强类型配置结构定义
- **全局单例**：提供全局配置实例和便捷函数

### 🔧 技术特性

- **双配置管理器**：保持现有简单配置管理器的兼容性
- **命名空间隔离**：避免与现有 ConfigManager 的命名冲突
- **完整测试覆盖**：包含单元测试和集成测试
- **生产就绪**：支持默认配置文件自动创建

## 📦 快速开始

### 基本使用

```go
package main

import (
    "log"
    "github.com/zsy619/yyhertz/framework/config"
)

func main() {
    // 获取全局配置管理器实例
    cm := config.GetViperConfigManager()
    
    // 获取基本配置值
    appName := cm.GetString("app.name")
    port := cm.GetInt("app.port")
    debug := cm.GetBool("app.debug")
    
    fmt.Printf("应用: %s, 端口: %d, 调试: %v\n", appName, port, debug)
}
```

### 获取完整配置结构

```go
// 获取完整的配置结构
appConfig, err := cm.GetConfig()
if err != nil {
    log.Fatal("获取配置失败:", err)
}

fmt.Printf("应用名称: %s\n", appConfig.App.Name)
fmt.Printf("数据库地址: %s:%d\n", appConfig.Database.Host, appConfig.Database.Port)
fmt.Printf("Redis地址: %s:%d\n", appConfig.Redis.Host, appConfig.Redis.Port)
```

### 使用全局便捷函数

```go
// 使用全局便捷函数
config, err := config.GetGlobalConfig()
appName := config.GetConfigString("app.name")
port := config.GetConfigInt("app.port")
debug := config.GetConfigBool("app.debug")
```

## 📝 配置文件格式

### 默认配置文件 (config.yaml)

系统会自动在以下位置搜索配置文件：
- `./config/config.yaml` - 项目配置目录
- `./config.yaml` - 当前目录
- `/etc/yyhertz/config.yaml` - 系统配置目录
- `$HOME/.yyhertz/config.yaml` - 用户配置目录

```yaml
# YYHertz Framework Configuration

# 应用基础配置
app:
  name: "YYHertz"
  version: "1.0.0"
  environment: "development"  # development, testing, production
  debug: true
  port: 8888
  host: "0.0.0.0"
  timezone: "Asia/Shanghai"

# 数据库配置
database:
  driver: "mysql"
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: ""
  database: "yyhertz"
  charset: "utf8mb4"
  max_idle: 10
  max_open: 100
  max_life: 3600  # 秒
  ssl_mode: "disable"

# Redis配置
redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  database: 0
  max_retries: 3
  pool_size: 10
  min_idle: 2
  dial_timeout: 5  # 秒
  read_timeout: 3  # 秒

# 日志配置
log:
  level: "info"          # debug, info, warn, error, fatal, panic
  format: "json"         # json, text
  enable_console: true
  enable_file: false
  file_path: "./logs/app.log"
  max_size: 100          # MB
  max_age: 7            # 天
  max_backups: 10
  compress: true
  show_caller: true
  show_timestamp: true

# TLS配置
tls:
  enable: false
  cert_file: ""
  key_file: ""
  min_version: "1.2"
  max_version: "1.3"
  auto_reload: false
  reload_interval: 300

# 中间件配置
middleware:
  # CORS跨域配置
  cors:
    enable: true
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["*"]
    expose_headers: []
    allow_credentials: false
    max_age: 3600

  # 限流配置
  rate_limit:
    enable: false
    rate: 100              # 请求/秒
    burst: 200             # 突发容量
    strategy: "token_bucket"  # token_bucket, sliding_window

  # 认证配置
  auth:
    enable: false
    jwt_secret: "your-secret-key-change-me"
    token_ttl: 24          # 小时
    refresh_ttl: 168       # 小时

# 外部服务配置
services:
  # 邮件服务
  email:
    provider: "smtp"       # smtp, sendgrid, ses
    host: "smtp.gmail.com"
    port: 587
    username: ""
    password: ""
    from: "noreply@example.com"

  # 文件存储
  storage:
    provider: "local"      # local, s3, oss
    local_path: "./uploads"
    bucket: ""
    region: ""
    access_key: ""
    secret_key: ""
    cdn_domain: ""

# 监控配置
monitor:
  enable: false
  endpoint: "/metrics"
  interval: 30          # 秒
  timeout: 10           # 秒
```

## 🌍 环境变量支持

### 环境变量命名规则

环境变量使用 `YYHERTZ_` 前缀，支持嵌套配置：

- `YYHERTZ_APP_NAME` → `app.name`
- `YYHERTZ_APP_PORT` → `app.port`
- `YYHERTZ_DATABASE_HOST` → `database.host`
- `YYHERTZ_DATABASE_PASSWORD` → `database.password`
- `YYHERTZ_REDIS_HOST` → `redis.host`
- `YYHERTZ_LOG_LEVEL` → `log.level`

### 使用示例

```bash
# 设置环境变量
export YYHERTZ_APP_NAME="MyApp"
export YYHERTZ_APP_PORT="9000"
export YYHERTZ_APP_DEBUG="false"
export YYHERTZ_DATABASE_HOST="prod-db.example.com"
export YYHERTZ_DATABASE_PASSWORD="secret"
export YYHERTZ_REDIS_HOST="redis.example.com"
export YYHERTZ_LOG_LEVEL="error"

# 启动应用（环境变量会自动覆盖配置文件中的值）
./yyhertz
```

## 🔧 高级用法

### 自定义配置管理器

```go
// 创建独立的配置管理器实例
cm := config.NewViperConfigManager()

// 设置配置文件名和类型
cm.SetConfigName("myconfig")
cm.SetConfigType("json")

// 添加搜索路径
cm.AddConfigPath("/etc/myapp/")
cm.AddConfigPath("$HOME/.myapp")
cm.AddConfigPath(".")

// 设置环境变量前缀
cm.SetEnvPrefix("MYAPP")

// 初始化配置
err := cm.Initialize()
if err != nil {
    log.Fatal("配置初始化失败:", err)
}
```

### 配置文件监听

```go
// 启用配置文件监听
cm.WatchConfig()

// 配置文件变化时会自动重载，并记录日志
```

### 动态配置设置

```go
// 动态设置配置值
cm.Set("custom.api_key", "your-api-key-here")
cm.Set("custom.timeout", 30)
cm.Set("custom.enabled", true)
cm.Set("custom.tags", []string{"web", "framework", "go"})

// 读取配置
apiKey := cm.GetString("custom.api_key")
timeout := cm.GetInt("custom.timeout")
enabled := cm.GetBool("custom.enabled")
tags := cm.GetStringSlice("custom.tags")
```

### 配置存在性检查

```go
// 检查配置是否存在
if cm.IsSet("database.password") {
    password := cm.GetString("database.password")
    // 使用密码连接数据库
}

// 获取所有配置键
allKeys := cm.AllKeys()
for _, key := range allKeys {
    value := cm.Get(key)
    fmt.Printf("%s = %v\n", key, value)
}
```

### 写入配置文件

```go
// 写入当前配置到文件
err := cm.WriteConfig()
if err != nil {
    log.Printf("写入配置失败: %v", err)
}

// 写入配置到指定文件
err = cm.WriteConfigAs("/path/to/new/config.yaml")
```

## 🛠️ 集成示例

### 在 main.go 中使用

```go
func main() {
    // 解析命令行参数
    var configFile = flag.String("config", "", "配置文件路径")
    flag.Parse()
    
    // 初始化配置管理器
    configManager := config.GetViperConfigManager()
    if *configFile != "" {
        configManager.SetConfigFile(*configFile)
    }
    
    if err := configManager.Initialize(); err != nil {
        log.Fatal("配置初始化失败:", err)
    }
    
    // 启用配置文件监听
    configManager.WatchConfig()
    
    // 获取应用配置
    appConfig, err := configManager.GetConfig()
    if err != nil {
        log.Fatal("获取配置失败:", err)
    }
    
    // 使用配置启动服务器
    app := controller.NewApp()
    
    // 从配置创建TLS中间件
    if appConfig.TLS.Enable {
        tlsConfig := middleware.DefaultTLSConfig()
        tlsConfig.Enable = appConfig.TLS.Enable
        tlsConfig.CertFile = appConfig.TLS.CertFile
        tlsConfig.KeyFile = appConfig.TLS.KeyFile
        app.Use(middleware.TLSSupportMiddleware(tlsConfig))
    }
    
    // 启动服务器
    app.Spin()
}
```

### 数据库连接示例

```go
func connectDatabase() (*sql.DB, error) {
    config, err := config.GetGlobalConfig()
    if err != nil {
        return nil, err
    }
    
    var dsn string
    switch config.Database.Driver {
    case "mysql":
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
            config.Database.Username,
            config.Database.Password,
            config.Database.Host,
            config.Database.Port,
            config.Database.Database,
            config.Database.Charset,
        )
    case "postgres":
        dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
            config.Database.Host,
            config.Database.Port,
            config.Database.Username,
            config.Database.Password,
            config.Database.Database,
            config.Database.SSLMode,
        )
    default:
        return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Database.Driver)
    }
    
    db, err := sql.Open(config.Database.Driver, dsn)
    if err != nil {
        return nil, err
    }
    
    // 设置连接池参数
    db.SetMaxIdleConns(config.Database.MaxIdle)
    db.SetMaxOpenConns(config.Database.MaxOpen)
    db.SetConnMaxLifetime(time.Duration(config.Database.MaxLife) * time.Second)
    
    return db, nil
}
```

## 🧪 测试和演示

### 运行配置演示程序

```bash
# 构建演示程序
cd cmd/config_demo
go build -o config_demo main.go

# 运行基本演示
./config_demo

# 运行特定示例
./config_demo example     # 基本配置使用示例
./config_demo database    # 数据库配置示例
./config_demo redis       # Redis配置示例
./config_demo log         # 日志配置示例
./config_demo tls         # TLS配置示例
./config_demo middleware  # 中间件配置示例
./config_demo env         # 环境变量配置示例
./config_demo all         # 运行所有示例

# 查看帮助
./config_demo help
```

### 运行单元测试

```bash
# 运行所有配置相关测试
go test ./framework/config/ -v

# 运行特定测试
go test ./framework/config/ -run TestViperConfigManager -v
```

## 📚 API 参考

### ViperConfigManager 方法

#### 创建和初始化
- `NewViperConfigManager() *ViperConfigManager` - 创建新的配置管理器
- `GetViperConfigManager() *ViperConfigManager` - 获取全局实例
- `Initialize() error` - 初始化配置管理器

#### 配置设置
- `SetConfigFile(file string)` - 设置配置文件路径
- `SetConfigName(name string)` - 设置配置文件名
- `SetConfigType(configType string)` - 设置配置文件类型
- `AddConfigPath(path string)` - 添加搜索路径
- `SetEnvPrefix(prefix string)` - 设置环境变量前缀

#### 配置读取
- `GetConfig() (*AppConfig, error)` - 获取完整配置结构
- `Get(key string) interface{}` - 获取任意类型配置值
- `GetString(key string) string` - 获取字符串配置值
- `GetInt(key string) int` - 获取整数配置值
- `GetBool(key string) bool` - 获取布尔配置值
- `GetStringSlice(key string) []string` - 获取字符串数组配置值
- `GetDuration(key string) time.Duration` - 获取时间间隔配置值

#### 配置操作
- `Set(key string, value interface{})` - 设置配置值
- `IsSet(key string) bool` - 检查配置是否存在
- `AllKeys() []string` - 获取所有配置键

#### 高级功能
- `WatchConfig()` - 监听配置文件变化
- `WriteConfig() error` - 写入配置文件
- `WriteConfigAs(filename string) error` - 写入到指定文件
- `ConfigFileUsed() string` - 获取当前使用的配置文件路径

### 全局便捷函数

- `GetGlobalConfig() (*AppConfig, error)` - 获取全局配置
- `GetConfigValue(key string) interface{}` - 获取配置值
- `GetConfigString(key string) string` - 获取字符串配置值
- `GetConfigInt(key string) int` - 获取整数配置值
- `GetConfigBool(key string) bool` - 获取布尔配置值

## 🔄 兼容性说明

### 与现有系统的兼容性

- **现有 ConfigManager**：保持现有简单配置管理器的完整功能
- **命名空间隔离**：新的 Viper 配置管理器使用 `ViperConfigManager` 命名
- **双配置系统**：可以同时使用两套配置系统，互不干扰
- **逐步迁移**：可以逐步将现有代码迁移到新的配置系统

### 迁移指南

1. **保持现有代码不变**：现有使用 `config.ConfigManager` 的代码无需修改
2. **新功能使用新系统**：新开发的功能建议使用 `config.ViperConfigManager`
3. **逐步替换**：可以逐步将现有代码从简单配置管理器迁移到 Viper 配置管理器

## 📖 最佳实践

### 1. 配置文件组织

```yaml
# 推荐的配置文件结构
# 按功能模块组织配置项
app:          # 应用基础配置
database:     # 数据库相关配置
redis:        # Redis 相关配置
log:          # 日志相关配置
middleware:   # 中间件相关配置
services:     # 外部服务配置
```

### 2. 环境特定配置

```bash
# 不同环境使用不同的配置文件
config/
  ├── config.yaml          # 基础配置
  ├── config.dev.yaml      # 开发环境
  ├── config.test.yaml     # 测试环境
  └── config.prod.yaml     # 生产环境
```

### 3. 敏感信息处理

```bash
# 敏感信息通过环境变量设置，不要写在配置文件中
export YYHERTZ_DATABASE_PASSWORD="secret"
export YYHERTZ_JWT_SECRET="your-jwt-secret"
export YYHERTZ_REDIS_PASSWORD="redis-password"
```

### 4. 配置验证

```go
// 在应用启动时验证关键配置
config, err := config.GetGlobalConfig()
if err != nil {
    log.Fatal("获取配置失败:", err)
}

// 验证必要的配置项
if config.App.Name == "" {
    log.Fatal("应用名称不能为空")
}

if config.Database.Host == "" {
    log.Fatal("数据库主机地址不能为空")
}
```

## 🎯 总结

YYHertz Viper 配置管理系统提供了企业级的配置管理能力，支持：

✅ **多数据源**：配置文件、环境变量、默认值、手动设置  
✅ **类型安全**：强类型配置结构和类型转换  
✅ **热重载**：配置文件变化时自动重载  
✅ **兼容性**：与现有配置系统完全兼容  
✅ **易用性**：简单的 API 和全局便捷函数  
✅ **生产就绪**：完整的测试覆盖和错误处理  

这个配置系统为 YYHertz 框架提供了强大而灵活的配置管理能力，支持从简单的单机应用到复杂的分布式系统的各种配置需求。