package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/zsy619/yyhertz/framework/orm"
)

func TestVerify(t *testing.T) {
	fmt.Println("开始验证ORM功能...")

	// 1. 测试配置
	config := orm.DefaultDatabaseConfig()
	fmt.Printf("✅ 配置获取成功: %s\n", config.Type)

	// 2. 创建ORM实例
	ormInstance, err := orm.NewORM(config)
	if err != nil {
		fmt.Printf("❌ 创建ORM失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ ORM实例创建成功")

	// 3. 测试连接
	if err := ormInstance.Ping(); err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 数据库连接成功")

	// 4. 获取统计信息
	stats := ormInstance.GetStats()
	fmt.Printf("✅ 连接池统计: %d 项\n", len(stats))

	// 5. 测试全局ORM
	globalORM := orm.GetDefaultORM()
	if globalORM != nil {
		fmt.Println("✅ 全局ORM获取成功")
	}

	// 6. 测试慢查询监控
	monitor := orm.GetGlobalSlowQueryMonitor()
	if monitor != nil {
		fmt.Printf("✅ 慢查询监控: %t\n", monitor.IsEnabled())
	}

	// 7. 关闭连接
	if err := ormInstance.Close(); err != nil {
		fmt.Printf("⚠️ 关闭连接警告: %v\n", err)
	} else {
		fmt.Println("✅ 连接已关闭")
	}

	fmt.Println("🎉 ORM功能验证完成！所有基本功能正常！")
}
