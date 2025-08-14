package devtools

import (
	"database/sql"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc"
	_ "github.com/mattn/go-sqlite3"
)

// TestNewPanelsConfiguration 测试新面板配置
func TestNewPanelsConfiguration(t *testing.T) {
	app := mvc.NewApp()
	
	// 创建临时SQLite数据库用于测试
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	config := &DevToolsConfig{
		Enabled:           true,
		Environment:       "development",
		EnableDebug:       true,
		EnableHealthCheck: true,
		EnableQPS:         true,
		EnableMetrics:     true,
		EnableProfiler:    true,
		EnableRateLimit:   true,
		EnableHotReload:   false, // 测试中不启用热重载
		EnableCache:       true,  // 启用缓存监控
		EnableDatabase:    true,  // 启用数据库监控
		EnableSecurity:    true,  // 启用安全监控
		Database:          db,
		ExternalServices:  []string{},
		RateLimitRules:    []*RateLimit{},
		WhitelistIPs:      []string{"127.0.0.1"},
	}

	err = SetupDevTools(app, config)
	if err != nil {
		t.Fatalf("SetupDevTools failed: %v", err)
	}

	t.Log("New panels setup successful")
}

// TestCacheMetricsCollector 测试缓存指标收集器
func TestCacheMetricsCollector(t *testing.T) {
	config := &CacheMetricsConfig{
		Enabled:               true,
		CacheType:             CacheTypeRedis,
		CollectInterval:       1 * time.Second,
		HotKeyThreshold:       10,
		MaxHotKeys:            5,
		EnableKeyTracking:     true,
		EnableLatencyTracking: true,
	}

	collector := NewCacheMetricsCollector(config)
	if !collector.IsEnabled() {
		t.Error("Cache metrics collector should be enabled")
	}

	// 测试缓存命中记录
	collector.RecordCacheHit("test_key", true)
	collector.RecordCacheHit("test_key", false)

	stats := collector.GetHitStats()
	if stats.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", stats.TotalRequests)
	}
	if stats.HitCount != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.HitCount)
	}
	if stats.MissCount != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.MissCount)
	}

	t.Log("Cache metrics collector test successful")
}

// TestDatabaseMetricsCollector 测试数据库指标收集器
func TestDatabaseMetricsCollector(t *testing.T) {
	// 创建临时SQLite数据库用于测试
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	config := &DatabaseMetricsConfig{
		Enabled:              true,
		DatabaseType:         DatabaseTypeSQLite,
		SlowQueryThreshold:   1 * time.Millisecond,
		MaxSlowQueryExamples: 10,
		CollectInterval:      1 * time.Second,
		EnableStackTrace:     false,
		EnableQueryAnalysis:  true,
	}

	collector := NewDatabaseMetricsCollector(db, config)
	if !collector.IsEnabled() {
		t.Error("Database metrics collector should be enabled")
	}

	// 测试查询记录
	collector.RecordQuery("SELECT * FROM users", 2*time.Millisecond, nil)
	collector.RecordQuery("INSERT INTO users VALUES (?)", 5*time.Millisecond, nil, "test")

	queryMetrics := collector.GetQueryMetrics()
	if queryMetrics.TotalQueries != 2 {
		t.Errorf("Expected 2 total queries, got %d", queryMetrics.TotalQueries)
	}

	connectionMetrics := collector.GetConnectionPoolMetrics()
	if connectionMetrics == nil {
		t.Error("Connection pool metrics should not be nil")
	}

	t.Log("Database metrics collector test successful")
}

// TestSecurityMetricsCollector 测试安全指标收集器
func TestSecurityMetricsCollector(t *testing.T) {
	config := &SecurityMetricsConfig{
		Enabled:                     true,
		EnableSQLInjectionDetection: true,
		EnableXSSDetection:          true,
		EnableBruteForceDetection:   true,
		MaxFailedLogins:             3,
		BruteForceWindow:            5 * time.Minute,
		IPLockDuration:              10 * time.Minute,
		UserLockDuration:            5 * time.Minute,
		MaxSecurityEvents:           100,
		MaxIPThreatInfo:             50,
		CleanupInterval:             1 * time.Hour,
		EnableAlerts:                false, // 测试中不启用告警
		AlertThreshold:              5,
	}

	collector := NewSecurityMetricsCollector(config)
	if !collector.IsEnabled() {
		t.Error("Security metrics collector should be enabled")
	}

	// 测试威胁检测
	detected, threatLevel, description := collector.detectThreats(
		"POST", 
		"/api/login", 
		"username=admin' OR '1'='1", 
		"Mozilla/5.0",
	)

	if detected {
		t.Logf("Detected threat: level=%s, description=%s", threatLevel, description)
	} else {
		t.Log("No threat detected in test payload")
	}
	
	// 测试更明显的SQL注入攻击
	detected2, threatLevel2, description2 := collector.detectThreats(
		"POST", 
		"/api/login", 
		"username=admin'; DROP TABLE users; --", 
		"Mozilla/5.0",
	)
	
	if detected2 {
		t.Logf("Detected SQL injection: level=%s, description=%s", threatLevel2, description2)
	}

	// 测试白名单功能
	collector.AddToWhitelist("192.168.1.1")
	if !collector.isWhitelisted("192.168.1.1") {
		t.Error("IP should be whitelisted")
	}

	metrics := collector.GetMetrics()
	if metrics == nil {
		t.Error("Security metrics should not be nil")
	}

	t.Log("Security metrics collector test successful")
}