# Spring Boot风格注解路由系统

这个包提供了类似Spring Boot的注解路由系统，支持通过struct标签和方法注册来定义RESTful API和MVC路由。

## 特性

- ✅ **struct级别注解** - 使用Go struct标签定义控制器类型和基础路径
- ✅ **方法级别注解** - 通过注册器定义HTTP方法映射
- ✅ **自动路由扫描** - 自动解析和注册控制器路由
- ✅ **参数绑定** - 支持路径参数、查询参数、请求体、请求头等
- ✅ **类型安全** - 强制controller继承`core.IController`接口，基于反射的类型安全参数绑定
- ✅ **兼容现有系统** - 与现有BaseController系统完全兼容
- ✅ **与comment包兼容** - 可以与Go注释注解混合使用
- ✅ **链式配置** - 流畅的API设计

## 快速开始

### 1. 定义控制器

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc/core"
)

// 使用struct标签定义REST控制器
// 注意：必须继承core.BaseController以实现core.IController接口
type UserController struct {
    core.BaseController `rest:"" mapping:"/api/users"`
}

// 传统MVC控制器
type WebController struct {
    core.BaseController `controller:"" mapping:"/web"`
}
```

### 2. 注册方法映射

```go
package controllers

import (
    "reflect"
    "github.com/zsy619/yyhertz/framework/mvc/annotation"
)

func init() {
    userType := reflect.TypeOf((*UserController)(nil)).Elem()
    
    // GET /api/users - 获取用户列表
    annotation.RegisterGetMethod(userType, "GetUsers", "/").
        WithDescription("获取用户列表").
        WithQueryParam("page", false, "1").
        WithQueryParam("size", false, "10")
    
    // GET /api/users/{id} - 获取单个用户
    annotation.RegisterGetMethod(userType, "GetUser", "/{id}").
        WithDescription("获取用户详情").
        WithPathParam("id", true)
    
    // POST /api/users - 创建用户
    annotation.RegisterPostMethod(userType, "CreateUser", "/").
        WithDescription("创建用户").
        WithBodyParam(true)
    
    // PUT /api/users/{id} - 更新用户
    annotation.RegisterPutMethod(userType, "UpdateUser", "/{id}").
        WithDescription("更新用户").
        WithPathParam("id", true).
        WithBodyParam(true)
    
    // DELETE /api/users/{id} - 删除用户
    annotation.RegisterDeleteMethod(userType, "DeleteUser", "/{id}").
        WithDescription("删除用户").
        WithPathParam("id", true)
}
```

### 3. 实现控制器方法

```go
type UserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"min=0,max=120"`
}

type UserResponse struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// GetUsers 获取用户列表
func (c *UserController) GetUsers() ([]*UserResponse, error) {
    page := c.GetQuery("page", "1")
    size := c.GetQuery("size", "10")
    
    // 业务逻辑...
    users := []*UserResponse{
        {ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25},
    }
    
    return users, nil
}

// GetUser 获取单个用户
func (c *UserController) GetUser() (*UserResponse, error) {
    id := c.GetParam("id")
    
    // 业务逻辑...
    user := &UserResponse{
        ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25,
    }
    
    return user, nil
}

// CreateUser 创建用户
func (c *UserController) CreateUser(req *UserRequest) (*UserResponse, error) {
    // req 会自动从请求体绑定
    user := &UserResponse{
        ID: 100,
        Name: req.Name,
        Email: req.Email,
        Age: req.Age,
    }
    
    return user, nil
}
```

### 4. 启动应用

```go
package main

import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/zsy619/yyhertz/framework/mvc/annotation"
)

func main() {
    // 创建Hertz引擎
    h := server.Default()
    
    // 创建支持注解的应用
    app := annotation.NewAppWithAnnotations(h.Engine)
    
    // 自动注册控制器
    app.AutoRegister(
        &UserController{},
        &WebController{},
    )
    
    h.Spin()
}
```

## 注解类型

### Struct级别注解

| 注解 | 用途 | 示例 |
|------|------|------|
| `rest:""` | REST控制器 | `rest:""` |
| `controller:""` | MVC控制器 | `controller:""` |
| `mapping:"/path"` | 基础路径 | `mapping:"/api/users"` |
| `service:"name"` | 服务组件 | `service:"userService"` |
| `repository:"name"` | 数据访问组件 | `repository:"userRepo"` |
| `component:"name"` | 通用组件 | `component:"userComponent"` |

### 方法级别注解

| 方法 | HTTP方法 | 用途 |
|------|----------|------|
| `RegisterGetMethod` | GET | 查询操作 |
| `RegisterPostMethod` | POST | 创建操作 |
| `RegisterPutMethod` | PUT | 更新操作 |
| `RegisterDeleteMethod` | DELETE | 删除操作 |
| `RegisterPatchMethod` | PATCH | 部分更新 |
| `RegisterAnyMethod` | ANY | 任意方法 |

### 参数注解

| 方法 | 用途 | 示例 |
|------|------|------|
| `WithPathParam(name, required)` | 路径参数 | `/{id}` |
| `WithQueryParam(name, required, default)` | 查询参数 | `?page=1` |
| `WithBodyParam(required)` | 请求体 | JSON/XML |
| `WithHeaderParam(name, required, default)` | 请求头 | `Authorization` |
| `WithCookieParam(name, required, default)` | Cookie | `session_id` |

## 高级用法

### 1. 混合使用注解和传统路由

```go
app := annotation.NewAppWithAnnotations(h.Engine)

// struct标签注解方式注册
app.AutoRegister(&APIController{})

// 传统方式注册
app.AutoRouters(&TraditionalController{})
```

### 2. 与comment包混合使用

```go
import (
    "github.com/zsy619/yyhertz/framework/mvc/annotation"
    "github.com/zsy619/yyhertz/framework/mvc/comment"
)

// 可以同时使用struct标签注解和Go注释注解
type HybridController struct {
    core.BaseController `rest:"" mapping:"/api/hybrid"`
}

func init() {
    // struct标签 + init()注册方式
    hybridType := reflect.TypeOf((*HybridController)(nil)).Elem()
    annotation.RegisterGetMethod(hybridType, "GetByInit", "/init")
}

// @GetMapping("/comment") 
// @Description("通过Go注释注解的方法")
func (c *HybridController) GetByComment() (interface{}, error) {
    return map[string]string{"method": "comment"}, nil
}

func main() {
    h := server.Default()
    
    // 同时使用两种注解系统
    annotationApp := annotation.NewAppWithAnnotations(h.Engine)
    commentApp := comment.NewApp(h.Engine)
    
    controller := &HybridController{}
    
    // 注册到annotation系统（处理struct标签和init()注册）
    annotationApp.AutoRegister(controller)
    
    // 注册到comment系统（处理Go注释注解）
    commentApp.AutoScanAndRegister(controller)
    
    h.Spin()
}
```

### 3. 自定义参数绑定

```go
annotation.RegisterPostMethod(userType, "CreateUser", "/").
    WithBodyParam(true).
    WithHeaderParam("X-Request-ID", false, "").
    WithQueryParam("format", false, "json")
```

### 4. 中间件支持

```go
annotation.RegisterGetMethod(userType, "GetUsers", "/").
    WithMiddleware("auth", "ratelimit").
    WithDescription("需要认证的用户列表接口")
```

### 5. 获取路由信息

```go
routes := app.GetAnnotatedRoutes()
for _, route := range routes {
    fmt.Printf("%s %s -> %s.%s\n", 
        route.HTTPMethod, 
        route.Path, 
        route.ControllerType.Name(), 
        route.MethodName)
}
```

## 与其他系统的兼容性

### 1. 与BaseController系统兼容

这个注解系统与现有的BaseController系统完全兼容：

1. **强制继承BaseController** - 所有controller必须实现`core.IController`接口
2. **模板渲染** - 支持原有的模板渲染机制
3. **数据绑定** - 支持原有的数据绑定方法
4. **生命周期** - 支持Prepare()和Finish()方法

### 2. 与comment包的区别

| 特性 | annotation包 | comment包 |
|------|-------------|-----------|
| **注解方式** | Go struct标签 | Go注释 |
| **语法风格** | `rest:"" mapping:"/path"` | `// @RestController` |
| **方法注册** | init()函数中手动注册 | 自动解析源码注释 |
| **性能** | 编译时确定，运行时高效 | 需要AST解析，稍慢 |
| **IDE支持** | struct标签高亮 | 注释语法高亮 |
| **学习成本** | Go原生语法 | SpringBoot风格 |
| **适用场景** | 简单快速，性能优先 | 复杂配置，可读性优先 |

推荐使用场景：
- **annotation包**：适合简单API，性能敏感的场景
- **comment包**：适合复杂配置，团队熟悉SpringBoot的场景
- **混合使用**：一个项目可以同时使用两种方式

## 最佳实践

### 1. 控制器组织

```go
// API控制器
type APIUserController struct {
    core.BaseController `rest:"" mapping:"/api/v1/users"`
}

// Web控制器  
type WebUserController struct {
    core.BaseController `controller:"" mapping:"/users"`
}

// 管理员控制器
type AdminUserController struct {
    core.BaseController `rest:"" mapping:"/admin/users"`
}
```

### 2. 请求响应结构

```go
// 请求结构
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

// 响应结构
type UserResponse struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// 分页响应
type PagedUsersResponse struct {
    Users    []*UserResponse `json:"users"`
    Page     int            `json:"page"`
    PageSize int            `json:"page_size"`
    Total    int            `json:"total"`
}
```

### 3. 错误处理

```go
func (c *UserController) GetUser() (*UserResponse, error) {
    id := c.GetParam("id")
    
    user, err := c.userService.GetByID(id)
    if err != nil {
        return nil, err // 会自动返回500错误
    }
    
    if user == nil {
        c.Ctx.Status(404)
        return nil, fmt.Errorf("用户不存在")
    }
    
    return user, nil
}
```

## 路由示例

基于上面的配置，将生成以下路由：

```
GET    /api/users          -> UserController.GetUsers
GET    /api/users/{id}     -> UserController.GetUser  
POST   /api/users          -> UserController.CreateUser
PUT    /api/users/{id}     -> UserController.UpdateUser
DELETE /api/users/{id}     -> UserController.DeleteUser
```

## 注意事项

1. **接口继承** - 所有controller必须继承`core.BaseController`以实现`core.IController`接口
2. **init函数** - 方法映射必须在init()函数中注册
3. **类型安全** - 确保方法签名与注册的参数匹配
4. **路径规范** - 路径会自动规范化（添加/删除前导/尾随斜杠）
5. **错误处理** - 方法可以返回error，会自动处理HTTP状态码
6. **性能考虑** - 基于反射实现，运行时高效，适合中大型应用
7. **与comment包配合** - 可以在同一个controller中混合使用两种注解方式

## 未来计划

- [ ] 支持OpenAPI/Swagger文档生成
- [ ] 支持参数验证注解
- [ ] 支持缓存注解
- [ ] 支持事务注解
- [ ] 支持权限注解