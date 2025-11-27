# 🛠️ 工具方法集合

YYHertz控制器提供了丰富的工具方法，简化日常开发中的常见任务。

## 🔧 核心工具方法

### 实用工具方法 (6个主要类别)

| 类别 | 方法 | 说明 | 示例 |
|------|------|------|------|
| **流程控制** | `StopRun()` | 停止后续处理 | `c.StopRun()` |
| | `Abort(code)` | 终止并返回状态码 | `c.Abort("404")` |
| | `CustomAbort(status, body)` | 自定义中止 | `c.CustomAbort(500, "错误")` |
| **数据管理** | `SetData(key, value)` | 设置模板数据 | `c.SetData("user", user)` |
| | `GetData(key)` | 获取模板数据 | `user := c.GetData("user")` |
| | `DelData(key)` | 删除模板数据 | `c.DelData("temp")` |
| **反射工具** | `GetControllerMethods()` | 获取控制器方法 | `methods := c.GetControllerMethods()` |
| | `ExtractControllerName()` | 提取控制器名 | `name := c.ExtractControllerName()` |
| | `CreateDefaultMethodMapping()` | 创建方法映射 | `mapping := c.CreateDefaultMethodMapping()` |

## ⚡ 性能优化工具

### 缓存辅助方法

```go
func (c *BaseController) withCache(key string, duration time.Duration, fn func() (interface{}, error)) (interface{}, error) {
    // 🔍 尝试从缓存获取
    if cached := c.getFromCache(key); cached != nil {
        c.LogDebug("缓存命中", map[string]any{"key": key})
        return cached, nil
    }
    
    // 📊 缓存未命中，执行函数
    result, err := fn()
    if err != nil {
        return nil, err
    }
    
    // 💾 存入缓存
    c.setToCache(key, result, duration)
    c.LogDebug("数据已缓存", map[string]any{"key": key, "duration": duration})
    
    return result, nil
}

// 使用示例
func (c *UserController) GetUserProfile() {
    userID := c.GetParam("id")
    cacheKey := fmt.Sprintf("user_profile:%s", userID)
    
    result, err := c.withCache(cacheKey, 10*time.Minute, func() (interface{}, error) {
        return c.userService.GetUserProfile(userID)
    })
    
    if err != nil {
        c.JSONError("获取用户资料失败")
        return
    }
    
    c.JSONSuccess("获取成功", result)
}
```

### 批量操作工具

```go
func (c *BaseController) batchProcess(items []interface{}, batchSize int, processor func(batch []interface{}) error) error {
    totalItems := len(items)
    
    for i := 0; i < totalItems; i += batchSize {
        end := i + batchSize
        if end > totalItems {
            end = totalItems
        }
        
        batch := items[i:end]
        
        c.LogInfo("处理批次", map[string]any{
            "batch_start": i,
            "batch_end":   end,
            "batch_size":  len(batch),
            "total":       totalItems,
        })
        
        if err := processor(batch); err != nil {
            return fmt.Errorf("批次处理失败 [%d:%d]: %v", i, end, err)
        }
    }
    
    return nil
}

// 使用示例：批量导入用户
func (c *AdminController) PostBatchImportUsers() {
    var users []User
    if err := c.BindJSON(&users); err != nil {
        c.JSONBadRequest("数据格式错误")
        return
    }
    
    // 转换为interface{}切片
    items := make([]interface{}, len(users))
    for i, user := range users {
        items[i] = user
    }
    
    err := c.batchProcess(items, 100, func(batch []interface{}) error {
        batchUsers := make([]User, len(batch))
        for i, item := range batch {
            batchUsers[i] = item.(User)
        }
        return c.userService.BatchCreateUsers(batchUsers)
    })
    
    if err != nil {
        c.JSONError("批量导入失败: " + err.Error())
        return
    }
    
    c.JSONSuccess("批量导入成功", map[string]any{
        "imported_count": len(users),
    })
}
```

## 📊 数据转换工具

### 类型转换辅助

```go
func (c *BaseController) safeStringToInt(s string, defaultValue int) int {
    if val, err := strconv.Atoi(s); err == nil {
        return val
    }
    return defaultValue
}

func (c *BaseController) safeStringToFloat(s string, defaultValue float64) float64 {
    if val, err := strconv.ParseFloat(s, 64); err == nil {
        return val
    }
    return defaultValue
}

func (c *BaseController) safeStringToBool(s string, defaultValue bool) bool {
    switch strings.ToLower(s) {
    case "true", "1", "yes", "on":
        return true
    case "false", "0", "no", "off":
        return false
    default:
        return defaultValue
    }
}

// 字符串切片转换
func (c *BaseController) stringToSlice(s string, separator string) []string {
    if s == "" {
        return []string{}
    }
    
    parts := strings.Split(s, separator)
    result := make([]string, 0, len(parts))
    
    for _, part := range parts {
        if trimmed := strings.TrimSpace(part); trimmed != "" {
            result = append(result, trimmed)
        }
    }
    
    return result
}

// 使用示例
func (c *ProductController) GetProducts() {
    // 📊 安全获取参数
    page := c.safeStringToInt(c.GetQuery("page"), 1)
    pageSize := c.safeStringToInt(c.GetQuery("page_size"), 20)
    minPrice := c.safeStringToFloat(c.GetQuery("min_price"), 0)
    onlyActive := c.safeStringToBool(c.GetQuery("active"), true)
    
    // 🏷️ 处理标签
    tagsStr := c.GetQuery("tags")
    tags := c.stringToSlice(tagsStr, ",")
    
    query := ProductQuery{
        Page:      page,
        PageSize:  pageSize,
        MinPrice:  minPrice,
        OnlyActive: onlyActive,
        Tags:      tags,
    }
    
    products, total, err := c.productService.GetProducts(query)
    if err != nil {
        c.JSONError("查询失败")
        return
    }
    
    c.JSONPage("查询成功", products, total)
}
```

## 🔍 验证工具集

### 通用验证方法

```go
// 邮箱验证
func (c *BaseController) isValidEmail(email string) bool {
    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    matched, _ := regexp.MatchString(pattern, email)
    return matched
}

// 手机号验证（中国）
func (c *BaseController) isValidPhone(phone string) bool {
    pattern := `^1[3-9]\d{9}$`
    matched, _ := regexp.MatchString(pattern, phone)
    return matched
}

// URL验证
func (c *BaseController) isValidURL(urlStr string) bool {
    _, err := url.Parse(urlStr)
    return err == nil
}

// 身份证验证（中国）
func (c *BaseController) isValidIDCard(idCard string) bool {
    // 简化版身份证验证
    if len(idCard) != 18 {
        return false
    }
    
    pattern := `^\d{17}[\dxX]$`
    matched, _ := regexp.MatchString(pattern, idCard)
    return matched
}

// 密码强度验证
func (c *BaseController) validatePasswordStrength(password string) (bool, string) {
    if len(password) < 8 {
        return false, "密码长度至少8位"
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasDigit := regexp.MustCompile(`\d`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
    
    if !hasUpper {
        return false, "密码必须包含大写字母"
    }
    if !hasLower {
        return false, "密码必须包含小写字母"
    }
    if !hasDigit {
        return false, "密码必须包含数字"
    }
    if !hasSpecial {
        return false, "密码必须包含特殊字符"
    }
    
    return true, ""
}
```

## 🎯 业务工具方法

### 分页计算工具

```go
type PaginationInfo struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    Total      int64 `json:"total"`
    TotalPages int `json:"total_pages"`
    HasPrev    bool `json:"has_prev"`
    HasNext    bool `json:"has_next"`
    PrevPage   int `json:"prev_page"`
    NextPage   int `json:"next_page"`
    Offset     int `json:"offset"`
}

func (c *BaseController) calculatePagination(page, pageSize int, total int64) PaginationInfo {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }
    if pageSize > 100 {
        pageSize = 100 // 限制最大页面大小
    }
    
    totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
    
    return PaginationInfo{
        Page:       page,
        PageSize:   pageSize,
        Total:      total,
        TotalPages: totalPages,
        HasPrev:    page > 1,
        HasNext:    page < totalPages,
        PrevPage:   page - 1,
        NextPage:   page + 1,
        Offset:     (page - 1) * pageSize,
    }
}

// 使用示例
func (c *ArticleController) GetArticleList() {
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    
    // 获取总数
    total, err := c.articleService.GetTotalCount()
    if err != nil {
        c.JSONError("查询失败")
        return
    }
    
    // 计算分页信息
    pagination := c.calculatePagination(page, pageSize, total)
    
    // 获取数据
    articles, err := c.articleService.GetArticles(pagination.Offset, pageSize)
    if err != nil {
        c.JSONError("获取文章失败")
        return
    }
    
    c.JSON(map[string]any{
        "code":       200,
        "message":    "获取成功",
        "data":       articles,
        "pagination": pagination,
    })
}
```

### 文件处理工具

```go
// 文件类型检测
func (c *BaseController) getFileType(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    
    imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
    videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"}
    documentExts := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"}
    
    for _, e := range imageExts {
        if ext == e {
            return "image"
        }
    }
    
    for _, e := range videoExts {
        if ext == e {
            return "video"
        }
    }
    
    for _, e := range documentExts {
        if ext == e {
            return "document"
        }
    }
    
    return "other"
}

// 文件大小格式化
func (c *BaseController) formatFileSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    
    units := []string{"KB", "MB", "GB", "TB", "PB"}
    return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
```

## 🔄 异步处理工具

### 后台任务队列

```go
type Task struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Payload   map[string]interface{} `json:"payload"`
    CreatedAt time.Time              `json:"created_at"`
}

func (c *BaseController) enqueueTask(taskType string, payload map[string]interface{}) string {
    task := Task{
        ID:        c.generateTaskID(),
        Type:      taskType,
        Payload:   payload,
        CreatedAt: time.Now(),
    }
    
    // 添加到队列
    c.taskQueue.Enqueue(task)
    
    c.LogInfo("任务已加入队列", map[string]any{
        "task_id":   task.ID,
        "task_type": taskType,
    })
    
    return task.ID
}

// 使用示例：发送邮件
func (c *UserController) PostSendEmail() {
    emails := c.GetStringSlice("emails")
    subject := c.GetString("subject")
    content := c.GetString("content")
    
    if len(emails) == 0 {
        c.JSONBadRequest("请提供收件人邮箱")
        return
    }
    
    // 🚀 异步发送邮件
    taskID := c.enqueueTask("send_bulk_email", map[string]interface{}{
        "emails":  emails,
        "subject": subject,
        "content": content,
        "sender":  c.GetCurrentUserID(),
    })
    
    c.JSONSuccess("邮件发送任务已提交", map[string]any{
        "task_id": taskID,
        "count":   len(emails),
    })
}
```

## ❓ 常见问题

**Q: 如何自定义工具方法？**
A: 在BaseController中添加方法，或创建工具类进行组合。

**Q: 工具方法的性能如何优化？**
A: 使用缓存、避免重复计算、合理使用并发。

**Q: 如何处理工具方法中的错误？**
A: 使用统一的错误处理机制，记录日志并返回友好提示。