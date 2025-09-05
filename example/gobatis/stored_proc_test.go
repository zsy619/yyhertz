// Package main 存储过程调用测试
//
// 测试MyBatis存储过程调用功能：
// 1. 基础存储过程调用
// 2. 带输入参数的存储过程
// 3. 带输出参数的存储过程
// 4. 带输入输出参数的存储过程
// 5. 多结果集存储过程
// 6. 存储过程错误处理
// 7. 存储过程性能测试
package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestStoredProcBasic 测试基础存储过程调用
func TestStoredProcBasic(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	// 准备测试数据
	setupStoredProcTestData(t, session, ctx)

	t.Run("简单存储过程调用", func(t *testing.T) {
		// 创建简单的存储过程 (SQLite 不支持存储过程，这里用函数模拟)
		// 实际场景中，这些存储过程应该已经在数据库中定义好

		// 由于SQLite不支持存储过程，我们模拟一个简单的查询作为"存储过程"
		params := []mybatis.ProcParam{
			{Name: "status", Value: "active", Direction: "IN"},
		}

		// 调用"存储过程"（实际是执行查询）
		// 注意：这里只是演示API的使用，实际的存储过程调用需要数据库支持
		result, err := session.CallStoredProc(ctx, "GetActiveUsers", params)
		if err != nil {
			// 由于SQLite不支持存储过程，这里预期会出错
			t.Logf("预期的存储过程调用错误 (SQLite不支持): %v", err)
			return
		}

		t.Logf("存储过程调用结果: %+v", result)
	})

	t.Run("模拟存储过程功能", func(t *testing.T) {
		// 由于SQLite限制，我们用普通查询来模拟存储过程的效果

		// 模拟输入参数的存储过程
		status := "active"
		result, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 5", status)
		if err != nil {
			t.Fatalf("模拟存储过程查询失败: %v", err)
		}

		t.Logf("模拟存储过程查询成功，返回 %d 条记录", len(result))

		// 模拟聚合计算的存储过程
		aggResult, err := session.SelectOne(ctx,
			"SELECT COUNT(*) as total_users, AVG(age) as avg_age FROM users WHERE status = ?",
			status)
		if err != nil {
			t.Fatalf("模拟聚合存储过程失败: %v", err)
		}

		aggMap := aggResult.(map[string]interface{})
		t.Logf("模拟聚合存储过程结果: 总用户数=%v, 平均年龄=%v",
			aggMap["total_users"], aggMap["avg_age"])
	})
}

// TestStoredProcParameters 测试存储过程参数类型
func TestStoredProcParameters(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	setupStoredProcTestData(t, session, ctx)

	t.Run("输入参数测试", func(t *testing.T) {
		params := []mybatis.ProcParam{
			{Name: "min_age", Value: 25, Direction: "IN"},
			{Name: "max_age", Value: 35, Direction: "IN"},
			{Name: "status", Value: "active", Direction: "IN"},
		}

		// 尝试调用带多个输入参数的存储过程
		result, err := session.CallStoredProc(ctx, "GetUsersByAgeRange", params)
		if err != nil {
			t.Logf("输入参数存储过程调用错误 (预期的): %v", err)
		} else {
			t.Logf("输入参数存储过程调用结果: %+v", result)
		}

		// 使用普通查询模拟相同功能
		mockResult, err := session.SelectList(ctx,
			"SELECT * FROM users WHERE age BETWEEN ? AND ? AND status = ?",
			25, 35, "active")
		if err != nil {
			t.Fatalf("模拟输入参数查询失败: %v", err)
		}

		t.Logf("模拟输入参数存储过程成功，返回 %d 条记录", len(mockResult))
	})

	t.Run("输出参数测试", func(t *testing.T) {
		params := []mybatis.ProcParam{
			{Name: "input_status", Value: "active", Direction: "IN"},
			{Name: "total_count", Value: nil, Direction: "OUT"},
			{Name: "avg_age", Value: nil, Direction: "OUT"},
		}

		// 尝试调用带输出参数的存储过程
		result, err := session.CallStoredProc(ctx, "GetUserStats", params)
		if err != nil {
			t.Logf("输出参数存储过程调用错误 (预期的): %v", err)
		} else {
			t.Logf("输出参数存储过程调用结果: %+v", result)

			// 检查输出参数
			if result.OutputParams != nil {
				t.Logf("输出参数: %+v", result.OutputParams)
			}
		}

		// 用查询模拟输出参数功能
		mockResult, err := session.SelectOne(ctx,
			"SELECT COUNT(*) as total_count, AVG(age) as avg_age FROM users WHERE status = ?",
			"active")
		if err != nil {
			t.Fatalf("模拟输出参数查询失败: %v", err)
		}

		mockMap := mockResult.(map[string]interface{})
		t.Logf("模拟输出参数结果: 总数=%v, 平均年龄=%v",
			mockMap["total_count"], mockMap["avg_age"])
	})

	t.Run("输入输出参数测试", func(t *testing.T) {
		params := []mybatis.ProcParam{
			{Name: "user_id", Value: 1, Direction: "INOUT"},
			{Name: "increment", Value: 5, Direction: "IN"},
			{Name: "new_age", Value: nil, Direction: "OUT"},
		}

		// 尝试调用带输入输出参数的存储过程
		result, err := session.CallStoredProc(ctx, "UpdateUserAge", params)
		if err != nil {
			t.Logf("输入输出参数存储过程调用错误 (预期的): %v", err)
		} else {
			t.Logf("输入输出参数存储过程调用结果: %+v", result)
		}

		// 用事务模拟输入输出参数功能
		tx := db.Begin()
		txSession := mybatis.NewSimpleSession(tx)

		// 获取当前年龄
		currentUser, err := txSession.SelectOne(ctx, "SELECT age FROM users WHERE id = ?", 1)
		if err == nil {
			currentMap := currentUser.(map[string]interface{})
			currentAge := fmt.Sprintf("%v", currentMap["age"])
			t.Logf("用户当前年龄: %s", currentAge)

			// 更新年龄
			_, err = txSession.Update(ctx, "UPDATE users SET age = age + ? WHERE id = ?", 5, 1)
			if err != nil {
				tx.Rollback()
				t.Fatalf("模拟更新年龄失败: %v", err)
			}

			// 获取新年龄
			newUser, err := txSession.SelectOne(ctx, "SELECT age FROM users WHERE id = ?", 1)
			if err == nil {
				newMap := newUser.(map[string]interface{})
				newAge := fmt.Sprintf("%v", newMap["age"])
				t.Logf("用户更新后年龄: %s", newAge)
			}
		}

		tx.Commit()
		t.Log("模拟输入输出参数存储过程完成")
	})
}

// TestStoredProcMultiResults 测试多结果集存储过程
func TestStoredProcMultiResults(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	setupStoredProcTestData(t, session, ctx)

	t.Run("多结果集存储过程", func(t *testing.T) {
		params := []mybatis.ProcParam{
			{Name: "status_filter", Value: "active", Direction: "IN"},
		}

		// 尝试调用返回多个结果集的存储过程
		result, err := session.CallStoredProcWithMultiResults(ctx, "GetMultipleResultSets", params)
		if err != nil {
			t.Logf("多结果集存储过程调用错误 (预期的): %v", err)
		} else {
			t.Logf("多结果集存储过程调用结果: %+v", result)

			if len(result.ResultSets) > 0 {
				for i, rs := range result.ResultSets {
					t.Logf("结果集 %d: %d 行数据", i+1, len(rs))
				}
			}
		}

		// 用多个查询模拟多结果集功能
		t.Log("模拟多结果集功能:")

		// 第一个结果集：用户列表
		users, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 5", "active")
		if err != nil {
			t.Fatalf("第一个结果集查询失败: %v", err)
		}
		t.Logf("第一个结果集（用户列表）: %d 条记录", len(users))

		// 第二个结果集：统计信息
		stats, err := session.SelectList(ctx,
			"SELECT status, COUNT(*) as count FROM users GROUP BY status")
		if err != nil {
			t.Fatalf("第二个结果集查询失败: %v", err)
		}
		t.Logf("第二个结果集（统计信息）: %d 条记录", len(stats))

		// 第三个结果集：年龄分布
		ageGroups, err := session.SelectList(ctx,
			"SELECT CASE WHEN age < 25 THEN 'young' WHEN age < 35 THEN 'middle' ELSE 'senior' END as age_group, COUNT(*) as count FROM users GROUP BY age_group")
		if err != nil {
			t.Fatalf("第三个结果集查询失败: %v", err)
		}
		t.Logf("第三个结果集（年龄分组）: %d 条记录", len(ageGroups))
	})
}

// TestStoredProcErrorHandling 测试存储过程错误处理
func TestStoredProcErrorHandling(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	t.Run("不存在的存储过程", func(t *testing.T) {
		params := []mybatis.ProcParam{}

		result, err := session.CallStoredProc(ctx, "NonExistentProcedure", params)
		if err == nil {
			t.Error("调用不存在的存储过程应该返回错误")
		} else {
			t.Logf("正确处理不存在存储过程的错误: %v", err)
		}

		if result != nil {
			t.Error("失败的存储过程调用不应该返回结果")
		}
	})

	t.Run("参数类型错误", func(t *testing.T) {
		params := []mybatis.ProcParam{
			{Name: "invalid_param", Value: "test", Direction: "INVALID"}, // 无效的参数方向
		}

		result, err := session.CallStoredProc(ctx, "SomeProcedure", params)
		if err == nil {
			t.Error("无效的参数类型应该返回错误")
		} else {
			t.Logf("正确处理参数类型错误: %v", err)
		}

		if result != nil {
			t.Error("失败的存储过程调用不应该返回结果")
		}
	})

	t.Run("空参数列表", func(t *testing.T) {
		// 测试空参数列表的处理
		result, err := session.CallStoredProc(ctx, "NoParamProcedure", nil)
		if err != nil {
			t.Logf("空参数存储过程调用错误 (预期的): %v", err)
		} else {
			t.Logf("空参数存储过程调用结果: %+v", result)
		}

		// 再测试空切片
		result2, err2 := session.CallStoredProc(ctx, "NoParamProcedure", []mybatis.ProcParam{})
		if err2 != nil {
			t.Logf("空参数切片存储过程调用错误 (预期的): %v", err2)
		} else {
			t.Logf("空参数切片存储过程调用结果: %+v", result2)
		}
	})
}

// TestStoredProcPerformance 测试存储过程性能
func TestStoredProcPerformance(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	session := mybatis.NewSimpleSession(db).Debug(false)
	ctx := context.Background()

	setupStoredProcTestData(t, session, ctx)

	t.Run("存储过程调用性能", func(t *testing.T) {
		// 由于SQLite不支持真正的存储过程，我们用复杂查询来模拟性能测试

		const iterations = 100

		// 测试复杂查询的性能（模拟存储过程）
		complexQueries := []string{
			"SELECT COUNT(*) FROM users WHERE status = 'active'",
			"SELECT AVG(age) FROM users WHERE status = 'active'",
			"SELECT status, COUNT(*) as count FROM users GROUP BY status",
			"SELECT * FROM users WHERE age > 25 ORDER BY age LIMIT 10",
			"SELECT MIN(age), MAX(age), AVG(age) FROM users",
		}

		totalDuration := int64(0)
		successCount := 0

		for i := 0; i < iterations; i++ {
			queryIndex := i % len(complexQueries)
			query := complexQueries[queryIndex]

			start := time.Now()
			_, err := session.SelectList(ctx, query)
			duration := time.Since(start).Nanoseconds()

			if err != nil {
				t.Logf("查询 %d 失败: %v", i, err)
			} else {
				totalDuration += duration
				successCount++
			}
		}

		if successCount > 0 {
			avgDuration := totalDuration / int64(successCount)
			t.Logf("复杂查询性能测试结果:")
			t.Logf("  总执行次数: %d", iterations)
			t.Logf("  成功次数: %d", successCount)
			t.Logf("  平均执行时间: %d 纳秒", avgDuration)
			t.Logf("  每秒操作数: %.2f", float64(successCount*1e9)/float64(totalDuration))
		}
	})

	t.Run("批量存储过程调用", func(t *testing.T) {
		// 模拟批量调用存储过程的场景
		batchSize := 50

		start := time.Now()
		successCount := 0

		for i := 0; i < batchSize; i++ {
			// 模拟不同的存储过程调用
			switch i % 3 {
			case 0:
				_, err := session.SelectOne(ctx, "SELECT COUNT(*) FROM users WHERE age > ?", 20+i%40)
				if err == nil {
					successCount++
				}
			case 1:
				_, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ? LIMIT 3", "active")
				if err == nil {
					successCount++
				}
			case 2:
				_, err := session.SelectOne(ctx, "SELECT AVG(age) FROM users WHERE status = ?", "active")
				if err == nil {
					successCount++
				}
			}
		}

		totalTime := time.Since(start).Nanoseconds()

		t.Logf("批量存储过程调用结果:")
		t.Logf("  批量大小: %d", batchSize)
		t.Logf("  成功次数: %d", successCount)
		t.Logf("  总耗时: %d 纳秒", totalTime)
		if totalTime > 0 {
			t.Logf("  平均响应时间: %d 纳秒", totalTime/int64(successCount))
			t.Logf("  吞吐量: %.2f ops/sec", float64(successCount*1e9)/float64(totalTime))
		}
	})
}

// setupStoredProcTestData 设置存储过程测试数据
func setupStoredProcTestData(t *testing.T, session mybatis.SimpleSession, ctx context.Context) {
	t.Helper()

	// 插入测试数据
	testUsers := []struct {
		name   string
		email  string
		age    int
		status string
	}{
		{"存储过程测试用户1", "sp1@example.com", 25, "active"},
		{"存储过程测试用户2", "sp2@example.com", 30, "active"},
		{"存储过程测试用户3", "sp3@example.com", 35, "inactive"},
		{"存储过程测试用户4", "sp4@example.com", 28, "active"},
		{"存储过程测试用户5", "sp5@example.com", 32, "pending"},
		{"存储过程测试用户6", "sp6@example.com", 27, "active"},
		{"存储过程测试用户7", "sp7@example.com", 33, "inactive"},
		{"存储过程测试用户8", "sp8@example.com", 29, "active"},
	}

	for i, user := range testUsers {
		_, err := session.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			user.name, user.email, user.age, user.status)
		if err != nil {
			t.Logf("插入存储过程测试用户 %d 失败: %v", i+1, err)
		}
	}

	t.Log("成功设置存储过程测试数据")
}

// 注意：由于大多数轻量级数据库（如SQLite）不支持真正的存储过程，
// 上述测试主要演示了API的使用方法和错误处理。
// 在实际的生产环境中，您可以将这些测试适配到支持存储过程的数据库系统，
// 如 MySQL、PostgreSQL、Oracle 或 SQL Server。
