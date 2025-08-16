package errors

import (
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 错误控制器工厂 =============

// DefaultErrorControllerFactory 默认错误控制器工厂
type DefaultErrorControllerFactory struct {
	configManager     ConfigManager
	statisticsManager StatisticsManager
	templateManager   TemplateManager
}

// NewDefaultErrorControllerFactory 创建默认错误控制器工厂
func NewDefaultErrorControllerFactory() *DefaultErrorControllerFactory {
	return &DefaultErrorControllerFactory{
		configManager:     GetGlobalConfigManager(),
		statisticsManager: GetGlobalStatisticsManager(),
		templateManager:   GetGlobalTemplateManager(),
	}
}

// CreateController 创建错误控制器（根据环境）
func (f *DefaultErrorControllerFactory) CreateController(env string) ErrorHandler {
	switch env {
	case "production", "prod":
		return f.CreateProductionController()
	case "development", "dev":
		return f.CreateDevelopmentController()
	case "testing", "test":
		return f.CreateTestingController()
	case "high_performance", "perf":
		return f.CreateHighPerformanceController()
	default:
		return f.CreateDevelopmentController() // 默认开发环境
	}
}

// CreateWithConfig 使用配置创建控制器
func (f *DefaultErrorControllerFactory) CreateWithConfig(config *ErrorControllerConfig) ErrorHandler {
	controller := NewOptimizedErrorController()
	controller.SetConfig(config)
	return controller
}

// CreateProductionController 创建生产环境控制器
func (f *DefaultErrorControllerFactory) CreateProductionController() ErrorHandler {
	config := &ErrorControllerConfig{
		ShowDetailedError: false,
		Language:         "zh-CN",
		CustomTitle:      "YYHertz Framework - Production",
		EnableDebugInfo:  false,
	}
	
	controller := f.CreateWithConfig(config)
	
	// 生产环境优化配置
	if opt, ok := controller.(*OptimizedErrorController); ok {
		// 设置高性能统计管理器
		opt.SetStatisticsManager(NewHighPerformanceStatisticsManager())
		
		// 添加生产环境专用钩子
		opt.AddHook("before", func(ctx *mvccontext.Context, statusCode int, err error) error {
			// 生产环境日志记录（可以集成到日志系统）
			return nil
		})
	}
	
	return controller
}

// CreateDevelopmentController 创建开发环境控制器
func (f *DefaultErrorControllerFactory) CreateDevelopmentController() ErrorHandler {
	config := &ErrorControllerConfig{
		ShowDetailedError: true,
		Language:         "zh-CN",
		CustomTitle:      "YYHertz Framework - Development",
		EnableDebugInfo:  true,
		SupportEmail:     "dev@example.com",
		SupportPhone:     "400-000-0000",
	}
	
	controller := f.CreateWithConfig(config)
	
	// 开发环境配置
	if opt, ok := controller.(*OptimizedErrorController); ok {
		// 启用热重载（如果模板管理器支持）
		if dtm, ok := opt.templateManager.(*DefaultTemplateManager); ok {
			dtm.EnableHotReload(true)
		}
		
		// 添加开发环境调试钩子
		opt.AddHook("before", func(ctx *mvccontext.Context, statusCode int, err error) error {
			// 开发环境详细日志
			return nil
		})
	}
	
	return controller
}

// CreateTestingController 创建测试环境控制器
func (f *DefaultErrorControllerFactory) CreateTestingController() ErrorHandler {
	config := &ErrorControllerConfig{
		ShowDetailedError: true,
		Language:         "zh-CN",
		CustomTitle:      "YYHertz Framework - Testing",
		EnableDebugInfo:  true,
	}
	
	controller := f.CreateWithConfig(config)
	
	// 测试环境配置
	if opt, ok := controller.(*OptimizedErrorController); ok {
		// 禁用缓存，确保测试的一致性（如果模板管理器支持）
		if dtm, ok := opt.templateManager.(*DefaultTemplateManager); ok {
			dtm.EnableCache(false)
		}
		
		// 重置统计，避免测试间的影响
		opt.ResetStatistics()
	}
	
	return controller
}

// CreateHighPerformanceController 创建高性能控制器
func (f *DefaultErrorControllerFactory) CreateHighPerformanceController() ErrorHandler {
	return NewHighPerformanceErrorController()
}

// ============= 组件工厂 =============

// ComponentFactory 组件工厂
type ComponentFactory struct{}

// NewComponentFactory 创建组件工厂
func NewComponentFactory() *ComponentFactory {
	return &ComponentFactory{}
}

// CreateConfigManager 创建配置管理器
func (f *ComponentFactory) CreateConfigManager(managerType string) ConfigManager {
	switch managerType {
	case "default":
		return NewDefaultConfigManager()
	case "cached":
		// 缓存版本配置管理器（可以后续实现）
		return NewDefaultConfigManager()
	case "file":
		// 文件版本配置管理器（可以后续实现）
		return NewDefaultConfigManager()
	default:
		return NewDefaultConfigManager()
	}
}

// CreateStatisticsManager 创建统计管理器
func (f *ComponentFactory) CreateStatisticsManager(managerType string) StatisticsManager {
	switch managerType {
	case "default":
		return NewDefaultStatisticsManager()
	case "high_performance":
		return NewHighPerformanceStatisticsManager()
	case "memory":
		return NewDefaultStatisticsManager()
	default:
		return NewDefaultStatisticsManager()
	}
}

// CreateTemplateManager 创建模板管理器
func (f *ComponentFactory) CreateTemplateManager(managerType string) TemplateManager {
	switch managerType {
	case "default":
		return NewDefaultTemplateManager()
	case "cached":
		manager := NewDefaultTemplateManager()
		manager.EnableCache(true)
		return manager
	case "hot_reload":
		manager := NewDefaultTemplateManager()
		manager.EnableHotReload(true)
		return manager
	default:
		return NewDefaultTemplateManager()
	}
}

// CreateStatusHandler 创建状态处理器
func (f *ComponentFactory) CreateStatusHandler(handlerType string, configManager ConfigManager) StatusHandler {
	switch handlerType {
	case "universal":
		return NewUniversalStatusHandler(configManager)
	case "4xx":
		return Create4xxHandler(configManager)
	case "5xx":
		return Create5xxHandler(configManager)
	case "business":
		return NewBusinessErrorStatusHandler(configManager)
	case "content_negotiation":
		return NewContentNegotiationHandler(configManager)
	case "fast":
		return NewFastStatusHandler()
	default:
		return NewContentNegotiationHandler(configManager)
	}
}

// ============= 环境预设配置 =============

// EnvironmentConfig 环境配置
type EnvironmentConfig struct {
	Name                string
	ErrorControllerConf *ErrorControllerConfig
	ConfigManagerType   string
	StatisticsType      string
	TemplateType        string
	StatusHandlerType   string
}

// GetPredefinedConfigs 获取预定义配置
func GetPredefinedConfigs() map[string]*EnvironmentConfig {
	return map[string]*EnvironmentConfig{
		"production": {
			Name: "Production Environment",
			ErrorControllerConf: &ErrorControllerConfig{
				ShowDetailedError: false,
				Language:         "zh-CN",
				CustomTitle:      "YYHertz Framework",
				EnableDebugInfo:  false,
			},
			ConfigManagerType:  "cached",
			StatisticsType:     "high_performance",
			TemplateType:       "cached",
			StatusHandlerType:  "fast",
		},
		"development": {
			Name: "Development Environment",
			ErrorControllerConf: &ErrorControllerConfig{
				ShowDetailedError: true,
				Language:         "zh-CN",
				CustomTitle:      "YYHertz Framework - Dev",
				EnableDebugInfo:  true,
				SupportEmail:     "dev@example.com",
			},
			ConfigManagerType:  "default",
			StatisticsType:     "default",
			TemplateType:       "hot_reload",
			StatusHandlerType:  "content_negotiation",
		},
		"testing": {
			Name: "Testing Environment",
			ErrorControllerConf: &ErrorControllerConfig{
				ShowDetailedError: true,
				Language:         "zh-CN",
				CustomTitle:      "YYHertz Framework - Test",
				EnableDebugInfo:  true,
			},
			ConfigManagerType:  "default",
			StatisticsType:     "memory",
			TemplateType:       "default",
			StatusHandlerType:  "universal",
		},
	}
}

// ============= 快速构建器 =============

// ErrorControllerBuilder 错误控制器构建器
type ErrorControllerBuilder struct {
	config            *ErrorControllerConfig
	configManager     ConfigManager
	statisticsManager StatisticsManager
	templateManager   TemplateManager
	statusHandler     StatusHandler
	hooks             map[string][]ErrorHook
}

// NewErrorControllerBuilder 创建错误控制器构建器
func NewErrorControllerBuilder() *ErrorControllerBuilder {
	return &ErrorControllerBuilder{
		hooks: make(map[string][]ErrorHook),
	}
}

// WithConfig 设置配置
func (b *ErrorControllerBuilder) WithConfig(config *ErrorControllerConfig) *ErrorControllerBuilder {
	b.config = config
	return b
}

// WithConfigManager 设置配置管理器
func (b *ErrorControllerBuilder) WithConfigManager(manager ConfigManager) *ErrorControllerBuilder {
	b.configManager = manager
	return b
}

// WithStatisticsManager 设置统计管理器
func (b *ErrorControllerBuilder) WithStatisticsManager(manager StatisticsManager) *ErrorControllerBuilder {
	b.statisticsManager = manager
	return b
}

// WithTemplateManager 设置模板管理器
func (b *ErrorControllerBuilder) WithTemplateManager(manager TemplateManager) *ErrorControllerBuilder {
	b.templateManager = manager
	return b
}

// WithStatusHandler 设置状态处理器
func (b *ErrorControllerBuilder) WithStatusHandler(handler StatusHandler) *ErrorControllerBuilder {
	b.statusHandler = handler
	return b
}

// WithHook 添加钩子
func (b *ErrorControllerBuilder) WithHook(phase string, hook ErrorHook) *ErrorControllerBuilder {
	if b.hooks[phase] == nil {
		b.hooks[phase] = make([]ErrorHook, 0)
	}
	b.hooks[phase] = append(b.hooks[phase], hook)
	return b
}

// WithEnvironment 使用环境预设
func (b *ErrorControllerBuilder) WithEnvironment(env string) *ErrorControllerBuilder {
	configs := GetPredefinedConfigs()
	if envConfig, exists := configs[env]; exists {
		factory := NewComponentFactory()
		
		b.config = envConfig.ErrorControllerConf
		b.configManager = factory.CreateConfigManager(envConfig.ConfigManagerType)
		b.statisticsManager = factory.CreateStatisticsManager(envConfig.StatisticsType)
		b.templateManager = factory.CreateTemplateManager(envConfig.TemplateType)
		b.statusHandler = factory.CreateStatusHandler(envConfig.StatusHandlerType, b.configManager)
	}
	return b
}

// Build 构建错误控制器
func (b *ErrorControllerBuilder) Build() ErrorHandler {
	// 设置默认值
	if b.config == nil {
		b.config = &ErrorControllerConfig{
			ShowDetailedError: true,
			Language:         "zh-CN",
			CustomTitle:      "YYHertz Framework",
			EnableDebugInfo:  true,
		}
	}
	
	if b.configManager == nil {
		b.configManager = GetGlobalConfigManager()
	}
	
	if b.statisticsManager == nil {
		b.statisticsManager = GetGlobalStatisticsManager()
	}
	
	if b.templateManager == nil {
		b.templateManager = GetGlobalTemplateManager()
	}
	
	// 创建控制器
	controller := &OptimizedErrorController{
		configManager:     b.configManager,
		statisticsManager: b.statisticsManager,
		templateManager:   b.templateManager,
		hooks:            make(map[string][]ErrorHook),
		config:           b.config,
	}
	
	// 设置状态处理器
	if b.statusHandler != nil {
		controller.statusHandler = b.statusHandler
	} else {
		controller.initStatusHandler()
	}
	
	// 初始化其他组件
	controller.initRenderers()
	
	// 添加自定义钩子
	for phase, hookList := range b.hooks {
		for _, hook := range hookList {
			controller.AddHook(phase, hook)
		}
	}
	
	// 初始化默认钩子
	controller.initDefaultHooks()
	
	return controller
}

// ============= 全局工厂实例 =============

var (
	globalErrorControllerFactory = NewDefaultErrorControllerFactory()
	globalComponentFactory       = NewComponentFactory()
)

// GetGlobalErrorControllerFactory 获取全局错误控制器工厂
func GetGlobalErrorControllerFactory() ErrorControllerFactory {
	return globalErrorControllerFactory
}

// GetGlobalComponentFactory 获取全局组件工厂
func GetGlobalComponentFactory() *ComponentFactory {
	return globalComponentFactory
}

// ============= 便捷方法 =============

// CreateErrorController 创建错误控制器（便捷方法）
func CreateErrorController(env string) ErrorHandler {
	return globalErrorControllerFactory.CreateController(env)
}

// CreateErrorControllerWithConfig 使用配置创建控制器（便捷方法）
func CreateErrorControllerWithConfig(config *ErrorControllerConfig) ErrorHandler {
	return globalErrorControllerFactory.CreateWithConfig(config)
}

// QuickBuild 快速构建错误控制器（便捷方法）
func QuickBuild() *ErrorControllerBuilder {
	return NewErrorControllerBuilder()
}

// QuickBuildForEnv 为特定环境快速构建（便捷方法）
func QuickBuildForEnv(env string) ErrorHandler {
	return NewErrorControllerBuilder().WithEnvironment(env).Build()
}