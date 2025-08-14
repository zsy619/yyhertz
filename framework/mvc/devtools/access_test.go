package devtools

import (
	"database/sql"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc"
	_ "github.com/mattn/go-sqlite3"
)

// TestFullDevToolsWithNewPanels 测试启用所有功能的完整设置
func TestFullDevToolsWithNewPanels(t *testing.T) {
	app := mvc.NewApp()
	
	// 创建临时SQLite数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// 完整配置，启用所有功能包括新面板
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
		Database:          db,    // 提供数据库连接
		RedisAddr:         "",    // 测试中不需要Redis
		ExternalServices:  []string{},
		RateLimitRules:    []*RateLimit{},
		WhitelistIPs:      []string{"127.0.0.1"},
	}

	err = SetupDevTools(app, config)
	if err != nil {
		t.Fatalf("SetupDevTools failed: %v", err)
	}

	// 验证所有路由是否注册
	engine := app.Engine
	routes := engine.Routes()
	
	expectedRoutes := []string{
		"/yyhertz/debug/panel",
		"/yyhertz/health/panel",
		"/yyhertz/qps/panel",
		"/yyhertz/metrics/panel",
		"/yyhertz/profile/panel",
		"/yyhertz/ratelimit/panel",
		"/yyhertz/cache/panel",     // 新面板
		"/yyhertz/database/panel",  // 新面板
		"/yyhertz/security/panel",  // 新面板
	}
	
	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Path] = true
	}
	
	foundCount := 0
	for _, expectedRoute := range expectedRoutes {
		if routeMap[expectedRoute] {
			t.Logf("✅ %s - 路由已注册", expectedRoute)
			foundCount++
		} else {
			t.Errorf("❌ %s - 路由未找到", expectedRoute)
		}
	}
	
	t.Logf("总计：%d/%d 个面板路由已注册", foundCount, len(expectedRoutes))
	
	if foundCount == len(expectedRoutes) {
		t.Log("🎉 所有面板都已成功注册！")
	} else {
		t.Errorf("❌ 有 %d 个面板路由未注册", len(expectedRoutes)-foundCount)
	}
}

// TestDefaultConfigAccess 测试默认配置下的面板访问
func TestDefaultConfigAccess(t *testing.T) {
	app := mvc.NewApp()

	// 使用默认配置
	err := SetupBasicDevTools(app)
	if err != nil {
		t.Fatalf("SetupBasicDevTools failed: %v", err)
	}

	// 检查哪些面板被启用了
	engine := app.Engine
	routes := engine.Routes()
	
	newPanelRoutes := []string{
		"/yyhertz/cache/panel",
		"/yyhertz/database/panel",
		"/yyhertz/security/panel",
	}
	
	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Path] = true
	}
	
	t.Log("默认配置下新面板的状态：")
	for _, route := range newPanelRoutes {
		if routeMap[route] {
			t.Logf("✅ %s - 可访问", route)
		} else {
			t.Logf("❌ %s - 不可访问（需要在配置中启用）", route)
		}
	}
}