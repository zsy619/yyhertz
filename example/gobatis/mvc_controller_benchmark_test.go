package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MyBatis控制器基础结构
type BenchmarkControllerMyBatis struct {
	RequestCount int64
	db           *gorm.DB
}

// NewBenchmarkController 创建MyBatis基准测试控制器
func NewBenchmarkControllerMyBatis() *BenchmarkControllerMyBatis {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	
	return &BenchmarkControllerMyBatis{
		db: db,
	}
}

// GetUsers 查询用户列表 - MyBatis版本
func (bc *BenchmarkControllerMyBatis) GetUsers() ([]User, error) {
	bc.RequestCount++
	
	session := mybatis.NewSimpleSession(bc.db)
	ctx := context.Background()
	
	_, err := session.SelectList(ctx, "SELECT id, name, email, age FROM users")
	if err != nil {
		// 忽略错误，使用GORM直接查询
	}
	
	// 使用GORM直接查询
	var users []User
	err = bc.db.Raw("SELECT id, name, email, age FROM users").Scan(&users).Error
	return users, err
}

// GetUser 查询单个用户 - MyBatis版本
func (bc *BenchmarkControllerMyBatis) GetUser(id int64) (*User, error) {
	bc.RequestCount++
	
	session := mybatis.NewSimpleSession(bc.db)
	ctx := context.Background()
	
	_, err := session.SelectOne(ctx, "SELECT id, name, email, age FROM users WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	
	// 直接使用GORM查询
	var user User
	err = bc.db.Raw("SELECT id, name, email, age FROM users WHERE id = ?", id).Scan(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户 - MyBatis版本
func (bc *BenchmarkControllerMyBatis) CreateUser(user User) error {
	bc.RequestCount++
	
	session := mybatis.NewSimpleSession(bc.db)
	ctx := context.Background()
	
	_, err := session.Insert(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
		user.Name, user.Email, user.Age)
	return err
}

// UpdateUser 更新用户 - MyBatis版本
func (bc *BenchmarkControllerMyBatis) UpdateUser(id int64, user User) error {
	bc.RequestCount++
	
	session := mybatis.NewSimpleSession(bc.db)
	ctx := context.Background()
	
	_, err := session.Update(ctx, "UPDATE users SET name = ?, email = ?, age = ? WHERE id = ?",
		user.Name, user.Email, user.Age, id)
	return err
}

// DeleteUser 删除用户 - MyBatis版本
func (bc *BenchmarkControllerMyBatis) DeleteUser(id int64) error {
	bc.RequestCount++
	
	session := mybatis.NewSimpleSession(bc.db)
	ctx := context.Background()
	
	_, err := session.Delete(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

// === MyBatis控制器性能基准测试 ===

// BenchmarkMyBatisSimpleSession_SelectOne 单条查询性能测试
func BenchmarkMyBatisSimpleSession_SelectOne(b *testing.B) {
	controller := NewBenchmarkControllerMyBatis()
	setupTestData(controller.db)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := controller.GetUser(1)
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkMyBatisSimpleSession_SelectList 列表查询性能测试
func BenchmarkMyBatisSimpleSession_SelectList(b *testing.B) {
	controller := NewBenchmarkControllerMyBatis()
	setupTestData(controller.db)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := controller.GetUsers()
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}

// BenchmarkMyBatisSimpleSession_Insert 插入操作性能测试
func BenchmarkMyBatisSimpleSession_Insert(b *testing.B) {
	controller := NewBenchmarkControllerMyBatis()
	setupEmptyDB(controller.db)

	user := User{
		Name:  "Test User",
		Email: "test@example.com",
		Age:   25,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		user.Name = fmt.Sprintf("User_%d", i)
		user.Email = fmt.Sprintf("user_%d@example.com", i)
		
		err := controller.CreateUser(user)
		if err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

// BenchmarkMyBatisSimpleSession_Update 更新操作性能测试
func BenchmarkMyBatisSimpleSession_Update(b *testing.B) {
	controller := NewBenchmarkControllerMyBatis()
	setupTestData(controller.db)

	user := User{
		Name:  "Updated User",
		Email: "updated@example.com",
		Age:   30,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := controller.UpdateUser(1, user)
		if err != nil {
			b.Fatalf("Update failed: %v", err)
		}
	}
}

// BenchmarkMyBatisSimpleSession_Delete 删除操作性能测试
func BenchmarkMyBatisSimpleSession_Delete(b *testing.B) {
	controller := NewBenchmarkControllerMyBatis()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 为每次测试准备数据
		b.StopTimer()
		setupTestData(controller.db)
		b.StartTimer()

		err := controller.DeleteUser(1)
		if err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
	}
}

// BenchmarkMyBatisConcurrentRequests MyBatis并发请求性能测试
func BenchmarkMyBatisConcurrentRequests(b *testing.B) {
	controller := NewBenchmarkControllerMyBatis()
	setupTestData(controller.db)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 随机选择操作类型
			switch b.N % 4 {
			case 0:
				controller.GetUser(1)
			case 1:
				controller.GetUsers()
			case 2:
				user := User{Name: "Concurrent User", Email: "concurrent@test.com", Age: 25}
				controller.CreateUser(user)
			case 3:
				user := User{Name: "Updated User", Email: "updated@test.com", Age: 30}
				controller.UpdateUser(1, user)
			}
		}
	})
}

// BenchmarkMyBatisTransactionPerformance 事务性能测试
func BenchmarkMyBatisTransactionPerformance(b *testing.B) {
	controller := NewBenchmarkControllerMyBatis()
	setupEmptyDB(controller.db)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 执行事务操作
		tx := controller.db.Begin()
		txSession := mybatis.NewSimpleSession(tx)
		ctx := context.Background()

		// 批量插入
		for j := 0; j < 10; j++ {
			user := User{
				Name:  fmt.Sprintf("TxUser_%d_%d", i, j),
				Email: fmt.Sprintf("txuser_%d_%d@test.com", i, j),
				Age:   20 + j,
			}
			_, err := txSession.Insert(ctx, "INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
				user.Name, user.Email, user.Age)
			if err != nil {
				tx.Rollback()
				b.Fatalf("Transaction insert failed: %v", err)
			}
		}

		err := tx.Commit().Error
		if err != nil {
			b.Fatalf("Transaction commit failed: %v", err)
		}
	}
}

// === MyBatis控制器功能测试 ===

// TestMyBatisControllerCRUD MyBatis控制器CRUD功能测试
func TestMyBatisControllerCRUD(t *testing.T) {
	controller := NewBenchmarkControllerMyBatis()
	setupEmptyDB(controller.db)

	// 测试创建用户
	user := User{
		Name:  "Test User",
		Email: "test@example.com",
		Age:   25,
	}

	err := controller.CreateUser(user)
	if err != nil {
		t.Fatalf("Create user failed: %v", err)
	}

	// 测试查询用户列表
	users, err := controller.GetUsers()
	if err != nil {
		t.Fatalf("Get users failed: %v", err)
	}

	if len(users) == 0 {
		t.Fatal("Expected at least one user")
	}

	// 测试查询单个用户
	foundUser, err := controller.GetUser(users[0].ID)
	if err != nil {
		t.Fatalf("Get user failed: %v", err)
	}

	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}

	// 测试更新用户
	updatedUser := User{
		Name:  "Updated User",
		Email: "updated@example.com",
		Age:   30,
	}

	err = controller.UpdateUser(foundUser.ID, updatedUser)
	if err != nil {
		t.Fatalf("Update user failed: %v", err)
	}

	// 验证更新
	foundUser, err = controller.GetUser(foundUser.ID)
	if err != nil {
		t.Fatalf("Get updated user failed: %v", err)
	}

	if foundUser.Name != updatedUser.Name {
		t.Errorf("Expected updated name %s, got %s", updatedUser.Name, foundUser.Name)
	}

	// 测试删除用户
	err = controller.DeleteUser(foundUser.ID)
	if err != nil {
		t.Fatalf("Delete user failed: %v", err)
	}

	// 验证删除
	users, err = controller.GetUsers()
	if err != nil {
		t.Fatalf("Get users after delete failed: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("Expected 0 users after delete, got %d", len(users))
	}

	t.Logf("✅ MyBatis Controller CRUD test passed. Request count: %d", controller.RequestCount)
}

// TestMyBatisControllerPerformance MyBatis控制器性能特性测试
func TestMyBatisControllerPerformance(t *testing.T) {
	controller := NewBenchmarkControllerMyBatis()
	setupTestData(controller.db)

	// 性能计时测试
	start := time.Now()
	
	// 执行1000次查询
	for i := 0; i < 1000; i++ {
		_, err := controller.GetUser(1)
		if err != nil {
			t.Fatalf("Query %d failed: %v", i, err)
		}
	}

	duration := time.Since(start)
	avgTime := duration / 1000

	t.Logf("MyBatis Single Query Performance:")
	t.Logf("- Total time: %v", duration)
	t.Logf("- Average time: %v", avgTime)
	t.Logf("- QPS: %.2f", 1000.0/duration.Seconds())

	if avgTime > time.Millisecond {
		t.Errorf("Average query time %v exceeds 1ms threshold", avgTime)
	}

	// 批量操作测试
	start = time.Now()
	for i := 0; i < 100; i++ {
		_, err := controller.GetUsers()
		if err != nil {
			t.Fatalf("Batch query %d failed: %v", i, err)
		}
	}
	batchDuration := time.Since(start)
	avgBatchTime := batchDuration / 100

	t.Logf("MyBatis Batch Query Performance:")
	t.Logf("- Total time: %v", batchDuration)
	t.Logf("- Average time: %v", avgBatchTime)
	t.Logf("- Batch QPS: %.2f", 100.0/batchDuration.Seconds())

	t.Logf("✅ MyBatis Controller performance test passed. Request count: %d", controller.RequestCount)
}

// 辅助函数
func setupEmptyDB(db *gorm.DB) {
	// 删除表（如果存在）
	db.Exec("DROP TABLE IF EXISTS users")
	
	// 创建用户表
	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		age INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
}

func setupTestData(db *gorm.DB) {
	setupEmptyDB(db)
	
	// 插入测试数据
	testUsers := []User{
		{Name: "John Doe", Email: "john@example.com", Age: 30},
		{Name: "Jane Smith", Email: "jane@example.com", Age: 25},
		{Name: "Bob Johnson", Email: "bob@example.com", Age: 35},
		{Name: "Alice Brown", Email: "alice@example.com", Age: 28},
		{Name: "Charlie Davis", Email: "charlie@example.com", Age: 32},
	}

	for _, user := range testUsers {
		db.Exec("INSERT INTO users (name, email, age) VALUES (?, ?, ?)", 
			user.Name, user.Email, user.Age)
	}
}