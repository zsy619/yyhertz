# 🎨 视图模板概览

YYHertz提供了强大而灵活的视图模板系统，支持多种模板引擎、布局管理、组件化开发和动态内容渲染。本文档将详细介绍模板系统的架构、特性和使用方法。

## 🌟 核心特性

### 🚀 多引擎支持
- **HTML模板**: Go原生html/template引擎，安全高效
- **Pug模板**: 简洁的缩进式语法，快速开发
- **Handlebars**: 兼容JavaScript的模板语法
- **自定义引擎**: 支持集成第三方模板引擎

### 🎯 高级功能
- **布局系统**: 嵌套布局，模板继承
- **组件化**: 可重用的UI组件
- **缓存机制**: 智能模板缓存，生产优化
- **热重载**: 开发环境自动刷新
- **国际化**: 多语言支持
- **安全防护**: XSS防护，自动转义

## 🏗️ 模板架构

### 目录结构
```
views/
├── layout/                 # 布局模板
│   ├── layout.html        # 主布局
│   ├── admin.html         # 管理后台布局
│   └── mobile.html        # 移动端布局
├── partials/              # 部分模板/组件
│   ├── header.html        # 页头组件
│   ├── footer.html        # 页脚组件
│   ├── sidebar.html       # 侧边栏组件
│   └── pagination.html    # 分页组件
├── home/                  # 首页模板
│   ├── index.html         
│   └── about.html         
├── user/                  # 用户模块模板
│   ├── profile.html       
│   ├── settings.html      
│   └── list.html          
└── errors/                # 错误页面模板
    ├── 404.html           
    ├── 500.html           
    └── maintenance.html   
```

### 模板层次结构
```
┌─────────────────────────┐
│      Layout 布局         │
│  ┌─────────────────────┐ │
│  │   Partials 组件     │ │
│  │ ┌─────────────────┐ │ │
│  │ │  Content 内容   │ │ │
│  │ │                 │ │ │
│  │ └─────────────────┘ │ │
│  └─────────────────────┘ │
└─────────────────────────┘
```

## 📝 模板语法

### 基本语法

YYHertz使用Go模板语法，提供了丰富的内置函数：

```html
<!DOCTYPE html>
<html lang="{{.Lang | default "zh-CN"}}">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}} - {{.SiteName}}</title>
    <meta name="description" content="{{.Description}}">
    <meta name="keywords" content="{{join .Keywords ","}}">
</head>
<body class="{{.BodyClass}}">
    <!-- 条件渲染 -->
    {{if .User}}
        <div class="user-info">
            欢迎，{{.User.Name}}！
        </div>
    {{else}}
        <div class="login-prompt">
            <a href="/login">请登录</a>
        </div>
    {{end}}
    
    <!-- 循环渲染 -->
    {{range .Posts}}
        <article class="post">
            <h2><a href="/post/{{.ID}}">{{.Title}}</a></h2>
            <div class="meta">
                <time datetime="{{.CreatedAt | date "2006-01-02"}}">
                    {{.CreatedAt | humanize}}
                </time>
                <span class="author">作者: {{.Author.Name}}</span>
            </div>
            <div class="content">
                {{.Content | markdown | safe}}
            </div>
        </article>
    {{end}}
    
    <!-- 模板函数 -->
    {{$currentYear := now | date "2006"}}
    <footer>
        <p>&copy; {{$currentYear}} {{.SiteName}}. All rights reserved.</p>
    </footer>
</body>
</html>
```

### 内置模板函数

YYHertz提供了丰富的内置函数：

```html
<!-- 字符串函数 -->
{{.Text | upper}}           <!-- 转大写 -->
{{.Text | lower}}           <!-- 转小写 -->
{{.Text | title}}           <!-- 首字母大写 -->
{{.Text | truncate 100}}    <!-- 截断文本 -->
{{.HTML | stripTags}}       <!-- 去除HTML标签 -->

<!-- 日期函数 -->
{{.Date | date "2006-01-02 15:04:05"}}  <!-- 格式化日期 -->
{{.Date | humanize}}                     <!-- 人性化时间 -->
{{.Date | timeAgo}}                      <!-- 相对时间 -->

<!-- 数组函数 -->
{{.Items | length}}         <!-- 数组长度 -->
{{.Items | first}}          <!-- 第一个元素 -->
{{.Items | last}}           <!-- 最后一个元素 -->
{{.Items | slice 0 5}}      <!-- 数组切片 -->
{{.Tags | join ", "}}       <!-- 连接数组 -->

<!-- 数学函数 -->
{{add .Price .Tax}}         <!-- 加法 -->
{{sub .Total .Discount}}    <!-- 减法 -->
{{mul .Price .Quantity}}    <!-- 乘法 -->
{{div .Total .Count}}       <!-- 除法 -->

<!-- 条件函数 -->
{{.Status | eq "active"}}   <!-- 等于 -->
{{.Age | gt 18}}            <!-- 大于 -->
{{.Score | lt 60}}          <!-- 小于 -->
{{or .Title .Name}}         <!-- 逻辑或 -->
{{and .IsActive .IsValid}}  <!-- 逻辑与 -->

<!-- URL函数 -->
{{url "/user" .User.ID}}    <!-- 生成URL -->
{{asset "css/style.css"}}   <!-- 静态资源URL -->
{{.Avatar | gravatar 80}}   <!-- Gravatar头像 -->
```

## 🏛️ 布局系统

### 主布局模板

```html
<!-- views/layout/layout.html -->
<!DOCTYPE html>
<html lang="{{.Lang | default "zh-CN"}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}{{.Title}} - {{.SiteName}}{{end}}</title>
    
    <!-- 基础CSS -->
    <link href="{{asset "css/bootstrap.min.css"}}" rel="stylesheet">
    <link href="{{asset "css/app.css"}}" rel="stylesheet">
    
    <!-- 页面特定CSS -->
    {{block "css" .}}{{end}}
    
    <!-- SEO Meta -->
    {{block "meta" .}}
    <meta name="description" content="{{.Description}}">
    <meta name="keywords" content="{{join .Keywords ","}}">
    {{end}}
</head>
<body class="{{block "body-class" .}}{{.BodyClass}}{{end}}">
    <!-- 页头 -->
    {{template "partials/header.html" .}}
    
    <!-- 面包屑 -->
    {{if .Breadcrumbs}}
        {{template "partials/breadcrumb.html" .}}
    {{end}}
    
    <!-- 主要内容 -->
    <main class="main-content">
        {{block "content" .}}
        <div class="container">
            <h1>默认内容</h1>
        </div>
        {{end}}
    </main>
    
    <!-- 页脚 -->
    {{template "partials/footer.html" .}}
    
    <!-- 基础JS -->
    <script src="{{asset "js/jquery.min.js"}}"></script>
    <script src="{{asset "js/bootstrap.min.js"}}"></script>
    <script src="{{asset "js/app.js"}}"></script>
    
    <!-- 页面特定JS -->
    {{block "js" .}}{{end}}
</body>
</html>
```

### 页面模板继承

```html
<!-- views/home/index.html -->
{{define "title"}}首页 - {{.SiteName}}{{end}}

{{define "meta"}}
<meta name="description" content="{{.SiteDescription}}">
<meta property="og:title" content="{{.SiteName}}">
<meta property="og:description" content="{{.SiteDescription}}">
<meta property="og:image" content="{{asset "img/og-image.jpg"}}">
{{end}}

{{define "css"}}
<link href="{{asset "css/home.css"}}" rel="stylesheet">
<style>
.hero-section {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 100px 0;
}
</style>
{{end}}

{{define "body-class"}}home-page{{end}}

{{define "content"}}
<section class="hero-section">
    <div class="container">
        <div class="row">
            <div class="col-lg-8 mx-auto text-center">
                <h1 class="display-4 fw-bold">{{.Hero.Title}}</h1>
                <p class="lead">{{.Hero.Subtitle}}</p>
                <div class="mt-4">
                    <a href="{{.Hero.PrimaryButton.URL}}" class="btn btn-light btn-lg me-3">
                        {{.Hero.PrimaryButton.Text}}
                    </a>
                    <a href="{{.Hero.SecondaryButton.URL}}" class="btn btn-outline-light btn-lg">
                        {{.Hero.SecondaryButton.Text}}
                    </a>
                </div>
            </div>
        </div>
    </div>
</section>

<section class="features py-5">
    <div class="container">
        <div class="row">
            {{range .Features}}
            <div class="col-md-4 mb-4">
                <div class="card h-100">
                    <div class="card-body text-center">
                        <i class="{{.Icon}} fa-3x text-primary mb-3"></i>
                        <h5 class="card-title">{{.Title}}</h5>
                        <p class="card-text">{{.Description}}</p>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
    </div>
</section>
{{end}}

{{define "js"}}
<script>
$(document).ready(function() {
    // 首页特定的JavaScript代码
    $('.hero-section').animate({opacity: 1}, 1000);
    
    // 特性卡片悬停效果
    $('.card').hover(
        function() { $(this).addClass('shadow-lg'); },
        function() { $(this).removeClass('shadow-lg'); }
    );
});
</script>
{{end}}

{{template "layout/layout.html" .}}
```

## 🧩 组件系统

### 可重用组件

```html
<!-- views/partials/card.html -->
<div class="card {{.Class}}">
    {{if .Image}}
    <img src="{{.Image}}" class="card-img-top" alt="{{.Title}}">
    {{end}}
    
    <div class="card-body">
        {{if .Title}}
        <h5 class="card-title">{{.Title}}</h5>
        {{end}}
        
        {{if .Subtitle}}
        <h6 class="card-subtitle mb-2 text-muted">{{.Subtitle}}</h6>
        {{end}}
        
        <p class="card-text">{{.Content}}</p>
        
        {{if .Actions}}
        <div class="card-actions">
            {{range .Actions}}
            <a href="{{.URL}}" class="btn btn-{{.Type | default "primary"}} {{.Class}}">
                {{if .Icon}}<i class="{{.Icon}}"></i>{{end}}
                {{.Text}}
            </a>
            {{end}}
        </div>
        {{end}}
    </div>
    
    {{if .Footer}}
    <div class="card-footer text-muted">
        {{.Footer}}
    </div>
    {{end}}
</div>
```

### 分页组件

```html
<!-- views/partials/pagination.html -->
{{if gt .TotalPages 1}}
<nav aria-label="分页导航">
    <ul class="pagination justify-content-center">
        <!-- 首页 -->
        {{if gt .CurrentPage 1}}
        <li class="page-item">
            <a class="page-link" href="{{url .BaseURL 1}}">
                <i class="fas fa-angle-double-left"></i>
            </a>
        </li>
        {{end}}
        
        <!-- 上一页 -->
        {{if gt .CurrentPage 1}}
        <li class="page-item">
            <a class="page-link" href="{{url .BaseURL (sub .CurrentPage 1)}}">
                <i class="fas fa-angle-left"></i>
            </a>
        </li>
        {{end}}
        
        <!-- 页码 -->
        {{range .Pages}}
        {{if eq . $.CurrentPage}}
        <li class="page-item active">
            <span class="page-link">{{.}}</span>
        </li>
        {{else}}
        <li class="page-item">
            <a class="page-link" href="{{url $.BaseURL .}}">{{.}}</a>
        </li>
        {{end}}
        {{end}}
        
        <!-- 下一页 -->
        {{if lt .CurrentPage .TotalPages}}
        <li class="page-item">
            <a class="page-link" href="{{url .BaseURL (add .CurrentPage 1)}}">
                <i class="fas fa-angle-right"></i>
            </a>
        </li>
        {{end}}
        
        <!-- 末页 -->
        {{if lt .CurrentPage .TotalPages}}
        <li class="page-item">
            <a class="page-link" href="{{url .BaseURL .TotalPages}}">
                <i class="fas fa-angle-double-right"></i>
            </a>
        </li>
        {{end}}
    </ul>
</nav>
{{end}}
```

## 🎛️ 控制器集成

### 模板渲染方法

```go
// controllers/home_controller.go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
)

type HomeController struct {
    mvc.BaseController
}

// 基础模板渲染
func (c *HomeController) GetIndex() {
    // 设置页面数据
    c.SetData("Title", "首页")
    c.SetData("SiteName", "YYHertz官网")
    c.SetData("Description", "高性能Go Web框架")
    
    // 设置Hero区域数据
    c.SetData("Hero", map[string]any{
        "Title": "YYHertz Web框架",
        "Subtitle": "基于CloudWeGo-Hertz的高性能Go框架",
        "PrimaryButton": map[string]string{
            "Text": "开始使用",
            "URL":  "/docs",
        },
        "SecondaryButton": map[string]string{
            "Text": "查看源码",
            "URL":  "https://github.com/zsy619/yyhertz",
        },
    })
    
    // 设置特性列表
    c.SetData("Features", []map[string]any{
        {
            "Icon": "fas fa-rocket",
            "Title": "高性能",
            "Description": "基于CloudWeGo-Hertz，提供极致性能",
        },
        {
            "Icon": "fas fa-code",
            "Title": "易开发",
            "Description": "类似Beego的开发体验，学习成本低",
        },
        {
            "Icon": "fas fa-shield-alt",
            "Title": "生产就绪",
            "Description": "内置安全防护，适合生产环境",
        },
    })
    
    // 渲染模板
    c.RenderHTML("home/index.html")
}

// 带分页的列表页面
func (c *HomeController) GetPosts() {
    page := c.GetQueryInt("page", 1)
    pageSize := 10
    
    // 获取文章列表和总数
    posts, total := c.getPostsList(page, pageSize)
    
    // 计算分页信息
    totalPages := (total + pageSize - 1) / pageSize
    
    // 生成页码列表
    pages := c.generatePageNumbers(page, totalPages)
    
    // 设置模板数据
    c.SetData("Title", "文章列表")
    c.SetData("Posts", posts)
    c.SetData("CurrentPage", page)
    c.SetData("TotalPages", totalPages)
    c.SetData("Total", total)
    c.SetData("Pages", pages)
    c.SetData("BaseURL", "/posts")
    
    c.RenderHTML("home/posts.html")
}

// 自定义模板函数
func (c *HomeController) GetProfile() {
    user := c.getCurrentUser()
    
    // 添加自定义模板函数
    c.AddTemplateFunc("avatar", func(email string, size int) string {
        return fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=%d", 
            md5.Sum([]byte(email)), size)
    })
    
    c.AddTemplateFunc("shortName", func(fullName string) string {
        parts := strings.Split(fullName, " ")
        if len(parts) >= 2 {
            return fmt.Sprintf("%s %s.", parts[0], string(parts[1][0]))
        }
        return fullName
    })
    
    c.SetData("Title", "用户资料")
    c.SetData("User", user)
    c.RenderHTML("user/profile.html")
}
```

### 响应式渲染

```go
// 根据设备类型渲染不同模板
func (c *HomeController) GetResponsive() {
    userAgent := c.GetHeader("User-Agent")
    
    var template string
    if strings.Contains(strings.ToLower(userAgent), "mobile") {
        template = "home/mobile.html"
        c.SetData("IsMobile", true)
    } else {
        template = "home/desktop.html"
        c.SetData("IsMobile", false)
    }
    
    c.SetData("Title", "响应式页面")
    c.RenderHTML(template)
}

// AJAX局部渲染
func (c *HomeController) GetPartial() {
    if c.IsAjax() {
        // 只渲染内容部分
        c.SetData("Posts", c.getLatestPosts(5))
        c.RenderHTML("partials/post-list.html")
    } else {
        // 渲染完整页面
        c.GetPosts()
    }
}
```

## 🔧 高级特性

### 模板缓存

```go
// 配置模板缓存
app.TemplateCache = mvc.TemplateCacheConfig{
    Enabled:    true,
    MaxSize:    1000,
    TTL:        time.Hour,
    Debug:      false, // 生产环境设为false
    Precompile: []string{
        "layout/layout.html",
        "partials/*.html",
    },
}
```

### 国际化支持

```go
// controllers/base_controller.go
func (c *BaseController) setLocale() {
    // 从URL参数或Cookie获取语言设置
    lang := c.GetQuery("lang")
    if lang == "" {
        lang = c.GetCookie("lang", "zh-CN")
    }
    
    // 设置语言
    c.SetData("Lang", lang)
    
    // 加载语言包
    c.SetData("T", c.loadTranslations(lang))
}

// 在模板中使用
// {{T "welcome_message"}}
// {{T "user_count" .UserCount}}
```

### 模板安全

```go
// 自动转义配置
app.TemplateConfig = mvc.TemplateConfig{
    AutoEscape:    true,  // 自动HTML转义
    TrustedSources: []string{
        "admin/*",  // 管理员模板可以使用原始HTML
    },
    CSPNonce:      true,  // 生成CSP nonce
}
```

## 📱 移动端适配

### 响应式布局

```html
<!-- views/layout/responsive.html -->
<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    
    <!-- 响应式CSS -->
    <link href="{{asset "css/responsive.css"}}" rel="stylesheet">
    
    <!-- 移动端特定CSS -->
    {{if .IsMobile}}
    <link href="{{asset "css/mobile.css"}}" rel="stylesheet">
    {{end}}
</head>
<body class="{{if .IsMobile}}mobile{{else}}desktop{{end}}">
    <!-- 移动端导航 -->
    {{if .IsMobile}}
        {{template "partials/mobile-nav.html" .}}
    {{else}}
        {{template "partials/desktop-nav.html" .}}
    {{end}}
    
    <main>
        {{template "content" .}}
    </main>
    
    {{template "partials/footer.html" .}}
</body>
</html>
```

### PWA支持

```html
<!-- 添加PWA支持 -->
<link rel="manifest" href="/manifest.json">
<meta name="theme-color" content="#667eea">

<script>
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js');
}
</script>
```

---

**🎨 视图模板系统为您提供了强大而灵活的前端开发能力，让您能够快速构建美观、高效的Web界面！**