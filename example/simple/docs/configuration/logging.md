# 系统日志

YYHertz框架提供了强大的日志系统，支持多种输出格式、日志级别和轮转策略，帮助开发者有效地记录和监控应用运行状态。

## 快速开始

### 1. 日志配置

在 `conf/log.yaml` 中配置日志系统：

```yaml
log:
  # 日志级别: trace, debug, info, warn, error, fatal, panic
  level: "info"
  
  # 输出格式: json, text
  format: "json"
  
  # 输出目标
  outputs:
    - type: "console"     # 控制台输出
    - type: "file"        # 文件输出
      filename: "logs/app.log"
      max_size: 100       # 单个文件最大尺寸(MB)
      max_backups: 10     # 保留的备份文件数量
      max_age: 30         # 文件保留天数
      compress: true      # 是否压缩备份文件
  
  # 开发模式配置
  development: false
  
  # 调用者信息
  disable_caller: false
  
  # 时间格式
  time_format: "2006-01-02 15:04:05"
```

### 2. 基本使用

```go
import "github.com/zsy619/yyhertz/framework/log"

func (c *HomeController) GetIndex() {
    // 不同级别的日志
    log.Debug("调试信息")
    log.Info("普通信息")
    log.Warn("警告信息")
    log.Error("错误信息")
    
    // 带字段的结构化日志
    log.WithFields(log.Fields{
        "user_id": 123,
        "action":  "login",
        "ip":      "192.168.1.1",
    }).Info("用户登录")
    
    // 带上下文的日志
    ctx := c.GetContext()
    log.WithContext(ctx).Info("处理请求")
}
```

## 日志级别

### 级别说明

- **TRACE**: 非常详细的调试信息
- **DEBUG**: 调试信息，开发时使用
- **INFO**: 一般信息，记录程序正常运行状态
- **WARN**: 警告信息，程序可以继续运行但需要注意
- **ERROR**: 错误信息，程序遇到错误但仍可继续
- **FATAL**: 致命错误，程序无法继续运行
- **PANIC**: 恐慌级别，会触发panic

### 动态设置日志级别

```go
import "github.com/zsy619/yyhertz/framework/log"

// 设置全局日志级别
log.SetLevel(log.InfoLevel)

// 运行时修改日志级别
func (c *AdminController) PostSetLogLevel() {
    level := c.GetForm("level")
    
    switch level {
    case "debug":
        log.SetLevel(log.DebugLevel)
    case "info":
        log.SetLevel(log.InfoLevel)
    case "warn":
        log.SetLevel(log.WarnLevel)
    case "error":
        log.SetLevel(log.ErrorLevel)
    default:
        c.Error(400, "无效的日志级别")
        return
    }
    
    log.WithField("level", level).Info("日志级别已更新")
    c.JSON(map[string]interface{}{
        "success": true,
        "message": "日志级别设置成功",
    })
}
```

## 结构化日志

### 字段日志

```go
// 单个字段
log.WithField("user_id", 123).Info("用户操作")

// 多个字段
log.WithFields(log.Fields{
    "method":     "POST",
    "url":        "/api/users",
    "status":     201,
    "duration":   "45ms",
    "request_id": "req-123456",
}).Info("API请求完成")

// 嵌套结构
log.WithFields(log.Fields{
    "user": map[string]interface{}{
        "id":       123,
        "username": "john",
        "role":     "admin",
    },
    "action": "update_profile",
}).Info("用户资料更新")
```

### 上下文日志

```go
import (
    "context"
    "github.com/zsy619/yyhertz/framework/log"
)

func (c *UserController) PostCreate() {
    ctx := c.GetContext()
    
    // 为上下文添加字段
    ctx = log.WithContext(ctx, log.Fields{
        "operation": "create_user",
        "ip":        c.GetClientIP(),
    })
    
    // 使用带上下文的日志
    log.FromContext(ctx).Info("开始创建用户")
    
    // 调用服务层
    err := c.userService.CreateUser(ctx, userData)
    if err != nil {
        log.FromContext(ctx).WithError(err).Error("创建用户失败")
        c.Error(500, "创建失败")
        return
    }
    
    log.FromContext(ctx).Info("用户创建成功")
}
```

## 中间件集成

### 请求日志中间件

```go
import (
    "time"
    "github.com/zsy619/yyhertz/framework/log"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
    app := mvc.HertzApp
    
    // 添加请求日志中间件
    app.Use(middleware.RequestLogger())
    
    // 自定义请求日志中间件
    app.Use(func(c *app.RequestContext) {
        start := time.Now()
        
        // 请求开始日志
        log.WithFields(log.Fields{
            "method":     string(c.Method()),
            "path":       string(c.Path()),
            "ip":         c.ClientIP(),
            "user_agent": string(c.UserAgent()),
        }).Info("请求开始")
        
        c.Next()
        
        // 请求完成日志
        duration := time.Since(start)
        log.WithFields(log.Fields{
            "method":     string(c.Method()),
            "path":       string(c.Path()),
            "status":     c.Response.StatusCode(),
            "duration":   duration.String(),
            "ip":         c.ClientIP(),
        }).Info("请求完成")
    })
    
    app.Run()
}
```

### 错误日志中间件

```go
app.Use(func(c *app.RequestContext) {
    defer func() {
        if r := recover(); r != nil {
            log.WithFields(log.Fields{
                "method": string(c.Method()),
                "path":   string(c.Path()),
                "ip":     c.ClientIP(),
                "panic":  r,
            }).Error("请求发生panic")
            
            c.JSON(500, map[string]interface{}{
                "error": "Internal Server Error",
            })
        }
    }()
    
    c.Next()
})
```

## 日志输出配置

### 控制台输出

```yaml
log:
  outputs:
    - type: "console"
      format: "text"        # text 或 json
      color: true           # 是否启用颜色
      timestamp: true       # 是否显示时间戳
```

### 文件输出

```yaml
log:
  outputs:
    - type: "file"
      filename: "logs/app.log"
      max_size: 100         # 单个文件最大尺寸(MB)
      max_backups: 10       # 保留的备份文件数量  
      max_age: 30           # 文件保留天数
      compress: true        # 是否压缩备份文件
      format: "json"        # 输出格式
```

### 多重输出

```yaml
log:
  outputs:
    # 控制台输出 - 开发环境
    - type: "console"
      format: "text"
      color: true
      min_level: "debug"
    
    # 应用日志文件
    - type: "file"
      filename: "logs/app.log"
      format: "json"
      max_size: 100
      max_backups: 10
      max_age: 30
      min_level: "info"
    
    # 错误日志文件
    - type: "file"
      filename: "logs/error.log"
      format: "json"
      max_size: 50
      max_backups: 20
      max_age: 60
      min_level: "error"
```

## 高级功能

### 异步日志

```go
import "github.com/zsy619/yyhertz/framework/log/async"

func main() {
    // 启用异步日志，提高性能
    async.Enable(async.Config{
        BufferSize: 1000,     // 缓冲区大小
        Workers:    2,        // 工作协程数量
        FlushInterval: "1s",  // 刷新间隔
    })
    
    defer async.Flush() // 确保所有日志都被写入
    
    app := mvc.HertzApp
    app.Run()
}
```

### 日志采样

```go
import "github.com/zsy619/yyhertz/framework/log/sampling"

// 配置日志采样，避免日志过多
sampling.Config{
    Initial:    100,  // 初始日志数量
    Thereafter: 100,  // 之后每隔100条记录一次
}
```

### 自定义日志格式

```go
import "github.com/zsy619/yyhertz/framework/log/formatter"

// 自定义JSON格式化器
customFormatter := &formatter.JSONFormatter{
    TimestampFormat: "2006-01-02 15:04:05.000",
    PrettyPrint:     false,
    FieldMap: formatter.FieldMap{
        "level": "severity",
        "time":  "timestamp",
        "msg":   "message",
    },
}

log.SetFormatter(customFormatter)
```

## 监控和告警

### 错误统计

```go
import "github.com/zsy619/yyhertz/framework/log/metrics"

// 启用日志指标收集
metrics.Enable()

// 获取错误统计
errorCount := metrics.GetErrorCount()
warnCount := metrics.GetWarnCount()

log.WithFields(log.Fields{
    "error_count": errorCount,
    "warn_count":  warnCount,
}).Info("日志统计")
```

### 告警集成

```go
import "github.com/zsy619/yyhertz/framework/log/alert"

// 配置告警规则
alert.Configure(alert.Config{
    // 错误日志超过阈值时发送告警
    Rules: []alert.Rule{
        {
            Level:     "error",
            Threshold: 10,        // 10分钟内超过10条错误日志
            Window:    "10m",
            Action:    "webhook", // 通过webhook发送告警
            URL:       "https://hooks.slack.com/xxx",
        },
    },
})
```

## 性能优化

### 1. 异步写入

```yaml
log:
  async: true
  buffer_size: 1000
  flush_interval: "1s"
```

### 2. 日志级别过滤

```go
// 生产环境只记录INFO及以上级别的日志
if env := os.Getenv("ENV"); env == "production" {
    log.SetLevel(log.InfoLevel)
}
```

### 3. 条件日志

```go
// 避免不必要的字符串格式化
if log.IsDebugEnabled() {
    log.WithFields(log.Fields{
        "expensive_operation": expensiveFunction(),
    }).Debug("调试信息")
}
```

## 最佳实践

### 1. 日志级别使用

- **DEBUG**: 详细的调试信息，仅开发环境使用
- **INFO**: 重要的业务流程节点
- **WARN**: 需要关注但不影响正常流程的情况
- **ERROR**: 程序错误，需要修复但不会导致崩溃
- **FATAL**: 致命错误，程序无法继续运行

### 2. 字段命名规范

```go
// 推荐的字段命名
log.WithFields(log.Fields{
    "user_id":    123,
    "request_id": "req-abc123",
    "operation":  "user_login",
    "duration":   "150ms",
    "status":     "success",
}).Info("用户操作完成")
```

### 3. 错误处理

```go
func (c *UserController) GetUser() {
    user, err := c.userService.GetUser(userID)
    if err != nil {
        // 记录错误详情
        log.WithFields(log.Fields{
            "user_id": userID,
            "error":   err.Error(),
            "stack":   fmt.Sprintf("%+v", err),
        }).Error("获取用户信息失败")
        
        // 返回用户友好的错误信息
        c.Error(500, "获取用户信息失败")
        return
    }
    
    // 记录成功操作
    log.WithField("user_id", userID).Info("获取用户信息成功")
    c.JSON(user)
}
```

### 4. 避免敏感信息

```go
// ❌ 不要记录敏感信息
log.WithFields(log.Fields{
    "password": password,
    "token":    authToken,
}).Info("用户登录")

// ✅ 使用脱敏或省略敏感信息
log.WithFields(log.Fields{
    "user_id": userID,
    "success": true,
}).Info("用户登录成功")
```

## 配置示例

### 开发环境配置

```yaml
log:
  level: "debug"
  format: "text"
  development: true
  outputs:
    - type: "console"
      color: true
```

### 生产环境配置

```yaml
log:
  level: "info"
  format: "json"
  development: false
  outputs:
    - type: "file"
      filename: "/var/log/app/app.log"
      max_size: 100
      max_backups: 30
      max_age: 30
      compress: true
    - type: "file"
      filename: "/var/log/app/error.log"
      max_size: 50
      max_backups: 60
      max_age: 60
      min_level: "error"
      compress: true
```