package main

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/framework/mvc/routing"
)

// 实际业务模型定义

// User 用户模型
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username" validate:"required,min=3,max=20"`
	Email     string    `json:"email" validate:"required,email"`
	Age       int       `json:"age" validate:"min=1,max=150"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	Tags      []string  `json:"tags"`
}

// Product 商品模型
type Product struct {
	ID          int      `json:"id"`
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" validate:"min=0"`
	CategoryID  int      `json:"category_id"`
	IsAvailable bool     `json:"is_available"`
	Tags        []string `json:"tags"`
}

// SearchFilter 搜索过滤器
type SearchFilter struct {
	Keyword    string   `json:"keyword"`
	MinPrice   float64  `json:"min_price"`
	MaxPrice   float64  `json:"max_price"`
	Categories []int    `json:"categories"`
	Tags       []string `json:"tags"`
	IsActive   *bool    `json:"is_active,omitempty"` // 指针类型支持null
	DateFrom   string   `json:"date_from"`
	DateTo     string   `json:"date_to"`
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page    int    `json:"page" validate:"min=1"`
	Size    int    `json:"size" validate:"min=1,max=100"`
	SortBy  string `json:"sort_by"`
	SortDir string `json:"sort_dir" validate:"oneof=asc desc"`
}

// 实际业务控制器

// RealWorldUserController 实际应用用户管理控制器
type RealWorldUserController struct {
	pb *routing.ParamBinder
}

func NewRealWorldUserController() *RealWorldUserController {
	return &RealWorldUserController{
		pb: routing.NewParamBinder(),
	}
}

// GetUserDetails 获取单个用户详情
// GET /api/users/{id}?include_profile=true
func (uc *RealWorldUserController) GetUserDetails(c context.Context, ctx *app.RequestContext) {
	// 提取路径参数
	idStr := string(ctx.Param("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]any{
			"error":   "无效的用户ID",
			"message": err.Error(),
		})
		return
	}

	// 提取查询参数
	includeProfile := ctx.Query("include_profile") == "true"

	// 模拟业务逻辑
	user := &User{
		ID:        id,
		Username:  fmt.Sprintf("user_%d", id),
		Email:     fmt.Sprintf("user_%d@example.com", id),
		Age:       25 + id%50,
		IsActive:  true,
		CreatedAt: time.Now().Add(-time.Duration(id) * 24 * time.Hour),
		Tags:      []string{"user", "active"},
	}

	result := map[string]any{
		"user": user,
	}

	if includeProfile {
		result["profile"] = map[string]any{
			"bio":        fmt.Sprintf("用户 %s 的个人简介", user.Username),
			"avatar":     fmt.Sprintf("/avatars/%d.jpg", id),
			"last_login": time.Now().Add(-2 * time.Hour),
		}
	}

	ctx.JSON(consts.StatusOK, result)
}

// CreateUserAccount 创建用户账户
// POST /api/users
func (uc *RealWorldUserController) CreateUserAccount(c context.Context, ctx *app.RequestContext) {
	var userReq User
	if err := ctx.BindAndValidate(&userReq); err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]any{
			"error":   "请求数据验证失败",
			"message": err.Error(),
		})
		return
	}

	// 模拟创建用户
	userReq.ID = int(time.Now().Unix()) // 简单的ID生成
	userReq.CreatedAt = time.Now()
	userReq.IsActive = true

	ctx.JSON(consts.StatusCreated, map[string]any{
		"message": "用户创建成功",
		"user":    userReq,
	})
}

// SearchUserAccounts 搜索用户账户
// GET /api/users/search?keyword=john&page=1&size=10&sort_by=created_at&sort_dir=desc
func (uc *RealWorldUserController) SearchUserAccounts(c context.Context, ctx *app.RequestContext) {
	// 手动参数绑定和转换
	var params struct {
		Keyword string   `query:"keyword"`
		Page    int      `query:"page"`
		Size    int      `query:"size"`
		SortBy  string   `query:"sort_by"`
		SortDir string   `query:"sort_dir"`
		Tags    []string `query:"tags"`
	}

	// 使用ParamBinder进行参数转换
	keyword := ctx.Query("keyword")
	pageStr := ctx.Query("page")
	if pageStr == "" {
		pageStr = "1"
	}
	sizeStr := ctx.Query("size")
	if sizeStr == "" {
		sizeStr = "10"
	}
	sortBy := ctx.Query("sort_by")
	if sortBy == "" {
		sortBy = "id"
	}
	sortDir := ctx.Query("sort_dir")
	if sortDir == "" {
		sortDir = "asc"
	}
	tagsStr := ctx.Query("tags")

	// 转换参数
	page, err := uc.pb.ConvertValue(pageStr, reflect.TypeOf(0))
	if err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]any{
			"error": "无效的页码参数",
		})
		return
	}

	size, err := uc.pb.ConvertValue(sizeStr, reflect.TypeOf(0))
	if err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]any{
			"error": "无效的页面大小参数",
		})
		return
	}

	var tags []string
	if tagsStr != "" {
		tagsResult, err := uc.pb.ConvertValue(tagsStr, reflect.TypeOf([]string{}))
		if err == nil {
			tags = tagsResult.Interface().([]string)
		}
	}

	params.Keyword = keyword
	params.Page = page.Interface().(int)
	params.Size = size.Interface().(int)
	params.SortBy = sortBy
	params.SortDir = sortDir
	params.Tags = tags

	// 模拟搜索结果
	users := make([]User, 0, params.Size)
	for i := 0; i < params.Size; i++ {
		offset := (params.Page-1)*params.Size + i + 1
		users = append(users, User{
			ID:        offset,
			Username:  fmt.Sprintf("user_%d", offset),
			Email:     fmt.Sprintf("user_%d@example.com", offset),
			Age:       20 + offset%60,
			IsActive:  offset%2 == 0,
			CreatedAt: time.Now().Add(-time.Duration(offset) * time.Hour),
			Tags:      []string{"user", "search_result"},
		})
	}

	ctx.JSON(consts.StatusOK, map[string]any{
		"users": users,
		"pagination": map[string]any{
			"page":        params.Page,
			"size":        params.Size,
			"total":       1000, // 模拟总数
			"total_pages": 100,
		},
		"search_params": params,
		"cache_stats":   uc.pb.GetCacheStats(),
	})
}

// ProductController 商品管理控制器
type ProductController struct {
	pb *routing.ParamBinder
}

func NewProductController() *ProductController {
	return &ProductController{
		pb: routing.NewParamBinder(),
	}
}

// GetProduct 获取商品详情
// GET /api/products/{id}
func (pc *ProductController) GetProduct(c context.Context, ctx *app.RequestContext) {
	idStr := string(ctx.Param("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]any{
			"error": "无效的商品ID",
		})
		return
	}

	product := &Product{
		ID:          id,
		Name:        fmt.Sprintf("商品_%d", id),
		Description: fmt.Sprintf("这是商品_%d的详细描述", id),
		Price:       99.99 + float64(id%100),
		CategoryID:  id % 10,
		IsAvailable: id%2 == 0,
		Tags:        []string{"product", "featured"},
	}

	ctx.JSON(consts.StatusOK, map[string]any{
		"product": product,
	})
}

// SearchProducts 商品搜索
// GET /api/products/search?keyword=手机&min_price=100&max_price=5000&categories=1,2,3&page=1&size=20
func (pc *ProductController) SearchProducts(c context.Context, ctx *app.RequestContext) {
	// 复杂的参数绑定示例
	var filter SearchFilter
	var pagination PaginationParams

	// 手动提取和转换参数
	filter.Keyword = ctx.Query("keyword")

	if minPriceStr := ctx.Query("min_price"); minPriceStr != "" {
		if minPriceVal, err := pc.pb.ConvertValue(minPriceStr, reflect.TypeOf(0.0)); err == nil {
			filter.MinPrice = minPriceVal.Interface().(float64)
		}
	}

	if maxPriceStr := ctx.Query("max_price"); maxPriceStr != "" {
		if maxPriceVal, err := pc.pb.ConvertValue(maxPriceStr, reflect.TypeOf(0.0)); err == nil {
			filter.MaxPrice = maxPriceVal.Interface().(float64)
		}
	}

	if categoriesStr := ctx.Query("categories"); categoriesStr != "" {
		// 这里需要特殊处理，因为我们需要[]int而不是[]string
		if categoriesStrSlice, err := pc.pb.ConvertValue(categoriesStr, reflect.TypeOf([]string{})); err == nil {
			strSlice := categoriesStrSlice.Interface().([]string)
			var intSlice []int
			for _, s := range strSlice {
				if val, err := pc.pb.ConvertValue(s, reflect.TypeOf(0)); err == nil {
					intSlice = append(intSlice, val.Interface().(int))
				}
			}
			filter.Categories = intSlice
		}
	}

	if tagsStr := ctx.Query("tags"); tagsStr != "" {
		if tagsVal, err := pc.pb.ConvertValue(tagsStr, reflect.TypeOf([]string{})); err == nil {
			filter.Tags = tagsVal.Interface().([]string)
		}
	}

	// 分页参数
	pageStr := ctx.Query("page")
	if pageStr == "" {
		pageStr = "1"
	}
	sizeStr := ctx.Query("size")
	if sizeStr == "" {
		sizeStr = "20"
	}

	if pageVal, err := pc.pb.ConvertValue(pageStr, reflect.TypeOf(0)); err == nil {
		pagination.Page = pageVal.Interface().(int)
	}
	if sizeVal, err := pc.pb.ConvertValue(sizeStr, reflect.TypeOf(0)); err == nil {
		pagination.Size = sizeVal.Interface().(int)
	}

	pagination.SortBy = ctx.Query("sort_by")
	if pagination.SortBy == "" {
		pagination.SortBy = "id"
	}
	pagination.SortDir = ctx.Query("sort_dir")
	if pagination.SortDir == "" {
		pagination.SortDir = "asc"
	}

	// 模拟搜索结果
	products := make([]Product, 0, pagination.Size)
	for i := 0; i < pagination.Size; i++ {
		offset := (pagination.Page-1)*pagination.Size + i + 1
		products = append(products, Product{
			ID:          offset,
			Name:        fmt.Sprintf("商品_%d_%s", offset, filter.Keyword),
			Description: fmt.Sprintf("匹配关键词 '%s' 的商品描述", filter.Keyword),
			Price:       filter.MinPrice + float64(offset%100),
			CategoryID:  offset % 10,
			IsAvailable: offset%2 == 0,
			Tags:        append(filter.Tags, "search_result"),
		})
	}

	ctx.JSON(consts.StatusOK, map[string]any{
		"products":    products,
		"filter":      filter,
		"pagination":  pagination,
		"total":       500, // 模拟总数
		"cache_stats": pc.pb.GetCacheStats(),
	})
}

// FileController 文件管理控制器
type FileController struct {
	pb *routing.ParamBinder
}

func NewFileController() *FileController {
	return &FileController{
		pb: routing.NewParamBinder(),
	}
}

// UploadFile 上传文件
// POST /api/files/upload
func (fc *FileController) UploadFile(c context.Context, ctx *app.RequestContext) {
	// 模拟文件上传
	filename := string(ctx.FormValue("filename"))
	description := string(ctx.FormValue("description"))
	
	result := map[string]any{
		"message": "文件上传成功",
		"filename": filename,
		"description": description,
		"uploaded_at": time.Now(),
	}
	
	ctx.JSON(consts.StatusOK, result)
}

// DownloadFile 下载文件
// GET /api/files/download/{filename}
func (fc *FileController) DownloadFile(c context.Context, ctx *app.RequestContext) {
	filename := string(ctx.Param("filename"))
	
	ctx.JSON(consts.StatusOK, map[string]any{
		"message": "文件下载",
		"filename": filename,
		"download_url": fmt.Sprintf("/files/%s", filename),
	})
}

// AdminController 管理员控制器
type AdminController struct {
	pb *routing.ParamBinder
}

func NewAdminController() *AdminController {
	return &AdminController{
		pb: routing.NewParamBinder(),
	}
}

// GetSystemStats 获取系统统计信息
// GET /api/admin/stats?date_from=2023-01-01&date_to=2023-12-31&metrics=users,products,orders
func (ac *AdminController) GetSystemStats(c context.Context, ctx *app.RequestContext) {
	// 复杂的参数提取和验证
	dateFrom := ctx.Query("date_from")
	dateTo := ctx.Query("date_to")
	metricsStr := ctx.Query("metrics")

	// 参数验证
	if dateFrom == "" || dateTo == "" {
		ctx.JSON(consts.StatusBadRequest, map[string]any{
			"error": "日期范围参数是必需的",
		})
		return
	}

	var metrics []string
	if metricsStr != "" {
		if metricsVal, err := ac.pb.ConvertValue(metricsStr, reflect.TypeOf([]string{})); err == nil {
			metrics = metricsVal.Interface().([]string)
		}
	} else {
		metrics = []string{"users", "products", "orders"} // 默认指标
	}

	// 模拟统计数据
	stats := map[string]any{
		"date_range": map[string]string{
			"from": dateFrom,
			"to":   dateTo,
		},
		"metrics": map[string]any{},
	}

	for _, metric := range metrics {
		switch metric {
		case "users":
			stats["metrics"].(map[string]any)["users"] = map[string]any{
				"total":       10000,
				"active":      8500,
				"new_today":   150,
				"growth_rate": 12.5,
			}
		case "products":
			stats["metrics"].(map[string]any)["products"] = map[string]any{
				"total":         5000,
				"available":     4200,
				"out_of_stock":  800,
				"new_this_week": 45,
			}
		case "orders":
			stats["metrics"].(map[string]any)["orders"] = map[string]any{
				"total":     25000,
				"completed": 22000,
				"pending":   2500,
				"cancelled": 500,
				"revenue":   1250000.50,
			}
		}
	}

	ctx.JSON(consts.StatusOK, stats)
}

// 服务器设置和路由配置

func SetupRealWorldServer() *server.Hertz {
	h := server.Default()

	// 初始化控制器
	userCtrl := NewRealWorldUserController()
	productCtrl := NewProductController()
	fileCtrl := NewFileController()
	adminCtrl := NewAdminController()

	// API版本分组
	apiV1 := h.Group("/api")

	// 用户相关路由
	userGroup := apiV1.Group("/users")
	{
		userGroup.GET("/:id", userCtrl.GetUserDetails)
		userGroup.POST("/", userCtrl.CreateUserAccount)
		userGroup.GET("/search", userCtrl.SearchUserAccounts)
	}

	// 商品相关路由
	productGroup := apiV1.Group("/products")
	{
		productGroup.GET("/:id", productCtrl.GetProduct)
		productGroup.GET("/search", productCtrl.SearchProducts)
	}

	// 文件管理路由
	fileGroup := apiV1.Group("/files")
	{
		fileGroup.POST("/upload", fileCtrl.UploadFile)
		fileGroup.GET("/download/:filename", fileCtrl.DownloadFile)
	}

	// 管理员路由
	adminGroup := apiV1.Group("/admin")
	{
		adminGroup.GET("/stats", adminCtrl.GetSystemStats)
	}

	// 健康检查
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		pb := routing.NewParamBinder()
		ctx.JSON(consts.StatusOK, map[string]any{
			"status":             "healthy",
			"timestamp":          time.Now(),
			"version":            "1.0.0",
			"param_binder_stats": pb.GetCacheStats(),
		})
	})

	return h
}

// 演示函数

func DemonstrateRealWorldUsage() {
	fmt.Println("🌟 实际应用场景演示")
	fmt.Println("===================")

	// 创建参数绑定器
	pb := routing.NewParamBinder()

	fmt.Println("1. 用户注册参数处理")
	userParams := map[string]string{
		"username":  "john_doe",
		"email":     "john@example.com",
		"age":       "25",
		"is_active": "true",
		"tags":      "user,active,verified",
	}

	for param, value := range userParams {
		var targetType reflect.Type
		switch param {
		case "username", "email":
			targetType = reflect.TypeOf("")
		case "age":
			targetType = reflect.TypeOf(0)
		case "is_active":
			targetType = reflect.TypeOf(false)
		case "tags":
			targetType = reflect.TypeOf([]string{})
		}

		if result, err := pb.ConvertValue(value, targetType); err == nil {
			fmt.Printf("   ✅ %s: %s -> %v\n", param, value, result.Interface())
		} else {
			fmt.Printf("   ❌ %s: %s -> 转换失败: %v\n", param, value, err)
		}
	}

	fmt.Println("\n2. 商品搜索参数处理")
	searchParams := map[string]string{
		"keyword":    "手机",
		"min_price":  "100.50",
		"max_price":  "5000.99",
		"categories": "1,2,3,4",
		"page":       "2",
		"size":       "20",
	}

	for param, value := range searchParams {
		var targetType reflect.Type
		switch param {
		case "keyword":
			targetType = reflect.TypeOf("")
		case "min_price", "max_price":
			targetType = reflect.TypeOf(0.0)
		case "categories":
			targetType = reflect.TypeOf([]string{}) // 注意：这里得到string切片，需要进一步转换为int切片
		case "page", "size":
			targetType = reflect.TypeOf(0)
		}

		if result, err := pb.ConvertValue(value, targetType); err == nil {
			fmt.Printf("   ✅ %s: %s -> %v\n", param, value, result.Interface())
		} else {
			fmt.Printf("   ❌ %s: %s -> 转换失败: %v\n", param, value, err)
		}
	}

	fmt.Println("\n3. 性能统计")
	stats := pb.GetCacheStats()
	fmt.Printf("   📊 缓存统计: %+v\n", stats)

	fmt.Println("\n🚀 HTTP服务器配置示例")
	fmt.Println("服务器实例已配置，包含以下端点:")
	endpoints := []string{
		"GET  /api/users/:id",
		"POST /api/users",
		"GET  /api/users/search",
		"GET  /api/products/:id",
		"GET  /api/products/search",
		"POST /api/upload/avatar/:user_id",
		"GET  /api/admin/stats",
		"GET  /health",
	}

	for _, endpoint := range endpoints {
		fmt.Printf("   🔗 %s\n", endpoint)
	}

	fmt.Println("\n💡 使用建议:")
	fmt.Println("   1. 在生产环境中使用参数缓存以提高性能")
	fmt.Println("   2. 结合数据验证库进行完整的参数验证")
	fmt.Println("   3. 使用中间件进行统一的错误处理")
	fmt.Println("   4. 定期监控缓存统计信息")
}

// RunRealWorldDemo 运行实际应用演示
func RunRealWorldDemo() {
	fmt.Println("🔥 ParamBinder 实际应用示例")
	fmt.Println("===========================\n")

	// 运行演示
	DemonstrateRealWorldUsage()

	// 如果需要启动HTTP服务器，取消以下注释
	/*
		fmt.Println("\n🚀 启动HTTP服务器...")
		h := SetupRealWorldServer()

		fmt.Println("服务器将在 :8080 启动")
		fmt.Println("API文档:")
		fmt.Println("  用户管理: GET|POST /api/users")
		fmt.Println("  商品管理: GET /api/products")
		fmt.Println("  文件管理: POST /api/files")
		fmt.Println("  管理功能: GET /api/admin")
		fmt.Println("  健康检查: GET /health")

		h.Spin()
	*/
}

// main 函数演示
func main() {
	RunRealWorldDemo()
}
