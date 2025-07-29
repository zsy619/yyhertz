# 📦 version.go 使用说明

`version.go` 文件为Hertz MVC框架提供了完整的版本管理和系统信息功能。

## 🎯 主要功能

### 1. 版本信息管理
- 框架版本号和构建信息
- 依赖库版本跟踪
- 构建时间和平台信息
- 许可证和作者信息

### 2. 系统监控
- 运行时信息获取
- 内存使用统计
- Goroutine计数
- 健康状态检查

### 3. 美观的输出
- ASCII艺术横幅
- 格式化的版本信息
- 彩色emoji标识
- JSON格式数据输出

## 📋 常量定义

```go
const (
    FrameworkName    = "Hertz MVC"    // 框架名称
    FrameworkVersion = "1.0.0"        // 框架版本
    BuildDate        = "2024-07-29"   // 构建日期
    HertzVersion     = "v0.10.1"      // Hertz版本
    Author          = "CloudWeGo Team" // 作者
    License         = "Apache 2.0"    // 许可证
)
```

## 🔧 核心函数

### 版本信息函数

| 函数名 | 返回类型 | 功能描述 |
|--------|----------|----------|
| `GetVersionInfo()` | `*VersionInfo` | 获取完整版本信息结构体 |
| `GetVersionString()` | `string` | 获取简短版本字符串 |
| `GetBuildInfo()` | `string` | 获取构建信息字符串 |
| `PrintVersion()` | `void` | 打印格式化版本信息 |
| `PrintBanner()` | `void` | 打印ASCII艺术横幅 |

### 系统信息函数

| 函数名 | 返回类型 | 功能描述 |
|--------|----------|----------|
| `GetSystemInfo()` | `map[string]any` | 获取系统运行信息 |
| `GetHealthStatus()` | `map[string]any` | 获取健康状态 |
| `GetFeatures()` | `[]string` | 获取框架特性列表 |
| `CheckDependencies()` | `bool` | 检查依赖兼容性 |
| `IsDebugMode()` | `bool` | 检查是否为调试模式 |

## 💻 使用示例

### 1. 命令行版本信息

```bash
# 显示版本信息
./your-app -version

# 自定义端口启动
./your-app -port=8080

# 禁用启动横幅
./your-app -banner=false
```

### 2. 代码中使用

```go
package main

import "fmt"

func main() {
    // 打印启动横幅
    PrintBanner()
    
    // 获取版本字符串
    version := GetVersionString()
    fmt.Println("当前版本:", version)
    
    // 获取完整版本信息
    info := GetVersionInfo()
    fmt.Printf("框架: %s v%s\n", info.Framework, info.Version)
    fmt.Printf("平台: %s/%s\n", info.Platform, info.Arch)
    
    // 检查系统状态
    if CheckDependencies() {
        fmt.Println("✅ 所有依赖都兼容")
    }
    
    // 获取系统信息
    sysInfo := GetSystemInfo()
    fmt.Printf("CPU核心数: %d\n", sysInfo["cpu_count"])
    fmt.Printf("Goroutine数: %d\n", sysInfo["goroutine_count"])
}
```

### 3. HTTP接口使用

在增强版主程序中，version.go提供了以下HTTP接口：

```bash
# 获取版本信息
curl http://localhost:8888/system/version

# 获取健康状态
curl http://localhost:8888/system/health

# 获取完整系统信息
curl http://localhost:8888/system/info
```

### 4. JSON响应格式

#### 版本信息接口 (`/system/version`)
```json
{
  "framework": "Hertz MVC",
  "version": "1.0.0",
  "build_date": "2024-07-29",
  "build_time": "2025-07-29 10:44:23",
  "go_version": "go1.24.5",
  "platform": "darwin",
  "arch": "amd64",
  "dependencies": {
    "go": "go1.24.5",
    "hertz": "v0.10.1"
  },
  "author": "CloudWeGo Team",
  "license": "Apache 2.0",
  "repository": "https://github.com/cloudwego/hertz",
  "homepage": "https://www.cloudwego.io/zh/docs/hertz/"
}
```

#### 健康检查接口 (`/system/health`)
```json
{
  "status": "healthy",
  "timestamp": 1753757063,
  "uptime": "2h30m15s",
  "version": "1.0.0",
  "framework": "Hertz MVC"
}
```

#### 系统信息接口 (`/system/info`)
```json
{
  "version": { /* 版本信息 */ },
  "system": {
    "go_version": "go1.24.5",
    "go_os": "darwin",
    "go_arch": "amd64",
    "cpu_count": 8,
    "goroutine_count": 15,
    "memory_usage": {
      "alloc_mb": 2,
      "total_alloc_mb": 5,
      "sys_mb": 12,
      "num_gc": 3
    },
    "framework": {
      "name": "Hertz MVC",
      "version": "1.0.0",
      "hertz": "v0.10.1"
    }
  },
  "features": [
    "🎯 基于Controller的架构设计",
    "⚡ 高性能HTTP服务器(基于Hertz)",
    "🔄 自动路由注册机制",
    /* ... 更多特性 */
  ]
}
```

## 🎨 启动横幅示例

```
██   ██ ███████ ██████  ████████ ███████     ███    ███ ██    ██  ██████ 
██   ██ ██      ██   ██    ██    ██          ████  ████ ██    ██ ██      
███████ █████   ██████     ██    ███████     ██ ████ ██ ██    ██ ██      
██   ██ ██      ██   ██    ██         ██     ██  ██  ██  ██  ██  ██      
██   ██ ███████ ██   ██    ██    ███████     ██      ██   ████    ██████ 

                    Hertz MVC Framework v1.0.0
                基于CloudWeGo-Hertz的类Beego框架
                    Build: 2024-07-29 | go1.24.5
```

## 🔧 自定义配置

你可以根据需要修改常量值：

```go
const (
    FrameworkName    = "你的框架名"
    FrameworkVersion = "2.0.0"
    BuildDate        = "2024-12-31"
    // ... 其他配置
)
```

## 📈 性能监控

`GetSystemInfo()` 函数提供了有用的性能指标：

- **内存使用**: 当前分配、总分配、系统内存
- **Goroutine数量**: 并发协程数量
- **GC次数**: 垃圾回收执行次数
- **CPU核心数**: 可用处理器核心

## 🛠️ 扩展建议

1. **添加自定义指标**: 在`GetSystemInfo()`中添加业务相关指标
2. **持久化日志**: 将版本和系统信息记录到日志文件
3. **监控集成**: 与Prometheus等监控系统集成
4. **配置文件**: 支持从配置文件读取版本信息

## ✨ 最佳实践

1. 在应用启动时调用`PrintBanner()`显示横幅
2. 使用`-version`参数快速查看版本信息
3. 在API中提供`/health`和`/version`端点用于监控
4. 定期检查`CheckDependencies()`确保兼容性
5. 在错误日志中包含版本信息便于调试

---

这个`version.go`文件为Hertz MVC框架提供了专业级的版本管理功能，让你的应用更加规范和易于维护！🚀