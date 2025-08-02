# YYHertz MVC Framework

基于CloudWeGo-Hertz的现代化Go Web框架，提供完整的Beego风格开发体验，兼具高性能与开发效率。

## 🚀 核心特性

- **🏗️ MVC架构** - 完整的Model-View-Controller设计模式
- **📁 Beego风格Namespace** - 100%兼容Beego的命名空间路由系统
- **🎛️ 智能路由** - 自动路由注册 + 手动路由映射，支持RESTful设计
- **🎨 模板引擎** - 内置HTML模板支持，布局和组件化开发
- **🔌 中间件生态** - 丰富的中间件：认证、日志、限流、CORS、恢复等
- **⚡ 高性能** - 基于CloudWeGo-Hertz，提供卓越的性能表现
- **🔧 配置管理** - 基于Viper的配置系统，支持多种格式
- **📊 可观测性** - 内置日志、链路追踪、监控指标
- **🛡️ 生产就绪** - 完善的错误处理、优雅关闭、健康检查

## 📦 项目结构

```
YYHertz/
├── framework/                    # 🏗️ 框架核心
│   ├── mvc/                     # MVC核心组件
│   │   ├── core/               # 核心应用和控制器
│   │   │   ├── app.go          # 应用实例
│   │   │   └── controller.go   # 基础控制器
│   │   ├── router/             # 路由系统
│   │   │   └── group.go        # 路由组管理
│   │   ├── namespace.go        # 🆕 Beego风格命名空间
│   │   ├── controller.go       # 控制器接口
│   │   └── static.go           # 静态方法导出
│   ├── middleware/             # 🔌 中间件集合
│   │   ├── auth.go            # 身份认证中间件
│   │   ├── cors.go            # 跨域中间件
│   │   ├── logger.go          # 日志中间件
│   │   ├── recovery.go        # 恢复中间件
│   │   └── rate_limit.go      # 限流中间件
│   ├── config/                 # ⚙️ 配置管理
│   │   ├── viper_config.go    # Viper配置实现
│   │   └── logger_singleton.go # 日志单例
│   ├── validation/             # ✅ 数据验证
│   ├── i18n/                   # 🌍 国际化支持
│   ├── view/                   # 🎨 视图引擎
│   └── testing/                # 🧪 测试工具
├── example/                     # 📚 完整示例
│   ├── controllers/            # 示例控制器
│   ├── views/                  # 模板文件
│   ├── static/                 # 静态资源
│   └── main.go                # 示例入口
├── config/                     # 📋 配置文件
│   └── config.yaml            # 应用配置
├── go.mod                      # Go模块定义
└── README.md                   # 📖 项目文档
```

## 🛠️ 快速开始

### 1. 安装框架

```bash
# 克隆项目
git clone <repository-url>
cd YYHertz

# 安装依赖
go mod tidy

# 运行示例
go run example/main.go

# 访问应用
open http://localhost:8888
```

### 2. 创建第一个应用

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/middleware"
)

type HomeController struct {
    mvc.BaseController
}

func (c *HomeController) GetIndex() {
    c.JSON(map[string]any{
        "message": "Hello YYHertz!",
        "version": "1.0.0",
    })
}

func main() {
    app := mvc.NewApp()
    
    // 添加中间件
    app.Use(
        middleware.RecoveryMiddleware(),
        middleware.LoggerMiddleware(),
        middleware.CORSMiddleware(),
    )
    
    // 注册控制器
    app.AutoRouter(&HomeController{})
    
    app.Run(":8888")
}
```

## 📚 核心功能

### 🏗️ 控制器开发

YYHertz采用标准的MVC架构，控制器是处理请求的核心：

```go
type UserController struct {
    mvc.BaseController
}

// GET方法自动映射到GET请求
func (c *UserController) GetIndex() {
    users := []User{{ID: 1, Name: "张三"}}
    c.SetData("users", users)
    c.Render("user/index.html")
}

// POST方法自动映射到POST请求  
func (c *UserController) PostCreate() {
    name := c.GetForm("name")
    email := c.GetForm("email")
    
    // 业务逻辑处理
    user := CreateUser(name, email)
    
    c.JSON(map[string]any{
        "success": true,
        "user": user,
    })
}

// 支持任意HTTP方法
func (c *UserController) PutUpdate() {
    // 处理PUT请求
}

func (c *UserController) DeleteRemove() {
    // 处理DELETE请求
}
```

### 📁 Beego风格命名空间 🆕

YYHertz完全兼容Beego的Namespace语法，支持复杂的路由组织：

```go
// 创建API命名空间
nsApi := mvc.NewNamespace("/api",
    // 自动路由注册
    mvc.NSAutoRouter(&PageController{}),
    
    // 手动路由映射
    mvc.NSRouter("/auth/token", &AuthController{}, "*:GetToken"),
    mvc.NSRouter("/auth/refresh", &AuthController{}, "POST:RefreshToken"),
    
    // 嵌套命名空间
    mvc.NSNamespace("/user",
        mvc.NSRouter("/profile", &UserController{}, "GET:GetProfile"),
        mvc.NSRouter("/settings", &UserController{}, "PUT:UpdateSettings"),
        
        // 多层嵌套
        mvc.NSNamespace("/social",
            mvc.NSRouter("/friends", &SocialController{}, "GET:GetFriends"),
            mvc.NSRouter("/messages", &SocialController{}, "POST:SendMessage"),
        ),
    ),
    
    // 管理功能命名空间
    mvc.NSNamespace("/admin",
        mvc.NSAutoRouter(&AdminController{}),
        mvc.NSNamespace("/system",
            mvc.NSRouter("/config", &SystemController{}, "GET:GetConfig"),
            mvc.NSRouter("/logs", &SystemController{}, "GET:GetLogs"),
        ),
    ),
)

// 添加到全局应用
mvc.AddNamespace(nsApi)
```

**支持的路由方法格式**：
- `"*:MethodName"` - 支持所有HTTP方法
- `"GET:MethodName"` - 仅支持GET方法
- `"POST:MethodName"` - 仅支持POST方法
- `"PUT:MethodName"` - 仅支持PUT方法
- `"DELETE:MethodName"` - 仅支持DELETE方法

### 🎛️ 智能路由系统

YYHertz提供多种路由注册方式，满足不同开发需求：

```go
app := mvc.NewApp()

// 1. 自动路由 - 根据控制器方法名自动生成路由
app.AutoRouter(&UserController{})
// 生成路由：GET /user/index, POST /user/create 等

// 2. 手动路由 - 完全自定义路由规则
app.Router(&UserController{},
    "GetProfile", "GET:/user/profile",
    "PostUpdate", "PUT:/user/:id/update",
    "DeleteUser", "DELETE:/user/:id",
)

// 3. 带前缀的路由组
app.RouterPrefix("/api/v1", &ApiController{},
    "GetUsers", "GET:/users",
    "CreateUser", "POST:/users",
)

// 4. 混合使用
app.AutoRouter(&HomeController{})           // 自动路由
app.Router(&ApiController{}, ...)          // 手动路由
mvc.AddNamespace(nsApi)                    // 命名空间路由
```

### 🔌 中间件生态

内置丰富的中间件，开箱即用：

```go
import "github.com/zsy619/yyhertz/framework/middleware"

app.Use(
    // 🛡️ 异常恢复
    middleware.RecoveryMiddleware(),
    
    // 📋 请求日志
    middleware.LoggerMiddleware(),
    
    // 🌐 跨域支持
    middleware.CORSMiddleware(),
    
    // 🚦 请求限流 (100请求/分钟)
    middleware.RateLimitMiddleware(100, time.Minute),
    
    // 🔐 身份认证 (跳过指定路径)
    middleware.AuthMiddleware("/login", "/register"),
    
    // 📊 链路追踪
    middleware.TracingMiddleware(),
)
```

### 🎨 模板引擎

支持布局和组件化的模板开发：

```go
// 控制器中使用模板
func (c *UserController) GetIndex() {
    c.SetData("title", "用户管理")
    c.SetData("users", getUserList())
    
    // 使用布局渲染
    c.Render("user/index.html")
    
    // 或不使用布局
    c.RenderHTML("user/simple.html")
}
```

**布局模板** (`views/layout/layout.html`):
```html
{{define "layout"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
    <link rel="stylesheet" href="/static/css/app.css">
</head>
<body>
    <nav>{{template "nav" .}}</nav>
    <main>{{template "content" .}}</main>
    <footer>{{template "footer" .}}</footer>
</body>
</html>
{{end}}
```

**页面模板** (`views/user/index.html`):
```html
{{define "content"}}
<div class="user-list">
    <h1>{{.title}}</h1>
    <table>
        {{range .users}}
        <tr>
            <td>{{.ID}}</td>
            <td>{{.Name}}</td>
            <td>{{.Email}}</td>
        </tr>
        {{end}}
    </table>
</div>
{{end}}
```

## 📖 API 参考

### BaseController 核心方法

| 方法 | 说明 | 示例 |
|------|------|------|
| **响应方法** |
| `JSON(data)` | 返回JSON响应 | `c.JSON(map[string]any{"code": 200})` |
| `String(text)` | 返回纯文本响应 | `c.String("Hello World")` |
| `Render(view)` | 渲染模板(带布局) | `c.Render("user/index.html")` |
| `RenderHTML(view)` | 渲染模板(无布局) | `c.RenderHTML("simple.html")` |
| `Redirect(url)` | 重定向 | `c.Redirect("/login")` |
| `Error(code, msg)` | 返回错误响应 | `c.Error(404, "Not Found")` |
| **数据处理** |
| `SetData(key, value)` | 设置模板数据 | `c.SetData("user", userObj)` |
| `GetString(key, def...)` | 获取字符串参数 | `name := c.GetString("name", "默认值")` |
| `GetInt(key, def...)` | 获取整型参数 | `age := c.GetInt("age", 0)` |
| `GetForm(key)` | 获取表单数据 | `email := c.GetForm("email")` |
| `GetJSON()` | 获取JSON数据 | `data := c.GetJSON()` |
| **文件处理** |
| `GetFile(key)` | 获取上传文件 | `file := c.GetFile("avatar")` |
| `SaveFile(file, path)` | 保存文件 | `c.SaveFile(file, "./uploads/")` |

### Namespace API

| 函数 | 说明 | 示例 |
|------|------|------|
| `NewNamespace(prefix, ...funcs)` | 创建命名空间 | `ns := mvc.NewNamespace("/api", ...)` |
| `NSAutoRouter(controller)` | 自动路由注册 | `mvc.NSAutoRouter(&UserController{})` |
| `NSRouter(path, ctrl, method)` | 手动路由映射 | `mvc.NSRouter("/users", ctrl, "GET:GetUsers")` |
| `NSNamespace(prefix, ...funcs)` | 嵌套命名空间 | `mvc.NSNamespace("/v1", ...)` |
| `AddNamespace(ns)` | 全局注册命名空间 | `mvc.AddNamespace(ns)` |

### 中间件

| 中间件 | 说明 | 参数 |
|--------|------|------|
| `RecoveryMiddleware()` | 异常恢复 | 无 |
| `LoggerMiddleware()` | 请求日志 | 无 |
| `CORSMiddleware()` | 跨域支持 | 无 |
| `AuthMiddleware(skip...)` | 身份认证 | 跳过的路径列表 |
| `RateLimitMiddleware(max, duration)` | 请求限流 | 最大请求数, 时间窗口 |
| `TracingMiddleware()` | 链路追踪 | 无 |

## 🌟 完整示例

### 电商API示例

```go
package main

import (
    "time"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/middleware"
)

// 产品控制器
type ProductController struct {
    mvc.BaseController
}

func (c *ProductController) GetList() {
    c.JSON(map[string]any{
        "products": []map[string]any{
            {"id": 1, "name": "iPhone 15", "price": 7999},
            {"id": 2, "name": "MacBook Pro", "price": 14999},
        },
    })
}

func (c *ProductController) PostCreate() {
    name := c.GetForm("name")
    price := c.GetInt("price")
    
    // 业务逻辑...
    
    c.JSON(map[string]any{
        "success": true,
        "product": map[string]any{
            "name": name,
            "price": price,
        },
    })
}

// 订单控制器
type OrderController struct {
    mvc.BaseController
}

func (c *OrderController) GetList() {
    userID := c.GetInt("user_id")
    // 获取用户订单...
    c.JSON(map[string]any{"orders": []any{}})
}

func (c *OrderController) PostCreate() {
    // 创建订单逻辑...
    c.JSON(map[string]any{"success": true})
}

// 用户控制器
type UserController struct {
    mvc.BaseController
}

func (c *UserController) GetProfile() {
    c.JSON(map[string]any{
        "user": map[string]any{
            "id": 1,
            "name": "张三",
            "email": "zhangsan@example.com",
        },
    })
}

func main() {
    app := mvc.NewApp()
    
    // 全局中间件
    app.Use(
        middleware.RecoveryMiddleware(),
        middleware.LoggerMiddleware(),
        middleware.CORSMiddleware(),
        middleware.RateLimitMiddleware(1000, time.Minute),
    )
    
    // 创建API命名空间
    apiV1 := mvc.NewNamespace("/api/v1",
        // 产品管理
        mvc.NSNamespace("/products",
            mvc.NSRouter("/list", &ProductController{}, "GET:GetList"),
            mvc.NSRouter("/create", &ProductController{}, "POST:PostCreate"),
            mvc.NSRouter("/:id", &ProductController{}, "GET:GetDetail"),
            mvc.NSRouter("/:id", &ProductController{}, "PUT:Update"),
            mvc.NSRouter("/:id", &ProductController{}, "DELETE:Delete"),
        ),
        
        // 订单管理
        mvc.NSNamespace("/orders",
            mvc.NSAutoRouter(&OrderController{}),
        ),
        
        // 用户管理
        mvc.NSNamespace("/users",
            mvc.NSRouter("/profile", &UserController{}, "GET:GetProfile"),
            mvc.NSRouter("/settings", &UserController{}, "PUT:UpdateSettings"),
        ),
    )
    
    // 管理员API
    adminAPI := mvc.NewNamespace("/admin",
        middleware.AuthMiddleware(), // 需要认证
        mvc.NSNamespace("/system",
            mvc.NSRouter("/stats", &AdminController{}, "GET:GetStats"),
            mvc.NSRouter("/config", &AdminController{}, "GET:GetConfig"),
        ),
    )
    
    // 注册命名空间
    mvc.AddNamespace(apiV1)
    mvc.AddNamespace(adminAPI)
    
    // 启动服务
    app.Run(":8888")
}
```

### 生成的路由列表

运行上述示例后，会自动生成以下路由：

#### API V1 路由
- `GET /api/v1/products/list` - 产品列表
- `POST /api/v1/products/create` - 创建产品
- `GET /api/v1/products/:id` - 产品详情
- `PUT /api/v1/products/:id` - 更新产品
- `DELETE /api/v1/products/:id` - 删除产品
- `GET /api/v1/orders/list` - 订单列表 (自动路由)
- `POST /api/v1/orders/create` - 创建订单 (自动路由)
- `GET /api/v1/users/profile` - 用户资料
- `PUT /api/v1/users/settings` - 更新设置

#### 管理员路由
- `GET /admin/system/stats` - 系统统计
- `GET /admin/system/config` - 系统配置

## 🧪 测试示例

```bash
# 获取产品列表
curl http://localhost:8888/api/v1/products/list

# 创建产品
curl -X POST http://localhost:8888/api/v1/products/create \
  -d "name=新产品&price=999"

# 获取用户资料
curl http://localhost:8888/api/v1/users/profile

# 获取订单列表
curl "http://localhost:8888/api/v1/orders/list?user_id=1"

# 管理员接口 (需要认证)
curl -H "Authorization: Bearer admin-token" \
  http://localhost:8888/admin/system/stats
```

## 🏆 性能特性

- **🚀 高并发**: 基于CloudWeGo-Hertz，支持高并发处理
- **💾 低内存**: 优化的内存使用，减少GC压力  
- **⚡ 快速启动**: 秒级启动，适合微服务部署
- **🔄 热重载**: 开发模式支持代码热重载
- **📈 可扩展**: 模块化设计，易于水平扩展

## 🤝 社区与贡献

- **🐛 问题反馈**: [GitHub Issues](https://github.com/your-repo/issues)
- **💡 功能建议**: [GitHub Discussions](https://github.com/your-repo/discussions)  
- **🔀 贡献代码**: 欢迎提交Pull Request
- **📚 文档完善**: 帮助完善文档和示例

### 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 开源协议

本项目采用 **Apache 2.0** 开源协议 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关资源

### 官方文档
- [YYHertz 官方文档](https://docs.yyhertz.com) 
- [API 参考手册](https://docs.yyhertz.com/api)
- [最佳实践指南](https://docs.yyhertz.com/best-practices)

### 技术栈
- [CloudWeGo-Hertz](https://github.com/cloudwego/hertz) - 高性能HTTP框架
- [Beego Framework](https://github.com/beego/beego) - Go Web框架参考
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Logrus](https://github.com/sirupsen/logrus) - 结构化日志

### 示例项目
- [YYHertz-Blog](https://github.com/your-repo/yyhertz-blog) - 博客系统示例
- [YYHertz-Shop](https://github.com/your-repo/yyhertz-shop) - 电商系统示例
- [YYHertz-Admin](https://github.com/your-repo/yyhertz-admin) - 后台管理示例

---

<div align="center">

**🌟 YYHertz MVC Framework**

*让 Go Web 开发更简单、更高效*

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)
[![Stars](https://img.shields.io/github/stars/your-repo/yyhertz?style=social)](https://github.com/your-repo/yyhertz)

</div>