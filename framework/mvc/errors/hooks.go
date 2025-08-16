package errors

import (
	"fmt"
	"log"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// =============================================================================
// 模块：钩子系统
// 职责：提供可扩展的错误处理钩子机制，支持日志、统计、清理等
// =============================================================================

// logErrorHook 错误日志记录钩子
func LogErrorHook(ctx *mvccontext.Context, statusCode int, err error, enableErrorLogging bool, errorLogger *log.Logger) error {
	if !enableErrorLogging || errorLogger == nil {
		return nil
	}

	requestID := ""
	if val, exists := ctx.Get("request_id"); exists {
		if id, ok := val.(string); ok {
			requestID = id
		}
	}

	logMsg := fmt.Sprintf("Status: %d, Path: %s, Method: %s, RequestID: %s, Error: %v",
		statusCode, string(ctx.Path()), string(ctx.Method()), requestID, err)

	errorLogger.Println(logMsg)
	return nil
}

// statisticsHook 错误统计钩子
func StatisticsHook(ctx *mvccontext.Context, statusCode int, err error, errorStatistics *ErrorStatistics) error {
	if errorStatistics == nil {
		return nil
	}

	errorStatistics.mu.Lock()
	defer errorStatistics.mu.Unlock()

	// 更新统计信息
	errorStatistics.TotalErrors++
	errorStatistics.ErrorsByStatus[statusCode]++

	path := string(ctx.Path())
	errorStatistics.ErrorsByPath[path]++

	// 添加到最近错误列表（保留最近50个）
	record := ErrorRecord{
		StatusCode: statusCode,
		Path:       path,
		Method:     string(ctx.Method()),
		Timestamp:  time.Now(),
		UserAgent:  string(ctx.UserAgent()),
	}

	if err != nil {
		record.Message = err.Error()
	}

	if val, exists := ctx.Get("request_id"); exists {
		if id, ok := val.(string); ok {
			record.RequestID = id
		}
	}

	errorStatistics.LastErrors = append(errorStatistics.LastErrors, record)
	if len(errorStatistics.LastErrors) > MaxLastErrors {
		errorStatistics.LastErrors = errorStatistics.LastErrors[1:]
	}

	return nil
}

// cleanupHook 清理钩子
func CleanupHook(ctx *mvccontext.Context, statusCode int, err error) error {
	// 可以在这里进行资源清理，比如关闭数据库连接等
	// 当前实现为空，留给用户自定义
	return nil
}

// ExecuteHooks 执行钩子函数
func ExecuteHooks(hooks []ErrorHook, ctx *mvccontext.Context, statusCode int, err error, errorLogger *log.Logger) error {
	for _, hook := range hooks {
		if hookErr := hook(ctx, statusCode, err); hookErr != nil {
			// 钩子执行失败时记录日志但继续处理
			logError("Hook execution failed", hookErr, errorLogger)
			return hookErr
		}
	}
	return nil
}

// logError 内部日志记录方法
func logError(prefix string, err error, errorLogger *log.Logger) {
	if errorLogger != nil {
		errorLogger.Printf("%s: %v", prefix, err)
	}
}