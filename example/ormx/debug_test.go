package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/zsy619/yyhertz/framework/orm"
)

func TestDebug(t *testing.T) {
	fmt.Println("开始调试测试...")

	// 检查当前工作目录
	pwd, _ := os.Getwd()
	fmt.Printf("当前工作目录: %s\n", pwd)

	// 尝试获取默认配置
	fmt.Println("获取默认数据库配置...")
	config := orm.DefaultDatabaseConfig()
	if config != nil {
		fmt.Printf("数据库类型: %s\n", config.Type)
		fmt.Printf("数据库文件: %s\n", config.Database)
		fmt.Printf("主机: %s\n", config.Host)
		fmt.Printf("端口: %d\n", config.Port)
	} else {
		fmt.Println("❌ 获取配置失败")
		return
	}

	fmt.Println("尝试创建ORM实例...")
	ormInstance, err := orm.NewORM(config)
	if err != nil {
		fmt.Printf("❌ 创建ORM实例失败: %v\n", err)
		return
	}

	fmt.Println("✅ ORM实例创建成功")

	// 测试数据库连接
	fmt.Println("测试数据库连接...")
	if err := ormInstance.Ping(); err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
	} else {
		fmt.Println("✅ 数据库连接成功")
	}

	// 关闭连接
	if err := ormInstance.Close(); err != nil {
		fmt.Printf("⚠️ 关闭连接警告: %v\n", err)
	} else {
		fmt.Println("✅ 连接已关闭")
	}

	fmt.Println("调试测试完成")
}
