package session

import (
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
)

// SessionCookieUnifiedDemo 统一的Session和Cookie功能演示
func SessionCookieUnifiedDemo() {
	fmt.Println("🎯 YYHertz 统一Session & Cookie模块演示")
	fmt.Println("==========================================")
	
	// 创建演示用的上下文
	mockRequest := &app.RequestContext{}
	ext := NewExtensionForHertzContext(mockRequest)
	
	fmt.Println("\n📋 模块1: 基础Cookie操作")
	fmt.Println("--------------------------------------")
	
	fmt.Println("✅ 支持的基础Cookie方法:")
	fmt.Println("   - Cookie.Get(key) string               // 获取cookie")
	fmt.Println("   - Cookie.Set(name, value, ...)         // 设置cookie")
	fmt.Println("   - Cookie.Delete(name)                  // 删除cookie")
	fmt.Println("   - Cookie.Exists(key) bool              // 检查存在")
	fmt.Println("   - Cookie.GetAll() map[string]string    // 获取所有cookie")
	fmt.Println("   - Cookie.Count() int                   // cookie数量")
	fmt.Println("   - Cookie.Clear()                       // 清除所有cookie")
	
	// 演示基础cookie操作
	ext.SetCookie("demo_cookie", "demo_value", 3600, "/", "", false, false)
	fmt.Printf("✅ 设置Cookie: demo_cookie=demo_value\n")
	
	exists := ext.CookieExists("demo_cookie")
	fmt.Printf("✅ Cookie存在检查: %t\n", exists)
	
	fmt.Println("\n📋 模块2: 安全Cookie操作")
	fmt.Println("--------------------------------------")
	
	secret := "demo_hmac_secret_key_256bit_safe"
	
	fmt.Println("✅ 安全Cookie特性:")
	fmt.Println("   - HMAC-SHA256签名验证")
	fmt.Println("   - Base64编码保护")
	fmt.Println("   - 时间戳防重放攻击")
	fmt.Println("   - 完全兼容beego API")
	
	// 演示安全cookie操作
	ext.SetSecureCookie(secret, "secure_demo", "sensitive_data", 3600, "/")
	fmt.Printf("✅ 设置安全Cookie: secure_demo (已加密)\n")
	
	secureValue, valid := ext.GetSecureCookie(secret, "secure_demo")
	fmt.Printf("✅ 获取安全Cookie: value=%s, valid=%t\n", secureValue, valid)
	
	fmt.Println("\n📋 模块3: 高级安全选项")
	fmt.Println("--------------------------------------")
	
	// 高级安全配置
	securityOptions := cookie.CookieSecurityOptions{
		Secret:          secret,
		MaxAge:          time.Hour * 24,  // 24小时有效期
		ValidateExpiry:  true,            // 验证过期时间
		RequireHTTPS:    false,           // 演示环境设为false
	}
	
	fmt.Println("✅ 高级安全特性:")
	fmt.Printf("   Secret: %s\n", securityOptions.Secret[:20]+"...")
	fmt.Printf("   MaxAge: %v\n", securityOptions.MaxAge)
	fmt.Printf("   ValidateExpiry: %t\n", securityOptions.ValidateExpiry)
	fmt.Printf("   RequireHTTPS: %t\n", securityOptions.RequireHTTPS)
	
	fmt.Println("\n   方法:")
	fmt.Println("   - SetSecureWithOptions(name, value, options, ...)")
	fmt.Println("   - GetSecureWithOptions(key, options) (value, bool, error)")
	
	fmt.Println("\n📋 模块4: Session管理")
	fmt.Println("--------------------------------------")
	
	fmt.Println("✅ Session适配器功能:")
	fmt.Println("   - StartSession() *Adapter              // 启动session")
	fmt.Println("   - SetSession(key, value) error         // 设置session数据")
	fmt.Println("   - GetSession(key) any                  // 获取session数据")
	fmt.Println("   - DelSession(key) error                // 删除session数据")
	fmt.Println("   - GetSessionID() string                // 获取session ID")
	fmt.Println("   - IsSessionStarted() bool              // 检查session状态")
	
	// 演示session操作
	adapter := ext.StartSession()
	if adapter != nil {
		fmt.Printf("✅ Session已启动，ID: %s\n", ext.GetSessionID())
		
		// 设置session数据
		ext.SetSession("user_id", "12345")
		ext.SetSession("username", "demo_user")
		fmt.Println("✅ 设置Session数据: user_id, username")
		
		// 获取session数据
		userID := ext.GetSession("user_id")
		username := ext.GetSession("username")
		fmt.Printf("✅ 获取Session数据: user_id=%v, username=%v\n", userID, username)
	}
	
	fmt.Println("\n📋 模块5: Context扩展接口")
	fmt.Println("--------------------------------------")
	
	fmt.Println("✅ 统一的Context扩展:")
	fmt.Println("   - NewExtensionForHertzContext(ctx)")
	fmt.Println("   - NewExtensionForYYHertzContext(ctx)")
	fmt.Println("   - NewContextExtension(ctx) // 自动检测")
	
	fmt.Println("\n✅ 便利方法:")
	fmt.Println("   - ext.GetCookie(key)")
	fmt.Println("   - ext.SetCookie(name, value, ...)")
	fmt.Println("   - ext.GetSecureCookie(secret, key)")
	fmt.Println("   - ext.SetSecureCookie(secret, name, value, ...)")
	fmt.Println("   - ext.StartSession()")
	fmt.Println("   - ext.SetSession(key, value)")
	fmt.Println("   - ext.GetSession(key)")
	
	fmt.Println("\n🎯 架构优势总结")
	fmt.Println("==========================================")
	fmt.Println("✅ 模块化设计: session/cookie功能统一管理")
	fmt.Println("✅ 职责分离: context专注请求处理，session专注状态管理")
	fmt.Println("✅ 100%向后兼容: 现有代码无需修改")
	fmt.Println("✅ 类型安全: 完整的类型提示和IDE支持")
	fmt.Println("✅ 可扩展性: 便于添加新的存储后端")
	fmt.Println("✅ 高性能: 基于YYHertz高性能架构")
}

// CookieFeatureDemo Cookie功能专项演示
func CookieFeatureDemo() {
	fmt.Println("\n🍪 Cookie功能专项演示")
	fmt.Println("==========================================")
	
	// 创建Cookie操作器
	mockRequest := &app.RequestContext{}
	secureCookie := cookie.NewSecureCookie(mockRequest)
	
	fmt.Println("\n📋 功能1: 基础Cookie操作演示")
	fmt.Println("--------------------------------------")
	
	// 基础操作演示
	fmt.Println("✅ 基础Cookie操作:")
	fmt.Println(`
// 设置cookie
cookie.Set("username", "john", 3600, "/", "", false, true)

// 获取cookie
username := cookie.Get("username")

// 检查存在
exists := cookie.Exists("username")

// 删除cookie
cookie.Delete("username")

// 获取所有cookie
allCookies := cookie.GetAll()

// 清除所有cookie
cookie.Clear()`)
	
	fmt.Println("\n📋 功能2: 安全Cookie操作演示")
	fmt.Println("--------------------------------------")
	
	secret := "demo_secret_key_256bit"
	
	fmt.Println("✅ 安全Cookie操作:")
	fmt.Printf(`
// 设置安全cookie
secureCookie.SetSecure("%s", "user_token", "abc123", 3600)

// 获取安全cookie
token, valid := secureCookie.GetSecure("%s", "user_token")

// 验证安全cookie
isValid := secureCookie.Validate("%s", "user_token")
`, secret, secret, secret)
	
	// 实际演示
	secureCookie.SetSecure(secret, "demo_token", "demo_value_123", 3600)
	fmt.Println("✅ 已设置演示用安全Cookie")
	
	value, valid := secureCookie.GetSecure(secret, "demo_token")
	fmt.Printf("✅ 获取结果: value=%s, valid=%t\n", value, valid)
	
	fmt.Println("\n📋 功能3: 高级安全选项演示")
	fmt.Println("--------------------------------------")
	
	fmt.Println("✅ 高级安全配置:")
	fmt.Println(`
options := CookieSecurityOptions{
    Secret:          "your_secret_key",
    MaxAge:          time.Hour * 24,
    ValidateExpiry:  true,
    RequireHTTPS:    true, // 生产环境建议
}

err := secureCookie.SetSecureWithOptions("token", "value", options, 3600)
value, valid, err := secureCookie.GetSecureWithOptions("token", options)`)
	
	fmt.Println("\n📋 功能4: 工具函数演示")
	fmt.Println("--------------------------------------")
	
	fmt.Println("✅ Cookie工具函数:")
	
	// 演示解析cookie字符串
	cookieStr := "name1=value1; name2=value2; name3=value3"
	parsed := cookie.ParseCookieString(cookieStr)
	fmt.Printf("✅ 解析Cookie字符串: %s -> %d个cookie\n", cookieStr, len(parsed))
	
	// 演示格式化cookie
	testCookies := map[string]string{
		"session": "abc123",
		"theme":   "dark",
	}
	formatted := cookie.FormatCookieString(testCookies)
	fmt.Printf("✅ 格式化Cookie: %s\n", formatted)
	
	// 演示验证
	fmt.Printf("✅ 验证Cookie名称 'valid_name': %t\n", cookie.ValidateCookieName("valid_name"))
	fmt.Printf("✅ 验证Cookie名称 'invalid;name': %t\n", cookie.ValidateCookieName("invalid;name"))
	
	fmt.Println("\n📋 功能5: 安全工具函数演示")
	fmt.Println("--------------------------------------")
	
	fmt.Println("✅ 安全Cookie工具:")
	
	// 演示生成安全值
	testSecret := "test_secret"
	testValue := "test_data"
	secureValue := cookie.GenerateSecureValue(testSecret, testValue)
	fmt.Printf("✅ 生成安全值: %s -> %s\n", testValue, secureValue[:30]+"...")
	
	// 演示解析安全值
	parsedValue, isValid := cookie.ParseSecureValue(testSecret, secureValue)
	fmt.Printf("✅ 解析安全值: %s (valid: %t)\n", parsedValue, isValid)
	
	// 演示时间戳获取
	timestamp, _ := cookie.GetSecureCookieTimestamp(secureValue)
	fmt.Printf("✅ Cookie创建时间: %s\n", timestamp.Format("2006-01-02 15:04:05"))
}

// SessionFeatureDemo Session功能专项演示
func SessionFeatureDemo() {
	fmt.Println("\n🔧 Session功能专项演示")
	fmt.Println("==========================================")
	
	fmt.Println("\n📋 功能1: Session适配器演示")
	fmt.Println("--------------------------------------")
	
	// 创建Session Store和适配器
	store := NewMemoryStore("demo_session_123")
	adapter := NewAdapter(store, nil)
	
	fmt.Println("✅ Session适配器操作:")
	fmt.Printf(`
// 创建适配器
store := NewMemoryStore("session_123")
adapter := NewAdapter(store, ctx)

// 设置数据
adapter.Set("user_id", "12345")
adapter.Set("username", "john_doe")

// 获取数据
userID := adapter.Get("user_id")
username := adapter.Get("username")

// 检查存在
exists := adapter.Exists("user_id")

// 删除数据
adapter.Delete("user_id")

// 获取SessionID
sessionID := adapter.SessionID()

// 获取所有数据
allData := adapter.GetAll()

// 清空数据
adapter.Flush()

// 保存数据
adapter.Save()`)
	
	// 实际演示
	adapter.Set("demo_key", "demo_value")
	adapter.Set("user_info", map[string]string{
		"name": "Demo User",
		"role": "admin",
	})
	
	fmt.Printf("✅ SessionID: %s\n", adapter.SessionID())
	fmt.Printf("✅ 数据存在检查 'demo_key': %t\n", adapter.Exists("demo_key"))
	
	allData := adapter.GetAll()
	fmt.Printf("✅ 所有Session数据: %d个键值对\n", len(allData))
	
	fmt.Println("\n📋 功能2: Session管理器演示")
	fmt.Println("--------------------------------------")
	
	sessionMgr := NewSessionManagerFromConfig()
	
	fmt.Println("✅ Session管理器功能:")
	fmt.Printf(`
// 创建管理器
sessionMgr := NewSessionManagerFromConfig()

// 检查状态
enabled := sessionMgr.IsEnabled()

// 启用/禁用
sessionMgr.Enable()
sessionMgr.Disable()

// 创建Session
adapter := sessionMgr.CreateSession(ctx)

// 获取现有Session
adapter := sessionMgr.GetSession(ctx, sessionID)

// 获取配置
config := sessionMgr.GetConfig()`)
	
	fmt.Printf("✅ Session管理器状态: enabled=%t\n", sessionMgr.IsEnabled())
	
	config := sessionMgr.GetConfig()
	fmt.Printf("✅ Session配置: cookieName=%s, maxAge=%d\n", 
		config.CookieName, config.MaxAge)
	
	fmt.Println("\n📋 功能3: Context扩展Session演示")
	fmt.Println("--------------------------------------")
	
	mockRequest := &app.RequestContext{}
	ext := NewExtensionForHertzContext(mockRequest)
	
	fmt.Println("✅ Context扩展Session操作:")
	fmt.Printf(`
// 创建扩展
ext := NewExtensionForHertzContext(ctx)

// 启动Session
adapter := ext.StartSession()

// Session操作
ext.SetSession("user_id", "12345")
userID := ext.GetSession("user_id")
ext.DelSession("user_id")

// Session状态
sessionID := ext.GetSessionID()
started := ext.IsSessionStarted()

// Session管理
ext.ClearSession()      // 清空数据
ext.SaveSession()       // 保存数据
ext.DestroySession()    // 销毁Session
ext.SessionRegenerateID() // 重新生成ID`)
	
	// 实际演示
	ext.StartSession()
	ext.SetSession("demo_user", "context_demo")
	ext.SetSession("demo_time", time.Now().Format("15:04:05"))
	
	fmt.Printf("✅ Session启动状态: %t\n", ext.IsSessionStarted())
	fmt.Printf("✅ SessionID: %s\n", ext.GetSessionID())
	
	demoUser := ext.GetSession("demo_user")
	demoTime := ext.GetSession("demo_time")
	fmt.Printf("✅ Session数据: user=%v, time=%v\n", demoUser, demoTime)
}

// BeegoMigrationDemo beego迁移演示
func BeegoMigrationDemo() {
	fmt.Println("\n📚 Beego迁移演示")
	fmt.Println("==========================================")
	
	fmt.Println("\n📋 迁移对比: Beego -> YYHertz")
	fmt.Println("--------------------------------------")
	
	fmt.Println("✅ Beego原始代码:")
	fmt.Println(`
// Beego Controller
type MainController struct {
    beego.Controller
}

func (c *MainController) Get() {
    // Cookie操作
    c.Ctx.SetCookie("username", "john", 3600, "/", "", false, true)
    username := c.Ctx.GetCookie("username")
    
    // 安全Cookie操作
    secret := "my_secret"
    c.Ctx.SetSecureCookie(secret, "user_id", "123", 3600)
    userID, valid := c.Ctx.GetSecureCookie(secret, "user_id")
    
    // Session操作
    c.SetSession("user_info", "user_data")
    userInfo := c.GetSession("user_info")
    sessionID := c.CruSession.SessionID()
}`)
	
	fmt.Println("\n✅ YYHertz迁移后代码:")
	fmt.Println(`
// YYHertz Controller
type MainController struct {
    // 可以直接使用Context或通过扩展
}

func (c *MainController) Get() {
    // 方式1: 直接使用Context (100%兼容)
    c.Ctx.SetCookie("username", "john", 3600, "/", "", false, true)
    username := c.Ctx.GetCookie("username")
    
    // 方式2: 使用Session扩展 (推荐)
    ext := session.NewContextExtension(c.Ctx)
    
    // Cookie操作 - 完全兼容！
    ext.SetCookie("username", "john", 3600, "/", "", false, true)
    username := ext.GetCookie("username")
    
    // 安全Cookie操作 - 完全兼容！
    secret := "my_secret"
    ext.SetSecureCookie(secret, "user_id", "123", 3600)
    userID, valid := ext.GetSecureCookie(secret, "user_id")
    
    // Session操作 - 增强功能！
    ext.StartSession()
    ext.SetSession("user_info", "user_data")
    userInfo := ext.GetSession("user_info")
    sessionID := ext.GetSessionID()
    
    // 新增高级功能
    options := session.CookieSecurityOptions{
        Secret:          secret,
        MaxAge:          time.Hour * 24,
        ValidateExpiry:  true,
        RequireHTTPS:    true,
    }
    ext.SecureCookie.SetSecureWithOptions("token", "value", options, 3600)
}`)
	
	fmt.Println("\n🎯 迁移优势:")
	fmt.Println("--------------------------------------")
	fmt.Println("✅ 零代码修改 - 完全向后兼容")
	fmt.Println("✅ 性能大幅提升 - YYHertz高性能架构")
	fmt.Println("✅ 功能增强 - 新增安全选项和工具方法")
	fmt.Println("✅ 模块化设计 - 更好的代码组织")
	fmt.Println("✅ 类型安全 - 完整的类型提示")
	fmt.Println("✅ 扩展性强 - 支持更多自定义功能")
	fmt.Println("✅ 统一管理 - session/cookie功能集中")
	
	fmt.Println("\n📈 使用建议:")
	fmt.Println("--------------------------------------")
	fmt.Println("✅ 新项目: 直接使用session.NewContextExtension()")
	fmt.Println("✅ 迁移项目: 保持原有代码，逐步采用新功能")
	fmt.Println("✅ 高安全需求: 使用CookieSecurityOptions")
	fmt.Println("✅ 大型项目: 使用SessionManager统一管理")
	
	fmt.Println("\n🚀 立即开始使用YYHertz统一Session模块！")
}

// FullDemo 完整功能演示
func FullDemo() {
	fmt.Println("🌟 YYHertz Session & Cookie 完整功能演示")
	fmt.Println("============================================")
	
	// 运行各个演示
	SessionCookieUnifiedDemo()
	CookieFeatureDemo()
	SessionFeatureDemo()
	BeegoMigrationDemo()
	
	fmt.Println("\n🎉 演示完成！")
	fmt.Println("============================================")
	fmt.Println("感谢使用YYHertz统一Session & Cookie模块！")
	fmt.Println("更多信息请查看文档和测试用例。")
}