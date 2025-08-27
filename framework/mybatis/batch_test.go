package mybatis_test

import (
	"context"
	"testing"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBatchOperations(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建测试表
	err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			age INTEGER
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	session := mybatis.NewSimpleSession(db)
	ctx := context.Background()

	t.Run("BatchInsert", func(t *testing.T) {
		sql := "INSERT INTO users (name, email, age) VALUES (?, ?, ?)"
		batchArgs := [][]any{
			{"张三", "zhangsan@example.com", 25},
			{"李四", "lisi@example.com", 30},
			{"王五", "wangwu@example.com", 28},
		}

		affectedRows, err := session.BatchInsert(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("BatchInsert failed: %v", err)
		}

		if affectedRows != 3 {
			t.Errorf("Expected 3 affected rows, got %d", affectedRows)
		}

		// 验证插入的数据
		results, err := session.SelectList(ctx, "SELECT COUNT(*) as count FROM users")
		if err != nil {
			t.Fatalf("Failed to count users: %v", err)
		}

		countValue := results[0].(map[string]any)["count"]
		var count int64

		// 处理各种可能的类型包装
		for countValue != nil {
			switch v := countValue.(type) {
			case int64:
				count = v
				goto done
			case int:
				count = int64(v)
				goto done
			case int32:
				count = int64(v)
				goto done
			case *interface{}:
				// 解引用指针
				if v != nil {
					countValue = *v
				} else {
					countValue = nil
				}
			default:
				t.Fatalf("Unexpected count type: %T, value: %v", v, v)
			}
		}
		t.Fatalf("count value is nil")
	done:

		if count != 3 {
			t.Errorf("Expected 3 users, found %d", count)
		}
	})

	t.Run("BatchUpdate", func(t *testing.T) {
		sql := "UPDATE users SET age = ? WHERE name = ?"
		batchArgs := [][]any{
			{26, "张三"},
			{31, "李四"},
			{29, "王五"},
		}

		affectedRows, err := session.BatchUpdate(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("BatchUpdate failed: %v", err)
		}

		if affectedRows != 3 {
			t.Errorf("Expected 3 affected rows, got %d", affectedRows)
		}

		// 验证更新的数据
		result, err := session.SelectOne(ctx, "SELECT age FROM users WHERE name = ?", "张三")
		if err != nil {
			t.Fatalf("Failed to select updated user: %v", err)
		}

		age := result.(map[string]any)["age"]
		if age != int64(26) {
			t.Errorf("Expected age 26, got %v", age)
		}
	})

	t.Run("BatchDelete", func(t *testing.T) {
		sql := "DELETE FROM users WHERE name = ?"
		batchArgs := [][]any{
			{"张三"},
			{"李四"},
		}

		affectedRows, err := session.BatchDelete(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("BatchDelete failed: %v", err)
		}

		if affectedRows != 2 {
			t.Errorf("Expected 2 affected rows, got %d", affectedRows)
		}

		// 验证删除的数据
		results, err := session.SelectList(ctx, "SELECT COUNT(*) as count FROM users")
		if err != nil {
			t.Fatalf("Failed to count remaining users: %v", err)
		}

		countValue := results[0].(map[string]any)["count"]
		var count int64

		// 处理各种可能的类型包装
		for countValue != nil {
			switch v := countValue.(type) {
			case int64:
				count = v
				goto done
			case int:
				count = int64(v)
				goto done
			case int32:
				count = int64(v)
				goto done
			case *interface{}:
				// 解引用指针
				if v != nil {
					countValue = *v
				} else {
					countValue = nil
				}
			default:
				t.Fatalf("Unexpected count type: %T, value: %v", v, v)
			}
		}
		t.Fatalf("count value is nil")
	done:

		if count != 1 {
			t.Errorf("Expected 1 remaining user, found %d", count)
		}
	})

	t.Run("EmptyBatch", func(t *testing.T) {
		sql := "INSERT INTO users (name, email, age) VALUES (?, ?, ?)"
		batchArgs := [][]any{}

		affectedRows, err := session.BatchInsert(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("Empty BatchInsert failed: %v", err)
		}

		if affectedRows != 0 {
			t.Errorf("Expected 0 affected rows for empty batch, got %d", affectedRows)
		}
	})
}

func TestXMLSessionBatchOperations(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建测试表
	err = db.Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price DECIMAL(10,2),
			category_id INTEGER
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	session := mybatis.NewXMLSession(db)
	ctx := context.Background()

	t.Run("XMLSession_BatchInsert", func(t *testing.T) {
		sql := "INSERT INTO products (name, price, category_id) VALUES (?, ?, ?)"
		batchArgs := [][]any{
			{"iPhone 15", 999.99, 1},
			{"Samsung Galaxy", 899.99, 1},
			{"MacBook Pro", 1999.99, 2},
		}

		affectedRows, err := session.BatchInsert(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("XMLSession BatchInsert failed: %v", err)
		}

		if affectedRows != 3 {
			t.Errorf("Expected 3 affected rows, got %d", affectedRows)
		}

		// 验证插入的数据
		results, err := session.SelectList(ctx, "SELECT COUNT(*) as count FROM products")
		if err != nil {
			t.Fatalf("Failed to count products: %v", err)
		}

		countValue := results[0].(map[string]any)["count"]
		var count int64

		// 处理各种可能的类型包装
		for countValue != nil {
			switch v := countValue.(type) {
			case int64:
				count = v
				goto done
			case int:
				count = int64(v)
				goto done
			case int32:
				count = int64(v)
				goto done
			case *interface{}:
				// 解引用指针
				if v != nil {
					countValue = *v
				} else {
					countValue = nil
				}
			default:
				t.Fatalf("Unexpected count type: %T, value: %v", v, v)
			}
		}
		t.Fatalf("count value is nil")
	done:

		if count != 3 {
			t.Errorf("Expected 3 products, found %d", count)
		}
	})
}