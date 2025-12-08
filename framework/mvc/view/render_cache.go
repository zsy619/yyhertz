package view

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 缓存键管理器 =============

// CacheKeyManager 缓存键管理器
type CacheKeyManager struct {
	delimiter string // 分隔符
}

// DefaultCacheKeyManager 默认缓存键管理器
var DefaultCacheKeyManager = &CacheKeyManager{
	delimiter: "@",
}

// GenerateLayoutKey 生成布局模板缓存键
func (ckm *CacheKeyManager) GenerateLayoutKey(templateName, layoutName string) string {
	if layoutName == "" {
		return templateName
	}
	return fmt.Sprintf("%s%s%s", templateName, ckm.delimiter, layoutName)
}

// GenerateTemplateKey 生成模板缓存键
func (ckm *CacheKeyManager) GenerateTemplateKey(templateName string) string {
	// 🔧 关键修复：缓存键应该使用文件basename（与LoadTemplate中的模板名称保持一致）
	// 因为LoadTemplate中使用filepath.Base(templatePath)作为模板名，缓存键也应该对应

	// 先获取文件的basename
	baseName := strings.TrimPrefix(templateName, "/")
	baseName = strings.TrimSuffix(baseName, "/")
	// 再去掉扩展名
	if idx := strings.LastIndex(baseName, "."); idx > 0 {
		baseName = baseName[:idx]
	}

	config.Debugf("🔧 CacheKey generation: original=%s, final_key=%s", templateName, baseName)
	return baseName
}

// ParseLayoutKey 解析布局缓存键
func (ckm *CacheKeyManager) ParseLayoutKey(key string) (templateName, layoutName string) {
	parts := strings.Split(key, ckm.delimiter)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return key, ""
}

// IsLayoutKey 检查是否是布局缓存键
func (ckm *CacheKeyManager) IsLayoutKey(key string) bool {
	return strings.Contains(key, ckm.delimiter)
}

// ValidateKey 验证缓存键的有效性
func (ckm *CacheKeyManager) ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("cache key cannot be empty")
	}
	if strings.Contains(key, " ") {
		return fmt.Errorf("cache key cannot contain spaces: %s", key)
	}
	return nil
}

// ============= 增强的缓存统计和监控 =============

// CacheStats 详细缓存统计信息
type CacheStats struct {
	TemplateCount  int            `json:"template_count"`
	LayoutCount    int            `json:"layout_count"`
	ComponentCount int            `json:"component_count"`
	TotalCount     int            `json:"total_count"`
	MemoryEstimate int            `json:"memory_estimate_bytes"`
	CacheEnabled   bool           `json:"cache_enabled"`
	CacheHitRate   float64        `json:"cache_hit_rate"`
	CacheKeys      []string       `json:"cache_keys,omitempty"`
	LayoutKeys     []string       `json:"layout_keys,omitempty"`
	RecentAccess   []string       `json:"recent_access,omitempty"`
	KeyTypes       map[string]int `json:"key_types"`
}

// GetEnhancedCacheStats 获取增强的缓存统计信息
func (e *TemplateEngine) GetEnhancedCacheStats() *CacheStats {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	stats := &CacheStats{
		TemplateCount:  len(e.templates),
		LayoutCount:    len(e.layouts),
		ComponentCount: len(e.components),
		TotalCount:     len(e.templates) + len(e.layouts) + len(e.components),
		CacheEnabled:   e.enableCache,
		KeyTypes:       make(map[string]int),
		CacheKeys:      make([]string, 0, len(e.templates)),
		LayoutKeys:     make([]string, 0),
	}

	// 分析缓存键类型
	layoutKeyCount := 0
	for key := range e.templates {
		stats.CacheKeys = append(stats.CacheKeys, key)
		if DefaultCacheKeyManager.IsLayoutKey(key) {
			layoutKeyCount++
			stats.LayoutKeys = append(stats.LayoutKeys, key)
		}
	}

	stats.KeyTypes["template_keys"] = len(e.templates) - layoutKeyCount
	stats.KeyTypes["layout_keys"] = layoutKeyCount

	// 估算内存使用
	var totalMemory int
	for key, tmpl := range e.templates {
		if tmpl != nil {
			// 基础估算: 缓存键长度 + 模板对象开销
			totalMemory += len(key)*2 + 2048 // 每个模板约2KB开销
		}
	}
	for key := range e.layouts {
		totalMemory += len(key)*2 + 1024 // 每个布局约1KB开销
	}
	for key := range e.components {
		totalMemory += len(key)*2 + 512 // 每个组件约512B开销
	}

	stats.MemoryEstimate = totalMemory

	// 计算缓存命中率（简化版，实际应该维护独立的命中统计）
	if stats.TotalCount > 0 {
		// 假设有缓存的模板命中率较高
		stats.CacheHitRate = 0.85 // 示例值，实际实现中应该动态计算
	}

	return stats
}

// AnalyzeCachePerformance 分析缓存性能
func (e *TemplateEngine) AnalyzeCachePerformance() map[string]any {
	stats := e.GetEnhancedCacheStats()

	analysis := map[string]any{
		"basic_stats": stats,
		"performance_metrics": map[string]any{
			"cache_efficiency": "good", // 根据实际统计确定
			"memory_usage":     "acceptable",
			"recommendations":  []string{},
		},
	}

	// 性能分析建议
	recommendations := []string{}

	if stats.TotalCount > 1000 {
		recommendations = append(recommendations, "Consider implementing cache size limits")
	}

	if stats.MemoryEstimate > 50*1024*1024 { // > 50MB
		recommendations = append(recommendations, "High memory usage detected, consider cache cleanup")
	}

	if float64(stats.KeyTypes["layout_keys"])/float64(stats.TotalCount) > 0.7 {
		recommendations = append(recommendations, "High ratio of layout templates, consider layout optimization")
	}

	analysis["performance_metrics"].(map[string]any)["recommendations"] = recommendations

	return analysis
}

// ClearCache 清除模板缓存
func (e *TemplateEngine) ClearCache() {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	oldCount := len(e.templates) + len(e.layouts) + len(e.components)

	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)

	config.Infof("Template cache cleared: removed %d cached items", oldCount)
}

// ClearTemplateCache 清除指定类型的模板缓存
func (e *TemplateEngine) ClearTemplateCache(cacheType string) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	var clearedCount int

	switch cacheType {
	case "templates":
		clearedCount = len(e.templates)
		e.templates = make(map[string]*template.Template)
	case "layouts":
		clearedCount = len(e.layouts)
		e.layouts = make(map[string]*template.Template)
	case "components":
		clearedCount = len(e.components)
		e.components = make(map[string]*template.Template)
	default:
		config.Warnf("Unknown cache type: %s", cacheType)
		return
	}

	config.Infof("%s cache cleared: removed %d items", cacheType, clearedCount)
}

// ClearTemplateByPattern 根据模式清除模板缓存
func (e *TemplateEngine) ClearTemplateByPattern(pattern string) int {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	clearedCount := 0

	// 清理匹配的模板缓存
	for key := range e.templates {
		if strings.Contains(key, pattern) {
			delete(e.templates, key)
			clearedCount++
		}
	}

	if clearedCount > 0 {
		config.Infof("Cleared %d template cache entries matching pattern: %s", clearedCount, pattern)
	}

	return clearedCount
}

// OptimizeCacheWithLimit 优化缓存，移除不再需要的模板（带限制）
func (e *TemplateEngine) OptimizeCacheWithLimit(maxEntries int) int {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	currentCount := len(e.templates)
	if currentCount <= maxEntries {
		return 0 // 不需要优化
	}

	// 简单的LRU策略：保留最近的maxEntries个条目
	// 这里使用一个简化的实现，实际中可以使用更复杂的LRU算法

	// 将所有key收集到slice中（Go的map遍历是随机的，这里模拟LRU）
	keys := make([]string, 0, len(e.templates))
	for key := range e.templates {
		keys = append(keys, key)
	}

	// 保留前maxEntries个，删除其余的
	removedCount := 0
	if len(keys) > maxEntries {
		toRemove := keys[maxEntries:]
		for _, key := range toRemove {
			delete(e.templates, key)
			removedCount++
		}
	}

	if removedCount > 0 {
		config.Infof("Cache optimized: removed %d entries, %d remaining", removedCount, len(e.templates))
	}

	return removedCount
}

// GetTemplateList 获取模板列表
func (e *TemplateEngine) GetTemplateList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	templates := make([]string, 0, len(e.templates))
	for name := range e.templates {
		templates = append(templates, name)
	}
	return templates
}

// GetLayoutList 获取布局列表
func (e *TemplateEngine) GetLayoutList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	layouts := make([]string, 0, len(e.layouts))
	for name := range e.layouts {
		layouts = append(layouts, name)
	}
	return layouts
}

// GetComponentList 获取组件列表
func (e *TemplateEngine) GetComponentList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	components := make([]string, 0, len(e.components))
	for name := range e.components {
		components = append(components, name)
	}
	return components
}
