package controllers

import (
	"github.com/zsy619/yyhertz/framework/mvc"
)

type HomeController struct {
	mvc.BaseController
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
		"Github":    "https://github.com/cloudwego/hertz",
		"Docs":      "https://www.cloudwego.io/zh/docs/hertz/",
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
	c.SetData("Title", "快速开始")
	c.SetData("CurrentDoc", "quickstart")
	c.RenderHTML("home/docs/quickstart.html")
}

// 控制器文档
func (c *HomeController) GetController() {
	c.SetData("Title", "控制器")
	c.SetData("CurrentDoc", "controller")
	c.RenderHTML("home/docs/controller.html")
}

// 路由文档
func (c *HomeController) GetRouting() {
	c.SetData("Title", "路由")
	c.SetData("CurrentDoc", "routing")
	c.RenderHTML("home/docs/routing.html")
}

// 中间件文档
func (c *HomeController) GetMiddleware() {
	c.SetData("Title", "中间件")
	c.SetData("CurrentDoc", "middleware")
	c.RenderHTML("home/docs/middleware.html")
}

// 模板文档
func (c *HomeController) GetTemplate() {
	c.SetData("Title", "模板")
	c.SetData("CurrentDoc", "template")
	c.RenderHTML("home/docs/template.html")
}

// 数据库集成文档
func (c *HomeController) GetDatabase() {
	c.SetData("Title", "数据库集成")
	c.SetData("CurrentDoc", "database")
	c.RenderHTML("home/docs/database.html")
}

// 部署文档
func (c *HomeController) GetDeployment() {
	c.SetData("Title", "部署上线")
	c.SetData("CurrentDoc", "deployment")
	c.RenderHTML("home/docs/deployment.html")
}
