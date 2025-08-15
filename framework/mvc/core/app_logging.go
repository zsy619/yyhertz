package core

import (
	hertzlogrus "github.com/hertz-contrib/logger/logrus"
	"github.com/zsy619/yyhertz/framework/config"
)

// ============================================================================
// 应用日志功能模块
// ============================================================================
//
// 本文件提供了 MVC 应用的完整日志功能，包括：
//
// 1. 日志配置管理：动态配置、级别调整、实例获取
// 2. 基础日志方法：Info、Error、Warn、Debug 等标准日志级别
// 3. 格式化日志：支持 Printf 风格的格式化输出
// 4. 上下文日志：集成请求上下文的日志记录
// 5. 特殊日志：Fatal、Panic 等程序控制日志
//
// 设计特点：
// - 统一的日志接口，支持多种日志后端
// - 集成 Hertz 框架的日志中间件
// - 支持动态配置和级别调整
// - 提供上下文感知的日志记录
// - 兼容标准库和第三方日志库
//
// 使用示例：
//   app.LogInfo("Server started")
//   app.LogErrorf("Failed to connect: %v", err)
//   logger := app.GetLogger()
//
// ============================================================================

// ============= 日志实例和配置管理 =============

// GetLogger 获取全局日志实例
//
// 返回当前应用使用的 Hertz Logrus 日志实例，可用于：
// - 自定义日志格式和输出
// - 添加结构化字段
// - 集成第三方日志系统
//
// 返回:
//   *hertzlogrus.Logger: Hertz 集成的 Logrus 日志实例
//
// 示例:
//   logger := app.GetLogger()
//   logger.WithField("user_id", 123).Info("User logged in")
func (app *App) GetLogger() *hertzlogrus.Logger {
	return app.loggerManager.GetLogger()
}

// GetLogConfig 获取当前日志配置
//
// 返回应用当前使用的日志配置信息，包括：
// - 日志级别设置
// - 输出格式配置
// - 文件轮转设置
// - 其他日志相关参数
//
// 返回:
//   *config.LogConfig: 当前的日志配置对象
//
// 用途:
//   用于查看和验证当前的日志配置状态
func (app *App) GetLogConfig() *config.LogConfig {
	return app.loggerManager.GetConfig()
}

// SetLogConfig 设置日志配置
//
// 动态更新应用的日志配置，支持运行时修改：
// - 日志级别调整
// - 输出格式变更
// - 文件路径修改
// - 其他配置参数
//
// 参数:
//   logConfig *config.LogConfig: 新的日志配置对象
//
// 注意:
//   配置更新会立即生效，影响后续的所有日志输出
//
// 示例:
//   config := &config.LogConfig{
//       Level: config.InfoLevel,
//       Format: "json",
//   }
//   app.SetLogConfig(config)
func (app *App) SetLogConfig(logConfig *config.LogConfig) {
	app.loggerManager.UpdateConfig(logConfig)
}

// UpdateLogLevel 动态更新日志级别
//
// 快速调整应用的日志级别，常用于：
// - 运行时调试：临时提高日志详细度
// - 性能优化：降低生产环境日志级别
// - 故障排查：动态开启详细日志
//
// 参数:
//   level config.LogLevel: 新的日志级别 (Debug, Info, Warn, Error, Fatal, Panic)
//
// 支持的级别:
//   - DebugLevel: 最详细，包含调试信息
//   - InfoLevel: 一般信息
//   - WarnLevel: 警告信息
//   - ErrorLevel: 错误信息
//   - FatalLevel: 致命错误，程序退出
//   - PanicLevel: panic 级别
//
// 示例:
//   app.UpdateLogLevel(config.DebugLevel)  // 开启调试模式
//   app.UpdateLogLevel(config.ErrorLevel) // 只记录错误
func (app *App) UpdateLogLevel(level config.LogLevel) {
	app.loggerManager.UpdateLevel(level)
}

// ============= 基础日志记录方法 =============

// LogInfof 记录格式化信息日志
//
// 用于记录应用的一般性信息，如：
// - 服务启动/停止状态
// - 重要操作完成通知
// - 系统状态变更
// - 用户行为记录
//
// 参数:
//   format string: Printf 风格的格式字符串
//   args ...any: 格式化参数
//
// 示例:
//   app.LogInfof("Server started on port %d", 8080)
//   app.LogInfof("User %s logged in from %s", username, ip)
func (app *App) LogInfof(format string, args ...any) {
	config.Infof(format, args...)
}

// LogInfo 记录信息日志
//
// 用于记录不需要格式化的一般性信息，参数会被直接拼接输出。
//
// 参数:
//   args ...any: 要记录的信息内容
//
// 示例:
//   app.LogInfo("Application initialized successfully")
//   app.LogInfo("Cache cleared, entries:", count)
func (app *App) LogInfo(args ...any) {
	config.Info(args...)
}

// LogErrorf 记录格式化错误日志
//
// 用于记录应用错误和异常情况，如：
// - 数据库连接失败
// - 文件操作错误
// - 网络请求失败
// - 业务逻辑错误
//
// 参数:
//   format string: Printf 风格的格式字符串
//   args ...any: 格式化参数
//
// 示例:
//   app.LogErrorf("Database connection failed: %v", err)
//   app.LogErrorf("Failed to process user %d: %s", userID, err.Error())
func (app *App) LogErrorf(format string, args ...any) {
	config.Errorf(format, args...)
}

// LogError 记录错误日志
//
// 用于记录不需要格式化的错误信息。
//
// 参数:
//   args ...any: 错误信息内容
//
// 示例:
//   app.LogError("Failed to load configuration")
//   app.LogError("Critical error:", err)
func (app *App) LogError(args ...any) {
	config.Error(args...)
}

// LogWarnf 记录格式化警告日志
//
// 用于记录需要注意但不影响正常运行的情况，如：
// - 配置项使用默认值
// - 性能指标异常
// - 资源使用率过高
// - 不推荐的API使用
//
// 参数:
//   format string: Printf 风格的格式字符串
//   args ...any: 格式化参数
//
// 示例:
//   app.LogWarnf("Memory usage is high: %.2f%%", usage*100)
//   app.LogWarnf("Deprecated API used by client %s", clientID)
func (app *App) LogWarnf(format string, args ...any) {
	config.Warnf(format, args...)
}

// LogWarn 记录警告日志
//
// 用于记录不需要格式化的警告信息。
//
// 参数:
//   args ...any: 警告信息内容
//
// 示例:
//   app.LogWarn("Configuration file not found, using defaults")
//   app.LogWarn("High CPU usage detected:", cpuPercent)
func (app *App) LogWarn(args ...any) {
	config.Warn(args...)
}

// LogDebugf 记录格式化调试日志
//
// 用于开发和调试时的详细信息记录，如：
// - 函数调用跟踪
// - 变量状态输出
// - 流程执行步骤
// - 性能分析数据
//
// 注意:
//   只在 Debug 级别及以上时输出，生产环境通常被过滤
//
// 参数:
//   format string: Printf 风格的格式字符串
//   args ...any: 格式化参数
//
// 示例:
//   app.LogDebugf("Processing request: method=%s, path=%s", method, path)
//   app.LogDebugf("Cache hit rate: %.2f%% (%d/%d)", rate, hits, total)
func (app *App) LogDebugf(format string, args ...any) {
	config.Debugf(format, args...)
}

// LogDebug 记录调试日志
//
// 用于记录不需要格式化的调试信息。
//
// 参数:
//   args ...any: 调试信息内容
//
// 示例:
//   app.LogDebug("Entering user authentication flow")
//   app.LogDebug("Database query executed:", sql)
func (app *App) LogDebug(args ...any) {
	config.Debug(args...)
}

// ============= 上下文感知日志 =============

// GetLoggerWithContext 获取带请求上下文信息的日志实例
//
// 返回一个集成了当前请求上下文的日志实例，可自动包含：
// - 请求ID（用于请求跟踪）
// - 用户信息（如已认证）
// - 会话标识
// - 其他上下文相关信息
//
// 参数:
//   ctx *RequestContext: 当前请求的上下文对象
//
// 返回:
//   *hertzlogrus.Logger: 带上下文信息的日志实例
//
// 用途:
//   - 请求级别的日志跟踪
//   - 用户行为分析
//   - 错误定位和调试
//
// 示例:
//   logger := app.GetLoggerWithContext(ctx)
//   logger.Info("User action completed")
//   // 输出会自动包含请求相关的上下文信息
func (app *App) GetLoggerWithContext(ctx *RequestContext) *hertzlogrus.Logger {
	return config.GetGlobalLogger().GetLogger()
}

// ============= 特殊控制日志 =============

// LogFatal 记录致命错误日志并终止程序
//
// 用于记录无法恢复的严重错误，如：
// - 关键配置文件缺失
// - 必要服务连接失败
// - 系统资源不足
// - 数据完整性破坏
//
// 警告:
//   调用此方法后程序会立即退出（os.Exit(1)），请谨慎使用
//
// 参数:
//   args ...any: 致命错误信息
//
// 示例:
//   app.LogFatal("Cannot connect to database, shutting down")
//   app.LogFatal("Critical configuration error:", err)
func (app *App) LogFatal(args ...any) {
	config.Fatal(args...)
}

// LogFatalf 记录格式化致命错误日志并终止程序
//
// 功能同 LogFatal，但支持 Printf 风格的格式化输出。
//
// 警告:
//   调用此方法后程序会立即退出（os.Exit(1)），请谨慎使用
//
// 参数:
//   format string: Printf 风格的格式字符串
//   args ...any: 格式化参数
//
// 示例:
//   app.LogFatalf("Failed to bind to port %d: %v", port, err)
//   app.LogFatalf("Database schema version mismatch: expected %d, got %d", expected, actual)
func (app *App) LogFatalf(format string, args ...any) {
	config.Fatalf(format, args...)
}

// LogPanic 记录 panic 日志并触发 panic
//
// 用于记录程序逻辑错误或不应该发生的情况，如：
// - 断言失败
// - 无效的程序状态
// - 不可能的代码路径
// - 严重的逻辑错误
//
// 警告:
//   调用此方法会立即触发 panic，除非被 recover，否则程序会崩溃
//
// 参数:
//   args ...any: panic 信息
//
// 示例:
//   app.LogPanic("Invalid state: user session is nil")
//   app.LogPanic("Unreachable code executed")
func (app *App) LogPanic(args ...any) {
	config.Panic(args...)
}

// LogPanicf 记录格式化 panic 日志并触发 panic
//
// 功能同 LogPanic，但支持 Printf 风格的格式化输出。
//
// 警告:
//   调用此方法会立即触发 panic，除非被 recover，否则程序会崩溃
//
// 参数:
//   format string: Printf 风格的格式字符串
//   args ...any: 格式化参数
//
// 示例:
//   app.LogPanicf("Invalid enum value: %d (expected 0-%d)", value, maxValue)
//   app.LogPanicf("Assertion failed: %s should not be %v", name, value)
func (app *App) LogPanicf(format string, args ...any) {
	config.Panicf(format, args...)
}