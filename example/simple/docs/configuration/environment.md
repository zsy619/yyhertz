# 环境配置

YYHertz 框架提供了灵活的环境配置管理功能，支持多环境部署、配置文件管理、环境变量注入等特性，帮助开发者轻松管理不同环境下的应用配置。

## 概述

环境配置管理是现代应用部署的重要环节。YYHertz 的环境配置系统提供：

- 多环境支持（开发、测试、生产）
- 配置文件分层管理
- 环境变量自动注入
- 配置热重载
- 敏感信息加密
- 配置验证和类型转换

## 环境类型

### 预定义环境

```go
package config

const (
    EnvDevelopment = "development"
    EnvTesting     = "testing"
    EnvStaging     = "staging"
    EnvProduction  = "production"
)

// 获取当前环境
func GetEnvironment() string {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = EnvDevelopment
    }
    return env
}

// 环境检查函数
func IsDevelopment() bool {
    return GetEnvironment() == EnvDevelopment
}

func IsProduction() bool {
    return GetEnvironment() == EnvProduction
}

func IsTesting() bool {
    return GetEnvironment() == EnvTesting
}
```

### 应用启动配置

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/config"
)

func main() {
    // 设置环境
    env := config.GetEnvironment()
    
    app := mvc.HertzApp
    app.SetEnvironment(env)
    
    // 根据环境配置应用
    switch env {
    case config.EnvDevelopment:
        setupDevelopmentConfig(app)
    case config.EnvTesting:
        setupTestingConfig(app)
    case config.EnvProduction:
        setupProductionConfig(app)
    }
    
    app.Run()
}

func setupDevelopmentConfig(app *mvc.App) {
    app.SetDebug(true)
    app.SetLogLevel("debug")
    
    // 启用开发工具
    app.EnableDevTools()
    app.EnableHotReload()
    
    // 详细错误输出
    app.EnableDetailedErrors(true)
}

func setupProductionConfig(app *mvc.App) {
    app.SetDebug(false)
    app.SetLogLevel("error")
    
    // 禁用开发工具
    app.DisableDevTools()
    
    // 启用性能优化
    app.EnableGzip(true)
    app.EnableCaching(true)
    
    // 安全配置
    app.EnableSecurityHeaders(true)
}
```

## 配置文件管理

### 分层配置结构

```
config/
├── default.yaml          # 默认配置
├── development.yaml       # 开发环境配置
├── testing.yaml          # 测试环境配置
├── staging.yaml          # 预发布环境配置
├── production.yaml       # 生产环境配置
└── local.yaml            # 本地覆盖配置（不提交到版本控制）
```

### 基础配置文件

```yaml
# config/default.yaml
app:
  name: "YYHertz App"
  version: "1.0.0"
  host: "localhost"
  port: 8080
  debug: false

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  name: "yyhertz"
  username: "root"
  password: ""
  charset: "utf8mb4"
  max_connections: 100
  max_idle: 10

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file: ""

session:
  driver: "memory"
  lifetime: 7200
  cookie_name: "session_id"
  
cache:
  driver: "memory"
  prefix: "yyhertz_"
  ttl: 3600
```

### 环境特定配置

```yaml
# config/development.yaml
app:
  debug: true
  port: 8080

database:
  host: "localhost"
  name: "yyhertz_dev"
  username: "dev_user"
  password: "dev_password"

logging:
  level: "debug"
  output: "stdout"

cache:
  driver: "memory"

session:
  driver: "memory"
  
# 开发工具配置
dev_tools:
  hot_reload: true
  profiling: true
  
# 邮件配置（开发环境使用本地SMTP）
mail:
  driver: "log"  # 将邮件输出到日志而不是发送
```

```yaml
# config/production.yaml
app:
  debug: false
  host: "0.0.0.0"
  port: 80

database:
  host: "${DB_HOST}"
  port: "${DB_PORT:3306}"
  name: "${DB_NAME}"
  username: "${DB_USER}"
  password: "${DB_PASSWORD}"
  max_connections: 200
  
redis:
  host: "${REDIS_HOST}"
  port: "${REDIS_PORT:6379}"
  password: "${REDIS_PASSWORD}"
  
logging:
  level: "error"
  format: "json"
  output: "file"
  file: "/var/log/yyhertz/app.log"
  
cache:
  driver: "redis"
  prefix: "prod_yyhertz_"
  
session:
  driver: "redis"
  
# 安全配置
security:
  csrf_protection: true
  rate_limiting: true
  https_only: true
```

### 配置加载器

```go
package config

import (
    "fmt"
    "os"
    "strings"
    "gopkg.in/yaml.v3"
)

type ConfigLoader struct {
    env        string
    configPath string
    config     map[string]interface{}
}

func NewConfigLoader(env, configPath string) *ConfigLoader {
    return &ConfigLoader{
        env:        env,
        configPath: configPath,
        config:     make(map[string]interface{}),
    }
}

func (cl *ConfigLoader) Load() error {
    // 加载顺序：default -> environment -> local
    files := []string{
        "default.yaml",
        cl.env + ".yaml",
        "local.yaml",
    }
    
    for _, file := range files {
        filePath := filepath.Join(cl.configPath, file)
        if err := cl.loadFile(filePath); err != nil {
            // local.yaml 是可选的
            if file != "local.yaml" {
                return err
            }
        }
    }
    
    // 替换环境变量
    cl.replaceEnvVars()
    
    return nil
}

func (cl *ConfigLoader) loadFile(filePath string) error {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }
    
    var fileConfig map[string]interface{}
    if err := yaml.Unmarshal(data, &fileConfig); err != nil {
        return err
    }
    
    // 深度合并配置
    cl.mergeConfig(cl.config, fileConfig)
    
    return nil
}

func (cl *ConfigLoader) mergeConfig(target, source map[string]interface{}) {
    for key, value := range source {
        if targetValue, exists := target[key]; exists {
            if targetMap, ok := targetValue.(map[string]interface{}); ok {
                if sourceMap, ok := value.(map[string]interface{}); ok {
                    cl.mergeConfig(targetMap, sourceMap)
                    continue
                }
            }
        }
        target[key] = value
    }
}

func (cl *ConfigLoader) replaceEnvVars() {
    cl.replaceEnvVarsRecursive(cl.config)
}

func (cl *ConfigLoader) replaceEnvVarsRecursive(obj map[string]interface{}) {
    for key, value := range obj {
        switch v := value.(type) {
        case string:
            obj[key] = cl.expandEnvVar(v)
        case map[string]interface{}:
            cl.replaceEnvVarsRecursive(v)
        case []interface{}:
            for i, item := range v {
                if str, ok := item.(string); ok {
                    v[i] = cl.expandEnvVar(str)
                }
            }
        }
    }
}

func (cl *ConfigLoader) expandEnvVar(value string) string {
    // 支持 ${VAR} 和 ${VAR:default} 格式
    if !strings.Contains(value, "${") {
        return value
    }
    
    return os.Expand(value, func(key string) string {
        // 处理默认值 VAR:default
        parts := strings.SplitN(key, ":", 2)
        envVar := parts[0]
        
        envValue := os.Getenv(envVar)
        if envValue == "" && len(parts) > 1 {
            return parts[1] // 返回默认值
        }
        
        return envValue
    })
}

func (cl *ConfigLoader) Get(key string) interface{} {
    keys := strings.Split(key, ".")
    current := cl.config
    
    for _, k := range keys {
        if value, exists := current[k]; exists {
            if nextMap, ok := value.(map[string]interface{}); ok {
                current = nextMap
            } else {
                return value
            }
        } else {
            return nil
        }
    }
    
    return current
}

func (cl *ConfigLoader) GetString(key string) string {
    if value := cl.Get(key); value != nil {
        if str, ok := value.(string); ok {
            return str
        }
    }
    return ""
}

func (cl *ConfigLoader) GetInt(key string) int {
    if value := cl.Get(key); value != nil {
        if num, ok := value.(int); ok {
            return num
        }
        if num, ok := value.(float64); ok {
            return int(num)
        }
    }
    return 0
}

func (cl *ConfigLoader) GetBool(key string) bool {
    if value := cl.Get(key); value != nil {
        if b, ok := value.(bool); ok {
            return b
        }
    }
    return false
}
```

## 环境变量管理

### .env 文件支持

```bash
# .env.development
APP_ENV=development
APP_DEBUG=true
APP_PORT=8080

DB_HOST=localhost
DB_PORT=3306
DB_NAME=yyhertz_dev
DB_USER=dev_user
DB_PASSWORD=dev_password

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 邮件配置
MAIL_DRIVER=smtp
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=your-email@gmail.com
MAIL_PASSWORD=your-password

# 第三方服务
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
```

```bash
# .env.production
APP_ENV=production
APP_DEBUG=false
APP_PORT=80

DB_HOST=prod-db.example.com
DB_PORT=3306
DB_NAME=yyhertz_prod
DB_USER=prod_user
DB_PASSWORD=super_secret_password

REDIS_HOST=prod-redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=redis_password

# SSL 配置
SSL_CERT_PATH=/etc/ssl/certs/app.crt
SSL_KEY_PATH=/etc/ssl/private/app.key

# 监控配置
SENTRY_DSN=https://your-sentry-dsn
```

### 环境变量加载器

```go
package config

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type EnvLoader struct {
    envFile string
}

func NewEnvLoader(envFile string) *EnvLoader {
    return &EnvLoader{
        envFile: envFile,
    }
}

func (el *EnvLoader) Load() error {
    file, err := os.Open(el.envFile)
    if err != nil {
        // .env 文件是可选的
        if os.IsNotExist(err) {
            return nil
        }
        return err
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    lineNum := 0
    
    for scanner.Scan() {
        lineNum++
        line := strings.TrimSpace(scanner.Text())
        
        // 跳过空行和注释
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        
        // 解析键值对
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            return fmt.Errorf("invalid format at line %d: %s", lineNum, line)
        }
        
        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        
        // 移除引号
        value = el.unquote(value)
        
        // 只有在环境变量不存在时才设置
        if os.Getenv(key) == "" {
            os.Setenv(key, value)
        }
    }
    
    return scanner.Err()
}

func (el *EnvLoader) unquote(value string) string {
    // 移除双引号或单引号
    if len(value) >= 2 {
        if (value[0] == '"' && value[len(value)-1] == '"') ||
           (value[0] == '\'' && value[len(value)-1] == '\'') {
            return value[1 : len(value)-1]
        }
    }
    return value
}

// 自动加载环境文件
func LoadEnvFile() error {
    env := GetEnvironment()
    
    // 尝试加载环境特定的 .env 文件
    envFiles := []string{
        fmt.Sprintf(".env.%s.local", env),
        ".env.local",
        fmt.Sprintf(".env.%s", env),
        ".env",
    }
    
    for _, file := range envFiles {
        loader := NewEnvLoader(file)
        if err := loader.Load(); err != nil {
            return fmt.Errorf("failed to load %s: %w", file, err)
        }
    }
    
    return nil
}
```

## 配置结构体

### 强类型配置

```go
package config

import (
    "time"
)

type AppConfig struct {
    App      AppSettings      `yaml:"app"`
    Database DatabaseConfig  `yaml:"database"`
    Redis    RedisConfig     `yaml:"redis"`
    Logging  LoggingConfig   `yaml:"logging"`
    Session  SessionConfig   `yaml:"session"`
    Cache    CacheConfig     `yaml:"cache"`
    Mail     MailConfig      `yaml:"mail"`
    Security SecurityConfig  `yaml:"security"`
}

type AppSettings struct {
    Name    string `yaml:"name"`
    Version string `yaml:"version"`
    Host    string `yaml:"host"`
    Port    int    `yaml:"port"`
    Debug   bool   `yaml:"debug"`
    Env     string `yaml:"env"`
}

type DatabaseConfig struct {
    Driver         string        `yaml:"driver"`
    Host           string        `yaml:"host"`
    Port           int           `yaml:"port"`
    Name           string        `yaml:"name"`
    Username       string        `yaml:"username"`
    Password       string        `yaml:"password"`
    Charset        string        `yaml:"charset"`
    MaxConnections int           `yaml:"max_connections"`
    MaxIdle        int           `yaml:"max_idle"`
    ConnectTimeout time.Duration `yaml:"connect_timeout"`
}

type RedisConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
    PoolSize int    `yaml:"pool_size"`
}

type LoggingConfig struct {
    Level  string `yaml:"level"`
    Format string `yaml:"format"`
    Output string `yaml:"output"`
    File   string `yaml:"file"`
}

type SessionConfig struct {
    Driver     string `yaml:"driver"`
    Lifetime   int    `yaml:"lifetime"`
    CookieName string `yaml:"cookie_name"`
}

type CacheConfig struct {
    Driver string `yaml:"driver"`
    Prefix string `yaml:"prefix"`
    TTL    int    `yaml:"ttl"`
}

type MailConfig struct {
    Driver   string `yaml:"driver"`
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    From     string `yaml:"from"`
}

type SecurityConfig struct {
    CSRFProtection bool `yaml:"csrf_protection"`
    RateLimiting   bool `yaml:"rate_limiting"`
    HTTPSOnly      bool `yaml:"https_only"`
}
```

### 配置验证

```go
package config

import (
    "errors"
    "fmt"
    "net"
    "strconv"
)

func (ac *AppConfig) Validate() error {
    if err := ac.App.Validate(); err != nil {
        return fmt.Errorf("app config: %w", err)
    }
    
    if err := ac.Database.Validate(); err != nil {
        return fmt.Errorf("database config: %w", err)
    }
    
    if err := ac.Redis.Validate(); err != nil {
        return fmt.Errorf("redis config: %w", err)
    }
    
    return nil
}

func (as *AppSettings) Validate() error {
    if as.Name == "" {
        return errors.New("app name is required")
    }
    
    if as.Port <= 0 || as.Port > 65535 {
        return errors.New("invalid port number")
    }
    
    if as.Host == "" {
        as.Host = "localhost"
    }
    
    return nil
}

func (dc *DatabaseConfig) Validate() error {
    if dc.Driver == "" {
        return errors.New("database driver is required")
    }
    
    if dc.Host == "" {
        return errors.New("database host is required")
    }
    
    if dc.Name == "" {
        return errors.New("database name is required")
    }
    
    if dc.Port <= 0 || dc.Port > 65535 {
        return errors.New("invalid database port")
    }
    
    // 验证主机地址
    if net.ParseIP(dc.Host) == nil {
        if _, err := net.LookupHost(dc.Host); err != nil {
            return fmt.Errorf("invalid database host: %w", err)
        }
    }
    
    return nil
}

func (rc *RedisConfig) Validate() error {
    if rc.Host == "" {
        rc.Host = "localhost"
    }
    
    if rc.Port <= 0 {
        rc.Port = 6379
    }
    
    if rc.DB < 0 {
        rc.DB = 0
    }
    
    if rc.PoolSize <= 0 {
        rc.PoolSize = 10
    }
    
    return nil
}
```

## 配置热重载

### 配置监控

```go
package config

import (
    "log"
    "sync"
    "time"
    "github.com/fsnotify/fsnotify"
)

type ConfigWatcher struct {
    config     *AppConfig
    configPath string
    loader     *ConfigLoader
    mutex      sync.RWMutex
    callbacks  []func(*AppConfig)
}

func NewConfigWatcher(configPath string) *ConfigWatcher {
    return &ConfigWatcher{
        configPath: configPath,
        callbacks:  make([]func(*AppConfig), 0),
    }
}

func (cw *ConfigWatcher) Start() error {
    // 初始加载配置
    if err := cw.loadConfig(); err != nil {
        return err
    }
    
    // 创建文件监控器
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    
    // 监控配置目录
    if err := watcher.Add(cw.configPath); err != nil {
        return err
    }
    
    go cw.watchLoop(watcher)
    
    return nil
}

func (cw *ConfigWatcher) watchLoop(watcher *fsnotify.Watcher) {
    defer watcher.Close()
    
    debouncer := NewDebouncer(1 * time.Second)
    
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return
            }
            
            if event.Op&fsnotify.Write == fsnotify.Write {
                debouncer.Call(func() {
                    cw.reloadConfig()
                })
            }
            
        case err, ok := <-watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Config watcher error: %v", err)
        }
    }
}

func (cw *ConfigWatcher) reloadConfig() {
    log.Println("Reloading configuration...")
    
    oldConfig := cw.GetConfig()
    
    if err := cw.loadConfig(); err != nil {
        log.Printf("Failed to reload config: %v", err)
        return
    }
    
    newConfig := cw.GetConfig()
    
    // 检查配置是否真的发生了变化
    if !cw.configChanged(oldConfig, newConfig) {
        return
    }
    
    log.Println("Configuration changed, notifying callbacks...")
    
    // 通知回调函数
    for _, callback := range cw.callbacks {
        go callback(newConfig)
    }
}

func (cw *ConfigWatcher) loadConfig() error {
    env := GetEnvironment()
    loader := NewConfigLoader(env, cw.configPath)
    
    if err := loader.Load(); err != nil {
        return err
    }
    
    var config AppConfig
    if err := loader.Unmarshal(&config); err != nil {
        return err
    }
    
    if err := config.Validate(); err != nil {
        return err
    }
    
    cw.mutex.Lock()
    cw.config = &config
    cw.loader = loader
    cw.mutex.Unlock()
    
    return nil
}

func (cw *ConfigWatcher) GetConfig() *AppConfig {
    cw.mutex.RLock()
    defer cw.mutex.RUnlock()
    return cw.config
}

func (cw *ConfigWatcher) OnConfigChange(callback func(*AppConfig)) {
    cw.callbacks = append(cw.callbacks, callback)
}

func (cw *ConfigWatcher) configChanged(old, new *AppConfig) bool {
    // 简单的配置比较，实际可以更精细
    return old.App.Debug != new.App.Debug ||
           old.Logging.Level != new.Logging.Level ||
           old.Database.MaxConnections != new.Database.MaxConnections
}
```

## Docker 环境配置

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 复制构建产物
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY --from=builder /app/static ./static
COPY --from=builder /app/views ./views

# 设置环境变量
ENV APP_ENV=production
ENV APP_PORT=8080

EXPOSE 8080

CMD ["./main"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=db
      - DB_NAME=yyhertz
      - DB_USER=root
      - DB_PASSWORD=secret
      - REDIS_HOST=redis
    depends_on:
      - db
      - redis
    volumes:
      - ./logs:/var/log/yyhertz
      
  db:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=secret
      - MYSQL_DATABASE=yyhertz
    volumes:
      - db_data:/var/lib/mysql
      
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  db_data:
  redis_data:
```

## 最佳实践

### 1. 配置安全

```go
// 敏感信息加密
type SecureConfig struct {
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
}

func (sc *SecureConfig) Decrypt(key []byte) error {
    if err := decryptField(&sc.Database.Password, key); err != nil {
        return err
    }
    
    if err := decryptField(&sc.Redis.Password, key); err != nil {
        return err
    }
    
    return nil
}

// 配置验证
func validateConfig() {
    requiredEnvVars := []string{
        "DB_PASSWORD",
        "REDIS_PASSWORD",
        "JWT_SECRET",
    }
    
    for _, envVar := range requiredEnvVars {
        if os.Getenv(envVar) == "" {
            log.Fatalf("Required environment variable %s is not set", envVar)
        }
    }
}
```

### 2. 配置版本控制

```bash
# .gitignore
.env
.env.local
.env.*.local
config/local.yaml
config/production.yaml  # 如果包含敏感信息

# 提交模板文件
.env.example
config/production.yaml.example
```

### 3. 配置文档

```yaml
# config/README.md
# 配置说明

## 环境变量

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| APP_ENV | 否 | development | 应用环境 |
| APP_DEBUG | 否 | false | 调试模式 |
| DB_HOST | 是 | - | 数据库主机 |
| DB_PASSWORD | 是 | - | 数据库密码 |

## 配置文件优先级

1. 环境变量
2. local.yaml
3. {environment}.yaml
4. default.yaml
```

YYHertz 的环境配置系统提供了灵活而强大的配置管理功能，支持多环境部署、配置热重载和安全管理，帮助开发者构建可维护的应用程序。
