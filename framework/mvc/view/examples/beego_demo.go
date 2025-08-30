package main

import (
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

func beegoDemo() {
	log.Println("🚀 YYHertz Beego模板引擎演示程序启动...")

	// 演示1：基础统一API使用
	demonstrateBasicUsage()

	// 演示2：模板缓存优化效果
	demonstrateCacheOptimization()

	// 演示3：Beego风格函数
	demonstrateBeegoFunctions()

	log.Println("✅ YYHertz Beego模板引擎演示完成!")
}

func demonstrateBasicUsage() {
	log.Println("📋 演示1: 基础统一API使用")

	// 获取统一API
	api := view.GetUnifiedAPI()

	// 测试数据
	data := map[string]interface{}{
		"UserName":    "张三",
		"CurrentTime": time.Now().Format("2006-01-02 15:04:05"),
		"Message":     "欢迎使用YYHertz Beego模板引擎!",
	}

	// 使用统一API渲染
	result, err := api.Render("test-template", data)
	if err != nil {
		log.Printf("❌ 统一API渲染失败: %v", err)
	} else {
		log.Printf("✅ 统一API渲染成功: %s", result)
	}

	// 获取引擎统计
	stats := view.GetUnifiedStats()
	log.Printf("📊 引擎统计: %+v", stats)
}

func demonstrateCacheOptimization() {
	log.Println("📋 演示2: 模板缓存优化效果")

	// 创建测试数据
	data := map[string]interface{}{
		"Title":   "缓存测试页面",
		"Content": "这是一个用于测试Beego模板缓存功能的页面",
		"Items": []string{
			"缓存命中优化",
			"预编译加速",
			"内存管理",
			"热重载支持",
		},
	}

	// 第一次渲染（预期缓存未命中）
	start1 := time.Now()
	result1, err1 := view.UnifiedRender("cache-test", data)
	duration1 := time.Since(start1)

	if err1 != nil {
		log.Printf("❌ 第一次渲染失败: %v", err1)
	} else {
		log.Printf("✅ 第一次渲染成功，耗时: %v", duration1)
	}

	// 第二次渲染（预期缓存命中）
	start2 := time.Now()
	result2, err2 := view.UnifiedRender("cache-test", data)
	duration2 := time.Since(start2)

	if err2 != nil {
		log.Printf("❌ 第二次渲染失败: %v", err2)
	} else {
		log.Printf("✅ 第二次渲染成功，耗时: %v", duration2)
		if duration2 < duration1 {
			log.Printf("🎯 缓存优化效果明显！加速比: %.2fx", float64(duration1)/float64(duration2))
		}
	}

	// 输出渲染结果对比
	if result1 == result2 {
		log.Println("✅ 缓存一致性验证通过")
	} else {
		log.Println("❌ 缓存一致性验证失败")
	}
}

func demonstrateBeegoFunctions() {
	log.Println("📋 演示3: Beego风格函数")

	// 启用Beego开发模式
	err := view.EnableBeegoDevMode()
	if err != nil {
		log.Printf("⚠️ 启用Beego开发模式失败: %v", err)
	} else {
		log.Println("✅ Beego开发模式已启用")
	}

	// 设置Beego视图路径
	err = view.SetBeegoViewPaths("./examples/views", "./examples/templates")
	if err != nil {
		log.Printf("⚠️ 设置Beego视图路径失败: %v", err)
	} else {
		log.Println("✅ Beego视图路径已设置")
	}

	// 启用Beego Gzip压缩
	view.EnableBeegoGzip(6)
	log.Println("✅ Beego Gzip压缩已启用")

	// 获取Beego性能统计
	perfStats := view.GetBeegoPerformanceStats()
	log.Printf("📈 Beego性能统计: %+v", perfStats)

	// 获取Beego引擎统计
	engineStats := view.GetBeegoEngineStats()
	log.Printf("⚙️ Beego引擎统计: %+v", engineStats)

	// 清除所有Beego缓存
	view.ClearAllBeegoCaches()
	log.Println("🧹 所有Beego缓存已清除")

	// 重建所有Beego模板
	err = view.RebuildAllBeegoTemplates()
	if err != nil {
		log.Printf("⚠️ 重建Beego模板失败: %v", err)
	} else {
		log.Println("🔄 所有Beego模板已重建")
	}

	// 注册Beego布局
	layoutContent := `<html><head><title>{{.Title}}</title></head><body><main>{{.Content}}</main></body></html>`
	err = view.RegisterBeegoLayout("demo-layout", layoutContent)
	if err != nil {
		log.Printf("⚠️ 注册Beego布局失败: %v", err)
	} else {
		log.Println("✅ Beego演示布局已注册")
	}

	// 使用布局渲染
	layoutData := map[string]interface{}{
		"Title":   "Beego布局演示",
		"Content": "这是通过Beego布局系统渲染的内容",
	}

	layoutResult, err := view.UnifiedRenderWithLayout("demo-content", "demo-layout", layoutData)
	if err != nil {
		log.Printf("❌ 布局渲染失败: %v", err)
	} else {
		log.Printf("✅ 布局渲染成功: %s", layoutResult)
	}
}

func demonstrateAdvancedFeatures() {
	log.Println("📋 演示4: 高级功能")

	// 演示字符串模板渲染
	templateContent := `
Hello {{.Name}}!
Current time: {{.Time}}
Features enabled: {{range .Features}}
- {{.}}{{end}}
`

	stringData := map[string]interface{}{
		"Name": "Beego用户",
		"Time": time.Now().Format("15:04:05"),
		"Features": []string{
			"模板缓存",
			"热重载", 
			"布局继承",
			"性能监控",
		},
	}

	stringResult, err := view.UnifiedRenderString(templateContent, stringData)
	if err != nil {
		log.Printf("❌ 字符串模板渲染失败: %v", err)
	} else {
		log.Printf("✅ 字符串模板渲染成功:\n%s", stringResult)
	}

	// 演示批量渲染
	templates := []string{"template1", "template2", "template3"}
	batchData := map[string]interface{}{
		"BatchID": "BATCH-001",
		"Count":   len(templates),
	}

	batchResults, err := view.GetUnifiedAPI().BatchRender(templates, batchData)
	if err != nil {
		log.Printf("❌ 批量渲染失败: %v", err)
	} else {
		log.Printf("✅ 批量渲染成功，完成 %d 个模板", len(batchResults))
		for template, result := range batchResults {
			log.Printf("  - %s: %s", template, result)
		}
	}
}

func showSummary() {
	log.Println("📋 功能总结")
	log.Println("✅ 已实现的Beego模板功能:")
	log.Println("  1. 🚀 Beego风格的模板引擎核心")
	log.Println("  2. 🏗️ 布局继承和块系统")
	log.Println("  3. 🔄 模板热重载和开发模式")
	log.Println("  4. ⚡ 高级模板功能和性能优化")
	log.Println("  5. 🎯 统一的模板引擎接口")
	log.Println("  6. 📊 完整的演示和测试案例")
	log.Println("")
	log.Println("🎉 YYHertz Framework现已完整支持Beego风格的模板系统!")
	log.Println("📖 详细文档和示例请参考: ./examples/views/beego/")
}