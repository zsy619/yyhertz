# 🏆 YYHertz 错误处理最佳实践

本文档总结了在YYHertz框架中进行错误处理的最佳实践，帮助开发者建立健壮、可维护的错误处理体系。

## 🎯 核心原则

### 1. 分层错误处理

```go
// ✅ 推荐的分层错误处理模式
func (c *UserController) CreateUser() {
    // 1. 参数验证层
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.HandleValidationError(err)
        return
    }
    
    // 2. 业务逻辑层
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        c.HandleBusinessError(err)
        return
    }
    
    // 3. 成功响应
    c.JSONSuccess("用户创建成功", user)
}

// 专门的错误处理方法
func (c *UserController) HandleValidationError(err error) {
    businessErr := NewBusinessError(ErrInvalidParameter).
        WithDetails(map[string]any{"validation_error": err.Error()})
    errors.Handle(c.Ctx, 400, businessErr)
}

func (c *UserController) HandleBusinessError(err error) {
    if businessErr, ok := err.(*CustomBusinessError); ok {
        errors.Handle(c.Ctx, 400, businessErr)
    } else {
        errors.Handle(c.Ctx, 500, err)
    }
}
```

### 2. 错误传播原则

```go
// ✅ 正确的错误传播
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    // 业务验证
    if err := s.validateUserRequest(req); err != nil {
        return nil, err // 直接传播业务错误
    }
    
    // 数据库操作
    user, err := s.userRepo.Create(req)
    if err != nil {
        // 包装系统错误为业务错误
        return nil, NewBusinessError(ErrUserCreationFailed).
            WithDetails(map[string]any{"original_error": err.Error()})
    }
    
    return user, nil
}

// ❌ 错误的做法 - 丢失错误信息
func (s *UserService) BadCreateUser(req *CreateUserRequest) (*User, error) {
    user, err := s.userRepo.Create(req)
    if err != nil {
        return nil, errors.New("创建失败") // 丢失了原始错误信息
    }
    return user, nil
}
```

## 🛠️ 实现最佳实践

### 1. 统一错误码管理

```go
// errors/codes.go - 集中管理错误码
package errors

// 错误码分段规则
const (
    // 系统级错误 (10000-19999)
    SystemErrorBase = 10000
    
    // 用户相关错误 (20000-20999)
    UserErrorBase = 20000
    ErrUserNotFound       = UserErrorBase + 1
    ErrUserAlreadyExists  = UserErrorBase + 2
    ErrInvalidCredentials = UserErrorBase + 3
    
    // 订单相关错误 (21000-21999)
    OrderErrorBase = 21000
    ErrOrderNotFound      = OrderErrorBase + 1
    ErrOrderAlreadyPaid   = OrderErrorBase + 2
    ErrInsufficientStock  = OrderErrorBase + 3
)

// 错误消息映射
var ErrorMessages = map[int]ErrorInfo{
    ErrUserNotFound: {
        Code:        ErrUserNotFound,
        Message:     "用户不存在",
        Description: "系统中未找到指定的用户信息",
        Category:    CategoryValidation,
        Severity:    SeverityMedium,
        Solutions: []string{
            "检查用户ID是否正确",
            "确认用户是否已注册",
        },
    },
    // ... 更多错误定义
}
```

### 2. 错误处理中间件

```go
// ErrorHandlingMiddleware 全局错误处理中间件
func ErrorHandlingMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        defer func() {
            if r := recover(); r != nil {
                var err error
                switch v := r.(type) {
                case error:
                    err = v
                case string:
                    err = errors.New(v)
                default:
                    err = fmt.Errorf("未知panic: %v", v)
                }
                
                // 记录panic错误
                log.Printf("Panic recovered: %+v", err)
                
                // 处理panic作为服务器错误
                HandleError(NewContext(c), 500, err)
                c.Abort()
            }
        }()
        
        c.Next(ctx)
        
        // 检查响应状态码，处理HTTP错误
        if c.Response.StatusCode() >= 400 {
            HandleHTTPError(NewContext(c), c.Response.StatusCode())
        }
    }
}
```

### 3. 结构化错误日志

```go
// 错误日志结构
type ErrorLogEntry struct {
    Timestamp   time.Time          `json:"timestamp"`
    Level       string             `json:"level"`
    RequestID   string             `json:"request_id"`
    UserID      string             `json:"user_id,omitempty"`
    ErrorCode   int                `json:"error_code,omitempty"`
    ErrorType   string             `json:"error_type"`
    Message     string             `json:"message"`
    StackTrace  string             `json:"stack_trace,omitempty"`
    Context     map[string]any     `json:"context"`
    Path        string             `json:"path"`
    Method      string             `json:"method"`
    UserAgent   string             `json:"user_agent"`
    ClientIP    string             `json:"client_ip"`
}

// 结构化错误日志记录器
func LogStructuredError(ctx *Context, err error, level string) {
    entry := ErrorLogEntry{
        Timestamp: time.Now(),
        Level:     level,
        RequestID: ctx.GetString("request_id"),
        Message:   err.Error(),
        Path:      string(ctx.Path()),
        Method:    string(ctx.Method()),
        UserAgent: string(ctx.UserAgent()),
        ClientIP:  ctx.ClientIP(),
        Context:   make(map[string]any),
    }
    
    // 根据错误类型添加特定信息
    switch e := err.(type) {
    case *CustomBusinessError:
        entry.ErrorCode = e.Code
        entry.ErrorType = "business"
        entry.UserID = e.UserID
        entry.Context = e.Details
    case *ValidationError:
        entry.ErrorType = "validation"
        entry.Context["fields"] = e.Fields
    default:
        entry.ErrorType = "system"
        if level == "error" || level == "fatal" {
            entry.StackTrace = getStackTrace()
        }
    }
    
    // 输出JSON格式日志
    logJSON, _ := json.Marshal(entry)
    log.Printf("%s", logJSON)
}
```

### 4. 错误恢复策略

```go
// GracefulErrorRecovery 优雅错误恢复
type GracefulErrorRecovery struct {
    maxRetries      int
    retryInterval   time.Duration
    fallbackHandler FallbackHandler
    circuitBreaker  *CircuitBreaker
}

func (g *GracefulErrorRecovery) Execute(operation Operation) error {
    return g.circuitBreaker.Execute(func() error {
        var lastErr error
        
        for attempt := 0; attempt <= g.maxRetries; attempt++ {
            err := operation.Execute()
            if err == nil {
                return nil
            }
            
            lastErr = err
            
            // 判断是否应该重试
            if !g.shouldRetry(err, attempt) {
                break
            }
            
            // 等待后重试
            time.Sleep(g.calculateDelay(attempt))
        }
        
        // 重试失败，执行降级策略
        if g.fallbackHandler != nil {
            return g.fallbackHandler.Handle(lastErr)
        }
        
        return lastErr
    })
}

func (g *GracefulErrorRecovery) shouldRetry(err error, attempt int) bool {
    // 不重试业务逻辑错误
    if _, ok := err.(*CustomBusinessError); ok {
        return false
    }
    
    // 检查是否为可重试的错误类型
    return isRetryableError(err) && attempt < g.maxRetries
}
```

## 🔧 开发环境最佳实践

### 1. 开发时错误详情

```go
// DevelopmentErrorHandler 开发环境错误处理器
type DevelopmentErrorHandler struct {
    showStackTrace bool
    showSQLQueries bool
}

func (h *DevelopmentErrorHandler) Handle(ctx *Context, statusCode int, err error) error {
    response := map[string]any{
        "success":   false,
        "timestamp": time.Now().Unix(),
        "path":      string(ctx.Path()),
        "method":    string(ctx.Method()),
    }
    
    // 开发环境显示详细错误信息
    if isDevelopment() {
        response["error"] = map[string]any{
            "message": err.Error(),
            "type":    reflect.TypeOf(err).String(),
        }
        
        if h.showStackTrace {
            response["stack_trace"] = getStackTrace()
        }
        
        // 显示请求上下文
        response["context"] = map[string]any{
            "headers":    ctx.Request.Header,
            "query":      ctx.QueryArgs(),
            "user_agent": string(ctx.UserAgent()),
            "client_ip":  ctx.ClientIP(),
        }
        
        // 数据库查询信息
        if h.showSQLQueries && ctx.Get("sql_queries") != nil {
            response["sql_queries"] = ctx.Get("sql_queries")
        }
    } else {
        // 生产环境只显示安全的错误信息
        response["message"] = "服务器内部错误"
        response["code"] = statusCode
    }
    
    ctx.JSON(statusCode, response)
    return nil
}
```

### 2. 测试错误场景

```go
// 错误处理单元测试
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name           string
        setupError     func() error
        expectedStatus int
        expectedCode   int
    }{
        {
            name: "用户不存在错误",
            setupError: func() error {
                return NewBusinessError(ErrUserNotFound)
            },
            expectedStatus: 400,
            expectedCode:   ErrUserNotFound,
        },
        {
            name: "系统错误",
            setupError: func() error {
                return errors.New("database connection failed")
            },
            expectedStatus: 500,
            expectedCode:   0,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 创建测试上下文
            ctx := createTestContext()
            
            // 触发错误
            err := tt.setupError()
            HandleError(ctx, tt.expectedStatus, err)
            
            // 验证响应
            assert.Equal(t, tt.expectedStatus, ctx.Response.StatusCode())
            
            var response map[string]any
            json.Unmarshal(ctx.Response.Body(), &response)
            
            if tt.expectedCode > 0 {
                assert.Equal(t, tt.expectedCode, int(response["code"].(float64)))
            }
        })
    }
}
```

## 🚀 生产环境最佳实践

### 1. 错误监控和告警

```go
// ProductionErrorMonitor 生产环境错误监控
type ProductionErrorMonitor struct {
    alertManager AlertManager
    metrics      MetricsCollector
    notifier     NotificationService
}

func (m *ProductionErrorMonitor) OnError(ctx *Context, err error) {
    // 1. 收集错误指标
    m.collectErrorMetrics(err)
    
    // 2. 检查是否需要告警
    if m.shouldAlert(err) {
        go m.sendAlert(ctx, err)
    }
    
    // 3. 记录错误到监控系统
    m.recordToMonitoring(ctx, err)
}

func (m *ProductionErrorMonitor) collectErrorMetrics(err error) {
    labels := map[string]string{
        "error_type": getErrorType(err),
        "severity":   getErrorSeverity(err),
    }
    
    m.metrics.Counter("errors_total").WithLabels(labels).Inc()
    
    if businessErr, ok := err.(*CustomBusinessError); ok {
        m.metrics.Counter("business_errors_total").
            WithLabels(map[string]string{
                "error_code": fmt.Sprintf("%d", businessErr.Code),
            }).Inc()
    }
}

func (m *ProductionErrorMonitor) shouldAlert(err error) bool {
    // 系统错误需要立即告警
    if isSystemError(err) {
        return true
    }
    
    // 高频业务错误需要告警
    if businessErr, ok := err.(*CustomBusinessError); ok {
        errorRate := m.getErrorRate(businessErr.Code)
        return errorRate > m.alertManager.GetThreshold(businessErr.Code)
    }
    
    return false
}
```

### 2. 错误日志轮转和清理

```go
// LogRotationConfig 日志轮转配置
type LogRotationConfig struct {
    MaxFileSize   int64         // 最大文件大小（字节）
    MaxFiles      int           // 最大保留文件数
    RotateDaily   bool          // 是否按日轮转
    CompressOld   bool          // 是否压缩旧日志
    RetentionDays int           // 日志保留天数
}

// RotatingErrorLogger 轮转错误日志记录器
type RotatingErrorLogger struct {
    config     LogRotationConfig
    currentLog *os.File
    logDir     string
    mutex      sync.Mutex
}

func (r *RotatingErrorLogger) Log(entry ErrorLogEntry) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    // 检查是否需要轮转
    if r.needsRotation() {
        r.rotate()
    }
    
    // 写入日志
    logData, _ := json.Marshal(entry)
    r.currentLog.WriteString(string(logData) + "\n")
}

func (r *RotatingErrorLogger) needsRotation() bool {
    if r.currentLog == nil {
        return true
    }
    
    // 检查文件大小
    if stat, err := r.currentLog.Stat(); err == nil {
        if stat.Size() >= r.config.MaxFileSize {
            return true
        }
    }
    
    // 检查是否需要按日轮转
    if r.config.RotateDaily {
        return time.Now().Day() != getFileDay(r.currentLog)
    }
    
    return false
}
```

### 3. 错误恢复和自愈

```go
// SelfHealingSystem 自愈系统
type SelfHealingSystem struct {
    healthCheckers map[string]HealthChecker
    recoverActions map[string]RecoverAction
    monitoring     *HealthMonitoring
}

func (s *SelfHealingSystem) StartMonitoring() {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            s.performHealthChecks()
        }
    }()
}

func (s *SelfHealingSystem) performHealthChecks() {
    for name, checker := range s.healthCheckers {
        health := checker.Check()
        
        if !health.Healthy {
            log.Printf("Health check failed for %s: %s", name, health.Message)
            
            // 尝试自动恢复
            if action, exists := s.recoverActions[name]; exists {
                go func(name string, action RecoverAction) {
                    if err := action.Recover(); err != nil {
                        log.Printf("Auto-recovery failed for %s: %v", name, err)
                        s.escalateAlert(name, err)
                    } else {
                        log.Printf("Auto-recovery successful for %s", name)
                    }
                }(name, action)
            }
        }
    }
}

// 数据库连接恢复
type DatabaseRecoverAction struct {
    db *sql.DB
}

func (d *DatabaseRecoverAction) Recover() error {
    // 关闭现有连接
    d.db.Close()
    
    // 重新建立连接
    newDB, err := sql.Open("mysql", getDatabaseDSN())
    if err != nil {
        return err
    }
    
    // 验证连接
    if err := newDB.Ping(); err != nil {
        return err
    }
    
    d.db = newDB
    return nil
}
```

## 📊 性能最佳实践

### 1. 错误处理性能优化

```go
// 错误对象池，避免频繁分配
var businessErrorPool = sync.Pool{
    New: func() interface{} {
        return &CustomBusinessError{}
    },
}

// 高性能错误创建
func NewBusinessErrorFromPool(code int) *CustomBusinessError {
    err := businessErrorPool.Get().(*CustomBusinessError)
    err.Code = code
    err.Timestamp = time.Now()
    err.Details = nil
    return err
}

// 回收错误对象
func ReleaseBusinessError(err *CustomBusinessError) {
    err.Reset()
    businessErrorPool.Put(err)
}

// 批量错误处理
type BatchErrorProcessor struct {
    errors []error
    mutex  sync.Mutex
}

func (b *BatchErrorProcessor) Add(err error) {
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    b.errors = append(b.errors, err)
    
    // 达到批处理阈值时处理
    if len(b.errors) >= 100 {
        go b.processBatch()
    }
}

func (b *BatchErrorProcessor) processBatch() {
    b.mutex.Lock()
    batch := make([]error, len(b.errors))
    copy(batch, b.errors)
    b.errors = b.errors[:0]
    b.mutex.Unlock()
    
    // 批量处理错误
    for _, err := range batch {
        processErrorAsync(err)
    }
}
```

### 2. 缓存错误响应

```go
// ErrorResponseCache 错误响应缓存
type ErrorResponseCache struct {
    cache map[string][]byte
    mutex sync.RWMutex
    ttl   time.Duration
}

func (c *ErrorResponseCache) GetCachedResponse(errorKey string) ([]byte, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    response, exists := c.cache[errorKey]
    return response, exists
}

func (c *ErrorResponseCache) CacheResponse(errorKey string, response []byte) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.cache[errorKey] = response
    
    // 设置TTL清理
    time.AfterFunc(c.ttl, func() {
        c.mutex.Lock()
        delete(c.cache, errorKey)
        c.mutex.Unlock()
    })
}

// 生成错误缓存键
func generateErrorCacheKey(code int, lang string) string {
    return fmt.Sprintf("error:%d:%s", code, lang)
}
```

## 🔒 安全最佳实践

### 1. 错误信息安全

```go
// SecureErrorHandler 安全错误处理器
type SecureErrorHandler struct {
    production bool
}

func (h *SecureErrorHandler) Handle(ctx *Context, statusCode int, err error) error {
    response := map[string]any{
        "success":   false,
        "timestamp": time.Now().Unix(),
    }
    
    if h.production {
        // 生产环境 - 过滤敏感信息
        response["message"] = h.getSafeErrorMessage(statusCode)
        response["code"] = statusCode
    } else {
        // 开发环境 - 显示详细信息
        response["message"] = err.Error()
        response["details"] = h.getErrorDetails(err)
    }
    
    // 记录完整错误到安全日志
    h.logSecureError(ctx, err)
    
    ctx.JSON(statusCode, response)
    return nil
}

func (h *SecureErrorHandler) getSafeErrorMessage(statusCode int) string {
    safeMessages := map[int]string{
        400: "请求参数错误",
        401: "身份认证失败",
        403: "权限不足",
        404: "资源不存在",
        500: "服务器内部错误",
    }
    
    if msg, exists := safeMessages[statusCode]; exists {
        return msg
    }
    return "请求处理失败"
}

func (h *SecureErrorHandler) logSecureError(ctx *Context, err error) {
    // 过滤敏感信息后记录日志
    sanitizedContext := h.sanitizeContext(ctx)
    
    logEntry := map[string]any{
        "timestamp": time.Now(),
        "error":     err.Error(),
        "context":   sanitizedContext,
        "trace_id":  ctx.GetString("trace_id"),
    }
    
    securityLogger.Printf("ERROR: %+v", logEntry)
}

func (h *SecureErrorHandler) sanitizeContext(ctx *Context) map[string]any {
    context := make(map[string]any)
    
    // 只包含安全的上下文信息
    context["path"] = string(ctx.Path())
    context["method"] = string(ctx.Method())
    context["client_ip"] = maskIP(ctx.ClientIP())
    context["user_agent"] = string(ctx.UserAgent())
    
    // 不包含请求体和敏感头部
    return context
}
```

### 2. 错误日志安全

```go
// SecureLogger 安全日志记录器
type SecureLogger struct {
    sensitiveFields []string
    maskPatterns    map[string]*regexp.Regexp
}

func NewSecureLogger() *SecureLogger {
    return &SecureLogger{
        sensitiveFields: []string{
            "password", "token", "secret", "key", 
            "authorization", "cookie", "session",
        },
        maskPatterns: map[string]*regexp.Regexp{
            "email":       regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
            "phone":       regexp.MustCompile(`\d{3}-?\d{3,4}-?\d{4}`),
            "id_card":     regexp.MustCompile(`\d{17}[\dXx]`),
            "credit_card": regexp.MustCompile(`\d{4}-?\d{4}-?\d{4}-?\d{4}`),
        },
    }
}

func (s *SecureLogger) SanitizeLogData(data map[string]any) map[string]any {
    sanitized := make(map[string]any)
    
    for key, value := range data {
        if s.isSensitiveField(key) {
            sanitized[key] = "***MASKED***"
            continue
        }
        
        if strValue, ok := value.(string); ok {
            sanitized[key] = s.maskSensitivePatterns(strValue)
        } else {
            sanitized[key] = value
        }
    }
    
    return sanitized
}

func (s *SecureLogger) isSensitiveField(field string) bool {
    field = strings.ToLower(field)
    for _, sensitive := range s.sensitiveFields {
        if strings.Contains(field, sensitive) {
            return true
        }
    }
    return false
}

func (s *SecureLogger) maskSensitivePatterns(text string) string {
    for name, pattern := range s.maskPatterns {
        text = pattern.ReplaceAllStringFunc(text, func(match string) string {
            return fmt.Sprintf("***%s***", strings.ToUpper(name))
        })
    }
    return text
}
```

## 📈 监控和度量

### 1. 错误指标收集

```go
// ErrorMetrics 错误指标
type ErrorMetrics struct {
    TotalErrors          *prometheus.CounterVec
    ErrorsByType         *prometheus.CounterVec
    ErrorsByCode         *prometheus.CounterVec
    ErrorResponseTime    *prometheus.HistogramVec
    ActiveErrorHandlers  *prometheus.GaugeVec
}

func NewErrorMetrics() *ErrorMetrics {
    return &ErrorMetrics{
        TotalErrors: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_errors_total",
                Help: "Total number of HTTP errors",
            },
            []string{"status_code", "method", "path"},
        ),
        ErrorsByType: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "errors_by_type_total",
                Help: "Total errors by type",
            },
            []string{"error_type", "severity"},
        ),
        ErrorsByCode: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "business_errors_total",
                Help: "Total business errors by code",
            },
            []string{"error_code", "category"},
        ),
        ErrorResponseTime: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "error_response_time_seconds",
                Help: "Time spent processing errors",
            },
            []string{"handler_type"},
        ),
    }
}

// 记录错误指标
func (m *ErrorMetrics) RecordError(ctx *Context, statusCode int, err error, duration time.Duration) {
    // 记录HTTP错误
    m.TotalErrors.WithLabelValues(
        fmt.Sprintf("%d", statusCode),
        string(ctx.Method()),
        string(ctx.Path()),
    ).Inc()
    
    // 记录错误类型
    errorType, severity := classifyError(err)
    m.ErrorsByType.WithLabelValues(errorType, severity).Inc()
    
    // 记录业务错误
    if businessErr, ok := err.(*CustomBusinessError); ok {
        m.ErrorsByCode.WithLabelValues(
            fmt.Sprintf("%d", businessErr.Code),
            getErrorCategory(businessErr.Code),
        ).Inc()
    }
    
    // 记录响应时间
    handlerType := getHandlerType(err)
    m.ErrorResponseTime.WithLabelValues(handlerType).Observe(duration.Seconds())
}
```

### 2. 错误趋势分析

```go
// ErrorTrendAnalyzer 错误趋势分析器
type ErrorTrendAnalyzer struct {
    timeWindow   time.Duration
    dataPoints   []ErrorDataPoint
    alertManager AlertManager
    mutex        sync.RWMutex
}

type ErrorDataPoint struct {
    Timestamp  time.Time
    ErrorCount int
    ErrorRate  float64
    TopErrors  []TopError
}

type TopError struct {
    Code    int
    Count   int
    Message string
}

func (a *ErrorTrendAnalyzer) AnalyzeTrend() {
    a.mutex.RLock()
    defer a.mutex.RUnlock()
    
    if len(a.dataPoints) < 2 {
        return
    }
    
    recent := a.dataPoints[len(a.dataPoints)-1]
    previous := a.dataPoints[len(a.dataPoints)-2]
    
    // 计算错误率变化
    rateChange := recent.ErrorRate - previous.ErrorRate
    
    // 异常检测
    if rateChange > 0.1 { // 错误率增加超过10%
        alert := &Alert{
            Type:        AlertTypeErrorSpike,
            Severity:    SeverityHigh,
            Title:       "错误率异常升高",
            Description: fmt.Sprintf("错误率从 %.2f%% 增加到 %.2f%%", previous.ErrorRate*100, recent.ErrorRate*100),
            Timestamp:   time.Now(),
            Data: map[string]any{
                "previous_rate": previous.ErrorRate,
                "current_rate":  recent.ErrorRate,
                "change":        rateChange,
            },
        }
        
        a.alertManager.Send(alert)
    }
    
    // 新错误类型检测
    newErrors := a.detectNewErrors(recent, previous)
    if len(newErrors) > 0 {
        alert := &Alert{
            Type:        AlertTypeNewError,
            Severity:    SeverityMedium,
            Title:       "检测到新的错误类型",
            Description: fmt.Sprintf("发现 %d 种新的错误类型", len(newErrors)),
            Data: map[string]any{
                "new_errors": newErrors,
            },
        }
        
        a.alertManager.Send(alert)
    }
}
```

## 📚 相关文档

- **[快速开始](quick-start.md)** - 了解基础错误处理配置
- **[自动恢复机制](recovery.md)** - 实现错误自动恢复
- **[错误监控](monitoring.md)** - 监控错误指标和趋势
- **[故障排除](troubleshooting.md)** - 常见问题解决方案

---

> 💡 **总结**: 良好的错误处理不仅能提升系统的稳定性和用户体验，还能为运维团队提供有价值的诊断信息。建议在项目初期就建立完善的错误处理体系，并持续优化改进。