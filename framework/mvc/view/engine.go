// Package view 提供增强的模板引擎功能
//
// 这个包提供了类似Beego的模板引擎功能，包括：
// - 布局继承和组件系统
// - 模板热重载
// - 丰富的模板函数
// - 模板缓存管理
// - 多主题支持
package view

import (
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/zsy619/yyhertz/framework/config"
)

// TemplateEngine 模板引擎 (合并了 BeegoTemplateEngine 功能)
type TemplateEngine struct {
	// 核心配置
	viewPaths     []string // 模板搜索路径
	layoutPath    string   // 布局文件路径
	componentPath string   // 组件文件路径
	extension     string   // 模板文件扩展名
	delimLeft     string   // 左分隔符
	delimRight    string   // 右分隔符

	// 缓存和管理
	templates     map[string]*template.Template // 模板缓存
	layouts       map[string]*template.Template // 布局缓存
	components    map[string]*template.Template // 组件缓存
	templateMutex sync.RWMutex                  // 模板缓存锁

	// 功能开关
	enableCache    bool // 启用模板缓存
	enableReload   bool // 启用热重载
	enableCompress bool // 启用压缩

	// 热重载
	watcher    *fsnotify.Watcher // 文件监控器
	watchPaths map[string]bool   // 监控路径

	// 模板函数
	funcMap template.FuncMap // 模板函数映射

	// 主题支持
	currentTheme string                         // 当前主题
	themes       map[string]*config.ThemeConfig // 主题配置

	// Beego 兼容功能 (从 BeegoTemplateEngine 合并)
	RunMode          string   // 运行模式 (dev/prod)
	EnableGzip       bool     // 是否启用Gzip压缩
	DirectoryIndex   bool     // 是否显示目录索引
	EnableAutoRender bool     // 是否启用自动渲染
	AutoRender       bool     // 是否自动渲染
	TemplateExt      []string // 支持的模板扩展名

	// Beego 高级功能
	templateDirs map[string]os.FileInfo // 目录信息缓存
	lastModTime  map[string]time.Time   // 文件修改时间缓存
	watchedFiles map[string]bool        // 监控的文件列表

	// 高级功能组件 (从 BeegoAdvancedFeatures 合并)
	performanceStats *PerformanceStats     // 性能统计
	cacheOptimizer   *CacheOptimizer       // 缓存优化器
	compressionMgr   *CompressionManager   // 压缩管理器
	errorHandler     *TemplateErrorHandler // 错误处理器
}

// 使用 config 包中的配置结构体
// ThemeConfig 和 TemplateConfig 已迁移到 config 包

// DefaultTemplateConfig 默认模板配置（向后兼容）
// 该函数返回 config 包中的默认配置
func DefaultTemplateConfig() *config.TemplateConfig {
	return config.GlobalTemplate
}

// TemplateFunctionManager 统一的模板函数管理器
type TemplateFunctionManager struct {
	builtinFuncs template.FuncMap // 内置函数（原 TemplateFuncs）
	globalFuncs  template.FuncMap // 全局自定义函数
	mu           sync.RWMutex
}

// FunctionSource 函数来源类型
type FunctionSource string

const (
	FunctionSourceBuiltin FunctionSource = "builtin"
	FunctionSourceGlobal  FunctionSource = "global"
	FunctionSourceEngine  FunctionSource = "engine"
)

var (
	defaultEngine *TemplateEngine
	engineOnce    sync.Once
	engineMutex   sync.Mutex

	// 全局模板函数管理器
	globalFunctionManager     *TemplateFunctionManager
	globalFunctionManagerOnce sync.Once

	// 全局引擎统一管理
	unifiedEngineInstance *TemplateEngine
	unifiedEngineOnce     sync.Once
	unifiedEngineMutex    sync.RWMutex
)

// NewTemplateFunctionManager 创建新的模板函数管理器
func NewTemplateFunctionManager() *TemplateFunctionManager {
	return &TemplateFunctionManager{
		builtinFuncs: make(template.FuncMap),
		globalFuncs:  make(template.FuncMap),
	}
}

// GetGlobalFunctionManager 获取全局函数管理器
func GetGlobalFunctionManager() *TemplateFunctionManager {
	globalFunctionManagerOnce.Do(func() {
		globalFunctionManager = NewTemplateFunctionManager()
		// 初始化内置函数
		globalFunctionManager.initBuiltinFunctions()
	})
	return globalFunctionManager
}

// initBuiltinFunctions 初始化内置函数
func (tfm *TemplateFunctionManager) initBuiltinFunctions() {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	// 从 template.go 获取内置函数
	builtinFuncs := GetBuiltinTemplateFunctions()
	maps.Copy(tfm.builtinFuncs, builtinFuncs)

	// 添加 Beego 风格的模板函数
	beegoFuncs := GetBeegoTemplateFuncs()
	maps.Copy(tfm.builtinFuncs, beegoFuncs)
}

// AddGlobalFunction 添加全局函数
func (tfm *TemplateFunctionManager) AddGlobalFunction(name string, fn any) {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	if tfm.globalFuncs == nil {
		tfm.globalFuncs = make(template.FuncMap)
	}

	tfm.globalFuncs[name] = fn
	config.Debugf("Global template function added: %s", name)
}

// RemoveGlobalFunction 移除全局函数
func (tfm *TemplateFunctionManager) RemoveGlobalFunction(name string) {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	delete(tfm.globalFuncs, name)
	config.Debugf("Global template function removed: %s", name)
}

// GetMergedFunctions 获取合并后的函数映射
// 优先级: 引擎自定义 > 全局自定义 > 内置函数
func (tfm *TemplateFunctionManager) GetMergedFunctions(engineFuncs template.FuncMap) template.FuncMap {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	merged := make(template.FuncMap)

	// 1. 首先添加内置函数
	maps.Copy(merged, tfm.builtinFuncs)

	// 2. 然后添加全局自定义函数（可能覆盖内置函数）
	for name := range tfm.globalFuncs {
		if _, exists := merged[name]; exists {
			config.Warnf("Global function '%s' overrides builtin function", name)
		}
	}
	maps.Copy(merged, tfm.globalFuncs)

	// 3. 最后添加引擎自定义函数（可能覆盖前面的函数）
	for name, fn := range engineFuncs {
		if _, exists := merged[name]; exists {
			source := tfm.GetFunctionSource(name)
			config.Warnf("Engine function '%s' overrides %s function", name, source)
		}
		merged[name] = fn
	}

	return merged
}

// GetFunctionSource 获取函数来源
func (tfm *TemplateFunctionManager) GetFunctionSource(name string) FunctionSource {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	if _, exists := tfm.builtinFuncs[name]; exists {
		return FunctionSourceBuiltin
	}
	if _, exists := tfm.globalFuncs[name]; exists {
		return FunctionSourceGlobal
	}
	return FunctionSourceEngine
}

// HasFunction 检查函数是否存在
func (tfm *TemplateFunctionManager) HasFunction(name string) bool {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	_, builtin := tfm.builtinFuncs[name]
	_, global := tfm.globalFuncs[name]
	return builtin || global
}

// ListFunctions 列出所有函数按来源分类
func (tfm *TemplateFunctionManager) ListFunctions() map[string][]string {
	tfm.mu.RLock()
	defer tfm.mu.RUnlock()

	result := map[string][]string{
		"builtin": make([]string, 0, len(tfm.builtinFuncs)),
		"global":  make([]string, 0, len(tfm.globalFuncs)),
	}

	for name := range tfm.builtinFuncs {
		result["builtin"] = append(result["builtin"], name)
	}

	for name := range tfm.globalFuncs {
		result["global"] = append(result["global"], name)
	}

	return result
}

// RegisterFunctionGroup 批量注册函数组
func (tfm *TemplateFunctionManager) RegisterFunctionGroup(groupName string, funcs template.FuncMap) {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	if tfm.globalFuncs == nil {
		tfm.globalFuncs = make(template.FuncMap)
	}

	for name, fn := range funcs {
		tfm.globalFuncs[name] = fn
		config.Debugf("Function '%s' registered in group '%s'", name, groupName)
	}
}

// ============= 向后兼容的全局函数接口 =============

// AddGlobalFunction 添加全局模板函数（向后兼容接口）
func AddGlobalFunction(name string, fn any) {
	manager := GetGlobalFunctionManager()
	manager.AddGlobalFunction(name, fn)
}

// GetGlobalFunctions 获取全局模板函数映射（向后兼容接口）
func GetGlobalFunctions() template.FuncMap {
	manager := GetGlobalFunctionManager()
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	// 创建副本以避免并发修改
	funcMapCopy := make(template.FuncMap, len(manager.globalFuncs))
	maps.Copy(funcMapCopy, manager.globalFuncs)

	return funcMapCopy
}

// RemoveGlobalFunction 移除全局模板函数（向后兼容接口）
func RemoveGlobalFunction(name string) {
	manager := GetGlobalFunctionManager()
	manager.RemoveGlobalFunction(name)
}

// ============= 新的统一管理接口 =============

// GetTemplateFunctionManager 获取模板函数管理器
func GetTemplateFunctionManager() *TemplateFunctionManager {
	return GetGlobalFunctionManager()
}

// RegisterTemplateFunctionGroup 注册模板函数组
func RegisterTemplateFunctionGroup(groupName string, funcs template.FuncMap) {
	manager := GetGlobalFunctionManager()
	manager.RegisterFunctionGroup(groupName, funcs)
}

// ListTemplateFunctions 列出所有模板函数
func ListTemplateFunctions() map[string][]string {
	manager := GetGlobalFunctionManager()
	return manager.ListFunctions()
}

// HasTemplateFunction 检查模板函数是否存在
func HasTemplateFunction(name string) bool {
	manager := GetGlobalFunctionManager()
	return manager.HasFunction(name)
}

// GetTemplateFunctionSource 获取模板函数来源
func GetTemplateFunctionSource(name string) FunctionSource {
	manager := GetGlobalFunctionManager()
	return manager.GetFunctionSource(name)
}

// ============= 高级功能组件结构体定义 =============

// PerformanceStats 性能统计
type PerformanceStats struct {
	mutex             sync.RWMutex
	RenderCount       int64                    // 渲染次数
	TotalRenderTime   time.Duration            // 总渲染时间
	AverageRenderTime time.Duration            // 平均渲染时间
	CacheHitCount     int64                    // 缓存命中次数
	CacheMissCount    int64                    // 缓存未命中次数
	TemplateErrors    int64                    // 模板错误次数
	LastRenderTime    time.Time                // 最后渲染时间
	RenderTimes       map[string]time.Duration // 各模板渲染时间
	HotTemplates      map[string]int64         // 热点模板统计
}

// CacheOptimizer 缓存优化器
type CacheOptimizer struct {
	mutex            sync.RWMutex
	precompiledCache map[string]*template.Template // 预编译缓存
	frequencyTracker map[string]int64              // 访问频率追踪
	lastAccessTime   map[string]time.Time          // 最后访问时间
	maxCacheSize     int                           // 最大缓存大小
	cleanupInterval  time.Duration                 // 清理间隔
	cleanupTicker    *time.Ticker                  // 清理定时器
}

// CompressionManager 压缩管理器
type CompressionManager struct {
	mutex             sync.RWMutex
	enableGzip        bool              // 是否启用Gzip
	gzipLevel         int               // Gzip压缩级别
	compressionCache  map[string][]byte // 压缩结果缓存
	minSizeToCompress int               // 最小压缩大小
}

// TemplateErrorHandler 模板错误处理器
type TemplateErrorHandler struct {
	mutex         sync.RWMutex
	showDebugInfo bool               // 是否显示调试信息
	errorTemplate *template.Template // 错误模板
	errorLog      []TemplateError    // 错误日志
	maxErrorLog   int                // 最大错误日志数
}

// TemplateError 模板错误
type TemplateError struct {
	Timestamp   time.Time `json:"timestamp"`
	Template    string    `json:"template"`
	Error       string    `json:"error"`
	Stack       string    `json:"stack,omitempty"`
	RequestPath string    `json:"request_path,omitempty"`
}

func (e *TemplateEngine) GetTemplateCount() int {
	return len(e.templates)
}

// ============= 模板函数管理器的引擎方法扩展 =============

// GetFunctionNames 获取引擎中所有函数名称
func (e *TemplateEngine) GetFunctionNames() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	names := make([]string, 0, len(e.funcMap))
	for name := range e.funcMap {
		names = append(names, name)
	}
	return names
}

// HasEngineFunction 检查引擎是否有指定函数
func (e *TemplateEngine) HasEngineFunction(name string) bool {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	_, exists := e.funcMap[name]
	return exists
}

// GetEngineFunctionCount 获取引擎中函数数量
func (e *TemplateEngine) GetEngineFunctionCount() int {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	return len(e.funcMap)
}

// GetDefaultEngine 获取默认模板引擎（统一管理）
func GetDefaultEngine() *TemplateEngine {
	engineOnce.Do(func() {
		engineMutex.Lock()
		defer engineMutex.Unlock()

		// 【重要修复】直接使用统一引擎实例管理
		defaultEngine = GetUnifiedEngine()
		config.Infof("✅ Default engine now references unified engine instance")
	})
	return defaultEngine
}

// GetUnifiedEngine 获取统一的模板引擎实例
// 这是所有组件应该使用的统一入口，确保全局只有一个引擎实例
func GetUnifiedEngine() *TemplateEngine {
	unifiedEngineOnce.Do(func() {
		unifiedEngineMutex.Lock()
		defer unifiedEngineMutex.Unlock()

		// 1. 优先从全局管理器获取引擎
		if templateManager != nil {
			if engine := templateManager.GetEngine(); engine != nil {
				unifiedEngineInstance = engine
				config.Infof("✅ Using unified engine from template manager")
				return
			}
		}

		// 2. 其次从增强管理器获取引擎
		if enhancedTemplateManager != nil {
			if engine := enhancedTemplateManager.GetEngine(); engine != nil {
				unifiedEngineInstance = engine
				config.Infof("✅ Using unified engine from enhanced manager")
				return
			}
		}

		// 3. 如果已有默认引擎，直接使用
		if defaultEngine != nil {
			unifiedEngineInstance = defaultEngine
			config.Infof("✅ Using existing default engine as unified instance")
			return
		}

		// 4. 最后创建新的引擎实例
		// 🔧 确保使用最新的配置 - 重新加载配置文件
		cfg, err := loadCurrentTemplateConfig()
		if err != nil {
			config.Warnf("⚠️ Failed to load current template config, using default: %v", err)
			cfg = config.GlobalTemplate
			if cfg == nil {
				cfg = DefaultTemplateConfig()
			}
		}

		// 📋 输出配置信息供调试
		config.Infof("🛠️ Creating unified template engine with config:")
		config.Infof("   ViewPaths: %v", cfg.Paths.ViewPaths)
		config.Infof("   LayoutPath: %s", cfg.Paths.LayoutPath)
		config.Infof("   ComponentPath: %s", cfg.Paths.ComponentPath)
		config.Infof("   EnableCache: %v", cfg.Cache.EnableCache)

		unifiedEngineInstance, err = NewTemplateEngine(cfg)
		if err != nil {
			config.Fatalf("Failed to create unified template engine: %v", err)
		}

		// 设置为默认引擎，保持向后兼容
		defaultEngine = unifiedEngineInstance
		SetGlobalEngineForFunctions(unifiedEngineInstance)
		config.Infof("✅ Created new unified template engine instance")
	})

	return unifiedEngineInstance
}

// SetUnifiedEngine 设置统一的模板引擎实例
// 允许外部设置引擎实例，确保全局统一
func SetUnifiedEngine(engine *TemplateEngine) {
	unifiedEngineMutex.Lock()
	defer unifiedEngineMutex.Unlock()

	if engine == nil {
		config.Warnf("⚠️ Attempting to set nil unified engine")
		return
	}

	unifiedEngineInstance = engine
	defaultEngine = engine // 保持向后兼容
	SetGlobalEngineForFunctions(engine)
	config.Infof("✅ Unified engine instance updated")
}

// EnsureUnifiedEngineInitialization 确保统一引擎已初始化
// 这个函数应该在系统启动时调用
func EnsureUnifiedEngineInitialization() error {
	engine := GetUnifiedEngine()
	if engine == nil {
		return fmt.Errorf("failed to initialize unified template engine")
	}

	// 验证引擎功能
	if err := engine.HealthCheck(); err != nil {
		config.Warnf("Unified engine health check failed: %v", err)
		// 尝试恢复
		if recoverErr := engine.Recovery(); recoverErr != nil {
			return fmt.Errorf("failed to recover unified engine: %w", recoverErr)
		}
	}

	config.Infof("✅ Unified template engine initialized successfully")
	return nil
}

// ReloadDefaultTemplates 重新加载默认模板引擎的所有模板
// 用于在函数注册后刷新模板，确保新注册的函数能被识别
func ReloadDefaultTemplates() error {
	engine := GetDefaultEngine()
	if engine != nil {
		return engine.ReloadAllTemplates()
	}
	return fmt.Errorf("default template engine not initialized")
}

// NewTemplateEngine 创建新的模板引擎
func NewTemplateEngine(cfg *config.TemplateConfig) (*TemplateEngine, error) {
	if cfg == nil {
		cfg = DefaultTemplateConfig()
	}

	// 🔧 修复路径配置：将相对路径转换为绝对路径
	absoluteViewPaths, err := resolveViewPaths(cfg.Paths.ViewPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve view paths: %w", err)
	}

	absoluteLayoutPath, err := resolveAbsolutePath(cfg.Paths.LayoutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve layout path: %w", err)
	}

	absoluteComponentPath, err := resolveAbsolutePath(cfg.Paths.ComponentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve component path: %w", err)
	}

	// 输出路径解析日志，便于调试
	config.Infof("📁 Template Engine Path Resolution:")
	config.Infof("   Working Directory: %s", getCurrentWorkingDir())
	config.Infof("   Original View Paths: %v", cfg.Paths.ViewPaths)
	config.Infof("   Resolved View Paths: %v", absoluteViewPaths)
	config.Infof("   Original Layout Path: %s", cfg.Paths.LayoutPath)
	config.Infof("   Resolved Layout Path: %s", absoluteLayoutPath)
	if cfg.Paths.ComponentPath != "" {
		config.Infof("   Original Component Path: %s", cfg.Paths.ComponentPath)
		config.Infof("   Resolved Component Path: %s", absoluteComponentPath)
	}

	engine := &TemplateEngine{
		// 核心配置 - 使用解析后的绝对路径
		viewPaths:      absoluteViewPaths,
		layoutPath:     absoluteLayoutPath,
		componentPath:  absoluteComponentPath,
		extension:      cfg.Paths.Extension,
		delimLeft:      cfg.Syntax.DelimLeft,
		delimRight:     cfg.Syntax.DelimRight,
		enableCache:    cfg.Cache.EnableCache,
		enableReload:   cfg.Reload.Enabled,
		enableCompress: cfg.Performance.EnableCompress,
		currentTheme:   cfg.Theme.Current,

		// 缓存映射
		templates:  make(map[string]*template.Template),
		layouts:    make(map[string]*template.Template),
		components: make(map[string]*template.Template),
		watchPaths: make(map[string]bool),
		funcMap:    make(template.FuncMap),
		themes:     cfg.Theme.Themes,

		// Beego 兼容配置
		RunMode:          "dev", // 默认开发模式
		EnableGzip:       cfg.Performance.EnableCompress,
		DirectoryIndex:   false,
		EnableAutoRender: true,
		AutoRender:       true,
		TemplateExt:      []string{".html", ".tpl", ".gohtml"},

		// Beego 高级功能初始化
		templateDirs: make(map[string]os.FileInfo),
		lastModTime:  make(map[string]time.Time),
		watchedFiles: make(map[string]bool),
	}

	// 注册默认模板函数
	engine.registerDefaultFunctions()

	// 初始化 Beego 风格的模板函数
	engine.initBeegoFunctions()

	// 设置全局引擎引用（用于Beego函数访问）
	SetGlobalEngineForFunctions(engine)

	// 初始化热重载
	if engine.enableReload {
		if err := engine.initWatcher(); err != nil {
			config.Warnf("Failed to initialize template watcher: %v", err)
		}
	}

	// 预加载模板
	if err := engine.loadAllTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	config.Infof("Template engine initialized with theme: %s", engine.currentTheme)
	return engine, nil
}

// registerDefaultFunctions 注册默认模板函数
func (e *TemplateEngine) registerDefaultFunctions() {
	// 使用统一的函数管理器获取所有函数
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	// 清空原有函数映射并使用合并后的函数
	e.funcMap = make(template.FuncMap)
	maps.Copy(e.funcMap, mergedFuncs)

	// 添加新的增强函数
	e.funcMap["include"] = e.includeTemplate
	e.funcMap["component"] = e.renderComponent
	e.funcMap["theme"] = e.getThemeVariable
	e.funcMap["asset"] = e.getAssetURL
	e.funcMap["url"] = e.buildURL
	e.funcMap["csrf"] = e.getCSRFToken
	e.funcMap["csrf_token"] = e.getCSRFToken // 下划线别名
	e.funcMap["flash"] = e.getFlashMessage
	e.funcMap["getFlashMessage"] = e.getFlashMessage
	e.funcMap["truncate"] = e.truncateString
	e.funcMap["truncateString"] = e.truncateString
	e.funcMap["markdown"] = e.renderMarkdown
	e.funcMap["renderMarkdown"] = e.renderMarkdown
	e.funcMap["json"] = e.toJSON
	e.funcMap["toJSON"] = e.toJSON
	e.funcMap["safe"] = e.safeHTML
	e.funcMap["safeHTML"] = e.safeHTML
	e.funcMap["dict"] = e.createDict
	e.funcMap["createDict"] = e.createDict
	e.funcMap["slice"] = e.createSlice
	e.funcMap["createSlice"] = e.createSlice
	e.funcMap["range"] = e.createRange
	e.funcMap["createRange"] = e.createRange
	// 保留引擎自有的格式化函数，但不与Beego函数冲突
	e.funcMap["dateFormat"] = e.formatDate
	e.funcMap["formatDate"] = e.formatDate
	e.funcMap["currency"] = e.formatCurrency
	e.funcMap["formatCurrency"] = e.formatCurrency
	e.funcMap["filesize"] = e.formatFileSize
	e.funcMap["formatFileSize"] = e.formatFileSize
}

// AddFunction 添加自定义模板函数到引擎
func (e *TemplateEngine) AddFunction(name string, fn any) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 获取统一管理器的合并函数
	manager := GetGlobalFunctionManager()

	// 先设置引擎自定义函数
	if e.funcMap == nil {
		e.funcMap = make(template.FuncMap)
	}
	e.funcMap[name] = fn

	// 重新获取合并后的函数映射
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)
	e.funcMap = mergedFuncs

	// 如果已经加载了模板，需要重新编译
	if e.enableCache {
		e.templates = make(map[string]*template.Template)
		e.layouts = make(map[string]*template.Template)
		e.components = make(map[string]*template.Template)

		if err := e.loadAllTemplates(); err != nil {
			config.Errorf("Failed to reload templates after adding function: %v", err)
		}
	}
}

// SetTheme 设置当前主题
func (e *TemplateEngine) SetTheme(themeName string) error {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	theme, exists := e.themes[themeName]
	if !exists {
		return fmt.Errorf("theme '%s' not found", themeName)
	}

	if !theme.Enabled {
		return fmt.Errorf("theme '%s' is disabled", themeName)
	}

	// 更新当前主题配置
	e.currentTheme = themeName
	e.viewPaths = theme.ViewPaths
	e.layoutPath = theme.LayoutPath
	e.componentPath = theme.ComponentPath

	// 清除缓存并重新加载
	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)

	if err := e.loadAllTemplates(); err != nil {
		return fmt.Errorf("failed to load templates for theme '%s': %w", themeName, err)
	}

	config.Infof("Switched to theme: %s", themeName)
	return nil
}

// SetDelimiters 设置模板分隔符
func (e *TemplateEngine) SetDelimiters(left, right string) error {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 验证分隔符
	if err := validateDelimiters(left, right); err != nil {
		return fmt.Errorf("invalid delimiters: %w", err)
	}

	// 更新分隔符
	oldLeft, oldRight := e.delimLeft, e.delimRight
	e.delimLeft = left
	e.delimRight = right

	// 清除所有模板缓存，因为需要用新分隔符重新解析
	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)

	// 重新加载所有模板
	if err := e.loadAllTemplates(); err != nil {
		// 如果重新加载失败，恢复原分隔符
		e.delimLeft = oldLeft
		e.delimRight = oldRight
		return fmt.Errorf("failed to reload templates with new delimiters '%s' and '%s': %w", left, right, err)
	}

	config.Infof("Template delimiters updated from '%s'/'%s' to '%s'/'%s'", oldLeft, oldRight, left, right)
	return nil
}

// validateDelimiters 验证分隔符的有效性
func validateDelimiters(left, right string) error {
	if left == "" || right == "" {
		return fmt.Errorf("delimiters cannot be empty")
	}

	if left == right {
		return fmt.Errorf("left and right delimiters cannot be the same")
	}

	// // 检查是否包含可能导致解析问题的模式
	// problematicPatterns := []string{
	// 	"{{{", "}}}", // 三重花括号会导致Go模板解析错误
	// }

	// for _, pattern := range problematicPatterns {
	// 	if strings.Contains(left, pattern) || strings.Contains(right, pattern) {
	// 		return fmt.Errorf("delimiter contains problematic pattern '%s' that may cause parsing errors", pattern)
	// 	}
	// }

	// 建议的安全分隔符组合
	safeDelimiters := map[string]string{
		"{{{": "}}}",
		"{[{": "}]}",
		"<<%": "%>>",
		"[[":  "]]",
		"{%":  "%}",
		"{{":  "}}", // 默认的Go模板分隔符
	}

	// 检查是否是已知的安全组合
	if expectedRight, isSafe := safeDelimiters[left]; isSafe {
		if right != expectedRight {
			config.Warnf("Left delimiter '%s' is typically paired with '%s', but you're using '%s'. This might cause issues.", left, expectedRight, right)
		}
	} else {
		config.Warnf("Using non-standard delimiters '%s' and '%s'. Please ensure they don't conflict with your template content.", left, right)
	}

	return nil
}

// GetDelimiters 获取当前的模板分隔符
func (e *TemplateEngine) GetDelimiters() (string, string) {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()
	return e.delimLeft, e.delimRight
}

// AddTheme 添加新主题
func (e *TemplateEngine) AddTheme(name string, theme *config.ThemeConfig) error {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	if e.themes == nil {
		e.themes = make(map[string]*config.ThemeConfig)
	}

	theme.Name = name
	e.themes[name] = theme

	config.Infof("Added theme: %s", name)
	return nil
}

// GetTheme 获取主题配置
func (e *TemplateEngine) GetTheme(name string) (*config.ThemeConfig, bool) {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	theme, exists := e.themes[name]
	return theme, exists
}

// GetCurrentTheme 获取当前主题名称
func (e *TemplateEngine) GetCurrentTheme() string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	return e.currentTheme
}

// ============= 测试辅助方法 =============

// CreateInlineTemplate 创建内联模板（用于测试）
func (e *TemplateEngine) CreateInlineTemplate(name, content string) (*template.Template, error) {
	// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	tmpl := template.New(name).
		Delims(e.delimLeft, e.delimRight).
		Funcs(mergedFuncs)

	return tmpl.Parse(content)
}

// ExecuteTemplate 执行模板（用于测试）
func (e *TemplateEngine) ExecuteTemplate(tmpl *template.Template, data any) (string, error) {
	var buf strings.Builder
	err := tmpl.Execute(&buf, data)
	return buf.String(), err
}

// createInlineTemplate 创建内联模板（用于测试）- 内部方法兼容性
func (e *TemplateEngine) createInlineTemplate(name, content string) (*template.Template, error) {
	return e.CreateInlineTemplate(name, content)
}

// executeTemplate 执行模板（用于测试）- 内部方法兼容性
func (e *TemplateEngine) executeTemplate(tmpl *template.Template, data any) (string, error) {
	return e.ExecuteTemplate(tmpl, data)
}

// GetAvailableThemes 获取所有可用主题
func (e *TemplateEngine) GetAvailableThemes() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	themes := make([]string, 0, len(e.themes))
	for name, theme := range e.themes {
		if theme.Enabled {
			themes = append(themes, name)
		}
	}
	return themes
}

// initWatcher 初始化文件监控器
func (e *TemplateEngine) initWatcher() error {
	var err error
	e.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// 添加监控路径
	allPaths := append(e.viewPaths, e.layoutPath, e.componentPath)
	for _, path := range allPaths {
		if err := e.addWatchPath(path); err != nil {
			config.Warnf("Failed to watch path %s: %v", path, err)
		}
	}

	// 启动监控协程
	go e.watchFiles()

	return nil
}

// addWatchPath 添加监控路径
func (e *TemplateEngine) addWatchPath(path string) error {
	if e.watchPaths[path] {
		return nil // 已经在监控
	}

	err := filepath.WalkDir(path, func(walkPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续
		}

		if d.IsDir() {
			if err := e.watcher.Add(walkPath); err != nil {
				return err
			}
		}
		return nil
	})

	if err == nil {
		e.watchPaths[path] = true
	}

	return err
}

// watchFiles 监控文件变化（增强版，带防抖机制）
func (e *TemplateEngine) watchFiles() {
	// 防抖机制：收集一段时间内的所有事件，然后批量处理
	debounceTimer := time.NewTimer(0)
	debounceTimer.Stop()
	pendingEvents := make(map[string]fsnotify.Event)

	for {
		select {
		case event, ok := <-e.watcher.Events:
			if !ok {
				return
			}

			// 只处理模板文件相关事件
			if !e.isTemplateFile(event.Name) {
				continue
			}

			config.Debugf("Template file event: %s %s", event.Name, event.Op.String())

			// 处理写入、创建、删除和重命名事件
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				// 将事件添加到待处理列表
				pendingEvents[event.Name] = event

				// 重置防抖定时器
				debounceTimer.Stop()
				debounceTimer.Reset(500 * time.Millisecond) // 500ms防抖延迟
			}

		case <-debounceTimer.C:
			// 防抖时间到，处理所有积累的事件
			if len(pendingEvents) > 0 {
				config.Infof("Processing %d template file changes after debounce delay", len(pendingEvents))

				// 处理每个变化的文件
				for filePath, event := range pendingEvents {
					config.Debugf("  Processing: %s (%s)", filePath, event.Op.String())

					switch event.Op & (fsnotify.Write | fsnotify.Create | fsnotify.Remove | fsnotify.Rename) {
					case fsnotify.Write, fsnotify.Create:
						// 文件被写入或创建，重新加载
						e.ReloadTemplate(filePath)
					case fsnotify.Remove, fsnotify.Rename:
						// 文件被删除或重命名，从缓存中移除
						e.RemoveTemplateFromCache(filePath)
					}
				}

				// 清空待处理事件
				pendingEvents = make(map[string]fsnotify.Event)
			}

		case err, ok := <-e.watcher.Errors:
			if !ok {
				return
			}
			config.Errorf("Template watcher error: %v", err)

			// 尝试重新初始化watcher
			config.Infof("Attempting to reinitialize template watcher...")
			if reinitErr := e.reinitWatcher(); reinitErr != nil {
				config.Errorf("Failed to reinitialize watcher: %v", reinitErr)
			} else {
				config.Infof("Template watcher reinitialized successfully")
			}
		}
	}
}

// ReloadTemplate 重新加载特定模板
func (e *TemplateEngine) ReloadTemplate(filePath string) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 清除相关缓存
	templateName := e.getTemplateName(filePath)
	delete(e.templates, templateName)

	// 如果是布局或组件文件，清除所有缓存
	if strings.Contains(filePath, e.layoutPath) || strings.Contains(filePath, e.componentPath) {
		e.templates = make(map[string]*template.Template)
		e.layouts = make(map[string]*template.Template)
		e.components = make(map[string]*template.Template)
	}

	config.Debugf("Template cache cleared for: %s", templateName)
}

// RemoveTemplateFromCache 从缓存中移除模板
func (e *TemplateEngine) RemoveTemplateFromCache(filePath string) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	templateName := e.getTemplateName(filePath)

	// 从各种缓存中移除
	delete(e.templates, templateName)
	delete(e.layouts, templateName)
	delete(e.components, templateName)

	// 如果是布局或组件文件变化，需要清空所有模板缓存
	// 因为其他模板可能依赖于这些变化的布局或组件
	if strings.Contains(filePath, e.layoutPath) {
		config.Infof("Layout file removed/renamed: %s, clearing all template cache", templateName)
		e.templates = make(map[string]*template.Template)
	} else if strings.Contains(filePath, e.componentPath) {
		config.Infof("Component file removed/renamed: %s, clearing all template cache", templateName)
		e.templates = make(map[string]*template.Template)
	}

	config.Debugf("Template removed from cache: %s", templateName)
}

// reinitWatcher 重新初始化文件监控器
func (e *TemplateEngine) reinitWatcher() error {
	// 关闭现有的watcher
	if e.watcher != nil {
		e.watcher.Close()
	}

	// 重新初始化
	return e.initWatcher()
}

// ReloadAllTemplates 重新加载所有模板（用于函数注册后刷新模板）
func (e *TemplateEngine) ReloadAllTemplates() error {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 清除所有缓存
	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)

	config.Infof("Reloading all templates with updated functions...")

	// 重新加载所有模板
	if err := e.loadAllTemplates(); err != nil {
		return fmt.Errorf("failed to reload templates: %w", err)
	}

	config.Infof("Successfully reloaded all templates")
	return nil
}

// Close 关闭模板引擎（增强版，合并 Beego 兼容）
func (e *TemplateEngine) Close() error {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 清空缓存
	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)
	if e.lastModTime != nil {
		e.lastModTime = make(map[string]time.Time)
	}
	if e.watchedFiles != nil {
		e.watchedFiles = make(map[string]bool)
	}

	// 关闭文件监控器
	if e.watcher != nil {
		if err := e.watcher.Close(); err != nil {
			config.Warnf("Error closing template watcher: %v", err)
		}
	}

	config.Info("Template engine closed successfully")
	return nil
}

// getTemplateName 从文件路径获取模板名称
func (e *TemplateEngine) getTemplateName(filePath string) string {
	// 移除扩展名
	name := strings.TrimSuffix(filepath.Base(filePath), e.extension)

	// 如果包含目录，保留相对路径
	for _, viewPath := range e.viewPaths {
		if strings.HasPrefix(filePath, viewPath) {
			relPath, _ := filepath.Rel(viewPath, filePath)
			name = strings.TrimSuffix(relPath, e.extension)
			break
		}
	}

	return strings.ReplaceAll(name, "\\", "/") // 标准化路径分隔符
}

// ============= 模板加载方法 =============

// loadAllTemplates 加载所有模板
func (e *TemplateEngine) loadAllTemplates() error {
	// 加载布局
	if err := e.loadLayouts(); err != nil {
		return fmt.Errorf("failed to load layouts: %w", err)
	}

	// 加载组件
	if err := e.loadComponents(); err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	// 加载视图模板
	if err := e.loadViewTemplates(); err != nil {
		return fmt.Errorf("failed to load view templates: %w", err)
	}

	config.Infof("Loaded %d templates, %d layouts, %d components",
		len(e.templates), len(e.layouts), len(e.components))

	return nil
}

// loadLayouts 加载布局模板
func (e *TemplateEngine) loadLayouts() error {
	if e.layoutPath == "" {
		return nil
	}

	return filepath.WalkDir(e.layoutPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续
		}

		if !d.IsDir() && strings.HasSuffix(path, e.extension) {
			layoutName := e.getTemplateName(path)

			// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
			manager := GetGlobalFunctionManager()
			mergedFuncs := manager.GetMergedFunctions(e.funcMap)

			tmpl := template.New(layoutName).
				Delims(e.delimLeft, e.delimRight).
				Funcs(mergedFuncs)

			if _, err := tmpl.ParseFiles(path); err != nil {
				config.Errorf("Failed to parse layout %s: %v", path, err)
				return nil
			}

			e.layouts[layoutName] = tmpl
			config.Debugf("Loaded layout: %s", layoutName)
		}

		return nil
	})
}

// loadComponents 加载组件模板
func (e *TemplateEngine) loadComponents() error {
	if e.componentPath == "" {
		return nil
	}

	return filepath.WalkDir(e.componentPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续
		}

		if !d.IsDir() && strings.HasSuffix(path, e.extension) {
			componentName := e.getTemplateName(path)

			// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
			manager := GetGlobalFunctionManager()
			mergedFuncs := manager.GetMergedFunctions(e.funcMap)

			tmpl := template.New(componentName).
				Delims(e.delimLeft, e.delimRight).
				Funcs(mergedFuncs)

			if _, err := tmpl.ParseFiles(path); err != nil {
				config.Errorf("Failed to parse component %s: %v", path, err)
				return nil
			}

			e.components[componentName] = tmpl
			config.Debugf("Loaded component: %s", componentName)
		}

		return nil
	})
}

// loadViewTemplates 加载视图模板
func (e *TemplateEngine) loadViewTemplates() error {
	for _, viewPath := range e.viewPaths {
		err := filepath.WalkDir(viewPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // 忽略错误，继续
			}

			if !d.IsDir() && strings.HasSuffix(path, e.extension) {
				// 跳过布局和组件目录
				if strings.Contains(path, e.layoutPath) || strings.Contains(path, e.componentPath) {
					return nil
				}

				templateName := e.getTemplateName(path)

				// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
				manager := GetGlobalFunctionManager()
				mergedFuncs := manager.GetMergedFunctions(e.funcMap)

				tmpl := template.New(templateName).
					Delims(e.delimLeft, e.delimRight).
					Funcs(mergedFuncs)

				parsedTmpl, err := tmpl.ParseFiles(path)
				if err != nil {
					config.Errorf("Failed to parse template %s: %v", path, err)
					return nil
				}

				// 查找实际的模板（类似loadTemplate中的逻辑）
				templates := parsedTmpl.Templates()
				var actualTemplate *template.Template

				for _, t := range templates {
					if t.Tree != nil && t.Tree.Root != nil {
						actualTemplate = t
						break
					}
				}

				if actualTemplate != nil {
					e.templates[templateName] = actualTemplate
					config.Debugf("Loaded template: %s -> %s", templateName, actualTemplate.Name())
				} else {
					config.Warnf("Template %s is empty or invalid", templateName)
				}
			}

			return nil
		})

		if err != nil {
			config.Warnf("Error walking view path %s: %v", viewPath, err)
		}
	}

	return nil
}

// ============= 便捷函数 =============

// Render 渲染模板（使用默认引擎）
func Render(templateName string, data any) (string, error) {
	return GetDefaultEngine().Render(templateName, data)
}

// RenderWithLayout 使用布局渲染模板（使用默认引擎）
func RenderWithLayout(templateName, layoutName string, data any) (string, error) {
	return GetDefaultEngine().RenderWithLayout(templateName, layoutName, data)
}

// RenderComponent 渲染组件（使用默认引擎）
func RenderComponent(componentName string, data any) (string, error) {
	return GetDefaultEngine().RenderComponent(componentName, data)
}

// AddTemplateFunction 添加模板函数（使用默认引擎）
func AddTemplateFunction(name string, fn any) {
	GetDefaultEngine().AddFunction(name, fn)
}

// SetCurrentTheme 设置当前主题（使用默认引擎）
func SetCurrentTheme(themeName string) error {
	return GetDefaultEngine().SetTheme(themeName)
}

// SetTemplateDelimiters 设置模板分隔符（使用默认引擎）
func SetTemplateDelimiters(left, right string) error {
	return GetDefaultEngine().SetDelimiters(left, right)
}

// GetTemplateDelimiters 获取模板分隔符（使用默认引擎）
func GetTemplateDelimiters() (string, string) {
	return GetDefaultEngine().GetDelimiters()
}

// ============= 模板管理器 =============

// TemplateManager 模板管理器（单实例）
type TemplateManager struct {
	engine *TemplateEngine
	config *config.TemplateConfig
	mutex  sync.RWMutex
}

var (
	templateManager *TemplateManager
	templateOnce    sync.Once
)

// GetTemplateManager 获取模板管理器单实例
func GetTemplateManager() *TemplateManager {
	templateOnce.Do(func() {
		var err error
		templateManager, err = NewTemplateManager()
		if err != nil {
			config.Fatalf("Failed to initialize template manager: %v", err)
		}
	})
	return templateManager
}

// NewTemplateManager 创建新的模板管理器（统一引擎管理）
func NewTemplateManager() (*TemplateManager, error) {
	// 【重要修复】使用统一的模板引擎实例，避免重复创建
	engine := GetUnifiedEngine()
	if engine == nil {
		return nil, fmt.Errorf("failed to get unified template engine")
	}

	// 使用从config包加载的模板配置
	templateConfig := config.GlobalTemplate

	manager := &TemplateManager{
		engine: engine,
		config: templateConfig,
	}

	config.Infof("Template manager initialized with unified engine successfully")
	return manager, nil
}

// LoadTemplateConfigFromFile 从配置文件加载模板配置
func LoadTemplateConfigFromFile() (*config.TemplateConfig, error) {
	// 直接使用config包的LoadTemplateConfig
	cfg := config.LoadTemplateConfig()

	config.Infof("Loaded template configuration from config file")
	return cfg, nil
}

// GetEngine 获取模板引擎
func (tm *TemplateManager) GetEngine() *TemplateEngine {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.engine
}

// GetConfig 获取模板配置
func (tm *TemplateManager) GetConfig() *config.TemplateConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.config
}

// Render 渲染模板
func (tm *TemplateManager) Render(templateName string, data any) (string, error) {
	return tm.engine.Render(templateName, data)
}

// RenderWithLayout 使用布局渲染模板
func (tm *TemplateManager) RenderWithLayout(templateName, layoutName string, data any) (string, error) {
	return tm.engine.RenderWithLayout(templateName, layoutName, data)
}

// RenderComponent 渲染组件
func (tm *TemplateManager) RenderComponent(componentName string, data any) (string, error) {
	return tm.engine.RenderComponent(componentName, data)
}

// SetTheme 设置当前主题
func (tm *TemplateManager) SetTheme(themeName string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	return tm.engine.SetTheme(themeName)
}

// GetCurrentTheme 获取当前主题
func (tm *TemplateManager) GetCurrentTheme() string {
	return tm.engine.GetCurrentTheme()
}

// GetAvailableThemes 获取可用主题列表
func (tm *TemplateManager) GetAvailableThemes() []string {
	return tm.engine.GetAvailableThemes()
}

// AddFunction 添加模板函数
func (tm *TemplateManager) AddFunction(name string, fn any) {
	tm.engine.AddFunction(name, fn)
}

// AddTheme 添加新主题
func (tm *TemplateManager) AddTheme(name string, theme *config.ThemeConfig) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	return tm.engine.AddTheme(name, theme)
}

// SetDelimiters 设置模板分隔符
func (tm *TemplateManager) SetDelimiters(left, right string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 调用引擎的SetDelimiters方法
	if err := tm.engine.SetDelimiters(left, right); err != nil {
		return err
	}

	// 更新配置中的分隔符设置
	tm.config.Syntax.DelimLeft = left
	tm.config.Syntax.DelimRight = right

	config.Infof("Template manager delimiters updated to '%s' and '%s'", left, right)
	return nil
}

// GetDelimiters 获取当前的模板分隔符
func (tm *TemplateManager) GetDelimiters() (string, string) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return tm.engine.GetDelimiters()
}

// ReloadConfig 重新加载配置
func (tm *TemplateManager) ReloadConfig() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 重新加载配置
	newConfig, err := LoadTemplateConfigFromFile()
	if err != nil {
		return fmt.Errorf("failed to reload template config: %w", err)
	}

	// 关闭当前引擎
	if tm.engine != nil {
		_ = tm.engine.Close() // 忽略关闭错误，继续执行
	}

	// 创建新引擎
	newEngine, err := NewTemplateEngine(newConfig)
	if err != nil {
		return fmt.Errorf("failed to create new template engine: %w", err)
	}

	// 更新引擎和配置
	tm.engine = newEngine
	tm.config = newConfig

	config.Infof("Template configuration reloaded successfully")
	return nil
}

// Close 关闭模板管理器
func (tm *TemplateManager) Close() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.engine != nil {
		return tm.engine.Close()
	}
	return nil
}

// ============= 增强模板管理器 =============

// EnhancedTemplateManager 增强的模板管理器（基于Beego机制）
type EnhancedTemplateManager struct {
	*TemplateManager // 嵌入原有管理器

	// Beego风格的增强功能
	autoDiscover   bool                 // 自动发现模板
	fileWatcher    *fsnotify.Watcher    // 文件监控器
	watcherEnabled bool                 // 监控是否启用
	templatePaths  map[string]string    // 模板路径映射
	lastModified   map[string]time.Time // 文件修改时间缓存
	discoverMutex  sync.RWMutex         // 发现过程锁
}

// NewEnhancedTemplateManager 创建增强的模板管理器（统一引擎管理）
func NewEnhancedTemplateManager() (*EnhancedTemplateManager, error) {
	// 【重要修复】使用统一的模板管理器，避免重复创建引擎
	baseManager := GetTemplateManager()
	if baseManager == nil {
		return nil, fmt.Errorf("failed to get template manager")
	}

	enhanced := &EnhancedTemplateManager{
		TemplateManager: baseManager,
		autoDiscover:    true,
		templatePaths:   make(map[string]string),
		lastModified:    make(map[string]time.Time),
	}

	// 执行自动发现
	if enhanced.autoDiscover {
		err := enhanced.DiscoverTemplates()
		if err != nil {
			config.Warnf("Template auto-discovery failed: %v", err)
		}
	}

	// 启用文件监控（开发模式）
	if enhanced.config.Reload.Enabled {
		err := enhanced.EnableFileWatcher()
		if err != nil {
			config.Warnf("Failed to enable file watcher: %v", err)
		}
	}

	config.Infof("Enhanced template manager initialized with unified engine")
	return enhanced, nil
}

// DiscoverTemplates 自动发现模板文件（类似Beego的模板扫描）
func (etm *EnhancedTemplateManager) DiscoverTemplates() error {
	etm.discoverMutex.Lock()
	defer etm.discoverMutex.Unlock()

	config.Infof("Starting template discovery...")
	startTime := time.Now()
	discoveredCount := 0

	// 扫描所有配置的视图路径
	for _, viewPath := range etm.config.Paths.ViewPaths {
		count, err := etm.discoverTemplatesInPath(viewPath)
		if err != nil {
			config.Warnf("Failed to discover templates in %s: %v", viewPath, err)
			continue
		}
		discoveredCount += count
	}

	// 扫描布局路径
	if etm.config.Paths.LayoutPath != "" {
		count, err := etm.discoverTemplatesInPath(etm.config.Paths.LayoutPath)
		if err != nil {
			config.Warnf("Failed to discover layouts in %s: %v", etm.config.Paths.LayoutPath, err)
		} else {
			discoveredCount += count
		}
	}

	// 扫描组件路径
	if etm.config.Paths.ComponentPath != "" {
		count, err := etm.discoverTemplatesInPath(etm.config.Paths.ComponentPath)
		if err != nil {
			config.Warnf("Failed to discover components in %s: %v", etm.config.Paths.ComponentPath, err)
		} else {
			discoveredCount += count
		}
	}

	elapsed := time.Since(startTime)
	config.Infof("Template discovery completed: %d templates found in %v", discoveredCount, elapsed)

	return nil
}

// discoverTemplatesInPath 在指定路径中发现模板
func (etm *EnhancedTemplateManager) discoverTemplatesInPath(rootPath string) (int, error) {
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		config.Debugf("Template path does not exist: %s", rootPath)
		return 0, nil
	}

	count := 0
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名
		if !etm.isTemplateFile(path) {
			return nil
		}

		// 记录模板路径和修改时间
		info, err := d.Info()
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			relativePath = path
		}

		etm.templatePaths[relativePath] = path
		etm.lastModified[path] = info.ModTime()
		count++

		config.Debugf("Discovered template: %s -> %s", relativePath, path)

		return nil
	})

	return count, err
}

// isTemplateFile 检查是否是模板文件
func (etm *EnhancedTemplateManager) isTemplateFile(filePath string) bool {
	ext := filepath.Ext(filePath)

	// 检查配置的扩展名
	if etm.config.Paths.Extension != "" && ext == etm.config.Paths.Extension {
		return true
	}

	// 检查常见的模板扩展名
	commonExts := []string{".html", ".htm", ".tpl", ".tmpl", ".gohtml"}
	return slices.Contains(commonExts, ext)
}

// EnableFileWatcher 启用文件监控（类似Beego的文件监控）
func (etm *EnhancedTemplateManager) EnableFileWatcher() error {
	if etm.watcherEnabled {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	etm.fileWatcher = watcher
	etm.watcherEnabled = true

	// 监控所有模板路径
	watchPaths := append(etm.config.Paths.ViewPaths, etm.config.Paths.LayoutPath, etm.config.Paths.ComponentPath)
	for _, path := range watchPaths {
		if path != "" {
			if err := etm.addWatchPath(path); err != nil {
				config.Warnf("Failed to watch path %s: %v", path, err)
			}
		}
	}

	// 启动监控协程
	go etm.handleFileEvents()

	config.Infof("File watcher enabled for template hot-reload")
	return nil
}

// addWatchPath 添加监控路径
func (etm *EnhancedTemplateManager) addWatchPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		config.Debugf("Skip watching non-existent path: %s", path)
		return nil
	}

	// 递归添加所有子目录
	return filepath.WalkDir(path, func(walkPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if err := etm.fileWatcher.Add(walkPath); err != nil {
				config.Warnf("Failed to watch directory %s: %v", walkPath, err)
			} else {
				config.Debugf("Watching directory: %s", walkPath)
			}
		}

		return nil
	})
}

// handleFileEvents 处理文件事件
func (etm *EnhancedTemplateManager) handleFileEvents() {
	for {
		select {
		case event, ok := <-etm.fileWatcher.Events:
			if !ok {
				return
			}
			etm.handleFileEvent(event)

		case err, ok := <-etm.fileWatcher.Errors:
			if !ok {
				return
			}
			config.Errorf("File watcher error: %v", err)
		}
	}
}

// handleFileEvent 处理单个文件事件
func (etm *EnhancedTemplateManager) handleFileEvent(event fsnotify.Event) {
	// 只处理模板文件
	if !etm.isTemplateFile(event.Name) {
		return
	}

	switch {
	case event.Op&fsnotify.Write == fsnotify.Write:
		etm.handleFileModified(event.Name)
	case event.Op&fsnotify.Create == fsnotify.Create:
		etm.handleFileCreated(event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		etm.handleFileRemoved(event.Name)
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		etm.handleFileRenamed(event.Name)
	}
}

// handleFileModified 处理文件修改事件
func (etm *EnhancedTemplateManager) handleFileModified(filePath string) {
	config.Debugf("Template file modified: %s", filePath)

	// 检查文件是否真的被修改了（避免重复事件）
	if info, err := os.Stat(filePath); err == nil {
		if lastMod, exists := etm.lastModified[filePath]; exists {
			if !info.ModTime().After(lastMod) {
				return // 文件没有真正修改
			}
		}
		etm.lastModified[filePath] = info.ModTime()
	}

	// 清除相关模板缓存
	etm.clearTemplateCache(filePath)

	config.Infof("Template cache cleared for: %s", filePath)
}

// handleFileCreated 处理文件创建事件
func (etm *EnhancedTemplateManager) handleFileCreated(filePath string) {
	config.Debugf("Template file created: %s", filePath)

	// 重新发现模板
	go func() {
		time.Sleep(100 * time.Millisecond) // 短暂延迟确保文件写入完成
		if err := etm.DiscoverTemplates(); err != nil {
			config.Warnf("Failed to rediscover templates after file creation: %v", err)
		}
	}()
}

// handleFileRemoved 处理文件删除事件
func (etm *EnhancedTemplateManager) handleFileRemoved(filePath string) {
	config.Debugf("Template file removed: %s", filePath)

	// 清除缓存和记录
	etm.clearTemplateCache(filePath)
	delete(etm.lastModified, filePath)

	// 从路径映射中移除
	for name, path := range etm.templatePaths {
		if path == filePath {
			delete(etm.templatePaths, name)
			break
		}
	}
}

// handleFileRenamed 处理文件重命名事件
func (etm *EnhancedTemplateManager) handleFileRenamed(filePath string) {
	config.Debugf("Template file renamed: %s", filePath)
	etm.handleFileRemoved(filePath)
	etm.handleFileCreated(filePath)
}

// clearTemplateCache 清除模板缓存
func (etm *EnhancedTemplateManager) clearTemplateCache(filePath string) {
	// 调用底层模板引擎的缓存清除方法
	if etm.engine != nil {
		config.Debugf("Template cache clearing requested for: %s", filePath)
	}
}

// GetTemplateInfo 获取模板信息
func (etm *EnhancedTemplateManager) GetTemplateInfo() map[string]any {
	etm.discoverMutex.RLock()
	defer etm.discoverMutex.RUnlock()

	info := map[string]any{
		"discovered_templates": len(etm.templatePaths),
		"watcher_enabled":      etm.watcherEnabled,
		"auto_discover":        etm.autoDiscover,
		"template_paths":       etm.templatePaths,
	}

	return info
}

// DisableFileWatcher 禁用文件监控
func (etm *EnhancedTemplateManager) DisableFileWatcher() error {
	if !etm.watcherEnabled || etm.fileWatcher == nil {
		return nil
	}

	err := etm.fileWatcher.Close()
	etm.watcherEnabled = false
	etm.fileWatcher = nil

	if err != nil {
		return fmt.Errorf("failed to close file watcher: %w", err)
	}

	config.Infof("File watcher disabled")
	return nil
}

// Close 关闭增强模板管理器
func (etm *EnhancedTemplateManager) Close() error {
	// 禁用文件监控
	if err := etm.DisableFileWatcher(); err != nil {
		config.Warnf("Error closing file watcher: %v", err)
	}

	// 关闭基础管理器
	return etm.TemplateManager.Close()
}

// ============= 全局增强管理器实例 =============

var (
	enhancedTemplateManager *EnhancedTemplateManager
	enhancedTemplateOnce    sync.Once
)

// GetEnhancedTemplateManager 获取增强模板管理器单实例
func GetEnhancedTemplateManager() *EnhancedTemplateManager {
	enhancedTemplateOnce.Do(func() {
		var err error
		enhancedTemplateManager, err = NewEnhancedTemplateManager()
		if err != nil {
			config.Fatalf("Failed to initialize enhanced template manager: %v", err)
		}
	})
	return enhancedTemplateManager
}

// ============= Beego自动推导功能 =============

// AutoTemplateInference Beego风格的自动模板推导器
type AutoTemplateInference struct {
	controllerMappings map[string]string // 控制器名映射
	actionMappings     map[string]string // 动作名映射
	pathTemplates      []PathTemplate    // 路径模板
	mu                 sync.RWMutex
}

// PathTemplate 路径模板结构
type PathTemplate struct {
	Pattern     string // 路径模式，如 "{controller}/{action}"
	Priority    int    // 优先级，数字越大优先级越高
	Description string // 描述
}

var (
	globalInference     *AutoTemplateInference
	globalInferenceOnce sync.Once
)

// GetAutoTemplateInference 获取全局自动推导器
func GetAutoTemplateInference() *AutoTemplateInference {
	globalInferenceOnce.Do(func() {
		globalInference = NewAutoTemplateInference()
	})
	return globalInference
}

// NewAutoTemplateInference 创建自动推导器
func NewAutoTemplateInference() *AutoTemplateInference {
	inference := &AutoTemplateInference{
		controllerMappings: make(map[string]string),
		actionMappings:     make(map[string]string),
		pathTemplates: []PathTemplate{
			// 默认的Beego路径模板，按优先级排序
			{Pattern: "{controller}/{action}", Priority: 100, Description: "标准Beego模式"},
			{Pattern: "{controller}/{action}/index", Priority: 90, Description: "带index后缀"},
			{Pattern: "{controller}/index", Priority: 80, Description: "控制器index"},
			{Pattern: "{action}", Priority: 70, Description: "仅动作名"},
			{Pattern: "shared/{action}", Priority: 60, Description: "共享模板"},
			{Pattern: "common/{action}", Priority: 50, Description: "通用模板"},
		},
	}

	// 初始化默认映射
	inference.initDefaultMappings()
	return inference
}

// initDefaultMappings 初始化默认映射
func (ati *AutoTemplateInference) initDefaultMappings() {
	// 常见控制器名映射（CamelCase -> lowercase）
	commonControllers := map[string]string{
		"UserController":    "user",
		"AdminController":   "admin",
		"LoginController":   "login",
		"IndexController":   "index",
		"HomeController":    "home",
		"ProductController": "product",
		"OrderController":   "order",
		"SystemController":  "system",
		"ConfigController":  "config",
		"ProfileController": "profile",
	}

	// 常见动作名映射
	commonActions := map[string]string{
		"Index":   "index",
		"List":    "list",
		"Show":    "show",
		"Edit":    "edit",
		"Create":  "create",
		"Update":  "update",
		"Delete":  "delete",
		"Login":   "login",
		"Logout":  "logout",
		"Profile": "profile",
		"Setting": "setting",
		"Detail":  "detail",
		"Search":  "search",
	}

	ati.mu.Lock()
	defer ati.mu.Unlock()

	maps.Copy(ati.controllerMappings, commonControllers)
	maps.Copy(ati.actionMappings, commonActions)
}

// InferTemplatePath 推导模板路径（核心功能 - 增强版）
func (ati *AutoTemplateInference) InferTemplatePath(controllerName, actionName string) []string {
	ati.mu.RLock()
	defer ati.mu.RUnlock()

	// 获取约定管理器
	conventionManager := GetConventionManager()

	// 首先尝试直接映射
	directKey := fmt.Sprintf("%s.%s", controllerName, actionName)
	if mapped, exists := ati.controllerMappings[directKey]; exists {
		return []string{mapped}
	}

	// 应用命名约定进行标准化
	normalizedController := ati.normalizeControllerName(controllerName)
	normalizedAction := ati.normalizeActionName(actionName)

	// 使用约定管理器进一步标准化
	conventionInput := fmt.Sprintf("%s.%s", controllerName, actionName)
	conventionPath := conventionManager.ApplyConvention(conventionInput, conventionManager.GetDefaultConvention())

	var candidates []string

	// 1. 添加约定管理器的结果（优先级最高）
	if conventionPath != conventionInput {
		candidates = append(candidates, conventionPath)
	}

	// 2. 根据路径模板生成候选路径
	for _, template := range ati.pathTemplates {
		path := ati.applyPathTemplate(template.Pattern, normalizedController, normalizedAction)
		if path != "" {
			candidates = append(candidates, path)
		}
	}

	// 3. 添加直接组合的路径
	if normalizedController != "" && normalizedAction != "" {
		candidates = append(candidates, fmt.Sprintf("%s/%s", normalizedController, normalizedAction))
	}

	// 4. 添加其他约定的变体
	for _, convention := range []NamingConvention{SnakeCase, KebabCase, LowerCase} {
		altPath := conventionManager.ApplyConvention(conventionInput, convention)
		if altPath != conventionInput && altPath != conventionPath {
			candidates = append(candidates, altPath)
		}
	}

	// 去重并保持顺序
	return ati.deduplicatePaths(candidates)
}

// normalizeControllerName 标准化控制器名
func (ati *AutoTemplateInference) normalizeControllerName(controllerName string) string {
	if controllerName == "" {
		return ""
	}

	// 检查是否有直接映射
	if mapped, exists := ati.controllerMappings[controllerName]; exists {
		return mapped
	}

	// 移除Controller后缀
	name := strings.TrimSuffix(controllerName, "Controller")

	// CamelCase转换为lowercase
	return ati.camelToLowercase(name)
}

// normalizeActionName 标准化动作名
func (ati *AutoTemplateInference) normalizeActionName(actionName string) string {
	if actionName == "" {
		return "index" // 默认动作
	}

	// 检查是否有直接映射
	if mapped, exists := ati.actionMappings[actionName]; exists {
		return mapped
	}

	// CamelCase转换为lowercase
	return ati.camelToLowercase(actionName)
}

// camelToLowercase 将CamelCase转换为lowercase
func (ati *AutoTemplateInference) camelToLowercase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			_, _ = result.WriteRune('_') // 忽略写入错误
		}
		if r >= 'A' && r <= 'Z' {
			_, _ = result.WriteRune(r - 'A' + 'a') // 忽略写入错误
		} else {
			_, _ = result.WriteRune(r) // 忽略写入错误
		}
	}
	return strings.ToLower(result.String())
}

// applyPathTemplate 应用路径模板
func (ati *AutoTemplateInference) applyPathTemplate(template, controller, action string) string {
	path := template
	path = strings.ReplaceAll(path, "{controller}", controller)
	path = strings.ReplaceAll(path, "{action}", action)

	// 清理路径
	path = strings.Trim(path, "/")
	path = strings.ReplaceAll(path, "//", "/")

	// 验证路径有效性
	if path == "" || strings.Contains(path, "{") {
		return "" // 模板变量未完全替换
	}

	return path
}

// deduplicatePaths 去重路径但保持顺序
func (ati *AutoTemplateInference) deduplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, path := range paths {
		if path != "" && !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}

	return result
}

// AddControllerMapping 添加控制器映射
func (ati *AutoTemplateInference) AddControllerMapping(controllerName, templatePath string) {
	ati.mu.Lock()
	defer ati.mu.Unlock()
	ati.controllerMappings[controllerName] = templatePath
	config.Debugf("Added controller mapping: %s -> %s", controllerName, templatePath)
}

// AddActionMapping 添加动作映射
func (ati *AutoTemplateInference) AddActionMapping(actionName, templatePath string) {
	ati.mu.Lock()
	defer ati.mu.Unlock()
	ati.actionMappings[actionName] = templatePath
	config.Debugf("Added action mapping: %s -> %s", actionName, templatePath)
}

// AddPathTemplate 添加路径模板
func (ati *AutoTemplateInference) AddPathTemplate(pattern string, priority int, description string) {
	ati.mu.Lock()
	defer ati.mu.Unlock()

	template := PathTemplate{
		Pattern:     pattern,
		Priority:    priority,
		Description: description,
	}

	// 插入到正确位置以保持优先级顺序
	inserted := false
	for i, existing := range ati.pathTemplates {
		if template.Priority > existing.Priority {
			ati.pathTemplates = append(ati.pathTemplates[:i], append([]PathTemplate{template}, ati.pathTemplates[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		ati.pathTemplates = append(ati.pathTemplates, template)
	}

	config.Debugf("Added path template: %s (priority: %d)", pattern, priority)
}

// ============= 路径映射和命名约定支持 =============

// PathMappingRule 路径映射规则
type PathMappingRule struct {
	Name        string              // 规则名称
	Pattern     string              // 匹配模式（支持正则表达式）
	Transform   func(string) string // 转换函数
	Priority    int                 // 优先级
	Description string              // 描述
	Enabled     bool                // 是否启用
}

// NamingConvention 命名约定类型
type NamingConvention string

const (
	CamelCase     NamingConvention = "camelCase"      // userProfile
	PascalCase    NamingConvention = "PascalCase"     // UserProfile
	SnakeCase     NamingConvention = "snake_case"     // user_profile
	KebabCase     NamingConvention = "kebab-case"     // user-profile
	LowerCase     NamingConvention = "lowercase"      // userprofile
	UpperCase     NamingConvention = "UPPERCASE"      // USERPROFILE
	BeegoStandard NamingConvention = "beego_standard" // user/profile
)

// ConventionManager 命名约定管理器
type ConventionManager struct {
	rules             []PathMappingRule // 映射规则
	conventions       map[string]string // 约定映射
	defaultConvention NamingConvention  // 默认约定
	mu                sync.RWMutex
}

var (
	globalConventionManager     *ConventionManager
	globalConventionManagerOnce sync.Once
)

// GetConventionManager 获取全局约定管理器
func GetConventionManager() *ConventionManager {
	globalConventionManagerOnce.Do(func() {
		globalConventionManager = NewConventionManager()
	})
	return globalConventionManager
}

// NewConventionManager 创建约定管理器
func NewConventionManager() *ConventionManager {
	manager := &ConventionManager{
		conventions:       make(map[string]string),
		defaultConvention: BeegoStandard,
		rules: []PathMappingRule{
			// 预定义的映射规则
			{
				Name:        "ControllerSuffix",
				Pattern:     `(.+)Controller$`,
				Transform:   func(s string) string { return strings.ToLower(s[:len(s)-10]) },
				Priority:    100,
				Description: "移除Controller后缀并转小写",
				Enabled:     true,
			},
			{
				Name:        "CamelToSnake",
				Pattern:     `[A-Z][a-z]*`,
				Transform:   func(s string) string { return camelToSnake(s) },
				Priority:    90,
				Description: "CamelCase转snake_case",
				Enabled:     true,
			},
			{
				Name:        "RemoveCommonPrefixes",
				Pattern:     `^(Get|Set|Show|View|Display)(.+)$`,
				Transform:   func(s string) string { return strings.ToLower(s[1:]) },
				Priority:    80,
				Description: "移除常见的动作前缀",
				Enabled:     true,
			},
		},
	}

	// 初始化预定义约定
	manager.initDefaultConventions()
	return manager
}

// initDefaultConventions 初始化默认约定
func (cm *ConventionManager) initDefaultConventions() {
	// Beego标准约定
	beegoConventions := map[string]string{
		"UserController.Index":   "user/index",
		"UserController.Profile": "user/profile",
		"AdminController.Index":  "admin/index",
		"LoginController.Login":  "login/login",
		"LoginController.Logout": "login/logout",
		"HomeController.Index":   "home/index",
		"ProductController.List": "product/list",
		"ProductController.Show": "product/show",
		"ProductController.Edit": "product/edit",
		"OrderController.Create": "order/create",
		"OrderController.Update": "order/update",
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	maps.Copy(cm.conventions, beegoConventions)
}

// ApplyConvention 应用命名约定
func (cm *ConventionManager) ApplyConvention(input string, convention NamingConvention) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 检查是否有直接映射
	if mapped, exists := cm.conventions[input]; exists {
		return mapped
	}

	// 应用命名约定
	switch convention {
	case CamelCase:
		return toCamelCase(input)
	case PascalCase:
		return toPascalCase(input)
	case SnakeCase:
		return toSnakeCase(input)
	case KebabCase:
		return toKebabCase(input)
	case LowerCase:
		return strings.ToLower(input)
	case UpperCase:
		return strings.ToUpper(input)
	case BeegoStandard:
		return cm.applyBeegoStandard(input)
	default:
		return input
	}
}

// applyBeegoStandard 应用Beego标准约定
func (cm *ConventionManager) applyBeegoStandard(input string) string {
	// 分析输入格式: "ControllerName.ActionName" 或单独的名称
	parts := strings.Split(input, ".")

	var controller, action string
	if len(parts) == 2 {
		controller = parts[0]
		action = parts[1]
	} else {
		controller = parts[0]
		action = "index" // 默认动作
	}

	// 应用映射规则
	controllerPath := cm.applyMappingRules(controller)
	actionPath := cm.applyMappingRules(action)

	// 组合路径
	if controllerPath != "" && actionPath != "" {
		return fmt.Sprintf("%s/%s", controllerPath, actionPath)
	} else if controllerPath != "" {
		return controllerPath
	}

	return input
}

// applyMappingRules 应用映射规则
func (cm *ConventionManager) applyMappingRules(input string) string {
	result := input

	// 按优先级顺序应用规则
	for _, rule := range cm.rules {
		if !rule.Enabled {
			continue
		}

		// 应用规则转换
		if rule.Transform != nil {
			transformed := rule.Transform(result)
			if transformed != result {
				result = transformed
				break // 应用第一个匹配的规则
			}
		}
	}

	return result
}

// AddConventionMapping 添加约定映射
func (cm *ConventionManager) AddConventionMapping(from, to string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conventions[from] = to
}

// AddMappingRule 添加映射规则
func (cm *ConventionManager) AddMappingRule(rule PathMappingRule) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 按优先级插入
	inserted := false
	for i, existing := range cm.rules {
		if rule.Priority > existing.Priority {
			cm.rules = append(cm.rules[:i], append([]PathMappingRule{rule}, cm.rules[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		cm.rules = append(cm.rules, rule)
	}
}

// SetDefaultConvention 设置默认约定
func (cm *ConventionManager) SetDefaultConvention(convention NamingConvention) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.defaultConvention = convention
}

// GetDefaultConvention 获取默认约定
func (cm *ConventionManager) GetDefaultConvention() NamingConvention {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.defaultConvention
}

// ============= 命名约定转换函数 =============

// camelToSnake CamelCase转snake_case
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			_, _ = result.WriteRune('_') // 忽略写入错误
		}
		if r >= 'A' && r <= 'Z' {
			_, _ = result.WriteRune(r - 'A' + 'a') // 忽略写入错误
		} else {
			_, _ = result.WriteRune(r) // 忽略写入错误
		}
	}
	return result.String()
}

// toCamelCase 转换为CamelCase
func toCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	if len(words) == 0 {
		return s
	}

	var result strings.Builder
	_, _ = result.WriteString(strings.ToLower(words[0])) // 忽略写入错误

	for _, word := range words[1:] {
		if len(word) > 0 {
			_, _ = result.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:])) // 忽略写入错误
		}
	}

	return result.String()
}

// toPascalCase 转换为PascalCase
func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			_, _ = result.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:])) // 忽略写入错误
		}
	}

	return result.String()
}

// toSnakeCase 转换为snake_case
func toSnakeCase(s string) string {
	return camelToSnake(s)
}

// toKebabCase 转换为kebab-case
func toKebabCase(s string) string {
	snake := camelToSnake(s)
	return strings.ReplaceAll(snake, "_", "-")
}

// ============= 性能优化和稳定性改进 =============

// TemplateCache 高性能模板缓存
type TemplateCache struct {
	cache           map[string]*CacheEntry
	maxSize         int
	cleanupTimer    *time.Timer
	cleanupInterval time.Duration
	mu              sync.RWMutex
}

// CacheEntry 缓存条目
type CacheEntry struct {
	template    *template.Template
	lastAccess  time.Time
	accessCount uint64
	createTime  time.Time
	size        int
}

// GetCacheStats 获取缓存统计信息（增强版，合并 Beego 兼容）
func (e *TemplateEngine) GetCacheStats() map[string]interface{} {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	stats := map[string]interface{}{
		"templates":      len(e.templates),
		"layouts":        len(e.layouts),
		"components":     len(e.components),
		"total":          len(e.templates) + len(e.layouts) + len(e.components),
		"view_paths":     e.viewPaths,
		"template_ext":   e.TemplateExt,
		"run_mode":       e.RunMode,
		"hot_reload":     e.enableReload,
		"auto_render":    e.AutoRender,
		"enable_gzip":    e.EnableGzip,
		"template_left":  e.delimLeft,
		"template_right": e.delimRight,
		"watched_files":  len(e.watchedFiles),
	}

	// 计算内存使用情况（估算）
	var totalMemory int
	for name, tmpl := range e.templates {
		if tmpl != nil {
			// 粗略估算：模板名长度 + 一些固定开销
			totalMemory += len(name)*2 + 1024 // 每个模板大约1KB基础开销
		}
	}
	stats["memory_estimate_bytes"] = totalMemory

	return stats
}

// OptimizeCache 优化缓存（增强版）
func (e *TemplateEngine) OptimizeCache() {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	if !e.enableCache {
		return
	}

	originalCount := len(e.templates) + len(e.layouts) + len(e.components)

	// 清理可能无效的模板引用
	cleanupCount := 0

	// 检查templates
	for name, tmpl := range e.templates {
		if tmpl == nil {
			delete(e.templates, name)
			cleanupCount++
		}
	}

	// 检查layouts
	for name, tmpl := range e.layouts {
		if tmpl == nil {
			delete(e.layouts, name)
			cleanupCount++
		}
	}

	// 检查components
	for name, tmpl := range e.components {
		if tmpl == nil {
			delete(e.components, name)
			cleanupCount++
		}
	}

	config.Debugf("Cache optimization completed: cleaned %d invalid entries, %d -> %d total entries",
		cleanupCount, originalCount, len(e.templates)+len(e.layouts)+len(e.components))

	// 触发GC以释放内存
	if cleanupCount > 0 {
		// runtime.GC() // 可以取消注释以强制GC
		config.Debugf("Recommended running GC after cache cleanup")
	}
}

// PreloadTemplates 预加载常用模板（性能优化）
func (e *TemplateEngine) PreloadTemplates(templateNames []string) error {
	if len(templateNames) == 0 {
		return nil
	}

	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	successCount := 0
	errorCount := 0

	for _, templateName := range templateNames {
		// 检查是否已经缓存
		if _, exists := e.templates[templateName]; exists {
			continue
		}

		// 尝试加载模板
		tmpl, err := e.loadSingleTemplate(templateName)
		if err != nil {
			config.Warnf("Failed to preload template %s: %v", templateName, err)
			errorCount++
			continue
		}

		e.templates[templateName] = tmpl
		successCount++
	}

	config.Infof("Preloaded %d templates successfully, %d failed", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("failed to preload %d out of %d templates", errorCount, len(templateNames))
	}

	return nil
}

// loadSingleTemplate 加载单个模板（内部方法）
func (e *TemplateEngine) loadSingleTemplate(templateName string) (*template.Template, error) {
	// 查找模板文件
	templateFile, err := e.FindTemplateFile(templateName)
	if err != nil {
		return nil, fmt.Errorf("template file not found: %s", templateName)
	}

	// 动态获取最新的合并函数
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	// 创建和解析模板
	tmpl := template.New(templateName).
		Delims(e.delimLeft, e.delimRight).
		Funcs(mergedFuncs)

	_, err = tmpl.ParseFiles(templateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templateFile, err)
	}

	return tmpl, nil
}

// GetMemoryUsage 获取内存使用情况（估算）
func (e *TemplateEngine) GetMemoryUsage() map[string]int64 {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	usage := map[string]int64{
		"template_count":  int64(len(e.templates)),
		"layout_count":    int64(len(e.layouts)),
		"component_count": int64(len(e.components)),
		"function_count":  int64(len(e.funcMap)),
	}

	// 粗略计算内存使用
	var templateMemory, layoutMemory, componentMemory, functionMemory int64

	for name := range e.templates {
		templateMemory += int64(len(name)*2 + 1024) // 基础估算
	}

	for name := range e.layouts {
		layoutMemory += int64(len(name)*2 + 1024)
	}

	for name := range e.components {
		componentMemory += int64(len(name)*2 + 1024)
	}

	for name := range e.funcMap {
		functionMemory += int64(len(name)*2 + 64) // 函数开销较小
	}

	usage["template_memory"] = templateMemory
	usage["layout_memory"] = layoutMemory
	usage["component_memory"] = componentMemory
	usage["function_memory"] = functionMemory
	usage["total_memory"] = templateMemory + layoutMemory + componentMemory + functionMemory

	return usage
}

// BulkRender 批量渲染模板（性能优化版）
func (e *TemplateEngine) BulkRender(requests []RenderRequest) (map[string]RenderResult, error) {
	if len(requests) == 0 {
		return make(map[string]RenderResult), nil
	}

	results := make(map[string]RenderResult, len(requests))

	// 并发渲染以提高性能
	const maxConcurrency = 10 // 限制并发数以避免资源耗尽
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var resultsMu sync.Mutex

	for i, request := range requests {
		wg.Add(1)
		go func(index int, req RenderRequest) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行渲染
			var result RenderResult
			result.Index = index
			result.TemplateName = req.TemplateName

			html, err := e.Render(req.TemplateName, req.Data)
			if err != nil {
				result.Error = err
				result.Success = false
			} else {
				result.HTML = html
				result.Success = true
			}

			// 安全地写入结果
			resultsMu.Lock()
			results[req.TemplateName] = result
			resultsMu.Unlock()
		}(i, request)
	}

	wg.Wait()

	config.Debugf("Bulk render completed: %d templates processed", len(results))
	return results, nil
}

// RenderRequest 渲染请求
type RenderRequest struct {
	TemplateName string
	Data         any
	LayoutName   string // 可选的布局
}

// RenderResult 渲染结果
type RenderResult struct {
	Index        int
	TemplateName string
	HTML         string
	Error        error
	Success      bool
	Duration     time.Duration
}

// ============= 稳定性改进 =============

// HealthCheck 健康检查
func (e *TemplateEngine) HealthCheck() error {
	// 检查基本配置
	if len(e.viewPaths) == 0 {
		return fmt.Errorf("no view paths configured")
	}

	// 检查模板目录是否存在
	for _, path := range e.viewPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("view path does not exist: %s", path)
		}
	}

	// 检查缓存状态
	if e.enableCache {
		stats := e.GetCacheStats()
		if stats["total"] == 0 {
			config.Warnf("Cache is enabled but empty")
		}
	}

	// 检查函数映射
	if len(e.funcMap) == 0 {
		config.Warnf("No template functions registered")
	}

	config.Debugf("Template engine health check passed")
	return nil
}

// Recovery 错误恢复
func (e *TemplateEngine) Recovery() error {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	config.Infof("Starting template engine recovery...")

	// 清理可能损坏的缓存
	originalCount := len(e.templates) + len(e.layouts) + len(e.components)

	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)

	// 重新注册函数
	e.registerDefaultFunctions()

	// 重新加载所有模板
	if err := e.loadAllTemplates(); err != nil {
		return fmt.Errorf("failed to reload templates during recovery: %w", err)
	}

	newCount := len(e.templates) + len(e.layouts) + len(e.components)
	config.Infof("Template engine recovery completed: %d -> %d total templates", originalCount, newCount)

	return nil
}

// ValidateConfiguration 验证配置
func (e *TemplateEngine) ValidateConfiguration() error {
	var errors []string

	// 检查必需的配置
	if len(e.viewPaths) == 0 {
		errors = append(errors, "viewPaths is empty")
	}

	if e.extension == "" {
		errors = append(errors, "extension is empty")
	}

	if e.delimLeft == "" || e.delimRight == "" {
		errors = append(errors, "delimiters are empty")
	}

	if e.delimLeft == e.delimRight {
		errors = append(errors, "left and right delimiters cannot be the same")
	}

	// 检查路径是否存在且可访问
	for _, path := range e.viewPaths {
		if _, err := os.Stat(path); err != nil {
			errors = append(errors, fmt.Sprintf("viewPath not accessible: %s (%v)", path, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	return nil
}

// RenderWithAutoDiscovery 自动发现并渲染模板
func RenderWithAutoDiscovery(templateName string, data any) (string, error) {
	manager := GetEnhancedTemplateManager()

	// 如果模板不存在，尝试重新发现
	if _, exists := manager.templatePaths[templateName]; !exists {
		if err := manager.DiscoverTemplates(); err != nil {
			config.Warnf("Failed to rediscover templates: %v", err)
		}
	}

	return manager.Render(templateName, data)
}

// RenderWithAutoInference Beego风格的自动推导渲染
func RenderWithAutoInference(controllerName, actionName string, data any) (string, error) {
	// 获取自动推导器
	inference := GetAutoTemplateInference()

	// 推导候选模板路径
	candidates := inference.InferTemplatePath(controllerName, actionName)
	if len(candidates) == 0 {
		return "", fmt.Errorf("no template candidates found for controller: %s, action: %s", controllerName, actionName)
	}

	// 获取模板管理器
	manager := GetEnhancedTemplateManager()

	// 尝试按优先级渲染候选模板
	var lastErr error
	for _, templatePath := range candidates {
		config.Debugf("Trying template path: %s", templatePath)

		result, err := manager.Render(templatePath, data)
		if err == nil {
			config.Debugf("Successfully rendered template: %s", templatePath)
			return result, nil
		}

		lastErr = err
		config.Debugf("Failed to render template %s: %v", templatePath, err)
	}

	// 如果所有候选模板都失败，返回最后一个错误
	return "", fmt.Errorf("failed to render any candidate template for %s.%s, last error: %w, candidates tried: %v",
		controllerName, actionName, lastErr, candidates)
}

// GetTemplateStatistics 获取模板统计信息
func GetTemplateStatistics() map[string]any {
	return GetEnhancedTemplateManager().GetTemplateInfo()
}

// ============= Beego 兼容方法 (从 BeegoTemplateEngine 合并) =============

// initBeegoFunctions 初始化Beego风格的模板函数
func (e *TemplateEngine) initBeegoFunctions() {
	// 获取现有的Beego函数
	beegoFuncs := GetBeegoTemplateFuncs()

	// 复制到当前引擎 - 确保Beego函数不被覆盖
	for name, fn := range beegoFuncs {
		e.funcMap[name] = fn
		config.Debugf("Registered Beego template function: %s", name)
	}

	// 特别确保关键函数的注册
	if dateformatFunc, exists := beegoFuncs["dateformat"]; exists {
		e.funcMap["dateformat"] = dateformatFunc
		config.Infof("✅ Critical function 'dateformat' registered successfully")
	} else {
		config.Errorf("❌ Critical function 'dateformat' not found in Beego functions")
	}

	// 添加Beego特有的额外函数
	e.funcMap["yield"] = func() string {
		return "{{.LayoutContent}}"
	}

	e.funcMap["content"] = func() string {
		return "{{.Content}}"
	}

	e.funcMap["partial"] = func(name string, data ...any) string {
		return fmt.Sprintf("{{template \"%s\" %v}}", name, data)
	}

	e.funcMap["block"] = func(name string, data ...any) string {
		return fmt.Sprintf("{{block \"%s\" %v}}{{end}}", name, data)
	}

	e.funcMap["section"] = func(name string) string {
		return fmt.Sprintf("{{define \"%s\"}}", name)
	}

	e.funcMap["endsection"] = func() string {
		return "{{end}}"
	}

	// Beego布局相关函数
	e.funcMap["layout"] = func(layoutName string) string {
		return fmt.Sprintf("{{template \"%s\" .}}", layoutName)
	}

	// 静态资源函数
	e.funcMap["static"] = func(path string) string {
		return "/static/" + strings.TrimPrefix(path, "/")
	}

	e.funcMap["css"] = func(path string) string {
		return fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\">", e.funcMap["static"].(func(string) string)(path))
	}

	e.funcMap["js"] = func(path string) string {
		return fmt.Sprintf("<script src=\"%s\"></script>", e.funcMap["static"].(func(string) string)(path))
	}

	// Beego路由函数
	e.funcMap["urlfor"] = func(controller, action string, params ...string) string {
		url := "/" + strings.ToLower(controller) + "/" + strings.ToLower(action)
		if len(params) > 0 {
			url += "?" + strings.Join(params, "&")
		}
		return url
	}

	// 开发/生产模式检查
	e.funcMap["isDev"] = func() bool {
		return e.RunMode == "dev"
	}

	e.funcMap["isProd"] = func() bool {
		return e.RunMode == "prod"
	}

	config.Infof("Initialized %d Beego-style template functions", len(e.funcMap))
}

// BuildAllTemplates 构建所有模板 (Beego风格)
func (e *TemplateEngine) BuildAllTemplates() error {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	config.Infof("🔄 Building all templates in Beego style...")

	// 清空现有缓存
	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)

	// 遍历所有视图路径
	for _, viewPath := range e.viewPaths {
		if err := e.buildTemplatesInPath(viewPath); err != nil {
			return fmt.Errorf("failed to build templates in path %s: %w", viewPath, err)
		}
	}

	config.Infof("✅ Successfully built %d templates", len(e.templates))
	return nil
}

// buildTemplatesInPath 构建指定路径下的所有模板
func (e *TemplateEngine) buildTemplatesInPath(viewPath string) error {
	return filepath.WalkDir(viewPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			config.Warnf("Template walk error at %s: %v", path, err)
			return nil // 忽略错误，继续处理
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名
		if !e.isTemplateFile(path) {
			return nil
		}

		// 计算相对路径作为模板名
		templateName, err := filepath.Rel(viewPath, path)
		if err != nil {
			config.Warnf("Failed to get relative path for %s: %v", path, err)
			return nil
		}

		// 统一路径分隔符
		templateName = filepath.ToSlash(templateName)

		// 移除扩展名
		templateName = e.removeExtension(templateName)

		// 构建模板
		if err := e.buildSingleTemplate(templateName, path); err != nil {
			config.Errorf("Failed to build template %s at %s: %v", templateName, path, err)
			return nil // 不中断构建过程
		}

		config.Debugf("Built template: %s -> %s", templateName, path)
		return nil
	})
}

// buildSingleTemplate 构建单个模板
func (e *TemplateEngine) buildSingleTemplate(templateName, filePath string) error {
	// 获取文件修改时间
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	// 检查是否需要重新构建
	if lastMod, exists := e.lastModTime[templateName]; exists {
		if !fileInfo.ModTime().After(lastMod) && e.RunMode == "prod" {
			return nil // 生产模式下不重建未修改的模板
		}
	}

	// 创建新模板
	tmpl := template.New(templateName)
	tmpl = tmpl.Delims(e.delimLeft, e.delimRight)
	tmpl = tmpl.Funcs(e.funcMap)

	// 解析模板文件
	tmpl, err = tmpl.ParseFiles(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", filePath, err)
	}

	// 处理模板依赖和包含
	if err := e.processTemplateIncludes(tmpl, filePath); err != nil {
		return fmt.Errorf("failed to process template includes for %s: %w", templateName, err)
	}

	// 缓存模板
	e.templates[templateName] = tmpl
	e.lastModTime[templateName] = fileInfo.ModTime()

	return nil
}

// processTemplateIncludes 处理模板包含和依赖 (Beego风格)
func (e *TemplateEngine) processTemplateIncludes(tmpl *template.Template, basePath string) error {
	baseDir := filepath.Dir(basePath)

	// 查找当前目录和父目录中的模板文件进行自动包含
	for _, ext := range e.TemplateExt {
		// 检查布局文件
		layoutPath := filepath.Join(baseDir, "layout"+ext)
		if _, err := os.Stat(layoutPath); err == nil {
			if _, err := tmpl.ParseFiles(layoutPath); err != nil {
				config.Warnf("Failed to include layout %s: %v", layoutPath, err)
			}
		}

		// 检查公共文件
		commonPath := filepath.Join(baseDir, "common"+ext)
		if _, err := os.Stat(commonPath); err == nil {
			if _, err := tmpl.ParseFiles(commonPath); err != nil {
				config.Warnf("Failed to include common %s: %v", commonPath, err)
			}
		}
	}

	return nil
}

// RenderTemplate 渲染模板 (Beego风格)
func (e *TemplateEngine) RenderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, err := e.GetTemplate(templateName)
	if err != nil {
		return "", err
	}

	// 准备渲染数据
	renderData := e.prepareBeegoRenderData(data)

	// 执行渲染
	var result strings.Builder
	if err := tmpl.Execute(&result, renderData); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	output := result.String()

	// 生产模式下启用Gzip压缩
	if e.RunMode == "prod" && e.EnableGzip {
		// 这里可以添加Gzip压缩逻辑
		config.Debugf("Template rendered with Gzip compression: %s", templateName)
	}

	return output, nil
}

// prepareBeegoRenderData 准备Beego风格渲染数据
func (e *TemplateEngine) prepareBeegoRenderData(data interface{}) map[string]interface{} {
	renderData := make(map[string]interface{})

	// 添加全局变量
	renderData["RunMode"] = e.RunMode
	renderData["TemplateLeft"] = e.delimLeft
	renderData["TemplateRight"] = e.delimRight
	renderData["EnableGzip"] = e.EnableGzip

	// 处理用户数据
	if data != nil {
		switch v := data.(type) {
		case map[string]interface{}:
			for k, val := range v {
				renderData[k] = val
			}
		default:
			renderData["Data"] = data
		}
	}

	return renderData
}

// shouldReloadTemplate 检查是否应该重新加载模板
func (e *TemplateEngine) shouldReloadTemplate(templateName string) bool {
	lastMod, exists := e.lastModTime[templateName]
	if !exists {
		return true
	}

	// 查找对应的文件路径
	for _, viewPath := range e.viewPaths {
		for _, ext := range e.TemplateExt {
			fullPath := filepath.Join(viewPath, templateName+ext)
			if fileInfo, err := os.Stat(fullPath); err == nil {
				return fileInfo.ModTime().After(lastMod)
			}
		}
	}

	return false
}

// reloadTemplate 重新加载模板
func (e *TemplateEngine) reloadTemplate(templateName string) error {
	// 查找模板文件
	for _, viewPath := range e.viewPaths {
		// 先检查是否已经带扩展名
		if filepath.Ext(templateName) != "" {
			// 如果已经带扩展名，直接使用
			fullPath := filepath.Join(viewPath, templateName)
			if _, err := os.Stat(fullPath); err == nil {
				return e.buildSingleTemplate(templateName, fullPath)
			}
		} else {
			// 如果没有扩展名，尝试添加各种扩展名
			for _, ext := range e.TemplateExt {
				fullPath := filepath.Join(viewPath, templateName+ext)
				if _, err := os.Stat(fullPath); err == nil {
					return e.buildSingleTemplate(templateName, fullPath)
				}
			}
		}
	}

	return fmt.Errorf("template file not found for %s", templateName)
}

// isTemplateFile 检查是否为模板文件
func (e *TemplateEngine) isTemplateFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	for _, templateExt := range e.TemplateExt {
		if ext == templateExt {
			return true
		}
	}
	return false
}

// removeExtension 移除文件扩展名
func (e *TemplateEngine) removeExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// AddFuncMap 添加模板函数 (Beego风格)
func (e *TemplateEngine) AddFuncMap(name string, fn interface{}) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	e.funcMap[name] = fn

	// 在开发模式下立即重建模板以包含新函数
	if e.RunMode == "dev" {
		go func() {
			if err := e.BuildAllTemplates(); err != nil {
				config.Errorf("Failed to rebuild templates after adding function %s: %v", name, err)
			}
		}()
	}
}

// SetRunMode 设置运行模式
func (e *TemplateEngine) SetRunMode(mode string) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	e.RunMode = mode

	if mode == "prod" {
		// 生产模式：禁用热重载，启用缓存
		e.enableReload = false
	} else {
		// 开发模式：启用热重载
		e.enableReload = true
	}

	config.Infof("Template engine run mode set to: %s", mode)
}

// ============= 高级功能集成方法 =============

// EnableAdvancedFeatures 启用高级功能（性能统计、缓存优化等）
func (e *TemplateEngine) EnableAdvancedFeatures() {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 启用性能统计
	if e.performanceStats == nil {
		e.performanceStats = &PerformanceStats{
			RenderTimes:  make(map[string]time.Duration),
			HotTemplates: make(map[string]int64),
		}
	}

	// 启用缓存优化器
	if e.cacheOptimizer == nil {
		e.cacheOptimizer = &CacheOptimizer{
			precompiledCache: make(map[string]*template.Template),
			frequencyTracker: make(map[string]int64),
			lastAccessTime:   make(map[string]time.Time),
			maxCacheSize:     1000,
			cleanupInterval:  time.Minute * 10,
		}
		go e.startCacheCleanup()
	}

	// 启用压缩管理器
	if e.compressionMgr == nil {
		e.compressionMgr = &CompressionManager{
			enableGzip:        e.EnableGzip,
			gzipLevel:         6, // 默认压缩级别
			compressionCache:  make(map[string][]byte),
			minSizeToCompress: 1024, // 1KB
		}
	}

	// 启用错误处理器
	if e.errorHandler == nil {
		e.errorHandler = &TemplateErrorHandler{
			showDebugInfo: e.RunMode == "dev",
			errorLog:      make([]TemplateError, 0),
			maxErrorLog:   100,
		}
	}

	config.Infof("Advanced features enabled for TemplateEngine")
}

// GetAdvancedPerformanceStats 获取高级性能统计信息
func (e *TemplateEngine) GetAdvancedPerformanceStats() map[string]any {
	if e.performanceStats == nil {
		return map[string]any{
			"render_count":   0,
			"cache_hit_rate": "0%",
			"error_message":  "Performance stats not enabled - call EnableAdvancedFeatures() first",
		}
	}

	e.performanceStats.mutex.RLock()
	defer e.performanceStats.mutex.RUnlock()

	// 计算缓存命中率
	var hitRate float64
	total := e.performanceStats.CacheHitCount + e.performanceStats.CacheMissCount
	if total > 0 {
		hitRate = float64(e.performanceStats.CacheHitCount) / float64(total) * 100
	}

	// 获取最热门的模板
	var hotTemplate string
	var maxCount int64
	for template, count := range e.performanceStats.HotTemplates {
		if count > maxCount {
			maxCount = count
			hotTemplate = template
		}
	}

	return map[string]any{
		"render_count":        e.performanceStats.RenderCount,
		"total_render_time":   e.performanceStats.TotalRenderTime.String(),
		"average_render_time": e.performanceStats.AverageRenderTime.String(),
		"cache_hit_count":     e.performanceStats.CacheHitCount,
		"cache_miss_count":    e.performanceStats.CacheMissCount,
		"cache_hit_rate":      fmt.Sprintf("%.2f%%", hitRate),
		"template_errors":     e.performanceStats.TemplateErrors,
		"last_render_time":    e.performanceStats.LastRenderTime,
		"hot_template":        hotTemplate,
		"hot_template_count":  maxCount,
		"render_times":        e.performanceStats.RenderTimes,
	}
}

// startCacheCleanup 启动缓存清理定时器
func (e *TemplateEngine) startCacheCleanup() {
	if e.cacheOptimizer == nil {
		return
	}

	ticker := time.NewTicker(e.cacheOptimizer.cleanupInterval)
	e.cacheOptimizer.cleanupTicker = ticker

	go func() {
		for range ticker.C {
			e.cleanupAdvancedCache()
		}
	}()
}

// cleanupAdvancedCache 清理高级缓存
func (e *TemplateEngine) cleanupAdvancedCache() {
	if e.cacheOptimizer == nil {
		return
	}

	e.cacheOptimizer.mutex.Lock()
	defer e.cacheOptimizer.mutex.Unlock()

	cutoffTime := time.Now().Add(-e.cacheOptimizer.cleanupInterval * 2)
	deleted := 0

	for templateName, lastAccess := range e.cacheOptimizer.lastAccessTime {
		if lastAccess.Before(cutoffTime) {
			delete(e.cacheOptimizer.precompiledCache, templateName)
			delete(e.cacheOptimizer.lastAccessTime, templateName)
			delete(e.cacheOptimizer.frequencyTracker, templateName)
			deleted++
		}
	}

	if deleted > 0 {
		config.Infof("Cleaned up %d expired templates from advanced cache", deleted)
	}
}

// recordRenderTime 记录渲染时间（用于性能统计）
func (e *TemplateEngine) recordRenderTime(templateName string, duration time.Duration) {
	if e.performanceStats == nil {
		return
	}

	e.performanceStats.mutex.Lock()
	defer e.performanceStats.mutex.Unlock()

	e.performanceStats.TotalRenderTime += duration
	e.performanceStats.RenderTimes[templateName] = duration
	if e.performanceStats.RenderCount > 0 {
		e.performanceStats.AverageRenderTime = time.Duration(int64(e.performanceStats.TotalRenderTime) / e.performanceStats.RenderCount)
	}
	e.performanceStats.LastRenderTime = time.Now()
}

// ============= 路径解析辅助函数 =============

// resolveViewPaths 解析视图路径列表，将相对路径转换为绝对路径
func resolveViewPaths(viewPaths []string) ([]string, error) {
	resolvedPaths := make([]string, 0, len(viewPaths))

	for _, path := range viewPaths {
		absolutePath, err := resolveAbsolutePath(path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve view path '%s': %w", path, err)
		}
		resolvedPaths = append(resolvedPaths, absolutePath)
	}

	return resolvedPaths, nil
}

// resolveAbsolutePath 解析单个路径，将相对路径转换为绝对路径
func resolveAbsolutePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// 将相对路径转换为绝对路径
	absolutePath := filepath.Join(cwd, path)
	return filepath.Clean(absolutePath), nil
}

// getCurrentWorkingDir 获取当前工作目录
func getCurrentWorkingDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return cwd
}

// loadCurrentTemplateConfig 加载当前的模板配置
// 尝试从配置文件加载最新的模板配置
func loadCurrentTemplateConfig() (*config.TemplateConfig, error) {
	// 尝试从配置文件加载
	if cfg, err := config.GetTemplateConfig(); err == nil && cfg != nil {
		config.Infof("✅ Loaded template config from file")
		return cfg, nil
	}

	// 如果有全局配置，使用全局配置
	if config.GlobalTemplate != nil {
		config.Infof("✅ Using global template config")
		return config.GlobalTemplate, nil
	}

	// 最后使用默认配置
	cfg := DefaultTemplateConfig()
	config.Infof("✅ Using default template config")
	return cfg, nil
}
