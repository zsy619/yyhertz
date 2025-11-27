// Package gin - 生产环境支持功能
// 提供限流、监控、健康检查、性能分析等生产环境必需功能
package gin

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// =============================================================================
// 限流器实现
// =============================================================================

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(key string) bool
	Reset(key string)
	GetStats(key string) RateLimitStats
}

// RateLimitStats 限流统计
type RateLimitStats struct {
	Requests     int64     `json:"requests"`
	Allowed      int64     `json:"allowed"`
	Rejected     int64     `json:"rejected"`
	LastRequest  time.Time `json:"last_request"`
	ResetTime    time.Time `json:"reset_time"`
	Remaining    int64     `json:"remaining"`
}

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*tokenBucket
	rate     int64  // 每秒生成的令牌数
	capacity int64  // 桶容量
	cleanup  time.Duration
}

type tokenBucket struct {
	tokens     int64
	lastUpdate time.Time
	stats      RateLimitStats
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(rate, capacity int64) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
		cleanup:  5 * time.Minute,
	}
	
	// 启动清理协程
	go limiter.cleanupLoop()
	
	return limiter
}

// Allow 检查是否允许请求
func (l *TokenBucketLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     l.capacity,
			lastUpdate: time.Now(),
		}
		l.buckets[key] = bucket
	}
	
	// 更新令牌
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate)
	tokensToAdd := int64(elapsed.Seconds()) * l.rate
	
	bucket.tokens += tokensToAdd
	if bucket.tokens > l.capacity {
		bucket.tokens = l.capacity
	}
	bucket.lastUpdate = now
	
	// 更新统计
	bucket.stats.Requests++
	bucket.stats.LastRequest = now
	bucket.stats.Remaining = bucket.tokens
	
	// 检查是否有可用令牌
	if bucket.tokens > 0 {
		bucket.tokens--
		bucket.stats.Allowed++
		return true
	}
	
	bucket.stats.Rejected++
	return false
}

// Reset 重置限流器
func (l *TokenBucketLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// GetStats 获取统计信息
func (l *TokenBucketLimiter) GetStats(key string) RateLimitStats {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	if bucket, exists := l.buckets[key]; exists {
		return bucket.stats
	}
	return RateLimitStats{}
}

// cleanupLoop 清理过期的桶
func (l *TokenBucketLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()
	
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, bucket := range l.buckets {
			if now.Sub(bucket.lastUpdate) > l.cleanup {
				delete(l.buckets, key)
			}
		}
		l.mu.Unlock()
	}
}

// =============================================================================
// 限流中间件
// =============================================================================

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Skipper      func(*Context) bool
	Limiter      RateLimiter
	KeyGenerator func(*Context) string
	OnLimitReached func(*Context)
}

// DefaultRateLimitConfig 默认限流配置
var DefaultRateLimitConfig = RateLimitConfig{
	Skipper: func(*Context) bool { return false },
	Limiter: NewTokenBucketLimiter(100, 1000), // 每秒100个令牌，容量1000
	KeyGenerator: func(c *Context) string {
		return c.ClientIP()
	},
	OnLimitReached: func(c *Context) {
		c.JSON(http.StatusTooManyRequests, H{
			"error": "Rate limit exceeded",
			"code":  "RATE_LIMIT_EXCEEDED",
		})
		c.Abort()
	},
}

// RateLimitWithConfig 带配置的限流中间件
func RateLimitWithConfig(config RateLimitConfig) HandlerFunc {
	if config.Limiter == nil {
		config.Limiter = DefaultRateLimitConfig.Limiter
	}
	if config.KeyGenerator == nil {
		config.KeyGenerator = DefaultRateLimitConfig.KeyGenerator
	}
	if config.OnLimitReached == nil {
		config.OnLimitReached = DefaultRateLimitConfig.OnLimitReached
	}
	
	return func(c *Context) {
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}
		
		key := config.KeyGenerator(c)
		if !config.Limiter.Allow(key) {
			config.OnLimitReached(c)
			return
		}
		
		c.Next()
	}
}

// RateLimit 默认限流中间件
func RateLimit() HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig)
}

// =============================================================================
// 健康检查
// =============================================================================

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Name() string
	Check() HealthStatus
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status  string            `json:"status"`
	Message string            `json:"message,omitempty"`
	Details map[string]any    `json:"details,omitempty"`
}

// HealthCheckManager 健康检查管理器
type HealthCheckManager struct {
	checkers map[string]HealthChecker
	mu       sync.RWMutex
}

// NewHealthCheckManager 创建健康检查管理器
func NewHealthCheckManager() *HealthCheckManager {
	return &HealthCheckManager{
		checkers: make(map[string]HealthChecker),
	}
}

// AddChecker 添加检查器
func (hm *HealthCheckManager) AddChecker(name string, checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checkers[name] = checker
}

// RemoveChecker 移除检查器
func (hm *HealthCheckManager) RemoveChecker(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.checkers, name)
}

// CheckAll 执行所有健康检查
func (hm *HealthCheckManager) CheckAll() map[string]HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	
	results := make(map[string]HealthStatus)
	for name, checker := range hm.checkers {
		results[name] = checker.Check()
	}
	return results
}

// GetOverallStatus 获取整体健康状态
func (hm *HealthCheckManager) GetOverallStatus() HealthStatus {
	results := hm.CheckAll()
	
	overallStatus := "healthy"
	var failedChecks []string
	
	for name, status := range results {
		if status.Status != "healthy" {
			overallStatus = "unhealthy"
			failedChecks = append(failedChecks, name)
		}
	}
	
	status := HealthStatus{
		Status: overallStatus,
		Details: map[string]any{
			"checks": results,
		},
	}
	
	if len(failedChecks) > 0 {
		status.Message = fmt.Sprintf("Failed checks: %v", failedChecks)
	}
	
	return status
}

// 全局健康检查管理器
var globalHealthManager = NewHealthCheckManager()

// =============================================================================
// 内置健康检查器
// =============================================================================

// MemoryHealthChecker 内存健康检查器
type MemoryHealthChecker struct {
	MaxMemoryMB int64
}

// Name 返回检查器名称
func (c *MemoryHealthChecker) Name() string {
	return "memory"
}

// Check 执行检查
func (c *MemoryHealthChecker) Check() HealthStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	currentMB := int64(m.Alloc / 1024 / 1024)
	
	status := HealthStatus{
		Status: "healthy",
		Details: map[string]any{
			"allocated_mb": currentMB,
			"max_mb":       c.MaxMemoryMB,
			"gc_cycles":    m.NumGC,
		},
	}
	
	if c.MaxMemoryMB > 0 && currentMB > c.MaxMemoryMB {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Memory usage %dMB exceeds limit %dMB", 
			currentMB, c.MaxMemoryMB)
	}
	
	return status
}

// GoroutineHealthChecker Goroutine健康检查器
type GoroutineHealthChecker struct {
	MaxGoroutines int
}

// Name 返回检查器名称
func (c *GoroutineHealthChecker) Name() string {
	return "goroutines"
}

// Check 执行检查
func (c *GoroutineHealthChecker) Check() HealthStatus {
	current := runtime.NumGoroutine()
	
	status := HealthStatus{
		Status: "healthy",
		Details: map[string]any{
			"current": current,
			"max":     c.MaxGoroutines,
		},
	}
	
	if c.MaxGoroutines > 0 && current > c.MaxGoroutines {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Goroutine count %d exceeds limit %d", 
			current, c.MaxGoroutines)
	}
	
	return status
}

// =============================================================================
// 健康检查中间件和端点
// =============================================================================

// HealthCheckMiddleware 健康检查中间件
func HealthCheckMiddleware() HandlerFunc {
	return func(c *Context) {
		if c.Request().URL.Path == "/health" {
			status := globalHealthManager.GetOverallStatus()
			
			httpStatus := http.StatusOK
			if status.Status != "healthy" {
				httpStatus = http.StatusServiceUnavailable
			}
			
			c.JSON(httpStatus, status)
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RegisterHealthCheckers 注册默认健康检查器
func RegisterHealthCheckers() {
	globalHealthManager.AddChecker("memory", &MemoryHealthChecker{MaxMemoryMB: 1024})
	globalHealthManager.AddChecker("goroutines", &GoroutineHealthChecker{MaxGoroutines: 1000})
}

// =============================================================================
// 监控和指标
// =============================================================================

// Metrics 监控指标
type Metrics struct {
	mu sync.RWMutex
	
	// 请求指标
	TotalRequests    int64            `json:"total_requests"`
	RequestsPerSec   float64          `json:"requests_per_sec"`
	AvgResponseTime  time.Duration    `json:"avg_response_time"`
	ErrorRate        float64          `json:"error_rate"`
	
	// 状态码统计
	StatusCodes      map[int]int64    `json:"status_codes"`
	
	// 路径统计
	PathStats        map[string]*PathMetrics `json:"path_stats"`
	
	// 系统指标
	MemoryUsage      int64            `json:"memory_usage_mb"`
	GoroutineCount   int              `json:"goroutine_count"`
	
	// 时间窗口
	StartTime        time.Time        `json:"start_time"`
	LastReset        time.Time        `json:"last_reset"`
}

// PathMetrics 路径指标
type PathMetrics struct {
	Count       int64         `json:"count"`
	TotalTime   time.Duration `json:"total_time"`
	AvgTime     time.Duration `json:"avg_time"`
	ErrorCount  int64         `json:"error_count"`
	ErrorRate   float64       `json:"error_rate"`
}

// NewMetrics 创建监控指标
func NewMetrics() *Metrics {
	return &Metrics{
		StatusCodes: make(map[int]int64),
		PathStats:   make(map[string]*PathMetrics),
		StartTime:   time.Now(),
		LastReset:   time.Now(),
	}
}

// RecordRequest 记录请求
func (m *Metrics) RecordRequest(path string, statusCode int, responseTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 更新总请求数
	m.TotalRequests++
	
	// 更新状态码统计
	m.StatusCodes[statusCode]++
	
	// 更新路径统计
	pathMetrics, exists := m.PathStats[path]
	if !exists {
		pathMetrics = &PathMetrics{}
		m.PathStats[path] = pathMetrics
	}
	
	pathMetrics.Count++
	pathMetrics.TotalTime += responseTime
	pathMetrics.AvgTime = pathMetrics.TotalTime / time.Duration(pathMetrics.Count)
	
	if statusCode >= 400 {
		pathMetrics.ErrorCount++
	}
	pathMetrics.ErrorRate = float64(pathMetrics.ErrorCount) / float64(pathMetrics.Count) * 100
	
	// 更新平均响应时间和错误率
	m.updateAggregates()
}

// updateAggregates 更新聚合指标
func (m *Metrics) updateAggregates() {
	if m.TotalRequests == 0 {
		return
	}
	
	// 计算错误率
	var errorCount int64
	for statusCode, count := range m.StatusCodes {
		if statusCode >= 400 {
			errorCount += count
		}
	}
	m.ErrorRate = float64(errorCount) / float64(m.TotalRequests) * 100
	
	// 计算平均响应时间
	var totalTime time.Duration
	for _, pathMetrics := range m.PathStats {
		totalTime += pathMetrics.TotalTime
	}
	m.AvgResponseTime = totalTime / time.Duration(m.TotalRequests)
	
	// 计算每秒请求数
	elapsed := time.Since(m.LastReset)
	if elapsed.Seconds() > 0 {
		m.RequestsPerSec = float64(m.TotalRequests) / elapsed.Seconds()
	}
	
	// 更新系统指标
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	m.MemoryUsage = int64(mem.Alloc / 1024 / 1024)
	m.GoroutineCount = runtime.NumGoroutine()
}

// GetSnapshot 获取指标快照
func (m *Metrics) GetSnapshot() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 更新聚合指标
	m.updateAggregates()
	
	// 创建快照
	snapshot := &Metrics{
		TotalRequests:   m.TotalRequests,
		RequestsPerSec:  m.RequestsPerSec,
		AvgResponseTime: m.AvgResponseTime,
		ErrorRate:       m.ErrorRate,
		MemoryUsage:     m.MemoryUsage,
		GoroutineCount:  m.GoroutineCount,
		StartTime:       m.StartTime,
		LastReset:       m.LastReset,
		StatusCodes:     make(map[int]int64),
		PathStats:       make(map[string]*PathMetrics),
	}
	
	// 复制映射
	for k, v := range m.StatusCodes {
		snapshot.StatusCodes[k] = v
	}
	
	for k, v := range m.PathStats {
		snapshot.PathStats[k] = &PathMetrics{
			Count:      v.Count,
			TotalTime:  v.TotalTime,
			AvgTime:    v.AvgTime,
			ErrorCount: v.ErrorCount,
			ErrorRate:  v.ErrorRate,
		}
	}
	
	return snapshot
}

// Reset 重置指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.TotalRequests = 0
	m.RequestsPerSec = 0
	m.AvgResponseTime = 0
	m.ErrorRate = 0
	m.StatusCodes = make(map[int]int64)
	m.PathStats = make(map[string]*PathMetrics)
	m.LastReset = time.Now()
}

// 全局指标
var globalMetrics = NewMetrics()

// =============================================================================
// 监控中间件
// =============================================================================

// MetricsMiddleware 监控中间件
func MetricsMiddleware() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		
		c.Next()
		
		// 记录指标
		responseTime := time.Since(start)
		path := c.Request().URL.Path
		statusCode := c.Writer().(*responseWriterWrapper).RequestContext.Response.StatusCode()
		
		globalMetrics.RecordRequest(path, statusCode, responseTime)
	}
}

// MetricsHandler 监控指标端点
func MetricsHandler() HandlerFunc {
	return func(c *Context) {
		metrics := globalMetrics.GetSnapshot()
		c.JSON(http.StatusOK, metrics)
	}
}

// =============================================================================
// 性能分析
// =============================================================================

// Profiler 性能分析器
type Profiler struct {
	mu       sync.RWMutex
	enabled  bool
	samples  []ProfileData
	maxSamples int
}

// ProfileData 性能分析数据
type ProfileData struct {
	Timestamp    time.Time     `json:"timestamp"`
	Path         string        `json:"path"`
	Method       string        `json:"method"`
	Duration     time.Duration `json:"duration"`
	StatusCode   int           `json:"status_code"`
	MemoryAlloc  uint64        `json:"memory_alloc"`
	GoroutineNum int           `json:"goroutine_num"`
	UserAgent    string        `json:"user_agent"`
	IP           string        `json:"ip"`
}

// NewProfiler 创建性能分析器
func NewProfiler(maxSamples int) *Profiler {
	return &Profiler{
		maxSamples: maxSamples,
		samples:    make([]ProfileData, 0, maxSamples),
	}
}

// Enable 启用性能分析
func (p *Profiler) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

// Disable 禁用性能分析
func (p *Profiler) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

// IsEnabled 检查是否启用
func (p *Profiler) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// AddSample 添加性能样本
func (p *Profiler) AddSample(data ProfileData) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.enabled {
		return
	}
	
	// 如果达到最大样本数，移除最旧的
	if len(p.samples) >= p.maxSamples {
		p.samples = p.samples[1:]
	}
	
	p.samples = append(p.samples, data)
}

// GetSamples 获取性能样本
func (p *Profiler) GetSamples() []ProfileData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	result := make([]ProfileData, len(p.samples))
	copy(result, p.samples)
	return result
}

// Clear 清除样本
func (p *Profiler) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.samples = p.samples[:0]
}

// GetStats 获取统计信息
func (p *Profiler) GetStats() map[string]any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if len(p.samples) == 0 {
		return map[string]any{
			"sample_count": 0,
			"enabled":      p.enabled,
		}
	}
	
	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration time.Duration = time.Hour // 初始化为一个大值
	
	pathCounts := make(map[string]int)
	statusCounts := make(map[int]int)
	
	for _, sample := range p.samples {
		totalDuration += sample.Duration
		
		if sample.Duration > maxDuration {
			maxDuration = sample.Duration
		}
		if sample.Duration < minDuration {
			minDuration = sample.Duration
		}
		
		pathCounts[sample.Path]++
		statusCounts[sample.StatusCode]++
	}
	
	avgDuration := totalDuration / time.Duration(len(p.samples))
	
	return map[string]any{
		"sample_count":   len(p.samples),
		"enabled":        p.enabled,
		"avg_duration":   avgDuration.String(),
		"max_duration":   maxDuration.String(),
		"min_duration":   minDuration.String(),
		"path_counts":    pathCounts,
		"status_counts":  statusCounts,
	}
}

// 全局生产环境性能分析器
var globalProductionProfiler = NewProfiler(1000)

// =============================================================================
// 性能分析中间件
// =============================================================================

// ProfilerMiddleware 性能分析中间件
func ProfilerMiddleware() HandlerFunc {
	return func(c *Context) {
		if !globalProductionProfiler.IsEnabled() {
			c.Next()
			return
		}
		
		start := time.Now()
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		c.Next()
		
		runtime.ReadMemStats(&m2)
		duration := time.Since(start)
		
		// 收集性能数据
		data := ProfileData{
			Timestamp:    start,
			Path:         c.Request().URL.Path,
			Method:       c.Request().Method,
			Duration:     duration,
			StatusCode:   c.Writer().(*responseWriterWrapper).RequestContext.Response.StatusCode(),
			MemoryAlloc:  m2.Alloc - m1.Alloc,
			GoroutineNum: runtime.NumGoroutine(),
			UserAgent:    c.GetHeader("User-Agent"),
			IP:           c.ClientIP(),
		}
		
		globalProductionProfiler.AddSample(data)
	}
}

// =============================================================================
// 生产环境中间件组合
// =============================================================================

// ProductionMiddleware 生产环境中间件组合
func ProductionMiddleware() []HandlerFunc {
	return []HandlerFunc{
		Recovery(),                    // 错误恢复
		Logger(),                     // 请求日志
		RateLimit(),                  // 限流
		MetricsMiddleware(),          // 监控指标
		HealthCheckMiddleware(),      // 健康检查
		ProfilerMiddleware(),         // 性能分析
	}
}

// SetupProductionRoutes 设置生产环境路由
func SetupProductionRoutes(r *Engine) {
	// 健康检查端点
	r.GET("/health", func(c *Context) {
		status := globalHealthManager.GetOverallStatus()
		httpStatus := http.StatusOK
		if status.Status != "healthy" {
			httpStatus = http.StatusServiceUnavailable
		}
		c.JSON(httpStatus, status)
	})
	
	// 指标端点
	r.GET("/metrics", MetricsHandler())
	
	// 性能分析端点
	r.GET("/debug/profile", func(c *Context) {
		c.JSON(http.StatusOK, H{
			"stats":   globalProductionProfiler.GetStats(),
			"samples": globalProductionProfiler.GetSamples(),
		})
	})
	
	// 性能分析控制端点
	r.POST("/debug/profile/enable", func(c *Context) {
		globalProductionProfiler.Enable()
		c.JSON(http.StatusOK, H{"message": "Profiler enabled"})
	})
	
	r.POST("/debug/profile/disable", func(c *Context) {
		globalProductionProfiler.Disable()
		c.JSON(http.StatusOK, H{"message": "Profiler disabled"})
	})
	
	r.POST("/debug/profile/clear", func(c *Context) {
		globalProductionProfiler.Clear()
		c.JSON(http.StatusOK, H{"message": "Profile data cleared"})
	})
	
	// 运行时信息端点
	r.GET("/debug/runtime", func(c *Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		c.JSON(http.StatusOK, H{
			"goroutines":    runtime.NumGoroutine(),
			"memory_alloc":  m.Alloc,
			"memory_sys":    m.Sys,
			"gc_cycles":     m.NumGC,
			"cpu_count":     runtime.NumCPU(),
			"go_version":    runtime.Version(),
		})
	})
}

// =============================================================================
// 生产环境配置助手
// =============================================================================

// ProductionConfig 生产环境配置
type ProductionConfig struct {
	EnableMetrics    bool
	EnableProfiling  bool
	EnableHealthCheck bool
	EnableRateLimit  bool
	RateLimitRate    int64
	RateLimitCapacity int64
	MetricsPath      string
	HealthPath       string
	ProfilePath      string
}

// DefaultProductionConfig 默认生产环境配置
var DefaultProductionConfig = ProductionConfig{
	EnableMetrics:     true,
	EnableProfiling:   false, // 默认关闭，性能开销较大
	EnableHealthCheck: true,
	EnableRateLimit:   true,
	RateLimitRate:     100,
	RateLimitCapacity: 1000,
	MetricsPath:       "/metrics",
	HealthPath:        "/health",
	ProfilePath:       "/debug/profile",
}

// SetupProductionWithConfig 使用配置设置生产环境
func SetupProductionWithConfig(r *Engine, config ProductionConfig) {
	// 注册健康检查器
	if config.EnableHealthCheck {
		RegisterHealthCheckers()
	}
	
	// 添加中间件
	middlewares := []HandlerFunc{
		// Recovery(), // 使用中间件包中的Recovery
		// Logger(),   // 使用中间件包中的Logger
	}
	
	if config.EnableRateLimit {
		limiter := NewTokenBucketLimiter(config.RateLimitRate, config.RateLimitCapacity)
		rateLimitConfig := DefaultRateLimitConfig
		rateLimitConfig.Limiter = limiter
		middlewares = append(middlewares, RateLimitWithConfig(rateLimitConfig))
	}
	
	if config.EnableMetrics {
		middlewares = append(middlewares, MetricsMiddleware())
	}
	
	if config.EnableHealthCheck {
		middlewares = append(middlewares, HealthCheckMiddleware())
	}
	
	if config.EnableProfiling {
		globalProductionProfiler.Enable()
		middlewares = append(middlewares, ProfilerMiddleware())
	}
	
	// 应用中间件
	r.Use(middlewares...)
	
	// 设置端点
	if config.EnableHealthCheck {
		r.GET(config.HealthPath, func(c *Context) {
			status := globalHealthManager.GetOverallStatus()
			httpStatus := http.StatusOK
			if status.Status != "healthy" {
				httpStatus = http.StatusServiceUnavailable
			}
			c.JSON(httpStatus, status)
		})
	}
	
	if config.EnableMetrics {
		r.GET(config.MetricsPath, MetricsHandler())
	}
	
	if config.EnableProfiling {
		r.GET(config.ProfilePath, func(c *Context) {
			c.JSON(http.StatusOK, H{
				"stats":   globalProductionProfiler.GetStats(),
				"samples": globalProductionProfiler.GetSamples(),
			})
		})
	}
}

// GetGlobalMetrics 获取全局指标
func GetGlobalMetrics() *Metrics {
	return globalMetrics.GetSnapshot()
}

// GetGlobalProfileStats 获取全局性能分析统计
func GetGlobalProfileStats() map[string]any {
	return globalProductionProfiler.GetStats()
}

// EnableGlobalProfiling 启用全局性能分析
func EnableGlobalProfiling() {
	globalProductionProfiler.Enable()
}

// DisableGlobalProfiling 禁用全局性能分析
func DisableGlobalProfiling() {
	globalProductionProfiler.Disable()
}