# MVC框架验证码模块

基于CloudWeGo-Hertz的高性能验证码模块，提供图形验证码生成、验证和中间件支持。

## 特性

- 🎨 **自定义图形验证码生成** - 支持自定义尺寸、长度、字符集
- 🏪 **多种存储方式** - 内存存储、Session存储、Redis存储（可扩展）
- 🛡️ **中间件支持** - 开箱即用的Hertz中间件
- ⚡ **高性能** - 优化的图像生成算法，支持高并发
- 🔒 **安全性** - 验证后自动销毁，防止重放攻击
- 🎯 **易于集成** - 简单的API设计，快速集成到现有项目

## 快速开始

### 1. 基本使用

```go
package main

import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/zsy619/yyhertz/framework/mvc/captcha"
)

func main() {
    // 创建验证码配置
    config := captcha.DefaultConfig()
    config.Width = 120
    config.Height = 40
    config.Length = 4
    
    // 创建存储
    store := captcha.NewMemoryStore()
    defer store.Close()
    
    // 创建验证码生成器
    generator := captcha.NewGenerator(config, store)
    
    // 创建Hertz应用
    h := server.Default()
    
    // 注册验证码路由
    h.GET("/captcha/generate", captcha.GenerateHandler(generator))
    h.GET("/captcha/image/:id", captcha.ImageHandler(generator))
    h.POST("/captcha/verify", captcha.VerifyHandler(generator))
    
    h.Spin()
}
```

### 2. 中间件使用

```go
// 创建中间件配置
middlewareConfig := &captcha.MiddlewareConfig{
    SkipPaths: []string{"/captcha/", "/public/"},
    ErrorHandler: func(c context.Context, ctx *app.RequestContext, err error) {
        ctx.JSON(400, map[string]any{
            "error": err.Error(),
        })
    },
}

// 创建中间件
middleware := captcha.NewMiddleware(generator, middlewareConfig)

// 应用到需要保护的路由
protected := h.Group("/api", middleware.Handler())
{
    protected.POST("/login", loginHandler)
    protected.POST("/register", registerHandler)
}
```

### 3. 自定义存储

```go
// 使用Session存储
sessionStore := captcha.NewSessionStore(
    "captcha_session_key",
    func() any {
        // 返回session数据
        return getSessionData()
    },
    func(key string, value any) error {
        // 设置session数据
        return setSessionData(key, value)
    },
)

generator := captcha.NewGenerator(config, sessionStore)
```

## API文档

### 配置选项

```go
type Config struct {
    Width   int    // 图片宽度，默认120
    Height  int    // 图片高度，默认40
    Length  int    // 验证码长度，默认4
    TTL     int64  // 过期时间(秒)，默认300
    Charset string // 字符集，默认"0123456789"
}
```

### 核心方法

#### Generator

```go
// 生成验证码
func (g *Generator) Generate() (*Captcha, error)

// 验证验证码
func (g *Generator) Verify(id, code string) bool

// 获取验证码图片
func (g *Generator) GetImage(id string) ([]byte, error)
```

#### Store接口

```go
type Store interface {
    Set(id string, captcha *Captcha) error
    Get(id string) (*Captcha, error)
    Delete(id string) error
    Clear() error
}
```

### HTTP路由

| 方法 | 路径 | 描述 | 响应 |
|------|------|------|------|
| GET | `/captcha/generate` | 生成新的验证码 | `{"captcha_id": "xxx"}` |
| GET | `/captcha/image/:id` | 获取验证码图片 | PNG图片 |
| POST | `/captcha/verify` | 验证验证码 | `{"code": 0, "message": "success"}` |

### 前端集成示例

```html
<!DOCTYPE html>
<html>
<head>
    <title>验证码示例</title>
</head>
<body>
    <div>
        <img id="captcha-image" src="" alt="验证码" onclick="refreshCaptcha()">
        <button onclick="refreshCaptcha()">刷新</button>
    </div>
    <form>
        <input type="hidden" id="captcha-id" name="captcha_id">
        <input type="text" id="captcha-code" name="captcha_code" placeholder="请输入验证码">
        <button type="submit">提交</button>
    </form>

    <script>
        let currentCaptchaId = '';

        // 刷新验证码
        async function refreshCaptcha() {
            try {
                const response = await fetch('/captcha/generate');
                const data = await response.json();
                
                currentCaptchaId = data.captcha_id;
                document.getElementById('captcha-id').value = currentCaptchaId;
                document.getElementById('captcha-image').src = `/captcha/image/${currentCaptchaId}`;
            } catch (error) {
                console.error('刷新验证码失败:', error);
            }
        }

        // 页面加载时生成验证码
        window.onload = refreshCaptcha;

        // 表单提交
        document.querySelector('form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = new FormData();
            formData.append('captcha_id', currentCaptchaId);
            formData.append('captcha_code', document.getElementById('captcha-code').value);
            
            try {
                const response = await fetch('/captcha/verify', {
                    method: 'POST',
                    body: formData
                });
                
                const result = await response.json();
                if (result.code === 0) {
                    alert('验证成功');
                } else {
                    alert('验证失败: ' + result.message);
                    refreshCaptcha(); // 验证失败后刷新验证码
                }
            } catch (error) {
                console.error('验证失败:', error);
            }
        });
    </script>
</body>
</html>
```

## 性能优化

### 1. 图像生成优化

- 使用高效的点阵字体渲染
- 优化图像绘制算法
- 减少内存分配

### 2. 存储优化

- 内存存储支持自动清理过期验证码
- Session存储复用现有会话机制
- Redis存储支持分布式部署

### 3. 并发支持

- 无锁设计，支持高并发
- 验证码生成器线程安全
- 存储操作原子性保证

## 安全考虑

1. **验证后销毁** - 验证成功后立即删除验证码，防止重放攻击
2. **过期机制** - 自动清理过期验证码，防止内存泄露
3. **随机生成** - 使用加密安全的随机数生成器
4. **大小写不敏感** - 提升用户体验的同时保持安全性

## 扩展开发

### 自定义字体

```go
// 扩展字符集支持字母
config := &captcha.Config{
    Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",
    // ... 其他配置
}
```

### 自定义存储实现

```go
type CustomStore struct {
    // 自定义存储实现
}

func (s *CustomStore) Set(id string, captcha *Captcha) error {
    // 实现存储逻辑
    return nil
}

func (s *CustomStore) Get(id string) (*Captcha, error) {
    // 实现获取逻辑
    return nil, nil
}

func (s *CustomStore) Delete(id string) error {
    // 实现删除逻辑
    return nil
}

func (s *CustomStore) Clear() error {
    // 实现清空逻辑
    return nil
}
```

### 自定义图像样式

通过修改`generateImage`方法可以自定义：
- 背景色和字体颜色
- 干扰线样式和数量
- 噪点密度和分布
- 字符倾斜和扭曲效果

## 最佳实践

1. **生产环境配置**
   - 使用Redis存储支持集群部署
   - 设置合适的过期时间（推荐5-10分钟）
   - 配置合适的图片尺寸（推荐120x40）

2. **前端集成**
   - 提供刷新验证码功能
   - 验证失败后自动刷新
   - 显示验证码加载状态

3. **安全设置**
   - 限制验证码生成频率
   - 记录验证失败次数
   - 实施IP访问限制

## 贡献

欢迎提交Issue和Pull Request来改进这个验证码模块。

## 许可证

MIT License
