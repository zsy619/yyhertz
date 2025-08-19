# YYHertz 日志系统完整指南

## 📋 概述

YYHertz框架提供了一个功能强大、高度可配置的日志系统，完全集成了[hertz-contrib/logger/logrus](https://github.com/hertz-contrib/logger/tree/main/logrus)，支持多种日志格式、输出目标和流行Go框架的专用格式。基于logrus构建，提供了丰富的格式化器和配置选项，满足从开发调试到生产监控的各种需求。

## 🎯 核心特性

- **多格式支持**: 支持17种日志格式，覆盖主流Go框架
- **灵活配置**: 支持控制台、文件、远程等多种输出方式
- **结构化日志**: 完整的字段支持和上下文追踪
- **高性能**: 针对高并发场景优化的异步输出
- **企业级**: 支持日志聚合、监控和分析平台集成
- **开箱即用**: 提供多种预设配置，快速上手

## 支持的日志格式

### 基础格式
- `json` - 标准JSON格式
- `text` - Logrus文本格式
- `beego` - Beego框架风格格式
- `log4go` - Log4go风格格式

### 企业级格式
- `logstash` - Logstash JSON格式，适用于ELK栈
- `syslog` - 标准Syslog格式
- `fluentd` - Fluentd JSON格式，适用于日志聚合
- `cloudwatch` - AWS CloudWatch格式
- `azure_insights` - Azure Application Insights格式

### 流行框架格式
- `gin` - Gin框架专用格式
- `iris` - Iris框架专用格式
- `ent` - Ent ORM专用格式
- `go_zero` - go-zero微服务格式
- `fiber` - Fiber框架专用格式
- `echo` - Echo框架专用格式
- `revel` - Revel MVC框架格式
- `buffalo` - Buffalo框架专用格式

## 格式详解

### Gin格式 (`gin`)
```
[GIN] [I] 2006/01/02 - 15:04:05 | 200 | 13ms | 127.0.0.1 | GET | /api/users HTTP请求处理完成
```
- **特点**: 支持彩色输出，显示级别标识、状态码、延迟、客户端IP
- **适用场景**: Gin Web应用开发
- **配置选项**:
  - `TimestampFormat`: 时间格式（默认："2006/01/02 - 15:04:05"）
  - `ShowColors`: 是否显示颜色（默认：true）

### Iris格式 (`iris`)
```
[IRIS] 2006/01/02 15:04:05 | 200 | 13ms | 127.0.0.1 | GET /api/users
```
- **特点**: 简洁清晰的格式
- **适用场景**: Iris Web应用开发
- **配置选项**:
  - `TimestampFormat`: 时间格式
  - `ShowColors`: 是否显示颜色（默认：false）

### Ent格式 (`ent`)
```
[ENT] 2006/01/02 15:04:05 [INFO] SELECT * FROM users WHERE id = ? [1] (13ms)
```
- **特点**: 专注于数据库操作，显示SQL语句和执行时间
- **适用场景**: 使用Ent ORM的数据库操作日志
- **配置选项**:
  - `ShowSQL`: 是否显示SQL语句（默认：true）
  - `TimestampFormat`: 时间格式

### go-zero格式 (`go_zero`)
```json
{
  "@timestamp": "2006-01-02T15:04:05Z",
  "level": "info",
  "service": "user-service",
  "trace": "abc123",
  "span": "def456",
  "message": "处理用户请求",
  "caller": "handler.go:42"
}
```
- **特点**: JSON格式，包含微服务相关字段
- **适用场景**: 微服务架构，分布式追踪
- **配置选项**:
  - `ServiceName`: 服务名称
  - `Environment`: 环境标识

### Fiber格式 (`fiber`)
```
15:04:05 | 200 | 13ms | 127.0.0.1 | GET | /api/users
```
- **特点**: 简洁高效，适合高性能场景
- **适用场景**: Fiber高性能Web应用
- **配置选项**:
  - `TimestampFormat`: 时间格式（默认："15:04:05"）

### Echo格式 (`echo`)
```json
{
  "time": "2006-01-02T15:04:05Z",
  "level": "INFO",
  "prefix": "echo",
  "file": "main.go:42",
  "message": "服务器启动"
}
```
- **特点**: JSON格式，包含文件位置信息
- **适用场景**: Echo框架Web应用
- **配置选项**:
  - `Prefix`: 日志前缀（默认："echo"）

### Revel格式 (`revel`)
```
INFO 2006/01/02 15:04:05 controller.go:123: 请求处理完成 [user_id=1001 action=login]
```
- **特点**: MVC框架风格，强调控制器信息
- **适用场景**: Revel MVC应用
- **配置选项**:
  - `ShowCaller`: 是否显示调用位置（默认：true）
  - `TimestampFormat`: 时间格式

### Buffalo格式 (`buffalo`)
```
--> GET /api/users (13ms) | 127.0.0.1 | 200 OK
```
- **特点**: Rails风格的清晰请求响应格式
- **适用场景**: Buffalo Web应用
- **配置选项**:
  - `TimestampFormat`: 时间格式

## 🚀 快速开始

### 1. 基础使用 (推荐新手)

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/config"
    "github.com/zsy619/yyhertz/framework/controller"
)

func main() {
    // 方式1: 使用全局配置函数
    config.InitGlobalLogger(config.DefaultLogConfig())
    config.Info("应用启动成功")
    
    // 方式2: 使用应用级配置 (推荐)
    app := controller.NewApp()
    app.LogInfo("应用启动成功")
    app.Run(":8080")
}
```

### 2. 应用级日志配置

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/config"
    "github.com/zsy619/yyhertz/framework/controller"
)

func main() {
    // 使用预设配置
    app := controller.NewAppWithLogConfig(config.GinLogConfig())
    
    // 记录不同级别的日志
    app.LogDebug("调试信息: %s", "debug message")
    app.LogInfo("信息日志: %s", "info message") 
    app.LogWarn("警告日志: %s", "warning message")
    app.LogError("错误日志: %s", "error message")
    
    app.Run(":8080")
}
```

### 3. 控制器中使用日志

```go
type UserController struct {
    controller.BaseController
}

func (c *UserController) GetIndex() {
    // 基础日志
    c.LogInfo("获取用户列表")
    c.LogRequest() // 自动记录请求详情
    
    // 结构化日志
    c.LogWithFields("info", "用户操作", map[string]any{
        "user_id": 123,
        "action":  "get_list",
        "ip":      c.GetClientIP(),
    })
    
    // 记录响应
    c.LogResponse(200, "获取成功")
    c.JSON(map[string]string{"message": "success"})
}
```

### 4. 框架专用配置

```go
// Gin应用 - Web开发首选
logger := config.InitGlobalLogger(config.GinLogConfig())

// 微服务应用 - 分布式架构
logger := config.InitGlobalLogger(config.MicroserviceLogConfig())

// 数据库应用 - ORM操作追踪
logger := config.InitGlobalLogger(config.DatabaseLogConfig())

// 高性能应用 - 减少日志开销
logger := config.InitGlobalLogger(config.HighPerformanceFrameworkLogConfig())

// 指定框架格式
logger := config.InitGlobalLogger(config.WebFrameworkLogConfig("iris"))
```

### 5. 自定义配置

```go
customConfig := &config.LogConfig{
    Level:           config.LogLevelInfo,
    Format:          config.LogFormatGin,
    EnableConsole:   true,
    EnableFile:      true,
    FilePath:        "./logs/custom.log",
    MaxSize:         100,
    MaxAge:          7,
    MaxBackups:      10,
    Compress:        true,
    ShowCaller:      true,
    ShowTimestamp:   true,
    TimestampFormat: "2006-01-02 15:04:05",
    Fields: map[string]any{
        "service": "my-service",
        "version": "1.0.0",
    },
}

// 方式1: 全局配置
logger := config.InitGlobalLogger(customConfig)

// 方式2: 应用配置 (推荐)
app := controller.NewAppWithLogConfig(customConfig)
```

### 6. 中间件配置

```go
import "github.com/zsy619/yyhertz/framework/mvc/middleware"

func main() {
    app := controller.NewApp()
    
    // 使用默认日志中间件
    app.Use(middleware.LoggerMiddleware())
    
    // 自定义中间件配置
    loggerConfig := &middleware.LoggerConfig{
        EnableRequestBody:  true,  // 记录请求体
        EnableResponseBody: false, // 不记录响应体
        SkipPaths:         []string{"/health", "/ping"}, // 跳过健康检查
        MaxBodySize:       512,    // 最大记录512字节
    }
    app.Use(middleware.LoggerMiddlewareWithConfig(loggerConfig))
    
    // 简化版访问日志
    app.Use(middleware.AccessLogMiddleware())
    
    app.Run(":8080")
}
```

## ⚙️ 配置选项详解

### LogConfig 结构体字段说明

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Level` | `LogLevel` | `LogLevelInfo` | 日志级别 (debug/info/warn/error/fatal/panic) |
| `Format` | `LogFormat` | `LogFormatGin` | 日志格式 (支持17种格式) |
| `EnableConsole` | `bool` | `true` | 是否输出到控制台 |
| `EnableFile` | `bool` | `true` | 是否输出到文件 |
| `FilePath` | `string` | `"./logs/app.log"` | 日志文件路径 |
| `MaxSize` | `int` | `100` | 单个日志文件最大大小(MB) |
| `MaxAge` | `int` | `7` | 日志文件保留天数 |
| `MaxBackups` | `int` | `10` | 最大备份数量 |
| `Compress` | `bool` | `true` | 是否压缩旧日志 |
| `ShowCaller` | `bool` | `true` | 是否显示调用位置 |
| `ShowTimestamp` | `bool` | `true` | 是否显示时间戳 |
| `TimestampFormat` | `string` | `time.RFC3339` | 时间戳格式 |
| `Fields` | `map[string]any` | `{}` | 全局字段 |
| `Outputs` | `[]string` | `["console", "file"]` | 启用的输出类型 |
| `OutputConfig` | `map[string]OutputConfig` | `{}` | 各输出的配置 |

### 日志级别

| 级别 | 说明 | 使用场景 |
|------|------|----------|
| `debug` | 调试信息 | 开发环境详细调试 |
| `info` | 信息日志 | 常规应用信息 |
| `warn` | 警告日志 | 潜在问题警告 |
| `error` | 错误日志 | 应用错误 |
| `fatal` | 致命错误 | 导致程序退出的错误 |
| `panic` | 恐慌日志 | 严重错误，触发panic |

## 预设配置函数

### DefaultLogConfig()
- **格式**: Gin格式
- **级别**: Info
- **输出**: 控制台 + 文件
- **适用**: 一般Web应用

### DevelopmentLogConfig()
- **格式**: Gin格式
- **级别**: Debug
- **输出**: 控制台 + 文件
- **特点**: 详细调试信息，不压缩日志

### ProductionLogConfig()
- **格式**: Logstash格式
- **级别**: Info
- **输出**: 文件 + Fluentd
- **特点**: 生产环境优化，支持日志聚合

### TestLogConfig()
- **格式**: Beego格式
- **级别**: Warn
- **输出**: 仅控制台
- **适用**: 测试环境

### HighPerformanceLogConfig()
- **格式**: JSON格式
- **级别**: Error
- **输出**: 仅文件
- **特点**: 最小日志开销，高性能场景

### CloudLogConfig()
- **格式**: CloudWatch格式
- **输出**: 文件 + CloudWatch + Azure Insights
- **适用**: 云端部署

## 📊 日志输出示例

### Gin格式输出
```text
[GIN] [I] 2024/01/01 - 15:04:05 | 200 | 13ms | 127.0.0.1 | GET | /api/users HTTP请求处理完成
```

### JSON格式输出
```json
{
  "level": "info",
  "msg": "Request started: map[client_ip:127.0.0.1 method:GET path:/user/index request_id:abc123 timestamp:2024-01-01T12:00:00Z user_agent:curl/7.68.0]",
  "time": "2024-01-01T12:00:00Z",
  "caller": "middleware/logger.go:74"
}
```

### 文本格式输出
```text
time="2024-01-01T12:00:00Z" level=info msg="Request started" method=GET path="/user/index" request_id=abc123 client_ip="127.0.0.1"
```

### go-zero微服务格式输出
```json
{
  "@timestamp": "2024-01-01T12:00:00Z",
  "level": "info", 
  "service": "user-service",
  "trace": "abc123def456",
  "span": "span789",
  "message": "处理用户请求",
  "caller": "handler.go:42"
}
```

## 🔧 使用方法

### 1. 全局函数方式 (简单场景)

```go
// 基础日志函数
config.Debug("调试信息")
config.Info("信息日志") 
config.Warn("警告信息")
config.Error("错误信息")
config.Fatal("致命错误")

// 格式化日志
config.Debugf("用户 %s 登录成功", username)
config.Infof("服务启动在端口 %d", port)
```

### 2. 应用级日志方式 (推荐)

```go
app := controller.NewApp()

app.LogDebug("调试信息: %s", "debug message")
app.LogInfo("信息日志: %s", "info message") 
app.LogWarn("警告日志: %s", "warning message")
app.LogError("错误日志: %s", "error message")
app.LogFatal("致命错误: %s", "fatal message")
```

### 3. 控制器日志方式 (Web应用推荐)
```go
type UserController struct {
    controller.BaseController
}

func (c *UserController) PostCreate() {
    // 基础日志
    c.LogInfo("创建用户")
    c.LogRequest() // 自动记录请求详情
    
    // 结构化日志
    c.LogWithFields("info", "用户创建", map[string]any{
        "user_id": 123,
        "username": "john", 
        "action": "create",
        "ip": c.GetClientIP(),
    })
    
    // 记录响应
    c.LogResponse(200, "创建成功")
    c.JSON(map[string]string{"message": "created"})
}
```

### 4. 结构化日志 (推荐)
```go
// 全局函数方式
config.WithFields(map[string]any{
    "user_id": 1001,
    "action":  "login",
    "ip":      "192.168.1.1",
}).Info("用户操作")

config.WithField("request_id", "abc123").Error("请求失败")

// 请求追踪
requestID := generateRequestID()
logger := config.WithRequestID(requestID)
logger.Info("开始处理请求")
logger.Info("请求处理完成")
```

### 专用日志函数

#### HTTP请求日志
```go
config.LogHTTPRequest("GET", "/api/users", "127.0.0.1", 200, 15.5)
```

#### 数据库操作日志
```go
config.LogDBQuery("SELECT * FROM users WHERE id = ?", 12.5, 1, nil)
```

#### 性能日志
```go
config.LogPerformance("user_query", 25.3)
```

#### 业务事件日志
```go
config.LogBusinessEvent("user_register", "user123", "user", map[string]any{
    "email": "user@example.com",
    "source": "web",
})
```

#### 安全事件日志
```go
config.LogSecurityEvent("failed_login", "user123", "192.168.1.100", map[string]any{
    "attempts": 3,
    "reason": "invalid_password",
})
```

## 动态配置

### 运行时更新日志级别
```go
config.UpdateGlobalLogLevel(config.LogLevelDebug)
```

### 运行时更新日志格式
```go
config.UpdateGlobalLogFormat(config.LogFormatJSON)
```

### 动态添加输出
```go
fluentdConfig := config.FluentdConfig{
    Host: "localhost",
    Port: 24224,
    Tag:  "myapp.logs",
}
config.AddGlobalLogOutput("fluentd", fluentdConfig)
```

## 输出目标配置

### 文件输出
```go
config := &config.LogConfig{
    EnableFile: true,
    FilePath:   "./logs/app.log",
    MaxSize:    100,    // 100MB
    MaxAge:     7,      // 7天
    MaxBackups: 10,     // 10个备份
    Compress:   true,   // 压缩旧文件
}
```

### Syslog输出
```go
syslogConfig := config.SyslogConfig{
    Network:  "udp",
    Address:  "localhost:514",
    Priority: 16, // local0.info
    Tag:      "myapp",
}

config := &config.LogConfig{
    Outputs: []string{"console", "syslog"},
    OutputConfig: map[string]config.OutputConfig{
        "syslog": syslogConfig,
    },
}
```

### Fluentd输出
```go
fluentdConfig := config.FluentdConfig{
    Host:    "localhost",
    Port:    24224,
    Tag:     "myapp.logs",
    Timeout: 3 * time.Second,
    Extra: map[string]string{
        "environment": "production",
    },
}
```

### CloudWatch输出
```go
cloudWatchConfig := config.CloudWatchConfig{
    Region:          "us-east-1",
    LogGroupName:    "/aws/myapp/application",
    LogStreamName:   "myapp-instance-001",
    AccessKeyID:     "your-access-key",
    SecretAccessKey: "your-secret-key",
}
```

## 🎯 最佳实践

### 1. 选择合适的日志格式

| 场景 | 推荐格式 | 原因 |
|------|---------|------|
| 开发调试 | `gin`, `iris`, `revel` | 易读，带颜色，信息完整 |
| 生产部署 | `json`, `logstash` | 结构化，易于解析和聚合 |
| 高性能 | `fiber`, `text` | 开销小，输出简洁 |
| 微服务 | `go_zero`, `json` | 支持分布式追踪 |
| 数据库 | `ent` | 专门显示SQL和执行时间 |
| 云端部署 | `cloudwatch`, `azure_insights` | 原生集成云服务 |

### 2. 合理设置日志级别

```go
// 开发环境配置
config := &config.LogConfig{
    Level: config.LogLevelDebug, // 详细调试信息
    Format: config.LogFormatGin,
}

// 测试环境配置  
config := &config.LogConfig{
    Level: config.LogLevelInfo,  // 关注关键流程
    Format: config.LogFormatText,
}

// 生产环境配置
config := &config.LogConfig{
    Level: config.LogLevelWarn,  // 只记录警告以上级别
    Format: config.LogFormatJSON,
}
```

### 3. 结构化日志规范

```go
// ✅ 推荐：使用WithFields添加结构化字段
config.WithFields(map[string]any{
    "user_id":    1001,
    "request_id": "abc123", 
    "operation":  "user_login",
    "duration":   "15ms",
    "status":     "success",
}).Info("用户登录成功")

// ❌ 不推荐：在消息中拼接信息
config.Info("用户 1001 登录成功，请求ID abc123，耗时15ms")
```

### 4. 请求追踪最佳实践

```go
// 在中间件中设置请求ID
func LoggerMiddleware() middleware.HandlerFunc {
    return func(c *Context) {
        requestID := generateRequestID()
        c.Set("request_id", requestID)
        
        // 使用WithRequestID记录
        logger := config.WithRequestID(requestID)
        logger.Info("请求开始处理")
        
        c.Next()
        
        logger.Info("请求处理完成")
    }
}

// 在控制器中继续使用
func (c *UserController) GetUser() {
    requestID := c.GetString("request_id")
    config.WithRequestID(requestID).Info("获取用户信息")
}
```

### 5. 敏感信息处理

```go
// ✅ 安全的日志记录
config.WithFields(map[string]any{
    "username": user.Username,
    "email":    maskEmail(user.Email), // 脱敏处理
    "login_ip": user.IP,
    // "password": user.Password, // 绝对不记录密码
}).Info("用户登录")

// 脱敏函数示例
func maskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    username := parts[0]
    if len(username) > 2 {
        username = username[:2] + "***"
    }
    return username + "@" + parts[1]
}
```

### 6. 文件轮转配置

```go
// 高频应用配置
config := &config.LogConfig{
    MaxSize:    50,   // 50MB，避免单文件过大
    MaxAge:     7,    // 7天，根据业务需求
    MaxBackups: 20,   // 更多备份
    Compress:   true, // 压缩节省空间
}

// 低频应用配置  
config := &config.LogConfig{
    MaxSize:    200,  // 200MB，减少文件数量
    MaxAge:     30,   // 30天，更长保留期
    MaxBackups: 5,    // 较少备份
    Compress:   true,
}
```

### 7. 性能优化建议

```go
// 高并发场景优化
config := &config.LogConfig{
    Level:       config.LogLevelWarn,  // 减少日志量
    ShowCaller:  false,                // 关闭调用栈
    EnableFile:  true,
    EnableConsole: false,              // 生产环境关闭控制台输出
}

// 异步日志处理
config.AddGlobalLogOutput("async_file", AsyncFileConfig{
    BufferSize: 1024,
    FlushInterval: time.Second,
})
```

## ❓ 常见问题

### Q: 如何在不同的包中使用日志？
A: 有两种推荐方式：
```go
// 方式1: 使用全局函数
import "github.com/zsy619/yyhertz/framework/config"

func someFunction() {
    config.Info("这是在其他包中的日志")
}

// 方式2: 传递日志实例 (推荐)
func someFunction(logger *logrus.Logger) {
    logger.Info("这是传递的日志实例")
}
```

### Q: 如何在Gin中间件中记录请求日志？
A: 使用YYHertz的内置中间件或自定义：
```go
// 使用内置中间件 (推荐)
app.Use(middleware.LoggerMiddleware())

// 或自定义Gin风格日志
gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
    config.WithFields(map[string]any{
        "status_code": param.StatusCode,
        "latency":     param.Latency.String(), 
        "client_ip":   param.ClientIP,
        "method":      param.Method,
        "path":        param.Path,
    }).Info("HTTP请求")
    return ""
})
```

### Q: 日志文件过大怎么办？
A: 配置合适的文件轮转参数：
```go
config := &config.LogConfig{
    MaxSize:    10,   // 单文件最大10MB
    MaxAge:     7,    // 保留7天
    MaxBackups: 5,    // 最多5个备份文件
    Compress:   true, // 压缩旧文件
}
```

### Q: 如何避免敏感信息被记录？
A: 实施数据脱敏和字段过滤：
```go
// 创建脱敏函数
func sanitizeUser(user User) map[string]any {
    return map[string]any{
        "user_id":  user.ID,
        "username": user.Username,
        "email":    maskEmail(user.Email),
        // 绝对不记录: password, token, 身份证号等
    }
}

// 使用脱敏后的数据记录日志
config.WithFields(sanitizeUser(user)).Info("用户操作")
```

### Q: 如何处理高并发下的日志性能？
A: 多种优化策略：
```go
// 1. 提高日志级别
config := &config.LogConfig{
    Level: config.LogLevelWarn, // 只记录警告以上
}

// 2. 关闭调用栈
config.ShowCaller = false

// 3. 使用异步输出
config.EnableConsole = false  // 关闭控制台输出
config.EnableFile = true      // 只写文件

// 4. 批量写入
config.AddGlobalLogOutput("buffer", BufferConfig{
    BufferSize: 1024,
    FlushInterval: 100 * time.Millisecond,
})
```

### Q: 不同环境如何配置？
A: 使用环境变量或配置文件：
```go
func getLogConfig() *config.LogConfig {
    env := os.Getenv("APP_ENV")
    switch env {
    case "development":
        return config.DevelopmentLogConfig()
    case "production": 
        return config.ProductionLogConfig()
    case "test":
        return config.TestLogConfig()
    default:
        return config.DefaultLogConfig()
    }
}
```

## 🔍 故障排查

### 1. 日志文件权限问题
**现象**: 程序启动报错，无法创建日志文件
**解决方案**:
```bash
# 确保日志目录存在且有写权限
mkdir -p logs
chmod 755 logs

# 检查磁盘空间
df -h

# 检查文件权限
ls -la logs/
```

### 2. 日志文件过大导致磁盘满
**现象**: 磁盘空间不足，系统性能下降
**解决方案**:
```go
// 调整日志轮转配置
config := &config.LogConfig{
    MaxSize:    10,   // 减小单文件大小
    MaxAge:     3,    // 缩短保留时间
    MaxBackups: 3,    // 减少备份数量
    Compress:   true, // 启用压缩
}

// 或临时清理
rm logs/*.log.gz  // 删除压缩的旧日志
```

### 3. 日志格式乱码或显示异常
**现象**: 控制台或文件中出现乱码
**解决方案**:
```go
// 确保使用UTF-8编码
config := &config.LogConfig{
    TimestampFormat: "2006-01-02 15:04:05", // 避免特殊字符
    ShowColors:      false, // 在某些环境下关闭颜色
}

// 检查终端编码
echo $LANG  // 应该包含UTF-8
```

### 4. 性能问题：日志拖慢应用
**现象**: 应用响应变慢，CPU使用率高
**解决方案**:
```go
// 优化日志配置
config := &config.LogConfig{
    Level:         config.LogLevelWarn,  // 提高日志级别
    EnableConsole: false,                // 关闭控制台输出
    ShowCaller:    false,                // 关闭调用栈
    Fields:        nil,                  // 减少全局字段
}

// 避免在高频路径记录Debug日志
if config.GetGlobalLogger().IsLevelEnabled(logrus.DebugLevel) {
    config.Debug("这是耗时的调试信息")
}
```

### 5. 中间件日志重复记录
**现象**: 同一个请求被记录多次
**解决方案**:
```go
// 检查中间件配置，避免重复注册
app.Use(middleware.LoggerMiddleware()) // 只注册一次

// 或配置跳过路径
loggerConfig := &middleware.LoggerConfig{
    SkipPaths: []string{"/health", "/ping", "/metrics"},
}
app.Use(middleware.LoggerMiddlewareWithConfig(loggerConfig))
```

## 🏗️ 环境配置示例

### 开发环境 (development)
```go
func DevelopmentLogConfig() *config.LogConfig {
    return &config.LogConfig{
        Level:           config.LogLevelDebug,
        Format:          config.LogFormatGin,    // 易读格式
        EnableConsole:   true,                   // 控制台输出
        EnableFile:      true,
        FilePath:        "logs/dev.log",
        MaxSize:         50,
        MaxAge:          3,
        MaxBackups:      5,
        Compress:        false,                  // 不压缩，便于查看
        ShowCaller:      true,                   // 显示调用位置
        ShowTimestamp:   true,
        TimestampFormat: "2006/01/02 15:04:05",
        Fields: map[string]any{
            "env":     "development",
            "service": "yyhertz-dev",
        },
    }
}
```

### 生产环境 (production)
```go
func ProductionLogConfig() *config.LogConfig {
    return &config.LogConfig{
        Level:           config.LogLevelInfo,
        Format:          config.LogFormatJSON,   // 结构化格式
        EnableConsole:   false,                  // 不输出控制台
        EnableFile:      true,
        FilePath:        "/var/log/app/prod.log",
        MaxSize:         100,
        MaxAge:          30,                     // 保留30天
        MaxBackups:      10,
        Compress:        true,                   // 压缩旧日志
        ShowCaller:      false,                  // 不显示调用位置
        ShowTimestamp:   true,
        TimestampFormat: time.RFC3339,
        Fields: map[string]any{
            "env":     "production",
            "service": "yyhertz",
            "version": "1.0.0",
        },
        Outputs: []string{"file", "syslog"},     // 多输出
    }
}
```

### 测试环境 (test)
```go
func TestLogConfig() *config.LogConfig {
    return &config.LogConfig{
        Level:         config.LogLevelWarn,      // 只记录警告以上
        Format:        config.LogFormatText,
        EnableConsole: true,
        EnableFile:    false,                    // 不写文件
        ShowTimestamp: false,                    // 测试时不需要时间戳
        ShowCaller:    false,
        Fields: map[string]any{
            "env": "test",
        },
    }
}
```

## 📚 更多资源

- **官方文档**: [YYHertz框架文档](https://github.com/zsy619/yyhertz)
- **Logrus文档**: [sirupsen/logrus](https://github.com/sirupsen/logrus)
- **Hertz-Contrib**: [hertz-contrib/logger](https://github.com/hertz-contrib/logger)
- **日志最佳实践**: [12-Factor App Logs](https://12factor.net/logs)
- **结构化日志**: [Structured Logging](https://www.honeycomb.io/blog/structured-logging-and-your-team/)

## 📋 更新日志

### v1.3.0 (当前)
- 📚 完善文档，合并所有日志相关文档
- ✨ 新增故障排查和环境配置示例
- 🔧 优化最佳实践建议
- 📊 增加更多输出格式示例

### v1.2.0
- ✨ 新增8个流行框架的日志格式支持
- ✨ 添加框架专用预设配置函数
- 🐛 修复Gin格式化器的级别显示问题
- 📚 完善文档和使用示例

### v1.1.0
- ✨ 新增企业级日志输出支持
- ✨ 添加动态配置功能
- 🔧 优化文件轮转机制

### v1.0.0
- 🎉 初始版本发布
- ✨ 基础日志格式支持
- ✨ 多输出目标支持

---

💡 **提示**: 本文档涵盖了YYHertz日志系统的完整使用方法。如有疑问，请参考示例代码或提交Issue。