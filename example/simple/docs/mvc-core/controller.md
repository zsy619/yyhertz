# 🎮 控制器详解

控制器是MVC架构的核心组件，负责处理请求和响应逻辑。

## 基础概念

### 控制器结构

所有控制器都需要嵌入 `mvc.BaseController`：

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
)

type UserController struct {
    mvc.BaseController
}
```

### 方法命名规则

控制器方法遵循RESTful命名约定：

| HTTP方法 | 控制器方法 | 说明 |
|----------|------------|------|
| GET | Get* | 获取资源 |
| POST | Post* | 创建资源 |
| PUT | Put* | 更新资源 |
| DELETE | Delete* | 删除资源 |

## 请求处理

### 获取请求参数

```go
func (c *UserController) GetUser() {
    // 获取URL参数
    id := c.GetParam("id")
    
    // 获取查询参数
    page := c.GetString("page")
    size := c.GetInt("size")
    
    // 获取表单数据
    name := c.GetForm("name")
    email := c.GetForm("email")
    
    // 获取JSON数据
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.Error(400, "Invalid JSON")
        return
    }
}
```

### 请求头和Cookies

```go
func (c *UserController) GetProfile() {
    // 获取请求头
    userAgent := c.GetHeader("User-Agent")
    authorization := c.GetHeader("Authorization")
    
    // 获取Cookie
    sessionId := c.GetCookie("session_id")
    
    // 设置Cookie
    c.SetCookie("user_id", "123", 3600, "/", "", false, true)
}
```

## 响应处理

### JSON响应

```go
func (c *UserController) GetUserList() {
    users := []User{
        {ID: 1, Name: "张三", Email: "zhang@example.com"},
        {ID: 2, Name: "李四", Email: "li@example.com"},
    }
    
    c.JSON(map[string]interface{}{
        "code": 200,
        "data": users,
        "message": "success",
    })
}
```

### HTML响应

```go
func (c *UserController) GetUserProfile() {
    user := getUserById(123)
    
    c.SetData("Title", "用户资料")
    c.SetData("User", user)
    c.RenderHTML("user/profile.html")
}
```

### 文件响应

```go
func (c *UserController) DownloadAvatar() {
    userId := c.GetParam("id")
    filePath := fmt.Sprintf("/uploads/avatars/%s.jpg", userId)
    
    c.File(filePath)
}

func (c *UserController) UploadAvatar() {
    file, header, err := c.GetFile("avatar")
    if err != nil {
        c.Error(400, "File upload failed")
        return
    }
    defer file.Close()
    
    // 保存文件
    savePath := fmt.Sprintf("/uploads/avatars/%s", header.Filename)
    c.SaveFile(file, savePath)
    
    c.JSON(map[string]interface{}{
        "message": "Upload successful",
        "filename": header.Filename,
    })
}
```

## 数据验证

### 基础验证

```go
func (c *UserController) PostCreate() {
    name := c.GetForm("name")
    email := c.GetForm("email")
    age := c.GetInt("age")
    
    // 验证必填字段
    if name == "" {
        c.Error(400, "姓名不能为空")
        return
    }
    
    if email == "" {
        c.Error(400, "邮箱不能为空")
        return
    }
    
    // 验证格式
    if !isValidEmail(email) {
        c.Error(400, "邮箱格式不正确")
        return
    }
    
    if age < 18 || age > 100 {
        c.Error(400, "年龄必须在18-100之间")
        return
    }
    
    // 创建用户
    user := &User{
        Name:  name,
        Email: email,
        Age:   age,
    }
    
    if err := createUser(user); err != nil {
        c.Error(500, "创建用户失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "message": "用户创建成功",
        "user": user,
    })
}
```

### 结构体验证

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,min=18,max=100"`
}

func (c *UserController) PostCreateWithValidation() {
    var req CreateUserRequest
    
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "JSON格式错误")
        return
    }
    
    if err := c.Validate(&req); err != nil {
        c.Error(400, err.Error())
        return
    }
    
    // 创建用户逻辑...
}
```

## 错误处理

### 自定义错误

```go
func (c *UserController) GetUser() {
    id := c.GetParam("id")
    
    user, err := getUserById(id)
    if err != nil {
        switch err {
        case ErrUserNotFound:
            c.Error(404, "用户不存在")
        case ErrDatabaseError:
            c.Error(500, "数据库错误")
        default:
            c.Error(500, "服务器内部错误")
        }
        return
    }
    
    c.JSON(user)
}
```

### 全局错误处理

```go
func (c *BaseController) HandleError(err error) {
    switch e := err.(type) {
    case *ValidationError:
        c.Error(400, e.Message)
    case *NotFoundError:
        c.Error(404, e.Message)
    case *AuthError:
        c.Error(401, e.Message)
    default:
        c.Error(500, "服务器内部错误")
    }
}
```

## 控制器组织

### 按功能分组

```
controllers/
├── auth/
│   ├── login_controller.go
│   └── register_controller.go
├── user/
│   ├── profile_controller.go
│   └── settings_controller.go
├── admin/
│   ├── dashboard_controller.go
│   └── users_controller.go
└── api/
    ├── v1/
    └── v2/
```

### 基础控制器

```go
// controllers/base_controller.go
type BaseController struct {
    mvc.BaseController
}

func (c *BaseController) RequireAuth() *User {
    token := c.GetHeader("Authorization")
    if token == "" {
        c.Error(401, "需要登录")
        return nil
    }
    
    user, err := validateToken(token)
    if err != nil {
        c.Error(401, "无效的token")
        return nil
    }
    
    return user
}

func (c *BaseController) RequireAdmin() *User {
    user := c.RequireAuth()
    if user == nil {
        return nil
    }
    
    if !user.IsAdmin {
        c.Error(403, "需要管理员权限")
        return nil
    }
    
    return user
}
```

## 最佳实践

### 1. 保持控制器精简

```go
// 好的做法 ✅
func (c *UserController) GetUser() {
    id := c.GetParam("id")
    user, err := c.userService.GetById(id)
    if err != nil {
        c.HandleError(err)
        return
    }
    c.JSON(user)
}

// 避免的做法 ❌
func (c *UserController) GetUser() {
    // 大量的业务逻辑代码...
    // 数据库操作...
    // 复杂的计算...
}
```

### 2. 使用服务层

```go
type UserController struct {
    mvc.BaseController
    userService *services.UserService
}

func NewUserController() *UserController {
    return &UserController{
        userService: services.NewUserService(),
    }
}
```

### 3. 统一响应格式

```go
type APIResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func (c *BaseController) Success(data interface{}) {
    c.JSON(APIResponse{
        Code:    200,
        Message: "success",
        Data:    data,
    })
}

func (c *BaseController) Fail(code int, message string) {
    c.JSON(APIResponse{
        Code:    code,
        Message: message,
    })
}
```

## 高级特性

### 中间件集成

```go
func (c *UserController) GetProfile() {
    // 中间件已经验证了用户身份
    user := c.GetData("current_user").(*User)
    
    c.SetData("Title", "个人资料")
    c.SetData("User", user)
    c.RenderHTML("user/profile.html")
}
```

### 依赖注入

```go
type UserController struct {
    mvc.BaseController
    UserService  services.UserServiceInterface  `inject:""`
    EmailService services.EmailServiceInterface `inject:""`
    Logger       logger.LoggerInterface         `inject:""`
}
```

---

控制器是连接HTTP请求和业务逻辑的桥梁，掌握好控制器的使用是开发高质量Web应用的基础！