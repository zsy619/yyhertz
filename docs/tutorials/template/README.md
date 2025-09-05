# YYHertz 模板引擎入门教程

<div align="center">

🎨 **掌握YYHertz模板系统** | 从基础语法到高级应用

</div>

---

## 📋 目录

- [模板引擎概述](#模板引擎概述)
- [基础语法入门](#基础语法入门)
- [模板布局系统](#模板布局系统)
- [组件化开发](#组件化开发)
- [模板函数使用](#模板函数使用)
- [高级特性应用](#高级特性应用)
- [性能优化技巧](#性能优化技巧)
- [实战案例](#实战案例)

---

## 🎯 模板引擎概述

### YYHertz模板系统特色

- **🎨 丰富语法**: 150+内置模板函数，覆盖各种使用场景
- **🏗️ 组件化**: 支持组件化开发，模板复用性强
- **⚡ 高性能**: 模板预编译和缓存机制
- **🔄 热重载**: 开发模式下自动重载模板文件
- **🛡️ 安全性**: 内置XSS防护和HTML转义

### 基本文件结构

```
views/
├── layouts/          # 布局模板
│   ├── base.html
│   ├── admin.html
│   └── mobile.html
├── components/       # 组件模板
│   ├── navbar.html
│   ├── footer.html
│   └── sidebar.html
├── pages/           # 页面模板
│   ├── home/
│   ├── user/
│   └── product/
└── partials/        # 片段模板
    ├── forms/
    └── widgets/
```

---

## 📚 基础语法入门

### 1. 变量输出

```html
<!-- 基本变量输出 -->
<h1>{{.title}}</h1>
<p>用户名: {{.user.name}}</p>
<span>当前时间: {{.currentTime}}</span>

<!-- 安全HTML输出 -->
<div>{{.htmlContent | safeHTML}}</div>

<!-- 默认值处理 -->
<p>描述: {{.description | default "暂无描述"}}</p>

<!-- 数字格式化 -->
<span>价格: {{.price | printf "%.2f"}}</span>
```

### 2. 条件判断

```html
<!-- 简单条件 -->
{{if .user}}
    <p>欢迎，{{.user.name}}!</p>
{{else}}
    <p>请先登录</p>
{{end}}

<!-- 多重条件 -->
{{if eq .status "active"}}
    <span class="status-active">激活</span>
{{else if eq .status "pending"}}
    <span class="status-pending">待审核</span>
{{else}}
    <span class="status-inactive">停用</span>
{{end}}

<!-- 复合条件 -->
{{if and .user.isLogin (gt .user.level 1)}}
    <a href="/admin">管理后台</a>
{{end}}

{{if or .user.isAdmin .user.isModerator}}
    <button class="btn-manage">管理操作</button>
{{end}}
```

### 3. 循环遍历

```html
<!-- 数组遍历 -->
<ul>
{{range .users}}
    <li>{{.name}} - {{.email}}</li>
{{else}}
    <li>暂无用户数据</li>
{{end}}
</ul>

<!-- 带索引的遍历 -->
<table>
{{range $index, $user := .users}}
    <tr class="{{if even $index}}even{{else}}odd{{end}}">
        <td>{{add $index 1}}</td>
        <td>{{$user.name}}</td>
        <td>{{$user.email}}</td>
    </tr>
{{end}}
</table>

<!-- Map遍历 -->
<dl>
{{range $key, $value := .settings}}
    <dt>{{$key}}</dt>
    <dd>{{$value}}</dd>
{{end}}
</dl>
```

### 4. 变量定义

```html
<!-- 定义局部变量 -->
{{$userName := .user.name}}
{{$isAdmin := .user.isAdmin}}

<div class="user-info">
    <h3>{{$userName}}</h3>
    {{if $isAdmin}}
        <span class="badge">管理员</span>
    {{end}}
</div>

<!-- 在循环中使用变量 -->
{{range $user := .users}}
    {{$fullName := printf "%s %s" $user.firstName $user.lastName}}
    <p>全名: {{$fullName}}</p>
{{end}}
```

---

## 🏗️ 模板布局系统

### 1. 基础布局模板

```html
<!-- views/layouts/base.html -->
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.title}} - {{.siteName | default "YYHertz应用"}}</title>
    
    <!-- CSS文件 -->
    <link href="/static/css/bootstrap.min.css" rel="stylesheet">
    <link href="/static/css/app.css" rel="stylesheet">
    
    <!-- 页面特定CSS -->
    {{block "extra_css" .}}{{end}}
</head>
<body class="{{.bodyClass | default "default"}}">
    <!-- 导航栏 -->
    {{template "components/navbar.html" .}}
    
    <!-- 面包屑导航 -->
    {{if .breadcrumbs}}
        {{template "components/breadcrumb.html" .}}
    {{end}}
    
    <!-- 主要内容区域 -->
    <main class="main-content">
        <div class="container{{if .fullWidth}}-fluid{{end}}">
            <!-- 闪现消息 -->
            {{template "components/flash-messages.html" .}}
            
            <!-- 页面内容 -->
            {{template "content" .}}
        </div>
    </main>
    
    <!-- 页脚 -->
    {{template "components/footer.html" .}}
    
    <!-- JavaScript文件 -->
    <script src="/static/js/jquery.min.js"></script>
    <script src="/static/js/bootstrap.bundle.min.js"></script>
    <script src="/static/js/app.js"></script>
    
    <!-- 页面特定JS -->
    {{block "extra_js" .}}{{end}}
</body>
</html>
```

### 2. 页面模板

```html
<!-- views/pages/user/profile.html -->
{{define "content"}}
<div class="user-profile">
    <div class="row">
        <div class="col-md-3">
            <!-- 用户头像 -->
            <div class="avatar-section">
                <img src="{{.user.avatar | default "/static/images/default-avatar.png"}}" 
                     alt="{{.user.name}}" 
                     class="avatar-img">
                <h4>{{.user.name}}</h4>
                <p class="text-muted">{{.user.title | default "普通用户"}}</p>
            </div>
        </div>
        
        <div class="col-md-9">
            <!-- 用户信息 -->
            <div class="profile-info">
                <h2>个人资料</h2>
                
                <div class="info-grid">
                    <div class="info-item">
                        <label>邮箱地址</label>
                        <span>{{.user.email}}</span>
                    </div>
                    
                    <div class="info-item">
                        <label>注册时间</label>
                        <span>{{.user.createdAt | dateformat "2006-01-02"}}</span>
                    </div>
                    
                    <div class="info-item">
                        <label>最后登录</label>
                        <span>{{.user.lastLoginAt | timeago}}</span>
                    </div>
                    
                    <div class="info-item">
                        <label>账户状态</label>
                        <span class="status-{{.user.status}}">
                            {{if eq .user.status 1}}激活{{else}}停用{{end}}
                        </span>
                    </div>
                </div>
            </div>
            
            <!-- 最近活动 -->
            {{if .recentActivities}}
            <div class="recent-activities">
                <h3>最近活动</h3>
                <div class="activity-list">
                    {{range .recentActivities}}
                    <div class="activity-item">
                        <i class="icon-{{.type}}"></i>
                        <div class="activity-content">
                            <p>{{.description}}</p>
                            <small class="text-muted">{{.createdAt | timeago}}</small>
                        </div>
                    </div>
                    {{end}}
                </div>
            </div>
            {{end}}
        </div>
    </div>
</div>
{{end}}

<!-- 页面特定CSS -->
{{define "extra_css"}}
<link href="/static/css/profile.css" rel="stylesheet">
{{end}}

<!-- 页面特定JavaScript -->
{{define "extra_js"}}
<script src="/static/js/profile.js"></script>
<script>
$(document).ready(function() {
    // 初始化用户资料页面
    ProfilePage.init({
        userId: {{.user.id}},
        canEdit: {{.canEdit | toJSON}}
    });
});
</script>
{{end}}
```

### 3. 控制器中使用布局

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
)

type UserController struct {
    mvc.BaseController
}

func (c *UserController) GetProfile() {
    userID, _ := c.GetInt("id")
    
    // 获取用户数据
    user, err := c.userService.GetUserByID(userID)
    if err != nil {
        c.ErrorHTML(404, "用户不存在")
        return
    }
    
    // 获取最近活动
    activities, _ := c.activityService.GetUserRecentActivities(userID, 10)
    
    // 准备模板数据
    c.Data["title"] = user.Name + "的个人资料"
    c.Data["user"] = user
    c.Data["recentActivities"] = activities
    c.Data["canEdit"] = c.isCurrentUser(userID)
    c.Data["breadcrumbs"] = []map[string]string{
        {"name": "首页", "url": "/"},
        {"name": "用户", "url": "/users"},
        {"name": user.Name, "url": ""},
    }
    
    // 渲染模板，使用base布局
    c.HTMLWithLayout("pages/user/profile.html", "layouts/base.html", c.Data)
}
```

---

## 🧩 组件化开发

### 1. 导航栏组件

```html
<!-- views/components/navbar.html -->
<nav class="navbar navbar-expand-lg navbar-dark bg-primary">
    <div class="container">
        <a class="navbar-brand" href="/">
            <img src="/static/images/logo.png" alt="{{.siteName}}" height="30">
            {{.siteName | default "YYHertz"}}
        </a>
        
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
            <span class="navbar-toggler-icon"></span>
        </button>
        
        <div class="collapse navbar-collapse" id="navbarNav">
            <!-- 主导航菜单 -->
            <ul class="navbar-nav me-auto">
                {{range .navItems}}
                <li class="nav-item {{if .active}}active{{end}}">
                    <a class="nav-link" href="{{.url}}">
                        {{if .icon}}<i class="{{.icon}}"></i>{{end}}
                        {{.name}}
                    </a>
                </li>
                {{end}}
            </ul>
            
            <!-- 用户菜单 -->
            <ul class="navbar-nav">
                {{if .currentUser}}
                    <!-- 已登录用户菜单 -->
                    <li class="nav-item dropdown">
                        <a class="nav-link dropdown-toggle" href="#" data-bs-toggle="dropdown">
                            <img src="{{.currentUser.avatar | default "/static/images/default-avatar.png"}}" 
                                 alt="{{.currentUser.name}}" 
                                 class="avatar-sm rounded-circle me-1">
                            {{.currentUser.name}}
                        </a>
                        <ul class="dropdown-menu">
                            <li><a class="dropdown-item" href="/profile">个人资料</a></li>
                            <li><a class="dropdown-item" href="/settings">账户设置</a></li>
                            {{if .currentUser.isAdmin}}
                            <li><hr class="dropdown-divider"></li>
                            <li><a class="dropdown-item" href="/admin">管理后台</a></li>
                            {{end}}
                            <li><hr class="dropdown-divider"></li>
                            <li><a class="dropdown-item" href="/logout">退出登录</a></li>
                        </ul>
                    </li>
                {{else}}
                    <!-- 未登录用户菜单 -->
                    <li class="nav-item">
                        <a class="nav-link" href="/login">登录</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" href="/register">注册</a>
                    </li>
                {{end}}
            </ul>
        </div>
    </div>
</nav>
```

### 2. 分页组件

```html
<!-- views/components/pagination.html -->
{{if gt .totalPages 1}}
<nav aria-label="分页导航">
    <ul class="pagination justify-content-center">
        <!-- 上一页 -->
        {{if gt .currentPage 1}}
        <li class="page-item">
            <a class="page-link" href="{{.buildURL (sub .currentPage 1)}}" aria-label="上一页">
                <span aria-hidden="true">&laquo;</span>
            </a>
        </li>
        {{else}}
        <li class="page-item disabled">
            <span class="page-link" aria-label="上一页">
                <span aria-hidden="true">&laquo;</span>
            </span>
        </li>
        {{end}}
        
        <!-- 页码 -->
        {{$currentPage := .currentPage}}
        {{$totalPages := .totalPages}}
        {{$buildURL := .buildURL}}
        
        {{range $page := paginate $currentPage $totalPages}}
            {{if eq $page $currentPage}}
            <li class="page-item active" aria-current="page">
                <span class="page-link">{{$page}}</span>
            </li>
            {{else if eq $page -1}}
            <li class="page-item disabled">
                <span class="page-link">…</span>
            </li>
            {{else}}
            <li class="page-item">
                <a class="page-link" href="{{$buildURL $page}}">{{$page}}</a>
            </li>
            {{end}}
        {{end}}
        
        <!-- 下一页 -->
        {{if lt .currentPage .totalPages}}
        <li class="page-item">
            <a class="page-link" href="{{.buildURL (add .currentPage 1)}}" aria-label="下一页">
                <span aria-hidden="true">&raquo;</span>
            </a>
        </li>
        {{else}}
        <li class="page-item disabled">
            <span class="page-link" aria-label="下一页">
                <span aria-hidden="true">&raquo;</span>
            </span>
        </li>
        {{end}}
    </ul>
</nav>

<!-- 分页信息 -->
<div class="pagination-info text-center text-muted mt-2">
    显示第 {{.startIndex}} - {{.endIndex}} 条，共 {{.total}} 条记录
</div>
{{end}}
```

### 3. 表单组件

```html
<!-- views/components/forms/input.html -->
<div class="mb-3">
    <label for="{{.name}}" class="form-label">
        {{.label}}
        {{if .required}}<span class="text-danger">*</span>{{end}}
    </label>
    
    <input type="{{.type | default "text"}}" 
           class="form-control{{if .error}} is-invalid{{end}}" 
           id="{{.name}}" 
           name="{{.name}}" 
           value="{{.value}}"
           placeholder="{{.placeholder}}"
           {{if .required}}required{{end}}
           {{if .readonly}}readonly{{end}}
           {{if .disabled}}disabled{{end}}>
    
    {{if .error}}
    <div class="invalid-feedback">{{.error}}</div>
    {{end}}
    
    {{if .help}}
    <div class="form-text">{{.help}}</div>
    {{end}}
</div>
```

---

## ⚙️ 模板函数使用

### 1. 常用内置函数

```html
<!-- 字符串处理 -->
<p>{{.text | upper}}</p>                    <!-- 转大写 -->
<p>{{.text | lower}}</p>                    <!-- 转小写 -->
<p>{{.text | title}}</p>                    <!-- 标题格式 -->
<p>{{.text | trim}}</p>                     <!-- 去空格 -->
<p>{{.text | truncate 50}}</p>              <!-- 截取字符 -->
<p>{{.html | stripHTML}}</p>                <!-- 去HTML标签 -->

<!-- 数字处理 -->
<span>{{.price | printf "%.2f"}}</span>     <!-- 格式化数字 -->
<span>{{.count | comma}}</span>             <!-- 千分位格式 -->
<span>{{.percent | percent 2}}</span>       <!-- 百分比格式 -->

<!-- 时间处理 -->
<time>{{.date | dateformat "2006-01-02"}}</time>           <!-- 日期格式化 -->
<span>{{.datetime | dateformat "2006-01-02 15:04:05"}}</span> <!-- 日期时间格式化 -->
<small>{{.createdAt | timeago}}</small>                     <!-- 相对时间 -->

<!-- 数组处理 -->
<p>总共 {{.items | length}} 项</p>          <!-- 获取长度 -->
<p>{{.tags | join ", "}}</p>                <!-- 数组连接 -->
<p>首项: {{.items | first}}</p>             <!-- 获取首项 -->
<p>末项: {{.items | last}}</p>              <!-- 获取末项 -->
```

### 2. 数学运算函数

```html
<!-- 基本运算 -->
<span>{{add .price .tax}}</span>            <!-- 加法 -->
<span>{{sub .total .discount}}</span>       <!-- 减法 -->
<span>{{mul .price .quantity}}</span>       <!-- 乘法 -->
<span>{{div .total .count}}</span>          <!-- 除法 -->

<!-- 比较函数 -->
{{if gt .score 80}}优秀{{end}}              <!-- 大于 -->
{{if lt .age 18}}未成年{{end}}              <!-- 小于 -->
{{if ge .level 5}}高级用户{{end}}           <!-- 大于等于 -->
{{if le .attempts 3}}继续尝试{{end}}        <!-- 小于等于 -->
{{if equal .status "active"}}激活{{end}}    <!-- 等于 -->

<!-- 逻辑函数 -->
{{if and .isLogin .isActive}}已激活用户{{end}}      <!-- 逻辑与 -->
{{if or .isAdmin .isModerator}}管理员{{end}}        <!-- 逻辑或 -->
{{if not .isDeleted}}正常状态{{end}}               <!-- 逻辑非 -->
```

### 3. 自定义模板函数

```go
// 在控制器或初始化代码中注册自定义函数
func init() {
    mvc.AddTemplateFunc("currency", func(amount float64) string {
        return fmt.Sprintf("¥%.2f", amount)
    })
    
    mvc.AddTemplateFunc("avatar", func(email string) string {
        if email == "" {
            return "/static/images/default-avatar.png"
        }
        // 生成Gravatar链接
        hash := md5.Sum([]byte(strings.ToLower(email)))
        return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=identicon", hash)
    })
    
    mvc.AddTemplateFunc("roleClass", func(role string) string {
        switch role {
        case "admin":
            return "badge bg-danger"
        case "moderator":
            return "badge bg-warning"
        case "vip":
            return "badge bg-success"
        default:
            return "badge bg-secondary"
        }
    })
}
```

```html
<!-- 使用自定义函数 -->
<span class="price">{{.product.price | currency}}</span>
<img src="{{.user.email | avatar}}" alt="头像">
<span class="{{.user.role | roleClass}}">{{.user.role}}</span>
```

---

## 🚀 高级特性应用

### 1. 模板缓存配置

```go
// 在应用初始化时配置模板缓存
func initTemplateEngine() {
    mvc.ConfigureTemplate(&mvc.TemplateConfig{
        // 模板根目录
        ViewsPath: "views",
        
        // 模板文件扩展名
        Extension: ".html",
        
        // 是否启用缓存
        EnableCache: true,
        
        // 缓存过期时间
        CacheTimeout: 30 * time.Minute,
        
        // 开发模式（自动重载）
        DevMode: false,
        
        // 自定义分隔符
        LeftDelim:  "{{",
        RightDelim: "}}",
        
        // 模板函数映射
        FuncMap: template.FuncMap{
            "customFunc": customTemplateFunction,
        },
    })
}
```

### 2. 条件加载模板

```html
<!-- 根据用户权限加载不同模板 -->
{{if .user.isAdmin}}
    {{template "admin/dashboard.html" .}}
{{else if .user.isModerator}}
    {{template "moderator/dashboard.html" .}}
{{else}}
    {{template "user/dashboard.html" .}}
{{end}}

<!-- 根据设备类型加载模板 -->
{{if .isMobile}}
    {{template "mobile/layout.html" .}}
{{else}}
    {{template "desktop/layout.html" .}}
{{end}}
```

### 3. 动态模板加载

```go
func (c *BaseController) RenderDynamicTemplate(templateName string, data interface{}) {
    // 根据条件选择模板
    var selectedTemplate string
    
    switch {
    case c.IsMobileRequest():
        selectedTemplate = fmt.Sprintf("mobile/%s", templateName)
    case c.IsTabletRequest():
        selectedTemplate = fmt.Sprintf("tablet/%s", templateName)
    default:
        selectedTemplate = fmt.Sprintf("desktop/%s", templateName)
    }
    
    // 检查模板是否存在
    if !c.TemplateExists(selectedTemplate) {
        selectedTemplate = fmt.Sprintf("default/%s", templateName)
    }
    
    c.HTML(selectedTemplate, data)
}
```

---

## ⚡ 性能优化技巧

### 1. 模板预编译

```go
// 预编译模板提升性能
func precompileTemplates() {
    templates := []string{
        "layouts/base.html",
        "components/navbar.html",
        "components/footer.html",
        "pages/home/index.html",
    }
    
    for _, tmpl := range templates {
        mvc.PrecompileTemplate(tmpl)
    }
}
```

### 2. 模板片段缓存

```html
<!-- 缓存耗时的模板片段 -->
{{cache "sidebar" 3600}}
    {{template "components/sidebar.html" .}}
{{end}}

{{cache "user_stats" 1800}}
    {{template "components/user-stats.html" .}}
{{end}}
```

### 3. 条件渲染优化

```html
<!-- 避免不必要的模板渲染 -->
{{if .showSidebar}}
    {{template "components/sidebar.html" .}}
{{end}}

{{if ne .user.role "guest"}}
    {{template "components/user-menu.html" .}}
{{end}}

<!-- 延迟加载内容 -->
<div class="lazy-content" data-src="/api/content/{{.contentId}}">
    <div class="loading">加载中...</div>
</div>
```

---

## 💼 实战案例

### 1. 用户列表页面

```html
<!-- views/pages/user/list.html -->
{{define "content"}}
<div class="user-list-page">
    <!-- 页面标题 -->
    <div class="d-flex justify-content-between align-items-center mb-4">
        <h1>用户管理</h1>
        <a href="/users/create" class="btn btn-primary">
            <i class="fa fa-plus"></i> 新增用户
        </a>
    </div>
    
    <!-- 搜索筛选 -->
    <div class="card mb-4">
        <div class="card-body">
            <form method="GET" class="row g-3">
                <div class="col-md-3">
                    <input type="text" name="keyword" class="form-control" 
                           placeholder="搜索用户名或邮箱" value="{{.filters.keyword}}">
                </div>
                <div class="col-md-2">
                    <select name="status" class="form-select">
                        <option value="">全部状态</option>
                        <option value="1" {{if eq .filters.status 1}}selected{{end}}>激活</option>
                        <option value="0" {{if eq .filters.status 0}}selected{{end}}>停用</option>
                    </select>
                </div>
                <div class="col-md-2">
                    <select name="role" class="form-select">
                        <option value="">全部角色</option>
                        {{range .roles}}
                        <option value="{{.id}}" {{if eq $.filters.role .id}}selected{{end}}>
                            {{.name}}
                        </option>
                        {{end}}
                    </select>
                </div>
                <div class="col-md-2">
                    <button type="submit" class="btn btn-outline-primary w-100">
                        <i class="fa fa-search"></i> 搜索
                    </button>
                </div>
                <div class="col-md-2">
                    <a href="/users" class="btn btn-outline-secondary w-100">
                        <i class="fa fa-refresh"></i> 重置
                    </a>
                </div>
            </form>
        </div>
    </div>
    
    <!-- 用户列表 -->
    <div class="card">
        <div class="card-body">
            {{if .users}}
            <div class="table-responsive">
                <table class="table table-hover">
                    <thead>
                        <tr>
                            <th width="50">
                                <input type="checkbox" class="form-check-input" id="selectAll">
                            </th>
                            <th>用户信息</th>
                            <th>角色</th>
                            <th>状态</th>
                            <th>注册时间</th>
                            <th>最后登录</th>
                            <th width="120">操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .users}}
                        <tr>
                            <td>
                                <input type="checkbox" class="form-check-input user-checkbox" 
                                       value="{{.id}}">
                            </td>
                            <td>
                                <div class="d-flex align-items-center">
                                    <img src="{{.email | avatar}}" alt="头像" 
                                         class="rounded-circle me-2" width="32" height="32">
                                    <div>
                                        <div class="fw-bold">{{.name}}</div>
                                        <small class="text-muted">{{.email}}</small>
                                    </div>
                                </div>
                            </td>
                            <td>
                                {{range .roles}}
                                <span class="{{.name | roleClass}}">{{.displayName}}</span>
                                {{end}}
                            </td>
                            <td>
                                {{if eq .status 1}}
                                    <span class="badge bg-success">激活</span>
                                {{else}}
                                    <span class="badge bg-secondary">停用</span>
                                {{end}}
                            </td>
                            <td>
                                <small>{{.createdAt | dateformat "2006-01-02"}}</small>
                            </td>
                            <td>
                                {{if .lastLoginAt}}
                                    <small>{{.lastLoginAt | timeago}}</small>
                                {{else}}
                                    <small class="text-muted">从未登录</small>
                                {{end}}
                            </td>
                            <td>
                                <div class="btn-group btn-group-sm">
                                    <a href="/users/{{.id}}" class="btn btn-outline-primary" 
                                       title="查看详情">
                                        <i class="fa fa-eye"></i>
                                    </a>
                                    <a href="/users/{{.id}}/edit" class="btn btn-outline-secondary" 
                                       title="编辑">
                                        <i class="fa fa-edit"></i>
                                    </a>
                                    <button class="btn btn-outline-danger" 
                                            onclick="deleteUser({{.id}})" title="删除">
                                        <i class="fa fa-trash"></i>
                                    </button>
                                </div>
                            </td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            
            <!-- 批量操作 -->
            <div class="batch-actions mt-3" style="display: none;">
                <div class="btn-group">
                    <button class="btn btn-sm btn-success" onclick="batchAction('activate')">
                        <i class="fa fa-check"></i> 批量激活
                    </button>
                    <button class="btn btn-sm btn-warning" onclick="batchAction('deactivate')">
                        <i class="fa fa-pause"></i> 批量停用
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="batchAction('delete')">
                        <i class="fa fa-trash"></i> 批量删除
                    </button>
                </div>
            </div>
            
            <!-- 分页 -->
            {{template "components/pagination.html" .pagination}}
            
            {{else}}
            <div class="text-center py-5">
                <i class="fa fa-users fa-3x text-muted mb-3"></i>
                <h5 class="text-muted">暂无用户数据</h5>
                <p class="text-muted">点击上方"新增用户"按钮创建第一个用户</p>
            </div>
            {{end}}
        </div>
    </div>
</div>
{{end}}

<!-- 页面特定JavaScript -->
{{define "extra_js"}}
<script>
$(document).ready(function() {
    // 全选功能
    $('#selectAll').change(function() {
        $('.user-checkbox').prop('checked', this.checked);
        toggleBatchActions();
    });
    
    // 单选功能
    $('.user-checkbox').change(function() {
        toggleBatchActions();
    });
    
    // 显示/隐藏批量操作按钮
    function toggleBatchActions() {
        const checked = $('.user-checkbox:checked').length;
        $('.batch-actions').toggle(checked > 0);
    }
});

// 删除用户
function deleteUser(userId) {
    if (confirm('确定要删除这个用户吗？')) {
        $.ajax({
            url: '/api/users/' + userId,
            method: 'DELETE',
            success: function(response) {
                if (response.success) {
                    location.reload();
                } else {
                    alert('删除失败：' + response.message);
                }
            },
            error: function() {
                alert('删除失败，请稍后重试');
            }
        });
    }
}

// 批量操作
function batchAction(action) {
    const userIds = $('.user-checkbox:checked').map(function() {
        return $(this).val();
    }).get();
    
    if (userIds.length === 0) {
        alert('请选择要操作的用户');
        return;
    }
    
    let confirmText = '';
    switch (action) {
        case 'activate':
            confirmText = '确定要激活选中的用户吗？';
            break;
        case 'deactivate':
            confirmText = '确定要停用选中的用户吗？';
            break;
        case 'delete':
            confirmText = '确定要删除选中的用户吗？此操作不可恢复！';
            break;
    }
    
    if (confirm(confirmText)) {
        $.ajax({
            url: '/api/users/batch',
            method: 'POST',
            data: {
                action: action,
                user_ids: userIds
            },
            success: function(response) {
                if (response.success) {
                    location.reload();
                } else {
                    alert('操作失败：' + response.message);
                }
            },
            error: function() {
                alert('操作失败，请稍后重试');
            }
        });
    }
}
</script>
{{end}}
```

### 2. 控制器代码

```go
func (c *UserController) GetList() {
    // 获取查询参数
    keyword := c.GetString("keyword")
    status := c.GetInt("status", -1)
    role := c.GetInt("role")
    page := c.GetInt("page", 1)
    pageSize := 20
    
    // 构建查询条件
    filters := &UserFilters{
        Keyword: keyword,
        Role:    role,
    }
    if status >= 0 {
        filters.Status = &status
    }
    
    // 查询用户列表
    users, total, err := c.userService.GetUserList(filters, page, pageSize)
    if err != nil {
        c.ErrorHTML(500, "查询失败")
        return
    }
    
    // 查询所有角色
    roles, _ := c.roleService.GetAllRoles()
    
    // 准备模板数据
    c.Data["title"] = "用户管理"
    c.Data["users"] = users
    c.Data["filters"] = filters
    c.Data["roles"] = roles
    
    // 分页数据
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    c.Data["pagination"] = map[string]interface{}{
        "currentPage": page,
        "totalPages":  totalPages,
        "total":       total,
        "startIndex":  (page-1)*pageSize + 1,
        "endIndex":    int(math.Min(float64(page*pageSize), float64(total))),
        "buildURL": func(p int) string {
            return c.BuildPaginationURL(p, map[string]string{
                "keyword": keyword,
                "status":  fmt.Sprintf("%d", status),
                "role":    fmt.Sprintf("%d", role),
            })
        },
    }
    
    // 渲染页面
    c.HTMLWithLayout("pages/user/list.html", "layouts/base.html", c.Data)
}
```

---

<div align="center">

**🎨 掌握YYHertz模板引擎，让前端开发更高效！**

**组件化思维 + 模板函数 = 强大的视图层解决方案 🚀**

</div>