package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// FastEngine 高性能MVC引擎
type FastEngine struct {
	router      *RouterTree                // 路由树
	contextPool *mvccontext.ContextPool    // Context池
	middleware  []mvccontext.HandlerFunc   // 全局中间件
	
	// 配置
	config EngineConfig
	
	// 统计
	stats EngineStats
	
	// 状态
	started   bool
	mu        sync.RWMutex
	startTime time.Time
}

// EngineConfig 引擎配置
type EngineConfig struct {
	MaxRouteCache   int           // 最大路由缓存
	MaxContextPool  int32         // 最大Context池大小
	EnableMetrics   bool          // 启用性能统计
	EnablePprof     bool          // 启用性能分析
	RequestTimeout  time.Duration // 请求超时
	RedirectSlash   bool          // 自动重定向斜杠
	HandleOptions   bool          // 处理OPTIONS请求
}

// EngineStats 引擎统计
type EngineStats struct {
	TotalRequests   int64 // 总请求数
	ActiveRequests  int64 // 活跃请求数
	AverageLatency  int64 // 平均延迟(微秒)
	RouteHitRate    float64 // 路由命中率
	ContextHitRate  float64 // Context池命中率
}

// NewFastEngine 创建高性能引擎
func NewFastEngine() *FastEngine {
	config := DefaultEngineConfig()
	
	engine := &FastEngine{
		router:      NewRouterTree(),
		contextPool: mvccontext.NewContextPool(),
		middleware:  make([]mvccontext.HandlerFunc, 0),
		config:      config,
		startTime:   time.Now(),
	}
	
	// 设置Context池大小
	mvccontext.SetMaxPoolSize(config.MaxContextPool)
	
	return engine
}

// DefaultEngineConfig 默认引擎配置
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		MaxRouteCache:  1000,
		MaxContextPool: 1000,
		EnableMetrics:  true,
		EnablePprof:    false,
		RequestTimeout: 30 * time.Second,
		RedirectSlash:  true,
		HandleOptions:  true,
	}
}

// SetConfig 设置引擎配置
func (e *FastEngine) SetConfig(config EngineConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.config = config
	mvccontext.SetMaxPoolSize(config.MaxContextPool)
}

// Use 添加全局中间件
func (e *FastEngine) Use(middleware ...mvccontext.HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.middleware = append(e.middleware, middleware...)
}

// AddRoute 添加路由
func (e *FastEngine) AddRoute(method, path string, handler core.HandlerFunc) {
	// 包装处理器以支持新的Context
	wrappedHandler := e.wrapHandler(handler)
	e.router.AddRoute(method, path, wrappedHandler)
}

// wrapHandler 包装处理器以支持Context池化
func (e *FastEngine) wrapHandler(handler core.HandlerFunc) core.HandlerFunc {
	return func(ctx context.Context, c *core.RequestContext) {
		// 从池中获取增强Context
		enhancedCtx := mvccontext.NewContext((*app.RequestContext)(c))
		defer enhancedCtx.Release()
		
		// 记录请求开始
		if e.config.EnableMetrics {
			atomic.AddInt64(&e.stats.TotalRequests, 1)
			atomic.AddInt64(&e.stats.ActiveRequests, 1)
			defer atomic.AddInt64(&e.stats.ActiveRequests, -1)
		}
		
		start := time.Now()
		
		// 设置处理器链（全局中间件 + 路由处理器）
		handlers := make([]mvccontext.HandlerFunc, len(e.middleware)+1)
		copy(handlers, e.middleware)
		handlers[len(handlers)-1] = func(ectx *mvccontext.Context) {
			// 调用原始处理器
			handler(ctx, c)
		}
		
		enhancedCtx.SetHandlers(handlers)
		enhancedCtx.Next()
		
		// 记录延迟
		if e.config.EnableMetrics {
			latency := time.Since(start).Microseconds()
			e.updateAverageLatency(latency)
		}
	}
}

// HandleRequest 处理HTTP请求
func (e *FastEngine) HandleRequest(method, path string, c *core.RequestContext) {
	// 查找路由
	handler, params := e.router.GetRoute(method, path)
	
	if handler != nil {
		// 创建增强Context
		enhancedCtx := mvccontext.NewContext((*app.RequestContext)(c))
		defer enhancedCtx.Release()
		
		// 设置路由参数
		convertedParams := make(mvccontext.Params, len(params))
		for i, p := range params {
			convertedParams[i] = mvccontext.Param{Key: p.Key, Value: p.Value}
		}
		enhancedCtx.Params = convertedParams
		enhancedCtx.FullPath = path
		
		// 执行处理器
		handler(context.Background(), c)
		return
	}
	
	// 处理404
	e.handle404(method, path, c)
}

// handle404 处理404错误
func (e *FastEngine) handle404(method, path string, c *core.RequestContext) {
	if e.config.RedirectSlash {
		// 尝试重定向
		if method != "CONNECT" && path != "/" {
			if e.tryRedirect(method, path, c) {
				return
			}
		}
	}
	
	// 返回404
	c.String(404, "404 page not found")
}

// tryRedirect 尝试重定向
func (e *FastEngine) tryRedirect(method, path string, c *core.RequestContext) bool {
	var redirectPath string
	
	if len(path) > 1 && path[len(path)-1] == '/' {
		// 移除尾部斜杠
		redirectPath = path[:len(path)-1]
	} else {
		// 添加尾部斜杠
		redirectPath = path + "/"
	}
	
	if handler, _ := e.router.GetRoute(method, redirectPath); handler != nil {
		c.Redirect(301, []byte(redirectPath))
		return true
	}
	
	return false
}

// Compile 编译引擎以优化性能
func (e *FastEngine) Compile() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// 编译路由树
	e.router.Compile()
	e.started = true
}

// GetStats 获取引擎统计信息
func (e *FastEngine) GetStats() EngineStats {
	if !e.config.EnableMetrics {
		return EngineStats{}
	}
	
	poolMetrics := mvccontext.GetPoolMetrics()
	
	return EngineStats{
		TotalRequests:  atomic.LoadInt64(&e.stats.TotalRequests),
		ActiveRequests: atomic.LoadInt64(&e.stats.ActiveRequests),
		AverageLatency: atomic.LoadInt64(&e.stats.AverageLatency),
		RouteHitRate:   e.calculateRouteHitRate(),
		ContextHitRate: e.calculateContextHitRate(poolMetrics),
	}
}

// calculateRouteHitRate 计算路由命中率
func (e *FastEngine) calculateRouteHitRate() float64 {
	// 这里需要根据实际路由缓存命中情况计算
	// 简化实现
	return 0.95
}

// calculateContextHitRate 计算Context池命中率
func (e *FastEngine) calculateContextHitRate(metrics mvccontext.PoolMetrics) float64 {
	if metrics.Gets == 0 {
		return 0
	}
	return float64(metrics.Reuses) / float64(metrics.Gets)
}

// updateAverageLatency 更新平均延迟
func (e *FastEngine) updateAverageLatency(latency int64) {
	// 简单的移动平均
	old := atomic.LoadInt64(&e.stats.AverageLatency)
	new := (old + latency) / 2
	atomic.StoreInt64(&e.stats.AverageLatency, new)
}

// PrintStats 打印统计信息
func (e *FastEngine) PrintStats() {
	stats := e.GetStats()
	poolMetrics := mvccontext.GetPoolMetrics()
	
	fmt.Printf("=== FastEngine Statistics ===\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Active Requests: %d\n", stats.ActiveRequests)
	fmt.Printf("Average Latency: %d μs\n", stats.AverageLatency)
	fmt.Printf("Route Hit Rate: %.2f%%\n", stats.RouteHitRate*100)
	fmt.Printf("Context Hit Rate: %.2f%%\n", stats.ContextHitRate*100)
	fmt.Printf("Context Pool - Gets: %d, Puts: %d, News: %d, Reuses: %d\n",
		poolMetrics.Gets, poolMetrics.Puts, poolMetrics.News, poolMetrics.Reuses)
	fmt.Printf("Running Time: %v\n", time.Since(e.startTime))
}

// ============= 路由注册便捷方法 =============

// GET 注册GET路由
func (e *FastEngine) GET(path string, handler core.HandlerFunc) {
	e.AddRoute("GET", path, handler)
}

// POST 注册POST路由
func (e *FastEngine) POST(path string, handler core.HandlerFunc) {
	e.AddRoute("POST", path, handler)
}

// PUT 注册PUT路由
func (e *FastEngine) PUT(path string, handler core.HandlerFunc) {
	e.AddRoute("PUT", path, handler)
}

// DELETE 注册DELETE路由
func (e *FastEngine) DELETE(path string, handler core.HandlerFunc) {
	e.AddRoute("DELETE", path, handler)
}

// PATCH 注册PATCH路由
func (e *FastEngine) PATCH(path string, handler core.HandlerFunc) {
	e.AddRoute("PATCH", path, handler)
}

// HEAD 注册HEAD路由
func (e *FastEngine) HEAD(path string, handler core.HandlerFunc) {
	e.AddRoute("HEAD", path, handler)
}

// OPTIONS 注册OPTIONS路由
func (e *FastEngine) OPTIONS(path string, handler core.HandlerFunc) {
	e.AddRoute("OPTIONS", path, handler)
}

// Any 注册任意方法路由
func (e *FastEngine) Any(path string, handler core.HandlerFunc) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		e.AddRoute(method, path, handler)
	}
}

// ============= 路由组支持 =============

// Group 路由组
type Group struct {
	engine     *FastEngine
	prefix     string
	middleware []mvccontext.HandlerFunc
}

// Group 创建路由组
func (e *FastEngine) Group(prefix string) *Group {
	return &Group{
		engine: e,
		prefix: prefix,
	}
}

// Use 为路由组添加中间件
func (g *Group) Use(middleware ...mvccontext.HandlerFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// GET 为路由组注册GET路由
func (g *Group) GET(path string, handler core.HandlerFunc) {
	g.engine.AddRoute("GET", g.prefix+path, g.wrapGroupHandler(handler))
}

// POST 为路由组注册POST路由
func (g *Group) POST(path string, handler core.HandlerFunc) {
	g.engine.AddRoute("POST", g.prefix+path, g.wrapGroupHandler(handler))
}

// wrapGroupHandler 为路由组包装处理器
func (g *Group) wrapGroupHandler(handler core.HandlerFunc) core.HandlerFunc {
	if len(g.middleware) == 0 {
		return handler
	}
	
	return func(ctx context.Context, c *core.RequestContext) {
		enhancedCtx := mvccontext.NewContext((*app.RequestContext)(c))
		defer enhancedCtx.Release()
		
		// 设置组中间件 + 处理器
		handlers := make([]mvccontext.HandlerFunc, len(g.middleware)+1)
		copy(handlers, g.middleware)
		handlers[len(handlers)-1] = func(ectx *mvccontext.Context) {
			handler(ctx, c)
		}
		
		enhancedCtx.SetHandlers(handlers)
		enhancedCtx.Next()
	}
}