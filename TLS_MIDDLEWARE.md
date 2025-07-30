# TLS支持中间件文档

## 概述

TLSSupportMiddleware 是一个为 Hertz MVC 框架提供 TLS/HTTPS 安全支持的中间件。它提供了完整的 HTTPS 重定向、安全头设置、HSTS 支持等功能。

## 特性

### 🔒 核心安全功能
- **HTTPS 强制**: 可配置强制要求 HTTPS 连接
- **自动重定向**: HTTP 请求自动重定向到 HTTPS
- **HSTS 支持**: HTTP 严格传输安全头设置
- **安全响应头**: 自动设置多种安全相关的 HTTP 头

### 🛡️ 安全头支持
- `Strict-Transport-Security`: HSTS 安全传输
- `X-Content-Type-Options`: 防止 MIME 类型嗅探
- `X-Frame-Options`: 防止点击劫持
- `X-XSS-Protection`: XSS 保护
- `Referrer-Policy`: 引用来源策略
- `Content-Security-Policy`: 内容安全策略

### 🔧 代理环境支持
- `X-Forwarded-Proto`: 代理协议检测
- `X-Forwarded-SSL`: SSL 状态检测
- `Front-End-Https`: 前端 HTTPS 检测

## 使用方法

### 基本使用

```go
import "github.com/zsy619/yyhertz/framework/middleware"

// 使用默认配置
app.Use(middleware.TLSSupportMiddleware(nil))

// 或者使用默认配置
tlsConfig := middleware.DefaultTLSConfig()
app.Use(middleware.TLSSupportMiddleware(tlsConfig))
```

### 自定义配置

```go
tlsConfig := &middleware.TLSConfig{
    Enable:         true,
    RequireHTTPS:   true,
    HSTSEnabled:    true,
    HSTSMaxAge:     31536000, // 1年
    HSTSSubdomains: true,
    HTTPSRedirect:  true,
    RedirectPort:   443,
}
app.Use(middleware.TLSSupportMiddleware(tlsConfig))
```

### 开发环境配置

```go
// 开发环境 - 不强制 HTTPS
devConfig := &middleware.TLSConfig{
    Enable:         false,
    RequireHTTPS:   false,
    HSTSEnabled:    false,
    HTTPSRedirect:  false,
}
app.Use(middleware.TLSSupportMiddleware(devConfig))
```

### 生产环境配置

```go
// 生产环境 - 强制 HTTPS
prodConfig := &middleware.TLSConfig{
    Enable:         true,
    CertFile:       "/etc/ssl/certs/server.crt",
    KeyFile:        "/etc/ssl/private/server.key",
    RequireHTTPS:   true,
    HSTSEnabled:    true,
    HSTSMaxAge:     31536000,
    HSTSSubdomains: true,
    HTTPSRedirect:  true,
    RedirectPort:   443,
}
app.Use(middleware.TLSSupportMiddleware(prodConfig))
```

## 配置选项

### TLSConfig 结构

```go
type TLSConfig struct {
    // 基础配置
    Enable         bool     // 是否启用TLS
    CertFile       string   // 证书文件路径
    KeyFile        string   // 私钥文件路径
    MinVersion     uint16   // 最小TLS版本
    MaxVersion     uint16   // 最大TLS版本
    
    // 安全配置
    RequireHTTPS   bool     // 是否强制HTTPS
    HSTSEnabled    bool     // 是否启用HSTS
    HSTSMaxAge     int      // HSTS最大年龄（秒）
    HSTSSubdomains bool     // HSTS是否包含子域名
    
    // 密码套件配置
    CipherSuites   []uint16 // 支持的密码套件
    PreferServer   bool     // 是否优先服务器密码套件
    
    // 客户端证书配置
    ClientAuth     tls.ClientAuthType // 客户端认证模式
    ClientCAFile   string             // 客户端CA证书文件
    
    // 重定向配置
    HTTPSRedirect  bool     // HTTP是否重定向到HTTPS
    RedirectPort   int      // HTTPS重定向端口
}
```

### 默认配置值

```go
Enable:         false
MinVersion:     tls.VersionTLS12
MaxVersion:     tls.VersionTLS13
RequireHTTPS:   false
HSTSEnabled:    true
HSTSMaxAge:     31536000 // 1年
HSTSSubdomains: true
HTTPSRedirect:  true
RedirectPort:   443
PreferServer:   true
ClientAuth:     tls.NoClientCert
```

## 命令行集成

在 `version.go` 中已集成命令行参数支持：

```bash
# 启用 HTTPS
./main --https --cert=/path/to/cert.pem --key=/path/to/key.pem

# 强制要求 HTTPS
./main --require-https

# 完整示例
./main --https --cert=/etc/ssl/certs/server.crt --key=/etc/ssl/private/server.key --require-https --port=8443
```

### 支持的命令行参数

- `--https`: 启用 HTTPS
- `--cert string`: TLS 证书文件路径
- `--key string`: TLS 私钥文件路径
- `--require-https`: 强制要求 HTTPS 连接

## TLS 配置管理

### TLS 服务器配置

```go
import "github.com/zsy619/yyhertz/framework/config"

// 创建 TLS 管理器
serverConfig := config.DefaultTLSServerConfig()
serverConfig.Enable = true
serverConfig.CertFile = "/path/to/cert.pem"
serverConfig.KeyFile = "/path/to/key.pem"
serverConfig.AutoReload = true

tlsManager, err := config.NewTLSManager(serverConfig)
if err != nil {
    log.Fatal("TLS管理器创建失败:", err)
}

// 获取 TLS 配置
tlsConfig := tlsManager.GetTLSConfig()
```

### 证书自动重载

```go
serverConfig := &config.TLSServerConfig{
    Enable:         true,
    CertFile:       "/path/to/cert.pem",
    KeyFile:        "/path/to/key.pem",
    AutoReload:     true,
    ReloadInterval: 300, // 5分钟检查一次
}
```

## 安全响应头详解

### HSTS (HTTP Strict Transport Security)

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

- 强制浏览器使用 HTTPS
- 防止协议降级攻击
- 支持子域名包含

### 内容安全策略 (CSP)

```
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; ...
```

- 防止 XSS 攻击
- 控制资源加载来源
- 升级不安全请求

### 其他安全头

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

## 错误处理

### 配置验证

中间件提供配置验证功能：

```go
if err := middleware.ValidateTLSConfig(tlsConfig); err != nil {
    log.Fatal("TLS配置验证失败:", err)
}
```

### 常见错误

1. **证书文件缺失**
   ```
   validation error in field 'cert_file': TLS证书文件路径不能为空
   ```

2. **私钥文件缺失**
   ```
   validation error in field 'key_file': TLS私钥文件路径不能为空
   ```

3. **TLS版本配置错误**
   ```
   validation error in field 'tls_version': 最小TLS版本不能大于最大TLS版本
   ```

## 日志记录

中间件使用结构化日志记录所有操作：

```json
{
  "level": "info",
  "msg": "TLS中间件处理开始",
  "path": "/api/users",
  "method": "GET",
  "tls_enabled": true,
  "time": "2025-07-29T22:55:03+08:00"
}
```

### 日志字段说明

- `path`: 请求路径
- `method`: HTTP 方法
- `tls_enabled`: TLS 是否启用
- `from`/`to`: 重定向源和目标URL
- `cert_file`/`key_file`: 证书文件路径

## 测试

运行 TLS 中间件测试：

```bash
go test ./framework/middleware -v
```

### 测试覆盖

- ✅ 默认配置测试
- ✅ 配置验证测试
- ✅ HTTPS 检测测试
- ✅ 安全头设置测试

## 最佳实践

### 1. 生产环境配置

```go
// 生产环境必须启用的配置
tlsConfig := &middleware.TLSConfig{
    RequireHTTPS:   true,    // 强制 HTTPS
    HSTSEnabled:    true,    // 启用 HSTS
    HSTSMaxAge:     31536000, // 1年有效期
    HSTSSubdomains: true,    // 包含子域名
    HTTPSRedirect:  true,    // 自动重定向
}
```

### 2. 开发环境配置

```go
// 开发环境灵活配置
tlsConfig := &middleware.TLSConfig{
    RequireHTTPS:  false,   // 允许 HTTP
    HSTSEnabled:   false,   // 禁用 HSTS
    HTTPSRedirect: false,   // 禁用重定向
}
```

### 3. 证书管理

- 使用自动续期的证书（如 Let's Encrypt）
- 启用证书自动重载功能
- 定期检查证书有效期
- 使用强加密套件

### 4. 监控和日志

- 监控证书过期时间
- 记录 HTTPS 重定向次数
- 监控 TLS 握手错误
- 定期审查安全头配置

## 常见问题

### Q: 如何在负载均衡器后使用？

A: 配置代理头检测：

```go
// 中间件会自动检测以下头：
// X-Forwarded-Proto: https
// X-Forwarded-SSL: on
// Front-End-Https: on
```

### Q: 如何自定义 CSP 策略？

A: 目前使用默认策略，可以通过修改 `setSecurityHeaders` 函数自定义。

### Q: 证书自动重载如何工作？

A: TLS 管理器会定期检查证书文件变化并自动重载，无需重启服务。

## 相关文件

- `framework/middleware/tls.go` - TLS 中间件实现
- `framework/config/tls_config.go` - TLS 配置管理
- `framework/types/constants.go` - 错误代码定义
- `version.go` - 命令行集成示例