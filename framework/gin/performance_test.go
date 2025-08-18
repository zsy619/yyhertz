// Package gin - 性能测试和基准对比
package gin

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
	
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
)

// BenchmarkRadixTreeRouting 测试Radix Tree路由性能
func BenchmarkRadixTreeRouting(b *testing.B) {
	router := NewRouterEngine()
	
	// 添加多个路由
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/users"},
		{"GET", "/users/:id"},
		{"GET", "/users/:id/posts"},
		{"GET", "/users/:id/posts/:pid"},
		{"POST", "/users"},
		{"PUT", "/users/:id"},
		{"DELETE", "/users/:id"},
		{"GET", "/api/v1/users"},
		{"GET", "/api/v1/users/:id"},
		{"GET", "/api/v2/users"},
		{"GET", "/static/*filepath"},
	}
	
	handlers := make([]HandlerFunc, 1)
	handlers[0] = func(c *Context) {
		c.String(200, "OK")
	}
	
	for _, route := range routes {
		router.addRoute(route.method, route.path, handlers)
	}
	
	// 创建测试Context
	ctx := &Context{
		RequestContext: &app.RequestContext{},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			router.handleHTTPRequest(ctx)
		}
	})
}

// BenchmarkContextPool 测试Context对象池性能
func BenchmarkContextPool(b *testing.B) {
	pool := NewContextPool()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := pool.Get()
			// 模拟使用Context
			ctx.Set("key", "value")
			pool.Put(ctx)
		}
	})
}

// BenchmarkMiddlewareChain 测试中间件链性能
func BenchmarkMiddlewareChain(b *testing.B) {
	// 创建多个中间件
	middlewares := make([]HandlerFunc, 5)
	for i := range middlewares {
		middlewares[i] = func(c *Context) {
			c.Set(fmt.Sprintf("middleware_%d", i), "executed")
			c.Next()
		}
	}
	
	chain := NewMiddlewareChain(middlewares)
	chain.Compile()
	
	ctx := &Context{
		Keys: make(map[string]any),
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			chain.Execute(ctx)
		}
	})
}

// BenchmarkParamParsing 测试参数解析性能
func BenchmarkParamParsing(b *testing.B) {
	router := NewRouterEngine()
	
	handler := []HandlerFunc{
		func(c *Context) {
			id := c.Param("id")
			name := c.Query("name")
			_ = id + name
		},
	}
	
	router.addRoute("GET", "/users/:id", handler)
	
	// 模拟请求
	req := &protocol.Request{}
	req.SetRequestURI("/users/123?name=john")
	req.Header.SetMethod("GET")
	
	ctx := &Context{
		RequestContext: &app.RequestContext{},
	}
	// 设置请求信息，避免直接复制
	ctx.RequestContext.Request.SetRequestURI("/users/123?name=john")
	ctx.RequestContext.Request.Header.SetMethod("GET")
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			router.handleHTTPRequest(ctx)
		}
	})
}

// BenchmarkJSONRendering 测试JSON渲染性能
func BenchmarkJSONRendering(b *testing.B) {
	ctx := &Context{
		RequestContext: &app.RequestContext{},
	}
	
	data := H{
		"message": "Hello World",
		"status":  200,
		"data": H{
			"users": []H{
				{"id": 1, "name": "John"},
				{"id": 2, "name": "Jane"},
			},
		},
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx.JSON(200, data)
		}
	})
}

// BenchmarkConcurrentRouting 测试并发路由性能
func BenchmarkConcurrentRouting(b *testing.B) {
	router := NewRouterEngine()
	
	// 添加大量路由
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/api/v1/resource%d/:id", i)
		router.addRoute("GET", path, []HandlerFunc{
			func(c *Context) {
				c.String(200, "OK")
			},
		})
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := &Context{
				RequestContext: &app.RequestContext{},
			}
			router.handleHTTPRequest(ctx)
		}
	})
}

// PerformanceTest 性能测试结构
type PerformanceTest struct {
	Name        string
	RequestPath string
	Method      string
	Setup       func() *Engine
	Iterations  int
}

// RunPerformanceTests 运行性能测试套件
func RunPerformanceTests() {
	tests := []PerformanceTest{
		{
			Name:        "Simple Route",
			RequestPath: "/ping",
			Method:      "GET",
			Setup: func() *Engine {
				r := New()
				r.GET("/ping", func(c *Context) {
					c.String(200, "pong")
				})
				return r
			},
			Iterations: 100000,
		},
		{
			Name:        "Parameterized Route",
			RequestPath: "/users/123",
			Method:      "GET",
			Setup: func() *Engine {
				r := New()
				r.GET("/users/:id", func(c *Context) {
					id := c.Param("id")
					c.JSON(200, H{"user_id": id})
				})
				return r
			},
			Iterations: 100000,
		},
		{
			Name:        "Complex Route with Middleware",
			RequestPath: "/api/v1/users/123/posts/456",
			Method:      "GET",
			Setup: func() *Engine {
				r := New()
				r.Use(func(c *Context) {
					c.Header("X-Custom", "middleware")
					c.Next()
				})
				r.GET("/api/v1/users/:uid/posts/:pid", func(c *Context) {
					uid := c.Param("uid")
					pid := c.Param("pid")
					c.JSON(200, H{"uid": uid, "pid": pid})
				})
				return r
			},
			Iterations: 50000,
		},
	}
	
	fmt.Println("=== YYHertz-Gin Performance Test Results ===")
	
	for _, test := range tests {
		result := runSinglePerformanceTest(test)
		printPerformanceResult(test.Name, result)
	}
}

// PerformanceResult 性能测试结果
type PerformanceResult struct {
	TotalDuration    time.Duration
	AverageDuration  time.Duration
	RequestsPerSec   float64
	MemoryAllocated  uint64
	GCCycles         uint64
}

// runSinglePerformanceTest 运行单个性能测试
func runSinglePerformanceTest(test PerformanceTest) PerformanceResult {
	engine := test.Setup()
	
	// 预热
	for i := 0; i < 1000; i++ {
		runSingleRequest(engine, test.RequestPath, test.Method)
	}
	
	// 记录GC统计
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	start := time.Now()
	
	// 并发执行测试
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	requestsPerWorker := test.Iterations / numWorkers
	
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				runSingleRequest(engine, test.RequestPath, test.Method)
			}
		}()
	}
	
	wg.Wait()
	totalDuration := time.Since(start)
	
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	return PerformanceResult{
		TotalDuration:    totalDuration,
		AverageDuration:  totalDuration / time.Duration(test.Iterations),
		RequestsPerSec:   float64(test.Iterations) / totalDuration.Seconds(),
		MemoryAllocated:  m2.TotalAlloc - m1.TotalAlloc,
		GCCycles:         uint64(m2.NumGC - m1.NumGC),
	}
}

// runSingleRequest 执行单个请求
func runSingleRequest(engine *Engine, path, method string) {
	// 创建模拟请求
	req := &protocol.Request{}
	req.SetRequestURI(path)
	req.Header.SetMethod(method)
	
	ctx := &app.RequestContext{}
	// 设置请求信息，避免直接复制
	ctx.Request.SetRequestURI(path)
	ctx.Request.Header.SetMethod(method)
	
	// 执行请求
	ginCtx := engine.createContext(ctx, nil)
	if engine.router != nil {
		engine.router.handleHTTPRequest(ginCtx)
	}
}

// printPerformanceResult 打印性能测试结果
func printPerformanceResult(name string, result PerformanceResult) {
	fmt.Printf("\n--- %s ---\n", name)
	fmt.Printf("Total Duration: %v\n", result.TotalDuration)
	fmt.Printf("Average Duration: %v\n", result.AverageDuration)
	fmt.Printf("Requests/sec: %.2f\n", result.RequestsPerSec)
	fmt.Printf("Memory Allocated: %d bytes\n", result.MemoryAllocated)
	fmt.Printf("GC Cycles: %d\n", result.GCCycles)
}

// MemoryLeakTest 内存泄漏测试
func MemoryLeakTest() {
	fmt.Println("=== Memory Leak Test ===")
	
	engine := New()
	engine.GET("/test", func(c *Context) {
		c.JSON(200, H{"message": "test"})
	})
	
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// 执行大量请求
	for i := 0; i < 100000; i++ {
		runSingleRequest(engine, "/test", "GET")
		
		if i%10000 == 0 {
			runtime.GC()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Printf("Iteration %d: Alloc=%d KB, Sys=%d KB, NumGC=%d\n",
				i, m.Alloc/1024, m.Sys/1024, m.NumGC)
		}
	}
	
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	fmt.Printf("Memory before: %d KB\n", m1.Alloc/1024)
	fmt.Printf("Memory after: %d KB\n", m2.Alloc/1024)
	fmt.Printf("Memory growth: %d KB\n", (m2.Alloc-m1.Alloc)/1024)
	
	if m2.Alloc-m1.Alloc > 1024*1024 { // 1MB增长
		fmt.Println("⚠️  Potential memory leak detected!")
	} else {
		fmt.Println("✅ No significant memory leak detected")
	}
}

// LoadTest 负载测试
func LoadTest(duration time.Duration, concurrency int) {
	fmt.Printf("=== Load Test (Duration: %v, Concurrency: %d) ===\n", duration, concurrency)
	
	engine := New()
	engine.GET("/api/users/:id", func(c *Context) {
		id := c.Param("id")
		c.JSON(200, H{
			"user_id": id,
			"name":    "Test User",
			"email":   "test@example.com",
		})
	})
	
	var (
		totalRequests uint64
		successCount  uint64
		errorCount    uint64
		mu            sync.Mutex
	)
	
	start := time.Now()
	var wg sync.WaitGroup
	
	// 启动多个goroutine并发执行请求
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for time.Since(start) < duration {
				path := fmt.Sprintf("/api/users/%d", workerID%1000+1)
				
				func() {
					defer func() {
						if r := recover(); r != nil {
							mu.Lock()
							errorCount++
							mu.Unlock()
						}
					}()
					
					runSingleRequest(engine, path, "GET")
					
					mu.Lock()
					successCount++
					mu.Unlock()
				}()
				
				mu.Lock()
				totalRequests++
				mu.Unlock()
			}
		}(i)
	}
	
	wg.Wait()
	elapsed := time.Since(start)
	
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful Requests: %d\n", successCount)
	fmt.Printf("Failed Requests: %d\n", errorCount)
	fmt.Printf("Requests/sec: %.2f\n", float64(totalRequests)/elapsed.Seconds())
	fmt.Printf("Success Rate: %.2f%%\n", float64(successCount)/float64(totalRequests)*100)
}

// ComparisonTest 与原版Gin的性能对比测试
func ComparisonTest() {
	fmt.Println("=== Performance Comparison ===")
	
	// 测试相同的路由配置
	testCases := []struct {
		name string
		path string
	}{
		{"Simple Route", "/ping"},
		{"One Parameter", "/users/:id"},
		{"Two Parameters", "/users/:id/posts/:pid"},
		{"Wildcard", "/static/*filepath"},
	}
	
	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.name)
		
		// YYHertz-Gin测试
		yyHertzResult := benchmarkYYHertzGin(tc.path, 50000)
		fmt.Printf("YYHertz-Gin: %.2f req/sec, %v avg latency\n",
			yyHertzResult.RequestsPerSec, yyHertzResult.AverageDuration)
		
		// 这里可以添加与原版Gin的对比
		// ginResult := benchmarkOriginalGin(tc.path, 50000)
		// fmt.Printf("Original Gin: %.2f req/sec, %v avg latency\n",
		//     ginResult.RequestsPerSec, ginResult.AverageDuration)
		
		// improvement := (yyHertzResult.RequestsPerSec - ginResult.RequestsPerSec) / ginResult.RequestsPerSec * 100
		// fmt.Printf("Performance improvement: %.2f%%\n", improvement)
	}
}

// benchmarkYYHertzGin YYHertz-Gin基准测试
func benchmarkYYHertzGin(path string, iterations int) PerformanceResult {
	engine := New()
	
	// 根据路径添加相应的路由
	switch {
	case path == "/ping":
		engine.GET("/ping", func(c *Context) {
			c.String(200, "pong")
		})
	case path == "/users/:id":
		engine.GET("/users/:id", func(c *Context) {
			id := c.Param("id")
			c.JSON(200, H{"id": id})
		})
	case path == "/users/:id/posts/:pid":
		engine.GET("/users/:id/posts/:pid", func(c *Context) {
			id := c.Param("id")
			pid := c.Param("pid")
			c.JSON(200, H{"id": id, "pid": pid})
		})
	case path == "/static/*filepath":
		engine.GET("/static/*filepath", func(c *Context) {
			filepath := c.Param("filepath")
			c.String(200, "File: %s", filepath)
		})
	}
	
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		testPath := path
		if path == "/users/:id" {
			testPath = "/users/123"
		} else if path == "/users/:id/posts/:pid" {
			testPath = "/users/123/posts/456"
		} else if path == "/static/*filepath" {
			testPath = "/static/js/app.js"
		}
		
		runSingleRequest(engine, testPath, "GET")
	}
	
	duration := time.Since(start)
	runtime.ReadMemStats(&m2)
	
	return PerformanceResult{
		TotalDuration:   duration,
		AverageDuration: duration / time.Duration(iterations),
		RequestsPerSec:  float64(iterations) / duration.Seconds(),
		MemoryAllocated: m2.TotalAlloc - m1.TotalAlloc,
		GCCycles:        uint64(m2.NumGC - m1.NumGC),
	}
}

// TestPerformanceRegression 性能回归测试
func TestPerformanceRegression(t *testing.T) {
	// 基准性能指标
	const (
		minRequestsPerSec = 10000  // 最小请求每秒
		maxAvgLatency     = 100    // 最大平均延迟(微秒)
		maxMemoryPerReq   = 1024   // 每请求最大内存(字节)
	)
	
	engine := New()
	engine.GET("/benchmark", func(c *Context) {
		c.JSON(200, H{"status": "ok"})
	})
	
	result := benchmarkYYHertzGin("/benchmark", 10000)
	
	if result.RequestsPerSec < minRequestsPerSec {
		t.Errorf("Performance regression: requests/sec %.2f < threshold %d",
			result.RequestsPerSec, minRequestsPerSec)
	}
	
	avgLatencyMicros := result.AverageDuration.Microseconds()
	if avgLatencyMicros > maxAvgLatency {
		t.Errorf("Performance regression: avg latency %d μs > threshold %d μs",
			avgLatencyMicros, maxAvgLatency)
	}
	
	memoryPerReq := result.MemoryAllocated / 10000
	if memoryPerReq > maxMemoryPerReq {
		t.Errorf("Performance regression: memory/req %d bytes > threshold %d bytes",
			memoryPerReq, maxMemoryPerReq)
	}
	
	t.Logf("Performance test passed: %.2f req/sec, %v avg latency, %d bytes/req",
		result.RequestsPerSec, result.AverageDuration, memoryPerReq)
}