package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
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
		filename := f.getShortFilename(entry.Caller.File)
		logLine += fmt.Sprintf(" [%s:%d]", filename, entry.Caller.Line)
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
		data["source"] = map[string]interface{}{
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
		"timestamp":      entry.Time.UnixMilli(),
		"level":          entry.Level.String(),
		"message":        entry.Message,
		"logGroup":       f.LogGroupName,
		"logStream":      f.LogStreamName,
		"awsRequestId":   "", // 可以从context中获取
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
	data := map[string]interface{}{
		"time":    entry.Time.Format(time.RFC3339),
		"iKey":    f.InstrumentationKey,
		"name":    "Microsoft.ApplicationInsights.Message",
		"tags": map[string]string{
			"ai.application.ver": "1.0.0",
			"ai.cloud.role":      f.ServiceName,
		},
		"data": map[string]interface{}{
			"baseType": "MessageData",
			"baseData": map[string]interface{}{
				"ver":         2,
				"message":     entry.Message,
				"severityLevel": f.getInsightsSeverity(entry.Level),
				"properties":  entry.Data,
			},
		},
	}
	
	// 添加调用位置信息
	if entry.HasCaller() {
		baseData := data["data"].(map[string]interface{})["baseData"].(map[string]interface{})
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
	default:
		return &BeegoFormatter{
			TimestampFormat: config.TimestampFormat,
			ShowCaller:      config.ShowCaller,
		}
	}
}