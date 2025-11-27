# 🎨 YYHertz 自定义错误处理器开发指南

本文档将详细介绍如何开发和使用自定义错误处理器，满足特定业务需求。

## 🏗️ 错误处理器架构

### 核心接口

YYHertz的错误处理器基于以下核心接口：

```go
// ErrorHandler 错误处理器接口
type ErrorHandler interface {
    Handle(ctx *Context, statusCode int, err error) error
    CanHandle(statusCode int, err error) bool
    Priority() int
}

// ErrorHandlerFunc 统一的错误处理函数类型
type ErrorHandlerFunc func(ctx *Context, statusCode int, err error) error
```

### 处理器特性

| 特性 | 说明 | 用途 |
|------|------|------|
| **Handle** | 执行错误处理逻辑 | 生成响应、记录日志、发送通知等 |
| **CanHandle** | 判断是否能处理该错误 | 错误分发、处理器选择 |
| **Priority** | 处理器优先级 | 数值越小优先级越高 |

## 🛠️ 创建自定义处理器

### 1. 实现ErrorHandler接口

#### 基础结构体处理器

```go
// APIErrorHandler API接口错误处理器
type APIErrorHandler struct {
    logger    *logger.Logger
    config    *APIErrorConfig
    notifier  NotificationService
}

type APIErrorConfig struct {
    ShowDetailedError bool              `json:"show_detailed_error"`
    EnableNotification bool             `json:"enable_notification"`
    NotificationThreshold int           `json:"notification_threshold"`
    ResponseFormat     string           `json:"response_format"` // json, xml, yaml
    EnableCORS        bool              `json:"enable_cors"`
    ErrorMapping      map[string]int    `json:"error_mapping"`
}

func NewAPIErrorHandler(config *APIErrorConfig) *APIErrorHandler {
    return &APIErrorHandler{
        logger:   logger.GetLogger("api_error"),
        config:   config,
        notifier: NewNotificationService(),
    }
}

func (h *APIErrorHandler) Handle(ctx *errors.Context, statusCode int, err error) error {
    // 记录错误日志
    h.logError(ctx, statusCode, err)
    
    // 发送告警通知
    if h.shouldNotify(statusCode, err) {
        go h.sendNotification(ctx, statusCode, err)
    }
    
    // 生成API响应
    return h.generateResponse(ctx, statusCode, err)
}

func (h *APIErrorHandler) CanHandle(statusCode int, err error) bool {
    // 只处理API相关的错误
    if strings.HasPrefix(ctx.Path(), "/api/") {
        return true
    }
    
    // 或者基于错误类型判断
    if _, ok := err.(*APIError); ok {
        return true
    }
    
    return false
}

func (h *APIErrorHandler) Priority() int {
    return 50 // 中等优先级
}

// 私有方法实现
func (h *APIErrorHandler) logError(ctx *errors.Context, statusCode int, err error) {
    logLevel := h.determineLogLevel(statusCode)
    
    h.logger.WithFields(map[string]any{
        "status_code":   statusCode,
        "path":         string(ctx.Path()),
        "method":       string(ctx.Method()),
        "user_agent":   string(ctx.UserAgent()),
        "remote_addr":  ctx.RemoteAddr().String(),
        "error":        err.Error(),
        "timestamp":    time.Now().Unix(),
    }).Log(logLevel, "API error occurred")
}

func (h *APIErrorHandler) shouldNotify(statusCode int, err error) bool {
    if !h.config.EnableNotification {
        return false
    }
    
    // 只对5xx错误发送通知
    return statusCode >= 500 && statusCode < 600
}

func (h *APIErrorHandler) sendNotification(ctx *errors.Context, statusCode int, err error) {
    notification := &ErrorNotification{
        Title:      fmt.Sprintf("API错误 - %d", statusCode),
        Message:    err.Error(),
        StatusCode: statusCode,
        Path:       string(ctx.Path()),
        Method:     string(ctx.Method()),
        Timestamp:  time.Now(),
    }
    
    h.notifier.Send(notification)
}

func (h *APIErrorHandler) generateResponse(ctx *errors.Context, statusCode int, err error) error {
    response := h.buildErrorResponse(statusCode, err)
    
    // 设置CORS头
    if h.config.EnableCORS {
        ctx.Header("Access-Control-Allow-Origin", "*")
        ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
    }
    
    // 根据配置的响应格式输出
    switch h.config.ResponseFormat {
    case "xml":
        ctx.XML(statusCode, response)
    case "yaml":
        ctx.YAML(statusCode, response)
    default:
        ctx.JSON(statusCode, response)
    }
    
    return nil
}

func (h *APIErrorHandler) buildErrorResponse(statusCode int, err error) map[string]any {
    response := map[string]any{
        "code":      statusCode,
        "success":   false,
        "timestamp": time.Now().Unix(),
    }
    
    // 自定义错误消息映射
    if mappedCode, exists := h.config.ErrorMapping[err.Error()]; exists {
        response["business_code"] = mappedCode
    }
    
    // 根据错误类型设置不同的消息
    switch {
    case statusCode >= 400 && statusCode < 500:
        response["message"] = h.getClientErrorMessage(statusCode, err)
    case statusCode >= 500:
        response["message"] = h.getServerErrorMessage(statusCode, err)
    default:
        response["message"] = http.StatusText(statusCode)
    }
    
    // 开发环境显示详细错误
    if h.config.ShowDetailedError && err != nil {
        response["error_detail"] = err.Error()
        response["error_type"] = fmt.Sprintf("%T", err)
    }
    
    return response
}

func (h *APIErrorHandler) determineLogLevel(statusCode int) logger.Level {
    switch {
    case statusCode >= 500:
        return logger.ErrorLevel
    case statusCode >= 400:
        return logger.WarnLevel
    default:
        return logger.InfoLevel
    }
}

func (h *APIErrorHandler) getClientErrorMessage(statusCode int, err error) string {
    messages := map[int]string{
        400: "请求参数有误，请检查后重试",
        401: "请先登录后再访问该资源",
        403: "您没有访问该资源的权限",
        404: "请求的资源不存在",
        405: "不支持该请求方法",
        409: "资源冲突，请重新操作",
        422: "请求数据验证失败",
        429: "请求过于频繁，请稍后再试",
    }
    
    if msg, exists := messages[statusCode]; exists {
        return msg
    }
    
    return "客户端请求错误"
}

func (h *APIErrorHandler) getServerErrorMessage(statusCode int, err error) string {
    return "服务器内部错误，请稍后重试"
}
```

#### 注册和使用

```go
func setupAPIErrorHandler() {
    config := &APIErrorConfig{
        ShowDetailedError:     false,
        EnableNotification:    true,
        NotificationThreshold: 500,
        ResponseFormat:       "json",
        EnableCORS:          true,
        ErrorMapping: map[string]int{
            "user not found":     10001,
            "invalid password":   10002,
            "permission denied":  10003,
        },
    }
    
    apiHandler := NewAPIErrorHandler(config)
    
    // 注册到全局注册器
    errors.RegisterErrorHandler(500, apiHandler)
    errors.RegisterErrorHandler(404, apiHandler)
    errors.RegisterErrorHandler(401, apiHandler)
}
```

### 2. 使用函数式处理器

#### 快速创建处理器

```go
// 为特定业务创建简单处理器
func setupBusinessHandlers() {
    // 用户相关错误处理器
    errors.RegisterFunc(404, "user-not-found", 100,
        func(statusCode int, err error) bool {
            return strings.Contains(err.Error(), "user") && statusCode == 404
        },
        func(ctx *errors.Context, statusCode int, err error) error {
            response := map[string]any{
                "code":    10404,
                "message": "用户不存在",
                "tips":    "请检查用户ID是否正确",
                "success": false,
            }
            
            ctx.JSON(404, response)
            return nil
        },
    )
    
    // 订单相关错误处理器
    errors.RegisterFunc(400, "order-validation", 200,
        func(statusCode int, err error) bool {
            return strings.Contains(strings.ToLower(err.Error()), "order") && statusCode == 400
        },
        func(ctx *errors.Context, statusCode int, err error) error {
            response := map[string]any{
                "code":    20400,
                "message": "订单参数验证失败",
                "tips":    "请检查订单信息是否完整",
                "success": false,
            }
            
            ctx.JSON(400, response)
            return nil
        },
    )
    
    // 支付相关错误处理器
    errors.RegisterFunc(402, "payment-required", 150,
        func(statusCode int, err error) bool {
            return strings.Contains(strings.ToLower(err.Error()), "payment")
        },
        func(ctx *errors.Context, statusCode int, err error) error {
            response := map[string]any{
                "code":    30402,
                "message": "支付失败",
                "tips":    "请检查账户余额或支付方式",
                "redirect": "/payment",
                "success": false,
            }
            
            ctx.JSON(402, response)
            return nil
        },
    )
}
```

#### 链式处理器

```go
// 创建处理器链，按顺序尝试处理
func setupChainHandlers() {
    // 第一级：业务特定处理器（高优先级）
    errors.Register(500, &BusinessSpecificHandler{priority: 10})
    
    // 第二级：通用业务处理器（中优先级）
    errors.Register(500, &GeneralBusinessHandler{priority: 50})
    
    // 第三级：系统默认处理器（低优先级）
    errors.Register(500, &SystemDefaultHandler{priority: 100})
}

type BusinessSpecificHandler struct {
    priority int
}

func (h *BusinessSpecificHandler) Handle(ctx *errors.Context, statusCode int, err error) error {
    // 只处理特定业务错误
    if businessErr, ok := err.(*BusinessError); ok {
        return h.handleBusinessError(ctx, statusCode, businessErr)
    }
    
    // 不能处理，让下一个处理器尝试
    return fmt.Errorf("cannot handle this error type")
}

func (h *BusinessSpecificHandler) CanHandle(statusCode int, err error) bool {
    _, ok := err.(*BusinessError)
    return ok && statusCode == 500
}

func (h *BusinessSpecificHandler) Priority() int {
    return h.priority
}
```

### 3. 中间件式处理器

#### 预处理和后处理

```go
// MiddlewareErrorHandler 中间件式错误处理器
type MiddlewareErrorHandler struct {
    next      errors.ErrorHandler
    preHooks  []PreProcessHook
    postHooks []PostProcessHook
}

type PreProcessHook func(ctx *errors.Context, statusCode int, err error) error
type PostProcessHook func(ctx *errors.Context, statusCode int, err error, result error)

func NewMiddlewareErrorHandler(next errors.ErrorHandler) *MiddlewareErrorHandler {
    return &MiddlewareErrorHandler{
        next:      next,
        preHooks:  make([]PreProcessHook, 0),
        postHooks: make([]PostProcessHook, 0),
    }
}

func (h *MiddlewareErrorHandler) AddPreHook(hook PreProcessHook) {
    h.preHooks = append(h.preHooks, hook)
}

func (h *MiddlewareErrorHandler) AddPostHook(hook PostProcessHook) {
    h.postHooks = append(h.postHooks, hook)
}

func (h *MiddlewareErrorHandler) Handle(ctx *errors.Context, statusCode int, err error) error {
    // 执行预处理钩子
    for _, hook := range h.preHooks {
        if hookErr := hook(ctx, statusCode, err); hookErr != nil {
            return hookErr
        }
    }
    
    // 执行核心处理逻辑
    result := h.next.Handle(ctx, statusCode, err)
    
    // 执行后处理钩子
    for _, hook := range h.postHooks {
        hook(ctx, statusCode, err, result)
    }
    
    return result
}

func (h *MiddlewareErrorHandler) CanHandle(statusCode int, err error) bool {
    return h.next.CanHandle(statusCode, err)
}

func (h *MiddlewareErrorHandler) Priority() int {
    return h.next.Priority()
}

// 使用示例
func setupMiddlewareHandler() {
    baseHandler := &APIErrorHandler{}
    middlewareHandler := NewMiddlewareErrorHandler(baseHandler)
    
    // 添加请求日志钩子
    middlewareHandler.AddPreHook(func(ctx *errors.Context, statusCode int, err error) error {
        log.Printf("Processing error: %d - %v", statusCode, err)
        return nil
    })
    
    // 添加性能监控钩子
    middlewareHandler.AddPreHook(func(ctx *errors.Context, statusCode int, err error) error {
        ctx.Set("error_start_time", time.Now())
        return nil
    })
    
    // 添加响应后钩子
    middlewareHandler.AddPostHook(func(ctx *errors.Context, statusCode int, err error, result error) {
        if startTime, exists := ctx.Get("error_start_time"); exists {
            duration := time.Since(startTime.(time.Time))
            log.Printf("Error handled in %v", duration)
        }
    })
    
    errors.Register(500, middlewareHandler)
}
```

## 🎯 高级错误处理模式

### 1. 错误分类处理器

```go
// ClassifiedErrorHandler 基于错误分类的处理器
type ClassifiedErrorHandler struct {
    classifier errors.Classifier
    handlers   map[errors.ErrorCategory]errors.ErrorHandler
}

func NewClassifiedErrorHandler() *ClassifiedErrorHandler {
    return &ClassifiedErrorHandler{
        classifier: errors.GetGlobalClassifier(),
        handlers:   make(map[errors.ErrorCategory]errors.ErrorHandler),
    }
}

func (h *ClassifiedErrorHandler) RegisterCategoryHandler(category errors.ErrorCategory, handler errors.ErrorHandler) {
    h.handlers[category] = handler
}

func (h *ClassifiedErrorHandler) Handle(ctx *errors.Context, statusCode int, err error) error {
    // 对错误进行分类
    classification := h.classifier.Classify(err, ctx)
    
    // 根据分类选择处理器
    if handler, exists := h.handlers[classification.Category]; exists {
        return handler.Handle(ctx, statusCode, err)
    }
    
    // 使用默认处理逻辑
    return h.handleDefault(ctx, statusCode, err, classification)
}

func (h *ClassifiedErrorHandler) CanHandle(statusCode int, err error) bool {
    classification := h.classifier.Classify(err, nil)
    _, exists := h.handlers[classification.Category]
    return exists
}

func (h *ClassifiedErrorHandler) Priority() int {
    return 30 // 较高优先级
}

func (h *ClassifiedErrorHandler) handleDefault(ctx *errors.Context, statusCode int, err error, classification *errors.ErrorClassification) error {
    response := map[string]any{
        "code":     statusCode,
        "message":  h.getCategoryMessage(classification.Category),
        "category": classification.Category.String(),
        "severity": classification.Severity.String(),
        "success":  false,
    }
    
    ctx.JSON(statusCode, response)
    return nil
}

// 使用示例
func setupClassifiedHandler() {
    handler := NewClassifiedErrorHandler()
    
    // 为不同分类注册专门的处理器
    handler.RegisterCategoryHandler(errors.CategoryDatabase, &DatabaseErrorHandler{})
    handler.RegisterCategoryHandler(errors.CategoryNetwork, &NetworkErrorHandler{})
    handler.RegisterCategoryHandler(errors.CategoryAuthentication, &AuthErrorHandler{})
    
    errors.Register(500, handler)
}
```

### 2. 自适应错误处理器

```go
// AdaptiveErrorHandler 自适应错误处理器
type AdaptiveErrorHandler struct {
    strategies map[string]errors.ErrorHandler
    selector   StrategySelector
    learner    *HandlerLearner
}

type StrategySelector interface {
    SelectStrategy(ctx *errors.Context, statusCode int, err error) string
}

type HandlerLearner struct {
    successRates map[string]float64
    callCounts   map[string]int64
    mutex        sync.RWMutex
}

func NewAdaptiveErrorHandler() *AdaptiveErrorHandler {
    return &AdaptiveErrorHandler{
        strategies: make(map[string]errors.ErrorHandler),
        selector:   &SmartStrategySelector{},
        learner:    &HandlerLearner{
            successRates: make(map[string]float64),
            callCounts:   make(map[string]int64),
        },
    }
}

func (h *AdaptiveErrorHandler) RegisterStrategy(name string, handler errors.ErrorHandler) {
    h.strategies[name] = handler
}

func (h *AdaptiveErrorHandler) Handle(ctx *errors.Context, statusCode int, err error) error {
    // 选择最佳策略
    strategyName := h.selector.SelectStrategy(ctx, statusCode, err)
    
    if strategy, exists := h.strategies[strategyName]; exists {
        start := time.Now()
        result := strategy.Handle(ctx, statusCode, err)
        
        // 记录处理结果用于学习
        h.learner.Record(strategyName, time.Since(start), result == nil)
        
        return result
    }
    
    return fmt.Errorf("no suitable strategy found")
}

type SmartStrategySelector struct{}

func (s *SmartStrategySelector) SelectStrategy(ctx *errors.Context, statusCode int, err error) string {
    // 根据多种因素选择策略
    userAgent := string(ctx.UserAgent())
    path := string(ctx.Path())
    
    switch {
    case strings.Contains(userAgent, "Mobile"):
        return "mobile_optimized"
    case strings.HasPrefix(path, "/api/"):
        return "api_json"
    case strings.HasPrefix(path, "/admin/"):
        return "admin_detailed"
    default:
        return "default_web"
    }
}
```

### 3. 异步错误处理器

```go
// AsyncErrorHandler 异步错误处理器
type AsyncErrorHandler struct {
    queue      chan *ErrorTask
    workers    int
    processors []ErrorProcessor
    metrics    *AsyncMetrics
}

type ErrorTask struct {
    Context    *errors.Context
    StatusCode int
    Error      error
    Timestamp  time.Time
}

type ErrorProcessor interface {
    Process(task *ErrorTask) error
}

func NewAsyncErrorHandler(workers int) *AsyncErrorHandler {
    handler := &AsyncErrorHandler{
        queue:   make(chan *ErrorTask, 1000),
        workers: workers,
        processors: []ErrorProcessor{
            &LogProcessor{},
            &MetricsProcessor{},
            &NotificationProcessor{},
        },
        metrics: &AsyncMetrics{},
    }
    
    // 启动工作协程
    for i := 0; i < workers; i++ {
        go handler.worker()
    }
    
    return handler
}

func (h *AsyncErrorHandler) Handle(ctx *errors.Context, statusCode int, err error) error {
    // 立即响应用户
    h.sendImmediateResponse(ctx, statusCode, err)
    
    // 异步处理其他任务
    task := &ErrorTask{
        Context:    ctx,
        StatusCode: statusCode,
        Error:      err,
        Timestamp:  time.Now(),
    }
    
    select {
    case h.queue <- task:
        // 任务已排队
    default:
        // 队列满，记录丢弃
        h.metrics.IncDropped()
    }
    
    return nil
}

func (h *AsyncErrorHandler) worker() {
    for task := range h.queue {
        h.processTask(task)
    }
}

func (h *AsyncErrorHandler) processTask(task *ErrorTask) {
    defer h.metrics.IncProcessed()
    
    for _, processor := range h.processors {
        if err := processor.Process(task); err != nil {
            log.Printf("Processor error: %v", err)
        }
    }
}

func (h *AsyncErrorHandler) sendImmediateResponse(ctx *errors.Context, statusCode int, err error) {
    response := map[string]any{
        "code":    statusCode,
        "message": "请求处理完成",
        "success": statusCode < 400,
    }
    
    ctx.JSON(statusCode, response)
}

// 处理器实现
type LogProcessor struct{}

func (p *LogProcessor) Process(task *ErrorTask) error {
    log.Printf("Async log: %d - %v at %v", 
        task.StatusCode, task.Error, task.Timestamp)
    return nil
}

type MetricsProcessor struct{}

func (p *MetricsProcessor) Process(task *ErrorTask) error {
    // 发送到监控系统
    metrics.Counter("error_total").
        With("status_code", fmt.Sprintf("%d", task.StatusCode)).
        Inc()
    return nil
}

type NotificationProcessor struct{}

func (p *NotificationProcessor) Process(task *ErrorTask) error {
    if task.StatusCode >= 500 {
        // 发送告警通知
        return sendAlert(task)
    }
    return nil
}
```

## 🧪 测试自定义处理器

### 单元测试

```go
func TestAPIErrorHandler(t *testing.T) {
    // 创建测试处理器
    config := &APIErrorConfig{
        ShowDetailedError: true,
        ResponseFormat:   "json",
    }
    handler := NewAPIErrorHandler(config)
    
    // 创建模拟上下文
    ctx := createMockContext()
    err := errors.New("test error")
    
    // 测试处理器能力
    assert.True(t, handler.CanHandle(500, err))
    assert.Equal(t, 50, handler.Priority())
    
    // 测试处理逻辑
    result := handler.Handle(ctx, 500, err)
    assert.NoError(t, result)
    
    // 验证响应
    response := ctx.GetResponse()
    assert.Equal(t, 500, response.StatusCode)
    assert.Contains(t, response.Body, "test error")
}

func TestHandlerChaining(t *testing.T) {
    registry := errors.NewRegistry(nil)
    
    // 注册多个处理器
    registry.RegisterHandler(500, &HighPriorityHandler{})
    registry.RegisterHandler(500, &LowPriorityHandler{})
    
    // 测试处理器选择
    ctx := createMockContext()
    err := &SpecificError{}
    
    result := registry.HandleError(ctx, 500, err)
    assert.NoError(t, result)
    
    // 验证使用了正确的处理器
    assert.Equal(t, "HighPriorityHandler", ctx.Get("handler_used"))
}

func BenchmarkErrorHandler(b *testing.B) {
    handler := NewAPIErrorHandler(&APIErrorConfig{})
    ctx := createMockContext()
    err := errors.New("benchmark error")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        handler.Handle(ctx, 500, err)
    }
}
```

### 集成测试

```go
func TestErrorHandlerIntegration(t *testing.T) {
    // 设置完整的错误处理链
    app := setupTestApp()
    
    // 注册自定义处理器
    setupAPIErrorHandler()
    setupBusinessHandlers()
    
    // 测试不同错误场景
    testCases := []struct {
        path       string
        method     string
        error      error
        statusCode int
        expected   map[string]any
    }{
        {
            path:       "/api/user/999",
            method:     "GET",
            error:      errors.New("user not found"),
            statusCode: 404,
            expected:   map[string]any{"code": 10404},
        },
        {
            path:       "/api/order",
            method:     "POST",
            error:      errors.New("invalid order data"),
            statusCode: 400,
            expected:   map[string]any{"code": 20400},
        },
    }
    
    for _, tc := range testCases {
        resp := app.Request(tc.method, tc.path).
            WithError(tc.error).
            Expect().
            Status(tc.statusCode)
        
        var response map[string]any
        json.Unmarshal(resp.Body().Raw(), &response)
        
        for key, expectedValue := range tc.expected {
            assert.Equal(t, expectedValue, response[key])
        }
    }
}
```

## 📊 性能优化建议

### 1. 处理器优先级设计

```go
const (
    // 业务特定处理器 - 最高优先级
    PriorityBusinessSpecific = 10
    
    // API接口处理器 - 高优先级
    PriorityAPI = 50
    
    // 通用业务处理器 - 中等优先级
    PriorityGeneralBusiness = 100
    
    // 系统默认处理器 - 低优先级
    PrioritySystemDefault = 500
    
    // 兜底处理器 - 最低优先级
    PriorityFallback = 1000
)
```

### 2. 缓存和池化

```go
// 使用对象池减少内存分配
var errorContextPool = sync.Pool{
    New: func() interface{} {
        return &ErrorContext{
            Metadata: make(map[string]any),
        }
    },
}

func (h *APIErrorHandler) Handle(ctx *errors.Context, statusCode int, err error) error {
    // 从池中获取对象
    errorCtx := errorContextPool.Get().(*ErrorContext)
    defer func() {
        // 清理并归还池
        for k := range errorCtx.Metadata {
            delete(errorCtx.Metadata, k)
        }
        errorContextPool.Put(errorCtx)
    }()
    
    // 使用errorCtx进行处理...
    return nil
}

// 缓存处理结果
type CachedHandler struct {
    base  errors.ErrorHandler
    cache sync.Map
}

func (h *CachedHandler) Handle(ctx *errors.Context, statusCode int, err error) error {
    cacheKey := fmt.Sprintf("%d:%s", statusCode, err.Error())
    
    if cached, ok := h.cache.Load(cacheKey); ok {
        cachedResponse := cached.(map[string]any)
        ctx.JSON(statusCode, cachedResponse)
        return nil
    }
    
    // 执行原处理器并缓存结果
    result := h.base.Handle(ctx, statusCode, err)
    // 缓存逻辑...
    
    return result
}
```

## 🚀 部署和监控

### 生产环境配置

```go
func setupProductionHandlers() {
    // 1. 高性能API处理器
    apiConfig := &APIErrorConfig{
        ShowDetailedError:     false,
        EnableNotification:    true,
        NotificationThreshold: 500,
        ResponseFormat:       "json",
        EnableCORS:          true,
    }
    
    // 2. 异步处理器减少响应时间
    asyncHandler := NewAsyncErrorHandler(10)
    
    // 3. 分类处理器提高准确性
    classifiedHandler := NewClassifiedErrorHandler()
    
    // 注册处理器链
    errors.Register(500, &CachedHandler{
        base: &RateLimitedHandler{
            base: classifiedHandler,
            limit: rate.NewLimiter(100, 1000),
        },
    })
}
```

### 监控指标

```go
// 处理器性能监控
func monitorHandlers() {
    go func() {
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        
        for range ticker.C {
            stats := errors.GetGlobalErrorRegistry().GetMetrics()
            
            for handlerName, metrics := range stats.HandlerPerformance {
                // 发送到监控系统
                prometheus.HandlerCallsTotal.
                    WithLabelValues(handlerName).
                    Set(float64(metrics.CallCount))
                    
                prometheus.HandlerSuccessRate.
                    WithLabelValues(handlerName).
                    Set(float64(metrics.SuccessCount) / float64(metrics.CallCount))
                    
                prometheus.HandlerAvgDuration.
                    WithLabelValues(handlerName).
                    Set(metrics.AverageTime.Seconds())
            }
        }
    }()
}
```

## 📚 最佳实践总结

### 1. 设计原则
- **单一职责**: 每个处理器只负责特定类型的错误
- **优先级合理**: 按照业务重要性设置优先级
- **性能优先**: 避免在处理器中进行重量级操作
- **可测试性**: 确保处理器逻辑易于测试

### 2. 常见错误
- ❌ 在CanHandle中进行复杂逻辑判断
- ❌ 处理器之间存在循环依赖
- ❌ 忽略错误处理器本身的异常
- ❌ 优先级设置不合理导致处理器失效

### 3. 推荐模式
- ✅ 使用责任链模式组织处理器
- ✅ 实现异步处理减少响应时间
- ✅ 基于错误分类选择处理策略
- ✅ 提供丰富的配置选项

## 📖 相关文档

- **[默认处理器详解](default-handlers.md)** - 了解系统内置的处理器
- **[错误页面定制](error-pages.md)** - 定制错误页面展示
- **[错误监控](monitoring.md)** - 建立完善的监控体系
- **[最佳实践](best-practices.md)** - 错误处理的最佳实践

---

> 💡 **提示**: 自定义错误处理器是构建健壮应用的关键，建议根据业务特点设计专门的处理逻辑，并充分测试各种错误场景。