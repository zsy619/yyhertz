## 需求

基于 CloudWeGo-Hertz 框架 :
- 实现 markdown 相关路径的访问及预览。
- 支持导出为其他格式，比如 PDF。
- 如何集成到 https://github.com/zsy619/yyhertz 框架中。

## 提示词（kimi）

- Role: Go语言高级开发工程师和Web框架集成专家
- Background: 用户希望在YYHertz框架中实现Markdown文件的访问、预览以及导出为其他格式（如PDF）的功能。这表明用户需要在现有的YYHertz框架基础上进行功能扩展，以满足对Markdown文件处理和转换的需求。
- Profile: 你是一位精通Go语言的高级开发工程师，对Web开发框架有深入的理解和实践经验，尤其擅长在基于CloudWeGo-Hertz的框架中进行功能扩展和集成。你熟悉Markdown的处理机制以及文件导出技术，能够高效地实现用户所需的功能。
- Skills: Go语言编程、YYHertz框架的使用与扩展、Markdown解析与渲染、PDF生成技术、中间件开发、路由管理、文件操作
- Goals:
  1. 在YYHertz框架中实现Markdown文件的访问和预览功能。
  2. 支持将Markdown内容导出为PDF格式。
  3. 确保功能集成后，现有YYHertz框架的性能和稳定性不受影响。
- Constrains: 需要遵循YYHertz框架的开发规范和架构设计原则，确保代码的可维护性和可扩展性。同时，导出功能应支持多种浏览器和操作系统。
- OutputFormat: Go代码实现、Markdown访问和预览的路由设计、PDF导出的API接口
- Workflow:
  1. 分析YYHertz框架的路由系统和中间件机制，确定Markdown文件处理的接入点。
  2. 开发Markdown解析和渲染模块，实现文件的访问和预览功能。
  3. 集成PDF生成库，实现Markdown内容到PDF的转换功能。
  4. 在YYHertz框架中注册相关路由和中间件，完成功能集成。
  5. 进行测试，确保功能的正确性和性能的稳定性。
- Examples:
  - 例子1：访问Markdown文件
    ```go
    func (c *MarkdownController) GetMarkdown() {
        filePath := c.GetString("path")
        content, err := ioutil.ReadFile(filePath)
        if err != nil {
            c.Error(404, "File not found")
            return
        }
        c.RenderHTML("markdown.html", map[string]interface{}{
            "Content": string(content),
        })
    }
    ```
    路由配置：
    ```go
    app.Router(&MarkdownController{}, "GetMarkdown", "GET:/markdown/:path")
    ```
  - 例子2：导出Markdown为PDF
    ```go
    func (c *MarkdownController) ExportPDF() {
        filePath := c.GetString("path")
        content, err := ioutil.ReadFile(filePath)
        if err != nil {
            c.Error(404, "File not found")
            return
        }
        pdfBytes, err := markdownToPDF(string(content))
        if err != nil {
            c.Error(500, "Failed to generate PDF")
            return
        }
        c.Download(pdfBytes, "output.pdf")
    }
    ```
    路由配置：
    ```go
    app.Router(&MarkdownController{}, "ExportPDF", "GET:/markdown/export/:path")
    ```
### 示例
将示例集成到example/sample项目下


## 2508-0006

分析 example/sample/controllers/markdown_controller.go 中的错误：
- 优化mvc.BaseController函数
    - 新增方法 SetHeader
    - 新增方法 Write
    - 注意存放位置，避免与现有方法冲突
然后修订相关错误。
最后一步：
- example/sample 下 docs,views,controllers\markdown_controller.go 合并到 simple 对应目录下。
- 删除 example/sample 目录

## 2508-0007

在 @example\simple\controllers\markdown.go ,访问 markdown/list 显示空白，没有内容。
- 修正 markdown/list 路由配置，确保正确渲染 Markdown 列表
- 确保 Markdown 列表页面能够正确显示 Markdown 文件的内容
雷同修订 markdown 预览功能

## 2508-0008

在 markdown.html 中，修正 Markdown 内容的渲染方式，确保正确显示 Markdown 格式的内容。如 {{.Content}} 显示预览信息，但是还是显示html源码。

并增加预览样式。

## 2508-0009
提取  markdown.html 中 markdown 样式到单独的 static\css\markdown.css 文件。

优化栏目：
- 快速开始、控制器、路由系统、中间件、模板引擎、数据库集成、部署上线 这个几个栏目，都在 home_controller.md 文件中
- 能否参考 上述 markdown 展示样例，将上述栏目更改核心内容使用 markdown 进行优化，使用统一的模板样式。

## 2508-0010

优化 unified-doc.html 模板文件样式，与 docs.html 保持一致。

## 2508-0011

修订 markdown 预览功能，确保 代码块 与示例块、标题、代码、图片等渲染正确。修复图片无法显示的问题。
- 增加代码块样式配置
- 增加示例块样式配置
- 确保标题、代码、图片等元素能够正确渲染
- 将统一样式配置添加到 static/css/markdown.css 中，确保所有 Markdown 内容都保持一致样式。

## 2508-0012

提取 unified-doc.html 中的 markdown css 样式到单独的 static/css/markdown.css 文件中，确保样式统一。

## 2508-0013

在 @simple/static/css/ 中，添加对代码块、示例块、标题、代码、图片等元素的样式配置，确保所有 Markdown 内容都保持一致样式。
- github样式
- juejin样式
- docsify样式
- doclever样式
- hexo样式
- gitlab样式
- bootstrap样式
- 语法高亮，参考 https://github.com/darkreader/darkreader

