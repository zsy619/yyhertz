package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/zsy619/yyhertz/framework/mvc"
)

type DocsController struct {
	mvc.BaseController
}

// 渲染文档的辅助方法
func (c *DocsController) renderDoc(category, doc, title string, trimCategory ...bool) {
	var docPath string

	// 构建文档路径
	if category == "" {
		docPath = filepath.Join("./docs", doc+".md")
	} else {
		if len(trimCategory) > 0 && trimCategory[0] {
			myDoc := strings.TrimPrefix(doc, category+"-")
			docPath = filepath.Join("./docs", category, myDoc+".md")
		} else {
			docPath = filepath.Join("./docs", category, doc+".md")
		}
	}

	log.Printf("尝试读取文档: %s", docPath)
	content, err := os.ReadFile(docPath)
	if err != nil {
		log.Printf("读取文档失败: %s, 错误: %v", docPath, err)
		c.Error(404, fmt.Sprintf("文档不存在: %s", doc))
		return
	}

	// 配置Goldmark解析器
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			extension.DefinitionList,
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
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
	c.SetData("CurrentDoc", doc)
	c.SetData("Content", template.HTML(htmlBuf.String()))
	c.SetData("Category", category)

	// 设置分组名称用于面包屑导航
	if category != "" {
		categoryNames := map[string]string{
			"getting-started": "📖 开始使用",
			"mvc-core":        "🏗️ MVC核心",
			"middleware":      "🔌 中间件",
			"data-access":     "🗄️ 数据访问",
			"view-template":   "🎨 视图渲染",
			"configuration":   "⚙️ 配置管理",
			"advanced":        "🔧 高级功能",
			"deployment":      "☁️ 部署运维",
			"dev-tools":       "🛠️ 开发工具",
		}
		c.SetData("CategoryName", categoryNames[category])
	}

	// 渲染统一模板
	c.RenderHTML("home/docs/unified-doc.html")
}

// ============= 开始使用 =============
func (c *DocsController) GetGettingStartedOverview() {
	c.renderDoc("getting-started", "overview", "概览与安装")
}

func (c *DocsController) GetGettingStartedQuickstart() {
	c.renderDoc("getting-started", "quickstart", "快速开始")
}

func (c *DocsController) GetGettingStartedStructure() {
	c.renderDoc("getting-started", "structure", "项目结构")
}

// ============= MVC核心 =============
func (c *DocsController) GetMvcCoreApplication() {
	c.renderDoc("mvc-core", "application", "应用程序")
}

func (c *DocsController) GetMvcCoreController() {
	c.renderDoc("mvc-core", "controller", "控制器")
}

func (c *DocsController) GetMvcCoreRouting() {
	c.renderDoc("mvc-core", "routing", "路由系统")
}

func (c *DocsController) GetMvcCoreNamespace() {
	c.renderDoc("mvc-core", "namespace", "命名空间")
}

func (c *DocsController) GetMvcCoreAnnotation() {
	c.renderDoc("mvc-core", "annotation", "注解路由系统")
}

func (c *DocsController) GetMvcCoreComment() {
	c.renderDoc("mvc-core", "comment", "注释路由系统")
}

// ============= 中间件 =============
func (c *DocsController) GetMiddlewareOverview() {
	c.renderDoc("middleware", "middleware-overview", "中间件概览", true)
}

func (c *DocsController) GetMiddlewareBuiltin() {
	c.renderDoc("middleware", "middleware-builtin", "内置中间件", true)
}

func (c *DocsController) GetMiddlewareCustom() {
	c.renderDoc("middleware", "custom", "自定义中间件")
}

func (c *DocsController) GetMiddlewareConfig() {
	c.renderDoc("middleware", "config", "中间件配置")
}

// ============= 数据访问 =============
func (c *DocsController) GetDataAccessGorm() {
	c.renderDoc("data-access", "gorm", "GORM集成")
}

// MyBatis基础集成
func (c *DocsController) GetDataAccessMybatisBasic() {
	c.renderDoc("data-access", "mybatis-basic", "MyBatis基础集成")
}

// MyBatis高级特性
func (c *DocsController) GetDataAccessMybatisAdvanced() {
	c.renderDoc("data-access", "mybatis-advanced", "MyBatis高级特性")
}

// MyBatis性能优化
func (c *DocsController) GetDataAccessMybatisPerformance() {
	c.renderDoc("data-access", "mybatis-performance", "MyBatis性能优化")
}

// 保留旧的MyBatis路由以向后兼容
func (c *DocsController) GetDataAccessMybatis() {
	// 重定向到基础集成页面
	c.renderDoc("data-access", "mybatis-basic", "MyBatis基础集成")
}

func (c *DocsController) GetDataAccessDatabaseConfig() {
	c.renderDoc("data-access", "database-config", "数据库配置")
}

func (c *DocsController) GetDataAccessTransaction() {
	c.renderDoc("data-access", "transaction", "事务管理")
}

// 新增的数据库调优文档
func (c *DocsController) GetDataAccessDatabaseTuning() {
	c.renderDoc("data-access", "database-tuning", "数据库调优")
}

// 新增的缓存策略文档
func (c *DocsController) GetDataAccessCachingStrategies() {
	c.renderDoc("data-access", "caching-strategies", "缓存策略")
}

// 新增的监控告警文档
func (c *DocsController) GetDataAccessMonitoringAlerting() {
	c.renderDoc("data-access", "monitoring-alerting", "监控告警")
}

// ============= 视图渲染 =============
func (c *DocsController) GetViewTemplateOverview() {
	c.renderDoc("view-template", "overview", "视图概览")
}

func (c *DocsController) GetViewTemplateTemplateEngine() {
	c.renderDoc("view-template", "template-engine", "模板引擎")
}

func (c *DocsController) GetViewTemplateViewRendering() {
	c.renderDoc("view-template", "view-rendering", "视图渲染")
}

func (c *DocsController) GetViewTemplateStaticAssets() {
	c.renderDoc("view-template", "static-assets", "静态资源")
}

// ============= 配置管理 =============
func (c *DocsController) GetConfigurationAppConfig() {
	c.renderDoc("configuration", "app-config", "应用配置")
}

func (c *DocsController) GetConfigurationEnvironment() {
	c.renderDoc("configuration", "environment", "环境配置")
}

func (c *DocsController) GetConfigurationLogging() {
	c.renderDoc("configuration", "logging", "日志配置")
}

// ============= 高级功能 =============
func (c *DocsController) GetAdvancedSession() {
	c.renderDoc("advanced", "session", "会话管理")
}

func (c *DocsController) GetAdvancedCache() {
	c.renderDoc("advanced", "cache", "缓存系统")
}

func (c *DocsController) GetAdvancedValidation() {
	c.renderDoc("advanced", "validation", "验证系统")
}

func (c *DocsController) GetAdvancedCaptcha() {
	c.renderDoc("advanced", "captcha", "验证码功能")
}

func (c *DocsController) GetAdvancedScheduler() {
	c.renderDoc("advanced", "scheduler", "任务调度")
}

// ============= 部署运维 =============
func (c *DocsController) GetDeploymentOverview() {
	c.renderDoc("deployment", "deployment-overview", "部署概览", true)
}

func (c *DocsController) GetDeploymentDocker() {
	c.renderDoc("deployment", "docker", "Docker部署")
}

func (c *DocsController) GetDeploymentKubernetes() {
	c.renderDoc("deployment", "kubernetes", "K8s部署")
}

func (c *DocsController) GetDeploymentMonitoring() {
	c.renderDoc("deployment", "monitoring", "监控告警")
}

// ============= 开发工具 =============
func (c *DocsController) GetDevToolsCodegen() {
	c.renderDoc("dev-tools", "codegen", "代码生成")
}

func (c *DocsController) GetDevToolsHotReload() {
	c.renderDoc("dev-tools", "hot-reload", "热重载")
}

func (c *DocsController) GetDevToolsPerformance() {
	c.renderDoc("dev-tools", "performance", "性能监控")
}

func (c *DocsController) GetDevToolsTesting() {
	c.renderDoc("dev-tools", "testing", "测试工具")
}
