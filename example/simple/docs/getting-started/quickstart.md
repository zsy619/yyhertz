# 🚀 快速开始

通过这个15分钟教程，您将掌握YYHertz框架的核心概念，包括最新的多Handler类型系统，并创建一个完整的现代Web应用。

## 🛠️ 环境准备

### 系统要求
- **Go版本**: 1.19+ (推荐1.21+)
- **操作系统**: Linux, macOS, Windows
- **内存**: 最低512MB RAM
- **工具**: Git, IDE (推荐VS Code + Go插件)

### 验证环境
```bash
# 检查Go版本
go version
# 输出: go version go1.21.0 darwin/amd64

# 检查Go环境
go env GOPATH GOROOT
```

## 📦 创建项目

### 方式一：标准创建 (推荐)
```bash
# 1. 创建项目目录
mkdir my-hertz-app && cd my-hertz-app

# 2. 初始化Go模块
go mod init my-hertz-app

# 3. 安装YYHertz框架
go get -u github.com/zsy619/yyhertz

# 4. 验证安装
go list -m github.com/zsy619/yyhertz
```

### 方式二：使用模板
```bash
# 克隆官方模板
git clone https://github.com/zsy619/yyhertz-template.git my-app
cd my-app

# 重新初始化模块
rm go.mod go.sum
go mod init my-app
go mod tidy
```

## 🎯 第一个应用 - 多Handler类型示例

### 1. 创建主文件

创建 `main.go` 文件，展示YYHertz最新的多Handler类型系统：

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/router"
    mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
)

func main() {
    // 创建应用实例
    app := mvc.HertzApp
    
    // 创建API路由组
    apiGroup := mvc.CreateGroup("/api/v1")
    
    // 设置各种Handler类型示例
    setupHandlers(apiGroup)
    
    log.Println("🚀 多Handler类型应用启动成功!")
    log.Println("📍 访问: http://localhost:8080")
    
    // 启动服务器
    app.Run(":8080")
}

func setupHandlers(group *router.Group) {
    // 1. 轻量级处理器 - 健康检查
    group.GETLight("/health", func() {
        log.Println("🟢 Health check performed")
    })
    
    // 2. 简单处理器 - Ping响应
    group.GETSimple("/ping", func(ctx context.Context) {
        log.Println("📡 Ping received")
    })
    
    // 3. 直接处理器 - 系统信息
    group.GETDirect("/info", func(c *mvcContext.Context) {
        c.Set("handler_type", "DirectHandler")
        reqCtx := c.RequestContext()
        reqCtx.SetContentType("application/json")
        
        info := fmt.Sprintf(`{
            "message": "YYHertz 框架",
            "handler": "DirectHandler (增强Context)",
            "timestamp": "%s",
            "keys_count": %d
        }`, time.Now().Format(time.RFC3339), c.KeysCount())
        
        reqCtx.WriteString(info)
    })
    
    // 4. 响应处理器 - 用户数据
    group.GETResponse("/users", func(c *mvcContext.Context) any {
        c.Set("handler_type", "ResponseHandler")
        return map[string]any{
            "success": true,
            "users": []map[string]any{
                {"id": 1, "name": "Alice", "role": "admin"},
                {"id": 2, "name": "Bob", "role": "user"},
            },
            "total": 2,
            "enhanced_context": true,
        }
    })
    
    // 5. 异步处理器 - 模拟数据处理
    group.POSTAsync("/process", func(c *mvcContext.Context) <-chan any {
        c.Set("handler_type", "AsyncHandler")
        resultChan := make(chan any, 1)
        
        go func() {
            defer close(resultChan)
            time.Sleep(1 * time.Second) // 模拟处理时间
            resultChan <- map[string]any{
                "success": true,
                "message": "异步处理完成",
                "processed_at": time.Now().Format(time.RFC3339),
                "context_enhanced": true,
            }
        }()
        
        return resultChan
    })
}
```

### 2. 运行应用

```bash
# 安装依赖
go mod tidy

# 启动应用
go run main.go
```

### 3. 测试Handler类型

应用启动后，可以测试各种Handler类型：

```bash
# 1. 轻量级处理器 - 健康检查
curl http://localhost:8080/api/v1/health
# 返回: HTTP 200 (无响应体，最小开销)

# 2. 简单处理器 - Ping
curl http://localhost:8080/api/v1/ping
# 返回: HTTP 200 (自动响应)

# 3. 直接处理器 - 系统信息
curl http://localhost:8080/api/v1/info
# 返回: {"message":"YYHertz 框架","handler":"DirectHandler (增强Context)","timestamp":"...","keys_count":1}

# 4. 响应处理器 - 用户数据
curl http://localhost:8080/api/v1/users
# 返回: {"success":true,"users":[{"id":1,"name":"Alice","role":"admin"},{"id":2,"name":"Bob","role":"user"}],"total":2,"enhanced_context":true}

# 5. 异步处理器 - 数据处理
curl -X POST http://localhost:8080/api/v1/process
# 返回: {"success":true,"message":"异步处理完成","processed_at":"...","context_enhanced":true}
```

### 4. 理解Handler类型

YYHertz框架提供7种Handler类型，每种都有特定的用途：

| Handler类型 | 签名 | 适用场景 | 特点 |
|------------|------|----------|------|
| **LightHandler** | `func()` | 健康检查、静态响应 | 零参数，最小开销 |
| **SimpleHandler** | `func(context.Context)` | 简单业务逻辑 | 支持上下文取消 |
| **DirectHandler** | `func(*mvcContext.Context)` | 直接控制响应 | 增强Context，高性能 |
| **ResponseHandler** | `func(*mvcContext.Context) any` | REST API | 自动JSON序列化 |
| **AsyncHandler** | `func(*mvcContext.Context) <-chan any` | 异步处理 | 支持超时控制 |
| **StreamHandler** | `func(*mvcContext.Context, chan<- []byte) error` | 流式传输 | 实时数据流 |
| **HandlerFunc** | `func(context.Context, *RequestContext)` | 传统方式 | 向后兼容 |

## 项目结构

标准的Hertz MVC项目结构：

```
my-hertz-app/
├── controllers/          # 控制器目录
│   └── home_controller.go
├── views/               # 视图模板目录
│   └── home/
│       └── index.html
├── static/              # 静态资源目录
│   ├── css/
│   ├── js/
│   └── images/
├── conf/               # 配置文件目录
├── models/             # 数据模型目录
├── middleware/         # 中间件目录
├── main.go            # 应用入口文件
└── go.mod             # Go模块文件
```

## 配置说明

### 基本配置

创建 `conf/app.yaml`：

```yaml
app:
  name: "my-hertz-app"
  version: "1.0.0"
  debug: true
  port: 8080
  host: "0.0.0.0"

log:
  level: "info"
  format: "json"
  enable_console: true
```

### 环境变量

支持通过环境变量覆盖配置：

```bash
export HERTZ_PORT=9000
export HERTZ_DEBUG=false
go run main.go
```

## 🚀 下一步

现在你已经掌握了YYHertz的多Handler类型系统，可以继续深入学习：

### 📚 核心概念
- [**多Handler类型详解**](/docs/mvc-core/routing) - 深入了解7种Handler类型的应用场景
- [**增强Context系统**](/docs/mvc-core/controller) - 掌握高性能并发Context的使用
- [**路由系统升级**](/docs/mvc-core/routing) - 学习新的路由注册方式和参数处理

### 🛠️ 高级功能  
- [**统一ORM解决方案**](/docs/data-access/orm-unified) - GORM + MyBatis双引擎架构
- [**中间件系统**](/docs/middleware/overview) - 4层架构的中间件设计
- [**模板引擎**](/docs/view-template/overview) - 灵活的视图渲染系统

### ⚡ 性能优化
- [**对象池化**](/docs/dev-tools/performance) - 了解Context池化机制
- [**并发优化**](/docs/advanced/scheduler) - 掌握高并发编程技巧
- [**监控告警**](/docs/data-access/monitoring-alerting) - 性能监控和优化建议

### 🔧 实战项目
- [**完整示例**](https://github.com/zsy619/yyhertz-examples) - 查看真实项目案例
- [**最佳实践**](/docs/mvc-core/controller/faq-best-practices) - 学习开发规范和技巧

## 常见问题

### Q: 端口被占用怎么办？

A: 修改 `main.go` 中的端口号或配置文件中的端口设置。

### Q: 模板文件找不到？

A: 确保视图文件路径与 `RenderHTML` 中指定的路径一致。

### Q: 如何处理静态文件？

A: 静态文件可以通过 `SetStaticPath` 方法配置。框架默认从 `static/` 目录提供服务，访问路径为 `/static/文件路径`。

**自定义静态路径：**
```go
// 双参数形式：指定本地目录和URL路径
mvc.SetStaticPath("public", "/assets")

// 单参数形式：自动推导URL路径
mvc.SetStaticPath("uploads")  // 自动映射到 /uploads
```

---

**恭喜！** 🎉 你已经成功创建了第一个Hertz MVC应用！