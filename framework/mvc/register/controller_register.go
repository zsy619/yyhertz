// Package register 实现 Beego 风格的 ControllerRegister
// 提供完整的路由注册、控制器管理和请求处理功能
package register

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// ============= 路由相关结构定义 =============

// ControllerRegister Beego 风格的控制器注册器
type ControllerRegister struct {
	routers      map[string]*ControllerTree // HTTP方法到路由树的映射
	enablePolicy bool                       // 是否启用策略
	policies     map[string]*FilterFunc     // 策略列表
	filters      []*FilterRouter            // 过滤器列表
	pool         sync.Pool                  // 控制器实例池

	// 配置选项
	enableCompress   bool  // 是否启用压缩
	enableGzip       bool  // 是否启用Gzip
	maxMemory        int64 // 最大内存使用
	enableAutoPrefix bool  // 是否启用自动前缀

	// 统计信息
	requestCount int64        // 请求计数
	mu           sync.RWMutex // 读写锁
}

// ControllerTree 控制器路由树
type ControllerTree struct {
	fixedRoutes   map[string]*ControllerInfo // 固定路由
	regexRoutes   []*ControllerInfo          // 正则路由
	wildcardRoute *ControllerInfo            // 通配符路由
}

// ControllerInfo 控制器信息
type ControllerInfo struct {
	pattern        string                 // 路由模式
	controllerType reflect.Type           // 控制器类型
	methods        map[string]*MethodInfo // HTTP方法到方法信息的映射
	routeRegex     *regexp.Regexp         // 路由正则表达式
	params         []string               // 参数名列表
	filters        []*FilterFunc          // 过滤器列表

	// 控制器配置
	initialize bool   // 是否已初始化
	runMethod  string // 运行方法
	actionName string // 动作名称
}

// MethodInfo 方法信息
type MethodInfo struct {
	methodName  string        // 方法名
	httpMethod  string        // HTTP方法
	methodType  reflect.Type  // 方法类型
	methodValue reflect.Value // 方法值
	params      []ParamInfo   // 参数信息

	// 方法配置
	filters     []*FilterFunc // 方法级过滤器
	checkXSRF   bool          // 是否检查XSRF
	prepareFunc string        // 准备函数名
	finishFunc  string        // 结束函数名
}

// ParamInfo 参数信息
type ParamInfo struct {
	name         string       // 参数名
	paramType    reflect.Type // 参数类型
	required     bool         // 是否必需
	defaultValue any          // 默认值
}

// FilterFunc 过滤器函数（简化版）
type FilterFunc func(*contextenhanced.Context, *FilterChain)

// HandlerFunc 简化的处理函数类型
type HandlerFunc func(*contextenhanced.Context)

// FilterChain 过滤器链
type FilterChain struct {
	filters []FilterFunc
	index   int
}

// FilterRouter 过滤器路由
type FilterRouter struct {
	tree           *FilterTree
	pattern        string
	filterFunc     FilterFunc
	returnOnOutput bool
}

// FilterTree 过滤器树
type FilterTree struct {
	fixedRoutes map[string]*FilterFunc
	regexRoutes []*FilterRouter
}

// ============= 过滤器位置常量 =============

const (
	BeforeStatic = iota // 静态文件之前
	BeforeRouter        // 路由之前
	BeforeExec          // 执行之前
	AfterExec           // 执行之后
	FinishRouter        // 路由完成
)

// NewControllerRegister 创建新的控制器注册器
func NewControllerRegister() *ControllerRegister {
	cr := &ControllerRegister{
		routers:          make(map[string]*ControllerTree),
		policies:         make(map[string]*FilterFunc),
		filters:          make([]*FilterRouter, 0),
		enableCompress:   false,
		enableGzip:       false,
		maxMemory:        1 << 26, // 64MB
		enableAutoPrefix: false,
	}

	// 初始化HTTP方法的路由树
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "ANY"}
	for _, method := range methods {
		cr.routers[method] = &ControllerTree{
			fixedRoutes: make(map[string]*ControllerInfo),
			regexRoutes: make([]*ControllerInfo, 0),
		}
	}

	// 初始化控制器实例池
	cr.pool.New = func() any {
		return make(map[string]any)
	}

	return cr
}

// ============= 核心注册方法 =============

// Add 添加控制器路由（Beego兼容）
func (cr *ControllerRegister) Add(pattern string, c core.IController, mappingMethods ...string) {
	cr.addController(pattern, c, mappingMethods...)
}

// AddAuto 自动添加控制器路由
func (cr *ControllerRegister) AddAuto(c core.IController) {
	cr.addAutoController(c)
}

// AddAutoPrefix 添加带前缀的自动路由
func (cr *ControllerRegister) AddAutoPrefix(prefix string, c core.IController) {
	cr.addAutoPrefixController(prefix, c)
}

// AddMethod 添加指定HTTP方法的路由（支持简化的HandlerFunc）
func (cr *ControllerRegister) AddMethod(method, pattern string, f any) {
	cr.addMethodRoute(method, pattern, f)
}

// Get 添加GET路由
func (cr *ControllerRegister) Get(pattern string, f any) {
	cr.AddMethod("GET", pattern, f)
}

// Post 添加POST路由
func (cr *ControllerRegister) Post(pattern string, f any) {
	cr.AddMethod("POST", pattern, f)
}

// Put 添加PUT路由
func (cr *ControllerRegister) Put(pattern string, f any) {
	cr.AddMethod("PUT", pattern, f)
}

// Delete 添加DELETE路由
func (cr *ControllerRegister) Delete(pattern string, f any) {
	cr.AddMethod("DELETE", pattern, f)
}

// Head 添加HEAD路由
func (cr *ControllerRegister) Head(pattern string, f any) {
	cr.AddMethod("HEAD", pattern, f)
}

// Options 添加OPTIONS路由
func (cr *ControllerRegister) Options(pattern string, f any) {
	cr.AddMethod("OPTIONS", pattern, f)
}

// Any 添加任意方法路由
func (cr *ControllerRegister) Any(pattern string, f any) {
	cr.AddMethod("*", pattern, f)
}

// ============= 过滤器方法 =============

// InsertFilter 插入过滤器
func (cr *ControllerRegister) InsertFilter(pattern string, pos int, filter FilterFunc, params ...bool) error {
	return cr.insertFilter(pattern, pos, filter, params...)
}

// InsertFilterChain 插入过滤器链
func (cr *ControllerRegister) InsertFilterChain(pattern string, filterChain ...FilterFunc) {
	for i, filter := range filterChain {
		cr.InsertFilter(pattern, BeforeRouter+i, filter)
	}
}

// ============= 策略方法 =============

// Policy 设置策略
func (cr *ControllerRegister) Policy(pattern string, policy FilterFunc) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.policies[pattern] = &policy
	cr.enablePolicy = true
}

// ============= 核心处理方法 =============

// ServeHTTP 处理HTTP请求（简化接口，只使用Context）
func (cr *ControllerRegister) ServeHTTP(ctx *contextenhanced.Context) {
	cr.serveHTTP(ctx)
}

// FindRouter 查找路由
func (cr *ControllerRegister) FindRouter(ctx *contextenhanced.Context) (routerInfo *ControllerInfo, isFind bool) {
	return cr.findRouter(ctx)
}

// ============= 内部实现方法 =============

// addController 添加控制器
func (cr *ControllerRegister) addController(pattern string, c core.IController, mappingMethods ...string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// 获取控制器类型
	controllerType := reflect.TypeOf(c)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	// 创建控制器信息
	controllerInfo := &ControllerInfo{
		pattern:        pattern,
		controllerType: controllerType,
		methods:        make(map[string]*MethodInfo),
		filters:        make([]*FilterFunc, 0),
	}

	// 编译路由正则表达式
	if err := cr.compilePattern(controllerInfo); err != nil {
		config.Errorf("Failed to compile pattern %s: %v", pattern, err)
		return
	}

	// 扫描控制器方法
	cr.scanControllerMethods(controllerInfo, c, mappingMethods...)

	// 注册到路由树
	cr.addToRouteTree(controllerInfo)
}

// addAutoController 自动添加控制器
func (cr *ControllerRegister) addAutoController(c core.IController) {
	controllerType := reflect.TypeOf(c)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	// 从控制器名称生成路由模式
	controllerName := strings.ToLower(controllerType.Name())
	if strings.HasSuffix(controllerName, "controller") {
		controllerName = controllerName[:len(controllerName)-10] // 去除"controller"后缀
	}

	pattern := "/" + controllerName + "/*"
	cr.addController(pattern, c)
}

// addAutoPrefixController 添加带前缀的自动控制器
func (cr *ControllerRegister) addAutoPrefixController(prefix string, c core.IController) {
	controllerType := reflect.TypeOf(c)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	controllerName := strings.ToLower(controllerType.Name())
	if strings.HasSuffix(controllerName, "controller") {
		controllerName = controllerName[:len(controllerName)-10]
	}

	pattern := "/" + strings.Trim(prefix, "/") + "/" + controllerName + "/*"
	cr.addController(pattern, c)
}

// addMethodRoute 添加方法路由（支持简化的HandlerFunc）
func (cr *ControllerRegister) addMethodRoute(method, pattern string, f any) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// 创建函数型控制器信息
	controllerInfo := &ControllerInfo{
		pattern: pattern,
		methods: make(map[string]*MethodInfo),
		filters: make([]*FilterFunc, 0),
	}

	// 编译路由模式
	if err := cr.compilePattern(controllerInfo); err != nil {
		config.Errorf("Failed to compile pattern %s: %v", pattern, err)
		return
	}

	// 创建方法信息
	methodInfo := &MethodInfo{
		methodName: "handler",
		httpMethod: strings.ToUpper(method),
		filters:    make([]*FilterFunc, 0),
	}

	// 根据函数类型设置方法值
	funcValue := reflect.ValueOf(f)
	funcType := funcValue.Type()
	methodInfo.methodType = funcType
	methodInfo.methodValue = funcValue

	controllerInfo.methods[strings.ToUpper(method)] = methodInfo

	// 注册到路由树
	cr.addToRouteTree(controllerInfo)
}

// compilePattern 编译路由模式
func (cr *ControllerRegister) compilePattern(info *ControllerInfo) error {
	pattern := info.pattern

	// 处理简单路径
	if !strings.ContainsAny(pattern, ":*?()[]{}") {
		return nil
	}

	// 构建正则表达式
	regexPattern := "^"
	params := make([]string, 0)

	// 分割路径段
	segments := strings.Split(strings.Trim(pattern, "/"), "/")

	for i, segment := range segments {
		if i > 0 {
			regexPattern += "/"
		}

		if strings.HasPrefix(segment, ":") {
			// 命名参数 :id
			paramName := segment[1:]
			params = append(params, paramName)
			regexPattern += "([^/]+)"
		} else if segment == "*" {
			// 通配符参数
			params = append(params, "splat")
			regexPattern += "(.*)"
		} else if strings.Contains(segment, "*") {
			// 部分通配符
			escaped := regexp.QuoteMeta(segment)
			escaped = strings.ReplaceAll(escaped, "\\*", "([^/]*)")
			regexPattern += escaped
		} else {
			// 字面量段
			regexPattern += regexp.QuoteMeta(segment)
		}
	}

	regexPattern += "$"

	// 编译正则表达式
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return fmt.Errorf("invalid route pattern: %s", pattern)
	}

	info.routeRegex = regex
	info.params = params

	return nil
}

// scanControllerMethods 扫描控制器方法
func (cr *ControllerRegister) scanControllerMethods(info *ControllerInfo, c core.IController, mappingMethods ...string) {
	controllerType := reflect.TypeOf(c)
	controllerValue := reflect.ValueOf(c)

	// 如果指定了映射方法，只处理指定的方法
	if len(mappingMethods) > 0 {
		for _, mapping := range mappingMethods {
			parts := strings.Split(mapping, ":")
			if len(parts) != 2 {
				continue
			}

			httpMethod := strings.ToUpper(strings.TrimSpace(parts[0]))
			methodName := strings.TrimSpace(parts[1])

			if method := controllerValue.MethodByName(methodName); method.IsValid() {
				methodInfo := &MethodInfo{
					methodName:  methodName,
					httpMethod:  httpMethod,
					methodType:  method.Type(),
					methodValue: method,
					filters:     make([]*FilterFunc, 0),
				}
				info.methods[httpMethod] = methodInfo
			}
		}
		return
	}

	// 自动扫描所有公开方法
	numMethods := controllerType.NumMethod()
	for i := 0; i < numMethods; i++ {
		method := controllerType.Method(i)
		methodName := method.Name

		// 跳过生命周期方法和私有方法
		if cr.isReservedMethod(methodName) {
			continue
		}

		// 根据方法名前缀确定HTTP方法
		httpMethod := cr.extractHTTPMethod(methodName)
		if httpMethod == "" {
			continue
		}

		methodInfo := &MethodInfo{
			methodName:  methodName,
			httpMethod:  httpMethod,
			methodType:  method.Type,
			methodValue: controllerValue.Method(i),
			filters:     make([]*FilterFunc, 0),
		}

		info.methods[httpMethod] = methodInfo
	}
}

// isReservedMethod 检查是否为保留方法
func (cr *ControllerRegister) isReservedMethod(methodName string) bool {
	reservedMethods := []string{
		"Init", "Prepare", "Finish",
		"InitWithContext", "Init",
		"GetControllerName", "GetActionName",
		"SetData", "GetData", "DelData",
		"Render", "JSON", "XML", "YAML",
		"Redirect", "Error", "Abort",
	}

	for _, reserved := range reservedMethods {
		if methodName == reserved {
			return true
		}
	}

	return false
}

// extractHTTPMethod 从方法名提取HTTP方法
func (cr *ControllerRegister) extractHTTPMethod(methodName string) string {
	httpMethods := []string{"Get", "Post", "Put", "Delete", "Patch", "Head", "Options", "Connect", "Trace"}

	for _, httpMethod := range httpMethods {
		if strings.HasPrefix(methodName, httpMethod) {
			return strings.ToUpper(httpMethod)
		}
	}

	return ""
}

// addToRouteTree 添加到路由树
func (cr *ControllerRegister) addToRouteTree(info *ControllerInfo) {
	// 根据支持的HTTP方法添加到对应的路由树
	for httpMethod := range info.methods {
		tree := cr.routers[httpMethod]
		if tree == nil {
			tree = &ControllerTree{
				fixedRoutes: make(map[string]*ControllerInfo),
				regexRoutes: make([]*ControllerInfo, 0),
			}
			cr.routers[httpMethod] = tree
		}

		if info.routeRegex != nil {
			// 正则路由
			tree.regexRoutes = append(tree.regexRoutes, info)
			// 按模式长度排序，长的优先匹配
			sort.Slice(tree.regexRoutes, func(i, j int) bool {
				return len(tree.regexRoutes[i].pattern) > len(tree.regexRoutes[j].pattern)
			})
		} else {
			// 固定路由
			tree.fixedRoutes[info.pattern] = info
		}
	}

	// 也添加到ANY路由树
	anyTree := cr.routers["ANY"]
	if anyTree == nil {
		anyTree = &ControllerTree{
			fixedRoutes: make(map[string]*ControllerInfo),
			regexRoutes: make([]*ControllerInfo, 0),
		}
		cr.routers["ANY"] = anyTree
	}

	if info.routeRegex != nil {
		anyTree.regexRoutes = append(anyTree.regexRoutes, info)
		sort.Slice(anyTree.regexRoutes, func(i, j int) bool {
			return len(anyTree.regexRoutes[i].pattern) > len(anyTree.regexRoutes[j].pattern)
		})
	} else {
		anyTree.fixedRoutes[info.pattern] = info
	}
}
