# 性能监控

YYHertz 框架提供了全面的性能监控功能，帮助开发者实时了解应用程序的性能状况，快速定位性能瓶颈，优化系统表现。

## 概述

性能监控是现代 Web 应用程序的重要组成部分。YYHertz 的性能监控系统提供：

- 实时性能指标收集
- 请求链路追踪
- 数据库查询监控
- 内存和 CPU 监控
- 自定义性能指标
- 性能告警系统
- 可视化监控面板
- 性能分析报告

## 基本使用

### 启用性能监控

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/monitoring"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

func main() {
    app := mvc.HertzApp
    
    // 创建性能监控器
    monitor := monitoring.New(monitoring.Config{
        Enabled:     true,
        MetricsPath: "/metrics",
        MetricsPort: 9090,
        
        // Prometheus 配置
        PrometheusEnabled: true,
        
        // 链路追踪配置
        TracingEnabled: true,
        JaegerEndpoint: "http://localhost:14268/api/traces",
        
        // 性能分析配置
        ProfilingEnabled: true,
        PprofPath:       "/debug/pprof",
    })
    
    // 启动监控器
    monitor.Start()
    
    // 添加性能监控中间件
    app.Use(middleware.PerformanceMiddleware(middleware.PerformanceConfig{
        Monitor:          monitor,
        EnableTracing:    true,
        EnableProfiling:  true,
        SlowRequestTime:  1 * time.Second,
    }))
    
    // 添加指标收集中间件
    app.Use(middleware.MetricsMiddleware(monitor))
    
    defer monitor.Stop()
    
    app.Run()
}
```

### 基本指标收集

```go
// 在控制器中收集性能指标
type UserController struct {
    mvc.Controller
    monitor monitoring.Monitor
}

func (c *UserController) GetUsers() {
    // 开始追踪
    span := c.monitor.StartSpan("user.get_users")
    defer span.Finish()
    
    // 计时器
    timer := c.monitor.Timer("database.query_time")
    defer timer.Stop()
    
    // 计数器
    c.monitor.Counter("user.requests").Inc()
    
    // 查询用户数据
    users, err := c.userService.GetUsers()
    if err != nil {
        c.monitor.Counter("user.errors").Inc()
        span.SetError(err)
        c.JSON(500, gin.H{"error": "Failed to get users"})
        return
    }
    
    // 记录查询结果数量
    c.monitor.Gauge("user.count").Set(float64(len(users)))
    
    // 直方图 - 记录响应大小
    responseSize := c.calculateResponseSize(users)
    c.monitor.Histogram("response.size").Observe(float64(responseSize))
    
    c.JSON(200, users)
}

func (c *UserController) CreateUser() {
    span := c.monitor.StartSpan("user.create")
    defer span.Finish()
    
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.monitor.Counter("user.validation_errors").Inc()
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // 添加标签到追踪
    span.SetTag("user.email", req.Email)
    span.SetTag("user.role", req.Role)
    
    // 创建用户
    user, err := c.userService.CreateUser(req)
    if err != nil {
        c.monitor.Counter("user.creation_errors").Inc()
        span.SetError(err)
        c.JSON(500, gin.H{"error": "Failed to create user"})
        return
    }
    
    c.monitor.Counter("user.created").Inc()
    c.JSON(201, user)
}
```

## 指标类型

### 基本指标

```go
type Metrics interface {
    // 计数器 - 只增不减的累计值
    Counter(name string) Counter
    
    // 测量仪 - 可增可减的瞬时值
    Gauge(name string) Gauge
    
    // 直方图 - 观察值的分布
    Histogram(name string) Histogram
    
    // 摘要 - 分位数统计
    Summary(name string) Summary
    
    // 计时器 - 便捷的持续时间测量
    Timer(name string) Timer
}

// 使用示例
func collectMetrics(monitor monitoring.Monitor) {
    // 计数器 - 记录事件次数
    requestCounter := monitor.Counter("http_requests_total")
    requestCounter.Inc()
    requestCounter.Add(5)
    
    // 测量仪 - 记录当前值
    activeConnections := monitor.Gauge("active_connections")
    activeConnections.Set(100)
    activeConnections.Inc()
    activeConnections.Dec()
    activeConnections.Add(10)
    activeConnections.Sub(5)
    
    // 直方图 - 记录数值分布
    requestDuration := monitor.Histogram("http_request_duration_seconds")
    requestDuration.Observe(0.25)
    
    // 摘要 - 记录分位数
    responseSize := monitor.Summary("http_response_size_bytes")
    responseSize.Observe(1024)
    
    // 计时器 - 测量执行时间
    timer := monitor.Timer("operation_duration")
    defer timer.Stop()
    // 或者
    start := time.Now()
    // ... 执行操作
    timer.Record(time.Since(start))
}
```

### 标签和维度

```go
// 带标签的指标
func collectLabeledMetrics(monitor monitoring.Monitor) {
    // 使用标签区分不同维度
    httpRequests := monitor.Counter("http_requests_total").WithLabels(map[string]string{
        "method": "GET",
        "path":   "/api/users",
        "status": "200",
    })
    httpRequests.Inc()
    
    // 数据库查询指标
    dbQueries := monitor.Histogram("database_query_duration").WithLabels(map[string]string{
        "database": "mysql",
        "table":    "users",
        "operation": "select",
    })
    dbQueries.Observe(0.1)
    
    // 缓存指标
    cacheHits := monitor.Counter("cache_operations_total").WithLabels(map[string]string{
        "cache_type": "redis",
        "operation":  "hit",
    })
    cacheHits.Inc()
}
```

## 中间件监控

### 性能监控中间件

```go
package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/zsy619/yyhertz/framework/mvc/monitoring"
)

type PerformanceConfig struct {
    Monitor         monitoring.Monitor
    EnableTracing   bool
    EnableProfiling bool
    SlowRequestTime time.Duration
}

func PerformanceMiddleware(config PerformanceConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 创建追踪 span
        var span monitoring.Span
        if config.EnableTracing {
            span = config.Monitor.StartSpan("http.request")
            span.SetTag("http.method", c.Request.Method)
            span.SetTag("http.url", c.Request.URL.String())
            span.SetTag("http.user_agent", c.Request.UserAgent())
            defer span.Finish()
        }
        
        // 记录请求开始
        config.Monitor.Counter("http_requests_total").WithLabels(map[string]string{
            "method": c.Request.Method,
            "path":   c.FullPath(),
        }).Inc()
        
        // 记录并发连接数
        config.Monitor.Gauge("http_active_requests").Inc()
        defer config.Monitor.Gauge("http_active_requests").Dec()
        
        // 执行请求
        c.Next()
        
        // 记录响应时间
        duration := time.Since(start)
        
        // 记录响应状态
        status := c.Writer.Status()
        config.Monitor.Counter("http_responses_total").WithLabels(map[string]string{
            "method": c.Request.Method,
            "path":   c.FullPath(),
            "status": strconv.Itoa(status),
        }).Inc()
        
        // 记录响应时间分布
        config.Monitor.Histogram("http_request_duration_seconds").WithLabels(map[string]string{
            "method": c.Request.Method,
            "path":   c.FullPath(),
        }).Observe(duration.Seconds())
        
        // 记录响应大小
        responseSize := c.Writer.Size()
        config.Monitor.Histogram("http_response_size_bytes").WithLabels(map[string]string{
            "method": c.Request.Method,
            "path":   c.FullPath(),
        }).Observe(float64(responseSize))
        
        // 追踪信息
        if span != nil {
            span.SetTag("http.status_code", status)
            span.SetTag("http.response_size", responseSize)
            
            if status >= 400 {
                span.SetError(fmt.Errorf("HTTP %d", status))
            }
        }
        
        // 慢请求告警
        if duration > config.SlowRequestTime {
            config.Monitor.Counter("http_slow_requests_total").WithLabels(map[string]string{
                "method": c.Request.Method,
                "path":   c.FullPath(),
            }).Inc()
            
            // 记录慢请求日志
            log.Printf("Slow request: %s %s took %v", c.Request.Method, c.Request.URL.String(), duration)
        }
    }
}
```

### 数据库监控中间件

```go
type DatabaseMonitor struct {
    monitor monitoring.Monitor
}

func NewDatabaseMonitor(monitor monitoring.Monitor) *DatabaseMonitor {
    return &DatabaseMonitor{monitor: monitor}
}

func (dm *DatabaseMonitor) WrapDB(db *gorm.DB) *gorm.DB {
    return db.Callback().Query().Before("gorm:query").Register("monitoring:before_query", dm.beforeQuery).
        Callback().Query().After("gorm:query").Register("monitoring:after_query", dm.afterQuery).
        Callback().Create().Before("gorm:create").Register("monitoring:before_create", dm.beforeCreate).
        Callback().Create().After("gorm:create").Register("monitoring:after_create", dm.afterCreate).
        Callback().Update().Before("gorm:update").Register("monitoring:before_update", dm.beforeUpdate).
        Callback().Update().After("gorm:update").Register("monitoring:after_update", dm.afterUpdate).
        Callback().Delete().Before("gorm:delete").Register("monitoring:before_delete", dm.beforeDelete).
        Callback().Delete().After("gorm:delete").Register("monitoring:after_delete", dm.afterDelete)
}

func (dm *DatabaseMonitor) beforeQuery(db *gorm.DB) {
    db.Set("monitoring:start_time", time.Now())
    
    dm.monitor.Counter("database_queries_total").WithLabels(map[string]string{
        "operation": "select",
        "table":     db.Statement.Table,
    }).Inc()
}

func (dm *DatabaseMonitor) afterQuery(db *gorm.DB) {
    startTime, exists := db.Get("monitoring:start_time")
    if !exists {
        return
    }
    
    duration := time.Since(startTime.(time.Time))
    
    dm.monitor.Histogram("database_query_duration_seconds").WithLabels(map[string]string{
        "operation": "select",
        "table":     db.Statement.Table,
    }).Observe(duration.Seconds())
    
    if db.Error != nil {
        dm.monitor.Counter("database_errors_total").WithLabels(map[string]string{
            "operation": "select",
            "table":     db.Statement.Table,
        }).Inc()
    }
}

// 类似的方法用于 Create, Update, Delete 操作
func (dm *DatabaseMonitor) beforeCreate(db *gorm.DB) {
    db.Set("monitoring:start_time", time.Now())
    dm.monitor.Counter("database_queries_total").WithLabels(map[string]string{
        "operation": "insert",
        "table":     db.Statement.Table,
    }).Inc()
}

func (dm *DatabaseMonitor) afterCreate(db *gorm.DB) {
    startTime, exists := db.Get("monitoring:start_time")
    if !exists {
        return
    }
    
    duration := time.Since(startTime.(time.Time))
    
    dm.monitor.Histogram("database_query_duration_seconds").WithLabels(map[string]string{
        "operation": "insert",
        "table":     db.Statement.Table,
    }).Observe(duration.Seconds())
    
    if db.Error != nil {
        dm.monitor.Counter("database_errors_total").WithLabels(map[string]string{
            "operation": "insert",
            "table":     db.Statement.Table,
        }).Inc()
    } else {
        dm.monitor.Gauge("database_last_insert_id").WithLabels(map[string]string{
            "table": db.Statement.Table,
        }).Set(float64(db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, reflect.ValueOf(db.Statement.Dest)).Int()))
    }
}
```

## 链路追踪

### 分布式追踪

```go
package monitoring

import (
    "context"
    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
)

type TracingConfig struct {
    ServiceName     string  `yaml:"service_name"`
    JaegerEndpoint  string  `yaml:"jaeger_endpoint"`
    SamplingRate    float64 `yaml:"sampling_rate"`
    ReporterLogSpans bool   `yaml:"reporter_log_spans"`
}

func InitTracing(config TracingConfig) (opentracing.Tracer, io.Closer, error) {
    cfg := jaegerConfig.Configuration{
        ServiceName: config.ServiceName,
        Sampler: &jaegerConfig.SamplerConfig{
            Type:  jaeger.SamplerTypeConst,
            Param: config.SamplingRate,
        },
        Reporter: &jaegerConfig.ReporterConfig{
            LogSpans:           config.ReporterLogSpans,
            CollectorEndpoint:  config.JaegerEndpoint,
        },
    }
    
    tracer, closer, err := cfg.NewTracer()
    if err != nil {
        return nil, nil, err
    }
    
    opentracing.SetGlobalTracer(tracer)
    return tracer, closer, nil
}

// 创建追踪 span
func StartSpan(operationName string, parent ...opentracing.SpanContext) opentracing.Span {
    var span opentracing.Span
    
    if len(parent) > 0 && parent[0] != nil {
        span = opentracing.StartSpan(operationName, opentracing.ChildOf(parent[0]))
    } else {
        span = opentracing.StartSpan(operationName)
    }
    
    return span
}

// 从上下文获取 span
func SpanFromContext(ctx context.Context) opentracing.Span {
    return opentracing.SpanFromContext(ctx)
}

// 将 span 添加到上下文
func ContextWithSpan(ctx context.Context, span opentracing.Span) context.Context {
    return opentracing.ContextWithSpan(ctx, span)
}
```

### 服务间追踪

```go
// HTTP 客户端追踪
func (c *HttpClient) DoWithTracing(req *http.Request, parentSpan opentracing.Span) (*http.Response, error) {
    span := opentracing.StartSpan("http.client.request", opentracing.ChildOf(parentSpan.Context()))
    defer span.Finish()
    
    // 添加追踪头
    span.SetTag("http.method", req.Method)
    span.SetTag("http.url", req.URL.String())
    
    // 注入追踪上下文到请求头
    opentracing.GlobalTracer().Inject(
        span.Context(),
        opentracing.HTTPHeaders,
        opentracing.HTTPHeadersCarrier(req.Header),
    )
    
    resp, err := c.client.Do(req)
    if err != nil {
        span.SetTag("error", true)
        span.LogFields(
            log.String("event", "error"),
            log.String("message", err.Error()),
        )
        return nil, err
    }
    
    span.SetTag("http.status_code", resp.StatusCode)
    return resp, nil
}

// 数据库操作追踪
func (r *UserRepository) GetByIDWithTracing(ctx context.Context, id int) (*User, error) {
    span, ctx := opentracing.StartSpanFromContext(ctx, "db.user.get_by_id")
    defer span.Finish()
    
    span.SetTag("db.table", "users")
    span.SetTag("db.operation", "select")
    span.SetTag("user.id", id)
    
    var user User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        span.SetTag("error", true)
        span.LogFields(
            log.String("event", "db_error"),
            log.String("message", err.Error()),
        )
        return nil, err
    }
    
    span.SetTag("db.rows_affected", 1)
    return &user, nil
}
```

## 系统监控

### 系统资源监控

```go
package monitoring

import (
    "runtime"
    "time"
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/shirou/gopsutil/v3/disk"
    "github.com/shirou/gopsutil/v3/net"
)

type SystemMonitor struct {
    monitor  monitoring.Monitor
    interval time.Duration
    done     chan struct{}
}

func NewSystemMonitor(monitor monitoring.Monitor, interval time.Duration) *SystemMonitor {
    return &SystemMonitor{
        monitor:  monitor,
        interval: interval,
        done:     make(chan struct{}),
    }
}

func (sm *SystemMonitor) Start() {
    go sm.collectMetrics()
}

func (sm *SystemMonitor) Stop() {
    close(sm.done)
}

func (sm *SystemMonitor) collectMetrics() {
    ticker := time.NewTicker(sm.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            sm.collectCPUMetrics()
            sm.collectMemoryMetrics()
            sm.collectDiskMetrics()
            sm.collectNetworkMetrics()
            sm.collectGoRuntimeMetrics()
        case <-sm.done:
            return
        }
    }
}

func (sm *SystemMonitor) collectCPUMetrics() {
    cpuPercent, err := cpu.Percent(0, false)
    if err != nil {
        return
    }
    
    if len(cpuPercent) > 0 {
        sm.monitor.Gauge("system_cpu_usage_percent").Set(cpuPercent[0])
    }
    
    // CPU 核心数
    cpuCount, _ := cpu.Counts(true)
    sm.monitor.Gauge("system_cpu_cores").Set(float64(cpuCount))
}

func (sm *SystemMonitor) collectMemoryMetrics() {
    memInfo, err := mem.VirtualMemory()
    if err != nil {
        return
    }
    
    sm.monitor.Gauge("system_memory_total_bytes").Set(float64(memInfo.Total))
    sm.monitor.Gauge("system_memory_used_bytes").Set(float64(memInfo.Used))
    sm.monitor.Gauge("system_memory_available_bytes").Set(float64(memInfo.Available))
    sm.monitor.Gauge("system_memory_usage_percent").Set(memInfo.UsedPercent)
    
    // 交换分区
    swapInfo, err := mem.SwapMemory()
    if err == nil {
        sm.monitor.Gauge("system_swap_total_bytes").Set(float64(swapInfo.Total))
        sm.monitor.Gauge("system_swap_used_bytes").Set(float64(swapInfo.Used))
        sm.monitor.Gauge("system_swap_usage_percent").Set(swapInfo.UsedPercent)
    }
}

func (sm *SystemMonitor) collectDiskMetrics() {
    diskUsage, err := disk.Usage("/")
    if err != nil {
        return
    }
    
    sm.monitor.Gauge("system_disk_total_bytes").Set(float64(diskUsage.Total))
    sm.monitor.Gauge("system_disk_used_bytes").Set(float64(diskUsage.Used))
    sm.monitor.Gauge("system_disk_free_bytes").Set(float64(diskUsage.Free))
    sm.monitor.Gauge("system_disk_usage_percent").Set(diskUsage.UsedPercent)
}

func (sm *SystemMonitor) collectNetworkMetrics() {
    netStats, err := net.IOCounters(false)
    if err != nil || len(netStats) == 0 {
        return
    }
    
    stat := netStats[0]
    sm.monitor.Counter("system_network_bytes_sent_total").Set(float64(stat.BytesSent))
    sm.monitor.Counter("system_network_bytes_recv_total").Set(float64(stat.BytesRecv))
    sm.monitor.Counter("system_network_packets_sent_total").Set(float64(stat.PacketsSent))
    sm.monitor.Counter("system_network_packets_recv_total").Set(float64(stat.PacketsRecv))
}

func (sm *SystemMonitor) collectGoRuntimeMetrics() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // Go 运行时内存指标
    sm.monitor.Gauge("go_memory_alloc_bytes").Set(float64(m.Alloc))
    sm.monitor.Gauge("go_memory_total_alloc_bytes").Set(float64(m.TotalAlloc))
    sm.monitor.Gauge("go_memory_sys_bytes").Set(float64(m.Sys))
    sm.monitor.Gauge("go_memory_heap_alloc_bytes").Set(float64(m.HeapAlloc))
    sm.monitor.Gauge("go_memory_heap_sys_bytes").Set(float64(m.HeapSys))
    sm.monitor.Gauge("go_memory_heap_idle_bytes").Set(float64(m.HeapIdle))
    sm.monitor.Gauge("go_memory_heap_inuse_bytes").Set(float64(m.HeapInuse))
    sm.monitor.Gauge("go_memory_stack_inuse_bytes").Set(float64(m.StackInuse))
    sm.monitor.Gauge("go_memory_stack_sys_bytes").Set(float64(m.StackSys))
    
    // GC 指标
    sm.monitor.Counter("go_gc_runs_total").Set(float64(m.NumGC))
    sm.monitor.Gauge("go_gc_pause_total_ns").Set(float64(m.PauseTotalNs))
    
    // Goroutine 数量
    sm.monitor.Gauge("go_goroutines").Set(float64(runtime.NumGoroutine()))
    
    // CPU 核心数
    sm.monitor.Gauge("go_max_procs").Set(float64(runtime.GOMAXPROCS(0)))
}
```

## 性能分析

### CPU 性能分析

```go
package monitoring

import (
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "runtime/pprof"
    "time"
)

type Profiler struct {
    enabled bool
    server  *http.Server
}

func NewProfiler(enabled bool, port string) *Profiler {
    if !enabled {
        return &Profiler{enabled: false}
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/debug/pprof/", pprof.Index)
    mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
    mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
    
    server := &http.Server{
        Addr:    ":" + port,
        Handler: mux,
    }
    
    return &Profiler{
        enabled: true,
        server:  server,
    }
}

func (p *Profiler) Start() error {
    if !p.enabled {
        return nil
    }
    
    go func() {
        if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("Profiler server error: %v", err)
        }
    }()
    
    return nil
}

func (p *Profiler) Stop() error {
    if !p.enabled || p.server == nil {
        return nil
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return p.server.Shutdown(ctx)
}

// 手动触发 CPU 性能分析
func ProfileCPU(duration time.Duration, filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    
    if err := pprof.StartCPUProfile(f); err != nil {
        return err
    }
    
    time.Sleep(duration)
    pprof.StopCPUProfile()
    
    return nil
}

// 手动触发内存性能分析
func ProfileMemory(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    
    runtime.GC() // 强制垃圾回收
    
    if err := pprof.WriteHeapProfile(f); err != nil {
        return err
    }
    
    return nil
}
```

### 自动性能分析

```go
type AutoProfiler struct {
    enabled       bool
    cpuThreshold  float64
    memThreshold  float64
    monitor       monitoring.Monitor
    profiling     bool
    mutex         sync.Mutex
}

func NewAutoProfiler(enabled bool, cpuThreshold, memThreshold float64, monitor monitoring.Monitor) *AutoProfiler {
    return &AutoProfiler{
        enabled:      enabled,
        cpuThreshold: cpuThreshold,
        memThreshold: memThreshold,
        monitor:      monitor,
    }
}

func (ap *AutoProfiler) Start() {
    if !ap.enabled {
        return
    }
    
    go ap.monitorLoop()
}

func (ap *AutoProfiler) monitorLoop() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        ap.checkAndProfile()
    }
}

func (ap *AutoProfiler) checkAndProfile() {
    ap.mutex.Lock()
    defer ap.mutex.Unlock()
    
    if ap.profiling {
        return
    }
    
    // 检查 CPU 使用率
    cpuUsage := ap.getCurrentCPUUsage()
    if cpuUsage > ap.cpuThreshold {
        ap.startCPUProfile()
        return
    }
    
    // 检查内存使用率
    memUsage := ap.getCurrentMemoryUsage()
    if memUsage > ap.memThreshold {
        ap.startMemoryProfile()
        return
    }
}

func (ap *AutoProfiler) startCPUProfile() {
    ap.profiling = true
    
    filename := fmt.Sprintf("cpu_profile_%d.prof", time.Now().Unix())
    
    go func() {
        defer func() {
            ap.mutex.Lock()
            ap.profiling = false
            ap.mutex.Unlock()
        }()
        
        if err := ProfileCPU(30*time.Second, filename); err != nil {
            log.Printf("Failed to profile CPU: %v", err)
        } else {
            log.Printf("CPU profile saved to %s", filename)
            ap.monitor.Counter("profiler_cpu_profiles_created").Inc()
        }
    }()
}

func (ap *AutoProfiler) startMemoryProfile() {
    ap.profiling = true
    
    filename := fmt.Sprintf("mem_profile_%d.prof", time.Now().Unix())
    
    go func() {
        defer func() {
            ap.mutex.Lock()
            ap.profiling = false
            ap.mutex.Unlock()
        }()
        
        if err := ProfileMemory(filename); err != nil {
            log.Printf("Failed to profile memory: %v", err)
        } else {
            log.Printf("Memory profile saved to %s", filename)
            ap.monitor.Counter("profiler_memory_profiles_created").Inc()
        }
    }()
}
```

## 告警系统

### 告警规则

```go
type AlertRule struct {
    Name        string            `yaml:"name"`
    Metric      string            `yaml:"metric"`
    Condition   string            `yaml:"condition"` // >, <, >=, <=, ==, !=
    Threshold   float64           `yaml:"threshold"`
    Duration    time.Duration     `yaml:"duration"`
    Labels      map[string]string `yaml:"labels"`
    Annotations map[string]string `yaml:"annotations"`
    Webhook     string            `yaml:"webhook"`
    Email       []string          `yaml:"email"`
}

type AlertManager struct {
    rules    []AlertRule
    states   map[string]*AlertState
    monitor  monitoring.Monitor
    mutex    sync.RWMutex
}

type AlertState struct {
    Rule      AlertRule
    Active    bool
    StartTime time.Time
    LastSent  time.Time
    Value     float64
}

func NewAlertManager(rules []AlertRule, monitor monitoring.Monitor) *AlertManager {
    return &AlertManager{
        rules:   rules,
        states:  make(map[string]*AlertState),
        monitor: monitor,
    }
}

func (am *AlertManager) Start() {
    go am.evaluateLoop()
}

func (am *AlertManager) evaluateLoop() {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        am.evaluateRules()
    }
}

func (am *AlertManager) evaluateRules() {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    
    for _, rule := range am.rules {
        am.evaluateRule(rule)
    }
}

func (am *AlertManager) evaluateRule(rule AlertRule) {
    // 获取指标当前值
    value := am.monitor.GetMetricValue(rule.Metric, rule.Labels)
    
    // 检查条件
    triggered := am.checkCondition(rule.Condition, value, rule.Threshold)
    
    state, exists := am.states[rule.Name]
    if !exists {
        state = &AlertState{
            Rule: rule,
        }
        am.states[rule.Name] = state
    }
    
    state.Value = value
    
    if triggered {
        if !state.Active {
            state.Active = true
            state.StartTime = time.Now()
        }
        
        // 检查是否达到持续时间
        if time.Since(state.StartTime) >= rule.Duration {
            am.sendAlert(state)
        }
    } else {
        if state.Active {
            state.Active = false
            am.sendResolution(state)
        }
    }
}

func (am *AlertManager) checkCondition(condition string, value, threshold float64) bool {
    switch condition {
    case ">":
        return value > threshold
    case "<":
        return value < threshold
    case ">=":
        return value >= threshold
    case "<=":
        return value <= threshold
    case "==":
        return value == threshold
    case "!=":
        return value != threshold
    default:
        return false
    }
}

func (am *AlertManager) sendAlert(state *AlertState) {
    // 避免重复发送
    if time.Since(state.LastSent) < 5*time.Minute {
        return
    }
    
    state.LastSent = time.Now()
    
    alert := Alert{
        Name:        state.Rule.Name,
        Value:       state.Value,
        Threshold:   state.Rule.Threshold,
        StartTime:   state.StartTime,
        Labels:      state.Rule.Labels,
        Annotations: state.Rule.Annotations,
        Status:      "firing",
    }
    
    // 发送 Webhook 告警
    if state.Rule.Webhook != "" {
        go am.sendWebhookAlert(state.Rule.Webhook, alert)
    }
    
    // 发送邮件告警
    if len(state.Rule.Email) > 0 {
        go am.sendEmailAlert(state.Rule.Email, alert)
    }
    
    am.monitor.Counter("alerts_sent_total").WithLabels(map[string]string{
        "alert": state.Rule.Name,
    }).Inc()
}

type Alert struct {
    Name        string            `json:"name"`
    Value       float64           `json:"value"`
    Threshold   float64           `json:"threshold"`
    StartTime   time.Time         `json:"start_time"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    Status      string            `json:"status"` // firing, resolved
}

func (am *AlertManager) sendWebhookAlert(webhook string, alert Alert) {
    jsonData, _ := json.Marshal(alert)
    
    resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Failed to send webhook alert: %v", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        log.Printf("Webhook alert failed with status: %d", resp.StatusCode)
    }
}
```

## 监控面板

### REST API

```go
type MonitoringController struct {
    mvc.Controller
    monitor monitoring.Monitor
}

// 获取实时指标
func (c *MonitoringController) GetMetrics() {
    metrics := c.monitor.GetAllMetrics()
    c.JSON(200, metrics)
}

// 获取系统健康状态
func (c *MonitoringController) GetHealth() {
    health := HealthCheck{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]CheckResult),
    }
    
    // 检查数据库连接
    if err := c.checkDatabase(); err != nil {
        health.Status = "unhealthy"
        health.Checks["database"] = CheckResult{
            Status: "fail",
            Error:  err.Error(),
        }
    } else {
        health.Checks["database"] = CheckResult{Status: "pass"}
    }
    
    // 检查 Redis 连接
    if err := c.checkRedis(); err != nil {
        health.Status = "degraded"
        health.Checks["redis"] = CheckResult{
            Status: "fail",
            Error:  err.Error(),
        }
    } else {
        health.Checks["redis"] = CheckResult{Status: "pass"}
    }
    
    // 检查磁盘空间
    if err := c.checkDiskSpace(); err != nil {
        health.Status = "degraded"
        health.Checks["disk"] = CheckResult{
            Status: "fail",
            Error:  err.Error(),
        }
    } else {
        health.Checks["disk"] = CheckResult{Status: "pass"}
    }
    
    c.JSON(200, health)
}

type HealthCheck struct {
    Status    string                   `json:"status"`
    Timestamp time.Time                `json:"timestamp"`
    Checks    map[string]CheckResult   `json:"checks"`
}

type CheckResult struct {
    Status string `json:"status"`
    Error  string `json:"error,omitempty"`
}

// 获取性能统计
func (c *MonitoringController) GetStats() {
    stats := PerformanceStats{
        RequestsPerSecond: c.monitor.GetMetricValue("http_requests_per_second", nil),
        AverageResponse:   c.monitor.GetMetricValue("http_average_response_time", nil),
        ErrorRate:         c.monitor.GetMetricValue("http_error_rate", nil),
        ActiveConnections: c.monitor.GetMetricValue("http_active_connections", nil),
        CPUUsage:         c.monitor.GetMetricValue("system_cpu_usage_percent", nil),
        MemoryUsage:      c.monitor.GetMetricValue("system_memory_usage_percent", nil),
        GoroutineCount:   c.monitor.GetMetricValue("go_goroutines", nil),
    }
    
    c.JSON(200, stats)
}

type PerformanceStats struct {
    RequestsPerSecond float64 `json:"requests_per_second"`
    AverageResponse   float64 `json:"average_response_time"`
    ErrorRate         float64 `json:"error_rate"`
    ActiveConnections float64 `json:"active_connections"`
    CPUUsage         float64 `json:"cpu_usage"`
    MemoryUsage      float64 `json:"memory_usage"`
    GoroutineCount   float64 `json:"goroutine_count"`
}

// 获取告警状态
func (c *MonitoringController) GetAlerts() {
    alerts := c.monitor.GetActiveAlerts()
    c.JSON(200, alerts)
}
```

## 最佳实践

### 1. 指标命名规范

```go
// 好的指标命名
const (
    HTTPRequestsTotal        = "http_requests_total"
    HTTPRequestDuration     = "http_request_duration_seconds"
    HTTPResponseSize        = "http_response_size_bytes"
    DatabaseQueryDuration   = "database_query_duration_seconds"
    CacheOperationsTotal    = "cache_operations_total"
    WorkerQueueSize         = "worker_queue_size"
    BusinessProcessingTime  = "business_processing_duration_seconds"
)

// 避免的命名
const (
    Requests      = "requests"           // 不够具体
    ResponseTime  = "response_time"      // 缺少单位
    DBTime        = "db_time"           // 缩写不清楚
)
```

### 2. 标签使用

```go
// 合理的标签使用
monitor.Counter("http_requests_total").WithLabels(map[string]string{
    "method": "GET",
    "path":   "/api/users",     // 使用路径模板，不是具体路径
    "status": "200",
})

// 避免高基数标签
monitor.Counter("requests").WithLabels(map[string]string{
    "user_id": "12345",         // 高基数，会产生大量时间序列
    "request_id": "uuid",       // 每个请求都不同
    "timestamp": "1234567890",  // 时间戳不应该作为标签
})
```

### 3. 性能优化

```go
// 使用对象池减少内存分配
var labelPool = sync.Pool{
    New: func() interface{} {
        return make(map[string]string, 5)
    },
}

func recordMetric(method, path, status string) {
    labels := labelPool.Get().(map[string]string)
    defer labelPool.Put(labels)
    
    // 清空并重用
    for k := range labels {
        delete(labels, k)
    }
    
    labels["method"] = method
    labels["path"] = path
    labels["status"] = status
    
    monitor.Counter("http_requests_total").WithLabels(labels).Inc()
}

// 批量记录指标
type MetricsBatch struct {
    metrics []MetricEntry
    mutex   sync.Mutex
}

func (mb *MetricsBatch) Add(name string, value float64, labels map[string]string) {
    mb.mutex.Lock()
    defer mb.mutex.Unlock()
    
    mb.metrics = append(mb.metrics, MetricEntry{
        Name:   name,
        Value:  value,
        Labels: labels,
    })
}

func (mb *MetricsBatch) Flush() {
    mb.mutex.Lock()
    defer mb.mutex.Unlock()
    
    for _, metric := range mb.metrics {
        monitor.Gauge(metric.Name).WithLabels(metric.Labels).Set(metric.Value)
    }
    
    mb.metrics = mb.metrics[:0] // 重用切片
}
```

YYHertz 的性能监控系统提供了全面的监控解决方案，从基本的指标收集到高级的性能分析，帮助开发者构建高性能、可观测的应用程序。
