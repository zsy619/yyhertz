package view

import (
	"fmt"
	"html/template"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 模板引擎示例 =============

// RunTemplateExample 运行模板引擎示例
func RunTemplateExample() error {
	config.Info("Starting template engine example...")

	// 1. 创建模板引擎
	templateConfig := DefaultTemplateConfig()
	templateConfig.ViewPaths = []string{"example/views", "views"}
	templateConfig.EnableReload = true
	templateConfig.EnableCache = true

	engine, err := NewTemplateEngine(templateConfig)
	if err != nil {
		return fmt.Errorf("failed to create template engine: %w", err)
	}
	defer engine.Close()

	// 2. 添加自定义模板函数
	engine.AddFunction("customGreeting", func(name string) string {
		return fmt.Sprintf("Hello, %s! Welcome to our site.", name)
	})

	engine.AddFunction("formatPrice", func(price float64, currency string) string {
		return fmt.Sprintf("%.2f %s", price, currency)
	})

	// 3. 基本模板渲染示例
	config.Info("=== Basic Template Rendering ===")

	// 准备示例数据
	userData := map[string]any{
		"name":     "张三",
		"email":    "zhangsan@example.com",
		"age":      28,
		"active":   true,
		"balance":  1256.78,
		"joinDate": time.Now().AddDate(-1, -3, -15),
		"tags":     []string{"VIP", "活跃用户", "早期用户"},
		"profile": map[string]any{
			"avatar":   "/static/images/avatar.jpg",
			"bio":      "热爱技术的开发者",
			"location": "北京",
		},
	}

	// 使用RenderData结构
	renderData := &RenderData{
		Data: userData,
		Meta: &MetaData{
			Title:       "用户信息页面",
			Description: "查看用户详细信息",
			Keywords:    "用户,个人资料,信息",
		},
		Theme: "default",
		Request: &RequestData{
			Method:    "GET",
			Path:      "/user/profile",
			Timestamp: time.Now().Unix(),
		},
	}

	// 模拟渲染基本模板
	config.Info("Rendering user profile template...")
	profileHTML := simulateTemplateRender("user/profile", renderData)
	config.Infof("Profile template rendered (%d characters)", len(profileHTML))

	// 4. 布局继承示例
	config.Info("=== Layout Inheritance Example ===")

	// 创建布局管理器
	layoutManager := NewLayoutManager(engine)
	if err := RegisterDefaultLayouts(layoutManager); err != nil {
		return fmt.Errorf("failed to register layouts: %w", err)
	}

	// 创建布局渲染器
	_ = NewLayoutRenderer(layoutManager, engine)

	// 渲染带布局的页面
	pageData := map[string]any{
		"title": "产品列表",
		"products": []map[string]any{
			{"id": 1, "name": "笔记本电脑", "price": 5999.0, "category": "电子产品"},
			{"id": 2, "name": "无线鼠标", "price": 129.0, "category": "配件"},
			{"id": 3, "name": "机械键盘", "price": 399.0, "category": "配件"},
		},
		"totalCount":  3,
		"currentPage": 1,
	}

	config.Info("Rendering product list with app layout...")
	productHTML := simulateLayoutRender("product/list", "app", pageData)
	config.Infof("Product list rendered with layout (%d characters)", len(productHTML))

	// 5. 组件系统示例
	config.Info("=== Component System Example ===")

	// 创建组件管理器
	componentManager := NewComponentManager(engine)
	if err := RegisterDefaultComponents(componentManager); err != nil {
		return fmt.Errorf("failed to register components: %w", err)
	}

	// 渲染导航栏组件
	navProps := map[string]any{
		"brand": "我的网站",
		"items": []map[string]any{
			{"name": "首页", "url": "/", "active": true},
			{"name": "产品", "url": "/products", "active": false},
			{"name": "关于", "url": "/about", "active": false},
		},
		"theme": "dark",
	}

	navSlots := map[string]template.HTML{
		"actions": template.HTML(`<button class="btn btn-primary">登录</button>`),
	}

	config.Info("Rendering navbar component...")
	navHTML := simulateComponentRender("navbar", navProps, navSlots)
	config.Infof("Navbar component rendered (%d characters)", len(navHTML))

	// 渲染表格组件
	tableProps := map[string]any{
		"columns": []map[string]any{
			{"key": "id", "title": "ID", "width": "80px"},
			{"key": "name", "title": "名称", "sortable": true},
			{"key": "price", "title": "价格", "align": "right"},
			{"key": "category", "title": "分类"},
		},
		"data":       pageData["products"],
		"striped":    true,
		"pagination": true,
	}

	config.Info("Rendering table component...")
	tableHTML := simulateComponentRender("table", tableProps, nil)
	config.Infof("Table component rendered (%d characters)", len(tableHTML))

	// 6. 主题切换示例
	config.Info("=== Theme Switching Example ===")

	// 添加自定义主题
	adminTheme := &ThemeConfig{
		Name:          "admin",
		ViewPaths:     []string{"admin/views", "views"},
		LayoutPath:    "admin/layouts",
		ComponentPath: "admin/components",
		StaticPath:    "admin/static",
		Enabled:       true,
		Variables: map[string]string{
			"primaryColor":   "#007bff",
			"secondaryColor": "#6c757d",
			"brandName":      "管理后台",
		},
	}

	if err := engine.AddTheme("admin", adminTheme); err != nil {
		return fmt.Errorf("failed to add admin theme: %w", err)
	}

	// 切换到管理主题
	if err := engine.SetTheme("admin"); err != nil {
		return fmt.Errorf("failed to switch to admin theme: %w", err)
	}

	config.Infof("Switched to theme: %s", engine.GetCurrentTheme())

	// 在新主题下渲染页面
	adminData := map[string]any{
		"title": "用户管理",
		"users": []map[string]any{
			{"id": 1, "username": "admin", "role": "管理员", "status": "活跃"},
			{"id": 2, "username": "editor", "role": "编辑", "status": "活跃"},
			{"id": 3, "username": "viewer", "role": "查看者", "status": "禁用"},
		},
		"stats": map[string]any{
			"totalUsers":    3,
			"activeUsers":   2,
			"inactiveUsers": 1,
		},
	}

	config.Info("Rendering admin page with admin theme...")
	adminHTML := simulateLayoutRender("admin/users", "admin", adminData)
	config.Infof("Admin page rendered (%d characters)", len(adminHTML))

	// 7. 模板函数示例
	config.Info("=== Template Functions Example ===")

	// 演示各种模板函数的使用
	funcTestData := map[string]any{
		"user": map[string]any{
			"name":     "李四",
			"balance":  1234.56,
			"joinDate": time.Now().AddDate(-2, 0, 0),
			"bio":      "这是一个很长的个人简介，需要截断显示。这里有更多的内容来演示截断功能。",
			"fileSize": 1024 * 1024 * 15, // 15MB
		},
		"products": []map[string]any{
			{"name": "产品A", "price": 199.99},
			{"name": "产品B", "price": 299.99},
			{"name": "产品C", "price": 399.99},
		},
	}

	config.Info("Demonstrating template functions...")
	funcHTML := simulateTemplateFunctions(funcTestData)
	config.Infof("Template functions demo rendered (%d characters)", len(funcHTML))

	// 8. 缓存和性能示例
	config.Info("=== Caching and Performance Example ===")

	// 获取引擎统计信息
	stats := engine.GetStats()
	config.Infof("Template engine stats: %+v", stats)

	// 性能测试
	startTime := time.Now()
	for i := 0; i < 100; i++ {
		_ = simulateTemplateRender("user/profile", renderData)
	}
	duration := time.Since(startTime)
	config.Infof("Rendered 100 templates in %v (%.2fms per template)",
		duration, float64(duration.Nanoseconds())/1000000/100)

	// 清除缓存测试
	engine.ClearCache()
	config.Info("Template cache cleared")

	// 9. 错误处理示例
	config.Info("=== Error Handling Example ===")

	// 尝试渲染不存在的模板
	if _, err := engine.Render("nonexistent/template", nil); err != nil {
		config.Infof("Expected error for nonexistent template: %v", err)
	}

	// 尝试使用不存在的布局
	if err := engine.SetTheme("nonexistent-theme"); err != nil {
		config.Infof("Expected error for nonexistent theme: %v", err)
	}

	config.Info("Template engine example completed successfully!")
	return nil
}

// ============= 模拟渲染函数 =============

// simulateTemplateRender 模拟模板渲染
func simulateTemplateRender(templateName string, data any) string {
	// 模拟HTML输出
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <meta charset="UTF-8">
</head>
<body>
    <div class="container">
        <h1>模板: %s</h1>
        <div class="content">
            <!-- 这里是模拟的模板内容 -->
            <p>Data: %T</p>
            <p>Rendered at: %s</p>
        </div>
    </div>
</body>
</html>`, templateName, templateName, data, time.Now().Format("2006-01-02 15:04:05"))

	return html
}

// simulateLayoutRender 模拟布局渲染
func simulateLayoutRender(templateName, layoutName string, data any) string {
	// 模拟带布局的HTML输出
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Layout: %s | Template: %s</title>
    <meta charset="UTF-8">
    <link rel="stylesheet" href="/static/css/app.css">
</head>
<body>
    <header class="header">
        <nav class="navbar">Layout Navigation</nav>
    </header>
    
    <main class="main-content">
        <div class="container">
            <h1>Content from: %s</h1>
            <div class="template-content">
                <!-- 这里是模拟的模板内容 -->
                <p>Layout: %s</p>
                <p>Data: %T</p>
                <p>Rendered at: %s</p>
            </div>
        </div>
    </main>
    
    <footer class="footer">
        <div class="container">Layout Footer</div>
    </footer>
    
    <script src="/static/js/app.js"></script>
</body>
</html>`, layoutName, templateName, templateName, layoutName, data, time.Now().Format("2006-01-02 15:04:05"))

	return html
}

// simulateComponentRender 模拟组件渲染
func simulateComponentRender(componentName string, props map[string]any, slots map[string]template.HTML) string {
	// 模拟组件HTML输出
	html := fmt.Sprintf(`<!-- Component: %s -->
<div class="component component-%s">
    <div class="component-props">
        <!-- Props: %+v -->
    </div>
    <div class="component-slots">
        <!-- Slots: %d slots -->
    </div>
    <div class="component-content">
        <p>Component: %s rendered successfully</p>
        <p>Rendered at: %s</p>
    </div>
</div>
<!-- End Component: %s -->`,
		componentName, componentName, props, len(slots), componentName,
		time.Now().Format("2006-01-02 15:04:05"), componentName)

	return html
}

// simulateTemplateFunctions 模拟模板函数使用
func simulateTemplateFunctions(data map[string]any) string {
	// 模拟使用各种模板函数的HTML输出
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Template Functions Demo</title>
</head>
<body>
    <div class="functions-demo">
        <h1>模板函数演示</h1>
        
        <section class="date-functions">
            <h2>日期格式化函数</h2>
            <p>日期: {{ dateFormat .user.joinDate "date" }}</p>
            <p>日期时间: {{ dateFormat .user.joinDate "datetime" }}</p>
            <p>ISO格式: {{ dateFormat .user.joinDate "iso" }}</p>
        </section>
        
        <section class="string-functions">
            <h2>字符串函数</h2>
            <p>截断文本: {{ truncate .user.bio 50 }}</p>
            <p>大写: {{ upper .user.name }}</p>
            <p>小写: {{ lower .user.name }}</p>
        </section>
        
        <section class="number-functions">
            <h2>数字格式化函数</h2>
            <p>货币: {{ currency .user.balance "CNY" }}</p>
            <p>文件大小: {{ filesize .user.fileSize }}</p>
        </section>
        
        <section class="array-functions">
            <h2>数组和对象函数</h2>
            <p>产品数量: {{ len .products }}</p>
            {{ range .products }}
            <div>产品: {{ .name }} - {{ currency .price "CNY" }}</div>
            {{ end }}
        </section>
        
        <section class="utility-functions">
            <h2>实用函数</h2>
            <p>创建字典: {{ dict "key1" "value1" "key2" "value2" }}</p>
            <p>创建数组: {{ slice 1 2 3 4 5 }}</p>
            <p>数字范围: {{ range createRange 1 5 }}{{ . }} {{ end }}</p>
        </section>
    </div>
</body>
</html>`

	return html
}

// ============= 模板示例生成器 =============

// GenerateExampleTemplates 生成示例模板文件
func GenerateExampleTemplates() error {
	config.Info("Generating example template files...")

	// 这里可以生成实际的模板文件到文件系统
	// 为了演示，我们只打印模板内容

	templates := map[string]string{
		"layouts/base.html":      generateBaseLayoutTemplate(),
		"layouts/app.html":       generateAppLayoutTemplate(),
		"components/navbar.html": generateNavbarComponentTemplate(),
		"components/table.html":  generateTableComponentTemplate(),
		"user/profile.html":      generateUserProfileTemplate(),
		"product/list.html":      generateProductListTemplate(),
	}

	for path, content := range templates {
		config.Infof("Generated template: %s (%d characters)", path, len(content))
		// 在实际应用中，这里会写入文件系统
		// ioutil.WriteFile(path, []byte(content), 0644)
	}

	config.Info("Example templates generated successfully!")
	return nil
}

// generateBaseLayoutTemplate 生成基础布局模板
func generateBaseLayoutTemplate() string {
	return `<!DOCTYPE html>
<html lang="{{ .lang | default "zh-CN" }}">
<head>
    <meta charset="{{ .charset | default "UTF-8" }}">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ block "title" . }}{{ .meta.title | default "默认标题" }}{{ end }}</title>
    
    {{ block "meta" . }}
    <meta name="description" content="{{ .meta.description }}">
    <meta name="keywords" content="{{ .meta.keywords }}">
    {{ end }}
    
    {{ block "styles" . }}
    <link rel="stylesheet" href="{{ asset "css/bootstrap.min.css" }}">
    <link rel="stylesheet" href="{{ asset "css/app.css" }}">
    {{ end }}
</head>
<body>
    {{ block "content" . }}{{ end }}
    
    {{ block "scripts" . }}
    <script src="{{ asset "js/jquery.min.js" }}"></script>
    <script src="{{ asset "js/bootstrap.min.js" }}"></script>
    <script src="{{ asset "js/app.js" }}"></script>
    {{ end }}
</body>
</html>`
}

// generateAppLayoutTemplate 生成应用布局模板
func generateAppLayoutTemplate() string {
	return `{{ define "app" }}
{{ template "base" . }}

{{ define "content" }}
<div class="app-layout">
    {{ component "header" (dict "title" .meta.title) }}
    
    <div class="app-body">
        {{ block "sidebar" . }}
        {{ component "sidebar" .sidebar }}
        {{ end }}
        
        <main class="main-content">
            {{ component "breadcrumb" .breadcrumb }}
            
            <div class="content-wrapper">
                {{ block "main" . }}{{ end }}
            </div>
        </main>
    </div>
    
    {{ component "footer" .footer }}
</div>
{{ end }}
{{ end }}`
}

// generateNavbarComponentTemplate 生成导航栏组件模板
func generateNavbarComponentTemplate() string {
	return `<nav class="navbar navbar-{{ prop "theme" "light" }}">
    <div class="container-fluid">
        {{ slot "brand" }}
        <a class="navbar-brand" href="/">{{ prop "brand" "Brand" }}</a>
        {{ end }}
        
        <div class="navbar-nav">
            {{ slot "items" }}
            {{ range prop "items" }}
            <a class="nav-link {{ if .active }}active{{ end }}" href="{{ .url }}">
                {{ .name }}
            </a>
            {{ end }}
            {{ end }}
        </div>
        
        {{ slot "actions" }}
        <div class="navbar-actions">
            <!-- Default actions -->
        </div>
        {{ end }}
    </div>
</nav>`
}

// generateTableComponentTemplate 生成表格组件模板
func generateTableComponentTemplate() string {
	return `<div class="table-wrapper">
    {{ slot "header" }}
    <div class="table-header">
        <h3>{{ prop "title" "数据表格" }}</h3>
    </div>
    {{ end }}
    
    <table class="table {{ if prop "striped" }}table-striped{{ end }} {{ if prop "bordered" }}table-bordered{{ end }}">
        <thead>
            <tr>
                {{ range prop "columns" }}
                <th style="{{ if .width }}width: {{ .width }};{{ end }} text-align: {{ .align | default "left" }};">
                    {{ .title }}
                    {{ if .sortable }}<i class="sort-icon"></i>{{ end }}
                </th>
                {{ end }}
            </tr>
        </thead>
        <tbody>
            {{ range prop "data" }}
            <tr>
                {{ range $ := . }}{{ range $col := $.columns }}
                <td style="text-align: {{ $col.align | default "left" }};">
                    {{ index . $col.key }}
                </td>
                {{ end }}{{ end }}
            </tr>
            {{ end }}
        </tbody>
    </table>
    
    {{ if prop "pagination" }}
    {{ slot "pagination" }}
    {{ component "pagination" .pagination }}
    {{ end }}
    {{ end }}
</div>`
}

// generateUserProfileTemplate 生成用户资料模板
func generateUserProfileTemplate() string {
	return `{{ define "title" }}{{ .data.name }} - 用户资料{{ end }}

<div class="user-profile">
    <div class="profile-header">
        <div class="avatar">
            <img src="{{ .data.profile.avatar }}" alt="Avatar">
        </div>
        <div class="basic-info">
            <h1>{{ customGreeting .data.name }}</h1>
            <p class="email">{{ .data.email }}</p>
            <p class="join-date">加入时间: {{ dateFormat .data.joinDate "date" }}</p>
        </div>
    </div>
    
    <div class="profile-body">
        <div class="row">
            <div class="col-md-8">
                <div class="card">
                    <div class="card-header">
                        <h3>个人信息</h3>
                    </div>
                    <div class="card-body">
                        <p><strong>年龄:</strong> {{ .data.age }}</p>
                        <p><strong>状态:</strong> {{ if .data.active }}活跃{{ else }}非活跃{{ end }}</p>
                        <p><strong>余额:</strong> {{ currency .data.balance "CNY" }}</p>
                        <p><strong>简介:</strong> {{ .data.profile.bio }}</p>
                        <p><strong>位置:</strong> {{ .data.profile.location }}</p>
                    </div>
                </div>
            </div>
            
            <div class="col-md-4">
                <div class="card">
                    <div class="card-header">
                        <h3>用户标签</h3>
                    </div>
                    <div class="card-body">
                        {{ range .data.tags }}
                        <span class="badge badge-primary">{{ . }}</span>
                        {{ end }}
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>`
}

// generateProductListTemplate 生成产品列表模板
func generateProductListTemplate() string {
	return `{{ define "title" }}{{ .title }}{{ end }}

<div class="product-list">
    <div class="page-header">
        <h1>{{ .title }}</h1>
        <p>共找到 {{ .totalCount }} 个产品</p>
    </div>
    
    {{ component "table" (dict 
        "columns" (slice 
            (dict "key" "id" "title" "ID" "width" "80px")
            (dict "key" "name" "title" "产品名称" "sortable" true)
            (dict "key" "price" "title" "价格" "align" "right")
            (dict "key" "category" "title" "分类")
        )
        "data" .products
        "striped" true
        "pagination" true
    ) }}
    
    <div class="pagination-wrapper">
        {{ component "pagination" (dict 
            "current" .currentPage
            "total" .totalCount
            "pageSize" 10
        ) }}
    </div>
</div>`
}
