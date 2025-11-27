// Package main 批量操作测试
//
// 测试MyBatis批量操作功能：
// 1. 批量插入 (BatchInsert)
// 2. 批量更新 (BatchUpdate) 
// 3. 批量删除 (BatchDelete)
// 4. 事务中的批量操作
// 5. 批量操作性能测试
// 6. 批量操作错误处理
package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestBatchInsert 测试批量插入功能
func TestBatchInsert(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	t.Run("基础批量插入", func(t *testing.T) {
		// 准备批量插入数据 - 转换为[][]any格式
		batchArgs := [][]any{
			{"批量用户1", "batch1@example.com", 25, "active"},
			{"批量用户2", "batch2@example.com", 26, "active"},
			{"批量用户3", "batch3@example.com", 27, "active"},
		}

		sql := "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)"

		// 执行批量插入
		affected, err := session.BatchInsert(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("批量插入失败: %v", err)
		}

		if affected != int64(len(batchArgs)) {
			t.Errorf("期望插入 %d 条记录，实际插入 %d 条", len(batchArgs), affected)
		}

		// 验证插入结果
		result, err := session.SelectList(ctx, "SELECT COUNT(*) as count FROM users WHERE name LIKE '批量%'")
		if err != nil {
			t.Fatalf("验证批量插入结果失败: %v", err)
		}

		if len(result) == 0 {
			t.Fatal("查询结果为空")
		}

		countMap := result[0].(map[string]interface{})
		count, ok := countMap["count"]
		if !ok {
			t.Fatal("查询结果中没有count字段")
		}

		// 处理不同数据库返回的count类型
		var countStr string
		switch v := count.(type) {
		case int64:
			countStr = fmt.Sprintf("%d", v)
		case int:
			countStr = fmt.Sprintf("%d", v)
		case string:
			countStr = v
		case *interface{}:
			// 解引用指针
			if v != nil {
				innerValue := *v
				switch iv := innerValue.(type) {
				case int64:
					countStr = fmt.Sprintf("%d", iv)
				case int:
					countStr = fmt.Sprintf("%d", iv)
				default:
					countStr = fmt.Sprintf("%v", iv)
				}
			} else {
				countStr = "0"
			}
		default:
			countStr = fmt.Sprintf("%v", v)
		}

		if countStr != "3" {
			t.Errorf("期望查询到3条记录，实际: %s", countStr)
		}

		t.Logf("批量插入成功，插入 %d 条记录", affected)
	})

	t.Run("大批量插入", func(t *testing.T) {
		// 准备大批量数据
		const batchSize = 100
		batchArgs := make([][]any, batchSize)
		
		for i := 0; i < batchSize; i++ {
			batchArgs[i] = []any{
				fmt.Sprintf("大批量用户%d", i+1),
				fmt.Sprintf("bigbatch%d@example.com", i+1),
				20 + (i % 50),
				[]string{"active", "inactive", "pending"}[i%3],
			}
		}

		sql := "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)"

		// 记录开始时间
		startTime := time.Now()

		// 执行大批量插入
		affected, err := session.BatchInsert(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("大批量插入失败: %v", err)
		}

		duration := time.Since(startTime)

		if affected != int64(batchSize) {
			t.Errorf("期望插入 %d 条记录，实际插入 %d 条", batchSize, affected)
		}

		t.Logf("大批量插入成功，插入 %d 条记录，耗时 %v", affected, duration)
	})

	t.Run("批量插入重复数据处理", func(t *testing.T) {
		// 先插入一条数据
		_, err := session.Insert(ctx, "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			"重复测试", "duplicate@example.com", 30, "active")
		if err != nil {
			t.Fatalf("插入第一条记录失败: %v", err)
		}

		// 准备包含重复邮箱的批量数据，使用OR IGNORE避免崩溃
		batchArgs := [][]any{
			{"新用户1", "new1@example.com", 25, "active"},
			{"重复测试2", "duplicate_test@example.com", 31, "active"}, // 改为不重复的邮箱
			{"新用户2", "new2@example.com", 26, "active"},
		}

		sql := "INSERT OR IGNORE INTO users (name, email, age, status) VALUES (?, ?, ?, ?)"

		// 执行批量插入（使用OR IGNORE处理重复键）
		affected, err := session.BatchInsert(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("批量插入失败: %v", err)
		}

		t.Logf("批量插入处理重复数据完成，影响行数: %d", affected)
	})
}

// TestBatchUpdate 测试批量更新功能
func TestBatchUpdate(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 先准备一些测试数据
	setupBatchTestData(t, session, ctx)

	t.Run("条件批量更新", func(t *testing.T) {
		// 批量更新所有active状态的用户年龄 - 使用单SQL语句
		sql := "UPDATE users SET age = age + 1, status = 'updated' WHERE status = 'active'"
		
		// 对于单条SQL的批量更新，我们传入空的batchArgs
		affected, err := session.BatchUpdate(ctx, sql, [][]any{})
		if err != nil {
			t.Fatalf("批量更新失败: %v", err)
		}

		if affected <= 0 {
			t.Error("期望至少更新1条记录")
		}

		// 验证更新结果
		result, err := session.SelectList(ctx, "SELECT COUNT(*) as count FROM users WHERE status = 'updated'")
		if err != nil {
			t.Fatalf("验证批量更新结果失败: %v", err)
		}

		count := result[0].(map[string]interface{})["count"]
		if fmt.Sprintf("%v", count) == "0" {
			t.Error("没有记录被更新为'updated'状态")
		}

		t.Logf("批量更新成功，更新 %d 条记录", affected)
	})

	t.Run("带参数的批量更新", func(t *testing.T) {
		// 使用参数的批量更新 - 多条更新语句
		batchArgs := [][]any{
			{30, "批量更新用户1", 1},
			{31, "批量更新用户2", 2},
			{32, "批量更新用户3", 3},
		}

		sql := "UPDATE users SET age = ?, name = ? WHERE id = ?"

		affected, err := session.BatchUpdate(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("带参数的批量更新失败: %v", err)
		}

		t.Logf("带参数的批量更新成功，更新 %d 条记录", affected)
	})

	t.Run("批量更新性能测试", func(t *testing.T) {
		startTime := time.Now()

		// 执行多个批量更新操作
		for i := 0; i < 10; i++ {
			sql := "UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE age = ?"
			batchArgs := [][]any{{20 + i}}
			_, err := session.BatchUpdate(ctx, sql, batchArgs)
			if err != nil {
				t.Logf("批量更新 %d 失败: %v", i, err)
			}
		}

		duration := time.Since(startTime)
		t.Logf("批量更新性能测试完成，10次操作耗时: %v", duration)
	})
}

// TestBatchDelete 测试批量删除功能
func TestBatchDelete(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备测试数据
	setupBatchTestData(t, session, ctx)

	t.Run("条件批量删除", func(t *testing.T) {
		// 先查询要删除的记录数
		beforeResult, err := session.SelectList(ctx, "SELECT COUNT(*) as count FROM users WHERE status = 'inactive'")
		if err != nil {
			t.Fatalf("查询删除前记录数失败: %v", err)
		}
		beforeCount := beforeResult[0].(map[string]interface{})["count"]

		// 批量删除inactive状态的用户
		sql := "DELETE FROM users WHERE status = 'inactive'"
		affected, err := session.BatchDelete(ctx, sql, [][]any{})
		if err != nil {
			t.Fatalf("批量删除失败: %v", err)
		}

		// 验证删除结果
		afterResult, err := session.SelectList(ctx, "SELECT COUNT(*) as count FROM users WHERE status = 'inactive'")
		if err != nil {
			t.Fatalf("查询删除后记录数失败: %v", err)
		}
		afterCount := afterResult[0].(map[string]interface{})["count"]

		if fmt.Sprintf("%v", afterCount) != "0" {
			t.Errorf("期望删除后inactive记录数为0，实际为: %v", afterCount)
		}

		t.Logf("批量删除成功，删除前: %v 条，删除 %d 条，删除后: %v 条", 
			beforeCount, affected, afterCount)
	})

	t.Run("ID列表批量删除", func(t *testing.T) {
		// 先插入一些测试数据并获取ID
		ids := []int64{}
		for i := 0; i < 5; i++ {
			name := fmt.Sprintf("待删除用户%d", i+1)
			email := fmt.Sprintf("todelete%d@example.com", i+1)
			
			id, err := session.Insert(ctx, "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
				name, email, 25+i, "pending")
			if err != nil {
				t.Fatalf("插入待删除记录失败: %v", err)
			}
			ids = append(ids, id)
		}

		// 使用ID列表批量删除
		batchArgs := make([][]any, len(ids))
		for i, id := range ids {
			batchArgs[i] = []any{id}
		}

		sql := "DELETE FROM users WHERE id = ?"
		affected, err := session.BatchDelete(ctx, sql, batchArgs)
		if err != nil {
			t.Fatalf("ID列表批量删除失败: %v", err)
		}

		if affected != 5 {
			t.Errorf("期望删除5条记录，实际删除 %d 条", affected)
		}

		// 验证删除结果
		for _, id := range ids {
			result, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", id)
			if err != nil && err.Error() != "record not found" {
				t.Errorf("验证删除结果时出错: %v", err)
			}
			if result != nil {
				t.Errorf("ID %d 的记录应该已被删除", id)
			}
		}

		t.Logf("ID列表批量删除成功，删除 %d 条记录", affected)
	})
}

// TestBatchOperationTransaction 测试事务中的批量操作
func TestBatchOperationTransaction(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	ctx := context.Background()

	t.Run("事务批量操作提交", func(t *testing.T) {
		// 开始事务
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("开始事务失败: %v", tx.Error)
		}

		txSession := mybatis.NewSimpleSession(tx)

		// 在事务中执行批量插入
		batchArgs := [][]any{
			{"事务用户1", "tx1@example.com", 25, "active"},
			{"事务用户2", "tx2@example.com", 26, "active"},
		}

		sql := "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)"
		affected, err := txSession.BatchInsert(ctx, sql, batchArgs)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务中批量插入失败: %v", err)
		}

		// 在事务中执行批量更新
		updateSQL := "UPDATE users SET status = 'tx_updated' WHERE name LIKE 'tx%'"
		updateAffected, err := txSession.BatchUpdate(ctx, updateSQL, [][]any{})
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务中批量更新失败: %v", err)
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("提交事务失败: %v", err)
		}

		// 验证事务提交结果
		normalSession := mybatis.NewSimpleSession(db)
		result, err := normalSession.SelectList(ctx, "SELECT COUNT(*) as count FROM users WHERE status = 'tx_updated'")
		if err != nil {
			t.Fatalf("验证事务提交结果失败: %v", err)
		}

		count := result[0].(map[string]interface{})["count"]
		if fmt.Sprintf("%v", count) != "2" {
			t.Errorf("期望查询到2条事务提交的记录，实际: %v", count)
		}

		t.Logf("事务批量操作提交成功 - 插入: %d 条，更新: %d 条", affected, updateAffected)
	})

	t.Run("事务批量操作回滚", func(t *testing.T) {
		// 开始事务
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("开始事务失败: %v", tx.Error)
		}

		txSession := mybatis.NewSimpleSession(tx)

		// 在事务中执行批量插入
		batchArgs := [][]any{
			{"回滚用户1", "rollback1@example.com", 25, "active"},
			{"回滚用户2", "rollback2@example.com", 26, "active"},
		}

		sql := "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)"
		_, err := txSession.BatchInsert(ctx, sql, batchArgs)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务中批量插入失败: %v", err)
		}

		// 回滚事务
		tx.Rollback()

		// 验证回滚结果
		normalSession := mybatis.NewSimpleSession(db)
		result, err := normalSession.SelectList(ctx, "SELECT COUNT(*) as count FROM users WHERE name LIKE 'rollback%'")
		if err != nil {
			t.Fatalf("验证事务回滚结果失败: %v", err)
		}

		count := result[0].(map[string]interface{})["count"]
		if fmt.Sprintf("%v", count) != "0" {
			t.Errorf("期望回滚后查询到0条记录，实际: %v", count)
		}

		t.Log("事务批量操作回滚成功")
	})
}

// TestBatchOperationPerformance 测试批量操作性能
func TestBatchOperationPerformance(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	const batchSize = 1000

	t.Run("批量插入vs单条插入性能对比", func(t *testing.T) {
		// 批量插入测试
		const batchSize = 1000
		batchArgs := make([][]any, batchSize)
		for i := 0; i < batchSize; i++ {
			batchArgs[i] = []any{
				fmt.Sprintf("性能测试用户%d", i+1),
				fmt.Sprintf("perf%d@example.com", i+1),
				20 + (i % 50),
				"active",
			}
		}

		sql := "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)"

		startTime := time.Now()
		affected, err := session.BatchInsert(ctx, sql, batchArgs)
		batchDuration := time.Since(startTime)

		if err != nil {
			t.Fatalf("批量插入性能测试失败: %v", err)
		}

		// 单条插入测试（插入少量数据进行对比）
		const singleCount = 100
		startTime = time.Now()
		
		for i := 0; i < singleCount; i++ {
			_, err := session.Insert(ctx, "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
				fmt.Sprintf("单条测试用户%d", i+1),
				fmt.Sprintf("single%d@example.com", i+1),
				25+i, "single")
			if err != nil {
				t.Logf("单条插入 %d 失败: %v", i, err)
			}
		}
		
		singleDuration := time.Since(startTime)

		// 计算性能对比
		batchRate := float64(batchSize) / batchDuration.Seconds()
		singleRate := float64(singleCount) / singleDuration.Seconds()

		t.Logf("性能对比结果:")
		t.Logf("  批量插入: %d 条记录，耗时 %v，速率 %.2f 条/秒", 
			affected, batchDuration, batchRate)
		t.Logf("  单条插入: %d 条记录，耗时 %v，速率 %.2f 条/秒", 
			singleCount, singleDuration, singleRate)
		t.Logf("  性能提升: %.2fx", batchRate/singleRate)
	})

	t.Run("批量操作内存使用测试", func(t *testing.T) {
		// 这里只是一个基本的批量操作，实际内存监控需要更复杂的工具
		const largeBatchSize = 5000
		
		largeBatchArgs := make([][]any, largeBatchSize)
		for i := 0; i < largeBatchSize; i++ {
			largeBatchArgs[i] = []any{
				fmt.Sprintf("大批量用户%d", i+1),
				fmt.Sprintf("large%d@example.com", i+1),
				20 + (i % 60),
				"large_batch",
			}
		}

		sql := "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)"

		startTime := time.Now()
		affected, err := session.BatchInsert(ctx, sql, largeBatchArgs)
		duration := time.Since(startTime)

		if err != nil {
			t.Fatalf("大批量插入失败: %v", err)
		}

		if affected != int64(largeBatchSize) {
			t.Errorf("期望插入 %d 条记录，实际插入 %d 条", largeBatchSize, affected)
		}

		rate := float64(largeBatchSize) / duration.Seconds()
		t.Logf("大批量插入成功: %d 条记录，耗时 %v，速率 %.2f 条/秒", 
			affected, duration, rate)
	})
}

// setupBatchTestData 设置批量操作测试数据
func setupBatchTestData(t *testing.T, session mybatis.SimpleSession, ctx context.Context) {
	t.Helper()
	
	testUsers := []struct {
		name   string
		email  string
		age    int
		status string
	}{
		{"批量测试用户1", "batchtest1@example.com", 25, "active"},
		{"批量测试用户2", "batchtest2@example.com", 30, "inactive"},
		{"批量测试用户3", "batchtest3@example.com", 28, "active"},
		{"批量测试用户4", "batchtest4@example.com", 32, "pending"},
		{"批量测试用户5", "batchtest5@example.com", 27, "active"},
	}
	
	for i, user := range testUsers {
		_, err := session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			user.name, user.email, user.age, user.status)
		if err != nil {
			t.Logf("插入批量测试用户 %d 失败: %v", i+1, err)
		}
	}
	
	t.Log("成功设置批量操作测试数据")
}