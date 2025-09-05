# YYHertz 监控告警体系教程

<div align="center">

📊 **企业级监控解决方案** | 从基础监控到智能告警

</div>

---

## 📋 目录

- [监控体系概述](#监控体系概述)
- [系统性能监控](#系统性能监控)
- [应用层监控](#应用层监控)
- [数据库监控](#数据库监控)
- [业务指标监控](#业务指标监控)
- [日志监控](#日志监控)
- [告警系统](#告警系统)
- [可视化dashboard](#可视化dashboard)
- [监控最佳实践](#监控最佳实践)

---

## 🎯 监控体系概述

### YYHertz监控架构

```
                    监控体系架构
                           │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
    📊 数据采集层      🔍 数据处理层      📈 数据展示层
        │                 │                 │
   ┌────▼────┐       ┌────▼────┐       ┌────▼────┐
   │Metrics  │       │时序数据库 │       │Grafana  │
   │Logs     │       │ElasticSearch│     │Kibana   │
   │Traces   │       │InfluxDB   │     │自定义面板 │
   └─────────┘       └─────────┘       └─────────┘
        │                 │                 │
   ┌────▼────┐       ┌────▼────┐       ┌────▼────┐
   │Agent    │       │聚合计算  │       │告警规则  │
   │Exporter │       │数据清理  │       │通知渠道  │
   │SDK埋点  │       │索引优化  │       │报告生成  │
   └─────────┘       └─────────┘       └─────────┘
```

### 监控组件架构

```go
package monitoring

import (
    "context"
    "sync"
    "time"
)

// MonitoringSystem 监控系统
type MonitoringSystem struct {
    metricsCollector  *MetricsCollector
    logsCollector     *LogsCollector
    tracesCollector   *TracesCollector
    alertManager      *AlertManager
    dashboard         *Dashboard
    config            *MonitoringConfig
    mutex             sync.RWMutex
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
    // 采集配置
    Collection CollectionConfig `json:"collection"`
    
    // 存储配置
    Storage StorageConfig `json:"storage"`
    
    // 告警配置
    Alerting AlertingConfig `json:"alerting"`
    
    // 可视化配置
    Visualization VisualizationConfig `json:"visualization"`
    
    // 性能配置
    Performance PerformanceConfig `json:"performance"`
}

// CollectionConfig 采集配置
type CollectionConfig struct {
    MetricsInterval   time.Duration `json:"metrics_interval"`
    LogsBufferSize    int          `json:"logs_buffer_size"`
    TracingSampleRate float64      `json:"tracing_sample_rate"`
    EnableAutoInstr   bool         `json:"enable_auto_instrumentation"`
}

// StorageConfig 存储配置
type StorageConfig struct {
    MetricsRetention string `json:"metrics_retention"`
    LogsRetention    string `json:"logs_retention"`
    TracesRetention  string `json:"traces_retention"`
    CompactionPolicy string `json:"compaction_policy"`
}

// AlertingConfig 告警配置
type AlertingConfig struct {
    EnableSlack     bool `json:"enable_slack"`
    EnableEmail     bool `json:"enable_email"`
    EnableWebhook   bool `json:"enable_webhook"`
    EvaluationInterval time.Duration `json:"evaluation_interval"`
}

// NewMonitoringSystem 创建监控系统
func NewMonitoringSystem(config *MonitoringConfig) *MonitoringSystem {
    if config == nil {
        config = &MonitoringConfig{
            Collection: CollectionConfig{
                MetricsInterval:   30 * time.Second,
                LogsBufferSize:    1000,
                TracingSampleRate: 0.1,
                EnableAutoInstr:   true,
            },
            Storage: StorageConfig{
                MetricsRetention: "30d",
                LogsRetention:    "7d",
                TracesRetention:  "3d",
                CompactionPolicy: "daily",
            },
            Alerting: AlertingConfig{
                EnableSlack:        true,
                EnableEmail:        true,
                EnableWebhook:      false,
                EvaluationInterval: 1 * time.Minute,
            },
        }
    }
    
    system := &MonitoringSystem{
        config:           config,
        metricsCollector: NewMetricsCollector(config.Collection),
        logsCollector:    NewLogsCollector(config.Collection),
        tracesCollector:  NewTracesCollector(config.Collection),
        alertManager:     NewAlertManager(config.Alerting),
        dashboard:        NewDashboard(config.Visualization),
    }
    
    return system
}

// Start 启动监控系统
func (ms *MonitoringSystem) Start(ctx context.Context) error {
    // 启动指标收集
    if err := ms.metricsCollector.Start(ctx); err != nil {
        return fmt.Errorf("failed to start metrics collector: %w", err)
    }
    
    // 启动日志收集
    if err := ms.logsCollector.Start(ctx); err != nil {
        return fmt.Errorf("failed to start logs collector: %w", err)
    }
    
    // 启动链路追踪收集
    if err := ms.tracesCollector.Start(ctx); err != nil {
        return fmt.Errorf("failed to start traces collector: %w", err)
    }
    
    // 启动告警管理
    if err := ms.alertManager.Start(ctx); err != nil {
        return fmt.Errorf("failed to start alert manager: %w", err)
    }
    
    // 启动仪表板
    if err := ms.dashboard.Start(ctx); err != nil {
        return fmt.Errorf("failed to start dashboard: %w", err)
    }
    
    return nil
}

// Stop 停止监控系统
func (ms *MonitoringSystem) Stop() error {
    var errors []error
    
    if err := ms.dashboard.Stop(); err != nil {
        errors = append(errors, err)
    }
    
    if err := ms.alertManager.Stop(); err != nil {
        errors = append(errors, err)
    }
    
    if err := ms.tracesCollector.Stop(); err != nil {
        errors = append(errors, err)
    }
    
    if err := ms.logsCollector.Stop(); err != nil {
        errors = append(errors, err)
    }
    
    if err := ms.metricsCollector.Stop(); err != nil {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("errors stopping monitoring system: %v", errors)
    }
    
    return nil
}

// GetHealthStatus 获取健康状态
func (ms *MonitoringSystem) GetHealthStatus() *HealthStatus {
    ms.mutex.RLock()
    defer ms.mutex.RUnlock()
    
    return &HealthStatus{
        Overall:          "healthy",
        MetricsCollector: ms.metricsCollector.GetStatus(),
        LogsCollector:    ms.logsCollector.GetStatus(),
        TracesCollector:  ms.tracesCollector.GetStatus(),
        AlertManager:     ms.alertManager.GetStatus(),
        Dashboard:        ms.dashboard.GetStatus(),
        Timestamp:        time.Now(),
    }
}

// HealthStatus 健康状态
type HealthStatus struct {
    Overall          string    `json:"overall"`
    MetricsCollector string    `json:"metrics_collector"`
    LogsCollector    string    `json:"logs_collector"`
    TracesCollector  string    `json:"traces_collector"`
    AlertManager     string    `json:"alert_manager"`
    Dashboard        string    `json:"dashboard"`
    Timestamp        time.Time `json:"timestamp"`
}

// MonitoringMiddleware 监控中间件
func MonitoringMiddleware(system *MonitoringSystem) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        start := time.Now()
        
        // 记录请求开始
        system.metricsCollector.IncrementCounter("http_requests_total", map[string]string{
            "method": c.Request.Method,
            "path":   c.Request.URL.Path,
        })
        
        // 开始追踪
        traceID := system.tracesCollector.StartTrace(c.Request.Context(), 
            "http_request", map[string]interface{}{
                "method": c.Request.Method,
                "path":   c.Request.URL.Path,
                "user_agent": c.Request.UserAgent(),
            })
        
        c.Set("trace_id", traceID)
        
        c.Next()
        
        // 记录请求完成
        duration := time.Since(start)
        status := c.Writer.Status()
        
        // 记录响应时间
        system.metricsCollector.RecordHistogram("http_request_duration_seconds", 
            duration.Seconds(), map[string]string{
                "method": c.Request.Method,
                "path":   c.Request.URL.Path,
                "status": fmt.Sprintf("%d", status),
            })
        
        // 记录状态码
        system.metricsCollector.IncrementCounter("http_responses_total", map[string]string{
            "method": c.Request.Method,
            "path":   c.Request.URL.Path,
            "status": fmt.Sprintf("%d", status),
        })
        
        // 结束追踪
        system.tracesCollector.EndTrace(traceID, map[string]interface{}{
            "status_code": status,
            "duration":    duration.Seconds(),
        })
        
        // 记录访问日志
        system.logsCollector.LogAccess(&AccessLog{
            Timestamp:    time.Now(),
            Method:       c.Request.Method,
            Path:         c.Request.URL.Path,
            StatusCode:   status,
            Duration:     duration,
            ClientIP:     c.ClientIP(),
            UserAgent:    c.Request.UserAgent(),
            TraceID:      traceID,
        })
        
        // 检查是否需要告警
        if status >= 500 {
            system.alertManager.TriggerAlert("http_error", map[string]interface{}{
                "message":     fmt.Sprintf("HTTP %d error on %s", status, c.Request.URL.Path),
                "method":      c.Request.Method,
                "path":        c.Request.URL.Path,
                "status_code": status,
                "client_ip":   c.ClientIP(),
                "trace_id":    traceID,
            })
        }
    }
}

// AccessLog 访问日志
type AccessLog struct {
    Timestamp  time.Time     `json:"timestamp"`
    Method     string        `json:"method"`
    Path       string        `json:"path"`
    StatusCode int           `json:"status_code"`
    Duration   time.Duration `json:"duration"`
    ClientIP   string        `json:"client_ip"`
    UserAgent  string        `json:"user_agent"`
    TraceID    string        `json:"trace_id"`
}
```

---

## 📊 系统性能监控

### 1. 系统指标收集器

```go
package monitoring

import (
    "runtime"
    "syscall"
    "time"
    "os"
    "strings"
    "strconv"
)

// SystemMetricsCollector 系统指标收集器
type SystemMetricsCollector struct {
    interval    time.Duration
    registry    *MetricsRegistry
    stopCh      chan struct{}
    lastCPUTimes CPUTimes
}

// CPUTimes CPU时间统计
type CPUTimes struct {
    User   uint64
    System uint64
    Idle   uint64
    Total  uint64
}

// SystemMetrics 系统指标
type SystemMetrics struct {
    // CPU指标
    CPUUsage        float64 `json:"cpu_usage"`
    CPULoadAvg1     float64 `json:"cpu_load_avg_1"`
    CPULoadAvg5     float64 `json:"cpu_load_avg_5"`
    CPULoadAvg15    float64 `json:"cpu_load_avg_15"`
    NumCPU          int     `json:"num_cpu"`
    NumGoroutines   int     `json:"num_goroutines"`
    
    // 内存指标
    MemoryAlloc     uint64  `json:"memory_alloc"`
    MemoryTotalAlloc uint64 `json:"memory_total_alloc"`
    MemorySys       uint64  `json:"memory_sys"`
    MemoryHeapAlloc uint64  `json:"memory_heap_alloc"`
    MemoryHeapSys   uint64  `json:"memory_heap_sys"`
    MemoryHeapIdle  uint64  `json:"memory_heap_idle"`
    MemoryHeapInuse uint64  `json:"memory_heap_inuse"`
    
    // GC指标
    GCNumCollections uint32  `json:"gc_num_collections"`
    GCPauseTotal     uint64  `json:"gc_pause_total"`
    GCPauseLast      uint64  `json:"gc_pause_last"`
    GCCPUFraction    float64 `json:"gc_cpu_fraction"`
    
    // 网络指标
    NetworkInBytes  uint64 `json:"network_in_bytes"`
    NetworkOutBytes uint64 `json:"network_out_bytes"`
    NetworkInPackets uint64 `json:"network_in_packets"`
    NetworkOutPackets uint64 `json:"network_out_packets"`
    
    // 磁盘指标
    DiskUsage       float64 `json:"disk_usage"`
    DiskInodeUsage  float64 `json:"disk_inode_usage"`
    DiskReadBytes   uint64  `json:"disk_read_bytes"`
    DiskWriteBytes  uint64  `json:"disk_write_bytes"`
    
    // 文件描述符
    OpenFileDescriptors uint64 `json:"open_file_descriptors"`
    MaxFileDescriptors  uint64 `json:"max_file_descriptors"`
    
    Timestamp       time.Time `json:"timestamp"`
}

// NewSystemMetricsCollector 创建系统指标收集器
func NewSystemMetricsCollector(interval time.Duration, registry *MetricsRegistry) *SystemMetricsCollector {
    return &SystemMetricsCollector{
        interval: interval,
        registry: registry,
        stopCh:   make(chan struct{}),
    }
}

// Start 启动收集器
func (smc *SystemMetricsCollector) Start(ctx context.Context) error {
    ticker := time.NewTicker(smc.interval)
    defer ticker.Stop()
    
    // 初始化CPU时间
    smc.lastCPUTimes = smc.getCPUTimes()
    
    go func() {
        for {
            select {
            case <-ticker.C:
                metrics := smc.collectSystemMetrics()
                smc.reportMetrics(metrics)
            case <-smc.stopCh:
                return
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return nil
}

// Stop 停止收集器
func (smc *SystemMetricsCollector) Stop() error {
    close(smc.stopCh)
    return nil
}

// collectSystemMetrics 收集系统指标
func (smc *SystemMetricsCollector) collectSystemMetrics() *SystemMetrics {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    metrics := &SystemMetrics{
        // CPU指标
        NumCPU:        runtime.NumCPU(),
        NumGoroutines: runtime.NumGoroutine(),
        
        // 内存指标
        MemoryAlloc:     memStats.Alloc,
        MemoryTotalAlloc: memStats.TotalAlloc,
        MemorySys:       memStats.Sys,
        MemoryHeapAlloc: memStats.HeapAlloc,
        MemoryHeapSys:   memStats.HeapSys,
        MemoryHeapIdle:  memStats.HeapIdle,
        MemoryHeapInuse: memStats.HeapInuse,
        
        // GC指标
        GCNumCollections: memStats.NumGC,
        GCPauseTotal:     memStats.PauseTotalNs,
        GCCPUFraction:    memStats.GCCPUFraction,
        
        Timestamp: time.Now(),
    }
    
    // 收集CPU使用率
    currentCPUTimes := smc.getCPUTimes()
    if smc.lastCPUTimes.Total > 0 {
        totalDiff := currentCPUTimes.Total - smc.lastCPUTimes.Total
        idleDiff := currentCPUTimes.Idle - smc.lastCPUTimes.Idle
        if totalDiff > 0 {
            metrics.CPUUsage = float64(totalDiff-idleDiff) / float64(totalDiff) * 100
        }
    }
    smc.lastCPUTimes = currentCPUTimes
    
    // 收集负载平均值
    loadAvg := smc.getLoadAverage()
    metrics.CPULoadAvg1 = loadAvg[0]
    metrics.CPULoadAvg5 = loadAvg[1]
    metrics.CPULoadAvg15 = loadAvg[2]
    
    // 收集网络指标
    netStats := smc.getNetworkStats()
    metrics.NetworkInBytes = netStats.InBytes
    metrics.NetworkOutBytes = netStats.OutBytes
    metrics.NetworkInPackets = netStats.InPackets
    metrics.NetworkOutPackets = netStats.OutPackets
    
    // 收集磁盘指标
    diskStats := smc.getDiskStats("/")
    metrics.DiskUsage = diskStats.UsagePercent
    metrics.DiskInodeUsage = diskStats.InodeUsagePercent
    
    // 收集文件描述符信息
    fdStats := smc.getFileDescriptorStats()
    metrics.OpenFileDescriptors = fdStats.Open
    metrics.MaxFileDescriptors = fdStats.Max
    
    // GC暂停时间
    if len(memStats.PauseNs) > 0 {
        metrics.GCPauseLast = memStats.PauseNs[(memStats.NumGC+255)%256]
    }
    
    return metrics
}

// getCPUTimes 获取CPU时间
func (smc *SystemMetricsCollector) getCPUTimes() CPUTimes {
    data, err := os.ReadFile("/proc/stat")
    if err != nil {
        return CPUTimes{}
    }
    
    lines := strings.Split(string(data), "\n")
    if len(lines) == 0 {
        return CPUTimes{}
    }
    
    fields := strings.Fields(lines[0])
    if len(fields) < 5 || fields[0] != "cpu" {
        return CPUTimes{}
    }
    
    user, _ := strconv.ParseUint(fields[1], 10, 64)
    system, _ := strconv.ParseUint(fields[3], 10, 64)
    idle, _ := strconv.ParseUint(fields[4], 10, 64)
    
    total := user + system + idle
    
    return CPUTimes{
        User:   user,
        System: system,
        Idle:   idle,
        Total:  total,
    }
}

// getLoadAverage 获取负载平均值
func (smc *SystemMetricsCollector) getLoadAverage() [3]float64 {
    data, err := os.ReadFile("/proc/loadavg")
    if err != nil {
        return [3]float64{}
    }
    
    fields := strings.Fields(string(data))
    if len(fields) < 3 {
        return [3]float64{}
    }
    
    var loadAvg [3]float64
    for i := 0; i < 3; i++ {
        if val, err := strconv.ParseFloat(fields[i], 64); err == nil {
            loadAvg[i] = val
        }
    }
    
    return loadAvg
}

// NetworkStats 网络统计
type NetworkStats struct {
    InBytes    uint64
    OutBytes   uint64
    InPackets  uint64
    OutPackets uint64
}

// getNetworkStats 获取网络统计
func (smc *SystemMetricsCollector) getNetworkStats() NetworkStats {
    data, err := os.ReadFile("/proc/net/dev")
    if err != nil {
        return NetworkStats{}
    }
    
    lines := strings.Split(string(data), "\n")
    var stats NetworkStats
    
    for _, line := range lines[2:] { // 跳过头两行
        if strings.TrimSpace(line) == "" {
            continue
        }
        
        fields := strings.Fields(line)
        if len(fields) < 17 {
            continue
        }
        
        // 跳过loopback接口
        if strings.HasPrefix(fields[0], "lo:") {
            continue
        }
        
        // 接收字节数和包数
        if inBytes, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
            stats.InBytes += inBytes
        }
        if inPackets, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
            stats.InPackets += inPackets
        }
        
        // 发送字节数和包数
        if outBytes, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
            stats.OutBytes += outBytes
        }
        if outPackets, err := strconv.ParseUint(fields[10], 10, 64); err == nil {
            stats.OutPackets += outPackets
        }
    }
    
    return stats
}

// DiskStats 磁盘统计
type DiskStats struct {
    UsagePercent      float64
    InodeUsagePercent float64
    ReadBytes         uint64
    WriteBytes        uint64
}

// getDiskStats 获取磁盘统计
func (smc *SystemMetricsCollector) getDiskStats(path string) DiskStats {
    var stat syscall.Statfs_t
    err := syscall.Statfs(path, &stat)
    if err != nil {
        return DiskStats{}
    }
    
    // 计算磁盘使用率
    total := stat.Blocks * uint64(stat.Bsize)
    free := stat.Bavail * uint64(stat.Bsize)
    used := total - free
    usagePercent := float64(used) / float64(total) * 100
    
    // 计算inode使用率
    totalInodes := stat.Files
    freeInodes := stat.Ffree
    usedInodes := totalInodes - freeInodes
    inodeUsagePercent := float64(usedInodes) / float64(totalInodes) * 100
    
    return DiskStats{
        UsagePercent:      usagePercent,
        InodeUsagePercent: inodeUsagePercent,
    }
}

// FileDescriptorStats 文件描述符统计
type FileDescriptorStats struct {
    Open uint64
    Max  uint64
}

// getFileDescriptorStats 获取文件描述符统计
func (smc *SystemMetricsCollector) getFileDescriptorStats() FileDescriptorStats {
    // 获取当前打开的文件描述符数量
    data, err := os.ReadFile("/proc/sys/fs/file-nr")
    if err != nil {
        return FileDescriptorStats{}
    }
    
    fields := strings.Fields(string(data))
    if len(fields) < 3 {
        return FileDescriptorStats{}
    }
    
    open, _ := strconv.ParseUint(fields[0], 10, 64)
    max, _ := strconv.ParseUint(fields[2], 10, 64)
    
    return FileDescriptorStats{
        Open: open,
        Max:  max,
    }
}

// reportMetrics 报告指标
func (smc *SystemMetricsCollector) reportMetrics(metrics *SystemMetrics) {
    // CPU指标
    smc.registry.SetGauge("system_cpu_usage_percent", metrics.CPUUsage, nil)
    smc.registry.SetGauge("system_load_average_1m", metrics.CPULoadAvg1, nil)
    smc.registry.SetGauge("system_load_average_5m", metrics.CPULoadAvg5, nil)
    smc.registry.SetGauge("system_load_average_15m", metrics.CPULoadAvg15, nil)
    smc.registry.SetGauge("system_num_goroutines", float64(metrics.NumGoroutines), nil)
    
    // 内存指标
    smc.registry.SetGauge("system_memory_alloc_bytes", float64(metrics.MemoryAlloc), nil)
    smc.registry.SetGauge("system_memory_sys_bytes", float64(metrics.MemorySys), nil)
    smc.registry.SetGauge("system_memory_heap_alloc_bytes", float64(metrics.MemoryHeapAlloc), nil)
    smc.registry.SetGauge("system_memory_heap_sys_bytes", float64(metrics.MemoryHeapSys), nil)
    smc.registry.SetGauge("system_memory_heap_idle_bytes", float64(metrics.MemoryHeapIdle), nil)
    smc.registry.SetGauge("system_memory_heap_inuse_bytes", float64(metrics.MemoryHeapInuse), nil)
    
    // GC指标
    smc.registry.SetGauge("system_gc_collections_total", float64(metrics.GCNumCollections), nil)
    smc.registry.SetGauge("system_gc_pause_total_seconds", float64(metrics.GCPauseTotal)/1e9, nil)
    smc.registry.SetGauge("system_gc_pause_last_seconds", float64(metrics.GCPauseLast)/1e9, nil)
    smc.registry.SetGauge("system_gc_cpu_fraction", metrics.GCCPUFraction, nil)
    
    // 网络指标
    smc.registry.SetGauge("system_network_receive_bytes_total", float64(metrics.NetworkInBytes), nil)
    smc.registry.SetGauge("system_network_transmit_bytes_total", float64(metrics.NetworkOutBytes), nil)
    smc.registry.SetGauge("system_network_receive_packets_total", float64(metrics.NetworkInPackets), nil)
    smc.registry.SetGauge("system_network_transmit_packets_total", float64(metrics.NetworkOutPackets), nil)
    
    // 磁盘指标
    smc.registry.SetGauge("system_disk_usage_percent", metrics.DiskUsage, nil)
    smc.registry.SetGauge("system_disk_inodes_usage_percent", metrics.DiskInodeUsage, nil)
    
    // 文件描述符指标
    smc.registry.SetGauge("system_file_descriptors_open", float64(metrics.OpenFileDescriptors), nil)
    smc.registry.SetGauge("system_file_descriptors_max", float64(metrics.MaxFileDescriptors), nil)
}
```

### 2. 应用性能监控

```go
package monitoring

import (
    "context"
    "sync"
    "sync/atomic"
    "time"
)

// ApplicationMetricsCollector 应用指标收集器
type ApplicationMetricsCollector struct {
    registry    *MetricsRegistry
    httpMetrics *HTTPMetrics
    dbMetrics   *DatabaseMetrics
    cacheMetrics *CacheMetrics
    customMetrics map[string]*CustomMetric
    mutex       sync.RWMutex
}

// HTTPMetrics HTTP指标
type HTTPMetrics struct {
    RequestCount     int64             `json:"request_count"`
    RequestDuration  *HistogramMetric  `json:"request_duration"`
    ResponseSizes    *HistogramMetric  `json:"response_sizes"`
    ActiveRequests   int64             `json:"active_requests"`
    RequestsByStatus map[int]int64     `json:"requests_by_status"`
    RequestsByMethod map[string]int64  `json:"requests_by_method"`
    mutex           sync.RWMutex
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
    ConnectionsActive int64   `json:"connections_active"`
    ConnectionsIdle   int64   `json:"connections_idle"`
    QueryCount        int64   `json:"query_count"`
    QueryDuration     *HistogramMetric `json:"query_duration"`
    QueryErrors       int64   `json:"query_errors"`
    TransactionCount  int64   `json:"transaction_count"`
    SlowQueries       int64   `json:"slow_queries"`
    mutex            sync.RWMutex
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
    HitCount    int64   `json:"hit_count"`
    MissCount   int64   `json:"miss_count"`
    HitRatio    float64 `json:"hit_ratio"`
    KeyCount    int64   `json:"key_count"`
    Memory      int64   `json:"memory_bytes"`
    Evictions   int64   `json:"evictions"`
    Operations  map[string]int64 `json:"operations"`
    mutex       sync.RWMutex
}

// CustomMetric 自定义指标
type CustomMetric struct {
    Name      string                 `json:"name"`
    Type      string                 `json:"type"` // counter, gauge, histogram
    Value     float64                `json:"value"`
    Labels    map[string]string      `json:"labels"`
    Histogram *HistogramMetric       `json:"histogram,omitempty"`
    UpdatedAt time.Time             `json:"updated_at"`
}

// HistogramMetric 直方图指标
type HistogramMetric struct {
    Count   int64     `json:"count"`
    Sum     float64   `json:"sum"`
    Buckets []Bucket  `json:"buckets"`
    mutex   sync.RWMutex
}

// Bucket 直方图桶
type Bucket struct {
    UpperBound float64 `json:"upper_bound"`
    Count      int64   `json:"count"`
}

// NewApplicationMetricsCollector 创建应用指标收集器
func NewApplicationMetricsCollector(registry *MetricsRegistry) *ApplicationMetricsCollector {
    return &ApplicationMetricsCollector{
        registry:    registry,
        httpMetrics: NewHTTPMetrics(),
        dbMetrics:   NewDatabaseMetrics(),
        cacheMetrics: NewCacheMetrics(),
        customMetrics: make(map[string]*CustomMetric),
    }
}

// NewHTTPMetrics 创建HTTP指标
func NewHTTPMetrics() *HTTPMetrics {
    return &HTTPMetrics{
        RequestDuration: NewHistogramMetric([]float64{
            0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
        }),
        ResponseSizes: NewHistogramMetric([]float64{
            100, 1000, 10000, 100000, 1000000, 10000000,
        }),
        RequestsByStatus: make(map[int]int64),
        RequestsByMethod: make(map[string]int64),
    }
}

// NewHistogramMetric 创建直方图指标
func NewHistogramMetric(buckets []float64) *HistogramMetric {
    histogram := &HistogramMetric{
        Buckets: make([]Bucket, len(buckets)),
    }
    
    for i, bound := range buckets {
        histogram.Buckets[i] = Bucket{UpperBound: bound}
    }
    
    return histogram
}

// RecordHTTPRequest 记录HTTP请求
func (amc *ApplicationMetricsCollector) RecordHTTPRequest(method string, statusCode int, duration time.Duration, responseSize int64) {
    amc.httpMetrics.mutex.Lock()
    defer amc.httpMetrics.mutex.Unlock()
    
    // 增加请求计数
    atomic.AddInt64(&amc.httpMetrics.RequestCount, 1)
    
    // 记录请求时长
    amc.httpMetrics.RequestDuration.Observe(duration.Seconds())
    
    // 记录响应大小
    amc.httpMetrics.ResponseSizes.Observe(float64(responseSize))
    
    // 按状态码统计
    amc.httpMetrics.RequestsByStatus[statusCode]++
    
    // 按方法统计
    amc.httpMetrics.RequestsByMethod[method]++
    
    // 报告到注册表
    amc.registry.IncrementCounter("http_requests_total", map[string]string{
        "method": method,
        "status": fmt.Sprintf("%d", statusCode),
    })
    
    amc.registry.RecordHistogram("http_request_duration_seconds", duration.Seconds(), map[string]string{
        "method": method,
    })
}

// StartHTTPRequest 开始HTTP请求
func (amc *ApplicationMetricsCollector) StartHTTPRequest() {
    atomic.AddInt64(&amc.httpMetrics.ActiveRequests, 1)
    amc.registry.SetGauge("http_requests_active", float64(amc.httpMetrics.ActiveRequests), nil)
}

// EndHTTPRequest 结束HTTP请求
func (amc *ApplicationMetricsCollector) EndHTTPRequest() {
    atomic.AddInt64(&amc.httpMetrics.ActiveRequests, -1)
    amc.registry.SetGauge("http_requests_active", float64(amc.httpMetrics.ActiveRequests), nil)
}

// RecordDatabaseQuery 记录数据库查询
func (amc *ApplicationMetricsCollector) RecordDatabaseQuery(operation string, duration time.Duration, success bool) {
    amc.dbMetrics.mutex.Lock()
    defer amc.dbMetrics.mutex.Unlock()
    
    // 增加查询计数
    atomic.AddInt64(&amc.dbMetrics.QueryCount, 1)
    
    // 记录查询时长
    amc.dbMetrics.QueryDuration.Observe(duration.Seconds())
    
    if !success {
        atomic.AddInt64(&amc.dbMetrics.QueryErrors, 1)
    }
    
    // 检查慢查询
    if duration > 2*time.Second {
        atomic.AddInt64(&amc.dbMetrics.SlowQueries, 1)
    }
    
    // 报告到注册表
    labels := map[string]string{"operation": operation}
    amc.registry.IncrementCounter("database_queries_total", labels)
    amc.registry.RecordHistogram("database_query_duration_seconds", duration.Seconds(), labels)
    
    if !success {
        amc.registry.IncrementCounter("database_errors_total", labels)
    }
}

// UpdateDatabaseConnections 更新数据库连接数
func (amc *ApplicationMetricsCollector) UpdateDatabaseConnections(active, idle int64) {
    atomic.StoreInt64(&amc.dbMetrics.ConnectionsActive, active)
    atomic.StoreInt64(&amc.dbMetrics.ConnectionsIdle, idle)
    
    amc.registry.SetGauge("database_connections_active", float64(active), nil)
    amc.registry.SetGauge("database_connections_idle", float64(idle), nil)
}

// RecordCacheOperation 记录缓存操作
func (amc *ApplicationMetricsCollector) RecordCacheOperation(operation string, hit bool) {
    amc.cacheMetrics.mutex.Lock()
    defer amc.cacheMetrics.mutex.Unlock()
    
    if amc.cacheMetrics.Operations == nil {
        amc.cacheMetrics.Operations = make(map[string]int64)
    }
    
    amc.cacheMetrics.Operations[operation]++
    
    if operation == "get" {
        if hit {
            atomic.AddInt64(&amc.cacheMetrics.HitCount, 1)
        } else {
            atomic.AddInt64(&amc.cacheMetrics.MissCount, 1)
        }
        
        // 计算命中率
        totalRequests := amc.cacheMetrics.HitCount + amc.cacheMetrics.MissCount
        if totalRequests > 0 {
            amc.cacheMetrics.HitRatio = float64(amc.cacheMetrics.HitCount) / float64(totalRequests)
        }
    }
    
    // 报告到注册表
    labels := map[string]string{
        "operation": operation,
        "result":    fmt.Sprintf("%t", hit),
    }
    amc.registry.IncrementCounter("cache_operations_total", labels)
    amc.registry.SetGauge("cache_hit_ratio", amc.cacheMetrics.HitRatio, nil)
}

// UpdateCacheStats 更新缓存统计
func (amc *ApplicationMetricsCollector) UpdateCacheStats(keyCount, memoryBytes, evictions int64) {
    atomic.StoreInt64(&amc.cacheMetrics.KeyCount, keyCount)
    atomic.StoreInt64(&amc.cacheMetrics.Memory, memoryBytes)
    atomic.StoreInt64(&amc.cacheMetrics.Evictions, evictions)
    
    amc.registry.SetGauge("cache_keys_total", float64(keyCount), nil)
    amc.registry.SetGauge("cache_memory_bytes", float64(memoryBytes), nil)
    amc.registry.SetGauge("cache_evictions_total", float64(evictions), nil)
}

// SetCustomMetric 设置自定义指标
func (amc *ApplicationMetricsCollector) SetCustomMetric(name, metricType string, value float64, labels map[string]string) {
    amc.mutex.Lock()
    defer amc.mutex.Unlock()
    
    metric := &CustomMetric{
        Name:      name,
        Type:      metricType,
        Value:     value,
        Labels:    labels,
        UpdatedAt: time.Now(),
    }
    
    amc.customMetrics[name] = metric
    
    // 报告到注册表
    switch metricType {
    case "counter":
        amc.registry.IncrementCounter(name, labels)
    case "gauge":
        amc.registry.SetGauge(name, value, labels)
    case "histogram":
        if metric.Histogram == nil {
            metric.Histogram = NewHistogramMetric([]float64{
                0.1, 0.5, 1, 2, 5, 10, 20, 50, 100,
            })
        }
        metric.Histogram.Observe(value)
        amc.registry.RecordHistogram(name, value, labels)
    }
}

// Observe 观察直方图指标
func (h *HistogramMetric) Observe(value float64) {
    h.mutex.Lock()
    defer h.mutex.Unlock()
    
    h.Count++
    h.Sum += value
    
    for i := range h.Buckets {
        if value <= h.Buckets[i].UpperBound {
            h.Buckets[i].Count++
        }
    }
}

// GetPercentile 获取百分位数
func (h *HistogramMetric) GetPercentile(percentile float64) float64 {
    h.mutex.RLock()
    defer h.mutex.RUnlock()
    
    if h.Count == 0 {
        return 0
    }
    
    targetCount := int64(float64(h.Count) * percentile / 100)
    
    for _, bucket := range h.Buckets {
        if bucket.Count >= targetCount {
            return bucket.UpperBound
        }
    }
    
    if len(h.Buckets) > 0 {
        return h.Buckets[len(h.Buckets)-1].UpperBound
    }
    
    return 0
}

// GetApplicationMetrics 获取应用指标
func (amc *ApplicationMetricsCollector) GetApplicationMetrics() map[string]interface{} {
    amc.mutex.RLock()
    defer amc.mutex.RUnlock()
    
    return map[string]interface{}{
        "http": map[string]interface{}{
            "request_count":      amc.httpMetrics.RequestCount,
            "active_requests":    amc.httpMetrics.ActiveRequests,
            "requests_by_status": amc.httpMetrics.RequestsByStatus,
            "requests_by_method": amc.httpMetrics.RequestsByMethod,
            "avg_duration":       amc.httpMetrics.RequestDuration.Sum / float64(amc.httpMetrics.RequestDuration.Count),
            "p95_duration":       amc.httpMetrics.RequestDuration.GetPercentile(95),
            "p99_duration":       amc.httpMetrics.RequestDuration.GetPercentile(99),
        },
        "database": map[string]interface{}{
            "connections_active": amc.dbMetrics.ConnectionsActive,
            "connections_idle":   amc.dbMetrics.ConnectionsIdle,
            "query_count":        amc.dbMetrics.QueryCount,
            "query_errors":       amc.dbMetrics.QueryErrors,
            "slow_queries":       amc.dbMetrics.SlowQueries,
            "avg_query_duration": amc.dbMetrics.QueryDuration.Sum / float64(amc.dbMetrics.QueryDuration.Count),
        },
        "cache": map[string]interface{}{
            "hit_count":   amc.cacheMetrics.HitCount,
            "miss_count":  amc.cacheMetrics.MissCount,
            "hit_ratio":   amc.cacheMetrics.HitRatio,
            "key_count":   amc.cacheMetrics.KeyCount,
            "memory":      amc.cacheMetrics.Memory,
            "evictions":   amc.cacheMetrics.Evictions,
            "operations":  amc.cacheMetrics.Operations,
        },
        "custom": amc.customMetrics,
    }
}

// ApplicationMonitoringMiddleware 应用监控中间件
func ApplicationMonitoringMiddleware(collector *ApplicationMetricsCollector) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        start := time.Now()
        
        // 开始请求监控
        collector.StartHTTPRequest()
        defer collector.EndHTTPRequest()
        
        // 处理请求
        c.Next()
        
        // 记录请求指标
        duration := time.Since(start)
        statusCode := c.Writer.Status()
        responseSize := int64(c.Writer.Size())
        
        collector.RecordHTTPRequest(c.Request.Method, statusCode, duration, responseSize)
        
        // 记录业务指标
        if userID, exists := c.Get("user_id"); exists {
            collector.SetCustomMetric("active_users", "gauge", 1, map[string]string{
                "user_id": fmt.Sprintf("%v", userID),
            })
        }
    }
}
```

---

## 🔔 告警系统

### 1. 告警管理器

```go
package monitoring

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "sync"
    "time"
)

// AlertManager 告警管理器
type AlertManager struct {
    config       *AlertingConfig
    rules        map[string]*AlertRule
    alerts       map[string]*Alert
    channels     []NotificationChannel
    evaluator    *RuleEvaluator
    mutex        sync.RWMutex
    stopCh       chan struct{}
}

// AlertRule 告警规则
type AlertRule struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Query       string                 `json:"query"`
    Condition   AlertCondition         `json:"condition"`
    Threshold   float64                `json:"threshold"`
    Duration    time.Duration          `json:"duration"`
    Labels      map[string]string      `json:"labels"`
    Annotations map[string]string      `json:"annotations"`
    Severity    AlertSeverity          `json:"severity"`
    Enabled     bool                   `json:"enabled"`
    CreatedAt   time.Time             `json:"created_at"`
    UpdatedAt   time.Time             `json:"updated_at"`
    
    // 内部状态
    LastEvaluatedAt time.Time         `json:"last_evaluated_at"`
    LastTriggeredAt time.Time         `json:"last_triggered_at"`
    EvaluationCount int64             `json:"evaluation_count"`
    TriggerCount    int64             `json:"trigger_count"`
}

// AlertCondition 告警条件
type AlertCondition string

const (
    ConditionGreaterThan    AlertCondition = "gt"
    ConditionLessThan       AlertCondition = "lt"
    ConditionEquals         AlertCondition = "eq"
    ConditionNotEquals      AlertCondition = "ne"
    ConditionContains       AlertCondition = "contains"
    ConditionRegex          AlertCondition = "regex"
)

// AlertSeverity 告警级别
type AlertSeverity string

const (
    SeverityCritical AlertSeverity = "critical"
    SeverityHigh     AlertSeverity = "high"
    SeverityMedium   AlertSeverity = "medium"
    SeverityLow      AlertSeverity = "low"
    SeverityInfo     AlertSeverity = "info"
)

// Alert 告警
type Alert struct {
    ID           string                 `json:"id"`
    RuleID       string                 `json:"rule_id"`
    RuleName     string                 `json:"rule_name"`
    Severity     AlertSeverity          `json:"severity"`
    Status       AlertStatus            `json:"status"`
    Message      string                 `json:"message"`
    Value        float64                `json:"value"`
    Labels       map[string]string      `json:"labels"`
    Annotations  map[string]string      `json:"annotations"`
    StartsAt     time.Time             `json:"starts_at"`
    EndsAt       time.Time             `json:"ends_at"`
    UpdatedAt    time.Time             `json:"updated_at"`
    
    // 通知状态
    NotificationsSent map[string]time.Time `json:"notifications_sent"`
    AcknowledgedBy   string               `json:"acknowledged_by"`
    AcknowledgedAt   time.Time            `json:"acknowledged_at"`
}

// AlertStatus 告警状态
type AlertStatus string

const (
    StatusFiring     AlertStatus = "firing"
    StatusResolved   AlertStatus = "resolved"
    StatusAcknowledged AlertStatus = "acknowledged"
    StatusSuppressed AlertStatus = "suppressed"
)

// NotificationChannel 通知渠道接口
type NotificationChannel interface {
    GetName() string
    Send(alert *Alert) error
    IsEnabled() bool
    GetConfig() map[string]interface{}
}

// NewAlertManager 创建告警管理器
func NewAlertManager(config *AlertingConfig) *AlertManager {
    am := &AlertManager{
        config:    config,
        rules:     make(map[string]*AlertRule),
        alerts:    make(map[string]*Alert),
        channels:  make([]NotificationChannel, 0),
        evaluator: NewRuleEvaluator(),
        stopCh:    make(chan struct{}),
    }
    
    // 初始化通知渠道
    if config.EnableSlack {
        am.AddChannel(NewSlackChannel())
    }
    if config.EnableEmail {
        am.AddChannel(NewEmailChannel())
    }
    if config.EnableWebhook {
        am.AddChannel(NewWebhookChannel())
    }
    
    return am
}

// Start 启动告警管理器
func (am *AlertManager) Start(ctx context.Context) error {
    ticker := time.NewTicker(am.config.EvaluationInterval)
    defer ticker.Stop()
    
    go func() {
        for {
            select {
            case <-ticker.C:
                am.evaluateRules()
            case <-am.stopCh:
                return
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return nil
}

// Stop 停止告警管理器
func (am *AlertManager) Stop() error {
    close(am.stopCh)
    return nil
}

// AddRule 添加告警规则
func (am *AlertManager) AddRule(rule *AlertRule) {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    
    rule.CreatedAt = time.Now()
    rule.UpdatedAt = time.Now()
    am.rules[rule.ID] = rule
}

// UpdateRule 更新告警规则
func (am *AlertManager) UpdateRule(rule *AlertRule) {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    
    if existingRule, exists := am.rules[rule.ID]; exists {
        rule.CreatedAt = existingRule.CreatedAt
        rule.UpdatedAt = time.Now()
        am.rules[rule.ID] = rule
    }
}

// DeleteRule 删除告警规则
func (am *AlertManager) DeleteRule(ruleID string) {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    
    delete(am.rules, ruleID)
    
    // 清理相关告警
    for alertID, alert := range am.alerts {
        if alert.RuleID == ruleID {
            delete(am.alerts, alertID)
        }
    }
}

// AddChannel 添加通知渠道
func (am *AlertManager) AddChannel(channel NotificationChannel) {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    
    am.channels = append(am.channels, channel)
}

// evaluateRules 评估告警规则
func (am *AlertManager) evaluateRules() {
    am.mutex.RLock()
    rules := make([]*AlertRule, 0, len(am.rules))
    for _, rule := range am.rules {
        if rule.Enabled {
            rules = append(rules, rule)
        }
    }
    am.mutex.RUnlock()
    
    for _, rule := range rules {
        am.evaluateRule(rule)
    }
}

// evaluateRule 评估单个告警规则
func (am *AlertManager) evaluateRule(rule *AlertRule) {
    now := time.Now()
    
    // 更新评估统计
    rule.LastEvaluatedAt = now
    rule.EvaluationCount++
    
    // 执行查询
    result, err := am.evaluator.EvaluateQuery(rule.Query)
    if err != nil {
        log.Printf("Error evaluating rule %s: %v", rule.ID, err)
        return
    }
    
    // 检查是否满足告警条件
    triggered := am.checkCondition(result, rule.Condition, rule.Threshold)
    
    alertID := fmt.Sprintf("%s_%s", rule.ID, rule.Name)
    
    am.mutex.Lock()
    existingAlert, alertExists := am.alerts[alertID]
    am.mutex.Unlock()
    
    if triggered {
        if !alertExists || existingAlert.Status == StatusResolved {
            // 创建新告警
            alert := &Alert{
                ID:           alertID,
                RuleID:       rule.ID,
                RuleName:     rule.Name,
                Severity:     rule.Severity,
                Status:       StatusFiring,
                Message:      am.formatAlertMessage(rule, result),
                Value:        result,
                Labels:       copyMap(rule.Labels),
                Annotations:  copyMap(rule.Annotations),
                StartsAt:     now,
                UpdatedAt:    now,
                NotificationsSent: make(map[string]time.Time),
            }
            
            am.mutex.Lock()
            am.alerts[alertID] = alert
            am.mutex.Unlock()
            
            // 更新规则统计
            rule.LastTriggeredAt = now
            rule.TriggerCount++
            
            // 发送通知
            am.sendNotifications(alert)
        } else {
            // 更新现有告警
            existingAlert.UpdatedAt = now
            existingAlert.Value = result
        }
    } else {
        if alertExists && existingAlert.Status == StatusFiring {
            // 解决告警
            existingAlert.Status = StatusResolved
            existingAlert.EndsAt = now
            existingAlert.UpdatedAt = now
            
            // 发送解决通知
            am.sendResolutionNotifications(existingAlert)
        }
    }
}

// checkCondition 检查告警条件
func (am *AlertManager) checkCondition(value float64, condition AlertCondition, threshold float64) bool {
    switch condition {
    case ConditionGreaterThan:
        return value > threshold
    case ConditionLessThan:
        return value < threshold
    case ConditionEquals:
        return value == threshold
    case ConditionNotEquals:
        return value != threshold
    default:
        return false
    }
}

// formatAlertMessage 格式化告警消息
func (am *AlertManager) formatAlertMessage(rule *AlertRule, value float64) string {
    return fmt.Sprintf("Alert: %s - Value: %.2f, Threshold: %.2f", 
        rule.Description, value, rule.Threshold)
}

// sendNotifications 发送告警通知
func (am *AlertManager) sendNotifications(alert *Alert) {
    for _, channel := range am.channels {
        if !channel.IsEnabled() {
            continue
        }
        
        // 检查是否已发送过通知
        channelName := channel.GetName()
        if _, sent := alert.NotificationsSent[channelName]; sent {
            continue
        }
        
        go func(ch NotificationChannel) {
            if err := ch.Send(alert); err != nil {
                log.Printf("Failed to send alert via %s: %v", ch.GetName(), err)
            } else {
                am.mutex.Lock()
                alert.NotificationsSent[channelName] = time.Now()
                am.mutex.Unlock()
            }
        }(channel)
    }
}

// sendResolutionNotifications 发送解决通知
func (am *AlertManager) sendResolutionNotifications(alert *Alert) {
    // 创建解决通知
    resolvedAlert := *alert
    resolvedAlert.Message = fmt.Sprintf("RESOLVED: %s", alert.Message)
    
    for _, channel := range am.channels {
        if !channel.IsEnabled() {
            continue
        }
        
        go func(ch NotificationChannel) {
            if err := ch.Send(&resolvedAlert); err != nil {
                log.Printf("Failed to send resolution via %s: %v", ch.GetName(), err)
            }
        }(channel)
    }
}

// AcknowledgeAlert 确认告警
func (am *AlertManager) AcknowledgeAlert(alertID, acknowledgedBy string) error {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    
    alert, exists := am.alerts[alertID]
    if !exists {
        return fmt.Errorf("alert not found: %s", alertID)
    }
    
    alert.Status = StatusAcknowledged
    alert.AcknowledgedBy = acknowledgedBy
    alert.AcknowledgedAt = time.Now()
    alert.UpdatedAt = time.Now()
    
    return nil
}

// GetAlerts 获取告警列表
func (am *AlertManager) GetAlerts(filters map[string]string) []*Alert {
    am.mutex.RLock()
    defer am.mutex.RUnlock()
    
    var result []*Alert
    for _, alert := range am.alerts {
        if am.matchFilters(alert, filters) {
            result = append(result, alert)
        }
    }
    
    return result
}

// matchFilters 匹配过滤条件
func (am *AlertManager) matchFilters(alert *Alert, filters map[string]string) bool {
    for key, value := range filters {
        switch key {
        case "severity":
            if string(alert.Severity) != value {
                return false
            }
        case "status":
            if string(alert.Status) != value {
                return false
            }
        case "rule_name":
            if alert.RuleName != value {
                return false
            }
        default:
            // 检查标签
            if labelValue, exists := alert.Labels[key]; !exists || labelValue != value {
                return false
            }
        }
    }
    return true
}

// GetStatus 获取告警管理器状态
func (am *AlertManager) GetStatus() string {
    am.mutex.RLock()
    defer am.mutex.RUnlock()
    
    activeAlerts := 0
    for _, alert := range am.alerts {
        if alert.Status == StatusFiring {
            activeAlerts++
        }
    }
    
    if activeAlerts > 0 {
        return fmt.Sprintf("alerting (%d active)", activeAlerts)
    }
    
    return "healthy"
}

// Slack通知渠道实现
type SlackChannel struct {
    enabled   bool
    webhookURL string
    channel   string
    username  string
}

// NewSlackChannel 创建Slack通知渠道
func NewSlackChannel() *SlackChannel {
    return &SlackChannel{
        enabled:   config.GetBool("slack.enabled", true),
        webhookURL: config.GetString("slack.webhook_url", ""),
        channel:   config.GetString("slack.channel", "#alerts"),
        username:  config.GetString("slack.username", "AlertBot"),
    }
}

func (sc *SlackChannel) GetName() string {
    return "slack"
}

func (sc *SlackChannel) IsEnabled() bool {
    return sc.enabled && sc.webhookURL != ""
}

func (sc *SlackChannel) GetConfig() map[string]interface{} {
    return map[string]interface{}{
        "enabled":     sc.enabled,
        "channel":     sc.channel,
        "username":    sc.username,
        "webhook_set": sc.webhookURL != "",
    }
}

func (sc *SlackChannel) Send(alert *Alert) error {
    if !sc.IsEnabled() {
        return fmt.Errorf("slack channel not enabled")
    }
    
    color := sc.getSeverityColor(alert.Severity)
    
    payload := map[string]interface{}{
        "channel":  sc.channel,
        "username": sc.username,
        "attachments": []map[string]interface{}{
            {
                "color": color,
                "title": fmt.Sprintf("[%s] %s", strings.ToUpper(string(alert.Severity)), alert.RuleName),
                "text":  alert.Message,
                "fields": []map[string]interface{}{
                    {
                        "title": "Status",
                        "value": string(alert.Status),
                        "short": true,
                    },
                    {
                        "title": "Value",
                        "value": fmt.Sprintf("%.2f", alert.Value),
                        "short": true,
                    },
                    {
                        "title": "Started At",
                        "value": alert.StartsAt.Format("2006-01-02 15:04:05"),
                        "short": true,
                    },
                },
                "ts": alert.StartsAt.Unix(),
            },
        },
    }
    
    jsonData, _ := json.Marshal(payload)
    
    resp, err := http.Post(sc.webhookURL, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
    }
    
    return nil
}

func (sc *SlackChannel) getSeverityColor(severity AlertSeverity) string {
    switch severity {
    case SeverityCritical:
        return "danger"
    case SeverityHigh:
        return "warning"
    case SeverityMedium:
        return "good"
    case SeverityLow:
        return "#36a64f"
    case SeverityInfo:
        return "#439FE0"
    default:
        return "#808080"
    }
}

// 辅助函数
func copyMap(original map[string]string) map[string]string {
    copied := make(map[string]string)
    for k, v := range original {
        copied[k] = v
    }
    return copied
}

// RuleEvaluator 规则评估器
type RuleEvaluator struct {
    metricsRegistry *MetricsRegistry
}

// NewRuleEvaluator 创建规则评估器
func NewRuleEvaluator() *RuleEvaluator {
    return &RuleEvaluator{
        metricsRegistry: GetGlobalMetricsRegistry(),
    }
}

// EvaluateQuery 评估查询
func (re *RuleEvaluator) EvaluateQuery(query string) (float64, error) {
    // 这里应该实现实际的查询评估逻辑
    // 可以集成Prometheus PromQL或者自定义查询语言
    
    // 示例实现：解析简单的指标查询
    if strings.HasPrefix(query, "metric:") {
        metricName := strings.TrimPrefix(query, "metric:")
        value := re.metricsRegistry.GetGauge(metricName)
        return value, nil
    }
    
    return 0, fmt.Errorf("unsupported query: %s", query)
}
```

---

## 📝 总结

通过本教程，你已经掌握了YYHertz框架的全面监控告警技能：

### 🎯 核心监控能力
- **系统性能监控** - CPU、内存、网络、磁盘全方位监控
- **应用层监控** - HTTP请求、数据库查询、缓存性能监控  
- **业务指标监控** - 自定义指标收集和分析
- **日志监控** - 结构化日志收集和分析
- **告警系统** - 智能告警规则和多渠道通知
- **可视化dashboard** - 实时监控面板和报表

### 💡 监控最佳实践
- **四个黄金信号** - 延迟、流量、错误、饱和度
- **分层监控** - 基础设施、应用、业务三层监控
- **主动监控** - 预警机制和异常检测
- **可观测性** - Metrics、Logs、Traces三支柱
- **告警优化** - 合理的告警阈值和通知策略

### 🚀 进阶监控方向  
- **APM集成** - Application Performance Monitoring
- **分布式追踪** - 微服务链路追踪
- **智能告警** - 基于ML的异常检测
- **SRE实践** - 服务可靠性工程

---

<div align="center">

**📊 构建YYHertz监控告警体系，保障系统稳定运行！**

**从基础监控到智能运维，全方位可观测性解决方案！📈**

</div>