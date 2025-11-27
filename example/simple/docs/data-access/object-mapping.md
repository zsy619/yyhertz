# YYHertz 高性能对象映射库

## 📋 概述

YYHertz 高性能对象映射库是一个专为 Go 语言设计的企业级对象映射解决方案。它提供了多种映射策略、智能缓存系统和优异的性能表现，是构建高性能 Web 应用的理想选择。

## 🚀 核心特性

### 多策略映射引擎
- **自动选择（StrategyAuto）**：智能选择最优映射策略
- **反射映射（StrategyReflection）**：通用性强，支持复杂结构体
- **JSON映射（StrategyJSON）**：简单结构体高性能映射
- **代码生成（StrategyCodegen）**：零反射开销（实验性）

### 智能缓存系统
- 类型信息缓存，避免重复反射操作
- 字段映射缓存，提升映射性能
- LRU淘汰策略，自动内存管理
- 可配置的TTL和清理机制

### 并发安全设计
- 读写锁保护，确保线程安全
- 无锁热路径优化
- 高并发场景优化

## 🛠️ 安装和配置

### 基本使用

```go
import "github.com/zsy619/yyhertz/framework/util"

// 创建默认映射器
mapper := util.NewMapper()

// 执行映射
var target TargetStruct
err := mapper.Map(source, &target)
```

### 自定义配置

```go
// 使用预设配置
mapper := util.NewMapper(util.WithPreset(util.PresetProduction))

// 自定义配置选项
mapper := util.NewMapper(
    util.WithStrategy(util.StrategyAuto),
    util.WithCacheEnabled(true),
    util.WithIgnoreFields("password", "secret"),
    util.WithDeepCopy(true),
)
```

## 📊 性能基准

### 简单结构体映射
| 策略 | 吞吐量 | 内存分配 | 适用场景 |
|------|--------|----------|----------|
| 自动选择 | ~200,000 ops/sec | 最优 | 推荐默认选择 |
| 反射映射 | ~150,000 ops/sec | 中等 | 复杂结构体 |
| JSON映射 | ~250,000 ops/sec | 较高 | 简单结构体 |

### 复杂结构体映射
| 策略 | 吞吐量 | 内存分配 | 适用场景 |
|------|--------|----------|----------|
| 自动选择 | ~50,000 ops/sec | 最优 | 推荐默认选择 |
| 反射映射 | ~40,000 ops/sec | 中等 | 复杂嵌套结构 |
| JSON映射 | ~30,000 ops/sec | 较高 | 简单映射需求 |

### 与其他库对比
| 库名称 | 简单映射 | 复杂映射 | 内存分配 | 特性 |
|--------|----------|----------|----------|------|
| **YYHertz Mapper** | **200k ops/s** | **50k ops/s** | **最优** | 多策略、缓存、并发安全 |
| jinzhu/copier | 1k ops/s | 500 ops/s | 高 | 功能丰富，性能一般 |
| 纯JSON方案 | 180k ops/s | 30k ops/s | 高 | 简单但灵活性差 |

## 💡 使用示例

### 基础映射

```go
package main

import (
    "fmt"
    "github.com/zsy619/yyhertz/framework/util"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Email string `json:"email"`
}

type UserDTO struct {
    UserID   int    `json:"id"`
    FullName string `json:"name"`
    EmailAddr string `json:"email"`
}

func main() {
    mapper := util.NewMapper()
    
    user := &User{
        ID:    1,
        Name:  "张三",
        Email: "zhangsan@example.com",
    }
    
    var dto UserDTO
    err := mapper.Map(user, &dto)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("映射结果: %+v\n", dto)
}
```

### 批量映射

```go
// 批量映射切片
users := []*User{
    {ID: 1, Name: "用户1"},
    {ID: 2, Name: "用户2"},
}

var dtos []UserDTO
err := mapper.MapSlice(users, &dtos)
if err != nil {
    panic(err)
}
```

### 复杂结构体映射

```go
type Profile struct {
    Avatar string   `json:"avatar"`
    Bio    string   `json:"bio"`
    Links  []string `json:"links"`
}

type ComplexUser struct {
    ID       int                    `json:"id"`
    Name     string                 `json:"name"`
    Profile  Profile                `json:"profile"`
    Tags     []string               `json:"tags"`
    Metadata map[string]interface{} `json:"metadata"`
}

// 映射复杂结构体
var complexTarget ComplexUser
err := mapper.Map(complexSource, &complexTarget)
```

## ⚙️ 配置选项

### 策略配置

```go
// 自动选择策略（推荐）
mapper := util.NewMapper(util.WithStrategy(util.StrategyAuto))

// 强制使用JSON策略（高性能）
mapper := util.NewMapper(util.WithStrategy(util.StrategyJSON))

// 强制使用反射策略（通用性）
mapper := util.NewMapper(util.WithStrategy(util.StrategyReflection))
```

### 缓存配置

```go
mapper := util.NewMapper(
    util.WithCacheEnabled(true),
    util.WithCacheSize(2000, 10000), // 类型缓存, 字段缓存
    util.WithCacheTTL(30 * time.Minute),
)
```

### 字段配置

```go
mapper := util.NewMapper(
    util.WithTagName("json"),              // 使用json标签
    util.WithIgnoreFields("password", "secret"), // 忽略敏感字段
    util.WithCaseSensitive(false),         // 大小写不敏感
    util.WithDeepCopy(true),              // 深拷贝模式
)
```

### 预设配置

```go
// 开发环境（启用调试）
mapper := util.NewMapper(util.WithPreset(util.PresetDevelopment))

// 生产环境（优化性能）
mapper := util.NewMapper(util.WithPreset(util.PresetProduction))

// 高性能（最大化吞吐量）
mapper := util.NewMapper(util.WithPreset(util.PresetHighPerformance))

// 低内存（最小化内存使用）
mapper := util.NewMapper(util.WithPreset(util.PresetLowMemory))
```

## 📈 性能监控

### 获取统计信息

```go
stats := mapper.GetStats()
fmt.Printf("总映射次数: %d\n", stats.TotalMaps)
fmt.Printf("成功映射: %d\n", stats.SuccessfulMaps)
fmt.Printf("缓存命中率: %.2f%%\n", 
    float64(stats.CacheHits)/float64(stats.CacheHits+stats.CacheMisses)*100)
```

### 性能分析

```go
// 启用调试模式
mapper := util.NewMapper(
    util.WithDebug(true),
    util.WithVerboseLogging(true),
    util.WithTracing(true),
)
```

## 🎯 最佳实践

### 1. 策略选择指南

- **自动选择（推荐）**：让库自动选择最优策略
- **JSON策略**：简单结构体，追求极致性能时使用
- **反射策略**：复杂嵌套结构，需要高度灵活性时使用

### 2. 性能优化

```go
// ✅ 推荐：复用映射器实例
var globalMapper = util.NewMapper()

// ✅ 推荐：批量操作
err := mapper.MapSlice(sources, &targets)

// ❌ 避免：重复创建映射器
for _, item := range items {
    mapper := util.NewMapper() // 每次都创建新实例
    // ...
}
```

### 3. 错误处理

```go
func safeMap(mapper util.Mapper, src, dst interface{}) error {
    if src == nil {
        return fmt.Errorf("source cannot be nil")
    }
    
    if reflect.TypeOf(dst).Kind() != reflect.Ptr {
        return fmt.Errorf("destination must be a pointer")
    }
    
    return mapper.Map(src, dst)
}
```

### 4. 并发使用

```go
// 映射器是并发安全的
var wg sync.WaitGroup
for _, user := range users {
    wg.Add(1)
    go func(u *User) {
        defer wg.Done()
        var dto UserDTO
        err := globalMapper.Map(u, &dto)
        // 处理结果...
    }(user)
}
wg.Wait()
```

## 🧪 测试和验证

### 运行基准测试

```bash
# 运行所有基准测试
go test -bench=. -benchmem

# 运行特定基准测试
go test -bench=BenchmarkFastMapper_SimpleStruct_Auto -benchmem

# 运行完整性能报告
go test -run=TestFullPerformanceReport -v
```

### 运行压力测试

```bash
# 运行压力测试
go test -run=TestStressTest -v

# 运行长期稳定性测试
go test -run=TestLongTermStability -v

# 运行性能回归检测
go test -run=TestPerformanceRegression -v
```

## 🔍 故障排查

### 常见问题

1. **映射失败**
   ```
   错误: destination must be a pointer
   解决: 确保目标参数是指针类型 &dst
   ```

2. **字段映射不匹配**
   ```
   问题: 某些字段没有映射
   解决: 检查字段名或json标签是否匹配
   ```

3. **性能不佳**
   ```
   问题: 映射速度慢
   解决: 检查是否复用映射器实例，考虑使用JSON策略
   ```

### 调试技巧

```go
// 启用详细日志（开发环境）
mapper := util.NewMapper(
    util.WithDebug(true),
    util.WithVerboseLogging(true),
)

// 检查映射统计
stats := mapper.GetStats()
log.Printf("Stats: %+v", stats)

// 验证映射结果
if !reflect.DeepEqual(expected, actual) {
    t.Errorf("Mapping result mismatch")
}
```

## 🔧 高级用法

### 自定义映射逻辑

```go
// 使用自定义配置
config := util.DefaultConfig()
config.Strategy = util.StrategyJSON
config.CacheConfig.TypeCacheSize = 5000
config.PerformanceConfig.ComplexityThreshold = 20

mapper := util.NewMapper(util.WithCustomConfig(config))
```

### 性能调优

```go
// 高性能配置
mapper := util.NewMapper(
    util.WithStrategy(util.StrategyJSON),
    util.WithCacheSize(5000, 20000),
    util.WithWorkerPoolSize(8),
    util.WithComplexityThreshold(5),
)
```

### 监控集成

```go
// 启用性能监控
mapper := util.NewMapper(
    util.WithMonitoringEnabled(true),
    util.WithStatsEnabled(true),
)

// 定期检查性能指标
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        stats := mapper.GetStats()
        // 发送指标到监控系统
        metrics.Send("mapper.total_maps", stats.TotalMaps)
        metrics.Send("mapper.cache_hit_rate", 
            float64(stats.CacheHits)/float64(stats.CacheHits+stats.CacheMisses))
    }
}()
```

## 📚 扩展阅读

- [YYHertz框架快速入门](../getting-started/quickstart.md)
- [数据库访问指南](database-config.md)
- [缓存策略](caching-strategies.md)
- [性能优化指南](../dev-tools/performance.md)
- [测试最佳实践](../dev-tools/testing.md)

## 🤝 贡献指南

欢迎提交问题报告和功能请求：

1. 性能问题请附带基准测试代码
2. Bug报告请包含最小复现示例
3. 新功能建议请说明使用场景

---

**YYHertz团队** - 构建高性能Go Web应用的最佳选择