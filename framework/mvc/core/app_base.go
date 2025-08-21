package core

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc/define"
	"github.com/zsy619/yyhertz/framework/mvc/errors"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
)

var (
	appInstance *App
	once        sync.Once
	appMutex    sync.Mutex
)

// ===== 多种处理器函数类型定义 =====

// 过滤器位置常量 - 使用统一常量
const (
	BeforeStatic = constant.BeforeStatic // 静态文件处理前
	BeforeRouter = constant.BeforeRouter // 路由匹配前
	BeforeExec   = constant.BeforeExec   // 控制器执行前
	AfterExec    = constant.AfterExec    // 控制器执行后
	FinishRouter = constant.FinishRouter // 请求处理完成后
)

// FilterPattern 过滤器模式匹配结构
type FilterPattern struct {
	Pattern  string            // 路径模式 (支持通配符)
	Position int               // 过滤器位置
	Filter   define.FilterFunc // 过滤器函数
	Enabled  bool              // 是否启用
	Priority int               // 优先级
}

// ============= 错误处理相关类型别名 =============

// 类型别名，统一引用errors包中的定义
type ErrorHandler = errors.ErrorHandler
type ErrorRegistry = errors.ErrorRegistry
type ErrorConfig = errors.ErrorConfig
type ErrorHandlerFunc = errors.ErrorHandlerFunc

// DefaultErrorConfig 返回默认错误配置
func DefaultErrorConfig() ErrorConfig {
	config := errors.DefaultErrorConfig()
	return *config
}

// AdaptHandler 将HandlerFunc适配为app.HandlerFunc
func AdaptHandler(handler define.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		handler(ctx, (*define.RequestContext)(c))
	}
}

// App 应用结构（精简版，只保留核心功能）
type App struct {
	*server.Hertz
	ViewPaths     []string
	StaticPaths   map[string]string // URL路径 -> 本地路径映射
	startTime     time.Time
	address       string
	loggerManager *config.LoggerManager

	// 过滤器管理
	filters      map[int][]*FilterPattern // 按位置分组的过滤器
	filtersMutex sync.RWMutex             // 过滤器读写锁
	nextFilterID int64                    // 下一个过滤器ID (用于排序)

	// 错误处理管理
	errorRegistry *ErrorRegistry // 错误处理注册器接口
	errorConfig   ErrorConfig    // 错误处理配置
	errorMutex    sync.RWMutex   // 错误处理读写锁
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
	port := config.GetAppConfigInt("app.port")
	if port == 0 {
		port = 8080 // 默认端口
	}
	host := config.GetAppConfigString("app.host")
	if host == "" {
		host = "0.0.0.0"
	}

	// 创建Hertz服务器实例
	h := server.Default(server.WithHostPorts(host + ":" + strconv.Itoa(port)))

	// 初始化全局日志管理器
	loggerManager := config.InitGlobalLogger(logConfig)

	viewPaths := config.GlobalTemplate.Paths.ViewPaths
	staticPaths := config.GlobalTemplate.Paths.StaticPaths

	// 如果配置中没有静态路径，使用默认值
	if len(staticPaths) == 0 {
		staticPaths = map[string]string{
			"/favicon.ico": "./static/favicon.ico",
			"/static":      "./static",    // 基础静态文件
			"/assets":      "./assets",    // 前端资源
			"/uploads":     "./uploads",   // 用户上传
			"/downloads":   "./downloads", // 用户下载
			"/public":      "./public",    // 公共资源
		}
	}

	app := &App{
		Hertz:         h,                                // 使用Hertz服务器实例
		ViewPaths:     viewPaths,                        // 从配置读取视图路径
		StaticPaths:   staticPaths,                      // 从配置读取静态文件路径映射
		startTime:     time.Now(),                       // 记录应用启动时间
		address:       fmt.Sprintf("%s:%d", host, port), // 应用监听地址
		loggerManager: loggerManager,                    // 日志管理器

		// 初始化过滤器管理
		filters:      make(map[int][]*FilterPattern),
		nextFilterID: 0,

		// 初始化错误处理
		errorConfig: DefaultErrorConfig(),
	}

	// 注册默认静态路径
	for urlPath, _ := range app.StaticPaths {
		app.Static(urlPath, ".")
	}

	// 配置增强的日志中间件
	loggerConfig := &middleware.MiddlewareLoggerConfig{
		EnableRequestBody:  true,
		EnableResponseBody: false,
		SkipPaths:          []string{"/health", "/ping"},
		MaxBodySize:        512,
	}

	// 添加基础全局中间件
	app.Use(
		middleware.RecoveryMiddleware(),
		middleware.TracingMiddleware(),
		middleware.LoggerMiddlewareWithConfig(loggerConfig),
		middleware.CORSMiddleware(),
		middleware.RateLimitMiddleware(100, time.Minute),
	)

	// 设置基础路由
	app.setupBasicRoutes()

	return app
}

// setupBasicRoutes 设置基础路由
func (app *App) setupBasicRoutes() {
	// 健康检查路由
	app.GET("/health", func(c context.Context, ctx *define.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// ping路由
	app.GET("/ping", func(c context.Context, ctx *define.RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{"message": "pong"})
	})
}

// SetViewPath 设置视图路径
func (app *App) SetViewPath(path []string) {
	app.ViewPaths = path
}

// GetViewPath 获取视图路径
func (app *App) GetViewPath() []string {
	return app.ViewPaths
}

// Use 添加中间件
func (app *App) Use(middleware ...define.HandlerFunc) {
	for _, m := range middleware {
		app.Hertz.Use(m)
	}
}

// Run 启动服务器
func (app *App) Run(addr ...string) {
	if len(addr) > 0 {
		app.address = addr[0]
	}
	app.Hertz.Spin()
}
