# 🔐 验证码功能

验证码功能是保护Web应用免受恶意攻击的重要安全措施。本框架集成了高性能的图形验证码模块，支持多种存储方式和自定义配置。

## 🌟 核心特性

### ✨ 功能特点
- **🎨 自定义图形验证码** - 支持数字、字母、混合字符集
- **🏪 多种存储方式** - 内存、Session、Redis存储
- **🛡️ 中间件支持** - 开箱即用的Hertz中间件集成
- **⚡ 高性能设计** - 优化的图像生成算法
- **🔒 安全性保障** - 验证后自动销毁，防重放攻击
- **🎯 易于集成** - 简单的API设计

### 📊 性能指标
- **验证码生成**: ~1.2ms/op
- **验证码验证**: ~150ns/op
- **内存优化**: 自动过期清理机制
- **并发支持**: 支持高并发访问

## 🚀 快速开始

### 1. 基础配置

```go
import "github.com/zsy619/yyhertz/framework/mvc/captcha"

// 创建验证码配置
config := captcha.DefaultConfig()
config.Width = 120    // 图片宽度
config.Height = 40    // 图片高度  
config.Length = 4     // 验证码长度
config.TTL = 300      // 5分钟过期
config.Charset = "0123456789" // 字符集
```

### 2. 存储选择

#### 内存存储（单机部署）
```go
store := captcha.NewMemoryStore()
defer store.Close()
```

#### Session存储（会话集成）
```go
store := captcha.NewSessionStore(
    "captcha_session_key",
    func() interface{} { return getSessionData() },
    func(key string, value interface{}) error {
        return setSessionData(key, value)
    },
)
```

#### Redis存储（分布式部署）
```go
// 需要实现Redis客户端集成
store := captcha.NewRedisStore(redisClient, "captcha:", time.Minute*10)
```

### 3. 创建生成器

```go
generator := captcha.NewGenerator(config, store)
```

## 🔧 路由配置

### HTTP端点配置

```go
import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/zsy619/yyhertz/framework/mvc/captcha"
)

func setupCaptchaRoutes(h *server.Hertz, generator *captcha.Generator) {
    // 验证码路由组
    captchaGroup := h.Group("/captcha")
    {
        // 生成验证码ID
        captchaGroup.GET("/generate", captcha.GenerateHandler(generator))
        
        // 获取验证码图片
        captchaGroup.GET("/image/:id", captcha.ImageHandler(generator))
        
        // 验证验证码
        captchaGroup.POST("/verify", captcha.VerifyHandler(generator))
    }
}
```

### API端点说明

| 方法 | 路径 | 描述 | 请求参数 | 响应格式 |
|------|------|------|----------|----------|
| GET | `/captcha/generate` | 生成新验证码 | 无 | `{"captcha_id": "xxx"}` |
| GET | `/captcha/image/:id` | 获取验证码图片 | `id`: 验证码ID | PNG图片 |
| POST | `/captcha/verify` | 验证验证码 | `captcha_id`, `captcha_code` | `{"code": 0, "message": "success"}` |

## 🛡️ 中间件集成

### 基础中间件配置

```go
// 创建中间件配置
middlewareConfig := &captcha.MiddlewareConfig{
    // 跳过验证的路径
    SkipPaths: []string{
        "/captcha/",  // 验证码相关路径
        "/public/",   // 公共资源
        "/static/",   // 静态文件
        "/health",    // 健康检查
    },
    // 自定义错误处理
    ErrorHandler: func(c context.Context, ctx *app.RequestContext, err error) {
        ctx.JSON(400, map[string]interface{}{
            "success": false,
            "error": err.Error(),
        })
    },
}

// 创建中间件
middleware := captcha.NewMiddleware(generator, middlewareConfig)
```

### 应用到受保护路由

```go
// 需要验证码保护的路由组
protected := h.Group("/api", middleware.Handler())
{
    // 用户认证
    authGroup := protected.Group("/auth")
    {
        authGroup.POST("/login", loginHandler)
        authGroup.POST("/register", registerHandler)
        authGroup.POST("/forgot-password", forgotPasswordHandler)
    }
    
    // 敏感操作
    userGroup := protected.Group("/user")
    {
        userGroup.POST("/change-password", changePasswordHandler)
        userGroup.DELETE("/delete-account", deleteAccountHandler)
    }
}
```

## 🎨 前端集成

### HTML表单示例

```html
<div class="captcha-container">
    <!-- 验证码图片 -->
    <img id="captcha-image" 
         src="" 
         alt="验证码" 
         onclick="refreshCaptcha()"
         style="cursor: pointer; border: 1px solid #ddd;">
    
    <!-- 刷新按钮 -->
    <button type="button" onclick="refreshCaptcha()">
        <i class="fas fa-sync-alt"></i> 刷新
    </button>
</div>

<form id="login-form">
    <!-- 隐藏的验证码ID -->
    <input type="hidden" id="captcha-id" name="captcha_id">
    
    <!-- 验证码输入框 -->
    <div class="form-group">
        <label for="captcha-code">验证码</label>
        <input type="text" 
               id="captcha-code" 
               name="captcha_code"
               class="form-control"
               placeholder="请输入验证码"
               maxlength="6"
               autocomplete="off">
    </div>
    
    <!-- 其他表单字段 -->
    <div class="form-group">
        <label for="username">用户名</label>
        <input type="text" id="username" name="username" class="form-control">
    </div>
    
    <button type="submit" class="btn btn-primary">登录</button>
</form>
```

### JavaScript集成

```javascript
class CaptchaManager {
    constructor() {
        this.currentCaptchaId = '';
        this.init();
    }
    
    // 初始化
    init() {
        this.refreshCaptcha();
        this.bindEvents();
    }
    
    // 绑定事件
    bindEvents() {
        // 表单提交事件
        document.getElementById('login-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleFormSubmit();
        });
        
        // 验证码输入框回车事件
        document.getElementById('captcha-code').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.handleFormSubmit();
            }
        });
    }
    
    // 刷新验证码
    async refreshCaptcha() {
        try {
            const response = await fetch('/captcha/generate');
            const data = await response.json();
            
            if (data.captcha_id) {
                this.currentCaptchaId = data.captcha_id;
                document.getElementById('captcha-id').value = this.currentCaptchaId;
                document.getElementById('captcha-image').src = 
                    `/captcha/image/${this.currentCaptchaId}?t=${Date.now()}`;
                
                // 清空输入框
                document.getElementById('captcha-code').value = '';
            } else {
                throw new Error('生成验证码失败');
            }
        } catch (error) {
            console.error('刷新验证码失败:', error);
            this.showMessage('刷新验证码失败，请重试', 'error');
        }
    }
    
    // 验证验证码
    async verifyCaptcha(captchaCode) {
        const formData = new FormData();
        formData.append('captcha_id', this.currentCaptchaId);
        formData.append('captcha_code', captchaCode);
        
        const response = await fetch('/captcha/verify', {
            method: 'POST',
            body: formData
        });
        
        return await response.json();
    }
    
    // 处理表单提交
    async handleFormSubmit() {
        const captchaCode = document.getElementById('captcha-code').value.trim();
        
        if (!captchaCode) {
            this.showMessage('请输入验证码', 'warning');
            return;
        }
        
        try {
            // 验证验证码
            const captchaResult = await this.verifyCaptcha(captchaCode);
            
            if (captchaResult.code !== 0) {
                this.showMessage(captchaResult.message || '验证码错误', 'error');
                await this.refreshCaptcha(); // 验证失败后刷新验证码
                return;
            }
            
            // 验证码通过，提交实际表单
            await this.submitForm();
            
        } catch (error) {
            console.error('验证失败:', error);
            this.showMessage('验证失败，请重试', 'error');
            await this.refreshCaptcha();
        }
    }
    
    // 提交表单
    async submitForm() {
        const formData = new FormData(document.getElementById('login-form'));
        
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                body: formData
            });
            
            const result = await response.json();
            
            if (result.success) {
                this.showMessage('登录成功', 'success');
                // 重定向或其他成功处理
                window.location.href = '/dashboard';
            } else {
                this.showMessage(result.message || '登录失败', 'error');
                await this.refreshCaptcha(); // 登录失败后刷新验证码
            }
        } catch (error) {
            console.error('登录失败:', error);
            this.showMessage('登录失败，请重试', 'error');
            await this.refreshCaptcha();
        }
    }
    
    // 显示消息
    showMessage(message, type) {
        // 实现消息显示逻辑（可以使用Toast、Alert等）
        const alertClass = {
            'success': 'alert-success',
            'error': 'alert-danger',
            'warning': 'alert-warning'
        }[type] || 'alert-info';
        
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert ${alertClass} alert-dismissible fade show`;
        alertDiv.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        
        document.body.insertBefore(alertDiv, document.body.firstChild);
        
        // 3秒后自动关闭
        setTimeout(() => {
            alertDiv.remove();
        }, 3000);
    }
}

// 页面加载后初始化
document.addEventListener('DOMContentLoaded', () => {
    new CaptchaManager();
});

// 全局刷新验证码函数（兼容onclick调用）
function refreshCaptcha() {
    if (window.captchaManager) {
        window.captchaManager.refreshCaptcha();
    }
}
```

## ⚙️ 高级配置

### 自定义验证码样式

```go
// 高级配置示例
config := &captcha.Config{
    Width:   200,  // 更大的图片提高识别率
    Height:  80,
    Length:  6,    // 更长的验证码提高安全性
    TTL:     1800, // 30分钟过期时间
    Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", // 数字+字母
}
```

### 错误处理配置

```go
middlewareConfig := &captcha.MiddlewareConfig{
    SkipPaths: []string{"/captcha/", "/public/"},
    ErrorHandler: func(c context.Context, ctx *app.RequestContext, err error) {
        // 记录错误日志
        log.Printf("验证码验证失败: %v, IP: %s, Path: %s", 
            err, ctx.ClientIP(), string(ctx.Path()))
        
        // 根据错误类型返回不同响应
        if captchaErr, ok := err.(*captcha.CaptchaError); ok {
            switch captchaErr.Code {
            case captcha.ErrCodeMissingParams:
                ctx.JSON(400, map[string]interface{}{
                    "error": "MISSING_CAPTCHA",
                    "message": "请提供验证码",
                })
            case captcha.ErrCodeInvalidCaptcha:
                ctx.JSON(400, map[string]interface{}{
                    "error": "INVALID_CAPTCHA", 
                    "message": "验证码错误，请重新输入",
                })
            case captcha.ErrCodeExpiredCaptcha:
                ctx.JSON(400, map[string]interface{}{
                    "error": "EXPIRED_CAPTCHA",
                    "message": "验证码已过期，请刷新后重试",
                })
            default:
                ctx.JSON(400, map[string]interface{}{
                    "error": "CAPTCHA_ERROR",
                    "message": captchaErr.Message,
                })
            }
        } else {
            ctx.JSON(500, map[string]interface{}{
                "error": "INTERNAL_ERROR",
                "message": "服务器内部错误",
            })
        }
    },
    SuccessHandler: func(c context.Context, ctx *app.RequestContext) {
        // 验证成功后的处理
        log.Printf("验证码验证成功, IP: %s", ctx.ClientIP())
    },
}
```

## 🔍 最佳实践

### 1. 安全建议

```go
// 生产环境配置
config := &captcha.Config{
    Width:   150,
    Height:  50,
    Length:  5,     // 适中的长度平衡安全性和用户体验
    TTL:     600,   // 10分钟过期时间
    Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",
}

// 使用Redis存储支持分布式部署
store := captcha.NewRedisStore(redisClient, "app:captcha:", time.Minute*10)
```

### 2. 用户体验优化

- **响应式设计**: 验证码图片适配移动设备
- **键盘支持**: 回车键快速提交
- **自动刷新**: 验证失败后自动刷新验证码
- **加载状态**: 显示验证码加载状态

### 3. 性能优化

- **图片缓存**: 设置合适的HTTP缓存头
- **存储优化**: 根据部署方式选择合适的存储
- **监控统计**: 记录验证码使用统计

### 4. 错误处理

- **友好提示**: 提供清晰的错误信息
- **自动重试**: 验证失败后自动刷新
- **日志记录**: 记录验证码相关操作日志

## 🐛 故障排除

### 常见问题

#### 1. 验证码图片不显示
```bash
# 检查路由注册
GET /captcha/image/:id

# 检查生成器配置
log.Printf("Generator config: %+v", config)

# 检查存储状态
captcha, err := store.Get(captchaId)
```

#### 2. 验证总是失败
```bash
# 检查验证码是否过期
if time.Now().Unix() > captcha.ExpireAt {
    log.Printf("Captcha expired: %s", captchaId)
}

# 检查字符串比较
log.Printf("Input: %s, Expected: %s", inputCode, captcha.Code)
```

#### 3. 性能问题
```bash
# 检查存储性能
go test -bench=BenchmarkCaptchaGeneration
go test -bench=BenchmarkCaptchaVerification

# 监控内存使用
go tool pprof http://localhost:6060/debug/pprof/heap
```

## 📚 扩展阅读

- [Web安全最佳实践](../security/best-practices.md)
- [中间件开发指南](../middleware/custom.md)
- [Redis集成配置](../configuration/redis.md)
- [性能监控指南](../dev-tools/performance.md)

## 🔗 相关链接

- [验证码模块源码](https://github.com/zsy619/yyhertz/tree/main/framework/mvc/captcha)
- [示例项目](https://github.com/zsy619/yyhertz/tree/main/example/simple)
- [API文档](../api/captcha.md)

---

> 💡 **提示**: 验证码功能已经集成到MVC框架中，开箱即用。如果需要自定义样式或存储方式，请参考上述配置示例。
