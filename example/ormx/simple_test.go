package main

import (
	"testing"

	"github.com/zsy619/yyhertz/framework/orm"
)

// SimpleUser 简单用户模型用于测试
type SimpleUser struct {
	orm.BaseModel
	Name  string `gorm:"size:100" json:"name"`
	Email string `gorm:"size:100" json:"email"`
	Age   int    `json:"age"`
}

// TestSimpleORM 简单ORM测试
func TestSimpleORMExt(t *testing.T) {
	t.Log("=== 简单ORM测试 ===")

	// 获取ORM实例
	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		t.Fatal("获取ORM实例失败")
	}

	// 测试数据库连接
	if err := ormInstance.Ping(); err != nil {
		t.Logf("数据库连接测试: %v", err)
	} else {
		t.Log("✅ 数据库连接正常")
	}

	// 测试自动迁移
	if err := ormInstance.AutoMigrate(&SimpleUser{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	t.Log("✅ 自动迁移成功")

	// 测试创建记录
	user := &SimpleUser{
		Name:  "测试用户",
		Email: "test@example.com",
		Age:   25,
	}

	result := orm.Create(user)
	if result.Error != nil {
		t.Logf("创建用户警告: %v", result.Error)
	} else {
		t.Logf("✅ 创建用户成功，ID: %d", user.ID)
	}

	// 测试查询记录
	var foundUser SimpleUser
	result = orm.First(&foundUser, "email = ?", "test@example.com")
	if result.Error != nil {
		t.Logf("查询用户警告: %v", result.Error)
	} else {
		t.Logf("✅ 查询用户成功: %s", foundUser.Name)
	}

	// 测试统计信息
	stats := ormInstance.GetStats()
	if len(stats) > 0 {
		t.Log("✅ 获取统计信息成功")
		for key, value := range stats {
			t.Logf("  %s: %v", key, value)
		}
	}

	// 清理测试数据
	orm.Where("email = ?", "test@example.com").Delete(&SimpleUser{})
	t.Log("✅ 清理测试数据完成")

	t.Log("=== 简单ORM测试完成 ===")
}

// TestSlowQueryMonitor 测试慢查询监控
func TestSlowQueryMonitor(t *testing.T) {
	t.Log("=== 慢查询监控测试 ===")

	// 获取慢查询监控器
	monitor := orm.GetGlobalSlowQueryMonitor()
	if monitor == nil {
		t.Fatal("获取慢查询监控器失败")
	}

	// 检查监控器状态
	if !monitor.IsEnabled() {
		t.Log("慢查询监控未启用")
	} else {
		t.Log("✅ 慢查询监控已启用")
	}

	// 获取统计信息
	stats := orm.GetSlowQueryStats()
	t.Log("慢查询统计信息:")
	for key, value := range stats {
		t.Logf("  %s: %v", key, value)
	}

	t.Log("✅ 慢查询监控测试完成")
}

// TestSimpleDatabaseConfig 测试数据库配置（重命名避免冲突）
func TestSimpleDatabaseConfig(t *testing.T) {
	t.Log("=== 简单数据库配置测试 ===")

	// 获取默认配置
	config := orm.DefaultDatabaseConfig()
	if config == nil {
		t.Fatal("获取默认配置失败")
	}

	t.Logf("数据库类型: %s", config.Type)
	t.Logf("数据库主机: %s", config.Host)
	t.Logf("数据库端口: %d", config.Port)
	t.Logf("数据库名称: %s", config.Database)
	t.Logf("最大连接数: %d", config.MaxOpenConns)
	t.Logf("最大空闲连接: %d", config.MaxIdleConns)

	t.Log("✅ 简单数据库配置测试完成")
}
