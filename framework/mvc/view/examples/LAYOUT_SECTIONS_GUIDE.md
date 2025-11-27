# LayoutSections 使用指南

## 🎯 概述

LayoutSections 是 YYHertz 模板引擎的强大布局区块系统，允许 Controller 通过 `map[string]string` 类型的 `LayoutSections` 来设置不同区块的内容，实现灵活的页面布局管理。

## 🚀 核心特性

- ✅ 支持 `{{.HtmlHead}}`、`{{.SideBar}}`、`{{.Scripts}}` 等多种布局区块
- ✅ Controller 中简单的 `map[string]string` 配置
- ✅ 自动占位符替换和内容嵌入
- ✅ 高性能缓存机制
- ✅ 多种占位符格式兼容

## 📝 使用方法

### 1. 布局模板定义

在布局文件中使用区块占位符：

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <!-- HTML头部区块 -->
    {{.HtmlHead}}
</head>
<body>
    <header>{{.Header}}</header>
    <nav>{{.Navigation}}</nav>
    
    <div class="container">
        <!-- 侧边栏区块 -->
        <aside>{{.SideBar}}</aside>
        
        <!-- 主要内容 -->
        <main>{{.LayoutContent}}</main>
    </div>
    
    <footer>{{.Footer}}</footer>
    
    <!-- 脚本区块 -->
    {{.Scripts}}
</body>
</html>
```

### 2. Controller 中使用

```go
func (c *Controller) MyAction(ctx context.Context, c2 *app.RequestContext) {
    // 页面数据
    data := map[string]any{
        "Title":   "我的页面",
        "Content": "页面内容...",
    }

    // 定义布局区块内容
    layoutSections := map[string]string{
        "HtmlHead": `
            <meta name="description" content="页面描述">
            <style>
                body { font-family: Arial, sans-serif; }
                .sidebar { background: #f8f9fa; }
            </style>`,
            
        "SideBar": `
            <div class="menu">
                <h3>菜单</h3>
                <ul>
                    <li><a href="/home">首页</a></li>
                    <li><a href="/about">关于</a></li>
                </ul>
            </div>`,
            
        "Scripts": `
            <script>
                console.log('页面已加载');
                // 其他JS代码...
            </script>`,
    }

    // 使用布局区块渲染
    result, err := engine.RenderWithLayoutSections(
        "my_template", 
        "layouts/main_layout", 
        data, 
        layoutSections
    )
    
    if err != nil {
        c2.String(500, "渲染失败: %v", err)
        return
    }

    c2.HTML(200, result)
}
```

## 🎨 支持的区块类型

### 常用区块占位符

| 占位符 | 说明 | 使用场景 |
|--------|------|----------|
| `{{.HtmlHead}}` | HTML头部内容 | CSS样式、Meta标签、外部资源 |
| `{{.SideBar}}` / `{{.Sidebar}}` | 侧边栏内容 | 导航菜单、快捷链接、广告 |
| `{{.Scripts}}` | JavaScript代码 | 页面脚本、统计代码、交互功能 |
| `{{.Header}}` | 页面头部 | 标题、Logo、顶部导航 |
| `{{.Footer}}` | 页面底部 | 版权信息、链接、联系方式 |
| `{{.Navigation}}` | 导航菜单 | 主导航、面包屑导航 |
| `{{.Alert}}` | 提示信息 | 成功、错误、警告消息 |
| `{{.Modal}}` | 模态框 | 弹出窗口、确认对话框 |

### 多种格式支持

系统支持多种占位符格式：

```html
{{.HtmlHead}}           <!-- 标准格式 -->
{{ .HtmlHead }}         <!-- 带空格格式 -->
{{.HtmlHead | safehtml}} <!-- 带过滤器格式 -->
{{template "htmlhead" .}} <!-- Beego风格 -->
{{block "htmlhead" .}}{{end}} <!-- Go标准模板 -->
```

## 🔧 高级功能

### 1. 条件区块

```go
layoutSections := map[string]string{
    "SideBar": func() string {
        if userLoggedIn {
            return `<div>用户菜单...</div>`
        }
        return `<div>访客菜单...</div>`
    }(),
}
```

### 2. 动态内容

```go
layoutSections := map[string]string{
    "Scripts": fmt.Sprintf(`
        <script>
            var userId = %d;
            var userName = "%s";
            // 其他动态内容...
        </script>
    `, user.ID, user.Name),
}
```

### 3. 缓存优化

系统会自动为不同的 LayoutSections 组合生成唯一的缓存键，确保高性能渲染。

## 📊 性能特性

- **智能缓存**: 基于区块内容哈希的缓存机制
- **毫秒级渲染**: 高效的模板处理和内容替换
- **内存优化**: 合理的缓存策略避免内存泄露

## 🧪 测试验证

运行完整的测试套件：

```bash
go test -v -run TestLayoutSections
```

## 🌟 最佳实践

1. **区块职责明确**: 每个区块负责特定的页面区域
2. **内容适度**: 避免在单个区块中放入过多内容
3. **缓存友好**: 相似页面使用相同的区块组合以提高缓存效率
4. **安全考虑**: 对用户输入的内容进行适当的转义处理

## 🔄 与现有功能兼容

LayoutSections 与现有的模板功能完全兼容：

- ✅ 支持 `{{.LayoutContent}}` 主内容嵌入
- ✅ 兼容 Beego 风格的模板函数
- ✅ 支持模板继承和组件系统
- ✅ 集成 CSRF、Flash 消息等功能

---

**完成✨**: LayoutSections 布局区块系统已全面实现，支持 Controller 通过 `map[string]string` 类型的 LayoutSections 来灵活设置页面各个区块的内容！