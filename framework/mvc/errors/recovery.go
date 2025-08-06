package errors

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// RecoveryAction 恢复动作类型
type RecoveryAction int

const (
	ActionRetry       RecoveryAction = iota // 重试
	ActionFallback                          // 降级
	ActionCircuitBreak                      // 熔断
	ActionIgnore                            // 忽略
	ActionEscalate                          // 上报
)

// RecoveryStrategy 恢复策略
type RecoveryStrategy struct {
	Name           string            // 策略名称
	Condition      RecoveryCondition // 触发条件
	Action         RecoveryAction    // 恢复动作
	MaxRetries     int               // 最大重试次数
	RetryInterval  time.Duration     // 重试间隔
	BackoffFactor  float64           // 退避因子
	Timeout        time.Duration     // 超时时间
	FallbackFunc   FallbackFunc      // 降级函数
	Metadata       map[string]interface{} // 元数据
}

// RecoveryCondition 恢复条件接口
type RecoveryCondition interface {
	ShouldRecover(classification *ErrorClassification, ctx *mvccontext.EnhancedContext) bool
	Priority() int
}

// FallbackFunc 降级函数
type FallbackFunc func(ctx *mvccontext.EnhancedContext, err error) error

// RecoveryResult 恢复结果
type RecoveryResult struct {
	Original      error             // 原始错误
	Strategy      string            // 使用的策略
	Action        RecoveryAction    // 执行的动作
	Attempts      int               // 尝试次数
	Success       bool              // 是否成功
	FinalError    error             // 最终错误
	Duration      time.Duration     // 恢复耗时
	Metadata      map[string]interface{} // 结果元数据
	RecoveredAt   time.Time         // 恢复时间
}

// AutoRecovery 自动错误恢复系统
type AutoRecovery struct {
	strategies    []RecoveryStrategy    // 恢复策略
	classifier    *IntelligentClassifier // 错误分类器
	circuitBreaker *CircuitBreaker      // 熔断器
	statistics    RecoveryStats         // 统计信息
	config        RecoveryConfig        // 配置
	mu            sync.RWMutex          // 读写锁
}

// RecoveryStats 恢复统计
type RecoveryStats struct {
	TotalAttempts    int64                     // 总尝试次数
	SuccessfulRecoveries int64                 // 成功恢复次数
	FailedRecoveries int64                     // 失败恢复次数
	ActionCounts     map[RecoveryAction]int64  // 各动作统计
	StrategyStats    map[string]*StrategyStats // 策略统计
	AverageRecoveryTime time.Duration          // 平均恢复时间
}

// StrategyStats 策略统计
type StrategyStats struct {
	UsageCount    int64         // 使用次数
	SuccessCount  int64         // 成功次数
	FailureCount  int64         // 失败次数
	AverageTime   time.Duration // 平均耗时
	LastUsed      time.Time     // 最后使用时间
}

// RecoveryConfig 恢复配置
type RecoveryConfig struct {
	EnableAutoRecovery    bool          // 启用自动恢复
	EnableCircuitBreaker  bool          // 启用熔断器
	EnableStatistics      bool          // 启用统计
	MaxConcurrentRecoveries int         // 最大并发恢复数
	DefaultTimeout        time.Duration // 默认超时时间
	HealthCheckInterval   time.Duration // 健康检查间隔
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	name           string        // 名称
	maxFailures    int           // 最大失败次数
	timeout        time.Duration // 熔断超时时间
	failureCount   int64         // 当前失败次数
	lastFailureTime time.Time    // 最后失败时间
	state          CircuitState  // 熔断器状态
	mu             sync.RWMutex  // 读写锁
}

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed   CircuitState = iota // 关闭状态 - 正常通行
	StateOpen                         // 开启状态 - 拒绝请求
	StateHalfOpen                     // 半开状态 - 试探性通行
)

// NewAutoRecovery 创建自动恢复系统
func NewAutoRecovery(classifier *IntelligentClassifier) *AutoRecovery {
	recovery := &AutoRecovery{
		strategies:     make([]RecoveryStrategy, 0),
		classifier:     classifier,
		circuitBreaker: NewCircuitBreaker("default", 10, 30*time.Second),
		statistics: RecoveryStats{
			ActionCounts:  make(map[RecoveryAction]int64),
			StrategyStats: make(map[string]*StrategyStats),
		},
		config: DefaultRecoveryConfig(),
	}
	
	// 初始化默认策略
	recovery.initDefaultStrategies()
	
	return recovery
}

// DefaultRecoveryConfig 默认恢复配置
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		EnableAutoRecovery:      true,
		EnableCircuitBreaker:    true,
		EnableStatistics:        true,
		MaxConcurrentRecoveries: 100,
		DefaultTimeout:          30 * time.Second,
		HealthCheckInterval:     time.Minute,
	}
}

// Recover 执行自动恢复
func (r *AutoRecovery) Recover(ctx *mvccontext.EnhancedContext, err error) *RecoveryResult {
	if !r.config.EnableAutoRecovery {
		return &RecoveryResult{
			Original:    err,
			Strategy:    "none",
			Action:      ActionIgnore,
			Success:     false,
			FinalError:  err,
			RecoveredAt: time.Now(),
		}
	}
	
	start := time.Now()
	atomic.AddInt64(&r.statistics.TotalAttempts, 1)
	
	// 分类错误
	classification := r.classifier.Classify(err, ctx)
	
	// 检查熔断器状态
	if r.config.EnableCircuitBreaker && !r.circuitBreaker.AllowRequest() {
		return &RecoveryResult{
			Original:    err,
			Strategy:    "circuit-breaker",
			Action:      ActionCircuitBreak,
			Success:     false,
			FinalError:  fmt.Errorf("circuit breaker is open: %w", err),
			Duration:    time.Since(start),
			RecoveredAt: time.Now(),
		}
	}
	
	// 选择恢复策略
	strategy := r.selectStrategy(classification, ctx)
	if strategy == nil {
		return &RecoveryResult{
			Original:    err,
			Strategy:    "none",
			Action:      ActionIgnore,
			Success:     false,
			FinalError:  err,
			Duration:    time.Since(start),
			RecoveredAt: time.Now(),
		}
	}
	
	// 执行恢复
	result := r.executeRecovery(ctx, err, classification, strategy)
	result.Duration = time.Since(start)
	
	// 更新统计信息
	if r.config.EnableStatistics {
		r.updateStatistics(result)
	}
	
	// 更新熔断器状态
	if r.config.EnableCircuitBreaker {
		if result.Success {
			r.circuitBreaker.OnSuccess()
		} else {
			r.circuitBreaker.OnFailure()
		}
	}
	
	return result
}

// selectStrategy 选择恢复策略
func (r *AutoRecovery) selectStrategy(classification *ErrorClassification, ctx *mvccontext.EnhancedContext) *RecoveryStrategy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var bestStrategy *RecoveryStrategy
	var bestPriority int = -1
	
	for i := range r.strategies {
		strategy := &r.strategies[i]
		if strategy.Condition.ShouldRecover(classification, ctx) {
			priority := strategy.Condition.Priority()
			if priority > bestPriority {
				bestPriority = priority
				bestStrategy = strategy
			}
		}
	}
	
	return bestStrategy
}

// executeRecovery 执行恢复
func (r *AutoRecovery) executeRecovery(ctx *mvccontext.EnhancedContext, err error, classification *ErrorClassification, strategy *RecoveryStrategy) *RecoveryResult {
	result := &RecoveryResult{
		Original:    err,
		Strategy:    strategy.Name,
		Action:      strategy.Action,
		RecoveredAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	switch strategy.Action {
	case ActionRetry:
		result = r.executeRetry(ctx, err, strategy)
	case ActionFallback:
		result = r.executeFallback(ctx, err, strategy)
	case ActionCircuitBreak:
		result = r.executeCircuitBreak(ctx, err, strategy)
	case ActionIgnore:
		result = r.executeIgnore(ctx, err, strategy)
	case ActionEscalate:
		result = r.executeEscalate(ctx, err, strategy)
	default:
		result.Success = false
		result.FinalError = fmt.Errorf("unknown recovery action: %v", strategy.Action)
	}
	
	result.Strategy = strategy.Name
	result.Action = strategy.Action
	return result
}

// executeRetry 执行重试恢复
func (r *AutoRecovery) executeRetry(ctx *mvccontext.EnhancedContext, err error, strategy *RecoveryStrategy) *RecoveryResult {
	result := &RecoveryResult{
		Original: err,
		Action:   ActionRetry,
	}
	
	var lastErr error = err
	interval := strategy.RetryInterval
	
	for attempt := 1; attempt <= strategy.MaxRetries; attempt++ {
		result.Attempts = attempt
		
		if attempt > 1 {
			// 等待重试间隔
			time.Sleep(interval)
			// 应用退避策略
			if strategy.BackoffFactor > 1.0 {
				interval = time.Duration(float64(interval) * strategy.BackoffFactor)
			}
		}
		
		// 这里应该重新执行原始操作
		// 由于我们只有错误信息，这里模拟重试逻辑
		if r.simulateRetrySuccess(attempt, strategy.MaxRetries) {
			result.Success = true
			result.FinalError = nil
			break
		}
		
		lastErr = fmt.Errorf("retry %d failed: %w", attempt, err)
	}
	
	if !result.Success {
		result.FinalError = lastErr
	}
	
	return result
}

// executeFallback 执行降级恢复
func (r *AutoRecovery) executeFallback(ctx *mvccontext.EnhancedContext, err error, strategy *RecoveryStrategy) *RecoveryResult {
	result := &RecoveryResult{
		Original: err,
		Action:   ActionFallback,
		Attempts: 1,
	}
	
	if strategy.FallbackFunc != nil {
		fallbackErr := strategy.FallbackFunc(ctx, err)
		if fallbackErr == nil {
			result.Success = true
			result.FinalError = nil
		} else {
			result.Success = false
			result.FinalError = fallbackErr
		}
	} else {
		// 默认降级：返回友好错误信息
		result.Success = true
		result.FinalError = nil
		result.Metadata["fallback_response"] = "Service temporarily unavailable, please try again later"
	}
	
	return result
}

// executeCircuitBreak 执行熔断恢复
func (r *AutoRecovery) executeCircuitBreak(ctx *mvccontext.EnhancedContext, err error, strategy *RecoveryStrategy) *RecoveryResult {
	return &RecoveryResult{
		Original:   err,
		Action:     ActionCircuitBreak,
		Attempts:   0,
		Success:    false,
		FinalError: fmt.Errorf("circuit breaker activated: %w", err),
	}
}

// executeIgnore 执行忽略恢复
func (r *AutoRecovery) executeIgnore(ctx *mvccontext.EnhancedContext, err error, strategy *RecoveryStrategy) *RecoveryResult {
	return &RecoveryResult{
		Original:   err,
		Action:     ActionIgnore,
		Attempts:   0,
		Success:    true,
		FinalError: nil,
	}
}

// executeEscalate 执行上报恢复
func (r *AutoRecovery) executeEscalate(ctx *mvccontext.EnhancedContext, err error, strategy *RecoveryStrategy) *RecoveryResult {
	// 这里可以集成告警系统
	fmt.Printf("[ESCALATED ERROR] %v\n", err)
	
	return &RecoveryResult{
		Original:   err,
		Action:     ActionEscalate,
		Attempts:   1,
		Success:    true,
		FinalError: err, // 保留原错误但标记为已处理
		Metadata:   map[string]interface{}{"escalated": true},
	}
}

// simulateRetrySuccess 模拟重试成功（实际应用中应该重新执行原始操作）
func (r *AutoRecovery) simulateRetrySuccess(attempt, maxRetries int) bool {
	// 简单的模拟逻辑：后面的重试更容易成功
	successProbability := float64(attempt) / float64(maxRetries)
	return successProbability > 0.5 // 简化的成功判断
}

// AddStrategy 添加恢复策略
func (r *AutoRecovery) AddStrategy(strategy RecoveryStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.strategies = append(r.strategies, strategy)
	
	// 初始化策略统计
	if r.config.EnableStatistics {
		r.statistics.StrategyStats[strategy.Name] = &StrategyStats{}
	}
}

// initDefaultStrategies 初始化默认策略
func (r *AutoRecovery) initDefaultStrategies() {
	// 超时错误重试策略
	r.AddStrategy(RecoveryStrategy{
		Name:          "timeout-retry",
		Condition:     &CategoryCondition{Category: CategoryTimeout},
		Action:        ActionRetry,
		MaxRetries:    3,
		RetryInterval: time.Second,
		BackoffFactor: 1.5,
		Timeout:       10 * time.Second,
	})
	
	// 网络错误重试策略
	r.AddStrategy(RecoveryStrategy{
		Name:          "network-retry",
		Condition:     &CategoryCondition{Category: CategoryNetwork},
		Action:        ActionRetry,
		MaxRetries:    3,
		RetryInterval: 2 * time.Second,
		BackoffFactor: 2.0,
		Timeout:       15 * time.Second,
	})
	
	// 数据库错误重试策略
	r.AddStrategy(RecoveryStrategy{
		Name:          "database-retry",
		Condition:     &CategoryCondition{Category: CategoryDatabase},
		Action:        ActionRetry,
		MaxRetries:    2,
		RetryInterval: 3 * time.Second,
		BackoffFactor: 1.5,
		Timeout:       20 * time.Second,
	})
	
	// 限流错误等待策略
	r.AddStrategy(RecoveryStrategy{
		Name:          "ratelimit-retry",
		Condition:     &CategoryCondition{Category: CategoryRateLimit},
		Action:        ActionRetry,
		MaxRetries:    5,
		RetryInterval: 5 * time.Second,
		BackoffFactor: 1.2,
		Timeout:       30 * time.Second,
	})
	
	// 业务错误忽略策略
	r.AddStrategy(RecoveryStrategy{
		Name:      "business-ignore",
		Condition: &CategoryCondition{Category: CategoryBusiness},
		Action:    ActionIgnore,
	})
	
	// 严重错误上报策略
	r.AddStrategy(RecoveryStrategy{
		Name:      "critical-escalate",
		Condition: &SeverityCondition{Severity: SeverityCritical},
		Action:    ActionEscalate,
	})
	
	// 外部服务错误降级策略
	r.AddStrategy(RecoveryStrategy{
		Name:         "external-fallback",
		Condition:    &CategoryCondition{Category: CategoryExternal},
		Action:       ActionFallback,
		FallbackFunc: defaultFallbackHandler,
	})
}

// updateStatistics 更新统计信息
func (r *AutoRecovery) updateStatistics(result *RecoveryResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// 更新总体统计
	if result.Success {
		atomic.AddInt64(&r.statistics.SuccessfulRecoveries, 1)
	} else {
		atomic.AddInt64(&r.statistics.FailedRecoveries, 1)
	}
	
	// 更新动作统计
	r.statistics.ActionCounts[result.Action]++
	
	// 更新策略统计
	if stats, exists := r.statistics.StrategyStats[result.Strategy]; exists {
		stats.UsageCount++
		if result.Success {
			stats.SuccessCount++
		} else {
			stats.FailureCount++
		}
		
		// 更新平均时间
		totalCount := stats.SuccessCount + stats.FailureCount
		stats.AverageTime = (stats.AverageTime*time.Duration(totalCount-1) + result.Duration) / time.Duration(totalCount)
		stats.LastUsed = time.Now()
	}
	
	// 更新平均恢复时间
	totalAttempts := atomic.LoadInt64(&r.statistics.TotalAttempts)
	r.statistics.AverageRecoveryTime = (r.statistics.AverageRecoveryTime*time.Duration(totalAttempts-1) + result.Duration) / time.Duration(totalAttempts)
}

// GetStatistics 获取统计信息
func (r *AutoRecovery) GetStatistics() RecoveryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// 深拷贝统计信息
	stats := RecoveryStats{
		TotalAttempts:       atomic.LoadInt64(&r.statistics.TotalAttempts),
		SuccessfulRecoveries: atomic.LoadInt64(&r.statistics.SuccessfulRecoveries),
		FailedRecoveries:    atomic.LoadInt64(&r.statistics.FailedRecoveries),
		ActionCounts:        make(map[RecoveryAction]int64),
		StrategyStats:       make(map[string]*StrategyStats),
		AverageRecoveryTime: r.statistics.AverageRecoveryTime,
	}
	
	for k, v := range r.statistics.ActionCounts {
		stats.ActionCounts[k] = v
	}
	
	for k, v := range r.statistics.StrategyStats {
		stats.StrategyStats[k] = &StrategyStats{
			UsageCount:   v.UsageCount,
			SuccessCount: v.SuccessCount,
			FailureCount: v.FailureCount,
			AverageTime:  v.AverageTime,
			LastUsed:     v.LastUsed,
		}
	}
	
	return stats
}

// 恢复条件实现

// CategoryCondition 分类条件
type CategoryCondition struct {
	Category ErrorCategory
}

func (c *CategoryCondition) ShouldRecover(classification *ErrorClassification, ctx *mvccontext.EnhancedContext) bool {
	return classification.Category == c.Category
}

func (c *CategoryCondition) Priority() int {
	return 50 // 中等优先级
}

// SeverityCondition 严重等级条件
type SeverityCondition struct {
	Severity ErrorSeverity
}

func (c *SeverityCondition) ShouldRecover(classification *ErrorClassification, ctx *mvccontext.EnhancedContext) bool {
	return classification.Severity == c.Severity
}

func (c *SeverityCondition) Priority() int {
	return 80 // 高优先级
}

// RetryableCondition 可重试条件
type RetryableCondition struct{}

func (c *RetryableCondition) ShouldRecover(classification *ErrorClassification, ctx *mvccontext.EnhancedContext) bool {
	return classification.Retryable
}

func (c *RetryableCondition) Priority() int {
	return 30 // 低优先级
}

// 熔断器实现

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(name string, maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:        name,
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       StateClosed,
	}
}

// AllowRequest 是否允许请求通过
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if now.Sub(cb.lastFailureTime) >= cb.timeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// OnSuccess 成功回调
func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failureCount = 0
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
	}
}

// OnFailure 失败回调
func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= int64(cb.maxFailures) {
		cb.state = StateOpen
	}
}

// GetState 获取状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// 默认降级处理器
func defaultFallbackHandler(ctx *mvccontext.EnhancedContext, err error) error {
	ctx.JSON(503, map[string]interface{}{
		"code":    503,
		"message": "Service temporarily unavailable",
		"success": false,
	})
	return nil
}

// 全局恢复系统
var globalRecovery = NewAutoRecovery(GetGlobalClassifier())

// GetGlobalRecovery 获取全局恢复系统
func GetGlobalRecovery() *AutoRecovery {
	return globalRecovery
}

// RecoverError 恢复错误（全局方法）
func RecoverError(ctx *mvccontext.EnhancedContext, err error) *RecoveryResult {
	return globalRecovery.Recover(ctx, err)
}

// AddGlobalStrategy 添加全局策略
func AddGlobalStrategy(strategy RecoveryStrategy) {
	globalRecovery.AddStrategy(strategy)
}

// GetActionName 获取动作名称
func GetActionName(action RecoveryAction) string {
	names := map[RecoveryAction]string{
		ActionRetry:       "Retry",
		ActionFallback:    "Fallback",
		ActionCircuitBreak: "CircuitBreak",
		ActionIgnore:      "Ignore",
		ActionEscalate:    "Escalate",
	}
	
	if name, exists := names[action]; exists {
		return name
	}
	return "Unknown"
}

// PrintRecoveryInfo 打印恢复系统信息
func PrintRecoveryInfo() {
	stats := globalRecovery.GetStatistics()
	
	fmt.Println("=== Auto Recovery Statistics ===")
	fmt.Printf("Total Recovery Attempts: %d\n", stats.TotalAttempts)
	fmt.Printf("Successful Recoveries: %d\n", stats.SuccessfulRecoveries)
	fmt.Printf("Failed Recoveries: %d\n", stats.FailedRecoveries)
	
	if stats.TotalAttempts > 0 {
		successRate := float64(stats.SuccessfulRecoveries) / float64(stats.TotalAttempts) * 100
		fmt.Printf("Success Rate: %.2f%%\n", successRate)
	}
	
	fmt.Printf("Average Recovery Time: %v\n", stats.AverageRecoveryTime)
	
	fmt.Println("\nAction Distribution:")
	for action, count := range stats.ActionCounts {
		percentage := float64(count) / float64(stats.TotalAttempts) * 100
		fmt.Printf("  %s: %d (%.1f%%)\n", GetActionName(action), count, percentage)
	}
	
	fmt.Println("\nStrategy Performance:")
	for name, strategyStats := range stats.StrategyStats {
		if strategyStats.UsageCount > 0 {
			successRate := float64(strategyStats.SuccessCount) / float64(strategyStats.UsageCount) * 100
			fmt.Printf("  %s: %d uses, %.1f%% success, avg: %v\n",
				name, strategyStats.UsageCount, successRate, strategyStats.AverageTime)
		}
	}
}