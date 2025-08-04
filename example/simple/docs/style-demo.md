# 🎨 Markdown样式主题演示

这是一个综合性的Markdown样式演示页面，展示了各种元素在不同主题下的渲染效果。

## 📝 基本文本样式

这是一个普通段落，包含**粗体文本**、*斜体文本*、~~删除线文本~~和`行内代码`。

### 链接和强调

- 这是一个[外部链接](https://github.com/darkreader/darkreader)
- 这是**重要的粗体文本**
- 这是*强调的斜体文本*
- 这是==高亮文本==（如果支持）

## 💻 代码块展示

### Go语言代码

```go
package main

import (
    "fmt"
    "github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
    // 创建应用实例
    app := mvc.HertzApp
    
    // 创建控制器
    homeController := &controllers.HomeController{}
    
    // 注册路由
    app.RouterPrefix("/", homeController, "GetIndex", "GET:/")
    
    // 启动服务器
    app.Run(":8080")
}
```

### JavaScript代码

```javascript
class MarkdownThemeSwitcher {
    constructor() {
        this.themes = {
            github: { name: 'GitHub风格' },
            juejin: { name: '掘金风格' }
        };
    }
    
    loadTheme(themeKey) {
        const theme = this.themes[themeKey];
        console.log(`Loading theme: ${theme.name}`);
    }
}
```

### Python代码

```python
def fibonacci(n):
    """计算斐波那契数列的第n项"""
    if n <= 1:
        return n
    else:
        return fibonacci(n-1) + fibonacci(n-2)

# 测试函数
for i in range(10):
    print(f"F({i}) = {fibonacci(i)}")
```

### CSS样式

```css
.markdown-content {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto;
    line-height: 1.6;
    color: #333;
}

.markdown-content h1 {
    font-size: 2em;
    border-bottom: 2px solid #667eea;
    padding-bottom: 10px;
}
```

### Shell脚本

```bash
#!/bin/bash

# 启动Hertz MVC服务器
echo "正在启动服务器..."
go build -o app main.go
./app > server.log 2>&1 &

echo "服务器已启动，访问: http://localhost:8888"
```

## 📋 列表展示

### 无序列表

- 第一项内容
- 第二项内容
  - 嵌套子项目1
  - 嵌套子项目2
    - 深层嵌套项目
- 第三项内容

### 有序列表

1. 首先需要安装Go环境
2. 然后克隆项目代码
3. 运行以下命令：
   1. `go mod tidy`
   2. `go build`
   3. `./app`

### 任务列表

- [x] 创建GitHub风格样式
- [x] 创建掘金风格样式  
- [x] 创建Docsify风格样式
- [ ] 测试所有主题兼容性
- [ ] 优化移动端显示效果

## 📊 表格展示

| 主题名称 | 样式特点 | 适用场景 | 推荐指数 |
|---------|---------|---------|----------|
| GitHub | 简洁清晰 | 代码文档 | ⭐⭐⭐⭐⭐ |
| 掘金 | 现代美观 | 技术博客 | ⭐⭐⭐⭐⭐ |
| Docsify | 文档专用 | 项目文档 | ⭐⭐⭐⭐ |
| DocLever | API专用 | 接口文档 | ⭐⭐⭐⭐ |
| Hexo | 博客风格 | 个人博客 | ⭐⭐⭐⭐ |
| GitLab | 企业级 | 团队协作 | ⭐⭐⭐⭐ |
| Bootstrap | 组件丰富 | Web应用 | ⭐⭐⭐⭐⭐ |
| DarkReader | 护眼模式 | 夜间阅读 | ⭐⭐⭐⭐⭐ |

## 💡 引用和提示

> 这是一个重要的引用块。
> 
> 引用可以包含多个段落，也可以包含**格式化文本**和`代码`。
> 
> —— 来自YYHertz框架开发团队

### 信息提示框

💡 **提示**：使用右上角的主题切换器可以实时预览不同样式效果。

⚠️ **警告**：某些主题可能需要刷新页面才能完全生效。

❌ **错误**：如果样式无法加载，请检查静态文件服务配置。

✅ **成功**：所有主题都已正确配置并可以正常使用。

## 🖼️ 图片展示

![Hertz Logo](https://avatars.githubusercontent.com/u/44036562?s=200&v=4 "CloudWeGo Hertz")

*图片说明：CloudWeGo Hertz框架Logo*

## 📐 数学公式（如果支持）

行内公式：$E = mc^2$

块级公式：
$$
\sum_{i=1}^{n} x_i = x_1 + x_2 + \cdots + x_n
$$

## 🔗 分割线

上面是内容部分

---

下面是总结部分

## 📚 嵌套内容

### 代码中的注释

```go
// 这是一个示例函数
func ExampleFunction() {
    /*
     * 多行注释示例
     * 展示不同的注释风格
     */
    fmt.Println("Hello, World!") // 行末注释
}
```

### 引用中的列表

> 重要提醒：
> 
> 1. 选择合适的主题样式
> 2. 考虑目标用户群体
> 3. 测试各种设备兼容性
> 
> 记住：好的样式能提升阅读体验！

## 🎯 总结

这个演示页面展示了YYHertz框架支持的多种Markdown样式主题：

- **GitHub风格**：简洁专业，适合代码文档
- **掘金风格**：现代美观，适合技术博客
- **Docsify风格**：清晰明了，适合项目文档
- **DocLever风格**：专业简约，适合API文档
- **Hexo风格**：优雅精致，适合个人博客
- **GitLab风格**：企业级设计，适合团队协作
- **Bootstrap风格**：组件丰富，适合Web应用
- **DarkReader风格**：暗色主题，适合夜间阅读

每种主题都经过精心设计，确保在不同场景下都能提供优秀的阅读体验。使用右上角的主题切换器可以实时预览效果！

---

*最后更新：2025年8月4日*