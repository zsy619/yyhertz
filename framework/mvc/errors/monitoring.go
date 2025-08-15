package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 错误监控功能 =============

// ErrorMonitor 错误监控器
type ErrorMonitor struct {
	controller *DefaultErrorController
}

// NewErrorMonitor 创建错误监控器
func NewErrorMonitor(controller *DefaultErrorController) *ErrorMonitor {
	return &ErrorMonitor{
		controller: controller,
	}
}

// GetStatisticsHandler 获取错误统计信息的HTTP处理器
func (m *ErrorMonitor) GetStatisticsHandler(ctx *mvccontext.Context) {
	if m.controller == nil || m.controller.ErrorStatistics == nil {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"error": "错误统计功能不可用",
		})
		return
	}

	stats := m.controller.GetErrorStatistics()
	if stats == nil {
		ctx.JSON(http.StatusNoContent, map[string]any{
			"message": "暂无错误统计数据",
		})
		return
	}

	// 计算运行时长
	uptime := time.Since(stats.StartTime)
	
	// 计算错误率
	errorRate := float64(0)
	if stats.TotalErrors > 0 {
		errorRate = float64(stats.TotalErrors) / uptime.Hours() // 每小时错误数
	}

	response := map[string]any{
		"total_errors":     stats.TotalErrors,
		"errors_by_status": stats.ErrorsByStatus,
		"errors_by_path":   stats.ErrorsByPath,
		"error_rate_per_hour": errorRate,
		"uptime_hours":     uptime.Hours(),
		"start_time":       stats.StartTime,
		"last_errors":      stats.LastErrors,
		"summary": map[string]any{
			"most_common_errors": m.getMostCommonErrors(stats),
			"error_trends":       m.getErrorTrends(stats),
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// ResetStatisticsHandler 重置错误统计信息的HTTP处理器
func (m *ErrorMonitor) ResetStatisticsHandler(ctx *mvccontext.Context) {
	if m.controller == nil {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"error": "错误控制器不可用",
		})
		return
	}

	m.controller.ResetStatistics()
	
	ctx.JSON(http.StatusOK, map[string]any{
		"message": "错误统计信息已重置",
		"reset_time": time.Now(),
	})
}

// GetHealthHandler 获取错误处理健康状态的HTTP处理器
func (m *ErrorMonitor) GetHealthHandler(ctx *mvccontext.Context) {
	if m.controller == nil {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"status": "unhealthy",
			"reason": "错误控制器不可用",
		})
		return
	}

	stats := m.controller.GetErrorStatistics()
	health := "healthy"
	reasons := make([]string, 0)

	if stats != nil {
		// 检查最近1小时内的错误率
		recentErrors := m.getRecentErrors(stats, time.Hour)
		if len(recentErrors) > 100 { // 如果1小时内错误超过100个
			health = "degraded"
			reasons = append(reasons, "高错误率")
		}

		// 检查是否有大量5xx错误
		serverErrors := int64(0)
		for status, count := range stats.ErrorsByStatus {
			if status >= 500 && status < 600 {
				serverErrors += count
			}
		}
		
		if serverErrors > stats.TotalErrors/2 { // 如果服务器错误超过总错误的一半
			health = "unhealthy"
			reasons = append(reasons, "大量服务器错误")
		}
	}

	response := map[string]any{
		"status": health,
		"timestamp": time.Now(),
	}

	if len(reasons) > 0 {
		response["reasons"] = reasons
	}

	statusCode := http.StatusOK
	if health == "degraded" {
		statusCode = http.StatusPartialContent
	} else if health == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	ctx.JSON(statusCode, response)
}

// GetErrorDetailsHandler 获取特定错误详情的HTTP处理器
func (m *ErrorMonitor) GetErrorDetailsHandler(ctx *mvccontext.Context) {
	statusCodeParam := ctx.Query("status_code")
	pathParam := ctx.Query("path")

	if m.controller == nil || m.controller.ErrorStatistics == nil {
		ctx.JSON(http.StatusServiceUnavailable, map[string]any{
			"error": "错误统计功能不可用",
		})
		return
	}

	stats := m.controller.GetErrorStatistics()
	if stats == nil {
		ctx.JSON(http.StatusNoContent, map[string]any{
			"message": "暂无错误统计数据",
		})
		return
	}

	response := make(map[string]any)

	// 按状态码筛选
	if statusCodeParam != "" {
		// 解析状态码并查找相关错误
		response["filtered_by_status"] = statusCodeParam
		response["errors"] = m.filterErrorsByStatus(stats.LastErrors, statusCodeParam)
	}

	// 按路径筛选
	if pathParam != "" {
		response["filtered_by_path"] = pathParam
		response["errors"] = m.filterErrorsByPath(stats.LastErrors, pathParam)
	}

	// 如果没有筛选条件，返回最近的错误
	if statusCodeParam == "" && pathParam == "" {
		response["recent_errors"] = stats.LastErrors
	}

	ctx.JSON(http.StatusOK, response)
}

// ============= 辅助方法 =============

// getMostCommonErrors 获取最常见的错误
func (m *ErrorMonitor) getMostCommonErrors(stats *ErrorStatistics) []map[string]any {
	type errorCount struct {
		Status int   `json:"status"`
		Count  int64 `json:"count"`
	}

	errors := make([]errorCount, 0)
	for status, count := range stats.ErrorsByStatus {
		errors = append(errors, errorCount{Status: status, Count: count})
	}

	// 简单排序（冒泡排序）
	for i := 0; i < len(errors)-1; i++ {
		for j := 0; j < len(errors)-1-i; j++ {
			if errors[j].Count < errors[j+1].Count {
				errors[j], errors[j+1] = errors[j+1], errors[j]
			}
		}
	}

	// 返回前5个
	result := make([]map[string]any, 0)
	limit := 5
	if len(errors) < limit {
		limit = len(errors)
	}

	for i := 0; i < limit; i++ {
		result = append(result, map[string]any{
			"status_code": errors[i].Status,
			"count":       errors[i].Count,
			"percentage":  float64(errors[i].Count) / float64(stats.TotalErrors) * 100,
		})
	}

	return result
}

// getErrorTrends 获取错误趋势
func (m *ErrorMonitor) getErrorTrends(stats *ErrorStatistics) map[string]any {
	now := time.Now()
	
	// 统计最近24小时内每小时的错误数
	hourlyErrors := make(map[int]int)
	for _, record := range stats.LastErrors {
		hoursAgo := int(now.Sub(record.Timestamp).Hours())
		if hoursAgo < 24 {
			hourlyErrors[hoursAgo]++
		}
	}

	// 计算趋势
	trend := "stable"
	recentHours := 0
	olderHours := 0

	for hour, count := range hourlyErrors {
		if hour < 6 { // 最近6小时
			recentHours += count
		} else if hour < 12 { // 6-12小时前
			olderHours += count
		}
	}

	if recentHours > olderHours*2 {
		trend = "increasing"
	} else if recentHours*2 < olderHours {
		trend = "decreasing"
	}

	return map[string]any{
		"trend":               trend,
		"recent_6h_errors":    recentHours,
		"previous_6h_errors":  olderHours,
		"hourly_distribution": hourlyErrors,
	}
}

// getRecentErrors 获取最近指定时间内的错误
func (m *ErrorMonitor) getRecentErrors(stats *ErrorStatistics, duration time.Duration) []ErrorRecord {
	cutoff := time.Now().Add(-duration)
	recent := make([]ErrorRecord, 0)

	for _, record := range stats.LastErrors {
		if record.Timestamp.After(cutoff) {
			recent = append(recent, record)
		}
	}

	return recent
}

// filterErrorsByStatus 按状态码筛选错误
func (m *ErrorMonitor) filterErrorsByStatus(errors []ErrorRecord, statusCode string) []ErrorRecord {
	filtered := make([]ErrorRecord, 0)
	
	for _, record := range errors {
		if fmt.Sprintf("%d", record.StatusCode) == statusCode {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// filterErrorsByPath 按路径筛选错误
func (m *ErrorMonitor) filterErrorsByPath(errors []ErrorRecord, path string) []ErrorRecord {
	filtered := make([]ErrorRecord, 0)
	
	for _, record := range errors {
		if record.Path == path {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// ============= 路由注册辅助函数 =============

// RegisterMonitoringRoutes 注册错误监控路由
func RegisterMonitoringRoutes(app interface{}, controller *DefaultErrorController) {
	_ = NewErrorMonitor(controller)
	
	// 这里需要根据具体的路由注册方式来实现
	// 由于我们不确定app的具体类型，这里提供一个示例结构
	
	// 示例路由注册（需要根据实际的App接口调整）
	// app.GET("/admin/errors/statistics", monitor.GetStatisticsHandler)
	// app.POST("/admin/errors/reset", monitor.ResetStatisticsHandler)
	// app.GET("/admin/errors/health", monitor.GetHealthHandler)
	// app.GET("/admin/errors/details", monitor.GetErrorDetailsHandler)
}

// ExportStatistics 导出错误统计数据为JSON
func (m *ErrorMonitor) ExportStatistics() ([]byte, error) {
	if m.controller == nil || m.controller.ErrorStatistics == nil {
		return nil, fmt.Errorf("错误统计功能不可用")
	}

	stats := m.controller.GetErrorStatistics()
	if stats == nil {
		return nil, fmt.Errorf("暂无错误统计数据")
	}

	return json.MarshalIndent(stats, "", "  ")
}

// ImportStatistics 从JSON导入错误统计数据
func (m *ErrorMonitor) ImportStatistics(data []byte) error {
	if m.controller == nil || m.controller.ErrorStatistics == nil {
		return fmt.Errorf("错误统计功能不可用")
	}

	var stats ErrorStatistics
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("解析统计数据失败: %w", err)
	}

	// 更新统计数据
	m.controller.ErrorStatistics.mu.Lock()
	defer m.controller.ErrorStatistics.mu.Unlock()

	m.controller.ErrorStatistics.TotalErrors = stats.TotalErrors
	m.controller.ErrorStatistics.ErrorsByStatus = stats.ErrorsByStatus
	m.controller.ErrorStatistics.ErrorsByPath = stats.ErrorsByPath
	m.controller.ErrorStatistics.LastErrors = stats.LastErrors
	m.controller.ErrorStatistics.StartTime = stats.StartTime

	return nil
}