# 🎨 模板系统详解

YYHertz控制器集成了强大的模板引擎，支持模板渲染、布局管理和自定义模板函数。

## 🏗️ 模板系统架构

### 核心模板方法 (18个)

| 方法 | 说明 | 示例 |
|------|------|------|
| `Render(template)` | 渲染指定模板 | `c.Render("user/profile.html")` |
| `RenderBytes(template)` | 渲染为字节数组 | `data := c.RenderBytes("email.html")` |
| `RenderString(template)` | 渲染为字符串 | `html := c.RenderString("widget.html")` |
| `SetLayout(layout)` | 设置布局模板 | `c.SetLayout("layouts/main.html")` |
| `SetViewPath(path)` | 设置视图路径 | `c.SetViewPath("custom/views")` |
| `AddTemplateFunc(name, fn)` | 添加模板函数 | `c.AddTemplateFunc("upper", strings.ToUpper)` |

## 📁 模板结构规划

```
views/
├── layouts/           # 布局模板
│   ├── main.html     # 主布局
│   ├── admin.html    # 管理后台布局
│   └── api.html      # API文档布局
├── partials/         # 部分模板
│   ├── header.html   # 页头
│   ├── footer.html   # 页脚
│   └── sidebar.html  # 侧边栏
├── users/            # 用户相关模板
│   ├── list.html     # 用户列表
│   ├── profile.html  # 用户资料
│   └── edit.html     # 编辑用户
└── errors/           # 错误页面
    ├── 404.html      # 404错误
    └── 500.html      # 500错误
```

## 🖼️ 基础模板使用

### 布局模板示例

```html
<!-- layouts/main.html -->
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - YYHertz应用</title>
    <link href="/static/css/app.css" rel="stylesheet">
    {{if .Description}}
    <meta name="description" content="{{.Description}}">
    {{end}}
    {{block "css" .}}{{end}}
</head>
<body>
    <!-- 页头 -->
    {{template "partials/header.html" .}}
    
    <!-- 主要内容 -->
    <main class="main-content">
        {{block "content" .}}{{end}}
    </main>
    
    <!-- 页脚 -->
    {{template "partials/footer.html" .}}
    
    <script src="/static/js/app.js"></script>
    {{block "js" .}}{{end}}
</body>
</html>
```

### 页面模板示例

```html
<!-- users/profile.html -->
{{define "content"}}
<div class="user-profile">
    <div class="profile-header">
        <img src="{{.User.Avatar}}" alt="{{.User.Name}}" class="avatar">
        <h1>{{.User.Name}}</h1>
        <p class="bio">{{.User.Bio}}</p>
    </div>
    
    <div class="profile-details">
        <div class="detail-item">
            <label>邮箱:</label>
            <span>{{.User.Email}}</span>
        </div>
        <div class="detail-item">
            <label>注册时间:</label>
            <span>{{formatDate .User.CreatedAt "2006-01-02 15:04:05"}}</span>
        </div>
        {{if .User.Website}}
        <div class="detail-item">
            <label>网站:</label>
            <a href="{{.User.Website}}" target="_blank">{{.User.Website}}</a>
        </div>
        {{end}}
    </div>
</div>
{{end}}

{{define "css"}}
<link href="/static/css/user-profile.css" rel="stylesheet">
{{end}}

{{define "js"}}
<script>
    console.log('用户资料页面已加载');
</script>
{{end}}
```

## 🔧 控制器模板操作

### 基础模板渲染

```go
func (c *UserController) GetProfile() {
    userID := c.GetParam("id")
    user, err := c.userService.GetUser(userID)
    if err != nil {
        c.Error(404, "用户不存在")
        return
    }
    
    // 设置页面数据
    c.SetData("Title", user.Name + "的资料")
    c.SetData("Description", user.Bio)
    c.SetData("User", user)
    c.SetData("IsOwner", c.isProfileOwner(user))
    
    // 设置布局
    c.SetLayout("layouts/main.html")
    
    // 渲染模板
    c.Render("users/profile.html")
}
```

### 动态模板选择

```go
func (c *PageController) GetContent() {
    contentType := c.GetParam("type")
    
    // 根据内容类型选择不同模板
    var template string
    var layout string
    
    switch contentType {
    case "article":
        template = "content/article.html"
        layout = "layouts/blog.html"
    case "product":
        template = "content/product.html"
        layout = "layouts/shop.html"
    case "news":
        template = "content/news.html"
        layout = "layouts/news.html"
    default:
        template = "content/default.html"
        layout = "layouts/main.html"
    }
    
    c.SetLayout(layout)
    c.SetData("ContentType", contentType)
    c.Render(template)
}
```

## 🛠️ 自定义模板函数

### 注册模板函数

```go
func (c *BaseController) initTemplateFunc() {
    // 📅 时间格式化
    c.AddTemplateFunc("formatDate", func(t time.Time, layout string) string {
        return t.Format(layout)
    })
    
    // 📅 相对时间
    c.AddTemplateFunc("timeAgo", func(t time.Time) string {
        duration := time.Since(t)
        switch {
        case duration < time.Minute:
            return "刚刚"
        case duration < time.Hour:
            return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
        case duration < 24*time.Hour:
            return fmt.Sprintf("%d小时前", int(duration.Hours()))
        case duration < 30*24*time.Hour:
            return fmt.Sprintf("%d天前", int(duration.Hours()/24))
        default:
            return t.Format("2006-01-02")
        }
    })
    
    // 🔢 数字格式化
    c.AddTemplateFunc("formatNumber", func(num int64) string {
        return humanize.Comma(num)
    })
    
    // 📏 文件大小格式化
    c.AddTemplateFunc("formatSize", func(size int64) string {
        return humanize.Bytes(uint64(size))
    })
    
    // ✂️ 字符串截取
    c.AddTemplateFunc("truncate", func(s string, length int) string {
        if len(s) <= length {
            return s
        }
        return s[:length] + "..."
    })
    
    // 🔐 HTML转义
    c.AddTemplateFunc("safeHTML", func(s string) template.HTML {
        return template.HTML(s)
    })
    
    // 🔗 URL生成
    c.AddTemplateFunc("url", func(route string, params ...interface{}) string {
        return c.URLFor(route, params...)
    })
    
    // 🎨 CSS类名生成
    c.AddTemplateFunc("class", func(base string, condition bool, activeClass string) string {
        if condition {
            return base + " " + activeClass
        }
        return base
    })
}
```

### 模板函数使用示例

```html
<!-- 时间格式化 -->
<p>发布时间: {{formatDate .Article.CreatedAt "2006-01-02 15:04:05"}}</p>
<p>更新于: {{timeAgo .Article.UpdatedAt}}</p>

<!-- 数字格式化 -->
<p>浏览量: {{formatNumber .Article.ViewCount}}</p>
<p>文件大小: {{formatSize .File.Size}}</p>

<!-- 字符串处理 -->
<p>{{truncate .Article.Content 200}}</p>

<!-- 安全HTML输出 -->
<div class="content">{{safeHTML .Article.HTMLContent}}</div>

<!-- URL生成 -->
<a href="{{url "user.profile" .User.ID}}">查看资料</a>

<!-- 条件CSS类名 -->
<div class="{{class "nav-item" .IsActive "active"}}">
    导航项
</div>
```

## 🎭 高级模板特性

### 模板继承链

```html
<!-- layouts/base.html - 基础布局 -->
<!DOCTYPE html>
<html>
<head>
    <title>{{block "title" .}}默认标题{{end}}</title>
    {{block "meta" .}}{{end}}
    {{block "css" .}}{{end}}
</head>
<body>
    {{block "body" .}}{{end}}
    {{block "js" .}}{{end}}
</body>
</html>

<!-- layouts/main.html - 继承基础布局 -->
{{define "body"}}
<header>{{template "partials/header.html" .}}</header>
<main>{{block "content" .}}{{end}}</main>
<footer>{{template "partials/footer.html" .}}</footer>
{{end}}

{{define "css"}}
<link href="/static/css/main.css" rel="stylesheet">
{{block "page-css" .}}{{end}}
{{end}}
```

### 模板数据处理

```go
func (c *ArticleController) GetList() {
    // 📊 获取文章列表
    page := c.GetInt("page", 1)
    articles, total, err := c.articleService.GetPaginatedList(page, 10)
    if err != nil {
        c.Error(500, "获取文章失败")
        return
    }
    
    // 📈 准备模板数据
    templateData := map[string]interface{}{
        "Title":       "文章列表",
        "Articles":    articles,
        "Total":       total,
        "CurrentPage": page,
        "TotalPages":  (total + 9) / 10, // 向上取整
        "HasPrev":     page > 1,
        "HasNext":     page < (total+9)/10,
        "PrevPage":    page - 1,
        "NextPage":    page + 1,
        "Categories":  c.getCategories(),
        "Tags":        c.getPopularTags(),
    }
    
    // 🎨 设置SEO数据
    templateData["MetaDescription"] = "最新文章列表，包含技术分享、生活感悟等内容"
    templateData["MetaKeywords"] = "博客,文章,技术,分享"
    
    // 📱 响应式处理
    if c.isMobile() {
        templateData["IsMobile"] = true
        c.Render("articles/mobile_list.html", templateData)
    } else {
        c.Render("articles/list.html", templateData)
    }
}
```

## 🔧 模板缓存优化

```go
func (c *BaseController) initTemplateCache() {
    // 🚀 启用模板缓存
    if c.app.Config.Environment == "production" {
        c.templateEngine.EnableCache()
        c.templateEngine.SetCacheSize(100) // 缓存100个模板
    }
    
    // 🔄 模板热重载（开发环境）
    if c.app.Config.Environment == "development" {
        c.templateEngine.EnableHotReload()
        c.templateEngine.WatchDirectory("views/")
    }
}
```

## ❓ 常见问题

**Q: 如何在模板中处理循环和条件？**
A: 使用Go模板的range、if、with等控制结构。

**Q: 模板函数如何传递多个参数？**
A: 在模板函数定义时使用可变参数，在模板中用空格分隔参数。

**Q: 如何实现模板的国际化？**
A: 使用i18n模板函数，根据语言设置返回不同的文本。

**Q: 大型项目如何管理模板？**
A: 按功能模块组织目录，使用模板继承，建立统一的命名规范。