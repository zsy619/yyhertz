# YYHertz 开发工具

基于 CloudWeGo-Hertz 的二次封装框架开发工具集，提供代码生成、热重载、调试和性能监控功能。

## 功能特性

### 1. 代码生成工具

#### 自动路由生成
- 支持注解驱动的路由生成
- RESTful 路由自动生成
- 路由分组和中间件配置
- 类型安全的路由代码

#### API文档生成
- OpenAPI 3.0 规范文档
- Swagger UI 集成
- 自动解析控制器注解
- 支持 JSON/YAML 格式输出

#### 客户端代码生成
- TypeScript 客户端生成
- Go 客户端 SDK 生成
- 类型安全的请求/响应定义
- 完整的错误处理机制

### 2. 开发工具增强

#### 热重载支持
- 文件变化自动监控
- 智能防抖机制
- 可配置的监控目录和文件类型
- 优雅的服务器重启

#### 调试中间件
- 请求生命周期追踪
- 内存使用监控
- 错误堆栈跟踪
- 可视化调试面板

#### 性能监控
- 实时性能指标监控
- 端点级别的统计分析
- 历史趋势图表展示
- 系统资源监控

## 快速开始

### 1. 基本使用

```go
package main

import (
    "log"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/devtools"
)

func main() {
    // 创建应用
    app := mvc.NewApp()
    
    // 设置开发工具
    if err := setupDevTools(app); err != nil {
        log.Fatalf("设置开发工具失败: %v", err)
    }
    
    // 启动服务器
    app.Run(":8080")
}

func setupDevTools(app *mvc.App) error {
    // 1. 热重载配置
    hotReloadConfig := devtools.DefaultHotReloadConfig()
    hotReloader, err := devtools.NewHotReloadServer(app, hotReloadConfig)
    if err != nil {
        return err
    }
    
    // 2. 调试中间件
    debugMiddleware := devtools.NewDebugMiddleware()
    debugPanel := devtools.NewDebugPanel(debugMiddleware)
    
    // 3. 性能监控
    performanceMonitor := devtools.NewPerformanceMonitor()
    performancePanel := devtools.NewPerformancePanel(performanceMonitor)
    
    // 启动监控
    performanceMonitor.Start()
    
    // 注册中间件
    app.Use(debugMiddleware.Handler())
    app.Use(performanceMonitor.Middleware())
    
    // 注册路由
    debugPanel.RegisterRoutes(app.Engine)
    performancePanel.RegisterRoutes(app.Engine)
    
    // 启动热重载（开发环境）
    go hotReloader.Run(":8080")
    
    return nil
}
```

### 2. 代码生成

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc/codegen"
)

func main() {
    // 路由生成
    routeGen := codegen.NewRouteGenerator(".")
    if err := routeGen.GenerateFromAnnotations(); err != nil {
        panic(err)
    }
    
    // API文档生成
    docGen := codegen.NewDocGenerator(".")
    if err := docGen.GenerateOpenAPI(); err != nil {
        panic(err)
    }
    
    // 客户端代码生成
    clientGen := codegen.NewClientGenerator(".")
    if err := clientGen.Generate(controllers); err != nil {
        panic(err)
    }
}
```

## 访问地址

启动应用后，可以通过以下地址访问各种工具：

- **调试面板**: http://localhost:8080/debug/panel
- **性能监控**: http://localhost:8080/performance/panel
- **API文档**: http://localhost:8080/docs (需要生成)

## 配置选项

### 热重载配置

```go
config := devtools.HotReloadConfig{
    WatchDirs:   []string{".", "controllers", "views"},
    ExcludeDirs: []string{"logs", "tmp", ".git"},
    Extensions:  []string{".go", ".html", ".css", ".js"},
    Debounce:    500 * time.Millisecond,
    OnReload: func() error {
        // 自定义重载逻辑
        return nil
    },
}
```

### 调试中间件配置

```go
debugMiddleware := devtools.NewDebugMiddleware()
debugMiddleware.SetMaxBodySize(1024 * 1024) // 1MB
debugMiddleware.Enable() // 启用调试
```

### 性能监控配置

```go
monitor := devtools.NewPerformanceMonitor()
monitor.Start() // 启动监控
```

## 注意事项

1. **开发环境使用**: 这些工具主要用于开发环境，生产环境请谨慎使用
2. **性能影响**: 调试和监控功能会有一定的性能开销
3. **内存使用**: 调试信息会占用内存，注意设置合理的存储限制
4. **安全考虑**: 调试面板包含敏感信息，请勿在生产环境暴露

## 示例项目

完整的使用示例请参考 `example/devtools/main.go` 文件。

## 依赖项

- CloudWeGo-Hertz
- fsnotify (文件监控)
- Chart.js (性能图表)

## 许可证

本项目遵循与主项目相同的许可证。
