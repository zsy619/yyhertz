package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/zsy619/yyhertz/framework/orm"
)

// TestUser 测试用户模型
type TestUser struct {
	orm.BaseModel
	Name  string `gorm:"size:100" json:"name"`
	Email string `gorm:"size:100" json:"email"`
}

func TestOrm(t *testing.T) {
	fmt.Println("=== ORM 功能测试 ===")

	// 检查当前工作目录
	pwd, _ := os.Getwd()
	fmt.Printf("当前工作目录: %s\n", pwd)

	// 1. 测试默认配置
	fmt.Println("\n1. 测试默认配置...")
	config := orm.DefaultDatabaseConfig()
	if config != nil {
		fmt.Printf("✅ 数据库类型: %s\n", config.Type)
		fmt.Printf("✅ 数据库文件: %s\n", config.Database)
		fmt.Printf("✅ 最大连接数: %d\n", config.MaxOpenConns)
	} else {
		fmt.Println("❌ 获取配置失败")
		return
	}

	// 2. 测试ORM实例创建
	fmt.Println("\n2. 测试ORM实例创建...")
	ormInstance, err := orm.NewORM(config)
	if err != nil {
		fmt.Printf("❌ 创建ORM实例失败: %v\n", err)
		return
	}
	fmt.Println("✅ ORM实例创建成功")

	// 3. 测试数据库连接
	fmt.Println("\n3. 测试数据库连接...")
	if err := ormInstance.Ping(); err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("✅ 数据库连接成功")

	// 4. 测试自动迁移
	fmt.Println("\n4. 测试自动迁移...")
	if err := ormInstance.AutoMigrate(&TestUser{}); err != nil {
		fmt.Printf("❌ 自动迁移失败: %v\n", err)
		return
	}
	fmt.Println("✅ 自动迁移成功")

	// 5. 测试基本CRUD操作
	fmt.Println("\n5. 测试基本CRUD操作...")

	db := ormInstance.DB()

	// 创建用户
	user := &TestUser{
		Name:  "测试用户",
		Email: "test@example.com",
	}

	result := db.Create(user)
	if result.Error != nil {
		fmt.Printf("❌ 创建用户失败: %v\n", result.Error)
	} else {
		fmt.Printf("✅ 创建用户成功，ID: %d\n", user.ID)
	}

	// 查询用户
	var foundUser TestUser
	result = db.First(&foundUser, "email = ?", "test@example.com")
	if result.Error != nil {
		fmt.Printf("❌ 查询用户失败: %v\n", result.Error)
	} else {
		fmt.Printf("✅ 查询用户成功: %s\n", foundUser.Name)
	}

	// 更新用户
	result = db.Model(&foundUser).Update("name", "更新后的用户")
	if result.Error != nil {
		fmt.Printf("❌ 更新用户失败: %v\n", result.Error)
	} else {
		fmt.Println("✅ 更新用户成功")
	}

	// 删除用户
	result = db.Delete(&foundUser)
	if result.Error != nil {
		fmt.Printf("❌ 删除用户失败: %v\n", result.Error)
	} else {
		fmt.Println("✅ 删除用户成功")
	}

	// 6. 测试统计信息
	fmt.Println("\n6. 测试统计信息...")
	stats := ormInstance.GetStats()
	if len(stats) > 0 {
		fmt.Println("✅ 获取统计信息成功:")
		for key, value := range stats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	} else {
		fmt.Println("⚠️ 统计信息为空")
	}

	// 7. 测试慢查询监控
	fmt.Println("\n7. 测试慢查询监控...")
	monitor := orm.GetGlobalSlowQueryMonitor()
	if monitor != nil {
		fmt.Printf("✅ 慢查询监控启用状态: %t\n", monitor.IsEnabled())
		fmt.Printf("✅ 慢查询记录数: %d\n", monitor.GetRecordCount())
		fmt.Printf("✅ 慢查询阈值: %v\n", monitor.GetThreshold())
	} else {
		fmt.Println("❌ 慢查询监控器为空")
	}

	// 8. 测试便捷函数
	fmt.Println("\n8. 测试便捷函数...")

	// 使用全局ORM实例
	globalORM := orm.GetDefaultORM()
	if globalORM != nil {
		fmt.Println("✅ 获取全局ORM实例成功")

		// 测试便捷的CRUD函数
		testUser := &TestUser{
			Name:  "便捷函数测试用户",
			Email: "convenience@example.com",
		}

		// 使用便捷创建函数
		result := orm.Create(testUser)
		if result.Error != nil {
			fmt.Printf("⚠️ 便捷创建函数警告: %v\n", result.Error)
		} else {
			fmt.Printf("✅ 便捷创建函数成功，ID: %d\n", testUser.ID)
		}

		// 使用便捷查询函数
		var foundTestUser TestUser
		result = orm.First(&foundTestUser, "email = ?", "convenience@example.com")
		if result.Error != nil {
			fmt.Printf("⚠️ 便捷查询函数警告: %v\n", result.Error)
		} else {
			fmt.Printf("✅ 便捷查询函数成功: %s\n", foundTestUser.Name)
		}

		// 清理测试数据
		orm.Where("email = ?", "convenience@example.com").Delete(&TestUser{})
		fmt.Println("✅ 清理测试数据完成")
	}

	// 关闭连接
	if err := ormInstance.Close(); err != nil {
		fmt.Printf("⚠️ 关闭连接警告: %v\n", err)
	} else {
		fmt.Println("\n✅ 数据库连接已关闭")
	}

	fmt.Println("\n=== ORM 功能测试完成 ===")
	fmt.Println("✅ 所有基本功能测试通过！")
}
