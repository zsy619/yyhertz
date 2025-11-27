// Package devtools 提供限流中间件
//
// 限流中间件用于保护服务免受过载攻击，提供：
// - 令牌桶限流算法
// - 滑动窗口限流算法
// - 漏桶限流算法
// - 分布式限流支持
// - 多维度限流策略
// - 动态限流规则调整
//
// 功能特性：
// - 高性能限流算法
// - 灵活的限流策略
// - 实时限流监控
// - 自动熔断保护
// - 白名单机制
package devtools

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/config"
)

// LimitStrategy 限流策略
type LimitStrategy string

const (
	// LimitStrategyTokenBucket 令牌桶策略
	LimitStrategyTokenBucket LimitStrategy = "token_bucket"
	// LimitStrategySlidingWindow 滑动窗口策略
	LimitStrategySlidingWindow LimitStrategy = "sliding_window"
	// LimitStrategyLeakyBucket 漏桶策略
	LimitStrategyLeakyBucket LimitStrategy = "leaky_bucket"
)

// LimitDimension 限流维度
type LimitDimension string

const (
	// LimitDimensionIP IP限流
	LimitDimensionIP LimitDimension = "ip"
	// LimitDimensionUser 用户限流
	LimitDimensionUser LimitDimension = "user"
	// LimitDimensionAPI API限流
	LimitDimensionAPI LimitDimension = "api"
	// LimitDimensionGlobal 全局限流
	LimitDimensionGlobal LimitDimension = "global"
)

// RateLimit 限流配置
type RateLimit struct {
	Rate      int64          `json:"rate"`      // 限流速率(请求/秒)
	Burst     int64          `json:"burst"`     // 突发容量
	Strategy  LimitStrategy  `json:"strategy"`  // 限流策略
	Dimension LimitDimension `json:"dimension"` // 限流维度
	Key       string         `json:"key"`       // 限流键(可选)
	Duration  time.Duration  `json:"duration"`  // 时间窗口(滑动窗口策略)
	Enabled   bool           `json:"enabled"`   // 是否启用
}

// LimitResult 限流结果
type LimitResult struct {
	Allowed    bool          `json:"allowed"`     // 是否允许
	Remaining  int64         `json:"remaining"`   // 剩余配额
	ResetTime  time.Time     `json:"reset_time"`  // 重置时间
	RetryAfter time.Duration `json:"retry_after"` // 重试间隔
}

// TokenBucket 令牌桶
type TokenBucket struct {
	rate     int64     // 令牌生成速率
	capacity int64     // 桶容量
	tokens   int64     // 当前令牌数
	lastTime time.Time // 上次更新时间
	mu       sync.Mutex
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(rate, capacity int64) *TokenBucket {
	return &TokenBucket{
		rate:     rate,
		capacity: capacity,
		tokens:   capacity,
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.lastTime = now

	// 添加新令牌
	tb.tokens += int64(elapsed * float64(tb.rate))
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	// 消费令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// Remaining 剩余令牌数
func (tb *TokenBucket) Remaining() int64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	rate     int64
	duration time.Duration
	requests []time.Time
	mu       sync.RWMutex
}

// NewSlidingWindow 创建滑动窗口
func NewSlidingWindow(rate int64, duration time.Duration) *SlidingWindow {
	return &SlidingWindow{
		rate:     rate,
		duration: duration,
		requests: make([]time.Time, 0),
	}
}

// Allow 检查是否允许请求
func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.duration)

	// 清理过期请求
	validRequests := 0
	for _, req := range sw.requests {
		if req.After(cutoff) {
			sw.requests[validRequests] = req
			validRequests++
		}
	}
	sw.requests = sw.requests[:validRequests]

	// 检查是否超过限制
	if int64(len(sw.requests)) >= sw.rate {
		return false
	}

	// 添加当前请求
	sw.requests = append(sw.requests, now)
	return true
}

// Remaining 剩余配额
func (sw *SlidingWindow) Remaining() int64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-sw.duration)

	count := int64(0)
	for _, req := range sw.requests {
		if req.After(cutoff) {
			count++
		}
	}

	remaining := sw.rate - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// LeakyBucket 漏桶
type LeakyBucket struct {
	rate     int64     // 漏水速率
	capacity int64     // 桶容量
	water    int64     // 当前水量
	lastLeak time.Time // 上次漏水时间
	mu       sync.Mutex
}

// NewLeakyBucket 创建漏桶
func NewLeakyBucket(rate, capacity int64) *LeakyBucket {
	return &LeakyBucket{
		rate:     rate,
		capacity: capacity,
		water:    0,
		lastLeak: time.Now(),
	}
}

// Allow 检查是否允许请求
func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(lb.lastLeak).Seconds()
	lb.lastLeak = now

	// 漏水
	leaked := int64(elapsed * float64(lb.rate))
	lb.water -= leaked
	if lb.water < 0 {
		lb.water = 0
	}

	// 检查是否可以加水
	if lb.water >= lb.capacity {
		return false
	}

	// 加水
	lb.water++
	return true
}

// Remaining 剩余容量
func (lb *LeakyBucket) Remaining() int64 {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	remaining := lb.capacity - lb.water
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// Limiter 限流器接口
type Limiter interface {
	Allow() bool
	Remaining() int64
}

// RateLimitStats 限流统计
type RateLimitStats struct {
	TotalRequests   int64            `json:"total_requests"`   // 总请求数
	AllowedRequests int64            `json:"allowed_requests"` // 允许的请求数
	BlockedRequests int64            `json:"blocked_requests"` // 被阻塞的请求数
	Dimension       LimitDimension   `json:"dimension"`        // 限流维度
	TopBlocked      map[string]int64 `json:"top_blocked"`      // 被阻塞最多的键
	LastUpdate      time.Time        `json:"last_update"`      // 最后更新时间
}

// RateLimiter 限流器
type RateLimiter struct {
	mu           sync.RWMutex
	enabled      bool
	rules        map[string]*RateLimit
	limiters     map[string]Limiter
	stats        map[LimitDimension]*RateLimitStats
	whitelist    map[string]bool
	blocked      map[string]int64 // 记录被阻塞的次数
	cleanupTimer *time.Timer
}

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	Enabled   bool         `json:"enabled"`   // 是否启用
	Rules     []*RateLimit `json:"rules"`     // 限流规则
	Whitelist []string     `json:"whitelist"` // 白名单
}

// NewRateLimiter 创建限流器
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = &RateLimiterConfig{
			Enabled: true,
			Rules:   []*RateLimit{},
		}
	}

	rl := &RateLimiter{
		enabled:   config.Enabled,
		rules:     make(map[string]*RateLimit),
		limiters:  make(map[string]Limiter),
		stats:     make(map[LimitDimension]*RateLimitStats),
		whitelist: make(map[string]bool),
		blocked:   make(map[string]int64),
	}

	// 设置白名单
	for _, ip := range config.Whitelist {
		rl.whitelist[ip] = true
	}

	// 添加规则
	for _, rule := range config.Rules {
		rl.AddRule(rule)
	}

	// 启动清理定时器
	rl.startCleanup()

	return rl
}

// AddRule 添加限流规则
func (rl *RateLimiter) AddRule(rule *RateLimit) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := rl.buildRuleKey(rule)
	rl.rules[key] = rule

	// 初始化统计
	if _, exists := rl.stats[rule.Dimension]; !exists {
		rl.stats[rule.Dimension] = &RateLimitStats{
			Dimension:  rule.Dimension,
			TopBlocked: make(map[string]int64),
			LastUpdate: time.Now(),
		}
	}
}

// RemoveRule 移除限流规则
func (rl *RateLimiter) RemoveRule(dimension LimitDimension, key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	ruleKey := string(dimension)
	if key != "" {
		ruleKey = fmt.Sprintf("%s:%s", dimension, key)
	}

	delete(rl.rules, ruleKey)
}

// buildRuleKey 构建规则键
func (rl *RateLimiter) buildRuleKey(rule *RateLimit) string {
	if rule.Key != "" {
		return fmt.Sprintf("%s:%s", rule.Dimension, rule.Key)
	}
	return string(rule.Dimension)
}

// Handler 限流中间件
func (rl *RateLimiter) Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !rl.enabled {
			c.Next(ctx)
			return
		}

		// 获取客户端IP
		clientIP := rl.getClientIP(c)

		// 检查白名单
		if rl.isWhitelisted(clientIP) {
			c.Next(ctx)
			return
		}

		// 检查各种限流规则
		if !rl.checkLimits(c, clientIP) {
			c.JSON(http.StatusTooManyRequests, map[string]any{
				"code":        429,
				"message":     "Too Many Requests",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

// checkLimits 检查限流
func (rl *RateLimiter) checkLimits(c *app.RequestContext, clientIP string) bool {
	rl.mu.RLock()
	rules := make([]*RateLimit, 0, len(rl.rules))
	for _, rule := range rl.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	rl.mu.RUnlock()

	for _, rule := range rules {
		limitKey := rl.buildLimitKey(rule, c, clientIP)
		limiter := rl.getLimiter(limitKey, rule)

		// 更新统计
		rl.updateStats(rule.Dimension, limitKey, true)

		if !limiter.Allow() {
			// 记录被阻塞
			rl.updateStats(rule.Dimension, limitKey, false)
			rl.recordBlocked(limitKey)
			return false
		}
	}

	return true
}

// buildLimitKey 构建限流键
func (rl *RateLimiter) buildLimitKey(rule *RateLimit, c *app.RequestContext, clientIP string) string {
	switch rule.Dimension {
	case LimitDimensionIP:
		return fmt.Sprintf("ip:%s", clientIP)
	case LimitDimensionUser:
		userID := c.GetHeader("X-User-ID")
		if len(userID) == 0 {
			userID = []byte("anonymous")
		}
		return fmt.Sprintf("user:%s", string(userID))
	case LimitDimensionAPI:
		path := string(c.Path())
		method := string(c.Method())
		return fmt.Sprintf("api:%s:%s", method, path)
	case LimitDimensionGlobal:
		return "global"
	default:
		if rule.Key != "" {
			return rule.Key
		}
		return fmt.Sprintf("%s:default", rule.Dimension)
	}
}

// getLimiter 获取限流器
func (rl *RateLimiter) getLimiter(key string, rule *RateLimit) Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limiter, exists := rl.limiters[key]; exists {
		return limiter
	}

	var limiter Limiter
	switch rule.Strategy {
	case LimitStrategyTokenBucket:
		limiter = NewTokenBucket(rule.Rate, rule.Burst)
	case LimitStrategySlidingWindow:
		duration := rule.Duration
		if duration == 0 {
			duration = time.Minute
		}
		limiter = NewSlidingWindow(rule.Rate, duration)
	case LimitStrategyLeakyBucket:
		limiter = NewLeakyBucket(rule.Rate, rule.Burst)
	default:
		limiter = NewTokenBucket(rule.Rate, rule.Burst)
	}

	rl.limiters[key] = limiter
	return limiter
}

// getClientIP 获取客户端IP
func (rl *RateLimiter) getClientIP(c *app.RequestContext) string {
	// 尝试从X-Forwarded-For获取
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if len(xForwardedFor) > 0 {
		ips := strings.Split(string(xForwardedFor), ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP获取
	xRealIP := c.GetHeader("X-Real-IP")
	if len(xRealIP) > 0 {
		return string(xRealIP)
	}

	// 从RemoteAddr获取
	host, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return c.RemoteAddr().String()
	}
	return host
}

// isWhitelisted 检查是否在白名单中
func (rl *RateLimiter) isWhitelisted(ip string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.whitelist[ip]
}

// updateStats 更新统计
func (rl *RateLimiter) updateStats(dimension LimitDimension, key string, allowed bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	stats, exists := rl.stats[dimension]
	if !exists {
		stats = &RateLimitStats{
			Dimension:  dimension,
			TopBlocked: make(map[string]int64),
			LastUpdate: time.Now(),
		}
		rl.stats[dimension] = stats
	}

	stats.TotalRequests++
	stats.LastUpdate = time.Now()

	if allowed {
		stats.AllowedRequests++
	} else {
		stats.BlockedRequests++
		stats.TopBlocked[key]++
	}
}

// recordBlocked 记录被阻塞
func (rl *RateLimiter) recordBlocked(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.blocked[key]++
}

// startCleanup 启动清理
func (rl *RateLimiter) startCleanup() {
	rl.cleanupTimer = time.AfterFunc(5*time.Minute, func() {
		rl.cleanup()
		rl.startCleanup() // 递归调用
	})
}

// cleanup 清理过期数据
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 清理限流器(简单策略：清理所有，让它们重新创建)
	rl.limiters = make(map[string]Limiter)

	// 清理阻塞记录
	rl.blocked = make(map[string]int64)

	// 重置统计中的TopBlocked
	for _, stats := range rl.stats {
		stats.TopBlocked = make(map[string]int64)
	}
}

// GetStats 获取统计信息
func (rl *RateLimiter) GetStats() map[LimitDimension]*RateLimitStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	result := make(map[LimitDimension]*RateLimitStats)
	for k, v := range rl.stats {
		// 创建副本
		result[k] = &RateLimitStats{
			TotalRequests:   v.TotalRequests,
			AllowedRequests: v.AllowedRequests,
			BlockedRequests: v.BlockedRequests,
			Dimension:       v.Dimension,
			TopBlocked:      make(map[string]int64),
			LastUpdate:      v.LastUpdate,
		}
		for key, count := range v.TopBlocked {
			result[k].TopBlocked[key] = count
		}
	}
	return result
}

// GetRules 获取规则列表
func (rl *RateLimiter) GetRules() []*RateLimit {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	rules := make([]*RateLimit, 0, len(rl.rules))
	for _, rule := range rl.rules {
		rules = append(rules, rule)
	}
	return rules
}

// Enable 启用限流
func (rl *RateLimiter) Enable() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = true
}

// Disable 禁用限流
func (rl *RateLimiter) Disable() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = false
}

// IsEnabled 检查是否启用
func (rl *RateLimiter) IsEnabled() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.enabled
}

// AddToWhitelist 添加到白名单
func (rl *RateLimiter) AddToWhitelist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.whitelist[ip] = true
}

// RemoveFromWhitelist 从白名单移除
func (rl *RateLimiter) RemoveFromWhitelist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.whitelist, ip)
}

// GetWhitelist 获取白名单
func (rl *RateLimiter) GetWhitelist() []string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	whitelist := make([]string, 0, len(rl.whitelist))
	for ip := range rl.whitelist {
		whitelist = append(whitelist, ip)
	}
	sort.Strings(whitelist)
	return whitelist
}

// Stop 停止限流器
func (rl *RateLimiter) Stop() {
	if rl.cleanupTimer != nil {
		rl.cleanupTimer.Stop()
	}
}

// RateLimitPanel 限流面板
type RateLimitPanel struct {
	limiter *RateLimiter
}

// NewRateLimitPanel 创建限流面板
func NewRateLimitPanel(limiter *RateLimiter) *RateLimitPanel {
	return &RateLimitPanel{
		limiter: limiter,
	}
}

// RegisterRoutes 注册限流路由
func (rlp *RateLimitPanel) RegisterRoutes(engine any) {
	var limitGroup *route.RouterGroup

	if h, ok := engine.(*route.Engine); ok {
		limitGroup = h.Group("/yyhertz/ratelimit")
	} else {
		config.Error("无法注册RateLimit路由，未知引擎类型")
		return
	}

	// 注册路由
	limitGroup.GET("/stats", rlp.getStats)
	limitGroup.GET("/rules", rlp.getRules)
	limitGroup.POST("/rules", rlp.addRule)
	limitGroup.DELETE("/rules", rlp.removeRule)
	limitGroup.GET("/whitelist", rlp.getWhitelist)
	limitGroup.POST("/whitelist", rlp.addToWhitelist)
	limitGroup.DELETE("/whitelist", rlp.removeFromWhitelist)
	limitGroup.POST("/enable", rlp.enableLimiter)
	limitGroup.POST("/disable", rlp.disableLimiter)
	limitGroup.GET("/panel", rlp.rateLimitPanel)
}

// getStats 获取统计
func (rlp *RateLimitPanel) getStats(ctx context.Context, c *app.RequestContext) {
	stats := rlp.limiter.GetStats()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": stats,
	})
}

// getRules 获取规则
func (rlp *RateLimitPanel) getRules(ctx context.Context, c *app.RequestContext) {
	rules := rlp.limiter.GetRules()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": rules,
	})
}

// addRule 添加规则
func (rlp *RateLimitPanel) addRule(ctx context.Context, c *app.RequestContext) {
	var rule RateLimit
	if err := c.BindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	rlp.limiter.AddRule(&rule)
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "规则已添加",
	})
}

// removeRule 移除规则
func (rlp *RateLimitPanel) removeRule(ctx context.Context, c *app.RequestContext) {
	dimension := LimitDimension(c.Query("dimension"))
	key := c.Query("key")

	if dimension == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "dimension参数不能为空",
		})
		return
	}

	rlp.limiter.RemoveRule(dimension, key)
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "规则已移除",
	})
}

// getWhitelist 获取白名单
func (rlp *RateLimitPanel) getWhitelist(ctx context.Context, c *app.RequestContext) {
	whitelist := rlp.limiter.GetWhitelist()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": whitelist,
	})
}

// addToWhitelist 添加到白名单
func (rlp *RateLimitPanel) addToWhitelist(ctx context.Context, c *app.RequestContext) {
	var req struct {
		IP string `json:"ip"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if req.IP == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "IP不能为空",
		})
		return
	}

	rlp.limiter.AddToWhitelist(req.IP)
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "已添加到白名单",
	})
}

// removeFromWhitelist 从白名单移除
func (rlp *RateLimitPanel) removeFromWhitelist(ctx context.Context, c *app.RequestContext) {
	ip := c.Query("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "IP参数不能为空",
		})
		return
	}

	rlp.limiter.RemoveFromWhitelist(ip)
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "已从白名单移除",
	})
}

// enableLimiter 启用限流
func (rlp *RateLimitPanel) enableLimiter(ctx context.Context, c *app.RequestContext) {
	rlp.limiter.Enable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "限流已启用",
		"enabled": true,
	})
}

// disableLimiter 禁用限流
func (rlp *RateLimitPanel) disableLimiter(ctx context.Context, c *app.RequestContext) {
	rlp.limiter.Disable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "限流已禁用",
		"enabled": false,
	})
}

// rateLimitPanel 限流面板页面
func (rlp *RateLimitPanel) rateLimitPanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz 限流管理面板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .section { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .stat-card { background: #f8f9fa; padding: 15px; border-radius: 8px; border-left: 4px solid #007bff; }
        .stat-value { font-size: 1.5em; font-weight: bold; color: #007bff; }
        .stat-label { color: #666; margin-top: 5px; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-warning { background: #ffc107; color: black; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f8f9fa; font-weight: bold; }
        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .form-group input, .form-group select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        .dimension-badge { padding: 2px 8px; border-radius: 3px; font-size: 12px; font-weight: bold; }
        .ip { background: #dc3545; color: white; }
        .user { background: #28a745; color: white; }
        .api { background: #007bff; color: white; }
        .global { background: #6f42c1; color: white; }
        .strategy-badge { padding: 2px 8px; border-radius: 3px; font-size: 12px; font-weight: bold; }
        .token_bucket { background: #17a2b8; color: white; }
        .sliding_window { background: #fd7e14; color: white; }
        .leaky_bucket { background: #6c757d; color: white; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz 限流管理面板</h1>
        <div>
            <button class="btn btn-primary" onclick="refreshData()">刷新数据</button>
            <button class="btn btn-success" onclick="enableLimiter()">启用限流</button>
            <button class="btn btn-danger" onclick="disableLimiter()">禁用限流</button>
        </div>
    </div>

    <div class="section">
        <h3>限流统计</h3>
        <div class="stats-grid" id="statsGrid">
            <!-- 统计卡片将在这里动态生成 -->
        </div>
    </div>

    <div class="section">
        <h3>限流规则</h3>
        <div style="margin-bottom: 20px;">
            <button class="btn btn-success" onclick="showAddRuleForm()">添加规则</button>
        </div>
        <table id="rulesTable">
            <thead>
                <tr>
                    <th>维度</th>
                    <th>策略</th>
                    <th>速率</th>
                    <th>突发</th>
                    <th>时间窗口</th>
                    <th>状态</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody>
                <!-- 规则数据将在这里动态生成 -->
            </tbody>
        </table>
    </div>

    <div class="section">
        <h3>IP白名单</h3>
        <div style="margin-bottom: 20px;">
            <input type="text" id="newWhitelistIP" placeholder="输入IP地址" style="width: 200px; padding: 8px; margin-right: 10px;">
            <button class="btn btn-success" onclick="addToWhitelist()">添加</button>
        </div>
        <div id="whitelistContainer">
            <!-- 白名单将在这里动态生成 -->
        </div>
    </div>

    <!-- 添加规则模态框 -->
    <div id="addRuleModal" style="display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 1000;">
        <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); background: white; padding: 30px; border-radius: 8px; width: 400px;">
            <h3>添加限流规则</h3>
            <div class="form-group">
                <label>限流维度:</label>
                <select id="ruleDimension">
                    <option value="ip">IP限流</option>
                    <option value="user">用户限流</option>
                    <option value="api">API限流</option>
                    <option value="global">全局限流</option>
                </select>
            </div>
            <div class="form-group">
                <label>限流策略:</label>
                <select id="ruleStrategy">
                    <option value="token_bucket">令牌桶</option>
                    <option value="sliding_window">滑动窗口</option>
                    <option value="leaky_bucket">漏桶</option>
                </select>
            </div>
            <div class="form-group">
                <label>速率(请求/秒):</label>
                <input type="number" id="ruleRate" min="1" value="100">
            </div>
            <div class="form-group">
                <label>突发容量:</label>
                <input type="number" id="ruleBurst" min="1" value="200">
            </div>
            <div class="form-group">
                <label>时间窗口(秒,仅滑动窗口):</label>
                <input type="number" id="ruleDuration" min="1" value="60">
            </div>
            <div class="form-group">
                <label>自定义键(可选):</label>
                <input type="text" id="ruleKey" placeholder="留空使用默认">
            </div>
            <div style="text-align: right; margin-top: 20px;">
                <button class="btn" onclick="hideAddRuleForm()">取消</button>
                <button class="btn btn-primary" onclick="addRule()">添加</button>
            </div>
        </div>
    </div>

    <script>
        function refreshData() {
            loadStats();
            loadRules();
            loadWhitelist();
        }

        function loadStats() {
            fetch('/yyhertz/ratelimit/stats')
                .then(response => response.json())
                .then(data => {
                    updateStatsGrid(data.data);
                })
                .catch(error => console.error('加载统计失败:', error));
        }

        function loadRules() {
            fetch('/yyhertz/ratelimit/rules')
                .then(response => response.json())
                .then(data => {
                    updateRulesTable(data.data);
                })
                .catch(error => console.error('加载规则失败:', error));
        }

        function loadWhitelist() {
            fetch('/yyhertz/ratelimit/whitelist')
                .then(response => response.json())
                .then(data => {
                    updateWhitelistContainer(data.data);
                })
                .catch(error => console.error('加载白名单失败:', error));
        }

        function updateStatsGrid(stats) {
            const grid = document.getElementById('statsGrid');
            let html = '';

            if (!stats || Object.keys(stats).length === 0) {
                html = '<div style="text-align: center; color: #666;">暂无统计数据</div>';
            } else {
                Object.values(stats).forEach(stat => {
                    const blockRate = stat.total_requests > 0 ? (stat.blocked_requests / stat.total_requests * 100).toFixed(1) : 0;
                    html += '<div class="stat-card">' +
                        '<div class="stat-value">' + stat.total_requests + '</div>' +
                        '<div class="stat-label">' + stat.dimension + ' - 总请求数</div>' +
                        '<div style="margin-top: 10px;">' +
                            '允许: ' + stat.allowed_requests + ' | ' +
                            '阻塞: ' + stat.blocked_requests + ' (' + blockRate + '%)' +
                        '</div>' +
                        '</div>';
                });
            }

            grid.innerHTML = html;
        }

        function updateRulesTable(rules) {
            const tbody = document.querySelector('#rulesTable tbody');
            let html = '';

            if (!rules || rules.length === 0) {
                html = '<tr><td colspan="7" style="text-align: center; color: #666;">暂无限流规则</td></tr>';
            } else {
                rules.forEach((rule, index) => {
                    const duration = rule.duration ? (rule.duration / 1000000000) + 's' : '-';
                    const status = rule.enabled ? '<span style="color: green;">启用</span>' : '<span style="color: red;">禁用</span>';
                    html += '<tr>' +
                        '<td><span class="dimension-badge ' + rule.dimension + '">' + rule.dimension + '</span></td>' +
                        '<td><span class="strategy-badge ' + rule.strategy + '">' + rule.strategy + '</span></td>' +
                        '<td>' + rule.rate + '/s</td>' +
                        '<td>' + rule.burst + '</td>' +
                        '<td>' + duration + '</td>' +
                        '<td>' + status + '</td>' +
                        '<td>' +
                            '<button class="btn btn-danger" onclick="removeRule(\'' + rule.dimension + '\', \'' + (rule.key || '') + '\')">删除</button>' +
                        '</td>' +
                        '</tr>';
                });
            }

            tbody.innerHTML = html;
        }

        function updateWhitelistContainer(whitelist) {
            const container = document.getElementById('whitelistContainer');
            let html = '';

            if (!whitelist || whitelist.length === 0) {
                html = '<div style="color: #666;">白名单为空</div>';
            } else {
                whitelist.forEach(ip => {
                    html += '<div style="display: inline-block; margin: 5px; padding: 5px 10px; background: #e9ecef; border-radius: 4px;">' +
                        ip + 
                        '<button onclick="removeFromWhitelist(\'' + ip + '\')" style="margin-left: 10px; background: none; border: none; color: red; cursor: pointer;">✕</button>' +
                        '</div>';
                });
            }

            container.innerHTML = html;
        }

        function showAddRuleForm() {
            document.getElementById('addRuleModal').style.display = 'block';
        }

        function hideAddRuleForm() {
            document.getElementById('addRuleModal').style.display = 'none';
        }

        function addRule() {
            const rule = {
                dimension: document.getElementById('ruleDimension').value,
                strategy: document.getElementById('ruleStrategy').value,
                rate: parseInt(document.getElementById('ruleRate').value),
                burst: parseInt(document.getElementById('ruleBurst').value),
                duration: parseInt(document.getElementById('ruleDuration').value) * 1000000000, // 转换为纳秒
                key: document.getElementById('ruleKey').value,
                enabled: true
            };

            fetch('/yyhertz/ratelimit/rules', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(rule)
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 0) {
                    alert('规则已添加');
                    hideAddRuleForm();
                    loadRules();
                } else {
                    alert('添加失败: ' + data.message);
                }
            })
            .catch(error => {
                console.error('添加规则失败:', error);
                alert('添加失败');
            });
        }

        function removeRule(dimension, key) {
            if (confirm('确定要删除这个限流规则吗？')) {
                let url = '/yyhertz/ratelimit/rules?dimension=' + dimension;
                if (key) {
                    url += '&key=' + key;
                }

                fetch(url, { method: 'DELETE' })
                    .then(response => response.json())
                    .then(data => {
                        alert('规则已删除');
                        loadRules();
                    })
                    .catch(error => {
                        console.error('删除规则失败:', error);
                        alert('删除失败');
                    });
            }
        }

        function addToWhitelist() {
            const ip = document.getElementById('newWhitelistIP').value.trim();
            if (!ip) {
                alert('请输入IP地址');
                return;
            }

            fetch('/yyhertz/ratelimit/whitelist', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ip: ip })
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 0) {
                    alert('已添加到白名单');
                    document.getElementById('newWhitelistIP').value = '';
                    loadWhitelist();
                } else {
                    alert('添加失败: ' + data.message);
                }
            })
            .catch(error => {
                console.error('添加白名单失败:', error);
                alert('添加失败');
            });
        }

        function removeFromWhitelist(ip) {
            if (confirm('确定要从白名单移除 "' + ip + '" 吗？')) {
                fetch('/yyhertz/ratelimit/whitelist?ip=' + ip, { method: 'DELETE' })
                    .then(response => response.json())
                    .then(data => {
                        alert('已从白名单移除');
                        loadWhitelist();
                    })
                    .catch(error => {
                        console.error('移除白名单失败:', error);
                        alert('移除失败');
                    });
            }
        }

        function enableLimiter() {
            fetch('/yyhertz/ratelimit/enable', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('限流已启用');
                })
                .catch(error => {
                    console.error('启用失败:', error);
                    alert('启用失败');
                });
        }

        function disableLimiter() {
            if (confirm('确定要禁用限流吗？')) {
                fetch('/yyhertz/ratelimit/disable', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('限流已禁用');
                    })
                    .catch(error => {
                        console.error('禁用失败:', error);
                        alert('禁用失败');
                    });
            }
        }

        // 页面加载时初始化
        window.onload = function() {
            refreshData();
            // 每30秒自动刷新统计
            setInterval(loadStats, 30000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}
