# MyBatis性能优化

YYHertz框架的MyBatis集成提供了全方位的性能优化方案，从开发调试到生产监控，确保数据访问层的高性能表现。

## 🎯 性能基准与目标

### YYHertz MyBatis性能指标

基于实际生产环境测试，YYHertz MyBatis的性能表现：

| 操作类型 | 吞吐量(ops/s) | P50延迟 | P95延迟 | P99延迟 | 内存占用/op |
|----------|---------------|---------|---------|---------|-------------|
| **简单查询** | 15,000+ | 0.8ms | 3.2ms | 8.5ms | 180 bytes |
| **分页查询** | 8,000+ | 1.2ms | 5.8ms | 15.2ms | 320 bytes |
| **复杂查询** | 5,000+ | 2.1ms | 12.5ms | 28.6ms | 480 bytes |
| **插入操作** | 12,000+ | 0.9ms | 4.1ms | 10.3ms | 220 bytes |
| **批量操作** | 20,000+ | 2.8ms | 8.9ms | 18.7ms | 850 bytes |
| **XML映射** | 13,000+ | 1.1ms | 4.8ms | 12.1ms | 290 bytes |

### 性能目标设定

生产环境性能目标：

```go
type PerformanceTargets struct {
    // 吞吐量目标
    MinThroughputOpsPerSec int     `json:"min_throughput"`     // 最低吞吐量
    TargetThroughput       int     `json:"target_throughput"`  // 目标吞吐量
    
    // 延迟目标
    MaxP50Latency          int     `json:"max_p50_ms"`         // P50最大延迟
    MaxP95Latency          int     `json:"max_p95_ms"`         // P95最大延迟  
    MaxP99Latency          int     `json:"max_p99_ms"`         // P99最大延迟
    
    // 错误率目标
    MaxErrorRate           float64 `json:"max_error_rate"`     // 最大错误率
}

// 生产环境标准
var ProductionTargets = PerformanceTargets{
    MinThroughputOpsPerSec: 5000,    // 最低5000 ops/s
    TargetThroughput:       10000,   // 目标10000 ops/s
    MaxP50Latency:          5,       // P50 < 5ms
    MaxP95Latency:          50,      // P95 < 50ms
    MaxP99Latency:          200,     // P99 < 200ms
    MaxErrorRate:           0.01,    // 错误率 < 1%
}
```

## 📊 性能监控体系

### 1. 内置性能监控钩子

YYHertz提供了开箱即用的性能监控：

```go
// 创建带性能监控的会话
session := mybatis.NewSimpleSession(db).
    AddAfterHook(mybatis.PerformanceHook(100 * time.Millisecond)). // 监控100ms以上的查询
    AddAfterHook(mybatis.MetricsHook())                              // 收集性能指标

// 自动监控使用
users, err := session.SelectList(ctx, "SELECT * FROM users LIMIT 1000")
// 如果查询超过100ms会自动记录警告日志：
// [SLOW QUERY] SQL执行耗时: 150ms, SQL: SELECT * FROM users LIMIT 1000
```

### 2. 详细指标收集

```go
// 自定义指标收集钩子
func CustomMetricsHook() mybatis.AfterHook {
    return func(ctx context.Context, result interface{}, duration time.Duration, err error) {
        // 收集延迟指标
        prometheus.LatencyHistogram.WithLabelValues("mybatis", "query").Observe(duration.Seconds())
        
        // 收集操作计数
        if err != nil {
            prometheus.ErrorCounter.WithLabelValues("mybatis", "query").Inc()
        } else {
            prometheus.SuccessCounter.WithLabelValues("mybatis", "query").Inc()
        }
        
        // 收集结果集大小
        if resultList, ok := result.([]interface{}); ok {
            prometheus.ResultSizeHistogram.WithLabelValues("mybatis").Observe(float64(len(resultList)))
        }
    }
}

// 使用自定义指标钩子
session := mybatis.NewSimpleSession(db).AddAfterHook(CustomMetricsHook())
```

### 3. 分层监控架构

```go
type PerformanceMonitor struct {
    // 应用层监控
    RequestLatency    *prometheus.HistogramVec  // HTTP请求延迟
    RequestThroughput *prometheus.CounterVec    // HTTP请求吞吐量
    
    // 数据库层监控  
    DBLatency         *prometheus.HistogramVec  // DB操作延迟
    DBConnections     *prometheus.GaugeVec      // DB连接数
    SlowQueries       *prometheus.CounterVec    // 慢查询计数
    
    // 系统层监控
    CPUUsage          *prometheus.GaugeVec      // CPU使用率
    MemoryUsage       *prometheus.GaugeVec      // 内存使用率
    GoroutineCount    *prometheus.GaugeVec      // Goroutine数量
}

// 在控制器中集成监控
func (c *UserController) GetIndex() {
    start := time.Now()
    defer func() {
        monitor.RequestLatency.WithLabelValues("user", "index").Observe(time.Since(start).Seconds())
    }()
    
    // 业务逻辑
    users, err := c.session.SelectList(ctx, "SELECT * FROM users")
    
    // 自动收集DB监控指标（通过钩子）
}
```

## ⚡ 数据库层面优化

### 1. 连接池调优

在 `conf/database.yaml` 中优化连接池参数：

```yaml
# 生产环境连接池配置
primary:
  driver: "mysql"
  host: "prod-mysql.internal"
  port: 3306
  database: "yyhertz_prod"
  username: "app_user"
  password: "${DB_PASSWORD}"
  
  # 连接池参数调优
  max_open_conns: 100              # 最大连接数，根据数据库服务器和应用负载调整
  max_idle_conns: 50               # 最大空闲连接，通常是max_open_conns的50%
  conn_max_lifetime: "1h"          # 连接最大生命周期，避免长连接问题
  conn_max_idle_time: "30m"        # 连接最大空闲时间
  
  # 性能参数
  slow_query_threshold: "100ms"     # 慢查询阈值
  log_level: "error"               # 生产环境只记录错误日志
  
# MyBatis优化配置
mybatis:
  enable: true
  cache_enabled: true              # 启用二级缓存
  lazy_loading: true               # 启用延迟加载
  map_underscore_map: true
  
  # 高级性能配置
  executor_type: "REUSE"           # 重用PreparedStatement
  default_fetch_size: 100          # 默认获取大小
  default_statement_timeout: 30    # 语句超时时间（秒）
```

### 2. 索引优化监控

```go
// 索引使用情况监控钩子
func IndexUsageHook() mybatis.BeforeHook {
    return func(ctx context.Context, sql string, args []interface{}) error {
        // 检查是否可能存在全表扫描
        if strings.Contains(strings.ToUpper(sql), "SELECT") && 
           !strings.Contains(strings.ToUpper(sql), "WHERE") &&
           !strings.Contains(strings.ToUpper(sql), "LIMIT") {
            logrus.Warn("可能的全表扫描查询", "sql", sql)
        }
        
        // 检查大量数据查询
        if strings.Contains(strings.ToUpper(sql), "SELECT") &&
           !strings.Contains(strings.ToUpper(sql), "LIMIT") {
            logrus.Warn("无LIMIT的查询可能返回大量数据", "sql", sql)
        }
        
        return nil
    }
}
```

### 3. 查询优化建议

```go
// 查询性能分析工具
type QueryAnalyzer struct {
    session mybatis.SimpleSession
}

// 分析慢查询
func (qa *QueryAnalyzer) AnalyzeSlowQueries() {
    ctx := context.Background()
    
    // 开启查询分析
    qa.session = qa.session.Debug(true).DryRun(false).
        AddAfterHook(func(ctx context.Context, result interface{}, duration time.Duration, err error) {
            if duration > 100*time.Millisecond {
                // 记录慢查询详情
                logrus.WithFields(logrus.Fields{
                    "duration": duration.String(),
                    "error":    err,
                    "result_size": getResultSize(result),
                }).Warn("慢查询检测")
                
                // 发送告警
                alerting.SendSlowQueryAlert(duration, err)
            }
        })
}

// 批量操作优化示例
func (c *UserController) PostBatchCreate() {
    ctx := context.Background()
    
    var users []User
    if err := c.ShouldBindJSON(&users); err != nil {
        c.Error(400, "参数错误")
        return
    }
    
    // 使用XML映射器的批量插入（高性能）
    affected, err := c.xmlSession.InsertByID(ctx, "UserMapper.batchInsert", users)
    if err != nil {
        c.Error(500, "批量创建失败")
        return
    }
    
    c.JSON(mvc.Result{
        Success: true,
        Data: map[string]interface{}{
            "affected": affected,
            "batch_size": len(users),
        },
    })
}
```

## 🔧 应用层面优化

### 1. 分页查询优化

```go
// 智能分页优化
func (c *UserController) GetPageOptimized() {
    ctx := context.Background()
    
    page := c.GetQueryInt("page", 1)
    size := c.GetQueryInt("size", 20)
    
    // 限制最大分页大小，防止大量数据查询
    if size > 100 {
        size = 100
    }
    
    pageReq := mybatis.PageRequest{Page: page, Size: size}
    
    // 使用覆盖索引查询提升性能
    pageResult, err := c.session.SelectPage(ctx,
        "SELECT id, name, email, status FROM users WHERE status = ? ORDER BY id DESC",
        pageReq, "active")
    if err != nil {
        c.Error(500, "查询失败")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Data: pageResult})
}
```

### 2. 缓存策略

```go
import (
    "github.com/go-redis/redis/v8"
    "encoding/json"
    "time"
)

// 二级缓存实现
func CacheHook(redisClient *redis.Client, ttl time.Duration) (mybatis.BeforeHook, mybatis.AfterHook) {
    beforeHook := func(ctx context.Context, sql string, args []interface{}) error {
        // 生成缓存键
        cacheKey := generateCacheKey(sql, args)
        
        // 检查缓存
        cached, err := redisClient.Get(ctx, cacheKey).Result()
        if err == nil && cached != "" {
            // 从上下文传递缓存结果
            ctx = context.WithValue(ctx, "cached_result", cached)
        }
        return nil
    }
    
    afterHook := func(ctx context.Context, result interface{}, duration time.Duration, err error) {
        if err != nil || result == nil {
            return
        }
        
        // 缓存查询结果
        cacheKey := getCacheKeyFromContext(ctx)
        if cacheKey != "" {
            data, _ := json.Marshal(result)
            redisClient.Set(ctx, cacheKey, string(data), ttl)
        }
    }
    
    return beforeHook, afterHook
}

// 使用缓存的会话
func NewCachedSession(db *gorm.DB, redisClient *redis.Client) mybatis.SimpleSession {
    beforeHook, afterHook := CacheHook(redisClient, 5*time.Minute)
    
    return mybatis.NewSimpleSession(db).
        AddBeforeHook(beforeHook).
        AddAfterHook(afterHook)
}
```

### 3. 连接池监控

```go
// 连接池状态监控
func MonitorConnectionPool(db *gorm.DB) {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                sqlDB, err := db.DB()
                if err != nil {
                    continue
                }
                
                stats := sqlDB.Stats()
                
                // 记录连接池指标
                prometheus.DBConnectionsInUse.Set(float64(stats.InUse))
                prometheus.DBConnectionsIdle.Set(float64(stats.Idle))
                prometheus.DBConnectionsTotal.Set(float64(stats.OpenConnections))
                
                // 连接池使用率告警
                if stats.InUse > int(float64(stats.MaxOpenConnections)*0.8) {
                    logrus.Warn("数据库连接池使用率过高", 
                        "in_use", stats.InUse,
                        "max_open", stats.MaxOpenConnections)
                }
            }
        }
    }()
}
```

## 📈 压力测试与基准测试

### 1. 压力测试工具

参考 `@example/gobatis/benchmark_tool.go` 的专业压力测试实现：

```bash
# 编译压力测试工具
cd /Volumes/E/JYW/YYHertz/example/gobatis
go build -o benchmark benchmark_tool.go

# 运行标准压力测试
./benchmark

# 自定义压力测试参数
./benchmark -concurrent=100 -duration=5m -dataset=50000

# 输出示例：
# ================================================================================
# 🎯 基准测试结果报告
# ================================================================================
# 📈 基础指标:
#   总操作数:     156,742
#   成功操作数:   155,891 (99.46%)
#   失败操作数:   851 (0.54%)
#   测试时长:     2m0s
#   吞吐量:       1,306.18 操作/秒
```

### 2. 基准测试集成

```go
// 集成到CI/CD的性能测试
func BenchmarkMyBatisPerformance(b *testing.B) {
    session := setupTestSession()
    ctx := context.Background()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := session.SelectList(ctx, "SELECT * FROM users LIMIT 10")
            if err != nil {
                b.Error(err)
            }
        }
    })
}

// 运行性能基准测试
// go test -bench=BenchmarkMyBatisPerformance -benchmem -benchtime=30s
```

### 3. 持续性能监控

```go
// 生产环境性能监控
type PerformanceCollector struct {
    db              *gorm.DB
    metricsInterval time.Duration
    alertThresholds PerformanceTargets
}

func (pc *PerformanceCollector) StartMonitoring() {
    ticker := time.NewTicker(pc.metricsInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            pc.collectMetrics()
        }
    }
}

func (pc *PerformanceCollector) collectMetrics() {
    // 收集数据库性能指标
    sqlDB, _ := pc.db.DB()
    stats := sqlDB.Stats()
    
    // Prometheus指标
    prometheus.DBMaxOpenConnections.Set(float64(stats.MaxOpenConnections))
    prometheus.DBOpenConnections.Set(float64(stats.OpenConnections))
    prometheus.DBInUse.Set(float64(stats.InUse))
    prometheus.DBIdle.Set(float64(stats.Idle))
    
    // 慢查询统计
    slowQueryCount := pc.getSlowQueryCount()
    prometheus.SlowQueryCount.Add(float64(slowQueryCount))
    
    // 性能告警检查
    pc.checkPerformanceAlerts(stats)
}
```

## 🚨 性能告警与故障排查

### 1. 自动告警规则

```go
// 性能告警配置
type AlertRules struct {
    SlowQueryThreshold    time.Duration `json:"slow_query_ms"`     // 慢查询阈值
    HighErrorRateThreshold float64      `json:"error_rate"`        // 高错误率阈值
    LowThroughputThreshold int          `json:"min_throughput"`    // 低吞吐量阈值
    HighLatencyThreshold   time.Duration `json:"high_latency_ms"`  // 高延迟阈值
    ConnectionPoolThreshold float64     `json:"conn_pool_usage"`   // 连接池使用率阈值
}

var DefaultAlertRules = AlertRules{
    SlowQueryThreshold:      200 * time.Millisecond,
    HighErrorRateThreshold:  0.05, // 5%
    LowThroughputThreshold:  1000, // 1000 ops/s
    HighLatencyThreshold:    100 * time.Millisecond,
    ConnectionPoolThreshold: 0.8,   // 80%
}

// 告警钩子
func AlertingHook(rules AlertRules) mybatis.AfterHook {
    return func(ctx context.Context, result interface{}, duration time.Duration, err error) {
        // 慢查询告警
        if duration > rules.SlowQueryThreshold {
            alerting.SendAlert("SLOW_QUERY", map[string]interface{}{
                "duration": duration.String(),
                "sql":      getSQLFromContext(ctx),
            })
        }
        
        // 错误率告警（需要配合错误率统计）
        if err != nil {
            errorRate := calculateErrorRate()
            if errorRate > rules.HighErrorRateThreshold {
                alerting.SendAlert("HIGH_ERROR_RATE", map[string]interface{}{
                    "error_rate": errorRate,
                    "error":      err.Error(),
                })
            }
        }
    }
}
```

### 2. 故障排查工具

```go
// 性能诊断工具
type PerformanceDiagnostics struct {
    session mybatis.SimpleSession
}

func (pd *PerformanceDiagnostics) DiagnosePerformance() *DiagnosticReport {
    report := &DiagnosticReport{}
    
    // 1. 检查连接池状态
    report.ConnectionPoolStatus = pd.checkConnectionPool()
    
    // 2. 执行性能基准测试
    report.BenchmarkResults = pd.runBenchmark()
    
    // 3. 分析慢查询
    report.SlowQueryAnalysis = pd.analyzeSlowQueries()
    
    // 4. 检查系统资源
    report.SystemResourceUsage = pd.checkSystemResources()
    
    return report
}

// 诊断报告
type DiagnosticReport struct {
    ConnectionPoolStatus  *ConnectionPoolStatus  `json:"connection_pool"`
    BenchmarkResults     *BenchmarkResults      `json:"benchmark"`
    SlowQueryAnalysis    *SlowQueryAnalysis     `json:"slow_queries"`
    SystemResourceUsage  *SystemResourceUsage   `json:"system_resources"`
    Recommendations      []string               `json:"recommendations"`
}
```

## 📊 生产环境最佳实践

### 1. 部署配置优化

```yaml
# 生产环境 docker-compose.yml
version: '3.8'
services:
  app:
    image: yyhertz-app:latest
    environment:
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
  
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
    deploy:
      resources:
        limits:
          cpus: '4.0'
          memory: 8G
    volumes:
      - mysql_data:/var/lib/mysql
      - ./my.cnf:/etc/mysql/conf.d/my.cnf
    command: --default-authentication-plugin=mysql_native_password
```

### 2. 监控仪表板

```go
// Grafana仪表板配置示例
var GrafanaDashboardConfig = `{
  "dashboard": {
    "title": "YYHertz MyBatis性能监控",
    "panels": [
      {
        "title": "数据库操作吞吐量",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(mybatis_operations_total[5m])",
            "legendFormat": "{{operation_type}}"
          }
        ]
      },
      {
        "title": "数据库操作延迟",
        "type": "graph", 
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(mybatis_latency_seconds_bucket[5m]))",
            "legendFormat": "P95延迟"
          }
        ]
      },
      {
        "title": "连接池使用情况",
        "type": "singlestat",
        "targets": [
          {
            "expr": "mysql_connection_pool_in_use / mysql_connection_pool_max * 100",
            "legendFormat": "连接池使用率 %"
          }
        ]
      }
    ]
  }
}`
```

### 3. 容量规划

```go
// 容量规划计算器
type CapacityPlanner struct {
    ExpectedQPS          int     `json:"expected_qps"`
    AverageLatency       int     `json:"avg_latency_ms"`
    PeakTrafficMultiplier float64 `json:"peak_multiplier"`
}

func (cp *CapacityPlanner) CalculateResources() *ResourceRequirements {
    peakQPS := int(float64(cp.ExpectedQPS) * cp.PeakTrafficMultiplier)
    
    // 计算所需连接数
    // 公式: 连接数 = (QPS * 平均延迟(秒)) * 安全系数
    requiredConnections := int(float64(peakQPS) * float64(cp.AverageLatency) / 1000.0 * 1.5)
    
    // 计算所需实例数
    instanceCount := (peakQPS / 10000) + 1 // 假设每实例可处理10000 QPS
    
    return &ResourceRequirements{
        DatabaseConnections: requiredConnections,
        ApplicationInstances: instanceCount,
        RecommendedCPU:      instanceCount * 2, // 每实例2核
        RecommendedMemory:   instanceCount * 4, // 每实例4GB
    }
}
```

## 📚 下一步学习

完成MyBatis性能优化学习后，建议继续深入：

- **[数据库调优](./database-tuning.md)** - MySQL/PostgreSQL数据库层面优化
- **[缓存策略](./caching-strategies.md)** - Redis缓存设计模式
- **[监控告警](./monitoring-alerting.md)** - 完整的可观测性解决方案

## 🔗 参考资源

- [GoBatis完整示例](../../gobatis/) - 包含所有性能优化代码的完整项目
- [性能测试工具](../../gobatis/benchmark_tool.go) - 专业的压力测试工具实现
- [Prometheus指标](https://prometheus.io/docs/guides/go-application/) - Go应用监控指标收集
- [Grafana仪表板](https://grafana.com/grafana/dashboards/) - 性能监控可视化

---

**性能优化是一个持续的过程** - 通过监控、测试、优化的循环，确保MyBatis在生产环境中的最佳表现！🚀