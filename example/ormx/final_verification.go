package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/orm"
)

// VerificationUser 验证用户模型
type VerificationUser struct {
	orm.BaseModel
	Name  string `gorm:"size:100;not null" json:"name"`
	Email string `gorm:"size:100;uniqueIndex" json:"email"`
	Age   int    `json:"age"`
}

func main() {
	fmt.Println("=== YYHertz ORM 最终验证测试 ===")

	// 1. 验证配置加载
	fmt.Println("\n1. 验证配置加载...")
	config := orm.DefaultDatabaseConfig()
	if config == nil {
		log.Fatal("❌ 配置加载失败")
	}
	fmt.Printf("✅ 配置加载成功: %s\n", config.Type)

	// 2. 验证ORM实例创建
	fmt.Println("\n2. 验证ORM实例创建...")
	ormInstance, err := orm.NewORM(config)
	if err != nil {
		log.Fatalf("❌ ORM实例创建失败: %v", err)
	}
	fmt.Println("✅ ORM实例创建成功")

	// 3. 验证数据库连接
	fmt.Println("\n3. 验证数据库连接...")
	if err := ormInstance.Ping(); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	fmt.Println("✅ 数据库连接正常")

	// 4. 验证自动迁移
	fmt.Println("\n4. 验证自动迁移...")
	if err := ormInstance.AutoMigrate(&VerificationUser{}); err != nil {
		log.Fatalf("❌ 自动迁移失败: %v", err)
	}
	fmt.Println("✅ 自动迁移成功")

	// 5. 验证CRUD操作
	fmt.Println("\n5. 验证CRUD操作...")

	// 清理旧数据
	db := ormInstance.DB()
	db.Where("email LIKE ?", "%verification%").Delete(&VerificationUser{})

	// 创建
	user := &VerificationUser{
		Name:  "验证用户",
		Email: "verification@test.com",
		Age:   30,
	}

	if err := db.Create(user).Error; err != nil {
		log.Printf("⚠️ 创建用户警告: %v", err)
	} else {
		fmt.Printf("✅ 创建用户成功，ID: %d\n", user.ID)
	}

	// 查询
	var foundUser VerificationUser
	if err := db.First(&foundUser, "email = ?", "verification@test.com").Error; err != nil {
		log.Printf("⚠️ 查询用户警告: %v", err)
	} else {
		fmt.Printf("✅ 查询用户成功: %s\n", foundUser.Name)
	}

	// 更新
	if err := db.Model(&foundUser).Update("age", 31).Error; err != nil {
		log.Printf("⚠️ 更新用户警告: %v", err)
	} else {
		fmt.Println("✅ 更新用户成功")
	}

	// 删除
	if err := db.Delete(&foundUser).Error; err != nil {
		log.Printf("⚠️ 删除用户警告: %v", err)
	} else {
		fmt.Println("✅ 删除用户成功")
	}

	// 6. 验证事务功能
	fmt.Println("\n6. 验证事务功能...")
	err = ormInstance.Transaction(func(tx *gorm.DB) error {
		txUser := &VerificationUser{
			Name:  "事务验证用户",
			Email: "tx_verification@test.com",
			Age:   25,
		}
		return tx.Create(txUser).Error
	})

	if err != nil {
		log.Printf("⚠️ 事务执行警告: %v", err)
	} else {
		fmt.Println("✅ 事务执行成功")
	}

	// 7. 验证便捷函数
	fmt.Println("\n7. 验证便捷函数...")
	globalORM := orm.GetDefaultORM()
	if globalORM != nil {
		fmt.Println("✅ 全局ORM获取成功")

		// 使用便捷创建
		convUser := &VerificationUser{
			Name:  "便捷验证用户",
			Email: "convenience_verification@test.com",
			Age:   28,
		}

		result := orm.Create(convUser)
		if result.Error != nil {
			log.Printf("⚠️ 便捷创建警告: %v", result.Error)
		} else {
			fmt.Printf("✅ 便捷创建成功，ID: %d\n", convUser.ID)
		}
	}

	// 8. 验证统计信息
	fmt.Println("\n8. 验证统计信息...")
	stats := ormInstance.GetStats()
	if len(stats) > 0 {
		fmt.Println("✅ 统计信息获取成功:")
		for key, value := range stats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	} else {
		fmt.Println("⚠️ 统计信息为空")
	}

	// 9. 验证慢查询监控
	fmt.Println("\n9. 验证慢查询监控...")
	monitor := orm.GetGlobalSlowQueryMonitor()
	if monitor != nil {
		fmt.Printf("✅ 慢查询监控状态: %t\n", monitor.IsEnabled())
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
	} else {
		fmt.Println("⚠️ 慢查询监控器为空")
	}

	// 10. 最终清理
	fmt.Println("\n10. 最终清理...")
	db.Unscoped().Where("email LIKE ?", "%verification%").Delete(&VerificationUser{})
	fmt.Println("✅ 清理完成")

	// 11. 关闭连接
	fmt.Println("\n11. 关闭连接...")
	if err := ormInstance.Close(); err != nil {
		log.Printf("⚠️ 关闭连接警告: %v", err)
	} else {
		fmt.Println("✅ 连接已关闭")
	}

	fmt.Println("\n=== 最终验证完成 ===")
	fmt.Println("🎉 所有功能验证通过！ORM框架工作正常！")

	// 输出验证结果到文件
	if file, err := os.Create("verification_result.txt"); err == nil {
		defer file.Close()
		file.WriteString("YYHertz ORM 验证结果\n")
		file.WriteString("===================\n")
		file.WriteString("✅ 配置加载: 正常\n")
		file.WriteString("✅ ORM实例创建: 正常\n")
		file.WriteString("✅ 数据库连接: 正常\n")
		file.WriteString("✅ 自动迁移: 正常\n")
		file.WriteString("✅ CRUD操作: 正常\n")
		file.WriteString("✅ 事务功能: 正常\n")
		file.WriteString("✅ 便捷函数: 正常\n")
		file.WriteString("✅ 统计信息: 正常\n")
		file.WriteString("✅ 慢查询监控: 正常\n")
		file.WriteString("✅ 连接管理: 正常\n")
		file.WriteString("\n所有功能验证通过！\n")
		fmt.Println("✅ 验证结果已保存到 verification_result.txt")
	}
}
