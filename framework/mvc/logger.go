package mvc

// 基础日志方法
func LogInfof(format string, args ...any) {
	HertzApp.LogInfof(format, args...)
}

func LogInfo(args ...any) {
	HertzApp.LogInfo(args...)
}

func LogErrorf(format string, args ...any) {
	HertzApp.LogErrorf(format, args...)
}

func LogError(args ...any) {
	HertzApp.LogError(args...)
}

func LogWarnf(format string, args ...any) {
	HertzApp.LogWarnf(format, args...)
}

func LogWarn(args ...any) {
	HertzApp.LogWarn(args...)
}

func LogDebugf(format string, args ...any) {
	HertzApp.LogDebugf(format, args...)
}

func LogDebug(args ...any) {
	HertzApp.LogDebug(args...)
}

// Fatal 全局致命错误日志
func LogFatal(args ...any) {
	HertzApp.LogFatal(args...)
}

// Fatalf 全局格式化致命错误日志
func LogFatalf(format string, args ...any) {
	HertzApp.LogFatalf(format, args...)
}

// Panic 全局panic日志
func LogPanic(args ...any) {
	HertzApp.LogPanic(args...)
}

// Panicf 全局格式化panic日志
func LogPanicf(format string, args ...any) {
	HertzApp.LogPanicf(format, args...)
}
