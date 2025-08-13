package main

import (
	"fmt"

	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// DemoController 演示控制器
type DemoController struct {
	*core.BaseController
}

func (c *DemoController) GetApi() {
	c.Ctx.JSON(200, map[string]any{
		"message": "API endpoint",
		"path":    string(c.Ctx.Request().Path()),
	})
}

func (c *DemoController) GetUser() {
	c.Ctx.JSON(200, map[string]any{
		"message": "User endpoint",
		"path":    string(c.Ctx.Request().Path()),
	})
}

func (c *DemoController) GetAdmin() {
	c.Ctx.JSON(200, map[string]any{
		"message": "Admin endpoint",
		"path":    string(c.Ctx.Request().Path()),
	})
}

func main() {
	fmt.Println("🎯 Pattern Matching 自动匹配演示")
	fmt.Println("================================")

	// 演示：过滤器函数不需要做路径判断，框架自动匹配

	// 1. API路径过滤器 - 只对 /demo/api* 路径有效
	apiFilter := func(ctx *context.Context) {
		// 注意：这里不需要检查路径，框架已经自动匹配了！
		fmt.Printf("🔹 API过滤器执行 - 路径: %s\n", string(ctx.Request().Path()))
		ctx.Set("api_filtered", true)
	}

	// 2. 用户路径过滤器 - 只对 /demo/user* 路径有效
	userFilter := func(ctx *context.Context) {
		// 不需要判断路径，框架保证只在匹配时执行
		fmt.Printf("👤 用户过滤器执行 - 路径: %s\n", string(ctx.Request().Path()))
		ctx.Set("user_filtered", true)
	}

	// 3. 管理员路径过滤器 - 只对 /demo/admin* 路径有效
	adminFilter := func(ctx *context.Context) {
		// 框架已经做了pattern匹配，这里直接执行业务逻辑
		fmt.Printf("👑 管理员过滤器执行 - 路径: %s\n", string(ctx.Request().Path()))
		ctx.Set("admin_filtered", true)
	}

	// 4. 全局过滤器 - 对所有路径有效
	globalFilter := func(ctx *context.Context) {
		fmt.Printf("🌐 全局过滤器执行 - 路径: %s\n", string(ctx.Request().Path()))
		ctx.Set("global_filtered", true)
	}

	// 5. JSON响应过滤器 - 只对.json结尾的路径有效
	jsonFilter := func(ctx *context.Context) {
		fmt.Printf("📄 JSON过滤器执行 - 路径: %s\n", string(ctx.Request().Path()))
		ctx.SetHeader("Content-Type", "application/json")
	}

	// 插入过滤器 - 指定精确的pattern
	fmt.Println("\n📋 注册过滤器:")

	mvc.InsertFilter("/demo/api*", constant.BeforeRouter, apiFilter)
	fmt.Println("   ✓ API过滤器: /demo/api*")

	mvc.InsertFilter("/demo/user*", constant.BeforeRouter, userFilter)
	fmt.Println("   ✓ 用户过滤器: /demo/user*")

	mvc.InsertFilter("/demo/admin*", constant.BeforeRouter, adminFilter)
	fmt.Println("   ✓ 管理员过滤器: /demo/admin*")

	mvc.InsertFilter("*", constant.BeforeExec, globalFilter)
	fmt.Println("   ✓ 全局过滤器: *")

	mvc.InsertFilter("*.json", constant.AfterExec, jsonFilter)
	fmt.Println("   ✓ JSON过滤器: *.json")

	// 演示中间通配符
	middlewareFilter := func(ctx *context.Context) {
		fmt.Printf("🔗 中间件过滤器执行 - 路径: %s\n", string(ctx.Request().Path()))
	}
	mvc.InsertFilter("/demo/*/info", constant.BeforeExec, middlewareFilter)
	fmt.Println("   ✓ 中间件过滤器: /demo/*/info")

	// 注册控制器
	app := mvc.HertzApp
	app.AutoRouters(&DemoController{})

	fmt.Println("\n🚀 服务器启动在端口 8080")
	fmt.Println("\n📝 测试URL (观察过滤器自动匹配):")
	fmt.Println("   GET http://localhost:8080/demo/api     - 只触发: 全局+API过滤器")
	fmt.Println("   GET http://localhost:8080/demo/user    - 只触发: 全局+用户过滤器")
	fmt.Println("   GET http://localhost:8080/demo/admin   - 只触发: 全局+管理员过滤器")
	fmt.Println("   GET http://localhost:8080/demo/api.json - 触发: 全局+API+JSON过滤器")
	fmt.Println("   GET http://localhost:8080/demo/test/info - 触发: 全局+中间件过滤器")
	fmt.Println("   GET http://localhost:8080/other        - 只触发: 全局过滤器")

	// 显示过滤器统计
	showPatternStats()

	app.Run(":8080")
}

// showPatternStats 显示pattern匹配统计
func showPatternStats() {
	fmt.Println("\n📊 过滤器Pattern统计:")

	allFilters := mvc.GetAllFilters()
	for position, filters := range allFilters {
		if len(filters) > 0 {
			positionName := getPositionName(position)
			fmt.Printf("   %s:\n", positionName)

			for i, filter := range filters {
				fmt.Printf("     %d. Pattern: '%s' (优先级: %d)\n",
					i+1, filter.Pattern, filter.Priority)
			}
		}
	}
}

func getPositionName(position int) string {
	names := map[int]string{
		mvc.BeforeStatic:      "BeforeStatic",
		constant.BeforeRouter: "BeforeRouter",
		constant.BeforeExec:   "BeforeExec",
		constant.AfterExec:    "AfterExec",
		mvc.FinishRouter:      "FinishRouter",
	}
	if name, exists := names[position]; exists {
		return name
	}
	return fmt.Sprintf("Position_%d", position)
}

/*
运行此示例后，您会看到：

1. 访问 /demo/api 时：
   🌐 全局过滤器执行 - 路径: /demo/api
   🔹 API过滤器执行 - 路径: /demo/api
   (用户和管理员过滤器不会执行，因为pattern不匹配)

2. 访问 /demo/user 时：
   🌐 全局过滤器执行 - 路径: /demo/user
   👤 用户过滤器执行 - 路径: /demo/user
   (API和管理员过滤器不会执行)

3. 访问 /other 时：
   🌐 全局过滤器执行 - 路径: /other
   (其他所有过滤器都不会执行)

这证明了框架自动进行pattern匹配，过滤器函数内部不需要做路径判断！
*/
