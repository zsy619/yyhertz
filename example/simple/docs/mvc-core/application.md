# 🖥️ 应用程序

YYHertz应用程序是整个框架的核心，负责管理HTTP服务器、路由注册、中间件链和生命周期管理。理解应用程序的工作原理对于构建高质量的Web应用至关重要。

## 🏗️ 应用架构

### 核心组件图
```
┌─────────────────────────────────────────┐
│            YYHertz Application          │
├─────────────────────────────────────────┤
│  🌐 HTTP Server (CloudWeGo-Hertz)       │
│  ├── Router Engine                      │
│  ├── Middleware Chain                   │
│  ├── Controller Registry                │
│  └── Template Engine                    │
├─────────────────────────────────────────┤
│  🔌 Middleware System                   │
│  ├── Global Middleware                  │
│  ├── Group Middleware                   │
│  ├── Route Middleware                   │
│  └── Controller Middleware              │
├─────────────────────────────────────────┤
│  🎛️ MVC Components                      │
│  ├── Controllers                        │
│  ├── Models                             │
│  ├── Views/Templates                    │
│  └── Services                           │
├─────────────────────────────────────────┤
│  ⚙️ Configuration                       │
│  ├── Server Config                      │
│  ├── Database Config                    │
│  ├── Middleware Config                  │
│  └── Template Config                    │
└─────────────────────────────────────────┘
```

## 🚀 应用初始化

### 基础应用创建
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
    // 获取全局应用实例
    app := mvc.HertzApp
    
    // 添加全局中间件
    app.Use(
        middleware.Recovery(),  // 异常恢复
        middleware.Logger(),    // 请求日志
        middleware.CORS(),      // 跨域支持
    )
    
    // 注册控制器
    app.AutoRouters(&HomeController{})
    
    // 启动HTTP服务
    app.Run(":8888")
}
```

### 高级应用配置
```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
    "github.com/zsy619/yyhertz/framework/config"
)

func main() {
    // 创建应用实例
    app := mvc.HertzApp
    
    // 配置服务器参数
    app.Configure(mvc.AppConfig{
        Name:         "YYHertz API",
        Version:      "2.0.0",
        Mode:         "release", // debug, release, test
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    })
    
    // 设置中间件
    setupMiddleware(app)
    
    // 注册路由
    setupRoutes(app)
    
    // 优雅关闭
    gracefulShutdown(app)
}

func setupMiddleware(app *mvc.Application) {
    app.Use(
        middleware.Recovery(),
        middleware.Logger(),
        middleware.CORS(),
        middleware.RateLimit(100, time.Minute),
    )
}

func setupRoutes(app *mvc.Application) {
    // API路由组
    api := app.Group("/api")
    api.Use(middleware.Auth(middleware.AuthConfig{
        Strategy: middleware.AuthJWT,
    }))
    
    // 注册API控制器
    api.AutoRouters(&APIController{})
    
    // Web路由组
    web := app.Group("/")
    web.AutoRouters(&WebController{})
}

func gracefulShutdown(app *mvc.Application) {
    // 创建接收系统信号的channel
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    // 启动服务器
    go func() {
        if err := app.Run(":8888"); err != nil {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()
    
    log.Println("Server started on :8888")
    
    // 等待关闭信号
    <-quit
    log.Println("Shutting down server...")
    
    // 创建关闭超时context
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // 优雅关闭服务器
    if err := app.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited")
}
```

## ⚙️ 应用配置

### 配置结构体
```go
// framework/mvc/app.go
type AppConfig struct {
    Name            string        // 应用名称
    Version         string        // 版本号
    Mode            string        // 运行模式: debug, release, test
    Host            string        // 监听地址
    Port            int           // 监听端口
    ReadTimeout     time.Duration // 读取超时
    WriteTimeout    time.Duration // 写入超时
    IdleTimeout     time.Duration // 空闲超时
    MaxHeaderBytes  int           // 最大请求头字节数
    TLSConfig       *TLSConfig    // TLS配置
    TemplateConfig  *TemplateConfig // 模板配置
    StaticConfig    *StaticConfig   // 静态文件配置
}

type TLSConfig struct {
    Enabled  bool   // 启用TLS
    CertFile string // 证书文件路径
    KeyFile  string // 私钥文件路径
}

type TemplateConfig struct {
    Dir      string   // 模板目录
    Suffix   string   // 模板后缀
    Funcs    template.FuncMap // 自定义函数
    Reload   bool     // 开发模式重载
}

type StaticConfig struct {
    Dir    string // 静态文件目录
    Prefix string // URL前缀
    MaxAge int    // 缓存时间(秒)
}

// 静态路径映射配置
type StaticPathConfig struct {
    Mappings map[string]string // URL路径 -> 本地目录映射
    Default  string            // 默认静态目录
}
```

### YAML配置文件
```yaml
# config/app.yaml
app:
  name: "YYHertz Application"
  version: "2.0.0"
  mode: "debug"
  host: "0.0.0.0"
  port: 8888

server:
  read_timeout: "30s"
  write_timeout: "30s" 
  idle_timeout: "60s"
  max_header_bytes: 1048576  # 1MB

tls:
  enabled: false
  cert_file: "./certs/server.crt"
  key_file: "./certs/server.key"

template:
  dir: "./views"
  suffix: ".html"
  reload: true  # 开发模式

static:
  dir: "./static"
  prefix: "/static"
  max_age: 3600  # 1 hour
  
# 多静态路径配置
static_paths:
  "/static": "./public"        # 主要静态资源
  "/files": "./uploads"        # 用户上传文件
  "/assets": "./assets"        # 编译后的资源
  "/docs": "./documents"       # 文档资源
```

### 从配置文件加载
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/config"
)

func main() {
    // 加载配置文件
    cfg, err := config.LoadFromFile("./config/app.yaml")
    if err != nil {
        panic(err)
    }
    
    app := mvc.HertzApp
    
    // 应用配置
    app.Configure(mvc.AppConfig{
        Name:         cfg.App.Name,
        Version:      cfg.App.Version,
        Mode:         cfg.App.Mode,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
    })
    
    // 配置模板
    app.SetTemplateConfig(mvc.TemplateConfig{
        Dir:    cfg.Template.Dir,
        Suffix: cfg.Template.Suffix,
        Reload: cfg.Template.Reload,
    })
    
    // 配置静态文件
    app.SetStaticConfig(mvc.StaticConfig{
        Dir:    cfg.Static.Dir,
        Prefix: cfg.Static.Prefix,
        MaxAge: cfg.Static.MaxAge,
    })
    
    // 配置多静态路径（推荐方式）
    for urlPath, localDir := range cfg.StaticPaths {
        mvc.SetStaticPath(localDir, urlPath)
    }
    
    // 或者使用简化方式
    mvc.SetStaticPath("public", "/static")   // 主要资源
    mvc.SetStaticPath("uploads", "/files")   // 上传文件
    mvc.SetStaticPath("assets")              // 自动推导为 /assets
    
    app.Run(fmt.Sprintf(":%d", cfg.App.Port))
}
```

## 🔄 应用生命周期

### 生命周期钩子
```go
package main

import (
    "context"
    "log"
    "github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
    app := mvc.HertzApp
    
    // 应用启动前钩子
    app.OnBeforeStart(func(ctx context.Context) error {
        log.Println("应用启动前初始化...")
        
        // 初始化数据库
        if err := initDatabase(); err != nil {
            return err
        }
        
        // 初始化缓存
        if err := initCache(); err != nil {
            return err
        }
        
        // 初始化任务调度器
        if err := initScheduler(); err != nil {
            return err
        }
        
        log.Println("初始化完成")
        return nil
    })
    
    // 应用启动后钩子
    app.OnAfterStart(func(ctx context.Context) error {
        log.Println("应用已启动，执行启动后任务...")
        
        // 预热缓存
        if err := warmupCache(); err != nil {
            log.Printf("缓存预热失败: %v", err)
        }
        
        // 注册服务发现
        if err := registerService(); err != nil {
            log.Printf("服务注册失败: %v", err)
        }
        
        return nil
    })
    
    // 应用关闭前钩子
    app.OnBeforeStop(func(ctx context.Context) error {
        log.Println("应用关闭前清理...")
        
        // 注销服务发现
        unregisterService()
        
        // 停止任务调度器
        stopScheduler()
        
        // 关闭数据库连接
        closeDatabaseConnections()
        
        log.Println("清理完成")
        return nil
    })
    
    // 应用关闭后钩子
    app.OnAfterStop(func(ctx context.Context) error {
        log.Println("应用已关闭")
        return nil
    })
    
    // 启动应用
    app.Run(":8888")
}

func initDatabase() error {
    // 数据库初始化逻辑
    log.Println("初始化数据库连接...")
    return nil
}

func initCache() error {
    // 缓存初始化逻辑
    log.Println("初始化Redis连接...")
    return nil
}

func initScheduler() error {
    // 调度器初始化逻辑
    log.Println("初始化任务调度器...")
    return nil
}
```

### 健康检查
```go
package main

import (
    "net/http"
    "github.com/zsy619/yyhertz/framework/mvc"
)

type HealthController struct {
    mvc.BaseController
}

// GET /health
func (c *HealthController) GetIndex() {
    // 检查数据库连接
    dbStatus := checkDatabase()
    
    // 检查Redis连接
    redisStatus := checkRedis()
    
    // 检查外部服务
    servicesStatus := checkExternalServices()
    
    // 整体健康状态
    overallStatus := "healthy"
    if !dbStatus || !redisStatus || !servicesStatus {
        overallStatus = "unhealthy"
        c.SetStatusCode(http.StatusServiceUnavailable)
    }
    
    c.JSON(map[string]any{
        "status": overallStatus,
        "timestamp": time.Now().Unix(),
        "checks": map[string]any{
            "database": map[string]any{
                "status": getStatusString(dbStatus),
                "response_time": "2ms",
            },
            "redis": map[string]any{
                "status": getStatusString(redisStatus),
                "response_time": "1ms",
            },
            "external_services": map[string]any{
                "status": getStatusString(servicesStatus),
                "response_time": "15ms",
            },
        },
        "version": "2.0.0",
        "uptime": getUptime(),
    })
}

func main() {
    app := mvc.HertzApp
    
    // 注册健康检查路由
    app.AutoRouters(&HealthController{})
    
    // 添加就绪检查
    app.GET("/ready", func(c *mvc.Context) {
        if isApplicationReady() {
            c.JSON(map[string]string{"status": "ready"})
        } else {
            c.Status(http.StatusServiceUnavailable)
            c.JSON(map[string]string{"status": "not ready"})
        }
    })
    
    app.Run(":8888")
}
```

## 📊 应用监控

### 内置指标收集
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
    "github.com/zsy619/yyhertz/framework/metrics"
)

func main() {
    app := mvc.HertzApp
    
    // 启用指标收集
    app.EnableMetrics(metrics.Config{
        Enabled: true,
        Path:    "/metrics", // Prometheus指标路径
        Namespace: "yyhertz",
        Subsystem: "app",
    })
    
    // 添加指标中间件
    app.Use(middleware.Metrics())
    
    // 自定义指标
    requestCounter := metrics.NewCounter(metrics.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    })
    
    requestDuration := metrics.NewHistogram(metrics.HistogramOpts{
        Name: "http_request_duration_seconds", 
        Help: "HTTP request latencies in seconds",
    })
    
    app.Use(func(c *mvc.Context) {
        start := time.Now()
        
        // 增加请求计数
        requestCounter.Inc()
        
        // 处理请求
        c.Next()
        
        // 记录请求耗时
        duration := time.Since(start).Seconds()
        requestDuration.Observe(duration)
    })
    
    app.Run(":8888")
}
```

### 日志配置
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/log"
)

func main() {
    app := mvc.HertzApp
    
    // 配置结构化日志
    logger := log.New(log.Config{
        Level:      "info",
        Format:     "json",      // json, text
        Output:     "file",      // stdout, file, both
        FilePath:   "./logs/app.log",
        MaxSize:    100,         // MB
        MaxAge:     30,          // days
        MaxBackups: 5,
        Compress:   true,
    })
    
    // 设置全局logger
    app.SetLogger(logger)
    
    // 应用级别日志
    app.Logger().Info("应用启动", log.Fields{
        "version": "2.0.0",
        "port":    8888,
    })
    
    app.Run(":8888")
}
```

## 🔧 高级特性

### 多应用实例
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
    // 创建API应用
    apiApp := mvc.New("api-app")
    apiApp.Use(middleware.CORS(), middleware.Auth())
    apiApp.AutoRouters(&APIController{})
    
    // 创建Web应用
    webApp := mvc.New("web-app")
    webApp.Use(middleware.Session(), middleware.CSRF())
    webApp.AutoRouters(&WebController{})
    
    // 创建管理后台应用
    adminApp := mvc.New("admin-app")
    adminApp.Use(middleware.BasicAuth(map[string]string{
        "admin": "secret",
    }))
    adminApp.AutoRouters(&AdminController{})
    
    // 并发启动多个应用
    go apiApp.Run(":8080")   // API服务
    go webApp.Run(":8081")   // Web服务
    adminApp.Run(":8082")    // 管理后台
}
```

### 应用插件系统
```go
// 定义插件接口
type Plugin interface {
    Name() string
    Version() string
    Init(app *mvc.Application) error
    Start() error
    Stop() error
}

// 示例插件实现
type LogPlugin struct {
    logger log.Logger
}

func (p *LogPlugin) Name() string { return "log-plugin" }
func (p *LogPlugin) Version() string { return "1.0.0" }

func (p *LogPlugin) Init(app *mvc.Application) error {
    p.logger = app.Logger()
    app.Use(middleware.Logger())
    return nil
}

func (p *LogPlugin) Start() error {
    p.logger.Info("Log plugin started")
    return nil
}

func (p *LogPlugin) Stop() error {
    p.logger.Info("Log plugin stopped")
    return nil
}

func main() {
    app := mvc.HertzApp
    
    // 注册插件
    app.RegisterPlugin(&LogPlugin{})
    app.RegisterPlugin(&MetricsPlugin{})
    app.RegisterPlugin(&AuthPlugin{})
    
    app.Run(":8888")
}
```

## 🚀 最佳实践

### 1. 应用分层
```
Application Layer    (HTTP处理、路由、中间件)
    ↓
Business Layer       (业务逻辑、服务层)
    ↓  
Data Access Layer    (数据访问、仓储层)
    ↓
Infrastructure Layer (数据库、缓存、消息队列)
```

### 2. 配置管理
- 使用环境变量覆盖配置文件
- 敏感信息不要写入代码
- 支持多环境配置 (dev/staging/prod)
- 配置验证和默认值

### 3. 错误处理
```go
// 统一错误处理中间件
func ErrorHandler() mvc.HandlerFunc {
    return mvc.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *mvc.Context, recovered interface{}) {
        if err, ok := recovered.(string); ok {
            c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
        }
        c.AbortWithStatus(http.StatusInternalServerError)
    })
}
```

### 4. 性能优化
- 合理设置超时时间
- 使用连接池
- 启用响应压缩
- 静态资源缓存
- 数据库查询优化

## 📖 下一步

现在您已经掌握了YYHertz应用程序的核心概念，建议继续学习：

1. 🎛️ [控制器基础](/home/controller) - 掌握请求处理
2. 🌐 [路由系统](/home/routing) - 了解URL映射
3. 🔌 [中间件系统](/home/middleware-overview) - 理解请求处理链
4. 📁 [命名空间](/home/namespace) - 组织复杂路由结构

---

**💡 应用程序是框架的心脏，理解它的工作机制将让您的开发事半功倍！**