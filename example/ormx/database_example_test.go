package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/orm"
)

// User 用户模型示例
func RunDatabaseExampleExt() {
	fmt.Println("=== YYHertz ORM 数据库示例 ===")

	// 1. 获取默认ORM实例
	ormInstance := orm.GetDefaultORM()
	defer func() {
		if err := ormInstance.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}()

	// 2. 自动迁移模型
	if err := ormInstance.AutoMigrate(&SimpleUser{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 3. 演示基本的CRUD操作
	demonstrateBasicOperations()

	// 4. 演示事务操作
	demonstrateTransactions()

	// 5. 演示数据库统计
	demonstrateStats()

	// 6. 演示上下文操作
	demonstrateContext()

	// 7. 演示原生SQL
	demonstrateRawSQL()

	// 8. 演示慢查询监控
	demonstrateSlowQuery()

	fmt.Println("=== 示例执行完成 ===")
}

// demonstrateBasicOperations 演示基本操作
func demonstrateBasicOperations() {
	fmt.Println("\n--- 基本CRUD操作演示 ---")

	// 创建用户
	user := &SimpleUser{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}

	// 使用便捷函数创建
	result := orm.Create(user)
	if result.Error != nil {
		log.Printf("创建用户失败: %v", result.Error)
		return
	}
	fmt.Printf("创建用户成功, ID: %d\n", user.ID)

	// 查询用户
	var foundUser SimpleUser
	result = orm.First(&foundUser, user.ID)
	if result.Error != nil {
		log.Printf("查询用户失败: %v", result.Error)
		return
	}
	fmt.Printf("查询用户成功: %+v\n", foundUser)

	// 更新用户
	result = orm.Where("id = ?", user.ID).Updates(&SimpleUser{Age: 26})
	if result.Error != nil {
		log.Printf("更新用户失败: %v", result.Error)
	} else {
		fmt.Println("更新用户成功")
	}

	// 查询所有用户
	var users []SimpleUser
	result = orm.Find(&users)
	if result.Error != nil {
		log.Printf("查询所有用户失败: %v", result.Error)
	} else {
		fmt.Printf("查询到 %d 个用户\n", len(users))
	}
}

// demonstrateTransactions 演示事务操作
func demonstrateTransactions() {
	fmt.Println("\n--- 事务操作演示 ---")

	ormInstance := orm.GetDefaultORM()

	// 成功事务
	err := ormInstance.Transaction(func(tx *gorm.DB) error {
		user1 := &SimpleUser{
			Name:  "事务用户1",
			Email: "tx1@example.com",
			Age:   20,
		}
		if err := tx.Create(user1).Error; err != nil {
			return err
		}

		user2 := &SimpleUser{
			Name:  "事务用户2",
			Email: "tx2@example.com",
			Age:   25,
		}
		if err := tx.Create(user2).Error; err != nil {
			return err
		}

		fmt.Println("事务中创建了两个用户")
		return nil
	})

	if err != nil {
		log.Printf("事务执行失败: %v", err)
	} else {
		fmt.Println("事务执行成功")
	}

	// 失败事务（回滚）
	err = ormInstance.Transaction(func(tx *gorm.DB) error {
		user3 := &SimpleUser{
			Name:  "事务用户3",
			Email: "tx3@example.com",
			Age:   30,
		}
		if err := tx.Create(user3).Error; err != nil {
			return err
		}

		// 故意返回错误来触发回滚
		return fmt.Errorf("故意的错误")
	})

	if err != nil {
		fmt.Printf("事务回滚成功: %v\n", err)
	}
}

// demonstrateStats 演示数据库统计
func demonstrateStats() {
	fmt.Println("\n--- 数据库统计演示 ---")

	ormInstance := orm.GetDefaultORM()
	stats := ormInstance.GetStats()

	fmt.Println("数据库统计信息:")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// demonstrateContext 演示上下文操作
func demonstrateContext() {
	fmt.Println("\n--- 上下文操作演示 ---")

	ormInstance := orm.GetDefaultORM()

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用上下文进行查询
	var users []SimpleUser
	if err := ormInstance.WithContext(ctx).Limit(10).Find(&users).Error; err != nil {
		log.Printf("上下文查询失败: %v", err)
	} else {
		fmt.Printf("使用上下文查询到 %d 条用户记录\n", len(users))
	}
}

// demonstrateRawSQL 演示原生SQL
func demonstrateRawSQL() {
	fmt.Println("\n--- 原生SQL演示 ---")

	// 测试原生查询
	var result struct {
		Count int `json:"count"`
	}

	if err := orm.Raw("SELECT COUNT(*) as count FROM users").Scan(&result).Error; err != nil {
		log.Printf("原生查询失败: %v", err)
	} else {
		fmt.Printf("用户总数: %d\n", result.Count)
	}

	// 执行原生SQL
	if err := orm.Exec("UPDATE users SET age = age + 1 WHERE age < 30").Error; err != nil {
		log.Printf("原生SQL执行失败: %v", err)
	} else {
		fmt.Println("原生SQL执行成功")
	}
}

// demonstrateSlowQuery 演示慢查询监控
func demonstrateSlowQuery() {
	fmt.Println("\n--- 慢查询监控演示 ---")

	// 获取慢查询监控器
	monitor := orm.GetGlobalSlowQueryMonitor()

	// 设置较低的阈值来演示慢查询检测
	monitor.SetThreshold(time.Millisecond * 10)

	// 执行一些查询来触发慢查询记录
	var users []SimpleUser
	orm.Find(&users)

	// 模拟一个可能较慢的查询
	orm.Raw("SELECT * FROM users WHERE name LIKE '%test%'").Scan(&users)

	// 打印慢查询统计
	fmt.Println("慢查询统计:")
	stats := orm.GetSlowQueryStats()
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 打印前5条最慢的查询
	orm.PrintSlowQueryStats(5)

	// 恢复默认阈值
	monitor.SetThreshold(time.Millisecond * 500)
}

// demonstrateErrorHandling 演示错误处理
func demonstrateErrorHandling() {
	fmt.Println("\n--- 错误处理演示 ---")

	// 1. 演示记录未找到错误
	var user SimpleUser
	result := orm.First(&user, 99999) // 查询不存在的记录
	if result.Error != nil {
		fmt.Printf("捕获错误: %v\n", result.Error)
	}

	// 2. 演示约束错误（重复键）
	user1 := &SimpleUser{
		Name:  "测试用户",
		Email: "test@example.com",
		Age:   25,
	}
	orm.Create(user1) // 第一次创建

	user2 := &SimpleUser{
		Name:  "测试用户2",
		Email: "test@example.com", // 相同邮箱，应该触发唯一约束错误
		Age:   30,
	}
	result = orm.Create(user2)
	if result.Error != nil {
		fmt.Printf("约束错误: %v\n", result.Error)
	}

	// 3. 演示验证错误（缺少WHERE子句的更新）
	result = orm.Updates(&SimpleUser{Age: 100}) // 没有WHERE条件的全局更新
	if result.Error != nil {
		fmt.Printf("验证错误: %v\n", result.Error)
	}

	// 4. 演示连接检查
	ormInstance := orm.GetDefaultORM()
	if err := ormInstance.Ping(); err != nil {
		fmt.Printf("连接错误: %v\n", err)
	} else {
		fmt.Println("数据库连接正常")
	}
}

// demonstrateAdvancedQueries 演示高级查询
func demonstrateAdvancedQueries() {
	fmt.Println("\n--- 高级查询演示 ---")

	ormInstance := orm.GetDefaultORM()
	db := ormInstance.DB()

	// 条件查询
	var users []SimpleUser
	db.Where("age > ?", 20).Where("name LIKE ?", "%用户%").Find(&users)
	fmt.Printf("条件查询结果: %d 条记录\n", len(users))

	// 排序查询
	db.Order("age DESC").Limit(5).Find(&users)
	fmt.Printf("排序查询结果: %d 条记录\n", len(users))

	// 分页查询
	var count int64
	db.Model(&SimpleUser{}).Count(&count)
	fmt.Printf("总记录数: %d\n", count)

	db.Offset(0).Limit(10).Find(&users)
	fmt.Printf("分页查询结果: %d 条记录\n", len(users))

	// 聚合查询
	var avgAge float64
	db.Model(&SimpleUser{}).Select("AVG(age)").Row().Scan(&avgAge)
	fmt.Printf("平均年龄: %.2f\n", avgAge)
}

// demonstrateModelHooks 演示模型钩子
func demonstrateModelHooks() {
	fmt.Println("\n--- 模型钩子演示 ---")

	// BaseModel 已经包含了 BeforeCreate 和 BeforeUpdate 钩子
	user := &SimpleUser{
		Name:  "钩子测试用户",
		Email: "hooks@example.com",
		Age:   25,
	}

	// 创建时会自动设置 CreatedAt 和 UpdatedAt
	result := orm.Create(user)
	if result.Error == nil {
		fmt.Printf("创建用户成功，创建时间: %v\n", user.CreatedAt)
	}

	// 更新时会自动更新 UpdatedAt
	time.Sleep(time.Second) // 等待一秒确保时间不同
	result = orm.Where("id = ?", user.ID).Updates(&SimpleUser{Age: 26})
	if result.Error == nil {
		// 重新查询获取更新后的时间
		var updatedUser SimpleUser
		orm.First(&updatedUser, user.ID)
		fmt.Printf("更新用户成功，更新时间: %v\n", updatedUser.UpdatedAt)
	}
}

// demonstrateCleanup 演示清理操作
func demonstrateCleanup() {
	fmt.Println("\n--- 清理测试数据 ---")

	// 删除测试数据
	result := orm.Where("email LIKE ?", "%@example.com").Delete(&SimpleUser{})
	if result.Error != nil {
		log.Printf("清理数据失败: %v", result.Error)
	} else {
		fmt.Printf("清理了 %d 条测试数据\n", result.RowsAffected)
	}
}

// main 函数，程序入口点
func TestExample(t *testing.T) {
	RunDatabaseExampleExt()
	demonstrateErrorHandling()
	demonstrateAdvancedQueries()
	demonstrateModelHooks()
	demonstrateCleanup()
}
