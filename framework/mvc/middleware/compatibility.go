package middleware

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/config"
)

// CompatibilityLayer 兼容性层 - 提供向下兼容的API接口
type CompatibilityLayer struct {
	unifiedManager *UnifiedMiddlewareManager
}

// NewCompatibilityLayer 创建兼容性层
func NewCompatibilityLayer() *CompatibilityLayer {
	return &CompatibilityLayer{
		unifiedManager: GetGlobalUnifiedManager(),
	}
}

// ===== 基础中间件系统兼容接口 =====

// BasicMiddlewareWrapper 基础中间件包装器
type BasicMiddlewareWrapper struct {
	engine *Engine
	compat *CompatibilityLayer
}

// NewBasicEngine 创建兼容的基础引擎（保持原API）
func NewBasicEngine() *BasicMiddlewareWrapper {
	return &BasicMiddlewareWrapper{
		engine: NewEngine(),
		compat: NewCompatibilityLayer(),
	}
}

// Use 使用中间件 - 兼容原基础系统API
func (w *BasicMiddlewareWrapper) Use(middleware ...HandlerFunc) {
	for i, handler := range middleware {
		// 通过统一管理器注册中间件
		name := fmt.Sprintf("basic-middleware-%d", i)
		w.compat.unifiedManager.Use(name, handler, WithLayer(LayerGlobal), WithPriority(i*10))
	}
	
	// 同时在原引擎中注册（保持兼容性）
	w.engine.Use(middleware...)
}

// NewContext 创建上下文 - 兼容原API
func (w *BasicMiddlewareWrapper) NewContext(c *app.RequestContext) *Context {
	return w.engine.NewContext(c)
}

// ===== MVC中间件系统兼容接口 =====

// MVCMiddlewareWrapper MVC中间件包装器
type MVCMiddlewareWrapper struct {
	manager *MiddlewareManager
	compat  *CompatibilityLayer
}

// NewMVCManager 创建兼容的MVC管理器（保持原API）
func NewMVCManager() *MVCMiddlewareWrapper {
	return &MVCMiddlewareWrapper{
		manager: NewMiddlewareManager(),
		compat:  NewCompatibilityLayer(),
	}
}

// RegisterCustom 注册自定义中间件 - 兼容原API
func (w *MVCMiddlewareWrapper) RegisterCustom(name string, handler MiddlewareFunc, metadata MiddlewareMetadata) error {
	// 通过统一管理器注册
	err := w.compat.unifiedManager.Use(name, handler, WithLayer(LayerGlobal))
	if err != nil {
		return err
	}
	
	// 同时在原管理器中注册（保持兼容性）
	return w.manager.RegisterCustom(name, handler, metadata)
}

// UseBuiltin 使用内置中间件 - 兼容原API
func (w *MVCMiddlewareWrapper) UseBuiltin(layer MiddlewareLayer, name string, config interface{}, priority int) error {
	// 通过统一管理器注册
	err := w.compat.unifiedManager.UseBuiltin(name, config, WithLayer(layer), WithPriority(priority))
	if err != nil {
		return err
	}
	
	// 同时在原管理器中注册（保持兼容性）
	return w.manager.UseBuiltin(layer, name, config, priority)
}

// UseCustom 使用自定义中间件 - 兼容原API
func (w *MVCMiddlewareWrapper) UseCustom(layer MiddlewareLayer, name string, priority int) error {
	return w.manager.UseCustom(layer, name, priority)
}

// GetStatistics 获取统计信息 - 兼容原API
func (w *MVCMiddlewareWrapper) GetStatistics() ManagerStatistics {
	return w.manager.GetStatistics()
}

// ===== 统一兼容接口 =====

// MiddlewareCompat 中间件兼容性接口
type MiddlewareCompat interface {
	// 基础功能
	Use(name string, handler interface{}, options ...interface{}) error
	UseBuiltin(name string, config interface{}, options ...interface{}) error
	
	// 模式控制
	SwitchMode(mode string) error
	GetCurrentMode() string
	
	// 统计信息
	GetStats() map[string]interface{}
}

// UnifiedCompat 统一兼容性实现
type UnifiedCompat struct {
	manager *UnifiedMiddlewareManager
}

// NewUnifiedCompat 创建统一兼容性层
func NewUnifiedCompat() MiddlewareCompat {
	return &UnifiedCompat{
		manager: GetGlobalUnifiedManager(),
	}
}

// Use 使用中间件 - 统一接口
func (c *UnifiedCompat) Use(name string, handler interface{}, options ...interface{}) error {
	// 解析选项
	opts := c.parseOptions(options...)
	return c.manager.Use(name, handler, opts...)
}

// UseBuiltin 使用内置中间件 - 统一接口
func (c *UnifiedCompat) UseBuiltin(name string, config interface{}, options ...interface{}) error {
	opts := c.parseOptions(options...)
	return c.manager.UseBuiltin(name, config, opts...)
}

// SwitchMode 切换模式 - 统一接口
func (c *UnifiedCompat) SwitchMode(mode string) error {
	var middlewareMode config.MiddlewareMode
	switch mode {
	case "basic":
		middlewareMode = config.BasicMode
	case "advanced", "mvc":
		middlewareMode = config.AdvancedMode
	case "auto":
		middlewareMode = config.AutoMode
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}
	
	return c.manager.SwitchMode(middlewareMode)
}

// GetCurrentMode 获取当前模式 - 统一接口
func (c *UnifiedCompat) GetCurrentMode() string {
	mode := c.manager.GetCurrentMode()
	return string(mode)
}

// GetStats 获取统计信息 - 统一接口
func (c *UnifiedCompat) GetStats() map[string]interface{} {
	stats := c.manager.GetStats()
	return map[string]interface{}{
		"current_mode":         stats.CurrentMode,
		"total_requests":      stats.TotalRequests,
		"basic_mode_requests": stats.BasicModeRequests,
		"mvc_mode_requests":   stats.MVCModeRequests,
		"average_response_time": stats.AverageResponseTime.String(),
		"mode_switch_count":   stats.ModeSwitchCount,
		"last_switch_time":    stats.LastSwitchTime,
	}
}

// parseOptions 解析选项
func (c *UnifiedCompat) parseOptions(options ...interface{}) []MiddlewareOption {
	var opts []MiddlewareOption
	
	for _, option := range options {
		switch v := option.(type) {
		case string:
			// 字符串参数可能是层级
			if layer := c.parseLayer(v); layer != nil {
				opts = append(opts, WithLayer(*layer))
			}
		case int:
			// 整数参数是优先级
			opts = append(opts, WithPriority(v))
		case MiddlewareLayer:
			// 直接的层级参数
			opts = append(opts, WithLayer(v))
		case map[string]interface{}:
			// 配置参数
			opts = append(opts, WithConfig(v))
		}
	}
	
	return opts
}

// parseLayer 解析层级字符串
func (c *UnifiedCompat) parseLayer(layerStr string) *MiddlewareLayer {
	switch layerStr {
	case "global":
		layer := LayerGlobal
		return &layer
	case "group":
		layer := LayerGroup
		return &layer
	case "route":
		layer := LayerRoute
		return &layer
	case "controller":
		layer := LayerController
		return &layer
	default:
		return nil
	}
}

// ===== 向下兼容的全局函数 =====

// 保持原有的全局函数接口不变

// UseBasicMiddleware 使用基础中间件（向下兼容）
func UseBasicMiddleware(middleware Middleware) {
	globalUnifiedManager.Use("basic-global", middleware, WithLayer(LayerGlobal))
}

// UseBasicHandlerFunc 使用基础HandlerFunc（向下兼容）
func UseBasicHandlerFunc(handler HandlerFunc) {
	globalUnifiedManager.Use("basic-handler", handler, WithLayer(LayerGlobal))
}

// UseMVCMiddleware 使用MVC中间件（向下兼容）
func UseMVCMiddleware(layer MiddlewareLayer, name string, handler MiddlewareFunc, priority int) error {
	return globalUnifiedManager.Use(name, handler, WithLayer(layer), WithPriority(priority))
}

// UseMVCBuiltin 使用MVC内置中间件（向下兼容）
func UseMVCBuiltin(layer MiddlewareLayer, name string, config interface{}, priority int) error {
	return globalUnifiedManager.UseBuiltin(name, config, WithLayer(layer), WithPriority(priority))
}

// ===== 迁移辅助函数 =====

// MigrationHelper 迁移辅助器
type MigrationHelper struct {
	compat *CompatibilityLayer
}

// NewMigrationHelper 创建迁移辅助器
func NewMigrationHelper() *MigrationHelper {
	return &MigrationHelper{
		compat: NewCompatibilityLayer(),
	}
}

// MigrateBasicEngine 迁移基础引擎到统一系统
func (h *MigrationHelper) MigrateBasicEngine(engine *Engine) error {
	// 这里需要反射或其他方法来提取已注册的中间件
	// 由于基础引擎没有提供获取中间件列表的方法，这里提供框架
	
	fmt.Println("Migrating basic engine to unified system...")
	// TODO: 实现具体的迁移逻辑
	return nil
}

// MigrateMVCManager 迁移MVC管理器到统一系统
func (h *MigrationHelper) MigrateMVCManager(manager *MiddlewareManager) error {
	fmt.Println("Migrating MVC manager to unified system...")
	
	// 获取所有注册的中间件
	middlewares := manager.ListMiddlewares()
	
	for _, middleware := range middlewares {
		// 根据中间件元数据进行迁移
		if middleware.IsBuiltin {
			h.compat.unifiedManager.UseBuiltin(middleware.Name, nil)
		} else {
			// 自定义中间件需要从注册表中获取处理函数
			// 这里需要扩展MiddlewareManager API来支持获取处理函数
		}
	}
	
	return nil
}

// GenerateMigrationReport 生成迁移报告
func (h *MigrationHelper) GenerateMigrationReport() MigrationReport {
	return MigrationReport{
		TotalMiddlewares: 0, // TODO: 计算实际数量
		BasicMiddlewares: 0,
		MVCMiddlewares:  0,
		Recommendations: []string{
			"Consider upgrading to advanced mode for better performance",
			"Enable auto mode for dynamic switching",
		},
	}
}

// MigrationReport 迁移报告
type MigrationReport struct {
	TotalMiddlewares int      `json:"total_middlewares"`
	BasicMiddlewares int      `json:"basic_middlewares"`
	MVCMiddlewares   int      `json:"mvc_middlewares"`
	Recommendations  []string `json:"recommendations"`
}

// ===== 全局兼容性实例 =====

var (
	globalCompat       = NewUnifiedCompat()
	globalMigrationHelper = NewMigrationHelper()
)

// GetGlobalCompat 获取全局兼容性层
func GetGlobalCompat() MiddlewareCompat {
	return globalCompat
}

// GetGlobalMigrationHelper 获取全局迁移辅助器
func GetGlobalMigrationHelper() *MigrationHelper {
	return globalMigrationHelper
}

// ===== 便捷的迁移函数 =====

// QuickMigrate 快速迁移到统一系统
func QuickMigrate(mode string) error {
	return globalCompat.SwitchMode(mode)
}

// GetMigrationReport 获取迁移报告
func GetMigrationReport() MigrationReport {
	return globalMigrationHelper.GenerateMigrationReport()
}

// EnableAutoMode 启用自动模式
func EnableAutoMode() error {
	return globalCompat.SwitchMode("auto")
}

// GetSystemStatus 获取系统状态
func GetSystemStatus() map[string]interface{} {
	stats := globalCompat.GetStats()
	stats["current_mode"] = globalCompat.GetCurrentMode()
	return stats
}