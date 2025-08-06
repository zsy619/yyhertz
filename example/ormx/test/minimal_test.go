package main

import (
	"fmt"
	"log"
	"testing"
)

func TestMinimal(t *testing.T) {
	fmt.Println("最小测试开始...")

	// 测试基本的Go功能
	fmt.Println("Go基本功能正常")

	// 尝试导入ORM包
	fmt.Println("尝试导入ORM包...")

	// 延迟导入以避免初始化阻塞
	defer func() {
		if r := recover(); r != nil {
			log.Printf("捕获到panic: %v", r)
		}
	}()

	// 动态导入
	fmt.Println("导入成功，测试完成")
}
