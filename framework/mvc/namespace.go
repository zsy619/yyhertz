package mvc

import (
	"strings"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// NamespaceFunc 定义命名空间配置函数类型
type NamespaceFunc func(*Namespace)

// Namespace 命名空间结构，类似Beego的Namespace
type Namespace struct {
	prefix      string
	controllers []controllerInfo
	routers     []routerInfo
	namespaces  []*Namespace
	middlewares []core.HandlerFunc
}

type controllerInfo struct {
	controller core.IController
	autoRoute  bool
}

type routerInfo struct {
	path       string
	controller core.IController
	method     string
}

// NewNamespace 创建新的命名空间（类似beego.NewNamespace）
func NewNamespace(prefix string, funcs ...NamespaceFunc) *Namespace {
	ns := &Namespace{
		prefix:      prefix,
		controllers: make([]controllerInfo, 0),
		routers:     make([]routerInfo, 0),
		namespaces:  make([]*Namespace, 0),
		middlewares: make([]core.HandlerFunc, 0),
	}

	// 执行配置函数
	for _, fn := range funcs {
		fn(ns)
	}

	return ns
}

// NSAutoRouter 自动路由注册（类似beego.NSAutoRouter）
func NSAutoRouter(ctrl core.IController) NamespaceFunc {
	return func(ns *Namespace) {
		ns.controllers = append(ns.controllers, controllerInfo{
			controller: ctrl,
			autoRoute:  true,
		})
	}
}

// NSRouter 手动路由映射（类似beego.NSRouter）
func NSRouter(path string, ctrl core.IController, method string) NamespaceFunc {
	return func(ns *Namespace) {
		ns.routers = append(ns.routers, routerInfo{
			path:       path,
			controller: ctrl,
			method:     method,
		})
	}
}

// NSNamespace 嵌套命名空间（类似beego.NSNamespace）
func NSNamespace(prefix string, funcs ...NamespaceFunc) NamespaceFunc {
	return func(ns *Namespace) {
		subNs := NewNamespace(prefix, funcs...)
		ns.namespaces = append(ns.namespaces, subNs)
	}
}

// NSMiddleware 添加命名空间中间件
func NSMiddleware(middlewares ...core.HandlerFunc) NamespaceFunc {
	return func(ns *Namespace) {
		ns.middlewares = append(ns.middlewares, middlewares...)
	}
}

// Register 将命名空间注册到应用（内部方法）
func (ns *Namespace) Register(app *core.App) {
	// 注册自动路由控制器
	for _, ctrl := range ns.controllers {
		if ctrl.autoRoute {
			app.AutoRouterPrefix(ns.prefix, ctrl.controller)
		}
	}

	// 注册手动路由
	for _, router := range ns.routers {
		ns.registerRouter(app, router)
	}

	// 递归注册子命名空间
	for _, subNs := range ns.namespaces {
		// 构建嵌套路径
		fullPrefix := ns.prefix
		if !strings.HasSuffix(fullPrefix, "/") {
			fullPrefix += "/"
		}
		fullPrefix += strings.TrimPrefix(subNs.prefix, "/")

		// 创建子命名空间副本，更新前缀
		subNsCopy := &Namespace{
			prefix:      fullPrefix,
			controllers: subNs.controllers,
			routers:     subNs.routers,
			namespaces:  subNs.namespaces,
			middlewares: append(ns.middlewares, subNs.middlewares...), // 继承父级中间件
		}

		subNsCopy.Register(app)
	}
}

// registerRouter 注册单个路由
func (ns *Namespace) registerRouter(app *core.App, router routerInfo) {
	// 解析方法规格："*:MethodName" 或 "GET:MethodName" 或 "MethodName"
	var httpMethod, methodName string

	if strings.Contains(router.method, ":") {
		parts := strings.SplitN(router.method, ":", 2)
		httpMethod = strings.ToUpper(parts[0])
		methodName = parts[1]

		if httpMethod == "*" {
			httpMethod = "ANY"
		}
	} else {
		// 默认为ANY方法
		httpMethod = "ANY"
		methodName = router.method
	}

	// 使用手动路由注册，传递prefix作为basePath，router.path作为相对路径
	routeSpec := httpMethod + ":" + router.path
	app.RouterPrefix(ns.prefix, router.controller, methodName, routeSpec)
}

// GetPrefix 获取命名空间前缀
func (ns *Namespace) GetPrefix() string {
	return ns.prefix
}

// GetControllers 获取控制器列表
func (ns *Namespace) GetControllers() []core.IController {
	var controllers []core.IController
	for _, ctrl := range ns.controllers {
		controllers = append(controllers, ctrl.controller)
	}
	return controllers
}

// GetRouters 获取路由信息
func (ns *Namespace) GetRouters() []string {
	var routes []string
	for _, router := range ns.routers {
		routes = append(routes, router.path+" -> "+router.method)
	}
	return routes
}

// GetNamespaces 获取子命名空间
func (ns *Namespace) GetNamespaces() []*Namespace {
	return ns.namespaces
}
