package controllers

import (
	"strings"
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

// GetS3DistSearchBill 搜索S3分发账单
func (c *AdminController) GetS3DistSearchBill() {
	status := c.GetInt("status", 0)
	limit := c.GetInt("limit", 10)
	page := c.GetInt("page", 1)
	keyword := c.GetString("keyword")
	
	// 模拟账单数据
	bills := []map[string]any{
		{
			"ID":          1001,
			"BillNo":      "S3-2024-001",
			"Amount":      1250.50,
			"Status":      1, // 1:已支付 2:未支付 3:已取消
			"StatusText":  "已支付",
			"Customer":    "张三企业",
			"CreateTime":  "2024-08-25 10:30:00",
			"PayTime":     "2024-08-25 11:15:00",
			"Description": "S3存储费用-8月份",
		},
		{
			"ID":          1002,
			"BillNo":      "S3-2024-002", 
			"Amount":      875.25,
			"Status":      2,
			"StatusText":  "未支付",
			"Customer":    "李四科技",
			"CreateTime":  "2024-08-25 09:45:00",
			"PayTime":     "",
			"Description": "S3流量费用-8月份",
		},
		{
			"ID":          1003,
			"BillNo":      "S3-2024-003",
			"Amount":      2340.75,
			"Status":      3,
			"StatusText":  "已取消", 
			"Customer":    "王五集团",
			"CreateTime":  "2024-08-24 16:20:00",
			"PayTime":     "",
			"Description": "S3备份费用-8月份",
		},
		{
			"ID":          1004,
			"BillNo":      "S3-2024-004",
			"Amount":      650.00,
			"Status":      1,
			"StatusText":  "已支付",
			"Customer":    "赵六公司",
			"CreateTime":  "2024-08-24 14:10:00", 
			"PayTime":     "2024-08-24 15:30:00",
			"Description": "S3 CDN费用-8月份",
		},
	}
	
	// 根据状态过滤
	filteredBills := []map[string]any{}
	for _, bill := range bills {
		if status == 0 || bill["Status"].(int) == status {
			// 如果有关键词，进行简单的模糊匹配
			if keyword == "" || 
				strings.Contains(strings.ToLower(bill["BillNo"].(string)), strings.ToLower(keyword)) ||
				strings.Contains(strings.ToLower(bill["Customer"].(string)), strings.ToLower(keyword)) {
				filteredBills = append(filteredBills, bill)
			}
		}
	}
	
	// 应用分页
	total := len(filteredBills)
	start := (page - 1) * limit
	end := start + limit
	
	if start > total {
		filteredBills = []map[string]any{}
	} else if end > total {
		filteredBills = filteredBills[start:]
	} else {
		filteredBills = filteredBills[start:end]
	}
	
	c.JSON(map[string]any{
		"success": true,
		"message": "搜索账单成功",
		"data": map[string]any{
			"bills": filteredBills,
			"pagination": map[string]any{
				"page":  page,
				"limit": limit,
				"total": total,
			},
			"filters": map[string]any{
				"status":  status,
				"keyword": keyword,
			},
		},
	})
}