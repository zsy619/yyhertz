// Package main 示例代码
//
// 演示MyBatis框架的各种使用方式：
// 1. 基本CRUD操作
// 2. XML映射使用
// 3. 动态SQL构建
// 4. 批量操作
// 5. 事务处理
// 6. 缓存使用
// 7. 性能监控
// 8. 高级特性
package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleUser 示例用户模型
type ExampleUser struct {
	ID       int64     `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Email    string    `json:"email" db:"email"`
	Age      int       `json:"age" db:"age"`
	Status   string    `json:"status" db:"status"`
	CreateAt time.Time `json:"createAt" db:"created_at"`
}

// ExampleBasicCRUD 基本CRUD操作示例
func ExampleBasicCRUD() {
	fmt.Println("=== 基本CRUD操作示例 ===")

	// 1. 创建配置
	config := DefaultConfig()

	// 2. 创建Session
	session, err := CreateSimpleSession(config)
	if err != nil {
		log.Fatalf("创建Session失败: %v", err)
	}

	ctx := context.Background()

	// 3. 插入数据
	fmt.Println("插入用户...")
	userID, err := session.Insert(ctx,
		"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
		"张三", "zhangsan@example.com", 25, "active")
	if err != nil {
		log.Printf("插入失败: %v", err)
		return
	}
	fmt.Printf("插入成功，ID: %d\n", userID)

	// 4. 查询单个记录
	fmt.Println("查询用户...")
	user, err := session.SelectOne(ctx,
		"SELECT * FROM users WHERE id = ?", userID)
	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}

	if user != nil {
		userMap := user.(map[string]interface{})
		fmt.Printf("查询结果: %+v\n", userMap)
	}

	// 5. 更新数据
	fmt.Println("更新用户...")
	affected, err := session.Update(ctx,
		"UPDATE users SET age = ? WHERE id = ?", 26, userID)
	if err != nil {
		log.Printf("更新失败: %v", err)
		return
	}
	fmt.Printf("更新成功，影响行数: %d\n", affected)

	// 6. 查询列表
	fmt.Println("查询用户列表...")
	users, err := session.SelectList(ctx,
		"SELECT * FROM users WHERE status = ?", "active")
	if err != nil {
		log.Printf("查询列表失败: %v", err)
		return
	}
	fmt.Printf("查询到 %d 个活跃用户\n", len(users))

	// 7. 删除数据
	fmt.Println("删除用户...")
	affected, err = session.Delete(ctx,
		"DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		log.Printf("删除失败: %v", err)
		return
	}
	fmt.Printf("删除成功，影响行数: %d\n", affected)
}

// ExampleXMLMapping XML映射示例
func ExampleXMLMapping() {
	fmt.Println("\n=== XML映射示例 ===")

	// 1. 创建XMLSession
	config := DefaultConfig()
	xmlSession, err := CreateXMLSession(config)
	if err != nil {
		log.Fatalf("创建XMLSession失败: %v", err)
	}

	// 2. 加载XML映射
	mapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="User">
    <select id="selectById" parameterType="map" resultType="map">
        SELECT * FROM users WHERE id = #{id}
    </select>
    
    <select id="selectByStatus" parameterType="map" resultType="map">
        SELECT * FROM users WHERE status = #{status} 
        <if test="limit != null">
            LIMIT #{limit}
        </if>
    </select>
    
    <insert id="insertUser" parameterType="map">
        INSERT INTO users (name, email, age, status) 
        VALUES (#{name}, #{email}, #{age}, #{status})
    </insert>
</mapper>`

	err = xmlSession.LoadMapperXMLFromString(mapperXML)
	if err != nil {
		log.Printf("加载XML映射失败: %v", err)
		return
	}

	ctx := context.Background()

	// 3. 使用XML映射插入数据
	fmt.Println("通过XML映射插入用户...")
	insertParams := map[string]interface{}{
		"name":   "李四",
		"email":  "lisi@example.com",
		"age":    30,
		"status": "active",
	}

	userID, err := xmlSession.InsertByID(ctx, "User.insertUser", insertParams)
	if err != nil {
		log.Printf("XML插入失败: %v", err)
		return
	}
	fmt.Printf("XML插入成功，ID: %d\n", userID)

	// 4. 使用XML映射查询数据
	fmt.Println("通过XML映射查询用户...")
	queryParams := map[string]interface{}{
		"id": userID,
	}

	user, err := xmlSession.SelectOneByID(ctx, "User.selectById", queryParams)
	if err != nil {
		log.Printf("XML查询失败: %v", err)
		return
	}

	if user != nil {
		userMap := user.(map[string]interface{})
		fmt.Printf("XML查询结果: %+v\n", userMap)
	}

	// 5. 动态条件查询
	fmt.Println("通过XML映射动态查询...")
	dynamicParams := map[string]interface{}{
		"status": "active",
		"limit":  5,
	}

	users, err := xmlSession.SelectListByID(ctx, "User.selectByStatus", dynamicParams)
	if err != nil {
		log.Printf("动态查询失败: %v", err)
		return
	}
	fmt.Printf("动态查询到 %d 个用户\n", len(users))
}

// ExampleBatchOperations 批量操作示例
func ExampleBatchOperations() {
	fmt.Println("\n=== 批量操作示例 ===")

	config := DefaultConfig()
	session, err := CreateSimpleSession(config)
	if err != nil {
		log.Fatalf("创建Session失败: %v", err)
	}

	ctx := context.Background()

	// 准备批量数据
	batchData := [][]interface{}{
		{"王五", "wangwu@example.com", 28, "active"},
		{"赵六", "zhaoliu@example.com", 32, "active"},
		{"孙七", "sunqi@example.com", 25, "inactive"},
		{"周八", "zhouba@example.com", 29, "active"},
		{"吴九", "wujiu@example.com", 27, "pending"},
	}

	// 批量插入
	fmt.Println("批量插入用户...")
	start := time.Now()
	affected, err := session.BatchInsert(ctx,
		"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
		batchData)
	duration := time.Since(start)

	if err != nil {
		log.Printf("批量插入失败: %v", err)
		return
	}

	fmt.Printf("批量插入成功: %d 行数据，耗时: %v\n", affected, duration)

	// 批量更新
	fmt.Println("批量更新用户状态...")
	updateData := [][]interface{}{
		{"active", "inactive"},
		{"pending", "active"},
	}

	start = time.Now()
	totalAffected := int64(0)
	for _, params := range updateData {
		affected, err := session.Update(ctx,
			"UPDATE users SET status = ? WHERE status = ?", params...)
		if err != nil {
			log.Printf("批量更新失败: %v", err)
			continue
		}
		totalAffected += affected
	}
	duration = time.Since(start)

	fmt.Printf("批量更新完成: %d 行数据，耗时: %v\n", totalAffected, duration)
}

// ExampleTransactionHandling 事务处理示例
func ExampleTransactionHandling() {
	fmt.Println("\n=== 事务处理示例 ===")

	config := DefaultConfig()
	session, err := CreateSimpleSession(config)
	if err != nil {
		log.Fatalf("创建Session失败: %v", err)
	}

	ctx := context.Background()

	// 模拟事务操作（注意：当前版本可能没有事务API，这里演示概念）
	fmt.Println("模拟事务操作...")

	// 开始事务操作
	userID, err := session.Insert(ctx,
		"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
		"事务用户", "tx@example.com", 30, "active")
	if err != nil {
		log.Printf("插入用户失败: %v", err)
		return
	}

	// 验证插入结果
	user, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
	if err != nil {
		log.Printf("验证用户失败: %v", err)
		return
	}

	if user != nil {
		fmt.Printf("事务操作成功，用户ID: %d\n", userID)
	}
}

// ExampleCacheUsage 缓存使用示例
func ExampleCacheUsage() {
	fmt.Println("\n=== 缓存使用示例 ===")

	// 创建启用缓存的配置
	config := DefaultConfig()
	config.Cache.L1Enabled = true
	config.Cache.L1Size = 100
	config.Cache.L1TTL = 5 * time.Minute

	session, err := CreateSimpleSession(config)
	if err != nil {
		log.Fatalf("创建启用缓存的Session失败: %v", err)
	}

	ctx := context.Background()

	// 准备测试数据
	userID, err := session.Insert(ctx,
		"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
		"缓存测试用户", "cache@example.com", 25, "active")
	if err != nil {
		log.Printf("准备缓存测试数据失败: %v", err)
		return
	}

	// 第一次查询（从数据库）
	fmt.Println("第一次查询（从数据库）...")
	start := time.Now()
	user1, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
	firstDuration := time.Since(start)

	if err != nil {
		log.Printf("第一次查询失败: %v", err)
		return
	}

	// 第二次相同查询（从缓存）
	fmt.Println("第二次查询（从缓存）...")
	start = time.Now()
	user2, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
	secondDuration := time.Since(start)

	if err != nil {
		log.Printf("第二次查询失败: %v", err)
		return
	}

	// 比较查询时间
	fmt.Printf("查询时间对比:\n")
	fmt.Printf("  第一次查询: %v\n", firstDuration)
	fmt.Printf("  第二次查询: %v\n", secondDuration)

	if secondDuration < firstDuration {
		speedup := float64(firstDuration) / float64(secondDuration)
		fmt.Printf("  缓存命中，速度提升: %.2fx\n", speedup)
	}

	// 验证数据一致性
	user1Map := user1.(map[string]interface{})
	user2Map := user2.(map[string]interface{})

	if user1Map["name"] == user2Map["name"] {
		fmt.Println("缓存数据一致性验证通过")
	} else {
		fmt.Println("缓存数据一致性验证失败")
	}
}

// ExamplePerformanceMonitoring 性能监控示例
func ExamplePerformanceMonitoring() {
	fmt.Println("\n=== 性能监控示例 ===")

	config := DefaultConfig()
	config.Log.ShowSQL = true
	config.Log.Level = "debug"

	session, err := CreateSimpleSession(config)
	if err != nil {
		log.Fatalf("创建Session失败: %v", err)
	}

	ctx := context.Background()

	// 执行一系列操作并监控性能
	operations := []struct {
		name string
		op   func() error
	}{
		{"插入操作", func() error {
			_, err := session.Insert(ctx,
				"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
				"监控用户", "monitor@example.com", 30, "active")
			return err
		}},
		{"查询操作", func() error {
			_, err := session.SelectList(ctx, "SELECT * FROM users WHERE status = ?", "active")
			return err
		}},
		{"聚合操作", func() error {
			_, err := session.SelectOne(ctx, "SELECT COUNT(*) as count FROM users")
			return err
		}},
		{"更新操作", func() error {
			_, err := session.Update(ctx, "UPDATE users SET age = age + 1 WHERE status = ?", "active")
			return err
		}},
	}

	fmt.Println("执行操作并监控性能:")
	totalTime := time.Duration(0)

	for _, operation := range operations {
		start := time.Now()
		err := operation.op()
		duration := time.Since(start)
		totalTime += duration

		if err != nil {
			fmt.Printf("  %s: 失败 - %v\n", operation.name, err)
		} else {
			fmt.Printf("  %s: 成功 - 耗时 %v\n", operation.name, duration)
		}
	}

	fmt.Printf("总耗时: %v\n", totalTime)
	fmt.Printf("平均耗时: %v\n", totalTime/time.Duration(len(operations)))
}

// RunExamples 运行所有示例
func RunExamples() {
	fmt.Println("MyBatis框架使用示例")
	fmt.Println("=====================")

	// 运行基本CRUD示例
	ExampleBasicCRUD()

	// 运行XML映射示例
	ExampleXMLMapping()

	// 运行批量操作示例
	ExampleBatchOperations()

	// 运行事务处理示例
	ExampleTransactionHandling()

	// 运行缓存使用示例
	ExampleCacheUsage()

	// 运行性能监控示例
	ExamplePerformanceMonitoring()

	fmt.Println("\n=== 所有示例运行完成 ===")
}
