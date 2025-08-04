# YYHertz Markdown 功能演示

欢迎使用 YYHertz 框架的 Markdown 功能演示！

## 🌟 功能特性

- ✅ **Markdown 解析**：基于 Goldmark 库，支持 GitHub Flavored Markdown
- ✅ **实时预览**：在线预览 Markdown 内容，支持语法高亮
- ✅ **PDF 导出**：一键导出为 PDF 格式，支持自定义样式
- ✅ **文件管理**：文档列表浏览，支持文件摘要显示
- ✅ **响应式设计**：适配移动端和桌面端

## 📋 支持的 Markdown 语法

### 标题

支持 1-6 级标题：

# 一级标题
## 二级标题
### 三级标题
#### 四级标题
##### 五级标题
###### 六级标题

### 文本样式

- **粗体文本**
- *斜体文本*
- ~~删除线~~
- `行内代码`

### 列表

无序列表：
- 项目 1
- 项目 2
  - 子项目 2.1
  - 子项目 2.2
- 项目 3

有序列表：
1. 第一项
2. 第二项
3. 第三项

任务列表：
- [x] 已完成任务
- [ ] 待完成任务
- [ ] 另一个待完成任务

### 链接和图片

[YYHertz 框架](https://github.com/zsy619/yyhertz)

### 代码块

```go
package main

import (
    "fmt"
    "github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
    app := mvc.HertzApp
    
    app.RouterPrefix("/", &controllers.HomeController{}, "GetIndex", "GET:/")
    
    fmt.Println("🚀 YYHertz 启动成功!")
    app.Run(":8080")
}
```

```javascript
// JavaScript 示例
function hello() {
    console.log('Hello, YYHertz!');
}

hello();
```

### 表格

| 功能 | 支持状态 | 说明 |
|------|----------|------|
| Markdown 解析 | ✅ | 基于 Goldmark |
| PDF 导出 | ✅ | 使用 Rod 浏览器 |
| 语法高亮 | ✅ | 代码块高亮显示 |
| 表格支持 | ✅ | 如本表格所示 |

### 引用

> 这是一个引用块。
> 
> 支持多行引用，可以包含 **粗体** 和 *斜体* 文本。
> 
> > 嵌套引用也是支持的。

### 分隔线

---

## 🔧 技术实现

### 后端技术栈

- **Go 语言**：高性能的编程语言
- **YYHertz 框架**：基于 CloudWeGo-Hertz 的 Web 框架
- **Goldmark**：Markdown 解析库
- **Rod**：无头浏览器，用于 PDF 生成

### 前端技术

- **HTML5**：语义化标记
- **CSS3**：现代样式设计
- **JavaScript**：交互功能
- **响应式设计**：适配各种设备

## 📁 项目结构

```
example/sample/
├── controllers/
│   └── markdown_controller.go  # Markdown 控制器
├── views/
│   └── markdown/
│       ├── markdown.html       # 预览模板
│       └── list.html          # 列表模板
├── docs/                      # Markdown 文档目录
│   ├── sample.md             # 示例文档
│   └── guide.md              # 使用指南
└── main.go                   # 应用入口
```

## 🚀 快速开始

1. **启动服务**：
   ```bash
   cd example/sample
   go run main.go
   ```

2. **访问应用**：
   - 打开浏览器访问：http://localhost:8891
   - 查看文档列表
   - 点击文档进行预览
   - 导出 PDF 文件

3. **添加文档**：
   - 在 `docs/` 目录下添加 `.md` 文件
   - 刷新页面即可看到新文档

## 📝 使用说明

### 查看文档
访问 `/markdown/{filename}` 查看指定文档，例如：
- `/markdown/sample` - 查看本示例文档
- `/markdown/guide` - 查看使用指南

### 导出PDF
访问 `/markdown/export/{filename}` 导出 PDF，例如：
- `/markdown/export/sample` - 导出本文档的 PDF

### API 接口
- `GET /markdown/list` - 获取文档列表
- `GET /markdown/{path}` - 预览文档
- `GET /markdown/export/{path}` - 导出 PDF

## 🎨 自定义样式

可以通过修改模板文件来自定义预览样式：

- **预览样式**：编辑 `views/markdown/markdown.html`
- **列表样式**：编辑 `views/markdown/list.html`
- **PDF 样式**：修改控制器中的 PDF 样式配置

## 🔍 更多功能

- **语法高亮**：支持多种编程语言
- **数学公式**：可扩展支持 LaTeX 数学公式
- **图表支持**：可集成 Mermaid 等图表库
- **主题切换**：可实现明暗主题切换

---

**© 2025 YYHertz Markdown 功能演示**

这是一个基于 YYHertz 框架的完整 Markdown 解决方案示例。
