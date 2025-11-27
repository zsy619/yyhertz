# YYHertz DevTools 开发工具集

YYHertz DevTools 是一套完整的Web开发调试和监控工具集，为基于CloudWego Hertz的YYHertz框架提供全方位的开发支持。

## 🚀 功能特性

### 1. 调试中间件 (Debug Middleware)
- **请求追踪**: 详细记录每个HTTP请求的完整生命周期
- **错误分析**: 自动捕获和分析运行时错误
- **性能监控**: 实时监控请求响应时间和内存使用
- **堆栈跟踪**: 出错时自动生成详细的调用堆栈
- **可视化面板**: 提供友好的Web界面查看调试信息

### 2. 健康检查 (Health Check)
- **多维度检查**: 支持启动、就绪、存活三种检查类型
- **系统资源监控**: CPU、内存、协程数、GC统计
- **数据库检查**: MySQL、PostgreSQL等数据库连接检查
- **缓存检查**: Redis连接和响应检查
- **外部服务检查**: HTTP服务连通性检查
- **磁盘空间检查**: 存储空间使用率监控
- **Kubernetes集成**: 兼容K8s探针标准

### 3. 统一监控系统 (Integrated Metrics System)
- **性能监控**: 实时统计请求数、错误数、响应时间
- **端点分析**: 按API端点分组的性能指标
- **运行时监控**: 详细的GC分析、内存分配、协程管理
- **健康评估**: 智能评分、优化建议、趋势分析
- **内存分析**: 堆内存、栈内存、GC开销详细统计
- **系统监控**: CPU使用率、内存详情、系统资源

### 4. QPS监控 (QPS Monitor)
- **实时QPS**: 基于滑动窗口的精确QPS计算
- **并发监控**: 实时并发请求数统计
- **端点QPS**: 按API端点分组的QPS统计
- **告警机制**: 可配置的QPS和并发告警阈值
- **历史数据**: QPS历史趋势图表展示

### 5. 企业级指标收集 (Enterprise Metrics Collector)
- **Prometheus兼容**: 标准Prometheus格式指标输出
- **多种指标类型**: Counter、Gauge、Histogram、Summary
- **自动标签**: 自动为HTTP请求添加method、path、status标签
- **系统指标**: 协程数、内存使用、运行时间等系统指标
- **自定义指标**: 支持注册业务自定义指标
- **集成监控**: 包含性能监控、运行时分析、健康评估

### 6. 性能分析 (Profiler)
- **CPU分析**: 集成Go原生CPU profiling
- **内存分析**: 堆内存分配和泄漏检测
- **协程分析**: Goroutine数量和状态分析
- **阻塞分析**: 同步原语阻塞分析
- **互斥锁分析**: 锁竞争和等待分析
- **pprof集成**: 完整集成Go pprof工具链
- **文件下载**: 支持分析文件下载和离线分析

### 7. 限流中间件 (Rate Limiter)
- **多种算法**: 令牌桶、滑动窗口、漏桶算法
- **多维限流**: IP、用户、API、全局四个维度
- **动态配置**: 运行时动态调整限流规则
- **白名单机制**: 支持IP白名单绕过限流
- **实时统计**: 限流效果实时统计和监控

### 8. 热重载 (Hot Reload)
- **文件监控**: 自动监控Go、HTML、CSS、JS文件变化
- **智能过滤**: 排除临时文件和不相关目录
- **防抖处理**: 避免频繁文件变化导致的重复重载
- **自定义回调**: 支持自定义重载逻辑

## 📦 快速开始

### 基础用法

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/devtools"
)

func main() {
    app := mvc.NewApp()
    
    // 启用基础开发工具
    devtools.SetupBasicDevTools(app)
    
    app.Run(":8080")
}
```

### 完整配置

```go
package main

import (
    "database/sql"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/devtools"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    app := mvc.NewApp()
    
    // 数据库连接
    db, _ := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
    
    // 完整配置
    config := &devtools.DevToolsConfig{
        Enabled:           true,
        Environment:       "development",
        EnableDebug:       true,
        EnableHealthCheck: true,
        // 性能监控功能已集成到EnableMetrics中
        EnableQPS:         true,
        EnableMetrics:     true,
        EnableProfiler:    true,
        EnableRateLimit:   true,
        EnableHotReload:   true,
        Database:          db,
        RedisAddr:         "localhost:6379",
        ExternalServices:  []string{"http://api.example.com/health"},
        RateLimitRules: []*devtools.RateLimit{
            {
                Rate:      100,
                Burst:     200,
                Strategy:  devtools.LimitStrategyTokenBucket,
                Dimension: devtools.LimitDimensionGlobal,
                Enabled:   true,
            },
        },
        WhitelistIPs: []string{"127.0.0.1", "192.168.1.100"},
    }
    
    devtools.SetupDevTools(app, config)
    
    app.Run(":8080")
}
```

### 生产环境用法

```go
func main() {
    app := mvc.NewApp()
    
    db, _ := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
    
    // 生产环境配置（不包含调试和热重载）
    devtools.SetupProductionDevTools(app, db, "localhost:6379")
    
    app.Run(":8080")
}
```

## 🌐 访问地址

启动应用后，可以通过以下地址访问各种工具：

**访问地址**:
- 调试面板: `http://localhost/yyhertz/debug/panel`
- 健康检查面板: `http://localhost/yyhertz/health/panel`
- **统一监控面板**: `http://localhost/yyhertz/metrics/panel`
  - 性能监控: `http://localhost/yyhertz/metrics/performance`
  - 运行时监控: `http://localhost/yyhertz/metrics/runtime`
  - 端点统计: `http://localhost/yyhertz/metrics/performance/endpoints`
- QPS监控面板: `http://localhost/yyhertz/qps/panel`
- 性能分析面板: `http://localhost/yyhertz/profile/panel`
- 限流管理面板: `http://localhost/yyhertz/ratelimit/panel`
- 数据库监控: `http://localhost/yyhertz/database/panel`
- 缓存监控: `http://localhost/yyhertz/cache/panel`
- 安全监控: `http://localhost/yyhertz/security/panel`

**K8s健康检查端点**:
- 存活检查: `http://localhost/yyhertz/health/live`
- 就绪检查: `http://localhost/yyhertz/health/ready`
- 启动检查: `http://localhost/yyhertz/health/startup`

**Prometheus指标端点**:
- Prometheus格式: `http://localhost/yyhertz/metrics/prometheus`
- JSON格式: `http://localhost/yyhertz/metrics/json`

**pprof性能分析**:
- pprof界面: `http://localhost/yyhertz/profile/debug/pprof/`

## 🔧 高级用法

### 自定义健康检查器

```go
// 创建自定义检查器
customChecker := devtools.NewCustomChecker("custom_service", func(ctx context.Context) devtools.HealthCheckResult {
    // 执行自定义检查逻辑
    return devtools.HealthCheckResult{
        Name:      "custom_service",
        Status:    devtools.HealthStatusHealthy,
        Message:   "服务运行正常",
        Timestamp: time.Now(),
        Details:   map[string]interface{}{"version": "1.0.0"},
    }
})

// 添加到健康检查中间件
healthMiddleware.AddChecker(customChecker)
```

### 自定义指标

```go
// 创建自定义计数器
requestCounter := devtools.NewCounter("my_requests_total", "Total requests", map[string]string{
    "service": "user_service",
})

// 注册指标
metricsCollector.RegisterMetric(requestCounter)

// 在业务代码中使用
requestCounter.Inc()
```

### 动态限流规则

```go
// 添加IP限流规则
ipLimit := &devtools.RateLimit{
    Rate:      50,
    Burst:     100,
    Strategy:  devtools.LimitStrategyTokenBucket,
    Dimension: devtools.LimitDimensionIP,
    Enabled:   true,
}
rateLimiter.AddRule(ipLimit)

// 添加IP到白名单
rateLimiter.AddToWhitelist("192.168.1.100")
```

## 📊 监控面板特性

所有工具都提供了美观、响应式的Web管理界面：

- **实时数据更新**: 自动刷新最新监控数据
- **交互式图表**: 基于Chart.js的动态图表
- **响应式设计**: 支持桌面和移动设备
- **暗色主题**: 适合长时间使用的护眼设计

## 🛡️ 安全考虑

1. **生产环境**: 建议在生产环境中禁用调试中间件
2. **访问控制**: 监控面板建议配置访问控制
3. **数据敏感性**: 调试信息可能包含敏感数据，请妥善保护
4. **性能影响**: 虽然工具经过优化，但仍建议在高负载场景下谨慎使用

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个工具集！

## 📄 许可证

MIT License

## 🙏 致谢

- CloudWego Hertz: 底层HTTP框架
- Chart.js: 图表展示
- Go pprof: 性能分析
- Prometheus: 指标规范

---

> 💡 **提示**: 更多详细文档和使用示例，请参考各个模块的源码注释。
