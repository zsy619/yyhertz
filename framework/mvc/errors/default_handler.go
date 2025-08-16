package errors

import (
	"sync"
	
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 新的优化版DefaultErrorController =============

// OptimizedErrorController 优化版错误控制器
type OptimizedErrorController struct {
	// 核心组件
	configManager     ConfigManager
	statisticsManager StatisticsManager
	templateManager   TemplateManager
	
	// 渲染器
	multiRenderer *MultiRenderer
	
	// 状态处理器
	statusHandler StatusHandler
	
	// 钩子管理
	hooks map[string][]ErrorHook
	
	// 配置
	config *ErrorControllerConfig
	
	// 并发安全
	mu sync.RWMutex
}

// NewOptimizedErrorController 创建优化版错误控制器
func NewOptimizedErrorController() *OptimizedErrorController {
	controller := &OptimizedErrorController{
		configManager:     GetGlobalConfigManager(),
		statisticsManager: GetGlobalStatisticsManager(),
		templateManager:   GetGlobalTemplateManager(),
		hooks:            make(map[string][]ErrorHook),
		config: &ErrorControllerConfig{
			ShowDetailedError: true,
			Language:         "zh-CN",
			CustomTitle:      "YYHertz Framework",
			EnableDebugInfo:  true,
		},
	}
	
	// 初始化渲染器
	controller.initRenderers()
	
	// 初始化状态处理器
	controller.initStatusHandler()
	
	// 初始化默认钩子
	controller.initDefaultHooks()
	
	return controller
}

// Handle 处理错误（实现ErrorHandler接口）
func (c *OptimizedErrorController) Handle(ctx *mvccontext.Context, statusCode int, err error) error {
	// 记录统计
	c.statisticsManager.RecordError(statusCode, string(ctx.Path()), string(ctx.Method()), err)
	
	// 执行前置钩子
	if err := c.executeHooks("before", ctx, statusCode, err); err != nil {
		// 钩子失败不影响错误处理，只记录
		c.logError("Before hook failed", err)
	}
	
	// 使用状态处理器处理
	handleErr := c.statusHandler.HandleStatus(ctx, statusCode, err)
	
	// 执行后置钩子
	if err := c.executeHooks("after", ctx, statusCode, err); err != nil {
		c.logError("After hook failed", err)
	}
	
	return handleErr
}

// CanHandle 检查是否能处理该错误
func (c *OptimizedErrorController) CanHandle(statusCode int, err error) bool {
	// 优化版控制器可以处理所有错误
	return true
}

// Priority 返回处理器优先级
func (c *OptimizedErrorController) Priority() int {
	return 1000 // 默认优先级
}

// ============= 初始化方法 =============

// initRenderers 初始化渲染器
func (c *OptimizedErrorController) initRenderers() {
	c.multiRenderer = CreateDefaultMultiRenderer()
	
	// 配置HTML渲染器
	htmlRenderer := NewHTMLRenderer(c.templateManager)
	htmlRenderer.SetShowDetails(c.config.ShowDetailedError)
	htmlRenderer.SetSupportInfo(c.config.SupportEmail, c.config.SupportPhone)
	
	// 替换默认HTML渲染器
	c.multiRenderer.renderers[1] = htmlRenderer // 假设索引1是HTML渲染器
}

// initStatusHandler 初始化状态处理器
func (c *OptimizedErrorController) initStatusHandler() {
	// 使用内容协商处理器
	handler := NewContentNegotiationHandler(c.configManager)
	handler.SetRenderers(
		NewHTMLRenderer(c.templateManager),
		NewJSONRenderer(c.config.ShowDetailedError, false),
		NewXMLRenderer(false),
	)
	c.statusHandler = handler
}

// initDefaultHooks 初始化默认钩子
func (c *OptimizedErrorController) initDefaultHooks() {
	// 前置钩子：日志记录
	c.AddHook("before", func(ctx *mvccontext.Context, statusCode int, err error) error {
		// 简单的日志记录
		path := string(ctx.Path())
		method := string(ctx.Method())
		c.logErrorWithContext(statusCode, path, method, err)
		return nil
	})
	
	// 后置钩子：清理工作
	c.AddHook("after", func(ctx *mvccontext.Context, statusCode int, err error) error {
		// 这里可以做清理工作
		return nil
	})
}

// ============= 配置方法 =============

// SetConfig 设置配置
func (c *OptimizedErrorController) SetConfig(config *ErrorControllerConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.config = config
	
	// 重新初始化组件
	c.initRenderers()
}

// GetConfig 获取配置
func (c *OptimizedErrorController) GetConfig() *ErrorControllerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// 返回配置副本
	configCopy := *c.config
	return &configCopy
}

// SetConfigManager 设置配置管理器
func (c *OptimizedErrorController) SetConfigManager(manager ConfigManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.configManager = manager
	c.initStatusHandler() // 重新初始化状态处理器
}

// SetStatisticsManager 设置统计管理器
func (c *OptimizedErrorController) SetStatisticsManager(manager StatisticsManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.statisticsManager = manager
}

// SetTemplateManager 设置模板管理器
func (c *OptimizedErrorController) SetTemplateManager(manager TemplateManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.templateManager = manager
	c.initRenderers() // 重新初始化渲染器
}

// ============= 钩子管理方法 =============

// AddHook 添加钩子
func (c *OptimizedErrorController) AddHook(phase string, hook ErrorHook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.hooks[phase] == nil {
		c.hooks[phase] = make([]ErrorHook, 0)
	}
	c.hooks[phase] = append(c.hooks[phase], hook)
}

// RemoveHooks 移除钩子
func (c *OptimizedErrorController) RemoveHooks(phase string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.hooks, phase)
}

// executeHooks 执行钩子
func (c *OptimizedErrorController) executeHooks(phase string, ctx *mvccontext.Context, statusCode int, err error) error {
	c.mu.RLock()
	hooks := c.hooks[phase]
	c.mu.RUnlock()
	
	for _, hook := range hooks {
		if hookErr := hook(ctx, statusCode, err); hookErr != nil {
			return hookErr
		}
	}
	
	return nil
}

// ============= 统计和监控方法 =============

// GetStatistics 获取错误统计
func (c *OptimizedErrorController) GetStatistics() *ErrorStatistics {
	return c.statisticsManager.GetStatistics()
}

// ResetStatistics 重置统计
func (c *OptimizedErrorController) ResetStatistics() {
	c.statisticsManager.Reset()
}

// GetErrorRate 获取错误率
func (c *OptimizedErrorController) GetErrorRate(timeWindow int) float64 {
	return c.statisticsManager.GetErrorRate(timeWindow)
}

// ============= 模板和渲染方法 =============

// RenderTemplate 渲染模板
func (c *OptimizedErrorController) RenderTemplate(name string, data interface{}) (string, error) {
	return c.templateManager.RenderTemplate(name, data)
}

// SetTemplate 设置自定义模板
func (c *OptimizedErrorController) SetTemplate(name, content string) error {
	return c.templateManager.LoadTemplate(name, content)
}

// ============= 辅助方法 =============

// logError 记录错误
func (c *OptimizedErrorController) logError(message string, err error) {
	// 这里可以集成更强大的日志系统
	// 现在只是简单输出
	_ = message
	_ = err
}

// logErrorWithContext 记录带上下文的错误
func (c *OptimizedErrorController) logErrorWithContext(statusCode int, path, method string, err error) {
	// 这里可以集成更强大的日志系统
	// 现在只是简单输出
	_ = statusCode
	_ = path
	_ = method
	_ = err
}

// ============= 兼容性方法 =============

// 为了与原有的DefaultErrorController保持兼容，提供一些兼容性方法

// ShowDetailedError 是否显示详细错误
func (c *OptimizedErrorController) ShowDetailedError() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.ShowDetailedError
}

// SetShowDetailedError 设置是否显示详细错误
func (c *OptimizedErrorController) SetShowDetailedError(show bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.ShowDetailedError = show
}

// GetLanguage 获取语言
func (c *OptimizedErrorController) GetLanguage() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.Language
}

// SetLanguage 设置语言
func (c *OptimizedErrorController) SetLanguage(language string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Language = language
}

// ============= 工厂方法 =============

// NewProductionOptimizedController 创建生产环境优化控制器
func NewProductionOptimizedController() *OptimizedErrorController {
	controller := NewOptimizedErrorController()
	controller.config.ShowDetailedError = false
	controller.config.EnableDebugInfo = false
	controller.config.CustomTitle = "YYHertz Framework - Production"
	
	// 重新初始化组件
	controller.initRenderers()
	
	return controller
}

// NewDevelopmentOptimizedController 创建开发环境优化控制器
func NewDevelopmentOptimizedController() *OptimizedErrorController {
	controller := NewOptimizedErrorController()
	controller.config.ShowDetailedError = true
	controller.config.EnableDebugInfo = true
	controller.config.CustomTitle = "YYHertz Framework - Development"
	
	return controller
}

// ============= 性能优化版本 =============

// HighPerformanceErrorController 高性能错误控制器
type HighPerformanceErrorController struct {
	// 预编译的处理器
	fastHandler *FastStatusHandler
	
	// 最小化配置
	showDetails bool
	language    string
	
	// 预分配的缓冲区
	bufferPool sync.Pool
}

// NewHighPerformanceErrorController 创建高性能错误控制器
func NewHighPerformanceErrorController() *HighPerformanceErrorController {
	controller := &HighPerformanceErrorController{
		fastHandler: NewFastStatusHandler(),
		showDetails: false, // 高性能版本默认不显示详情
		language:    "zh-CN",
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // 预分配1KB缓冲区
			},
		},
	}
	
	return controller
}

// Handle 高性能错误处理
func (c *HighPerformanceErrorController) Handle(ctx *mvccontext.Context, statusCode int, err error) error {
	// 使用预编译的快速处理器
	return c.fastHandler.HandleStatus(ctx, statusCode, err)
}

// CanHandle 检查是否能处理该错误
func (c *HighPerformanceErrorController) CanHandle(statusCode int, err error) bool {
	return true
}

// Priority 返回处理器优先级
func (c *HighPerformanceErrorController) Priority() int {
	return 100 // 高优先级
}

// ============= 全局优化控制器实例 =============

var globalOptimizedController = NewOptimizedErrorController()

// GetGlobalOptimizedController 获取全局优化控制器
func GetGlobalOptimizedController() *OptimizedErrorController {
	return globalOptimizedController
}

// HandleErrorOptimized 使用优化控制器处理错误（全局方法）
func HandleErrorOptimized(ctx *mvccontext.Context, statusCode int, err error) error {
	return globalOptimizedController.Handle(ctx, statusCode, err)
}