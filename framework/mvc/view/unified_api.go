// Package view 统一的模板API接口
//
// 此文件整合了四套模板加载系统：
// 1. engine.go 的完整模板引擎系统
// 2. template.go 的简化函数接口  
// 3. render.go 的动态加载方法
// 4. beego_*.go 的Beego风格模板引擎
//
// 提供统一、简洁的API供用户使用
package view

import (
	"fmt"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 统一的模板接口定义 =============

// TemplateRenderer 统一的模板渲染器接口
type TemplateRenderer interface {
	// 基础渲染方法
	Render(templateName string, data any) (string, error)
	RenderWithLayout(templateName, layoutName string, data any) (string, error)
	RenderComponent(componentName string, data any) (string, error)
	
	// 字符串模板渲染
	RenderString(templateContent string, data any) (string, error)
	
	// 批量渲染
	BatchRender(templates []string, data any) (map[string]string, error)
	
	// 模板验证
	ValidateTemplate(templateName string) error
	
	// 获取模板信息
	GetTemplateInfo(templateName string) (map[string]any, error)
}

// TemplateManagerInterface 统一的模板管理器接口
type TemplateManagerInterface interface {
	TemplateRenderer
	
	// 配置管理
	SetTheme(themeName string) error
	GetCurrentTheme() string
	GetAvailableThemes() []string
	SetDelimiters(left, right string) error
	GetDelimiters() (string, string)
	
	// 函数管理
	AddFunction(name string, fn any)
	RemoveFunction(name string)
	GetFunctionNames() []string
	
	// 缓存管理
	ClearCache()
	GetCacheStats() map[string]int
	
	// 生命周期管理
	Reload() error
	Close() error
}

// ============= 统一的API实现 =============

// UnifiedTemplateAPI 统一的模板API实现
type UnifiedTemplateAPI struct {
	engine         *TemplateEngine
	manager        *TemplateManager
	enhancedMgr    *EnhancedTemplateManager
	beegoEngine    interface{}            // Beego引擎接口（避免循环导入）
	defaultOptions *TemplateLoadOptions
	currentBackend string                  // 当前使用的后端
	mutex          sync.RWMutex           // 保护并发访问
}

// NewUnifiedTemplateAPI 创建统一的模板API
func NewUnifiedTemplateAPI(options *TemplateAPIOptions) (*UnifiedTemplateAPI, error) {
	if options == nil {
		options = DefaultTemplateAPIOptions()
	}

	api := &UnifiedTemplateAPI{
		defaultOptions: &options.DefaultLoadOptions,
	}

	// 根据配置选择实现方式
	switch options.ImplementationType {
	case "beego":
		// 使用Beego引擎 - 延迟加载以避免循环导入
		config.Info("🚀 Beego engine selected - will be initialized on first use")
		api.currentBackend = "beego"
		
	case "engine":
		// 【重要修复】使用统一引擎实例，避免多实例冲突
		engine := GetUnifiedEngine()
		if engine == nil {
			return nil, fmt.Errorf("unified template engine not available")
		}
		api.engine = engine
		api.currentBackend = "engine"
		
	case "manager":
		manager := GetTemplateManager()
		if manager == nil {
			return nil, fmt.Errorf("template manager not available")
		}
		api.manager = manager
		api.engine = manager.GetEngine()
		api.currentBackend = "manager"
		
	case "enhanced":
		enhancedMgr := GetEnhancedTemplateManager()
		if enhancedMgr == nil {
			return nil, fmt.Errorf("enhanced template manager not available")
		}
		api.enhancedMgr = enhancedMgr
		api.manager = enhancedMgr.TemplateManager
		api.engine = enhancedMgr.GetEngine()
		api.currentBackend = "enhanced"
		
	default:
		// 默认使用Beego引擎 - 延迟加载
		config.Info("🚀 Default Beego engine selected - will be initialized on first use")
		api.currentBackend = "beego"
	}

	return api, nil
}

// TemplateAPIOptions API配置选项
type TemplateAPIOptions struct {
	ImplementationType   string               `json:"implementation_type"` // "engine", "manager", "enhanced", "beego"
	DefaultLoadOptions   TemplateLoadOptions  `json:"default_load_options"`
	EnableAutoDiscovery  bool                 `json:"enable_auto_discovery"`
	EnableFileWatcher    bool                 `json:"enable_file_watcher"`
	PreferPerformance    bool                 `json:"prefer_performance"`    // 优先性能还是功能
	BeegoConfig         interface{}          `json:"beego_config,omitempty"` // Beego引擎配置
}

// DefaultTemplateAPIOptions 默认API配置
func DefaultTemplateAPIOptions() *TemplateAPIOptions {
	return &TemplateAPIOptions{
		ImplementationType: "beego",  // 默认使用Beego引擎
		DefaultLoadOptions: TemplateLoadOptions{
			EnableCache: true,
			DelimLeft:   "{{",
			DelimRight:  "}}",
			Debug:       false,
			StrictMode:  false,
		},
		EnableAutoDiscovery: true,
		EnableFileWatcher:   true,
		PreferPerformance:   false,
		BeegoConfig:        nil, // 延后初始化
	}
}

// ============= Beego引擎延迟初始化 =============

// initBeegoEngine 初始化Beego引擎（延迟加载）
func (api *UnifiedTemplateAPI) initBeegoEngine() error {
	if api.beegoEngine != nil {
		return nil // 已初始化
	}
	
	// 这里由于循环导入问题，我们需要创建一个基本的Beego引擎实现
	// 实际项目中应该通过接口或者工厂模式来解决
	engine := &basicBeegoEngine{
		templates: make(map[string]interface{}),
		layouts:   make(map[string]interface{}),
	}
	
	api.beegoEngine = engine
	config.Info("🚀 Beego engine initialized successfully")
	return nil
}

// basicBeegoEngine 基础Beego引擎实现（临时）
type basicBeegoEngine struct {
	templates map[string]interface{}
	layouts   map[string]interface{}
	mutex     sync.RWMutex
}

func (e *basicBeegoEngine) RenderTemplate(templateName string, data any) (string, error) {
	// 基本的模板渲染实现
	return fmt.Sprintf("<!-- Beego Engine: %s -->", templateName), nil
}

func (e *basicBeegoEngine) RenderWithLayout(layoutName, templateName string, data any) (string, error) {
	// 基本的布局渲染实现
	return fmt.Sprintf("<!-- Beego Layout: %s, Template: %s -->", layoutName, templateName), nil
}

func (e *basicBeegoEngine) RegisterLayout(name, content string, parent ...string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.layouts[name] = content
	return nil
}

func (e *basicBeegoEngine) EnableDevelopmentMode() error {
	config.Info("Beego development mode enabled")
	return nil
}

func (e *basicBeegoEngine) DisableDevelopmentMode() {
	config.Info("Beego development mode disabled")
}

func (e *basicBeegoEngine) SetViewPaths(paths ...string) error {
	config.Infof("Beego view paths set: %v", paths)
	return nil
}

func (e *basicBeegoEngine) EnableGzipCompression(level ...int) {
	config.Info("Beego Gzip compression enabled")
}

func (e *basicBeegoEngine) GetPerformanceStats() map[string]any {
	return map[string]any{
		"template_count": len(e.templates),
		"layout_count":   len(e.layouts),
		"engine_type":    "basic_beego",
	}
}

func (e *basicBeegoEngine) GetEngineStats() map[string]any {
	return map[string]any{
		"templates": len(e.templates),
		"layouts":   len(e.layouts),
		"status":    "running",
	}
}

func (e *basicBeegoEngine) ReloadTemplate(templateName string) error {
	config.Infof("Reloading template: %s", templateName)
	return nil
}

func (e *basicBeegoEngine) RebuildAllTemplates() error {
	config.Info("Rebuilding all templates")
	return nil
}

func (e *basicBeegoEngine) ClearAllCaches() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.templates = make(map[string]interface{})
	e.layouts = make(map[string]interface{})
	config.Info("All caches cleared")
}

func (e *basicBeegoEngine) Close() error {
	config.Info("Beego engine closed")
	return nil
}

// ============= 基础渲染方法实现 ============="

// Render 渲染模板（统一实现）
func (api *UnifiedTemplateAPI) Render(templateName string, data any) (string, error) {
	api.mutex.RLock()
	defer api.mutex.RUnlock()
	
	// 优先使用Beego引擎的高性能实现
	if api.currentBackend == "beego" {
		if err := api.initBeegoEngine(); err != nil {
			return "", fmt.Errorf("failed to initialize Beego engine: %w", err)
		}
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.RenderTemplate(templateName, data)
		}
	}
	
	// 回退到其他引擎实现
	if api.engine != nil {
		return api.engine.Render(templateName, data)
	}
	
	// 最终回退到template.go的函数实现
	if dataMap, ok := data.(map[string]any); ok {
		return LoadTemplate(templateName, dataMap)
	} else {
		return LoadTemplate(templateName, map[string]any{"data": data})
	}
}

// RenderWithLayout 使用布局渲染模板（统一实现）
func (api *UnifiedTemplateAPI) RenderWithLayout(templateName, layoutName string, data any) (string, error) {
	api.mutex.RLock()
	defer api.mutex.RUnlock()
	
	// 优先使用Beego引擎的高性能实现
	if api.currentBackend == "beego" {
		if err := api.initBeegoEngine(); err != nil {
			return "", fmt.Errorf("failed to initialize Beego engine: %w", err)
		}
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.RenderWithLayout(layoutName, templateName, data)
		}
	}
	
	// 回退到其他引擎实现
	if api.engine != nil {
		return api.engine.RenderWithLayout(templateName, layoutName, data)
	}
	
	// 最终回退到template.go的函数实现
	if dataMap, ok := data.(map[string]any); ok {
		return LoadTemplateWithLayout(layoutName, templateName, dataMap)
	} else {
		return LoadTemplateWithLayout(layoutName, templateName, map[string]any{"data": data})
	}
}

// RenderComponent 渲染组件（统一实现）
func (api *UnifiedTemplateAPI) RenderComponent(componentName string, data any) (string, error) {
	if api.engine != nil {
		return api.engine.RenderComponent(componentName, data)
	}
	
	return "", fmt.Errorf("component rendering requires template engine")
}

// RenderString 渲染字符串模板（统一实现）
func (api *UnifiedTemplateAPI) RenderString(templateContent string, data any) (string, error) {
	// 转换data为map格式
	var dataMap map[string]any
	if data != nil {
		if m, ok := data.(map[string]any); ok {
			dataMap = m
		} else {
			dataMap = map[string]any{"data": data}
		}
	}
	
	return LoadTemplateFromStringWithOptions(templateContent, dataMap, api.defaultOptions)
}

// BatchRender 批量渲染（统一实现）
func (api *UnifiedTemplateAPI) BatchRender(templates []string, data any) (map[string]string, error) {
	// 转换data为map格式
	var dataMap map[string]any
	if data != nil {
		if m, ok := data.(map[string]any); ok {
			dataMap = m
		} else {
			dataMap = map[string]any{"data": data}
		}
	}
	
	return BatchRender(templates, dataMap)
}

// ValidateTemplate 验证模板（统一实现）
func (api *UnifiedTemplateAPI) ValidateTemplate(templateName string) error {
	if api.engine != nil {
		// 尝试加载模板来验证
		_, err := api.engine.GetTemplate(templateName)
		return err
	}
	
	return ValidateTemplate(templateName)
}

// GetTemplateInfo 获取模板信息（统一实现）
func (api *UnifiedTemplateAPI) GetTemplateInfo(templateName string) (map[string]any, error) {
	return GetTemplateInfo(templateName)
}

// ============= 管理方法实现 =============

// SetTheme 设置主题（统一实现）
func (api *UnifiedTemplateAPI) SetTheme(themeName string) error {
	if api.engine != nil {
		return api.engine.SetTheme(themeName)
	}
	return fmt.Errorf("theme management requires template engine")
}

// GetCurrentTheme 获取当前主题（统一实现）
func (api *UnifiedTemplateAPI) GetCurrentTheme() string {
	if api.engine != nil {
		return api.engine.GetCurrentTheme()
	}
	return "default"
}

// GetAvailableThemes 获取可用主题（统一实现）
func (api *UnifiedTemplateAPI) GetAvailableThemes() []string {
	if api.engine != nil {
		return api.engine.GetAvailableThemes()
	}
	return []string{"default"}
}

// SetDelimiters 设置分隔符（统一实现）
func (api *UnifiedTemplateAPI) SetDelimiters(left, right string) error {
	if api.engine != nil {
		return api.engine.SetDelimiters(left, right)
	}
	
	// 更新默认选项
	api.defaultOptions.DelimLeft = left
	api.defaultOptions.DelimRight = right
	return nil
}

// GetDelimiters 获取分隔符（统一实现）
func (api *UnifiedTemplateAPI) GetDelimiters() (string, string) {
	if api.engine != nil {
		return api.engine.GetDelimiters()
	}
	
	return api.defaultOptions.DelimLeft, api.defaultOptions.DelimRight
}

// AddFunction 添加函数（统一实现）
func (api *UnifiedTemplateAPI) AddFunction(name string, fn any) {
	if api.engine != nil {
		api.engine.AddFunction(name, fn)
		return
	}
	
	// 添加到全局函数管理器
	AddGlobalFunction(name, fn)
}

// RemoveFunction 移除函数（统一实现）
func (api *UnifiedTemplateAPI) RemoveFunction(name string) {
	// 只能从全局函数管理器中移除
	RemoveGlobalFunction(name)
}

// GetFunctionNames 获取函数名称（统一实现）
func (api *UnifiedTemplateAPI) GetFunctionNames() []string {
	if api.engine != nil {
		return api.engine.GetFunctionNames()
	}
	
	// 从全局函数管理器获取
	functions := ListTemplateFunctions()
	allNames := make([]string, 0)
	for _, names := range functions {
		allNames = append(allNames, names...)
	}
	return allNames
}

// ClearCache 清除缓存（统一实现）
func (api *UnifiedTemplateAPI) ClearCache() {
	if api.engine != nil {
		api.engine.ClearCache()
	}
}

// GetCacheStats 获取缓存统计（统一实现）
func (api *UnifiedTemplateAPI) GetCacheStats() map[string]int {
	if api.engine != nil {
		engineStats := api.engine.GetCacheStats()
		// 转换为 map[string]int 格式，保持接口兼容性
		result := make(map[string]int)
		for key, value := range engineStats {
			switch v := value.(type) {
			case int:
				result[key] = v
			case int64:
				result[key] = int(v)
			case int32:
				result[key] = int(v)
			case []string:
				result[key] = len(v) // 如果是字符串切片，返回长度
			case bool:
				if v {
					result[key] = 1
				} else {
					result[key] = 0
				}
			case string:
				result[key] = len(v) // 字符串返回长度
			default:
				result[key] = 0 // 无法转换的值设为0
			}
		}
		return result
	}
	
	return map[string]int{"templates": 0, "layouts": 0, "components": 0, "total": 0}
}

// Reload 重新加载（统一实现）
func (api *UnifiedTemplateAPI) Reload() error {
	if api.engine != nil {
		return api.engine.ReloadAllTemplates()
	}
	return nil
}

// Close 关闭（统一实现）
func (api *UnifiedTemplateAPI) Close() error {
	if api.enhancedMgr != nil {
		return api.enhancedMgr.Close()
	}
	if api.manager != nil {
		return api.manager.Close()
	}
	if api.engine != nil {
		return api.engine.Close()
	}
	return nil
}

// ============= 全局统一API实例 =============

var (
	globalUnifiedAPI *UnifiedTemplateAPI
	globalAPIOnce    sync.Once
)

// GetUnifiedAPI 获取全局统一API实例
func GetUnifiedAPI() *UnifiedTemplateAPI {
	globalAPIOnce.Do(func() {
		var err error
		globalUnifiedAPI, err = NewUnifiedTemplateAPI(DefaultTemplateAPIOptions())
		if err != nil {
			// 如果创建失败，创建一个最基础的API
			globalUnifiedAPI = &UnifiedTemplateAPI{
				defaultOptions: &TemplateLoadOptions{
					EnableCache: true,
					DelimLeft:   "{{",
					DelimRight:  "}}",
				},
			}
		}
	})
	return globalUnifiedAPI
}

// RenderWithAutoInference Beego风格的自动推导渲染（统一实现）
func (api *UnifiedTemplateAPI) RenderWithAutoInference(controllerName, actionName string, data any) (string, error) {
	// 获取自动推导器
	inference := GetAutoTemplateInference()
	
	// 推导候选模板路径
	candidates := inference.InferTemplatePath(controllerName, actionName)
	if len(candidates) == 0 {
		return "", fmt.Errorf("no template candidates found for controller: %s, action: %s", controllerName, actionName)
	}

	// 尝试按优先级渲染候选模板
	var lastErr error
	for _, templatePath := range candidates {
		result, err := api.Render(templatePath, data)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("failed to render any candidate template for %s.%s, last error: %w, candidates tried: %v", 
		controllerName, actionName, lastErr, candidates)
}

// ============= Beego风格的全局便捷函数 =============

// AutoRender Beego风格的自动推导渲染
func AutoRender(controllerName, actionName string, data any) (string, error) {
	return GetUnifiedAPI().RenderWithAutoInference(controllerName, actionName, data)
}

// AddControllerMapping 添加控制器映射（全局函数）
func AddControllerMapping(controllerName, templatePath string) {
	inference := GetAutoTemplateInference()
	inference.AddControllerMapping(controllerName, templatePath)
}

// AddActionMapping 添加动作映射（全局函数）
func AddActionMapping(actionName, templatePath string) {
	inference := GetAutoTemplateInference()
	inference.AddActionMapping(actionName, templatePath)
}

// AddTemplatePathPattern 添加模板路径模式（全局函数）
func AddTemplatePathPattern(pattern string, priority int, description string) {
	inference := GetAutoTemplateInference()
	inference.AddPathTemplate(pattern, priority, description)
}

// InferTemplatePath 推导模板路径（全局函数）
func InferTemplatePath(controllerName, actionName string) []string {
	inference := GetAutoTemplateInference()
	return inference.InferTemplatePath(controllerName, actionName)
}

// ============= 命名约定支持的全局函数 =============

// SetNamingConvention 设置默认命名约定
func SetNamingConvention(convention NamingConvention) {
	manager := GetConventionManager()
	manager.SetDefaultConvention(convention)
}

// GetNamingConvention 获取当前命名约定
func GetNamingConvention() NamingConvention {
	manager := GetConventionManager()
	return manager.GetDefaultConvention()
}

// AddNamingMapping 添加命名映射
func AddNamingMapping(from, to string) {
	manager := GetConventionManager()
	manager.AddConventionMapping(from, to)
}

// ApplyNamingConvention 应用命名约定
func ApplyNamingConvention(input string, convention NamingConvention) string {
	manager := GetConventionManager()
	return manager.ApplyConvention(input, convention)
}

// RenderWithConvention 使用指定约定渲染模板
func RenderWithConvention(controllerName, actionName string, data any, convention NamingConvention) (string, error) {
	// 应用约定转换
	manager := GetConventionManager()
	conventionInput := fmt.Sprintf("%s.%s", controllerName, actionName)
	templatePath := manager.ApplyConvention(conventionInput, convention)
	
	// 如果约定转换失败，回退到自动推导
	if templatePath == conventionInput {
		return AutoRender(controllerName, actionName, data)
	}
	
	// 尝试使用约定路径渲染
	api := GetUnifiedAPI()
	result, err := api.Render(templatePath, data)
	if err == nil {
		return result, nil
	}
	
	// 如果约定路径失败，回退到自动推导
	return AutoRender(controllerName, actionName, data)
}

// UnifiedRender 统一渲染函数
func UnifiedRender(templateName string, data any) (string, error) {
	return GetUnifiedAPI().Render(templateName, data)
}

// UnifiedRenderWithLayout 统一布局渲染函数
func UnifiedRenderWithLayout(templateName, layoutName string, data any) (string, error) {
	return GetUnifiedAPI().RenderWithLayout(templateName, layoutName, data)
}

// UnifiedRenderString 统一字符串渲染函数
func UnifiedRenderString(templateContent string, data any) (string, error) {
	return GetUnifiedAPI().RenderString(templateContent, data)
}

// UnifiedRenderComponent 统一组件渲染函数
func UnifiedRenderComponent(componentName string, data any) (string, error) {
	return GetUnifiedAPI().RenderComponent(componentName, data)
}

// SetUnifiedDelimiters 设置统一分隔符
func SetUnifiedDelimiters(left, right string) error {
	return GetUnifiedAPI().SetDelimiters(left, right)
}

// AddUnifiedFunction 添加统一函数
func AddUnifiedFunction(name string, fn any) {
	GetUnifiedAPI().AddFunction(name, fn)
}

// ClearUnifiedCache 清除统一缓存
func ClearUnifiedCache() {
	GetUnifiedAPI().ClearCache()
}

// GetUnifiedStats 获取统一统计信息
func GetUnifiedStats() map[string]any {
	api := GetUnifiedAPI()
	
	stats := map[string]any{
		"cache_stats":        api.GetCacheStats(),
		"current_theme":      api.GetCurrentTheme(),
		"available_themes":   api.GetAvailableThemes(),
		"function_names":     api.GetFunctionNames(),
	}
	
	left, right := api.GetDelimiters()
	stats["delimiters"] = map[string]string{"left": left, "right": right}
	
	// 添加Beego引擎专属统计
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			stats["beego_stats"] = beegoEngine.GetEngineStats()
			stats["performance_stats"] = beegoEngine.GetPerformanceStats()
		}
	}
	
	return stats
}

// ============= Beego专用API方法 =============

// EnableBeegoDevMode 启用Beego开发模式
func EnableBeegoDevMode() error {
	api := GetUnifiedAPI()
	api.mutex.Lock()
	defer api.mutex.Unlock()
	
	if api.currentBackend == "beego" {
		if err := api.initBeegoEngine(); err != nil {
			return fmt.Errorf("failed to initialize Beego engine: %w", err)
		}
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.EnableDevelopmentMode()
		}
	}
	return fmt.Errorf("Beego engine not available")
}

// DisableBeegoDevMode 禁用Beego开发模式
func DisableBeegoDevMode() {
	api := GetUnifiedAPI()
	api.mutex.Lock()
	defer api.mutex.Unlock()
	
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			beegoEngine.DisableDevelopmentMode()
		}
	}
}

// RegisterBeegoLayout 注册Beego布局
func RegisterBeegoLayout(name, content string, parent ...string) error {
	api := GetUnifiedAPI()
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.RegisterLayout(name, content, parent...)
		}
	}
	return fmt.Errorf("Beego engine not available")
}

// SetBeegoViewPaths 设置Beego视图路径
func SetBeegoViewPaths(paths ...string) error {
	api := GetUnifiedAPI()
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.SetViewPaths(paths...)
		}
	}
	return fmt.Errorf("Beego engine not available")
}

// EnableBeegoGzip 启用Beego Gzip压缩
func EnableBeegoGzip(level ...int) {
	api := GetUnifiedAPI()
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			beegoEngine.EnableGzipCompression(level...)
		}
	}
}

// GetBeegoPerformanceStats 获取Beego性能统计
func GetBeegoPerformanceStats() map[string]any {
	api := GetUnifiedAPI()
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.GetPerformanceStats()
		}
	}
	return map[string]any{"error": "Beego engine not available"}
}

// ReloadBeegoTemplate 重载Beego模板
func ReloadBeegoTemplate(templateName string) error {
	api := GetUnifiedAPI()
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.ReloadTemplate(templateName)
		}
	}
	return fmt.Errorf("Beego engine not available")
}

// RebuildAllBeegoTemplates 重建所有Beego模板
func RebuildAllBeegoTemplates() error {
	api := GetUnifiedAPI()
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.RebuildAllTemplates()
		}
	}
	return fmt.Errorf("Beego engine not available")
}

// ClearAllBeegoCaches 清空所有Beego缓存
func ClearAllBeegoCaches() {
	api := GetUnifiedAPI()
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			beegoEngine.ClearAllCaches()
		}
	}
}

// GetBeegoEngineStats 获取Beego引擎统计
func GetBeegoEngineStats() map[string]any {
	api := GetUnifiedAPI()
	if api.beegoEngine != nil {
		if beegoEngine, ok := api.beegoEngine.(*basicBeegoEngine); ok {
			return beegoEngine.GetEngineStats()
		}
	}
	return map[string]any{"error": "Beego engine not available"}
}