package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// UnifiedMiddlewareManager 统一中间件管理器
type UnifiedMiddlewareManager struct {
	// 基础系统引擎
	basicEngine *Engine
	
	// 高级系统管理器  
	mvcManager *MiddlewareManager
	
	// 当前模式
	currentMode config.MiddlewareMode
	
	// 配置
	config *config.MiddlewareUnifiedConfig
	
	// 适配器
	adapter *MiddlewareAdapter
	converter *ContextConverter
	
	// 统计信息
	stats UnifiedStats
	
	// 同步控制
	mu sync.RWMutex
	
	// 自动切换控制
	autoSwitcher *AutoModeSwitcher
}

// UnifiedStats 统一统计信息
type UnifiedStats struct {
	CurrentMode        config.MiddlewareMode `json:"current_mode"`
	TotalRequests      int64                 `json:"total_requests"`
	BasicModeRequests  int64                 `json:"basic_mode_requests"`
	MVCModeRequests    int64                 `json:"mvc_mode_requests"`
	AverageResponseTime time.Duration       `json:"average_response_time"`
	ModeSwitchCount     int64                `json:"mode_switch_count"`
	LastSwitchTime      time.Time            `json:"last_switch_time"`
}

// AutoModeSwitcher 自动模式切换器
type AutoModeSwitcher struct {
	manager     *UnifiedMiddlewareManager
	stopCh      chan struct{}
	running     bool
	mu          sync.RWMutex
	
	// 性能监控
	requestCount    int64
	responseTimeSum time.Duration
	lastCheckTime   time.Time
}

// NewUnifiedMiddlewareManager 创建统一中间件管理器
func NewUnifiedMiddlewareManager() *UnifiedMiddlewareManager {
	// 加载配置
	cfg, err := config.GetMiddlewareUnifiedConfig()
	if err != nil {
		// 使用默认配置
		defaultCfg := config.MiddlewareUnifiedConfig{}
		cfg = &defaultCfg
	}
	
	manager := &UnifiedMiddlewareManager{
		basicEngine: NewEngine(),
		mvcManager:  NewMiddlewareManager(),
		currentMode: cfg.GetMode(),
		config:      cfg,
		adapter:     NewMiddlewareAdapter("unified"),
		converter:   NewContextConverter(),
		stats:       UnifiedStats{CurrentMode: cfg.GetMode()},
	}
	
	// 如果是自动模式，启动自动切换器
	if cfg.IsAutoMode() {
		manager.autoSwitcher = NewAutoModeSwitcher(manager)
	}
	
	// 初始化内置中间件
	manager.initBuiltinMiddlewares()
	
	return manager
}

// NewAutoModeSwitcher 创建自动模式切换器
func NewAutoModeSwitcher(manager *UnifiedMiddlewareManager) *AutoModeSwitcher {
	return &AutoModeSwitcher{
		manager:       manager,
		stopCh:        make(chan struct{}),
		lastCheckTime: time.Now(),
	}
}

// Use 注册中间件 (统一接口)
func (m *UnifiedMiddlewareManager) Use(name string, handler interface{}, options ...MiddlewareOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	opts := parseMiddlewareOptions(options...)
	
	switch m.currentMode {
	case config.BasicMode:
		return m.useInBasicMode(name, handler, opts)
	case config.AdvancedMode:
		return m.useInMVCMode(name, handler, opts)
	case config.AutoMode:
		// 自动模式下同时注册到两套系统
		if err := m.useInBasicMode(name, handler, opts); err != nil {
			return err
		}
		return m.useInMVCMode(name, handler, opts)
	default:
		return fmt.Errorf("unsupported middleware mode: %s", m.currentMode)
	}
}

// UseBuiltin 使用内置中间件
func (m *UnifiedMiddlewareManager) UseBuiltin(name string, config interface{}, options ...MiddlewareOption) error {
	opts := parseMiddlewareOptions(options...)
	
	switch m.currentMode {
	case "basic":
		return m.useBuiltinInBasicMode(name, config, opts)
	case "advanced":
		return m.useBuiltinInMVCMode(name, config, opts)
	case "auto":
		// 自动模式下优先使用MVC系统的内置中间件
		return m.useBuiltinInMVCMode(name, config, opts)
	default:
		return fmt.Errorf("unsupported middleware mode: %s", m.currentMode)
	}
}

// SwitchMode 切换中间件模式
func (m *UnifiedMiddlewareManager) SwitchMode(mode config.MiddlewareMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.currentMode == mode {
		return nil // 已经是目标模式
	}
	
	oldMode := m.currentMode
	m.currentMode = mode
	m.stats.CurrentMode = mode
	m.stats.ModeSwitchCount++
	m.stats.LastSwitchTime = time.Now()
	
	fmt.Printf("Middleware mode switched: %s -> %s\n", oldMode, mode)
	return nil
}

// GetCurrentMode 获取当前模式
func (m *UnifiedMiddlewareManager) GetCurrentMode() config.MiddlewareMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentMode
}

// ExecuteMiddleware 执行中间件链 (统一接口)
func (m *UnifiedMiddlewareManager) ExecuteMiddleware(ctx interface{}) error {
	switch m.currentMode {
	case config.BasicMode:
		return m.executeInBasicMode(ctx)
	case config.AdvancedMode:
		return m.executeInMVCMode(ctx)
	case config.AutoMode:
		// 自动模式根据当前性能选择执行方式
		return m.executeInAutoMode(ctx)
	default:
		return fmt.Errorf("unsupported middleware mode: %s", m.currentMode)
	}
}

// GetStats 获取统计信息
func (m *UnifiedMiddlewareManager) GetStats() UnifiedStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// StartAutoSwitcher 启动自动切换器
func (m *UnifiedMiddlewareManager) StartAutoSwitcher() error {
	if m.autoSwitcher == nil {
		return fmt.Errorf("auto switcher not initialized")
	}
	return m.autoSwitcher.Start()
}

// StopAutoSwitcher 停止自动切换器
func (m *UnifiedMiddlewareManager) StopAutoSwitcher() error {
	if m.autoSwitcher == nil {
		return fmt.Errorf("auto switcher not initialized")
	}
	return m.autoSwitcher.Stop()
}

// 私有方法实现

func (m *UnifiedMiddlewareManager) useInBasicMode(name string, handler interface{}, opts MiddlewareOptions) error {
	switch h := handler.(type) {
	case HandlerFunc:
		m.basicEngine.Use(h)
		return nil
	case MiddlewareFunc:
		// MVC中间件转换为基础中间件
		convertedHandler := func(ctx *Context) {
			mvcCtx := m.converter.BasicToMVCContext(ctx)
			h(mvcCtx)
			SyncContextState(ctx, mvcCtx)
		}
		m.basicEngine.Use(convertedHandler)
		return nil
	default:
		return fmt.Errorf("unsupported handler type for basic mode")
	}
}

func (m *UnifiedMiddlewareManager) useInMVCMode(name string, handler interface{}, opts MiddlewareOptions) error {
	layer := opts.Layer
	if layer == nil {
		defaultLayer := LayerGlobal
		layer = &defaultLayer
	}
	
	priority := opts.Priority
	if priority == nil {
		defaultPriority := 50
		priority = &defaultPriority
	}
	
	switch h := handler.(type) {
	case MiddlewareFunc:
		err := m.mvcManager.RegisterCustom(name, h, MiddlewareMetadata{
			Name:        name,
			Description: "User registered middleware",
			Author:      "User",
		})
		if err != nil {
			return err
		}
		return m.mvcManager.UseCustom(*layer, name, *priority)
	case HandlerFunc:
		// 基础中间件转换为MVC中间件
		mvcHandler := HandlerFuncToMVC(h)
		err := m.mvcManager.RegisterCustom(name, mvcHandler, MiddlewareMetadata{
			Name:        name,
			Description: "Converted from basic middleware",
			Author:      "Adapter",
		})
		if err != nil {
			return err
		}
		return m.mvcManager.UseCustom(*layer, name, *priority)
	default:
		return fmt.Errorf("unsupported handler type for MVC mode")
	}
}

func (m *UnifiedMiddlewareManager) useBuiltinInBasicMode(name string, config interface{}, opts MiddlewareOptions) error {
	// 基础模式下使用基础中间件系统的内置中间件
	switch name {
	case "logger":
		m.basicEngine.Use(Logger())
	case "recovery":
		m.basicEngine.Use(Recovery())
	default:
		return fmt.Errorf("builtin middleware %s not supported in basic mode", name)
	}
	return nil
}

func (m *UnifiedMiddlewareManager) useBuiltinInMVCMode(name string, config interface{}, opts MiddlewareOptions) error {
	layer := opts.Layer
	if layer == nil {
		defaultLayer := LayerGlobal
		layer = &defaultLayer
	}
	
	priority := opts.Priority
	if priority == nil {
		defaultPriority := 50
		priority = &defaultPriority
	}
	
	return m.mvcManager.UseBuiltin(*layer, name, config, *priority)
}

func (m *UnifiedMiddlewareManager) executeInBasicMode(ctx interface{}) error {
	if basicCtx, ok := ctx.(*Context); ok {
		basicCtx.Next()
		m.updateStats(config.BasicMode, time.Since(time.Now()))
		return nil
	}
	return fmt.Errorf("invalid context type for basic mode")
}

func (m *UnifiedMiddlewareManager) executeInMVCMode(ctx interface{}) error {
	if mvcCtx, ok := ctx.(*mvccontext.EnhancedContext); ok {
		mvcCtx.Next()
		m.updateStats(config.AdvancedMode, time.Since(time.Now()))
		return nil
	}
	return fmt.Errorf("invalid context type for MVC mode")
}

func (m *UnifiedMiddlewareManager) executeInAutoMode(ctx interface{}) error {
	start := time.Now()
	
	// 根据当前性能决定使用哪种模式执行
	if m.shouldUseAdvancedMode() {
		err := m.executeInMVCMode(ctx)
		m.updateStats(config.AdvancedMode, time.Since(start))
		return err
	} else {
		err := m.executeInBasicMode(ctx)
		m.updateStats(config.BasicMode, time.Since(start))
		return err
	}
}

func (m *UnifiedMiddlewareManager) shouldUseAdvancedMode() bool {
	// 简单的启发式规则
	if m.stats.TotalRequests > int64(m.config.Auto.RequestThreshold) {
		return true
	}
	
	if m.stats.AverageResponseTime > m.config.Auto.ResponseTimeLimit {
		return true
	}
	
	return false
}

func (m *UnifiedMiddlewareManager) updateStats(mode config.MiddlewareMode, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.stats.TotalRequests++
	
	switch mode {
	case config.BasicMode:
		m.stats.BasicModeRequests++
	case config.AdvancedMode:
		m.stats.MVCModeRequests++
	}
	
	// 更新平均响应时间
	if m.stats.TotalRequests == 1 {
		m.stats.AverageResponseTime = duration
	} else {
		m.stats.AverageResponseTime = (m.stats.AverageResponseTime*time.Duration(m.stats.TotalRequests-1) + duration) / time.Duration(m.stats.TotalRequests)
	}
}

func (m *UnifiedMiddlewareManager) initBuiltinMiddlewares() {
	// 根据配置初始化内置中间件
	if m.config.Builtin.Logger.Enable {
		m.UseBuiltin("logger", nil)
	}
	
	if m.config.Builtin.Recovery.Enable {
		m.UseBuiltin("recovery", nil)
	}
	
	if m.config.Builtin.CORS.Enable {
		m.UseBuiltin("cors", m.config.Builtin.CORS)
	}
	
	if m.config.Builtin.RequestID.Enable {
		m.UseBuiltin("requestid", m.config.Builtin.RequestID)
	}
	
	// ... 其他内置中间件
}

// AutoModeSwitcher 方法实现

func (switcher *AutoModeSwitcher) Start() error {
	switcher.mu.Lock()
	defer switcher.mu.Unlock()
	
	if switcher.running {
		return fmt.Errorf("auto switcher already running")
	}
	
	switcher.running = true
	go switcher.run()
	return nil
}

func (switcher *AutoModeSwitcher) Stop() error {
	switcher.mu.Lock()
	defer switcher.mu.Unlock()
	
	if !switcher.running {
		return fmt.Errorf("auto switcher not running")
	}
	
	close(switcher.stopCh)
	switcher.running = false
	return nil
}

func (switcher *AutoModeSwitcher) run() {
	ticker := time.NewTicker(switcher.manager.config.Auto.SwitchCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-switcher.stopCh:
			return
		case <-ticker.C:
			switcher.checkAndSwitch()
		}
	}
}

func (switcher *AutoModeSwitcher) checkAndSwitch() {
	stats := switcher.manager.GetStats()
	
	// 检查是否需要升级到高级模式
	if switcher.manager.currentMode == config.BasicMode {
		if switcher.shouldUpgrade(stats) {
			switcher.manager.SwitchMode(config.AdvancedMode)
		}
	}
	
	// 检查是否需要降级到基础模式
	if switcher.manager.currentMode == config.AdvancedMode {
		if switcher.shouldDowngrade(stats) {
			switcher.manager.SwitchMode(config.BasicMode)
		}
	}
}

func (switcher *AutoModeSwitcher) shouldUpgrade(stats UnifiedStats) bool {
	config := switcher.manager.config
	
	// 请求量达到阈值
	if stats.TotalRequests > int64(config.Auto.RequestThreshold) {
		return config.Auto.EnableUpgrade
	}
	
	// 响应时间超过限制
	if stats.AverageResponseTime > config.Auto.ResponseTimeLimit {
		return config.Auto.EnableUpgrade
	}
	
	return false
}

func (switcher *AutoModeSwitcher) shouldDowngrade(stats UnifiedStats) bool {
	config := switcher.manager.config
	
	// 性能良好且请求量不高时降级
	if stats.TotalRequests < int64(config.Auto.RequestThreshold/2) &&
		stats.AverageResponseTime < config.Auto.ResponseTimeLimit/2 {
		return config.Auto.EnableDowngrade
	}
	
	return false
}

// 中间件选项

type MiddlewareOptions struct {
	Layer    *MiddlewareLayer
	Priority *int
	Config   interface{}
}

type MiddlewareOption func(*MiddlewareOptions)

func WithLayer(layer MiddlewareLayer) MiddlewareOption {
	return func(opts *MiddlewareOptions) {
		opts.Layer = &layer
	}
}

func WithPriority(priority int) MiddlewareOption {
	return func(opts *MiddlewareOptions) {
		opts.Priority = &priority
	}
}

func WithConfig(config interface{}) MiddlewareOption {
	return func(opts *MiddlewareOptions) {
		opts.Config = config
	}
}

func parseMiddlewareOptions(options ...MiddlewareOption) MiddlewareOptions {
	opts := MiddlewareOptions{}
	for _, option := range options {
		option(&opts)
	}
	return opts
}

// 全局统一管理器实例
var globalUnifiedManager *UnifiedMiddlewareManager

func init() {
	globalUnifiedManager = NewUnifiedMiddlewareManager()
}

// GetGlobalUnifiedManager 获取全局统一管理器
func GetGlobalUnifiedManager() *UnifiedMiddlewareManager {
	return globalUnifiedManager
}

// 全局便捷函数

// UseMiddleware 全局使用中间件
func UseMiddleware(name string, handler interface{}, options ...MiddlewareOption) error {
	return globalUnifiedManager.Use(name, handler, options...)
}

// UseBuiltinMiddleware 全局使用内置中间件
func UseBuiltinMiddleware(name string, config interface{}, options ...MiddlewareOption) error {
	return globalUnifiedManager.UseBuiltin(name, config, options...)
}

// SwitchMiddlewareMode 全局切换中间件模式
func SwitchMiddlewareMode(mode config.MiddlewareMode) error {
	return globalUnifiedManager.SwitchMode(mode)
}

// GetMiddlewareStats 获取全局中间件统计
func GetMiddlewareStats() UnifiedStats {
	return globalUnifiedManager.GetStats()
}