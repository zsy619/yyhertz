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
	"maps"
	"os"
	"path/filepath"
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

// DefaultTemplateConfig 默认模板配置（向后兼容）
// 该函数返回 config 包中的默认配置
func DefaultTemplateConfig() *config.TemplateConfig {
	return config.GlobalTemplate
}

var (
	defaultEngine *TemplateEngine
	engineOnce    sync.Once
	engineMutex   sync.Mutex

	// 全局引擎统一管理
	unifiedEngineInstance *TemplateEngine
	unifiedEngineOnce     sync.Once
	unifiedEngineMutex    sync.RWMutex
)

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

// ============= 测试辅助方法 =============

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
// 路径映射和约定管理功能已移至 engine_path.go 文件中

// ============= 路径映射和命名约定支持 =============

// ============= 性能优化和稳定性改进 =============

// GetCacheStats 方法已移至 engine_performance.go 文件中

// OptimizeCache 方法已移至 engine_performance.go 文件中

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

// GetTemplateStatistics 获取模板统计信息
func GetTemplateStatistics() map[string]any {
	return GetEnhancedTemplateManager().GetTemplateInfo()
}

// removeExtension 移除文件扩展名
func (e *TemplateEngine) removeExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
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
// 性能监控功能已移至 engine_performance.go 文件中
func (e *TemplateEngine) EnableAdvancedFeatures() {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 启用错误处理器
	if e.errorHandler == nil {
		e.errorHandler = &TemplateErrorHandler{
			showDebugInfo: e.RunMode == "dev",
			errorLog:      make([]TemplateError, 0),
			maxErrorLog:   100,
		}
	}

	// 启用性能监控功能
	e.enablePerformanceFeatures()

	config.Infof("Advanced features enabled for TemplateEngine")
}
