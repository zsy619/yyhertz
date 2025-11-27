# 📥 请求处理完全指南

YYHertz控制器提供了丰富的请求处理方法，支持各种参数获取、数据绑定和请求信息检查功能。

## 📋 基础参数获取方法

### 核心参数获取方法 (15个)

| 方法 | 说明 | 返回类型 | 示例 |
|------|------|----------|------|
| `GetString(key, def...)` | 获取字符串参数 | string | `name := c.GetString("name", "默认值")` |
| `GetInt(key, def...)` | 获取整型参数 | int | `age := c.GetInt("age", 18)` |
| `GetInt64(key, def...)` | 获取长整型参数 | int64 | `id := c.GetInt64("id", 0)` |
| `GetFloat(key, def...)` | 获取浮点参数 | float64 | `price := c.GetFloat("price", 0.0)` |
| `GetBool(key, def...)` | 获取布尔参数 | bool | `active := c.GetBool("active", true)` |
| `GetParam(key)` | 获取路径参数 | string | `id := c.GetParam("id")` |
| `GetQuery(key, def...)` | 获取查询参数 | string | `page := c.GetQuery("page", "1")` |
| `GetForm(key, def...)` | 获取表单参数 | string | `email := c.GetForm("email")` |
| `GetHeader(key)` | 获取请求头 | string | `token := c.GetHeader("Authorization")` |
| `GetFile(key)` | 获取上传文件 | *multipart.FileHeader | `file := c.GetFile("avatar")` |
| `GetFiles(key)` | 获取多个上传文件 | []*multipart.FileHeader | `files := c.GetFiles("documents")` |
| `GetBody()` | 获取原始请求体 | []byte | `body := c.GetBody()` |
| `GetJSON()` | 获取JSON数据 | map[string]any | `data := c.GetJSON()` |
| `GetClientIP()` | 获取客户端IP | string | `ip := c.GetClientIP()` |
| `GetUserAgent()` | 获取用户代理 | string | `ua := c.GetUserAgent()` |

### 参数获取示例

```go
func (c *UserController) GetUserList() {
    // 📝 分页参数
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    
    // 🔍 搜索参数
    keyword := c.GetString("keyword", "")
    category := c.GetString("category", "")
    
    // 📅 时间范围参数
    startDate := c.GetString("start_date")
    endDate := c.GetString("end_date")
    
    // ✅ 布尔参数
    isActive := c.GetBool("active", true)
    includeDeleted := c.GetBool("include_deleted", false)
    
    // 🔢 数值参数
    minAge := c.GetInt("min_age", 0)
    maxAge := c.GetInt("max_age", 150)
    minSalary := c.GetFloat("min_salary", 0.0)
    
    // 构建查询条件
    query := &UserQuery{
        Page:           page,
        PageSize:       pageSize,
        Keyword:        keyword,
        Category:       category,
        StartDate:      startDate,
        EndDate:        endDate,
        IsActive:       isActive,
        IncludeDeleted: includeDeleted,
        MinAge:         minAge,
        MaxAge:         maxAge,
        MinSalary:      minSalary,
    }
    
    // 执行查询
    users, total, err := c.userService.GetUserList(query)
    if err != nil {
        c.JSONError("查询失败: " + err.Error())
        return
    }
    
    c.JSONPage("查询成功", users, total)
}
```

## 🔍 请求信息检查

### HTTP方法判断

```go
func (c *BaseController) HandleDifferentMethods() {
    // 📋 HTTP方法检查
    switch {
    case c.IsGet():
        c.handleGetRequest()
    case c.IsPost():
        c.handlePostRequest()
    case c.IsPut():
        c.handlePutRequest()
    case c.IsDelete():
        c.handleDeleteRequest()
    case c.IsPatch():
        c.handlePatchRequest()
    case c.IsHead():
        c.handleHeadRequest()
    case c.IsOptions():
        c.handleOptionsRequest()
    default:
        c.Error(405, "不支持的HTTP方法")
    }
}

// 🌐 请求类型判断
func (c *BaseController) CheckRequestType() {
    // Ajax请求检查
    if c.IsAjax() {
        c.LogInfo("检测到Ajax请求")
        c.handleAjaxRequest()
        return
    }
    
    // WebSocket升级请求检查
    if c.IsWebSocket() {
        c.LogInfo("WebSocket升级请求")
        c.handleWebSocketUpgrade()
        return
    }
    
    // 普通HTTP请求
    c.handleNormalRequest()
}
```

### 请求头分析

```go
func (c *BaseController) AnalyzeHeaders() {
    // 🔑 认证信息
    authHeader := c.GetHeader("Authorization")
    if strings.HasPrefix(authHeader, "Bearer ") {
        token := authHeader[7:]
        if user := c.validateJWTToken(token); user != nil {
            c.SetData("current_user", user)
        }
    }
    
    // 🌐 内容协商
    accept := c.GetHeader("Accept")
    switch {
    case strings.Contains(accept, "application/json"):
        c.SetData("response_type", "json")
    case strings.Contains(accept, "text/html"):
        c.SetData("response_type", "html")
    case strings.Contains(accept, "application/xml"):
        c.SetData("response_type", "xml")
    }
    
    // 🌍 语言偏好
    acceptLang := c.GetHeader("Accept-Language")
    if lang := c.parseLanguage(acceptLang); lang != "" {
        c.SetData("language", lang)
    }
    
    // 📱 客户端信息
    userAgent := c.GetUserAgent()
    if device := c.parseDeviceInfo(userAgent); device != nil {
        c.SetData("device_info", device)
    }
    
    // 🔗 来源检查
    referer := c.GetHeader("Referer")
    origin := c.GetHeader("Origin")
    c.LogInfo("请求来源", map[string]any{
        "referer": referer,
        "origin":  origin,
    })
}
```

## 🔗 数据绑定与验证

### JSON数据绑定

```go
// 📋 用户注册请求结构
type UserRegisterRequest struct {
    Name            string `json:"name" validate:"required,min=2,max=50"`
    Email           string `json:"email" validate:"required,email"`
    Password        string `json:"password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
    Age             int    `json:"age" validate:"required,min=18,max=120"`
    Gender          string `json:"gender" validate:"oneof=male female other"`
    Phone           string `json:"phone" validate:"required,phone"`
    Address         string `json:"address" validate:"max=200"`
}

func (c *UserController) PostRegister() {
    // 🔗 绑定JSON数据
    var req UserRegisterRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("JSON格式错误: " + err.Error())
        return
    }
    
    // ✅ 数据验证
    if err := c.ValidateStruct(&req); err != nil {
        c.JSONBadRequest("数据验证失败: " + err.Error())
        return
    }
    
    // 🔐 额外业务验证
    if exists := c.userService.EmailExists(req.Email); exists {
        c.JSONBadRequest("邮箱已被使用")
        return
    }
    
    // 💾 创建用户
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        c.JSONInternalServerError("创建用户失败: " + err.Error())
        return
    }
    
    c.JSONSuccess("注册成功", user)
}
```

### 表单数据绑定

```go
// 📋 表单绑定结构
type ProfileUpdateForm struct {
    Name      string `form:"name" validate:"required,min=2,max=50"`
    Bio       string `form:"bio" validate:"max=500"`
    Website   string `form:"website" validate:"omitempty,url"`
    Location  string `form:"location" validate:"max=100"`
    BirthDate string `form:"birth_date" validate:"omitempty,datetime=2006-01-02"`
}

func (c *UserController) PostUpdateProfile() {
    // 🔗 绑定表单数据
    var form ProfileUpdateForm
    if err := c.BindForm(&form); err != nil {
        c.JSONBadRequest("表单数据错误: " + err.Error())
        return
    }
    
    // ✅ 验证数据
    if err := c.ValidateStruct(&form); err != nil {
        c.JSONBadRequest("数据验证失败: " + err.Error())
        return
    }
    
    // 💾 更新资料
    userID := c.GetCurrentUserID()
    if err := c.userService.UpdateProfile(userID, &form); err != nil {
        c.JSONInternalServerError("更新失败: " + err.Error())
        return
    }
    
    c.JSONSuccess("更新成功", nil)
}
```

### 查询参数绑定

```go
// 📋 查询参数结构
type UserQueryParams struct {
    Page       int    `query:"page" validate:"min=1"`
    PageSize   int    `query:"page_size" validate:"min=1,max=100"`
    Keyword    string `query:"keyword" validate:"max=100"`
    SortBy     string `query:"sort_by" validate:"oneof=name email created_at"`
    SortOrder  string `query:"sort_order" validate:"oneof=asc desc"`
    Status     string `query:"status" validate:"omitempty,oneof=active inactive"`
    CreatedAt  string `query:"created_at" validate:"omitempty,datetime=2006-01-02"`
}

func (c *UserController) GetUsers() {
    // 🔗 绑定查询参数
    var params UserQueryParams
    if err := c.BindQuery(&params); err != nil {
        c.JSONBadRequest("查询参数错误: " + err.Error())
        return
    }
    
    // ✅ 设置默认值
    if params.Page <= 0 {
        params.Page = 1
    }
    if params.PageSize <= 0 {
        params.PageSize = 20
    }
    if params.SortBy == "" {
        params.SortBy = "created_at"
    }
    if params.SortOrder == "" {
        params.SortOrder = "desc"
    }
    
    // 📊 执行查询
    users, total, err := c.userService.QueryUsers(&params)
    if err != nil {
        c.JSONInternalServerError("查询失败: " + err.Error())
        return
    }
    
    c.JSONPage("查询成功", users, total)
}
```

## 📎 文件上传处理

### 单文件上传

```go
func (c *FileController) PostUploadSingle() {
    // 📎 获取上传文件
    file := c.GetFile("file")
    if file == nil {
        c.JSONBadRequest("请选择要上传的文件")
        return
    }
    
    // 📏 检查文件大小 (5MB限制)
    maxSize := int64(5 * 1024 * 1024)
    if file.Size > maxSize {
        c.JSONBadRequest("文件大小不能超过5MB")
        return
    }
    
    // 📋 检查文件类型
    allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"}
    ext := strings.ToLower(filepath.Ext(file.Filename))
    if !c.isAllowedFileType(ext, allowedTypes) {
        c.JSONBadRequest("不支持的文件类型")
        return
    }
    
    // 💾 保存文件
    savedPath, err := c.fileService.SaveUploadedFile(file)
    if err != nil {
        c.JSONInternalServerError("文件保存失败: " + err.Error())
        return
    }
    
    // 📊 记录文件信息
    fileInfo := map[string]any{
        "filename":     file.Filename,
        "size":         file.Size,
        "content_type": file.Header.Get("Content-Type"),
        "saved_path":   savedPath,
        "uploaded_at":  time.Now(),
    }
    
    c.JSONSuccess("文件上传成功", fileInfo)
}
```

### 多文件上传

```go
func (c *FileController) PostUploadMultiple() {
    // 📎 获取多个文件
    files := c.GetFiles("files")
    if len(files) == 0 {
        c.JSONBadRequest("请选择要上传的文件")
        return
    }
    
    // 📊 检查文件数量限制
    maxFiles := 10
    if len(files) > maxFiles {
        c.JSONBadRequest(fmt.Sprintf("最多只能上传%d个文件", maxFiles))
        return
    }
    
    var uploadedFiles []map[string]any
    var errors []string
    
    for i, file := range files {
        // 📏 检查单个文件
        if file.Size > int64(5*1024*1024) {
            errors = append(errors, fmt.Sprintf("文件%d: 大小超过限制", i+1))
            continue
        }
        
        // 💾 保存文件
        savedPath, err := c.fileService.SaveUploadedFile(file)
        if err != nil {
            errors = append(errors, fmt.Sprintf("文件%d: 保存失败", i+1))
            continue
        }
        
        uploadedFiles = append(uploadedFiles, map[string]any{
            "filename":   file.Filename,
            "size":       file.Size,
            "saved_path": savedPath,
        })
    }
    
    // 📊 返回结果
    result := map[string]any{
        "uploaded_count": len(uploadedFiles),
        "uploaded_files": uploadedFiles,
        "errors":        errors,
    }
    
    if len(errors) > 0 {
        c.JSONError("部分文件上传失败", result)
    } else {
        c.JSONSuccess("所有文件上传成功", result)
    }
}
```

## 🔧 高级请求处理

### 流式数据处理

```go
func (c *APIController) PostStreamData() {
    // 📡 处理流式JSON数据
    decoder := json.NewDecoder(c.Ctx.Request.Body)
    defer c.Ctx.Request.Body.Close()
    
    var processedCount int
    var errors []string
    
    for decoder.More() {
        var item map[string]any
        if err := decoder.Decode(&item); err != nil {
            errors = append(errors, fmt.Sprintf("解析第%d项失败: %v", processedCount+1, err))
            continue
        }
        
        // 处理单项数据
        if err := c.processItem(item); err != nil {
            errors = append(errors, fmt.Sprintf("处理第%d项失败: %v", processedCount+1, err))
            continue
        }
        
        processedCount++
        
        // 限制处理数量
        if processedCount >= 1000 {
            break
        }
    }
    
    result := map[string]any{
        "processed_count": processedCount,
        "error_count":    len(errors),
        "errors":         errors,
    }
    
    c.JSONSuccess("数据处理完成", result)
}
```

### 请求重试机制

```go
func (c *BaseController) WithRetry(operation func() error, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        if err := operation(); err == nil {
            return nil
        } else {
            lastErr = err
            c.LogWarn(fmt.Sprintf("操作失败，第%d次重试", i+1), map[string]any{
                "error": err.Error(),
                "retry": i + 1,
            })
            
            // 指数退避
            time.Sleep(time.Duration(1<<uint(i)) * time.Second)
        }
    }
    
    return fmt.Errorf("操作失败，已重试%d次: %v", maxRetries, lastErr)
}
```

## ❓ 常见问题与解答

**Q: 如何处理大文件上传？**
A: 使用流式处理，分块上传，设置合理的超时时间和内存限制。

**Q: 参数验证失败如何返回详细错误信息？**
A: 使用结构化验证，返回字段级别的错误信息，便于前端处理。

**Q: 如何防止文件上传攻击？**
A: 检查文件类型、大小、文件名，使用病毒扫描，存储在安全位置。

**Q: 请求超时如何处理？**
A: 设置合理的超时时间，使用上下文取消机制，及时释放资源。