# 视图渲染

YYHertz 框架的视图渲染系统提供了灵活而强大的模板渲染功能，支持多种数据绑定、条件渲染、循环渲染等特性，让前端开发更加高效。

## 概述

视图渲染是 MVC 架构中的重要组成部分。YYHertz 的视图渲染系统提供：

- 数据绑定和传递
- 条件渲染和循环渲染
- 部分视图和组件化
- 布局继承和嵌套
- 自定义渲染函数
- 模板缓存和性能优化

## 基本视图渲染

### 简单数据渲染

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
)

type UserController struct {
    mvc.BaseController
}

func (c *UserController) GetProfile() {
    user := &User{
        ID:    1,
        Name:  "John Doe",
        Email: "john@example.com",
        Age:   25,
    }
    
    // 设置单个数据
    c.SetData("User", user)
    c.SetData("Title", "用户资料")
    c.SetData("IsLoggedIn", true)
    
    // 渲染模板
    c.RenderHTML("user/profile.html")
}

func (c *UserController) GetDashboard() {
    // 设置多个数据
    data := map[string]interface{}{
        "Title":      "控制面板",
        "UserCount":  1250,
        "PostCount":  3400,
        "ViewCount":  45000,
        "Statistics": map[string]int{
            "today":     120,
            "week":      850,
            "month":     3200,
        },
    }
    
    // 批量设置数据
    c.SetData(data)
    
    c.RenderHTML("dashboard/index.html")
}
```

### 模板中的数据访问

```html
<!-- user/profile.html -->
{{define "user/profile"}}
<div class="user-profile">
    <h1>{{.Title}}</h1>
    
    {{if .IsLoggedIn}}
        <div class="user-info">
            <h2>欢迎，{{.User.Name}}!</h2>
            <p>邮箱: {{.User.Email}}</p>
            <p>年龄: {{.User.Age}}</p>
        </div>
    {{else}}
        <p>请先登录</p>
    {{end}}
</div>
{{end}}
```

## 条件渲染

### 基本条件判断

```html
<!-- 简单条件 -->
{{if .User}}
    <p>用户已登录</p>
{{end}}

{{if not .User}}
    <p>用户未登录</p>
{{end}}

<!-- if-else 结构 -->
{{if eq .User.Role "admin"}}
    <p>管理员权限</p>
{{else if eq .User.Role "moderator"}}
    <p>版主权限</p>
{{else}}
    <p>普通用户</p>
{{end}}

<!-- 复杂条件 -->
{{if and .User (gt .User.Age 18)}}
    <p>成年用户</p>
{{end}}

{{if or (eq .User.Role "admin") (eq .User.Role "moderator")}}
    <p>管理权限</p>
{{end}}
```

### 比较操作符

```html
<!-- 数值比较 -->
{{if gt .Score 90}}
    <span class="badge badge-success">优秀</span>
{{else if gt .Score 80}}
    <span class="badge badge-warning">良好</span>
{{else if gt .Score 60}}
    <span class="badge badge-info">及格</span>
{{else}}
    <span class="badge badge-danger">不及格</span>
{{end}}

<!-- 字符串比较 -->
{{if eq .Status "active"}}
    <span class="status-active">激活</span>
{{else if eq .Status "pending"}}
    <span class="status-pending">待审核</span>
{{else}}
    <span class="status-inactive">未激活</span>
{{end}}

<!-- 存在性检查 -->
{{if .User.Avatar}}
    <img src="{{.User.Avatar}}" alt="头像">
{{else}}
    <img src="/static/images/default-avatar.png" alt="默认头像">
{{end}}
```

## 循环渲染

### 列表渲染

```go
func (c *UserController) GetUserList() {
    users := []User{
        {ID: 1, Name: "Alice", Email: "alice@example.com"},
        {ID: 2, Name: "Bob", Email: "bob@example.com"},
        {ID: 3, Name: "Charlie", Email: "charlie@example.com"},
    }
    
    c.SetData("Users", users)
    c.SetData("Title", "用户列表")
    c.RenderHTML("user/list.html")
}
```

```html
<!-- user/list.html -->
{{define "user/list"}}
<div class="user-list">
    <h1>{{.Title}}</h1>
    
    {{if .Users}}
        <table class="table">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>姓名</th>
                    <th>邮箱</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody>
                {{range $index, $user := .Users}}
                <tr>
                    <td>{{$user.ID}}</td>
                    <td>{{$user.Name}}</td>
                    <td>{{$user.Email}}</td>
                    <td>
                        <a href="/users/{{$user.ID}}" class="btn btn-sm btn-primary">查看</a>
                        <a href="/users/{{$user.ID}}/edit" class="btn btn-sm btn-warning">编辑</a>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    {{else}}
        <p>暂无用户数据</p>
    {{end}}
</div>
{{end}}
```

### 复杂循环

```html
<!-- 带索引的循环 -->
{{range $index, $item := .Items}}
    <div class="item-{{$index}}">
        <span class="index">{{add $index 1}}</span>
        <span class="name">{{$item.Name}}</span>
    </div>
{{end}}

<!-- 嵌套循环 -->
{{range .Categories}}
    <div class="category">
        <h3>{{.Name}}</h3>
        {{if .Products}}
            <ul class="products">
                {{range .Products}}
                    <li>
                        <strong>{{.Name}}</strong> - 
                        <span class="price">${{.Price}}</span>
                    </li>
                {{end}}
            </ul>
        {{else}}
            <p>此分类下暂无商品</p>
        {{end}}
    </div>
{{end}}

<!-- 字典循环 -->
{{range $key, $value := .Settings}}
    <div class="setting">
        <label>{{$key}}:</label>
        <span>{{$value}}</span>
    </div>
{{end}}
```

## 部分视图

### 包含部分模板

```html
<!-- layouts/main.html -->
<!DOCTYPE html>
<html>
<head>
    {{template "common/head" .}}
</head>
<body>
    {{template "common/navbar" .}}
    
    <main class="container">
        {{template "content" .}}
    </main>
    
    {{template "common/footer" .}}
    {{template "common/scripts" .}}
</body>
</html>
```

```html
<!-- common/head.html -->
{{define "common/head"}}
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Title}} - YYHertz</title>
<link href="/static/css/bootstrap.min.css" rel="stylesheet">
<link href="/static/css/app.css" rel="stylesheet">
{{end}}
```

```html
<!-- common/navbar.html -->
{{define "common/navbar"}}
<nav class="navbar navbar-expand-lg navbar-dark bg-dark">
    <div class="container">
        <a class="navbar-brand" href="/">YYHertz</a>
        
        <div class="navbar-nav ms-auto">
            {{if .CurrentUser}}
                <span class="navbar-text">欢迎，{{.CurrentUser.Name}}</span>
                <a class="nav-link" href="/logout">退出</a>
            {{else}}
                <a class="nav-link" href="/login">登录</a>
                <a class="nav-link" href="/register">注册</a>
            {{end}}
        </div>
    </div>
</nav>
{{end}}
```

### 动态包含

```go
func (c *BaseController) SetPartial(name string, template string, data interface{}) {
    c.SetData(name+"_template", template)
    c.SetData(name+"_data", data)
}

func (c *UserController) GetDashboard() {
    // 设置侧边栏
    c.SetPartial("sidebar", "dashboard/sidebar", map[string]interface{}{
        "MenuItems": []string{"概览", "用户", "设置"},
        "ActiveItem": "概览",
    })
    
    // 设置主要内容
    c.SetData("Stats", getStats())
    c.RenderHTML("dashboard/index.html")
}
```

```html
<!-- dashboard/index.html -->
{{define "dashboard/index"}}
<div class="dashboard">
    <div class="row">
        <div class="col-md-3">
            {{if .sidebar_template}}
                {{template .sidebar_template .sidebar_data}}
            {{end}}
        </div>
        <div class="col-md-9">
            <div class="stats">
                {{range .Stats}}
                    <div class="stat-card">
                        <h3>{{.Title}}</h3>
                        <p>{{.Value}}</p>
                    </div>
                {{end}}
            </div>
        </div>
    </div>
</div>
{{end}}
```

## 自定义函数

### 注册自定义函数

```go
package main

import (
    "html/template"
    "strings"
    "time"
    "github.com/zsy619/yyhertz/framework/mvc"
)

func init() {
    // 注册自定义模板函数
    funcMap := template.FuncMap{
        "add":        add,
        "sub":        sub,
        "formatDate": formatDate,
        "truncate":   truncate,
        "upper":      strings.ToUpper,
        "lower":      strings.ToLower,
        "contains":   strings.Contains,
        "join":       strings.Join,
        "default":    defaultValue,
        "safe":       safeHTML,
    }
    
    mvc.RegisterTemplateFuncs(funcMap)
}

func add(a, b int) int {
    return a + b
}

func sub(a, b int) int {
    return a - b
}

func formatDate(t time.Time, layout string) string {
    return t.Format(layout)
}

func truncate(s string, length int) string {
    if len(s) <= length {
        return s
    }
    return s[:length] + "..."
}

func defaultValue(value, defaultVal interface{}) interface{} {
    if value == nil || value == "" {
        return defaultVal
    }
    return value
}

func safeHTML(s string) template.HTML {
    return template.HTML(s)
}
```

### 使用自定义函数

```html
<!-- 使用自定义函数 -->
<div class="article">
    <h2>{{.Title}}</h2>
    <p class="meta">
        发布于 {{formatDate .CreatedAt "2006-01-02 15:04"}}
        • 作者: {{.Author.Name | upper}}
    </p>
    <p class="summary">
        {{.Content | truncate 200}}
    </p>
    <div class="tags">
        {{join .Tags ", "}}
    </div>
    
    <!-- 数学运算 -->
    <p>总页数: {{add .CurrentPage .RemainingPages}}</p>
    
    <!-- 条件默认值 -->
    <img src="{{.Avatar | default "/static/images/default.jpg"}}" alt="头像">
    
    <!-- 安全HTML -->
    <div class="content">
        {{.HTMLContent | safe}}
    </div>
</div>
```

## 错误处理

### 渲染错误页面

```go
func (c *BaseController) RenderError(code int, message string) {
    c.SetData("Code", code)
    c.SetData("Message", message)
    c.SetData("Title", fmt.Sprintf("错误 %d", code))
    
    // 根据错误代码选择模板
    template := "errors/default.html"
    switch code {
    case 404:
        template = "errors/404.html"
    case 500:
        template = "errors/500.html"
    case 403:
        template = "errors/403.html"
    }
    
    c.RenderHTML(template)
}

func (c *UserController) GetUser() {
    userID := c.GetIntParam("id")
    
    user, err := c.userService.GetByID(userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.RenderError(404, "用户不存在")
        } else {
            c.RenderError(500, "服务器内部错误")
        }
        return
    }
    
    c.SetData("User", user)
    c.RenderHTML("user/detail.html")
}
```

### 错误模板

```html
<!-- errors/404.html -->
{{define "errors/404"}}
<div class="error-page">
    <div class="error-content">
        <h1 class="error-code">{{.Code}}</h1>
        <h2 class="error-title">页面未找到</h2>
        <p class="error-message">{{.Message}}</p>
        <div class="error-actions">
            <a href="/" class="btn btn-primary">返回首页</a>
            <button onclick="history.back()" class="btn btn-secondary">返回上页</button>
        </div>
    </div>
</div>
{{end}}
```

## 性能优化

### 模板缓存

```go
package mvc

import (
    "html/template"
    "sync"
)

type TemplateCache struct {
    templates map[string]*template.Template
    mutex     sync.RWMutex
    enabled   bool
}

func NewTemplateCache(enabled bool) *TemplateCache {
    return &TemplateCache{
        templates: make(map[string]*template.Template),
        enabled:   enabled,
    }
}

func (tc *TemplateCache) Get(name string) (*template.Template, bool) {
    if !tc.enabled {
        return nil, false
    }
    
    tc.mutex.RLock()
    defer tc.mutex.RUnlock()
    
    tmpl, exists := tc.templates[name]
    return tmpl, exists
}

func (tc *TemplateCache) Set(name string, tmpl *template.Template) {
    if !tc.enabled {
        return
    }
    
    tc.mutex.Lock()
    defer tc.mutex.Unlock()
    
    tc.templates[name] = tmpl
}

func (tc *TemplateCache) Clear() {
    tc.mutex.Lock()
    defer tc.mutex.Unlock()
    
    tc.templates = make(map[string]*template.Template)
}
```

### 预编译模板

```go
func (app *App) PrecompileTemplates() error {
    templateDir := app.ViewPath
    
    return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !strings.HasSuffix(path, ".html") {
            return nil
        }
        
        relPath, err := filepath.Rel(templateDir, path)
        if err != nil {
            return err
        }
        
        templateName := strings.TrimSuffix(relPath, ".html")
        templateName = strings.ReplaceAll(templateName, "\\", "/")
        
        tmpl, err := template.New(templateName).
            Funcs(app.templateFuncs).
            ParseFiles(path)
        if err != nil {
            return fmt.Errorf("failed to parse template %s: %w", path, err)
        }
        
        app.templateCache.Set(templateName, tmpl)
        
        return nil
    })
}
```

## 最佳实践

### 1. 模板组织

```
views/
├── layouts/
│   ├── main.html
│   ├── admin.html
│   └── auth.html
├── common/
│   ├── head.html
│   ├── navbar.html
│   ├── sidebar.html
│   └── footer.html
├── user/
│   ├── list.html
│   ├── detail.html
│   └── edit.html
├── admin/
│   ├── dashboard.html
│   └── users.html
└── errors/
    ├── 404.html
    ├── 500.html
    └── default.html
```

### 2. 数据传递规范

```go
// 好的做法：使用结构化数据
type PageData struct {
    Title       string
    Description string
    User        *User
    Posts       []Post
    Pagination  *Pagination
    Meta        map[string]interface{}
}

func (c *BlogController) GetPosts() {
    posts, pagination := c.postService.GetPosts(c.GetPageParam())
    
    data := &PageData{
        Title:       "博客列表",
        Description: "最新博客文章",
        User:        c.GetCurrentUser(),
        Posts:       posts,
        Pagination:  pagination,
        Meta: map[string]interface{}{
            "canonical": c.Request.URL.String(),
            "keywords":  "博客,文章,技术",
        },
    }
    
    c.SetData("Data", data)
    c.RenderHTML("blog/list.html")
}
```

### 3. 模板复用

```html
<!-- 定义可复用的组件 -->
{{define "pagination"}}
<nav aria-label="分页导航">
    <ul class="pagination">
        {{if .HasPrev}}
            <li class="page-item">
                <a class="page-link" href="?page={{.PrevPage}}">上一页</a>
            </li>
        {{end}}
        
        {{range .Pages}}
            <li class="page-item {{if .IsCurrent}}active{{end}}">
                <a class="page-link" href="?page={{.Number}}">{{.Number}}</a>
            </li>
        {{end}}
        
        {{if .HasNext}}
            <li class="page-item">
                <a class="page-link" href="?page={{.NextPage}}">下一页</a>
            </li>
        {{end}}
    </ul>
</nav>
{{end}}

<!-- 在其他模板中使用 -->
{{template "pagination" .Data.Pagination}}
```

### 4. 安全考虑

```html
<!-- 避免 XSS 攻击 -->
<div class="user-content">
    <!-- 危险：直接输出用户内容 -->
    {{.UserContent}}
    
    <!-- 安全：转义输出 -->
    {{.UserContent | html}}
    
    <!-- 受信任的HTML（谨慎使用） -->
    {{.TrustedHTML | safe}}
</div>

<!-- 避免 CSRF 攻击 -->
<form method="POST" action="/user/update">
    <input type="hidden" name="_token" value="{{.CSRFToken}}">
    <!-- 表单字段 -->
</form>
```

YYHertz 的视图渲染系统提供了强大而灵活的模板功能，通过合理的组织和使用，可以构建出高效、可维护的前端界面。
