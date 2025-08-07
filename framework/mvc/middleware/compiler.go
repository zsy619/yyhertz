package middleware

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"sync"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// MiddlewareCompiler 中间件编译器
type MiddlewareCompiler struct {
	pipeline   *MiddlewarePipeline
	cache      map[string]*CompiledMiddleware
	mu         sync.RWMutex
	stats      CompilerStats
	config     CompilerConfig
}

// CompiledMiddleware 编译后的中间件
type CompiledMiddleware struct {
	ID             string                     // 编译ID
	Name           string                     // 名称
	Handler        MiddlewareFunc             // 优化后的处理函数
	Dependencies   []string                   // 依赖关系
	OptimizedChain []MiddlewareFunc           // 优化后的中间件链
	Metadata       CompiledMiddlewareMetadata // 元数据
	CreatedAt      time.Time                  // 创建时间
}

// CompiledMiddlewareMetadata 编译中间件元数据
type CompiledMiddlewareMetadata struct {
	OriginalCount   int           // 原始中间件数量
	OptimizedCount  int           // 优化后数量
	CompileTime     time.Duration // 编译耗时
	OptimizationLevel int         // 优化级别
	CacheHit        bool          // 是否命中缓存
}

// CompilerStats 编译器统计
type CompilerStats struct {
	CompileCount    int64         // 编译次数
	CacheHitCount   int64         // 缓存命中次数
	CacheMissCount  int64         // 缓存未命中次数
	TotalCompileTime time.Duration // 总编译时间
	AverageCompileTime time.Duration // 平均编译时间
}

// CompilerConfig 编译器配置
type CompilerConfig struct {
	EnableOptimization    bool          // 启用优化
	EnableInlining        bool          // 启用内联
	EnableDeadCodeElimination bool      // 启用死代码消除
	OptimizationLevel     int           // 优化级别 (0-3)
	MaxCacheSize          int           // 最大缓存大小
	CacheExpireTime       time.Duration // 缓存过期时间
	EnableDependencyAnalysis bool       // 启用依赖分析
}

// MiddlewareGraph 中间件依赖图
type MiddlewareGraph struct {
	nodes map[string]*MiddlewareNode
	edges map[string][]string // 依赖关系: key依赖于values
}

// MiddlewareNode 中间件节点
type MiddlewareNode struct {
	Info         *MiddlewareInfo
	Dependencies []string // 依赖的中间件名称
	Dependents   []string // 依赖此中间件的名称
	Order        int      // 执行顺序
}

// NewMiddlewareCompiler 创建中间件编译器
func NewMiddlewareCompiler(pipeline *MiddlewarePipeline) *MiddlewareCompiler {
	return &MiddlewareCompiler{
		pipeline: pipeline,
		cache:    make(map[string]*CompiledMiddleware),
		config:   DefaultCompilerConfig(),
	}
}

// DefaultCompilerConfig 默认编译器配置
func DefaultCompilerConfig() CompilerConfig {
	return CompilerConfig{
		EnableOptimization:       true,
		EnableInlining:          true,
		EnableDeadCodeElimination: true,
		OptimizationLevel:       2,
		MaxCacheSize:            500,
		CacheExpireTime:         30 * time.Minute,
		EnableDependencyAnalysis: true,
	}
}

// CompileChain 编译中间件链
func (c *MiddlewareCompiler) CompileChain(chainID string, layers ...MiddlewareLayer) (*CompiledMiddleware, error) {
	start := time.Now()
	
	// 检查缓存
	if compiled := c.getCachedCompiled(chainID); compiled != nil {
		c.stats.CacheHitCount++
		compiled.Metadata.CacheHit = true
		return compiled, nil
	}
	
	c.stats.CacheMissCount++
	c.stats.CompileCount++
	
	// 构建依赖图
	graph := c.buildDependencyGraph(layers...)
	
	// 执行拓扑排序
	sortedMiddlewares, err := c.topologicalSort(graph)
	if err != nil {
		return nil, fmt.Errorf("middleware dependency cycle detected: %v", err)
	}
	
	// 构建原始链
	originalChain := c.buildOriginalChain(sortedMiddlewares)
	
	// 执行优化
	optimizedChain := c.optimizeChain(originalChain)
	
	// 创建编译结果
	compiled := &CompiledMiddleware{
		ID:             chainID,
		Name:           fmt.Sprintf("Compiled-%s", chainID),
		Handler:        c.createCompiledHandler(optimizedChain),
		Dependencies:   c.extractDependencies(graph),
		OptimizedChain: optimizedChain,
		Metadata: CompiledMiddlewareMetadata{
			OriginalCount:     len(originalChain),
			OptimizedCount:    len(optimizedChain),
			CompileTime:       time.Since(start),
			OptimizationLevel: c.config.OptimizationLevel,
			CacheHit:          false,
		},
		CreatedAt: time.Now(),
	}
	
	// 缓存结果
	c.cacheCompiled(chainID, compiled)
	
	// 更新统计
	c.stats.TotalCompileTime += compiled.Metadata.CompileTime
	c.stats.AverageCompileTime = c.stats.TotalCompileTime / time.Duration(c.stats.CompileCount)
	
	return compiled, nil
}

// buildDependencyGraph 构建依赖图
func (c *MiddlewareCompiler) buildDependencyGraph(layers ...MiddlewareLayer) *MiddlewareGraph {
	graph := &MiddlewareGraph{
		nodes: make(map[string]*MiddlewareNode),
		edges: make(map[string][]string),
	}
	
	c.pipeline.mu.RLock()
	defer c.pipeline.mu.RUnlock()
	
	// 按层级顺序添加节点
	order := 0
	for _, layer := range []MiddlewareLayer{LayerGlobal, LayerGroup, LayerRoute, LayerController} {
		// 检查是否需要包含此层级
		shouldInclude := false
		for _, targetLayer := range layers {
			if targetLayer == layer {
				shouldInclude = true
				break
			}
		}
		
		if !shouldInclude {
			continue
		}
		
		middlewares := c.pipeline.layers[layer]
		for _, middleware := range middlewares {
			if !middleware.Enabled {
				continue
			}
			
			node := &MiddlewareNode{
				Info:         middleware,
				Dependencies: c.extractMiddlewareDependencies(middleware),
				Order:        order,
			}
			
			graph.nodes[middleware.Name] = node
			order++
		}
	}
	
	// 构建边（依赖关系）
	for name, node := range graph.nodes {
		for _, dep := range node.Dependencies {
			if _, exists := graph.nodes[dep]; exists {
				graph.edges[name] = append(graph.edges[name], dep)
				// 同时更新被依赖节点的依赖者列表
				if depNode, exists := graph.nodes[dep]; exists {
					depNode.Dependents = append(depNode.Dependents, name)
				}
			}
		}
	}
	
	return graph
}

// extractMiddlewareDependencies 提取中间件依赖
func (c *MiddlewareCompiler) extractMiddlewareDependencies(middleware *MiddlewareInfo) []string {
	// 这里可以根据中间件名称或注解来推断依赖关系
	// 简化实现，基于常见的依赖模式
	dependencies := make([]string, 0)
	
	switch middleware.Name {
	case "auth":
		dependencies = append(dependencies, "cors", "logger")
	case "ratelimit":
		dependencies = append(dependencies, "logger")
	case "recovery":
		// recovery通常在最后执行
	default:
		// 默认依赖logger和recovery
		if middleware.Name != "logger" && middleware.Name != "recovery" {
			dependencies = append(dependencies, "logger")
		}
	}
	
	return dependencies
}

// topologicalSort 拓扑排序
func (c *MiddlewareCompiler) topologicalSort(graph *MiddlewareGraph) ([]*MiddlewareNode, error) {
	var result []*MiddlewareNode
	inDegree := make(map[string]int)
	
	// 计算入度
	for name := range graph.nodes {
		inDegree[name] = len(graph.edges[name])
	}
	
	// 找到入度为0的节点
	var queue []*MiddlewareNode
	for name, node := range graph.nodes {
		if inDegree[name] == 0 {
			queue = append(queue, node)
		}
	}
	
	// 处理队列
	for len(queue) > 0 {
		// 从队列中取出节点
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		
		// 更新依赖此节点的其他节点的入度
		for _, dependent := range current.Dependents {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, graph.nodes[dependent])
			}
		}
	}
	
	// 检查是否有环
	if len(result) != len(graph.nodes) {
		return nil, fmt.Errorf("circular dependency detected")
	}
	
	// 按原始顺序和优先级排序
	sort.Slice(result, func(i, j int) bool {
		// 首先按层级排序
		if result[i].Info.Layer != result[j].Info.Layer {
			return result[i].Info.Layer < result[j].Info.Layer
		}
		// 然后按优先级排序
		return result[i].Info.Priority < result[j].Info.Priority
	})
	
	return result, nil
}

// buildOriginalChain 构建原始中间件链
func (c *MiddlewareCompiler) buildOriginalChain(sortedMiddlewares []*MiddlewareNode) []MiddlewareFunc {
	chain := make([]MiddlewareFunc, 0, len(sortedMiddlewares))
	
	for _, node := range sortedMiddlewares {
		chain = append(chain, node.Info.Handler)
	}
	
	return chain
}

// optimizeChain 优化中间件链
func (c *MiddlewareCompiler) optimizeChain(originalChain []MiddlewareFunc) []MiddlewareFunc {
	if !c.config.EnableOptimization {
		return originalChain
	}
	
	optimizedChain := originalChain
	
	// 死代码消除
	if c.config.EnableDeadCodeElimination {
		optimizedChain = c.eliminateDeadCode(optimizedChain)
	}
	
	// 内联优化
	if c.config.EnableInlining {
		optimizedChain = c.inlineMiddlewares(optimizedChain)
	}
	
	return optimizedChain
}

// eliminateDeadCode 死代码消除
func (c *MiddlewareCompiler) eliminateDeadCode(chain []MiddlewareFunc) []MiddlewareFunc {
	// 简化实现：移除空的或无效的中间件
	optimized := make([]MiddlewareFunc, 0, len(chain))
	
	for _, middleware := range chain {
		if middleware != nil {
			optimized = append(optimized, middleware)
		}
	}
	
	return optimized
}

// inlineMiddlewares 内联中间件
func (c *MiddlewareCompiler) inlineMiddlewares(chain []MiddlewareFunc) []MiddlewareFunc {
	// 简化实现：对于简单的中间件，可以考虑内联
	// 这里保持原链，实际实现可以做更复杂的内联优化
	return chain
}

// createCompiledHandler 创建编译后的处理器
func (c *MiddlewareCompiler) createCompiledHandler(optimizedChain []MiddlewareFunc) MiddlewareFunc {
	return func(ctx *mvccontext.Context) {
		// 转换并设置优化后的中间件链
		handlers := make([]mvccontext.HandlerFunc, len(optimizedChain))
		for i, middleware := range optimizedChain {
			handlers[i] = mvccontext.HandlerFunc(middleware)
		}
		ctx.SetHandlers(handlers)
		
		// 开始执行
		ctx.Next()
	}
}

// extractDependencies 提取依赖列表
func (c *MiddlewareCompiler) extractDependencies(graph *MiddlewareGraph) []string {
	var dependencies []string
	
	for name, node := range graph.nodes {
		if len(node.Dependencies) > 0 {
			dependencies = append(dependencies, fmt.Sprintf("%s -> %v", name, node.Dependencies))
		}
	}
	
	return dependencies
}

// getCachedCompiled 获取缓存的编译结果
func (c *MiddlewareCompiler) getCachedCompiled(chainID string) *CompiledMiddleware {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	compiled, exists := c.cache[chainID]
	if !exists {
		return nil
	}
	
	// 检查是否过期
	if time.Since(compiled.CreatedAt) > c.config.CacheExpireTime {
		return nil
	}
	
	return compiled
}

// cacheCompiled 缓存编译结果
func (c *MiddlewareCompiler) cacheCompiled(chainID string, compiled *CompiledMiddleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 检查缓存大小限制
	if len(c.cache) >= c.config.MaxCacheSize {
		// LRU淘汰：删除最旧的条目
		var oldestID string
		var oldestTime time.Time = time.Now()
		
		for id, cached := range c.cache {
			if cached.CreatedAt.Before(oldestTime) {
				oldestTime = cached.CreatedAt
				oldestID = id
			}
		}
		
		if oldestID != "" {
			delete(c.cache, oldestID)
		}
	}
	
	c.cache[chainID] = compiled
}

// GenerateChainID 生成中间件链ID
func (c *MiddlewareCompiler) GenerateChainID(layers ...MiddlewareLayer) string {
	// 构建唯一标识字符串
	var builder strings.Builder
	for _, layer := range layers {
		builder.WriteString(fmt.Sprintf("%d-", layer))
	}
	
	// 添加配置信息
	builder.WriteString(fmt.Sprintf("opt:%d-inline:%t-dce:%t", 
		c.config.OptimizationLevel, 
		c.config.EnableInlining, 
		c.config.EnableDeadCodeElimination))
	
	// 计算哈希
	hasher := fnv.New64a()
	hasher.Write([]byte(builder.String()))
	
	return fmt.Sprintf("chain_%x", hasher.Sum64())
}

// GetStatistics 获取编译器统计信息
func (c *MiddlewareCompiler) GetStatistics() CompilerStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.stats
}

// ClearCache 清空缓存
func (c *MiddlewareCompiler) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.cache = make(map[string]*CompiledMiddleware)
}

// PrintCompilerInfo 打印编译器信息
func (c *MiddlewareCompiler) PrintCompilerInfo() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	fmt.Println("=== Middleware Compiler Info ===")
	fmt.Printf("Compile Count: %d\n", c.stats.CompileCount)
	fmt.Printf("Cache Hit Rate: %.2f%%\n", 
		float64(c.stats.CacheHitCount)/float64(c.stats.CacheHitCount+c.stats.CacheMissCount)*100)
	fmt.Printf("Average Compile Time: %v\n", c.stats.AverageCompileTime)
	fmt.Printf("Cache Size: %d/%d\n", len(c.cache), c.config.MaxCacheSize)
	
	fmt.Printf("\nConfig:\n")
	fmt.Printf("  Optimization Level: %d\n", c.config.OptimizationLevel)
	fmt.Printf("  Enable Inlining: %v\n", c.config.EnableInlining)
	fmt.Printf("  Enable Dead Code Elimination: %v\n", c.config.EnableDeadCodeElimination)
	fmt.Printf("  Enable Dependency Analysis: %v\n", c.config.EnableDependencyAnalysis)
}

// 全局编译器实例
var defaultCompiler = NewMiddlewareCompiler(GetDefaultPipeline())

// GetDefaultCompiler 获取默认编译器
func GetDefaultCompiler() *MiddlewareCompiler {
	return defaultCompiler
}

// CompileGlobalChain 编译全局中间件链
func CompileGlobalChain() (*CompiledMiddleware, error) {
	chainID := defaultCompiler.GenerateChainID(LayerGlobal)
	return defaultCompiler.CompileChain(chainID, LayerGlobal)
}

// CompileFullChain 编译完整中间件链
func CompileFullChain() (*CompiledMiddleware, error) {
	chainID := defaultCompiler.GenerateChainID(LayerGlobal, LayerGroup, LayerRoute, LayerController)
	return defaultCompiler.CompileChain(chainID, LayerGlobal, LayerGroup, LayerRoute, LayerController)
}