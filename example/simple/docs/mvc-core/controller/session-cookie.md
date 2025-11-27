# 🍪 Session和Cookie管理

YYHertz控制器提供了完整的Session和Cookie管理功能，支持安全存储和高效访问。

## 🍪 Cookie操作方法

### Cookie基础方法 (6个)

| 方法 | 说明 | 示例 |
|------|------|------|
| `SetCookie(name, value, options...)` | 设置Cookie | `c.SetCookie("user_id", "123")` |
| `GetCookie(name)` | 获取Cookie | `userID := c.GetCookie("user_id")` |
| `DeleteCookie(name, path...)` | 删除Cookie | `c.DeleteCookie("user_id")` |
| `HasCookie(name)` | 检查Cookie是否存在 | `exists := c.HasCookie("user_id")` |
| `SetSecureCookie(secret, name, value, others...)` | 设置加密Cookie | `c.SetSecureCookie("key", "token", jwt)` |
| `GetSecureCookie(secret, name)` | 获取加密Cookie | `token, ok := c.GetSecureCookie("key", "token")` |

### Cookie使用示例

```go
func (c *AuthController) PostLogin() {
    username := c.GetString("username")
    password := c.GetString("password")
    
    // 验证用户身份
    user, err := c.authService.Login(username, password)
    if err != nil {
        c.JSONError("登录失败: " + err.Error())
        return
    }
    
    // 🍪 设置登录Cookie
    options := &cookie.Options{
        MaxAge:   7 * 24 * 3600, // 7天
        Path:     "/",
        HttpOnly: true,           // 防止XSS
        Secure:   true,           // HTTPS环境
        SameSite: "Strict",       // CSRF防护
    }
    
    c.SetCookie("user_id", strconv.Itoa(user.ID), options)
    c.SetCookie("username", user.Username, options)
    
    // 🔐 设置加密Token
    token := c.generateJWTToken(user)
    c.SetSecureCookie("jwt_secret", "auth_token", token, 7*24*3600)
    
    c.JSONSuccess("登录成功", user)
}
```

## 🗝️ Session管理方法

### Session核心方法 (4个)

| 方法 | 说明 | 示例 |
|------|------|------|
| `SetSession(key, value)` | 设置Session | `c.SetSession("user_id", 123)` |
| `GetSession(key)` | 获取Session | `userID := c.GetSession("user_id")` |
| `DelSession(key)` | 删除Session | `c.DelSession("user_id")` |
| `DestroySession()` | 销毁整个Session | `c.DestroySession()` |

### Session使用示例

```go
func (c *UserController) handleUserSession() {
    // 📝 设置用户信息到Session
    user := c.getCurrentUser()
    c.SetSession("user_id", user.ID)
    c.SetSession("username", user.Username)
    c.SetSession("role", user.Role)
    c.SetSession("login_time", time.Now())
    
    // 📊 获取Session数据
    userID := c.GetSession("user_id")
    if userID != nil {
        c.LogInfo("用户已登录", map[string]any{
            "user_id": userID,
            "path":    c.Ctx.Path(),
        })
    }
    
    // 🔄 更新Session数据
    c.SetSession("last_activity", time.Now())
    c.SetSession("page_views", c.getSessionInt("page_views")+1)
}

// 退出登录
func (c *AuthController) PostLogout() {
    // 📝 记录退出日志
    userID := c.GetSession("user_id")
    c.LogInfo("用户退出登录", map[string]any{"user_id": userID})
    
    // 🧹 清理Session
    c.DestroySession()
    
    // 🍪 清理Cookie
    c.DeleteCookie("user_id")
    c.DeleteCookie("username")
    c.DeleteCookie("auth_token")
    
    c.JSONSuccess("退出成功", nil)
}
```

## 🔐 安全Cookie实现

### 加密Cookie示例

```go
func (c *BaseController) handleSecureCookies() {
    // 🔑 安全密钥管理
    secretKey := c.app.Config.GetString("cookie.secret_key")
    if secretKey == "" {
        secretKey = "your-secure-secret-key-here"
    }
    
    // 💾 存储敏感数据
    sensitiveData := map[string]any{
        "user_id":    123,
        "role":       "admin",
        "created_at": time.Now(),
    }
    
    dataJSON, _ := json.Marshal(sensitiveData)
    c.SetSecureCookie(secretKey, "secure_data", string(dataJSON), 3600)
    
    // 📖 读取敏感数据
    if encryptedData, ok := c.GetSecureCookie(secretKey, "secure_data"); ok {
        var data map[string]any
        if err := json.Unmarshal([]byte(encryptedData), &data); err == nil {
            c.LogInfo("安全Cookie数据", data)
        }
    }
}
```

## 🛡️ 安全最佳实践

### Cookie安全配置

```go
type SecureCookieConfig struct {
    HttpOnly bool   // 防止XSS攻击
    Secure   bool   // 仅HTTPS传输
    SameSite string // CSRF防护
    Domain   string // 作用域限制
    Path     string // 路径限制
    MaxAge   int    // 有效期
}

func (c *BaseController) setSecureCookie(name, value string) {
    config := SecureCookieConfig{
        HttpOnly: true,
        Secure:   c.isHTTPS(),
        SameSite: "Strict",
        Path:     "/",
        MaxAge:   24 * 3600, // 24小时
    }
    
    options := &cookie.Options{
        HttpOnly: config.HttpOnly,
        Secure:   config.Secure,
        SameSite: config.SameSite,
        Path:     config.Path,
        MaxAge:   config.MaxAge,
    }
    
    // 🔐 加密存储
    if c.shouldEncrypt(name) {
        secretKey := c.getSecretKey()
        c.SetSecureCookie(secretKey, name, value, config.MaxAge)
    } else {
        c.SetCookie(name, value, options)
    }
}
```

### Session安全管理

```go
func (c *BaseController) handleSessionSecurity() {
    // 🔄 Session ID再生（防止会话固定攻击）
    if c.shouldRegenerateSessionID() {
        c.regenerateSessionID()
    }
    
    // ⏰ Session超时检查
    lastActivity := c.GetSession("last_activity")
    if lastActivity != nil {
        if time.Since(lastActivity.(time.Time)) > 30*time.Minute {
            c.LogWarn("Session超时", map[string]any{
                "session_id": c.GetSessionID(),
                "last_activity": lastActivity,
            })
            c.DestroySession()
            c.JSONUnauthorized("会话已超时，请重新登录")
            return
        }
    }
    
    // 🔒 IP地址验证（防止会话劫持）
    sessionIP := c.GetSession("ip_address")
    currentIP := c.GetClientIP()
    if sessionIP != nil && sessionIP.(string) != currentIP {
        c.LogError("检测到会话劫持", map[string]any{
            "session_ip": sessionIP,
            "current_ip": currentIP,
        })
        c.DestroySession()
        c.JSONForbidden("安全验证失败")
        return
    }
    
    // 📝 更新活动时间
    c.SetSession("last_activity", time.Now())
}
```

## 📊 购物车实现示例

```go
type CartItem struct {
    ProductID int     `json:"product_id"`
    Name      string  `json:"name"`
    Price     float64 `json:"price"`
    Quantity  int     `json:"quantity"`
    AddedAt   time.Time `json:"added_at"`
}

type ShoppingCart struct {
    Items     []CartItem `json:"items"`
    Total     float64    `json:"total"`
    UpdatedAt time.Time  `json:"updated_at"`
}

func (c *ShopController) AddToCart() {
    productID := c.GetInt("product_id")
    quantity := c.GetInt("quantity", 1)
    
    // 📦 获取商品信息
    product, err := c.productService.GetProduct(productID)
    if err != nil {
        c.JSONError("商品不存在")
        return
    }
    
    // 🛒 获取当前购物车
    cart := c.getCartFromSession()
    
    // ➕ 添加商品到购物车
    cart.AddItem(CartItem{
        ProductID: product.ID,
        Name:      product.Name,
        Price:     product.Price,
        Quantity:  quantity,
        AddedAt:   time.Now(),
    })
    
    // 💾 保存购物车到Session
    c.saveCartToSession(cart)
    
    c.JSONSuccess("添加成功", map[string]any{
        "cart_total": cart.Total,
        "item_count": len(cart.Items),
    })
}

func (c *ShopController) getCartFromSession() *ShoppingCart {
    cartData := c.GetSession("shopping_cart")
    if cartData == nil {
        return &ShoppingCart{
            Items:     make([]CartItem, 0),
            UpdatedAt: time.Now(),
        }
    }
    
    var cart ShoppingCart
    if err := json.Unmarshal(cartData.([]byte), &cart); err != nil {
        return &ShoppingCart{Items: make([]CartItem, 0)}
    }
    
    return &cart
}

func (c *ShopController) saveCartToSession(cart *ShoppingCart) {
    cart.UpdatedAt = time.Now()
    cart.calculateTotal()
    
    cartJSON, _ := json.Marshal(cart)
    c.SetSession("shopping_cart", cartJSON)
}
```

## ❓ 常见问题解答

**Q: Cookie和Session有什么区别？**
A: Cookie存储在客户端浏览器，Session存储在服务器端，通过Session ID关联。

**Q: 如何设置Cookie的过期时间？**
A: 使用MaxAge参数设置秒数，或使用Expires设置具体时间。

**Q: Session数据丢失怎么办？**
A: 检查Session存储后端配置，确认Session ID传输正常，考虑使用持久化存储。

**Q: 如何在分布式环境中使用Session？**
A: 使用Redis、Memcached等共享存储作为Session后端。