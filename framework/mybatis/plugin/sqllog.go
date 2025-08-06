// Package plugin SQL日志插件实现
//
// 提供SQL执行日志记录功能，支持不同日志级别和格式化输出
package plugin

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// SqlLogPlugin SQL日志插件
type SqlLogPlugin struct {
	*BasePlugin
	logLevel     string // 日志级别
	logSql       bool   // 是否记录SQL
	logResult    bool   // 是否记录结果
	logParameter bool   // 是否记录参数
	logger       Logger // 日志记录器
}

// Logger 日志记录器接口
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// DefaultLogger 默认日志记录器
type DefaultLogger struct{}

// SqlLogEntry SQL日志条目
type SqlLogEntry struct {
	Timestamp     time.Time     // 时间戳
	Method        string        // 调用方法
	SQL           string        // SQL语句
	Parameters    []any         // 参数
	ExecutionTime time.Duration // 执行时间
	Success       bool          // 是否成功
	Error         error         // 错误信息
	RowsAffected  int64         // 影响行数
	Result        any           // 查询结果
}

// NewSqlLogPlugin 创建SQL日志插件
func NewSqlLogPlugin() *SqlLogPlugin {
	plugin := &SqlLogPlugin{
		BasePlugin:   NewBasePlugin("sqllog", 3),
		logLevel:     "INFO",
		logSql:       true,
		logResult:    false,
		logParameter: true,
		logger:       &DefaultLogger{},
	}
	return plugin
}

// Intercept 拦截方法调用
func (plugin *SqlLogPlugin) Intercept(invocation *Invocation) (any, error) {
	startTime := time.Now()

	// 执行原方法
	result, err := invocation.Proceed()

	// 记录日志
	plugin.logExecution(&SqlLogEntry{
		Timestamp:     startTime,
		Method:        invocation.Method.Name,
		SQL:           plugin.extractSQL(invocation),
		Parameters:    plugin.extractParameters(invocation),
		ExecutionTime: time.Since(startTime),
		Success:       err == nil,
		Error:         err,
		Result:        result,
	})

	return result, err
}

// Plugin 包装目标对象
func (plugin *SqlLogPlugin) Plugin(target any) any {
	return target
}

// SetProperties 设置插件属性
func (plugin *SqlLogPlugin) SetProperties(properties map[string]any) {
	plugin.BasePlugin.SetProperties(properties)

	plugin.logLevel = plugin.GetPropertyString("logLevel", "INFO")
	plugin.logSql = plugin.GetPropertyBool("logSql", true)
	plugin.logResult = plugin.GetPropertyBool("logResult", false)
	plugin.logParameter = plugin.GetPropertyBool("logParameter", true)
}

// SetLogger 设置日志记录器
func (plugin *SqlLogPlugin) SetLogger(logger Logger) {
	plugin.logger = logger
}

// extractSQL 提取SQL语句
func (plugin *SqlLogPlugin) extractSQL(invocation *Invocation) string {
	for _, arg := range invocation.Args {
		if sql, ok := arg.(string); ok && plugin.looksLikeSQL(sql) {
			return sql
		}
	}

	if sql, exists := invocation.Properties["sql"]; exists {
		if sqlStr, ok := sql.(string); ok {
			return sqlStr
		}
	}

	return ""
}

// extractParameters 提取参数
func (plugin *SqlLogPlugin) extractParameters(invocation *Invocation) []any {
	params := make([]any, 0)

	for _, arg := range invocation.Args {
		if sql, ok := arg.(string); ok && plugin.looksLikeSQL(sql) {
			continue
		}
		params = append(params, arg)
	}

	return params
}

// looksLikeSQL 判断字符串是否像SQL
func (plugin *SqlLogPlugin) looksLikeSQL(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	return strings.HasPrefix(s, "SELECT") ||
		strings.HasPrefix(s, "INSERT") ||
		strings.HasPrefix(s, "UPDATE") ||
		strings.HasPrefix(s, "DELETE") ||
		strings.HasPrefix(s, "CREATE") ||
		strings.HasPrefix(s, "DROP") ||
		strings.HasPrefix(s, "ALTER")
}

// logExecution 记录执行日志
func (plugin *SqlLogPlugin) logExecution(entry *SqlLogEntry) {
	if !plugin.shouldLog(entry) {
		return
	}

	message := plugin.formatLogMessage(entry)

	switch strings.ToUpper(plugin.logLevel) {
	case "DEBUG":
		plugin.logger.Debug(message)
	case "INFO":
		plugin.logger.Info(message)
	case "WARN":
		plugin.logger.Warn(message)
	case "ERROR":
		plugin.logger.Error(message)
	default:
		plugin.logger.Info(message)
	}
}

// shouldLog 判断是否应该记录日志
func (plugin *SqlLogPlugin) shouldLog(entry *SqlLogEntry) bool {
	// 根据配置决定是否记录
	if !plugin.logSql && entry.SQL != "" {
		return false
	}

	// 可以根据更多条件判断
	return true
}

// formatLogMessage 格式化日志消息
func (plugin *SqlLogPlugin) formatLogMessage(entry *SqlLogEntry) string {
	var builder strings.Builder

	// 基本信息
	builder.WriteString(fmt.Sprintf("[MyBatis] %s - %s",
		entry.Timestamp.Format("2006-01-02 15:04:05.000"),
		entry.Method))

	// SQL语句
	if plugin.logSql && entry.SQL != "" {
		builder.WriteString(fmt.Sprintf("\n  SQL: %s", plugin.formatSQL(entry.SQL)))
	}

	// 参数
	if plugin.logParameter && len(entry.Parameters) > 0 {
		builder.WriteString(fmt.Sprintf("\n  参数: %v", entry.Parameters))
	}

	// 执行时间
	builder.WriteString(fmt.Sprintf("\n  执行时间: %v", entry.ExecutionTime))

	// 结果信息
	if entry.Success {
		if entry.RowsAffected > 0 {
			builder.WriteString(fmt.Sprintf("\n  影响行数: %d", entry.RowsAffected))
		}

		if plugin.logResult && entry.Result != nil {
			builder.WriteString(fmt.Sprintf("\n  结果: %v", plugin.formatResult(entry.Result)))
		}
	} else {
		builder.WriteString(fmt.Sprintf("\n  错误: %v", entry.Error))
	}

	return builder.String()
}

// formatSQL 格式化SQL语句
func (plugin *SqlLogPlugin) formatSQL(sql string) string {
	// 简单的SQL格式化
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")

	// 移除多余空格
	for strings.Contains(sql, "  ") {
		sql = strings.ReplaceAll(sql, "  ", " ")
	}

	return strings.TrimSpace(sql)
}

// formatResult 格式化结果
func (plugin *SqlLogPlugin) formatResult(result any) string {
	if result == nil {
		return "null"
	}

	// 如果是切片，只显示长度
	if results, ok := result.([]any); ok {
		return fmt.Sprintf("列表[长度=%d]", len(results))
	}

	// 限制结果长度
	resultStr := fmt.Sprintf("%v", result)
	if len(resultStr) > 200 {
		return resultStr[:200] + "..."
	}

	return resultStr
}

// DefaultLogger 实现

// Debug 调试日志
func (logger *DefaultLogger) Debug(msg string, args ...any) {
	log.Printf("[DEBUG] "+msg, args...)
}

// Info 信息日志
func (logger *DefaultLogger) Info(msg string, args ...any) {
	log.Printf("[INFO] "+msg, args...)
}

// Warn 警告日志
func (logger *DefaultLogger) Warn(msg string, args ...any) {
	log.Printf("[WARN] "+msg, args...)
}

// Error 错误日志
func (logger *DefaultLogger) Error(msg string, args ...any) {
	log.Printf("[ERROR] "+msg, args...)
}

// SqlLogConfig SQL日志配置
type SqlLogConfig struct {
	Enabled      bool   `json:"enabled" yaml:"enabled"`           // 是否启用
	Level        string `json:"level" yaml:"level"`               // 日志级别
	LogSql       bool   `json:"logSql" yaml:"logSql"`             // 记录SQL
	LogParameter bool   `json:"logParameter" yaml:"logParameter"` // 记录参数
	LogResult    bool   `json:"logResult" yaml:"logResult"`       // 记录结果
	LogError     bool   `json:"logError" yaml:"logError"`         // 记录错误
	MaxSqlLength int    `json:"maxSqlLength" yaml:"maxSqlLength"` // SQL最大长度
}

// NewSqlLogConfig 创建默认SQL日志配置
func NewSqlLogConfig() *SqlLogConfig {
	return &SqlLogConfig{
		Enabled:      true,
		Level:        "INFO",
		LogSql:       true,
		LogParameter: true,
		LogResult:    false,
		LogError:     true,
		MaxSqlLength: 1000,
	}
}

// SqlLogFormatter SQL日志格式化器
type SqlLogFormatter interface {
	Format(entry *SqlLogEntry) string
}

// SimpleSqlLogFormatter 简单SQL日志格式化器
type SimpleSqlLogFormatter struct{}

// Format 格式化日志条目
func (formatter *SimpleSqlLogFormatter) Format(entry *SqlLogEntry) string {
	return fmt.Sprintf("[%s] %s - 执行时间: %v",
		entry.Timestamp.Format("15:04:05.000"),
		entry.Method,
		entry.ExecutionTime)
}

// DetailedSqlLogFormatter 详细SQL日志格式化器
type DetailedSqlLogFormatter struct{}

// Format 格式化日志条目
func (formatter *DetailedSqlLogFormatter) Format(entry *SqlLogEntry) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("==> 执行方法: %s\n", entry.Method))
	builder.WriteString(fmt.Sprintf("==> 执行时间: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05.000")))

	if entry.SQL != "" {
		builder.WriteString(fmt.Sprintf("==> SQL语句: %s\n", entry.SQL))
	}

	if len(entry.Parameters) > 0 {
		builder.WriteString(fmt.Sprintf("==> 参数列表: %v\n", entry.Parameters))
	}

	builder.WriteString(fmt.Sprintf("==> 耗时: %v\n", entry.ExecutionTime))

	if entry.Success {
		builder.WriteString("==> 执行成功")
		if entry.RowsAffected > 0 {
			builder.WriteString(fmt.Sprintf(", 影响行数: %d", entry.RowsAffected))
		}
	} else {
		builder.WriteString(fmt.Sprintf("==> 执行失败: %v", entry.Error))
	}

	return builder.String()
}
