# YYHertz ORM 框架

YYHertz ORM 是一个基于 GORM v2 构建的企业级数据库 ORM 框架，提供了简洁、高效、功能丰富的数据库操作接口。

## ✨ 核心特性

- 🚀 **简化的 CRUD 操作** - 基于 GORM 最佳实践的简化接口
- 🔧 **多数据库支持** - 支持 MySQL、PostgreSQL、SQLite、SQL Server、Oracle 等
- 🏊 **连接池管理** - 高性能的数据库连接池和读写分离
- 📊 **查询构建器** - 流畅的链式查询构建器
- 🎯 **通用仓库模式** - 支持泛型的仓库模式
- 💾 **多层缓存** - 内存缓存、Redis缓存、查询缓存
- 🔒 **事务管理** - 完整的事务支持和嵌套事务
- 📈 **性能监控** - 慢查询监控、连接池监控、指标收集
- 🛡️ **错误处理** - 完善的错误处理和重试机制
- 🔄 **自动迁移** - 数据库结构自动迁移和版本管理

## 🚀 快速开始

### 安装

```go
import "github.com/zsy619/yyhertz/framework/orm"
```

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/zsy619/yyhertz/framework/orm"
)

// 定义模型
type User struct {
    orm.BaseModel
    Name  string `gorm:"size:100" json:"name"`
    Email string `gorm:"size:100;uniqueIndex" json:"email"`
    Age   int    `json:"age"`
}

func main() {
    // 1. 初始化数据库配置
    config := &orm.DatabaseConfig{
        Type:         "mysql",
        Host:         "localhost",
        Port:         3306,
        Username:     "root",
        Password:     "password",
        Database:     "myapp",
        Charset:      "utf8mb4",
        Timezone:     "Asia/Shanghai",
        MaxIdleConns: 10,
        MaxOpenConns: 100,
    }

    // 2. 创建 ORM 实例
    ormInstance, err := orm.NewORM(config)
    if err != nil {
        log.Fatal(err)
    }

    // 3. 自动迁移
    ormInstance.AutoMigrate(&User{})

    // 4. 使用简化的 CRUD 操作
    crud := orm.GetSimpleCRUD()
    
    // 创建用户
    user := &User{Name: "张三", Email: "zhangsan@example.com", Age: 25}
    err = crud.Create(user)
    
    // 查询用户
    var users []User
    err = crud.Find(&users)
    
    // 更新用户
    err = crud.Update(user, "age", 26)
    
    // 删除用户
    err = crud.Delete(user)
}
```

## 📖 API 参考

### SimpleCRUD - 简化的 CRUD 操作

```go
// 获取 SimpleCRUD 实例
crud := orm.GetSimpleCRUD()

// 创建记录
err := crud.Create(&user)

// 批量创建
err := crud.BatchCreate(users, 100) // 每批100条

// 查询记录
var users []User
err := crud.Find(&users, "age > ?", 18)

// 查询第一条
var user User
err := crud.First(&user, "email = ?", "test@example.com")

// 更新记录
err := crud.Update(&user, "age", 26)
err := crud.Updates(&user, map[string]interface{}{"age": 26, "name": "新名字"})

// 删除记录
err := crud.Delete(&user)

// 统计记录
count, err := crud.Count(&User{})

// 检查存在
exists, err := crud.Exists(&User{}, "email = ?", "test@example.com")

// 分页查询
total, err := crud.Paginate(&User{}, 1, 10, &users)
```

### 链式查询构建器

```go
crud := orm.GetSimpleCRUD()

// 链式查询
users, err := crud.
    Where("age > ?", 18).
    WhereIn("status", []string{"active", "pending"}).
    WhereBetween("created_at", startTime, endTime).
    OrderBy("created_at", "desc").
    Limit(10).
    Find(&users)
```

### 通用仓库模式

```go
// 获取用户仓库
userRepo := orm.GetRepository[User]()

// 基本操作
user, err := userRepo.FindByID(1)
users, err := userRepo.FindAll()
users, err := userRepo.FindWhere("age > ?", 18)

// 分页查询
users, total, err := userRepo.Paginate(1, 10)
users, total, err := userRepo.PaginateWhere("status = ?", 1, 10, "active")

// 统计操作
count, err := userRepo.Count()
count, err := userRepo.CountWhere("age > ?", 18)
exists, err := userRepo.Exists(1)

// 更新操作
err = userRepo.UpdateColumns(1, map[string]any{"age": 26})
err = userRepo.UpdateWhere("status = ?", map[string]any{"updated_at": time.Now()}, "pending")
```

### 查询构建器

```go
// 获取查询构建器
query := orm.GetQueryBuilder[User]()

// 复杂查询
users, err := query.
    Where("age > ?", 18).
    Or("vip = ?", true).
    Order("created_at DESC").
    Limit(10).
    Offset(20).
    Preload("Profile", "Orders").
    Find()

// 统计查询
count, err := query.Where("status = ?", "active").Count()

// 分页查询
users, total, err := query.
    Where("age BETWEEN ? AND ?", 18, 65).
    Paginate(1, 20)
```

### 事务操作

```go
// 简单事务
err := orm.QuickTransaction(func(crud *orm.SimpleCRUD) error {
    // 在事务中执行多个操作
    if err := crud.Create(&user); err != nil {
        return err
    }
    if err := crud.Create(&profile); err != nil {
        return err
    }
    return nil
})

// 手动事务管理
crud := orm.GetSimpleCRUD()
err := crud.Transaction(func(tx *orm.SimpleCRUD) error {
    // 事务操作
    return nil
})
```

## 🔧 高级功能

### 1. 多数据库支持

```go
// MySQL
mysqlConfig := &orm.DatabaseConfig{
    Type:     "mysql",
    Host:     "localhost",
    Port:     3306,
    Username: "root",
    Password: "password",
    Database: "myapp",
}

// PostgreSQL
postgresConfig := &orm.DatabaseConfig{
    Type:     "postgres",
    Host:     "localhost",
    Port:     5432,
    Username: "postgres",
    Password: "password",
    Database: "myapp",
    SSLMode:  "disable",
    Schema:   "public",
}

// SQLite
sqliteConfig := &orm.DatabaseConfig{
    Type:     "sqlite",
    Database: "app.db",
}
```

### 2. 读写分离

```go
// 读写分离配置
rwConfig := &orm.ReadWriteConfig{
    Master: masterConfig,
    Slaves: []*orm.DatabaseConfig{slave1Config, slave2Config},
    LoadBalanceStrategy: "round_robin",
    FailoverEnabled: true,
    RetryAttempts: 3,
}

// 创建连接池管理器
poolManager, err := orm.NewConnectionPoolManager(rwConfig, orm.DefaultPoolConfig())
```

### 3. 查询缓存

```go
// 创建缓存查询
cachedQuery := orm.NewCachedQuery(db).
    WithCacheKey("user_list_active").
    WithTTL(5 * time.Minute)

// 带缓存的查询
var users []User
err := cachedQuery.Where("status", "=", "active").Find(&users)
```

### 4. 性能监控

```go
// 获取连接统计
stats, err := orm.GetConnectionStats()
fmt.Printf("活跃连接: %d, 空闲连接: %d\n", stats.InUse, stats.Idle)

// 打印连接统计
orm.PrintConnectionStats()

// 性能监控器
monitor := orm.NewPerformanceMonitor(500 * time.Millisecond)
stats := monitor.GetStats()
```

### 5. 错误处理和重试

```go
crud := orm.GetSimpleCRUD()

// 安全执行带重试
err := crud.SafeExecute(func() error {
    return crud.Create(&user)
}, 3) // 最大重试3次
```

### 6. 调试和开发工具

```go
// 调试模式 - 打印 SQL
debugCRUD := orm.DebugMode(true)
debugCRUD.Find(&users)

// DryRun 模式 - 预览 SQL 不执行
dryRunCRUD := orm.DryRun()
dryRunCRUD.Create(&user) // 只打印SQL，不执行

// 显示 SQL 语句
orm.ShowSQL(func(crud *orm.SimpleCRUD) {
    crud.Find(&users)
})
```

## 📊 模型定义

### 基础模型

```go
// 使用内置基础模型
type User struct {
    orm.BaseModel // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
    Name  string `gorm:"size:100" json:"name"`
    Email string `gorm:"size:100;uniqueIndex" json:"email"`
}

// 使用时间戳模型
type Log struct {
    orm.TimestampModel // 只包含时间字段
    Level   string `json:"level"`
    Message string `json:"message"`
}

// 使用用户追踪模型
type Article struct {
    orm.UserTrackingModel // 包含创建用户和更新用户字段
    Title   string `json:"title"`
    Content string `json:"content"`
}
```

### 模型关联

```go
type User struct {
    orm.BaseModel
    Name    string    `json:"name"`
    Profile Profile   `json:"profile"`
    Orders  []Order   `json:"orders"`
}

type Profile struct {
    orm.BaseModel
    UserID   uint   `json:"user_id"`
    Avatar   string `json:"avatar"`
    Bio      string `json:"bio"`
}

type Order struct {
    orm.BaseModel
    UserID uint    `json:"user_id"`
    Amount float64 `json:"amount"`
    Status string  `json:"status"`
}
```

## 🎯 最佳实践

### 1. 配置管理

```go
// 生产环境配置
func ProductionConfig() *orm.DatabaseConfig {
    return &orm.DatabaseConfig{
        Type:         "mysql",
        Host:         os.Getenv("DB_HOST"),
        Port:         3306,
        Username:     os.Getenv("DB_USER"),
        Password:     os.Getenv("DB_PASSWORD"),
        Database:     os.Getenv("DB_NAME"),
        Charset:      "utf8mb4",
        Timezone:     "UTC",
        MaxIdleConns: 20,
        MaxOpenConns: 200,
        MaxLifetime:  3600,
        LogLevel:     "warn",
        SlowQuery:    1000,
    }
}
```

### 2. 错误处理

```go
user, err := crud.First(&User{}, "email = ?", email)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("用户不存在")
    }
    return nil, fmt.Errorf("查询用户失败: %w", err)
}
```

### 3. 事务最佳实践

```go
err := orm.QuickTransaction(func(crud *orm.SimpleCRUD) error {
    // 1. 创建用户
    if err := crud.Create(&user); err != nil {
        return fmt.Errorf("创建用户失败: %w", err)
    }
    
    // 2. 创建用户资料
    profile.UserID = user.ID
    if err := crud.Create(&profile); err != nil {
        return fmt.Errorf("创建用户资料失败: %w", err)
    }
    
    // 3. 发送欢迎邮件（这里应该用异步队列）
    // 事务中不要执行耗时操作
    
    return nil
})
```

### 4. 分页查询优化

```go
// 使用仓库模式进行分页
userRepo := orm.GetRepository[User]()
users, total, err := userRepo.PaginateWhere(
    "status = ? AND created_at > ?", 
    page, pageSize, 
    "active", time.Now().AddDate(0, -1, 0),
)

// 计算分页信息
totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
hasNext := page < int(totalPages)
hasPrev := page > 1
```

### 5. 性能优化

```go
// 1. 使用批量操作
users := make([]*User, 1000)
err := crud.BatchCreate(users, 100) // 每批100条

// 2. 使用索引字段查询
err := crud.Find(&users, "email = ?", email) // email有唯一索引

// 3. 预加载关联数据
query := orm.GetQueryBuilder[User]()
users, err := query.Preload("Profile", "Orders").Find()

// 4. 使用缓存
cachedQuery := orm.NewCachedQuery(db).WithCacheKey("active_users").WithTTL(5*time.Minute)
err := cachedQuery.Find(&users, "status = ?", "active")
```

## 🛠️ 配置参考

### 数据库配置

```go
type DatabaseConfig struct {
    // 基础连接配置
    Type         string `json:"type"`          // mysql, postgres, sqlite, sqlserver
    Host         string `json:"host"`          // 数据库主机
    Port         int    `json:"port"`          // 端口
    Username     string `json:"username"`      // 用户名
    Password     string `json:"password"`      // 密码
    Database     string `json:"database"`      // 数据库名
    
    // 连接池配置
    MaxIdleConns int `json:"max_idle_conns"` // 最大空闲连接数
    MaxOpenConns int `json:"max_open_conns"` // 最大打开连接数
    MaxLifetime  int `json:"max_lifetime"`   // 连接最大生存时间(秒)
    
    // 性能配置
    LogLevel     string `json:"log_level"`    // 日志级别: silent, error, warn, info
    SlowQuery    int    `json:"slow_query"`   // 慢查询阈值(毫秒)
    
    // 扩展配置
    Charset      string `json:"charset"`      // 字符集
    Timezone     string `json:"timezone"`     // 时区
    SSLMode      string `json:"ssl_mode"`     // SSL模式
    Schema       string `json:"schema"`       // 数据库schema
    TablePrefix  string `json:"table_prefix"` // 表前缀
}
```

### 连接池配置

```go
type PoolConfig struct {
    MaxIdleConns    int           `json:"max_idle_conns"`    // 最大空闲连接数
    MaxOpenConns    int           `json:"max_open_conns"`    // 最大打开连接数
    ConnMaxLifetime time.Duration `json:"conn_max_lifetime"` // 连接最大生存时间
    ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"` // 连接最大空闲时间
}
```

## 🔍 健康检查

```go
// 简单健康检查
err := orm.HealthCheck()
if err != nil {
    log.Printf("数据库连接异常: %v", err)
}

// 带超时的健康检查
err := orm.HealthCheckWithTimeout(5 * time.Second)
if err != nil {
    log.Printf("数据库健康检查超时: %v", err)
}
```

## 📈 监控和指标

```go
// 获取连接池统计
stats, err := orm.GetConnectionStats()
if err == nil {
    fmt.Printf("连接池状态:\n")
    fmt.Printf("  最大连接数: %d\n", stats.MaxOpenConnections)
    fmt.Printf("  活跃连接数: %d\n", stats.InUse)
    fmt.Printf("  空闲连接数: %d\n", stats.Idle)
    fmt.Printf("  等待连接数: %d\n", stats.WaitCount)
    fmt.Printf("  等待时长: %v\n", stats.WaitDuration)
}

// 性能监控
monitor := orm.NewPerformanceMonitor(500 * time.Millisecond)
// ... 执行查询操作
performanceStats := monitor.GetStats()
fmt.Printf("性能统计: %+v\n", performanceStats)
```

## 🚨 故障排除

### 常见问题

1. **连接超时**
   ```go
   // 调整连接超时设置
   config.MaxOpenConns = 50
   config.MaxIdleConns = 10
   config.MaxLifetime = 1800 // 30分钟
   ```

2. **慢查询优化**
   ```go
   // 启用慢查询日志
   config.LogLevel = "info"
   config.SlowQuery = 500 // 500ms
   
   // 使用调试模式查看SQL
   orm.DebugMode(true).Find(&users)
   ```

3. **内存占用过高**
   ```go
   // 使用分页查询避免大量数据
   // 启用查询缓存减少数据库压力
   // 调整连接池大小
   ```

## 🔗 相关链接

- [GORM 官方文档](https://gorm.io/)
- [YYHertz 框架文档](../README.md)
- [MyBatis 集成](../mybatis/README.md)

## 📄 许可证

MIT License