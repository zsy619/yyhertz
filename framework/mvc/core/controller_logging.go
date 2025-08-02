package core

import (
	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 日志方法 =============

// LogInfof 记录信息日志
func (c *BaseController) LogInfof(format string, args ...any) {
	config.Infof(format, args...)
}

// LogError 记录错误日志
func (c *BaseController) LogError(args ...any) {
	config.Error(args...)
}

// LogDebug 记录调试日志
func (c *BaseController) LogDebug(args ...any) {
	config.Debug(args...)
}

// LogWarn 记录警告日志
func (c *BaseController) LogWarn(args ...any) {
	config.Warn(args...)
}

// LogInfo 记录信息日志
func (c *BaseController) LogInfo(args ...any) {
	config.Info(args...)
}

// LogErrorf 记录格式化错误日志
func (c *BaseController) LogErrorf(format string, args ...any) {
	config.Errorf(format, args...)
}

// LogDebugf 记录格式化调试日志
func (c *BaseController) LogDebugf(format string, args ...any) {
	config.Debugf(format, args...)
}

// LogFetal 记录致命错误日志并终止程序
func (c *BaseController) LogFetal(args ...any) {
	config.Fatal(args...)
}

// LogFetalf 记录格式化致命错误日志并终止程序
func (c *BaseController) LogFetalf(format string, args ...any) {
	config.Fatalf(format, args...)
}

// LogPanic 记录恐慌日志并触发恐慌
func (c *BaseController) LogPanic(args ...any) {
	config.Panic(args...)
}

// LogPanicsf 记录格式化恐慌日志并触发恐慌
func (c *BaseController) LogPanicsf(format string, args ...any) {
	config.Panicf(format, args...)
}