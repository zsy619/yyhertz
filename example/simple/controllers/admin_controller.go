package controllers

import (
	"github.com/zsy619/yyhertz/framework/mvc"
)

type AdminController struct {
	mvc.BaseController
}

func (c *AdminController) Prepare() {
	// 验证管理员权限
	token := string(c.Ctx.GetHeader("Authorization"))
	if token == "" {
		c.Redirect("/auth/login")
		return
	}
	
	// 这里应该验证token的有效性
	if token != "Bearer admin-token" {
		c.Error(403, "权限不足")
		return
	}
}

func (c *AdminController) GetDashboard() {
	// 模拟统计数据
	stats := map[string]any{
		"Users":    150,
		"Products": 75,
		"Orders":   320,
		"Revenue":  "125,680",
	}
	
	// 模拟最近活动
	recentActivities := []map[string]any{
		{
			"Time":   "2024-07-29 10:30:00",
			"User":   "张三",
			"Action": "登录系统",
			"Status": "success",
		},
		{
			"Time":   "2024-07-29 10:25:00",
			"User":   "李四",
			"Action": "修改个人资料",
			"Status": "success",
		},
		{
			"Time":   "2024-07-29 10:20:00",
			"User":   "王五",
			"Action": "尝试访问管理页面",
			"Status": "warning",
		},
		{
			"Time":   "2024-07-29 10:15:00",
			"User":   "系统",
			"Action": "数据备份",
			"Status": "success",
		},
	}
	
	// 模拟系统状态
	systemStatus := map[string]any{
		"CPU":    45,
		"Memory": 68,
		"Disk":   32,
	}
	
	c.SetData("Title", "管理员控制台")
	c.SetData("Stats", stats)
	c.SetData("RecentActivities", recentActivities)
	c.SetData("SystemStatus", systemStatus)
	c.RenderWithViewName("admin/dashboard.html")
}

func (c *AdminController) GetUsers() {
	page := c.GetInt("page", 1)
	size := c.GetInt("size", 10)
	
	// 模拟管理员用户列表（比普通用户列表包含更多信息）
	users := []map[string]any{
		{
			"ID":       1,
			"Name":     "张三",
			"Email":    "zhangsan@example.com",
			"Role":     "用户",
			"Status":   "活跃",
			"LastIP":   "192.168.1.100",
			"Created":  "2024-01-15",
		},
		{
			"ID":       2,
			"Name":     "李四",
			"Email":    "lisi@example.com",
			"Role":     "管理员",
			"Status":   "活跃",
			"LastIP":   "192.168.1.101",
			"Created":  "2024-02-20",
		},
		{
			"ID":       3,
			"Name":     "王五",
			"Email":    "wangwu@example.com",
			"Role":     "用户",
			"Status":   "禁用",
			"LastIP":   "192.168.1.102",
			"Created":  "2024-03-10",
		},
	}
	
	c.JSON(map[string]any{
		"success": true,
		"message": "获取用户列表成功",
		"data": map[string]any{
			"users": users,
			"page":  page,
			"size":  size,
			"total": 150,
		},
	})
}

func (c *AdminController) PostClearCache() {
	// 模拟清除缓存操作
	c.JSON(map[string]any{
		"success": true,
		"message": "缓存清除成功",
		"time":    "2024-07-29 10:30:00",
	})
}

func (c *AdminController) GetSettings() {
	// 模拟系统设置
	settings := map[string]any{
		"SiteName":        "Hertz MVC Framework",
		"SiteDescription": "基于CloudWeGo-Hertz的类Beego框架",
		"AdminEmail":      "admin@example.com",
		"MaxUploadSize":   "10MB",
		"AllowRegister":   true,
		"MaintenanceMode": false,
	}
	
	c.JSON(map[string]any{
		"success": true,
		"data":    settings,
	})
}

func (c *AdminController) PostSettings() {
	siteName := c.GetForm("site_name")
	siteDesc := c.GetForm("site_description")
	adminEmail := c.GetForm("admin_email")
	
	// 这里应该是保存设置到数据库的逻辑
	c.JSON(map[string]any{
		"success": true,
		"message": "设置保存成功",
		"data": map[string]any{
			"site_name":        siteName,
			"site_description": siteDesc,
			"admin_email":      adminEmail,
		},
	})
}