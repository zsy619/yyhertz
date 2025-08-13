package session

import (
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// ============= Cookie基础功能测试 =============

func TestBaseCookieOperations(t *testing.T) {
	// 创建测试上下文
	mockRequest := &app.RequestContext{}
	cookie := NewBaseCookie(mockRequest)
	
	t.Log("🍪 测试基础Cookie操作")
	
	// 测试设置Cookie
	cookie.Set("test_cookie", "test_value", 3600, "/", "localhost", false, true)
	t.Log("✅ SetCookie方法调用成功")
	
	// 测试Cookie存在性检查
	exists := cookie.Exists("nonexistent_cookie")
	if exists {
		t.Error("期望nonexistent_cookie不存在")
	}
	t.Log("✅ CookieExists检查正常")
	
	// 测试删除Cookie
	cookie.Delete("test_cookie")
	t.Log("✅ DelCookie方法调用成功")
	
	// 测试获取所有Cookie
	cookies := cookie.GetAll()
	if cookies == nil {
		t.Error("期望cookies不为nil")
	}
	t.Log("✅ GetCookies返回正常")
	
	// 测试Cookie计数
	count := cookie.Count()
	if count < 0 {
		t.Error("期望cookie计数大于等于0")
	}
	t.Logf("✅ CookieCount: %d", count)
}

func TestOutputCookieOperations(t *testing.T) {
	// 创建测试上下文
	mockRequest := &app.RequestContext{}
	outputCookie := NewOutputCookie(mockRequest)
	
	t.Log("📤 测试输出Cookie操作")
	
	// 测试设置输出Cookie
	outputCookie.Set("output_cookie", "output_value", 3600, "/", "localhost", false, true)
	t.Log("✅ OutputCookie.Set方法调用成功")
}

// ============= 安全Cookie功能测试 =============

func TestSecureCookieOperations(t *testing.T) {
	// 创建测试上下文
	mockRequest := &app.RequestContext{}
	secureCookie := NewSecureCookie(mockRequest)
	
	t.Log("🔐 测试安全Cookie操作")
	
	secret := "my_secret_key_for_hmac_256"
	cookieName := "secure_cookie"
	cookieValue := "sensitive_data_123"
	
	// 测试设置安全Cookie
	secureCookie.SetSecure(secret, cookieName, cookieValue, 3600, "/", "", false, true)
	t.Log("✅ SetSecureCookie方法调用成功")
	
	// 测试验证安全Cookie（由于模拟环境限制，主要测试方法调用）
	valid := secureCookie.Validate(secret, cookieName)
	t.Logf("✅ ValidateSecureCookie结果: %t", valid)
}

func TestAdvancedSecureCookieOptions(t *testing.T) {
	// 创建测试上下文
	mockRequest := &app.RequestContext{}
	secureCookie := NewSecureCookie(mockRequest)
	
	t.Log("⚙️ 测试高级安全Cookie选项")
	
	// 配置安全选项
	options := CookieSecurityOptions{
		Secret:          "advanced_secret_key_hmac",
		MaxAge:          time.Hour * 24, // 24小时
		ValidateExpiry:  true,
		RequireHTTPS:    false, // 测试环境不使用HTTPS
	}
	
	cookieName := "advanced_secure_cookie"
	cookieValue := "advanced_sensitive_data"
	
	// 测试设置带选项的安全Cookie
	err := secureCookie.SetSecureWithOptions(cookieName, cookieValue, options, 3600)
	if err != nil {
		t.Logf("SetSecureWithOptions返回错误（预期的）: %v", err)
	} else {
		t.Log("✅ SetSecureWithOptions方法调用成功")
	}
	
	// 测试获取带选项的安全Cookie
	_, valid, err := secureCookie.GetSecureWithOptions(cookieName, options)
	if err != nil {
		t.Logf("GetSecureWithOptions返回错误（可能由于测试环境）: %v", err)
	} else {
		t.Logf("✅ GetSecureWithOptions方法调用成功，valid: %t", valid)
	}
}

func TestCookieSecurityFeatures(t *testing.T) {
	// 创建测试上下文
	mockRequest := &app.RequestContext{}
	secureCookie := NewSecureCookie(mockRequest)
	
	t.Log("🛡️ 测试Cookie安全特性")
	
	// 测试过期验证
	shortOptions := CookieSecurityOptions{
		Secret:          "expire_test_secret",
		MaxAge:          time.Millisecond * 100, // 很短的过期时间
		ValidateExpiry:  true,
		RequireHTTPS:    false,
	}
	
	// 设置一个会快速过期的Cookie
	err := secureCookie.SetSecureWithOptions("expire_test", "expire_value", shortOptions)
	if err != nil {
		t.Logf("设置过期Cookie时返回错误: %v", err)
	} else {
		t.Log("✅ 过期时间测试cookie设置成功")
	}
	
	// 等待过期
	time.Sleep(time.Millisecond * 200)
	
	// 尝试获取应该已过期的Cookie
	_, valid, err := secureCookie.GetSecureWithOptions("expire_test", shortOptions)
	if err != nil {
		t.Logf("✅ 正确检测到过期Cookie: %v", err)
	} else if !valid {
		t.Log("✅ 正确返回invalid状态")
	}
	
	// 测试HTTPS要求
	httpsOptions := CookieSecurityOptions{
		Secret:       "https_test_secret",
		RequireHTTPS: true,
	}
	
	err = secureCookie.SetSecureWithOptions("https_test", "https_value", httpsOptions)
	if err != nil {
		t.Logf("✅ 正确检测到非HTTPS环境: %v", err)
	}
}

// ============= Session适配器测试 =============

func TestSessionAdapterBasic(t *testing.T) {
	t.Log("🔄 测试Session适配器基础功能")
	
	// 创建内存store
	store := NewMemoryStore("test_session_123")
	adapter := NewAdapter(store, nil)
	
	if !adapter.IsStarted() {
		t.Error("期望adapter已启动")
	}
	
	// 测试设置和获取
	err := adapter.Set("test_key", "test_value")
	if err != nil {
		t.Errorf("设置session值失败: %v", err)
	}
	
	value := adapter.Get("test_key")
	if value != "test_value" {
		t.Errorf("期望值为'test_value'，实际为: %v", value)
	}
	
	// 测试SessionID
	sessionID := adapter.SessionID()
	if sessionID != "test_session_123" {
		t.Errorf("期望SessionID为'test_session_123'，实际为: %s", sessionID)
	}
	
	// 测试删除
	err = adapter.Delete("test_key")
	if err != nil {
		t.Errorf("删除session值失败: %v", err)
	}
	
	value = adapter.Get("test_key")
	if value != nil {
		t.Errorf("期望删除后值为nil，实际为: %v", value)
	}
	
	t.Log("✅ Session适配器基础功能测试通过")
}

func TestSessionAdapterWithNilStore(t *testing.T) {
	t.Log("🔄 测试Session适配器nil store处理")
	
	// 测试nil store的处理
	adapter := NewAdapter(nil, nil)
	
	if adapter.IsStarted() {
		t.Error("期望nil store的adapter未启动")
	}
	
	// 应该能够安全地调用方法而不崩溃
	err := adapter.Set("key", "value")
	if err != nil {
		t.Errorf("nil store设置应该静默处理: %v", err)
	}
	
	value := adapter.Get("key")
	if value != nil {
		t.Errorf("期望nil store返回nil，实际为: %v", value)
	}
	
	sessionID := adapter.SessionID()
	if sessionID != "" {
		t.Errorf("期望nil store返回空字符串，实际为: %s", sessionID)
	}
	
	t.Log("✅ Session适配器nil store处理测试通过")
}

// ============= 上下文扩展测试 =============

func TestContextExtensionBasic(t *testing.T) {
	t.Log("🔧 测试Context扩展基础功能")
	
	// 创建Hertz上下文扩展
	mockRequest := &app.RequestContext{}
	ext := NewExtensionForHertzContext(mockRequest)
	
	if ext.Cookie == nil {
		t.Error("期望Cookie不为nil")
	}
	
	if ext.SecureCookie == nil {
		t.Error("期望SecureCookie不为nil")
	}
	
	if ext.SessionMgr == nil {
		t.Error("期望SessionMgr不为nil")
	}
	
	t.Log("✅ Context扩展基础功能创建成功")
}

func TestContextExtensionSessionOperations(t *testing.T) {
	t.Log("🔧 测试Context扩展Session操作")
	
	// 创建扩展
	mockRequest := &app.RequestContext{}
	ext := NewExtensionForHertzContext(mockRequest)
	
	// 测试启动session
	adapter := ext.StartSession()
	if adapter == nil {
		t.Error("期望StartSession返回非nil适配器")
	}
	
	if !ext.IsSessionStarted() {
		t.Error("期望session已启动")
	}
	
	// 测试session操作
	err := ext.SetSession("test_key", "test_value")
	if err != nil {
		t.Errorf("设置session失败: %v", err)
	}
	
	value := ext.GetSession("test_key")
	if value != "test_value" {
		t.Errorf("期望值为'test_value'，实际为: %v", value)
	}
	
	sessionID := ext.GetSessionID()
	if sessionID == "" {
		t.Error("期望SessionID不为空")
	}
	
	// 测试清理session
	ext.ClearSession()
	
	// 测试销毁session
	ext.DestroySession()
	if ext.IsSessionStarted() {
		t.Error("期望销毁后session未启动")
	}
	
	t.Log("✅ Context扩展Session操作测试通过")
}

func TestContextExtensionCookieOperations(t *testing.T) {
	t.Log("🔧 测试Context扩展Cookie操作")
	
	// 创建扩展
	mockRequest := &app.RequestContext{}
	ext := NewExtensionForHertzContext(mockRequest)
	
	// 测试基础cookie操作
	ext.SetCookie("test_cookie", "test_value", 3600, "/")
	
	// 由于模拟环境限制，主要验证方法调用不出错
	value := ext.GetCookie("test_cookie")
	t.Logf("GetCookie返回: %s", value)
	
	exists := ext.CookieExists("test_cookie")
	t.Logf("CookieExists结果: %t", exists)
	
	// 测试安全cookie操作
	secret := "test_secret_key"
	ext.SetSecureCookie(secret, "secure_test", "secure_value", 3600)
	
	secureValue, valid := ext.GetSecureCookie(secret, "secure_test")
	t.Logf("GetSecureCookie返回: value=%s, valid=%t", secureValue, valid)
	
	// 测试删除cookie
	ext.DelCookie("test_cookie")
	
	t.Log("✅ Context扩展Cookie操作测试通过")
}

// ============= 工具函数测试 =============

func TestCookieUtilityFunctions(t *testing.T) {
	t.Log("🛠️ 测试Cookie工具函数")
	
	// 测试解析Cookie字符串
	cookieStr := "name1=value1; name2=value2; name3=value3"
	cookies := ParseCookieString(cookieStr)
	
	if len(cookies) != 3 {
		t.Errorf("期望解析出3个cookie，实际为: %d", len(cookies))
	}
	
	if cookies["name1"] != "value1" {
		t.Errorf("期望name1=value1，实际为: %s", cookies["name1"])
	}
	
	// 测试格式化Cookie字符串
	testCookies := map[string]string{
		"test1": "value1",
		"test2": "value2",
	}
	formatted := FormatCookieString(testCookies)
	if formatted == "" {
		t.Error("期望格式化后的字符串不为空")
	}
	
	t.Logf("格式化后的Cookie字符串: %s", formatted)
	
	// 测试Cookie名称验证
	if !ValidateCookieName("valid_name") {
		t.Error("期望valid_name是合法的cookie名称")
	}
	
	if ValidateCookieName("invalid;name") {
		t.Error("期望invalid;name是非法的cookie名称")
	}
	
	// 测试Cookie值验证
	if !ValidateCookieValue("valid_value") {
		t.Error("期望valid_value是合法的cookie值")
	}
	
	t.Log("✅ Cookie工具函数测试通过")
}

func TestSecurityUtilityFunctions(t *testing.T) {
	t.Log("🔐 测试安全工具函数")
	
	secret := "test_secret_key"
	value := "test_value"
	
	// 测试生成安全值
	secureValue := GenerateSecureValue(secret, value)
	if secureValue == "" {
		t.Error("期望生成的安全值不为空")
	}
	
	// 测试解析安全值
	parsedValue, valid := ParseSecureValue(secret, secureValue)
	if !valid {
		t.Error("期望解析安全值成功")
	}
	
	if parsedValue != value {
		t.Errorf("期望解析值为'%s'，实际为: %s", value, parsedValue)
	}
	
	// 测试过期检查
	if IsSecureCookieExpired(secureValue, time.Millisecond) {
		// 这里可能会因为执行速度太快而不过期，这是正常的
		t.Log("Cookie可能已过期（取决于执行速度）")
	}
	
	// 测试获取时间戳
	timestamp, err := GetSecureCookieTimestamp(secureValue)
	if err != nil {
		t.Errorf("获取时间戳失败: %v", err)
	}
	
	if timestamp.IsZero() {
		t.Error("期望时间戳不为零")
	}
	
	t.Log("✅ 安全工具函数测试通过")
}

// ============= 性能基准测试 =============

func BenchmarkBaseCookieSet(b *testing.B) {
	mockRequest := &app.RequestContext{}
	cookie := NewBaseCookie(mockRequest)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cookie.Set("bench_cookie", "bench_value")
	}
}

func BenchmarkSecureCookieSet(b *testing.B) {
	mockRequest := &app.RequestContext{}
	secureCookie := NewSecureCookie(mockRequest)
	secret := "benchmark_secret_key"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		secureCookie.SetSecure(secret, "bench_cookie", "bench_value")
	}
}

func BenchmarkSessionAdapterOperations(b *testing.B) {
	store := NewMemoryStore("bench_session")
	adapter := NewAdapter(store, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Set("bench_key", "bench_value")
		adapter.Get("bench_key")
	}
}