# 🔄 YYHertz 自动恢复机制

本文档详细介绍YYHertz框架的智能错误恢复系统，包括自动重试、熔断器、降级策略等高级恢复机制。

## 🏗️ 恢复系统架构

### 核心组件

```go
// AutoRecovery 自动恢复系统
type AutoRecovery struct {
    config          *RecoveryConfig
    retryManager    *RetryManager
    circuitBreaker  *CircuitBreaker
    fallbackManager *FallbackManager
    healthChecker   *HealthChecker
    metrics         *RecoveryMetrics
    mutex           sync.RWMutex
}

// RecoveryConfig 恢复配置
type RecoveryConfig struct {
    // 基础配置
    EnableAutoRecovery    bool          `json:"enable_auto_recovery"`
    MaxRetries           int           `json:"max_retries"`
    RetryInterval        time.Duration `json:"retry_interval"`
    BackoffMultiplier    float64       `json:"backoff_multiplier"`
    MaxRetryInterval     time.Duration `json:"max_retry_interval"`
    
    // 重试策略配置
    RetryableStatusCodes []int                    `json:"retryable_status_codes"`
    RetryableCategories  []ErrorCategory         `json:"retryable_categories"`
    RetryCondition       func(error) bool        `json:"-"`
    
    // 熔断器配置
    EnableCircuitBreaker bool          `json:"enable_circuit_breaker"`
    FailureThreshold     int           `json:"failure_threshold"`
    RecoveryTimeout      time.Duration `json:"recovery_timeout"`
    HalfOpenRequests     int           `json:"half_open_requests"`
    
    // 降级配置
    EnableFallback       bool                    `json:"enable_fallback"`
    FallbackStrategies   map[string]FallbackFunc `json:"-"`
    
    // 健康检查配置
    EnableHealthCheck    bool          `json:"enable_health_check"`
    HealthCheckInterval  time.Duration `json:"health_check_interval"`
    HealthCheckTimeout   time.Duration `json:"health_check_timeout"`
}

// RetryStrategy 重试策略
type RetryStrategy interface {
    ShouldRetry(attempt int, err error) bool
    GetDelay(attempt int) time.Duration
    GetMaxAttempts() int
}
```

## 🔄 智能重试机制

### 1. 基础重试实现

```go
// RetryManager 重试管理器
type RetryManager struct {
    config    *RecoveryConfig
    strategies map[string]RetryStrategy
    metrics   *RetryMetrics
}

func NewRetryManager(config *RecoveryConfig) *RetryManager {
    return &RetryManager{
        config:     config,
        strategies: make(map[string]RetryStrategy),
        metrics:    NewRetryMetrics(),
    }
}

// ExecuteWithRetry 执行带重试的操作
func (r *RetryManager) ExecuteWithRetry(ctx *Context, operation Operation) error {
    var lastErr error
    
    for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
        // 记录重试尝试
        r.metrics.RecordAttempt()
        
        // 执行操作
        startTime := time.Now()
        err := operation.Execute(ctx)
        duration := time.Since(startTime)
        
        // 记录执行指标
        r.metrics.RecordExecution(duration, err == nil)
        
        if err == nil {
            // 操作成功
            if attempt > 0 {
                r.metrics.RecordSuccess(attempt)
            }
            return nil
        }
        
        lastErr = err
        
        // 检查是否应该重试
        if !r.shouldRetry(attempt, err) {
            break
        }
        
        // 计算重试延迟
        delay := r.calculateDelay(attempt)
        
        // 记录重试
        r.metrics.RecordRetry(attempt, delay)
        
        // 等待后重试
        select {
        case <-time.After(delay):
            continue
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    // 重试失败
    r.metrics.RecordFailure()
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// shouldRetry 判断是否应该重试
func (r *RetryManager) shouldRetry(attempt int, err error) bool {
    if attempt >= r.config.MaxRetries {
        return false
    }
    
    // 使用自定义条件判断
    if r.config.RetryCondition != nil {
        return r.config.RetryCondition(err)
    }
    
    // 检查错误分类
    classification := ClassifyError(err, nil)
    for _, category := range r.config.RetryableCategories {
        if classification.Category == category {
            return true
        }
    }
    
    // 检查HTTP状态码
    if httpErr, ok := err.(HTTPError); ok {
        for _, code := range r.config.RetryableStatusCodes {
            if httpErr.StatusCode() == code {
                return true
            }
        }
    }
    
    return false
}

// calculateDelay 计算重试延迟（指数退避）
func (r *RetryManager) calculateDelay(attempt int) time.Duration {
    delay := r.config.RetryInterval
    
    // 指数退避
    for i := 0; i < attempt; i++ {
        delay = time.Duration(float64(delay) * r.config.BackoffMultiplier)
    }
    
    // 限制最大延迟
    if delay > r.config.MaxRetryInterval {
        delay = r.config.MaxRetryInterval
    }
    
    // 添加随机抖动（避免雷群效应）
    jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
    return delay + jitter
}
```

### 2. 多种重试策略

```go
// ExponentialBackoffStrategy 指数退避策略
type ExponentialBackoffStrategy struct {
    baseDelay    time.Duration
    maxDelay     time.Duration
    multiplier   float64
    maxAttempts  int
}

func (s *ExponentialBackoffStrategy) ShouldRetry(attempt int, err error) bool {
    return attempt < s.maxAttempts
}

func (s *ExponentialBackoffStrategy) GetDelay(attempt int) time.Duration {
    delay := s.baseDelay
    for i := 0; i < attempt; i++ {
        delay = time.Duration(float64(delay) * s.multiplier)
    }
    
    if delay > s.maxDelay {
        delay = s.maxDelay
    }
    
    return delay
}

func (s *ExponentialBackoffStrategy) GetMaxAttempts() int {
    return s.maxAttempts
}

// LinearBackoffStrategy 线性退避策略
type LinearBackoffStrategy struct {
    baseDelay    time.Duration
    increment    time.Duration
    maxDelay     time.Duration
    maxAttempts  int
}

func (s *LinearBackoffStrategy) GetDelay(attempt int) time.Duration {
    delay := s.baseDelay + time.Duration(attempt)*s.increment
    
    if delay > s.maxDelay {
        delay = s.maxDelay
    }
    
    return delay
}

// FixedIntervalStrategy 固定间隔策略
type FixedIntervalStrategy struct {
    interval    time.Duration
    maxAttempts int
}

func (s *FixedIntervalStrategy) GetDelay(attempt int) time.Duration {
    return s.interval
}

// SmartRetryStrategy 智能重试策略
type SmartRetryStrategy struct {
    errorStrategies map[ErrorCategory]RetryStrategy
    defaultStrategy RetryStrategy
}

func (s *SmartRetryStrategy) ShouldRetry(attempt int, err error) bool {
    classification := ClassifyError(err, nil)
    
    if strategy, exists := s.errorStrategies[classification.Category]; exists {
        return strategy.ShouldRetry(attempt, err)
    }
    
    return s.defaultStrategy.ShouldRetry(attempt, err)
}

func (s *SmartRetryStrategy) GetDelay(attempt int) time.Duration {
    // 根据系统负载动态调整延迟
    loadFactor := GetSystemLoad()
    baseDelay := s.defaultStrategy.GetDelay(attempt)
    
    return time.Duration(float64(baseDelay) * (1 + loadFactor))
}
```

## ⚡ 熔断器机制

### 1. 熔断器实现

```go
// CircuitBreaker 熔断器
type CircuitBreaker struct {
    config       *CircuitBreakerConfig
    state        CircuitBreakerState
    failures     int64
    requests     int64
    lastFailTime time.Time
    halfOpenReqs int64
    mutex        sync.RWMutex
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
    StateClosed CircuitBreakerState = iota
    StateOpen
    StateHalfOpen
)

func (s CircuitBreakerState) String() string {
    switch s {
    case StateClosed:
        return "CLOSED"
    case StateOpen:
        return "OPEN"
    case StateHalfOpen:
        return "HALF_OPEN"
    default:
        return "UNKNOWN"
    }
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
    FailureThreshold    int           `json:"failure_threshold"`    // 失败阈值
    RecoveryTimeout     time.Duration `json:"recovery_timeout"`     // 恢复超时
    HalfOpenRequests    int           `json:"half_open_requests"`   // 半开状态请求数
    MinRequests         int           `json:"min_requests"`         // 最小请求数
    FailureRate         float64       `json:"failure_rate"`         // 失败率阈值
}

func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
    return &CircuitBreaker{
        config: config,
        state:  StateClosed,
    }
}

// Execute 执行操作（带熔断保护）
func (cb *CircuitBreaker) Execute(operation func() error) error {
    // 检查熔断器状态
    if !cb.allowRequest() {
        return ErrCircuitBreakerOpen
    }
    
    // 执行操作
    err := operation()
    
    // 记录结果
    cb.recordResult(err)
    
    return err
}

// allowRequest 判断是否允许请求通过
func (cb *CircuitBreaker) allowRequest() bool {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    switch cb.state {
    case StateClosed:
        return true
    case StateOpen:
        return cb.shouldAttemptReset()
    case StateHalfOpen:
        return cb.halfOpenReqs < int64(cb.config.HalfOpenRequests)
    default:
        return false
    }
}

// shouldAttemptReset 判断是否应该尝试重置
func (cb *CircuitBreaker) shouldAttemptReset() bool {
    return time.Since(cb.lastFailTime) >= cb.config.RecoveryTimeout
}

// recordResult 记录操作结果
func (cb *CircuitBreaker) recordResult(err error) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    cb.requests++
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        cb.onFailure()
    } else {
        cb.onSuccess()
    }
}

// onSuccess 处理成功结果
func (cb *CircuitBreaker) onSuccess() {
    switch cb.state {
    case StateHalfOpen:
        cb.halfOpenReqs++
        if cb.halfOpenReqs >= int64(cb.config.HalfOpenRequests) {
            cb.setState(StateClosed)
            cb.reset()
        }
    case StateClosed:
        // 保持关闭状态
    }
}

// onFailure 处理失败结果
func (cb *CircuitBreaker) onFailure() {
    switch cb.state {
    case StateClosed:
        if cb.shouldOpen() {
            cb.setState(StateOpen)
        }
    case StateHalfOpen:
        cb.setState(StateOpen)
    }
}

// shouldOpen 判断是否应该开启熔断
func (cb *CircuitBreaker) shouldOpen() bool {
    if cb.requests < int64(cb.config.MinRequests) {
        return false
    }
    
    failureRate := float64(cb.failures) / float64(cb.requests)
    return failureRate >= cb.config.FailureRate
}

// setState 设置熔断器状态
func (cb *CircuitBreaker) setState(state CircuitBreakerState) {
    if cb.state != state {
        cb.state = state
        cb.onStateChange(state)
    }
}

// onStateChange 状态变更回调
func (cb *CircuitBreaker) onStateChange(newState CircuitBreakerState) {
    switch newState {
    case StateOpen:
        log.Printf("Circuit breaker opened due to failures")
    case StateHalfOpen:
        log.Printf("Circuit breaker entering half-open state")
        cb.halfOpenReqs = 0
    case StateClosed:
        log.Printf("Circuit breaker closed - service recovered")
    }
}

// reset 重置熔断器
func (cb *CircuitBreaker) reset() {
    cb.requests = 0
    cb.failures = 0
    cb.halfOpenReqs = 0
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
    cb.mutex.RLock()
    defer cb.mutex.RUnlock()
    return cb.state
}
```

### 2. 熔断器错误处理

```go
// CircuitBreakerError 熔断器错误
type CircuitBreakerError struct {
    State     CircuitBreakerState
    Message   string
    Timestamp time.Time
}

func (e *CircuitBreakerError) Error() string {
    return fmt.Sprintf("circuit breaker %s: %s", e.State, e.Message)
}

var ErrCircuitBreakerOpen = &CircuitBreakerError{
    State:     StateOpen,
    Message:   "circuit breaker is open",
    Timestamp: time.Now(),
}

// CircuitBreakerHandler 熔断器错误处理器
type CircuitBreakerHandler struct{}

func (h *CircuitBreakerHandler) Handle(ctx *Context, statusCode int, err error) error {
    cbErr, ok := err.(*CircuitBreakerError)
    if !ok {
        return fmt.Errorf("not a circuit breaker error")
    }
    
    response := map[string]any{
        "success":   false,
        "code":      503, // Service Unavailable
        "message":   "服务暂时不可用，请稍后重试",
        "state":     cbErr.State.String(),
        "timestamp": cbErr.Timestamp.Unix(),
        "retry_after": 30, // 建议30秒后重试
    }
    
    ctx.Header("Retry-After", "30")
    ctx.JSON(503, response)
    return nil
}

func (h *CircuitBreakerHandler) CanHandle(statusCode int, err error) bool {
    _, ok := err.(*CircuitBreakerError)
    return ok
}

func (h *CircuitBreakerHandler) Priority() int {
    return 5 // 高优先级
}
```

## 🛡️ 降级策略

### 1. 降级管理器

```go
// FallbackManager 降级管理器
type FallbackManager struct {
    strategies map[string]FallbackFunc
    metrics    *FallbackMetrics
    mutex      sync.RWMutex
}

// FallbackFunc 降级函数
type FallbackFunc func(ctx *Context, err error) error

func NewFallbackManager() *FallbackManager {
    return &FallbackManager{
        strategies: make(map[string]FallbackFunc),
        metrics:    NewFallbackMetrics(),
    }
}

// RegisterFallback 注册降级策略
func (f *FallbackManager) RegisterFallback(name string, strategy FallbackFunc) {
    f.mutex.Lock()
    defer f.mutex.Unlock()
    
    f.strategies[name] = strategy
}

// ExecuteFallback 执行降级策略
func (f *FallbackManager) ExecuteFallback(ctx *Context, strategyName string, originalErr error) error {
    f.mutex.RLock()
    strategy, exists := f.strategies[strategyName]
    f.mutex.RUnlock()
    
    if !exists {
        return fmt.Errorf("fallback strategy not found: %s", strategyName)
    }
    
    // 记录降级执行
    f.metrics.RecordFallback(strategyName)
    
    // 执行降级策略
    start := time.Now()
    err := strategy(ctx, originalErr)
    duration := time.Since(start)
    
    // 记录执行结果
    f.metrics.RecordExecution(strategyName, duration, err == nil)
    
    return err
}

// GetAvailableStrategies 获取可用策略
func (f *FallbackManager) GetAvailableStrategies() []string {
    f.mutex.RLock()
    defer f.mutex.RUnlock()
    
    strategies := make([]string, 0, len(f.strategies))
    for name := range f.strategies {
        strategies = append(strategies, name)
    }
    
    return strategies
}
```

### 2. 常见降级策略

```go
// 缓存降级策略
func CacheFallbackStrategy(cache Cache) FallbackFunc {
    return func(ctx *Context, err error) error {
        // 尝试从缓存获取数据
        cacheKey := generateCacheKey(ctx)
        
        if data, found := cache.Get(cacheKey); found {
            response := map[string]any{
                "success": true,
                "data":    data,
                "source":  "cache",
                "message": "数据来自缓存",
            }
            
            ctx.JSON(200, response)
            return nil
        }
        
        // 缓存也没有，返回默认响应
        return DefaultFallbackStrategy(ctx, err)
    }
}

// 默认数据降级策略
func DefaultDataFallbackStrategy(defaultData any) FallbackFunc {
    return func(ctx *Context, err error) error {
        response := map[string]any{
            "success": true,
            "data":    defaultData,
            "source":  "default",
            "message": "服务异常，返回默认数据",
        }
        
        ctx.JSON(200, response)
        return nil
    }
}

// 静态页面降级策略
func StaticPageFallbackStrategy(templatePath string) FallbackFunc {
    return func(ctx *Context, err error) error {
        data := map[string]any{
            "Error":     err.Error(),
            "Timestamp": time.Now(),
            "Message":   "服务暂时不可用",
        }
        
        ctx.HTML(503, templatePath, data)
        return nil
    }
}

// 重定向降级策略
func RedirectFallbackStrategy(redirectURL string) FallbackFunc {
    return func(ctx *Context, err error) error {
        ctx.Redirect(302, redirectURL)
        return nil
    }
}

// 队列降级策略
func QueueFallbackStrategy(queue MessageQueue) FallbackFunc {
    return func(ctx *Context, err error) error {
        // 将请求放入队列，异步处理
        message := &QueueMessage{
            Path:      string(ctx.Path()),
            Method:    string(ctx.Method()),
            Body:      ctx.Request.Body(),
            Headers:   ctx.Request.Header,
            Timestamp: time.Now(),
        }
        
        if err := queue.Enqueue(message); err != nil {
            return err
        }
        
        response := map[string]any{
            "success": true,
            "message": "请求已排队处理，稍后查看结果",
            "queue_id": message.ID,
        }
        
        ctx.JSON(202, response) // 202 Accepted
        return nil
    }
}
```

## 💊 健康检查系统

### 1. 健康检查器

```go
// HealthChecker 健康检查器
type HealthChecker struct {
    config      *HealthCheckConfig
    endpoints   map[string]HealthEndpoint
    results     map[string]*HealthResult
    mutex       sync.RWMutex
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
    Interval    time.Duration `json:"interval"`
    Timeout     time.Duration `json:"timeout"`
    RetryCount  int           `json:"retry_count"`
    Enabled     bool          `json:"enabled"`
}

// HealthEndpoint 健康检查端点
type HealthEndpoint struct {
    Name        string                      `json:"name"`
    URL         string                      `json:"url"`
    Method      string                      `json:"method"`
    Headers     map[string]string           `json:"headers"`
    CheckFunc   func() *HealthResult        `json:"-"`
    Timeout     time.Duration               `json:"timeout"`
    Critical    bool                        `json:"critical"`
}

// HealthResult 健康检查结果
type HealthResult struct {
    Name        string                 `json:"name"`
    Healthy     bool                   `json:"healthy"`
    Message     string                 `json:"message"`
    Duration    time.Duration          `json:"duration"`
    Timestamp   time.Time              `json:"timestamp"`
    Details     map[string]any         `json:"details,omitempty"`
}

func NewHealthChecker(config *HealthCheckConfig) *HealthChecker {
    return &HealthChecker{
        config:    config,
        endpoints: make(map[string]HealthEndpoint),
        results:   make(map[string]*HealthResult),
    }
}

// RegisterEndpoint 注册健康检查端点
func (h *HealthChecker) RegisterEndpoint(endpoint HealthEndpoint) {
    h.mutex.Lock()
    defer h.mutex.Unlock()
    
    h.endpoints[endpoint.Name] = endpoint
}

// Start 启动健康检查
func (h *HealthChecker) Start() {
    if !h.config.Enabled {
        return
    }
    
    ticker := time.NewTicker(h.config.Interval)
    go func() {
        defer ticker.Stop()
        
        for range ticker.C {
            h.checkAll()
        }
    }()
}

// checkAll 检查所有端点
func (h *HealthChecker) checkAll() {
    h.mutex.RLock()
    endpoints := make(map[string]HealthEndpoint)
    for name, endpoint := range h.endpoints {
        endpoints[name] = endpoint
    }
    h.mutex.RUnlock()
    
    // 并发检查所有端点
    var wg sync.WaitGroup
    for name, endpoint := range endpoints {
        wg.Add(1)
        go func(name string, endpoint HealthEndpoint) {
            defer wg.Done()
            result := h.checkEndpoint(endpoint)
            
            h.mutex.Lock()
            h.results[name] = result
            h.mutex.Unlock()
        }(name, endpoint)
    }
    
    wg.Wait()
}

// checkEndpoint 检查单个端点
func (h *HealthChecker) checkEndpoint(endpoint HealthEndpoint) *HealthResult {
    start := time.Now()
    
    result := &HealthResult{
        Name:      endpoint.Name,
        Timestamp: start,
    }
    
    // 使用自定义检查函数
    if endpoint.CheckFunc != nil {
        customResult := endpoint.CheckFunc()
        customResult.Duration = time.Since(start)
        return customResult
    }
    
    // 默认HTTP检查
    timeout := endpoint.Timeout
    if timeout == 0 {
        timeout = h.config.Timeout
    }
    
    client := &http.Client{Timeout: timeout}
    
    req, err := http.NewRequest(endpoint.Method, endpoint.URL, nil)
    if err != nil {
        result.Healthy = false
        result.Message = fmt.Sprintf("创建请求失败: %v", err)
        result.Duration = time.Since(start)
        return result
    }
    
    // 添加请求头
    for key, value := range endpoint.Headers {
        req.Header.Set(key, value)
    }
    
    resp, err := client.Do(req)
    if err != nil {
        result.Healthy = false
        result.Message = fmt.Sprintf("请求失败: %v", err)
    } else {
        defer resp.Body.Close()
        
        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            result.Healthy = true
            result.Message = "健康检查通过"
        } else {
            result.Healthy = false
            result.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
        }
        
        result.Details = map[string]any{
            "status_code": resp.StatusCode,
            "headers":     resp.Header,
        }
    }
    
    result.Duration = time.Since(start)
    return result
}

// GetResults 获取健康检查结果
func (h *HealthChecker) GetResults() map[string]*HealthResult {
    h.mutex.RLock()
    defer h.mutex.RUnlock()
    
    results := make(map[string]*HealthResult)
    for name, result := range h.results {
        results[name] = result
    }
    
    return results
}

// IsHealthy 判断系统是否健康
func (h *HealthChecker) IsHealthy() bool {
    h.mutex.RLock()
    defer h.mutex.RUnlock()
    
    for name, endpoint := range h.endpoints {
        if endpoint.Critical {
            if result, exists := h.results[name]; exists {
                if !result.Healthy {
                    return false
                }
            } else {
                return false // 关键服务没有检查结果
            }
        }
    }
    
    return true
}
```

### 2. 内置健康检查

```go
// 数据库健康检查
func DatabaseHealthCheck(db *sql.DB) func() *HealthResult {
    return func() *HealthResult {
        start := time.Now()
        result := &HealthResult{
            Name:      "database",
            Timestamp: start,
        }
        
        if err := db.Ping(); err != nil {
            result.Healthy = false
            result.Message = fmt.Sprintf("数据库连接失败: %v", err)
        } else {
            result.Healthy = true
            result.Message = "数据库连接正常"
        }
        
        result.Duration = time.Since(start)
        return result
    }
}

// Redis健康检查
func RedisHealthCheck(redis *redis.Client) func() *HealthResult {
    return func() *HealthResult {
        start := time.Now()
        result := &HealthResult{
            Name:      "redis",
            Timestamp: start,
        }
        
        if err := redis.Ping().Err(); err != nil {
            result.Healthy = false
            result.Message = fmt.Sprintf("Redis连接失败: %v", err)
        } else {
            result.Healthy = true
            result.Message = "Redis连接正常"
        }
        
        result.Duration = time.Since(start)
        return result
    }
}

// 外部API健康检查
func ExternalAPIHealthCheck(apiURL string) func() *HealthResult {
    return func() *HealthResult {
        start := time.Now()
        result := &HealthResult{
            Name:      "external_api",
            Timestamp: start,
        }
        
        client := &http.Client{Timeout: 5 * time.Second}
        resp, err := client.Get(apiURL)
        
        if err != nil {
            result.Healthy = false
            result.Message = fmt.Sprintf("API调用失败: %v", err)
        } else {
            defer resp.Body.Close()
            
            if resp.StatusCode == 200 {
                result.Healthy = true
                result.Message = "API服务正常"
            } else {
                result.Healthy = false
                result.Message = fmt.Sprintf("API返回错误状态: %d", resp.StatusCode)
            }
        }
        
        result.Duration = time.Since(start)
        return result
    }
}
```

## 🚀 完整恢复系统集成

### 1. 统一恢复处理器

```go
// RecoveryHandler 恢复处理器
type RecoveryHandler struct {
    retryManager    *RetryManager
    circuitBreaker  *CircuitBreaker
    fallbackManager *FallbackManager
    healthChecker   *HealthChecker
    config          *RecoveryConfig
}

func NewRecoveryHandler(config *RecoveryConfig) *RecoveryHandler {
    return &RecoveryHandler{
        retryManager:    NewRetryManager(config),
        circuitBreaker:  NewCircuitBreaker(config.CircuitBreakerConfig),
        fallbackManager: NewFallbackManager(),
        healthChecker:   NewHealthChecker(config.HealthCheckConfig),
        config:          config,
    }
}

func (h *RecoveryHandler) Handle(ctx *Context, statusCode int, err error) error {
    // 1. 首先检查系统健康状态
    if !h.healthChecker.IsHealthy() {
        return h.fallbackManager.ExecuteFallback(ctx, "system_unhealthy", err)
    }
    
    // 2. 尝试通过熔断器执行重试
    operation := &RetryableOperation{
        ctx:          ctx,
        statusCode:   statusCode,
        originalErr:  err,
        retryManager: h.retryManager,
    }
    
    circuitErr := h.circuitBreaker.Execute(func() error {
        return h.retryManager.ExecuteWithRetry(ctx, operation)
    })
    
    // 3. 如果熔断器开启或重试失败，执行降级策略
    if circuitErr != nil {
        strategyName := h.selectFallbackStrategy(statusCode, err)
        return h.fallbackManager.ExecuteFallback(ctx, strategyName, circuitErr)
    }
    
    return nil
}

func (h *RecoveryHandler) CanHandle(statusCode int, err error) bool {
    return h.config.EnableAutoRecovery
}

func (h *RecoveryHandler) Priority() int {
    return 1 // 最高优先级
}

// selectFallbackStrategy 选择降级策略
func (h *RecoveryHandler) selectFallbackStrategy(statusCode int, err error) string {
    switch {
    case statusCode >= 500:
        return "server_error_fallback"
    case statusCode == 429:
        return "rate_limit_fallback"
    default:
        return "default_fallback"
    }
}
```

### 2. 生产环境配置

```go
// 生产环境恢复配置
func SetupProductionRecovery() {
    config := &RecoveryConfig{
        // 基础重试配置
        EnableAutoRecovery:   true,
        MaxRetries:          3,
        RetryInterval:       time.Second,
        BackoffMultiplier:   2.0,
        MaxRetryInterval:    time.Minute,
        
        // 可重试的状态码和分类
        RetryableStatusCodes: []int{500, 502, 503, 504},
        RetryableCategories: []ErrorCategory{
            CategoryNetwork,
            CategoryTimeout,
            CategorySystem,
        },
        
        // 熔断器配置
        EnableCircuitBreaker: true,
        FailureThreshold:     10,
        RecoveryTimeout:      30 * time.Second,
        HalfOpenRequests:     3,
        
        // 降级配置
        EnableFallback: true,
        
        // 健康检查配置
        EnableHealthCheck:   true,
        HealthCheckInterval: 30 * time.Second,
        HealthCheckTimeout:  5 * time.Second,
    }
    
    // 创建恢复处理器
    recoveryHandler := NewRecoveryHandler(config)
    
    // 注册降级策略
    recoveryHandler.fallbackManager.RegisterFallback(
        "server_error_fallback",
        CacheFallbackStrategy(GetCache()),
    )
    
    recoveryHandler.fallbackManager.RegisterFallback(
        "rate_limit_fallback",
        QueueFallbackStrategy(GetMessageQueue()),
    )
    
    recoveryHandler.fallbackManager.RegisterFallback(
        "default_fallback",
        DefaultDataFallbackStrategy(map[string]any{
            "message": "服务暂时不可用，请稍后重试",
            "code":    503,
        }),
    )
    
    // 注册健康检查端点
    recoveryHandler.healthChecker.RegisterEndpoint(HealthEndpoint{
        Name:     "database",
        CheckFunc: DatabaseHealthCheck(GetDB()),
        Critical: true,
    })
    
    recoveryHandler.healthChecker.RegisterEndpoint(HealthEndpoint{
        Name:     "redis",
        CheckFunc: RedisHealthCheck(GetRedis()),
        Critical: false,
    })
    
    // 启动健康检查
    recoveryHandler.healthChecker.Start()
    
    // 注册到错误处理系统
    errors.RegisterErrorHandler(500, recoveryHandler)
}
```

## 📊 恢复系统监控

### 监控API

```go
// RecoveryMonitorController 恢复监控控制器
type RecoveryMonitorController struct {
    recoveryHandler *RecoveryHandler
}

// GetRecoveryStats 获取恢复统计
func (c *RecoveryMonitorController) GetRecoveryStats(ctx *Context) {
    stats := map[string]any{
        "retry_stats":         c.recoveryHandler.retryManager.GetStats(),
        "circuit_breaker":     c.getCircuitBreakerStats(),
        "fallback_stats":      c.recoveryHandler.fallbackManager.GetStats(),
        "health_check_results": c.recoveryHandler.healthChecker.GetResults(),
        "overall_health":      c.recoveryHandler.healthChecker.IsHealthy(),
    }
    
    ctx.JSON(200, stats)
}

// GetHealthStatus 获取健康状态
func (c *RecoveryMonitorController) GetHealthStatus(ctx *Context) {
    results := c.recoveryHandler.healthChecker.GetResults()
    overallHealthy := c.recoveryHandler.healthChecker.IsHealthy()
    
    response := map[string]any{
        "healthy":    overallHealthy,
        "timestamp":  time.Now(),
        "checks":     results,
    }
    
    statusCode := 200
    if !overallHealthy {
        statusCode = 503
    }
    
    ctx.JSON(statusCode, response)
}
```

## 📚 相关文档

- **[快速开始](quick-start.md)** - 了解基础错误处理配置
- **[错误监控](monitoring.md)** - 监控恢复系统指标
- **[自定义处理器](custom-handlers.md)** - 开发恢复处理器
- **[最佳实践](best-practices.md)** - 恢复系统最佳实践

---

> 💡 **提示**: 自动恢复机制是构建高可用系统的关键，建议根据业务特点配置合适的重试策略和降级方案，并持续监控恢复效果。