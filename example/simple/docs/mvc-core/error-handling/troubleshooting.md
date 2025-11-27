# 🔍 YYHertz 错误处理故障排除指南

本文档提供YYHertz框架错误处理系统的常见问题诊断和解决方案，帮助开发者快速定位和修复错误处理相关问题。

## 🚨 常见问题诊断

### 1. 错误处理器不生效

#### 问题现象
```
ERROR: 自定义错误处理器没有被调用
HTTP 500 错误显示默认错误页面而不是自定义处理
```

#### 排查步骤

**步骤1: 检查错误处理器注册**
```go
// ✅ 正确的注册方式
func init() {
    // 注册错误处理器
    errors.RegisterErrorHandler(404, &CustomNotFoundHandler{})
    errors.RegisterErrorHandler(500, &CustomServerErrorHandler{})
    
    // 注册业务错误处理器
    errors.RegisterBusinessErrorHandler(&BusinessErrorHandler{})
}

// ❌ 常见错误 - 忘记注册
func init() {
    handler := &CustomErrorHandler{}
    // 创建了处理器但没有注册
}
```

**步骤2: 验证处理器接口实现**
```go
// 检查是否正确实现了ErrorHandler接口
type CustomErrorHandler struct{}

// ✅ 必须实现的方法
func (h *CustomErrorHandler) Handle(ctx *Context, statusCode int, err error) error {
    // 实现处理逻辑
    return nil
}

func (h *CustomErrorHandler) CanHandle(statusCode int, err error) bool {
    // 实现判断逻辑
    return statusCode == 404
}

func (h *CustomErrorHandler) Priority() int {
    // 返回优先级
    return 10
}
```

**步骤3: 检查调用方式**
```go
// ✅ 正确的错误处理调用
func (c *UserController) GetUser() {
    user, err := c.userService.GetUser(id)
    if err != nil {
        // 使用错误处理系统
        errors.Handle(c.Ctx, 404, err)
        return
    }
}

// ❌ 错误做法 - 直接返回HTTP响应
func (c *UserController) GetUserBad() {
    user, err := c.userService.GetUser(id)
    if err != nil {
        // 绕过了错误处理系统
        c.JSON(404, map[string]any{"error": "user not found"})
        return
    }
}
```

#### 解决方案

**方案1: 启用错误处理调试**
```go
// 在main.go中启用调试模式
func main() {
    // 启用错误处理调试
    errors.SetDebugMode(true)
    
    app := mvc.HertzApp
    app.Run()
}

// 查看调试日志
// DEBUG: Registered error handler for status 404 with priority 10
// DEBUG: No handler found for status 500, using default handler
```

**方案2: 验证处理器链**
```go
// 添加处理器链调试信息
func DebugHandlerChain() {
    handlers := errors.GetRegisteredHandlers()
    for statusCode, handlerList := range handlers {
        fmt.Printf("Status %d handlers:\n", statusCode)
        for i, handler := range handlerList {
            fmt.Printf("  %d. %T (priority: %d)\n", i+1, handler, handler.Priority())
        }
    }
}
```

### 2. 错误页面显示异常

#### 问题现象
```
ERROR: 错误页面显示空白或模板渲染失败
Template execution error: template not found
```

#### 排查步骤

**步骤1: 检查模板文件路径**
```bash
# 确认错误模板文件存在
ls -la ./views/errors/
# 应该看到:
# 404.html
# 500.html
# error.html
```

**步骤2: 验证模板语法**
```html
<!-- ✅ 正确的错误模板结构 -->
{{define "errors/404.html"}}
<!DOCTYPE html>
<html>
<head>
    <title>页面未找到 - 404</title>
</head>
<body>
    <div class="error-container">
        <h1>404 - 页面未找到</h1>
        <p>{{.Message}}</p>
        {{if .Debug}}
            <pre>{{.Error}}</pre>
        {{end}}
    </div>
</body>
</html>
{{end}}

<!-- ❌ 常见错误 - 缺少define声明 -->
<!DOCTYPE html>
<html>
<!-- 缺少 {{define}} 包装 -->
```

**步骤3: 检查模板数据传递**
```go
// 检查错误处理器中的数据设置
func (h *CustomErrorHandler) Handle(ctx *Context, statusCode int, err error) error {
    data := map[string]any{
        "Title":     fmt.Sprintf("错误 %d", statusCode),
        "Message":   getErrorMessage(statusCode),
        "Error":     err.Error(),
        "Debug":     isDevelopmentMode(),
        "RequestID": ctx.GetString("request_id"),
    }
    
    // ✅ 确保传递了必要的数据
    ctx.HTML(statusCode, "errors/error.html", data)
    return nil
}
```

#### 解决方案

**方案1: 模板路径诊断**
```go
// 添加模板路径调试
func DiagnoseTemplatePath(templateName string) {
    app := mvc.HertzApp
    
    // 获取模板路径
    viewPath := app.GetViewPath()
    fullPath := filepath.Join(viewPath, templateName)
    
    fmt.Printf("Template path: %s\n", fullPath)
    
    // 检查文件是否存在
    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        fmt.Printf("❌ Template file not found: %s\n", fullPath)
    } else {
        fmt.Printf("✅ Template file exists: %s\n", fullPath)
    }
    
    // 检查模板内容
    content, err := os.ReadFile(fullPath)
    if err == nil {
        fmt.Printf("Template content length: %d bytes\n", len(content))
    }
}
```

**方案2: 错误模板热重载**
```go
// 开发环境下启用模板热重载
func EnableTemplateHotReload() {
    if isDevelopmentMode() {
        app := mvc.HertzApp
        app.EnableTemplateHotReload(true)
    }
}
```

### 3. 错误日志记录问题

#### 问题现象
```
ERROR: 错误日志没有记录或格式异常
日志文件权限错误
日志轮转失败
```

#### 排查步骤

**步骤1: 检查日志配置**
```go
// 验证日志配置
func DiagnoseLogConfiguration() {
    config := GetLogConfig()
    
    fmt.Printf("Log level: %s\n", config.Level)
    fmt.Printf("Log file: %s\n", config.File)
    fmt.Printf("Log format: %s\n", config.Format)
    
    // 检查日志文件权限
    if config.File != "" {
        if info, err := os.Stat(config.File); err == nil {
            fmt.Printf("Log file exists: %s\n", info.Mode())
        } else {
            fmt.Printf("❌ Log file error: %v\n", err)
        }
        
        // 检查目录权限
        dir := filepath.Dir(config.File)
        if info, err := os.Stat(dir); err == nil {
            fmt.Printf("Log directory permissions: %s\n", info.Mode())
        }
    }
}
```

**步骤2: 测试日志写入**
```go
// 测试日志写入功能
func TestLogWriting() {
    testError := errors.New("test error for log writing")
    
    // 测试不同级别的日志
    log.Debug("Debug message")
    log.Info("Info message")
    log.Warn("Warning message") 
    log.Error("Error message: %v", testError)
    
    fmt.Println("✅ Log writing test completed")
}
```

#### 解决方案

**方案1: 日志权限修复**
```bash
# 修复日志目录权限
sudo chmod 755 /var/log/yyhertz
sudo chown app:app /var/log/yyhertz

# 修复日志文件权限
sudo chmod 644 /var/log/yyhertz/error.log
sudo chown app:app /var/log/yyhertz/error.log
```

**方案2: 日志配置优化**
```go
// 生产环境日志配置
func SetupProductionLogging() {
    config := &LogConfig{
        Level:      "info",
        Format:     "json",
        File:       "/var/log/yyhertz/app.log",
        MaxSize:    100, // MB
        MaxFiles:   30,  // 保留30个文件
        MaxAge:     7,   // 保留7天
        Compress:   true,
        LocalTime:  true,
    }
    
    logger := NewRotatingLogger(config)
    SetGlobalLogger(logger)
}
```

### 4. 性能问题诊断

#### 问题现象
```
ERROR: 错误处理响应时间过长
内存使用异常增长
CPU使用率过高
```

#### 排查步骤

**步骤1: 错误处理性能分析**
```go
// 性能分析工具
type ErrorHandlingProfiler struct {
    metrics map[string]*PerformanceMetrics
    mutex   sync.RWMutex
}

type PerformanceMetrics struct {
    TotalCalls    int64
    TotalDuration time.Duration
    MaxDuration   time.Duration
    MinDuration   time.Duration
    AvgDuration   time.Duration
}

func (p *ErrorHandlingProfiler) MeasureHandler(handlerName string, handler ErrorHandler) ErrorHandler {
    return &ProfiledErrorHandler{
        name:     handlerName,
        handler:  handler,
        profiler: p,
    }
}

type ProfiledErrorHandler struct {
    name     string
    handler  ErrorHandler
    profiler *ErrorHandlingProfiler
}

func (p *ProfiledErrorHandler) Handle(ctx *Context, statusCode int, err error) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        p.profiler.RecordMetrics(p.name, duration)
    }()
    
    return p.handler.Handle(ctx, statusCode, err)
}

func (p *ErrorHandlingProfiler) GetReport() string {
    p.mutex.RLock()
    defer p.mutex.RUnlock()
    
    var report strings.Builder
    report.WriteString("Error Handler Performance Report:\n")
    
    for name, metrics := range p.metrics {
        report.WriteString(fmt.Sprintf(
            "Handler: %s\n"+
            "  Total Calls: %d\n"+
            "  Average Duration: %v\n"+
            "  Max Duration: %v\n"+
            "  Min Duration: %v\n\n",
            name, metrics.TotalCalls, metrics.AvgDuration,
            metrics.MaxDuration, metrics.MinDuration,
        ))
    }
    
    return report.String()
}
```

**步骤2: 内存泄露检测**
```go
// 内存使用监控
func MonitorErrorHandlingMemory() {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            fmt.Printf("Memory Stats:\n")
            fmt.Printf("  Alloc: %d KB\n", m.Alloc/1024)
            fmt.Printf("  TotalAlloc: %d KB\n", m.TotalAlloc/1024)
            fmt.Printf("  Sys: %d KB\n", m.Sys/1024)
            fmt.Printf("  NumGC: %d\n", m.NumGC)
            
            // 检查是否有内存泄露迹象
            if m.Alloc/1024 > 500000 { // 超过500MB
                log.Printf("⚠️  Memory usage high: %d KB", m.Alloc/1024)
            }
        }
    }()
}
```

#### 解决方案

**方案1: 错误处理优化**
```go
// 使用对象池减少分配
var errorResponsePool = sync.Pool{
    New: func() interface{} {
        return make(map[string]any)
    },
}

func OptimizedErrorResponse(statusCode int, message string) map[string]any {
    response := errorResponsePool.Get().(map[string]any)
    
    // 清空之前的数据
    for k := range response {
        delete(response, k)
    }
    
    // 设置新数据
    response["success"] = false
    response["code"] = statusCode
    response["message"] = message
    response["timestamp"] = time.Now().Unix()
    
    return response
}

func ReleaseErrorResponse(response map[string]any) {
    errorResponsePool.Put(response)
}
```

**方案2: 异步错误处理**
```go
// 异步错误日志记录
type AsyncErrorLogger struct {
    logChan chan ErrorLogEntry
    workers int
}

func NewAsyncErrorLogger(workers int) *AsyncErrorLogger {
    logger := &AsyncErrorLogger{
        logChan: make(chan ErrorLogEntry, 1000),
        workers: workers,
    }
    
    // 启动worker goroutines
    for i := 0; i < workers; i++ {
        go logger.worker()
    }
    
    return logger
}

func (a *AsyncErrorLogger) Log(entry ErrorLogEntry) {
    select {
    case a.logChan <- entry:
        // 日志已入队
    default:
        // 队列满，丢弃日志或使用降级策略
        log.Printf("Warning: Error log queue full, dropping log entry")
    }
}

func (a *AsyncErrorLogger) worker() {
    for entry := range a.logChan {
        // 实际写入日志
        a.writeLog(entry)
    }
}
```

## 🔧 调试工具和技巧

### 1. 错误处理调试器

```go
// ErrorHandlingDebugger 错误处理调试器
type ErrorHandlingDebugger struct {
    enabled     bool
    traceLevel  int
    output      io.Writer
}

func NewDebugger(enabled bool) *ErrorHandlingDebugger {
    return &ErrorHandlingDebugger{
        enabled:    enabled,
        traceLevel: 2,
        output:     os.Stdout,
    }
}

func (d *ErrorHandlingDebugger) TraceHandlerCall(handlerType string, statusCode int, err error) {
    if !d.enabled {
        return
    }
    
    fmt.Fprintf(d.output, "[DEBUG] %s Handler called for status %d\n", handlerType, statusCode)
    fmt.Fprintf(d.output, "        Error: %v\n", err)
    
    if d.traceLevel >= 2 {
        fmt.Fprintf(d.output, "        Stack trace:\n")
        debug.PrintStack()
    }
}

func (d *ErrorHandlingDebugger) TraceHandlerResult(handlerType string, success bool, duration time.Duration) {
    if !d.enabled {
        return
    }
    
    status := "✅ SUCCESS"
    if !success {
        status = "❌ FAILED"
    }
    
    fmt.Fprintf(d.output, "[DEBUG] %s Handler %s (took %v)\n", handlerType, status, duration)
}
```

### 2. 实时错误监控面板

```go
// ErrorDashboard 错误监控面板
type ErrorDashboard struct {
    server     *http.Server
    stats      *ErrorStats
    template   *template.Template
}

func NewErrorDashboard(port int, stats *ErrorStats) *ErrorDashboard {
    dashboard := &ErrorDashboard{
        stats: stats,
    }
    
    // 设置HTTP路由
    mux := http.NewServeMux()
    mux.HandleFunc("/", dashboard.handleDashboard)
    mux.HandleFunc("/api/stats", dashboard.handleStats)
    mux.HandleFunc("/api/errors", dashboard.handleErrors)
    
    dashboard.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    
    return dashboard
}

func (d *ErrorDashboard) Start() {
    log.Printf("Error dashboard starting on http://localhost%s", d.server.Addr)
    go d.server.ListenAndServe()
}

func (d *ErrorDashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
    data := struct {
        Stats   *ErrorStats
        Errors  []RecentError
        Uptime  time.Duration
    }{
        Stats:  d.stats,
        Errors: d.stats.GetRecentErrors(50),
        Uptime: time.Since(startTime),
    }
    
    tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Error Dashboard</title>
    <meta http-equiv="refresh" content="5">
    <style>
        .stats { display: flex; gap: 20px; margin-bottom: 20px; }
        .stat-box { 
            border: 1px solid #ddd; 
            padding: 15px; 
            border-radius: 5px; 
            flex: 1; 
        }
        .error-list { max-height: 400px; overflow-y: auto; }
        .error-item { 
            border-bottom: 1px solid #eee; 
            padding: 10px; 
        }
        .error-high { background-color: #ffebee; }
        .error-medium { background-color: #fff3e0; }
        .error-low { background-color: #f3e5f5; }
    </style>
</head>
<body>
    <h1>🚨 Error Monitoring Dashboard</h1>
    
    <div class="stats">
        <div class="stat-box">
            <h3>Total Errors</h3>
            <h2>{{.Stats.TotalErrors}}</h2>
        </div>
        <div class="stat-box">
            <h3>Error Rate</h3>
            <h2>{{printf "%.2f%%" .Stats.ErrorRate}}</h2>
        </div>
        <div class="stat-box">
            <h3>System Uptime</h3>
            <h2>{{.Uptime}}</h2>
        </div>
    </div>
    
    <div class="error-list">
        <h3>Recent Errors</h3>
        {{range .Errors}}
        <div class="error-item error-{{.Severity}}">
            <strong>{{.Timestamp.Format "15:04:05"}}</strong> 
            [{{.Code}}] {{.Message}}
            <br><small>Path: {{.Path}} | User: {{.UserID}}</small>
        </div>
        {{end}}
    </div>
    
    <script>
        // 自动刷新统计数据
        setInterval(() => {
            fetch('/api/stats')
                .then(response => response.json())
                .then(data => {
                    // 更新统计数据
                    console.log('Stats updated:', data);
                });
        }, 5000);
    </script>
</body>
</html>
    `
    
    t := template.Must(template.New("dashboard").Parse(tmpl))
    t.Execute(w, data)
}
```

### 3. 错误重现工具

```go
// ErrorReproducer 错误重现工具
type ErrorReproducer struct {
    scenarios map[string]ErrorScenario
}

type ErrorScenario struct {
    Name        string
    Description string
    Setup       func() error
    Trigger     func() error
    Cleanup     func() error
}

func NewErrorReproducer() *ErrorReproducer {
    reproducer := &ErrorReproducer{
        scenarios: make(map[string]ErrorScenario),
    }
    
    // 预定义常见错误场景
    reproducer.registerCommonScenarios()
    
    return reproducer
}

func (r *ErrorReproducer) registerCommonScenarios() {
    // 数据库连接错误
    r.scenarios["db_connection_error"] = ErrorScenario{
        Name:        "Database Connection Error",
        Description: "Simulate database connection failure",
        Setup: func() error {
            return nil
        },
        Trigger: func() error {
            // 临时断开数据库连接
            db := GetDatabase()
            return db.Close()
        },
        Cleanup: func() error {
            // 重新连接数据库
            return InitializeDatabase()
        },
    }
    
    // 内存不足错误
    r.scenarios["out_of_memory"] = ErrorScenario{
        Name:        "Out of Memory Error",
        Description: "Simulate memory exhaustion",
        Setup: func() error {
            return nil
        },
        Trigger: func() error {
            // 分配大量内存
            bigSlice := make([]byte, 1024*1024*1024) // 1GB
            _ = bigSlice
            return errors.New("out of memory simulation")
        },
        Cleanup: func() error {
            runtime.GC()
            return nil
        },
    }
}

func (r *ErrorReproducer) RunScenario(scenarioName string) error {
    scenario, exists := r.scenarios[scenarioName]
    if !exists {
        return fmt.Errorf("scenario not found: %s", scenarioName)
    }
    
    fmt.Printf("🧪 Running error scenario: %s\n", scenario.Name)
    fmt.Printf("   Description: %s\n", scenario.Description)
    
    // 设置环境
    if err := scenario.Setup(); err != nil {
        return fmt.Errorf("scenario setup failed: %v", err)
    }
    
    // 触发错误
    err := scenario.Trigger()
    
    // 清理环境
    if cleanupErr := scenario.Cleanup(); cleanupErr != nil {
        fmt.Printf("⚠️  Cleanup failed: %v\n", cleanupErr)
    }
    
    return err
}
```

## 🚀 生产环境故障处理

### 1. 快速故障定位

```go
// QuickDiagnostics 快速诊断工具
type QuickDiagnostics struct {
    startTime time.Time
}

func NewQuickDiagnostics() *QuickDiagnostics {
    return &QuickDiagnostics{
        startTime: time.Now(),
    }
}

func (q *QuickDiagnostics) RunFullDiagnostics() *DiagnosticReport {
    report := &DiagnosticReport{
        Timestamp: time.Now(),
        Duration:  time.Since(q.startTime),
    }
    
    // 1. 系统健康检查
    report.SystemHealth = q.checkSystemHealth()
    
    // 2. 错误处理器状态
    report.HandlerStatus = q.checkHandlerStatus()
    
    // 3. 数据库连接状态
    report.DatabaseStatus = q.checkDatabaseStatus()
    
    // 4. 内存和CPU使用情况
    report.ResourceUsage = q.checkResourceUsage()
    
    // 5. 最近错误统计
    report.RecentErrors = q.getRecentErrorStats()
    
    return report
}

type DiagnosticReport struct {
    Timestamp      time.Time
    Duration       time.Duration
    SystemHealth   HealthStatus
    HandlerStatus  HandlerStatus
    DatabaseStatus DatabaseStatus
    ResourceUsage  ResourceUsage
    RecentErrors   ErrorStatistics
}
```

### 2. 应急处理流程

```go
// EmergencyResponsePlan 应急响应计划
type EmergencyResponsePlan struct {
    escalationRules []EscalationRule
    actionPlans     map[string]ActionPlan
    notifier        NotificationService
}

type EscalationRule struct {
    Condition   func(*DiagnosticReport) bool
    Severity    AlertSeverity
    ActionPlan  string
    NotifyList  []string
}

func (e *EmergencyResponsePlan) HandleEmergency(report *DiagnosticReport) {
    // 检查所有升级规则
    for _, rule := range e.escalationRules {
        if rule.Condition(report) {
            fmt.Printf("🚨 Emergency condition detected: %s\n", rule.ActionPlan)
            
            // 执行应急处理计划
            if plan, exists := e.actionPlans[rule.ActionPlan]; exists {
                go e.executeActionPlan(plan, report)
            }
            
            // 发送通知
            alert := &EmergencyAlert{
                Severity:    rule.Severity,
                Description: fmt.Sprintf("Emergency condition: %s", rule.ActionPlan),
                Report:      report,
                Timestamp:   time.Now(),
            }
            
            e.notifier.SendEmergencyAlert(alert, rule.NotifyList)
        }
    }
}

// 预定义应急处理计划
func (e *EmergencyResponsePlan) setupDefaultPlans() {
    // 高错误率应急处理
    e.actionPlans["high_error_rate"] = ActionPlan{
        Name:        "High Error Rate Response",
        Description: "Handle high error rate emergency",
        Steps: []ActionStep{
            {
                Name:        "Enable Circuit Breaker",
                Action:      func() error { return enableEmergencyCircuitBreaker() },
                Timeout:     30 * time.Second,
                Critical:    true,
            },
            {
                Name:        "Scale Up Services",
                Action:      func() error { return scaleUpServices() },
                Timeout:     2 * time.Minute,
                Critical:    false,
            },
            {
                Name:        "Enable Fallback Mode",
                Action:      func() error { return enableFallbackMode() },
                Timeout:     10 * time.Second,
                Critical:    true,
            },
        },
    }
    
    // 系统资源不足应急处理
    e.actionPlans["resource_exhaustion"] = ActionPlan{
        Name:        "Resource Exhaustion Response",
        Description: "Handle system resource exhaustion",
        Steps: []ActionStep{
            {
                Name:        "Force Garbage Collection",
                Action:      func() error { runtime.GC(); return nil },
                Timeout:     10 * time.Second,
                Critical:    false,
            },
            {
                Name:        "Clear Caches",
                Action:      func() error { return clearAllCaches() },
                Timeout:     30 * time.Second,
                Critical:    true,
            },
            {
                Name:        "Restart Services",
                Action:      func() error { return restartNonCriticalServices() },
                Timeout:     5 * time.Minute,
                Critical:    false,
            },
        },
    }
}
```

## 📋 故障排除检查清单

### 开发环境检查清单

- [ ] 错误处理器是否正确注册
- [ ] 错误模板文件是否存在且语法正确
- [ ] 日志配置是否正确
- [ ] 数据库连接是否正常
- [ ] 依赖服务是否可用
- [ ] 配置文件是否正确加载

### 生产环境检查清单

- [ ] 系统资源使用情况（CPU、内存、磁盘）
- [ ] 网络连接状态
- [ ] 数据库连接池状态
- [ ] 缓存服务状态
- [ ] 日志文件权限和空间
- [ ] 监控系统是否正常工作
- [ ] 负载均衡器健康检查
- [ ] 安全证书是否有效

### 错误处理系统检查清单

- [ ] 错误处理器链是否完整
- [ ] 错误分类是否正确
- [ ] 错误码是否唯一且有意义
- [ ] 错误消息是否用户友好
- [ ] 敏感信息是否已脱敏
- [ ] 错误恢复机制是否工作
- [ ] 监控和告警是否配置
- [ ] 错误统计是否准确

## 📞 技术支持

### 获取帮助的步骤

1. **收集诊断信息**
   ```bash
   # 运行快速诊断
   ./yyhertz-cli diagnose --full
   
   # 收集错误日志
   tail -n 1000 /var/log/yyhertz/error.log > error_report.log
   
   # 导出配置信息
   ./yyhertz-cli config export > config_dump.json
   ```

2. **准备问题描述**
   - 问题发生的时间和频率
   - 错误的完整堆栈跟踪
   - 相关的请求URL和参数
   - 系统环境信息
   - 已尝试的解决方案

3. **联系技术支持**
   - GitHub Issues: `https://github.com/yyhertz/yyhertz/issues`
   - 技术论坛: `https://forum.yyhertz.com`
   - 邮件支持: `support@yyhertz.com`

## 📚 相关文档

- **[错误处理概览](overview.md)** - 了解错误处理系统架构
- **[快速开始](quick-start.md)** - 基础错误处理配置
- **[最佳实践](best-practices.md)** - 错误处理最佳实践
- **[错误监控](monitoring.md)** - 监控和告警配置

---

> 🔧 **提示**: 遇到问题时，首先启用调试模式并查看详细日志。大多数问题都可以通过仔细分析错误信息和检查配置来解决。如果问题持续存在，请收集完整的诊断信息并联系技术支持。