package errors

import (
	"time"
)

// =============================================================================
// 模块：统计管理
// 职责：管理错误统计信息，提供查询、重置、重试判断等功能
// =============================================================================

// GetErrorStatistics 获取错误统计信息（返回副本）
func GetErrorStatistics(stats *ErrorStatistics) *ErrorStatistics {
	if stats == nil {
		return nil
	}

	stats.mu.RLock()
	defer stats.mu.RUnlock()

	// 返回副本以避免并发问题
	statsCopy := &ErrorStatistics{
		TotalErrors:    stats.TotalErrors,
		ErrorsByStatus: make(map[int]int64),
		ErrorsByPath:   make(map[string]int64),
		LastErrors:     make([]ErrorRecord, len(stats.LastErrors)),
		StartTime:      stats.StartTime,
	}

	for k, v := range stats.ErrorsByStatus {
		statsCopy.ErrorsByStatus[k] = v
	}

	for k, v := range stats.ErrorsByPath {
		statsCopy.ErrorsByPath[k] = v
	}

	copy(statsCopy.LastErrors, stats.LastErrors)

	return statsCopy
}

// ResetStatistics 重置错误统计信息
func ResetStatistics(stats *ErrorStatistics) {
	if stats == nil {
		return
	}

	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.TotalErrors = 0
	stats.ErrorsByStatus = make(map[int]int64)
	stats.ErrorsByPath = make(map[string]int64)
	stats.LastErrors = make([]ErrorRecord, 0)
	stats.StartTime = time.Now()
}

// IsRetryable 检查错误是否可重试
func IsRetryable(statusCode int, retryableErrors map[int]bool) bool {
	return retryableErrors[statusCode]
}

// SetRetryable 设置错误是否可重试
func SetRetryable(statusCode int, retryable bool, retryableErrors map[int]bool) {
	retryableErrors[statusCode] = retryable
}

// IsRetryableError 判断错误是否可重试（增强版）
func IsRetryableError(statusCode int, retryableErrors map[int]bool) bool {
	// 基础可重试检查
	if IsRetryable(statusCode, retryableErrors) {
		return true
	}

	// 额外的智能判断
	switch statusCode {
	case 408, 429: // 超时和频率限制总是可重试
		return true
	case 500, 502, 503, 504: // 服务器临时错误
		return true
	case 507: // 存储空间不足
		return true
	default:
		return false
	}
}