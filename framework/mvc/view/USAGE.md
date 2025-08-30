# Beego风格模板系统使用指南

本系统实现了完整的Beego风格模板功能，提供了统一、高性能的模板渲染解决方案。

## 快速开始

### 1. 基本模板渲染
```go
// 直接渲染模板
html, err := view.UnifiedRender("user/profile", map[string]any{
    "user": user,
    "title": "用户资料",
})

// 使用布局渲染
html, err := view.UnifiedRenderWithLayout("content", "main", data)
```

### 2. 字符串模板渲染
```go
// 动态模板内容
template := "欢迎 {{.name}}，今天是 {{dateformat .date \"2006-01-02\"}}"
html, err := view.UnifiedRenderString(template, map[string]any{
    "name": "张三",
    "date": time.Now(),
})
```

## 核心功能

### 1. 自动模板推导
```go
// 自动根据控制器和动作名推导模板路径
html, err := view.AutoRender("UserController", "Login", map[string]any{
    "username": "john",
    "message": "欢迎登录",
})

// UserController.Login 会自动尝试以下路径：
// - user/login
// - user/login/index  
// - user/index
// - login
// - shared/login
// - common/login
```

### 2. 命名约定支持
```go
// 设置命名约定
view.SetNamingConvention(view.BeegoStandard) // 默认
view.SetNamingConvention(view.SnakeCase)
view.SetNamingConvention(view.CamelCase)

// 添加自定义映射
view.AddControllerMapping("UserController", "user")
view.AddActionMapping("ShowProfile", "profile")

// 应用约定转换
result := view.ApplyNamingConvention("UserController.ShowProfile", view.BeegoStandard)
// 结果: "user/profile"
```

### 3. 统一模板API
```go
// 获取统一API实例
api := view.GetUnifiedAPI()

// 基本渲染
html, err := api.Render("template_name", data)

// 使用布局渲染
html, err := api.RenderWithLayout("content", "layout", data)

// 渲染字符串模板
html, err := api.RenderString("Hello {{.name}}!", map[string]any{"name": "World"})

// 批量渲染
results, err := api.BatchRender([]string{"tmpl1", "tmpl2"}, data)
```

### 4. Beego模板函数
```go
// 在模板中使用Beego函数
template := `
<h1>{{str2html .title}}</h1>
<p>日期: {{dateformat .date "2006-01-02"}}</p>
<p>截断: {{truncate .content 100}}</p>
<p>计算: {{add .price .tax}}</p>
<p>比较: {{if gt .score 60}}及格{{else}}不及格{{end}}</p>

{{/* 表单渲染 */}}
{{renderform .form}}

{{/* 资源引入 */}}
{{assets_css "style.css" "bootstrap.css"}}
{{assets_js "jquery.js" "main.js"}}

{{/* CSRF Token */}}
<input type="hidden" name="csrf_token" value="{{csrf}}">

{{/* URL生成 */}}
<a href="{{urlfor "user/profile" "id" .user.ID}}">个人资料</a>

{{/* 组件渲染 */}}
{{component "user_card" .user}}
`
```

### 5. 便捷的全局函数
```go
// 统一渲染
html, err := view.UnifiedRender("template", data)

// 字符串渲染
html, err := view.UnifiedRenderString("Hello {{.name}}!", data)

// 使用约定渲染
html, err := view.RenderWithConvention("UserController", "Profile", data, view.BeegoStandard)

// 推导模板路径
candidates := view.InferTemplatePath("UserController", "Login")
```

### 6. 性能优化功能
```go
// 获取默认引擎并优化
engine := view.GetDefaultEngine()

// 预加载常用模板
err := engine.PreloadTemplates([]string{
    "user/login", "user/profile", "admin/dashboard",
})

// 缓存优化
engine.OptimizeCache()

// 获取性能统计
stats := engine.GetCacheStats()
memUsage := engine.GetMemoryUsage()

// 健康检查
err := engine.HealthCheck()

// 错误恢复
err := engine.Recovery()
```

### 7. 高级配置
```go
// 创建自定义配置的引擎
cfg := view.DefaultTemplateConfig()
cfg.Paths.ViewPaths = []string{"views", "templates"}
cfg.Syntax.DelimLeft = "{%"
cfg.Syntax.DelimRight = "%}"
cfg.Cache.EnableCache = true

engine, err := view.NewTemplateEngine(cfg)

// 添加自定义函数
engine.AddFunction("myFunc", func(s string) string {
    return strings.ToUpper(s)
})

// 设置主题
err = engine.SetTheme("dark")

// 添加路径模板
view.AddTemplatePathPattern("{theme}/{controller}/{action}", 120, "主题化路径")
```

## 实际应用案例

### 案例1: 用户管理系统
```go
type UserController struct {
    // controller implementation
}

func (c *UserController) Login() {
    // 使用自动推导
    html, err := view.AutoRender("UserController", "Login", map[string]any{
        "title": "用户登录",
        "action": "/login",
        "csrf": c.GetCSRFToken(),
    })
    if err != nil {
        // 降级处理
        html = "<p>模板加载失败</p>"
    }
    c.WriteHTML(html)
}

func (c *UserController) Profile() {
    user := c.GetCurrentUser()
    
    // 使用布局渲染
    html, err := view.UnifiedRenderWithLayout("user/profile", "main", map[string]any{
        "user": user,
        "title": "个人资料",
        "meta": map[string]string{
            "description": "用户个人资料页面",
            "keywords": "用户,资料,设置",
        },
    })
    
    c.WriteHTML(html)
}
```

### 案例2: 电商商品展示
```go
func (c *ProductController) List() {
    products := c.GetProducts()
    
    // 批量渲染相关模板
    templates := []string{"product/list", "component/pagination"}
    results, err := view.GetUnifiedAPI().BatchRender(templates, map[string]any{
        "products": products,
        "total": len(products),
        "page": c.GetPage(),
    })
    
    if err != nil {
        c.HandleError(err)
        return
    }
    
    c.WriteHTML(results["product/list"])
}
```

### 案例3: 动态邮件模板
```go
func SendWelcomeEmail(user User) error {
    // 动态生成邮件模板
    emailTemplate := `
    <h1>欢迎 {{.name}}！</h1>
    <p>感谢您在 {{dateformat .joinDate "2006年1月2日"}} 加入我们。</p>
    <p>您的用户ID是: {{.id}}</p>
    
    {{if gt .level 1}}
    <p>恭喜您获得了VIP{{.level}}级别的特权！</p>
    {{end}}
    
    <p>请点击 <a href="{{urlfor "user/activate" "token" .token}}">这里</a> 激活账户。</p>
    `
    
    html, err := view.UnifiedRenderString(emailTemplate, map[string]any{
        "name": user.Name,
        "id": user.ID,
        "level": user.Level,
        "token": user.ActivationToken,
        "joinDate": user.CreatedAt,
    })
    
    return sendEmail(user.Email, "欢迎加入", html)
}
```

## 最佳实践

### 1. 模板组织结构
```
views/
├── layouts/
│   ├── main.html
│   └── admin.html
├── components/
│   ├── header.html
│   └── footer.html  
├── user/
│   ├── login.html
│   ├── profile.html
│   └── index.html
└── admin/
    ├── dashboard.html
    └── users.html
```

### 2. 性能优化建议
```go
// 启用缓存
cfg.Cache.EnableCache = true

// 预加载常用模板
engine.PreloadTemplates([]string{"user/login", "user/profile"})

// 定期优化缓存
go func() {
    ticker := time.NewTicker(30 * time.Minute)
    for range ticker.C {
        engine.OptimizeCache()
    }
}()

// 使用批量渲染
requests := []view.RenderRequest{
    {TemplateName: "email1", Data: data1},
    {TemplateName: "email2", Data: data2},
}
results, err := engine.BulkRender(requests)
```

### 3. 错误处理
```go
// 带错误恢复的渲染
func SafeRender(templateName string, data any) (string, error) {
    html, err := view.UnifiedRender(templateName, data)
    if err != nil {
        // 尝试自动推导
        if strings.Contains(err.Error(), "not found") {
            return view.AutoRender("Default", "Error", map[string]any{
                "error": err.Error(),
                "data": data,
            })
        }
        
        // 尝试错误恢复
        engine := view.GetDefaultEngine()
        if recoverErr := engine.Recovery(); recoverErr == nil {
            return view.UnifiedRender(templateName, data)
        }
    }
    return html, err
}
```

### 4. 中间件集成
```go
// 模板渲染中间件
func TemplateMiddleware() MiddlewareFunc {
    return func(ctx Context, next HandlerFunc) error {
        // 注入模板函数
        view.AddUnifiedFunction("currentUser", func() any {
            return ctx.GetUser()
        })
        
        // 注入CSRF token
        if csrf := ctx.GetCSRFToken(); csrf != "" {
            ctx.Set("csrf_token", csrf)
        }
        
        return next(ctx)
    }
}
```

## 调试和监控

```go
// 启用调试模式
options := &view.TemplateLoadOptions{
    Debug: true,
    StrictMode: false,
}
html, err := view.LoadTemplateWithOptions("template", data, options)

// 获取统计信息
stats := view.GetUnifiedStats()
templateStats := view.GetTemplateStatistics()

// 健康检查
engine := view.GetDefaultEngine()
if err := engine.HealthCheck(); err != nil {
    log.Printf("Template engine health check failed: %v", err)
}

// 性能监控
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        stats := engine.GetMemoryUsage()
        if stats["total_memory"] > 100*1024*1024 { // 超过100MB
            log.Printf("模板引擎内存使用过高: %d bytes", stats["total_memory"])
            engine.OptimizeCache()
        }
    }
}()
```

## 注意事项

1. **并发安全**: 所有公共API都是并发安全的
2. **内存管理**: 定期调用`OptimizeCache()`清理无效缓存
3. **错误处理**: 使用`Recovery()`方法处理严重错误
4. **性能监控**: 定期检查`GetMemoryUsage()`和`GetCacheStats()`
5. **模板热重载**: 在开发环境中启用`cfg.Reload.Enabled = true`
6. **CSRF保护**: 所有表单都应包含CSRF token
7. **XSS防护**: 使用`str2html`函数时要确保内容安全

这个系统完全兼容Beego的模板使用习惯，同时提供了更强的性能和稳定性。