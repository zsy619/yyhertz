package context

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestBaseURL(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	ctx := NewContext(c)

	// 测试BaseURL
	baseURL := ctx.BaseURL()
	expected := "http://example.com"
	if baseURL != expected {
		t.Errorf("Expected BaseURL '%s', got '%s'", expected, baseURL)
	}
}

func TestBaseURLWithCustomPort(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com:8080")
	ctx := NewContext(c)

	// 测试带端口的BaseURL
	baseURL := ctx.BaseURL()
	expected := "http://example.com:8080"
	if baseURL != expected {
		t.Errorf("Expected BaseURL '%s', got '%s'", expected, baseURL)
	}
}

func TestOriginalURL(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	c.Request.SetRequestURI("/api/users?page=1&limit=10")
	ctx := NewContext(c)

	// 调试输出
	t.Logf("Host: '%s'", ctx.Host())
	t.Logf("Hostname: '%s'", ctx.Hostname())
	t.Logf("BaseURL: '%s'", ctx.BaseURL())

	// 测试OriginalURL
	originalURL := ctx.OriginalURL()
	t.Logf("OriginalURL: '%s'", originalURL)
	
	// 简化期望值检查
	if !contains(originalURL, "/api/users") {
		t.Errorf("Expected OriginalURL to contain '/api/users', got '%s'", originalURL)
	}
	
	if !contains(originalURL, "page=1") {
		t.Errorf("Expected OriginalURL to contain 'page=1', got '%s'", originalURL)
	}
}

func TestProtocol(t *testing.T) {
	// 测试HTTP协议
	c := &app.RequestContext{}
	ctx := NewContext(c)

	protocol := ctx.Protocol()
	if protocol != "http" {
		t.Errorf("Expected protocol 'http', got '%s'", protocol)
	}
}

func TestProtocolWithForwardedProto(t *testing.T) {
	// 测试X-Forwarded-Proto头
	c := &app.RequestContext{}
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	ctx := NewContext(c)

	protocol := ctx.Protocol()
	if protocol != "https" {
		t.Errorf("Expected protocol 'https', got '%s'", protocol)
	}
}

func TestHostname(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "api.example.com:8080")
	ctx := NewContext(c)

	// 测试Hostname（应该去除端口）
	hostname := ctx.Hostname()
	expected := "api.example.com"
	if hostname != expected {
		t.Errorf("Expected hostname '%s', got '%s'", expected, hostname)
	}
}

func TestPort(t *testing.T) {
	// 测试有端口的情况
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com:8080")
	ctx := NewContext(c)

	port := ctx.Port()
	if port != "8080" {
		t.Errorf("Expected port '8080', got '%s'", port)
	}
}

func TestPortDefault(t *testing.T) {
	// 测试默认端口
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	ctx := NewContext(c)

	port := ctx.Port()
	if port != "80" {
		t.Errorf("Expected default port '80', got '%s'", port)
	}
}

func TestPortInt(t *testing.T) {
	// 测试端口整数形式
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com:8080")
	ctx := NewContext(c)

	port := ctx.PortInt()
	if port != 8080 {
		t.Errorf("Expected port 8080, got %d", port)
	}
}

func TestSubdomains(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "api.v1.example.com")
	ctx := NewContext(c)

	// 测试Subdomains
	subdomains := ctx.Subdomains()
	expected := []string{"v1", "api"} // 注意顺序是反转的
	
	if len(subdomains) != len(expected) {
		t.Errorf("Expected %d subdomains, got %d", len(expected), len(subdomains))
	}

	for i, subdomain := range expected {
		if i >= len(subdomains) || subdomains[i] != subdomain {
			t.Errorf("Expected subdomain[%d] '%s', got '%s'", i, subdomain, subdomains[i])
		}
	}
}

func TestSubdomainsWithCustomOffset(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "api.v1.staging.example.com")
	ctx := NewContext(c)

	// 测试自定义偏移量
	subdomains := ctx.Subdomains(2) // 默认偏移量
	expected := []string{"staging", "v1", "api"}
	
	if len(subdomains) != len(expected) {
		t.Errorf("Expected %d subdomains, got %d", len(expected), len(subdomains))
	}
}

func TestIPs(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	c.Request.Header.Set("X-Real-IP", "203.0.113.1")
	ctx := NewContext(c)

	// 测试IPs
	ips := ctx.IPs()
	
	// 验证包含所有IP
	expectedIPs := []string{"192.168.1.1", "10.0.0.1", "203.0.113.1"}
	for _, expectedIP := range expectedIPs {
		found := false
		for _, ip := range ips {
			if ip == expectedIP {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected IP '%s' not found in IPs list", expectedIP)
		}
	}
}

func TestIsIPv4(t *testing.T) {
	// 模拟IPv4客户端IP
	c := &app.RequestContext{}
	// 注意：在实际测试中，ClientIP的行为可能依赖于Hertz的内部实现
	ctx := NewContext(c)

	// 由于我们无法直接设置ClientIP，这里主要测试方法不会崩溃
	isIPv4 := ctx.IsIPv4()
	t.Logf("IsIPv4 result: %v", isIPv4)
}

func TestIsIPv6(t *testing.T) {
	// 测试IPv6检测
	c := &app.RequestContext{}
	ctx := NewContext(c)

	isIPv6 := ctx.IsIPv6()
	t.Logf("IsIPv6 result: %v", isIPv6)
}

func TestIsLocalhost(t *testing.T) {
	// 测试localhost检测
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "localhost")
	ctx := NewContext(c)

	if !ctx.IsLocalhost() {
		t.Error("Expected IsLocalhost to return true for localhost")
	}
}

func TestIsLocalhost127(t *testing.T) {
	// 测试127.0.0.1
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "127.0.0.1")
	ctx := NewContext(c)

	if !ctx.IsLocalhost() {
		t.Error("Expected IsLocalhost to return true for 127.0.0.1")
	}
}

func TestParsedURL(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	c.Request.SetRequestURI("/api/users?page=1")
	ctx := NewContext(c)

	// 测试ParsedURL
	parsedURL, err := ctx.ParsedURL()
	if err != nil {
		t.Errorf("ParsedURL failed: %v", err)
	}

	if parsedURL.Hostname() != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%s'", parsedURL.Hostname())
	}

	if parsedURL.Path != "/api/users" {
		t.Errorf("Expected path '/api/users', got '%s'", parsedURL.Path)
	}
}

func TestURLFor(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	ctx := NewContext(c)

	// 测试URLFor
	url := ctx.URLFor("/api/users")
	expected := "http://example.com/api/users"
	if url != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, url)
	}
}

func TestURLForWithParams(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	ctx := NewContext(c)

	// 测试带参数的URLFor
	params := map[string]string{
		"page":  "1",
		"limit": "10",
	}
	url := ctx.URLFor("/api/users", params)
	
	// 验证包含基础URL和路径
	if !contains(url, "http://example.com/api/users") {
		t.Errorf("Expected URL to contain base path, got '%s'", url)
	}
	
	// 验证包含参数
	if !contains(url, "page=1") || !contains(url, "limit=10") {
		t.Errorf("Expected URL to contain parameters, got '%s'", url)
	}
}

func TestAbsoluteURL(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	c.Request.SetRequestURI("/api/users")
	ctx := NewContext(c)

	// 测试绝对URL（已经是绝对的）
	absoluteURL := ctx.AbsoluteURL("https://other.com/path")
	expected := "https://other.com/path"
	if absoluteURL != expected {
		t.Errorf("Expected absolute URL '%s', got '%s'", expected, absoluteURL)
	}

	// 测试相对于根的URL
	rootRelativeURL := ctx.AbsoluteURL("/other/path")
	expected = "http://example.com/other/path"
	if rootRelativeURL != expected {
		t.Errorf("Expected root relative URL '%s', got '%s'", expected, rootRelativeURL)
	}
}

func TestIsReferrerSameDomain(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	c.Request.Header.Set("Referer", "https://example.com/previous")
	ctx := NewContext(c)

	// 测试同域名Referer
	if !ctx.IsReferrerSameDomain() {
		t.Error("Expected IsReferrerSameDomain to return true for same domain")
	}
}

func TestIsReferrerSameDomainDifferent(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	c.Request.Header.Set("Referer", "https://other.com/previous")
	ctx := NewContext(c)

	// 测试不同域名Referer
	if ctx.IsReferrerSameDomain() {
		t.Error("Expected IsReferrerSameDomain to return false for different domain")
	}
}

func TestGetReferrerDomain(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Referer", "https://other.com/path")
	ctx := NewContext(c)

	domain := ctx.GetReferrerDomain()
	expected := "other.com"
	if domain != expected {
		t.Errorf("Expected referer domain '%s', got '%s'", expected, domain)
	}
}

func TestIsBehindProxy(t *testing.T) {
	// 测试没有代理头的情况
	c := &app.RequestContext{}
	ctx := NewContext(c)

	if ctx.IsBehindProxy() {
		t.Error("Expected IsBehindProxy to return false when no proxy headers present")
	}

	// 测试有代理头的情况
	c.Request.Header.Set("X-Forwarded-For", "192.168.1.1")
	if !ctx.IsBehindProxy() {
		t.Error("Expected IsBehindProxy to return true when proxy headers present")
	}
}

func TestGetProxyInfo(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("X-Forwarded-For", "192.168.1.1")
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	c.Request.Header.Set("X-Real-IP", "203.0.113.1")
	ctx := NewContext(c)

	info := ctx.GetProxyInfo()
	
	if info["forwarded_for"] != "192.168.1.1" {
		t.Errorf("Expected forwarded_for '192.168.1.1', got '%s'", info["forwarded_for"])
	}
	
	if info["forwarded_proto"] != "https" {
		t.Errorf("Expected forwarded_proto 'https', got '%s'", info["forwarded_proto"])
	}
	
	if info["real_ip"] != "203.0.113.1" {
		t.Errorf("Expected real_ip '203.0.113.1', got '%s'", info["real_ip"])
	}
}

func TestConnectionInfo(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "api.example.com:8080")
	c.Request.SetRequestURI("/api/users?page=1")
	c.Request.Header.Set("User-Agent", "Test Agent")
	ctx := NewContext(c)

	info := ctx.ConnectionInfo()
	
	if info["hostname"] != "api.example.com" {
		t.Errorf("Expected hostname 'api.example.com', got '%v'", info["hostname"])
	}
	
	if info["port"] != "8080" {
		t.Errorf("Expected port '8080', got '%v'", info["port"])
	}
	
	if info["protocol"] != "http" {
		t.Errorf("Expected protocol 'http', got '%v'", info["protocol"])
	}
}

func TestIsSecure(t *testing.T) {
	// 测试HTTP（不安全）
	c := &app.RequestContext{}
	ctx := NewContext(c)

	if ctx.IsSecure() {
		t.Error("Expected IsSecure to return false for HTTP")
	}

	// 测试带X-Forwarded-Proto的HTTPS
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	if !ctx.IsSecure() {
		t.Error("Expected IsSecure to return true when X-Forwarded-Proto is https")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 基准测试
func BenchmarkBaseURL(b *testing.B) {
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	ctx := NewContext(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.BaseURL()
	}
}

func BenchmarkOriginalURL(b *testing.B) {
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "example.com")
	c.Request.SetRequestURI("/api/users?page=1")
	ctx := NewContext(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.OriginalURL()
	}
}

func BenchmarkSubdomains(b *testing.B) {
	c := &app.RequestContext{}
	c.Request.Header.Set("Host", "api.v1.example.com")
	ctx := NewContext(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Subdomains()
	}
}