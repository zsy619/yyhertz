// Package view 提供模板引擎的性能监控功能
//
// 这个文件包含了模板引擎的性能监控相关功能：
// - PerformanceStats 性能统计
// - CacheOptimizer 缓存优化器  
// - CompressionManager 压缩管理器
// - 相关的性能监控方法
package view

import (
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// PerformanceStats 性能统计
type PerformanceStats struct {
	mutex             sync.RWMutex
	RenderCount       int64                    // 渲染次数
	TotalRenderTime   time.Duration            // 总渲染时间
	AverageRenderTime time.Duration            // 平均渲染时间
	CacheHitCount     int64                    // 缓存命中次数
	CacheMissCount    int64                    // 缓存未命中次数
	TemplateErrors    int64                    // 模板错误次数
	LastRenderTime    time.Time                // 最后渲染时间
	RenderTimes       map[string]time.Duration // 各模板渲染时间
	HotTemplates      map[string]int64         // 热点模板统计
}

// CacheOptimizer 缓存优化器
type CacheOptimizer struct {
	mutex            sync.RWMutex
	precompiledCache map[string]*template.Template // 预编译缓存
	frequencyTracker map[string]int64              // 访问频率追踪
	lastAccessTime   map[string]time.Time          // 最后访问时间
	maxCacheSize     int                           // 最大缓存大小
	cleanupInterval  time.Duration                 // 清理间隔
	cleanupTicker    *time.Ticker                  // 清理定时器
}

// CompressionManager 压缩管理器
type CompressionManager struct {
	mutex             sync.RWMutex
	enableGzip        bool              // 是否启用Gzip
	gzipLevel         int               // Gzip压缩级别
	compressionCache  map[string][]byte // 压缩结果缓存
	minSizeToCompress int               // 最小压缩大小
}

// GetCacheStats 获取缓存统计信息（增强版，合并 Beego 兼容）
func (e *TemplateEngine) GetCacheStats() map[string]interface{} {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	stats := map[string]interface{}{
		"templates":      len(e.templates),
		"layouts":        len(e.layouts),
		"components":     len(e.components),
		"total":          len(e.templates) + len(e.layouts) + len(e.components),
		"view_paths":     e.viewPaths,
		"template_ext":   e.TemplateExt,
		"enable_cache":   e.enableCache,
		"auto_render":    e.AutoRender,
		"delim_left":     e.delimLeft,
		"delim_right":    e.delimRight,
		"func_map":       len(e.funcMap),
	}

	// 添加主题相关信息
	if e.currentTheme != "" {
		stats["current_theme"] = e.currentTheme
	}
	if len(e.themes) > 0 {
		themeNames := make([]string, 0, len(e.themes))
		for name := range e.themes {
			themeNames = append(themeNames, name)
		}
		stats["themes"] = themeNames
	}

	return stats
}

// OptimizeCache 优化缓存（增强版）
func (e *TemplateEngine) OptimizeCache() {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	if !e.enableCache {
		return
	}

	originalCount := len(e.templates) + len(e.layouts) + len(e.components)

	// 清理可能无效的模板引用
	validTemplates := make(map[string]*template.Template)
	validLayouts := make(map[string]*template.Template)
	validComponents := make(map[string]*template.Template)

	for name, tmpl := range e.templates {
		if tmpl != nil {
			validTemplates[name] = tmpl
		}
	}

	for name, tmpl := range e.layouts {
		if tmpl != nil {
			validLayouts[name] = tmpl
		}
	}

	for name, tmpl := range e.components {
		if tmpl != nil {
			validComponents[name] = tmpl
		}
	}

	e.templates = validTemplates
	e.layouts = validLayouts
	e.components = validComponents

	newCount := len(e.templates) + len(e.layouts) + len(e.components)
	if originalCount > newCount {
		config.Infof("Cache optimization completed: %d -> %d templates", originalCount, newCount)
	}
}

// GetAdvancedPerformanceStats 获取高级性能统计信息
func (e *TemplateEngine) GetAdvancedPerformanceStats() map[string]any {
	if e.performanceStats == nil {
		return map[string]any{
			"render_count":   0,
			"cache_hit_rate": "0%",
			"error_message":  "Performance stats not enabled - call EnableAdvancedFeatures() first",
		}
	}

	e.performanceStats.mutex.RLock()
	defer e.performanceStats.mutex.RUnlock()

	// 计算缓存命中率
	var hitRate float64
	total := e.performanceStats.CacheHitCount + e.performanceStats.CacheMissCount
	if total > 0 {
		hitRate = float64(e.performanceStats.CacheHitCount) / float64(total) * 100
	}

	// 获取最热门的模板
	var hotTemplate string
	var maxCount int64
	for template, count := range e.performanceStats.HotTemplates {
		if count > maxCount {
			maxCount = count
			hotTemplate = template
		}
	}

	return map[string]any{
		"render_count":        e.performanceStats.RenderCount,
		"total_render_time":   e.performanceStats.TotalRenderTime.String(),
		"average_render_time": e.performanceStats.AverageRenderTime.String(),
		"cache_hit_count":     e.performanceStats.CacheHitCount,
		"cache_miss_count":    e.performanceStats.CacheMissCount,
		"cache_hit_rate":      fmt.Sprintf("%.2f%%", hitRate),
		"template_errors":     e.performanceStats.TemplateErrors,
		"last_render_time":    e.performanceStats.LastRenderTime,
		"hot_template":        hotTemplate,
		"hot_template_count":  maxCount,
		"render_times":        e.performanceStats.RenderTimes,
	}
}

// startCacheCleanup 启动缓存清理定时器
func (e *TemplateEngine) startCacheCleanup() {
	if e.cacheOptimizer == nil {
		return
	}

	ticker := time.NewTicker(e.cacheOptimizer.cleanupInterval)
	e.cacheOptimizer.cleanupTicker = ticker

	go func() {
		for range ticker.C {
			e.cleanupAdvancedCache()
		}
	}()
}

// cleanupAdvancedCache 清理高级缓存
func (e *TemplateEngine) cleanupAdvancedCache() {
	if e.cacheOptimizer == nil {
		return
	}

	e.cacheOptimizer.mutex.Lock()
	defer e.cacheOptimizer.mutex.Unlock()

	cutoffTime := time.Now().Add(-e.cacheOptimizer.cleanupInterval * 2)
	deleted := 0

	for templateName, lastAccess := range e.cacheOptimizer.lastAccessTime {
		if lastAccess.Before(cutoffTime) {
			delete(e.cacheOptimizer.precompiledCache, templateName)
			delete(e.cacheOptimizer.lastAccessTime, templateName)
			delete(e.cacheOptimizer.frequencyTracker, templateName)
			deleted++
		}
	}

	if deleted > 0 {
		config.Debugf("Cleaned up %d cached templates from advanced cache", deleted)
	}
}

// recordRenderTime 记录渲染时间（用于性能统计）
func (e *TemplateEngine) recordRenderTime(templateName string, duration time.Duration) {
	if e.performanceStats == nil {
		return
	}

	e.performanceStats.mutex.Lock()
	defer e.performanceStats.mutex.Unlock()

	e.performanceStats.TotalRenderTime += duration
	e.performanceStats.RenderTimes[templateName] = duration
	if e.performanceStats.RenderCount > 0 {
		e.performanceStats.AverageRenderTime = time.Duration(int64(e.performanceStats.TotalRenderTime) / e.performanceStats.RenderCount)
	}
	e.performanceStats.LastRenderTime = time.Now()
}

// enablePerformanceFeatures 启用性能监控功能
func (e *TemplateEngine) enablePerformanceFeatures() {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 启用性能统计
	if e.performanceStats == nil {
		e.performanceStats = &PerformanceStats{
			RenderTimes:  make(map[string]time.Duration),
			HotTemplates: make(map[string]int64),
		}
	}

	// 启用缓存优化器
	if e.cacheOptimizer == nil {
		e.cacheOptimizer = &CacheOptimizer{
			precompiledCache: make(map[string]*template.Template),
			frequencyTracker: make(map[string]int64),
			lastAccessTime:   make(map[string]time.Time),
			maxCacheSize:     1000,
			cleanupInterval:  time.Minute * 10,
		}
		go e.startCacheCleanup()
	}

	// 启用压缩管理器
	if e.compressionMgr == nil {
		e.compressionMgr = &CompressionManager{
			enableGzip:        e.EnableGzip,
			gzipLevel:         6, // 默认压缩级别
			compressionCache:  make(map[string][]byte),
			minSizeToCompress: 1024, // 1KB
		}
	}

	config.Infof("Advanced features enabled for TemplateEngine")
}