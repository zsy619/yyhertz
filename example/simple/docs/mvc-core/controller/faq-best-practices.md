# ❓ FAQ与最佳实践

本文档汇总了YYHertz控制器开发中的常见问题和最佳实践指南。

## 🤔 常见问题解答

### 1. 控制器初始化相关

**Q: 控制器初始化时出现空指针错误怎么办？**
A: 检查以下几点：
- 确保在Init方法中正确设置了Ctx
- 验证依赖的Service是否正确注入
- 检查中间件的执行顺序

```go
func (c *UserController) Init(ctx *context.Context, controllerName, actionName string, app any) {
    // 必须调用父类初始化方法
    c.BaseController.Init(ctx, controllerName, actionName, app)
    
    // 初始化业务依赖
    if c.userService == nil {
        c.userService = app.(*Application).UserService
    }
}
```

**Q: Prepare方法什么时候被调用？**
A: Prepare方法在Action方法执行之前自动调用，用于预处理逻辑。

```go
func (c *UserController) Prepare() {
    // 🔥 调用父类方法是最佳实践
    c.BaseController.Prepare()
    
    // 添加控制器特定的预处理逻辑
    c.checkAuthentication()
    c.loadCommonData()
}
```

### 2. 参数获取和验证

**Q: 如何安全地获取请求参数？**
A: 使用类型安全的获取方法，并提供默认值：

```go
func (c *UserController) GetUserList() {
    // ✅ 推荐：提供默认值
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    
    // ❌ 避免：不安全的类型转换
    // pageStr := c.GetString("page")
    // page, _ := strconv.Atoi(pageStr) // 可能panic
    
    // ✅ 范围验证
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
}
```

**Q: 复杂数据结构如何验证？**
A: 使用结构体绑定和验证标签：

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
    Tags     []string `json:"tags" validate:"max=10,dive,required"`
}

func (c *UserController) PostCreate() {
    var req CreateUserRequest
    
    // 绑定数据
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("数据格式错误: " + err.Error())
        return
    }
    
    // 验证数据
    if err := c.ValidateStruct(&req); err != nil {
        c.JSONBadRequest("数据验证失败: " + err.Error())
        return
    }
}
```

### 3. Session和Cookie问题

**Q: Session数据丢失怎么解决？**
A: 常见原因和解决方案：

```go
// ✅ 正确的Session使用
func (c *AuthController) PostLogin() {
    // 1. 确保在设置Session前未输出内容
    c.SetSession("user_id", user.ID)
    c.SetSession("username", user.Username)
    
    // 2. 检查Session配置
    if c.app.Config.Session.Provider == "" {
        c.LogError("Session provider未配置")
    }
    
    // 3. 分布式环境使用Redis等共享存储
    c.JSONSuccess("登录成功", user)
}

// ❌ 错误示例
func (c *AuthController) BadLogin() {
    c.JSONSuccess("登录成功", user) // 已输出响应
    c.SetSession("user_id", user.ID) // 此时设置Session可能失败
}
```

**Q: Cookie设置后无法读取？**
A: 检查Cookie配置：

```go
func (c *BaseController) setSecureCookie() {
    options := &cookie.Options{
        Path:     "/",           // 确保路径正确
        HttpOnly: true,          // 提高安全性
        Secure:   c.isHTTPS(),   // HTTPS环境启用
        SameSite: "Lax",         // 跨站请求策略
        MaxAge:   3600,          // 有效期
    }
    
    c.SetCookie("session_id", "value", options)
}
```

### 4. 模板渲染问题

**Q: 模板找不到或渲染失败？**
A: 检查模板路径和数据：

```go
func (c *PageController) GetIndex() {
    // ✅ 检查模板是否存在
    templatePath := "home/index.html"
    if !c.templateExists(templatePath) {
        c.LogError("模板不存在", map[string]any{"template": templatePath})
        c.Error(500, "页面模板缺失")
        return
    }
    
    // ✅ 设置必要数据
    c.SetData("Title", "首页")
    c.SetData("User", c.getCurrentUser()) // 确保数据不为nil
    
    // ✅ 设置布局
    c.SetLayout("layouts/main.html")
    
    c.Render(templatePath)
}
```

**Q: 模板函数报错如何处理？**
A: 安全地定义和使用模板函数：

```go
func (c *BaseController) initTemplateFunc() {
    // ✅ 安全的模板函数
    c.AddTemplateFunc("formatDate", func(t time.Time, layout string) string {
        if t.IsZero() {
            return ""
        }
        return t.Format(layout)
    })
    
    // ✅ 处理空值情况
    c.AddTemplateFunc("safeString", func(s *string) string {
        if s == nil {
            return ""
        }
        return *s
    })
}
```

### 5. 性能优化问题

**Q: 控制器响应慢如何优化？**
A: 多方面优化策略：

```go
func (c *UserController) GetProfile() {
    userID := c.GetParam("id")
    
    // ✅ 使用缓存
    cacheKey := fmt.Sprintf("user_profile:%s", userID)
    if cached := c.getFromCache(cacheKey); cached != nil {
        c.JSONSuccess("获取成功", cached)
        return
    }
    
    // ✅ 并发获取数据
    var user *User
    var posts []Post
    var followers int
    
    var wg sync.WaitGroup
    wg.Add(3)
    
    go func() {
        defer wg.Done()
        user, _ = c.userService.GetUser(userID)
    }()
    
    go func() {
        defer wg.Done()
        posts, _ = c.postService.GetUserPosts(userID, 5)
    }()
    
    go func() {
        defer wg.Done()
        followers, _ = c.followService.GetFollowerCount(userID)
    }()
    
    wg.Wait()
    
    result := map[string]any{
        "user": user,
        "recent_posts": posts,
        "follower_count": followers,
    }
    
    // ✅ 缓存结果
    c.setToCache(cacheKey, result, 5*time.Minute)
    
    c.JSONSuccess("获取成功", result)
}
```

## 🎯 最佳实践指南

### 1. 控制器设计原则

#### 单一职责原则
```go
// ✅ 好的设计：每个控制器负责一个资源
type UserController struct {
    core.BaseController
    userService *UserService
}

type PostController struct {
    core.BaseController
    postService *PostService
}

// ❌ 避免：一个控制器处理多种不相关资源
type MixedController struct {
    core.BaseController
    // 避免在一个控制器中混合多种业务
}
```

#### 依赖注入
```go
// ✅ 推荐：通过构造函数或初始化方法注入依赖
func NewUserController(userService *UserService) *UserController {
    return &UserController{
        userService: userService,
    }
}

// ❌ 避免：在方法内部直接创建依赖
func (c *UserController) BadMethod() {
    service := &UserService{} // 紧耦合，难以测试
}
```

### 2. 错误处理最佳实践

#### 统一错误处理
```go
type ErrorHandler struct{}

func (e *ErrorHandler) HandleError(c *core.BaseController, err error) {
    switch e := err.(type) {
    case *ValidationError:
        c.JSONBadRequest(e.Error())
    case *NotFoundError:
        c.JSONNotFound(e.Error())
    case *PermissionError:
        c.JSONForbidden(e.Error())
    default:
        c.LogError("未处理的错误", map[string]any{"error": err.Error()})
        c.JSONInternalServerError("系统错误")
    }
}

// 使用示例
func (c *UserController) GetUser() {
    user, err := c.userService.GetUser(userID)
    if err != nil {
        c.errorHandler.HandleError(c, err)
        return
    }
    
    c.JSONSuccess("获取成功", user)
}
```

#### 错误恢复机制
```go
func (c *BaseController) Prepare() {
    defer func() {
        if r := recover(); r != nil {
            c.LogError("Panic recovered", map[string]any{
                "error": r,
                "stack": string(debug.Stack()),
            })
            c.JSONInternalServerError("服务暂时不可用")
        }
    }()
    
    // 正常的预处理逻辑
}
```

### 3. 安全最佳实践

#### 输入验证
```go
func (c *BaseController) sanitizeInput(input string) string {
    // 1. HTML转义
    input = html.EscapeString(input)
    
    // 2. 移除危险字符
    input = strings.ReplaceAll(input, "<script", "&lt;script")
    
    // 3. 长度限制
    if len(input) > 1000 {
        input = input[:1000]
    }
    
    return input
}
```

#### 权限控制
```go
func (c *BaseController) CheckResourcePermission(resourceType, resourceID string, action string) bool {
    userID := c.GetCurrentUserID()
    
    // 1. 检查基础权限
    if !c.permissionService.HasPermission(userID, resourceType, action) {
        return false
    }
    
    // 2. 检查资源所有权（如果需要）
    if action == "edit" || action == "delete" {
        return c.resourceService.IsOwner(resourceID, userID)
    }
    
    return true
}
```

### 4. 测试最佳实践

#### 控制器单元测试
```go
func TestUserController_GetUser(t *testing.T) {
    // 设置测试环境
    mockService := &MockUserService{}
    controller := &UserController{
        userService: mockService,
    }
    
    // 模拟请求上下文
    ctx := &context.Context{
        Params: map[string]string{"id": "123"},
    }
    controller.Init(ctx, "User", "GetUser", nil)
    
    // 设置期望
    mockService.On("GetUser", "123").Return(&User{ID: 123}, nil)
    
    // 执行测试
    controller.GetUser()
    
    // 验证结果
    assert.Equal(t, 200, ctx.ResponseCode)
    mockService.AssertExpectations(t)
}
```

### 5. 性能优化指南

#### 数据库查询优化
```go
func (c *UserController) GetUserList() {
    // ✅ 使用分页，避免全量查询
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    
    // ✅ 只查询需要的字段
    users, err := c.userService.GetUserList(&UserQuery{
        Page:     page,
        PageSize: pageSize,
        Fields:   []string{"id", "username", "email", "created_at"},
    })
    
    // ✅ 预加载关联数据
    c.userService.PreloadRelations(users, []string{"profile", "roles"})
    
    c.JSONPage("获取成功", users, total)
}
```

#### 缓存策略
```go
func (c *BaseController) getCachedData(key string, fetchFunc func() (interface{}, error)) (interface{}, error) {
    // 1. L1缓存：内存缓存
    if data := c.memoryCache.Get(key); data != nil {
        return data, nil
    }
    
    // 2. L2缓存：Redis缓存
    if data := c.redisCache.Get(key); data != nil {
        c.memoryCache.Set(key, data, 1*time.Minute)
        return data, nil
    }
    
    // 3. 从数据源获取
    data, err := fetchFunc()
    if err != nil {
        return nil, err
    }
    
    // 4. 写入缓存
    c.redisCache.Set(key, data, 10*time.Minute)
    c.memoryCache.Set(key, data, 1*time.Minute)
    
    return data, nil
}
```

## 📚 推荐学习资源

### 1. 深入学习建议
- **阅读源码** - 理解BaseController的实现原理
- **实践项目** - 通过实际项目加深理解
- **性能测试** - 使用压力测试工具验证性能
- **安全审计** - 定期进行安全代码审查

### 2. 开发工具推荐
- **IDE插件** - 使用支持Go的IDE和插件
- **调试工具** - 熟练使用debugger和profiler
- **测试工具** - 单元测试、集成测试工具
- **监控工具** - APM和日志分析工具

### 3. 社区资源
- **官方文档** - 定期查看最新文档更新
- **GitHub Issues** - 参与讨论和问题反馈
- **技术博客** - 关注相关技术文章
- **开源项目** - 参考优秀的开源实现

通过遵循这些最佳实践，您可以构建出高质量、可维护、安全的YYHertz应用程序。