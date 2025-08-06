package middleware

import (
	"fmt"
	"sync"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// MiddlewareManager 中间件管理器
type MiddlewareManager struct {
	pipeline    *MiddlewarePipeline
	compiler    *MiddlewareCompiler
	registry    *MiddlewareRegistry
	config      ManagerConfig
	mu          sync.RWMutex
	initialized bool
}

// MiddlewareRegistry 中间件注册表
type MiddlewareRegistry struct {
	builtins    map[string]BuiltinMiddlewareFactory // 内置中间件工厂
	customs     map[string]MiddlewareFunc           // 自定义中间件
	metadata    map[string]MiddlewareMetadata       // 中间件元数据
	mu          sync.RWMutex
}

// BuiltinMiddlewareFactory 内置中间件工厂函数
type BuiltinMiddlewareFactory func(config interface{}) MiddlewareFunc

// MiddlewareMetadata 中间件元数据
type MiddlewareMetadata struct {
	Name         string      // 名称
	Version      string      // 版本
	Description  string      // 描述
	Author       string      // 作者
	Dependencies []string    // 依赖
	Config       interface{} // 配置
	IsBuiltin    bool        // 是否内置
	CreatedAt    time.Time   // 创建时间
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	EnableAutoCompile     bool          // 启用自动编译
	EnableHotReload       bool          // 启用热重载
	CompileInterval       time.Duration // 编译间隔
	EnableHealthCheck     bool          // 启用健康检查
	HealthCheckInterval   time.Duration // 健康检查间隔
	EnablePerformanceMonitor bool       // 启用性能监控
}

// NewMiddlewareManager 创建中间件管理器
func NewMiddlewareManager() *MiddlewareManager {
	pipeline := NewMiddlewarePipeline()
	compiler := NewMiddlewareCompiler(pipeline)
	registry := NewMiddlewareRegistry()
	
	manager := &MiddlewareManager{
		pipeline: pipeline,
		compiler: compiler,
		registry: registry,
		config:   DefaultManagerConfig(),
	}
	
	// 注册内置中间件
	manager.registerBuiltinMiddlewares()
	
	return manager
}

// NewMiddlewareRegistry 创建中间件注册表
func NewMiddlewareRegistry() *MiddlewareRegistry {
	return &MiddlewareRegistry{
		builtins: make(map[string]BuiltinMiddlewareFactory),
		customs:  make(map[string]MiddlewareFunc),
		metadata: make(map[string]MiddlewareMetadata),
	}
}

// DefaultManagerConfig 默认管理器配置
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		EnableAutoCompile:        true,
		EnableHotReload:         false,
		CompileInterval:         5 * time.Minute,
		EnableHealthCheck:       true,
		HealthCheckInterval:     time.Minute,
		EnablePerformanceMonitor: true,
	}
}

// Initialize 初始化管理器
func (m *MiddlewareManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.initialized {
		return nil
	}
	
	// 启动自动编译
	if m.config.EnableAutoCompile {
		go m.autoCompileLoop()
	}
	
	// 启动健康检查
	if m.config.EnableHealthCheck {
		go m.healthCheckLoop()
	}
	
	m.initialized = true
	return nil
}

// RegisterBuiltin 注册内置中间件
func (m *MiddlewareManager) RegisterBuiltin(name string, factory BuiltinMiddlewareFactory, metadata MiddlewareMetadata) error {
	m.registry.mu.Lock()
	defer m.registry.mu.Unlock()
	
	if _, exists := m.registry.builtins[name]; exists {
		return fmt.Errorf("builtin middleware %s already registered", name)
	}
	
	metadata.IsBuiltin = true
	metadata.CreatedAt = time.Now()
	
	m.registry.builtins[name] = factory
	m.registry.metadata[name] = metadata
	
	return nil
}

// RegisterCustom 注册自定义中间件
func (m *MiddlewareManager) RegisterCustom(name string, handler MiddlewareFunc, metadata MiddlewareMetadata) error {
	m.registry.mu.Lock()
	defer m.registry.mu.Unlock()
	
	if _, exists := m.registry.customs[name]; exists {
		return fmt.Errorf("custom middleware %s already registered", name)
	}
	
	metadata.IsBuiltin = false
	metadata.CreatedAt = time.Now()
	
	m.registry.customs[name] = handler
	m.registry.metadata[name] = metadata
	
	return nil
}

// UseBuiltin 使用内置中间件
func (m *MiddlewareManager) UseBuiltin(layer MiddlewareLayer, name string, config interface{}, priority int) error {
	m.registry.mu.RLock()
	factory, exists := m.registry.builtins[name]
	m.registry.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("builtin middleware %s not found", name)
	}
	
	// 通过工厂创建中间件实例
	handler := factory(config)
	
	// 注册到管道
	m.pipeline.Use(layer, name, handler, priority)
	
	return nil
}

// UseCustom 使用自定义中间件
func (m *MiddlewareManager) UseCustom(layer MiddlewareLayer, name string, priority int) error {
	m.registry.mu.RLock()
	handler, exists := m.registry.customs[name]
	m.registry.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("custom middleware %s not found", name)
	}
	
	// 注册到管道
	m.pipeline.Use(layer, name, handler, priority)
	
	return nil
}

// GetCompiledChain 获取编译后的中间件链
func (m *MiddlewareManager) GetCompiledChain(layers ...MiddlewareLayer) (*CompiledMiddleware, error) {
	chainID := m.compiler.GenerateChainID(layers...)
	return m.compiler.CompileChain(chainID, layers...)
}

// ExecuteCompiledChain 执行编译后的中间件链
func (m *MiddlewareManager) ExecuteCompiledChain(ctx *mvccontext.EnhancedContext, layers ...MiddlewareLayer) error {
	compiled, err := m.GetCompiledChain(layers...)
	if err != nil {
		return fmt.Errorf("failed to compile middleware chain: %v", err)
	}
	
	// 执行编译后的处理器
	compiled.Handler(ctx)
	
	return nil
}

// GetMiddlewareInfo 获取中间件信息
func (m *MiddlewareManager) GetMiddlewareInfo(name string) (MiddlewareMetadata, bool) {
	m.registry.mu.RLock()
	defer m.registry.mu.RUnlock()
	
	metadata, exists := m.registry.metadata[name]
	return metadata, exists
}

// ListMiddlewares 列出所有注册的中间件
func (m *MiddlewareManager) ListMiddlewares() []MiddlewareMetadata {
	m.registry.mu.RLock()
	defer m.registry.mu.RUnlock()
	
	var list []MiddlewareMetadata
	for _, metadata := range m.registry.metadata {
		list = append(list, metadata)
	}
	
	return list
}

// EnableMiddleware 启用中间件
func (m *MiddlewareManager) EnableMiddleware(layer MiddlewareLayer, name string) bool {
	return m.pipeline.EnableMiddleware(layer, name, true)
}

// DisableMiddleware 禁用中间件
func (m *MiddlewareManager) DisableMiddleware(layer MiddlewareLayer, name string) bool {
	return m.pipeline.EnableMiddleware(layer, name, false)
}

// GetStatistics 获取统计信息
func (m *MiddlewareManager) GetStatistics() ManagerStatistics {
	pipelineStats := m.pipeline.GetStatistics()
	compilerStats := m.compiler.GetStatistics()
	
	return ManagerStatistics{
		Pipeline: pipelineStats,
		Compiler: compilerStats,
		Registry: m.getRegistryStatistics(),
	}
}

// ManagerStatistics 管理器统计信息
type ManagerStatistics struct {
	Pipeline PipelineStats
	Compiler CompilerStats
	Registry RegistryStatistics
}

// RegistryStatistics 注册表统计信息
type RegistryStatistics struct {
	BuiltinCount int
	CustomCount  int
	TotalCount   int
}

// getRegistryStatistics 获取注册表统计
func (m *MiddlewareManager) getRegistryStatistics() RegistryStatistics {
	m.registry.mu.RLock()
	defer m.registry.mu.RUnlock()
	
	return RegistryStatistics{
		BuiltinCount: len(m.registry.builtins),
		CustomCount:  len(m.registry.customs),
		TotalCount:   len(m.registry.metadata),
	}
}

// autoCompileLoop 自动编译循环
func (m *MiddlewareManager) autoCompileLoop() {
	ticker := time.NewTicker(m.config.CompileInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		// 检查是否需要重新编译
		if !m.pipeline.compiled {
			// 编译常用的中间件链组合
			commonChains := [][]MiddlewareLayer{
				{LayerGlobal},
				{LayerGlobal, LayerGroup},
				{LayerGlobal, LayerGroup, LayerRoute},
				{LayerGlobal, LayerGroup, LayerRoute, LayerController},
			}
			
			for _, layers := range commonChains {
				_, err := m.GetCompiledChain(layers...)
				if err != nil {
					// 记录编译错误，但不中断
					continue
				}
			}
		}
	}
}

// healthCheckLoop 健康检查循环
func (m *MiddlewareManager) healthCheckLoop() {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		// 检查中间件健康状态
		m.performHealthCheck()
	}
}

// performHealthCheck 执行健康检查
func (m *MiddlewareManager) performHealthCheck() {
	// 检查编译器缓存大小
	compilerStats := m.compiler.GetStatistics()
	if compilerStats.CompileCount > 0 {
		hitRate := float64(compilerStats.CacheHitCount) / float64(compilerStats.CacheHitCount + compilerStats.CacheMissCount)
		if hitRate < 0.5 { // 命中率低于50%
			// 可以触发缓存优化或告警
		}
	}
	
	// 检查管道统计
	pipelineStats := m.pipeline.GetStatistics()
	if pipelineStats.ExecutionCount > 0 {
		errorRate := float64(pipelineStats.ErrorCount) / float64(pipelineStats.ExecutionCount)
		if errorRate > 0.1 { // 错误率超过10%
			// 可以触发告警
		}
	}
}

// registerBuiltinMiddlewares 注册内置中间件
func (m *MiddlewareManager) registerBuiltinMiddlewares() {
	// Logger中间件
	m.RegisterBuiltin("logger", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.EnhancedContext) {
			start := time.Now()
			path := string(ctx.Request.Path())
			method := string(ctx.Request.Method())
			
			ctx.Next()
			
			// 记录日志
			duration := time.Since(start)
			status := ctx.Writer.Status()
			
			// 这里可以集成实际的日志系统
			if status >= 400 {
				fmt.Printf("[ERROR] %s %s - %d (%v)\n", method, path, status, duration)
			} else {
				fmt.Printf("[INFO] %s %s - %d (%v)\n", method, path, status, duration)
			}
		}
	}, MiddlewareMetadata{
		Name:        "logger",
		Version:     "1.0.0",
		Description: "HTTP请求日志记录中间件",
		Author:      "YYHertz Team",
	})
	
	// Recovery中间件
	m.RegisterBuiltin("recovery", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.EnhancedContext) {
			defer func() {
				if err := recover(); err != nil {
					// 记录panic信息
					fmt.Printf("[PANIC] %v\n", err)
					
					// 返回错误响应
					ctx.JSON(500, map[string]interface{}{
						"error": "Internal Server Error",
						"code":  500,
					})
					ctx.Abort()
				}
			}()
			
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "recovery",
		Version:     "1.0.0",
		Description: "Panic恢复中间件",
		Author:      "YYHertz Team",
	})
	
	// CORS中间件
	m.RegisterBuiltin("cors", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.EnhancedContext) {
			// 设置CORS头
			ctx.Request.Header("Access-Control-Allow-Origin", "*")
			ctx.Request.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			ctx.Request.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
			
			// 处理预检请求
			if string(ctx.Request.Method()) == "OPTIONS" {
				ctx.JSON(204, nil)
				ctx.Abort()
				return
			}
			
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "cors",
		Version:     "1.0.0",
		Description: "跨域资源共享中间件",
		Author:      "YYHertz Team",
	})
	
	// Auth中间件
	m.RegisterBuiltin("auth", func(config interface{}) MiddlewareFunc {
		return func(ctx *mvccontext.EnhancedContext) {
			// 简化的认证逻辑
			token := ctx.Header("Authorization")
			if token == "" {
				ctx.JSON(401, map[string]interface{}{
					"error": "Authorization required",
					"code":  401,
				})
				ctx.Abort()
				return
			}
			
			// 这里可以集成实际的认证逻辑
			ctx.Set("user_id", "authenticated_user")
			ctx.Next()
		}
	}, MiddlewareMetadata{
		Name:        "auth",
		Version:     "1.0.0",
		Description: "用户认证中间件",
		Author:      "YYHertz Team",
		Dependencies: []string{"cors"},
	})
}

// PrintManagerInfo 打印管理器信息
func (m *MiddlewareManager) PrintManagerInfo() {
	stats := m.GetStatistics()
	
	fmt.Println("=== Middleware Manager Info ===")
	fmt.Printf("Registered Middlewares: %d (Builtin: %d, Custom: %d)\n", 
		stats.Registry.TotalCount, stats.Registry.BuiltinCount, stats.Registry.CustomCount)
	fmt.Printf("Pipeline Executions: %d (Errors: %d)\n", 
		stats.Pipeline.ExecutionCount, stats.Pipeline.ErrorCount)
	fmt.Printf("Compiler Cache Hit Rate: %.2f%%\n", 
		float64(stats.Compiler.CacheHitCount)/float64(stats.Compiler.CacheHitCount+stats.Compiler.CacheMissCount)*100)
	
	fmt.Println("\nRegistered Middlewares:")
	middlewares := m.ListMiddlewares()
	for i, middleware := range middlewares {
		fmt.Printf("  %d. %s v%s (%s)\n", 
			i+1, middleware.Name, middleware.Version, getMiddlewareType(middleware.IsBuiltin))
		if middleware.Description != "" {
			fmt.Printf("     %s\n", middleware.Description)
		}
	}
}

// getMiddlewareType 获取中间件类型描述
func getMiddlewareType(isBuiltin bool) string {
	if isBuiltin {
		return "Builtin"
	}
	return "Custom"
}

// 全局管理器实例
var defaultManager = NewMiddlewareManager()

// GetDefaultManager 获取默认管理器
func GetDefaultManager() *MiddlewareManager {
	return defaultManager
}

// InitializeManager 初始化默认管理器
func InitializeManager() error {
	return defaultManager.Initialize()
}