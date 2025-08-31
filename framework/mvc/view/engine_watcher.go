package view

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 文件监控和热重载功能 =============

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