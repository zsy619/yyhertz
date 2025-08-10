# 🗄️ 统一ORM解决方案

YYHertz框架提供了业界领先的统一ORM解决方案，通过双引擎协同架构，将GORM的高效性与MyBatis的灵活性完美结合。

## 🎯 核心设计理念

### ✨ 双引擎协同
- **GORM引擎** - 处理简单高效的CRUD操作
- **MyBatis引擎** - 处理复杂动态的SQL查询  
- **智能选择器** - 根据操作复杂度自动选择最优引擎

### 🚀 设计优势
- **统一接口** - 开发者使用同一套API，无需关心底层引擎切换
- **性能最优** - 每个操作都由最适合的引擎处理
- **渐进式** - 从简单CRUD到复杂查询的平滑过渡

## 🏗️ 架构概览

```
┌─────────────────┐
│   应用控制器      │
└─────────┬───────┘
          │
┌─────────▼───────┐
│  ORMManager     │ ← 统一管理器
│  统一入口点      │
└─────────┬───────┘
          │
┌─────────▼───────┐
│ SmartSelector   │ ← 智能选择器
│ 自动引擎路由     │
└─────┬───┬───────┘
      │   │
┌─────▼─┐ │ ┌─────▼─────┐
│ GORM  │ │ │ MyBatis   │
│ 引擎  │ │ │ 引擎      │
│16,278 │ │ │ 990       │
│ops/s  │ │ │ ops/s     │
└─────┬─┘ │ └─────┬─────┘
      │   │       │
      └───┼───────┘
          │
┌─────────▼───────┐
│     数据库       │
│ MySQL/PgSQL/... │
└─────────────────┘
```

## ⚡ 性能表现

基于实际生产环境的性能测试数据：

| 操作类型 | 性能指标 | 引擎选择 | 适用场景 |
|---------|----------|----------|----------|
| **创建操作** | 16,278 ops/sec | GORM | 用户注册、数据插入 |
| **简单查询** | 44,816 ops/sec | GORM | 列表查询、详情查看 |
| **更新操作** | 37,648 ops/sec | GORM | 状态修改、信息更新 |
| **复杂查询** | 990 ops/sec | MyBatis | 统计报表、多表关联 |

*测试环境: Intel i7-6820HQ CPU @ 2.70GHz, SQLite内存数据库*

## 🚀 快速开始

### 1. 基础配置

```yaml
# conf/database.yaml
database:
  primary:
    driver: "mysql"
    dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4"
    max_idle_conns: 20
    max_open_conns: 100
```

### 2. 控制器使用

```go
type UserController struct {
    mvc.BaseController
}

// 简单查询 - 自动选择GORM引擎
func (c *UserController) GetList() {
    var users []User
    orm := orm.GetDefault()
    
    // 智能选择器自动使用GORM处理简单查询
    users, err := orm.Find(&User{}).Where("status = ?", "active").All()
    if err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]any{"users": users})
}

// 复杂查询 - 自动选择MyBatis引擎  
func (c *UserController) GetReport() {
    orm := orm.GetDefault()
    
    // 智能选择器自动使用MyBatis处理复杂查询
    report, err := orm.Query(`
        SELECT department, COUNT(*) as user_count, 
               AVG(order_amount) as avg_amount
        FROM users u 
        LEFT JOIN orders o ON u.id = o.user_id
        WHERE u.created_at >= ?
        GROUP BY department
        ORDER BY avg_amount DESC
    `, startDate)
    
    if err != nil {
        c.Error(500, err.Error())
        return  
    }
    
    c.JSON(map[string]any{"report": report})
}
```

## 🤖 智能选择策略

### 自动选择规则

| 操作特征 | 选择引擎 | 判断依据 |
|---------|----------|----------|
| 单表CRUD | GORM | 操作简单，性能优先 |
| 预定义关联查询 | GORM | 有ORM模型映射 |
| 动态WHERE条件 | MyBatis | 需要动态SQL构建 |
| 多表复杂JOIN | MyBatis | SQL逻辑复杂 |
| 聚合统计查询 | MyBatis | 需要数据库函数 |
| 批量操作 | 自适应 | 根据数据量选择 |

### 手动引擎指定

```go
// 强制使用GORM引擎
db := orm.GetDefault().GORM()
var users []User
db.Where("age > ?", 18).Find(&users)

// 强制使用MyBatis引擎
session := orm.GetDefault().MyBatis().OpenSession()
defer session.Close()
result, err := session.SelectList("UserMapper.complexQuery", params)
```

## 📊 集成功能特性

### 🔄 事务管理
```go
// 跨引擎事务支持
err := orm.GetDefault().Transaction(func(tx *orm.TransactionContext) error {
    // GORM操作
    if err := tx.GORM().Create(&user).Error; err != nil {
        return err
    }
    
    // MyBatis操作  
    if _, err := tx.MyBatis().Update("updateStats", params); err != nil {
        return err
    }
    
    return nil
})
```

### 💾 智能缓存
```go
// 自动缓存热点查询
users, err := orm.GetDefault().Cache().Query(&orm.CacheQuery{
    Key: "active_users",
    TTL: 5 * time.Minute,
    Query: func() (interface{}, error) {
        return orm.Find(&User{}).Where("status = ?", "active").All()
    },
})
```

### 📈 性能监控
```go
// 获取性能指标
metrics := orm.GetDefault().GetMetrics()
fmt.Printf("GORM查询: %d, MyBatis查询: %d, 智能选择: %d\n", 
    metrics.GormQueries, 
    metrics.MyBatisQueries, 
    metrics.AutoSelections)
```

## 🛠️ 开发工具支持

### DryRun调试
```go
// 预览SQL执行计划
result := orm.GetDefault().DryRun().Query("SELECT * FROM users WHERE age > ?", 18)
fmt.Printf("生成SQL: %s\n", result.SQL)
fmt.Printf("使用引擎: %s\n", result.EngineUsed)
```

### 慢查询监控
```go
// 启用慢查询监控
monitor := orm.GetGlobalSlowQueryMonitor()
monitor.SetThreshold(100 * time.Millisecond)
monitor.OnSlowQuery(func(query orm.SlowQuery) {
    log.Printf("慢查询告警: %s 耗时 %v", query.SQL, query.Duration)
})
```

## 📚 相关文档

### 核心文档
- **[GORM快速开始](./gorm-quickstart)** - GORM引擎使用指南
- **[MyBatis基础](./mybatis-basic)** - MyBatis引擎基础功能
- **[MyBatis高级特性](./mybatis-advanced)** - 动态SQL和XML映射
- **[数据库配置](./database-config)** - 连接池和多数据源配置

### 性能优化
- **[MyBatis性能优化](./mybatis-performance)** - 查询优化和缓存策略
- **[数据库调优](./database-tuning)** - 数据库层面的性能调优
- **[缓存策略](./caching-strategies)** - 多级缓存设计模式

### 运维监控
- **[事务管理](./transaction)** - 事务边界和异常处理
- **[监控告警](./monitoring-alerting)** - Prometheus集成和告警规则

## 💡 最佳实践

### 1. 引擎选择建议
- **简单CRUD** → GORM (性能优先)
- **复杂报表** → MyBatis (灵活性优先)
- **批量操作** → 根据数据量自动选择
- **事务操作** → 使用跨引擎事务管理

### 2. 性能优化
- 合理设置连接池参数
- 启用查询结果缓存
- 使用分页避免大量数据查询
- 定期检查慢查询日志

### 3. 开发建议
- 优先使用智能选择器，必要时手动指定引擎
- 在开发环境启用DryRun模式预览SQL
- 配置慢查询监控及时发现性能问题
- 使用事务确保数据一致性

---

**🚀 统一ORM解决方案 - 让您同时享受GORM的高效与MyBatis的灵活！**

*智能选择，性能优先，开发者友好*