# Hertz MVC Framework

基于CloudWeGo-Hertz的类Beego框架，提供简洁、高效的Go Web开发体验。

## 🚀 特性

- **基于Controller的架构** - 类似Beego的Controller结构，开发更简单
- **HTML模板支持** - 内置模板引擎，支持布局和组件化开发  
- **中间件机制** - 丰富的中间件支持，包括认证、日志、限流等
- **RESTful路由** - 支持RESTful风格的路由设计，API开发更规范
- **高性能** - 基于CloudWeGo-Hertz，提供卓越的性能表现
- **易扩展** - 模块化设计，易于扩展和定制

## 📦 项目结构

```
hertz-mvc/
├── framework/              # 框架核心代码
│   ├── controller/         # 控制器核心包
│   │   ├── base_controller.go  # 基础控制器实现
│   │   └── router.go           # 路由注册机制
│   └── middleware/         # 中间件包
│       └── middleware.go       # 中间件实现
├── example/                # 示例应用
│   ├── controllers/        # 示例控制器
│   │   ├── home_controller.go
│   │   ├── user_controller.go
│   │   └── admin_controller.go
│   ├── views/              # 模板文件
│   │   ├── layout/
│   │   │   └── layout.html     # 布局模板
│   │   ├── user/
│   │   │   ├── index.html      # 用户列表
│   │   │   └── info.html       # 用户详情
│   │   └── admin/
│   │       └── dashboard.html  # 管理面板
│   ├── static/             # 静态资源
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│   └── main.go             # 示例入口文件
├── go.mod
└── README.md
```

## 🛠️ 安装

1. 克隆项目：
```bash
git clone <repository-url>
cd hertz-mvc
```

2. 安装依赖：
```bash
go mod tidy
```

3. 运行示例：
```bash
cd example
go run main.go
```

4. 访问应用：
```
http://localhost:8888
```

## 📚 快速开始

### 1. 创建控制器

```go
package controllers

import (
    "hertz-mvc/framework/controller"
)

type HomeController struct {
    controller.BaseController
}

func (c *HomeController) GetIndex() {
    c.SetData("Title", "欢迎")
    c.SetData("Message", "Hello World!")
    c.Render("home/index.html")
}

func (c *HomeController) PostCreate() {
    name := c.GetForm("name")
    c.JSON(map[string]any{
        "success": true,
        "message": "创建成功",
        "name":    name,
    })
}
```

### 2. 注册路由

```go
package main

import (
    "./controllers"
    "./framework/controller"
    "./framework/middleware"
)

func main() {
    app := controller.NewApp()
    
    // 添加中间件
    app.Use(
        middleware.LoggerMiddleware(),
        middleware.CORSMiddleware(),
    )
    
    // 注册控制器
    homeController := &controllers.HomeController{}
    app.Include(homeController)
    
    // 自定义路由
    app.Router("/api", homeController,
        "GetProfile", "GET:/api/profile",
        "PostLogin", "POST:/api/login",
    )
    
    app.Run(":8888")
}
```

### 3. 创建模板

**布局模板** (`views/layout/layout.html`):
```html
{{define "layout"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    {{template "content" .}}
</body>
</html>
{{end}}
```

**页面模板** (`views/home/index.html`):
```html
{{define "content"}}
<h1>{{.Title}}</h1>
<p>{{.Message}}</p>
{{end}}
```

## 🔧 核心概念

### 控制器 (Controller)

控制器是处理HTTP请求的核心组件：

```go
type UserController struct {
    controller.BaseController
}

// GET方法自动映射到GET请求
func (c *UserController) GetIndex() {
    // 处理GET /user/index
}

// POST方法自动映射到POST请求  
func (c *UserController) PostCreate() {
    // 处理POST /user/create
}
```

### 路由 (Routing)

支持多种路由注册方式：

```go
// 自动路由 - 根据方法名自动生成路由
app.Include(userController)

// 手动路由 - 自定义路由规则
app.Router("/user", userController,
    "GetProfile", "GET:/user/profile",
    "PostUpdate", "PUT:/user/update",
)
```

### 中间件 (Middleware)

内置多种中间件：

```go
app.Use(
    middleware.RecoveryMiddleware(),    // 异常恢复
    middleware.LoggerMiddleware(),     // 请求日志
    middleware.CORSMiddleware(),       // 跨域支持
    middleware.RateLimitMiddleware(100, time.Minute), // 限流
    middleware.AuthMiddleware("/login"), // 认证
)
```

### 模板 (Templates)

支持布局和组件化：

```go
// 使用布局渲染
c.Render("user/index.html")

// 不使用布局
c.RenderHTML("user/simple.html") 

// 设置数据
c.SetData("key", "value")
```

## 📖 API 参考

### BaseController 方法

| 方法 | 说明 |
|------|------|
| `JSON(data)` | 返回JSON响应 |
| `String(text)` | 返回文本响应 |
| `Render(view)` | 渲染模板(带布局) |
| `RenderHTML(view)` | 渲染模板(无布局) |
| `Redirect(url)` | 重定向 |
| `Error(code, msg)` | 返回错误响应 |
| `SetData(key, value)` | 设置模板数据 |
| `GetString(key)` | 获取查询参数 |
| `GetInt(key)` | 获取整型参数 |
| `GetForm(key)` | 获取表单数据 |

### 中间件

| 中间件 | 说明 |
|--------|------|
| `LoggerMiddleware()` | 请求日志记录 |
| `CORSMiddleware()` | 跨域请求支持 |
| `AuthMiddleware(skip...)` | 身份认证 |
| `RecoveryMiddleware()` | 异常恢复 |
| `RateLimitMiddleware(max, duration)` | 请求限流 |

## 🌟 示例

查看 `example` 目录获取完整示例，包括：

- **首页展示** - 框架特性介绍
- **用户管理** - CRUD操作示例
- **管理后台** - 权限控制示例
- **RESTful API** - API接口示例

## 📋 路由列表

### 页面路由
- `GET /` - 首页
- `GET /about` - 关于页面  
- `GET /docs` - 文档页面
- `GET /user/index` - 用户列表
- `GET /user/info` - 用户详情
- `GET /admin/dashboard` - 管理面板

### API路由
- `POST /user/create` - 创建用户
- `PUT /user/update` - 更新用户
- `DELETE /user/remove` - 删除用户
- `GET /admin/users` - 管理员获取用户列表
- `POST /admin/settings` - 保存系统设置

## 🧪 测试

```bash
# 测试首页
curl http://localhost:8888/

# 测试用户列表
curl http://localhost:8888/user/index

# 测试创建用户
curl -X POST http://localhost:8888/user/create \
  -d "name=张三&email=test@example.com&password=123456"

# 测试管理员接口(需要认证)
curl -H "Authorization: Bearer admin-token" \
  http://localhost:8888/admin/dashboard
```

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📄 许可证

本项目采用Apache 2.0许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [CloudWeGo-Hertz](https://github.com/cloudwego/hertz)
- [Hertz 文档](https://www.cloudwego.io/zh/docs/hertz/)
- [Beego 框架](https://github.com/beego/beego)

---

**Hertz MVC Framework** - 让Go Web开发更简单！