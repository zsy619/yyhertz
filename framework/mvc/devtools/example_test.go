package devtools

import (
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
)

// TestBasicSetup 测试基础设置
func TestBasicSetup(t *testing.T) {
	app := mvc.NewApp()

	err := SetupBasicDevTools(app)
	if err != nil {
		t.Fatalf("SetupBasicDevTools failed: %v", err)
	}

	t.Log("Basic devtools setup successful")
}

// TestFullSetup 测试完整设置
func TestFullSetup(t *testing.T) {
	app := mvc.NewApp()

	config := &DevToolsConfig{
		Enabled:           true,
		Environment:       "development",
		EnableDebug:       true,
		EnableHealthCheck: true,
		EnableQPS:         true,
		EnableMetrics:     true, // 包含性能监控功能
		EnableProfiler:    true,
		EnableRateLimit:   false, // 测试中不启用限流
		EnableHotReload:   false, // 测试中不启用热重载
		ExternalServices:  []string{},
		RateLimitRules:    []*RateLimit{},
		WhitelistIPs:      []string{"127.0.0.1"},
	}

	err := SetupDevTools(app, config)
	if err != nil {
		t.Fatalf("SetupDevTools failed: %v", err)
	}

	t.Log("Full devtools setup successful")
}

// TestHealthCheckers 测试健康检查器
func TestHealthCheckers(t *testing.T) {
	// 测试磁盘检查器
	diskChecker := CreateDiskChecker("test_disk", "/")
	if diskChecker.Name() != "test_disk" {
		t.Errorf("Expected name 'test_disk', got '%s'", diskChecker.Name())
	}

	// 测试自定义检查器
	customChecker := CreateFileExistsChecker("test_file", "/etc/hosts")
	if customChecker.Name() != "test_file" {
		t.Errorf("Expected name 'test_file', got '%s'", customChecker.Name())
	}

	t.Log("Health checkers test successful")
}

// TestQPSMonitor 测试QPS监控
func TestQPSMonitor(t *testing.T) {
	config := &QPSConfig{
		Enabled:             true,
		WindowSize:          10 * time.Second,
		MaxHistorySize:      100,
		QPSThreshold:        1000,
		ConcurrentThreshold: 1000,
	}

	monitor := NewQPSMonitor(config)
	if !monitor.IsEnabled() {
		t.Error("QPS monitor should be enabled")
	}

	stats := monitor.GetStats()
	if stats.CurrentQPS < 0 {
		t.Error("Current QPS should not be negative")
	}

	t.Log("QPS monitor test successful")
}

// TestMetricsCollector 测试指标收集
func TestMetricsCollector(t *testing.T) {
	config := &MetricsConfig{
		Enabled:       true,
		Namespace:     "test",
		Subsystem:     "unit",
		IncludeSystem: false,
	}

	collector := NewMetricsCollector(config)
	if !collector.IsEnabled() {
		t.Error("Metrics collector should be enabled")
	}

	// 创建自定义指标
	counter := NewCounter("test_counter", "Test counter", map[string]string{"type": "test"})
	collector.RegisterMetric(counter)

	counter.Inc()
	if counter.Value() != 1 {
		t.Errorf("Expected counter value 1, got %f", counter.Value())
	}

	t.Log("Metrics collector test successful")
}

// TestProfiler 测试性能分析器
func TestProfiler(t *testing.T) {
	config := &ProfilerConfig{
		Enabled:         true,
		MaxProfiles:     5,
		AutoProfile:     false,
		ProfileInterval: 1 * time.Minute,
	}

	profiler := NewProfiler(config)
	if !profiler.IsEnabled() {
		t.Error("Profiler should be enabled")
	}

	// 测试获取基线
	baseline := profiler.CreateBaseline()
	if baseline == nil {
		t.Error("Baseline should not be nil")
	}

	if baseline.Timestamp.IsZero() {
		t.Error("Baseline timestamp should not be zero")
	}

	t.Log("Profiler test successful")
}

// TestRateLimiter 测试限流器
func TestRateLimiter(t *testing.T) {
	config := &RateLimiterConfig{
		Enabled: true,
		Rules: []*RateLimit{
			{
				Rate:      10,
				Burst:     20,
				Strategy:  LimitStrategyTokenBucket,
				Dimension: LimitDimensionGlobal,
				Enabled:   true,
			},
		},
		Whitelist: []string{"127.0.0.1"},
	}

	limiter := NewRateLimiter(config)
	if !limiter.IsEnabled() {
		t.Error("Rate limiter should be enabled")
	}

	rules := limiter.GetRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	whitelist := limiter.GetWhitelist()
	if len(whitelist) != 1 || whitelist[0] != "127.0.0.1" {
		t.Errorf("Expected whitelist ['127.0.0.1'], got %v", whitelist)
	}

	t.Log("Rate limiter test successful")
}

// TestTokenBucket 测试令牌桶
func TestTokenBucket(t *testing.T) {
	bucket := NewTokenBucket(10, 20) // 10 tokens/sec, capacity 20

	// 初始应该有满桶的令牌
	if bucket.Remaining() != 20 {
		t.Errorf("Expected 20 tokens, got %d", bucket.Remaining())
	}

	// 消费一个令牌
	if !bucket.Allow() {
		t.Error("Should allow first request")
	}

	if bucket.Remaining() != 19 {
		t.Errorf("Expected 19 tokens after consumption, got %d", bucket.Remaining())
	}

	t.Log("Token bucket test successful")
}

// TestSlidingWindow 测试滑动窗口
func TestSlidingWindow(t *testing.T) {
	window := NewSlidingWindow(5, 1*time.Second) // 5 requests per second

	// 应该允许前5个请求
	for i := 0; i < 5; i++ {
		if !window.Allow() {
			t.Errorf("Should allow request %d", i+1)
		}
	}

	// 第6个请求应该被拒绝
	if window.Allow() {
		t.Error("Should reject 6th request")
	}

	remaining := window.Remaining()
	if remaining != 0 {
		t.Errorf("Expected 0 remaining requests, got %d", remaining)
	}

	t.Log("Sliding window test successful")
}
