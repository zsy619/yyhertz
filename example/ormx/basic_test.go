package main

import (
	"fmt"
	"log"

	"github.com/zsy619/yyhertz/framework/orm"
)

// BasicUser 基础用户模型
type BasicUser struct {
	orm.BaseModel
	Name string `gorm:"size:100" json:"name"`
}

func TestBasicFunctionality() {
	fmt.Println("=== 基础功能测试 ===")

	// 1. 测试ORM实例获取
	fmt.Println("1. 获取ORM实例...")
	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		log.Fatal("❌ 获取ORM实例失败")
	}
	fmt.Println("✅ ORM实例获取成功")

	// 2. 测试数据库连接
	fmt.Println("2. 测试数据库连接...")
	if err := ormInstance.Ping(); err != nil {
		fmt.Printf("⚠️ 数据库连接测试: %v\n", err)
	} else {
		fmt.Println("✅ 数据库连接正常")
	}

	// 3. 测试自动迁移
	fmt.Println("3. 测试自动迁移...")
	if err := ormInstance.AutoMigrate(&BasicUser{}); err != nil {
		fmt.Printf("⚠️ 自动迁移警告: %v\n", err)
	} else {
		fmt.Println("✅ 自动迁移成功")
	}

	// 4. 测试基本CRUD操作
	fmt.Println("4. 测试基本CRUD操作...")

	// 创建
	user := &BasicUser{Name: "测试用户"}
	result := orm.Create(user)
	if result.Error != nil {
		fmt.Printf("⚠️ 创建用户警告: %v\n", result.Error)
	} else {
		fmt.Printf("✅ 创建用户成功，ID: %d\n", user.ID)
	}

	// 查询
	var foundUser BasicUser
	result = orm.First(&foundUser, "name = ?", "测试用户")
	if result.Error != nil {
		fmt.Printf("⚠️ 查询用户警告: %v\n", result.Error)
	} else {
		fmt.Printf("✅ 查询用户成功: %s\n", foundUser.Name)
	}

	// 更新
	result = orm.Where("name = ?", "测试用户").Updates(&BasicUser{Name: "更新用户"})
	if result.Error != nil {
		fmt.Printf("⚠️ 更新用户警告: %v\n", result.Error)
	} else {
		fmt.Println("✅ 更新用户成功")
	}

	// 删除
	result = orm.Where("name = ?", "更新用户").Delete(&BasicUser{})
	if result.Error != nil {
		fmt.Printf("⚠️ 删除用户警告: %v\n", result.Error)
	} else {
		fmt.Println("✅ 删除用户成功")
	}

	// 5. 测试统计信息
	fmt.Println("5. 测试统计信息...")
	stats := ormInstance.GetStats()
	if len(stats) > 0 {
		fmt.Println("✅ 获取统计信息成功:")
		for key, value := range stats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	} else {
		fmt.Println("⚠️ 统计信息为空")
	}

	// 6. 测试慢查询监控
	fmt.Println("6. 测试慢查询监控...")
	monitor := orm.GetGlobalSlowQueryMonitor()
	if monitor != nil {
		fmt.Printf("✅ 慢查询监控状态: %t\n", monitor.IsEnabled())
		fmt.Printf("✅ 慢查询记录数: %d\n", monitor.GetRecordCount())
	} else {
		fmt.Println("⚠️ 慢查询监控器为空")
	}

	fmt.Println("=== 基础功能测试完成 ===")
}

// 如果需要单独运行，可以取消注释下面的 main 函数
// func main() {
// 	TestBasicFunctionality()
// }
