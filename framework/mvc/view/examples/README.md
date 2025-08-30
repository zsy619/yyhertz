# YYHertz 模板引擎演示应用

这是一个基于 YYHertz MVC 框架的模板引擎演示应用，展示了 Beego 风格的模板引擎功能。

## 🚀 快速开始

### 前置条件

- Go 1.19+ 
- YYHertz 框架

### 安装和运行

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd YYHertz/framework/mvc/view/examples
   ```

2. **编译应用**
   ```bash
   go build -o template-demo main.go
   ```

3. **运行应用**
   ```bash
   ./template-demo
   ```

4. **访问应用**
   打开浏览器访问：http://localhost:8888

## 📋 功能特性

### 🎨 模板引擎功能
- **Layout 继承系统** - 支持模板布局继承
- **Beego 风格函数库** - 丰富的内置模板函数
- **高级功能** - CSRF 保护、自动模板推导、命名约定
- **性能监控** - 实时查看模板引擎性能指标
- **模板管理** - 在线浏览和预览模板文件

### 🛠️ 技术特点
- **MVC 架构** - 基于 YYHertz MVC 框架
- **RESTful 路由** - 支持自动路由和手动路由注册
- **中间件支持** - 内置多种中间件（日志、CORS、限流等）
- **静态资源** - 完整的 CSS/JS 资源管理
- **响应式设计** - 现代化的用户界面

## 🗂️ 项目结构

```
examples/
├── main.go                 # 应用入口文件
├── controllers/            # 控制器目录
│   ├── example_controller.go    # 示例控制器
│   ├── demo_controller.go       # 演示控制器
│   └── template_controller.go   # 模板管理控制器
├── views/                  # 视图模板目录
│   ├── index.html               # 首页模板
│   ├── example/                 # 示例模板
│   │   ├── layout.html          # Layout 演示
│   │   └── beego-functions.html # 函数演示
│   ├── demo/                    # 演示模板
│   │   ├── advanced.html        # 高级功能
│   │   └── performance.html     # 性能监控
│   └── template/               # 模板管理
│       ├── index.html          # 模板列表
│       └── show.html           # 模板详情
├── static/                 # 静态资源目录
│   ├── css/app.css             # 主样式文件
│   ├── js/app.js               # 主 JavaScript 文件
│   └── README.md               # 静态资源说明
├── conf/                   # 配置文件目录
└── template-demo          # 编译后的可执行文件
```

## 🌐 路由说明

| 方法 | 路径 | 控制器 | 说明 |
|------|------|--------|------|
| GET | `/` | ExampleController.Index | 模板引擎演示首页 |
| GET | `/layout` | ExampleController.LayoutDemo | Layout继承演示 |
| GET | `/beego-functions` | ExampleController.BeegoFunctions | Beego函数演示 |
| GET | `/advanced` | DemoController.AdvancedFeatures | 高级功能演示 |
| GET | `/performance` | DemoController.Performance | 性能监控 |
| POST | `/csrf-test` | DemoController.CsrfTest | CSRF测试 |
| GET | `/templates` | TemplateController.Index | 模板管理 |
| GET | `/templates/:name` | TemplateController.Show | 查看特定模板 |
| POST | `/templates/preview` | TemplateController.Preview | 模板预览 |

## 📖 使用指南

### 1. 首页功能概览
- 访问 http://localhost:8888/ 查看所有功能模块
- 每个功能卡片都有详细的说明和链接

### 2. Layout 继承演示
- URL: `/layout`
- 演示模板布局继承系统
- 包含用户信息、导航菜单等

### 3. Beego 函数演示
- URL: `/beego-functions`
- 展示丰富的内置模板函数
- 包括日期时间、字符串、数学、集合处理函数

### 4. 高级功能演示
- URL: `/advanced`
- CSRF Token 安全防护
- 自动模板路径推导
- 命名约定系统
- 自定义模板函数

### 5. 性能监控
- URL: `/performance`
- 实时查看引擎健康状态
- 缓存统计和内存使用情况
- 全局统计信息

### 6. 模板管理
- URL: `/templates`
- 浏览所有模板文件
- 支持搜索和过滤
- 在线预览功能

## 🎯 开发说明

### 控制器开发
继承 `mvc.BaseController` 创建新的控制器：

```go
type MyController struct {
    mvc.BaseController
}

func (c *MyController) Index() {
    c.SetData("Title", "页面标题")
    c.SetData("Message", "欢迎信息")
    c.TplName = "my/index.html"
    c.Render()
}
```

### 模板开发
使用标准的 Go 模板语法和 Beego 函数：

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>当前时间: {{dateformat now "2006-01-02 15:04:05"}}</p>
    <p>{{.Message}}</p>
</body>
</html>
```

### 路由注册
支持自动路由和手动路由：

```go
// 自动路由
app.RouterAuto(controller)

// 手动路由
app.RouterPrefix("/", controller, true, "Index", "*:/")
```

## 🔧 配置文件

应用使用 YAML 配置文件进行配置管理：

- `conf/app.yaml` - 应用基础配置
- `conf/template.yaml` - 模板引擎配置
- `conf/middleware.yaml` - 中间件配置
- 等其他配置文件...

## 🐛 故障排除

### 常见问题

1. **端口占用**
   ```bash
   # 查看端口占用
   lsof -i :8888
   # 修改端口（如果需要）
   # 在配置文件中更改端口设置
   ```

2. **模板找不到**
   - 检查模板文件路径是否正确
   - 确保控制器中设置了正确的 `TplName`

3. **静态资源无法加载**
   - 确认静态文件目录结构
   - 检查静态资源路径配置

### 日志调试
应用会输出详细的日志信息，包括：
- 请求处理日志
- 模板渲染日志
- 错误信息日志

## 📈 性能优化

1. **模板缓存**
   - 生产环境启用模板缓存
   - 定期清理无用的缓存

2. **静态资源**
   - 使用 CDN 加速静态资源
   - 启用 gzip 压缩

3. **数据库优化**
   - 使用连接池
   - 优化查询语句

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

本项目基于 Apache 2.0 许可证开源。

## 🔗 相关链接

- [YYHertz 框架文档](https://github.com/zsy619/yyhertz)
- [CloudWeGo Hertz](https://www.cloudwego.io/zh/docs/hertz/)
- [Go 模板语法](https://pkg.go.dev/text/template)

## 📞 联系我们

如有问题或建议，请通过以下方式联系：
- 提交 Issue
- 发送邮件
- 参与讨论

---

© 2024 YYHertz 模板引擎演示 | 基于 CloudWeGo/Hertz 构建
