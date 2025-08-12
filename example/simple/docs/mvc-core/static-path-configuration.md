# 静态路径配置 - SetStaticPath

YYHertz 框架提供了灵活的静态资源路径配置功能，通过 `SetStaticPath` 方法可以轻松配置多个静态目录到不同的URL路径映射。

## 🎯 API 概览

### 函数签名
```go
func SetStaticPath(localDir string, urlPath ...string)
```

### 参数说明
- `localDir`: 静态文件本地目录（相对应用所在目录）
- `urlPath`: URL路径（可选），如果不提供则自动推导

## 🚀 基本用法

### 双参数形式（推荐）
明确指定本地目录和URL路径的映射关系：

```go
package main

import "github.com/zsy619/yyhertz/framework/mvc"

func main() {
    // 设置 public 目录映射到 /static URL路径
    mvc.SetStaticPath("public", "/static")
    
    // 设置 uploads 目录映射到 /files URL路径
    mvc.SetStaticPath("uploads", "/files")
    
    // 设置 assets 目录映射到 /assets URL路径
    mvc.SetStaticPath("assets", "/assets")
    
    mvc.HertzApp.Run(":8080")
}
```

**访问示例：**
- 本地文件: `public/css/app.css` → 访问URL: `http://localhost:8080/static/css/app.css`
- 本地文件: `uploads/avatar.jpg` → 访问URL: `http://localhost:8080/files/avatar.jpg`
- 本地文件: `assets/logo.png` → 访问URL: `http://localhost:8080/assets/logo.png`

### 单参数形式（自动推导）
只提供本地目录，URL路径会自动推导：

```go
func main() {
    // 自动推导URL路径
    mvc.SetStaticPath("images")      // 自动映射到 /images
    mvc.SetStaticPath("css")         // 自动映射到 /css
    mvc.SetStaticPath("js")          // 自动映射到 /js
    mvc.SetStaticPath("documents")   // 自动映射到 /documents
    
    mvc.HertzApp.Run(":8080")
}
```

**访问示例：**
- 本地文件: `images/banner.jpg` → 访问URL: `http://localhost:8080/images/banner.jpg`
- 本地文件: `css/style.css` → 访问URL: `http://localhost:8080/css/style.css`

## 🎨 高级用法

### 相对路径处理
支持带 `./` 前缀的相对路径：

```go
func main() {
    // 支持相对路径前缀
    mvc.SetStaticPath("./static/vendor", "/vendor")
    mvc.SetStaticPath("./uploads", "/files")
    
    // 推荐不带前缀的写法
    mvc.SetStaticPath("static/vendor", "/vendor")
    mvc.SetStaticPath("uploads", "/files")
}
```

### 复杂项目配置
适用于大型项目的多目录配置：

```go
func setupStaticPaths() {
    // 主要静态资源
    mvc.SetStaticPath("public", "/static")
    
    // 用户上传内容
    mvc.SetStaticPath("uploads", "/files")
    mvc.SetStaticPath("temp", "/temp")
    
    // 开发和构建资源
    mvc.SetStaticPath("src", "/src")       // 开发环境源文件
    mvc.SetStaticPath("dist", "/assets")   // 生产环境构建文件
    
    // 文档和下载
    mvc.SetStaticPath("documents", "/docs")
    mvc.SetStaticPath("downloads", "/dl")
    
    // 第三方库和插件
    mvc.SetStaticPath("vendor", "/vendor")
    mvc.SetStaticPath("plugins", "/plugins")
    
    // 自动推导的目录
    mvc.SetStaticPath("images")    // -> /images
    mvc.SetStaticPath("fonts")     // -> /fonts
    mvc.SetStaticPath("videos")    // -> /videos
}

func main() {
    setupStaticPaths()
    mvc.HertzApp.Run(":8080")
}
```

## 📁 推荐目录结构

### 标准Web应用结构
```
my-hertz-app/
├── public/                 # 主要静态资源 → /static
│   ├── css/
│   ├── js/
│   ├── images/
│   ├── fonts/
│   └── vendor/
├── uploads/                # 用户上传 → /files  
│   ├── avatars/
│   ├── documents/
│   └── gallery/
├── assets/                 # 编译资源 → /assets
│   ├── app.min.css
│   ├── app.min.js
│   └── manifest.json
└── docs/                  # 文档资源 → /docs
    ├── api/
    └── guides/
```

**对应配置：**
```go
func main() {
    mvc.SetStaticPath("public", "/static")
    mvc.SetStaticPath("uploads", "/files") 
    mvc.SetStaticPath("assets")            // 自动推导为 /assets
    mvc.SetStaticPath("docs")              // 自动推导为 /docs
    
    mvc.HertzApp.Run(":8080")
}
```

### 微服务架构结构
```
microservice-app/
├── web-assets/             # Web前端资源 → /web
│   ├── app/
│   ├── admin/
│   └── mobile/
├── api-docs/              # API文档 → /api-docs
│   ├── v1/
│   └── v2/
├── cdn-cache/             # CDN缓存 → /cache
└── shared/                # 共享资源 → /shared
    ├── images/
    └── icons/
```

**对应配置：**
```go
func main() {
    mvc.SetStaticPath("web-assets", "/web")
    mvc.SetStaticPath("api-docs")           // 自动推导为 /api-docs  
    mvc.SetStaticPath("cdn-cache", "/cache")
    mvc.SetStaticPath("shared")             // 自动推导为 /shared
    
    mvc.HertzApp.Run(":8080")
}
```

## 🔧 路径映射规则

### 自动推导规则
当只提供 `localDir` 参数时，URL路径按以下规则推导：

| 本地目录 | 自动推导URL路径 | 说明 |
|---------|---------------|------|
| `"assets"` | `/assets` | 直接映射 |
| `"./images"` | `/images` | 移除 `./` 前缀 |
| `"static/css"` | `/static/css` | 保持路径结构 |
| `""` | `/static` | 空目录使用默认值 |

### 路径处理逻辑
```go
// 伪代码展示路径处理逻辑
func processPath(localDir string, urlPath ...string) (string, string) {
    var targetUrlPath string
    
    if len(urlPath) > 0 && urlPath[0] != "" {
        // 明确指定URL路径
        targetUrlPath = urlPath[0]
    } else {
        // 自动推导URL路径
        cleanDir := strings.TrimLeft(strings.TrimPrefix(localDir, "./"), "/")
        if cleanDir == "" {
            targetUrlPath = "/static"  // 默认值
        } else {
            targetUrlPath = "/" + cleanDir
        }
    }
    
    // 确保URL路径以 / 开头
    if !strings.HasPrefix(targetUrlPath, "/") {
        targetUrlPath = "/" + targetUrlPath
    }
    
    return localDir, targetUrlPath
}
```

## 🔍 实际使用示例

### 博客系统
```go
func setupBlogStaticPaths() {
    // 主题资源
    mvc.SetStaticPath("themes/default", "/theme")
    mvc.SetStaticPath("themes/admin", "/admin-theme")
    
    // 内容资源
    mvc.SetStaticPath("uploads", "/uploads")    // 文章图片
    mvc.SetStaticPath("attachments", "/files")  // 文件下载
    
    // 缓存资源
    mvc.SetStaticPath("cache/thumbnails", "/thumbs")
    
    // 插件资源
    mvc.SetStaticPath("plugins")  // 自动推导为 /plugins
}
```

### 电商系统
```go
func setupEcommerceStaticPaths() {
    // 产品图片
    mvc.SetStaticPath("products/images", "/product-images")
    mvc.SetStaticPath("products/videos", "/product-videos")
    
    // 用户内容
    mvc.SetStaticPath("users/avatars", "/avatars")
    mvc.SetStaticPath("users/uploads", "/user-files")
    
    // 营销素材
    mvc.SetStaticPath("marketing/banners", "/banners")
    mvc.SetStaticPath("marketing/promotions", "/promo")
    
    // 第三方集成
    mvc.SetStaticPath("integrations", "/integrations")
}
```

### 开发环境配置
```go
func setupDevelopmentPaths() {
    if isProduction() {
        // 生产环境：使用构建后的资源
        mvc.SetStaticPath("dist", "/static")
        mvc.SetStaticPath("uploads", "/files")
    } else {
        // 开发环境：使用源文件
        mvc.SetStaticPath("src", "/static")
        mvc.SetStaticPath("dev-uploads", "/files")
        
        // 开发工具
        mvc.SetStaticPath("dev-tools", "/dev")
    }
}
```

## ⚡ 性能优化建议

### 1. 合理的目录结构
```go
// ✅ 推荐：按功能分类
mvc.SetStaticPath("css", "/css")
mvc.SetStaticPath("js", "/js") 
mvc.SetStaticPath("images", "/images")

// ❌ 避免：过于复杂的嵌套
mvc.SetStaticPath("assets/frontend/user/css", "/user-css")
```

### 2. 缓存友好的路径
```go
// ✅ 推荐：版本化资源路径
mvc.SetStaticPath("dist/v2.1.0", "/assets")

// ✅ 推荐：CDN友好的路径
mvc.SetStaticPath("cdn-ready", "/static")
```

### 3. 安全考虑
```go
// ✅ 推荐：不暴露敏感目录
mvc.SetStaticPath("public", "/static")      // 仅暴露public目录

// ❌ 避免：暴露整个项目根目录
mvc.SetStaticPath(".", "/")                 // 危险：暴露所有文件
```

## 🔗 与其他功能的集成

### 与模板引擎集成
```html
<!-- 模板中引用静态资源 -->
<link href="/static/css/app.css" rel="stylesheet">
<script src="/assets/app.min.js"></script>
<img src="/images/logo.png" alt="Logo">
```

### 与中间件集成  
```go
func main() {
    // 配置静态路径
    mvc.SetStaticPath("public", "/static")
    
    // 添加静态文件中间件（自动处理缓存、压缩等）
    mvc.HertzApp.Use(middleware.StaticCache())
    
    mvc.HertzApp.Run(":8080")
}
```

### 与配置文件集成
```yaml
# config/static.yaml
static_paths:
  public: "/static"
  uploads: "/files"
  assets: "/assets"
  docs: "/docs"
```

```go
func loadStaticConfig() {
    config := loadConfig("config/static.yaml")
    
    for localDir, urlPath := range config.StaticPaths {
        mvc.SetStaticPath(localDir, urlPath)
    }
}
```

---

**💡 提示**: `SetStaticPath` 是 YYHertz 框架中配置静态资源的推荐方式，它提供了灵活性和易用性的完美平衡，支持从简单的单目录配置到复杂的多目录映射需求。