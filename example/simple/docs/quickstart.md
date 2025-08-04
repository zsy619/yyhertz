# 🚀 快速开始

几分钟内搭建你的第一个Hertz MVC应用。

## 环境要求

- Go 1.19 或更高版本
- Git（用于下载依赖）

## 安装框架

使用 go mod 初始化项目并安装Hertz MVC框架：

```bash
mkdir my-hertz-app
cd my-hertz-app
go mod init my-hertz-app
go get github.com/zsy619/yyhertz
```

## 创建第一个应用

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

- [控制器详解](/home/controller) - 了解控制器的高级用法
- [路由系统](/home/routing) - 学习路由配置和参数处理
- [模板引擎](/home/template) - 掌握模板语法和布局
- [中间件](/home/middleware) - 添加认证、日志等功能
- [数据库集成](/home/database) - 连接和操作数据库

## 常见问题

### Q: 端口被占用怎么办？

A: 修改 `main.go` 中的端口号或配置文件中的端口设置。

### Q: 模板文件找不到？

A: 确保视图文件路径与 `RenderHTML` 中指定的路径一致。

### Q: 如何处理静态文件？

A: 静态文件会自动从 `static/` 目录提供服务，访问路径为 `/static/文件路径`。

---

**恭喜！** 🎉 你已经成功创建了第一个Hertz MVC应用！