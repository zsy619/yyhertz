// Package examples 完整的YYHertz Session使用示例
//
// 这个示例展示了如何在实际项目中使用YYHertz的Session和Cookie功能
// 包含用户认证、购物车、安全令牌等常见Web应用场景
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/framework/mvc/session"
)

// 用户信息结构
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 购物车商品结构
type CartItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// 全局配置
var (
	hmacSecret = getEnvOrDefault("HMAC_SECRET", "default-hmac-secret-key-change-in-production")
	serverPort = getEnvOrDefault("SERVER_PORT", "8080")
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Test_Complete(t *testing.T) {
	// 创建Hertz服务器
	h := server.Default(server.WithHostPorts(":" + serverPort))

	// 注册路由
	setupRoutes(h)

	fmt.Printf("🚀 YYHertz Session示例服务器启动在端口 %s\n", serverPort)
	fmt.Println("📖 访问以下端点来测试功能:")
	fmt.Println("   GET  /                    - 首页")
	fmt.Println("   POST /auth/login          - 用户登录")
	fmt.Println("   POST /auth/logout         - 用户注销")
	fmt.Println("   GET  /profile             - 用户资料（需要登录）")
	fmt.Println("   POST /cart/add            - 添加商品到购物车")
	fmt.Println("   GET  /cart                - 查看购物车")
	fmt.Println("   GET  /settings            - 用户设置")
	fmt.Println("   POST /settings/theme      - 更新主题设置")

	// 启动服务器
	h.Spin()
}

func setupRoutes(h *server.Hertz) {
	// 首页
	h.GET("/", homeHandler)

	// 认证相关
	auth := h.Group("/auth")
	{
		auth.POST("/login", loginHandler)
		auth.POST("/logout", logoutHandler)
	}

	// 需要认证的路由
	h.GET("/profile", requireAuth, profileHandler)

	// 购物车功能
	cart := h.Group("/cart")
	{
		cart.POST("/add", addToCartHandler)
		cart.GET("/", viewCartHandler)
		cart.DELETE("/item/:id", removeFromCartHandler)
		cart.POST("/clear", clearCartHandler)
	}

	// 用户设置
	settings := h.Group("/settings")
	{
		settings.GET("/", settingsHandler)
		settings.POST("/theme", updateThemeHandler)
		settings.POST("/language", updateLanguageHandler)
	}

	// API端点
	api := h.Group("/api")
	{
		api.GET("/session/info", sessionInfoHandler)
		api.GET("/cookies/info", cookieInfoHandler)
		api.POST("/security/csrf", csrfTokenHandler)
	}

	// 演示安全Cookie
	h.GET("/demo/secure-cookie", secureCookieDemo)
}

// ============= 中间件 =============

// requireAuth 认证中间件
func requireAuth(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	userID := extension.GetSession("user_id")
	if userID == nil {
		c.JSON(consts.StatusUnauthorized, utils.H{
			"error": "请先登录",
			"code":  401,
		})
		c.Abort()
		return
	}

	// 验证CSRF token（如果是POST/PUT/DELETE请求）
	if c.Request.Header.IsPost() || c.Request.Header.IsPut() || c.Request.Header.IsDelete() {
		csrfToken := string(c.Request.Header.Peek("X-CSRF-Token"))
		if csrfToken == "" {
			csrfToken = c.Query("csrf_token")
		}

		if csrfToken != "" {
			if storedToken, ok := extension.GetSecureCookie(hmacSecret, "csrf_token"); !ok || storedToken != csrfToken {
				c.JSON(consts.StatusForbidden, utils.H{
					"error": "CSRF token无效",
					"code":  403,
				})
				c.Abort()
				return
			}
		}
	}

	c.Next(ctx)
}

// ============= 处理函数 =============

// homeHandler 首页
func homeHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	// 记录访问统计
	visitorID := extension.GetCookie("visitor_id")
	if visitorID == "" {
		visitorID = fmt.Sprintf("visitor_%d", time.Now().UnixNano())
		extension.SetCookie("visitor_id", visitorID, 86400*30) // 30天
	}

	// 更新最后访问时间
	extension.SetCookie("last_visit", time.Now().Format(time.RFC3339))

	// 获取用户信息（如果已登录）
	var user *User
	if userID := extension.GetSession("user_id"); userID != nil {
		user = &User{
			ID:       userID.(string),
			Username: extension.GetSession("username").(string),
		}
	}

	// 获取用户偏好
	theme := extension.GetCookie("theme")
	if theme == "" {
		theme = "light"
	}

	language := extension.GetCookie("language")
	if language == "" {
		language = "zh-CN"
	}

	c.JSON(consts.StatusOK, utils.H{
		"message":    "欢迎使用YYHertz Session示例",
		"visitor_id": visitorID,
		"user":       user,
		"preferences": utils.H{
			"theme":    theme,
			"language": language,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// loginHandler 用户登录
func loginHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	// 获取登录参数
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(consts.StatusBadRequest, utils.H{
			"error": "用户名和密码不能为空",
			"code":  400,
		})
		return
	}

	// 模拟用户验证（实际项目中应该查询数据库）
	if username == "demo" && password == "password" {
		// 登录成功，创建Session
		userID := fmt.Sprintf("user_%d", time.Now().Unix())

		err := extension.SetSession("user_id", userID)
		if err != nil {
			log.Printf("设置user_id session失败: %v", err)
		}

		err = extension.SetSession("username", username)
		if err != nil {
			log.Printf("设置username session失败: %v", err)
		}

		err = extension.SetSession("login_time", time.Now().Unix())
		if err != nil {
			log.Printf("设置login_time session失败: %v", err)
		}

		// 生成CSRF token
		csrfToken := fmt.Sprintf("csrf_%d_%s", time.Now().UnixNano(), userID)
		extension.SetSecureCookie(hmacSecret, "csrf_token", csrfToken)

		// 设置记住登录状态的Cookie（可选）
		rememberMe := c.PostForm("remember_me")
		if rememberMe == "true" {
			extension.SetCookie("remember_token", userID, 86400*30) // 30天
		}

		c.JSON(consts.StatusOK, utils.H{
			"message":    "登录成功",
			"user_id":    userID,
			"username":   username,
			"csrf_token": csrfToken,
		})
	} else {
		c.JSON(consts.StatusUnauthorized, utils.H{
			"error": "用户名或密码错误",
			"code":  401,
		})
	}
}

// logoutHandler 用户注销
func logoutHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	// 清除Session
	extension.ClearSession()

	// 删除相关Cookie
	extension.DelCookie("csrf_token")
	extension.DelCookie("remember_token")

	c.JSON(consts.StatusOK, utils.H{
		"message": "注销成功",
	})
}

// profileHandler 用户资料
func profileHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	userID := extension.GetSession("user_id").(string)
	username := extension.GetSession("username").(string)
	loginTime := extension.GetSession("login_time").(int64)

	// 获取购物车商品数量
	cartItemsRaw := extension.GetSession("cart_items")
	cartItemCount := 0
	if cartItemsRaw != nil {
		if cartItems, ok := cartItemsRaw.([]CartItem); ok {
			for _, item := range cartItems {
				cartItemCount += item.Quantity
			}
		}
	}

	c.JSON(consts.StatusOK, utils.H{
		"user": utils.H{
			"id":         userID,
			"username":   username,
			"login_time": time.Unix(loginTime, 0).Format(time.RFC3339),
		},
		"cart_item_count": cartItemCount,
		"session_id":      extension.GetSessionID(),
	})
}

// ============= 购物车功能 =============

// addToCartHandler 添加商品到购物车
func addToCartHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	// 获取商品信息
	itemID := c.PostForm("item_id")
	itemName := c.PostForm("item_name")
	priceStr := c.PostForm("price")
	quantityStr := c.PostForm("quantity")

	if itemID == "" || itemName == "" || priceStr == "" {
		c.JSON(consts.StatusBadRequest, utils.H{
			"error": "商品信息不完整",
			"code":  400,
		})
		return
	}

	// 解析价格和数量
	var price float64 = 0
	var quantity int = 1
	fmt.Sscanf(priceStr, "%f", &price)
	fmt.Sscanf(quantityStr, "%d", &quantity)

	// 获取现有购物车
	var cartItems []CartItem
	if cartItemsRaw := extension.GetSession("cart_items"); cartItemsRaw != nil {
		if items, ok := cartItemsRaw.([]CartItem); ok {
			cartItems = items
		}
	}

	// 检查商品是否已存在
	found := false
	for i, item := range cartItems {
		if item.ID == itemID {
			cartItems[i].Quantity += quantity
			found = true
			break
		}
	}

	// 如果商品不存在，添加新商品
	if !found {
		cartItems = append(cartItems, CartItem{
			ID:       itemID,
			Name:     itemName,
			Price:    price,
			Quantity: quantity,
		})
	}

	// 保存购物车
	err := extension.SetSession("cart_items", cartItems)
	if err != nil {
		log.Printf("保存购物车失败: %v", err)
		c.JSON(consts.StatusInternalServerError, utils.H{
			"error": "保存购物车失败",
			"code":  500,
		})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"message":    "商品已添加到购物车",
		"cart_items": cartItems,
	})
}

// viewCartHandler 查看购物车
func viewCartHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	var cartItems []CartItem
	if cartItemsRaw := extension.GetSession("cart_items"); cartItemsRaw != nil {
		if items, ok := cartItemsRaw.([]CartItem); ok {
			cartItems = items
		}
	}

	// 计算总价
	var totalPrice float64 = 0
	var totalQuantity int = 0
	for _, item := range cartItems {
		totalPrice += item.Price * float64(item.Quantity)
		totalQuantity += item.Quantity
	}

	c.JSON(consts.StatusOK, utils.H{
		"cart_items":     cartItems,
		"total_price":    totalPrice,
		"total_quantity": totalQuantity,
		"cart_id":        extension.GetSessionID(),
	})
}

// removeFromCartHandler 从购物车删除商品
func removeFromCartHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(consts.StatusBadRequest, utils.H{
			"error": "商品ID不能为空",
			"code":  400,
		})
		return
	}

	// 获取现有购物车
	var cartItems []CartItem
	if cartItemsRaw := extension.GetSession("cart_items"); cartItemsRaw != nil {
		if items, ok := cartItemsRaw.([]CartItem); ok {
			cartItems = items
		}
	}

	// 删除指定商品
	newCartItems := make([]CartItem, 0)
	for _, item := range cartItems {
		if item.ID != itemID {
			newCartItems = append(newCartItems, item)
		}
	}

	// 保存更新后的购物车
	err := extension.SetSession("cart_items", newCartItems)
	if err != nil {
		log.Printf("更新购物车失败: %v", err)
		c.JSON(consts.StatusInternalServerError, utils.H{
			"error": "更新购物车失败",
			"code":  500,
		})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"message":    "商品已从购物车删除",
		"cart_items": newCartItems,
	})
}

// clearCartHandler 清空购物车
func clearCartHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	err := extension.DelSession("cart_items")
	if err != nil {
		log.Printf("清空购物车失败: %v", err)
		c.JSON(consts.StatusInternalServerError, utils.H{
			"error": "清空购物车失败",
			"code":  500,
		})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"message": "购物车已清空",
	})
}

// ============= 用户设置 =============

// settingsHandler 用户设置页面
func settingsHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	// 获取用户偏好设置
	theme := extension.GetCookie("theme")
	if theme == "" {
		theme = "light"
	}

	language := extension.GetCookie("language")
	if language == "" {
		language = "zh-CN"
	}

	timezone := extension.GetCookie("timezone")
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}

	c.JSON(consts.StatusOK, utils.H{
		"settings": utils.H{
			"theme":    theme,
			"language": language,
			"timezone": timezone,
		},
	})
}

// updateThemeHandler 更新主题设置
func updateThemeHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	theme := c.PostForm("theme")
	if theme != "light" && theme != "dark" && theme != "auto" {
		c.JSON(consts.StatusBadRequest, utils.H{
			"error": "无效的主题设置",
			"code":  400,
		})
		return
	}

	// 保存主题设置到Cookie（30天有效期）
	extension.SetCookie("theme", theme, 86400*30)

	c.JSON(consts.StatusOK, utils.H{
		"message": "主题设置已更新",
		"theme":   theme,
	})
}

// updateLanguageHandler 更新语言设置
func updateLanguageHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	language := c.PostForm("language")
	if language == "" {
		c.JSON(consts.StatusBadRequest, utils.H{
			"error": "语言设置不能为空",
			"code":  400,
		})
		return
	}

	// 保存语言设置到Cookie
	extension.SetCookie("language", language, 86400*30)

	c.JSON(consts.StatusOK, utils.H{
		"message":  "语言设置已更新",
		"language": language,
	})
}

// ============= API端点 =============

// sessionInfoHandler Session信息API
func sessionInfoHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	sessionInfo := utils.H{
		"session_id":      extension.GetSessionID(),
		"is_started":      extension.IsSessionStarted(),
		"user_id":         extension.GetSession("user_id"),
		"username":        extension.GetSession("username"),
		"login_time":      extension.GetSession("login_time"),
		"cart_item_count": 0,
	}

	// 计算购物车商品数量
	if cartItemsRaw := extension.GetSession("cart_items"); cartItemsRaw != nil {
		if cartItems, ok := cartItemsRaw.([]CartItem); ok {
			count := 0
			for _, item := range cartItems {
				count += item.Quantity
			}
			sessionInfo["cart_item_count"] = count
		}
	}

	c.JSON(consts.StatusOK, utils.H{
		"session": sessionInfo,
	})
}

// cookieInfoHandler Cookie信息API
func cookieInfoHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	// 获取所有Cookie
	allCookies := extension.Cookie.GetAll()

	// 安全起见，过滤敏感Cookie
	safeCookies := make(map[string]string)
	for name, value := range allCookies {
		// 过滤敏感Cookie
		if name != "csrf_token" && name != "remember_token" {
			safeCookies[name] = value
		} else {
			safeCookies[name] = "[HIDDEN]"
		}
	}

	c.JSON(consts.StatusOK, utils.H{
		"cookies": safeCookies,
		"count":   extension.Cookie.Count(),
	})
}

// csrfTokenHandler 获取CSRF令牌
func csrfTokenHandler(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	// 生成新的CSRF token
	csrfToken := fmt.Sprintf("csrf_%d", time.Now().UnixNano())
	extension.SetSecureCookie(hmacSecret, "csrf_token", csrfToken)

	c.JSON(consts.StatusOK, utils.H{
		"csrf_token": csrfToken,
		"expires_in": 3600, // 1小时
	})
}

// secureCookieDemo 安全Cookie演示
func secureCookieDemo(ctx context.Context, c *app.RequestContext) {
	extension := session.NewExtensionForHertzContext(c)

	// 设置安全Cookie
	demoData := fmt.Sprintf("demo_data_%d", time.Now().Unix())
	extension.SetSecureCookie(hmacSecret, "demo_secure", demoData)

	// 立即读取验证
	retrievedData, ok := extension.GetSecureCookie(hmacSecret, "demo_secure")

	// 演示篡改检测
	// 手动设置一个被篡改的Cookie来展示安全性
	tamperedValue := "tampered_value|" + fmt.Sprintf("%d", time.Now().Unix()) + "|invalid_signature"
	extension.SetCookie("demo_tampered", tamperedValue)

	// 尝试验证被篡改的Cookie
	_, tamperedOk := extension.GetSecureCookie(hmacSecret, "demo_tampered")

	c.JSON(consts.StatusOK, utils.H{
		"secure_cookie": utils.H{
			"original_data":  demoData,
			"retrieved_data": retrievedData,
			"verification":   ok,
		},
		"tamper_detection": utils.H{
			"tampered_cookie": tamperedValue,
			"verification":    tamperedOk,
			"message":         "被篡改的Cookie应该验证失败",
		},
		"hmac_algorithm": "SHA256",
		"timestamp":      time.Now().Format(time.RFC3339),
	})
}

/*
使用方法:

1. 运行服务器:
   go run complete_example.go

2. 测试登录:
   curl -X POST http://localhost:8080/auth/login \
     -d "username=demo&password=password&remember_me=true"

3. 查看用户资料:
   curl http://localhost:8080/profile \
     -H "Cookie: YYHERTZ_SESSID=your_session_id"

4. 添加商品到购物车:
   curl -X POST http://localhost:8080/cart/add \
     -d "item_id=item1&item_name=iPhone&price=999&quantity=1" \
     -H "Cookie: YYHERTZ_SESSID=your_session_id"

5. 查看购物车:
   curl http://localhost:8080/cart \
     -H "Cookie: YYHERTZ_SESSID=your_session_id"

6. 更新主题:
   curl -X POST http://localhost:8080/settings/theme \
     -d "theme=dark"

7. 查看Session信息:
   curl http://localhost:8080/api/session/info \
     -H "Cookie: YYHERTZ_SESSID=your_session_id"

8. 测试安全Cookie:
   curl http://localhost:8080/demo/secure-cookie

环境变量:
- HMAC_SECRET: HMAC密钥（生产环境必须设置）
- SERVER_PORT: 服务器端口（默认8080）
*/
