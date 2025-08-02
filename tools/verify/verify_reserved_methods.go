package main

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

func main() {
	fmt.Println("=== ReservedMethods 覆盖验证 ===")
	
	// 创建BaseController实例
	controller := core.NewBaseController()
	controllerType := reflect.TypeOf(controller)
	
	var allMethods []string
	var missingMethods []string
	var reservedCount int
	
	// 收集所有公共方法
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		methodName := method.Name
		
		// 只统计公共方法（首字母大写）
		if len(methodName) > 0 && methodName[0] >= 'A' && methodName[0] <= 'Z' {
			allMethods = append(allMethods, methodName)
			
			// 检查是否在ReservedMethods中
			if core.ReservedMethods == nil {
				fmt.Println("错误：ReservedMethods 未初始化")
				return
			}
			
			if core.ReservedMethods[methodName] {
				reservedCount++
			} else {
				missingMethods = append(missingMethods, methodName)
			}
		}
	}
	
	sort.Strings(allMethods)
	sort.Strings(missingMethods)
	
	fmt.Printf("BaseController公共方法总数: %d\n", len(allMethods))
	fmt.Printf("ReservedMethods覆盖数量: %d\n", reservedCount)
	fmt.Printf("覆盖率: %.1f%%\n", float64(reservedCount)/float64(len(allMethods))*100)
	
	if len(missingMethods) > 0 {
		fmt.Printf("\n❌ 仍有 %d 个方法未被包含:\n", len(missingMethods))
		for _, method := range missingMethods {
			fmt.Printf("- %s\n", method)
		}
	} else {
		fmt.Println("\n✅ 所有BaseController公共方法都已包含在ReservedMethods中！")
	}
	
	// 检查ReservedMethods是否有多余的项目
	fmt.Println("\n=== 检查ReservedMethods完整性 ===")
	methodSet := make(map[string]bool)
	for _, method := range allMethods {
		methodSet[method] = true
	}
	
	var extraReserved []string
	for method := range core.ReservedMethods {
		if !methodSet[method] {
			extraReserved = append(extraReserved, method)
		}
	}
	
	sort.Strings(extraReserved)
	if len(extraReserved) > 0 {
		fmt.Printf("⚠️  ReservedMethods中有 %d 个不存在的方法（可能是私有方法或已删除的方法）:\n", len(extraReserved))
		for _, method := range extraReserved {
			fmt.Printf("- %s\n", method)
		}
	} else {
		fmt.Println("✅ ReservedMethods中没有多余的方法项目")
	}
}