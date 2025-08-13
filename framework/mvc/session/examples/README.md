# YYHertz Session模块示例

本目录包含了YYHertz Session & Cookie模块的完整使用示例，帮助开发者快速理解和使用新的Session架构。

## 📁 示例文件

### 1. complete_example.go
**完整的Web应用示例**

演示了一个完整的Web应用场景，包括：
- 用户认证（登录/注销）
- 购物车功能
- 用户设置管理
- 安全Cookie使用
- CSRF防护
- API端点

```bash
# 运行示例
go run complete_example.go

# 访问端点
curl http://localhost:8080/
curl -X POST http://localhost:8080/auth/login -d "username=demo&password=password"
curl http://localhost:8080/profile -H "Cookie: YYHERTZ_SESSID=your_session_id"
```

**特性展示:**
- ✅ 用户认证和Session管理
- ✅ 购物车Session存储
- ✅ 安全Cookie与CSRF防护
- ✅ 用户偏好设置
- ✅ 完整的错误处理

### 2. migration_example.go
**迁移策略示例**

展示了如何从旧版本无缝迁移到新架构：
- 传统方式（100%兼容）
- 新方式（推荐用法）
- 混合方式（渐进迁移）
- 兼容性验证

```bash
# 运行迁移示例
go run migration_example.go

# 测试不同的使用方式
curl http://localhost:8081/legacy        # 传统方式
curl http://localhost:8081/new           # 新方式
curl http://localhost:8081/mixed         # 混合方式
curl http://localhost:8081/compatibility # 兼容性测试
```

**迁移策略:**
- 🔄 **立即升级**: 现有代码零修改
- 🔄 **渐进迁移**: 新功能使用新API
- 🔄 **全面迁移**: 完全使用新架构

### 3. performance_comparison.go
**性能对比测试**

提供详细的性能基准测试，对比重构前后的性能表现：
- 直接调用vs代理调用性能
- 批量操作性能
- 安全Cookie性能
- 内存使用分析

```bash
# 运行性能测试
go run performance_comparison.go

# 输出示例：
# 测试类型     耗时          操作数/秒        平均延迟      内存分配       对象数
# 直接调用     125ms         320000          312ns        1.2MB         1500
# 代理调用     128ms         312500          320ns        1.3MB         1600
```

**性能指标:**
- 📊 代理开销 < 2%
- 📊 内存效率优秀
- 📊 并发性能良好

## 🚀 快速开始

### 环境要求
- Go 1.18+
- YYHertz框架 v1.4.0+

### 运行所有示例

```bash
# 克隆或下载示例文件
cd examples/

# 运行完整示例（端口8080）
go run complete_example.go &

# 运行迁移示例（端口8081）
go run migration_example.go &

# 运行性能测试
go run performance_comparison.go

# 停止示例服务器
pkill -f "complete_example\|migration_example"
```

### 环境变量配置

```bash
# 完整示例配置
export HMAC_SECRET="your-32-byte-secret-key-here"
export SERVER_PORT="8080"

# 性能测试配置
export PERF_ITERATIONS="10000"
export PERF_WARMUP="100"
```

## 📖 使用指南

### 1. 新项目推荐用法

```go
import "github.com/zsy619/yyhertz/framework/mvc/session"

func handleRequest(ctx context.Context, c *app.RequestContext) {
    // 创建session扩展
    extension := session.NewExtensionForHertzContext(c)
    
    // Cookie操作
    extension.SetCookie("theme", "dark")
    theme := extension.GetCookie("theme")
    
    // Session操作
    extension.SetSession("user_id", "12345")
    userID := extension.GetSession("user_id")
    
    // 安全Cookie
    secret := "your-hmac-secret"
    extension.SetSecureCookie(secret, "csrf", "token")
    token, ok := extension.GetSecureCookie(secret, "csrf")
}
```

### 2. 现有项目升级

```go
// 现有代码无需修改，继续正常工作
inputData.Cookie("key")
inputData.SetCookie("key", "value")
inputData.SetSession("key", "value")
outputData.SetCookie("key", "value")

// 新功能可以逐步引入
extension := session.NewExtensionForHertzContext(ctx)
extension.SetSecureCookie("secret", "secure_key", "secure_value")
```

### 3. 高级功能使用

```go
// 批量Session操作
adapter := extension.StartSession()
adapter.Set("key1", "value1")
adapter.Set("key2", "value2")
adapter.Save()

// 安全Cookie配置
options := session.CookieSecurityOptions{
    Secret:         "your-secret",
    MaxAge:         time.Hour * 24,
    ValidateExpiry: true,
    RequireHTTPS:   true,
}
extension.SecureCookie.SetSecureWithOptions("secure_data", "value", options)
```

## 🧪 测试场景

### 用户认证流程测试

```bash
# 1. 用户登录
curl -X POST http://localhost:8080/auth/login \
  -d "username=demo&password=password&remember_me=true"

# 2. 获取用户资料
curl http://localhost:8080/profile \
  -H "Cookie: YYHERTZ_SESSID=returned_session_id"

# 3. 用户注销
curl -X POST http://localhost:8080/auth/logout \
  -H "Cookie: YYHERTZ_SESSID=returned_session_id"
```

### 购物车功能测试

```bash
# 1. 添加商品
curl -X POST http://localhost:8080/cart/add \
  -d "item_id=item1&item_name=iPhone&price=999&quantity=1" \
  -H "Cookie: YYHERTZ_SESSID=session_id"

# 2. 查看购物车
curl http://localhost:8080/cart \
  -H "Cookie: YYHERTZ_SESSID=session_id"

# 3. 删除商品
curl -X DELETE http://localhost:8080/cart/item/item1 \
  -H "Cookie: YYHERTZ_SESSID=session_id"
```

### 安全Cookie测试

```bash
# 测试安全Cookie演示
curl http://localhost:8080/demo/secure-cookie

# 返回结果包含：
# - 安全Cookie设置/获取
# - 篡改检测演示
# - HMAC验证过程
```

## 📊 性能基准

基于示例的性能测试结果：

| 操作类型 | 性能表现 | 内存使用 | 推荐使用场景 |
|---------|----------|----------|-------------|
| 基础Cookie | ~15 ns/op | 极低 | 用户偏好设置 |
| 安全Cookie | ~4200 ns/op | 低 | 认证令牌 |
| Session操作 | ~1000 ns/op | 低 | 用户状态管理 |
| 批量操作 | 显著提升 | 中 | 大量数据处理 |

### 性能优化建议

1. **复用Extension对象**
```go
// ✅ 好的做法
extension := session.NewExtensionForHertzContext(ctx)
// 多次使用extension

// ❌ 避免的做法
for range items {
    ext := session.NewExtensionForHertzContext(ctx) // 重复创建
}
```

2. **使用批量操作**
```go
// ✅ 批量操作
adapter := extension.StartSession()
for key, value := range data {
    adapter.Set(key, value)
}
adapter.Save()

// ❌ 逐个操作
for key, value := range data {
    extension.SetSession(key, value)
}
```

3. **适当的安全级别**
```go
// 普通数据
extension.SetCookie("theme", "dark")

// 敏感数据
extension.SetSecureCookie("secret", "csrf_token", "token_value")
```

## 🔧 自定义配置

### Session配置

```yaml
# session.yaml
session:
  provider: "memory"
  cookie_name: "YYHERTZ_SESSID"
  cookie_lifetime: 3600
  gc_maxlifetime: 3600
  cookie_secure: true
  cookie_httponly: true
  cookie_samesite: "Strict"
```

### 安全配置

```go
// 生产环境安全配置
options := session.CookieSecurityOptions{
    Secret:         os.Getenv("COOKIE_SECRET"), // 从环境变量读取
    MaxAge:         time.Hour * 24,             // 24小时过期
    ValidateExpiry: true,                       // 启用过期验证
    RequireHTTPS:   true,                       // 强制HTTPS
}
```

## 🚨 注意事项

### 安全建议

1. **HMAC密钥管理**
```bash
# 生成强密钥
openssl rand -hex 32

# 环境变量设置
export COOKIE_SECRET="your-generated-32-byte-key"
```

2. **生产环境配置**
```go
// 生产环境必须启用
options.RequireHTTPS = true
options.ValidateExpiry = true
```

3. **敏感数据处理**
```go
// ✅ 安全存储敏感数据
extension.SetSecureCookie("secret", "auth_token", token)

// ❌ 避免明文存储敏感数据
extension.SetCookie("auth_token", token) // 不安全
```

### 调试技巧

1. **启用调试日志**
```go
config := session.DefaultConfig()
config.EnableDebug = true
```

2. **Session状态检查**
```bash
curl http://localhost:8080/api/session/info
curl http://localhost:8080/api/cookies/info
```

3. **性能监控**
```bash
go run performance_comparison.go
```

## 🔗 相关资源

- [Session模块文档](../README.md)
- [架构更新文档](../../ARCHITECTURE_UPDATE.md)
- [迁移指南](../../MIGRATION_GUIDE.md)
- [性能基准测试](../benchmark_test.go)
- [生产环境测试](../production_compatibility_test.go)

## 🆘 常见问题

### Q: 如何从beego迁移？
A: 参考 `migration_example.go`，API几乎一致，可以无缝迁移。

### Q: 性能有影响吗？
A: 运行 `performance_comparison.go` 查看详细对比，代理开销 < 2%。

### Q: 如何启用分布式Session？
A: 配置Session存储提供者为Redis或数据库。

### Q: 安全Cookie如何工作？
A: 使用HMAC-SHA256签名防篡改，参考 `complete_example.go` 的安全Cookie演示。

---

**如有问题或建议，请参考主文档或提交Issue。**