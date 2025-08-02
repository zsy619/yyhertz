package mvc

// AutoRouters 自动注册多个控制器路由（根据控制器方法名自动推导路由）
func AutoRouters(controllers ...IController) *App {
	return HertzApp.AutoRouters(controllers...)
}

// AutoRoutersPrefix 自动注册多个控制器路由，使用指定的路径前缀
func AutoRoutersPrefix(prefix string, ctrls ...IController) *App {
	return HertzApp.AutoRoutersPrefix(prefix, ctrls...)
}

// AutoRouter 自动注册单个控制器
func AutoRouter(ctrl IController) *App {
	return HertzApp.AutoRouter(ctrl)
}

// 注册单个控制器（无routes时自动注册，有routes时手动注册）
func AutoRouterPrefix(prefix string, ctrl IController) *App {
	return HertzApp.AutoRouterPrefix(prefix, ctrl)
}

// ManualRouter 手动注册控制器路由
func ManualRouter(ctrl IController, routes ...string) *App {
	return HertzApp.Router(ctrl, routes...)
}

// ManualRouterPrefix 手动注册控制器路由
func ManualRouterPrefix(prefix string, ctrl IController, routes ...string) *App {
	return HertzApp.RouterPrefix(prefix, ctrl, routes...)
}

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

// AddNamespace 添加命名空间到全局应用（类似beego.AddNamespace）
func AddNamespace(ns *Namespace) {
	if HertzApp != nil {
		ns.Register(HertzApp)
	}
}
