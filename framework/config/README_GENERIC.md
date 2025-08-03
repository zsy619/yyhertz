# YYHertz 泛型配置管理器

## 概述

YYHertz 框架现在支持泛型配置管理器，允许你创建类型安全、可复用的配置管理系统。这个新系统保持了与原有 `ViperConfigManager` 的完全兼容性，同时提供了更强大和灵活的配置管理能力。

## 主要特性

- ✨ **类型安全的泛型** - 编译时类型检查，避免运行时错误
- 🔄 **完全可复用** - 任何结构体都可以成为配置类型
- 🛠️ **统一接口** - 所有配置类型使用相同的管理接口
- ⚡ **高性能** - 单例模式和并发安全设计
- 🔧 **自动化** - 自动默认值设置和配置文件生成
- 📁 **多文件支持** - 支持不同配置类型使用不同的配置文件
- 🔄 **向后兼容** - 原有代码无需修改即可工作

## 快速开始

### 1. 使用内置配置类型

#### 应用配置 (AppConfig)

```go
import "github.com/zsy619/yyhertz/framework/config"

// 方式1：获取完整配置
appConfig, err := config.GetAppConfig()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("应用名称: %s\n", appConfig.App.Name)
fmt.Printf("数据库主机: %s\n", appConfig.Database.Host)

// 方式2：使用配置管理器
manager := config.GetAppConfigManager()
appName := manager.GetString("app.name")
dbHost := manager.GetString("database.host")

// 方式3：使用泛型便捷函数
appPort := config.GetGenericConfigInt(config.AppConfig{}, "app.port")
debugMode := config.GetGenericConfigBool(config.AppConfig{}, "app.debug")
```

#### 模板配置 (TemplateConfig)

```go
// 方式1：获取完整配置
templateConfig, err := config.GetTemplateConfig()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("模板目录: %s\n", templateConfig.Engine.Directory)
fmt.Printf("启用缓存: %v\n", templateConfig.Cache.Enable)

// 方式2：使用配置管理器
manager := config.GetTemplateConfigManager()
templateType := manager.GetString("engine.type")
reloadEnabled := manager.GetBool("engine.reload")

// 方式3：使用泛型便捷函数
staticRoot := config.GetGenericConfigString(config.TemplateConfig{}, "static.root")
liveReload := config.GetGenericConfigBool(config.TemplateConfig{}, "development.live_reload")
```

### 2. 创建自定义配置类型

#### 步骤1：定义配置结构体

```go
type DatabaseConfig struct {
    Host     string `mapstructure:"host" yaml:"host"`
    Port     int    `mapstructure:"port" yaml:"port"`
    Username string `mapstructure:"username" yaml:"username"`
    Password string `mapstructure:"password" yaml:"password"`
    Database string `mapstructure:"database" yaml:"database"`
    MaxConns int    `mapstructure:"max_conns" yaml:"max_conns"`
}
```

#### 步骤2：实现 ConfigInterface 接口

```go
func (c DatabaseConfig) GetConfigName() string {
    return "database"  // 配置文件名（不含扩展名）
}

func (c DatabaseConfig) SetDefaults(v *viper.Viper) {
    v.SetDefault("host", "localhost")
    v.SetDefault("port", 3306)
    v.SetDefault("username", "root")
    v.SetDefault("password", "")
    v.SetDefault("database", "myapp")
    v.SetDefault("max_conns", 100)
}

func (c DatabaseConfig) GenerateDefaultContent() string {
    return `# 数据库配置
host: "localhost"
port: 3306
username: "root"
password: ""
database: "myapp"
max_conns: 100
`
}
```

#### 步骤3：使用自定义配置

```go
// 获取配置管理器
manager := config.GetGenericConfigManager(DatabaseConfig{})

// 获取配置值
host := manager.GetString("host")
port := manager.GetInt("port")

// 设置配置值
manager.Set("host", "192.168.1.100")
manager.Set("port", 3307)

// 获取完整配置
dbConfig, err := manager.GetConfig()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("数据库配置: %+v\n", dbConfig)
```

## 配置文件管理

### 自动配置文件生成

当配置文件不存在时，系统会自动创建默认配置文件：

- `conf/app.yaml` - 应用配置
- `conf/template.yaml` - 模板配置  
- `conf/database.yaml` - 自定义数据库配置
- `conf/{ConfigName}.yaml` - 其他自定义配置

### 配置文件监听

```go
// 监听配置文件变化
config.WatchGenericConfig(config.AppConfig{})        // 监听应用配置
config.WatchGenericConfig(config.TemplateConfig{})   // 监听模板配置
config.WatchGenericConfig(DatabaseConfig{})          // 监听自定义配置
```

## 高级用法

### 动态配置设置

```go
// 动态设置应用配置
config.SetGenericConfigValue(config.AppConfig{}, "app.debug", false)
config.SetGenericConfigValue(config.AppConfig{}, "database.host", "prod-db.example.com")

// 动态设置模板配置
config.SetGenericConfigValue(config.TemplateConfig{}, "engine.type", "pug")
config.SetGenericConfigValue(config.TemplateConfig{}, "cache.enable", false)

// 动态设置自定义配置
config.SetGenericConfigValue(DatabaseConfig{}, "max_conns", 200)
```

### 批量配置操作

```go
// 批量获取配置值
appName := config.GetGenericConfigString(config.AppConfig{}, "app.name")
appPort := config.GetGenericConfigInt(config.AppConfig{}, "app.port")
appDebug := config.GetGenericConfigBool(config.AppConfig{}, "app.debug")

// 批量设置配置值
manager := config.GetAppConfigManager()
manager.Set("app.name", "ProductionApp")
manager.Set("app.port", 80)
manager.Set("app.debug", false)
```

## 并发安全

泛型配置管理器是并发安全的，你可以在多个 goroutine 中安全地读取和写入配置：

```go
go func() {
    // 协程1：读取配置
    appName := config.GetGenericConfigString(config.AppConfig{}, "app.name")
    fmt.Println("应用名称:", appName)
}()

go func() {
    // 协程2：设置配置
    config.SetGenericConfigValue(config.AppConfig{}, "app.debug", false)
}()
```

## 性能特性

- **单例模式** - 每种配置类型只创建一个管理器实例
- **读写锁** - 支持多读单写，提高并发性能
- **延迟初始化** - 配置管理器在首次访问时才初始化
- **内存缓存** - 配置值被缓存在内存中，避免重复解析

## 迁移指南

### 从旧版本迁移

原有代码无需修改即可继续工作：

```go
// 原有代码仍然有效
config, err := config.GetGlobalConfig()
appName := config.GetConfigString("app.name")
```

### 推荐的新写法

```go
// 推荐的新写法
config, err := config.GetAppConfig()
appName := config.GetGenericConfigString(config.AppConfig{}, "app.name")
```

## 最佳实践

1. **使用类型安全的方法** - 优先使用泛型配置管理器而不是字符串键值
2. **创建专用配置类型** - 为不同的功能模块创建专门的配置结构体
3. **合理设置默认值** - 在 `SetDefaults` 方法中设置合理的默认值
4. **监听配置变化** - 在生产环境中启用配置文件监听
5. **配置验证** - 在配置结构体中添加验证逻辑

## 示例项目

查看 `framework/examples/config_usage.go` 文件以获取完整的使用示例。

## 常见问题

### Q: 如何处理配置验证？
A: 在配置结构体中实现验证方法，或在获取配置后进行验证。

### Q: 可以使用嵌套配置吗？
A: 是的，支持任意深度的嵌套配置结构。

### Q: 如何处理敏感配置？
A: 建议使用环境变量或外部密钥管理系统，而不是直接写在配置文件中。

### Q: 性能如何？
A: 泛型配置管理器具有优异的性能，支持高并发访问，配置值被缓存在内存中。

## 技术支持

如果遇到问题或需要技术支持，请：

1. 查看框架文档
2. 检查示例代码
3. 提交 Issue 到项目仓库