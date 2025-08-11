# 📤 响应输出系统

YYHertz控制器提供了丰富的响应输出方法，支持JSON、HTML、文件下载等多种响应格式。

## 🎯 JSON响应方法

### 核心JSON响应方法 (16个)

| 方法 | 说明 | HTTP状态码 | 示例 |
|------|------|-----------|------|
| `JSON(data)` | 标准JSON响应 | 200 | `c.JSON(user)` |
| `JSONSuccess(msg, data)` | 成功响应 | 200 | `c.JSONSuccess("操作成功", result)` |
| `JSONError(msg, data...)` | 错误响应 | 200 | `c.JSONError("参数错误")` |
| `JSONPage(msg, data, total)` | 分页响应 | 200 | `c.JSONPage("查询成功", users, 100)` |
| `JSONOK(data)` | 200状态响应 | 200 | `c.JSONOK(data)` |
| `JSONBadRequest(msg)` | 400错误 | 400 | `c.JSONBadRequest("请求参数错误")` |
| `JSONUnauthorized(msg)` | 401错误 | 401 | `c.JSONUnauthorized("未授权访问")` |
| `JSONForbidden(msg)` | 403错误 | 403 | `c.JSONForbidden("禁止访问")` |
| `JSONNotFound(msg)` | 404错误 | 404 | `c.JSONNotFound("资源未找到")` |
| `JSONInternalServerError(msg)` | 500错误 | 500 | `c.JSONInternalServerError("服务器错误")` |

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
        RequestID: c.GetRequestID(),
    }
    c.JSON(response)
}
```

## 📄 模板渲染响应

### HTML模板渲染

```go
func (c *PageController) GetUserProfile() {
    userID := c.GetParam("id")
    user, err := c.userService.GetUser(userID)
    if err != nil {
        c.Error(404, "用户不存在")
        return
    }
    
    // 设置模板数据
    c.SetData("User", user)
    c.SetData("Title", "用户资料")
    c.SetData("Description", user.Name + "的个人资料")
    
    // 渲染模板
    c.Render("users/profile.html")
}

// 使用布局模板
func (c *PageController) GetDashboard() {
    c.SetLayout("layouts/admin.html")
    c.SetData("Title", "管理面板")
    c.SetData("ActiveMenu", "dashboard")
    
    // 获取统计数据
    stats := c.getDashboardStats()
    c.SetData("Stats", stats)
    
    c.Render("admin/dashboard.html")
}
```

## 📎 文件响应处理

### 文件下载

```go
func (c *FileController) GetDownload() {
    fileID := c.GetParam("id")
    
    // 查找文件信息
    fileInfo, err := c.fileService.GetFileInfo(fileID)
    if err != nil {
        c.JSONNotFound("文件不存在")
        return
    }
    
    // 权限检查
    if !c.canAccessFile(fileInfo) {
        c.JSONForbidden("无权限访问此文件")
        return
    }
    
    // 设置响应头
    filename := url.QueryEscape(fileInfo.OriginalName)
    c.Ctx.Header("Content-Disposition", "attachment; filename*=UTF-8''"+filename)
    c.Ctx.Header("Content-Type", "application/octet-stream")
    c.Ctx.Header("Content-Length", strconv.FormatInt(fileInfo.Size, 10))
    
    // 流式输出文件内容
    c.Ctx.File(fileInfo.StoragePath)
}

// 图片预览
func (c *FileController) GetPreview() {
    fileID := c.GetParam("id")
    
    fileInfo, err := c.fileService.GetFileInfo(fileID)
    if err != nil {
        c.Error(404, "文件不存在")
        return
    }
    
    // 检查是否为图片
    if !c.isImageFile(fileInfo.Extension) {
        c.Error(400, "不支持预览的文件类型")
        return
    }
    
    // 设置缓存头
    c.Ctx.Header("Cache-Control", "public, max-age=3600")
    c.Ctx.Header("Content-Type", fileInfo.MimeType)
    
    c.Ctx.File(fileInfo.StoragePath)
}
```

### 流式响应

```go
func (c *APIController) GetDataStream() {
    // 设置SSE响应头
    c.Ctx.Header("Content-Type", "text/event-stream")
    c.Ctx.Header("Cache-Control", "no-cache")
    c.Ctx.Header("Connection", "keep-alive")
    c.Ctx.Header("Access-Control-Allow-Origin", "*")
    
    // 获取数据流
    dataStream := c.dataService.GetRealTimeData()
    
    // 流式发送数据
    for data := range dataStream {
        event := fmt.Sprintf("data: %s\n\n", c.toJSON(data))
        c.Ctx.WriteString(event)
        c.Ctx.Writer().Flush()
        
        // 检查连接状态
        if c.Ctx.IsAborted() {
            break
        }
    }
}
```

## 🌐 内容协商

```go
func (c *APIController) GetUserData() {
    userID := c.GetParam("id")
    user, err := c.userService.GetUser(userID)
    if err != nil {
        c.handleError(err)
        return
    }
    
    // 根据Accept头返回不同格式
    accept := c.GetHeader("Accept")
    
    switch {
    case strings.Contains(accept, "application/json"):
        c.JSONSuccess("获取成功", user)
        
    case strings.Contains(accept, "application/xml"):
        c.XMLResponse(user)
        
    case strings.Contains(accept, "text/csv"):
        c.CSVResponse([]User{user})
        
    case strings.Contains(accept, "text/html"):
        c.SetData("User", user)
        c.Render("users/detail.html")
        
    default:
        c.JSONSuccess("获取成功", user)
    }
}

// XML响应
func (c *BaseController) XMLResponse(data interface{}) {
    c.Ctx.Header("Content-Type", "application/xml; charset=utf-8")
    xmlData, _ := xml.Marshal(data)
    c.Ctx.WriteString(xml.Header + string(xmlData))
}

// CSV响应
func (c *BaseController) CSVResponse(data interface{}) {
    c.Ctx.Header("Content-Type", "text/csv; charset=utf-8")
    c.Ctx.Header("Content-Disposition", "attachment; filename=export.csv")
    
    writer := csv.NewWriter(c.Ctx.Writer())
    defer writer.Flush()
    
    // 写入CSV数据
    c.writeCSVData(writer, data)
}
```

## 🚨 错误响应处理

```go
// 统一错误处理
func (c *BaseController) handleError(err error) {
    switch e := err.(type) {
    case *ValidationError:
        c.JSONBadRequest("数据验证失败: " + e.Error())
        
    case *NotFoundError:
        c.JSONNotFound(e.Error())
        
    case *ForbiddenError:
        c.JSONForbidden(e.Error())
        
    case *UnauthorizedError:
        c.JSONUnauthorized(e.Error())
        
    default:
        c.LogError("未处理的错误", map[string]any{
            "error": err.Error(),
            "path":  c.Ctx.Path(),
        })
        c.JSONInternalServerError("服务器内部错误")
    }
}

// 自定义错误页面
func (c *BaseController) ShowErrorPage(code int, message string) {
    c.SetData("ErrorCode", code)
    c.SetData("ErrorMessage", message)
    c.SetData("Title", fmt.Sprintf("错误 %d", code))
    
    // 根据状态码显示不同错误页面
    switch code {
    case 404:
        c.Render("errors/404.html")
    case 500:
        c.Render("errors/500.html")
    default:
        c.Render("errors/general.html")
    }
}
```

## 🎨 响应格式化

```go
// 美化JSON输出
func (c *BaseController) JSONPretty(data interface{}) {
    c.Ctx.Header("Content-Type", "application/json; charset=utf-8")
    
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        c.JSONInternalServerError("JSON序列化失败")
        return
    }
    
    c.Ctx.Write(jsonData)
}

// JSONP响应
func (c *BaseController) JSONP(callback string, data interface{}) {
    if callback == "" {
        callback = "callback"
    }
    
    jsonData, _ := json.Marshal(data)
    response := fmt.Sprintf("%s(%s);", callback, string(jsonData))
    
    c.Ctx.Header("Content-Type", "application/javascript; charset=utf-8")
    c.Ctx.WriteString(response)
}
```

## ❓ 常见问题

**Q: 如何设置响应缓存？**
A: 使用HTTP缓存头，如Cache-Control、ETag、Last-Modified等。

**Q: 大文件响应如何优化？**
A: 使用流式响应、分片传输、断点续传等技术。

**Q: 如何处理跨域请求？**
A: 设置CORS响应头，或使用专门的CORS中间件。

**Q: JSON响应如何处理敏感字段？**
A: 使用结构体标签、自定义序列化方法或响应过滤器。