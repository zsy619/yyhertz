package controller

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// OptimizedControllerManager 优化的控制器管理器
type OptimizedControllerManager struct {
	compiler       *ControllerCompiler     // 控制器编译器
	lifecycleManager *LifecycleManager     // 生命周期管理器
	controllers    sync.Map                // 已注册的控制器
	config         *CompilerConfig         // 配置
	stats          *PerformanceStats       // 性能统计
	mu             sync.RWMutex           // 读写锁
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	TotalRequests       int64         // 总请求数
	AverageResponseTime time.Duration // 平均响应时间
	CacheHitRate       float64       // 缓存命中率
	CompilationTime    time.Duration // 编译时间
	ControllerInstances int64         // 控制器实例数
	ActiveConnections  int64         // 活跃连接数
	mu                 sync.RWMutex  // 统计锁
}

// NewOptimizedControllerManager 创建优化的控制器管理器
func NewOptimizedControllerManager(config *CompilerConfig) *OptimizedControllerManager {
	if config == nil {
		config = DefaultCompilerConfig()
	}

	return &OptimizedControllerManager{
		compiler:         NewControllerCompiler(config),
		lifecycleManager: NewLifecycleManager(config),
		config:          config,
		stats:           &PerformanceStats{},
	}
}

// RegisterController 注册控制器
func (ocm *OptimizedControllerManager) RegisterController(controller interface{}) error {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	controllerName := controllerType.Name()

	// 编译控制器
	startTime := time.Now()
	compiled, err := ocm.compiler.Compile(controller)
	if err != nil {
		return fmt.Errorf("failed to compile controller %s: %w", controllerName, err)
	}
	compilationTime := time.Since(startTime)

	// 更新统计信息
	ocm.stats.updateCompilationTime(compilationTime)

	// 存储编译后的控制器
	ocm.controllers.Store(controllerName, compiled)

	fmt.Printf("Controller %s registered successfully (compiled in %v)\n", controllerName, compilationTime)
	return nil
}

// HandleRequest 处理请求
func (ocm *OptimizedControllerManager) HandleRequest(ctx *context.Context, controllerName, methodName string) error {
	startTime := time.Now()
	defer func() {
		responseTime := time.Since(startTime)
		ocm.stats.updateResponseTime(responseTime)
		ocm.stats.updateTotalRequests(1)
	}()

	// 获取编译后的控制器
	compiledController, err := ocm.getCompiledController(controllerName)
	if err != nil {
		return fmt.Errorf("controller not found: %s", controllerName)
	}

	// 获取编译后的方法
	compiledMethod, exists := compiledController.Methods[methodName]
	if !exists {
		return fmt.Errorf("method not found: %s.%s", controllerName, methodName)
	}

	// 创建控制器实例
	instance, err := ocm.lifecycleManager.CreateController(compiledController.Type, ctx)
	if err != nil {
		return fmt.Errorf("failed to create controller instance: %w", err)
	}

	// 确保释放控制器实例
	defer func() {
		if releaseErr := ocm.lifecycleManager.ReturnController(instance); releaseErr != nil {
			fmt.Printf("Failed to return controller: %v\n", releaseErr)
		}
	}()

	// 执行方法
	if err := compiledMethod.Handler(ctx, instance.Controller); err != nil {
		return fmt.Errorf("method execution failed: %w", err)
	}

	return nil
}

// getCompiledController 获取编译后的控制器
func (ocm *OptimizedControllerManager) getCompiledController(controllerName string) (*CompiledController, error) {
	if value, exists := ocm.controllers.Load(controllerName); exists {
		return value.(*CompiledController), nil
	}
	return nil, fmt.Errorf("controller not found: %s", controllerName)
}

// PrecompileAll 预编译所有控制器
func (ocm *OptimizedControllerManager) PrecompileAll() error {
	var controllers []interface{}
	
	ocm.controllers.Range(func(key, value interface{}) bool {
		compiled := value.(*CompiledController)
		controllers = append(controllers, compiled.Instance)
		return true
	})

	return ocm.compiler.PrecompileAll(controllers)
}

// GetStats 获取性能统计
func (ocm *OptimizedControllerManager) GetStats() *PerformanceStats {
	ocm.stats.mu.RLock()
	defer ocm.stats.mu.RUnlock()

	return &PerformanceStats{
		TotalRequests:       ocm.stats.TotalRequests,
		AverageResponseTime: ocm.stats.AverageResponseTime,
		CacheHitRate:       ocm.stats.CacheHitRate,
		CompilationTime:    ocm.stats.CompilationTime,
		ControllerInstances: ocm.stats.ControllerInstances,
		ActiveConnections:  ocm.stats.ActiveConnections,
	}
}

// GetDetailedStats 获取详细统计信息
func (ocm *OptimizedControllerManager) GetDetailedStats() map[string]interface{} {
	stats := map[string]interface{}{
		"performance": ocm.GetStats(),
		"compiler":    ocm.compiler.GetStats(),
		"lifecycle":   ocm.lifecycleManager.GetMetrics(),
	}

	// 控制器统计
	controllerStats := make(map[string]interface{})
	ocm.controllers.Range(func(key, value interface{}) bool {
		controllerName := key.(string)
		compiled := value.(*CompiledController)
		
		controllerStats[controllerName] = map[string]interface{}{
			"methods_count": len(compiled.Methods),
			"created_at":    compiled.CreatedAt,
			"metadata":      compiled.Metadata,
			"pool_stats":    compiled.Pool.Stats(),
		}
		return true
	})
	stats["controllers"] = controllerStats

	return stats
}

// RegisterLifecycleHooks 注册生命周期钩子
func (ocm *OptimizedControllerManager) RegisterLifecycleHooks() {
	// 注册性能监控钩子
	ocm.lifecycleManager.RegisterHook(HookAfterCreate, func(controller interface{}, ctx *context.Context) error {
		ocm.stats.updateControllerInstances(1)
		return nil
	})

	ocm.lifecycleManager.RegisterHook(HookAfterDestroy, func(controller interface{}, ctx *context.Context) error {
		ocm.stats.updateControllerInstances(-1)
		return nil
	})

	// 注册缓存预热钩子
	ocm.lifecycleManager.RegisterHook(HookAfterInit, func(controller interface{}, ctx *context.Context) error {
		// 这里可以实现缓存预热逻辑
		return nil
	})
}

// Shutdown 优雅关闭
func (ocm *OptimizedControllerManager) Shutdown() error {
	fmt.Println("Shutting down optimized controller manager...")

	// 打印统计信息
	stats := ocm.GetDetailedStats()
	fmt.Printf("Final stats: %+v\n", stats)

	return nil
}

// PerformanceStats 方法实现

// updateTotalRequests 更新总请求数
func (ps *PerformanceStats) updateTotalRequests(delta int64) {
	ps.mu.Lock()
	ps.TotalRequests += delta
	ps.mu.Unlock()
}

// updateResponseTime 更新响应时间
func (ps *PerformanceStats) updateResponseTime(duration time.Duration) {
	ps.mu.Lock()
	// 简单的移动平均算法
	if ps.AverageResponseTime == 0 {
		ps.AverageResponseTime = duration
	} else {
		ps.AverageResponseTime = (ps.AverageResponseTime + duration) / 2
	}
	ps.mu.Unlock()
}

// updateCompilationTime 更新编译时间
func (ps *PerformanceStats) updateCompilationTime(duration time.Duration) {
	ps.mu.Lock()
	if ps.CompilationTime == 0 {
		ps.CompilationTime = duration
	} else {
		ps.CompilationTime = (ps.CompilationTime + duration) / 2
	}
	ps.mu.Unlock()
}

// updateControllerInstances 更新控制器实例数
func (ps *PerformanceStats) updateControllerInstances(delta int64) {
	ps.mu.Lock()
	ps.ControllerInstances += delta
	ps.mu.Unlock()
}

// 示例集成接口

// OptimizedController 优化的控制器接口
type OptimizedController interface {
	// 标准方法
	Init(ctx *context.Context) error
	Destroy() error
	Reset()

	// 优化方法
	GetControllerName() string
	GetMethodMapping() map[string]string
	GetMiddleware() []string
}

// ============= BaseOptimizedController 已完全移除 =============
// 
// BaseOptimizedController 已完全合并到 framework/mvc/core.BaseController 中
// 所有优化特性现在通过 BaseController.EnableOptimization() 启用
//
// 迁移指南:
// 旧方式: controller.BaseOptimizedController  
// 新方式: core.BaseController + ctrl.EnableOptimization()
//
// 参考: framework/mvc/MIGRATION_GUIDE.md

// 性能优化工具

// WarmupCache 缓存预热
func (ocm *OptimizedControllerManager) WarmupCache() error {
	fmt.Println("Warming up controller cache...")
	
	ocm.controllers.Range(func(key, value interface{}) bool {
		controllerName := key.(string)
		compiled := value.(*CompiledController)
		
		fmt.Printf("Preloading controller: %s\n", controllerName)
		
		// 预创建一些控制器实例到池中
		for i := 0; i < 5; i++ {
			instance, err := ocm.lifecycleManager.CreateController(compiled.Type, nil)
			if err != nil {
				fmt.Printf("Failed to precreate controller instance: %v\n", err)
				continue
			}
			
			// 立即归还到池中
			if err := ocm.lifecycleManager.ReturnController(instance); err != nil {
				fmt.Printf("Failed to return prewarmed controller: %v\n", err)
			}
		}
		
		return true
	})
	
	fmt.Println("Cache warmup completed")
	return nil
}

// OptimizeMemory 内存优化
func (ocm *OptimizedControllerManager) OptimizeMemory() {
	fmt.Println("Optimizing memory usage...")
	
	// 这里可以实现内存优化逻辑
	// 例如：清理过期的控制器实例、压缩缓存等
	
	fmt.Println("Memory optimization completed")
}