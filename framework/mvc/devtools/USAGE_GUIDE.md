# YYHertz DevTools 使用指南

## 🎯 问题解决

所有面板现在都可以正常访问！以下是完整的使用指南：

## 📋 可用的监控面板

### ✅ 基础面板（默认可用）
使用 `SetupBasicDevTools(app)` 即可访问：

- **调试面板**: `http://localhost/yyhertz/debug/panel`
- **健康检查面板**: `http://localhost/yyhertz/health/panel`
- **🆕 缓存监控**: `http://localhost/yyhertz/cache/panel`
- **🆕 安全监控**: `http://localhost/yyhertz/security/panel`

### ✅ 完整面板（需要完整配置）
使用带配置的 `SetupDevTools(app, config)` 可访问：

- **QPS监控面板**: `http://localhost/yyhertz/qps/panel`
- **统一监控面板**: `http://localhost/yyhertz/metrics/panel`
- **性能分析面板**: `http://localhost/yyhertz/profile/panel`
- **限流管理面板**: `http://localhost/yyhertz/ratelimit/panel`
- **🆕 数据库监控**: `http://localhost/yyhertz/database/panel`

## 🚀 快速开始

### 1. 最简单的使用方式
```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/devtools"
)

func main() {
    app := mvc.NewApp()
    
    // 启用基础开发工具（包含新面板！）
    devtools.SetupBasicDevTools(app)
    
    app.Run(":8080")
}
```

**现在可访问：**
- `http://localhost:8080/yyhertz/debug/panel`
- `http://localhost:8080/yyhertz/health/panel`
- `http://localhost:8080/yyhertz/cache/panel` ⭐ **新增**
- `http://localhost:8080/yyhertz/security/panel` ⭐ **新增**

### 2. 完整功能使用方式
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
        EnableQPS:         true,
        EnableMetrics:     true,
        EnableProfiler:    true,
        EnableRateLimit:   true,
        EnableHotReload:   true,
        EnableCache:       true,  // ⭐ 启用缓存监控
        EnableDatabase:    true,  // ⭐ 启用数据库监控
        EnableSecurity:    true,  // ⭐ 启用安全监控
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

**现在可访问所有面板：**
- `http://localhost:8080/yyhertz/debug/panel`
- `http://localhost:8080/yyhertz/health/panel`
- `http://localhost:8080/yyhertz/qps/panel`
- `http://localhost:8080/yyhertz/metrics/panel`
- `http://localhost:8080/yyhertz/profile/panel`
- `http://localhost:8080/yyhertz/ratelimit/panel`
- `http://localhost:8080/yyhertz/cache/panel` ⭐ **新增**
- `http://localhost:8080/yyhertz/database/panel` ⭐ **新增**
- `http://localhost:8080/yyhertz/security/panel` ⭐ **新增**

### 3. 生产环境使用方式
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
    
    db, _ := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
    
    // 生产环境配置（自动包含所有新面板）
    devtools.SetupProductionDevTools(app, db, "localhost:6379")
    
    app.Run(":8080")
}
```

## 🆕 新增面板功能说明

### 1. 缓存监控 (`/yyhertz/cache/panel`)
- **功能**: 监控Redis缓存性能
- **指标**: 命中率、延迟统计、热点Key、内存使用
- **实时数据**: 连接状态、操作统计
- **无依赖**: 可独立运行，不需要Redis连接

### 2. 数据库监控 (`/yyhertz/database/panel`)  
- **功能**: 监控数据库连接和查询性能
- **指标**: 慢查询、连接池状态、事务统计
- **实时数据**: 查询统计、连接详情
- **依赖要求**: 需要提供数据库连接 (`config.Database`)

### 3. 安全监控 (`/yyhertz/security/panel`)
- **功能**: 实时安全威胁检测和防护
- **检测能力**: SQL注入、XSS攻击、暴力破解、路径遍历
- **防护机制**: 自动IP黑名单、白名单管理、威胁评分
- **无依赖**: 可独立运行，无需额外配置

## 🔧 配置说明

### 功能开关
```go
type DevToolsConfig struct {
    // 新增的配置项
    EnableCache    bool // 启用缓存监控
    EnableDatabase bool // 启用数据库监控  
    EnableSecurity bool // 启用安全监控
    
    // 依赖项
    Database *sql.DB // 数据库连接（数据库监控需要）
}
```

### 默认配置
- **基础配置** (`SetupBasicDevTools`): 启用 `Cache` 和 `Security`，不启用 `Database`
- **完整配置** (`DefaultDevToolsConfig`): 启用所有功能
- **生产配置** (`SetupProductionDevTools`): 启用所有功能，禁用调试和热重载

## ✅ 验证步骤

1. **启动应用**
2. **检查日志输出** - 会显示所有可用面板的访问地址
3. **访问面板** - 根据配置访问对应的监控面板

## 🎉 问题已解决

现在所有的监控面板都可以正常访问了！根据你的配置选择：

- 使用 `SetupBasicDevTools()` - 可访问 4 个面板
- 使用完整配置 - 可访问所有 9 个面板

每个面板都有完整的功能实现和Web界面！