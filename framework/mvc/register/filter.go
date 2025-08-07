// 过滤器和中间件支持实现
package register

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/config"
	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ============= 过滤器实现 =============

// insertFilter 插入过滤器的具体实现
func (cr *ControllerRegister) insertFilter(pattern string, pos int, filter FilterFunc, params ...bool) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// 验证位置参数
	if pos < BeforeStatic || pos > FinishRouter {
		return fmt.Errorf("invalid filter position: %d", pos)
	}

	// 处理可选参数
	returnOnOutput := true
	if len(params) > 0 {
		returnOnOutput = params[0]
	}

	// 创建过滤器路由
	filterRouter := &FilterRouter{
		pattern:        pattern,
		filterFunc:     filter,
		returnOnOutput: returnOnOutput,
		tree:           cr.createFilterTree(pattern),
	}

	// 插入到指定位置
	if len(cr.filters) <= pos {
		// 扩展切片
		newFilters := make([]*FilterRouter, pos+1)
		copy(newFilters, cr.filters)
		cr.filters = newFilters
	}

	cr.filters = append(cr.filters, nil)
	copy(cr.filters[pos+1:], cr.filters[pos:])
	cr.filters[pos] = filterRouter

	return nil
}

// createFilterTree 创建过滤器树
func (cr *ControllerRegister) createFilterTree(pattern string) *FilterTree {
	tree := &FilterTree{
		fixedRoutes: make(map[string]*FilterFunc),
		regexRoutes: make([]*FilterRouter, 0),
	}

	return tree
}

// ============= 内置过滤器 =============

// LoggingFilter 日志记录过滤器
func LoggingFilter(ctx *contextenhanced.Context, chain *FilterChain) {
	start := time.Now()
	method := string(ctx.RequestContext.Method())
	path := string(ctx.RequestContext.URI().Path())

	// 继续执行
	chain.Next(ctx)

	// 记录日志
	duration := time.Since(start)
	status := ctx.ResponseWriter.Status()

	config.Infof("[%s] %s %s - %d - %v",
		start.Format("2006/01/02 15:04:05"),
		method,
		path,
		status,
		duration)
}

// CORSFilter CORS过滤器
func CORSFilter(ctx *contextenhanced.Context, chain *FilterChain) {
	origin := ctx.Input.Header("Origin")

	// 设置CORS头
	ctx.Output.Header("Access-Control-Allow-Origin", origin)
	ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	ctx.Output.Header("Access-Control-Allow-Credentials", "true")
	ctx.Output.Header("Access-Control-Max-Age", "3600")

	// 处理预检请求
	if string(ctx.RequestContext.Method()) == "OPTIONS" {
		ctx.Output.SetStatus(200)
		return
	}

	chain.Next(ctx)
}

// AuthFilter 认证过滤器
func AuthFilter(ctx *contextenhanced.Context, chain *FilterChain) {
	// 检查认证token
	token := ctx.Input.Header("Authorization")
	if token == "" {
		token = ctx.Input.Cookie("auth_token")
	}

	if token == "" {
		ctx.Output.SetStatus(401)
		ctx.Output.JSON(map[string]any{
			"error": "Authentication required",
			"code":  401,
		}, false, true)
		return
	}

	// 验证token（这里是简化实现）
	if !validateToken(token) {
		ctx.Output.SetStatus(401)
		ctx.Output.JSON(map[string]any{
			"error": "Invalid token",
			"code":  401,
		}, false, true)
		return
	}

	// 设置用户信息到上下文
	userInfo := getUserFromToken(token)
	ctx.Input.Data("user", userInfo)

	chain.Next(ctx)
}

// RateLimitFilter 限流过滤器
func RateLimitFilter(limit int, window time.Duration) FilterFunc {
	limiter := newRateLimiter(limit, window)

	return func(ctx *contextenhanced.Context, chain *FilterChain) {
		clientIP := ctx.Input.IP()

		if !limiter.Allow(clientIP) {
			ctx.Output.SetStatus(429)
			ctx.Output.Header("Retry-After", strconv.Itoa(int(window.Seconds())))
			ctx.Output.JSON(map[string]any{
				"error": "Rate limit exceeded",
				"code":  429,
			}, false, true)
			return
		}

		chain.Next(ctx)
	}
}

// SecurityFilter 安全过滤器
func SecurityFilter(ctx *contextenhanced.Context, chain *FilterChain) {
	// 设置安全头
	ctx.Output.Header("X-Content-Type-Options", "nosniff")
	ctx.Output.Header("X-Frame-Options", "DENY")
	ctx.Output.Header("X-XSS-Protection", "1; mode=block")
	ctx.Output.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	ctx.Output.Header("Content-Security-Policy", "default-src 'self'")

	// 检查HTTP方法
	method := string(ctx.RequestContext.Method())
	allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	allowed := false
	for _, allowedMethod := range allowedMethods {
		if method == allowedMethod {
			allowed = true
			break
		}
	}

	if !allowed {
		ctx.Output.SetStatus(405)
		ctx.Output.JSON(map[string]any{
			"error": "Method not allowed",
			"code":  405,
		}, false, true)
		return
	}

	chain.Next(ctx)
}

// CompressFilter 压缩过滤器
func CompressFilter(ctx *contextenhanced.Context, chain *FilterChain) {
	acceptEncoding := ctx.Input.Header("Accept-Encoding")

	// 检查是否支持gzip
	if strings.Contains(acceptEncoding, "gzip") {
		ctx.Output.Header("Content-Encoding", "gzip")
		ctx.Output.Header("Vary", "Accept-Encoding")
	}

	chain.Next(ctx)
}

// ============= 中间件适配 =============

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(*contextenhanced.Context) error

// WrapMiddleware 将中间件包装为过滤器
func WrapMiddleware(middleware MiddlewareFunc) FilterFunc {
	return func(ctx *contextenhanced.Context, chain *FilterChain) {
		if err := middleware(ctx); err != nil {
			ctx.Output.SetStatus(500)
			ctx.Output.JSON(map[string]any{
				"error": err.Error(),
				"code":  500,
			}, false, true)
			return
		}

		chain.Next(ctx)
	}
}

// WrapHertzMiddleware 将Hertz中间件适配为过滤器
func WrapHertzMiddleware(hertzMiddleware app.HandlerFunc) FilterFunc {
	return func(ctx *contextenhanced.Context, chain *FilterChain) {
		// 执行Hertz中间件
		hertzMiddleware(context.Background(), ctx.RequestContext)

		// 继续执行过滤器链
		chain.Next(ctx)
	}
}

// ============= 路由级过滤器 =============

// RouteFilter 路由级过滤器
type RouteFilter struct {
	pattern string
	filter  FilterFunc
	methods []string
}

// NewRouteFilter 创建路由过滤器
func NewRouteFilter(pattern string, filter FilterFunc, methods ...string) *RouteFilter {
	if len(methods) == 0 {
		methods = []string{"*"}
	}

	return &RouteFilter{
		pattern: pattern,
		filter:  filter,
		methods: methods,
	}
}

// Match 检查是否匹配路由
func (rf *RouteFilter) Match(path, method string) bool {
	// 检查方法
	methodMatch := false
	for _, m := range rf.methods {
		if m == "*" || m == method {
			methodMatch = true
			break
		}
	}

	if !methodMatch {
		return false
	}

	// 检查路径
	return rf.matchPath(path)
}

// matchPath 匹配路径
func (rf *RouteFilter) matchPath(path string) bool {
	// 支持简单的通配符匹配
	if rf.pattern == "*" {
		return true
	}

	if rf.pattern == path {
		return true
	}

	// 支持前缀匹配
	if strings.HasSuffix(rf.pattern, "*") {
		prefix := rf.pattern[:len(rf.pattern)-1]
		return strings.HasPrefix(path, prefix)
	}

	// 支持正则表达式
	if strings.HasPrefix(rf.pattern, "^") && strings.HasSuffix(rf.pattern, "$") {
		regex, err := regexp.Compile(rf.pattern)
		if err != nil {
			return false
		}
		return regex.MatchString(path)
	}

	return false
}

// ============= 过滤器链管理 =============

// FilterManager 过滤器管理器
type FilterManager struct {
	globalFilters []FilterFunc
	routeFilters  []*RouteFilter
	beforeFilters []FilterFunc
	afterFilters  []FilterFunc
	mu            sync.RWMutex
}

// NewFilterManager 创建过滤器管理器
func NewFilterManager() *FilterManager {
	return &FilterManager{
		globalFilters: make([]FilterFunc, 0),
		routeFilters:  make([]*RouteFilter, 0),
		beforeFilters: make([]FilterFunc, 0),
		afterFilters:  make([]FilterFunc, 0),
	}
}

// AddGlobalFilter 添加全局过滤器
func (fm *FilterManager) AddGlobalFilter(filter FilterFunc) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.globalFilters = append(fm.globalFilters, filter)
}

// AddRouteFilter 添加路由过滤器
func (fm *FilterManager) AddRouteFilter(routeFilter *RouteFilter) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.routeFilters = append(fm.routeFilters, routeFilter)
}

// AddBeforeFilter 添加前置过滤器
func (fm *FilterManager) AddBeforeFilter(filter FilterFunc) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.beforeFilters = append(fm.beforeFilters, filter)
}

// AddAfterFilter 添加后置过滤器
func (fm *FilterManager) AddAfterFilter(filter FilterFunc) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.afterFilters = append(fm.afterFilters, filter)
}

// GetMatchingFilters 获取匹配的过滤器
func (fm *FilterManager) GetMatchingFilters(path, method string) []FilterFunc {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	filters := make([]FilterFunc, 0)

	// 添加全局过滤器
	filters = append(filters, fm.globalFilters...)

	// 添加匹配的路由过滤器
	for _, routeFilter := range fm.routeFilters {
		if routeFilter.Match(path, method) {
			filters = append(filters, routeFilter.filter)
		}
	}

	return filters
}

// ============= 辅助函数 =============

// validateToken 验证token（简化实现）
func validateToken(token string) bool {
	// 这里应该实现真正的token验证逻辑
	// 例如JWT验证、数据库查询等
	return len(token) > 10
}

// getUserFromToken 从token获取用户信息（简化实现）
func getUserFromToken(token string) map[string]any {
	// 这里应该从token中解析用户信息
	// 或从数据库中查询用户信息
	return map[string]any{
		"id":   1,
		"name": "test_user",
		"role": "user",
	}
}

// ============= 限流器实现 =============

// RateLimiter 简单的限流器
type RateLimiter struct {
	limit    int
	window   time.Duration
	requests map[string][]time.Time
	mu       sync.Mutex
}

// newRateLimiter 创建限流器
func newRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string][]time.Time),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// 获取该key的请求记录
	requests := rl.requests[key]

	// 清理过期的请求记录
	validRequests := make([]time.Time, 0)
	for _, req := range requests {
		if now.Sub(req) < rl.window {
			validRequests = append(validRequests, req)
		}
	}

	// 检查是否超过限制
	if len(validRequests) >= rl.limit {
		return false
	}

	// 记录当前请求
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}
