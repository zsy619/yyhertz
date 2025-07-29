package util

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPerformanceMonitor(t *testing.T) {
	monitor := NewPerformanceMonitor()
	
	t.Run("测试执行跟踪", func(t *testing.T) {
		// 跟踪一个简单的函数执行
		monitor.TrackExecution("test_function", func() {
			time.Sleep(10 * time.Millisecond)
		})
		
		// 获取指标
		metric, exists := monitor.GetMetric("test_function")
		assert.True(t, exists)
		assert.Equal(t, int64(1), metric.Count)
		assert.True(t, metric.Total >= 10*time.Millisecond)
		assert.True(t, metric.Average >= 10*time.Millisecond)
	})
	
	t.Run("测试手动跟踪", func(t *testing.T) {
		stop := monitor.StartTracking("manual_track")
		time.Sleep(5 * time.Millisecond)
		stop()
		
		metric, exists := monitor.GetMetric("manual_track")
		assert.True(t, exists)
		assert.Equal(t, int64(1), metric.Count)
		assert.True(t, metric.Total >= 5*time.Millisecond)
	})
	
	t.Run("测试多次执行", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			monitor.TrackExecution("multi_exec", func() {
				time.Sleep(time.Millisecond)
			})
		}
		
		metric, exists := monitor.GetMetric("multi_exec")
		assert.True(t, exists)
		assert.Equal(t, int64(5), metric.Count)
		assert.True(t, metric.Min <= metric.Average)
		assert.True(t, metric.Average <= metric.Max)
	})
	
	t.Run("测试指标重置", func(t *testing.T) {
		monitor.TrackExecution("reset_test", func() {
			time.Sleep(time.Millisecond)
		})
		
		// 重置前应该有指标
		_, exists := monitor.GetMetric("reset_test")
		assert.True(t, exists)
		
		// 重置后应该没有指标
		monitor.Reset()
		_, exists = monitor.GetMetric("reset_test")
		assert.False(t, exists)
	})
	
	t.Run("测试获取所有指标", func(t *testing.T) {
		monitor.TrackExecution("metric1", func() { time.Sleep(time.Millisecond) })
		monitor.TrackExecution("metric2", func() { time.Sleep(time.Millisecond) })
		
		metrics := monitor.GetMetrics()
		assert.Len(t, metrics, 2)
		assert.Contains(t, metrics, "metric1")
		assert.Contains(t, metrics, "metric2")
	})
	
	t.Run("测试运行时间", func(t *testing.T) {
		uptime := monitor.GetUptime()
		assert.True(t, uptime > 0)
		
		time.Sleep(10 * time.Millisecond)
		newUptime := monitor.GetUptime()
		assert.True(t, newUptime > uptime)
	})
}

func TestMemoryPool(t *testing.T) {
	// 测试字符串内存池
	pool := NewMemoryPool(func() *string {
		s := ""
		return &s
	})
	
	t.Run("测试基本获取和归还", func(t *testing.T) {
		obj1 := pool.Get()
		assert.NotNil(t, obj1)
		
		*obj1 = "test"
		pool.Put(obj1)
		
		obj2 := pool.Get()
		assert.NotNil(t, obj2)
		// 注意：从池中获取的对象可能是重用的
	})
	
	t.Run("测试并发安全", func(t *testing.T) {
		var wg sync.WaitGroup
		const numGoroutines = 100
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				obj := pool.Get()
				*obj = "concurrent test"
				time.Sleep(time.Microsecond)
				pool.Put(obj)
			}()
		}
		
		wg.Wait()
	})
}

func TestBytePool(t *testing.T) {
	pool := NewBytePool()
	
	t.Run("测试不同大小的缓冲区", func(t *testing.T) {
		sizes := []int{1024, 2048, 4096}
		
		for _, size := range sizes {
			buf := pool.Get(size)
			assert.Len(t, buf, size)
			
			// 写入一些数据
			for i := 0; i < len(buf); i++ {
				buf[i] = byte(i % 256)
			}
			
			pool.Put(buf)
			
			// 再次获取，应该被清零
			buf2 := pool.Get(size)
			assert.Len(t, buf2, size)
			for i := 0; i < len(buf2); i++ {
				assert.Equal(t, byte(0), buf2[i])
			}
		}
	})
	
	t.Run("测试nil缓冲区处理", func(t *testing.T) {
		// 不应该panic
		assert.NotPanics(t, func() {
			pool.Put(nil)
		})
	})
}

func TestStringPool(t *testing.T) {
	pool := NewStringPool()
	
	t.Run("测试字符串构建器复用", func(t *testing.T) {
		sb1 := pool.Get()
		assert.NotNil(t, sb1)
		
		sb1.WriteString("hello")
		sb1.WriteString(" world")
		result1 := sb1.String()
		
		pool.Put(sb1)
		
		// 获取另一个构建器，应该是重置过的
		sb2 := pool.Get()
		assert.NotNil(t, sb2)
		assert.Equal(t, 0, sb2.Len()) // 应该被重置
		
		sb2.WriteString("test")
		result2 := sb2.String()
		
		assert.Equal(t, "hello world", result1)
		assert.Equal(t, "test", result2)
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("测试基本限流", func(t *testing.T) {
		limiter := NewRateLimiter(5, 1) // 5个令牌，每秒补充1个
		
		// 应该能够立即获得5个令牌
		for i := 0; i < 5; i++ {
			assert.True(t, limiter.Allow(), "第%d次请求应该被允许", i+1)
		}
		
		// 第6个请求应该被拒绝
		assert.False(t, limiter.Allow(), "第6次请求应该被拒绝")
		
		// 等待令牌补充
		time.Sleep(1100 * time.Millisecond)
		assert.True(t, limiter.Allow(), "等待后的请求应该被允许")
	})
	
	t.Run("测试批量请求", func(t *testing.T) {
		limiter := NewRateLimiter(10, 2)
		
		// 请求5个令牌
		assert.True(t, limiter.AllowN(5))
		// 再请求5个令牌
		assert.True(t, limiter.AllowN(5))
		// 再请求1个令牌，应该被拒绝
		assert.False(t, limiter.AllowN(1))
	})
	
	t.Run("测试令牌数获取", func(t *testing.T) {
		limiter := NewRateLimiter(3, 1)
		
		assert.Equal(t, int64(3), limiter.GetTokens())
		limiter.Allow()
		assert.Equal(t, int64(2), limiter.GetTokens())
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("测试正常状态", func(t *testing.T) {
		cb := NewCircuitBreaker(3, time.Second)
		
		// 成功调用应该正常工作
		err := cb.Call(func() error {
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, CBStateClosed, cb.GetState())
	})
	
	t.Run("测试熔断触发", func(t *testing.T) {
		cb := NewCircuitBreaker(2, time.Second)
		
		// 失败2次应该触发熔断
		for i := 0; i < 2; i++ {
			err := cb.Call(func() error {
				return assert.AnError
			})
			assert.NotNil(t, err)
		}
		
		assert.Equal(t, CBStateOpen, cb.GetState())
		
		// 熔断状态下的调用应该被拒绝
		err := cb.Call(func() error {
			return nil
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "circuit breaker is open")
	})
	
	t.Run("测试熔断恢复", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 100*time.Millisecond)
		
		// 触发熔断
		cb.Call(func() error {
			return assert.AnError
		})
		assert.Equal(t, CBStateOpen, cb.GetState())
		
		// 等待重置时间
		time.Sleep(150 * time.Millisecond)
		
		// 成功调用应该恢复熔断器
		err := cb.Call(func() error {
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, CBStateClosed, cb.GetState())
	})
}

func TestSystemInfo(t *testing.T) {
	t.Run("测试系统信息获取", func(t *testing.T) {
		info := GetSystemInfo()
		
		assert.NotEmpty(t, info.GoVersion)
		assert.True(t, info.NumCPU > 0)
		assert.True(t, info.NumGoroutine > 0)
		assert.True(t, info.MemStats.Alloc > 0)
		assert.True(t, info.MemStats.TotalAlloc > 0)
		assert.True(t, info.MemStats.Sys > 0)
	})
	
	t.Run("测试内存使用获取", func(t *testing.T) {
		usage := GetMemoryUsage()
		
		expectedKeys := []string{
			"alloc_mb", "total_alloc_mb", "sys_mb", 
			"heap_alloc_mb", "heap_sys_mb",
		}
		
		for _, key := range expectedKeys {
			assert.Contains(t, usage, key)
			assert.True(t, usage[key] > 0, "内存使用量%s应该大于0", key)
		}
	})
	
	t.Run("测试强制GC", func(t *testing.T) {
		beforeGC := GetSystemInfo()
		
		// 分配一些内存
		data := make([][]byte, 1000)
		for i := range data {
			data[i] = make([]byte, 1024)
		}
		
		ForceGC()
		
		afterGC := GetSystemInfo()
		
		// GC后，NumGC应该增加
		assert.True(t, afterGC.MemStats.NumGC >= beforeGC.MemStats.NumGC)
		
		// 释放引用
		data = nil
	})
}

// 并发测试
func TestConcurrentPerformanceMonitor(t *testing.T) {
	monitor := NewPerformanceMonitor()
	
	t.Run("测试并发跟踪", func(t *testing.T) {
		var wg sync.WaitGroup
		const numGoroutines = 100
		const numCalls = 10
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numCalls; j++ {
					monitor.TrackExecution("concurrent_test", func() {
						time.Sleep(time.Microsecond)
					})
				}
			}(i)
		}
		
		wg.Wait()
		
		metric, exists := monitor.GetMetric("concurrent_test")
		assert.True(t, exists)
		assert.Equal(t, int64(numGoroutines*numCalls), metric.Count)
	})
}

// 基准测试
func BenchmarkPerformanceMonitor(b *testing.B) {
	monitor := NewPerformanceMonitor()
	
	b.Run("TrackExecution", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			monitor.TrackExecution("benchmark", func() {
				// 模拟一些工作
			})
		}
	})
	
	b.Run("StartTracking", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stop := monitor.StartTracking("benchmark_manual")
			stop()
		}
	})
}

func BenchmarkMemoryPool(b *testing.B) {
	pool := NewMemoryPool(func() *[]byte {
		buf := make([]byte, 1024)
		return &buf
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := pool.Get()
		pool.Put(obj)
	}
}

func BenchmarkBytePool(b *testing.B) {
	pool := NewBytePool()
	
	b.Run("Size1024", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := pool.Get(1024)
			pool.Put(buf)
		}
	})
	
	b.Run("Size4096", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := pool.Get(4096)
			pool.Put(buf)
		}
	})
}

func BenchmarkStringPool(b *testing.B) {
	pool := NewStringPool()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb := pool.Get()
		sb.WriteString("benchmark test")
		_ = sb.String()
		pool.Put(sb)
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	limiter := NewRateLimiter(1000000, 1000000) // 足够大，避免限流影响测试
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow()
	}
}

func BenchmarkCircuitBreaker(b *testing.B) {
	cb := NewCircuitBreaker(1000, time.Hour) // 设置很高的阈值，避免熔断
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Call(func() error {
			return nil
		})
	}
}

// 内存基准测试
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("DirectAllocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 1024)
			_ = buf
		}
	})
	
	b.Run("PooledAllocation", func(b *testing.B) {
		pool := NewBytePool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := pool.Get(1024)
			pool.Put(buf)
		}
	})
}

// 测试全局实例
func TestGlobalInstances(t *testing.T) {
	t.Run("测试全局监控器", func(t *testing.T) {
		assert.NotNil(t, GlobalMonitor)
		
		GlobalMonitor.TrackExecution("global_test", func() {
			time.Sleep(time.Millisecond)
		})
		
		metric, exists := GlobalMonitor.GetMetric("global_test")
		assert.True(t, exists)
		assert.Equal(t, int64(1), metric.Count)
	})
	
	t.Run("测试全局字节池", func(t *testing.T) {
		assert.NotNil(t, GlobalBytePool)
		
		buf := GlobalBytePool.Get(1024)
		assert.Len(t, buf, 1024)
		GlobalBytePool.Put(buf)
	})
	
	t.Run("测试全局字符串池", func(t *testing.T) {
		assert.NotNil(t, GlobalStringPool)
		
		sb := GlobalStringPool.Get()
		assert.NotNil(t, sb)
		sb.WriteString("test")
		GlobalStringPool.Put(sb)
	})
}