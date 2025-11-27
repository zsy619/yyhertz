package devtools

import (
	"database/sql"
	"testing"

	"github.com/zsy619/yyhertz/framework/mvc"
	_ "github.com/mattn/go-sqlite3"
)

// TestPanelRoutes 测试面板路由是否正确注册
func TestPanelRoutes(t *testing.T) {
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
		EnableDebug:       false, // 只测试新面板
		EnableHealthCheck: false,
		EnableQPS:         false,
		EnableMetrics:     false,
		EnableProfiler:    false,
		EnableRateLimit:   false,
		EnableHotReload:   false,
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

	// 获取路由引擎
	engine := app.Engine
	
	// 检查路由是否存在
	routes := engine.Routes()
	
	expectedRoutes := []string{
		"/yyhertz/cache/panel",
		"/yyhertz/database/panel", 
		"/yyhertz/security/panel",
	}
	
	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Path] = true
	}
	
	for _, expectedRoute := range expectedRoutes {
		if !routeMap[expectedRoute] {
			t.Errorf("Route %s not found in registered routes", expectedRoute)
		} else {
			t.Logf("✅ Route %s found", expectedRoute)
		}
	}
	
	// 打印所有注册的路由供调试
	t.Logf("All registered routes:")
	for _, route := range routes {
		if route.Path[:9] == "/yyhertz/" {
			t.Logf("  %s %s", route.Method, route.Path)
		}
	}

	t.Log("Route registration test completed")
}