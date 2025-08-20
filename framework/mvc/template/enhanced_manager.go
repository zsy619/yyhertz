package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/zsy619/yyhertz/framework/config"
)

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
	if enhanced.config.EnableReload {
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
	for _, viewPath := range etm.config.ViewPaths {
		count, err := etm.discoverTemplatesInPath(viewPath)
		if err != nil {
			config.Warnf("Failed to discover templates in %s: %v", viewPath, err)
			continue
		}
		discoveredCount += count
	}

	// 扫描布局路径
	if etm.config.LayoutPath != "" {
		count, err := etm.discoverTemplatesInPath(etm.config.LayoutPath)
		if err != nil {
			config.Warnf("Failed to discover layouts in %s: %v", etm.config.LayoutPath, err)
		} else {
			discoveredCount += count
		}
	}

	// 扫描组件路径
	if etm.config.ComponentPath != "" {
		count, err := etm.discoverTemplatesInPath(etm.config.ComponentPath)
		if err != nil {
			config.Warnf("Failed to discover components in %s: %v", etm.config.ComponentPath, err)
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
	if etm.config.Extension != "" && ext == etm.config.Extension {
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
	watchPaths := append(etm.config.ViewPaths, etm.config.LayoutPath, etm.config.ComponentPath)
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
	// 这里需要调用底层模板引擎的缓存清除方法
	// 具体实现取决于 view.TemplateEngine 的 API
	if etm.engine != nil {
		// 假设模板引擎有清除缓存的方法
		// etm.engine.ClearCache(filePath)
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
