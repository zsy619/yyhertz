package main

import (
	"fmt"
	"testing"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// BenchmarkTemplateLoading 模板加载性能测试
func BenchmarkTemplateLoading(b *testing.B) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}
	cfg.Cache.EnableCache = true

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.LoadTemplate("Login/Login")
		if err != nil {
			b.Errorf("加载模板失败: %v", err)
		}
	}
}

// BenchmarkTemplateLoadingWithoutCache 无缓存模板加载性能测试
func BenchmarkTemplateLoadingWithoutCache(b *testing.B) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}
	cfg.Cache.EnableCache = false

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.LoadTemplate("Login/Login")
		if err != nil {
			b.Errorf("加载模板失败: %v", err)
		}
	}
}

// BenchmarkSimpleTemplateRendering 简单模板渲染性能测试
func BenchmarkSimpleTemplateRendering(b *testing.B) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}
	cfg.Cache.EnableCache = true

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	testData := map[string]interface{}{
		"username": "testuser",
		"message":  "Hello World",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.RenderWithLayout("Login/Login", "", testData)
			if err != nil {
				b.Errorf("渲染模板失败: %v", err)
			}
		}
	})
}

// BenchmarkComplexTemplateRendering 复杂模板渲染性能测试
func BenchmarkComplexTemplateRendering(b *testing.B) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}
	cfg.Cache.EnableCache = true

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	complexData := &view.RenderData{
		Data: map[string]interface{}{
			"username": "complexuser",
			"message":  "Welcome to YYHertz!",
			"items": []string{
				"Item 1", "Item 2", "Item 3", "Item 4", "Item 5",
			},
			"userInfo": map[string]interface{}{
				"id":    12345,
				"email": "user@example.com",
				"roles": []string{"user", "admin"},
			},
		},
		CSRF: "complex-csrf-token-12345",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.RenderWithLayout("Login/LoginWithCSRF", "", complexData)
			if err != nil {
				b.Errorf("渲染复杂模板失败: %v", err)
			}
		}
	})
}

// BenchmarkLayoutInheritance 布局继承性能测试
func BenchmarkLayoutInheritance(b *testing.B) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}
	cfg.Cache.EnableCache = true

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	testData := map[string]interface{}{
		"Title":   "Layout Test",
		"Heading": "Performance Test",
		"Content": "This is a layout inheritance performance test",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.RenderWithLayout("content", "layout", testData)
			if err != nil {
				// 如果布局文件不存在，使用简单渲染
				_, err = engine.RenderWithLayout("test", "", testData)
				if err != nil {
					b.Errorf("布局继承渲染失败: %v", err)
				}
			}
		}
	})
}

// BenchmarkTemplateFunctions 模板函数执行性能测试
func BenchmarkTemplateFunctions(b *testing.B) {
	b.Run("MakeSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.MakeSlice("a", "b", "c", "d", "e")
		}
	})

	b.Run("ConcatString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.ConcatString("Hello", " ", "World", "!")
		}
	})

	b.Run("ContainString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.ContainString("a,b,c,d,e", "c")
		}
	})

	b.Run("FmtByte", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.FmtByte(1048576) // 1MB
		}
	})

	b.Run("AuthContain", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.AuthContain("1,2,3,4,5", 3)
		}
	})
}

// BenchmarkComparisonFunctions 比较函数性能测试
func BenchmarkComparisonFunctions(b *testing.B) {
	b.Run("Eq", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = view.Eq(5, 5)
		}
	})

	b.Run("Ne", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = view.Ne(5, 3)
		}
	})

	b.Run("Lt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = view.Lt(3, 5)
		}
	})

	b.Run("Gt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = view.Gt(5, 3)
		}
	})
}

// BenchmarkLogicalFunctions 逻辑函数性能测试
func BenchmarkLogicalFunctions(b *testing.B) {
	b.Run("And", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.And(true, true)
		}
	})

	b.Run("Or", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.Or(false, true)
		}
	})

	b.Run("Not", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.Not(false)
		}
	})
}

// BenchmarkCollectionFunctions 集合函数性能测试
func BenchmarkCollectionFunctions(b *testing.B) {
	slice := []any{"a", "b", "c", "d", "e"}

	b.Run("Len", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.Len(slice)
		}
	})

	b.Run("Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.Index(slice, 2)
		}
	})

	b.Run("Slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = view.Slice(slice, 1, 4)
		}
	})
}

// BenchmarkCSRFTokenAccess CSRF Token访问性能测试
func BenchmarkCSRFTokenAccess(b *testing.B) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	testData := &view.RenderData{
		Data: map[string]interface{}{
			"username": "testuser",
		},
		CSRF: "benchmark-csrf-token",
	}

	b.Run("DirectFieldAccess", func(b *testing.B) {
		template := `{{.CSRF}}`
		tmpl, err := engine.CreateInlineTemplate("csrf_direct", template)
		if err != nil {
			b.Fatalf("创建内联模板失败: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.ExecuteTemplate(tmpl, testData)
			if err != nil {
				b.Errorf("执行模板失败: %v", err)
			}
		}
	})

	b.Run("PreparedFieldAccess", func(b *testing.B) {
		template := `{{.CsrfToken}}`
		tmpl, err := engine.CreateInlineTemplate("csrf_prepared", template)
		if err != nil {
			b.Fatalf("创建内联模板失败: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			preparedData := engine.PrepareRenderData(testData)
			_, err := engine.ExecuteTemplate(tmpl, preparedData)
			if err != nil {
				b.Errorf("执行模板失败: %v", err)
			}
		}
	})

	b.Run("FunctionAccess", func(b *testing.B) {
		template := `{{csrf}}`
		tmpl, err := engine.CreateInlineTemplate("csrf_func", template)
		if err != nil {
			b.Fatalf("创建内联模板失败: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.ExecuteTemplate(tmpl, testData)
			if err != nil {
				b.Errorf("执行模板失败: %v", err)
			}
		}
	})
}

// BenchmarkFunctionManager 函数管理器性能测试
func BenchmarkFunctionManager(b *testing.B) {
	manager := view.GetGlobalFunctionManager()

	b.Run("HasFunction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = manager.HasFunction("add")
		}
	})

	b.Run("GetFunctionSource", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = manager.GetFunctionSource("add")
		}
	})

	b.Run("AddGlobalFunction", func(b *testing.B) {
		testFunc := func() string { return "test" }
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			funcName := fmt.Sprintf("benchmarkFunc%d", i%1000)
			manager.AddGlobalFunction(funcName, testFunc)
		}
	})
}

// BenchmarkTemplateCache 模板缓存性能测试
func BenchmarkTemplateCache(b *testing.B) {
	b.Run("WithCache", func(b *testing.B) {
		cfg := config.GlobalTemplate
		cfg.Paths.ViewPaths = []string{"views"}
		cfg.Cache.EnableCache = true

		engine, err := view.NewTemplateEngine(cfg)
		if err != nil {
			b.Fatalf("创建缓存引擎失败: %v", err)
		}

		testData := map[string]interface{}{
			"username": "cacheuser",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.RenderWithLayout("Login/Login", "", testData)
			if err != nil {
				b.Errorf("缓存渲染失败: %v", err)
			}
		}
	})

	b.Run("WithoutCache", func(b *testing.B) {
		cfg := config.GlobalTemplate
		cfg.Paths.ViewPaths = []string{"views"}
		cfg.Cache.EnableCache = false

		engine, err := view.NewTemplateEngine(cfg)
		if err != nil {
			b.Fatalf("创建非缓存引擎失败: %v", err)
		}

		testData := map[string]interface{}{
			"username": "nocacheuser",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.RenderWithLayout("Login/Login", "", testData)
			if err != nil {
				b.Errorf("非缓存渲染失败: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentRendering 并发渲染性能测试
func BenchmarkConcurrentRendering(b *testing.B) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}
	cfg.Cache.EnableCache = true

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	testData := map[string]interface{}{
		"username": "concurrentuser",
		"message":  "Concurrent rendering test",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := engine.RenderWithLayout("Login/Login", "", testData)
			if err != nil {
				b.Errorf("并发渲染失败: %v", err)
			}
		}
	})
}

// BenchmarkInlineTemplateCreation 内联模板创建性能测试
func BenchmarkInlineTemplateCreation(b *testing.B) {
	cfg := config.GlobalTemplate
	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	templateContent := `<h1>{{.Title}}</h1><p>{{.Message}}</p>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		templateName := "inline_" + string(rune(i%1000))
		_, err := engine.CreateInlineTemplate(templateName, templateContent)
		if err != nil {
			b.Errorf("创建内联模板失败: %v", err)
		}
	}
}

// BenchmarkMemoryAllocation 内存分配分析
func BenchmarkMemoryAllocation(b *testing.B) {
	cfg := config.GlobalTemplate
	cfg.Paths.ViewPaths = []string{"views"}
	cfg.Cache.EnableCache = true

	engine, err := view.NewTemplateEngine(cfg)
	if err != nil {
		b.Fatalf("创建模板引擎失败: %v", err)
	}

	testData := map[string]interface{}{
		"username": "memoryuser",
		"items":    []string{"item1", "item2", "item3"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := engine.RenderWithLayout("Login/Login", "", testData)
		if err != nil {
			b.Errorf("内存分配测试失败: %v", err)
		}
	}
}
