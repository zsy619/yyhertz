package controllers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/mvc"
)

// BlogsController 博客控制器
type BlogsController struct {
	mvc.BaseController
}

// Get 处理博客列表页面
func (this *BlogsController) Get(ctx context.Context, c *app.RequestContext) {
	// 设置布局模板
	this.Layout = "layouts/layout_blog.html"
	// 设置内容模板
	this.TplName = "blogs/index.html"

	// 初始化 LayoutSections
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["HtmlHead"] = "blogs/html_head.html"
	this.LayoutSections["Scripts"] = "blogs/scripts.html"
	this.LayoutSections["Sidebar"] = ""

	// 设置页面数据
	this.Data = map[string]any{
		"Title":       "技术博客",
		"PageTitle":   "最新博客文章",
		"Description": "分享技术心得与经验",
		"Articles": []map[string]any{
			{
				"Title":   "YYHertz 框架入门指南",
				"Summary": "详细介绍 YYHertz 框架的核心功能和使用方法",
				"Author":  "开发团队",
				"Date":    "2024-08-28",
				"Tags":    []string{"Go", "框架", "教程"},
			},
			{
				"Title":   "LayoutSections 高级用法",
				"Summary": "深入探讨布局区块系统的高级特性和最佳实践",
				"Author":  "架构师",
				"Date":    "2024-08-27",
				"Tags":    []string{"模板", "布局", "进阶"},
			},
			{
				"Title":   "性能优化实战经验",
				"Summary": "分享在实际项目中的性能优化技巧和解决方案",
				"Author":  "性能专家",
				"Date":    "2024-08-26",
				"Tags":    []string{"性能", "优化", "实战"},
			},
		},
	}

	// 渲染页面
	if err := this.Render(); err != nil {
		this.Ctx.WriteString(err.Error())
	}
}
