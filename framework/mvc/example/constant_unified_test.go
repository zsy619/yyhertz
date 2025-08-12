package example

import (
	"fmt"
	"log"
	"testing"

	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
)

// 演示常量统一后的使用方式
func TestConstantUnifiedDemo(t *testing.T) {
	fmt.Println("=== YYHertz 框架常量统一系统演示 ===\n")

	// 1. 展示统一的执行层级常量
	demonstrateExecutionLayers()

	// 2. 展示过滤器位置兼容性
	demonstrateFilterPositions()

	// 3. 展示系统集成使用
	demonstrateSystemIntegration()

	fmt.Println("\n=== 演示完成 ===")
}

// demonstrateExecutionLayers 演示统一执行层级常量
func demonstrateExecutionLayers() {
	fmt.Println("1. 统一执行层级常量:")

	layers := constant.GetAllExecutionLayers()
	for _, layer := range layers {
		name := constant.GetExecutionLayerName(layer)
		fmt.Printf("   - %s (Layer %d)\n", name, int(layer))
	}
	fmt.Println()
}

// demonstrateFilterPositions 演示过滤器位置兼容性
func demonstrateFilterPositions() {
	fmt.Println("2. 过滤器位置常量 (向后兼容):")

	positions := []int{
		constant.BeforeStatic,
		constant.BeforeRouter,
		constant.BeforeExec,
		constant.AfterExec,
		constant.FinishRouter,
	}

	for _, position := range positions {
		name := constant.GetFilterPositionName(position)
		layer := constant.FilterPositionToLayer(position)
		layerName := constant.GetExecutionLayerName(layer)
		fmt.Printf("   - %s (Position %d) -> %s (Layer %d)\n",
			name, position, layerName, int(layer))
	}
	fmt.Println()
}

// demonstrateSystemIntegration 演示系统集成
func demonstrateSystemIntegration() {
	fmt.Println("3. 系统集成演示:")

	// 创建应用实例
	app := mvc.NewApp()

	// 使用新的统一常量注册过滤器
	fmt.Println("   注册过滤器:")

	// 使用过滤器位置常量 (向后兼容)
	app.InsertFilter("/*", mvc.BeforeStatic, func(ctx *context.Context) {
		fmt.Printf("     -> BeforeStatic 过滤器执行 (位置: %d)\n", mvc.BeforeStatic)
	})

	app.InsertFilter("/*", mvc.BeforeRouter, func(ctx *context.Context) {
		fmt.Printf("     -> BeforeRouter 过滤器执行 (位置: %d)\n", mvc.BeforeRouter)
	})

	app.InsertFilter("/*", mvc.BeforeExec, func(ctx *context.Context) {
		fmt.Printf("     -> BeforeExec 过滤器执行 (位置: %d)\n", mvc.BeforeExec)
	})

	// 使用中间件管道的统一层级常量
	fmt.Println("   注册中间件:")

	pipeline := middleware.NewMiddlewarePipeline()

	// 使用新的统一常量
	pipeline.Use(middleware.LayerGlobal, "global-auth", func(ctx *context.Context) {
		fmt.Printf("     -> Global 中间件执行 (层级: %d)\n", int(middleware.LayerGlobal))
	}, 10)

	pipeline.Use(middleware.LayerRoute, "route-validation", func(ctx *context.Context) {
		fmt.Printf("     -> Route 中间件执行 (层级: %d)\n", int(middleware.LayerRoute))
	}, 20)

	pipeline.Use(middleware.LayerController, "controller-logging", func(ctx *context.Context) {
		fmt.Printf("     -> Controller 中间件执行 (层级: %d)\n", int(middleware.LayerController))
	}, 30)

	// 展示转换功能
	fmt.Println("   常量转换:")
	filterPosition := mvc.BeforeRouter
	middlewareLayer := constant.FilterPositionToLayer(filterPosition)
	fmt.Printf("     过滤器位置 %d (%s) 对应中间件层级 %d (%s)\n",
		filterPosition,
		constant.GetFilterPositionName(filterPosition),
		int(middlewareLayer),
		constant.GetExecutionLayerName(middlewareLayer))

	// 验证双向转换
	backToPosition := constant.LayerToFilterPosition(middlewareLayer)
	fmt.Printf("     反向转换: 层级 %d -> 位置 %d ✓\n", int(middlewareLayer), backToPosition)

	// 显示所有过滤器
	fmt.Println("   已注册的过滤器:")
	allFilters := app.GetAllFilters()
	for position, filters := range allFilters {
		positionName := constant.GetFilterPositionName(position)
		fmt.Printf("     位置 %s (%d): %d 个过滤器\n", positionName, position, len(filters))
	}
}

// 示例控制器，展示常量在实际应用中的使用
type DemoController struct {
	mvc.BaseController
}

// GetIndex 示例方法
func (c *DemoController) GetIndex() {
	// 可以在控制器中使用统一常量
	fmt.Printf("控制器执行中，当前可用的执行层级：\n")
	for _, layer := range constant.GetAllExecutionLayers() {
		name := constant.GetExecutionLayerName(layer)
		fmt.Printf("  - %s\n", name)
	}

	c.Ctx.JSON(200, map[string]interface{}{
		"message": "演示控制器",
		"layers":  constant.GetAllExecutionLayers(),
	})
}

func init() {
	// 在包初始化时可以进行验证
	fmt.Println("包初始化：验证常量一致性...")

	// 验证所有过滤器位置都能正确映射到中间件层级
	positions := []int{
		constant.BeforeStatic,
		constant.BeforeRouter,
		constant.BeforeExec,
		constant.AfterExec,
		constant.FinishRouter,
	}

	for _, pos := range positions {
		if !constant.IsValidFilterPosition(pos) {
			log.Fatalf("无效的过滤器位置: %d", pos)
		}

		layer := constant.FilterPositionToLayer(pos)
		if !constant.IsValidExecutionLayer(layer) {
			log.Fatalf("转换结果无效的层级: %d", int(layer))
		}

		// 验证双向转换
		backPos := constant.LayerToFilterPosition(layer)
		if backPos != pos {
			log.Fatalf("转换不一致: %d -> %d -> %d", pos, int(layer), backPos)
		}
	}

	fmt.Println("常量验证通过 ✓")
}
