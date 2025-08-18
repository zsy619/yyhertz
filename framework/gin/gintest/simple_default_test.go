// Package gin - 简化的Default()性能测试
package gin

import (
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/gin"
)

// 简单的基准测试：对比单例vs新实例创建
func BenchmarkSimpleDefault(b *testing.B) {
	b.Run("Singleton", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = gin.Default()
		}
	})

	b.Run("NewInstance", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = gin.Default(func(e *gin.Engine) {})
		}
	})

	b.Run("ManualNew", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine := gin.New()
			engine.Use(gin.Logger(), gin.Recovery())
		}
	})
}

// 验证单例确实工作
func TestSingletonWorking(t *testing.T) {

	// 获取多个实例
	e1 := gin.Default()
	e2 := gin.Default()
	e3 := gin.Default()

	// 验证都是同一个实例
	if e1 != e2 || e2 != e3 {
		t.Error("Default() should return the same instance")
	}

	// 验证与DefaultGlobal()返回相同实例
	e4 := gin.DefaultGlobal()
	if e1 != e4 {
		t.Error("Default() and DefaultGlobal() should return the same instance")
	}

	t.Log("✅ Singleton pattern working correctly")
}

// 验证自定义配置创建新实例
func TestCustomConfigNewInstance(t *testing.T) {
	// 获取单例
	singleton := gin.Default()

	// 获取自定义配置实例
	custom1 := gin.Default(func(e *gin.Engine) {})
	custom2 := gin.Default(func(e *gin.Engine) {})

	// 验证自定义实例与单例不同
	if custom1 == singleton {
		t.Error("Custom instance should be different from singleton")
	}

	// 验证多个自定义实例互不相同
	if custom1 == custom2 {
		t.Error("Multiple custom instances should be different")
	}

	t.Log("✅ Custom configuration creates new instances correctly")
}

// 性能对比测试
func TestPerformanceComparison(t *testing.T) {
	const iterations = 10000

	// 测试单例模式性能
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = gin.Default()
	}
	singletonDuration := time.Since(start)

	// 测试新实例创建性能
	start = time.Now()
	for i := 0; i < iterations; i++ {
		engine := gin.New()
		engine.Use(gin.Logger(), gin.Recovery())
	}
	newInstanceDuration := time.Since(start)

	// 计算性能提升
	speedup := float64(newInstanceDuration) / float64(singletonDuration)

	t.Logf("Performance comparison (%d iterations):", iterations)
	t.Logf("  Singleton mode: %v", singletonDuration)
	t.Logf("  New instance mode: %v", newInstanceDuration)
	t.Logf("  Speedup: %.2fx", speedup)

	// 单例应该显著更快
	if speedup < 2.0 {
		t.Logf("Warning: Expected at least 2x speedup, got %.2fx", speedup)
	} else {
		t.Logf("✅ Significant performance improvement achieved")
	}
}
