# 基于注释的Spring Boot风格注解路由系统

这个包提供了基于Go注释的Spring Boot风格注解路由系统，通过解析源码注释来定义RESTful API和MVC路由，更符合Go语言的编程习惯。

## 🌟 特性

- ✅ **基于注释** - 使用标准Go注释，符合Go语言习惯
- ✅ **Spring Boot风格** - 熟悉的`@RestController`、`@GetMapping`等注解
- ✅ **自动源码解析** - 自动解析Go源文件中的注释注解
- ✅ **注释即文档** - 注释同时作为代码文档和路由配置
- ✅ **完全兼容现有系统** - 与BaseController系统无缝集成
- ✅ **路由分析工具** - 提供路由分析和诊断功能
- ✅ **RESTful支持** - 完整的RESTful API支持

## 🚀 快速开始

### 1. 定义控制器

```go
package controllers

import "github.com/zsy619/yyhertz/framework/mvc/core"

// UserController 用户控制器
// @RestController
// @RequestMapping("/api/users")  
// @Description("用户管理控制器")
type UserController struct {
    core.BaseController
}
```

### 2. 定义方法

```go
// GetUsers 获取用户列表
// @GetMapping("/")
// @Description("获取用户列表")
// @RequestParam(name="page", required=false, defaultValue="1")
// @RequestParam(name="size", required=false, defaultValue="10")
func (c *UserController) GetUsers() ([]*UserResponse, error) {
    page := c.GetQuery("page", "1")
    size := c.GetQuery("size", "10")
    
    // 业务逻辑...
    users := []*UserResponse{
        {ID: 1, Name: "张三", Email: "zhang@example.com"},
    }
    
    return users, nil
}

// GetUser 获取单个用户
// @GetMapping("/{id}")
// @Description("根据ID获取用户详情")
// @PathVariable("id")
func (c *UserController) GetUser() (*UserResponse, error) {
    id := c.GetParam("id")
    
    // 业务逻辑...
    user := &UserResponse{
        ID: 1, Name: "张三", Email: "zhang@example.com",
    }
    
    return user, nil
}

// CreateUser 创建用户
// @PostMapping("/")
// @Description("创建新用户")
// @RequestBody
func (c *UserController) CreateUser(req *UserRequest) (*UserResponse, error) {
    // req 会自动从请求体绑定
    user := &UserResponse{
        ID: 100, Name: req.Name, Email: req.Email,
    }
    
    return user, nil
}
```

### 3. 启动应用

```go
package main

import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/zsy619/yyhertz/framework/mvc/annotation"
)

func main() {
    // 创建Hertz引擎
    h := server.Default()
    
    // 创建支持注释注解的应用
    app := annotation.NewAppWithComments(h.Engine)
    
    // 自动扫描并注册控制器
    app.AutoScanAndRegister(
        &UserController{},
        &ProductController{},
    )
    
    h.Spin()
}
```

## 📝 注解类型

### 控制器级别注解

| 注解 | 用途 | 示例 |
|------|------|------|
| `@RestController` | REST控制器 | `// @RestController` |
| `@Controller` | MVC控制器 | `// @Controller` |
| `@RequestMapping("/path")` | 基础路径 | `// @RequestMapping("/api/users")` |
| `@Description("desc")` | 描述信息 | `// @Description("用户管理控制器")` |

### 方法级别注解

| 注解 | HTTP方法 | 示例 |
|------|----------|------|
| `@GetMapping("/path")` | GET | `// @GetMapping("/")` |
| `@PostMapping("/path")` | POST | `// @PostMapping("/")` |
| `@PutMapping("/path")` | PUT | `// @PutMapping("/{id}")` |
| `@DeleteMapping("/path")` | DELETE | `// @DeleteMapping("/{id}")` |
| `@PatchMapping("/path")` | PATCH | `// @PatchMapping("/{id}")` |

### 参数注解

| 注解 | 用途 | 示例 |
|------|------|------|
| `@PathVariable("name")` | 路径参数 | `// @PathVariable("id")` |
| `@RequestParam(name="page", required=false, defaultValue="1")` | 查询参数 | 支持必需性和默认值 |
| `@RequestBody` | 请求体 | `// @RequestBody` |
| `@RequestHeader(name="Auth", required=false)` | 请求头 | 支持必需性和默认值 |
| `@CookieValue("sessionId")` | Cookie值 | `// @CookieValue("sessionId")` |

### 其他注解

| 注解 | 用途 | 示例 |
|------|------|------|
| `@Description("desc")` | 描述信息 | `// @Description("获取用户列表")` |
| `@Middleware("auth", "ratelimit")` | 中间件 | 支持多个中间件 |

## 🎯 完整示例

### REST API控制器

```go
// UserController 用户控制器
// @RestController
// @RequestMapping("/api/v1/users")
// @Description("用户管理REST API")
type UserController struct {
    core.BaseController
}

// GetUsers 获取用户列表
// @GetMapping("/")
// @Description("分页获取用户列表")
// @RequestParam(name="page", required=false, defaultValue="1")
// @RequestParam(name="size", required=false, defaultValue="10")
// @RequestParam(name="keyword", required=false)
func (c *UserController) GetUsers() ([]*UserResponse, error) {
    page := c.GetQuery("page", "1")
    size := c.GetQuery("size", "10")
    keyword := c.GetQuery("keyword", "")

    log.Printf("获取用户列表: page=%s, size=%s, keyword=%s", page, size, keyword)

    users := []*UserResponse{
        {ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25, Status: "active"},
        {ID: 2, Name: "李四", Email: "li@example.com", Age: 30, Status: "active"},
    }

    return users, nil
}

// GetUser 获取用户详情
// @GetMapping("/{id}")
// @Description("根据ID获取用户详情")
// @PathVariable("id")
func (c *UserController) GetUser() (*UserResponse, error) {
    id := c.GetParam("id")

    log.Printf("获取用户详情: id=%s", id)

    user := &UserResponse{
        ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25, Status: "active",
    }

    return user, nil
}

// CreateUser 创建用户
// @PostMapping("/")
// @Description("创建新用户")
// @RequestBody
func (c *UserController) CreateUser(req *UserRequest) (*UserResponse, error) {
    log.Printf("创建用户: %+v", req)

    user := &UserResponse{
        ID: 100, Name: req.Name, Email: req.Email, Age: req.Age, Status: "active",
    }

    return user, nil
}

// UpdateUser 更新用户
// @PutMapping("/{id}")
// @Description("更新用户信息")
// @PathVariable("id")
// @RequestBody
func (c *UserController) UpdateUser(req *UserRequest) (*UserResponse, error) {
    id := c.GetParam("id")

    log.Printf("更新用户: id=%s, data=%+v", id, req)

    user := &UserResponse{
        ID: 1, Name: req.Name, Email: req.Email, Age: req.Age, Status: "active",
    }

    return user, nil
}

// DeleteUser 删除用户
// @DeleteMapping("/{id}")
// @Description("删除用户")
// @PathVariable("id")
func (c *UserController) DeleteUser() (map[string]interface{}, error) {
    id := c.GetParam("id")

    log.Printf("删除用户: id=%s", id)

    return map[string]interface{}{
        "success": true,
        "message": "用户删除成功",
        "id":      id,
    }, nil
}

// SearchUsers 搜索用户
// @GetMapping("/search")
// @Description("搜索用户")
// @RequestParam(name="q", required=true)
// @RequestParam(name="type", required=false, defaultValue="name")
// @RequestHeader(name="X-Request-ID", required=false)
func (c *UserController) SearchUsers() ([]*UserResponse, error) {
    query := c.GetQuery("q", "")
    searchType := c.GetQuery("type", "name")
    requestID := c.GetHeader("X-Request-ID")

    log.Printf("搜索用户: q=%s, type=%s, requestID=%s", query, searchType, string(requestID))

    users := []*UserResponse{
        {ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25, Status: "active"},
    }

    return users, nil
}
```

### MVC Web控制器

```go
// WebController Web控制器
// @Controller
// @RequestMapping("/web")
// @Description("Web页面控制器")
type WebController struct {
    core.BaseController
}

// Index 首页
// @GetMapping("/")
// @Description("网站首页")
func (c *WebController) Index() {
    c.Data["Title"] = "首页"
    c.Data["Message"] = "欢迎来到基于注释注解的YYHertz框架!"
    c.TplName = "index.html"
}

// UserList 用户列表页面
// @GetMapping("/users")
// @Description("用户列表页面")
// @RequestParam(name="page", required=false, defaultValue="1")
func (c *WebController) UserList() {
    page := c.GetQuery("page", "1")

    c.Data["Title"] = "用户列表"
    c.Data["Users"] = []map[string]interface{}{
        {"ID": 1, "Name": "张三", "Email": "zhang@example.com"},
        {"ID": 2, "Name": "李四", "Email": "li@example.com"},
    }
    c.Data["Page"] = page
    c.TplName = "users/list.html"
}
```

### 带中间件的控制器

```go
// AdminController 管理员控制器
// @RestController
// @RequestMapping("/api/admin")
// @Description("管理员控制器")
// @Middleware("auth", "admin")
type AdminController struct {
    core.BaseController
}

// GetDashboard 获取仪表板数据
// @GetMapping("/dashboard")
// @Description("获取管理员仪表板数据")
func (c *AdminController) GetDashboard() (map[string]interface{}, error) {
    dashboard := map[string]interface{}{
        "userCount":    1000,
        "productCount": 500,
        "orderCount":   2000,
        "revenue":      100000.0,
        "systemStatus": "healthy",
    }

    return dashboard, nil
}

// BackupSystem 系统备份
// @PostMapping("/system/backup")
// @Description("执行系统备份")
// @RequestBody
func (c *AdminController) BackupSystem(req *BackupRequest) (map[string]interface{}, error) {
    log.Printf("执行系统备份: %+v", req)

    result := map[string]interface{}{
        "success":   true,
        "message":   "备份任务已启动",
        "backupId":  "backup_20240801_001",
        "type":      req.Type,
        "timestamp": "2024-08-01T22:00:00Z",
    }

    return result, nil
}
```

## 🛠️ 高级功能

### 1. 路由分析

```go
// 收集路由信息
collector := annotation.NewRouteCollector().CollectFromApp(app)

// 统计信息
fmt.Printf("总路由数: %d\n", collector.GetRouteCount())
fmt.Printf("控制器数: %d\n", collector.GetControllerCount())

// 方法统计
methodCounts := collector.GetMethodCount()
for method, count := range methodCounts {
    fmt.Printf("%s: %d\n", method, count)
}

// 路由分析
analyzer := annotation.NewRouteAnalyzer(collector)

// 检查重复路由
duplicates := analyzer.AnalyzeDuplicates()
if len(duplicates) > 0 {
    fmt.Println("发现重复路由:")
    for _, duplicate := range duplicates {
        fmt.Printf("  %s -> %v\n", duplicate[0], duplicate[1:])
    }
}

// RESTful分析
restPatterns := analyzer.AnalyzeRESTfulness()
fmt.Println("RESTful模式:")
for pattern, paths := range restPatterns {
    fmt.Printf("  %s: %v\n", pattern, paths)
}
```

### 2. 自动发现

```go
// 自动发现控制器
discovery := annotation.NewAutoDiscovery(app).
    WithScanPaths("./controllers", "./api").
    WithExcludePaths("./test").
    WithControllerSuffix("Controller")

err := discovery.Discover()
if err != nil {
    log.Fatal(err)
}
```

### 3. 手动源码扫描

```go
// 扫描特定包
err := annotation.ScanPackage("./controllers")

// 扫描特定文件
err := annotation.ScanSourceFile("user_controller.go")

// 获取注解信息
controllerInfo := annotation.GetGlobalControllerInfo("controllers", "UserController")
methodInfo := annotation.GetGlobalMethodInfo("controllers", "UserController", "GetUsers")
```

## 🎨 生成的路由示例

基于上面的注解配置，将生成以下路由：

```
用户管理API:
GET    /api/v1/users              -> UserController.GetUsers
GET    /api/v1/users/{id}         -> UserController.GetUser
GET    /api/v1/users/search       -> UserController.SearchUsers
POST   /api/v1/users              -> UserController.CreateUser
PUT    /api/v1/users/{id}         -> UserController.UpdateUser
DELETE /api/v1/users/{id}         -> UserController.DeleteUser

Web页面:
GET    /web/                      -> WebController.Index
GET    /web/users                 -> WebController.UserList

管理员:
GET    /api/admin/dashboard       -> AdminController.GetDashboard
POST   /api/admin/system/backup   -> AdminController.BackupSystem
```

## 🔧 配置选项

### 参数配置详解

```go
// 查询参数配置
// @RequestParam(name="page", required=false, defaultValue="1")
// - name: 参数名称
// - required: 是否必需 (true/false)
// - defaultValue: 默认值

// 请求头配置
// @RequestHeader(name="Authorization", required=true)
// - name: 请求头名称
// - required: 是否必需

// Cookie配置
// @CookieValue("session_id", required=false, defaultValue="")
// - 第一个参数: Cookie名称
// - required: 是否必需
// - defaultValue: 默认值
```

### 中间件配置

```go
// 单个中间件
// @Middleware("auth")

// 多个中间件
// @Middleware("auth", "ratelimit", "cors")
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test -v ./example/comments/

# 运行指定测试
go test -v -run TestCommentAnnotationParsing ./example/comments/

# 运行基准测试
go test -v -bench=. ./example/comments/
```

### 测试覆盖

```bash
# 生成测试覆盖报告
go test -cover ./example/comments/

# 生成详细覆盖报告
go test -coverprofile=coverage.out ./example/comments/
go tool cover -html=coverage.out
```

## 📊 与Struct标签注解的对比

| 特性 | 注释注解 | Struct标签注解 |
|------|----------|---------------|
| **可读性** | ✅ 更符合Go习惯 | ❌ 标签较难阅读 |
| **维护性** | ✅ 注释即文档 | ❌ 需要额外文档 |
| **IDE支持** | ✅ 更好的语法高亮 | ❌ 标签支持有限 |
| **版本控制** | ✅ 变更更清晰 | ❌ 标签变更难追踪 |
| **学习成本** | ✅ 熟悉的注解风格 | ❌ 需要学习标签语法 |
| **性能** | ❌ 需要源码解析 | ✅ 编译时解析 |
| **部署要求** | ❌ 需要源码访问 | ✅ 无额外要求 |

## 🚨 注意事项

1. **源码解析** - 需要访问Go源文件，部署时确保源码可访问
2. **注释格式** - 严格按照示例格式编写注释
3. **参数绑定** - 确保方法签名与注解参数匹配
4. **错误处理** - 方法可以返回error，会自动处理HTTP状态码
5. **文件路径** - 确保源文件路径正确，以便正确解析

## 🔄 迁移指南

### 从Struct标签迁移到注释

**旧方式 (struct标签):**
```go
type UserController struct {
    core.BaseController `rest:"" mapping:"/api/users"`
}

func init() {
    annotation.RegisterGetMethod(userType, "GetUsers", "/").
        WithQueryParam("page", false, "1")
}
```

**新方式 (注释):**
```go
// @RestController
// @RequestMapping("/api/users")
type UserController struct {
    core.BaseController
}

// @GetMapping("/")
// @RequestParam(name="page", required=false, defaultValue="1")
func (c *UserController) GetUsers() { ... }
```

## 🛣️ 路线图

- [ ] 支持OpenAPI/Swagger文档自动生成
- [ ] 支持参数验证注解增强
- [ ] 支持缓存注解
- [ ] 支持事务注解
- [ ] 支持权限注解
- [ ] 支持限流注解
- [ ] 完善IDE插件支持

## 💡 最佳实践

1. **注释即文档** - 在注解中提供清晰的描述
2. **RESTful设计** - 遵循RESTful API设计原则
3. **参数验证** - 在结构体中使用binding标签进行验证
4. **错误处理** - 统一的错误处理和响应格式
5. **中间件使用** - 合理使用中间件进行横切关注点处理
6. **源码管理** - 确保部署环境能够访问到源码文件

## 🎯 实际应用场景

### 微服务架构

```go
// @RestController
// @RequestMapping("/api/v1/order")
// @Description("订单微服务")
// @Middleware("auth", "ratelimit")
type OrderController struct {
    core.BaseController
}
```

### API版本控制

```go
// @RestController  
// @RequestMapping("/api/v2/users")
// @Description("用户API v2版本")
type UserV2Controller struct {
    core.BaseController
}
```

### 多租户系统

```go
// @RestController
// @RequestMapping("/api/{tenant}/users")
// @Description("多租户用户管理")
type TenantUserController struct {
    core.BaseController
}

// @GetMapping("/")
// @PathVariable("tenant")
// @RequestParam(name="page", required=false, defaultValue="1")
func (c *TenantUserController) GetUsers() { ... }
```

这个基于注释的注解系统让你可以用更自然的Go方式编写Spring Boot风格的Web应用！