package controllers

import (
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
)

// ExampleController 展示各种模板功能的控制器
type ExampleController struct {
	mvc.BaseController
}

// Index 首页，显示所有功能链接
func (c *ExampleController) Index() {
	c.SetData("Title", "YYHertz 模板引擎演示")
	c.SetData("Description", "展示YYHertz Beego风格模板引擎的强大功能")

	// index.html是完整的HTML文档，不需要layout包装
	c.Layout = ""

	// 功能链接列表
	features := []map[string]any{
		{
			"title":       "Layout 继承演示",
			"description": "展示模板布局继承系统的强大功能",
			"url":         "/layout",
			"icon":        "🎨",
		},
		{
			"title":       "Beego 函数演示",
			"description": "体验丰富的内置模板函数库",
			"url":         "/beego-functions",
			"icon":        "⚡",
		},
		{
			"title":       "高级功能演示",
			"description": "CSRF安全、自动推导、命名约定等高级特性",
			"url":         "/advanced",
			"icon":        "🚀",
		},
		{
			"title":       "性能监控",
			"description": "查看模板引擎性能统计和监控信息",
			"url":         "/performance",
			"icon":        "📊",
		},
		{
			"title":       "模板管理",
			"description": "管理和预览各种模板文件",
			"url":         "/templates",
			"icon":        "📝",
		},
		{
			"title":       "管理后台",
			"description": "完整的管理后台界面演示",
			"url":         "/admin",
			"icon":        "🛠️",
		},
	}

	c.SetData("Features", features)
	c.SetData("Now", time.Now())
	c.TplName = "index.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString(err.Error())
	}
}

// LayoutDemo Layout继承演示
func (c *ExampleController) LayoutDemo() {
	c.SetData("Title", "Layout 继承演示")
	c.SetData("Heading", "Layout继承系统演示")
	c.SetData("Content", "这个内容通过layout模板进行包装，展示了强大的布局继承功能。")

	// 用户信息
	c.SetData("User", map[string]any{
		"Name":  "张三",
		"Email": "zhangsan@example.com",
		"Role":  "管理员",
	})

	// 导航信息
	navigation := []map[string]string{
		{"name": "首页", "url": "/"},
		{"name": "Layout演示", "url": "/layout"},
		{"name": "函数演示", "url": "/beego-functions"},
		{"name": "高级功能", "url": "/advanced"},
	}
	c.SetData("Navigation", navigation)
	c.SetData("Now", time.Now())

	c.Layout = ""
	c.TplName = "example/layoutx.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString(err.Error())
	}
}

// BeegoFunctions Beego模板函数演示
func (c *ExampleController) BeegoFunctions() {
	c.SetData("Title", "Beego 函数演示")
	c.SetData("Now", time.Now())
	c.SetData("Past", time.Now().Add(-2*time.Hour))

	// 演示数据
	c.SetData("Price", 99.99)
	c.SetData("Text", "This is a very long text that needs to be truncated for display")
	c.SetData("HTML", "<script>alert('XSS')</script>")

	// 商品列表
	items := []map[string]any{
		{"name": "笔记本电脑", "price": 5999},
		{"name": "无线鼠标", "price": 89},
		{"name": "机械键盘", "price": 299},
	}
	c.SetData("Items", items)

	c.Layout = ""
	c.TplName = "example/beego-functions.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString(err.Error())
	}
}
