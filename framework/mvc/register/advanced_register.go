// Package register 高级路由管理器
// 实现更复杂的路由管理功能，包括路由组、中间件链、动态路由等
package register

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
	contextenhanced "github.com/zsy619/yyhertz/framework/context"
	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// AdvancedControllerRegister 高级控制器注册器
type AdvancedControllerRegister struct {
	*ControllerRegister

	// 高级功能
	routeGroups       map[string]*RouteGroup      // 路由组
	middlewareManager *MiddlewareManager          // 中间件管理器
	routeConstraints  map[string]*RouteConstraint // 路由约束
	dynamicRoutes     []*DynamicRoute             // 动态路由
	versionManager    *VersionManager             // 版本管理器

	// 缓存和性能
	compiledRoutes map[string]*CompiledRoute // 编译后的路由
	routeMetrics   map[string]*RouteMetrics  // 路由指标
	warmupRoutes   []string                  // 预热路由

	// 高级配置
	enableRouteDebug bool          // 是否启用路由调试
	routeTimeout     time.Duration // 路由超时
	maxRouteDepth    int           // 最大路由深度
	caseSensitive    bool          // 是否区分大小写

	mu sync.RWMutex
}

// RouteGroup 路由组
type RouteGroup struct {
	Name        string
	Prefix      string
	Middleware  []string
	Constraints map[string]any
	Routes      []*GroupRoute
	Parent      *RouteGroup
	Children    []*RouteGroup
}

// GroupRoute 组路由
type GroupRoute struct {
	Pattern    string
	Controller core.IController
	Methods    []string
	Name       string
	Metadata   map[string]any
}

// MiddlewareManager 中间件管理器
type MiddlewareManager struct {
	middleware map[string]FilterFunc
	chains     map[string][]string
	mu         sync.RWMutex
}

// RouteConstraint 路由约束
type RouteConstraint struct {
	Name      string
	Pattern   *regexp.Regexp
	Validator func(string) bool
	Converter func(string) any
	ErrorMsg  string
}

// DynamicRoute 动态路由
type DynamicRoute struct {
	ID         string
	Pattern    string
	Handler    any
	Conditions []DynamicCondition
	Active     bool
	CreatedAt  time.Time
	ExpiresAt  *time.Time
}

// DynamicCondition 动态条件
type DynamicCondition struct {
	Type     string // header, query, time, etc.
	Key      string
	Operator string // eq, ne, gt, lt, in, etc.
	Value    any
}

// VersionManager API版本管理器
type VersionManager struct {
	versions       map[string]*APIVersion
	defaultVersion string
	headerName     string // 版本头名称，如 "API-Version"
	queryParam     string // 版本查询参数，如 "version"
}

// APIVersion API版本
type APIVersion struct {
	Version    string
	Routes     map[string]*ControllerInfo
	Deprecated bool
	SunsetDate *time.Time
	Middleware []string
}

// CompiledRoute 编译后的路由
type CompiledRoute struct {
	Pattern     *regexp.Regexp
	ParamNames  []string
	Controller  core.IController
	Methods     map[string]*MethodInfo
	Constraints []*RouteConstraint
	CompiledAt  time.Time
}

// RouteMetrics 路由指标
type RouteMetrics struct {
	Pattern        string
	RequestCount   int64
	ErrorCount     int64
	AverageLatency time.Duration
	LastAccessed   time.Time
	PeakRPS        float64
}

// NewAdvancedControllerRegister 创建高级控制器注册器
func NewAdvancedControllerRegister() *AdvancedControllerRegister {
	return &AdvancedControllerRegister{
		ControllerRegister: NewControllerRegister(),
		routeGroups:        make(map[string]*RouteGroup),
		middlewareManager:  NewMiddlewareManager(),
		routeConstraints:   make(map[string]*RouteConstraint),
		dynamicRoutes:      make([]*DynamicRoute, 0),
		versionManager:     NewVersionManager(),
		compiledRoutes:     make(map[string]*CompiledRoute),
		routeMetrics:       make(map[string]*RouteMetrics),
		warmupRoutes:       make([]string, 0),
		enableRouteDebug:   false,
		routeTimeout:       30 * time.Second,
		maxRouteDepth:      10,
		caseSensitive:      false,
	}
}

// RegisterControllerAdvanced 高级控制器注册
func (acr *AdvancedControllerRegister) RegisterControllerAdvanced(options *AdvancedRegistrationOptions) error {
	acr.mu.Lock()
	defer acr.mu.Unlock()

	// 验证选项
	if err := acr.validateAdvancedOptions(options); err != nil {
		return fmt.Errorf("invalid registration options: %v", err)
	}

	// 编译路由
	compiledRoute, err := acr.compileRoute(options)
	if err != nil {
		return fmt.Errorf("failed to compile route: %v", err)
	}

	// 应用版本管理
	if options.Version != "" {
		if err := acr.registerVersionedRoute(options, compiledRoute); err != nil {
			return fmt.Errorf("failed to register versioned route: %v", err)
		}
	}

	// 应用路由组
	if options.Group != "" {
		if err := acr.registerGroupRoute(options, compiledRoute); err != nil {
			return fmt.Errorf("failed to register group route: %v", err)
		}
	}

	// 应用约束
	if err := acr.applyRouteConstraints(options, compiledRoute); err != nil {
		return fmt.Errorf("failed to apply route constraints: %v", err)
	}

	// 注册到基础ControllerRegister
	acr.ControllerRegister.Add(options.Pattern, options.Controller, options.Methods...)

	// 缓存编译后的路由
	acr.compiledRoutes[options.Pattern] = compiledRoute

	// 初始化路由指标
	acr.routeMetrics[options.Pattern] = &RouteMetrics{
		Pattern:      options.Pattern,
		RequestCount: 0,
		ErrorCount:   0,
		LastAccessed: time.Now(),
	}

	config.Infof("Advanced controller registered: %s", options.Pattern)
	return nil
}

// CreateRouteGroup 创建路由组
func (acr *AdvancedControllerRegister) CreateRouteGroup(name, prefix string) *RouteGroup {
	acr.mu.Lock()
	defer acr.mu.Unlock()

	group := &RouteGroup{
		Name:        name,
		Prefix:      prefix,
		Middleware:  make([]string, 0),
		Constraints: make(map[string]any),
		Routes:      make([]*GroupRoute, 0),
		Children:    make([]*RouteGroup, 0),
	}

	acr.routeGroups[name] = group
	return group
}

// AddDynamicRoute 添加动态路由
func (acr *AdvancedControllerRegister) AddDynamicRoute(route *DynamicRoute) error {
	acr.mu.Lock()
	defer acr.mu.Unlock()

	// 验证动态路由
	if err := acr.validateDynamicRoute(route); err != nil {
		return fmt.Errorf("invalid dynamic route: %v", err)
	}

	// 设置创建时间
	route.CreatedAt = time.Now()
	route.Active = true

	// 添加到动态路由列表
	acr.dynamicRoutes = append(acr.dynamicRoutes, route)

	// 如果是函数处理器，直接注册
	if handler, ok := route.Handler.(func(*contextenhanced.Context)); ok {
		acr.ControllerRegister.Any(route.Pattern, handler)
	}

	config.Infof("Dynamic route added: %s", route.Pattern)
	return nil
}

// RegisterVersionedController 注册版本化控制器
func (acr *AdvancedControllerRegister) RegisterVersionedController(version, pattern string, ctrl core.IController) error {
	versionedPattern := fmt.Sprintf("/v%s%s", version, pattern)

	options := &AdvancedRegistrationOptions{
		Pattern:    versionedPattern,
		Controller: ctrl,
		Version:    version,
		Methods:    []string{"GET", "POST", "PUT", "DELETE"},
	}

	return acr.RegisterControllerAdvanced(options)
}

// AddRouteConstraint 添加路由约束
func (acr *AdvancedControllerRegister) AddRouteConstraint(name string, constraint *RouteConstraint) {
	acr.mu.Lock()
	defer acr.mu.Unlock()

	constraint.Name = name
	acr.routeConstraints[name] = constraint
}

// WarmupRoutes 预热路由
func (acr *AdvancedControllerRegister) WarmupRoutes() error {
	acr.mu.RLock()
	routes := make([]string, len(acr.warmupRoutes))
	copy(routes, acr.warmupRoutes)
	acr.mu.RUnlock()

	for _, route := range routes {
		// 预编译路由
		if _, exists := acr.compiledRoutes[route]; !exists {
			config.Warnf("Route not found for warmup: %s", route)
			continue
		}

		config.Debugf("Warmed up route: %s", route)
	}

	config.Infof("Warmed up %d routes", len(routes))
	return nil
}

// GetRouteMetrics 获取路由指标
func (acr *AdvancedControllerRegister) GetRouteMetrics(pattern string) *RouteMetrics {
	acr.mu.RLock()
	defer acr.mu.RUnlock()

	if metrics, exists := acr.routeMetrics[pattern]; exists {
		return metrics
	}

	return nil
}

// ListDynamicRoutes 列出动态路由
func (acr *AdvancedControllerRegister) ListDynamicRoutes() []*DynamicRoute {
	acr.mu.RLock()
	defer acr.mu.RUnlock()

	// 清理过期的动态路由
	acr.cleanupExpiredRoutes()

	activeRoutes := make([]*DynamicRoute, 0)
	for _, route := range acr.dynamicRoutes {
		if route.Active {
			activeRoutes = append(activeRoutes, route)
		}
	}

	return activeRoutes
}

// ============= 辅助结构 =============

// AdvancedRegistrationOptions 高级注册选项
type AdvancedRegistrationOptions struct {
	Pattern     string
	Controller  core.IController
	Methods     []string
	Group       string
	Version     string
	Constraints []string
	Middleware  []string
	Metadata    map[string]any
	Timeout     time.Duration
	RateLimit   *RateLimit
	Cache       *CacheConfig
}

// RateLimit 限流配置
type RateLimit struct {
	RequestsPerSecond int
	BurstSize         int
	KeyFunc           func(*contextenhanced.Context) string
}

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL        time.Duration
	KeyFunc    func(*contextenhanced.Context) string
	Conditions []CacheCondition
}

// CacheCondition 缓存条件
type CacheCondition struct {
	Method     string
	StatusCode int
	Headers    map[string]string
}

// ============= 内部实现方法 =============

// validateAdvancedOptions 验证高级选项
func (acr *AdvancedControllerRegister) validateAdvancedOptions(options *AdvancedRegistrationOptions) error {
	if options.Pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	if options.Controller == nil {
		return fmt.Errorf("controller cannot be nil")
	}

	if len(options.Methods) == 0 {
		options.Methods = []string{"GET", "POST", "PUT", "DELETE"}
	}

	return nil
}

// compileRoute 编译路由
func (acr *AdvancedControllerRegister) compileRoute(options *AdvancedRegistrationOptions) (*CompiledRoute, error) {
	// 创建正则表达式
	pattern := options.Pattern
	paramNames := make([]string, 0)

	// 处理路径参数
	paramRegex := regexp.MustCompile(`:(\w+)`)
	matches := paramRegex.FindAllStringSubmatch(pattern, -1)
	for _, match := range matches {
		paramNames = append(paramNames, match[1])
		pattern = strings.Replace(pattern, match[0], `([^/]+)`, 1)
	}

	// 编译正则表达式
	compiledPattern, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return nil, fmt.Errorf("failed to compile pattern: %v", err)
	}

	// 获取约束
	constraints := make([]*RouteConstraint, 0)
	for _, constraintName := range options.Constraints {
		if constraint, exists := acr.routeConstraints[constraintName]; exists {
			constraints = append(constraints, constraint)
		}
	}

	return &CompiledRoute{
		Pattern:     compiledPattern,
		ParamNames:  paramNames,
		Controller:  options.Controller,
		Constraints: constraints,
		CompiledAt:  time.Now(),
	}, nil
}

// registerVersionedRoute 注册版本化路由
func (acr *AdvancedControllerRegister) registerVersionedRoute(options *AdvancedRegistrationOptions, compiledRoute *CompiledRoute) error {
	return acr.versionManager.RegisterRoute(options.Version, options.Pattern, compiledRoute)
}

// registerGroupRoute 注册组路由
func (acr *AdvancedControllerRegister) registerGroupRoute(options *AdvancedRegistrationOptions, compiledRoute *CompiledRoute) error {
	if group, exists := acr.routeGroups[options.Group]; exists {
		groupRoute := &GroupRoute{
			Pattern:    options.Pattern,
			Controller: options.Controller,
			Methods:    options.Methods,
			Metadata:   options.Metadata,
		}

		group.Routes = append(group.Routes, groupRoute)
		return nil
	}

	return fmt.Errorf("route group not found: %s", options.Group)
}

// applyRouteConstraints 应用路由约束
func (acr *AdvancedControllerRegister) applyRouteConstraints(options *AdvancedRegistrationOptions, compiledRoute *CompiledRoute) error {
	for _, constraintName := range options.Constraints {
		if constraint, exists := acr.routeConstraints[constraintName]; exists {
			compiledRoute.Constraints = append(compiledRoute.Constraints, constraint)
		}
	}

	return nil
}

// validateDynamicRoute 验证动态路由
func (acr *AdvancedControllerRegister) validateDynamicRoute(route *DynamicRoute) error {
	if route.Pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	if route.Handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	return nil
}

// cleanupExpiredRoutes 清理过期路由
func (acr *AdvancedControllerRegister) cleanupExpiredRoutes() {
	now := time.Now()
	activeRoutes := make([]*DynamicRoute, 0)

	for _, route := range acr.dynamicRoutes {
		if route.ExpiresAt == nil || now.Before(*route.ExpiresAt) {
			activeRoutes = append(activeRoutes, route)
		} else {
			route.Active = false
			config.Infof("Dynamic route expired: %s", route.Pattern)
		}
	}

	acr.dynamicRoutes = activeRoutes
}

// ============= 工厂方法 =============

// NewMiddlewareManager 创建中间件管理器
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{
		middleware: make(map[string]FilterFunc),
		chains:     make(map[string][]string),
	}
}

// NewVersionManager 创建版本管理器
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions:       make(map[string]*APIVersion),
		defaultVersion: "1",
		headerName:     "API-Version",
		queryParam:     "version",
	}
}

// RegisterRoute 注册版本路由
func (vm *VersionManager) RegisterRoute(version, pattern string, compiledRoute *CompiledRoute) error {
	if _, exists := vm.versions[version]; !exists {
		vm.versions[version] = &APIVersion{
			Version:    version,
			Routes:     make(map[string]*ControllerInfo),
			Deprecated: false,
			Middleware: make([]string, 0),
		}
	}

	// 创建该版本的控制器信息
	controllerInfo := &ControllerInfo{
		pattern:        pattern,
		controllerType: reflect.TypeOf(compiledRoute.Controller),
		methods:        compiledRoute.Methods,
		routeRegex:     compiledRoute.Pattern,
		params:         compiledRoute.ParamNames,
		filters:        make([]*FilterFunc, 0),
	}

	// 注册到版本路由映射
	vm.versions[version].Routes[pattern] = controllerInfo

	config.Infof("Versioned route registered: v%s %s", version, pattern)
	return nil
}

// RegisterMiddleware 注册中间件
func (mm *MiddlewareManager) RegisterMiddleware(name string, middleware FilterFunc) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.middleware[name] = middleware
}

// CreateChain 创建中间件链
func (mm *MiddlewareManager) CreateChain(name string, middlewareNames []string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.chains[name] = middlewareNames
}

// GetChain 获取中间件链
func (mm *MiddlewareManager) GetChain(name string) []FilterFunc {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if middlewareNames, exists := mm.chains[name]; exists {
		chain := make([]FilterFunc, 0, len(middlewareNames))
		for _, middlewareName := range middlewareNames {
			if middleware, exists := mm.middleware[middlewareName]; exists {
				chain = append(chain, middleware)
			}
		}
		return chain
	}

	return nil
}
