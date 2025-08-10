package main

import (
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/orm"
)

// ================== 模型定义 ==================

// User 用户模型
type User struct {
	orm.BaseModel
	Username string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email    string    `gorm:"size:100;uniqueIndex" json:"email"`
	Password string    `gorm:"size:255" json:"-"`
	Status   string    `gorm:"size:20;default:'active'" json:"status"`
	Profile  *Profile  `json:"profile,omitempty"`
	Articles []Article `json:"articles,omitempty"`
	Orders   []Order   `json:"orders,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// Profile 用户资料模型
type Profile struct {
	orm.BaseModel
	UserID    uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	FirstName string `gorm:"size:50" json:"first_name"`
	LastName  string `gorm:"size:50" json:"last_name"`
	Avatar    string `gorm:"size:255" json:"avatar"`
	Bio       string `gorm:"type:text" json:"bio"`
	Birthday  *time.Time `json:"birthday"`
}

// Article 文章模型
type Article struct {
	orm.BaseModel
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Slug        string    `gorm:"size:250;uniqueIndex" json:"slug"`
	Content     string    `gorm:"type:longtext" json:"content"`
	Summary     string    `gorm:"type:text" json:"summary"`
	Status      string    `gorm:"size:20;default:'draft'" json:"status"`
	ViewCount   int       `gorm:"default:0" json:"view_count"`
	PublishedAt *time.Time `json:"published_at"`
	Tags        []Tag     `gorm:"many2many:article_tags;" json:"tags,omitempty"`
}

// Tag 标签模型
type Tag struct {
	orm.BaseModel
	Name        string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Color       string    `gorm:"size:7;default:'#007bff'" json:"color"`
	Description string    `gorm:"type:text" json:"description"`
	Articles    []Article `gorm:"many2many:article_tags;" json:"articles,omitempty"`
}

// Order 订单模型
type Order struct {
	orm.BaseModel
	UserID      uint        `gorm:"not null;index" json:"user_id"`
	OrderNo     string      `gorm:"size:32;uniqueIndex;not null" json:"order_no"`
	Amount      float64     `gorm:"type:decimal(10,2)" json:"amount"`
	Status      string      `gorm:"size:20;default:'pending'" json:"status"`
	PaymentType string      `gorm:"size:20" json:"payment_type"`
	PaidAt      *time.Time  `json:"paid_at"`
	Items       []OrderItem `json:"items,omitempty"`
}

// OrderItem 订单项模型
type OrderItem struct {
	orm.BaseModel
	OrderID     uint    `gorm:"not null;index" json:"order_id"`
	ProductName string  `gorm:"size:200;not null" json:"product_name"`
	Price       float64 `gorm:"type:decimal(8,2)" json:"price"`
	Quantity    int     `gorm:"not null" json:"quantity"`
	Total       float64 `gorm:"type:decimal(10,2)" json:"total"`
}

// ================== 示例函数 ==================

// CompleteORMExample 完整的ORM功能演示
func CompleteORMExample() {
	fmt.Println("=== YYHertz ORM 完整功能演示 ===")

	// 1. 初始化数据库
	ormInstance, err := initDatabase()
	if err != nil {
		log.Fatal("初始化数据库失败:", err)
	}

	// 2. 基础 CRUD 操作演示
	demonstrateCRUDOperations(ormInstance)

	// 3. 高级查询演示
	demonstrateAdvancedQueries(ormInstance)

	// 4. 关联查询演示
	demonstrateRelationQueries(ormInstance)

	// 5. 事务操作演示
	demonstrateTransactions(ormInstance)

	// 6. 分页查询演示
	demonstratePagination(ormInstance)

	// 7. 缓存操作演示
	demonstrateCaching(ormInstance)

	// 8. 性能监控演示
	demonstrateMonitoring(ormInstance)

	fmt.Println("\n✅ 所有演示完成!")
}

// initDatabase 初始化数据库
func initDatabase() (*orm.ORM, error) {
	fmt.Println("1. 初始化数据库...")

	// 创建数据库配置
	config := &orm.DatabaseConfig{
		Type:         "sqlite",
		Database:     "complete_example.db",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		LogLevel:     "info",
		SlowQuery:    200, // 200ms慢查询阈值
	}

	// 创建 ORM 实例
	ormInstance, err := orm.NewORM(config)
	if err != nil {
		return nil, fmt.Errorf("创建ORM实例失败: %w", err)
	}

	// 自动迁移
	models := []interface{}{
		&User{}, &Profile{}, &Article{}, &Tag{}, &Order{}, &OrderItem{},
	}

	for _, model := range models {
		if err := ormInstance.AutoMigrate(model); err != nil {
			return nil, fmt.Errorf("迁移模型失败: %w", err)
		}
	}

	fmt.Println("✅ 数据库初始化完成")
	return ormInstance, nil
}

// demonstrateCRUDOperations 演示基础CRUD操作
func demonstrateCRUDOperations(ormInstance *orm.ORM) {
	fmt.Println("2. 基础 CRUD 操作演示...")
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	// 创建用户
	users := []*User{
		{Username: "alice", Email: "alice@example.com", Password: "password123", Status: "active"},
		{Username: "bob", Email: "bob@example.com", Password: "password123", Status: "active"},
		{Username: "charlie", Email: "charlie@example.com", Password: "password123", Status: "inactive"},
	}

	for _, user := range users {
		if err := crud.Create(user); err != nil {
			fmt.Printf("❌ 创建用户失败: %v\n", err)
			continue
		}
		fmt.Printf("✅ 创建用户: %s (ID: %d)\n", user.Username, user.ID)
	}

	// 创建用户资料
	profiles := []*Profile{
		{UserID: users[0].ID, FirstName: "Alice", LastName: "Smith", Bio: "Frontend Developer"},
		{UserID: users[1].ID, FirstName: "Bob", LastName: "Johnson", Bio: "Backend Developer"},
	}

	for _, profile := range profiles {
		if err := crud.Create(profile); err != nil {
			fmt.Printf("❌ 创建资料失败: %v\n", err)
			continue
		}
		fmt.Printf("✅ 创建资料: %s %s\n", profile.FirstName, profile.LastName)
	}

	// 读取操作
	var foundUser User
	if err := crud.First(&foundUser, "username = ?", "alice"); err != nil {
		fmt.Printf("❌ 查询用户失败: %v\n", err)
	} else {
		fmt.Printf("✅ 查询到用户: %s\n", foundUser.Username)
	}

	// 更新操作
	if err := crud.Update(&foundUser, "status", "premium"); err != nil {
		fmt.Printf("❌ 更新用户失败: %v\n", err)
	} else {
		fmt.Printf("✅ 更新用户状态: %s\n", foundUser.Status)
	}

	// 统计操作
	count, err := crud.Count(&User{})
	if err != nil {
		fmt.Printf("❌ 统计用户失败: %v\n", err)
	} else {
		fmt.Printf("✅ 用户总数: %d\n", count)
	}

	fmt.Println("")
}

// demonstrateAdvancedQueries 演示高级查询
func demonstrateAdvancedQueries(ormInstance *orm.ORM) {
	fmt.Println("3. 高级查询演示...")
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	// 链式查询
	var activeUsers []User
	err := crud.Where("status = ?", "active").
		OrderBy("created_at", "desc").
		Limit(10).
		Find(&activeUsers)

	if err != nil {
		fmt.Printf("❌ 链式查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 活跃用户数量: %d\n", len(activeUsers))
	}

	// IN查询
	var users []User
	err = crud.WhereIn("status", []string{"active", "premium"}).Find(&users)
	if err != nil {
		fmt.Printf("❌ IN查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 活跃和高级用户数量: %d\n", len(users))
	}

	// LIKE查询
	var searchUsers []User
	err = crud.WhereLike("username", "a").Find(&searchUsers)
	if err != nil {
		fmt.Printf("❌ LIKE查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 用户名包含'a'的用户数量: %d\n", len(searchUsers))
	}

	// 时间范围查询
	yesterday := time.Now().AddDate(0, 0, -1)
	var recentUsers []User
	err = crud.WhereBetween("created_at", yesterday, time.Now()).Find(&recentUsers)
	if err != nil {
		fmt.Printf("❌ 时间范围查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 最近创建的用户数量: %d\n", len(recentUsers))
	}

	fmt.Println("")
}

// demonstrateRelationQueries 演示关联查询
func demonstrateRelationQueries(ormInstance *orm.ORM) {
	fmt.Println("4. 关联查询演示...")

	// 创建文章和标签数据
	setupArticlesAndTags(ormInstance)

	// 使用查询构建器进行关联查询
	query := orm.NewQueryBuilder[User](ormInstance.DB())
	users, err := query.Preload("Profile", "Articles", "Orders").Find()

	if err != nil {
		fmt.Printf("❌ 关联查询失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 查询到 %d 个用户及其关联数据\n", len(users))
	for _, user := range users {
		fmt.Printf("  用户: %s\n", user.Username)
		if user.Profile != nil {
			fmt.Printf("    资料: %s %s\n", user.Profile.FirstName, user.Profile.LastName)
		}
		fmt.Printf("    文章数: %d\n", len(user.Articles))
		fmt.Printf("    订单数: %d\n", len(user.Orders))
	}

	fmt.Println("")
}

// demonstrateTransactions 演示事务操作
func demonstrateTransactions(ormInstance *orm.ORM) {
	fmt.Println("5. 事务操作演示...")

	// 手动开始事务
	db := ormInstance.DB()
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err := func() error {
		// 创建新用户
		user := &User{
			Username: "transaction_user",
			Email:    "transaction@example.com",
			Password: "password123",
			Status:   "active",
		}

		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}

		// 创建订单
		order := &Order{
			UserID:      user.ID,
			OrderNo:     fmt.Sprintf("ORD%d%d", time.Now().Unix(), user.ID),
			Amount:      199.99,
			Status:      "pending",
			PaymentType: "credit_card",
		}

		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 创建订单项
		items := []*OrderItem{
			{OrderID: order.ID, ProductName: "产品A", Price: 99.99, Quantity: 1, Total: 99.99},
			{OrderID: order.ID, ProductName: "产品B", Price: 100.00, Quantity: 1, Total: 100.00},
		}

		for _, item := range items {
			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("创建订单项失败: %w", err)
			}
		}

		fmt.Printf("✅ 事务成功: 用户 %s, 订单 %s\n", user.Username, order.OrderNo)
		return nil
	}()

	if err != nil {
		tx.Rollback()
		fmt.Printf("❌ 事务失败: %v\n", err)
	} else {
		tx.Commit()
	}


	fmt.Println("")
}

// demonstratePagination 演示分页查询
func demonstratePagination(ormInstance *orm.ORM) {
	fmt.Println("6. 分页查询演示...")

	// 创建更多测试用户
	crud := orm.NewSimpleCRUD(ormInstance.DB())
	for i := 1; i <= 25; i++ {
		user := &User{
			Username: fmt.Sprintf("user%02d", i),
			Email:    fmt.Sprintf("user%02d@example.com", i),
			Password: "password123",
			Status:   []string{"active", "inactive", "premium"}[i%3],
		}
		crud.Create(user)
	}

	// 使用简单分页方法
	db := ormInstance.DB()
	
	page, pageSize := 1, 5
	var users []User
	var total int64

	// 计算总数
	db.Model(&User{}).Count(&total)
	
	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Limit(pageSize).Offset(offset).Find(&users).Error

	if err != nil {
		fmt.Printf("❌ 分页查询失败: %v\n", err)
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	fmt.Printf("✅ 分页结果: 第%d页/共%d页, 总记录数:%d\n", page, totalPages, total)
	
	for i, user := range users {
		fmt.Printf("  %d. %s (%s)\n", i+1, user.Username, user.Status)
	}

	// 条件分页查询
	var activeUsers []User
	var activeTotal int64
	db.Model(&User{}).Where("status = ?", "active").Count(&activeTotal)
	err = db.Where("status = ?", "active").Limit(5).Find(&activeUsers).Error
	if err != nil {
		fmt.Printf("❌ 条件分页查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 活跃用户分页: %d条记录\n", activeTotal)
		for i, user := range activeUsers {
			fmt.Printf("  %d. %s\n", i+1, user.Username)
		}
	}

	fmt.Println("")
}

// demonstrateCaching 演示缓存操作
func demonstrateCaching(ormInstance *orm.ORM) {
	fmt.Println("7. 缓存操作演示...")

	db := ormInstance.DB()
	cachedQuery := orm.NewCachedQuery(db).
		WithCacheKey("active_users_demo").
		WithTTL(2 * time.Minute)

	// 第一次查询（从数据库）
	start := time.Now()
	var users []User
	err := cachedQuery.Find(&users, "status = ?", "active")
	duration1 := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 缓存查询失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 第一次查询: %d条记录, 耗时: %v\n", len(users), duration1)

	// 第二次查询（从缓存）
	start = time.Now()
	var cachedUsers []User
	err = cachedQuery.Find(&cachedUsers, "status = ?", "active")
	duration2 := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 缓存查询失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 第二次查询（缓存): %d条记录, 耗时: %v\n", len(cachedUsers), duration2)
	fmt.Printf("✅ 性能提升: %.2fx\n", float64(duration1)/float64(duration2))

	fmt.Println("")
}

// demonstrateMonitoring 演示性能监控
func demonstrateMonitoring(ormInstance *orm.ORM) {
	fmt.Println("8. 性能监控演示...")

	// 连接池统计
	sqlDB, _ := ormInstance.DB().DB()
	stats := sqlDB.Stats()
	fmt.Printf("✅ 连接池统计:\n")
	fmt.Printf("  最大连接数: %d\n", stats.MaxOpenConnections)
	fmt.Printf("  活跃连接数: %d\n", stats.InUse)
	fmt.Printf("  空闲连接数: %d\n", stats.Idle)
	fmt.Printf("  等待次数: %d\n", stats.WaitCount)
	fmt.Printf("  等待时长: %v\n", stats.WaitDuration)

	// 性能监控器
	monitor := orm.NewPerformanceMonitor(100 * time.Millisecond)
	
	// 执行一些查询来生成监控数据
	crud := orm.NewSimpleCRUD(ormInstance.DB())
	for i := 0; i < 10; i++ {
		start := time.Now()
		var users []User
		crud.Find(&users, "status = ?", "active")
		monitor.RecordQuery(time.Since(start))
	}

	// 获取性能统计
	performanceStats := monitor.GetStats()
	fmt.Printf("✅ 性能统计:\n")
	fmt.Printf("  总查询次数: %v\n", performanceStats["total_queries"])
	fmt.Printf("  慢查询次数: %v\n", performanceStats["slow_queries"])
	fmt.Printf("  慢查询比例: %.2f%%\n", performanceStats["slow_query_rate"])
	fmt.Printf("  慢查询阈值: %v\n", performanceStats["threshold"])

	fmt.Println("")
}

// setupArticlesAndTags 设置文章和标签数据
func setupArticlesAndTags(ormInstance *orm.ORM) {
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	// 创建标签
	tags := []*Tag{
		{Name: "Go", Color: "#00ADD8", Description: "Go programming language"},
		{Name: "Database", Color: "#336791", Description: "Database related topics"},
		{Name: "Web", Color: "#61DAFB", Description: "Web development"},
	}

	for _, tag := range tags {
		crud.Create(tag)
	}

	// 获取用户
	var users []User
	crud.Find(&users)
	if len(users) == 0 {
		return
	}

	// 创建文章
	articles := []*Article{
		{
			UserID:      users[0].ID,
			Title:       "Go语言入门指南",
			Slug:        "go-getting-started",
			Content:     "这是一篇关于Go语言入门的文章...",
			Summary:     "Go语言基础教程",
			Status:      "published",
			ViewCount:   150,
			PublishedAt: &time.Time{},
		},
		{
			UserID:      users[0].ID,
			Title:       "数据库设计最佳实践",
			Slug:        "database-best-practices",
			Content:     "这是一篇关于数据库设计的文章...",
			Summary:     "数据库设计指南",
			Status:      "published",
			ViewCount:   200,
			PublishedAt: &time.Time{},
		},
	}

	for _, article := range articles {
		now := time.Now()
		article.PublishedAt = &now
		crud.Create(article)
	}

	// 创建订单
	orders := []*Order{
		{
			UserID:      users[0].ID,
			OrderNo:     "ORD001",
			Amount:      299.99,
			Status:      "paid",
			PaymentType: "credit_card",
		},
		{
			UserID:      users[1].ID,
			OrderNo:     "ORD002",
			Amount:      599.99,
			Status:      "pending",
			PaymentType: "paypal",
		},
	}

	for _, order := range orders {
		if order.Status == "paid" {
			now := time.Now()
			order.PaidAt = &now
		}
		crud.Create(order)
	}
}

func main() {
	CompleteORMExample()
}