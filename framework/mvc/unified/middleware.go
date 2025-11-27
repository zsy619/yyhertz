package unified

import (
	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// MiddlewareAdapter 中间件适配器
//
// 将统一管理器的过滤器系统适配到不同的中间件架构中，
// 支持多种Web框架的中间件接口。
//
// 设计特点：
// - 框架无关：支持多种Web框架
// - 过滤器桥接：将过滤器转换为中间件
// - 错误处理：统一的错误处理机制
// - 性能优化：最小化适配开销
type MiddlewareAdapter struct {
	manager *Manager // 统一管理器实例
}

// NewMiddlewareAdapter 创建新的中间件适配器
//
// 返回：
//   - *MiddlewareAdapter: 中间件适配器实例
//
// 示例：
//
//	adapter := unified.NewMiddlewareAdapter()
//	middleware := adapter.ToHertzMiddleware()
func NewMiddlewareAdapter() *MiddlewareAdapter {
	return &MiddlewareAdapter{
		manager: GetManager(),
	}
}

// ToMiddlewareFunc 转换为中间件函数
//
// 将过滤器系统转换为通用的中间件函数。
// 中间件会按优先级执行所有匹配的过滤器。
//
// 返回：
//   - func(*context.Context, func()): 中间件函数
//
// 示例：
//
//	// 添加SSO过滤器
//	manager.AddFilter(&Filter{
//	    Name: "sso",
//	    Pattern: "/*",
//	    FilterFunc: FilterSSO,
//	    Priority: 10,
//	    Enabled: true,
//	})
//
//	// 转换为中间件
//	adapter := NewMiddlewareAdapter()
//	middleware := adapter.ToMiddlewareFunc()
func (ma *MiddlewareAdapter) ToMiddlewareFunc() func(*context.Context, func()) {
	return func(c *context.Context, next func()) {
		if ma.manager == nil {
			// 如果管理器未初始化，直接继续
			if next != nil {
				next()
			}
			return
		}

		// 获取请求路径
		path := string(c.Request().URI().Path())

		// 执行过滤器链
		result, err := ma.manager.ExecuteFilters(c, path)

		// 处理过滤器执行结果
		switch result {
		case FilterContinue:
			// 继续执行后续中间件和处理器
			if next != nil {
				next()
			}
		case FilterStop:
			// 停止执行，如果有错误则记录
			if err != nil {
				// 可以在这里记录错误日志
				// 或者设置错误响应
				handleFilterError(c, err)
			}
			// 不调用 next()，停止后续处理
		case FilterSkip:
			// 跳过当前请求处理，但继续执行后续中间件
			if next != nil {
				next()
			}
		default:
			// 未知结果，默认继续
			if next != nil {
				next()
			}
		}
	}
}

// ToMiddlewareFuncWithConfig 带配置的中间件转换
//
// 允许自定义中间件行为，如错误处理方式等。
//
// 参数：
//   - config: 中间件配置
//
// 返回：
//   - func(*context.Context, func()): 配置化的中间件函数
//
// 示例：
//
//	config := &MiddlewareConfig{
//	    HandleError: true,
//	    LogErrors: true,
//	    SkipPaths: []string{"/health", "/metrics"},
//	}
//	middleware := adapter.ToMiddlewareFuncWithConfig(config)
func (ma *MiddlewareAdapter) ToMiddlewareFuncWithConfig(config *MiddlewareConfig) func(*context.Context, func()) {
	return func(c *context.Context, next func()) {
		if ma.manager == nil {
			if next != nil {
				next()
			}
			return
		}

		// 获取请求路径
		path := string(c.Request().URI().Path())

		// 检查是否跳过此路径
		if config != nil && config.shouldSkipPath(path) {
			if next != nil {
				next()
			}
			return
		}

		// 执行过滤器链
		result, err := ma.manager.ExecuteFilters(c, path)

		// 处理错误
		if err != nil && config != nil {
			if config.LogErrors {
				// 记录错误（这里可以集成日志系统）
				logFilterError(path, err)
			}

			if config.HandleError {
				handleFilterError(c, err)
				return
			}
		}

		// 处理结果
		switch result {
		case FilterContinue:
			if next != nil {
				next()
			}
		case FilterStop:
			// 不调用 next()
		case FilterSkip:
			if next != nil {
				next()
			}
		default:
			if next != nil {
				next()
			}
		}
	}
}

// ToGenericMiddleware 转换为通用中间件函数
//
// 返回一个通用的中间件函数，可以适配到任何支持类似接口的框架。
//
// 返回：
//   - func(*context.Context) (FilterResult, error): 通用中间件函数
//
// 示例：
//
//	middleware := adapter.ToGenericMiddleware()
//	result, err := middleware(ctx)
func (ma *MiddlewareAdapter) ToGenericMiddleware() func(*context.Context) (FilterResult, error) {
	return func(ctx *context.Context) (FilterResult, error) {
		if ma.manager == nil {
			return FilterContinue, nil
		}

		// 获取请求路径
		path := string(ctx.Request().URI().Path())

		// 执行过滤器链
		return ma.manager.ExecuteFilters(ctx, path)
	}
}

// MiddlewareConfig 中间件配置
//
// 用于自定义中间件适配器的行为。
type MiddlewareConfig struct {
	// 错误处理
	HandleError bool // 是否处理过滤器错误
	LogErrors   bool // 是否记录错误日志

	// 路径控制
	SkipPaths []string // 跳过处理的路径列表

	// 自定义处理函数
	ErrorHandler func(*context.Context, error) // 自定义错误处理函数
	Logger       func(string, error)           // 自定义日志函数
}

// DefaultMiddlewareConfig 默认中间件配置
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		HandleError: true,
		LogErrors:   true,
		SkipPaths:   []string{},
	}
}

// shouldSkipPath 检查是否应该跳过指定路径
func (config *MiddlewareConfig) shouldSkipPath(path string) bool {
	for _, skipPath := range config.SkipPaths {
		if matchPath(skipPath, path) {
			return true
		}
	}
	return false
}

// ============= 错误处理辅助函数 =============

// handleFilterError 处理过滤器执行错误
func handleFilterError(ctx *context.Context, err error) {
	// 设置错误响应
	ctx.Request().SetStatusCode(500)
	ctx.Request().Response.SetBodyString("Internal Server Error: Filter execution failed")

	// 可以在这里添加更详细的错误处理逻辑
	// 比如根据错误类型返回不同的响应
}

// logFilterError 记录过滤器错误
func logFilterError(path string, err error) {
	// 这里可以集成具体的日志系统
	// 目前使用简单的打印
	println("Filter error on path", path, ":", err.Error())
}

// ============= 高级中间件适配器 =============

// ChainableMiddleware 可链式调用的中间件适配器
//
// 支持链式调用多个过滤器，提供更灵活的中间件组合方式。
type ChainableMiddleware struct {
	filters []string // 要执行的过滤器名称列表
	manager *Manager // 统一管理器实例
}

// NewChainableMiddleware 创建可链式调用的中间件
//
// 参数：
//   - filterNames: 要执行的过滤器名称列表
//
// 返回：
//   - *ChainableMiddleware: 可链式中间件实例
//
// 示例：
//
//	middleware := unified.NewChainableMiddleware([]string{"auth", "csrf", "sso"})
//	app.Use(middleware.ToHertzHandler())
func NewChainableMiddleware(filterNames []string) *ChainableMiddleware {
	return &ChainableMiddleware{
		filters: filterNames,
		manager: GetManager(),
	}
}

// ToMiddlewareHandler 转换为中间件处理器
func (cm *ChainableMiddleware) ToMiddlewareHandler() func(*context.Context, func()) {
	return func(c *context.Context, next func()) {
		if cm.manager == nil {
			if next != nil {
				next()
			}
			return
		}

		path := string(c.Request().URI().Path())

		// 按指定顺序执行过滤器
		for _, filterName := range cm.filters {
			filter := cm.manager.GetFilter(filterName)
			if filter == nil || !filter.Enabled {
				continue
			}

			// 检查路径匹配
			if !cm.manager.matchFilterPattern(filter, path) {
				continue
			}

			// 执行过滤器
			result, err := filter.FilterFunc(c)
			if err != nil {
				handleFilterError(c, err)
				return
			}

			// 根据结果决定是否继续
			switch result {
			case FilterStop:
				return // 停止执行
			case FilterSkip:
				if next != nil {
					next() // 跳过后续处理但继续中间件链
				}
				return
			case FilterContinue:
				continue // 继续执行下一个过滤器
			}
		}

		// 所有过滤器都通过，继续执行
		if next != nil {
			next()
		}
	}
}

// ============= 特殊用途中间件适配器 =============

// SSOMiddleware 专门的SSO中间件适配器
//
// 提供简化的SSO中间件，自动配置SSO过滤器。
type SSOMiddleware struct {
	config *SSOConfig // SSO配置
}

// NewSSOMiddleware 创建SSO中间件
//
// 参数：
//   - config: SSO配置，如果为nil则使用默认配置
//
// 返回：
//   - *SSOMiddleware: SSO中间件实例
//
// 示例：
//
//	ssoMiddleware := unified.NewSSOMiddleware(nil) // 使用默认配置
//	app.Use(ssoMiddleware.ToHertzHandler())
func NewSSOMiddleware(config *SSOConfig) *SSOMiddleware {
	if config == nil {
		config = DefaultSSOConfig()
	}

	// 设置全局SSO配置
	SetSSOConfig(config)

	return &SSOMiddleware{
		config: config,
	}
}

// ToMiddlewareHandler 转换为中间件处理器
func (sm *SSOMiddleware) ToMiddlewareHandler() func(*context.Context, func()) {
	return func(c *context.Context, next func()) {
		if !sm.config.Enabled {
			if next != nil {
				next()
			}
			return
		}

		// 直接执行SSO过滤器
		result, err := FilterSSO(c)
		if err != nil {
			handleFilterError(c, err)
			return
		}

		// 处理结果
		switch result {
		case FilterContinue:
			if next != nil {
				next()
			}
		case FilterStop:
			// SSO过滤器已经设置了响应，不需要继续
			return
		case FilterSkip:
			if next != nil {
				next()
			}
		}
	}
}

// ============= 实用工具函数 =============

// CreateMiddlewareFromFilter 从单个过滤器创建中间件
//
// 快速将单个过滤器转换为中间件的工具函数。
//
// 参数：
//   - filter: 要转换的过滤器
//
// 返回：
//   - func(*context.Context, func()): 中间件函数
//
// 示例：
//
//	middleware := unified.CreateMiddlewareFromFilter(&Filter{
//	    Name: "custom",
//	    Pattern: "/api/*",
//	    FilterFunc: myCustomFilter,
//	    Priority: 100,
//	    Enabled: true,
//	})
func CreateMiddlewareFromFilter(filter *Filter) func(*context.Context, func()) {
	return func(c *context.Context, next func()) {
		if filter == nil || !filter.Enabled {
			if next != nil {
				next()
			}
			return
		}

		path := string(c.Request().URI().Path())

		// 检查路径匹配
		manager := GetManager()
		if !manager.matchFilterPattern(filter, path) {
			if next != nil {
				next()
			}
			return
		}

		// 执行过滤器
		result, err := filter.FilterFunc(c)
		if err != nil {
			handleFilterError(c, err)
			return
		}

		// 处理结果
		switch result {
		case FilterContinue:
			if next != nil {
				next()
			}
		case FilterStop:
			return
		case FilterSkip:
			if next != nil {
				next()
			}
		}
	}
}

// CreateGlobalSSO 创建全局SSO中间件
//
// 便捷函数，快速创建SSO中间件。
//
// 参数：
//   - config: SSO配置，如果为nil则使用默认配置
//
// 返回：
//   - func(*context.Context, func()): SSO中间件函数
//
// 示例：
//
//	ssoMiddleware := unified.CreateGlobalSSO(nil) // 使用默认SSO配置
func CreateGlobalSSO(config *SSOConfig) func(*context.Context, func()) {
	ssoMiddleware := NewSSOMiddleware(config)
	return ssoMiddleware.ToMiddlewareHandler()
}