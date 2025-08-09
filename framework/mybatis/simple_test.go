// Package mybatis 简化版功能测试
package mybatis

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 设置测试数据库
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	
	// 创建测试表
	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		create_at DATETIME
	)`)
	
	db.Exec(`CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT,
		content TEXT,
		create_at DATETIME
	)`)
	
	// 插入测试数据
	db.Exec("INSERT INTO users (name, email, create_at) VALUES (?, ?, ?)", 
		"John Doe", "john@example.com", time.Now())
	db.Exec("INSERT INTO users (name, email, create_at) VALUES (?, ?, ?)", 
		"Jane Smith", "jane@example.com", time.Now())
	db.Exec("INSERT INTO users (name, email, create_at) VALUES (?, ?, ?)", 
		"Bob Wilson", "bob@example.com", time.Now())
	
	return db
}

// TestSimpleSessionBasicOperations 测试简单会话基本操作
func TestSimpleSessionBasicOperations(t *testing.T) {
	db := setupTestDB()
	session := NewSimpleSession(db)
	ctx := context.Background()
	
	// 测试 SelectOne
	result, err := session.SelectOne(ctx, "SELECT COUNT(*) as count FROM users")
	if err != nil {
		t.Fatalf("SelectOne failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	// 测试 SelectList
	results, err := session.SelectList(ctx, "SELECT * FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("SelectList failed: %v", err)
	}
	
	if len(results) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(results))
	}
	
	// 测试 Insert
	affectedRows, err := session.Insert(ctx, "INSERT INTO users (name, email, create_at) VALUES (?, ?, ?)",
		"Test User", "test@example.com", time.Now())
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	
	if affectedRows != 1 {
		t.Fatalf("Expected 1 affected row, got %d", affectedRows)
	}
	
	// 测试 Update
	affectedRows, err = session.Update(ctx, "UPDATE users SET name = ? WHERE id = ?", "Updated User", 1)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	
	if affectedRows != 1 {
		t.Fatalf("Expected 1 affected row, got %d", affectedRows)
	}
	
	log.Println("TestSimpleSessionBasicOperations passed")
}

// TestDryRunMode 测试DryRun模式
func TestDryRunMode(t *testing.T) {
	session := NewSimpleSession(nil).DryRun(true)
	ctx := context.Background()
	
	// DryRun模式下，这些操作不会真正执行
	result, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("DryRun SelectOne failed: %v", err)
	}
	
	// DryRun模式应该返回空结果
	if result != nil {
		t.Fatal("DryRun should return nil result")
	}
	
	// 测试DryRun插入
	affectedRows, err := session.Insert(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Test", "test@example.com")
	if err != nil {
		t.Fatalf("DryRun Insert failed: %v", err)
	}
	
	// DryRun模式应该返回0
	if affectedRows != 0 {
		t.Fatalf("DryRun should return 0 affected rows, got %d", affectedRows)
	}
	
	log.Println("TestDryRunMode passed")
}

// TestPagination 测试分页功能
func TestPagination(t *testing.T) {
	db := setupTestDB()
	session := NewSimpleSession(db)
	ctx := context.Background()
	
	// 插入更多测试数据
	for i := 0; i < 10; i++ {
		session.Insert(ctx, "INSERT INTO users (name, email, create_at) VALUES (?, ?, ?)",
			fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i), time.Now())
	}
	
	// 测试分页查询
	pageReq := PageRequest{Page: 1, Size: 5}
	pageResult, err := session.SelectPage(ctx, "SELECT * FROM users ORDER BY id", pageReq)
	if err != nil {
		t.Fatalf("SelectPage failed: %v", err)
	}
	
	if len(pageResult.Items) != 5 {
		t.Fatalf("Expected 5 items in page, got %d", len(pageResult.Items))
	}
	
	if pageResult.Total < 10 {
		t.Fatalf("Expected total >= 10, got %d", pageResult.Total)
	}
	
	if pageResult.TotalPages < 2 {
		t.Fatalf("Expected totalPages >= 2, got %d", pageResult.TotalPages)
	}
	
	// 测试第二页
	pageReq = PageRequest{Page: 2, Size: 5}
	pageResult, err = session.SelectPage(ctx, "SELECT * FROM users ORDER BY id", pageReq)
	if err != nil {
		t.Fatalf("SelectPage page 2 failed: %v", err)
	}
	
	if len(pageResult.Items) == 0 {
		t.Fatal("Expected items in page 2")
	}
	
	log.Println("TestPagination passed")
}

// TestHooks 测试钩子系统
func TestHooks(t *testing.T) {
	db := setupTestDB()
	
	var beforeCalled, afterCalled bool
	
	beforeHook := func(ctx context.Context, sql string, args []interface{}) error {
		beforeCalled = true
		log.Printf("Before hook: SQL = %s", sql)
		return nil
	}
	
	afterHook := func(ctx context.Context, result interface{}, duration time.Duration, err error) {
		afterCalled = true
		log.Printf("After hook: Duration = %v, Error = %v", duration, err)
	}
	
	session := NewSimpleSession(db).
		AddBeforeHook(beforeHook).
		AddAfterHook(afterHook)
	
	ctx := context.Background()
	
	// 执行查询，应该触发钩子
	_, err := session.SelectOne(ctx, "SELECT COUNT(*) FROM users")
	if err != nil {
		t.Fatalf("Query with hooks failed: %v", err)
	}
	
	if !beforeCalled {
		t.Fatal("Before hook was not called")
	}
	
	if !afterCalled {
		t.Fatal("After hook was not called")
	}
	
	log.Println("TestHooks passed")
}

// TestTransactionAware 测试事务感知会话
func TestTransactionAware(t *testing.T) {
	db := setupTestDB()
	txSession := NewTransactionAwareSession(db)
	
	ctx := context.WithValue(context.Background(), UserIDKey, "test_user")
	
	// 测试事务执行
	err := txSession.ExecuteInTransaction(ctx, "test_user", func(txCtx context.Context, session SimpleSession) error {
		// 在事务中插入用户
		_, err := session.Insert(txCtx, "INSERT INTO users (name, email, create_at) VALUES (?, ?, ?)",
			"TX User", "tx@example.com", time.Now())
		if err != nil {
			return err
		}
		
		// 在事务中插入文章
		_, err = session.Insert(txCtx, "INSERT INTO posts (user_id, title, content, create_at) VALUES (?, ?, ?, ?)",
			1, "TX Post", "Transaction test", time.Now())
		return err
	})
	
	if err != nil {
		t.Fatalf("Transaction execution failed: %v", err)
	}
	
	// 验证数据已提交
	result, err := txSession.SelectOne(context.Background(), "SELECT COUNT(*) as count FROM posts")
	if err != nil {
		t.Fatalf("Failed to verify transaction result: %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected non-nil result after transaction")
	}
	
	log.Println("TestTransactionAware passed")
}

// TestPerformanceHook 测试性能监控钩子
func TestPerformanceHook(t *testing.T) {
	db := setupTestDB()
	
	var slowQueryDetected bool
	
	beforeHook, afterHook := PerformanceHook(1 * time.Millisecond) // 设置很小的阈值来触发慢查询检测
	
	// 重写afterHook来捕获慢查询
	afterHookWrapper := func(ctx context.Context, result interface{}, duration time.Duration, err error) {
		afterHook(ctx, result, duration, err) // 调用原始钩子
		if duration > 1*time.Millisecond {
			slowQueryDetected = true
		}
	}
	
	session := NewSimpleSession(db).
		AddBeforeHook(beforeHook).
		AddAfterHook(afterHookWrapper)
	
	ctx := context.WithValue(context.Background(), UserIDKey, "perf_test")
	
	// 执行一个查询（可能会被检测为慢查询）
	_, err := session.SelectList(ctx, "SELECT * FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("Performance test query failed: %v", err)
	}
	
	log.Printf("Slow query detected: %v", slowQueryDetected)
	log.Println("TestPerformanceHook passed")
}

// TestMain 测试入口
func TestMain(m *testing.M) {
	log.Println("Starting MyBatis simplified version tests...")
	
	// 运行所有测试
	code := m.Run()
	
	log.Println("Tests completed.")
	os.Exit(code)
}

// 需要修复导入问题
func init() {
	// 这里可以添加一些初始化逻辑
}