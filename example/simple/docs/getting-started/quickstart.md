# 🚀 快速开始

通过这个10分钟教程，您将掌握YYHertz框架的核心概念，并创建一个完整的Web应用。

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

## 🎯 第一个应用 - Hello World

### 1. 创建主文件

创建 `main.go` 文件：

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/example/simple/controllers"
)

func main() {
    // 创建应用实例
    app := mvc.HertzApp
    
    // 创建控制器
    homeController := &controllers.HomeController{}
    
    // 注册路由
    app.RouterPrefix("/", homeController, "GetIndex", "GET:/")
    
    // 启动服务器
    app.Run(":8080")
}
```

### 2. 创建控制器

创建 `controllers/home_controller.go`：

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
)

type HomeController struct {
    mvc.BaseController
}

func (c *HomeController) GetIndex() {
    c.SetData("Title", "欢迎使用 Hertz MVC")
    c.SetData("Message", "Hello, World!")
    c.RenderHTML("home/index.html")
}
```

### 3. 创建视图模板

创建 `views/home/index.html`：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            padding: 40px; 
            text-align: center; 
        }
        h1 { color: #667eea; }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>{{.Message}}</p>
</body>
</html>
```

### 4. 运行应用

```bash
go run main.go
```

访问 http://localhost:8080 查看结果！

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

## 下一步

现在你已经有了一个基本的Hertz MVC应用，可以继续学习：

- [控制器详解](/docs/mvc-core/controller) - 了解控制器的高级用法
- [路由系统](/docs/mvc-core/routing) - 学习路由配置和参数处理
- [模板引擎](/docs/view-template/overview) - 掌握模板语法和布局
- [中间件](/docs/middleware/overview) - 添加认证、日志等功能
- [数据库集成](/docs/data-access/orm-unified) - 连接和操作数据库

## 常见问题

### Q: 端口被占用怎么办？

A: 修改 `main.go` 中的端口号或配置文件中的端口设置。

### Q: 模板文件找不到？

A: 确保视图文件路径与 `RenderHTML` 中指定的路径一致。

### Q: 如何处理静态文件？

A: 静态文件会自动从 `static/` 目录提供服务，访问路径为 `/static/文件路径`。

---

**恭喜！** 🎉 你已经成功创建了第一个Hertz MVC应用！