package view

// 注意：CacheStats 和相关统计功能已经在 render_cache.go 中定义
// 这个文件将包含额外的统计和分析功能，避免重复定义

import (
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	RenderCount       int64         `json:"render_count"`
	TotalRenderTime   time.Duration `json:"total_render_time"`
	AverageRenderTime time.Duration `json:"average_render_time"`
	CacheHitRate      float64       `json:"cache_hit_rate"`
	LastResetTime     time.Time     `json:"last_reset_time"`
}

// ResetPerformanceMetrics 重置性能指标
func (pm *PerformanceMetrics) Reset() {
	pm.RenderCount = 0
	pm.TotalRenderTime = 0
	pm.AverageRenderTime = 0
	pm.CacheHitRate = 0
	pm.LastResetTime = time.Now()
}

// AddRenderTime 添加渲染时间
func (pm *PerformanceMetrics) AddRenderTime(duration time.Duration) {
	pm.RenderCount++
	pm.TotalRenderTime += duration
	pm.AverageRenderTime = pm.TotalRenderTime / time.Duration(pm.RenderCount)
}

// GetPerformanceReport 获取性能报告
func (e *TemplateEngine) GetPerformanceReport() map[string]any {
	stats := e.GetEnhancedCacheStats()
	
	// 基础指标
	report := map[string]any{
		"cache_stats": stats,
		"engine_info": map[string]any{
			"cache_enabled":    e.enableCache,
			"reload_enabled":   e.enableReload,
			"compress_enabled": e.enableCompress,
			"run_mode":         e.RunMode,
			"current_theme":    e.currentTheme,
			"view_paths_count": len(e.viewPaths),
		},
		"template_distribution": e.getTemplateDistribution(),
		"recommendations":       e.generateRecommendations(),
	}
	
	return report
}

// getTemplateDistribution 获取模板分布情况
func (e *TemplateEngine) getTemplateDistribution() map[string]any {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()
	
	distribution := map[string]any{
		"total_templates":  len(e.templates),
		"total_layouts":    len(e.layouts),
		"total_components": len(e.components),
	}
	
	// 分析模板类型分布
	layoutTemplateCount := 0
	singleTemplateCount := 0
	
	for key := range e.templates {
		if DefaultCacheKeyManager.IsLayoutKey(key) {
			layoutTemplateCount++
		} else {
			singleTemplateCount++
		}
	}
	
	distribution["layout_templates"] = layoutTemplateCount
	distribution["single_templates"] = singleTemplateCount
	
	return distribution
}

// generateRecommendations 生成优化建议
func (e *TemplateEngine) generateRecommendations() []string {
	var recommendations []string
	
	e.templateMutex.RLock()
	totalTemplates := len(e.templates) + len(e.layouts) + len(e.components)
	e.templateMutex.RUnlock()
	
	// 模板数量建议
	if totalTemplates > 1000 {
		recommendations = append(recommendations, "Consider implementing template cache cleanup strategy for large template count")
	}
	
	// 缓存建议
	if !e.enableCache {
		recommendations = append(recommendations, "Enable template caching to improve performance")
	}
	
	// 开发模式建议
	if e.RunMode == "dev" && e.enableReload {
		recommendations = append(recommendations, "Disable hot reload in production for better performance")
	}
	
	// 压缩建议
	if !e.enableCompress {
		recommendations = append(recommendations, "Enable HTML compression to reduce response size")
	}
	
	// 预加载建议
	if totalTemplates < 50 {
		recommendations = append(recommendations, "Consider using template preloading for frequently used templates")
	}
	
	return recommendations
}

// GetHealthCheck 获取模板引擎健康检查信息
func (e *TemplateEngine) GetHealthCheck() map[string]any {
	health := map[string]any{
		"status": "healthy",
		"checks": map[string]any{},
	}
	
	checks := health["checks"].(map[string]any)
	isHealthy := true
	
	// 检查配置
	if len(e.viewPaths) == 0 {
		checks["view_paths"] = map[string]any{
			"status": "warning",
			"message": "No view paths configured",
		}
	} else {
		checks["view_paths"] = map[string]any{
			"status": "ok",
			"count": len(e.viewPaths),
		}
	}
	
	// 检查函数映射
	if len(e.funcMap) < 5 {
		checks["functions"] = map[string]any{
			"status": "warning",
			"message": "Very few template functions registered",
			"count": len(e.funcMap),
		}
	} else {
		checks["functions"] = map[string]any{
			"status": "ok",
			"count": len(e.funcMap),
		}
	}
	
	// 检查模板数量
	e.templateMutex.RLock()
	templateCount := len(e.templates)
	e.templateMutex.RUnlock()
	
	if templateCount == 0 {
		checks["templates"] = map[string]any{
			"status": "warning",
			"message": "No templates cached",
		}
	} else {
		checks["templates"] = map[string]any{
			"status": "ok",
			"cached_count": templateCount,
		}
	}
	
	// 检查关键函数
	criticalFunctions := []string{"csrf_token", "now", "safeHTML"}
	missingFunctions := []string{}
	for _, funcName := range criticalFunctions {
		if _, exists := e.funcMap[funcName]; !exists {
			missingFunctions = append(missingFunctions, funcName)
		}
	}
	
	if len(missingFunctions) > 0 {
		checks["critical_functions"] = map[string]any{
			"status": "error",
			"message": "Critical functions missing",
			"missing": missingFunctions,
		}
		isHealthy = false
	} else {
		checks["critical_functions"] = map[string]any{
			"status": "ok",
			"message": "All critical functions available",
		}
	}
	
	// 设置整体状态
	if !isHealthy {
		health["status"] = "unhealthy"
	}
	
	// 添加时间戳
	health["timestamp"] = time.Now().Format(time.RFC3339)
	
	return health
}

// GetDiagnostics 获取详细诊断信息
func (e *TemplateEngine) GetDiagnostics() map[string]any {
	diagnostics := map[string]any{
		"engine_config": map[string]any{
			"extension":        e.extension,
			"delims":          []string{e.delimLeft, e.delimRight},
			"cache_enabled":   e.enableCache,
			"reload_enabled":  e.enableReload,
			"compress_enabled": e.enableCompress,
			"run_mode":        e.RunMode,
			"current_theme":   e.currentTheme,
		},
		"paths": map[string]any{
			"view_paths":     e.viewPaths,
			"layout_path":    e.layoutPath,
			"component_path": e.componentPath,
		},
		"cache_info": e.GetEnhancedCacheStats(),
		"function_info": map[string]any{
			"total_functions": len(e.funcMap),
			"function_names":  e.getFunctionNames(),
		},
		"themes": map[string]any{
			"current":   e.currentTheme,
			"available": e.GetAvailableThemes(),
		},
		"system_info": map[string]any{
			"timestamp": time.Now().Format(time.RFC3339),
			"go_version": "1.21+", // 可以通过runtime.Version()获取
		},
	}
	
	return diagnostics
}

// getFunctionNames 获取已注册的函数名称列表
func (e *TemplateEngine) getFunctionNames() []string {
	names := make([]string, 0, len(e.funcMap))
	for name := range e.funcMap {
		names = append(names, name)
	}
	return names
}

// LogEngineStatus 记录引擎状态
func (e *TemplateEngine) LogEngineStatus() {
	e.templateMutex.RLock()
	templateCount := len(e.templates)
	layoutCount := len(e.layouts)
	componentCount := len(e.components)
	e.templateMutex.RUnlock()
	
	config.Infof("Template Engine Status:")
	config.Infof("  - Templates cached: %d", templateCount)
	config.Infof("  - Layouts loaded: %d", layoutCount)
	config.Infof("  - Components loaded: %d", componentCount)
	config.Infof("  - Functions registered: %d", len(e.funcMap))
	config.Infof("  - Cache enabled: %v", e.enableCache)
	config.Infof("  - Reload enabled: %v", e.enableReload)
	config.Infof("  - Run mode: %s", e.RunMode)
	config.Infof("  - Current theme: %s", e.currentTheme)
}

// GetDetailedMemoryUsage 获取详细内存使用估算（避免与engine.go冲突）
func (e *TemplateEngine) GetDetailedMemoryUsage() map[string]any {
	e.templateMutex.RLock()
	templateCount := len(e.templates)
	layoutCount := len(e.layouts)
	componentCount := len(e.components)
	e.templateMutex.RUnlock()
	
	// 简单估算（每个模板约2KB）
	templateMemory := templateCount * 2048
	layoutMemory := layoutCount * 1024
	componentMemory := componentCount * 512
	totalMemory := templateMemory + layoutMemory + componentMemory
	
	return map[string]any{
		"templates_bytes":  templateMemory,
		"layouts_bytes":    layoutMemory,
		"components_bytes": componentMemory,
		"total_bytes":      totalMemory,
		"total_mb":         float64(totalMemory) / (1024 * 1024),
	}
}