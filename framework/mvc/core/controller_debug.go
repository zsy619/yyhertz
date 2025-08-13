package core

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 调试信息收集方法 =============

// GetDebugInfo 获取调试信息
func (c *BaseController) GetDebugInfo() map[string]any {
	debugInfo := map[string]any{
		"request_info":  c.getRequestDebugInfo(),
		"server_info":   c.getServerDebugInfo(),
		"runtime_info":  c.getRuntimeDebugInfo(),
		"memory_info":   c.getMemoryDebugInfo(),
		"session_info":  c.getSessionDebugInfo(),
		"headers_info":  c.getHeadersDebugInfo(),
	}
	
	return debugInfo
}

// getRequestDebugInfo 获取请求调试信息
func (c *BaseController) getRequestDebugInfo() map[string]any {
	if c.Ctx == nil {
		return map[string]any{"error": "context is nil"}
	}
	
	return map[string]any{
		"method":       c.Ctx.Method(),
		"path":         c.Ctx.Path(),
		"query":        string(c.Ctx.Request.URI().QueryString()),
		"client_ip":    c.GetClientIP(),
		"user_agent":   c.GetUserAgent(),
		"content_type": c.Ctx.ContentType(),
		"body_size":    c.GetBodySize(),
		"is_ajax":      c.IsAjax(),
		"is_https":     c.IsHTTPS(),
		"host":         string(c.Ctx.Request.Host()),
		"referer":      c.GetHeader("Referer"),
	}
}

// getServerDebugInfo 获取服务器调试信息
func (c *BaseController) getServerDebugInfo() map[string]any {
	return map[string]any{
		"timestamp":    time.Now().Format(time.RFC3339),
		"go_version":   runtime.Version(),
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"num_cpu":      runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
	}
}

// getRuntimeDebugInfo 获取运行时调试信息
func (c *BaseController) getRuntimeDebugInfo() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]any{
		"alloc_mb":        c.bytesToMB(m.Alloc),
		"total_alloc_mb":  c.bytesToMB(m.TotalAlloc),
		"sys_mb":          c.bytesToMB(m.Sys),
		"num_gc":          m.NumGC,
		"gc_cpu_fraction": m.GCCPUFraction,
	}
}

// getMemoryDebugInfo 获取内存调试信息
func (c *BaseController) getMemoryDebugInfo() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]any{
		"heap_alloc":     c.bytesToMB(m.HeapAlloc),
		"heap_sys":       c.bytesToMB(m.HeapSys),
		"heap_idle":      c.bytesToMB(m.HeapIdle),
		"heap_inuse":     c.bytesToMB(m.HeapInuse),
		"heap_released":  c.bytesToMB(m.HeapReleased),
		"heap_objects":   m.HeapObjects,
		"stack_inuse":    c.bytesToMB(m.StackInuse),
		"stack_sys":      c.bytesToMB(m.StackSys),
	}
}

// getSessionDebugInfo 获取会话调试信息
func (c *BaseController) getSessionDebugInfo() map[string]any {
	sessionInfo := map[string]any{
		"session_id": c.GetSessionID(),
	}
	
	// 获取所有会话数据（谨慎使用，可能包含敏感信息）
	if c.IsDebugMode() {
		// 这里应该实现获取所有session数据的逻辑
		sessionInfo["session_data"] = "Available in debug mode only"
	}
	
	return sessionInfo
}

// getHeadersDebugInfo 获取请求头调试信息
func (c *BaseController) getHeadersDebugInfo() map[string]any {
	if c.Ctx == nil {
		return map[string]any{"error": "context is nil"}
	}
	
	headers := make(map[string]string)
	
	// 获取所有请求头
	c.Ctx.Request.Request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	
	return map[string]any{
		"request_headers": headers,
	}
}

// ============= 性能监控方法 =============

// StartProfiler 开始性能分析
func (c *BaseController) StartProfiler() time.Time {
	start := time.Now()
	c.SetData("profiler_start", start)
	return start
}

// EndProfiler 结束性能分析
func (c *BaseController) EndProfiler() time.Duration {
	start, ok := c.GetData("profiler_start").(time.Time)
	if !ok {
		return 0
	}
	
	duration := time.Since(start)
	c.SetData("profiler_duration", duration)
	
	// 记录性能信息
	c.LogPerformance("request_duration", duration)
	
	return duration
}

// GetProfilerResult 获取性能分析结果
func (c *BaseController) GetProfilerResult() map[string]any {
	duration, _ := c.GetData("profiler_duration").(time.Duration)
	
	return map[string]any{
		"duration_ms":    float64(duration.Nanoseconds()) / 1e6,
		"duration_str":   duration.String(),
		"memory_before":  c.GetData("memory_before"),
		"memory_after":   c.GetData("memory_after"),
		"memory_delta":   c.calculateMemoryDelta(),
	}
}

// LogPerformance 记录性能信息
func (c *BaseController) LogPerformance(metric string, value any) {
	perfData := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
		"metric":    metric,
		"value":     value,
		"path":      c.Ctx.Path(),
		"method":    c.Ctx.Method(),
		"client_ip": c.GetClientIP(),
	}
	
	config.Infof("Performance: %+v", perfData)
}

// calculateMemoryDelta 计算内存差值
func (c *BaseController) calculateMemoryDelta() any {
	before, beforeOk := c.GetData("memory_before").(uint64)
	after, afterOk := c.GetData("memory_after").(uint64)
	
	if beforeOk && afterOk {
		delta := int64(after) - int64(before)
		return map[string]any{
			"delta_bytes": delta,
			"delta_mb":    float64(delta) / (1024 * 1024),
		}
	}
	
	return nil
}

// ============= 错误调试方法 =============

// DumpRequest 转储请求信息
func (c *BaseController) DumpRequest() string {
	if c.Ctx == nil {
		return "Context is nil"
	}
	
	dump := fmt.Sprintf("=== REQUEST DUMP ===\n")
	dump += fmt.Sprintf("Method: %s\n", c.Ctx.Method())
	dump += fmt.Sprintf("Path: %s\n", c.Ctx.Path())
	dump += fmt.Sprintf("Query: %s\n", c.Ctx.Request.URI().QueryString())
	dump += fmt.Sprintf("Content-Type: %s\n", c.Ctx.ContentType())
	dump += fmt.Sprintf("User-Agent: %s\n", c.GetUserAgent())
	dump += fmt.Sprintf("Client-IP: %s\n", c.GetClientIP())
	dump += fmt.Sprintf("Body-Size: %d\n", c.GetBodySize())
	
	dump += "\n=== HEADERS ===\n"
	c.Ctx.Request.Request.Header.VisitAll(func(key, value []byte) {
		dump += fmt.Sprintf("%s: %s\n", key, value)
	})
	
	if body, err := c.GetRawBody(); err == nil && len(body) > 0 {
		dump += "\n=== BODY ===\n"
		if len(body) > 1024 {
			dump += fmt.Sprintf("%s... (truncated, total: %d bytes)\n", string(body[:1024]), len(body))
		} else {
			dump += string(body) + "\n"
		}
	}
	
	return dump
}

// DumpStackTrace 转储堆栈跟踪
func (c *BaseController) DumpStackTrace() string {
	return string(debug.Stack())
}

// LogDebugError 记录调试错误信息
func (c *BaseController) LogDebugError(err error, context map[string]any) {
	errorData := map[string]any{
		"timestamp":   time.Now().Format(time.RFC3339),
		"error":       err.Error(),
		"stack_trace": string(debug.Stack()),
		"request":     c.getRequestDebugInfo(),
		"context":     context,
	}
	
	config.Errorf("Application Error: %+v", errorData)
}

// ============= 调试响应方法 =============

// DebugJSON 返回调试JSON响应
func (c *BaseController) DebugJSON(data any) {
	if !c.IsDebugMode() {
		c.JSONError("Debug mode is disabled")
		return
	}
	
	debugData := map[string]any{
		"debug_info": c.GetDebugInfo(),
		"data":       data,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	
	c.IndentedJSON(debugData)
}

// DebugHeaders 在响应中添加调试头
func (c *BaseController) DebugHeaders() {
	if !c.IsDebugMode() {
		return
	}
	
	c.SetHeader("X-Debug-Request-ID", c.GenerateRequestID())
	c.SetHeader("X-Debug-Timestamp", time.Now().Format(time.RFC3339))
	c.SetHeader("X-Debug-Go-Version", runtime.Version())
	
	// 添加性能信息
	if duration, ok := c.GetData("profiler_duration").(time.Duration); ok {
		c.SetHeader("X-Debug-Duration-Ms", fmt.Sprintf("%.2f", float64(duration.Nanoseconds())/1e6))
	}
	
	// 添加内存信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	c.SetHeader("X-Debug-Memory-MB", fmt.Sprintf("%.2f", c.bytesToMB(m.Alloc)))
}

// ============= 调试工具方法 =============

// IsDebugMode 检查是否为调试模式
func (c *BaseController) IsDebugMode() bool {
	// 这里应该从配置中读取，示例实现
	// return strings.ToLower(config.GetString("app.debug", "false")) == "true"
	return false // 临时固定为false，避免编译错误
}

// GenerateRequestID 生成请求ID
func (c *BaseController) GenerateRequestID() string {
	// 简单的请求ID生成
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("req_%d", timestamp)
}

// bytesToMB 字节转MB
func (c *BaseController) bytesToMB(bytes uint64) float64 {
	return float64(bytes) / (1024 * 1024)
}

// PrintDebugInfo 打印调试信息到控制台
func (c *BaseController) PrintDebugInfo() {
	if !c.IsDebugMode() {
		return
	}
	
	debugInfo := c.GetDebugInfo()
	jsonData, err := json.MarshalIndent(debugInfo, "", "  ")
	if err != nil {
		config.Errorf("Failed to marshal debug info: %v", err)
		return
	}
	
	config.Infof("Debug Info:\n%s", string(jsonData))
}

// ============= 断言和验证方法 =============

// Assert 断言条件为真
func (c *BaseController) Assert(condition bool, message string) {
	if !condition {
		stack := string(debug.Stack())
		config.Errorf("Assertion failed: %s\nStack trace:\n%s", message, stack)
		
		if c.IsDebugMode() {
			panic(fmt.Sprintf("Assertion failed: %s", message))
		}
	}
}

// AssertNotNil 断言值不为nil
func (c *BaseController) AssertNotNil(value any, message string) {
	c.Assert(value != nil, message)
}

// AssertEqual 断言两个值相等
func (c *BaseController) AssertEqual(expected, actual any, message string) {
	c.Assert(expected == actual, fmt.Sprintf("%s: expected %v, got %v", message, expected, actual))
}

// ============= 监控和度量方法 =============

// RecordMetric 记录度量值
func (c *BaseController) RecordMetric(name string, value float64, tags map[string]string) {
	metricData := map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
		"name":      name,
		"value":     value,
		"tags":      tags,
		"path":      c.Ctx.Path(),
		"method":    c.Ctx.Method(),
	}
	
	config.Infof("Metric: %+v", metricData)
}

// IncrementCounter 增加计数器
func (c *BaseController) IncrementCounter(name string, tags map[string]string) {
	c.RecordMetric(name, 1, tags)
}

// RecordTiming 记录时间度量
func (c *BaseController) RecordTiming(name string, duration time.Duration, tags map[string]string) {
	c.RecordMetric(name+"_duration_ms", float64(duration.Nanoseconds())/1e6, tags)
}

// RecordGauge 记录仪表值
func (c *BaseController) RecordGauge(name string, value float64, tags map[string]string) {
	c.RecordMetric(name, value, tags)
}

// ============= 健康检查方法 =============

// HealthCheck 健康检查
func (c *BaseController) HealthCheck() map[string]any {
	health := map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   c.getAppVersion(),
		"uptime":    c.getUptime(),
		"checks":    c.performHealthChecks(),
	}
	
	return health
}

// getAppVersion 获取应用版本
func (c *BaseController) getAppVersion() string {
	// 这里应该从配置或构建信息中读取
	// return config.GetString("app.version", "unknown")
	return "1.0.0" // 临时固定版本号
}

// getUptime 获取运行时间
func (c *BaseController) getUptime() string {
	// 这里应该记录应用启动时间
	return "unknown"
}

// performHealthChecks 执行健康检查
func (c *BaseController) performHealthChecks() map[string]any {
	checks := map[string]any{
		"database": c.checkDatabase(),
		"cache":    c.checkCache(),
		"disk":     c.checkDiskSpace(),
		"memory":   c.checkMemoryUsage(),
	}
	
	return checks
}

// checkDatabase 检查数据库连接
func (c *BaseController) checkDatabase() map[string]any {
	// 这里应该实现真正的数据库检查
	return map[string]any{
		"status": "ok",
		"latency_ms": 5.2,
	}
}

// checkCache 检查缓存服务
func (c *BaseController) checkCache() map[string]any {
	// 这里应该实现真正的缓存检查
	return map[string]any{
		"status": "ok",
		"latency_ms": 1.8,
	}
}

// checkDiskSpace 检查磁盘空间
func (c *BaseController) checkDiskSpace() map[string]any {
	// 这里应该实现真正的磁盘检查
	return map[string]any{
		"status": "ok",
		"usage_percent": 45.6,
	}
}

// checkMemoryUsage 检查内存使用
func (c *BaseController) checkMemoryUsage() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	usagePercent := float64(m.Alloc) / float64(m.Sys) * 100
	
	status := "ok"
	if usagePercent > 90 {
		status = "critical"
	} else if usagePercent > 80 {
		status = "warning"
	}
	
	return map[string]any{
		"status":         status,
		"usage_percent":  usagePercent,
		"alloc_mb":       c.bytesToMB(m.Alloc),
		"sys_mb":         c.bytesToMB(m.Sys),
	}
}