// 继续 ControllerRegister 的路由匹配和请求处理实现
package register

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/zsy619/yyhertz/framework/config"
	contextenhanced "github.com/zsy619/yyhertz/framework/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// ============= 路由匹配和查找 =============

// serveHTTP 处理HTTP请求的核心方法（简化版）
func (cr *ControllerRegister) serveHTTP(ctx *contextenhanced.Context) {
	// 增加请求计数
	cr.mu.Lock()
	cr.requestCount++
	cr.mu.Unlock()

	// 执行前置过滤器
	if !cr.executeFilters(ctx, BeforeRouter) {
		return
	}

	// 查找路由
	controllerInfo, found := cr.findRouter(ctx)
	if !found {
		cr.handleNotFound(ctx)
		return
	}

	// 执行路由前过滤器
	if !cr.executeFilters(ctx, BeforeExec) {
		return
	}

	// 执行控制器
	cr.executeController(ctx, controllerInfo)

	// 执行后置过滤器
	cr.executeFilters(ctx, AfterExec)
}

// findRouter 查找匹配的路由
func (cr *ControllerRegister) findRouter(ctx *contextenhanced.Context) (*ControllerInfo, bool) {
	if ctx == nil || ctx.RequestContext == nil {
		return nil, false
	}

	method := strings.ToUpper(string(ctx.RequestContext.Method()))
	requestPath := string(ctx.RequestContext.URI().Path())

	// 清理路径
	requestPath = cr.cleanPath(requestPath)

	// 首先在指定HTTP方法的路由树中查找
	if tree := cr.routers[method]; tree != nil {
		if info := cr.matchRoute(tree, requestPath, ctx); info != nil {
			return info, true
		}
	}

	// 然后在ANY路由树中查找
	if tree := cr.routers["ANY"]; tree != nil {
		if info := cr.matchRoute(tree, requestPath, ctx); info != nil {
			return info, true
		}
	}

	return nil, false
}

// matchRoute 在路由树中匹配路由
func (cr *ControllerRegister) matchRoute(tree *ControllerTree, requestPath string, ctx *contextenhanced.Context) *ControllerInfo {
	// 1. 首先尝试固定路由
	if info := tree.fixedRoutes[requestPath]; info != nil {
		return info
	}

	// 2. 尝试正则路由
	for _, info := range tree.regexRoutes {
		if matches := info.routeRegex.FindStringSubmatch(requestPath); matches != nil {
			// 设置路由参数
			cr.setRouteParams(ctx, info, matches[1:])
			return info
		}
	}

	// 3. 尝试通配符路由
	if tree.wildcardRoute != nil {
		return tree.wildcardRoute
	}

	return nil
}

// setRouteParams 设置路由参数
func (cr *ControllerRegister) setRouteParams(ctx *contextenhanced.Context, info *ControllerInfo, matches []string) {
	if len(info.params) == 0 || len(matches) == 0 {
		return
	}

	// 创建参数映射
	params := make([]contextenhanced.Param, 0, len(info.params))

	for i, paramName := range info.params {
		if i < len(matches) {
			params = append(params, contextenhanced.Param{
				Key:   paramName,
				Value: matches[i],
			})
		}
	}

	ctx.Params = params
}

// cleanPath 清理路径
func (cr *ControllerRegister) cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	// 标准化路径
	if p[0] != '/' {
		p = "/" + p
	}

	// 简单的路径清理，移除多余的斜杠
	cleaned := path.Clean(p)

	// 保持尾部斜杠（如果原路径有的话）
	if len(p) > 1 && strings.HasSuffix(p, "/") && !strings.HasSuffix(cleaned, "/") {
		cleaned += "/"
	}

	return cleaned
}

// ============= 控制器执行 =============

// executeController 执行控制器
func (cr *ControllerRegister) executeController(ctx *contextenhanced.Context, info *ControllerInfo) {
	defer func() {
		if err := recover(); err != nil {
			cr.handlePanic(ctx, err)
		}
	}()

	// 获取HTTP方法
	method := strings.ToUpper(string(ctx.RequestContext.Method()))

	// 查找方法信息
	methodInfo := info.methods[method]
	if methodInfo == nil {
		methodInfo = info.methods["ANY"]
	}

	if methodInfo == nil {
		cr.handleMethodNotAllowed(ctx, info)
		return
	}

	// 执行控制器方法
	if info.controllerType != nil {
		// 控制器类型路由
		cr.executeControllerMethod(ctx, info, methodInfo)
	} else {
		// 函数类型路由
		cr.executeFunctionHandler(ctx, methodInfo)
	}
}

// executeControllerMethod 执行控制器方法
func (cr *ControllerRegister) executeControllerMethod(ctx *contextenhanced.Context, info *ControllerInfo, methodInfo *MethodInfo) {
	// 创建控制器实例
	controller := cr.createControllerInstance(info.controllerType)
	if controller == nil {
		cr.handleError(ctx, fmt.Errorf("failed to create controller instance"))
		return
	}

	// 类型断言为IController
	iController, ok := controller.(core.IController)
	if !ok {
		cr.handleError(ctx, fmt.Errorf("controller does not implement IController interface"))
		return
	}

	// 执行控制器生命周期
	cr.executeControllerLifecycle(ctx, iController, methodInfo)
}

// createControllerInstance 创建控制器实例
func (cr *ControllerRegister) createControllerInstance(controllerType reflect.Type) any {
	// 从对象池获取实例
	pooledMap := cr.pool.Get().(map[string]any)
	defer cr.pool.Put(pooledMap)

	typeName := controllerType.String()

	// 检查池中是否有可用实例
	if instance, exists := pooledMap[typeName]; exists {
		delete(pooledMap, typeName) // 从池中移除
		return instance
	}

	// 创建新实例
	if controllerType.Kind() == reflect.Ptr {
		// 指针类型
		return reflect.New(controllerType.Elem()).Interface()
	} else {
		// 值类型
		return reflect.New(controllerType).Interface()
	}
}

// getControllerName 获取控制器名称
func (cr *ControllerRegister) getControllerName(controller core.IController) string {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}
	return controllerType.Name()
}

// executeControllerLifecycle 执行控制器生命周期
func (cr *ControllerRegister) executeControllerLifecycle(ctx *contextenhanced.Context, controller core.IController, methodInfo *MethodInfo) {
	// 1. 初始化控制器（使用Beego兼容的参数名ct）
	controllerName := cr.getControllerName(controller)
	// 检查是否有Controller后缀
	for suffix := range core.ControllerNameSuffixReserved {
		if strings.HasSuffix(controllerName, suffix) {
			controllerName = strings.TrimSuffix(controllerName, suffix)
			break
		}
	}
	actionName := methodInfo.methodName
	controller.Init(ctx, controllerName, actionName, nil)

	// 2. 执行Prepare方法
	controller.Prepare()

	// 3. 执行具体的业务方法
	cr.invokeMethod(controller, methodInfo)

	// 4. 执行Finish方法
	controller.Finish()
}

// invokeMethod 调用方法
func (cr *ControllerRegister) invokeMethod(controller core.IController, methodInfo *MethodInfo) {
	defer func() {
		if err := recover(); err != nil {
			config.Errorf("Method execution panic: %v", err)
		}
	}()

	// 获取方法
	controllerValue := reflect.ValueOf(controller)
	method := controllerValue.MethodByName(methodInfo.methodName)

	if !method.IsValid() {
		config.Errorf("Method %s not found", methodInfo.methodName)
		return
	}

	// 调用方法
	method.Call([]reflect.Value{})
}

// executeFunctionHandler 执行函数处理器
func (cr *ControllerRegister) executeFunctionHandler(ctx *contextenhanced.Context, methodInfo *MethodInfo) {
	defer func() {
		if err := recover(); err != nil {
			cr.handlePanic(ctx, err)
		}
	}()

	// 根据函数签名调用
	funcType := methodInfo.methodType
	funcValue := methodInfo.methodValue

	if !funcValue.IsValid() {
		cr.handleError(ctx, fmt.Errorf("invalid function handler"))
		return
	}

	// 准备参数
	args := cr.prepareFunctionArgs(ctx, funcType)

	// 调用函数
	results := funcValue.Call(args)

	// 处理返回值
	cr.handleFunctionResults(ctx, results)
}

// prepareFunctionArgs 准备函数参数（简化版）
func (cr *ControllerRegister) prepareFunctionArgs(ctx *contextenhanced.Context, funcType reflect.Type) []reflect.Value {
	numIn := funcType.NumIn()
	args := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		paramType := funcType.In(i)

		switch paramType {
		case reflect.TypeOf((*contextenhanced.Context)(nil)):
			// *Context 参数
			args[i] = reflect.ValueOf(ctx)
		default:
			// 其他类型，尝试从请求中解析
			args[i] = reflect.Zero(paramType)
		}
	}

	return args
}

// handleFunctionResults 处理函数返回值
func (cr *ControllerRegister) handleFunctionResults(ctx *contextenhanced.Context, results []reflect.Value) {
	if len(results) == 0 {
		return
	}

	for _, result := range results {
		if !result.IsValid() {
			continue
		}

		// 检查是否为error类型
		if result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !result.IsNil() {
				err := result.Interface().(error)
				cr.handleError(ctx, err)
				return
			}
		}

		// 其他类型作为响应内容
		if result.CanInterface() {
			ctx.Output.JSON(result.Interface(), false, true)
		}
	}
}

// ============= 错误处理 =============

// handleNotFound 处理404错误
func (cr *ControllerRegister) handleNotFound(ctx *contextenhanced.Context) {
	ctx.Output.SetStatus(404)
	ctx.Output.JSON(map[string]any{
		"error": "Not Found",
		"code":  404,
		"path":  string(ctx.RequestContext.URI().Path()),
	}, false, true)
}

// handleMethodNotAllowed 处理405错误
func (cr *ControllerRegister) handleMethodNotAllowed(ctx *contextenhanced.Context, info *ControllerInfo) {
	// 获取支持的方法列表
	allowedMethods := make([]string, 0, len(info.methods))
	for method := range info.methods {
		if method != "ANY" {
			allowedMethods = append(allowedMethods, method)
		}
	}

	ctx.Output.Header("Allow", strings.Join(allowedMethods, ", "))
	ctx.Output.SetStatus(405)
	ctx.Output.JSON(map[string]any{
		"error":           "Method Not Allowed",
		"code":            405,
		"allowed_methods": allowedMethods,
	}, false, true)
}

// handlePanic 处理panic
func (cr *ControllerRegister) handlePanic(ctx *contextenhanced.Context, err any) {
	config.Errorf("Request panic: %v", err)

	ctx.Output.SetStatus(500)
	ctx.Output.JSON(map[string]any{
		"error": "Internal Server Error",
		"code":  500,
	}, false, true)
}

// handleError 处理一般错误
func (cr *ControllerRegister) handleError(ctx *contextenhanced.Context, err error) {
	config.Errorf("Request error: %v", err)

	ctx.Output.SetStatus(500)
	ctx.Output.JSON(map[string]any{
		"error": err.Error(),
		"code":  500,
	}, false, true)
}

// ============= 过滤器执行 =============

// executeFilters 执行过滤器
func (cr *ControllerRegister) executeFilters(ctx *contextenhanced.Context, pos int) bool {
	cr.mu.RLock()
	filters := cr.filters
	cr.mu.RUnlock()

	for _, filterRouter := range filters {
		if cr.matchFilterPattern(filterRouter.pattern, string(ctx.RequestContext.URI().Path())) {
			// 创建过滤器链
			chain := &FilterChain{
				filters: []FilterFunc{filterRouter.filterFunc},
				index:   0,
			}

			// 执行过滤器
			filterRouter.filterFunc(ctx, chain)

			// 检查是否应该继续处理
			if filterRouter.returnOnOutput && ctx.ResponseWriter.Written() {
				return false
			}
		}
	}

	return true
}

// matchFilterPattern 匹配过滤器模式
func (cr *ControllerRegister) matchFilterPattern(pattern, path string) bool {
	// 简单的模式匹配，支持通配符
	if pattern == "*" || pattern == "/*" {
		return true
	}

	if pattern == path {
		return true
	}

	// 支持前缀匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// Next 执行过滤器链中的下一个过滤器（简化版）
func (chain *FilterChain) Next(ctx *contextenhanced.Context) {
	chain.index++
	if chain.index < len(chain.filters) {
		chain.filters[chain.index](ctx, chain)
	}
}

// ============= 统计和监控 =============

// GetRequestCount 获取请求计数
func (cr *ControllerRegister) GetRequestCount() int64 {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.requestCount
}

// GetRouteCount 获取路由数量
func (cr *ControllerRegister) GetRouteCount() map[string]int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	counts := make(map[string]int)
	for method, tree := range cr.routers {
		count := len(tree.fixedRoutes) + len(tree.regexRoutes)
		if tree.wildcardRoute != nil {
			count++
		}
		counts[method] = count
	}

	return counts
}

// ListRoutes 列出所有路由
func (cr *ControllerRegister) ListRoutes() map[string][]string {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	routes := make(map[string][]string)

	for method, tree := range cr.routers {
		routeList := make([]string, 0)

		// 固定路由
		for pattern := range tree.fixedRoutes {
			routeList = append(routeList, pattern)
		}

		// 正则路由
		for _, info := range tree.regexRoutes {
			routeList = append(routeList, info.pattern)
		}

		// 通配符路由
		if tree.wildcardRoute != nil {
			routeList = append(routeList, tree.wildcardRoute.pattern)
		}

		routes[method] = routeList
	}

	return routes
}
