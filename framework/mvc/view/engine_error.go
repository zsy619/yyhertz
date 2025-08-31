// Package view 提供错误处理功能
//
// 这个文件包含了模板引擎的错误处理相关功能，包括：
// - 错误记录和日志管理
// - 错误统计和分析
// - 调试信息显示
package view

import (
	"fmt"
	"html/template"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// TemplateErrorHandler 模板错误处理器
type TemplateErrorHandler struct {
	mutex         sync.RWMutex
	showDebugInfo bool               // 是否显示调试信息
	errorTemplate *template.Template // 错误模板
	errorLog      []TemplateError    // 错误日志
	maxErrorLog   int                // 最大错误日志数
}

// TemplateError 模板错误
type TemplateError struct {
	Timestamp   time.Time `json:"timestamp"`
	Template    string    `json:"template"`
	Error       string    `json:"error"`
	Stack       string    `json:"stack,omitempty"`
	RequestPath string    `json:"request_path,omitempty"`
}

// RecordError 记录模板错误
func (eh *TemplateErrorHandler) RecordError(templateName, errorMsg, requestPath string) {
	if eh == nil {
		return
	}

	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	// 获取堆栈信息
	stack := ""
	if eh.showDebugInfo {
		stack = eh.getStackTrace()
	}

	// 创建错误记录
	templateError := TemplateError{
		Timestamp:   time.Now(),
		Template:    templateName,
		Error:       errorMsg,
		Stack:       stack,
		RequestPath: requestPath,
	}

	// 添加到错误日志
	eh.errorLog = append(eh.errorLog, templateError)

	// 限制日志大小
	if len(eh.errorLog) > eh.maxErrorLog {
		eh.errorLog = eh.errorLog[len(eh.errorLog)-eh.maxErrorLog:]
	}

	// 记录到配置日志
	if eh.showDebugInfo {
		config.Errorf("Template error in '%s': %s", templateName, errorMsg)
		if stack != "" {
			config.Errorf("Stack trace: %s", stack)
		}
	} else {
		config.Errorf("Template error in '%s': %s", templateName, errorMsg)
	}
}

// GetErrorLog 获取错误日志
func (eh *TemplateErrorHandler) GetErrorLog() []TemplateError {
	if eh == nil {
		return nil
	}

	eh.mutex.RLock()
	defer eh.mutex.RUnlock()

	// 返回错误日志的副本
	logCopy := make([]TemplateError, len(eh.errorLog))
	copy(logCopy, eh.errorLog)
	return logCopy
}

// GetErrorCount 获取错误数量
func (eh *TemplateErrorHandler) GetErrorCount() int {
	if eh == nil {
		return 0
	}

	eh.mutex.RLock()
	defer eh.mutex.RUnlock()

	return len(eh.errorLog)
}

// ClearErrorLog 清空错误日志
func (eh *TemplateErrorHandler) ClearErrorLog() {
	if eh == nil {
		return
	}

	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	eh.errorLog = make([]TemplateError, 0)
	config.Infof("Template error log cleared")
}

// SetShowDebugInfo 设置是否显示调试信息
func (eh *TemplateErrorHandler) SetShowDebugInfo(show bool) {
	if eh == nil {
		return
	}

	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	eh.showDebugInfo = show
	config.Infof("Template error debug info display set to: %v", show)
}

// IsShowDebugInfo 检查是否显示调试信息
func (eh *TemplateErrorHandler) IsShowDebugInfo() bool {
	if eh == nil {
		return false
	}

	eh.mutex.RLock()
	defer eh.mutex.RUnlock()

	return eh.showDebugInfo
}

// SetMaxErrorLog 设置最大错误日志数
func (eh *TemplateErrorHandler) SetMaxErrorLog(max int) {
	if eh == nil {
		return
	}

	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	if max <= 0 {
		max = 100 // 默认值
	}

	eh.maxErrorLog = max
	
	// 如果当前日志数量超过新的限制，截断日志
	if len(eh.errorLog) > max {
		eh.errorLog = eh.errorLog[len(eh.errorLog)-max:]
	}

	config.Infof("Template error log max size set to: %d", max)
}

// GetMaxErrorLog 获取最大错误日志数
func (eh *TemplateErrorHandler) GetMaxErrorLog() int {
	if eh == nil {
		return 0
	}

	eh.mutex.RLock()
	defer eh.mutex.RUnlock()

	return eh.maxErrorLog
}

// getStackTrace 获取堆栈跟踪信息
func (eh *TemplateErrorHandler) getStackTrace() string {
	var buf strings.Builder
	
	// 跳过当前函数和调用者
	for i := 2; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		// 过滤掉不相关的堆栈信息
		funcName := fn.Name()
		if strings.Contains(funcName, "runtime.") {
			continue
		}
		
		buf.WriteString(fmt.Sprintf("%s:%d %s\n", file, line, funcName))
	}
	
	return buf.String()
}

// CreateErrorTemplate 创建错误显示模板
func (eh *TemplateErrorHandler) CreateErrorTemplate() error {
	if eh == nil {
		return fmt.Errorf("error handler is nil")
	}

	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	errorHTML := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>模板错误</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f8f8f8; }
        .error-container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .error-title { color: #d32f2f; font-size: 24px; margin-bottom: 20px; }
        .error-info { background: #fff3cd; padding: 15px; border-left: 4px solid #ffc107; margin-bottom: 20px; }
        .error-details { background: #f5f5f5; padding: 15px; border-radius: 4px; font-family: monospace; }
        .stack-trace { background: #f8f8f8; padding: 10px; margin-top: 10px; border-radius: 4px; font-size: 12px; }
        .timestamp { color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="error-container">
        <h1 class="error-title">模板渲染错误</h1>
        <div class="error-info">
            <p><strong>模板:</strong> {{.Template}}</p>
            <p><strong>错误:</strong> {{.Error}}</p>
            <p class="timestamp"><strong>时间:</strong> {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
            {{if .RequestPath}}<p><strong>请求路径:</strong> {{.RequestPath}}</p>{{end}}
        </div>
        {{if .Stack}}
        <div class="error-details">
            <h3>堆栈跟踪:</h3>
            <pre class="stack-trace">{{.Stack}}</pre>
        </div>
        {{end}}
    </div>
</body>
</html>`

	tmpl, err := template.New("error").Parse(errorHTML)
	if err != nil {
		return fmt.Errorf("failed to create error template: %w", err)
	}

	eh.errorTemplate = tmpl
	config.Infof("Error template created successfully")
	return nil
}

// RenderError 渲染错误页面
func (eh *TemplateErrorHandler) RenderError(templateError TemplateError) (string, error) {
	if eh == nil {
		return "", fmt.Errorf("error handler is nil")
	}

	eh.mutex.RLock()
	defer eh.mutex.RUnlock()

	if eh.errorTemplate == nil {
		if err := eh.CreateErrorTemplate(); err != nil {
			return "", fmt.Errorf("failed to create error template: %w", err)
		}
	}

	var buf strings.Builder
	if err := eh.errorTemplate.Execute(&buf, templateError); err != nil {
		return "", fmt.Errorf("failed to render error template: %w", err)
	}

	return buf.String(), nil
}

// ============= TemplateEngine 错误处理方法扩展 =============

// RecordTemplateError 记录模板错误（Engine方法）
func (e *TemplateEngine) RecordTemplateError(templateName, errorMsg, requestPath string) {
	// 记录到错误处理器
	if e.errorHandler != nil {
		e.errorHandler.RecordError(templateName, errorMsg, requestPath)
	}

	// 更新性能统计
	if e.performanceStats != nil {
		e.performanceStats.mutex.Lock()
		e.performanceStats.TemplateErrors++
		e.performanceStats.mutex.Unlock()
	}
}

// GetTemplateErrors 获取模板错误日志
func (e *TemplateEngine) GetTemplateErrors() []TemplateError {
	if e.errorHandler == nil {
		return nil
	}
	return e.errorHandler.GetErrorLog()
}

// ClearTemplateErrors 清空模板错误日志
func (e *TemplateEngine) ClearTemplateErrors() {
	if e.errorHandler != nil {
		e.errorHandler.ClearErrorLog()
	}

	// 重置性能统计中的错误计数
	if e.performanceStats != nil {
		e.performanceStats.mutex.Lock()
		e.performanceStats.TemplateErrors = 0
		e.performanceStats.mutex.Unlock()
	}
}

// GetTemplateErrorCount 获取模板错误数量
func (e *TemplateEngine) GetTemplateErrorCount() int {
	if e.errorHandler == nil {
		return 0
	}
	return e.errorHandler.GetErrorCount()
}

// SetErrorDebugMode 设置错误调试模式
func (e *TemplateEngine) SetErrorDebugMode(enable bool) {
	if e.errorHandler != nil {
		e.errorHandler.SetShowDebugInfo(enable)
	}
}

// IsErrorDebugMode 检查是否启用错误调试模式
func (e *TemplateEngine) IsErrorDebugMode() bool {
	if e.errorHandler == nil {
		return false
	}
	return e.errorHandler.IsShowDebugInfo()
}