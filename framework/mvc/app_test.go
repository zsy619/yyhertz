package mvc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zsy619/yyhertz/framework/config"
)

// TestGetAppInstance 测试单例应用实例获取
func TestGetAppInstance(t *testing.T) {
	app1 := GetAppInstance()
	app2 := GetAppInstance()
	assert.NotNil(t, app1, "GetAppInstance should return a non-nil app instance")
	assert.Equal(t, app1, app2, "GetAppInstance should return the same instance (singleton)")
}

// TestNewAppWithLogConfig 测试应用实例创建
func TestNewAppWithLogConfig(t *testing.T) {
	logConfig := config.DefaultLogConfig()
	app := NewAppWithLogConfig(logConfig)
	assert.NotNil(t, app, "NewAppWithLogConfig should return a non-nil app instance")
	assert.Equal(t, "./views", app.GetViewPath(), "Default view path should be ./views")
	assert.Equal(t, "./static", app.GetStaticPath(), "Default static path should be ./static")
}

// TestSetViewPath 测试视图路径设置
func TestSetViewPath(t *testing.T) {
	app := NewApp()
	app.SetViewPath("/custom/views")
	assert.Equal(t, "/custom/views", app.GetViewPath(), "SetViewPath should update the view path")
}

// TestStaticPathSetting 测试静态文件路径设置（新API）
func TestStaticPathSetting(t *testing.T) {
	app := NewApp()
	// 只测试路径设置，不测试路由注册
	originalPath := app.GetStaticPath()
	assert.NotEmpty(t, originalPath, "Static path should not be empty")

	// 测试双参数调用：指定本地目录和URL路径
	app.SetStaticPath("custom/assets", "/assets")
	staticPaths := app.GetStaticPaths()
	assert.Equal(t, "custom/assets", staticPaths["/assets"], "Custom assets path should be set correctly")

	// 测试单参数调用：自动推导URL路径
	app.SetStaticPath("public")
	staticPaths = app.GetStaticPaths()
	assert.Equal(t, "public", staticPaths["/public"], "Public path should be auto-deduced")
	
	// 测试带 "./" 前缀的目录
	app.SetStaticPath("./uploads", "/files")
	staticPaths = app.GetStaticPaths()
	assert.Equal(t, "./uploads", staticPaths["/files"], "Uploads path should be set correctly")
}

// TestNewApp 测试基本应用创建
func TestNewApp(t *testing.T) {
	app := NewApp()
	assert.NotNil(t, app, "NewApp should return a non-nil app instance")
	assert.NotNil(t, app.GetLogger(), "App should have a logger")
	assert.NotNil(t, app.GetLogConfig(), "App should have log config")
}

// TestAppLogging 测试应用日志功能
func TestAppLogging(t *testing.T) {
	app := NewApp()

	// 测试基础日志方法（这些方法应该不会panic）
	assert.NotPanics(t, func() {
		app.LogInfo("Test info message")
		app.LogInfof("Test info message: %s", "formatted")
		app.LogError("Test error message")
		app.LogErrorf("Test error message: %s", "formatted")
		app.LogWarn("Test warn message")
		app.LogWarnf("Test warn message: %s", "formatted")
		app.LogDebug("Test debug message")
		app.LogDebugf("Test debug message: %s", "formatted")
	}, "Basic logging methods should not panic")
}

// TestAppConfig 测试应用配置功能
func TestAppConfig(t *testing.T) {
	app := NewApp()

	// 测试日志配置
	originalConfig := app.GetLogConfig()
	assert.NotNil(t, originalConfig, "App should have initial log config")

	// 测试更新日志级别
	assert.NotPanics(t, func() {
		app.UpdateLogLevel(config.LogLevelDebug)
	}, "UpdateLogLevel should not panic")
}

// TestSetStaticPathStaticMethod 测试SetStaticPath静态方法
func TestSetStaticPathStaticMethod(t *testing.T) {
	// 确保HertzApp已初始化
	assert.NotNil(t, HertzApp, "HertzApp should be initialized")
	
	// 测试静态方法调用
	assert.NotPanics(t, func() {
		SetStaticPath("test-assets", "/test-static")
		SetStaticPath("another-dir")
	}, "SetStaticPath static method should not panic")
	
	// 验证路径是否正确设置
	staticPaths := HertzApp.GetStaticPaths()
	assert.Equal(t, "test-assets", staticPaths["/test-static"], "Static method should set path correctly")
	assert.Equal(t, "another-dir", staticPaths["/another-dir"], "Static method should auto-deduce URL path")
}
