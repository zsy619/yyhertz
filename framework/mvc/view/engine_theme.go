package view

import (
	"fmt"
	"html/template"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 主题管理相关功能 =============

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