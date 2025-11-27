# 💬 注释路由系统

YYHertz框架提供了基于Go注释的注解路由系统，通过解析Go源码中的注释来自动生成RESTful API路由。采用SpringBoot风格的注释注解，让Go代码更加直观和易于维护。这是一种声明式的路由定义方式，可读性强，团队协作友好。

## 特性

- ✅ **Go注释注解** - 使用标准Go注释语法定义路由
- ✅ **自动源码解析** - 基于AST解析Go源文件获取注释信息
- ✅ **SpringBoot风格** - 熟悉的注解语法：`@RestController`、`@GetMapping`等
- ✅ **类型安全** - 强制controller继承`core.IController`接口
- ✅ **参数自动绑定** - 支持路径参数、查询参数、请求体、请求头等
- ✅ **灵活的路由配置** - 支持各种HTTP方法和参数配置
- ✅ **与annotation包兼容** - 可以与struct标签注解混合使用
- ✅ **完整的生命周期** - 支持Prepare()和Finish()方法

## 快速开始

### 1. 定义控制器

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/example/simple/services"
)

// @RestController
// @RequestMapping("/api/users")
// @Description("用户管理API控制器")
type UserController struct {
    mvc.BaseController
    userService *services.UserService
}

// @Controller  
// @RequestMapping("/web/users")
// @Description("用户管理Web控制器")
type WebUserController struct {
    mvc.BaseController
}
```

### 2. 定义API方法

```go
// @GetMapping("/")
// @Description("获取用户列表") 
// @QueryParam("page", false, "1")
// @QueryParam("size", false, "10")
// @QueryParam("keyword", false, "")
func (c *UserController) GetUsers() ([]*UserResponse, error) {
    page := c.GetQuery("page", "1")
    size := c.GetQuery("size", "10")
    keyword := c.GetQuery("keyword", "")
    
    // 业务逻辑...
    users := []*UserResponse{
        {ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25},
    }
    
    return users, nil
}

// @GetMapping("/{id}")
// @Description("获取用户详情")
// @PathParam("id", true)
func (c *UserController) GetUser() (*UserResponse, error) {
    id := c.GetParam("id")
    
    // 业务逻辑...
    user := &UserResponse{
        ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25,
    }
    
    return user, nil
}

// @PostMapping("/")
// @Description("创建用户")
// @BodyParam(true)
func (c *UserController) CreateUser(req *UserRequest) (*UserResponse, error) {
    // req会自动从请求体绑定
    user := &UserResponse{
        ID: 100,
        Name: req.Name,
        Email: req.Email,
        Age: req.Age,
    }
    
    return user, nil
}

// @PutMapping("/{id}")
// @Description("更新用户")
// @PathParam("id", true)
// @BodyParam(true)
func (c *UserController) UpdateUser(req *UserRequest) (*UserResponse, error) {
    id := c.GetParam("id")
    
    // 更新逻辑...
    user := &UserResponse{
        ID: parseInt(id),
        Name: req.Name,
        Email: req.Email,
        Age: req.Age,
    }
    
    return user, nil
}

// @DeleteMapping("/{id}")
// @Description("删除用户")
// @PathParam("id", true)
func (c *UserController) DeleteUser() (map[string]interface{}, error) {
    id := c.GetParam("id")
    
    // 删除逻辑...
    return map[string]interface{}{
        "message": "用户删除成功",
        "id": id,
    }, nil
}
```

### 3. 定义请求响应结构

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
```

### 4. 启动应用

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/comment"
    "github.com/zsy619/yyhertz/example/simple/controllers"
)

func main() {
    // 创建YYHertz应用
    app := mvc.HertzApp
    
    // 创建支持注释注解的应用
    commentApp := comment.NewApp(app.Engine)
    
    // 创建控制器实例
    userController := &controllers.UserController{}
    webController := &controllers.WebUserController{}
    
    // 自动扫描并注册注释注解控制器
    commentApp.AutoScanAndRegister(
        userController,
        webController,
    )
    
    // 可以混合使用传统路由
    app.AutoRouters(userController, webController)
    
    // 启动应用
    app.Run()
}
```

## 支持的注释注解

### 控制器级别注解

| 注解 | 用途 | 示例 |
|------|------|------|
| `@RestController` | REST控制器 | `// @RestController` |
| `@Controller` | MVC控制器 | `// @Controller` |
| `@RequestMapping("/path")` | 基础路径 | `// @RequestMapping("/api/users")` |
| `@Description("说明")` | 控制器描述 | `// @Description("用户管理控制器")` |

### 方法级别注解

| 注解 | HTTP方法 | 用途 | 示例 |
|------|----------|------|------|
| `@GetMapping("/path")` | GET | 查询操作 | `// @GetMapping("/")` |
| `@PostMapping("/path")` | POST | 创建操作 | `// @PostMapping("/")` |
| `@PutMapping("/path")` | PUT | 更新操作 | `// @PutMapping("/{id}")` |
| `@DeleteMapping("/path")` | DELETE | 删除操作 | `// @DeleteMapping("/{id}")` |
| `@PatchMapping("/path")` | PATCH | 部分更新 | `// @PatchMapping("/{id}")` |
| `@RequestMapping("/path", "METHOD")` | 自定义 | 任意方法 | `// @RequestMapping("/test", "OPTIONS")` |

### 参数注解

| 注解 | 用途 | 示例 |
|------|------|------|
| `@PathParam("name", required)` | 路径参数 | `// @PathParam("id", true)` |
| `@QueryParam("name", required, "default")` | 查询参数 | `// @QueryParam("page", false, "1")` |
| `@BodyParam(required)` | 请求体 | `// @BodyParam(true)` |
| `@HeaderParam("name", required, "default")` | 请求头 | `// @HeaderParam("Authorization", true, "")` |
| `@CookieParam("name", required, "default")` | Cookie | `// @CookieParam("session_id", false, "")` |

### 其他注解

| 注解 | 用途 | 示例 |
|------|------|------|
| `@Description("说明")` | 方法描述 | `// @Description("获取用户列表")` |
| `@Middleware("name1", "name2")` | 中间件 | `// @Middleware("auth", "ratelimit")` |
| `@Tag("key", "value")` | 自定义标签 | `// @Tag("version", "v1")` |

## 高级用法

### 1. 复杂参数配置

```go
// @GetMapping("/search")
// @Description("高级搜索用户")
// @QueryParam("keyword", false, "")
// @QueryParam("status", false, "all")
// @QueryParam("page", false, "1")
// @QueryParam("size", false, "10")
// @HeaderParam("X-Request-ID", false, "")
// @Middleware("auth", "ratelimit")
func (c *UserController) SearchUsers() ([]*UserResponse, error) {
    keyword := c.GetQuery("keyword", "")
    status := c.GetQuery("status", "all")
    page := c.GetQuery("page", "1")
    size := c.GetQuery("size", "10")
    requestID := c.GetHeader("X-Request-ID")
    
    // 搜索逻辑...
    return users, nil
}
```

### 2. 混合注解使用

可以与annotation包的struct标签注解混合使用：

```go
// 使用struct标签定义基础路径
type MixedController struct {
    core.BaseController `rest:"" mapping:"/api/mixed"`
}

// 同时使用Go注释注解
// @GetMapping("/comment")
// @Description("通过注释注解的方法")
func (c *MixedController) GetByComment() (map[string]interface{}, error) {
    return map[string]interface{}{
        "message": "通过Go注释注解",
        "method": "GetByComment",
    }, nil
}
```

### 3. 复杂请求体处理

```go
type CreateUserRequest struct {
    Name     string            `json:"name" binding:"required"`
    Email    string            `json:"email" binding:"required,email"`
    Age      int              `json:"age" binding:"min=0,max=120"`
    Tags     []string         `json:"tags"`
    Metadata map[string]string `json:"metadata"`
}

// @PostMapping("/complex")
// @Description("复杂用户创建")
// @BodyParam(true)
// @HeaderParam("Content-Type", false, "application/json")
func (c *UserController) CreateComplexUser(req *CreateUserRequest) (*UserResponse, error) {
    // 复杂业务逻辑...
    user := &UserResponse{
        ID: 200,
        Name: req.Name,
        Email: req.Email,
        Age: req.Age,
    }
    
    return user, nil
}
```

### 4. 获取路由信息

```go
routes := app.GetRoutes()
for _, route := range routes {
    fmt.Printf("%s %s -> %s.%s - %s\n", 
        route.HTTPMethod, 
        route.Path, 
        route.TypeName,
        route.MethodName,
        route.Description)
}
```

## 与其他包的兼容性

### 1. 与annotation包混合使用

```go
// 支持同时使用两种注解方式
app := comment.NewApp(h.Engine)

// Go注释注解
app.AutoScanAndRegister(&CommentController{})

// 可以同时使用annotation包
annotationApp := annotation.NewAppWithAnnotations(h.Engine)
annotationApp.AutoRegister(&AnnotationController{})
```

### 2. 与YYHertz MVC完全兼容

所有controller必须继承`mvc.BaseController`，享受完整的YYHertz框架功能：

```go
// @RestController
// @RequestMapping("/api/v1/users")
type UserController struct {
    mvc.BaseController
    userService *services.UserService
}

// @GetMapping("/{id}")
// @Description("获取用户详情")
// @PathParam("id", true)
func (c *UserController) GetUser() {
    id := c.GetParamInt("id")
    if id == 0 {
        c.Error(400, "用户ID无效")
        return
    }
    
    // 使用YYHertz的业务服务
    user, err := c.userService.GetUserByID(id)
    if err != nil {
        c.LogError("获取用户失败", err)
        c.Error(500, "服务器内部错误")
        return
    }
    
    // 统一响应格式
    c.JSON(map[string]interface{}{
        "code": 200,
        "data": user,
        "message": "获取成功",
    })
}

// 生命周期方法
func (c *UserController) Prepare() {
    // 权限验证
    if !c.CheckAuth() {
        c.Error(401, "未授权访问")
        return
    }
    
    // 初始化服务
    c.userService = services.NewUserService(c.GetDB())
}

func (c *UserController) Finish() {
    // 记录请求日志
    c.LogRequest()
}
```

**支持的YYHertz功能**：
- **模板渲染**：`c.RenderHTML("template.html")`, `c.SetData(key, value)`
- **数据绑定**：`c.GetQuery()`, `c.GetParam()`, `c.ParseJSON()`, `c.ValidateStruct()`
- **响应处理**：`c.JSON()`, `c.String()`, `c.HTML()`, `c.Error()`
- **生命周期**：`Prepare()`, `Finish()`
- **中间件集成**：完全兼容YYHertz中间件系统
- **数据库访问**：`c.GetDB()`, GORM集成
- **日志记录**：`c.LogInfo()`, `c.LogError()`

## 工作原理

### 1. 源码解析

```go
// 解析器会扫描Go源文件
parser := comment.GetGlobalParser()
err := parser.ParseSourceFile("controllers/user.go")

// 获取解析结果
controllerInfo := parser.GetControllerInfo("main", "UserController")
methods := parser.GetControllerMethods("main", "UserController")
```

### 2. 路由注册

```go
// 自动创建处理函数
handler := func(ctx context.Context, c *app.RequestContext) {
    // 创建控制器实例
    controller := &UserController{}
    
    // 初始化BaseController
    controller.Init(enhancedCtx, "UserController", "GetUsers", app)
    
    // 调用Prepare
    controller.Prepare()
    
    // 执行业务方法
    result := controller.GetUsers()
    
    // 处理响应
    c.JSON(200, result)
    
    // 调用Finish
    controller.Finish()
}

// 注册到路由器
engine.GET("/api/users", handler)
```

## 最佳实践

### 1. 控制器组织

```go
// API控制器
// @RestController
// @RequestMapping("/api/v1/users")
type APIUserController struct {
    core.BaseController
}

// Web控制器
// @Controller
// @RequestMapping("/users")
type WebUserController struct {
    core.BaseController
}

// 管理员控制器
// @RestController
// @RequestMapping("/admin/users")
type AdminUserController struct {
    core.BaseController
}
```

### 2. 注释规范

```go
// 控制器注释应该在struct定义前
// @RestController
// @RequestMapping("/api/users")
// @Description("用户管理API控制器")
type UserController struct {
    core.BaseController
}

// 方法注释应该在方法定义前
// @GetMapping("/")
// @Description("获取用户列表，支持分页和搜索")
// @QueryParam("page", false, "1")
// @QueryParam("size", false, "10")
// @QueryParam("keyword", false, "")
func (c *UserController) GetUsers() ([]*UserResponse, error) {
    // 实现...
}
```

### 3. 错误处理

```go
// @GetMapping("/{id}")
// @PathParam("id", true)
func (c *UserController) GetUser() (*UserResponse, error) {
    id := c.GetParam("id")
    
    // 参数验证
    if id == "" {
        return nil, fmt.Errorf("用户ID不能为空")
    }
    
    // 业务逻辑
    user, err := c.userService.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("获取用户失败: %w", err)
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
GET    /api/users          -> UserController.GetUsers - 获取用户列表
GET    /api/users/{id}     -> UserController.GetUser - 获取用户详情
POST   /api/users          -> UserController.CreateUser - 创建用户
PUT    /api/users/{id}     -> UserController.UpdateUser - 更新用户
DELETE /api/users/{id}     -> UserController.DeleteUser - 删除用户
GET    /api/users/search   -> UserController.SearchUsers - 高级搜索用户
POST   /api/users/complex  -> UserController.CreateComplexUser - 复杂用户创建
```

## 注意事项

1. **注释格式** - 必须使用标准的Go注释格式（`//`开头），注解紧挨在struct或方法定义前
2. **注解语法** - 注解必须以`@`开头，遵循SpringBoot风格，大小写敏感
3. **框架集成** - 所有controller必须继承`mvc.BaseController`获得完整YYHertz功能
4. **源码解析** - 需要能够访问Go源码文件进行AST解析，建议在项目根目录运行
5. **路径规范** - 路径会自动规范化和合并，支持RESTful风格参数
6. **性能考虑** - 基于AST解析和反射，启动时解析一次，适合中大型应用
7. **依赖服务** - 支持在控制器中注入业务服务，便于分层架构
8. **错误处理** - 推荐使用统一的错误响应格式
9. **开发友好** - 注释即文档，便于团队协作和API维护

## 全局函数

```go
// 扫描源文件
comment.ScanSourceFile("controllers/user.go")

// 扫描包
comment.ScanPackage("github.com/example/controllers")

// 获取控制器信息
info := comment.GetGlobalControllerInfo("main", "UserController")

// 获取方法信息
method := comment.GetGlobalMethodInfo("main", "UserController", "GetUsers")

// 列出所有注解
controllers, methods := comment.ListAllAnnotations()
```

## 未来计划

- [ ] 支持更多SpringBoot注解
- [ ] 支持注解继承和组合
- [ ] 支持条件注解（@ConditionalOnProperty等）
- [ ] 支持OpenAPI文档自动生成
- [ ] 支持注解验证（@Valid, @NotNull等）
- [ ] 支持缓存注解（@Cacheable等）
- [ ] 支持事务注解（@Transactional等）
- [ ] 支持异步处理注解（@Async等）