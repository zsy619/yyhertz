package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// ============= 全局便捷函数 =============

// Debug 全局调试日志
func Debug(args ...any) {
	GetGlobalLogger().Debug(args...)
}

// Debugf 全局格式化调试日志
func Debugf(format string, args ...any) {
	GetGlobalLogger().Debugf(format, args...)
}

// Info 全局信息日志
func Info(args ...any) {
	GetGlobalLogger().Info(args...)
}

// Infof 全局格式化信息日志
func Infof(format string, args ...any) {
	GetGlobalLogger().Infof(format, args...)
}

// Warn 全局警告日志
func Warn(args ...any) {
	GetGlobalLogger().Warn(args...)
}

// Warnf 全局格式化警告日志
func Warnf(format string, args ...any) {
	GetGlobalLogger().Warnf(format, args...)
}

// Error 全局错误日志
func Error(args ...any) {
	GetGlobalLogger().Error(args...)
}

// Errorf 全局格式化错误日志
func Errorf(format string, args ...any) {
	GetGlobalLogger().Errorf(format, args...)
}

// Fatal 全局致命错误日志
func Fatal(args ...any) {
	GetGlobalLogger().Fatal(args...)
}

// Fatalf 全局格式化致命错误日志
func Fatalf(format string, args ...any) {
	GetGlobalLogger().Fatalf(format, args...)
}

// Panic 全局panic日志
func Panic(args ...any) {
	GetGlobalLogger().Panic(args...)
}

// Panicf 全局格式化panic日志
func Panicf(format string, args ...any) {
	GetGlobalLogger().Panicf(format, args...)
}

// WithFields 全局带字段日志
func WithFields(fields map[string]any) *logrus.Entry {
	return GetGlobalLogger().WithFields(fields)
}

// WithField 全局带单个字段日志
func WithField(key string, value any) *logrus.Entry {
	return GetGlobalLogger().WithField(key, value)
}

// WithError 全局带错误字段日志
func WithError(err error) *logrus.Entry {
	return GetGlobalLogger().WithError(err)
}

// ============= 改进的全局请求ID函数 =============

var (
	// 请求ID验证正则表达式：支持UUID、时间戳、自定义格式等
	requestIDPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_\.]{8,64}$`)
)

// WithRequestID 全局添加请求ID字段，包含防重复和校验功能
// 参数校验：
// - requestID不能为空
// - 长度必须在8-64字符之间
// - 只允许字母、数字、连字符、下划线、点号
// 防重复：如果当前Entry已有request_id字段，会在日志中添加警告但不会覆盖
func WithRequestID(requestID string) *logrus.Entry {
	// 1. 参数校验
	if err := validateRequestID(requestID); err != nil {
		// 记录校验错误并返回带错误信息的Entry
		return GetGlobalLogger().WithError(err).WithField("invalid_request_id", requestID)
	}

	// 2. 获取全局logger的Entry
	globalLogger := GetGlobalLogger()
	baseEntry := globalLogger.GetRawLogger().WithFields(logrus.Fields{})

	// 3. 检查是否已存在request_id字段（防重复）
	if existingID, exists := baseEntry.Data["request_id"]; exists {
		// 如果已存在相同的request_id，直接返回
		if existingID == requestID {
			return baseEntry
		}
		// 如果存在不同的request_id，添加警告但不覆盖
		return baseEntry.WithField("request_id_conflict", fmt.Sprintf("尝试覆盖现有request_id: %v -> %s", existingID, requestID))
	}

	// 4. 添加新的request_id
	return baseEntry.WithField("request_id", requestID)
}

// validateRequestID 验证请求ID的合法性
func validateRequestID(requestID string) error {
	// 检查是否为空
	if requestID == "" {
		return fmt.Errorf("request_id不能为空")
	}

	// 检查长度
	if len(requestID) < 8 {
		return fmt.Errorf("request_id长度不能少于8个字符，当前长度: %d", len(requestID))
	}
	if len(requestID) > 64 {
		return fmt.Errorf("request_id长度不能超过64个字符，当前长度: %d", len(requestID))
	}

	// 检查字符格式
	if !requestIDPattern.MatchString(requestID) {
		return fmt.Errorf("request_id格式无效，只允许字母、数字、连字符(-)、下划线(_)、点号(.)，当前值: %s", requestID)
	}

	// 检查是否包含连续的特殊字符（可选的额外验证）
	if strings.Contains(requestID, "--") || strings.Contains(requestID, "__") || strings.Contains(requestID, "..") {
		return fmt.Errorf("request_id不能包含连续的特殊字符，当前值: %s", requestID)
	}

	return nil
}

// WithRequestIDUnsafe 不安全版本的WithRequestID，跳过所有校验（仅用于特殊场景）
// 警告：使用此函数需要确保requestID已经过外部验证
func WithRequestIDUnsafe(requestID string) *logrus.Entry {
	return GetGlobalLogger().WithField("request_id", requestID)
}

// IsValidRequestID 检查请求ID是否有效（公开的校验函数）
func IsValidRequestID(requestID string) bool {
	return validateRequestID(requestID) == nil
}

// ============= 特定用途的便捷函数 =============

// LoggerWithRequestID 为logger添加请求ID
func LoggerWithRequestID(requestID string) *logrus.Entry {
	return GetGlobalLogger().WithField("request_id", requestID)
}

// LoggerWithUserID 为logger添加用户ID
func LoggerWithUserID(userID string) *logrus.Entry {
	return GetGlobalLogger().WithField("user_id", userID)
}

// LoggerWithTraceID 为logger添加链路追踪ID
func LoggerWithTraceID(traceID string) *logrus.Entry {
	return GetGlobalLogger().WithField("trace_id", traceID)
}

// LoggerWithContext 为logger添加上下文信息
func LoggerWithContext(ctx map[string]any) *logrus.Entry {
	return GetGlobalLogger().WithFields(ctx)
}

// ============= 性能和监控相关函数 =============

// LogPerformance 记录性能日志
func LogPerformance(operation string, duration float64, args ...any) {
	GetGlobalLogger().WithFields(map[string]any{
		"operation": operation,
		"duration":  duration,
		"unit":      "ms",
	}).Info(args...)
}

// LogMetric 记录指标日志
func LogMetric(name string, value float64, labels map[string]string, args ...any) {
	fields := map[string]any{
		"metric_name":  name,
		"metric_value": value,
	}
	for k, v := range labels {
		fields["label_"+k] = v
	}
	GetGlobalLogger().WithFields(fields).Info(args...)
}

// LogEvent 记录事件日志
func LogEvent(eventType string, event map[string]any, args ...any) {
	fields := map[string]any{
		"event_type": eventType,
	}
	for k, v := range event {
		fields["event_"+k] = v
	}
	GetGlobalLogger().WithFields(fields).Info(args...)
}

// ============= HTTP和API相关函数 =============

// LogHTTPRequest 记录HTTP请求日志
func LogHTTPRequest(method, path, clientIP string, statusCode int, duration float64) {
	GetGlobalLogger().WithFields(map[string]any{
		"method":      method,
		"path":        path,
		"client_ip":   clientIP,
		"status_code": statusCode,
		"duration":    duration,
		"unit":        "ms",
	}).Info("HTTP Request")
}

// LogAPICall 记录API调用日志
func LogAPICall(api string, success bool, duration float64, errorMsg string) {
	fields := map[string]any{
		"api":      api,
		"success":  success,
		"duration": duration,
		"unit":     "ms",
	}
	if errorMsg != "" {
		fields["error"] = errorMsg
	}
	
	if success {
		GetGlobalLogger().WithFields(fields).Info("API Call Success")
	} else {
		GetGlobalLogger().WithFields(fields).Error("API Call Failed")
	}
}

// ============= 数据库相关函数 =============

// LogDBQuery 记录数据库查询日志
func LogDBQuery(query string, duration float64, rowsAffected int64, err error) {
	fields := map[string]any{
		"query":         query,
		"duration":      duration,
		"unit":          "ms",
		"rows_affected": rowsAffected,
	}
	
	if err != nil {
		fields["error"] = err.Error()
		GetGlobalLogger().WithFields(fields).Error("Database Query Failed")
	} else {
		GetGlobalLogger().WithFields(fields).Debug("Database Query Success")
	}
}

// LogDBConnection 记录数据库连接日志
func LogDBConnection(host string, database string, connected bool, err error) {
	fields := map[string]any{
		"host":      host,
		"database":  database,
		"connected": connected,
	}
	
	if err != nil {
		fields["error"] = err.Error()
		GetGlobalLogger().WithFields(fields).Error("Database Connection Failed")
	} else {
		GetGlobalLogger().WithFields(fields).Info("Database Connected")
	}
}

// ============= 安全相关函数 =============

// LogSecurityEvent 记录安全事件日志
func LogSecurityEvent(eventType string, userID string, ip string, details map[string]any) {
	fields := map[string]any{
		"security_event": eventType,
		"user_id":        userID,
		"client_ip":      ip,
	}
	for k, v := range details {
		fields[k] = v
	}
	GetGlobalLogger().WithFields(fields).Warn("Security Event")
}

// LogAuthEvent 记录认证事件日志
func LogAuthEvent(eventType string, userID string, success bool, reason string) {
	fields := map[string]any{
		"auth_event": eventType,
		"user_id":    userID,
		"success":    success,
	}
	if reason != "" {
		fields["reason"] = reason
	}
	
	if success {
		GetGlobalLogger().WithFields(fields).Info("Authentication Success")
	} else {
		GetGlobalLogger().WithFields(fields).Warn("Authentication Failed")
	}
}

// ============= 业务相关函数 =============

// LogBusinessEvent 记录业务事件日志
func LogBusinessEvent(eventType string, entityID string, entityType string, details map[string]any) {
	fields := map[string]any{
		"business_event": eventType,
		"entity_id":      entityID,
		"entity_type":    entityType,
	}
	for k, v := range details {
		fields[k] = v
	}
	GetGlobalLogger().WithFields(fields).Info("Business Event")
}

// LogTransaction 记录事务日志
func LogTransaction(txnID string, operation string, success bool, duration float64, err error) {
	fields := map[string]any{
		"transaction_id": txnID,
		"operation":      operation,
		"success":        success,
		"duration":       duration,
		"unit":           "ms",
	}
	
	if err != nil {
		fields["error"] = err.Error()
		GetGlobalLogger().WithFields(fields).Error("Transaction Failed")
	} else {
		GetGlobalLogger().WithFields(fields).Info("Transaction Success")
	}
}

// ============= 系统相关函数 =============

// LogSystemEvent 记录系统事件日志
func LogSystemEvent(eventType string, component string, details map[string]any) {
	fields := map[string]any{
		"system_event": eventType,
		"component":    component,
	}
	for k, v := range details {
		fields[k] = v
	}
	GetGlobalLogger().WithFields(fields).Info("System Event")
}

// LogStartup 记录启动日志
func LogStartup(service string, version string, port int, config map[string]any) {
	fields := map[string]any{
		"startup":   true,
		"service":   service,
		"version":   version,
		"port":      port,
	}
	for k, v := range config {
		fields["config_"+k] = v
	}
	GetGlobalLogger().WithFields(fields).Info("Service Started")
}

// LogShutdown 记录关闭日志
func LogShutdown(service string, reason string, duration float64) {
	GetGlobalLogger().WithFields(map[string]any{
		"shutdown": true,
		"service":  service,
		"reason":   reason,
		"duration": duration,
		"unit":     "ms",
	}).Info("Service Shutdown")
}

// ============= 配置动态更新函数 =============

// UpdateGlobalLogLevel 动态更新全局日志级别
func UpdateGlobalLogLevel(level LogLevel) {
	globalLogger := GetGlobalLogger()
	globalLogger.UpdateLevel(level)
	Info("日志级别已更新", "new_level", level)
}

// UpdateGlobalLogFormat 动态更新全局日志格式
func UpdateGlobalLogFormat(format LogFormat) {
	globalLogger := GetGlobalLogger()
	globalLogger.UpdateFormat(format)
	Info("日志格式已更新", "new_format", format)
}

// AddGlobalLogOutput 动态添加日志输出
func AddGlobalLogOutput(output string, outputConfig OutputConfig) {
	globalLogger := GetGlobalLogger()
	config := globalLogger.GetConfig()
	newConfig := config.AddOutput(output, outputConfig)
	globalLogger.UpdateConfig(newConfig)
	Info("日志输出已添加", "output", output)
}

// RemoveGlobalLogOutput 动态移除日志输出
func RemoveGlobalLogOutput(output string) {
	globalLogger := GetGlobalLogger()
	config := globalLogger.GetConfig()
	newConfig := config.RemoveOutput(output)
	globalLogger.UpdateConfig(newConfig)
	Info("日志输出已移除", "output", output)
}

// ============= 调试和开发相关函数 =============

// LogDebugInfo 记录调试信息（仅在Debug级别下输出）
func LogDebugInfo(component string, details map[string]any) {
	fields := map[string]any{
		"debug_info": true,
		"component":  component,
	}
	for k, v := range details {
		fields[k] = v
	}
	GetGlobalLogger().WithFields(fields).Debug("Debug Information")
}

// LogDump 转储变量内容（调试用）
func LogDump(name string, value any) {
	GetGlobalLogger().WithFields(map[string]any{
		"dump":     true,
		"var_name": name,
		"var_dump": value,
	}).Debug("Variable Dump")
}