// Package view 提供增强的模板引擎功能
//
// 这个文件包含了模板管理器相关的代码，包括：
// - TemplateManager（基础模板管理器）
// - EnhancedTemplateManager（增强模板管理器）
// - ConventionManager（命名约定管理器）
package view

import (
	"fmt"
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
		etm.engine.RemoveTemplateFromCache(filePath)
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

// ============= 命名约定管理器 =============

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
				Transform:   func(s string) string { 
					if len(s) >= 10 && strings.HasSuffix(s, "Controller") {
						return strings.ToLower(s[:len(s)-10])
					}
					return strings.ToLower(s)
				},
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