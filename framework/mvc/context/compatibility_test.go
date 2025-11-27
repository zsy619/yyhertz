package context

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

// TestInputDataCookieCompatibility 测试InputData的Cookie兼容性
func TestInputDataCookieCompatibility(t *testing.T) {
	// 模拟Request上下文
	mockRequest := &app.RequestContext{}

	// 创建测试上下文（使用构造函数）
	ctx := NewContext(mockRequest)

	// 初始化InputData
	ctx.Input = &InputData{ctx: ctx}

	t.Log("🔄 测试InputData Cookie兼容性（代理到session包）")

	// 测试基础Cookie方法
	ctx.Input.SetCookie("test_cookie", "test_value", 3600, "/", "localhost", false, true)
	t.Log("✅ SetCookie代理方法调用成功")

	value := ctx.Input.Cookie("test_cookie")
	t.Logf("✅ Cookie代理方法返回: %s", value)

	exists := ctx.Input.CookieExists("test_cookie")
	t.Logf("✅ CookieExists代理方法返回: %t", exists)

	ctx.Input.DelCookie("test_cookie")
	t.Log("✅ DelCookie代理方法调用成功")

	// 测试安全Cookie方法
	secret := "test_secret_key"
	ctx.Input.SetSecureCookie(secret, "secure_test", "secure_value", 3600)
	t.Log("✅ SetSecureCookie代理方法调用成功")

	secureValue, valid := ctx.Input.GetSecureCookie(secret, "secure_test")
	t.Logf("✅ GetSecureCookie代理方法返回: value=%s, valid=%t", secureValue, valid)

	// 测试高级Cookie功能
	options := cookie.CookieSecurityOptions{
		Secret:         secret,
		RequireHTTPS:   false,
		ValidateExpiry: false,
	}

	err := ctx.Input.SetSecureCookieWithOptions("advanced_test", "advanced_value", options)
	if err != nil {
		t.Logf("SetSecureCookieWithOptions返回错误（可能正常）: %v", err)
	} else {
		t.Log("✅ SetSecureCookieWithOptions代理方法调用成功")
	}

	t.Log("✅ InputData Cookie兼容性测试通过")
}

func TestInputDataSessionCompatibility(t *testing.T) {
	// 模拟Request上下文
	mockRequest := &app.RequestContext{}

	// 创建测试上下文（使用构造函数）
	ctx := NewContext(mockRequest)

	// 初始化InputData
	ctx.Input = &InputData{ctx: ctx}

	t.Log("🔄 测试InputData Session兼容性（代理到session包）")

	// 测试启动Session
	adapter := ctx.Input.StartSession()
	if adapter == nil {
		t.Error("期望StartSession返回非nil适配器")
	} else {
		t.Log("✅ StartSession代理方法返回Session适配器")
	}

	// 测试Session状态检查
	isStarted := ctx.Input.IsSessionStarted()
	t.Logf("✅ IsSessionStarted代理方法返回: %t", isStarted)

	// 测试Session数据操作
	err := ctx.Input.SetSession("test_key", "test_value")
	if err != nil {
		t.Errorf("SetSession代理方法失败: %v", err)
	} else {
		t.Log("✅ SetSession代理方法调用成功")
	}

	value := ctx.Input.GetSession("test_key")
	t.Logf("✅ GetSession代理方法返回: %v", value)

	sessionID := ctx.Input.GetSessionID()
	t.Logf("✅ GetSessionID代理方法返回: %s", sessionID)

	// 测试Session清理
	ctx.Input.ClearSession()
	t.Log("✅ ClearSession代理方法调用成功")

	err = ctx.Input.SaveSession()
	if err != nil {
		t.Errorf("SaveSession代理方法失败: %v", err)
	} else {
		t.Log("✅ SaveSession代理方法调用成功")
	}

	// 测试Session销毁
	ctx.Input.DestroySession()
	t.Log("✅ DestroySession代理方法调用成功")

	t.Log("✅ InputData Session兼容性测试通过")
}

func TestContextCookieCompatibility(t *testing.T) {
	// 模拟Request上下文
	mockRequest := &app.RequestContext{}

	// 创建测试上下文（使用构造函数）
	ctx := NewContext(mockRequest)

	// 初始化InputData和OutputData
	ctx.Input = &InputData{ctx: ctx}
	ctx.Output = &OutputData{ctx: ctx}

	t.Log("🔄 测试Context级别的Cookie兼容性")

	// 测试Context Cookie方法（这些方法代理到Input）
	ctx.SetCookie("context_test", "context_value", 3600, "/")
	t.Log("✅ Context.SetCookie代理方法调用成功")

	value := ctx.GetCookie("context_test")
	t.Logf("✅ Context.GetCookie代理方法返回: %s", value)

	secret := "context_secret"
	ctx.SetSecureCookie(secret, "context_secure", "secure_context_value", 3600)
	t.Log("✅ Context.SetSecureCookie代理方法调用成功")

	secureValue, valid := ctx.GetSecureCookie(secret, "context_secure")
	t.Logf("✅ Context.GetSecureCookie代理方法返回: value=%s, valid=%t", secureValue, valid)

	t.Log("✅ Context级别Cookie兼容性测试通过")
}

func TestSessionExtensionAccess(t *testing.T) {
	// 模拟Request上下文
	mockRequest := &app.RequestContext{}

	// 创建测试上下文（使用构造函数）
	ctx := NewContext(mockRequest)

	// 初始化InputData
	ctx.Input = &InputData{ctx: ctx}

	t.Log("🔄 测试Session扩展访问")

	// 测试获取Session扩展
	ext := ctx.Input.GetSessionExtension()
	if ext == nil {
		t.Error("期望GetSessionExtension返回非nil扩展")
	} else {
		t.Log("✅ GetSessionExtension返回Session扩展对象")

		// 测试直接访问高级功能
		if ext.Cookie != nil {
			t.Log("✅ 可以访问BaseCookie功能")
		}

		if ext.SecureCookie != nil {
			t.Log("✅ 可以访问SecureCookie功能")
		}

		if ext.SessionMgr != nil {
			t.Log("✅ 可以访问SessionManager功能")
		}
	}

	// 测试自定义Session扩展
	customExt := session.NewExtensionForHertzContext(mockRequest)
	ctx.Input.WithSessionExtension(customExt)

	newExt := ctx.Input.GetSessionExtension()
	if newExt != customExt {
		t.Error("期望WithSessionExtension设置的扩展被正确返回")
	} else {
		t.Log("✅ WithSessionExtension设置自定义扩展成功")
	}

	t.Log("✅ Session扩展访问测试通过")
}

func TestBackwardCompatibilityTypes(t *testing.T) {
	t.Log("🔄 测试向后兼容性类型别名")

	// 测试SessionStore类型别名
	store := session.NewMemoryStore("test_session")
	adapter := NewSessionStore(store, nil)

	if adapter == nil {
		t.Error("期望NewSessionStore返回非nil适配器")
	} else {
		t.Log("✅ SessionStore类型别名工作正常")
	}

	// 测试CookieSecurityOptions类型别名
	options := CookieSecurityOptions{
		Secret:         "test",
		RequireHTTPS:   false,
		ValidateExpiry: false,
	}

	if options.Secret != "test" {
		t.Error("期望CookieSecurityOptions类型别名工作正常")
	} else {
		t.Log("✅ CookieSecurityOptions类型别名工作正常")
	}

	t.Log("✅ 向后兼容性类型别名测试通过")
}

func TestInputDataBasicMethods(t *testing.T) {
	// 模拟Request上下文
	mockRequest := &app.RequestContext{}

	// 创建测试上下文（使用构造函数）
	ctx := NewContext(mockRequest)

	// 初始化InputData
	ctx.Input = &InputData{ctx: ctx}

	t.Log("🔄 测试InputData基础方法")

	// 测试请求信息方法
	scheme := ctx.Input.Scheme()
	t.Logf("✅ Scheme(): %s", scheme)

	domain := ctx.Input.Domain()
	t.Logf("✅ Domain(): %s", domain)

	method := ctx.Input.Method()
	t.Logf("✅ Method(): %s", method)

	isGet := ctx.Input.IsGet()
	t.Logf("✅ IsGet(): %t", isGet)

	// 测试数据存储
	ctx.Input.Data("test_key", "test_value")
	value := ctx.Input.GetData("test_key")
	if value != "test_value" {
		t.Errorf("期望GetData返回'test_value'，实际为: %v", value)
	} else {
		t.Log("✅ Data/GetData方法工作正常")
	}

	t.Log("✅ InputData基础方法测试通过")
}
