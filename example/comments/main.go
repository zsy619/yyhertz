package main

import (
	"log"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/comment"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// 基于注释的注解示例

// UserController 用户控制器
// @RestController
// @RequestMapping("/api/v1/users")
// @Description("用户管理REST API控制器")
type UserController struct {
	core.BaseController
}

// GetUsers 获取用户列表
// @GetMapping("/")
// @Description("分页获取用户列表")
// @RequestParam(name="page", required=false, defaultValue="1")
// @RequestParam(name="size", required=false, defaultValue="10")
// @RequestParam(name="keyword", required=false, defaultValue="")
func (c *UserController) GetUsers() ([]*UserResponse, error) {
	page := c.GetQuery("page", "1")
	size := c.GetQuery("size", "10")
	keyword := c.GetQuery("keyword", "")

	log.Printf("获取用户列表: page=%s, size=%s, keyword=%s", page, size, keyword)

	// 模拟数据
	users := []*UserResponse{
		{ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25, Status: "active"},
		{ID: 2, Name: "李四", Email: "li@example.com", Age: 30, Status: "active"},
		{ID: 3, Name: "王五", Email: "wang@example.com", Age: 28, Status: "inactive"},
	}

	return users, nil
}

// GetUser 获取单个用户
// @GetMapping("/{id}")
// @Description("根据ID获取用户详情")
// @PathVariable("id")
func (c *UserController) GetUser() (*UserResponse, error) {
	id := c.GetParam("id")

	log.Printf("获取用户详情: id=%s", id)

	user := &UserResponse{
		ID:     1,
		Name:   "张三",
		Email:  "zhang@example.com",
		Age:    25,
		Status: "active",
	}

	return user, nil
}

// CreateUser 创建用户
// @PostMapping("/")
// @Description("创建新用户")
// @RequestBody
func (c *UserController) CreateUser(req *UserRequest) (*UserResponse, error) {
	log.Printf("创建用户: %+v", req)

	user := &UserResponse{
		ID:     100,
		Name:   req.Name,
		Email:  req.Email,
		Age:    req.Age,
		Status: "active",
	}

	return user, nil
}

// UpdateUser 更新用户
// @PutMapping("/{id}")
// @Description("更新用户信息")
// @PathVariable("id")
// @RequestBody
func (c *UserController) UpdateUser(req *UserRequest) (*UserResponse, error) {
	id := c.GetParam("id")

	log.Printf("更新用户: id=%s, data=%+v", id, req)

	user := &UserResponse{
		ID:     1,
		Name:   req.Name,
		Email:  req.Email,
		Age:    req.Age,
		Status: "active",
	}

	return user, nil
}

// DeleteUser 删除用户
// @DeleteMapping("/{id}")
// @Description("删除用户")
// @PathVariable("id")
func (c *UserController) DeleteUser() (map[string]interface{}, error) {
	id := c.GetParam("id")

	log.Printf("删除用户: id=%s", id)

	return map[string]interface{}{
		"success": true,
		"message": "用户删除成功",
		"id":      id,
	}, nil
}

// SearchUsers 搜索用户
// @GetMapping("/search")
// @Description("搜索用户")
// @RequestParam(name="q", required=true)
// @RequestParam(name="type", required=false, defaultValue="name")
// @RequestHeader(name="X-Request-ID", required=false)
func (c *UserController) SearchUsers() ([]*UserResponse, error) {
	query := c.GetQuery("q", "")
	searchType := c.GetQuery("type", "name")
	requestID := c.GetHeader("X-Request-ID")

	log.Printf("搜索用户: q=%s, type=%s, requestID=%s", query, searchType, string(requestID))

	users := []*UserResponse{
		{ID: 1, Name: "张三", Email: "zhang@example.com", Age: 25, Status: "active"},
	}

	return users, nil
}

// ProductController 产品控制器
// @RestController
// @RequestMapping("/api/v1/products")
// @Description("产品管理REST API控制器")
type ProductController struct {
	core.BaseController
}

// GetProducts 获取产品列表
// @GetMapping("/")
// @Description("获取产品列表")
// @RequestParam(name="category", required=false)
// @RequestParam(name="page", required=false, defaultValue="1")
// @RequestParam(name="limit", required=false, defaultValue="20")
func (c *ProductController) GetProducts() ([]*ProductResponse, error) {
	category := c.GetQuery("category", "")
	page := c.GetQuery("page", "1")
	limit := c.GetQuery("limit", "20")

	log.Printf("获取产品列表: category=%s, page=%s, limit=%s", category, page, limit)

	products := []*ProductResponse{
		{ID: 1, Name: "iPhone 15", Category: "electronics", Price: 6999.0, Stock: 100},
		{ID: 2, Name: "MacBook Pro", Category: "electronics", Price: 15999.0, Stock: 50},
		{ID: 3, Name: "iPad Pro", Category: "electronics", Price: 8999.0, Stock: 75},
	}

	return products, nil
}

// GetProduct 获取单个产品
// @GetMapping("/{id}")
// @Description("根据ID获取产品详情")
// @PathVariable("id")
func (c *ProductController) GetProduct() (*ProductResponse, error) {
	id := c.GetParam("id")

	log.Printf("获取产品详情: id=%s", id)

	product := &ProductResponse{
		ID:       1,
		Name:     "iPhone 15",
		Category: "electronics",
		Price:    6999.0,
		Stock:    100,
	}

	return product, nil
}

// CreateProduct 创建产品
// @PostMapping("/")
// @Description("创建新产品")
// @RequestBody
// @Middleware("auth", "ratelimit")
func (c *ProductController) CreateProduct(req *ProductRequest) (*ProductResponse, error) {
	log.Printf("创建产品: %+v", req)

	product := &ProductResponse{
		ID:       100,
		Name:     req.Name,
		Category: req.Category,
		Price:    req.Price,
		Stock:    req.Stock,
	}

	return product, nil
}

// WebController Web页面控制器
// @Controller
// @RequestMapping("/web")
// @Description("Web页面控制器")
type WebController struct {
	core.BaseController
}

// Index 首页
// @GetMapping("/")
// @Description("网站首页")
func (c *WebController) Index() {
	c.Data["Title"] = "首页"
	c.Data["Message"] = "欢迎来到基于注释注解的YYHertz框架!"
	c.TplName = "index.html"
}

// UserList 用户列表页面
// @GetMapping("/users")
// @Description("用户列表页面")
// @RequestParam(name="page", required=false, defaultValue="1")
func (c *WebController) UserList() {
	page := c.GetQuery("page", "1")

	c.Data["Title"] = "用户列表"
	c.Data["Users"] = []map[string]interface{}{
		{"ID": 1, "Name": "张三", "Email": "zhang@example.com"},
		{"ID": 2, "Name": "李四", "Email": "li@example.com"},
		{"ID": 3, "Name": "王五", "Email": "wang@example.com"},
	}
	c.Data["Page"] = page
	c.TplName = "users/list.html"
}

// AdminController 管理员控制器
// @RestController
// @RequestMapping("/api/admin")
// @Description("管理员控制器")
// @Middleware("auth", "admin")
type AdminController struct {
	core.BaseController
}

// GetDashboard 获取仪表板数据
// @GetMapping("/dashboard")
// @Description("获取管理员仪表板数据")
func (c *AdminController) GetDashboard() (map[string]interface{}, error) {
	dashboard := map[string]interface{}{
		"userCount":    1000,
		"productCount": 500,
		"orderCount":   2000,
		"revenue":      100000.0,
		"systemStatus": "healthy",
		"timestamp":    "2024-08-01T22:00:00Z",
	}

	return dashboard, nil
}

// GetSystemInfo 获取系统信息
// @GetMapping("/system/info")
// @Description("获取系统信息")
func (c *AdminController) GetSystemInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{
		"version":     "1.0.0",
		"environment": "production",
		"uptime":      "30 days",
		"memory":      "512MB",
		"cpu":         "2 cores",
		"goroutines":  100,
		"connections": 50,
	}

	return info, nil
}

// BackupSystem 系统备份
// @PostMapping("/system/backup")
// @Description("执行系统备份")
// @RequestBody
func (c *AdminController) BackupSystem(req *BackupRequest) (map[string]interface{}, error) {
	log.Printf("执行系统备份: %+v", req)

	result := map[string]interface{}{
		"success":    true,
		"message":    "备份任务已启动",
		"backupId":   "backup_20240801_001",
		"type":       req.Type,
		"timestamp":  "2024-08-01T22:00:00Z",
		"compressed": req.Compression,
	}

	return result, nil
}

// 请求/响应结构定义

// UserRequest 用户请求结构
type UserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"min=0,max=120"`
}

// UserResponse 用户响应结构
type UserResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	Status string `json:"status"`
}

// ProductRequest 产品请求结构
type ProductRequest struct {
	Name     string  `json:"name" binding:"required"`
	Category string  `json:"category" binding:"required"`
	Price    float64 `json:"price" binding:"required,min=0"`
	Stock    int     `json:"stock" binding:"min=0"`
}

// ProductResponse 产品响应结构
type ProductResponse struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
}

// BackupRequest 备份请求结构
type BackupRequest struct {
	Type        string   `json:"type" binding:"required"` // full, incremental
	Tables      []string `json:"tables"`                  // 指定表名
	Compression bool     `json:"compression"`             // 是否压缩
	Schedule    string   `json:"schedule"`                // 定时备份
}

func main() {
	// 创建Hertz引擎
	h := mvc.HertzApp

	// 创建支持注释注解的应用
	app := comment.NewCommentWithApp(h)

	// 自动扫描并注册控制器
	app.AutoScanAndRegister(
		&UserController{},
		&ProductController{},
		&WebController{},
		&AdminController{},
	)

	// 获取所有注册的路由信息
	routes := app.GetRoutes()
	log.Printf("🚀 注册了 %d 个基于注释的路由:", len(routes))
	for _, route := range routes {
		log.Printf("  %s %s -> %s.%s - %s",
			route.HTTPMethod,
			route.Path,
			route.TypeName,
			route.MethodName,
			route.Description)
	}

	// 路由分析
	collector := comment.NewRouteCollector().CollectFromApp(app)
	analyzer := comment.NewRouteAnalyzer(collector)

	log.Printf("\n📊 路由统计:")
	log.Printf("  总路由数: %d", collector.GetRouteCount())
	log.Printf("  控制器数: %d", collector.GetControllerCount())

	methodCounts := collector.GetMethodCount()
	for method, count := range methodCounts {
		log.Printf("  %s: %d", method, count)
	}

	// 分析重复路由
	duplicates := analyzer.AnalyzeDuplicates()
	if len(duplicates) > 0 {
		log.Printf("\n⚠️ 发现重复路由:")
		for _, duplicate := range duplicates {
			log.Printf("  %s -> %v", duplicate[0], duplicate[1:])
		}
	} else {
		log.Printf("\n✅ 未发现重复路由")
	}

	// RESTful分析
	restPatterns := analyzer.AnalyzeRESTfulness()
	log.Printf("\n🎯 RESTful模式分析:")
	for pattern, paths := range restPatterns {
		log.Printf("  %s: %v", pattern, paths)
	}

	// 启动服务器
	log.Println("\n🌟 基于注释注解的示例服务器启动在 :8888")
	log.Println("\n📋 API接口:")
	log.Println("用户管理:")
	log.Println("  GET    /api/v1/users              - 获取用户列表")
	log.Println("  GET    /api/v1/users/1            - 获取用户详情")
	log.Println("  GET    /api/v1/users/search?q=张三 - 搜索用户")
	log.Println("  POST   /api/v1/users              - 创建用户")
	log.Println("  PUT    /api/v1/users/1            - 更新用户")
	log.Println("  DELETE /api/v1/users/1            - 删除用户")
	log.Println("")
	log.Println("产品管理:")
	log.Println("  GET    /api/v1/products           - 获取产品列表")
	log.Println("  GET    /api/v1/products/1         - 获取产品详情")
	log.Println("  POST   /api/v1/products           - 创建产品")
	log.Println("")
	log.Println("Web页面:")
	log.Println("  GET    /web/                      - 首页")
	log.Println("  GET    /web/users                 - 用户列表页面")
	log.Println("")
	log.Println("管理员:")
	log.Println("  GET    /api/admin/dashboard       - 仪表板数据")
	log.Println("  GET    /api/admin/system/info     - 系统信息")
	log.Println("  POST   /api/admin/system/backup   - 系统备份")

	log.Println("\n🔧 测试命令:")
	log.Println("curl -X GET 'http://localhost:8888/api/v1/users?page=1&size=5'")
	log.Println("curl -X POST http://localhost:8888/api/v1/users -H 'Content-Type: application/json' -d '{\"name\":\"新用户\",\"email\":\"new@example.com\",\"age\":25}'")
	log.Println("curl -X GET http://localhost:8888/api/admin/dashboard")

	h.Spin()
}
