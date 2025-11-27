# 静态路径配置

YYHertz 框架提供了灵活而强大的静态文件服务配置，支持多种静态资源管理策略。

## 概述

静态文件服务是Web应用的重要组成部分，用于提供CSS、JavaScript、图片、字体等静态资源。YYHertz框架通过统一的配置接口管理静态文件路径映射。

## 基本配置

### 单一路径配置

最简单的静态文件配置方式：

```go
func main() {
    app := mvc.HertzApp
    
    // 基本的单一路径配置
    app.SetStaticPath("./static")
    
    // 或者指定URL路径
    app.SetStaticPath("./static", "/assets")
    
    app.Run()
}
```

**参数说明：**
- 第一个参数：本地文件系统路径（相对或绝对路径）
- 第二个参数：URL访问路径前缀

**访问示例：**
```
本地文件: ./static/css/style.css
访问URL:  http://localhost:8080/assets/css/style.css

本地文件: ./static/js/app.js  
访问URL:  http://localhost:8080/assets/js/app.js
```

### 多路径配置

支持配置多个静态文件路径：

```go
func main() {
    app := mvc.HertzApp
    
    // 使用 SetStaticPaths 批量配置（推荐方式）
    app.SetStaticPaths(map[string]string{
        "./assets":  "/assets",
        "./uploads": "/uploads",
        "./cdn":     "/cdn",
        "./public":  "/public",
    })
    
    // 或者单独配置
    app.SetStaticPath("./assets/css", "/css")
    app.SetStaticPath("./assets/js", "/js")
    app.SetStaticPath("./assets/images", "/images")
    app.SetStaticPath("./uploads", "/uploads")
    
    app.Run()
}
```

**多路径访问：**
```
./assets/css/bootstrap.css  → /css/bootstrap.css
./assets/js/jquery.js       → /js/jquery.js  
./assets/images/logo.png    → /images/logo.png
./uploads/avatar.jpg        → /uploads/avatar.jpg
```

## 实际应用示例

根据当前项目的main.go配置：

```go
func main() {
    // 创建应用实例
    app := mvc.HertzApp
    
    // 基础静态路径配置
    app.SetStaticPath("./static")
    
    // 设置静态文件路径 - 避免与默认路径冲突
    app.SetStaticPaths(map[string]string{
        "./assets":  "/assets",
        "./uploads": "/uploads",
        "./cdn":     "/cdn",
        "./public":  "/public",
    })
    
    // 可以通过以下方式访问：
    // http://localhost:8888/static/css/style.css  （默认路径）
    // http://localhost:8888/assets/css/style.css  （批量配置）
    // http://localhost:8888/uploads/file.pdf
    // http://localhost:8888/cdn/images/logo.png
    
    app.Run()
}
```

## 完整API方法说明

YYHertz框架提供了完整的静态路径配置API：

### SetStaticPath - 单个路径配置

```go
// 方法签名
func (app *App) SetStaticPath(localDir string, urlPath ...string)

// 使用示例
app.SetStaticPath("./static")           // 自动映射到 /static
app.SetStaticPath("./assets", "/cdn")   // 映射到 /cdn
app.SetStaticPath("./public")           // 自动映射到 /public
```

**参数说明：**
- `localDir`: 本地文件系统路径（相对或绝对路径）
- `urlPath`: 可选的URL路径前缀，如果不提供则自动推导

### SetStaticPaths - 批量路径配置（推荐）

```go
// 方法签名  
func (app *App) SetStaticPaths(pathMap map[string]string)

// 使用示例
app.SetStaticPaths(map[string]string{
    "./assets":  "/assets",
    "./uploads": "/uploads", 
    "./cdn":     "/cdn",
    "./public":  "/public",
})
```

**优势：**
- 一次性配置多个路径映射
- 代码更简洁，配置更集中
- 减少重复代码

### AddStaticPath - 添加单个路径

```go
// 方法签名
func (app *App) AddStaticPath(localPath, urlPath string)

// 使用示例
app.AddStaticPath("./downloads", "/downloads")
```

### AddStaticPaths - 添加多个路径

```go
// 方法签名
func (app *App) AddStaticPaths(pathMap map[string]string)

// 使用示例
additionalPaths := map[string]string{
    "./temp":   "/temp",
    "./backup": "/backup",
}
app.AddStaticPaths(additionalPaths)
```

### GetStaticPaths - 获取配置信息

```go
// 方法签名
func (app *App) GetStaticPaths() map[string]string

// 使用示例
func printStaticPathInfo(app *mvc.App) {
    staticPaths := app.GetStaticPaths()
    fmt.Println("已配置的静态文件路径:")
    for urlPath, localPath := range staticPaths {
        fmt.Printf("  %s -> %s\n", urlPath, localPath)
    }
}
```

### GetStaticPath - 获取默认路径（向后兼容）

```go
// 方法签名
func (app *App) GetStaticPath() string

// 使用示例
defaultPath := app.GetStaticPath()
fmt.Printf("默认静态路径: %s\n", defaultPath)
```

## YYHertz框架特性集成

### 统一常量系统

YYHertz框架提供了统一的执行层级常量系统，静态文件处理过程中可以使用这些常量：

```go
// 导入统一常量包
import "github.com/zsy619/yyhertz/framework/constant"

// 过滤器位置常量 - 使用统一常量
const (
    BeforeStatic = constant.BeforeStatic // 静态文件处理前 (0)
    BeforeRouter = constant.BeforeRouter // 路由匹配前 (1)  
    BeforeExec   = constant.BeforeExec   // 控制器执行前 (3)
    AfterExec    = constant.AfterExec    // 控制器执行后 (4)
    FinishRouter = constant.FinishRouter // 请求处理完成后 (5)
)

// 验证过滤器位置是否有效
func setupStaticFilters(app *mvc.App) {
    position := constant.BeforeStatic
    
    if constant.IsValidFilterPosition(position) {
        mvc.InsertFilter("/*", position, func(ctx *mvc.Context) {
            // 静态文件处理前的逻辑
            path := string(ctx.Request.Path())
            log.Printf("[STATIC] 处理静态文件请求: %s", path)
        })
    }
}
```

### 过滤器层级详解

静态文件请求在YYHertz框架中会经过以下层级：

```go
func setupComprehensiveStaticFilters(app *mvc.App) {
    // 0. BeforeStatic - 静态文件处理前
    // 这是最早的拦截点，可以在静态文件处理之前进行预处理
    mvc.InsertFilter("/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        path := string(ctx.Request.Path())
        method := string(ctx.Request.Method())
        log.Printf("[BeforeStatic] %s %s - 请求开始", method, path)
        
        // 可以在这里添加：
        // - 请求日志记录
        // - 访问权限检查
        // - 请求头处理
    })
    
    // 1. BeforeRouter - 路由匹配前
    // 在路由解析之前，但在静态文件处理之后
    mvc.InsertFilter("/*", mvc.BeforeRouter, func(ctx *mvc.Context) {
        path := string(ctx.Request.Path())
        
        // 检查是否为静态文件请求
        if isStaticRequest(path) {
            log.Printf("[BeforeRouter] 静态文件请求: %s", path)
            // 添加静态文件特定的处理逻辑
        }
    })
    
    // 3. BeforeExec - 控制器执行前
    // 对于静态文件，这个层级通常不会被触发（除非路由到控制器）
    mvc.InsertFilter("/*", mvc.BeforeExec, func(ctx *mvc.Context) {
        ctx.Set("start_time", tool.GetCurrentTimeMillis())
    })
    
    // 4. AfterExec - 控制器执行后
    // 记录处理时间和响应信息
    mvc.InsertFilter("/*", mvc.AfterExec, func(ctx *mvc.Context) {
        if startTime, exists := ctx.Get("start_time"); exists {
            if start, ok := startTime.(int64); ok {
                duration := tool.GetCurrentTimeMillis() - start
                path := string(ctx.Request.Path())
                log.Printf("[AfterExec] 请求处理完成: %s - 耗时: %dms", path, duration)
            }
        }
    })
    
    // 5. FinishRouter - 请求处理完成后
    // 最终的清理和日志记录
    mvc.InsertFilter("/*", mvc.FinishRouter, func(ctx *mvc.Context) {
        path := string(ctx.Request.Path())
        status := ctx.Ctx.Response.StatusCode()
        log.Printf("[FinishRouter] 请求完成: %s - 状态码: %d", path, status)
    })
}

// 辅助函数：判断是否为静态文件请求
func isStaticRequest(path string) bool {
    staticPrefixes := []string{"/static", "/assets", "/uploads", "/cdn", "/public"}
    for _, prefix := range staticPrefixes {
        if strings.HasPrefix(path, prefix) {
            return true
        }
    }
    return false
}
```

### 开发工具集成

YYHertz框架提供了开发工具支持，可以与静态文件配置结合使用：

```go
import "github.com/zsy619/yyhertz/framework/mvc/devtools"

func setupDevelopmentStaticFiles(app *mvc.App) {
    // 设置开发工具
    if err := devtools.SetupDevTools(app); err != nil {
        log.Printf("设置开发工具失败: %v", err)
    }
    
    // 开发环境的静态文件配置
    if isDevelopment() {
        // 禁用静态文件缓存
        mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
            ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
            ctx.Header("Pragma", "no-cache")
            ctx.Header("Expires", "0")
        })
        
        // 热重载支持
        mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
            // 检查文件是否有更新
            path := string(ctx.Request.Path())
            if strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js") {
                ctx.Header("X-Dev-Mode", "true")
            }
        })
        
        log.Println("开发模式已启用:")
        log.Println("- 静态文件缓存已禁用")
        log.Println("- 热重载监控已启用")
        log.Println("- 调试面板: http://localhost:8888/debug/panel")
    }
}

### 环境相关配置

根据不同环境配置不同的静态路径：

```go
func setupStaticPaths(app *mvc.App) {
    env := os.Getenv("ENV")
    
    switch env {
    case "development":
        // 开发环境 - 使用相对路径，便于调试
        app.SetStaticPath("./static")
        app.SetStaticPaths(map[string]string{
            "./assets":  "/assets",
            "./uploads": "/uploads",
        })
        
        // 开发环境特定配置
        setupDevelopmentStaticFiles(app)
        
    case "production":
        // 生产环境 - 使用绝对路径或CDN
        app.SetStaticPaths(map[string]string{
            "/var/www/static":  "/assets",
            "/var/www/uploads": "/uploads",
        })
        
        // 生产环境优化
        setupProductionStaticOptimization(app)
        
    case "staging":
        // 测试环境
        app.SetStaticPaths(map[string]string{
            "./static":  "/assets",
            "./uploads": "/uploads",
        })
        
    default:
        // 默认配置
        app.SetStaticPath("./static")
    }
}
```

### 容器化部署配置

在Docker容器中部署时的静态文件配置：

```go
func setupDockerStaticPaths(app *mvc.App) {
    // Docker容器内的路径配置
    app.SetStaticPaths(map[string]string{
        "/app/static":  "/assets",
        "/app/uploads": "/uploads",
        "/app/public":  "/public",
    })
    
    // 挂载卷的配置
    if volumePath := os.Getenv("STATIC_VOLUME_PATH"); volumePath != "" {
        app.SetStaticPath(volumePath, "/shared")
    }
    
    // 健康检查静态文件
    mvc.InsertFilter("/health/static", mvc.BeforeRouter, func(ctx *mvc.Context) {
        staticPaths := app.GetStaticPaths()
        healthInfo := make(map[string]interface{})
        
        for urlPath, localPath := range staticPaths {
            if _, err := os.Stat(localPath); err == nil {
                healthInfo[urlPath] = "healthy"
            } else {
                healthInfo[urlPath] = "error: " + err.Error()
            }
        }
        
        ctx.JSON(200, map[string]interface{}{
            "static_paths": healthInfo,
            "timestamp":    time.Now().Unix(),
        })
    })
}
```

### 微服务架构配置

在微服务环境中的静态文件策略：

```go
func setupMicroserviceStaticPaths(app *mvc.App) {
    serviceName := os.Getenv("SERVICE_NAME")
    
    // 服务特定的静态路径
    serviceStaticPath := fmt.Sprintf("/service/%s/assets", serviceName)
    app.SetStaticPath("./static", serviceStaticPath)
    
    // 共享资源路径
    if sharedPath := os.Getenv("SHARED_STATIC_PATH"); sharedPath != "" {
        app.SetStaticPath(sharedPath, "/shared")
    }
    
    // 注册服务发现的静态资源信息
    registerStaticResourceInfo(app, serviceName)
}

func registerStaticResourceInfo(app *mvc.App, serviceName string) {
    // 向服务注册中心注册静态资源信息
    staticInfo := map[string]interface{}{
        "service": serviceName,
        "paths":   app.GetStaticPaths(),
        "version": os.Getenv("SERVICE_VERSION"),
    }
    
    // 提供静态资源信息的API端点
    app.GET("/api/static/info", func(ctx context.Context, c *mvc.RequestContext) {
        c.JSON(200, staticInfo)
    })
}
```

### 负载均衡配置

在负载均衡环境中的静态文件处理：

```go
func setupLoadBalancedStaticPaths(app *mvc.App) {
    // 多节点共享存储
    sharedStoragePath := os.Getenv("SHARED_STORAGE_PATH")
    if sharedStoragePath != "" {
        app.SetStaticPaths(map[string]string{
            sharedStoragePath + "/assets":  "/assets",
            sharedStoragePath + "/uploads": "/uploads",
        })
    }
    
    // 节点本地缓存
    localCachePath := os.Getenv("LOCAL_CACHE_PATH")
    if localCachePath != "" {
        app.SetStaticPath(localCachePath, "/cache")
    }
    
    // 添加节点标识头
    nodeId := os.Getenv("NODE_ID")
    mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        if nodeId != "" {
            ctx.Header("X-Served-By", nodeId)
        }
        ctx.Header("X-Cache-Status", "HIT")
    })
}
```

### CDN集成配置

结合CDN使用的配置策略：

```go
func setupCDNStaticPaths(app *mvc.App) {
    enableCDN := os.Getenv("ENABLE_CDN") == "true"
    cdnHost := os.Getenv("CDN_HOST")
    
    if enableCDN && cdnHost != "" {
        // CDN环境下，本地只提供上传文件和动态内容
        app.SetStaticPaths(map[string]string{
            "./uploads": "/uploads",
            "./temp":    "/temp",
        })
        
        // 注册CDN相关的模板函数
        mvc.AddFuncMap("assetURL", func(path string) string {
            return fmt.Sprintf("%s/assets%s", cdnHost, path)
        })
        
        mvc.AddFuncMap("cdnURL", func(path string) string {
            return fmt.Sprintf("%s%s", cdnHost, path)
        })
        
        // CDN回源处理
        mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
            // 如果是CDN回源请求，提供本地文件
            if origin := string(ctx.Request.Header.Get("X-Origin-Request")); origin == "cdn" {
                log.Printf("[CDN] CDN回源请求: %s", string(ctx.Request.Path()))
                // 设置CDN缓存头
                ctx.Header("Cache-Control", "public, max-age=31536000")
                ctx.Header("X-CDN-Cache", "MISS")
            }
        })
        
    } else {
        // 本地环境提供所有静态文件
        app.SetStaticPaths(map[string]string{
            "./static":  "/assets",
            "./uploads": "/uploads",
        })
        
        mvc.AddFuncMap("assetURL", func(path string) string {
            return "/assets" + path
        })
        
        mvc.AddFuncMap("cdnURL", func(path string) string {
            return path
        })
    }
}
```

**模板中使用CDN函数：**
```html
<!-- 使用CDN或本地资源 -->
<link rel="stylesheet" href="{{assetURL "/css/style.css"}}">
<script src="{{assetURL "/js/app.js"}}"></script>

<!-- 直接使用CDN -->
<img src="{{cdnURL "/images/logo.png"}}" alt="Logo">
```

### 版本控制配置

为静态资源添加版本控制：

```go
func setupVersionedStaticPaths(app *mvc.App) {
    version := getAssetVersion() // 获取资源版本号
    
    // 带版本号的路径配置
    versionedPath := fmt.Sprintf("/assets/v%s", version)
    app.SetStaticPath("./static", versionedPath)
    
    // 注册版本化资源URL函数
    mvc.AddFuncMap("versionedAsset", func(path string) string {
        return fmt.Sprintf("%s%s", versionedPath, path)
    })
    
    // 兼容旧版本的重定向
    app.SetStaticPath("./static", "/assets")
}

func getAssetVersion() string {
    // 可以从配置文件、环境变量或构建时间获取
    if version := os.Getenv("ASSET_VERSION"); version != "" {
        return version
    }
    return "1.0.0"
}
```

## 性能优化

### 缓存控制

配置静态文件的缓存策略：

```go
func setupStaticCacheControl(app *mvc.App) {
    // 为不同类型的静态资源设置不同的缓存策略
    
    // CSS和JS文件 - 长时间缓存
    mvc.InsertFilter("/assets/css/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        ctx.Header("Cache-Control", "public, max-age=31536000") // 1年
        ctx.Header("ETag", generateETag(ctx.Request.Path()))
    })
    
    mvc.InsertFilter("/assets/js/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        ctx.Header("Cache-Control", "public, max-age=31536000") // 1年
        ctx.Header("ETag", generateETag(ctx.Request.Path()))
    })
    
    // 图片文件 - 中等时间缓存
    mvc.InsertFilter("/assets/images/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        ctx.Header("Cache-Control", "public, max-age=2592000") // 30天
    })
    
    // 上传文件 - 较短时间缓存
    mvc.InsertFilter("/uploads/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        ctx.Header("Cache-Control", "public, max-age=86400") // 1天
    })
}

func generateETag(path string) string {
    // 基于文件修改时间生成ETag
    if info, err := os.Stat("./static" + path); err == nil {
        return fmt.Sprintf(`"%x"`, info.ModTime().Unix())
    }
    return `"static"`
}
```

### 过滤器层级的配置

YYHertz提供了统一的过滤器层级系统，可以在不同层级配置静态文件处理：

```go
func setupStaticFilters(app *mvc.App) {
    // BeforeStatic (0) - 静态文件处理前
    mvc.InsertFilter("/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        path := string(ctx.Request.Path())
        method := string(ctx.Request.Method())
        log.Printf("[%s] %s - 请求开始", method, path)
    })
    
    // 当前项目中的过滤器示例（来自main.go）
    mvc.InsertFilter("/*", mvc.BeforeRouter, func(ctx *mvc.Context) {
        method := string(ctx.Request.Method())
        path := string(ctx.Request.Path())
        log.Printf("[%s] %s - 请求开始", method, path)
    })
    
    // BeforeExec (3) - 控制器执行前
    mvc.InsertFilter("/*", mvc.BeforeExec, func(ctx *mvc.Context) {
        ctx.Set("start_time", tool.GetCurrentTimeMillis())
    })
    
    // AfterExec (4) - 控制器执行后
    mvc.InsertFilter("/*", mvc.AfterExec, func(ctx *mvc.Context) {
        if startTime, exists := ctx.Get("start_time"); exists {
            if start, ok := startTime.(int64); ok {
                duration := tool.GetCurrentTimeMillis() - start
                path := string(ctx.Request.Path())
                log.Printf("请求处理完成: %s - 耗时: %dms", path, duration)
            }
        }
    })
}
```

## 安全配置

### 路径安全

防止路径遍历攻击：

```go
func setupStaticSecurity(app *mvc.App) {
    // 路径安全过滤器
    mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        path := string(ctx.Request.Path())
        
        // 检查路径遍历
        if strings.Contains(path, "..") {
            ctx.SetStatusCode(400)
            ctx.WriteString("非法路径")
            return
        }
        
        // 检查隐藏文件
        if strings.HasPrefix(filepath.Base(path), ".") {
            ctx.SetStatusCode(404)
            return
        }
    })
    
    // 限制访问的文件类型
    allowedExtensions := map[string]bool{
        ".css":   true,
        ".js":    true,
        ".png":   true,
        ".jpg":   true,
        ".jpeg":  true,
        ".gif":   true,
        ".svg":   true,
        ".ico":   true,
        ".woff":  true,
        ".woff2": true,
        ".ttf":   true,
        ".eot":   true,
    }
    
    mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        path := string(ctx.Request.Path())
        ext := strings.ToLower(filepath.Ext(path))
        
        if !allowedExtensions[ext] {
            ctx.SetStatusCode(403)
            ctx.WriteString("禁止访问的文件类型")
            return
        }
    })
}
```

## 监控和调试

### 访问日志

基于当前项目的过滤器配置，记录静态文件访问：

```go
func setupDetailedStaticLogging(app *mvc.App) {
    // 扩展现有的日志记录功能
    mvc.InsertFilter("/assets/*", mvc.BeforeRouter, func(ctx *mvc.Context) {
        method := string(ctx.Request.Method())
        path := string(ctx.Request.Path())
        userAgent := string(ctx.Request.Header.Get("User-Agent"))
        log.Printf("[STATIC] [%s] %s - User-Agent: %s", method, path, userAgent)
    })
    
    mvc.InsertFilter("/assets/*", mvc.BeforeExec, func(ctx *mvc.Context) {
        ctx.Set("static_start_time", tool.GetCurrentTimeMillis())
    })
    
    mvc.InsertFilter("/assets/*", mvc.AfterExec, func(ctx *mvc.Context) {
        if startTime, exists := ctx.Get("static_start_time"); exists {
            if start, ok := startTime.(int64); ok {
                duration := tool.GetCurrentTimeMillis() - start
                path := string(ctx.Request.Path())
                status := ctx.Ctx.Response.StatusCode()
                log.Printf("[STATIC] 静态文件访问完成: %s - 状态:%d - 耗时:%dms", path, status, duration)
            }
        }
    })
}
```

### 调试信息

获取和显示静态路径配置信息：

```go
func printStaticPathInfo(app *mvc.App) {
    fmt.Println("\n========== 静态路径配置信息 ==========")
    
    // 获取静态路径信息
    staticPaths := app.GetStaticPaths()
    fmt.Println("静态文件路径:")
    for urlPath, localPath := range staticPaths {
        fmt.Printf("  %s -> %s\n", urlPath, localPath)
        
        // 检查目录是否存在
        if _, err := os.Stat(localPath); os.IsNotExist(err) {
            fmt.Printf("    ⚠️  警告: 本地目录不存在: %s\n", localPath)
        } else {
            fmt.Printf("    ✅ 本地目录存在: %s\n", localPath)
        }
    }
    
    // 获取过滤器信息（现有功能）
    allFilters := app.GetAllFilters()
    fmt.Printf("静态文件相关过滤器: %d 个位置\n", len(allFilters))
    for position, filters := range allFilters {
        positionName := getFilterPositionName(position)
        fmt.Printf("  %s: %d 个过滤器\n", positionName, len(filters))
    }
    
    fmt.Println("=====================================")
}

// 复用main.go中的函数
func getFilterPositionName(position int) string {
    names := map[int]string{
        0: "BeforeStatic",
        1: "BeforeRouter", 
        3: "BeforeExec",
        4: "AfterExec",
        5: "FinishRouter",
    }
    if name, exists := names[position]; exists {
        return name
    }
    return "Unknown"
}
```

## 最佳实践

### 1. 目录结构组织

```
static/
├── css/                # 样式文件
│   ├── vendor/        # 第三方CSS
│   ├── components/    # 组件样式  
│   └── pages/         # 页面特定样式
├── js/                 # JavaScript文件
│   ├── vendor/        # 第三方JS库
│   ├── modules/       # 模块化JS
│   └── pages/         # 页面特定JS
├── images/             # 图片资源
│   ├── icons/         # 图标
│   ├── backgrounds/   # 背景图
│   └── content/       # 内容图片
└── fonts/              # 字体文件
```

### 2. 配置管理

```go
func setupStaticConfiguration(app *mvc.App) {
    // 基础静态路径配置
    app.SetStaticPath("./static", "/assets")
    
    // 注册模板函数
    registerTemplateFunctions(app)
    
    // 设置过滤器
    registerGlobalFilters(app)
    
    // 打印配置信息
    printStaticPathInfo(app)
}
```

### 3. 开发环境配置

```go
func setupDevelopmentAssets(app *mvc.App) {
    if os.Getenv("ENV") == "development" {
        // 开发环境禁用缓存
        mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
            ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
            ctx.Header("Pragma", "no-cache")
            ctx.Header("Expires", "0")
        })
        
        // 启用详细日志
        setupDetailedStaticLogging(app)
    }
}
```

## 故障排除和诊断

### 常见问题及解决方案

#### 1. 404错误 - 静态文件无法访问

**问题现象：**
```
GET /assets/css/style.css -> 404 Not Found
```

**诊断步骤：**
```go
func diagnoseStaticFile404(app *mvc.App, requestPath string) {
    fmt.Printf("=== 诊断静态文件404错误 ===\n")
    fmt.Printf("请求路径: %s\n", requestPath)
    
    // 1. 检查静态路径配置
    staticPaths := app.GetStaticPaths()
    fmt.Println("\n已配置的静态路径:")
    for urlPath, localPath := range staticPaths {
        fmt.Printf("  %s -> %s\n", urlPath, localPath)
        
        // 检查是否匹配
        if strings.HasPrefix(requestPath, urlPath) {
            fmt.Printf("    ✓ 匹配URL路径: %s\n", urlPath)
            
            // 构建本地文件路径
            relativePath := strings.TrimPrefix(requestPath, urlPath)
            fullPath := filepath.Join(localPath, relativePath)
            
            // 检查文件是否存在
            if info, err := os.Stat(fullPath); err == nil {
                fmt.Printf("    ✓ 本地文件存在: %s\n", fullPath)
                fmt.Printf("    📁 文件信息: 大小=%d, 修改时间=%s\n", 
                    info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
            } else {
                fmt.Printf("    ❌ 本地文件不存在: %s\n", fullPath)
                fmt.Printf("    📁 错误信息: %v\n", err)
                
                // 检查目录是否存在
                dir := filepath.Dir(fullPath)
                if _, dirErr := os.Stat(dir); os.IsNotExist(dirErr) {
                    fmt.Printf("    ❌ 目录不存在: %s\n", dir)
                } else {
                    fmt.Printf("    ✓ 目录存在: %s\n", dir)
                    // 列出目录内容
                    if files, err := os.ReadDir(dir); err == nil {
                        fmt.Printf("    📁 目录内容:\n")
                        for _, file := range files {
                            fmt.Printf("      - %s\n", file.Name())
                        }
                    }
                }
            }
        }
    }
    
    // 2. 检查路径冲突
    checkRouteConflicts(requestPath)
}

func checkRouteConflicts(path string) {
    fmt.Println("\n=== 检查路由冲突 ===")
    
    // 常见的路由冲突模式
    conflictPatterns := []string{"/api", "/admin", "/user", "/docs"}
    
    for _, pattern := range conflictPatterns {
        if strings.HasPrefix(path, pattern) {
            fmt.Printf("⚠️  可能的路由冲突: %s 可能与控制器路由冲突\n", pattern)
        }
    }
}
```

**解决方案：**
```go
// 方案1: 修正路径配置
app.SetStaticPath("./static", "/assets")  // 确保路径正确

// 方案2: 检查文件权限
// chmod 644 static/css/style.css
// chmod 755 static/

// 方案3: 避免路由冲突
app.SetStaticPath("./static", "/static")  // 使用不冲突的路径
```

#### 2. 权限错误 - 文件无法读取

**问题现象：**
```
Error: permission denied
```

**诊断函数：**
```go
func diagnoseFilePermissions(filePath string) {
    fmt.Printf("=== 诊断文件权限问题 ===\n")
    fmt.Printf("文件路径: %s\n", filePath)
    
    if info, err := os.Stat(filePath); err == nil {
        fmt.Printf("文件权限: %s\n", info.Mode().String())
        fmt.Printf("文件大小: %d bytes\n", info.Size())
        fmt.Printf("修改时间: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
        
        // 检查读取权限
        if file, err := os.Open(filePath); err == nil {
            file.Close()
            fmt.Println("✓ 文件可读取")
        } else {
            fmt.Printf("❌ 文件读取失败: %v\n", err)
            fmt.Println("💡 建议执行: chmod 644 " + filePath)
        }
    } else {
        fmt.Printf("❌ 文件状态检查失败: %v\n", err)
    }
    
    // 检查目录权限
    dir := filepath.Dir(filePath)
    if info, err := os.Stat(dir); err == nil {
        fmt.Printf("目录权限: %s\n", info.Mode().String())
        
        if info.Mode().Perm()&0111 == 0 {
            fmt.Println("❌ 目录缺少执行权限")
            fmt.Println("💡 建议执行: chmod 755 " + dir)
        }
    }
}
```

#### 3. 性能问题 - 静态文件加载缓慢

**诊断工具：**
```go
func diagnoseStaticPerformance(app *mvc.App) {
    // 添加性能监控过滤器
    mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        ctx.Set("static_start_time", time.Now())
    })
    
    mvc.InsertFilter("/assets/*", mvc.AfterExec, func(ctx *mvc.Context) {
        if startTime, exists := ctx.Get("static_start_time"); exists {
            if start, ok := startTime.(time.Time); ok {
                duration := time.Since(start)
                path := string(ctx.Request.Path())
                
                // 记录慢请求
                if duration > 100*time.Millisecond {
                    log.Printf("⚠️  慢静态文件请求: %s - 耗时: %v", path, duration)
                    
                    // 文件大小检查
                    if info, err := os.Stat(getLocalPath(path)); err == nil {
                        log.Printf("📁 文件大小: %d bytes", info.Size())
                        if info.Size() > 1024*1024 { // 1MB
                            log.Printf("💡 建议: 文件过大，考虑压缩或使用CDN")
                        }
                    }
                }
            }
        }
    })
}

func getLocalPath(urlPath string) string {
    // 根据URL路径获取本地文件路径的辅助函数
    // 这里需要根据实际的静态路径配置来实现
    return "./static" + strings.TrimPrefix(urlPath, "/assets")
}
```

#### 4. 缓存问题 - 文件更新不生效

**诊断和解决：**
```go
func setupCacheDebugging(app *mvc.App) {
    mvc.InsertFilter("/assets/*", mvc.BeforeStatic, func(ctx *mvc.Context) {
        path := string(ctx.Request.Path())
        
        // 检查文件修改时间
        if localPath := getLocalPath(path); localPath != "" {
            if info, err := os.Stat(localPath); err == nil {
                modTime := info.ModTime()
                
                // 添加调试头
                ctx.Header("X-File-ModTime", modTime.Format(time.RFC3339))
                ctx.Header("X-Debug-Path", localPath)
                
                // 开发环境禁用缓存
                if os.Getenv("ENV") == "development" {
                    ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
                    ctx.Header("Pragma", "no-cache")
                    ctx.Header("Expires", "0")
                } else {
                    // 生产环境使用ETag
                    etag := fmt.Sprintf(`"%x"`, modTime.Unix())
                    ctx.Header("ETag", etag)
                    
                    // 检查If-None-Match
                    if ifNoneMatch := string(ctx.Request.Header.Get("If-None-Match")); ifNoneMatch == etag {
                        ctx.SetStatusCode(304)
                        return
                    }
                }
            }
        }
    })
}
```

### 诊断工具函数

提供完整的静态文件配置诊断工具：

```go
func RunStaticPathDiagnostics(app *mvc.App) {
    fmt.Println("========== YYHertz 静态路径诊断报告 ==========\n")
    
    // 1. 基本配置信息
    printBasicStaticInfo(app)
    
    // 2. 文件系统检查
    checkFileSystemStatus(app)
    
    // 3. 过滤器配置检查
    checkFilterConfiguration(app)
    
    // 4. 性能检查
    checkPerformanceSettings(app)
    
    // 5. 安全检查
    checkSecuritySettings(app)
    
    fmt.Println("===============================================")
}

func printBasicStaticInfo(app *mvc.App) {
    fmt.Println("=== 基本配置信息 ===")
    
    staticPaths := app.GetStaticPaths()
    fmt.Printf("已配置静态路径数量: %d\n", len(staticPaths))
    
    for urlPath, localPath := range staticPaths {
        fmt.Printf("  %s -> %s\n", urlPath, localPath)
    }
    fmt.Println()
}

func checkFileSystemStatus(app *mvc.App) {
    fmt.Println("=== 文件系统状态检查 ===")
    
    staticPaths := app.GetStaticPaths()
    for urlPath, localPath := range staticPaths {
        if info, err := os.Stat(localPath); err == nil {
            fmt.Printf("✓ %s: 目录存在\n", localPath)
            fmt.Printf("  权限: %s\n", info.Mode().String())
            
            // 检查目录内容
            if files, err := os.ReadDir(localPath); err == nil {
                fmt.Printf("  文件数量: %d\n", len(files))
            }
        } else {
            fmt.Printf("❌ %s: 目录不存在或无法访问\n", localPath)
            fmt.Printf("  错误: %v\n", err)
        }
    }
    fmt.Println()
}

func checkFilterConfiguration(app *mvc.App) {
    fmt.Println("=== 过滤器配置检查 ===")
    
    allFilters := app.GetAllFilters()
    for position, filters := range allFilters {
        positionName := getFilterPositionName(position)
        fmt.Printf("%s (%d): %d 个过滤器\n", positionName, position, len(filters))
        
        for i, filter := range filters {
            fmt.Printf("  %d. 模式: %s, 启用: %t\n", i+1, filter.Pattern, filter.Enabled)
        }
    }
    fmt.Println()
}

func checkPerformanceSettings(app *mvc.App) {
    fmt.Println("=== 性能设置检查 ===")
    
    // 检查环境变量
    env := os.Getenv("ENV")
    fmt.Printf("运行环境: %s\n", env)
    
    if env == "development" {
        fmt.Println("💡 开发环境建议:")
        fmt.Println("  - 禁用静态文件缓存")
        fmt.Println("  - 启用详细日志记录")
        fmt.Println("  - 启用热重载监控")
    } else if env == "production" {
        fmt.Println("💡 生产环境建议:")
        fmt.Println("  - 启用静态文件缓存")
        fmt.Println("  - 配置CDN")
        fmt.Println("  - 启用文件压缩")
    }
    fmt.Println()
}

func checkSecuritySettings(app *mvc.App) {
    fmt.Println("=== 安全设置检查 ===")
    
    staticPaths := app.GetStaticPaths()
    for urlPath, localPath := range staticPaths {
        // 检查是否包含敏感路径
        if strings.Contains(localPath, "..") {
            fmt.Printf("⚠️  安全警告: %s 包含相对路径\n", localPath)
        }
        
        // 检查是否在Web根目录外
        if !strings.HasPrefix(localPath, "./") && !filepath.IsAbs(localPath) {
            fmt.Printf("💡 建议: %s 使用绝对路径或相对路径\n", localPath)
        }
    }
    fmt.Println()
}
```

## 总结

YYHertz的静态路径配置系统提供了全面而强大的静态文件管理能力：

### 核心特性

- **简单配置**: 通过 `SetStaticPath` 和 `SetStaticPaths` 方法轻松配置静态文件路径
- **多路径支持**: 可以配置多个不同的静态文件路径映射，满足复杂应用需求
- **统一常量系统**: 集成框架的统一执行层级常量，确保配置一致性
- **过滤器集成**: 与YYHertz的过滤器系统无缝集成，支持五个层级的拦截处理

### 高级能力

- **环境适配**: 支持开发、测试、生产等多环境配置
- **容器化支持**: 完整的Docker容器部署配置方案  
- **微服务架构**: 适配微服务环境的静态资源管理
- **负载均衡**: 支持多节点共享存储和本地缓存策略
- **CDN集成**: 完整的CDN集成方案，包括回源处理和模板函数支持

### 性能与安全

- **缓存优化**: 支持多层级缓存控制和ETag机制
- **性能监控**: 集成性能监控和慢请求检测
- **安全保护**: 提供路径安全、文件类型限制等安全措施
- **开发工具**: 集成热重载、调试面板等开发辅助功能

### 诊断与运维

- **完整诊断工具**: 提供全面的配置诊断和问题排查工具
- **详细错误分析**: 针对常见问题提供详细的诊断步骤和解决方案
- **监控集成**: 支持健康检查、日志记录和性能监控
- **故障排除**: 涵盖404错误、权限问题、性能问题等常见故障的解决方案

### 最佳实践

- **目录结构**: 提供合理的静态资源目录组织建议
- **配置管理**: 模块化的配置管理方法
- **环境隔离**: 不同环境的差异化配置策略
- **部署优化**: 针对容器化、微服务等现代部署方式的优化建议

通过合理配置和使用YYHertz的静态路径系统，可以有效管理Web应用的静态资源，提升用户体验和应用性能，同时确保系统的安全性和可维护性。

### 快速入门示例

```go
func main() {
    // 创建应用
    app := mvc.HertzApp
    
    // 基础配置
    app.SetStaticPaths(map[string]string{
        "./static":  "/assets",
        "./uploads": "/uploads",
    })
    
    // 可选：添加诊断工具（开发环境）
    if os.Getenv("ENV") == "development" {
        RunStaticPathDiagnostics(app)
    }
    
    app.Run()
}
```

这个全面的静态路径配置文档将帮助开发者充分利用YYHertz框架的静态文件管理能力，构建高性能、安全可靠的Web应用。