package template

import (
	"fmt"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/view"
)

// TemplateManager 模板管理器（单实例）
type TemplateManager struct {
	engine *view.TemplateEngine
	config *view.TemplateConfig
	mutex  sync.RWMutex
}

var (
	templateManager *TemplateManager
	templateOnce    sync.Once
	configEngine    = config.GetViperConfigManagerWithName(config.TemplateConfigName)
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
	// 使用默认模板配置（暂时简化，避免配置读取问题）
	templateConfig := view.DefaultTemplateConfig()

	// 创建模板引擎
	engine, err := view.NewTemplateEngine(templateConfig)
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
func loadTemplateConfigFromFile() (*view.TemplateConfig, error) {
	cfg := &view.TemplateConfig{}

	// 基础配置 - 使用默认值的方式
	cfg.ViewPaths = []string{"views", "templates"}
	cfg.LayoutPath = "views/layouts"
	cfg.ComponentPath = "views/components"
	cfg.Extension = ".html"
	cfg.DelimLeft = "{{"
	cfg.DelimRight = "}}"
	cfg.EnableCache = true
	cfg.EnableReload = true
	cfg.EnableCompress = false
	cfg.CurrentTheme = "default"
	cfg.Themes = view.DefaultTemplateConfig().Themes

	// 尝试从配置文件读取模板配置 (如果配置文件存在的话)
	if viewPaths := configEngine.GetStringSlice("template.view_paths"); len(viewPaths) > 0 {
		cfg.ViewPaths = viewPaths
	}

	if layoutPath := configEngine.GetString("template.layout_path"); layoutPath != "" {
		cfg.LayoutPath = layoutPath
	}

	if componentPath := configEngine.GetString("template.component_path"); componentPath != "" {
		cfg.ComponentPath = componentPath
	}

	if extension := configEngine.GetString("template.extension"); extension != "" {
		cfg.Extension = extension
	}

	if delimLeft := configEngine.GetString("template.delim_left"); delimLeft != "" {
		cfg.DelimLeft = delimLeft
	}

	if delimRight := configEngine.GetString("template.delim_right"); delimRight != "" {
		cfg.DelimRight = delimRight
	}

	// 性能配置
	cfg.EnableCache = configEngine.GetBool("template.enable_cache")
	cfg.EnableReload = configEngine.GetBool("template.enable_reload")
	cfg.EnableCompress = configEngine.GetBool("template.enable_compress")

	// 主题配置
	if currentTheme := configEngine.GetString("template.current_theme"); currentTheme != "" {
		cfg.CurrentTheme = currentTheme
	}

	config.Infof("Loaded template configuration from config file")
	return cfg, nil
}

// GetEngine 获取模板引擎
func (tm *TemplateManager) GetEngine() *view.TemplateEngine {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.engine
}

// GetConfig 获取模板配置
func (tm *TemplateManager) GetConfig() *view.TemplateConfig {
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
func (tm *TemplateManager) AddTheme(name string, theme *view.ThemeConfig) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	return tm.engine.AddTheme(name, theme)
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
	newEngine, err := view.NewTemplateEngine(newConfig)
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

// ============= 便捷函数 =============

// Render 渲染模板（使用默认管理器）
func Render(templateName string, data any) (string, error) {
	return GetTemplateManager().Render(templateName, data)
}

// RenderWithLayout 使用布局渲染模板（使用默认管理器）
func RenderWithLayout(templateName, layoutName string, data any) (string, error) {
	return GetTemplateManager().RenderWithLayout(templateName, layoutName, data)
}

// RenderComponent 渲染组件（使用默认管理器）
func RenderComponent(componentName string, data any) (string, error) {
	return GetTemplateManager().RenderComponent(componentName, data)
}

// SetCurrentTheme 设置当前主题（使用默认管理器）
func SetCurrentTheme(themeName string) error {
	return GetTemplateManager().SetTheme(themeName)
}

// GetCurrentTheme 获取当前主题（使用默认管理器）
func GetCurrentTheme() string {
	return GetTemplateManager().GetCurrentTheme()
}

// AddTemplateFunction 添加模板函数（使用默认管理器）
func AddTemplateFunction(name string, fn any) {
	GetTemplateManager().AddFunction(name, fn)
}
