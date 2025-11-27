package main

import (
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/controller"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
)

// 用户控制器示例 - 统一使用BaseController + 启用优化特性
type OptimizedUserController struct {
	core.BaseController
}

// NewOptimizedUserController 创建用户控制器并启用优化特性
func NewOptimizedUserController() *OptimizedUserController {
	ctrl := &OptimizedUserController{}
	ctrl.EnableOptimization()
	ctrl.SetMiddleware([]string{"auth", "logging", "validation"})
	return ctrl
}

// 新的用户控制器示例 - 方式2：直接使用BaseController + 启用优化
type ModernUserController struct {
	core.BaseController
}

// NewModernUserController 创建现代用户控制器并启用优化
func NewModernUserController() *ModernUserController {
	controller := &ModernUserController{}
	// 启用优化特性
	controller.EnableOptimization()
	// 设置中间件
	controller.SetMiddleware([]string{"auth", "logging", "validation"})
	return controller
}

// 用户创建请求结构
type UserCreateRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50" binding:"required"`
	Email    string `json:"email" validate:"required,email" binding:"required"`
	Age      int    `json:"age" validate:"min=18,max=120"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"oneof=admin user guest" default:"user"`
}

// 用户更新请求结构
type UserUpdateRequest struct {
	Name  string `json:"name" validate:"min=2,max=50"`
	Email string `json:"email" validate:"email"`
	Age   int    `json:"age" validate:"min=18,max=120"`
}

// 用户响应结构
type UserResponse struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Age      int       `json:"age"`
	Role     string    `json:"role"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// GetIndex 获取用户列表 - 自动从查询参数绑定
func (uc *OptimizedUserController) GetIndex(page int, limit int, search string) ([]UserResponse, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	fmt.Printf("Getting users: page=%d, limit=%d, search=%s\n", page, limit, search)

	// 模拟数据库查询
	users := []UserResponse{
		{
			ID:       1,
			Name:     "张三",
			Email:    "zhangsan@example.com",
			Age:      25,
			Role:     "admin",
			Created:  time.Now().Add(-24 * time.Hour),
			Modified: time.Now(),
		},
		{
			ID:       2,
			Name:     "李四",
			Email:    "lisi@example.com",
			Age:      30,
			Role:     "user",
			Created:  time.Now().Add(-12 * time.Hour),
			Modified: time.Now(),
		},
	}

	// 过滤搜索结果
	if search != "" {
		var filtered []UserResponse
		for _, user := range users {
			if user.Name == search || user.Email == search {
				filtered = append(filtered, user)
			}
		}
		users = filtered
	}

	return users, nil
}

// GetShow 获取单个用户 - 从路径参数绑定
func (uc *OptimizedUserController) GetShow(id int64) (UserResponse, error) {
	fmt.Printf("Getting user with ID: %d\n", id)

	if id <= 0 {
		return UserResponse{}, fmt.Errorf("invalid user ID: %d", id)
	}

	// 模拟数据库查询
	user := UserResponse{
		ID:       id,
		Name:     "用户" + fmt.Sprintf("%d", id),
		Email:    fmt.Sprintf("user%d@example.com", id),
		Age:      25,
		Role:     "user",
		Created:  time.Now().Add(-24 * time.Hour),
		Modified: time.Now(),
	}

	return user, nil
}

// PostCreate 创建用户 - 从JSON体绑定和验证
func (uc *OptimizedUserController) PostCreate(req UserCreateRequest) (UserResponse, error) {
	fmt.Printf("Creating user: %+v\n", req)

	// 业务逻辑验证
	if req.Email == "admin@example.com" {
		return UserResponse{}, fmt.Errorf("email already exists")
	}

	// 模拟创建用户
	user := UserResponse{
		ID:       time.Now().Unix(),
		Name:     req.Name,
		Email:    req.Email,
		Age:      req.Age,
		Role:     req.Role,
		Created:  time.Now(),
		Modified: time.Now(),
	}

	return user, nil
}

// PutUpdate 更新用户 - 混合参数绑定
func (uc *OptimizedUserController) PutUpdate(id int64, req UserUpdateRequest) (UserResponse, error) {
	fmt.Printf("Updating user %d: %+v\n", id, req)

	if id <= 0 {
		return UserResponse{}, fmt.Errorf("invalid user ID: %d", id)
	}

	// 模拟更新用户
	user := UserResponse{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Age:      req.Age,
		Role:     "user",
		Created:  time.Now().Add(-24 * time.Hour),
		Modified: time.Now(),
	}

	return user, nil
}

// DeleteRemove 删除用户
func (uc *OptimizedUserController) DeleteRemove(id int64) error {
	fmt.Printf("Deleting user with ID: %d\n", id)

	if id <= 0 {
		return fmt.Errorf("invalid user ID: %d", id)
	}

	// 模拟删除用户
	fmt.Printf("User %d deleted successfully\n", id)
	return nil
}

// GetControllerName 实现OptimizedController接口
func (uc *OptimizedUserController) GetControllerName() string {
	return "OptimizedUserController"
}

// GetMethodMapping 获取方法映射
func (uc *OptimizedUserController) GetMethodMapping() map[string]string {
	return map[string]string{
		"GetIndex":     "GET:/users",
		"GetShow":      "GET:/users/{id}",
		"PostCreate":   "POST:/users",
		"PutUpdate":    "PUT:/users/{id}",
		"DeleteRemove": "DELETE:/users/{id}",
	}
}

// GetMiddleware 获取中间件
func (uc *OptimizedUserController) GetMiddleware() []string {
	return []string{"auth", "logging", "validation"}
}

// 产品控制器示例 - 统一使用BaseController + 启用优化特性
type OptimizedProductController struct {
	core.BaseController
}

// NewOptimizedProductController 创建产品控制器并启用优化特性
func NewOptimizedProductController() *OptimizedProductController {
	ctrl := &OptimizedProductController{}
	ctrl.EnableOptimization()
	ctrl.SetMiddleware([]string{"auth", "logging", "rateLimit"})
	return ctrl
}

type ProductCreateRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Category    string  `json:"category" validate:"required,oneof=electronics books clothing"`
	InStock     bool    `json:"in_stock" default:"true"`
}

type ProductResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	InStock     bool      `json:"in_stock"`
	Created     time.Time `json:"created"`
}

func (pc *OptimizedProductController) GetIndex(category string, minPrice float64, maxPrice float64) ([]ProductResponse, error) {
	fmt.Printf("Getting products: category=%s, minPrice=%.2f, maxPrice=%.2f\n", category, minPrice, maxPrice)

	products := []ProductResponse{
		{
			ID:          1,
			Name:        "iPhone 15",
			Description: "Latest iPhone model",
			Price:       999.99,
			Category:    "electronics",
			InStock:     true,
			Created:     time.Now().Add(-48 * time.Hour),
		},
		{
			ID:          2,
			Name:        "Programming Book",
			Description: "Learn Go programming",
			Price:       59.99,
			Category:    "books",
			InStock:     true,
			Created:     time.Now().Add(-24 * time.Hour),
		},
	}

	// 过滤产品
	var filtered []ProductResponse
	for _, product := range products {
		if category != "" && product.Category != category {
			continue
		}
		if minPrice > 0 && product.Price < minPrice {
			continue
		}
		if maxPrice > 0 && product.Price > maxPrice {
			continue
		}
		filtered = append(filtered, product)
	}

	return filtered, nil
}

func (pc *OptimizedProductController) PostCreate(req ProductCreateRequest) (ProductResponse, error) {
	fmt.Printf("Creating product: %+v\n", req)

	product := ProductResponse{
		ID:          time.Now().Unix(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		InStock:     req.InStock,
		Created:     time.Time{},
	}

	return product, nil
}

func main() {
	fmt.Println("🚀 启动优化的MVC示例应用...")

	// 创建优化的控制器管理器
	config := &controller.CompilerConfig{
		EnableCache:     true,
		CacheSize:       1000,
		PrecompileAll:   true,
		OptimizeLevel:   3,
		EnableLifecycle: true,
		PoolSize:        100,
		MaxIdleTime:     30 * time.Minute,
	}

	manager := controller.NewOptimizedControllerManager(config)
	manager.RegisterLifecycleHooks()

	// 注册控制器
	fmt.Println("📋 注册控制器...")

	// 使用统一的BaseController + 启用优化特性
	userController := NewOptimizedUserController()
	if err := manager.RegisterController(userController); err != nil {
		log.Fatalf("Failed to register user controller: %v", err)
	}

	productController := NewOptimizedProductController()
	if err := manager.RegisterController(productController); err != nil {
		log.Fatalf("Failed to register product controller: %v", err)
	}

	// 展示另一种创建方式
	modernController := NewModernUserController()
	if err := manager.RegisterController(modernController); err != nil {
		log.Fatalf("Failed to register modern controller: %v", err)
	}

	// 预编译所有控制器
	fmt.Println("⚡ 预编译控制器...")
	if err := manager.PrecompileAll(); err != nil {
		log.Fatalf("Failed to precompile controllers: %v", err)
	}

	// 缓存预热
	fmt.Println("🔥 缓存预热...")
	if err := manager.WarmupCache(); err != nil {
		log.Printf("Cache warmup failed: %v", err)
	}

	// 创建MVC应用
	app := mvc.HertzApp

	// 添加中间件
	app.Use(
		middleware.RecoveryMiddleware(),
		middleware.TracingMiddleware(),
		middleware.CORSMiddleware(),
	)

	fmt.Println("✅ 中间件配置已跳过（演示用）")

	// 模拟一些请求来展示优化效果
	fmt.Println("\n🧪 模拟请求测试...")

	// 创建模拟上下文
	testRequests := []struct {
		controller string
		method     string
		desc       string
	}{
		{"OptimizedUserController", "GetIndex", "获取用户列表"},
		{"OptimizedUserController", "GetShow", "获取单个用户"},
		{"OptimizedUserController", "PostCreate", "创建用户"},
		{"OptimizedProductController", "GetIndex", "获取产品列表"},
		{"OptimizedProductController", "PostCreate", "创建产品"},
	}

	// 执行测试请求
	for i, req := range testRequests {
		fmt.Printf("\n%d. %s\n", i+1, req.desc)

		ctx := &context.Context{}

		start := time.Now()
		err := manager.HandleRequest(ctx, req.controller, req.method)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("   ❌ 请求失败: %v (耗时: %v)\n", err, duration)
		} else {
			fmt.Printf("   ✅ 请求成功 (耗时: %v)\n", duration)
		}
	}

	// 性能压测示例
	fmt.Println("\n🚀 性能压测...")

	testConcurrentRequests(manager, 1000, 10)

	// 显示统计信息
	fmt.Println("\n📊 性能统计:")
	stats := manager.GetDetailedStats()
	printStats(stats)

	// 内存优化
	fmt.Println("\n🧹 内存优化...")
	manager.OptimizeMemory()

	fmt.Println("\n✅ 示例完成")
}

// testConcurrentRequests 并发请求测试
func testConcurrentRequests(manager *controller.OptimizedControllerManager, requestCount int, concurrency int) {
	start := time.Now()

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, concurrency)
	results := make(chan error, requestCount)

	// 发起并发请求
	for i := 0; i < requestCount; i++ {
		go func(reqID int) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			ctx := &context.Context{}

			err := manager.HandleRequest(ctx, "OptimizedUserController", "GetIndex")
			results <- err
		}(i)
	}

	// 等待所有请求完成
	var errors []error
	for i := 0; i < requestCount; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	duration := time.Since(start)
	successCount := requestCount - len(errors)

	fmt.Printf("   📈 总请求: %d\n", requestCount)
	fmt.Printf("   ✅ 成功: %d\n", successCount)
	fmt.Printf("   ❌ 失败: %d\n", len(errors))
	fmt.Printf("   ⏱️  总耗时: %v\n", duration)
	fmt.Printf("   🚀 吞吐量: %.0f 请求/秒\n", float64(requestCount)/duration.Seconds())
	fmt.Printf("   ⚡ 平均延迟: %v\n", duration/time.Duration(requestCount))
}

// printStats 打印统计信息
func printStats(stats map[string]interface{}) {
	if perfStats, ok := stats["performance"]; ok {
		if ps, ok := perfStats.(*controller.PerformanceStats); ok {
			fmt.Printf("   📊 总请求数: %d\n", ps.TotalRequests)
			fmt.Printf("   ⏱️  平均响应时间: %v\n", ps.AverageResponseTime)
			fmt.Printf("   🎯 缓存命中率: %.2f%%\n", ps.CacheHitRate*100)
			fmt.Printf("   🏗️  编译时间: %v\n", ps.CompilationTime)
			fmt.Printf("   🎮 控制器实例: %d\n", ps.ControllerInstances)
		}
	}

	if compilerStats, ok := stats["compiler"]; ok {
		if cs, ok := compilerStats.(*controller.CompilerStats); ok {
			fmt.Printf("   📝 编译的控制器: %d\n", cs.CompiledControllers)
			fmt.Printf("   🔧 编译的方法: %d\n", cs.CompiledMethods)
		}
	}

	if lifecycleStats, ok := stats["lifecycle"]; ok {
		if ls, ok := lifecycleStats.(*controller.LifecycleMetrics); ok {
			fmt.Printf("   🆕 创建数量: %d\n", ls.CreatedCount)
			fmt.Printf("   🗑️  销毁数量: %d\n", ls.DestroyedCount)
			fmt.Printf("   🔄 活跃数量: %d\n", ls.ActiveCount)
		}
	}
}
