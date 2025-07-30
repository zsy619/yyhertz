package mvc

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	hertzlogrus "github.com/hertz-contrib/logger/logrus"
	"github.com/sirupsen/logrus"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/middleware"
)

var (
	appInstance *App
	once        sync.Once
	appMutex    sync.Mutex
)

// 类型别名定义
type RequestContext = app.RequestContext

// HandlerFunc 定义处理函数类型
// 这里的HandlerFunc是一个函数类型，接收context.Context和*RequestContext
// 返回值为void（即没有返回值），用于处理HTTP请求。
type HandlerFunc = func(context.Context, *RequestContext)

// AdaptHandler 将HandlerFunc适配为app.HandlerFunc
// 这个函数将自定义的HandlerFunc适配为Hertz框架所需的app.HandlerFunc类型。
func AdaptHandler(handler HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		handler(ctx, (*RequestContext)(c))
	}
}

// 控制器接口定义
type IController interface {
	Init()
	Prepare()
	Finish()
}

// App 应用结构
type App struct {
	*server.Hertz
	ViewPath      string
	StaticPath    string
	startTime     time.Time
	address       string
	loggerManager *config.LoggerManager
}

// GetAppInstance 获取单例应用实例
func GetAppInstance() *App {
	once.Do(func() {
		appMutex.Lock()
		defer appMutex.Unlock()
		appInstance = NewAppWithLogConfig(config.DefaultLogConfig())
	})
	return appInstance
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return NewAppWithLogConfig(config.DefaultLogConfig())
}

// NewAppWithLogConfig 使用指定日志配置创建应用实例
func NewAppWithLogConfig(logConfig *config.LogConfig) *App {
	// 创建Hertz服务器实例
	port := config.GetConfigInt("app.port")
	if port == 0 {
		port = 8080 // 默认端口
	}
	host := config.GetConfigString("app.host")
	if host == "" {
		host = "0.0.0.0"
	}
	// 创建Hertz服务器实例，使用指定的端口
	// 这里使用了server.WithHostPorts来指定监听的地址和端口
	h := server.Default(server.WithHostPorts(host + ":" + strconv.Itoa(port)))

	// 初始化全局日志管理器
	loggerManager := config.InitGlobalLogger(logConfig)

	app := &App{
		Hertz:         h,
		ViewPath:      "views",
		StaticPath:    "static",
		startTime:     time.Now(),
		address:       ":8080",
		loggerManager: loggerManager,
	}

	// 配置视图和静态文件路径
	app.SetViewPath("/views")
	app.SetStaticPath("/static")

	// 配置增强的日志中间件
	loggerConfig := &middleware.MiddlewareLoggerConfig{
		EnableRequestBody:  true,  // 启用请求体记录用于演示
		EnableResponseBody: false, // 不记录响应体以提高性能
		SkipPaths:          []string{"/health", "/ping"},
		MaxBodySize:        512, // 限制记录的请求体大小
	}

	// 添加全局中间件
	app.Use(
		middleware.RecoveryMiddleware(),
		middleware.TracingMiddleware(),
		middleware.LoggerMiddlewareWithConfig(loggerConfig),
		middleware.CORSMiddleware(),
		middleware.RateLimitMiddleware(100, time.Minute),
	)

	// 设置路由
	{
		// 健康检查路由（会被日志中间件跳过）
		app.GET("/health", func(c context.Context, ctx *RequestContext) {
			ctx.JSON(consts.StatusOK, map[string]string{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
		})

		// ping路由（也会被跳过）
		app.GET("/ping", func(c context.Context, ctx *RequestContext) {
			ctx.JSON(consts.StatusOK, map[string]string{"message": "pong"})
		})
	}

	return app
}

// SetViewPath 设置视图路径
func (app *App) SetViewPath(path string) {
	app.ViewPath = path
}

func (app *App) GetViewPath() string {
	return app.ViewPath
}

// SetStaticPath 设置静态文件路径
func (app *App) SetStaticPath(path string) {
	app.StaticPath = path
	app.Static("/static", path)
}

func (app *App) GetStaticPath() string {
	return app.StaticPath
}

// Use 添加中间件
func (app *App) Use(middleware ...HandlerFunc) {
	for _, m := range middleware {
		app.Hertz.Use(m)
	}
}

// Run 启动服务器
func (app *App) Run(addr ...string) {
	if len(addr) > 0 {
		app.address = addr[0]
	}

	// 启动服务器，忽略地址参数，使用默认配置
	app.Hertz.Spin()
}

// ============= 日志方法 =============

// GetLogger 获取日志实例
func (app *App) GetLogger() *hertzlogrus.Logger {
	return app.loggerManager.GetLogger()
}

// GetLogConfig 获取当前日志配置
func (app *App) GetLogConfig() *config.LogConfig {
	return app.loggerManager.GetConfig()
}

// SetLogConfig 设置日志配置
func (app *App) SetLogConfig(logConfig *config.LogConfig) {
	app.loggerManager.UpdateConfig(logConfig)
}

// UpdateLogLevel 动态更新日志级别
func (app *App) UpdateLogLevel(level config.LogLevel) {
	app.loggerManager.UpdateLevel(level)
}

// LogInfo 记录信息日志
func (app *App) LogInfof(format string, args ...any) {
	config.Infof(format, args...)
}

// LogInfo 记录信息日志
func (app *App) LogInfo(args ...any) {
	config.Info(args...)
}

// LogError 记录错误日志
func (app *App) LogErrorf(format string, args ...any) {
	config.Errorf(format, args...)
}

// LogError 记录错误日志
func (app *App) LogError(args ...any) {
	config.Error(args...)
}

// LogWarn 记录警告日志
func (app *App) LogWarnf(format string, args ...any) {
	config.Warnf(format, args...)
}

// LogWarn 记录警告日志
func (app *App) LogWarn(args ...any) {
	config.Warn(args...)
}

// LogDebug 记录调试日志
func (app *App) LogDebugf(format string, args ...any) {
	config.Debugf(format, args...)
}

// LogDebug 记录调试日志
func (app *App) LogDebug(args ...any) {
	config.Debug(args...)
}

// LogFatal 记录致命错误日志
func (app *App) LogFatalf(format string, args ...any) {
	config.Fatalf(format, args...)
}

// LogFatal 记录致命错误日志
func (app *App) LogFatal(args ...any) {
	config.Fatal(args...)
}

// LogPanic 记录恐慌日志
func (app *App) LogPanicf(format string, args ...any) {
	config.Panicf(format, args...)
}

// LogPanic 记录恐慌日志
func (app *App) LogPanic(args ...any) {
	config.Panic(args...)
}

// LogWithFields 记录带字段的日志
func (app *App) LogWithFields(level config.LogLevel, msg string, fields map[string]any) {
	entry := config.WithFields(fields)

	app.log(entry, level, "", msg)
}

// LogfWithFields 记录带上下文的格式日志
func (app *App) LogfWithFields(level config.LogLevel, format, msg string, fields map[string]any) {
	entry := config.WithFields(fields)

	app.log(entry, level, format, msg)
}

// LogWithRequestID 记录带请求ID的日志
func (app *App) LogWithRequestID(level config.LogLevel, msg string, requestID string) {
	entry := config.WithRequestID(requestID)

	app.log(entry, level, "", msg)
}

// LogfWithContext 记录带上下文的格式日志
func (app *App) LogfWithContext(level config.LogLevel, format, msg string, requestID string) {
	entry := config.WithRequestID(requestID)

	app.log(entry, level, format, msg)
}

// LogWithUserID 记录带用户ID的日志
func (app *App) LogWithUserID(level config.LogLevel, msg string, userID string) {
	entry := config.WithUserID(userID)

	app.log(entry, level, "", msg)
}

// LogfWithUserID 记录带用户ID的格式日志
func (app *App) LogfWithUserID(level config.LogLevel, format, msg string, userID string) {
	entry := config.WithUserID(userID)

	app.log(entry, level, format, msg)
}

func (app *App) log(entry *logrus.Entry, level config.LogLevel, format, msg string) {
	if format == "" {
		switch level {
		case config.LogLevelDebug:
			entry.Debug(msg)
		case config.LogLevelInfo:
			entry.Info(msg)
		case config.LogLevelWarn:
			entry.Warn(msg)
		case config.LogLevelError:
			entry.Error(msg)
		case config.LogLevelFatal:
			entry.Fatal(msg)
		case config.LogLevelPanic:
			entry.Panic(msg)
		default:
			entry.Info(msg)
		}
	} else {
		switch level {
		case config.LogLevelDebug:
			entry.Debugf(format, msg)
		case config.LogLevelInfo:
			entry.Infof(format, msg)
		case config.LogLevelWarn:
			entry.Warnf(format, msg)
		case config.LogLevelError:
			entry.Errorf(format, msg)
		case config.LogLevelFatal:
			entry.Fatalf(format, msg)
		case config.LogLevelPanic:
			entry.Panicf(format, msg)
		default:
			entry.Infof(format, msg)
		}
	}
}

// GetLoggerWithContext 获取带上下文信息的logger
func (app *App) GetLoggerWithContext(ctx *RequestContext) *hertzlogrus.Logger {
	// 使用全局日志管理器
	return config.GetGlobalLogger().GetLogger()
}
