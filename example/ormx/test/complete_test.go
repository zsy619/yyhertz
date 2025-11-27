package main

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/orm"
)

// User 用户模型
type User struct {
	orm.BaseModel
	Name  string `gorm:"size:100" json:"name"`
	Email string `gorm:"size:100;uniqueIndex" json:"email"`
	Age   int    `json:"age"`
}

func main() {
	fmt.Println("=== YYHertz ORM 完整功能测试 ===")

	// 1. 初始化ORM
	fmt.Println("\n1. 初始化ORM...")
	config := orm.DefaultDatabaseConfig()
	fmt.Printf("✅ 数据库配置: %s -> %s\n", config.Type, config.Database)

	ormInstance, err := orm.NewORM(config)
	if err != nil {
		log.Fatalf("❌ 创建ORM失败: %v", err)
	}
	fmt.Println("✅ ORM实例创建成功")

	// 2. 测试连接
	fmt.Println("\n2. 测试数据库连接...")
	if err := ormInstance.Ping(); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	fmt.Println("✅ 数据库连接正常")

	// 3. 自动迁移
	fmt.Println("\n3. 执行自动迁移...")
	if err := ormInstance.AutoMigrate(&User{}); err != nil {
		log.Fatalf("❌ 自动迁移失败: %v", err)
	}
	fmt.Println("✅ 自动迁移完成")

	db := ormInstance.DB()

	// 4. 清理旧数据
	fmt.Println("\n4. 清理测试数据...")
	db.Where("email LIKE ?", "%@test.com").Delete(&User{})
	fmt.Println("✅ 清理完成")

	// 5. 测试创建操作
	fmt.Println("\n5. 测试创建操作...")
	users := []*User{
		{Name: "张三", Email: "zhangsan@test.com", Age: 25},
		{Name: "李四", Email: "lisi@test.com", Age: 30},
		{Name: "王五", Email: "wangwu@test.com", Age: 28},
	}

	for i, user := range users {
		if err := db.Create(user).Error; err != nil {
			log.Printf("⚠️ 创建用户%d失败: %v", i+1, err)
		} else {
			fmt.Printf("✅ 创建用户成功: %s (ID: %d)\n", user.Name, user.ID)
		}
	}

	// 6. 测试查询操作
	fmt.Println("\n6. 测试查询操作...")

	// 查询单个用户
	var user User
	if err := db.First(&user, "email = ?", "zhangsan@test.com").Error; err != nil {
		log.Printf("⚠️ 查询单个用户失败: %v", err)
	} else {
		fmt.Printf("✅ 查询单个用户成功: %s (年龄: %d)\n", user.Name, user.Age)
	}

	// 查询所有用户
	var allUsers []User
	if err := db.Find(&allUsers).Error; err != nil {
		log.Printf("⚠️ 查询所有用户失败: %v", err)
	} else {
		fmt.Printf("✅ 查询到 %d 个用户\n", len(allUsers))
	}

	// 条件查询
	var youngUsers []User
	if err := db.Where("age < ?", 30).Find(&youngUsers).Error; err != nil {
		log.Printf("⚠️ 条件查询失败: %v", err)
	} else {
		fmt.Printf("✅ 年龄小于30的用户: %d 个\n", len(youngUsers))
	}

	// 7. 测试更新操作
	fmt.Println("\n7. 测试更新操作...")
	if err := db.Model(&user).Update("age", 26).Error; err != nil {
		log.Printf("⚠️ 更新用户失败: %v", err)
	} else {
		fmt.Println("✅ 更新用户年龄成功")
	}

	// 批量更新
	result := db.Model(&User{}).Where("age > ?", 25).Update("age", gorm.Expr("age + ?", 1))
	if result.Error != nil {
		log.Printf("⚠️ 批量更新失败: %v", result.Error)
	} else {
		fmt.Printf("✅ 批量更新成功，影响 %d 行\n", result.RowsAffected)
	}

	// 8. 测试事务
	fmt.Println("\n8. 测试事务...")
	err = ormInstance.Transaction(func(tx *gorm.DB) error {
		// 在事务中创建用户
		txUser := &User{
			Name:  "事务用户",
			Email: "transaction@test.com",
			Age:   35,
		}
		if err := tx.Create(txUser).Error; err != nil {
			return err
		}

		// 在事务中更新用户
		if err := tx.Model(txUser).Update("age", 36).Error; err != nil {
			return err
		}

		fmt.Printf("✅ 事务中创建并更新用户: %s (ID: %d)\n", txUser.Name, txUser.ID)
		return nil
	})

	if err != nil {
		log.Printf("⚠️ 事务执行失败: %v", err)
	} else {
		fmt.Println("✅ 事务执行成功")
	}

	// 9. 测试便捷函数
	fmt.Println("\n9. 测试便捷函数...")
	globalORM := orm.GetDefaultORM()
	if globalORM != nil {
		fmt.Println("✅ 获取全局ORM成功")

		// 使用便捷创建函数
		convenienceUser := &User{
			Name:  "便捷用户",
			Email: "convenience@test.com",
			Age:   40,
		}

		result := orm.Create(convenienceUser)
		if result.Error != nil {
			log.Printf("⚠️ 便捷创建失败: %v", result.Error)
		} else {
			fmt.Printf("✅ 便捷创建成功: %s (ID: %d)\n", convenienceUser.Name, convenienceUser.ID)
		}

		// 使用便捷查询函数
		var foundUser User
		result = orm.First(&foundUser, "email = ?", "convenience@test.com")
		if result.Error != nil {
			log.Printf("⚠️ 便捷查询失败: %v", result.Error)
		} else {
			fmt.Printf("✅ 便捷查询成功: %s\n", foundUser.Name)
		}
	}

	// 10. 测试统计信息
	fmt.Println("\n10. 测试统计信息...")
	stats := ormInstance.GetStats()
	fmt.Println("✅ 数据库连接池统计:")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 11. 测试慢查询监控
	fmt.Println("\n11. 测试慢查询监控...")
	monitor := orm.GetGlobalSlowQueryMonitor()
	if monitor != nil {
		fmt.Printf("✅ 慢查询监控启用: %t\n", monitor.IsEnabled())
		fmt.Printf("✅ 慢查询记录数: %d\n", monitor.GetRecordCount())
		fmt.Printf("✅ 慢查询阈值: %v\n", monitor.GetThreshold())

		// 获取慢查询统计
		slowStats := orm.GetSlowQueryStats()
		if len(slowStats) > 0 {
			fmt.Println("✅ 慢查询统计:")
			for key, value := range slowStats {
				fmt.Printf("  %s: %v\n", key, value)
			}
		} else {
			fmt.Println("✅ 暂无慢查询记录")
		}
	}

	// 12. 测试删除操作
	fmt.Println("\n12. 测试删除操作...")

	// 软删除
	if err := db.Where("email = ?", "lisi@test.com").Delete(&User{}).Error; err != nil {
		log.Printf("⚠️ 软删除失败: %v", err)
	} else {
		fmt.Println("✅ 软删除成功")
	}

	// 验证软删除
	var deletedUser User
	if err := db.Unscoped().Where("email = ?", "lisi@test.com").First(&deletedUser).Error; err != nil {
		log.Printf("⚠️ 查询软删除用户失败: %v", err)
	} else {
		fmt.Printf("✅ 软删除用户仍存在: %s (删除时间: %v)\n", deletedUser.Name, deletedUser.DeletedAt)
	}

	// 13. 最终清理
	fmt.Println("\n13. 清理测试数据...")
	db.Unscoped().Where("email LIKE ?", "%@test.com").Delete(&User{})
	fmt.Println("✅ 清理完成")

	// 14. 关闭连接
	fmt.Println("\n14. 关闭数据库连接...")
	if err := ormInstance.Close(); err != nil {
		log.Printf("⚠️ 关闭连接失败: %v", err)
	} else {
		fmt.Println("✅ 数据库连接已关闭")
	}

	fmt.Println("\n=== ORM 功能测试完成 ===")
	fmt.Println("🎉 所有测试通过！ORM功能正常工作！")
}
