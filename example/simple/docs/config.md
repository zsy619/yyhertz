# 系统配置

YYHertz框架提供了强大而灵活的配置管理系统，支持多种配置格式、环境变量注入、热更新等功能，让应用配置管理变得简单高效。

## 快速开始

### 1. 配置文件结构

```
conf/
├── app.yaml          # 应用基础配置
├── database.yaml     # 数据库配置
├── redis.yaml        # Redis配置
├── log.yaml          # 日志配置
├── session.yaml      # 会话配置
└── tls.yaml          # TLS/SSL配置
```

### 2. 基本使用

```go
import (
    "github.com/zsy619/yyhertz/framework/config"
)

// 定义配置结构体
type AppConfig struct {
    Name        string `yaml:"name" json:"name"`
    Version     string `yaml:"version" json:"version"`
    Port        int    `yaml:"port" json:"port"`
    Debug       bool   `yaml:"debug" json:"debug"`
    Environment string `yaml:"environment" json:"environment"`
}

func main() {
    // 加载配置
    var appConfig AppConfig
    err := config.LoadConfig("app", &appConfig)
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    // 使用配置
    fmt.Printf("应用名称: %s\n", appConfig.Name)
    fmt.Printf("运行端口: %d\n", appConfig.Port)
}
```

## 配置文件格式

### YAML格式

`conf/app.yaml`:
```yaml
# 应用基本配置
app:
  name: "YYHertz应用"
  version: "1.0.0"
  port: 8080
  debug: true
  environment: "development"
  
  # 服务器配置
  server:
    read_timeout: "30s"
    write_timeout: "30s"
    idle_timeout: "60s"
    max_header_bytes: 1048576
  
  # 跨域配置
  cors:
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["*"]
    allow_credentials: true
    max_age: 86400
```

### JSON格式

`conf/app.json`:
```json
{
  "app": {
    "name": "YYHertz应用",
    "version": "1.0.0",
    "port": 8080,
    "debug": true,
    "environment": "development",
    "server": {
      "read_timeout": "30s",
      "write_timeout": "30s",
      "idle_timeout": "60s",
      "max_header_bytes": 1048576
    }
  }
}
```

### TOML格式

`conf/app.toml`:
```toml
[app]
name = "YYHertz应用"
version = "1.0.0"
port = 8080
debug = true
environment = "development"

[app.server]
read_timeout = "30s"
write_timeout = "30s"
idle_timeout = "60s"
max_header_bytes = 1048576
```

## 环境变量支持

### 1. 环境变量注入

```yaml
# conf/app.yaml
app:
  name: "${APP_NAME:MyApp}"              # 环境变量APP_NAME，默认值MyApp
  port: "${APP_PORT:8080}"               # 环境变量APP_PORT，默认值8080
  debug: "${APP_DEBUG:false}"            # 环境变量APP_DEBUG，默认值false
  database_url: "${DATABASE_URL}"        # 必须的环境变量
```

### 2. 环境变量配置

```bash
# .env 文件
APP_NAME=生产环境应用
APP_PORT=80
APP_DEBUG=false
DATABASE_URL=mysql://user:pass@localhost:3306/db
REDIS_URL=redis://localhost:6379/0
```

### 3. 代码中使用

```go
import "github.com/zsy619/yyhertz/framework/config"

func main() {
    // 自动加载.env文件
    config.LoadEnv()
    
    // 加载配置（会自动替换环境变量）
    var appConfig AppConfig
    config.LoadConfig("app", &appConfig)
    
    // 直接获取环境变量
    dbURL := config.GetEnv("DATABASE_URL", "sqlite://app.db")
    redisURL := config.GetEnvRequired("REDIS_URL") // 必须存在的环境变量
}
```

## 配置管理

### 1. 全局配置管理器

```go
import "github.com/zsy619/yyhertz/framework/config"

func main() {
    // 初始化配置管理器
    configManager := config.NewManager()
    
    // 设置配置文件搜索路径
    configManager.AddConfigPath("./conf")
    configManager.AddConfigPath("./config")
    configManager.AddConfigPath("/etc/myapp")
    
    // 设置配置文件名
    configManager.SetConfigName("app")
    configManager.SetConfigType("yaml")
    
    // 读取配置
    err := configManager.ReadConfig()
    if err != nil {
        log.Fatal("读取配置失败:", err)
    }
    
    // 获取配置值
    appName := configManager.GetString("app.name")
    appPort := configManager.GetInt("app.port")
    appDebug := configManager.GetBool("app.debug")
}
```

### 2. 配置热更新

```go
import "github.com/zsy619/yyhertz/framework/config"

func main() {
    configManager := config.NewManager()
    
    // 启用配置文件监听
    configManager.WatchConfig()
    
    // 设置配置变更回调
    configManager.OnConfigChange(func(event config.Event) {
        log.Printf("配置文件发生变更: %s", event.Name)
        
        // 重新加载配置
        var newConfig AppConfig
        configManager.Unmarshal(&newConfig)
        
        // 应用新配置
        updateAppConfig(newConfig)
    })
    
    app := mvc.HertzApp
    app.Run()
}
```

### 3. 配置验证

```go
import (
    "github.com/zsy619/yyhertz/framework/config"
    "github.com/go-playground/validator/v10"
)

type AppConfig struct {
    Name        string `yaml:"name" validate:"required,min=1,max=50"`
    Port        int    `yaml:"port" validate:"required,min=1,max=65535"`
    Environment string `yaml:"environment" validate:"required,oneof=development staging production"`
    Email       string `yaml:"email" validate:"required,email"`
    URL         string `yaml:"url" validate:"required,url"`
}

func loadAndValidateConfig() (*AppConfig, error) {
    var appConfig AppConfig
    
    // 加载配置
    err := config.LoadConfig("app", &appConfig)
    if err != nil {
        return nil, err
    }
    
    // 验证配置
    validate := validator.New()
    err = validate.Struct(&appConfig)
    if err != nil {
        return nil, fmt.Errorf("配置验证失败: %w", err)
    }
    
    return &appConfig, nil
}
```

## 分层配置

### 1. 基础配置

`conf/base.yaml`:
```yaml
# 基础配置
app:
  name: "YYHertz应用"
  version: "1.0.0"
  
server:
  read_timeout: "30s"
  write_timeout: "30s"
  
database:
  driver: "mysql"
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10
```

### 2. 环境特定配置

`conf/development.yaml`:
```yaml
# 开发环境配置
app:
  debug: true
  port: 8080
  
database:
  host: "localhost"
  port: 3306
  username: "dev_user"
  password: "dev_pass"
  dbname: "myapp_dev"
  
log:
  level: "debug"
  format: "text"
```

`conf/production.yaml`:
```yaml
# 生产环境配置
app:
  debug: false
  port: 80
  
database:
  host: "${DB_HOST}"
  port: "${DB_PORT:3306}"
  username: "${DB_USER}"
  password: "${DB_PASS}"
  dbname: "${DB_NAME}"
  
log:
  level: "info"
  format: "json"
```

### 3. 配置合并

```go
func loadEnvironmentConfig() error {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    // 加载基础配置
    err := config.LoadConfig("base", &baseConfig)
    if err != nil {
        return err
    }
    
    // 加载环境特定配置并合并
    err = config.LoadConfigAndMerge(env, &baseConfig)
    if err != nil {
        return err
    }
    
    return nil
}
```

## 配置加密

### 1. 敏感配置加密

```yaml
# conf/app.yaml
app:
  name: "YYHertz应用"
  
database:
  # 使用ENC()标记加密字段
  password: "ENC(AQICAHhwm0YAXVoXdayqXjY9...)"
  
redis:
  password: "ENC(AQICAHhwm0YAXVoXdayqXjY9...)"
```

### 2. 配置解密

```go
import "github.com/zsy619/yyhertz/framework/config/crypto"

func main() {
    // 设置解密密钥
    crypto.SetEncryptionKey("your-encryption-key")
    
    // 加载并自动解密配置
    var appConfig AppConfig
    err := config.LoadConfigWithDecryption("app", &appConfig)
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
}
```

### 3. 配置加密工具

```bash
# 使用CLI工具加密敏感配置
yyhertz config encrypt --key="your-key" --value="sensitive-password"
# 输出: ENC(AQICAHhwm0YAXVoXdayqXjY9...)

# 解密配置验证
yyhertz config decrypt --key="your-key" --value="ENC(AQICAHhwm0YAXVoXdayqXjY9...)"
# 输出: sensitive-password
```

## 配置中心集成

### 1. Consul集成

```go
import "github.com/zsy619/yyhertz/framework/config/consul"

func main() {
    // 配置Consul客户端
    consulConfig := consul.Config{
        Address: "localhost:8500",
        Scheme:  "http",
        Token:   "your-consul-token",
    }
    
    // 创建Consul配置提供者
    provider, err := consul.NewProvider(consulConfig)
    if err != nil {
        log.Fatal("创建Consul提供者失败:", err)
    }
    
    // 加载远程配置
    configManager := config.NewManager()
    configManager.AddRemoteProvider(provider)
    
    // 从Consul加载配置
    err = configManager.ReadRemoteConfig("myapp/config")
    if err != nil {
        log.Fatal("读取远程配置失败:", err)
    }
}
```

### 2. Etcd集成

```go
import "github.com/zsy619/yyhertz/framework/config/etcd"

func main() {
    // 配置Etcd客户端
    etcdConfig := etcd.Config{
        Endpoints: []string{"localhost:2379"},
        Username:  "etcd-user",
        Password:  "etcd-pass",
    }
    
    // 创建Etcd配置提供者
    provider, err := etcd.NewProvider(etcdConfig)
    if err != nil {
        log.Fatal("创建Etcd提供者失败:", err)
    }
    
    // 监听配置变更
    provider.Watch("/myapp/config", func(value string) {
        log.Printf("配置发生变更: %s", value)
        // 重新加载配置
        reloadConfig()
    })
}
```

## 高级功能

### 1. 配置模板

```yaml
# conf/template.yaml
app:
  name: "{{.APP_NAME}}"
  port: {{.APP_PORT}}
  database:
    host: "{{.DB_HOST}}"
    port: {{.DB_PORT}}
    
# 模板变量文件 vars.yaml
APP_NAME: "生产应用"
APP_PORT: 8080
DB_HOST: "prod-db.example.com"
DB_PORT: 3306
```

```go
// 渲染配置模板
configContent, err := config.RenderTemplate("template.yaml", "vars.yaml")
if err != nil {
    log.Fatal("渲染配置模板失败:", err)
}
```

### 2. 配置版本管理

```go
import "github.com/zsy619/yyhertz/framework/config/version"

func main() {
    // 启用配置版本管理
    versionManager := version.NewManager()
    
    // 保存配置快照
    err := versionManager.SaveSnapshot("v1.0.0", currentConfig)
    if err != nil {
        log.Error("保存配置快照失败:", err)
    }
    
    // 回滚到指定版本
    err = versionManager.Rollback("v1.0.0")
    if err != nil {
        log.Error("配置回滚失败:", err)
    }
    
    // 获取版本历史
    versions := versionManager.GetVersionHistory()
    for _, v := range versions {
        fmt.Printf("版本: %s, 时间: %s\n", v.Version, v.Timestamp)
    }
}
```

### 3. 配置API

```go
// 配置管理API
func (c *AdminController) GetConfig() {
    configData := config.GetAllConfig()
    c.JSON(map[string]interface{}{
        "success": true,
        "data":    configData,
    })
}

func (c *AdminController) PostUpdateConfig() {
    var updateData map[string]interface{}
    err := c.BindJSON(&updateData)
    if err != nil {
        c.Error(400, "请求数据格式错误")
        return
    }
    
    // 更新配置
    for key, value := range updateData {
        config.Set(key, value)
    }
    
    // 保存配置
    err = config.WriteConfig()
    if err != nil {
        c.Error(500, "保存配置失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "message": "配置更新成功",
    })
}
```

## 最佳实践

### 1. 配置结构设计

```go
// 推荐的配置结构
type Config struct {
    App      AppConfig      `yaml:"app"`
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
    Log      LogConfig      `yaml:"log"`
    Email    EmailConfig    `yaml:"email"`
}

type AppConfig struct {
    Name        string `yaml:"name" validate:"required"`
    Version     string `yaml:"version" validate:"required"`
    Environment string `yaml:"environment" validate:"required,oneof=dev test prod"`
    Debug       bool   `yaml:"debug"`
    TimeZone    string `yaml:"timezone" validate:"required"`
}
```

### 2. 环境管理

```bash
# 不同环境使用不同的配置文件
APP_ENV=development go run main.go  # 加载 development.yaml
APP_ENV=production go run main.go   # 加载 production.yaml
APP_ENV=testing go run main.go      # 加载 testing.yaml
```

### 3. 配置安全

- 敏感信息使用环境变量或加密存储
- 不要将包含敏感信息的配置文件提交到版本控制
- 使用配置验证确保配置的正确性
- 定期轮换敏感配置如密钥、密码等

### 4. 配置监控

```go
// 配置变更监控
func monitorConfigChanges() {
    config.WatchConfig()
    config.OnConfigChange(func(event config.Event) {
        // 记录配置变更日志
        log.WithFields(log.Fields{
            "file":      event.Name,
            "operation": event.Op,
            "timestamp": time.Now(),
        }).Info("配置文件发生变更")
        
        // 发送告警通知
        alerting.SendConfigChangeNotification(event)
        
        // 验证新配置
        if err := validateNewConfig(); err != nil {
            log.Error("新配置验证失败:", err)
            // 可以选择回滚到上一个版本
        }
    })
}
```

## 配置示例

### 完整的应用配置

```yaml
# conf/app.yaml
app:
  name: "YYHertz Web应用"
  version: "1.2.0"
  environment: "${APP_ENV:development}"
  debug: "${APP_DEBUG:true}"
  timezone: "Asia/Shanghai"
  
server:
  host: "${SERVER_HOST:0.0.0.0}"
  port: "${SERVER_PORT:8080}"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  max_header_bytes: 1048576
  
database:
  driver: "mysql"
  host: "${DB_HOST:localhost}"
  port: "${DB_PORT:3306}"
  username: "${DB_USER:root}"
  password: "${DB_PASS:password}"
  dbname: "${DB_NAME:myapp}"
  charset: "utf8mb4"
  parse_time: true
  loc: "Local"
  max_open_conns: "${DB_MAX_OPEN_CONNS:100}"
  max_idle_conns: "${DB_MAX_IDLE_CONNS:10}"
  conn_max_lifetime: "3600s"
  
redis:
  host: "${REDIS_HOST:localhost}"
  port: "${REDIS_PORT:6379}"
  password: "${REDIS_PASS:}"
  db: "${REDIS_DB:0}"
  pool_size: "${REDIS_POOL_SIZE:10}"
  
log:
  level: "${LOG_LEVEL:info}"
  format: "${LOG_FORMAT:json}"
  outputs:
    - type: "console"
      color: true
    - type: "file"
      filename: "logs/app.log"
      max_size: 100
      max_backups: 10
      max_age: 30
      compress: true
```