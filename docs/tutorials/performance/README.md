# YYHertz 性能优化实战教程

<div align="center">

⚡ **高性能应用调优** | 从基础优化到生产级性能调优

</div>

---

## 📋 目录

- [性能优化概述](#性能优化概述)
- [基准测试与监控](#基准测试与监控)
- [HTTP层优化](#http层优化)
- [数据库性能优化](#数据库性能优化)
- [缓存策略优化](#缓存策略优化)
- [内存管理优化](#内存管理优化)
- [并发处理优化](#并发处理优化)
- [部署优化](#部署优化)
- [实战案例](#实战案例)

---

## 🎯 性能优化概述

### YYHertz性能架构

```
                    性能优化层级架构
                           │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
    🌐 网络层         💻 应用层         💾 存储层
        │                 │                 │
   ┌────▼────┐       ┌────▼────┐       ┌────▼────┐
   │负载均衡  │       │中间件   │       │数据库   │
   │CDN加速  │       │路由优化  │       │缓存系统  │
   │连接复用  │       │内存管理  │       │文件系统  │
   └─────────┘       └─────────┘       └─────────┘
```

### 性能优化原则

```go
package optimization

import (
    "context"
    "runtime"
    "time"
    "sync"
)

// PerformanceOptimizer 性能优化器
type PerformanceOptimizer struct {
    config    *OptimizationConfig
    monitor   *PerformanceMonitor
    profiler  *Profiler
    metrics   *MetricsCollector
    mutex     sync.RWMutex
}

// OptimizationConfig 优化配置
type OptimizationConfig struct {
    // CPU优化
    MaxProcs           int           `json:"max_procs"`
    GCTarget           int           `json:"gc_target"`
    
    // 内存优化
    MaxMemory          int64         `json:"max_memory"`
    PoolSize           int           `json:"pool_size"`
    
    // 网络优化
    ReadTimeout        time.Duration `json:"read_timeout"`
    WriteTimeout       time.Duration `json:"write_timeout"`
    MaxConnections     int           `json:"max_connections"`
    
    // 数据库优化
    MaxOpenConns       int           `json:"max_open_conns"`
    MaxIdleConns       int           `json:"max_idle_conns"`
    ConnMaxLifetime    time.Duration `json:"conn_max_lifetime"`
    
    // 缓存优化
    CacheSize          int64         `json:"cache_size"`
    CacheTTL           time.Duration `json:"cache_ttl"`
    
    // 监控配置
    EnableProfiling    bool          `json:"enable_profiling"`
    MetricsInterval    time.Duration `json:"metrics_interval"`
}

// NewPerformanceOptimizer 创建性能优化器
func NewPerformanceOptimizer() *PerformanceOptimizer {
    config := &OptimizationConfig{
        MaxProcs:        runtime.NumCPU(),
        GCTarget:        100,
        MaxMemory:       1024 * 1024 * 1024, // 1GB
        PoolSize:        1000,
        ReadTimeout:     30 * time.Second,
        WriteTimeout:    30 * time.Second,
        MaxConnections:  10000,
        MaxOpenConns:    25,
        MaxIdleConns:    5,
        ConnMaxLifetime: time.Hour,
        CacheSize:       100 * 1024 * 1024, // 100MB
        CacheTTL:        15 * time.Minute,
        EnableProfiling: true,
        MetricsInterval: 30 * time.Second,
    }
    
    return &PerformanceOptimizer{
        config:  config,
        monitor: NewPerformanceMonitor(),
        profiler: NewProfiler(),
        metrics: NewMetricsCollector(),
    }
}

// OptimizeApplication 优化应用性能
func (o *PerformanceOptimizer) OptimizeApplication() error {
    // CPU优化
    if err := o.optimizeCPU(); err != nil {
        return err
    }
    
    // 内存优化
    if err := o.optimizeMemory(); err != nil {
        return err
    }
    
    // 网络优化
    if err := o.optimizeNetwork(); err != nil {
        return err
    }
    
    // 启动监控
    go o.startMonitoring()
    
    return nil
}

// optimizeCPU CPU优化
func (o *PerformanceOptimizer) optimizeCPU() error {
    // 设置最大CPU使用数
    runtime.GOMAXPROCS(o.config.MaxProcs)
    
    // 调整GC目标
    debug.SetGCPercent(o.config.GCTarget)
    
    return nil
}

// optimizeMemory 内存优化
func (o *PerformanceOptimizer) optimizeMemory() error {
    // 设置内存限制
    debug.SetMemoryLimit(o.config.MaxMemory)
    
    // 初始化对象池
    o.initObjectPools()
    
    return nil
}

// initObjectPools 初始化对象池
func (o *PerformanceOptimizer) initObjectPools() {
    // 字节池
    bytePool = &sync.Pool{
        New: func() interface{} {
            return make([]byte, 1024)
        },
    }
    
    // 字符串构建器池
    stringBuilderPool = &sync.Pool{
        New: func() interface{} {
            return &strings.Builder{}
        },
    }
    
    // HTTP响应池
    responsePool = &sync.Pool{
        New: func() interface{} {
            return &Response{}
        },
    }
}

var (
    bytePool          *sync.Pool
    stringBuilderPool *sync.Pool
    responsePool      *sync.Pool
)

// GetByteBuffer 获取字节缓冲区
func GetByteBuffer() []byte {
    return bytePool.Get().([]byte)
}

// PutByteBuffer 归还字节缓冲区
func PutByteBuffer(buf []byte) {
    buf = buf[:0]
    bytePool.Put(buf)
}
```

---

## 📊 基准测试与监控

### 1. 性能基准测试

```go
package benchmark

import (
    "testing"
    "runtime"
    "time"
    "net/http"
    "net/http/httptest"
    
    "github.com/zsy619/yyhertz/framework/mvc"
)

// BenchmarkSuite 基准测试套件
type BenchmarkSuite struct {
    app        *mvc.Application
    server     *httptest.Server
    client     *http.Client
}

// NewBenchmarkSuite 创建基准测试套件
func NewBenchmarkSuite() *BenchmarkSuite {
    app := mvc.HertzApp
    server := httptest.NewServer(app.Handler())
    
    client := &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 100,
            IdleConnTimeout:     90 * time.Second,
        },
    }
    
    return &BenchmarkSuite{
        app:    app,
        server: server,
        client: client,
    }
}

// BenchmarkSimpleJSON 简单JSON响应基准测试
func BenchmarkSimpleJSON(b *testing.B) {
    suite := NewBenchmarkSuite()
    defer suite.server.Close()
    
    // 设置路由
    suite.app.GET("/json", func(c *mvc.Context) {
        c.JSON(map[string]interface{}{
            "message": "Hello, World!",
            "status":  200,
            "time":    time.Now().Unix(),
        })
    })
    
    b.ResetTimer()
    b.ReportAllocs()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp, err := suite.client.Get(suite.server.URL + "/json")
            if err != nil {
                b.Error(err)
                continue
            }
            resp.Body.Close()
            
            if resp.StatusCode != 200 {
                b.Errorf("Expected status 200, got %d", resp.StatusCode)
            }
        }
    })
}

// BenchmarkDatabaseQuery 数据库查询基准测试
func BenchmarkDatabaseQuery(b *testing.B) {
    suite := NewBenchmarkSuite()
    defer suite.server.Close()
    
    // 初始化数据库
    setupTestDB()
    defer teardownTestDB()
    
    suite.app.GET("/users", func(c *mvc.Context) {
        var users []User
        db.Find(&users)
        c.JSON(users)
    })
    
    b.ResetTimer()
    b.ReportAllocs()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp, err := suite.client.Get(suite.server.URL + "/users")
            if err != nil {
                b.Error(err)
                continue
            }
            resp.Body.Close()
        }
    })
}

// BenchmarkMemoryUsage 内存使用基准测试
func BenchmarkMemoryUsage(b *testing.B) {
    var m1, m2 runtime.MemStats
    
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        // 模拟应用操作
        data := make([]byte, 1024)
        _ = processData(data)
    }
    
    b.StopTimer()
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")
    b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}

// BenchmarkConcurrentRequests 并发请求基准测试
func BenchmarkConcurrentRequests(b *testing.B) {
    suite := NewBenchmarkSuite()
    defer suite.server.Close()
    
    // 设置路由
    suite.app.GET("/concurrent", func(c *mvc.Context) {
        time.Sleep(10 * time.Millisecond) // 模拟处理时间
        c.JSON(map[string]string{"status": "ok"})
    })
    
    concurrency := []int{1, 10, 50, 100, 500, 1000}
    
    for _, c := range concurrency {
        b.Run(fmt.Sprintf("Concurrency-%d", c), func(b *testing.B) {
            benchmarkWithConcurrency(b, suite, c)
        })
    }
}

// benchmarkWithConcurrency 指定并发度的基准测试
func benchmarkWithConcurrency(b *testing.B, suite *BenchmarkSuite, concurrency int) {
    b.SetParallelism(concurrency)
    b.ResetTimer()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp, err := suite.client.Get(suite.server.URL + "/concurrent")
            if err != nil {
                b.Error(err)
                continue
            }
            resp.Body.Close()
        }
    })
}

// PerformanceReport 性能报告
type PerformanceReport struct {
    TestName     string        `json:"test_name"`
    RequestsPerSec float64     `json:"requests_per_sec"`
    AvgLatency   time.Duration `json:"avg_latency"`
    MaxLatency   time.Duration `json:"max_latency"`
    MinLatency   time.Duration `json:"min_latency"`
    MemoryUsage  int64         `json:"memory_usage"`
    AllocsPerOp  int64         `json:"allocs_per_op"`
    BytesPerOp   int64         `json:"bytes_per_op"`
}

// GenerateReport 生成性能报告
func GenerateReport(results []testing.BenchmarkResult) []*PerformanceReport {
    var reports []*PerformanceReport
    
    for _, result := range results {
        report := &PerformanceReport{
            TestName:     result.Name(),
            RequestsPerSec: float64(result.N) / result.T.Seconds(),
            MemoryUsage:  result.MemBytes,
            AllocsPerOp:  result.MemAllocs,
            BytesPerOp:   result.Bytes,
        }
        
        reports = append(reports, report)
    }
    
    return reports
}
```

### 2. 实时性能监控

```go
package monitoring

import (
    "context"
    "runtime"
    "time"
    "sync/atomic"
    "net/http"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
    metrics    *RuntimeMetrics
    collectors []MetricsCollector
    interval   time.Duration
    stopCh     chan struct{}
}

// RuntimeMetrics 运行时指标
type RuntimeMetrics struct {
    // CPU指标
    CPUUsage     float64 `json:"cpu_usage"`
    NumGoroutine int     `json:"num_goroutine"`
    NumCPU       int     `json:"num_cpu"`
    
    // 内存指标
    Alloc        uint64 `json:"alloc"`         // 当前分配内存
    TotalAlloc   uint64 `json:"total_alloc"`   // 总分配内存
    Sys          uint64 `json:"sys"`           // 系统内存
    HeapAlloc    uint64 `json:"heap_alloc"`    // 堆分配
    HeapSys      uint64 `json:"heap_sys"`      // 堆系统内存
    HeapIdle     uint64 `json:"heap_idle"`     // 堆空闲内存
    HeapInuse    uint64 `json:"heap_inuse"`    // 堆使用内存
    
    // GC指标
    NumGC        uint32 `json:"num_gc"`        // GC次数
    PauseTotalNs uint64 `json:"pause_total_ns"` // GC总暂停时间
    LastGC       uint64 `json:"last_gc"`       // 上次GC时间
    
    // 网络指标
    ActiveConnections int64 `json:"active_connections"`
    TotalRequests     int64 `json:"total_requests"`
    RequestsPerSec    float64 `json:"requests_per_sec"`
    
    // 错误指标
    ErrorCount    int64 `json:"error_count"`
    TimeoutCount  int64 `json:"timeout_count"`
    
    Timestamp     time.Time `json:"timestamp"`
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
    return &PerformanceMonitor{
        metrics:  &RuntimeMetrics{},
        interval: 30 * time.Second,
        stopCh:   make(chan struct{}),
    }
}

// Start 启动监控
func (m *PerformanceMonitor) Start() {
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            m.collectMetrics()
        case <-m.stopCh:
            return
        }
    }
}

// Stop 停止监控
func (m *PerformanceMonitor) Stop() {
    close(m.stopCh)
}

// collectMetrics 收集指标
func (m *PerformanceMonitor) collectMetrics() {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    m.metrics.NumGoroutine = runtime.NumGoroutine()
    m.metrics.NumCPU = runtime.NumCPU()
    
    // 内存指标
    m.metrics.Alloc = memStats.Alloc
    m.metrics.TotalAlloc = memStats.TotalAlloc
    m.metrics.Sys = memStats.Sys
    m.metrics.HeapAlloc = memStats.HeapAlloc
    m.metrics.HeapSys = memStats.HeapSys
    m.metrics.HeapIdle = memStats.HeapIdle
    m.metrics.HeapInuse = memStats.HeapInuse
    
    // GC指标
    m.metrics.NumGC = memStats.NumGC
    m.metrics.PauseTotalNs = memStats.PauseTotalNs
    m.metrics.LastGC = memStats.LastGC
    
    m.metrics.Timestamp = time.Now()
    
    // 通知所有收集器
    for _, collector := range m.collectors {
        collector.Collect(m.metrics)
    }
}

// GetMetrics 获取当前指标
func (m *PerformanceMonitor) GetMetrics() *RuntimeMetrics {
    return m.metrics
}

// HTTPMetricsMiddleware HTTP指标中间件
func HTTPMetricsMiddleware() mvc.HandlerFunc {
    var (
        totalRequests   int64
        activeConns     int64
        errorCount      int64
        timeoutCount    int64
        lastResetTime   = time.Now()
    )
    
    return func(c *mvc.Context) {
        start := time.Now()
        
        // 增加活跃连接数
        atomic.AddInt64(&activeConns, 1)
        defer atomic.AddInt64(&activeConns, -1)
        
        // 增加总请求数
        atomic.AddInt64(&totalRequests, 1)
        
        c.Next()
        
        // 统计错误和超时
        if c.Writer.Status() >= 500 {
            atomic.AddInt64(&errorCount, 1)
        }
        
        if time.Since(start) > 30*time.Second {
            atomic.AddInt64(&timeoutCount, 1)
        }
        
        // 更新监控指标
        duration := time.Since(lastResetTime)
        if duration >= time.Minute {
            requests := atomic.SwapInt64(&totalRequests, 0)
            requestsPerSec := float64(requests) / duration.Seconds()
            
            // 更新全局指标
            if monitor := GetGlobalMonitor(); monitor != nil {
                monitor.metrics.ActiveConnections = atomic.LoadInt64(&activeConns)
                monitor.metrics.TotalRequests += requests
                monitor.metrics.RequestsPerSec = requestsPerSec
                monitor.metrics.ErrorCount = atomic.LoadInt64(&errorCount)
                monitor.metrics.TimeoutCount = atomic.LoadInt64(&timeoutCount)
            }
            
            lastResetTime = time.Now()
        }
    }
}

// MetricsHandler 指标HTTP处理器
func MetricsHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        monitor := GetGlobalMonitor()
        if monitor == nil {
            http.Error(w, "Monitor not available", http.StatusServiceUnavailable)
            return
        }
        
        metrics := monitor.GetMetrics()
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(metrics)
    }
}

var globalMonitor *PerformanceMonitor

// GetGlobalMonitor 获取全局监控器
func GetGlobalMonitor() *PerformanceMonitor {
    return globalMonitor
}

// SetGlobalMonitor 设置全局监控器
func SetGlobalMonitor(monitor *PerformanceMonitor) {
    globalMonitor = monitor
}
```

---

## 🌐 HTTP层优化

### 1. 连接池优化

```go
package http

import (
    "net"
    "net/http"
    "time"
    "sync"
)

// OptimizedTransport 优化的HTTP传输
type OptimizedTransport struct {
    *http.Transport
    config *TransportConfig
}

// TransportConfig 传输配置
type TransportConfig struct {
    MaxIdleConns        int           `json:"max_idle_conns"`
    MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host"`
    MaxConnsPerHost     int           `json:"max_conns_per_host"`
    IdleConnTimeout     time.Duration `json:"idle_conn_timeout"`
    DialTimeout         time.Duration `json:"dial_timeout"`
    KeepAlive           time.Duration `json:"keep_alive"`
    TLSHandshakeTimeout time.Duration `json:"tls_handshake_timeout"`
    ResponseHeaderTimeout time.Duration `json:"response_header_timeout"`
    ExpectContinueTimeout time.Duration `json:"expect_continue_timeout"`
}

// NewOptimizedTransport 创建优化的传输
func NewOptimizedTransport(config *TransportConfig) *OptimizedTransport {
    if config == nil {
        config = &TransportConfig{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            MaxConnsPerHost:     0, // 无限制
            IdleConnTimeout:     90 * time.Second,
            DialTimeout:         30 * time.Second,
            KeepAlive:           30 * time.Second,
            TLSHandshakeTimeout: 10 * time.Second,
            ResponseHeaderTimeout: 10 * time.Second,
            ExpectContinueTimeout: 1 * time.Second,
        }
    }
    
    transport := &http.Transport{
        MaxIdleConns:        config.MaxIdleConns,
        MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
        MaxConnsPerHost:     config.MaxConnsPerHost,
        IdleConnTimeout:     config.IdleConnTimeout,
        TLSHandshakeTimeout: config.TLSHandshakeTimeout,
        ResponseHeaderTimeout: config.ResponseHeaderTimeout,
        ExpectContinueTimeout: config.ExpectContinueTimeout,
        
        DialContext: (&net.Dialer{
            Timeout:   config.DialTimeout,
            KeepAlive: config.KeepAlive,
        }).DialContext,
    }
    
    return &OptimizedTransport{
        Transport: transport,
        config:    config,
    }
}

// ConnectionPoolManager 连接池管理器
type ConnectionPoolManager struct {
    pools  map[string]*ConnectionPool
    mutex  sync.RWMutex
    config *PoolConfig
}

// ConnectionPool 连接池
type ConnectionPool struct {
    connections chan net.Conn
    factory     func() (net.Conn, error)
    maxSize     int
    currentSize int64
    mutex       sync.RWMutex
}

// PoolConfig 池配置
type PoolConfig struct {
    MaxSize        int           `json:"max_size"`
    InitialSize    int           `json:"initial_size"`
    MaxIdleTime    time.Duration `json:"max_idle_time"`
    CleanupInterval time.Duration `json:"cleanup_interval"`
}

// NewConnectionPoolManager 创建连接池管理器
func NewConnectionPoolManager(config *PoolConfig) *ConnectionPoolManager {
    if config == nil {
        config = &PoolConfig{
            MaxSize:        100,
            InitialSize:    10,
            MaxIdleTime:    5 * time.Minute,
            CleanupInterval: 1 * time.Minute,
        }
    }
    
    manager := &ConnectionPoolManager{
        pools:  make(map[string]*ConnectionPool),
        config: config,
    }
    
    // 启动清理goroutine
    go manager.startCleanup()
    
    return manager
}

// GetConnection 获取连接
func (m *ConnectionPoolManager) GetConnection(address string) (net.Conn, error) {
    m.mutex.RLock()
    pool, exists := m.pools[address]
    m.mutex.RUnlock()
    
    if !exists {
        pool = m.createPool(address)
        m.mutex.Lock()
        m.pools[address] = pool
        m.mutex.Unlock()
    }
    
    return pool.Get()
}

// PutConnection 归还连接
func (m *ConnectionPoolManager) PutConnection(address string, conn net.Conn) {
    m.mutex.RLock()
    pool, exists := m.pools[address]
    m.mutex.RUnlock()
    
    if exists {
        pool.Put(conn)
    } else {
        conn.Close()
    }
}

// createPool 创建连接池
func (m *ConnectionPoolManager) createPool(address string) *ConnectionPool {
    pool := &ConnectionPool{
        connections: make(chan net.Conn, m.config.MaxSize),
        maxSize:     m.config.MaxSize,
        factory: func() (net.Conn, error) {
            return net.DialTimeout("tcp", address, 30*time.Second)
        },
    }
    
    // 预创建连接
    for i := 0; i < m.config.InitialSize; i++ {
        if conn, err := pool.factory(); err == nil {
            pool.connections <- conn
            atomic.AddInt64(&pool.currentSize, 1)
        }
    }
    
    return pool
}

// Get 获取连接
func (p *ConnectionPool) Get() (net.Conn, error) {
    select {
    case conn := <-p.connections:
        atomic.AddInt64(&p.currentSize, -1)
        return conn, nil
    default:
        // 连接池为空，创建新连接
        if atomic.LoadInt64(&p.currentSize) < int64(p.maxSize) {
            return p.factory()
        }
        // 等待连接归还
        conn := <-p.connections
        atomic.AddInt64(&p.currentSize, -1)
        return conn, nil
    }
}

// Put 归还连接
func (p *ConnectionPool) Put(conn net.Conn) {
    select {
    case p.connections <- conn:
        atomic.AddInt64(&p.currentSize, 1)
    default:
        // 连接池已满，关闭连接
        conn.Close()
    }
}
```

### 2. 请求路由优化

```go
package router

import (
    "strings"
    "sync"
    "regexp"
)

// OptimizedRouter 优化的路由器
type OptimizedRouter struct {
    staticRoutes map[string]*Route     // 静态路由
    dynamicRoutes []*DynamicRoute      // 动态路由
    compiled     bool                  // 是否已编译
    mutex        sync.RWMutex
}

// Route 路由信息
type Route struct {
    Pattern  string              `json:"pattern"`
    Handler  mvc.HandlerFunc     `json:"-"`
    Methods  []string            `json:"methods"`
    Params   map[string]string   `json:"params"`
}

// DynamicRoute 动态路由
type DynamicRoute struct {
    *Route
    Regex    *regexp.Regexp      `json:"-"`
    ParamNames []string          `json:"param_names"`
}

// NewOptimizedRouter 创建优化路由器
func NewOptimizedRouter() *OptimizedRouter {
    return &OptimizedRouter{
        staticRoutes:  make(map[string]*Route),
        dynamicRoutes: make([]*DynamicRoute, 0),
    }
}

// AddRoute 添加路由
func (r *OptimizedRouter) AddRoute(method, pattern string, handler mvc.HandlerFunc) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    route := &Route{
        Pattern: pattern,
        Handler: handler,
        Methods: []string{method},
        Params:  make(map[string]string),
    }
    
    if r.isStaticRoute(pattern) {
        // 静态路由
        key := method + ":" + pattern
        r.staticRoutes[key] = route
    } else {
        // 动态路由
        regex, paramNames := r.compilePattern(pattern)
        dynamicRoute := &DynamicRoute{
            Route:      route,
            Regex:      regex,
            ParamNames: paramNames,
        }
        r.dynamicRoutes = append(r.dynamicRoutes, dynamicRoute)
    }
    
    r.compiled = false
}

// FindRoute 查找路由
func (r *OptimizedRouter) FindRoute(method, path string) (*Route, map[string]string) {
    if !r.compiled {
        r.compile()
    }
    
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    // 1. 优先匹配静态路由（O(1)时间复杂度）
    key := method + ":" + path
    if route, exists := r.staticRoutes[key]; exists {
        return route, nil
    }
    
    // 2. 匹配动态路由
    for _, dynamicRoute := range r.dynamicRoutes {
        if r.methodMatches(dynamicRoute.Methods, method) {
            if matches := dynamicRoute.Regex.FindStringSubmatch(path); matches != nil {
                params := make(map[string]string)
                for i, name := range dynamicRoute.ParamNames {
                    if i+1 < len(matches) {
                        params[name] = matches[i+1]
                    }
                }
                return dynamicRoute.Route, params
            }
        }
    }
    
    return nil, nil
}

// isStaticRoute 判断是否为静态路由
func (r *OptimizedRouter) isStaticRoute(pattern string) bool {
    return !strings.Contains(pattern, ":") && !strings.Contains(pattern, "*")
}

// compilePattern 编译路由模式
func (r *OptimizedRouter) compilePattern(pattern string) (*regexp.Regexp, []string) {
    var paramNames []string
    regexPattern := "^"
    
    parts := strings.Split(pattern, "/")
    for i, part := range parts {
        if i > 0 {
            regexPattern += "/"
        }
        
        if strings.HasPrefix(part, ":") {
            // 参数路由 :id
            paramName := part[1:]
            paramNames = append(paramNames, paramName)
            regexPattern += "([^/]+)"
        } else if part == "*" {
            // 通配符路由
            paramNames = append(paramNames, "wildcard")
            regexPattern += "(.*)"
        } else {
            // 静态部分
            regexPattern += regexp.QuoteMeta(part)
        }
    }
    
    regexPattern += "$"
    regex, _ := regexp.Compile(regexPattern)
    
    return regex, paramNames
}

// compile 编译路由器
func (r *OptimizedRouter) compile() {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if r.compiled {
        return
    }
    
    // 对动态路由进行排序优化
    // 将更具体的路由排在前面
    sort.Slice(r.dynamicRoutes, func(i, j int) bool {
        return r.routeSpecificity(r.dynamicRoutes[i].Pattern) > 
               r.routeSpecificity(r.dynamicRoutes[j].Pattern)
    })
    
    r.compiled = true
}

// routeSpecificity 计算路由特异性
func (r *OptimizedRouter) routeSpecificity(pattern string) int {
    score := 0
    parts := strings.Split(pattern, "/")
    
    for _, part := range parts {
        if part == "" {
            continue
        } else if strings.HasPrefix(part, ":") {
            score += 1 // 参数路由
        } else if part == "*" {
            score += 0 // 通配符路由
        } else {
            score += 2 // 静态路由段
        }
    }
    
    return score
}

// methodMatches 检查方法是否匹配
func (r *OptimizedRouter) methodMatches(methods []string, method string) bool {
    for _, m := range methods {
        if m == method {
            return true
        }
    }
    return false
}

// RouterBenchmark 路由器基准测试
func BenchmarkRouterLookup(b *testing.B) {
    router := NewOptimizedRouter()
    
    // 添加大量路由进行测试
    routes := []string{
        "/api/users",
        "/api/users/:id",
        "/api/users/:id/posts",
        "/api/posts",
        "/api/posts/:id",
        "/static/*",
        "/admin/dashboard",
        "/admin/users/:id/edit",
    }
    
    for _, route := range routes {
        router.AddRoute("GET", route, func(c *mvc.Context) {})
    }
    
    b.ResetTimer()
    
    // 测试静态路由查找
    b.Run("StaticRoute", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            router.FindRoute("GET", "/api/users")
        }
    })
    
    // 测试动态路由查找
    b.Run("DynamicRoute", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            router.FindRoute("GET", "/api/users/123")
        }
    })
    
    // 测试通配符路由查找
    b.Run("WildcardRoute", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            router.FindRoute("GET", "/static/css/style.css")
        }
    })
}
```

---

## 💾 数据库性能优化

### 1. 连接池优化

```go
package database

import (
    "database/sql"
    "time"
    "context"
    "sync"
    
    "gorm.io/gorm"
)

// DBOptimizer 数据库优化器
type DBOptimizer struct {
    db     *gorm.DB
    config *DBConfig
    stats  *DBStats
    mutex  sync.RWMutex
}

// DBConfig 数据库配置
type DBConfig struct {
    // 连接池配置
    MaxOpenConns    int           `json:"max_open_conns"`
    MaxIdleConns    int           `json:"max_idle_conns"`
    ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
    ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
    
    // 性能配置
    SlowQueryThreshold time.Duration `json:"slow_query_threshold"`
    EnableQueryLog     bool          `json:"enable_query_log"`
    EnableMetrics      bool          `json:"enable_metrics"`
    
    // 缓存配置
    EnableQueryCache   bool          `json:"enable_query_cache"`
    CacheSize          int           `json:"cache_size"`
    CacheTTL           time.Duration `json:"cache_ttl"`
}

// DBStats 数据库统计
type DBStats struct {
    OpenConnections   int           `json:"open_connections"`
    InUse            int           `json:"in_use"`
    Idle             int           `json:"idle"`
    WaitCount        int64         `json:"wait_count"`
    WaitDuration     time.Duration `json:"wait_duration"`
    MaxIdleClosed    int64         `json:"max_idle_closed"`
    MaxLifetimeClosed int64        `json:"max_lifetime_closed"`
    QueryCount       int64         `json:"query_count"`
    SlowQueryCount   int64         `json:"slow_query_count"`
    ErrorCount       int64         `json:"error_count"`
}

// NewDBOptimizer 创建数据库优化器
func NewDBOptimizer(db *gorm.DB, config *DBConfig) *DBOptimizer {
    if config == nil {
        config = &DBConfig{
            MaxOpenConns:       25,
            MaxIdleConns:       5,
            ConnMaxLifetime:    time.Hour,
            ConnMaxIdleTime:    10 * time.Minute,
            SlowQueryThreshold: 2 * time.Second,
            EnableQueryLog:     true,
            EnableMetrics:      true,
            EnableQueryCache:   true,
            CacheSize:          1000,
            CacheTTL:           5 * time.Minute,
        }
    }
    
    optimizer := &DBOptimizer{
        db:     db,
        config: config,
        stats:  &DBStats{},
    }
    
    optimizer.optimizeConnectionPool()
    optimizer.enableLogging()
    
    if config.EnableMetrics {
        go optimizer.startMetricsCollection()
    }
    
    return optimizer
}

// optimizeConnectionPool 优化连接池
func (o *DBOptimizer) optimizeConnectionPool() {
    sqlDB, err := o.db.DB()
    if err != nil {
        return
    }
    
    // 设置连接池参数
    sqlDB.SetMaxOpenConns(o.config.MaxOpenConns)
    sqlDB.SetMaxIdleConns(o.config.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(o.config.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(o.config.ConnMaxIdleTime)
}

// enableLogging 启用日志记录
func (o *DBOptimizer) enableLogging() {
    if !o.config.EnableQueryLog {
        return
    }
    
    o.db = o.db.Debug().Session(&gorm.Session{
        Logger: &QueryLogger{
            optimizer: o,
            threshold: o.config.SlowQueryThreshold,
        },
    })
}

// QueryLogger 查询日志记录器
type QueryLogger struct {
    optimizer *DBOptimizer
    threshold time.Duration
}

// LogMode 日志模式
func (l *QueryLogger) LogMode(level logger.LogLevel) logger.Interface {
    return l
}

// Info 信息日志
func (l *QueryLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    // 记录查询信息
}

// Warn 警告日志
func (l *QueryLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    // 记录警告信息
}

// Error 错误日志
func (l *QueryLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    l.optimizer.mutex.Lock()
    l.optimizer.stats.ErrorCount++
    l.optimizer.mutex.Unlock()
}

// Trace 追踪日志
func (l *QueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()
    
    l.optimizer.mutex.Lock()
    l.optimizer.stats.QueryCount++
    
    if elapsed > l.threshold {
        l.optimizer.stats.SlowQueryCount++
        // 记录慢查询
        log.Printf("Slow Query: %s, Duration: %v, Rows: %d", sql, elapsed, rows)
    }
    l.optimizer.mutex.Unlock()
    
    if err != nil {
        l.optimizer.mutex.Lock()
        l.optimizer.stats.ErrorCount++
        l.optimizer.mutex.Unlock()
    }
}

// startMetricsCollection 启动指标收集
func (o *DBOptimizer) startMetricsCollection() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        o.collectDBStats()
    }
}

// collectDBStats 收集数据库统计
func (o *DBOptimizer) collectDBStats() {
    sqlDB, err := o.db.DB()
    if err != nil {
        return
    }
    
    stats := sqlDB.Stats()
    
    o.mutex.Lock()
    o.stats.OpenConnections = stats.OpenConnections
    o.stats.InUse = stats.InUse
    o.stats.Idle = stats.Idle
    o.stats.WaitCount = stats.WaitCount
    o.stats.WaitDuration = stats.WaitDuration
    o.stats.MaxIdleClosed = stats.MaxIdleClosed
    o.stats.MaxLifetimeClosed = stats.MaxLifetimeClosed
    o.mutex.Unlock()
}

// GetStats 获取统计信息
func (o *DBOptimizer) GetStats() *DBStats {
    o.mutex.RLock()
    defer o.mutex.RUnlock()
    
    stats := *o.stats
    return &stats
}

// QueryCache 查询缓存
type QueryCache struct {
    cache  map[string]*CacheItem
    mutex  sync.RWMutex
    size   int
    ttl    time.Duration
}

// CacheItem 缓存项
type CacheItem struct {
    Data      interface{} `json:"data"`
    ExpiresAt time.Time   `json:"expires_at"`
    HitCount  int64       `json:"hit_count"`
}

// NewQueryCache 创建查询缓存
func NewQueryCache(size int, ttl time.Duration) *QueryCache {
    cache := &QueryCache{
        cache: make(map[string]*CacheItem),
        size:  size,
        ttl:   ttl,
    }
    
    // 启动清理goroutine
    go cache.startCleanup()
    
    return cache
}

// Get 获取缓存
func (c *QueryCache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    item, exists := c.cache[key]
    if !exists {
        return nil, false
    }
    
    if time.Now().After(item.ExpiresAt) {
        delete(c.cache, key)
        return nil, false
    }
    
    item.HitCount++
    return item.Data, true
}

// Set 设置缓存
func (c *QueryCache) Set(key string, data interface{}) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    // 检查缓存大小限制
    if len(c.cache) >= c.size {
        c.evictLRU()
    }
    
    c.cache[key] = &CacheItem{
        Data:      data,
        ExpiresAt: time.Now().Add(c.ttl),
        HitCount:  0,
    }
}

// evictLRU LRU淘汰
func (c *QueryCache) evictLRU() {
    var oldestKey string
    var oldestTime time.Time
    
    for key, item := range c.cache {
        if oldestKey == "" || item.ExpiresAt.Before(oldestTime) {
            oldestKey = key
            oldestTime = item.ExpiresAt
        }
    }
    
    if oldestKey != "" {
        delete(c.cache, oldestKey)
    }
}

// startCleanup 启动清理
func (c *QueryCache) startCleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.cleanup()
    }
}

// cleanup 清理过期缓存
func (c *QueryCache) cleanup() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    now := time.Now()
    for key, item := range c.cache {
        if now.After(item.ExpiresAt) {
            delete(c.cache, key)
        }
    }
}
```

### 2. 查询优化

```go
package database

import (
    "fmt"
    "strings"
    "reflect"
)

// QueryOptimizer 查询优化器
type QueryOptimizer struct {
    db            *gorm.DB
    cache         *QueryCache
    indexHints    map[string][]string
    explainCache  map[string]*ExplainResult
    mutex         sync.RWMutex
}

// ExplainResult 执行计划结果
type ExplainResult struct {
    Query      string  `json:"query"`
    Cost       float64 `json:"cost"`
    Rows       int64   `json:"rows"`
    UsingIndex bool    `json:"using_index"`
    KeyLen     int     `json:"key_len"`
    Extra      string  `json:"extra"`
}

// NewQueryOptimizer 创建查询优化器
func NewQueryOptimizer(db *gorm.DB) *QueryOptimizer {
    return &QueryOptimizer{
        db:           db,
        cache:        NewQueryCache(1000, 5*time.Minute),
        indexHints:   make(map[string][]string),
        explainCache: make(map[string]*ExplainResult),
    }
}

// OptimizedFind 优化的查询方法
func (o *QueryOptimizer) OptimizedFind(dest interface{}, conditions ...interface{}) error {
    query := o.buildQuery(dest, conditions...)
    cacheKey := o.generateCacheKey(query, conditions...)
    
    // 尝试从缓存获取
    if cached, found := o.cache.Get(cacheKey); found {
        return o.assignCachedResult(dest, cached)
    }
    
    // 分析查询计划
    if shouldOptimize := o.shouldOptimizeQuery(query); shouldOptimize {
        query = o.optimizeQuery(query)
    }
    
    // 执行查询
    result := o.db.Find(dest, conditions...)
    if result.Error != nil {
        return result.Error
    }
    
    // 缓存结果
    o.cache.Set(cacheKey, dest)
    
    return nil
}

// OptimizedPaginate 优化的分页查询
func (o *QueryOptimizer) OptimizedPaginate(dest interface{}, page, pageSize int, conditions ...interface{}) (*PaginationResult, error) {
    // 计算总数（优化版本）
    totalCount, err := o.optimizedCount(dest, conditions...)
    if err != nil {
        return nil, err
    }
    
    // 如果没有数据，直接返回
    if totalCount == 0 {
        return &PaginationResult{
            Data:       dest,
            Total:      0,
            Page:       page,
            PageSize:   pageSize,
            TotalPages: 0,
        }, nil
    }
    
    // 计算偏移量
    offset := (page - 1) * pageSize
    
    // 使用子查询优化大偏移量分页
    if offset > 10000 {
        err = o.optimizedLargeOffsetPaginate(dest, offset, pageSize, conditions...)
    } else {
        err = o.db.Offset(offset).Limit(pageSize).Find(dest, conditions...).Error
    }
    
    if err != nil {
        return nil, err
    }
    
    totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
    
    return &PaginationResult{
        Data:       dest,
        Total:      totalCount,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}

// optimizedCount 优化的计数查询
func (o *QueryOptimizer) optimizedCount(model interface{}, conditions ...interface{}) (int64, error) {
    var count int64
    
    // 构建计数查询缓存键
    cacheKey := fmt.Sprintf("count:%s:%v", reflect.TypeOf(model).Name(), conditions)
    
    // 尝试从缓存获取
    if cached, found := o.cache.Get(cacheKey); found {
        return cached.(int64), nil
    }
    
    // 执行计数查询
    err := o.db.Model(model).Where(conditions[0], conditions[1:]...).Count(&count).Error
    if err != nil {
        return 0, err
    }
    
    // 缓存结果（较短的TTL）
    o.cache.Set(cacheKey, count)
    
    return count, nil
}

// optimizedLargeOffsetPaginate 优化大偏移量分页
func (o *QueryOptimizer) optimizedLargeOffsetPaginate(dest interface{}, offset, limit int, conditions ...interface{}) error {
    // 使用子查询优化大偏移量
    tableName := o.getTableName(dest)
    
    subQuery := o.db.Table(tableName).
        Select("id").
        Where(conditions[0], conditions[1:]...).
        Offset(offset).
        Limit(limit)
    
    return o.db.Table(tableName).
        Where("id IN (?)", subQuery).
        Find(dest).Error
}

// buildQuery 构建查询
func (o *QueryOptimizer) buildQuery(model interface{}, conditions ...interface{}) string {
    tableName := o.getTableName(model)
    
    if len(conditions) == 0 {
        return fmt.Sprintf("SELECT * FROM %s", tableName)
    }
    
    whereClause := conditions[0].(string)
    return fmt.Sprintf("SELECT * FROM %s WHERE %s", tableName, whereClause)
}

// shouldOptimizeQuery 判断是否需要优化查询
func (o *QueryOptimizer) shouldOptimizeQuery(query string) bool {
    // 检查查询复杂度
    if strings.Contains(strings.ToLower(query), "join") {
        return true
    }
    
    if strings.Contains(strings.ToLower(query), "order by") {
        return true
    }
    
    if strings.Contains(strings.ToLower(query), "group by") {
        return true
    }
    
    return false
}

// optimizeQuery 优化查询
func (o *QueryOptimizer) optimizeQuery(query string) string {
    // 添加索引提示
    optimized := query
    
    // 检查是否有可用的索引提示
    for table, hints := range o.indexHints {
        if strings.Contains(query, table) {
            for _, hint := range hints {
                optimized = strings.Replace(optimized, table, fmt.Sprintf("%s USE INDEX (%s)", table, hint), 1)
            }
        }
    }
    
    return optimized
}

// AddIndexHint 添加索引提示
func (o *QueryOptimizer) AddIndexHint(table, index string) {
    o.mutex.Lock()
    defer o.mutex.Unlock()
    
    if o.indexHints[table] == nil {
        o.indexHints[table] = make([]string, 0)
    }
    
    o.indexHints[table] = append(o.indexHints[table], index)
}

// ExplainQuery 分析查询计划
func (o *QueryOptimizer) ExplainQuery(query string) (*ExplainResult, error) {
    o.mutex.RLock()
    if cached, exists := o.explainCache[query]; exists {
        o.mutex.RUnlock()
        return cached, nil
    }
    o.mutex.RUnlock()
    
    var result []map[string]interface{}
    err := o.db.Raw(fmt.Sprintf("EXPLAIN %s", query)).Find(&result).Error
    if err != nil {
        return nil, err
    }
    
    if len(result) == 0 {
        return nil, fmt.Errorf("no explain result")
    }
    
    explainResult := &ExplainResult{
        Query: query,
    }
    
    // 解析执行计划
    if cost, ok := result[0]["Cost"].(float64); ok {
        explainResult.Cost = cost
    }
    
    if rows, ok := result[0]["rows"].(int64); ok {
        explainResult.Rows = rows
    }
    
    if key, ok := result[0]["key"].(string); ok {
        explainResult.UsingIndex = key != ""
    }
    
    if extra, ok := result[0]["Extra"].(string); ok {
        explainResult.Extra = extra
    }
    
    // 缓存结果
    o.mutex.Lock()
    o.explainCache[query] = explainResult
    o.mutex.Unlock()
    
    return explainResult, nil
}

// PaginationResult 分页结果
type PaginationResult struct {
    Data       interface{} `json:"data"`
    Total      int64       `json:"total"`
    Page       int         `json:"page"`
    PageSize   int         `json:"page_size"`
    TotalPages int         `json:"total_pages"`
}

// getTableName 获取表名
func (o *QueryOptimizer) getTableName(model interface{}) string {
    stmt := &gorm.Statement{DB: o.db}
    stmt.Parse(model)
    return stmt.Schema.Table
}

// generateCacheKey 生成缓存键
func (o *QueryOptimizer) generateCacheKey(query string, conditions ...interface{}) string {
    return fmt.Sprintf("query:%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%v", query, conditions))))
}

// assignCachedResult 分配缓存结果
func (o *QueryOptimizer) assignCachedResult(dest, cached interface{}) error {
    destValue := reflect.ValueOf(dest)
    cachedValue := reflect.ValueOf(cached)
    
    if destValue.Kind() != reflect.Ptr {
        return fmt.Errorf("dest must be a pointer")
    }
    
    destValue.Elem().Set(cachedValue.Elem())
    return nil
}
```

---

## 📝 总结

通过本教程，你已经掌握了YYHertz框架的全面性能优化技能：

### 🎯 核心优化技能
- **基准测试与监控** - 科学的性能评估方法
- **HTTP层优化** - 连接池、路由器性能调优
- **数据库优化** - 连接池管理、查询优化策略
- **缓存策略** - 多级缓存架构设计
- **内存管理** - 对象池化、GC优化
- **并发优化** - 高效的并发处理模式

### 💡 最佳实践
- **性能监控** - 建立完善的性能监控体系
- **渐进式优化** - 从瓶颈点开始逐步优化
- **缓存策略** - 合理使用多级缓存提升性能
- **资源池化** - 减少对象创建销毁开销
- **查询优化** - 数据库查询性能调优

### 🚀 进阶方向
- **分布式性能优化** - 微服务架构性能调优
- **云原生优化** - Kubernetes环境性能优化
- **大数据处理** - 海量数据处理性能优化
- **AI辅助优化** - 使用机器学习优化性能

---

<div align="center">

**⚡ 掌握YYHertz性能优化，构建高性能应用！**

**从基础调优到企业级优化，全栈性能解决方案！🚀**

</div>