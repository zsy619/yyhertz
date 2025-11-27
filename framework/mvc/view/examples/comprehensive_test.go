package main

import (
	"slices"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// TestBeegoTemplateSystem 测试完整的Beego模板系统
func TestBeegoTemplateSystem(t *testing.T) {
	t.Log("=== 开始测试Beego模板系统完整功能 ===")

	// 1. 测试自动推导功能
	t.Run("AutoTemplateInference", func(t *testing.T) {
		inference := view.GetAutoTemplateInference()
		
		// 测试常见控制器映射
		candidates := inference.InferTemplatePath("UserController", "Login")
		t.Logf("UserController.Login 推导结果: %v", candidates)
		
		if len(candidates) == 0 {
			t.Error("Expected at least one template candidate")
		}
		
		// 检查是否包含预期的标准路径
		if !slices.Contains(candidates, "user/login") {
			t.Error("Expected to find 'user/login' in candidates")
		}
	})

	// 2. 测试命名约定
	t.Run("NamingConventions", func(t *testing.T) {
		manager := view.GetConventionManager()
		
		// 测试Beego标准约定
		result := manager.ApplyConvention("UserController.Profile", view.BeegoStandard)
		t.Logf("UserController.Profile -> %s", result)
		
		expected := "user/profile"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
		
		// 测试snake_case约定
		result = manager.ApplyConvention("UserController.Profile", view.SnakeCase)
		t.Logf("Snake case result: %s", result)
	})

	// 3. 测试统一API
	t.Run("UnifiedAPI", func(t *testing.T) {
		api := view.GetUnifiedAPI()
		
		// 测试基本渲染（会返回错误因为没有实际模板文件，但不应该panic）
		_, err := api.Render("test", map[string]any{"message": "hello"})
		t.Logf("Unified API render result: %v", err) // 预期会有错误，因为没有模板文件
		
		// 测试字符串渲染
		result, err := api.RenderString("Hello {{.name}}!", map[string]any{"name": "World"})
		if err != nil {
			t.Errorf("String render failed: %v", err)
		} else {
			t.Logf("String render result: %s", result)
			if result != "Hello World!" {
				t.Errorf("Expected 'Hello World!', got '%s'", result)
			}
		}
	})

	// 4. 测试模板函数
	t.Run("TemplateFunctions", func(t *testing.T) {
		// 测试内置函数
		builtinFuncs := view.GetBuiltinTemplateFunctions()
		t.Logf("内置函数数量: %d", len(builtinFuncs))
		
		if len(builtinFuncs) == 0 {
			t.Error("Expected builtin functions to be available")
		}
		
		// 检查关键函数是否存在
		expectedFuncs := []string{"add", "sub", "eq", "ne", "toString", "formatTime"}
		for _, funcName := range expectedFuncs {
			if _, exists := builtinFuncs[funcName]; !exists {
				t.Errorf("Expected builtin function '%s' not found", funcName)
			}
		}
		
		// 测试全局函数管理器
		manager := view.GetGlobalFunctionManager()
		functions := manager.ListFunctions()
		t.Logf("全局函数分类: %v", functions)
		
		if len(functions["builtin"]) == 0 {
			t.Error("Expected builtin functions in global manager")
		}
	})

	// 5. 测试Beego函数
	t.Run("BeegoFunctions", func(t *testing.T) {
		// 测试字符串函数
		api := view.GetUnifiedAPI()
		
		template := `
{{- $text := "Hello World" -}}
Length: {{len $text}}
Upper: {{toupper $text}}
Lower: {{tolower $text}}
Truncate: {{truncate $text 5}}
`
		
		result, err := api.RenderString(template, nil)
		if err != nil {
			t.Errorf("Beego functions test failed: %v", err)
		} else {
			t.Logf("Beego functions result:\n%s", result)
		}
	})

	// 6. 性能测试
	t.Run("PerformanceTest", func(t *testing.T) {
		api := view.GetUnifiedAPI()
		
		template := "Hello {{.name}}, current time is {{.time}}"
		data := map[string]any{
			"name": "Performance Test",
			"time": time.Now().Format("2006-01-02 15:04:05"),
		}
		
		// 测试单次渲染性能
		start := time.Now()
		_, err := api.RenderString(template, data)
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("Performance test failed: %v", err)
		} else {
			t.Logf("Single render took: %v", duration)
		}
		
		// 测试批量渲染性能
		const numRenders = 1000
		start = time.Now()
		
		for i := range numRenders {
			data["iteration"] = i
			_, err := api.RenderString(template, data)
			if err != nil {
				t.Errorf("Batch render failed at iteration %d: %v", i, err)
				break
			}
		}
		
		batchDuration := time.Since(start)
		avgDuration := batchDuration / numRenders
		t.Logf("Batch render (%d iterations) took: %v, avg per render: %v", 
			numRenders, batchDuration, avgDuration)
	})

	// 7. 测试错误处理和稳定性
	t.Run("ErrorHandling", func(t *testing.T) {
		api := view.GetUnifiedAPI()
		
		// 测试无效模板
		_, err := api.RenderString("{{.invalid.nested.property}}", map[string]any{"data": "test"})
		if err == nil {
			t.Log("Invalid template handled gracefully")
		} else {
			t.Logf("Invalid template error (expected): %v", err)
		}
		
		// 测试空数据
		result, err := api.RenderString("Static content", nil)
		if err != nil {
			t.Errorf("Static content render failed: %v", err)
		} else if result != "Static content" {
			t.Errorf("Expected 'Static content', got '%s'", result)
		}
	})

	t.Log("=== Beego模板系统测试完成 ===")
}

// TestSystemIntegration 测试系统集成
func TestSystemIntegration(t *testing.T) {
	t.Log("=== 开始系统集成测试 ===")
	
	// 测试多个组件协同工作
	t.Run("ComponentIntegration", func(t *testing.T) {
		// 1. 配置命名约定
		view.SetNamingConvention(view.BeegoStandard)
		
		// 2. 添加自定义映射
		view.AddControllerMapping("CustomController", "special")
		view.AddActionMapping("SpecialAction", "action")
		
		// 3. 测试推导
		candidates := view.InferTemplatePath("CustomController", "SpecialAction")
		t.Logf("Custom mapping result: %v", candidates)
		
		// 应该包含自定义映射的结果
		if !slices.Contains(candidates, "special/action") {
			t.Error("Expected custom mapping to be reflected in inference results")
		}
	})
	
	// 测试缓存和性能
	t.Run("CachePerformance", func(t *testing.T) {
		api := view.GetUnifiedAPI()
		template := "Cache test {{.value}}"
		
		// 多次渲染相同模板测试缓存效果
		const iterations = 100
		start := time.Now()
		
		for i := range iterations {
			_, err := api.RenderString(template, map[string]any{"value": i})
			if err != nil {
				t.Errorf("Cache test failed at iteration %d: %v", i, err)
				break
			}
		}
		
		duration := time.Since(start)
		t.Logf("Cache performance test (%d iterations): %v", iterations, duration)
	})
	
	t.Log("=== 系统集成测试完成 ===")
}

// BenchmarkTemplateRendering Benchmark测试
func BenchmarkTemplateRendering(b *testing.B) {
	api := view.GetUnifiedAPI()
	template := "Benchmark test {{.name}} - {{.value}}"
	data := map[string]any{
		"name":  "Performance",
		"value": 42,
	}
	
	b.ResetTimer()
	for range b.N {
		_, err := api.RenderString(template, data)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// TestExampleUsage 使用示例测试
func TestExampleUsage(t *testing.T) {
	t.Log("=== 使用示例测试 ===")
	
	// 示例1：基本模板渲染
	t.Run("BasicUsage", func(t *testing.T) {
		html, err := view.UnifiedRenderString("Hello {{.name}}!", map[string]any{
			"name": "Beego Template System",
		})
		
		if err != nil {
			t.Errorf("Basic usage failed: %v", err)
		} else {
			t.Logf("Basic usage result: %s", html)
		}
	})
	
	// 示例2：使用Beego函数
	t.Run("BeegoFunctionUsage", func(t *testing.T) {
		html, err := view.UnifiedRenderString(`
<div>
	<h1>{{toupper .title}}</h1>
	<p>Created: {{dateformat .date "2006-01-02 15:04:05"}}</p>
	<p>Truncated: {{truncate .content 50}}</p>
	<p>Math: {{add .price .tax}}</p>
</div>`, map[string]any{
			"title":   "welcome to our site",
			"date":    time.Now(),
			"content": "This is a very long content that should be truncated for display purposes",
			"price":   100,
			"tax":     15,
		})
		
		if err != nil {
			t.Errorf("Beego function usage failed: %v", err)
		} else {
			t.Logf("Beego function usage result:\n%s", html)
		}
	})
	
	// 示例3：自动推导渲染
	t.Run("AutoInferenceUsage", func(t *testing.T) {
		// 这个测试会失败因为没有实际的模板文件，但展示了使用方法
		_, err := view.AutoRender("UserController", "Profile", map[string]any{
			"username": "john_doe",
			"email":    "john@example.com",
		})
		
		if err != nil {
			t.Logf("Auto inference (expected to fail without template files): %v", err)
		}
	})
	
	t.Log("=== 使用示例测试完成 ===")
}

// 注释：要运行测试，请使用 'go test' 命令