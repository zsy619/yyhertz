# YYHertz 快速开始指南

<div align="center">

🚀 **5分钟快速上手YYHertz** | 从零到第一个Web应用

</div>

---

## 目录

- [环境要求](#环境要求)
- [快速安装](#快速安装)
- [创建第一个应用](#创建第一个应用)
- [核心概念速览](#核心概念速览)
- [常用功能示例](#常用功能示例)
- [下一步学习](#下一步学习)

---

## 🔧 环境要求

### 必备环境
- **Go版本**: ≥ 1.19 (推荐使用Go 1.21+)
- **操作系统**: Linux、macOS、Windows
- **内存**: 最少512MB，推荐2GB+
- **磁盘**: 至少100MB可用空间

### 验证环境
```bash
# 检查Go版本
go version
# 输出示例: go version go1.21.0 darwin/amd64

# 检查Go模块支持
go env GO111MODULE
# 输出: on
```

---

## ⬇️ 快速安装

### 方式1: 克隆仓库（推荐新手）
```bash
# 1. 克隆项目
git clone https://github.com/zsy619/yyhertz.git
cd yyhertz

# 2. 初始化依赖
go mod tidy

# 3. 运行示例
cd example/simple
go run main.go

# 4. 访问应用
curl http://localhost:8888/home/index
```

### 方式2: Go模块引入（推荐生产）
```bash
# 1. 初始化新项目
mkdir my-yyhertz-app && cd my-yyhertz-app
go mod init my-yyhertz-app

# 2. 添加YYHertz依赖
go get github.com/zsy619/yyhertz@latest

# 3. 创建main.go
cat > main.go << 'EOF'
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

type HomeController struct {
    mvc.BaseController
}

func (c *HomeController) GetIndex() {
    c.JSON(map[string]any{
        "message": "Hello YYHertz!",
        "version": "2.0.0",
    })
}

func main() {
    app := mvc.HertzApp
    
    // 添加中间件
    app.Use(
        middleware.Recovery(),
        middleware.Logger(),
    )
    
    // 自动路由注册
    app.AutoRouters(&HomeController{})
    
    app.Run(":8888")
}
EOF

# 4. 启动应用
go run main.go
```

---

## 🎯 创建第一个应用

### Step 1: 项目结构
创建标准的YYHertz项目结构：

```
my-app/
├── main.go              # 应用入口
├── controllers/         # 控制器目录
│   └── home.go
├── models/             # 模型目录
│   └── user.go
├── views/              # 视图模板目录
│   ├── layouts/
│   └── home/
├── static/             # 静态资源目录
│   ├── css/
│   ├── js/
│   └── images/
├── config/             # 配置文件目录
│   └── app.conf
└── go.mod              # Go模块文件
```

### Step 2: 创建控制器
```bash
mkdir -p controllers
cat > controllers/home.go << 'EOF'
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "time"
)

type HomeController struct {
    mvc.BaseController
}

// GetIndex 处理 GET /home/index
func (c *HomeController) GetIndex() {
    data := map[string]any{
        "title":     "欢迎使用YYHertz",
        "message":   "这是你的第一个YYHertz应用！",
        "timestamp": time.Now().Format("2006-01-02 15:04:05"),
        "features": []string{
            "高性能Web框架",
            "MVC架构设计", 
            "智能路由系统",
            "丰富的中间件",
            "强大的模板引擎",
        },
    }
    c.JSON(data)
}

// PostCreate 处理 POST /home/create
func (c *HomeController) PostCreate() {
    name := c.GetForm("name")
    email := c.GetForm("email")
    
    if name == "" || email == "" {
        c.ErrorJSON(400, "姓名和邮箱不能为空")
        return
    }
    
    c.JSON(map[string]any{
        "success": true,
        "message": "用户创建成功",
        "data": map[string]string{
            "name":  name,
            "email": email,
        },
    })
}

// GetUsers 处理 GET /home/users
func (c *HomeController) GetUsers() {
    users := []map[string]any{
        {"id": 1, "name": "张三", "email": "zhangsan@example.com"},
        {"id": 2, "name": "李四", "email": "lisi@example.com"},
        {"id": 3, "name": "王五", "email": "wangwu@example.com"},
    }
    
    c.JSON(map[string]any{
        "users": users,
        "total": len(users),
    })
}
EOF
```

### Step 3: 创建模型
```bash
mkdir -p models
cat > models/user.go << 'EOF'
package models

import (
    "time"
    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    Name      string         `gorm:"size:100;not null" json:"name"`
    Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
    Phone     string         `gorm:"size:20" json:"phone"`
    Status    int            `gorm:"default:1" json:"status"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
    return "users"
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
    return u.Status == 1
}

// GetDisplayName 获取显示名称
func (u *User) GetDisplayName() string {
    if u.Name != "" {
        return u.Name
    }
    return u.Email
}
EOF
```

### Step 4: 完善主程序
```bash
cat > main.go << 'EOF'
package main

import (
    "my-yyhertz-app/controllers"
    
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
    "github.com/zsy619/yyhertz/framework/config"
)

func main() {
    // 获取应用实例
    app := mvc.HertzApp
    
    // 配置中间件
    app.Use(
        middleware.Recovery(),          // 异常恢复
        middleware.Logger(),            // 访问日志
        middleware.CORS(),              // 跨域支持
        middleware.RateLimit(100, 60),  // 限流: 100次/分钟
    )
    
    // 注册控制器 - 自动路由
    app.AutoRouters(&controllers.HomeController{})
    
    // 手动路由示例
    app.GET("/health", func(c *mvc.Context) {
        c.JSON(map[string]any{
            "status": "ok",
            "time":   c.Now().Format("2006-01-02 15:04:05"),
        })
    })
    
    // 静态文件服务
    app.Static("/static", "./static")
    
    // 启动服务器
    app.Run(":8888")
}
EOF
```

---

## 💡 核心概念速览

### 1. MVC架构
```go
// Model - 数据模型
type User struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

// View - 通过模板渲染或JSON响应
func (c *HomeController) GetIndex() {
    c.JSON(data)  // JSON视图
    // c.HTML("index.html", data)  // HTML视图
}

// Controller - 业务逻辑控制
type HomeController struct {
    mvc.BaseController
}
```

### 2. 自动路由规则
```go
type UserController struct {
    mvc.BaseController
}

// HTTP方法 + 函数名 = 路由路径
func (c *UserController) GetIndex()  {}  // GET  /user/index
func (c *UserController) PostCreate() {} // POST /user/create
func (c *UserController) PutUpdate()  {} // PUT  /user/update
func (c *UserController) DeleteRemove() {} // DELETE /user/remove
```

### 3. 中间件系统
```go
// 全局中间件
app.Use(middleware.Logger())

// 路由级中间件
app.GET("/api/*", middleware.Auth(), handler)

// 控制器级中间件
func (c *UserController) Prepare() {
    // 在每个控制器方法前执行
    c.CheckAuth()
}
```

### 4. 上下文对象
```go
func (c *BaseController) HandleRequest() {
    // 获取请求参数
    name := c.GetForm("name")          // 表单参数
    id := c.GetParam("id")             // 路径参数
    token := c.GetHeader("Authorization") // 请求头
    
    // 设置响应
    c.SetHeader("Content-Type", "application/json")
    c.JSON(map[string]any{"status": "ok"})
}
```

---

## 🛠️ 常用功能示例

### 1. 数据库操作
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/orm"
    "gorm.io/gorm"
)

func initDatabase() {
    // 初始化数据库连接
    db, err := orm.InitDB("sqlite", "app.db")
    if err != nil {
        panic(err)
    }
    
    // 自动迁移
    db.AutoMigrate(&User{})
}

func (c *UserController) GetList() {
    var users []User
    db := orm.GetDB()
    
    // 查询用户列表
    result := db.Where("status = ?", 1).
                 Order("created_at desc").
                 Limit(10).
                 Find(&users)
    
    if result.Error != nil {
        c.ErrorJSON(500, "查询失败")
        return
    }
    
    c.JSON(map[string]any{
        "users": users,
        "total": len(users),
    })
}
```

### 2. 模板渲染
```bash
# 创建模板目录
mkdir -p views/layouts views/home

# 布局模板
cat > views/layouts/base.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
    <meta charset="utf-8">
</head>
<body>
    <header>
        <h1>YYHertz App</h1>
    </header>
    <main>
        {{template "content" .}}
    </main>
</body>
</html>
EOF

# 页面模板
cat > views/home/index.html << 'EOF'
{{define "content"}}
<div class="container">
    <h2>{{.message}}</h2>
    <p>当前时间: {{.timestamp}}</p>
    <ul>
        {{range .features}}
        <li>{{.}}</li>
        {{end}}
    </ul>
</div>
{{end}}
EOF
```

```go
// 控制器中使用模板
func (c *HomeController) GetIndex() {
    data := map[string]any{
        "title":     "首页",
        "message":   "欢迎使用YYHertz",
        "timestamp": time.Now().Format("2006-01-02 15:04:05"),
        "features":  []string{"高性能", "易用性", "可扩展"},
    }
    
    // 渲染模板
    c.HTMLWithLayout("home/index.html", "layouts/base.html", data)
}
```

### 3. API接口开发
```go
// API控制器
type APIController struct {
    mvc.BaseController
}

// 统一API响应格式
type APIResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}

func (c *APIController) Success(data any) {
    c.JSON(APIResponse{
        Code:    200,
        Message: "success",
        Data:    data,
    })
}

func (c *APIController) Error(code int, message string) {
    c.JSON(APIResponse{
        Code:    code,
        Message: message,
    })
}

// 用户API
type UserAPIController struct {
    APIController
}

func (c *UserAPIController) GetList() {
    users := getUserList()
    c.Success(users)
}

func (c *UserAPIController) PostCreate() {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.Error(400, "参数错误")
        return
    }
    
    if err := createUser(&user); err != nil {
        c.Error(500, "创建失败")
        return
    }
    
    c.Success(user)
}
```

### 4. 配置管理
```bash
# 创建配置文件
cat > config/app.conf << 'EOF'
# 应用配置
app.name = my-yyhertz-app
app.version = 1.0.0
app.host = localhost
app.port = 8888

# 数据库配置
db.driver = sqlite
db.host = localhost
db.port = 3306
db.name = app.db
db.user = root
db.password = 

# Redis配置
redis.host = localhost
redis.port = 6379
redis.password = 
redis.db = 0

# 日志配置
log.level = info
log.file = logs/app.log
EOF
```

```go
// 加载配置
import "github.com/zsy619/yyhertz/framework/config"

func init() {
    config.LoadConfig("config/app.conf")
}

func main() {
    // 读取配置
    appName := config.GetString("app.name")
    port := config.GetInt("app.port")
    
    app := mvc.HertzApp
    app.Run(fmt.Sprintf(":%d", port))
}
```

---

## 🧪 测试应用

### 1. 启动应用
```bash
go run main.go
```

### 2. 测试接口
```bash
# 测试首页
curl http://localhost:8888/home/index

# 测试用户列表
curl http://localhost:8888/home/users

# 测试创建用户
curl -X POST http://localhost:8888/home/create \
  -d "name=张三&email=zhangsan@example.com"

# 测试健康检查
curl http://localhost:8888/health
```

### 3. 预期响应
```json
{
  "title": "欢迎使用YYHertz",
  "message": "这是你的第一个YYHertz应用！",
  "timestamp": "2024-09-05 15:30:25",
  "features": [
    "高性能Web框架",
    "MVC架构设计",
    "智能路由系统",
    "丰富的中间件",
    "强大的模板引擎"
  ]
}
```

---

## 📚 下一步学习

### 🎯 推荐学习路径

1. **基础概念** (30分钟)
   - [架构设计文档](./ARCHITECTURE.md)
   - [MVC模式详解](./tutorials/mvc/README.md)

2. **核心功能** (2小时)
   - [路由系统详解](./tutorials/routing/README.md)
   - [中间件开发](./tutorials/middleware/README.md)
   - [模板引擎使用](./template/FUNCTIONS.md)

3. **数据操作** (1小时)
   - [数据库操作基础](./tutorials/database/README.md)
   - [ORM使用指南](../MYBATIS_SAMPLES.md)

4. **进阶特性** (3小时)
   - [性能优化](./performance/OPTIMIZATION.md)
   - [安全配置](./security/SECURITY.md)
   - [生产部署](./best-practices/DEPLOYMENT.md)

### 🔗 相关资源

- **[完整API文档](./API.md)** - 详细的接口参考
- **[示例项目](../example/)** - 丰富的代码示例
- **[最佳实践](./best-practices/)** - 生产环境建议
- **[社区支持](../CONTRIBUTING.md)** - 获取帮助和参与贡献

### 🆘 遇到问题？

- **查看**: [故障排除指南](./troubleshooting/README.md)
- **搜索**: [GitHub Issues](https://github.com/zsy619/yyhertz/issues)
- **提问**: [讨论区](https://github.com/zsy619/yyhertz/discussions)
- **联系**: support@yyhertz.com

---

<div align="center">

🎉 **恭喜！你已经成功创建了第一个YYHertz应用！** 🎉

**接下来探索更多YYHertz的强大功能吧！**

</div>