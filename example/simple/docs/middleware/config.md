# ⚙️ 中间件配置

中间件配置是定制化中间件行为的关键。通过合理的配置，可以让中间件适应不同的业务场景和环境需求。

## 🎯 配置设计原则

### 1. 配置结构设计

```go
// 基础配置结构
type BaseConfig struct {
    // 是否启用
    Enabled bool
    // 调试模式
    Debug bool
    // 跳过的路径
    SkipPaths []string
    // 环境类型
    Environment string
}

// 具体中间件配置继承基础配置
type LoggerConfig struct {
    BaseConfig
    // 日志格式
    Format string
    // 输出文件
    Output string
    // 是否启用颜色
    EnableColor bool
    // 日志级别
    Level LogLevel
}
```

### 2. 配置验证

```go
type Validator interface {
    Validate() error
}

func (c *LoggerConfig) Validate() error {
    if c.Format == "" {
        return errors.New("logger format cannot be empty")
    }
    if c.Level < 0 || c.Level > 5 {
        return errors.New("invalid log level")
    }
    return nil
}
```

### 3. 默认配置

```go
func DefaultLoggerConfig() *LoggerConfig {
    return &LoggerConfig{
        BaseConfig: BaseConfig{
            Enabled:     true,
            Debug:       false,
            Environment: "development",
        },
        Format:      "${time} | ${status} | ${latency} | ${method} ${path}",
        EnableColor: true,
        Level:       InfoLevel,
    }
}
```

## 📁 配置文件管理

### 1. YAML配置文件

```yaml
# config/middleware.yaml
middleware:
  logger:
    enabled: true
    debug: false
    format: "${time} | ${status} | ${latency} | ${ip} | ${method} ${path}"
    output: "logs/access.log"
    enable_color: false
    level: "info"
    skip_paths:
      - "/health"
      - "/metrics"
      - "/static/*"
  
  cors:
    enabled: true
    allow_origins:
      - "https://example.com"
      - "https://app.example.com"
    allow_methods:
      - "GET"
      - "POST" 
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allow_headers:
      - "Origin"
      - "Content-Type"
      - "Accept"
      - "Authorization"
    allow_credentials: true
    max_age: "12h"
  
  rate_limit:
    enabled: true
    window: 60 # seconds
    limit: 100
    key_func: "ip" # ip, user, custom
    storage: "memory" # memory, redis
    redis:
      addr: "localhost:6379"
      password: ""
      db: 0
  
  security:
    enabled: true
    frame_options: "SAMEORIGIN"
    content_type_nosniff: true
    xss_protection: "1; mode=block"
    hsts_max_age: 31536000
    hsts_include_subdomains: true
    csp: "default-src 'self'; script-src 'self' 'unsafe-inline';"
  
  cache:
    enabled: true
    ttl: "5m"
    max_size: 1000
    storage: "memory"
    redis:
      addr: "localhost:6379"
      password: ""
      db: 1
```

### 2. 配置加载器

```go
package config

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

type MiddlewareConfig struct {
    Logger    LoggerConfig    `yaml:"logger"`
    CORS      CORSConfig      `yaml:"cors"`
    RateLimit RateLimitConfig `yaml:"rate_limit"`
    Security  SecurityConfig  `yaml:"security"`
    Cache     CacheConfig     `yaml:"cache"`
}

type ConfigLoader struct {
    configPath string
    config     *MiddlewareConfig
}

func NewConfigLoader(configPath string) *ConfigLoader {
    return &ConfigLoader{
        configPath: configPath,
    }
}

func (cl *ConfigLoader) Load() (*MiddlewareConfig, error) {
    // 读取配置文件
    data, err := os.ReadFile(cl.configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    // 解析YAML
    var config MiddlewareConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    // 应用环境变量覆盖
    cl.applyEnvOverrides(&config)
    
    // 验证配置
    if err := cl.validateConfig(&config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    cl.config = &config
    return &config, nil
}

func (cl *ConfigLoader) applyEnvOverrides(config *MiddlewareConfig) {
    // 环境变量覆盖配置
    if env := os.Getenv("LOGGER_LEVEL"); env != "" {
        config.Logger.Level = env
    }
    
    if env := os.Getenv("REDIS_ADDR"); env != "" {
        config.RateLimit.Redis.Addr = env
        config.Cache.Redis.Addr = env
    }
    
    if env := os.Getenv("CORS_ALLOW_ORIGINS"); env != "" {
        config.CORS.AllowOrigins = strings.Split(env, ",")
    }
}

func (cl *ConfigLoader) validateConfig(config *MiddlewareConfig) error {
    // 验证各个中间件配置
    if err := config.Logger.Validate(); err != nil {
        return fmt.Errorf("logger config: %w", err)
    }
    
    if err := config.CORS.Validate(); err != nil {
        return fmt.Errorf("cors config: %w", err)
    }
    
    if err := config.RateLimit.Validate(); err != nil {
        return fmt.Errorf("rate limit config: %w", err)
    }
    
    return nil
}
```

## 🔧 环境特定配置

### 1. 多环境配置

```go
type EnvironmentConfig struct {
    Development MiddlewareConfig `yaml:"development"`
    Testing     MiddlewareConfig `yaml:"testing"`
    Staging     MiddlewareConfig `yaml:"staging"`
    Production  MiddlewareConfig `yaml:"production"`
}

func LoadEnvironmentConfig(env string) (*MiddlewareConfig, error) {
    configFile := fmt.Sprintf("config/middleware.%s.yaml", env)
    
    // 如果环境特定配置不存在，使用默认配置
    if _, err := os.Stat(configFile); os.IsNotExist(err) {
        configFile = "config/middleware.yaml"
    }
    
    loader := NewConfigLoader(configFile)
    return loader.Load()
}
```

### 2. 配置合并

```go
func MergeConfigs(base, override *MiddlewareConfig) *MiddlewareConfig {
    merged := *base
    
    // 合并Logger配置
    if override.Logger.Format != "" {
        merged.Logger.Format = override.Logger.Format
    }
    if override.Logger.Output != "" {
        merged.Logger.Output = override.Logger.Output
    }
    
    // 合并CORS配置
    if len(override.CORS.AllowOrigins) > 0 {
        merged.CORS.AllowOrigins = override.CORS.AllowOrigins
    }
    
    return &merged
}
```

## 🎛️ 动态配置

### 1. 配置热重载

```go
import (
    "github.com/fsnotify/fsnotify"
)

type ConfigWatcher struct {
    configPath string
    watcher    *fsnotify.Watcher
    onChange   func(*MiddlewareConfig)
    loader     *ConfigLoader
}

func NewConfigWatcher(configPath string, onChange func(*MiddlewareConfig)) *ConfigWatcher {
    return &ConfigWatcher{
        configPath: configPath,
        onChange:   onChange,
        loader:     NewConfigLoader(configPath),
    }
}

func (cw *ConfigWatcher) Start() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    cw.watcher = watcher
    
    // 监听配置文件
    err = watcher.Add(cw.configPath)
    if err != nil {
        return err
    }
    
    go cw.watchLoop()
    return nil
}

func (cw *ConfigWatcher) watchLoop() {
    for {
        select {
        case event, ok := <-cw.watcher.Events:
            if !ok {
                return
            }
            
            if event.Op&fsnotify.Write == fsnotify.Write {
                log.Printf("Config file modified: %s", event.Name)
                
                // 重新加载配置
                config, err := cw.loader.Load()
                if err != nil {
                    log.Printf("Failed to reload config: %v", err)
                    continue
                }
                
                // 调用回调函数
                if cw.onChange != nil {
                    cw.onChange(config)
                }
            }
            
        case err, ok := <-cw.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Config watcher error: %v", err)
        }
    }
}

func (cw *ConfigWatcher) Stop() error {
    if cw.watcher != nil {
        return cw.watcher.Close()
    }
    return nil
}
```

### 2. 远程配置

```go
import (
    "encoding/json"
    "net/http"
    "time"
)

type RemoteConfigClient struct {
    baseURL    string
    httpClient *http.Client
    apiKey     string
}

func NewRemoteConfigClient(baseURL, apiKey string) *RemoteConfigClient {
    return &RemoteConfigClient{
        baseURL: baseURL,
        apiKey:  apiKey,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (rcc *RemoteConfigClient) FetchConfig(service string) (*MiddlewareConfig, error) {
    url := fmt.Sprintf("%s/api/config/%s", rcc.baseURL, service)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+rcc.apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := rcc.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch config: %d", resp.StatusCode)
    }
    
    var config MiddlewareConfig
    if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

// 定期拉取配置
func (rcc *RemoteConfigClient) StartPolling(service string, interval time.Duration, onChange func(*MiddlewareConfig)) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        config, err := rcc.FetchConfig(service)
        if err != nil {
            log.Printf("Failed to fetch remote config: %v", err)
            continue
        }
        
        if onChange != nil {
            onChange(config)
        }
    }
}
```

## 🔄 配置应用

### 1. 中间件管理器

```go
type MiddlewareManager struct {
    config     *MiddlewareConfig
    app        *app.Engine
    middleware map[string]app.HandlerFunc
    mutex      sync.RWMutex
}

func NewMiddlewareManager(app *app.Engine) *MiddlewareManager {
    return &MiddlewareManager{
        app:        app,
        middleware: make(map[string]app.HandlerFunc),
    }
}

func (mm *MiddlewareManager) LoadConfig(config *MiddlewareConfig) error {
    mm.mutex.Lock()
    defer mm.mutex.Unlock()
    
    mm.config = config
    return mm.applyConfig()
}

func (mm *MiddlewareManager) applyConfig() error {
    // 清除现有中间件
    mm.clearMiddleware()
    
    // 应用Logger中间件
    if mm.config.Logger.Enabled {
        logger := middleware.LoggerMiddleware(mm.config.Logger)
        mm.app.Use(logger)
        mm.middleware["logger"] = logger
    }
    
    // 应用CORS中间件
    if mm.config.CORS.Enabled {
        cors := middleware.CORSMiddleware(mm.config.CORS)
        mm.app.Use(cors)
        mm.middleware["cors"] = cors
    }
    
    // 应用限流中间件
    if mm.config.RateLimit.Enabled {
        rateLimit := middleware.RateLimitMiddleware(mm.config.RateLimit)
        mm.app.Use(rateLimit)
        mm.middleware["rate_limit"] = rateLimit
    }
    
    return nil
}

func (mm *MiddlewareManager) clearMiddleware() {
    // 注意：Hertz引擎不支持动态移除中间件
    // 这里需要重新创建引擎实例或使用其他策略
    mm.middleware = make(map[string]app.HandlerFunc)
}

func (mm *MiddlewareManager) ReloadConfig(config *MiddlewareConfig) error {
    return mm.LoadConfig(config)
}
```

### 2. 配置更新通知

```go
type ConfigUpdateNotifier struct {
    subscribers []func(*MiddlewareConfig)
    mutex       sync.RWMutex
}

func NewConfigUpdateNotifier() *ConfigUpdateNotifier {
    return &ConfigUpdateNotifier{
        subscribers: make([]func(*MiddlewareConfig), 0),
    }
}

func (cun *ConfigUpdateNotifier) Subscribe(callback func(*MiddlewareConfig)) {
    cun.mutex.Lock()
    defer cun.mutex.Unlock()
    cun.subscribers = append(cun.subscribers, callback)
}

func (cun *ConfigUpdateNotifier) Notify(config *MiddlewareConfig) {
    cun.mutex.RLock()
    defer cun.mutex.RUnlock()
    
    for _, callback := range cun.subscribers {
        go func(cb func(*MiddlewareConfig)) {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("Config update callback panic: %v", r)
                }
            }()
            cb(config)
        }(callback)
    }
}
```

## 📊 配置监控

### 1. 配置指标收集

```go
type ConfigMetrics struct {
    ReloadCount     int64
    LastReloadTime  time.Time
    ReloadErrors    int64
    ActiveConfig    string
    ConfigVersion   string
}

func (cm *ConfigMetrics) RecordReload(success bool, version string) {
    atomic.AddInt64(&cm.ReloadCount, 1)
    cm.LastReloadTime = time.Now()
    cm.ConfigVersion = version
    
    if !success {
        atomic.AddInt64(&cm.ReloadErrors, 1)
    }
}

func (cm *ConfigMetrics) GetMetrics() map[string]interface{} {
    return map[string]interface{}{
        "reload_count":      atomic.LoadInt64(&cm.ReloadCount),
        "last_reload_time":  cm.LastReloadTime,
        "reload_errors":     atomic.LoadInt64(&cm.ReloadErrors),
        "active_config":     cm.ActiveConfig,
        "config_version":    cm.ConfigVersion,
    }
}
```

### 2. 配置审计

```go
type ConfigAudit struct {
    Timestamp time.Time `json:"timestamp"`
    Action    string    `json:"action"`
    User      string    `json:"user"`
    Source    string    `json:"source"`
    Changes   []Change  `json:"changes"`
}

type Change struct {
    Field    string      `json:"field"`
    OldValue interface{} `json:"old_value"`
    NewValue interface{} `json:"new_value"`
}

func AuditConfigChange(old, new *MiddlewareConfig, user, source string) {
    changes := detectChanges(old, new)
    if len(changes) == 0 {
        return
    }
    
    audit := ConfigAudit{
        Timestamp: time.Now(),
        Action:    "config_update",
        User:      user,
        Source:    source,
        Changes:   changes,
    }
    
    // 记录审计日志
    auditJSON, _ := json.Marshal(audit)
    log.Printf("CONFIG_AUDIT: %s", string(auditJSON))
    
    // 发送到审计系统
    sendToAuditSystem(audit)
}

func detectChanges(old, new *MiddlewareConfig) []Change {
    var changes []Change
    
    // 检测Logger配置变化
    if old.Logger.Level != new.Logger.Level {
        changes = append(changes, Change{
            Field:    "logger.level",
            OldValue: old.Logger.Level,
            NewValue: new.Logger.Level,
        })
    }
    
    // 检测其他配置变化...
    
    return changes
}
```

## 🧪 配置测试

### 1. 配置验证测试

```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  MiddlewareConfig
        wantErr bool
    }{
        {
            name: "valid config",
            config: MiddlewareConfig{
                Logger: LoggerConfig{
                    Format: "${time} | ${status}",
                    Level:  "info",
                },
            },
            wantErr: false,
        },
        {
            name: "invalid log level",
            config: MiddlewareConfig{
                Logger: LoggerConfig{
                    Format: "${time} | ${status}",
                    Level:  "invalid",
                },
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 2. 配置加载测试

```go
func TestConfigLoader(t *testing.T) {
    // 创建临时配置文件
    tempFile, err := os.CreateTemp("", "test_config_*.yaml")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tempFile.Name())
    
    configYAML := `
middleware:
  logger:
    enabled: true
    format: "${time} | ${status}"
    level: "info"
`
    
    _, err = tempFile.WriteString(configYAML)
    if err != nil {
        t.Fatal(err)
    }
    tempFile.Close()
    
    // 测试配置加载
    loader := NewConfigLoader(tempFile.Name())
    config, err := loader.Load()
    
    assert.NoError(t, err)
    assert.True(t, config.Logger.Enabled)
    assert.Equal(t, "info", config.Logger.Level)
}
```

## 📚 最佳实践

### 1. 配置设计原则

- **向后兼容**: 新版本配置应兼容旧版本
- **安全默认**: 提供安全的默认配置
- **环境感知**: 区分开发、测试、生产环境
- **验证完整**: 提供全面的配置验证

### 2. 配置管理建议

- **版本控制**: 配置文件纳入版本控制
- **敏感信息**: 使用环境变量或密钥管理系统
- **文档维护**: 保持配置文档与代码同步
- **监控告警**: 监控配置变更和错误

### 3. 部署建议

```go
// 生产环境配置检查
func ValidateProductionConfig(config *MiddlewareConfig) error {
    // 检查日志配置
    if config.Logger.EnableColor {
        return errors.New("production should not use colored logs")
    }
    
    // 检查CORS配置
    for _, origin := range config.CORS.AllowOrigins {
        if origin == "*" {
            return errors.New("production should not allow all origins")
        }
    }
    
    // 检查安全配置
    if !config.Security.Enabled {
        return errors.New("security middleware must be enabled in production")
    }
    
    return nil
}
```

## 🔗 相关资源

- [中间件概览](./overview.md)
- [内置中间件](./builtin.md)
- [自定义中间件](./custom.md)
- [配置管理指南](../configuration/app-config.md)

---

> 💡 **提示**: 合理的配置管理是中间件系统稳定运行的基础。建议使用配置中心统一管理多环境配置。
