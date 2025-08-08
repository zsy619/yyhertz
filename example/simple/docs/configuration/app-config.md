# ⚙️ 应用配置

应用配置是YYHertz MVC框架的核心组成部分，提供了灵活的配置管理机制，支持多环境配置、动态加载和热更新等功能。

## 🌟 核心特性

### ✨ 配置功能
- **📁 多格式支持** - YAML、JSON、TOML、Properties等格式
- **🌍 环境隔离** - 开发、测试、生产环境配置隔离
- **🔄 动态加载** - 运行时配置更新和热重载
- **🔒 安全管理** - 敏感信息加密和环境变量注入
- **📊 验证机制** - 配置格式验证和完整性检查

### 🎪 高级功能
- **🎯 配置继承** - 配置文件继承和覆盖机制
- **🔍 配置发现** - 自动配置发现和加载
- **📈 配置监控** - 配置变更监控和日志记录
- **🌐 远程配置** - 支持远程配置中心集成

## 🚀 配置文件结构

### 1. 基础配置目录

```
config/
├── app.yaml              # 应用主配置
├── app.development.yaml  # 开发环境配置
├── app.testing.yaml      # 测试环境配置
├── app.production.yaml   # 生产环境配置
├── database.yaml         # 数据库配置
├── cache.yaml           # 缓存配置
├── middleware.yaml      # 中间件配置
├── logging.yaml         # 日志配置
└── custom/              # 自定义配置目录
    ├── features.yaml    # 功能特性配置
    └── integrations.yaml # 第三方集成配置
```

### 2. 主配置文件示例

```yaml
# config/app.yaml
app:
  # 应用基本信息
  name: "YYHertz MVC App"
  version: "1.0.0"
  description: "基于YYHertz构建的MVC应用"
  
  # 服务器配置
  server:
    host: "0.0.0.0"
    port: 8888
    mode: "debug" # debug, test, release
    read_timeout: "30s"
    write_timeout: "30s"
    idle_timeout: "60s"
    max_header_bytes: 1048576
    
    # TLS配置
    tls:
      enabled: false
      cert_file: ""
      key_file: ""
      auto_cert: false
      
    # HTTP/2配置
    http2:
      enabled: true
      max_concurrent_streams: 250
      
  # 应用路径配置
  paths:
    static: "./static"
    views: "./views"
    uploads: "./uploads"
    logs: "./logs"
    temp: "./temp"
    
  # 模板配置
  template:
    engine: "html/template"
    cache: true
    reload: false
    delims:
      left: "{{"
      right: "}}"
    functions: []
    
  # 会话配置
  session:
    provider: "memory" # memory, redis, database
    cookie_name: "session_id"
    cookie_path: "/"
    cookie_domain: ""
    cookie_secure: false
    cookie_http_only: true
    cookie_same_site: "lax"
    max_age: 3600
    
  # 安全配置
  security:
    csrf:
      enabled: true
      token_length: 32
      cookie_name: "csrf_token"
      header_name: "X-CSRF-Token"
      
    cors:
      enabled: true
      allow_origins: ["*"]
      allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
      allow_headers: ["*"]
      allow_credentials: false
      max_age: 86400
      
    rate_limit:
      enabled: true
      requests_per_minute: 60
      burst: 10
      
  # 数据库配置引用
  database:
    default: "${DATABASE_URL:sqlite://./data/app.db}"
    
  # 缓存配置引用
  cache:
    default: "memory"
    
  # 日志配置引用
  logging:
    level: "info"
    format: "json"
    
# 功能开关
features:
  user_registration: true
  email_verification: true
  two_factor_auth: false
  api_versioning: true
  
# 第三方服务配置
services:
  email:
    provider: "smtp"
    smtp:
      host: "${SMTP_HOST:localhost}"
      port: ${SMTP_PORT:587}
      username: "${SMTP_USER:}"
      password: "${SMTP_PASS:}"
      tls: true
      
  sms:
    provider: "aliyun"
    aliyun:
      access_key: "${SMS_ACCESS_KEY:}"
      secret_key: "${SMS_SECRET_KEY:}"
      sign_name: "YYHertz"
      
  storage:
    provider: "local" # local, oss, s3
    local:
      path: "./uploads"
    oss:
      endpoint: "${OSS_ENDPOINT:}"
      bucket: "${OSS_BUCKET:}"
      access_key: "${OSS_ACCESS_KEY:}"
      secret_key: "${OSS_SECRET_KEY:}"
```

### 3. 环境特定配置

```yaml
# config/app.development.yaml
app:
  server:
    mode: "debug"
    port: 8888
    
  template:
    cache: false
    reload: true
    
  security:
    csrf:
      enabled: false
      
  database:
    default: "sqlite://./data/development.db"
    
features:
  user_registration: true
  email_verification: false

# config/app.production.yaml
app:
  server:
    mode: "release"
    port: ${PORT:8080}
    tls:
      enabled: true
      cert_file: "/etc/ssl/cert.pem"
      key_file: "/etc/ssl/key.pem"
      
  template:
    cache: true
    reload: false
    
  security:
    csrf:
      enabled: true
    cors:
      allow_origins: ["https://example.com"]
      
  session:
    provider: "redis"
    cookie_secure: true
    
features:
  user_registration: false
  email_verification: true
  two_factor_auth: true
```

## 🔧 配置管理器

### 1. 配置结构定义

```go
// config/config.go
package config

import (
    "time"
)

// AppConfig 应用配置主结构
type AppConfig struct {
    App      *App                   `yaml:"app"`
    Features map[string]bool        `yaml:"features"`
    Services map[string]interface{} `yaml:"services"`
}

// App 应用基础配置
type App struct {
    Name        string     `yaml:"name"`
    Version     string     `yaml:"version"`
    Description string     `yaml:"description"`
    Server      *Server    `yaml:"server"`
    Paths       *Paths     `yaml:"paths"`
    Template    *Template  `yaml:"template"`
    Session     *Session   `yaml:"session"`
    Security    *Security  `yaml:"security"`
    Database    string     `yaml:"database"`
    Cache       string     `yaml:"cache"`
    Logging     *Logging   `yaml:"logging"`
}

// Server 服务器配置
type Server struct {
    Host           string        `yaml:"host"`
    Port           int           `yaml:"port"`
    Mode           string        `yaml:"mode"`
    ReadTimeout    time.Duration `yaml:"read_timeout"`
    WriteTimeout   time.Duration `yaml:"write_timeout"`
    IdleTimeout    time.Duration `yaml:"idle_timeout"`
    MaxHeaderBytes int           `yaml:"max_header_bytes"`
    TLS            *TLS          `yaml:"tls"`
    HTTP2          *HTTP2        `yaml:"http2"`
}

// TLS TLS配置
type TLS struct {
    Enabled  bool   `yaml:"enabled"`
    CertFile string `yaml:"cert_file"`
    KeyFile  string `yaml:"key_file"`
    AutoCert bool   `yaml:"auto_cert"`
}

// HTTP2 HTTP/2配置
type HTTP2 struct {
    Enabled              bool `yaml:"enabled"`
    MaxConcurrentStreams int  `yaml:"max_concurrent_streams"`
}

// Paths 路径配置
type Paths struct {
    Static  string `yaml:"static"`
    Views   string `yaml:"views"`
    Uploads string `yaml:"uploads"`
    Logs    string `yaml:"logs"`
    Temp    string `yaml:"temp"`
}

// Template 模板配置
type Template struct {
    Engine    string            `yaml:"engine"`
    Cache     bool              `yaml:"cache"`
    Reload    bool              `yaml:"reload"`
    Delims    *Delims           `yaml:"delims"`
    Functions []string          `yaml:"functions"`
}

type Delims struct {
    Left  string `yaml:"left"`
    Right string `yaml:"right"`
}

// Session 会话配置
type Session struct {
    Provider       string `yaml:"provider"`
    CookieName     string `yaml:"cookie_name"`
    CookiePath     string `yaml:"cookie_path"`
    CookieDomain   string `yaml:"cookie_domain"`
    CookieSecure   bool   `yaml:"cookie_secure"`
    CookieHTTPOnly bool   `yaml:"cookie_http_only"`
    CookieSameSite string `yaml:"cookie_same_site"`
    MaxAge         int    `yaml:"max_age"`
}

// Security 安全配置
type Security struct {
    CSRF      *CSRF      `yaml:"csrf"`
    CORS      *CORS      `yaml:"cors"`
    RateLimit *RateLimit `yaml:"rate_limit"`
}

type CSRF struct {
    Enabled    bool   `yaml:"enabled"`
    TokenLength int   `yaml:"token_length"`
    CookieName string `yaml:"cookie_name"`
    HeaderName string `yaml:"header_name"`
}

type CORS struct {
    Enabled          bool     `yaml:"enabled"`
    AllowOrigins     []string `yaml:"allow_origins"`
    AllowMethods     []string `yaml:"allow_methods"`
    AllowHeaders     []string `yaml:"allow_headers"`
    AllowCredentials bool     `yaml:"allow_credentials"`
    MaxAge           int      `yaml:"max_age"`
}

type RateLimit struct {
    Enabled           bool `yaml:"enabled"`
    RequestsPerMinute int  `yaml:"requests_per_minute"`
    Burst             int  `yaml:"burst"`
}

// Logging 日志配置
type Logging struct {
    Level  string `yaml:"level"`
    Format string `yaml:"format"`
}
```

### 2. 配置加载器

```go
// config/loader.go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    
    "gopkg.in/yaml.v3"
    "github.com/spf13/viper"
)

// Loader 配置加载器
type Loader struct {
    environment string
    configPaths []string
    viper       *viper.Viper
}

// NewLoader 创建配置加载器
func NewLoader(environment string) *Loader {
    v := viper.New()
    
    // 设置配置文件类型
    v.SetConfigType("yaml")
    
    // 设置环境变量前缀
    v.SetEnvPrefix("APP")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    return &Loader{
        environment: environment,
        configPaths: []string{"./config", "./configs", "/etc/app"},
        viper:       v,
    }
}

// AddConfigPath 添加配置路径
func (l *Loader) AddConfigPath(path string) {
    l.configPaths = append(l.configPaths, path)
    l.viper.AddConfigPath(path)
}

// LoadConfig 加载配置
func (l *Loader) LoadConfig() (*AppConfig, error) {
    // 加载主配置文件
    if err := l.loadMainConfig(); err != nil {
        return nil, err
    }
    
    // 加载环境特定配置
    if err := l.loadEnvironmentConfig(); err != nil {
        return nil, err
    }
    
    // 加载扩展配置
    if err := l.loadExtendedConfigs(); err != nil {
        return nil, err
    }
    
    // 解析配置到结构体
    var config AppConfig
    if err := l.viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    // 验证配置
    if err := l.validateConfig(&config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    return &config, nil
}

// loadMainConfig 加载主配置文件
func (l *Loader) loadMainConfig() error {
    l.viper.SetConfigName("app")
    
    if err := l.viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // 配置文件不存在，使用默认配置
            return l.loadDefaultConfig()
        }
        return fmt.Errorf("failed to read main config: %w", err)
    }
    
    return nil
}

// loadEnvironmentConfig 加载环境特定配置
func (l *Loader) loadEnvironmentConfig() error {
    if l.environment == "" {
        return nil
    }
    
    envConfigName := fmt.Sprintf("app.%s", l.environment)
    
    // 查找环境配置文件
    for _, path := range l.configPaths {
        envConfigFile := filepath.Join(path, envConfigName+".yaml")
        if _, err := os.Stat(envConfigFile); err == nil {
            // 找到环境配置文件，合并配置
            envViper := viper.New()
            envViper.SetConfigType("yaml")
            envViper.SetConfigFile(envConfigFile)
            
            if err := envViper.ReadInConfig(); err != nil {
                return fmt.Errorf("failed to read environment config: %w", err)
            }
            
            // 合并配置
            return l.viper.MergeConfigMap(envViper.AllSettings())
        }
    }
    
    return nil
}

// loadExtendedConfigs 加载扩展配置文件
func (l *Loader) loadExtendedConfigs() error {
    extendedConfigs := []string{
        "database",
        "cache", 
        "middleware",
        "logging",
    }
    
    for _, configName := range extendedConfigs {
        if err := l.loadExtendedConfig(configName); err != nil {
            // 扩展配置加载失败不应该中断主流程
            fmt.Printf("Warning: failed to load %s config: %v\n", configName, err)
        }
    }
    
    return nil
}

func (l *Loader) loadExtendedConfig(configName string) error {
    for _, path := range l.configPaths {
        configFile := filepath.Join(path, configName+".yaml")
        if _, err := os.Stat(configFile); err == nil {
            extViper := viper.New()
            extViper.SetConfigType("yaml")
            extViper.SetConfigFile(configFile)
            
            if err := extViper.ReadInConfig(); err != nil {
                return err
            }
            
            // 将扩展配置合并到主配置中
            extSettings := extViper.AllSettings()
            if len(extSettings) > 0 {
                l.viper.Set(configName, extSettings[configName])
            }
            
            return nil
        }
    }
    
    return nil
}

// loadDefaultConfig 加载默认配置
func (l *Loader) loadDefaultConfig() error {
    defaultConfig := map[string]interface{}{
        "app": map[string]interface{}{
            "name":    "YYHertz MVC App",
            "version": "1.0.0",
            "server": map[string]interface{}{
                "host": "0.0.0.0",
                "port": 8888,
                "mode": "debug",
            },
            "paths": map[string]interface{}{
                "static":  "./static",
                "views":   "./views",
                "uploads": "./uploads",
                "logs":    "./logs",
                "temp":    "./temp",
            },
        },
    }
    
    return l.viper.MergeConfigMap(defaultConfig)
}

// validateConfig 验证配置
func (l *Loader) validateConfig(config *AppConfig) error {
    if config.App == nil {
        return fmt.Errorf("app configuration is required")
    }
    
    if config.App.Server == nil {
        return fmt.Errorf("server configuration is required")
    }
    
    if config.App.Server.Port <= 0 || config.App.Server.Port > 65535 {
        return fmt.Errorf("invalid server port: %d", config.App.Server.Port)
    }
    
    validModes := []string{"debug", "test", "release"}
    if !contains(validModes, config.App.Server.Mode) {
        return fmt.Errorf("invalid server mode: %s", config.App.Server.Mode)
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

### 3. 配置管理器

```go
// config/manager.go
package config

import (
    "sync"
    "fmt"
)

// Manager 配置管理器
type Manager struct {
    config *AppConfig
    loader *Loader
    mutex  sync.RWMutex
    watchers []func(*AppConfig)
}

var (
    defaultManager *Manager
    once           sync.Once
)

// GetManager 获取配置管理器实例
func GetManager() *Manager {
    once.Do(func() {
        defaultManager = &Manager{
            watchers: make([]func(*AppConfig), 0),
        }
    })
    return defaultManager
}

// Initialize 初始化配置管理器
func (m *Manager) Initialize(environment string) error {
    m.loader = NewLoader(environment)
    return m.Reload()
}

// Reload 重新加载配置
func (m *Manager) Reload() error {
    config, err := m.loader.LoadConfig()
    if err != nil {
        return err
    }
    
    m.mutex.Lock()
    oldConfig := m.config
    m.config = config
    m.mutex.Unlock()
    
    // 通知配置变更
    m.notifyWatchers(config)
    
    if oldConfig != nil {
        fmt.Println("Configuration reloaded successfully")
    } else {
        fmt.Println("Configuration loaded successfully")
    }
    
    return nil
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *AppConfig {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.config
}

// GetApp 获取应用配置
func (m *Manager) GetApp() *App {
    config := m.GetConfig()
    if config != nil {
        return config.App
    }
    return nil
}

// GetFeature 获取功能开关
func (m *Manager) GetFeature(name string) bool {
    config := m.GetConfig()
    if config != nil && config.Features != nil {
        return config.Features[name]
    }
    return false
}

// GetService 获取服务配置
func (m *Manager) GetService(name string) interface{} {
    config := m.GetConfig()
    if config != nil && config.Services != nil {
        return config.Services[name]
    }
    return nil
}

// Watch 注册配置变更监听器
func (m *Manager) Watch(watcher func(*AppConfig)) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.watchers = append(m.watchers, watcher)
}

// notifyWatchers 通知配置变更监听器
func (m *Manager) notifyWatchers(config *AppConfig) {
    for _, watcher := range m.watchers {
        go func(w func(*AppConfig)) {
            defer func() {
                if r := recover(); r != nil {
                    fmt.Printf("Config watcher panic: %v\n", r)
                }
            }()
            w(config)
        }(watcher)
    }
}

// 便捷访问函数
func GetConfig() *AppConfig {
    return GetManager().GetConfig()
}

func GetApp() *App {
    return GetManager().GetApp()
}

func GetFeature(name string) bool {
    return GetManager().GetFeature(name)
}

func GetService(name string) interface{} {
    return GetManager().GetService(name)
}
```

## 🔄 配置热更新

### 1. 文件监控

```go
// config/watcher.go
package config

import (
    "log"
    "path/filepath"
    "time"
    
    "github.com/fsnotify/fsnotify"
)

// FileWatcher 配置文件监控器
type FileWatcher struct {
    manager *Manager
    watcher *fsnotify.Watcher
    paths   []string
}

// NewFileWatcher 创建文件监控器
func NewFileWatcher(manager *Manager) *FileWatcher {
    return &FileWatcher{
        manager: manager,
        paths:   []string{"./config"},
    }
}

// Start 启动文件监控
func (fw *FileWatcher) Start() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    fw.watcher = watcher
    
    // 监控配置目录
    for _, path := range fw.paths {
        if err := watcher.Add(path); err != nil {
            log.Printf("Failed to watch config path %s: %v", path, err)
        }
    }
    
    go fw.watchLoop()
    return nil
}

// Stop 停止文件监控
func (fw *FileWatcher) Stop() error {
    if fw.watcher != nil {
        return fw.watcher.Close()
    }
    return nil
}

// watchLoop 监控循环
func (fw *FileWatcher) watchLoop() {
    defer fw.watcher.Close()
    
    for {
        select {
        case event, ok := <-fw.watcher.Events:
            if !ok {
                return
            }
            
            if fw.isConfigFile(event.Name) && event.Op&fsnotify.Write == fsnotify.Write {
                log.Printf("Config file changed: %s", event.Name)
                
                // 延迟重新加载，避免文件写入过程中的部分读取
                time.Sleep(100 * time.Millisecond)
                
                if err := fw.manager.Reload(); err != nil {
                    log.Printf("Failed to reload config: %v", err)
                }
            }
            
        case err, ok := <-fw.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Config file watcher error: %v", err)
        }
    }
}

// isConfigFile 检查是否为配置文件
func (fw *FileWatcher) isConfigFile(filename string) bool {
    ext := filepath.Ext(filename)
    configExts := []string{".yaml", ".yml", ".json", ".toml"}
    
    for _, configExt := range configExts {
        if ext == configExt {
            return true
        }
    }
    return false
}
```

### 2. 远程配置支持

```go
// config/remote.go
package config

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// RemoteConfigClient 远程配置客户端
type RemoteConfigClient struct {
    baseURL    string
    token      string
    httpClient *http.Client
    manager    *Manager
}

// NewRemoteConfigClient 创建远程配置客户端
func NewRemoteConfigClient(baseURL, token string, manager *Manager) *RemoteConfigClient {
    return &RemoteConfigClient{
        baseURL: baseURL,
        token:   token,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
        manager: manager,
    }
}

// StartPolling 启动配置轮询
func (rcc *RemoteConfigClient) StartPolling(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := rcc.fetchAndUpdate(); err != nil {
            log.Printf("Failed to fetch remote config: %v", err)
        }
    }
}

// fetchAndUpdate 获取并更新远程配置
func (rcc *RemoteConfigClient) fetchAndUpdate() error {
    url := fmt.Sprintf("%s/api/config", rcc.baseURL)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+rcc.token)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := rcc.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to fetch config: %d", resp.StatusCode)
    }
    
    var remoteConfig AppConfig
    if err := json.NewDecoder(resp.Body).Decode(&remoteConfig); err != nil {
        return err
    }
    
    // 更新本地配置
    rcc.manager.mutex.Lock()
    rcc.manager.config = &remoteConfig
    rcc.manager.mutex.Unlock()
    
    // 通知配置变更
    rcc.manager.notifyWatchers(&remoteConfig)
    
    return nil
}
```

## 🧪 配置测试

### 1. 配置测试工具

```go
// config/testing.go
package config

import (
    "os"
    "testing"
    "path/filepath"
)

// TestConfig 测试配置
func TestConfig(t *testing.T) *AppConfig {
    // 创建临时配置文件
    configContent := `
app:
  name: "Test App"
  server:
    host: "localhost"
    port: 0
    mode: "test"
  paths:
    static: "./test_static"
    views: "./test_views"
features:
  test_feature: true
`
    
    tempDir := t.TempDir()
    configFile := filepath.Join(tempDir, "app.yaml")
    
    if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
        t.Fatal(err)
    }
    
    // 加载测试配置
    loader := NewLoader("test")
    loader.AddConfigPath(tempDir)
    
    config, err := loader.LoadConfig()
    if err != nil {
        t.Fatal(err)
    }
    
    return config
}

// SetTestConfig 设置测试配置
func SetTestConfig(config *AppConfig) {
    manager := GetManager()
    manager.mutex.Lock()
    manager.config = config
    manager.mutex.Unlock()
}
```

### 2. 配置验证测试

```go
// config/config_test.go
package config

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestConfigLoader(t *testing.T) {
    t.Run("Load Valid Config", func(t *testing.T) {
        config := TestConfig(t)
        
        assert.NotNil(t, config)
        assert.NotNil(t, config.App)
        assert.Equal(t, "Test App", config.App.Name)
        assert.Equal(t, "test", config.App.Server.Mode)
        assert.True(t, config.Features["test_feature"])
    })
    
    t.Run("Environment Override", func(t *testing.T) {
        // 测试环境变量覆盖
        os.Setenv("APP_APP_SERVER_PORT", "9999")
        defer os.Unsetenv("APP_APP_SERVER_PORT")
        
        config := TestConfig(t)
        assert.Equal(t, 9999, config.App.Server.Port)
    })
    
    t.Run("Feature Toggle", func(t *testing.T) {
        config := TestConfig(t)
        SetTestConfig(config)
        
        assert.True(t, GetFeature("test_feature"))
        assert.False(t, GetFeature("non_existent_feature"))
    })
}

func TestConfigValidation(t *testing.T) {
    t.Run("Invalid Port", func(t *testing.T) {
        loader := NewLoader("test")
        config := &AppConfig{
            App: &App{
                Server: &Server{Port: -1},
            },
        }
        
        err := loader.validateConfig(config)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid server port")
    })
    
    t.Run("Invalid Mode", func(t *testing.T) {
        loader := NewLoader("test")
        config := &AppConfig{
            App: &App{
                Server: &Server{
                    Port: 8080,
                    Mode: "invalid",
                },
            },
        }
        
        err := loader.validateConfig(config)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid server mode")
    })
}
```

## 🔗 相关资源

- [环境配置管理](./environment.md)
- [数据库配置](../data-access/database-config.md)
- [中间件配置](../middleware/config.md)
- [应用程序架构](../mvc-core/application.md)

---

> 💡 **提示**: 良好的配置管理是应用维护的基础。建议采用分层配置和环境隔离的方式，确保配置的安全性和可维护性。
