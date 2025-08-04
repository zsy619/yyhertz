# YYHertz Markdown 使用指南

## 🎯 概述

本指南将帮助您快速上手 YYHertz 框架的 Markdown 功能。

## 📦 安装依赖

确保您的项目已添加以下依赖：

```bash
go get github.com/yuin/goldmark
go get github.com/yuin/goldmark/extension
go get github.com/yuin/goldmark/renderer/html
go get github.com/go-rod/rod
```

## 🏗️ 架构设计

### MVC 模式

YYHertz 采用经典的 MVC（Model-View-Controller）架构：

- **Model**：数据层，处理 Markdown 文件读取
- **View**：视图层，HTML 模板渲染
- **Controller**：控制层，业务逻辑处理

### 组件说明

#### MarkdownController

```go
type MarkdownController struct {
    core.Controller
}
```

主要方法：
- `GetMarkdown()`：获取并渲染 Markdown 文档
- `ExportPDF()`：导出 PDF 文件
- `GetList()`：获取文档列表

#### 模板引擎

使用 YYHertz 内置的模板引擎，支持：
- 模板继承
- 部分模板
- 自定义函数
- 数据绑定

## 🔧 配置说明

### 路由配置

```go
// 自动路由
app.AutoRouters(markdownController)

// 手动路由
app.RouterPrefix("/markdown", markdownController, "GetList", "GET:/list")
app.RouterPrefix("/markdown", markdownController, "GetMarkdown", "GET:/:path")
app.RouterPrefix("/markdown/export", markdownController, "ExportPDF", "GET:/:path")
```

### 中间件配置

```go
app.Use(
    middleware.RecoveryMiddleware(),    // 异常恢复
    middleware.TracingMiddleware(),     // 链路追踪
    middleware.LoggerMiddleware(),      // 日志记录
    middleware.CORSMiddleware(),        // CORS 支持
    middleware.RateLimitMiddleware(100, time.Minute), // 限流
)
```

## 📁 目录结构

```
example/sample/
├── controllers/          # 控制器目录
├── views/               # 视图模板目录
│   └── markdown/       # Markdown 相关模板
├── docs/               # 文档存储目录
├── static/             # 静态资源目录
├── conf/               # 配置文件目录
└── main.go            # 应用入口文件
```

## 🎨 模板开发

### 数据传递

控制器向模板传递数据：

```go
c.RenderHTML("markdown.html", map[string]interface{}{
    "Title":       title,
    "Content":     htmlContent,
    "RawContent":  rawContent,
    "FilePath":    filePath,
})
```

### 模板语法

```html
<!-- 变量输出 -->
{{.Title}}

<!-- 条件判断 -->
{{if .Files}}
    <!-- 内容 -->
{{else}}
    <!-- 空状态 -->
{{end}}

<!-- 循环遍历 -->
{{range .Files}}
    <div>{{.Name}}</div>
{{end}}

<!-- 函数调用 -->
{{len .Files}}
```

## 🔐 安全考虑

### 路径安全

```go
// 防止路径遍历攻击
if strings.Contains(filePath, "..") {
    c.Error(403, "Invalid file path")
    return
}
```

### 文件类型限制

```go
// 确保文件扩展名
if !strings.HasSuffix(fullPath, ".md") {
    fullPath += ".md"
}
```

### 内容过滤

Goldmark 配置：

```go
md := goldmark.New(
    goldmark.WithRendererOptions(
        html.WithUnsafe(), // 根据需要启用/禁用
    ),
)
```

## 📊 性能优化

### 缓存策略

可以添加文件内容缓存：

```go
// 伪代码示例
type FileCache struct {
    content    string
    modTime    time.Time
    htmlCache  string
}

var fileCache = make(map[string]*FileCache)
```

### PDF 生成优化

```go
// 复用浏览器实例
var browserInstance *rod.Browser

func init() {
    launcher := launcher.New().Headless(true)
    browserInstance = rod.New().ControlURL(launcher.MustLaunch())
}
```

## 🐛 调试技巧

### 日志记录

```go
log.Printf("Processing file: %s", filePath)
log.Printf("Generated HTML length: %d", len(htmlContent))
```

### 错误处理

```go
if err != nil {
    log.Printf("Error reading file %s: %v", filePath, err)
    c.Error(404, "File not found")
    return
}
```

## 🧪 测试

### 单元测试

```go
func TestMarkdownController_GetMarkdown(t *testing.T) {
    // 测试代码
}
```

### 集成测试

```bash
curl -X GET http://localhost:8891/markdown/sample
curl -X GET http://localhost:8891/markdown/export/sample
```

## 🔄 扩展功能

### 添加语法高亮

```go
import "github.com/yuin/goldmark-highlighting"

md := goldmark.New(
    goldmark.WithExtensions(
        highlighting.Highlighting,
    ),
)
```

### 添加数学公式支持

```go
import "github.com/litao91/goldmark-mathjax"

md := goldmark.New(
    goldmark.WithExtensions(
        mathjax.MathJax,
    ),
)
```

### 自定义渲染器

```go
type CustomRenderer struct {
    html.Config
}

func (r *CustomRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
    // 自定义渲染逻辑
}
```

## 📚 常见问题

### Q: PDF 导出失败怎么办？

A: 检查以下几点：
1. 确保 Chrome/Chromium 已安装
2. 检查系统权限
3. 查看错误日志

### Q: 如何添加自定义样式？

A: 修改模板文件中的 CSS 样式，或者添加外部样式表。

### Q: 支持哪些 Markdown 扩展？

A: 当前支持：
- GitHub Flavored Markdown
- 表格
- 删除线
- 任务列表
- 定义列表

## 🔗 相关资源

- [YYHertz 官方文档](https://github.com/zsy619/yyhertz)
- [Goldmark 文档](https://github.com/yuin/goldmark)
- [Rod 浏览器文档](https://github.com/go-rod/rod)
- [CloudWeGo Hertz](https://www.cloudwego.io/zh/docs/hertz/)

---

**祝您使用愉快！** 🎉
