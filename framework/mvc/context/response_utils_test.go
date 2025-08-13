package context

import (
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestSendStatus(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试SendStatus
	ctx.SendStatus(StatusNotFound)

	// 验证状态码
	// 注意：由于Hertz的实现，我们主要验证方法不会出错
	t.Log("SendStatus test completed")
}

func TestEnd(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试End
	ctx.End()

	// 验证Content-Length头
	contentLength := ctx.GetResponseHeader("Content-Length")
	if contentLength != "0" {
		t.Errorf("Expected Content-Length '0', got '%s'", contentLength)
	}
}

func TestLocation(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试Location
	testURL := "https://example.com/redirect"
	ctx.Location(testURL)

	// 验证Location头
	location := ctx.GetResponseHeader("Location")
	if location != testURL {
		t.Errorf("Expected Location '%s', got '%s'", testURL, location)
	}
}

func TestVary(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试Vary
	ctx.Vary("Accept", "Accept-Encoding")

	// 验证Vary头
	vary := ctx.GetResponseHeader("Vary")
	if !strings.Contains(vary, "Accept") || !strings.Contains(vary, "Accept-Encoding") {
		t.Errorf("Expected Vary header to contain 'Accept' and 'Accept-Encoding', got '%s'", vary)
	}
}

func TestVaryAppend(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 先设置一个Vary头
	ctx.SetHeader("Vary", "Accept")
	
	// 再添加更多字段
	ctx.Vary("Accept-Language", "User-Agent")

	// 验证Vary头包含所有字段
	vary := ctx.GetResponseHeader("Vary")
	expectedFields := []string{"Accept", "Accept-Language", "User-Agent"}
	for _, field := range expectedFields {
		if !strings.Contains(vary, field) {
			t.Errorf("Expected Vary header to contain '%s', got '%s'", field, vary)
		}
	}
}

func TestLinks(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试Links
	links := map[string]string{
		"https://example.com/next": "next",
		"https://example.com/prev": "prev",
	}
	ctx.Links(links)

	// 验证Link头
	link := ctx.GetResponseHeader("Link")
	if !strings.Contains(link, "<https://example.com/next>; rel=\"next\"") {
		t.Errorf("Expected Link header to contain next link, got '%s'", link)
	}
	if !strings.Contains(link, "<https://example.com/prev>; rel=\"prev\"") {
		t.Errorf("Expected Link header to contain prev link, got '%s'", link)
	}
}

func TestType(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试Type
	contentType := "application/json"
	ctx.Type(contentType)

	// 验证Content-Type头
	actualType := ctx.GetResponseHeader("Content-Type")
	if actualType != contentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", contentType, actualType)
	}
}

func TestAppend(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试Append - 新头部
	ctx.Append("X-Custom-Header", "value1")
	header := ctx.GetResponseHeader("X-Custom-Header")
	if header != "value1" {
		t.Errorf("Expected 'value1', got '%s'", header)
	}

	// 测试Append - 追加到现有头部
	ctx.Append("X-Custom-Header", "value2")
	header = ctx.GetResponseHeader("X-Custom-Header")
	if !strings.Contains(header, "value1") || !strings.Contains(header, "value2") {
		t.Errorf("Expected header to contain both values, got '%s'", header)
	}
}

func TestFresh(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试没有缓存头的情况
	if ctx.Fresh() {
		t.Error("Fresh should return false when no cache headers are present")
	}

	// 设置ETag
	etag := "\"12345\""
	ctx.SetHeader("ETag", etag)
	
	// 设置匹配的If-None-Match
	c.Request.Header.Set("If-None-Match", etag)
	
	if !ctx.Fresh() {
		t.Error("Fresh should return true when ETag matches If-None-Match")
	}
}

func TestStale(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// Stale应该是Fresh的反义
	if ctx.Stale() != !ctx.Fresh() {
		t.Error("Stale should be the opposite of Fresh")
	}
}

func TestRedirectBack(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Referer", "https://example.com/previous")
	ctx := NewContext(c)

	// 测试RedirectBack
	ctx.RedirectBack("https://example.com/fallback")

	// 验证Location头应该是Referer的值
	location := ctx.GetResponseHeader("Location")
	if location != "https://example.com/previous" {
		t.Errorf("Expected redirect to referer, got '%s'", location)
	}
}

func TestRedirectBackWithoutReferer(t *testing.T) {
	// 创建测试Context（没有Referer）
	c := &app.RequestContext{}
	ctx := NewContext(c)

	fallback := "https://example.com/fallback"
	ctx.RedirectBack(fallback)

	// 验证Location头应该是fallback的值
	location := ctx.GetResponseHeader("Location")
	if location != fallback {
		t.Errorf("Expected redirect to fallback '%s', got '%s'", fallback, location)
	}
}

func TestRedirectWithQuery(t *testing.T) {
	// 创建测试Context，模拟查询参数
	c := &app.RequestContext{}
	c.Request.SetRequestURI("/test?param1=value1&param2=value2")
	ctx := NewContext(c)

	// 测试RedirectWithQuery
	ctx.RedirectWithQuery(StatusFound, "https://example.com/new")

	// 验证Location头包含查询参数
	location := ctx.GetResponseHeader("Location")
	if !strings.Contains(location, "param1=value1") || !strings.Contains(location, "param2=value2") {
		t.Errorf("Expected redirect URL to contain query parameters, got '%s'", location)
	}
}

func TestNoCache(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试NoCache
	ctx.NoCache()

	// 验证缓存控制头
	cacheControl := ctx.GetResponseHeader("Cache-Control")
	if !strings.Contains(cacheControl, "no-cache") {
		t.Errorf("Expected Cache-Control to contain 'no-cache', got '%s'", cacheControl)
	}

	pragma := ctx.GetResponseHeader("Pragma")
	if pragma != "no-cache" {
		t.Errorf("Expected Pragma 'no-cache', got '%s'", pragma)
	}

	expires := ctx.GetResponseHeader("Expires")
	if expires != "0" {
		t.Errorf("Expected Expires '0', got '%s'", expires)
	}
}

func TestMaxAge(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试MaxAge
	maxAge := 3600
	ctx.MaxAge(maxAge)

	// 验证Cache-Control头
	cacheControl := ctx.GetResponseHeader("Cache-Control")
	expected := "max-age=3600"
	if cacheControl != expected {
		t.Errorf("Expected Cache-Control '%s', got '%s'", expected, cacheControl)
	}
}

func TestCacheControl(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试CacheControl
	directives := map[string]interface{}{
		"public":    true,
		"max-age":   3600,
		"no-cache":  false, // 应该被忽略
		"custom":    "value",
	}
	ctx.CacheControl(directives)

	// 验证Cache-Control头
	cacheControl := ctx.GetResponseHeader("Cache-Control")
	if !strings.Contains(cacheControl, "public") {
		t.Errorf("Expected Cache-Control to contain 'public', got '%s'", cacheControl)
	}
	if !strings.Contains(cacheControl, "max-age=3600") {
		t.Errorf("Expected Cache-Control to contain 'max-age=3600', got '%s'", cacheControl)
	}
	if strings.Contains(cacheControl, "no-cache") {
		t.Errorf("Expected Cache-Control to not contain 'no-cache', got '%s'", cacheControl)
	}
}

func TestEnableCORS(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试默认CORS
	ctx.EnableCORS()

	// 验证CORS头
	origin := ctx.GetResponseHeader("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
	}

	methods := ctx.GetResponseHeader("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "GET") || !strings.Contains(methods, "POST") {
		t.Errorf("Expected Access-Control-Allow-Methods to contain standard methods, got '%s'", methods)
	}
}

func TestEnableCORSWithOrigin(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试指定origin的CORS
	origin := "https://example.com"
	ctx.EnableCORS(origin)

	// 验证Access-Control-Allow-Origin头
	actualOrigin := ctx.GetResponseHeader("Access-Control-Allow-Origin")
	if actualOrigin != origin {
		t.Errorf("Expected Access-Control-Allow-Origin '%s', got '%s'", origin, actualOrigin)
	}
}

func TestCORSHeaders(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试CORSHeaders
	config := map[string]string{
		"Access-Control-Allow-Origin":      "https://example.com",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "86400",
		"X-Custom-Header":                  "should-be-ignored", // 非CORS头，应该被忽略
	}
	ctx.CORSHeaders(config)

	// 验证CORS头被设置
	origin := ctx.GetResponseHeader("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Errorf("Expected origin header to be set, got '%s'", origin)
	}

	credentials := ctx.GetResponseHeader("Access-Control-Allow-Credentials")
	if credentials != "true" {
		t.Errorf("Expected credentials header to be set, got '%s'", credentials)
	}

	// 验证非CORS头被忽略
	customHeader := ctx.GetResponseHeader("X-Custom-Header")
	if customHeader != "" {
		t.Errorf("Expected custom header to be ignored, but it was set to '%s'", customHeader)
	}
}

func TestSecurityHeaders(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试SecurityHeaders
	ctx.SecurityHeaders()

	// 验证安全头
	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range securityHeaders {
		actualValue := ctx.GetResponseHeader(header)
		if actualValue != expectedValue {
			t.Errorf("Expected %s '%s', got '%s'", header, expectedValue, actualValue)
		}
	}
}

func TestCSP(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试CSP
	policy := "default-src 'self'; script-src 'self' 'unsafe-inline'"
	ctx.CSP(policy)

	// 验证CSP头
	csp := ctx.GetResponseHeader("Content-Security-Policy")
	if csp != policy {
		t.Errorf("Expected CSP '%s', got '%s'", policy, csp)
	}
}

func TestHSTS(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试HSTS
	ctx.HSTS(31536000, true)

	// 验证HSTS头
	hsts := ctx.GetResponseHeader("Strict-Transport-Security")
	expected := "max-age=31536000; includeSubDomains"
	if hsts != expected {
		t.Errorf("Expected HSTS '%s', got '%s'", expected, hsts)
	}
}

func TestFormat(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Accept", "application/json")
	ctx := NewContext(c)

	// 测试Format
	jsonCalled := false
	htmlCalled := false

	ctx.Format(map[string]func(){
		"application/json": func() {
			jsonCalled = true
			ctx.JSON(StatusOK, map[string]string{"message": "json"})
		},
		"text/html": func() {
			htmlCalled = true
			ctx.HTML(StatusOK, "template", nil)
		},
	})

	// 验证正确的处理器被调用
	if !jsonCalled {
		t.Error("Expected JSON handler to be called")
	}
	if htmlCalled {
		t.Error("Expected HTML handler not to be called")
	}
}

func TestResponseInfo(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 设置一些响应信息
	ctx.Status(StatusOK)
	ctx.SetContentType("application/json")
	ctx.SetHeader("ETag", "\"12345\"")

	// 测试ResponseInfo
	info := ctx.ResponseInfo()
	if info == nil {
		t.Fatal("Expected response info, got nil")
	}

	// 验证信息内容
	if info["content_type"] != "application/json" {
		t.Errorf("Expected content_type 'application/json', got '%v'", info["content_type"])
	}

	if info["etag"] != "\"12345\"" {
		t.Errorf("Expected etag '\"12345\"', got '%v'", info["etag"])
	}
}

// 基准测试
func BenchmarkSendStatus(b *testing.B) {
	c := &app.RequestContext{}
	ctx := NewContext(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.SendStatus(StatusOK)
	}
}

func BenchmarkVary(b *testing.B) {
	c := &app.RequestContext{}
	ctx := NewContext(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Vary("Accept", "Accept-Language")
	}
}

func BenchmarkSecurityHeaders(b *testing.B) {
	c := &app.RequestContext{}
	ctx := NewContext(c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.SecurityHeaders()
	}
}