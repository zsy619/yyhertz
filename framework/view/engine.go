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
	"path/filepath"
	"strings"
	"sync"

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
	currentTheme string                  // 当前主题
	themes       map[string]*ThemeConfig // 主题配置
}

// ThemeConfig 主题配置
type ThemeConfig struct {
	Name          string            `json:"name"`
	ViewPaths     []string          `json:"view_paths"`
	LayoutPath    string            `json:"layout_path"`
	ComponentPath string            `json:"component_path"`
	StaticPath    string            `json:"static_path"`
	Enabled       bool              `json:"enabled"`
	Default       bool              `json:"default"`
	Variables     map[string]string `json:"variables"`
}

// TemplateConfig 模板引擎配置
type TemplateConfig struct {
	ViewPaths      []string                `json:"view_paths" yaml:"view_paths"`
	LayoutPath     string                  `json:"layout_path" yaml:"layout_path"`
	ComponentPath  string                  `json:"component_path" yaml:"component_path"`
	Extension      string                  `json:"extension" yaml:"extension"`
	DelimLeft      string                  `json:"delim_left" yaml:"delim_left"`
	DelimRight     string                  `json:"delim_right" yaml:"delim_right"`
	EnableCache    bool                    `json:"enable_cache" yaml:"enable_cache"`
	EnableReload   bool                    `json:"enable_reload" yaml:"enable_reload"`
	EnableCompress bool                    `json:"enable_compress" yaml:"enable_compress"`
	CurrentTheme   string                  `json:"current_theme" yaml:"current_theme"`
	Themes         map[string]*ThemeConfig `json:"themes" yaml:"themes"`
}

// DefaultTemplateConfig 默认模板配置
func DefaultTemplateConfig() *TemplateConfig {
	return &TemplateConfig{
		ViewPaths:      []string{"views", "templates"},
		LayoutPath:     "views/layouts",
		ComponentPath:  "views/components",
		Extension:      ".html",
		DelimLeft:      "{{",
		DelimRight:     "}}",
		EnableCache:    true,
		EnableReload:   true,
		EnableCompress: false,
		CurrentTheme:   "default",
		Themes: map[string]*ThemeConfig{
			"default": {
				Name:          "default",
				ViewPaths:     []string{"views"},
				LayoutPath:    "views/layouts",
				ComponentPath: "views/components",
				StaticPath:    "static",
				Enabled:       true,
				Default:       true,
				Variables:     make(map[string]string),
			},
		},
	}
}

var (
	defaultEngine *TemplateEngine
	engineOnce    sync.Once
	engineMutex   sync.Mutex
)

// GetDefaultEngine 获取默认模板引擎
func GetDefaultEngine() *TemplateEngine {
	engineOnce.Do(func() {
		engineMutex.Lock()
		defer engineMutex.Unlock()

		cfg := DefaultTemplateConfig()
		var err error
		defaultEngine, err = NewTemplateEngine(cfg)
		if err != nil {
			config.Fatalf("Failed to initialize default template engine: %v", err)
		}
	})
	return defaultEngine
}

// NewTemplateEngine 创建新的模板引擎
func NewTemplateEngine(cfg *TemplateConfig) (*TemplateEngine, error) {
	if cfg == nil {
		cfg = DefaultTemplateConfig()
	}

	engine := &TemplateEngine{
		viewPaths:      cfg.ViewPaths,
		layoutPath:     cfg.LayoutPath,
		componentPath:  cfg.ComponentPath,
		extension:      cfg.Extension,
		delimLeft:      cfg.DelimLeft,
		delimRight:     cfg.DelimRight,
		enableCache:    cfg.EnableCache,
		enableReload:   cfg.EnableReload,
		enableCompress: cfg.EnableCompress,
		currentTheme:   cfg.CurrentTheme,

		templates:  make(map[string]*template.Template),
		layouts:    make(map[string]*template.Template),
		components: make(map[string]*template.Template),
		watchPaths: make(map[string]bool),
		funcMap:    make(template.FuncMap),
		themes:     cfg.Themes,
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
	// 从现有的TemplateFuncs复制
	for name, fn := range TemplateFuncs {
		e.funcMap[name] = fn
	}

	// 添加新的增强函数
	e.funcMap["include"] = e.includeTemplate
	e.funcMap["component"] = e.renderComponent
	e.funcMap["theme"] = e.getThemeVariable
	e.funcMap["asset"] = e.getAssetURL
	e.funcMap["url"] = e.buildURL
	e.funcMap["csrf"] = e.getCSRFToken
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

// AddFunction 添加自定义模板函数
func (e *TemplateEngine) AddFunction(name string, fn any) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	e.funcMap[name] = fn

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

// AddTheme 添加新主题
func (e *TemplateEngine) AddTheme(name string, theme *ThemeConfig) error {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	if e.themes == nil {
		e.themes = make(map[string]*ThemeConfig)
	}

	theme.Name = name
	e.themes[name] = theme

	config.Infof("Added theme: %s", name)
	return nil
}

// GetTheme 获取主题配置
func (e *TemplateEngine) GetTheme(name string) (*ThemeConfig, bool) {
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

				if _, err := tmpl.ParseFiles(path); err != nil {
					config.Errorf("Failed to parse template %s: %v", path, err)
					return nil
				}

				e.templates[templateName] = tmpl
				config.Debugf("Loaded template: %s", templateName)
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
