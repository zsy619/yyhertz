package core

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	hertzlogrus "github.com/hertz-contrib/logger/logrus"

	"github.com/zsy619/yyhertz/framework/config"
	contextenhanced "github.com/zsy619/yyhertz/framework/context"
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
type HandlerFunc = func(context.Context, *RequestContext)

// AdaptHandler 将HandlerFunc适配为app.HandlerFunc
func AdaptHandler(handler HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		handler(ctx, (*RequestContext)(c))
	}
}

// App 应用结构（精简版，只保留核心功能）
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

	// 创建Hertz服务器实例
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
	app.GET("/health", func(c context.Context, ctx *RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// ping路由
	app.GET("/ping", func(c context.Context, ctx *RequestContext) {
		ctx.JSON(consts.StatusOK, map[string]string{"message": "pong"})
	})
}

// SetViewPath 设置视图路径
func (app *App) SetViewPath(path string) {
	app.ViewPath = path
}

// GetViewPath 获取视图路径
func (app *App) GetViewPath() string {
	return app.ViewPath
}

// SetStaticPath 设置静态文件路径
func (app *App) SetStaticPath(path string) {
	app.StaticPath = path
	app.Static("/static", path)
}

// GetStaticPath 获取静态文件路径
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

// 基础日志方法
func (app *App) LogInfof(format string, args ...any) {
	config.Infof(format, args...)
}

func (app *App) LogInfo(args ...any) {
	config.Info(args...)
}

func (app *App) LogErrorf(format string, args ...any) {
	config.Errorf(format, args...)
}

func (app *App) LogError(args ...any) {
	config.Error(args...)
}

func (app *App) LogWarnf(format string, args ...any) {
	config.Warnf(format, args...)
}

func (app *App) LogWarn(args ...any) {
	config.Warn(args...)
}

func (app *App) LogDebugf(format string, args ...any) {
	config.Debugf(format, args...)
}

func (app *App) LogDebug(args ...any) {
	config.Debug(args...)
}

// GetLoggerWithContext 获取带上下文信息的logger
func (app *App) GetLoggerWithContext(ctx *RequestContext) *hertzlogrus.Logger {
	return config.GetGlobalLogger().GetLogger()
}

// ============= 路由注册方法 =============

// AutoRouters 自动注册多个控制器路由（根据控制器方法名自动推导路由）
func (app *App) AutoRouters(controllers ...IController) *App {
	return app.AutoRoutersPrefix("", controllers...)
}

// AutoRoutersPrefix 自动注册多个控制器路由，使用指定的路径前缀
func (app *App) AutoRoutersPrefix(prefix string, ctrls ...IController) *App {
	for _, ctrl := range ctrls {
		app.registerAutoRoutes(prefix, ctrl)
	}
	return app
}

// AutoRouter 自动注册单个控制器
func (app *App) AutoRouter(ctrl IController) *App {
	return app.AutoRouterPrefix("", ctrl)
}

// 注册单个控制器（无routes时自动注册，有routes时手动注册）
func (app *App) AutoRouterPrefix(prefix string, ctrl IController) *App {
	app.registerManualRoutes(prefix, ctrl)
	return app
}

// Router 手动注册控制器路由
func (app *App) Router(ctrl IController, routes ...string) *App {
	return app.RouterPrefix("", ctrl, routes...)
}

// RouterPrefix 手动注册控制器路由
func (app *App) RouterPrefix(prefix string, ctrl IController, routes ...string) *App {
	if len(routes) == 0 {
		return app
	}
	app.registerManualRoutes(prefix, ctrl, routes...)
	return app
}

// ============= 向后兼容的别名方法 =============

// registerAutoRoutes 自动注册控制器路由
func (app *App) registerAutoRoutes(basePath string, controller IController) {
	// 确保控制器实例正确设置（提前初始化）
	if method := reflect.ValueOf(controller).MethodByName("SetControllerInstance"); method.IsValid() {
		method.Call([]reflect.Value{reflect.ValueOf(controller)})
	}

	// 使用反射获取控制器类型信息
	reflectVal := reflect.ValueOf(controller)
	rt := reflectVal.Type() // 获取指针类型的方法，而不是值类型

	// 从控制器名称推导基础路径
	controllerName := rt.Elem().Name() // 获取指针指向的类型名称
	if basePath == "" {
		basePath = "/"
	}
	for suffix := range ControllerNameSuffixReserved {
		if strings.HasSuffix(controllerName, suffix) {
			name := strings.TrimSuffix(controllerName, suffix)
			// if name != "Home" && name != "Index" {
			name = strings.ToLower(name)
			if basePath == "/" {
				basePath += name
			} else {
				basePath = path.Join(basePath, name)
			}
			// }
			break
		}
	}

	// 遍历所有公共方法
	for i := 0; i < rt.NumMethod(); i++ {
		method := rt.Method(i)
		methodName := method.Name

		// 跳过生命周期方法和BaseController方法
		if _, ok := ReservedMethods[methodName]; ok {
			continue
		}

		// 根据方法名前缀确定HTTP方法
		httpMethod := "ANY" // 默认ANY
		actionName := methodName

		switch {
		case strings.HasPrefix(methodName, "Get"):
			httpMethod = "GET"
			actionName = strings.TrimPrefix(methodName, "Get")
		case strings.HasPrefix(methodName, "Post"):
			httpMethod = "POST"
			actionName = strings.TrimPrefix(methodName, "Post")
		case strings.HasPrefix(methodName, "Put"):
			httpMethod = "PUT"
			actionName = strings.TrimPrefix(methodName, "Put")
		case strings.HasPrefix(methodName, "Delete"):
			httpMethod = "DELETE"
			actionName = strings.TrimPrefix(methodName, "Delete")
		case strings.HasPrefix(methodName, "Patch"):
			httpMethod = "PATCH"
			actionName = strings.TrimPrefix(methodName, "Patch")
		case strings.HasPrefix(methodName, "Head"):
			httpMethod = "HEAD"
			actionName = strings.TrimPrefix(methodName, "Head")
		case strings.HasPrefix(methodName, "Options"):
			httpMethod = "OPTIONS"
			actionName = strings.TrimPrefix(methodName, "Options")
		}

		// 构建路由路径
		routePath := basePath
		if actionName != "" && actionName != "Index" {
			if !strings.HasSuffix(routePath, "/") {
				routePath += "/"
			}
			routePath += strings.ToLower(actionName)
		}

		// 为根路径特殊处理
		if routePath == "//" {
			routePath = "/"
		}

		// 创建处理函数
		handler := app.createControllerHandler(controller, method)

		// 注册路由
		app.registerRoute(httpMethod, routePath, handler)
	}
}

// registerManualRoutes 手动注册路由
func (app *App) registerManualRoutes(basePath string, controller IController, routes ...string) {
	t := reflect.TypeOf(controller)                       // 返回 *controllers.UserController
	controllerName := strings.TrimPrefix(t.String(), "*") // 得到 "controllers.UserController"
	controllerName = strings.TrimSuffix(controllerName, "Controller")
	fmt.Printf("Registering routes for controller: %s\n", controllerName)

	for i := 0; i < len(routes); i += 2 {
		if i+1 >= len(routes) {
			break
		}

		methodName := routes[i]
		routeSpec := routes[i+1]

		// 解析路由规格: "GET:/path" 或 "/path" 或 "*:/path"
		httpMethod := "ANY"
		routePath := routeSpec

		if colonIndex := strings.Index(routeSpec, ":"); colonIndex != -1 {
			httpMethod = routeSpec[:colonIndex]
			if httpMethod == "*" { // 兼容旧格式的路由语法: *:path
				httpMethod = "ANY"
			}
			routePath = routeSpec[colonIndex+1:]
		}

		// 确保路由路径以基础路径开头
		if !strings.HasPrefix(routePath, basePath) {
			routePath = basePath + routePath
		}

		// 获取控制器方法
		reflectVal := reflect.ValueOf(controller)
		method := reflectVal.MethodByName(methodName)

		if !method.IsValid() {
			app.LogErrorf("Method %s not found in controller", methodName)
			continue
		}

		// 创建处理函数
		handler := app.createMethodHandler(controller, methodName)

		// 注册路由
		app.registerRoute(httpMethod, routePath, handler)
	}
}

// getControllerName 获取控制器名称
func (app *App) getControllerName(controller IController) string {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}
	controllerName := controllerType.Name()
	for suffix := range ControllerNameSuffixReserved {
		if strings.HasSuffix(controllerName, suffix) {
			controllerName = strings.TrimSuffix(controllerName, suffix)
			break
		}
	}
	return controllerName
}

// createControllerHandler 创建控制器处理函数
func (app *App) createControllerHandler(controller IController, method reflect.Method) HandlerFunc {
	return func(ctx context.Context, c *RequestContext) {
		// 确保控制器实例正确设置（关键修复）
		if method := reflect.ValueOf(controller).MethodByName("SetControllerInstance"); method.IsValid() {
			method.Call([]reflect.Value{reflect.ValueOf(controller)})
		}

		// 初始化控制器
		enhancedCtx := contextenhanced.NewContext(c)
		controllerName := controller.GetControllerName() // 使用修复后的方法
		methodName := method.Name
		controller.Init(enhancedCtx, controllerName, methodName, app)

		// 设置控制器上下文（如果控制器有Ctx字段）
		app.setControllerContext(controller, c)

		// 执行前置处理
		controller.Prepare()

		// 执行具体方法
		methodValue := reflect.ValueOf(controller).MethodByName(method.Name)
		if methodValue.IsValid() {
			// 根据方法签名调用
			methodType := methodValue.Type()
			if methodType.NumIn() == 2 {
				// 方法签名: func(context.Context, *RequestContext)
				methodValue.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(c),
				})
			} else if methodType.NumIn() == 0 {
				// 方法签名: func()
				methodValue.Call([]reflect.Value{})
			}
		}

		// 执行后置处理
		controller.Finish()
	}
}

// createMethodHandler 创建方法处理函数
func (app *App) createMethodHandler(controller IController, methodName string) HandlerFunc {
	return func(ctx context.Context, c *RequestContext) {
		// 初始化控制器
		enhancedCtx := contextenhanced.NewContext(c)
		controllerName := app.getControllerName(controller)
		controller.Init(enhancedCtx, controllerName, methodName, app)

		// 设置控制器上下文
		app.setControllerContext(controller, c)

		// 执行前置处理
		controller.Prepare()

		// 执行具体方法
		methodValue := reflect.ValueOf(controller).MethodByName(methodName)
		if methodValue.IsValid() {
			methodType := methodValue.Type()
			if methodType.NumIn() == 2 {
				methodValue.Call([]reflect.Value{
					reflect.ValueOf(ctx),
					reflect.ValueOf(c),
				})
			} else if methodType.NumIn() == 0 {
				methodValue.Call([]reflect.Value{})
			}
		}

		// 执行后置处理
		controller.Finish()
	}
}

// setControllerContext 设置控制器上下文（重构后版本）
func (app *App) setControllerContext(controller IController, ctx *RequestContext) {
	// 创建增强的Context
	enhancedCtx := contextenhanced.NewContext(ctx)

	// 使用反射设置控制器的Ctx字段
	reflectVal := reflect.ValueOf(controller)
	if reflectVal.Kind() == reflect.Ptr {
		reflectVal = reflectVal.Elem()
	}

	// 查找Ctx字段
	if reflectVal.Kind() == reflect.Struct {
		ctxField := reflectVal.FieldByName("Ctx")
		if ctxField.IsValid() && ctxField.CanSet() {
			ctxField.Set(reflect.ValueOf(enhancedCtx))
		}

		// 也尝试设置BaseController字段的Ctx
		baseField := reflectVal.FieldByName("BaseController")
		if baseField.IsValid() && baseField.Kind() == reflect.Ptr {
			baseController := baseField.Elem()
			if baseController.IsValid() {
				baseCtxField := baseController.FieldByName("Ctx")
				if baseCtxField.IsValid() && baseCtxField.CanSet() {
					baseCtxField.Set(reflect.ValueOf(enhancedCtx))
				}
			}
		}
	}
}

// registerRoute 注册路由到应用
func (app *App) registerRoute(method, path string, handler HandlerFunc) {
	switch strings.ToUpper(method) {
	case "GET":
		app.GET(path, handler)
	case "POST":
		app.POST(path, handler)
	case "PUT":
		app.PUT(path, handler)
	case "DELETE":
		app.DELETE(path, handler)
	case "PATCH":
		app.PATCH(path, handler)
	case "HEAD":
		app.HEAD(path, handler)
	case "OPTIONS":
		app.OPTIONS(path, handler)
	default:
		app.Any(path, handler)
	}

	app.LogInfof("Route registered: %s %s", method, path)
}
