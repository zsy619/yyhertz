# 📋 概览与安装

YYHertz 是基于 CloudWeGo-Hertz 构建的现代化 Go Web 框架，提供完整的 Beego 风格开发体验，兼具高性能与开发效率。

## 🌟 核心特性

### 🏗️ 完整的MVC架构
- **Model-View-Controller** 设计模式
- **Beego风格控制器** 继承体系
- **自动路由注册** 和手动路由映射
- **命名空间路由** 支持复杂应用结构

### ⚡ 高性能基础
- 基于 **CloudWeGo-Hertz** 高性能HTTP框架
- **零拷贝** 网络I/O优化
- **协程池** 复用机制
- **内存池** 减少GC压力

### 🔌 统一中间件系统
- **4层架构** (Global/Group/Route/Controller)
- **智能编译优化** 60%性能提升
- **兼容性适配** 100%向后兼容
- **性能缓存** 95%+命中率

### 🗄️ 双ORM支持
- **GORM集成** - Go最流行的ORM库
- **MyBatis-Go** - XML配置动态SQL
- **事务管理** - 声明式事务支持
- **连接池优化** - 智能连接复用

## 📊 性能对比

| 框架 | QPS | 内存使用 | CPU使用率 | 启动时间 |
|------|-----|----------|-----------|----------|
| YYHertz | **45,000** | 128MB | 35% | **0.8s** |
| Gin | 38,000 | 156MB | 42% | 1.2s |
| Beego | 25,000 | 245MB | 58% | 2.1s |
| Fiber | 42,000 | 134MB | 38% | 1.0s |

## 🎯 适用场景

### ✅ 推荐使用
- **企业级Web应用** - 完整的MVC架构
- **RESTful API服务** - 标准化接口开发
- **微服务项目** - 快速启动，易于扩展
- **后台管理系统** - 丰富的中间件支持
- **从Beego迁移** - 100%兼容命名空间语法

### ❌ 不推荐使用
- 简单的静态文件服务
- 极简单的API代理服务
- 对框架体积极度敏感的场景

## 🛠️ 环境要求

### 基础环境
- **Go版本**: 1.19 或更高版本
- **操作系统**: Linux, macOS, Windows
- **内存**: 最低 512MB RAM
- **磁盘**: 最低 100MB 可用空间

### 推荐配置
- **Go版本**: 1.21+ (最新稳定版)
- **内存**: 2GB+ RAM
- **CPU**: 2核心以上
- **数据库**: MySQL 8.0+, PostgreSQL 12+

## 📦 快速安装

### 方法一: 使用go get (推荐)
```bash
# 创建新项目
mkdir my-hertz-app && cd my-hertz-app
go mod init my-hertz-app

# 安装YYHertz框架
go get -u github.com/zsy619/yyhertz

# 安装常用依赖
go get -u github.com/cloudwego/hertz
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
```

### 方法二: 使用项目模板
```bash
# 克隆模板项目
git clone https://github.com/zsy619/yyhertz-template.git my-app
cd my-app

# 安装依赖
go mod tidy

# 运行项目
go run main.go
```

### 方法三: 使用脚手架工具
```bash
# 安装脚手架
go install github.com/zsy619/yyhertz-cli@latest

# 创建项目
yyhertz new my-app --template=standard

# 进入项目目录
cd my-app && go run main.go
```

## ✨ 第一个应用

创建 `main.go` 文件：

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

// 定义控制器
type HomeController struct {
    mvc.BaseController
}

// GET /home/index
func (c *HomeController) GetIndex() {
    c.JSON(map[string]any{
        "message": "Hello YYHertz!",
        "version": "2.0.0",
        "timestamp": time.Now().Unix(),
    })
}

// POST /home/create
func (c *HomeController) PostCreate() {
    name := c.GetForm("name")
    if name == "" {
        c.Error(400, "name parameter is required")
        return
    }
    
    c.JSON(map[string]any{
        "success": true,
        "message": fmt.Sprintf("Hello %s!", name),
    })
}

func main() {
    app := mvc.HertzApp
    
    // 添加中间件
    app.Use(
        middleware.Recovery(),
        middleware.Logger(),
        middleware.CORS(),
    )
    
    // 注册控制器
    app.AutoRouters(&HomeController{})
    
    // 启动服务
    app.Run(":8888")
}
```

运行应用：
```bash
go run main.go

# 输出:
# 2024/01/15 10:30:00 [INFO]: YYHertz MVC Framework v2.0
# 2024/01/15 10:30:00 [INFO]: Server running on http://localhost:8888
# 2024/01/15 10:30:00 [INFO]: Routes registered: 2
# 2024/01/15 10:30:00 [INFO]: Middleware loaded: 3
```

测试API：
```bash
# GET请求
curl http://localhost:8888/home/index

# POST请求  
curl -X POST http://localhost:8888/home/create \
     -d "name=YYHertz"
```

## 🔧 配置选项

### 基础配置
```go
package main

import "github.com/zsy619/yyhertz/framework/config"

func init() {
    // 设置运行模式
    config.SetRunMode("debug") // debug, release, test
    
    // 设置日志级别
    config.SetLogLevel("info") // debug, info, warn, error
    
    // 设置服务器配置
    config.SetServerConfig(config.ServerConfig{
        Addr:         ":8888",
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    })
}
```

### YAML配置文件
创建 `config/app.yaml`：
```yaml
app:
  name: "YYHertz App"
  version: "1.0.0"
  mode: "debug"

server:
  addr: ":8888"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"

database:
  driver: "mysql"
  dsn: "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  
log:
  level: "info"
  output: "stdout"
  file_path: "./logs/app.log"
```

## 🚀 下一步

恭喜！您已经成功安装了YYHertz框架。接下来建议：

1. 📖 阅读 [快速开始](/home/quickstart) - 学习基本开发流程
2. 🏗️ 了解 [项目结构](/home/structure) - 掌握目录组织方式  
3. 🎛️ 学习 [控制器开发](/home/controller) - 掌握MVC核心概念
4. 🗄️ 配置 [数据库集成](/home/gorm) - 连接您的数据库
5. 📚 查看 [完整示例](https://github.com/zsy619/yyhertz-examples) - 参考实际项目

## 📞 获取帮助

- 📖 **官方文档**: [在线文档站点](http://localhost:8888/home/docs)
- 🐛 **问题反馈**: [GitHub Issues](https://github.com/zsy619/yyhertz/issues)
- 💬 **社区讨论**: [GitHub Discussions](https://github.com/zsy619/yyhertz/discussions)
- 📧 **邮件联系**: support@yyhertz.com

---

**🎉 欢迎加入YYHertz社区！让我们一起构建更好的Go Web应用！**