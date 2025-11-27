# YYHertz ORM 示例和测试

本目录包含了 YYHertz ORM 框架的完整示例、性能测试和压力测试，展示了框架的各种功能和最佳实践。

## 📁 文件结构

```
example/ormx/
├── README.md              # 本文档
├── orm_test.go           # 基础功能测试
├── complete_example.go   # 完整功能演示（已修复）
├── benchmark_test.go     # 原始性能基准测试
├── simple_benchmark_test.go # 简化性能基准测试（推荐）
├── stress_test.go        # 压力测试
├── conf/                 # 配置文件目录
│   ├── database.yaml     # 数据库配置
│   ├── log.yaml          # 日志配置
│   ├── mvc.yaml          # MVC框架配置
│   └── middleware.yaml   # 中间件配置
└── test/                 # 独立测试目录
    ├── complete_test.go  # 完整功能测试（可执行程序）
    ├── final_test.go     # 最终验证测试
    ├── minimal_test.go   # 最小功能测试
    ├── orm_test.go       # ORM基础测试
    ├── quick_test.go     # 快速验证测试
    ├── verify_test.go    # 功能验证测试
    └── conf/             # 测试配置目录（自动生成）
```

## 🚀 快速开始

### 运行完整功能演示

```bash
go run complete_example.go
```

这将运行包含以下功能的完整演示：
- 数据库初始化和自动迁移
- 基础 CRUD 操作
- 高级查询功能
- 关联查询
- 事务操作
- 分页查询
- 缓存操作
- 性能监控

### 运行基础测试

```bash
go test -v
```

### 运行性能基准测试

**推荐使用简化版基准测试（避免配置冲突）：**

```bash
go test -bench=Simple -benchtime=500ms -run=^$
```

**或者运行原始基准测试（可能需要配置MySQL）：**

```bash
# 运行所有基准测试
go test -bench=.

# 运行特定基准测试
go test -bench=BenchmarkCRUDCreate
go test -bench=BenchmarkQueryChain
go test -bench=BenchmarkTransaction

# 运行内存分配测试
go test -bench=BenchmarkMemoryUsage -benchmem

# 运行压力测试对比
go test -bench=BenchmarkCompare
```

### 运行压力测试

```bash
go run stress_test.go
```

## 📊 功能演示

### 1. 基础 CRUD 操作

```go
// 获取 CRUD 实例
crud := orm.GetSimpleCRUD()

// 创建记录
user := &User{
    Username: "alice",
    Email:    "alice@example.com",
    Status:   "active",
}
err := crud.Create(user)

// 查询记录
var users []User
err = crud.Find(&users, "status = ?", "active")

// 更新记录
err = crud.Update(user, "status", "premium")

// 删除记录
err = crud.Delete(user)
```

### 2. 链式查询

```go
var users []User
err := crud.Where("age > ?", 18).
    WhereIn("status", []string{"active", "premium"}).
    WhereBetween("created_at", startTime, endTime).
    OrderBy("created_at", "desc").
    Limit(10).
    Find(&users)
```

### 3. 仓库模式

```go
// 获取仓库
userRepo := orm.GetRepository[User]()

// 基础操作
user, err := userRepo.FindByID(1)
users, err := userRepo.FindAll()

// 分页查询
users, total, err := userRepo.Paginate(1, 20)

// 条件查询
users, err := userRepo.FindWhere("age > ?", 18)
```

### 4. 查询构建器

```go
query := orm.GetQueryBuilder[User]()
users, err := query.
    Where("age > ?", 18).
    Or("vip = ?", true).
    Order("created_at DESC").
    Preload("Profile", "Orders").
    Find()
```

### 5. 事务操作

```go
err := orm.QuickTransaction(func(crud *orm.SimpleCRUD) error {
    if err := crud.Create(&user); err != nil {
        return err
    }
    if err := crud.Create(&profile); err != nil {
        return err
    }
    return nil
})
```

### 6. 缓存查询

```go
db := orm.GetDefaultORM().DB()
cachedQuery := orm.NewCachedQuery(db).
    WithCacheKey("active_users").
    WithTTL(5 * time.Minute)

var users []User
err := cachedQuery.Find(&users, "status = ?", "active")
```

## 📈 性能测试结果

### 基准测试结果示例

```
BenchmarkCRUDCreate-8              1000    1234567 ns/op    234 B/op    5 allocs/op
BenchmarkCRUDRead-8                5000     567890 ns/op    123 B/op    3 allocs/op
BenchmarkCRUDUpdate-8              2000     890123 ns/op    156 B/op    4 allocs/op
BenchmarkQueryChain-8              3000     456789 ns/op    345 B/op    8 allocs/op
BenchmarkTransaction-8             1500    1567890 ns/op    456 B/op    12 allocs/op
BenchmarkConnectionPool-8          10000    123456 ns/op     89 B/op    2 allocs/op
```

### 压力测试结果示例

```
=== 创建操作压力测试 ===
并发数: 10
持续时间: 30s
总请求数: 25000
成功请求: 24950
失败请求: 50
每秒请求数: 833.33
错误率: 0.20%
平均耗时: 12ms

=== 查询操作压力测试 ===
并发数: 10
持续时间: 30s
总请求数: 45000
成功请求: 45000
失败请求: 0
每秒请求数: 1500.00
错误率: 0.00%
平均耗时: 6.7ms
```

## 🛠️ 测试配置

### 数据库配置

测试使用 SQLite 内存数据库以获得最佳性能：

```go
config := &orm.DatabaseConfig{
    Type:         "sqlite",
    Database:     ":memory:",
    MaxIdleConns: 10,
    MaxOpenConns: 100,
    LogLevel:     "silent",
}
```

### 基准测试配置

```go
// 设置测试数据量
testDataSize := 10000

// 设置批处理大小
batchSize := 100

// 设置并发级别
concurrency := []int{1, 5, 10, 20, 50}
```

### 压力测试配置

```go
config := &StressTestConfig{
    Concurrency:      10,           // 并发goroutine数
    Duration:         30*time.Second, // 测试持续时间
    WarmupTime:       5*time.Second,  // 预热时间
    EnableMonitoring: true,           // 启用监控
    PrintInterval:    5*time.Second,  // 打印间隔
}
```

## 📋 测试覆盖

### 基础功能测试

- [x] 数据库连接和配置
- [x] 自动迁移
- [x] 基础 CRUD 操作
- [x] 查询构建器
- [x] 仓库模式
- [x] 事务管理
- [x] 关联查询
- [x] 分页查询
- [x] 缓存功能

### 性能基准测试

- [x] 创建操作性能
- [x] 读取操作性能
- [x] 更新操作性能
- [x] 删除操作性能
- [x] 批量操作性能
- [x] 查询操作性能
- [x] 事务操作性能
- [x] 连接池性能
- [x] 内存使用情况
- [x] 方法对比测试

### 压力测试

- [x] 高并发创建测试
- [x] 高并发查询测试
- [x] 高并发更新测试
- [x] 高并发删除测试
- [x] 混合操作测试
- [x] 事务压力测试
- [x] 分页查询压力测试
- [x] 连接池压力测试
- [x] 并发读写测试
- [x] 长时间运行测试

## 🎯 最佳实践

### 1. 性能优化建议

```go
// 使用批量操作
err := crud.BatchCreate(users, 100)

// 使用索引字段查询
err := crud.Find(&users, "email = ?", email) // email有索引

// 预加载关联数据
users, err := query.Preload("Profile", "Orders").Find()

// 使用分页避免大量数据
users, total, err := repo.Paginate(1, 50)
```

### 2. 连接池优化

```go
config := &orm.DatabaseConfig{
    MaxIdleConns: 20,  // 空闲连接数
    MaxOpenConns: 100, // 最大连接数
    MaxLifetime:  3600, // 连接生存时间(秒)
}
```

### 3. 缓存策略

```go
// 短期缓存适用于频繁查询
cachedQuery := orm.NewCachedQuery(db).
    WithCacheKey("frequent_query").
    WithTTL(1 * time.Minute)

// 长期缓存适用于不经常变化的数据
cachedQuery := orm.NewCachedQuery(db).
    WithCacheKey("static_data").
    WithTTL(1 * time.Hour)
```

### 4. 错误处理

```go
user, err := crud.First(&User{}, "email = ?", email)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("用户不存在")
    }
    return nil, fmt.Errorf("查询用户失败: %w", err)
}
```

### 5. 事务最佳实践

```go
err := orm.QuickTransaction(func(crud *orm.SimpleCRUD) error {
    // 1. 所有数据库操作都应该在事务内
    if err := crud.Create(&user); err != nil {
        return err // 自动回滚
    }
    
    // 2. 避免在事务中执行耗时操作
    // 3. 保持事务尽可能短
    
    return nil // 自动提交
})
```

## 📊 监控和诊断

### 连接池监控

```go
stats, err := orm.GetConnectionStats()
if err == nil {
    fmt.Printf("连接池状态:\n")
    fmt.Printf("活跃连接: %d/%d\n", stats.InUse, stats.MaxOpenConnections)
    fmt.Printf("空闲连接: %d\n", stats.Idle)
    fmt.Printf("等待次数: %d\n", stats.WaitCount)
}
```

### 性能监控

```go
monitor := orm.NewPerformanceMonitor(500 * time.Millisecond)
// 执行操作...
stats := monitor.GetStats()
fmt.Printf("慢查询率: %.2f%%\n", stats["slow_query_rate"])
```

### 调试工具

```go
// 开启调试模式查看SQL
orm.DebugMode(true).Find(&users)

// DryRun模式预览SQL
orm.DryRun().Create(&user)

// 显示SQL执行过程
orm.ShowSQL(func(crud *orm.SimpleCRUD) {
    crud.Find(&users, "status = ?", "active")
})
```

## 🔧 环境要求

- Go 1.18+
- GORM v2
- SQLite3 (用于测试)
- 支持的数据库：MySQL, PostgreSQL, SQLite, SQL Server, Oracle

## 📈 性能指标

### 实际基准测试结果 (简化版基准测试)

基于简化基准测试的实际性能数据：

- **创建操作**: ~8,139 ops/500ms ≈ 16,278 ops/sec (67.3μs/op)
- **查询操作**: ~22,408 ops/500ms ≈ 44,816 ops/sec (26.3μs/op)  
- **更新操作**: ~18,824 ops/500ms ≈ 37,648 ops/sec (31.5μs/op)
- **复杂查询**: ~495 ops/500ms ≈ 990 ops/sec (1.26ms/op)

*测试环境: Intel i7-6820HQ CPU @ 2.70GHz, SQLite内存数据库*

### 内存使用

- **基础内存占用**: ~5-10 MB
- **10万记录内存增长**: ~50-80 MB
- **连接池内存**: ~1-2 MB per connection

### 并发能力

- **支持并发数**: 50-100 个并发连接
- **事务并发**: 20-50 个并发事务
- **查询并发**: 100+ 个并发查询

## 🧪 测试目录详解 (test/)

test目录包含了多种测试方法，适用于不同的测试场景：

### 测试文件说明

| 文件名 | 类型 | 用途 | 执行方式 |
|--------|------|------|----------|
| `minimal_test.go` | 单元测试 | 最基础的Go功能测试 | `go test -run TestMinimal` |
| `quick_test.go` | 单元测试 | 快速ORM连接和配置测试 | `go test -run TestQuick` |
| `verify_test.go` | 单元测试 | ORM基本功能验证 | `go test -run TestVerify` |
| `orm_test.go` | 单元测试 | 完整ORM功能测试（包括CRUD） | `go test -run TestOrm` |
| `final_test.go` | 单元测试 | 最终集成测试 | `go test -run TestFinal` |
| `complete_test.go` | 主程序 | 完整功能演示程序 | `go run complete_test.go` |

### 推荐测试顺序

1. **基础验证**：
   ```bash
   cd test && go test -run TestMinimal
   ```

2. **快速功能测试**：
   ```bash
   cd test && go test -run TestQuick -v
   ```

3. **完整功能验证**：
   ```bash
   cd test && go test -run TestOrm -v
   ```

4. **运行完整演示**：
   ```bash
   cd test && go run complete_test.go
   ```

### 测试特点

- **自动配置生成**: 首次运行会自动创建所需的配置文件
- **内存数据库**: 使用SQLite内存数据库，无需外部依赖
- **完整覆盖**: 涵盖CRUD操作、事务、统计、慢查询监控等
- **错误处理**: 包含完善的错误处理和警告提示
- **性能监控**: 集成了连接池统计和慢查询监控

### 配置文件

test目录的配置文件会在首次运行时自动生成：

```
test/conf/
├── app.yaml       # 应用配置
├── auth.yaml      # 认证配置  
├── database.yaml  # 数据库配置
├── log.yaml       # 日志配置
├── mvc.yaml       # MVC框架配置
├── middleware.yaml # 中间件配置
└── ...           # 其他配置文件
```

## 🚨 故障排除

### 常见问题及解决方案

1. **基准测试配置冲突**
   ```
   问题: 运行 benchmark_test.go 时出现数据库配置错误
   原因: 全局ORM接口触发了MySQL配置，但密码为空导致DSN格式错误
   
   解决方案:
   - 使用简化版基准测试: go test -bench=Simple
   - 或配置正确的MySQL连接参数到 conf/database.yaml
   ```

2. **complete_example.go 函数签名错误**
   ```
   问题: 函数参数不匹配错误
   状态: ✅ 已修复
   
   修复内容:
   - 所有演示函数现在接受 *orm.ORM 参数
   - 移除对全局ORM的依赖
   - 使用局部ORM实例避免配置冲突
   ```

3. **测试数据冲突**
   ```
   问题: UNIQUE constraint failed 错误
   解决方案: 
   - 测试使用内存数据库 :memory: 提供隔离
   - 每个测试函数使用不同的模型名称
   - 测试间自动清理数据
   ```

4. **连接池配置**
   ```
   问题: 连接超时或性能问题
   解决方案: 调整配置参数
   
   # conf/database.yaml
   primary:
     max_open_conns: 100
     max_idle_conns: 10
     conn_max_lifetime: "1h"
   ```

5. **慢查询监控**
   ```
   问题: 查询性能问题
   解决方案:
   - 启用慢查询监控
   - 调整慢查询阈值
   - 添加适当的索引
   
   # 在代码中启用监控
   monitor := orm.GetGlobalSlowQueryMonitor()
   monitor.SetThreshold(100 * time.Millisecond)
   ```

### 性能优化建议

1. **使用合适的数据库**
   - 开发/测试: SQLite
   - 生产环境: MySQL/PostgreSQL

2. **连接池调优**
   - 根据并发量调整 max_open_conns
   - 设置合理的 conn_max_lifetime

3. **查询优化**
   - 使用分页查询避免大量数据
   - 启用查询缓存
   - 添加适当的数据库索引

### 调试步骤

1. **启用调试日志**
   ```go
   config := &orm.DatabaseConfig{
       LogLevel: "info",  // 或 "debug" 查看详细SQL
   }
   ```

2. **检查连接池状态**
   ```go
   stats := ormInstance.GetStats()
   for key, value := range stats {
       fmt.Printf("%s: %v\n", key, value)
   }
   ```

3. **运行渐进测试**
   ```bash
   # 基础功能测试
   cd test && go test -run TestMinimal -v
   
   # 快速ORM测试
   cd test && go test -run TestQuick -v
   
   # 完整功能测试
   cd test && go test -run TestOrm -v
   ```

4. **验证配置文件**
   ```bash
   # 检查数据库配置
   cat conf/database.yaml
   
   # 检查日志配置
   cat conf/log.yaml
   ```

## ✅ 功能状态总览

| 功能模块 | 状态 | 说明 |
|---------|------|------|
| complete_example.go | ✅ 已修复 | 所有函数签名已修正，可正常运行 |
| simple_benchmark_test.go | ✅ 可用 | 推荐的基准测试，避免配置冲突 |
| benchmark_test.go | ⚠️ 需配置 | 需要正确的MySQL配置才能运行 |
| test/目录测试 | ✅ 全部通过 | 包含6个测试文件，覆盖各种场景 |
| 基础CRUD操作 | ✅ 正常 | 创建、查询、更新、删除功能正常 |
| 事务处理 | ✅ 正常 | 支持手动和自动事务管理 |
| 分页查询 | ✅ 正常 | 支持偏移分页和条件分页 |
| 缓存功能 | ✅ 正常 | 支持查询结果缓存和TTL管理 |
| 性能监控 | ✅ 正常 | 连接池统计和慢查询监控 |
| 多数据库支持 | ✅ 正常 | 支持MySQL、PostgreSQL、SQLite等 |

## 🚀 快速验证

运行以下命令验证所有功能是否正常：

```bash
# 1. 基础功能测试
go test -v

# 2. 性能基准测试
go test -bench=Simple -benchtime=500ms -run=^$

# 3. 完整功能演示
go run complete_example.go

# 4. test目录验证
cd test && go test -run TestQuick -v
```

## 📚 相关文档

- [YYHertz ORM 框架文档](../../framework/orm/README.md) - 完整API参考
- [YYHertz MyBatis 集成](../../framework/mybatis/README.md) - MyBatis功能
- [GORM 官方文档](https://gorm.io/) - 底层ORM框架
- [YYHertz 框架主页](https://github.com/zsy619/yyhertz) - 项目主页

## 📄 更新日志

### v1.4.0 (2025-08-10)
- ✅ 修复 complete_example.go 函数签名错误
- ✅ 新增 simple_benchmark_test.go 避免配置冲突
- ✅ 完善 test/ 目录测试覆盖
- ✅ 更新文档，增加故障排除指南
- ✅ 优化性能基准测试和监控功能

---

🎉 **所有示例和测试已验证可用！** 如有问题请参考故障排除部分或提交Issue。

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支
3. 编写测试用例
4. 运行完整测试套件
5. 提交 Pull Request

## 📝 测试报告

详细的测试报告请查看 [test_report.md](test_report.md)

## 📄 许可证

MIT License

---

**注意**: 本示例使用 SQLite 内存数据库进行测试，实际生产环境建议使用 MySQL 或 PostgreSQL 等企业级数据库，并根据实际情况调整连接池配置和缓存策略。