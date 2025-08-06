package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/core"
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// BenchmarkRouterTree_AddRoute 测试路由添加性能
func BenchmarkRouterTree_AddRoute(b *testing.B) {
	handler := func(ctx context.Context, c *core.RequestContext) {}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		tree := NewRouterTree() // 每个goroutine使用独立的tree
		i := 0
		for pb.Next() {
			path := fmt.Sprintf("/test/%d", i)
			tree.AddRoute("GET", path, handler)
			i++
		}
	})
}

// BenchmarkRouterTree_GetRoute 测试路由查找性能
func BenchmarkRouterTree_GetRoute(b *testing.B) {
	tree := NewRouterTree()
	handler := func(ctx context.Context, c *core.RequestContext) {}
	
	// 预先添加路由
	paths := generateTestPaths(1000)
	for _, path := range paths {
		tree.AddRoute("GET", path, handler)
	}
	tree.Compile()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := paths[i%len(paths)]
			_, _ = tree.GetRoute("GET", path)
			i++
		}
	})
}

// BenchmarkRouterCache 测试路由缓存性能
func BenchmarkRouterCache(b *testing.B) {
	tree := NewRouterTree()
	handler := func(ctx context.Context, c *core.RequestContext) {}
	
	// 添加一些路由
	paths := []string{
		"/users/:id",
		"/users/:id/posts",
		"/users/:id/posts/:post_id",
		"/api/v1/users",
		"/api/v1/posts",
	}
	
	for _, path := range paths {
		tree.AddRoute("GET", path, handler)
	}
	tree.Compile()
	
	testPaths := []string{
		"/users/123",
		"/users/456/posts",
		"/users/789/posts/101",
		"/api/v1/users",
		"/api/v1/posts",
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := testPaths[i%len(testPaths)]
			_, _ = tree.GetRoute("GET", path)
			i++
		}
	})
}

// BenchmarkContextPool 测试Context池性能
func BenchmarkContextPool(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx := mvccontext.NewContext(nil)
				ctx.Set("key", "value")
				ctx.Release()
			}
		})
	})
	
	b.Run("WithoutPool", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx := &mvccontext.EnhancedContext{
					Keys: make(map[string]interface{}),
				}
				ctx.Set("key", "value")
			}
		})
	})
}

// BenchmarkFastEngine 测试完整引擎性能
func BenchmarkFastEngine(b *testing.B) {
	engine := NewFastEngine()
	
	// 添加测试路由
	engine.GET("/", func(ctx context.Context, c *core.RequestContext) {
		// 简化测试，不调用c的方法
	})
	engine.GET("/users/:id", func(ctx context.Context, c *core.RequestContext) {
		// 简化测试
	})
	engine.POST("/users", func(ctx context.Context, c *core.RequestContext) {
		// 简化测试
	})
	
	engine.Compile()
	
	// 模拟请求
	requests := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/users/123"},
		{"POST", "/users"},
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := requests[i%len(requests)]
			// 只测试路由查找，不调用处理器
			_, _ = engine.router.GetRoute(req.method, req.path)
			i++
		}
	})
}

// generateTestPaths 生成测试路径
func generateTestPaths(count int) []string {
	paths := make([]string, count)
	
	for i := 0; i < count; i++ {
		// 生成唯一路径
		paths[i] = fmt.Sprintf("/test/%d", i)
	}
	
	return paths
}

// TestRouterTree_BasicFunctionality 测试路由树基本功能
func TestRouterTree_BasicFunctionality(t *testing.T) {
	tree := NewRouterTree()
	handler := func(ctx context.Context, c *core.RequestContext) {}
	
	// 测试添加路由
	tree.AddRoute("GET", "/users", handler)
	tree.AddRoute("GET", "/users/:id", handler)
	tree.AddRoute("POST", "/users", handler)
	
	// 测试路由查找
	h, params := tree.GetRoute("GET", "/users")
	if h == nil {
		t.Error("Expected handler for /users")
	}
	
	h, params = tree.GetRoute("GET", "/users/123")
	if h == nil {
		t.Error("Expected handler for /users/:id")
	}
	if len(params) != 1 || params.ByName("id") != "123" {
		t.Error("Expected param id=123")
	}
	
	h, params = tree.GetRoute("POST", "/users")
	if h == nil {
		t.Error("Expected handler for POST /users")
	}
}

// TestRouterCache_Functionality 测试路由缓存功能
func TestRouterCache_Functionality(t *testing.T) {
	cache := NewRouterCache(2)
	
	handler := func(ctx context.Context, c *core.RequestContext) {}
	params := Params{{Key: "id", Value: "123"}}
	
	entry := &CacheEntry{
		handler: handler,
		params:  params,
	}
	
	// 测试设置和获取
	cache.Set("GET:/users/123", entry)
	retrieved := cache.Get("GET:/users/123")
	
	if retrieved == nil {
		t.Error("Expected cached entry")
	}
	
	if retrieved.params.ByName("id") != "123" {
		t.Error("Expected param id=123")
	}
}

// TestContextPool_Functionality 测试Context池功能
func TestContextPool_Functionality(t *testing.T) {
	ctx1 := mvccontext.NewContext(nil)
	ctx1.Set("test", "value1")
	
	// 释放到池中
	ctx1.Release()
	
	// 从池中获取新的Context
	ctx2 := mvccontext.NewContext(nil)
	
	// 验证Context被正确重置
	if val, exists := ctx2.Get("test"); exists {
		t.Errorf("Expected context to be reset, got %v", val)
	}
	
	ctx2.Release()
}

// LoadTest 负载测试
func TestFastEngine_LoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	engine := NewFastEngine()
	
	// 添加大量路由
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/test/%d", i)
		engine.GET(path, func(ctx context.Context, c *core.RequestContext) {
			// 简化测试，不调用c的方法
		})
	}
	
	engine.Compile()
	
	// 并发测试
	concurrency := 100
	requestsPerGoroutine := 1000
	
	start := time.Now()
	
	// 启动多个goroutine
	done := make(chan bool, concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			for j := 0; j < requestsPerGoroutine; j++ {
				path := fmt.Sprintf("/test/%d", j%1000)
				// 只测试路由查找
				_, _ = engine.router.GetRoute("GET", path)
			}
			done <- true
		}(i)
	}
	
	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}
	
	elapsed := time.Since(start)
	totalRequests := concurrency * requestsPerGoroutine
	qps := float64(totalRequests) / elapsed.Seconds()
	
	t.Logf("Load test completed:")
	t.Logf("Total requests: %d", totalRequests)
	t.Logf("Time elapsed: %v", elapsed)
	t.Logf("QPS: %.2f", qps)
	
	// 打印引擎统计
	engine.PrintStats()
}

// 内存泄漏测试
func TestMemoryLeak(t *testing.T) {
	engine := NewFastEngine()
	
	engine.GET("/test", func(ctx context.Context, c *core.RequestContext) {
		// 简化测试
	})
	
	// 执行大量请求
	for i := 0; i < 10000; i++ {
		_, _ = engine.router.GetRoute("GET", "/test")
	}
	
	// 检查池大小是否正常
	poolSize := mvccontext.GetCurrentPoolSize()
	if poolSize > 1000 {
		t.Errorf("Pool size too large: %d", poolSize)
	}
	
	// 检查统计信息
	stats := engine.GetStats()
	if stats.ActiveRequests > 0 {
		t.Errorf("Active requests should be 0, got %d", stats.ActiveRequests)
	}
}