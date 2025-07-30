# Hertz MVC 框架集成文档

## 概述

本文档记录了将 `main_with_framework.go` 集成到 `version.go` 文件的过程和相关信息。

## 集成内容

### 1. 版本信息模块

保留了原有的版本信息常量和函数：
- `FrameworkName`, `FrameworkVersion`, `BuildDate` 等常量
- `GetVersionInfo()`, `PrintVersion()`, `PrintBanner()` 等函数
- `GetSystemInfo()`, `GetHealthStatus()` 等系统信息函数

### 2. 控制器定义

集成了三个主要控制器：

#### SystemController（系统控制器）
- `GetVersion()` - 获取版本信息
- `GetHealth()` - 健康检查
- `GetInfo()` - 获取完整系统信息

#### UserController（用户控制器）
- `GetIndex()` - 获取用户列表
- `GetInfo()` - 获取用户详情
- `PostCreate()` - 创建用户

#### HomeController（首页控制器）
- `GetIndex()` - 获取首页信息

### 3. 中间件

集成了日志中间件：
- `LoggerMiddleware()` - 记录请求日志和响应时间

### 4. 主函数

完整的应用启动逻辑：
- 命令行参数解析（version, banner, port）
- 控制器注册
- 路由配置
- 服务器启动

## API 接口文档

### 系统接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | / | 首页信息 |
| GET | /api | API文档 |
| GET | /system/version | 版本信息 |
| GET | /system/health | 健康检查 |
| GET | /system/info | 系统信息 |

### 业务接口

| 方法 | 路径 | 描述 | 参数 |
|------|------|------|------|
| GET | /home/index | 首页 | - |
| GET | /user/index | 用户列表 | - |
| GET | /user/info | 用户详情 | id, name |
| POST | /user/create | 创建用户 | name, email |

## 使用方法

### 构建应用

```bash
go build -o main version.go
```

### 启动应用

```bash
# 默认启动
./main

# 查看版本
./main --version

# 指定端口启动
./main --port 8889

# 不显示横幅
./main --banner=false
```

### 测试接口

```bash
# 首页
curl http://localhost:8888/

# 版本信息
curl http://localhost:8888/system/version

# 用户列表
curl http://localhost:8888/user/index

# 创建用户
curl -X POST http://localhost:8888/user/create -d 'name=张三&email=test@example.com'
```

## 特性

1. **完整的版本管理系统** - 包含框架版本、构建信息、依赖版本等
2. **系统监控接口** - 提供健康检查、系统信息等监控接口
3. **RESTful API设计** - 标准的REST接口设计
4. **中间件支持** - 内置日志中间件，支持扩展
5. **命令行参数** - 支持版本查看、端口配置等
6. **用户友好的启动信息** - 显示横幅、路由列表、测试命令等

## 框架依赖

- CloudWeGo Hertz v0.10.1
- Go 1.18+
- 自定义 github.com/zsy619/yyhertz/framework/controller 模块

## 注意事项

1. 确保 `github.com/zsy619/yyhertz/framework/controller` 模块正确实现
2. 所有中文字符已正确编码
3. 端口默认为 8888，可通过 `--port` 参数修改
4. 支持 `--version` 和 `--banner` 命令行选项

## 更新历史

- 2025-07-29: 完成 main_with_framework.go 到 version.go 的集成
- 集成了完整的控制器、中间件和主函数
- **日志系统升级**: 将所有日志输出迁移到 `framework/config/logger_singleton.go` 单例日志系统
- 更新了相关文档和使用说明

## 日志系统特性

### 新的日志功能
- **结构化日志**: 使用JSON格式输出，便于日志分析和监控
- **单例模式**: 全局统一的日志实例，确保配置一致性
- **多级别支持**: Debug, Info, Warn, Error, Fatal, Panic
- **字段化日志**: 支持添加自定义字段进行上下文追踪
- **线程安全**: 支持并发环境下的安全日志记录

### 日志示例
```json
{"action":"login","file":"version.go:280","func":"main.UserController.PostCreate","ip":"192.168.1.100","level":"info","msg":"创建用户请求","time":"2025-07-29T22:55:03+08:00","user_id":123}
```

### 使用方法
```go
// 基本日志
config.Info("这是信息日志")
config.Warn("这是警告日志")

// 格式化日志
config.Infof("用户 %s 执行了 %s 操作", "张三", "登录")

// 结构化日志
config.WithFields(map[string]any{
    "user_id": 123,
    "action":  "login",
    "ip":      "192.168.1.100",
}).Info("用户登录成功")
```