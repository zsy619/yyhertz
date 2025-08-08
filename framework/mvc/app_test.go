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
	assert.Equal(t, "/views", app.GetViewPath(), "Default view path should be /views")
	assert.Equal(t, "/static", app.GetStaticPath(), "Default static path should be /static")
}

// TestSetViewPath 测试视图路径设置
func TestSetViewPath(t *testing.T) {
	app := NewApp()
	app.SetViewPath("/custom/views")
	assert.Equal(t, "/custom/views", app.GetViewPath(), "SetViewPath should update the view path")
}

// TestStaticPathSetting 测试静态文件路径设置（不测试路由注册）
func TestStaticPathSetting(t *testing.T) {
	app := NewApp()
	// 只测试路径设置，不测试路由注册
	originalPath := app.GetStaticPath()
	assert.NotEmpty(t, originalPath, "Static path should not be empty")

	// 直接设置StaticPath字段，避免路由冲突
	app.SetStaticPath("/custom/assets")
	assert.Equal(t, "/custom/assets", app.GetStaticPath(), "StaticPath field should be updated")
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
