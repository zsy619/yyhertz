# 🎨 模板引擎

YYHertz提供了强大的模板引擎，支持模板继承、组件化、自定义函数等特性。

## 基础模板语法

### 变量输出

```html
<!-- 基本变量输出 -->
<h1>{{.Title}}</h1>
<p>{{.Content}}</p>

<!-- 对象属性访问 -->
<h2>{{.User.Name}}</h2>
<p>{{.User.Email}}</p>

<!-- 数组/切片访问 -->
<p>第一个元素: {{index .Items 0}}</p>
```

### 条件判断

```html
<!-- if-else 语句 -->
{{if .User}}
    <p>欢迎, {{.User.Name}}!</p>
{{else}}
    <p>请先登录</p>
{{end}}

<!-- 多条件判断 -->
{{if eq .Status "active"}}
    <span class="badge bg-success">活跃</span>
{{else if eq .Status "inactive"}}
    <span class="badge bg-warning">非活跃</span>
{{else}}
    <span class="badge bg-danger">未知状态</span>
{{end}}

<!-- 复杂条件 -->
{{if and .User (gt .User.Age 18)}}
    <p>成年用户</p>
{{end}}
```

### 循环遍历

```html
<!-- 遍历数组/切片 -->
{{range .Users}}
    <div class="user-card">
        <h3>{{.Name}}</h3>
        <p>{{.Email}}</p>
    </div>
{{else}}
    <p>没有用户数据</p>
{{end}}

<!-- 带索引的遍历 -->
{{range $index, $user := .Users}}
    <div class="user-item" data-index="{{$index}}">
        <span>{{add $index 1}}. {{$user.Name}}</span>
    </div>
{{end}}

<!-- 遍历Map -->
{{range $key, $value := .Settings}}
    <div>{{$key}}: {{$value}}</div>
{{end}}
```

## 模板函数

### 内置函数

```html
<!-- 字符串函数 -->
<p>{{printf "Hello, %s!" .Name}}</p>
<p>{{len .Items}} 个项目</p>

<!-- 数学函数 -->
<p>总价: {{add .Price .Tax}}</p>
<p>折扣后: {{sub .Price .Discount}}</p>

<!-- 比较函数 -->
{{if gt .Score 80}}
    <span class="text-success">优秀</span>
{{else if gt .Score 60}}
    <span class="text-warning">及格</span>
{{else}}
    <span class="text-danger">不及格</span>
{{end}}

<!-- 字符串处理 -->
<p>{{.Content | html}}</p>  <!-- HTML转义 -->
<p>{{.Content | js}}</p>    <!-- JavaScript转义 -->
<p>{{.Content | urlquery}}</p> <!-- URL编码 -->
```

### 自定义函数

```go
// 在控制器中注册自定义函数
func (c *BaseController) SetTemplateFuncs() {
    funcs := template.FuncMap{
        // 格式化日期
        "formatDate": func(t time.Time) string {
            return t.Format("2006-01-02 15:04:05")
        },
        
        // 截取字符串
        "truncate": func(s string, length int) string {
            if len(s) <= length {
                return s
            }
            return s[:length] + "..."
        },
        
        // 货币格式化
        "currency": func(amount float64) string {
            return fmt.Sprintf("¥%.2f", amount)
        },
        
        // 数组包含检查
        "contains": func(slice []string, item string) bool {
            for _, v := range slice {
                if v == item {
                    return true
                }
            }
            return false
        },
        
        // Markdown渲染
        "markdown": func(content string) template.HTML {
            md := goldmark.New()
            var buf bytes.Buffer
            md.Convert([]byte(content), &buf)
            return template.HTML(buf.String())
        },
    }
    
    c.SetFuncMap(funcs)
}
```

## 模板继承

### 基础布局

创建 `views/layout/base.html`：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}默认标题{{end}} - YYHertz</title>
    
    <!-- 基础样式 -->
    <link href="/static/css/bootstrap.min.css" rel="stylesheet">
    <link href="/static/css/app.css" rel="stylesheet">
    
    <!-- 页面特定样式 -->
    {{block "styles" .}}{{end}}
</head>
<body>
    <!-- 导航栏 -->
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
        <div class="container">
            <a class="navbar-brand" href="/">YYHertz</a>
            
            <div class="navbar-nav ms-auto">
                {{if .User}}
                    <a class="nav-link" href="/profile">{{.User.Name}}</a>
                    <a class="nav-link" href="/logout">退出</a>
                {{else}}
                    <a class="nav-link" href="/login">登录</a>
                    <a class="nav-link" href="/register">注册</a>
                {{end}}
            </div>
        </div>
    </nav>
    
    <!-- 主要内容 -->
    <main class="container mt-4">
        {{block "content" .}}
            <p>内容区域</p>
        {{end}}
    </main>
    
    <!-- 页脚 -->
    <footer class="bg-light mt-5 py-4">
        <div class="container text-center">
            <p>&copy; 2025 YYHertz Framework. All rights reserved.</p>
        </div>
    </footer>
    
    <!-- 基础脚本 -->
    <script src="/static/js/bootstrap.min.js"></script>
    
    <!-- 页面特定脚本 -->
    {{block "scripts" .}}{{end}}
</body>
</html>
```

### 页面模板

创建 `views/user/profile.html`：

```html
{{define "title"}}用户资料{{end}}

{{define "styles"}}
<style>
.profile-card { 
    max-width: 600px; 
    margin: 0 auto; 
}
</style>
{{end}}

{{define "content"}}
<div class="profile-card">
    <div class="card">
        <div class="card-header">
            <h2>用户资料</h2>
        </div>
        
        <div class="card-body">
            <div class="row mb-3">
                <div class="col-sm-3">
                    <strong>姓名:</strong>
                </div>
                <div class="col-sm-9">
                    {{.User.Name}}
                </div>
            </div>
            
            <div class="row mb-3">
                <div class="col-sm-3">
                    <strong>邮箱:</strong>
                </div>
                <div class="col-sm-9">
                    {{.User.Email}}
                </div>
            </div>
            
            <div class="row mb-3">
                <div class="col-sm-3">
                    <strong>注册时间:</strong>
                </div>
                <div class="col-sm-9">
                    {{formatDate .User.CreatedAt}}
                </div>
            </div>
        </div>
        
        <div class="card-footer">
            <a href="/user/edit" class="btn btn-primary">编辑资料</a>
        </div>
    </div>
</div>
{{end}}

{{define "scripts"}}
<script>
console.log('用户资料页面加载完成');
</script>
{{end}}
```

## 模板组件

### 可复用组件

创建 `views/components/user-card.html`：

```html
{{define "user-card"}}
<div class="card user-card">
    {{if .User.Avatar}}
        <img src="{{.User.Avatar}}" class="card-img-top" alt="{{.User.Name}}">
    {{end}}
    
    <div class="card-body">
        <h5 class="card-title">{{.User.Name}}</h5>
        <p class="card-text">{{truncate .User.Bio 100}}</p>
        
        <div class="user-meta">
            <small class="text-muted">
                注册于 {{formatDate .User.CreatedAt}}
            </small>
        </div>
    </div>
    
    <div class="card-footer">
        <a href="/user/{{.User.ID}}" class="btn btn-primary btn-sm">查看详情</a>
        {{if .ShowActions}}
            <a href="/user/{{.User.ID}}/edit" class="btn btn-outline-secondary btn-sm">编辑</a>
        {{end}}
    </div>
</div>
{{end}}
```

### 使用组件

在页面中使用组件：

```html
{{define "content"}}
<div class="row">
    {{range .Users}}
        <div class="col-md-4 mb-4">
            {{template "user-card" dict "User" . "ShowActions" true}}
        </div>
    {{end}}
</div>
{{end}}
```

## 模板数据处理

### 控制器中准备数据

```go
func (c *UserController) GetProfile() {
    userID := c.GetParam("id")
    user := getUserByID(userID)
    
    // 基础数据
    c.SetData("Title", "用户资料")
    c.SetData("User", user)
    
    // 计算字段
    c.SetData("Age", calculateAge(user.Birthday))
    c.SetData("PostCount", getPostCountByUser(userID))
    
    // 格式化数据
    c.SetData("FormattedJoinDate", user.CreatedAt.Format("2006年01月02日"))
    
    // 权限数据
    c.SetData("CanEdit", c.currentUser.ID == user.ID || c.currentUser.IsAdmin)
    
    // 渲染模板，使用布局
    c.RenderHTMLWithLayout("user/profile.html", "layout/base.html")
}
```

### 复杂数据结构

```go
func (c *DashboardController) GetIndex() {
    // 统计数据
    stats := map[string]interface{}{
        "TotalUsers":    getUserCount(),
        "ActiveUsers":   getActiveUserCount(),
        "TotalPosts":    getPostCount(),
        "TodayPosts":    getTodayPostCount(),
    }
    
    // 图表数据
    chartData := map[string]interface{}{
        "Labels": []string{"1月", "2月", "3月", "4月", "5月", "6月"},
        "Data":   []int{120, 150, 180, 220, 200, 250},
    }
    
    // 近期活动
    recentActivities := getRecentActivities(10)
    
    c.SetData("Stats", stats)
    c.SetData("ChartData", chartData)
    c.SetData("Activities", recentActivities)
    
    c.RenderHTML("dashboard/index.html")
}
```

## 模板缓存

### 开发环境

```go
// 开发模式 - 每次重新加载模板
func (app *App) SetDevelopmentMode() {
    app.Config.TemplateCache = false
    app.Config.AutoReload = true
}
```

### 生产环境

```go
// 生产模式 - 缓存模板
func (app *App) SetProductionMode() {
    app.Config.TemplateCache = true
    app.Config.AutoReload = false
    
    // 预编译所有模板
    app.PrecompileTemplates()
}
```

## 模板安全

### XSS防护

```html
<!-- 自动转义 (推荐) -->
<p>{{.UserInput}}</p>

<!-- 输出原始HTML (谨慎使用) -->
<div>{{.TrustedHTML | raw}}</div>

<!-- 手动转义 -->
<script>
var data = {{.JSONData | js}};
</script>
```

### CSRF保护

```html
<!-- 表单中添加CSRF令牌 -->
<form method="POST" action="/user/update">
    {{.CSRFToken}}
    
    <input type="text" name="name" value="{{.User.Name}}">
    <button type="submit">更新</button>
</form>
```

## 模板调试

### 调试信息

```html
{{if .Debug}}
<div class="debug-info">
    <h4>调试信息</h4>
    <pre>{{printf "%+v" .}}</pre>
</div>
{{end}}
```

### 模板错误处理

```go
func (c *BaseController) RenderTemplate(template string, data map[string]interface{}) {
    defer func() {
        if err := recover(); err != nil {
            log.Printf("模板渲染错误: %v", err)
            c.Error(500, "页面渲染失败")
        }
    }()
    
    c.RenderHTML(template, data)
}
```

## 最佳实践

### 1. 目录结构

```
views/
├── layout/           # 布局模板
│   ├── base.html
│   ├── admin.html
│   └── auth.html
├── components/       # 可复用组件
│   ├── header.html
│   ├── footer.html
│   ├── sidebar.html
│   └── user-card.html
├── user/            # 用户相关页面
│   ├── profile.html
│   ├── settings.html
│   └── list.html
└── admin/           # 管理页面
    ├── dashboard.html
    └── users.html
```

### 2. 命名规范

```html
<!-- 页面模板 -->
{{define "user-profile"}}...{{end}}

<!-- 组件模板 -->
{{define "component-user-card"}}...{{end}}

<!-- 布局模板 -->
{{define "layout-base"}}...{{end}}
```

### 3. 性能优化

```go
// 预编译模板
func (app *App) PrecompileTemplates() {
    templates := []string{
        "layout/base.html",
        "components/*.html",
        "user/*.html",
    }
    
    for _, pattern := range templates {
        app.CompileTemplatePattern(pattern)
    }
}

// 模板缓存
var templateCache = make(map[string]*template.Template)

func GetTemplate(name string) *template.Template {
    if tmpl, exists := templateCache[name]; exists {
        return tmpl
    }
    
    tmpl := template.Must(template.ParseFiles(name))
    templateCache[name] = tmpl
    return tmpl
}
```

---

模板引擎是前端展示的核心，合理使用模板可以构建出美观、高效的用户界面！