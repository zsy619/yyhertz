// Package view 提供增强的模板引擎缓存管理功能
//
// 这个文件包含模板缓存相关的结构体和方法：
// - TemplateCache 高性能缓存结构
// - CacheEntry 缓存条目定义
// - RenderRequest 和 RenderResult 批量渲染结构
// - 缓存管理相关方法
package view

import (
	"html/template"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// TemplateCache 高性能模板缓存
type TemplateCache struct {
	cache           map[string]*CacheEntry
	maxSize         int
	cleanupTimer    *time.Timer
	cleanupInterval time.Duration
	mu              sync.RWMutex
}

// CacheEntry 缓存条目
type CacheEntry struct {
	template    *template.Template
	lastAccess  time.Time
	accessCount uint64
	createTime  time.Time
	size        int
}

// RenderRequest 渲染请求
type RenderRequest struct {
	TemplateName string
	Data         any
	LayoutName   string // 可选的布局
}

// RenderResult 渲染结果
type RenderResult struct {
	Index        int
	TemplateName string
	HTML         string
	Error        error
	Success      bool
	Duration     time.Duration
}

// BulkRender 批量渲染模板（性能优化版）
func (e *TemplateEngine) BulkRender(requests []RenderRequest) (map[string]RenderResult, error) {
	if len(requests) == 0 {
		return make(map[string]RenderResult), nil
	}

	results := make(map[string]RenderResult, len(requests))

	// 并发渲染以提高性能
	const maxConcurrency = 10 // 限制并发数以避免资源耗尽
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var resultsMu sync.Mutex

	for i, request := range requests {
		wg.Add(1)
		go func(index int, req RenderRequest) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行渲染
			var result RenderResult
			result.Index = index
			result.TemplateName = req.TemplateName

			html, err := e.Render(req.TemplateName, req.Data)
			if err != nil {
				result.Error = err
				result.Success = false
			} else {
				result.HTML = html
				result.Success = true
			}

			// 安全地写入结果
			resultsMu.Lock()
			results[req.TemplateName] = result
			resultsMu.Unlock()
		}(i, request)
	}

	wg.Wait()

	config.Debugf("Bulk render completed: %d templates processed", len(results))
	return results, nil
}

// RemoveTemplateFromCache 从缓存中移除模板
func (e *TemplateEngine) RemoveTemplateFromCache(filePath string) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	templateName := e.getTemplateName(filePath)

	// 从各种缓存中移除
	delete(e.templates, templateName)
	delete(e.layouts, templateName)
	delete(e.components, templateName)

	// 如果是布局或组件文件变化，需要清空所有模板缓存
	// 因为其他模板可能依赖于这些变化的布局或组件
	if strings.Contains(filePath, e.layoutPath) {
		config.Infof("Layout file removed/renamed: %s, clearing all template cache", templateName)
		e.templates = make(map[string]*template.Template)
	} else if strings.Contains(filePath, e.componentPath) {
		config.Infof("Component file removed/renamed: %s, clearing all template cache", templateName)
		e.templates = make(map[string]*template.Template)
	}

	config.Debugf("Template removed from cache: %s", templateName)
}