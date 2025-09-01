package view

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// compressHTML 压缩HTML
func (e *TemplateEngine) compressHTML(html string) string {
	// 简单的HTML压缩：移除多余的空白字符
	lines := strings.Split(html, "\n")
	compressed := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			compressed = append(compressed, trimmed)
		}
	}

	return strings.Join(compressed, "\n")
}

// checkReloadNeeded 检查模板是否需要重新加载（避免与engine_template.go冲突）
func (e *TemplateEngine) checkReloadNeeded(templateName string) bool {
	if !e.enableReload {
		return false
	}

	templatePath, err := e.FindTemplateFile(templateName)
	if err != nil {
		return false
	}

	// 获取文件的修改时间
	fileInfo, err := os.Stat(templatePath)
	if err != nil {
		return false
	}

	_ = fileInfo.ModTime().Unix() // 读取但不使用，避免未使用变量警告
	
	// 简化版本：总是返回true（因为原结构体没有lastModified字段）
	// TODO: 如果需要更精确的热重载控制，可以在TemplateEngine中添加lastModified字段
	return true
}

// reloadTemplateFromDisk 重新加载模板（避免与engine_template.go冲突）
func (e *TemplateEngine) reloadTemplateFromDisk(templateName string) error {
	config.Debugf("🔄 Reloading template: %s", templateName)
	
	// 从缓存中移除旧模板
	cacheKey := DefaultCacheKeyManager.GenerateTemplateKey(templateName)
	e.templateMutex.Lock()
	delete(e.templates, cacheKey)
	e.templateMutex.Unlock()

	// 重新加载模板
	_, err := e.GetTemplate(templateName)
	if err != nil {
		config.Errorf("Failed to reload template %s: %v", templateName, err)
		return err
	}

	// 简化版本：不更新修改时间（因为原结构体没有lastModified字段）
	// TODO: 如果需要跟踪修改时间，可以在TemplateEngine中添加lastModified字段

	config.Debugf("✅ Template reloaded successfully: %s", templateName)
	return nil
}

// setupDefaultTemplateFunctions 注册默认函数（避免与engine.go冲突）
func (e *TemplateEngine) setupDefaultTemplateFunctions() {
	if e.funcMap == nil {
		e.funcMap = make(map[string]any)
	}

	// 基础函数
	e.funcMap["now"] = func() time.Time { return time.Now() }
	e.funcMap["dateformat"] = e.formatDate
	e.funcMap["formatFileSize"] = e.formatFileSize
	e.funcMap["safeHTML"] = e.safeHTML
	e.funcMap["json"] = e.toJSON
	e.funcMap["dict"] = e.createDict
	e.funcMap["slice"] = e.createSlice
	e.funcMap["range"] = e.createRange
	
	// CSRF相关
	e.funcMap["csrf_token"] = e.getCSRFToken
	
	// 主题和资源
	e.funcMap["theme_var"] = e.getThemeVariable
	e.funcMap["asset_url"] = e.getAssetURL
	e.funcMap["url"] = e.buildURL
	
	// 模板包含
	e.funcMap["include"] = e.includeTemplate
	e.funcMap["component"] = e.renderComponent
	
	// 文本处理
	e.funcMap["truncate"] = e.truncateString
	e.funcMap["markdown"] = e.renderMarkdown
	
	// 货币格式化
	e.funcMap["currency"] = e.formatCurrency
	
	// Flash消息
	e.funcMap["flash"] = e.getFlashMessage

	config.Debugf("Default template functions registered: %d functions", len(e.funcMap))
}

// initializeTemplateEngine 初始化模板引擎的默认设置
func (e *TemplateEngine) initializeTemplateEngine() {
	if e.templates == nil {
		e.templates = make(map[string]*template.Template)
	}
	if e.layouts == nil {
		e.layouts = make(map[string]*template.Template)
	}
	if e.components == nil {
		e.components = make(map[string]*template.Template)
	}
	// 注意：themes字段可能已在engine.go中定义，这里简化处理
	// TODO: 检查现有engine结构体中的themes字段类型
	
	// 设置默认值
	if e.extension == "" {
		e.extension = ".html"
	}
	if e.delimLeft == "" {
		e.delimLeft = "{{"
	}
	if e.delimRight == "" {
		e.delimRight = "}}"
	}
	if e.RunMode == "" {
		e.RunMode = "dev"
	}
	if e.currentTheme == "" {
		e.currentTheme = "default"
	}

	// 注册默认函数
	e.setupDefaultTemplateFunctions()

	config.Debugf("Template engine initialized with extension: %s, delims: %s/%s, mode: %s", 
		e.extension, e.delimLeft, e.delimRight, e.RunMode)
}

// validateEngineConfig 验证引擎配置
func (e *TemplateEngine) validateEngineConfig() error {
	if len(e.viewPaths) == 0 {
		config.Warnf("No view paths configured")
	}

	// 检查视图路径是否存在
	for _, path := range e.viewPaths {
		if _, err := os.Stat(path); err != nil {
			config.Warnf("View path does not exist or is not accessible: %s", path)
		}
	}

	// 检查布局路径
	if e.layoutPath != "" {
		if _, err := os.Stat(e.layoutPath); err != nil {
			config.Warnf("Layout path does not exist or is not accessible: %s", e.layoutPath)
		}
	}

	// 检查组件路径
	if e.componentPath != "" {
		if _, err := os.Stat(e.componentPath); err != nil {
			config.Warnf("Component path does not exist or is not accessible: %s", e.componentPath)
		}
	}

	return nil
}

// cleanupExpiredCache 清理过期缓存（基于时间或使用频率）
func (e *TemplateEngine) cleanupExpiredCache(maxAge time.Duration) int {
	if !e.enableCache {
		return 0
	}

	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	_ = time.Now().Unix() // 读取但不使用，避免未使用变量警告
	cleanedCount := 0

	// 简化版本的缓存清理策略
	// TODO: 实现更完整的基于时间的缓存清理策略
	// 当前版本只是示例，实际使用时需要根据具体需求实现
	config.Debugf("Cache cleanup requested, but simplified implementation - no cleanup performed")

	if cleanedCount > 0 {
		config.Infof("Cleaned up %d expired templates from cache", cleanedCount)
	}

	return cleanedCount
}

// warmupCache 预热缓存 - 加载常用模板
func (e *TemplateEngine) warmupCache(commonTemplates []string) error {
	config.Infof("Starting cache warmup for %d templates", len(commonTemplates))
	
	successCount := 0
	errorCount := 0

	for _, templateName := range commonTemplates {
		if _, err := e.GetTemplate(templateName); err != nil {
			config.Warnf("Failed to warmup template %s: %v", templateName, err)
			errorCount++
		} else {
			successCount++
		}
	}

	config.Infof("Cache warmup completed: %d success, %d failed", successCount, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("failed to warmup %d templates", errorCount)
	}
	
	return nil
}

// GetCacheMetrics 获取缓存指标
func (e *TemplateEngine) GetCacheMetrics() map[string]any {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	templateCount := len(e.templates)
	layoutCount := len(e.layouts)
	componentCount := len(e.components)
	
	metrics := map[string]any{
		"template_count":  templateCount,
		"layout_count":    layoutCount,
		"component_count": componentCount,
		"total_count":     templateCount + layoutCount + componentCount,
		"cache_enabled":   e.enableCache,
		"reload_enabled":  e.enableReload,
		"compress_enabled": e.enableCompress,
		"run_mode":        e.RunMode,
		"view_paths":      len(e.viewPaths),
	}

	// 估算内存使用（简化计算）
	estimatedMemory := (templateCount + layoutCount + componentCount) * 2048 // 每个模板约2KB
	metrics["estimated_memory_bytes"] = estimatedMemory

	return metrics
}

// IsTemplateExists 检查模板是否存在
func (e *TemplateEngine) IsTemplateExists(templateName string) bool {
	_, err := e.FindTemplateFile(templateName)
	return err == nil
}

// GetTemplateInfo 获取模板信息
func (e *TemplateEngine) GetTemplateInfo(templateName string) map[string]any {
	info := map[string]any{
		"name":     templateName,
		"exists":   false,
		"cached":   false,
		"path":     "",
		"size":     0,
		"modified": "",
	}

	// 检查文件是否存在
	templatePath, err := e.FindTemplateFile(templateName)
	if err != nil {
		return info
	}

	info["exists"] = true
	info["path"] = templatePath

	// 获取文件信息
	if fileInfo, err := os.Stat(templatePath); err == nil {
		info["size"] = fileInfo.Size()
		info["modified"] = fileInfo.ModTime().Format(time.RFC3339)
	}

	// 检查是否在缓存中
	cacheKey := DefaultCacheKeyManager.GenerateTemplateKey(templateName)
	e.templateMutex.RLock()
	_, cached := e.templates[cacheKey]
	e.templateMutex.RUnlock()
	
	info["cached"] = cached

	return info
}