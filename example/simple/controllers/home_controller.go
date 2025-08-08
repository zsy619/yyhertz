package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/zsy619/yyhertz/framework/mvc"
)

type HomeController struct {
	mvc.BaseController
}

// 渲染Markdown文档的辅助方法 (支持分组目录)
func (c *HomeController) renderMarkdownDoc(docName, title string) {
	c.renderMarkdownDocWithGroup("", docName, title)
}

// 渲染分组Markdown文档的方法
func (c *HomeController) renderMarkdownDocWithGroup(group, docName, title string) {
	var docPath string
	if group == "" {
		// 兼容旧版本的文档路径
		docPath = filepath.Join("./docs", docName+".md")
	} else {
		// 新的分组文档路径
		docPath = filepath.Join("./docs", group, docName+".md")
	}

	// 读取markdown文件
	log.Printf("尝试读取文档: %s", docPath)
	content, err := os.ReadFile(docPath)
	if err != nil {
		log.Printf("读取文档失败: %s, 错误: %v", docPath, err)
		c.Error(404, fmt.Sprintf("文档不存在: %s", docName))
		return
	}
	log.Printf("成功读取文档，长度: %d", len(content))

	// 配置Goldmark解析器
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,            // GitHub Flavored Markdown
			extension.Table,          // 表格支持
			extension.Strikethrough,  // 删除线
			extension.TaskList,       // 任务列表
			extension.DefinitionList, // 定义列表
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // 硬换行
			html.WithXHTML(),     // XHTML兼容
			html.WithUnsafe(),    // 允许HTML标签
		),
	)

	// 解析Markdown为HTML
	var htmlBuf bytes.Buffer
	if err := md.Convert(content, &htmlBuf); err != nil {
		log.Printf("Markdown解析失败: %v", err)
		c.Error(500, "文档解析失败")
		return
	}

	// 设置模板数据
	c.SetData("Title", title)
	c.SetData("CurrentDoc", docName)
	c.SetData("Content", template.HTML(htmlBuf.String()))
	if group != "" {
		// 设置分组名称用于面包屑导航
		groupNames := map[string]string{
			"getting-started": "📖 开始使用",
			"mvc-core":        "🏗️ MVC核心",
			"middleware":      "🔌 中间件系统",
			"data-access":     "🗄️ 数据访问",
			"view-template":   "🎨 视图模板",
			"configuration":   "⚙️ 配置管理",
			"advanced":        "🔧 高级功能",
			"dev-tools":       "🛠️ 开发工具",
		}
		c.SetData("GroupName", groupNames[group])
	}

	// 渲染统一模板
	log.Printf("=== 准备渲染模板: home/docs/unified-doc.html ===")
	log.Printf("=== 模板数据: Title=%s, CurrentDoc=%s, Content长度=%d ===",
		title, docName, len(htmlBuf.String()))
	c.RenderHTML("home/docs/unified-doc.html")
	log.Printf("=== 模板渲染完成 ===")
}

func (c *HomeController) GetIndex() {
	// 模拟首页数据
	features := []map[string]any{
		{
			"Title":       "基于Controller",
			"Description": "类似Beego的Controller结构，让开发更简单",
			"Icon":        "fas fa-code",
		},
		{
			"Title":       "HTML模板支持",
			"Description": "内置模板引擎，支持布局和组件化开发",
			"Icon":        "fas fa-file-code",
		},
		{
			"Title":       "中间件机制",
			"Description": "丰富的中间件支持，包括认证、日志、限流等",
			"Icon":        "fas fa-layer-group",
		},
		{
			"Title":       "RESTful路由",
			"Description": "支持RESTful风格的路由设计，API开发更规范",
			"Icon":        "fas fa-route",
		},
	}

	statistics := map[string]any{
		"Controllers": 15,
		"Routes":      45,
		"Middleware":  8,
		"Templates":   12,
	}

	c.SetData("Title", "首页")
	c.SetData("Features", features)
	c.SetData("Statistics", statistics)
	c.SetData("Message", "欢迎使用Hertz MVC框架！")

	// 暂时使用完整HTML版本，确保页面正常显示
	c.RenderHTML("home/index.html")
}

func (c *HomeController) GetAbout() {
	about := map[string]any{
		"Framework": "Hertz MVC",
		"Version":   "1.0.0",
		"Author":    "CloudWeGo Team",
		"License":   "Apache 2.0",
		"Github":    "https://github.com/zsy619/yyhertz",
		"Docs":      "https://yyhertz.hn24365.com",
	}

	c.SetData("Title", "关于我们")
	c.SetData("About", about)
	c.RenderHTML("home/about.html")
}

func (c *HomeController) GetDocs() {
	docs := []map[string]any{
		{
			"Title":       "快速开始",
			"Description": "学习如何快速搭建一个Hertz MVC应用",
			"Link":        "/home/quickstart",
		},
		{
			"Title":       "控制器",
			"Description": "了解如何创建和使用控制器",
			"Link":        "/home/controller",
		},
		{
			"Title":       "路由",
			"Description": "掌握路由配置和RESTful API设计",
			"Link":        "/home/routing",
		},
		{
			"Title":       "中间件",
			"Description": "学习中间件的使用和自定义开发",
			"Link":        "/home/middleware",
		},
		{
			"Title":       "模板",
			"Description": "了解模板引擎的使用方法",
			"Link":        "/home/template",
		},
		{
			"Title":       "日志",
			"Description": "了解日志系统的集成",
			"Link":        "/home/logging",
		},
	}

	c.SetData("Title", "文档")
	c.SetData("Docs", docs)
	c.RenderHTML("home/docs.html")
}

func (c *HomeController) PostContact() {
	name := c.GetForm("name")
	email := c.GetForm("email")
	message := c.GetForm("message")

	if name == "" || email == "" || message == "" {
		c.JSON(map[string]any{
			"success": false,
			"message": "请填写完整信息",
		})
		return
	}

	// 这里应该是发送邮件或保存留言的逻辑
	c.JSON(map[string]any{
		"success": true,
		"message": "感谢您的留言，我们会尽快回复！",
		"data": map[string]any{
			"name":    name,
			"email":   email,
			"message": message,
		},
	})
}

// ============= 文档系统路由 =============

// 快速开始文档
func (c *HomeController) GetQuickstart() {
	c.renderMarkdownDoc("quickstart", "快速开始")
}

// 控制器文档
func (c *HomeController) GetController() {
	c.renderMarkdownDoc("controller", "控制器")
}

// 路由文档
func (c *HomeController) GetRouting() {
	c.renderMarkdownDoc("routing", "路由系统")
}

// 中间件文档
func (c *HomeController) GetMiddlewares() {
	c.renderMarkdownDoc("middlewares", "中间件系统")
}

// 模板文档
func (c *HomeController) GetTemplate() {
	c.renderMarkdownDoc("template", "模板引擎")
}

// 数据库集成文档
func (c *HomeController) GetDatabase() {
	c.renderMarkdownDoc("database", "数据库集成")
}

// MyBatis集成文档
func (c *HomeController) GetMybatis() {
	c.renderMarkdownDoc("mybatis", "MyBatis集成")
}

// 系统日志文档
func (c *HomeController) GetLogging() {
	c.renderMarkdownDoc("logging", "系统日志")
}

// 系统配置文档
func (c *HomeController) GetConfig() {
	c.renderMarkdownDoc("config", "系统配置")
}

// 部署文档
func (c *HomeController) GetDeployment() {
	c.renderMarkdownDoc("deployment", "部署上线")
}

// ============= 新文档体系路由 (基于8大分组) =============

// ============= 📖 开始使用分组 =============

// 概览与安装文档
func (c *HomeController) GetOverview() {
	log.Printf("=== GetOverview方法被调用 ===")
	log.Printf("=== 开始调用renderMarkdownDocWithGroup ===")
	c.renderMarkdownDocWithGroup("getting-started", "overview", "概览与安装")
	log.Printf("=== renderMarkdownDocWithGroup调用结束 ===")
}

// 简单测试方法
func (c *HomeController) GetTest() {
	log.Printf("=== GetTest方法被调用 ===")

	// 测试简单模板渲染
	c.SetData("Title", "Simple Test")
	c.SetData("Content", "This is a simple test content")
	log.Printf("=== 尝试渲染简单模板 ===")
	c.RenderHTML("home/index.html")
	log.Printf("=== 简单模板渲染完成 ===")
}

// HTML测试方法
func (c *HomeController) GetHtmlTest() {
	log.Printf("=== GetHtmlTest方法被调用 ===")
	c.SetData("Title", "HTML测试")
	c.SetData("Message", "这是直接的HTML测试内容")
	c.RenderHTML("home/docs/unified-doc.html")
	log.Printf("=== HTML测试渲染完成 ===")
}

// 项目结构文档
func (c *HomeController) GetStructure() {
	c.renderMarkdownDocWithGroup("getting-started", "structure", "项目结构")
}

// ============= 🏗️ MVC核心分组 =============

// 应用程序文档
func (c *HomeController) GetApplication() {
	c.renderMarkdownDocWithGroup("mvc-core", "application", "应用程序")
}

// 命名空间文档
func (c *HomeController) GetNamespace() {
	c.renderMarkdownDocWithGroup("mvc-core", "namespace", "命名空间")
}

// ============= 🔌 中间件系统分组 =============

// 中间件概览文档
func (c *HomeController) GetMiddlewareOverview() {
	c.renderMarkdownDocWithGroup("middleware", "overview", "中间件概览")
}

// 内置中间件文档
func (c *HomeController) GetBuiltinMiddleware() {
	c.renderMarkdownDocWithGroup("middleware", "builtin", "内置中间件")
}

// 自定义中间件文档
func (c *HomeController) GetCustomMiddleware() {
	c.renderMarkdownDocWithGroup("middleware", "custom", "自定义中间件")
}

// 中间件配置文档
func (c *HomeController) GetMiddlewareConfig() {
	c.renderMarkdownDocWithGroup("middleware", "config", "中间件配置")
}

// ============= 🗄️ 数据访问分组 =============

// GORM集成文档
func (c *HomeController) GetGorm() {
	c.renderMarkdownDocWithGroup("data-access", "gorm", "GORM集成")
}

// 数据库配置文档
func (c *HomeController) GetDatabaseConfig() {
	c.renderMarkdownDocWithGroup("data-access", "database-config", "数据库配置")
}

// 事务管理文档
func (c *HomeController) GetTransaction() {
	c.renderMarkdownDocWithGroup("data-access", "transaction", "事务管理")
}

// ============= 🎨 视图模板分组 =============

// 模板引擎文档
func (c *HomeController) GetTemplateEngine() {
	c.renderMarkdownDocWithGroup("view-template", "template-engine", "模板引擎")
}

// 视图渲染文档
func (c *HomeController) GetViewRendering() {
	c.renderMarkdownDocWithGroup("view-template", "view-rendering", "视图渲染")
}

// 静态资源文档
func (c *HomeController) GetStaticAssets() {
	c.renderMarkdownDocWithGroup("view-template", "static-assets", "静态资源")
}

// ============= ⚙️ 配置管理分组 =============

// 应用配置文档
func (c *HomeController) GetAppConfig() {
	c.renderMarkdownDocWithGroup("configuration", "app-config", "应用配置")
}

// 环境配置文档
func (c *HomeController) GetEnvironment() {
	c.renderMarkdownDocWithGroup("configuration", "environment", "环境配置")
}

// ============= 🔧 高级功能分组 =============

// 会话管理文档
func (c *HomeController) GetSession() {
	c.renderMarkdownDocWithGroup("advanced", "session", "会话管理")
}

// 缓存系统文档
func (c *HomeController) GetCache() {
	c.renderMarkdownDocWithGroup("advanced", "cache", "缓存系统")
}

// 验证系统文档
func (c *HomeController) GetValidation() {
	c.renderMarkdownDocWithGroup("advanced", "validation", "验证系统")
}

// 验证码功能文档
func (c *HomeController) GetCaptcha() {
	c.renderMarkdownDocWithGroup("advanced", "captcha", "验证码功能")
}

// 任务调度文档
func (c *HomeController) GetScheduler() {
	c.renderMarkdownDocWithGroup("advanced", "scheduler", "任务调度")
}

// ============= 🛠️ 开发工具分组 =============

// 代码生成文档
func (c *HomeController) GetCodegen() {
	c.renderMarkdownDocWithGroup("dev-tools", "codegen", "代码生成")
}

// 热重载文档
func (c *HomeController) GetHotReload() {
	c.renderMarkdownDocWithGroup("dev-tools", "hot-reload", "热重载")
}

// 性能监控文档
func (c *HomeController) GetPerformance() {
	c.renderMarkdownDocWithGroup("dev-tools", "performance", "性能监控")
}

// 测试工具文档
func (c *HomeController) GetTesting() {
	c.renderMarkdownDocWithGroup("dev-tools", "testing", "测试工具")
}
