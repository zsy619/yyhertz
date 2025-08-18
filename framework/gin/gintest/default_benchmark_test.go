// Package gin - Default() 函数性能测试
// 验证全局单例优化的性能提升效果
package gin

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/gin"
)

// BenchmarkDefault_WithoutOptions 测试无参数调用的性能（单例模式）
func BenchmarkDefault_WithoutOptions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gin.Default()
	}
}

// BenchmarkDefault_WithOptions 测试有参数调用的性能（新实例模式）
func BenchmarkDefault_WithOptions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gin.Default(func(e *gin.Engine) {
			// 简单的配置函数
		})
	}
}

// BenchmarkNew_Baseline 基线测试：直接使用New()的性能
func BenchmarkNew_Baseline(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := gin.New()
		engine.Use(gin.Logger(), gin.Recovery())
	}
}

// BenchmarkDefaultGlobal 测试专用的全局单例函数性能
func BenchmarkDefaultGlobal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gin.DefaultGlobal()
	}
}

// TestDefault_MemoryUsage 测试内存使用情况
func TestDefault_MemoryUsage(t *testing.T) {
	var m1, m2 runtime.MemStats

	// 记录初始内存状态
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// 创建1000个实例（应该只有一个单例）
	engines := make([]*gin.Engine, 1000)
	for i := 0; i < 1000; i++ {
		engines[i] = gin.Default()
	}

	// 验证所有引用都指向同一个实例
	for i := 1; i < 1000; i++ {
		if engines[i] != engines[0] {
			t.Errorf("Expected all engines to be the same instance, but engine[%d] is different", i)
		}
	}

	// 记录内存使用
	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocatedBytes := m2.TotalAlloc - m1.TotalAlloc
	t.Logf("Memory allocated for 1000 Default() calls: %d bytes", allocatedBytes)
	t.Logf("Memory per instance (should be near zero): %.2f bytes", float64(allocatedBytes)/1000.0)

	// 验证内存使用应该很小（因为是单例）
	if allocatedBytes > 10000 { // 允许10KB的误差
		t.Errorf("Expected minimal memory allocation, but got %d bytes", allocatedBytes)
	}
}

// TestDefault_ThreadSafety 测试线程安全性
func TestDefault_ThreadSafety(t *testing.T) {
	const numGoroutines = 100
	const callsPerGoroutine = 100

	var wg sync.WaitGroup
	engines := make([][]*gin.Engine, numGoroutines)

	// 启动多个goroutine并发调用Default()
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			engines[index] = make([]*gin.Engine, callsPerGoroutine)
			for j := 0; j < callsPerGoroutine; j++ {
				engines[index][j] = gin.Default()
			}
		}(i)
	}

	wg.Wait()

	// 验证所有goroutine获得的都是同一个实例
	firstEngine := engines[0][0]
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < callsPerGoroutine; j++ {
			if engines[i][j] != firstEngine {
				t.Errorf("Thread safety violation: engine[%d][%d] is not the same instance", i, j)
			}
		}
	}

	t.Logf("Thread safety test passed: %d goroutines × %d calls all returned the same instance",
		numGoroutines, callsPerGoroutine)
}

// TestDefault_WithVsWithoutOptions 对比有参数和无参数的行为
func TestDefault_WithVsWithoutOptions(t *testing.T) {
	// 无参数调用（应该返回单例）
	engine1 := gin.Default()
	engine2 := gin.Default()

	if engine1 != engine2 {
		t.Error("Expected Default() calls without options to return the same instance")
	}

	// 有参数调用（应该返回新实例）
	engine3 := gin.Default(func(e *gin.Engine) {})
	engine4 := gin.Default(func(e *gin.Engine) {})

	if engine3 == engine4 {
		t.Error("Expected Default() calls with options to return different instances")
	}

	// 无参数和有参数不应该是同一个实例
	if engine1 == engine3 {
		t.Error("Expected Default() with options to return different instance from singleton")
	}

	t.Logf("Behavior validation passed:")
	t.Logf("  Default() without options: singleton behavior ✓")
	t.Logf("  Default() with options: new instance behavior ✓")
}

// BenchmarkConcurrentDefault 并发访问基准测试
func BenchmarkConcurrentDefault(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = gin.Default()
		}
	})
}

// TestDefault_InitializationTime 测试初始化时间
func TestDefault_InitializationTime(t *testing.T) {
	// 第一次调用（需要初始化）
	start := time.Now()
	engine1 := gin.Default()
	firstCallDuration := time.Since(start)

	// 后续调用（直接返回）
	start = time.Now()
	engine2 := gin.Default()
	secondCallDuration := time.Since(start)

	if engine1 != engine2 {
		t.Error("Expected both calls to return the same instance")
	}

	t.Logf("First call (with initialization): %v", firstCallDuration)
	t.Logf("Second call (singleton return): %v", secondCallDuration)

	// 后续调用应该显著更快
	if secondCallDuration > firstCallDuration/2 {
		t.Logf("Warning: Second call not significantly faster (expected optimization)")
	}
}

// ExampleDefault_usage 展示使用方式的示例
func ExampleDefault_usage() {
	// 推荐用法：无参数调用，使用单例实例
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 需要自定义配置时，传入配置函数
	customR := gin.Default(func(e *gin.Engine) {
		// 自定义配置
	})
	customR.GET("/custom", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "custom"})
	})
}
