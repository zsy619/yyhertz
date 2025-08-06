package middleware

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// MiddlewareFunc 统一的中间件函数类型
type MiddlewareFunc func(*mvccontext.EnhancedContext)

// MiddlewareLayer 中间件层级枚举
type MiddlewareLayer int

const (
	LayerGlobal     MiddlewareLayer = iota // 全局中间件
	LayerGroup                             // 路由组中间件
	LayerRoute                             // 路由级中间件
	LayerController                        // 控制器中间件
)

// MiddlewareInfo 中间件信息
type MiddlewareInfo struct {
	Name       string           // 中间件名称
	Handler    MiddlewareFunc   // 处理函数
	Layer      MiddlewareLayer  // 所属层级
	Priority   int              // 优先级 (数字越小优先级越高)
	Enabled    bool             // 是否启用
	Statistics *MiddlewareStats // 统计信息
}

// MiddlewareStats 中间件统计信息
type MiddlewareStats struct {
	ExecutionCount  int64         // 执行次数
	TotalDuration   time.Duration // 总执行时间
	AverageDuration time.Duration // 平均执行时间
	LastExecution   time.Time     // 最后执行时间
	ErrorCount      int64         // 错误次数
}

// MiddlewarePipeline 中间件管道
type MiddlewarePipeline struct {
	// 分层中间件存储
	layers map[MiddlewareLayer][]*MiddlewareInfo

	// 编译后的中间件链
	compiledChains map[string]CompiledChain

	// 配置选项
	config PipelineConfig

	// 同步控制
	mu       sync.RWMutex
	compiled bool

	// 统计信息
	stats PipelineStats
}

// CompiledChain 编译后的中间件链
type CompiledChain struct {
	ID          string           // 链ID
	Middlewares []MiddlewareFunc // 编译后的中间件函数
	Layer       string           // 应用层级
	CreatedAt   time.Time        // 创建时间
	Stats       *ChainStats      // 链统计信息
}

// ChainStats 中间件链统计
type ChainStats struct {
	ExecutionCount  int64         // 执行次数
	TotalDuration   time.Duration // 总执行时间
	AverageDuration time.Duration // 平均执行时间
	LastExecution   time.Time     // 最后执行时间
}

// PipelineConfig 管道配置
type PipelineConfig struct {
	EnableStatistics   bool          // 启用统计
	EnableCompilation  bool          // 启用编译优化
	MaxCachedChains    int           // 最大缓存链数量
	StatisticsInterval time.Duration // 统计间隔
	EnableDebug        bool          // 启用调试模式
}

// PipelineStats 管道总体统计
type PipelineStats struct {
	RegisteredCount int64 // 注册的中间件数量
	CompiledCount   int64 // 编译的链数量
	ExecutionCount  int64 // 总执行次数
	ErrorCount      int64 // 总错误次数
}

// NewMiddlewarePipeline 创建新的中间件管道
func NewMiddlewarePipeline() *MiddlewarePipeline {
	return &MiddlewarePipeline{
		layers:         make(map[MiddlewareLayer][]*MiddlewareInfo),
		compiledChains: make(map[string]CompiledChain),
		config:         DefaultPipelineConfig(),
		compiled:       false,
	}
}

// DefaultPipelineConfig 默认管道配置
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		EnableStatistics:   true,
		EnableCompilation:  true,
		MaxCachedChains:    1000,
		StatisticsInterval: time.Minute,
		EnableDebug:        false,
	}
}

// Use 注册中间件到指定层级
func (p *MiddlewarePipeline) Use(layer MiddlewareLayer, name string, handler MiddlewareFunc, priority int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	info := &MiddlewareInfo{
		Name:       name,
		Handler:    handler,
		Layer:      layer,
		Priority:   priority,
		Enabled:    true,
		Statistics: &MiddlewareStats{},
	}

	// 添加到对应层级
	p.layers[layer] = append(p.layers[layer], info)

	// 按优先级排序
	p.sortLayerByPriority(layer)

	// 标记需要重新编译
	p.compiled = false

	// 更新统计
	atomic.AddInt64(&p.stats.RegisteredCount, 1)
}

// UseGlobal 注册全局中间件
func (p *MiddlewarePipeline) UseGlobal(name string, handler MiddlewareFunc, priority int) {
	p.Use(LayerGlobal, name, handler, priority)
}

// UseGroup 注册路由组中间件
func (p *MiddlewarePipeline) UseGroup(name string, handler MiddlewareFunc, priority int) {
	p.Use(LayerGroup, name, handler, priority)
}

// UseRoute 注册路由中间件
func (p *MiddlewarePipeline) UseRoute(name string, handler MiddlewareFunc, priority int) {
	p.Use(LayerRoute, name, handler, priority)
}

// UseController 注册控制器中间件
func (p *MiddlewarePipeline) UseController(name string, handler MiddlewareFunc, priority int) {
	p.Use(LayerController, name, handler, priority)
}

// sortLayerByPriority 按优先级对层级内的中间件排序
func (p *MiddlewarePipeline) sortLayerByPriority(layer MiddlewareLayer) {
	middlewares := p.layers[layer]
	if len(middlewares) <= 1 {
		return
	}

	// 简单的冒泡排序，按优先级升序
	for i := 0; i < len(middlewares)-1; i++ {
		for j := 0; j < len(middlewares)-1-i; j++ {
			if middlewares[j].Priority > middlewares[j+1].Priority {
				middlewares[j], middlewares[j+1] = middlewares[j+1], middlewares[j]
			}
		}
	}
}

// BuildChain 构建指定层级组合的中间件链
func (p *MiddlewarePipeline) BuildChain(layers ...MiddlewareLayer) []MiddlewareFunc {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var chain []MiddlewareFunc

	// 按层级优先级组合中间件
	layerOrder := []MiddlewareLayer{LayerGlobal, LayerGroup, LayerRoute, LayerController}

	for _, targetLayer := range layerOrder {
		// 检查是否需要包含此层级
		shouldInclude := false
		for _, requestedLayer := range layers {
			if requestedLayer == targetLayer {
				shouldInclude = true
				break
			}
		}

		if !shouldInclude {
			continue
		}

		// 添加该层级的中间件
		middlewares := p.layers[targetLayer]
		for _, middleware := range middlewares {
			if middleware.Enabled {
				// 如果启用统计，包装中间件函数
				if p.config.EnableStatistics {
					chain = append(chain, p.wrapWithStats(middleware))
				} else {
					chain = append(chain, middleware.Handler)
				}
			}
		}
	}

	return chain
}

// wrapWithStats 用统计信息包装中间件
func (p *MiddlewarePipeline) wrapWithStats(info *MiddlewareInfo) MiddlewareFunc {
	return func(ctx *mvccontext.EnhancedContext) {
		start := time.Now()

		// 执行原始中间件
		defer func() {
			duration := time.Since(start)

			// 更新统计信息
			atomic.AddInt64(&info.Statistics.ExecutionCount, 1)
			info.Statistics.TotalDuration += duration
			info.Statistics.AverageDuration = info.Statistics.TotalDuration /
				time.Duration(atomic.LoadInt64(&info.Statistics.ExecutionCount))
			info.Statistics.LastExecution = time.Now()

			// 更新管道统计
			atomic.AddInt64(&p.stats.ExecutionCount, 1)

			// 检查错误
			if len(ctx.GetErrors()) > 0 {
				atomic.AddInt64(&info.Statistics.ErrorCount, 1)
				atomic.AddInt64(&p.stats.ErrorCount, 1)
			}
		}()

		info.Handler(ctx)
	}
}

// GetCompiledChain 获取编译后的中间件链
func (p *MiddlewarePipeline) GetCompiledChain(chainID string, layers ...MiddlewareLayer) CompiledChain {
	p.mu.RLock()
	// 检查是否已有编译的链
	if chain, exists := p.compiledChains[chainID]; exists {
		p.mu.RUnlock()
		return chain
	}
	p.mu.RUnlock()

	// 需要编译新链
	p.mu.Lock()
	defer p.mu.Unlock()

	// 双重检查锁定模式
	if chain, exists := p.compiledChains[chainID]; exists {
		return chain
	}

	// 构建新链
	middlewares := p.BuildChain(layers...)

	chain := CompiledChain{
		ID:          chainID,
		Middlewares: middlewares,
		Layer:       fmt.Sprintf("%v", layers),
		CreatedAt:   time.Now(),
		Stats:       &ChainStats{},
	}

	// 检查缓存限制
	if len(p.compiledChains) >= p.config.MaxCachedChains {
		// 简单的LRU：删除最旧的链
		var oldestID string
		var oldestTime time.Time = time.Now()

		for id, cachedChain := range p.compiledChains {
			if cachedChain.CreatedAt.Before(oldestTime) {
				oldestTime = cachedChain.CreatedAt
				oldestID = id
			}
		}

		if oldestID != "" {
			delete(p.compiledChains, oldestID)
		}
	}

	// 缓存新链
	p.compiledChains[chainID] = chain
	atomic.AddInt64(&p.stats.CompiledCount, 1)

	return chain
}

// ExecuteChain 执行中间件链
func (p *MiddlewarePipeline) ExecuteChain(ctx *mvccontext.EnhancedContext, chain CompiledChain) {
	start := time.Now()

	defer func() {
		// 更新链统计
		if p.config.EnableStatistics {
			duration := time.Since(start)
			atomic.AddInt64(&chain.Stats.ExecutionCount, 1)
			chain.Stats.TotalDuration += duration
			chain.Stats.AverageDuration = chain.Stats.TotalDuration /
				time.Duration(atomic.LoadInt64(&chain.Stats.ExecutionCount))
			chain.Stats.LastExecution = time.Now()
		}
	}()

	// 转换并设置中间件链
	handlers := convertMiddlewaresToHandlers(chain.Middlewares)
	ctx.SetHandlers(handlers)

	// 开始执行
	ctx.Next()
}

// convertMiddlewaresToHandlers 将 MiddlewareFunc 转换为 HandlerFunc
func convertMiddlewaresToHandlers(middlewares []MiddlewareFunc) []mvccontext.HandlerFunc {
	handlers := make([]mvccontext.HandlerFunc, len(middlewares))
	for i, middleware := range middlewares {
		handlers[i] = mvccontext.HandlerFunc(middleware)
	}
	return handlers
}

// GetMiddlewareInfo 获取中间件信息
func (p *MiddlewarePipeline) GetMiddlewareInfo(layer MiddlewareLayer, name string) (*MiddlewareInfo, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	middlewares := p.layers[layer]
	for _, middleware := range middlewares {
		if middleware.Name == name {
			return middleware, true
		}
	}

	return nil, false
}

// EnableMiddleware 启用/禁用中间件
func (p *MiddlewarePipeline) EnableMiddleware(layer MiddlewareLayer, name string, enabled bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	middlewares := p.layers[layer]
	for _, middleware := range middlewares {
		if middleware.Name == name {
			middleware.Enabled = enabled
			p.compiled = false // 标记需要重新编译
			return true
		}
	}

	return false
}

// GetStatistics 获取管道统计信息
func (p *MiddlewarePipeline) GetStatistics() PipelineStats {
	return p.stats
}

// GetLayerStatistics 获取指定层级的统计信息
func (p *MiddlewarePipeline) GetLayerStatistics(layer MiddlewareLayer) []MiddlewareStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var stats []MiddlewareStats
	middlewares := p.layers[layer]

	for _, middleware := range middlewares {
		if middleware.Statistics != nil {
			stats = append(stats, *middleware.Statistics)
		}
	}

	return stats
}

// Reset 重置管道（清空所有中间件和缓存）
func (p *MiddlewarePipeline) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.layers = make(map[MiddlewareLayer][]*MiddlewareInfo)
	p.compiledChains = make(map[string]CompiledChain)
	p.compiled = false
	p.stats = PipelineStats{}
}

// PrintDebugInfo 打印调试信息
func (p *MiddlewarePipeline) PrintDebugInfo() {
	if !p.config.EnableDebug {
		return
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	fmt.Println("=== Middleware Pipeline Debug Info ===")
	fmt.Printf("Registered Middlewares: %d\n", p.stats.RegisteredCount)
	fmt.Printf("Compiled Chains: %d\n", p.stats.CompiledCount)
	fmt.Printf("Total Executions: %d\n", p.stats.ExecutionCount)
	fmt.Printf("Total Errors: %d\n", p.stats.ErrorCount)

	for layer, middlewares := range p.layers {
		fmt.Printf("\nLayer %d (%s):\n", layer, getLayerName(layer))
		for i, middleware := range middlewares {
			fmt.Printf("  %d. %s (Priority: %d, Enabled: %v)\n",
				i+1, middleware.Name, middleware.Priority, middleware.Enabled)
			if middleware.Statistics != nil {
				fmt.Printf("     Executions: %d, Errors: %d, Avg Duration: %v\n",
					middleware.Statistics.ExecutionCount,
					middleware.Statistics.ErrorCount,
					middleware.Statistics.AverageDuration)
			}
		}
	}
}

// getLayerName 获取层级名称
func getLayerName(layer MiddlewareLayer) string {
	switch layer {
	case LayerGlobal:
		return "Global"
	case LayerGroup:
		return "Group"
	case LayerRoute:
		return "Route"
	case LayerController:
		return "Controller"
	default:
		return "Unknown"
	}
}

// 全局默认管道实例
var defaultPipeline = NewMiddlewarePipeline()

// GetDefaultPipeline 获取默认管道实例
func GetDefaultPipeline() *MiddlewarePipeline {
	return defaultPipeline
}

// UseGlobal 使用默认管道注册全局中间件
func UseGlobal(name string, handler MiddlewareFunc, priority int) {
	defaultPipeline.UseGlobal(name, handler, priority)
}

// UseGroup 使用默认管道注册路由组中间件
func UseGroup(name string, handler MiddlewareFunc, priority int) {
	defaultPipeline.UseGroup(name, handler, priority)
}

// UseRoute 使用默认管道注册路由中间件
func UseRoute(name string, handler MiddlewareFunc, priority int) {
	defaultPipeline.UseRoute(name, handler, priority)
}

// UseController 使用默认管道注册控制器中间件
func UseController(name string, handler MiddlewareFunc, priority int) {
	defaultPipeline.UseController(name, handler, priority)
}
