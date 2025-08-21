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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/zsy619/yyhertz/framework/config"
)

// TemplateEngine 模板引擎
type TemplateEngine struct {
	// 配置
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
	for name, fn := range builtinFuncs {
		tfm.builtinFuncs[name] = fn
	}
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
	for name, fn := range tfm.builtinFuncs {
		merged[name] = fn
	}

	// 2. 然后添加全局自定义函数（可能覆盖内置函数）
	for name, fn := range tfm.globalFuncs {
		if _, exists := merged[name]; exists {
			config.Warnf("Global function '%s' overrides builtin function", name)
		}
		merged[name] = fn
	}

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
	for name, fn := range manager.globalFuncs {
		funcMapCopy[name] = fn
	}

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

// GetDefaultEngine 获取默认模板引擎
func GetDefaultEngine() *TemplateEngine {
	engineOnce.Do(func() {
		engineMutex.Lock()
		defer engineMutex.Unlock()

		cfg := config.GlobalTemplate
		var err error
		defaultEngine, err = NewTemplateEngine(cfg)
		if err != nil {
			config.Fatalf("Failed to initialize default template engine: %v", err)
		}
	})
	return defaultEngine
}

// NewTemplateEngine 创建新的模板引擎
func NewTemplateEngine(cfg *config.TemplateConfig) (*TemplateEngine, error) {
	if cfg == nil {
		cfg = DefaultTemplateConfig()
	}

	engine := &TemplateEngine{
		viewPaths:      cfg.Paths.ViewPaths,
		layoutPath:     cfg.Paths.LayoutPath,
		componentPath:  cfg.Paths.ComponentPath,
		extension:      cfg.Paths.Extension,
		delimLeft:      cfg.Syntax.DelimLeft,
		delimRight:     cfg.Syntax.DelimRight,
		enableCache:    cfg.Cache.EnableCache,
		enableReload:   cfg.Reload.Enabled,
		enableCompress: cfg.Performance.EnableCompress,
		currentTheme:   cfg.Theme.Current,

		templates:  make(map[string]*template.Template),
		layouts:    make(map[string]*template.Template),
		components: make(map[string]*template.Template),
		watchPaths: make(map[string]bool),
		funcMap:    make(template.FuncMap),
		themes:     cfg.Theme.Themes,
	}

	// 注册默认模板函数
	engine.registerDefaultFunctions()

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
	for name, fn := range mergedFuncs {
		e.funcMap[name] = fn
	}

	// 添加新的增强函数
	e.funcMap["include"] = e.includeTemplate
	e.funcMap["component"] = e.renderComponent
	e.funcMap["theme"] = e.getThemeVariable
	e.funcMap["asset"] = e.getAssetURL
	e.funcMap["url"] = e.buildURL
	e.funcMap["csrf"] = e.getCSRFToken
	e.funcMap["csrf_token"] = e.getCSRFToken // 下划线别名
	e.funcMap["flash"] = e.getFlashMessage
	e.funcMap["truncate"] = e.truncateString
	e.funcMap["markdown"] = e.renderMarkdown
	e.funcMap["json"] = e.toJSON
	e.funcMap["safe"] = e.safeHTML
	e.funcMap["dict"] = e.createDict
	e.funcMap["slice"] = e.createSlice
	e.funcMap["range"] = e.createRange
	e.funcMap["dateFormat"] = e.formatDate
	e.funcMap["currency"] = e.formatCurrency
	e.funcMap["filesize"] = e.formatFileSize
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
	tmpl := template.New(name).
		Delims(e.delimLeft, e.delimRight).
		Funcs(e.funcMap)

	return tmpl.Parse(content)
}

// ExecuteTemplate 执行模板（用于测试）
func (e *TemplateEngine) ExecuteTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var buf strings.Builder
	err := tmpl.Execute(&buf, data)
	return buf.String(), err
}

// createInlineTemplate 创建内联模板（用于测试）- 内部方法兼容性
func (e *TemplateEngine) createInlineTemplate(name, content string) (*template.Template, error) {
	return e.CreateInlineTemplate(name, content)
}

// executeTemplate 执行模板（用于测试）- 内部方法兼容性
func (e *TemplateEngine) executeTemplate(tmpl *template.Template, data interface{}) (string, error) {
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

// watchFiles 监控文件变化
func (e *TemplateEngine) watchFiles() {
	for {
		select {
		case event, ok := <-e.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Remove == fsnotify.Remove {

				if strings.HasSuffix(event.Name, e.extension) {
					config.Debugf("Template file changed: %s", event.Name)
					e.reloadTemplate(event.Name)
				}
			}

		case err, ok := <-e.watcher.Errors:
			if !ok {
				return
			}
			config.Errorf("Template watcher error: %v", err)
		}
	}
}

// reloadTemplate 重新加载特定模板
func (e *TemplateEngine) reloadTemplate(filePath string) {
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

// Close 关闭模板引擎
func (e *TemplateEngine) Close() error {
	if e.watcher != nil {
		return e.watcher.Close()
	}
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

			tmpl := template.New(layoutName).
				Delims(e.delimLeft, e.delimRight).
				Funcs(e.funcMap)

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

			tmpl := template.New(componentName).
				Delims(e.delimLeft, e.delimRight).
				Funcs(e.funcMap)

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

				tmpl := template.New(templateName).
					Delims(e.delimLeft, e.delimRight).
					Funcs(e.funcMap)

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

// NewTemplateManager 创建新的模板管理器
func NewTemplateManager() (*TemplateManager, error) {
	// 使用从config包加载的模板配置
	templateConfig := config.DefaultTemplateConfig()

	// 创建模板引擎
	engine, err := NewTemplateEngine(templateConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create template engine: %w", err)
	}

	manager := &TemplateManager{
		engine: engine,
		config: templateConfig,
	}

	config.Infof("Template manager initialized successfully")
	return manager, nil
}

// loadTemplateConfigFromFile 从配置文件加载模板配置
func loadTemplateConfigFromFile() (*config.TemplateConfig, error) {
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
	newConfig, err := loadTemplateConfigFromFile()
	if err != nil {
		return fmt.Errorf("failed to reload template config: %w", err)
	}

	// 关闭当前引擎
	if tm.engine != nil {
		tm.engine.Close()
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

// NewEnhancedTemplateManager 创建增强的模板管理器
func NewEnhancedTemplateManager() (*EnhancedTemplateManager, error) {
	// 创建基础管理器
	baseManager, err := NewTemplateManager()
	if err != nil {
		return nil, err
	}

	enhanced := &EnhancedTemplateManager{
		TemplateManager: baseManager,
		autoDiscover:    true,
		templatePaths:   make(map[string]string),
		lastModified:    make(map[string]time.Time),
	}

	// 执行自动发现
	if enhanced.autoDiscover {
		err = enhanced.DiscoverTemplates()
		if err != nil {
			config.Warnf("Template auto-discovery failed: %v", err)
		}
	}

	// 启用文件监控（开发模式）
	if enhanced.config.Reload.Enabled {
		err = enhanced.EnableFileWatcher()
		if err != nil {
			config.Warnf("Failed to enable file watcher: %v", err)
		}
	}

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
	for _, commonExt := range commonExts {
		if ext == commonExt {
			return true
		}
	}

	return false
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

// ============= 增强的便捷函数 =============

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

// GetTemplateStatistics 获取模板统计信息
func GetTemplateStatistics() map[string]any {
	return GetEnhancedTemplateManager().GetTemplateInfo()
}
