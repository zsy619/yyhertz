package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/orm"
	"gorm.io/gorm"
)

// User 测试用户模型
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;size:100" json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Article 测试文章模型
type Article struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"user_id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Content     string    `gorm:"type:text" json:"content"`
	ViewCount   int       `json:"view_count"`
	PublishedAt time.Time `json:"published_at"`
}

// TestBasicORM 测试基本ORM功能
func TestBasicORM(t *testing.T) {
	// 1. 初始化数据库连接
	dbConfig := &orm.DatabaseConfig{
		Type:         "sqlite",
		Database:     "test.db",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		LogLevel:     "info",
		SlowQuery:    500,
	}

	// 创建ORM实例
	ormInstance, err := orm.NewORM(dbConfig)
	if err != nil {
		t.Fatal("Failed to initialize ORM:", err)
	}

	db := ormInstance.DB()

	// 2. 自动迁移
	if err := db.AutoMigrate(&User{}, &Article{}); err != nil {
		t.Fatal("Failed to migrate:", err)
	}

	// 3. 测试CRUD操作
	testCRUD(t, db)

	// 4. 测试分页
	testPagination(t, db)

	// 5. 测试事务
	testTransaction(t, db)

	fmt.Println("All tests passed!")
}

// testCRUD 测试CRUD操作
func testCRUD(t *testing.T, db *gorm.DB) {
	fmt.Println("\n=== Testing CRUD Operations ===")

	// Create
	user := User{
		Name:  "Test User",
		Email: "test@example.com",
		Age:   25,
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatal("Failed to create user:", err)
	}
	fmt.Printf("Created user with ID: %d\n", user.ID)

	// Read
	var foundUser User
	if err := db.First(&foundUser, user.ID).Error; err != nil {
		t.Fatal("Failed to find user:", err)
	}
	fmt.Printf("Found user: %+v\n", foundUser)

	// Update
	if err := db.Model(&foundUser).Update("age", 26).Error; err != nil {
		t.Fatal("Failed to update user:", err)
	}
	fmt.Println("Updated user age to 26")

	// Delete
	if err := db.Delete(&foundUser).Error; err != nil {
		t.Fatal("Failed to delete user:", err)
	}
	fmt.Println("Deleted user")
}

// testPagination 测试分页
func testPagination(t *testing.T, db *gorm.DB) {
	fmt.Println("\n=== Testing Pagination ===")

	// 创建测试数据
	for i := 1; i <= 20; i++ {
		user := User{
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   20 + i,
		}
		db.Create(&user)
	}

	// 使用偏移分页
	config := &orm.PaginationConfig{
		DefaultPageSize: 5,
		MaxPageSize: 100,
		IncludeTotal: true,
	}
	paginator := orm.NewOffsetPaginator(db.Model(&User{}), config)
	var users []User
	
	result, err := paginator.Paginate(1, 5, &users)
	if err != nil {
		t.Fatal("Pagination failed:", err)
	}

	fmt.Printf("Page %d of %d (Total: %d records)\n", 
		result.Pagination.CurrentPage, result.Pagination.TotalPages, result.Pagination.Total)
	fmt.Printf("Found %d users on page 1\n", len(users))

	// 测试第二页
	var users2 []User
	result2, err := paginator.Paginate(2, 5, &users2)
	if err != nil {
		t.Fatal("Pagination failed:", err)
	}
	fmt.Printf("Found %d users on page 2\n", len(users2))
	_ = result2 // 使用result2避免未使用的警告
}

// testTransaction 测试事务
func testTransaction(t *testing.T, db *gorm.DB) {
	fmt.Println("\n=== Testing Transaction ===")

	// 开始事务
	tx := db.Begin()
	
	// 创建用户
	user := User{
		Name:  "Transaction User",
		Email: "transaction@example.com",
		Age:   30,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		t.Fatal("Transaction failed:", err)
	}

	// 创建文章
	article := Article{
		UserID:      user.ID,
		Title:       "Transaction Article",
		Content:     "This article was created in a transaction",
		ViewCount:   0,
		PublishedAt: time.Now(),
	}

	if err := tx.Create(&article).Error; err != nil {
		tx.Rollback()
		t.Fatal("Transaction failed:", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Fatal("Failed to commit transaction:", err)
	}

	fmt.Printf("Transaction committed: User ID=%d, Article ID=%d\n", user.ID, article.ID)

	// 验证数据
	var count int64
	db.Model(&User{}).Where("email = ?", "transaction@example.com").Count(&count)
	if count != 1 {
		t.Fatal("Transaction data not found")
	}
	fmt.Println("Transaction data verified")
}

// TestConnectionPool 测试连接池
func TestConnectionPool(t *testing.T) {
	fmt.Println("\n=== Testing Connection Pool ===")

	// 配置连接池
	poolConfig := orm.DefaultPoolConfig()
	poolConfig.MaxIdleConns = 5
	poolConfig.MaxOpenConns = 10

	// 创建节点
	nodes := []*orm.DatabaseNode{
		{
			ID: "node1",
			Config: &orm.DatabaseConfig{
				Type:     "sqlite",
				Database: "test1.db",
			},
			IsMaster: true,
			Weight:   100,
		},
	}

	// 创建连接池
	pool, err := orm.NewMultiNodePool(poolConfig, nodes)
	if err != nil {
		t.Fatal("Failed to create connection pool:", err)
	}

	// 获取连接
	ctx := context.Background()
	conn, err := pool.GetConnection(ctx)
	if err != nil {
		t.Fatal("Failed to get connection:", err)
	}

	fmt.Printf("Got connection: %v\n", conn != nil)

	// 获取统计信息
	stats := pool.Stats()
	fmt.Printf("Pool stats: OpenConnections=%d, InUse=%d\n", 
		stats.OpenConnections, stats.InUse)

	// 关闭连接池
	if err := pool.Close(); err != nil {
		t.Fatal("Failed to close pool:", err)
	}
	fmt.Println("Connection pool closed")
}

// TestMultiDatabase 测试多数据库支持
func TestMultiDatabase(t *testing.T) {
	fmt.Println("\n=== Testing Multi-Database Support ===")

	// 获取驱动管理器
	driverManager := orm.GetDriverManager()
	
	// 列出支持的驱动
	drivers := driverManager.ListDrivers()
	fmt.Printf("Supported drivers: %v\n", drivers)

	// 测试不同的数据库类型
	databases := []struct {
		Type     string
		Expected bool
	}{
		{"mysql", true},
		{"postgres", true},
		{"sqlite", true},
		{"sqlserver", true},
		{"oracle", true},
		{"unknown", false},
	}

	for _, db := range databases {
		driver, err := driverManager.GetDriver(db.Type)
		if db.Expected {
			if err != nil {
				t.Errorf("Expected driver %s to be supported", db.Type)
			} else {
				fmt.Printf("Driver %s is supported\n", db.Type)
				_ = driver // 使用driver避免未使用的警告
			}
		} else {
			if err == nil {
				t.Errorf("Expected driver %s to be unsupported", db.Type)
			}
		}
	}
}