# 静态资源

YYHertz 框架提供了完善的静态资源管理功能，支持静态文件服务、资源打包、版本控制、CDN 集成等特性，帮助开发者高效管理前端资源。

## 概述

静态资源管理是现代 Web 应用的重要组成部分。YYHertz 的静态资源系统提供：

- 静态文件服务
- 资源路径管理
- 文件压缩和缓存
- 版本控制和缓存清除
- CDN 集成
- 资源打包和优化
- 热重载支持

## 基本配置

### 静态文件服务

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
)

func main() {
    app := mvc.HertzApp
    
    // 基本静态文件配置
    app.StaticPath = "./static"
    app.StaticURL = "/static"
    
    // 多个静态目录
    app.AddStaticPath("/assets", "./assets")
    app.AddStaticPath("/uploads", "./uploads")
    app.AddStaticPath("/images", "./static/images")
    
    // 单文件映射
    app.AddStaticFile("/favicon.ico", "./static/favicon.ico")
    app.AddStaticFile("/robots.txt", "./static/robots.txt")
    
    app.Run()
}
```

### 目录结构

```
static/
├── css/
│   ├── bootstrap.min.css
│   ├── app.css
│   └── admin.css
├── js/
│   ├── jquery.min.js
│   ├── bootstrap.min.js
│   ├── app.js
│   └── modules/
│       ├── user.js
│       └── dashboard.js
├── images/
│   ├── logo.png
│   ├── icons/
│   └── uploads/
├── fonts/
│   ├── roboto.woff2
│   └── icons.ttf
└── vendor/
    ├── bootstrap/
    ├── jquery/
    └── fontawesome/
```

## 资源引用

### 模板中的资源引用

```html
<!-- layouts/main.html -->
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    
    <!-- CSS 资源 -->
    <link href="{{static "css/bootstrap.min.css"}}" rel="stylesheet">
    <link href="{{static "css/app.css"}}" rel="stylesheet">
    
    <!-- 条件CSS -->
    {{if eq .Layout "admin"}}
        <link href="{{static "css/admin.css"}}" rel="stylesheet">
    {{end}}
    
    <!-- Favicon -->
    <link rel="icon" href="{{static "favicon.ico"}}" type="image/x-icon">
</head>
<body>
    <!-- 页面内容 -->
    {{template "content" .}}
    
    <!-- JavaScript 资源 -->
    <script src="{{static "js/jquery.min.js"}}"></script>
    <script src="{{static "js/bootstrap.min.js"}}"></script>
    <script src="{{static "js/app.js"}}"></script>
    
    <!-- 页面特定的JS -->
    {{if .PageJS}}
        {{range .PageJS}}
            <script src="{{static .}}"></script>
        {{end}}
    {{end}}
</body>
</html>
```

### 自定义静态资源函数

```go
package mvc

import (
    "html/template"
    "path"
    "strings"
)

func init() {
    // 注册静态资源函数
    funcMap := template.FuncMap{
        "static":    staticURL,
        "css":       cssURL,
        "js":        jsURL,
        "image":     imageURL,
        "asset":     assetURL,
        "versioned": versionedURL,
    }
    
    RegisterTemplateFuncs(funcMap)
}

func staticURL(path string) string {
    return "/static/" + strings.TrimPrefix(path, "/")
}

func cssURL(filename string) template.HTML {
    url := staticURL("css/" + filename)
    return template.HTML(fmt.Sprintf(`<link href="%s" rel="stylesheet">`, url))
}

func jsURL(filename string) template.HTML {
    url := staticURL("js/" + filename)
    return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, url))
}

func imageURL(filename string) string {
    return staticURL("images/" + filename)
}

func assetURL(filename string) string {
    // 支持版本化资源
    if config.AssetVersion != "" {
        return fmt.Sprintf("/static/%s?v=%s", filename, config.AssetVersion)
    }
    return staticURL(filename)
}

func versionedURL(filename string) string {
    // 基于文件修改时间的版本
    stat, err := os.Stat(path.Join(config.StaticPath, filename))
    if err != nil {
        return staticURL(filename)
    }
    
    version := stat.ModTime().Unix()
    return fmt.Sprintf("/static/%s?v=%d", filename, version)
}
```

### 使用静态资源函数

```html
<!-- 基本使用 -->
<img src="{{image "logo.png"}}" alt="Logo">

<!-- CSS 快捷方式 -->
{{css "app.css"}}
{{css "admin.css"}}

<!-- JavaScript 快捷方式 -->
{{js "jquery.min.js"}}
{{js "app.js"}}

<!-- 版本化资源 -->
<link href="{{versioned "css/app.css"}}" rel="stylesheet">
<script src="{{versioned "js/app.js"}}"></script>

<!-- 带版本的资源 -->
<img src="{{asset "images/banner.jpg"}}" alt="Banner">
```

## 资源打包

### 资源配置文件

```yaml
# assets.yaml
assets:
  css:
    app:
      - "css/bootstrap.min.css"
      - "css/app.css"
      - "css/components/*.css"
    admin:
      - "css/bootstrap.min.css"
      - "css/admin.css"
      - "css/admin/*.css"
  
  js:
    app:
      - "js/jquery.min.js"
      - "js/bootstrap.min.js"
      - "js/app.js"
      - "js/modules/*.js"
    admin:
      - "js/jquery.min.js"
      - "js/bootstrap.min.js"
      - "js/admin.js"
      - "js/admin/*.js"

compression:
  enabled: true
  gzip: true
  brotli: true
  
cache:
  enabled: true
  max_age: 31536000  # 1 year
  etag: true

cdn:
  enabled: false
  base_url: "https://cdn.example.com"
  fallback: true
```

### 资源打包器

```go
package assets

import (
    "compress/gzip"
    "crypto/md5"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

type AssetBundle struct {
    Name        string
    Type        string
    Files       []string
    OutputPath  string
    Compressed  bool
    Hash        string
}

type AssetManager struct {
    config     AssetConfig
    bundles    map[string]*AssetBundle
    staticPath string
    outputPath string
}

func NewAssetManager(config AssetConfig) *AssetManager {
    return &AssetManager{
        config:     config,
        bundles:    make(map[string]*AssetBundle),
        staticPath: config.StaticPath,
        outputPath: config.OutputPath,
    }
}

func (am *AssetManager) RegisterBundle(name, assetType string, files []string) {
    bundle := &AssetBundle{
        Name:  name,
        Type:  assetType,
        Files: files,
    }
    
    am.bundles[name] = bundle
}

func (am *AssetManager) BuildBundles() error {
    for name, bundle := range am.bundles {
        if err := am.buildBundle(bundle); err != nil {
            return fmt.Errorf("failed to build bundle %s: %w", name, err)
        }
    }
    
    return nil
}

func (am *AssetManager) buildBundle(bundle *AssetBundle) error {
    // 创建输出文件
    outputFile := fmt.Sprintf("%s.%s", bundle.Name, bundle.Type)
    outputPath := filepath.Join(am.outputPath, outputFile)
    
    output, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer output.Close()
    
    hasher := md5.New()
    multiWriter := io.MultiWriter(output, hasher)
    
    // 合并文件
    for _, file := range bundle.Files {
        files, err := am.expandGlob(file)
        if err != nil {
            return err
        }
        
        for _, f := range files {
            if err := am.appendFile(multiWriter, f); err != nil {
                return err
            }
        }
    }
    
    // 计算哈希
    bundle.Hash = fmt.Sprintf("%x", hasher.Sum(nil))[:8]
    bundle.OutputPath = outputPath
    
    // 创建带哈希的文件名
    if am.config.Versioning {
        versionedFile := fmt.Sprintf("%s.%s.%s", bundle.Name, bundle.Hash, bundle.Type)
        versionedPath := filepath.Join(am.outputPath, versionedFile)
        
        if err := am.copyFile(outputPath, versionedPath); err != nil {
            return err
        }
    }
    
    // 压缩文件
    if am.config.Compression {
        if err := am.compressFile(outputPath); err != nil {
            return err
        }
    }
    
    return nil
}

func (am *AssetManager) expandGlob(pattern string) ([]string, error) {
    fullPattern := filepath.Join(am.staticPath, pattern)
    return filepath.Glob(fullPattern)
}

func (am *AssetManager) appendFile(writer io.Writer, filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    _, err = io.Copy(writer, file)
    return err
}

func (am *AssetManager) compressFile(filename string) error {
    // Gzip 压缩
    if am.config.Gzip {
        if err := am.gzipFile(filename); err != nil {
            return err
        }
    }
    
    // Brotli 压缩
    if am.config.Brotli {
        if err := am.brotliFile(filename); err != nil {
            return err
        }
    }
    
    return nil
}

func (am *AssetManager) gzipFile(filename string) error {
    input, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer input.Close()
    
    output, err := os.Create(filename + ".gz")
    if err != nil {
        return err
    }
    defer output.Close()
    
    gzipWriter := gzip.NewWriter(output)
    defer gzipWriter.Close()
    
    _, err = io.Copy(gzipWriter, input)
    return err
}
```

### 在应用中使用打包资源

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/assets"
)

func main() {
    app := mvc.HertzApp
    
    // 加载资源配置
    assetConfig := assets.LoadConfig("assets.yaml")
    assetManager := assets.NewAssetManager(assetConfig)
    
    // 注册资源束
    assetManager.RegisterBundle("app", "css", []string{
        "css/bootstrap.min.css",
        "css/app.css",
    })
    
    assetManager.RegisterBundle("app", "js", []string{
        "js/jquery.min.js",
        "js/bootstrap.min.js",
        "js/app.js",
    })
    
    // 构建资源（在生产环境中）
    if !app.IsDebug() {
        if err := assetManager.BuildBundles(); err != nil {
            panic(err)
        }
    }
    
    // 注册资源管理器
    app.SetAssetManager(assetManager)
    
    app.Run()
}
```

## CDN 集成

### CDN 配置

```go
type CDNConfig struct {
    Enabled     bool   `yaml:"enabled"`
    BaseURL     string `yaml:"base_url"`
    Fallback    bool   `yaml:"fallback"`
    Domains     []string `yaml:"domains"`
    StaticTypes []string `yaml:"static_types"`
}

type CDNManager struct {
    config CDNConfig
    index  int
}

func NewCDNManager(config CDNConfig) *CDNManager {
    return &CDNManager{
        config: config,
    }
}

func (cm *CDNManager) GetURL(path string) string {
    if !cm.config.Enabled {
        return "/static/" + path
    }
    
    // 轮询使用多个CDN域名
    if len(cm.config.Domains) > 0 {
        domain := cm.config.Domains[cm.index%len(cm.config.Domains)]
        cm.index++
        return fmt.Sprintf("https://%s/%s", domain, path)
    }
    
    return cm.config.BaseURL + "/" + path
}

func (cm *CDNManager) ShouldUseCDN(path string) bool {
    if !cm.config.Enabled {
        return false
    }
    
    ext := filepath.Ext(path)
    for _, staticType := range cm.config.StaticTypes {
        if ext == staticType {
            return true
        }
    }
    
    return false
}
```

### CDN 资源函数

```go
func staticURLWithCDN(path string) string {
    cdnManager := GetCDNManager()
    
    if cdnManager.ShouldUseCDN(path) {
        return cdnManager.GetURL(path)
    }
    
    return "/static/" + path
}

func cdnURL(path string) template.HTML {
    url := staticURLWithCDN(path)
    return template.HTML(url)
}

func cdnFallback(path string) template.HTML {
    cdnURL := staticURLWithCDN(path)
    localURL := "/static/" + path
    
    switch filepath.Ext(path) {
    case ".css":
        return template.HTML(fmt.Sprintf(`
            <link href="%s" rel="stylesheet" onerror="this.onerror=null;this.href='%s'">
        `, cdnURL, localURL))
    
    case ".js":
        return template.HTML(fmt.Sprintf(`
            <script src="%s" onerror="var s=document.createElement('script');s.src='%s';document.head.appendChild(s)"></script>
        `, cdnURL, localURL))
    
    default:
        return template.HTML(cdnURL)
    }
}
```

## 缓存优化

### HTTP 缓存头

```go
package middleware

import (
    "path/filepath"
    "strconv"
    "strings"
    "time"
)

func StaticCacheMiddleware(config StaticCacheConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path
        
        // 检查是否是静态资源
        if !strings.HasPrefix(path, "/static/") {
            c.Next()
            return
        }
        
        ext := filepath.Ext(path)
        cacheConfig := config.GetCacheConfig(ext)
        
        if cacheConfig != nil {
            // 设置缓存头
            c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", cacheConfig.MaxAge))
            c.Header("Expires", time.Now().Add(time.Duration(cacheConfig.MaxAge)*time.Second).Format(http.TimeFormat))
            
            // ETag 支持
            if cacheConfig.ETag {
                etag := generateETag(path)
                c.Header("ETag", etag)
                
                // 检查 If-None-Match
                if match := c.GetHeader("If-None-Match"); match == etag {
                    c.Status(304)
                    return
                }
            }
            
            // Last-Modified 支持
            if cacheConfig.LastModified {
                if stat, err := os.Stat(getFilePath(path)); err == nil {
                    lastModified := stat.ModTime().Format(http.TimeFormat)
                    c.Header("Last-Modified", lastModified)
                    
                    // 检查 If-Modified-Since
                    if since := c.GetHeader("If-Modified-Since"); since != "" {
                        if sinceTime, err := time.Parse(http.TimeFormat, since); err == nil {
                            if !stat.ModTime().After(sinceTime) {
                                c.Status(304)
                                return
                            }
                        }
                    }
                }
            }
        }
        
        c.Next()
    }
}

type CacheConfig struct {
    MaxAge       int  `yaml:"max_age"`
    ETag         bool `yaml:"etag"`
    LastModified bool `yaml:"last_modified"`
}

type StaticCacheConfig struct {
    Configs map[string]*CacheConfig `yaml:"configs"`
}

func (scc *StaticCacheConfig) GetCacheConfig(ext string) *CacheConfig {
    if config, exists := scc.Configs[ext]; exists {
        return config
    }
    
    // 默认配置
    return scc.Configs["default"]
}
```

### 资源版本管理

```go
type VersionManager struct {
    versions map[string]string
    strategy VersionStrategy
}

type VersionStrategy string

const (
    VersionStrategyHash      VersionStrategy = "hash"
    VersionStrategyTimestamp VersionStrategy = "timestamp"
    VersionStrategyManual    VersionStrategy = "manual"
)

func NewVersionManager(strategy VersionStrategy) *VersionManager {
    return &VersionManager{
        versions: make(map[string]string),
        strategy: strategy,
    }
}

func (vm *VersionManager) GetVersion(path string) string {
    if version, exists := vm.versions[path]; exists {
        return version
    }
    
    switch vm.strategy {
    case VersionStrategyHash:
        return vm.generateHashVersion(path)
    case VersionStrategyTimestamp:
        return vm.generateTimestampVersion(path)
    case VersionStrategyManual:
        return vm.getManualVersion(path)
    default:
        return ""
    }
}

func (vm *VersionManager) generateHashVersion(path string) string {
    filePath := getFilePath(path)
    content, err := os.ReadFile(filePath)
    if err != nil {
        return ""
    }
    
    hash := md5.Sum(content)
    version := fmt.Sprintf("%x", hash)[:8]
    vm.versions[path] = version
    
    return version
}

func (vm *VersionManager) generateTimestampVersion(path string) string {
    filePath := getFilePath(path)
    stat, err := os.Stat(filePath)
    if err != nil {
        return ""
    }
    
    version := strconv.FormatInt(stat.ModTime().Unix(), 10)
    vm.versions[path] = version
    
    return version
}
```

## 开发环境优化

### 热重载支持

```go
package devtools

import (
    "path/filepath"
    "strings"
)

type AssetWatcher struct {
    config    AssetWatchConfig
    manager   *AssetManager
    clients   []chan string
    clientsMu sync.RWMutex
}

func NewAssetWatcher(config AssetWatchConfig, manager *AssetManager) *AssetWatcher {
    return &AssetWatcher{
        config:  config,
        manager: manager,
        clients: make([]chan string, 0),
    }
}

func (aw *AssetWatcher) Start() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    defer watcher.Close()
    
    // 监控静态资源目录
    for _, path := range aw.config.WatchPaths {
        watcher.Add(path)
    }
    
    debouncer := NewDebouncer(500 * time.Millisecond)
    
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                return nil
            }
            
            if aw.shouldReload(event.Name) {
                debouncer.Call(func() {
                    aw.handleAssetChange(event.Name)
                })
            }
            
        case err, ok := <-watcher.Errors:
            if !ok {
                return nil
            }
            log.Printf("Asset watcher error: %v", err)
        }
    }
}

func (aw *AssetWatcher) shouldReload(path string) bool {
    ext := filepath.Ext(path)
    for _, watchExt := range aw.config.Extensions {
        if ext == watchExt {
            return true
        }
    }
    return false
}

func (aw *AssetWatcher) handleAssetChange(path string) {
    log.Printf("Asset changed: %s", path)
    
    // 重新构建相关的资源束
    if err := aw.manager.RebuildAffectedBundles(path); err != nil {
        log.Printf("Failed to rebuild bundles: %v", err)
        return
    }
    
    // 通知浏览器刷新
    aw.notifyClients(path)
}

func (aw *AssetWatcher) notifyClients(path string) {
    message := fmt.Sprintf(`{"type":"reload","path":"%s"}`, path)
    
    aw.clientsMu.RLock()
    for _, client := range aw.clients {
        select {
        case client <- message:
        default:
            // 客户端缓冲区满，跳过
        }
    }
    aw.clientsMu.RUnlock()
}
```

## 最佳实践

### 1. 资源组织

```
static/
├── dist/                 # 构建后的资源
│   ├── css/
│   ├── js/
│   └── images/
├── src/                  # 源文件
│   ├── scss/
│   ├── js/
│   └── images/
├── vendor/               # 第三方库
│   ├── bootstrap/
│   ├── jquery/
│   └── fontawesome/
└── uploads/              # 用户上传文件
    ├── images/
    └── documents/
```

### 2. 性能优化

```go
// 资源预加载
func preloadAssets(c *gin.Context) {
    criticalCSS := []string{
        "css/critical.css",
        "css/app.css",
    }
    
    for _, css := range criticalCSS {
        c.Header("Link", fmt.Sprintf("<%s>; rel=preload; as=style", staticURL(css)))
    }
    
    criticalJS := []string{
        "js/critical.js",
    }
    
    for _, js := range criticalJS {
        c.Header("Link", fmt.Sprintf("<%s>; rel=preload; as=script", staticURL(js)))
    }
}

// 资源压缩
func enableCompression() gin.HandlerFunc {
    return gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{
        "/api/",  // API 响应不压缩
    }))
}
```

### 3. 安全配置

```go
// 静态资源安全头
func securityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        if strings.HasPrefix(c.Request.URL.Path, "/static/") {
            // 防止 MIME 类型嗅探
            c.Header("X-Content-Type-Options", "nosniff")
            
            // 防止在 frame 中嵌入
            c.Header("X-Frame-Options", "DENY")
            
            // CSP 头
            c.Header("Content-Security-Policy", "default-src 'self'")
        }
        
        c.Next()
    }
}

// 文件上传安全
func validateUpload(file *multipart.FileHeader) error {
    // 检查文件大小
    if file.Size > maxFileSize {
        return errors.New("file too large")
    }
    
    // 检查文件类型
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "image/gif":  true,
    }
    
    contentType := file.Header.Get("Content-Type")
    if !allowedTypes[contentType] {
        return errors.New("file type not allowed")
    }
    
    return nil
}
```

YYHertz 的静态资源管理系统提供了完整的解决方案，从基本的文件服务到高级的 CDN 集成和性能优化，帮助开发者构建高效的 Web 应用程序。
