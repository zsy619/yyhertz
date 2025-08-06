# YYHertz 数据库配置说明

本目录包含了 YYHertz 框架的数据库配置文件示例，支持多种数据库类型和配置场景。

## 配置文件列表

### 1. `database.yaml` - 默认配置文件
- **用途**: 开发环境默认配置，使用 SQLite 数据库
- **特点**: 
  - 零配置启动，适合快速开发
  - 使用文件数据库，无需安装额外软件
  - 包含完整的配置选项说明

### 2. `database-mysql.yaml` - MySQL 配置示例
- **用途**: 生产环境 MySQL 数据库配置
- **特点**:
  - 支持读写分离
  - 优化的连接池配置
  - 启用 Redis 缓存
  - 生产环境安全设置

### 3. `database-postgres.yaml` - PostgreSQL 配置示例
- **用途**: PostgreSQL 数据库配置
- **特点**:
  - 支持多租户 Schema 策略
  - 优化的 PostgreSQL 特性配置
  - 企业级部署建议

## 使用方法

### 1. 选择配置文件
根据你的数据库类型，选择对应的配置文件：

```bash
# 使用 SQLite (默认)
cp database.yaml database-active.yaml

# 使用 MySQL
cp database-mysql.yaml database.yaml

# 使用 PostgreSQL  
cp database-postgres.yaml database.yaml
```

### 2. 修改配置参数
编辑选中的配置文件，修改以下关键参数：

```yaml
primary:
  host: "your-database-host"      # 数据库主机
  port: 3306                      # 数据库端口
  database: "your-database-name"  # 数据库名称
  username: "your-username"       # 用户名
  password: "your-password"       # 密码
```

### 3. 环境变量覆盖
你也可以使用环境变量覆盖配置文件中的设置：

```bash
export YYHERTZ_PRIMARY_HOST="localhost"
export YYHERTZ_PRIMARY_PORT="3306"
export YYHERTZ_PRIMARY_DATABASE="yyhertz"
export YYHERTZ_PRIMARY_USERNAME="root"
export YYHERTZ_PRIMARY_PASSWORD="password"
```

### 4. 程序中使用
在 Go 程序中使用数据库配置：

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/config"
    "github.com/zsy619/yyhertz/framework/orm"
)

func main() {
    // 获取数据库配置
    dbConfig, err := config.GetDatabaseConfig()
    if err != nil {
        log.Fatal("Failed to get database config:", err)
    }

    // 使用默认 ORM 实例
    db := orm.GetDefaultORM()
    
    // 或者创建新的 ORM 实例
    customORM, err := orm.NewORMWithConfig(dbConfig)
    if err != nil {
        log.Fatal("Failed to create ORM:", err)
    }
}
```

## 配置选项详解

### 主数据库配置 (primary)
- `driver`: 数据库驱动类型 (`mysql`, `postgres`, `sqlite`, `sqlserver`)
- `dsn`: 完整连接字符串（优先级高于单独配置）
- `host`: 数据库主机地址
- `port`: 数据库端口
- `database`: 数据库名称（SQLite 为文件路径）
- `username/password`: 认证信息
- `max_open_conns`: 最大打开连接数
- `max_idle_conns`: 最大空闲连接数
- `conn_max_lifetime`: 连接最大生存时间
- `slow_query_threshold`: 慢查询阈值

### GORM 配置 (gorm)
- `enable`: 是否启用 GORM
- `disable_foreign_key_constrain`: 禁用外键约束
- `skip_default_transaction`: 跳过默认事务
- `prepare_stmt`: 启用预编译语句
- `create_batch_size`: 批量创建大小
- `naming_strategy`: 命名策略（`snake_case` 或 `camel_case`）

### 连接池配置 (pool)
- `enable`: 启用连接池
- `max_active_conns`: 最大活跃连接数
- `test_on_borrow`: 借用时测试连接
- `validation_query`: 验证查询 SQL

### 缓存配置 (cache)
- `enable`: 启用查询缓存
- `type`: 缓存类型（`memory`, `redis`, `memcached`）
- `redis_addr`: Redis 服务器地址
- `ttl`: 缓存生存时间

### 监控配置 (monitoring)
- `enable`: 启用监控
- `slow_query_log`: 慢查询日志
- `export_format`: 导出格式（`prometheus`, `json`, `text`）

### 迁移配置 (migration)
- `enable`: 启用迁移
- `auto_migrate`: 自动迁移模型
- `path`: 迁移文件路径

## 数据库特定注意事项

### MySQL
- 推荐使用 `utf8mb4` 字符集
- 注意时区设置
- 生产环境建议启用 SSL

### PostgreSQL
- 支持强大的多租户 Schema 功能
- 推荐使用 `snake_case` 命名策略
- 支持丰富的数据类型

### SQLite
- 适合开发和测试环境
- 单文件数据库，便于备份
- 不支持并发写入

## 安全建议

1. **生产环境**:
   - 不要在配置文件中明文存储密码
   - 使用环境变量或密钥管理服务
   - 启用 SSL/TLS 连接

2. **权限控制**:
   - 为应用创建专用数据库用户
   - 仅授予必要的数据库权限
   - 定期轮换密码

3. **网络安全**:
   - 限制数据库服务器的网络访问
   - 使用防火墙规则
   - 考虑使用 VPN 或专用网络

## 性能优化建议

1. **连接池**:
   - 根据应用负载调整连接池大小
   - 监控连接使用情况
   - 设置合适的连接超时时间

2. **查询优化**:
   - 启用慢查询日志
   - 定期分析查询性能
   - 创建适当的索引

3. **缓存策略**:
   - 为频繁查询启用缓存
   - 设置合理的缓存过期时间
   - 考虑使用 Redis 集群

## 故障排除

### 常见问题

1. **连接失败**:
   ```
   Error: failed to connect to database
   ```
   - 检查数据库服务是否启动
   - 验证连接参数是否正确
   - 确认网络连通性

2. **权限错误**:
   ```
   Error: Access denied for user
   ```
   - 检查用户名和密码
   - 验证用户权限
   - 确认数据库是否存在

3. **连接池耗尽**:
   ```
   Error: connection pool exhausted
   ```
   - 增加最大连接数
   - 检查是否有连接泄漏
   - 优化查询性能

### 调试方法

1. **启用 SQL 日志**:
   ```yaml
   primary:
     log_level: "info"
   development:
     show_sql: true
   ```

2. **查看连接统计**:
   ```go
   stats := orm.GetDefaultORM().GetStats()
   fmt.Printf("Database stats: %+v\n", stats)
   ```

3. **健康检查**:
   ```go
   if err := orm.GetDefaultORM().Ping(); err != nil {
       log.Printf("Database health check failed: %v", err)
   }
   ```

## 测试

运行数据库测试套件：

```bash
# 运行所有数据库测试
go test -v ./database_test.go

# 运行特定测试
go test -v -run TestDatabaseConnection ./database_test.go

# 运行性能测试
go test -v -bench=. ./database_test.go
```

## 更多资源

- [GORM 官方文档](https://gorm.io/docs/)
- [Viper 配置管理](https://github.com/spf13/viper)
- [数据库驱动文档](https://gorm.io/docs/connecting_to_the_database.html)
- [YYHertz 框架文档](../../../README.md)