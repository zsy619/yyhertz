# 🎮 YYHertz 控制器完全指南

YYHertz的BaseController是一个功能强大、高度模块化的控制器基类，提供了Web开发所需的全部功能。本文档将深入解析其架构设计、方法分组、使用示例和最佳实践。

## 🏗️ 控制器架构概览

### 核心结构设计

BaseController采用模块化设计，将功能划分为13个专门模块：

```go
type BaseController struct {
    // 🔧 核心组件
    Ctx            *context.Context     // 统一上下文
    ControllerName string               // 控制器名称
    ActionName     string               // 当前动作名称
    
    // 🎨 模板系统
    ViewPath       string               // 视图路径
    Layout         string               // 布局文件
    Data           map[string]any       // 模板数据
    TplFuncs       template.FuncMap     // 自定义模板函数
    
    // 🛣️ 路由管理
    RoutePattern   string                      // 路由模式
    RouteParams    map[string]string           // 路由参数
    URLGenerator   func(string, ...any) string // URL生成函数
    
    // 🛠️ 辅助工具
    cookieHelper   *cookie.Helper              // Cookie助手
    sessionHelper  *session.Manager            // Session管理
    templateEngine *view.TemplateEngine        // 模板引擎
    
    // ⚡ 优化特性
    optimizationEnabled bool               // 优化特性开关
    middlewareList      []string           // 中间件列表
}
```

### 功能模块分布

| 模块 | 文件 | 主要功能 | 方法数量 |
|------|------|----------|----------|
| 🔄 **生命周期** | controller.go | 初始化、资源管理 | 8个 |
| 📥 **请求处理** | controller_request.go | 参数获取、类型转换 | 15个 |
| 📤 **响应输出** | controller_response.go | JSON、HTML、文件输出 | 16个 |
| 🎨 **模板引擎** | controller_template.go | 模板渲染、布局管理 | 18个 |
| 🍪 **Session/Cookie** | controller_session.go<br/>controller_cookie.go | 会话和Cookie管理 | 10个 |
| 🛣️ **路由管理** | controller_routing.go | URL映射、路由参数 | 14个 |
| 📊 **数据管理** | controller_data.go | 模板数据存储 | 3个 |
| 🔒 **安全功能** | controller_security.go | XSRF防护、安全Cookie | 6个 |
| 🚦 **流程控制** | controller_flow.go | 请求流程控制 | 3个 |
| 🛠️ **工具方法** | controller_utils.go | 反射工具、方法映射 | 6个 |

---

## 🔄 生命周期管理

### 核心生命周期方法

```go
// 1️⃣ 初始化控制器
func (c *BaseController) Init(ct *context.Context, controllerName, actionName string, app any)

// 2️⃣ 预处理（可重写）
func (c *BaseController) Prepare()

// 3️⃣ 执行业务逻辑（自动路由到具体方法）
// GetUser(), PostCreate(), PutUpdate(), DeleteUser() 等

// 4️⃣ 后处理（可重写）
func (c *BaseController) Finish()

// 5️⃣ 资源清理
func (c *BaseController) Destroy() error
```

### 完整生命周期示例

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc/core"
)

type UserController struct {
    core.BaseController
    userService *UserService // 业务服务
}

// 预处理 - 每个请求前执行
func (c *UserController) Prepare() {
    // 🔐 身份验证
    if !c.IsAuthenticated() {
        c.JSONError("未登录")
        c.StopRun() // 终止后续处理
        return
    }
    
    // 📊 设置通用数据
    c.SetData("CurrentUser", c.GetCurrentUser())
    c.SetData("Timestamp", time.Now())
    
    // 🛡️ 启用XSRF保护（POST/PUT/DELETE请求）
    if c.IsPost() || c.IsPut() || c.IsDelete() {
        c.EnableXSRF(3600) // 1小时有效期
        if !c.CheckXSRFCookie() {
            c.JSONError("XSRF令牌验证失败")
            c.StopRun()
            return
        }
    }
    
    // 📝 记录请求日志
    c.LogInfo("处理请求", map[string]any{
        "method": c.Ctx.Method(),
        "path":   c.Ctx.Path(),
        "ip":     c.GetClientIP(),
    })
}

// 后处理 - 每个请求后执行
func (c *UserController) Finish() {
    // ⏱️ 记录响应时间
    duration := time.Since(c.GetData("start_time").(time.Time))
    c.LogInfo("请求完成", map[string]any{
        "duration": duration,
        "status":   c.Ctx.Response.StatusCode(),
    })
    
    // 🧹 清理敏感数据
    c.DelData("password")
    c.DelData("token")
    
    // 🔄 调用父类方法进行资源清理
    c.BaseController.Finish()
}
```

---

## 📥 请求处理完全指南

### 基础参数获取 (15个方法)

| 方法 | 说明 | 示例 |
|------|------|------|
| `GetString(key, def...)` | 获取字符串参数 | `name := c.GetString("name", "默认值")` |
| `GetInt(key, def...)` | 获取整型参数 | `age := c.GetInt("age", 18)` |
| `GetFloat(key, def...)` | 获取浮点参数 | `price := c.GetFloat("price", 0.0)` |
| `GetBool(key, def...)` | 获取布尔参数 | `active := c.GetBool("active", true)` |
| `GetParam(key)` | 获取路径参数 | `id := c.GetParam("id")` |
| `GetQuery(key, def...)` | 获取查询参数 | `page := c.GetQuery("page", "1")` |
| `GetForm(key, def...)` | 获取表单参数 | `email := c.GetForm("email")` |

### 高级请求处理

```go
// 🔍 请求信息检查
func (c *UserController) GetUserInfo() {
    // 1️⃣ 基础信息获取
    userAgent := c.GetUserAgent()
    clientIP := c.GetClientIP()
    
    // 2️⃣ 请求方法判断
    if c.IsAjax() {
        c.LogInfo("Ajax请求")
    }
    
    // 3️⃣ HTTP方法检查
    switch {
    case c.IsGet():
        c.handleGetRequest()
    case c.IsPost():
        c.handlePostRequest()
    case c.IsPut():
        c.handlePutRequest()
    case c.IsDelete():
        c.handleDeleteRequest()
    default:
        c.Error(405, "不支持的HTTP方法")
    }
}

// 📋 复杂参数绑定
func (c *UserController) PostCreate() {
    // 4️⃣ JSON数据绑定
    var userReq CreateUserRequest
    if err := c.BindJSON(&userReq); err != nil {
        c.JSONError("JSON格式错误: " + err.Error())
        return
    }
    
    // 5️⃣ 数据验证
    if err := c.ValidateStruct(&userReq); err != nil {
        c.JSONError("数据验证失败: " + err.Error())
        return
    }
    
    // 6️⃣ 处理业务逻辑
    user, err := c.userService.CreateUser(&userReq)
    if err != nil {
        c.JSONError("创建用户失败: " + err.Error())
        return
    }
    
    c.JSONSuccess("用户创建成功", user)
}

// 🔨 自定义验证规则
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
    Password string `json:"password" validate:"required,min=8"`
}
```

### 请求头处理

```go
func (c *BaseController) HandleHeaders() {
    // 🔑 认证头处理
    authHeader := c.GetHeader("Authorization")
    if strings.HasPrefix(authHeader, "Bearer ") {
        token := authHeader[7:]
        user, err := c.validateJWTToken(token)
        if err != nil {
            c.Error(401, "无效的认证令牌")
            return
        }
        c.SetData("current_user", user)
    }
    
    // 🌐 内容协商
    accept := c.GetHeader("Accept")
    switch {
    case strings.Contains(accept, "application/json"):
        c.handleJSONRequest()
    case strings.Contains(accept, "text/html"):
        c.handleHTMLRequest()
    case strings.Contains(accept, "application/xml"):
        c.handleXMLRequest()
    default:
        c.handleDefaultRequest()
    }
    
    // 📱 客户端信息
    if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
        c.LogInfo("Ajax请求检测")
    }
}
```

---

## 📤 响应输出系统

### JSON响应方法 (16个方法)

| 方法 | 说明 | 示例 |
|------|------|------|
| `JSON(data)` | 标准JSON响应 | `c.JSON(user)` |
| `JSONSuccess(msg, data)` | 成功响应 | `c.JSONSuccess("操作成功", result)` |
| `JSONError(msg)` | 错误响应 | `c.JSONError("参数错误")` |
| `JSONPage(msg, data, count)` | 分页响应 | `c.JSONPage("查询成功", users, 100)` |
| `JSONOK(data)` | 200状态响应 | `c.JSONOK(data)` |
| `JSONBadRequest(msg)` | 400错误 | `c.JSONBadRequest("请求参数错误")` |
| `JSONUnauthorized(msg)` | 401错误 | `c.JSONUnauthorized("未授权访问")` |
| `JSONForbidden(msg)` | 403错误 | `c.JSONForbidden("禁止访问")` |
| `JSONNotFound(msg)` | 404错误 | `c.JSONNotFound("资源未找到")` |
| `JSONInternalServerError(msg)` | 500错误 | `c.JSONInternalServerError("服务器错误")` |

### 统一响应格式

```go
// 🎯 标准API响应格式
type APIResponse struct {
    Code      int         `json:"code"`      // 业务状态码
    Message   string      `json:"message"`   // 响应消息
    Data      interface{} `json:"data,omitempty"`      // 数据内容
    Timestamp int64       `json:"timestamp"` // 时间戳
    RequestID string      `json:"request_id,omitempty"` // 请求ID
}

// 📊 分页响应格式
type PageResponse struct {
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data"`
    Total     int64       `json:"total"`     // 总记录数
    Page      int         `json:"page"`      // 当前页码
    PageSize  int         `json:"page_size"` // 每页大小
    TotalPage int         `json:"total_page"` // 总页数
    Timestamp int64       `json:"timestamp"`
}

// 🔧 自定义响应方法
func (c *BaseController) CustomAPIResponse(code int, message string, data interface{}) {
    response := APIResponse{
        Code:      code,
        Message:   message,
        Data:      data,
        Timestamp: time.Now().Unix(),
        RequestID: c.GetRequestID(), // 自定义方法获取请求ID
    }
    c.JSON(response)
}
```

### 文件响应处理

```go
func (c *FileController) HandleFiles() {
    action := c.GetParam("action")
    
    switch action {
    case "download":
        c.downloadFile()
    case "upload":
        c.uploadFile()
    case "preview":
        c.previewFile()
    default:
        c.JSONNotFound("不支持的操作")
    }
}

// 📥 文件下载
func (c *FileController) downloadFile() {
    fileID := c.GetParam("id")
    
    // 🔍 查找文件信息
    fileInfo, err := c.fileService.GetFileInfo(fileID)
    if err != nil {
        c.JSONNotFound("文件不存在")
        return
    }
    
    // 🔐 权限检查
    if !c.hasFileAccess(fileInfo) {
        c.JSONForbidden("无权访问该文件")
        return
    }
    
    // 📁 设置响应头
    c.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name))
    c.SetHeader("Content-Type", "application/octet-stream")
    
    // 📤 发送文件
    c.File(fileInfo.Path)
}

// 📤 文件上传
func (c *FileController) uploadFile() {
    // 🔍 获取上传文件
    file, header, err := c.GetFile("file")
    if err != nil {
        c.JSONBadRequest("文件上传失败: " + err.Error())
        return
    }
    defer file.Close()
    
    // ✅ 文件验证
    if header.Size > 10*1024*1024 { // 10MB限制
        c.JSONBadRequest("文件大小不能超过10MB")
        return
    }
    
    allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"}
    if !c.isAllowedFileType(header.Filename, allowedTypes) {
        c.JSONBadRequest("不支持的文件类型")
        return
    }
    
    // 💾 保存文件
    savePath := c.generateFilePath(header.Filename)
    if err := c.SaveFile(file, savePath); err != nil {
        c.JSONInternalServerError("文件保存失败")
        return
    }
    
    c.JSONSuccess("文件上传成功", map[string]string{
        "filename": header.Filename,
        "path":     savePath,
        "size":     fmt.Sprintf("%d", header.Size),
    })
}
```

---

## 🎨 模板引擎系统

### 模板渲染方法 (18个方法)

| 方法 | 功能描述 | 使用场景 |
|------|----------|----------|
| `RenderHTML(viewName, data...)` | 渲染HTML模板 | 页面渲染 |
| `RenderWithLayout(viewName, layoutName)` | 使用指定布局渲染 | 复杂页面 |
| `RenderTemplate(templateName, data)` | 渲染指定模板 | 组件渲染 |
| `RenderBytes()` | 渲染为字节数组 | 自定义处理 |
| `RenderString()` | 渲染为字符串 | 邮件模板 |
| `SetTplName(name)` | 设置模板名称 | 动态模板 |
| `SetLayout(layout)` | 设置布局文件 | 布局切换 |
| `AddTplFunc(name, func)` | 添加模板函数 | 自定义函数 |

### 完整模板示例

```go
type ProductController struct {
    core.BaseController
    productService *ProductService
}

// 🎨 产品列表页面
func (c *ProductController) GetList() {
    // 1️⃣ 获取查询参数
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("pageSize", 20)
    category := c.GetString("category")
    keyword := c.GetString("keyword")
    
    // 2️⃣ 业务逻辑处理
    products, total, err := c.productService.GetProductList(page, pageSize, category, keyword)
    if err != nil {
        c.SetData("Error", "获取产品列表失败")
        c.RenderHTML("error/500.html")
        return
    }
    
    // 3️⃣ 设置模板数据
    c.SetData("Title", "产品列表")
    c.SetData("Products", products)
    c.SetData("Pagination", map[string]any{
        "Page":      page,
        "PageSize":  pageSize,
        "Total":     total,
        "TotalPage": (total + int64(pageSize) - 1) / int64(pageSize),
    })
    c.SetData("Category", category)
    c.SetData("Keyword", keyword)
    
    // 4️⃣ 添加自定义模板函数
    c.AddTplFunc("formatPrice", func(price float64) string {
        return fmt.Sprintf("￥%.2f", price)
    })
    c.AddTplFunc("formatDate", func(t time.Time) string {
        return t.Format("2006-01-02")
    })
    
    // 5️⃣ 渲染模板
    c.RenderHTML("product/list.html")
}

// 🎯 产品详情页面（带布局）
func (c *ProductController) GetDetail() {
    id := c.GetParam("id")
    
    // 查询产品信息
    product, err := c.productService.GetProductByID(id)
    if err != nil {
        c.SetData("Title", "产品未找到")
        c.SetData("Message", "请求的产品不存在")
        c.RenderHTML("error/404.html")
        return
    }
    
    // 获取相关产品
    relatedProducts, _ := c.productService.GetRelatedProducts(product.CategoryID, 4)
    
    // 设置页面数据
    c.SetData("Title", product.Name)
    c.SetData("Product", product)
    c.SetData("RelatedProducts", relatedProducts)
    c.SetData("PageType", "product-detail")
    
    // 使用专门的产品详情布局
    c.RenderWithLayout("product/detail.html", "layout/product.html")
}

// 📧 邮件模板渲染
func (c *ProductController) sendProductNotification(productID string, userEmail string) error {
    product, err := c.productService.GetProductByID(productID)
    if err != nil {
        return err
    }
    
    // 渲染邮件模板
    c.SetData("Product", product)
    c.SetData("UserEmail", userEmail)
    
    emailContent, err := c.RenderString("email/product_notification.html")
    if err != nil {
        return err
    }
    
    // 发送邮件（伪代码）
    return c.emailService.Send(userEmail, "产品通知", emailContent)
}
```

### 主题和布局系统

```go
// 🎨 动态主题切换
func (c *BaseController) SetTheme(themeName string) {
    if err := c.SetTemplateTheme(themeName); err != nil {
        c.LogError("设置主题失败", err)
        return
    }
    
    // 更新布局路径
    c.SetTemplatePath(
        fmt.Sprintf("themes/%s/views", themeName),
        fmt.Sprintf("themes/%s/layouts", themeName),
    )
    
    // 保存主题偏好到Session
    c.SetSession("user_theme", themeName)
}

// 📱 响应式布局选择
func (c *BaseController) SelectLayout() string {
    userAgent := c.GetUserAgent()
    
    if strings.Contains(userAgent, "Mobile") {
        return "layout/mobile.html"
    } else if strings.Contains(userAgent, "Tablet") {
        return "layout/tablet.html"
    }
    
    return "layout/desktop.html"
}
```

---

## 🍪 Session与Cookie管理

### Session管理方法 (6个方法)

| 方法 | 功能 | 示例 |
|------|------|------|
| `SetSession(key, value)` | 设置Session值 | `c.SetSession("user_id", 123)` |
| `GetSession(key)` | 获取Session值 | `userID := c.GetSession("user_id")` |
| `DeleteSession(key)` | 删除Session项 | `c.DeleteSession("temp_data")` |
| `HasSession(key)` | 检查Session是否存在 | `if c.HasSession("user_id")` |
| `GetSessionID()` | 获取Session ID | `sessionID := c.GetSessionID()` |
| `DestroySession()` | 销毁整个Session | `c.DestroySession()` |

### Cookie操作方法 (4+2个方法)

| 方法 | 功能 | 示例 |
|------|------|------|
| `SetCookie(name, value, options...)` | 设置Cookie | 见下方示例 |
| `GetCookie(name)` | 获取Cookie | `token := c.GetCookie("auth_token")` |
| `DeleteCookie(name, path...)` | 删除Cookie | `c.DeleteCookie("temp_cookie")` |
| `HasCookie(name)` | 检查Cookie存在性 | `if c.HasCookie("remember_me")` |
| `SetSecureCookie(secret, name, value, options...)` | 设置加密Cookie | 安全存储 |
| `GetSecureCookie(secret, name)` | 获取加密Cookie | 安全读取 |

### 完整的用户认证示例

```go
type AuthController struct {
    core.BaseController
    authService *AuthService
}

// 🔐 用户登录
func (c *AuthController) PostLogin() {
    // 1️⃣ 获取登录信息
    username := c.GetForm("username")
    password := c.GetForm("password")
    rememberMe := c.GetBool("remember_me", false)
    
    // 2️⃣ 验证用户凭证
    user, err := c.authService.ValidateCredentials(username, password)
    if err != nil {
        c.JSONError("用户名或密码错误")
        return
    }
    
    // 3️⃣ 设置Session
    c.SetSession("user_id", user.ID)
    c.SetSession("username", user.Username)
    c.SetSession("login_time", time.Now())
    c.SetSession("permissions", user.Permissions)
    
    // 4️⃣ 设置记住我功能
    if rememberMe {
        // 生成记住我令牌
        rememberToken, err := c.authService.GenerateRememberToken(user.ID)
        if err == nil {
            // 设置长期Cookie（30天）
            c.SetCookie("remember_token", rememberToken, &cookie.Options{
                MaxAge:   30 * 24 * 3600, // 30天
                HttpOnly: true,            // 防止XSS
                Secure:   true,           // HTTPS only
                SameSite: "Strict",       // CSRF防护
                Path:     "/",
            })
        }
    }
    
    // 5️⃣ 设置用户偏好Cookie
    c.SetCookie("user_theme", user.Theme, &cookie.Options{
        MaxAge: 365 * 24 * 3600, // 1年
        Path:   "/",
    })
    
    c.JSONSuccess("登录成功", map[string]any{
        "user":         user,
        "redirect_url": c.GetString("redirect", "/dashboard"),
    })
}

// 🚪 用户登出
func (c *AuthController) PostLogout() {
    // 1️⃣ 清理Session数据
    c.DeleteSession("user_id")
    c.DeleteSession("username")
    c.DeleteSession("login_time")
    c.DeleteSession("permissions")
    
    // 2️⃣ 删除记住我Cookie
    c.DeleteCookie("remember_token", "/")
    
    // 3️⃣ 保留用户偏好（主题等）
    // c.DeleteCookie("user_theme", "/") // 可选择保留
    
    // 4️⃣ 记录登出日志
    c.LogInfo("用户登出", map[string]any{
        "session_id": c.GetSessionID(),
        "user_agent": c.GetUserAgent(),
        "ip":         c.GetClientIP(),
    })
    
    c.JSONSuccess("登出成功", nil)
}

// 🔍 检查登录状态
func (c *AuthController) checkAuthentication() *User {
    // 1️⃣ 检查Session
    if userID := c.GetSession("user_id"); userID != nil {
        if user, err := c.authService.GetUserByID(userID.(int)); err == nil {
            return user
        }
    }
    
    // 2️⃣ 检查记住我令牌
    if rememberToken := c.GetCookie("remember_token"); rememberToken != "" {
        if user, err := c.authService.ValidateRememberToken(rememberToken); err == nil {
            // 重新建立Session
            c.SetSession("user_id", user.ID)
            c.SetSession("username", user.Username)
            c.SetSession("login_time", time.Now())
            return user
        }
    }
    
    return nil
}

// 🔐 安全Cookie示例
func (c *AuthController) setSecureUserPreferences(userID int, preferences map[string]any) {
    preferencesJSON, _ := json.Marshal(preferences)
    
    // 使用加密Cookie存储敏感的用户偏好
    secret := "your-secret-key-here" // 应该从配置文件读取
    c.SetSecureCookie(secret, "user_preferences", string(preferencesJSON), 
        &cookie.Options{
            MaxAge:   7 * 24 * 3600, // 7天
            HttpOnly: true,
            Secure:   true,
            SameSite: "Strict",
            Path:     "/",
        })
}
```

---

## 🛣️ 路由管理系统

### 路由管理方法 (14个方法)

| 分类 | 方法 | 功能描述 |
|------|------|----------|
| **方法映射** | `AddMethodMapping(http, method)` | 添加HTTP方法映射 |
|  | `GetMethodMapping()` | 获取方法映射表 |
|  | `SetMethodMapping(mapping)` | 设置完整映射 |
|  | `GetMappedMethod(httpMethod)` | 获取映射的控制器方法 |
| **路由参数** | `SetRouteParam(key, value)` | 设置路由参数 |
|  | `GetRouteParam(key)` | 获取路由参数 |
|  | `GetRouteParams()` | 获取所有路由参数 |
|  | `SetRouteParams(params)` | 设置所有路由参数 |
| **URL构建** | `URLFor(endpoint, values...)` | 构建URL |
|  | `BuildURL(controller, action, params...)` | 构建控制器URL |
| **URL映射** | `AddURLMapping(pattern, method)` | 添加URL映射 |
|  | `GetURLMappings()` | 获取URL映射表 |
| **工具方法** | `HandlerFunc(name)` | 检查处理器函数 |
|  | `URLMapping()` | URL映射配置（可重写） |

### RESTful路由设计

```go
type ProductController struct {
    core.BaseController
    productService *ProductService
}

// 🎯 重写URLMapping配置RESTful路由
func (c *ProductController) URLMapping() {
    // 产品CRUD操作映射
    c.AddMethodMapping("GET", "GetList")      // GET /products
    c.AddMethodMapping("POST", "PostCreate")   // POST /products
    c.AddMethodMapping("PUT", "PutUpdate")     // PUT /products/:id  
    c.AddMethodMapping("DELETE", "DeleteRemove") // DELETE /products/:id
    
    // 自定义URL映射
    c.AddURLMapping("/products", "GetList")
    c.AddURLMapping("/products/create", "GetCreateForm")
    c.AddURLMapping("/products/:id", "GetDetail")
    c.AddURLMapping("/products/:id/edit", "GetEditForm")
    c.AddURLMapping("/products/category/:category", "GetByCategory")
    c.AddURLMapping("/products/search", "GetSearch")
}

// 📋 产品列表 - GET /products
func (c *ProductController) GetList() {
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    category := c.GetString("category")
    
    products, total, err := c.productService.GetProducts(page, pageSize, category)
    if err != nil {
        c.JSONInternalServerError("获取产品列表失败")
        return
    }
    
    c.JSONPage("获取成功", products, total)
}

// ➕ 创建产品 - POST /products  
func (c *ProductController) PostCreate() {
    var req CreateProductRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("请求格式错误")
        return
    }
    
    product, err := c.productService.CreateProduct(&req)
    if err != nil {
        c.JSONInternalServerError("创建产品失败")
        return
    }
    
    // 构建新产品的URL
    productURL := c.BuildURL("Product", "Detail", product.ID)
    
    c.JSONSuccess("产品创建成功", map[string]any{
        "product": product,
        "url":     productURL,
    })
}

// 🔍 产品详情 - GET /products/:id
func (c *ProductController) GetDetail() {
    id := c.GetRouteParam("id")
    if id == "" {
        c.JSONBadRequest("产品ID不能为空")
        return
    }
    
    product, err := c.productService.GetProductByID(id)
    if err != nil {
        c.JSONNotFound("产品不存在")
        return
    }
    
    c.JSONSuccess("获取成功", product)
}

// ✏️ 更新产品 - PUT /products/:id
func (c *ProductController) PutUpdate() {
    id := c.GetRouteParam("id")
    
    var req UpdateProductRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("请求格式错误")
        return
    }
    
    product, err := c.productService.UpdateProduct(id, &req)
    if err != nil {
        c.JSONInternalServerError("更新产品失败")
        return
    }
    
    c.JSONSuccess("产品更新成功", product)
}

// 🗑️ 删除产品 - DELETE /products/:id
func (c *ProductController) DeleteRemove() {
    id := c.GetRouteParam("id")
    
    if err := c.productService.DeleteProduct(id); err != nil {
        c.JSONInternalServerError("删除产品失败")
        return
    }
    
    c.JSONSuccess("产品删除成功", nil)
}
```

### 高级路由功能

```go
// 🎯 动态路由处理
type APIController struct {
    core.BaseController
}

func (c *APIController) URLMapping() {
    // 版本化API路由
    c.AddURLMapping("/api/v1/:resource", "HandleV1")
    c.AddURLMapping("/api/v2/:resource", "HandleV2") 
    c.AddURLMapping("/api/:version/:resource/:action", "HandleDynamic")
}

// 🔄 动态处理不同版本的API
func (c *APIController) HandleDynamic() {
    version := c.GetRouteParam("version")
    resource := c.GetRouteParam("resource")
    action := c.GetRouteParam("action")
    
    // 构建处理器方法名
    methodName := fmt.Sprintf("Handle%s%s%s", 
        strings.Title(version),
        strings.Title(resource), 
        strings.Title(action))
    
    // 动态调用方法
    if c.HandlerFunc(methodName) {
        c.callMethod(methodName)
    } else {
        c.JSONNotFound("API不存在")
    }
}

// 🔗 URL构建工具
func (c *BaseController) BuildResourceURL(resource, action string, params ...any) string {
    if c.URLGenerator != nil {
        endpoint := fmt.Sprintf("%s_%s", resource, action)
        return c.URLGenerator(endpoint, params...)
    }
    
    return c.BuildURL(strings.Title(resource), strings.Title(action), params...)
}

// 🏠 面包屑导航构建
func (c *BaseController) buildBreadcrumb() []map[string]string {
    breadcrumb := []map[string]string{}
    
    // 根据当前路由构建面包屑
    controllerName := strings.ToLower(c.ControllerName)
    actionName := strings.ToLower(c.ActionName)
    
    // 添加首页
    breadcrumb = append(breadcrumb, map[string]string{
        "title": "首页",
        "url":   c.BuildURL("Home", "Index"),
    })
    
    // 添加控制器级别
    if controllerName != "home" {
        breadcrumb = append(breadcrumb, map[string]string{
            "title": c.getControllerTitle(controllerName),
            "url":   c.BuildURL(c.ControllerName, "Index"),
        })
    }
    
    // 添加动作级别
    if actionName != "index" {
        breadcrumb = append(breadcrumb, map[string]string{
            "title": c.getActionTitle(actionName),
            "url":   "", // 当前页面不需要链接
        })
    }
    
    return breadcrumb
}
```

---

## 🔒 安全功能详解

### XSRF/CSRF防护 (6个方法)

| 方法 | 功能 | 使用场景 |
|------|------|----------|
| `XSRFToken()` | 生成XSRF令牌 | 表单防护 |
| `CheckXSRFCookie()` | 验证XSRF令牌 | 请求验证 |
| `EnableXSRF(expire...)` | 启用XSRF防护 | 安全配置 |
| `DisableXSRF()` | 禁用XSRF防护 | 测试环境 |
| `SetSecureCookie(secret, name, value, options...)` | 设置加密Cookie | 敏感数据 |
| `GetSecureCookie(secret, name)` | 获取加密Cookie | 数据读取 |

### 完整的安全实践

```go
type SecureController struct {
    core.BaseController
}

// 🛡️ 预处理：统一安全检查
func (c *SecureController) Prepare() {
    // 1️⃣ HTTPS强制跳转
    if !c.IsHTTPS() && c.isProductionEnv() {
        httpsURL := strings.Replace(c.GetCurrentURL(), "http://", "https://", 1)
        c.Redirect(httpsURL, 301)
        return
    }
    
    // 2️⃣ 设置安全响应头
    c.setSecurityHeaders()
    
    // 3️⃣ IP白名单检查（管理员功能）
    if c.requiresAdminAccess() && !c.isAllowedIP() {
        c.JSONForbidden("IP地址不在允许范围内")
        c.StopRun()
        return
    }
    
    // 4️⃣ 频率限制检查
    if c.isRateLimited() {
        c.JSONWithStatus(429, map[string]string{
            "error": "请求过于频繁，请稍后再试",
        })
        c.StopRun()
        return
    }
    
    // 5️⃣ XSRF保护（非GET请求）
    if !c.IsGet() && !c.IsHead() && !c.IsOptions() {
        c.EnableXSRF(3600) // 1小时有效期
        if !c.CheckXSRFCookie() {
            c.JSONForbidden("XSRF令牌验证失败")
            c.StopRun()
            return
        }
    }
}

// 🔒 设置安全响应头
func (c *SecureController) setSecurityHeaders() {
    // 防止点击劫持
    c.SetHeader("X-Frame-Options", "DENY")
    
    // 防止MIME类型嗅探
    c.SetHeader("X-Content-Type-Options", "nosniff")
    
    // XSS防护
    c.SetHeader("X-XSS-Protection", "1; mode=block")
    
    // HSTS（HTTPS环境）
    if c.IsHTTPS() {
        c.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    }
    
    // Content Security Policy
    csp := "default-src 'self'; " +
           "script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
           "style-src 'self' 'unsafe-inline'; " +
           "img-src 'self' data: https:; " +
           "font-src 'self' https://fonts.gstatic.com"
    c.SetHeader("Content-Security-Policy", csp)
    
    // 引用者策略
    c.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
}

// 🔐 JWT令牌验证
func (c *SecureController) ValidateJWTToken() (*User, error) {
    // 从Authorization头获取令牌
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        return nil, errors.New("缺少认证头")
    }
    
    // 检查Bearer格式
    if !strings.HasPrefix(authHeader, "Bearer ") {
        return nil, errors.New("认证格式错误")
    }
    
    tokenString := authHeader[7:]
    
    // 解析JWT令牌
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // 验证签名方法
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
        }
        return []byte(c.getJWTSecret()), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // 验证令牌声明
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        userID := claims["user_id"].(float64)
        user, err := c.userService.GetUserByID(int(userID))
        if err != nil {
            return nil, err
        }
        
        // 检查令牌黑名单
        if c.isTokenBlacklisted(tokenString) {
            return nil, errors.New("令牌已被吊销")
        }
        
        return user, nil
    }
    
    return nil, errors.New("无效的令牌")
}

// 🚦 频率限制检查
func (c *SecureController) isRateLimited() bool {
    clientIP := c.GetClientIP()
    endpoint := c.ControllerName + "." + c.ActionName
    
    // 从Redis或内存缓存获取访问计数
    key := fmt.Sprintf("rate_limit:%s:%s", clientIP, endpoint)
    
    // 基础频率限制：每分钟60次请求
    count := c.getAccessCount(key)
    if count > 60 {
        return true
    }
    
    // 递增计数器
    c.incrementAccessCount(key, 60) // 60秒过期
    
    return false
}

// 🔍 IP白名单验证
func (c *SecureController) isAllowedIP() bool {
    clientIP := c.GetClientIP()
    
    // 管理员IP白名单
    allowedIPs := []string{
        "127.0.0.1",
        "::1",
        "192.168.1.0/24", // 支持CIDR格式
    }
    
    for _, allowedIP := range allowedIPs {
        if c.matchesIPPattern(clientIP, allowedIP) {
            return true
        }
    }
    
    return false
}

// 🛡️ SQL注入防护示例
func (c *SecureController) sanitizeInput(input string) string {
    // 移除或转义危险字符
    dangerous := []string{"'", "\"", ";", "--", "/*", "*/", "xp_", "drop ", "delete ", "insert ", "update "}
    
    for _, danger := range dangerous {
        input = strings.ReplaceAll(strings.ToLower(input), danger, "")
    }
    
    return input
}

// 📝 安全日志记录
func (c *SecureController) logSecurityEvent(event string, details map[string]any) {
    logData := map[string]any{
        "event":      event,
        "ip":         c.GetClientIP(),
        "user_agent": c.GetUserAgent(),
        "timestamp":  time.Now(),
        "session_id": c.GetSessionID(),
        "details":    details,
    }
    
    c.LogWarn("安全事件", logData)
}
```

---

## 🚦 流程控制方法

### 流程控制方法 (3个方法)

| 方法 | 功能 | 使用场景 |
|------|------|----------|
| `StopRun()` | 停止后续处理 | 中间件拦截 |
| `Abort(code)` | 终止并返回状态码 | 错误处理 |
| `CustomAbort(status, body)` | 自定义终止响应 | 特殊错误页面 |

### 流程控制最佳实践

```go
type FlowController struct {
    core.BaseController
}

// 🔐 认证中间件示例
func (c *FlowController) Prepare() {
    // 检查是否需要认证
    if c.requiresAuth() {
        user := c.getCurrentUser()
        if user == nil {
            // 根据请求类型返回不同响应
            if c.IsAjax() {
                c.JSONUnauthorized("请先登录")
                c.StopRun() // 停止后续处理
                return
            } else {
                // 重定向到登录页面
                loginURL := c.BuildURL("Auth", "Login")
                c.Redirect(loginURL + "?redirect=" + c.GetCurrentURL())
                c.StopRun()
                return
            }
        }
        
        // 设置当前用户到上下文
        c.SetData("CurrentUser", user)
    }
    
    // 权限检查
    if c.requiresPermission() {
        if !c.hasPermission() {
            c.CustomAbort(403, "您没有权限访问此页面")
            return
        }
    }
}

// ⚡ 条件处理示例
func (c *FlowController) GetUserProfile() {
    userID := c.GetParam("id")
    
    // 参数验证
    if userID == "" {
        c.Abort("400") // 直接返回400错误
        return
    }
    
    user, err := c.userService.GetUserByID(userID)
    if err != nil {
        // 根据错误类型选择不同的处理方式
        switch err {
        case ErrUserNotFound:
            if c.IsAjax() {
                c.JSONNotFound("用户不存在")
            } else {
                c.CustomAbort(404, c.render404Page("用户不存在"))
            }
        case ErrPermissionDenied:
            c.CustomAbort(403, c.render403Page("无权访问此用户"))
        default:
            c.CustomAbort(500, c.render500Page("系统错误"))
        }
        return // 这里的return其实可以省略，因为CustomAbort会停止执行
    }
    
    // 正常处理逻辑
    c.handleUserProfile(user)
}

// 🔄 业务流程控制
func (c *FlowController) PostProcessOrder() {
    orderID := c.GetForm("order_id")
    action := c.GetForm("action")
    
    // 获取订单
    order, err := c.orderService.GetOrderByID(orderID)
    if err != nil {
        c.JSONNotFound("订单不存在")
        return
    }
    
    // 根据不同动作进行不同处理
    switch action {
    case "pay":
        c.processPayment(order)
    case "cancel":
        c.processCancellation(order)
    case "refund":
        c.processRefund(order)
    default:
        c.JSONBadRequest("不支持的操作")
        c.StopRun() // 明确停止后续处理
    }
}

// 💳 支付流程控制
func (c *FlowController) processPayment(order *Order) {
    // 1. 预检查
    if order.Status != "pending" {
        c.JSONBadRequest("订单状态不允许支付")
        c.StopRun()
        return
    }
    
    // 2. 风险检查
    if c.isHighRiskTransaction(order) {
        c.LogWarn("高风险交易", map[string]any{
            "order_id": order.ID,
            "amount":   order.Amount,
            "ip":       c.GetClientIP(),
        })
        
        // 需要额外验证
        if !c.validateHighRiskTransaction(order) {
            c.CustomAbort(403, "交易需要额外验证，请联系客服")
            return
        }
    }
    
    // 3. 执行支付
    result, err := c.paymentService.ProcessPayment(order)
    if err != nil {
        c.LogError("支付失败", map[string]any{
            "order_id": order.ID,
            "error":    err.Error(),
        })
        
        c.JSONInternalServerError("支付处理失败，请稍后重试")
        return
    }
    
    // 4. 支付成功处理
    c.JSONSuccess("支付成功", map[string]any{
        "order_id":      order.ID,
        "transaction_id": result.TransactionID,
        "amount":        result.Amount,
    })
}

// 🎯 错误页面渲染
func (c *FlowController) render404Page(message string) string {
    c.SetData("Title", "页面未找到")
    c.SetData("Message", message)
    c.SetData("BackURL", c.GetHeader("Referer"))
    
    content, err := c.RenderString("error/404.html")
    if err != nil {
        return "<h1>404 - 页面未找到</h1><p>" + message + "</p>"
    }
    return content
}
```

---

## 📊 数据管理方法

### 数据管理方法 (3个方法)

| 方法 | 功能 | 示例 |
|------|------|------|
| `SetData(key, value)` | 设置模板数据 | `c.SetData("Title", "首页")` |
| `GetData(key)` | 获取模板数据 | `title := c.GetData("Title")` |
| `DelData(key)` | 删除模板数据 | `c.DelData("sensitive_data")` |

### 数据管理最佳实践

```go
// 🎯 统一数据设置
func (c *BaseController) setCommonData() {
    // 🏠 页面基础信息
    c.SetData("SiteName", "YYHertz应用")
    c.SetData("PageTitle", c.getPageTitle())
    c.SetData("MetaDescription", c.getMetaDescription())
    c.SetData("MetaKeywords", c.getMetaKeywords())
    
    // 👤 用户信息
    if user := c.getCurrentUser(); user != nil {
        c.SetData("CurrentUser", user)
        c.SetData("IsLoggedIn", true)
        c.SetData("UserPermissions", user.Permissions)
    } else {
        c.SetData("IsLoggedIn", false)
    }
    
    // 🍪 用户偏好
    c.SetData("Theme", c.GetCookie("user_theme"))
    c.SetData("Language", c.GetCookie("language"))
    
    // 🛣️ 导航信息
    c.SetData("CurrentController", c.ControllerName)
    c.SetData("CurrentAction", c.ActionName)
    c.SetData("Breadcrumb", c.buildBreadcrumb())
    
    // 🔒 安全数据
    c.SetData("XSRFToken", c.XSRFToken())
    c.SetData("CSPNonce", c.generateCSPNonce())
    
    // ⏰ 时间信息
    c.SetData("ServerTime", time.Now())
    c.SetData("RequestTime", c.GetData("request_start_time"))
}

// 📄 分页数据设置
func (c *BaseController) setPaginationData(page, pageSize int, total int64) {
    totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
    
    c.SetData("Pagination", map[string]any{
        "CurrentPage":  page,
        "PageSize":     pageSize,
        "Total":        total,
        "TotalPages":   totalPages,
        "HasPrevious":  page > 1,
        "HasNext":      page < int(totalPages),
        "PreviousPage": page - 1,
        "NextPage":     page + 1,
        "Pages":        c.generatePageNumbers(page, int(totalPages)),
    })
}

// 📊 列表数据处理
func (c *BaseController) setListData(items any, itemCount int64, filters map[string]any) {
    c.SetData("Items", items)
    c.SetData("ItemCount", itemCount)
    c.SetData("Filters", filters)
    c.SetData("HasFilters", len(filters) > 0)
    
    // 排序信息
    sortField := c.GetString("sort", "id")
    sortOrder := c.GetString("order", "desc")
    c.SetData("SortField", sortField)
    c.SetData("SortOrder", sortOrder)
    c.SetData("SortURL", c.buildSortURL)
}

// 📱 响应式数据
func (c *BaseController) setResponsiveData() {
    userAgent := c.GetUserAgent()
    
    isMobile := strings.Contains(userAgent, "Mobile")
    isTablet := strings.Contains(userAgent, "Tablet")
    isDesktop := !isMobile && !isTablet
    
    c.SetData("DeviceInfo", map[string]any{
        "IsMobile":  isMobile,
        "IsTablet":  isTablet,
        "IsDesktop": isDesktop,
        "UserAgent": userAgent,
    })
}
```

---

## 🛠️ 工具方法详解

### 工具方法 (6个核心方法)

| 方法 | 功能 | 使用场景 |
|------|------|----------|
| `getControllerMethods(controller)` | 获取控制器所有方法 | 动态路由 |
| `ExtractControllerName(controller)` | 提取控制器名称 | 路由解析 |
| `ExtractActionName(methodName)` | 提取动作名称 | URL构建 |
| `ValidateMethodMapping(mapping, controller)` | 验证方法映射 | 路由验证 |
| `CreateDefaultMethodMapping(controller)` | 创建默认方法映射 | 自动路由 |

### 反射与动态调用

```go
// 🔍 动态方法发现与调用
func (c *BaseController) callDynamicMethod(methodName string, args ...any) ([]reflect.Value, error) {
    // 获取控制器的反射值
    controllerValue := reflect.ValueOf(c.AppController)
    if controllerValue.Kind() == reflect.Ptr {
        controllerValue = controllerValue.Elem()
    }
    
    // 查找方法
    method := controllerValue.MethodByName(methodName)
    if !method.IsValid() {
        return nil, fmt.Errorf("方法 %s 不存在", methodName)
    }
    
    // 准备参数
    methodType := method.Type()
    if len(args) != methodType.NumIn() {
        return nil, fmt.Errorf("参数数量不匹配: 期望 %d, 实际 %d", methodType.NumIn(), len(args))
    }
    
    values := make([]reflect.Value, len(args))
    for i, arg := range args {
        values[i] = reflect.ValueOf(arg)
    }
    
    // 调用方法
    return method.Call(values), nil
}

// 🎯 智能路由处理
func (c *BaseController) handleSmartRouting() {
    // 根据URL模式智能匹配处理方法
    path := c.Ctx.Path()
    method := c.Ctx.Method()
    
    // 提取路径段
    segments := strings.Split(strings.Trim(path, "/"), "/")
    if len(segments) < 2 {
        c.JSONNotFound("路由不存在")
        return
    }
    
    // 构建可能的方法名
    possibleMethods := []string{
        method + strings.Title(segments[len(segments)-1]),     // GetUsers
        method + strings.Title(segments[len(segments)-1])[:len(segments[len(segments)-1])-1], // GetUser (单数形式)
        "Handle" + strings.Title(segments[len(segments)-1]),   // HandleUsers
        strings.Title(segments[len(segments)-1]),               // Users
    }
    
    // 尝试调用方法
    for _, methodName := range possibleMethods {
        if c.methodExists(methodName) {
            c.callMethod(methodName)
            return
        }
    }
    
    c.JSONNotFound("处理方法不存在")
}

// 📝 方法文档生成
func (c *BaseController) generateMethodDocumentation() map[string]any {
    methods := getControllerMethods(c.AppController)
    docs := make(map[string]any)
    
    for _, methodName := range methods {
        methodInfo := c.analyzeMethod(methodName)
        docs[methodName] = map[string]any{
            "name":        methodName,
            "http_method": c.extractHTTPMethod(methodName),
            "action":      ExtractActionName(methodName),
            "parameters":  methodInfo.Parameters,
            "description": methodInfo.Description,
            "example":     methodInfo.Example,
        }
    }
    
    return docs
}

// 🔧 性能优化工具
func (c *BaseController) optimizePerformance() {
    // 启用方法缓存
    c.EnableOptimization()
    
    // 预加载常用数据
    c.preloadCommonData()
    
    // 设置适当的缓存头
    if c.isCacheable() {
        c.SetHeader("Cache-Control", "public, max-age=300") // 5分钟缓存
        c.SetHeader("ETag", c.generateETag())
    }
}
```

---

## 🎯 实际应用场景

### 1. 电商系统完整示例

```go
// 🛍️ 电商产品控制器
type EcommerceProductController struct {
    core.BaseController
    productService  *ProductService
    cartService     *CartService
    reviewService   *ReviewService
    analyticsService *AnalyticsService
}

// 🔄 生命周期管理
func (c *EcommerceProductController) Prepare() {
    // 🔐 用户认证（可选）
    user := c.checkAuthentication()
    if user != nil {
        c.SetData("CurrentUser", user)
        c.SetData("UserCart", c.cartService.GetUserCart(user.ID))
    }
    
    // 📊 访问统计
    c.analyticsService.TrackPageView(c.GetClientIP(), c.Ctx.Path())
    
    // 🎨 设置通用模板数据
    c.setEcommerceData()
}

// 🏪 产品列表页面（支持搜索、分类、排序、分页）
func (c *EcommerceProductController) GetList() {
    // 1️⃣ 获取查询参数
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("per_page", 20)
    category := c.GetString("category")
    keyword := c.GetString("q")
    sortBy := c.GetString("sort", "created_at")
    sortOrder := c.GetString("order", "desc")
    minPrice := c.GetFloat("min_price", 0)
    maxPrice := c.GetFloat("max_price", 0)
    
    // 2️⃣ 构建查询条件
    filters := &ProductFilters{
        Category:  category,
        Keyword:   keyword,
        MinPrice:  minPrice,
        MaxPrice:  maxPrice,
        SortBy:    sortBy,
        SortOrder: sortOrder,
    }
    
    // 3️⃣ 查询产品数据
    products, total, err := c.productService.GetProductList(page, pageSize, filters)
    if err != nil {
        c.LogError("获取产品列表失败", err)
        c.SetData("Error", "获取产品列表失败")
        c.RenderHTML("error/500.html")
        return
    }
    
    // 4️⃣ 获取分类列表（用于筛选）
    categories, _ := c.productService.GetCategoryList()
    
    // 5️⃣ 设置模板数据
    c.SetData("Title", "产品列表")
    c.SetData("Products", products)
    c.SetData("Categories", categories)
    c.SetData("Filters", filters)
    c.setPaginationData(page, pageSize, total)
    
    // 6️⃣ SEO优化
    c.setSEOData("产品列表", "浏览我们的精选产品", "产品,商品,购物")
    
    // 7️⃣ 判断响应格式
    if c.IsAjax() {
        // Ajax请求返回JSON
        c.JSONSuccess("获取成功", map[string]any{
            "products":   products,
            "pagination": c.GetData("Pagination"),
            "total":      total,
        })
    } else {
        // 正常请求渲染HTML
        c.RenderHTML("product/list.html")
    }
}

// 📋 产品详情页面
func (c *EcommerceProductController) GetDetail() {
    id := c.GetParam("id")
    
    // 1️⃣ 获取产品详情
    product, err := c.productService.GetProductDetail(id)
    if err != nil {
        if err == ErrProductNotFound {
            c.CustomAbort(404, c.render404Page("产品不存在"))
        } else {
            c.CustomAbort(500, "获取产品详情失败")
        }
        return
    }
    
    // 2️⃣ 获取相关数据
    reviews, _ := c.reviewService.GetProductReviews(product.ID, 1, 10)
    relatedProducts, _ := c.productService.GetRelatedProducts(product.CategoryID, 4)
    
    // 3️⃣ 记录浏览历史
    if user := c.getCurrentUser(); user != nil {
        c.productService.AddToViewHistory(user.ID, product.ID)
    }
    
    // 4️⃣ 检查用户是否已购买（用于显示评价按钮）
    var hasPurchased bool
    if user := c.getCurrentUser(); user != nil {
        hasPurchased, _ = c.productService.HasUserPurchased(user.ID, product.ID)
    }
    
    // 5️⃣ 设置模板数据
    c.SetData("Title", product.Name)
    c.SetData("Product", product)
    c.SetData("Reviews", reviews)
    c.SetData("RelatedProducts", relatedProducts)
    c.SetData("HasPurchased", hasPurchased)
    
    // 6️⃣ 结构化数据（SEO）
    c.setProductStructuredData(product)
    
    c.RenderHTML("product/detail.html")
}

// 🛒 添加到购物车
func (c *EcommerceProductController) PostAddToCart() {
    // 1️⃣ 用户认证检查
    user := c.getCurrentUser()
    if user == nil {
        c.JSONUnauthorized("请先登录")
        return
    }
    
    // 2️⃣ 获取请求参数
    productID := c.GetForm("product_id")
    quantity := c.GetInt("quantity", 1)
    specifications := c.GetForm("specifications") // JSON格式的规格选择
    
    // 3️⃣ 参数验证
    if productID == "" {
        c.JSONBadRequest("产品ID不能为空")
        return
    }
    
    if quantity < 1 || quantity > 99 {
        c.JSONBadRequest("数量必须在1-99之间")
        return
    }
    
    // 4️⃣ 检查产品是否存在且有库存
    product, err := c.productService.GetProductByID(productID)
    if err != nil {
        c.JSONNotFound("产品不存在")
        return
    }
    
    if product.Stock < quantity {
        c.JSONBadRequest(fmt.Sprintf("库存不足，仅剩%d件", product.Stock))
        return
    }
    
    // 5️⃣ 添加到购物车
    cartItem := &CartItem{
        UserID:        user.ID,
        ProductID:     productID,
        Quantity:      quantity,
        Specifications: specifications,
    }
    
    if err := c.cartService.AddToCart(cartItem); err != nil {
        c.JSONInternalServerError("添加到购物车失败")
        return
    }
    
    // 6️⃣ 返回更新后的购物车信息
    cart, _ := c.cartService.GetUserCart(user.ID)
    
    c.JSONSuccess("添加成功", map[string]any{
        "cart_count": cart.TotalItems,
        "cart_total": cart.TotalAmount,
    })
}

// 🎨 设置电商通用数据
func (c *EcommerceProductController) setEcommerceData() {
    // 🛍️ 购物车信息
    if user := c.getCurrentUser(); user != nil {
        cart, _ := c.cartService.GetUserCart(user.ID)
        c.SetData("Cart", cart)
        c.SetData("CartCount", cart.TotalItems)
    }
    
    // 📂 分类导航
    categories, _ := c.productService.GetTopCategories()
    c.SetData("TopCategories", categories)
    
    // 🔥 热门产品
    hotProducts, _ := c.productService.GetHotProducts(8)
    c.SetData("HotProducts", hotProducts)
    
    // 💱 货币设置
    c.SetData("Currency", "CNY")
    c.SetData("CurrencySymbol", "￥")
}

// 📊 SEO数据设置
func (c *EcommerceProductController) setSEOData(title, description, keywords string) {
    c.SetData("PageTitle", title)
    c.SetData("MetaDescription", description)
    c.SetData("MetaKeywords", keywords)
}

// 🏷️ 产品结构化数据（用于搜索引擎优化）
func (c *EcommerceProductController) setProductStructuredData(product *Product) {
    structuredData := map[string]any{
        "@context": "https://schema.org/",
        "@type":    "Product",
        "name":     product.Name,
        "image":    product.Images,
        "description": product.Description,
        "brand": map[string]any{
            "@type": "Brand",
            "name":  product.Brand,
        },
        "offers": map[string]any{
            "@type":      "Offer",
            "price":      product.Price,
            "priceCurrency": "CNY",
            "availability": "https://schema.org/InStock",
        },
    }
    
    c.SetData("StructuredData", structuredData)
}
```

### 2. 用户管理系统

```go
// 👤 用户管理控制器
type UserManagementController struct {
    core.BaseController
    userService *UserService
    roleService *RoleService
    logService  *LogService
}

// 🔐 权限预检查
func (c *UserManagementController) Prepare() {
    // 检查管理员权限
    if !c.hasAdminPermission() {
        c.CustomAbort(403, "需要管理员权限")
        return
    }
    
    // 启用XSRF保护
    c.EnableXSRF(3600)
    if !c.IsGet() && !c.CheckXSRFCookie() {
        c.JSONForbidden("XSRF令牌验证失败")
        c.StopRun()
        return
    }
    
    c.setAdminCommonData()
}

// 👥 用户列表（支持搜索、筛选、排序）
func (c *UserManagementController) GetUsers() {
    // 查询参数
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("per_page", 50)
    keyword := c.GetString("keyword")
    status := c.GetString("status")
    role := c.GetString("role")
    sortBy := c.GetString("sort", "created_at")
    sortOrder := c.GetString("order", "desc")
    
    // 构建查询条件
    filters := &UserFilters{
        Keyword:   keyword,
        Status:    status,
        Role:      role,
        SortBy:    sortBy,
        SortOrder: sortOrder,
    }
    
    // 查询用户数据
    users, total, err := c.userService.GetUserList(page, pageSize, filters)
    if err != nil {
        c.LogError("获取用户列表失败", err)
        if c.IsAjax() {
            c.JSONInternalServerError("获取用户列表失败")
        } else {
            c.SetData("Error", "获取用户列表失败")
            c.RenderHTML("admin/error.html")
        }
        return
    }
    
    // 获取角色列表（用于筛选）
    roles, _ := c.roleService.GetAllRoles()
    
    if c.IsAjax() {
        // Ajax分页加载
        c.JSONPage("获取成功", users, total)
    } else {
        // 完整页面
        c.SetData("Title", "用户管理")
        c.SetData("Users", users)
        c.SetData("Roles", roles)
        c.SetData("Filters", filters)
        c.setPaginationData(page, pageSize, total)
        c.RenderHTML("admin/user/list.html")
    }
}

// ➕ 创建用户
func (c *UserManagementController) PostCreateUser() {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("请求格式错误")
        return
    }
    
    // 数据验证
    if err := c.ValidateStruct(&req); err != nil {
        c.JSONBadRequest("数据验证失败: " + err.Error())
        return
    }
    
    // 检查邮箱是否已存在
    if exists, _ := c.userService.CheckEmailExists(req.Email); exists {
        c.JSONBadRequest("邮箱已存在")
        return
    }
    
    // 创建用户
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        c.LogError("创建用户失败", err)
        c.JSONInternalServerError("创建用户失败")
        return
    }
    
    // 记录操作日志
    c.logService.LogUserOperation(c.getCurrentUser().ID, "create_user", map[string]any{
        "target_user_id": user.ID,
        "email":          user.Email,
    })
    
    c.JSONSuccess("用户创建成功", user)
}

// ✏️ 更新用户
func (c *UserManagementController) PutUpdateUser() {
    userID := c.GetParam("id")
    
    var req UpdateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("请求格式错误")
        return
    }
    
    // 检查用户是否存在
    existingUser, err := c.userService.GetUserByID(userID)
    if err != nil {
        c.JSONNotFound("用户不存在")
        return
    }
    
    // 权限检查：不能修改比自己权限高的用户
    currentUser := c.getCurrentUser()
    if !c.canModifyUser(currentUser, existingUser) {
        c.JSONForbidden("权限不足")
        return
    }
    
    // 更新用户
    updatedUser, err := c.userService.UpdateUser(userID, &req)
    if err != nil {
        c.LogError("更新用户失败", err)
        c.JSONInternalServerError("更新用户失败")
        return
    }
    
    // 记录操作日志
    c.logService.LogUserOperation(currentUser.ID, "update_user", map[string]any{
        "target_user_id": userID,
        "changes":        c.getUserChanges(existingUser, updatedUser),
    })
    
    c.JSONSuccess("用户更新成功", updatedUser)
}

// 🔒 重置用户密码
func (c *UserManagementController) PostResetPassword() {
    userID := c.GetParam("id")
    newPassword := c.generateRandomPassword()
    
    if err := c.userService.ResetUserPassword(userID, newPassword); err != nil {
        c.JSONInternalServerError("重置密码失败")
        return
    }
    
    // 发送密码重置通知邮件
    user, _ := c.userService.GetUserByID(userID)
    c.sendPasswordResetNotification(user, newPassword)
    
    // 记录操作日志
    c.logService.LogUserOperation(c.getCurrentUser().ID, "reset_password", map[string]any{
        "target_user_id": userID,
    })
    
    c.JSONSuccess("密码重置成功，新密码已发送到用户邮箱", nil)
}

// 📊 用户统计数据
func (c *UserManagementController) GetUserStats() {
    stats, err := c.userService.GetUserStatistics()
    if err != nil {
        c.JSONInternalServerError("获取统计数据失败")
        return
    }
    
    c.JSONSuccess("获取成功", stats)
}
```

### 3. 内容管理系统

```go
// 📄 文章管理控制器
type ArticleController struct {
    core.BaseController
    articleService  *ArticleService
    categoryService *CategoryService
    commentService  *CommentService
    tagService      *TagService
}

// 📝 文章列表（支持多种视图模式）
func (c *ArticleController) GetList() {
    viewMode := c.GetString("view", "grid") // grid, list, card
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("per_page", 12)
    category := c.GetString("category")
    tag := c.GetString("tag")
    status := c.GetString("status", "published")
    
    // 根据用户权限调整状态筛选
    currentUser := c.getCurrentUser()
    if currentUser == nil || !currentUser.IsEditor() {
        status = "published" // 非编辑只能看已发布的文章
    }
    
    filters := &ArticleFilters{
        Category: category,
        Tag:      tag,
        Status:   status,
        ViewMode: viewMode,
    }
    
    articles, total, err := c.articleService.GetArticleList(page, pageSize, filters)
    if err != nil {
        c.handleError("获取文章列表失败", err)
        return
    }
    
    // 获取侧边栏数据
    categories, _ := c.categoryService.GetCategoryTree()
    popularTags, _ := c.tagService.GetPopularTags(20)
    recentArticles, _ := c.articleService.GetRecentArticles(5)
    
    c.SetData("Title", "文章列表")
    c.SetData("Articles", articles)
    c.SetData("Categories", categories)
    c.SetData("PopularTags", popularTags)
    c.SetData("RecentArticles", recentArticles)
    c.SetData("ViewMode", viewMode)
    c.setPaginationData(page, pageSize, total)
    
    // 根据视图模式选择模板
    templateName := fmt.Sprintf("article/list_%s.html", viewMode)
    c.RenderHTML(templateName)
}

// 📖 文章详情（包含评论、相关文章等）
func (c *ArticleController) GetDetail() {
    id := c.GetParam("id")
    
    // 获取文章详情
    article, err := c.articleService.GetArticleDetail(id)
    if err != nil {
        if err == ErrArticleNotFound {
            c.CustomAbort(404, c.render404Page("文章不存在"))
        }
        return
    }
    
    // 权限检查：未发布的文章只有作者和编辑能看
    currentUser := c.getCurrentUser()
    if article.Status != "published" {
        if currentUser == nil || (!currentUser.IsEditor() && currentUser.ID != article.AuthorID) {
            c.CustomAbort(403, "无权访问该文章")
            return
        }
    }
    
    // 增加浏览次数
    c.articleService.IncrementViewCount(article.ID)
    
    // 获取相关数据
    comments, _ := c.commentService.GetArticleComments(article.ID, 1, 20)
    relatedArticles, _ := c.articleService.GetRelatedArticles(article.ID, 5)
    tags, _ := c.tagService.GetArticleTags(article.ID)
    
    // 检查用户是否已收藏
    var isFavorited bool
    if currentUser != nil {
        isFavorited, _ = c.articleService.IsUserFavorited(currentUser.ID, article.ID)
    }
    
    c.SetData("Title", article.Title)
    c.SetData("Article", article)
    c.SetData("Comments", comments)
    c.SetData("RelatedArticles", relatedArticles)
    c.SetData("Tags", tags)
    c.SetData("IsFavorited", isFavorited)
    c.SetData("CanEdit", c.canEditArticle(currentUser, article))
    
    // SEO优化
    c.SetData("MetaDescription", article.Summary)
    c.SetData("MetaKeywords", strings.Join(c.extractTagNames(tags), ","))
    
    // 设置文章结构化数据
    c.setArticleStructuredData(article)
    
    c.RenderHTML("article/detail.html")
}

// ❤️ 收藏文章
func (c *ArticleController) PostToggleFavorite() {
    user := c.getCurrentUser()
    if user == nil {
        c.JSONUnauthorized("请先登录")
        return
    }
    
    articleID := c.GetForm("article_id")
    
    // 检查文章是否存在
    article, err := c.articleService.GetArticleByID(articleID)
    if err != nil {
        c.JSONNotFound("文章不存在")
        return
    }
    
    // 切换收藏状态
    isFavorited, err := c.articleService.ToggleFavorite(user.ID, articleID)
    if err != nil {
        c.JSONInternalServerError("操作失败")
        return
    }
    
    // 返回结果
    action := "已取消收藏"
    if isFavorited {
        action = "收藏成功"
    }
    
    c.JSONSuccess(action, map[string]any{
        "is_favorited": isFavorited,
        "favorite_count": article.FavoriteCount,
    })
}

// 💬 发表评论
func (c *ArticleController) PostComment() {
    user := c.getCurrentUser()
    if user == nil {
        c.JSONUnauthorized("请先登录")
        return
    }
    
    var req CreateCommentRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("请求格式错误")
        return
    }
    
    // 数据验证
    if len(req.Content) < 5 || len(req.Content) > 1000 {
        c.JSONBadRequest("评论内容长度必须在5-1000字符之间")
        return
    }
    
    // 检查是否在冷却期（防止刷评论）
    if c.isCommentCooldown(user.ID) {
        c.JSONBadRequest("评论过于频繁，请稍后再试")
        return
    }
    
    // 内容过滤（敏感词检查）
    if c.containsSensitiveWords(req.Content) {
        c.JSONBadRequest("评论包含敏感内容")
        return
    }
    
    // 创建评论
    comment := &Comment{
        ArticleID: req.ArticleID,
        UserID:    user.ID,
        Content:   req.Content,
        ParentID:  req.ParentID, // 支持回复评论
        Status:    "pending",    // 待审核
    }
    
    if err := c.commentService.CreateComment(comment); err != nil {
        c.JSONInternalServerError("发表评论失败")
        return
    }
    
    c.JSONSuccess("评论已提交，等待审核", comment)
}

// 🔍 文章搜索
func (c *ArticleController) GetSearch() {
    keyword := c.GetString("q")
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("per_page", 10)
    
    if keyword == "" {
        c.JSONBadRequest("搜索关键词不能为空")
        return
    }
    
    // 执行搜索
    results, total, err := c.articleService.SearchArticles(keyword, page, pageSize)
    if err != nil {
        c.JSONInternalServerError("搜索失败")
        return
    }
    
    // 记录搜索日志
    c.logService.LogSearch(keyword, c.GetClientIP(), total)
    
    if c.IsAjax() {
        c.JSONPage("搜索完成", results, total)
    } else {
        c.SetData("Title", fmt.Sprintf("搜索: %s", keyword))
        c.SetData("Keyword", keyword)
        c.SetData("Results", results)
        c.SetData("Total", total)
        c.setPaginationData(page, pageSize, total)
        c.RenderHTML("article/search.html")
    }
}
```

---

## 🚨 常见问题解答 (FAQ)

### Q1: Context为nil错误如何解决？

**问题**：经常遇到"Context is nil"错误

**原因**：控制器未正确初始化或在错误的时机访问Context

**解决方案**：
```go
// ❌ 错误的做法
type MyController struct {
    core.BaseController
    userID string // 在结构体初始化时就尝试获取
}

func NewMyController() *MyController {
    c := &MyController{}
    c.userID = c.GetParam("id") // 此时Context还未初始化
    return c
}

// ✅ 正确的做法
type MyController struct {
    core.BaseController
}

func (c *MyController) Prepare() {
    // 在Prepare或具体的处理方法中访问Context
    userID := c.GetParam("id")
    c.SetData("UserID", userID)
}
```

### Q2: 模板渲染失败怎么办？

**问题**：模板文件找不到或渲染失败

**解决方案**：
```go
func (c *MyController) GetPage() {
    // 1️⃣ 检查模板路径配置
    c.LogInfo("模板路径", map[string]any{
        "ViewPath":   c.ViewPath,
        "LayoutPath": c.LayoutPath,
        "TplName":    c.TplName,
    })
    
    // 2️⃣ 设置必要的模板数据
    if c.Data == nil {
        c.Data = make(map[string]any)
    }
    c.SetData("Title", "页面标题")
    
    // 3️⃣ 使用fallback渲染
    if err := c.RenderHTML("page.html"); err != nil {
        c.LogError("模板渲染失败", err)
        c.String(500, "页面渲染失败")
        return
    }
}
```

### Q3: Session数据丢失问题

**问题**：Session数据无法保存或读取失败

**解决方案**：
```go
func (c *MyController) handleSession() {
    // 1️⃣ 检查Session配置
    sessionID := c.GetSessionID()
    if sessionID == "" {
        c.LogWarn("Session ID为空")
        return
    }
    
    // 2️⃣ 确保数据类型正确
    userID := 123
    c.SetSession("user_id", userID) // 存储时确保类型
    
    // 3️⃣ 读取时进行类型检查
    if val := c.GetSession("user_id"); val != nil {
        if userID, ok := val.(int); ok {
            c.LogInfo("用户ID", userID)
        }
    }
}
```

### Q4: XSRF令牌验证失败

**问题**：XSRF保护导致请求被拒绝

**解决方案**：
```go
// 前端JavaScript
function makeSecureRequest(url, data) {
    // 从meta标签获取XSRF令牌
    const token = document.querySelector('meta[name="xsrf-token"]').content;
    
    return fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-Xsrftoken': token, // 或 'X-CSRF-Token'
        },
        body: JSON.stringify(data)
    });
}

// 控制器端
func (c *MyController) Prepare() {
    // 生成令牌并添加到模板
    token := c.XSRFToken()
    c.SetData("XSRFToken", token)
}

// 模板中添加meta标签
// <meta name="xsrf-token" content="{{.XSRFToken}}">
```

### Q5: 内存泄漏问题

**问题**：长时间运行后内存占用过高

**解决方案**：
```go
func (c *MyController) Prepare() {
    // 启用优化特性
    c.EnableOptimization()
}

func (c *MyController) Finish() {
    // 清理大对象引用
    c.DelData("large_data")
    
    // 调用父类清理方法
    c.BaseController.Finish()
}

// 在路由配置中启用控制器复用
func setupRoutes() {
    // 使用单例模式减少对象创建
    userController := controllers.NewUserController()
    router.GET("/users", userController.GetList)
}
```

### Q6: 文件上传大小限制

**问题**：文件上传失败或大小限制问题

**解决方案**：
```go
func (c *FileController) PostUpload() {
    // 1️⃣ 检查Content-Length
    if c.Ctx.Request.Header.ContentLength() > 10*1024*1024 { // 10MB
        c.JSONBadRequest("文件大小超过限制")
        return
    }
    
    // 2️⃣ 获取文件并检查实际大小
    file, header, err := c.GetFile("upload")
    if err != nil {
        c.JSONBadRequest("文件上传失败: " + err.Error())
        return
    }
    defer file.Close()
    
    if header.Size > 10*1024*1024 {
        c.JSONBadRequest("文件大小不能超过10MB")
        return
    }
    
    // 3️⃣ 检查文件类型
    if !c.isAllowedFileType(header.Filename) {
        c.JSONBadRequest("不支持的文件类型")
        return
    }
    
    // 4️⃣ 安全保存文件
    savePath := c.generateSecureFilename(header.Filename)
    if err := c.SaveFile(file, savePath); err != nil {
        c.JSONInternalServerError("文件保存失败")
        return
    }
    
    c.JSONSuccess("上传成功", map[string]string{
        "filename": header.Filename,
        "path":     savePath,
    })
}
```

### Q7: 数据库连接超时

**问题**：数据库操作超时或连接池耗尽

**解决方案**：
```go
func (c *MyController) GetData() {
    // 1️⃣ 设置超时上下文
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // 2️⃣ 使用带超时的查询
    data, err := c.dataService.GetDataWithContext(ctx, "some_id")
    if err != nil {
        if err == context.DeadlineExceeded {
            c.JSONWithStatus(504, map[string]string{
                "error": "查询超时，请稍后重试",
            })
        } else {
            c.JSONInternalServerError("查询失败")
        }
        return
    }
    
    c.JSONSuccess("获取成功", data)
}
```

### Q8: 并发安全问题

**问题**：多个请求同时访问共享数据

**解决方案**：
```go
type SafeController struct {
    core.BaseController
    mu    sync.RWMutex
    cache map[string]any
}

func (c *SafeController) GetCachedData(key string) any {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    return c.cache[key]
}

func (c *SafeController) SetCachedData(key string, value any) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.cache == nil {
        c.cache = make(map[string]any)
    }
    c.cache[key] = value
}

// 或使用sync.Map
type SafeController2 struct {
    core.BaseController
    cache sync.Map
}

func (c *SafeController2) GetData(key string) any {
    if val, ok := c.cache.Load(key); ok {
        return val
    }
    return nil
}
```

---

## 🏆 最佳实践总结

### 1. 🏗️ 架构设计原则

```go
// ✅ 推荐：分层架构
type UserController struct {
    core.BaseController
    userService    services.UserServiceInterface    // 业务逻辑层
    authService    services.AuthServiceInterface    // 认证服务
    cacheService   services.CacheServiceInterface   // 缓存服务
    logger         logger.LoggerInterface           // 日志服务
}

// ❌ 避免：控制器直接操作数据库
type BadUserController struct {
    core.BaseController
    db *sql.DB // 直接依赖数据库
}
```

### 2. 🔄 生命周期管理

```go
// ✅ 标准生命周期实现
func (c *UserController) Prepare() {
    // 1. 认证检查
    // 2. 权限验证  
    // 3. 通用数据设置
    // 4. 安全检查
}

func (c *UserController) Finish() {
    // 1. 清理敏感数据
    // 2. 记录操作日志
    // 3. 性能统计
    // 4. 资源清理
}
```

### 3. 🔒 安全最佳实践

```go
// ✅ 完整的安全配置
func (c *SecureController) Prepare() {
    // HTTPS强制
    if !c.IsHTTPS() && c.isProduction() {
        c.redirectToHTTPS()
        return
    }
    
    // 安全响应头
    c.setSecurityHeaders()
    
    // XSRF保护
    if c.isStateChangingRequest() {
        c.EnableXSRF(3600)
        if !c.CheckXSRFCookie() {
            c.abortWithError(403, "XSRF验证失败")
            return
        }
    }
    
    // 频率限制
    if c.isRateLimited() {
        c.abortWithError(429, "请求过于频繁")
        return
    }
}
```

### 4. 🎯 错误处理策略

```go
// ✅ 统一错误处理
func (c *BaseController) HandleServiceError(err error) {
    switch e := err.(type) {
    case *ValidationError:
        c.JSONBadRequest(e.Message)
    case *NotFoundError:
        c.JSONNotFound(e.Message)
    case *PermissionError:
        c.JSONForbidden(e.Message)
    case *RateLimitError:
        c.JSONWithStatus(429, map[string]string{"error": e.Message})
    default:
        c.LogError("未处理的错误", err)
        c.JSONInternalServerError("系统错误")
    }
}
```

### 5. ⚡ 性能优化技巧

```go
// ✅ 性能优化实践
func (c *OptimizedController) Prepare() {
    // 启用优化特性
    c.EnableOptimization()
    
    // 预加载常用数据
    c.preloadCommonData()
    
    // 设置缓存策略
    c.setCacheHeaders()
}

func (c *OptimizedController) preloadCommonData() {
    // 使用goroutine并发加载
    var wg sync.WaitGroup
    var mu sync.Mutex
    data := make(map[string]any)
    
    wg.Add(3)
    
    // 并发加载用户信息
    go func() {
        defer wg.Done()
        if user := c.getCurrentUser(); user != nil {
            mu.Lock()
            data["CurrentUser"] = user
            mu.Unlock()
        }
    }()
    
    // 并发加载导航菜单
    go func() {
        defer wg.Done()
        menu := c.menuService.GetUserMenu()
        mu.Lock()
        data["Menu"] = menu
        mu.Unlock()
    }()
    
    // 并发加载通知
    go func() {
        defer wg.Done()
        notifications := c.notificationService.GetRecentNotifications()
        mu.Lock()
        data["Notifications"] = notifications
        mu.Unlock()
    }()
    
    wg.Wait()
    
    // 批量设置数据
    for k, v := range data {
        c.SetData(k, v)
    }
}
```

### 6. 📊 监控和日志

```go
// ✅ 完整的监控实现
func (c *MonitoredController) Prepare() {
    // 记录请求开始时间
    c.SetData("request_start_time", time.Now())
    
    // 设置请求ID
    requestID := c.generateRequestID()
    c.SetData("request_id", requestID)
    c.SetHeader("X-Request-ID", requestID)
}

func (c *MonitoredController) Finish() {
    // 计算响应时间
    startTime := c.GetData("request_start_time").(time.Time)
    duration := time.Since(startTime)
    
    // 记录访问日志
    c.LogInfo("请求完成", map[string]any{
        "method":      c.Ctx.Method(),
        "path":        c.Ctx.Path(),
        "status":      c.Ctx.Response.StatusCode(),
        "duration":    duration,
        "ip":          c.GetClientIP(),
        "user_agent":  c.GetUserAgent(),
        "request_id":  c.GetData("request_id"),
    })
    
    // 性能监控
    if duration > 2*time.Second {
        c.LogWarn("慢请求", map[string]any{
            "path":     c.Ctx.Path(),
            "duration": duration,
        })
    }
    
    c.BaseController.Finish()
}
```

---

## 📋 控制器方法完整清单

### 🔄 生命周期方法 (8个)
- `Init(ct, controllerName, actionName, app)` - 初始化控制器
- `Prepare()` - 预处理（可重写）
- `Finish()` - 后处理（可重写）
- `Destroy()` - 资源清理
- `Reset()` - 重置控制器状态
- `InitWithContext(ctx)` - 使用上下文初始化
- `EnableOptimization()` - 启用优化特性
- `IsOptimizationEnabled()` - 检查优化状态

### 📥 请求处理方法 (15个)
- `GetString(key, def...)` - 获取字符串参数
- `GetInt(key, def...)` - 获取整型参数  
- `GetFloat(key, def...)` - 获取浮点参数
- `GetBool(key, def...)` - 获取布尔参数
- `GetParam(key)` - 获取路径参数
- `GetQuery(key, def...)` - 获取查询参数
- `GetForm(key, def...)` - 获取表单参数
- `GetHeader(key)` - 获取请求头
- `GetUserAgent()` - 获取用户代理
- `GetClientIP()` - 获取客户端IP
- `IsAjax()` - 检查是否Ajax请求
- `IsGet/IsPost/IsPut/IsDelete()` - HTTP方法检查
- `IsPatch/IsHead/IsOptions()` - 其他HTTP方法检查

### 📤 响应输出方法 (16个)
- `JSON(data)` - 标准JSON响应
- `JSONWithStatus(status, data)` - 指定状态JSON响应
- `JSONSuccess(msg, data)` - 成功响应
- `JSONError(msg)` - 错误响应
- `JSONPage(msg, data, count)` - 分页响应
- `JSONOK(data)` - 200状态响应
- `JSONBadRequest(msg)` - 400错误响应
- `JSONUnauthorized(msg)` - 401错误响应
- `JSONForbidden(msg)` - 403错误响应
- `JSONNotFound(msg)` - 404错误响应
- `JSONInternalServerError(msg)` - 500错误响应
- `String(status, format, args...)` - 字符串响应
- `Redirect(url, status...)` - 重定向响应
- `File(filepath)` - 文件响应
- `Data(status, contentType, data)` - 原始数据响应
- `Abort(code)` - 中止响应

### 🎨 模板系统方法 (18个)
- `RenderHTML(viewName, data...)` - 渲染HTML模板
- `RenderWithLayout(view, layout)` - 使用布局渲染
- `RenderTemplate(name, data)` - 渲染指定模板
- `RenderBytes()` - 渲染为字节数组
- `RenderString()` - 渲染为字符串  
- `SetTplName(name)` - 设置模板名称
- `GetTplName()` - 获取模板名称
- `SetLayout(layout)` - 设置布局文件
- `GetLayout()` - 获取布局文件
- `AddTplFunc(name, fn)` - 添加模板函数
- `SetTemplatePath(view, layout)` - 设置模板路径
- `SetTemplateTheme(theme)` - 设置模板主题
- `GetTemplateTheme()` - 获取当前主题
- `AddTemplateFunction(name, fn)` - 添加模板函数
- `CreateTemplateDefinition(name, content)` - 创建模板定义
- `ListAvailableTemplates()` - 列出可用模板
- `RenderHTMLWithIncludes(view, data...)` - 支持include渲染
- `GetTemplateManager()` - 获取模板管理器

### 🍪 Session与Cookie方法 (10个)
- `SetSession(key, value)` - 设置Session值
- `GetSession(key)` - 获取Session值
- `DeleteSession(key)` - 删除Session项
- `HasSession(key)` - 检查Session存在
- `GetSessionID()` - 获取Session ID
- `DestroySession()` - 销毁Session
- `SetCookie(name, value, options...)` - 设置Cookie
- `GetCookie(name)` - 获取Cookie值
- `DeleteCookie(name, path...)` - 删除Cookie
- `HasCookie(name)` - 检查Cookie存在

### 🛣️ 路由管理方法 (14个)
- `AddMethodMapping(http, method)` - 添加方法映射
- `GetMethodMapping()` - 获取方法映射
- `SetMethodMapping(mapping)` - 设置方法映射
- `GetMappedMethod(httpMethod)` - 获取映射方法
- `SetRouteParam(key, value)` - 设置路由参数
- `GetRouteParam(key)` - 获取路由参数
- `GetRouteParams()` - 获取所有路由参数
- `SetRouteParams(params)` - 设置所有路由参数
- `URLFor(endpoint, values...)` - 构建URL
- `BuildURL(controller, action, params...)` - 构建控制器URL
- `AddURLMapping(pattern, method)` - 添加URL映射
- `GetURLMappings()` - 获取URL映射
- `HandlerFunc(name)` - 检查处理器函数
- `URLMapping()` - URL映射配置

### 🔒 安全功能方法 (6个)
- `XSRFToken()` - 生成XSRF令牌
- `CheckXSRFCookie()` - 检查XSRF令牌
- `EnableXSRF(expire...)` - 启用XSRF保护
- `DisableXSRF()` - 禁用XSRF保护
- `SetSecureCookie(secret, name, value, options...)` - 设置安全Cookie
- `GetSecureCookie(secret, name)` - 获取安全Cookie

### 🚦 流程控制方法 (3个)
- `StopRun()` - 停止后续处理
- `Abort(code)` - 中止并返回状态码
- `CustomAbort(status, body)` - 自定义中止响应

### 📊 数据管理方法 (3个)
- `SetData(key, value)` - 设置模板数据
- `GetData(key)` - 获取模板数据
- `DelData(key)` - 删除模板数据

### 🛠️ 中间件管理方法 (3个)
- `GetMiddleware()` - 获取中间件列表
- `SetMiddleware(middlewares)` - 设置中间件列表
- `AddMiddleware(middleware)` - 添加中间件

---

## 📚 扩展阅读

- [YYHertz框架官方文档](../../README.md)
- [MVC架构最佳实践](./mvc-patterns.md)
- [模板引擎使用指南](../template/template-guide.md)
- [安全开发指南](../security/security-best-practices.md)
- [性能优化指南](../performance/optimization-guide.md)

---

**YYHertz团队** - 让Web开发更简单、更高效、更安全！ 🚀