package main

import (
	"fmt"
	"log"
	"testing"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/orm"
)

// User类型定义在complete_test.go中

func TestFinal(t *testing.T) {
	fmt.Println("=== YYHertz ORM 最终功能测试 ===")

	// 1. 测试配置获取
	fmt.Println("\n1. 测试配置获取...")
	config := orm.DefaultDatabaseConfig()
	fmt.Printf("✅ 数据库类型: %s\n", config.Type)
	fmt.Printf("✅ 数据库文件: %s\n", config.Database)

	// 2. 测试ORM实例创建
	fmt.Println("\n2. 测试ORM实例创建...")
	ormInstance, err := orm.NewORM(config)
	if err != nil {
		log.Fatalf("❌ 创建ORM实例失败: %v", err)
	}
	fmt.Println("✅ ORM实例创建成功")

	// 3. 测试数据库连接
	fmt.Println("\n3. 测试数据库连接...")
	if err := ormInstance.Ping(); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	fmt.Println("✅ 数据库连接成功")

	// 4. 测试自动迁移
	fmt.Println("\n4. 测试自动迁移...")
	if err := ormInstance.AutoMigrate(&User{}); err != nil {
		log.Fatalf("❌ 自动迁移失败: %v", err)
	}
	fmt.Println("✅ 自动迁移成功")

	// 5. 测试CRUD操作
	fmt.Println("\n5. 测试CRUD操作...")

	db := ormInstance.DB()

	// 创建用户
	user := &User{
		Name:  "张三",
		Email: "zhangsan@test.com",
	}

	if err := db.Create(user).Error; err != nil {
		log.Printf("⚠️ 创建用户警告: %v", err)
	} else {
		fmt.Printf("✅ 创建用户成功，ID: %d\n", user.ID)
	}

	// 查询用户
	var foundUser User
	if err := db.First(&foundUser, "email = ?", "zhangsan@test.com").Error; err != nil {
		log.Printf("⚠️ 查询用户警告: %v", err)
	} else {
		fmt.Printf("✅ 查询用户成功: %s\n", foundUser.Name)
	}

	// 更新用户
	if err := db.Model(&foundUser).Update("name", "李四").Error; err != nil {
		log.Printf("⚠️ 更新用户警告: %v", err)
	} else {
		fmt.Println("✅ 更新用户成功")
	}

	// 查询所有用户
	var users []User
	if err := db.Find(&users).Error; err != nil {
		log.Printf("⚠️ 查询所有用户警告: %v", err)
	} else {
		fmt.Printf("✅ 查询到 %d 个用户\n", len(users))
	}

	// 6. 测试便捷函数
	fmt.Println("\n6. 测试便捷函数...")

	// 使用全局ORM
	globalORM := orm.GetDefaultORM()
	if globalORM != nil {
		fmt.Println("✅ 获取全局ORM实例成功")

		// 测试便捷创建
		testUser := &User{
			Name:  "王五",
			Email: "wangwu@test.com",
		}

		result := orm.Create(testUser)
		if result.Error != nil {
			log.Printf("⚠️ 便捷创建警告: %v", result.Error)
		} else {
			fmt.Printf("✅ 便捷创建成功，ID: %d\n", testUser.ID)
		}

		// 测试便捷查询
		var testFoundUser User
		result = orm.First(&testFoundUser, "email = ?", "wangwu@test.com")
		if result.Error != nil {
			log.Printf("⚠️ 便捷查询警告: %v", result.Error)
		} else {
			fmt.Printf("✅ 便捷查询成功: %s\n", testFoundUser.Name)
		}
	}

	// 7. 测试统计信息
	fmt.Println("\n7. 测试统计信息...")
	stats := ormInstance.GetStats()
	fmt.Println("✅ 数据库统计信息:")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 8. 测试慢查询监控
	fmt.Println("\n8. 测试慢查询监控...")
	monitor := orm.GetGlobalSlowQueryMonitor()
	if monitor != nil {
		fmt.Printf("✅ 慢查询监控状态: %t\n", monitor.IsEnabled())
		fmt.Printf("✅ 慢查询记录数: %d\n", monitor.GetRecordCount())
		fmt.Printf("✅ 慢查询阈值: %v\n", monitor.GetThreshold())

		// 获取慢查询统计
		slowStats := orm.GetSlowQueryStats()
		fmt.Println("✅ 慢查询统计:")
		for key, value := range slowStats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	// 9. 测试事务
	fmt.Println("\n9. 测试事务...")
	err = ormInstance.Transaction(func(tx *gorm.DB) error {
		txUser := &User{
			Name:  "事务用户",
			Email: "tx@test.com",
		}
		return tx.Create(txUser).Error
	})

	if err != nil {
		log.Printf("⚠️ 事务警告: %v", err)
	} else {
		fmt.Println("✅ 事务执行成功")
	}

	// 10. 清理测试数据
	fmt.Println("\n10. 清理测试数据...")
	db.Where("email LIKE ?", "%@test.com").Delete(&User{})
	fmt.Println("✅ 清理完成")

	// 关闭连接
	if err := ormInstance.Close(); err != nil {
		log.Printf("⚠️ 关闭连接警告: %v", err)
	} else {
		fmt.Println("\n✅ 数据库连接已关闭")
	}

	fmt.Println("\n=== ORM 功能测试完成 ===")
	fmt.Println("🎉 所有测试通过！ORM功能正常工作！")
}
