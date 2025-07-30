package yyhertz

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zsy619/yyhertz/framework/config"
)

// TestGetAppInstance 测试单例模式
func TestGetAppInstance(t *testing.T) {
	app1 := GetAppInstance()
	app2 := GetAppInstance()
	assert.Equal(t, app1, app2, "GetAppInstance should return the same instance")
}

// TestNewAppWithLogConfig 测试应用实例创建
func TestNewAppWithLogConfig(t *testing.T) {
	logConfig := config.DefaultLogConfig()
	app := NewAppWithLogConfig(logConfig)
	assert.NotNil(t, app, "NewAppWithLogConfig should return a non-nil app instance")
	assert.Equal(t, ":8080", app.address, "Default address should be :8080")
	assert.Equal(t, "views", app.ViewPath, "Default view path should be views")
	assert.Equal(t, "static", app.StaticPath, "Default static path should be static")
}

// TestSetViewPath 测试视图路径设置
func TestSetViewPath(t *testing.T) {
	app := NewApp()
	app.SetViewPath("/custom/views")
	assert.Equal(t, "/custom/views", app.ViewPath, "SetViewPath should update the view path")
}

// TestSetStaticPath 测试静态文件路径设置
func TestSetStaticPath(t *testing.T) {
	app := NewApp()
	app.SetStaticPath("/custom/static")
	assert.Equal(t, "/custom/static", app.StaticPath, "SetStaticPath should update the static path")
}

// TestUseMiddleware 测试中间件添加
func TestUseMiddleware(t *testing.T) {
	app := NewApp()
	middlewareCalled := false
	middleware := func(ctx context.Context, r *RequestContext) {
		middlewareCalled = true
	}
	app.Use(middleware)
	assert.True(t, middlewareCalled, "Middleware should be called")
}

// TestLogFunctions 测试日志功能
func TestLogFunctions(t *testing.T) {
	app := NewApp()
	app.LogInfo("Test info log")
	app.LogError("Test error log")
	app.LogWarn("Test warn log")
	app.LogDebug("Test debug log")
	app.LogFatal("Test fatal log")
	app.LogPanic("Test panic log")

	fields := map[string]any{"key": "value"}
	app.LogWithFields(config.LogLevelInfo, "Test log with fields", fields)
	app.LogWithRequestID(config.LogLevelInfo, "Test log with request ID", "req123")
	app.LogWithUserID(config.LogLevelInfo, "Test log with user ID", "user123")
}

// TestGetLoggerWithContext 测试带上下文的日志获取
func TestGetLoggerWithContext(t *testing.T) {
	app := NewApp()
	ctx := &RequestContext{}
	logger := app.GetLoggerWithContext(ctx)
	assert.NotNil(t, logger, "GetLoggerWithContext should return a non-nil logger")
}

// TestRunServer 测试服务器启动逻辑
func TestRunServer(t *testing.T) {
	app := NewApp()
	app.address = ":0" // Use a random port for testing
	go func() {
		app.Run()
	}()
	time.Sleep(100 * time.Millisecond) // Give the server time to start
	assert.NotEmpty(t, app.address, "Server should be running on a non-empty address")
}
