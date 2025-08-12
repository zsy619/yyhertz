package core

import (
	"context"
	"fmt"
	"html/template"
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
	"github.com/zsy619/yyhertz/framework/constant"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
	"github.com/zsy619/yyhertz/framework/view"
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

// FilterFunc 过滤器函数类型
type FilterFunc = func(*contextenhanced.Context)

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
	Pattern  string     // 路径模式 (支持通配符)
	Position int        // 过滤器位置
	Filter   FilterFunc // 过滤器函数
	Enabled  bool       // 是否启用
	Priority int        // 优先级
}

// 位置映射到中间件层级已移除 - 现在使用统一常量和转换函数

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
	StaticPaths   map[string]string // URL路径 -> 本地路径映射
	startTime     time.Time
	address       string
	loggerManager *config.LoggerManager

	// 全局模板函数管理
	globalFuncMap template.FuncMap
	funcMapMutex  sync.RWMutex

	// 过滤器管理
	filters      map[int][]*FilterPattern // 按位置分组的过滤器
	filtersMutex sync.RWMutex             // 过滤器读写锁
	nextFilterID int64                    // 下一个过滤器ID (用于排序)
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

	app := &App{
		Hertz:         h,                                        // 使用Hertz服务器实例
		ViewPath:      "./views",                                // 默认视图路径
		StaticPaths:   map[string]string{"/static": "./static"}, // 默认静态文件路径映射
		startTime:     time.Now(),                               // 记录应用启动时间
		address:       fmt.Sprintf("%s:%d", host, port),         // 应用监听地址
		loggerManager: loggerManager,                            // 日志管理器

		// 初始化全局模板函数映射
		globalFuncMap: make(template.FuncMap),

		// 初始化过滤器管理
		filters:      make(map[int][]*FilterPattern),
		nextFilterID: 0,
	}

	// 配置视图路径
	app.SetViewPath("./views")
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
// 参数：localDir - 静态文件本地目录（相对应用所在目录）
//
//	urlPath - URL路径（可选），如果不提供则自动推导
//
// 示例：SetStaticPath("public", "/static") 或 SetStaticPath("public")
func (app *App) SetStaticPath(localDir string, urlPath ...string) {
	if app.StaticPaths == nil {
		app.StaticPaths = make(map[string]string)
	}

	// 确定URL路径
	var targetUrlPath string
	if len(urlPath) > 0 && urlPath[0] != "" {
		targetUrlPath = urlPath[0]
	} else {
		// 自动推导：移除 "./" 前缀，确保以 "/" 开头
		cleanDir := strings.TrimLeft(strings.TrimPrefix(localDir, "./"), "/")
		if cleanDir == "" {
			targetUrlPath = "/static" // 默认URL路径
		} else {
			targetUrlPath = "/" + cleanDir
		}
	}

	// 确保URL路径以/开头
	if !strings.HasPrefix(targetUrlPath, "/") {
		targetUrlPath = "/" + targetUrlPath
	}

	// 只有当路径不存在或者发生变化时才注册
	if existing, exists := app.StaticPaths[targetUrlPath]; !exists || existing != localDir {
		app.StaticPaths[targetUrlPath] = localDir
		// 注册静态文件路由
		app.Static(targetUrlPath, localDir)
	}
}

// SetStaticPaths 设置多个静态文件路径映射
func (app *App) SetStaticPaths(pathMap map[string]string) {
	app.StaticPaths = make(map[string]string)
	for localPath, urlPath := range pathMap {
		app.SetStaticPath(localPath, urlPath)
	}
}

// AddStaticPath 添加单个静态路径映射
func (app *App) AddStaticPath(localPath, urlPath string) {
	app.SetStaticPath(localPath, urlPath)
}

// AddStaticPaths 添加多个静态路径映射
func (app *App) AddStaticPaths(pathMap map[string]string) {
	app.SetStaticPaths(pathMap)
}

// GetStaticPath 获取默认静态文件路径（向后兼容）
func (app *App) GetStaticPath() string {
	if path, exists := app.StaticPaths["/static"]; exists {
		return path
	}
	// 返回第一个路径作为默认值
	for _, path := range app.StaticPaths {
		return path
	}
	return "./static"
}

// GetStaticPaths 获取所有静态路径映射
func (app *App) GetStaticPaths() map[string]string {
	return app.StaticPaths
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

// Fatal 全局致命错误日志
func (app *App) LogFatal(args ...any) {
	config.Fatal(args...)
}

// Fatalf 全局格式化致命错误日志
func (app *App) LogFatalf(format string, args ...any) {
	config.Fatalf(format, args...)
}

// Panic 全局panic日志
func (app *App) LogPanic(args ...any) {
	config.Panic(args...)
}

// Panicf 全局格式化panic日志
func (app *App) LogPanicf(format string, args ...any) {
	config.Panicf(format, args...)
}

// ============= 路由注册方法 =============

// AutoRouters 自动注册多个控制器路由（根据控制器方法名自动推导路由）
func (app *App) AutoRouters(controllers ...IController) *App {
	return app.AutoRoutersPrefix("", controllers...)
}

// Include 向后兼容的别名方法，自动注册多个控制器路由
func (app *App) Include(controllers ...IController) *App {
	return app.AutoRouters(controllers...)
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

		// 初始化增强上下文
		enhancedCtx := contextenhanced.NewContext(c)

		// 执行 BeforeStatic 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeStatic)
		if enhancedCtx.IsAborted() {
			return
		}

		// 执行 BeforeRouter 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeRouter)
		if enhancedCtx.IsAborted() {
			return
		}

		// 初始化控制器
		controllerName := controller.GetControllerName() // 使用修复后的方法
		methodName := method.Name
		controller.Init(enhancedCtx, controllerName, methodName, app)

		// 设置控制器上下文（如果控制器有Ctx字段）
		app.setControllerContext(controller, c)

		// 执行 BeforeExec 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeExec)
		if enhancedCtx.IsAborted() {
			return
		}

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

		// 执行 AfterExec 过滤器
		app.ExecuteFilters(enhancedCtx, AfterExec)

		// 执行后置处理
		controller.Finish()

		// 执行 FinishRouter 过滤器
		app.ExecuteFilters(enhancedCtx, FinishRouter)
	}
}

// createMethodHandler 创建方法处理函数
func (app *App) createMethodHandler(controller IController, methodName string) HandlerFunc {
	return func(ctx context.Context, c *RequestContext) {
		// 初始化增强上下文
		enhancedCtx := contextenhanced.NewContext(c)

		// 执行 BeforeStatic 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeStatic)
		if enhancedCtx.IsAborted() {
			return
		}

		// 执行 BeforeRouter 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeRouter)
		if enhancedCtx.IsAborted() {
			return
		}

		// 初始化控制器
		controllerName := app.getControllerName(controller)
		controller.Init(enhancedCtx, controllerName, methodName, app)

		// 设置控制器上下文
		app.setControllerContext(controller, c)

		// 执行 BeforeExec 过滤器
		app.ExecuteFilters(enhancedCtx, BeforeExec)
		if enhancedCtx.IsAborted() {
			return
		}

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

		// 执行 AfterExec 过滤器
		app.ExecuteFilters(enhancedCtx, AfterExec)

		// 执行后置处理
		controller.Finish()

		// 执行 FinishRouter 过滤器
		app.ExecuteFilters(enhancedCtx, FinishRouter)
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

// ============= 模板函数管理方法 =============

// AddFuncMap 添加全局模板函数
// 参数：name - 函数名字符串，fn - 函数实现
// 示例：AddFuncMap("containString", tool.ContainString)
func (app *App) AddFuncMap(name string, fn any) {
	app.funcMapMutex.Lock()
	defer app.funcMapMutex.Unlock()

	// 添加到应用级别的全局模板函数映射
	app.globalFuncMap[name] = fn

	// 同时添加到view引擎的全局存储中
	view.AddGlobalFunction(name, fn)

	app.LogInfof("Template function registered: %s", name)
}

// GetGlobalFuncMap 获取全局模板函数映射（只读副本）
func (app *App) GetGlobalFuncMap() template.FuncMap {
	app.funcMapMutex.RLock()
	defer app.funcMapMutex.RUnlock()

	// 创建副本以避免并发修改
	funcMapCopy := make(template.FuncMap, len(app.globalFuncMap))
	for name, fn := range app.globalFuncMap {
		funcMapCopy[name] = fn
	}

	return funcMapCopy
}

// RemoveFuncMap 移除全局模板函数
func (app *App) RemoveFuncMap(name string) {
	app.funcMapMutex.Lock()
	defer app.funcMapMutex.Unlock()

	delete(app.globalFuncMap, name)

	// 同时从view引擎的全局存储中移除
	view.RemoveGlobalFunction(name)

	app.LogInfof("Template function removed: %s", name)
}

// ListFuncMap 列出所有已注册的模板函数名称
func (app *App) ListFuncMap() []string {
	app.funcMapMutex.RLock()
	defer app.funcMapMutex.RUnlock()

	names := make([]string, 0, len(app.globalFuncMap))
	for name := range app.globalFuncMap {
		names = append(names, name)
	}

	return names
}

// ============= 过滤器管理方法 =============

// InsertFilter 插入过滤器到指定位置
// 参数：pattern - 路径模式 (支持通配符 *)
//
//	position - 过滤器位置 (BeforeStatic, BeforeRouter, BeforeExec, AfterExec, FinishRouter)
//	filter - 过滤器函数
//	params - 可选参数 (第一个bool值表示是否启用，默认true)
func (app *App) InsertFilter(pattern string, position int, filter FilterFunc, params ...bool) {
	// 验证位置参数
	if !constant.IsValidFilterPosition(position) {
		app.LogErrorf("Invalid filter position: %d", position)
		return
	}

	// 处理可选参数
	enabled := true
	if len(params) > 0 {
		enabled = params[0]
	}

	app.filtersMutex.Lock()
	defer app.filtersMutex.Unlock()

	// 创建过滤器模式
	filterPattern := &FilterPattern{
		Pattern:  pattern,
		Position: position,
		Filter:   filter,
		Enabled:  enabled,
		Priority: int(app.nextFilterID), // 使用ID作为优先级，保证插入顺序
	}

	// 添加到对应位置的过滤器列表
	app.filters[position] = append(app.filters[position], filterPattern)
	app.nextFilterID++

	app.LogInfof("Filter inserted: pattern=%s, position=%d", pattern, position)
}

// RemoveFilter 移除指定模式和位置的过滤器
func (app *App) RemoveFilter(pattern string, position int) bool {
	app.filtersMutex.Lock()
	defer app.filtersMutex.Unlock()

	filters := app.filters[position]
	for i, filter := range filters {
		if filter.Pattern == pattern {
			// 从切片中移除
			app.filters[position] = append(filters[:i], filters[i+1:]...)
			app.LogInfof("Filter removed: pattern=%s, position=%d", pattern, position)
			return true
		}
	}

	return false
}

// ListFilters 列出指定位置的所有过滤器
func (app *App) ListFilters(position int) []*FilterPattern {
	app.filtersMutex.RLock()
	defer app.filtersMutex.RUnlock()

	filters := app.filters[position]
	// 返回副本，避免并发修改
	result := make([]*FilterPattern, len(filters))
	copy(result, filters)
	return result
}

// GetAllFilters 获取所有位置的过滤器
func (app *App) GetAllFilters() map[int][]*FilterPattern {
	app.filtersMutex.RLock()
	defer app.filtersMutex.RUnlock()

	// 创建深度副本
	result := make(map[int][]*FilterPattern)
	for position, filters := range app.filters {
		result[position] = make([]*FilterPattern, len(filters))
		copy(result[position], filters)
	}

	return result
}

// matchPattern 检查路径是否匹配模式 (支持 * 通配符)
func (app *App) matchPattern(pattern, path string) bool {
	// 简单的通配符匹配实现
	if pattern == "*" || pattern == "/*" {
		return true
	}

	// 精确匹配
	if pattern == path {
		return true
	}

	// 通配符匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}

	// 中间通配符支持
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(path, parts[0]) && strings.HasSuffix(path, parts[1])
		}
	}

	return false
}

// ExecuteFilters 执行指定位置的过滤器
func (app *App) ExecuteFilters(ctx *contextenhanced.Context, position int) {
	app.filtersMutex.RLock()
	filters := app.filters[position]
	app.filtersMutex.RUnlock()

	if len(filters) == 0 {
		return
	}

	// 获取请求路径
	path := string(ctx.Request.Path())

	// 执行匹配的过滤器
	for _, filter := range filters {
		if !filter.Enabled {
			continue
		}

		// 检查路径是否匹配
		if app.matchPattern(filter.Pattern, path) {
			// 执行过滤器
			filter.Filter(ctx)

			// 如果请求被中止，停止执行后续过滤器
			if ctx.IsAborted() {
				break
			}
		}
	}
}
