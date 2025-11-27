# 🏗️ 项目结构

YYHertz框架采用清晰的目录结构设计，遵循MVC架构模式和Go项目最佳实践。本文档将详细介绍标准项目结构和各目录的作用。

## 📁 标准项目结构

### 完整项目模板
```
my-hertz-app/
├── 📁 cmd/                          # 应用程序入口
│   └── server/
│       └── main.go                  # 服务器启动文件
├── 📁 internal/                     # 私有代码目录  
│   ├── controllers/                 # 控制器层
│   │   ├── api/                     # API控制器
│   │   │   ├── v1/                  # API版本控制
│   │   │   │   ├── user_controller.go
│   │   │   │   └── product_controller.go
│   │   │   └── v2/
│   │   │       └── user_controller.go
│   │   └── web/                     # Web控制器
│   │       ├── home_controller.go
│   │       └── admin_controller.go
│   ├── models/                      # 数据模型
│   │   ├── user.go
│   │   ├── product.go
│   │   └── base.go                  # 基础模型
│   ├── services/                    # 业务服务层
│   │   ├── user_service.go
│   │   ├── product_service.go
│   │   └── interfaces/              # 服务接口
│   │       ├── user_service.go
│   │       └── product_service.go
│   ├── repositories/                # 数据访问层
│   │   ├── user_repository.go
│   │   ├── product_repository.go
│   │   └── interfaces/
│   │       ├── user_repository.go
│   │       └── product_repository.go
│   └── middleware/                  # 自定义中间件
│       ├── auth.go
│       ├── rate_limit.go
│       └── custom_logger.go
├── 📁 config/                       # 配置文件
│   ├── app.yaml                     # 应用配置
│   ├── database.yaml                # 数据库配置
│   ├── redis.yaml                   # Redis配置
│   └── local.yaml                   # 本地开发配置
├── 📁 docs/                         # 文档目录
│   ├── api/                         # API文档
│   │   ├── swagger.yaml
│   │   └── postman_collection.json
│   ├── deployment/                  # 部署文档
│   │   ├── docker.md
│   │   └── kubernetes.md
│   └── development/                 # 开发文档
│       ├── setup.md
│       └── coding_standards.md
├── 📁 pkg/                          # 公共库代码
│   ├── utils/                       # 工具函数
│   │   ├── crypto.go
│   │   ├── time.go
│   │   └── validator.go
│   ├── constants/                   # 常量定义
│   │   ├── errors.go
│   │   └── status.go
│   └── types/                       # 公共类型
│       ├── response.go
│       └── request.go
├── 📁 views/                        # 模板文件
│   ├── layout/                      # 布局模板
│   │   ├── base.html
│   │   └── admin.html
│   ├── web/                         # Web页面模板
│   │   ├── home/
│   │   │   ├── index.html
│   │   │   └── about.html
│   │   └── user/
│   │       ├── profile.html
│   │       └── settings.html
│   └── partials/                    # 组件模板
│       ├── header.html
│       ├── footer.html
│       └── navbar.html
├── 📁 static/                       # 静态资源
│   ├── css/                         # 样式文件
│   │   ├── app.css
│   │   └── admin.css
│   ├── js/                          # JavaScript文件
│   │   ├── app.js
│   │   └── components/
│   │       └── datatable.js
│   ├── images/                      # 图片资源
│   │   ├── logo.png
│   │   └── icons/
│   └── fonts/                       # 字体文件
├── 📁 migrations/                   # 数据库迁移
│   ├── 20240101_create_users_table.sql
│   ├── 20240102_create_products_table.sql
│   └── schema.sql
├── 📁 tests/                        # 测试文件
│   ├── integration/                 # 集成测试
│   │   ├── api_test.go
│   │   └── web_test.go
│   ├── unit/                        # 单元测试  
│   │   ├── services/
│   │   │   └── user_service_test.go
│   │   └── utils/
│   │       └── crypto_test.go
│   └── fixtures/                    # 测试数据
│       ├── users.json
│       └── products.json
├── 📁 scripts/                      # 脚本文件
│   ├── build.sh                     # 构建脚本
│   ├── deploy.sh                    # 部署脚本
│   └── migrate.sh                   # 数据库迁移脚本
├── 📁 deployments/                  # 部署配置
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── kubernetes/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── nginx/
│       └── nginx.conf
├── 📄 go.mod                        # Go模块文件
├── 📄 go.sum                        # 依赖校验和
├── 📄 .env                          # 环境变量文件
├── 📄 .env.example                  # 环境变量示例
├── 📄 .gitignore                    # Git忽略文件
├── 📄 Makefile                      # Make构建文件
└── 📄 README.md                     # 项目说明文档
```

## 🗂️ 核心目录说明

### 1. `cmd/` - 应用入口
**职责**: 应用程序启动点，包含main函数
```go
// cmd/server/main.go
package main

import (
    "log"
    "github.com/zsy619/yyhertz/framework/mvc"
    "my-hertz-app/internal/controllers"
    "my-hertz-app/internal/middleware"
)

func main() {
    app := mvc.HertzApp
    
    // 注册中间件
    app.Use(
        middleware.Logger(),
        middleware.Recovery(),
        middleware.AuthMiddleware(),
    )
    
    // 注册控制器
    app.AutoRouters(&controllers.HomeController{})
    
    // 启动服务
    log.Println("Server starting on :8888")
    app.Run(":8888")
}
```

### 2. `internal/` - 私有代码
**职责**: 应用核心业务逻辑，外部不可访问

#### controllers/ - 控制器层
```go
// internal/controllers/api/v1/user_controller.go
package v1

import "github.com/zsy619/yyhertz/framework/mvc"

type UserController struct {
    mvc.BaseController
    userService services.UserService
}

func (c *UserController) GetList() {
    users, err := c.userService.GetAllUsers()
    if err != nil {
        c.Error(500, err.Error())
        return
    }
    c.JSON(users)
}
```

#### models/ - 数据模型
```go
// internal/models/user.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    Username string `gorm:"uniqueIndex;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;size:100" json:"email"`
    Password string `gorm:"size:255" json:"-"`
    Avatar   string `gorm:"size:255" json:"avatar"`
    Status   int    `gorm:"default:1" json:"status"`
}

func (User) TableName() string {
    return "users"
}
```

#### services/ - 业务服务层
```go
// internal/services/user_service.go
package services

import (
    "my-hertz-app/internal/models"
    "my-hertz-app/internal/repositories"
)

type UserService struct {
    userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
    return &UserService{userRepo: userRepo}
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
    return s.userRepo.FindAll()
}

func (s *UserService) CreateUser(user *models.User) error {
    // 业务逻辑处理
    if err := s.validateUser(user); err != nil {
        return err
    }
    return s.userRepo.Create(user)
}
```

### 3. `pkg/` - 公共库
**职责**: 可被外部项目引用的公共代码
```go
// pkg/utils/crypto.go
package utils

import (
    "crypto/md5"
    "fmt"
)

func MD5Hash(text string) string {
    hash := md5.Sum([]byte(text))
    return fmt.Sprintf("%x", hash)
}

// pkg/types/response.go  
package types

type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Code    int         `json:"code"`
}
```

### 4. `config/` - 配置管理
**职责**: 应用配置文件和环境变量
```yaml
# config/app.yaml
app:
  name: "My Hertz App"
  version: "1.0.0"
  mode: "debug"
  
server:
  addr: ":8888"
  read_timeout: "30s"
  write_timeout: "30s"

database:
  driver: "mysql"
  dsn: "${DB_DSN}"
  max_open_conns: 100
  max_idle_conns: 10

redis:
  addr: "${REDIS_ADDR:localhost:6379}"
  password: "${REDIS_PASSWORD:}"
  db: 0
```

## 🎨 模板和静态资源

### Views目录结构
```
views/
├── layout/                 # 布局模板
│   ├── base.html          # 基础布局
│   ├── admin.html         # 后台布局
│   └── api.html           # API文档布局
├── web/                   # 前端页面
│   ├── home/
│   │   ├── index.html
│   │   └── about.html
│   └── user/
│       ├── login.html
│       └── register.html
├── partials/              # 组件模板
│   ├── header.html
│   ├── footer.html
│   ├── navbar.html
│   └── sidebar.html
└── emails/                # 邮件模板
    ├── welcome.html
    └── reset_password.html
```

### Static目录结构
```
static/
├── css/
│   ├── app.css           # 主样式
│   ├── admin.css         # 后台样式
│   └── vendor/           # 第三方CSS
│       └── bootstrap.css
├── js/
│   ├── app.js            # 主脚本
│   ├── admin.js          # 后台脚本
│   ├── components/       # 组件脚本
│   │   ├── datatable.js
│   │   └── modal.js
│   └── vendor/           # 第三方JS
│       ├── jquery.js
│       └── bootstrap.js
├── images/
│   ├── logo.png
│   ├── icons/
│   │   ├── favicon.ico
│   │   └── apple-touch-icon.png
│   └── uploads/          # 用户上传
└── fonts/
    ├── custom-font.woff2
    └── icons.woff
```

## 🧪 测试目录组织

### 测试结构最佳实践
```
tests/
├── integration/           # 集成测试
│   ├── api/
│   │   ├── user_api_test.go
│   │   └── product_api_test.go
│   └── web/
│       ├── home_test.go
│       └── admin_test.go
├── unit/                 # 单元测试
│   ├── controllers/
│   │   └── user_controller_test.go
│   ├── services/
│   │   └── user_service_test.go
│   ├── repositories/
│   │   └── user_repository_test.go
│   └── utils/
│       └── crypto_test.go
├── fixtures/             # 测试数据
│   ├── users.json
│   ├── products.json
│   └── sql/
│       └── test_data.sql
├── mocks/                # Mock对象
│   ├── user_service_mock.go
│   └── user_repository_mock.go
└── helpers/              # 测试助手
    ├── database.go
    ├── http_client.go
    └── fixtures.go
```

## ⚙️ 配置文件管理

### 多环境配置策略
```
config/
├── base.yaml            # 基础配置
├── development.yaml     # 开发环境
├── staging.yaml         # 预发布环境
├── production.yaml      # 生产环境
└── local.yaml          # 本地开发配置 (git ignore)
```

### 环境变量文件
```bash
# .env
APP_ENV=development
APP_DEBUG=true
APP_KEY=your-secret-key

DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=hertz_app
DB_USERNAME=root
DB_PASSWORD=secret

REDIS_HOST=localhost  
REDIS_PORT=6379
REDIS_PASSWORD=

LOG_LEVEL=debug
LOG_OUTPUT=stdout
```

## 📦 依赖管理

### go.mod 示例
```go
module my-hertz-app

go 1.21

require (
    github.com/zsy619/yyhertz v2.0.0
    github.com/cloudwego/hertz v0.8.0
    gorm.io/gorm v1.25.5
    gorm.io/driver/mysql v1.5.2
    github.com/go-redis/redis/v8 v8.11.5
    github.com/spf13/viper v1.17.0
    github.com/golang-jwt/jwt/v5 v5.2.0
)

require (
    // 间接依赖...
)
```

### Makefile 构建脚本
```makefile
.PHONY: build run test clean docker

# 变量定义
APP_NAME=my-hertz-app
VERSION=1.0.0
BUILD_DIR=build

# 构建应用
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

# 运行应用
run:
	go run cmd/server/main.go

# 运行测试
test:
	go test -v ./tests/...

# 代码格式化
fmt:
	go fmt ./...
	goimports -w .

# 代码检查
lint:
	golangci-lint run

# 清理构建文件
clean:
	rm -rf $(BUILD_DIR)

# 构建Docker镜像
docker:
	docker build -t $(APP_NAME):$(VERSION) .
```

## 🚀 最佳实践建议

### 1. 目录命名规范
- **小写字母** + **下划线** (snake_case)
- **复数形式** 用于集合目录 (controllers, models)
- **单数形式** 用于功能目录 (config, static)

### 2. 文件命名规范
- Go文件: **snake_case.go** (user_controller.go)
- 模板文件: **kebab-case.html** (user-profile.html)
- 配置文件: **kebab-case.yaml** (database-config.yaml)

### 3. 包导入顺序
```go
import (
    // 标准库
    "fmt"
    "time"
    "context"
    
    // 第三方库
    "github.com/cloudwego/hertz"
    "gorm.io/gorm"
    
    // 本项目包
    "my-hertz-app/internal/models"
    "my-hertz-app/internal/services"
)
```

### 4. 接口设计原则
- 接口定义在使用方包中，而不是实现方
- 接口应该小而专一 (Interface Segregation)
- 优先使用组合而非继承

## 📖 下一步

现在您已经了解了YYHertz的项目结构，建议继续学习：

1. 🎛️ [控制器开发](/home/controller) - 掌握MVC核心
2. 🗄️ [数据库集成](/home/gorm) - 配置数据访问层
3. 🔌 [中间件系统](/home/middleware-overview) - 了解请求处理流程
4. ⚙️ [配置管理](/home/app-config) - 掌握配置文件使用

---

**💡 记住：好的项目结构是成功应用的基础！**