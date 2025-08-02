package annotation

import (
	"reflect"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// 示例：使用struct标签的控制器

// UserController 用户控制器示例
type UserController struct {
	core.BaseController `rest:"" mapping:"/api/users"` // REST控制器，基础路径为 /api/users
}

// ProductController 产品控制器示例
type ProductController struct {
	core.BaseController `controller:"" mapping:"/api/products"` // MVC控制器，基础路径为 /api/products
}

// AdminController 管理员控制器示例
type AdminController struct {
	core.BaseController `rest:"" mapping:"/api/admin"`
}

// UserService 用户服务示例
type UserService struct {
	_ string `service:"userService"` // 服务组件
}

// UserRepository 用户仓库示例
type UserRepository struct {
	_ string `repository:"userRepo"` // 数据访问组件
}

// 示例：在init()函数中注册方法映射

func init() {
	// 获取控制器类型
	userControllerType := reflect.TypeOf((*UserController)(nil)).Elem()
	productControllerType := reflect.TypeOf((*ProductController)(nil)).Elem()
	adminControllerType := reflect.TypeOf((*AdminController)(nil)).Elem()

	// 注册UserController的方法映射
	RegisterGetMethod(userControllerType, "GetUsers", "/").
		WithDescription("获取用户列表").
		WithQueryParam("page", false, "1").
		WithQueryParam("size", false, "10")

	RegisterGetMethod(userControllerType, "GetUser", "/{id}").
		WithDescription("根据ID获取用户").
		WithPathParam("id", true)

	RegisterPostMethod(userControllerType, "CreateUser", "/").
		WithDescription("创建新用户").
		WithBodyParam(true)

	RegisterPutMethod(userControllerType, "UpdateUser", "/{id}").
		WithDescription("更新用户信息").
		WithPathParam("id", true).
		WithBodyParam(true)

	RegisterDeleteMethod(userControllerType, "DeleteUser", "/{id}").
		WithDescription("删除用户").
		WithPathParam("id", true)

	// 注册ProductController的方法映射
	RegisterGetMethod(productControllerType, "GetProducts", "/").
		WithDescription("获取产品列表").
		WithQueryParam("category", false, "").
		WithQueryParam("page", false, "1")

	RegisterGetMethod(productControllerType, "GetProduct", "/{id}").
		WithDescription("根据ID获取产品").
		WithPathParam("id", true)

	RegisterPostMethod(productControllerType, "CreateProduct", "/").
		WithDescription("创建新产品").
		WithBodyParam(true)

	// 注册AdminController的方法映射
	RegisterGetMethod(adminControllerType, "GetDashboard", "/dashboard").
		WithDescription("管理员仪表板")

	RegisterGetMethod(adminControllerType, "GetSystemInfo", "/system/info").
		WithDescription("获取系统信息")

	RegisterPostMethod(adminControllerType, "SystemBackup", "/system/backup").
		WithDescription("系统备份")
}

// 示例：控制器方法实现

// UserRequest 用户请求结构
type UserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"min=0,max=120"`
}

// UserResponse 用户响应结构
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// GetUsers 获取用户列表
// 对应路由: GET /api/users?page=1&size=10
func (c *UserController) GetUsers() ([]*UserResponse, error) {
	page := c.GetQuery("page", "1")
	size := c.GetQuery("size", "10")

	// 模拟获取用户列表
	users := []*UserResponse{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25},
		{ID: 2, Name: "李四", Email: "lisi@example.com", Age: 30},
	}

	c.Data["page"] = page
	c.Data["size"] = size

	return users, nil
}

// GetUser 根据ID获取用户
// 对应路由: GET /api/users/{id}
func (c *UserController) GetUser() (*UserResponse, error) {
	id := c.GetParam("id")

	// 模拟根据ID获取用户
	user := &UserResponse{
		ID:    1,
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}

	c.Data["id"] = id

	return user, nil
}

// CreateUser 创建新用户
// 对应路由: POST /api/users
func (c *UserController) CreateUser(req *UserRequest) (*UserResponse, error) {
	// 模拟创建用户
	user := &UserResponse{
		ID:    100,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	return user, nil
}

// UpdateUser 更新用户信息
// 对应路由: PUT /api/users/{id}
func (c *UserController) UpdateUser(req *UserRequest) (*UserResponse, error) {
	id := c.GetParam("id")

	// 模拟更新用户
	user := &UserResponse{
		ID:    1,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	c.Data["id"] = id

	return user, nil
}

// DeleteUser 删除用户
// 对应路由: DELETE /api/users/{id}
func (c *UserController) DeleteUser() (map[string]interface{}, error) {
	id := c.GetParam("id")

	// 模拟删除用户
	result := map[string]interface{}{
		"success": true,
		"message": "用户删除成功",
		"id":      id,
	}

	return result, nil
}

// ProductController 方法实现示例

// ProductRequest 产品请求结构
type ProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Category    string  `json:"category" binding:"required"`
}

// ProductResponse 产品响应结构
type ProductResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

// GetProducts 获取产品列表
// 对应路由: GET /api/products?category=electronics&page=1
func (c *ProductController) GetProducts() ([]*ProductResponse, error) {
	category := c.GetQuery("category", "")
	page := c.GetQuery("page", "1")

	// 模拟获取产品列表
	products := []*ProductResponse{
		{ID: 1, Name: "iPhone 15", Description: "最新的iPhone", Price: 6999.0, Category: "electronics"},
		{ID: 2, Name: "MacBook Pro", Description: "专业笔记本", Price: 15999.0, Category: "electronics"},
	}

	c.Data["category"] = category
	c.Data["page"] = page

	return products, nil
}

// GetProduct 根据ID获取产品
// 对应路由: GET /api/products/{id}
func (c *ProductController) GetProduct() (*ProductResponse, error) {
	id := c.GetParam("id")

	// 模拟根据ID获取产品
	product := &ProductResponse{
		ID:          1,
		Name:        "iPhone 15",
		Description: "最新的iPhone",
		Price:       6999.0,
		Category:    "electronics",
	}

	c.Data["id"] = id

	return product, nil
}

// CreateProduct 创建新产品
// 对应路由: POST /api/products
func (c *ProductController) CreateProduct(req *ProductRequest) (*ProductResponse, error) {
	// 模拟创建产品
	product := &ProductResponse{
		ID:          100,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
	}

	return product, nil
}

// AdminController 方法实现示例

// GetDashboard 管理员仪表板
// 对应路由: GET /api/admin/dashboard
func (c *AdminController) GetDashboard() (map[string]interface{}, error) {
	dashboard := map[string]interface{}{
		"userCount":    1000,
		"productCount": 500,
		"orderCount":   2000,
		"revenue":      100000.0,
	}

	return dashboard, nil
}

// GetSystemInfo 获取系统信息
// 对应路由: GET /api/admin/system/info
func (c *AdminController) GetSystemInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{
		"version":    "1.0.0",
		"uptime":     "30 days",
		"memory":     "512MB",
		"cpu":        "2 cores",
		"goroutines": 100,
	}

	return info, nil
}

// SystemBackup 系统备份
// 对应路由: POST /api/admin/system/backup
func (c *AdminController) SystemBackup() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"success":   true,
		"message":   "系统备份已启动",
		"backupId":  "backup_20240801_001",
		"timestamp": "2024-08-01T22:00:00Z",
	}

	return result, nil
}