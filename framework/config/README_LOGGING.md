# Logrus 日志集成使用指南

## 📋 概述

框架已完全集成 [hertz-contrib/logger/logrus](https://github.com/hertz-contrib/logger/tree/main/logrus)，提供强大的结构化日志功能。

## 🚀 快速开始

### 1. 使用默认日志配置

```go
package main

import (
    "hertz-controller/framework/controller"
)

func main() {
    // 使用默认logrus配置
    app := controller.NewApp()
    
    app.LogInfo("应用启动成功")
    app.Run(":8080")
}
```

### 2. 自定义日志配置

```go
package main

import (
    "time"
    "hertz-controller/framework/config"
    "hertz-controller/framework/controller"
)

func main() {
    // 创建自定义日志配置
    logConfig := &config.LogConfig{
        Level:           config.LogLevelDebug,
        Format:          config.LogFormatJSON,
        EnableConsole:   true,
        EnableFile:      true,
        FilePath:        "logs/app.log",
        MaxSize:         100,
        MaxAge:          7,
        MaxBackups:      10,
        Compress:        true,
        ShowCaller:      true,
        ShowTimestamp:   true,
        TimestampFormat: time.RFC3339,
        Fields: map[string]interface{}{
            "service": "my-service",
            "version": "1.0.0",
        },
    }

    app := controller.NewAppWithLogConfig(logConfig)
    app.Run(":8080")
}
```

## ⚙️ 配置选项

### LogConfig 结构体字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Level` | `LogLevel` | `LogLevelInfo` | 日志级别 |
| `Format` | `LogFormat` | `LogFormatJSON` | 日志格式 |
| `EnableConsole` | `bool` | `true` | 是否输出到控制台 |
| `EnableFile` | `bool` | `true` | 是否输出到文件 |
| `FilePath` | `string` | `"logs/app.log"` | 日志文件路径 |
| `MaxSize` | `int` | `100` | 单个日志文件最大大小(MB) |
| `MaxAge` | `int` | `7` | 日志文件保留天数 |
| `MaxBackups` | `int` | `10` | 最大备份数量 |
| `Compress` | `bool` | `true` | 是否压缩旧日志 |
| `ShowCaller` | `bool` | `true` | 是否显示调用位置 |
| `ShowTimestamp` | `bool` | `true` | 是否显示时间戳 |
| `TimestampFormat` | `string` | `time.RFC3339` | 时间戳格式 |
| `Fields` | `map[string]interface{}` | `{}` | 全局字段 |

### 日志级别

```go
config.LogLevelDebug  // debug
config.LogLevelInfo   // info
config.LogLevelWarn   // warn
config.LogLevelError  // error
config.LogLevelFatal  // fatal
config.LogLevelPanic  // panic
```

### 日志格式

```go
config.LogFormatJSON  // JSON格式
config.LogFormatText  // 文本格式
```

## 🔧 使用方法

### 1. 在应用级别记录日志

```go
app := controller.NewApp()

app.LogDebug("调试信息: %s", "debug message")
app.LogInfo("信息日志: %s", "info message")
app.LogWarn("警告日志: %s", "warning message")
app.LogError("错误日志: %s", "error message")
app.LogFatal("致命错误: %s", "fatal message")
```

### 2. 在控制器中记录日志

```go
type UserController struct {
    controller.BaseController
}

func (c *UserController) GetIndex() {
    c.LogInfo("获取用户列表")
    c.LogRequest() // 记录请求详情
    
    // 业务逻辑...
    
    c.LogResponse(200, "获取成功") // 记录响应
    c.JSON(map[string]string{"message": "success"})
}

func (c *UserController) PostCreate() {
    c.LogInfo("创建用户")
    
    // 带字段的日志
    c.LogWithFields("info", "用户创建", map[string]interface{}{
        "user_id": 123,
        "username": "john",
        "action": "create",
    })
    
    c.JSON(map[string]string{"message": "created"})
}
```

### 3. 日志中间件配置

```go
import "hertz-controller/framework/middleware"

// 使用默认配置
app.Use(middleware.LoggerMiddleware())

// 或者自定义配置
loggerConfig := &middleware.LoggerConfig{
    EnableRequestBody:  true,  // 记录请求体
    EnableResponseBody: false, // 不记录响应体
    SkipPaths:         []string{"/health", "/ping"}, // 跳过健康检查
    MaxBodySize:       512,    // 最大记录512字节
}
app.Use(middleware.LoggerMiddlewareWithConfig(loggerConfig))

// 简化版访问日志
app.Use(middleware.AccessLogMiddleware())
```

## 📊 日志输出示例

### JSON 格式输出

```json
{
  "level": "info",
  "msg": "Request started: map[client_ip:127.0.0.1 method:GET path:/user/index request_id:abc123 timestamp:2024-01-01T12:00:00Z user_agent:curl/7.68.0]",
  "time": "2024-01-01T12:00:00Z",
  "caller": "middleware/logger.go:74"
}
```

### 文本格式输出

```
time="2024-01-01T12:00:00Z" level=info msg="Request started" method=GET path="/user/index" request_id=abc123 client_ip="127.0.0.1"
```

## 🏗️ 预设配置示例

### 开发环境

```go
func DevelopmentConfig() *config.LogConfig {
    return &config.LogConfig{
        Level:           config.LogLevelDebug,
        Format:          config.LogFormatText,
        EnableConsole:   true,
        EnableFile:      true,
        FilePath:        "logs/dev.log",
        ShowCaller:      true,
        ShowTimestamp:   true,
        Fields: map[string]interface{}{
            "env": "development",
        },
    }
}
```

### 生产环境

```go
func ProductionConfig() *config.LogConfig {
    return &config.LogConfig{
        Level:           config.LogLevelInfo,
        Format:          config.LogFormatJSON,
        EnableConsole:   false,
        EnableFile:      true,
        FilePath:        "logs/prod.log",
        MaxSize:         100,
        MaxAge:          30,
        MaxBackups:      10,
        Compress:        true,
        ShowCaller:      false,
        Fields: map[string]interface{}{
            "env":     "production",
            "service": "my-service",
        },
    }
}
```

### 测试环境

```go
func TestConfig() *config.LogConfig {
    return &config.LogConfig{
        Level:         config.LogLevelWarn,
        Format:        config.LogFormatText,
        EnableConsole: true,
        EnableFile:    false,
        ShowTimestamp: false,
    }
}
```

## 🎯 最佳实践

### 1. 结构化日志

```go
// 推荐：使用结构化字段
c.LogWithFields("info", "用户操作", map[string]interface{}{
    "user_id":   123,
    "action":    "login",
    "ip":        "192.168.1.1",
    "timestamp": time.Now(),
})

// 不推荐：纯文本日志
c.LogInfo("用户123从192.168.1.1登录")
```

### 2. 合理的日志级别

- `Debug`: 开发调试信息
- `Info`: 一般信息，如请求处理
- `Warn`: 警告，需要注意但不影响功能
- `Error`: 错误，影响功能但不致命
- `Fatal/Panic`: 致命错误，导致程序退出

### 3. 请求追踪

```go
// 中间件会自动生成request_id
// 在控制器中可以获取并使用
requestID := c.Ctx.GetString("request_id")
c.LogWithFields("info", "处理业务逻辑", map[string]interface{}{
    "request_id": requestID,
    "step": "validation",
})
```

### 4. 敏感信息处理

```go
// 不要记录敏感信息
c.LogWithFields("info", "用户登录", map[string]interface{}{
    "username": user.Username,
    // "password": user.Password, // 绝对不要记录密码
    "login_time": time.Now(),
})
```

## 🔍 故障排查

### 1. 日志文件权限问题

确保应用有权限写入日志目录：

```bash
mkdir -p logs
chmod 755 logs
```

### 2. 日志文件过大

配置日志轮转：

```go
logConfig := &config.LogConfig{
    MaxSize:    10,  // 10MB
    MaxAge:     7,   // 7天
    MaxBackups: 5,   // 最多5个备份
    Compress:   true, // 压缩旧文件
}
```

### 3. 性能考虑

- 生产环境避免使用`Debug`级别
- 大量日志时考虑异步写入
- 合理配置日志轮转避免磁盘占满

## 📚 更多资源

- [Logrus官方文档](https://github.com/sirupsen/logrus)
- [Hertz-Contrib Logger](https://github.com/hertz-contrib/logger)
- [日志最佳实践](https://12factor.net/logs)