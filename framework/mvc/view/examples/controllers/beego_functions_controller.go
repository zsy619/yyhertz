package controllers

import (
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
)

// BeeGoFunctionsController 专门用于测试Beego风格模板函数的控制器
type BeeGoFunctionsController struct {
	mvc.BaseController
}

// ============= 单独函数测试 =============

// TestInclude 测试Include模板函数
func (c *BeeGoFunctionsController) TestInclude() {
	c.SetData("Title", "Include 函数测试")
	c.SetData("Message", "这是通过Include函数包含的内容测试")
	c.SetData("TestData", map[string]any{
		"name":        "测试用户",
		"description": "Include函数可以包含其他模板文件并传递数据",
		"timestamp":   time.Now(),
	})

	// 准备包含模板需要的数据
	c.SetData("HeaderTitle", "Include测试页面")
	c.SetData("Navigation", []map[string]string{
		{"name": "首页", "url": "/test"},
		{"name": "Include测试", "url": "/test/include"},
	})

	c.Layout = ""
	c.TplName = "test/bf_include.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestTemplateInclude 测试TemplateInclude模板函数
func (c *BeeGoFunctionsController) TestTemplateInclude() {
	c.SetData("Title", "TemplateInclude 函数测试")
	c.SetData("Message", "这是通过TemplateInclude函数包含的内容测试")
	c.SetData("TestData", map[string]any{
		"name":        "测试用户",
		"description": "TemplateInclude函数是Include的别名，功能相同",
		"timestamp":   time.Now(),
	})

	// 准备包含模板需要的数据
	c.SetData("HeaderTitle", "TemplateInclude测试页面")
	c.SetData("UserInfo", map[string]any{
		"username": "admin",
		"role":     "管理员",
		"avatar":   "/static/images/avatar.png",
	})

	c.Layout = ""
	c.TplName = "test/bf_templateinclude.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestPartial 测试Partial模板函数
func (c *BeeGoFunctionsController) TestPartial() {
	c.SetData("Title", "Partial 函数测试")
	c.SetData("Message", "这是通过Partial函数包含的部分模板测试")
	c.SetData("TestData", map[string]any{
		"name":        "测试用户",
		"description": "Partial函数用于包含部分模板，常用于页面片段",
		"timestamp":   time.Now(),
	})

	// 准备侧边栏数据
	c.SetData("SidebarItems", []map[string]any{
		{"title": "仪表盘", "icon": "dashboard", "url": "/dashboard"},
		{"title": "用户管理", "icon": "users", "url": "/users"},
		{"title": "设置", "icon": "settings", "url": "/settings"},
	})

	// 准备表格数据
	c.SetData("TableData", []map[string]any{
		{"id": 1, "name": "张三", "email": "zhangsan@example.com", "status": "active"},
		{"id": 2, "name": "李四", "email": "lisi@example.com", "status": "inactive"},
		{"id": 3, "name": "王五", "email": "wangwu@example.com", "status": "active"},
	})

	c.Layout = ""
	c.TplName = "test/bf_partial.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestComponentTemplate 测试ComponentTemplate模板函数
func (c *BeeGoFunctionsController) TestComponentTemplate() {
	c.SetData("Title", "ComponentTemplate 函数测试")
	c.SetData("Message", "这是通过ComponentTemplate函数渲染的组件测试")
	c.SetData("TestData", map[string]any{
		"name":        "测试用户",
		"description": "ComponentTemplate函数用于渲染可复用的组件",
		"timestamp":   time.Now(),
	})

	// 准备组件数据
	c.SetData("ComponentData", map[string]any{
		"cardTitle":   "用户信息卡片",
		"cardContent": "这是一个通过ComponentTemplate渲染的卡片组件",
		"buttonText":  "了解更多",
		"buttonUrl":   "#",
	})

	// 准备表单组件数据
	c.SetData("FormData", map[string]any{
		"action": "/submit",
		"method": "post",
		"fields": []map[string]any{
			{"name": "username", "type": "text", "label": "用户名", "required": true},
			{"name": "email", "type": "email", "label": "邮箱", "required": true},
			{"name": "message", "type": "textarea", "label": "留言"},
		},
	})

	c.Layout = ""
	c.TplName = "test/bf_componenttemplate.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestRenderTemplate 测试RenderTemplate模板函数
func (c *BeeGoFunctionsController) TestRenderTemplate() {
	c.SetData("Title", "RenderTemplate 函数测试")
	c.SetData("Message", "这是通过RenderTemplate函数渲染的模板测试")
	c.SetData("TestData", map[string]any{
		"name":        "测试用户",
		"description": "RenderTemplate函数是Include的另一个别名，功能相同",
		"timestamp":   time.Now(),
	})

	// 准备页面数据
	c.SetData("PageConfig", map[string]any{
		"showHeader": true,
		"showFooter": true,
		"theme":      "default",
		"language":   "zh-CN",
	})

	// 准备内容数据
	c.SetData("ContentData", map[string]any{
		"articles": []map[string]any{
			{
				"title":   "模板函数使用指南",
				"summary": "详细介绍各种模板函数的使用方法",
				"author":  "开发团队",
				"date":    time.Now().Format("2006-01-02"),
			},
			{
				"title":   "最佳实践分享",
				"summary": "分享模板开发的最佳实践经验",
				"author":  "技术专家",
				"date":    time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			},
		},
	})

	c.Layout = ""
	c.TplName = "test/bf_rendertemplate.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestTemplate 测试Template模板函数
func (c *BeeGoFunctionsController) TestTemplate() {
	c.SetData("Title", "Template 函数测试")
	c.SetData("Message", "这是通过Template函数包含的内容测试")
	c.SetData("TestData", map[string]any{
		"name":        "Template测试用户",
		"description": "Template函数是TemplateInclude的简化别名，提供更简洁的模板包含语法",
		"timestamp":   time.Now(),
	})

	// 准备模板渲染数据
	c.SetData("TemplateData", map[string]any{
		"renderType": "template",
		"features": []string{
			"简洁的函数名称",
			"与TemplateInclude功能完全相同", 
			"更直观的语义表达",
			"减少代码冗余",
		},
		"usageExamples": []map[string]any{
			{
				"scenario": "基础模板包含",
				"syntax":   "{{template \"partial.html\" .data}}",
				"description": "最常用的模板包含方式",
			},
			{
				"scenario": "条件模板包含",
				"syntax":   "{{if .showHeader}}{{template \"header.html\" .}}{{end}}",
				"description": "根据条件动态包含模板",
			},
			{
				"scenario": "循环模板包含",
				"syntax":   "{{range .items}}{{template \"item.html\" .}}{{end}}",
				"description": "在循环中包含模板片段",
			},
		},
	})

	// 准备对比数据
	c.SetData("ComparisonData", map[string]any{
		"template": map[string]any{
			"name": "template",
			"length": 8,
			"readability": "excellent",
			"commonality": "high",
		},
		"templateinclude": map[string]any{
			"name": "templateinclude", 
			"length": 15,
			"readability": "good",
			"commonality": "medium",
		},
		"include": map[string]any{
			"name": "include",
			"length": 7,
			"readability": "excellent",
			"commonality": "very high",
		},
	})

	c.Layout = ""
	c.TplName = "test/bf_template.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// ============= 组合函数测试 =============

// TestIncludeTemplateInclude 测试Include + TemplateInclude组合
func (c *BeeGoFunctionsController) TestIncludeTemplateInclude() {
	c.SetData("Title", "Include + TemplateInclude 组合测试")
	c.SetData("Message", "测试Include和TemplateInclude函数的组合使用")
	c.SetData("TestData", map[string]any{
		"name":        "组合测试",
		"description": "演示多个模板函数的组合使用场景",
		"timestamp":   time.Now(),
	})

	// 准备头部数据
	c.SetData("Header", map[string]any{
		"title":      "组合测试页面",
		"breadcrumb": []string{"首页", "测试", "组合测试"},
	})

	// 准备主要内容数据
	c.SetData("MainContent", map[string]any{
		"sections": []map[string]any{
			{
				"title":   "Include部分",
				"content": "这部分内容通过Include函数加载",
				"type":    "include",
			},
			{
				"title":   "TemplateInclude部分",
				"content": "这部分内容通过TemplateInclude函数加载",
				"type":    "templateinclude",
			},
		},
	})

	c.Layout = ""
	c.TplName = "test/combinations/bf_include_templateinclude.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestPartialComponent 测试Partial + ComponentTemplate组合
func (c *BeeGoFunctionsController) TestPartialComponent() {
	c.SetData("Title", "Partial + ComponentTemplate 组合测试")
	c.SetData("Message", "测试Partial和ComponentTemplate函数的组合使用")
	c.SetData("TestData", map[string]any{
		"name":        "组合测试",
		"description": "演示部分模板和组件的组合使用",
		"timestamp":   time.Now(),
	})

	// 准备侧边栏数据（用于Partial）
	c.SetData("Sidebar", map[string]any{
		"title": "功能菜单",
		"items": []map[string]any{
			{"name": "用户管理", "url": "/users", "icon": "user"},
			{"name": "角色管理", "url": "/roles", "icon": "shield"},
			{"name": "权限管理", "url": "/permissions", "icon": "key"},
		},
	})

	// 准备组件数据（用于ComponentTemplate）
	c.SetData("Cards", []map[string]any{
		{
			"title":       "统计卡片1",
			"value":       "1,234",
			"description": "总用户数",
			"icon":        "users",
			"color":       "blue",
		},
		{
			"title":       "统计卡片2",
			"value":       "89.5%",
			"description": "系统利用率",
			"icon":        "chart",
			"color":       "green",
		},
		{
			"title":       "统计卡片3",
			"value":       "156",
			"description": "在线用户",
			"icon":        "activity",
			"color":       "orange",
		},
	})

	c.Layout = ""
	c.TplName = "test/combinations/bf_partial_component.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestAllFunctions 测试所有模板函数的综合使用
func (c *BeeGoFunctionsController) TestAllFunctions() {
	c.SetData("Title", "所有模板函数综合测试")
	c.SetData("Message", "展示所有模板函数在一个页面中的协同工作")
	c.SetData("TestData", map[string]any{
		"name":        "综合测试",
		"description": "演示Include、TemplateInclude、Partial、ComponentTemplate、RenderTemplate的综合使用",
		"timestamp":   time.Now(),
	})

	// 全局页面配置
	c.SetData("PageConfig", map[string]any{
		"title":       "模板函数综合演示",
		"description": "全面展示各种模板函数的强大功能",
		"version":     "1.0.0",
	})

	// 头部导航数据
	c.SetData("Navigation", []map[string]any{
		{"name": "首页", "url": "/", "active": false},
		{"name": "测试中心", "url": "/test", "active": true},
		{"name": "文档", "url": "/docs", "active": false},
	})

	// 侧边栏数据
	c.SetData("SidebarMenu", []map[string]any{
		{"title": "单独测试", "items": []map[string]any{
			{"name": "Include", "url": "/test/include"},
			{"name": "TemplateInclude", "url": "/test/templateinclude"},
			{"name": "Partial", "url": "/test/partial"},
		}},
		{"title": "组合测试", "items": []map[string]any{
			{"name": "Include+TemplateInclude", "url": "/test/include-templateinclude"},
			{"name": "Partial+Component", "url": "/test/partial-component"},
		}},
	})

	// 主要内容区域数据
	c.SetData("MainSections", []map[string]any{
		{
			"type":    "component",
			"name":    "welcome-card",
			"title":   "欢迎使用",
			"content": "这是通过ComponentTemplate渲染的欢迎卡片",
		},
		{
			"type":    "partial",
			"name":    "statistics",
			"title":   "统计信息",
			"content": "这部分通过Partial包含统计信息",
		},
		{
			"type":    "include",
			"name":    "recent-activity",
			"title":   "最近活动",
			"content": "这部分通过Include包含最近活动",
		},
	})

	// 页脚数据
	c.SetData("FooterData", map[string]any{
		"copyright": "© 2024 YYHertz Framework. All rights reserved.",
		"links": []map[string]string{
			{"name": "关于我们", "url": "/about"},
			{"name": "联系我们", "url": "/contact"},
			{"name": "帮助中心", "url": "/help"},
		},
	})

	c.Layout = ""
	c.TplName = "test/combinations/bf_all_functions.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestNestedIncludes 测试嵌套包含
func (c *BeeGoFunctionsController) TestNestedIncludes() {
	c.SetData("Title", "嵌套包含测试")
	c.SetData("Message", "测试模板函数的嵌套调用能力")
	c.SetData("TestData", map[string]any{
		"name":        "嵌套测试",
		"description": "演示模板可以包含其他模板，而被包含的模板还可以再包含更多模板",
		"timestamp":   time.Now(),
		"level":       1,
	})

	// 多层嵌套数据
	c.SetData("NestedData", map[string]any{
		"level1": map[string]any{
			"title": "第一层",
			"level2": map[string]any{
				"title": "第二层",
				"level3": map[string]any{
					"title": "第三层",
					"content": "这是最深层的内容",
				},
			},
		},
	})

	c.Layout = ""
	c.TplName = "test/nested/bf_nested_includes.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestErrorHandling 测试错误处理
func (c *BeeGoFunctionsController) TestErrorHandling() {
	c.SetData("Title", "错误处理测试")
	c.SetData("Message", "测试模板函数的错误处理机制")
	c.SetData("TestData", map[string]any{
		"name":        "错误处理测试",
		"description": "验证当包含不存在的模板或传递错误数据时的处理情况",
		"timestamp":   time.Now(),
	})

	// 错误场景数据
	c.SetData("ErrorScenarios", []map[string]any{
		{
			"name":        "不存在的模板",
			"template":    "non_existent_template.html",
			"description": "测试包含不存在的模板文件",
		},
		{
			"name":        "空数据",
			"template":    "test_template.html",
			"data":        nil,
			"description": "测试传递空数据到模板",
		},
		{
			"name":        "错误数据类型",
			"template":    "test_template.html",
			"data":        "invalid_data",
			"description": "测试传递错误类型的数据",
		},
	})

	c.Layout = ""
	c.TplName = "test/error/bf_error_handling.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// ============= 辅助方法 =============

// Index 测试首页，显示所有测试链接
func (c *BeeGoFunctionsController) Index() {
	c.SetData("Title", "Beego模板函数测试中心")
	c.SetData("Description", "全面测试Beego风格模板函数的功能和性能")

	// 单独函数测试链接
	singleTests := []map[string]any{
		{
			"name":        "Include 测试",
			"url":         "/test/beego-functions/include",
			"description": "测试Include模板包含函数",
			"icon":        "📄",
		},
		{
			"name":        "TemplateInclude 测试",
			"url":         "/test/beego-functions/templateinclude",
			"description": "测试TemplateInclude模板包含函数",
			"icon":        "📋",
		},
		{
			"name":        "Partial 测试",
			"url":         "/test/beego-functions/partial",
			"description": "测试Partial部分模板函数",
			"icon":        "📃",
		},
		{
			"name":        "ComponentTemplate 测试",
			"url":         "/test/beego-functions/componenttemplate",
			"description": "测试ComponentTemplate组件模板函数",
			"icon":        "🧩",
		},
		{
			"name":        "RenderTemplate 测试",
			"url":         "/test/beego-functions/rendertemplate",
			"description": "测试RenderTemplate渲染模板函数",
			"icon":        "🎨",
		},
		{
			"name":        "Template 测试",
			"url":         "/test/beego-functions/template",
			"description": "测试Template模板函数（TemplateInclude简化别名）",
			"icon":        "📝",
		},
	}

	// Layout版本测试链接
	layoutTests := []map[string]any{
		{
			"name":        "Include + Layout",
			"url":         "/test/beego-functions/include_layout",
			"description": "使用Layout布局测试Include模板函数",
			"icon":        "📄",
		},
		{
			"name":        "TemplateInclude + Layout",
			"url":         "/test/beego-functions/templateinclude_layout",
			"description": "使用Layout布局测试TemplateInclude模板函数",
			"icon":        "📋",
		},
		{
			"name":        "Partial + Layout",
			"url":         "/test/beego-functions/partial_layout",
			"description": "使用Layout布局测试Partial部分模板函数",
			"icon":        "📃",
		},
		{
			"name":        "ComponentTemplate + Layout",
			"url":         "/test/beego-functions/componenttemplate_layout",
			"description": "使用Layout布局测试ComponentTemplate组件模板函数",
			"icon":        "🧩",
		},
		{
			"name":        "RenderTemplate + Layout",
			"url":         "/test/beego-functions/rendertemplate_layout",
			"description": "使用Layout布局测试RenderTemplate渲染模板函数",
			"icon":        "🎨",
		},
		{
			"name":        "Template + Layout",
			"url":         "/test/beego-functions/template_layout",
			"description": "使用Layout布局测试Template模板函数",
			"icon":        "📝",
		},
	}

	// 组合测试链接
	combinationTests := []map[string]any{
		{
			"name":        "Include + TemplateInclude",
			"url":         "/test/beego-functions/include-templateinclude",
			"description": "测试Include和TemplateInclude的组合使用",
			"icon":        "🔗",
		},
		{
			"name":        "Partial + ComponentTemplate",
			"url":         "/test/beego-functions/partial-component",
			"description": "测试Partial和ComponentTemplate的组合使用",
			"icon":        "⚙️",
		},
		{
			"name":        "所有函数综合测试",
			"url":         "/test/beego-functions/all-functions",
			"description": "测试所有模板函数的综合使用场景",
			"icon":        "🚀",
		},
		{
			"name":        "嵌套包含测试",
			"url":         "/test/beego-functions/nested-includes",
			"description": "测试模板函数的嵌套调用能力",
			"icon":        "🎯",
		},
		{
			"name":        "错误处理测试",
			"url":         "/test/beego-functions/error-handling",
			"description": "测试模板函数的错误处理机制",
			"icon":        "⚠️",
		},
	}

	// Layout版本组合测试链接
	layoutCombinationTests := []map[string]any{
		{
			"name":        "Include + TemplateInclude (Layout版)",
			"url":         "/test/beego-functions/includetemplate_layout",
			"description": "使用Layout布局测试Include和TemplateInclude的组合使用",
			"icon":        "🔗",
		},
		{
			"name":        "Partial + ComponentTemplate (Layout版)",
			"url":         "/test/beego-functions/partialcomponent_layout",
			"description": "使用Layout布局测试Partial和ComponentTemplate的组合使用",
			"icon":        "⚙️",
		},
		{
			"name":        "所有函数综合测试 (Layout版)",
			"url":         "/test/beego-functions/allfunctions_layout",
			"description": "使用Layout布局测试所有模板函数的综合使用场景",
			"icon":        "🚀",
		},
		{
			"name":        "嵌套包含测试 (Layout版)",
			"url":         "/test/beego-functions/nestedincludes_layout",
			"description": "使用Layout布局测试模板函数的嵌套调用能力",
			"icon":        "🎯",
		},
		{
			"name":        "错误处理测试 (Layout版)",
			"url":         "/test/beego-functions/errorhandling_layout",
			"description": "使用Layout布局测试模板函数的错误处理机制",
			"icon":        "⚠️",
		},
	}

	c.SetData("SingleTests", singleTests)
	c.SetData("LayoutTests", layoutTests)
	c.SetData("CombinationTests", combinationTests)
	c.SetData("LayoutCombinationTests", layoutCombinationTests)
	c.SetData("Now", time.Now())

	c.Layout = ""
	c.TplName = "test/bf_index.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// ============= Layout版本测试方法 =============

// TestIncludeLayout 测试Include函数 + Layout布局
func (c *BeeGoFunctionsController) TestIncludeLayout() {
	c.SetData("Title", "Include 函数测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试Include模板包含函数")
	c.SetData("TestType", "Include + Layout")
	c.SetData("Message", "这是通过Include函数包含的内容测试，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "Layout测试用户",
		"description": "Include函数可以包含其他模板文件并传递数据，配合Layout使用",
		"timestamp":   time.Now(),
	})

	// 准备包含模板需要的数据
	c.SetData("HeaderTitle", "Include测试页面 (Layout版)")
	c.SetData("Navigation", []map[string]string{
		{"name": "首页", "url": "/test/beego-functions"},
		{"name": "Include Layout", "url": "/test/beego-functions/include_layout"},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>🔧 Include函数特性</h3>
		<ul style="list-style: none; padding: 0;">
			<li style="padding: 8px 0; border-bottom: 1px solid #eee;">
				📝 动态包含模板文件
			</li>
			<li style="padding: 8px 0; border-bottom: 1px solid #eee;">
				📊 支持数据传递
			</li>
			<li style="padding: 8px 0; border-bottom: 1px solid #eee;">
				🔄 支持嵌套包含
			</li>
			<li style="padding: 8px 0;">
				⚡ 高性能渲染
			</li>
		</ul>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_include_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestTemplateIncludeLayout 测试TemplateInclude函数 + Layout布局
func (c *BeeGoFunctionsController) TestTemplateIncludeLayout() {
	c.SetData("Title", "TemplateInclude 函数测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试TemplateInclude模板包含函数")
	c.SetData("TestType", "TemplateInclude + Layout")
	c.SetData("Message", "这是通过TemplateInclude函数包含的内容测试，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "TemplateInclude Layout测试",
		"description": "TemplateInclude是Include的别名，功能完全相同，配合Layout使用",
		"timestamp":   time.Now(),
	})

	// 用户信息数据
	c.SetData("UserInfo", map[string]any{
		"username": "Layout测试用户",
		"role":     "模板测试工程师",
		"avatar":   "layout_avatar.png",
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>📋 TemplateInclude特性</h3>
		<div style="background: #f8f9fa; padding: 15px; border-radius: 8px; margin: 10px 0;">
			<p style="margin: 5px 0;"><strong>别名关系:</strong> Include = TemplateInclude</p>
			<p style="margin: 5px 0;"><strong>功能一致性:</strong> 100%</p>
			<p style="margin: 5px 0;"><strong>性能表现:</strong> 相同</p>
			<p style="margin: 5px 0;"><strong>使用场景:</strong> 完全互换</p>
		</div>
	`)

	// 添加安全演示数据（用于模板中的安全示例演示）
	c.SetData("UserPath", "test/partials/user_profile.html")         // 安全的预定义路径
	c.SetData("UserType", "user")                                     // 用于条件判断
	c.SetData("TrustedHTML", "<strong>这是安全的HTML内容</strong>")       // 用于unescaped演示
	c.SetData("UserDefinedPath", "test/partials/user_profile.html")   // 用于安全路径演示，避免动态模板路径注入

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_templateinclude_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestPartialLayout 测试Partial函数 + Layout布局
func (c *BeeGoFunctionsController) TestPartialLayout() {
	c.SetData("Title", "Partial 函数测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试Partial部分模板函数")
	c.SetData("TestType", "Partial + Layout")
	c.SetData("Message", "这是通过Partial函数包含的内容测试，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "Partial Layout测试",
		"description": "Partial函数用于包含可重用的部分模板，配合Layout系统使用",
		"timestamp":   time.Now(),
	})

	// 部分模板数据
	c.SetData("PartialData", map[string]any{
		"sections": []map[string]any{
			{
				"title":   "用户管理模块",
				"content": "这是通过Partial包含的用户管理部分",
				"icon":    "👥",
			},
			{
				"title":   "系统设置模块", 
				"content": "这是通过Partial包含的系统设置部分",
				"icon":    "⚙️",
			},
			{
				"title":   "数据统计模块",
				"content": "这是通过Partial包含的数据统计部分",
				"icon":    "📊",
			},
		},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>🧩 Partial函数优势</h3>
		<div style="background: linear-gradient(135deg, #ffeaa7, #fab1a0); padding: 15px; border-radius: 10px; color: #2d3436;">
			<h4 style="margin: 0 0 10px 0;">✨ 主要特点</h4>
			<ul style="margin: 0; padding-left: 20px;">
				<li>模块化开发</li>
				<li>代码重用性高</li>
				<li>维护成本低</li>
				<li>结构更清晰</li>
			</ul>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_partial_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestComponentTemplateLayout 测试ComponentTemplate函数 + Layout布局
func (c *BeeGoFunctionsController) TestComponentTemplateLayout() {
	c.SetData("Title", "ComponentTemplate 函数测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试ComponentTemplate组件模板函数")
	c.SetData("TestType", "ComponentTemplate + Layout")
	c.SetData("Message", "这是通过ComponentTemplate函数渲染的组件测试，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "ComponentTemplate Layout测试",
		"description": "ComponentTemplate专门用于渲染可重用组件，配合Layout系统",
		"timestamp":   time.Now(),
	})

	// 组件数据
	c.SetData("ComponentsData", map[string]any{
		"cards": []map[string]any{
			{
				"type":        "info",
				"title":       "信息卡片",
				"content":     "这是一个信息展示组件",
				"color":       "#3498db",
				"icon":        "ℹ️",
			},
			{
				"type":        "success", 
				"title":       "成功卡片",
				"content":     "这是一个成功状态组件",
				"color":       "#2ecc71",
				"icon":        "✅",
			},
			{
				"type":        "warning",
				"title":       "警告卡片", 
				"content":     "这是一个警告提示组件",
				"color":       "#f39c12",
				"icon":        "⚠️",
			},
		},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>⚙️ ComponentTemplate</h3>
		<div style="background: linear-gradient(135deg, #a8edea, #fed6e3); padding: 15px; border-radius: 10px;">
			<h4 style="margin: 0 0 10px 0; color: #2d3436;">🎨 组件化开发</h4>
			<div style="background: rgba(255,255,255,0.7); padding: 10px; border-radius: 5px; margin: 5px 0;">
				<strong>可重用性:</strong> 极高
			</div>
			<div style="background: rgba(255,255,255,0.7); padding: 10px; border-radius: 5px; margin: 5px 0;">
				<strong>维护性:</strong> 优秀
			</div>
			<div style="background: rgba(255,255,255,0.7); padding: 10px; border-radius: 5px;">
				<strong>扩展性:</strong> 灵活
			</div>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_componenttemplate_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestRenderTemplateLayout 测试RenderTemplate函数 + Layout布局
func (c *BeeGoFunctionsController) TestRenderTemplateLayout() {
	c.SetData("Title", "RenderTemplate 函数测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试RenderTemplate渲染模板函数")
	c.SetData("TestType", "RenderTemplate + Layout")
	c.SetData("Message", "这是通过RenderTemplate函数渲染的内容测试，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "RenderTemplate Layout测试",
		"description": "RenderTemplate提供完整的模板渲染功能，配合Layout系统使用",
		"timestamp":   time.Now(),
	})

	// 渲染内容数据
	c.SetData("RenderData", map[string]any{
		"templates": []map[string]any{
			{
				"name":        "用户列表模板",
				"description": "动态渲染用户列表数据",
				"status":      "active",
				"users":       []string{"张三", "李四", "王五", "赵六"},
			},
			{
				"name":        "统计图表模板",
				"description": "渲染各种数据统计图表",
				"status":      "active", 
				"charts":      []string{"柱状图", "饼图", "折线图", "散点图"},
			},
			{
				"name":        "表单模板",
				"description": "动态生成各种表单界面",
				"status":      "pending",
				"forms":       []string{"登录表单", "注册表单", "搜索表单", "设置表单"},
			},
		},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>🎨 RenderTemplate</h3>
		<div style="background: linear-gradient(135deg, #667eea, #764ba2); color: white; padding: 15px; border-radius: 10px;">
			<h4 style="margin: 0 0 15px 0;">🚀 核心功能</h4>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				完整渲染控制
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				灵活数据传递
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				高级模板功能
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px;">
				性能优化支持
			</div>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_rendertemplate_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestTemplateLayout 测试Template函数 + Layout布局
func (c *BeeGoFunctionsController) TestTemplateLayout() {
	c.SetData("Title", "Template 函数测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试Template模板函数")
	c.SetData("TestType", "Template + Layout")
	c.SetData("Message", "这是通过Template函数包含的内容测试，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "Template Layout测试",
		"description": "Template函数是TemplateInclude的简化别名，在Layout环境中提供更简洁的模板包含语法",
		"timestamp":   time.Now(),
	})

	// Layout环境下的Template特性数据
	c.SetData("LayoutTemplateData", map[string]any{
		"renderType": "template-layout",
		"features": []string{
			"Layout环境完美集成",
			"简洁的函数名称",
			"与TemplateInclude功能完全相同",
			"更直观的语义表达",
			"Layout级别缓存支持",
		},
		"layoutAdvantages": []map[string]any{
			{
				"feature": "统一布局管理",
				"description": "Template函数在Layout框架中提供一致的包含体验",
				"benefit": "提升开发效率和维护性",
			},
			{
				"feature": "响应式设计支持",
				"description": "Layout环境下Template函数支持响应式模板包含",
				"benefit": "适配多种设备和屏幕尺寸",
			},
			{
				"feature": "嵌套包含优化",
				"description": "Layout中的Template调用经过性能优化",
				"benefit": "更快的模板渲染速度",
			},
		},
	})

	// Template函数语法示例
	c.SetData("SyntaxExamples", []map[string]any{
		{
			"category": "基础语法",
			"examples": []map[string]string{
				{"syntax": "{{template \"header.html\" .}}", "description": "包含头部模板"},
				{"syntax": "{{template \"sidebar.html\" .sidebarData}}", "description": "包含侧边栏并传递数据"},
			},
		},
		{
			"category": "条件包含",
			"examples": []map[string]string{
				{"syntax": "{{if .showNav}}{{template \"nav.html\" .}}{{end}}", "description": "条件包含导航"},
				{"syntax": "{{with .userInfo}}{{template \"profile.html\" .}}{{end}}", "description": "数据存在时包含用户配置"},
			},
		},
		{
			"category": "循环包含",
			"examples": []map[string]string{
				{"syntax": "{{range .articles}}{{template \"article.html\" .}}{{end}}", "description": "循环渲染文章"},
				{"syntax": "{{range $i, $item := .items}}{{template \"item.html\" $item}}{{end}}", "description": "带索引的循环包含"},
			},
		},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>📋 Template函数特性</h3>
		<div style="background: linear-gradient(135deg, #667eea, #764ba2); color: white; padding: 15px; border-radius: 10px;">
			<h4 style="margin: 0 0 15px 0;">🎨 Layout环境优势</h4>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				简洁语法: template
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Layout集成: 完美支持
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				性能优化: 缓存机制
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px;">
				响应式: 自适应设计
			</div>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_template_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// ============= 组合函数Layout版本测试方法 =============

// TestAllFunctionsLayout 测试所有模板函数的综合使用 + Layout布局
func (c *BeeGoFunctionsController) TestAllFunctionsLayout() {
	c.SetData("Title", "所有模板函数综合测试 (Layout版)")
	c.SetData("Description", "使用Layout布局展示所有模板函数在一个页面中的协同工作")
	c.SetData("TestType", "AllFunctions + Layout")
	c.SetData("Message", "展示所有模板函数在一个页面中的协同工作，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "综合测试 Layout版",
		"description": "演示Include、TemplateInclude、Partial、ComponentTemplate、RenderTemplate在Layout系统中的综合使用",
		"timestamp":   time.Now(),
	})

	// 全局页面配置
	c.SetData("PageConfig", map[string]any{
		"title":       "模板函数综合演示 (Layout版)",
		"description": "全面展示各种模板函数在Layout系统中的强大功能",
		"version":     "1.0.0",
	})

	// 头部导航数据
	c.SetData("Navigation", []map[string]any{
		{"name": "首页", "url": "/", "active": false},
		{"name": "测试中心", "url": "/test", "active": true},
		{"name": "文档", "url": "/docs", "active": false},
	})

	// 侧边栏数据
	c.SetData("SidebarMenu", []map[string]any{
		{"title": "单独测试", "items": []map[string]any{
			{"name": "Include", "url": "/test/beego-functions/include"},
			{"name": "TemplateInclude", "url": "/test/beego-functions/templateinclude"},
			{"name": "Partial", "url": "/test/beego-functions/partial"},
		}},
		{"title": "组合测试", "items": []map[string]any{
			{"name": "Include+TemplateInclude", "url": "/test/beego-functions/include-templateinclude"},
			{"name": "Partial+Component", "url": "/test/beego-functions/partial-component"},
		}},
	})

	// 主要内容区域数据
	c.SetData("MainSections", []map[string]any{
		{
			"type":    "component",
			"name":    "welcome-card",
			"title":   "欢迎使用Layout版本",
			"content": "这是通过ComponentTemplate渲染的欢迎卡片 (Layout版)",
		},
		{
			"type":    "partial",
			"name":    "statistics",
			"title":   "统计信息",
			"content": "这部分通过Partial包含统计信息 (Layout版)",
		},
		{
			"type":    "include",
			"name":    "recent-activity",
			"title":   "最近活动",
			"content": "这部分通过Include包含最近活动 (Layout版)",
		},
	})

	// 页脚数据
	c.SetData("FooterData", map[string]any{
		"copyright": "© 2024 YYHertz Framework Layout版. All rights reserved.",
		"links": []map[string]string{
			{"name": "关于我们", "url": "/about"},
			{"name": "联系我们", "url": "/contact"},
			{"name": "帮助中心", "url": "/help"},
		},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>🚀 综合功能展示</h3>
		<div style="background: linear-gradient(135deg, #667eea, #764ba2); color: white; padding: 15px; border-radius: 10px;">
			<h4 style="margin: 0 0 15px 0;">🎨 Layout集成特性</h4>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				完整的模板函数集成
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				统一的Layout框架
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				企业级功能演示
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px;">
				交互式测试工具
			</div>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_allfunctions_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestNestedIncludesLayout 测试嵌套包含 + Layout布局
func (c *BeeGoFunctionsController) TestNestedIncludesLayout() {
	c.SetData("Title", "嵌套包含测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试模板函数的嵌套调用能力")
	c.SetData("TestType", "NestedIncludes + Layout")
	c.SetData("Message", "测试模板函数的嵌套调用能力，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "嵌套测试 Layout版",
		"description": "演示模板可以包含其他模板，而被包含的模板还可以再包含更多模板，在Layout系统中的表现",
		"timestamp":   time.Now(),
		"level":       1,
	})

	// 多层嵌套数据
	c.SetData("NestedData", map[string]any{
		"level1": map[string]any{
			"title": "第一层 (Layout版)",
			"level2": map[string]any{
				"title": "第二层 (Layout版)",
				"level3": map[string]any{
					"title":   "第三层 (Layout版)",
					"content": "这是最深层的内容 (Layout版)",
				},
			},
		},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>🎯 嵌套包含系统</h3>
		<div style="background: linear-gradient(135deg, #4facfe, #00f2fe); color: white; padding: 15px; border-radius: 10px;">
			<h4 style="margin: 0 0 15px 0;">🔄 嵌套层级结构</h4>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Level 1: 主Layout框架
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Level 2: 内容模板包含
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Level 3: 部分模板嵌套
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px;">
				Level 4: 组件深度包含
			</div>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_nestedincludes_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestErrorHandlingLayout 测试错误处理 + Layout布局
func (c *BeeGoFunctionsController) TestErrorHandlingLayout() {
	c.SetData("Title", "错误处理测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试模板函数的错误处理机制")
	c.SetData("TestType", "ErrorHandling + Layout")
	c.SetData("Message", "测试模板函数的错误处理机制，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "错误处理测试 Layout版",
		"description": "验证当包含不存在的模板或传递错误数据时在Layout系统中的处理情况",
		"timestamp":   time.Now(),
	})

	// 错误场景数据
	c.SetData("ErrorScenarios", []map[string]any{
		{
			"name":        "不存在的模板 (Layout环境)",
			"template":    "non_existent_template.html",
			"description": "测试在Layout环境中包含不存在的模板文件",
		},
		{
			"name":        "空数据 (Layout环境)",
			"template":    "test_template.html",
			"data":        nil,
			"description": "测试在Layout环境中传递空数据到模板",
		},
		{
			"name":        "错误数据类型 (Layout环境)",
			"template":    "test_template.html",
			"data":        "invalid_data",
			"description": "测试在Layout环境中传递错误类型的数据",
		},
		{
			"name":        "Layout嵌套错误",
			"template":    "nested_error_template.html",
			"data":        map[string]any{"invalid": true},
			"description": "测试Layout系统中的嵌套包含错误处理",
		},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>⚠️ 错误处理机制</h3>
		<div style="background: linear-gradient(135deg, #ff9a9e, #fecfef); color: #2d3436; padding: 15px; border-radius: 10px;">
			<h4 style="margin: 0 0 15px 0;">🛡️ Layout错误防护</h4>
			<div style="background: rgba(255,255,255,0.7); padding: 8px; border-radius: 5px; margin: 5px 0;">
				模板文件缺失检测
			</div>
			<div style="background: rgba(255,255,255,0.7); padding: 8px; border-radius: 5px; margin: 5px 0;">
				数据类型验证
			</div>
			<div style="background: rgba(255,255,255,0.7); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Layout嵌套错误处理
			</div>
			<div style="background: rgba(255,255,255,0.7); padding: 8px; border-radius: 5px;">
				优雅降级机制
			</div>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_errorhandling_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestIncludeTemplateIncludeLayout 测试Include + TemplateInclude组合 + Layout布局
func (c *BeeGoFunctionsController) TestIncludeTemplateIncludeLayout() {
	c.SetData("Title", "Include + TemplateInclude 组合测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试Include和TemplateInclude函数的组合使用")
	c.SetData("TestType", "Include+TemplateInclude + Layout")
	c.SetData("Message", "测试Include和TemplateInclude函数的组合使用，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "组合测试 Layout版",
		"description": "演示多个模板函数在Layout系统中的组合使用场景",
		"timestamp":   time.Now(),
	})

	// 准备头部数据
	c.SetData("Header", map[string]any{
		"title":      "组合测试页面 (Layout版)",
		"breadcrumb": []string{"首页", "测试", "组合测试", "Layout版"},
	})

	// 准备主要内容数据
	c.SetData("MainContent", map[string]any{
		"sections": []map[string]any{
			{
				"title":   "Include部分 (Layout版)",
				"content": "这部分内容通过Include函数在Layout中加载",
				"type":    "include",
			},
			{
				"title":   "TemplateInclude部分 (Layout版)",
				"content": "这部分内容通过TemplateInclude函数在Layout中加载",
				"type":    "templateinclude",
			},
		},
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>🔗 函数组合系统</h3>
		<div style="background: linear-gradient(135deg, #a8edea, #fed6e3); padding: 15px; border-radius: 10px; color: #2d3436;">
			<h4 style="margin: 0 0 15px 0;">🎯 Layout组合特性</h4>
			<div style="background: rgba(255,255,255,0.7); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Include + TemplateInclude
			</div>
			<div style="background: rgba(255,255,255,0.7); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Layout框架集成
			</div>
			<div style="background: rgba(255,255,255,0.7); padding: 8px; border-radius: 5px; margin: 5px 0;">
				统一数据传递
			</div>
			<div style="background: rgba(255,255,255,0.7); padding: 8px; border-radius: 5px;">
				响应式布局支持
			</div>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_includetemplate_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}

// TestPartialComponentLayout 测试Partial + ComponentTemplate组合 + Layout布局
func (c *BeeGoFunctionsController) TestPartialComponentLayout() {
	c.SetData("Title", "Partial + ComponentTemplate 组合测试 (Layout版)")
	c.SetData("Description", "使用Layout布局测试Partial和ComponentTemplate函数的组合使用")
	c.SetData("TestType", "Partial+Component + Layout")
	c.SetData("Message", "测试Partial和ComponentTemplate函数的组合使用，使用Layout布局")
	c.SetData("TestData", map[string]any{
		"name":        "组合测试 Layout版",
		"description": "演示部分模板和组件在Layout系统中的组合使用",
		"timestamp":   time.Now(),
	})

	// 准备侧边栏数据（用于Partial）
	c.SetData("SidebarData", map[string]any{
		"title": "功能菜单 (Layout版)",
		"items": []map[string]any{
			{"name": "用户管理", "url": "/users", "icon": "user"},
			{"name": "角色管理", "url": "/roles", "icon": "shield"},
			{"name": "权限管理", "url": "/permissions", "icon": "key"},
			{"name": "Layout管理", "url": "/layouts", "icon": "grid"},
		},
	})

	// 准备组件数据（用于ComponentTemplate）
	c.SetData("Cards", []map[string]any{
		{
			"title":       "统计卡片1 (Layout版)",
			"value":       "1,234",
			"description": "总用户数",
			"icon":        "users",
			"color":       "blue",
		},
		{
			"title":       "统计卡片2 (Layout版)",
			"value":       "89.5%",
			"description": "系统利用率",
			"icon":        "chart",
			"color":       "green",
		},
		{
			"title":       "统计卡片3 (Layout版)",
			"value":       "156",
			"description": "在线用户",
			"icon":        "activity",
			"color":       "orange",
		},
	})

	// 添加卡片组件数据（修复template中{{component "card" .CardData}}的数据缺失问题）
	c.SetData("CardData", map[string]any{
		"cardTitle":   "ComponentTemplate Layout演示卡片",
		"cardContent": "这是在Layout环境中通过ComponentTemplate渲染的卡片组件示例，展示了组件化开发的强大功能。",
		"buttonText":  "查看详情",
		"buttonUrl":   "/test/beego-functions/componenttemplate_layout",
	})

	// 侧边栏数据
	c.SetData("Sidebar", `
		<h3>⚙️ Partial+Component</h3>
		<div style="background: linear-gradient(135deg, #89f7fe, #66a6ff); color: white; padding: 15px; border-radius: 10px;">
			<h4 style="margin: 0 0 15px 0;">🧩 Layout组合架构</h4>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Partial模块化设计
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Component组件化开发
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px; margin: 5px 0;">
				Layout统一管理
			</div>
			<div style="background: rgba(255,255,255,0.2); padding: 8px; border-radius: 5px;">
				响应式交互体验
			</div>
		</div>
	`)

	c.Layout = "test_functions_layout.html"
	c.TplName = "test/layout/bf_partialcomponent_layout.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString("渲染错误: " + err.Error())
	}
}