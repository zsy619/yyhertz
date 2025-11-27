// Package mvc 提供MVC框架的全局日志记录功能
//
// 本文件提供了对底层日志系统的封装，允许在应用的任何地方使用统一的日志接口。
// 所有日志操作都委托给全局HertzApp实例的日志系统。
//
// 特性：
// - 支持格式化和非格式化日志记录
// - 支持多种日志级别：Debug、Info、Warn、Error、Fatal、Panic
// - 自动使用全局应用配置的日志格式和输出目标
// - 线程安全，可在并发环境中使用
package mvc

// ============= 基础日志方法 =============

// LogInfof 记录格式化信息日志
//
// 用于记录应用运行过程中的一般信息，如操作成功、状态变更等。
// 支持printf风格的格式化字符串。
//
// 参数：
//   - format: string - 格式化字符串，支持%s、%d、%v等占位符
//   - args: ...any - 格式化参数列表
//
// 使用示例：
//
//	mvc.LogInfof("用户 %s 登录成功，IP: %s", username, clientIP)
//	mvc.LogInfof("处理请求耗时: %dms", duration)
func LogInfof(format string, args ...any) {
	HertzApp.LogInfof(format, args...)
}

// LogInfo 记录信息日志
//
// 用于记录应用运行过程中的一般信息。支持传入多个参数，
// 这些参数会被自动连接成一条日志消息。
//
// 参数：
//   - args: ...any - 要记录的信息列表
//
// 使用示例：
//
//	mvc.LogInfo("应用启动成功")
//	mvc.LogInfo("用户", username, "执行操作", operation)
func LogInfo(args ...any) {
	HertzApp.LogInfo(args...)
}

// LogErrorf 记录格式化错误日志
//
// 用于记录应用运行过程中的错误信息，如操作失败、异常情况等。
// 支持printf风格的格式化字符串。
//
// 参数：
//   - format: string - 格式化字符串
//   - args: ...any - 格式化参数列表
//
// 使用示例：
//
//	mvc.LogErrorf("数据库连接失败: %v", err)
//	mvc.LogErrorf("用户 %s 认证失败，尝试次数: %d", username, attempts)
func LogErrorf(format string, args ...any) {
	HertzApp.LogErrorf(format, args...)
}

// LogError 记录错误日志
//
// 用于记录应用运行过程中的错误信息。支持传入多个参数。
//
// 参数：
//   - args: ...any - 要记录的错误信息列表
//
// 使用示例：
//
//	mvc.LogError("数据库连接失败:", err)
//	mvc.LogError("处理用户请求时发生错误")
func LogError(args ...any) {
	HertzApp.LogError(args...)
}

// LogWarnf 记录格式化警告日志
//
// 用于记录可能存在问题但不影响正常运行的情况。
// 支持printf风格的格式化字符串。
//
// 参数：
//   - format: string - 格式化字符串
//   - args: ...any - 格式化参数列表
//
// 使用示例：
//
//	mvc.LogWarnf("缓存连接不稳定，重试次数: %d", retryCount)
//	mvc.LogWarnf("用户 %s 权限不足，尝试访问: %s", username, resource)
func LogWarnf(format string, args ...any) {
	HertzApp.LogWarnf(format, args...)
}

// LogWarn 记录警告日志
//
// 用于记录可能存在问题但不影响正常运行的情况。
//
// 参数：
//   - args: ...any - 要记录的警告信息列表
//
// 使用示例：
//
//	mvc.LogWarn("缓存未命中，将查询数据库")
//	mvc.LogWarn("检测到异常流量")
func LogWarn(args ...any) {
	HertzApp.LogWarn(args...)
}

// LogDebugf 记录格式化调试日志
//
// 用于记录详细的调试信息，通常只在开发和测试环境中启用。
// 支持printf风格的格式化字符串。
//
// 参数：
//   - format: string - 格式化字符串
//   - args: ...any - 格式化参数列表
//
// 使用示例：
//
//	mvc.LogDebugf("处理请求参数: %+v", requestParams)
//	mvc.LogDebugf("SQL查询: %s, 参数: %v", sql, params)
func LogDebugf(format string, args ...any) {
	HertzApp.LogDebugf(format, args...)
}

// LogDebug 记录调试日志
//
// 用于记录详细的调试信息，通常只在开发和测试环境中启用。
//
// 参数：
//   - args: ...any - 要记录的调试信息列表
//
// 使用示例：
//
//	mvc.LogDebug("进入用户认证流程")
//	mvc.LogDebug("变量状态:", variable)
func LogDebug(args ...any) {
	HertzApp.LogDebug(args...)
}

// ============= 严重错误日志方法 =============

// LogFatal 记录致命错误日志并终止程序
//
// 用于记录导致程序无法继续运行的严重错误。记录日志后程序会调用os.Exit(1)终止。
// 应谨慎使用，只在确实需要立即终止程序时使用。
//
// 参数：
//   - args: ...any - 要记录的致命错误信息列表
//
// 注意事项：
//   - 调用此方法后程序会立即退出
//   - 不会执行defer语句
//   - 适用于初始化失败、关键资源不可用等情况
//
// 使用示例：
//
//	if err := initDatabase(); err != nil {
//		mvc.LogFatal("数据库初始化失败:", err)
//	}
func LogFatal(args ...any) {
	HertzApp.LogFatal(args...)
}

// LogFatalf 记录格式化致命错误日志并终止程序
//
// 类似LogFatal，但支持printf风格的格式化字符串。
//
// 参数：
//   - format: string - 格式化字符串
//   - args: ...any - 格式化参数列表
//
// 使用示例：
//
//	mvc.LogFatalf("配置文件 %s 加载失败: %v", configFile, err)
func LogFatalf(format string, args ...any) {
	HertzApp.LogFatalf(format, args...)
}

// LogPanic 记录panic日志并触发panic
//
// 用于记录严重错误并触发panic。与LogFatal不同，panic可以被recover捕获。
// 适用于检测到不可恢复的程序状态错误时使用。
//
// 参数：
//   - args: ...any - 要记录的panic信息列表
//
// 注意事项：
//   - 会触发panic，可被defer+recover捕获
//   - 适用于程序逻辑错误、不变式违反等情况
//   - 应配合恢复机制使用
//
// 使用示例：
//
//	if user == nil {
//		mvc.LogPanic("用户对象不能为nil")
//	}
func LogPanic(args ...any) {
	HertzApp.LogPanic(args...)
}

// LogPanicf 记录格式化panic日志并触发panic
//
// 类似LogPanic，但支持printf风格的格式化字符串。
//
// 参数：
//   - format: string - 格式化字符串
//   - args: ...any - 格式化参数列表
//
// 使用示例：
//
//	mvc.LogPanicf("数组越界访问: index=%d, length=%d", index, length)
func LogPanicf(format string, args ...any) {
	HertzApp.LogPanicf(format, args...)
}
