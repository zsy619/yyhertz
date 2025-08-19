package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	IgnorePackages = []string{
		"github.com/sirupsen/logrus",
		"github.com/zsy619/yyhertz/framework/config",
		"runtime",
		"strings",
	}
)

const (
	// 根据您的封装层数调整（通常 4-6）
	defaultSkip = 7
)

// BeegoFormatter Beego风格的日志格式化器
type BeegoFormatter struct {
	TimestampFormat string
	ShowCaller      bool
}

// Format 实现 logrus.Formatter 接口
func (f *BeegoFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 获取日志级别标识
	levelChar := f.getLevelChar(entry.Level)

	// 格式化时间
	timestamp := entry.Time.Format(f.TimestampFormat)
	if f.TimestampFormat == "" {
		timestamp = entry.Time.Format("2006/01/02 15:04:05.000")
	}

	// 构建基础日志信息
	logLine := fmt.Sprintf("[%s] %s", levelChar, timestamp)

	// 添加调用位置信息
	if f.ShowCaller && entry.HasCaller() {
		// // 获取正确的调用栈
		// pc := make([]uintptr, 1)
		// num := runtime.Callers(defaultSkip, pc)
		// if num > 0 {
		// 	frames := runtime.CallersFrames(pc)
		// 	frame, _ := frames.Next()

		// 	// 更新调用者信息
		// 	entry.Caller = &frame
		// 	filename := f.getShortFilename(entry.Caller.File)
		// 	logLine += fmt.Sprintf(" [%s:%d]", filename, entry.Caller.Line)
		// }
	}

	// 添加日志消息
	logLine += fmt.Sprintf(" %s", entry.Message)

	// 添加字段信息（如果有）
	if len(entry.Data) > 0 {
		fields := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		logLine += fmt.Sprintf(" {%s}", strings.Join(fields, ", "))
	}

	logLine += "\n"
	return []byte(logLine), nil
}

// getLevelChar 获取日志级别字符
func (f *BeegoFormatter) getLevelChar(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "D"
	case logrus.InfoLevel:
		return "I"
	case logrus.WarnLevel:
		return "W"
	case logrus.ErrorLevel:
		return "E"
	case logrus.FatalLevel:
		return "F"
	case logrus.PanicLevel:
		return "P"
	default:
		return "I"
	}
}

// getShortFilename 获取短文件名
func (f *BeegoFormatter) getShortFilename(fullPath string) string {
	// 只保留文件名，不包含完整路径
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// Log4GoFormatter Log4Go风格的日志格式化器
type Log4GoFormatter struct {
	TimestampFormat string
	ShowCaller      bool
}

// Format 实现 logrus.Formatter 接口
func (f *Log4GoFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 格式化时间
	timestamp := entry.Time.Format(f.TimestampFormat)
	if f.TimestampFormat == "" {
		timestamp = entry.Time.Format("2006/01/02 15:04:05")
	}

	// 获取级别名称
	levelName := strings.ToUpper(entry.Level.String())

	// 构建基础日志信息
	logLine := fmt.Sprintf("[%s] [%s]", timestamp, levelName)

	// 添加调用位置信息
	if f.ShowCaller && entry.HasCaller() {
		filename := f.getShortFilename(entry.Caller.File)
		logLine += fmt.Sprintf(" (%s:%d)", filename, entry.Caller.Line)
	}

	// 添加日志消息
	logLine += fmt.Sprintf(" %s", entry.Message)

	// 添加字段信息（如果有）
	if len(entry.Data) > 0 {
		fields := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		logLine += fmt.Sprintf(" [%s]", strings.Join(fields, ", "))
	}

	logLine += "\n"
	return []byte(logLine), nil
}

// getShortFilename 获取短文件名
func (f *Log4GoFormatter) getShortFilename(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// LogstashFormatter Logstash格式化器
type LogstashFormatter struct {
	TimestampFormat string
	ServiceName     string
	Version         string
}

// Format 实现 logrus.Formatter 接口
func (f *LogstashFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 创建Logstash格式的数据结构
	data := logrus.Fields{
		"@timestamp": entry.Time.Format(time.RFC3339),
		"@version":   "1",
		"level":      entry.Level.String(),
		"message":    entry.Message,
		"logger":     "yyhertz",
	}

	// 添加服务信息
	if f.ServiceName != "" {
		data["service"] = f.ServiceName
	}
	if f.Version != "" {
		data["version"] = f.Version
	}

	// 添加调用位置信息
	if entry.HasCaller() {
		data["file"] = entry.Caller.File
		data["line"] = entry.Caller.Line
		data["function"] = entry.Caller.Function
	}

	// 合并字段数据
	for k, v := range entry.Data {
		data[k] = v
	}

	// 序列化为JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 添加换行符
	bytes = append(bytes, '\n')
	return bytes, nil
}

// SyslogFormatter Syslog格式化器
type SyslogFormatter struct {
	TimestampFormat string
	Hostname        string
	Tag             string
	Facility        int
}

// Format 实现 logrus.Formatter 接口
func (f *SyslogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 计算优先级
	priority := f.Facility*8 + f.getSyslogSeverity(entry.Level)

	// 格式化时间（RFC3164格式）
	timestamp := entry.Time.Format("Jan 2 15:04:05")
	if f.TimestampFormat != "" {
		timestamp = entry.Time.Format(f.TimestampFormat)
	}

	// 获取主机名
	hostname := f.Hostname
	if hostname == "" {
		hostname = "localhost"
	}

	// 获取标签
	tag := f.Tag
	if tag == "" {
		tag = "yyhertz"
	}

	// 构建Syslog消息
	message := fmt.Sprintf("<%d>%s %s %s: %s", priority, timestamp, hostname, tag, entry.Message)

	// 添加字段信息
	if len(entry.Data) > 0 {
		fields := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		message += fmt.Sprintf(" [%s]", strings.Join(fields, " "))
	}

	message += "\n"
	return []byte(message), nil
}

// getSyslogSeverity 获取Syslog严重级别
func (f *SyslogFormatter) getSyslogSeverity(level logrus.Level) int {
	switch level {
	case logrus.PanicLevel:
		return 0 // Emergency
	case logrus.FatalLevel:
		return 2 // Critical
	case logrus.ErrorLevel:
		return 3 // Error
	case logrus.WarnLevel:
		return 4 // Warning
	case logrus.InfoLevel:
		return 6 // Info
	case logrus.DebugLevel:
		return 7 // Debug
	default:
		return 6 // Info
	}
}

// FluentdFormatter Fluentd格式化器
type FluentdFormatter struct {
	TimestampFormat string
	ServiceName     string
	Environment     string
}

// Format 实现 logrus.Formatter 接口
func (f *FluentdFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 创建Fluentd格式的数据结构
	data := logrus.Fields{
		"timestamp": entry.Time.Format(time.RFC3339),
		"level":     entry.Level.String(),
		"message":   entry.Message,
	}

	// 添加服务信息
	if f.ServiceName != "" {
		data["service"] = f.ServiceName
	}
	if f.Environment != "" {
		data["environment"] = f.Environment
	}

	// 添加调用位置信息
	if entry.HasCaller() {
		data["source"] = map[string]any{
			"file":     entry.Caller.File,
			"line":     entry.Caller.Line,
			"function": entry.Caller.Function,
		}
	}

	// 合并字段数据
	for k, v := range entry.Data {
		data[k] = v
	}

	// 序列化为JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, '\n')
	return bytes, nil
}

// CloudWatchFormatter AWS CloudWatch格式化器
type CloudWatchFormatter struct {
	TimestampFormat string
	ServiceName     string
	LogGroupName    string
	LogStreamName   string
}

// Format 实现 logrus.Formatter 接口
func (f *CloudWatchFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 创建CloudWatch格式的数据结构
	data := logrus.Fields{
		"timestamp":    entry.Time.UnixMilli(),
		"level":        entry.Level.String(),
		"message":      entry.Message,
		"logGroup":     f.LogGroupName,
		"logStream":    f.LogStreamName,
		"awsRequestId": "", // 可以从context中获取
	}

	// 添加服务信息
	if f.ServiceName != "" {
		data["service"] = f.ServiceName
	}

	// 添加调用位置信息
	if entry.HasCaller() {
		data["source"] = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
	}

	// 合并字段数据
	for k, v := range entry.Data {
		data[k] = v
	}

	// 序列化为JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, '\n')
	return bytes, nil
}

// AzureInsightsFormatter Azure Application Insights格式化器
type AzureInsightsFormatter struct {
	TimestampFormat    string
	ServiceName        string
	InstrumentationKey string
}

// Format 实现 logrus.Formatter 接口
func (f *AzureInsightsFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 创建Application Insights格式的数据结构
	data := map[string]any{
		"time": entry.Time.Format(time.RFC3339),
		"iKey": f.InstrumentationKey,
		"name": "Microsoft.ApplicationInsights.Message",
		"tags": map[string]string{
			"ai.application.ver": "1.0.0",
			"ai.cloud.role":      f.ServiceName,
		},
		"data": map[string]any{
			"baseType": "MessageData",
			"baseData": map[string]any{
				"ver":           2,
				"message":       entry.Message,
				"severityLevel": f.getInsightsSeverity(entry.Level),
				"properties":    entry.Data,
			},
		},
	}

	// 添加调用位置信息
	if entry.HasCaller() {
		baseData := data["data"].(map[string]any)["baseData"].(map[string]any)
		baseData["source"] = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
	}

	// 序列化为JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, '\n')
	return bytes, nil
}

// getInsightsSeverity 获取Application Insights严重级别
func (f *AzureInsightsFormatter) getInsightsSeverity(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel:
		return 0 // Verbose
	case logrus.InfoLevel:
		return 1 // Information
	case logrus.WarnLevel:
		return 2 // Warning
	case logrus.ErrorLevel:
		return 3 // Error
	case logrus.FatalLevel, logrus.PanicLevel:
		return 4 // Critical
	default:
		return 1 // Information
	}
}

// ============= 流行框架格式化器 =============

// GinFormatter Gin框架日志格式化器
// 格式: [GIN] 2006/01/02 - 15:04:05 | 200 | 13ms | 127.0.0.1 | GET | /api/users
type GinFormatter struct {
	TimestampFormat string
	ShowColors      bool
}

// Format 实现 logrus.Formatter 接口
func (f *GinFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 格式化时间
	timestamp := entry.Time.Format("2006/01/02 - 15:04:05")
	if f.TimestampFormat != "" {
		timestamp = entry.Time.Format(f.TimestampFormat)
	}

	// 获取日志级别标识
	levelChar := f.getLevelChar(entry.Level)
	// 获取状态码、延迟等信息
	statusCode := f.getFieldString(entry, "status_code", "200")
	latency := f.getFieldString(entry, "latency", "0ms")
	clientIP := f.getFieldString(entry, "client_ip", "127.0.0.1")
	method := f.getFieldString(entry, "method", "GET")
	path := f.getFieldString(entry, "path", "/")

	// 构建Gin风格的日志格式
	var message string
	if f.ShowColors {
		statusColor := f.getStatusColor(statusCode)
		methodColor := f.getMethodColor(method)
		resetColor := "\033[0m"

		message = fmt.Sprintf("[GIN] %s %s |%s %3s %s| %13s | %15s |%s %-7s %s %s",
			levelChar,
			timestamp,
			statusColor, statusCode, resetColor,
			latency,
			clientIP,
			methodColor, method, resetColor,
			path)
	} else {
		message = fmt.Sprintf("[GIN] %s %s | %3s | %13s | %15s | %-7s %s",
			levelChar, timestamp, statusCode, latency, clientIP, method, path)
	}

	// 添加原始消息
	if entry.Message != "" {
		message += fmt.Sprintf(" %s", entry.Message)
	}

	// 添加错误信息
	if errorMsg := f.getFieldString(entry, "error", ""); errorMsg != "" {
		message += fmt.Sprintf("\n%s", errorMsg)
	}

	message += "\n"
	return []byte(message), nil
}

// getFieldString 安全获取字段值
func (f *GinFormatter) getFieldString(entry *logrus.Entry, key, defaultValue string) string {
	if value, exists := entry.Data[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}

// getStatusColor 根据状态码获取颜色
func (f *GinFormatter) getStatusColor(statusCode string) string {
	switch {
	case strings.HasPrefix(statusCode, "2"):
		return "\033[97;42m" // 绿色背景
	case strings.HasPrefix(statusCode, "3"):
		return "\033[90;47m" // 白色背景
	case strings.HasPrefix(statusCode, "4"):
		return "\033[90;43m" // 黄色背景
	case strings.HasPrefix(statusCode, "5"):
		return "\033[97;41m" // 红色背景
	default:
		return "\033[97;46m" // 青色背景
	}
}

// getMethodColor 根据方法获取颜色
func (f *GinFormatter) getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[97;44m" // 蓝色背景
	case "POST":
		return "\033[97;46m" // 青色背景
	case "PUT":
		return "\033[97;43m" // 黄色背景
	case "DELETE":
		return "\033[97;41m" // 红色背景
	case "PATCH":
		return "\033[97;42m" // 绿色背景
	case "HEAD":
		return "\033[97;45m" // 品红色背景
	case "OPTIONS":
		return "\033[90;47m" // 白色背景
	default:
		return "\033[0m" // 重置颜色
	}
}

func (f *GinFormatter) getLevelChar(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "[D]"
	case logrus.InfoLevel:
		return "[I]"
	case logrus.WarnLevel:
		return "[W]"
	case logrus.ErrorLevel:
		return "[E]"
	case logrus.FatalLevel:
		return "[F]"
	case logrus.PanicLevel:
		return "[P]"
	default:
		return "[I]"
	}
}

// IrisFormatter Iris框架日志格式化器
// 格式: [IRIS] 2006/01/02 15:04:05 | 200 | 13ms | 127.0.0.1 | GET /api/users
type IrisFormatter struct {
	TimestampFormat string
	ShowColors      bool
}

// Format 实现 logrus.Formatter 接口
func (f *IrisFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 格式化时间
	timestamp := entry.Time.Format("2006/01/02 15:04:05")
	if f.TimestampFormat != "" {
		timestamp = entry.Time.Format(f.TimestampFormat)
	}

	// 获取字段信息
	statusCode := f.getFieldString(entry, "status_code", "200")
	latency := f.getFieldString(entry, "latency", "0ms")
	clientIP := f.getFieldString(entry, "client_ip", "127.0.0.1")
	method := f.getFieldString(entry, "method", "GET")
	path := f.getFieldString(entry, "path", "/")

	// 构建Iris风格的日志格式
	message := fmt.Sprintf("[IRIS] %s | %s | %s | %s | %s %s",
		timestamp, statusCode, latency, clientIP, method, path)

	// 添加原始消息
	if entry.Message != "" {
		message += fmt.Sprintf(" - %s", entry.Message)
	}

	message += "\n"
	return []byte(message), nil
}

// getFieldString 安全获取字段值
func (f *IrisFormatter) getFieldString(entry *logrus.Entry, key, defaultValue string) string {
	if value, exists := entry.Data[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}

// EntFormatter Ent ORM日志格式化器
// 格式: [ENT] 2006/01/02 15:04:05 [INFO] SELECT * FROM users WHERE id = ? [1] (13ms)
type EntFormatter struct {
	TimestampFormat string
	ShowSQL         bool
}

// Format 实现 logrus.Formatter 接口
func (f *EntFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 格式化时间
	timestamp := entry.Time.Format("2006/01/02 15:04:05")
	if f.TimestampFormat != "" {
		timestamp = entry.Time.Format(f.TimestampFormat)
	}

	// 获取级别
	level := strings.ToUpper(entry.Level.String())

	// 构建Ent风格的日志格式
	message := fmt.Sprintf("[ENT] %s [%s]", timestamp, level)

	// 添加SQL信息（如果存在）
	if f.ShowSQL {
		if sql := f.getFieldString(entry, "sql", ""); sql != "" {
			message += fmt.Sprintf(" %s", sql)
		}
		if args := f.getFieldString(entry, "args", ""); args != "" {
			message += fmt.Sprintf(" %s", args)
		}
	}

	// 添加执行时间
	if duration := f.getFieldString(entry, "duration", ""); duration != "" {
		message += fmt.Sprintf(" (%s)", duration)
	}

	// 添加原始消息
	if entry.Message != "" {
		message += fmt.Sprintf(" %s", entry.Message)
	}

	message += "\n"
	return []byte(message), nil
}

// getFieldString 安全获取字段值
func (f *EntFormatter) getFieldString(entry *logrus.Entry, key, defaultValue string) string {
	if value, exists := entry.Data[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}

// GoZeroFormatter go-zero微服务框架日志格式化器
// JSON格式，包含服务名、追踪ID等微服务相关字段
type GoZeroFormatter struct {
	ServiceName string
	Environment string
}

// Format 实现 logrus.Formatter 接口
func (f *GoZeroFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 创建go-zero格式的数据结构
	data := logrus.Fields{
		"@timestamp": entry.Time.Format(time.RFC3339),
		"level":      entry.Level.String(),
		"service":    f.ServiceName,
		"message":    entry.Message,
	}

	// 添加环境信息
	if f.Environment != "" {
		data["env"] = f.Environment
	}

	// 添加微服务相关字段
	if traceID := f.getFieldString(entry, "trace_id", ""); traceID != "" {
		data["trace"] = traceID
	}
	if spanID := f.getFieldString(entry, "span_id", ""); spanID != "" {
		data["span"] = spanID
	}

	// 添加调用位置信息
	if entry.HasCaller() {
		data["caller"] = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
	}

	// 合并其他字段
	for k, v := range entry.Data {
		if k != "trace_id" && k != "span_id" {
			data[k] = v
		}
	}

	// 序列化为JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, '\n')
	return bytes, nil
}

// getFieldString 安全获取字段值
func (f *GoZeroFormatter) getFieldString(entry *logrus.Entry, key, defaultValue string) string {
	if value, exists := entry.Data[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}

// FiberFormatter Fiber框架日志格式化器
// 格式: 15:04:05 | 200 | 13ms | 127.0.0.1 | GET | /api/users
type FiberFormatter struct {
	TimestampFormat string
}

// Format 实现 logrus.Formatter 接口
func (f *FiberFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 格式化时间
	timestamp := entry.Time.Format("15:04:05")
	if f.TimestampFormat != "" {
		timestamp = entry.Time.Format(f.TimestampFormat)
	}

	// 获取字段信息
	statusCode := f.getFieldString(entry, "status_code", "200")
	latency := f.getFieldString(entry, "latency", "0ms")
	clientIP := f.getFieldString(entry, "client_ip", "127.0.0.1")
	method := f.getFieldString(entry, "method", "GET")
	path := f.getFieldString(entry, "path", "/")

	// 构建Fiber风格的日志格式
	message := fmt.Sprintf("%s | %s | %s | %s | %s | %s",
		timestamp, statusCode, latency, clientIP, method, path)

	// 添加原始消息
	if entry.Message != "" {
		message += fmt.Sprintf(" - %s", entry.Message)
	}

	message += "\n"
	return []byte(message), nil
}

// getFieldString 安全获取字段值
func (f *FiberFormatter) getFieldString(entry *logrus.Entry, key, defaultValue string) string {
	if value, exists := entry.Data[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}

// EchoFormatter Echo框架日志格式化器
// JSON格式，包含时间、级别、前缀、文件、消息
type EchoFormatter struct {
	Prefix string
}

// Format 实现 logrus.Formatter 接口
func (f *EchoFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 创建Echo格式的数据结构
	data := logrus.Fields{
		"time":    entry.Time.Format(time.RFC3339),
		"level":   strings.ToUpper(entry.Level.String()),
		"prefix":  f.Prefix,
		"message": entry.Message,
	}

	// 添加文件位置信息
	if entry.HasCaller() {
		data["file"] = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
	}

	// 合并其他字段
	for k, v := range entry.Data {
		data[k] = v
	}

	// 序列化为JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, '\n')
	return bytes, nil
}

// RevelFormatter Revel MVC框架日志格式化器
// 格式: INFO 2006/01/02 15:04:05 controller.go:123: Request processed
type RevelFormatter struct {
	TimestampFormat string
	ShowCaller      bool
}

// Format 实现 logrus.Formatter 接口
func (f *RevelFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 格式化时间
	timestamp := entry.Time.Format("2006/01/02 15:04:05")
	if f.TimestampFormat != "" {
		timestamp = entry.Time.Format(f.TimestampFormat)
	}

	// 获取级别
	level := strings.ToUpper(entry.Level.String())

	// 构建Revel风格的日志格式
	message := fmt.Sprintf("%s %s", level, timestamp)

	// 添加调用位置信息
	if f.ShowCaller && entry.HasCaller() {
		filename := f.getShortFilename(entry.Caller.File)
		message += fmt.Sprintf(" %s:%d:", filename, entry.Caller.Line)
	}

	// 添加原始消息
	message += fmt.Sprintf(" %s", entry.Message)

	// 添加字段信息
	if len(entry.Data) > 0 {
		fields := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		message += fmt.Sprintf(" [%s]", strings.Join(fields, " "))
	}

	message += "\n"
	return []byte(message), nil
}

// getShortFilename 获取短文件名
func (f *RevelFormatter) getShortFilename(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// BuffaloFormatter Buffalo框架日志格式化器
// 格式: --> GET /api/users (13ms) | 127.0.0.1 | 200 OK
type BuffaloFormatter struct {
	TimestampFormat string
}

// Format 实现 logrus.Formatter 接口
func (f *BuffaloFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 获取字段信息
	method := f.getFieldString(entry, "method", "GET")
	path := f.getFieldString(entry, "path", "/")
	latency := f.getFieldString(entry, "latency", "0ms")
	clientIP := f.getFieldString(entry, "client_ip", "127.0.0.1")
	statusCode := f.getFieldString(entry, "status_code", "200")

	// 构建Buffalo风格的日志格式
	message := fmt.Sprintf("--> %s %s (%s) | %s | %s",
		method, path, latency, clientIP, statusCode)

	// 根据状态码添加状态描述
	if strings.HasPrefix(statusCode, "2") {
		message += " OK"
	} else if strings.HasPrefix(statusCode, "4") {
		message += " Client Error"
	} else if strings.HasPrefix(statusCode, "5") {
		message += " Server Error"
	}

	// 添加原始消息
	if entry.Message != "" {
		message += fmt.Sprintf(" - %s", entry.Message)
	}

	message += "\n"
	return []byte(message), nil
}

// getFieldString 安全获取字段值
func (f *BuffaloFormatter) getFieldString(entry *logrus.Entry, key, defaultValue string) string {
	if value, exists := entry.Data[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}

// GetFormatter 根据格式类型获取对应的格式化器
func GetFormatter(format LogFormat, config *LogConfig) logrus.Formatter {
	switch format {
	case LogFormatBeego:
		return &BeegoFormatter{
			TimestampFormat: config.TimestampFormat,
			ShowCaller:      config.ShowCaller,
		}
	case LogFormatLog4Go:
		return &Log4GoFormatter{
			TimestampFormat: config.TimestampFormat,
			ShowCaller:      config.ShowCaller,
		}
	case LogFormatLogstash:
		serviceName := "yyhertz"
		version := "1.0.0"
		if config.Fields != nil {
			if s, ok := config.Fields["service"].(string); ok {
				serviceName = s
			}
			if v, ok := config.Fields["version"].(string); ok {
				version = v
			}
		}
		return &LogstashFormatter{
			TimestampFormat: config.TimestampFormat,
			ServiceName:     serviceName,
			Version:         version,
		}
	case LogFormatSyslog:
		hostname := "localhost"
		tag := "yyhertz"
		if config.Fields != nil {
			if h, ok := config.Fields["hostname"].(string); ok {
				hostname = h
			}
			if t, ok := config.Fields["tag"].(string); ok {
				tag = t
			}
		}
		return &SyslogFormatter{
			TimestampFormat: config.TimestampFormat,
			Hostname:        hostname,
			Tag:             tag,
			Facility:        16, // local0
		}
	case LogFormatFluentd:
		serviceName := "yyhertz"
		environment := "production"
		if config.Fields != nil {
			if s, ok := config.Fields["service"].(string); ok {
				serviceName = s
			}
			if e, ok := config.Fields["environment"].(string); ok {
				environment = e
			}
		}
		return &FluentdFormatter{
			TimestampFormat: config.TimestampFormat,
			ServiceName:     serviceName,
			Environment:     environment,
		}
	case LogFormatCloudWatch:
		serviceName := "yyhertz"
		logGroupName := "/aws/yyhertz/application"
		logStreamName := "yyhertz-instance-001"
		if config.Fields != nil {
			if s, ok := config.Fields["service"].(string); ok {
				serviceName = s
			}
		}
		if outputConfig, exists := config.GetOutputConfig("cloudwatch"); exists {
			if cwConfig, ok := outputConfig.(CloudWatchConfig); ok {
				logGroupName = cwConfig.LogGroupName
				logStreamName = cwConfig.LogStreamName
			}
		}
		return &CloudWatchFormatter{
			TimestampFormat: config.TimestampFormat,
			ServiceName:     serviceName,
			LogGroupName:    logGroupName,
			LogStreamName:   logStreamName,
		}
	case LogFormatApplicationInsights:
		serviceName := "yyhertz"
		instrumentationKey := ""
		if config.Fields != nil {
			if s, ok := config.Fields["service"].(string); ok {
				serviceName = s
			}
		}
		if outputConfig, exists := config.GetOutputConfig("azure_insights"); exists {
			if aiConfig, ok := outputConfig.(AzureInsightsConfig); ok {
				instrumentationKey = aiConfig.InstrumentationKey
			}
		}
		return &AzureInsightsFormatter{
			TimestampFormat:    config.TimestampFormat,
			ServiceName:        serviceName,
			InstrumentationKey: instrumentationKey,
		}
	case LogFormatJSON:
		return &logrus.JSONFormatter{
			TimestampFormat: config.TimestampFormat,
		}
	case LogFormatText:
		return &logrus.TextFormatter{
			FullTimestamp:   config.ShowTimestamp,
			TimestampFormat: config.TimestampFormat,
			DisableColors:   false,
		}
	case LogFormatGin:
		return &GinFormatter{
			TimestampFormat: config.TimestampFormat,
			ShowColors:      true, // Gin通常显示彩色输出
		}
	case LogFormatIris:
		return &IrisFormatter{
			TimestampFormat: config.TimestampFormat,
			ShowColors:      false, // Iris默认不显示颜色
		}
	case LogFormatEnt:
		return &EntFormatter{
			TimestampFormat: config.TimestampFormat,
			ShowSQL:         true, // Ent默认显示SQL
		}
	case LogFormatGoZero:
		serviceName := "yyhertz"
		environment := "production"
		if config.Fields != nil {
			if s, ok := config.Fields["service"].(string); ok {
				serviceName = s
			}
			if e, ok := config.Fields["environment"].(string); ok {
				environment = e
			}
		}
		return &GoZeroFormatter{
			ServiceName: serviceName,
			Environment: environment,
		}
	case LogFormatFiber:
		return &FiberFormatter{
			TimestampFormat: config.TimestampFormat,
		}
	case LogFormatEcho:
		prefix := "echo"
		if config.Fields != nil {
			if p, ok := config.Fields["prefix"].(string); ok {
				prefix = p
			}
		}
		return &EchoFormatter{
			Prefix: prefix,
		}
	case LogFormatRevel:
		return &RevelFormatter{
			TimestampFormat: config.TimestampFormat,
			ShowCaller:      config.ShowCaller,
		}
	case LogFormatBuffalo:
		return &BuffaloFormatter{
			TimestampFormat: config.TimestampFormat,
		}
	default:
		return &BeegoFormatter{
			TimestampFormat: config.TimestampFormat,
			ShowCaller:      config.ShowCaller,
		}
	}
}
