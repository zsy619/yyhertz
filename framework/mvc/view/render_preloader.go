package view

import (
	"fmt"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 模板预加载机制 =============

// TemplatePreloader 模板预加载器
type TemplatePreloader struct {
	engine      *TemplateEngine
	preloadList []string          // 预加载模板列表
	layoutPairs map[string]string // 模板-布局配对
	enabled     bool
	mu          sync.RWMutex
}

// NewTemplatePreloader 创建模板预加载器
func NewTemplatePreloader(engine *TemplateEngine) *TemplatePreloader {
	return &TemplatePreloader{
		engine:      engine,
		preloadList: []string{},
		layoutPairs: make(map[string]string),
		enabled:     true,
	}
}

// AddPreloadTemplate 添加预加载模板
func (tp *TemplatePreloader) AddPreloadTemplate(templateName string) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// 避免重复添加
	for _, existing := range tp.preloadList {
		if existing == templateName {
			return
		}
	}

	tp.preloadList = append(tp.preloadList, templateName)
	config.Debugf("Added template to preload list: %s", templateName)
}

// AddPreloadTemplateWithLayout 添加带布局的预加载模板
func (tp *TemplatePreloader) AddPreloadTemplateWithLayout(templateName, layoutName string) {
	tp.AddPreloadTemplate(templateName)

	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.layoutPairs[templateName] = layoutName
	config.Debugf("Added template-layout pair: %s@%s", templateName, layoutName)
}

// PreloadAll 预加载所有模板
func (tp *TemplatePreloader) PreloadAll() error {
	if !tp.enabled {
		config.Debugf("Template preloader is disabled")
		return nil
	}

	tp.mu.RLock()
	preloadList := make([]string, len(tp.preloadList))
	copy(preloadList, tp.preloadList)
	layoutPairs := make(map[string]string)
	for k, v := range tp.layoutPairs {
		layoutPairs[k] = v
	}
	tp.mu.RUnlock()

	if len(preloadList) == 0 {
		config.Debugf("No templates to preload")
		return nil
	}

	config.Infof("Starting template preloading: %d templates", len(preloadList))
	startTime := time.Now()

	successCount := 0
	errorCount := 0

	for _, templateName := range preloadList {
		// 检查是否需要预加载布局组合
		if layoutName, hasLayout := layoutPairs[templateName]; hasLayout {
			err := tp.preloadTemplateWithLayout(templateName, layoutName)
			if err != nil {
				config.Warnf("Failed to preload template with layout %s@%s: %v", templateName, layoutName, err)
				errorCount++
			} else {
				successCount++
			}
		} else {
			err := tp.preloadSingleTemplate(templateName)
			if err != nil {
				config.Warnf("Failed to preload template %s: %v", templateName, err)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	elapsed := time.Since(startTime)
	config.Infof("Template preloading completed: %d success, %d failed, took %v",
		successCount, errorCount, elapsed)

	if errorCount > 0 {
		return fmt.Errorf("failed to preload %d out of %d templates", errorCount, len(preloadList))
	}

	return nil
}

// preloadSingleTemplate 预加载单个模板
func (tp *TemplatePreloader) preloadSingleTemplate(templateName string) error {
	// 检查是否已缓存
	cacheKey := DefaultCacheKeyManager.GenerateTemplateKey(templateName)
	if tp.engine.enableCache {
		tp.engine.templateMutex.RLock()
		if _, exists := tp.engine.templates[cacheKey]; exists {
			tp.engine.templateMutex.RUnlock()
			config.Debugf("Template %s already cached, skipping preload", templateName)
			return nil
		}
		tp.engine.templateMutex.RUnlock()
	}

	// 预加载模板
	_, err := tp.engine.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("failed to preload template %s: %w", templateName, err)
	}

	config.Debugf("Successfully preloaded template: %s", templateName)
	return nil
}

// preloadTemplateWithLayout 预加载带布局的模板
func (tp *TemplatePreloader) preloadTemplateWithLayout(templateName, layoutName string) error {
	// 检查是否已缓存
	cacheKey := DefaultCacheKeyManager.GenerateLayoutKey(templateName, layoutName)
	if tp.engine.enableCache {
		tp.engine.templateMutex.RLock()
		if _, exists := tp.engine.templates[cacheKey]; exists {
			tp.engine.templateMutex.RUnlock()
			config.Debugf("Template with layout %s@%s already cached, skipping preload", templateName, layoutName)
			return nil
		}
		tp.engine.templateMutex.RUnlock()
	}

	// 预加载模板
	_, err := tp.engine.GetTemplateWithLayout(templateName, layoutName)
	if err != nil {
		return fmt.Errorf("failed to preload template with layout %s@%s: %w", templateName, layoutName, err)
	}

	config.Debugf("Successfully preloaded template with layout: %s@%s", templateName, layoutName)
	return nil
}

// GetPreloadStats 获取预加载统计
func (tp *TemplatePreloader) GetPreloadStats() map[string]any {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	return map[string]any{
		"enabled":            tp.enabled,
		"preload_list_count": len(tp.preloadList),
		"layout_pairs_count": len(tp.layoutPairs),
		"preload_list":       tp.preloadList,
		"layout_pairs":       tp.layoutPairs,
	}
}

// SetEnabled 设置预加载器启用状态
func (tp *TemplatePreloader) SetEnabled(enabled bool) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.enabled = enabled
	config.Debugf("Template preloader enabled: %v", enabled)
}

// Clear 清空预加载列表
func (tp *TemplatePreloader) Clear() {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.preloadList = []string{}
	tp.layoutPairs = make(map[string]string)
	config.Debugf("Template preloader list cleared")
}

// AddCommonTemplates 添加常用模板到预加载列表
func (tp *TemplatePreloader) AddCommonTemplates() {
	commonTemplates := []string{
		"index.html",
		"error.html",
		"404.html",
		"500.html",
		"login.html",
	}

	commonLayoutPairs := map[string]string{
		"index.html": "app.html",
		"error.html": "app.html",
		"login.html": "auth.html",
	}

	for _, templateName := range commonTemplates {
		tp.AddPreloadTemplate(templateName)
	}

	tp.mu.Lock()
	for template, layout := range commonLayoutPairs {
		tp.layoutPairs[template] = layout
	}
	tp.mu.Unlock()

	config.Debugf("Added %d common templates to preload list", len(commonTemplates))
}

// 全局预加载器实例
var (
	globalPreloader *TemplatePreloader
	preloaderOnce   sync.Once
)

// GetGlobalTemplatePreloader 获取全局模板预加载器
func GetGlobalTemplatePreloader() *TemplatePreloader {
	preloaderOnce.Do(func() {
		defaultEngine := GetDefaultEngine()
		globalPreloader = NewTemplatePreloader(defaultEngine)
	})
	return globalPreloader
}

// PreloadCommonTemplates 预加载常用模板（便捷函数）
func PreloadCommonTemplates() error {
	preloader := GetGlobalTemplatePreloader()
	preloader.AddCommonTemplates()
	return preloader.PreloadAll()
}

// AddPreloadTemplate 添加预加载模板（便捷函数）
func AddPreloadTemplate(templateName string) {
	GetGlobalTemplatePreloader().AddPreloadTemplate(templateName)
}

// AddPreloadTemplateWithLayout 添加带布局的预加载模板（便捷函数）
func AddPreloadTemplateWithLayout(templateName, layoutName string) {
	GetGlobalTemplatePreloader().AddPreloadTemplateWithLayout(templateName, layoutName)
}