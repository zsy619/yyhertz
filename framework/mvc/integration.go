// Package mvc MVC框架集成文件
// 将优化后的中间件管道和错误处理系统与FastEngine整合
package mvc

import (
	"fmt"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/engine"
	"github.com/zsy619/yyhertz/framework/mvc/errors"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
)

// EnhancedFastEngine 增强的FastEngine
// 集成了优化的中间件管道和智能错误处理
type EnhancedFastEngine struct {
	*engine.FastEngine                               // 嵌入原始引擎
	middlewareManager  *middleware.MiddlewareManager // 中间件管理器
	errorDispatcher    *errors.ErrorDispatcher       // 错误分发器
	errorClassifier    *errors.IntelligentClassifier // 错误分类器
	autoRecovery       *errors.AutoRecovery          // 自动恢复系统
	config             *config.MVCConfig             // MVC配置
}

// NewEnhancedFastEngine 创建增强的FastEngine
func NewEnhancedFastEngine(mvcConfig ...*config.MVCConfig) *EnhancedFastEngine {
	// 使用默认配置或提供的配置
	var cfg *config.MVCConfig
	if len(mvcConfig) > 0 && mvcConfig[0] != nil {
		cfg = mvcConfig[0]
	} else {
		// 尝试从配置文件加载
		if loadedConfig, err := config.GetMVCConfig(); err == nil {
			cfg = loadedConfig
		} else {
			// 使用默认配置
			defaultConfig := config.GetDefaultMVCConfig()
			cfg = &defaultConfig
		}
	}

	// 创建基础引擎
	baseEngine := engine.NewFastEngine()

	// 创建增强组件
	middlewareManager := middleware.NewMiddlewareManager()
	errorDispatcher := errors.NewErrorDispatcher()
	errorClassifier := errors.NewIntelligentClassifier()
	autoRecovery := errors.NewAutoRecovery(errorClassifier)

	enhanced := &EnhancedFastEngine{
		FastEngine:        baseEngine,
		middlewareManager: middlewareManager,
		errorDispatcher:   errorDispatcher,
		errorClassifier:   errorClassifier,
		autoRecovery:      autoRecovery,
		config:            cfg,
	}

	// 初始化增强功能
	enhanced.initialize()

	return enhanced
}

// initialize 初始化增强功能
func (e *EnhancedFastEngine) initialize() {
	// 初始化中间件管理器
	if e.config.Middleware.EnableOptimization {
		if err := e.middlewareManager.Initialize(); err != nil {
			fmt.Printf("Failed to initialize middleware manager: %v\n", err)
		}

		// 预编译常用中间件链
		if e.config.Middleware.PrecompileChains {
			e.precompileCommonChains()
		}
	}

	// 设置全局错误处理
	if e.config.ErrorHandling.EnableIntelligent {
		e.setupGlobalErrorHandling()
	}

	// 启动统计报告
	if e.config.Performance.EnableStatistics && e.config.Performance.StatsReportInterval > 0 {
		go e.statsReportLoop()
	}

	// 打印系统信息
	if e.config.Debug.EnableMode {
		e.printSystemInfo()
	}
}

// precompileCommonChains 预编译常用中间件链
func (e *EnhancedFastEngine) precompileCommonChains() {
	commonChains := [][]middleware.MiddlewareLayer{
		{middleware.LayerGlobal},
		{middleware.LayerGlobal, middleware.LayerGroup},
		{middleware.LayerGlobal, middleware.LayerGroup, middleware.LayerRoute},
		{middleware.LayerGlobal, middleware.LayerGroup, middleware.LayerRoute, middleware.LayerController},
	}

	for _, layers := range commonChains {
		_, err := e.middlewareManager.GetCompiledChain(layers...)
		if err != nil {
			fmt.Printf("Failed to precompile chain %v: %v\n", layers, err)
		}
	}

	fmt.Println("✓ Middleware chains precompiled successfully")
}

// setupGlobalErrorHandling 设置全局错误处理
func (e *EnhancedFastEngine) setupGlobalErrorHandling() {
	// 注册智能错误处理中间件
	e.Use("intelligent-error-handler", e.createIntelligentErrorHandler(), 999) // 最低优先级，最后执行

	fmt.Println("✓ Intelligent error handling enabled")
}

// createIntelligentErrorHandler 创建智能错误处理中间件
func (e *EnhancedFastEngine) createIntelligentErrorHandler() middleware.MiddlewareFunc {
	return func(ctx *context.Context) {
		// 执行后续处理
		ctx.Next()

		// 检查是否有错误
		if len(ctx.GetErrors()) > 0 {
			lastError := ctx.GetErrors()[len(ctx.GetErrors())-1]

			// 使用错误分类器分类错误
			var classification *errors.ErrorClassification
			if e.config.ErrorHandling.EnableClassification {
				classification = e.errorClassifier.Classify(lastError, ctx)
			}

			// 尝试自动恢复
			var recoveryResult *errors.RecoveryResult
			if e.config.ErrorHandling.EnableAutoRecovery && classification != nil && classification.Retryable {
				recoveryResult = e.autoRecovery.Recover(ctx, lastError)

				// 如果恢复成功，清除错误
				if recoveryResult.Success {
					ctx.ClearErrors()
					return
				}
			}

			// 使用错误分发器处理错误
			err := e.errorDispatcher.Dispatch(ctx, lastError)
			if err != nil {
				// 如果分发器也无法处理，使用默认处理
				ctx.JSON(500, map[string]interface{}{
					"code":    500,
					"message": "Internal Server Error",
					"success": false,
				})
			}
		}
	}
}

// statsReportLoop 统计报告循环
func (e *EnhancedFastEngine) statsReportLoop() {
	ticker := time.NewTicker(e.config.Performance.StatsReportInterval)
	defer ticker.Stop()

	for range ticker.C {
		if e.config.Debug.PrintMiddleware {
			e.middlewareManager.PrintManagerInfo()
		}

		if e.config.Debug.PrintError {
			errors.PrintErrorHandlerInfo()
			errors.PrintClassifierInfo()
			errors.PrintRecoveryInfo()
		}
	}
}

// printSystemInfo 打印系统信息
func (e *EnhancedFastEngine) printSystemInfo() {
	fmt.Println("=== Enhanced FastEngine System Info ===")

	// 中间件系统信息
	if e.config.Middleware.EnableOptimization {
		fmt.Println("✓ Middleware optimization enabled")
		stats := e.middlewareManager.GetStatistics()
		fmt.Printf("  - Registered middlewares: %d\n", stats.Registry.TotalCount)
		fmt.Printf("  - Pipeline executions: %d\n", stats.Pipeline.ExecutionCount)
		fmt.Printf("  - Compiler cache hits: %d\n", stats.Compiler.CacheHitCount)
	}

	// 错误处理系统信息
	if e.config.ErrorHandling.EnableIntelligent {
		fmt.Println("✓ Intelligent error handling enabled")

		if e.config.ErrorHandling.EnableClassification {
			fmt.Println("  - Error classification enabled")
		}

		if e.config.ErrorHandling.EnableAutoRecovery {
			fmt.Println("  - Auto recovery enabled")
		}
	}

	// 性能监控信息
	if e.config.Performance.EnableMonitoring {
		fmt.Println("✓ Performance monitoring enabled")
		fmt.Printf("  - Stats report interval: %v\n", e.config.Performance.StatsReportInterval)
	}

	fmt.Println("=====================================")
}

// 增强的中间件注册方法

// Use 注册全局中间件（重写基础方法）
func (e *EnhancedFastEngine) Use(name string, handler middleware.MiddlewareFunc, priority int) *EnhancedFastEngine {
	if e.config.Middleware.EnableOptimization {
		// 使用优化的中间件管道
		e.middlewareManager.RegisterCustom(name, handler, middleware.MiddlewareMetadata{
			Name:        name,
			Description: "Custom middleware registered via Use()",
			Author:      "Application",
		})
		e.middlewareManager.UseCustom(middleware.LayerGlobal, name, priority)
	} else {
		// 使用原始方法（这里需要根据实际的 FastEngine API 调整）
		// e.FastEngine.Use(name, func(ctx *context.Context) {
		// 	handler(ctx)
		// }, priority)
		fmt.Printf("Warning: Original FastEngine Use method not implemented\n")
	}

	return e
}

// UseBuiltin 使用内置中间件
func (e *EnhancedFastEngine) UseBuiltin(name string, config interface{}, priority int) *EnhancedFastEngine {
	if e.config.Middleware.EnableOptimization {
		err := e.middlewareManager.UseBuiltin(middleware.LayerGlobal, name, config, priority)
		if err != nil {
			fmt.Printf("Failed to use builtin middleware %s: %v\n", name, err)
		}
	} else {
		fmt.Printf("Warning: UseBuiltin requires middleware optimization to be enabled\n")
	}

	return e
}

// UseGroup 在路由组级别使用中间件
func (e *EnhancedFastEngine) UseGroup(name string, handler middleware.MiddlewareFunc, priority int) *EnhancedFastEngine {
	if e.config.Middleware.EnableOptimization {
		e.middlewareManager.RegisterCustom(name, handler, middleware.MiddlewareMetadata{
			Name:        name,
			Description: "Group middleware",
			Author:      "Application",
		})
		e.middlewareManager.UseCustom(middleware.LayerGroup, name, priority)
	}

	return e
}

// UseRoute 在路由级别使用中间件
func (e *EnhancedFastEngine) UseRoute(name string, handler middleware.MiddlewareFunc, priority int) *EnhancedFastEngine {
	if e.config.Middleware.EnableOptimization {
		e.middlewareManager.RegisterCustom(name, handler, middleware.MiddlewareMetadata{
			Name:        name,
			Description: "Route middleware",
			Author:      "Application",
		})
		e.middlewareManager.UseCustom(middleware.LayerRoute, name, priority)
	}

	return e
}

// UseController 在控制器级别使用中间件
func (e *EnhancedFastEngine) UseController(name string, handler middleware.MiddlewareFunc, priority int) *EnhancedFastEngine {
	if e.config.Middleware.EnableOptimization {
		e.middlewareManager.RegisterCustom(name, handler, middleware.MiddlewareMetadata{
			Name:        name,
			Description: "Controller middleware",
			Author:      "Application",
		})
		e.middlewareManager.UseCustom(middleware.LayerController, name, priority)
	}

	return e
}

// 增强的错误处理方法

// RegisterErrorHandler 注册错误处理器
func (e *EnhancedFastEngine) RegisterErrorHandler(handler errors.ErrorHandler) *EnhancedFastEngine {
	if e.config.ErrorHandling.EnableIntelligent {
		e.errorDispatcher.RegisterHandler(handler)
	}
	return e
}

// RegisterErrorHandlerFunc 注册错误处理函数
func (e *EnhancedFastEngine) RegisterErrorHandlerFunc(name string, priority int, canHandle func(error) bool, handleFunc errors.ErrorHandlerFunc) *EnhancedFastEngine {
	if e.config.ErrorHandling.EnableIntelligent {
		e.errorDispatcher.RegisterHandlerFunc(name, priority, canHandle, handleFunc)
	}
	return e
}

// LearnError 学习错误分类（用于提高分类准确性）
func (e *EnhancedFastEngine) LearnError(err error, category errors.ErrorCategory, severity errors.ErrorSeverity) {
	if e.config.ErrorHandling.EnableClassification {
		e.errorClassifier.Learn(err, category, severity)
	}
}

// AddRecoveryStrategy 添加恢复策略
func (e *EnhancedFastEngine) AddRecoveryStrategy(strategy errors.RecoveryStrategy) *EnhancedFastEngine {
	if e.config.ErrorHandling.EnableAutoRecovery {
		e.autoRecovery.AddStrategy(strategy)
	}
	return e
}

// 系统管理方法

// GetMiddlewareManager 获取中间件管理器
func (e *EnhancedFastEngine) GetMiddlewareManager() *middleware.MiddlewareManager {
	return e.middlewareManager
}

// GetErrorDispatcher 获取错误分发器
func (e *EnhancedFastEngine) GetErrorDispatcher() *errors.ErrorDispatcher {
	return e.errorDispatcher
}

// GetErrorClassifier 获取错误分类器
func (e *EnhancedFastEngine) GetErrorClassifier() *errors.IntelligentClassifier {
	return e.errorClassifier
}

// GetAutoRecovery 获取自动恢复系统
func (e *EnhancedFastEngine) GetAutoRecovery() *errors.AutoRecovery {
	return e.autoRecovery
}

// GetSystemStatistics 获取系统统计信息
func (e *EnhancedFastEngine) GetSystemStatistics() SystemStatistics {
	return SystemStatistics{
		Middleware: e.middlewareManager.GetStatistics(),
		Error:      e.errorDispatcher.GetStatistics(),
		Classifier: e.errorClassifier.GetStatistics(),
		Recovery:   e.autoRecovery.GetStatistics(),
	}
}

// SystemStatistics 系统统计信息
type SystemStatistics struct {
	Middleware middleware.ManagerStatistics
	Error      errors.DispatcherStats
	Classifier errors.ClassifierStats
	Recovery   errors.RecoveryStats
}

// EnableDebugMode 启用调试模式
func (e *EnhancedFastEngine) EnableDebugMode() *EnhancedFastEngine {
	e.config.Debug.EnableMode = true
	e.config.Debug.PrintMiddleware = true
	e.config.Debug.PrintError = true
	return e
}

// DisableDebugMode 禁用调试模式
func (e *EnhancedFastEngine) DisableDebugMode() *EnhancedFastEngine {
	e.config.Debug.EnableMode = false
	e.config.Debug.PrintMiddleware = false
	e.config.Debug.PrintError = false
	return e
}

// PrintSystemStatistics 打印系统统计信息
func (e *EnhancedFastEngine) PrintSystemStatistics() {
	fmt.Println("=== Enhanced FastEngine Statistics ===")

	if e.config.Middleware.EnableOptimization {
		e.middlewareManager.PrintManagerInfo()
		fmt.Println()
	}

	if e.config.ErrorHandling.EnableIntelligent {
		errors.PrintErrorHandlerInfo()
		fmt.Println()
	}

	if e.config.ErrorHandling.EnableClassification {
		errors.PrintClassifierInfo()
		fmt.Println()
	}

	if e.config.ErrorHandling.EnableAutoRecovery {
		errors.PrintRecoveryInfo()
		fmt.Println()
	}

	fmt.Println("======================================")
}

// 便捷的创建函数

// New 创建默认的增强FastEngine
func New() *EnhancedFastEngine {
	return NewEnhancedFastEngine()
}

// NewWithConfig 使用配置创建增强FastEngine
func NewWithConfig(config *config.MVCConfig) *EnhancedFastEngine {
	return NewEnhancedFastEngine(config)
}

// NewForProduction 创建生产环境配置的增强FastEngine
func NewForProduction() *EnhancedFastEngine {
	cfg := config.GetProductionConfig()
	return NewEnhancedFastEngine(&cfg)
}

// NewForDevelopment 创建开发环境配置的增强FastEngine
func NewForDevelopment() *EnhancedFastEngine {
	cfg := config.GetDevelopmentConfig()
	return NewEnhancedFastEngine(&cfg)
}
