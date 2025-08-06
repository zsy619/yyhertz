package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/zsy619/yyhertz/framework/orm"
)

func TestQuick(t *testing.T) {
	fmt.Println("=== 快速ORM测试 ===")

	// 测试配置
	fmt.Println("1. 测试配置...")
	config := orm.DefaultDatabaseConfig()
	fmt.Printf("数据库类型: %s\n", config.Type)

	// 测试ORM创建
	fmt.Println("2. 测试ORM创建...")
	ormInstance, err := orm.NewORM(config)
	if err != nil {
		log.Printf("创建ORM失败: %v", err)
		return
	}
	fmt.Println("ORM创建成功")

	// 测试连接
	fmt.Println("3. 测试连接...")
	if err := ormInstance.Ping(); err != nil {
		log.Printf("连接失败: %v", err)
		return
	}
	fmt.Println("连接成功")

	// 关闭连接
	if err := ormInstance.Close(); err != nil {
		log.Printf("关闭连接失败: %v", err)
	} else {
		fmt.Println("连接已关闭")
	}

	fmt.Println("=== 测试完成 ===")
}
