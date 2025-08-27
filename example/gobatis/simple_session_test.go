// Package main SimpleSession基础CRUD测试
//
// 测试SimpleSession的基本功能：
// 1. CRUD操作（Create, Read, Update, Delete）
// 2. 分页查询
// 3. DryRun模式
// 4. 错误处理
// 5. 事务支持
package main

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestSimpleSessionCRUD 测试基础CRUD操作
func TestSimpleSessionCRUD(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	t.Run("Create-插入操作", func(t *testing.T) {
		// 测试插入单条记录
		insertSQL := `INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`
		userID, err := session.Insert(ctx, insertSQL, "测试用户", "test@example.com", 25, "active")
		
		if err != nil {
			t.Fatalf("插入用户失败: %v", err)
		}
		
		if userID <= 0 {
			t.Errorf("期望用户ID > 0，得到: %d", userID)
		}
		
		t.Logf("成功插入用户，ID: %d", userID)
	})

	t.Run("Read-查询操作", func(t *testing.T) {
		// 先插入一个测试用户
		userID, err := session.Insert(ctx, 
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			"查询测试用户", "query@example.com", 30, "active")
		if err != nil {
			t.Fatalf("准备查询测试数据失败: %v", err)
		}

		// 测试查询单条记录
		selectSQL := `SELECT * FROM users WHERE id = ?`
		result, err := session.SelectOne(ctx, selectSQL, userID)
		if err != nil {
			t.Fatalf("查询单条记录失败: %v", err)
		}
		
		if result == nil {
			t.Fatal("期望查询到结果，但得到nil")
		}
		
		// 验证结果内容
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("期望结果类型为map[string]interface{}，得到: %T", result)
		}
		
		if resultMap["name"] != "查询测试用户" {
			t.Errorf("期望name='查询测试用户'，得到: %v", resultMap["name"])
		}
		
		t.Logf("成功查询到用户: %+v", result)

		// 测试查询多条记录
		listSQL := `SELECT * FROM users WHERE status = ? LIMIT 5`
		results, err := session.SelectList(ctx, listSQL, "active")
		if err != nil {
			t.Fatalf("查询多条记录失败: %v", err)
		}
		
		if len(results) == 0 {
			t.Error("期望查询到结果，但结果为空")
		}
		
		t.Logf("成功查询到 %d 条记录", len(results))
	})

	t.Run("Update-更新操作", func(t *testing.T) {
		// 先插入一个测试用户
		userID, err := session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			"更新测试用户", "update@example.com", 28, "inactive")
		if err != nil {
			t.Fatalf("准备更新测试数据失败: %v", err)
		}

		// 测试更新操作
		updateSQL := `UPDATE users SET age = ?, status = ? WHERE id = ?`
		affected, err := session.Update(ctx, updateSQL, 29, "active", userID)
		if err != nil {
			t.Fatalf("更新记录失败: %v", err)
		}
		
		if affected != 1 {
			t.Errorf("期望影响1行，实际影响 %d 行", affected)
		}

		// 验证更新结果
		result, err := session.SelectOne(ctx, `SELECT age, status FROM users WHERE id = ?`, userID)
		if err != nil {
			t.Fatalf("验证更新结果失败: %v", err)
		}
		
		resultMap := result.(map[string]interface{})
		if fmt.Sprintf("%v", resultMap["age"]) != "29" {
			t.Errorf("期望年龄为29，得到: %v", resultMap["age"])
		}
		if resultMap["status"] != "active" {
			t.Errorf("期望状态为active，得到: %v", resultMap["status"])
		}
		
		t.Logf("成功更新记录，影响 %d 行", affected)
	})

	t.Run("Delete-删除操作", func(t *testing.T) {
		// 先插入一个测试用户
		userID, err := session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			"删除测试用户", "delete@example.com", 35, "active")
		if err != nil {
			t.Fatalf("准备删除测试数据失败: %v", err)
		}

		// 测试删除操作
		deleteSQL := `DELETE FROM users WHERE id = ?`
		affected, err := session.Delete(ctx, deleteSQL, userID)
		if err != nil {
			t.Fatalf("删除记录失败: %v", err)
		}
		
		if affected != 1 {
			t.Errorf("期望删除1行，实际删除 %d 行", affected)
		}

		// 验证删除结果
		result, err := session.SelectOne(ctx, `SELECT * FROM users WHERE id = ?`, userID)
		if err != nil {
			// GORM的RecordNotFound错误是正常的
			if err.Error() != "record not found" {
				t.Fatalf("验证删除结果时出现意外错误: %v", err)
			}
		} else if result != nil {
			t.Error("期望记录已被删除，但仍能查询到")
		}
		
		t.Logf("成功删除记录，影响 %d 行", affected)
	})
}

// TestSimpleSessionPagination 测试分页查询
func TestSimpleSessionPagination(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备测试数据
	setupPaginationTestData(t, session, ctx, 50)

	t.Run("基本分页查询", func(t *testing.T) {
		pageSQL := `SELECT * FROM users WHERE status = 'active' ORDER BY id`
		pageRequest := mybatis.PageRequest{
			Page: 1,
			Size: 10,
		}

		pageResult, err := session.SelectPage(ctx, pageSQL, pageRequest)
		if err != nil {
			t.Fatalf("分页查询失败: %v", err)
		}

		// 验证分页结果
		if pageResult.Page != 1 {
			t.Errorf("期望当前页为1，得到: %d", pageResult.Page)
		}
		if pageResult.Size != 10 {
			t.Errorf("期望页面大小为10，得到: %d", pageResult.Size)
		}
		if len(pageResult.Items) == 0 {
			t.Error("期望查询到数据，但结果为空")
		}
		if pageResult.Total <= 0 {
			t.Errorf("期望总记录数 > 0，得到: %d", pageResult.Total)
		}
		
		expectedTotalPages := (int(pageResult.Total) + pageResult.Size - 1) / pageResult.Size
		if pageResult.TotalPages != expectedTotalPages {
			t.Errorf("期望总页数为 %d，得到: %d", expectedTotalPages, pageResult.TotalPages)
		}

		t.Logf("分页查询成功 - 总记录数: %d, 当前页: %d/%d, 当前页数据: %d条", 
			pageResult.Total, pageResult.Page, pageResult.TotalPages, len(pageResult.Items))
	})

	t.Run("不同页码测试", func(t *testing.T) {
		pageSQL := `SELECT * FROM users ORDER BY id`
		
		// 测试第1页
		page1, err := session.SelectPage(ctx, pageSQL, mybatis.PageRequest{Page: 1, Size: 5})
		if err != nil {
			t.Fatalf("查询第1页失败: %v", err)
		}
		
		// 测试第2页
		page2, err := session.SelectPage(ctx, pageSQL, mybatis.PageRequest{Page: 2, Size: 5})
		if err != nil {
			t.Fatalf("查询第2页失败: %v", err)
		}

		// 验证两页数据不同（如果总数据 > 5）
		if page1.Total > 5 {
			if len(page1.Items) == 0 || len(page2.Items) == 0 {
				t.Error("分页数据不应为空")
			}
			
			// 简单验证：第一页第一条记录的ID应该不同于第二页第一条记录的ID
			page1Item := page1.Items[0].(map[string]interface{})
			page2Item := page2.Items[0].(map[string]interface{})
			if page1Item["id"] == page2Item["id"] {
				t.Error("不同页面的数据不应该相同")
			}
		}

		t.Logf("多页测试成功 - 第1页: %d条, 第2页: %d条", len(page1.Items), len(page2.Items))
	})
}

// TestSimpleSessionDryRun 测试DryRun模式
func TestSimpleSessionDryRun(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	// 创建DryRun会话
	dryRunSession := mybatis.NewSimpleSession(db).DryRun(true).Debug(true)
	ctx := context.Background()

	t.Run("DryRun查询测试", func(t *testing.T) {
		// DryRun模式下，查询应该不实际执行但不报错
		_, err := dryRunSession.SelectOne(ctx, `SELECT * FROM users WHERE id = ?`, 1)
		if err != nil {
			t.Errorf("DryRun查询不应该报错: %v", err)
		}
		
		t.Log("DryRun查询测试通过")
	})

	t.Run("DryRun插入测试", func(t *testing.T) {
		// DryRun模式下，插入应该不实际执行
		_, err := dryRunSession.Insert(ctx, 
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			"DryRun测试", "dryrun@example.com", 25, "active")
		if err != nil {
			t.Errorf("DryRun插入不应该报错: %v", err)
		}
		
		// 验证数据没有实际插入 - 使用正常session查询
		normalSession := mybatis.NewSimpleSession(db)
		result, err := normalSession.SelectOne(ctx, `SELECT * FROM users WHERE email = ?`, "dryrun@example.com")
		if err != nil && err.Error() != "record not found" {
			t.Errorf("验证DryRun时出现意外错误: %v", err)
		}
		if result != nil {
			t.Error("DryRun模式不应该实际插入数据")
		}
		
		t.Log("DryRun插入测试通过")
	})
}

// TestSimpleSessionTransaction 测试事务支持
func TestSimpleSessionTransaction(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	ctx := context.Background()

	t.Run("事务提交测试", func(t *testing.T) {
		// 开始事务
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("开始事务失败: %v", tx.Error)
		}

		// 创建事务会话
		txSession := mybatis.NewSimpleSession(tx)

		// 在事务中插入数据
		userID, err := txSession.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			"事务测试用户", "tx@example.com", 30, "active")
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务中插入数据失败: %v", err)
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("提交事务失败: %v", err)
		}

		// 验证数据已提交
		normalSession := mybatis.NewSimpleSession(db)
		result, err := normalSession.SelectOne(ctx, `SELECT * FROM users WHERE id = ?`, userID)
		if err != nil {
			t.Fatalf("验证事务提交失败: %v", err)
		}
		if result == nil {
			t.Error("事务提交后应该能查询到数据")
		}

		t.Log("事务提交测试通过")
	})

	t.Run("事务回滚测试", func(t *testing.T) {
		// 开始事务
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("开始事务失败: %v", tx.Error)
		}

		// 创建事务会话
		txSession := mybatis.NewSimpleSession(tx)

		// 在事务中插入数据
		_, err := txSession.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			"回滚测试用户", "rollback@example.com", 25, "active")
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务中插入数据失败: %v", err)
		}

		// 回滚事务
		tx.Rollback()

		// 验证数据已回滚
		normalSession := mybatis.NewSimpleSession(db)
		result, err := normalSession.SelectOne(ctx, `SELECT * FROM users WHERE email = ?`, "rollback@example.com")
		if err != nil && err.Error() != "record not found" {
			t.Errorf("验证事务回滚时出现意外错误: %v", err)
		}
		if result != nil {
			t.Error("事务回滚后不应该查询到数据")
		}

		t.Log("事务回滚测试通过")
	})
}

// TestSimpleSessionErrorHandling 测试错误处理
func TestSimpleSessionErrorHandling(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	t.Run("SQL语法错误", func(t *testing.T) {
		// 测试无效SQL
		_, err := session.SelectOne(ctx, `INVALID SQL STATEMENT`, 1)
		if err == nil {
			t.Error("期望SQL语法错误，但没有返回错误")
		}
		
		t.Logf("正确捕获SQL语法错误: %v", err)
	})

	t.Run("记录不存在", func(t *testing.T) {
		// 查询不存在的记录
		result, err := session.SelectOne(ctx, `SELECT * FROM users WHERE id = ?`, 99999)
		if err != nil && err.Error() != "record not found" {
			t.Errorf("期望'record not found'错误，得到: %v", err)
		}
		if result != nil {
			t.Error("期望查询结果为nil")
		}
		
		t.Log("正确处理记录不存在情况")
	})

	t.Run("唯一约束违反", func(t *testing.T) {
		// 先插入一条记录
		email := fmt.Sprintf("unique_%d@example.com", time.Now().UnixNano())
		_, err := session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			"唯一约束测试1", email, 25, "active")
		if err != nil {
			t.Fatalf("插入第一条记录失败: %v", err)
		}

		// 尝试插入相同email的记录
		_, err = session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			"唯一约束测试2", email, 30, "active")
		if err == nil {
			t.Error("期望唯一约束违反错误，但没有返回错误")
		}
		
		t.Logf("正确捕获唯一约束违反错误: %v", err)
	})
}

// setupPaginationTestData 准备分页测试数据
func setupPaginationTestData(t *testing.T, session mybatis.SimpleSession, ctx context.Context, count int) {
	t.Helper()
	
	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("分页用户%d", i)
		email := fmt.Sprintf("page_%d@example.com", i)
		age := rand.Intn(50) + 20
		status := []string{"active", "inactive", "pending"}[rand.Intn(3)]
		
		_, err := session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			name, email, age, status)
		if err != nil {
			t.Logf("插入分页测试数据失败 %d: %v", i, err)
		}
	}
	
	t.Logf("成功准备 %d 条分页测试数据", count)
}

// setupTestDatabase 设置测试数据库
func setupTestDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.New(
			nil, // 禁用日志输出
			logger.Config{
				LogLevel: logger.Error,
				Colorful: false,
			},
		),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// cleanupDatabase 清理测试数据库
func cleanupDatabase(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}
}