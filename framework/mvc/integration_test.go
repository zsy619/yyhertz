package mvc

import (
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/errors"
	"github.com/zsy619/yyhertz/framework/mvc/context"
	mvcerrros "github.com/zsy619/yyhertz/framework/mvc/errors"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
)

// TestEnhancedFastEngine 测试增强的FastEngine
func TestEnhancedFastEngine(t *testing.T) {
	// 创建增强引擎
	engine := NewForDevelopment()
	
	// 测试中间件注册
	testMiddlewareRegistration(t, engine)
	
	// 测试错误处理
	testErrorHandling(t, engine)
	
	// 测试统计信息
	testStatistics(t, engine)
	
	// 打印系统信息
	engine.PrintSystemStatistics()
}

// testMiddlewareRegistration 测试中间件注册
func testMiddlewareRegistration(t *testing.T, engine *EnhancedFastEngine) {
	fmt.Println("\n=== Testing Middleware Registration ===")
	
	// 注册内置中间件
	engine.UseBuiltin("logger", nil, 10)
	engine.UseBuiltin("recovery", nil, 5)
	engine.UseBuiltin("cors", nil, 20)
	
	// 注册自定义中间件
	engine.Use("test-middleware", func(ctx *context.Context) {
		fmt.Println("Test middleware executed")
		ctx.Next()
	}, 15)
	
	// 注册不同层级的中间件
	engine.UseGroup("group-middleware", func(ctx *context.Context) {
		fmt.Println("Group middleware executed")
		ctx.Next()
	}, 10)
	
	engine.UseRoute("route-middleware", func(ctx *context.Context) {
		fmt.Println("Route middleware executed")
		ctx.Next()
	}, 10)
	
	engine.UseController("controller-middleware", func(ctx *context.Context) {
		fmt.Println("Controller middleware executed")
		ctx.Next()
	}, 10)
	
	// 获取编译后的中间件链
	compiled, err := engine.GetMiddlewareManager().GetCompiledChain(
		middleware.LayerGlobal,
		middleware.LayerGroup,
		middleware.LayerRoute,
		middleware.LayerController,
	)
	
	if err != nil {
		t.Errorf("Failed to get compiled chain: %v", err)
		return
	}
	
	fmt.Printf("✓ Compiled middleware chain created with %d middlewares\n", len(compiled.OptimizedChain))
}

// testErrorHandling 测试错误处理
func testErrorHandling(t *testing.T, engine *EnhancedFastEngine) {
	fmt.Println("\n=== Testing Error Handling ===")
	
	// 注册自定义错误处理器
	engine.RegisterErrorHandlerFunc("test-handler", 50, func(err error) bool {
		return err.Error() == "test error"
	}, func(ctx *context.Context, err error) error {
		fmt.Printf("✓ Custom error handler processed: %v\n", err)
		return nil
	})
	
	// 测试错误分类
	testErrors := []error{
		errors.TimeoutError,
		errors.NetworkError,
		errors.UserNotExist,
		errors.PermissionDenied,
		fmt.Errorf("unknown error"),
	}
	
	for _, err := range testErrors {
		classification := engine.GetErrorClassifier().Classify(err, nil)
		fmt.Printf("✓ Error '%v' classified as: Category=%s, Severity=%s, Retryable=%v\n",
			err,
			mvcerrros.GetCategoryName(classification.Category),
			mvcerrros.GetSeverityName(classification.Severity),
			classification.Retryable)
	}
	
	// 测试自动恢复
	timeoutErr := errors.TimeoutError
	result := engine.GetAutoRecovery().Recover(nil, timeoutErr)
	fmt.Printf("✓ Recovery result for timeout error: Strategy=%s, Action=%s, Success=%v\n",
		result.Strategy, mvcerrros.GetActionName(result.Action), result.Success)
}

// testStatistics 测试统计信息
func testStatistics(t *testing.T, engine *EnhancedFastEngine) {
	fmt.Println("\n=== Testing Statistics ===")
	
	stats := engine.GetSystemStatistics()
	
	fmt.Printf("✓ Middleware statistics: %d registered, %d executions\n",
		stats.Middleware.Registry.TotalCount,
		stats.Middleware.Pipeline.ExecutionCount)
	
	fmt.Printf("✓ Error statistics: %d total errors, %d handled\n",
		stats.Error.TotalErrors,
		stats.Error.HandledErrors)
	
	fmt.Printf("✓ Classifier statistics: %d classifications, %.2f avg score\n",
		stats.Classifier.TotalClassified,
		stats.Classifier.AverageScore)
	
	fmt.Printf("✓ Recovery statistics: %d attempts, %d successful\n",
		stats.Recovery.TotalAttempts,
		stats.Recovery.SuccessfulRecoveries)
}

// BenchmarkMiddlewareChain 基准测试中间件链
func BenchmarkMiddlewareChain(b *testing.B) {
	engine := New()
	
	// 注册多个中间件
	for i := 0; i < 10; i++ {
		engine.Use(fmt.Sprintf("middleware-%d", i), func(ctx *context.Context) {
			ctx.Next()
		}, i)
	}
	
	// 获取编译后的链
	compiled, err := engine.GetMiddlewareManager().GetCompiledChain(middleware.LayerGlobal)
	if err != nil {
		b.Fatalf("Failed to compile chain: %v", err)
	}
	
	b.ResetTimer()
	
	// 基准测试中间件执行
	for i := 0; i < b.N; i++ {
		// 这里应该创建实际的上下文来测试
		// 由于依赖关系复杂，这里仅作为示例
		_ = compiled
	}
}

// BenchmarkErrorClassification 基准测试错误分类
func BenchmarkErrorClassification(b *testing.B) {
	engine := New()
	classifier := engine.GetErrorClassifier()
	
	testErr := fmt.Errorf("connection timeout")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		classification := classifier.Classify(testErr, nil)
		_ = classification
	}
}

// Example_basicUsage 基本使用示例
func Example_basicUsage() {
	// 创建增强的FastEngine
	engine := New()
	
	// 使用内置中间件
	engine.UseBuiltin("logger", nil, 10)
	engine.UseBuiltin("recovery", nil, 5)
	engine.UseBuiltin("cors", nil, 20)
	
	// 注册自定义中间件
	engine.Use("auth", func(ctx *context.Context) {
		// 认证逻辑
		fmt.Println("Authentication middleware")
		ctx.Next()
	}, 15)
	
	// 注册错误处理器
	engine.RegisterErrorHandlerFunc("business-error", 100, func(err error) bool {
		_, ok := err.(*errors.ErrNo)
		return ok
	}, func(ctx *context.Context, err error) error {
		fmt.Println("Business error handled")
		return nil
	})
	
	// 添加恢复策略
	engine.AddRecoveryStrategy(mvcerrros.RecoveryStrategy{
		Name:          "custom-retry",
		Condition:     &mvcerrros.CategoryCondition{Category: mvcerrros.CategoryNetwork},
		Action:        mvcerrros.ActionRetry,
		MaxRetries:    3,
		RetryInterval: time.Second,
	})
	
	fmt.Println("Enhanced FastEngine configured successfully")
	
	// Output: Enhanced FastEngine configured successfully
}

// Example_errorHandling 错误处理示例
func Example_errorHandling() {
	engine := NewForDevelopment()
	
	// 模拟错误处理
	testErr := errors.NetworkError
	
	// 分类错误
	classification := engine.GetErrorClassifier().Classify(testErr, nil)
	fmt.Printf("Error category: %s\n", mvcerrros.GetCategoryName(classification.Category))
	
	// 尝试恢复
	result := engine.GetAutoRecovery().Recover(nil, testErr)
	fmt.Printf("Recovery action: %s\n", mvcerrros.GetActionName(result.Action))
	
	// Output: Error category: Network
	// Recovery action: Retry
}

// Example_statistics 统计信息示例
func Example_statistics() {
	engine := New()
	
	// 模拟一些操作来产生统计数据
	engine.UseBuiltin("logger", nil, 10)
	
	// 获取统计信息
	stats := engine.GetSystemStatistics()
	
	fmt.Printf("Registered middlewares: %d\n", stats.Middleware.Registry.TotalCount)
	
	// Output: Registered middlewares: 1
}

// TestIntegrationConfig 测试集成配置
func TestIntegrationConfig(t *testing.T) {
	// 测试默认配置
	defaultConfig := config.GetDefaultMVCConfig()
	if !defaultConfig.Middleware.EnableOptimization {
		t.Error("Expected middleware optimization to be enabled by default")
	}
	
	// 测试生产环境配置
	engine := NewForProduction()
	if engine.config.Debug.EnableMode {
		t.Error("Expected debug mode to be disabled in production")
	}
	
	// 测试开发环境配置
	devEngine := NewForDevelopment()
	if !devEngine.config.Debug.EnableMode {
		t.Error("Expected debug mode to be enabled in development")
	}
}