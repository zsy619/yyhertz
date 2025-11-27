# YYHertz Session 模块

## 概述

YYHertz Session 模块提供了统一、安全、高性能的会话管理功能。通过架构重构优化，session功能现在专注于会话管理，而Cookie功能已独立为 `@framework/mvc/cookie` 包，提供更好的模块化和维护性。

## 🚀 主要特性

### ✨ 核心功能
- **统一的API接口** - 兼容Beego Session API，零学习成本
- **多种Context支持** - 同时支持Hertz和YYHertz Context
- **高性能设计** - 代理模式实现，几乎零性能开销
- **完全向后兼容** - 保持100%向后兼容性
- **灵活存储** - 支持内存、Redis等多种存储后端

### 🔐 安全特性
- 安全的Session ID生成
- Session劫持保护
- 可配置的安全选项
- 自动过期管理

### ⚡ 性能优化
- 延迟初始化减少内存开销
- 代理模式性能损失 < 10%
- 并发安全的设计
- 内存泄漏检测通过

## 📁 模块架构

```
framework/mvc/session/
├── session_adapter.go     # Session适配器
├── context_extension.go   # Context扩展（现已整合Cookie功能）
├── manager.go            # Session管理器
├── store.go              # Session存储接口
├── config.go             # 配置管理
├── demo.go               # 使用示例和Cookie迁移演示
├── session_test.go       # 功能测试
├── benchmark_test.go     # 性能基准测试
├── production_compatibility_test.go  # 生产环境测试
└── README.md             # 本文档
```

> **注意**: Cookie相关功能已迁移至 `@framework/mvc/cookie` 包。Session包中的Context扩展提供了向后兼容的Cookie接口。

## 🔧 快速开始

### 基础Session操作

```go
import "github.com/zsy619/yyhertz/framework/mvc/session"

// 创建Context扩展
ctx := // 你的RequestContext
extension := session.NewExtensionForHertzContext(ctx)

// 设置Session
err := extension.SetSession("user_id", "12345")
err = extension.SetSession("username", "testuser")

// 获取Session
userID := extension.GetSession("user_id")
username := extension.GetSession("username")

// 删除Session
err = extension.DelSession("user_id")

// 检查Session是否存在
exists := extension.GetSession("user_id") != nil
```

### 高级Session操作

```go
// 清空所有Session
extension.ClearSession()

// 销毁Session
extension.DestroySession()

// 重新生成Session ID
extension.SessionRegenerateID()

// 获取Session ID
sessionID := extension.GetSessionID()

// 检查Session是否已启动
started := extension.IsSessionStarted()
```

### 高级用法 - Session适配器

```go
// 获取Session适配器进行批量操作
adapter := extension.StartSession()

// 批量设置
adapter.Set("key1", "value1")
adapter.Set("key2", "value2")

// 批量获取
value1 := adapter.Get("key1")
value2 := adapter.Get("key2")

// 检查数据存在
exists := adapter.Has("key1")

// 删除特定数据
adapter.Delete("key1")

// 保存Session（通常自动完成）
adapter.Save()

// 获取Session ID
sessionID := adapter.GetSessionID()
```

## 🔄 迁移指南

### 从旧版本迁移

如果你之前使用的是 `framework/mvc/context` 中的session功能，**无需修改任何代码**！新版本通过代理模式保持100%向后兼容。

```go
// 这些Session代码继续正常工作
inputData.SetSession("key", "value")
inputData.GetSession("key")
inputData.DelSession("key")
// ... 等等
```

### Cookie功能迁移

Cookie相关功能已迁移到独立的 `@framework/mvc/cookie` 包：

```go
// 旧用法（仍然有效，通过代理）
inputData.SetCookie("key", "value")
inputData.Cookie("key")

// 新推荐用法
import "github.com/zsy619/yyhertz/framework/mvc/cookie"
helper := cookie.NewHelper(cookie.DefaultConfig())
helper.Set(ctx, "key", "value")
value := helper.Get(ctx, "key")
```

### 推荐的新用法

虽然旧API继续工作，但推荐使用新的统一接口：

```go
// 旧方式（仍然支持）
value := inputData.Cookie("key")

// 新方式（推荐）
extension := session.NewExtensionForHertzContext(ctx)
value := extension.GetCookie("key")
```

### 从Beego迁移

对于从Beego迁移的项目，API几乎一致：

```go
// Beego风格的API完全支持
extension.SetCookie("name", "value", maxAge, path, domain, secure, httpOnly)
value := extension.GetCookie("name")
extension.DelCookie("name")

// Session API也保持一致
extension.SetSession("key", "value")
value := extension.GetSession("key")
```

## 🎯 最佳实践

### 1. 安全Cookie使用

```go
// ✅ 正确：使用强密钥和适当的选项
options := session.CookieSecurityOptions{
    Secret:         generateSecureSecret(), // 至少32字节
    MaxAge:         time.Hour * 2,          // 适当的过期时间
    ValidateExpiry: true,                   // 启用过期验证
    RequireHTTPS:   true,                   // 生产环境要求HTTPS
}

// ❌ 错误：使用弱密钥
options := session.CookieSecurityOptions{
    Secret: "123456", // 太弱
}
```

### 2. Session管理

```go
// ✅ 正确：及时清理Session
defer func() {
    if shouldLogout {
        extension.DestroySession()
    }
}()

// ✅ 正确：定期重新生成Session ID
if shouldRegenerateID {
    extension.SessionRegenerateID()
}
```

### 3. 错误处理

```go
// ✅ 正确：检查错误
if err := extension.SetSession("key", "value"); err != nil {
    log.Printf("Session设置失败: %v", err)
    // 处理错误
}

// ✅ 正确：验证安全Cookie
if token, ok := extension.GetSecureCookie(secret, "csrf"); !ok {
    return errors.New("CSRF token无效")
}
```

### 4. 性能优化

```go
// ✅ 正确：复用Context扩展
extension := session.NewExtensionForHertzContext(ctx)
// 多次使用extension进行操作

// ❌ 错误：重复创建
for i := range items {
    ext := session.NewExtensionForHertzContext(ctx) // 不要在循环中创建
}
```

## 📊 性能数据

基于基准测试的性能数据：

### Cookie操作性能
- **基础Cookie获取**: ~15 ns/op, 0 allocs
- **基础Cookie设置**: ~4000 ns/op, 16 allocs
- **安全Cookie设置**: ~4200 ns/op, 16 allocs
- **安全Cookie获取**: ~10 ns/op, 0 allocs

### 代理层开销
- **直接调用**: 224.5 ns/op
- **代理调用**: 221.9 ns/op
- **性能损失**: < 2% (几乎可以忽略)

### 并发性能
- **高并发Cookie操作**: 10,000+ operations/goroutine 无错误
- **高并发Session操作**: 25,000+ operations/goroutine 无错误
- **内存使用**: 10,000次操作仅增长 0.01MB

## 🧪 测试覆盖

- ✅ **功能测试**: 12个测试函数，覆盖所有核心功能
- ✅ **兼容性测试**: 6个测试函数，确保向后兼容
- ✅ **性能基准测试**: 15个基准测试，覆盖各种场景
- ✅ **生产环境测试**: 高并发、内存泄漏、长时间运行测试
- ✅ **安全测试**: 错误恢复、边界条件测试

## 🔍 调试和监控

### 启用调试日志

```go
// 在配置中启用调试
config := session.DefaultConfig()
config.EnableDebug = true
manager := session.NewSessionManager(config)
```

### 性能监控

```go
// 运行性能基准测试
go test -bench=. -benchmem

// 内存泄漏检测
go test -run TestMemoryLeak -memprofile=mem.prof
go tool pprof mem.prof
```

### 竞态条件检测

```go
// 并发安全测试
go test -race -run TestHighConcurrency
```

## 🛠️ 配置选项

### Session配置

```yaml
# session.yaml
session:
  provider: "memory"           # memory, redis, database
  cookie_name: "YYHERTZ_SESSID"
  cookie_lifetime: 3600        # 秒
  gc_maxlifetime: 3600         # 垃圾回收间隔
  cookie_secure: true          # HTTPS only
  cookie_httponly: true        # 防止XSS
  cookie_samesite: "Strict"    # CSRF防护
```

### Cookie安全配置

```go
options := session.CookieSecurityOptions{
    Secret:         os.Getenv("COOKIE_SECRET"),
    MaxAge:         time.Hour * 24,
    ValidateExpiry: true,
    RequireHTTPS:   true,
}
```

## 🚨 常见问题

### Q: 重构后性能有影响吗？
A: 几乎没有。代理模式的性能损失 < 2%，在实际应用中可以忽略。

### Q: 需要修改现有代码吗？
A: 不需要。保持100%向后兼容，现有代码无需修改。

### Q: 如何启用安全Cookie？
A: 使用 `SetSecureCookie` 方法并提供HMAC密钥即可。

### Q: 支持分布式Session吗？
A: 是的，通过SessionManager支持Redis、数据库等分布式存储。

### Q: 如何处理Session过期？
A: 框架自动处理Session过期，也可以手动调用 `DestroySession()`。

## 🔗 相关链接

- [API文档](./demo.go) - 完整的API使用示例
- [性能测试](./benchmark_test.go) - 性能基准测试
- [迁移指南](../MIGRATION_GUIDE.md) - 详细迁移说明
- [安全最佳实践](./security.go) - 安全功能实现

## 📄 许可证

本项目采用与YYHertz框架相同的许可证。

---

**注意**: 本文档反映了重构后的新架构。如有疑问，请参考示例代码或联系维护团队。