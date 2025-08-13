// Package session 性能基准测试
//
// 这个文件包含对比重构前后性能表现的基准测试，包括：
// - Cookie操作性能测试
// - Session操作性能测试  
// - 安全Cookie性能测试
// - 代理层开销测试
// - 内存分配测试
package session

import (
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/route"
)

// 创建测试用的RequestContext
func createTestRequestContext() *app.RequestContext {
	ctx := app.NewContext(10)
	_ = route.NewEngine(config.NewOptions([]config.Option{}))
	ctx.Request.SetRequestURI("http://localhost:8080/test")
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.Header.Set("User-Agent", "test-agent")
	ctx.Request.Header.Set("Cookie", "test_cookie=test_value; session_id=abc123")
	
	return ctx
}

// ============= Cookie 操作基准测试 =============

// BenchmarkBaseCookieGet 测试基础Cookie获取性能
func BenchmarkBaseCookieGet(b *testing.B) {
	ctx := createTestRequestContext()
	cookie := NewBaseCookie(ctx)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = cookie.Get("test_cookie")
	}
}

// BenchmarkBaseCookieSetPerf 测试基础Cookie设置性能
func BenchmarkBaseCookieSetPerf(b *testing.B) {
	ctx := createTestRequestContext()
	cookie := NewBaseCookie(ctx)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		cookie.Set("bench_cookie", "bench_value", 3600)
	}
}

// BenchmarkBaseCookieDelete 测试Cookie删除性能
func BenchmarkBaseCookieDelete(b *testing.B) {
	ctx := createTestRequestContext()
	cookie := NewBaseCookie(ctx)
	
	// 预设一些cookie
	for i := 0; i < 10; i++ {
		cookie.Set("test_cookie_"+string(rune(i)), "value", 3600)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		cookie.Delete("test_cookie_" + string(rune(i%10)))
	}
}

// BenchmarkBaseCookieGetAll 测试获取所有Cookie性能
func BenchmarkBaseCookieGetAll(b *testing.B) {
	ctx := createTestRequestContext()
	cookie := NewBaseCookie(ctx)
	
	// 预设一些cookie
	for i := 0; i < 20; i++ {
		cookie.Set("test_cookie_"+string(rune(i)), "value", 3600)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = cookie.GetAll()
	}
}

// ============= 安全Cookie基准测试 =============

// BenchmarkSecureCookieSetPerf 测试安全Cookie设置性能
func BenchmarkSecureCookieSetPerf(b *testing.B) {
	ctx := createTestRequestContext()
	secureCookie := NewSecureCookie(ctx)
	secret := "test-secret-key-for-hmac-256"
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		secureCookie.SetSecure(secret, "secure_test", "secure_value")
	}
}

// BenchmarkSecureCookieGet 测试安全Cookie获取性能
func BenchmarkSecureCookieGet(b *testing.B) {
	ctx := createTestRequestContext()
	secureCookie := NewSecureCookie(ctx)
	secret := "test-secret-key-for-hmac-256"
	
	// 预设安全cookie
	secureCookie.SetSecure(secret, "secure_test", "secure_value")
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, _ = secureCookie.GetSecure("secure_test", secret)
	}
}

// BenchmarkSecureCookieWithOptions 测试带选项的安全Cookie性能
func BenchmarkSecureCookieWithOptions(b *testing.B) {
	ctx := createTestRequestContext()
	secureCookie := NewSecureCookie(ctx)
	
	options := CookieSecurityOptions{
		Secret:         "test-secret-key-for-hmac-256",
		MaxAge:         time.Hour,
		ValidateExpiry: true,
		RequireHTTPS:   false,
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = secureCookie.SetSecureWithOptions("secure_opt", "value", options)
	}
}

// ============= Session操作基准测试 =============

// BenchmarkSessionStart 测试Session启动性能
func BenchmarkSessionStart(b *testing.B) {
	ctx := createTestRequestContext()
	extension := NewExtensionForHertzContext(ctx)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		adapter := extension.StartSession()
		_ = adapter
	}
}

// BenchmarkSessionSetGet 测试Session设置和获取性能
func BenchmarkSessionSetGet(b *testing.B) {
	ctx := createTestRequestContext()
	extension := NewExtensionForHertzContext(ctx)
	adapter := extension.StartSession()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = adapter.Set("test_key", "test_value")
		_ = adapter.Get("test_key")
	}
}

// BenchmarkSessionBatchOperations 测试Session批量操作性能
func BenchmarkSessionBatchOperations(b *testing.B) {
	ctx := createTestRequestContext()
	extension := NewExtensionForHertzContext(ctx)
	adapter := extension.StartSession()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// 批量设置
		for j := 0; j < 10; j++ {
			_ = adapter.Set("key_"+string(rune(j)), "value")
		}
		
		// 批量获取
		for j := 0; j < 10; j++ {
			_ = adapter.Get("key_" + string(rune(j)))
		}
		
		// 清空
		adapter.Flush()
	}
}

// ============= ContextExtension基准测试 =============

// BenchmarkContextExtensionCreation 测试Context扩展创建性能
func BenchmarkContextExtensionCreation(b *testing.B) {
	ctx := createTestRequestContext()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		ext := NewExtensionForHertzContext(ctx)
		_ = ext
	}
}

// BenchmarkContextExtensionLazyInit 测试延迟初始化性能
func BenchmarkContextExtensionLazyInit(b *testing.B) {
	ctx := createTestRequestContext()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		ext := NewExtensionForHertzContext(ctx)
		_ = ext
	}
}

// BenchmarkContextExtensionMixedOperations 测试混合操作性能
func BenchmarkContextExtensionMixedOperations(b *testing.B) {
	ctx := createTestRequestContext()
	extension := NewExtensionForHertzContext(ctx)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Cookie操作
		extension.SetCookie("bench_cookie", "bench_value")
		_ = extension.GetCookie("bench_cookie")
		
		// 安全Cookie操作
		extension.SetSecureCookie("secret", "secure_bench", "secure_value")
		_, _ = extension.GetSecureCookie("secret", "secure_bench")
		
		// Session操作
		_ = extension.SetSession("session_key", "session_value")
		_ = extension.GetSession("session_key")
	}
}

// ============= 代理层开销测试 (模拟context/adapter.go中的代理) =============

// 模拟InputData结构和代理方法
type BenchmarkInputData struct {
	ctx       *app.RequestContext
	extension *ContextExtension
}

func (i *BenchmarkInputData) getExtension() *ContextExtension {
	if i.extension == nil && i.ctx != nil {
		i.extension = NewExtensionForHertzContext(i.ctx)
	}
	return i.extension
}

func (i *BenchmarkInputData) Cookie(key string) string {
	if ext := i.getExtension(); ext != nil {
		return ext.GetCookie(key)
	}
	return ""
}

func (i *BenchmarkInputData) SetCookie(name, value string, others ...interface{}) {
	if ext := i.getExtension(); ext != nil {
		ext.SetCookie(name, value, others...)
	}
}

// BenchmarkProxyLayerOverhead 测试代理层开销
func BenchmarkProxyLayerOverhead(b *testing.B) {
	ctx := createTestRequestContext()
	inputData := &BenchmarkInputData{ctx: ctx}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		inputData.SetCookie("proxy_test", "proxy_value")
		_ = inputData.Cookie("proxy_test")
	}
}

// BenchmarkDirectVsProxy 对比直接调用vs代理调用性能
func BenchmarkDirectVsProxy(b *testing.B) {
	ctx := createTestRequestContext()
	
	b.Run("Direct", func(b *testing.B) {
		ext := NewExtensionForHertzContext(ctx)
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			ext.SetCookie("direct_test", "direct_value")
			_ = ext.GetCookie("direct_test")
		}
	})
	
	b.Run("Proxy", func(b *testing.B) {
		inputData := &BenchmarkInputData{ctx: ctx}
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			inputData.SetCookie("proxy_test", "proxy_value")
			_ = inputData.Cookie("proxy_test")
		}
	})
}

// ============= 内存分配测试 =============

// BenchmarkMemoryAllocation 测试内存分配模式
func BenchmarkMemoryAllocation(b *testing.B) {
	ctx := createTestRequestContext()
	
	b.Run("Cookie_Operations", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cookie := NewBaseCookie(ctx)
			cookie.Set("mem_test", "mem_value", 3600)
			_ = cookie.Get("mem_test")
			_ = cookie.GetAll()
		}
	})
	
	b.Run("SecureCookie_Operations", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			secureCookie := NewSecureCookie(ctx)
			secureCookie.SetSecure("secret", "mem_secure", "mem_value")
			_, _ = secureCookie.GetSecure("secret", "mem_secure")
		}
	})
	
	b.Run("Session_Operations", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ext := NewExtensionForHertzContext(ctx)
			adapter := ext.StartSession()
			_ = adapter.Set("mem_session", "mem_value")
			_ = adapter.Get("mem_session")
		}
	})
}

// ============= 并发性能测试 =============

// BenchmarkConcurrentCookieOperations 测试并发Cookie操作
func BenchmarkConcurrentCookieOperations(b *testing.B) {
	ctx := createTestRequestContext()
	extension := NewExtensionForHertzContext(ctx)
	
	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(10)
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "concurrent_" + string(rune(i%100))
			extension.SetCookie(key, "value")
			_ = extension.GetCookie(key)
			i++
		}
	})
}

// BenchmarkConcurrentSessionOperations 测试并发Session操作
func BenchmarkConcurrentSessionOperations(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(10)
	
	b.RunParallel(func(pb *testing.PB) {
		ctx := createTestRequestContext() // 每个goroutine一个独立context
		extension := NewExtensionForHertzContext(ctx)
		adapter := extension.StartSession()
		
		i := 0
		for pb.Next() {
			key := "session_" + string(rune(i%100))
			_ = adapter.Set(key, "value")
			_ = adapter.Get(key)
			i++
		}
	})
}

// ============= 性能回归测试辅助函数 =============

// BenchmarkSuite 运行完整的性能测试套件
func BenchmarkSuite(b *testing.B) {
	ctx := createTestRequestContext()
	
	// 这个基准测试用于CI/CD中的性能回归检测
	b.Run("Critical_Path", func(b *testing.B) {
		extension := NewExtensionForHertzContext(ctx)
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			// 模拟典型的web请求场景
			extension.SetCookie("session_id", "abc123")
			_ = extension.GetCookie("session_id")
			
			adapter := extension.StartSession()
			_ = adapter.Set("user_id", "12345")
			_ = adapter.Get("user_id")
			
			extension.SetSecureCookie("secret-key", "csrf_token", "token_value")
			_, _ = extension.GetSecureCookie("secret-key", "csrf_token")
		}
	})
}

// ============= 基准测试辅助函数 =============

// 运行基准测试的辅助脚本
/*
运行方法:
go test -bench=. -benchmem -count=5 -cpu=1,2,4

详细分析:
go test -bench=BenchmarkBaseCookieGet -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
go tool pprof cpu.prof
go tool pprof mem.prof

性能对比:
benchcmp old.txt new.txt

示例输出解读:
BenchmarkBaseCookieGet-4    	1000000	  1234 ns/op	  128 B/op	   3 allocs/op
                               |       |           |       |         |
                               |       |           |       |         分配次数
                               |       |           |       每次操作分配字节数  
                               |       |           每次操作耗时
                               |       总执行次数
                               CPU核心数

性能目标:
- Cookie操作: < 500ns/op, < 64B/op
- 安全Cookie: < 2000ns/op, < 256B/op  
- Session操作: < 1000ns/op, < 128B/op
- 代理开销: < 10% 性能损失
*/