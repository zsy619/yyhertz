package controllers

import (
	"fmt"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// DemoController 演示高级功能的控制器
type DemoController struct {
	mvc.BaseController
}

// AdvancedFeatures 高级功能演示
func (c *DemoController) AdvancedFeatures() {
	c.SetData("Title", "高级功能演示")

	// CSRF Token演示
	csrfProvider := &SimpleMockCSRFProvider{}
	view.SetCSRFTokenProvider(csrfProvider)
	c.SetData("CSRFToken", csrfProvider.GenerateSimpleToken())

	// 自动模板推导演示
	candidates := view.InferTemplatePath("UserController", "Profile")
	c.SetData("TemplateCandidates", candidates)

	// 命名约定演示
	conventions := []map[string]interface{}{
		{
			"name":   "BeegoStandard",
			"input":  "UserController.ShowProfile",
			"output": view.ApplyNamingConvention("UserController.ShowProfile", view.BeegoStandard),
		},
		{
			"name":   "SnakeCase",
			"input":  "UserController.ShowProfile",
			"output": view.ApplyNamingConvention("UserController.ShowProfile", view.SnakeCase),
		},
		{
			"name":   "CamelCase",
			"input":  "UserController.ShowProfile",
			"output": view.ApplyNamingConvention("UserController.ShowProfile", view.CamelCase),
		},
		{
			"name":   "KebabCase",
			"input":  "UserController.ShowProfile",
			"output": view.ApplyNamingConvention("UserController.ShowProfile", view.KebabCase),
		},
	}
	c.SetData("NamingConventions", conventions)

	// 自定义函数演示数据
	fileSizeData := map[string]int64{
		"small":  1024,
		"medium": 5 * 1024 * 1024,
		"large":  2 * 1024 * 1024 * 1024,
	}
	c.SetData("FileSizeData", fileSizeData)

	c.Layout = ""
	c.TplName = "demo/advanced.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString(err.Error())
	}
}

// Performance 性能监控
func (c *DemoController) Performance() {
	c.SetData("Title", "性能监控")

	engine := view.GetDefaultEngine()

	// 缓存统计
	cacheStats := engine.GetCacheStats()
	c.SetData("CacheStats", cacheStats)

	// 内存使用情况
	memStats := engine.GetMemoryUsage()
	c.SetData("MemStats", memStats)

	// 健康检查
	healthStatus := "良好"
	if err := engine.HealthCheck(); err != nil {
		healthStatus = fmt.Sprintf("异常: %v", err)
	}
	c.SetData("HealthStatus", healthStatus)

	// 统计信息总览
	globalStats := view.GetUnifiedStats()
	c.SetData("GlobalStats", globalStats)

	c.SetData("Now", time.Now())

	c.Layout = ""
	c.TplName = "demo/performance.html"
	if err := c.Render(); err != nil {
		c.Ctx.WriteString(err.Error())
	}
}

// CsrfTest CSRF测试处理
func (c *DemoController) CsrfTest() {
	if string(c.Ctx.Method()) != "POST" {
		c.Ctx.JSON(405, map[string]string{
			"error": "Method not allowed",
		})
		return
	}

	// 这里可以添加CSRF验证逻辑
	username := c.GetString("username")
	csrfToken := c.GetString("csrf_token")

	result := map[string]interface{}{
		"status":  "success",
		"message": "CSRF保护测试通过",
		"data": map[string]interface{}{
			"username":   username,
			"csrf_token": csrfToken,
			"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	c.Ctx.JSON(200, result)
}

// SimpleMockCSRFProvider 简单的CSRF提供者实现（演示用）
type SimpleMockCSRFProvider struct{}

func (p *SimpleMockCSRFProvider) GenerateSimpleToken() string {
	return fmt.Sprintf("csrf-token-%d", time.Now().UnixNano())
}
